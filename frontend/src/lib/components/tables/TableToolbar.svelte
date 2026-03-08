<script lang="ts">
	/**
	 * TableToolbar - Filter, sort, hide columns, search
	 */
	import {
		Search,
		Filter,
		ArrowUpDown,
		EyeOff,
		Plus,
		Download,
		Upload,
		MoreHorizontal
	} from 'lucide-svelte';
	import type { Column, Filter as FilterType, Sort } from '$lib/api/tables/types';

	interface Props {
		columns: Column[];
		filters: FilterType[];
		sorts: Sort[];
		searchQuery: string;
		selectedCount: number;
		onSearchChange: (query: string) => void;
		onAddFilter?: () => void;
		onAddSort?: () => void;
		onHideFields?: () => void;
		onAddRow: () => void;
		onDeleteSelected?: () => void;
		onExport?: () => void;
		onImport?: () => void;
	}

	let {
		columns,
		filters,
		sorts,
		searchQuery,
		selectedCount,
		onSearchChange,
		onAddFilter,
		onAddSort,
		onHideFields,
		onAddRow,
		onDeleteSelected,
		onExport,
		onImport
	}: Props = $props();

	let showMoreMenu = $state(false);

	function handleClickOutside() {
		showMoreMenu = false;
	}
</script>

<svelte:window onclick={handleClickOutside} />

<div class="flex items-center justify-between border-b border-gray-100 bg-gray-50 px-4 py-2">
	<!-- Left: Search and Filters -->
	<div class="flex items-center gap-2">
		<!-- Search -->
		<div class="relative">
			<Search class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
			<input
				type="text"
				placeholder="Search..."
				value={searchQuery}
				oninput={(e) => onSearchChange((e.target as HTMLInputElement).value)}
				class="w-48 rounded-lg border border-gray-200 bg-white py-1.5 pl-8 pr-3 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
			/>
		</div>

		<!-- Filter Button -->
		{#if onAddFilter}
			<button
				type="button"
				class="btn-pill btn-pill-sm flex items-center gap-1.5 {filters.length > 0 ? 'btn-pill-soft' : 'btn-pill-secondary'}"
				onclick={onAddFilter}
			>
				<Filter class="h-4 w-4" />
				Filter
				{#if filters.length > 0}
					<span class="rounded bg-blue-100 px-1.5 py-0.5 text-xs font-medium text-blue-600">
						{filters.length}
					</span>
				{/if}
			</button>
		{/if}

		<!-- Sort Button -->
		{#if onAddSort}
			<button
				type="button"
				class="btn-pill btn-pill-sm flex items-center gap-1.5 {sorts.length > 0 ? 'btn-pill-soft' : 'btn-pill-secondary'}"
				onclick={onAddSort}
			>
				<ArrowUpDown class="h-4 w-4" />
				Sort
				{#if sorts.length > 0}
					<span class="rounded bg-blue-100 px-1.5 py-0.5 text-xs font-medium text-blue-600">
						{sorts.length}
					</span>
				{/if}
			</button>
		{/if}

		<!-- Hide Fields Button -->
		{#if onHideFields}
			<button
				type="button"
				class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-1.5"
				onclick={onHideFields}
			>
				<EyeOff class="h-4 w-4" />
				Hide fields
			</button>
		{/if}
	</div>

	<!-- Right: Actions -->
	<div class="flex items-center gap-2">
		<!-- Selection Actions -->
		{#if selectedCount > 0}
			<div class="flex items-center gap-2 border-r border-gray-200 pr-2">
				<span class="text-sm text-gray-600">{selectedCount} selected</span>
				{#if onDeleteSelected}
					<button
						type="button"
						class="btn-pill btn-pill-ghost btn-pill-xs"
						onclick={onDeleteSelected}
					>
						Delete
					</button>
				{/if}
			</div>
		{/if}

		<!-- Add Row -->
		<button
			type="button"
			class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-1.5"
			onclick={onAddRow}
		>
			<Plus class="h-4 w-4" />
			Add row
		</button>

		<!-- More Menu -->
		<div class="relative">
			<button
				type="button"
				class="btn-pill btn-pill-ghost btn-pill-icon"
				onclick={(e) => {
					e.stopPropagation();
					showMoreMenu = !showMoreMenu;
				}}
			>
				<MoreHorizontal class="h-5 w-5" />
			</button>

			{#if showMoreMenu}
				<div
					class="absolute right-0 top-full z-10 mt-1 w-40 rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
				>
					{#if onExport}
						<button
							type="button"
							class="flex w-full items-center gap-2 px-3 py-2 rounded-lg text-sm hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
							onclick={() => {
								onExport();
								showMoreMenu = false;
							}}
						>
							<Download class="h-4 w-4" />
							Export
						</button>
					{/if}
					{#if onImport}
						<button
							type="button"
							class="flex w-full items-center gap-2 px-3 py-2 rounded-lg text-sm hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
							onclick={() => {
								onImport();
								showMoreMenu = false;
							}}
						>
							<Upload class="h-4 w-4" />
							Import
						</button>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</div>
