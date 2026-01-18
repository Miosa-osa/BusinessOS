<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { crm, lifecycleStageColors, lifecycleStageLabels, formatCurrency } from '$lib/stores/crm';
	import type { Company, CreateCompanyData } from '$lib/api/crm';

	// Check if we're in embed mode
	const embedSuffix = $derived(
		$page.url.searchParams.get('embed') === 'true' ? '?embed=true' : ''
	);

	// State from store
	let companies = $state<Company[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let searchQuery = $state('');

	// Modal state
	let showAddModal = $state(false);

	// Subscribe to store
	$effect(() => {
		const unsubscribe = crm.subscribe((state) => {
			companies = state.companies;
			loading = state.loading;
			error = state.error;
		});
		return unsubscribe;
	});

	// Load companies on mount
	onMount(() => {
		crm.loadCompanies();
	});

	// Filtered companies
	const filteredCompanies = $derived(() => {
		if (!searchQuery) return companies;
		const query = searchQuery.toLowerCase();
		return companies.filter(
			(c) =>
				c.name.toLowerCase().includes(query) ||
				(c.industry && c.industry.toLowerCase().includes(query)) ||
				(c.email && c.email.toLowerCase().includes(query))
		);
	});

	// Handlers
	function handleCompanyClick(id: string) {
		goto(`/crm/companies/${id}${embedSuffix}`);
	}

	async function handleCreateCompany(data: CreateCompanyData) {
		try {
			const company = await crm.createCompany(data);
			showAddModal = false;
			goto(`/crm/companies/${company.id}`);
		} catch (err) {
			console.error('Failed to create company:', err);
		}
	}
</script>

<div class="flex flex-col h-full bg-white">
	<!-- Header -->
	<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
		<div>
			<h1 class="text-2xl font-semibold text-gray-900">Companies</h1>
			<p class="text-sm text-gray-500 mt-0.5">Manage organizations in your CRM</p>
		</div>
		<div class="flex items-center gap-3">
			<a
				href="/crm{embedSuffix}"
				class="btn-pill btn-pill-secondary btn-pill-sm"
			>
				Back to Pipeline
			</a>
			<button
				onclick={() => (showAddModal = true)}
				class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"
					></path>
				</svg>
				Add Company
			</button>
		</div>
	</div>

	<!-- Search -->
	<div class="px-6 py-3 border-b border-gray-100">
		<div class="relative max-w-md">
			<svg
				class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
				></path>
			</svg>
			<input
				type="text"
				placeholder="Search companies..."
				bind:value={searchQuery}
				class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
			/>
		</div>
	</div>

	<!-- Error State -->
	{#if error}
		<div class="mx-6 mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
			<p class="text-sm text-red-700">{error}</p>
			<button
				onclick={() => crm.loadCompanies()}
				class="btn-pill btn-pill-danger btn-pill-sm mt-2"
			>
				Try again
			</button>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading && companies.length === 0}
		<div class="flex-1 flex items-center justify-center">
			<div class="flex flex-col items-center gap-3 text-gray-500">
				<svg class="w-8 h-8 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
					></path>
				</svg>
				<p class="text-sm">Loading companies...</p>
			</div>
		</div>
	{:else if filteredCompanies().length === 0 && !loading}
		<!-- Empty State -->
		<div class="flex-1 flex items-center justify-center">
			<div class="flex flex-col items-center gap-3 text-gray-500">
				<svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
					></path>
				</svg>
				<p class="text-lg font-medium text-gray-900">No companies yet</p>
				<p class="text-sm">Add your first company to get started</p>
				<button
					onclick={() => (showAddModal = true)}
					class="btn-pill btn-pill-primary mt-2"
				>
					Add Company
				</button>
			</div>
		</div>
	{:else}
		<!-- Companies Grid -->
		<div class="flex-1 overflow-auto p-6">
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each filteredCompanies() as company}
					<button
						onclick={() => handleCompanyClick(company.id)}
						class="text-left bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
					>
						<div class="flex items-start gap-3">
							{#if company.logo_url}
								<img
									src={company.logo_url}
									alt={company.name}
									class="w-10 h-10 rounded object-cover"
								/>
							{:else}
								<div
									class="w-10 h-10 rounded bg-gray-100 flex items-center justify-center text-gray-500 font-medium"
								>
									{company.name.charAt(0).toUpperCase()}
								</div>
							{/if}
							<div class="flex-1 min-w-0">
								<h3 class="font-medium text-gray-900 truncate">{company.name}</h3>
								{#if company.industry}
									<p class="text-sm text-gray-500 truncate">{company.industry}</p>
								{/if}
							</div>
						</div>

						<div class="mt-3 flex items-center justify-between">
							{#if company.lifecycle_stage}
								<span
									class="px-2 py-0.5 text-xs font-medium rounded {lifecycleStageColors[
										company.lifecycle_stage
									] || 'bg-gray-100 text-gray-600'}"
								>
									{lifecycleStageLabels[company.lifecycle_stage] || company.lifecycle_stage}
								</span>
							{:else}
								<span></span>
							{/if}
							{#if company.annual_revenue}
								<span class="text-sm text-gray-600">
									{formatCurrency(company.annual_revenue, company.currency)}
								</span>
							{/if}
						</div>

						{#if company.website || company.email}
							<div class="mt-2 pt-2 border-t border-gray-100 text-xs text-gray-500 truncate">
								{company.website || company.email}
							</div>
						{/if}
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Add Company Modal -->
{#if showAddModal}
	<div
		class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
		onclick={() => (showAddModal = false)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto"
			onclick={(e) => e.stopPropagation()}
			role="document"
		>
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Add New Company</h2>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					const formData = new FormData(e.currentTarget);
					handleCreateCompany({
						name: formData.get('name') as string,
						industry: (formData.get('industry') as string) || undefined,
						website: (formData.get('website') as string) || undefined,
						email: (formData.get('email') as string) || undefined,
						phone: (formData.get('phone') as string) || undefined,
						lifecycle_stage: (formData.get('lifecycle_stage') as string) || undefined
					});
				}}
			>
				<div class="space-y-4">
					<div>
						<label for="company-name" class="block text-sm font-medium text-gray-700 mb-1"
							>Company Name *</label
						>
						<input
							id="company-name"
							name="name"
							type="text"
							required
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							placeholder="Acme Inc."
						/>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<label for="company-industry" class="block text-sm font-medium text-gray-700 mb-1"
								>Industry</label
							>
							<input
								id="company-industry"
								name="industry"
								type="text"
								class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								placeholder="Technology"
							/>
						</div>
						<div>
							<label for="company-stage" class="block text-sm font-medium text-gray-700 mb-1"
								>Lifecycle Stage</label
							>
							<select
								id="company-stage"
								name="lifecycle_stage"
								class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							>
								<option value="">Select stage...</option>
								<option value="lead">Lead</option>
								<option value="opportunity">Opportunity</option>
								<option value="customer">Customer</option>
								<option value="partner">Partner</option>
							</select>
						</div>
					</div>

					<div>
						<label for="company-website" class="block text-sm font-medium text-gray-700 mb-1"
							>Website</label
						>
						<input
							id="company-website"
							name="website"
							type="url"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							placeholder="https://example.com"
						/>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<label for="company-email" class="block text-sm font-medium text-gray-700 mb-1"
								>Email</label
							>
							<input
								id="company-email"
								name="email"
								type="email"
								class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								placeholder="contact@example.com"
							/>
						</div>
						<div>
							<label for="company-phone" class="block text-sm font-medium text-gray-700 mb-1"
								>Phone</label
							>
							<input
								id="company-phone"
								name="phone"
								type="tel"
								class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								placeholder="+1 (555) 123-4567"
							/>
						</div>
					</div>
				</div>

				<div class="flex justify-end gap-3 mt-6">
					<button
						type="button"
						onclick={() => (showAddModal = false)}
						class="btn-pill btn-pill-secondary"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="btn-pill btn-pill-primary"
					>
						Create Company
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
