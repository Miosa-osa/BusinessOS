<script lang="ts">
	import { windowStore } from '$lib/stores/windowStore';
	import { desktopSettings } from '$lib/stores/desktopStore';
	import { userAppsStore } from '$lib/stores/userAppsStore';
	import { api, apiClient } from '$lib/api';

	const iconStyle = $derived($desktopSettings.iconStyle);
	const iconLibrary = $derived($desktopSettings.iconLibrary);

	// Different libraries have EXTREMELY DRAMATIC different styles
	const libraryStrokeWidth = $derived({
		lucide: 2,        // Lucide - balanced, clean
		phosphor: 3,      // Phosphor - VERY bold, thick
		tabler: 1.2,      // Tabler - very thin, hairline
		heroicons: 2.5    // Heroicons - solid, medium-bold
	}[iconLibrary] || 2);

	const libraryLineCap = $derived<'round' | 'square' | 'butt'>(
		iconLibrary === 'tabler' ? 'square' : 'round'
	);

	const libraryLineJoin = $derived<'round' | 'miter' | 'bevel'>(
		iconLibrary === 'tabler' ? 'miter' : 'round'
	);

	// Icon scale varies EXTREMELY by library
	const libraryIconScale = $derived({
		lucide: 1,
		phosphor: 1.25,   // Phosphor icons 25% larger
		tabler: 0.85,     // Tabler icons 15% smaller
		heroicons: 1.15   // Heroicons 15% larger
	}[iconLibrary] || 1);

	// Different SVG filters per library for OBVIOUS visual differences
	const librarySvgFilter = $derived({
		lucide: 'none',
		phosphor: 'drop-shadow(0 2px 3px rgba(0,0,0,0.25))',    // Noticeable shadow
		tabler: 'saturate(0.7)',                                  // Desaturated look
		heroicons: 'drop-shadow(0 1px 2px rgba(0,0,0,0.2)) saturate(1.2)'  // Shadow + vivid
	}[iconLibrary] || 'none');

	// Check if any windows are open (for collapsed chat bubble state)
	const hasOpenWindows = $derived($windowStore.windows.filter(w => !w.minimized).length > 0);

	// Collapsed bubble state
	let isHoveringCollapsed = $state(false);
	let collapsedVoiceActive = $state(false);

	// Quick chat state
	let chatInput = $state('');
	let isExpanded = $state(false);
	let isLoading = $state(false);
	let lastResponse = $state('');
	let showResponse = $state(false);
	let chatInputElement: HTMLTextAreaElement | undefined = $state(undefined);

	// Context/Project selection
	let selectedProject = $state<{ id: string; name: string } | null>(null);
	let showProjectSelector = $state(false);
	let highlightedProjectIndex = $state(-1);
	let projects = $state<{ id: string; name: string }[]>([]);
	let loadingProjects = $state(false);

	// Model selection
	let selectedModel = $state<{ id: string; name: string; isLocal: boolean } | null>(null);
	let showModelSelector = $state(false);
	let localModels = $state<{ id: string; name: string }[]>([]);
	let cloudModels = $state<{ id: string; name: string }[]>([]);
	let loadingModels = $state(false);
	let activeProvider = $state<string>('');
	let defaultModelId = $state<string>('');
	let ollamaAvailable = $state<boolean | null>(null); // null = unknown, true = available, false = not available

	// Model pull state
	let isPulling = $state(false);
	let pullingModel = $state('');
	let pullProgress = $state<{ status: string; percent?: number } | null>(null);
	let recommendedModels = $state<{ id: string; name: string; size: string }[]>([]);

	// Voice recording state
	let isRecording = $state(false);
	let mediaRecorder: MediaRecorder | null = null;
	let audioChunks: Blob[] = [];
	let recordingDuration = $state(0);
	let recordingInterval: number | null = null;

	// Audio visualization
	let audioContext: AudioContext | null = null;
	let analyser: AnalyserNode | null = null;
	let audioDataArray: Uint8Array | null = null;
	let waveformBars = $state<number[]>(Array(20).fill(2));
	let animationFrameId: number | null = null;

	// File upload state
	let isDraggingFile = $state(false);
	let attachedFiles = $state<File[]>([]);
	let fileInputElement: HTMLInputElement | undefined = $state(undefined);

	// Content size detection for dynamic width
	let contentSize = $derived(() => {
		const lineCount = (chatInput.match(/\n/g) || []).length + 1;
		const charCount = chatInput.length;

		if (lineCount >= 3 || charCount > 150) return 'large';
		if (lineCount >= 2 || charCount > 80) return 'medium';
		if (charCount > 40) return 'small-expand';
		return 'default';
	});

	// Load contexts on mount
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';

	// Global keyboard handler for Ctrl+Space dictation shortcut
	function handleGlobalKeydown(e: KeyboardEvent) {
		// Ctrl+Space to start dictation when collapsed bubble is visible
		if (e.ctrlKey && e.code === 'Space' && hasOpenWindows && !collapsedVoiceActive && !isRecording) {
			e.preventDefault();
			handleCollapsedBubbleClick();
		}
	}

	onMount(() => {
		if (browser) {
			// Load in background - don't block UI
			loadProjects();
			loadModels();
			// Add global keyboard listener for dictation shortcut
			window.addEventListener('keydown', handleGlobalKeydown);
		}
	});

	onDestroy(() => {
		if (browser) {
			window.removeEventListener('keydown', handleGlobalKeydown);
		}
	});

	async function loadProjects() {
		loadingProjects = true;
		try {
			const projectsList = await api.getProjects();
			projects = projectsList.map((p: any) => ({ id: p.id, name: p.name }));
		} catch (error) {
			console.error('Failed to load projects:', error);
		} finally {
			loadingProjects = false;
		}
	}

	async function loadModels() {
		loadingModels = true;
		try {
			// 1. Load providers config to get active provider and default model from settings
			const providersRes = await apiClient.get('/ai/providers');
			if (providersRes.ok) {
				const data = await providersRes.json();
				activeProvider = data.active_provider || 'ollama_local';
				defaultModelId = data.default_model || '';
			}

			// 2. Load all models (cloud models from configured providers)
			const allModelsRes = await apiClient.get('/ai/models');
			if (allModelsRes.ok) {
				const data = await allModelsRes.json();
				const allModels = data.models || [];
				// Cloud models are non-local (anthropic, openai, etc.)
				cloudModels = allModels
					.filter((m: any) => m.provider !== 'ollama_local' && m.provider !== 'ollama')
					.map((m: any) => ({
						id: m.id || m.name,
						name: m.name || m.id
					}));
				// Use default_model from response if not from providers
				if (!defaultModelId && data.default_model) {
					defaultModelId = data.default_model;
				}
			}

			// 3. Load local models from Ollama API
			const localRes = await apiClient.get('/ai/models/local');
			if (localRes.ok) {
				const data = await localRes.json();
				// 200 OK means Ollama is running and available
				ollamaAvailable = true;

				localModels = (data.models || [])
					.filter((m: any) => {
						// Filter out tiny cloud reference stubs
						const nameOrId = (m.id || '') + (m.name || '');
						const isCloudRef = nameOrId.toLowerCase().includes('cloud') &&
							(m.size === '< 1 KB' || m.size === '0 B' || !m.size);
						return !isCloudRef;
					})
					.map((m: any) => ({
						id: m.id || m.name,
						name: m.name || m.id
					}));
			} else {
				// 503 or other error means Ollama not available (not installed or not running)
				ollamaAvailable = false;
				localModels = [];
			}

			// 4. Load system info for recommended models (for Ollama pulls)
			const systemRes = await apiClient.get('/ai/system');
			if (systemRes.ok) {
				const systemInfo = await systemRes.json();
				if (systemInfo.recommended_models && systemInfo.recommended_models.length > 0) {
					recommendedModels = systemInfo.recommended_models.map((m: any) => ({
						id: m.name || m.id,
						name: m.name || m.id,
						size: m.ram_required || '~4GB'
					}));
				}
			}
			// Fallback to sensible defaults if no recommended models from API
			if (recommendedModels.length === 0) {
				recommendedModels = [
					{ id: 'llama3.2:3b', name: 'Llama 3.2 3B', size: '~2GB' },
					{ id: 'llama3.2:latest', name: 'Llama 3.2 7B', size: '~4GB' },
					{ id: 'mistral:7b', name: 'Mistral 7B', size: '~4GB' },
					{ id: 'qwen2.5:7b', name: 'Qwen 2.5 7B', size: '~4GB' }
				];
			}

			// 5. Set selected model based on default from settings
			if (!selectedModel) {
				// Try to find the default model in local or cloud models
				const defaultInLocal = localModels.find(m => m.id === defaultModelId || m.name === defaultModelId);
				const defaultInCloud = cloudModels.find(m => m.id === defaultModelId || m.name === defaultModelId);

				if (defaultInLocal) {
					selectedModel = { ...defaultInLocal, isLocal: true };
				} else if (defaultInCloud) {
					selectedModel = { ...defaultInCloud, isLocal: false };
				} else if (localModels.length > 0) {
					// Fallback to first local model
					selectedModel = { ...localModels[0], isLocal: true };
				} else if (cloudModels.length > 0) {
					// Fallback to first cloud model
					selectedModel = { ...cloudModels[0], isLocal: false };
				}
			}
		} catch (error) {
			console.error('Failed to load models:', error);
			// Default to first available model
			if (!selectedModel) {
				if (localModels.length > 0) {
					selectedModel = { ...localModels[0], isLocal: true };
				} else if (cloudModels.length > 0) {
					selectedModel = { ...cloudModels[0], isLocal: false };
				}
			}
		} finally {
			loadingModels = false;
		}
	}

	async function pullModel(modelId: string) {
		if (isPulling) return;

		isPulling = true;
		pullingModel = modelId;
		pullProgress = { status: 'Starting...' };

		try {
			const response = await fetch('/api/ai/models/pull', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ model: modelId }),
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error('Failed to pull model');
			}

			const reader = response.body?.getReader();
			if (!reader) throw new Error('No response body');

			const decoder = new TextDecoder();
			let buffer = '';

			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				buffer += decoder.decode(value, { stream: true });
				const lines = buffer.split('\n');
				buffer = lines.pop() || '';

				for (const line of lines) {
					if (!line.trim()) continue;
					try {
						const data = JSON.parse(line);
						if (data.status) {
							pullProgress = {
								status: data.status,
								percent: data.completed && data.total ? Math.round((data.completed / data.total) * 100) : undefined
							};
						}
					} catch {}
				}
			}

			// Refresh models after pull
			await loadModels();
			pullProgress = null;
		} catch (error) {
			console.error('Failed to pull model:', error);
			pullProgress = { status: 'Failed to pull model' };
			setTimeout(() => { pullProgress = null; }, 3000);
		} finally {
			isPulling = false;
			pullingModel = '';
		}
	}

	interface DockItem {
		id: string;
		module: string;
		label: string;
		isOpen: boolean;
		isMinimized: boolean;
		windowId?: string;
		folderId?: string;
		folderColor?: string;
	}

	// Icon data for each module
	const moduleIcons: Record<string, { path?: string; color: string; bgColor: string; isTerminal?: boolean; isFolder?: boolean; isFinder?: boolean; imageUrl?: string }> = {
		'app-store': {
			imageUrl: '/logos/integrations/AppleStore_whitelogo.png',
			color: '#0D84FF',
			bgColor: '#0D84FF'
		},
		platform: {
			path: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5',
			color: '#333333',
			bgColor: '#F5F5F5'
		},
		folder: {
			path: 'M3 7V17C3 18.1046 3.89543 19 5 19H19C20.1046 19 21 18.1046 21 17V9C21 7.89543 20.1046 7 19 7H12L10 5H5C3.89543 5 3 5.89543 3 7Z',
			color: '#3B82F6',
			bgColor: '#EFF6FF',
			isFolder: true
		},
		terminal: {
			path: 'M4 17l6-6-6-6M12 19h8',
			color: '#00FF00',
			bgColor: '#1E1E1E',
			isTerminal: true
		},
		dashboard: {
			path: 'M4 5a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v2a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1h-4a1 1 0 01-1-1v-5zm-10 1a1 1 0 011-1h4a1 1 0 011 1v3a1 1 0 01-1 1H5a1 1 0 01-1-1v-3z',
			color: '#1E88E5',
			bgColor: '#E3F2FD'
		},
		chat: {
			path: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z',
			color: '#43A047',
			bgColor: '#E8F5E9'
		},
		tasks: {
			path: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4',
			color: '#FB8C00',
			bgColor: '#FFF3E0'
		},
		projects: {
			path: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z',
			color: '#8E24AA',
			bgColor: '#F3E5F5'
		},
		team: {
			path: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z',
			color: '#00ACC1',
			bgColor: '#E0F7FA'
		},
		clients: {
			path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4',
			color: '#7B1FA2',
			bgColor: '#F3E5F5'
		},
		contexts: {
			path: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
			color: '#5E35B1',
			bgColor: '#EDE7F6'
		},
		nodes: {
			path: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z',
			color: '#E53935',
			bgColor: '#FFEBEE'
		},
		daily: {
			path: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
			color: '#039BE5',
			bgColor: '#E1F5FE'
		},
		settings: {
			path: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
			color: '#546E7A',
			bgColor: '#ECEFF1'
		},
		trash: {
			path: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16',
			color: '#78909C',
			bgColor: '#ECEFF1'
		},
		files: {
			path: 'M3 7V17C3 18.1046 3.89543 19 5 19H19C20.1046 19 21 18.1046 21 17V9C21 7.89543 20.1046 7 19 7H12L10 5H5C3.89543 5 3 5.89543 3 7Z M7 13h10M7 16h6',
			color: '#2196F3',
			bgColor: '#E3F2FD'
		},
		calendar: {
			path: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
			color: '#E91E63',
			bgColor: '#FCE4EC'
		},
		'ai-settings': {
			path: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z',
			color: '#9C27B0',
			bgColor: '#F3E5F5'
		},
		help: {
			path: 'M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
			color: '#0EA5E9',
			bgColor: '#E0F2FE'
		},
		finder: {
			path: 'M5 3h14a2 2 0 012 2v14a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2z',
			color: '#1C9BF6',
			bgColor: 'linear-gradient(180deg, #3FBBF7 0%, #1C7FE6 100%)',
			isFinder: true
		}
	};

	const staticModuleLabels: Record<string, string> = {
		platform: 'Business OS',
		terminal: 'Terminal',
		dashboard: 'Dashboard',
		chat: 'Chat',
		tasks: 'Tasks',
		projects: 'Projects',
		team: 'Team',
		clients: 'Clients',
		contexts: 'Contexts',
		nodes: 'Nodes',
		daily: 'Daily Log',
		settings: 'Settings',
		trash: 'Trash',
		files: 'Files',
		calendar: 'Calendar',
		'ai-settings': 'AI Settings',
		help: 'Help',
		finder: 'Finder'
	};

	// Merge static labels with dynamic user app labels
	const moduleLabels = $derived(() => {
		const labels = { ...staticModuleLabels };

		// Add user app labels
		for (const app of $userAppsStore.apps) {
			labels[`user-app-${app.id}`] = app.name;
		}

		return labels;
	});

	let hoveredIndex = $state<number | null>(null);
	let isDragOver = $state(false);

	// Handle drag over for pinning apps
	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		if (event.dataTransfer) {
			event.dataTransfer.dropEffect = 'copy';
		}
		isDragOver = true;
	}

	function handleDragLeave() {
		isDragOver = false;
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		isDragOver = false;
		const module = event.dataTransfer?.getData('text/plain');
		const iconId = event.dataTransfer?.getData('text/icon-id');

		if (module && module !== 'trash' && module !== 'separator') {
			let moduleToAdd = module;

			// If it's a folder, look up the folder ID
			if (module === 'folder' && iconId) {
				const icon = $windowStore.desktopIcons.find(i => i.id === iconId);
				if (icon?.type === 'folder' && icon.folderId) {
					moduleToAdd = `folder-${icon.folderId}`;
				}
			}

			// Add to pinned items if not already there
			if (!$windowStore.dockPinnedItems.includes(moduleToAdd)) {
				windowStore.addToDock(moduleToAdd);
			}
		}
	}

	// Helper to check if a module is a folder
	function isFolder(module: string): boolean {
		return module === 'folder' || module.startsWith('folder-');
	}

	// Get folder data from a folder module string
	function getFolderData(module: string) {
		if (module === 'folder') return null;
		const folderId = module.replace('folder-', '');
		const folderIcon = $windowStore.desktopIcons.find(
			i => i.type === 'folder' && i.folderId === folderId
		);
		const folder = $windowStore.folders.find(f => f.id === folderId);
		return {
			folderId,
			label: folder?.name || folderIcon?.label || 'Folder',
			color: folder?.color || folderIcon?.folderColor || '#3B82F6'
		};
	}

	// Build dock items from pinned items and open windows
	const dockItems = $derived(() => {
		const items: DockItem[] = [];

		// Add pinned items
		for (const module of $windowStore.dockPinnedItems) {
			const windows = $windowStore.windows.filter(w => w.module === module);
			const isOpen = windows.length > 0;
			const minimizedWindow = windows.find(w => w.minimized);

			// Check if it's a folder
			const folderData = isFolder(module) ? getFolderData(module) : null;

			items.push({
				id: `pinned-${module}`,
				module,
				label: folderData?.label || moduleLabels()[module] || module,
				isOpen,
				isMinimized: !!minimizedWindow,
				windowId: minimizedWindow?.id || windows[0]?.id,
				folderId: folderData?.folderId,
				folderColor: folderData?.color
			});
		}

		// Add separator marker
		items.push({
			id: 'separator',
			module: 'separator',
			label: '',
			isOpen: false,
			isMinimized: false
		});

		// Add open windows that aren't pinned
		const pinnedModules = new Set($windowStore.dockPinnedItems);
		const openWindowModules = [...new Set($windowStore.windows.map(w => w.module))];

		for (const module of openWindowModules) {
			if (!pinnedModules.has(module)) {
				const windows = $windowStore.windows.filter(w => w.module === module);
				const minimizedWindow = windows.find(w => w.minimized);

				items.push({
					id: `open-${module}`,
					module,
					label: moduleLabels()[module] || module,
					isOpen: true,
					isMinimized: !!minimizedWindow,
					windowId: minimizedWindow?.id || windows[0]?.id
				});
			}
		}

		// Add trash at the end
		if (!pinnedModules.has('trash')) {
			items.push({
				id: 'separator-2',
				module: 'separator',
				label: '',
				isOpen: false,
				isMinimized: false
			});
			items.push({
				id: 'trash',
				module: 'trash',
				label: 'Trash',
				isOpen: $windowStore.windows.some(w => w.module === 'trash'),
				isMinimized: $windowStore.windows.some(w => w.module === 'trash' && w.minimized),
				windowId: $windowStore.windows.find(w => w.module === 'trash')?.id
			});
		}

		return items;
	});

	// Format recording duration
	function formatDuration(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	// Voice recording handlers
	async function startRecording() {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

			// Set up audio context for visualization
			audioContext = new AudioContext();
			analyser = audioContext.createAnalyser();
			const source = audioContext.createMediaStreamSource(stream);
			source.connect(analyser);
			analyser.fftSize = 64;
			audioDataArray = new Uint8Array(analyser.frequencyBinCount);

			mediaRecorder = new MediaRecorder(stream);
			audioChunks = [];

			mediaRecorder.ondataavailable = (event) => {
				audioChunks.push(event.data);
			};

			mediaRecorder.onstop = async () => {
				const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
				await transcribeAudio(audioBlob);
				stream.getTracks().forEach(track => track.stop());
			};

			mediaRecorder.start();
			isRecording = true;
			isExpanded = true;
			recordingDuration = 0;

			recordingInterval = window.setInterval(() => {
				recordingDuration++;
			}, 1000);

			updateWaveform();
		} catch (error) {
			console.error('Failed to start recording:', error);
		}
	}

	function stopRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.stop();
		}
		isRecording = false;

		if (recordingInterval) {
			clearInterval(recordingInterval);
			recordingInterval = null;
		}

		if (animationFrameId) {
			cancelAnimationFrame(animationFrameId);
			animationFrameId = null;
		}

		waveformBars = Array(20).fill(2);
	}

	function updateWaveform() {
		if (!isRecording || !analyser || !audioDataArray) return;

		analyser.getByteTimeDomainData(audioDataArray as Uint8Array<ArrayBuffer>);

		const newBars: number[] = [];
		const step = Math.floor(audioDataArray.length / 20);

		for (let i = 0; i < 20; i++) {
			const index = i * step;
			const value = audioDataArray[index];
			const deviation = Math.abs(value - 128);
			const height = Math.max(2, Math.min(20, 2 + (deviation / 128) * 36));
			newBars.push(height);
		}

		waveformBars = newBars;
		animationFrameId = requestAnimationFrame(updateWaveform);
	}

	async function transcribeAudio(audioBlob: Blob) {
		isLoading = true;
		showResponse = false;

		try {
			const formData = new FormData();
			formData.append('audio', audioBlob, 'recording.webm');

			const response = await fetch('/api/transcribe', {
				method: 'POST',
				body: formData
			});

			if (response.ok) {
				const data = await response.json();
				if (data.text) {
					chatInput = data.text;
				}
			}

			// TODO: Auto-save voice note when API endpoint is implemented
			// apiClient.uploadVoiceNote(audioBlob).catch(err => {
			// 	console.warn('Voice note auto-save failed (non-critical):', err);
			// });
		} catch (error) {
			console.error('Transcription error:', error);
		} finally {
			isLoading = false;
		}
	}

	// Quick chat handlers
	function handleChatFocus() {
		isExpanded = true;
	}

	function handleChatBlur() {
		if (!chatInput.trim() && !showResponse && !isRecording) {
			setTimeout(() => {
				if (!chatInput.trim() && !showResponse && !isRecording) {
					isExpanded = false;
				}
			}, 150);
		}
	}

	async function handleChatSubmit() {
		if (!chatInput.trim() || isLoading) return;

		// Require project selection
		if (!selectedProject) {
			showProjectSelector = true;
			highlightedProjectIndex = 0;
			return;
		}

		const message = chatInput.trim();

		// Store the message to pass to chat module
		if (browser) {
			// Convert files to serializable format (store file names/sizes)
			const fileData = attachedFiles.map(f => ({
				name: f.name,
				size: f.size,
				type: f.type
			}));

			// Store message in sessionStorage so chat module can pick it up
			sessionStorage.setItem('quickChatMessage', JSON.stringify({
				message,
				projectId: selectedProject.id,
				projectName: selectedProject.name,
				model: selectedModel?.name,
				isLocalModel: selectedModel?.isLocal,
				files: fileData,
				timestamp: Date.now(),
				isNewConversation: true
			}));
		}

		// Clear input and reset
		chatInput = '';
		attachedFiles = [];
		isExpanded = false;
		showResponse = false;
		resetTextareaHeight();

		// Open the chat module
		windowStore.openWindow('chat');
	}

	function handleChatKeyDown(e: KeyboardEvent) {
		// Handle keyboard navigation when project selector is open
		if (showProjectSelector && projects.length > 0) {
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				highlightedProjectIndex = Math.min(highlightedProjectIndex + 1, projects.length - 1);
				return;
			} else if (e.key === 'ArrowUp') {
				e.preventDefault();
				highlightedProjectIndex = Math.max(highlightedProjectIndex - 1, 0);
				return;
			} else if (e.key === 'Enter') {
				e.preventDefault();
				if (highlightedProjectIndex >= 0 && highlightedProjectIndex < projects.length) {
					selectedProject = projects[highlightedProjectIndex];
					showProjectSelector = false;
					highlightedProjectIndex = -1;
				}
				return;
			} else if (e.key === 'Escape') {
				e.preventDefault();
				showProjectSelector = false;
				highlightedProjectIndex = -1;
				return;
			}
		}

		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleChatSubmit();
		} else if (e.key === 'Escape') {
			chatInput = '';
			showResponse = false;
			isExpanded = false;
			resetTextareaHeight();
			if (isRecording) stopRecording();
			chatInputElement?.blur();
		} else if (e.key === 'd' && (e.metaKey || e.ctrlKey)) {
			e.preventDefault();
			toggleRecording();
		}
	}

	// Auto-resize textarea based on content
	function autoResizeTextarea() {
		if (!chatInputElement) return;
		// Reset height to auto to get the proper scrollHeight
		chatInputElement.style.height = 'auto';
		// Set height to scrollHeight, with min and max constraints
		const newHeight = Math.min(Math.max(chatInputElement.scrollHeight, 24), 120);
		chatInputElement.style.height = `${newHeight}px`;
	}

	function resetTextareaHeight() {
		if (!chatInputElement) return;
		chatInputElement.style.height = '24px';
	}

	// Watch for input changes to auto-resize
	$effect(() => {
		if (chatInput !== undefined && chatInputElement) {
			// Use requestAnimationFrame to ensure DOM has updated
			requestAnimationFrame(autoResizeTextarea);
		}
	});

	function openFullChat() {
		windowStore.openWindow('chat');
		showResponse = false;
		isExpanded = false;
	}

	// Handle collapsed bubble click/hover - start dictation immediately
	async function handleCollapsedBubbleClick() {
		collapsedVoiceActive = true;
		// Start recording without expanding (we'll stay in collapsed mode)
		await startRecordingCollapsed();
	}

	// Start recording for collapsed bubble mode (doesn't expand)
	async function startRecordingCollapsed() {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

			// Set up audio context for visualization
			audioContext = new AudioContext();
			analyser = audioContext.createAnalyser();
			const source = audioContext.createMediaStreamSource(stream);
			source.connect(analyser);
			analyser.fftSize = 64;
			audioDataArray = new Uint8Array(analyser.frequencyBinCount);

			mediaRecorder = new MediaRecorder(stream);
			audioChunks = [];

			mediaRecorder.ondataavailable = (event) => {
				audioChunks.push(event.data);
			};

			mediaRecorder.onstop = async () => {
				const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
				await transcribeAudio(audioBlob);
				stream.getTracks().forEach(track => track.stop());
			};

			mediaRecorder.start();
			isRecording = true;
			// Don't set isExpanded = true for collapsed mode
			recordingDuration = 0;

			recordingInterval = window.setInterval(() => {
				recordingDuration++;
			}, 1000);

			updateWaveform();
		} catch (error) {
			console.error('Failed to start recording:', error);
			collapsedVoiceActive = false;
		}
	}

	// Handle collapsed voice done - stop recording and open chat with transcribed text
	async function handleCollapsedVoiceDone() {
		// Store current chatInput to check for changes
		const previousInput = chatInput;

		// Stop recording (triggers transcription via onstop)
		stopRecording();

		// Wait for transcription to complete by polling for chatInput changes
		// Transcription typically takes 1-3 seconds
		let attempts = 0;
		const maxAttempts = 30; // 3 seconds max wait

		const waitForTranscription = () => {
			return new Promise<void>((resolve) => {
				const checkInterval = setInterval(() => {
					attempts++;
					if (chatInput !== previousInput || attempts >= maxAttempts) {
						clearInterval(checkInterval);
						resolve();
					}
				}, 100);
			});
		};

		await waitForTranscription();

		collapsedVoiceActive = false;
		isHoveringCollapsed = false;
		isRecording = false;

		// Open chat window with transcribed text
		if (chatInput.trim() && browser) {
			// Store transcript in sessionStorage so chat module can pick it up
			// Use autoSend: false to just put text in input without sending
			sessionStorage.setItem('voiceTranscript', JSON.stringify({
				message: chatInput.trim(),
				timestamp: Date.now(),
				autoSend: false
			}));

			// Clear local chatInput since it will be in the chat window now
			chatInput = '';

			windowStore.openWindow('chat');
		}
	}

	// Handle collapsed voice cancel - stop recording without opening chat
	function handleCollapsedVoiceCancel() {
		stopRecording();
		collapsedVoiceActive = false;
		isHoveringCollapsed = false;
		chatInput = ''; // Clear any partial transcription
	}

	// File handling
	function handleFileDragOver(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		isDraggingFile = true;
	}

	function handleFileDragLeave(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		isDraggingFile = false;
	}

	function handleFileDrop(e: DragEvent) {
		e.preventDefault();
		e.stopPropagation();
		isDraggingFile = false;

		const files = e.dataTransfer?.files;
		if (files && files.length > 0) {
			addFiles(Array.from(files));
		}
	}

	function handleFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		if (input.files && input.files.length > 0) {
			addFiles(Array.from(input.files));
			input.value = ''; // Reset for next selection
		}
	}

	function addFiles(files: File[]) {
		// Limit to 5 files
		const newFiles = [...attachedFiles, ...files].slice(0, 5);
		attachedFiles = newFiles;
		isExpanded = true;
	}

	function removeFile(index: number) {
		attachedFiles = attachedFiles.filter((_, i) => i !== index);
	}

	function openFileDialog() {
		fileInputElement?.click();
	}

	function toggleRecording() {
		if (isRecording) {
			stopRecording();
		} else {
			startRecording();
		}
	}

	function handleItemClick(item: DockItem) {
		if (item.module === 'separator') return;

		if (item.isMinimized && item.windowId) {
			windowStore.restoreWindow(item.windowId);
		} else if (item.isOpen && item.windowId) {
			windowStore.focusWindow(item.windowId);
		} else if (item.folderId) {
			// Open folder
			windowStore.openFolder(item.folderId);
		} else {
			windowStore.openWindow(item.module);
		}
	}

	// Get the icon data for an item, handling folders specially
	function getItemIcon(item: DockItem) {
		if (item.folderId) {
			return {
				path: moduleIcons.folder.path,
				color: item.folderColor || '#3B82F6',
				bgColor: `${item.folderColor || '#3B82F6'}20`,
				isFolder: true
			};
		}

		// Check if it's a user app
		if (item.module.startsWith('user-app-')) {
			const appId = item.module.replace('user-app-', '');
			const app = $userAppsStore.apps.find(a => a.id === appId);

			if (app && app.logo_url) {
				return {
					imageUrl: app.logo_url,
					color: app.color || '#6366F1',
					bgColor: 'white',
					isUserApp: true
				};
			} else if (app) {
				// Fallback to default icon if no logo
				return {
					path: moduleIcons.dashboard.path,
					color: app.color || '#6366F1',
					bgColor: `${app.color || '#6366F1'}20`
				};
			}
		}

		return moduleIcons[item.module] || moduleIcons.dashboard;
	}

	function getScale(index: number): number {
		if (hoveredIndex === null) return 1;

		const distance = Math.abs(index - hoveredIndex);
		if (distance === 0) return 1.4;
		if (distance === 1) return 1.2;
		if (distance === 2) return 1.1;
		return 1;
	}

	function getTranslateY(index: number): number {
		if (hoveredIndex === null) return 0;

		const distance = Math.abs(index - hoveredIndex);
		if (distance === 0) return -12;
		if (distance === 1) return -6;
		if (distance === 2) return -2;
		return 0;
	}
</script>

<div class="dock-container">
	<!-- Hidden file input -->
	<input
		type="file"
		bind:this={fileInputElement}
		onchange={handleFileSelect}
		multiple
		accept="*/*"
		class="hidden-file-input"
	/>

	<!-- Collapsed Chat Bubble (shown when windows are open) -->
	{#if hasOpenWindows && (!isRecording || collapsedVoiceActive) && !showResponse && !isExpanded}
		<div
			class="collapsed-chat-bubble"
			class:hovering={isHoveringCollapsed && !collapsedVoiceActive}
			class:voice-active={collapsedVoiceActive}
			onmouseenter={() => { if (!collapsedVoiceActive) isHoveringCollapsed = true; }}
			onmouseleave={() => { if (!collapsedVoiceActive) isHoveringCollapsed = false; }}
			role="button"
			tabindex="0"
			aria-label="Click to start voice input"
		>
			{#if collapsedVoiceActive}
				<!-- Voice recording active - matches example design -->
				<div class="collapsed-voice-recording">
					<button class="btn-pill btn-pill-danger btn-pill-icon btn-pill-sm" onclick={handleCollapsedVoiceCancel} title="Cancel recording">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
							<line x1="18" y1="6" x2="6" y2="18"/>
							<line x1="6" y1="6" x2="18" y2="18"/>
						</svg>
					</button>
					<div class="collapsed-waveform-container">
						<div class="collapsed-waveform">
							{#each waveformBars.slice(0, 10) as height}
								<div class="collapsed-bar" style="height: {Math.max(4, height * 0.7)}px"></div>
							{/each}
						</div>
						<span class="collapsed-timer">{formatDuration(recordingDuration)}</span>
					</div>
					<button class="btn-pill btn-pill-primary btn-pill-icon btn-pill-sm" onclick={handleCollapsedVoiceDone} title="Stop and send to chat">
						<svg viewBox="0 0 24 24" fill="currentColor">
							<rect x="6" y="6" width="12" height="12" rx="2"/>
						</svg>
					</button>
				</div>
			{:else if isHoveringCollapsed}
				<!-- Hover state - macOS dictation style popup -->
				<div class="collapsed-hover-state">
					<div class="collapsed-hint-bubble">
						<span>Click or hold</span>
						<kbd>⌃ Ctrl</kbd>
						<span>to start dictating</span>
					</div>
					<button
						class="btn-pill btn-pill-secondary btn-pill-sm"
						onclick={handleCollapsedBubbleClick}
						title="Start dictating"
					>
						<span class="collapsed-dots">............</span>
					</button>
				</div>
			{:else}
				<!-- Default collapsed state - dots pill -->
				<button
					class="btn-pill btn-pill-ghost btn-pill-sm"
					onclick={handleCollapsedBubbleClick}
					title="Click to start voice input"
				>
					<span class="collapsed-dots-default">•••</span>
				</button>
			{/if}
		</div>
	{:else}
		<!-- Full Quick Chat Input Bar (shown when no windows are open) -->
		<div
			class="quick-chat"
			class:expanded={isExpanded || showResponse || isRecording}
			class:size-small-expand={contentSize() === 'small-expand'}
			class:size-medium={contentSize() === 'medium'}
			class:size-large={contentSize() === 'large'}
			class:dragging-file={isDraggingFile}
			ondragover={handleFileDragOver}
			ondragleave={handleFileDragLeave}
			ondrop={handleFileDrop}
		>
		{#if showResponse}
			<!-- Response display -->
			<div class="quick-chat-response">
				<div class="response-header">
					<div class="ai-badge">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
						</svg>
						<span>AI</span>
					</div>
					<div class="response-actions">
						<button class="btn-pill btn-pill-secondary btn-pill-icon btn-pill-sm" onclick={openFullChat} title="Open full chat (⌘+Enter)">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
								<polyline points="15 3 21 3 21 9"/>
								<line x1="10" y1="14" x2="21" y2="3"/>
							</svg>
						</button>
						<button class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-sm" onclick={() => { showResponse = false; isExpanded = false; }} title="Close (Esc)">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="18" y1="6" x2="6" y2="18"/>
								<line x1="6" y1="6" x2="18" y2="18"/>
							</svg>
						</button>
					</div>
				</div>
				<div class="response-content">
					{#if isLoading}
						<div class="loading-dots">
							<span></span><span></span><span></span>
						</div>
					{:else}
						{lastResponse}
					{/if}
				</div>
			</div>
		{/if}

		<!-- Recording indicator - matches popup-chat style -->
		{#if isRecording}
			<div class="recording-waveform-bar">
				<button class="btn-pill btn-pill-danger btn-pill-icon btn-pill-sm" onclick={stopRecording} title="Cancel">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
					</svg>
				</button>
				<div class="recording-waveform">
					{#each waveformBars as height}
						<div class="recording-bar" style="height: {height}px"></div>
					{/each}
				</div>
				<span class="recording-duration">{formatDuration(recordingDuration)}</span>
				<button class="btn-pill btn-pill-primary btn-pill-icon btn-pill-sm" onclick={stopRecording} title="Done">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polyline points="20 6 9 17 4 12"/>
					</svg>
				</button>
			</div>
		{/if}

		<!-- Text area (top) -->
		<div class="quick-chat-textarea-wrapper">
			<textarea
				bind:this={chatInputElement}
				bind:value={chatInput}
				placeholder={selectedProject ? `Ask about ${selectedProject.name}...` : 'Ask anything...'}
				onfocus={handleChatFocus}
				onblur={handleChatBlur}
				onkeydown={handleChatKeyDown}
				rows="1"
			></textarea>
		</div>

		<!-- Attached files display -->
		{#if attachedFiles.length > 0}
			<div class="attached-files">
				{#each attachedFiles as file, index}
					<div class="attached-file">
						<svg class="file-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
							<polyline points="14 2 14 8 20 8"/>
						</svg>
						<span class="file-name">{file.name}</span>
						<button class="btn-pill btn-pill-danger btn-pill-icon btn-pill-sm" onclick={() => removeFile(index)}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
							</svg>
						</button>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Bottom toolbar (fixed) -->
		<div class="quick-chat-toolbar">
			<!-- Left side - context and options -->
			<div class="toolbar-left">
				<!-- Attachment button -->
				<button class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-sm" onclick={openFileDialog} title="Attach files">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48"/>
					</svg>
				</button>

				<!-- Project selector (required) -->
				<div class="context-selector">
					<button
						class="btn-pill btn-pill-sm {selectedProject ? 'btn-pill-secondary' : 'btn-pill-ghost'}"
						class:required={!selectedProject}
						onclick={() => { showProjectSelector = !showProjectSelector; highlightedProjectIndex = showProjectSelector ? 0 : -1; }}
						title="Select project (required)"
					>
						{#if selectedProject}
							<svg class="project-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
							</svg>
							<span class="context-name">{selectedProject.name}</span>
						{:else}
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
							</svg>
							<span class="placeholder-text">Project</span>
						{/if}
						<svg class="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M6 9l6 6 6-6"/>
						</svg>
					</button>

					{#if showProjectSelector}
						<div class="context-dropdown">
							<div class="dropdown-header">Select Project</div>
							{#if loadingProjects}
								<div class="context-option empty">Loading...</div>
							{:else if projects.length === 0}
								<div class="context-option empty">No projects found</div>
							{:else}
								{#each projects as proj, index}
									<button
										class="btn-pill btn-pill-ghost btn-pill-sm btn-pill-block justify-start"
										class:selected={selectedProject?.id === proj.id}
										class:highlighted={highlightedProjectIndex === index}
										onclick={() => { selectedProject = proj; showProjectSelector = false; highlightedProjectIndex = -1; }}
										onmouseenter={() => highlightedProjectIndex = index}
									>
										<svg class="option-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
										</svg>
										<span>{proj.name}</span>
									</button>
								{/each}
							{/if}
							<div class="dropdown-divider"></div>
							<button
								class="btn-pill btn-pill-ghost btn-pill-sm btn-pill-block justify-start"
								onclick={() => { showProjectSelector = false; windowStore.openWindow('projects'); }}
							>
								<svg class="option-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<line x1="12" y1="5" x2="12" y2="19"/>
									<line x1="5" y1="12" x2="19" y2="12"/>
								</svg>
								<span>Create New Project</span>
							</button>
						</div>
					{/if}
				</div>

				<!-- Model selector -->
				<div class="context-selector">
					<button
						class="btn-pill btn-pill-sm {selectedModel ? 'btn-pill-secondary' : 'btn-pill-ghost'}"
						onclick={() => showModelSelector = !showModelSelector}
						title="Select AI model"
					>
						{#if selectedModel?.isLocal}
							<svg class="model-icon local" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<rect x="2" y="3" width="20" height="14" rx="2"/>
								<path d="M8 21h8M12 17v4"/>
							</svg>
						{:else}
							<svg class="model-icon cloud" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
							</svg>
						{/if}
						<span class="model-name">{selectedModel?.name?.split(':')[0] || 'Model'}</span>
						<svg class="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M6 9l6 6 6-6"/>
						</svg>
					</button>

					{#if showModelSelector}
						<div class="context-dropdown model-dropdown">
							<div class="dropdown-header">LOCAL MODELS</div>
							{#if loadingModels}
								<div class="context-option empty">Loading...</div>
							{:else if ollamaAvailable === false}
								<div class="context-option empty">
									<svg class="option-model-icon text-amber-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
										<line x1="12" y1="9" x2="12" y2="13"/>
										<line x1="12" y1="17" x2="12.01" y2="17"/>
									</svg>
									Ollama not installed or not running
								</div>
							{:else if localModels.length === 0}
								<div class="context-option empty">No local models installed</div>
							{:else}
								{#each localModels as model}
									<button
										class="context-option"
										class:selected={selectedModel?.id === model.id}
										onclick={() => { selectedModel = { ...model, isLocal: true }; showModelSelector = false; }}
									>
										<svg class="option-model-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<rect x="2" y="3" width="20" height="14" rx="2"/>
											<path d="M8 21h8M12 17v4"/>
										</svg>
										<span>{model.name}</span>
										<span class="model-status ready">Ready</span>
									</button>
								{/each}
							{/if}

							<!-- Recommended models to pull - only show if Ollama is available -->
							{#if ollamaAvailable !== false}
								<div class="dropdown-divider"></div>
								<div class="dropdown-header">AVAILABLE TO PULL</div>
								{#if isPulling && pullProgress}
									<div class="context-option pull-progress">
										<span class="pull-model">{pullingModel}</span>
										<span class="pull-status">
											{pullProgress.status}
											{#if pullProgress.percent !== undefined}
												({pullProgress.percent}%)
											{/if}
										</span>
									</div>
								{/if}
								{#each recommendedModels as model}
									{@const isInstalled = localModels.some(m => m.id === model.id || m.id.startsWith(model.id.split(':')[0]))}
									{#if !isInstalled}
										<button
											class="context-option"
											disabled={isPulling}
											onclick={() => pullModel(model.id)}
										>
											<svg class="option-model-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
												<polyline points="7 10 12 15 17 10"/>
												<line x1="12" y1="15" x2="12" y2="3"/>
											</svg>
											<span>{model.name}</span>
											<span class="model-size">{model.size}</span>
											<span class="model-status pull">Pull</span>
										</button>
									{/if}
								{/each}
							{/if}

							<div class="dropdown-divider"></div>
							<div class="dropdown-header">CLOUD MODELS</div>
							{#if cloudModels.length === 0}
								<div class="context-option empty">
									<svg class="option-model-icon text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
									</svg>
									No API keys configured
								</div>
							{:else}
								{#each cloudModels as model}
									<button
										class="context-option"
										class:selected={selectedModel?.id === model.id}
										onclick={() => { selectedModel = { ...model, isLocal: false }; showModelSelector = false; }}
									>
										<svg class="option-model-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/>
										</svg>
										<span>{model.name}</span>
										<span class="model-status ready">Ready</span>
									</button>
								{/each}
							{/if}

							<div class="dropdown-divider"></div>
							<a href="/settings/ai" class="context-option settings-link" onclick={() => showModelSelector = false}>
								<svg class="option-model-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="3"/>
									<path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z"/>
								</svg>
								<span>AI Settings</span>
							</a>
						</div>
					{/if}
				</div>

				<!-- Quick hints in toolbar -->
				{#if isExpanded && !isRecording}
					<div class="toolbar-hints">
						<span class="hint">⌘D</span>
						<span class="hint-text">Voice</span>
						<span class="hint-divider">•</span>
						<span class="hint">Enter</span>
						<span class="hint-text">Send</span>
					</div>
				{/if}
			</div>

			<!-- Right side - action buttons -->
			<div class="toolbar-right">
				<!-- Voice button -->
				<button
					class="voice-btn"
					class:recording={isRecording}
					onclick={toggleRecording}
					title={isRecording ? 'Stop recording' : 'Voice input (⌘D)'}
				>
					{#if isRecording}
						<svg viewBox="0 0 24 24" fill="currentColor">
							<rect x="6" y="6" width="12" height="12" rx="2"/>
						</svg>
					{:else}
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
							<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
							<line x1="12" y1="19" x2="12" y2="23"/>
							<line x1="8" y1="23" x2="16" y2="23"/>
						</svg>
					{/if}
				</button>

				<!-- Send button -->
				<button
					class="btn-pill btn-pill-icon {chatInput.trim() ? 'btn-pill-primary' : 'btn-pill-ghost'}"
					onclick={chatInput.trim() ? handleChatSubmit : openFullChat}
					disabled={isLoading}
					title={chatInput.trim() ? 'Send (Enter)' : 'Open full chat'}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
						<path d="M12 19V5M5 12l7-7 7 7"/>
					</svg>
				</button>
			</div>
		</div>
	</div>
	{/if}

	<div
		class="dock"
		class:drag-over={isDragOver}
		ondragover={handleDragOver}
		ondragleave={handleDragLeave}
		ondrop={handleDrop}
		role="toolbar"
		aria-label="Application dock"
		tabindex="0"
	>
		{#each dockItems() as item, index (item.id)}
			{#if item.module === 'separator'}
				<div class="dock-separator"></div>
			{:else}
				{@const icon = getItemIcon(item)}
				<button
					class="dock-item style-{iconStyle}"
					class:has-indicator={item.isOpen}
					style="
						transform: scale({getScale(index)}) translateY({getTranslateY(index)}px);
					"
					onmouseenter={() => hoveredIndex = index}
					onmouseleave={() => hoveredIndex = null}
					onclick={() => handleItemClick(item)}
					aria-label={item.label}
				>
					<div
						class="dock-icon"
						class:terminal={icon.isTerminal}
						class:finder={icon.isFinder}
						class:user-app={icon.isUserApp}
						style="
							{icon.isFinder ? `background: ${icon.bgColor};` : `background-color: ${iconStyle === 'minimal' ? 'transparent' : icon.bgColor};`}
							{iconStyle === 'outlined' && !icon.isFinder ? `border: 2px solid ${icon.color}; background-color: transparent;` : ''}
							{iconStyle === 'neon' ? `color: ${icon.color};` : ''}
							{iconStyle === 'gradient' ? `--gradient-start: ${icon.color}; --gradient-end: ${icon.bgColor};` : ''}
						"
					>
						{#if icon.imageUrl}
							<!-- Module or user app with logo image -->
							<img src={icon.imageUrl} alt={item.label} class="dock-icon-image" />
						{:else if icon.isTerminal}
							<span class="terminal-prompt">&gt;_</span>
						{:else if icon.isFinder}
							<!-- Finder happy face icon -->
							<svg class="dock-icon-svg finder-face" viewBox="0 0 24 24" fill="none">
								<!-- Left eye with outline for visibility -->
								<rect x="6" y="7" width="4" height="6" rx="2" stroke="rgba(0, 0, 0, 0.5)" stroke-width="1" fill="white"/>
								<!-- Right eye with outline for visibility -->
								<rect x="14" y="7" width="4" height="6" rx="2" stroke="rgba(0, 0, 0, 0.5)" stroke-width="1" fill="white"/>
								<!-- Smile with dark outline for visibility on any background -->
								<path d="M6 17 C8 20, 16 20, 18 17" stroke="rgba(0, 0, 0, 0.6)" stroke-width="4" stroke-linecap="round" fill="none"/>
								<path d="M6 17 C8 20, 16 20, 18 17" stroke="white" stroke-width="2.5" stroke-linecap="round" fill="none"/>
							</svg>
						{:else if icon.isFolder}
							<!-- Folder icon with fill -->
							<svg
								class="dock-icon-svg"
								viewBox="0 0 24 24"
								fill={icon.color}
								stroke="none"
							>
								<path d={icon.path} />
							</svg>
						{:else}
							<svg
								class="dock-icon-svg library-{iconLibrary}"
								viewBox="0 0 24 24"
								fill="none"
								stroke={icon.color}
								stroke-width={libraryStrokeWidth}
								stroke-linecap={libraryLineCap}
								stroke-linejoin={libraryLineJoin}
								style="
									transform: scale({libraryIconScale});
									filter: {librarySvgFilter};
								"
							>
								<path d={icon.path} />
							</svg>
						{/if}
					</div>

					<!-- Open indicator dot -->
					{#if item.isOpen}
						<div class="dock-indicator"></div>
					{/if}

					<!-- Tooltip -->
					{#if hoveredIndex === index}
						<div class="dock-tooltip">{item.label}</div>
					{/if}
				</button>
			{/if}
		{/each}
	</div>
</div>

<style>
	.dock-container {
		position: fixed;
		bottom: 8px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 9999;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
	}

	/* Quick Chat Input Bar - dynamic width */
	.quick-chat {
		display: flex;
		flex-direction: column;
		background: var(--quick-chat-bg, rgba(255, 255, 255, 0.95));
		backdrop-filter: blur(24px);
		-webkit-backdrop-filter: blur(24px);
		border: 1px solid var(--quick-chat-border, rgba(0, 0, 0, 0.08));
		border-radius: 16px;
		box-shadow: var(--quick-chat-shadow, 0 0 0 0.5px rgba(0, 0, 0, 0.06), 0 8px 32px rgba(0, 0, 0, 0.12));
		overflow: visible;
		transition: min-width 0.2s ease, box-shadow 0.2s ease;
		width: auto;
		min-width: 420px;
		max-width: 700px;
		position: relative;
		z-index: 1;
	}

	:global(.dark) {
		--quick-chat-bg: rgba(28, 28, 30, 0.95);
		--quick-chat-border: rgba(255, 255, 255, 0.12);
		--quick-chat-shadow: 0 0 0 0.5px rgba(255, 255, 255, 0.08), 0 8px 32px rgba(0, 0, 0, 0.4);
	}

	/* Collapsed Chat Bubble - macOS dictation style */
	.collapsed-chat-bubble {
		display: flex;
		align-items: center;
		justify-content: center;
		background: transparent;
		border-radius: 22px;
		padding: 0;
		transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
		position: relative;
		z-index: 1;
	}

	/* Light mode - default also transparent */
	:global(html:not(.dark)) .collapsed-chat-bubble {
		background: transparent;
	}

	/* Only show background when hovering (expanded hint state) */
	.collapsed-chat-bubble.hovering {
		background: rgba(30, 30, 32, 0.95);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		padding: 12px 16px;
		border-radius: 16px;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.25);
	}

	:global(html:not(.dark)) .collapsed-chat-bubble.hovering {
		background: rgba(255, 255, 255, 0.95);
		border: 1px solid rgba(0, 0, 0, 0.12);
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
	}

	.collapsed-chat-bubble.voice-active {
		background: transparent;
		padding: 0;
		box-shadow: none;
	}

	:global(html:not(.dark)) .collapsed-chat-bubble.voice-active {
		background: transparent;
	}

	/* Hover state - macOS dictation style */
	.collapsed-hover-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
	}

	.collapsed-hint-bubble {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 13px;
		color: rgba(255, 255, 255, 0.95);
		font-weight: 500;
		white-space: nowrap;
	}

	:global(html:not(.dark)) .collapsed-hint-bubble {
		color: rgba(0, 0, 0, 0.85);
	}

	.collapsed-hint-bubble kbd {
		display: inline-flex;
		align-items: center;
		padding: 2px 6px;
		background: rgba(255, 255, 255, 0.15);
		border-radius: 4px;
		font-family: -apple-system, BlinkMacSystemFont, sans-serif;
		font-size: 12px;
		font-weight: 600;
	}

	:global(html:not(.dark)) .collapsed-hint-bubble kbd {
		background: rgba(0, 0, 0, 0.08);
	}

	.collapsed-dots {
		color: rgba(255, 255, 255, 0.5);
		font-size: 11px;
		letter-spacing: 2px;
	}

	:global(html:not(.dark)) .collapsed-dots {
		color: rgba(0, 0, 0, 0.4);
	}

	.collapsed-dots-default {
		font-size: 16px;
		letter-spacing: 3px;
		font-weight: 700;
		color: rgba(255, 255, 255, 0.5);
	}

	:global(html:not(.dark)) .collapsed-dots-default {
		color: rgba(0, 0, 0, 0.35);
	}

	/* Voice recording in collapsed mode - matches example design */
	.collapsed-voice-recording {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 6px 8px 6px 10px;
		background: rgba(30, 30, 30, 0.95);
		border-radius: 24px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
	}

	:global(html:not(.dark)) .collapsed-voice-recording {
		background: rgba(255, 255, 255, 0.98);
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
	}


	.collapsed-waveform-container {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.collapsed-waveform {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 3px;
		height: 24px;
		min-width: 60px;
	}

	.collapsed-timer {
		font-size: 13px;
		font-family: ui-monospace, 'SF Mono', monospace;
		font-weight: 500;
		color: rgba(255, 255, 255, 0.9);
		min-width: 40px;
		letter-spacing: 0.5px;
	}

	:global(html:not(.dark)) .collapsed-timer {
		color: rgba(0, 0, 0, 0.8);
	}

	.collapsed-bar {
		width: 3px;
		background: rgba(255, 255, 255, 0.85);
		border-radius: 2px;
		transition: height 0.08s ease-out;
		animation: waveform-pulse 0.5s ease-in-out infinite alternate;
	}

	.collapsed-bar:nth-child(odd) {
		animation-delay: 0.1s;
	}

	.collapsed-bar:nth-child(3n) {
		animation-delay: 0.2s;
	}

	@keyframes waveform-pulse {
		from { opacity: 0.7; }
		to { opacity: 1; }
	}

	:global(html:not(.dark)) .collapsed-bar {
		background: rgba(0, 0, 0, 0.7);
	}

	.collapsed-duration {
		font-size: 12px;
		font-family: ui-monospace, monospace;
		color: rgba(255, 255, 255, 0.9);
		min-width: 36px;
		text-align: center;
	}

	:global(html:not(.dark)) .collapsed-duration {
		color: rgba(0, 0, 0, 0.8);
	}


	.quick-chat.expanded {
		min-width: 460px;
		box-shadow:
			0 0 0 0.5px rgba(0, 0, 0, 0.08),
			0 12px 40px rgba(0, 0, 0, 0.15);
	}

	/* Dynamic width based on content */
	.quick-chat.size-small-expand {
		min-width: 500px;
	}

	.quick-chat.size-medium {
		min-width: 600px;
	}

	.quick-chat.size-large {
		min-width: 720px;
		max-width: 800px;
	}

	.quick-chat.dragging-file {
		border-color: #3B82F6;
		background: rgba(239, 246, 255, 0.95);
		box-shadow:
			0 0 0 2px rgba(59, 130, 246, 0.3),
			0 12px 40px rgba(0, 0, 0, 0.15);
	}

	/* Focus state - no scale to avoid blur */
	.quick-chat:focus-within {
		box-shadow:
			0 0 0 0.5px rgba(0, 0, 0, 0.1),
			0 0 0 3px rgba(59, 130, 246, 0.15),
			0 12px 40px rgba(0, 0, 0, 0.12);
	}

	/* Textarea wrapper (top section) */
	.quick-chat-textarea-wrapper {
		padding: 12px 14px 8px 14px;
	}

	/* Textarea input */
	.quick-chat textarea {
		width: 100%;
		border: none;
		background: transparent;
		font-size: 14px;
		font-family: inherit;
		outline: none;
		color: var(--quick-chat-text, #333);
		resize: none;
		line-height: 1.5;
		min-height: 24px;
		max-height: 150px;
		height: 24px;
		overflow-y: auto;
		overflow-x: hidden;
		transition: height 0.15s ease;
		padding: 0;
		/* Hide scrollbar but keep functionality */
		scrollbar-width: none; /* Firefox */
		-ms-overflow-style: none; /* IE/Edge */
	}

	:global(.dark) {
		--quick-chat-text: #f5f5f7;
		--quick-chat-placeholder: #6e6e73;
		--quick-chat-icon: #a1a1a6;
		--quick-chat-icon-hover: #f5f5f7;
		--quick-chat-btn-bg: #2c2c2e;
		--quick-chat-btn-hover: #3a3a3c;
	}

	/* Remove focus outline in dark mode */
	:global(.dark) .quick-chat textarea:focus {
		outline: none !important;
		box-shadow: none !important;
	}

	:global(.dark) .quick-chat *:focus {
		outline: none !important;
	}

	.quick-chat textarea::-webkit-scrollbar {
		display: none; /* Chrome/Safari/Opera */
	}

	.quick-chat textarea::placeholder {
		color: var(--quick-chat-placeholder, #999);
	}

	/* Bottom toolbar (fixed at bottom) - no border */
	.quick-chat-toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 14px 10px 14px;
	}

	.toolbar-left {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.toolbar-right {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	/* Toolbar hints */
	.toolbar-hints {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 11px;
		color: #999;
		margin-left: 8px;
	}

	.toolbar-hints .hint {
		background: rgba(0, 0, 0, 0.05);
		padding: 2px 5px;
		border-radius: 3px;
		font-family: ui-monospace, monospace;
		font-size: 10px;
	}

	.toolbar-hints .hint-text {
		color: #aaa;
	}

	.toolbar-hints .hint-divider {
		color: #ddd;
		margin: 0 2px;
	}

	/* Voice button - minimal, no background */
	.quick-chat .voice-btn {
		width: 32px;
		height: 32px;
		border: none;
		background: transparent;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #888;
		transition: all 0.2s;
		flex-shrink: 0;
	}

	.quick-chat .voice-btn:hover {
		color: #333;
	}

	.quick-chat .voice-btn.recording {
		background: #EF4444;
		color: white;
		animation: pulse 1.5s infinite;
	}

	.quick-chat .voice-btn svg {
		width: 18px;
		height: 18px;
	}

	@keyframes pulse {
		0%, 100% { transform: scale(1); }
		50% { transform: scale(1.05); }
	}

	/* Hidden file input */
	.hidden-file-input {
		display: none;
	}

	/* Attached files display */
	.attached-files {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		padding: 8px 14px 0 14px;
	}

	.attached-file {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 4px 8px;
		background: #F3F4F6;
		border-radius: 6px;
		font-size: 12px;
		color: #374151;
	}

	.attached-file .file-icon {
		width: 14px;
		height: 14px;
		color: #6B7280;
	}

	.attached-file .file-name {
		max-width: 120px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}



	/* Context selector */
	.context-selector {
		position: relative;
		z-index: 10000;
	}

	.required {
		animation: pulse-required 2s infinite;
	}

	@keyframes pulse-required {
		0%, 100% { border-color: #FCA5A5; }
		50% { border-color: #EF4444; }
	}

	.project-icon {
		color: #8B5CF6;
	}

	.placeholder-text {
		color: #9CA3AF;
	}

	.model-icon {
		width: 14px;
		height: 14px;
	}

	.model-icon.local {
		color: #10B981;
	}

	.model-icon.cloud {
		color: #3B82F6;
	}

	.model-name {
		max-width: 80px;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.chevron {
		width: 12px;
		height: 12px;
		opacity: 0.5;
	}

	/* Dropdown enhancements */
	.dropdown-header {
		padding: 8px 12px 4px;
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: #9CA3AF;
	}

	.dropdown-divider {
		height: 1px;
		background: #E5E7EB;
		margin: 4px 0;
	}


	.option-model-icon {
		width: 14px;
		height: 14px;
		flex-shrink: 0;
	}

	.model-dropdown {
		min-width: 240px;
	}

	.model-status {
		font-size: 10px;
		padding: 2px 6px;
		border-radius: 4px;
		font-weight: 500;
		margin-left: auto;
	}

	.model-status.ready {
		background: #D1FAE5;
		color: #059669;
	}

	.model-status.pull {
		background: #DBEAFE;
		color: #2563EB;
	}

	.model-size {
		font-size: 10px;
		color: #9CA3AF;
		margin-left: 4px;
	}

	.pull-progress {
		flex-direction: column !important;
		align-items: flex-start !important;
		gap: 4px !important;
		padding: 12px !important;
		background: #F3F4F6;
	}

	.pull-model {
		font-weight: 500;
		color: #374151;
	}

	.pull-status {
		font-size: 11px;
		color: #2563EB;
	}

	.settings-link {
		color: #6B7280 !important;
		text-decoration: none;
	}

	.settings-link:hover {
		background: #F3F4F6;
	}


	.context-name {
		max-width: 80px;
		overflow: hidden;
		text-overflow: ellipsis;
		font-weight: 500;
	}

	.context-dropdown {
		position: absolute;
		bottom: 100%;
		left: 0;
		margin-bottom: 8px;
		background: white;
		border: 1px solid #E5E7EB;
		border-radius: 12px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
		min-width: 220px;
		max-height: 280px;
		overflow-y: auto;
		z-index: 10002;
		animation: dropdownSlideUp 0.2s cubic-bezier(0.16, 1, 0.3, 1);
		transform-origin: bottom center;
	}

	@keyframes dropdownSlideUp {
		from {
			opacity: 0;
			transform: translateY(8px) scale(0.96);
		}
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}

	.option-icon {
		width: 14px;
		height: 14px;
		flex-shrink: 0;
	}

	.selected {
		background: #EEF2FF;
		color: #4F46E5;
	}

	.highlighted {
		background: #E5E7EB;
	}

	.highlighted.selected {
		background: #DDD6FE;
	}

	.justify-start {
		justify-content: flex-start;
	}

	.option-icon {
		font-size: 14px;
	}

	/* Recording waveform bar - matches popup-chat style */
	.recording-waveform-bar {
		display: flex;
		align-items: center;
		gap: 8px;
		background: #1f2937;
		border-radius: 24px;
		padding: 8px 12px;
		margin: 8px 10px;
	}

	.recording-waveform {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 2px;
		height: 24px;
	}

	.recording-bar {
		width: 2px;
		background: white;
		border-radius: 1px;
		transition: height 0.05s ease-out;
		min-height: 2px;
	}

	.recording-duration {
		font-size: 12px;
		font-family: ui-monospace, 'SF Mono', monospace;
		color: white;
		min-width: 32px;
		text-align: right;
	}


	/* Response display */
	.quick-chat-response {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px 16px;
		background: rgba(249, 250, 251, 0.9);
		border-bottom: 1px solid rgba(0, 0, 0, 0.05);
	}

	.response-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.ai-badge {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 11px;
		font-weight: 600;
		color: #666;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.ai-badge svg {
		width: 14px;
		height: 14px;
		color: #8B5CF6;
	}

	.response-actions {
		display: flex;
		gap: 4px;
	}


	.response-content {
		font-size: 13px;
		line-height: 1.6;
		color: #333;
		max-height: 150px;
		overflow-y: auto;
	}

	/* Quick hints */
	.quick-hints {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		padding: 6px 12px;
		background: rgba(0, 0, 0, 0.02);
		border-top: 1px solid rgba(0, 0, 0, 0.04);
		font-size: 11px;
	}

	.hint {
		padding: 2px 6px;
		background: rgba(0, 0, 0, 0.06);
		border-radius: 4px;
		font-weight: 500;
		color: #666;
		font-family: system-ui, -apple-system, sans-serif;
	}

	.hint-text {
		color: #999;
	}

	.hint-divider {
		color: #ddd;
	}

	/* Loading dots animation */
	.loading-dots {
		display: flex;
		gap: 4px;
	}

	.loading-dots span {
		width: 6px;
		height: 6px;
		background: #8B5CF6;
		border-radius: 50%;
		animation: loadingBounce 1.4s infinite ease-in-out both;
	}

	.loading-dots span:nth-child(1) { animation-delay: -0.32s; }
	.loading-dots span:nth-child(2) { animation-delay: -0.16s; }

	@keyframes loadingBounce {
		0%, 80%, 100% { transform: scale(0.8); opacity: 0.5; }
		40% { transform: scale(1); opacity: 1; }
	}

	.dock {
		display: flex;
		align-items: flex-end;
		gap: 4px;
		padding: 6px 10px 8px;
		background: rgba(255, 255, 255, 0.75);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.6);
		border-radius: 16px;
		box-shadow:
			0 0 0 0.5px rgba(0, 0, 0, 0.1),
			0 8px 32px rgba(0, 0, 0, 0.12);
		transition: all 0.2s ease;
		position: relative;
		z-index: 10;
	}

	.dock.drag-over {
		background: rgba(0, 102, 255, 0.15);
		border-color: rgba(0, 102, 255, 0.5);
		box-shadow:
			0 0 0 2px rgba(0, 102, 255, 0.3),
			0 8px 32px rgba(0, 0, 0, 0.15);
	}

	.dock-separator {
		width: 1px;
		height: 40px;
		background: rgba(0, 0, 0, 0.15);
		margin: 0 4px;
		align-self: center;
	}

	.dock-item {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		transition: transform 0.15s ease-out;
		transform-origin: bottom center;
	}

	.dock-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: box-shadow 0.15s ease;
	}

	.dock-item:hover .dock-icon {
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
	}

	.dock-icon-svg {
		width: 24px;
		height: 24px;
	}

	.dock-icon-image {
		width: 100%;
		height: 100%;
		object-fit: contain;
		padding: 4px;
	}

	.dock-icon.user-app {
		background: white;
		padding: 0;
		overflow: hidden;
	}

	.dock-icon.terminal {
		background: #1E1E1E !important;
	}

	.dock-icon.finder {
		border-radius: 10px;
		box-shadow:
			0 2px 8px rgba(0, 0, 0, 0.15),
			inset 0 1px 0 rgba(255, 255, 255, 0.2);
	}

	.finder-face {
		width: 32px;
		height: 32px;
	}

	.terminal-prompt {
		font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Code', 'Courier New', monospace;
		font-size: 16px;
		font-weight: bold;
		color: #00FF00;
		text-shadow: 0 0 8px rgba(0, 255, 0, 0.5);
	}

	.dock-indicator {
		width: 4px;
		height: 4px;
		border-radius: 50%;
		background: #333;
		margin-top: 4px;
	}

	.dock-tooltip {
		position: absolute;
		bottom: 100%;
		left: 50%;
		transform: translateX(-50%);
		margin-bottom: 8px;
		padding: 6px 12px;
		background: rgba(30, 30, 30, 0.95);
		color: white;
		font-size: 12px;
		font-weight: 500;
		border-radius: 6px;
		white-space: nowrap;
		pointer-events: none;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
		z-index: 10010;
	}

	.dock-tooltip::after {
		content: '';
		position: absolute;
		top: 100%;
		left: 50%;
		transform: translateX(-50%);
		border: 6px solid transparent;
		border-top-color: rgba(30, 30, 30, 0.9);
	}

	/* Icon Style Variants for Dock */

	/* Minimal */
	.dock-item.style-minimal .dock-icon {
		box-shadow: none;
		background: transparent !important;
	}

	.dock-item.style-minimal:hover .dock-icon {
		background: rgba(0, 0, 0, 0.08) !important;
	}

	/* Rounded */
	.dock-item.style-rounded .dock-icon {
		border-radius: 50%;
	}

	/* Square */
	.dock-item.style-square .dock-icon {
		border-radius: 4px;
	}

	/* macOS */
	.dock-item.style-macos .dock-icon {
		border-radius: 22%;
		width: 52px;
		height: 52px;
	}

	.dock-item.style-macos .dock-icon-svg {
		width: 28px;
		height: 28px;
	}

	/* Outlined */
	.dock-item.style-outlined .dock-icon {
		box-shadow: none;
	}

	.dock-item.style-outlined:hover .dock-icon {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	/* Retro */
	.dock-item.style-retro .dock-icon {
		border-radius: 0;
		box-shadow:
			3px 3px 0 rgba(0, 0, 0, 0.3),
			inset -2px -2px 0 rgba(0, 0, 0, 0.2),
			inset 2px 2px 0 rgba(255, 255, 255, 0.3);
	}

	/* Win95 */
	.dock-item.style-win95 .dock-icon {
		border-radius: 0;
		box-shadow: none;
		border: 2px solid;
		border-color: #DFDFDF #808080 #808080 #DFDFDF;
		background: #C0C0C0 !important;
	}

	.dock-item.style-win95:hover .dock-icon {
		border-color: #808080 #DFDFDF #DFDFDF #808080;
	}

	/* Glassmorphism */
	.dock-item.style-glassmorphism .dock-icon {
		background: rgba(255, 255, 255, 0.2) !important;
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		border: 1px solid rgba(255, 255, 255, 0.3);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
	}

	.dock-item.style-glassmorphism:hover .dock-icon {
		background: rgba(255, 255, 255, 0.3) !important;
	}

	/* Neon */
	.dock-item.style-neon .dock-icon {
		background: #1a1a2e !important;
		box-shadow:
			0 0 8px currentColor,
			0 0 16px currentColor,
			inset 0 0 8px rgba(255, 255, 255, 0.1);
		border: 1px solid currentColor;
	}

	.dock-item.style-neon:hover .dock-icon {
		box-shadow:
			0 0 12px currentColor,
			0 0 24px currentColor,
			0 0 36px currentColor,
			inset 0 0 8px rgba(255, 255, 255, 0.1);
	}

	.dock-item.style-neon .dock-icon-svg {
		filter: drop-shadow(0 0 3px currentColor);
	}

	/* Flat */
	.dock-item.style-flat .dock-icon {
		box-shadow: none;
		border-radius: 8px;
	}

	.dock-item.style-flat:hover .dock-icon {
		filter: brightness(0.95);
	}

	/* Gradient */
	.dock-item.style-gradient .dock-icon {
		background: linear-gradient(135deg, var(--gradient-start, #667eea) 0%, var(--gradient-end, #764ba2) 100%) !important;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
	}

	.dock-item.style-gradient .dock-icon-svg {
		stroke: white !important;
		filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.2));
	}

	.dock-item.style-gradient:hover .dock-icon {
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.25);
	}

	/* macOS Classic */
	.dock-item.style-macos-classic .dock-icon {
		border-radius: 4px;
		background: linear-gradient(180deg, #EAEAEA 0%, #D4D4D4 50%, #C4C4C4 100%) !important;
		border: 1px solid;
		border-color: #FFFFFF #888888 #888888 #FFFFFF;
		box-shadow:
			1px 1px 0 #666666,
			inset 1px 1px 0 rgba(255, 255, 255, 0.8);
	}

	.dock-item.style-macos-classic:hover .dock-icon {
		background: linear-gradient(180deg, #F0F0F0 0%, #E0E0E0 50%, #D0D0D0 100%) !important;
	}

	/* Paper */
	.dock-item.style-paper .dock-icon {
		background: #FFFFFF !important;
		border-radius: 8px;
		box-shadow:
			0 1px 3px rgba(0, 0, 0, 0.08),
			0 4px 12px rgba(0, 0, 0, 0.05);
		border: 1px solid rgba(0, 0, 0, 0.06);
	}

	.dock-item.style-paper:hover .dock-icon {
		box-shadow:
			0 2px 8px rgba(0, 0, 0, 0.1),
			0 8px 24px rgba(0, 0, 0, 0.08);
		transform: translateY(-2px);
	}

	/* Pixel */
	.dock-item.style-pixel .dock-icon {
		border-radius: 0;
		image-rendering: pixelated;
		box-shadow:
			3px 0 0 #000,
			-3px 0 0 #000,
			0 3px 0 #000,
			0 -3px 0 #000;
		border: none;
	}

	.dock-item.style-pixel:hover .dock-icon {
		box-shadow:
			3px 0 0 #333,
			-3px 0 0 #333,
			0 3px 0 #333,
			0 -3px 0 #333;
		filter: brightness(1.1);
	}

	/* Frosted - clean frosted glass */
	.dock-item.style-frosted .dock-icon {
		background: rgba(255, 255, 255, 0.6) !important;
		backdrop-filter: blur(12px) saturate(180%);
		-webkit-backdrop-filter: blur(12px) saturate(180%);
		border-radius: 14px;
		border: 1px solid rgba(255, 255, 255, 0.4);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
	}

	.dock-item.style-frosted:hover .dock-icon {
		background: rgba(255, 255, 255, 0.75) !important;
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12);
	}

	/* Terminal */
	.dock-item.style-terminal .dock-icon {
		background: #0a0a0a !important;
		border-radius: 4px;
		border: 1px solid #00ff00;
		box-shadow: 0 0 8px rgba(0, 255, 0, 0.3);
	}

	.dock-item.style-terminal:hover .dock-icon {
		box-shadow: 0 0 12px rgba(0, 255, 0, 0.5);
	}

	.dock-item.style-terminal .dock-icon-svg {
		color: #00ff00 !important;
		filter: drop-shadow(0 0 2px #00ff00);
	}

	/* Glow - soft colored glow */
	.dock-item.style-glow .dock-icon {
		border-radius: 14px;
		box-shadow:
			0 0 15px currentColor,
			0 0 30px rgba(100, 100, 255, 0.3);
		border: none;
	}

	.dock-item.style-glow:hover .dock-icon {
		box-shadow:
			0 0 20px currentColor,
			0 0 40px rgba(100, 100, 255, 0.4);
		transform: scale(1.02);
	}

	.dock-item.style-glow .dock-icon-svg {
		filter: drop-shadow(0 0 3px currentColor);
	}

	/* Brutalist */
	.dock-item.style-brutalist .dock-icon {
		background: #fff !important;
		border-radius: 0;
		border: 3px solid #000;
		box-shadow: 4px 4px 0 #000;
	}

	.dock-item.style-brutalist:hover .dock-icon {
		transform: translate(-2px, -2px);
		box-shadow: 6px 6px 0 #000;
	}

	.dock-item.style-brutalist .dock-icon-svg {
		color: #000 !important;
	}

	/* Depth - layered 3D shadows */
	.dock-item.style-depth .dock-icon {
		border-radius: 12px;
		border: none;
		box-shadow:
			0 2px 4px rgba(0, 0, 0, 0.1),
			0 4px 8px rgba(0, 0, 0, 0.1),
			0 8px 16px rgba(0, 0, 0, 0.08);
	}

	.dock-item.style-depth:hover .dock-icon {
		transform: translateY(-3px);
		box-shadow:
			0 4px 8px rgba(0, 0, 0, 0.12),
			0 8px 16px rgba(0, 0, 0, 0.1),
			0 16px 32px rgba(0, 0, 0, 0.08);
	}

	/* ===== DARK MODE - MODERN APPLE STYLE ===== */
	:global(.dark) .dock {
		background: rgba(44, 44, 46, 0.85);
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow:
			0 0 0 0.5px rgba(255, 255, 255, 0.08),
			0 8px 32px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .dock.drag-over {
		background: rgba(59, 130, 246, 0.15);
		border-color: rgba(59, 130, 246, 0.5);
	}

	:global(.dark) .dock-separator {
		background: rgba(255, 255, 255, 0.15);
	}

	:global(.dark) .dock-indicator {
		background: #fff;
	}

	:global(.dark) .dock-tooltip {
		background: rgba(44, 44, 46, 0.98);
		border: 1px solid rgba(255, 255, 255, 0.12);
	}

	:global(.dark) .dock-tooltip::after {
		border-top-color: rgba(44, 44, 46, 0.98);
	}

	:global(.dark) .quick-chat {
		background: rgba(28, 28, 30, 0.95) !important;
		border-color: rgba(255, 255, 255, 0.12) !important;
		box-shadow:
			0 0 0 0.5px rgba(255, 255, 255, 0.08),
			0 8px 32px rgba(0, 0, 0, 0.4) !important;
	}

	:global(.dark) .quick-chat:focus-within {
		box-shadow:
			0 0 0 0.5px rgba(255, 255, 255, 0.15),
			0 12px 40px rgba(0, 0, 0, 0.4) !important;
		outline: none !important;
		border-color: rgba(255, 255, 255, 0.2) !important;
	}

	:global(.dark) .quick-chat.dragging-file {
		background: rgba(30, 60, 90, 0.95) !important;
		border-color: #3B82F6 !important;
	}

	:global(.dark) .quick-chat.expanded {
		background: rgba(28, 28, 30, 0.95) !important;
	}

	:global(.dark) .quick-chat textarea {
		color: #f5f5f7;
	}

	:global(.dark) .quick-chat textarea::placeholder {
		color: #6e6e73;
	}

	:global(.dark) .quick-chat .voice-btn {
		color: #a1a1a6;
	}

	:global(.dark) .quick-chat .voice-btn:hover {
		color: #f5f5f7;
	}


	:global(.dark) .context-dropdown {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
	}


	:global(.dark) .dropdown-header {
		color: #6e6e73;
	}

	:global(.dark) .dropdown-divider {
		background: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .attached-file {
		background: #3a3a3c;
		color: #f5f5f7;
	}

	:global(.dark) .attached-file .file-icon {
		color: #a1a1a6;
	}


	:global(.dark) .toolbar-hints .hint {
		background: rgba(255, 255, 255, 0.1);
		color: #a1a1a6;
	}

	:global(.dark) .toolbar-hints .hint-text {
		color: #6e6e73;
	}

	:global(.dark) .toolbar-hints .hint-divider {
		color: #48484a;
	}

	:global(.dark) .recording-waveform-bar {
		background: #1c1c1e;
		border: 1px solid rgba(255, 255, 255, 0.12);
	}

	:global(.dark) .quick-chat-response {
		background: rgba(44, 44, 46, 0.95);
		border-bottom-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .response-content {
		color: #f5f5f7;
	}

	:global(.dark) .ai-badge {
		color: #a1a1a6;
	}


	:global(.dark) .model-status.ready {
		background: rgba(16, 185, 129, 0.2);
		color: #34d399;
	}

	:global(.dark) .model-status.pull {
		background: rgba(10, 132, 255, 0.2);
		color: #60a5fa;
	}

	:global(.dark) .pull-progress {
		background: #3a3a3c;
	}

	:global(.dark) .pull-model {
		color: #f5f5f7;
	}

	/* Dark mode icon styles */
	:global(.dark) .dock-item.style-glassmorphism .dock-icon {
		background: rgba(255, 255, 255, 0.1) !important;
		border-color: rgba(255, 255, 255, 0.15);
	}

	:global(.dark) .dock-item.style-glassmorphism:hover .dock-icon {
		background: rgba(255, 255, 255, 0.15) !important;
	}

	:global(.dark) .dock-item.style-minimal:hover .dock-icon {
		background: rgba(255, 255, 255, 0.1) !important;
	}

	:global(.dark) .dock-item.style-paper .dock-icon {
		background: #3a3a3c !important;
		border-color: rgba(255, 255, 255, 0.12);
	}
</style>
