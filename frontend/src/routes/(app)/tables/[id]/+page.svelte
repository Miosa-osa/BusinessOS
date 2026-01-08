<script lang="ts">
	/**
	 * Table Detail Page
	 * Shows table data with grid view and column management
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { ArrowLeft, Loader2, AlertCircle, Plus } from 'lucide-svelte';
	import {
		tables,
		visibleColumns,
		selectedRowCount
	} from '$lib/stores/tables';
	import type { Table, TableView, Row, Column, ViewType, CreateViewData } from '$lib/api/tables/types';
	import { TableHeader, TableToolbar, GridView } from '$lib/components/tables';

	// Get table ID from route
	const tableId = $derived($page.params.id);

	// Embed mode support
	const embedSuffix = $derived(
		$page.url.searchParams.get('embed') === 'true' ? '?embed=true' : ''
	);

	// State from store
	let table = $state<Table | null>(null);
	let currentView = $state<TableView | null>(null);
	let rows = $state<Row[]>([]);
	let columns = $state<Column[]>([]);
	let selectedRowIds = $state<Set<string>>(new Set());
	let editingCell = $state<{ rowId: string; columnId: string } | null>(null);
	let loading = $state(true);
	let loadingRows = $state(false);
	let error = $state<string | null>(null);
	let searchQuery = $state('');

	// Subscribe to stores
	$effect(() => {
		const unsubscribe = tables.subscribe((state) => {
			table = state.currentTable;
			currentView = state.currentView;
			rows = state.rows;
			selectedRowIds = state.selectedRowIds;
			editingCell = state.editingCell;
			loading = state.loading;
			loadingRows = state.loadingRows;
			error = state.error;
		});
		return unsubscribe;
	});

	$effect(() => {
		const unsubscribe = visibleColumns.subscribe((cols) => {
			columns = cols;
		});
		return unsubscribe;
	});

	// Load table on mount or ID change
	$effect(() => {
		if (tableId) {
			loadTable();
		}
	});

	async function loadTable() {
		if (!tableId) return;
		const loadedTable = await tables.loadTable(tableId);
		if (loadedTable) {
			await tables.loadRows();
		}
	}

	// Navigation
	function handleBack() {
		goto(`/tables${embedSuffix}`);
	}

	// View management
	function handleViewChange(viewId: string) {
		tables.setCurrentView(viewId);
		tables.loadRows();
	}

	async function handleCreateView(type: ViewType) {
		const viewData: CreateViewData = {
			name: `New ${type.charAt(0).toUpperCase() + type.slice(1)} View`,
			type
		};
		await tables.createView(viewData);
		await tables.loadRows();
	}

	function handleFavoriteToggle() {
		if (table) {
			tables.toggleFavorite(table.id);
		}
	}

	// Row management
	async function handleAddRow() {
		// Create empty row with default values
		const emptyData: Record<string, unknown> = {};
		for (const col of columns) {
			if (col.default_value !== undefined) {
				emptyData[col.id] = col.default_value;
			}
		}
		await tables.createRow(emptyData);
	}

	function handleRowSelect(rowId: string) {
		tables.toggleRowSelection(rowId);
	}

	function handleSelectAll() {
		if (selectedRowIds.size === rows.length) {
			tables.clearSelection();
		} else {
			tables.selectAllRows();
		}
	}

	async function handleDeleteSelected() {
		if (confirm(`Delete ${selectedRowIds.size} selected rows?`)) {
			await tables.deleteSelectedRows();
		}
	}

	// Cell management
	function handleCellEdit(rowId: string, columnId: string) {
		tables.setEditingCell(rowId, columnId);
	}

	function handleCellBlur() {
		tables.setEditingCell(null, null);
	}

	async function handleCellChange(rowId: string, columnId: string, value: unknown) {
		await tables.updateCell(rowId, columnId, value);
	}

	// Column management
	function handleAddColumn() {
		// TODO: Open add column modal
		console.log('Add column');
	}

	function handleColumnResize(columnId: string, width: number) {
		if (currentView) {
			tables.updateView(currentView.id, {
				column_widths: {
					...currentView.column_widths,
					[columnId]: width
				}
			});
		}
	}

	// Search
	function handleSearchChange(query: string) {
		searchQuery = query;
		tables.loadRows({ search: query });
	}

	// Filter/Sort
	function handleAddFilter() {
		// TODO: Open filter modal
		console.log('Add filter');
	}

	function handleAddSort() {
		// TODO: Open sort modal
		console.log('Add sort');
	}

	function handleHideFields() {
		// TODO: Open hide fields modal
		console.log('Hide fields');
	}

	// Export/Import
	function handleExport() {
		// TODO: Implement export
		console.log('Export');
	}

	function handleImport() {
		// TODO: Implement import
		console.log('Import');
	}
</script>

<svelte:head>
	<title>{table?.name ?? 'Table'} | BusinessOS</title>
</svelte:head>

<div class="flex h-full flex-col bg-white">
	{#if loading && !table}
		<!-- Loading State -->
		<div class="flex h-full flex-col items-center justify-center">
			<Loader2 class="mb-4 h-8 w-8 animate-spin text-blue-600" />
			<p class="text-sm text-gray-500">Loading table...</p>
		</div>
	{:else if error && !table}
		<!-- Error State -->
		<div class="flex h-full flex-col items-center justify-center p-6">
			<div class="flex flex-col items-center rounded-lg border border-red-200 bg-red-50 p-8">
				<AlertCircle class="mb-3 h-10 w-10 text-red-500" />
				<h2 class="mb-2 text-lg font-semibold text-red-900">Failed to load table</h2>
				<p class="mb-4 text-sm text-red-700">{error}</p>
				<div class="flex gap-3">
					<button
						type="button"
						class="rounded-lg px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
						onclick={handleBack}
					>
						Go Back
					</button>
					<button
						type="button"
						class="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
						onclick={loadTable}
					>
						Try Again
					</button>
				</div>
			</div>
		</div>
	{:else if table}
		<!-- Table Header -->
		<TableHeader
			{table}
			{currentView}
			onViewChange={handleViewChange}
			onCreateView={handleCreateView}
			onFavoriteToggle={handleFavoriteToggle}
		/>

		<!-- Toolbar -->
		<TableToolbar
			{columns}
			filters={currentView?.filters ?? []}
			sorts={currentView?.sorts ?? []}
			{searchQuery}
			selectedCount={selectedRowIds.size}
			onSearchChange={handleSearchChange}
			onAddFilter={handleAddFilter}
			onAddSort={handleAddSort}
			onHideFields={handleHideFields}
			onAddRow={handleAddRow}
			onDeleteSelected={handleDeleteSelected}
			onExport={handleExport}
			onImport={handleImport}
		/>

		<!-- Grid View -->
		<div class="flex-1 overflow-hidden">
			{#if loadingRows && rows.length === 0}
				<div class="flex h-full items-center justify-center">
					<Loader2 class="h-6 w-6 animate-spin text-gray-400" />
				</div>
			{:else}
				<GridView
					{columns}
					{rows}
					{selectedRowIds}
					{editingCell}
					columnWidths={currentView?.column_widths ?? {}}
					onCellChange={handleCellChange}
					onRowSelect={handleRowSelect}
					onSelectAll={handleSelectAll}
					onCellEdit={handleCellEdit}
					onCellBlur={handleCellBlur}
					onAddRow={handleAddRow}
					onAddColumn={handleAddColumn}
					onColumnResize={handleColumnResize}
				/>
			{/if}
		</div>

		<!-- Status Bar -->
		<div class="flex items-center justify-between border-t border-gray-200 bg-gray-50 px-4 py-2 text-sm text-gray-500">
			<div class="flex items-center gap-4">
				<span>{table.row_count.toLocaleString()} rows</span>
				{#if selectedRowIds.size > 0}
					<span class="text-blue-600">{selectedRowIds.size} selected</span>
				{/if}
			</div>
			<div>
				{#if loadingRows}
					<span class="flex items-center gap-1">
						<Loader2 class="h-3 w-3 animate-spin" />
						Loading...
					</span>
				{:else}
					<span>Last updated: {new Date(table.updated_at).toLocaleString()}</span>
				{/if}
			</div>
		</div>
	{/if}
</div>
