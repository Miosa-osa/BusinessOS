<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { fade, scale } from 'svelte/transition';
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

	// Reactive store access via auto-subscription
	let pipelines = $derived($crm.pipelines);
	let currentPipeline = $derived($crm.currentPipeline);
	let stages = $derived($crm.stages);
	let deals = $derived($crm.deals);
	let loading = $derived($crm.loading);
	let error = $derived($crm.error);
	let viewMode = $derived($crm.viewMode);
	let dealStats = $derived($crm.dealStats);

	// Modal state
	let showAddDealModal = $state(false);
	let selectedStageId = $state<string | null>(null);

	// Load data on mount
	onMount(async () => {
		try {
			await crm.loadPipelines();
		} catch {
			// Backend unavailable — empty state will show
		}
	});

	// Load deals when pipeline changes
	let pipelineLoadedFor = $state<string | null>(null);
	$effect(() => {
		if (currentPipeline && currentPipeline.id !== pipelineLoadedFor) {
			pipelineLoadedFor = currentPipeline.id;
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

<div class="cr-page">
	<!-- Header -->
	<div class="cr-page__header">
		<div class="cr-page__header-left">
			<div>
				<h1 class="cr-page__title">Sales Pipeline</h1>
				<p class="cr-page__subtitle">Manage deals and track your sales process</p>
			</div>

			<!-- Pipeline Selector -->
			{#if pipelines.length > 1}
				<select
					class="cr-pipeline-select"
					value={currentPipeline?.id || ''}
					onchange={(e) => {
						const pipeline = pipelines.find((p) => p.id === e.currentTarget.value);
						if (pipeline) handlePipelineChange(pipeline);
					}}
					aria-label="Select pipeline"
				>
					{#each pipelines as pipeline}
						<option value={pipeline.id}>{pipeline.name}</option>
					{/each}
				</select>
			{/if}
		</div>

		<div class="cr-page__header-right">
			<!-- Stats -->
			{#if dealStats}
				<div class="cr-stats-strip">
					<div class="cr-stat">
						<span class="cr-stat__value">{dealStats.total_deals}</span>
						<span class="cr-stat__label">deals</span>
					</div>
					<div class="cr-stat">
						<span class="cr-stat__label">Open:</span>
						<span class="cr-stat__value cr-stat__value--open">{formatCurrency(dealStats.open_value)}</span>
					</div>
					<div class="cr-stat">
						<span class="cr-stat__label">Won:</span>
						<span class="cr-stat__value cr-stat__value--won">{formatCurrency(dealStats.won_value)}</span>
					</div>
				</div>
			{/if}

			<!-- View Switcher -->
			<div class="cr-view-switcher">
				<button
					onclick={() => handleViewChange('kanban')}
					class="cr-view-switcher__tab {viewMode === 'kanban' ? 'cr-view-switcher__tab--active' : ''}"
					aria-label="Kanban view"
					aria-pressed={viewMode === 'kanban'}
				>
					Kanban
				</button>
				<button
					onclick={() => handleViewChange('list')}
					class="cr-view-switcher__tab {viewMode === 'list' ? 'cr-view-switcher__tab--active' : ''}"
					aria-label="List view"
					aria-pressed={viewMode === 'list'}
				>
					List
				</button>
			</div>

			<!-- Add Deal Button -->
			<button
				onclick={() => handleAddDeal(stages[0]?.id)}
				class="btn-rounded btn-rounded-primary"
				aria-label="Add new deal"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
				</svg>
				Add Deal
			</button>
		</div>
	</div>

	<!-- Error State -->
	{#if error}
		<div class="cr-page__error">
			<p class="cr-page__error-text">{error}</p>
			<button onclick={() => crm.loadPipelines()} class="cr-page__error-retry" aria-label="Retry loading">
				Try again
			</button>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading && stages.length === 0}
		<div class="cr-page__center">
			<div class="cr-page__center-content">
				<svg class="w-8 h-8 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
				</svg>
				<p class="cr-page__center-text">Loading pipeline...</p>
			</div>
		</div>
	{:else if viewMode === 'kanban'}
		<!-- Kanban View -->
		<div class="cr-kanban-container">
			<div class="cr-kanban">
				{#each stages as stage}
					{@const stageDeals = dealsByStage()[stage.id] || []}
					{@const stageTotal = getStageTotal(stage.id)}
					<div
						class="cr-kanban-col {dragOverStageId === stage.id ? 'cr-kanban-col--dragover' : ''}"
						ondragover={(e) => handleDragOver(e, stage.id)}
						ondragleave={handleDragLeave}
						ondrop={(e) => handleDrop(e, stage.id)}
						role="region"
						aria-label={stage.name}
					>
						<!-- Stage Header -->
						<div class="cr-kanban-col__header">
							{#if stage.color}
								<span class="cr-kanban-col__dot" style="background: {stage.color}"></span>
							{/if}
							<span class="cr-kanban-col__stage">{stage.name}</span>
							<span class="cr-kanban-col__count">{stageDeals.length}</span>
						</div>
						<div class="cr-kanban-col__total">{formatCurrency(stageTotal)}</div>

						<!-- Deals -->
						<div class="cr-kanban-col__cards">
							{#each stageDeals as deal}
								<div
									class="cr-kanban-card"
									style="--stage-color: {stage.color || 'var(--dbd2)'}"
									draggable="true"
									ondragstart={(e) => handleDragStart(e, deal.id)}
									onclick={() => handleDealClick(deal.id)}
									role="button"
									tabindex="0"
									onkeydown={(e) => e.key === 'Enter' && handleDealClick(deal.id)}
									aria-label="View deal: {deal.name}"
								>
									<div class="cr-kanban-card__bar"></div>
									<div class="cr-kanban-card__body">
										<div class="cr-kanban-card__header-row">
											<span class="cr-kanban-card__name">{deal.name}</span>
											{#if deal.priority}
												<span class="cr-priority-pill cr-priority-pill--{deal.priority}">{deal.priority}</span>
											{/if}
										</div>

										{#if deal.company_name}
											<div class="cr-kanban-card__contact">{deal.company_name}</div>
										{/if}

										<div class="cr-kanban-card__meta">
											<span class="cr-kanban-card__value">{formatCurrency(deal.amount, deal.currency)}</span>
											{#if deal.probability !== undefined}
												<span class="cr-kanban-card__prob">{formatProbability(deal.probability)}</span>
											{/if}
										</div>

										{#if deal.expected_close_date}
											<div class="cr-kanban-card__date">
												Close: {new Date(deal.expected_close_date).toLocaleDateString()}
											</div>
										{/if}
									</div>
								</div>
							{/each}

							<!-- Add Deal to Stage -->
							<button
								onclick={() => handleAddDeal(stage.id)}
								class="cr-kanban-col__add-btn"
								aria-label="Add deal to {stage.name}"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
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
		<div class="cr-list-container">
			<div class="cr-table-wrap">
				<table class="cr-table">
					<thead>
						<tr class="cr-table__head-row">
							<th class="cr-table__th">Deal</th>
							<th class="cr-table__th">Company</th>
							<th class="cr-table__th">Stage</th>
							<th class="cr-table__th cr-table__th--right">Amount</th>
							<th class="cr-table__th cr-table__th--center">Probability</th>
							<th class="cr-table__th">Close Date</th>
							<th class="cr-table__th">Status</th>
						</tr>
					</thead>
					<tbody>
						{#each deals as deal}
							<tr
								class="cr-table__row"
								onclick={() => handleDealClick(deal.id)}
								role="button"
								tabindex="0"
								onkeydown={(e) => e.key === 'Enter' && handleDealClick(deal.id)}
							>
								<td class="cr-table__td cr-table__td--name">{deal.name}</td>
								<td class="cr-table__td cr-table__td--muted">{deal.company_name || '-'}</td>
								<td class="cr-table__td cr-table__td--muted">{deal.stage_name || '-'}</td>
								<td class="cr-table__td cr-table__td--right cr-table__td--name">
									{formatCurrency(deal.amount, deal.currency)}
								</td>
								<td class="cr-table__td cr-table__td--center cr-table__td--muted">
									{formatProbability(deal.probability)}
								</td>
								<td class="cr-table__td cr-table__td--muted">
									{deal.expected_close_date
										? new Date(deal.expected_close_date).toLocaleDateString()
										: '-'}
								</td>
								<td class="cr-table__td">
									{#if deal.status}
										<span class="cr-status-pill cr-status-pill--{deal.status}">
											{deal.status}
										</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>

				{#if deals.length === 0 && !loading}
					<div class="cr-page__empty">
						<p class="cr-page__empty-text">No deals in this pipeline yet.</p>
						<button
							onclick={() => handleAddDeal(stages[0]?.id)}
							class="cr-page__empty-link"
							aria-label="Add your first deal"
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
		class="cr-modal-overlay"
		onclick={() => (showAddDealModal = false)}
		role="dialog"
		aria-modal="true"
		transition:fade={{ duration: 150 }}
	>
		<div
			class="cr-modal"
			onclick={(e) => e.stopPropagation()}
			role="document"
			transition:scale={{ duration: 200, start: 0.95 }}
		>
			<h2 class="cr-modal__title">Create New Deal</h2>
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
				<div class="cr-modal__fields">
					<div class="cr-modal__field">
						<label for="deal-name" class="cr-modal__label">Deal Name</label>
						<input
							id="deal-name"
							name="name"
							type="text"
							required
							class="cr-modal__input"
							placeholder="e.g., Enterprise License Q1"
						/>
					</div>

					<div class="cr-modal__field">
						<label for="deal-amount" class="cr-modal__label">Amount</label>
						<input
							id="deal-amount"
							name="amount"
							type="number"
							class="cr-modal__input"
							placeholder="50000"
						/>
					</div>

					<div class="cr-modal__field">
						<label for="deal-close-date" class="cr-modal__label">Expected Close Date</label>
						<input
							id="deal-close-date"
							name="expected_close_date"
							type="date"
							class="cr-modal__input"
						/>
					</div>
				</div>

				<div class="cr-modal__actions">
					<button
						type="button"
						onclick={() => (showAddDealModal = false)}
						class="btn-rounded btn-rounded-ghost"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="btn-pill btn-pill-primary btn-pill-sm"
					>
						Create Deal
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	/* ── CRM Page Layout ── */
	.cr-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--dbg2);
	}

	/* ── Header ── */
	.cr-page__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 1.5rem;
		background: var(--dbg);
		border-bottom: 1px solid var(--dbd);
	}
	.cr-page__header-left {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.cr-page__header-right {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.cr-page__title {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0;
	}
	.cr-page__subtitle {
		font-size: 0.8125rem;
		color: var(--dt3);
		margin: 0.125rem 0 0;
	}

	/* ── Pipeline Selector ── */
	.cr-pipeline-select {
		padding: 0.375rem 0.75rem;
		border: 1px solid var(--dbd);
		border-radius: 8px;
		font-size: 0.875rem;
		color: var(--dt);
		background: var(--dbg);
		outline: none;
		cursor: pointer;
	}
	.cr-pipeline-select:focus {
		border-color: var(--dt3);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--dt3) 25%, transparent);
	}

	/* ── Stats Strip ── */
	.cr-stats-strip {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-right: 0.5rem;
		font-size: 0.875rem;
	}
	.cr-stat {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		color: var(--dt3);
	}
	.cr-stat__value {
		font-weight: 600;
		color: var(--dt);
	}
	.cr-stat__label {
		color: var(--dt3);
	}
	.cr-stat__value--open {
		color: #3b82f6;
	}
	.cr-stat__value--won {
		color: #22c55e;
	}

	/* ── View Switcher ── */
	.cr-view-switcher {
		display: flex;
		border: 1px solid var(--dbd);
		border-radius: 8px;
		overflow: hidden;
	}
	.cr-view-switcher__tab {
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		font-weight: 500;
		border: none;
		cursor: pointer;
		background: var(--dbg);
		color: var(--dt3);
		transition: background 0.15s, color 0.15s;
	}
	.cr-view-switcher__tab:hover {
		background: var(--dbg2);
		color: var(--dt);
	}
	.cr-view-switcher__tab--active {
		background: var(--dt);
		color: var(--dbg);
	}
	.cr-view-switcher__tab--active:hover {
		background: var(--dt);
		color: var(--dbg);
	}

	/* ── Error State ── */
	.cr-page__error {
		margin: 1rem 1.5rem 0;
		padding: 1rem;
		background: color-mix(in srgb, #ef4444 10%, var(--dbg));
		border: 1px solid color-mix(in srgb, #ef4444 30%, var(--dbd));
		border-radius: 8px;
	}
	.cr-page__error-text {
		font-size: 0.875rem;
		color: #ef4444;
		margin: 0;
	}
	.cr-page__error-retry {
		margin-top: 0.5rem;
		font-size: 0.875rem;
		color: #ef4444;
		text-decoration: underline;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}
	.cr-page__error-retry:hover {
		opacity: 0.8;
	}

	/* ── Center / Loading ── */
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
		gap: 0.75rem;
		color: var(--dt3);
	}
	.cr-page__center-text {
		font-size: 0.875rem;
		color: var(--dt3);
		margin: 0;
	}

	/* ── Kanban ── */
	.cr-kanban-container {
		flex: 1;
		overflow-x: auto;
		padding: 1.5rem;
	}
	.cr-kanban {
		display: flex;
		gap: 1rem;
		height: 100%;
		min-width: max-content;
	}
	.cr-kanban-col {
		width: 18rem;
		display: flex;
		flex-direction: column;
		background: var(--dbg3);
		border-radius: 8px;
		transition: box-shadow 0.2s, background 0.2s;
	}
	.cr-kanban-col--dragover {
		box-shadow: 0 0 0 2px #3b82f6;
		background: color-mix(in srgb, #3b82f6 8%, var(--dbg3));
	}
	.cr-kanban-col__header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem;
		border-bottom: 1px solid var(--dbd2);
	}
	.cr-kanban-col__dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.cr-kanban-col__stage {
		font-weight: 600;
		font-size: 0.875rem;
		color: var(--dt);
		flex: 1;
	}
	.cr-kanban-col__count {
		padding: 0.125rem 0.5rem;
		font-size: 0.75rem;
		font-weight: 600;
		background: var(--dbg2);
		color: var(--dt3);
		border-radius: 4px;
	}
	.cr-kanban-col__total {
		padding: 0 0.75rem 0.5rem;
		font-size: 0.75rem;
		color: var(--dt3);
	}
	.cr-kanban-col__cards {
		flex: 1;
		overflow-y: auto;
		padding: 0.5rem;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.cr-kanban-col__add-btn {
		width: 100%;
		padding: 0.5rem;
		font-size: 0.875rem;
		color: var(--dt3);
		background: none;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.25rem;
		transition: background 0.15s, color 0.15s;
	}
	.cr-kanban-col__add-btn:hover {
		background: var(--dbg2);
		color: var(--dt);
	}

	/* ── Kanban Card ── */
	.cr-kanban-card {
		background: var(--dbg);
		border: 1px solid var(--dbd2);
		border-radius: 8px;
		cursor: pointer;
		overflow: hidden;
		transition: box-shadow 0.15s;
	}
	.cr-kanban-card:hover {
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
	}
	.cr-kanban-card__bar {
		height: 3px;
		background: var(--stage-color, var(--dbd2));
	}
	.cr-kanban-card__body {
		padding: 0.75rem;
	}
	.cr-kanban-card__header-row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 0.5rem;
	}
	.cr-kanban-card__name {
		font-weight: 600;
		font-size: 0.875rem;
		color: var(--dt);
	}
	.cr-kanban-card__contact {
		font-size: 0.75rem;
		color: var(--dt3);
		margin-bottom: 0.5rem;
	}
	.cr-kanban-card__meta {
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-size: 0.875rem;
	}
	.cr-kanban-card__value {
		font-weight: 600;
		color: var(--dt);
	}
	.cr-kanban-card__prob {
		font-size: 0.75rem;
		color: var(--dt3);
	}
	.cr-kanban-card__date {
		margin-top: 0.5rem;
		font-size: 0.75rem;
		color: var(--dt4);
	}

	/* ── Priority Pill ── */
	.cr-priority-pill {
		padding: 0.125rem 0.5rem;
		font-size: 0.6875rem;
		font-weight: 600;
		border-radius: 9999px;
		text-transform: capitalize;
	}
	.cr-priority-pill--high {
		background: color-mix(in srgb, #ef4444 15%, var(--dbg));
		color: #ef4444;
	}
	.cr-priority-pill--medium {
		background: color-mix(in srgb, #f59e0b 15%, var(--dbg));
		color: #f59e0b;
	}
	.cr-priority-pill--low {
		background: color-mix(in srgb, #22c55e 15%, var(--dbg));
		color: #22c55e;
	}

	/* ── List View / Table ── */
	.cr-list-container {
		flex: 1;
		overflow: auto;
		padding: 1.5rem;
	}
	.cr-table-wrap {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 8px;
		overflow: hidden;
	}
	.cr-table {
		width: 100%;
		border-collapse: collapse;
	}
	.cr-table__head-row {
		background: var(--dbg2);
		border-bottom: 1px solid var(--dbd);
	}
	.cr-table__th {
		text-align: left;
		padding: 0.75rem 1rem;
		font-size: 0.6875rem;
		font-weight: 600;
		color: var(--dt3);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.cr-table__th--right {
		text-align: right;
	}
	.cr-table__th--center {
		text-align: center;
	}
	.cr-table__row {
		border-bottom: 1px solid var(--dbd2);
		cursor: pointer;
		transition: background 0.1s;
	}
	.cr-table__row:hover {
		background: var(--dbg2);
	}
	.cr-table__row:last-child {
		border-bottom: none;
	}
	.cr-table__td {
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		color: var(--dt);
	}
	.cr-table__td--name {
		font-weight: 600;
	}
	.cr-table__td--muted {
		color: var(--dt3);
	}
	.cr-table__td--right {
		text-align: right;
	}
	.cr-table__td--center {
		text-align: center;
	}

	/* ── Status Pill (table) ── */
	.cr-status-pill {
		display: inline-block;
		padding: 0.25rem 0.625rem;
		font-size: 0.6875rem;
		font-weight: 600;
		border-radius: 9999px;
		text-transform: capitalize;
	}
	.cr-status-pill--open {
		background: color-mix(in srgb, #3b82f6 15%, var(--dbg));
		color: #3b82f6;
	}
	.cr-status-pill--won {
		background: color-mix(in srgb, #22c55e 15%, var(--dbg));
		color: #22c55e;
	}
	.cr-status-pill--lost {
		background: color-mix(in srgb, #ef4444 15%, var(--dbg));
		color: #ef4444;
	}

	/* ── Empty State ── */
	.cr-page__empty {
		padding: 2rem;
		text-align: center;
	}
	.cr-page__empty-text {
		color: var(--dt3);
		margin: 0;
	}
	.cr-page__empty-link {
		margin-top: 0.5rem;
		font-size: 0.875rem;
		color: #3b82f6;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}
	.cr-page__empty-link:hover {
		opacity: 0.8;
	}

	/* ── Modal ── */
	.cr-modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 50;
	}
	.cr-modal {
		background: var(--dbg);
		border-radius: 12px;
		box-shadow: 0 16px 48px rgba(0, 0, 0, 0.15);
		width: 100%;
		max-width: 28rem;
		padding: 1.5rem;
	}
	.cr-modal__title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
		margin: 0 0 1rem;
	}
	.cr-modal__fields {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.cr-modal__field {
		display: flex;
		flex-direction: column;
	}
	.cr-modal__label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt2);
		margin-bottom: 0.25rem;
	}
	.cr-modal__input {
		width: 100%;
		padding: 0.5rem 0.75rem;
		border: 1px solid var(--dbd);
		border-radius: 8px;
		font-size: 0.875rem;
		color: var(--dt);
		background: var(--dbg);
		outline: none;
		transition: border-color 0.15s;
		box-sizing: border-box;
	}
	.cr-modal__input:focus {
		border-color: var(--dt3);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--dt3) 25%, transparent);
	}
	.cr-modal__input::placeholder {
		color: var(--dt4);
	}
	.cr-modal__actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 1.5rem;
	}

	/* ── Dark Mode tweaks ── */
	:global(.dark) .cr-kanban-card:hover {
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
	}
	:global(.dark) .cr-modal {
		box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
	}
</style>
