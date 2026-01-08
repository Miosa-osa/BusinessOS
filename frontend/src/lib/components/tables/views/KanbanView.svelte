<script lang="ts">
	/**
	 * KanbanView - Board view with cards grouped by a select column
	 * Features: Drag-and-drop cards, column headers with counts, card previews
	 */
	import { Plus, GripVertical, MoreHorizontal } from 'lucide-svelte';
	import type { Column, Row, SelectChoice } from '$lib/api/tables/types';

	interface Props {
		columns: Column[];
		rows: Row[];
		groupColumnId: string | null;
		onCardClick: (rowId: string) => void;
		onCardMove: (rowId: string, newGroupValue: string) => void;
		onAddCard: (groupValue: string) => void;
		onAddGroup?: () => void;
	}

	let {
		columns,
		rows,
		groupColumnId,
		onCardClick,
		onCardMove,
		onAddCard,
		onAddGroup
	}: Props = $props();

	// Get the grouping column
	const groupColumn = $derived(columns.find((c) => c.id === groupColumnId));

	// Get primary display column (first text column or first column)
	const primaryColumn = $derived(
		columns.find((c) => c.is_primary) || columns.find((c) => c.type === 'text') || columns[0]
	);

	// Get available groups from the select column choices
	const groups = $derived.by(() => {
		if (!groupColumn || !groupColumn.options?.choices) {
			// If no group column, create a single "All Items" group
			return [{ id: 'all', label: 'All Items', color: '#64748b', order: 0 }];
		}
		// Add an "Uncategorized" group for items without a value
		return [
			...groupColumn.options.choices,
			{ id: '__uncategorized__', label: 'Uncategorized', color: '#94a3b8', order: 999 }
		];
	});

	// Group rows by the group column value
	const rowsByGroup = $derived.by(() => {
		const grouped: Record<string, Row[]> = {};

		// Initialize all groups
		for (const group of groups) {
			grouped[group.id] = [];
		}

		// Sort rows into groups
		for (const row of rows) {
			if (!groupColumnId) {
				// No grouping, all in one group
				grouped['all'] = grouped['all'] || [];
				grouped['all'].push(row);
			} else {
				const groupValue = row.data[groupColumnId];
				if (groupValue && typeof groupValue === 'string' && grouped[groupValue]) {
					grouped[groupValue].push(row);
				} else {
					// Uncategorized
					grouped['__uncategorized__'].push(row);
				}
			}
		}

		return grouped;
	});

	// Drag state
	let draggingRowId = $state<string | null>(null);
	let dragOverGroupId = $state<string | null>(null);

	function handleDragStart(e: DragEvent, rowId: string) {
		draggingRowId = rowId;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
			e.dataTransfer.setData('text/plain', rowId);
		}
	}

	function handleDragEnd() {
		draggingRowId = null;
		dragOverGroupId = null;
	}

	function handleDragOver(e: DragEvent, groupId: string) {
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}
		dragOverGroupId = groupId;
	}

	function handleDragLeave() {
		dragOverGroupId = null;
	}

	function handleDrop(e: DragEvent, groupId: string) {
		e.preventDefault();
		if (draggingRowId && groupId !== '__uncategorized__') {
			onCardMove(draggingRowId, groupId);
		}
		draggingRowId = null;
		dragOverGroupId = null;
	}

	// Get display value for a row
	function getDisplayValue(row: Row, column: Column | undefined): string {
		if (!column) return '';
		const value = row.data[column.id];
		if (value === null || value === undefined) return '';
		if (typeof value === 'boolean') return value ? 'Yes' : 'No';
		return String(value);
	}

	// Get preview fields (first few columns excluding group column)
	const previewColumns = $derived(
		columns
			.filter((c) => c.id !== groupColumnId && c.id !== primaryColumn?.id && !c.is_hidden)
			.slice(0, 3)
	);
</script>

<div class="flex h-full gap-4 overflow-x-auto p-4">
	{#each groups as group (group.id)}
		{@const groupRows = rowsByGroup[group.id] || []}
		<div
			class="flex w-72 flex-shrink-0 flex-col rounded-lg bg-gray-100"
			ondragover={(e) => handleDragOver(e, group.id)}
			ondragleave={handleDragLeave}
			ondrop={(e) => handleDrop(e, group.id)}
		>
			<!-- Column Header -->
			<div class="flex items-center justify-between p-3">
				<div class="flex items-center gap-2">
					<div
						class="h-3 w-3 rounded-full"
						style="background-color: {group.color}"
					></div>
					<h3 class="text-sm font-semibold text-gray-700">{group.label}</h3>
					<span class="rounded-full bg-gray-200 px-2 py-0.5 text-xs text-gray-500">
						{groupRows.length}
					</span>
				</div>
				<button
					type="button"
					class="rounded p-1 text-gray-400 hover:bg-gray-200 hover:text-gray-600"
				>
					<MoreHorizontal class="h-4 w-4" />
				</button>
			</div>

			<!-- Cards Container -->
			<div
				class="flex-1 space-y-2 overflow-y-auto p-2 {dragOverGroupId === group.id
					? 'bg-blue-50'
					: ''}"
			>
				{#each groupRows as row (row.id)}
					<button
						type="button"
						draggable="true"
						ondragstart={(e) => handleDragStart(e, row.id)}
						ondragend={handleDragEnd}
						onclick={() => onCardClick(row.id)}
						class="group w-full rounded-lg border bg-white p-3 text-left shadow-sm transition-shadow hover:shadow-md {draggingRowId ===
						row.id
							? 'opacity-50'
							: ''}"
					>
						<!-- Drag Handle -->
						<div class="mb-2 flex items-center justify-between">
							<GripVertical class="h-4 w-4 cursor-grab text-gray-300 group-hover:text-gray-400" />
						</div>

						<!-- Primary Value -->
						<div class="mb-2 font-medium text-gray-900">
							{getDisplayValue(row, primaryColumn) || 'Untitled'}
						</div>

						<!-- Preview Fields -->
						{#if previewColumns.length > 0}
							<div class="space-y-1">
								{#each previewColumns as col}
									{@const value = getDisplayValue(row, col)}
									{#if value}
										<div class="flex items-center gap-2 text-xs">
											<span class="text-gray-400">{col.name}:</span>
											<span class="truncate text-gray-600">{value}</span>
										</div>
									{/if}
								{/each}
							</div>
						{/if}
					</button>
				{/each}

				<!-- Empty State -->
				{#if groupRows.length === 0}
					<div class="py-8 text-center text-sm text-gray-400">
						No items
					</div>
				{/if}
			</div>

			<!-- Add Card Button -->
			{#if group.id !== '__uncategorized__'}
				<div class="p-2">
					<button
						type="button"
						class="flex w-full items-center justify-center gap-1 rounded-lg border border-dashed border-gray-300 py-2 text-sm text-gray-500 hover:border-gray-400 hover:bg-gray-50 hover:text-gray-600"
						onclick={() => onAddCard(group.id)}
					>
						<Plus class="h-4 w-4" />
						Add card
					</button>
				</div>
			{/if}
		</div>
	{/each}

	<!-- Add Group Button -->
	{#if onAddGroup && groupColumn}
		<div class="flex w-72 flex-shrink-0 items-start">
			<button
				type="button"
				class="flex items-center gap-2 rounded-lg border border-dashed border-gray-300 px-4 py-3 text-sm text-gray-500 hover:border-gray-400 hover:bg-gray-50 hover:text-gray-600"
				onclick={onAddGroup}
			>
				<Plus class="h-4 w-4" />
				Add group
			</button>
		</div>
	{/if}
</div>
