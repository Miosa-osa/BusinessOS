<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Plus, Loader2, LayoutGrid, List } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ModuleCard from '$lib/components/modules/ModuleCard.svelte';
	import ModuleFilters from '$lib/components/modules/ModuleFilters.svelte';

	const BUILTIN_MODULES = [
		{ id: 'dashboard', name: 'Dashboard', description: 'Overview and quick access', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6', href: '/dashboard', color: '#3b82f6' },
		{ id: 'chat', name: 'Chat', description: 'AI conversation interface', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z', href: '/chat', color: '#8b5cf6' },
		{ id: 'tasks', name: 'Tasks', description: 'Task management', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4', href: '/tasks', color: '#22c55e' },
		{ id: 'projects', name: 'Projects', description: 'Project tracking', icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', href: '/projects', color: '#a855f7' },
		{ id: 'daily', name: 'Daily Log', description: 'Daily rhythm and schedule', icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z', href: '/daily', color: '#f59e0b' },
		{ id: 'communication', name: 'Communication', description: 'Calendar and messaging', icon: 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z', href: '/communication/calendar', color: '#06b6d4' },
		{ id: 'team', name: 'Team', description: 'Team management', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z', href: '/team', color: '#ef4444' },
		{ id: 'clients', name: 'Clients', description: 'Client management', icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4', href: '/clients', color: '#10b981' },
		{ id: 'tables', name: 'Tables', description: 'Data tables', icon: 'M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z', href: '/tables', color: '#6366f1' },
		{ id: 'pages', name: 'Pages', description: 'Knowledge base', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z', href: '/pages', color: '#ec4899' },
		{ id: 'agents', name: 'Agents', description: 'AI agent management', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z', href: '/agents', color: '#f97316' },
		{ id: 'nodes', name: 'Nodes', description: 'OptimalOS nodes', icon: 'M13 10V3L4 14h7v7l9-11h-7z', href: '/nodes', color: '#14b8a6' },
		{ id: 'terminal', name: 'Terminal', description: 'Terminal interface', icon: 'M4 17l6-6-6-6m8 14h8', href: '/terminal', color: '#84cc16' },
		{ id: 'code', name: 'Code', description: 'Code editor', icon: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4', href: '/code', color: '#0ea5e9' },
	];

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

	const hasCustomModules = $derived(!storeState.loading && !storeState.error && filteredModules.length > 0);
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

	<!-- Stats Bar + View Toggle (only when custom modules loaded) -->
	{#if hasCustomModules}
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

		<!-- Built-in Modules — always visible -->
		<div class="am-section">
			<h2 class="am-section__title">Built-in Modules</h2>
			<div class="am-builtin-grid">
				{#each BUILTIN_MODULES as mod}
					<button
						class="am-builtin-card"
						style="--am-accent: {mod.color}"
						onclick={() => goto(mod.href)}
						aria-label="Open {mod.name}"
					>
						<span class="am-builtin-card__border"></span>
						<span class="am-builtin-card__icon">
							<svg class="am-builtin-card__svg" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d={mod.icon} />
							</svg>
						</span>
						<span class="am-builtin-card__body">
							<span class="am-builtin-card__name">{mod.name}</span>
							<span class="am-builtin-card__desc">{mod.description}</span>
						</span>
					</button>
				{/each}
			</div>
		</div>

		<!-- Custom Modules — shown when API succeeds -->
		{#if storeState.loading}
			<div class="am-section">
				<h2 class="am-section__title">Custom Modules</h2>
				<div class="am-page__center am-page__center--sm">
					<Loader2 class="am-page__spinner" />
					<p class="am-page__muted">Loading custom modules...</p>
				</div>
			</div>
		{:else if storeState.error}
			<div class="am-section">
				<h2 class="am-section__title">Custom Modules</h2>
				<div class="am-page__center am-page__center--sm">
					<div class="am-page__error-orb">
						<svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<p class="am-page__text">Could not load custom modules</p>
					<p class="am-page__muted">{storeState.error}</p>
					<button
						onclick={() => store.loadModules()}
						class="btn-pill btn-pill-ghost"
					>
						Try Again
					</button>
				</div>
			</div>
		{:else if filteredModules.length === 0}
			<div class="am-section">
				<h2 class="am-section__title">Custom Modules</h2>
				<div class="am-page__center am-page__center--sm">
					<div class="am-page__empty-orb">
						<svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
						</svg>
					</div>
					<p class="am-page__text">No custom modules yet</p>
					<p class="am-page__muted">
						{#if storeState.filters.search || storeState.filters.category}
							Try adjusting your filters or create a new module.
						{:else}
							Extend BusinessOS with your own custom modules.
						{/if}
					</p>
					<button onclick={handleCreateModule} class="btn-cta">
						Create Module
					</button>
				</div>
			</div>
		{:else}
			<div class="am-section">
				<h2 class="am-section__title">Custom Modules</h2>
				{#if viewMode === 'list'}
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
				<div class="am-page__summary">
					<p>Showing {filteredModules.length} of {storeState.total} modules</p>
				</div>
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

	/* Section wrapper */
	.am-section {
		margin-bottom: 32px;
	}
	.am-section__title {
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		color: var(--dt3);
		margin-bottom: 12px;
	}

	/* Built-in module cards */
	.am-builtin-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 8px;
	}

	.am-builtin-card {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 8px;
		border: 1px solid var(--dbd2);
		background: var(--dbg2);
		cursor: pointer;
		text-align: left;
		transition: background 0.12s, border-color 0.12s, box-shadow 0.12s;
		overflow: hidden;
	}

	.am-builtin-card:hover {
		background: var(--dbg3);
		border-color: var(--dbd);
		box-shadow: 0 1px 6px rgba(0,0,0,0.06);
	}

	/* Colored left accent border */
	.am-builtin-card__border {
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		width: 3px;
		border-radius: 8px 0 0 8px;
		background: var(--am-accent, #3b82f6);
	}

	.am-builtin-card__icon {
		flex-shrink: 0;
		width: 32px;
		height: 32px;
		border-radius: 6px;
		background: color-mix(in srgb, var(--am-accent, #3b82f6) 12%, transparent);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--am-accent, #3b82f6);
	}

	.am-builtin-card__svg {
		width: 16px;
		height: 16px;
	}

	.am-builtin-card__body {
		display: flex;
		flex-direction: column;
		gap: 1px;
		min-width: 0;
	}

	.am-builtin-card__name {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.am-builtin-card__desc {
		font-size: 11px;
		color: var(--dt3);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	/* Smaller centered state for the custom modules section */
	.am-page__center--sm {
		min-height: 160px;
	}
</style>
