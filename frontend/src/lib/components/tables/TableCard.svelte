<script lang="ts">
	/**
	 * TableCard - Rich table card with preview and quick actions
	 * Features: Stats preview, column badges, last modified, quick actions
	 */
	import {
		Table2,
		MoreHorizontal,
		Star,
		StarOff,
		Edit3,
		Copy,
		Trash2,
		ExternalLink,
		Grid3X3,
		LayoutGrid,
		Calendar,
		Columns3,
		Database,
		Upload,
		Link
	} from 'lucide-svelte';
	import type { TableListItem } from '$lib/api/tables/types';
	import type { ComponentType, SvelteComponent } from 'svelte';

	type IconComponent = ComponentType<SvelteComponent>;

	interface Props {
		table: TableListItem;
		isFavorite?: boolean;
		onOpen: (id: string) => void;
		onToggleFavorite: (id: string) => void;
		onRename: (id: string) => void;
		onDuplicate: (id: string) => void;
		onDelete: (id: string) => void;
	}

	let {
		table,
		isFavorite = false,
		onOpen,
		onToggleFavorite,
		onRename,
		onDuplicate,
		onDelete
	}: Props = $props();

	let showMenu = $state(false);

	// Get icon and color based on source
	function getSourceIcon(): { icon: IconComponent; color: string; bg: string; label: string } {
		switch (table.source) {
			case 'import':
				return {
					icon: Upload as unknown as IconComponent,
					color: 'text-orange-600',
					bg: 'bg-orange-100',
					label: 'Imported'
				};
			case 'integration':
				return {
					icon: Link as unknown as IconComponent,
					color: 'text-green-600',
					bg: 'bg-green-100',
					label: 'Connected'
				};
			default:
				return {
					icon: Database as unknown as IconComponent,
					color: 'text-blue-600',
					bg: 'bg-blue-100',
					label: 'Custom'
				};
		}
	}

	// Format relative time
	function getRelativeTime(dateString: string): string {
		const date = new Date(dateString);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMins / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays < 7) return `${diffDays}d ago`;
		return date.toLocaleDateString();
	}

	// Get view type icon
	function getViewIcon(viewType: string): IconComponent {
		switch (viewType) {
			case 'kanban':
				return Columns3 as unknown as IconComponent;
			case 'gallery':
				return LayoutGrid as unknown as IconComponent;
			case 'calendar':
				return Calendar as unknown as IconComponent;
			default:
				return Grid3X3 as unknown as IconComponent;
		}
	}

	const sourceInfo = $derived(getSourceIcon());

	function handleMenuClick(e: MouseEvent) {
		e.stopPropagation();
		showMenu = !showMenu;
	}

	function handleAction(action: () => void) {
		action();
		showMenu = false;
	}

	function handleClickOutside() {
		if (showMenu) showMenu = false;
	}
</script>

<svelte:window onclick={handleClickOutside} />

<div
	class="group relative flex flex-col rounded-xl border border-gray-200 bg-white shadow-sm transition-all hover:border-gray-300 hover:shadow-md"
>
	<!-- Card Header -->
	<div class="flex items-start gap-3 p-4">
		<!-- Clickable area for opening table -->
		<button
			type="button"
			class="flex flex-1 items-start gap-3 text-left"
			onclick={() => onOpen(table.id)}
		>
			<!-- Icon -->
			<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg {sourceInfo.bg}">
				<svelte:component this={sourceInfo.icon} class="h-5 w-5 {sourceInfo.color}" />
			</div>

			<!-- Title and Description -->
			<div class="min-w-0 flex-1">
				<div class="flex items-center gap-2">
					<h3 class="truncate font-medium text-gray-900 group-hover:text-blue-600">
						{table.name}
					</h3>
					{#if isFavorite}
						<Star class="h-4 w-4 shrink-0 fill-amber-400 text-amber-400" />
					{/if}
				</div>
				{#if table.description}
					<p class="mt-0.5 line-clamp-1 text-sm text-gray-500">{table.description}</p>
				{/if}
			</div>
		</button>

		<!-- Menu Button (separate from clickable area) -->
		<div class="relative shrink-0">
			<button
				type="button"
				class="rounded-lg p-1.5 text-gray-400 opacity-0 transition-opacity hover:bg-gray-100 hover:text-gray-600 group-hover:opacity-100"
				onclick={handleMenuClick}
			>
				<MoreHorizontal class="h-5 w-5" />
			</button>

			<!-- Dropdown Menu -->
			{#if showMenu}
				<div
					class="absolute right-0 top-full z-20 mt-1 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
				>
					<button
						type="button"
						class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
						onclick={() => handleAction(() => onOpen(table.id))}
					>
						<ExternalLink class="h-4 w-4" />
						Open table
					</button>
					<button
						type="button"
						class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
						onclick={() => handleAction(() => onToggleFavorite(table.id))}
					>
						{#if isFavorite}
							<StarOff class="h-4 w-4" />
							Remove from favorites
						{:else}
							<Star class="h-4 w-4" />
							Add to favorites
						{/if}
					</button>
					<div class="my-1 border-t border-gray-200"></div>
					<button
						type="button"
						class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
						onclick={() => handleAction(() => onRename(table.id))}
					>
						<Edit3 class="h-4 w-4" />
						Rename
					</button>
					<button
						type="button"
						class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
						onclick={() => handleAction(() => onDuplicate(table.id))}
					>
						<Copy class="h-4 w-4" />
						Duplicate
					</button>
					<div class="my-1 border-t border-gray-200"></div>
					<button
						type="button"
						class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
						onclick={() => handleAction(() => onDelete(table.id))}
					>
						<Trash2 class="h-4 w-4" />
						Delete
					</button>
				</div>
			{/if}
		</div>
	</div>

	<!-- Card Stats -->
	<div class="flex items-center gap-4 border-t border-gray-100 px-4 py-3">
		<!-- Row Count -->
		<div class="flex items-center gap-1.5 text-xs text-gray-500">
			<Table2 class="h-3.5 w-3.5" />
			<span>{table.row_count.toLocaleString()} rows</span>
		</div>

		<!-- Column Count -->
		<div class="flex items-center gap-1.5 text-xs text-gray-500">
			<Grid3X3 class="h-3.5 w-3.5" />
			<span>{table.columns?.length || 0} columns</span>
		</div>

		<!-- Source Badge -->
		<span
			class="ml-auto rounded-full px-2 py-0.5 text-xs font-medium {sourceInfo.bg} {sourceInfo.color}"
		>
			{sourceInfo.label}
		</span>
	</div>

	<!-- Column Preview (if available) -->
	{#if table.columns && table.columns.length > 0}
		<div class="border-t border-gray-100 px-4 py-3">
			<div class="flex flex-wrap gap-1">
				{#each table.columns.slice(0, 5) as col}
					<span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-600">
						{col.name}
					</span>
				{/each}
				{#if table.columns.length > 5}
					<span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-400">
						+{table.columns.length - 5}
					</span>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Card Footer -->
	<div class="flex items-center justify-between border-t border-gray-100 px-4 py-2">
		<span class="text-xs text-gray-400">
			Updated {getRelativeTime(table.updated_at)}
		</span>

		<!-- Quick Views (if available) -->
		{#if table.views && table.views.length > 0}
			<div class="flex items-center gap-1">
				{#each table.views.slice(0, 3) as view}
					<button
						type="button"
						class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
						title="{view.name} ({view.type})"
						onclick={(e) => {
							e.stopPropagation();
							// TODO: Open specific view
							onOpen(table.id);
						}}
					>
						<svelte:component this={getViewIcon(view.type)} class="h-3.5 w-3.5" />
					</button>
				{/each}
				{#if table.views.length > 3}
					<span class="text-xs text-gray-400">+{table.views.length - 3}</span>
				{/if}
			</div>
		{/if}
	</div>
</div>
