<script lang="ts">
	/**
	 * Tables List Page
	 * Main page showing all tables with list/card views
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Table2, Plus, Search, Database, Upload, Loader2 } from 'lucide-svelte';
	import { tables, filteredTables, favoriteTables, type TableViewMode } from '$lib/stores/tables';
	import type { TableListItem, CreateTableData, TableSource } from '$lib/api/tables/types';
	import {
		AddTableModal,
		TableListView,
		TableCardView,
		TableViewSwitcher
	} from '$lib/components/tables';

	// Embed mode support
	const embedSuffix = $derived(
		$page.url.searchParams.get('embed') === 'true' ? '?embed=true' : ''
	);

	// State
	let showAddModal = $state(false);
	let tablesList = $state<TableListItem[]>([]);
	let favorites = $state<TableListItem[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let viewMode = $state<TableViewMode>('list');
	let searchQuery = $state('');
	let sourceFilter = $state<TableSource | null>(null);

	// Subscribe to stores
	$effect(() => {
		const unsubscribe = tables.subscribe((state) => {
			loading = state.loading;
			error = state.error;
			viewMode = state.viewMode;
			searchQuery = state.filters.search;
			sourceFilter = state.filters.source;
		});
		return unsubscribe;
	});

	$effect(() => {
		const unsubscribe = filteredTables.subscribe((items) => {
			tablesList = items;
		});
		return unsubscribe;
	});

	$effect(() => {
		const unsubscribe = favoriteTables.subscribe((items) => {
			favorites = items;
		});
		return unsubscribe;
	});

	// Load tables on mount
	onMount(() => {
		tables.loadTables();
	});

	// Event handlers
	function handleTableClick(id: string) {
		goto(`/tables/${id}${embedSuffix}`);
	}

	function handleFavoriteToggle(id: string) {
		tables.toggleFavorite(id);
	}

	async function handleDelete(id: string) {
		if (confirm('Are you sure you want to delete this table? This action cannot be undone.')) {
			try {
				await tables.deleteTable(id);
			} catch (err) {
				console.error('Failed to delete table:', err);
			}
		}
	}

	async function handleCreateTable(data: CreateTableData) {
		try {
			const table = await tables.createTable(data);
			showAddModal = false;
			goto(`/tables/${table.id}${embedSuffix}`);
		} catch (err) {
			console.error('Failed to create table:', err);
		}
	}

	function handleViewChange(mode: TableViewMode) {
		tables.setViewMode(mode);
	}

	function handleSearchChange(e: Event) {
		const target = e.target as HTMLInputElement;
		tables.setFilters({ search: target.value });
	}

	function handleSourceFilter(source: TableSource | null) {
		tables.setFilters({ source });
		tables.loadTables();
	}

	const sourceFilters: { value: TableSource | null; label: string; icon: typeof Table2 }[] = [
		{ value: null, label: 'All', icon: Table2 },
		{ value: 'custom', label: 'Custom', icon: Table2 },
		{ value: 'import', label: 'Imported', icon: Upload },
		{ value: 'integration', label: 'Integrations', icon: Database }
	];
</script>

<svelte:head>
	<title>Tables | BusinessOS</title>
</svelte:head>

<div class="flex h-full flex-col bg-gray-50">
	<!-- Header -->
	<div class="border-b border-gray-200 bg-white px-6 py-4">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">Tables</h1>
				<p class="mt-1 text-sm text-gray-500">
					Manage your data tables, imports, and integrations
				</p>
			</div>
			<button
				type="button"
				class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
				onclick={() => (showAddModal = true)}
			>
				<Plus class="h-4 w-4" />
				New Table
			</button>
		</div>
	</div>

	<!-- Filters & Search -->
	<div class="border-b border-gray-200 bg-white px-6 py-3">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-4">
				<!-- Source Tabs -->
				<div class="flex items-center gap-1">
					{#each sourceFilters as filter}
						<button
							type="button"
							class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {sourceFilter ===
							filter.value
								? 'bg-gray-100 text-gray-900'
								: 'text-gray-500 hover:text-gray-700'}"
							onclick={() => handleSourceFilter(filter.value)}
						>
							<filter.icon class="h-4 w-4" />
							{filter.label}
						</button>
					{/each}
				</div>

				<!-- Search -->
				<div class="relative">
					<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
					<input
						type="text"
						placeholder="Search tables..."
						value={searchQuery}
						oninput={handleSearchChange}
						class="w-64 rounded-lg border border-gray-200 py-1.5 pl-9 pr-3 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					/>
				</div>
			</div>

			<!-- View Switcher -->
			<TableViewSwitcher {viewMode} onChange={handleViewChange} />
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-auto p-6">
		{#if loading}
			<div class="flex h-64 flex-col items-center justify-center">
				<Loader2 class="mb-4 h-8 w-8 animate-spin text-blue-600" />
				<p class="text-sm text-gray-500">Loading tables...</p>
			</div>
		{:else if error}
			<div class="rounded-lg border border-red-200 bg-red-50 p-4">
				<p class="text-sm text-red-600">{error}</p>
				<button
					type="button"
					class="mt-2 text-sm font-medium text-red-600 hover:text-red-700"
					onclick={() => tables.loadTables()}
				>
					Try again
				</button>
			</div>
		{:else}
			<!-- Favorites Section -->
			{#if favorites.length > 0 && !searchQuery}
				<div class="mb-8">
					<h2 class="mb-3 text-sm font-medium text-gray-500">Favorites</h2>
					{#if viewMode === 'list'}
						<TableListView
							tables={favorites}
							onTableClick={handleTableClick}
							onFavoriteToggle={handleFavoriteToggle}
							onDelete={handleDelete}
						/>
					{:else}
						<TableCardView
							tables={favorites}
							onTableClick={handleTableClick}
							onFavoriteToggle={handleFavoriteToggle}
						/>
					{/if}
				</div>
			{/if}

			<!-- All Tables Section -->
			<div>
				{#if favorites.length > 0 && !searchQuery}
					<h2 class="mb-3 text-sm font-medium text-gray-500">All Tables</h2>
				{/if}

				{#if tablesList.length === 0}
					<!-- Empty State -->
					<div class="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-200 py-16">
						<div class="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100">
							<Table2 class="h-8 w-8 text-gray-400" />
						</div>
						<h3 class="mb-1 text-lg font-medium text-gray-900">No tables yet</h3>
						<p class="mb-4 text-sm text-gray-500">
							{#if searchQuery}
								No tables match your search
							{:else}
								Create your first table to get started
							{/if}
						</p>
						{#if !searchQuery}
							<button
								type="button"
								class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
								onclick={() => (showAddModal = true)}
							>
								<Plus class="h-4 w-4" />
								Create Table
							</button>
						{/if}
					</div>
				{:else if viewMode === 'list'}
					<TableListView
						tables={tablesList}
						onTableClick={handleTableClick}
						onFavoriteToggle={handleFavoriteToggle}
						onDelete={handleDelete}
					/>
				{:else}
					<TableCardView
						tables={tablesList}
						onTableClick={handleTableClick}
						onFavoriteToggle={handleFavoriteToggle}
					/>
				{/if}
			</div>
		{/if}
	</div>
</div>

<!-- Create Table Modal -->
<AddTableModal
	open={showAddModal}
	onClose={() => (showAddModal = false)}
	onCreate={handleCreateTable}
/>
