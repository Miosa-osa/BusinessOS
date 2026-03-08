<script lang="ts">
	/**
	 * Sidebar Header - BusinessOS Style
	 * Modern document-centric workspace selector and quick-search patterns
	 */
	import { Search, Plus, Settings } from 'lucide-svelte';
	import { Tooltip } from '$lib/ui';

	interface Props {
		onNewDocument?: () => void;
		onOpenSearch?: () => void;
		onOpenSettings?: () => void;
	}

	let { onNewDocument, onOpenSearch, onOpenSettings }: Props = $props();
</script>

<div class="bos-sidebar-header">
	<div class="bos-sidebar-header__workspace">
		<span class="bos-sidebar-header__logo">P</span>
		<span class="bos-sidebar-header__name">Pages</span>
	</div>

	<div class="bos-sidebar-header__actions">
		<!-- Quick Search Button -->
		<button
			class="bos-sidebar-header__search"
			onclick={onOpenSearch}
			aria-label="Search"
		>
			<Search />
			<span>Search</span>
			<kbd>⌘K</kbd>
		</button>

		<!-- Action buttons -->
		<div class="bos-sidebar-header__buttons">
			<Tooltip content="Settings" side="bottom">
				<button
					class="bos-sidebar-header__action"
					onclick={onOpenSettings}
					aria-label="Settings"
				>
					<Settings />
				</button>
			</Tooltip>
			<Tooltip content="New page" side="bottom">
				<button
					class="bos-sidebar-header__action bos-sidebar-header__action--primary"
					onclick={onNewDocument}
					aria-label="New page"
				>
					<Plus />
				</button>
			</Tooltip>
		</div>
	</div>
</div>

<style>
	.bos-sidebar-header {
		display: flex;
		flex-direction: column;
		padding: 12px;
		gap: 12px;
	}

	.bos-sidebar-header__workspace {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 4px;
	}

	.bos-sidebar-header__logo {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		border-radius: 4px;
		background: linear-gradient(135deg, var(--bos-brand-color, #1e96eb) 0%, #1976d2 100%);
		color: #ffffff;
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.5px;
	}

	.bos-sidebar-header__name {
		font-size: var(--bos-font-sm, 14px);
		font-weight: 600;
		color: var(--bos-v2-text-primary, #121212);
	}

	.bos-sidebar-header__actions {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	/* Search button - quick-search-input style */
	.bos-sidebar-header__search {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		height: 32px;
		padding: 0 12px;
		border-radius: 8px;
		background-color: var(--bos-v2-layer-background-primary, #ffffff);
		border: 1px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		color: var(--bos-v2-text-secondary, #8e8d91);
		font-size: var(--bos-font-sm, 14px);
		cursor: pointer;
		transition: border-color 0.15s;
	}

	.bos-sidebar-header__search:hover {
		border-color: var(--bos-v2-layer-insideBorder-primaryBorder, #1e96eb);
	}

	.bos-sidebar-header__search :global(svg) {
		width: 14px;
		height: 14px;
		flex-shrink: 0;
	}

	.bos-sidebar-header__search span {
		flex: 1;
		text-align: left;
	}

	.bos-sidebar-header__search kbd {
		display: inline-flex;
		align-items: center;
		padding: 2px 6px;
		border-radius: 4px;
		background-color: var(--bos-v2-layer-background-secondary, #f4f4f5);
		font-family: var(--bos-font-code-family);
		font-size: 10px;
		color: var(--bos-v2-text-tertiary, #bfbfc3);
	}

	/* Action buttons */
	.bos-sidebar-header__buttons {
		display: flex;
		gap: 4px;
	}

	.bos-sidebar-header__action {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: 4px;
		background: transparent;
		border: none;
		color: var(--bos-v2-icon-primary, #77757d);
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.bos-sidebar-header__action:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.bos-sidebar-header__action :global(svg) {
		width: 16px;
		height: 16px;
	}

	/* Primary action (New page) - subtle emphasis */
	.bos-sidebar-header__action--primary {
		color: var(--bos-brand-color, #1e96eb);
	}

	.bos-sidebar-header__action--primary:hover {
		background: var(--bos-brand-color, #1e96eb);
		color: #ffffff;
	}

	/* Dark mode */
	:global(.dark) .bos-sidebar-header__name {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .bos-sidebar-header__search {
		background-color: var(--bos-v2-layer-background-primary, #1e1e1e);
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
	}

	:global(.dark) .bos-sidebar-header__search kbd {
		background-color: var(--bos-v2-layer-background-tertiary, #3a3a3a);
		color: var(--bos-v2-text-tertiary, #545459);
	}

	:global(.dark) .bos-sidebar-header__action {
		color: var(--bos-v2-icon-primary, #a6a6ad);
	}

	:global(.dark) .bos-sidebar-header__action:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.08));
	}

	:global(.dark) .bos-sidebar-header__action--primary {
		color: var(--bos-brand-color, #4da6f0);
	}

	:global(.dark) .bos-sidebar-header__action--primary:hover {
		background: var(--bos-brand-color, #4da6f0);
		color: #ffffff;
	}
</style>
