<script lang="ts">
	import { tick, onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { api, type ArtifactListItem, type Artifact, type Node } from '$lib/api/client';

	// Message interface
	interface ChatMessage {
		id: string;
		role: 'user' | 'assistant';
		content: string;
		artifacts?: { title: string; type: string; content: string }[];
	}

	// UI State
	let messagesContainer: HTMLDivElement | undefined = $state(undefined);
	let inputRef: HTMLTextAreaElement | undefined = $state(undefined);
	let inputValue = $state('');
	let selectedModel = $state('qwen3-coder:480b-cloud');
	let selectedContext = $state('Default');
	let chatSidebarOpen = $state(true);
	let artifactsPanelOpen = $state(false);
	let searchQuery = $state('');
	let showContextDropdown = $state(false);
	let showModelDropdown = $state(false);
	let showNodeDropdown = $state(false);
	let copiedMessageId: string | null = $state(null);
	let filterTab: 'all' | 'pinned' | 'recent' = $state('all');

	// Chat state
	let messages: ChatMessage[] = $state([]);
	let isStreaming = $state(false);
	let conversationId: string | null = $state(null);
	let abortController: AbortController | null = $state(null);

	// Active node state
	let activeNode: Node | null = $state(null);
	let nodeContextPrompt: string | null = $state(null);

	// Artifacts state
	let artifacts: ArtifactListItem[] = $state([]);
	let selectedArtifact: Artifact | null = $state(null);
	let loadingArtifacts = $state(false);
	let artifactFilter: string = $state('all');

	// Artifact generation state (for live preview)
	let generatingArtifact = $state(false);
	let generatingArtifactTitle = $state('');
	let generatingArtifactType = $state('');
	let generatingArtifactContent = $state('');

	// Resizable panel state
	let artifactPanelWidth = $state(400);
	let isResizing = $state(false);
	let resizeStartX = $state(0);
	let resizeStartWidth = $state(0);

	// Currently viewing artifact in panel
	let viewingArtifactFromMessage: { title: string; type: string; content: string } | null = $state(null);

	// Editable artifact state
	let isEditingArtifact = $state(false);
	let editedArtifactContent = $state('');

	// Save to node modal
	let showSaveToNodeModal = $state(false);
	let availableNodes: Node[] = $state([]);
	let selectedNodeForSave: string = $state('');

	// Load available nodes for saving
	async function loadAvailableNodes() {
		try {
			availableNodes = await api.getNodes();
		} catch (e) {
			console.error('Failed to load nodes:', e);
		}
	}

	function startEditingArtifact() {
		if (viewingArtifactFromMessage) {
			editedArtifactContent = viewingArtifactFromMessage.content;
			isEditingArtifact = true;
		}
	}

	function saveArtifactEdit() {
		if (viewingArtifactFromMessage) {
			viewingArtifactFromMessage = {
				...viewingArtifactFromMessage,
				content: editedArtifactContent
			};
			isEditingArtifact = false;
		}
	}

	function cancelArtifactEdit() {
		isEditingArtifact = false;
		editedArtifactContent = '';
	}

	async function openSaveToNodeModal() {
		await loadAvailableNodes();
		showSaveToNodeModal = true;
	}

	async function saveArtifactToNode() {
		if (!selectedNodeForSave || !viewingArtifactFromMessage) return;

		try {
			// TODO: Implement backend endpoint to save artifact to node
			// For now, just close the modal
			console.log('Saving artifact to node:', selectedNodeForSave, viewingArtifactFromMessage);
			showSaveToNodeModal = false;
			selectedNodeForSave = '';
		} catch (e) {
			console.error('Failed to save artifact to node:', e);
		}
	}

	// Load active node on mount
	async function loadActiveNode() {
		try {
			activeNode = await api.getActiveNode();
			if (activeNode) {
				// Build context prompt from node data
				const focusItems = activeNode.this_week_focus?.map((f, i) => `${i + 1}. ${f}`).join('\n') || 'Not defined';
				nodeContextPrompt = `Current Active Node: ${activeNode.name}

Purpose: ${activeNode.purpose || 'Not defined'}

Current Status: ${activeNode.current_status || 'Not defined'}

This Week's Focus:
${focusItems}

Use this context to inform your responses.`;
			}
		} catch (e) {
			console.error('Failed to load active node:', e);
		}
	}

	async function handleDeactivateNode() {
		if (!activeNode) return;
		try {
			await api.deactivateNode(activeNode.id);
			activeNode = null;
			nodeContextPrompt = null;
			showNodeDropdown = false;
		} catch (e) {
			console.error('Failed to deactivate node:', e);
		}
	}

	onMount(() => {
		loadActiveNode();
	});

	// Load artifacts
	async function loadArtifacts() {
		loadingArtifacts = true;
		try {
			const filters: { type?: string; conversationId?: string } = {};
			if (artifactFilter !== 'all') filters.type = artifactFilter;
			if (conversationId) filters.conversationId = conversationId;
			artifacts = await api.getArtifacts(filters);
		} catch (error) {
			console.error('Failed to load artifacts:', error);
		} finally {
			loadingArtifacts = false;
		}
	}

	async function selectArtifact(id: string) {
		try {
			selectedArtifact = await api.getArtifact(id);
		} catch (error) {
			console.error('Failed to load artifact:', error);
		}
	}

	function closeArtifactDetail() {
		selectedArtifact = null;
	}

	function getArtifactIcon(type: string) {
		switch (type) {
			case 'proposal': return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			case 'sop': return 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4';
			case 'framework': return 'M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z';
			case 'agenda': return 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z';
			case 'report': return 'M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			case 'plan': return 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2';
			default: return 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z';
		}
	}

	function getArtifactColor(type: string) {
		switch (type) {
			case 'proposal': return 'text-blue-500 bg-blue-50';
			case 'sop': return 'text-green-500 bg-green-50';
			case 'framework': return 'text-purple-500 bg-purple-50';
			case 'agenda': return 'text-orange-500 bg-orange-50';
			case 'report': return 'text-red-500 bg-red-50';
			case 'plan': return 'text-teal-500 bg-teal-50';
			default: return 'text-gray-500 bg-gray-50';
		}
	}

	// Load artifacts when panel opens
	$effect(() => {
		if (artifactsPanelOpen) {
			loadArtifacts();
		}
	});

	// Available models
	interface ModelOption {
		id: string;
		name: string;
		description: string;
		type: 'cloud' | 'local';
	}

	const models: ModelOption[] = [
		// Cloud models (via Ollama Cloud API)
		{ id: 'qwen3-coder:480b-cloud', name: 'Qwen3 Coder 480B', description: 'Best for coding tasks (Cloud)', type: 'cloud' },
		// Local models (requires local Ollama)
		{ id: 'qwen3-coder:latest', name: 'Qwen3 Coder', description: 'Latest Qwen3 coder model', type: 'local' },
		{ id: 'qwen3-coder:30b', name: 'Qwen3 Coder 30B', description: 'Local 30B parameter model', type: 'local' },
		{ id: 'qwen2.5:7b', name: 'Qwen 2.5 7B', description: 'Fast general purpose', type: 'local' },
		{ id: 'llama3.2:latest', name: 'Llama 3.2', description: 'Meta\'s latest model', type: 'local' },
	];

	const cloudModels = models.filter(m => m.type === 'cloud');
	const localModels = models.filter(m => m.type === 'local');

	// Sidebar conversations
	interface SidebarConversation {
		id: string;
		title: string;
		timestamp: string;
		pinned?: boolean;
	}

	let conversations: SidebarConversation[] = $state([]);
	let activeConversationId = $state<string | null>(null);

	// Context options
	const contexts = ['Default', 'Marketing', 'Development', 'Sales', 'Operations'];

	// Quick action prompts
	const quickActions = [
		'Write a business proposal',
		'Analyze my data',
		'Plan my week'
	];

	// Derived state
	let hasConversation = $derived(messages.length > 0);
	let currentModelName = $derived(models.find(m => m.id === selectedModel)?.name ?? selectedModel);

	// Auto-scroll on new messages
	$effect(() => {
		if (messagesContainer && messages.length) {
			tick().then(() => {
				if (messagesContainer) {
					messagesContainer.scrollTop = messagesContainer.scrollHeight;
				}
			});
		}
	});

	function handleQuickAction(prompt: string) {
		inputValue = prompt;
		inputRef?.focus();
	}

	function handleNewChat() {
		messages = [];
		conversationId = null;
		activeConversationId = null;
	}

	function selectConversation(id: string) {
		activeConversationId = id;
		// TODO: Load conversation messages from backend
	}

	async function handleSendMessage() {
		if (!inputValue.trim() || isStreaming) return;

		const userMessage = inputValue.trim();
		inputValue = '';
		if (inputRef) inputRef.style.height = 'auto';

		// Add user message to UI
		const userMsgId = crypto.randomUUID();
		messages = [...messages, { id: userMsgId, role: 'user', content: userMessage }];

		// Create assistant message placeholder
		const assistantMsgId = crypto.randomUUID();
		messages = [...messages, { id: assistantMsgId, role: 'assistant', content: '' }];

		isStreaming = true;
		abortController = new AbortController();

		try {
			// Build request body with optional node context
			const requestBody: Record<string, unknown> = {
				message: userMessage,
				model: selectedModel,
				conversation_id: conversationId,
				system_prompt_key: selectedContext.toLowerCase() === 'default' ? 'default' : selectedContext.toLowerCase(),
			};

			// Include node context if there's an active node
			if (nodeContextPrompt) {
				requestBody.node_context = nodeContextPrompt;
			}

			const response = await fetch('/api/chat/message', {
				credentials: 'include',
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(requestBody),
				signal: abortController.signal,
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			// Get conversation ID from response header
			const newConvId = response.headers.get('X-Conversation-Id');
			if (newConvId) {
				conversationId = newConvId;
			}

			// Stream the response
			const reader = response.body?.getReader();
			const decoder = new TextDecoder();
			let fullContent = '';
			let artifactStarted = false;

			if (reader) {
				while (true) {
					const { done, value } = await reader.read();
					if (done) break;

					const chunk = decoder.decode(value, { stream: true });
					fullContent += chunk;

					// Update the assistant message content
					messages = messages.map(msg =>
						msg.id === assistantMsgId
							? { ...msg, content: msg.content + chunk }
							: msg
					);

					// Detect artifact generation starting
					if (fullContent.includes('```artifact') && !artifactStarted) {
						artifactStarted = true;
						generatingArtifact = true;
						artifactsPanelOpen = true;

						// Try to extract title early if available
						const titleMatch = fullContent.match(/"title":\s*"([^"]+)"/);
						if (titleMatch) {
							generatingArtifactTitle = titleMatch[1];
						}

						// Try to extract type early if available
						const typeMatch = fullContent.match(/"type":\s*"([^"]+)"/);
						if (typeMatch) {
							generatingArtifactType = typeMatch[1];
						}
					}

					// Update title/type if not yet found
					if (artifactStarted && generatingArtifact) {
						if (!generatingArtifactTitle) {
							const titleMatch = fullContent.match(/"title":\s*"([^"]+)"/);
							if (titleMatch) generatingArtifactTitle = titleMatch[1];
						}
						if (!generatingArtifactType) {
							const typeMatch = fullContent.match(/"type":\s*"([^"]+)"/);
							if (typeMatch) generatingArtifactType = typeMatch[1];
						}

						// Extract content being generated
						const contentMatch = fullContent.match(/"content":\s*"([\s\S]*?)(?:"\s*}|$)/);
						if (contentMatch) {
							// Unescape the JSON string content
							generatingArtifactContent = contentMatch[1]
								.replace(/\\n/g, '\n')
								.replace(/\\"/g, '"')
								.replace(/\\\\/g, '\\');
						}

						// Check if artifact block has been closed (look for closing ``` after artifact start)
						// Count occurrences of ``` - if there are at least 2 after ```artifact starts, it's complete
						const afterArtifactStart = fullContent.slice(fullContent.indexOf('```artifact'));
						const backtickMatches = afterArtifactStart.match(/```/g);
						if (backtickMatches && backtickMatches.length >= 2) {
							// Artifact block is complete - stop showing "Writing..." in panel
							generatingArtifact = false;

							// Try to parse and set the final artifact for viewing
							try {
								const artifactMatch = fullContent.match(/```artifact\s*\n([\s\S]*?)\n```/);
								if (artifactMatch) {
									const artifactData = JSON.parse(artifactMatch[1].trim());
									if (artifactData.title && artifactData.type && artifactData.content) {
										viewingArtifactFromMessage = {
											title: artifactData.title,
											type: artifactData.type,
											content: artifactData.content
												.replace(/\\n/g, '\n')
												.replace(/\\"/g, '"')
												.replace(/\\\\/g, '\\')
										};
									}
								}
							} catch {
								// Failed to parse, will rely on final parsing
							}
						}
					}
				}
			}

			// Check if the response contains artifact blocks - final cleanup
			if (fullContent.includes('```artifact')) {
				// Artifact was created - refresh artifacts list
				await loadArtifacts();
			}

			// Reset generation state after streaming completes
			generatingArtifact = false;
			generatingArtifactTitle = '';
			generatingArtifactType = '';
			generatingArtifactContent = '';
		} catch (error: any) {
			if (error.name === 'AbortError') {
				console.log('Request aborted');
			} else {
				console.error('Chat error:', error);
				// Update assistant message with error
				messages = messages.map(msg =>
					msg.id === assistantMsgId
						? { ...msg, content: 'Sorry, there was an error processing your request. Please try again.' }
						: msg
				);
			}
		} finally {
			isStreaming = false;
			abortController = null;
		}
	}

	// Parse artifact blocks from message content for rendering
	interface ParsedPart {
		type: 'text' | 'artifact';
		text?: string;
		artifact?: { title: string; type: string; content: string };
	}

	function parseMessageContent(content: string): ParsedPart[] {
		const parts: ParsedPart[] = [];
		// More flexible regex that matches artifact blocks with any field order
		// Match ```artifact followed by JSON block and closing ```
		const pattern = /```artifact\s*\n([\s\S]*?)\n```/g;
		let lastIndex = 0;
		let match;

		while ((match = pattern.exec(content)) !== null) {
			// Add text before the artifact block
			if (match.index > lastIndex) {
				const textBefore = content.slice(lastIndex, match.index).trim();
				if (textBefore) {
					parts.push({ type: 'text', text: textBefore });
				}
			}

			// Try to parse the JSON inside the artifact block
			try {
				const jsonStr = match[1].trim();
				const artifactData = JSON.parse(jsonStr);

				if (artifactData.title && artifactData.type && artifactData.content) {
					// Unescape content if needed
					const artifactContent = artifactData.content
						.replace(/\\n/g, '\n')
						.replace(/\\"/g, '"')
						.replace(/\\\\/g, '\\');

					parts.push({
						type: 'artifact',
						artifact: {
							title: artifactData.title,
							type: artifactData.type,
							content: artifactContent
						}
					});
				}
			} catch {
				// JSON parsing failed - this might be incomplete, skip it
				console.log('Failed to parse artifact JSON, possibly incomplete');
			}

			lastIndex = match.index + match[0].length;
		}

		// Add remaining text
		if (lastIndex < content.length) {
			const remainingText = content.slice(lastIndex).trim();
			if (remainingText) {
				// Check if remaining text contains an incomplete artifact block
				if (remainingText.includes('```artifact') && !remainingText.includes('```artifact') ||
					(remainingText.includes('```artifact') && remainingText.lastIndexOf('```') === remainingText.indexOf('```artifact') + 3)) {
					// Incomplete artifact block - don't show it
					const beforeArtifact = remainingText.split('```artifact')[0].trim();
					if (beforeArtifact) {
						parts.push({ type: 'text', text: beforeArtifact });
					}
				} else {
					parts.push({ type: 'text', text: remainingText });
				}
			}
		}

		// If no parts found, check if we're in the middle of generating an artifact
		if (parts.length === 0) {
			// Check if content contains an incomplete artifact block (started but not finished)
			if (content.includes('```artifact')) {
				// Extract text before the artifact block
				const beforeArtifact = content.split('```artifact')[0].trim();
				if (beforeArtifact) {
					return [{ type: 'text', text: beforeArtifact }];
				}
				// Nothing to show yet - artifact is being generated
				return [];
			}
			return [{ type: 'text', text: content }];
		}

		return parts;
	}

	// Simple markdown renderer
	function renderMarkdown(text: string): string {
		return text
			// Headers
			.replace(/^### (.+)$/gm, '<h3 class="text-lg font-semibold text-gray-900 mt-4 mb-2">$1</h3>')
			.replace(/^## (.+)$/gm, '<h2 class="text-xl font-semibold text-gray-900 mt-5 mb-3">$1</h2>')
			.replace(/^# (.+)$/gm, '<h1 class="text-2xl font-bold text-gray-900 mt-6 mb-4">$1</h1>')
			// Bold
			.replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold">$1</strong>')
			// Italic
			.replace(/\*(.+?)\*/g, '<em class="italic">$1</em>')
			// Lists
			.replace(/^- (.+)$/gm, '<li class="ml-4 list-disc text-gray-700">$1</li>')
			.replace(/^(\d+)\. (.+)$/gm, '<li class="ml-4 list-decimal text-gray-700">$2</li>')
			// Line breaks
			.replace(/\n\n/g, '</p><p class="mb-3">')
			.replace(/\n/g, '<br/>');
	}

	// Resize handlers
	function startResize(e: MouseEvent) {
		isResizing = true;
		resizeStartX = e.clientX;
		resizeStartWidth = artifactPanelWidth;
		document.addEventListener('mousemove', handleResize);
		document.addEventListener('mouseup', stopResize);
		document.body.style.cursor = 'col-resize';
		document.body.style.userSelect = 'none';
	}

	function handleResize(e: MouseEvent) {
		if (!isResizing) return;
		const delta = resizeStartX - e.clientX;
		const newWidth = Math.min(Math.max(resizeStartWidth + delta, 300), 800);
		artifactPanelWidth = newWidth;
	}

	function stopResize() {
		isResizing = false;
		document.removeEventListener('mousemove', handleResize);
		document.removeEventListener('mouseup', stopResize);
		document.body.style.cursor = '';
		document.body.style.userSelect = '';
	}

	function viewArtifactInPanel(artifact: { title: string; type: string; content: string }) {
		viewingArtifactFromMessage = artifact;
		selectedArtifact = null;
		artifactsPanelOpen = true;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSendMessage();
		}
	}

	function handleInput() {
		if (inputRef) {
			inputRef.style.height = 'auto';
			inputRef.style.height = Math.min(inputRef.scrollHeight, 200) + 'px';
		}
	}

	function handleStop() {
		if (abortController) {
			abortController.abort();
		}
	}

	function copyMessage(content: string, id: string) {
		navigator.clipboard.writeText(content);
		copiedMessageId = id;
		setTimeout(() => copiedMessageId = null, 2000);
	}

	function formatTime(dateStr: string) {
		const date = new Date(dateStr);
		const now = new Date();
		const diffHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60));
		if (diffHours < 1) return 'Just now';
		if (diffHours < 24) return `${diffHours}h ago`;
		return date.toLocaleDateString();
	}
</script>

<!-- Fixed height container that fills parent -->
<div class="h-full flex overflow-hidden">
	<!-- Chat Conversations Sidebar -->
	{#if chatSidebarOpen}
		<div class="w-64 h-full flex flex-col bg-white border-r border-gray-200 flex-shrink-0" transition:fly={{ x: -256, duration: 200 }}>
			<!-- Header -->
			<div class="p-4 flex-shrink-0">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">Chats</h2>
					<button
						onclick={handleNewChat}
						class="w-8 h-8 flex items-center justify-center bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
					</button>
				</div>

				<!-- Search -->
				<div class="relative">
					<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						placeholder="Search conversations..."
						bind:value={searchQuery}
						class="w-full pl-10 pr-4 py-2 text-sm bg-gray-50 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent"
					/>
				</div>

				<!-- Filter Tabs -->
				<div class="flex items-center gap-1 mt-3">
					<button
						onclick={() => filterTab = 'all'}
						class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors {filterTab === 'all' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
					>
						All
					</button>
					<button
						onclick={() => filterTab = 'pinned'}
						class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors {filterTab === 'pinned' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
					>
						Pinned
					</button>
					<button
						onclick={() => filterTab = 'recent'}
						class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors {filterTab === 'recent' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
					>
						Recent
					</button>
				</div>
			</div>

			<!-- Conversation List - scrollable -->
			<div class="flex-1 overflow-y-auto px-2">
				{#each conversations as conv (conv.id)}
					<button
						onclick={() => selectConversation(conv.id)}
						class="w-full text-left p-3 rounded-lg mb-1 transition-colors {activeConversationId === conv.id ? 'bg-gray-100' : 'hover:bg-gray-50'}"
					>
						<p class="text-sm font-medium text-gray-900 truncate">{conv.title}</p>
						<p class="text-xs text-gray-500 mt-1">{formatTime(conv.timestamp)}</p>
					</button>
				{/each}
			</div>

			<!-- Footer -->
			<div class="p-3 flex-shrink-0 border-t border-gray-100">
				<button class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
					</svg>
					View archived
				</button>
			</div>
		</div>
	{/if}

	<!-- Main Chat Area - fills remaining space -->
	<div class="flex-1 flex flex-col min-w-0 h-full bg-gray-50">
		<!-- Toggle button - fixed header -->
		<div class="h-12 flex items-center justify-between px-4 flex-shrink-0 border-b border-gray-100">
			<button
				onclick={() => chatSidebarOpen = !chatSidebarOpen}
				class="p-2 text-gray-400 hover:text-gray-600 hover:bg-white rounded-lg transition-colors"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					{#if chatSidebarOpen}
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
					{:else}
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
					{/if}
				</svg>
			</button>

			<div class="flex items-center gap-3">
				<!-- Active Node Indicator -->
				{#if activeNode}
					<div class="relative">
						<button
							onclick={() => showNodeDropdown = !showNodeDropdown}
							class="flex items-center gap-2 px-3 py-1.5 text-sm bg-blue-50 text-blue-700 rounded-lg hover:bg-blue-100 transition-colors"
						>
							<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
								<path d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
							{activeNode.name}
							<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
							</svg>
						</button>

						{#if showNodeDropdown}
							<div
								class="absolute right-0 top-full mt-2 w-64 bg-white border border-gray-200 rounded-xl shadow-lg p-3 z-20"
								transition:fly={{ y: -10, duration: 200 }}
							>
								<div class="text-xs font-semibold text-gray-500 uppercase mb-2">Active Node</div>
								<div class="mb-3">
									<p class="text-sm font-medium text-gray-900">{activeNode.name}</p>
									{#if activeNode.purpose}
										<p class="text-xs text-gray-500 mt-1 line-clamp-2">{activeNode.purpose}</p>
									{/if}
								</div>
								<div class="flex gap-2">
									<a
										href="/nodes/{activeNode.id}"
										class="flex-1 text-center px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										View
									</a>
									<button
										onclick={handleDeactivateNode}
										class="flex-1 px-3 py-1.5 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
									>
										Deactivate
									</button>
								</div>
							</div>
						{/if}
					</div>
				{:else}
					<a
						href="/nodes"
						class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
						No Active Node
					</a>
				{/if}

				<!-- Artifacts Toggle -->
				<button
					onclick={() => artifactsPanelOpen = !artifactsPanelOpen}
					class="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg transition-colors {artifactsPanelOpen ? 'bg-blue-100 text-blue-700' : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'}"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
					Artifacts
					{#if artifacts.length > 0}
						<span class="px-1.5 py-0.5 text-xs font-medium rounded-full {artifactsPanelOpen ? 'bg-blue-200' : 'bg-gray-200'}">{artifacts.length}</span>
					{/if}
				</button>
			</div>
		</div>

		{#if hasConversation}
			<!-- Messages container - scrollable, takes remaining height -->
			<div bind:this={messagesContainer} class="flex-1 overflow-y-auto min-h-0">
				<div class="max-w-3xl mx-auto px-4 py-6 space-y-6">
					{#each messages as message, i (message.id)}
						{@const isLastMessage = i === messages.length - 1}
						{@const parsedParts = parseMessageContent(message.content)}

						{#if message.role === 'user'}
							<!-- User message - dark bubble on right -->
							<div class="flex justify-end">
								<div class="max-w-[80%] bg-gray-900 text-white px-4 py-3 rounded-2xl rounded-br-md">
									<p class="text-[15px] leading-relaxed whitespace-pre-wrap">{message.content}</p>
								</div>
							</div>
						{:else if message.role === 'assistant'}
							<!-- Assistant message - left aligned -->
							<div class="max-w-[85%]">
								{#if parsedParts.length === 0 && isStreaming}
									<!-- Still generating artifact, show indicator -->
									<div class="flex items-center gap-2 text-sm text-gray-500">
										<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
										</svg>
										<span>Creating artifact...</span>
									</div>
								{:else}
									{#each parsedParts as part}
										{#if part.type === 'artifact' && part.artifact}
											<!-- Artifact Card - clickable to open in panel -->
											<button
												onclick={() => viewArtifactInPanel(part.artifact!)}
												class="my-3 flex items-center gap-3 px-4 py-3 bg-gradient-to-r from-blue-50 to-purple-50 border border-blue-200 rounded-xl hover:shadow-md hover:border-blue-300 transition-all cursor-pointer w-full text-left group"
											>
												<div class="w-10 h-10 rounded-lg {getArtifactColor(part.artifact.type)} flex items-center justify-center flex-shrink-0">
													<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getArtifactIcon(part.artifact.type)} />
													</svg>
												</div>
												<div class="flex-1 min-w-0">
													<p class="text-sm font-medium text-gray-900 truncate">{part.artifact.title}</p>
													<p class="text-xs text-gray-500 capitalize">{part.artifact.type} &bull; Click to view</p>
												</div>
												<svg class="w-5 h-5 text-gray-400 group-hover:text-blue-500 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
												</svg>
											</button>
										{:else if part.type === 'text' && part.text}
											<p class="text-[15px] leading-relaxed text-gray-800 whitespace-pre-wrap">{part.text}</p>
										{/if}
									{/each}
								{/if}
								{#if isLastMessage && isStreaming && parsedParts.length > 0}<span class="inline-block w-2 h-5 bg-blue-500 animate-pulse ml-1 rounded-sm"></span>{/if}
								{#if parsedParts.length > 0 && (!isStreaming || !isLastMessage)}
									<div class="flex items-center gap-2 mt-3">
										<button
											onclick={() => copyMessage(message.content, message.id)}
											class="flex items-center gap-1.5 px-2.5 py-1 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
										>
											{#if copiedMessageId === message.id}
												<svg class="w-3.5 h-3.5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
												</svg>
												<span class="text-green-600">Copied</span>
											{:else}
												<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
												</svg>
												<span>Copy</span>
											{/if}
										</button>
									</div>
								{/if}
							</div>
						{/if}
					{/each}

					{#if isStreaming && messages[messages.length - 1]?.role === 'user'}
						<!-- Typing indicator -->
						<div class="flex items-center gap-1.5">
							<div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0ms"></div>
							<div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 150ms"></div>
							<div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 300ms"></div>
						</div>
					{/if}
				</div>
			</div>

			<!-- Input Area - fixed at bottom -->
			<div class="flex-shrink-0 p-4 bg-gray-50 border-t border-gray-100">
				<div class="max-w-3xl mx-auto">
					<div class="bg-white rounded-3xl shadow-sm border border-gray-200 p-4 cursor-text" onclick={() => inputRef?.focus()}>
						<!-- Textarea -->
						<textarea
							bind:this={inputRef}
							bind:value={inputValue}
							placeholder="Ask OSA anything..."
							rows={1}
							disabled={isStreaming}
							class="w-full text-[15px] text-gray-900 placeholder-gray-400 bg-transparent resize-none focus:outline-none mb-3"
							style="min-height: 24px; max-height: 200px;"
							onkeydown={handleKeydown}
							oninput={handleInput}
						></textarea>

						<!-- Bottom row -->
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-1">
								<!-- Plus button -->
								<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" aria-label="Add">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
									</svg>
								</button>

								<!-- Attachment -->
								<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" aria-label="Attach file">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
									</svg>
								</button>

								<!-- Context selector -->
								<div class="relative">
									<button
										onclick={() => { showContextDropdown = !showContextDropdown; showModelDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
										</svg>
										{selectedContext}
									</button>

									{#if showContextDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[160px] z-10"
											transition:fly={{ y: 5, duration: 150 }}
										>
											{#each contexts as ctx}
												<button
													onclick={() => { selectedContext = ctx; showContextDropdown = false; }}
													class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors {selectedContext === ctx ? 'text-gray-900 font-medium bg-gray-50' : 'text-gray-600'}"
												>
													{ctx}
												</button>
											{/each}
										</div>
									{/if}
								</div>

								<!-- Model selector -->
								<div class="relative">
									<button
										onclick={() => { showModelDropdown = !showModelDropdown; showContextDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
										</svg>
										{currentModelName}
									</button>

									{#if showModelDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-2 min-w-[250px] z-10 max-h-80 overflow-y-auto"
											transition:fly={{ y: 5, duration: 150 }}
										>
											<!-- Cloud Models -->
											<div class="px-3 py-1.5">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Cloud</span>
											</div>
											{#each cloudModels as model}
												<button
													onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
													class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
												>
													<div class="flex items-center gap-2">
														<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
														</svg>
														<div>
															<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
															<div class="text-xs text-gray-500">{model.description}</div>
														</div>
													</div>
												</button>
											{/each}

											<div class="border-t border-gray-100 my-2"></div>

											<!-- Local Models -->
											<div class="px-3 py-1.5">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Local</span>
											</div>
											{#each localModels as model}
												<button
													onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
													class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
												>
													<div class="flex items-center gap-2">
														<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
														</svg>
														<div>
															<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
															<div class="text-xs text-gray-500">{model.description}</div>
														</div>
													</div>
												</button>
											{/each}
										</div>
									{/if}
								</div>
							</div>

							<!-- Send/Stop button -->
							{#if isStreaming}
								<button
									type="button"
									onclick={handleStop}
									class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-red-500 text-white rounded-xl hover:bg-red-600 transition-colors"
								>
									<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
										<rect x="6" y="6" width="12" height="12" rx="2" />
									</svg>
								</button>
							{:else}
								<button
									type="button"
									onclick={handleSendMessage}
									disabled={!inputValue.trim()}
									class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-blue-500 text-white rounded-xl hover:bg-blue-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
								>
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
									</svg>
								</button>
							{/if}
						</div>
					</div>
				</div>
			</div>
		{:else}
			<!-- Empty State - centered in available space -->
			<div class="flex-1 flex items-center justify-center overflow-auto">
				<div class="w-full max-w-3xl px-6">
					<!-- Title -->
					<div class="text-center mb-8">
						<h1 class="text-3xl font-semibold text-gray-900 mb-2">
							What would you like to know?
						</h1>
						<p class="text-gray-500">
							Ask <span class="text-gray-900 font-medium">OSA</span> about your business
						</p>
					</div>

					<!-- Input Box -->
					<div class="bg-white rounded-3xl shadow-lg border border-gray-200 p-4 cursor-text" onclick={() => inputRef?.focus()}>
						<textarea
							bind:this={inputRef}
							bind:value={inputValue}
							placeholder="Ask OSA anything..."
							rows={1}
							disabled={isStreaming}
							class="w-full text-[15px] text-gray-900 placeholder-gray-400 bg-transparent resize-none focus:outline-none mb-3"
							style="min-height: 24px; max-height: 200px;"
							onkeydown={handleKeydown}
							oninput={handleInput}
						></textarea>

						<div class="flex items-center justify-between">
							<div class="flex items-center gap-1">
								<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" aria-label="Add">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
									</svg>
								</button>

								<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors" aria-label="Attach file">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
									</svg>
								</button>

								<div class="relative">
									<button
										onclick={() => { showContextDropdown = !showContextDropdown; showModelDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
										</svg>
										{selectedContext}
									</button>

									{#if showContextDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[160px] z-10"
											transition:fly={{ y: 5, duration: 150 }}
										>
											{#each contexts as ctx}
												<button
													onclick={() => { selectedContext = ctx; showContextDropdown = false; }}
													class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors {selectedContext === ctx ? 'text-gray-900 font-medium bg-gray-50' : 'text-gray-600'}"
												>
													{ctx}
												</button>
											{/each}
										</div>
									{/if}
								</div>

								<div class="relative">
									<button
										onclick={() => { showModelDropdown = !showModelDropdown; showContextDropdown = false; }}
										class="flex items-center gap-1.5 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
										</svg>
										{currentModelName}
									</button>

									{#if showModelDropdown}
										<div
											class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-2 min-w-[250px] z-10 max-h-80 overflow-y-auto"
											transition:fly={{ y: 5, duration: 150 }}
										>
											<div class="px-3 py-1.5">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Cloud</span>
											</div>
											{#each cloudModels as model}
												<button
													onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
													class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
												>
													<div class="flex items-center gap-2">
														<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z" />
														</svg>
														<div>
															<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
															<div class="text-xs text-gray-500">{model.description}</div>
														</div>
													</div>
												</button>
											{/each}

											<div class="border-t border-gray-100 my-2"></div>

											<div class="px-3 py-1.5">
												<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Local</span>
											</div>
											{#each localModels as model}
												<button
													onclick={() => { selectedModel = model.id; showModelDropdown = false; }}
													class="w-full px-4 py-2 text-left hover:bg-gray-50 transition-colors {selectedModel === model.id ? 'bg-blue-50' : ''}"
												>
													<div class="flex items-center gap-2">
														<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
														</svg>
														<div>
															<div class="text-sm font-medium {selectedModel === model.id ? 'text-blue-600' : 'text-gray-700'}">{model.name}</div>
															<div class="text-xs text-gray-500">{model.description}</div>
														</div>
													</div>
												</button>
											{/each}
										</div>
									{/if}
								</div>
							</div>

							<button
								type="button"
								onclick={handleSendMessage}
								disabled={!inputValue.trim() || isStreaming}
								class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-blue-500 text-white rounded-xl hover:bg-blue-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
								</svg>
							</button>
						</div>
					</div>

					<!-- Quick Actions -->
					<div class="flex flex-wrap justify-center gap-2 mt-5">
						{#each quickActions as action}
							<button
								onclick={() => handleQuickAction(action)}
								class="px-4 py-2 bg-white border border-gray-200 rounded-full text-sm text-gray-600 hover:bg-gray-50 hover:border-gray-300 transition-all"
							>
								{action}
							</button>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Resizable Divider + Artifacts Panel -->
	{#if artifactsPanelOpen}
		<!-- Resize Handle -->
		<div
			class="w-1 h-full bg-gray-200 hover:bg-blue-400 cursor-col-resize flex-shrink-0 transition-colors relative group"
			onmousedown={startResize}
			role="separator"
			aria-orientation="vertical"
		>
			<div class="absolute inset-y-0 -left-1 -right-1 group-hover:bg-blue-400/20"></div>
			<!-- Visible grip indicator -->
			<div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 flex flex-col gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
				<div class="w-1 h-1 rounded-full bg-gray-400"></div>
				<div class="w-1 h-1 rounded-full bg-gray-400"></div>
				<div class="w-1 h-1 rounded-full bg-gray-400"></div>
			</div>
		</div>

		<div class="h-full flex flex-col bg-white flex-shrink-0" style="width: {artifactPanelWidth}px" transition:fly={{ x: 320, duration: 200 }}>
			<!-- Panel Header -->
			<div class="p-4 border-b border-gray-100 flex-shrink-0">
				<div class="flex items-center justify-between mb-3">
					<h3 class="font-semibold text-gray-900">Artifacts</h3>
					<button
						onclick={() => { artifactsPanelOpen = false; viewingArtifactFromMessage = null; }}
						class="p-1 text-gray-400 hover:text-gray-600 rounded"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<!-- Filter Tabs (only show when not viewing message artifact) -->
				{#if !viewingArtifactFromMessage}
					<div class="flex gap-1 overflow-x-auto">
						{#each ['all', 'proposal', 'sop', 'framework', 'plan', 'report'] as filter}
							<button
								onclick={() => { artifactFilter = filter; loadArtifacts(); }}
								class="px-2.5 py-1 text-xs font-medium rounded-lg whitespace-nowrap transition-colors {artifactFilter === filter ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
							>
								{filter === 'all' ? 'All' : filter.charAt(0).toUpperCase() + filter.slice(1)}
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Content Area: Generating | Message Artifact | Selected Artifact | List -->
			{#if generatingArtifact}
				<!-- Live Generation View -->
				<div class="flex-1 flex flex-col overflow-hidden">
					<!-- Generation Header -->
					<div class="p-4 border-b border-gray-100 flex-shrink-0">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-lg {generatingArtifactType ? getArtifactColor(generatingArtifactType) : 'bg-blue-50 text-blue-500'} flex items-center justify-center flex-shrink-0 relative">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={generatingArtifactType ? getArtifactIcon(generatingArtifactType) : 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'} />
								</svg>
								<!-- Generating indicator -->
								<div class="absolute -top-1 -right-1 w-3 h-3">
									<span class="absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75 animate-ping"></span>
									<span class="relative inline-flex rounded-full h-3 w-3 bg-blue-500"></span>
								</div>
							</div>
							<div class="min-w-0 flex-1">
								<h4 class="font-medium text-gray-900 truncate">
									{generatingArtifactTitle || 'Generating artifact...'}
								</h4>
								<p class="text-xs text-gray-500 flex items-center gap-1.5">
									{#if generatingArtifactType}
										<span class="capitalize">{generatingArtifactType}</span>
										<span>&bull;</span>
									{/if}
									<span class="flex items-center gap-1">
										<svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
										</svg>
										Writing...
									</span>
								</p>
							</div>
						</div>
					</div>

					<!-- Live Content Preview with Markdown -->
					<div class="flex-1 overflow-y-auto p-4 bg-gray-50">
						<div class="prose prose-sm max-w-none">
							{@html renderMarkdown(generatingArtifactContent || 'Waiting for content...')}
							<span class="inline-block w-2 h-4 bg-blue-500 animate-pulse ml-0.5"></span>
						</div>
					</div>
				</div>
			{:else if viewingArtifactFromMessage}
				<!-- Viewing artifact from message -->
				<div class="flex-1 flex flex-col overflow-hidden">
					<!-- Header -->
					<div class="p-4 border-b border-gray-100 flex-shrink-0">
						<div class="flex items-center justify-between mb-2">
							<button
								onclick={() => { viewingArtifactFromMessage = null; isEditingArtifact = false; }}
								class="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
								</svg>
								Back
							</button>
							{#if !isEditingArtifact}
								<button
									onclick={startEditingArtifact}
									class="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
									</svg>
									Edit
								</button>
							{/if}
						</div>
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-lg {getArtifactColor(viewingArtifactFromMessage.type)} flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(viewingArtifactFromMessage.type)} />
								</svg>
							</div>
							<div class="min-w-0">
								<h4 class="font-medium text-gray-900">{viewingArtifactFromMessage.title}</h4>
								<p class="text-xs text-gray-500 capitalize">{viewingArtifactFromMessage.type}</p>
							</div>
						</div>
					</div>

					<!-- Content - Editable or Rendered -->
					<div class="flex-1 overflow-y-auto p-4">
						{#if isEditingArtifact}
							<textarea
								bind:value={editedArtifactContent}
								class="w-full h-full min-h-[300px] p-3 text-sm font-mono text-gray-700 bg-white border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
								placeholder="Edit artifact content..."
							></textarea>
						{:else}
							<div
								class="prose prose-sm max-w-none cursor-text hover:bg-gray-50 rounded-lg p-2 -m-2 transition-colors"
								onclick={startEditingArtifact}
								role="button"
								tabindex="0"
								onkeydown={(e) => e.key === 'Enter' && startEditingArtifact()}
							>
								{@html renderMarkdown(viewingArtifactFromMessage.content)}
							</div>
						{/if}
					</div>

					<!-- Actions -->
					<div class="p-3 border-t border-gray-100 flex-shrink-0">
						{#if isEditingArtifact}
							<div class="flex gap-2">
								<button
									onclick={cancelArtifactEdit}
									class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
								>
									Cancel
								</button>
								<button
									onclick={saveArtifactEdit}
									class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-white bg-blue-500 hover:bg-blue-600 rounded-lg transition-colors"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									Save Changes
								</button>
							</div>
						{:else}
							<div class="flex gap-2 mb-2">
								<button
									onclick={() => { navigator.clipboard.writeText(viewingArtifactFromMessage?.content || ''); }}
									class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
									</svg>
									Copy
								</button>
								<button class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
									</svg>
									Export
								</button>
							</div>
							<button
								onclick={openSaveToNodeModal}
								class="w-full flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-white bg-gray-900 hover:bg-gray-800 rounded-lg transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
								</svg>
								Save to Node
							</button>
						{/if}
					</div>
				</div>
			{:else if selectedArtifact}
				<!-- Artifact Detail View (from API) -->
				<div class="flex-1 flex flex-col overflow-hidden">
					<!-- Detail Header -->
					<div class="p-4 border-b border-gray-100 flex-shrink-0">
						<button
							onclick={closeArtifactDetail}
							class="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700 mb-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
							</svg>
							Back
						</button>
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-lg {getArtifactColor(selectedArtifact.type)} flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(selectedArtifact.type)} />
								</svg>
							</div>
							<div class="min-w-0">
								<h4 class="font-medium text-gray-900 truncate">{selectedArtifact.title}</h4>
								<p class="text-xs text-gray-500 capitalize">{selectedArtifact.type} &bull; v{selectedArtifact.version}</p>
							</div>
						</div>
					</div>

					<!-- Content with Markdown -->
					<div class="flex-1 overflow-y-auto p-4">
						<div class="prose prose-sm max-w-none">
							{@html renderMarkdown(selectedArtifact.content)}
						</div>
					</div>

					<!-- Actions -->
					<div class="p-3 border-t border-gray-100 flex gap-2 flex-shrink-0">
						<button
							onclick={() => { navigator.clipboard.writeText(selectedArtifact?.content || ''); }}
							class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
							Copy
						</button>
						<button class="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 text-sm text-white bg-gray-900 hover:bg-gray-800 rounded-lg transition-colors">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
							</svg>
							Export
						</button>
					</div>
				</div>
			{:else}
				<!-- Artifacts List -->
				<div class="flex-1 overflow-y-auto">
					{#if loadingArtifacts}
						<div class="flex items-center justify-center h-32">
							<div class="animate-spin h-6 w-6 border-2 border-gray-900 border-t-transparent rounded-full"></div>
						</div>
					{:else if artifacts.length === 0}
						<div class="flex flex-col items-center justify-center h-48 text-center px-4">
							<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
								<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
							</div>
							<p class="text-sm text-gray-500">No artifacts yet</p>
							<p class="text-xs text-gray-400 mt-1">Ask OSA to create proposals, SOPs, or frameworks</p>
						</div>
					{:else}
						<div class="p-2 space-y-1">
							{#each artifacts as artifact (artifact.id)}
								<button
									onclick={() => selectArtifact(artifact.id)}
									class="w-full flex items-start gap-3 p-3 rounded-lg hover:bg-gray-50 transition-colors text-left"
								>
									<div class="w-9 h-9 rounded-lg {getArtifactColor(artifact.type)} flex items-center justify-center flex-shrink-0">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getArtifactIcon(artifact.type)} />
										</svg>
									</div>
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium text-gray-900 truncate">{artifact.title}</p>
										{#if artifact.summary}
											<p class="text-xs text-gray-500 line-clamp-2 mt-0.5">{artifact.summary}</p>
										{/if}
										<p class="text-xs text-gray-400 mt-1 capitalize">{artifact.type}</p>
									</div>
								</button>
							{/each}
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Click outside to close dropdowns -->
{#if showContextDropdown || showModelDropdown || showNodeDropdown}
	<button
		class="fixed inset-0 z-10 cursor-default"
		onclick={() => { showContextDropdown = false; showModelDropdown = false; showNodeDropdown = false; }}
		aria-label="Close dropdown"
	></button>
{/if}

<!-- Save to Node Modal -->
{#if showSaveToNodeModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<!-- Backdrop -->
		<button
			class="absolute inset-0 bg-black/50"
			onclick={() => showSaveToNodeModal = false}
			aria-label="Close modal"
		></button>

		<!-- Modal -->
		<div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
			<!-- Header -->
			<div class="p-4 border-b border-gray-100">
				<div class="flex items-center justify-between">
					<h3 class="text-lg font-semibold text-gray-900">Save Artifact to Node</h3>
					<button
						onclick={() => showSaveToNodeModal = false}
						class="p-1 text-gray-400 hover:text-gray-600 rounded"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
				<p class="text-sm text-gray-500 mt-1">Select a node to save this artifact to</p>
			</div>

			<!-- Content -->
			<div class="p-4 max-h-80 overflow-y-auto">
				{#if availableNodes.length === 0}
					<div class="text-center py-8">
						<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
							</svg>
						</div>
						<p class="text-sm text-gray-500">No nodes available</p>
						<a href="/nodes" class="text-sm text-blue-600 hover:underline mt-1 inline-block">Create a node first</a>
					</div>
				{:else}
					<div class="space-y-2">
						{#each availableNodes as node (node.id)}
							<button
								onclick={() => selectedNodeForSave = node.id}
								class="w-full flex items-center gap-3 p-3 rounded-xl border-2 transition-colors text-left {selectedNodeForSave === node.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
							>
								<div class="w-10 h-10 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center flex-shrink-0">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
									</svg>
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-900">{node.name}</p>
									{#if node.purpose}
										<p class="text-xs text-gray-500 truncate">{node.purpose}</p>
									{/if}
								</div>
								{#if selectedNodeForSave === node.id}
									<svg class="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 24 24">
										<path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="p-4 border-t border-gray-100 flex gap-3">
				<button
					onclick={() => showSaveToNodeModal = false}
					class="flex-1 px-4 py-2.5 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-xl transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={saveArtifactToNode}
					disabled={!selectedNodeForSave}
					class="flex-1 px-4 py-2.5 text-sm font-medium text-white bg-gray-900 hover:bg-gray-800 rounded-xl transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					Save to Node
				</button>
			</div>
		</div>
	</div>
{/if}
