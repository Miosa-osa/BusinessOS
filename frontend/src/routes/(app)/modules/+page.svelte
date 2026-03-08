<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Plus, Loader2 } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ModuleCard from '$lib/components/modules/ModuleCard.svelte';
	import ModuleFilters from '$lib/components/modules/ModuleFilters.svelte';

	let store = customModulesStore;
	let storeState = $derived($store);

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
</script>

<div class="am-page">
	<!-- Header -->
	<div class="am-page-header">
		<div class="am-page-header__top">
			<div>
				<h1 class="am-page-title">Custom Modules</h1>
				<p class="am-page-subtitle">Browse and manage custom modules for your workspace</p>
			</div>
			<button
				onclick={handleCreateModule}
				class="btn-pill btn-pill-primary"
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

	<!-- Content -->
	<div class="am-page-content">
		{#if storeState.loading}
			<!-- Loading State -->
			<div class="am-page-center">
				<Loader2 class="am-page-spinner" />
				<p class="am-page-muted">Loading modules...</p>
			</div>
		{:else if storeState.error}
			<!-- Error State -->
			<div class="am-page-center">
				<div class="am-error-icon">
					<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<p class="am-page-text">Failed to load modules</p>
				<p class="am-page-muted">{storeState.error}</p>
				<button
					onclick={() => store.loadModules()}
					class="am-retry-btn"
					aria-label="Try again"
				>
					Try Again
				</button>
			</div>
		{:else if storeState.modules.length === 0}
			<!-- Empty State -->
			<div class="am-page-center">
				<div class="am-empty-icon">
					<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
					</svg>
				</div>
				<p class="am-page-text">No modules found</p>
				<p class="am-page-muted">
					{#if storeState.filters.search || storeState.filters.category}
						Try adjusting your filters or create a new module.
					{:else}
						Get started by creating your first custom module.
					{/if}
				</p>
				<button
					onclick={handleCreateModule}
					class="btn-pill btn-pill-primary btn-pill-sm"
					aria-label="Create Module"
				>
					Create Module
				</button>
			</div>
		{:else}
			<!-- Modules Grid -->
			<div class="am-module-grid">
				{#each storeState.modules as module}
					<ModuleCard
						{module}
						onClick={() => handleModuleClick(module.id)}
					/>
				{/each}
			</div>

			<!-- Results Summary -->
			<div class="am-results-summary">
				<p class="am-page-muted">
					Showing {storeState.modules.length} of {storeState.total} modules
				</p>
			</div>
		{/if}
	</div>
</div>

<style>
	/* ══════════════════════════════════════════════════════════════ */
	/*  MODULES PAGE (am-page-) — Foundation Design Tokens          */
	/* ══════════════════════════════════════════════════════════════ */
	.am-page {
		background: var(--dbg, #fff);
		height: 100%;
		display: flex;
		flex-direction: column;
	}
	.am-page-header {
		flex-shrink: 0;
		padding: 24px 32px 16px;
		border-bottom: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg, #fff);
	}
	.am-page-header__top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 20px;
	}
	.am-page-header__top .btn-pill {
		display: inline-flex;
		align-items: center;
		gap: 8px;
	}
	.am-page-title {
		font-size: 22px;
		font-weight: 700;
		color: var(--dt, #111);
	}
	.am-page-subtitle {
		font-size: 13px;
		color: var(--dt2, #555);
		margin-top: 4px;
	}

	/* Content area */
	.am-page-content {
		flex: 1;
		overflow-y: auto;
		padding: 24px 32px;
	}

	/* Module grid */
	.am-module-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 14px;
	}
	@media (max-width: 1200px) {
		.am-module-grid { grid-template-columns: repeat(2, 1fr); }
	}
	@media (max-width: 700px) {
		.am-module-grid { grid-template-columns: 1fr; }
	}

	/* Center states (loading, error, empty) */
	.am-page-center {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		text-align: center;
		gap: 8px;
	}
	.am-page-text {
		font-size: 14px;
		font-weight: 500;
		color: var(--dt, #111);
	}
	.am-page-muted {
		font-size: 13px;
		color: var(--dt3, #888);
	}
	.am-page-center :global(.am-page-spinner) {
		width: 28px;
		height: 28px;
		color: var(--dt3, #888);
		animation: spin 1s linear infinite;
		margin-bottom: 4px;
	}
	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	/* Error state */
	.am-error-icon {
		width: 56px;
		height: 56px;
		border-radius: 50%;
		background: rgba(239, 68, 68, 0.1);
		display: flex;
		align-items: center;
		justify-content: center;
		color: #ef4444;
		margin-bottom: 8px;
	}
	.am-retry-btn {
		margin-top: 8px;
		padding: 7px 20px;
		border-radius: 999px;
		border: 1px solid var(--dbd, #e0e0e0);
		background: transparent;
		color: var(--dt, #111);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all .15s;
	}
	.am-retry-btn:hover {
		border-color: var(--accent-blue, #3b82f6);
		color: var(--accent-blue, #3b82f6);
	}

	/* Empty state */
	.am-empty-icon {
		width: 56px;
		height: 56px;
		border-radius: 50%;
		background: var(--dbg2, #f5f5f5);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--dt3, #888);
		margin-bottom: 8px;
	}

	/* Results summary */
	.am-results-summary {
		margin-top: 24px;
		text-align: center;
	}
</style>
