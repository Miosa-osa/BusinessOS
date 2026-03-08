<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import * as integrationsApi from '$lib/api/integrations';
	import type { UserIntegration, IntegrationSettings, IntegrationSyncStats, SyncHistoryEntry, AvailablePermission } from '$lib/api/integrations';

	const session = useSession();

	// State
	let isLoading = $state(true);
	let isSaving = $state(false);
	let isSyncing = $state(false);
	let integration = $state<UserIntegration | null>(null);
	let settings = $state<IntegrationSettings>({
		enabledSkills: [],
		notifications: true,
		syncSettings: {}
	});
	let error = $state<string | null>(null);
	let successMessage = $state<string | null>(null);

	// Skill execution state
	let executingSkillId = $state<string | null>(null);
	let skillExecutionMessage = $state<string | null>(null);

	// Track unsaved changes
	let initialSettings = $state<string>('');
	let hasUnsavedChanges = $derived(
		initialSettings !== '' && JSON.stringify(settings) !== initialSettings
	);

	// Get integration ID from route params
	let integrationId = $derived($page.params.id ?? '');

	onMount(async () => {
		if (!$session?.data?.user) {
			goto('/login');
			return;
		}

		if (integrationId) {
			await loadIntegration();
		}
	});

	async function loadIntegration() {
		if (!integrationId) return;
		isLoading = true;
		error = null;
		try {
			const response = await integrationsApi.getUserIntegration(integrationId);
			integration = response.integration;
			// Ensure we have default values for settings
			const apiSettings = response.integration.settings || {};
			settings = {
				enabledSkills: apiSettings.enabledSkills || [],
				notifications: apiSettings.notifications ?? true,
				syncSettings: apiSettings.syncSettings || {}
			};
			initialSettings = JSON.stringify(settings);
		} catch (err) {
			console.error('Failed to load integration:', err);
			error = 'Failed to load integration details';
		} finally {
			isLoading = false;
		}
	}

	async function saveSettings() {
		if (!integrationId) return;
		isSaving = true;
		error = null;
		successMessage = null;
		try {
			await integrationsApi.updateIntegrationSettings(integrationId, settings);
			initialSettings = JSON.stringify(settings);
			successMessage = 'Settings saved successfully';
			setTimeout(() => (successMessage = null), 3000);
		} catch (err) {
			console.error('Failed to save settings:', err);
			error = 'Failed to save settings';
		} finally {
			isSaving = false;
		}
	}

	async function handleDisconnect() {
		if (!integrationId) return;
		if (!confirm('Are you sure you want to disconnect this integration?')) {
			return;
		}

		try {
			await integrationsApi.disconnectUserIntegration(integrationId);
			goto('/integrations');
		} catch (err) {
			console.error('Failed to disconnect:', err);
			error = 'Failed to disconnect integration';
		}
	}

	async function triggerSync() {
		if (!integrationId || isSyncing) return;
		isSyncing = true;
		error = null;
		try {
			const response = await integrationsApi.triggerIntegrationSync(integrationId);
			successMessage = response.message || 'Sync completed';
			// Reload integration to get updated stats
			await loadIntegration();
			setTimeout(() => (successMessage = null), 5000);
		} catch (err) {
			console.error('Failed to trigger sync:', err);
			error = 'Failed to trigger sync';
		} finally {
			isSyncing = false;
		}
	}

	// Helper functions for formatting
	function formatDate(dateStr: string | null | undefined): string {
		if (!dateStr) return 'Never';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatRelativeTime(dateStr: string | null | undefined): string {
		if (!dateStr) return 'Never';
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(diff / 3600000);
		const days = Math.floor(diff / 86400000);

		if (minutes < 1) return 'Just now';
		if (minutes < 60) return `${minutes}m ago`;
		if (hours < 24) return `${hours}h ago`;
		if (days < 7) return `${days}d ago`;
		return formatDate(dateStr);
	}

	function getSyncStatusColor(status: string | null | undefined): string {
		switch (status) {
			case 'completed':
				return 'ih-sync-status--completed';
			case 'failed':
				return 'ih-sync-status--failed';
			case 'in_progress':
				return 'ih-sync-status--progress';
			default:
				return 'ih-sync-status--default';
		}
	}

	function toggleSkill(skillId: string) {
		const currentSkills = settings.enabledSkills || [];
		if (currentSkills.includes(skillId)) {
			settings.enabledSkills = currentSkills.filter((s) => s !== skillId);
		} else {
			settings.enabledSkills = [...currentSkills, skillId];
		}
	}

	function getStatusBadgeClass(status: string) {
		switch (status) {
			case 'connected':
				return 'ih-badge--connected';
			case 'expired':
				return 'ih-badge--expired';
			case 'error':
				return 'ih-badge--error';
			default:
				return 'ih-badge--default';
		}
	}

	async function handleReAuth() {
		if (!integration) return;
		const oauthProvider = integration.oauth_provider || integration.provider_id;
		try {
			const response = await integrationsApi.initiateAuth(
				oauthProvider as integrationsApi.IntegrationProvider
			);
			if (response.auth_url) {
				window.open(response.auth_url, '_blank', 'width=600,height=700');
			}
		} catch (err) {
			console.error('Failed to re-authorize:', err);
			error = 'Failed to initiate re-authorization';
		}
	}

	async function handleTriggerSkill(skillId: string) {
		executingSkillId = skillId;
		skillExecutionMessage = null;
		try {
			const result = await integrationsApi.triggerSkill(skillId);
			skillExecutionMessage = result.message || `Skill "${skillId}" triggered (${result.status})`;
			setTimeout(() => (skillExecutionMessage = null), 5000);
		} catch (err) {
			console.error('Failed to trigger skill:', err);
			error = `Failed to trigger skill "${skillId}"`;
		} finally {
			executingSkillId = null;
		}
	}
</script>

<svelte:head>
	<title>{integration?.provider_name || 'Integration'} Settings | BusinessOS</title>
</svelte:head>

<div class="ih-settings-page">
	<!-- Header -->
	<div class="ih-settings-header">
		<div class="ih-settings-header__inner">
			<!-- Breadcrumb -->
			<nav class="ih-breadcrumb">
				<a href="/integrations" class="ih-breadcrumb__link">Integrations</a>
				<svg class="w-4 h-4 ih-breadcrumb__sep" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
				<span class="ih-breadcrumb__current">{integration?.provider_name || 'Settings'}</span>
			</nav>
			<div class="ih-settings-header__row">
				<a
					href="/integrations"
					aria-label="Back to integrations"
					class="ih-back-btn"
				>
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M15 19l-7-7 7-7"
						/>
					</svg>
				</a>
				{#if integration}
					<div class="ih-settings-provider">
						<div class="ih-settings-provider__icon">
							{#if integration.icon_url}
								<img src={integration.icon_url} alt={integration.provider_name} class="ih-settings-provider__img" />
							{:else}
								<span class="ih-settings-provider__letter">
									{integration.provider_name.charAt(0)}
								</span>
							{/if}
						</div>
						<div>
							<div class="ih-settings-provider__name-row">
								<h1 class="ih-settings-provider__name">
									{integration.provider_name}
								</h1>
								<span
									class="ih-badge {getStatusBadgeClass(integration.status)}"
								>
									{integration.status}
								</span>
								{#if integration.status === 'expired' || integration.status === 'error'}
									<button
										onclick={handleReAuth}
										class="btn-pill btn-pill-primary btn-pill-sm"
									>
										Re-authorize
									</button>
								{/if}
							</div>
							<p class="ih-settings-provider__meta">
								{integration.external_account_name ||
									integration.external_workspace_name ||
									integration.category}
							</p>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="ih-settings-content">
		{#if isLoading}
			<div class="ih-spinner-wrap ih-spinner-wrap--tall">
				<div class="ih-spinner"></div>
			</div>
		{:else if !integration}
			<div class="ih-empty">
				<svg
					class="ih-empty__icon"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M12 12h.01M12 14h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
					/>
				</svg>
				<h3 class="ih-empty__title">
					Integration not found
				</h3>
				<p class="ih-empty__text">
					This integration doesn't exist or you don't have access to it.
				</p>
				<a
					href="/integrations"
					class="btn-pill btn-pill-primary ih-empty__action"
				>
					Back to Integrations
				</a>
			</div>
		{:else}
			<!-- Messages -->
			{#if error}
				<div class="ih-alert ih-alert--error ih-alert--banner">
					<p>{error}</p>
				</div>
			{/if}
			{#if successMessage}
				<div class="ih-alert ih-alert--success ih-alert--banner">
					<p>{successMessage}</p>
				</div>
			{/if}

			<div class="ih-settings-sections">
				<!-- Sync Stats Banner -->
				{#if integration.sync_stats}
					{@const stats = integration.sync_stats}
					<div class="ih-sync-banner">
						<div class="ih-sync-banner__top">
							<h2 class="ih-section-title">Sync Statistics</h2>
							<button
								onclick={triggerSync}
								disabled={isSyncing}
								class="btn-pill btn-pill-primary ih-sync-btn"
							>
								{#if isSyncing}
									<svg class="w-4 h-4 ih-spinner--inline" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
									</svg>
									Syncing...
								{:else}
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
									</svg>
									Sync Now
								{/if}
							</button>
						</div>
						<div class="ih-stats-grid">
							<div class="ih-stat-card">
								<div class="ih-stat-card__value ih-stat-card__value--blue">{stats.total_items}</div>
								<div class="ih-stat-card__label">Total Items</div>
							</div>
							<div class="ih-stat-card">
								<div class="ih-stat-card__value ih-stat-card__value--indigo">{stats.sync_count}</div>
								<div class="ih-stat-card__label">Total Syncs</div>
							</div>
							<div class="ih-stat-card">
								<div class="ih-stat-card__value-sm {getSyncStatusColor(stats.last_sync_status)}">{stats.last_sync_status || 'N/A'}</div>
								<div class="ih-stat-card__label">Last Status</div>
							</div>
							<div class="ih-stat-card">
								<div class="ih-stat-card__value-sm">{formatRelativeTime(stats.last_sync)}</div>
								<div class="ih-stat-card__label">Last Sync</div>
							</div>
						</div>
						{#if stats.items_by_type && Object.keys(stats.items_by_type).length > 0}
							<div class="ih-sync-breakdown">
								<div class="ih-sync-breakdown__title">Data Breakdown</div>
								<div class="ih-sync-breakdown__tags">
									{#each Object.entries(stats.items_by_type) as [type, count]}
										<span class="ih-breakdown-tag">
											<span class="ih-breakdown-tag__count">{count}</span>
											<span class="ih-breakdown-tag__type">{type}</span>
										</span>
									{/each}
								</div>
							</div>
						{/if}
						{#if stats.date_range}
							<div class="ih-sync-breakdown">
								<div class="ih-sync-breakdown__range">
									Data Range: <span class="ih-sync-breakdown__date">{formatDate(stats.date_range.from)}</span>
									to <span class="ih-sync-breakdown__date">{formatDate(stats.date_range.to)}</span>
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Connection Info -->
				<div class="ih-settings-card">
					<h2 class="ih-section-title">
						Connection Details
					</h2>
					<dl class="ih-detail-grid">
						<div>
							<dt class="ih-detail-label">Account</dt>
							<dd class="ih-detail-value">
								{integration.external_account_name || 'N/A'}
							</dd>
						</div>
						{#if integration.external_workspace_name}
							<div>
								<dt class="ih-detail-label">Workspace</dt>
								<dd class="ih-detail-value">
									{integration.external_workspace_name}
								</dd>
							</div>
						{/if}
						<div>
							<dt class="ih-detail-label">Connected</dt>
							<dd class="ih-detail-value">
								{formatDate(integration.connected_at)}
							</dd>
						</div>
						{#if integration.last_used_at}
							<div>
								<dt class="ih-detail-label">Last Used</dt>
								<dd class="ih-detail-value">
									{formatRelativeTime(integration.last_used_at)}
								</dd>
							</div>
						{/if}
						<div>
							<dt class="ih-detail-label">Category</dt>
							<dd class="ih-detail-value ih-detail-value--cap">
								{integration.category}
							</dd>
						</div>
						<div>
							<dt class="ih-detail-label">Provider ID</dt>
							<dd class="ih-detail-value ih-detail-value--mono">
								{integration.provider_id}
							</dd>
						</div>
					</dl>
				</div>

				<!-- Skills -->
				{#if integration.skills?.length > 0}
					<div class="ih-settings-card">
						<h2 class="ih-section-title">
							Available Skills
						</h2>
						<p class="ih-section-desc">
							Enable or disable specific AI skills for this integration.
						</p>
						<div class="ih-skill-list">
							{#each integration.skills as skill}
								<div class="ih-skill-row-wrap">
									<label class="ih-skill-row">
										<input
											type="checkbox"
											checked={settings.enabledSkills?.includes(skill) ?? false}
											onchange={() => toggleSkill(skill)}
											class="ih-checkbox"
										/>
										<span class="ih-skill-name">{skill}</span>
									</label>
									<button
										onclick={() => handleTriggerSkill(skill)}
										disabled={executingSkillId === skill || !(settings.enabledSkills?.includes(skill))}
										class="ih-skill-run-btn"
										title={settings.enabledSkills?.includes(skill) ? 'Run skill' : 'Enable skill first'}
									>
										{#if executingSkillId === skill}
											<svg class="w-3.5 h-3.5 ih-spin" viewBox="0 0 24 24" fill="none">
												<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
												<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
											</svg>
										{:else}
											<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
											</svg>
										{/if}
									</button>
								</div>
							{/each}
						</div>
						{#if skillExecutionMessage}
							<div class="ih-alert ih-alert--success ih-alert--sm">
								<p>{skillExecutionMessage}</p>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Modules -->
				{#if integration.modules?.length > 0}
					<div class="ih-settings-card">
						<h2 class="ih-section-title">
							Available In Modules
						</h2>
						<div class="ih-module-list">
							{#each integration.modules as mod}
								<a
									href="/{mod.toLowerCase()}"
									class="ih-module-link"
								>
									{mod}
								</a>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Available Permissions -->
				{#if integration.available_permissions && integration.available_permissions.length > 0}
					<div class="ih-settings-card">
						<h2 class="ih-section-title">
							Permissions
						</h2>
						<p class="ih-section-desc">
							Data access permissions for this integration.
						</p>
						<div class="ih-perm-list">
							{#each integration.available_permissions as permission}
								<div class="ih-perm-row">
									<div class="ih-perm-row__left">
										<div class="ih-perm-icon {permission.granted ? 'ih-perm-icon--granted' : 'ih-perm-icon--locked'}">
											{#if permission.granted}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
												</svg>
											{:else}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
												</svg>
											{/if}
										</div>
										<div>
											<div class="ih-perm-name">{permission.name}</div>
											<div class="ih-perm-desc">{permission.description}</div>
										</div>
									</div>
									<span class="ih-perm-badge {permission.granted ? 'ih-perm-badge--granted' : 'ih-perm-badge--locked'}">
										{permission.granted ? 'Granted' : 'Not Granted'}
									</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Sync History -->
				{#if integration.sync_history && integration.sync_history.length > 0}
					<div class="ih-settings-card">
						<h2 class="ih-section-title">
							Sync History
						</h2>
						<div class="ih-history-list">
							{#each integration.sync_history as sync}
								<div class="ih-history-row">
									<div class="ih-history-row__left">
										<div class="ih-history-icon {sync.status === 'completed' ? 'ih-history-icon--ok' : sync.status === 'failed' ? 'ih-history-icon--fail' : 'ih-history-icon--progress'}">
											{#if sync.status === 'completed'}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
												</svg>
											{:else if sync.status === 'failed'}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
												</svg>
											{:else}
												<svg class="w-4 h-4 ih-spin" fill="none" viewBox="0 0 24 24">
													<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
													<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
												</svg>
											{/if}
										</div>
										<div>
											<div class="ih-history-type">{sync.sync_type} sync</div>
											<div class="ih-history-meta">
												{formatRelativeTime(sync.started_at)}
												{#if sync.records_synced}
													 - {sync.records_synced} records
												{/if}
											</div>
										</div>
									</div>
									<span class="ih-history-badge {sync.status === 'completed' ? 'ih-history-badge--ok' : sync.status === 'failed' ? 'ih-history-badge--fail' : 'ih-history-badge--progress'}">
										{sync.status}
									</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Notifications -->
				<div class="ih-settings-card">
					<h2 class="ih-section-title">
						Notifications
					</h2>
					<label class="ih-notif-label">
						<input
							type="checkbox"
							bind:checked={settings.notifications}
							class="ih-checkbox"
						/>
						<span class="ih-notif-text">
							Enable notifications from this integration
						</span>
					</label>
				</div>

				<!-- Actions -->
				<div class="ih-settings-actions">
					<button
						onclick={handleDisconnect}
						class="btn-pill btn-pill-ghost"
					>
						Disconnect Integration
					</button>
					<div class="ih-settings-actions__right">
						{#if hasUnsavedChanges}
							<span class="ih-unsaved-indicator">Unsaved changes</span>
						{/if}
						<button
							onclick={saveSettings}
							disabled={isSaving}
							class="btn-pill btn-pill-primary"
						>
							{isSaving ? 'Saving...' : 'Save Settings'}
						</button>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>
<style>
	/* ─── Settings Page Shell ─── */
	.ih-settings-page {
		min-height: 100vh;
		overflow-y: auto;
		background: var(--dbg);
	}

	/* ─── Header ─── */
	.ih-settings-header {
		background: var(--dbg2);
		border-bottom: 1px solid var(--dbd);
	}
	.ih-settings-header__inner {
		max-width: 64rem;
		margin: 0 auto;
		padding: 1.5rem;
	}
	.ih-settings-header__row {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.ih-back-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: var(--dt3);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.375rem 0.75rem;
		border-radius: 0.5rem;
		transition: color 0.15s, background 0.15s;
	}
	.ih-back-btn:hover {
		color: var(--dt);
		background: var(--dbg3);
	}
	.ih-settings-provider {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.ih-settings-provider__icon {
		width: 3rem;
		height: 3rem;
		border-radius: 0.75rem;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		border: 1px solid var(--dbd);
		background: var(--dbg3);
	}
	.ih-settings-provider__img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.ih-settings-provider__letter {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--dt2);
	}
	.ih-settings-provider__name-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.ih-settings-provider__name {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--dt);
	}
	.ih-settings-provider__meta {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-top: 0.125rem;
	}

	/* ─── Content ─── */
	.ih-settings-content {
		max-width: 64rem;
		margin: 0 auto;
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	/* ─── Shared Card ─── */
	.ih-settings-card {
		background: var(--dbg2);
		border-radius: 0.75rem;
		border: 1px solid var(--dbd);
		padding: 1.5rem;
	}

	/* ─── Section Title / Desc ─── */
	.ih-section-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
		margin-bottom: 1rem;
	}
	.ih-section-desc {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-bottom: 1rem;
	}

	/* ─── Badges ─── */
	.ih-badge--connected {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
		font-size: 0.75rem;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-weight: 500;
	}
	.ih-badge--expired {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
		font-size: 0.75rem;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-weight: 500;
	}
	.ih-badge--error {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
		font-size: 0.75rem;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-weight: 500;
	}
	.ih-badge--default {
		background: rgba(107, 114, 128, 0.1);
		color: var(--dt3);
		font-size: 0.75rem;
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-weight: 500;
	}

	/* ─── Sync Banner ─── */
	.ih-sync-banner {
		background: var(--dbg2);
		border-radius: 0.75rem;
		border: 1px solid var(--dbd);
		padding: 1.5rem;
	}
	.ih-sync-banner__top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}
	.ih-sync-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	/* ─── Stats Grid ─── */
	.ih-stats-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin-bottom: 1rem;
	}
	@media (min-width: 640px) {
		.ih-stats-grid { grid-template-columns: repeat(4, 1fr); }
	}
	.ih-stat-card {
		background: var(--dbg3);
		border-radius: 0.5rem;
		padding: 1rem;
	}
	.ih-stat-card__value {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--dt);
	}
	.ih-stat-card__value--blue { color: #3b82f6; }
	.ih-stat-card__value--indigo { color: #6366f1; }
	.ih-stat-card__value-sm {
		font-size: 0.875rem;
		font-weight: 400;
		color: var(--dt3);
		margin-left: 0.25rem;
	}
	.ih-stat-card__label {
		font-size: 0.75rem;
		color: var(--dt3);
		margin-top: 0.25rem;
	}

	/* ─── Sync Breakdown ─── */
	.ih-sync-breakdown {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	.ih-sync-breakdown__tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	.ih-breakdown-tag {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.75rem;
		background: var(--dbg3);
		padding: 0.25rem 0.5rem;
		border-radius: 9999px;
		color: var(--dt3);
	}
	.ih-breakdown-tag__count { font-weight: 600; color: var(--dt); }
	.ih-breakdown-tag__type { color: var(--dt3); }
	.ih-sync-breakdown__range {
		font-size: 0.75rem;
		color: var(--dt4);
	}
	.ih-sync-breakdown__date { font-weight: 500; color: var(--dt3); }

	/* ─── Connection Details ─── */
	.ih-detail-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 1rem;
	}
	@media (min-width: 640px) {
		.ih-detail-grid { grid-template-columns: repeat(2, 1fr); }
	}
	.ih-detail-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt3);
	}
	.ih-detail-value {
		font-size: 0.875rem;
		color: var(--dt);
		margin-top: 0.25rem;
	}
	.ih-detail-value--cap { text-transform: capitalize; }
	.ih-detail-value--mono { font-family: monospace; }

	/* ─── Skills ─── */
	.ih-skill-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.ih-skill-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		background: var(--dbg3);
		border-radius: 0.5rem;
		cursor: pointer;
		transition: background 0.15s;
	}
	.ih-skill-row:hover { background: var(--dbg); }
	.ih-skill-name {
		font-size: 0.875rem;
		color: var(--dt);
	}

	/* ─── Checkbox ─── */
	.ih-checkbox {
		border-radius: 0.25rem;
		border: 1px solid var(--dbd);
		background: var(--dbg3);
		accent-color: #3b82f6;
	}

	/* ─── Module Links ─── */
	.ih-module-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	.ih-module-link {
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		background: var(--dbg3);
		color: var(--dt2);
		border-radius: 0.5rem;
		text-decoration: none;
		transition: background 0.15s;
	}
	.ih-module-link:hover { background: var(--dbg); }

	/* ─── Permissions ─── */
	.ih-perm-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.ih-perm-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem;
		background: var(--dbg3);
		border-radius: 0.5rem;
	}
	.ih-perm-row__left {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.ih-perm-icon {
		width: 2rem;
		height: 2rem;
		border-radius: 9999px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}
	.ih-perm-icon--granted {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}
	.ih-perm-icon--locked {
		background: var(--dbg);
		color: var(--dt4);
	}
	.ih-perm-name {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt);
	}
	.ih-perm-desc {
		font-size: 0.75rem;
		color: var(--dt3);
	}
	.ih-perm-badge {
		font-size: 0.75rem;
		padding: 0.25rem 0.5rem;
		border-radius: 9999px;
		white-space: nowrap;
	}
	.ih-perm-badge--granted {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}
	.ih-perm-badge--locked {
		background: rgba(107, 114, 128, 0.1);
		color: var(--dt3);
	}

	/* ─── Sync History ─── */
	.ih-history-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.ih-history-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem;
		background: var(--dbg3);
		border-radius: 0.5rem;
	}
	.ih-history-row__left {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.ih-history-icon {
		width: 2rem;
		height: 2rem;
		border-radius: 9999px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}
	.ih-history-icon--ok {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}
	.ih-history-icon--fail {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}
	.ih-history-icon--progress {
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}
	.ih-history-type {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt);
		text-transform: capitalize;
	}
	.ih-history-meta {
		font-size: 0.75rem;
		color: var(--dt3);
	}
	.ih-history-badge {
		font-size: 0.75rem;
		padding: 0.25rem 0.5rem;
		border-radius: 9999px;
		text-transform: capitalize;
		white-space: nowrap;
	}
	.ih-history-badge--ok {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}
	.ih-history-badge--fail {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}
	.ih-history-badge--progress {
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}

	/* ─── Notifications ─── */
	.ih-notif-label {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.ih-notif-text {
		font-size: 0.875rem;
		color: var(--dt);
	}

	/* ─── Actions ─── */
	.ih-settings-actions {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding-top: 1rem;
	}

	/* ─── Spinner ─── */
	.ih-spinner-wrap--tall {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 5rem 0;
	}
	.ih-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--dbd);
		border-top-color: #3b82f6;
		border-radius: 9999px;
		animation: ih-spin 0.6s linear infinite;
	}
	.ih-spin { animation: ih-spin 0.6s linear infinite; }
	@keyframes ih-spin { to { transform: rotate(360deg); } }

	/* ─── Empty / Not Found ─── */
	.ih-empty {
		text-align: center;
		padding: 3rem 1rem;
	}
	.ih-empty__title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
		margin-bottom: 0.5rem;
	}
	.ih-empty__text {
		font-size: 0.875rem;
		color: var(--dt3);
	}

	/* ─── Alerts ─── */
	.ih-alert--error {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
		padding: 0.75rem 1rem;
		border-radius: 0.5rem;
		font-size: 0.875rem;
	}
	.ih-alert--success {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
		padding: 0.75rem 1rem;
		border-radius: 0.5rem;
		font-size: 0.875rem;
	}
	.ih-alert--banner {
		border-radius: 0.75rem;
	}

	/* ─── Sync Status (helper classes) ─── */
	.ih-sync-status--completed { color: #22c55e; }
	.ih-sync-status--failed { color: #ef4444; }
	.ih-sync-status--progress { color: #3b82f6; }
	.ih-sync-status--default { color: var(--dt3); }

	/* ─── Breadcrumb ─── */
	.ih-breadcrumb {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		margin-bottom: 0.75rem;
		font-size: 0.8125rem;
	}
	.ih-breadcrumb__link {
		color: var(--dt3);
		text-decoration: none;
		transition: color 0.15s;
	}
	.ih-breadcrumb__link:hover {
		color: var(--dt);
	}
	.ih-breadcrumb__sep {
		color: var(--dt4);
	}
	.ih-breadcrumb__current {
		color: var(--dt);
		font-weight: 500;
	}

	/* ─── Skill Row with Trigger ─── */
	.ih-skill-row-wrap {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}
	.ih-skill-run-btn {
		padding: 0.25rem 0.5rem;
		border-radius: 0.375rem;
		border: 1px solid var(--dbd);
		background: none;
		color: var(--dt3);
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		transition: color 0.15s, border-color 0.15s, background 0.15s;
	}
	.ih-skill-run-btn:hover:not(:disabled) {
		color: #3b82f6;
		border-color: rgba(59, 130, 246, 0.3);
		background: rgba(59, 130, 246, 0.05);
	}
	.ih-skill-run-btn:disabled {
		opacity: 0.35;
		cursor: not-allowed;
	}

	/* ─── Unsaved Changes ─── */
	.ih-settings-actions__right {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.ih-unsaved-indicator {
		font-size: 0.75rem;
		color: #f59e0b;
		font-weight: 500;
		padding: 0.25rem 0.625rem;
		border-radius: 9999px;
		background: rgba(245, 158, 11, 0.1);
	}

	/* Alert small */
	.ih-alert--sm {
		margin-top: 0.5rem;
		padding: 0.5rem 0.75rem;
		font-size: 0.8125rem;
	}

	/* ─── Spin animation ─── */
	.ih-spin {
		animation: ih-spin 0.6s linear infinite;
	}
</style>