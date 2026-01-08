<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { useSession, clearSession } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import * as integrationsApi from '$lib/api/integrations';
	import type {
		IntegrationProviderInfo,
		UserIntegration,
		AIModelPreferences,
		PendingDecision,
		IntegrationCategory
	} from '$lib/api/integrations';

	const session = useSession();

	// State
	let isLoading = $state(true);
	let activeTab = $state<'connected' | 'available' | 'ai' | 'decisions'>('available');
	let hoveredId = $state<string | null>(null);
	let connectingId = $state<string | null>(null);
	let selectedProvider = $state<IntegrationProviderInfo | null>(null);
	let showDetailModal = $state(false);

	// Data
	let connectedIntegrations = $state<UserIntegration[]>([]);
	let availableProviders = $state<IntegrationProviderInfo[]>([]);
	let aiPreferences = $state<AIModelPreferences | null>(null);
	let pendingDecisions = $state<PendingDecision[]>([]);
	let selectedCategory = $state<IntegrationCategory | 'all'>('all');
	let isAuthenticated = $state(false);

	// Guards to prevent duplicate API calls
	let authDataLoading = $state(false);
	let authDataLoaded = $state(false);

	// AI providers that use file import instead of OAuth
	const fileImportProviders = ['chatgpt', 'claude', 'perplexity', 'gemini', 'granola'];

	// Category icons and labels
	const categories: { id: IntegrationCategory | 'all'; label: string; icon: string }[] = [
		{ id: 'all', label: 'All', icon: 'M4 6h16M4 12h16M4 18h16' },
		{ id: 'communication', label: 'Communication', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' },
		{ id: 'crm', label: 'CRM', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z' },
		{ id: 'tasks', label: 'Tasks', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4' },
		{ id: 'calendar', label: 'Calendar', icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'storage', label: 'Storage', icon: 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4' },
		{ id: 'meetings', label: 'Meetings', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
		{ id: 'ai', label: 'AI Assistants', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' }
	];

	// Category descriptions for integration info
	const categoryInfo: Record<string, { desc: string; features: string[] }> = {
		communication: { desc: 'Email and messaging integrations', features: ['Import conversations', 'Track threads', 'Send messages'] },
		crm: { desc: 'Customer relationship management', features: ['Sync contacts', 'Track deals', 'Manage pipelines'] },
		tasks: { desc: 'Task and project management', features: ['Sync tasks', 'Track progress', 'Bi-directional updates'] },
		calendar: { desc: 'Calendar and scheduling', features: ['Sync events', 'Track meetings', 'Auto-scheduling'] },
		storage: { desc: 'File storage and documents', features: ['Index files', 'Full-text search', 'Knowledge extraction'] },
		meetings: { desc: 'Video calls and recordings', features: ['Meeting summaries', 'Transcripts', 'Action items'] },
		ai: { desc: 'AI assistant exports', features: ['Import conversations', 'Knowledge extraction', 'Pattern learning'] },
		finance: { desc: 'Financial tools', features: ['Invoice tracking', 'Payment sync', 'Reports'] },
		code: { desc: 'Code repositories', features: ['PR tracking', 'Issue sync', 'Commit history'] }
	};

	// Sort and filter providers (ones with local logos first)
	let sortedProviders = $derived(
		[...availableProviders].sort((a, b) => {
			const aHasLocalLogo = a.icon_url?.startsWith('/logos/') ? 0 : 1;
			const bHasLocalLogo = b.icon_url?.startsWith('/logos/') ? 0 : 1;
			return aHasLocalLogo - bHasLocalLogo;
		})
	);

	// Filter providers by category
	let filteredProviders = $derived(
		selectedCategory === 'all'
			? sortedProviders
			: sortedProviders.filter((p) => p.category === selectedCategory)
	);

	// Check if provider is connected
	function isProviderConnected(providerId: string) {
		return connectedIntegrations.some(
			(i) => i.provider_id === providerId && i.status === 'connected'
		);
	}

	// Get connected integration for a provider
	function getConnectedIntegration(providerId: string) {
		return connectedIntegrations.find((i) => i.provider_id === providerId);
	}

	// Reactive auth check - updates when session changes
	$effect(() => {
		const sessionData = $session;
		// Only update if session is no longer pending
		if (!sessionData?.isPending) {
			isAuthenticated = !!sessionData?.data?.user;
		}
	});

	onMount(async () => {
		// Always load providers immediately (public endpoint)
		loadProviders();

		// Wait for session to resolve (give it up to 2 seconds)
		let attempts = 0;
		while ($session?.isPending && attempts < 20) {
			await new Promise((r) => setTimeout(r, 100));
			attempts++;
		}

		// Session resolved - check auth and load data
		const sessionData = $session;
		isAuthenticated = !sessionData?.isPending && !!sessionData?.data?.user;

		// Only load authenticated data if user is logged in AND we haven't already
		if (isAuthenticated && !authDataLoaded && !authDataLoading) {
			await loadAuthenticatedData();
		}
		isLoading = false;
	});

	async function loadProviders() {
		try {
			const providers = await integrationsApi.getProviders();
			availableProviders = providers.providers || [];
		} catch {
			availableProviders = [];
		}
	}

	async function loadAuthenticatedData() {
		// Guard against duplicate calls
		if (authDataLoading || authDataLoaded) {
			return;
		}
		authDataLoading = true;

		let authFailed = false;

		// Fetch connected integrations
		try {
			const connected = await integrationsApi.getConnectedIntegrations();
			connectedIntegrations = connected.integrations || [];
		} catch (e: unknown) {
			connectedIntegrations = [];
			// Check if this is a 401 error (session expired/invalid)
			if (e instanceof Error && e.message.includes('401')) {
				authFailed = true;
			}
		}

		// If first call got 401, clear session and don't make more authenticated calls
		if (authFailed) {
			clearSession();
			isAuthenticated = false;
			authDataLoading = false;
			isLoading = false;
			return;
		}

		// Fetch AI preferences
		try {
			const prefs = await integrationsApi.getAIModelPreferences();
			aiPreferences = prefs.preferences;
		} catch {
			aiPreferences = null;
		}

		// Fetch pending decisions
		try {
			const decisions = await integrationsApi.getPendingDecisions();
			pendingDecisions = decisions.decisions || [];
		} catch {
			pendingDecisions = [];
		}

		authDataLoading = false;
		authDataLoaded = true;
		isLoading = false;
	}

	async function loadData() {
		isLoading = true;
		// Reset guard to allow fresh load
		authDataLoaded = false;
		await loadProviders();
		if (isAuthenticated && !authDataLoading) {
			await loadAuthenticatedData();
		}
		isLoading = false;
	}

	function openProviderDetail(provider: IntegrationProviderInfo) {
		selectedProvider = provider;
		showDetailModal = true;
	}

	function closeDetailModal() {
		showDetailModal = false;
		selectedProvider = null;
	}

	async function handleConnect(provider: IntegrationProviderInfo) {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		// Check if this is a file import provider (AI assistants)
		if (fileImportProviders.includes(provider.id)) {
			// TODO: Open file import dialog
			alert(`File import for ${provider.name} coming soon. Export your data and upload here.`);
			return;
		}

		// Use oauth_provider if available (maps provider to OAuth endpoint, e.g., google_calendar -> google)
		// Fall back to provider.id if not mapped
		const oauthProvider = provider.oauth_provider || provider.id;
		if (!oauthProvider) {
			alert(`OAuth not configured for ${provider.name}. Please try again later.`);
			return;
		}

		console.log(`[handleConnect] Provider: ${provider.id}, OAuth Provider: ${oauthProvider}`);

		connectingId = provider.id;

		try {
			const response = await integrationsApi.initiateAuth(oauthProvider as integrationsApi.IntegrationProvider);
			if (response.auth_url) {
				window.open(response.auth_url, '_blank', 'width=600,height=700');
			}
		} catch (err) {
			console.error('Failed to initiate auth:', err);
			alert(`Failed to connect to ${provider.name}. Please try again.`);
		} finally {
			connectingId = null;
		}
	}

	async function handleDisconnect(integrationId: string) {
		try {
			await integrationsApi.disconnectUserIntegration(integrationId);
			await loadData();
		} catch (err) {
			console.error('Failed to disconnect:', err);
		}
	}

	async function handleDecision(decisionId: string, decision: string) {
		try {
			await integrationsApi.respondToDecision(decisionId, { decision });
			await loadData();
		} catch (err) {
			console.error('Failed to respond to decision:', err);
		}
	}

	function getStatusBadgeClass(status: string) {
		switch (status) {
			case 'connected':
				return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
			case 'available':
				return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400';
			case 'coming_soon':
				return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
			case 'error':
				return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}

	function getPriorityBadgeClass(priority: string) {
		switch (priority) {
			case 'urgent':
				return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
			case 'high':
				return 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400';
			case 'medium':
				return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
			default:
				return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
		}
	}
</script>

<svelte:head>
	<title>Integrations | BusinessOS</title>
</svelte:head>

<div class="h-full overflow-y-auto bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Integrations</h1>
					<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
						Connect your favorite tools and configure AI models
					</p>
				</div>
				{#if pendingDecisions.length > 0}
					<button
						onclick={() => (activeTab = 'decisions')}
						class="inline-flex items-center gap-2 px-4 py-2 bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-400 rounded-lg hover:bg-orange-200 dark:hover:bg-orange-900/50 transition-colors"
					>
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
						<span>{pendingDecisions.length} pending decisions</span>
					</button>
				{/if}
			</div>

			<!-- Tabs -->
			<div class="mt-6 flex gap-4 border-b border-gray-200 dark:border-gray-700 -mb-px">
				<button
					onclick={() => (activeTab = 'connected')}
					class="pb-3 px-1 font-medium text-sm transition-colors {activeTab === 'connected'
						? 'text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
						: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
				>
					Connected ({connectedIntegrations.length})
				</button>
				<button
					onclick={() => (activeTab = 'available')}
					class="pb-3 px-1 font-medium text-sm transition-colors {activeTab === 'available'
						? 'text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
						: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
				>
					Available ({availableProviders.length})
				</button>
				<button
					onclick={() => (activeTab = 'ai')}
					class="pb-3 px-1 font-medium text-sm transition-colors {activeTab === 'ai'
						? 'text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
						: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
				>
					AI Models
				</button>
				<button
					onclick={() => (activeTab = 'decisions')}
					class="pb-3 px-1 font-medium text-sm transition-colors relative {activeTab ===
					'decisions'
						? 'text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
						: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'}"
				>
					Decisions
					{#if pendingDecisions.length > 0}
						<span
							class="absolute -top-1 -right-2 w-5 h-5 flex items-center justify-center text-xs bg-orange-500 text-white rounded-full"
						>
							{pendingDecisions.length}
						</span>
					{/if}
				</button>
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		{#if isLoading}
			<div class="flex items-center justify-center h-64">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
			</div>
		{:else if activeTab === 'connected'}
			<!-- Connected Integrations -->
			{#if connectedIntegrations.length === 0}
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
							d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
						/>
					</svg>
					<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">
						No integrations connected
					</h3>
					<p class="mt-2 text-gray-500 dark:text-gray-400">
						Connect your favorite tools to get started.
					</p>
					<button
						onclick={() => (activeTab = 'available')}
						class="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
					>
						Browse Available Integrations
					</button>
				</div>
			{:else}
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each connectedIntegrations as integration}
						<div
							class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4"
						>
							<div class="flex items-start gap-3">
								<div
									class="w-10 h-10 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center"
								>
									{#if integration.icon_url}
										<img
											src={integration.icon_url}
											alt={integration.provider_name}
											class="w-6 h-6"
										/>
									{:else}
										<span class="text-lg font-medium text-gray-600 dark:text-gray-300">
											{integration.provider_name.charAt(0)}
										</span>
									{/if}
								</div>
								<div class="flex-1 min-w-0">
									<div class="flex items-center gap-2">
										<h3 class="font-medium text-gray-900 dark:text-white">
											{integration.provider_name}
										</h3>
										<span
											class="px-2 py-0.5 text-xs rounded-full {getStatusBadgeClass(
												integration.status
											)}"
										>
											{integration.status}
										</span>
									</div>
									<p class="text-sm text-gray-500 dark:text-gray-400 truncate">
										{integration.external_account_name ||
											integration.external_workspace_name ||
											'Connected'}
									</p>
								</div>
							</div>
							<div class="mt-4 flex gap-2">
								<button
									onclick={() => handleDisconnect(integration.id)}
									class="flex-1 px-3 py-1.5 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors"
								>
									Disconnect
								</button>
								<a
									href="/integrations/{integration.id}"
									class="flex-1 px-3 py-1.5 text-sm text-center text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors"
								>
									Settings
								</a>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{:else if activeTab === 'available'}
			<!-- Header Text -->
			<div class="mb-6">
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
					Let's bring all your data into a single place.
				</h2>
				<p class="text-sm text-gray-600 dark:text-gray-400 max-w-2xl">
					When you connect your apps, we will process raw data and extract essential information and turn it into nodes.
				</p>
			</div>

			<!-- Category Filter -->
			<div class="mb-6 flex gap-2 flex-wrap">
				{#each categories as category}
					<button
						onclick={() => (selectedCategory = category.id)}
						class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-colors {selectedCategory ===
						category.id
							? 'bg-gray-900 dark:bg-white text-white dark:text-gray-900'
							: 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'}"
					>
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={category.icon} />
						</svg>
						{category.label}
					</button>
				{/each}
			</div>

			<!-- Integrations Grid -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pb-8">
				{#each filteredProviders as provider}
					{@const isConnected = isProviderConnected(provider.id)}
					{@const isConnecting = connectingId === provider.id}
					{@const isComingSoon = provider.status === 'coming_soon'}
					<div
						class="relative bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 transition-all duration-200
							{isComingSoon ? 'opacity-60' : 'hover:shadow-lg hover:border-gray-300 dark:hover:border-gray-600'}
							{isConnecting ? 'border-blue-300 dark:border-blue-600 bg-blue-50/50 dark:bg-blue-900/10' : ''}"
						onmouseenter={() => hoveredId = provider.id}
						onmouseleave={() => hoveredId = null}
					>
						<!-- Tooltip -->
						{#if hoveredId === provider.id && provider.tooltip && !isComingSoon}
							<div
								class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 bg-gray-900 text-white text-xs px-3 py-2 rounded-lg max-w-[250px] text-center z-50 shadow-lg"
								transition:fade={{ duration: 150 }}
							>
								{provider.tooltip}
								<div class="absolute top-full left-1/2 -translate-x-1/2 w-0 h-0 border-l-6 border-r-6 border-t-6 border-l-transparent border-r-transparent border-t-gray-900"></div>
							</div>
						{/if}

						<!-- Card Header -->
						<div class="flex items-center justify-between mb-3">
							<div class="flex items-center gap-3">
								<!-- Icon -->
								<div
									class="w-8 h-8 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center overflow-hidden flex-shrink-0"
								>
									{#if provider.icon_url}
										<img
											src={provider.icon_url}
											alt={provider.name}
											class="w-5 h-5 object-contain"
											onerror={(e) => { const target = e.currentTarget as HTMLImageElement; target.style.display = 'none'; target.nextElementSibling?.classList.remove('hidden'); }}
										/>
										<span class="hidden text-sm font-semibold text-gray-600 dark:text-gray-300">
											{provider.name.charAt(0)}
										</span>
									{:else}
										<span class="text-sm font-semibold text-gray-600 dark:text-gray-300">
											{provider.name.charAt(0)}
										</span>
									{/if}
								</div>
								<!-- Name -->
								<span class="font-semibold text-gray-900 dark:text-white">{provider.name}</span>
								<!-- Auto Live-sync Badge -->
								{#if provider.auto_live_sync}
									<span class="inline-flex items-center gap-1 px-2 py-0.5 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 text-xs rounded">
										Auto Live-sync
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</span>
								{/if}
							</div>

							<!-- Status / Connect Button -->
							{#if isConnected}
								<span class="inline-flex items-center gap-1.5 px-3 py-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-green-600 dark:text-green-400 text-sm rounded-full">
									<span class="w-1.5 h-1.5 bg-green-500 rounded-full"></span>
									Live-Synced
								</span>
							{:else if isComingSoon}
								<span class="px-3 py-1 bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 text-sm rounded-full">
									Soon
								</span>
							{:else if isConnecting}
								<span class="inline-flex items-center gap-1.5 px-3 py-1 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-700 text-blue-600 dark:text-blue-400 text-sm rounded-full">
									<span class="w-3 h-3 border-2 border-blue-400 border-t-transparent rounded-full animate-spin"></span>
									Connecting...
								</span>
							{:else}
								<button
									onclick={() => handleConnect(provider)}
									class="px-4 py-1.5 bg-gray-900 dark:bg-white text-white dark:text-gray-900 text-sm font-medium rounded-full hover:bg-gray-800 dark:hover:bg-gray-100 transition-colors"
								>
									{fileImportProviders.includes(provider.id) ? 'Import' : 'Connect'}
								</button>
							{/if}
						</div>

						<!-- Description -->
						<p class="text-sm text-gray-600 dark:text-gray-400 mb-4 line-clamp-2">
							{provider.description || `Connect your ${provider.name} account`}
						</p>

						<!-- Stats Footer -->
						{#if provider.est_nodes || provider.initial_sync}
							<div class="flex flex-col gap-1.5 pt-3 border-t border-gray-100 dark:border-gray-700">
								{#if provider.est_nodes}
									<div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
										</svg>
										<span class="flex-1">{isConnected ? 'Tot. nodes' : 'Est. nodes'}</span>
										<span class="font-medium text-gray-700 dark:text-gray-300">{provider.est_nodes}</span>
									</div>
								{/if}
								{#if provider.initial_sync}
									<div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
										<span class="flex-1">Initial sync</span>
										<span class="font-medium text-gray-700 dark:text-gray-300">{provider.initial_sync}</span>
									</div>
								{/if}
							</div>
						{/if}

						<!-- Learn More Link -->
						<button
							onclick={() => openProviderDetail(provider)}
							class="mt-3 text-xs text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 hover:underline flex items-center gap-1"
						>
							Learn more
							<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
							</svg>
						</button>
					</div>
				{/each}
			</div>
		{:else if activeTab === 'ai'}
			<!-- AI Model Preferences -->
			<div class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
					AI Model Configuration
				</h2>
				<p class="text-gray-500 dark:text-gray-400 mb-6">
					Configure which AI models to use for different task tiers. The system automatically selects
					the appropriate tier based on task complexity.
				</p>

				{#if aiPreferences}
					<div class="space-y-6">
						<!-- Tier 2 -->
						<div class="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
							<h3 class="font-medium text-gray-900 dark:text-white">
								Tier 2: Fast Tasks
							</h3>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-3">
								Quick, low-complexity operations like formatting and simple lookups.
							</p>
							<div class="flex items-center gap-4">
								<span class="text-sm text-gray-600 dark:text-gray-300">
									{aiPreferences.tier_2_model.provider}: {aiPreferences.tier_2_model.model_id}
								</span>
							</div>
						</div>

						<!-- Tier 3 -->
						<div class="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
							<h3 class="font-medium text-gray-900 dark:text-white">
								Tier 3: Standard Tasks
							</h3>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-3">
								Medium-complexity tasks requiring analysis and synthesis.
							</p>
							<div class="flex items-center gap-4">
								<span class="text-sm text-gray-600 dark:text-gray-300">
									{aiPreferences.tier_3_model.provider}: {aiPreferences.tier_3_model.model_id}
								</span>
							</div>
						</div>

						<!-- Tier 4 -->
						<div class="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
							<h3 class="font-medium text-gray-900 dark:text-white">
								Tier 4: Complex Tasks
							</h3>
							<p class="text-sm text-gray-500 dark:text-gray-400 mb-3">
								High-complexity tasks requiring deep reasoning and multi-step analysis.
							</p>
							<div class="flex items-center gap-4">
								<span class="text-sm text-gray-600 dark:text-gray-300">
									{aiPreferences.tier_4_model.provider}: {aiPreferences.tier_4_model.model_id}
								</span>
							</div>
						</div>

						<!-- Settings -->
						<div class="border-t border-gray-200 dark:border-gray-600 pt-4">
							<h3 class="font-medium text-gray-900 dark:text-white mb-3">Settings</h3>
							<div class="space-y-2">
								<label class="flex items-center gap-2">
									<input
										type="checkbox"
										checked={aiPreferences.allow_model_upgrade_on_failure}
										class="rounded border-gray-300 dark:border-gray-600"
									/>
									<span class="text-sm text-gray-600 dark:text-gray-300">
										Allow automatic model upgrade on failure
									</span>
								</label>
								<label class="flex items-center gap-2">
									<input
										type="checkbox"
										checked={aiPreferences.prefer_local}
										class="rounded border-gray-300 dark:border-gray-600"
									/>
									<span class="text-sm text-gray-600 dark:text-gray-300">
										Prefer local models when available
									</span>
								</label>
								<div class="flex items-center gap-2">
									<span class="text-sm text-gray-600 dark:text-gray-300">Max latency:</span>
									<span class="text-sm font-mono text-gray-900 dark:text-white">
										{aiPreferences.max_latency_ms}ms
									</span>
								</div>
							</div>
						</div>
					</div>
				{:else}
					<p class="text-gray-500 dark:text-gray-400">
						AI preferences not configured. Default settings will be used.
					</p>
				{/if}
			</div>
		{:else if activeTab === 'decisions'}
			<!-- Pending Decisions -->
			{#if pendingDecisions.length === 0}
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
							d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">
						No pending decisions
					</h3>
					<p class="mt-2 text-gray-500 dark:text-gray-400">
						When AI agents need your input, decisions will appear here.
					</p>
				</div>
			{:else}
				<div class="space-y-4">
					{#each pendingDecisions as decision}
						<div
							class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4"
						>
							<div class="flex items-start justify-between">
								<div>
									<div class="flex items-center gap-2">
										<h3 class="font-medium text-gray-900 dark:text-white">{decision.question}</h3>
										<span class="px-2 py-0.5 text-xs rounded-full {getPriorityBadgeClass(decision.priority)}">
											{decision.priority}
										</span>
									</div>
									{#if decision.description}
										<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
											{decision.description}
										</p>
									{/if}
									<p class="mt-2 text-xs text-gray-400 dark:text-gray-500">
										Skill: {decision.skill_id} | Created: {new Date(decision.created_at).toLocaleString()}
									</p>
								</div>
							</div>
							{#if decision.options && decision.options.length > 0}
								<div class="mt-4 flex flex-wrap gap-2">
									{#each decision.options as option}
										<button
											onclick={() => handleDecision(decision.id, option)}
											class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
										>
											{option}
										</button>
									{/each}
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		{/if}
	</div>
</div>

<!-- Integration Detail Modal -->
{#if showDetailModal && selectedProvider}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
		onclick={closeDetailModal}
		transition:fade={{ duration: 200 }}
	>
		<div
			class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-lg w-full max-h-[85vh] overflow-hidden"
			onclick={(e) => e.stopPropagation()}
			transition:slide={{ duration: 200 }}
		>
			<!-- Header -->
			<div class="p-6 border-b border-gray-200 dark:border-gray-700">
				<div class="flex items-start justify-between">
					<div class="flex items-center gap-4">
						<div class="w-14 h-14 rounded-xl bg-gray-100 dark:bg-gray-700 flex items-center justify-center overflow-hidden">
							{#if selectedProvider.icon_url}
								<img src={selectedProvider.icon_url} alt={selectedProvider.name} class="w-9 h-9 object-contain" />
							{:else}
								<span class="text-xl font-bold text-gray-600 dark:text-gray-300">{selectedProvider.name.charAt(0)}</span>
							{/if}
						</div>
						<div>
							<h2 class="text-xl font-bold text-gray-900 dark:text-white">{selectedProvider.name}</h2>
							<span class="inline-flex items-center gap-1 px-2 py-0.5 mt-1 text-xs font-medium rounded-full bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 capitalize">
								{selectedProvider.category}
							</span>
						</div>
					</div>
					<button
						onclick={closeDetailModal}
						class="p-2 text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Body -->
			<div class="p-6 overflow-y-auto max-h-[50vh]">
				<!-- Description -->
				<p class="text-gray-600 dark:text-gray-400 mb-6">
					{selectedProvider.description || `Connect your ${selectedProvider.name} account to sync data and enable powerful automations.`}
				</p>

				<!-- Category Features -->
				{#if categoryInfo[selectedProvider.category]}
					<div class="mb-6">
						<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">What you can do</h3>
						<ul class="space-y-2">
							{#each categoryInfo[selectedProvider.category].features as feature}
								<li class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
									<svg class="w-4 h-4 text-green-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									{feature}
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				<!-- Sync Info -->
				<div class="bg-gray-50 dark:bg-gray-700/50 rounded-xl p-4 mb-6">
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">Sync details</h3>
					<div class="grid grid-cols-2 gap-4">
						{#if selectedProvider.auto_live_sync}
							<div class="flex items-center gap-2">
								<div class="w-8 h-8 rounded-lg bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
									<svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
									</svg>
								</div>
								<div>
									<div class="text-xs text-gray-500 dark:text-gray-400">Sync type</div>
									<div class="text-sm font-medium text-gray-900 dark:text-white">Live sync</div>
								</div>
							</div>
						{:else}
							<div class="flex items-center gap-2">
								<div class="w-8 h-8 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
									<svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
									</svg>
								</div>
								<div>
									<div class="text-xs text-gray-500 dark:text-gray-400">Sync type</div>
									<div class="text-sm font-medium text-gray-900 dark:text-white">Manual/Scheduled</div>
								</div>
							</div>
						{/if}

						{#if selectedProvider.est_nodes}
							<div class="flex items-center gap-2">
								<div class="w-8 h-8 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center">
									<svg class="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
									</svg>
								</div>
								<div>
									<div class="text-xs text-gray-500 dark:text-gray-400">Est. nodes</div>
									<div class="text-sm font-medium text-gray-900 dark:text-white">{selectedProvider.est_nodes}</div>
								</div>
							</div>
						{/if}

						{#if selectedProvider.initial_sync}
							<div class="flex items-center gap-2">
								<div class="w-8 h-8 rounded-lg bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center">
									<svg class="w-4 h-4 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								</div>
								<div>
									<div class="text-xs text-gray-500 dark:text-gray-400">Initial sync</div>
									<div class="text-sm font-medium text-gray-900 dark:text-white">{selectedProvider.initial_sync}</div>
								</div>
							</div>
						{/if}
					</div>
				</div>

				<!-- Skills -->
				{#if selectedProvider.skills && selectedProvider.skills.length > 0}
					<div class="mb-6">
						<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">Available skills</h3>
						<div class="flex flex-wrap gap-2">
							{#each selectedProvider.skills as skill}
								<span class="px-2 py-1 text-xs font-mono bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded">
									{skill}
								</span>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Modules -->
				{#if selectedProvider.modules && selectedProvider.modules.length > 0}
					<div>
						<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-3">Works with</h3>
						<div class="flex flex-wrap gap-2">
							{#each selectedProvider.modules as module}
								<span class="inline-flex items-center gap-1 px-2 py-1 text-xs bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 rounded capitalize">
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
									</svg>
									{module.replace('_', ' ')}
								</span>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="p-6 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50">
				{#if isProviderConnected(selectedProvider.id)}
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2 text-green-600 dark:text-green-400">
							<span class="w-2 h-2 bg-green-500 rounded-full"></span>
							<span class="text-sm font-medium">Connected</span>
							{#if getConnectedIntegration(selectedProvider.id)?.external_account_name}
								<span class="text-sm text-gray-500 dark:text-gray-400">as {getConnectedIntegration(selectedProvider.id)?.external_account_name}</span>
							{/if}
						</div>
						<div class="flex items-center gap-2">
							<!-- Auto-sync toggle (if supported) -->
							{#if selectedProvider.auto_live_sync}
								<label class="flex items-center gap-2 cursor-pointer">
									<span class="text-sm text-gray-600 dark:text-gray-400">Auto-sync</span>
									<div class="relative">
										<input type="checkbox" class="sr-only peer" checked />
										<div class="w-10 h-5 bg-gray-200 dark:bg-gray-600 rounded-full peer peer-checked:bg-green-500 transition-colors"></div>
										<div class="absolute left-0.5 top-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-5"></div>
									</div>
								</label>
							{/if}
							<a
								href="/integrations/{getConnectedIntegration(selectedProvider.id)?.id}"
								class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
							>
								Settings
							</a>
						</div>
					</div>
				{:else if selectedProvider.status === 'coming_soon'}
					<button
						disabled
						class="w-full px-4 py-2.5 text-sm font-medium text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 rounded-lg cursor-not-allowed"
					>
						Coming Soon
					</button>
				{:else}
					<button
						onclick={() => { if (selectedProvider) { closeDetailModal(); handleConnect(selectedProvider); } }}
						class="w-full px-4 py-2.5 text-sm font-medium text-white bg-gray-900 dark:bg-white dark:text-gray-900 rounded-lg hover:bg-gray-800 dark:hover:bg-gray-100 transition-colors"
					>
						{selectedProvider && fileImportProviders.includes(selectedProvider.id) ? 'Import Data' : 'Connect'}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}
