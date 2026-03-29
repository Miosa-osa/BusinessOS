<script lang="ts">
	import { fly, fade, scale } from 'svelte/transition';
	import { getFileIcon, formatFileSize } from './spotlightActions.ts';
	import type { AttachedFile } from './spotlightActions.ts';
	import type { SlashCommand, Project, AIModel } from './spotlightSearch.ts';
	import type { useVoiceRecorder } from '$lib/hooks/useVoiceRecorder.svelte';

	type VoiceRecorder = ReturnType<typeof useVoiceRecorder>;

	interface Props {
		inputValue: string;
		onInputChange: (value: string) => void;
		inputElement?: HTMLTextAreaElement;
		onKeyDown: (event: KeyboardEvent) => void;

		// Slash commands
		filteredCommands: SlashCommand[];
		showCommandsDropdown: boolean;
		commandDropdownIndex: number;
		onSelectCommand: (cmd: SlashCommand) => void;
		onCommandHover: (index: number) => void;

		// File attachments
		attachedFiles: AttachedFile[];
		fileInputRef?: HTMLInputElement;
		onFileSelect: (event: Event) => void;
		onRemoveFile: (fileId: string) => void;
		isDragging: boolean;
		onDragEnter: (e: DragEvent) => void;
		onDragLeave: (e: DragEvent) => void;
		onDragOver: (e: DragEvent) => void;
		onDrop: (e: DragEvent) => void;

		// Voice
		recorder: VoiceRecorder;

		// Project selector
		projectsList: Project[];
		selectedProjectId: string | null;
		showProjectDropdown: boolean;
		projectDropdownIndex: number;
		onProjectSelect: (id: string) => void;
		onToggleProjectDropdown: () => void;
		onProjectHover: (index: number) => void;
		onCloseAndNavigate: (appId: string) => void;

		// Model selector
		availableModels: AIModel[];
		selectedModel: string;
		activeProvider: string;
		showModelDropdown: boolean;
		currentModelName: string;
		onModelSelect: (id: string) => void;
		onToggleModelDropdown: () => void;

		// Send
		canSend: boolean;
		onSend: () => void;
	}

	let {
		inputValue,
		onInputChange,
		inputElement = $bindable(),
		onKeyDown,
		filteredCommands,
		showCommandsDropdown,
		commandDropdownIndex,
		onSelectCommand,
		onCommandHover,
		attachedFiles,
		fileInputRef = $bindable(),
		onFileSelect,
		onRemoveFile,
		isDragging,
		onDragEnter,
		onDragLeave,
		onDragOver,
		onDrop,
		recorder,
		projectsList,
		selectedProjectId,
		showProjectDropdown,
		projectDropdownIndex,
		onProjectSelect,
		onToggleProjectDropdown,
		onProjectHover,
		onCloseAndNavigate,
		availableModels,
		selectedModel,
		activeProvider,
		showModelDropdown,
		currentModelName,
		onModelSelect,
		onToggleModelDropdown,
		canSend,
		onSend
	}: Props = $props();

	let selectedProject = $derived(
		selectedProjectId ? projectsList.find((p) => p.id === selectedProjectId) : null
	);

	function handleInput(e: Event) {
		const ta = e.target as HTMLTextAreaElement;
		onInputChange(ta.value);
		// Auto-resize
		ta.style.height = 'auto';
		ta.style.height = Math.min(ta.scrollHeight, 120) + 'px';
	}
</script>

<div
	class="input-card"
	class:dragging={isDragging}
	ondragenter={onDragEnter}
	ondragleave={onDragLeave}
	ondragover={onDragOver}
	ondrop={onDrop}
>
	<!-- Hidden file input -->
	<input
		type="file"
		bind:this={fileInputRef}
		onchange={onFileSelect}
		multiple
		accept="image/*,.pdf,.doc,.docx,.txt,.csv,.json,.md"
		style="display: none;"
	/>

	<!-- Drag overlay -->
	{#if isDragging}
		<div class="drag-overlay" transition:fade={{ duration: 100 }}>
			<div class="drag-content">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" />
					<polyline points="17 8 12 3 7 8" />
					<line x1="12" y1="3" x2="12" y2="15" />
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
					<button
						class="attachment-remove"
						onclick={() => onRemoveFile(file.id)}
						aria-label="Remove {file.name}"
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
						</svg>
					</button>
				</div>
			{/each}
		</div>
	{/if}

	{#if recorder.isRecording}
		<!-- Recording UI -->
		<div class="recording-area">
			{#if recorder.liveTranscript}
				<div class="live-transcript">{recorder.liveTranscript}</div>
			{:else}
				<div class="live-transcript placeholder">Listening...</div>
			{/if}
			<div class="waveform-bar">
				<button class="cancel-btn" onclick={recorder.cancelRecording} title="Cancel">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
					</svg>
				</button>
				<div class="waveform">
					{#each recorder.waveformBars as height}
						<div class="bar" style="height: {height}px"></div>
					{/each}
				</div>
				<span class="duration">{recorder.recordingTimeDisplay}</span>
				<button class="confirm-btn" onclick={recorder.stopRecording} title="Done">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polyline points="20 6 9 17 4 12" />
					</svg>
				</button>
			</div>
		</div>
	{:else}
		<!-- Textarea -->
		<textarea
			bind:this={inputElement}
			value={inputValue}
			placeholder="Ask anything... (type / for commands)"
			rows={1}
			onkeydown={onKeyDown}
			oninput={handleInput}
			aria-label="Chat input"
		></textarea>

		<!-- Commands dropdown -->
		{#if showCommandsDropdown && filteredCommands.length > 0}
			<div class="commands-dropdown" transition:fly={{ y: 5, duration: 150 }}>
				{#each filteredCommands as cmd, i (cmd.id)}
					<button
						class="command-item"
						class:highlighted={commandDropdownIndex === i}
						onclick={() => onSelectCommand(cmd)}
						onmouseenter={() => onCommandHover(i)}
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
			<button
				class="icon-btn"
				title="Attach file"
				onclick={() => fileInputRef?.click()}
				aria-label="Attach file"
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path
						d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
					/>
				</svg>
			</button>

			<!-- Project Selector -->
			<div class="dropdown-wrapper">
				<button
					class="selector-btn"
					class:selected={selectedProject}
					class:required={!selectedProject && inputValue.trim()}
					onclick={onToggleProjectDropdown}
					aria-label={selectedProject ? selectedProject.name : 'Select project'}
					aria-expanded={showProjectDropdown}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
					<span>{selectedProject ? selectedProject.name : 'Project'}</span>
					{#if !selectedProject}
						<span class="required-dot"></span>
					{/if}
					<svg class="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
								onclick={() => onProjectSelect(project.id)}
								onmouseenter={() => onProjectHover(i)}
							>
								{project.name}
							</button>
						{/each}
						<button
							class="dropdown-item create-new"
							class:highlighted={projectDropdownIndex === projectsList.length}
							onclick={() => onCloseAndNavigate('projects')}
							onmouseenter={() => onProjectHover(projectsList.length)}
						>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
							</svg>
							New Project
						</button>
					</div>
				{/if}
			</div>

			<!-- Model Selector -->
			<div class="dropdown-wrapper">
				<button
					class="icon-btn"
					onclick={onToggleModelDropdown}
					title={currentModelName}
					aria-label="Select AI model: {currentModelName}"
					aria-expanded={showModelDropdown}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="4" y="4" width="16" height="16" rx="2" />
						<circle cx="9" cy="9" r="1.5" fill="currentColor" />
						<circle cx="15" cy="9" r="1.5" fill="currentColor" />
						<path d="M9 15h6" />
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
									onclick={() => onModelSelect(model.id)}
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
			<div class="hints">
				<span><kbd>⌘D</kbd> Voice</span>
				<span><kbd>↵</kbd> Send</span>
			</div>

			<!-- Voice Button -->
			<button
				class="icon-btn mic"
				class:recording={recorder.isRecording}
				onclick={recorder.toggleRecording}
				title="Voice input"
				aria-label={recorder.isRecording ? 'Stop recording' : 'Start voice input'}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 1a3 3 0 00-3 3v8a3 3 0 006 0V4a3 3 0 00-3-3z" />
					<path d="M19 10v2a7 7 0 01-14 0v-2" />
					<line x1="12" y1="19" x2="12" y2="23" />
					<line x1="8" y1="23" x2="16" y2="23" />
				</svg>
			</button>

			<!-- Send Button -->
			<button
				class="send-btn"
				onclick={onSend}
				disabled={!canSend}
				title={!selectedProjectId ? 'Select a project first' : 'Send message'}
				aria-label="Send message"
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M5 10l7-7m0 0l7 7m-7-7v18" />
				</svg>
			</button>
		</div>
	</div>
</div>

<style>
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

	.cancel-btn,
	.confirm-btn {
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

	.cancel-btn svg,
	.confirm-btn svg {
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

	.left-controls,
	.right-controls {
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
		0%,
		100% {
			box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.4);
		}
		50% {
			box-shadow: 0 0 0 8px rgba(239, 68, 68, 0);
		}
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

	/* Send Button */
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

	/* ===== DARK MODE ===== */
	:global(.dark) .input-card {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .input-card.dragging {
		border-color: #0a84ff;
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
		color: #0a84ff;
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
		color: #0a84ff;
	}

	:global(.dark) .dropdown-item.create-new {
		border-top-color: rgba(255, 255, 255, 0.1);
		color: #0a84ff;
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
		background: #0a84ff;
	}

	:global(.dark) .send-btn:hover:not(:disabled) {
		background: #0070e0;
	}

	:global(.dark) .send-btn:disabled {
		background: #3a3a3c;
		color: #6e6e73;
	}
</style>
