<script lang="ts">
	/**
	 * GridView - Spreadsheet-like table view
	 */
	import { Plus, GripVertical } from 'lucide-svelte';
	import type { Column, Row } from '$lib/api/tables/types';
	import CellRenderer from '../cells/CellRenderer.svelte';

	interface Props {
		columns: Column[];
		rows: Row[];
		selectedRowIds: Set<string>;
		editingCell: { rowId: string; columnId: string } | null;
		columnWidths: Record<string, number>;
		onCellChange: (rowId: string, columnId: string, value: unknown) => void;
		onRowSelect: (rowId: string) => void;
		onSelectAll: () => void;
		onCellEdit: (rowId: string, columnId: string) => void;
		onCellBlur: () => void;
		onAddRow: () => void;
		onAddColumn: () => void;
		onColumnResize?: (columnId: string, width: number) => void;
	}

	let {
		columns,
		rows,
		selectedRowIds,
		editingCell,
		columnWidths,
		onCellChange,
		onRowSelect,
		onSelectAll,
		onCellEdit,
		onCellBlur,
		onAddRow,
		onAddColumn,
		onColumnResize
	}: Props = $props();

	let resizingColumn = $state<string | null>(null);
	let resizeStartX = $state(0);
	let resizeStartWidth = $state(0);

	const allSelected = $derived(
		rows.length > 0 && rows.every((row) => selectedRowIds.has(row.id))
	);

	function getColumnWidth(column: Column): number {
		return columnWidths[column.id] || column.width || 150;
	}

	function handleResizeStart(e: MouseEvent, columnId: string) {
		e.preventDefault();
		resizingColumn = columnId;
		resizeStartX = e.clientX;
		resizeStartWidth = columnWidths[columnId] || 150;

		window.addEventListener('mousemove', handleResizeMove);
		window.addEventListener('mouseup', handleResizeEnd);
	}

	function handleResizeMove(e: MouseEvent) {
		if (!resizingColumn) return;

		const diff = e.clientX - resizeStartX;
		const newWidth = Math.max(80, resizeStartWidth + diff);

		if (onColumnResize) {
			onColumnResize(resizingColumn, newWidth);
		}
	}

	function handleResizeEnd() {
		resizingColumn = null;
		window.removeEventListener('mousemove', handleResizeMove);
		window.removeEventListener('mouseup', handleResizeEnd);
	}

	function handleCellKeydown(e: KeyboardEvent, rowIndex: number, colIndex: number) {
		if (e.key === 'Tab') {
			e.preventDefault();
			const nextCol = e.shiftKey ? colIndex - 1 : colIndex + 1;

			if (nextCol >= 0 && nextCol < columns.length) {
				onCellEdit(rows[rowIndex].id, columns[nextCol].id);
			} else if (nextCol >= columns.length && rowIndex < rows.length - 1) {
				onCellEdit(rows[rowIndex + 1].id, columns[0].id);
			} else if (nextCol < 0 && rowIndex > 0) {
				onCellEdit(rows[rowIndex - 1].id, columns[columns.length - 1].id);
			}
		} else if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			if (rowIndex < rows.length - 1) {
				onCellEdit(rows[rowIndex + 1].id, columns[colIndex].id);
			}
		} else if (e.key === 'Escape') {
			onCellBlur();
		}
	}
</script>

<div class="flex-1 overflow-auto">
	<table class="w-full border-collapse">
		<!-- Header -->
		<thead class="sticky top-0 z-10 bg-gray-50">
			<tr>
				<!-- Checkbox column -->
				<th class="w-10 border-b border-r border-gray-200 bg-gray-50 px-2 py-2">
					<input
						type="checkbox"
						checked={allSelected}
						onchange={onSelectAll}
						class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
					/>
				</th>

				<!-- Row number column -->
				<th class="w-12 border-b border-r border-gray-200 bg-gray-50 px-2 py-2 text-center text-xs font-medium text-gray-500">
					#
				</th>

				<!-- Data columns -->
				{#each columns as column, colIndex}
					<th
						class="relative border-b border-r border-gray-200 bg-gray-50 px-3 py-2 text-left text-sm font-medium text-gray-700"
						style="width: {getColumnWidth(column)}px; min-width: {getColumnWidth(column)}px;"
					>
						<div class="flex items-center gap-2">
							<span class="truncate">{column.name}</span>
							{#if column.is_required}
								<span class="text-red-500">*</span>
							{/if}
						</div>

						<!-- Resize handle -->
						{#if onColumnResize}
							<button
								type="button"
								class="absolute -right-1 top-0 h-full w-2 cursor-col-resize hover:bg-blue-500/20"
								onmousedown={(e) => handleResizeStart(e, column.id)}
							></button>
						{/if}
					</th>
				{/each}

				<!-- Add column button -->
				<th class="w-10 border-b border-gray-200 bg-gray-50 px-2 py-2">
					<button
						type="button"
						class="flex h-6 w-6 items-center justify-center rounded text-gray-400 hover:bg-gray-200 hover:text-gray-600"
						onclick={onAddColumn}
					>
						<Plus class="h-4 w-4" />
					</button>
				</th>
			</tr>
		</thead>

		<!-- Body -->
		<tbody>
			{#each rows as row, rowIndex (row.id)}
				<tr class="group hover:bg-blue-50/50">
					<!-- Checkbox -->
					<td class="border-b border-r border-gray-100 bg-white px-2 py-1">
						<input
							type="checkbox"
							checked={selectedRowIds.has(row.id)}
							onchange={() => onRowSelect(row.id)}
							class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
						/>
					</td>

					<!-- Row number -->
					<td class="border-b border-r border-gray-100 bg-white px-2 py-1 text-center text-xs text-gray-400">
						{rowIndex + 1}
					</td>

					<!-- Data cells -->
					{#each columns as column, colIndex}
						{@const isEditing =
							editingCell?.rowId === row.id && editingCell?.columnId === column.id}
						<td
							class="border-b border-r border-gray-100 bg-white p-0"
							style="width: {getColumnWidth(column)}px; min-width: {getColumnWidth(column)}px;"
						>
							<div
								class="h-full w-full cursor-text px-2 py-1"
								onclick={() => onCellEdit(row.id, column.id)}
								onkeydown={(e) => handleCellKeydown(e, rowIndex, colIndex)}
								role="gridcell"
								tabindex="0"
							>
								<CellRenderer
									type={column.type}
									value={row.data[column.id]}
									options={column.options}
									editing={isEditing}
									onChange={(value) => onCellChange(row.id, column.id, value)}
									onBlur={onCellBlur}
								/>
							</div>
						</td>
					{/each}

					<!-- Empty cell for add column -->
					<td class="border-b border-gray-100 bg-white"></td>
				</tr>
			{/each}

			<!-- Add row button -->
			<tr>
				<td colspan={columns.length + 3} class="border-b border-gray-100 bg-white p-0">
					<button
						type="button"
						class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-400 hover:bg-gray-50 hover:text-gray-600"
						onclick={onAddRow}
					>
						<Plus class="h-4 w-4" />
						Add row
					</button>
				</td>
			</tr>
		</tbody>
	</table>
</div>

<style>
	/* Custom scrollbar */
	div::-webkit-scrollbar {
		width: 8px;
		height: 8px;
	}

	div::-webkit-scrollbar-track {
		background: #f1f1f1;
	}

	div::-webkit-scrollbar-thumb {
		background: #c1c1c1;
		border-radius: 4px;
	}

	div::-webkit-scrollbar-thumb:hover {
		background: #a1a1a1;
	}
</style>
