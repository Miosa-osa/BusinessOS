<script lang="ts">
	/**
	 * Database View Tabs
	 * Tabs for switching between different database views
	 */
	import { Plus, LayoutGrid, Columns, Calendar, GalleryVertical, List, MoreHorizontal, Trash2, Copy } from 'lucide-svelte';
	import { Menu, MenuItem, MenuSeparator, Tooltip } from '$lib/ui';
	import type { DatabaseView, DatabaseViewType } from '../../entities/block';

	interface Props {
		views: DatabaseView[];
		activeViewId: string | null;
		onSelectView?: (viewId: string) => void;
		onAddView?: (type: DatabaseViewType) => void;
		onRenameView?: (viewId: string, name: string) => void;
		onDuplicateView?: (viewId: string) => void;
		onDeleteView?: (viewId: string) => void;
	}

	let {
		views,
		activeViewId,
		onSelectView,
		onAddView,
		onRenameView,
		onDuplicateView,
		onDeleteView
	}: Props = $props();

	let showAddMenu = $state(false);
	let editingViewId = $state<string | null>(null);
	let editValue = $state('');

	const viewIcons: Record<DatabaseViewType, typeof LayoutGrid> = {
		'table': LayoutGrid,
		'kanban': Columns,
		'calendar': Calendar,
		'gallery': GalleryVertical,
		'list': List
	};

	const viewLabels: Record<DatabaseViewType, string> = {
		'table': 'Table',
		'kanban': 'Kanban',
		'calendar': 'Calendar',
		'gallery': 'Gallery',
		'list': 'List'
	};

	function handleStartEdit(view: DatabaseView) {
		editingViewId = view.id;
		editValue = view.name;
	}

	function handleEditKeydown(e: KeyboardEvent, viewId: string) {
		if (e.key === 'Enter') {
			onRenameView?.(viewId, editValue);
			editingViewId = null;
		} else if (e.key === 'Escape') {
			editingViewId = null;
		}
	}

	function handleEditBlur(viewId: string) {
		if (editValue) {
			onRenameView?.(viewId, editValue);
		}
		editingViewId = null;
	}
</script>

<div class="bos-view-tabs">
	<div class="bos-view-tabs__list">
		{#each views as view (view.id)}
			{@const isActive = view.id === activeViewId}
			{@const isEditing = editingViewId === view.id}
			{@const Icon = viewIcons[view.type]}

			<div
				class="bos-view-tabs__tab"
				class:bos-view-tabs__tab--active={isActive}
				onclick={() => onSelectView?.(view.id)}
				onkeydown={(e) => e.key === 'Enter' && onSelectView?.(view.id)}
				role="tab"
				tabindex={0}
				aria-selected={isActive}
			>
				<svelte:component this={Icon} class="bos-view-tabs__icon" />

				{#if isEditing}
					<input
						type="text"
						class="bos-view-tabs__input"
						bind:value={editValue}
						onkeydown={(e) => handleEditKeydown(e, view.id)}
						onblur={() => handleEditBlur(view.id)}
						autofocus
						onclick={(e) => e.stopPropagation()}
					/>
				{:else}
					<span class="bos-view-tabs__name" ondblclick={() => handleStartEdit(view)}>
						{view.name}
					</span>
				{/if}

				{#if isActive && views.length > 1}
					<Menu>
						{#snippet trigger()}
							<button
								class="bos-view-tabs__menu"
								onclick={(e) => e.stopPropagation()}
							>
								<MoreHorizontal />
							</button>
						{/snippet}

						<MenuItem onSelect={() => handleStartEdit(view)}>
							Rename
						</MenuItem>
						<MenuItem onSelect={() => onDuplicateView?.(view.id)}>
							{#snippet prefix()}<Copy />{/snippet}
							Duplicate
						</MenuItem>
						{#if views.length > 1}
							<MenuSeparator />
							<MenuItem destructive onSelect={() => onDeleteView?.(view.id)}>
								{#snippet prefix()}<Trash2 />{/snippet}
								Delete
							</MenuItem>
						{/if}
					</Menu>
				{/if}
			</div>
		{/each}
	</div>

	<Menu bind:open={showAddMenu}>
		{#snippet trigger()}
			<Tooltip content="Add view" side="top">
				<button class="bos-view-tabs__add">
					<Plus />
				</button>
			</Tooltip>
		{/snippet}

		{#each Object.entries(viewLabels) as [type, label]}
			{@const Icon = viewIcons[type as DatabaseViewType]}
			<MenuItem onSelect={() => onAddView?.(type as DatabaseViewType)}>
				{#snippet prefix()}<svelte:component this={Icon} />{/snippet}
				{label}
			</MenuItem>
		{/each}
	</Menu>
</div>

<style>
	.bos-view-tabs {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 8px 12px;
		border-bottom: 1px solid var(--bos-v2-layer-insideBorder-border, rgba(0, 0, 0, 0.1));
		background: var(--bos-v2-layer-background-primary, #ffffff);
	}

	.bos-view-tabs__list {
		display: flex;
		align-items: center;
		gap: 2px;
	}

	.bos-view-tabs__tab {
		display: flex;
		align-items: center;
		gap: 6px;
		height: 28px;
		padding: 0 10px;
		border-radius: 6px;
		font-size: var(--bos-font-sm, 14px);
		color: var(--bos-v2-text-secondary, #8e8d91);
		background: transparent;
		cursor: pointer;
		transition: all 0.15s;
	}

	.bos-view-tabs__tab:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.bos-view-tabs__tab--active {
		background: var(--bos-v2-layer-background-secondary, #f4f4f5);
		color: var(--bos-v2-text-primary, #121212);
		font-weight: 500;
	}

	.bos-view-tabs__tab :global(.bos-view-tabs__icon) {
		width: 14px;
		height: 14px;
		flex-shrink: 0;
	}

	.bos-view-tabs__name {
		white-space: nowrap;
	}

	.bos-view-tabs__input {
		width: 100px;
		height: 22px;
		padding: 0 6px;
		border: 1px solid var(--bos-brand-color, #1e96eb);
		border-radius: 4px;
		font-size: var(--bos-font-sm, 14px);
		background: var(--bos-v2-layer-background-primary, #ffffff);
		color: var(--bos-v2-text-primary, #121212);
		outline: none;
	}

	.bos-view-tabs__menu {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 18px;
		height: 18px;
		padding: 0;
		border: none;
		background: transparent;
		border-radius: 4px;
		color: var(--bos-v2-icon-secondary, #a9a9ad);
		cursor: pointer;
		opacity: 0;
		transition: opacity 0.15s;
	}

	.bos-view-tabs__tab:hover .bos-view-tabs__menu,
	.bos-view-tabs__tab--active .bos-view-tabs__menu {
		opacity: 1;
	}

	.bos-view-tabs__menu:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
	}

	.bos-view-tabs__menu :global(svg) {
		width: 14px;
		height: 14px;
	}

	.bos-view-tabs__add {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		border: none;
		background: transparent;
		border-radius: 4px;
		color: var(--bos-v2-icon-secondary, #a9a9ad);
		cursor: pointer;
		transition: all 0.15s;
	}

	.bos-view-tabs__add:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(0, 0, 0, 0.04));
		color: var(--bos-v2-icon-primary, #77757d);
	}

	.bos-view-tabs__add :global(svg) {
		width: 16px;
		height: 16px;
	}

	/* Dark mode */
	:global(.dark) .bos-view-tabs {
		background: var(--bos-v2-layer-background-primary, #1e1e1e);
		border-color: var(--bos-v2-layer-insideBorder-border, rgba(255, 255, 255, 0.1));
	}

	:global(.dark) .bos-view-tabs__tab:hover {
		background: var(--bos-v2-layer-background-hoverOverlay, rgba(255, 255, 255, 0.08));
	}

	:global(.dark) .bos-view-tabs__tab--active {
		background: var(--bos-v2-layer-background-secondary, #2c2c2c);
		color: var(--bos-v2-text-primary, #e6e6e6);
	}

	:global(.dark) .bos-view-tabs__input {
		background: var(--bos-v2-layer-background-primary, #1e1e1e);
		color: var(--bos-v2-text-primary, #e6e6e6);
	}
</style>
