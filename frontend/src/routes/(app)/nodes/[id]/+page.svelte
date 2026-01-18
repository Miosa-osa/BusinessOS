<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { fly, slide } from 'svelte/transition';
	import { nodes } from '$lib/stores/nodes';
	import { team } from '$lib/stores/team';
	import { getNodeChildren } from '$lib/api/nodes';
	import { LinkingModal } from '$lib/components/nodes';
	import type { Node, NodeHealth, DecisionItem, DelegationItem } from '$lib/api/nodes/types';

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

	// Decision Queue state
	let showAddDecision = $state(false);
	let newDecisionQuestion = $state('');
	let decidingDecisionId: string | null = $state(null);
	let decisionAnswer = $state('');
	let showDecidedHistory = $state(false);

	// Delegation state
	let showAddDelegation = $state(false);
	let newDelegationTask = $state('');
	let newDelegationAssigneeId: string | null = $state(null);
	let editingDelegationId: string | null = $state(null);

	// Linking state
	let showLinkingModal = $state(false);
	let expandedLinkedSection: 'projects' | 'contexts' | 'conversations' | null = $state(null);

	// Derive linked items from store
	const linkedProjects = $derived($nodes.currentNodeLinks?.projects ?? []);
	const linkedContexts = $derived($nodes.currentNodeLinks?.contexts ?? []);
	const linkedConversations = $derived($nodes.currentNodeLinks?.conversations ?? []);

	// Node type config
	const nodeTypeConfig: Record<string, { icon: string; color: string; label: string }> = {
		business: { icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4', color: 'blue', label: 'Business' },
		project: { icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z', color: 'green', label: 'Project' },
		learning: { icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253', color: 'purple', label: 'Learning' },
		operational: { icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z', color: 'orange', label: 'Operational' },
	};
	const defaultTypeConfig = { icon: 'M4 6h16M4 12h16M4 18h16', color: 'gray', label: 'Unknown' };

	// Health config
	const healthConfig: Record<string, { color: string; bgColor: string; label: string }> = {
		healthy: { color: 'text-green-600', bgColor: 'bg-green-500', label: 'Healthy' },
		needs_attention: { color: 'text-yellow-600', bgColor: 'bg-yellow-500', label: 'Needs Attention' },
		critical: { color: 'text-red-600', bgColor: 'bg-red-500', label: 'Critical' },
		not_started: { color: 'text-gray-500', bgColor: 'bg-gray-400', label: 'Not Started' },
	};
	const defaultHealthConfig = { color: 'text-gray-500', bgColor: 'bg-gray-400', label: 'Unknown' };

	// Helper functions to safely get config
	function getTypeConfig(type: string | undefined | null) {
		return (type && nodeTypeConfig[type]) || defaultTypeConfig;
	}

	function getHealthConfig(health: string | undefined | null) {
		return (health && healthConfig[health]) || defaultHealthConfig;
	}

	const nodeId = $derived($page.params.id);

	// Derive node from store
	const node = $derived($nodes.currentNode);

	async function loadData() {
		if (!nodeId) {
			error = 'No node ID provided';
			return;
		}
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
		// Load team members for delegation dropdown
		team.loadMembers('active');
		// Clear current node on unmount
		return () => nodes.clearCurrent();
	});

	// Reload when nodeId changes
	$effect(() => {
		if (nodeId) {
			loadData();
		}
	});

	async function handleActivate() {
		if (!node) return;
		try {
			await nodes.activate(node.id);
		} catch (e) {
			console.error('Failed to activate node:', e);
		}
	}

	async function handleDeactivate() {
		if (!node) return;
		try {
			await nodes.deactivate(node.id);
		} catch (e) {
			console.error('Failed to deactivate node:', e);
		}
	}

	async function updateHealth(health: NodeHealth) {
		if (!node) return;
		isSaving = true;
		try {
			await nodes.update(node.id, { health });
		} catch (e) {
			console.error('Failed to update health:', e);
		} finally {
			isSaving = false;
		}
	}

	async function savePurpose() {
		if (!node) return;
		isSaving = true;
		try {
			await nodes.update(node.id, { purpose: purposeValue });
			editingPurpose = false;
		} catch (e) {
			console.error('Failed to save purpose:', e);
		} finally {
			isSaving = false;
		}
	}

	async function saveStatus() {
		if (!node) return;
		isSaving = true;
		try {
			await nodes.update(node.id, { current_status: statusValue });
			editingStatus = false;
		} catch (e) {
			console.error('Failed to save status:', e);
		} finally {
			isSaving = false;
		}
	}

	async function saveFocus() {
		if (!node) return;
		isSaving = true;
		try {
			await nodes.update(node.id, { this_week_focus: focusValue.filter(f => f.trim()) });
			editingFocus = false;
		} catch (e) {
			console.error('Failed to save focus:', e);
		} finally {
			isSaving = false;
		}
	}

	function addFocusItem() {
		focusValue = [...focusValue, ''];
	}

	function removeFocusItem(index: number) {
		focusValue = focusValue.filter((_, i) => i !== index);
	}

	function updateFocusItem(index: number, value: string) {
		focusValue = focusValue.map((item, i) => i === index ? value : item);
	}

	// Decision Queue functions
	function generateId() {
		return crypto.randomUUID();
	}

	async function addDecision() {
		if (!node || !newDecisionQuestion.trim()) return;
		isSaving = true;
		try {
			const newDecision: DecisionItem = {
				id: generateId(),
				question: newDecisionQuestion.trim(),
				added_at: new Date().toISOString(),
				decided: false,
				decision: null
			};
			const currentQueue = node.decision_queue || [];
			await nodes.update(node.id, { decision_queue: [...currentQueue, newDecision] });
			showAddDecision = false;
			newDecisionQuestion = '';
		} catch (e) {
			console.error('Failed to add decision:', e);
		} finally {
			isSaving = false;
		}
	}

	async function makeDecision(decisionId: string) {
		if (!node || !decisionAnswer.trim()) return;
		isSaving = true;
		try {
			const updatedQueue = (node.decision_queue || []).map(d =>
				d.id === decisionId
					? { ...d, decided: true, decision: decisionAnswer.trim() }
					: d
			);
			await nodes.update(node.id, { decision_queue: updatedQueue });
			decidingDecisionId = null;
			decisionAnswer = '';
		} catch (e) {
			console.error('Failed to make decision:', e);
		} finally {
			isSaving = false;
		}
	}

	async function deleteDecision(decisionId: string) {
		if (!node) return;
		if (!confirm('Are you sure you want to delete this decision?')) return;
		isSaving = true;
		try {
			const updatedQueue = (node.decision_queue || []).filter(d => d.id !== decisionId);
			await nodes.update(node.id, { decision_queue: updatedQueue });
		} catch (e) {
			console.error('Failed to delete decision:', e);
		} finally {
			isSaving = false;
		}
	}

	// Derived values for decisions
	const pendingDecisions = $derived(
		(node?.decision_queue || []).filter(d => !d.decided)
	);
	const decidedDecisions = $derived(
		(node?.decision_queue || []).filter(d => d.decided)
	);

	// Delegation functions
	type DelegationStatus = 'pending' | 'assigned' | 'in_progress' | 'done';

	async function addDelegation() {
		if (!node || !newDelegationTask.trim()) return;
		isSaving = true;
		try {
			const assignee = $team.members.find(m => m.id === newDelegationAssigneeId);
			const newItem: DelegationItem = {
				id: generateId(),
				task: newDelegationTask.trim(),
				assignee_id: newDelegationAssigneeId,
				assignee_name: assignee?.name || null,
				status: newDelegationAssigneeId ? 'assigned' : 'pending'
			};
			const currentItems = node.delegation_ready || [];
			await nodes.update(node.id, { delegation_ready: [...currentItems, newItem] });
			showAddDelegation = false;
			newDelegationTask = '';
			newDelegationAssigneeId = null;
		} catch (e) {
			console.error('Failed to add delegation:', e);
		} finally {
			isSaving = false;
		}
	}

	async function updateDelegationStatus(itemId: string, status: DelegationStatus) {
		if (!node) return;
		isSaving = true;
		try {
			const updatedItems = (node.delegation_ready || []).map(d =>
				d.id === itemId ? { ...d, status } : d
			);
			await nodes.update(node.id, { delegation_ready: updatedItems });
		} catch (e) {
			console.error('Failed to update delegation:', e);
		} finally {
			isSaving = false;
		}
	}

	async function updateDelegationAssignee(itemId: string, assigneeId: string | null) {
		if (!node) return;
		isSaving = true;
		try {
			const assignee = $team.members.find(m => m.id === assigneeId);
			const updatedItems = (node.delegation_ready || []).map(d =>
				d.id === itemId
					? {
						...d,
						assignee_id: assigneeId,
						assignee_name: assignee?.name || null,
						status: assigneeId && d.status === 'pending' ? 'assigned' : d.status
					}
					: d
			);
			await nodes.update(node.id, { delegation_ready: updatedItems });
			editingDelegationId = null;
		} catch (e) {
			console.error('Failed to update delegation assignee:', e);
		} finally {
			isSaving = false;
		}
	}

	async function deleteDelegation(itemId: string) {
		if (!node) return;
		if (!confirm('Are you sure you want to remove this delegation item?')) return;
		isSaving = true;
		try {
			const updatedItems = (node.delegation_ready || []).filter(d => d.id !== itemId);
			await nodes.update(node.id, { delegation_ready: updatedItems });
		} catch (e) {
			console.error('Failed to delete delegation:', e);
		} finally {
			isSaving = false;
		}
	}

	// Derived values for delegation
	const activeDelegations = $derived(
		(node?.delegation_ready || []).filter(d => d.status !== 'done')
	);
	const completedDelegations = $derived(
		(node?.delegation_ready || []).filter(d => d.status === 'done')
	);

	const delegationStatusConfig: Record<string, { color: string; label: string }> = {
		pending: { color: 'bg-gray-200 text-gray-700', label: 'Pending' },
		assigned: { color: 'bg-blue-100 text-blue-700', label: 'Assigned' },
		in_progress: { color: 'bg-yellow-100 text-yellow-700', label: 'In Progress' },
		done: { color: 'bg-green-100 text-green-700', label: 'Done' }
	};
</script>

{#if $nodes.loading}
	<div class="h-full flex items-center justify-center">
		<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
	</div>
{:else if error || !node}
	<div class="h-full flex flex-col items-center justify-center">
		<p class="text-red-500 mb-4">{error || 'Node not found'}</p>
		<a href="/nodes" class="px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800">
			Back to Nodes
		</a>
	</div>
{:else}
	<div class="h-full flex flex-col bg-white overflow-hidden">
		<!-- Header -->
		<div class="border-b border-gray-200 px-6 py-4 flex-shrink-0">
			<div class="flex items-center gap-2 text-sm text-gray-500 mb-3">
				<a href="/nodes" class="hover:text-gray-700">Nodes</a>
				<span>/</span>
				{#if node.parent_name}
					<span>{node.parent_name}</span>
					<span>/</span>
				{/if}
				<span class="text-gray-900">{node.name}</span>
			</div>

			<div class="flex items-center justify-between">
				<div class="flex items-center gap-4">
					<div class="w-12 h-12 rounded-xl bg-{getTypeConfig(node.type).color}-100 text-{getTypeConfig(node.type).color}-600 flex items-center justify-center">
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(node.type).icon} />
						</svg>
					</div>
					<div>
						<h1 class="text-2xl font-semibold text-gray-900">{node.name}</h1>
						<div class="flex items-center gap-3 mt-1">
							<span class="text-sm text-gray-500 capitalize">{getTypeConfig(node.type).label} Node</span>
							<span class="text-gray-300">|</span>
							<span class="flex items-center gap-1.5 text-sm {getHealthConfig(node.health).color}">
								<span class="w-2 h-2 rounded-full {getHealthConfig(node.health).bgColor}"></span>
								{getHealthConfig(node.health).label}
							</span>
							<span class="text-gray-300">|</span>
							<span class="text-sm text-gray-400">
								Updated {new Date(node.updated_at).toLocaleDateString()}
							</span>
						</div>
					</div>
				</div>

				<div class="flex items-center gap-3">
					{#if node.is_active}
						<button
							onclick={handleDeactivate}
							class="btn-pill btn-pill-primary"
						>
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
							Active
						</button>
					{:else}
						<button
							onclick={handleActivate}
							class="btn-pill btn-pill-secondary"
						>
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
		<div class="flex-1 overflow-y-auto p-6">
			<div class="max-w-4xl mx-auto space-y-6">
				<!-- Purpose Section -->
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<div class="flex items-center justify-between mb-3">
						<h2 class="text-lg font-semibold text-gray-900">Purpose</h2>
						{#if !editingPurpose}
							<button
								onclick={() => editingPurpose = true}
								class="btn-pill btn-pill-link btn-pill-xs"
							>
								Edit
							</button>
						{/if}
					</div>

					{#if editingPurpose}
						<div transition:slide={{ duration: 200 }}>
							<textarea
								bind:value={purposeValue}
								rows={4}
								class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
								placeholder="Why does this node exist? What's its goal?"
							></textarea>
							<div class="flex justify-end gap-2 mt-3">
								<button
									onclick={() => { editingPurpose = false; purposeValue = node?.purpose || ''; }}
									class="btn-pill btn-pill-secondary btn-pill-sm"
								>
									Cancel
								</button>
								<button
									onclick={savePurpose}
									disabled={isSaving}
									class="btn-pill btn-pill-primary btn-pill-sm"
								>
									{isSaving ? 'Saving...' : 'Save'}
								</button>
							</div>
						</div>
					{:else}
						<p class="text-gray-600 whitespace-pre-wrap">
							{node.purpose || 'No purpose defined yet. Click Edit to add one.'}
						</p>
					{/if}
				</div>

				<!-- Status and Focus Row -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<!-- Current Status -->
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<div class="flex items-center justify-between mb-3">
							<div class="flex items-center gap-3">
								<h2 class="text-lg font-semibold text-gray-900">Current Status</h2>
								<!-- Health Selector -->
								<select
									value={node.health}
									onchange={(e) => updateHealth((e.target as HTMLSelectElement).value as NodeHealth)}
									class="text-sm px-2 py-1 border border-gray-200 rounded-lg {getHealthConfig(node.health).color} focus:outline-none focus:ring-2 focus:ring-blue-500"
								>
									{#each Object.entries(healthConfig) as [health, config]}
										<option value={health}>{config.label}</option>
									{/each}
								</select>
							</div>
							{#if !editingStatus}
								<button
									onclick={() => editingStatus = true}
									class="btn-pill btn-pill-link btn-pill-xs"
								>
									Update
								</button>
							{/if}
						</div>

						{#if editingStatus}
							<div transition:slide={{ duration: 200 }}>
								<textarea
									bind:value={statusValue}
									rows={4}
									class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
									placeholder="What's the current state of this node?"
								></textarea>
								<div class="flex justify-end gap-2 mt-3">
									<button
										onclick={() => { editingStatus = false; statusValue = node?.current_status || ''; }}
										class="btn-pill btn-pill-secondary btn-pill-sm"
									>
										Cancel
									</button>
									<button
										onclick={saveStatus}
										disabled={isSaving}
										class="btn-pill btn-pill-primary btn-pill-sm"
									>
										{isSaving ? 'Saving...' : 'Save'}
									</button>
								</div>
							</div>
						{:else}
							<p class="text-gray-600 whitespace-pre-wrap">
								{node.current_status || 'No status update yet.'}
							</p>
							{#if node.updated_at}
								<p class="text-xs text-gray-400 mt-2">
									Last updated: {new Date(node.updated_at).toLocaleString()}
								</p>
							{/if}
						{/if}
					</div>

					<!-- This Week's Focus -->
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<div class="flex items-center justify-between mb-3">
							<h2 class="text-lg font-semibold text-gray-900">This Week's Focus</h2>
							{#if !editingFocus}
								<button
									onclick={() => { editingFocus = true; focusValue = node?.this_week_focus || []; }}
									class="btn-pill btn-pill-link btn-pill-xs"
								>
									Edit
								</button>
							{/if}
						</div>

						{#if editingFocus}
							<div transition:slide={{ duration: 200 }}>
								<div class="space-y-2">
									{#each focusValue as item, i}
										<div class="flex items-center gap-2">
											<span class="text-sm text-gray-400 w-4">{i + 1}.</span>
											<input
												type="text"
												value={item}
												oninput={(e) => updateFocusItem(i, (e.target as HTMLInputElement).value)}
												class="flex-1 px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
												placeholder="Focus item..."
											/>
											<button
												onclick={() => removeFocusItem(i)}
												class="p-1 text-gray-400 hover:text-red-600"
											>
												<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
												</svg>
											</button>
										</div>
									{/each}
								</div>
								{#if focusValue.length < 5}
									<button
										onclick={addFocusItem}
										class="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 mt-2"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
										Add Focus Item
									</button>
								{/if}
								<div class="flex justify-end gap-2 mt-3">
									<button
										onclick={() => { editingFocus = false; focusValue = node?.this_week_focus || []; }}
										class="btn-pill btn-pill-secondary btn-pill-sm"
									>
										Cancel
									</button>
									<button
										onclick={saveFocus}
										disabled={isSaving}
										class="btn-pill btn-pill-primary btn-pill-sm"
									>
										{isSaving ? 'Saving...' : 'Save'}
									</button>
								</div>
							</div>
						{:else}
							{#if node.this_week_focus && node.this_week_focus.length > 0}
								<ol class="space-y-1">
									{#each node.this_week_focus as item, i}
										<li class="text-gray-600">
											<span class="text-gray-400 mr-2">{i + 1}.</span>
											{item}
										</li>
									{/each}
								</ol>
							{:else}
								<p class="text-gray-400">No focus items set for this week.</p>
							{/if}
						{/if}
					</div>
				</div>

				<!-- Decision Queue and Delegation -->
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<!-- Decision Queue -->
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<div class="flex items-center justify-between mb-3">
							<h2 class="text-lg font-semibold text-gray-900">Decision Queue</h2>
							<button
								onclick={() => showAddDecision = true}
								class="btn-pill btn-pill-link btn-pill-xs"
							>
								+ Add
							</button>
						</div>

						<!-- Add Decision Form -->
						{#if showAddDecision}
							<div class="mb-4 p-3 bg-blue-50 rounded-lg" transition:slide={{ duration: 200 }}>
								<label class="block text-sm font-medium text-gray-700 mb-1">Question to decide</label>
								<textarea
									bind:value={newDecisionQuestion}
									rows={2}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none text-sm"
									placeholder="What needs to be decided?"
								></textarea>
								<div class="flex justify-end gap-2 mt-2">
									<button
										onclick={() => { showAddDecision = false; newDecisionQuestion = ''; }}
										class="btn-pill btn-pill-secondary btn-pill-sm"
									>
										Cancel
									</button>
									<button
										onclick={addDecision}
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
							<div class="space-y-2">
								{#each pendingDecisions as decision (decision.id)}
									<div class="flex flex-col gap-2 p-3 bg-gray-50 rounded-lg" transition:slide={{ duration: 200 }}>
										<div class="flex items-start gap-3">
											<svg class="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
											</svg>
											<div class="flex-1 min-w-0">
												<p class="text-sm text-gray-900">{decision.question}</p>
												<p class="text-xs text-gray-400 mt-1">
													Added {new Date(decision.added_at).toLocaleDateString()}
												</p>
											</div>
											<div class="flex items-center gap-1">
												<button
													onclick={() => { decidingDecisionId = decision.id; decisionAnswer = ''; }}
													class="btn-pill btn-pill-link btn-pill-xs"
												>
													Decide
												</button>
												<button
													onclick={() => deleteDecision(decision.id)}
													class="p-1 text-gray-400 hover:text-red-600"
													title="Delete"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
													</svg>
												</button>
											</div>
										</div>

										<!-- Decision Input -->
										{#if decidingDecisionId === decision.id}
											<div class="mt-2 pl-8" transition:slide={{ duration: 200 }}>
												<textarea
													bind:value={decisionAnswer}
													rows={2}
													class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none text-sm"
													placeholder="What's your decision?"
												></textarea>
												<div class="flex justify-end gap-2 mt-2">
													<button
														onclick={() => { decidingDecisionId = null; decisionAnswer = ''; }}
														class="btn-pill btn-pill-secondary btn-pill-sm"
													>
														Cancel
													</button>
													<button
														onclick={() => makeDecision(decision.id)}
														disabled={!decisionAnswer.trim() || isSaving}
														class="btn-pill btn-pill-primary btn-pill-sm"
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
							<p class="text-gray-400 text-sm">No pending decisions.</p>
						{/if}

						<!-- Decided History -->
						{#if decidedDecisions.length > 0}
							<div class="mt-4 pt-4 border-t border-gray-100">
								<button
									onclick={() => showDecidedHistory = !showDecidedHistory}
									class="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700"
								>
									<svg
										class="w-4 h-4 transition-transform {showDecidedHistory ? 'rotate-90' : ''}"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
									{decidedDecisions.length} decided
								</button>

								{#if showDecidedHistory}
									<div class="mt-2 space-y-2" transition:slide={{ duration: 200 }}>
										{#each decidedDecisions as decision (decision.id)}
											<div class="p-3 bg-green-50 rounded-lg">
												<p class="text-sm text-gray-600 line-through">{decision.question}</p>
												<p class="text-sm text-green-700 mt-1 font-medium">{decision.decision}</p>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Delegation Ready -->
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<div class="flex items-center justify-between mb-3">
							<h2 class="text-lg font-semibold text-gray-900">Delegation Ready</h2>
							<button
								onclick={() => showAddDelegation = true}
								class="btn-pill btn-pill-link btn-pill-xs"
							>
								+ Add
							</button>
						</div>

						<!-- Add Delegation Form -->
						{#if showAddDelegation}
							<div class="mb-4 p-3 bg-purple-50 rounded-lg" transition:slide={{ duration: 200 }}>
								<div class="space-y-3">
									<div>
										<label class="block text-sm font-medium text-gray-700 mb-1">Task to delegate</label>
										<input
											type="text"
											bind:value={newDelegationTask}
											class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 text-sm"
											placeholder="Describe the task..."
										/>
									</div>
									<div>
										<label class="block text-sm font-medium text-gray-700 mb-1">Assign to (optional)</label>
										<select
											bind:value={newDelegationAssigneeId}
											class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 text-sm"
										>
											<option value={null}>Unassigned</option>
											{#each $team.members as member}
												<option value={member.id}>{member.name}</option>
											{/each}
										</select>
									</div>
								</div>
								<div class="flex justify-end gap-2 mt-3">
									<button
										onclick={() => { showAddDelegation = false; newDelegationTask = ''; newDelegationAssigneeId = null; }}
										class="btn-pill btn-pill-secondary btn-pill-sm"
									>
										Cancel
									</button>
									<button
										onclick={addDelegation}
										disabled={!newDelegationTask.trim() || isSaving}
										class="btn-pill btn-pill-primary btn-pill-sm"
									>
										{isSaving ? 'Adding...' : 'Add Task'}
									</button>
								</div>
							</div>
						{/if}

						<!-- Active Delegations -->
						{#if activeDelegations.length > 0}
							<div class="space-y-2">
								{#each activeDelegations as item (item.id)}
									<div class="flex flex-col gap-2 p-3 bg-gray-50 rounded-lg group" transition:slide={{ duration: 200 }}>
										<div class="flex items-center gap-3">
											<svg class="w-5 h-5 text-purple-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
											</svg>
											<div class="flex-1 min-w-0">
												<p class="text-sm text-gray-900">{item.task}</p>
											</div>
											<div class="flex items-center gap-2">
												<!-- Status Selector -->
												<select
													value={item.status}
													onchange={(e) => updateDelegationStatus(item.id, (e.target as HTMLSelectElement).value as DelegationStatus)}
													class="text-xs px-2 py-1 border border-gray-200 rounded-lg {delegationStatusConfig[item.status]?.color || ''}"
												>
													<option value="pending">Pending</option>
													<option value="assigned">Assigned</option>
													<option value="in_progress">In Progress</option>
													<option value="done">Done</option>
												</select>
												<!-- Delete Button -->
												<button
													onclick={() => deleteDelegation(item.id)}
													class="p-1 text-gray-400 hover:text-red-600 opacity-0 group-hover:opacity-100 transition-opacity"
													title="Remove"
												>
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
													</svg>
												</button>
											</div>
										</div>

										<!-- Assignee Row -->
										<div class="flex items-center gap-2 pl-8">
											{#if editingDelegationId === item.id}
												<select
													value={item.assignee_id || ''}
													onchange={(e) => updateDelegationAssignee(item.id, (e.target as HTMLSelectElement).value || null)}
													class="text-xs px-2 py-1 border border-gray-200 rounded-lg flex-1"
												>
													<option value="">Unassigned</option>
													{#each $team.members as member}
														<option value={member.id}>{member.name}</option>
													{/each}
												</select>
												<button
													onclick={() => editingDelegationId = null}
													class="text-xs text-gray-500 hover:text-gray-700"
												>
													Cancel
												</button>
											{:else}
												<button
													onclick={() => editingDelegationId = item.id}
													class="btn-pill btn-pill-link btn-pill-xs"
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
							<p class="text-gray-400 text-sm">No items ready for delegation.</p>
						{/if}

						<!-- Completed Delegations -->
						{#if completedDelegations.length > 0}
							<div class="mt-4 pt-4 border-t border-gray-100">
								<p class="text-xs text-gray-500 mb-2">{completedDelegations.length} completed</p>
								<div class="space-y-1">
									{#each completedDelegations as item (item.id)}
										<div class="flex items-center gap-2 p-2 rounded-lg">
											<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
											</svg>
											<span class="text-sm text-gray-500 line-through flex-1">{item.task}</span>
											<span class="text-xs text-gray-400">{item.assignee_name || 'Unassigned'}</span>
											<button
												onclick={() => deleteDelegation(item.id)}
												class="p-1 text-gray-400 hover:text-red-600"
												title="Remove"
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
				</div>

				<!-- Child Nodes -->
				{#if children.length > 0}
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<div class="flex items-center justify-between mb-4">
							<h2 class="text-lg font-semibold text-gray-900">Child Nodes ({children.length})</h2>
							<a href="/nodes?parent={node.id}" class="btn-pill btn-pill-link btn-pill-xs">
								+ Add Child
							</a>
						</div>

						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
							{#each children as child}
								<a
									href="/nodes/{child.id}"
									class="flex items-center gap-3 p-3 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
								>
									<div class="w-9 h-9 rounded-lg bg-{getTypeConfig(child.type).color}-100 text-{getTypeConfig(child.type).color}-600 flex items-center justify-center flex-shrink-0">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(child.type).icon} />
										</svg>
									</div>
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium text-gray-900 truncate">{child.name}</p>
										<span class="flex items-center gap-1 text-xs text-gray-500">
											<span class="w-1.5 h-1.5 rounded-full {getHealthConfig(child.health).bgColor}"></span>
											{getHealthConfig(child.health).label}
										</span>
									</div>
								</a>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Linked Items -->
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<div class="flex items-center justify-between mb-4">
						<h2 class="text-lg font-semibold text-gray-900">Linked Items</h2>
						<button
							onclick={() => showLinkingModal = true}
							class="btn-pill btn-pill-link btn-pill-xs"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
							</svg>
							Manage Links
						</button>
					</div>

					<div class="space-y-3">
						<!-- Projects Section -->
						<div class="border border-gray-200 rounded-lg overflow-hidden">
							<button
								onclick={() => expandedLinkedSection = expandedLinkedSection === 'projects' ? null : 'projects'}
								class="w-full flex items-center justify-between p-3 bg-gray-50 hover:bg-gray-100 transition-colors"
							>
								<span class="flex items-center gap-3 text-gray-700">
									<div class="w-8 h-8 rounded-lg bg-green-100 text-green-600 flex items-center justify-center">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
										</svg>
									</div>
									<span class="font-medium">Projects</span>
								</span>
								<div class="flex items-center gap-2">
									<span class="px-2 py-0.5 text-xs font-medium bg-gray-200 text-gray-700 rounded-full">{linkedProjects.length}</span>
									<svg class="w-4 h-4 text-gray-400 transition-transform {expandedLinkedSection === 'projects' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
									</svg>
								</div>
							</button>
							{#if expandedLinkedSection === 'projects'}
								<div class="p-3 border-t border-gray-200" transition:slide={{ duration: 200 }}>
									{#if linkedProjects.length === 0}
										<p class="text-sm text-gray-400 text-center py-2">No projects linked</p>
									{:else}
										<div class="space-y-2">
											{#each linkedProjects as project}
												<a
													href="/projects/{project.id}"
													class="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-lg transition-colors group"
												>
													<div class="w-6 h-6 rounded bg-green-100 text-green-600 flex items-center justify-center flex-shrink-0">
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
														</svg>
													</div>
													<div class="flex-1 min-w-0">
														<p class="text-sm font-medium text-gray-900 truncate">{project.name}</p>
														<p class="text-xs text-gray-500 capitalize">{project.status}</p>
													</div>
													<svg class="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
													</svg>
												</a>
											{/each}
										</div>
									{/if}
								</div>
							{/if}
						</div>

						<!-- Contexts Section -->
						<div class="border border-gray-200 rounded-lg overflow-hidden">
							<button
								onclick={() => expandedLinkedSection = expandedLinkedSection === 'contexts' ? null : 'contexts'}
								class="w-full flex items-center justify-between p-3 bg-gray-50 hover:bg-gray-100 transition-colors"
							>
								<span class="flex items-center gap-3 text-gray-700">
									<div class="w-8 h-8 rounded-lg bg-purple-100 text-purple-600 flex items-center justify-center">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
										</svg>
									</div>
									<span class="font-medium">Context Profiles</span>
								</span>
								<div class="flex items-center gap-2">
									<span class="px-2 py-0.5 text-xs font-medium bg-gray-200 text-gray-700 rounded-full">{linkedContexts.length}</span>
									<svg class="w-4 h-4 text-gray-400 transition-transform {expandedLinkedSection === 'contexts' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
									</svg>
								</div>
							</button>
							{#if expandedLinkedSection === 'contexts'}
								<div class="p-3 border-t border-gray-200" transition:slide={{ duration: 200 }}>
									{#if linkedContexts.length === 0}
										<p class="text-sm text-gray-400 text-center py-2">No context profiles linked</p>
									{:else}
										<div class="space-y-2">
											{#each linkedContexts as context}
												<a
													href="/knowledge-v2/{context.id}"
													class="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-lg transition-colors group"
												>
													<div class="w-6 h-6 rounded bg-purple-100 text-purple-600 flex items-center justify-center flex-shrink-0">
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
														</svg>
													</div>
													<div class="flex-1 min-w-0">
														<p class="text-sm font-medium text-gray-900 truncate">{context.name}</p>
														<p class="text-xs text-gray-500 capitalize">{context.type}</p>
													</div>
													<svg class="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
													</svg>
												</a>
											{/each}
										</div>
									{/if}
								</div>
							{/if}
						</div>

						<!-- Conversations Section -->
						<div class="border border-gray-200 rounded-lg overflow-hidden">
							<button
								onclick={() => expandedLinkedSection = expandedLinkedSection === 'conversations' ? null : 'conversations'}
								class="w-full flex items-center justify-between p-3 bg-gray-50 hover:bg-gray-100 transition-colors"
							>
								<span class="flex items-center gap-3 text-gray-700">
									<div class="w-8 h-8 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
										</svg>
									</div>
									<span class="font-medium">Conversations</span>
								</span>
								<div class="flex items-center gap-2">
									<span class="px-2 py-0.5 text-xs font-medium bg-gray-200 text-gray-700 rounded-full">{linkedConversations.length}</span>
									<svg class="w-4 h-4 text-gray-400 transition-transform {expandedLinkedSection === 'conversations' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
									</svg>
								</div>
							</button>
							{#if expandedLinkedSection === 'conversations'}
								<div class="p-3 border-t border-gray-200" transition:slide={{ duration: 200 }}>
									{#if linkedConversations.length === 0}
										<p class="text-sm text-gray-400 text-center py-2">No conversations linked</p>
									{:else}
										<div class="space-y-2">
											{#each linkedConversations as conversation}
												<a
													href="/chat/{conversation.id}"
													class="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-lg transition-colors group"
												>
													<div class="w-6 h-6 rounded bg-blue-100 text-blue-600 flex items-center justify-center flex-shrink-0">
														<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
															<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
														</svg>
													</div>
													<div class="flex-1 min-w-0">
														<p class="text-sm font-medium text-gray-900 truncate">{conversation.title || 'Untitled Conversation'}</p>
														<p class="text-xs text-gray-500">{new Date(conversation.updated_at).toLocaleDateString()}</p>
													</div>
													<svg class="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
													</svg>
												</a>
											{/each}
										</div>
									{/if}
								</div>
							{/if}
						</div>
					</div>
				</div>
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
