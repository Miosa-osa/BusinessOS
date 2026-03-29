<script lang="ts">
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
		content?: string;
	}

	interface ContextItem {
		id: string;
		name: string;
		icon?: string;
	}

	interface Props {
		// Input state (bindable)
		inputValue: string;
		attachedFiles: AttachedFile[];
		// Mode badge
		selectedModeName?: string | null;
		onClearMode?: () => void;
		// Submit
		canSubmit: boolean;
		selectedProjectId?: string | null;
		onSubmit: () => void;
		// Commands
		commands?: SlashCommand[];
		activeCommand?: SlashCommand | null;
		onClearCommand?: () => void;
		// Context
		availableContexts?: ContextItem[];
		selectedContextIds?: string[];
		onContextToggle?: (contextId: string) => void;
		// Placeholder
		placeholderText?: string;
		// Input ref for parent focus control
		inputRef?: HTMLTextAreaElement | null;
	}

	let {
		inputValue = $bindable(),
		attachedFiles = $bindable(),
		selectedModeName = null,
		onClearMode,
		canSubmit,
		selectedProjectId = null,
		onSubmit,
		commands = [],
		activeCommand = $bindable(null),
		onClearCommand,
		availableContexts = [],
		selectedContextIds = [],
		onContextToggle,
		placeholderText = 'What would you like to do?',
		inputRef = $bindable(null)
	}: Props = $props();

	let fileInputRef = $state<HTMLInputElement | null>(null);
	let isDragging = $state(false);
	let showContextDropdown = $state(false);
	let showCommandSuggestions = $state(false);
	let filteredCommands = $state<SlashCommand[]>([]);
	let commandDropdownIndex = $state(0);

	function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files) {
			addFiles(Array.from(target.files));
		}
		target.value = '';
	}

	async function addFiles(files: File[]) {
		for (const file of files) {
			const newFile: AttachedFile = {
				id: crypto.randomUUID(),
				name: file.name,
				type: file.type,
				size: file.size
			};
			const reader = new FileReader();
			reader.onload = (e) => {
				newFile.content = e.target?.result as string;
				attachedFiles = [...attachedFiles, newFile];
				console.log('[FocusInputArea] File added:', file.name, 'size:', file.size, 'type:', file.type);
			};
			reader.onerror = (e) => {
				console.error('[FocusInputArea] Error reading file:', file.name, e);
				attachedFiles = [...attachedFiles, newFile];
			};
			reader.readAsDataURL(file);
		}
	}

	function removeFile(fileId: string) {
		attachedFiles = attachedFiles.filter(f => f.id !== fileId);
	}

	function formatFileSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
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

	function selectCommand(cmd: SlashCommand) {
		activeCommand = cmd;
		inputValue = '';
		showCommandSuggestions = false;
		inputRef?.focus();
	}

	function clearActiveCommand() {
		onClearCommand?.();
		inputRef?.focus();
	}

	function handleInput(e: Event) {
		const target = e.target as HTMLTextAreaElement;
		const value = target.value;

		// Auto-resize
		target.style.height = 'auto';
		target.style.height = Math.min(target.scrollHeight, 200) + 'px';

		// Slash command detection
		if (value.startsWith('/') && commands.length > 0) {
			const query = value.slice(1).toLowerCase();
			filteredCommands = query === ''
				? commands.slice(0, 8)
				: commands
					.filter(cmd =>
						cmd.name.toLowerCase().includes(query) ||
						cmd.display_name.toLowerCase().includes(query)
					)
					.slice(0, 8);
			showCommandSuggestions = filteredCommands.length > 0;
			commandDropdownIndex = 0;
		} else {
			showCommandSuggestions = false;
		}
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (showCommandSuggestions && filteredCommands.length > 0) {
			if (e.key === 'ArrowDown') {
				e.preventDefault();
				commandDropdownIndex = (commandDropdownIndex + 1) % filteredCommands.length;
			} else if (e.key === 'ArrowUp') {
				e.preventDefault();
				commandDropdownIndex = commandDropdownIndex <= 0
					? filteredCommands.length - 1
					: commandDropdownIndex - 1;
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
			onSubmit();
		}
	}
</script>

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
	class="input-container"
	class:dragging={isDragging}
	ondragover={handleDragOver}
	ondragleave={handleDragLeave}
	ondrop={handleDrop}
	role="region"
	aria-label="Message input with file drop"
>
	<!-- Selected mode badge -->
	{#if selectedModeName}
		<div class="selected-mode-badge">
			<span class="mode-name">{selectedModeName}</span>
			<button class="mode-clear" onclick={onClearMode} aria-label="Clear mode">
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="12" height="12">
					<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
				</svg>
			</button>
		</div>
	{/if}

	<!-- Attached files -->
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

	<!-- Command suggestions dropdown -->
	{#if showCommandSuggestions && filteredCommands.length > 0}
		<div class="command-dropdown">
			<div class="command-dropdown-header">
				<span class="command-dropdown-title">Commands</span>
			</div>
			<div class="command-dropdown-list">
				{#each filteredCommands as cmd, i (cmd.name)}
					<button
						class="command-item"
						class:selected={commandDropdownIndex === i}
						onclick={() => selectCommand(cmd)}
						onmouseenter={() => commandDropdownIndex = i}
					>
						<div class="command-icon" class:selected={commandDropdownIndex === i}>
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

	<!-- Active command chip -->
	{#if activeCommand}
		<div class="active-command-chip">
			<div class="active-command-badge">
				<div class="active-command-icon"><span>/</span></div>
				<span class="active-command-name">{activeCommand.display_name}</span>
				<button onclick={clearActiveCommand} class="active-command-clear" aria-label="Clear command">
					<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="12" height="12">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
					</svg>
				</button>
			</div>
			<span class="active-command-desc">{activeCommand.description}</span>
		</div>
	{/if}

	<!-- Textarea -->
	<textarea
		bind:this={inputRef}
		bind:value={inputValue}
		class="focus-input"
		placeholder={activeCommand
			? `Describe your ${activeCommand.display_name.toLowerCase()} request...`
			: placeholderText}
		rows="1"
		oninput={handleInput}
		onkeydown={handleKeyDown}
	></textarea>

	<!-- Bottom row -->
	<div class="input-row">
		<div class="input-row-left">
			<!-- Attach button -->
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
									class="context-item"
									class:selected={isSelected}
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

		<!-- Submit button -->
		<button
			class="submit-btn"
			onclick={onSubmit}
			disabled={!canSubmit}
			title={!selectedProjectId ? 'Select a project first' : ''}
		>
			<span>{!selectedProjectId && (inputValue.trim() || attachedFiles.length > 0) ? 'Select project' : "Let's go"}</span>
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="16" height="16">
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

<style>
	.hidden {
		display: none;
	}

	.input-container {
		position: relative;
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

	.input-container.dragging {
		border-color: var(--color-primary, #3b82f6);
		border-style: dashed;
		background: rgba(59, 130, 246, 0.05);
	}

	:global(.dark) .input-container.dragging {
		border-color: #0A84FF;
		background: rgba(10, 132, 255, 0.1);
	}

	/* Mode badge */
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

	/* Command dropdown */
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

	/* Active command chip */
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

	/* Textarea */
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

	/* Bottom row */
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

	/* Submit button */
	.submit-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		align-self: flex-end;
		padding: 10px 20px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: 24px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.submit-btn:hover:not(:disabled) {
		background: var(--color-primary-hover);
		transform: translateY(-1px);
	}

	.submit-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	:global(.dark) .submit-btn {
		background: #0A84FF;
	}

	:global(.dark) .submit-btn:hover:not(:disabled) {
		background: #0070E0;
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
</style>
