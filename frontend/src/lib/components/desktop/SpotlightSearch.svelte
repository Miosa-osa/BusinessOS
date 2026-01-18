<script lang="ts">
	import { windowStore } from '$lib/stores/windowStore';
	import { fade, scale, fly } from 'svelte/transition';
	import { apiClient } from '$lib/api';
	import { ImageSearchModal } from '$lib/components/search';

	interface Props {
		open: boolean;
		onClose: () => void;
	}

	let { open, onClose }: Props = $props();

	let inputValue = $state('');
	let selectedIndex = $state(0);
	let inputElement: HTMLTextAreaElement | undefined = $state(undefined);
	let mode = $state<'search' | 'chat'>('chat');

	// Voice recording state
	let isRecording = $state(false);
	let mediaRecorder: MediaRecorder | null = null;
	let audioContext: AudioContext | null = null;
	let analyser: AnalyserNode | null = null;
	let waveformBars = $state<number[]>(Array(30).fill(2));
	let animationId: number | null = null;
	let recordingDuration = $state(0);
	let recordingInterval: ReturnType<typeof setInterval> | null = null;
	let liveTranscript = $state('');
	let speechRecognition: SpeechRecognition | null = null;

	// Chat state
	let messages = $state<Array<{ role: 'user' | 'assistant'; content: string }>>([]);
	let isTyping = $state(false);

	// File attachment state
	interface AttachedFile {
		id: string;
		name: string;
		type: string;
		size: number;
		preview?: string; // base64 data URL for images
		file: File;
	}
	let attachedFiles = $state<AttachedFile[]>([]);
	let fileInputRef: HTMLInputElement | undefined = $state(undefined);
	let isDragging = $state(false);

	// Slash commands state
	let showCommandsDropdown = $state(false);
	let commandDropdownIndex = $state(0);
	const slashCommands = [
		{ id: 'analyze', name: '/analyze', description: 'Analyze content or data', icon: '📊' },
		{ id: 'summarize', name: '/summarize', description: 'Summarize text or document', icon: '📝' },
		{ id: 'explain', name: '/explain', description: 'Explain code or concept', icon: '💡' },
		{ id: 'generate', name: '/generate', description: 'Generate content or code', icon: '✨' },
		{ id: 'review', name: '/review', description: 'Review and provide feedback', icon: '🔍' },
		{ id: 'translate', name: '/translate', description: 'Translate to another language', icon: '🌐' },
		{ id: 'brainstorm', name: '/brainstorm', description: 'Generate ideas', icon: '🧠' },
		{ id: 'task', name: '/task', description: 'Create a new task', icon: '✅' },
		{ id: 'image', name: '/image', description: 'Multimodal image search', icon: '🖼️' },
	];

	// Project/Context state
	let selectedProjectId = $state<string | null>(null);
	let showProjectDropdown = $state(false);
	let projectsList = $state<{ id: string; name: string; description?: string }[]>([]);
	let projectDropdownIndex = $state(0);

	// Model state
	let selectedModel = $state('');
	let showModelDropdown = $state(false);
	let availableModels = $state<{ id: string; name: string; provider: string; size?: string }[]>([]);
	let activeProvider = $state('ollama_local');

	// Image search state
	let showImageSearch = $state(false);

	// Derived: Filter commands based on input
	let filteredCommands = $derived(() => {
		if (!inputValue.startsWith('/')) return [];
		const query = inputValue.slice(1).toLowerCase();
		return slashCommands.filter(cmd =>
			cmd.id.includes(query) || cmd.description.toLowerCase().includes(query)
		);
	});

	// All searchable items
	const searchItems = [
		{ id: 'dashboard', type: 'app', name: 'Dashboard', description: 'Overview and analytics', icon: 'M4 5a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v2a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1h-4a1 1 0 01-1-1v-5zm-10 1a1 1 0 011-1h4a1 1 0 011 1v3a1 1 0 01-1 1H5a1 1 0 01-1-1v-3z', color: '#1E88E5' },
		{ id: 'chat', type: 'app', name: 'Chat', description: 'AI Assistant', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z', color: '#43A047' },
		{ id: 'tasks', type: 'app', name: 'Tasks', description: 'Task management', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4', color: '#FB8C00' },
		{ id: 'projects', type: 'app', name: 'Projects', description: 'Project management', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', color: '#8E24AA' },
		{ id: 'team', type: 'app', name: 'Team', description: 'Team members', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z', color: '#00ACC1' },
		{ id: 'clients', type: 'app', name: 'Clients', description: 'Client management', icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4', color: '#7B1FA2' },
		{ id: 'calendar', type: 'app', name: 'Calendar', description: 'Schedule and events', icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z', color: '#E53935' },
		{ id: 'contexts', type: 'app', name: 'Contexts', description: 'Work contexts', icon: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10', color: '#5E35B1' },
		{ id: 'nodes', type: 'app', name: 'Nodes', description: 'Node management', icon: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z', color: '#E53935' },
		{ id: 'settings', type: 'app', name: 'Settings', description: 'System settings', icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z', color: '#546E7A' },
		{ id: 'terminal', type: 'app', name: 'Terminal', description: 'OS Agent terminal', icon: 'M4 17l6-6-6-6M12 19h8', color: '#00FF00' },
		{ id: 'files', type: 'app', name: 'Files', description: 'File browser', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', color: '#3B82F6' },
	];

	// Derived state
	let selectedProject = $derived(
		selectedProjectId ? projectsList.find(p => p.id === selectedProjectId) : null
	);

	let currentModelName = $derived(() => {
		if (!selectedModel) return 'Select Model';
		const model = availableModels.find(m => m.id === selectedModel);
		return model ? model.name : selectedModel.split(':')[0];
	});

	// Filter items based on input
	const filteredItems = $derived(() => {
		if (mode === 'chat') return [];
		if (!inputValue.trim()) {
			return searchItems.slice(0, 6);
		}
		const query = inputValue.toLowerCase();
		return searchItems.filter(item =>
			item.name.toLowerCase().includes(query) ||
			item.description.toLowerCase().includes(query)
		);
	});

	// Format recording duration as MM:SS
	function formatDuration(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	// Conversation ID for persisting to chat
	let conversationId = $state<string | null>(null);

	// Load projects and models on open
	$effect(() => {
		if (open) {
			loadProjects();
			loadModels();
			setTimeout(() => inputElement?.focus(), 50);
		}
		if (!open) {
			inputValue = '';
			selectedIndex = 0;
			messages = [];
			attachedFiles = [];
			conversationId = null;
			stopRecording();
			showProjectDropdown = false;
			showModelDropdown = false;
			showCommandsDropdown = false;
		}
	});

	// Show commands dropdown when typing /
	$effect(() => {
		if (inputValue.startsWith('/') && filteredCommands().length > 0) {
			showCommandsDropdown = true;
			commandDropdownIndex = 0;
		} else {
			showCommandsDropdown = false;
		}
	});

	async function loadProjects() {
		try {
			const response = await apiClient.get('/projects');
			if (response.ok) {
				const data = await response.json();
				projectsList = data.projects || data || [];
			}
		} catch (e) {
			console.error('Failed to load projects:', e);
		}
	}

	async function loadModels() {
		try {
			// Get provider info
			const providersRes = await apiClient.get('/ai/providers');
			if (providersRes.ok) {
				const data = await providersRes.json();
				activeProvider = data.active_provider || 'ollama_local';
				if (data.default_model && !selectedModel) {
					selectedModel = data.default_model;
				}
			}

			// Get available models
			const response = await apiClient.get('/ai/models');
			if (response.ok) {
				const data = await response.json();
				availableModels = data.models || [];
				if (!selectedModel && availableModels.length > 0) {
					selectedModel = availableModels[0].id;
				}
			}
		} catch (e) {
			console.error('Failed to load models:', e);
		}
	}

	// Voice recording functions
	async function startRecording() {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });

			audioContext = new AudioContext();
			analyser = audioContext.createAnalyser();
			analyser.fftSize = 64;

			const source = audioContext.createMediaStreamSource(stream);
			source.connect(analyser);

			mediaRecorder = new MediaRecorder(stream);
			const chunks: BlobPart[] = [];

			mediaRecorder.ondataavailable = (e) => chunks.push(e.data);
			mediaRecorder.onstop = async () => {
				const blob = new Blob(chunks, { type: 'audio/webm' });
				// Transcribe audio
				await transcribeAudio(blob);
			};

			mediaRecorder.start();
			isRecording = true;
			recordingDuration = 0;

			// Start duration timer
			recordingInterval = setInterval(() => {
				recordingDuration++;
			}, 1000);

			// Start Web Speech API for live transcript
			const SpeechRecognitionAPI = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
			if (SpeechRecognitionAPI) {
				const recognition = new SpeechRecognitionAPI() as SpeechRecognition;
				recognition.continuous = true;
				recognition.interimResults = true;
				recognition.lang = 'en-US';

				recognition.onresult = (event: SpeechRecognitionEvent) => {
					let interimTranscript = '';
					let finalTranscript = '';
					for (let i = event.resultIndex; i < event.results.length; i++) {
						const transcript = event.results[i][0].transcript;
						if (event.results[i].isFinal) {
							finalTranscript += transcript;
						} else {
							interimTranscript += transcript;
						}
					}
					liveTranscript = finalTranscript || interimTranscript;
				};

				recognition.start();
				speechRecognition = recognition;
			}

			updateWaveform();
		} catch (err) {
			console.error('Failed to start recording:', err);
		}
	}

	function stopRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.stop();
			mediaRecorder.stream.getTracks().forEach(track => track.stop());
		}
		if (animationId) {
			cancelAnimationFrame(animationId);
			animationId = null;
		}
		if (audioContext) {
			audioContext.close();
			audioContext = null;
		}
		if (recordingInterval) {
			clearInterval(recordingInterval);
			recordingInterval = null;
		}
		if (speechRecognition) {
			speechRecognition.stop();
			speechRecognition = null;
		}
		isRecording = false;
		waveformBars = Array(30).fill(2);
		liveTranscript = '';
		recordingDuration = 0;
	}

	function cancelRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.ondataavailable = null;
			mediaRecorder.onstop = null;
			mediaRecorder.stop();
			mediaRecorder.stream.getTracks().forEach(track => track.stop());
		}
		stopRecording();
	}

	async function transcribeAudio(blob: Blob) {
		try {
			const formData = new FormData();
			formData.append('audio', blob, 'recording.webm');

			const response = await apiClient.postFormData('/transcribe', formData);
			if (response.ok) {
				const data = await response.json();
				if (data.text) {
					inputValue = data.text;
				}
			}
		} catch (e) {
			console.error('Transcription failed:', e);
			// Use live transcript as fallback
			if (liveTranscript) {
				inputValue = liveTranscript;
			}
		}
	}

	function updateWaveform() {
		if (!analyser || !isRecording) return;

		const dataArray = new Uint8Array(analyser.frequencyBinCount);
		analyser.getByteTimeDomainData(dataArray);

		const newBars: number[] = [];
		const step = Math.floor(dataArray.length / 30);

		for (let i = 0; i < 30; i++) {
			const value = dataArray[i * step] || 128;
			const deviation = Math.abs(value - 128);
			const height = Math.max(4, Math.min(24, 4 + (deviation / 128) * 40));
			newBars.push(height);
		}

		waveformBars = newBars;
		animationId = requestAnimationFrame(updateWaveform);
	}

	function toggleRecording() {
		if (isRecording) {
			stopRecording();
		} else {
			startRecording();
		}
	}

	function handleKeyDown(event: KeyboardEvent) {
		const items = filteredItems();
		const commands = filteredCommands();

		// Handle commands dropdown navigation
		if (showCommandsDropdown && commands.length > 0) {
			switch (event.key) {
				case 'ArrowDown':
					event.preventDefault();
					commandDropdownIndex = Math.min(commandDropdownIndex + 1, commands.length - 1);
					return;
				case 'ArrowUp':
					event.preventDefault();
					commandDropdownIndex = Math.max(commandDropdownIndex - 1, 0);
					return;
				case 'Enter':
				case 'Tab':
					event.preventDefault();
					selectCommand(commands[commandDropdownIndex]);
					return;
				case 'Escape':
					event.preventDefault();
					showCommandsDropdown = false;
					return;
			}
		}

		// Handle project dropdown navigation
		if (showProjectDropdown) {
			const totalItems = projectsList.length + 1; // +1 for "New Project"
			switch (event.key) {
				case 'ArrowDown':
					event.preventDefault();
					projectDropdownIndex = Math.min(projectDropdownIndex + 1, totalItems - 1);
					return;
				case 'ArrowUp':
					event.preventDefault();
					projectDropdownIndex = Math.max(projectDropdownIndex - 1, 0);
					return;
				case 'Enter':
					event.preventDefault();
					if (projectDropdownIndex < projectsList.length) {
						selectedProjectId = projectsList[projectDropdownIndex].id;
						showProjectDropdown = false;
					} else {
						// New Project option
						showProjectDropdown = false;
						windowStore.openWindow('projects');
						onClose();
					}
					return;
				case 'Escape':
					event.preventDefault();
					showProjectDropdown = false;
					return;
			}
		}

		switch (event.key) {
			case 'ArrowDown':
				event.preventDefault();
				if (mode === 'search') {
					selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
				}
				break;
			case 'ArrowUp':
				event.preventDefault();
				if (mode === 'search') {
					selectedIndex = Math.max(selectedIndex - 1, 0);
				}
				break;
			case 'Enter':
				if (event.shiftKey) return; // Allow shift+enter for newlines
				event.preventDefault();
				if (mode === 'chat') {
					// If no project selected, open project dropdown
					if (!selectedProjectId && inputValue.trim()) {
						showProjectDropdown = true;
						showModelDropdown = false;
						projectDropdownIndex = 0;
						return;
					}
					sendMessage();
				} else if (items[selectedIndex]) {
					selectItem(items[selectedIndex]);
				}
				break;
			case 'Escape':
				onClose();
				break;
			case 'Tab':
				event.preventDefault();
				mode = mode === 'search' ? 'chat' : 'search';
				break;
		}
	}

	function selectItem(item: any) {
		if (item.type === 'app') {
			windowStore.openWindow(item.id);
			onClose();
		}
	}

	// Check if can send (requires project)
	let canSend = $derived(inputValue.trim() && selectedProjectId && !isTyping);

	async function sendMessage() {
		if (!inputValue.trim() || isTyping) return;

		// Require project selection
		if (!selectedProjectId) {
			// Flash the project button or show warning
			return;
		}

		const userMessage = inputValue.trim();
		messages = [...messages, { role: 'user', content: userMessage }];
		inputValue = '';
		isTyping = true;

		try {
			const response = await fetch('/api/chat/message', {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					message: userMessage,
					model: selectedModel || undefined,
					project_id: selectedProjectId,
					conversation_id: conversationId || undefined
				})
			});

			// Capture conversation ID from headers
			const newConvId = response.headers.get('X-Conversation-Id');
			if (newConvId) {
				conversationId = newConvId;
			}

			if (response.ok) {
				// Read streaming response
				const reader = response.body?.getReader();
				const decoder = new TextDecoder();
				let fullContent = '';

				if (reader) {
					while (true) {
						const { done, value } = await reader.read();
						if (done) break;
						const chunk = decoder.decode(value, { stream: true });
						fullContent += chunk;

						// Update message in real-time
						const lastIdx = messages.length - 1;
						if (messages[lastIdx]?.role === 'user') {
							messages = [...messages, { role: 'assistant', content: fullContent }];
						} else {
							messages = messages.map((m, i) =>
								i === messages.length - 1 ? { ...m, content: fullContent } : m
							);
						}
					}
				}
			} else {
				messages = [...messages, { role: 'assistant', content: 'Sorry, I encountered an error. Please try again.' }];
			}
		} catch (error) {
			console.error('Chat error:', error);
			messages = [...messages, { role: 'assistant', content: 'Connection error. Please check your network.' }];
		} finally {
			isTyping = false;
		}
	}

	// Open full Chat module with current conversation
	function openInChat() {
		// Pass conversation data via sessionStorage since modules are in iframes
		if (conversationId || messages.length > 0) {
			const chatData = {
				conversationId,
				messages: messages.map(m => ({
					role: m.role,
					content: m.content
				})),
				projectId: selectedProjectId
			};
			sessionStorage.setItem('spotlightChatTransfer', JSON.stringify(chatData));
		}
		windowStore.openWindow('chat');
		onClose();
	}

	function handleBackdropClick(event: MouseEvent) {
		if ((event.target as HTMLElement).classList.contains('spotlight-backdrop')) {
			onClose();
		}
	}

	function handleInput() {
		// Auto-resize textarea
		if (inputElement) {
			inputElement.style.height = 'auto';
			inputElement.style.height = Math.min(inputElement.scrollHeight, 120) + 'px';
		}
	}

	// File attachment handlers
	function handleFileSelect(event: Event) {
		const input = event.target as HTMLInputElement;
		const files = input.files;
		if (!files) return;
		processFiles(Array.from(files));
		if (fileInputRef) fileInputRef.value = '';
	}

	function processFiles(files: File[]) {
		for (const file of files) {
			// Max 10MB per file
			if (file.size > 10 * 1024 * 1024) {
				continue;
			}

			const newFile: AttachedFile = {
				id: crypto.randomUUID(),
				name: file.name,
				type: file.type,
				size: file.size,
				file
			};

			// Generate preview for images
			if (file.type.startsWith('image/')) {
				const reader = new FileReader();
				reader.onload = () => {
					newFile.preview = reader.result as string;
					attachedFiles = [...attachedFiles, newFile];
				};
				reader.readAsDataURL(file);
			} else {
				attachedFiles = [...attachedFiles, newFile];
			}
		}
	}

	function removeFile(fileId: string) {
		attachedFiles = attachedFiles.filter(f => f.id !== fileId);
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return bytes + ' B';
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
		return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
	}

	// Drag & drop handlers
	function handleDragEnter(e: DragEvent) {
		e.preventDefault();
		isDragging = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		// Only set to false if we're leaving the container (not entering a child)
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		if (
			e.clientX < rect.left ||
			e.clientX > rect.right ||
			e.clientY < rect.top ||
			e.clientY > rect.bottom
		) {
			isDragging = false;
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
		const files = e.dataTransfer?.files;
		if (files) {
			processFiles(Array.from(files));
		}
	}

	// Select a slash command
	function selectCommand(cmd: typeof slashCommands[0]) {
		// Special handling for /image command
		if (cmd.id === 'image') {
			showImageSearch = true;
			inputValue = '';
			showCommandsDropdown = false;
			return;
		}

		inputValue = cmd.name + ' ';
		showCommandsDropdown = false;
		inputElement?.focus();
	}

	// Get file icon based on type
	function getFileIcon(type: string): string {
		if (type.startsWith('image/')) return '🖼️';
		if (type.startsWith('video/')) return '🎬';
		if (type.startsWith('audio/')) return '🎵';
		if (type.includes('pdf')) return '📄';
		if (type.includes('word') || type.includes('document')) return '📝';
		if (type.includes('sheet') || type.includes('excel')) return '📊';
		if (type.includes('zip') || type.includes('archive')) return '📦';
		return '📎';
	}
</script>

{#if open}
	<div
		class="spotlight-backdrop"
		onclick={handleBackdropClick}
		role="dialog"
		aria-modal="true"
		aria-label="Quick Chat"
		transition:fade={{ duration: 150 }}
	>
		<div class="spotlight-container" transition:scale={{ duration: 150, start: 0.95 }}>
			<!-- Messages Area (when there are messages) -->
			{#if messages.length > 0}
				<div class="messages-area">
					<!-- Open in Chat icon - top right -->
					<button class="expand-chat-btn" onclick={openInChat} title="Open in Chat">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"/>
							<polyline points="15 3 21 3 21 9"/>
							<line x1="10" y1="14" x2="21" y2="3"/>
						</svg>
					</button>
					{#each messages as msg}
						<div class="message {msg.role}">
							<div class="message-content">{msg.content}</div>
						</div>
					{/each}
					{#if isTyping}
						<div class="message assistant">
							<div class="message-content typing">
								<span></span><span></span><span></span>
							</div>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Input Card -->
			<div
				class="input-card"
				class:dragging={isDragging}
				ondragenter={handleDragEnter}
				ondragleave={handleDragLeave}
				ondragover={handleDragOver}
				ondrop={handleDrop}
			>
				<!-- Hidden file input -->
				<input
					type="file"
					bind:this={fileInputRef}
					onchange={handleFileSelect}
					multiple
					accept="image/*,.pdf,.doc,.docx,.txt,.csv,.json,.md"
					style="display: none;"
				/>

				<!-- Drag overlay -->
				{#if isDragging}
					<div class="drag-overlay" transition:fade={{ duration: 100 }}>
						<div class="drag-content">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/>
								<polyline points="17 8 12 3 7 8"/>
								<line x1="12" y1="3" x2="12" y2="15"/>
							</svg>
							<span>Drop files here</span>
						</div>
					</div>
				{/if}

				<!-- File attachments preview -->
				{#if attachedFiles.length > 0}
					<div class="attachments-preview">
						{#each attachedFiles as file (file.id)}
							<div class="attachment-item" transition:scale={{ duration: 150 }}>
								{#if file.preview}
									<img src={file.preview} alt={file.name} class="attachment-thumb" />
								{:else}
									<div class="attachment-icon">{getFileIcon(file.type)}</div>
								{/if}
								<div class="attachment-info">
									<span class="attachment-name">{file.name}</span>
									<span class="attachment-size">{formatFileSize(file.size)}</span>
								</div>
								<button class="attachment-remove" onclick={() => removeFile(file.id)}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
									</svg>
								</button>
							</div>
						{/each}
					</div>
				{/if}

				{#if isRecording}
					<!-- Recording UI -->
					<div class="recording-area">
						{#if liveTranscript}
							<div class="live-transcript">{liveTranscript}</div>
						{:else}
							<div class="live-transcript placeholder">Listening...</div>
						{/if}
						<div class="waveform-bar">
							<button class="cancel-btn" onclick={cancelRecording} title="Cancel">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
								</svg>
							</button>
							<div class="waveform">
								{#each waveformBars as height}
									<div class="bar" style="height: {height}px"></div>
								{/each}
							</div>
							<span class="duration">{formatDuration(recordingDuration)}</span>
							<button class="confirm-btn" onclick={stopRecording} title="Done">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polyline points="20 6 9 17 4 12"/>
								</svg>
							</button>
						</div>
					</div>
				{:else}
					<!-- Textarea -->
					<textarea
						bind:this={inputElement}
						bind:value={inputValue}
						placeholder="Ask anything... (type / for commands)"
						rows={1}
						onkeydown={handleKeyDown}
						oninput={handleInput}
					></textarea>

					<!-- Commands dropdown -->
					{#if showCommandsDropdown && filteredCommands().length > 0}
						<div class="commands-dropdown" transition:fly={{ y: 5, duration: 150 }}>
							{#each filteredCommands() as cmd, i (cmd.id)}
								<button
									class="command-item"
									class:highlighted={commandDropdownIndex === i}
									onclick={() => selectCommand(cmd)}
									onmouseenter={() => commandDropdownIndex = i}
								>
									<span class="command-icon">{cmd.icon}</span>
									<div class="command-info">
										<span class="command-name">{cmd.name}</span>
										<span class="command-desc">{cmd.description}</span>
									</div>
								</button>
							{/each}
						</div>
					{/if}
				{/if}

				<!-- Bottom Controls -->
				<div class="controls-row">
					<div class="left-controls">
						<!-- Attachment -->
						<button class="icon-btn" title="Attach file" onclick={() => fileInputRef?.click()}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
							</svg>
						</button>

						<!-- Project Selector -->
						<div class="dropdown-wrapper">
							<button
								class="btn-pill btn-pill-sm selector-btn {selectedProject ? 'btn-pill-secondary' : 'btn-pill-ghost'}"
								class:required={!selectedProject && inputValue.trim()}
								onclick={() => { showProjectDropdown = !showProjectDropdown; showModelDropdown = false; }}
							>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="w-4 h-4">
									<path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
								</svg>
								<span>{selectedProject ? selectedProject.name : 'Continue'}</span>
								{#if !selectedProject}
									<span class="required-dot"></span>
								{/if}
								<svg class="chevron w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M19 9l-7 7-7-7" />
								</svg>
							</button>
							{#if showProjectDropdown}
								<div class="dropdown-menu" transition:fly={{ y: 5, duration: 150 }}>
									{#each projectsList as project, i}
										<button
											class="dropdown-item"
											class:active={selectedProjectId === project.id}
											class:highlighted={projectDropdownIndex === i}
											onclick={() => { selectedProjectId = project.id; showProjectDropdown = false; }}
											onmouseenter={() => projectDropdownIndex = i}
										>
											{project.name}
										</button>
									{/each}
									<button
										class="dropdown-item create-new"
										class:highlighted={projectDropdownIndex === projectsList.length}
										onclick={() => { showProjectDropdown = false; windowStore.openWindow('projects'); onClose(); }}
										onmouseenter={() => projectDropdownIndex = projectsList.length}
									>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
										</svg>
										New Project
									</button>
								</div>
							{/if}
						</div>

						<!-- Model Selector (icon only) -->
						<div class="dropdown-wrapper">
							<button
								class="icon-btn"
								onclick={() => { showModelDropdown = !showModelDropdown; showProjectDropdown = false; }}
								title={currentModelName()}
							>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="4" y="4" width="16" height="16" rx="2"/>
									<circle cx="9" cy="9" r="1.5" fill="currentColor"/>
									<circle cx="15" cy="9" r="1.5" fill="currentColor"/>
									<path d="M9 15h6"/>
								</svg>
							</button>
							{#if showModelDropdown}
								<div class="dropdown-menu model-menu" transition:fly={{ y: 5, duration: 150 }}>
									<div class="dropdown-header">
										Provider: {activeProvider === 'ollama_local' ? 'Local' : activeProvider}
									</div>
									{#if availableModels.length === 0}
										<div class="dropdown-empty">No models available</div>
									{:else}
										{#each availableModels as model}
											<button
												class="dropdown-item"
												class:active={selectedModel === model.id}
												onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
											>
												<span class="model-name">{model.name}</span>
												{#if model.size}
													<span class="model-size">{model.size}</span>
												{/if}
											</button>
										{/each}
									{/if}
								</div>
							{/if}
						</div>
					</div>

					<div class="right-controls">
						<!-- Keyboard hints -->
						<div class="hints">
							<span><kbd>⌘D</kbd> Voice</span>
							<span><kbd>↵</kbd> Send</span>
						</div>

						<!-- Voice Button -->
						<button
							class="icon-btn mic"
							class:recording={isRecording}
							onclick={toggleRecording}
							title="Voice input"
						>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M12 1a3 3 0 00-3 3v8a3 3 0 006 0V4a3 3 0 00-3-3z"/>
								<path d="M19 10v2a7 7 0 01-14 0v-2"/>
								<line x1="12" y1="19" x2="12" y2="23"/>
								<line x1="8" y1="23" x2="16" y2="23"/>
							</svg>
						</button>

						<!-- Send Button (circle) -->
						<button
							class="btn-pill btn-pill-icon {canSend ? 'btn-pill-primary' : 'btn-pill-ghost'} send-btn"
							onclick={sendMessage}
							disabled={!canSend}
							title={!selectedProjectId ? 'Select a project first' : 'Send message'}
						>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M5 10l7-7m0 0l7 7m-7-7v18" />
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Search Results (when in search mode) -->
			{#if mode === 'search' && filteredItems().length > 0}
				<div class="search-results">
					{#each filteredItems() as item, index (item.id)}
						<button
							class="search-item"
							class:selected={index === selectedIndex}
							onclick={() => selectItem(item)}
							onmouseenter={() => selectedIndex = index}
						>
							<div class="item-icon" style="background: {item.color}15;">
								<svg viewBox="0 0 24 24" fill="none" stroke={item.color} stroke-width="1.5">
									<path d={item.icon} />
								</svg>
							</div>
							<div class="item-info">
								<span class="item-name">{item.name}</span>
								<span class="item-desc">{item.description}</span>
							</div>
						</button>
					{/each}
				</div>
			{/if}

			<!-- Footer -->
			<div class="footer">
				<button
					class="btn-pill {mode === 'search' ? 'btn-pill-primary' : 'btn-pill-ghost'} btn-pill-sm"
					onclick={() => mode = 'search'}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="11" cy="11" r="8"/>
						<path d="m21 21-4.35-4.35"/>
					</svg>
					Search
				</button>
				<button
					class="btn-pill {mode === 'chat' ? 'btn-pill-primary' : 'btn-pill-ghost'} btn-pill-sm"
					onclick={() => mode = 'chat'}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
					</svg>
					Chat
				</button>
				<div class="footer-spacer"></div>
				<span class="footer-hint"><kbd>Tab</kbd> Switch · <kbd>Esc</kbd> Close</span>
			</div>
		</div>
	</div>
{/if}

<!-- Image Search Modal -->
<ImageSearchModal bind:show={showImageSearch} />

<style>
	.spotlight-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.3);
		backdrop-filter: blur(4px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 99999;
	}

	.spotlight-container {
		width: 100%;
		max-width: 600px;
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	/* Messages */
	.messages-area {
		position: relative;
		background: white;
		border-radius: 20px 20px 0 0;
		padding: 16px;
		padding-top: 40px;
		max-height: 300px;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 12px;
		border: 1px solid rgba(0, 0, 0, 0.08);
		border-bottom: none;
		scrollbar-width: none;
		-ms-overflow-style: none;
	}

	.messages-area::-webkit-scrollbar {
		display: none;
	}

	.expand-chat-btn {
		position: absolute;
		top: 8px;
		right: 8px;
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: white;
		border: 1px solid #e5e5e5;
		border-radius: 6px;
		color: #888;
		cursor: pointer;
		transition: all 0.15s;
		z-index: 10;
	}

	.expand-chat-btn:hover {
		background: #f5f5f5;
		color: #333;
		border-color: #ccc;
	}

	.expand-chat-btn svg {
		width: 14px;
		height: 14px;
	}

	.message {
		display: flex;
		max-width: 85%;
	}

	.message.user {
		align-self: flex-end;
	}

	.message.assistant {
		align-self: flex-start;
	}

	.message-content {
		padding: 10px 14px;
		border-radius: 16px;
		font-size: 14px;
		line-height: 1.5;
		white-space: pre-wrap;
	}

	.message.user .message-content {
		background: #111;
		color: white;
		border-bottom-right-radius: 4px;
	}

	.message.assistant .message-content {
		background: #f3f4f6;
		color: #111;
		border-bottom-left-radius: 4px;
	}

	.message-content.typing {
		display: flex;
		gap: 4px;
		padding: 14px 18px;
	}

	.message-content.typing span {
		width: 8px;
		height: 8px;
		background: #999;
		border-radius: 50%;
		animation: bounce 1.4s infinite ease-in-out both;
	}

	.message-content.typing span:nth-child(2) { animation-delay: 0.2s; }
	.message-content.typing span:nth-child(3) { animation-delay: 0.4s; }

	@keyframes bounce {
		0%, 80%, 100% { transform: scale(0.8); }
		40% { transform: scale(1); }
	}

	/* Input Card */
	.input-card {
		background: white;
		border-radius: 20px;
		padding: 16px;
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
		border: 1px solid rgba(0, 0, 0, 0.08);
		position: relative;
		transition: border-color 0.2s;
	}

	.input-card.dragging {
		border-color: #3b82f6;
		border-style: dashed;
	}

	.messages-area + .input-card {
		border-radius: 0 0 20px 20px;
		border-top: 1px solid rgba(0, 0, 0, 0.06);
	}

	/* Drag overlay */
	.drag-overlay {
		position: absolute;
		inset: 0;
		background: rgba(59, 130, 246, 0.1);
		border-radius: 20px;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 10;
	}

	.drag-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		color: #3b82f6;
	}

	.drag-content svg {
		width: 32px;
		height: 32px;
	}

	.drag-content span {
		font-size: 14px;
		font-weight: 500;
	}

	/* File attachments */
	.attachments-preview {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 12px;
		padding-bottom: 12px;
		border-bottom: 1px solid #f0f0f0;
	}

	.attachment-item {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 8px;
		background: #f5f5f5;
		border-radius: 8px;
		max-width: 200px;
	}

	.attachment-thumb {
		width: 36px;
		height: 36px;
		border-radius: 6px;
		object-fit: cover;
	}

	.attachment-icon {
		width: 36px;
		height: 36px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 20px;
		background: white;
		border-radius: 6px;
	}

	.attachment-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.attachment-name {
		font-size: 12px;
		font-weight: 500;
		color: #333;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.attachment-size {
		font-size: 10px;
		color: #888;
	}

	.attachment-remove {
		width: 20px;
		height: 20px;
		border: none;
		background: transparent;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #999;
		transition: all 0.15s;
		flex-shrink: 0;
	}

	.attachment-remove:hover {
		background: #e5e5e5;
		color: #666;
	}

	.attachment-remove svg {
		width: 12px;
		height: 12px;
	}

	/* Commands dropdown */
	.commands-dropdown {
		position: absolute;
		bottom: 100%;
		left: 16px;
		right: 16px;
		margin-bottom: 8px;
		background: white;
		border: 1px solid #e5e5e5;
		border-radius: 12px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
		overflow: hidden;
		z-index: 100;
	}

	.command-item {
		width: 100%;
		padding: 10px 12px;
		border: none;
		background: none;
		text-align: left;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 10px;
		transition: background 0.1s;
	}

	.command-item:hover,
	.command-item.highlighted {
		background: #f5f5f5;
	}

	.command-icon {
		font-size: 18px;
		width: 28px;
		text-align: center;
	}

	.command-info {
		flex: 1;
		display: flex;
		flex-direction: column;
	}

	.command-name {
		font-size: 13px;
		font-weight: 500;
		color: #333;
		font-family: monospace;
	}

	.command-desc {
		font-size: 11px;
		color: #888;
	}

	textarea {
		width: 100%;
		border: none;
		font-size: 15px;
		font-family: inherit;
		resize: none;
		outline: none;
		background: transparent;
		color: #111;
		line-height: 1.5;
		min-height: 24px;
		max-height: 120px;
	}

	textarea::placeholder {
		color: #999;
	}

	/* Recording Area */
	.recording-area {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	.live-transcript {
		font-size: 14px;
		color: #111;
		min-height: 24px;
	}

	.live-transcript.placeholder {
		color: #999;
	}

	.waveform-bar {
		display: flex;
		align-items: center;
		gap: 10px;
		background: #1f2937;
		border-radius: 24px;
		padding: 8px 14px;
	}

	.waveform {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 2px;
		height: 24px;
	}

	.waveform .bar {
		width: 2px;
		background: white;
		border-radius: 1px;
		transition: height 0.05s;
	}

	.duration {
		font-size: 12px;
		font-family: monospace;
		color: white;
		min-width: 36px;
		text-align: right;
	}

	.cancel-btn, .confirm-btn {
		width: 28px;
		height: 28px;
		border: none;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.15s;
	}

	.cancel-btn {
		background: transparent;
		color: #9ca3af;
	}

	.cancel-btn:hover {
		color: white;
	}

	.confirm-btn {
		background: white;
		color: #1f2937;
	}

	.confirm-btn:hover {
		background: #e5e7eb;
	}

	.cancel-btn svg, .confirm-btn svg {
		width: 16px;
		height: 16px;
	}

	/* Controls Row */
	.controls-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-top: 12px;
	}

	.left-controls, .right-controls {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.icon-btn {
		width: 36px;
		height: 36px;
		border: none;
		background: transparent;
		border-radius: 10px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #888;
		transition: all 0.15s;
	}

	.icon-btn:hover {
		background: #f3f3f3;
		color: #333;
	}

	.icon-btn:active {
		transform: scale(0.95);
	}

	.icon-btn.mic.recording {
		background: #ef4444;
		color: white;
		animation: pulse 1.5s infinite;
	}

	@keyframes pulse {
		0%, 100% { box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.4); }
		50% { box-shadow: 0 0 0 8px rgba(239, 68, 68, 0); }
	}

	.icon-btn svg {
		width: 18px;
		height: 18px;
	}

	/* Selector Buttons */
	.dropdown-wrapper {
		position: relative;
	}

	.selector-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 10px;
		border: 1px solid #e5e5e5;
		background: white;
		border-radius: 8px;
		cursor: pointer;
		font-size: 13px;
		color: #666;
		transition: all 0.15s;
	}

	.selector-btn:hover {
		border-color: #ccc;
		color: #333;
	}

	.selector-btn.selected {
		background: #f0f0ff;
		border-color: #c7c7ff;
		color: #5b5bd6;
	}

	.selector-btn.required {
		border-color: #ef4444;
	}

	.required-dot {
		width: 5px;
		height: 5px;
		background: #ef4444;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.selector-btn svg {
		width: 14px;
		height: 14px;
	}

	.selector-btn .chevron {
		width: 12px;
		height: 12px;
		opacity: 0.5;
	}

	/* Dropdown Menu */
	.dropdown-menu {
		position: absolute;
		bottom: 100%;
		left: 0;
		margin-bottom: 6px;
		min-width: 180px;
		background: white;
		border: 1px solid #e5e5e5;
		border-radius: 12px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
		overflow: hidden;
		z-index: 100;
	}

	.dropdown-menu.model-menu {
		min-width: 220px;
	}

	.dropdown-header {
		padding: 8px 12px;
		font-size: 11px;
		color: #666;
		background: #f9f9f9;
		border-bottom: 1px solid #eee;
	}

	.dropdown-item {
		width: 100%;
		padding: 10px 12px;
		border: none;
		background: none;
		text-align: left;
		font-size: 13px;
		color: #333;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: space-between;
		transition: background 0.1s;
	}

	.dropdown-item:hover,
	.dropdown-item.highlighted {
		background: #f5f5f5;
	}

	.dropdown-item.active {
		background: #f0f0ff;
		color: #5b5bd6;
	}

	.dropdown-item.active.highlighted {
		background: #e0e0ff;
	}

	.dropdown-item.create-new {
		display: flex;
		align-items: center;
		gap: 8px;
		color: #3b82f6;
		border-top: 1px solid #eee;
		margin-top: 4px;
		padding-top: 12px;
	}

	.dropdown-item.create-new:hover {
		background: #eff6ff;
	}

	.dropdown-item.create-new svg {
		width: 14px;
		height: 14px;
	}

	.dropdown-empty {
		padding: 16px;
		text-align: center;
		color: #999;
		font-size: 13px;
	}

	.model-name {
		flex: 1;
	}

	.model-size {
		font-size: 11px;
		color: #999;
	}

	/* Hints */
	.hints {
		display: flex;
		gap: 12px;
		font-size: 11px;
		color: #999;
	}

	.hints kbd {
		background: #f3f3f3;
		padding: 2px 5px;
		border-radius: 4px;
		font-family: inherit;
		font-size: 10px;
	}

	/* Send Button - Circle */
	.send-btn {
		width: 36px;
		height: 36px;
		border: none;
		background: #3b82f6;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		transition: all 0.15s;
	}

	.send-btn:hover:not(:disabled) {
		background: #2563eb;
		transform: scale(1.05);
	}

	.send-btn:active:not(:disabled) {
		transform: scale(0.95);
	}

	.send-btn:disabled {
		background: #e5e5e5;
		color: #bbb;
		cursor: not-allowed;
	}

	.send-btn svg {
		width: 18px;
		height: 18px;
	}

	/* Search Results */
	.search-results {
		background: white;
		border-radius: 16px;
		margin-top: 8px;
		padding: 8px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
		border: 1px solid rgba(0, 0, 0, 0.06);
		max-height: 320px;
		overflow-y: auto;
		scrollbar-width: none;
		-ms-overflow-style: none;
	}

	.search-results::-webkit-scrollbar {
		display: none;
	}

	.search-item {
		display: flex;
		align-items: center;
		gap: 12px;
		width: 100%;
		padding: 10px 12px;
		border: none;
		background: none;
		border-radius: 12px;
		cursor: pointer;
		text-align: left;
		transition: all 0.15s;
	}

	.search-item:hover,
	.search-item.selected {
		background: #f5f5f5;
	}

	.search-item:active {
		transform: scale(0.98);
	}

	.item-icon {
		width: 40px;
		height: 40px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		transition: transform 0.15s;
	}

	.search-item:hover .item-icon {
		transform: scale(1.05);
	}

	.item-icon svg {
		width: 20px;
		height: 20px;
	}

	.item-info {
		flex: 1;
		min-width: 0;
	}

	.item-name {
		display: block;
		font-size: 14px;
		font-weight: 500;
		color: #111;
	}

	.item-desc {
		display: block;
		font-size: 12px;
		color: #666;
		margin-top: 2px;
	}

	/* Footer */
	.footer {
		display: flex;
		align-items: center;
		gap: 4px;
		margin-top: 8px;
		padding: 0 4px;
	}

	.footer-spacer {
		flex: 1;
	}

	.footer-hint {
		font-size: 11px;
		color: #888;
	}

	.footer-hint kbd {
		background: rgba(0, 0, 0, 0.06);
		padding: 2px 5px;
		border-radius: 4px;
		font-family: inherit;
	}

	/* ===== DARK MODE FOR SPOTLIGHT ===== */
	:global(.dark) .spotlight-backdrop {
		background: rgba(0, 0, 0, 0.5);
	}

	:global(.dark) .messages-area {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.12);
	}

	:global(.dark) .expand-chat-btn {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		color: #a1a1a6;
	}

	:global(.dark) .expand-chat-btn:hover {
		background: #3a3a3c;
		color: #f5f5f7;
	}

	:global(.dark) .message.user .message-content {
		background: #0A84FF;
		color: white;
	}

	:global(.dark) .message.assistant .message-content {
		background: #2c2c2e;
		color: #f5f5f7;
	}

	:global(.dark) .input-card {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .input-card.dragging {
		border-color: #0A84FF;
	}

	:global(.dark) textarea {
		color: #f5f5f7;
	}

	:global(.dark) textarea::placeholder {
		color: #6e6e73;
	}

	:global(.dark) .attachments-preview {
		border-bottom-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .attachment-item {
		background: #2c2c2e;
	}

	:global(.dark) .attachment-icon {
		background: #3a3a3c;
	}

	:global(.dark) .attachment-name {
		color: #f5f5f7;
	}

	:global(.dark) .attachment-size {
		color: #6e6e73;
	}

	:global(.dark) .attachment-remove {
		color: #6e6e73;
	}

	:global(.dark) .attachment-remove:hover {
		background: #3a3a3c;
		color: #f5f5f7;
	}

	:global(.dark) .commands-dropdown {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .command-item:hover,
	:global(.dark) .command-item.highlighted {
		background: #3a3a3c;
	}

	:global(.dark) .command-name {
		color: #f5f5f7;
	}

	:global(.dark) .command-desc {
		color: #6e6e73;
	}

	:global(.dark) .live-transcript {
		color: #f5f5f7;
	}

	:global(.dark) .live-transcript.placeholder {
		color: #6e6e73;
	}

	:global(.dark) .icon-btn {
		color: #a1a1a6;
	}

	:global(.dark) .icon-btn:hover {
		background: #2c2c2e;
		color: #f5f5f7;
	}

	:global(.dark) .selector-btn {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		color: #a1a1a6;
	}

	:global(.dark) .selector-btn:hover {
		border-color: rgba(255, 255, 255, 0.2);
		color: #f5f5f7;
	}

	:global(.dark) .selector-btn.selected {
		background: rgba(10, 132, 255, 0.2);
		border-color: rgba(10, 132, 255, 0.4);
		color: #0A84FF;
	}

	:global(.dark) .dropdown-menu {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .dropdown-header {
		background: #1c1c1e;
		border-bottom-color: rgba(255, 255, 255, 0.1);
		color: #6e6e73;
	}

	:global(.dark) .dropdown-item {
		color: #f5f5f7;
	}

	:global(.dark) .dropdown-item:hover,
	:global(.dark) .dropdown-item.highlighted {
		background: #3a3a3c;
	}

	:global(.dark) .dropdown-item.active {
		background: rgba(10, 132, 255, 0.2);
		color: #0A84FF;
	}

	:global(.dark) .dropdown-item.create-new {
		border-top-color: rgba(255, 255, 255, 0.1);
		color: #0A84FF;
	}

	:global(.dark) .dropdown-item.create-new:hover {
		background: rgba(10, 132, 255, 0.1);
	}

	:global(.dark) .dropdown-empty {
		color: #6e6e73;
	}

	:global(.dark) .model-size {
		color: #6e6e73;
	}

	:global(.dark) .hints {
		color: #6e6e73;
	}

	:global(.dark) .hints kbd {
		background: #2c2c2e;
		color: #a1a1a6;
	}

	:global(.dark) .send-btn {
		background: #0A84FF;
	}

	:global(.dark) .send-btn:hover:not(:disabled) {
		background: #0070E0;
	}

	:global(.dark) .send-btn:disabled {
		background: #3a3a3c;
		color: #6e6e73;
	}

	:global(.dark) .search-results {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .search-item:hover,
	:global(.dark) .search-item.selected {
		background: #2c2c2e;
	}

	:global(.dark) .item-name {
		color: #f5f5f7;
	}

	:global(.dark) .item-desc {
		color: #6e6e73;
	}

	:global(.dark) .footer-hint {
		color: #6e6e73;
	}

	:global(.dark) .footer-hint kbd {
		background: rgba(255, 255, 255, 0.1);
		color: #a1a1a6;
	}
</style>
