<script lang="ts">
	import { onMount } from 'svelte';
	import FocusCard from './FocusCard.svelte';
	import { FOCUS_MODES, getDefaultOptions, type FocusMode } from './focusModes';

	interface SlashCommand {
		name: string;
		display_name: string;
		description: string;
		icon: string;
		category: string;
	}

	interface AttachedFile {
		id: string;
		name: string;
		type: string;
		size: number;
		content?: string; // base64 for images
	}

	interface ContextItem {
		id: string;
		name: string;
		icon?: string;
	}

	interface Props {
		onSubmit: (message: string, focusMode: string | null, focusOptions: Record<string, string>, files?: AttachedFile[]) => void;
		commands?: SlashCommand[];
		onModeChange?: (isFocusMode: boolean) => void;
		selectedProjectId?: string | null;
		onRequestProjectSelect?: () => void;
		availableContexts?: ContextItem[];
		selectedContextIds?: string[];
		onContextToggle?: (contextId: string) => void;
		initialInput?: string;
	}

	let {
		onSubmit,
		commands = [],
		onModeChange,
		selectedProjectId = null,
		onRequestProjectSelect,
		availableContexts = [],
		selectedContextIds = [],
		onContextToggle,
		initialInput = ''
	}: Props = $props();

	// React to initialInput changes (for voice transcripts)
	$effect(() => {
		if (initialInput && initialInput.trim()) {
			inputValue = initialInput;
			// Focus the input
			setTimeout(() => inputRef?.focus(), 100);
		}
	});

	// Context dropdown state
	let showContextDropdown = $state(false);

	let selectedFocusId = $state<string | null>(null);
	let focusOptions = $state<Record<string, string>>({});
	let inputValue = $state('');
	let inputRef = $state<HTMLTextAreaElement | null>(null);

	// File attachment state
	let attachedFiles = $state<AttachedFile[]>([]);
	let fileInputRef = $state<HTMLInputElement | null>(null);
	let isDragging = $state(false);

	// Command autocomplete state
	let showCommandSuggestions = $state(false);
	let filteredCommands = $state<SlashCommand[]>([]);
	let commandDropdownIndex = $state(0);
	let activeCommand = $state<SlashCommand | null>(null);

	// Derived: check if we can submit (need project selected)
	let canSubmit = $derived(!!selectedProjectId && (inputValue.trim() || attachedFiles.length > 0 || activeCommand));

	// Typewriter animation state
	const typewriterTexts = [
		'What would you like to do?',
		'Need help with research?',
		'Ready to build something?',
		'Time to write a document?',
		'Want to analyze some data?'
	];
	let typewriterIndex = $state(0);
	let typewriterCharIndex = $state(0);
	let typewriterText = $state('');
	let isDeleting = $state(false);
	let typewriterPaused = $state(false);

	onMount(() => {
		const typeSpeed = 80;
		const deleteSpeed = 40;
		const pauseTime = 2000;

		const tick = () => {
			const currentText = typewriterTexts[typewriterIndex];

			if (typewriterPaused) {
				return;
			}

			if (!isDeleting) {
				// Typing
				if (typewriterCharIndex < currentText.length) {
					typewriterText = currentText.substring(0, typewriterCharIndex + 1);
					typewriterCharIndex++;
					setTimeout(tick, typeSpeed);
				} else {
					// Finished typing, pause then delete
					typewriterPaused = true;
					setTimeout(() => {
						typewriterPaused = false;
						isDeleting = true;
						tick();
					}, pauseTime);
				}
			} else {
				// Deleting
				if (typewriterCharIndex > 0) {
					typewriterText = currentText.substring(0, typewriterCharIndex - 1);
					typewriterCharIndex--;
					setTimeout(tick, deleteSpeed);
				} else {
					// Finished deleting, move to next text
					isDeleting = false;
					typewriterIndex = (typewriterIndex + 1) % typewriterTexts.length;
					setTimeout(tick, typeSpeed);
				}
			}
		};

		tick();
	});

	// Get the selected focus mode object
	let selectedMode = $derived(
		selectedFocusId ? FOCUS_MODES.find(m => m.id === selectedFocusId) : null
	);

	// Dynamic placeholder - use typewriter when no mode selected, mode-specific otherwise
	let placeholderText = $derived(
		selectedMode
			? `Describe what you'd like to ${selectedMode.name.toLowerCase()}...`
			: typewriterText || 'What would you like to do?'
	);

	function handleSelectMode(mode: FocusMode) {
		selectedFocusId = mode.id;
		focusOptions = getDefaultOptions(mode);
	}

	function handleDeselectMode() {
		selectedFocusId = null;
		focusOptions = {};
	}

	function handleOptionChange(optionId: string, value: string) {
		focusOptions = { ...focusOptions, [optionId]: value };
	}

	function handleSubmit() {
		if (!inputValue.trim() && attachedFiles.length === 0 && !activeCommand) return;

		// Require project selection before submitting
		if (!selectedProjectId) {
			onRequestProjectSelect?.();
			return;
		}

		// Build message with command prefix if active
		let message = inputValue.trim();
		if (activeCommand) {
			message = `/${activeCommand.name} ${message}`.trim();
		}

		onSubmit(message, selectedFocusId, focusOptions, attachedFiles.length > 0 ? attachedFiles : undefined);
		inputValue = '';
		attachedFiles = [];
		activeCommand = null;  // Clear active command after submit
	}

	// File handling functions
	function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files) {
			addFiles(Array.from(target.files));
		}
		target.value = ''; // Reset input
	}

	async function addFiles(files: File[]) {
		for (const file of files) {
			const newFile: AttachedFile = {
				id: crypto.randomUUID(),
				name: file.name,
				type: file.type,
				size: file.size
			};

			// Read all files as base64 so they can be uploaded later
			// Images and binary files (like PDFs) are read as data URLs
			// Text files are also read as data URLs for consistency
			const reader = new FileReader();
			reader.onload = (e) => {
				newFile.content = e.target?.result as string;
				attachedFiles = [...attachedFiles, newFile];
				console.log('[FocusMode] File added:', file.name, 'size:', file.size, 'type:', file.type);
			};
			reader.onerror = (e) => {
				console.error('[FocusMode] Error reading file:', file.name, e);
				// Add file without content as fallback
				attachedFiles = [...attachedFiles, newFile];
			};
			// Read all files as base64 data URLs
			reader.readAsDataURL(file);
		}
	}

	function removeFile(fileId: string) {
		attachedFiles = attachedFiles.filter(f => f.id !== fileId);
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		isDragging = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		isDragging = false;
		if (e.dataTransfer?.files) {
			addFiles(Array.from(e.dataTransfer.files));
		}
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	function handleKeyDown(e: KeyboardEvent) {
		// Handle command suggestions navigation
		if (showCommandSuggestions && filteredCommands.length > 0) {
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				commandDropdownIndex = (commandDropdownIndex + 1) % filteredCommands.length;
			} else if (e.key === 'ArrowUp') {
				e.preventDefault();
				commandDropdownIndex = commandDropdownIndex <= 0 ? filteredCommands.length - 1 : commandDropdownIndex - 1;
			} else if (e.key === 'Enter' || e.key === 'Tab') {
				e.preventDefault();
				const cmd = filteredCommands[commandDropdownIndex];
				if (cmd) selectCommand(cmd);
				return;
			} else if (e.key === 'Escape') {
				showCommandSuggestions = false;
				return;
			}
		}

		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSubmit();
		}
	}

	function selectCommand(cmd: SlashCommand) {
		activeCommand = cmd;
		inputValue = '';  // Clear input, command is shown in chip
		showCommandSuggestions = false;
		inputRef?.focus();
	}

	function clearActiveCommand() {
		activeCommand = null;
		inputValue = '';
		inputRef?.focus();
	}

	function handleInput(e: Event) {
		const target = e.target as HTMLTextAreaElement;
		const value = target.value;

		// Auto-resize
		target.style.height = 'auto';
		target.style.height = Math.min(target.scrollHeight, 200) + 'px';

		// Check for slash commands
		if (value.startsWith('/') && commands.length > 0) {
			const query = value.slice(1).toLowerCase();
			if (query === '') {
				filteredCommands = commands.slice(0, 8);
			} else {
				filteredCommands = commands
					.filter(cmd => cmd.name.toLowerCase().includes(query) || cmd.display_name.toLowerCase().includes(query))
					.slice(0, 8);
			}
			showCommandSuggestions = filteredCommands.length > 0;
			commandDropdownIndex = 0;
		} else {
			showCommandSuggestions = false;
		}
	}

	// Auto-resize textarea
	function autoResize(e: Event) {
		const target = e.target as HTMLTextAreaElement;
		target.style.height = 'auto';
		target.style.height = Math.min(target.scrollHeight, 200) + 'px';
	}
</script>

<div class="focus-mode-selector">
	<!-- Header -->
	<div class="focus-header">
		<h2 class="focus-title">What's your focus?</h2>
		<p class="focus-subtitle">Choose a mode to help me assist you better</p>
	</div>

	<!-- Focus Cards Row -->
	<div class="focus-cards">
		{#each FOCUS_MODES as mode (mode.id)}
			<FocusCard
				{mode}
				isSelected={selectedFocusId === mode.id}
				onSelect={() => handleSelectMode(mode)}
				onDeselect={handleDeselectMode}
			/>
		{/each}
	</div>

	<!-- Options Panel (shown when mode is selected) -->
	{#if selectedMode && selectedMode.options.length > 0}
		<div class="options-panel">
			<div class="options-header">
				<span class="options-title">{selectedMode.name} options</span>
			</div>
			<div class="options-grid">
				{#each selectedMode.options as option}
					<div class="option-item">
						<span class="option-label">{option.label}</span>
						{#if option.type === 'segment'}
							<div class="segment-control">
								{#each option.choices || [] as choice}
									<button
										class="segment-btn"
										class:active={focusOptions[option.id] === choice.value}
										onclick={() => handleOptionChange(option.id, choice.value)}
										title={choice.tooltip || ''}
										type="button"
									>
										{choice.label}
									</button>
								{/each}
							</div>
						{:else if option.type === 'toggle'}
							<button
								class="toggle-btn"
								class:active={focusOptions[option.id] === 'on'}
								onclick={() => handleOptionChange(option.id, focusOptions[option.id] === 'on' ? 'off' : 'on')}
								type="button"
								aria-label="Toggle {option.label}"
							>
								<span class="toggle-track">
									<span class="toggle-thumb"></span>
								</span>
							</button>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Input Area -->
	<div class="input-area">
		<!-- Hidden file input -->
		<input
			bind:this={fileInputRef}
			type="file"
			multiple
			accept="image/*,.pdf,.txt,.md,.doc,.docx,.csv,.json"
			onchange={handleFileSelect}
			class="hidden"
		/>

		<div
			class="input-container {isDragging ? 'dragging' : ''}"
			ondragover={handleDragOver}
			ondragleave={handleDragLeave}
			ondrop={handleDrop}
			role="region"
			aria-label="Message input with file drop"
		>
			{#if selectedMode}
				<div class="selected-mode-badge">
					<span class="mode-name">{selectedMode.name}</span>
					<button class="mode-clear" onclick={handleDeselectMode} aria-label="Clear mode">
						<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="12" height="12">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			{/if}

			<!-- Attached files display -->
			{#if attachedFiles.length > 0}
				<div class="attached-files">
					{#each attachedFiles as file (file.id)}
						<div class="attached-file">
							{#if file.type.startsWith('image/') && file.content}
								<img src={file.content} alt={file.name} class="file-preview-img" />
							{:else}
								<div class="file-icon">
									<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" width="16" height="16">
										<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
									</svg>
								</div>
							{/if}
							<div class="file-info">
								<span class="file-name">{file.name}</span>
								<span class="file-size">{formatFileSize(file.size)}</span>
							</div>
							<button class="file-remove" onclick={() => removeFile(file.id)} aria-label="Remove file">
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="14" height="14">
									<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
								</svg>
							</button>
						</div>
					{/each}
				</div>
			{/if}

			<!-- Command Suggestions Dropdown (above input like Chat mode) -->
			{#if showCommandSuggestions && filteredCommands.length > 0}
				<div class="command-dropdown">
					<div class="command-dropdown-header">
						<span class="command-dropdown-title">Commands</span>
					</div>
					<div class="command-dropdown-list">
						{#each filteredCommands as cmd, i (cmd.name)}
							<button
								class="command-item {commandDropdownIndex === i ? 'selected' : ''}"
								onclick={() => selectCommand(cmd)}
								onmouseenter={() => commandDropdownIndex = i}
							>
								<div class="command-icon {commandDropdownIndex === i ? 'selected' : ''}">
									<span>/</span>
								</div>
								<div class="command-info">
									<span class="command-display-name">{cmd.display_name}</span>
									<span class="command-desc">{cmd.description}</span>
								</div>
								<span class="command-shortcut">/{cmd.name}</span>
							</button>
						{/each}
					</div>
					<div class="command-dropdown-footer">
						↑↓ Navigate · Enter/Tab Select · Esc Cancel
					</div>
				</div>
			{/if}

			<!-- Active Command Chip (when command is selected) -->
			{#if activeCommand}
				<div class="active-command-chip">
					<div class="active-command-badge">
						<div class="active-command-icon">
							<span>/</span>
						</div>
						<span class="active-command-name">{activeCommand.display_name}</span>
						<button
							onclick={clearActiveCommand}
							class="active-command-clear"
							aria-label="Clear command"
						>
							<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="12" height="12">
								<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
							</svg>
						</button>
					</div>
					<span class="active-command-desc">{activeCommand.description}</span>
				</div>
			{/if}

			<!-- Textarea (above controls like Chat mode) -->
			<textarea
				bind:this={inputRef}
				bind:value={inputValue}
				class="focus-input"
				placeholder={activeCommand ? `Describe your ${activeCommand.display_name.toLowerCase()} request...` : placeholderText}
				rows="1"
				oninput={handleInput}
				onkeydown={handleKeyDown}
			></textarea>

			<!-- Bottom row: left controls + right submit -->
			<div class="input-row">
				<div class="input-row-left">
					<!-- Plus button for attachments -->
					<button
						class="attach-btn"
						onclick={() => fileInputRef?.click()}
						title="Attach files"
						aria-label="Attach files"
					>
						<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="20" height="20">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
						</svg>
					</button>

					<!-- Context selector -->
					{#if availableContexts.length > 0}
						<div class="context-selector">
							<button
								class="context-btn"
								onclick={() => showContextDropdown = !showContextDropdown}
								title="Select context"
								aria-label="Select context"
							>
								<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="18" height="18">
									<path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
								</svg>
								{#if selectedContextIds.length > 0}
									<span class="context-count">{selectedContextIds.length}</span>
								{/if}
							</button>

							{#if showContextDropdown}
								<div class="context-dropdown">
									{#if selectedContextIds.length > 0}
										<button
											class="context-clear"
											onclick={() => {
												selectedContextIds.forEach(id => onContextToggle?.(id));
												showContextDropdown = false;
											}}
										>
											<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="14" height="14">
												<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
											</svg>
											Clear ({selectedContextIds.length})
										</button>
									{/if}
									{#each availableContexts as ctx (ctx.id)}
										{@const isSelected = selectedContextIds.includes(ctx.id)}
										<button
											class="context-item {isSelected ? 'selected' : ''}"
											onclick={() => onContextToggle?.(ctx.id)}
										>
											{#if isSelected}
												<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="14" height="14">
													<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
												</svg>
											{:else}
												<span class="context-icon">{ctx.icon || '📄'}</span>
											{/if}
											<span class="context-name">{ctx.name}</span>
										</button>
									{/each}
								</div>
							{/if}
						</div>
					{/if}
				</div>

				<button
					class="btn-pill btn-pill-primary"
					onclick={handleSubmit}
					disabled={!canSubmit}
					title={!selectedProjectId ? 'Select a project first' : ''}
				>
					<span>{!selectedProjectId && (inputValue.trim() || attachedFiles.length > 0) ? 'Select project' : 'Let\'s go'}</span>
					<svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
					</svg>
				</button>
			</div>

			<!-- Drag overlay -->
			{#if isDragging}
				<div class="drag-overlay">
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" width="32" height="32">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
					</svg>
					<span>Drop files here</span>
				</div>
			{/if}
		</div>

		<!-- Mode Toggle (below input) -->
		{#if onModeChange}
			<div class="mode-toggle-container">
				<button
					class="btn-pill btn-pill-secondary btn-pill-sm"
					onclick={() => onModeChange?.(false)}
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
					</svg>
					<span>Switch to Chat mode</span>
				</button>
			</div>
		{/if}
	</div>
</div>

<style>
	.focus-mode-selector {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 48px 24px;
		max-width: 800px;
		margin: 0 auto;
		gap: 32px;
	}

	.focus-header {
		text-align: center;
	}

	.focus-title {
		font-size: 28px;
		font-weight: 600;
		color: var(--color-text);
		margin: 0 0 8px 0;
	}

	:global(.dark) .focus-title {
		color: #f5f5f7;
	}

	.focus-subtitle {
		font-size: 15px;
		color: var(--color-text-secondary);
		margin: 0;
	}

	:global(.dark) .focus-subtitle {
		color: #a1a1a6;
	}

	.focus-cards {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
		justify-content: center;
		align-items: center;
		width: 100%;
	}

	/* Options Panel */
	.options-panel {
		width: 100%;
		max-width: 600px;
		background: var(--color-bg-secondary);
		border: 1px solid var(--color-border);
		border-radius: 16px;
		padding: 16px 20px;
	}

	:global(.dark) .options-panel {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
	}

	.options-header {
		margin-bottom: 16px;
	}

	.options-title {
		font-size: 13px;
		font-weight: 600;
		color: var(--color-text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	:global(.dark) .options-title {
		color: #a1a1a6;
	}

	.options-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 16px;
	}

	.option-item {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.option-label {
		font-size: 14px;
		color: var(--color-text);
		font-weight: 500;
		min-width: 80px;
	}

	:global(.dark) .option-label {
		color: #f5f5f7;
	}

	.segment-control {
		display: flex;
		background: var(--color-bg-tertiary);
		border-radius: 8px;
		padding: 3px;
	}

	:global(.dark) .segment-control {
		background: #1c1c1e;
	}

	.segment-btn {
		padding: 6px 12px;
		font-size: 13px;
		font-weight: 500;
		border: none;
		background: transparent;
		color: var(--color-text-secondary);
		cursor: pointer;
		border-radius: 6px;
		transition: all 0.15s ease;
		white-space: nowrap;
	}

	.segment-btn:hover:not(.active) {
		color: var(--color-text);
	}

	.segment-btn.active {
		background: var(--color-bg);
		color: var(--color-text);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	:global(.dark) .segment-btn {
		color: #8e8e93;
	}

	:global(.dark) .segment-btn:hover:not(.active) {
		color: #f5f5f7;
	}

	:global(.dark) .segment-btn.active {
		background: #3a3a3c;
		color: #f5f5f7;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
	}

	.toggle-btn {
		background: transparent;
		border: none;
		cursor: pointer;
		padding: 0;
	}

	.toggle-track {
		display: block;
		width: 44px;
		height: 24px;
		background: var(--color-bg-tertiary);
		border-radius: 12px;
		position: relative;
		transition: background 0.2s ease;
	}

	.toggle-btn.active .toggle-track {
		background: #34c759;
	}

	:global(.dark) .toggle-track {
		background: #3a3a3c;
	}

	:global(.dark) .toggle-btn.active .toggle-track {
		background: #30d158;
	}

	.toggle-thumb {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 20px;
		height: 20px;
		background: white;
		border-radius: 50%;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
		transition: transform 0.2s ease;
	}

	.toggle-btn.active .toggle-thumb {
		transform: translateX(20px);
	}

	.input-area {
		width: 100%;
		max-width: 640px;
	}

	.input-container {
		display: flex;
		flex-direction: column;
		gap: 12px;
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 16px;
		padding: 16px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
	}

	:global(.dark) .input-container {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
	}

	.selected-mode-badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		background: var(--color-bg-tertiary);
		padding: 4px 8px 4px 12px;
		border-radius: 20px;
		align-self: flex-start;
	}

	:global(.dark) .selected-mode-badge {
		background: #3a3a3c;
	}

	.mode-name {
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text);
	}

	:global(.dark) .mode-name {
		color: #f5f5f7;
	}

	.mode-clear {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 18px;
		height: 18px;
		border: none;
		background: transparent;
		color: var(--color-text-muted);
		cursor: pointer;
		border-radius: 50%;
		transition: all 0.15s ease;
	}

	.mode-clear:hover {
		background: var(--color-bg-secondary);
		color: var(--color-text);
	}

	:global(.dark) .mode-clear {
		color: #6e6e73;
	}

	:global(.dark) .mode-clear:hover {
		background: #48484a;
		color: #f5f5f7;
	}

	.focus-input {
		width: 100%;
		border: none;
		background: transparent;
		font-size: 16px;
		color: var(--color-text);
		resize: none;
		outline: none;
		line-height: 1.5;
		min-height: 24px;
		max-height: 200px;
	}

	.focus-input::placeholder {
		color: var(--color-text-muted);
	}

	:global(.dark) .focus-input {
		color: #f5f5f7;
	}

	:global(.dark) .focus-input::placeholder {
		color: #6e6e73;
	}


	/* Command dropdown styles - Chat mode style */
	.command-dropdown {
		background: var(--color-bg-secondary, #f9fafb);
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
	}

	:global(.dark) .command-dropdown {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
	}

	.command-dropdown-header {
		padding: 8px 12px;
		border-bottom: 1px solid var(--color-border, #e5e7eb);
		background: var(--color-bg, white);
	}

	:global(.dark) .command-dropdown-header {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.08);
	}

	.command-dropdown-title {
		font-size: 11px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--color-text-secondary, #6b7280);
	}

	:global(.dark) .command-dropdown-title {
		color: #8e8e93;
	}

	.command-dropdown-list {
		max-height: 260px;
		overflow-y: auto;
	}

	.command-dropdown-footer {
		padding: 6px 12px;
		border-top: 1px solid var(--color-border, #e5e7eb);
		background: var(--color-bg-tertiary, #f3f4f6);
		font-size: 11px;
		color: var(--color-text-muted, #9ca3af);
	}

	:global(.dark) .command-dropdown-footer {
		background: #1c1c1e;
		border-color: rgba(255, 255, 255, 0.08);
		color: #6e6e73;
	}

	.command-item {
		display: flex;
		align-items: center;
		gap: 12px;
		width: 100%;
		padding: 10px 12px;
		background: transparent;
		border: none;
		cursor: pointer;
		text-align: left;
		transition: background 0.1s ease;
	}

	.command-item:hover,
	.command-item.selected {
		background: rgba(59, 130, 246, 0.08);
	}

	:global(.dark) .command-item:hover,
	:global(.dark) .command-item.selected {
		background: rgba(10, 132, 255, 0.15);
	}

	.command-item.selected {
		color: var(--color-primary, #3b82f6);
	}

	:global(.dark) .command-item.selected {
		color: #0A84FF;
	}

	.command-icon {
		width: 32px;
		height: 32px;
		border-radius: 8px;
		background: var(--color-bg-tertiary, #e5e7eb);
		color: var(--color-text-secondary, #6b7280);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		font-size: 14px;
		font-weight: 600;
	}

	.command-icon.selected {
		background: var(--color-primary, #3b82f6);
		color: white;
	}

	:global(.dark) .command-icon {
		background: #3a3a3c;
		color: #a1a1a6;
	}

	:global(.dark) .command-icon.selected {
		background: #0A84FF;
		color: white;
	}

	.command-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.command-display-name {
		font-size: 14px;
		font-weight: 500;
		color: var(--color-text, #1f2937);
	}

	:global(.dark) .command-display-name {
		color: #f5f5f7;
	}

	.command-item.selected .command-display-name {
		color: var(--color-primary, #3b82f6);
	}

	:global(.dark) .command-item.selected .command-display-name {
		color: #0A84FF;
	}

	.command-desc {
		font-size: 12px;
		color: var(--color-text-muted, #6b7280);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	:global(.dark) .command-desc {
		color: #8e8e93;
	}

	.command-shortcut {
		font-size: 12px;
		font-family: ui-monospace, SFMono-Regular, monospace;
		color: var(--color-text-muted, #9ca3af);
		flex-shrink: 0;
	}

	:global(.dark) .command-shortcut {
		color: #6e6e73;
	}

	/* Active Command Chip */
	.active-command-chip {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.active-command-badge {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 6px 10px;
		background: rgba(59, 130, 246, 0.1);
		border: 1px solid rgba(59, 130, 246, 0.3);
		border-radius: 20px;
	}

	:global(.dark) .active-command-badge {
		background: rgba(10, 132, 255, 0.15);
		border-color: rgba(10, 132, 255, 0.3);
	}

	.active-command-icon {
		width: 20px;
		height: 20px;
		border-radius: 4px;
		background: var(--color-primary, #3b82f6);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
		font-weight: 600;
	}

	:global(.dark) .active-command-icon {
		background: #0A84FF;
	}

	.active-command-name {
		font-size: 13px;
		font-weight: 500;
		color: var(--color-primary, #3b82f6);
	}

	:global(.dark) .active-command-name {
		color: #0A84FF;
	}

	.active-command-clear {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		background: rgba(59, 130, 246, 0.2);
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--color-primary, #3b82f6);
		transition: all 0.15s ease;
	}

	.active-command-clear:hover {
		background: rgba(59, 130, 246, 0.3);
	}

	:global(.dark) .active-command-clear {
		background: rgba(10, 132, 255, 0.25);
		color: #0A84FF;
	}

	:global(.dark) .active-command-clear:hover {
		background: rgba(10, 132, 255, 0.4);
	}

	.active-command-desc {
		font-size: 12px;
		color: var(--color-text-muted, #6b7280);
	}

	:global(.dark) .active-command-desc {
		color: #8e8e93;
	}

	/* Hidden utility */
	.hidden {
		display: none;
	}

	/* Input row layout */
	.input-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		margin-top: 8px;
	}

	.input-row-left {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	/* Attach button */
	.attach-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border: none;
		background: transparent;
		color: var(--color-text-muted, #6b7280);
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.15s ease;
		flex-shrink: 0;
	}

	.attach-btn:hover {
		background: var(--color-bg-secondary, #f3f4f6);
		color: var(--color-text, #1f2937);
	}

	:global(.dark) .attach-btn {
		color: #6e6e73;
	}

	:global(.dark) .attach-btn:hover {
		background: #3a3a3c;
		color: #f5f5f7;
	}

	/* Context selector */
	.context-selector {
		position: relative;
	}

	.context-btn {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 8px;
		border: none;
		background: transparent;
		color: var(--color-text-muted, #6b7280);
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.15s ease;
	}

	.context-btn:hover {
		background: var(--color-bg-secondary, #f3f4f6);
		color: var(--color-text, #1f2937);
	}

	:global(.dark) .context-btn {
		color: #6e6e73;
	}

	:global(.dark) .context-btn:hover {
		background: #3a3a3c;
		color: #f5f5f7;
	}

	.context-count {
		font-size: 11px;
		font-weight: 600;
		background: var(--color-primary, #3b82f6);
		color: white;
		padding: 1px 5px;
		border-radius: 10px;
		min-width: 16px;
		text-align: center;
	}

	.context-dropdown {
		position: absolute;
		bottom: 100%;
		left: 0;
		margin-bottom: 8px;
		background: var(--color-bg, white);
		border: 1px solid var(--color-border, #e5e7eb);
		border-radius: 12px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
		min-width: 200px;
		max-height: 240px;
		overflow-y: auto;
		z-index: 50;
	}

	:global(.dark) .context-dropdown {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
	}

	.context-clear {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 10px 14px;
		border: none;
		background: transparent;
		color: var(--color-text-secondary, #6b7280);
		font-size: 13px;
		cursor: pointer;
		border-bottom: 1px solid var(--color-border, #e5e7eb);
		text-align: left;
	}

	.context-clear:hover {
		background: var(--color-bg-secondary, #f3f4f6);
	}

	:global(.dark) .context-clear {
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .context-clear:hover {
		background: #3a3a3c;
	}

	.context-item {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 10px 14px;
		border: none;
		background: transparent;
		color: var(--color-text, #1f2937);
		font-size: 13px;
		cursor: pointer;
		text-align: left;
	}

	.context-item:hover {
		background: var(--color-bg-secondary, #f3f4f6);
	}

	.context-item.selected {
		color: var(--color-primary, #3b82f6);
		font-weight: 500;
	}

	:global(.dark) .context-item {
		color: #f5f5f7;
	}

	:global(.dark) .context-item:hover {
		background: #3a3a3c;
	}

	:global(.dark) .context-item.selected {
		color: #0A84FF;
	}

	.context-icon {
		font-size: 14px;
	}

	.context-name {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* Attached files */
	.attached-files {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 4px;
	}

	.attached-file {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 8px;
		background: var(--color-bg-secondary, #f3f4f6);
		border-radius: 8px;
		max-width: 200px;
	}

	:global(.dark) .attached-file {
		background: #3a3a3c;
	}

	.file-preview-img {
		width: 32px;
		height: 32px;
		border-radius: 4px;
		object-fit: cover;
	}

	.file-icon {
		width: 32px;
		height: 32px;
		border-radius: 4px;
		background: var(--color-bg, white);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--color-text-muted, #6b7280);
	}

	:global(.dark) .file-icon {
		background: #2c2c2e;
		color: #a1a1a6;
	}

	.file-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.file-name {
		font-size: 12px;
		font-weight: 500;
		color: var(--color-text, #1f2937);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	:global(.dark) .file-name {
		color: #f5f5f7;
	}

	.file-size {
		font-size: 10px;
		color: var(--color-text-muted, #6b7280);
	}

	:global(.dark) .file-size {
		color: #a1a1a6;
	}

	.file-remove {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		border: none;
		background: transparent;
		color: var(--color-text-muted, #6b7280);
		cursor: pointer;
		border-radius: 4px;
		transition: all 0.15s ease;
	}

	.file-remove:hover {
		background: rgba(0, 0, 0, 0.1);
		color: #ef4444;
	}

	:global(.dark) .file-remove:hover {
		background: rgba(255, 255, 255, 0.1);
		color: #f87171;
	}

	/* Drag state */
	.input-container.dragging {
		border-color: var(--color-primary, #3b82f6);
		border-style: dashed;
		background: rgba(59, 130, 246, 0.05);
	}

	:global(.dark) .input-container.dragging {
		border-color: #0A84FF;
		background: rgba(10, 132, 255, 0.1);
	}

	/* Drag overlay */
	.drag-overlay {
		position: absolute;
		inset: 0;
		background: rgba(59, 130, 246, 0.1);
		border-radius: 16px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 8px;
		color: var(--color-primary, #3b82f6);
		font-size: 14px;
		font-weight: 500;
		pointer-events: none;
	}

	:global(.dark) .drag-overlay {
		background: rgba(10, 132, 255, 0.15);
		color: #0A84FF;
	}

	/* Make input container relative for overlay positioning */
	.input-container {
		position: relative;
	}

	/* Mode toggle */
	.mode-toggle-container {
		display: flex;
		justify-content: center;
		margin-top: 16px;
	}

	/* Responsive adjustments */
	@media (max-width: 640px) {
		.focus-mode-selector {
			padding: 24px 16px;
			gap: 24px;
		}

		.focus-title {
			font-size: 24px;
		}

		.focus-cards {
			gap: 8px;
		}
	}
</style>
