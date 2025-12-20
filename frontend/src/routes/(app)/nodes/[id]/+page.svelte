<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { fly, slide } from 'svelte/transition';
	import { api, type NodeDetail, type Node, type NodeType, type NodeHealth, type DecisionItem, type DelegationItem } from '$lib/api';

	// State
	let node: NodeDetail | null = $state(null);
	let children: Node[] = $state([]);
	let isLoading = $state(true);
	let error: string | null = $state(null);
	let isSaving = $state(false);

	// Edit states
	let editingPurpose = $state(false);
	let editingStatus = $state(false);
	let editingFocus = $state(false);
	let purposeValue = $state('');
	let statusValue = $state('');
	let focusValue = $state<string[]>([]);

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

	async function loadData() {
		isLoading = true;
		error = null;
		try {
			const [nodeData, childrenData] = await Promise.all([
				api.getNode(nodeId),
				api.getNodeChildren(nodeId)
			]);
			node = nodeData;
			children = childrenData;
			purposeValue = node.purpose || '';
			statusValue = node.current_status || '';
			focusValue = node.this_week_focus || [];
		} catch (e) {
			console.error('Failed to load node:', e);
			error = 'Failed to load node. Please try again.';
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadData();
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
			const result = await api.activateNode(node.id);
			node = { ...node, is_active: true };
		} catch (e) {
			console.error('Failed to activate node:', e);
		}
	}

	async function handleDeactivate() {
		if (!node) return;
		try {
			await api.deactivateNode(node.id);
			node = { ...node, is_active: false };
		} catch (e) {
			console.error('Failed to deactivate node:', e);
		}
	}

	async function updateHealth(health: NodeHealth) {
		if (!node) return;
		isSaving = true;
		try {
			await api.updateNode(node.id, { health });
			node = { ...node, health };
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
			await api.updateNode(node.id, { purpose: purposeValue });
			node = { ...node, purpose: purposeValue };
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
			await api.updateNode(node.id, { current_status: statusValue });
			node = { ...node, current_status: statusValue };
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
			await api.updateNode(node.id, { this_week_focus: focusValue.filter(f => f.trim()) });
			node = { ...node, this_week_focus: focusValue.filter(f => f.trim()) };
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
</script>

{#if isLoading}
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
							class="flex items-center gap-2 px-4 py-2 bg-blue-100 text-blue-700 rounded-lg hover:bg-blue-200 transition-colors"
						>
							<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
								<path d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
							Active
						</button>
					{:else}
						<button
							onclick={handleActivate}
							class="flex items-center gap-2 px-4 py-2 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
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
								class="text-sm text-blue-600 hover:text-blue-700"
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
									class="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg"
								>
									Cancel
								</button>
								<button
									onclick={savePurpose}
									disabled={isSaving}
									class="px-3 py-1.5 text-sm bg-gray-900 text-white rounded-lg hover:bg-gray-800 disabled:opacity-50"
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
									class="text-sm text-blue-600 hover:text-blue-700"
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
										class="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg"
									>
										Cancel
									</button>
									<button
										onclick={saveStatus}
										disabled={isSaving}
										class="px-3 py-1.5 text-sm bg-gray-900 text-white rounded-lg hover:bg-gray-800 disabled:opacity-50"
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
									class="text-sm text-blue-600 hover:text-blue-700"
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
										class="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-lg"
									>
										Cancel
									</button>
									<button
										onclick={saveFocus}
										disabled={isSaving}
										class="px-3 py-1.5 text-sm bg-gray-900 text-white rounded-lg hover:bg-gray-800 disabled:opacity-50"
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
							<button class="text-sm text-blue-600 hover:text-blue-700">+ Add</button>
						</div>

						{#if node.decision_queue && node.decision_queue.length > 0}
							<div class="space-y-2">
								{#each node.decision_queue as decision}
									<div class="flex items-start gap-3 p-3 bg-gray-50 rounded-lg">
										<svg class="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
										</svg>
										<div class="flex-1 min-w-0">
											<p class="text-sm text-gray-900">{decision.question}</p>
											<p class="text-xs text-gray-400 mt-1">Added {decision.added_at}</p>
										</div>
										<button class="text-sm text-blue-600 hover:text-blue-700">Decide</button>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-gray-400 text-sm">No pending decisions.</p>
						{/if}
					</div>

					<!-- Delegation Ready -->
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<div class="flex items-center justify-between mb-3">
							<h2 class="text-lg font-semibold text-gray-900">Delegation Ready</h2>
							<button class="text-sm text-blue-600 hover:text-blue-700">+ Add</button>
						</div>

						{#if node.delegation_ready && node.delegation_ready.length > 0}
							<div class="space-y-2">
								{#each node.delegation_ready as item}
									<div class="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
										<svg class="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
										</svg>
										<div class="flex-1 min-w-0">
											<p class="text-sm text-gray-900">{item.task}</p>
										</div>
										<span class="text-sm text-gray-500">
											{item.assignee_name || 'Unassigned'}
										</span>
									</div>
								{/each}
							</div>
						{:else}
							<p class="text-gray-400 text-sm">No items ready for delegation.</p>
						{/if}
					</div>
				</div>

				<!-- Child Nodes -->
				{#if children.length > 0}
					<div class="bg-white border border-gray-200 rounded-xl p-5">
						<div class="flex items-center justify-between mb-4">
							<h2 class="text-lg font-semibold text-gray-900">Child Nodes ({children.length})</h2>
							<a href="/nodes?parent={node.id}" class="text-sm text-blue-600 hover:text-blue-700">
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
					<h2 class="text-lg font-semibold text-gray-900 mb-4">Linked Items</h2>
					<div class="space-y-2">
						<a href="/contexts" class="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg transition-colors">
							<span class="flex items-center gap-3 text-gray-700">
								<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
								</svg>
								Context Profiles
							</span>
							<span class="text-sm text-gray-400">0 linked</span>
						</a>
						<a href="/projects" class="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg transition-colors">
							<span class="flex items-center gap-3 text-gray-700">
								<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
								Projects
							</span>
							<span class="text-sm text-gray-400">{node.linked_projects_count} linked</span>
						</a>
						<a href="/chat" class="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg transition-colors">
							<span class="flex items-center gap-3 text-gray-700">
								<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
								</svg>
								Conversations
							</span>
							<span class="text-sm text-gray-400">{node.linked_conversations_count} linked</span>
						</a>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
