<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { clients, statusColors, statusLabels, dealStageColors, dealStageLabels, interactionTypeLabels, interactionTypeIcons } from '$lib/stores/clients';
	import type {
		ClientDetailResponse,
		ClientStatus,
		ContactResponse,
		InteractionResponse,
		DealResponse,
		CreateContactData,
		CreateInteractionData,
		CreateDealData,
		InteractionType,
		DealStage
	} from '$lib/api';

	// Check if we're in embed mode to propagate to links
	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	type Tab = 'overview' | 'contacts' | 'interactions' | 'deals';

	// State
	let activeTab = $state<Tab>('overview');
	let client = $state<ClientDetailResponse | null>(null);
	let loading = $state(false);
	let error = $state<string | null>(null);

	// Modal states
	let showAddContactModal = $state(false);
	let showAddInteractionModal = $state(false);
	let showAddDealModal = $state(false);
	let showEditModal = $state(false);

	// Form states
	let contactForm = $state<CreateContactData>({ name: '' });
	let interactionForm = $state<CreateInteractionData>({ type: 'call', subject: '' });
	let dealForm = $state<CreateDealData>({ name: '' });

	$effect(() => {
		const unsubscribe = clients.subscribe((state) => {
			client = state.currentClient;
			loading = state.loading;
			error = state.error;
		});
		return unsubscribe;
	});

	// Load client on mount or when ID changes
	$effect(() => {
		const id = $page.params.id;
		if (id) {
			clients.loadClient(id);
		}
	});

	onMount(() => {
		return () => {
			clients.clearCurrent();
		};
	});

	function formatCurrency(value: number | null): string {
		if (value === null || value === undefined) return '-';
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 0
		}).format(value);
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function formatDateTime(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	async function handleStatusChange(status: ClientStatus) {
		if (!client) return;
		try {
			await clients.updateClientStatus(client.id, status);
		} catch (err) {
			console.error('Failed to update status:', err);
		}
	}

	async function handleAddContact() {
		if (!client || !contactForm.name.trim()) return;
		try {
			await clients.createContact(client.id, contactForm);
			showAddContactModal = false;
			contactForm = { name: '' };
		} catch (err) {
			console.error('Failed to add contact:', err);
		}
	}

	async function handleDeleteContact(contactId: string) {
		if (!client) return;
		if (!confirm('Are you sure you want to delete this contact?')) return;
		try {
			await clients.deleteContact(client.id, contactId);
		} catch (err) {
			console.error('Failed to delete contact:', err);
		}
	}

	async function handleAddInteraction() {
		if (!client || !interactionForm.subject.trim()) return;
		try {
			await clients.createInteraction(client.id, interactionForm);
			showAddInteractionModal = false;
			interactionForm = { type: 'call', subject: '' };
		} catch (err) {
			console.error('Failed to add interaction:', err);
		}
	}

	async function handleAddDeal() {
		if (!client || !dealForm.name.trim()) return;
		try {
			await clients.createDeal(client.id, dealForm);
			showAddDealModal = false;
			dealForm = { name: '' };
		} catch (err) {
			console.error('Failed to add deal:', err);
		}
	}

	async function handleDeleteClient() {
		if (!client) return;
		if (!confirm('Are you sure you want to delete this client? This will also delete all contacts, interactions, and deals.')) return;
		try {
			await clients.deleteClient(client.id);
			goto('/clients' + embedSuffix);
		} catch (err) {
			console.error('Failed to delete client:', err);
		}
	}

	function getTotalDealsValue(): number {
		if (!client) return 0;
		return client.deals.reduce((sum, d) => sum + d.value, 0);
	}

	function getActiveDealsValue(): number {
		if (!client) return 0;
		return client.deals
			.filter((d) => d.stage !== 'closed_won' && d.stage !== 'closed_lost')
			.reduce((sum, d) => sum + d.value, 0);
	}
</script>

{#if loading && !client}
	<div class="flex-1 flex items-center justify-center bg-white">
		<div class="flex flex-col items-center gap-3 text-gray-500">
			<svg class="w-8 h-8 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
				/>
			</svg>
			<p class="text-sm">Loading client...</p>
		</div>
	</div>
{:else if error}
	<div class="flex-1 flex items-center justify-center bg-white">
		<div class="text-center">
			<p class="text-red-600">{error}</p>
			<button onclick={() => goto('/clients' + embedSuffix)} class="mt-4 text-sm text-gray-600 hover:text-gray-900 underline">
				Back to clients
			</button>
		</div>
	</div>
{:else if client}
	<div class="flex flex-col h-full bg-gray-50">
		<!-- Header -->
		<div class="bg-white border-b border-gray-200">
			<div class="px-6 py-4">
				<div class="flex items-start justify-between">
					<div class="flex items-start gap-4">
						<button onclick={() => goto('/clients' + embedSuffix)} class="mt-1 p-1 hover:bg-gray-100 rounded-lg transition-colors">
							<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
							</svg>
						</button>
						<div class="w-14 h-14 rounded-xl bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-xl font-semibold text-gray-600">
							{getInitials(client.name)}
						</div>
						<div>
							<h1 class="text-2xl font-semibold text-gray-900">{client.name}</h1>
							<div class="flex items-center gap-3 mt-1">
								<span class="inline-flex items-center px-2 py-0.5 text-xs rounded-md {client.type === 'company' ? 'bg-blue-50 text-blue-700' : 'bg-violet-50 text-violet-700'}">
									{client.type === 'company' ? 'Company' : 'Individual'}
								</span>
								<select
									value={client.status}
									onchange={(e) => handleStatusChange((e.target as HTMLSelectElement).value as ClientStatus)}
									class="px-2 py-0.5 text-xs rounded-md border-0 {statusColors[client.status]} cursor-pointer focus:ring-2 focus:ring-gray-900"
								>
									<option value="lead">Lead</option>
									<option value="prospect">Prospect</option>
									<option value="active">Active</option>
									<option value="inactive">Inactive</option>
									<option value="churned">Churned</option>
								</select>
								{#if client.source}
									<span class="text-xs text-gray-500">via {client.source}</span>
								{/if}
							</div>
						</div>
					</div>
					<div class="flex items-center gap-2">
						<button
							onclick={() => (showAddInteractionModal = true)}
							class="px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
						>
							Log Interaction
						</button>
						<button
							onclick={() => (showAddDealModal = true)}
							class="px-3 py-1.5 text-sm font-medium bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors"
						>
							Add Deal
						</button>
						<button
							onclick={handleDeleteClient}
							class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
							title="Delete client"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Tabs -->
			<div class="px-6 flex gap-6 border-t border-gray-100">
				{#each [
					{ id: 'overview', label: 'Overview' },
					{ id: 'contacts', label: `Contacts (${client.contacts.length})` },
					{ id: 'interactions', label: `Interactions (${client.interactions.length})` },
					{ id: 'deals', label: `Deals (${client.deals.length})` }
				] as tab}
					<button
						onclick={() => (activeTab = tab.id as Tab)}
						class="py-3 text-sm font-medium border-b-2 transition-colors {activeTab === tab.id
							? 'border-gray-900 text-gray-900'
							: 'border-transparent text-gray-500 hover:text-gray-700'}"
					>
						{tab.label}
					</button>
				{/each}
			</div>
		</div>

		<!-- Content -->
		<div class="flex-1 overflow-auto">
			{#if activeTab === 'overview'}
				<div class="p-6">
					<div class="grid grid-cols-3 gap-6">
						<!-- Main Info -->
						<div class="col-span-2 space-y-6">
							<!-- Contact Info Card -->
							<div class="bg-white rounded-xl border border-gray-200 p-6">
								<h3 class="text-sm font-medium text-gray-900 mb-4">Contact Information</h3>
								<div class="grid grid-cols-2 gap-4">
									<div>
										<div class="text-xs text-gray-500">Email</div>
										<div class="text-sm text-gray-900">{client.email || '-'}</div>
									</div>
									<div>
										<div class="text-xs text-gray-500">Phone</div>
										<div class="text-sm text-gray-900">{client.phone || '-'}</div>
									</div>
									{#if client.website}
										<div>
											<div class="text-xs text-gray-500">Website</div>
											<a href={client.website} target="_blank" class="text-sm text-blue-600 hover:underline">{client.website}</a>
										</div>
									{/if}
									{#if client.type === 'company'}
										<div>
											<div class="text-xs text-gray-500">Industry</div>
											<div class="text-sm text-gray-900">{client.industry || '-'}</div>
										</div>
										<div>
											<div class="text-xs text-gray-500">Company Size</div>
											<div class="text-sm text-gray-900">{client.company_size || '-'}</div>
										</div>
									{/if}
								</div>

								{#if client.address || client.city}
									<div class="mt-4 pt-4 border-t border-gray-100">
										<div class="text-xs text-gray-500">Address</div>
										<div class="text-sm text-gray-900">
											{client.address || ''}{client.address && client.city ? ', ' : ''}{client.city || ''}{client.city && client.state ? ', ' : ''}{client.state || ''} {client.zip_code || ''}
											{#if client.country}<br />{client.country}{/if}
										</div>
									</div>
								{/if}
							</div>

							<!-- Notes -->
							{#if client.notes}
								<div class="bg-white rounded-xl border border-gray-200 p-6">
									<h3 class="text-sm font-medium text-gray-900 mb-2">Notes</h3>
									<p class="text-sm text-gray-600 whitespace-pre-wrap">{client.notes}</p>
								</div>
							{/if}

							<!-- Recent Interactions -->
							<div class="bg-white rounded-xl border border-gray-200 p-6">
								<div class="flex items-center justify-between mb-4">
									<h3 class="text-sm font-medium text-gray-900">Recent Activity</h3>
									<button onclick={() => (activeTab = 'interactions')} class="text-xs text-gray-500 hover:text-gray-700">
										View all
									</button>
								</div>
								{#if client.interactions.length === 0}
									<p class="text-sm text-gray-500">No interactions yet</p>
								{:else}
									<div class="space-y-3">
										{#each client.interactions.slice(0, 5) as interaction}
											<div class="flex items-start gap-3">
												<span class="text-lg">{interactionTypeIcons[interaction.type]}</span>
												<div class="flex-1 min-w-0">
													<div class="text-sm font-medium text-gray-900">{interaction.subject}</div>
													<div class="text-xs text-gray-500">{formatDateTime(interaction.occurred_at)}</div>
												</div>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						</div>

						<!-- Sidebar -->
						<div class="space-y-6">
							<!-- Stats Card -->
							<div class="bg-white rounded-xl border border-gray-200 p-6">
								<h3 class="text-sm font-medium text-gray-900 mb-4">Summary</h3>
								<div class="space-y-4">
									<div>
										<div class="text-xs text-gray-500">Lifetime Value</div>
										<div class="text-2xl font-semibold text-gray-900">{formatCurrency(client.lifetime_value)}</div>
									</div>
									<div>
										<div class="text-xs text-gray-500">Pipeline Value</div>
										<div class="text-lg font-medium text-emerald-600">{formatCurrency(getActiveDealsValue())}</div>
									</div>
									<div class="pt-3 border-t border-gray-100 grid grid-cols-2 gap-4">
										<div>
											<div class="text-2xl font-semibold text-gray-900">{client.deals.length}</div>
											<div class="text-xs text-gray-500">Total Deals</div>
										</div>
										<div>
											<div class="text-2xl font-semibold text-gray-900">{client.contacts.length}</div>
											<div class="text-xs text-gray-500">Contacts</div>
										</div>
									</div>
								</div>
							</div>

							<!-- Tags -->
							{#if client.tags && client.tags.length > 0}
								<div class="bg-white rounded-xl border border-gray-200 p-6">
									<h3 class="text-sm font-medium text-gray-900 mb-3">Tags</h3>
									<div class="flex flex-wrap gap-2">
										{#each client.tags as tag}
											<span class="px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded-md">{tag}</span>
										{/each}
									</div>
								</div>
							{/if}

							<!-- Dates -->
							<div class="bg-white rounded-xl border border-gray-200 p-6">
								<h3 class="text-sm font-medium text-gray-900 mb-3">Dates</h3>
								<div class="space-y-2 text-sm">
									<div class="flex justify-between">
										<span class="text-gray-500">Created</span>
										<span class="text-gray-900">{formatDate(client.created_at)}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-500">Last Updated</span>
										<span class="text-gray-900">{formatDate(client.updated_at)}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-500">Last Contact</span>
										<span class="text-gray-900">{formatDate(client.last_contacted_at)}</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			{:else if activeTab === 'contacts'}
				<div class="p-6">
					<div class="bg-white rounded-xl border border-gray-200">
						<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
							<h3 class="text-sm font-medium text-gray-900">Contacts</h3>
							<button
								onclick={() => (showAddContactModal = true)}
								class="flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
								Add Contact
							</button>
						</div>
						{#if client.contacts.length === 0}
							<div class="px-6 py-12 text-center text-gray-500">
								<p>No contacts yet</p>
							</div>
						{:else}
							<div class="divide-y divide-gray-100">
								{#each client.contacts as contact}
									<div class="px-6 py-4 flex items-center justify-between hover:bg-gray-50">
										<div class="flex items-center gap-4">
											<div class="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center text-sm font-medium text-gray-600">
												{getInitials(contact.name)}
											</div>
											<div>
												<div class="flex items-center gap-2">
													<span class="text-sm font-medium text-gray-900">{contact.name}</span>
													{#if contact.is_primary}
														<span class="px-1.5 py-0.5 text-xs bg-blue-50 text-blue-700 rounded">Primary</span>
													{/if}
												</div>
												<div class="text-xs text-gray-500">{contact.role || 'No role specified'}</div>
											</div>
										</div>
										<div class="flex items-center gap-4">
											<div class="text-right text-sm">
												<div class="text-gray-900">{contact.email || '-'}</div>
												<div class="text-gray-500">{contact.phone || '-'}</div>
											</div>
											<button
												onclick={() => handleDeleteContact(contact.id)}
												class="p-1 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
											>
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
												</svg>
											</button>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			{:else if activeTab === 'interactions'}
				<div class="p-6">
					<div class="bg-white rounded-xl border border-gray-200">
						<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
							<h3 class="text-sm font-medium text-gray-900">Interaction Timeline</h3>
							<button
								onclick={() => (showAddInteractionModal = true)}
								class="flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
								Log Interaction
							</button>
						</div>
						{#if client.interactions.length === 0}
							<div class="px-6 py-12 text-center text-gray-500">
								<p>No interactions yet</p>
							</div>
						{:else}
							<div class="divide-y divide-gray-100">
								{#each client.interactions as interaction}
									<div class="px-6 py-4">
										<div class="flex items-start gap-4">
											<div class="w-10 h-10 rounded-lg bg-gray-100 flex items-center justify-center text-lg">
												{interactionTypeIcons[interaction.type]}
											</div>
											<div class="flex-1">
												<div class="flex items-center gap-2">
													<span class="text-sm font-medium text-gray-900">{interaction.subject}</span>
													<span class="px-1.5 py-0.5 text-xs bg-gray-100 text-gray-600 rounded">{interactionTypeLabels[interaction.type]}</span>
												</div>
												{#if interaction.description}
													<p class="text-sm text-gray-600 mt-1">{interaction.description}</p>
												{/if}
												{#if interaction.outcome}
													<div class="mt-2 text-xs text-gray-500">
														<span class="font-medium">Outcome:</span> {interaction.outcome}
													</div>
												{/if}
												<div class="text-xs text-gray-400 mt-2">{formatDateTime(interaction.occurred_at)}</div>
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			{:else if activeTab === 'deals'}
				<div class="p-6">
					<div class="bg-white rounded-xl border border-gray-200">
						<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
							<h3 class="text-sm font-medium text-gray-900">Deals</h3>
							<button
								onclick={() => (showAddDealModal = true)}
								class="flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
								</svg>
								Add Deal
							</button>
						</div>
						{#if client.deals.length === 0}
							<div class="px-6 py-12 text-center text-gray-500">
								<p>No deals yet</p>
							</div>
						{:else}
							<div class="divide-y divide-gray-100">
								{#each client.deals as deal}
									<div class="px-6 py-4">
										<div class="flex items-center justify-between">
											<div>
												<div class="flex items-center gap-2">
													<span class="text-sm font-medium text-gray-900">{deal.name}</span>
													<span class="px-1.5 py-0.5 text-xs rounded {dealStageColors[deal.stage]}">{dealStageLabels[deal.stage]}</span>
												</div>
												<div class="text-xs text-gray-500 mt-1">
													{deal.probability}% probability
													{#if deal.expected_close_date}
														 &middot; Expected close: {formatDate(deal.expected_close_date)}
													{/if}
												</div>
											</div>
											<div class="text-right">
												<div class="text-lg font-semibold text-gray-900">{formatCurrency(deal.value)}</div>
												<div class="text-xs text-gray-500">Created {formatDate(deal.created_at)}</div>
											</div>
										</div>
										{#if deal.notes}
											<p class="text-sm text-gray-600 mt-2">{deal.notes}</p>
										{/if}
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	</div>

	<!-- Add Contact Modal -->
	{#if showAddContactModal}
		<div class="fixed inset-0 z-50 flex items-center justify-center">
			<div class="absolute inset-0 bg-black/50" onclick={() => (showAddContactModal = false)} role="presentation"></div>
			<div class="relative bg-white rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Add Contact</h3>
				<form onsubmit={(e) => { e.preventDefault(); handleAddContact(); }}>
					<div class="space-y-4">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
							<input type="text" bind:value={contactForm.name} required class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
						</div>
						<div class="grid grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
								<input type="email" bind:value={contactForm.email} class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Phone</label>
								<input type="tel" bind:value={contactForm.phone} class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
							</div>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
							<input type="text" bind:value={contactForm.role} placeholder="CEO, CTO, etc." class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
						</div>
						<div class="flex items-center gap-2">
							<input type="checkbox" id="isPrimary" bind:checked={contactForm.is_primary} class="rounded" />
							<label for="isPrimary" class="text-sm text-gray-700">Primary contact</label>
						</div>
					</div>
					<div class="flex justify-end gap-3 mt-6">
						<button type="button" onclick={() => (showAddContactModal = false)} class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg">Cancel</button>
						<button type="submit" class="px-4 py-2 text-sm font-medium bg-gray-900 text-white rounded-lg hover:bg-gray-800">Add Contact</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	<!-- Add Interaction Modal -->
	{#if showAddInteractionModal}
		<div class="fixed inset-0 z-50 flex items-center justify-center">
			<div class="absolute inset-0 bg-black/50" onclick={() => (showAddInteractionModal = false)} role="presentation"></div>
			<div class="relative bg-white rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Log Interaction</h3>
				<form onsubmit={(e) => { e.preventDefault(); handleAddInteraction(); }}>
					<div class="space-y-4">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
							<select bind:value={interactionForm.type} class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900">
								<option value="call">Call</option>
								<option value="email">Email</option>
								<option value="meeting">Meeting</option>
								<option value="note">Note</option>
							</select>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Subject *</label>
							<input type="text" bind:value={interactionForm.subject} required class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
							<textarea bind:value={interactionForm.description} rows="3" class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"></textarea>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Outcome</label>
							<input type="text" bind:value={interactionForm.outcome} placeholder="e.g., Scheduled follow-up call" class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
						</div>
					</div>
					<div class="flex justify-end gap-3 mt-6">
						<button type="button" onclick={() => (showAddInteractionModal = false)} class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg">Cancel</button>
						<button type="submit" class="px-4 py-2 text-sm font-medium bg-gray-900 text-white rounded-lg hover:bg-gray-800">Log Interaction</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	<!-- Add Deal Modal -->
	{#if showAddDealModal}
		<div class="fixed inset-0 z-50 flex items-center justify-center">
			<div class="absolute inset-0 bg-black/50" onclick={() => (showAddDealModal = false)} role="presentation"></div>
			<div class="relative bg-white rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
				<h3 class="text-lg font-semibold text-gray-900 mb-4">Add Deal</h3>
				<form onsubmit={(e) => { e.preventDefault(); handleAddDeal(); }}>
					<div class="space-y-4">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Deal Name *</label>
							<input type="text" bind:value={dealForm.name} required class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
						</div>
						<div class="grid grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Value</label>
								<input type="number" bind:value={dealForm.value} min="0" step="0.01" class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Probability %</label>
								<input type="number" bind:value={dealForm.probability} min="0" max="100" class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
							</div>
						</div>
						<div class="grid grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Stage</label>
								<select bind:value={dealForm.stage} class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900">
									<option value="qualification">Qualification</option>
									<option value="proposal">Proposal</option>
									<option value="negotiation">Negotiation</option>
									<option value="closed_won">Closed Won</option>
									<option value="closed_lost">Closed Lost</option>
								</select>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Expected Close</label>
								<input type="date" bind:value={dealForm.expected_close_date} class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900" />
							</div>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
							<textarea bind:value={dealForm.notes} rows="2" class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"></textarea>
						</div>
					</div>
					<div class="flex justify-end gap-3 mt-6">
						<button type="button" onclick={() => (showAddDealModal = false)} class="px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-lg">Cancel</button>
						<button type="submit" class="px-4 py-2 text-sm font-medium bg-gray-900 text-white rounded-lg hover:bg-gray-800">Add Deal</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
{/if}
