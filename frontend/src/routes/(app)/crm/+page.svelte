<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import {
		crm,
		type CRMViewMode,
		dealStatusColors,
		dealPriorityColors,
		formatCurrency,
		formatProbability
	} from '$lib/stores/crm';
	import type { Pipeline, PipelineStage, Deal, CreateDealData } from '$lib/api/crm';

	// Check if we're in embed mode
	const embedSuffix = $derived(
		$page.url.searchParams.get('embed') === 'true' ? '?embed=true' : ''
	);

	// State from store
	let pipelines = $state<Pipeline[]>([]);
	let currentPipeline = $state<Pipeline | null>(null);
	let stages = $state<PipelineStage[]>([]);
	let deals = $state<Deal[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let viewMode = $state<CRMViewMode>('kanban');
	let dealStats = $state<{ total_deals: number; open_value: number; won_value: number } | null>(
		null
	);

	// Modal state
	let showAddDealModal = $state(false);
	let selectedStageId = $state<string | null>(null);

	// Subscribe to store
	$effect(() => {
		const unsubscribe = crm.subscribe((state) => {
			pipelines = state.pipelines;
			currentPipeline = state.currentPipeline;
			stages = state.stages;
			deals = state.deals;
			loading = state.loading;
			error = state.error;
			viewMode = state.viewMode;
			dealStats = state.dealStats;
		});
		return unsubscribe;
	});

	// Load data on mount
	onMount(() => {
		crm.loadPipelines();
	});

	// Load deals when pipeline changes
	$effect(() => {
		if (currentPipeline) {
			crm.loadDeals({ pipeline_id: currentPipeline.id });
			crm.loadDealStats(currentPipeline.id);
		}
	});

	// Group deals by stage for kanban
	const dealsByStage = $derived(() => {
		const grouped: Record<string, Deal[]> = {};
		for (const stage of stages) {
			grouped[stage.id] = deals.filter((d) => d.stage_id === stage.id);
		}
		return grouped;
	});

	// Calculate stage totals
	function getStageTotal(stageId: string): number {
		return (dealsByStage()[stageId] || []).reduce((sum, d) => sum + (d.amount || 0), 0);
	}

	// Handlers
	function handlePipelineChange(pipeline: Pipeline) {
		crm.selectPipeline(pipeline);
	}

	function handleViewChange(mode: CRMViewMode) {
		crm.setViewMode(mode);
	}

	function handleDealClick(dealId: string) {
		goto(`/crm/deals/${dealId}${embedSuffix}`);
	}

	function handleAddDeal(stageId: string) {
		selectedStageId = stageId;
		showAddDealModal = true;
	}

	async function handleDragDrop(dealId: string, newStageId: string) {
		try {
			await crm.moveDealToStage(dealId, newStageId);
		} catch (err) {
			console.error('Failed to move deal:', err);
		}
	}

	function handleCreateDeal(data: CreateDealData) {
		crm.createDeal(data).then(() => {
			showAddDealModal = false;
			selectedStageId = null;
		});
	}

	// Drag state
	let draggedDealId = $state<string | null>(null);
	let dragOverStageId = $state<string | null>(null);

	function handleDragStart(e: DragEvent, dealId: string) {
		draggedDealId = dealId;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
			e.dataTransfer.setData('text/plain', dealId);
		}
	}

	function handleDragOver(e: DragEvent, stageId: string) {
		e.preventDefault();
		dragOverStageId = stageId;
	}

	function handleDragLeave() {
		dragOverStageId = null;
	}

	function handleDrop(e: DragEvent, stageId: string) {
		e.preventDefault();
		if (draggedDealId && draggedDealId !== stageId) {
			handleDragDrop(draggedDealId, stageId);
		}
		draggedDealId = null;
		dragOverStageId = null;
	}
</script>

<div class="flex flex-col h-full bg-gray-50">
	<!-- Header -->
	<div class="flex items-center justify-between px-6 py-4 bg-white border-b border-gray-200">
		<div class="flex items-center gap-4">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Sales Pipeline</h1>
				<p class="text-sm text-gray-500 mt-0.5">Manage deals and track your sales process</p>
			</div>

			<!-- Pipeline Selector -->
			{#if pipelines.length > 1}
				<select
					class="ml-4 px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					value={currentPipeline?.id || ''}
					onchange={(e) => {
						const pipeline = pipelines.find((p) => p.id === e.currentTarget.value);
						if (pipeline) handlePipelineChange(pipeline);
					}}
				>
					{#each pipelines as pipeline}
						<option value={pipeline.id}>{pipeline.name}</option>
					{/each}
				</select>
			{/if}
		</div>

		<div class="flex items-center gap-3">
			<!-- Stats -->
			{#if dealStats}
				<div class="flex items-center gap-4 mr-4 text-sm">
					<div class="text-gray-500">
						<span class="font-medium text-gray-900">{dealStats.total_deals}</span> deals
					</div>
					<div class="text-gray-500">
						Open: <span class="font-medium text-blue-600"
							>{formatCurrency(dealStats.open_value)}</span
						>
					</div>
					<div class="text-gray-500">
						Won: <span class="font-medium text-emerald-600"
							>{formatCurrency(dealStats.won_value)}</span
						>
					</div>
				</div>
			{/if}

			<!-- View Switcher -->
			<div class="btn-pill-group">
				<button
					onclick={() => handleViewChange('kanban')}
					class="btn-pill {viewMode === 'kanban' ? 'btn-pill-primary' : 'btn-pill-secondary'} btn-pill-sm"
				>
					Kanban
				</button>
				<button
					onclick={() => handleViewChange('list')}
					class="btn-pill {viewMode === 'list' ? 'btn-pill-primary' : 'btn-pill-secondary'} btn-pill-sm"
				>
					List
				</button>
			</div>

			<!-- Add Deal Button -->
			<button
				onclick={() => handleAddDeal(stages[0]?.id)}
				class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"
					></path>
				</svg>
				Add Deal
			</button>
		</div>
	</div>

	<!-- Error State -->
	{#if error}
		<div class="mx-6 mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
			<p class="text-sm text-red-700">{error}</p>
			<button
				onclick={() => crm.loadPipelines()}
				class="btn-pill btn-pill-danger btn-pill-sm mt-2"
			>
				Try again
			</button>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading && stages.length === 0}
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
				<p class="text-sm">Loading pipeline...</p>
			</div>
		</div>
	{:else if viewMode === 'kanban'}
		<!-- Kanban View -->
		<div class="flex-1 overflow-x-auto p-6">
			<div class="flex gap-4 h-full min-w-max">
				{#each stages as stage}
					{@const stageDeals = dealsByStage()[stage.id] || []}
					{@const stageTotal = getStageTotal(stage.id)}
					<div
						class="w-72 flex flex-col bg-gray-100 rounded-lg {dragOverStageId === stage.id
							? 'ring-2 ring-blue-500'
							: ''}"
						ondragover={(e) => handleDragOver(e, stage.id)}
						ondragleave={handleDragLeave}
						ondrop={(e) => handleDrop(e, stage.id)}
						role="region"
						aria-label={stage.name}
					>
						<!-- Stage Header -->
						<div class="flex items-center justify-between p-3 border-b border-gray-200">
							<div class="flex items-center gap-2">
								{#if stage.color}
									<div
										class="w-2 h-2 rounded-full"
										style="background-color: {stage.color}"
									></div>
								{/if}
								<h3 class="font-medium text-gray-900">{stage.name}</h3>
								<span
									class="px-1.5 py-0.5 text-xs font-medium bg-gray-200 text-gray-600 rounded"
								>
									{stageDeals.length}
								</span>
							</div>
							<div class="text-xs text-gray-500">{formatCurrency(stageTotal)}</div>
						</div>

						<!-- Deals -->
						<div class="flex-1 overflow-y-auto p-2 space-y-2">
							{#each stageDeals as deal}
								<div
									class="bg-white rounded-lg border border-gray-200 p-3 cursor-pointer hover:shadow-md transition-shadow"
									draggable="true"
									ondragstart={(e) => handleDragStart(e, deal.id)}
									onclick={() => handleDealClick(deal.id)}
									role="button"
									tabindex="0"
									onkeydown={(e) => e.key === 'Enter' && handleDealClick(deal.id)}
								>
									<div class="flex items-start justify-between mb-2">
										<h4 class="font-medium text-gray-900 text-sm">{deal.name}</h4>
										{#if deal.priority}
											<span
												class="px-1.5 py-0.5 text-xs rounded {dealPriorityColors[deal.priority]}"
											>
												{deal.priority}
											</span>
										{/if}
									</div>

									{#if deal.company_name}
										<p class="text-xs text-gray-500 mb-2">{deal.company_name}</p>
									{/if}

									<div class="flex items-center justify-between text-sm">
										<span class="font-medium text-gray-900">
											{formatCurrency(deal.amount, deal.currency)}
										</span>
										{#if deal.probability !== undefined}
											<span class="text-xs text-gray-500">
												{formatProbability(deal.probability)}
											</span>
										{/if}
									</div>

									{#if deal.expected_close_date}
										<div class="mt-2 text-xs text-gray-400">
											Close: {new Date(deal.expected_close_date).toLocaleDateString()}
										</div>
									{/if}
								</div>
							{/each}

							<!-- Add Deal to Stage -->
							<button
								onclick={() => handleAddDeal(stage.id)}
								class="btn-pill btn-pill-ghost btn-pill-sm w-full flex items-center justify-center gap-1"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M12 4v16m8-8H4"
									></path>
								</svg>
								Add Deal
							</button>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{:else}
		<!-- List View -->
		<div class="flex-1 overflow-auto p-6">
			<div class="bg-white rounded-lg border border-gray-200">
				<table class="w-full">
					<thead class="bg-gray-50 border-b border-gray-200">
						<tr>
							<th class="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Deal</th
							>
							<th class="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase"
								>Company</th
							>
							<th class="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Stage</th
							>
							<th class="text-right px-4 py-3 text-xs font-medium text-gray-500 uppercase"
								>Amount</th
							>
							<th class="text-center px-4 py-3 text-xs font-medium text-gray-500 uppercase"
								>Probability</th
							>
							<th class="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase"
								>Close Date</th
							>
							<th class="text-left px-4 py-3 text-xs font-medium text-gray-500 uppercase">Status</th
							>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each deals as deal}
							<tr
								class="hover:bg-gray-50 cursor-pointer"
								onclick={() => handleDealClick(deal.id)}
							>
								<td class="px-4 py-3">
									<div class="font-medium text-gray-900">{deal.name}</div>
								</td>
								<td class="px-4 py-3 text-sm text-gray-600">{deal.company_name || '-'}</td>
								<td class="px-4 py-3 text-sm text-gray-600">{deal.stage_name || '-'}</td>
								<td class="px-4 py-3 text-sm text-gray-900 text-right font-medium">
									{formatCurrency(deal.amount, deal.currency)}
								</td>
								<td class="px-4 py-3 text-sm text-gray-600 text-center">
									{formatProbability(deal.probability)}
								</td>
								<td class="px-4 py-3 text-sm text-gray-600">
									{deal.expected_close_date
										? new Date(deal.expected_close_date).toLocaleDateString()
										: '-'}
								</td>
								<td class="px-4 py-3">
									{#if deal.status}
										<span
											class="px-2 py-1 text-xs font-medium rounded-full {dealStatusColors[
												deal.status
											]}"
										>
											{deal.status}
										</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>

				{#if deals.length === 0 && !loading}
					<div class="p-8 text-center text-gray-500">
						<p>No deals in this pipeline yet.</p>
						<button
							onclick={() => handleAddDeal(stages[0]?.id)}
							class="btn-pill btn-pill-primary btn-pill-sm mt-2"
						>
							Add your first deal
						</button>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Add Deal Modal -->
{#if showAddDealModal && currentPipeline && selectedStageId}
	<div
		class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
		onclick={() => (showAddDealModal = false)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-md p-6"
			onclick={(e) => e.stopPropagation()}
			role="document"
		>
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Create New Deal</h2>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					const formData = new FormData(e.currentTarget);
					handleCreateDeal({
						pipeline_id: currentPipeline!.id,
						stage_id: selectedStageId!,
						name: formData.get('name') as string,
						amount: formData.get('amount') ? Number(formData.get('amount')) : undefined,
						expected_close_date: formData.get('expected_close_date') as string || undefined
					});
				}}
			>
				<div class="space-y-4">
					<div>
						<label for="deal-name" class="block text-sm font-medium text-gray-700 mb-1"
							>Deal Name</label
						>
						<input
							id="deal-name"
							name="name"
							type="text"
							required
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							placeholder="e.g., Enterprise License Q1"
						/>
					</div>

					<div>
						<label for="deal-amount" class="block text-sm font-medium text-gray-700 mb-1"
							>Amount</label
						>
						<input
							id="deal-amount"
							name="amount"
							type="number"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							placeholder="50000"
						/>
					</div>

					<div>
						<label for="deal-close-date" class="block text-sm font-medium text-gray-700 mb-1"
							>Expected Close Date</label
						>
						<input
							id="deal-close-date"
							name="expected_close_date"
							type="date"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
						/>
					</div>
				</div>

				<div class="flex justify-end gap-3 mt-6">
					<button
						type="button"
						onclick={() => (showAddDealModal = false)}
						class="btn-pill btn-pill-secondary"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="btn-pill btn-pill-primary"
					>
						Create Deal
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
