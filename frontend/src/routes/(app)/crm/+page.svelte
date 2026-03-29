<script lang="ts">
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import {
		crm,
		type CRMViewMode,
		formatCurrency,
		formatProbability
	} from '$lib/stores/crm';
	import type { Pipeline, Deal, CreateDealData } from '$lib/api/crm';
	import { Plus, LayoutGrid, List, Building2, Calendar, TrendingUp, DollarSign, Target, X } from 'lucide-svelte';

	// Embed mode check
	const embedSuffix = $derived(
		$page.url.searchParams.get('embed') === 'true' ? '?embed=true' : ''
	);

	// Reactive store slices
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
	let isSubmitting = $state(false);
	let modalError = $state<string | null>(null);

	// Drag state
	let draggedDealId = $state<string | null>(null);
	let dragOverStageId = $state<string | null>(null);

	// Load on mount — non-blocking with fast timeout so seed data shows quickly
	onMount(() => {
		loadCRMData();
	});

	async function loadCRMData() {
		// loadPipelines handles its own 3s timeout and seed data fallback internally
		try {
			await crm.loadPipelines();
		} catch {
			// Seed data already loaded by loadPipelines catch handler; nothing to do here
		}

		const pipelineId = get(crm).currentPipeline?.id;
		if (pipelineId) {
			// Fire deals + stats in parallel, don't block
			crm.loadDeals({ pipeline_id: pipelineId }).catch(() => {});
			crm.loadDealStats(pipelineId).catch(() => {});
		}
	}

	// Group deals by stage — expression form, not arrow function
	const dealsByStage = $derived(
		Object.fromEntries(
			stages.map((stage) => [stage.id, deals.filter((d) => d.stage_id === stage.id)])
		) as Record<string, Deal[]>
	);

	// Computed stats from dealStats
	const avgDealSize = $derived(
		dealStats && dealStats.total_deals > 0
			? Math.round(
					((dealStats.open_value || 0) + (dealStats.won_value || 0) + (dealStats.lost_value || 0)) /
						dealStats.total_deals
				)
			: 0
	);

	const winRate = $derived(
		dealStats && dealStats.total_deals > 0
			? Math.round((dealStats.won_deals / dealStats.total_deals) * 100)
			: 0
	);

	// Helpers
	function getStageTotal(stageId: string): number {
		return (dealsByStage[stageId] || []).reduce((sum, d) => sum + (d.amount || 0), 0);
	}

	function formatDate(dateStr: string | undefined | null): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function getStageName(stageId: string): string {
		return stages.find((s) => s.id === stageId)?.name || '-';
	}

	function getStageColor(stageId: string): string {
		return stages.find((s) => s.id === stageId)?.color || 'var(--dbd2)';
	}

	// Handlers
	function handlePipelineChange(pipeline: Pipeline) {
		crm.selectPipeline(pipeline);
	}

	function handleViewChange(mode: CRMViewMode) {
		console.log('[CRM] View change clicked:', mode);
		crm.setViewMode(mode);
	}

	function handleDealClick(dealId: string) {
		goto(`/crm/deals/${dealId}${embedSuffix}`);
	}

	function handleAddDeal(stageId?: string) {
		console.log('[CRM] Add Deal clicked, stageId:', stageId);
		selectedStageId = stageId || stages[0]?.id || null;
		showAddDealModal = true;
		console.log('[CRM] showAddDealModal set to:', showAddDealModal);
	}

	function closeModal() {
		showAddDealModal = false;
		selectedStageId = null;
		isSubmitting = false;
		modalError = null;
	}

	async function handleCreateDeal(e: SubmitEvent) {
		e.preventDefault();
		if (isSubmitting) return;
		modalError = null;

		const formData = new FormData(e.currentTarget as HTMLFormElement);
		const pipelineId = currentPipeline?.id || pipelines[0]?.id;
		const stageId = selectedStageId || stages[0]?.id;

		if (!pipelineId || !stageId) {
			modalError = 'No pipeline or stage available. Please reload the page.';
			return;
		}

		const name = (formData.get('name') as string)?.trim();
		if (!name) {
			modalError = 'Deal name is required.';
			return;
		}

		isSubmitting = true;
		try {
			await crm.createDeal({
				pipeline_id: pipelineId,
				stage_id: stageId,
				name,
				amount: formData.get('amount') ? Number(formData.get('amount')) : undefined,
				expected_close_date: (formData.get('expected_close_date') as string) || undefined
			} as CreateDealData);
			// Refresh deals and stats after successful creation
			crm.loadDeals({ pipeline_id: pipelineId }).catch(() => {});
			crm.loadDealStats(pipelineId).catch(() => {});
			closeModal();
		} catch (err) {
			console.error('[CRM] createDeal failed:', err);
			modalError = err instanceof Error ? err.message : 'Failed to create deal. Please try again.';
			isSubmitting = false;
		}
	}

	// Drag and drop
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

	async function handleDrop(e: DragEvent, stageId: string) {
		e.preventDefault();
		if (draggedDealId && draggedDealId !== stageId) {
			try {
				await crm.moveDealToStage(draggedDealId, stageId);
			} catch {
				// handled internally in store
			}
		}
		draggedDealId = null;
		dragOverStageId = null;
	}

	// Priority dot color via CSS variable name
	function priorityVar(priority: string): string {
		const map: Record<string, string> = {
			low: 'var(--bos-priority-low)',
			medium: 'var(--bos-priority-medium)',
			high: 'var(--bos-priority-high)',
			urgent: 'var(--bos-priority-critical)',
			critical: 'var(--bos-priority-critical)'
		};
		return map[priority] || 'var(--dt4)';
	}
</script>

<div class="cr-page">
	<!-- ── Header ── -->
	<header class="cr-header">
		<div class="cr-header__left">
			<div class="cr-header__title-block">
				<h1 class="cr-header__title">Sales Pipeline</h1>
				<p class="cr-header__subtitle">Manage deals and track your sales process</p>
			</div>

			{#if pipelines.length > 1}
				<select
					class="cr-pipeline-select"
					value={currentPipeline?.id || ''}
					onchange={(e) => {
						const p = pipelines.find((pl) => pl.id === e.currentTarget.value);
						if (p) handlePipelineChange(p);
					}}
					aria-label="Select pipeline"
				>
					{#each pipelines as pl}
						<option value={pl.id}>{pl.name}</option>
					{/each}
				</select>
			{/if}
		</div>

		<div class="cr-header__right">
			<!-- Stats strip -->
			{#if dealStats}
				<div class="cr-stats-strip">
					<div class="cr-stat">
						<span class="cr-stat__val">{dealStats.total_deals}</span>
						<span class="cr-stat__lbl">deals</span>
					</div>
					<div class="cr-stat__sep" aria-hidden="true"></div>
					<div class="cr-stat">
						<span class="cr-stat__lbl">Open</span>
						<span class="cr-stat__val cr-stat__val--info">{formatCurrency(dealStats.open_value)}</span>
					</div>
					<div class="cr-stat__sep" aria-hidden="true"></div>
					<div class="cr-stat">
						<span class="cr-stat__lbl">Won</span>
						<span class="cr-stat__val cr-stat__val--success">{formatCurrency(dealStats.won_value)}</span>
					</div>
				</div>
			{/if}

			<!-- View switcher -->
			<div class="cr-view-group" role="group" aria-label="View mode">
				<button
					class="cr-view-btn {viewMode === 'kanban' ? 'cr-view-btn--active' : ''}"
					onclick={() => handleViewChange('kanban')}
					aria-pressed={viewMode === 'kanban'}
					aria-label="Kanban view"
				>
					<LayoutGrid size={14} aria-hidden="true" />
					Kanban
				</button>
				<button
					class="cr-view-btn {viewMode === 'list' ? 'cr-view-btn--active' : ''}"
					onclick={() => handleViewChange('list')}
					aria-pressed={viewMode === 'list'}
					aria-label="List view"
				>
					<List size={14} aria-hidden="true" />
					List
				</button>
			</div>

			<button
				class="btn-cta"
				onclick={(e) => {
					e.stopPropagation();
					console.log('[CRM] btn-cta clicked');
					showAddDealModal = true;
					selectedStageId = stages[0]?.id || null;
				}}
				aria-label="Add new deal"
				style="position: relative; z-index: 10; pointer-events: auto;"
			>
				<Plus size={15} aria-hidden="true" />
				Add Deal
			</button>
		</div>
	</header>

	<!-- ── Stats Bar ── -->
	{#if dealStats}
		<div class="cr-statsbar">
			<div class="cr-statsbar__metric">
				<DollarSign size={13} class="cr-statsbar__icon" aria-hidden="true" />
				<span class="cr-statsbar__num">{formatCurrency(dealStats.open_value)}</span>
				<span class="cr-statsbar__label">Open Value</span>
			</div>
			<div class="cr-statsbar__divider" aria-hidden="true"></div>
			<div class="cr-statsbar__metric">
				<TrendingUp size={13} class="cr-statsbar__icon" aria-hidden="true" />
				<span class="cr-statsbar__num cr-statsbar__num--success">{formatCurrency(dealStats.won_value)}</span>
				<span class="cr-statsbar__label">Won Value</span>
			</div>
			<div class="cr-statsbar__divider" aria-hidden="true"></div>
			<div class="cr-statsbar__metric">
				<Target size={13} class="cr-statsbar__icon" aria-hidden="true" />
				<span class="cr-statsbar__num">{formatCurrency(avgDealSize)}</span>
				<span class="cr-statsbar__label">Avg Deal Size</span>
			</div>
			<div class="cr-statsbar__divider" aria-hidden="true"></div>
			<div class="cr-statsbar__metric">
				<Building2 size={13} class="cr-statsbar__icon" aria-hidden="true" />
				<span class="cr-statsbar__num">{dealStats.open_deals}</span>
				<span class="cr-statsbar__label">Open Deals</span>
			</div>
			<div class="cr-statsbar__divider" aria-hidden="true"></div>
			<div class="cr-statsbar__metric">
				<Calendar size={13} class="cr-statsbar__icon" aria-hidden="true" />
				<span class="cr-statsbar__num cr-statsbar__num--accent">{winRate}%</span>
				<span class="cr-statsbar__label">Win Rate</span>
			</div>
		</div>
	{/if}

	<!-- ── Error Banner ── -->
	{#if error}
		<div class="cr-error-banner" role="alert">
			<span class="cr-error-banner__text">{error}</span>
			<button
				class="cr-error-banner__retry"
				onclick={() => crm.loadPipelines()}
				aria-label="Retry loading pipeline"
			>
				Try again
			</button>
		</div>
	{/if}

	<!-- ── Loading indicator (non-blocking) ── -->
	{#if loading && stages.length === 0 && deals.length === 0}
		<div class="cr-loading-bar" role="status" aria-label="Loading pipeline">
			<div class="cr-loading-bar__track"></div>
		</div>
	{/if}

	<!-- ── Kanban View ── -->
	{#if viewMode === 'kanban'}
		<div class="cr-kanban-scroll">
			<div class="cr-kanban">
				{#each stages as stage}
					{@const stageDeals = dealsByStage[stage.id] || []}
					{@const stageTotal = getStageTotal(stage.id)}
					<div
						class="cr-col {dragOverStageId === stage.id ? 'cr-col--dragover' : ''}"
						style="--stage-color: {stage.color || 'var(--dbd2)'}"
						ondragover={(e) => handleDragOver(e, stage.id)}
						ondragleave={handleDragLeave}
						ondrop={(e) => handleDrop(e, stage.id)}
						role="region"
						aria-label="Stage: {stage.name}"
					>
						<!-- Column header -->
						<div class="cr-col__header">
							<span class="cr-col__dot" aria-hidden="true"></span>
							<span class="cr-col__name">{stage.name}</span>
							<span class="cr-col__count" aria-label="{stageDeals.length} deals">{stageDeals.length}</span>
						</div>
						<div class="cr-col__total">{formatCurrency(stageTotal)}</div>

						<!-- Cards -->
						<div class="cr-col__cards">
							{#each stageDeals as deal (deal.id)}
								<div
									class="cr-card"
									style="--stage-color: {stage.color || 'var(--dbd2)'}"
									draggable="true"
									ondragstart={(e) => handleDragStart(e, deal.id)}
									onclick={() => handleDealClick(deal.id)}
									onkeydown={(e) => e.key === 'Enter' && handleDealClick(deal.id)}
									role="button"
									tabindex="0"
									aria-label="Open deal: {deal.name}"
								>
									<div class="cr-card__bar" aria-hidden="true"></div>
									<div class="cr-card__body">
										<div class="cr-card__top">
											<span class="cr-card__name">{deal.name}</span>
											{#if deal.priority}
												<span
													class="cr-priority-dot"
													style="background: {priorityVar(deal.priority)}"
													title="Priority: {deal.priority}"
													aria-label="Priority: {deal.priority}"
												></span>
											{/if}
										</div>

										{#if deal.company_name}
											<div class="cr-card__company">{deal.company_name}</div>
										{/if}

										<div class="cr-card__meta">
											<span class="cr-card__amount">{formatCurrency(deal.amount, deal.currency)}</span>
											{#if deal.probability !== undefined}
												<span class="cr-card__prob">{formatProbability(deal.probability)}</span>
											{/if}
										</div>

										{#if deal.expected_close_date}
											<div class="cr-card__date">
												Close {formatDate(deal.expected_close_date)}
											</div>
										{/if}
									</div>
								</div>
							{/each}

							<!-- Empty column state -->
							{#if stageDeals.length === 0}
								<div class="cr-col__empty" aria-label="No deals in {stage.name}">
									No deals
								</div>
							{/if}

							<!-- Add deal to stage -->
							<button
								class="cr-col__add"
								onclick={() => handleAddDeal(stage.id)}
								aria-label="Add deal to {stage.name}"
							>
								<Plus size={13} aria-hidden="true" />
								Add Deal
							</button>
						</div>
					</div>
				{/each}

				<!-- Empty pipeline state inside kanban -->
				{#if stages.length === 0 && !loading}
					<div class="cr-empty">
						<div class="cr-empty__icon" aria-hidden="true">
							<TrendingUp size={32} />
						</div>
						<p class="cr-empty__title">No pipeline configured</p>
						<p class="cr-empty__subtitle">Add stages to get started</p>
					</div>
				{/if}
			</div>
		</div>

	<!-- ── List View ── -->
	{:else}
		<div class="cr-list-scroll">
			{#if deals.length === 0 && !loading}
				<div class="cr-empty cr-empty--list">
					<div class="cr-empty__icon" aria-hidden="true">
						<List size={32} />
					</div>
					<p class="cr-empty__title">No deals yet</p>
					<p class="cr-empty__subtitle">Add your first deal to get started</p>
					<button
						class="btn-pill btn-pill-primary btn-pill-sm"
						onclick={() => handleAddDeal(stages[0]?.id)}
						aria-label="Add your first deal"
					>
						<Plus size={13} aria-hidden="true" />
						Add Deal
					</button>
				</div>
			{:else}
				<table class="cr-table">
					<thead>
						<tr class="cr-table__head">
							<th class="cr-th" scope="col">Deal</th>
							<th class="cr-th" scope="col">Company</th>
							<th class="cr-th" scope="col">Stage</th>
							<th class="cr-th cr-th--right" scope="col">Amount</th>
							<th class="cr-th cr-th--center" scope="col">Probability</th>
							<th class="cr-th" scope="col">Close Date</th>
							<th class="cr-th" scope="col">Status</th>
						</tr>
					</thead>
					<tbody>
						{#each deals as deal (deal.id)}
							<tr
								class="cr-row"
								onclick={() => handleDealClick(deal.id)}
								onkeydown={(e) => e.key === 'Enter' && handleDealClick(deal.id)}
								role="button"
								tabindex="0"
								aria-label="Open deal: {deal.name}"
							>
								<td class="cr-td">
									<div class="cr-td__name-cell">
										<span
											class="cr-td__stage-bar"
											style="background: {getStageColor(deal.stage_id)}"
											aria-hidden="true"
										></span>
										<span class="cr-td__name">{deal.name}</span>
									</div>
								</td>
								<td class="cr-td cr-td--muted">{deal.company_name || '-'}</td>
								<td class="cr-td">
									{#if deal.stage_id}
										<span class="cr-stage-pill">{getStageName(deal.stage_id)}</span>
									{:else}
										<span class="cr-td--muted">-</span>
									{/if}
								</td>
								<td class="cr-td cr-td--right cr-td--num cr-td--bold">
									{formatCurrency(deal.amount, deal.currency)}
								</td>
								<td class="cr-td cr-td--center cr-td--muted">
									{formatProbability(deal.probability)}
								</td>
								<td class="cr-td cr-td--muted">
									{formatDate(deal.expected_close_date)}
								</td>
								<td class="cr-td">
									{#if deal.status}
										<span class="cr-status-dot cr-status-dot--{deal.status}" aria-hidden="true"></span>
										<span class="cr-status-label cr-status-label--{deal.status}">
											{deal.status === 'open' ? 'Open' : deal.status === 'won' ? 'Won' : 'Lost'}
										</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</div>
	{/if}
</div>

<!-- ── Add Deal Modal ── -->
{#if showAddDealModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="cr-modal-backdrop"
		onclick={closeModal}
		onkeydown={(e) => e.key === 'Escape' && closeModal()}
	>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="cr-modal-box"
			onclick={(e) => e.stopPropagation()}
			role="dialog"
			aria-modal="true"
			aria-label="Create new deal"
		>
			<!-- Header -->
			<div class="cr-modal-box__header">
				<h2 class="cr-modal-box__title">Create New Deal</h2>
				<button
					class="btn-pill btn-pill-ghost btn-pill-icon"
					onclick={closeModal}
					aria-label="Close modal"
				>
					<X size={16} aria-hidden="true" />
				</button>
			</div>

			<!-- Body -->
			<form onsubmit={handleCreateDeal}>
				<div class="cr-modal-box__body">
					{#if modalError}
						<div class="cr-modal-error" role="alert">{modalError}</div>
					{/if}

					{#if stages.length > 0}
						<div class="cr-field">
							<label class="cr-field__label" for="modal-stage">Stage</label>
							<select
								id="modal-stage"
								class="cr-field__input"
								value={selectedStageId || stages[0]?.id || ''}
								onchange={(e) => (selectedStageId = e.currentTarget.value)}
							>
								{#each stages as stage}
									<option value={stage.id}>{stage.name}</option>
								{/each}
							</select>
						</div>
					{/if}

					<div class="cr-field">
						<label class="cr-field__label cr-field__label--req" for="modal-name">Deal Name</label>
						<input
							id="modal-name"
							name="name"
							type="text"
							class="cr-field__input"
							placeholder="e.g., Enterprise License Q2"
							required
							autocomplete="off"
						/>
					</div>

					<div class="cr-field">
						<label class="cr-field__label" for="modal-amount">Amount ($)</label>
						<input
							id="modal-amount"
							name="amount"
							type="number"
							class="cr-field__input"
							placeholder="50000"
							min="0"
							step="1"
						/>
					</div>

					<div class="cr-field">
						<label class="cr-field__label" for="modal-date">Expected Close Date</label>
						<input
							id="modal-date"
							name="expected_close_date"
							type="date"
							class="cr-field__input"
						/>
					</div>
				</div>

				<!-- Footer -->
				<div class="cr-modal-box__footer">
					<button
						type="button"
						class="btn-pill btn-pill-ghost"
						onclick={closeModal}
						disabled={isSubmitting}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="btn-cta"
						disabled={isSubmitting}
						aria-busy={isSubmitting}
					>
						{isSubmitting ? 'Creating...' : 'Create Deal'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	/* ─────────────────────────────────────────
	   PAGE SHELL
	───────────────────────────────────────── */
	.cr-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--dbg2);
		font-family: var(--bos-font-family);
		overflow: hidden;
	}

	/* ─────────────────────────────────────────
	   HEADER
	───────────────────────────────────────── */
	.cr-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.875rem 1.5rem;
		background: var(--dbg);
		border-bottom: 1px solid var(--dbd);
		flex-shrink: 0;
	}
	.cr-header__left {
		display: flex;
		align-items: center;
		gap: 1rem;
		min-width: 0;
	}
	.cr-header__title-block {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}
	.cr-header__title {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--dt);
		letter-spacing: -0.01em;
		line-height: 1.3;
	}
	.cr-header__subtitle {
		margin: 0;
		font-size: 0.75rem;
		color: var(--dt3);
		line-height: 1.4;
	}
	.cr-header__right {
		display: flex;
		align-items: center;
		gap: 0.625rem;
		flex-shrink: 0;
	}

	/* ─────────────────────────────────────────
	   PIPELINE SELECT
	───────────────────────────────────────── */
	.cr-pipeline-select {
		height: 30px;
		padding: 0 0.625rem;
		border: 1px solid var(--dbd);
		border-radius: 8px;
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--dt);
		background: var(--dbg);
		outline: none;
		cursor: pointer;
		transition: border-color 0.15s, box-shadow 0.15s;
	}
	.cr-pipeline-select:focus {
		border-color: var(--bos-nav-active);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--bos-nav-active) 18%, transparent);
	}

	/* ─────────────────────────────────────────
	   INLINE STATS STRIP (header)
	───────────────────────────────────────── */
	.cr-stats-strip {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.8125rem;
		padding: 0 0.5rem;
	}
	.cr-stat {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}
	.cr-stat__lbl {
		color: var(--dt3);
		font-size: 0.75rem;
	}
	.cr-stat__val {
		font-weight: 700;
		color: var(--dt);
		font-family: var(--bos-font-number-family);
		font-size: 0.8125rem;
	}
	.cr-stat__val--info  { color: var(--bos-status-info); }
	.cr-stat__val--success { color: var(--bos-status-success); }
	.cr-stat__sep {
		width: 1px;
		height: 12px;
		background: var(--dbd);
	}

	/* View switcher */
	.cr-view-group {
		display: flex;
		border: 1px solid var(--dbd);
		border-radius: 8px;
		overflow: hidden;
	}
	.cr-view-btn {
		display: flex;
		align-items: center;
		gap: 0.3rem;
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
		font-weight: 500;
		font-family: var(--bos-font-family);
		border: none;
		cursor: pointer;
		background: var(--dbg);
		color: var(--dt3);
		transition: background 0.15s, color 0.15s;
	}
	.cr-view-btn:hover {
		background: var(--dbg2);
		color: var(--dt);
	}
	.cr-view-btn--active {
		background: var(--dt);
		color: var(--dbg);
	}
	.cr-view-btn--active:hover {
		background: var(--dt);
		color: var(--dbg);
	}

	/* ─────────────────────────────────────────
	   STATS BAR (secondary row)
	───────────────────────────────────────── */
	.cr-statsbar {
		display: flex;
		align-items: center;
		gap: 0;
		padding: 0 1.5rem;
		height: 38px;
		background: var(--dbg);
		border-bottom: 1px solid var(--dbd2);
		flex-shrink: 0;
		overflow-x: auto;
	}
	.cr-statsbar__metric {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0 1rem;
		white-space: nowrap;
		flex-shrink: 0;
	}
	.cr-statsbar__metric:first-child {
		padding-left: 0;
	}
	:global(.cr-statsbar__icon) {
		color: var(--dt4);
		flex-shrink: 0;
	}
	.cr-statsbar__num {
		font-family: var(--bos-font-number-family);
		font-size: 0.8125rem;
		font-weight: 700;
		color: var(--dt);
		letter-spacing: -0.02em;
	}
	.cr-statsbar__num--success { color: var(--bos-status-success); }
	.cr-statsbar__num--accent  { color: var(--bos-status-accent); }
	.cr-statsbar__label {
		font-size: 0.6875rem;
		color: var(--dt3);
	}
	.cr-statsbar__divider {
		width: 1px;
		height: 18px;
		background: var(--dbd2);
		flex-shrink: 0;
	}

	/* ─────────────────────────────────────────
	   ERROR BANNER
	───────────────────────────────────────── */
	.cr-error-banner {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin: 0.75rem 1.5rem 0;
		padding: 0.625rem 0.875rem;
		background: color-mix(in srgb, var(--bos-status-error) 8%, var(--dbg));
		border: 1px solid color-mix(in srgb, var(--bos-status-error) 25%, var(--dbd));
		border-radius: 8px;
		flex-shrink: 0;
	}
	.cr-error-banner__text {
		flex: 1;
		font-size: 0.8125rem;
		color: var(--bos-status-error);
	}
	.cr-error-banner__retry {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--bos-status-error);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
		text-decoration: underline;
		text-underline-offset: 2px;
		flex-shrink: 0;
	}
	.cr-error-banner__retry:hover { opacity: 0.75; }

	/* ─────────────────────────────────────────
	   LOADING / CENTER
	───────────────────────────────────────── */
	.cr-loading-bar {
		height: 2px;
		width: 100%;
		overflow: hidden;
		flex-shrink: 0;
	}
	.cr-loading-bar__track {
		height: 100%;
		width: 30%;
		background: var(--bos-nav-active);
		border-radius: 1px;
		animation: cr-slide 1s ease-in-out infinite;
	}
	@keyframes cr-slide {
		0% { transform: translateX(-100%); }
		100% { transform: translateX(400%); }
	}

	/* ─────────────────────────────────────────
	   KANBAN SCROLL + BOARD
	───────────────────────────────────────── */
	.cr-kanban-scroll {
		flex: 1;
		overflow-x: auto;
		overflow-y: hidden;
		padding: 1.25rem 1.5rem;
	}
	.cr-kanban {
		display: flex;
		gap: 0.875rem;
		height: 100%;
		min-width: max-content;
		align-items: flex-start;
	}

	/* ─────────────────────────────────────────
	   KANBAN COLUMN
	───────────────────────────────────────── */
	.cr-col {
		width: 17rem;
		flex-shrink: 0;
		display: flex;
		flex-direction: column;
		background: var(--dbg3);
		border-radius: 10px;
		max-height: calc(100vh - 200px);
		transition: box-shadow 0.18s;
	}
	.cr-col--dragover {
		box-shadow: 0 0 0 2px var(--bos-status-info);
	}
	.cr-col__header {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.625rem 0.75rem 0.5rem;
	}
	.cr-col__dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--stage-color, var(--dbd2));
		flex-shrink: 0;
	}
	.cr-col__name {
		flex: 1;
		font-size: 0.8125rem;
		font-weight: 700;
		color: var(--dt);
		letter-spacing: -0.005em;
	}
	.cr-col__count {
		font-size: 0.6875rem;
		font-weight: 600;
		color: var(--dt3);
		background: var(--dbg2);
		border-radius: 9999px;
		padding: 0.0625rem 0.4375rem;
		min-width: 18px;
		text-align: center;
	}
	.cr-col__total {
		padding: 0 0.75rem 0.5rem;
		font-size: 0.75rem;
		font-family: var(--bos-font-number-family);
		font-weight: 600;
		color: var(--dt3);
		border-bottom: 1px solid var(--dbd2);
	}
	.cr-col__cards {
		flex: 1;
		overflow-y: auto;
		padding: 0.5rem;
		display: flex;
		flex-direction: column;
		gap: 0.4375rem;
	}
	.cr-col__empty {
		text-align: center;
		font-size: 0.75rem;
		color: var(--dt4);
		padding: 0.75rem 0;
	}
	.cr-col__add {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.25rem;
		width: 100%;
		padding: 0.4375rem 0;
		font-size: 0.75rem;
		font-weight: 500;
		color: var(--dt4);
		background: none;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		transition: background 0.12s, color 0.12s;
		margin-top: 0.125rem;
	}
	.cr-col__add:hover {
		background: var(--dbg2);
		color: var(--dt2);
	}

	/* ─────────────────────────────────────────
	   KANBAN CARD
	───────────────────────────────────────── */
	.cr-card {
		background: var(--dbg);
		border: 1px solid var(--dbd2);
		border-radius: 8px;
		cursor: pointer;
		overflow: hidden;
		transition: box-shadow 0.14s, border-color 0.14s;
		outline: none;
	}
	.cr-card:hover {
		box-shadow: var(--bos-shadow-1);
		border-color: var(--dbd);
	}
	.cr-card:focus-visible {
		box-shadow: 0 0 0 2px var(--bos-nav-active);
		border-color: var(--bos-nav-active);
	}
	.cr-card__bar {
		height: 3px;
		background: var(--stage-color, var(--dbd2));
	}
	.cr-card__body {
		padding: 0.625rem 0.75rem;
	}
	.cr-card__top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 0.375rem;
		margin-bottom: 0.3125rem;
	}
	.cr-card__name {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--dt);
		line-height: 1.35;
		flex: 1;
	}
	.cr-priority-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		flex-shrink: 0;
		margin-top: 0.3rem;
	}
	.cr-card__company {
		font-size: 0.6875rem;
		color: var(--dt3);
		margin-bottom: 0.4375rem;
	}
	.cr-card__meta {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.cr-card__amount {
		font-size: 0.8125rem;
		font-weight: 700;
		font-family: var(--bos-font-number-family);
		color: var(--dt);
		letter-spacing: -0.02em;
	}
	.cr-card__prob {
		font-size: 0.6875rem;
		color: var(--dt3);
	}
	.cr-card__date {
		margin-top: 0.375rem;
		font-size: 0.6875rem;
		color: var(--dt4);
	}

	/* ─────────────────────────────────────────
	   LIST VIEW
	───────────────────────────────────────── */
	.cr-list-scroll {
		flex: 1;
		overflow: auto;
		padding: 1.25rem 1.5rem;
	}
	.cr-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.8125rem;
		background: var(--dbg);
		border-radius: 10px;
		overflow: hidden;
		box-shadow: var(--bos-shadow-1);
	}
	.cr-table__head {
		background: var(--dbg2);
	}
	.cr-th {
		padding: 0.625rem 0.875rem;
		text-align: left;
		font-size: 0.6875rem;
		font-weight: 700;
		color: var(--dt3);
		text-transform: uppercase;
		letter-spacing: 0.055em;
		border-bottom: 1px solid var(--dbd);
	}
	.cr-th--right  { text-align: right; }
	.cr-th--center { text-align: center; }

	.cr-row {
		border-bottom: 1px solid var(--dbd2);
		cursor: pointer;
		transition: background 0.1s;
		outline: none;
	}
	.cr-row:last-child { border-bottom: none; }
	.cr-row:hover { background: var(--dbg2); }
	.cr-row:focus-visible { background: color-mix(in srgb, var(--bos-nav-active) 6%, var(--dbg)); }

	.cr-td {
		padding: 0.6875rem 0.875rem;
		color: var(--dt);
		vertical-align: middle;
	}
	.cr-td--muted {
		color: var(--dt3);
	}
	.cr-td--bold {
		font-weight: 700;
	}
	.cr-td--right  { text-align: right; }
	.cr-td--center { text-align: center; }
	.cr-td--num {
		font-family: var(--bos-font-number-family);
		letter-spacing: -0.02em;
	}

	/* Deal name cell with stage color bar */
	.cr-td__name-cell {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.cr-td__stage-bar {
		width: 3px;
		height: 1.25rem;
		border-radius: 2px;
		flex-shrink: 0;
	}
	.cr-td__name {
		font-weight: 600;
		color: var(--dt);
	}

	/* Stage pill in list */
	.cr-stage-pill {
		display: inline-block;
		padding: 0.1875rem 0.5rem;
		font-size: 0.6875rem;
		font-weight: 600;
		color: var(--dt2);
		background: var(--dbg3);
		border-radius: 4px;
		letter-spacing: 0.01em;
	}

	/* Status indicator in list — last cell uses flex for dot + label */
	.cr-td:last-child {
		display: flex;
		align-items: center;
		gap: 0.375rem;
	}
	.cr-status-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.cr-status-dot--open { background: var(--bos-status-info); }
	.cr-status-dot--won  { background: var(--bos-status-success); }
	.cr-status-dot--lost { background: var(--bos-status-error); }

	.cr-status-label {
		font-size: 0.75rem;
		font-weight: 600;
	}
	.cr-status-label--open { color: var(--bos-status-info); }
	.cr-status-label--won  { color: var(--bos-status-success); }
	.cr-status-label--lost { color: var(--bos-status-error); }

	/* ─────────────────────────────────────────
	   EMPTY STATE
	───────────────────────────────────────── */
	.cr-empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 3rem 1.5rem;
		min-height: 240px;
		color: var(--dt3);
	}
	.cr-empty--list {
		background: var(--dbg);
		border-radius: 10px;
	}
	.cr-empty__icon {
		color: var(--dbd);
		margin-bottom: 0.25rem;
	}
	.cr-empty__title {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--dt2);
		margin: 0;
	}
	.cr-empty__subtitle {
		font-size: 0.75rem;
		color: var(--dt3);
		margin: 0;
	}
	.cr-empty .btn-pill {
		margin-top: 0.75rem;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	/* ─────────────────────────────────────────
	   MODAL — Fresh implementation
	───────────────────────────────────────── */
	.cr-modal-backdrop {
		position: fixed;
		inset: 0;
		z-index: 1000;
		background: var(--bos-modal-backdrop, rgba(0, 0, 0, 0.5));
		display: flex;
		align-items: center;
		justify-content: center;
		animation: cr-fade-in 0.15s ease-out;
	}
	@keyframes cr-fade-in {
		from { opacity: 0; }
		to { opacity: 1; }
	}
	.cr-modal-box {
		width: 100%;
		max-width: 480px;
		background: var(--bos-modal-bg, var(--dbg));
		border: 1px solid var(--bos-modal-border, var(--dbd));
		border-radius: var(--bos-modal-radius, 12px);
		box-shadow: var(--bos-modal-shadow, 0 20px 25px -5px rgba(0, 0, 0, 0.1));
		display: flex;
		flex-direction: column;
		max-height: 90vh;
		overflow: hidden;
		animation: cr-scale-in 0.15s ease-out;
	}
	@keyframes cr-scale-in {
		from { opacity: 0; transform: scale(0.96); }
		to { opacity: 1; transform: scale(1); }
	}
	.cr-modal-box__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 1.25rem;
		border-bottom: 1px solid var(--dbd);
	}
	.cr-modal-box__title {
		margin: 0;
		font-size: 1rem;
		font-weight: 700;
		color: var(--dt);
		letter-spacing: -0.01em;
	}
	.cr-modal-box__body {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		overflow-y: auto;
	}
	.cr-modal-box__footer {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		padding: 0.875rem 1.25rem;
		border-top: 1px solid var(--dbd);
	}

	.cr-field {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}
	.cr-field__label {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--dt2);
	}
	.cr-field__label--req::after {
		content: ' *';
		color: var(--bos-status-error-text, #dc2626);
	}
	.cr-field__input {
		width: 100%;
		height: 36px;
		padding: 0 0.75rem;
		font-size: 0.8125rem;
		font-family: inherit;
		color: var(--dt);
		background: var(--dbg2);
		border: 1px solid var(--dbd);
		border-radius: 8px;
		outline: none;
		transition: border-color 0.15s, box-shadow 0.15s;
	}
	.cr-field__input:focus {
		border-color: var(--bos-nav-active);
		box-shadow: 0 0 0 2px color-mix(in srgb, var(--bos-nav-active) 18%, transparent);
	}
	.cr-field__input::placeholder {
		color: var(--dt4);
	}
	select.cr-field__input {
		cursor: pointer;
	}

	.cr-modal-error {
		padding: 0.625rem 0.75rem;
		font-size: 0.8125rem;
		color: var(--bos-status-error-text);
		background: var(--bos-status-error-bg);
		border: 1px solid color-mix(in srgb, var(--bos-status-error) 20%, transparent);
		border-radius: 8px;
	}
</style>
