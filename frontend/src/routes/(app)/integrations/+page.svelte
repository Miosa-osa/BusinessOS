<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, slide } from 'svelte/transition';
	import { useSession, clearSession } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
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

	// File import state
	let showFileImportModal = $state(false);
	let fileImportProvider = $state<IntegrationProviderInfo | null>(null);
	let fileImportFile = $state<File | null>(null);
	let fileImporting = $state(false);
	let fileImportError = $state<string | null>(null);
	let fileImportSuccess = $state<string | null>(null);
	let fileInputRef = $state<HTMLInputElement | null>(null);

	// AI prefs save state
	let savingAiPrefs = $state(false);
	let aiPrefsMessage = $state<string | null>(null);
	let aiPrefsError = $state<string | null>(null);

	// Sync state for connected cards
	let syncingId = $state<string | null>(null);

	// Search filter for providers
	let searchQuery = $state('');

	// Decisions error state
	let decisionsError = $state<string | null>(null);

	// AI providers that use file import instead of OAuth
	const fileImportProviders = ['chatgpt', 'claude', 'perplexity', 'gemini', 'granola'];

	// Cleanup for OAuth message listener
	let oauthMessageCleanup: (() => void) | null = null;
	onDestroy(() => oauthMessageCleanup?.());

	// Category icons and labels
	const categories: { id: IntegrationCategory | 'all'; label: string; icon: string }[] = [
		{ id: 'all', label: 'All', icon: 'M4 6h16M4 12h16M4 18h16' },
		{ id: 'communication', label: 'Communication', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' },
		{ id: 'crm', label: 'CRM', icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z' },
		{ id: 'tasks', label: 'Tasks', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4' },
		{ id: 'calendar', label: 'Calendar', icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'storage', label: 'Storage', icon: 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4' },
		{ id: 'meetings', label: 'Meetings', icon: 'M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
		{ id: 'ai', label: 'AI Assistants', icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ id: 'custom', label: 'Custom', icon: 'M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z' },
		{ id: 'productivity', label: 'Productivity', icon: 'M13 10V3L4 14h7v7l9-11h-7z' }
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

	// Filter providers by category and search
	let filteredProviders = $derived.by(() => {
		let result = selectedCategory === 'all'
			? sortedProviders
			: sortedProviders.filter((p) => p.category === selectedCategory);
		if (searchQuery.trim()) {
			const q = searchQuery.trim().toLowerCase();
			result = result.filter((p) => p.name.toLowerCase().includes(q));
		}
		return result;
	});

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
		// Handle OAuth callback: if this page was opened in a popup after OAuth redirect
		if (browser) {
			const urlParams = new URLSearchParams(window.location.search);
			const connectedProvider = urlParams.get('connected');
			if (connectedProvider && window.opener) {
				// We're in the OAuth popup — notify parent and close
				window.opener.postMessage({ type: 'integration-connected', provider: connectedProvider }, window.location.origin);
				window.close();
				return;
			}
			if (connectedProvider) {
				// Direct navigation with ?connected= (popup blocked or manual redirect)
				activeTab = 'connected';
				// Clean up the URL
				const url = new URL(window.location.href);
				url.searchParams.delete('connected');
				window.history.replaceState({}, '', url.toString());
			}
		}

		// Listen for OAuth popup completion messages
		function handleOAuthMessage(event: MessageEvent) {
			if (event.origin !== window.location.origin) return;
			if (event.data?.type === 'integration-connected') {
				// Popup completed OAuth — refresh data
				loadData();
				activeTab = 'connected';
			}
		}
		window.addEventListener('message', handleOAuthMessage);
		oauthMessageCleanup = () => window.removeEventListener('message', handleOAuthMessage);

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
			decisionsError = null;
		} catch {
			pendingDecisions = [];
			decisionsError = 'Failed to load pending decisions';
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

	async function saveAiPreferences(updates: Partial<AIModelPreferences>) {
		if (!aiPreferences) return;
		savingAiPrefs = true;
		aiPrefsError = null;
		aiPrefsMessage = null;
		try {
			await integrationsApi.updateAIModelPreferences({ ...aiPreferences, ...updates });
			Object.assign(aiPreferences, updates);
			aiPrefsMessage = 'Preferences saved';
			setTimeout(() => (aiPrefsMessage = null), 3000);
		} catch (err) {
			console.error('Failed to save AI preferences:', err);
			aiPrefsError = 'Failed to save preferences';
			setTimeout(() => (aiPrefsError = null), 5000);
		} finally {
			savingAiPrefs = false;
		}
	}

	async function handleSyncCard(integrationId: string) {
		syncingId = integrationId;
		try {
			await integrationsApi.triggerIntegrationSync(integrationId);
			await loadData();
		} catch (err) {
			console.error('Failed to sync:', err);
		} finally {
			syncingId = null;
		}
	}

	async function handleConnect(provider: IntegrationProviderInfo) {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		// Check if this is a file import provider (AI assistants)
		if (fileImportProviders.includes(provider.id)) {
			fileImportProvider = provider;
			fileImportFile = null;
			fileImportError = null;
			fileImportSuccess = null;
			showFileImportModal = true;
			return;
		}

		// Use oauth_provider if available (maps provider to OAuth endpoint, e.g., google_calendar -> google)
		// Fall back to provider.id if not mapped
		const oauthProvider = provider.oauth_provider || provider.id;
		if (!oauthProvider) {
			alert(`OAuth not configured for ${provider.name}. Please try again later.`);
			return;
		}

		if (import.meta.env.DEV) console.log(`[handleConnect] Provider: ${provider.id}, OAuth Provider: ${oauthProvider}`);

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

	function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		if (target.files && target.files.length > 0) {
			fileImportFile = target.files[0];
			fileImportError = null;
		}
	}

	async function handleFileImport() {
		if (!fileImportFile || !fileImportProvider) return;
		fileImporting = true;
		fileImportError = null;
		fileImportSuccess = null;
		try {
			const source = (fileImportProvider.id === 'granola' ? 'other' : fileImportProvider.id) as 'chatgpt' | 'claude' | 'perplexity' | 'gemini' | 'other';
			const result = await integrationsApi.importFile(fileImportFile, source);
			fileImportSuccess = result.message || `Successfully imported ${result.imported_count} items.`;
			fileImportFile = null;
			await loadData();
		} catch (err) {
			fileImportError = err instanceof Error ? err.message : 'Import failed. Please try again.';
		} finally {
			fileImporting = false;
		}
	}

	function closeFileImportModal() {
		showFileImportModal = false;
		fileImportProvider = null;
		fileImportFile = null;
		fileImportError = null;
		fileImportSuccess = null;
	}

	function getStatusBadgeClass(status: string) {
		switch (status) {
			case 'connected': return 'ih-badge--connected';
			case 'available': return 'ih-badge--available';
			case 'coming_soon': return 'ih-badge--coming-soon';
			case 'error': return 'ih-badge--error';
			default: return 'ih-badge--default';
		}
	}

	function getPriorityBadgeClass(priority: string) {
		switch (priority) {
			case 'urgent': return 'ih-priority--urgent';
			case 'high': return 'ih-priority--high';
			case 'medium': return 'ih-priority--medium';
			default: return 'ih-priority--default';
		}
	}
</script>

<svelte:head>
	<title>Integrations | BusinessOS</title>
</svelte:head>

<div class="ih-page">
	<!-- Header -->
	<div class="ih-header">
		<div class="ih-header__inner">
			<div class="ih-header__top">
				<div>
					<h1 class="ih-header__title">Integrations</h1>
					<p class="ih-header__subtitle">
						Connect your favorite tools and configure AI models
					</p>
				</div>
				{#if pendingDecisions.length > 0}
					<button
						onclick={() => (activeTab = 'decisions')}
						class="ih-decisions-alert"
					>
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
						<span>{pendingDecisions.length} pending decisions</span>
					</button>
				{/if}
			</div>

			<!-- Tabs -->
			<div class="ih-tabs">
				<button
					onclick={() => (activeTab = 'connected')}
					class="ih-tab {activeTab === 'connected' ? 'ih-tab--active' : ''}"
				>
					Connected ({connectedIntegrations.length})
				</button>
				<button
					onclick={() => (activeTab = 'available')}
					class="ih-tab {activeTab === 'available' ? 'ih-tab--active' : ''}"
				>
					Available ({availableProviders.length})
				</button>
				<button
					onclick={() => (activeTab = 'ai')}
					class="ih-tab {activeTab === 'ai' ? 'ih-tab--active' : ''}"
				>
					AI Models
				</button>
				<button
					onclick={() => (activeTab = 'decisions')}
					class="ih-tab {activeTab === 'decisions' ? 'ih-tab--active' : ''}"
				>
					Decisions
					{#if pendingDecisions.length > 0}
						<span class="ih-tab__count">{pendingDecisions.length}</span>
					{/if}
				</button>
			</div>
		</div>
	</div>

	<!-- Content -->
	<div class="ih-content">
		{#key activeTab}
		<div in:fade={{ duration: 150 }}>
		{#if isLoading}
			<div class="ih-spinner-wrap">
				<div class="ih-spinner"></div>
			</div>
		{:else if activeTab === 'connected'}
			<!-- Connected Integrations -->
			{#if connectedIntegrations.length === 0}
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
							d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
						/>
					</svg>
					<h3 class="ih-empty__title">No integrations connected</h3>
					<p class="ih-empty__text">Connect your favorite tools to get started.</p>
					<button
						onclick={() => (activeTab = 'available')}
						class="btn-pill btn-pill-primary btn-pill-sm mt-4"
					>
						Browse Available Integrations
					</button>
				</div>
			{:else}
				<div class="ih-grid">
					{#each connectedIntegrations as integration}
						<div class="ih-card">
							<div class="ih-card__header">
								<div class="ih-card__icon-wrap">
									{#if integration.icon_url}
										<img
											src={integration.icon_url}
											alt={integration.provider_name}
											class="w-6 h-6"
										/>
									{:else}
										<span class="ih-card__icon-letter">
											{integration.provider_name.charAt(0)}
										</span>
									{/if}
								</div>
								<div class="ih-card__info">
									<div class="ih-card__name-row">
										<h3 class="ih-card__name">{integration.provider_name}</h3>
										<span class="ih-badge {getStatusBadgeClass(integration.status)}">
											{integration.status}
										</span>
									</div>
									<p class="ih-card__meta">
										{integration.external_account_name ||
											integration.external_workspace_name ||
											'Connected'}
									</p>
									{#if integration.last_used_at}
										<p class="ih-card__sub-meta">
											Last used {new Date(integration.last_used_at).toLocaleDateString()}
										</p>
									{/if}
								</div>
							</div>
							<div class="ih-card__actions">
								<button
									onclick={() => handleSyncCard(integration.id)}
									disabled={syncingId === integration.id}
									class="ih-card__sync-btn"
									title="Sync now"
								>
									{#if syncingId === integration.id}
										<svg class="w-4 h-4 ih-spinner--inline" viewBox="0 0 24 24">
											<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
											<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
										</svg>
									{:else}
										<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
										</svg>
									{/if}
								</button>
								<a
									href="/integrations/{integration.id}"
									class="ih-card__settings-link"
									title="Configure"
								>
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
									</svg>
								</a>
								<button
									onclick={() => handleDisconnect(integration.id)}
									class="btn-pill btn-pill-danger btn-pill-sm ih-card__actions-btn"
								>
									Disconnect
								</button>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{:else if activeTab === 'available'}
			<!-- Header Text -->
			<div class="ih-section-intro">
				<h2 class="ih-section-intro__title">
					Let's bring all your data into a single place.
				</h2>
				<p class="ih-section-intro__text">
					When you connect your apps, we will process raw data and extract essential information and turn it into nodes.
				</p>
			</div>

			<!-- Category Filter -->
			<div class="ih-category-filter">
				<div class="ih-search-wrap">
					<svg class="w-4 h-4 ih-search-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						placeholder="Search integrations..."
						bind:value={searchQuery}
						class="ih-search-input"
					/>
				</div>
				{#each categories as category}
					<button
						onclick={() => (selectedCategory = category.id)}
						class="ih-category-btn {selectedCategory === category.id ? 'ih-category-btn--active' : ''}"
					>
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={category.icon} />
						</svg>
						{category.label}
					</button>
				{/each}
			</div>

			<!-- Integrations Grid -->
			<div class="ih-grid ih-grid--pb">
				{#each filteredProviders as provider}
					{@const isConnected = isProviderConnected(provider.id)}
					{@const isConnecting = connectingId === provider.id}
					{@const isComingSoon = provider.status === 'coming_soon'}
					<div
						class="ih-provider-card
							{isComingSoon ? 'ih-provider-card--soon' : ''}
							{isConnecting ? 'ih-provider-card--connecting' : ''}"
						onmouseenter={() => hoveredId = provider.id}
						onmouseleave={() => hoveredId = null}
					>
						<!-- Tooltip -->
						{#if hoveredId === provider.id && provider.tooltip && !isComingSoon}
							<div
								class="ih-tooltip"
								transition:fade={{ duration: 150 }}
							>
								{provider.tooltip}
								<div class="ih-tooltip__arrow"></div>
							</div>
						{/if}

						<!-- Card Header -->
						<div class="ih-provider-card__header">
							<div class="ih-provider-card__left">
								<!-- Icon -->
								<div class="ih-provider-card__icon">
									{#if provider.icon_url}
										<img
											src={provider.icon_url}
											alt={provider.name}
											class="w-5 h-5 object-contain"
											onerror={(e) => { const target = e.currentTarget as HTMLImageElement; target.style.display = 'none'; target.nextElementSibling?.classList.remove('hidden'); }}
										/>
										<span class="hidden ih-card__icon-letter--sm">
											{provider.name.charAt(0)}
										</span>
									{:else}
										<span class="ih-card__icon-letter--sm">
											{provider.name.charAt(0)}
										</span>
									{/if}
								</div>
								<!-- Name -->
								<span class="ih-provider-card__name">{provider.name}</span>
								<!-- Auto Live-sync Badge -->
								{#if provider.auto_live_sync}
									<span class="ih-autosync-badge">
										Auto Live-sync
										<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
									</span>
								{/if}
							</div>

							<!-- Status / Connect Button -->
							{#if isConnected}
								<span class="ih-status-pill ih-status-pill--connected">
									<span class="ih-status-dot ih-status-dot--green"></span>
									Live-Synced
								</span>
							{:else if isComingSoon}
								<span class="ih-status-pill ih-status-pill--soon">Soon</span>
							{:else if isConnecting}
								<span class="ih-status-pill ih-status-pill--connecting">
									<span class="ih-spinner ih-spinner--sm"></span>
									Connecting...
								</span>
							{:else}
								<button
									onclick={() => handleConnect(provider)}
									class="btn-pill btn-pill-primary btn-pill-sm"
								>
									{fileImportProviders.includes(provider.id) ? 'Import' : 'Connect'}
								</button>
							{/if}
						</div>

						<!-- Description -->
						<p class="ih-provider-card__desc">
							{provider.description || `Connect your ${provider.name} account`}
						</p>

						<!-- Stats Footer -->
						{#if provider.est_nodes || provider.initial_sync}
							<div class="ih-provider-card__stats">
								{#if provider.est_nodes}
									<div class="ih-stat-row">
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
										</svg>
										<span class="ih-stat-row__label">{isConnected ? 'Tot. nodes' : 'Est. nodes'}</span>
										<span class="ih-stat-row__value">{provider.est_nodes}</span>
									</div>
								{/if}
								{#if provider.initial_sync}
									<div class="ih-stat-row">
										<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
										<span class="ih-stat-row__label">Initial sync</span>
										<span class="ih-stat-row__value">{provider.initial_sync}</span>
									</div>
								{/if}
							</div>
						{/if}

						<!-- Learn More Link -->
						<button
							onclick={() => openProviderDetail(provider)}
							class="ih-learn-more"
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
			<div class="ih-section">
				<h2 class="ih-section__title">AI Model Configuration</h2>
				<p class="ih-section__desc">
					Configure which AI models to use for different task tiers. The system automatically selects
					the appropriate tier based on task complexity.
				</p>

				{#if aiPreferences}
					<div class="ih-tier-list">
						<!-- Tier 2 -->
						<div class="ih-tier">
							<h3 class="ih-tier__name">Tier 2: Fast Tasks</h3>
							<p class="ih-tier__desc">
								Quick, low-complexity operations like formatting and simple lookups.
							</p>
							<div class="ih-tier__model">
								<span>
									{aiPreferences.tier_2_model.provider}: {aiPreferences.tier_2_model.model_id}
								</span>
							</div>
						</div>

						<!-- Tier 3 -->
						<div class="ih-tier">
							<h3 class="ih-tier__name">Tier 3: Standard Tasks</h3>
							<p class="ih-tier__desc">
								Medium-complexity tasks requiring analysis and synthesis.
							</p>
							<div class="ih-tier__model">
								<span>
									{aiPreferences.tier_3_model.provider}: {aiPreferences.tier_3_model.model_id}
								</span>
							</div>
						</div>

						<!-- Tier 4 -->
						<div class="ih-tier">
							<h3 class="ih-tier__name">Tier 4: Complex Tasks</h3>
							<p class="ih-tier__desc">
								High-complexity tasks requiring deep reasoning and multi-step analysis.
							</p>
							<div class="ih-tier__model">
								<span>
									{aiPreferences.tier_4_model.provider}: {aiPreferences.tier_4_model.model_id}
								</span>
							</div>
						</div>

						<!-- Settings -->
						<div class="ih-ai-settings">
							<h3 class="ih-ai-settings__title">Settings</h3>
							{#if aiPrefsMessage}
								<div class="ih-alert ih-alert--success ih-alert--sm">
									<p>{aiPrefsMessage}</p>
								</div>
							{/if}
							{#if aiPrefsError}
								<div class="ih-alert ih-alert--error ih-alert--sm">
									<p>{aiPrefsError}</p>
								</div>
							{/if}
							<div class="ih-ai-settings__list">
								<label class="ih-checkbox-label">
									<input
										type="checkbox"
										checked={aiPreferences.allow_model_upgrade_on_failure}
										onchange={(e) => {
											const target = e.currentTarget as HTMLInputElement;
											saveAiPreferences({ allow_model_upgrade_on_failure: target.checked });
										}}
										disabled={savingAiPrefs}
										class="ih-checkbox"
									/>
									<span>Allow automatic model upgrade on failure</span>
								</label>
								<label class="ih-checkbox-label">
									<input
										type="checkbox"
										checked={aiPreferences.prefer_local}
										onchange={(e) => {
											const target = e.currentTarget as HTMLInputElement;
											saveAiPreferences({ prefer_local: target.checked });
										}}
										disabled={savingAiPrefs}
										class="ih-checkbox"
									/>
									<span>Prefer local models when available</span>
								</label>
								<div class="ih-latency-row">
									<span>Max latency:</span>
									<input
										type="number"
										value={aiPreferences.max_latency_ms}
										onchange={(e) => {
											const target = e.currentTarget as HTMLInputElement;
											const val = parseInt(target.value, 10);
											if (!isNaN(val) && val > 0) {
												saveAiPreferences({ max_latency_ms: val });
											}
										}}
										disabled={savingAiPrefs}
										class="ih-latency-input"
										min="100"
										step="100"
									/>
									<span class="ih-latency-unit">ms</span>
								</div>
							</div>
						</div>
					</div>
				{:else}
					<p class="ih-empty__text">
						AI preferences not configured. Default settings will be used.
					</p>
				{/if}
			</div>
		{:else if activeTab === 'decisions'}
			<!-- Pending Decisions -->
			{#if decisionsError}
				<div class="ih-alert ih-alert--error ih-alert--banner">
					<p>{decisionsError}</p>
					<button onclick={loadData} class="btn-pill btn-pill-ghost btn-pill-sm">Retry</button>
				</div>
			{/if}
			{#if pendingDecisions.length === 0 && !decisionsError}
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
							d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<h3 class="ih-empty__title">No pending decisions</h3>
					<p class="ih-empty__text">
						When AI agents need your input, decisions will appear here.
					</p>
				</div>
			{:else}
				<div class="ih-decision-list">
					{#each pendingDecisions as decision}
						<div class="ih-card">
							<div class="ih-decision__top">
								<div>
									<div class="ih-decision__header">
										<h3 class="ih-card__name">{decision.question}</h3>
										<span class="ih-priority {getPriorityBadgeClass(decision.priority)}">
											{decision.priority}
										</span>
									</div>
									{#if decision.description}
										<p class="ih-decision__desc">{decision.description}</p>
									{/if}
									<p class="ih-decision__meta">
										Skill: {decision.skill_id} | Created: {new Date(decision.created_at).toLocaleString()}
									</p>
								</div>
							</div>
							{#if decision.options && decision.options.length > 0}
								<div class="ih-decision__options">
									{#each decision.options as option}
										<button
											onclick={() => handleDecision(decision.id, option)}
											class="btn-pill btn-pill-primary btn-pill-sm"
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
		{/key}
	</div>
</div>

<!-- Integration Detail Modal -->
{#if showDetailModal && selectedProvider}
	<div
		class="ih-modal-backdrop"
		onclick={closeDetailModal}
		transition:fade={{ duration: 200 }}
	>
		<div
			class="ih-modal"
			onclick={(e) => e.stopPropagation()}
			transition:slide={{ duration: 200 }}
		>
			<!-- Header -->
			<div class="ih-modal__header">
				<div class="ih-modal__header-inner">
					<div class="ih-modal__provider">
						<div class="ih-modal__provider-icon">
							{#if selectedProvider.icon_url}
								<img src={selectedProvider.icon_url} alt={selectedProvider.name} class="ih-modal__provider-img" />
							{:else}
								<span class="ih-modal__provider-letter">{selectedProvider.name.charAt(0)}</span>
							{/if}
						</div>
						<div>
							<h2 class="ih-modal__title">{selectedProvider.name}</h2>
							<span class="ih-modal__category-badge">
								{selectedProvider.category}
							</span>
						</div>
					</div>
					<button
						onclick={closeDetailModal}
						class="btn-pill btn-pill-ghost btn-pill-icon"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Body -->
			<div class="ih-modal__body">
				<!-- Description -->
				<p class="ih-modal__desc">
					{selectedProvider.description || `Connect your ${selectedProvider.name} account to sync data and enable powerful automations.`}
				</p>

				<!-- Category Features -->
				{#if categoryInfo[selectedProvider.category]}
					<div class="ih-modal__section">
						<h3 class="ih-modal__section-title">What you can do</h3>
						<ul class="ih-feature-list">
							{#each categoryInfo[selectedProvider.category].features as feature}
								<li class="ih-feature-item">
									<svg class="w-4 h-4 ih-feature-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									{feature}
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				<!-- Sync Info -->
				<div class="ih-sync-panel">
					<h3 class="ih-modal__section-title">Sync details</h3>
					<div class="ih-sync-grid">
						{#if selectedProvider.auto_live_sync}
							<div class="ih-sync-item">
								<div class="ih-sync-icon ih-sync-icon--green">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
									</svg>
								</div>
								<div>
									<div class="ih-sync-label">Sync type</div>
									<div class="ih-sync-value">Live sync</div>
								</div>
							</div>
						{:else}
							<div class="ih-sync-item">
								<div class="ih-sync-icon ih-sync-icon--blue">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
									</svg>
								</div>
								<div>
									<div class="ih-sync-label">Sync type</div>
									<div class="ih-sync-value">Manual/Scheduled</div>
								</div>
							</div>
						{/if}

						{#if selectedProvider.est_nodes}
							<div class="ih-sync-item">
								<div class="ih-sync-icon ih-sync-icon--purple">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
									</svg>
								</div>
								<div>
									<div class="ih-sync-label">Est. nodes</div>
									<div class="ih-sync-value">{selectedProvider.est_nodes}</div>
								</div>
							</div>
						{/if}

						{#if selectedProvider.initial_sync}
							<div class="ih-sync-item">
								<div class="ih-sync-icon ih-sync-icon--amber">
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								</div>
								<div>
									<div class="ih-sync-label">Initial sync</div>
									<div class="ih-sync-value">{selectedProvider.initial_sync}</div>
								</div>
							</div>
						{/if}
					</div>
				</div>

				<!-- Skills -->
				{#if selectedProvider.skills && selectedProvider.skills.length > 0}
					<div class="ih-modal__section">
						<h3 class="ih-modal__section-title">Available skills</h3>
						<div class="ih-tag-list">
							{#each selectedProvider.skills as skill}
								<span class="ih-skill-tag">{skill}</span>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Modules -->
				{#if selectedProvider.modules && selectedProvider.modules.length > 0}
					<div class="ih-modal__section">
						<h3 class="ih-modal__section-title">Works with</h3>
						<div class="ih-tag-list">
							{#each selectedProvider.modules as module}
								<span class="ih-module-tag">
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
			<div class="ih-modal__footer">
				{#if isProviderConnected(selectedProvider.id)}
					<div class="ih-modal__connected-row">
						<div class="ih-modal__connected-status">
							<span class="ih-status-dot--green"></span>
							<span class="ih-modal__connected-label">Connected</span>
							{#if getConnectedIntegration(selectedProvider.id)?.external_account_name}
								<span class="ih-modal__connected-account">as {getConnectedIntegration(selectedProvider.id)?.external_account_name}</span>
							{/if}
						</div>
						<div class="ih-modal__connected-actions">
							{#if selectedProvider.auto_live_sync}
								<label class="ih-toggle-label">
									<span class="ih-toggle-text">Auto-sync</span>
									<div class="ih-toggle">
										<input type="checkbox" class="sr-only peer" checked />
										<div class="ih-toggle__track"></div>
										<div class="ih-toggle__thumb"></div>
									</div>
								</label>
							{/if}
							<a
								href="/integrations/{getConnectedIntegration(selectedProvider.id)?.id}"
								class="btn-pill btn-pill-ghost btn-pill-sm"
							>
								Settings
							</a>
						</div>
					</div>
				{:else if selectedProvider.status === 'coming_soon'}
					<button
						disabled
						class="btn-pill btn-pill-soft btn-pill-sm ih-modal__full-btn"
					>
						Coming Soon
					</button>
				{:else}
					<button
						onclick={() => { if (selectedProvider) { closeDetailModal(); handleConnect(selectedProvider); } }}
						class="btn-pill btn-pill-primary ih-modal__full-btn"
					>
						{selectedProvider && fileImportProviders.includes(selectedProvider.id) ? 'Import Data' : 'Connect'}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<!-- File Import Modal -->
{#if showFileImportModal && fileImportProvider}
	<div class="ih-modal-backdrop" transition:fade={{ duration: 150 }}>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="ih-modal-backdrop__overlay"
			onclick={closeFileImportModal}
			onkeydown={(e) => { if (e.key === 'Escape') closeFileImportModal(); }}
		></div>
		<div
			class="ih-modal ih-modal--sm"
			transition:fade={{ duration: 150 }}
			role="dialog"
			aria-label="Import data from {fileImportProvider.name}"
		>
			<!-- Header -->
			<div class="ih-modal__header">
				<div class="ih-modal__header-inner">
					<div class="ih-modal__provider">
						{#if fileImportProvider.logo_url}
							<img src={fileImportProvider.logo_url} alt="" class="ih-import-icon" />
						{/if}
						<div>
							<h3 class="ih-modal__title ih-modal__title--sm">Import from {fileImportProvider.name}</h3>
							<p class="ih-modal__subtitle">Upload your exported data file</p>
						</div>
					</div>
					<button
						onclick={closeFileImportModal}
						class="btn-pill btn-pill-ghost btn-pill-icon"
						aria-label="Close import dialog"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<!-- Body -->
			<div class="ih-modal__body">
				<p class="ih-modal__help-text">
					Export your data from {fileImportProvider.name} and upload the file here. Supported formats: JSON, ZIP.
				</p>

				<!-- File Drop Zone -->
				<label class="ih-dropzone {fileImportFile ? 'ih-dropzone--active' : ''}">
					<input
						bind:this={fileInputRef}
						type="file"
						accept=".json,.zip,.csv,.txt"
						class="hidden"
						onchange={handleFileSelect}
					/>
					{#if fileImportFile}
						<svg class="w-8 h-8 ih-dropzone__icon--success" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						<span class="ih-dropzone__filename">{fileImportFile.name}</span>
						<span class="ih-dropzone__filesize">
							{(fileImportFile.size / 1024).toFixed(1)} KB
						</span>
					{:else}
						<svg class="w-8 h-8 ih-dropzone__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
						</svg>
						<span class="ih-dropzone__label">Click to select a file</span>
						<span class="ih-dropzone__formats">JSON, ZIP, CSV, or TXT</span>
					{/if}
				</label>

				<!-- Error / Success Messages -->
				{#if fileImportError}
					<div class="ih-alert ih-alert--error">
						<p>{fileImportError}</p>
					</div>
				{/if}
				{#if fileImportSuccess}
					<div class="ih-alert ih-alert--success">
						<p>{fileImportSuccess}</p>
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="ih-modal__footer">
				<button
					onclick={closeFileImportModal}
					class="btn-pill btn-pill-ghost btn-pill-sm"
				>
					Cancel
				</button>
				<button
					onclick={handleFileImport}
					disabled={!fileImportFile || fileImporting}
					class="btn-pill btn-pill-primary btn-pill-sm"
				>
					{#if fileImporting}
						<span class="ih-import-loading">
							<svg class="w-4 h-4 ih-spinner--inline" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
							Importing...
						</span>
					{:else}
						Import Data
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	/* ═══════════════════════════════════════════════════════════
	   INTEGRATIONS HUB — Foundation ih- Prefix System
	   ═══════════════════════════════════════════════════════════ */

	/* Page Layout */
	.ih-page {
		min-height: 100vh;
		overflow-y: auto;
		background: var(--dbg);
	}
	.ih-header {
		background: var(--dbg2);
		border-bottom: 1px solid var(--dbd);
		position: sticky;
		top: 0;
		z-index: 10;
	}
	.ih-header__inner {
		max-width: 80rem;
		margin: 0 auto;
		padding: 1.5rem 2rem 0;
	}
	.ih-header__top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 1rem;
	}
	.ih-header__title {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--dt);
	}
	.ih-header__subtitle {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-top: 0.25rem;
	}
	.ih-decisions-alert {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.375rem 0.75rem;
		border-radius: 9999px;
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
		border: none;
		transition: background 0.15s;
	}
	.ih-decisions-alert:hover {
		background: rgba(245, 158, 11, 0.2);
	}

	/* Tabs */
	.ih-tabs {
		display: flex;
		gap: 0;
		border-bottom: 1px solid var(--dbd);
		margin: 0 -2rem;
		padding: 0 2rem;
	}
	.ih-tab {
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt3);
		border-bottom: 2px solid transparent;
		cursor: pointer;
		background: none;
		border-top: none;
		border-left: none;
		border-right: none;
		transition: color 0.15s, border-color 0.15s;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.ih-tab:hover {
		color: var(--dt2);
	}
	.ih-tab--active {
		color: #3b82f6;
		border-bottom-color: #3b82f6;
	}
	.ih-tab__count {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 1.25rem;
		height: 1.25rem;
		padding: 0 0.375rem;
		border-radius: 9999px;
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
		font-size: 0.75rem;
		font-weight: 600;
	}

	/* Content Area */
	.ih-content {
		max-width: 80rem;
		margin: 0 auto;
		padding: 1.5rem 2rem;
	}
	.ih-spinner-wrap {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 3rem 0;
	}
	.ih-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--dbd);
		border-top-color: #3b82f6;
		border-radius: 50%;
		animation: ih-spin 0.7s linear infinite;
	}
	.ih-spinner--sm {
		width: 1rem;
		height: 1rem;
		border-width: 2px;
		border-color: var(--dbd);
		border-top-color: #3b82f6;
		border-radius: 50%;
		animation: ih-spin 0.7s linear infinite;
	}
	.ih-spinner--inline {
		animation: ih-spin 0.7s linear infinite;
	}
	@keyframes ih-spin {
		to { transform: rotate(360deg); }
	}

	/* Empty States */
	.ih-empty {
		text-align: center;
		padding: 3rem 0;
	}
	.ih-empty__icon {
		width: 3rem;
		height: 3rem;
		margin: 0 auto;
		color: var(--dt4);
	}
	.ih-empty__title {
		margin-top: 1rem;
		font-size: 1.125rem;
		font-weight: 500;
		color: var(--dt);
	}
	.ih-empty__text {
		margin-top: 0.5rem;
		color: var(--dt3);
	}

	/* Card Grid */
	.ih-grid {
		display: grid;
		grid-template-columns: repeat(1, 1fr);
		gap: 1rem;
	}
	@media (min-width: 768px) {
		.ih-grid { grid-template-columns: repeat(2, 1fr); }
	}
	@media (min-width: 1024px) {
		.ih-grid { grid-template-columns: repeat(3, 1fr); }
	}
	.ih-grid--pb {
		display: grid;
		grid-template-columns: repeat(1, 1fr);
		gap: 1rem;
		padding-bottom: 2rem;
	}
	@media (min-width: 768px) {
		.ih-grid--pb { grid-template-columns: repeat(2, 1fr); }
	}
	@media (min-width: 1024px) {
		.ih-grid--pb { grid-template-columns: repeat(3, 1fr); }
	}

	/* Connected Cards */
	.ih-card {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 0.75rem;
		padding: 1rem;
		transition: box-shadow 0.15s;
	}
	.ih-card:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
	}
	.ih-card__header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
	}
	.ih-card__icon-wrap {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 0.5rem;
		background: var(--dbg3);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}
	.ih-card__icon-wrap img {
		width: 1.5rem;
		height: 1.5rem;
		object-fit: contain;
	}
	.ih-card__icon-letter {
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--dt3);
	}
	.ih-card__icon-letter--sm {
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--dt3);
	}
	.ih-card__info {
		margin-left: 0.75rem;
		flex: 1;
		min-width: 0;
	}
	.ih-card__name-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.ih-card__name {
		font-weight: 500;
		color: var(--dt);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.ih-card__meta {
		font-size: 0.75rem;
		color: var(--dt4);
		margin-top: 0.25rem;
	}
	.ih-card__actions {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		margin-left: 0.5rem;
	}
	.ih-card__actions-btn {
		padding: 0.375rem;
		border-radius: 0.375rem;
		color: var(--dt4);
		background: none;
		border: none;
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
	}
	.ih-card__actions-btn:hover {
		color: var(--dt2);
		background: var(--dbg3);
	}
	.ih-card__settings-link {
		font-size: 0.75rem;
		color: #3b82f6;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		margin-top: 0.5rem;
	}
	.ih-card__settings-link:hover {
		text-decoration: underline;
	}

	/* Badges */
	.ih-badge {
		display: inline-flex;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
		white-space: nowrap;
	}
	.ih-badge--connected {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}
	.ih-badge--available {
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}
	.ih-badge--coming-soon {
		background: rgba(156, 163, 175, 0.1);
		color: var(--dt4);
	}
	.ih-badge--error {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}
	.ih-badge--default {
		background: rgba(156, 163, 175, 0.1);
		color: var(--dt3);
	}

	/* Priority Badges */
	.ih-priority {
		display: inline-flex;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
	}
	.ih-priority--urgent {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
	}
	.ih-priority--high {
		background: rgba(249, 115, 22, 0.1);
		color: #f97316;
	}
	.ih-priority--medium {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
	}
	.ih-priority--default {
		background: rgba(156, 163, 175, 0.1);
		color: var(--dt3);
	}

	/* Section Intro (Available tab) */
	.ih-section-intro {
		margin-bottom: 1.5rem;
	}
	.ih-section-intro__title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
	}
	.ih-section-intro__text {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-top: 0.25rem;
	}

	/* Category Filter */
	.ih-category-filter {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
	}
	.ih-category-btn {
		padding: 0.375rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.875rem;
		font-weight: 500;
		border: 1px solid var(--dbd);
		background: var(--dbg2);
		color: var(--dt3);
		cursor: pointer;
		transition: all 0.15s;
		text-transform: capitalize;
	}
	.ih-category-btn:hover {
		border-color: var(--dbd2);
		color: var(--dt2);
	}
	.ih-category-btn--active {
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
		border-color: rgba(59, 130, 246, 0.3);
	}

	/* Provider Cards */
	.ih-provider-card {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 0.75rem;
		padding: 1rem;
		cursor: pointer;
		transition: all 0.15s;
	}
	.ih-provider-card:hover {
		border-color: var(--dbd2);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
	}
	.ih-provider-card--soon {
		opacity: 0.6;
	}
	.ih-provider-card--connecting {
		border-color: rgba(59, 130, 246, 0.3);
	}
	.ih-provider-card__header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 0.5rem;
	}
	.ih-provider-card__left {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.ih-provider-card__icon {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 0.5rem;
		background: var(--dbg3);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}
	.ih-provider-card__icon img {
		width: 1.5rem;
		height: 1.5rem;
		object-fit: contain;
	}
	.ih-provider-card__name {
		font-weight: 500;
		color: var(--dt);
	}
	.ih-provider-card__desc {
		font-size: 0.75rem;
		color: var(--dt3);
		margin-bottom: 0.75rem;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.ih-provider-card__stats {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	/* Status Pills */
	.ih-status-pill {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
		font-weight: 500;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
	}
	.ih-status-pill--connected {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}
	.ih-status-pill--soon {
		background: rgba(156, 163, 175, 0.1);
		color: var(--dt4);
	}
	.ih-status-pill--connecting {
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}
	.ih-status-dot--green {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		background: #22c55e;
		display: inline-block;
	}

	/* Auto-sync Badge */
	.ih-autosync-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.625rem;
		font-weight: 500;
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
		background: rgba(168, 85, 247, 0.1);
		color: #a855f7;
	}

	/* Stat Row */
	.ih-stat-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-size: 0.75rem;
	}
	.ih-stat-row__label {
		color: var(--dt4);
	}
	.ih-stat-row__value {
		color: var(--dt2);
		font-weight: 500;
	}

	/* Learn More Link */
	.ih-learn-more {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		font-size: 0.75rem;
		color: #3b82f6;
		text-decoration: none;
		margin-top: 0.5rem;
	}
	.ih-learn-more:hover {
		text-decoration: underline;
	}

	/* Tooltip */
	.ih-tooltip {
		position: absolute;
		bottom: 100%;
		left: 50%;
		transform: translateX(-50%);
		margin-bottom: 0.5rem;
		padding: 0.375rem 0.75rem;
		background: var(--dbg3);
		color: var(--dt2);
		font-size: 0.75rem;
		border-radius: 0.375rem;
		white-space: nowrap;
		pointer-events: none;
		z-index: 10;
	}
	.ih-tooltip__arrow {
		position: absolute;
		top: 100%;
		left: 50%;
		transform: translateX(-50%);
		width: 0;
		height: 0;
		border-left: 4px solid transparent;
		border-right: 4px solid transparent;
		border-top: 4px solid var(--dbg3);
	}

	/* AI Models Section */
	.ih-section {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 0.75rem;
		padding: 1.5rem;
		margin-bottom: 1.5rem;
	}
	.ih-section__title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
		margin-bottom: 0.25rem;
	}
	.ih-section__desc {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-bottom: 1rem;
	}
	.ih-tier-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.ih-tier {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 0.5rem;
		padding: 1rem;
	}
	.ih-tier__name {
		font-weight: 600;
		color: var(--dt);
		text-transform: capitalize;
		margin-bottom: 0.25rem;
	}
	.ih-tier__desc {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-bottom: 0.5rem;
	}
	.ih-tier__model {
		font-size: 0.75rem;
		font-family: monospace;
		color: var(--dt4);
	}
	.ih-ai-settings {
		margin-top: 1rem;
	}
	.ih-ai-settings__title {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--dt);
		margin-bottom: 0.75rem;
	}
	.ih-ai-settings__list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.ih-checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: var(--dt2);
		cursor: pointer;
	}
	.ih-checkbox {
		width: 1rem;
		height: 1rem;
		border-radius: 0.25rem;
		accent-color: #3b82f6;
	}
	.ih-latency-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1rem;
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 0.5rem;
	}
	.ih-latency-row__value {
		font-family: monospace;
		font-size: 0.875rem;
		color: #22c55e;
	}

	/* Decisions Tab */
	.ih-decision-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.ih-decision__top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
	}
	.ih-decision__header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.ih-decision__desc {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-top: 0.25rem;
	}
	.ih-decision__meta {
		font-size: 0.75rem;
		color: var(--dt4);
		margin-top: 0.5rem;
	}
	.ih-decision__options {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-top: 1rem;
	}

	/* ═══════════════════════════════════════════
	   Detail Modal
	   ═══════════════════════════════════════════ */
	.ih-modal-backdrop {
		position: fixed;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		background: rgba(0, 0, 0, 0.5);
	}
	.ih-modal-backdrop__overlay {
		position: fixed;
		inset: 0;
	}
	.ih-modal {
		position: relative;
		z-index: 10;
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 1rem;
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.25);
		max-width: 32rem;
		width: 100%;
		max-height: 85vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}
	.ih-modal--sm {
		max-width: 28rem;
	}
	.ih-modal__header {
		padding: 1.5rem;
		border-bottom: 1px solid var(--dbd);
	}
	.ih-modal__header-inner {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
	}
	.ih-modal__provider {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.ih-modal__provider-icon {
		width: 3.5rem;
		height: 3.5rem;
		border-radius: 0.75rem;
		background: var(--dbg3);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		flex-shrink: 0;
	}
	.ih-modal__provider-img {
		width: 2.25rem;
		height: 2.25rem;
		object-fit: contain;
	}
	.ih-modal__provider-letter {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--dt3);
	}
	.ih-modal__title {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--dt);
	}
	.ih-modal__title--sm {
		font-size: 1rem;
	}
	.ih-modal__subtitle {
		font-size: 0.75rem;
		color: var(--dt4);
	}
	.ih-modal__category-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.125rem 0.5rem;
		margin-top: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
		border-radius: 9999px;
		background: var(--dbg3);
		color: var(--dt3);
		text-transform: capitalize;
	}
	.ih-modal__body {
		padding: 1.5rem;
		overflow-y: auto;
		flex: 1;
	}
	.ih-modal__desc {
		color: var(--dt3);
		margin-bottom: 1.5rem;
	}
	.ih-modal__help-text {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-bottom: 1rem;
	}
	.ih-modal__section {
		margin-bottom: 1.5rem;
	}
	.ih-modal__section-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--dt);
		margin-bottom: 0.75rem;
	}
	.ih-modal__footer {
		padding: 1.5rem;
		border-top: 1px solid var(--dbd);
		background: var(--dbg);
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 0.75rem;
	}
	.ih-modal__full-btn {
		width: 100%;
	}
	.ih-modal__connected-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
	}
	.ih-modal__connected-status {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: #22c55e;
	}
	.ih-modal__connected-label {
		font-size: 0.875rem;
		font-weight: 500;
	}
	.ih-modal__connected-account {
		font-size: 0.875rem;
		color: var(--dt3);
	}
	.ih-modal__connected-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	/* Feature List */
	.ih-feature-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.ih-feature-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: var(--dt3);
	}
	.ih-feature-icon {
		color: #22c55e;
		flex-shrink: 0;
	}

	/* Sync Panel */
	.ih-sync-panel {
		background: var(--dbg);
		border-radius: 0.75rem;
		padding: 1rem;
		margin-bottom: 1.5rem;
	}
	.ih-sync-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
	}
	.ih-sync-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.ih-sync-icon {
		width: 2rem;
		height: 2rem;
		border-radius: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}
	.ih-sync-icon--green {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
	}
	.ih-sync-icon--blue {
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
	}
	.ih-sync-icon--purple {
		background: rgba(168, 85, 247, 0.1);
		color: #a855f7;
	}
	.ih-sync-icon--amber {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
	}
	.ih-sync-label {
		font-size: 0.75rem;
		color: var(--dt4);
	}
	.ih-sync-value {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt);
	}

	/* Tags */
	.ih-tag-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	.ih-skill-tag {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
		font-family: monospace;
		background: var(--dbg3);
		color: var(--dt3);
		border-radius: 0.25rem;
	}
	.ih-module-tag {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
		background: rgba(59, 130, 246, 0.08);
		color: #3b82f6;
		border-radius: 0.25rem;
		text-transform: capitalize;
	}

	/* Toggle */
	.ih-toggle-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}
	.ih-toggle-text {
		font-size: 0.875rem;
		color: var(--dt3);
	}
	.ih-toggle {
		position: relative;
		width: 2.5rem;
		height: 1.25rem;
	}
	.ih-toggle__track {
		width: 100%;
		height: 100%;
		border-radius: 9999px;
		background: var(--dbg3);
		transition: background 0.2s;
	}
	.ih-toggle :checked ~ .ih-toggle__track {
		background: #22c55e;
	}
	.ih-toggle__thumb {
		position: absolute;
		left: 0.125rem;
		top: 0.125rem;
		width: 1rem;
		height: 1rem;
		border-radius: 50%;
		background: white;
		transition: transform 0.2s;
	}
	.ih-toggle :checked ~ .ih-toggle__thumb {
		transform: translateX(1.25rem);
	}

	/* ═══════════════════════════════════════════
	   File Import
	   ═══════════════════════════════════════════ */
	.ih-import-icon {
		width: 2rem;
		height: 2rem;
		border-radius: 0.5rem;
	}
	.ih-dropzone {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 8rem;
		border: 2px dashed var(--dbd);
		border-radius: 0.75rem;
		cursor: pointer;
		transition: border-color 0.15s, background 0.15s;
		background: var(--dbg);
	}
	.ih-dropzone:hover {
		border-color: var(--dbd2);
	}
	.ih-dropzone--active {
		border-color: #22c55e;
		background: rgba(34, 197, 94, 0.05);
	}
	.ih-dropzone__icon {
		color: var(--dt4);
		margin-bottom: 0.5rem;
	}
	.ih-dropzone__icon--success {
		color: #22c55e;
		margin-bottom: 0.5rem;
	}
	.ih-dropzone__filename {
		font-size: 0.875rem;
		font-weight: 500;
		color: #22c55e;
	}
	.ih-dropzone__filesize {
		font-size: 0.75rem;
		color: var(--dt4);
		margin-top: 0.25rem;
	}
	.ih-dropzone__label {
		font-size: 0.875rem;
		color: var(--dt3);
	}
	.ih-dropzone__formats {
		font-size: 0.75rem;
		color: var(--dt4);
		margin-top: 0.25rem;
	}

	/* Alerts */
	.ih-alert {
		margin-top: 0.75rem;
		padding: 0.75rem;
		border-radius: 0.5rem;
		font-size: 0.875rem;
	}
	.ih-alert--error {
		background: rgba(239, 68, 68, 0.08);
		border: 1px solid rgba(239, 68, 68, 0.2);
		color: #ef4444;
	}
	.ih-alert--success {
		background: rgba(34, 197, 94, 0.08);
		border: 1px solid rgba(34, 197, 94, 0.2);
		color: #22c55e;
	}
	.ih-import-loading {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	/* Search Input */
	.ih-search-wrap {
		position: relative;
		flex-shrink: 0;
	}
	.ih-search-icon {
		position: absolute;
		left: 0.625rem;
		top: 50%;
		transform: translateY(-50%);
		color: var(--dt4);
		pointer-events: none;
	}
	.ih-search-input {
		padding: 0.375rem 0.75rem 0.375rem 2rem;
		font-size: 0.8125rem;
		border-radius: 0.5rem;
		border: 1px solid var(--dbd);
		background: var(--dbg);
		color: var(--dt);
		outline: none;
		width: 14rem;
		transition: border-color 0.15s;
	}
	.ih-search-input:focus {
		border-color: #3b82f6;
	}
	.ih-search-input::placeholder {
		color: var(--dt4);
	}

	/* Card Sync Button */
	.ih-card__sync-btn {
		padding: 0.375rem;
		border-radius: 0.375rem;
		color: var(--dt4);
		background: none;
		border: 1px solid var(--dbd);
		cursor: pointer;
		transition: color 0.15s, background 0.15s;
		display: inline-flex;
		align-items: center;
	}
	.ih-card__sync-btn:hover:not(:disabled) {
		color: #3b82f6;
		background: rgba(59, 130, 246, 0.08);
		border-color: rgba(59, 130, 246, 0.3);
	}
	.ih-card__sync-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Card sub-meta (last used) */
	.ih-card__sub-meta {
		font-size: 0.6875rem;
		color: var(--dt4);
		margin-top: 0.125rem;
	}

	/* Latency Input */
	.ih-latency-input {
		width: 5rem;
		padding: 0.25rem 0.5rem;
		font-family: monospace;
		font-size: 0.875rem;
		color: #22c55e;
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 0.375rem;
		outline: none;
		text-align: right;
	}
	.ih-latency-input:focus {
		border-color: #3b82f6;
	}
	.ih-latency-unit {
		font-size: 0.75rem;
		color: var(--dt4);
	}

	/* Alert small variant */
	.ih-alert--sm {
		margin-top: 0.5rem;
		margin-bottom: 0.5rem;
		padding: 0.5rem 0.75rem;
		font-size: 0.8125rem;
	}
</style>
