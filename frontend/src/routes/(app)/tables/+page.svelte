<script lang="ts">
	/**
	 * Tables List Page - NocoDB-style Layout
	 * Features: Sidebar navigation, template gallery, rich table cards, import modal
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import {
		Table2,
		Plus,
		Search,
		Database,
		Upload,
		Loader2,
		LayoutGrid,
		List,
		PanelLeftClose,
		PanelLeft,
		Sparkles
	} from 'lucide-svelte';
	import { tables, filteredTables, favoriteTables, type TableViewMode } from '$lib/stores/tables';
	import type { TableListItem, CreateTableData, TableSource } from '$lib/api/tables/types';
	import {
		AddTableModal,
		TablesSidebar,
		TableCard,
		TemplateGallery,
		ImportModal
	} from '$lib/components/tables';

	// Embed mode support
	const embedSuffix = $derived(
		$page.url.searchParams.get('embed') === 'true' ? '?embed=true' : ''
	);

	// State
	let showAddModal = $state(false);
	let showImportModal = $state(false);
	let showTemplateGallery = $state(false);
	let sidebarCollapsed = $state(false);
	let tablesList = $state<TableListItem[]>([]);
	let favorites = $state<TableListItem[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let viewMode = $state<TableViewMode>('card');
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

	function handleRename(id: string) {
		const newName = prompt('Enter new table name:');
		if (newName) {
			tables.updateTable(id, { name: newName });
		}
	}

	function handleDuplicate(id: string) {
		// TODO: Implement table duplication
		alert('Duplicate feature coming soon!');
	}

	async function handleCreateTable(data: CreateTableData) {
		try {
			const table = await tables.createTable(data);
			showAddModal = false;
			if (table?.id) {
				goto(`/tables/${table.id}${embedSuffix}`);
			}
		} catch (err) {
			console.error('Failed to create table:', err);
		}
	}

	function handleCreateBase() {
		// For now, just show add modal with a base context
		showAddModal = true;
	}

	function handleSelectTemplate(template: any) {
		// Create table from template
		const data: CreateTableData = {
			name: template.name,
			description: template.description,
			source: 'custom',
			columns: template.columns.map((col: any, i: number) => ({
				name: col.name,
				type: col.type,
				order: i,
				is_primary: i === 0
			}))
		};
		handleCreateTable(data);
		showTemplateGallery = false;
	}

	function handleImport(importData: any) {
		// Create table from import
		const data: CreateTableData = {
			name: importData.tableName,
			source: 'import',
			columns: importData.columns.map((col: any, i: number) => ({
				name: col.name,
				type: col.selectedType,
				order: i,
				is_primary: i === 0
			}))
		};
		// TODO: Also import the data rows
		handleCreateTable(data);
		showImportModal = false;
	}

	function handleViewChange(mode: TableViewMode) {
		tables.setViewMode(mode);
	}

	function handleSearchChange(query: string) {
		tables.setFilters({ search: query });
	}

	function handleSourceFilter(source: TableSource | null) {
		tables.setFilters({ source });
		tables.loadTables();
	}

	// Derived states
	const hasNoTables = $derived(tablesList.length === 0 && !loading && !searchQuery);
	const showWelcome = $derived(hasNoTables && favorites.length === 0);
</script>

<svelte:head>
	<title>Tables | BusinessOS</title>
</svelte:head>

<div class="flex h-full bg-gray-50">
	<!-- Sidebar -->
	{#if !sidebarCollapsed}
		<TablesSidebar
			tables={tablesList}
			{favorites}
			onTableClick={handleTableClick}
			onCreateTable={() => (showAddModal = true)}
			onCreateBase={handleCreateBase}
			onImport={() => (showImportModal = true)}
		/>
	{/if}

	<!-- Main Content -->
	<div class="flex flex-1 flex-col overflow-hidden">
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-gray-200 bg-white px-6 py-4">
			<div class="flex items-center gap-4">
				<!-- Sidebar Toggle -->
				<button
					type="button"
					class="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
					onclick={() => (sidebarCollapsed = !sidebarCollapsed)}
					title={sidebarCollapsed ? 'Show sidebar' : 'Hide sidebar'}
				>
					{#if sidebarCollapsed}
						<PanelLeft class="h-5 w-5" />
					{:else}
						<PanelLeftClose class="h-5 w-5" />
					{/if}
				</button>

				<div>
					<h1 class="text-xl font-semibold text-gray-900">Tables</h1>
					<p class="text-sm text-gray-500">
						{tablesList.length} table{tablesList.length !== 1 ? 's' : ''}
						{#if favorites.length > 0}
							 | {favorites.length} favorite{favorites.length !== 1 ? 's' : ''}
						{/if}
					</p>
				</div>
			</div>

			<div class="flex items-center gap-3">
				<!-- Search -->
				<div class="relative">
					<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
					<input
						type="text"
						placeholder="Search tables..."
						value={searchQuery}
						oninput={(e) => handleSearchChange((e.target as HTMLInputElement).value)}
						class="w-64 rounded-lg border border-gray-200 bg-gray-50 py-2 pl-9 pr-3 text-sm focus:border-blue-500 focus:bg-white focus:outline-none focus:ring-1 focus:ring-blue-500"
					/>
				</div>

				<!-- View Switcher -->
				<div class="flex items-center rounded-lg border border-gray-200 bg-white p-1">
					<button
						type="button"
						class="rounded p-1.5 transition-colors {viewMode === 'card'
							? 'bg-gray-100 text-gray-900'
							: 'text-gray-400 hover:text-gray-600'}"
						onclick={() => handleViewChange('card')}
						title="Card view"
					>
						<LayoutGrid class="h-4 w-4" />
					</button>
					<button
						type="button"
						class="rounded p-1.5 transition-colors {viewMode === 'list'
							? 'bg-gray-100 text-gray-900'
							: 'text-gray-400 hover:text-gray-600'}"
						onclick={() => handleViewChange('list')}
						title="List view"
					>
						<List class="h-4 w-4" />
					</button>
				</div>

				<!-- Actions -->
				<button
					type="button"
					class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
					onclick={() => (showImportModal = true)}
				>
					<Upload class="h-4 w-4" />
					Import
				</button>

				<button
					type="button"
					class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
					onclick={() => (showTemplateGallery = true)}
				>
					<Plus class="h-4 w-4" />
					New Table
				</button>
			</div>
		</div>

		<!-- Source Filter Tabs -->
		{#if tablesList.length > 0 || searchQuery}
			<div class="flex items-center gap-2 border-b border-gray-200 bg-white px-6 py-2">
				<button
					type="button"
					class="rounded-full px-3 py-1 text-sm font-medium transition-colors {sourceFilter === null
						? 'bg-blue-100 text-blue-700'
						: 'text-gray-500 hover:bg-gray-100'}"
					onclick={() => handleSourceFilter(null)}
				>
					All
				</button>
				<button
					type="button"
					class="flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium transition-colors {sourceFilter === 'custom'
						? 'bg-blue-100 text-blue-700'
						: 'text-gray-500 hover:bg-gray-100'}"
					onclick={() => handleSourceFilter('custom')}
				>
					<Database class="h-3.5 w-3.5" />
					Custom
				</button>
				<button
					type="button"
					class="flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium transition-colors {sourceFilter === 'import'
						? 'bg-orange-100 text-orange-700'
						: 'text-gray-500 hover:bg-gray-100'}"
					onclick={() => handleSourceFilter('import')}
				>
					<Upload class="h-3.5 w-3.5" />
					Imported
				</button>
				<button
					type="button"
					class="flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium transition-colors {sourceFilter === 'integration'
						? 'bg-green-100 text-green-700'
						: 'text-gray-500 hover:bg-gray-100'}"
					onclick={() => handleSourceFilter('integration')}
				>
					<Sparkles class="h-3.5 w-3.5" />
					Connected
				</button>
			</div>
		{/if}

		<!-- Content Area -->
		<div class="flex-1 overflow-auto">
			{#if loading}
				<div class="flex h-full flex-col items-center justify-center">
					<Loader2 class="mb-4 h-8 w-8 animate-spin text-blue-600" />
					<p class="text-sm text-gray-500">Loading tables...</p>
				</div>
			{:else if error}
				<div class="m-6 rounded-lg border border-red-200 bg-red-50 p-4">
					<p class="text-sm text-red-600">{error}</p>
					<button
						type="button"
						class="mt-2 text-sm font-medium text-red-600 hover:text-red-700"
						onclick={() => tables.loadTables()}
					>
						Try again
					</button>
				</div>
			{:else if showWelcome}
				<!-- Welcome State with Template Gallery Inline -->
				<div class="h-full overflow-auto">
					<div class="mx-auto max-w-5xl px-6 py-12">
						<!-- Welcome Header -->
						<div class="mb-8 text-center">
							<div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-blue-100">
								<Database class="h-8 w-8 text-blue-600" />
							</div>
							<h2 class="text-2xl font-bold text-gray-900">Welcome to Tables</h2>
							<p class="mt-2 text-gray-500">
								Your central hub for managing all structured data. Start with a template or create from scratch.
							</p>
						</div>

						<!-- Quick Actions -->
						<div class="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
							<button
								type="button"
								class="flex items-center gap-4 rounded-xl border border-gray-200 bg-white p-4 text-left shadow-sm transition-all hover:border-blue-400 hover:shadow-md"
								onclick={() => (showAddModal = true)}
							>
								<div class="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-100">
									<Plus class="h-6 w-6 text-blue-600" />
								</div>
								<div>
									<h3 class="font-medium text-gray-900">Create Table</h3>
									<p class="text-sm text-gray-500">Start from scratch</p>
								</div>
							</button>

							<button
								type="button"
								class="flex items-center gap-4 rounded-xl border border-gray-200 bg-white p-4 text-left shadow-sm transition-all hover:border-orange-400 hover:shadow-md"
								onclick={() => (showImportModal = true)}
							>
								<div class="flex h-12 w-12 items-center justify-center rounded-lg bg-orange-100">
									<Upload class="h-6 w-6 text-orange-600" />
								</div>
								<div>
									<h3 class="font-medium text-gray-900">Import Data</h3>
									<p class="text-sm text-gray-500">CSV or Excel file</p>
								</div>
							</button>

							<button
								type="button"
								class="flex items-center gap-4 rounded-xl border border-gray-200 bg-white p-4 text-left shadow-sm transition-all hover:border-green-400 hover:shadow-md"
								onclick={() => (showTemplateGallery = true)}
							>
								<div class="flex h-12 w-12 items-center justify-center rounded-lg bg-green-100">
									<Sparkles class="h-6 w-6 text-green-600" />
								</div>
								<div>
									<h3 class="font-medium text-gray-900">Use Template</h3>
									<p class="text-sm text-gray-500">Pre-built templates</p>
								</div>
							</button>
						</div>

						<!-- Inline Template Preview -->
						<div class="rounded-xl border border-gray-200 bg-white shadow-sm">
							<TemplateGallery
								onSelectTemplate={handleSelectTemplate}
								onStartBlank={() => (showAddModal = true)}
							/>
						</div>
					</div>
				</div>
			{:else if tablesList.length === 0}
				<!-- No Results Empty State -->
				<div class="flex h-full flex-col items-center justify-center">
					<div class="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100">
						<Search class="h-8 w-8 text-gray-400" />
					</div>
					<h3 class="mb-1 text-lg font-medium text-gray-900">No tables found</h3>
					<p class="mb-4 text-sm text-gray-500">
						{#if searchQuery}
							No tables match "{searchQuery}"
						{:else}
							No tables match your filters
						{/if}
					</p>
					<button
						type="button"
						class="text-sm font-medium text-blue-600 hover:text-blue-700"
						onclick={() => {
							handleSearchChange('');
							handleSourceFilter(null);
						}}
					>
						Clear filters
					</button>
				</div>
			{:else}
				<!-- Tables Grid/List -->
				<div class="p-6">
					{#if viewMode === 'card'}
						<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
							{#each tablesList as table (table.id)}
								<TableCard
									{table}
									isFavorite={favorites.some((f) => f.id === table.id)}
									onOpen={handleTableClick}
									onToggleFavorite={handleFavoriteToggle}
									onRename={handleRename}
									onDuplicate={handleDuplicate}
									onDelete={handleDelete}
								/>
							{/each}
						</div>
					{:else}
						<!-- List View -->
						<div class="overflow-hidden rounded-xl border border-gray-200 bg-white">
							<table class="w-full text-sm">
								<thead class="bg-gray-50">
									<tr>
										<th class="px-4 py-3 text-left font-medium text-gray-600">Name</th>
										<th class="px-4 py-3 text-left font-medium text-gray-600">Rows</th>
										<th class="px-4 py-3 text-left font-medium text-gray-600">Columns</th>
										<th class="px-4 py-3 text-left font-medium text-gray-600">Source</th>
										<th class="px-4 py-3 text-left font-medium text-gray-600">Updated</th>
										<th class="w-10"></th>
									</tr>
								</thead>
								<tbody class="divide-y divide-gray-100">
									{#each tablesList as table (table.id)}
										<tr
											class="cursor-pointer hover:bg-gray-50"
											onclick={() => handleTableClick(table.id)}
										>
											<td class="px-4 py-3">
												<div class="flex items-center gap-3">
													<Table2 class="h-5 w-5 text-gray-400" />
													<span class="font-medium text-gray-900">{table.name}</span>
												</div>
											</td>
											<td class="px-4 py-3 text-gray-500">{table.row_count}</td>
											<td class="px-4 py-3 text-gray-500">{table.columns?.length || 0}</td>
											<td class="px-4 py-3">
												<span
													class="rounded-full px-2 py-0.5 text-xs font-medium {table.source === 'import'
														? 'bg-orange-100 text-orange-700'
														: table.source === 'integration'
															? 'bg-green-100 text-green-700'
															: 'bg-blue-100 text-blue-700'}"
												>
													{table.source}
												</span>
											</td>
											<td class="px-4 py-3 text-gray-500">
												{new Date(table.updated_at).toLocaleDateString()}
											</td>
											<td class="px-4 py-3">
												<button
													type="button"
													class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
													onclick={(e) => {
														e.stopPropagation();
														handleFavoriteToggle(table.id);
													}}
												>
													{#if favorites.some((f) => f.id === table.id)}
														<svg class="h-4 w-4 fill-amber-400 text-amber-400" viewBox="0 0 24 24">
															<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
														</svg>
													{:else}
														<svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
														</svg>
													{/if}
												</button>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Create Table Modal -->
<AddTableModal
	open={showAddModal}
	onClose={() => (showAddModal = false)}
	onCreate={handleCreateTable}
/>

<!-- Import Modal -->
<ImportModal
	isOpen={showImportModal}
	onClose={() => (showImportModal = false)}
	onImport={handleImport}
/>

<!-- Template Gallery Modal -->
{#if showTemplateGallery && !showWelcome}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
		<div
			class="flex max-h-[90vh] w-full max-w-4xl flex-col rounded-xl bg-white shadow-2xl"
			onclick={(e) => e.stopPropagation()}
		>
			<div class="flex items-center justify-between border-b border-gray-200 px-6 py-4">
				<h2 class="text-lg font-semibold text-gray-900">Choose a Template</h2>
				<button
					type="button"
					class="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
					onclick={() => (showTemplateGallery = false)}
				>
					<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 6L6 18M6 6l12 12" />
					</svg>
				</button>
			</div>
			<div class="flex-1 overflow-auto">
				<TemplateGallery
					onSelectTemplate={handleSelectTemplate}
					onStartBlank={() => {
						showTemplateGallery = false;
						showAddModal = true;
					}}
				/>
			</div>
		</div>
	</div>
{/if}
