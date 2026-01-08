<script lang="ts">
	/**
	 * Table Detail Page
	 * Shows table data with multiple view types, filtering, sorting, and row expand
	 */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { ArrowLeft, Loader2, AlertCircle, Plus, Download } from 'lucide-svelte';
	import {
		tables,
		visibleColumns,
		selectedRowCount
	} from '$lib/stores/tables';
	import type { Table, TableView, Row, Column, ViewType, CreateViewData, Filter, Sort, CreateColumnData } from '$lib/api/tables/types';
	import {
		TableHeader,
		TableToolbar,
		GridView,
		KanbanView,
		GalleryView,
		FilterBar,
		FilterModal,
		SortModal,
		FieldsPanel,
		RowExpandModal,
		AddColumnModal
	} from '$lib/components/tables';

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

	// Modal/Panel states
	let showAddColumnModal = $state(false);
	let showFilterModal = $state(false);
	let showSortModal = $state(false);
	let showFieldsPanel = $state(false);
	let showRowExpand = $state(false);
	let editingFilter = $state<Filter | null>(null);
	let expandedRow = $state<Row | null>(null);
	let expandedRowIndex = $state(0);

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
	async function handleAddRow(groupValue?: string) {
		const emptyData: Record<string, unknown> = {};
		for (const col of columns) {
			if (col.default_value !== undefined) {
				emptyData[col.id] = col.default_value;
			}
		}
		if (groupValue && currentView?.kanban_column_id) {
			emptyData[currentView.kanban_column_id] = groupValue;
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
		showAddColumnModal = true;
	}

	async function handleCreateColumn(columnData: CreateColumnData) {
		await tables.addColumn(columnData);
		showAddColumnModal = false;
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

	// Filter management
	function handleAddFilter() {
		editingFilter = null;
		showFilterModal = true;
	}

	function handleEditFilter(filter: Filter) {
		editingFilter = filter;
		showFilterModal = true;
	}

	async function handleSaveFilter(filterData: Omit<Filter, 'id'> & { id?: string }) {
		if (!currentView) return;

		const currentFilters = currentView.filters || [];

		if (filterData.id) {
			const updatedFilters = currentFilters.map((f) =>
				f.id === filterData.id ? { ...f, ...filterData } : f
			);
			await tables.updateView(currentView.id, { filters: updatedFilters });
		} else {
			const newFilter: Filter = {
				...filterData,
				id: crypto.randomUUID()
			};
			await tables.updateView(currentView.id, {
				filters: [...currentFilters, newFilter]
			});
		}

		await tables.loadRows();
		showFilterModal = false;
		editingFilter = null;
	}

	async function handleRemoveFilter(filterId: string) {
		if (!currentView) return;

		const updatedFilters = (currentView.filters || []).filter((f) => f.id !== filterId);
		await tables.updateView(currentView.id, { filters: updatedFilters });
		await tables.loadRows();
	}

	async function handleClearAllFilters() {
		if (!currentView) return;

		await tables.updateView(currentView.id, { filters: [] });
		await tables.loadRows();
	}

	// Sort management
	function handleAddSort() {
		showSortModal = true;
	}

	async function handleSaveSorts(sorts: Sort[]) {
		if (!currentView) return;

		await tables.updateView(currentView.id, { sorts });
		await tables.loadRows();
	}

	// Fields/Columns visibility
	function handleHideFields() {
		showFieldsPanel = true;
	}

	function handleToggleColumnVisibility(columnId: string) {
		if (!currentView) return;

		const hiddenColumns = currentView.hidden_columns || [];
		const isHidden = hiddenColumns.includes(columnId);

		const updatedHidden = isHidden
			? hiddenColumns.filter((id) => id !== columnId)
			: [...hiddenColumns, columnId];

		tables.updateView(currentView.id, { hidden_columns: updatedHidden });
	}

	function handleShowAllColumns() {
		if (!currentView) return;
		tables.updateView(currentView.id, { hidden_columns: [] });
	}

	function handleHideAllColumns() {
		if (!currentView) return;
		// Keep primary column visible
		const primaryColumn = columns.find((c) => c.is_primary);
		const allColumnIds = columns.filter((c) => c.id !== primaryColumn?.id).map((c) => c.id);
		tables.updateView(currentView.id, { hidden_columns: allColumnIds });
	}

	// Export
	function handleExport() {
		if (!table || rows.length === 0) return;

		// Build CSV content
		const visibleCols = columns.filter((c) => !currentView?.hidden_columns?.includes(c.id));
		const headers = visibleCols.map((c) => `"${c.name}"`).join(',');
		const rowData = rows.map((row) => {
			return visibleCols
				.map((col) => {
					const value = row.data[col.id];
					if (value === null || value === undefined) return '';
					if (typeof value === 'string') return `"${value.replace(/"/g, '""')}"`;
					return String(value);
				})
				.join(',');
		});

		const csv = [headers, ...rowData].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `${table.name}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	function handleImport() {
		// Redirect to import on tables list page
		goto(`/tables?import=true${embedSuffix}`);
	}

	// Row Expand (Card Click)
	function handleCardClick(rowId: string) {
		const rowIndex = rows.findIndex((r) => r.id === rowId);
		if (rowIndex >= 0) {
			expandedRow = rows[rowIndex];
			expandedRowIndex = rowIndex;
			showRowExpand = true;
		}
	}

	function handleRowExpandClose() {
		showRowExpand = false;
		expandedRow = null;
	}

	async function handleRowExpandCellChange(columnId: string, value: unknown) {
		if (!expandedRow) return;
		await tables.updateCell(expandedRow.id, columnId, value);
		// Update local state
		expandedRow = { ...expandedRow, data: { ...expandedRow.data, [columnId]: value } };
	}

	async function handleRowExpandDelete() {
		if (!expandedRow) return;
		if (confirm('Delete this row?')) {
			await tables.deleteRow(expandedRow.id);
			showRowExpand = false;
			expandedRow = null;
		}
	}

	function handleRowExpandDuplicate() {
		if (!expandedRow) return;
		// Create new row with same data
		const newData = { ...expandedRow.data };
		tables.createRow(newData);
		showRowExpand = false;
		expandedRow = null;
	}

	function handleRowExpandNavigate(direction: 'prev' | 'next') {
		const newIndex = direction === 'prev' ? expandedRowIndex - 1 : expandedRowIndex + 1;
		if (newIndex >= 0 && newIndex < rows.length) {
			expandedRowIndex = newIndex;
			expandedRow = rows[newIndex];
		}
	}

	// Kanban-specific handlers
	async function handleCardMove(rowId: string, newGroupValue: string) {
		if (!currentView?.kanban_column_id) return;
		await tables.updateCell(rowId, currentView.kanban_column_id, newGroupValue);
	}

	function handleAddGroup() {
		// TODO: Open modal to add new select choice to kanban column
		const kanbanCol = columns.find((c) => c.id === currentView?.kanban_column_id);
		if (kanbanCol) {
			alert(`Add new option to "${kanbanCol.name}" column. Feature coming soon!`);
		}
	}

	// Computed values
	const viewType = $derived(currentView?.type || 'grid');
	const kanbanGroupColumnId = $derived(currentView?.kanban_column_id || null);
	const galleryCoverColumnId = $derived(currentView?.gallery_cover_column_id || null);
	const activeFilters = $derived(currentView?.filters || []);
	const activeSorts = $derived(currentView?.sorts || []);
	const hiddenColumns = $derived(currentView?.hidden_columns || []);
	const allColumns = $derived(table?.columns || []);
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
			filters={activeFilters}
			sorts={activeSorts}
			{searchQuery}
			selectedCount={selectedRowIds.size}
			onSearchChange={handleSearchChange}
			onAddFilter={handleAddFilter}
			onAddSort={handleAddSort}
			onHideFields={handleHideFields}
			onAddRow={() => handleAddRow()}
			onDeleteSelected={handleDeleteSelected}
			onExport={handleExport}
			onImport={handleImport}
		/>

		<!-- Filter Bar (when filters are active) -->
		{#if activeFilters.length > 0}
			<FilterBar
				filters={activeFilters}
				{columns}
				onRemoveFilter={handleRemoveFilter}
				onClearAll={handleClearAllFilters}
				onAddFilter={handleAddFilter}
				onEditFilter={handleEditFilter}
			/>
		{/if}

		<!-- Dynamic View Rendering -->
		<div class="flex-1 overflow-hidden">
			{#if loadingRows && rows.length === 0}
				<div class="flex h-full items-center justify-center">
					<Loader2 class="h-6 w-6 animate-spin text-gray-400" />
				</div>
			{:else if viewType === 'grid'}
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
					onAddRow={() => handleAddRow()}
					onAddColumn={handleAddColumn}
					onColumnResize={handleColumnResize}
				/>
			{:else if viewType === 'kanban'}
				<KanbanView
					{columns}
					{rows}
					groupColumnId={kanbanGroupColumnId}
					onCardClick={handleCardClick}
					onCardMove={handleCardMove}
					onAddCard={handleAddRow}
					onAddGroup={handleAddGroup}
				/>
			{:else if viewType === 'gallery'}
				<GalleryView
					{columns}
					{rows}
					coverColumnId={galleryCoverColumnId}
					onCardClick={handleCardClick}
					onAddCard={() => handleAddRow()}
				/>
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
					onAddRow={() => handleAddRow()}
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
				{#if activeFilters.length > 0}
					<span class="text-orange-600">{activeFilters.length} filter{activeFilters.length !== 1 ? 's' : ''}</span>
				{/if}
				{#if activeSorts.length > 0}
					<span class="text-purple-600">{activeSorts.length} sort{activeSorts.length !== 1 ? 's' : ''}</span>
				{/if}
				{#if hiddenColumns.length > 0}
					<span class="text-gray-500">{hiddenColumns.length} hidden</span>
				{/if}
			</div>
			<div class="flex items-center gap-3">
				<span class="text-xs text-gray-400 capitalize">{viewType} view</span>
				{#if loadingRows}
					<span class="flex items-center gap-1">
						<Loader2 class="h-3 w-3 animate-spin" />
						Loading...
					</span>
				{:else}
					<span>Updated: {new Date(table.updated_at).toLocaleString()}</span>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Add Column Modal -->
<AddColumnModal
	open={showAddColumnModal}
	onClose={() => (showAddColumnModal = false)}
	onCreate={handleCreateColumn}
/>

<!-- Filter Modal -->
<FilterModal
	open={showFilterModal}
	{columns}
	editFilter={editingFilter}
	onClose={() => {
		showFilterModal = false;
		editingFilter = null;
	}}
	onSave={handleSaveFilter}
/>

<!-- Sort Modal -->
<SortModal
	open={showSortModal}
	columns={allColumns}
	sorts={activeSorts}
	onClose={() => (showSortModal = false)}
	onSave={handleSaveSorts}
/>

<!-- Fields Panel -->
<FieldsPanel
	open={showFieldsPanel}
	columns={allColumns}
	{hiddenColumns}
	onClose={() => (showFieldsPanel = false)}
	onToggleColumn={handleToggleColumnVisibility}
	onShowAll={handleShowAllColumns}
	onHideAll={handleHideAllColumns}
/>

<!-- Row Expand Modal -->
<RowExpandModal
	isOpen={showRowExpand}
	row={expandedRow}
	columns={allColumns}
	rowIndex={expandedRowIndex}
	totalRows={rows.length}
	onClose={handleRowExpandClose}
	onCellChange={handleRowExpandCellChange}
	onDelete={handleRowExpandDelete}
	onDuplicate={handleRowExpandDuplicate}
	onNavigate={handleRowExpandNavigate}
/>
