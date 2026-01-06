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
				return 'text-green-600 dark:text-green-400';
			case 'failed':
				return 'text-red-600 dark:text-red-400';
			case 'in_progress':
				return 'text-blue-600 dark:text-blue-400';
			default:
				return 'text-gray-600 dark:text-gray-400';
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
				return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
			case 'expired':
				return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
			case 'error':
				return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
			default:
				return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
		}
	}
</script>

<svelte:head>
	<title>{integration?.provider_name || 'Integration'} Settings | BusinessOS</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			<div class="flex items-center gap-4">
				<a
					href="/integrations"
					aria-label="Back to integrations"
					class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
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
					<div class="flex items-center gap-3">
						<div
							class="w-12 h-12 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center"
						>
							{#if integration.icon_url}
								<img src={integration.icon_url} alt={integration.provider_name} class="w-8 h-8" />
							{:else}
								<span class="text-xl font-medium text-gray-600 dark:text-gray-300">
									{integration.provider_name.charAt(0)}
								</span>
							{/if}
						</div>
						<div>
							<div class="flex items-center gap-2">
								<h1 class="text-xl font-bold text-gray-900 dark:text-white">
									{integration.provider_name}
								</h1>
								<span
									class="px-2 py-0.5 text-xs rounded-full {getStatusBadgeClass(integration.status)}"
								>
									{integration.status}
								</span>
							</div>
							<p class="text-sm text-gray-500 dark:text-gray-400">
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
	<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		{#if isLoading}
			<div class="flex items-center justify-center h-64">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
			</div>
		{:else if !integration}
			<div class="text-center py-12">
				<svg
					class="mx-auto h-12 w-12 text-gray-400"
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
				<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">
					Integration not found
				</h3>
				<p class="mt-2 text-gray-500 dark:text-gray-400">
					This integration doesn't exist or you don't have access to it.
				</p>
				<a
					href="/integrations"
					class="mt-4 inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
				>
					Back to Integrations
				</a>
			</div>
		{:else}
			<!-- Messages -->
			{#if error}
				<div class="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
					<p class="text-red-800 dark:text-red-400">{error}</p>
				</div>
			{/if}
			{#if successMessage}
				<div class="mb-6 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
					<p class="text-green-800 dark:text-green-400">{successMessage}</p>
				</div>
			{/if}

			<div class="space-y-6">
				<!-- Sync Stats Banner -->
				{#if integration.sync_stats}
					{@const stats = integration.sync_stats}
					<div class="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-lg border border-blue-200 dark:border-blue-800 p-6">
						<div class="flex items-center justify-between mb-4">
							<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Sync Statistics</h2>
							<button
								onclick={triggerSync}
								disabled={isSyncing}
								class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
							>
								{#if isSyncing}
									<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
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
						<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
							<div class="bg-white dark:bg-gray-800 rounded-lg p-4">
								<div class="text-2xl font-bold text-blue-600 dark:text-blue-400">{stats.total_items}</div>
								<div class="text-sm text-gray-500 dark:text-gray-400">Total Items</div>
							</div>
							<div class="bg-white dark:bg-gray-800 rounded-lg p-4">
								<div class="text-2xl font-bold text-indigo-600 dark:text-indigo-400">{stats.sync_count}</div>
								<div class="text-sm text-gray-500 dark:text-gray-400">Total Syncs</div>
							</div>
							<div class="bg-white dark:bg-gray-800 rounded-lg p-4">
								<div class="text-sm font-medium {getSyncStatusColor(stats.last_sync_status)}">{stats.last_sync_status || 'N/A'}</div>
								<div class="text-sm text-gray-500 dark:text-gray-400">Last Status</div>
							</div>
							<div class="bg-white dark:bg-gray-800 rounded-lg p-4">
								<div class="text-sm font-medium text-gray-900 dark:text-white">{formatRelativeTime(stats.last_sync)}</div>
								<div class="text-sm text-gray-500 dark:text-gray-400">Last Sync</div>
							</div>
						</div>
						{#if stats.items_by_type && Object.keys(stats.items_by_type).length > 0}
							<div class="mt-4 pt-4 border-t border-blue-200 dark:border-blue-700">
								<div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Data Breakdown</div>
								<div class="flex flex-wrap gap-2">
									{#each Object.entries(stats.items_by_type) as [type, count]}
										<span class="px-3 py-1 bg-white dark:bg-gray-800 rounded-full text-sm">
											<span class="font-medium text-gray-900 dark:text-white">{count}</span>
											<span class="text-gray-500 dark:text-gray-400 ml-1 capitalize">{type}</span>
										</span>
									{/each}
								</div>
							</div>
						{/if}
						{#if stats.date_range}
							<div class="mt-4 pt-4 border-t border-blue-200 dark:border-blue-700">
								<div class="text-sm text-gray-600 dark:text-gray-400">
									Data Range: <span class="font-medium text-gray-900 dark:text-white">{formatDate(stats.date_range.from)}</span>
									to <span class="font-medium text-gray-900 dark:text-white">{formatDate(stats.date_range.to)}</span>
								</div>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Connection Info -->
				<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
						Connection Details
					</h2>
					<dl class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<div>
							<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Account</dt>
							<dd class="mt-1 text-sm text-gray-900 dark:text-white">
								{integration.external_account_name || 'N/A'}
							</dd>
						</div>
						{#if integration.external_workspace_name}
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Workspace</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">
									{integration.external_workspace_name}
								</dd>
							</div>
						{/if}
						<div>
							<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Connected</dt>
							<dd class="mt-1 text-sm text-gray-900 dark:text-white">
								{formatDate(integration.connected_at)}
							</dd>
						</div>
						{#if integration.last_used_at}
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Last Used</dt>
								<dd class="mt-1 text-sm text-gray-900 dark:text-white">
									{formatRelativeTime(integration.last_used_at)}
								</dd>
							</div>
						{/if}
						<div>
							<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Category</dt>
							<dd class="mt-1 text-sm text-gray-900 dark:text-white capitalize">
								{integration.category}
							</dd>
						</div>
						<div>
							<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Provider ID</dt>
							<dd class="mt-1 text-sm text-gray-900 dark:text-white font-mono">
								{integration.provider_id}
							</dd>
						</div>
					</dl>
				</div>

				<!-- Skills -->
				{#if integration.skills?.length > 0}
					<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
						<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
							Available Skills
						</h2>
						<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
							Enable or disable specific AI skills for this integration.
						</p>
						<div class="space-y-2">
							{#each integration.skills as skill}
								<label class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
									<input
										type="checkbox"
										checked={settings.enabledSkills?.includes(skill) ?? false}
										onchange={() => toggleSkill(skill)}
										class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
									/>
									<span class="text-sm text-gray-900 dark:text-white">{skill}</span>
								</label>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Modules -->
				{#if integration.modules?.length > 0}
					<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
						<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
							Available In Modules
						</h2>
						<div class="flex flex-wrap gap-2">
							{#each integration.modules as mod}
								<a
									href="/{mod.toLowerCase()}"
									class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-sm"
								>
									{mod}
								</a>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Available Permissions -->
				{#if integration.available_permissions && integration.available_permissions.length > 0}
					<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
						<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
							Permissions
						</h2>
						<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
							Data access permissions for this integration.
						</p>
						<div class="space-y-3">
							{#each integration.available_permissions as permission}
								<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-full flex items-center justify-center {permission.granted ? 'bg-green-100 dark:bg-green-900/30' : 'bg-gray-200 dark:bg-gray-600'}">
											{#if permission.granted}
												<svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
												</svg>
											{:else}
												<svg class="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
												</svg>
											{/if}
										</div>
										<div>
											<div class="text-sm font-medium text-gray-900 dark:text-white">{permission.name}</div>
											<div class="text-xs text-gray-500 dark:text-gray-400">{permission.description}</div>
										</div>
									</div>
									<span class="text-xs px-2 py-1 rounded-full {permission.granted ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400' : 'bg-gray-100 dark:bg-gray-600 text-gray-600 dark:text-gray-400'}">
										{permission.granted ? 'Granted' : 'Not Granted'}
									</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Sync History -->
				{#if integration.sync_history && integration.sync_history.length > 0}
					<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
						<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
							Sync History
						</h2>
						<div class="space-y-2">
							{#each integration.sync_history as sync}
								<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-full flex items-center justify-center {sync.status === 'completed' ? 'bg-green-100 dark:bg-green-900/30' : sync.status === 'failed' ? 'bg-red-100 dark:bg-red-900/30' : 'bg-blue-100 dark:bg-blue-900/30'}">
											{#if sync.status === 'completed'}
												<svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
												</svg>
											{:else if sync.status === 'failed'}
												<svg class="w-4 h-4 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
												</svg>
											{:else}
												<svg class="w-4 h-4 text-blue-600 dark:text-blue-400 animate-spin" fill="none" viewBox="0 0 24 24">
													<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
													<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
												</svg>
											{/if}
										</div>
										<div>
											<div class="text-sm font-medium text-gray-900 dark:text-white capitalize">{sync.sync_type} sync</div>
											<div class="text-xs text-gray-500 dark:text-gray-400">
												{formatRelativeTime(sync.started_at)}
												{#if sync.records_synced}
													 - {sync.records_synced} records
												{/if}
											</div>
										</div>
									</div>
									<span class="text-xs px-2 py-1 rounded-full capitalize {sync.status === 'completed' ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400' : sync.status === 'failed' ? 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400' : 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400'}">
										{sync.status}
									</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Notifications -->
				<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
						Notifications
					</h2>
					<label class="flex items-center gap-3">
						<input
							type="checkbox"
							bind:checked={settings.notifications}
							class="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
						/>
						<span class="text-sm text-gray-900 dark:text-white">
							Enable notifications from this integration
						</span>
					</label>
				</div>

				<!-- Actions -->
				<div class="flex items-center justify-between pt-4">
					<button
						onclick={handleDisconnect}
						class="px-4 py-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
					>
						Disconnect Integration
					</button>
					<button
						onclick={saveSettings}
						disabled={isSaving}
						class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{isSaving ? 'Saving...' : 'Save Settings'}
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>
