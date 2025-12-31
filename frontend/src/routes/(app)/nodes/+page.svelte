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

	// Node type config with full Tailwind classes (static for purging)
	const nodeTypeConfig: Record<string, { icon: string; bgClass: string; textClass: string; label: string }> = {
		business: {
			icon: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4',
			bgClass: 'bg-blue-100',
			textClass: 'text-blue-600',
			label: 'Business'
		},
		project: {
			icon: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z',
			bgClass: 'bg-green-100',
			textClass: 'text-green-600',
			label: 'Project'
		},
		learning: {
			icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253',
			bgClass: 'bg-purple-100',
			textClass: 'text-purple-600',
			label: 'Learning'
		},
		operational: {
			icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
			bgClass: 'bg-orange-100',
			textClass: 'text-orange-600',
			label: 'Operational'
		},
	};
	const defaultTypeConfig = { icon: 'M4 6h16M4 12h16M4 18h16', bgClass: 'bg-gray-100', textClass: 'text-gray-600', label: 'Unknown' };

	// Health config
	const healthConfig: Record<string, { color: string; label: string }> = {
		healthy: { color: 'bg-green-500', label: 'Healthy' },
		needs_attention: { color: 'bg-yellow-500', label: 'Needs Attention' },
		critical: { color: 'bg-red-500', label: 'Critical' },
		not_started: { color: 'bg-gray-400', label: 'Not Started' },
	};
	const defaultHealthConfig = { color: 'bg-gray-400', label: 'Unknown' };

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

<div class="h-full flex flex-col bg-white">
	<!-- Header -->
	<div class="border-b border-gray-200 px-6 py-4 flex-shrink-0">
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Business Nodes</h1>
				<p class="text-sm text-gray-500 mt-1">Your cognitive operating system structure</p>
			</div>
			<button
				onclick={() => showNewNodeModal = true}
				class="flex items-center gap-2 px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				New Node
			</button>
		</div>
	</div>

	<!-- Toolbar -->
	<div class="border-b border-gray-200 px-6 py-3 flex-shrink-0">
		<div class="flex items-center justify-between gap-4">
			<!-- View Switcher -->
			<div class="flex items-center gap-1 bg-gray-100 rounded-lg p-1">
				<button
					onclick={() => viewMode = 'tree'}
					class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {viewMode === 'tree' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
				>
					<svg class="w-4 h-4 inline-block mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
					</svg>
					Tree
				</button>
				<button
					onclick={() => viewMode = 'list'}
					class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {viewMode === 'list' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
				>
					<svg class="w-4 h-4 inline-block mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
					List
				</button>
				<button
					onclick={() => viewMode = 'grid'}
					class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {viewMode === 'grid' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
				>
					<svg class="w-4 h-4 inline-block mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
					</svg>
					Grid
				</button>
				<button
					onclick={() => viewMode = 'graph'}
					class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {viewMode === 'graph' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
				>
					<svg class="w-4 h-4 inline-block mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
					</svg>
					Graph
				</button>
				<button
					onclick={() => viewMode = 'building'}
					class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {viewMode === 'building' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
				>
					<svg class="w-4 h-4 inline-block mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
					</svg>
					2D
				</button>
				<button
					onclick={() => viewMode = 'building3d'}
					class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {viewMode === 'building3d' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'}"
				>
					<svg class="w-4 h-4 inline-block mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
					</svg>
					3D
				</button>
			</div>

			<div class="flex items-center gap-3">
				<!-- Filter Dropdown -->
				<div class="relative">
					<button
						onclick={() => showFilterDropdown = !showFilterDropdown}
						class="flex items-center gap-2 px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
						</svg>
						Filter
						{#if filterType !== 'all' || filterHealth !== 'all' || showArchived}
							<span class="w-2 h-2 bg-blue-500 rounded-full"></span>
						{/if}
					</button>

					{#if showFilterDropdown}
						<div
							class="absolute right-0 top-full mt-2 w-64 bg-white border border-gray-200 rounded-xl shadow-lg p-4 z-20"
							transition:fly={{ y: -10, duration: 200 }}
						>
							<div class="space-y-4">
								<div>
									<label class="text-xs font-semibold text-gray-500 uppercase">Type</label>
									<div class="mt-2 space-y-1">
										<label class="flex items-center gap-2">
											<input type="radio" bind:group={filterType} value="all" class="text-blue-600" />
											<span class="text-sm">All Types</span>
										</label>
										{#each Object.entries(nodeTypeConfig) as [type, config]}
											<label class="flex items-center gap-2">
												<input type="radio" bind:group={filterType} value={type} class="text-blue-600" />
												<span class="text-sm">{config.label}</span>
											</label>
										{/each}
									</div>
								</div>

								<div>
									<label class="text-xs font-semibold text-gray-500 uppercase">Health</label>
									<div class="mt-2 space-y-1">
										<label class="flex items-center gap-2">
											<input type="radio" bind:group={filterHealth} value="all" class="text-blue-600" />
											<span class="text-sm">All Health</span>
										</label>
										{#each Object.entries(healthConfig) as [health, config]}
											<label class="flex items-center gap-2">
												<input type="radio" bind:group={filterHealth} value={health} class="text-blue-600" />
												<span class="w-2 h-2 rounded-full {config.color}"></span>
												<span class="text-sm">{config.label}</span>
											</label>
										{/each}
									</div>
								</div>

								<div>
									<label class="flex items-center gap-2">
										<input type="checkbox" bind:checked={showArchived} onchange={() => loadData()} class="text-blue-600 rounded" />
										<span class="text-sm">Show Archived</span>
									</label>
								</div>

								<div class="pt-3 border-t border-gray-100 flex justify-between">
									<button
										onclick={() => { filterType = 'all'; filterHealth = 'all'; showArchived = false; loadData(); }}
										class="text-sm text-gray-500 hover:text-gray-700"
									>
										Clear
									</button>
									<button
										onclick={() => showFilterDropdown = false}
										class="text-sm text-blue-600 font-medium hover:text-blue-700"
									>
										Apply
									</button>
								</div>
							</div>
						</div>
					{/if}
				</div>

				<!-- Search -->
				<div class="relative">
					<svg class="w-5 h-5 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
					<input
						type="text"
						placeholder="Search nodes..."
						bind:value={searchInput}
						oninput={(e) => updateSearchQuery(e.currentTarget.value)}
						class="pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent w-64"
					/>
				</div>
			</div>
		</div>
	</div>

	<!-- Active Node Banner -->
	{#if $nodes.activeNode}
		<div class="bg-blue-50 border-b border-blue-100 px-6 py-3 flex items-center justify-between flex-shrink-0" transition:slide>
			<div class="flex items-center gap-3">
				<svg class="w-5 h-5 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
					<path d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
				<span class="text-sm font-medium text-blue-900">
					Active Node: <span class="font-semibold">{$nodes.activeNode.name}</span>
				</span>
			</div>
			<div class="flex items-center gap-2">
				<a
					href="/nodes/{$nodes.activeNode.id}"
					class="px-3 py-1 text-sm text-blue-700 hover:bg-blue-100 rounded-lg transition-colors"
				>
					View
				</a>
				<button
					onclick={handleDeactivate}
					class="px-3 py-1 text-sm text-blue-700 hover:bg-blue-100 rounded-lg transition-colors"
				>
					Deactivate
				</button>
			</div>
		</div>
	{/if}

	<!-- Content -->
	<div class="flex-1 overflow-auto p-6">
		{#if $nodes.loading}
			<div class="flex items-center justify-center h-64">
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
			</div>
		{:else if error}
			<div class="flex flex-col items-center justify-center h-64">
				<p class="text-red-500 mb-4">{error}</p>
				<button onclick={loadData} class="px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800">
					Retry
				</button>
			</div>
		{:else if filteredNodes.length === 0}
			<div class="flex flex-col items-center justify-center h-64 text-center">
				<div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
					</svg>
				</div>
				<h3 class="text-lg font-medium text-gray-900 mb-1">No nodes yet</h3>
				<p class="text-gray-500 mb-4">Create your first node to organize your business into manageable focus areas.</p>
				<button
					onclick={() => showNewNodeModal = true}
					class="px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800"
				>
					Create your first node
				</button>
			</div>
		{:else if viewMode === 'tree'}
			<!-- Tree View -->
			<div class="space-y-1">
				{#snippet treeNode(node: NodeTree, depth: number = 0)}
					<div>
						<div
							class="flex items-center gap-2 py-2 px-3 rounded-lg hover:bg-gray-50 transition-colors group"
							style="padding-left: {depth * 24 + 12}px"
						>
							<!-- Expand/Collapse -->
							{#if node.children.length > 0}
								<button
									onclick={() => toggleExpand(node.id)}
									class="p-0.5 text-gray-400 hover:text-gray-600"
								>
									<svg
										class="w-4 h-4 transition-transform {expandedNodes.has(node.id) ? 'rotate-90' : ''}"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</button>
							{:else}
								<div class="w-5"></div>
							{/if}

							<!-- Type Icon -->
							<div class="w-8 h-8 rounded-lg {getTypeConfig(node.type).bgClass} {getTypeConfig(node.type).textClass} flex items-center justify-center flex-shrink-0">
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(node.type).icon} />
								</svg>
							</div>

							<!-- Name -->
							<a href="/nodes/{node.id}" class="flex-1 font-medium text-gray-900 hover:text-blue-600">
								{node.name}
							</a>

							<!-- Active indicator -->
							{#if node.is_active}
								<span class="px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded">Active</span>
							{/if}

							<!-- Health -->
							<div class="w-2.5 h-2.5 rounded-full {getHealthConfig(node.health).color}"></div>

							<!-- Actions -->
							<div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
								{#if !node.is_active}
									<button
										onclick={() => handleActivate(node.id)}
										class="p-1 text-gray-400 hover:text-blue-600"
										title="Activate"
									>
										<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
											<path d="M13 10V3L4 14h7v7l9-11h-7z" />
										</svg>
									</button>
								{/if}
								<button
									onclick={() => handleDelete(node.id)}
									class="p-1 text-gray-400 hover:text-red-600"
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
			<div class="border border-gray-200 rounded-lg overflow-hidden">
				<table class="w-full">
					<thead class="bg-gray-50 border-b border-gray-200">
						<tr>
							<th class="px-4 py-3 text-left text-xs font-semibold text-gray-500 uppercase">Name</th>
							<th class="px-4 py-3 text-left text-xs font-semibold text-gray-500 uppercase">Type</th>
							<th class="px-4 py-3 text-left text-xs font-semibold text-gray-500 uppercase">Health</th>
							<th class="px-4 py-3 text-left text-xs font-semibold text-gray-500 uppercase">Updated</th>
							<th class="px-4 py-3 text-right text-xs font-semibold text-gray-500 uppercase">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200">
						{#each flatNodes as node}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3">
									<div class="flex items-center gap-3" style="padding-left: {node.depth * 20}px">
										<div class="w-8 h-8 rounded-lg {getTypeConfig(node.type).bgClass} {getTypeConfig(node.type).textClass} flex items-center justify-center flex-shrink-0">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(node.type).icon} />
											</svg>
										</div>
										<a href="/nodes/{node.id}" class="font-medium text-gray-900 hover:text-blue-600">
											{node.name}
										</a>
										{#if node.is_active}
											<span class="px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded">Active</span>
										{/if}
									</div>
								</td>
								<td class="px-4 py-3 text-sm text-gray-600 capitalize">{node.type}</td>
								<td class="px-4 py-3">
									<span class="flex items-center gap-2">
										<span class="w-2 h-2 rounded-full {getHealthConfig(node.health).color}"></span>
										<span class="text-sm text-gray-600">{getHealthConfig(node.health).label}</span>
									</span>
								</td>
								<td class="px-4 py-3 text-sm text-gray-500">
									{new Date(node.updated_at).toLocaleDateString()}
								</td>
								<td class="px-4 py-3 text-right">
									<div class="flex items-center justify-end gap-1">
										{#if !node.is_active}
											<button
												onclick={() => handleActivate(node.id)}
												class="p-1 text-gray-400 hover:text-blue-600"
											>
												<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
													<path d="M13 10V3L4 14h7v7l9-11h-7z" />
												</svg>
											</button>
										{/if}
										<button
											onclick={() => handleDelete(node.id)}
											class="p-1 text-gray-400 hover:text-red-600"
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
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
				{#each flatNodes as node}
					<a
						href="/nodes/{node.id}"
						class="block p-4 bg-white border border-gray-200 rounded-xl hover:shadow-md transition-all hover:-translate-y-0.5"
					>
						<div class="flex items-start gap-3">
							<div class="w-10 h-10 rounded-lg {getTypeConfig(node.type).bgClass} {getTypeConfig(node.type).textClass} flex items-center justify-center flex-shrink-0">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeConfig(node.type).icon} />
								</svg>
							</div>
							<div class="flex-1 min-w-0">
								<h3 class="font-medium text-gray-900 truncate">{node.name}</h3>
								<div class="flex items-center gap-2 mt-1">
									<span class="w-2 h-2 rounded-full {getHealthConfig(node.health).color}"></span>
									<span class="text-sm text-gray-500">{getHealthConfig(node.health).label}</span>
								</div>
							</div>
							{#if node.is_active}
								<svg class="w-5 h-5 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
									<path d="M13 10V3L4 14h7v7l9-11h-7z" />
								</svg>
							{/if}
						</div>

						{#if node.this_week_focus && node.this_week_focus.length > 0}
							<div class="mt-3 pt-3 border-t border-gray-100">
								<p class="text-xs font-medium text-gray-500 mb-1">This week:</p>
								<p class="text-sm text-gray-600 truncate">{node.this_week_focus[0]}</p>
							</div>
						{/if}

						{#if node.children_count > 0}
							<p class="mt-2 text-xs text-gray-400">{node.children_count} child nodes</p>
						{/if}
					</a>
				{/each}

				<!-- Add New Card -->
				<button
					onclick={() => showNewNodeModal = true}
					class="flex flex-col items-center justify-center p-6 border-2 border-dashed border-gray-200 rounded-xl hover:border-gray-400 transition-colors min-h-[140px]"
				>
					<svg class="w-8 h-8 text-gray-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					<span class="text-sm text-gray-500">Add New Node</span>
				</button>
			</div>
		{:else if viewMode === 'graph'}
			<!-- Graph View -->
			<div class="h-[calc(100vh-280px)] min-h-[500px]">
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
			<div class="h-[calc(100vh-280px)] min-h-[500px] -mx-6 -mb-6">
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
			<div class="h-[calc(100vh-280px)] min-h-[500px] -mx-6 -mb-6">
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
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<button
			class="absolute inset-0 bg-black/50"
			onclick={() => showNewNodeModal = false}
		></button>

		<div
			class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg mx-4 overflow-hidden"
			transition:fly={{ y: 20, duration: 200 }}
		>
			<div class="p-6 border-b border-gray-200">
				<div class="flex items-center justify-between">
					<h2 class="text-xl font-semibold text-gray-900">Create New Node</h2>
					<button
						onclick={() => showNewNodeModal = false}
						class="p-1 text-gray-400 hover:text-gray-600"
					>
						<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			</div>

			<div class="p-6 space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Node name *</label>
					<input
						type="text"
						bind:value={newNodeName}
						placeholder="Enter node name"
						class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">Type *</label>
					<div class="grid grid-cols-4 gap-2">
						{#each Object.entries(nodeTypeConfig) as [type, config]}
							<button
								onclick={() => newNodeType = type as NodeType}
								class="flex flex-col items-center gap-1 p-3 border-2 rounded-lg transition-colors {newNodeType === type ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
							>
								<svg class="w-5 h-5 {config.textClass}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={config.icon} />
								</svg>
								<span class="text-xs font-medium text-gray-700">{config.label}</span>
							</button>
						{/each}
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Parent node (optional)</label>
					<select
						bind:value={newNodeParentId}
						class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value={null}>No parent (root node)</option>
						{#each allNodes as node}
							<option value={node.id}>{node.name}</option>
						{/each}
					</select>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Purpose</label>
					<textarea
						bind:value={newNodePurpose}
						placeholder="Why does this node exist? What's its goal?"
						rows={3}
						class="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
					></textarea>
				</div>
			</div>

			<div class="p-6 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={() => showNewNodeModal = false}
					class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={handleCreateNode}
					disabled={!newNodeName.trim() || isCreatingNode}
					class="px-4 py-2 bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
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
		class="fixed inset-0 z-10 cursor-default"
		onclick={() => showFilterDropdown = false}
	></button>
{/if}
