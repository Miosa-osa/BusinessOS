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
				class="flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50 {filters.length > 0 ? 'border-blue-300 bg-blue-50 text-blue-600' : ''}"
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
				class="flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50 {sorts.length > 0 ? 'border-blue-300 bg-blue-50 text-blue-600' : ''}"
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
				class="flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50"
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
						class="rounded-lg px-2 py-1 text-sm text-red-600 hover:bg-red-50"
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
			class="btn-pill btn-pill-primary btn-pill-sm"
			onclick={onAddRow}
		>
			<Plus class="h-4 w-4" />
			Add row
		</button>

		<!-- More Menu -->
		<div class="relative">
			<button
				type="button"
				class="rounded-lg p-1.5 text-gray-500 hover:bg-gray-100"
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
							class="flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
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
							class="flex w-full items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
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
