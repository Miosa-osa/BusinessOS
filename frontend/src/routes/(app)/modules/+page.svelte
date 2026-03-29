<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Plus, Loader2, LayoutGrid, List } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ModuleCard from '$lib/components/modules/ModuleCard.svelte';
	import ModuleFilters from '$lib/components/modules/ModuleFilters.svelte';

	let store = $state(customModulesStore);
	let storeState = $state($store);
	let viewMode = $state<'grid' | 'list'>('grid');

	$effect(() => {
		storeState = $store;
	});

	onMount(() => {
		store.loadModules();
	});

	function handleFiltersChange(filters: Parameters<typeof store.setFilters>[0]) {
		store.setFilters(filters);
		store.loadModules();
	}

	function handleModuleClick(moduleId: string) {
		goto(`/modules/${moduleId}`);
	}

	function handleCreateModule() {
		goto('/modules/create');
	}

	// Client-side filtering (fallback if backend doesn't filter)
	const filteredModules = $derived.by(() => {
		let result = storeState.modules;
		const { category, search } = storeState.filters;
		if (category) {
			result = result.filter(m => m.category === category);
		}
		if (search) {
			const q = search.toLowerCase();
			result = result.filter(m =>
				m.name.toLowerCase().includes(q) ||
				m.description.toLowerCase().includes(q)
			);
		}
		return result;
	});

	// Stats (based on all modules, not filtered)
	const moduleCount = $derived(filteredModules.length);
	const totalInstalls = $derived(storeState.modules.reduce((sum, m) => sum + m.install_count, 0));
	const activeCount = $derived(storeState.modules.filter(m => m.is_active).length);
	const categoryCount = $derived(new Set(storeState.modules.map(m => m.category)).size);
</script>

<div class="am-page">
	<!-- Header -->
	<div class="am-page__header">
		<div class="am-page__header-top">
			<div>
				<h1 class="am-page__title">Custom Modules</h1>
				<p class="am-page__subtitle">Browse and manage custom modules for your workspace</p>
			</div>
			<button
				onclick={handleCreateModule}
				class="btn-cta"
				aria-label="Create Module"
			>
				<Plus class="w-4 h-4" />
				<span>Create Module</span>
			</button>
		</div>

		<!-- Filters -->
		<ModuleFilters
			filters={storeState.filters}
			onFiltersChange={handleFiltersChange}
		/>
	</div>

	<!-- Stats Bar + View Toggle -->
	{#if !storeState.loading && !storeState.error && filteredModules.length > 0}
		<div class="am-stats-bar">
			<div class="am-stats-bar__left">
				<span class="am-stats-bar__item"><strong>{moduleCount}</strong> modules</span>
				<span class="am-stats-bar__sep"></span>
				<span class="am-stats-bar__item"><strong>{activeCount}</strong> active</span>
				<span class="am-stats-bar__sep"></span>
				<span class="am-stats-bar__item"><strong>{categoryCount}</strong> categories</span>
				<span class="am-stats-bar__sep"></span>
				<span class="am-stats-bar__item"><strong>{totalInstalls.toLocaleString()}</strong> total installs</span>
			</div>
			<div class="am-stats-bar__right">
				<button
					class="am-view-btn"
					class:am-view-btn--active={viewMode === 'grid'}
					onclick={() => viewMode = 'grid'}
					aria-label="Grid view"
				>
					<LayoutGrid class="w-4 h-4" />
				</button>
				<button
					class="am-view-btn"
					class:am-view-btn--active={viewMode === 'list'}
					onclick={() => viewMode = 'list'}
					aria-label="List view"
				>
					<List class="w-4 h-4" />
				</button>
			</div>
		</div>
	{/if}

	<!-- Content -->
	<div class="am-page__content">
		{#if storeState.loading}
			<div class="am-page__center">
				<Loader2 class="am-page__spinner" />
				<p class="am-page__muted">Loading modules...</p>
			</div>
		{:else if storeState.error}
			<div class="am-page__center">
				<div class="am-page__error-orb">
					<svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<p class="am-page__text">Failed to load modules</p>
				<p class="am-page__muted">{storeState.error}</p>
				<button
					onclick={() => store.loadModules()}
					class="btn-pill btn-pill-ghost"
				>
					Try Again
				</button>
			</div>
		{:else if filteredModules.length === 0}
			<div class="am-page__center">
				<div class="am-page__empty-orb">
					<svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
					</svg>
				</div>
				<p class="am-page__text">No modules found</p>
				<p class="am-page__muted">
					{#if storeState.filters.search || storeState.filters.category}
						Try adjusting your filters or create a new module.
					{:else}
						Get started by creating your first custom module.
					{/if}
				</p>
				<button
					onclick={handleCreateModule}
					class="btn-cta"
				>
					Create Module
				</button>
			</div>
		{:else if viewMode === 'list'}
			<!-- List View -->
			<div class="am-list">
				{#each filteredModules as module}
					<ModuleCard
						{module}
						compact={true}
						onClick={() => handleModuleClick(module.id)}
					/>
				{/each}
			</div>
		{:else}
			<!-- Grid View -->
			<div class="am-grid">
				{#each filteredModules as module}
					<ModuleCard
						{module}
						onClick={() => handleModuleClick(module.id)}
					/>
				{/each}
			</div>
		{/if}

		<!-- Results Summary -->
		{#if !storeState.loading && !storeState.error && storeState.modules.length > 0}
			<div class="am-page__summary">
				<p>Showing {filteredModules.length} of {storeState.total} modules</p>
			</div>
		{/if}
	</div>
</div>

<style>
	/* ══════════════════════════════════════════════════════════════ */
	/*  MODULES PAGE v2 (am-) — Foundation Design Tokens            */
	/* ══════════════════════════════════════════════════════════════ */
	.am-page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--dbg);
	}

	/* Header */
	.am-page__header {
		flex-shrink: 0;
		padding: 24px 32px 16px;
		border-bottom: 1px solid var(--dbd2);
		background: var(--dbg);
	}
	.am-page__header-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 16px;
	}
	.am-page__title {
		font-size: 20px;
		font-weight: 700;
		color: var(--dt);
		line-height: 1.3;
	}
	.am-page__subtitle {
		font-size: 13px;
		color: var(--dt3);
		margin-top: 2px;
	}

	/* Stats Bar */
	.am-stats-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 32px;
		border-bottom: 1px solid var(--dbd2);
		flex-shrink: 0;
	}
	.am-stats-bar__left {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.am-stats-bar__item {
		font-size: 11px;
		color: var(--dt3);
	}
	.am-stats-bar__item strong {
		color: var(--dt);
		font-weight: 600;
	}
	.am-stats-bar__sep {
		width: 3px;
		height: 3px;
		border-radius: 50%;
		background: var(--dbd);
	}
	.am-stats-bar__right {
		display: flex;
		align-items: center;
		gap: 2px;
	}
	.am-view-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		border-radius: 6px;
		border: none;
		background: transparent;
		color: var(--dt3);
		cursor: pointer;
		transition: all 0.15s;
	}
	.am-view-btn:hover {
		background: var(--dbg2);
		color: var(--dt);
	}
	.am-view-btn--active {
		background: var(--dbg2);
		color: var(--dt);
	}

	/* Content */
	.am-page__content {
		flex: 1;
		overflow-y: auto;
		padding: 24px 32px 40px;
	}

	/* Grid */
	.am-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: 16px;
	}

	/* List */
	.am-list {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	/* Center states */
	.am-page__center {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		gap: 8px;
		text-align: center;
	}
	.am-page__center :global(.am-page__spinner) {
		width: 28px;
		height: 28px;
		color: var(--dt3);
		animation: am-spin 1s linear infinite;
	}
	@keyframes am-spin {
		to { transform: rotate(360deg); }
	}
	.am-page__text {
		font-size: 14px;
		font-weight: 500;
		color: var(--dt);
	}
	.am-page__muted {
		font-size: 13px;
		color: var(--dt3);
	}
	.am-page__error-orb {
		width: 52px;
		height: 52px;
		border-radius: 50%;
		background: var(--bos-status-error-bg);
		color: var(--bos-status-error);
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 4px;
	}
	.am-page__empty-orb {
		width: 52px;
		height: 52px;
		border-radius: 50%;
		background: var(--dbg2);
		color: var(--dt3);
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 4px;
	}
	.am-page__summary {
		margin-top: 24px;
		text-align: center;
		font-size: 12px;
		color: var(--dt4);
	}
</style>
