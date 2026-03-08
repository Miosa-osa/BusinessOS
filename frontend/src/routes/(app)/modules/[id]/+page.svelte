<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Download, Star, Share2, Upload, Loader2, Trash2, Package } from 'lucide-svelte';
	import { customModulesStore } from '$lib/stores/customModulesStore';
	import ManifestViewer from '$lib/components/modules/ManifestViewer.svelte';
	import ShareDialog from '$lib/components/modules/ShareDialog.svelte';
	import { categoryLabels } from '$lib/types/modules';

	let store = $state(customModulesStore);
	let storeState = $state($store);

	$effect(() => {
		storeState = $store;
	});

	let moduleId = $derived($page.params.id);
	let activeTab = $state<'overview' | 'manifest' | 'versions' | 'settings'>('overview');
	let isShareDialogOpen = $state(false);
	let isInstalled = $state(false);
	let isProcessing = $state(false);

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

	function fmtNum(n: number): string {
		if (n >= 1000) return (n / 1000).toFixed(n >= 10000 ? 0 : 1) + 'K';
		return String(n);
	}

	onMount(async () => {
		await store.loadModule(moduleId);
		await store.loadVersions(moduleId);
	});

	async function handleInstall() {
		isProcessing = true;
		const success = await store.installModule(moduleId);
		if (success) {
			isInstalled = true;
		}
		isProcessing = false;
	}

	async function handleUninstall() {
		if (!confirm('Are you sure you want to uninstall this module?')) return;
		isProcessing = true;
		const success = await store.uninstallModule(moduleId);
		if (success) {
			isInstalled = false;
		}
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
		if (success) {
			goto('/modules');
		}
	}
</script>

<div class="am-detail-page">
	{#if storeState.loading}
		<!-- Loading State -->
		<div class="am-detail-center">
			<Loader2 class="am-detail-spinner" />
			<p class="am-detail-muted">Loading module...</p>
		</div>
	{:else if storeState.error || !storeState.currentModule}
		<!-- Error State -->
		<div class="am-detail-center">
			<div class="am-detail-error-icon">
				<svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</div>
			<p class="am-detail-text">Failed to load module</p>
			<p class="am-detail-muted">{storeState.error || 'Module not found'}</p>
			<button
				onclick={() => goto('/modules')}
				class="am-icon-btn"
				aria-label="Back to Modules"
			>
				Back to Modules
			</button>
		</div>
	{:else}
		{@const mod = storeState.currentModule}
		{@const catColor = categoryHexColors[mod.category] || '#6366f1'}

		<!-- Module Header -->
		<div class="am-detail-header">
			<!-- Back Button -->
			<button
				onclick={() => goto('/modules')}
				class="am-back-btn"
				aria-label="Back to Modules"
			>
				<ArrowLeft class="w-4 h-4" />
				<span>Back to Modules</span>
			</button>

			<!-- Detail header row -->
			<div class="am-detail">
				<div class="am-detail__header">
					<!-- Icon -->
					{#if mod.icon}
						<div class="am-detail__icon" style="background: {catColor}">
							{mod.icon}
						</div>
					{:else}
						<div class="am-detail__icon am-detail__icon--fallback">
							<Package class="w-8 h-8" />
						</div>
					{/if}

					<!-- Info -->
					<div class="am-detail__info">
						<h1 class="am-detail__name">{mod.name}</h1>
						<div class="am-detail__sub">
							{#if mod.creator_name}
								<span>by {mod.creator_name}</span>
								<span class="am-dot">·</span>
							{/if}
							<span>v{mod.version}</span>
							<span class="am-dot">·</span>
							<span
								class="am-cat-badge"
								style="background: {catColor}16; color: {catColor}"
							>{categoryLabels[mod.category]}</span>
							<span class="am-visibility-badge am-visibility-badge--{mod.visibility}">{mod.visibility}</span>
						</div>
						<p class="am-detail__desc">{mod.description}</p>
						<div class="am-detail__stats">
							<span class="am-stat">
								<Download class="w-3.5 h-3.5" />
								{fmtNum(mod.install_count)} installs
							</span>
							<span class="am-stat">
								<Star class="w-3.5 h-3.5" />
								{fmtNum(mod.star_count)} stars
							</span>
						</div>
					</div>
				</div>

				<!-- Action Buttons -->
				<div class="am-detail__actions">
					{#if isInstalled}
						<button
							onclick={handleUninstall}
							disabled={isProcessing}
							class="am-btn am-btn--ghost"
							aria-label="Uninstall module"
						>
							{isProcessing ? 'Uninstalling...' : 'Uninstall'}
						</button>
					{:else}
						<button
							onclick={handleInstall}
							disabled={isProcessing}
							class="am-btn am-btn--install"
							aria-label="Install module"
						>
							{isProcessing ? 'Installing...' : '+ Install Module'}
						</button>
					{/if}
					<button
						onclick={() => isShareDialogOpen = true}
						class="am-icon-btn"
						aria-label="Share module"
					>
						<Share2 class="w-4 h-4" />
						Share
					</button>
					<button
						onclick={handleExport}
						class="am-icon-btn"
						aria-label="Export module"
					>
						<Upload class="w-4 h-4" />
						Export
					</button>
					<button
						onclick={handleDelete}
						class="am-icon-btn am-icon-btn--danger"
						aria-label="Delete module"
					>
						<Trash2 class="w-4 h-4" />
						Delete
					</button>
				</div>
			</div>

			<!-- Tabs -->
			<div class="am-tabs" role="tablist">
				{#each (['overview', 'manifest', 'versions', 'settings'] as const) as tab}
					<button
						class="am-tab {activeTab === tab ? 'am-tab--active' : ''}"
						role="tab"
						aria-selected={activeTab === tab}
						onclick={() => activeTab = tab}
					>{tab.charAt(0).toUpperCase() + tab.slice(1)}</button>
				{/each}
			</div>
		</div>

		<!-- Tab Content -->
		<div class="am-detail-content">
			<div class="am-tab-content">
				{#if activeTab === 'overview'}
					<!-- Overview Tab -->
					<div class="am-section">
						<h2 class="am-section__title">Actions</h2>
						{#if mod.manifest.actions.length === 0}
							<p class="am-detail-muted">No actions defined for this module.</p>
						{:else}
							<div class="am-action-list">
								{#each mod.manifest.actions as action}
									<div class="am-action-card">
										<div class="am-action-card__header">
											<h3 class="am-action-card__name">{action.name}</h3>
											<span class="am-action-card__type">{action.type}</span>
										</div>
										<p class="am-action-card__desc">{action.description}</p>
									</div>
								{/each}
							</div>
						{/if}
					</div>

					{#if mod.creator_name}
						<div class="am-section">
							<h2 class="am-section__title">Author</h2>
							<p class="am-detail-muted">{mod.creator_name}</p>
						</div>
					{/if}
				{:else if activeTab === 'manifest'}
					<div class="am-section">
						<h2 class="am-section__title">Module Manifest</h2>
						<ManifestViewer manifest={mod.manifest} />
					</div>
				{:else if activeTab === 'versions'}
					<div class="am-section">
						<h2 class="am-section__title">Version History</h2>
						{#if storeState.versions.length === 0}
							<p class="am-detail-muted">No version history available.</p>
						{:else}
							<div class="am-action-list">
								{#each storeState.versions as version}
									<div class="am-action-card">
										<div class="am-action-card__header">
											<h3 class="am-action-card__name">v{version.version}</h3>
											<span class="am-action-card__type">
												{new Date(version.created_at).toLocaleDateString()}
											</span>
										</div>
										{#if version.changelog}
											<p class="am-action-card__desc">{version.changelog}</p>
										{/if}
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{:else if activeTab === 'settings'}
					<div class="am-section">
						<h2 class="am-section__title">Module Settings</h2>
						<div class="am-action-card">
							<label class="am-setting-row">
								<div>
									<p class="am-action-card__name">Active</p>
									<p class="am-action-card__desc">Enable or disable this module</p>
								</div>
								<input
									type="checkbox"
									checked={mod.is_active}
									class="am-checkbox"
									aria-label="Toggle module active state"
								/>
							</label>
						</div>
					</div>
				{/if}
			</div>
		</div>

		<!-- Share Dialog -->
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
	/*  MODULE DETAIL PAGE (am-detail-) — Foundation Tokens         */
	/* ══════════════════════════════════════════════════════════════ */
	.am-detail-page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--dbg, #fff);
	}
	.am-detail-header {
		flex-shrink: 0;
		padding: 20px 32px 0;
		border-bottom: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg, #fff);
	}
	.am-detail-content {
		flex: 1;
		overflow-y: auto;
		padding: 0 32px 32px;
	}

	/* Back button */
	.am-back-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		color: var(--dt3, #888);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
		margin-bottom: 16px;
		transition: color .15s;
	}
	.am-back-btn:hover {
		color: var(--dt, #111);
	}

	/* Center states */
	.am-detail-center {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		text-align: center;
		gap: 8px;
	}
	.am-detail-text {
		font-size: 14px;
		font-weight: 500;
		color: var(--dt, #111);
	}
	.am-detail-muted {
		font-size: 13px;
		color: var(--dt3, #888);
	}
	.am-detail-center :global(.am-detail-spinner) {
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
	.am-detail-error-icon {
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

	/* Detail header */
	.am-detail {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.am-detail__header {
		display: flex;
		align-items: flex-start;
		gap: 16px;
	}
	.am-detail__icon {
		width: 64px;
		height: 64px;
		border-radius: 18px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #fff;
		font-size: 20px;
		font-weight: 700;
		flex-shrink: 0;
	}
	.am-detail__icon--fallback {
		background: var(--dbg3, #eee);
		color: var(--dt3, #888);
	}
	.am-detail__info {
		flex: 1;
	}
	.am-detail__name {
		font-size: 20px;
		font-weight: 700;
		color: var(--dt, #111);
		margin-bottom: 6px;
	}
	.am-detail__sub {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 6px;
		font-size: 13px;
		color: var(--dt2, #555);
		margin-bottom: 8px;
	}
	.am-detail__desc {
		font-size: 13px;
		color: var(--dt2, #555);
		line-height: 1.5;
		margin-bottom: 8px;
	}
	.am-detail__stats {
		display: flex;
		gap: 12px;
	}
	.am-dot {
		margin: 0 2px;
		color: var(--dt4, #bbb);
	}

	/* Shared atoms */
	.am-cat-badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 8px;
		border-radius: 999px;
		font-size: 11px;
		font-weight: 600;
	}
	.am-visibility-badge {
		display: inline-flex;
		align-items: center;
		padding: 2px 8px;
		border-radius: 999px;
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.am-visibility-badge--public {
		background: #10b98122;
		color: #10b981;
	}
	.am-visibility-badge--workspace {
		background: #6366f122;
		color: #6366f1;
	}
	.am-visibility-badge--private {
		background: #6b728022;
		color: #6b7280;
	}
	.am-stat {
		display: inline-flex;
		align-items: center;
		gap: 3px;
		font-size: 12px;
		color: var(--dt3, #888);
	}

	/* Action buttons */
	.am-detail__actions {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 8px;
	}
	.am-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 7px 16px;
		border-radius: 999px;
		border: none;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		transition: all .18s;
		white-space: nowrap;
		gap: 6px;
	}
	.am-btn--install {
		background: var(--accent-blue, #3b82f6);
		color: #fff;
	}
	.am-btn--install:hover {
		filter: brightness(0.9);
		transform: translateY(-1px);
	}
	.am-btn--ghost {
		background: transparent;
		color: var(--dt2, #555);
		border: 1px solid var(--dbd, #e0e0e0);
	}
	.am-btn--ghost:hover {
		border-color: var(--accent-blue, #3b82f6);
		color: var(--accent-blue, #3b82f6);
	}
	.am-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
		transform: none;
	}
	.am-icon-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 7px 14px;
		border-radius: 999px;
		border: 1px solid var(--dbd, #e0e0e0);
		background: transparent;
		color: var(--dt2, #555);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all .15s;
	}
	.am-icon-btn:hover {
		border-color: var(--accent-blue, #3b82f6);
		color: var(--accent-blue, #3b82f6);
	}
	.am-icon-btn--danger:hover {
		border-color: #ef4444;
		color: #ef4444;
	}

	/* Tabs */
	.am-tabs {
		display: flex;
		gap: 0;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
		overflow-x: auto;
		margin-top: 16px;
	}
	.am-tab {
		padding: 10px 20px;
		border: none;
		background: transparent;
		color: var(--dt3, #888);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		border-bottom: 2px solid transparent;
		margin-bottom: -1px;
		transition: all .15s;
		white-space: nowrap;
	}
	.am-tab:hover {
		color: var(--dt, #111);
	}
	.am-tab--active {
		color: var(--accent-blue, #3b82f6);
		border-bottom-color: var(--accent-blue, #3b82f6);
	}

	/* Tab content area */
	.am-tab-content {
		padding-top: 24px;
	}

	/* Sections */
	.am-section {
		margin-bottom: 24px;
		max-width: 800px;
	}
	.am-section__title {
		font-size: 16px;
		font-weight: 600;
		color: var(--dt, #111);
		margin-bottom: 12px;
	}

	/* Action / version cards */
	.am-action-list {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.am-action-card {
		padding: 14px 16px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 12px;
		background: var(--dbg2, #f5f5f5);
		transition: border-color .15s;
	}
	.am-action-card:hover {
		border-color: var(--dbd2, #f0f0f0);
	}
	.am-action-card__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 6px;
	}
	.am-action-card__name {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt, #111);
	}
	.am-action-card__type {
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 999px;
		background: var(--dbg3, #eee);
		color: var(--dt3, #888);
		font-weight: 500;
	}
	.am-action-card__desc {
		font-size: 12px;
		color: var(--dt2, #555);
		line-height: 1.5;
	}

	/* Settings row */
	.am-setting-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		cursor: pointer;
		width: 100%;
	}
	.am-checkbox {
		width: 18px;
		height: 18px;
		accent-color: var(--accent-blue, #3b82f6);
		cursor: pointer;
	}
</style>
