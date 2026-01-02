<script lang="ts">
	import { fly } from 'svelte/transition';
	import { nodes } from '$lib/stores/nodes';
	import { getProjects } from '$lib/api/projects';
	import { getContexts } from '$lib/api/contexts';
	import { getConversations } from '$lib/api/conversations';
	import type { Project } from '$lib/api/projects/types';
	import type { ContextListItem } from '$lib/api/contexts/types';
	import type { Conversation } from '$lib/api/conversations/types';
	import type { LinkedProject, LinkedContext, LinkedConversation } from '$lib/api/nodes/types';

	interface Props {
		nodeId: string;
		nodeName: string;
		onClose: () => void;
	}

	let { nodeId, nodeName, onClose }: Props = $props();

	// Tab state
	type TabType = 'projects' | 'contexts' | 'conversations';
	let activeTab: TabType = $state('projects');
	let searchQuery = $state('');

	// Data state
	let allProjects: Project[] = $state([]);
	let allContexts: ContextListItem[] = $state([]);
	let allConversations: Conversation[] = $state([]);

	// Loading state
	let loadingAvailable = $state(true);
	let linking = $state(false);

	// Derived: linked items from store
	let linkedProjects = $derived($nodes.currentNodeLinks?.projects ?? []);
	let linkedContexts = $derived($nodes.currentNodeLinks?.contexts ?? []);
	let linkedConversations = $derived($nodes.currentNodeLinks?.conversations ?? []);
	let linksLoading = $derived($nodes.linksLoading);

	// Derived: IDs of already linked items
	let linkedProjectIds = $derived(new Set(linkedProjects.map(p => p.id)));
	let linkedContextIds = $derived(new Set(linkedContexts.map(c => c.id)));
	let linkedConversationIds = $derived(new Set(linkedConversations.map(c => c.id)));

	// Derived: filtered available items (not already linked)
	let availableProjects = $derived(
		allProjects
			.filter(p => !linkedProjectIds.has(p.id))
			.filter(p => searchQuery === '' || p.name.toLowerCase().includes(searchQuery.toLowerCase()))
	);

	let availableContexts = $derived(
		allContexts
			.filter(c => !linkedContextIds.has(c.id))
			.filter(c => searchQuery === '' || c.name.toLowerCase().includes(searchQuery.toLowerCase()))
	);

	let availableConversations = $derived(
		allConversations
			.filter(c => !linkedConversationIds.has(c.id))
			.filter(c => {
				const title = c.title || 'Untitled';
				return searchQuery === '' || title.toLowerCase().includes(searchQuery.toLowerCase());
			})
	);

	// Load available items
	async function loadAvailableItems() {
		loadingAvailable = true;
		try {
			const [projects, contexts, conversations] = await Promise.all([
				getProjects(),
				getContexts(),
				getConversations()
			]);
			allProjects = projects;
			allContexts = contexts;
			allConversations = conversations;
		} catch (error) {
			console.error('Failed to load available items:', error);
		} finally {
			loadingAvailable = false;
		}
	}

	// Load linked items on mount
	$effect(() => {
		loadAvailableItems();
		nodes.loadLinks(nodeId);
	});

	// Link handlers
	async function handleLinkProject(projectId: string) {
		linking = true;
		try {
			await nodes.linkProject(nodeId, projectId);
		} catch (error) {
			console.error('Failed to link project:', error);
		} finally {
			linking = false;
		}
	}

	async function handleUnlinkProject(projectId: string) {
		linking = true;
		try {
			await nodes.unlinkProject(nodeId, projectId);
		} catch (error) {
			console.error('Failed to unlink project:', error);
		} finally {
			linking = false;
		}
	}

	async function handleLinkContext(contextId: string) {
		linking = true;
		try {
			await nodes.linkContext(nodeId, contextId);
		} catch (error) {
			console.error('Failed to link context:', error);
		} finally {
			linking = false;
		}
	}

	async function handleUnlinkContext(contextId: string) {
		linking = true;
		try {
			await nodes.unlinkContext(nodeId, contextId);
		} catch (error) {
			console.error('Failed to unlink context:', error);
		} finally {
			linking = false;
		}
	}

	async function handleLinkConversation(conversationId: string) {
		linking = true;
		try {
			await nodes.linkConversation(nodeId, conversationId);
		} catch (error) {
			console.error('Failed to link conversation:', error);
		} finally {
			linking = false;
		}
	}

	async function handleUnlinkConversation(conversationId: string) {
		linking = true;
		try {
			await nodes.unlinkConversation(nodeId, conversationId);
		} catch (error) {
			console.error('Failed to unlink conversation:', error);
		} finally {
			linking = false;
		}
	}

	// Helper for status colors
	function getStatusColor(status: string) {
		switch (status) {
			case 'active': return 'bg-green-100 text-green-700';
			case 'completed': return 'bg-blue-100 text-blue-700';
			case 'paused': return 'bg-yellow-100 text-yellow-700';
			case 'archived': return 'bg-gray-100 text-gray-700';
			default: return 'bg-gray-100 text-gray-700';
		}
	}

	function getContextTypeIcon(type: string) {
		switch (type) {
			case 'person': return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
			case 'business': return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
			case 'project': return 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z';
			case 'document': return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			default: return 'M4 6h16M4 12h16M4 18h16';
		}
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center">
	<!-- Backdrop -->
	<button
		class="absolute inset-0 bg-black/50"
		onclick={onClose}
	></button>

	<!-- Modal -->
	<div
		class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl mx-4 max-h-[85vh] flex flex-col overflow-hidden"
		transition:fly={{ y: 20, duration: 200 }}
	>
		<!-- Header -->
		<div class="p-6 border-b border-gray-200 flex-shrink-0">
			<div class="flex items-center justify-between">
				<div>
					<h2 class="text-xl font-semibold text-gray-900">Link Items</h2>
					<p class="text-sm text-gray-500 mt-1">Connect projects, context profiles, and conversations to <span class="font-medium">{nodeName}</span></p>
				</div>
				<button
					onclick={onClose}
					class="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Tabs -->
			<div class="flex gap-1 mt-4 border-b border-gray-200 -mb-px">
				{#each [
					{ id: 'projects', label: 'Projects', count: linkedProjects.length },
					{ id: 'contexts', label: 'Context Profiles', count: linkedContexts.length },
					{ id: 'conversations', label: 'Conversations', count: linkedConversations.length }
				] as tab}
					<button
						onclick={() => { activeTab = tab.id as TabType; searchQuery = ''; }}
						class="px-4 py-2 text-sm font-medium border-b-2 transition-colors {activeTab === tab.id
							? 'border-blue-500 text-blue-600'
							: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
					>
						{tab.label}
						{#if tab.count > 0}
							<span class="ml-1.5 px-1.5 py-0.5 text-xs rounded-full bg-blue-100 text-blue-700">{tab.count}</span>
						{/if}
					</button>
				{/each}
			</div>
		</div>

		<!-- Search -->
		<div class="px-6 py-3 border-b border-gray-100 flex-shrink-0">
			<div class="relative">
				<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search {activeTab}..."
					class="w-full pl-10 pr-4 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
			</div>
		</div>

		<!-- Content -->
		<div class="flex-1 overflow-y-auto p-6">
			{#if loadingAvailable || linksLoading}
				<div class="flex items-center justify-center py-12">
					<div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
				</div>
			{:else}
				<!-- Linked Items Section -->
				{#if activeTab === 'projects' && linkedProjects.length > 0}
					<div class="mb-6">
						<h3 class="text-sm font-medium text-gray-700 mb-3">Linked Projects ({linkedProjects.length})</h3>
						<div class="space-y-2">
							{#each linkedProjects as project}
								<div class="flex items-center justify-between p-3 bg-blue-50 rounded-lg border border-blue-100">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-lg bg-green-100 text-green-600 flex items-center justify-center">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
											</svg>
										</div>
										<div>
											<p class="font-medium text-gray-900">{project.name}</p>
											<p class="text-xs text-gray-500">Linked {formatDate(project.linked_at)}</p>
										</div>
									</div>
									<button
										onclick={() => handleUnlinkProject(project.id)}
										disabled={linking}
										class="p-1.5 text-red-500 hover:bg-red-100 rounded-lg transition-colors disabled:opacity-50"
										title="Unlink"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
										</svg>
									</button>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				{#if activeTab === 'contexts' && linkedContexts.length > 0}
					<div class="mb-6">
						<h3 class="text-sm font-medium text-gray-700 mb-3">Linked Context Profiles ({linkedContexts.length})</h3>
						<div class="space-y-2">
							{#each linkedContexts as context}
								<div class="flex items-center justify-between p-3 bg-blue-50 rounded-lg border border-blue-100">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-lg bg-purple-100 text-purple-600 flex items-center justify-center">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getContextTypeIcon(context.type)} />
											</svg>
										</div>
										<div>
											<p class="font-medium text-gray-900">{context.name}</p>
											<p class="text-xs text-gray-500">Linked {formatDate(context.linked_at)}</p>
										</div>
									</div>
									<button
										onclick={() => handleUnlinkContext(context.id)}
										disabled={linking}
										class="p-1.5 text-red-500 hover:bg-red-100 rounded-lg transition-colors disabled:opacity-50"
										title="Unlink"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
										</svg>
									</button>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				{#if activeTab === 'conversations' && linkedConversations.length > 0}
					<div class="mb-6">
						<h3 class="text-sm font-medium text-gray-700 mb-3">Linked Conversations ({linkedConversations.length})</h3>
						<div class="space-y-2">
							{#each linkedConversations as conversation}
								<div class="flex items-center justify-between p-3 bg-blue-50 rounded-lg border border-blue-100">
									<div class="flex items-center gap-3">
										<div class="w-8 h-8 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
											</svg>
										</div>
										<div>
											<p class="font-medium text-gray-900">{conversation.title || 'Untitled Conversation'}</p>
											<p class="text-xs text-gray-500">Linked {formatDate(conversation.linked_at)}</p>
										</div>
									</div>
									<button
										onclick={() => handleUnlinkConversation(conversation.id)}
										disabled={linking}
										class="p-1.5 text-red-500 hover:bg-red-100 rounded-lg transition-colors disabled:opacity-50"
										title="Unlink"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
										</svg>
									</button>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Available Items Section -->
				{#if activeTab === 'projects'}
					<div>
						<h3 class="text-sm font-medium text-gray-700 mb-3">Available Projects ({availableProjects.length})</h3>
						{#if availableProjects.length === 0}
							<p class="text-sm text-gray-500 text-center py-8">
								{searchQuery ? 'No matching projects found' : 'All projects are already linked'}
							</p>
						{:else}
							<div class="space-y-2">
								{#each availableProjects as project}
									<button
										onclick={() => handleLinkProject(project.id)}
										disabled={linking}
										class="w-full flex items-center justify-between p-3 bg-white rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors disabled:opacity-50"
									>
										<div class="flex items-center gap-3">
											<div class="w-8 h-8 rounded-lg bg-green-100 text-green-600 flex items-center justify-center">
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
												</svg>
											</div>
											<div class="text-left">
												<p class="font-medium text-gray-900">{project.name}</p>
												<div class="flex items-center gap-2 mt-0.5">
													<span class="px-1.5 py-0.5 text-xs rounded {getStatusColor(project.status)}">{project.status}</span>
													{#if project.client_name}
														<span class="text-xs text-gray-500">{project.client_name}</span>
													{/if}
												</div>
											</div>
										</div>
										<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
									</button>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

				{#if activeTab === 'contexts'}
					<div>
						<h3 class="text-sm font-medium text-gray-700 mb-3">Available Context Profiles ({availableContexts.length})</h3>
						{#if availableContexts.length === 0}
							<p class="text-sm text-gray-500 text-center py-8">
								{searchQuery ? 'No matching context profiles found' : 'All context profiles are already linked'}
							</p>
						{:else}
							<div class="space-y-2">
								{#each availableContexts as context}
									<button
										onclick={() => handleLinkContext(context.id)}
										disabled={linking}
										class="w-full flex items-center justify-between p-3 bg-white rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors disabled:opacity-50"
									>
										<div class="flex items-center gap-3">
											<div class="w-8 h-8 rounded-lg bg-purple-100 text-purple-600 flex items-center justify-center">
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getContextTypeIcon(context.type)} />
												</svg>
											</div>
											<div class="text-left">
												<p class="font-medium text-gray-900">{context.name}</p>
												<p class="text-xs text-gray-500 capitalize">{context.type}</p>
											</div>
										</div>
										<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
									</button>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

				{#if activeTab === 'conversations'}
					<div>
						<h3 class="text-sm font-medium text-gray-700 mb-3">Available Conversations ({availableConversations.length})</h3>
						{#if availableConversations.length === 0}
							<p class="text-sm text-gray-500 text-center py-8">
								{searchQuery ? 'No matching conversations found' : 'All conversations are already linked'}
							</p>
						{:else}
							<div class="space-y-2">
								{#each availableConversations as conversation}
									<button
										onclick={() => handleLinkConversation(conversation.id)}
										disabled={linking}
										class="w-full flex items-center justify-between p-3 bg-white rounded-lg border border-gray-200 hover:border-blue-300 hover:bg-blue-50 transition-colors disabled:opacity-50"
									>
										<div class="flex items-center gap-3">
											<div class="w-8 h-8 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center">
												<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
												</svg>
											</div>
											<div class="text-left">
												<p class="font-medium text-gray-900">{conversation.title || 'Untitled Conversation'}</p>
												<p class="text-xs text-gray-500">{formatDate(conversation.updated_at)}</p>
											</div>
										</div>
										<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
										</svg>
									</button>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			{/if}
		</div>

		<!-- Footer -->
		<div class="px-6 py-4 border-t border-gray-200 flex-shrink-0 bg-gray-50">
			<div class="flex items-center justify-between">
				<p class="text-sm text-gray-500">
					{linkedProjects.length + linkedContexts.length + linkedConversations.length} items linked
				</p>
				<button
					onclick={onClose}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
				>
					Done
				</button>
			</div>
		</div>
	</div>
</div>
