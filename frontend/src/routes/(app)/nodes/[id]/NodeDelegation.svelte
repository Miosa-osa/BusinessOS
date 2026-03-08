<script lang="ts">
	import { slide } from 'svelte/transition';
	import { team } from '$lib/stores/team';
	import type { DelegationItem } from '$lib/api/nodes/types';

	type DelegationStatus = 'pending' | 'assigned' | 'in_progress' | 'done';

	interface Props {
		delegations: DelegationItem[];
		isSaving: boolean;
		onAdd: (task: string, assigneeId: string | null) => void;
		onUpdateStatus: (itemId: string, status: DelegationStatus) => void;
		onUpdateAssignee: (itemId: string, assigneeId: string | null) => void;
		onDelete: (itemId: string) => void;
	}

	let { delegations, isSaving, onAdd, onUpdateStatus, onUpdateAssignee, onDelete }: Props = $props();

	let showAddDelegation = $state(false);
	let newDelegationTask = $state('');
	let newDelegationAssigneeId: string | null = $state(null);
	let editingDelegationId: string | null = $state(null);

	const activeDelegations = $derived(delegations.filter(d => d.status !== 'done'));
	const completedDelegations = $derived(delegations.filter(d => d.status === 'done'));

	const delegationStatusConfig: Record<string, { statusClass: string; label: string }> = {
		pending: { statusClass: 'ng-status--pending', label: 'Pending' },
		assigned: { statusClass: 'ng-status--assigned', label: 'Assigned' },
		in_progress: { statusClass: 'ng-status--in-progress', label: 'In Progress' },
		done: { statusClass: 'ng-status--done', label: 'Done' }
	};

	function submitAdd() {
		if (!newDelegationTask.trim()) return;
		onAdd(newDelegationTask.trim(), newDelegationAssigneeId);
		showAddDelegation = false;
		newDelegationTask = '';
		newDelegationAssigneeId = null;
	}
</script>

<div class="ng-section">
	<div class="ng-section__header">
		<h2 class="ng-section__title">Delegation Ready</h2>
		<button
			onclick={() => showAddDelegation = true}
			class="btn-pill btn-pill-ghost btn-pill-xs"
		>
			+ Add
		</button>
	</div>

	<!-- Add Delegation Form -->
	{#if showAddDelegation}
		<div class="ng-form-area ng-form-area--purple" transition:slide={{ duration: 200 }}>
			<div class="ng-form-fields">
				<div>
					<label class="ng-label">Task to delegate</label>
					<input
						type="text"
						bind:value={newDelegationTask}
						class="ng-input"
						placeholder="Describe the task..."
					/>
				</div>
				<div>
					<label class="ng-label">Assign to (optional)</label>
					<select
						bind:value={newDelegationAssigneeId}
						class="ng-input"
					>
						<option value={null}>Unassigned</option>
						{#each $team.members as member}
							<option value={member.id}>{member.name}</option>
						{/each}
					</select>
				</div>
			</div>
			<div class="ng-form-actions">
				<button
					onclick={() => { showAddDelegation = false; newDelegationTask = ''; newDelegationAssigneeId = null; }}
					class="btn-pill btn-pill-ghost btn-pill-sm"
				>
					Cancel
				</button>
				<button
					onclick={submitAdd}
					disabled={!newDelegationTask.trim() || isSaving}
					class="btn-pill btn-pill-primary btn-pill-sm disabled:opacity-50"
				>
					{isSaving ? 'Adding...' : 'Add Task'}
				</button>
			</div>
		</div>
	{/if}

	<!-- Active Delegations -->
	{#if activeDelegations.length > 0}
		<div class="ng-item-list">
			{#each activeDelegations as item (item.id)}
				<div class="ng-item ng-item--hoverable" transition:slide={{ duration: 200 }}>
					<div class="ng-item__row">
						<svg class="w-5 h-5 ng-icon--purple" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
						</svg>
						<div class="ng-item__body">
							<p class="ng-item__text">{item.task}</p>
						</div>
						<div class="ng-item__actions">
							<!-- Status Selector -->
							<select
								value={item.status}
								onchange={(e) => onUpdateStatus(item.id, (e.target as HTMLSelectElement).value as DelegationStatus)}
								class="ng-status-select {delegationStatusConfig[item.status]?.statusClass || ''}"
							>
								<option value="pending">Pending</option>
								<option value="assigned">Assigned</option>
								<option value="in_progress">In Progress</option>
								<option value="done">Done</option>
							</select>
							<!-- Delete Button -->
							<button
								onclick={() => onDelete(item.id)}
								class="btn-pill btn-pill-danger btn-pill-icon ng-item__delete"
								title="Remove"
								aria-label="Remove delegation"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</button>
						</div>
					</div>

					<!-- Assignee Row -->
					<div class="ng-item__assignee">
						{#if editingDelegationId === item.id}
							<select
								value={item.assignee_id || ''}
								onchange={(e) => { onUpdateAssignee(item.id, (e.target as HTMLSelectElement).value || null); editingDelegationId = null; }}
								class="ng-input ng-input--sm"
							>
								<option value="">Unassigned</option>
								{#each $team.members as member}
									<option value={member.id}>{member.name}</option>
								{/each}
							</select>
							<button
								onclick={() => editingDelegationId = null}
								class="btn-pill btn-pill-ghost btn-pill-xs"
							>
								Cancel
							</button>
						{:else}
							<button
								onclick={() => editingDelegationId = item.id}
								class="btn-pill btn-pill-ghost btn-pill-xs ng-assignee-btn"
							>
								<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
								</svg>
								{item.assignee_name || 'Assign someone'}
							</button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{:else if !showAddDelegation}
		<p class="ng-empty-text">No items ready for delegation.</p>
	{/if}

	<!-- Completed Delegations -->
	{#if completedDelegations.length > 0}
		<div class="ng-divider">
			<p class="ng-completed-label">{completedDelegations.length} completed</p>
			<div class="ng-completed-list">
				{#each completedDelegations as item (item.id)}
					<div class="ng-completed-item">
						<svg class="w-4 h-4 ng-icon--green" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						<span class="ng-completed-item__text">{item.task}</span>
						<span class="ng-completed-item__assignee">{item.assignee_name || 'Unassigned'}</span>
						<button
							onclick={() => onDelete(item.id)}
							class="btn-pill btn-pill-danger btn-pill-icon"
							title="Remove"
							aria-label="Remove delegation"
						>
							<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				{/each}
			</div>
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
	.ng-form-area--purple { background: rgba(168, 85, 247, 0.08); }
	.ng-form-fields { display: flex; flex-direction: column; gap: 0.75rem; }
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
		box-shadow: 0 0 0 2px rgba(168, 85, 247, 0.4);
	}
	.ng-input--sm {
		font-size: 0.75rem;
		padding: 0.25rem 0.5rem;
		flex: 1;
	}
	.ng-form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
		margin-top: 0.75rem;
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
	.ng-item--hoverable .ng-item__delete {
		opacity: 0;
		transition: opacity 0.15s;
	}
	.ng-item--hoverable:hover .ng-item__delete { opacity: 1; }
	.ng-item__row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.ng-icon--purple { color: #a855f7; flex-shrink: 0; }
	.ng-icon--green { color: #22c55e; }
	.ng-item__body { flex: 1; min-width: 0; }
	.ng-item__text { font-size: 0.875rem; color: var(--dt); }
	.ng-item__actions { display: flex; align-items: center; gap: 0.5rem; }
	.ng-item__assignee {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding-left: 2rem;
	}
	.ng-assignee-btn { display: flex; align-items: center; gap: 0.25rem; }
	.ng-status-select {
		font-size: 0.75rem;
		padding: 0.25rem 0.5rem;
		border: 1px solid var(--dbd);
		border-radius: 0.5rem;
	}
	.ng-status--pending { background: var(--dbg2); color: var(--dt2); }
	.ng-status--assigned { background: rgba(59, 130, 246, 0.1); color: #2563eb; }
	.ng-status--in-progress { background: rgba(245, 158, 11, 0.1); color: #d97706; }
	.ng-status--done { background: rgba(34, 197, 94, 0.1); color: #16a34a; }
	.ng-empty-text { color: var(--dt4); font-size: 0.875rem; }
	.ng-divider {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--dbd2);
	}
	.ng-completed-label {
		font-size: 0.75rem;
		color: var(--dt3);
		margin-bottom: 0.5rem;
	}
	.ng-completed-list { display: flex; flex-direction: column; gap: 0.25rem; }
	.ng-completed-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem;
		border-radius: 0.5rem;
	}
	.ng-completed-item__text {
		font-size: 0.875rem;
		color: var(--dt3);
		text-decoration: line-through;
		flex: 1;
	}
	.ng-completed-item__assignee {
		font-size: 0.75rem;
		color: var(--dt4);
	}
</style>
