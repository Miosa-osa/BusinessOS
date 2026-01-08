<script lang="ts">
	/**
	 * TableHeader - Table name, views tabs, and actions
	 */
	import {
		ChevronDown,
		Plus,
		Star,
		Settings,
		Table2,
		Columns3,
		LayoutGrid,
		Calendar,
		FileInput
	} from 'lucide-svelte';
	import type { Table, TableView, ViewType } from '$lib/api/tables/types';
	import { VIEW_TYPES } from '$lib/api/tables/types';

	interface Props {
		table: Table;
		currentView: TableView | null;
		onViewChange: (viewId: string) => void;
		onCreateView: (type: ViewType) => void;
		onFavoriteToggle: () => void;
		onSettingsClick?: () => void;
	}

	let { table, currentView, onViewChange, onCreateView, onFavoriteToggle, onSettingsClick }: Props =
		$props();

	let showViewMenu = $state(false);

	function getViewIcon(type: ViewType) {
		switch (type) {
			case 'grid':
				return Table2;
			case 'kanban':
				return Columns3;
			case 'gallery':
				return LayoutGrid;
			case 'calendar':
				return Calendar;
			case 'form':
				return FileInput;
			default:
				return Table2;
		}
	}

	function handleClickOutside() {
		showViewMenu = false;
	}
</script>

<svelte:window onclick={handleClickOutside} />

<div class="flex items-center justify-between border-b border-gray-200 bg-white px-6 py-3">
	<!-- Left: Table name and views -->
	<div class="flex items-center gap-4">
		<!-- Table Icon & Name -->
		<div class="flex items-center gap-2">
			{#if table.icon}
				<span class="text-2xl">{table.icon}</span>
			{:else}
				<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-100">
					<Table2 class="h-4 w-4 text-blue-600" />
				</div>
			{/if}
			<h1 class="text-xl font-semibold text-gray-900">{table.name}</h1>
			<button
				type="button"
				class="rounded p-1 text-gray-300 transition-colors hover:text-yellow-400 {table.is_favorite
					? 'text-yellow-400'
					: ''}"
				onclick={onFavoriteToggle}
			>
				<Star class="h-5 w-5 {table.is_favorite ? 'fill-current' : ''}" />
			</button>
		</div>

		<!-- View Tabs -->
		<div class="flex items-center gap-1 border-l border-gray-200 pl-4">
			{#each table.views as view}
				<button
					type="button"
					class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {currentView?.id ===
					view.id
						? 'bg-blue-100 text-blue-700'
						: 'text-gray-600 hover:bg-gray-100'}"
					onclick={() => onViewChange(view.id)}
				>
					<svelte:component this={getViewIcon(view.type)} class="h-4 w-4" />
					{view.name}
				</button>
			{/each}

			<!-- Add View -->
			<div class="relative">
				<button
					type="button"
					class="flex items-center gap-1 rounded-lg px-2 py-1.5 text-sm text-gray-500 hover:bg-gray-100"
					onclick={(e) => {
						e.stopPropagation();
						showViewMenu = !showViewMenu;
					}}
				>
					<Plus class="h-4 w-4" />
				</button>

				{#if showViewMenu}
					<div
						class="absolute left-0 top-full z-10 mt-1 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
					>
						<div class="px-3 py-2 text-xs font-medium uppercase text-gray-400">Add View</div>
						{#each VIEW_TYPES as viewType}
							<button
								type="button"
								class="flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
								onclick={() => {
									onCreateView(viewType.type);
									showViewMenu = false;
								}}
							>
								<svelte:component this={getViewIcon(viewType.type)} class="h-4 w-4 text-gray-400" />
								{viewType.label}
							</button>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Right: Actions -->
	<div class="flex items-center gap-2">
		{#if onSettingsClick}
			<button
				type="button"
				class="rounded-lg p-2 text-gray-500 hover:bg-gray-100"
				onclick={onSettingsClick}
			>
				<Settings class="h-5 w-5" />
			</button>
		{/if}
	</div>
</div>
