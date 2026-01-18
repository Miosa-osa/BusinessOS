<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { clients, type ViewMode } from '$lib/stores/clients';
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

	// Subscribe to clients store
	let clientsList = $state<ClientListResponse[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let viewMode = $state<ViewMode>('table');
	let statusFilter = $state<ClientStatus | null>(null);
	let typeFilter = $state<ClientType | null>(null);
	let searchQuery = $state('');

	$effect(() => {
		const unsubscribe = clients.subscribe((state) => {
			clientsList = state.clients;
			loading = state.loading;
			error = state.error;
			viewMode = state.viewMode;
			statusFilter = state.filters.status;
			typeFilter = state.filters.type;
			searchQuery = state.filters.search;
		});
		return unsubscribe;
	});

	// Load clients on mount
	onMount(() => {
		clients.loadClients();
	});

	// Filtered clients based on local search (for immediate feedback)
	const filteredClients = $derived(() => {
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

<div class="flex flex-col h-full bg-white">
	<!-- Header -->
	<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
		<div>
			<h1 class="text-2xl font-semibold text-gray-900">Clients</h1>
			<p class="text-sm text-gray-500 mt-0.5">Manage your clients and track relationships</p>
		</div>
		<button
			onclick={() => (showAddModal = true)}
			class="btn-pill btn-pill-primary btn-pill-sm"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			Add Client
		</button>
	</div>

	<!-- View Switcher & Filters -->
	<ClientViewSwitcher
		bind:view={viewMode}
		bind:searchQuery
		{statusFilter}
		{typeFilter}
		onViewChange={handleViewChange}
		onSearchChange={handleSearchChange}
		onStatusChange={handleStatusFilterChange}
		onTypeChange={handleTypeFilterChange}
	/>

	<!-- Error State -->
	{#if error}
		<div class="mx-6 mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
			<p class="text-sm text-red-700">{error}</p>
			<button
				onclick={() => clients.loadClients()}
				class="btn-pill btn-pill-ghost btn-pill-xs mt-2"
			>
				Try again
			</button>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading && clientsList.length === 0}
		<div class="flex-1 flex items-center justify-center">
			<div class="flex flex-col items-center gap-3 text-gray-500">
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
				<p class="text-sm">Loading clients...</p>
			</div>
		</div>
	{:else if clientsList.length === 0 && !loading}
		<!-- Empty State -->
		<div class="flex-1 flex items-center justify-center">
			<div class="flex flex-col items-center gap-3 text-gray-500">
				<svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
					/>
				</svg>
				<p class="text-lg font-medium text-gray-900">No clients yet</p>
				<p class="text-sm">Add your first client to get started</p>
				<button
					onclick={() => (showAddModal = true)}
					class="btn-pill btn-pill-primary btn-pill-sm mt-2"
				>
					Add Client
				</button>
			</div>
		</div>
	{:else}
		<!-- Content -->
		{#if viewMode === 'table'}
			<ClientTableView
				clients={filteredClients()}
				onClientClick={handleClientClick}
				onStatusChange={handleStatusChange}
			/>
		{:else if viewMode === 'cards'}
			<ClientCardView clients={filteredClients()} onClientClick={handleClientClick} />
		{:else if viewMode === 'kanban'}
			<ClientKanbanView
				clients={filteredClients()}
				onClientClick={handleClientClick}
				onStatusChange={handleStatusChange}
			/>
		{/if}
	{/if}
</div>

<!-- Add Client Modal -->
<AddClientModal bind:open={showAddModal} onCreate={handleCreateClient} />
