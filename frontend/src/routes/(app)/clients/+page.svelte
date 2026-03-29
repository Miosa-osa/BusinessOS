<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { clients, type ViewMode } from '$lib/stores/clients';
	import { moduleEvents } from '$lib/stores/events';
	import type { ClientListResponse, ClientStatus, ClientType, CreateClientData } from '$lib/api';
	import {
		ClientViewSwitcher,
		ClientTableView,
		ClientCardView,
		ClientKanbanView,
		AddClientModal
	} from '$lib/components/clients';

	// Check if we're in embed mode to propagate to links
	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	// State
	let showAddModal = $state(false);

	// Reactive store access via auto-subscription (avoids $effect + subscribe infinite loop)
	let clientsList = $derived($clients.clients);
	let loading = $derived($clients.loading);
	let error = $derived($clients.error);
	let viewMode = $derived($clients.viewMode);
	let statusFilter = $derived($clients.filters.status);
	let typeFilter = $derived($clients.filters.type);
	let searchQuery = $derived($clients.filters.search);

	// Load clients on mount
	onMount(async () => {
		try {
			await clients.loadClients();
		} catch {
			// Backend unavailable — empty state will show
		}
	});

	// Reload when a client or deal is created from the chat slash command
	$effect(() => {
		const event = $moduleEvents;
		if (event?.type === 'client:created' || event?.type === 'deal:created') {
			clients.loadClients();
		}
	});

	// Filtered clients based on local search (for immediate feedback)
	const filteredClients = $derived.by(() => {
		if (!searchQuery) return clientsList;
		const query = searchQuery.toLowerCase();
		return clientsList.filter(
			(c) =>
				c.name.toLowerCase().includes(query) ||
				(c.email && c.email.toLowerCase().includes(query)) ||
				(c.phone && c.phone.includes(query))
		);
	});

	function handleClientClick(id: string) {
		goto(`/clients/${id}${embedSuffix}`);
	}

	async function handleStatusChange(id: string, status: ClientStatus) {
		try {
			await clients.updateClientStatus(id, status);
		} catch (err) {
			console.error('Failed to update status:', err);
		}
	}

	function handleViewChange(mode: ViewMode) {
		clients.setViewMode(mode);
	}

	function handleSearchChange(query: string) {
		clients.setFilters({ search: query });
	}

	function handleStatusFilterChange(status: ClientStatus | null) {
		clients.setFilters({ status });
		clients.loadClients();
	}

	function handleTypeFilterChange(type: ClientType | null) {
		clients.setFilters({ type });
		clients.loadClients();
	}

	async function handleCreateClient(data: CreateClientData) {
		try {
			const client = await clients.createClient(data);
			showAddModal = false;
			// Navigate to the new client
			goto(`/clients/${client.id}`);
		} catch (err) {
			console.error('Failed to create client:', err);
		}
	}
</script>

<div class="cr-page">
	<!-- Header -->
	<div class="cr-page__header">
		<div>
			<h1 class="cr-page__title">Clients</h1>
			<p class="cr-page__subtitle">Manage your clients and track relationships</p>
		</div>
		<button
			onclick={() => (showAddModal = true)}
			class="btn-cta"
			aria-label="Add new client"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			Add Client
		</button>
	</div>

	<!-- View Switcher & Filters -->
	<ClientViewSwitcher
		view={viewMode}
		{searchQuery}
		{statusFilter}
		{typeFilter}
		onViewChange={handleViewChange}
		onSearchChange={handleSearchChange}
		onStatusChange={handleStatusFilterChange}
		onTypeChange={handleTypeFilterChange}
	/>

	<!-- Error State -->
	{#if error}
		<div class="cr-page__error">
			<p class="cr-page__error-text">{error}</p>
			<button
				onclick={() => clients.loadClients()}
				class="cr-page__error-retry"
				aria-label="Retry loading clients"
			>
				Try again
			</button>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading && clientsList.length === 0}
		<div class="cr-page__center">
			<div class="cr-page__center-content">
				<svg
					class="w-8 h-8 animate-spin"
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
				<p class="cr-page__center-text">Loading clients...</p>
			</div>
		</div>
	{:else if clientsList.length === 0 && !loading}
		<!-- Empty State -->
		<div class="cr-page__center">
			<div class="cr-page__center-content">
				<svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
					/>
				</svg>
				<p class="cr-page__empty-title">No clients yet</p>
				<p class="cr-page__center-text">Add your first client to get started</p>
				<button
					onclick={() => (showAddModal = true)}
					class="btn-cta"
					aria-label="Add first client"
				>
					Add Client
				</button>
			</div>
		</div>
	{:else}
		<!-- Content -->
		{#if viewMode === 'table'}
			<ClientTableView
				clients={filteredClients}
				onClientClick={handleClientClick}
				onStatusChange={handleStatusChange}
			/>
		{:else if viewMode === 'cards'}
			<ClientCardView clients={filteredClients} onClientClick={handleClientClick} />
		{:else if viewMode === 'kanban'}
			<ClientKanbanView
				clients={filteredClients}
				onClientClick={handleClientClick}
				onStatusChange={handleStatusChange}
			/>
		{/if}
	{/if}
</div>

<!-- Add Client Modal -->
<AddClientModal bind:open={showAddModal} onCreate={handleCreateClient} />

<style>
	/* ─── Clients Page Layout (cr- prefix from Foundation CRM) ─── */
	.cr-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--dbg, #fff);
	}
	.cr-page__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 24px;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}
	.cr-page__title {
		font-size: 20px;
		font-weight: 700;
		color: var(--dt, #111);
		letter-spacing: -0.01em;
	}
	.cr-page__subtitle {
		font-size: 13px;
		color: var(--dt3, #888);
		margin-top: 2px;
	}
	.cr-page__error {
		margin: 16px 24px 0;
		padding: 12px 16px;
		border-radius: 10px;
		border: 1px solid rgba(239, 68, 68, 0.2);
		background: rgba(239, 68, 68, 0.06);
	}
	:global(.dark) .cr-page__error {
		border-color: rgba(239, 68, 68, 0.25);
		background: rgba(239, 68, 68, 0.1);
	}
	.cr-page__error-text {
		font-size: 13px;
		color: #ef4444;
		font-weight: 500;
	}
	.cr-page__error-retry {
		margin-top: 6px;
		font-size: 12px;
		color: #ef4444;
		text-decoration: underline;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}
	.cr-page__error-retry:hover { opacity: 0.8; }
	.cr-page__center {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.cr-page__center-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		color: var(--dt3, #888);
	}
	.cr-page__center-text {
		font-size: 13px;
		color: var(--dt3, #888);
	}
	.cr-page__empty-title {
		font-size: 16px;
		font-weight: 600;
		color: var(--dt, #111);
	}
</style>
