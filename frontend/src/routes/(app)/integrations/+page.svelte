<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade } from 'svelte/transition';
	import { useSession, clearSession } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import * as integrationsApi from '$lib/api/integrations';
	import type {
		IntegrationProviderInfo,
		UserIntegration,
		AIModelPreferences,
		PendingDecision,
		IntegrationCategory
	} from '$lib/api/integrations';
	import IntegrationDetailModal from '$lib/components/integrations/IntegrationDetailModal.svelte';
	import AIModelsConfig from '$lib/components/integrations/AIModelsConfig.svelte';
	import ConnectedIntegrations from '$lib/components/integrations/ConnectedIntegrations.svelte';
	import AvailableIntegrations from '$lib/components/integrations/AvailableIntegrations.svelte';

	const session = useSession();

	// State
	let isLoading = $state(true);
	let activeTab = $state<'connected' | 'available' | 'ai' | 'decisions'>('available');
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

	// Reactive auth check - updates when session changes
	$effect(() => {
		const sessionData = $session;
		if (!sessionData?.isPending) {
			isAuthenticated = !!sessionData?.data?.user;
		}
	});

	onMount(async () => {
		if (browser) {
			const urlParams = new URLSearchParams(window.location.search);
			const connectedProvider = urlParams.get('connected');
			if (connectedProvider && window.opener) {
				window.opener.postMessage({ type: 'integration-connected', provider: connectedProvider }, window.location.origin);
				window.close();
				return;
			}
			if (connectedProvider) {
				activeTab = 'connected';
				const url = new URL(window.location.href);
				url.searchParams.delete('connected');
				window.history.replaceState({}, '', url.toString());
			}
		}

		function handleOAuthMessage(event: MessageEvent) {
			if (event.origin !== window.location.origin) return;
			if (event.data?.type === 'integration-connected') {
				loadData();
				activeTab = 'connected';
			}
		}
		window.addEventListener('message', handleOAuthMessage);
		oauthMessageCleanup = () => window.removeEventListener('message', handleOAuthMessage);

		loadProviders();

		let attempts = 0;
		while ($session?.isPending && attempts < 20) {
			await new Promise((r) => setTimeout(r, 100));
			attempts++;
		}

		const sessionData = $session;
		isAuthenticated = !sessionData?.isPending && !!sessionData?.data?.user;

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
		if (authDataLoading || authDataLoaded) {
			return;
		}
		authDataLoading = true;

		let authFailed = false;

		try {
			const connected = await integrationsApi.getConnectedIntegrations();
			connectedIntegrations = connected.integrations || [];
		} catch (e: unknown) {
			connectedIntegrations = [];
			if (e instanceof Error && e.message.includes('401')) {
				authFailed = true;
			}
		}

		if (authFailed) {
			clearSession();
			isAuthenticated = false;
			authDataLoading = false;
			isLoading = false;
			return;
		}

		try {
			const prefs = await integrationsApi.getAIModelPreferences();
			aiPreferences = prefs.preferences;
		} catch {
			aiPreferences = null;
		}

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

		if (fileImportProviders.includes(provider.id)) {
			fileImportProvider = provider;
			fileImportFile = null;
			fileImportError = null;
			fileImportSuccess = null;
			showFileImportModal = true;
			return;
		}

		const oauthProvider = provider.oauth_provider || provider.id;
		if (!oauthProvider) {
			alert(`OAuth not configured for ${provider.name}. Please try again later.`);
			return;
		}

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

	function getPriorityBadgeClass(priority: string) {
		switch (priority) {
			case 'urgent': return 'ih-priority--urgent';
			case 'high': return 'ih-priority--high';
			case 'medium': return 'ih-priority--medium';
			default: return 'ih-priority--default';
		}
	}

	// Derived connected integration lookup for modal
	let selectedProviderConnected = $derived(
		selectedProvider
			? connectedIntegrations.find((i) => i.provider_id === selectedProvider!.id)
			: undefined
	);
	let selectedProviderIsConnected = $derived(
		selectedProvider
			? connectedIntegrations.some((i) => i.provider_id === selectedProvider!.id && i.status === 'connected')
			: false
	);
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
			<ConnectedIntegrations
				integrations={connectedIntegrations}
				{syncingId}
				onsync={handleSyncCard}
				ondisconnect={handleDisconnect}
				onbrowse={() => (activeTab = 'available')}
			/>
		{:else if activeTab === 'available'}
			<AvailableIntegrations
				providers={filteredProviders}
				{connectedIntegrations}
				{categories}
				{selectedCategory}
				{searchQuery}
				{connectingId}
				onselect={(cat) => (selectedCategory = cat)}
				onsearch={(q) => (searchQuery = q)}
				onconnect={handleConnect}
				ondisconnect={handleDisconnect}
				onviewdetail={openProviderDetail}
			/>
		{:else if activeTab === 'ai'}
			<AIModelsConfig
				{aiPreferences}
				{savingAiPrefs}
				{aiPrefsMessage}
				{aiPrefsError}
				onsave={saveAiPreferences}
			/>
		{:else if activeTab === 'decisions'}
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
	<IntegrationDetailModal
		provider={selectedProvider}
		connectedIntegration={selectedProviderConnected}
		isConnected={selectedProviderIsConnected}
		{fileImportProviders}
		{categoryInfo}
		onclose={closeDetailModal}
		onconnect={handleConnect}
	/>
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
						{#if fileImportProvider.icon_url}
							<img src={fileImportProvider.icon_url} alt="" class="ih-import-icon" />
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
	.ih-spinner--inline {
		animation: ih-spin 0.7s linear infinite;
	}
	@keyframes ih-spin {
		to { transform: rotate(360deg); }
	}

	/* Empty States (decisions tab) */
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

	/* Decisions Tab */
	.ih-decision-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.ih-card {
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 0.75rem;
		padding: 0.75rem;
		transition: box-shadow 0.15s;
	}
	.ih-card:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
	}
	.ih-card__name {
		font-weight: 500;
		color: var(--dt);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
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

	/* Priority Badges */
	.ih-priority {
		display: inline-flex;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
		font-size: 0.75rem;
		font-weight: 500;
	}
	.ih-priority--urgent { background: rgba(239, 68, 68, 0.1); color: #ef4444; }
	.ih-priority--high { background: rgba(249, 115, 22, 0.1); color: #f97316; }
	.ih-priority--medium { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
	.ih-priority--default { background: rgba(156, 163, 175, 0.1); color: var(--dt3); }

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
	.ih-alert--banner {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	/* File Import Modal */
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
	.ih-modal__body {
		padding: 1.5rem;
		overflow-y: auto;
		flex: 1;
	}
	.ih-modal__help-text {
		font-size: 0.875rem;
		color: var(--dt3);
		margin-bottom: 1rem;
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
	.ih-import-loading {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
</style>
