<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { fly, slide } from 'svelte/transition';
	import { nodes } from '$lib/stores/nodes';
	import { NodeGraphView, NodeBuildingView, NodeBuilding3D } from '$lib/components/nodes';
	import type { NodeTree, NodeType, NodeHealth, CreateNodeData } from '$lib/api/nodes/types';

	// Sanitize user input to prevent XSS and injection attacks
	function sanitizeInput(input: string): string {
		return input
			.replace(/[<>]/g, '') // Remove angle brackets
			.replace(/javascript:/gi, '') // Remove javascript: protocol
			.replace(/on\w+=/gi, '') // Remove event handlers
			.trim();
	}

	// Debounce utility for search input
	function debounce<T extends (...args: Parameters<T>) => void>(fn: T, delay: number): T {
		let timeoutId: ReturnType<typeof setTimeout>;
		return ((...args: Parameters<T>) => {
			clearTimeout(timeoutId);
			timeoutId = setTimeout(() => fn(...args), delay);
		}) as T;
	}

	// View state (local - UI specific)
	let viewMode: 'tree' | 'list' | 'grid' | 'graph' | 'building' | 'building3d' = $state('tree');
	let selectedGraphNode: string | null = $state(null);
	let selectedBuildingNode: string | null = $state(null);
	let searchInput = $state(''); // Raw input value
	let searchQuery = $state(''); // Debounced value used for filtering
	let showNewNodeModal = $state(false);
	let expandedNodes: Set<string> = $state(new Set());

	// Debounced search handler (300ms delay)
	const updateSearchQuery = debounce((value: string) => {
		searchQuery = value;
	}, 300);

	// Filter state
	let showFilterDropdown = $state(false);
	let filterType: NodeType | 'all' = $state('all');
	let filterHealth: NodeHealth | 'all' = $state('all');
	let showArchived = $state(false);

	// New node form
	let newNodeName = $state('');
	let newNodeType: NodeType = $state('business');
	let newNodeParentId: string | null = $state(null);
	let newNodePurpose = $state('');
	let isCreatingNode = $state(false);

	// Error state (local)
	let error: string | null = $state(null);

	// Node type config with Foundation CSS classes
	const nodeTypeConfig: Record<string, { icon: string; typeClass: string; label: string }> = {
		business: {
			icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4',
			typeClass: 'ng-type-icon--business',
			label: 'Business'
		},
		project: {
			icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z',
			typeClass: 'ng-type-icon--project',
			label: 'Project'
		},
		learning: {
			icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253',
			typeClass: 'ng-type-icon--learning',
			label: 'Learning'
		},
		operational: {
			icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
			typeClass: 'ng-type-icon--operational',
			label: 'Operational'
		},
	};
	const defaultTypeConfig = { icon: 'M4 6h16M4 12h16M4 18h16', typeClass: '', label: 'Unknown' };

	// Health config
	const healthConfig: Record<string, { colorClass: string; label: string }> = {
		healthy: { colorClass: 'ng-health-dot--healthy', label: 'Healthy' },
		needs_attention: { colorClass: 'ng-health-dot--attention', label: 'Needs Attention' },
		critical: { colorClass: 'ng-health-dot--critical', label: 'Critical' },
		not_started: { colorClass: 'ng-health-dot--not-started', label: 'Not Started' },
	};
	const defaultHealthConfig = { colorClass: '', label: 'Unknown' };

	// Helper functions to safely get config
	function getTypeConfig(type: string | undefined | null) {
		return (type && nodeTypeConfig[type]) || defaultTypeConfig;
	}

	function getHealthConfig(health: string | undefined | null) {
		return (health && healthConfig[health]) || defaultHealthConfig;
	}

	async function loadData() {
		error = null;
		try {
			await Promise.all([
				nodes.loadTree(showArchived),
				nodes.loadActive()
			]);
		} catch (e) {
			console.error('Failed to load nodes:', e);
			error = 'Failed to load nodes. Please try again.';
		}
	}

	onMount(() => {
		loadData();
	});

	function toggleExpand(nodeId: string) {
		const newExpanded = new Set(expandedNodes);
		if (newExpanded.has(nodeId)) {
			newExpanded.delete(nodeId);
		} else {
			newExpanded.add(nodeId);
		}
		expandedNodes = newExpanded;
	}

	async function handleActivate(nodeId: string) {
		try {
			await nodes.activate(nodeId);
			await nodes.loadTree(showArchived);
		} catch (e) {
			console.error('Failed to activate node:', e);
		}
	}

	async function handleDeactivate() {
		const activeNode = $nodes.activeNode;
		if (!activeNode) return;
		try {
			await nodes.deactivate(activeNode.id);
			await nodes.loadTree(showArchived);
		} catch (e) {
			console.error('Failed to deactivate node:', e);
		}
	}

	async function handleDelete(nodeId: string) {
		if (!confirm('Are you sure you want to delete this node? All children will also be deleted.')) return;
		try {
			await nodes.delete(nodeId);
			await nodes.loadTree(showArchived);
		} catch (e) {
			console.error('Failed to delete node:', e);
		}
	}

	async function handleCreateNode() {
		if (!newNodeName.trim()) return;
		isCreatingNode = true;
		try {
			// Sanitize user inputs before sending to backend
			const sanitizedName = sanitizeInput(newNodeName);
			const sanitizedPurpose = sanitizeInput(newNodePurpose);

			if (!sanitizedName) {
				console.error('Invalid node name after sanitization');
				return;
			}

			const data: CreateNodeData = {
				name: sanitizedName,
				type: newNodeType,
			};
			if (newNodeParentId) data.parent_id = newNodeParentId;
			if (sanitizedPurpose) data.purpose = sanitizedPurpose;

			await nodes.create(data);
			showNewNodeModal = false;
			newNodeName = '';
			newNodeType = 'business';
			newNodeParentId = null;
			newNodePurpose = '';
			await nodes.loadTree(showArchived);
		} catch (e) {
			console.error('Failed to create node:', e);
		} finally {
			isCreatingNode = false;
		}
	}

	// Filter nodes
	function filterNodes(nodeList: NodeTree[]): NodeTree[] {
		if (!nodeList) return [];
		return nodeList.filter(node => {
			const matchesSearch = !searchQuery || node.name.toLowerCase().includes(searchQuery.toLowerCase());
			const matchesType = filterType === 'all' || node.type === filterType;
			const matchesHealth = filterHealth === 'all' || node.health === filterHealth;
			return matchesSearch && matchesType && matchesHealth;
		}).map(node => ({
			...node,
			children: filterNodes(node.children || [])
		}));
	}

	const filteredNodes = $derived(filterNodes($nodes.nodeTree));

	// Flatten nodes for list view
	function flattenNodes(nodeList: NodeTree[], depth = 0): (NodeTree & { depth: number })[] {
		if (!nodeList) return [];
		let result: (NodeTree & { depth: number })[] = [];
		for (const node of nodeList) {
			result.push({ ...node, depth });
			result = result.concat(flattenNodes(node.children || [], depth + 1));
		}
		return result;
	}

	const flatNodes = $derived(flattenNodes(filteredNodes));

	// Get all nodes for parent selector
	function getAllNodes(nodeList: NodeTree[]): NodeTree[] {
		if (!nodeList) return [];
		let result: NodeTree[] = [];
		for (const node of nodeList) {
			result.push(node);
			result = result.concat(getAllNodes(node.children || []));
		}
		return result;
	}

	const allNodes = $derived(getAllNodes($nodes.nodeTree));
</script>

<div class="ng-page">
	<!-- Header -->
	<div class="ng-header">
		<div class="ng-header__row">
			<div>
				<h1 class="ng-header__title">Business Nodes</h1>
				<p class="ng-header__sub">Your cognitive operating system structure</p>
			</div>
			<button
				onclick={() => showNewNodeModal = true}
				class="btn-pill btn-pill-primary btn-pill-sm"
			>
				<svg class="ng-btn-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				New Node
			</button>
		</div>
	</div>

	<!-- Toolbar -->
	<div class="ng-toolbar">
		<div class="ng-toolbar__row">
			<!-- View Switcher -->
			<div class="ng-view-tabs">
				<button
					onclick={() => viewMode = 'tree'}
					class="ng-view-tab {viewMode === 'tree' ? 'ng-view-tab--active' : ''}"
				>
					<svg class="ng-view-tab__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
					</svg>
					Tree
				</button>
				<button
					onclick={() => viewMode = 'list'}
					class="ng-view-tab {viewMode === 'list' ? 'ng-view-tab--active' : ''}"
				>
					<svg class="ng-view-tab__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
					List
				</button>
				<button
					onclick={() => viewMode = 'grid'}
					class="ng-view-tab {viewMode === 'grid' ? 'ng-view-tab--active' : ''}"
				>
					<svg class="ng-view-tab__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
					</svg>
					Grid
				</button>
				<button
					onclick={() => viewMode = 'graph'}
					class="ng-view-tab {viewMode === 'graph' ? 'ng-view-tab--active' : ''}"
				>
					<svg class="ng-view-tab__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
					</svg>
					Graph
				</button>
				<button
					onclick={() => viewMode = 'building'}
					class="ng-view-tab {viewMode === 'building' ? 'ng-view-tab--active' : ''}"
				>
					<svg class="ng-view-tab__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
					</svg>
					2D
				</button>
				<button
					onclick={() => viewMode = 'building3d'}
					class="ng-view-tab {viewMode === 'building3d' ? 'ng-view-tab--active' : ''}"
				>
					<svg class="ng-view-tab__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
					</svg>
					3D
				</button>
			</div>

			<div class="ng-toolbar__right">
				<!-- Filter Dropdown -->
				<div class="ng-filter-wrap">
					<button
						onclick={() => showFilterDropdown = !showFilterDropdown}
						class="btn-pill btn-pill-ghost btn-pill-sm"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
						</svg>
						Filter
						{#if filterType !== 'all' || filterHealth !== 'all' || showArchived}
							<span class="ng-filter-indicator"></span>
						{/if}
					</button>

					{#if showFilterDropdown}
						<div
							class="ng-filter-panel"
							transition:fly={{ y: -10, duration: 200 }}
						>
							<div class="ng-filter-panel__sections">
								<div>
									<label class="ng-filter-panel__label">Type</label>
									<div class="ng-filter-panel__options">
										<label class="ng-filter-panel__option">
											<input type="radio" bind:group={filterType} value="all" />
											<span>All Types</span>
										</label>
										{#each Object.entries(nodeTypeConfig) as [type, config]}
											<label class="ng-filter-panel__option">
												<input type="radio" bind:group={filterType} value={type} />
												<span>{config.label}</span>
											</label>
										{/each}
									</div>
								</div>

								<div>
									<label class="ng-filter-panel__label">Health</label>
									<div class="ng-filter-panel__options">
										<label class="ng-filter-panel__option">
											<input type="radio" bind:group={filterHealth} value="all" />
											<span>All Health</span>
										</label>
										{#each Object.entries(healthConfig) as [health, config]}
											<label class="ng-filter-panel__option">
												<input type="radio" bind:group={filterHealth} value={health} />
												<span class="ng-health-dot {config.colorClass}"></span>
												<span>{config.label}</span>
											</label>
										{/each}
									</div>
								</div>

								<div>
									<label class="ng-filter-panel__option">
										<input type="checkbox" bind:checked={showArchived} onchange={() => loadData()} />
										<span>Show Archived</span>
									</label>
								</div>

								<div class="ng-filter-panel__footer">
									<button
										onclick={() => { filterType = 'all'; filterHealth = 'all'; showArchived = false; loadData(); }}
										class="ng-filter-panel__clear"
									>
										Clear
									</button>
									<button
										onclick={() => showFilterDropdown = false}
										class="ng-filter-panel__apply"
									>
										Apply
									</button>
								</div>
							</div>
						</div>
					{/if}
				</div>

				<!-- Search -->
				<div class="ng-search-wrap">
					<svg class="ng-search-wrap__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						placeholder="Search nodes..."
						bind:value={searchInput}
						oninput={(e) => updateSearchQuery(e.currentTarget.value)}
						class="ng-search-input"
					/>
				</div>
			</div>
		</div>
	</div>

	<!-- Active Node Banner -->
	{#if $nodes.activeNode}
		<div class="ng-active-banner" transition:slide>
			<div class="ng-active-banner__left">
				<svg class="w-5 h-5" style="color: #3b82f6" fill="currentColor" viewBox="0 0 24 24">
					<path d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
				<span class="ng-active-banner__text">
					Active Node: <strong>{$nodes.activeNode.name}</strong>
				</span>
			</div>
			<div class="ng-active-banner__right">
				<a
					href="/nodes/{$nodes.activeNode.id}"
					class="btn-pill btn-pill-ghost btn-pill-xs"
				>
					View
				</a>
				<button
					onclick={handleDeactivate}
					class="btn-pill btn-pill-ghost btn-pill-xs"
				>
					Deactivate
				</button>
			</div>
		</div>
	{/if}

	<!-- Content -->
	<div class="ng-content">
		{#if $nodes.loading}
			<div class="ng-empty">
				<div class="ng-spinner"></div>
			</div>
		{:else if error}
			<div class="ng-empty">
				<p style="color: #ef4444; margin-bottom: 12px;">{error}</p>
				<button onclick={loadData} class="btn-pill btn-pill-primary">
					Retry
				</button>
			</div>
		{:else if filteredNodes.length === 0}
			<div class="ng-empty">
				<div class="ng-empty__icon-wrap">
					<svg class="w-8 h-8" style="color: var(--dt4, #bbb)" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
					</svg>
				</div>
				<h3 class="ng-empty__title">No nodes yet</h3>
				<p class="ng-empty__text">Create your first node to organize your business into manageable focus areas.</p>
				<button
					onclick={() => showNewNodeModal = true}
					class="btn-pill btn-pill-primary btn-pill-sm"
				>
					Create your first node
				</button>
			</div>
		{:else if viewMode === 'tree'}
			<!-- Tree View -->
			<div class="ng-tree">
				{#snippet treeNode(node: NodeTree, depth: number = 0)}
					<div>
						<div
							class="ng-tree-item"
							style="padding-left: {depth * 24 + 12}px"
						>
							<!-- Expand/Collapse -->
							{#if node.children.length > 0}
								<button
									onclick={() => toggleExpand(node.id)}
									class="ng-tree-item__expand"
								>
									<svg
										class="w-4 h-4 {expandedNodes.has(node.id) ? 'ng-tree-item__chevron--open' : ''}"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</button>
							{:else}
								<div style="width: 20px;"></div>
							{/if}

							<!-- Type Icon -->
							<div class="ng-type-icon {getTypeConfig(node.type).typeClass}">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(node.type).icon} />
								</svg>
							</div>

							<!-- Name -->
							<a href="/nodes/{node.id}" class="ng-tree-item__name">
								{node.name}
							</a>

							<!-- Active indicator -->
							{#if node.is_active}
								<span class="ng-active-tag">Active</span>
							{/if}

							<!-- Health -->
							<div class="ng-health-dot {getHealthConfig(node.health).colorClass}"></div>

							<!-- Actions -->
							<div class="ng-tree-item__actions">
								{#if !node.is_active}
									<button
										onclick={() => handleActivate(node.id)}
										class="ng-tree-item__action-btn"
										title="Activate"
									>
										<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
											<path d="M13 10V3L4 14h7v7l9-11h-7z" />
										</svg>
									</button>
								{/if}
								<button
									onclick={() => handleDelete(node.id)}
									class="ng-tree-item__action-btn ng-tree-item__action-btn--danger"
									title="Delete"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
									</svg>
								</button>
							</div>
						</div>

						{#if expandedNodes.has(node.id) && node.children.length > 0}
							<div transition:slide={{ duration: 200 }}>
								{#each node.children as child}
									{@render treeNode(child, depth + 1)}
								{/each}
							</div>
						{/if}
					</div>
				{/snippet}

				{#each filteredNodes as node}
					{@render treeNode(node)}
				{/each}
			</div>
		{:else if viewMode === 'list'}
			<!-- List View -->
			<div class="ng-table-wrap">
				<table class="ng-table">
					<thead class="ng-table__head">
						<tr>
							<th class="ng-table__th">Name</th>
							<th class="ng-table__th">Type</th>
							<th class="ng-table__th">Health</th>
							<th class="ng-table__th">Updated</th>
							<th class="ng-table__th" style="text-align: right;">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each flatNodes as node}
							<tr class="ng-table__row">
								<td class="ng-table__td">
									<div class="ng-table__name-cell" style="padding-left: {node.depth * 20}px">
										<div class="ng-type-icon {getTypeConfig(node.type).typeClass}">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(node.type).icon} />
											</svg>
										</div>
										<a href="/nodes/{node.id}" class="ng-tree-item__name">
											{node.name}
										</a>
										{#if node.is_active}
											<span class="ng-active-tag">Active</span>
										{/if}
									</div>
								</td>
								<td class="ng-table__td ng-table__td--muted" style="text-transform: capitalize;">{node.type}</td>
								<td class="ng-table__td">
									<span class="ng-table__health-cell">
										<span class="ng-health-dot {getHealthConfig(node.health).colorClass}"></span>
										<span class="ng-table__td--muted">{getHealthConfig(node.health).label}</span>
									</span>
								</td>
								<td class="ng-table__td ng-table__td--muted">
									{new Date(node.updated_at).toLocaleDateString()}
								</td>
								<td class="ng-table__td" style="text-align: right;">
									<div class="ng-tree-item__actions" style="opacity: 1; justify-content: flex-end;">
										{#if !node.is_active}
											<button
												onclick={() => handleActivate(node.id)}
												class="ng-tree-item__action-btn"
											>
												<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
													<path d="M13 10V3L4 14h7v7l9-11h-7z" />
												</svg>
											</button>
										{/if}
										<button
											onclick={() => handleDelete(node.id)}
											class="ng-tree-item__action-btn ng-tree-item__action-btn--danger"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
											</svg>
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else if viewMode === 'grid'}
			<!-- Grid View -->
			<div class="ng-grid">
				{#each flatNodes as node}
					<a href="/nodes/{node.id}" class="ng-card">
						<div class="ng-card__top">
							<div class="ng-type-icon ng-type-icon--lg {getTypeConfig(node.type).typeClass}">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(node.type).icon} />
								</svg>
							</div>
							<div class="ng-card__info">
								<h3 class="ng-card__name">{node.name}</h3>
								<span class="ng-card__health">
									<span class="ng-health-dot {getHealthConfig(node.health).colorClass}"></span>
									<span>{getHealthConfig(node.health).label}</span>
								</span>
							</div>
							{#if node.is_active}
								<svg class="ng-card__active-icon" fill="currentColor" viewBox="0 0 24 24">
									<path d="M13 10V3L4 14h7v7l9-11h-7z" />
								</svg>
							{/if}
						</div>

						{#if node.this_week_focus && node.this_week_focus.length > 0}
							<div class="ng-card__focus">
								<p class="ng-card__focus-label">This week:</p>
								<p class="ng-card__focus-text">{node.this_week_focus[0]}</p>
							</div>
						{/if}

						{#if node.children_count > 0}
							<p class="ng-card__children">{node.children_count} child nodes</p>
						{/if}
					</a>
				{/each}

				<!-- Add New Card -->
				<button onclick={() => showNewNodeModal = true} class="ng-card ng-card--add">
					<svg class="ng-card--add__icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					<span class="ng-card--add__label">Add New Node</span>
				</button>
			</div>
		{:else if viewMode === 'graph'}
			<!-- Graph View -->
			<div class="ng-canvas-wrap">
				<NodeGraphView
					nodes={$nodes.nodeTree}
					activeNodeId={$nodes.activeNode?.id}
					selectedId={selectedGraphNode}
					onSelect={(node) => { if (node?.id) selectedGraphNode = node.id; }}
					onNavigate={(node) => { if (node?.id) goto(`/nodes/${node.id}`); }}
				/>
			</div>
		{:else if viewMode === 'building'}
			<!-- Building 2D View -->
			<div class="ng-canvas-wrap ng-canvas-wrap--full">
				<NodeBuildingView
					nodes={$nodes.nodeTree}
					activeNodeId={$nodes.activeNode?.id}
					selectedId={selectedBuildingNode}
					onSelect={(node) => { if (node?.id) selectedBuildingNode = node.id; }}
					onNavigate={(node) => { if (node?.id) goto(`/nodes/${node.id}`); }}
					onCreateRoom={(floorLevel) => {
						// Pre-select parent based on floor level
						const flatNodes = flattenNodes($nodes.nodeTree);
						const nodesAtDepth = flatNodes.filter(n => n.depth === floorLevel);
						if (nodesAtDepth.length > 0) {
							// If there are existing nodes at this level, use first one's parent
							const firstNode = nodesAtDepth[0];
							newNodeParentId = firstNode.parent_id || null;
						} else if (floorLevel > 0) {
							// For new floor, find nodes at parent level
							const parentNodes = flatNodes.filter(n => n.depth === floorLevel - 1);
							if (parentNodes.length > 0) {
								newNodeParentId = parentNodes[0].id;
							}
						} else {
							newNodeParentId = null;
						}
						showNewNodeModal = true;
					}}
				/>
			</div>
		{:else if viewMode === 'building3d'}
			<!-- Building 3D View -->
			<div class="ng-canvas-wrap ng-canvas-wrap--full">
				<NodeBuilding3D
					nodes={$nodes.nodeTree}
					activeNodeId={$nodes.activeNode?.id}
					selectedId={selectedBuildingNode}
					onSelect={(node) => { if (node?.id) selectedBuildingNode = node.id; }}
					onNavigate={(node) => { if (node?.id) goto(`/nodes/${node.id}`); }}
					onCreateRoom={() => { showNewNodeModal = true; }}
				/>
			</div>
		{/if}
	</div>
</div>

<!-- New Node Modal -->
{#if showNewNodeModal}
	<div class="ng-modal-backdrop">
		<button
			class="ng-modal-backdrop__dismiss"
			onclick={() => showNewNodeModal = false}
		></button>

		<div
			class="ng-modal"
			transition:fly={{ y: 20, duration: 200 }}
		>
			<div class="ng-modal__header">
				<h2 class="ng-modal__title">Create New Node</h2>
				<button
					onclick={() => showNewNodeModal = false}
					class="ng-modal__close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="ng-modal__body">
				<div class="ng-input-group">
					<label class="ng-label">Node name <span class="ng-label__req">*</span></label>
					<input
						type="text"
						bind:value={newNodeName}
						placeholder="Enter node name"
						class="ng-input"
					/>
				</div>

				<div class="ng-input-group">
					<label class="ng-label">Type <span class="ng-label__req">*</span></label>
					<div class="ng-type-grid">
						{#each Object.entries(nodeTypeConfig) as [type, config]}
							<button
								onclick={() => newNodeType = type as NodeType}
								class="ng-type-option {newNodeType === type ? 'ng-type-option--selected' : ''}"
							>
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={config.icon} />
								</svg>
								<span class="ng-type-option__label">{config.label}</span>
							</button>
						{/each}
					</div>
				</div>

				<div class="ng-input-group">
					<label class="ng-label">Parent node (optional)</label>
					<select bind:value={newNodeParentId} class="ng-input">
						<option value={null}>No parent (root node)</option>
						{#each allNodes as node}
							<option value={node.id}>{node.name}</option>
						{/each}
					</select>
				</div>

				<div class="ng-input-group">
					<label class="ng-label">Purpose</label>
					<textarea
						bind:value={newNodePurpose}
						placeholder="Why does this node exist? What's its goal?"
						rows={3}
						class="ng-input ng-input--textarea"
					></textarea>
				</div>
			</div>

			<div class="ng-modal__footer">
				<button
					onclick={() => showNewNodeModal = false}
					class="btn-pill btn-pill-ghost"
				>
					Cancel
				</button>
				<button
					onclick={handleCreateNode}
					disabled={!newNodeName.trim() || isCreatingNode}
					class="btn-pill btn-pill-primary"
				>
					{isCreatingNode ? 'Creating...' : 'Create Node'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Click outside to close filter -->
{#if showFilterDropdown}
	<button
		class="ng-modal-backdrop__dismiss"
		onclick={() => showFilterDropdown = false}
	></button>
{/if}

<style>
	/* ── Page Shell ── */
	.ng-page { display: flex; flex-direction: column; height: 100%; background: var(--dbg); color: var(--dt); }

	/* ── Header ── */
	.ng-header { padding: 1.5rem 1.5rem 0; }
	.ng-header__row { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
	.ng-header__title { font-size: 1.5rem; font-weight: 700; color: var(--dt); }
	.ng-header__sub { font-size: 0.875rem; color: var(--dt3); margin-top: 0.25rem; }

	/* ── Toolbar ── */
	.ng-toolbar { padding: 1rem 1.5rem; }
	.ng-toolbar__row { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
	.ng-toolbar__right { display: flex; align-items: center; gap: 0.75rem; }

	/* ── View Tabs ── */
	.ng-view-tabs { display: flex; gap: 0.25rem; background: var(--dbg2); border-radius: 0.5rem; padding: 0.25rem; }
	.ng-view-tab {
		padding: 0.375rem 0.75rem; border-radius: 0.375rem; font-size: 0.8125rem; font-weight: 500;
		color: var(--dt3); background: transparent; border: none; cursor: pointer; transition: all 0.15s;
		display: inline-flex; align-items: center; gap: 0.375rem;
	}
	.ng-view-tab:hover { color: var(--dt); background: var(--dbg3); }
	.ng-view-tab--active { color: var(--dt); background: var(--dbg); box-shadow: 0 1px 3px rgba(0,0,0,.25); }

	/* ── Search ── */
	.ng-search-wrap { position: relative; display: flex; align-items: center; }
	.ng-search-wrap__icon { position: absolute; left: 0.625rem; width: 1rem; height: 1rem; color: var(--dt4); pointer-events: none; }
	.ng-search-input {
		padding: 0.375rem 0.75rem 0.375rem 2rem; font-size: 0.8125rem; border-radius: 0.5rem;
		border: 1px solid var(--dbd); background: var(--dbg2); color: var(--dt); width: 14rem; transition: border-color 0.15s;
	}
	.ng-search-input::placeholder { color: var(--dt4); }
	.ng-search-input:focus { outline: none; border-color: #3b82f6; }

	/* ── Filter ── */
	.ng-filter-wrap { position: relative; }
	.ng-filter-indicator {
		position: absolute; top: -0.25rem; right: -0.25rem; width: 0.5rem; height: 0.5rem;
		border-radius: 50%; background: #3b82f6;
	}
	.ng-filter-panel {
		position: absolute; right: 0; top: 100%; margin-top: 0.5rem; width: 18rem;
		background: var(--dbg2); border: 1px solid var(--dbd); border-radius: 0.75rem;
		box-shadow: 0 8px 24px rgba(0,0,0,.35); z-index: 30; padding: 1rem;
	}
	.ng-filter-panel__sections { display: flex; flex-direction: column; gap: 1rem; }
	.ng-filter-panel__label { font-size: 0.75rem; font-weight: 600; color: var(--dt3); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 0.375rem; }
	.ng-filter-panel__options { display: flex; flex-wrap: wrap; gap: 0.375rem; }
	.ng-filter-panel__option {
		padding: 0.25rem 0.625rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 500;
		border: 1px solid var(--dbd); background: transparent; color: var(--dt3); cursor: pointer; transition: all 0.15s; text-transform: capitalize;
	}
	.ng-filter-panel__option:hover { border-color: var(--dt3); color: var(--dt); }
	.ng-filter-panel__option--active { background: #3b82f6; border-color: #3b82f6; color: #fff; }
	.ng-filter-panel__footer { display: flex; justify-content: space-between; padding-top: 0.75rem; border-top: 1px solid var(--dbd); }
	.ng-filter-panel__clear { font-size: 0.75rem; color: var(--dt3); background: none; border: none; cursor: pointer; padding: 0; }
	.ng-filter-panel__clear:hover { color: var(--dt); }
	.ng-filter-panel__apply { font-size: 0.75rem; font-weight: 600; color: #3b82f6; background: none; border: none; cursor: pointer; padding: 0; }

	/* ── Active Banner ── */
	.ng-active-banner {
		display: flex; align-items: center; justify-content: space-between;
		margin: 0 1.5rem 0.75rem; padding: 0.75rem 1rem; border-radius: 0.5rem;
		background: rgba(59,130,246,.1); border: 1px solid rgba(59,130,246,.25);
	}
	.ng-active-banner__left { display: flex; align-items: center; gap: 0.5rem; }
	.ng-active-banner__right { display: flex; align-items: center; gap: 0.5rem; }
	.ng-active-banner__text { font-size: 0.875rem; color: var(--dt2); }

	/* ── Content Area ── */
	.ng-content { flex: 1; overflow-y: auto; padding: 0 1.5rem 1.5rem; }
	.ng-spinner { width: 2rem; height: 2rem; border: 3px solid var(--dbd); border-top-color: #3b82f6; border-radius: 50%; animation: ng-spin 0.8s linear infinite; margin: 3rem auto; display: block; }
	@keyframes ng-spin { to { transform: rotate(360deg); } }
	.ng-empty { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 4rem 1rem; text-align: center; }
	.ng-empty__icon-wrap { width: 3rem; height: 3rem; border-radius: 50%; background: var(--dbg2); display: flex; align-items: center; justify-content: center; margin-bottom: 1rem; color: var(--dt4); }
	.ng-empty__title { font-size: 1rem; font-weight: 600; color: var(--dt2); margin-bottom: 0.25rem; }
	.ng-empty__text { font-size: 0.875rem; color: var(--dt3); }

	/* ── Tree View ── */
	.ng-tree { display: flex; flex-direction: column; gap: 0.125rem; }
	.ng-tree-item {
		display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem 0.75rem;
		border-radius: 0.5rem; transition: background 0.15s; position: relative;
	}
	.ng-tree-item:hover { background: var(--dbg2); }
	.ng-tree-item__expand { width: 1.25rem; height: 1.25rem; display: flex; align-items: center; justify-content: center; color: var(--dt4); background: none; border: none; cursor: pointer; flex-shrink: 0; transition: transform 0.15s; }
	.ng-tree-item__chevron--open { transform: rotate(90deg); }
	.ng-tree-item__name { font-size: 0.875rem; font-weight: 500; color: var(--dt); text-decoration: none; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.ng-tree-item__name:hover { color: #3b82f6; }
	.ng-tree-item__actions { display: flex; gap: 0.25rem; opacity: 0; transition: opacity 0.15s; }
	.ng-tree-item:hover .ng-tree-item__actions { opacity: 1; }
	.ng-tree-item__action-btn { padding: 0.25rem; color: var(--dt4); background: none; border: none; cursor: pointer; border-radius: 0.25rem; transition: color 0.15s; }
	.ng-tree-item__action-btn:hover { color: #3b82f6; }
	.ng-tree-item__action-btn--danger:hover { color: #ef4444; }

	/* ── Type Icon ── */
	.ng-type-icon {
		width: 2rem; height: 2rem; border-radius: 0.5rem; display: flex; align-items: center; justify-content: center; flex-shrink: 0;
	}
	.ng-type-icon--lg { width: 2.5rem; height: 2.5rem; }
	.ng-type-icon--business { background: rgba(59,130,246,.15); color: #3b82f6; }
	.ng-type-icon--project { background: rgba(168,85,247,.15); color: #a855f7; }
	.ng-type-icon--learning { background: rgba(34,197,94,.15); color: #22c55e; }
	.ng-type-icon--operational { background: rgba(245,158,11,.15); color: #f59e0b; }

	/* ── Health Dot ── */
	.ng-health-dot { width: 0.5rem; height: 0.5rem; border-radius: 50%; flex-shrink: 0; display: inline-block; }
	.ng-health-dot--healthy { background: #22c55e; }
	.ng-health-dot--attention { background: #f59e0b; }
	.ng-health-dot--critical { background: #ef4444; }
	.ng-health-dot--not-started { background: #9ca3af; }

	/* ── Active Tag ── */
	.ng-active-tag {
		display: inline-block; padding: 0.125rem 0.5rem; font-size: 0.6875rem; font-weight: 600;
		background: rgba(59,130,246,.15); color: #3b82f6; border-radius: 9999px;
	}

	/* ── Table (List View) ── */
	.ng-table-wrap { border: 1px solid var(--dbd); border-radius: 0.75rem; overflow: hidden; }
	.ng-table { width: 100%; border-collapse: collapse; }
	.ng-table__head { background: var(--dbg2); border-bottom: 1px solid var(--dbd); }
	.ng-table__th { padding: 0.75rem 1rem; text-align: left; font-size: 0.6875rem; font-weight: 600; color: var(--dt4); text-transform: uppercase; letter-spacing: 0.05em; }
	.ng-table__row { border-bottom: 1px solid var(--dbd); transition: background 0.15s; }
	.ng-table__row:last-child { border-bottom: none; }
	.ng-table__row:hover { background: var(--dbg2); }
	.ng-table__td { padding: 0.75rem 1rem; font-size: 0.875rem; color: var(--dt); }
	.ng-table__td--muted { color: var(--dt3); }
	.ng-table__name-cell { display: flex; align-items: center; gap: 0.75rem; }
	.ng-table__health-cell { display: flex; align-items: center; gap: 0.5rem; }

	/* ── Grid View ── */
	.ng-grid { display: grid; grid-template-columns: repeat(1, 1fr); gap: 1rem; }
	@media (min-width: 640px) { .ng-grid { grid-template-columns: repeat(2, 1fr); } }
	@media (min-width: 1024px) { .ng-grid { grid-template-columns: repeat(3, 1fr); } }
	@media (min-width: 1280px) { .ng-grid { grid-template-columns: repeat(4, 1fr); } }

	.ng-card {
		display: block; padding: 1rem; background: var(--dbg2); border: 1px solid var(--dbd);
		border-radius: 0.75rem; text-decoration: none; color: inherit; transition: all 0.15s;
	}
	.ng-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,.2); transform: translateY(-2px); }
	.ng-card__top { display: flex; align-items: flex-start; gap: 0.75rem; }
	.ng-card__info { flex: 1; min-width: 0; }
	.ng-card__name { font-size: 0.875rem; font-weight: 600; color: var(--dt); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.ng-card__health { display: flex; align-items: center; gap: 0.375rem; margin-top: 0.375rem; font-size: 0.8125rem; color: var(--dt3); }
	.ng-card__active-icon { width: 1.25rem; height: 1.25rem; color: #3b82f6; flex-shrink: 0; }
	.ng-card__focus { margin-top: 0.75rem; padding-top: 0.75rem; border-top: 1px solid var(--dbd); }
	.ng-card__focus-label { font-size: 0.6875rem; font-weight: 600; color: var(--dt4); margin-bottom: 0.25rem; }
	.ng-card__focus-text { font-size: 0.8125rem; color: var(--dt2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.ng-card__children { margin-top: 0.5rem; font-size: 0.75rem; color: var(--dt4); }

	.ng-card--add {
		display: flex; flex-direction: column; align-items: center; justify-content: center;
		min-height: 8.75rem; border-style: dashed; border-width: 2px; cursor: pointer;
	}
	.ng-card--add:hover { border-color: var(--dt3); }
	.ng-card--add__icon { width: 2rem; height: 2rem; color: var(--dt4); margin-bottom: 0.5rem; }
	.ng-card--add__label { font-size: 0.8125rem; color: var(--dt3); }

	/* ── Modal ── */
	.ng-modal-backdrop { position: fixed; inset: 0; z-index: 50; display: flex; align-items: center; justify-content: center; }
	.ng-modal-backdrop__dismiss { position: fixed; inset: 0; z-index: 10; background: rgba(0,0,0,.5); border: none; cursor: default; }
	.ng-modal {
		position: relative; z-index: 20; background: var(--dbg2); border: 1px solid var(--dbd);
		border-radius: 1rem; box-shadow: 0 16px 48px rgba(0,0,0,.4); width: 100%; max-width: 32rem;
		margin: 0 1rem; overflow: hidden;
	}
	.ng-modal__header { padding: 1.25rem 1.5rem; border-bottom: 1px solid var(--dbd); display: flex; align-items: center; justify-content: space-between; }
	.ng-modal__title { font-size: 1.125rem; font-weight: 600; color: var(--dt); }
	.ng-modal__close { padding: 0.25rem; color: var(--dt4); background: none; border: none; cursor: pointer; border-radius: 0.25rem; transition: color 0.15s; }
	.ng-modal__close:hover { color: var(--dt); }
	.ng-modal__body { padding: 1.5rem; display: flex; flex-direction: column; gap: 1rem; }
	.ng-modal__footer { padding: 1rem 1.5rem; border-top: 1px solid var(--dbd); display: flex; justify-content: flex-end; gap: 0.75rem; }

	/* ── Form Controls ── */
	.ng-input-group { display: flex; flex-direction: column; gap: 0.25rem; }
	.ng-label { font-size: 0.8125rem; font-weight: 500; color: var(--dt2); }
	.ng-label__req { color: #ef4444; }
	.ng-input {
		width: 100%; padding: 0.5rem 0.75rem; font-size: 0.875rem; border-radius: 0.5rem;
		border: 1px solid var(--dbd); background: var(--dbg); color: var(--dt); transition: border-color 0.15s;
	}
	.ng-input:focus { outline: none; border-color: #3b82f6; }
	.ng-input--textarea { resize: none; }

	/* ── Type Selector Grid ── */
	.ng-type-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.5rem; }
	.ng-type-option {
		display: flex; flex-direction: column; align-items: center; gap: 0.375rem;
		padding: 0.75rem; border: 2px solid var(--dbd); border-radius: 0.5rem;
		background: transparent; color: var(--dt3); cursor: pointer; transition: all 0.15s;
	}
	.ng-type-option:hover { border-color: var(--dt3); }
	.ng-type-option--selected { border-color: #3b82f6; background: rgba(59,130,246,.1); color: #3b82f6; }
	.ng-type-option__label { font-size: 0.6875rem; font-weight: 500; }

	/* ── Utility ── */
	.ng-btn-icon { width: 1rem; height: 1rem; display: inline-block; margin-right: 0.25rem; vertical-align: middle; }
	.ng-view-tab__icon { width: 1rem; height: 1rem; display: inline-block; margin-right: 0.25rem; vertical-align: middle; }
	.ng-canvas-wrap { height: calc(100vh - 280px); min-height: 500px; }
	.ng-canvas-wrap--full { margin: 0 -1.5rem -1.5rem; }
</style>
