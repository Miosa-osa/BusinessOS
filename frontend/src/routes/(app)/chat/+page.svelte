<script lang="ts">
	// ── External deps ───────────────────────────────────────────────────────
	import { tick, onMount } from 'svelte';
	import { initCSRF } from '$lib/api/base';

	// ── Component imports ───────────────────────────────────────────────────
	import ChatSidebar from '$lib/components/chat/panels/ChatSidebar.svelte';
	import ChatArtifactsPanel from '$lib/components/chat/panels/ChatArtifactsPanel.svelte';
	import { useVoiceRecorder } from '$lib/hooks/useVoiceRecorder.svelte';

	// ── Route-local components ───────────────────────────────────────────────
	import ChatMainArea from './components/ChatMainArea.svelte';
	import ChatRightPanel from './components/ChatRightPanel.svelte';
	import ChatPageModals from './components/ChatPageModals.svelte';

	// ── Stores ──────────────────────────────────────────────────────────────
	import { chatUIStore } from '$lib/stores/chat/chatUIStore.svelte';
	import { chatModelStore } from '$lib/stores/chat/chatModelStore.svelte';
	import { useSession } from '$lib/auth-client';
	import { chatContextStore } from '$lib/stores/chat/chatContextStore.svelte';
	import { chatAgentStore } from '$lib/stores/chat/chatAgentStore.svelte';
	import { chatArtifactStore } from '$lib/stores/chat/chatArtifactStore.svelte';
	import { chatConversationStore } from '$lib/stores/chat/chatConversationStore.svelte';
	import type { StreamingToolCall } from '$lib/stores/chat/types';
	import { currentWorkspaceId } from '$lib/stores/workspaces';
	import { thinking, thinkingEnabled } from '$lib/stores/thinking';

	// ── Utilities ───────────────────────────────────────────────────────────
	import { ChatStreamManager } from './ChatStreamManager';
	import type { SignalClassifiedEvent } from '$lib/utils/chatSSEParser';
	import { emitModuleEvent, type ModuleEventType } from '$lib/stores/events';
	import {
		renderMarkdown,
		getArtifactIcon,
		getArtifactColor,
	} from './chatActions';
	import { computeAllArtifacts, copyToClipboard, runTypewriterEffect } from './chatPageUtils';
	import {
		checkForSpotlightTransfer,
		checkForQuickChatMessage,
		checkForVoiceTranscript,
	} from './chatSessionTransfer';
	import { createStreamCallbacks } from './chatStreamCallbacks';
	import {
		handleInput as handleInputFn,
		handleKeydown as handleKeydownFn,
		handleProjectDropdownKeydown as handleProjectDropdownKeydownFn,
		selectCommand as selectCommandFn,
		clearActiveCommand as clearActiveCommandFn,
		selectAgent as selectAgentFn,
		clearDetectedAgent as clearDetectedAgentFn,
		type InputHandlerContext,
	} from './chatInputHandlers';
	import type { SlashCommand, AgentPreset } from '$lib/stores/chat/types';

	// ── CSS ─────────────────────────────────────────────────────────────────
	import './components/chat.css';

	// ── Store aliases (short names for template readability) ─────────────────
	const ui = chatUIStore;
	const ms = chatModelStore;
	const cx = chatContextStore;
	const ag = chatAgentStore;
	const ar = chatArtifactStore;
	const cs = chatConversationStore;

	// ── DOM refs (must stay in component) ────────────────────────────────────
	let messagesContainer: HTMLDivElement | undefined = $state(undefined);
	let inputRef: HTMLTextAreaElement | undefined = $state(undefined);
	let fileInputRef: HTMLInputElement | undefined = $state(undefined);

	// ── Ephemeral streaming state (not worth extracting to a store) ──────────
	let currentThinking = $state('');
	let thinkingExpanded = $state(true);
	let hasThinking = $state(false);
	let streamingToolCalls = $state<Map<string, StreamingToolCall>>(new Map());
	let streamingSignalMode = $state<{
		mode: SignalClassifiedEvent['mode'];
		confidence?: number;
		genre?: SignalClassifiedEvent['genre'];
		docType?: string;
		weight?: number;
	} | null>(null);

	// ── Trivial toggles (kept inline) ────────────────────────────────────────
	let showModelDropdown = $state(false);

	// ── Whisper availability (informational only) ────────────────────────────
	let whisperAvailable = false;
	async function checkWhisperStatus() {
		try {
			const { apiClient } = await import('$lib/api');
			const response = await apiClient.get('/transcribe/status');
			if (response.ok) {
				const data = await response.json();
				whisperAvailable = data.available;
			}
		} catch {
			whisperAvailable = false;
		}
	}

	// ── Voice recorder ──────────────────────────────────────────────────────
	const recorder = useVoiceRecorder({
		barCount: 40,
		maxBarHeight: 24,
		onTranscription: (text) => {
			cs.inputValue = cs.inputValue ? cs.inputValue + ' ' + text : text;
		},
		onTranscriptionError: (message) => {
			cs.messages = [...cs.messages, {
				id: crypto.randomUUID(),
				role: 'assistant',
				content: `Voice transcription failed: ${message}. Make sure whisper.cpp is installed.`
			}];
		},
	});

	// ── ChatStreamManager instance ──────────────────────────────────────────
	const streamManager = new ChatStreamManager();

	// ── allArtifacts derived (from API + message artifacts) ──────────────────
	let allArtifacts = $derived.by(() => computeAllArtifacts(ar, cs));

	// ── handleSendMessage (rewired to stores + ChatStreamManager) ────────────
	async function handleSendMessage() {
		if (!cs.inputValue.trim() || cs.isStreaming) return;

		if (!cx.selectedProjectId) {
			cx.showProjectDropdown = true;
			return;
		}

		const userMessage = cs.inputValue.trim();
		cs.inputValue = '';
		if (inputRef) inputRef.style.height = 'auto';

		ar.resetStreamingArtifactState();
		currentThinking = '';
		hasThinking = false;
		streamingToolCalls = new Map();
		streamingSignalMode = null;

		let command: string | undefined;
		let messageContent = userMessage;
		const commandMatch = userMessage.match(/^\/(\w+)(?:\s+(.*))?$/s);
		if (commandMatch) {
			command = commandMatch[1];
			messageContent = commandMatch[2]?.trim() || userMessage;
		}

		const documentIds = cx.getUploadedDocumentIds();
		cx.clearAttachedFiles();

		await streamManager.send(
			{
				message: messageContent,
				model: ms.selectedModel,
				conversationId: cs.conversationId,
				projectId: cx.selectedProjectId,
				workspaceId: $currentWorkspaceId,
				contextIds: cx.selectedContextIds,
				command,
				temperature: ms.aiTemperature,
				maxTokens: ms.aiMaxTokens,
				topP: ms.aiTopP,
				focusMode: ag.selectedFocusId,
				focusOptions: ag.focusOptions,
				agentId: ag.selectedAgent?.id,
				memoryIds: ag.selectedMemoryIds,
				documentIds,
				nodeContext: cx.nodeContextPrompt ?? undefined,
				useCOT: ms.useCOT,
			},
			createStreamCallbacks(cs, cx, ms, ag, ar, ui, {
				setCurrentThinking: (v) => { currentThinking = v; },
				setHasThinking: (v) => { hasThinking = v; },
				setThinkingExpanded: (v) => { thinkingExpanded = v; },
				setStreamingToolCalls: (v) => { streamingToolCalls = v; },
				setStreamingSignalMode: (v) => { streamingSignalMode = v; },
				getCurrentThinking: () => currentThinking,
				getHasThinking: () => hasThinking,
				getStreamingToolCalls: () => streamingToolCalls,
			}),
		);

		// Notify sibling modules to refresh their data after a slash command
		// that creates entities. The map covers all entity-creating commands.
		if (command) {
			const commandEventMap: Record<string, ModuleEventType> = {
				task:    'task:created',
				tasks:   'task:created',
				todo:    'task:created',
				project: 'project:created',
				plan:    'project:created',
				client:  'client:created',
				crm:     'client:created',
				deal:    'deal:created',
				note:    'note:created',
			};
			const eventType = commandEventMap[command];
			if (eventType) {
				emitModuleEvent(eventType);
			}
		}
	}

	// ── Focus Mode submission ────────────────────────────────────────────────

	async function handleFocusModeSubmit(
		message: string,
		focusMode: string | null,
		options: Record<string, string>,
		files?: { id: string; name: string; type: string; size: number; content?: string }[]
	) {
		ag.selectedFocusId = focusMode;
		ag.focusOptions = options;
		cs.inputValue = message;

		if (files && files.length > 0) {
			cx.attachedFiles = files.map((f) => ({
				id: f.id,
				name: f.name,
				type: f.type,
				size: f.size,
				content: f.content,
				uploading: false,
			}));
		}

		handleSendMessage();
	}

	// ── Input handler context (shared across all input handlers) ────────────

	function getInputCtx(): InputHandlerContext {
		return {
			cs, cx, ag, ui,
			getInputRef: () => inputRef,
			getShowModelDropdown: () => showModelDropdown,
			handleSendMessage,
		};
	}

	// ── Input handling ───────────────────────────────────────────────────────

	function handleInput() { handleInputFn(getInputCtx()); }
	function handleKeydown(e: KeyboardEvent) { handleKeydownFn(getInputCtx(), e); }
	function handleProjectDropdownKeydown(e: KeyboardEvent) { handleProjectDropdownKeydownFn(getInputCtx(), e); }
	function selectCommand(cmd: SlashCommand) { selectCommandFn(getInputCtx(), cmd); }
	function clearActiveCommand() { clearActiveCommandFn(getInputCtx()); }
	function selectAgent(agent: AgentPreset) { selectAgentFn(getInputCtx(), agent); }
	function clearDetectedAgent() { clearDetectedAgentFn(getInputCtx()); }

	// ── Quick actions ────────────────────────────────────────────────────────

	function handleQuickAction(prompt: string) {
		cs.inputValue = prompt;
		inputRef?.focus();
	}

	function handleNewChat() {
		cs.handleNewChat();
	}

	function handleStop() {
		streamManager.stop();
	}

	// ── onMount ─────────────────────────────────────────────────────────────

	const session = useSession();

	// Keep greeting in sync with auth session (fires when session loads)
	$effect(() => {
		const name = $session.data?.user?.name?.split(' ')[0];
		if (name) ui.userName = name;
	});

	onMount(async () => {
		await initCSRF();
		ui.initFromLocalStorage();
		await ms.loadUserSettings();
		await ms.loadModels();

		if (ms.selectedModel && ms.activeProvider === 'ollama_local') {
			ms.warmupModel(ms.selectedModel);
		}

		cx.loadActiveNode();
		cx.loadContexts();
		cs.loadConversations();
		cx.loadProjects();
		cx.loadTeamMembers();
		ag.loadCommands();

		await ag.loadAgentPresets();
		await ag.loadCustomAgents();

		checkWhisperStatus();
		await thinking.loadSettings();

		const availableWidth = window.innerWidth - 256;
		ui.artifactPanelWidth = Math.floor(availableWidth / 2);

		checkForQuickChatMessage(cs, cx, ms, handleSendMessage);
		checkForVoiceTranscript(cs, ag, inputRef);
		checkForSpotlightTransfer(cs, cx);
	});

	// ── Effects ──────────────────────────────────────────────────────────────

	$effect(() => {
		void ui.rightPanelOpen;
		void ui.rightPanelTab;
		void ui.rightPanelWidth;
		ui.saveToLocalStorage();
	});

	$effect(() => {
		ms.useCOT = $thinkingEnabled;
	});

	$effect(() => {
		if (ar.artifactsPanelOpen && !ar.artifactsLoadedOnce) {
			ar.artifactsLoadedOnce = true;
			ar.loadArtifacts();
		}
	});

	$effect(() => {
		if (messagesContainer && cs.messages.length) {
			tick().then(() => {
				if (messagesContainer) {
					messagesContainer.scrollTop = messagesContainer.scrollHeight;
				}
			});
		}
	});

	// Typewriter greeting effect
	$effect(() => runTypewriterEffect(ui, cs));

	// ── Derived values ───────────────────────────────────────────────────────

	let totalConversationTokens = $derived.by(() => {
		return cx.messageTokens(cs.messages) + cx.nodeContextTokens + cx.contextDocTokens;
	});

	let contextUsagePercent = $derived.by(() => {
		const limit = ms.currentContextLimit;
		const used = totalConversationTokens;
		return Math.min(100, Math.round((used / limit) * 100));
	});
</script>

<svelte:window onkeydown={handleProjectDropdownKeydown} />

<!-- Fixed height container that fills parent -->
<div class="h-full flex overflow-hidden">
	<!-- Chat Conversations Sidebar -->
	<ChatSidebar
		open={ui.chatSidebarOpen}
		conversations={cs.conversations}
		archivedConversations={cs.archivedConversations}
		activeConversationId={cs.activeConversationId ?? ''}
		showArchivedView={cs.showArchivedView}
		bind:searchQuery={ui.searchQuery}
		conversationsPage={cs.conversationsPage}
		conversationsPageSize={cs.conversationsPageSize}
		conversationsTotal={cs.conversationsTotal}
		onselect={cs.selectConversation}
		onnewchat={handleNewChat}
		onrename={cs.handleRenameConversation}
		onpin={cs.handlePinConversation}
		ondelete={cs.handleDeleteConversation}
		onarchive={cs.handleArchiveConversation}
		onunarchive={cs.handleUnarchiveConversation}
		onexport={cs.handleExportConversation}
		onlinkproject={cs.handleLinkProject}
		onviewarchived={cs.handleViewArchived}
		onbacktochats={cs.handleBackToChats}
		onpagechange={cs.handleConversationPageChange}
		onsearchchange={(q) => { ui.searchQuery = q; }}
	/>

	<!-- Main Chat Area -->
	<div class="{ar.artifactsPanelOpen && ar.isArtifactFocused ? 'w-1/2' : 'flex-1'} flex flex-col min-w-0 h-full" style="background: var(--dbg, #fff);">
		<ChatMainArea
			bind:messagesContainer
			bind:inputRef
			bind:fileInputRef
			{hasThinking}
			{currentThinking}
			{thinkingExpanded}
			{streamingToolCalls}
			{streamingSignalMode}
			{showModelDropdown}
			{totalConversationTokens}
			{contextUsagePercent}
			onSendMessage={handleSendMessage}
			onStop={handleStop}
			onInput={handleInput}
			onKeydown={handleKeydown}
			onThinkingToggle={() => { thinkingExpanded = !thinkingExpanded; }}
			onFocusModeSubmit={handleFocusModeSubmit}
			onSelectCommand={selectCommand}
			onClearCommand={clearActiveCommand}
			onSelectAgent={selectAgent}
			onClearAgent={clearDetectedAgent}
			onQuickAction={handleQuickAction}
			recorder={{
				isRecording: recorder.isRecording,
				isTranscribing: recorder.isTranscribing,
				waveformBars: recorder.waveformBars,
				liveTranscript: recorder.liveTranscript,
				recordingTimeDisplay: recorder.recordingTimeDisplay,
				toggleRecording: recorder.toggleRecording,
				cancelRecording: recorder.cancelRecording,
				stopRecording: recorder.stopRecording,
			}}
		/>
	</div>

	<!-- Resizable Artifacts Panel -->
	{#if ar.artifactsPanelOpen}
		<ChatArtifactsPanel
			{allArtifacts}
			selectedArtifact={ar.selectedArtifact}
			viewingArtifactFromMessage={ar.viewingArtifactFromMessage}
			generatingArtifact={ar.generatingArtifact}
			generatingArtifactTitle={ar.generatingArtifactTitle}
			generatingArtifactType={ar.generatingArtifactType}
			generatingArtifactContent={ar.generatingArtifactContent}
			loadingArtifacts={ar.loadingArtifacts}
			artifactFilter={ar.artifactFilter}
			{getArtifactIcon}
			{getArtifactColor}
			{renderMarkdown}
			onClose={() => { ar.artifactsPanelOpen = false; ar.viewingArtifactFromMessage = null; }}
			onFilterChange={(f) => { ar.artifactFilter = f; ar.loadArtifacts(); }}
			onSelectArtifact={ar.selectArtifact}
			onViewMessageArtifact={(a) => { ar.viewingArtifactFromMessage = a; }}
			onDeleteArtifact={ar.deleteArtifactById}
			onSaveToProfile={ar.openSaveToProfileModal}
			onCopyContent={copyToClipboard}
			onUpdateContent={ar.updateArtifactContent}
			onUpdateSelectedContent={ar.updateSelectedArtifactContent}
			onStartResize={ui.startResize}
		/>
	{/if}

	<!-- Right Panel: Progress / Context / Artifacts -->
	{#if ui.rightPanelOpen}
		<ChatRightPanel
			rightPanelTab={ui.rightPanelTab}
			rightPanelWidth={ui.rightPanelWidth}
			delegatedTasks={ui.delegatedTasks}
			activeResources={ui.activeResources}
			{allArtifacts}
			availableContexts={cx.availableContexts.map(c => ({ id: c.id, name: c.name }))}
			selectedContextIds={cx.selectedContextIds}
			onTabChange={(tab) => { ui.rightPanelTab = tab as 'progress' | 'context' | 'artifacts'; }}
			onClose={() => { ui.rightPanelOpen = false; }}
			onStartResize={ui.startRightPanelResize}
			onContextToggle={(id) => cx.handleContextToggle(id, !cx.selectedContextIds.includes(id))}
			onMemoriesSelected={ag.handleMemoriesSelected}
			onSelectArtifact={ar.selectArtifact}
			onViewMessageArtifact={(a) => { ar.viewingArtifactFromMessage = a; }}
		/>
	{/if}
</div>

<!-- Click outside to close dropdowns -->
{#if cx.showContextDropdown || showModelDropdown || cx.showNodeDropdown || cx.showHeaderContextDropdown}
	<button
		class="fixed inset-0 z-5 cursor-default"
		onclick={() => { cx.showContextDropdown = false; showModelDropdown = false; cx.showNodeDropdown = false; cx.showHeaderContextDropdown = false; }}
		aria-label="Close dropdown"
	></button>
{/if}

<!-- Modals (save to profile, task generation, create project, document upload, hybrid search) -->
<ChatPageModals />
