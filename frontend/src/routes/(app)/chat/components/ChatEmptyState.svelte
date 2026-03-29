<script lang="ts">
	import { fade, fly } from 'svelte/transition';
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

<div class="es-outer">
	<div class="w-full max-w-3xl px-6 relative">
		<!-- Stage-light radial spotlight — shared by both chat and focus mode -->
		<div class="es-stage-light" aria-hidden="true"></div>

		{#if focusModeEnabled}
			<div in:fade={{ duration: 200 }} out:fade={{ duration: 150 }}>
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
			</div>
		{:else}
			<div in:fade={{ duration: 220 }} out:fade={{ duration: 150 }}>
				<!-- Greeting + Typewriter subtitle -->
				<div class="text-center mb-6 es-anim" style="animation-delay: 80ms">
					<h1 class="es-greeting">
						{personalizedGreeting}
					</h1>
					<p class="es-subtitle h-6">
						Let me help you <span class="es-accent">{displayedSuggestion}</span><span class="cursor-blink es-accent">|</span>
					</p>
				</div>

				<!-- Quick Actions ABOVE the input -->
				<div class="flex flex-wrap justify-center gap-2 mb-5 es-anim" style="animation-delay: 160ms">
					{#each quickActions as action}
						<button
							onclick={() => onQuickAction(action)}
							class="btn-pill btn-pill-secondary btn-pill-sm"
						>
							{action}
						</button>
					{/each}
				</div>

				<!-- Input Box -->
				<div class="es-anim" style="animation-delay: 240ms">
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
				</div>

				<!-- Switch to Focus Mode -->
				<div class="flex justify-center mt-6 es-anim" style="animation-delay: 320ms">
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
			</div>
		{/if}
	</div>
</div>

<style>
	/* Outer flex container */
	.es-outer {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: flex-start;
		overflow: auto;
		padding-top: 72px;
	}

	/* Subtle radial stage-light from top-center */
	.es-stage-light {
		position: absolute;
		top: -80px;
		left: 50%;
		transform: translateX(-50%);
		width: 720px;
		height: 420px;
		background: radial-gradient(ellipse at 50% 0%, rgba(0, 0, 0, 0.042) 0%, transparent 68%);
		pointer-events: none;
		z-index: 0;
	}
	:global(.dark) .es-stage-light {
		background: radial-gradient(ellipse at 50% 0%, rgba(255, 255, 255, 0.065) 0%, transparent 68%);
	}

	/* OSA Identity Avatar */
	.es-avatar {
		width: 54px;
		height: 54px;
		background: var(--dt);
		border-radius: 15px;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 1.25rem;
		box-shadow: 0 2px 20px rgba(0, 0, 0, 0.13);
		position: relative;
		z-index: 1;
	}

	.es-avatar-label {
		font-size: 11px;
		font-weight: 800;
		letter-spacing: 0.12em;
		color: var(--dbg);
	}

	/* H1 greeting */
	.es-greeting {
		font-size: 26px;
		font-weight: 700;
		color: var(--dt);
		margin-bottom: 0.5rem;
		letter-spacing: -0.025em;
		line-height: 1.2;
	}

	/* Subtitle */
	.es-subtitle {
		font-size: 15px;
		color: var(--dt3);
		line-height: 1.5;
	}

	/* Accent text (typewriter highlight) */
	.es-accent {
		color: var(--dt);
		font-weight: 600;
	}

	/* Staggered entrance animation */
	@keyframes esSlideUp {
		from {
			opacity: 0;
			transform: translateY(14px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.es-anim {
		animation: esSlideUp 0.5s cubic-bezier(0.22, 1, 0.36, 1) both;
	}
</style>
