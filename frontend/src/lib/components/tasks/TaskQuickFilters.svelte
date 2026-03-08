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

<div class="tb-quick-filters flex items-center gap-2 px-6 py-3 overflow-x-auto">
	{#each filters as filter}
		{@const count = counts[filter.value]}
		<button
			onclick={() => handleFilterClick(filter.value)}
			class="btn-pill btn-pill-sm flex items-center gap-1.5 whitespace-nowrap {activeFilter === filter.value
				? 'btn-pill-primary'
				: 'btn-pill-ghost'}"
		>
			{filter.label}
			{#if count !== undefined && count > 0}
				<span class="tb-quick-badge px-1.5 py-0.5 text-xs rounded-full
					{activeFilter === filter.value ? 'tb-quick-badge--active' : ''}">
					{count}
				</span>
			{/if}
		</button>
	{/each}
</div>

<style>
	.tb-quick-filters {
		background: color-mix(in srgb, var(--dbg2, #f5f5f5) 50%, transparent);
		border-bottom: 1px solid var(--dbd2, #f0f0f0);
	}
	.tb-quick-badge {
		background: var(--dbg3, #eee);
	}
	.tb-quick-badge--active {
		background: rgba(255,255,255,0.2);
	}
</style>
