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

	async function handleDuplicate(id: string) {
		const table = tablesList.find(t => t.id === id);
		if (!table) return;

		try {
			// Create a duplicate with " (Copy)" suffix
			const duplicateData: CreateTableData = {
				name: `${table.name} (Copy)`,
				description: table.description,
				source: 'custom',
				columns: table.columns.map((col, i) => ({
					name: col.name,
					type: col.type,
					order: i,
					is_primary: i === 0
				}))
			};

			const newTable = await tables.createTable(duplicateData);
			if (newTable?.id) {
				goto(`/tables/${newTable.id}${embedSuffix}`);
			}
		} catch (error) {
			console.error('Failed to duplicate table:', error);
			alert('Failed to duplicate table. Please try again.');
		}
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

<div class="flex h-full" style="background: var(--dbg);">
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
		<div class="flex items-center justify-between px-6 py-4" style="border-bottom: 1px solid var(--dbd); background: var(--dbg);">
			<div class="flex items-center gap-4">
				<!-- Sidebar Toggle -->
				<button
					type="button"
					class="btn-pill btn-pill-ghost btn-pill-icon"
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
					<h1 class="text-xl font-semibold" style="color: var(--dt);">Tables</h1>
					<p class="text-sm" style="color: var(--dt3);">
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
					<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" style="color: var(--dt4);" />
					<input
						type="text"
						placeholder="Search tables..."
						value={searchQuery}
						oninput={(e) => handleSearchChange((e.target as HTMLInputElement).value)}
						class="dt2-search-input"
					/>
				</div>

				<!-- View Switcher -->
				<div class="flex items-center rounded-lg p-1" style="border: 1px solid var(--dbd); background: var(--dbg);">
					<button
						type="button"
						class="dt2-view-btn {viewMode === 'card' ? 'dt2-view-btn--active' : ''}"
						onclick={() => handleViewChange('card')}
						title="Card view"
					>
						<LayoutGrid class="h-4 w-4" />
					</button>
					<button
						type="button"
						class="dt2-view-btn {viewMode === 'list' ? 'dt2-view-btn--active' : ''}"
						onclick={() => handleViewChange('list')}
						title="List view"
					>
						<List class="h-4 w-4" />
					</button>
				</div>

				<!-- Actions -->
				<button
					type="button"
					class="btn-cta flex items-center gap-2"
					onclick={() => (showImportModal = true)}
				>
					<Upload class="h-4 w-4" />
					Import
				</button>

				<button
					type="button"
					class="btn-cta"
					onclick={() => (showTemplateGallery = true)}
				>
					<Plus class="h-4 w-4" />
					New Table
				</button>
			</div>
		</div>

		<!-- Source Filter Tabs -->
		{#if tablesList.length > 0 || searchQuery}
			<div class="flex items-center gap-2 px-6 py-2" style="border-bottom: 1px solid var(--dbd); background: var(--dbg);">
				<button
					type="button"
					class="dt2-filter-pill {sourceFilter === null ? 'dt2-filter-pill--active' : ''}"
					onclick={() => handleSourceFilter(null)}
				>
					All
				</button>
				<button
					type="button"
					class="dt2-filter-pill {sourceFilter === 'custom' ? 'dt2-filter-pill--active' : ''} flex items-center gap-1.5"
					onclick={() => handleSourceFilter('custom')}
				>
					<Database class="h-3.5 w-3.5" />
					Custom
				</button>
				<button
					type="button"
					class="dt2-filter-pill {sourceFilter === 'import' ? 'dt2-filter-pill--active dt2-filter-pill--orange' : ''} flex items-center gap-1.5"
					onclick={() => handleSourceFilter('import')}
				>
					<Upload class="h-3.5 w-3.5" />
					Imported
				</button>
				<button
					type="button"
					class="dt2-filter-pill {sourceFilter === 'integration' ? 'dt2-filter-pill--active dt2-filter-pill--green' : ''} flex items-center gap-1.5"
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
					<Loader2 class="mb-4 h-8 w-8 animate-spin" style="color: var(--accent-blue);" />
					<p class="text-sm" style="color: var(--dt3);">Loading tables...</p>
				</div>
			{:else if error}
				<div class="m-6 rounded-lg p-4" style="border: 1px solid var(--bos-status-error); background: color-mix(in srgb, var(--bos-status-error) 8%, var(--dbg));">
					<p class="text-sm" style="color: var(--bos-status-error-text);">{error}</p>
					<button
						type="button"
						class="btn-pill btn-pill-ghost btn-pill-sm mt-2"
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
						<div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full" style="background: color-mix(in srgb, var(--accent-blue) 15%, var(--dbg));">
							<Database class="h-8 w-8" style="color: var(--accent-blue);" />
						</div>
						<h2 class="text-2xl font-bold" style="color: var(--dt);">Welcome to Tables</h2>
						<p class="mt-2" style="color: var(--dt2);">
								Your central hub for managing all structured data. Start with a template or create from scratch.
							</p>
						</div>

						<!-- Quick Actions -->
						<div class="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
							<button
								type="button"
								class="btn-cta flex items-center gap-4 p-4 text-left"
								onclick={() => (showAddModal = true)}
							>
								<div class="flex h-12 w-12 items-center justify-center rounded-lg" style="background: color-mix(in srgb, var(--accent-blue) 15%, var(--dbg));">
									<Plus class="h-6 w-6" style="color: var(--accent-blue);" />
								</div>
								<div>
									<h3 class="font-medium" style="color: var(--dt);">Create Table</h3>
									<p class="text-sm" style="color: var(--dt2);">Start from scratch</p>
								</div>
							</button>

							<button
								type="button"
								class="btn-cta flex items-center gap-4 p-4 text-left"
								onclick={() => (showImportModal = true)}
							>
								<div class="flex h-12 w-12 items-center justify-center rounded-lg" style="background: color-mix(in srgb, var(--accent-orange) 15%, var(--dbg));">
									<Upload class="h-6 w-6" style="color: var(--accent-orange);" />
								</div>
								<div>
									<h3 class="font-medium" style="color: var(--dt);">Import Data</h3>
									<p class="text-sm" style="color: var(--dt2);">CSV or Excel file</p>
								</div>
							</button>

							<button
								type="button"
								class="btn-cta flex items-center gap-4 p-4 text-left"
								onclick={() => (showTemplateGallery = true)}
							>
								<div class="flex h-12 w-12 items-center justify-center rounded-lg" style="background: color-mix(in srgb, var(--accent-green) 15%, var(--dbg));">
									<Sparkles class="h-6 w-6" style="color: var(--accent-green);" />
								</div>
								<div>
									<h3 class="font-medium" style="color: var(--dt);">Use Template</h3>
									<p class="text-sm" style="color: var(--dt2);">Pre-built templates</p>
								</div>
							</button>
						</div>

						<!-- Inline Template Preview -->
						<div class="rounded-xl" style="border: 1px solid var(--dbd); background: var(--dbg); box-shadow: var(--shadow-sm);">
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
					<div class="mb-4 flex h-16 w-16 items-center justify-center rounded-full" style="background: var(--dbg2);">
						<Search class="h-8 w-8" style="color: var(--dt4);" />
					</div>
					<h3 class="mb-1 text-lg font-medium" style="color: var(--dt);">No tables found</h3>
					<p class="mb-4 text-sm" style="color: var(--dt2);">
						{#if searchQuery}
							No tables match "{searchQuery}"
						{:else}
							No tables match your filters
						{/if}
					</p>
					<button
						type="button"
						class="btn-pill btn-pill-ghost btn-pill-sm"
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
						<div class="overflow-hidden rounded-xl" style="border: 1px solid var(--dbd); background: var(--dbg);">
							<table class="w-full text-sm">
								<thead style="background: var(--dbg2);">
									<tr>
										<th class="px-4 py-3 text-left font-medium" style="color: var(--dt2);">Name</th>
										<th class="px-4 py-3 text-left font-medium" style="color: var(--dt2);">Rows</th>
										<th class="px-4 py-3 text-left font-medium" style="color: var(--dt2);">Columns</th>
										<th class="px-4 py-3 text-left font-medium" style="color: var(--dt2);">Source</th>
										<th class="px-4 py-3 text-left font-medium" style="color: var(--dt2);">Updated</th>
										<th class="w-10"></th>
									</tr>
								</thead>
								<tbody class="dt2-tbody">
									{#each tablesList as table (table.id)}
										<tr
											class="dt2-row"
											onclick={() => handleTableClick(table.id)}
										>
											<td class="px-4 py-3">
												<div class="flex items-center gap-3">
													<Table2 class="h-5 w-5" style="color: var(--dt4);" />
													<span class="font-medium" style="color: var(--dt);">{table.name}</span>
												</div>
											</td>
											<td class="px-4 py-3" style="color: var(--dt2);">{table.row_count}</td>
											<td class="px-4 py-3" style="color: var(--dt2);">{table.columns?.length || 0}</td>
											<td class="px-4 py-3">
												<span
													class="dt2-source-badge {table.source === 'import'
														? 'dt2-source-badge--orange'
														: table.source === 'integration'
															? 'dt2-source-badge--green'
															: 'dt2-source-badge--blue'}"
												>
													{table.source}
												</span>
											</td>
											<td class="px-4 py-3" style="color: var(--dt3);">
												{new Date(table.updated_at).toLocaleDateString()}
											</td>
											<td class="px-4 py-3">
												<button
													type="button"
													class="btn-pill btn-pill-ghost btn-pill-icon"
													onclick={(e) => {
														e.stopPropagation();
														handleFavoriteToggle(table.id);
													}}
												>
													{#if favorites.some((f) => f.id === table.id)}
														<svg class="h-4 w-4" style="fill: var(--bos-status-warning); color: var(--bos-status-warning)" viewBox="0 0 24 24">
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
	<div class="fixed inset-0 z-50 flex items-center justify-center" style="background: var(--bos-modal-backdrop)" onclick={() => (showTemplateGallery = false)}>
		<div
			class="flex max-h-[90vh] w-full max-w-4xl flex-col rounded-xl"
			style="background: var(--dbg); box-shadow: var(--shadow-xl);"
			onclick={(e) => e.stopPropagation()}
		>
			<div class="flex items-center justify-between px-6 py-4" style="border-bottom: 1px solid var(--dbd);">
				<h2 class="text-lg font-semibold" style="color: var(--dt);">Choose a Template</h2>
				<button
					type="button"
					class="btn-pill btn-pill-ghost btn-pill-icon"
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

<style>
	/* ── Tables Page Foundation Styles (dt2- prefix) ── */

	/* Search input */
	.dt2-search-input {
		width: 16rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--dbd);
		background: var(--dbg2);
		padding: 0.5rem 0.75rem 0.5rem 2.25rem;
		font-size: var(--text-sm);
		color: var(--dt);
		outline: none;
		transition: all 200ms ease;
	}
	.dt2-search-input::placeholder {
		color: var(--dt4);
	}
	.dt2-search-input:focus {
		border-color: var(--accent-blue);
		background: var(--dbg);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-blue) 20%, transparent);
	}

	/* View mode toggle buttons */
	.dt2-view-btn {
		border-radius: var(--radius-xs);
		padding: 0.375rem;
		transition: all 200ms ease;
		color: var(--dt4);
		background: none;
		border: none;
		cursor: pointer;
	}
	.dt2-view-btn:hover {
		color: var(--dt2);
	}
	.dt2-view-btn--active {
		background: var(--dbg2);
		color: var(--dt);
	}

	/* Filter pills */
	.dt2-filter-pill {
		border-radius: 9999px;
		padding: 0.25rem 0.75rem;
		font-size: var(--text-sm);
		font-weight: 500;
		transition: all 200ms ease;
		color: var(--dt3);
		background: none;
		border: none;
		cursor: pointer;
	}
	.dt2-filter-pill:hover {
		background: var(--dbg2);
		color: var(--dt2);
	}
	.dt2-filter-pill--active {
		background: color-mix(in srgb, var(--accent-blue) 15%, var(--dbg));
		color: var(--accent-blue);
	}
	.dt2-filter-pill--active.dt2-filter-pill--orange {
		background: color-mix(in srgb, var(--accent-orange) 15%, var(--dbg));
		color: var(--accent-orange);
	}
	.dt2-filter-pill--active.dt2-filter-pill--green {
		background: color-mix(in srgb, var(--accent-green) 15%, var(--dbg));
		color: var(--accent-green);
	}

	/* Table list view rows */
	:global(.dt2-tbody) {
		border-top: 1px solid var(--dbd2);
	}
	:global(.dt2-tbody > tr + tr) {
		border-top: 1px solid var(--dbd2);
	}
	.dt2-row {
		cursor: pointer;
		transition: background 150ms ease;
	}
	.dt2-row:hover {
		background: var(--dbg2);
	}

	/* Source badges */
	.dt2-source-badge {
		border-radius: 9999px;
		padding: 0.125rem 0.5rem;
		font-size: 0.75rem;
		font-weight: 500;
	}
	.dt2-source-badge--blue {
		background: color-mix(in srgb, var(--accent-blue) 15%, var(--dbg));
		color: var(--accent-blue);
	}
	.dt2-source-badge--orange {
		background: color-mix(in srgb, var(--accent-orange) 15%, var(--dbg));
		color: var(--accent-orange);
	}
	.dt2-source-badge--green {
		background: color-mix(in srgb, var(--accent-green) 15%, var(--dbg));
		color: var(--accent-green);
	}
</style>
