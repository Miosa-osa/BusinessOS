<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Download, Star, Share2, Upload, Loader2, Trash2, Play, Code, Workflow, Info, Lock, CheckCircle2, Package } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ManifestViewer from '$lib/components/modules/ManifestViewer.svelte';
	import ShareDialog from '$lib/components/modules/ShareDialog.svelte';
	import { categoryLabels } from '$lib/types/modules';
	import { getCategoryColor } from '$lib/constants/colors';
	import { getModuleIcon } from '$lib/components/modules/moduleIcons';

	let store = $state(customModulesStore);
	let storeState = $state($store);

	$effect(() => {
		storeState = $store;
	});

	let moduleId = $derived($page.params.id ?? '');
	let activeTab = $state<'overview' | 'manifest' | 'versions' | 'settings'>('overview');
	let isShareDialogOpen = $state(false);
	let isInstalled = $state(false);
	let isProcessing = $state(false);

	// Using shared getCategoryColor from $lib/constants/colors

	const actionTypeIcons: Record<string, typeof Play> = {
		function: Code,
		api: Workflow,
		workflow: Play,
	};

	function fmtNum(n: number): string {
		if (n >= 1000) return (n / 1000).toFixed(n >= 10000 ? 0 : 1) + 'K';
		return String(n);
	}

	function fmtDate(d: string): string {
		return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	onMount(async () => {
		await store.loadModule(moduleId);
		await store.loadVersions(moduleId);
	});

	async function handleInstall() {
		isProcessing = true;
		const success = await store.installModule(moduleId);
		if (success) isInstalled = true;
		isProcessing = false;
	}

	async function handleUninstall() {
		if (!confirm('Are you sure you want to uninstall this module?')) return;
		isProcessing = true;
		const success = await store.uninstallModule(moduleId);
		if (success) isInstalled = false;
		isProcessing = false;
	}

	async function handleExport() {
		const blob = await store.exportModule(moduleId);
		if (blob) {
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `${storeState.currentModule?.name || 'module'}.json`;
			a.click();
			URL.revokeObjectURL(url);
		}
	}

	async function handleShare(data: Parameters<typeof store.shareModule>[1]) {
		await store.shareModule(moduleId, data);
	}

	async function handleDelete() {
		if (!confirm('Are you sure you want to delete this module? This action cannot be undone.')) return;
		const success = await store.deleteModule(moduleId);
		if (success) goto('/modules');
	}
</script>

<div class="md-page">
	{#if storeState.loading}
		<div class="md-center">
			<Loader2 class="md-spinner" />
			<p class="md-muted">Loading module...</p>
		</div>
	{:else if storeState.error || !storeState.currentModule}
		<div class="md-center">
			<div class="md-error-orb">
				<svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</div>
			<p class="md-text-primary">Failed to load module</p>
			<p class="md-muted">{storeState.error || 'Module not found'}</p>
			<button onclick={() => goto('/modules')} class="btn-pill btn-pill-ghost" aria-label="Back to Modules">
				Back to Modules
			</button>
		</div>
	{:else}
		{@const mod = storeState.currentModule}
		{@const catColor = getCategoryColor(mod.category)}
		{@const ModIcon = getModuleIcon(mod.icon, mod.category)}

		<!-- Top bar -->
		<div class="md-topbar">
			<button onclick={() => goto('/modules')} class="md-back" aria-label="Back to Modules">
				<ArrowLeft class="w-4 h-4" />
				<span>Modules</span>
			</button>
			<div class="md-topbar__actions">
				<button onclick={() => isShareDialogOpen = true} class="md-action-btn" aria-label="Share module">
					<Share2 class="w-4 h-4" />
					<span>Share</span>
				</button>
				<button onclick={handleExport} class="md-action-btn" aria-label="Export module">
					<Upload class="w-4 h-4" />
					<span>Export</span>
				</button>
				<button onclick={handleDelete} class="md-action-btn md-action-btn--danger" aria-label="Delete module">
					<Trash2 class="w-4 h-4" />
				</button>
			</div>
		</div>

		<!-- Hero header -->
		<div class="md-hero">
			<div class="md-hero__icon-wrap">
				<div class="md-hero__icon" style="background: {catColor}">
					<ModIcon class="w-7 h-7" />
				</div>
			</div>
			<div class="md-hero__info">
				<div class="md-hero__title-row">
					<h1 class="md-hero__name">{mod.name}</h1>
					<span class="md-badge md-badge--vis md-badge--{mod.visibility}">{mod.visibility}</span>
				</div>
				<div class="md-hero__meta">
					<span class="md-badge md-badge--cat" style="background: color-mix(in srgb, {catColor} 10%, transparent); color: {catColor}; border-color: color-mix(in srgb, {catColor} 20%, transparent)">
						{categoryLabels[mod.category]}
					</span>
					<span class="md-hero__sep">v{mod.version}</span>
					{#if mod.creator_name}
						<span class="md-hero__sep">by {mod.creator_name}</span>
					{/if}
				</div>
				<p class="md-hero__desc">{mod.description}</p>
			</div>
			<div class="md-hero__cta">
				{#if isInstalled}
					<button onclick={handleUninstall} disabled={isProcessing} class="md-installed-btn" aria-label="Uninstall module">
						<CheckCircle2 class="w-4 h-4" />
						{isProcessing ? 'Removing...' : 'Installed'}
					</button>
				{:else}
					<button onclick={handleInstall} disabled={isProcessing} class="btn-pill btn-pill-primary" aria-label="Install module">
						<Download class="w-4 h-4" />
						{isProcessing ? 'Installing...' : 'Install'}
					</button>
				{/if}
			</div>
		</div>

		<!-- Stats bar -->
		<div class="md-stats-bar">
			<div class="md-stat-item">
				<Download class="w-4 h-4" />
				<span class="md-stat-item__val">{fmtNum(mod.install_count)}</span>
				<span class="md-stat-item__label">installs</span>
			</div>
			<div class="md-stat-item">
				<Star class="w-4 h-4" />
				<span class="md-stat-item__val">{fmtNum(mod.star_count)}</span>
				<span class="md-stat-item__label">stars</span>
			</div>
			<div class="md-stat-item">
				<Info class="w-4 h-4" />
				<span class="md-stat-item__val">{mod.manifest.actions.length}</span>
				<span class="md-stat-item__label">actions</span>
			</div>
			{#if mod.manifest.dependencies && mod.manifest.dependencies.length > 0}
				<div class="md-stat-item">
					<Package class="w-4 h-4" />
					<span class="md-stat-item__val">{mod.manifest.dependencies.length}</span>
					<span class="md-stat-item__label">deps</span>
				</div>
			{/if}
		</div>

		<!-- Tabs -->
		<div class="md-tabs" role="tablist">
			{#each (['overview', 'manifest', 'versions', 'settings'] as const) as tab}
				<button
					class="md-tab {activeTab === tab ? 'md-tab--active' : ''}"
					role="tab"
					aria-selected={activeTab === tab}
					onclick={() => activeTab = tab}
				>{tab.charAt(0).toUpperCase() + tab.slice(1)}</button>
			{/each}
		</div>

		<!-- Content -->
		<div class="md-content">
			{#if activeTab === 'overview'}
				<div class="md-two-col">
					<!-- Main column -->
					<div class="md-main">
						<section class="md-section">
							<h2 class="md-section__title">Actions</h2>
							{#if mod.manifest.actions.length === 0}
								<div class="md-empty-card">
									<p class="md-muted">No actions defined for this module.</p>
								</div>
							{:else}
								<div class="md-action-list">
									{#each mod.manifest.actions as action}
										{@const TypeIcon = actionTypeIcons[action.type] || Code}
										<div class="md-action-card">
											<div class="md-action-card__left">
												<div class="md-action-card__icon" style="background: color-mix(in srgb, {catColor} 8%, transparent); color: {catColor}">
													<TypeIcon class="w-4 h-4" />
												</div>
												<div>
													<h3 class="md-action-card__name">{action.name}</h3>
													<p class="md-action-card__desc">{action.description}</p>
												</div>
											</div>
											<span class="md-action-card__type">{action.type}</span>
										</div>
									{/each}
								</div>
							{/if}
						</section>

						{#if mod.manifest.permissions && mod.manifest.permissions.length > 0}
							<section class="md-section">
								<h2 class="md-section__title">Permissions</h2>
								<div class="md-perm-list">
									{#each mod.manifest.permissions as perm}
										<div class="md-perm-item">
											<Lock class="w-3.5 h-3.5" />
											<span>{perm}</span>
										</div>
									{/each}
								</div>
							</section>
						{/if}
					</div>

					<!-- Sidebar -->
					<aside class="md-sidebar">
						<div class="md-sidebar-card">
							<h3 class="md-sidebar-card__title">Details</h3>
							<dl class="md-detail-list">
								<div class="md-detail-list__item">
									<dt>Version</dt>
									<dd>{mod.version}</dd>
								</div>
								<div class="md-detail-list__item">
									<dt>Category</dt>
									<dd>{categoryLabels[mod.category]}</dd>
								</div>
								<div class="md-detail-list__item">
									<dt>Visibility</dt>
									<dd class="capitalize">{mod.visibility}</dd>
								</div>
								<div class="md-detail-list__item">
									<dt>Status</dt>
									<dd>
										<span class="md-status-dot {mod.is_active ? 'md-status-dot--active' : 'md-status-dot--inactive'}"></span>
										{mod.is_active ? 'Active' : 'Inactive'}
									</dd>
								</div>
								{#if mod.creator_name}
									<div class="md-detail-list__item">
										<dt>Author</dt>
										<dd>{mod.creator_name}</dd>
									</div>
								{/if}
								<div class="md-detail-list__item">
									<dt>Updated</dt>
									<dd>{fmtDate(mod.updated_at)}</dd>
								</div>
								<div class="md-detail-list__item">
									<dt>Created</dt>
									<dd>{fmtDate(mod.created_at)}</dd>
								</div>
							</dl>
						</div>

						{#if mod.manifest.dependencies && mod.manifest.dependencies.length > 0}
							<div class="md-sidebar-card">
								<h3 class="md-sidebar-card__title">Dependencies</h3>
								<div class="md-dep-list">
									{#each mod.manifest.dependencies as dep}
										<span class="md-dep-chip">{dep}</span>
									{/each}
								</div>
							</div>
						{/if}
					</aside>
				</div>

			{:else if activeTab === 'manifest'}
				<section class="md-section md-section--wide">
					<ManifestViewer manifest={mod.manifest} />
				</section>

			{:else if activeTab === 'versions'}
				<section class="md-section md-section--wide">
					{#if storeState.versions.length === 0}
						<div class="md-empty-card">
							<p class="md-muted">No version history available.</p>
						</div>
					{:else}
						<div class="md-version-timeline">
							{#each storeState.versions as version, i}
								<div class="md-version-item">
									<div class="md-version-item__dot {i === 0 ? 'md-version-item__dot--latest' : ''}"></div>
									<div class="md-version-item__content">
										<div class="md-version-item__header">
											<span class="md-version-item__tag">v{version.version}</span>
											<span class="md-version-item__date">{fmtDate(version.created_at)}</span>
											{#if i === 0}
												<span class="md-badge md-badge--latest">Latest</span>
											{/if}
										</div>
										{#if version.changelog}
											<p class="md-version-item__log">{version.changelog}</p>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</section>

			{:else if activeTab === 'settings'}
				<section class="md-section md-section--wide">
					<div class="md-settings-card">
						<label class="md-setting-row">
							<div>
								<p class="md-setting-row__title">Active</p>
								<p class="md-setting-row__desc">Enable or disable this module in your workspace</p>
							</div>
							<input type="checkbox" checked={mod.is_active} class="md-toggle" aria-label="Toggle module active state" />
						</label>
					</div>
					<div class="md-settings-card md-settings-card--danger">
						<div class="md-setting-row">
							<div>
								<p class="md-setting-row__title md-setting-row__title--danger">Delete Module</p>
								<p class="md-setting-row__desc">Permanently remove this module and all its data. This cannot be undone.</p>
							</div>
							<button onclick={handleDelete} class="md-danger-btn" aria-label="Delete module">
								<Trash2 class="w-4 h-4" />
								Delete
							</button>
						</div>
					</div>
				</section>
			{/if}
		</div>

		<ShareDialog
			{moduleId}
			moduleName={mod.name}
			isOpen={isShareDialogOpen}
			onClose={() => isShareDialogOpen = false}
			onShare={handleShare}
		/>
	{/if}
</div>

<style>
	/* ══════════════════════════════════════════════════════════════ */
	/*  MODULE DETAIL v2 — md- prefix, Foundation tokens            */
	/* ══════════════════════════════════════════════════════════════ */
	.md-page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--dbg);
	}

	/* Center states */
	.md-center {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		gap: 8px;
		text-align: center;
	}
	.md-center :global(.md-spinner) {
		width: 28px; height: 28px;
		color: var(--dt3);
		animation: md-spin 1s linear infinite;
	}
	@keyframes md-spin { to { transform: rotate(360deg); } }
	.md-text-primary { font-size: 14px; font-weight: 500; color: var(--dt); }
	.md-muted { font-size: 13px; color: var(--dt3); }
	.md-error-orb {
		width: 52px; height: 52px; border-radius: 50%;
		background: var(--bos-status-error-bg); color: var(--bos-status-error);
		display: flex; align-items: center; justify-content: center;
		margin-bottom: 4px;
	}

	/* Top bar */
	.md-topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 32px;
		border-bottom: 1px solid var(--dbd2);
		flex-shrink: 0;
	}
	.md-back {
		display: inline-flex; align-items: center; gap: 6px;
		font-size: 13px; color: var(--dt3);
		background: none; border: none; cursor: pointer; padding: 0;
		transition: color .15s;
	}
	.md-back:hover { color: var(--dt); }
	.md-topbar__actions {
		display: flex; align-items: center; gap: 4px;
	}
	.md-action-btn {
		display: inline-flex; align-items: center; gap: 6px;
		padding: 6px 12px; border-radius: 8px;
		font-size: 12px; font-weight: 500;
		border: 1px solid var(--dbd);
		background: transparent; color: var(--dt2);
		cursor: pointer; transition: all .15s;
	}
	.md-action-btn:hover {
		background: var(--dbg2);
		border-color: var(--dt3);
		color: var(--dt);
	}
	.md-action-btn--danger {
		border-color: color-mix(in srgb, var(--bos-status-error) 25%, transparent);
		color: var(--bos-status-error);
	}
	.md-action-btn--danger:hover {
		background: var(--bos-status-error-bg);
		border-color: color-mix(in srgb, var(--bos-status-error) 40%, transparent);
		color: var(--bos-status-error-text);
	}

	/* Hero */
	.md-hero {
		display: flex;
		align-items: flex-start;
		gap: 18px;
		padding: 24px 32px 20px;
		flex-shrink: 0;
	}
	.md-hero__icon-wrap { flex-shrink: 0; }
	.md-hero__icon {
		width: 56px; height: 56px; border-radius: 16px;
		display: flex; align-items: center; justify-content: center;
		color: var(--bos-surface-on-color);
	}
	.md-hero__icon--fallback {
		background: var(--dbg3); color: var(--dt3);
	}
	.md-hero__info { flex: 1; min-width: 0; }
	.md-hero__title-row {
		display: flex; align-items: center; gap: 10px; margin-bottom: 6px;
	}
	.md-hero__name {
		font-size: 20px; font-weight: 700; color: var(--dt);
		white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
	}
	.md-hero__meta {
		display: flex; align-items: center; flex-wrap: wrap; gap: 8px;
		font-size: 12px; color: var(--dt3); margin-bottom: 8px;
	}
	.md-hero__sep { color: var(--dt3); }
	.md-hero__desc {
		font-size: 13px; color: var(--dt2); line-height: 1.6;
		max-width: 640px;
	}
	.md-hero__cta { flex-shrink: 0; padding-top: 4px; }

	/* Installed button (green state) */
	.md-installed-btn {
		display: inline-flex; align-items: center; gap: 8px;
		padding: 10px 22px; border-radius: 10px;
		font-size: 13px; font-weight: 600;
		border: none; cursor: pointer;
		background: var(--bos-status-success-bg);
		color: var(--bos-status-success);
		transition: all .15s;
	}
	.md-installed-btn:hover:not(:disabled) {
		background: var(--bos-status-error-bg);
		color: var(--bos-status-error);
	}
	.md-installed-btn:disabled { opacity: 0.5; cursor: not-allowed; }

	/* Stats bar */
	.md-stats-bar {
		display: flex;
		gap: 24px;
		padding: 0 32px 16px;
		flex-shrink: 0;
	}
	.md-stat-item {
		display: inline-flex; align-items: center; gap: 5px;
		font-size: 12px; color: var(--dt3);
	}
	.md-stat-item__val { font-weight: 600; color: var(--dt); }
	.md-stat-item__label { color: var(--dt3); }

	/* Badges */
	.md-badge {
		display: inline-flex; align-items: center;
		padding: 2px 8px; border-radius: 999px;
		font-size: 11px; font-weight: 600;
	}
	.md-badge--cat {
		border: 1px solid transparent;
	}
	.md-badge--vis {
		font-size: 10px; text-transform: uppercase; letter-spacing: 0.04em;
	}
	.md-badge--public { background: var(--bos-status-success-bg); color: var(--bos-status-success); }
	.md-badge--workspace { background: var(--bos-status-info-bg); color: var(--bos-status-info); }
	.md-badge--private { background: var(--bos-status-neutral-bg); color: var(--bos-status-neutral); }
	.md-badge--latest {
		background: var(--bos-status-info-bg);
		color: var(--bos-status-info);
		font-size: 10px;
	}

	/* Tabs */
	.md-tabs {
		display: flex; gap: 0;
		padding: 0 32px;
		border-bottom: 1px solid var(--dbd);
		flex-shrink: 0;
		position: sticky;
		top: 0;
		z-index: 10;
		background: var(--dbg);
	}
	.md-tab {
		padding: 10px 18px; border: none; background: transparent;
		color: var(--dt3); font-size: 13px; font-weight: 500;
		cursor: pointer; border-bottom: 2px solid transparent;
		margin-bottom: -1px; transition: all .15s; white-space: nowrap;
	}
	.md-tab:hover { color: var(--dt); }
	.md-tab--active {
		color: var(--dt);
		border-bottom-color: var(--dt);
	}

	/* Content */
	.md-content {
		flex: 1; overflow-y: auto;
		padding: 24px 32px 40px;
	}

	/* Two-column layout */
	.md-two-col {
		display: grid;
		grid-template-columns: 1fr 280px;
		gap: 32px;
		align-items: flex-start;
	}
	@media (max-width: 900px) {
		.md-two-col { grid-template-columns: 1fr; }
	}
	.md-main { min-width: 0; }

	/* Sections */
	.md-section { margin-bottom: 28px; }
	.md-section--wide { max-width: 800px; }
	.md-section__title {
		font-size: 14px; font-weight: 600; color: var(--dt);
		margin-bottom: 12px; text-transform: uppercase; letter-spacing: 0.03em;
	}

	/* Action cards */
	.md-action-list { display: flex; flex-direction: column; gap: 8px; }
	.md-action-card {
		display: flex; align-items: center; justify-content: space-between;
		padding: 12px 14px; border-radius: 10px;
		border: 1px solid var(--dbd);
		background: var(--dbg);
		transition: border-color .15s, background .15s;
	}
	.md-action-card:hover {
		border-color: var(--dbd2);
		background: var(--dbg2);
	}
	.md-action-card__left {
		display: flex; align-items: center; gap: 12px; min-width: 0;
	}
	.md-action-card__icon {
		width: 34px; height: 34px; border-radius: 8px;
		display: flex; align-items: center; justify-content: center;
		flex-shrink: 0;
	}
	.md-action-card__name {
		font-size: 13px; font-weight: 600; color: var(--dt);
	}
	.md-action-card__desc {
		font-size: 12px; color: var(--dt3); margin-top: 2px;
	}
	.md-action-card__type {
		font-size: 11px; padding: 3px 10px; border-radius: 999px;
		background: var(--dbg2); color: var(--dt3);
		font-weight: 500; flex-shrink: 0;
	}

	/* Permissions */
	.md-perm-list { display: flex; flex-wrap: wrap; gap: 6px; }
	.md-perm-item {
		display: inline-flex; align-items: center; gap: 5px;
		padding: 5px 10px; border-radius: 8px;
		font-size: 12px; color: var(--dt2);
		border: 1px solid var(--dbd);
		background: var(--dbg);
	}

	/* Sidebar */
	.md-sidebar { display: flex; flex-direction: column; gap: 16px; }
	.md-sidebar-card {
		padding: 16px; border-radius: 12px;
		border: 1px solid var(--dbd);
		background: var(--dbg);
	}
	.md-sidebar-card__title {
		font-size: 12px; font-weight: 600; color: var(--dt);
		text-transform: uppercase; letter-spacing: 0.04em;
		margin-bottom: 12px;
	}

	/* Detail list */
	.md-detail-list { display: flex; flex-direction: column; gap: 0; }
	.md-detail-list__item {
		display: flex; justify-content: space-between; align-items: center;
		padding: 7px 0;
		border-bottom: 1px solid var(--dbd2);
	}
	.md-detail-list__item:last-child { border-bottom: none; }
	.md-detail-list__item dt {
		font-size: 12px; color: var(--dt3);
	}
	.md-detail-list__item dd {
		font-size: 12px; font-weight: 500; color: var(--dt);
		display: flex; align-items: center; gap: 5px;
	}
	.capitalize { text-transform: capitalize; }
	.md-status-dot {
		width: 7px; height: 7px; border-radius: 50%;
	}
	.md-status-dot--active { background: var(--bos-status-success); }
	.md-status-dot--inactive { background: var(--bos-status-neutral); }

	/* Dependencies */
	.md-dep-list { display: flex; flex-wrap: wrap; gap: 6px; }
	.md-dep-chip {
		padding: 4px 10px; border-radius: 6px; font-size: 11px;
		font-weight: 500; font-family: monospace;
		background: var(--dbg2); color: var(--dt2);
		border: 1px solid var(--dbd);
	}

	/* Empty card */
	.md-empty-card {
		padding: 24px; border-radius: 10px; text-align: center;
		border: 1px dashed var(--dbd);
		background: var(--dbg2);
	}

	/* Version timeline */
	.md-version-timeline {
		display: flex; flex-direction: column; gap: 0;
		padding-left: 16px;
		border-left: 2px solid var(--dbd);
	}
	.md-version-item {
		display: flex; align-items: flex-start; gap: 14px;
		padding: 12px 0;
		position: relative;
	}
	.md-version-item__dot {
		width: 10px; height: 10px; border-radius: 50%;
		background: var(--dbd); border: 2px solid var(--dbg);
		flex-shrink: 0; margin-top: 4px; margin-left: -21px;
	}
	.md-version-item__dot--latest {
		background: var(--bos-status-info);
	}
	.md-version-item__content { flex: 1; }
	.md-version-item__header {
		display: flex; align-items: center; gap: 8px; margin-bottom: 4px;
	}
	.md-version-item__tag {
		font-size: 13px; font-weight: 600; color: var(--dt);
		font-family: monospace;
	}
	.md-version-item__date {
		font-size: 12px; color: var(--dt3);
	}
	.md-version-item__log {
		font-size: 13px; color: var(--dt2); line-height: 1.5;
	}

	/* Settings */
	.md-settings-card {
		padding: 16px 18px; border-radius: 12px;
		border: 1px solid var(--dbd);
		background: var(--dbg);
		margin-bottom: 12px;
	}
	.md-settings-card--danger {
		border-color: color-mix(in srgb, var(--bos-status-error) 20%, transparent);
	}
	.md-setting-row {
		display: flex; align-items: center; justify-content: space-between;
		cursor: pointer; width: 100%; gap: 16px;
	}
	.md-setting-row__title {
		font-size: 13px; font-weight: 600; color: var(--dt);
	}
	.md-setting-row__title--danger { color: var(--bos-status-error); }
	.md-setting-row__desc {
		font-size: 12px; color: var(--dt3); margin-top: 2px;
	}

	/* Toggle checkbox */
	.md-toggle {
		width: 40px; height: 22px; appearance: none; -webkit-appearance: none;
		background: var(--dbd); border-radius: 999px;
		position: relative; cursor: pointer; transition: background .2s;
		flex-shrink: 0;
	}
	.md-toggle::after {
		content: ''; position: absolute;
		top: 2px; left: 2px; width: 18px; height: 18px;
		border-radius: 50%; background: var(--dbg);
		transition: transform .2s;
	}
	.md-toggle:checked { background: var(--bos-status-success); }
	.md-toggle:checked::after { transform: translateX(18px); }

	/* Danger button */
	.md-danger-btn {
		display: inline-flex; align-items: center; gap: 6px;
		padding: 8px 16px; border-radius: 8px;
		font-size: 12px; font-weight: 600;
		border: 1px solid color-mix(in srgb, var(--bos-status-error) 30%, transparent);
		background: var(--bos-status-error-bg);
		color: var(--bos-status-error); cursor: pointer;
		transition: all .15s; flex-shrink: 0;
	}
	.md-danger-btn:hover {
		background: color-mix(in srgb, var(--bos-status-error) 15%, transparent);
		border-color: color-mix(in srgb, var(--bos-status-error) 50%, transparent);
	}
</style>
