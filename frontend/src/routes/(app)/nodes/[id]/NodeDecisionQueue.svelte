<script lang="ts">
	import { slide } from 'svelte/transition';
	import type { DecisionItem } from '$lib/api/nodes/types';

	interface Props {
		decisions: DecisionItem[];
		isSaving: boolean;
		onAdd: (question: string) => void;
		onDecide: (decisionId: string, answer: string) => void;
		onDelete: (decisionId: string) => void;
	}

	let { decisions, isSaving, onAdd, onDecide, onDelete }: Props = $props();

	let showAddDecision = $state(false);
	let newDecisionQuestion = $state('');
	let decidingDecisionId: string | null = $state(null);
	let decisionAnswer = $state('');
	let showDecidedHistory = $state(false);

	const pendingDecisions = $derived(decisions.filter(d => !d.decided));
	const decidedDecisions = $derived(decisions.filter(d => d.decided));

	function submitAdd() {
		if (!newDecisionQuestion.trim()) return;
		onAdd(newDecisionQuestion.trim());
		showAddDecision = false;
		newDecisionQuestion = '';
	}

	function submitDecide(decisionId: string) {
		if (!decisionAnswer.trim()) return;
		onDecide(decisionId, decisionAnswer.trim());
		decidingDecisionId = null;
		decisionAnswer = '';
	}
</script>

<div class="ng-section">
	<div class="ng-section__header">
		<h2 class="ng-section__title">Decision Queue</h2>
		<button
			onclick={() => showAddDecision = true}
			class="btn-pill btn-pill-ghost btn-pill-xs"
		>
			+ Add
		</button>
	</div>

	<!-- Add Decision Form -->
	{#if showAddDecision}
		<div class="ng-form-area ng-form-area--blue" transition:slide={{ duration: 200 }}>
			<label class="ng-label">Question to decide</label>
			<textarea
				bind:value={newDecisionQuestion}
				rows={2}
				class="ng-input ng-input--textarea"
				placeholder="What needs to be decided?"
			></textarea>
			<div class="ng-form-actions">
				<button
					onclick={() => { showAddDecision = false; newDecisionQuestion = ''; }}
					class="btn-pill btn-pill-ghost btn-pill-sm"
				>
					Cancel
				</button>
				<button
					onclick={submitAdd}
					disabled={!newDecisionQuestion.trim() || isSaving}
					class="btn-pill btn-pill-primary btn-pill-sm"
				>
					{isSaving ? 'Adding...' : 'Add Question'}
				</button>
			</div>
		</div>
	{/if}

	<!-- Pending Decisions -->
	{#if pendingDecisions.length > 0}
		<div class="ng-item-list">
			{#each pendingDecisions as decision (decision.id)}
				<div class="ng-item" transition:slide={{ duration: 200 }}>
					<div class="ng-item__row">
						<svg class="w-5 h-5 ng-icon--amber" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<div class="ng-item__body">
							<p class="ng-item__text">{decision.question}</p>
							<p class="ng-item__meta">
								Added {new Date(decision.added_at).toLocaleDateString()}
							</p>
						</div>
						<div class="ng-item__actions">
							<button
								onclick={() => { decidingDecisionId = decision.id; decisionAnswer = ''; }}
								class="btn-pill btn-pill-ghost btn-pill-xs"
							>
								Decide
							</button>
							<button
								onclick={() => onDelete(decision.id)}
								class="btn-pill btn-pill-danger btn-pill-icon"
								title="Delete"
								aria-label="Delete decision"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
						</div>
					</div>

					<!-- Decision Input -->
					{#if decidingDecisionId === decision.id}
						<div class="ng-item__decide-area" transition:slide={{ duration: 200 }}>
							<textarea
								bind:value={decisionAnswer}
								rows={2}
								class="ng-input ng-input--textarea"
								placeholder="What's your decision?"
							></textarea>
							<div class="ng-form-actions">
								<button
									onclick={() => { decidingDecisionId = null; decisionAnswer = ''; }}
									class="btn-pill btn-pill-ghost btn-pill-sm"
								>
									Cancel
								</button>
								<button
									onclick={() => submitDecide(decision.id)}
									disabled={!decisionAnswer.trim() || isSaving}
									class="btn-pill btn-pill-primary btn-pill-sm disabled:opacity-50"
								>
									{isSaving ? 'Saving...' : 'Confirm Decision'}
								</button>
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{:else if !showAddDecision}
		<p class="ng-empty-text">No pending decisions.</p>
	{/if}

	<!-- Decided History -->
	{#if decidedDecisions.length > 0}
		<div class="ng-divider">
			<button
				onclick={() => showDecidedHistory = !showDecidedHistory}
				class="btn-pill btn-pill-ghost btn-pill-sm ng-toggle-btn"
			>
				<svg
					class="w-4 h-4 ng-toggle-chevron {showDecidedHistory ? 'ng-toggle-chevron--open' : ''}"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
				{decidedDecisions.length} decided
			</button>

			{#if showDecidedHistory}
				<div class="ng-item-list" transition:slide={{ duration: 200 }}>
					{#each decidedDecisions as decision (decision.id)}
						<div class="ng-item ng-item--decided">
							<p class="ng-item__text ng-item__text--struck">{decision.question}</p>
							<p class="ng-item__decision">{decision.decision}</p>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.ng-section {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 0.75rem;
		padding: 1.25rem;
	}
	.ng-section__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}
	.ng-section__title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt);
	}
	.ng-form-area {
		margin-bottom: 1rem;
		padding: 0.75rem;
		border-radius: 0.5rem;
	}
	.ng-form-area--blue { background: rgba(59, 130, 246, 0.08); }
	.ng-label {
		display: block;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt2);
		margin-bottom: 0.25rem;
	}
	.ng-input {
		width: 100%;
		padding: 0.5rem 0.75rem;
		border: 1px solid var(--dbd);
		border-radius: 0.5rem;
		font-size: 0.875rem;
		color: var(--dt);
		background: var(--dbg);
	}
	.ng-input:focus {
		outline: none;
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.4);
	}
	.ng-input--textarea { resize: none; }
	.ng-form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		margin-top: 0.5rem;
	}
	.ng-item-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.ng-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding: 0.75rem;
		background: var(--dbg2);
		border-radius: 0.5rem;
	}
	.ng-item__row {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
	}
	.ng-icon--amber { color: #f59e0b; flex-shrink: 0; margin-top: 0.125rem; }
	.ng-item__body { flex: 1; min-width: 0; }
	.ng-item__text { font-size: 0.875rem; color: var(--dt); }
	.ng-item__meta { font-size: 0.75rem; color: var(--dt4); margin-top: 0.25rem; }
	.ng-item__actions { display: flex; align-items: center; gap: 0.25rem; }
	.ng-item__decide-area { margin-top: 0.5rem; padding-left: 2rem; }
	.ng-empty-text { color: var(--dt4); font-size: 0.875rem; }
	.ng-divider {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--dbd2);
	}
	.ng-toggle-btn { display: flex; align-items: center; gap: 0.5rem; }
	.ng-toggle-chevron { transition: transform 0.2s; }
	.ng-toggle-chevron--open { transform: rotate(90deg); }
	.ng-item--decided {
		padding: 0.75rem;
		background: rgba(34, 197, 94, 0.06);
		border-radius: 0.5rem;
	}
	.ng-item__text--struck {
		text-decoration: line-through;
		color: var(--dt3);
	}
	.ng-item__decision {
		font-size: 0.875rem;
		color: #16a34a;
		margin-top: 0.25rem;
		font-weight: 500;
	}
</style>
