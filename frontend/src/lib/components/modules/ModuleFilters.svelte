<script lang="ts">
	import { Search } from 'lucide-svelte';
	import type { ModuleCategory, ModuleFilters } from '$lib/types/modules';
	import { categoryLabels } from '$lib/types/modules';

	interface Props {
		filters: ModuleFilters;
		onFiltersChange: (filters: Partial<ModuleFilters>) => void;
	}

	let { filters, onFiltersChange }: Props = $props();

	let searchInput = $state(filters.search);

	// Debounce search
	let debounceTimer: number;
	function handleSearchInput(value: string) {
		searchInput = value;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			onFiltersChange({ search: value });
		}, 300) as unknown as number;
	}

	const categories: (ModuleCategory | null)[] = [
		null,
		'productivity',
		'communication',
		'finance',
		'analytics',
		'automation',
		'integration',
		'utilities',
		'custom'
	];

	const categoryHexColors: Record<string, string> = {
		productivity: '#3b82f6',
		communication: '#a855f7',
		finance: '#10b981',
		analytics: '#f97316',
		automation: '#ec4899',
		integration: '#6366f1',
		utilities: '#6b7280',
		custom: '#06b6d4',
	};

	const sortOptions: Array<{ value: 'popular' | 'newest' | 'name' | 'installs'; label: string }> = [
		{ value: 'popular', label: 'Popular' },
		{ value: 'newest', label: 'Newest' },
		{ value: 'name', label: 'Name' },
		{ value: 'installs', label: 'Most Installed' }
	];
</script>

<!-- Search + Sort row -->
<div class="am-search-row">
	<div class="am-search-wrap">
		<Search class="am-search-icon" />
		<input
			type="text"
			placeholder="Search modules…"
			value={searchInput}
			oninput={(e) => handleSearchInput(e.currentTarget.value)}
			class="am-search-input"
			aria-label="Search modules"
		/>
	</div>
	<div class="am-sort-wrap">
		<span class="am-sort-label">Sort:</span>
		{#each sortOptions as option}
			<button
				class="am-sort-chip {filters.sort === option.value ? 'am-sort-chip--active' : ''}"
				onclick={() => onFiltersChange({ sort: option.value })}
				aria-pressed={filters.sort === option.value}
				aria-label="Sort by {option.label}"
			>{option.label}</button>
		{/each}
	</div>
</div>

<!-- Category filter chips -->
<div class="am-cat-row">
	{#each categories as cat}
		{@const isActive = filters.category === cat}
		{@const label = cat ? categoryLabels[cat] : 'All'}
		{@const color = cat ? categoryHexColors[cat] : undefined}
		<button
			class="am-cat-chip {isActive ? 'am-cat-chip--active' : ''}"
			onclick={() => onFiltersChange({ category: cat })}
			style={isActive && color ? `background:${color}16;border-color:${color};color:${color}` : ''}
			aria-pressed={isActive}
			aria-label="Filter by {label}"
		>{label}</button>
	{/each}
</div>

<style>
	/* ══════════════════════════════════════════════════════════════ */
	/*  MODULE FILTERS (am-) — Foundation AppMarketplace Pattern    */
	/* ══════════════════════════════════════════════════════════════ */

	/* Search row */
	.am-search-row {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		margin-bottom: 12px;
	}
	.am-search-wrap {
		position: relative;
		flex: 1;
		min-width: 200px;
	}
	.am-search-wrap :global(.am-search-icon) {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		color: var(--dt3, #888);
		pointer-events: none;
		width: 15px;
		height: 15px;
	}
	.am-search-input {
		width: 100%;
		padding: 9px 12px 9px 34px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
		font-size: 13px;
		outline: none;
		transition: border-color .15s;
	}
	.am-search-input::placeholder {
		color: var(--dt4, #bbb);
	}
	.am-search-input:focus {
		border-color: var(--accent-blue, #3b82f6);
	}

	/* Sort chips */
	.am-sort-wrap {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
	}
	.am-sort-label {
		font-size: 12px;
		color: var(--dt3, #888);
	}
	.am-sort-chip {
		padding: 5px 12px;
		border-radius: 999px;
		border: 1px solid var(--dbd, #e0e0e0);
		background: transparent;
		color: var(--dt2, #555);
		font-size: 12px;
		cursor: pointer;
		transition: all .15s;
	}
	.am-sort-chip:hover {
		border-color: var(--accent-blue, #3b82f6);
		color: var(--accent-blue, #3b82f6);
	}
	.am-sort-chip--active {
		background: var(--accent-blue, #3b82f6);
		border-color: var(--accent-blue, #3b82f6);
		color: #fff;
	}

	/* Category chips */
	.am-cat-row {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 4px;
	}
	.am-cat-chip {
		padding: 6px 14px;
		border-radius: 999px;
		border: 1px solid var(--dbd, #e0e0e0);
		background: transparent;
		color: var(--dt2, #555);
		font-size: 12px;
		font-weight: 500;
		cursor: pointer;
		transition: all .15s;
	}
	.am-cat-chip:hover {
		border-color: var(--accent-blue, #3b82f6);
		color: var(--accent-blue, #3b82f6);
	}
	.am-cat-chip--active {
		background: rgba(59, 130, 246, 0.07);
		border-color: var(--accent-blue, #3b82f6);
		color: var(--accent-blue, #3b82f6);
	}
</style>
