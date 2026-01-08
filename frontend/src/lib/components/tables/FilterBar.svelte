<script lang="ts">
	/**
	 * FilterBar - Displays active filters as pills with quick actions
	 * Shows current filters, allows removal, and opens filter modal
	 */
	import { X, Filter, Plus } from 'lucide-svelte';
	import type { Filter as FilterType, Column, FilterOperator } from '$lib/api/tables/types';

	interface Props {
		filters: FilterType[];
		columns: Column[];
		onRemoveFilter: (filterId: string) => void;
		onClearAll: () => void;
		onAddFilter: () => void;
		onEditFilter: (filter: FilterType) => void;
	}

	let { filters, columns, onRemoveFilter, onClearAll, onAddFilter, onEditFilter }: Props = $props();

	// Get column name by ID
	function getColumnName(columnId: string): string {
		const column = columns.find((c) => c.id === columnId);
		return column?.name ?? 'Unknown';
	}

	// Get human-readable operator label
	function getOperatorLabel(operator: FilterOperator): string {
		const labels: Record<FilterOperator, string> = {
			eq: 'is',
			neq: 'is not',
			gt: '>',
			gte: '>=',
			lt: '<',
			lte: '<=',
			contains: 'contains',
			not_contains: 'does not contain',
			starts_with: 'starts with',
			ends_with: 'ends with',
			is_empty: 'is empty',
			is_not_empty: 'is not empty',
			is_null: 'is null',
			is_not_null: 'is not null',
			in: 'is any of',
			not_in: 'is none of',
			is_within: 'is within',
			is_before: 'is before',
			is_after: 'is after',
			is_on_or_before: 'is on or before',
			is_on_or_after: 'is on or after'
		};
		return labels[operator] ?? operator;
	}

	// Format filter value for display
	function formatValue(value: unknown): string {
		if (value === null || value === undefined) return '';
		if (Array.isArray(value)) return value.join(', ');
		if (typeof value === 'boolean') return value ? 'Yes' : 'No';
		if (value instanceof Date) return value.toLocaleDateString();
		return String(value);
	}

	// Check if operator needs a value display
	function needsValue(operator: FilterOperator): boolean {
		return !['is_empty', 'is_not_empty', 'is_null', 'is_not_null'].includes(operator);
	}
</script>

{#if filters.length > 0}
	<div class="flex flex-wrap items-center gap-2 border-b border-gray-100 bg-gray-50/50 px-4 py-2">
		<div class="flex items-center gap-1.5 text-xs font-medium text-gray-500">
			<Filter class="h-3.5 w-3.5" />
			<span>Filters:</span>
		</div>

		{#each filters as filter (filter.id)}
			<div
				class="group flex items-center gap-1.5 rounded-full border border-blue-200 bg-blue-50 px-2.5 py-1 text-xs hover:border-blue-300 hover:bg-blue-100"
			>
				<!-- Filter content (clickable for edit) -->
				<button
					type="button"
					class="flex items-center gap-1.5"
					onclick={() => onEditFilter(filter)}
				>
					<span class="font-medium text-blue-700">{getColumnName(filter.column_id)}</span>
					<span class="text-blue-500">{getOperatorLabel(filter.operator)}</span>
					{#if needsValue(filter.operator) && filter.value !== undefined}
						<span class="max-w-[120px] truncate font-medium text-blue-700">
							"{formatValue(filter.value)}"
						</span>
					{/if}
				</button>

				<!-- Remove button -->
				<button
					type="button"
					class="ml-0.5 rounded-full p-0.5 text-blue-400 hover:bg-blue-200 hover:text-blue-600"
					onclick={() => onRemoveFilter(filter.id)}
					aria-label="Remove filter"
				>
					<X class="h-3 w-3" />
				</button>
			</div>
		{/each}

		<!-- Add filter button -->
		<button
			type="button"
			class="flex items-center gap-1 rounded-full border border-dashed border-gray-300 px-2.5 py-1 text-xs text-gray-500 hover:border-gray-400 hover:bg-gray-100 hover:text-gray-600"
			onclick={onAddFilter}
		>
			<Plus class="h-3 w-3" />
			Add
		</button>

		<!-- Clear all -->
		{#if filters.length > 1}
			<button
				type="button"
				class="ml-auto text-xs text-gray-500 hover:text-gray-700"
				onclick={onClearAll}
			>
				Clear all
			</button>
		{/if}
	</div>
{/if}
