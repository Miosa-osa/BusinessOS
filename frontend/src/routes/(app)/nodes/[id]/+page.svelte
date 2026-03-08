<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { slide } from 'svelte/transition';
	import { nodes } from '$lib/stores/nodes';
	import { team } from '$lib/stores/team';
	import { getNodeChildren } from '$lib/api/nodes';
	import { LinkingModal } from '$lib/components/nodes';
	import NodeDecisionQueue from './NodeDecisionQueue.svelte';
	import NodeDelegation from './NodeDelegation.svelte';
	import NodeLinkedItems from './NodeLinkedItems.svelte';
	import type { Node, NodeHealth, DelegationItem } from '$lib/api/nodes/types';

	// State - children loaded separately (not in store)
	let children: Node[] = $state([]);
	let error: string | null = $state(null);
	let isSaving = $state(false);

	// Edit states
	let editingPurpose = $state(false);
	let editingStatus = $state(false);
	let editingFocus = $state(false);
	let purposeValue = $state('');
	let statusValue = $state('');
	let focusValue = $state<string[]>([]);

	// Linking state
	let showLinkingModal = $state(false);

	// Derive linked items from store
	const linkedProjects = $derived($nodes.currentNodeLinks?.projects ?? []);
	const linkedContexts = $derived($nodes.currentNodeLinks?.contexts ?? []);
	const linkedConversations = $derived($nodes.currentNodeLinks?.conversations ?? []);

	// Node type config
	const nodeTypeConfig: Record<string, { icon: string; typeClass: string; label: string }> = {
		business: { icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4', typeClass: 'ng-type-icon--business', label: 'Business' },
		project: { icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', typeClass: 'ng-type-icon--project', label: 'Project' },
		learning: { icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253', typeClass: 'ng-type-icon--learning', label: 'Learning' },
		operational: { icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z', typeClass: 'ng-type-icon--operational', label: 'Operational' },
	};
	const defaultTypeConfig = { icon: 'M4 6h16M4 12h16M4 18h16', typeClass: '', label: 'Unknown' };

	// Health config
	const healthConfig: Record<string, { colorClass: string; dotClass: string; label: string }> = {
		healthy: { colorClass: 'ng-health-text--healthy', dotClass: 'ng-health-dot--healthy', label: 'Healthy' },
		needs_attention: { colorClass: 'ng-health-text--attention', dotClass: 'ng-health-dot--attention', label: 'Needs Attention' },
		critical: { colorClass: 'ng-health-text--critical', dotClass: 'ng-health-dot--critical', label: 'Critical' },
		not_started: { colorClass: 'ng-health-text--not-started', dotClass: 'ng-health-dot--not-started', label: 'Not Started' },
	};
	const defaultHealthConfig = { colorClass: '', dotClass: '', label: 'Unknown' };

	function getTypeConfig(type: string | undefined | null) {
		return (type && nodeTypeConfig[type]) || defaultTypeConfig;
	}

	function getHealthConfig(health: string | undefined | null) {
		return (health && healthConfig[health]) || defaultHealthConfig;
	}

	const nodeId = $derived($page.params.id);
	const node = $derived($nodes.currentNode);

	async function loadData() {
		if (!nodeId) { error = 'No node ID provided'; return; }
		error = null;
		try {
			const [loadedNode, childrenData] = await Promise.all([
				nodes.loadById(nodeId),
				getNodeChildren(nodeId),
				nodes.loadLinks(nodeId)
			]);
			children = childrenData;
			if (loadedNode) {
				purposeValue = loadedNode.purpose || '';
				statusValue = loadedNode.current_status || '';
				focusValue = loadedNode.this_week_focus || [];
			}
		} catch (e) {
			console.error('Failed to load node:', e);
			error = 'Failed to load node. Please try again.';
		}
	}

	onMount(() => {
		loadData();
		team.loadMembers('active');
		return () => nodes.clearCurrent();
	});

	$effect(() => {
		if (nodeId) loadData();
	});

	async function handleActivate() {
		if (!node) return;
		try { await nodes.activate(node.id); } catch (e) { console.error('Failed to activate node:', e); }
	}

	async function handleDeactivate() {
		if (!node) return;
		try { await nodes.deactivate(node.id); } catch (e) { console.error('Failed to deactivate node:', e); }
	}

	async function updateHealth(health: NodeHealth) {
		if (!node) return;
		isSaving = true;
		try { await nodes.update(node.id, { health }); } catch (e) { console.error('Failed to update health:', e); } finally { isSaving = false; }
	}

	async function savePurpose() {
		if (!node) return;
		isSaving = true;
		try { await nodes.update(node.id, { purpose: purposeValue }); editingPurpose = false; } catch (e) { console.error('Failed to save purpose:', e); } finally { isSaving = false; }
	}

	async function saveStatus() {
		if (!node) return;
		isSaving = true;
		try { await nodes.update(node.id, { current_status: statusValue }); editingStatus = false; } catch (e) { console.error('Failed to save status:', e); } finally { isSaving = false; }
	}

	async function saveFocus() {
		if (!node) return;
		isSaving = true;
		try { await nodes.update(node.id, { this_week_focus: focusValue.filter(f => f.trim()) }); editingFocus = false; } catch (e) { console.error('Failed to save focus:', e); } finally { isSaving = false; }
	}

	function addFocusItem() { focusValue = [...focusValue, '']; }
	function removeFocusItem(index: number) { focusValue = focusValue.filter((_, i) => i !== index); }
	function updateFocusItem(index: number, value: string) { focusValue = focusValue.map((item, i) => i === index ? value : item); }

	// Decision Queue handlers
	function generateId() { return crypto.randomUUID(); }

	async function addDecision(question: string) {
		if (!node) return;
		isSaving = true;
		try {
			const newDecision = { id: generateId(), question, added_at: new Date().toISOString(), decided: false, decision: null };
			await nodes.update(node.id, { decision_queue: [...(node.decision_queue || []), newDecision] });
		} catch (e) { console.error('Failed to add decision:', e); } finally { isSaving = false; }
	}

	async function makeDecision(decisionId: string, answer: string) {
		if (!node) return;
		isSaving = true;
		try {
			const updatedQueue = (node.decision_queue || []).map(d =>
				d.id === decisionId ? { ...d, decided: true, decision: answer } : d
			);
			await nodes.update(node.id, { decision_queue: updatedQueue });
		} catch (e) { console.error('Failed to make decision:', e); } finally { isSaving = false; }
	}

	async function deleteDecision(decisionId: string) {
		if (!node) return;
		if (!confirm('Are you sure you want to delete this decision?')) return;
		isSaving = true;
		try {
			const updatedQueue = (node.decision_queue || []).filter(d => d.id !== decisionId);
			await nodes.update(node.id, { decision_queue: updatedQueue });
		} catch (e) { console.error('Failed to delete decision:', e); } finally { isSaving = false; }
	}

	// Delegation handlers
	type DelegationStatus = 'pending' | 'assigned' | 'in_progress' | 'done';

	async function addDelegation(task: string, assigneeId: string | null) {
		if (!node) return;
		isSaving = true;
		try {
			const assignee = $team.members.find(m => m.id === assigneeId);
			const newItem: DelegationItem = {
				id: generateId(), task,
				assignee_id: assigneeId,
				assignee_name: assignee?.name || null,
				status: assigneeId ? 'assigned' : 'pending'
			};
			await nodes.update(node.id, { delegation_ready: [...(node.delegation_ready || []), newItem] });
		} catch (e) { console.error('Failed to add delegation:', e); } finally { isSaving = false; }
	}

	async function updateDelegationStatus(itemId: string, status: DelegationStatus) {
		if (!node) return;
		isSaving = true;
		try {
			const updatedItems = (node.delegation_ready || []).map(d => d.id === itemId ? { ...d, status } : d);
			await nodes.update(node.id, { delegation_ready: updatedItems });
		} catch (e) { console.error('Failed to update delegation:', e); } finally { isSaving = false; }
	}

	async function updateDelegationAssignee(itemId: string, assigneeId: string | null) {
		if (!node) return;
		isSaving = true;
		try {
			const assignee = $team.members.find(m => m.id === assigneeId);
			const updatedItems = (node.delegation_ready || []).map(d =>
				d.id === itemId
					? { ...d, assignee_id: assigneeId, assignee_name: assignee?.name || null, status: assigneeId && d.status === 'pending' ? 'assigned' : d.status }
					: d
			);
			await nodes.update(node.id, { delegation_ready: updatedItems });
		} catch (e) { console.error('Failed to update delegation assignee:', e); } finally { isSaving = false; }
	}

	async function deleteDelegation(itemId: string) {
		if (!node) return;
		if (!confirm('Are you sure you want to remove this delegation item?')) return;
		isSaving = true;
		try {
			const updatedItems = (node.delegation_ready || []).filter(d => d.id !== itemId);
			await nodes.update(node.id, { delegation_ready: updatedItems });
		} catch (e) { console.error('Failed to delete delegation:', e); } finally { isSaving = false; }
	}
</script>

{#if $nodes.loading}
	<div class="ng-page ng-page--center">
		<div class="ng-spinner"></div>
	</div>
{:else if error || !node}
	<div class="ng-page ng-page--center">
		<p class="ng-error-text">{error || 'Node not found'}</p>
		<a href="/nodes" class="btn-pill btn-pill-primary">Back to Nodes</a>
	</div>
{:else}
	<div class="ng-page">
		<!-- Header -->
		<div class="ng-detail-header">
			<div class="ng-breadcrumb">
				<a href="/nodes" class="ng-breadcrumb__link">Nodes</a>
				<span class="ng-breadcrumb__sep">/</span>
				{#if node.parent_name}
					<span class="ng-breadcrumb__text">{node.parent_name}</span>
					<span class="ng-breadcrumb__sep">/</span>
				{/if}
				<span class="ng-breadcrumb__current">{node.name}</span>
			</div>

			<div class="ng-detail-header__row">
				<div class="ng-detail-header__left">
					{#if node}
					<div class="ng-type-icon ng-type-icon--xl {getTypeConfig(node.type).typeClass}">
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(node.type).icon} />
						</svg>
					</div>
					{/if}
					<div>
						<h1 class="ng-detail-header__title">{node.name}</h1>
						<div class="ng-detail-header__meta">
							<span class="ng-detail-header__label">{getTypeConfig(node.type).label} Node</span>
							<span class="ng-detail-header__divider">|</span>
							<span class="ng-detail-header__health {getHealthConfig(node.health).colorClass}">
								<span class="ng-health-dot {getHealthConfig(node.health).dotClass}"></span>
								{getHealthConfig(node.health).label}
							</span>
							<span class="ng-detail-header__divider">|</span>
							<span class="ng-detail-header__date">
								Updated {new Date(node.updated_at).toLocaleDateString()}
							</span>
						</div>
					</div>
				</div>

				<div class="ng-detail-header__actions">
					{#if node.is_active}
						<button onclick={handleDeactivate} class="btn-pill btn-pill-soft">
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
							Active
						</button>
					{:else}
						<button onclick={handleActivate} class="btn-pill btn-pill-ghost">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
							Activate
						</button>
					{/if}
				</div>
			</div>
		</div>

		<!-- Content -->
		<div class="ng-detail-content">
			<div class="ng-detail-content__inner">
				<!-- Purpose Section -->
				<div class="ng-section">
					<div class="ng-section__header">
						<h2 class="ng-section__title">Purpose</h2>
						{#if !editingPurpose}
							<button onclick={() => editingPurpose = true} class="btn-pill btn-pill-ghost btn-pill-xs">
								Edit
							</button>
						{/if}
					</div>

					{#if editingPurpose}
						<div transition:slide={{ duration: 200 }}>
							<textarea
								bind:value={purposeValue}
								rows={4}
								class="ng-input ng-input--textarea"
								placeholder="Why does this node exist? What's its goal?"
							></textarea>
							<div class="ng-section__actions">
								<button onclick={() => { editingPurpose = false; purposeValue = node?.purpose || ''; }} class="btn-pill btn-pill-ghost btn-pill-sm">Cancel</button>
								<button onclick={savePurpose} disabled={isSaving} class="btn-pill btn-pill-primary btn-pill-sm">
									{isSaving ? 'Saving...' : 'Save'}
								</button>
							</div>
						</div>
					{:else}
						<p class="ng-section__text">
							{node.purpose || 'No purpose defined yet. Click Edit to add one.'}
						</p>
					{/if}
				</div>

				<!-- Status and Focus Row -->
				<div class="ng-detail-row">
					<!-- Current Status -->
					<div class="ng-section">
						<div class="ng-section__header">
							<div class="ng-section__header-left">
								<h2 class="ng-section__title">Current Status</h2>
								<select
									value={node.health}
									onchange={(e) => updateHealth((e.target as HTMLSelectElement).value as NodeHealth)}
									class="ng-health-select {getHealthConfig(node.health).colorClass}"
								>
									{#each Object.entries(healthConfig) as [health, config]}
										<option value={health}>{config.label}</option>
									{/each}
								</select>
							</div>
							{#if !editingStatus}
								<button onclick={() => editingStatus = true} class="btn-pill btn-pill-ghost btn-pill-xs">Update</button>
							{/if}
						</div>

						{#if editingStatus}
							<div transition:slide={{ duration: 200 }}>
								<textarea
									bind:value={statusValue}
									rows={4}
									class="ng-input ng-input--textarea"
									placeholder="What's the current state of this node?"
								></textarea>
								<div class="ng-section__actions">
									<button onclick={() => { editingStatus = false; statusValue = node?.current_status || ''; }} class="btn-pill btn-pill-ghost btn-pill-sm">Cancel</button>
									<button onclick={saveStatus} disabled={isSaving} class="btn-pill btn-pill-primary btn-pill-sm">
										{isSaving ? 'Saving...' : 'Save'}
									</button>
								</div>
							</div>
						{:else}
							<p class="ng-section__text">{node.current_status || 'No status update yet.'}</p>
							{#if node.updated_at}
								<p class="ng-section__date">Last updated: {new Date(node.updated_at).toLocaleString()}</p>
							{/if}
						{/if}
					</div>

					<!-- This Week's Focus -->
					<div class="ng-section">
						<div class="ng-section__header">
							<h2 class="ng-section__title">This Week's Focus</h2>
							{#if !editingFocus}
								<button onclick={() => { editingFocus = true; focusValue = node?.this_week_focus || []; }} class="btn-pill btn-pill-ghost btn-pill-xs">Edit</button>
							{/if}
						</div>

						{#if editingFocus}
							<div transition:slide={{ duration: 200 }}>
								<div class="ng-focus-list">
									{#each focusValue as item, i}
										<div class="ng-focus-item">
											<span class="ng-focus-item__num">{i + 1}.</span>
											<input
												type="text"
												value={item}
												oninput={(e) => updateFocusItem(i, (e.target as HTMLInputElement).value)}
												class="ng-input"
												placeholder="Focus item..."
											/>
											<button onclick={() => removeFocusItem(i)} class="btn-pill btn-pill-danger btn-pill-icon" aria-label="Remove focus item">
												<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
												</svg>
											</button>
										</div>
									{/each}
								</div>
								{#if focusValue.length < 5}
									<button onclick={addFocusItem} class="btn-pill btn-pill-ghost btn-pill-sm ng-add-btn">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
										Add Focus Item
									</button>
								{/if}
								<div class="ng-section__actions">
									<button onclick={() => { editingFocus = false; focusValue = node?.this_week_focus || []; }} class="btn-pill btn-pill-ghost btn-pill-sm">Cancel</button>
									<button onclick={saveFocus} disabled={isSaving} class="btn-pill btn-pill-primary btn-pill-sm">
										{isSaving ? 'Saving...' : 'Save'}
									</button>
								</div>
							</div>
						{:else}
							{#if node.this_week_focus && node.this_week_focus.length > 0}
								<ol class="ng-focus-display">
									{#each node.this_week_focus as item, i}
										<li class="ng-focus-display__item">
											<span class="ng-focus-display__num">{i + 1}.</span>{item}
										</li>
									{/each}
								</ol>
							{:else}
								<p class="ng-section__empty">No focus items set for this week.</p>
							{/if}
						{/if}
					</div>
				</div>

				<!-- Decision Queue and Delegation -->
				<div class="ng-detail-row">
					<NodeDecisionQueue
						decisions={node.decision_queue || []}
						{isSaving}
						onAdd={addDecision}
						onDecide={makeDecision}
						onDelete={deleteDecision}
					/>

					<NodeDelegation
						delegations={node.delegation_ready || []}
						{isSaving}
						onAdd={addDelegation}
						onUpdateStatus={updateDelegationStatus}
						onUpdateAssignee={updateDelegationAssignee}
						onDelete={deleteDelegation}
					/>
				</div>

				<!-- Child Nodes -->
				{#if children.length > 0}
					<div class="ng-section">
						<div class="ng-section__header">
							<h2 class="ng-section__title">Child Nodes ({children.length})</h2>
							<a href="/nodes?parent={node.id}" class="btn-pill btn-pill-ghost btn-pill-xs">+ Add Child</a>
						</div>

						<div class="ng-children-grid">
							{#each children as child}
								<a href="/nodes/{child.id}" class="ng-child-card">
									<div class="ng-type-icon {getTypeConfig(child.type).typeClass}">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(child.type).icon} />
										</svg>
									</div>
									<div class="ng-child-card__info">
										<p class="ng-child-card__name">{child.name}</p>
										<span class="ng-child-card__health">
											<span class="ng-health-dot ng-health-dot--sm {getHealthConfig(child.health).dotClass}"></span>
											{getHealthConfig(child.health).label}
										</span>
									</div>
								</a>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Linked Items -->
				<NodeLinkedItems
					{linkedProjects}
					{linkedContexts}
					{linkedConversations}
					onManageLinks={() => showLinkingModal = true}
				/>
			</div>
		</div>
	</div>
{/if}

<!-- Linking Modal -->
{#if showLinkingModal && node}
	<LinkingModal
		nodeId={node.id}
		nodeName={node.name}
		onClose={() => showLinkingModal = false}
	/>
{/if}

<style>
	/* ── Page Shell ── */
	.ng-page { display: flex; flex-direction: column; height: 100%; background: var(--dbg); color: var(--dt); overflow: hidden; }
	.ng-page--center { align-items: center; justify-content: center; }
	.ng-spinner { width: 2rem; height: 2rem; border: 3px solid var(--dbd); border-top-color: #3b82f6; border-radius: 50%; animation: ng-spin 0.8s linear infinite; }
	@keyframes ng-spin { to { transform: rotate(360deg); } }
	.ng-error-text { color: #ef4444; margin-bottom: 1rem; }

	/* ── Breadcrumb ── */
	.ng-breadcrumb { display: flex; align-items: center; gap: 0.5rem; font-size: 0.875rem; margin-bottom: 0.75rem; }
	.ng-breadcrumb__link { color: var(--dt3); text-decoration: none; }
	.ng-breadcrumb__link:hover { color: var(--dt); }
	.ng-breadcrumb__sep { color: var(--dt4); }
	.ng-breadcrumb__text { color: var(--dt3); }
	.ng-breadcrumb__current { color: var(--dt); }

	/* ── Detail Header ── */
	.ng-detail-header { border-bottom: 1px solid var(--dbd); padding: 1rem 1.5rem; flex-shrink: 0; }
	.ng-detail-header__row { display: flex; align-items: center; justify-content: space-between; }
	.ng-detail-header__left { display: flex; align-items: center; gap: 1rem; }
	.ng-detail-header__title { font-size: 1.5rem; font-weight: 600; color: var(--dt); }
	.ng-detail-header__meta { display: flex; align-items: center; gap: 0.75rem; margin-top: 0.25rem; }
	.ng-detail-header__label { font-size: 0.875rem; color: var(--dt3); text-transform: capitalize; }
	.ng-detail-header__divider { color: var(--dt4); }
	.ng-detail-header__health { display: flex; align-items: center; gap: 0.375rem; font-size: 0.875rem; }
	.ng-detail-header__date { font-size: 0.875rem; color: var(--dt4); }
	.ng-detail-header__actions { display: flex; align-items: center; gap: 0.75rem; }

	/* ── Type Icon ── */
	.ng-type-icon {
		width: 2rem; height: 2rem; border-radius: 0.5rem; display: flex;
		align-items: center; justify-content: center; flex-shrink: 0;
	}
	.ng-type-icon--xl { width: 3rem; height: 3rem; border-radius: 0.75rem; }
	.ng-type-icon--business { background: rgba(59,130,246,.15); color: #3b82f6; }
	.ng-type-icon--project { background: rgba(168,85,247,.15); color: #a855f7; }
	.ng-type-icon--learning { background: rgba(34,197,94,.15); color: #22c55e; }
	.ng-type-icon--operational { background: rgba(245,158,11,.15); color: #f59e0b; }

	/* ── Health ── */
	.ng-health-dot { width: 0.5rem; height: 0.5rem; border-radius: 50%; flex-shrink: 0; display: inline-block; }
	.ng-health-dot--sm { width: 0.375rem; height: 0.375rem; }
	.ng-health-dot--healthy { background: #22c55e; }
	.ng-health-dot--attention { background: #f59e0b; }
	.ng-health-dot--critical { background: #ef4444; }
	.ng-health-dot--not-started { background: #9ca3af; }
	.ng-health-text--healthy { color: #22c55e; }
	.ng-health-text--attention { color: #eab308; }
	.ng-health-text--critical { color: #ef4444; }
	.ng-health-text--not-started { color: var(--dt3); }

	.ng-health-select {
		font-size: 0.8125rem; padding: 0.25rem 0.5rem; border: 1px solid var(--dbd);
		border-radius: 0.5rem; background: var(--dbg); outline: none;
	}
	.ng-health-select:focus { border-color: #3b82f6; }

	/* ── Content ── */
	.ng-detail-content { flex: 1; overflow-y: auto; padding: 1.5rem; }
	.ng-detail-content__inner { max-width: 56rem; margin: 0 auto; display: flex; flex-direction: column; gap: 1.5rem; }

	/* ── Section Cards ── */
	.ng-section {
		background: var(--dbg2); border: 1px solid var(--dbd); border-radius: 0.75rem; padding: 1.25rem;
	}
	.ng-section__header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.75rem; }
	.ng-section__header-left { display: flex; align-items: center; gap: 0.75rem; }
	.ng-section__title { font-size: 1.0625rem; font-weight: 600; color: var(--dt); }
	.ng-section__text { color: var(--dt2); white-space: pre-wrap; font-size: 0.875rem; }
	.ng-section__date { font-size: 0.75rem; color: var(--dt4); margin-top: 0.5rem; }
	.ng-section__empty { color: var(--dt4); font-size: 0.875rem; }
	.ng-section__actions { display: flex; justify-content: flex-end; gap: 0.5rem; margin-top: 0.75rem; }

	/* ── Detail Row (two cols) ── */
	.ng-detail-row { display: grid; grid-template-columns: 1fr; gap: 1.5rem; }
	@media (min-width: 768px) { .ng-detail-row { grid-template-columns: 1fr 1fr; } }

	/* ── Form Controls ── */
	.ng-input {
		width: 100%; padding: 0.5rem 0.75rem; font-size: 0.875rem; border-radius: 0.5rem;
		border: 1px solid var(--dbd); background: var(--dbg); color: var(--dt); transition: border-color 0.15s;
	}
	.ng-input:focus { outline: none; border-color: #3b82f6; }
	.ng-input--textarea { resize: none; }

	/* ── Focus List ── */
	.ng-focus-list { display: flex; flex-direction: column; gap: 0.5rem; }
	.ng-focus-item { display: flex; align-items: center; gap: 0.5rem; }
	.ng-focus-item__num { font-size: 0.875rem; color: var(--dt4); width: 1rem; flex-shrink: 0; }
	.ng-add-btn { display: inline-flex; align-items: center; gap: 0.25rem; margin-top: 0.5rem; }
	.ng-focus-display { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 0.25rem; }
	.ng-focus-display__item { color: var(--dt2); font-size: 0.875rem; }
	.ng-focus-display__num { color: var(--dt4); margin-right: 0.5rem; }

	/* ── Child Nodes ── */
	.ng-children-grid { display: grid; grid-template-columns: 1fr; gap: 0.75rem; }
	@media (min-width: 640px) { .ng-children-grid { grid-template-columns: repeat(2, 1fr); } }
	@media (min-width: 1024px) { .ng-children-grid { grid-template-columns: repeat(3, 1fr); } }
	.ng-child-card {
		display: flex; align-items: center; gap: 0.75rem; padding: 0.625rem 0.75rem;
		background: var(--dbg); border: 1px solid var(--dbd); border-radius: 0.5rem;
		text-decoration: none; color: inherit; transition: all 0.15s;
	}
	.ng-child-card:hover { background: var(--dbg3); border-color: var(--dt4); }
	.ng-child-card__info { flex: 1; min-width: 0; }
	.ng-child-card__name { font-size: 0.875rem; font-weight: 500; color: var(--dt); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.ng-child-card__health { display: flex; align-items: center; gap: 0.25rem; font-size: 0.75rem; color: var(--dt3); }
</style>
