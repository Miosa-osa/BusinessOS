<script lang="ts">
	import ChatMessageList from '$lib/components/chat/messages/ChatMessageList.svelte';
	import ChatInputBar from '$lib/components/chat/input/ChatInputBar.svelte';
	import ChatEmptyState from './ChatEmptyState.svelte';
	import { chatUIStore } from '$lib/stores/chat/chatUIStore.svelte';
	import { chatModelStore } from '$lib/stores/chat/chatModelStore.svelte';
	import { chatContextStore } from '$lib/stores/chat/chatContextStore.svelte';
	import { chatAgentStore } from '$lib/stores/chat/chatAgentStore.svelte';
	import { chatArtifactStore } from '$lib/stores/chat/chatArtifactStore.svelte';
	import { chatConversationStore } from '$lib/stores/chat/chatConversationStore.svelte';
	import type { StreamingToolCall } from '$lib/stores/chat/types';
	import type { SignalClassifiedEvent } from '$lib/utils/chatSSEParser';
	import type { SlashCommand, AgentPreset } from '$lib/stores/chat/types';
	import { formatTokenCount } from '../chatActions';

	const ui = chatUIStore;
	const ms = chatModelStore;
	const cx = chatContextStore;
	const ag = chatAgentStore;
	const ar = chatArtifactStore;
	const cs = chatConversationStore;

	interface Props {
		// DOM refs (must be bound from parent)
		messagesContainer: HTMLDivElement | undefined;
		inputRef: HTMLTextAreaElement | undefined;
		fileInputRef: HTMLInputElement | undefined;
		// Streaming state (local to page, passed as props)
		hasThinking: boolean;
		currentThinking: string;
		thinkingExpanded: boolean;
		streamingToolCalls: Map<string, StreamingToolCall>;
		streamingSignalMode: {
			mode: SignalClassifiedEvent['mode'];
			confidence?: number;
			genre?: SignalClassifiedEvent['genre'];
			docType?: string;
			weight?: number;
		} | null;
		showModelDropdown: boolean;
		// Computed values
		totalConversationTokens: () => number;
		contextUsagePercent: () => number;
		// Handlers
		onSendMessage: () => void;
		onStop: () => void;
		onInput: () => void;
		onKeydown: (e: KeyboardEvent) => void;
		onThinkingToggle: () => void;
		onFocusModeSubmit: (
			message: string,
			focusMode: string | null,
			options: Record<string, string>,
			files?: { id: string; name: string; type: string; size: number; content?: string }[]
		) => void;
		onSelectCommand: (cmd: SlashCommand) => void;
		onClearCommand: () => void;
		onSelectAgent: (agent: AgentPreset) => void;
		onClearAgent: () => void;
		onQuickAction: (prompt: string) => void;
		// Recorder
		recorder: {
			isRecording: boolean;
			isTranscribing: boolean;
			waveformBars: number[];
			liveTranscript: string;
			recordingTimeDisplay: string;
			toggleRecording: () => void;
			cancelRecording: () => void;
			stopRecording: () => void;
		};
	}

	let {
		messagesContainer = $bindable(),
		inputRef = $bindable(),
		fileInputRef = $bindable(),
		hasThinking,
		currentThinking,
		thinkingExpanded,
		streamingToolCalls,
		streamingSignalMode,
		showModelDropdown,
		totalConversationTokens,
		contextUsagePercent,
		onSendMessage,
		onStop,
		onInput,
		onKeydown,
		onThinkingToggle,
		onFocusModeSubmit,
		onSelectCommand,
		onClearCommand,
		onSelectAgent,
		onClearAgent,
		onQuickAction,
		recorder,
	}: Props = $props();

	// Shared input bar props (used in both conversation and empty state)
	const sharedInputProps = $derived(() => ({
		isStreaming: cs.isStreaming,
		attachedFiles: cx.attachedFiles,
		selectedMemoryIds: ag.selectedMemoryIds,
		recorderIsRecording: recorder.isRecording,
		recorderIsTranscribing: recorder.isTranscribing,
		recorderWaveformBars: recorder.waveformBars,
		recorderLiveTranscript: recorder.liveTranscript,
		recorderRecordingTimeDisplay: recorder.recordingTimeDisplay,
		showCommandSuggestions: ag.showCommandSuggestions,
		showAgentSuggestions: ag.showAgentSuggestions,
		filteredCommands: ag.filteredCommands,
		filteredAgents: ag.filteredAgents,
		commandDropdownIndex: ag.commandDropdownIndex,
		agentDropdownIndex: ag.agentDropdownIndex,
		activeCommand: ag.activeCommand,
		detectedAgent: ag.detectedAgent,
		showInlineProjectPicker: ui.showInlineProjectPicker,
		projectsList: cx.projectsList,
		projectDropdownIndex: cx.projectDropdownIndex,
		selectedFocusId: ag.selectedFocusId,
		showPlusMenu: cx.showPlusMenu,
		showContextDropdown: cx.showContextDropdown,
		showFocusDropdown: ag.showFocusDropdown,
		selectedContextsCount: cx.selectedContexts.length,
		useCOT: ms.useCOT,
		formatTokenCount,
		onSend: onSendMessage,
		onStop,
		onInput,
		onKeydown,
		onToggleRecording: recorder.toggleRecording,
		onCancelRecording: recorder.cancelRecording,
		onStopRecording: recorder.stopRecording,
		onRemoveFile: cx.removeAttachedFile,
		onClearMemories: () => { ag.selectedMemoryIds = []; ag.activeMemories = []; },
		onSelectCommand,
		onSelectAgent,
		onClearCommand,
		onClearAgent,
		onSelectProject: (id: string) => { cx.selectedProjectId = id; ui.showInlineProjectPicker = false; },
		onShowNewProjectModal: () => { ui.showInlineProjectPicker = false; cx.showNewProjectModal = true; },
		onTogglePlusMenu: (e: MouseEvent) => { e.stopPropagation(); cx.showPlusMenu = !cx.showPlusMenu; cx.showContextDropdown = false; showModelDropdown = false; },
		onAttachFile: (e: MouseEvent) => { e.stopPropagation(); fileInputRef?.click(); },
		onToggleContextDropdown: () => { cx.showContextDropdown = !cx.showContextDropdown; showModelDropdown = false; },
		onContextToggle: (id: string) => cx.handleContextToggle(id, !cx.selectedContextIds.includes(id)),
		onClearContexts: () => { cx.selectedContextIds = []; },
		onToggleFocusDropdown: () => { ag.showFocusDropdown = !ag.showFocusDropdown; showModelDropdown = false; cx.showContextDropdown = false; },
		onSelectFocusMode: (id: string | null, opts: Record<string, string>) => { ag.selectedFocusId = id; ag.focusOptions = opts; ag.showFocusDropdown = false; },
		onClearFocusMode: () => { ag.selectedFocusId = null; ag.focusOptions = {}; ag.showFocusDropdown = false; },
		onToggleCOT: ms.toggleCOT,
		onShowDocumentUpload: () => { cx.showPlusMenu = false; cx.showDocumentUploadModal = true; },
		onShowHybridSearch: () => { cx.showPlusMenu = false; cx.showHybridSearchPanel = true; },
	}));
</script>

{#if cs.hasConversation}
	<!-- Messages container -->
	<div bind:this={messagesContainer} class="flex-1 overflow-y-auto min-h-0">
		<div class="max-w-5xl mx-auto px-2 sm:px-4 py-4 sm:py-6 space-y-4 sm:space-y-6">
			<ChatMessageList
				messages={cs.messages}
				isStreaming={cs.isStreaming}
				loadingConversation={cs.loadingConversation}
				selectedModel={ms.selectedModel}
				currentModelName={ms.currentModelName}
				activeProvider={ms.activeProvider}
				warmedUpModels={ms.warmedUpModels}
				{hasThinking}
				{currentThinking}
				{thinkingExpanded}
				{streamingToolCalls}
				{streamingSignalMode}
				artifactCompletedInStream={ar.artifactCompletedInStream}
				showInlineTaskCreation={ar.showInlineTaskCreation}
				creatingInlineTasks={ar.creatingInlineTasks}
				inlineTasksForArtifact={ar.inlineTasksForArtifact}
				availableTeamMembers={cx.availableTeamMembers}
				copiedMessageId={cs.copiedMessageId}
				conversationId={cs.conversationId}
				showUsageInChat={ms.showUsageInChat}
				onThinkingToggle={onThinkingToggle}
				onViewArtifact={(a) => { ar.viewingArtifactFromMessage = a; }}
				onGenerateTasks={(a) => ar.generateTasksFromArtifact(a, cx.availableTeamMembers)}
				onSaveToProfile={(a) => { ar.viewingArtifactFromMessage = a; ar.saveArtifactToProfile(); }}
				onCopyMessage={(content, id) => cs.copyMessage(content, id)}
				onConfirmInlineTasks={() => ar.confirmInlineTasks(cx.selectedProjectId)}
				onDismissInlineTasks={() => ar.dismissInlineTasks()}
				onUpdateInlineTaskAssignee={(idx, assigneeId) => ar.updateInlineTaskAssignee(idx, assigneeId)}
				onRemoveInlineTask={(idx) => ar.removeInlineTask(idx)}
			/>
		</div>
	</div>

	<!-- Input Area -->
	<div class="chat-input-area flex-shrink-0 p-4">
		<div class="max-w-3xl mx-auto">
			<input
				bind:this={fileInputRef}
				type="file"
				multiple
				accept="image/*,.pdf,.txt,.md,.json,.csv,.doc,.docx"
				class="hidden"
				onchange={(e) => cx.handleFileSelect(e, cx.selectedProjectId)}
			/>
			<ChatInputBar
				bind:inputValue={cs.inputValue}
				bind:inputRef
				{...sharedInputProps()}
				availableContexts={cx.availableContexts.map(c => ({ id: c.id, name: c.name, icon: c.icon ?? undefined }))}
				selectedContextIds={cx.selectedContextIds}
				selectedContextsLabel={cx.selectedContextsLabel()}
				showContextStats={cs.messages.length > 0}
				totalTokens={totalConversationTokens()}
				contextLimit={ms.currentContextLimit}
				contextUsagePercent={contextUsagePercent()}
				messageTokens={cx.messageTokens(cs.messages)}
				nodeContextTokens={cx.nodeContextTokens}
				contextDocTokens={cx.contextDocTokens}
				variant="conversation"
				onNewConversation={() => { cs.startNewConversation(); cx.showPlusMenu = false; }}
			/>
		</div>
	</div>
{:else}
	<!-- Empty State -->
	<ChatEmptyState
		personalizedGreeting={ui.personalizedGreeting}
		displayedSuggestion={ui.displayedSuggestion}
		focusModeEnabled={ag.focusModeEnabled}
		availableCommands={ag.availableCommands}
		focusModeInitialInput={ag.focusModeInitialInput}
		selectedProjectId={cx.selectedProjectId}
		selectedContextIds={cx.selectedContextIds}
		availableContexts={cx.availableContexts.map(c => ({ id: c.id, name: c.name, icon: c.icon ?? undefined }))}
		onFocusModeSubmit={onFocusModeSubmit}
		onModeChange={(isFocus) => { ag.focusModeEnabled = isFocus; }}
		onRequestProjectSelect={() => { cx.showProjectDropdown = true; }}
		bind:inputValue={cs.inputValue}
		bind:inputRef
		{...sharedInputProps()}
		selectedContextsLabel={cx.selectedContextsLabel()}
		onNewConversation={() => cs.startNewConversation()}
		onSwitchToFocusMode={() => { ag.focusModeEnabled = true; }}
		onQuickAction={onQuickAction}
	/>
{/if}
