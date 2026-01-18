<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import {
		crm,
		dealStatusColors,
		dealStatusLabels,
		dealPriorityColors,
		dealPriorityLabels,
		activityTypeColors,
		activityTypeLabels,
		formatCurrency,
		formatProbability
	} from '$lib/stores/crm';
	import type { Deal, CRMActivity, CreateActivityData, ActivityType } from '$lib/api/crm';

	// Get deal ID from route
	const dealId = $derived($page.params.id);

	// Check if we're in embed mode
	const embedSuffix = $derived(
		$page.url.searchParams.get('embed') === 'true' ? '?embed=true' : ''
	);

	// State from store
	let deal = $state<Deal | null>(null);
	let activities = $state<CRMActivity[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);

	// Modal state
	let showAddActivityModal = $state(false);
	let showStatusModal = $state(false);

	// Subscribe to store
	$effect(() => {
		const unsubscribe = crm.subscribe((state) => {
			deal = state.currentDeal;
			activities = state.activities;
			loading = state.loading;
			error = state.error;
		});
		return unsubscribe;
	});

	// Load deal on mount and when ID changes
	$effect(() => {
		if (dealId) {
			crm.loadDeal(dealId);
		}
	});

	// Cleanup on unmount
	onMount(() => {
		return () => {
			crm.clearCurrentDeal();
		};
	});

	// Handlers
	function handleBack() {
		goto(`/crm${embedSuffix}`);
	}

	async function handleUpdateStatus(status: string, lostReason?: string) {
		if (!deal) return;
		try {
			await crm.updateDealStatus(deal.id, status, lostReason);
			showStatusModal = false;
		} catch (err) {
			console.error('Failed to update status:', err);
		}
	}

	async function handleCreateActivity(data: CreateActivityData) {
		try {
			await crm.createActivity(data);
			showAddActivityModal = false;
		} catch (err) {
			console.error('Failed to create activity:', err);
		}
	}

	async function handleCompleteActivity(activityId: string) {
		try {
			await crm.completeActivity(activityId);
		} catch (err) {
			console.error('Failed to complete activity:', err);
		}
	}

	// Format date
	function formatDate(dateStr: string | undefined): string {
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
			hour: 'numeric',
			minute: '2-digit'
		});
	}
</script>

<div class="flex flex-col h-full bg-gray-50">
	<!-- Header -->
	<div class="flex items-center justify-between px-6 py-4 bg-white border-b border-gray-200">
		<div class="flex items-center gap-4">
			<button
				onclick={handleBack}
				class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M15 19l-7-7 7-7"
					></path>
				</svg>
			</button>
			{#if deal}
				<div>
					<h1 class="text-xl font-semibold text-gray-900">{deal.name}</h1>
					<p class="text-sm text-gray-500">
						{deal.company_name || 'No company'} &bull; {deal.pipeline_name}
					</p>
				</div>
			{/if}
		</div>

		{#if deal}
			<div class="flex items-center gap-3">
				{#if deal.status}
					<span class="px-3 py-1.5 text-sm font-medium rounded-full {dealStatusColors[deal.status]}">
						{dealStatusLabels[deal.status] || deal.status}
					</span>
				{/if}
				<button
					onclick={() => (showStatusModal = true)}
					class="btn-pill btn-pill-secondary btn-pill-sm"
				>
					Update Status
				</button>
				<button
					onclick={() => (showAddActivityModal = true)}
					class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 4v16m8-8H4"
						></path>
					</svg>
					Log Activity
				</button>
			</div>
		{/if}
	</div>

	<!-- Loading/Error States -->
	{#if loading && !deal}
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
				<p class="text-sm">Loading deal...</p>
			</div>
		</div>
	{:else if error}
		<div class="m-6 p-4 bg-red-50 border border-red-200 rounded-lg">
			<p class="text-sm text-red-700">{error}</p>
			<button
				onclick={handleBack}
				class="btn-pill btn-pill-danger btn-pill-sm mt-2"
			>
				Go back
			</button>
		</div>
	{:else if deal}
		<!-- Content -->
		<div class="flex-1 overflow-auto p-6">
			<div class="max-w-4xl mx-auto grid grid-cols-3 gap-6">
				<!-- Main Info -->
				<div class="col-span-2 space-y-6">
					<!-- Deal Details Card -->
					<div class="bg-white rounded-lg border border-gray-200 p-6">
						<h2 class="text-lg font-medium text-gray-900 mb-4">Deal Details</h2>
						<dl class="grid grid-cols-2 gap-4">
							<div>
								<dt class="text-sm text-gray-500">Amount</dt>
								<dd class="text-lg font-semibold text-gray-900">
									{formatCurrency(deal.amount, deal.currency)}
								</dd>
							</div>
							<div>
								<dt class="text-sm text-gray-500">Probability</dt>
								<dd class="text-lg font-semibold text-gray-900">
									{formatProbability(deal.probability)}
								</dd>
							</div>
							<div>
								<dt class="text-sm text-gray-500">Stage</dt>
								<dd class="text-gray-900">{deal.stage_name || '-'}</dd>
							</div>
							<div>
								<dt class="text-sm text-gray-500">Priority</dt>
								<dd>
									{#if deal.priority}
										<span class="px-2 py-1 text-xs rounded {dealPriorityColors[deal.priority]}">
											{dealPriorityLabels[deal.priority] || deal.priority}
										</span>
									{:else}
										<span class="text-gray-400">-</span>
									{/if}
								</dd>
							</div>
							<div>
								<dt class="text-sm text-gray-500">Expected Close</dt>
								<dd class="text-gray-900">{formatDate(deal.expected_close_date)}</dd>
							</div>
							<div>
								<dt class="text-sm text-gray-500">Lead Source</dt>
								<dd class="text-gray-900">{deal.lead_source || '-'}</dd>
							</div>
						</dl>

						{#if deal.description}
							<div class="mt-4 pt-4 border-t border-gray-100">
								<dt class="text-sm text-gray-500 mb-1">Description</dt>
								<dd class="text-gray-700 whitespace-pre-wrap">{deal.description}</dd>
							</div>
						{/if}
					</div>

					<!-- Activities Timeline -->
					<div class="bg-white rounded-lg border border-gray-200 p-6">
						<div class="flex items-center justify-between mb-4">
							<h2 class="text-lg font-medium text-gray-900">Activity Timeline</h2>
							<button
								onclick={() => (showAddActivityModal = true)}
								class="btn-pill btn-pill-primary btn-pill-xs"
							>
								+ Add Activity
							</button>
						</div>

						{#if activities.length === 0}
							<div class="text-center py-8 text-gray-500">
								<p>No activities yet.</p>
								<button
									onclick={() => (showAddActivityModal = true)}
									class="btn-pill btn-pill-primary btn-pill-sm mt-2"
								>
									Log your first activity
								</button>
							</div>
						{:else}
							<div class="space-y-4">
								{#each activities as activity}
									<div class="flex gap-3">
										<div
											class="flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center {activityTypeColors[
												activity.activity_type
											]}"
										>
											{#if activity.activity_type === 'call'}
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
													></path>
												</svg>
											{:else if activity.activity_type === 'email'}
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
													></path>
												</svg>
											{:else if activity.activity_type === 'meeting'}
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
													></path>
												</svg>
											{:else}
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path
														stroke-linecap="round"
														stroke-linejoin="round"
														stroke-width="2"
														d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
													></path>
												</svg>
											{/if}
										</div>
										<div class="flex-1 min-w-0">
											<div class="flex items-center justify-between">
												<p class="font-medium text-gray-900">{activity.subject}</p>
												<span class="text-xs text-gray-500">
													{formatDateTime(activity.activity_date)}
												</span>
											</div>
											{#if activity.description}
												<p class="text-sm text-gray-600 mt-1">{activity.description}</p>
											{/if}
											<div class="flex items-center gap-2 mt-2">
												<span
													class="text-xs px-2 py-0.5 rounded {activityTypeColors[
														activity.activity_type
													]}"
												>
													{activityTypeLabels[activity.activity_type] || activity.activity_type}
												</span>
												{#if activity.is_completed}
													<span class="text-xs text-emerald-600">Completed</span>
												{:else}
													<button
														onclick={() => handleCompleteActivity(activity.id)}
														class="btn-pill btn-pill-primary btn-pill-xs"
													>
														Mark Complete
													</button>
												{/if}
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>

				<!-- Sidebar -->
				<div class="space-y-6">
					<!-- Quick Stats -->
					<div class="bg-white rounded-lg border border-gray-200 p-4">
						<h3 class="text-sm font-medium text-gray-500 mb-3">Deal Score</h3>
						<div class="text-3xl font-bold text-gray-900">{deal.deal_score || '-'}</div>
					</div>

					<!-- Dates -->
					<div class="bg-white rounded-lg border border-gray-200 p-4">
						<h3 class="text-sm font-medium text-gray-500 mb-3">Important Dates</h3>
						<dl class="space-y-2 text-sm">
							<div class="flex justify-between">
								<dt class="text-gray-500">Created</dt>
								<dd class="text-gray-900">{formatDate(deal.created_at)}</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-gray-500">Last Updated</dt>
								<dd class="text-gray-900">{formatDate(deal.updated_at)}</dd>
							</div>
							{#if deal.actual_close_date}
								<div class="flex justify-between">
									<dt class="text-gray-500">Closed</dt>
									<dd class="text-gray-900">{formatDate(deal.actual_close_date)}</dd>
								</div>
							{/if}
						</dl>
					</div>

					<!-- Company Link -->
					{#if deal.company_id}
						<div class="bg-white rounded-lg border border-gray-200 p-4">
							<h3 class="text-sm font-medium text-gray-500 mb-2">Company</h3>
							<a
								href="/crm/companies/{deal.company_id}{embedSuffix}"
								class="btn-pill btn-pill-link btn-pill-sm"
							>
								{deal.company_name || 'View Company'}
							</a>
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Update Status Modal -->
{#if showStatusModal && deal}
	<div
		class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
		onclick={() => (showStatusModal = false)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm p-6"
			onclick={(e) => e.stopPropagation()}
			role="document"
		>
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Update Deal Status</h2>
			<div class="space-y-3">
				<button
					onclick={() => handleUpdateStatus('open')}
					class="w-full p-3 text-left rounded-lg border hover:bg-gray-50 {deal.status === 'open'
						? 'border-blue-500 bg-blue-50'
						: 'border-gray-200'}"
				>
					<span class="font-medium">Open</span>
					<p class="text-sm text-gray-500">Deal is in progress</p>
				</button>
				<button
					onclick={() => handleUpdateStatus('won')}
					class="w-full p-3 text-left rounded-lg border hover:bg-gray-50 {deal.status === 'won'
						? 'border-emerald-500 bg-emerald-50'
						: 'border-gray-200'}"
				>
					<span class="font-medium text-emerald-700">Won</span>
					<p class="text-sm text-gray-500">Deal was closed successfully</p>
				</button>
				<button
					onclick={() => handleUpdateStatus('lost', 'Lost to competitor')}
					class="w-full p-3 text-left rounded-lg border hover:bg-gray-50 {deal.status === 'lost'
						? 'border-red-500 bg-red-50'
						: 'border-gray-200'}"
				>
					<span class="font-medium text-red-700">Lost</span>
					<p class="text-sm text-gray-500">Deal was not successful</p>
				</button>
			</div>
			<button
				onclick={() => (showStatusModal = false)}
				class="btn-pill btn-pill-secondary w-full mt-4"
			>
				Cancel
			</button>
		</div>
	</div>
{/if}

<!-- Add Activity Modal -->
{#if showAddActivityModal && deal}
	<div
		class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
		onclick={() => (showAddActivityModal = false)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-md p-6"
			onclick={(e) => e.stopPropagation()}
			role="document"
		>
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Log Activity</h2>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					const formData = new FormData(e.currentTarget);
					handleCreateActivity({
						activity_type: formData.get('activity_type') as string,
						subject: formData.get('subject') as string,
						description: (formData.get('description') as string) || undefined,
						deal_id: deal!.id,
						activity_date: new Date().toISOString()
					});
				}}
			>
				<div class="space-y-4">
					<div>
						<label for="activity-type" class="block text-sm font-medium text-gray-700 mb-1"
							>Activity Type</label
						>
						<select
							id="activity-type"
							name="activity_type"
							required
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						>
							<option value="call">Call</option>
							<option value="email">Email</option>
							<option value="meeting">Meeting</option>
							<option value="demo">Demo</option>
							<option value="note">Note</option>
							<option value="task">Task</option>
						</select>
					</div>

					<div>
						<label for="activity-subject" class="block text-sm font-medium text-gray-700 mb-1"
							>Subject</label
						>
						<input
							id="activity-subject"
							name="subject"
							type="text"
							required
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							placeholder="e.g., Follow-up call with decision maker"
						/>
					</div>

					<div>
						<label for="activity-description" class="block text-sm font-medium text-gray-700 mb-1"
							>Description</label
						>
						<textarea
							id="activity-description"
							name="description"
							rows="3"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							placeholder="Add notes about this activity..."
						></textarea>
					</div>
				</div>

				<div class="flex justify-end gap-3 mt-6">
					<button
						type="button"
						onclick={() => (showAddActivityModal = false)}
						class="btn-pill btn-pill-secondary"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="btn-pill btn-pill-primary"
					>
						Log Activity
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
