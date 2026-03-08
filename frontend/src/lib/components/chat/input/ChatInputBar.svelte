<script lang="ts">
	import { fly } from 'svelte/transition';
	import ChatVoice from './ChatVoice.svelte';
	import ChatAttachments from './ChatAttachments.svelte';
	import { FOCUS_MODES, getDefaultOptions } from './focusModes';

	// Types passed in from parent
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

	interface Props {
		// Bound value
		inputValue: string;
		// Refs (bound)
		inputRef?: HTMLTextAreaElement;
		fileInputRef?: HTMLInputElement;
		// State
		isStreaming: boolean;
		attachedFiles: AttachedFile[];
		selectedMemoryIds: string[];
		// Recorder state (from useVoiceRecorder hook)
		recorderIsRecording: boolean;
		recorderIsTranscribing: boolean;
		recorderWaveformBars: number[];
		recorderLiveTranscript: string;
		recorderRecordingTimeDisplay: string;
		// Command/agent autocomplete
		showCommandSuggestions: boolean;
		showAgentSuggestions: boolean;
		filteredCommands: SlashCommand[];
		filteredAgents: AgentPreset[];
		commandDropdownIndex: number;
		agentDropdownIndex: number;
		activeCommand: SlashCommand | null;
		detectedAgent: AgentPreset | null;
		// Project picker
		showInlineProjectPicker: boolean;
		projectsList: ProjectItem[];
		projectDropdownIndex: number;
		// Focus mode
		selectedFocusId: string | null;
		// Context selector
		availableContexts: { id: string; name: string; icon?: string | null }[];
		selectedContextIds: string[];
		selectedContextsLabel: string;
		// Context window stats
		showContextStats: boolean;
		totalTokens: number;
		contextLimit: number;
		contextUsagePercent: number;
		messageTokens: number;
		nodeContextTokens: number;
		contextDocTokens: number;
		selectedContextsCount: number;
		// COT toggle
		useCOT: boolean;
		// Variant: 'conversation' | 'empty' — empty state has slightly different container styles
		variant?: 'conversation' | 'empty';
		// Callbacks
		onSend: () => void;
		onStop: () => void;
		onInput: () => void;
		onKeydown: (e: KeyboardEvent) => void;
		onToggleRecording: () => void;
		onCancelRecording: () => void;
		onStopRecording: () => void;
		onRemoveFile: (fileId: string) => void;
		onClearMemories: () => void;
		onSelectCommand: (cmd: SlashCommand) => void;
		onSelectAgent: (agent: AgentPreset) => void;
		onClearCommand: () => void;
		onClearAgent: () => void;
		onSelectProject: (projectId: string) => void;
		onShowNewProjectModal: () => void;
		onTogglePlusMenu: (e: MouseEvent) => void;
		onAttachFile: (e: MouseEvent) => void;
		onToggleContextDropdown: () => void;
		onContextToggle: (id: string) => void;
		onClearContexts: () => void;
		onToggleFocusDropdown: () => void;
		onSelectFocusMode: (modeId: string, options: Record<string, string>) => void;
		onClearFocusMode: () => void;
		onToggleCOT: () => void;
		onShowDocumentUpload: () => void;
		onNewConversation: () => void;
		onShowHybridSearch: () => void;
		// Dropdown visibility (controlled by parent)
		showPlusMenu: boolean;
		showContextDropdown: boolean;
		showFocusDropdown: boolean;
		formatTokenCount: (n: number) => string;
	}

	let {
		inputValue = $bindable(),
		inputRef = $bindable(),
		fileInputRef = $bindable(),
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
		availableContexts,
		selectedContextIds,
		selectedContextsLabel,
		showContextStats,
		totalTokens,
		contextLimit,
		contextUsagePercent,
		messageTokens,
		nodeContextTokens,
		contextDocTokens,
		selectedContextsCount,
		useCOT,
		variant = 'conversation',
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
		onContextToggle,
		onClearContexts,
		onToggleFocusDropdown,
		onSelectFocusMode,
		onClearFocusMode,
		onToggleCOT,
		onShowDocumentUpload,
		onNewConversation,
		onShowHybridSearch,
		showPlusMenu,
		showContextDropdown,
		showFocusDropdown,
		formatTokenCount,
	}: Props = $props();

	const containerClass = $derived(
		variant === 'conversation'
			? 'chat-input-box bg-white rounded-2xl shadow-sm border border-gray-200 p-3 cursor-text'
			: 'bg-white rounded-3xl shadow-lg border border-gray-200 p-4 cursor-text'
	);
</script>

<div
	class={containerClass}
	onclick={() => inputRef?.focus()}
	role="presentation"
>
	<!-- Hidden file input (managed by parent) -->
	<input
		bind:this={fileInputRef}
		type="file"
		multiple
		accept="image/*,.pdf,.txt,.md,.json,.csv,.doc,.docx"
		class="hidden"
		aria-hidden="true"
	/>

	<!-- Attachments: files + memories -->
	<ChatAttachments
		{attachedFiles}
		{selectedMemoryIds}
		onRemoveFile={onRemoveFile}
		onClearMemories={onClearMemories}
	/>

	<!-- Voice recording overlay -->
	<ChatVoice
		isRecording={recorderIsRecording}
		isTranscribing={recorderIsTranscribing}
		waveformBars={recorderWaveformBars}
		liveTranscript={recorderLiveTranscript}
		recordingTimeDisplay={recorderRecordingTimeDisplay}
		onToggleRecording={onToggleRecording}
		onCancelRecording={onCancelRecording}
		onStopRecording={onStopRecording}
	/>

	<!-- Non-recording state -->
	{#if !recorderIsRecording && !recorderIsTranscribing}
		<!-- Inline Project Picker -->
		{#if showInlineProjectPicker}
			<div
				class="mb-3 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden"
				transition:fly={{ y: 10, duration: 150 }}
			>
				<div class="px-3 py-2 border-b border-gray-200 bg-white">
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Select a project to continue</span>
				</div>
				<div class="max-h-48 overflow-y-auto">
					{#each projectsList as project, i (project.id)}
						<button
							onclick={() => onSelectProject(project.id)}
							class="w-full px-3 py-2 text-left transition-colors flex items-center gap-3 {projectDropdownIndex === i ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-100 text-gray-700'}"
						>
							<div class="w-7 h-7 rounded-lg {projectDropdownIndex === i ? 'bg-blue-500 text-white' : 'bg-purple-100 text-purple-600'} flex items-center justify-center flex-shrink-0">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
								</svg>
							</div>
							<span class="text-sm font-medium truncate">{project.name}</span>
						</button>
					{/each}
					<button
						onclick={onShowNewProjectModal}
						class="w-full flex items-center gap-3 text-left px-3 py-2 text-sm border-t border-gray-200 transition-colors {projectDropdownIndex === projectsList.length ? 'bg-gray-100 text-gray-900' : 'hover:bg-gray-100 text-gray-700'}"
					>
						<div class="w-7 h-7 rounded-lg {projectDropdownIndex === projectsList.length ? 'bg-gray-900 text-white' : 'bg-gray-200 text-gray-500'} flex items-center justify-center flex-shrink-0">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
						</div>
						<span class="text-sm font-medium">Create new project</span>
					</button>
				</div>
				<div class="px-3 py-1.5 bg-gray-100 border-t border-gray-200 text-xs text-gray-500">
					↑↓ Navigate · Enter Select · Esc Cancel
				</div>
			</div>
		{/if}

		<!-- Slash Command Suggestions -->
		{#if showCommandSuggestions}
			<div
				class="mb-3 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden"
				transition:fly={{ y: 10, duration: 150 }}
			>
				<div class="px-3 py-2 border-b border-gray-200 bg-white">
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Commands</span>
				</div>
				<div class="max-h-64 overflow-y-auto">
					{#each filteredCommands as cmd, i (cmd.name)}
						<button
							data-command-index={i}
							onclick={() => onSelectCommand(cmd)}
							class="w-full px-3 py-2.5 text-left transition-colors flex items-center gap-3 {commandDropdownIndex === i ? 'bg-blue-50 text-blue-700' : 'hover:bg-gray-100 text-gray-700'}"
						>
							<div class="w-8 h-8 rounded-lg {commandDropdownIndex === i ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-600'} flex items-center justify-center flex-shrink-0">
								<span class="text-sm font-bold">/</span>
							</div>
							<div class="flex-1 min-w-0">
								<div class="text-sm font-medium">{cmd.display_name}</div>
								<div class="text-xs text-gray-500 truncate">{cmd.description}</div>
							</div>
							<span class="text-xs text-gray-400 font-mono">/{cmd.name}</span>
						</button>
					{/each}
				</div>
				<div class="px-3 py-1.5 bg-gray-100 border-t border-gray-200 text-xs text-gray-500">
					↑↓ Navigate · Enter/Tab Select · Esc Cancel
				</div>
			</div>
		{/if}

		<!-- Agent Mention Suggestions -->
		{#if showAgentSuggestions}
			<div
				class="mb-3 bg-gray-50 rounded-xl border border-gray-200 overflow-hidden"
				transition:fly={{ y: 10, duration: 150 }}
			>
				<div class="px-3 py-2 border-b border-gray-200 bg-white">
					<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Agents</span>
				</div>
				<div class="max-h-64 overflow-y-auto">
					{#each filteredAgents as agent, i (agent.id)}
						<button
							data-agent-index={i}
							onclick={() => onSelectAgent(agent)}
							class="w-full px-3 py-2.5 text-left transition-colors flex items-center gap-3 {agentDropdownIndex === i ? 'bg-purple-50 text-purple-700' : 'hover:bg-gray-100 text-gray-700'}"
						>
							<div class="w-8 h-8 rounded-lg {agentDropdownIndex === i ? 'bg-purple-500 text-white' : 'bg-gray-200 text-gray-600'} flex items-center justify-center flex-shrink-0">
								<span class="text-sm font-bold">@</span>
							</div>
							<div class="flex-1 min-w-0">
								<div class="text-sm font-medium">{agent.display_name}</div>
								<div class="text-xs text-gray-500 truncate">{agent.description || agent.category || 'Agent'}</div>
							</div>
							<span class="text-xs text-gray-400 font-mono">@{agent.name}</span>
						</button>
					{/each}
				</div>
				<div class="px-3 py-1.5 bg-gray-100 border-t border-gray-200 text-xs text-gray-500">
					Use arrow keys + Enter/Tab to select
				</div>
			</div>
		{/if}

		<!-- Active Command Chip -->
		{#if activeCommand}
			<div class="mb-2 flex items-center gap-2" transition:fly={{ y: -5, duration: 150 }}>
				<div class="inline-flex items-center gap-2 px-3 py-1.5 bg-blue-50 border border-blue-200 rounded-full">
					<div class="w-5 h-5 rounded bg-blue-500 text-white flex items-center justify-center">
						<span class="text-xs font-bold">/</span>
					</div>
					<span class="text-sm font-medium text-blue-700">{activeCommand.display_name}</span>
					<button
						onclick={onClearCommand}
						class="w-4 h-4 rounded-full bg-blue-200 hover:bg-blue-300 text-blue-600 flex items-center justify-center transition-colors"
						aria-label="Clear command"
					>
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
						</svg>
					</button>
				</div>
				<span class="text-xs text-gray-500">{activeCommand.description}</span>
			</div>
		{/if}

		<!-- Detected Agent Chip -->
		{#if detectedAgent}
			<div class="mb-2 flex items-center gap-2" transition:fly={{ y: -5, duration: 150 }}>
				<div class="inline-flex items-center gap-2 px-3 py-1.5 bg-purple-50 border border-purple-200 rounded-full">
					<div class="w-5 h-5 rounded bg-purple-500 text-white flex items-center justify-center">
						<span class="text-xs font-bold">@</span>
					</div>
					<span class="text-sm font-medium text-purple-700">{detectedAgent.display_name}</span>
					<button
						onclick={onClearAgent}
						class="w-4 h-4 rounded-full bg-purple-200 hover:bg-purple-300 text-purple-600 flex items-center justify-center transition-colors"
						aria-label="Clear agent"
					>
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
						</svg>
					</button>
				</div>
				<span class="text-xs text-gray-500">{detectedAgent.description || detectedAgent.category || 'Specialist'}</span>
			</div>
		{/if}

		<!-- Textarea -->
		<textarea
			bind:this={inputRef}
			bind:value={inputValue}
			placeholder="Ask OSA anything... (type / for commands)"
			rows={1}
			disabled={isStreaming}
			class="w-full text-[15px] text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 bg-transparent resize-none focus:outline-none mb-3"
			style="min-height: 24px; max-height: 200px;"
			onkeydown={onKeydown}
			oninput={onInput}
		></textarea>
	{/if}

	<!-- Bottom toolbar row -->
	<div class="flex items-center justify-between">
		<!-- Left controls -->
		<div class="flex items-center gap-1">
			<!-- Plus menu -->
			<div class="relative">
				<button
					onclick={onTogglePlusMenu}
					class="p-2 text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors rounded-lg"
					aria-label="Add"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
				</button>
				{#if showPlusMenu}
					<div
						class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[180px] z-10"
						transition:fly={{ y: 5, duration: 150 }}
					>
						{#if variant === 'conversation'}
							<button
								onclick={onNewConversation}
								class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
								New conversation
							</button>
						{/if}
						<button
							onclick={() => onToggleContextDropdown()}
							class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
							</svg>
							Add context
						</button>
						{#if variant === 'conversation'}
							<button
								onclick={onShowHybridSearch}
								class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
								</svg>
								Search knowledge
							</button>
						{/if}
						<button
							onclick={onAttachFile}
							class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
							</svg>
							Attach image
						</button>
						<button
							onclick={onShowDocumentUpload}
							class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
							</svg>
							Upload document
						</button>
					</div>
				{/if}
			</div>

			<!-- Attachment button -->
			<button
				onclick={onAttachFile}
				class="p-2 text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors rounded-lg"
				aria-label="Attach image"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
				</svg>
			</button>

			<!-- Context selector -->
			<div class="relative">
				<button
					onclick={onToggleContextDropdown}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-lg bg-transparent border border-gray-200 dark:border-[#444] text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-white/5 transition-colors"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
					</svg>
					{selectedContextIds.length > 0 ? selectedContextsLabel : 'Default'}
				</button>
				{#if showContextDropdown}
					<div
						class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[220px] z-10 max-h-64 overflow-y-auto"
						transition:fly={{ y: 5, duration: 150 }}
					>
						{#if selectedContextIds.length > 0}
							<button
								onclick={onClearContexts}
								class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700 border-b border-gray-200"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
								Clear ({selectedContextIds.length})
							</button>
						{/if}
						{#each availableContexts as ctx (ctx.id)}
							{@const isSelected = selectedContextIds.includes(ctx.id)}
							<button
								onclick={() => onContextToggle(ctx.id)}
								class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 {isSelected ? 'text-blue-600 font-medium bg-blue-50' : 'text-gray-600'}"
							>
								{#if isSelected}
									<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								{:else}
									<span class="text-base">{ctx.icon || '📄'}</span>
								{/if}
								<span class="truncate">{ctx.name}</span>
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Focus Mode selector -->
			<div class="relative">
				<button
					onclick={onToggleFocusDropdown}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {selectedFocusId ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-800 dark:text-amber-300 border border-amber-200 dark:border-amber-700' : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-[#3a3a3c]'}"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
					</svg>
					{#if selectedFocusId}
						{@const mode = FOCUS_MODES.find(m => m.id === selectedFocusId)}
						<span class="font-medium">{mode?.name || 'Focus'}</span>
					{:else}
						<span>Focus</span>
					{/if}
					<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
					</svg>
				</button>
				{#if showFocusDropdown}
					<div
						class="absolute bottom-full left-0 mb-2 bg-white border border-gray-200 rounded-xl shadow-lg py-1 min-w-[200px] z-10"
						transition:fly={{ y: 5, duration: 150 }}
					>
						{#if selectedFocusId}
							<button
								onclick={onClearFocusMode}
								class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 text-gray-700 border-b border-gray-200"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
								Clear Focus Mode
							</button>
						{/if}
						{#each FOCUS_MODES as mode (mode.id)}
							{@const isSelected = selectedFocusId === mode.id}
							<button
								onclick={() => onSelectFocusMode(mode.id, getDefaultOptions(mode))}
								class="w-full px-4 py-2 text-sm text-left hover:bg-gray-50 transition-colors flex items-center gap-2 {isSelected ? 'text-purple-600 font-medium bg-purple-50' : 'text-gray-600'}"
							>
								{#if isSelected}
									<svg class="w-4 h-4 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								{:else}
									<span class="w-4 h-4 flex items-center justify-center text-gray-400">
										{#if mode.icon === 'magnifying-glass-chart'}
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
										{:else if mode.icon === 'chart-bar'}
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
										{:else if mode.icon === 'document-text'}
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
										{:else if mode.icon === 'cube'}
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" /></svg>
										{:else}
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
										{/if}
									</span>
								{/if}
								<span class="truncate">{mode.name}</span>
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Context window indicator (conversation variant only) -->
			{#if showContextStats}
				<div class="group relative flex items-center">
					<div class="flex items-center gap-1.5 px-2 py-1 text-xs text-gray-400 hover:text-gray-600 cursor-default transition-colors">
						<span class="tabular-nums font-medium">{formatTokenCount(totalTokens)}</span>
						<span class="text-gray-300">/</span>
						<span class="tabular-nums">{formatTokenCount(contextLimit)}</span>
						{#if contextUsagePercent >= 50}
							<div class="w-12 h-1 bg-gray-200 rounded-full overflow-hidden ml-1">
								<div
									class="h-full rounded-full transition-all duration-300 {contextUsagePercent > 80 ? 'bg-red-500' : 'bg-yellow-500'}"
									style="width: {contextUsagePercent}%"
								></div>
							</div>
						{/if}
					</div>
					<!-- Tooltip -->
					<div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-3 py-2 bg-gray-900 text-white text-xs rounded-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50">
						<div class="font-medium mb-1">Context Window</div>
						<div class="text-gray-300">{totalTokens.toLocaleString()} / {contextLimit.toLocaleString()} tokens</div>
						<div class="text-gray-400 mt-1">{contextUsagePercent}% used</div>
						{#if nodeContextTokens > 0 || contextDocTokens > 0}
							<div class="border-t border-gray-700 mt-2 pt-2 text-gray-400">
								<div class="flex justify-between gap-4">
									<span>Messages:</span>
									<span>{messageTokens.toLocaleString()}</span>
								</div>
								{#if nodeContextTokens > 0}
									<div class="flex justify-between gap-4">
										<span>Node context:</span>
										<span>~{nodeContextTokens.toLocaleString()}</span>
									</div>
								{/if}
								{#if contextDocTokens > 0}
									<div class="flex justify-between gap-4">
										<span>Documents ({selectedContextsCount}):</span>
										<span>~{contextDocTokens.toLocaleString()}</span>
									</div>
								{/if}
							</div>
						{/if}
						<div class="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
					</div>
				</div>
			{/if}
		</div>

		<!-- Right controls -->
		<div class="flex items-center gap-2">
			<!-- COT Toggle (conversation variant only) -->
			{#if variant === 'conversation'}
				<button
					type="button"
					onclick={onToggleCOT}
					class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-lg transition-colors group relative {useCOT ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-800 dark:text-amber-300 border border-amber-200 dark:border-amber-700' : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-[#3a3a3c]'}"
					aria-label="Toggle Chain of Thought reasoning"
					title="{useCOT ? 'Thinking enabled' : 'Thinking disabled'}"
				>
					<svg class="w-4 h-4 {useCOT ? 'animate-pulse' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
					</svg>
					<span class="text-xs font-medium">COT</span>
					<div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50">
						{useCOT ? 'Chain of Thought: ON' : 'Chain of Thought: OFF'}
						<div class="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-900"></div>
					</div>
				</button>
			{/if}

			<!-- Voice Recording Button -->
			{#if !recorderIsRecording && !recorderIsTranscribing}
				<button
					type="button"
					onclick={(e) => { e.stopPropagation(); onToggleRecording(); }}
					class="p-2 text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors rounded-lg"
					aria-label="Voice input"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
					</svg>
				</button>
			{/if}

			<!-- Send / Stop button -->
			{#if isStreaming}
				<button
					type="button"
					onclick={onStop}
					class="p-2 rounded-lg bg-red-600 text-white hover:bg-red-700 transition-colors"
					aria-label="Stop generation"
				>
					<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
						<rect x="6" y="6" width="12" height="12" rx="2" />
					</svg>
				</button>
			{:else}
				<button
					type="button"
					onclick={onSend}
					disabled={!inputValue.trim()}
					class="send-btn p-2 rounded-lg transition-colors"
					class:send-btn-active={inputValue.trim()}
					class:send-btn-disabled={!inputValue.trim()}
					aria-label="Send message"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
					</svg>
				</button>
			{/if}
		</div>
	</div>
</div>
