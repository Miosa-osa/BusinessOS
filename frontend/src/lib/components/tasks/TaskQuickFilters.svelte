<script lang="ts">
	type QuickFilter = 'my-tasks' | 'all' | 'overdue' | 'today' | 'this-week' | 'blocked' | 'unassigned';

	interface Props {
		activeFilter?: QuickFilter;
		counts?: Partial<Record<QuickFilter, number>>;
		onFilterChange?: (filter: QuickFilter) => void;
	}

	let { activeFilter = 'all', counts = {}, onFilterChange }: Props = $props();

	const filters: { value: QuickFilter; label: string }[] = [
		{ value: 'my-tasks', label: 'My Tasks' },
		{ value: 'all', label: 'All Tasks' },
		{ value: 'overdue', label: 'Overdue' },
		{ value: 'today', label: 'Due Today' },
		{ value: 'this-week', label: 'This Week' },
		{ value: 'blocked', label: 'Blocked' }
	];

	function handleFilterClick(filter: QuickFilter) {
		onFilterChange?.(filter);
	}
</script>

<div class="flex items-center gap-2 px-6 py-3 border-b border-gray-100 bg-gray-50/50 overflow-x-auto">
	{#each filters as filter}
		{@const count = counts[filter.value]}
		<button
			onclick={() => handleFilterClick(filter.value)}
			class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg whitespace-nowrap transition-colors
				{activeFilter === filter.value
					? 'bg-gray-900 text-white font-medium'
					: 'text-gray-600 hover:bg-gray-100'}"
		>
			{filter.label}
			{#if count !== undefined && count > 0}
				<span class="px-1.5 py-0.5 text-xs rounded-full
					{activeFilter === filter.value ? 'bg-white/20' : 'bg-gray-200'}">
					{count}
				</span>
			{/if}
		</button>
	{/each}
</div>
