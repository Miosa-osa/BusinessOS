<script lang="ts">
	interface ContextItem {
		id: string;
		name: string;
		icon?: string;
	}

	interface Props {
		canSubmit: boolean;
		selectedProjectId?: string | null;
		inputValue: string;
		attachedFilesCount: number;
		availableContexts?: ContextItem[];
		selectedContextIds?: string[];
		onAttach: () => void;
		onContextToggle?: (contextId: string) => void;
		onSubmit: () => void;
	}

	let {
		canSubmit,
		selectedProjectId = null,
		inputValue,
		attachedFilesCount,
		availableContexts = [],
		selectedContextIds = [],
		onAttach,
		onContextToggle,
		onSubmit
	}: Props = $props();

	let showContextDropdown = $state(false);
</script>

<div class="input-row">
	<div class="input-row-left">
		<!-- Attach button -->
		<button
			class="btn-pill btn-pill-ghost attach-btn"
			onclick={onAttach}
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
					class="btn-pill btn-pill-ghost context-btn"
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
								class="btn-pill btn-pill-ghost context-clear"
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
								class="btn-pill btn-pill-ghost context-item"
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
		class="btn-pill btn-pill-primary submit-btn"
		onclick={onSubmit}
		disabled={!canSubmit}
		title={!selectedProjectId ? 'Select a project first' : ''}
	>
		<span>{!selectedProjectId && (inputValue.trim() || attachedFilesCount > 0) ? 'Select project' : "Let's go"}</span>
		<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" width="16" height="16">
			<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
		</svg>
	</button>
</div>

<style>
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
</style>
