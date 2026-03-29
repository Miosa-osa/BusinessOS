<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Sparkles } from 'lucide-svelte';
	import { generatedAppsStore, type AppStatus } from '$lib/stores/generatedAppsStore';
	import GeneratedAppCard from '$lib/components/osa/GeneratedAppCard.svelte';
	import CreateAppModal from '$lib/components/osa/CreateAppModal.svelte';
	import { currentWorkspace } from '$lib/stores/workspaces';

	let { data } = $props();

	let showCreateModal = $state(false);
	let workspaceId = $derived($currentWorkspace?.id ?? data?.workspaceId ?? $page.data?.workspaceId ?? '');

	let loading = $state(false);
	let error = $state<string | null>(null);
	let selectedFilter = $state<AppStatus | 'all'>('all');
	let searchQuery = $state('');
	let viewMode = $state<'grid' | 'list'>('grid');

	// Subscribe to store
	let apps = $derived($generatedAppsStore.filteredApps);
	let storeState = $derived($generatedAppsStore);

	// Stats
	let stats = $derived({
		total: storeState.apps.length,
		generating: storeState.apps.filter((a) => a.status === 'generating').length,
		generated: storeState.apps.filter((a) => a.status === 'generated').length,
		deployed: storeState.apps.filter((a) => a.status === 'deployed').length,
		failed: storeState.apps.filter((a) => a.status === 'failed').length,
	});

	// Filter options
	const filterOptions: Array<{ value: AppStatus | 'all'; label: string; color: string }> = [
		{ value: 'all', label: 'All', color: 'bg-gray-500' },
		{ value: 'generating', label: 'Generating', color: 'bg-yellow-500' },
		{ value: 'generated', label: 'Generated', color: 'bg-blue-500' },
		{ value: 'deployed', label: 'Deployed', color: 'bg-green-500' },
		{ value: 'failed', label: 'Failed', color: 'bg-red-500' },
	];

	onMount(async () => {
		loading = true;
		try {
			await generatedAppsStore.fetchApps();

			// Subscribe to real-time updates for generating apps
			for (const app of storeState.apps) {
				if (app.status === 'generating') {
					generatedAppsStore.subscribeToAppProgress(app.id);
				}
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load apps';
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		// Cleanup SSE connections
		for (const app of storeState.apps) {
			if (app.status === 'generating') {
				generatedAppsStore.unsubscribeFromAppProgress(app.id);
			}
		}
	});

	function handleFilterChange(filter: AppStatus | 'all') {
		selectedFilter = filter;
		generatedAppsStore.setFilter(filter);
	}

	function handleSearchChange(e: Event) {
		const target = e.target as HTMLInputElement;
		searchQuery = target.value;
		generatedAppsStore.setSearchQuery(target.value);
	}

	function handleViewApp(app: any) {
		goto(`/generated-apps/${app.id}`);
	}

	async function handleDeployApp(app: any) {
		try {
			await generatedAppsStore.deployApp(app.id);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to deploy app';
		}
	}

	async function handleDeleteApp(app: any) {
		if (!confirm(`Are you sure you want to delete ${app.app_name}?`)) return;

		try {
			await generatedAppsStore.deleteApp(app.id);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete app';
		}
	}

	function handleRefresh() {
		generatedAppsStore.fetchApps();
	}
</script>

<div class="h-full flex flex-col bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			<div class="flex items-center justify-between mb-6">
				<div>
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Generated Apps</h1>
					<p class="text-sm text-gray-600 dark:text-gray-400 mt-1">
						Manage your OSA-generated applications
					</p>
				</div>

				<div class="flex items-center gap-3">
					<!-- Generation indicator -->
					{#if stats.generating > 0}
						<div class="flex items-center gap-2 px-3 py-1.5 bg-yellow-50 dark:bg-yellow-900/30 border border-yellow-200 dark:border-yellow-800 rounded-lg">
							<svg class="w-4 h-4 text-yellow-600 dark:text-yellow-400 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
							</svg>
							<span class="text-xs font-medium text-yellow-700 dark:text-yellow-300">
								{stats.generating} app{stats.generating !== 1 ? 's' : ''} generating
							</span>
						</div>
					{/if}

					<!-- Generate New App Button -->
					<button
						onclick={() => showCreateModal = true}
						class="px-4 py-2 text-sm font-medium text-white bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg hover:from-blue-700 hover:to-purple-700 transition-all flex items-center gap-2"
					>
						<Sparkles class="w-4 h-4" />
						Generate App
					</button>

					<!-- Refresh Button -->
					<button
						onclick={handleRefresh}
						disabled={loading}
						class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2"
					>
						<svg
							class="w-4 h-4"
							class:animate-spin={loading}
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
							/>
						</svg>
						Refresh
					</button>

					<!-- View Mode Toggle -->
					<div class="flex items-center bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg p-1">
						<button
							onclick={() => (viewMode = 'grid')}
							class="px-3 py-1.5 text-sm font-medium rounded {viewMode === 'grid'
								? 'bg-gray-100 dark:bg-gray-600 text-gray-900 dark:text-white'
								: 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'} transition-colors"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
								/>
							</svg>
						</button>
						<button
							onclick={() => (viewMode = 'list')}
							class="px-3 py-1.5 text-sm font-medium rounded {viewMode === 'list'
								? 'bg-gray-100 dark:bg-gray-600 text-gray-900 dark:text-white'
								: 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'} transition-colors"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M4 6h16M4 12h16M4 18h16"
								/>
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Stats -->
			<div class="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
				<div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
					<p class="text-sm text-gray-600 dark:text-gray-400">Total</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-white">{stats.total}</p>
				</div>
				<div class="bg-yellow-50 dark:bg-yellow-900/30 rounded-lg p-4">
					<p class="text-sm text-yellow-700 dark:text-yellow-400">Generating</p>
					<p class="text-2xl font-bold text-yellow-900 dark:text-yellow-300">{stats.generating}</p>
				</div>
				<div class="bg-blue-50 dark:bg-blue-900/30 rounded-lg p-4">
					<p class="text-sm text-blue-700 dark:text-blue-400">Generated</p>
					<p class="text-2xl font-bold text-blue-900 dark:text-blue-300">{stats.generated}</p>
				</div>
				<div class="bg-green-50 dark:bg-green-900/30 rounded-lg p-4">
					<p class="text-sm text-green-700 dark:text-green-400">Deployed</p>
					<p class="text-2xl font-bold text-green-900 dark:text-green-300">{stats.deployed}</p>
				</div>
				<div class="bg-red-50 dark:bg-red-900/30 rounded-lg p-4">
					<p class="text-sm text-red-700 dark:text-red-400">Failed</p>
					<p class="text-2xl font-bold text-red-900 dark:text-red-300">{stats.failed}</p>
				</div>
			</div>

			<!-- Filters and Search -->
			<div class="flex flex-col sm:flex-row gap-4">
				<!-- Status Filter -->
				<div class="flex gap-2 flex-wrap">
					{#each filterOptions as option}
						<button
							onclick={() => handleFilterChange(option.value)}
							class="px-4 py-2 text-sm font-medium rounded-lg transition-colors {selectedFilter ===
							option.value
								? 'bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900'
								: 'bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600'}"
						>
							<span class="flex items-center gap-2">
								<span class="w-2 h-2 rounded-full {option.color}"></span>
								{option.label}
							</span>
						</button>
					{/each}
				</div>

				<!-- Search -->
				<div class="flex-1 max-w-md">
					<div class="relative">
						<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
							<svg
								class="w-5 h-5 text-gray-400"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
								/>
							</svg>
						</div>
						<input
							type="text"
							placeholder="Search apps..."
							value={searchQuery}
							oninput={handleSearchChange}
							class="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400"
						/>
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-auto">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			{#if error}
				<!-- Error State -->
				<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg p-6">
					<div class="flex items-start gap-3">
						<svg
							class="w-6 h-6 text-red-600 dark:text-red-400 flex-shrink-0"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<div>
							<h3 class="text-lg font-semibold text-red-900 dark:text-red-200">Error Loading Apps</h3>
							<p class="text-sm text-red-700 dark:text-red-300 mt-1">{error}</p>
							<button
								onclick={handleRefresh}
								class="btn-pill btn-pill-danger btn-pill-sm mt-3"
							>
								Try Again
							</button>
						</div>
					</div>
				</div>
			{:else if loading}
				<!-- Loading State -->
				<div class="flex items-center justify-center py-12">
					<div class="flex flex-col items-center gap-4">
						<svg
							class="w-12 h-12 text-blue-500 animate-spin"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
							/>
						</svg>
						<p class="text-gray-600 dark:text-gray-400">Loading apps...</p>
					</div>
				</div>
			{:else if apps.length === 0}
				<!-- Empty State -->
				<div class="text-center py-12">
					<svg
						class="w-16 h-16 text-gray-400 mx-auto mb-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
						/>
					</svg>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
						{selectedFilter === 'all' ? 'No apps found' : `No ${selectedFilter} apps`}
					</h3>
					<p class="text-gray-600 dark:text-gray-400 mb-6">
						{searchQuery
							? 'Try adjusting your search or filters'
							: 'Start by creating your first app with OSA'}
					</p>
					<div class="flex items-center justify-center gap-3">
						{#if searchQuery || selectedFilter !== 'all'}
							<button
								onclick={() => {
									selectedFilter = 'all';
									searchQuery = '';
									generatedAppsStore.setFilter('all');
									generatedAppsStore.setSearchQuery('');
								}}
								class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
							>
								Clear Filters
							</button>
						{:else}
							<button
								onclick={() => showCreateModal = true}
								class="px-6 py-3 text-sm font-medium text-white bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg hover:from-blue-700 hover:to-purple-700 transition-all flex items-center gap-2"
							>
								<Sparkles class="w-5 h-5" />
								Generate Your First App
							</button>
						{/if}
					</div>
				</div>
			{:else}
				<!-- Apps Grid/List -->
				<div
					class="{viewMode === 'grid'
						? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6'
						: 'flex flex-col gap-4'}"
				>
					{#each apps as app (app.id)}
						<GeneratedAppCard
							{app}
							onView={handleViewApp}
							onDeploy={handleDeployApp}
							onDelete={handleDeleteApp}
						/>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>

<CreateAppModal {workspaceId} bind:open={showCreateModal} />
