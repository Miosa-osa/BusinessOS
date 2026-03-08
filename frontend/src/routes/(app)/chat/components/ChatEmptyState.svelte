<script lang="ts">
	import FocusModeSelector from '$lib/components/chat/focus/FocusModeSelector.svelte';
	import ChatInputBar from '$lib/components/chat/input/ChatInputBar.svelte';
	import { quickActions } from '../chatActions';

	interface AttachedFile {
		id: string;
		name: string;
		type: string;
		size: number;
		content?: string;
		documentId?: string;
		uploading?: boolean;
		uploadError?: string;
	}

	interface SlashCommand {
		name: string;
		display_name: string;
		description: string;
		icon: string;
		category: string;
	}

	interface AgentPreset {
		id: string;
		name: string;
		display_name: string;
		description: string | null;
		avatar: string | null;
		category: string | null;
	}

	interface ProjectItem {
		id: string;
		name: string;
		description?: string;
	}

	interface ContextItem {
		id: string;
		name: string;
		icon?: string;
	}

	interface Props {
		// Greeting
		personalizedGreeting: string;
		displayedSuggestion: string;

		// Focus mode
		focusModeEnabled: boolean;
		availableCommands: SlashCommand[];
		focusModeInitialInput: string;
		selectedProjectId: string | null;
		selectedContextIds: string[];
		availableContexts: ContextItem[];
		onFocusModeSubmit: (
			message: string,
			focusMode: string | null,
			options: Record<string, string>,
			files?: { id: string; name: string; type: string; size: number; content?: string }[]
		) => void;
		onModeChange: (isFocus: boolean) => void;
		onRequestProjectSelect: () => void;
		onContextToggle: (id: string) => void;

		// Input bar (empty variant)
		inputValue: string;
		inputRef?: HTMLTextAreaElement;
		isStreaming: boolean;
		attachedFiles: AttachedFile[];
		selectedMemoryIds: string[];
		recorderIsRecording: boolean;
		recorderIsTranscribing: boolean;
		recorderWaveformBars: number[];
		recorderLiveTranscript: string;
		recorderRecordingTimeDisplay: string;
		showCommandSuggestions: boolean;
		showAgentSuggestions: boolean;
		filteredCommands: SlashCommand[];
		filteredAgents: AgentPreset[];
		commandDropdownIndex: number;
		agentDropdownIndex: number;
		activeCommand: SlashCommand | null;
		detectedAgent: AgentPreset | null;
		showInlineProjectPicker: boolean;
		projectsList: ProjectItem[];
		projectDropdownIndex: number;
		selectedFocusId: string | null;
		selectedContextsLabel: string;
		useCOT: boolean;
		showPlusMenu: boolean;
		showContextDropdown: boolean;
		showFocusDropdown: boolean;
		formatTokenCount: (n: number) => string;
		selectedContextsCount: number;

		// Callbacks
		onSend: () => void;
		onStop: () => void;
		onInput: () => void;
		onKeydown: (e: KeyboardEvent) => void;
		onToggleRecording: () => void;
		onCancelRecording: () => void;
		onStopRecording: () => void;
		onRemoveFile: (id: string) => void;
		onClearMemories: () => void;
		onSelectCommand: (cmd: SlashCommand) => void;
		onSelectAgent: (agent: AgentPreset) => void;
		onClearCommand: () => void;
		onClearAgent: () => void;
		onSelectProject: (id: string) => void;
		onShowNewProjectModal: () => void;
		onTogglePlusMenu: (e: MouseEvent) => void;
		onAttachFile: (e: MouseEvent) => void;
		onToggleContextDropdown: () => void;
		onClearContexts: () => void;
		onToggleFocusDropdown: () => void;
		onSelectFocusMode: (id: string | null, opts: Record<string, string>) => void;
		onClearFocusMode: () => void;
		onToggleCOT: () => void;
		onShowDocumentUpload: () => void;
		onNewConversation: () => void;
		onShowHybridSearch: () => void;
		onSwitchToFocusMode: () => void;
		onQuickAction: (prompt: string) => void;
	}

	let {
		personalizedGreeting,
		displayedSuggestion,
		focusModeEnabled,
		availableCommands,
		focusModeInitialInput,
		selectedProjectId,
		selectedContextIds,
		availableContexts,
		onFocusModeSubmit,
		onModeChange,
		onRequestProjectSelect,
		onContextToggle,
		inputValue = $bindable(),
		inputRef = $bindable(),
		isStreaming,
		attachedFiles,
		selectedMemoryIds,
		recorderIsRecording,
		recorderIsTranscribing,
		recorderWaveformBars,
		recorderLiveTranscript,
		recorderRecordingTimeDisplay,
		showCommandSuggestions,
		showAgentSuggestions,
		filteredCommands,
		filteredAgents,
		commandDropdownIndex,
		agentDropdownIndex,
		activeCommand,
		detectedAgent,
		showInlineProjectPicker,
		projectsList,
		projectDropdownIndex,
		selectedFocusId,
		selectedContextsLabel,
		useCOT,
		showPlusMenu,
		showContextDropdown,
		showFocusDropdown,
		formatTokenCount,
		selectedContextsCount,
		onSend,
		onStop,
		onInput,
		onKeydown,
		onToggleRecording,
		onCancelRecording,
		onStopRecording,
		onRemoveFile,
		onClearMemories,
		onSelectCommand,
		onSelectAgent,
		onClearCommand,
		onClearAgent,
		onSelectProject,
		onShowNewProjectModal,
		onTogglePlusMenu,
		onAttachFile,
		onToggleContextDropdown,
		onClearContexts,
		onToggleFocusDropdown,
		onSelectFocusMode,
		onClearFocusMode,
		onToggleCOT,
		onShowDocumentUpload,
		onNewConversation,
		onShowHybridSearch,
		onSwitchToFocusMode,
		onQuickAction,
	}: Props = $props();
</script>

<div class="flex-1 flex items-center justify-center overflow-auto">
	<div class="w-full max-w-3xl px-6">
		{#if focusModeEnabled}
			<!-- Focus Mode UI -->
			<FocusModeSelector
				onSubmit={onFocusModeSubmit}
				commands={availableCommands}
				onModeChange={(isFocus: boolean) => onModeChange(isFocus)}
				{selectedProjectId}
				onRequestProjectSelect={onRequestProjectSelect}
				{availableContexts}
				{selectedContextIds}
				{onContextToggle}
				initialInput={focusModeInitialInput}
			/>
		{:else}
			<!-- Classic Mode - Personalized Title -->
			<div class="text-center mb-8">
				<h1 class="text-3xl font-semibold text-gray-900 dark:text-white mb-3">
					{personalizedGreeting}
				</h1>
				<p class="text-gray-500 h-6">
					Let me help you <span class="text-blue-600 font-medium">{displayedSuggestion}</span><span class="cursor-blink text-blue-600 font-light">|</span>
				</p>
			</div>

			<!-- Input Box (Classic Mode) -->
			<ChatInputBar
				bind:inputValue
				bind:inputRef
				{isStreaming}
				{attachedFiles}
				{selectedMemoryIds}
				{recorderIsRecording}
				{recorderIsTranscribing}
				{recorderWaveformBars}
				{recorderLiveTranscript}
				{recorderRecordingTimeDisplay}
				{showCommandSuggestions}
				{showAgentSuggestions}
				{filteredCommands}
				{filteredAgents}
				{commandDropdownIndex}
				{agentDropdownIndex}
				{activeCommand}
				{detectedAgent}
				{showInlineProjectPicker}
				{projectsList}
				{projectDropdownIndex}
				{selectedFocusId}
				{availableContexts}
				{selectedContextIds}
				selectedContextsLabel={selectedContextsLabel}
				showContextStats={false}
				totalTokens={0}
				contextLimit={0}
				contextUsagePercent={0}
				messageTokens={0}
				nodeContextTokens={0}
				contextDocTokens={0}
				{selectedContextsCount}
				{useCOT}
				variant="empty"
				{showPlusMenu}
				{showContextDropdown}
				{showFocusDropdown}
				{formatTokenCount}
				onSend={onSend}
				onStop={onStop}
				onInput={onInput}
				onKeydown={onKeydown}
				onToggleRecording={onToggleRecording}
				onCancelRecording={onCancelRecording}
				onStopRecording={onStopRecording}
				onRemoveFile={onRemoveFile}
				onClearMemories={onClearMemories}
				onSelectCommand={onSelectCommand}
				onSelectAgent={onSelectAgent}
				onClearCommand={onClearCommand}
				onClearAgent={onClearAgent}
				onSelectProject={(id) => onSelectProject(id)}
				onShowNewProjectModal={onShowNewProjectModal}
				onTogglePlusMenu={onTogglePlusMenu}
				onAttachFile={onAttachFile}
				onToggleContextDropdown={onToggleContextDropdown}
				onContextToggle={(id) => onContextToggle(id)}
				onClearContexts={onClearContexts}
				onToggleFocusDropdown={onToggleFocusDropdown}
				onSelectFocusMode={(id, opts) => onSelectFocusMode(id, opts)}
				onClearFocusMode={onClearFocusMode}
				onToggleCOT={onToggleCOT}
				onShowDocumentUpload={onShowDocumentUpload}
				onNewConversation={onNewConversation}
				onShowHybridSearch={onShowHybridSearch}
			/>

			<!-- Quick Actions (Classic Mode only) -->
			<div class="flex flex-wrap justify-center gap-2 mt-5">
				{#each quickActions as action}
					<button
						onclick={() => onQuickAction(action)}
						class="btn-pill btn-pill-secondary btn-pill-sm"
					>
						{action}
					</button>
				{/each}
			</div>

			<!-- Switch to Focus Mode -->
			<div class="flex justify-center mt-6">
				<button
					onclick={onSwitchToFocusMode}
					class="btn-pill btn-pill-ghost btn-pill-sm"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
					</svg>
					Switch to Focus mode
				</button>
			</div>
		{/if}
	</div>
</div>
