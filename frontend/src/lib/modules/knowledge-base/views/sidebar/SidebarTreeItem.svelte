<script lang="ts">
	/**
	 * Sidebar Tree Item - BusinessOS Style
	 * Modern document-centric menu-item patterns (30px height, hover overlay)
	 */
	import { type Snippet } from 'svelte';
	import { ChevronRight, FileText, Folder, MoreHorizontal, Plus, Star, Copy, Link2, Pencil, FolderInput, Trash2, Archive, ExternalLink } from 'lucide-svelte';
	import { Menu, MenuItem, MenuSeparator, Tooltip } from '$lib/ui';
	import type { DocumentMeta, DocumentIcon } from '../../entities/types';
	import { treeStore } from '../../stores/documents';
	import { iconLibrary } from '../editor/PageIconPicker.svelte';

	interface Props {
		document: DocumentMeta;
		depth?: number;
		hasChildren?: boolean;
		isExpanded?: boolean;
		isLoading?: boolean;
		isActive?: boolean;
		onSelect?: () => void;
		onToggleExpand?: () => void;
		onAddChild?: () => void;
		onDelete?: () => void;
		onDuplicate?: () => void;
		onToggleFavorite?: () => void;
		onRename?: () => void;
		onCopyLink?: () => void;
		onMoveTo?: () => void;
		onArchive?: () => void;
		onOpenInNewTab?: () => void;
		children?: Snippet;
	}

	let {
		document,
		depth = 0,
		hasChildren = false,
		isExpanded = false,
		isLoading = false,
		isActive = false,
		onSelect,
		onToggleExpand,
		onAddChild,
		onDelete,
		onDuplicate,
		onToggleFavorite,
		onRename,
		onCopyLink,
		onMoveTo,
		onArchive,
		onOpenInNewTab,
		children
	}: Props = $props();

	let showMenu = $state(false);
	let isHovered = $state(false);
	let contextMenuPosition = $state({ x: 0, y: 0 });
	let showContextMenu = $state(false);

	const paddingLeft = $derived(12 + depth * 16);

	// Right-click handler for context menu
	function handleContextMenu(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		contextMenuPosition = { x: e.clientX, y: e.clientY };
		showContextMenu = true;
	}

	// Close context menu when clicking outside
	function handleClickOutside(e: MouseEvent) {
		if (showContextMenu) {
			showContextMenu = false;
		}
	}

	// Copy link to clipboard
	function handleCopyLink() {
		const url = `${window.location.origin}/pages/${document.id}`;
		navigator.clipboard.writeText(url);
		showContextMenu = false;
		showMenu = false;
		onCopyLink?.();
	}

	// Open in new tab
	function handleOpenInNewTab() {
		window.open(`/pages/${document.id}`, '_blank');
		showContextMenu = false;
		showMenu = false;
		onOpenInNewTab?.();
	}

	function handleChevronClick(e: MouseEvent) {
		e.stopPropagation();
		treeStore.toggleExpanded(document.id);
		onToggleExpand?.();
	}

	function handleSelect() {
		onSelect?.();
	}

	// Get icon name from DocumentIcon - handles string and object formats
	function getIconName(icon: DocumentIcon | null): string | null {
		if (!icon) return null;
		if (typeof icon === 'string') return icon;
		if (icon.type === 'icon') return icon.value;
		return null;
	}

	// Get SVG path for icon
	function getIconPath(iconName: string | null): string | null {
		if (!iconName) return null;
		return iconLibrary[iconName] || null;
	}

	const iconName = $derived(getIconName(document.icon));
	const iconPath = $derived(getIconPath(iconName));
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="tree-item-container" onclick={handleClickOutside}>
	<!-- Main row - using div with role="button" to avoid nested button issues -->
	<div
		class="tree-item"
		class:tree-item--active={isActive}
		class:tree-item--hovered={isHovered}
		style:padding-left="{paddingLeft}px"
		onclick={handleSelect}
		oncontextmenu={handleContextMenu}
		onkeydown={(e) => e.key === 'Enter' && handleSelect()}
		onmouseenter={() => (isHovered = true)}
		onmouseleave={() => (isHovered = false)}
		role="button"
		tabindex={0}
	>
		<!-- Expand/collapse chevron -->
		<span
			class="tree-item__chevron"
			class:tree-item__chevron--visible={hasChildren}
			class:tree-item__chevron--expanded={isExpanded}
			onclick={handleChevronClick}
			onkeydown={(e) => e.key === 'Enter' && handleChevronClick(e as unknown as MouseEvent)}
			role="button"
			tabindex={hasChildren ? 0 : -1}
		>
			{#if isLoading}
				<div class="tree-item__spinner"></div>
			{:else}
				<ChevronRight />
			{/if}
		</span>

		<!-- Icon -->
		<span class="tree-item__icon">
			{#if iconPath}
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d={iconPath} />
				</svg>
			{:else if document.type === 'folder'}
				<Folder />
			{:else}
				<FileText />
			{/if}
		</span>

		<!-- Title -->
		<span class="tree-item__title">
			{document.title || 'Untitled'}
		</span>

		<!-- Favorite indicator -->
		{#if document.is_favorite}
			<Star class="tree-item__favorite" />
		{/if}

		<!-- Actions (visible on hover) -->
		{#if isHovered}
			<div class="tree-item__actions" onclick={(e) => e.stopPropagation()}>
				<Tooltip content="Add subpage" side="top">
					<span
						class="tree-item__action"
						onclick={onAddChild}
						onkeydown={(e) => e.key === 'Enter' && onAddChild?.()}
						role="button"
						tabindex={0}
					>
						<Plus />
					</span>
				</Tooltip>

				<Menu bind:open={showMenu}>
					{#snippet trigger()}
						<span class="tree-item__action" role="button" tabindex={0}>
							<MoreHorizontal />
						</span>
					{/snippet}

					<MenuItem onSelect={onToggleFavorite}>
						{#snippet prefix()}
							<Star class="h-4 w-4" />
						{/snippet}
						{document.is_favorite ? 'Remove from favorites' : 'Add to favorites'}
					</MenuItem>
					<MenuItem onSelect={handleCopyLink}>
						{#snippet prefix()}
							<Link2 class="h-4 w-4" />
						{/snippet}
						Copy link
					</MenuItem>
					<MenuItem onSelect={onDuplicate}>
						{#snippet prefix()}
							<Copy class="h-4 w-4" />
						{/snippet}
						Duplicate
					</MenuItem>
					<MenuItem onSelect={onRename}>
						{#snippet prefix()}
							<Pencil class="h-4 w-4" />
						{/snippet}
						Rename
					</MenuItem>
					<MenuSeparator />
					<MenuItem onSelect={onMoveTo}>
						{#snippet prefix()}
							<FolderInput class="h-4 w-4" />
						{/snippet}
						Move to
					</MenuItem>
					<MenuItem onSelect={handleOpenInNewTab}>
						{#snippet prefix()}
							<ExternalLink class="h-4 w-4" />
						{/snippet}
						Open in new tab
					</MenuItem>
					<MenuSeparator />
					<MenuItem onSelect={onArchive}>
						{#snippet prefix()}
							<Archive class="h-4 w-4" />
						{/snippet}
						{document.is_archived ? 'Unarchive' : 'Archive'}
					</MenuItem>
					<MenuItem destructive onSelect={onDelete}>
						{#snippet prefix()}
							<Trash2 class="h-4 w-4" />
						{/snippet}
						Delete
					</MenuItem>
				</Menu>
			</div>
		{/if}
	</div>

	<!-- Children -->
	{#if children && isExpanded}
		<div class="tree-item__children">
			{@render children()}
		</div>
	{/if}

	<!-- Right-click context menu -->
	{#if showContextMenu}
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="context-menu-overlay"
			onclick={() => showContextMenu = false}
			onkeydown={(e) => e.key === 'Escape' && (showContextMenu = false)}
		>
			<div
				class="context-menu"
				style="left: {contextMenuPosition.x}px; top: {contextMenuPosition.y}px;"
				onclick={(e) => e.stopPropagation()}
			>
				<button class="context-menu__item" onclick={() => { onToggleFavorite?.(); showContextMenu = false; }}>
					<Star class="h-4 w-4" />
					<span>{document.is_favorite ? 'Remove from favorites' : 'Add to favorites'}</span>
				</button>
				<button class="context-menu__item" onclick={handleCopyLink}>
					<Link2 class="h-4 w-4" />
					<span>Copy link</span>
				</button>
				<button class="context-menu__item" onclick={() => { onDuplicate?.(); showContextMenu = false; }}>
					<Copy class="h-4 w-4" />
					<span>Duplicate</span>
				</button>
				<button class="context-menu__item" onclick={() => { onRename?.(); showContextMenu = false; }}>
					<Pencil class="h-4 w-4" />
					<span>Rename</span>
				</button>
				<div class="context-menu__separator"></div>
				<button class="context-menu__item" onclick={() => { onMoveTo?.(); showContextMenu = false; }}>
					<FolderInput class="h-4 w-4" />
					<span>Move to</span>
				</button>
				<button class="context-menu__item" onclick={handleOpenInNewTab}>
					<ExternalLink class="h-4 w-4" />
					<span>Open in new tab</span>
				</button>
				<div class="context-menu__separator"></div>
				<button class="context-menu__item" onclick={() => { onArchive?.(); showContextMenu = false; }}>
					<Archive class="h-4 w-4" />
					<span>{document.is_archived ? 'Unarchive' : 'Archive'}</span>
				</button>
				<button class="context-menu__item context-menu__item--destructive" onclick={() => { onDelete?.(); showContextMenu = false; }}>
					<Trash2 class="h-4 w-4" />
					<span>Delete</span>
				</button>
			</div>
		</div>
	{/if}
</div>

<style>
	/* BusinessOS-style tree item */
	.tree-item-container {
		display: flex;
		flex-direction: column;
	}

	.tree-item {
		display: inline-flex;
		align-items: center;
		width: 100%;
		min-height: 30px;
		padding: 0 8px 0 0;
		margin-top: 2px;
		background: transparent;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		text-align: left;
		font-size: var(--bos-font-sm, 14px);
		color: inherit;
		user-select: none;
		position: relative;
		outline: none;
	}

	.tree-item-container:first-child .tree-item {
		margin-top: 0;
	}

	.tree-item:hover,
	.tree-item--hovered,
	.tree-item:focus-visible {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.tree-item--active {
		background: rgba(30, 150, 235, 0.12) !important;
	}

	.tree-item--active .tree-item__title {
		font-weight: 500;
		color: var(--bos-v2-text-activated, #1e96eb);
	}

	.tree-item--active .tree-item__icon {
		color: var(--bos-v2-icon-activated, #1e96eb);
	}

	.tree-item__chevron {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 16px;
		height: 16px;
		margin-right: 2px;
		flex-shrink: 0;
		border-radius: 2px;
		color: inherit;
		cursor: pointer;
		transition: transform 0.2s;
		opacity: 0;
	}

	.tree-item__chevron--visible {
		opacity: 1;
	}

	.tree-item__chevron:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.tree-item__chevron--expanded {
		transform: rotate(90deg);
	}

	.tree-item__chevron :global(svg) {
		width: 12px;
		height: 12px;
	}

	.tree-item__spinner {
		width: 10px;
		height: 10px;
		border: 2px solid var(--bos-v2-icon-tertiary, #c9ccd0);
		border-top-color: var(--bos-v2-icon-primary, #77757d);
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.tree-item__icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 18px;
		height: 18px;
		margin-right: 6px;
		flex-shrink: 0;
		font-size: 16px;
		color: var(--bos-v2-icon-primary, #77757d);
	}

	.tree-item__icon :global(svg) {
		width: 16px;
		height: 16px;
	}

	.tree-item__title {
		flex: 1;
		font-size: var(--bos-font-sm, 14px);
		color: var(--bos-v2-text-primary, #121212);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.tree-item__favorite {
		flex-shrink: 0;
		width: 12px;
		height: 12px;
		margin-left: 4px;
		color: var(--bos-warning-color, #f59e0b);
		fill: var(--bos-warning-color, #f59e0b);
	}

	/* Postfix actions - visible on hover */
	.tree-item__actions {
		display: flex;
		gap: 2px;
		margin-left: auto;
	}

	.tree-item__action {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		border-radius: 4px;
		background: transparent;
		color: var(--bos-v2-icon-primary, #77757d);
		cursor: pointer;
		transition: background-color 0.15s;
	}

	.tree-item__action:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.tree-item__action :global(svg) {
		width: 14px;
		height: 14px;
	}

	.tree-item__children {
		display: flex;
		flex-direction: column;
	}

	/* Dark mode overrides */
	:global(.dark) .tree-item:hover,
	:global(.dark) .tree-item--hovered {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.08));
	}

	:global(.dark) .tree-item--active {
		background: rgba(30, 150, 235, 0.2) !important;
	}

	:global(.dark) .tree-item--active .tree-item__title {
		color: var(--bos-v2-text-activated, #60b5ff);
	}

	:global(.dark) .tree-item--active .tree-item__icon {
		color: var(--bos-v2-icon-activated, #60b5ff);
	}

	:global(.dark) .tree-item__title {
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .tree-item__icon {
		color: var(--bos-v2-icon-primary, #a6a6ad);
	}

	:global(.dark) .tree-item__action {
		color: var(--bos-v2-icon-primary, #a6a6ad);
	}

	:global(.dark) .tree-item__action:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.08));
	}

	/* Context menu (right-click) */
	.context-menu-overlay {
		position: fixed;
		inset: 0;
		z-index: 1000;
	}

	.context-menu {
		position: fixed;
		min-width: 200px;
		background: hsl(var(--background));
		border: 1px solid hsl(var(--border));
		border-radius: 8px;
		box-shadow: 0 4px 16px hsl(var(--foreground) / 0.12);
		padding: 4px;
		z-index: 1001;
	}

	.context-menu__item {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 8px 12px;
		background: transparent;
		border: none;
		border-radius: 4px;
		font-size: 14px;
		color: hsl(var(--foreground));
		cursor: pointer;
		text-align: left;
		transition: background-color 0.1s;
	}

	.context-menu__item:hover {
		background: hsl(var(--muted));
	}

	.context-menu__item--destructive {
		color: hsl(var(--destructive));
	}

	.context-menu__item--destructive:hover {
		background: hsl(var(--destructive) / 0.1);
	}

	.context-menu__separator {
		height: 1px;
		background: hsl(var(--border));
		margin: 4px 0;
	}
</style>
