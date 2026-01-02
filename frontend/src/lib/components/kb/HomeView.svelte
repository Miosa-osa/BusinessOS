<script lang="ts">
	import { onMount } from 'svelte';
	import type { ContextListItem, ArtifactListItem } from '$lib/api/client';
	import { api } from '$lib/api/client';
	import TreeSearchPanel from '$lib/components/contexts/TreeSearchPanel.svelte';
	import type { TreeSearchResult } from '$lib/api/context-tree/types';

	interface Props {
		pages: ContextListItem[];
		recentPages: ContextListItem[];
		memories: any[];
		onSelectPage: (page: ContextListItem) => void;
		onSelectMemory: (memory: any) => void;
		onCreatePage: () => void;
	}

	let { pages, recentPages, memories, onSelectPage, onSelectMemory, onCreatePage }: Props = $props();

	let searchQuery = $state('');
	let viewMode = $state<'grid' | 'list'>('list');
	let showTreeSearch = $state(false);

	function handleTreeSearchItemSelect(item: TreeSearchResult) {
		console.log('Selected tree item:', item);
		showTreeSearch = false;
		// TODO: Navigate to the selected item based on entity_type
	}

	// Recent pages (last 8)
	const displayRecentPages = $derived(recentPages.slice(0, 8));

	// Recent memories (last 4)
	const displayRecentMemories = $derived(memories.slice(0, 4));

	// Filtered and sorted pages
	const filteredPages = $derived.by(() => {
		let filtered = [...pages];

		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			filtered = filtered.filter(p =>
				p.name?.toLowerCase().includes(query) ||
				p.type?.toLowerCase().includes(query)
			);
		}

		return filtered.sort((a, b) => {
			const aDate = a.updated_at ? new Date(a.updated_at).getTime() : 0;
			const bDate = b.updated_at ? new Date(b.updated_at).getTime() : 0;
			return bDate - aDate;
		});
	});

	function formatDate(dateStr: string | undefined): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));

		if (days === 0) return 'Today';
		if (days === 1) return 'Yesterday';
		if (days < 7) return `${days}d ago`;
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	// Icon presets for SVG rendering
	const iconPresets = [
		{ id: 'document', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		{ id: 'folder', path: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
		{ id: 'briefcase', path: 'M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ id: 'user', path: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
		{ id: 'users', path: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
		{ id: 'star', path: 'M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z' },
		{ id: 'bookmark', path: 'M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z' },
		{ id: 'home', path: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
		{ id: 'chat', path: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' },
		{ id: 'calendar', path: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'chart-bar', path: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' },
		{ id: 'lightbulb', path: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z' },
	];

	function getIconPath(iconId: string | null): string | null {
		if (!iconId) return null;
		const found = iconPresets.find(i => i.id === iconId);
		return found?.path || null;
	}
</script>

<div class="flex-1 overflow-auto bg-white dark:bg-[#191919]">
	<div class="max-w-4xl mx-auto px-8 py-16">

		<!-- Header -->
		<div class="mb-12">
			<h1 class="text-4xl font-bold text-gray-900 dark:text-white mb-2">Knowledge Base</h1>
			<p class="text-gray-500 dark:text-gray-400">Your workspace for pages, docs, and artifacts</p>
		</div>

		<!-- Quick Actions -->
		<div class="mb-12 flex gap-3">
			<button
				onclick={() => onCreatePage()}
				class="inline-flex items-center gap-2 px-4 py-2 text-gray-600 dark:text-gray-400 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white transition-colors text-sm"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
				</svg>
				New page
			</button>

			<button
				onclick={() => showTreeSearch = true}
				class="inline-flex items-center gap-2 px-4 py-2 text-gray-600 dark:text-gray-400 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white transition-colors text-sm"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
				Tree Search
			</button>
		</div>

		<!-- Recently Viewed -->
		{#if displayRecentPages.length > 0}
			<div class="mb-12">
				<h2 class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-4">Recently viewed</h2>
				<div class="space-y-0.5">
					{#each displayRecentPages as page (page.id)}
						<button
							onclick={() => onSelectPage(page)}
							class="group w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800/50 transition-colors text-left"
						>
							<div class="w-6 h-6 rounded flex items-center justify-center flex-shrink-0 bg-gray-100 dark:bg-gray-800">
								{#if page.icon && getIconPath(page.icon)}
									<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getIconPath(page.icon)} />
									</svg>
								{:else}
									<svg class="w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								{/if}
							</div>
							<span class="flex-1 text-sm text-gray-700 dark:text-gray-300 truncate">{page.name || 'New page'}</span>
							<span class="text-xs text-gray-400 dark:text-gray-500">{formatDate(page.updated_at)}</span>
						</button>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Recent Memories -->
		{#if displayRecentMemories.length > 0}
			<div class="mb-12">
				<h2 class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-4">Learned Memories</h2>
				<div class="grid grid-cols-2 gap-3">
					{#each displayRecentMemories as memory (memory.id)}
						<button
							onclick={() => onSelectMemory(memory)}
							class="group flex flex-col p-4 bg-pink-50/50 dark:bg-pink-900/10 rounded-xl border border-pink-100 dark:border-pink-900/30 hover:border-pink-200 dark:hover:border-pink-800 transition-all text-left"
						>
							<div class="w-8 h-8 rounded-lg bg-pink-100 dark:bg-pink-900/30 flex items-center justify-center mb-3">
								<span class="text-lg">🧠</span>
							</div>
							<div class="text-sm font-medium text-gray-900 dark:text-white line-clamp-2 mb-1">
								{memory.learning_summary || memory.learning_content}
							</div>
							<div class="text-xs text-gray-400 dark:text-gray-500 mt-auto">
								Learned {formatDate(memory.created_at)}
							</div>
						</button>
					{/each}
				</div>
			</div>
		{/if}

		<!-- All Pages -->
		<div>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider">
					All pages
					<span class="ml-1 text-gray-300 dark:text-gray-600">({filteredPages.length})</span>
				</h2>
				<div class="flex items-center gap-2">
					<!-- View Toggle -->
					<div class="flex items-center border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
						<button
							onclick={() => viewMode = 'list'}
							class="p-1.5 transition-colors {viewMode === 'list' ? 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300' : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
							</svg>
						</button>
						<button
							onclick={() => viewMode = 'grid'}
							class="p-1.5 transition-colors {viewMode === 'grid' ? 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300' : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
							</svg>
						</button>
					</div>
					<!-- Search -->
					<div class="relative">
						<svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
						</svg>
						<input
							type="text"
							bind:value={searchQuery}
							placeholder="Search..."
							class="pl-8 pr-3 py-1.5 w-48 bg-transparent border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-900 dark:text-white placeholder:text-gray-400 focus:outline-none focus:border-gray-400 dark:focus:border-gray-500 transition-colors"
						/>
					</div>
				</div>
			</div>

			{#if filteredPages.length === 0}
				<div class="text-center py-16">
					<div class="w-12 h-12 mx-auto mb-4 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center">
						<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
					</div>
					<p class="text-gray-500 dark:text-gray-400 mb-4">
						{#if searchQuery}
							No pages found
						{:else}
							No pages yet
						{/if}
					</p>
					{#if !searchQuery}
						<button
							onclick={() => onCreatePage()}
							class="text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white underline transition-colors"
						>
							Create your first page
						</button>
					{/if}
				</div>
			{:else if viewMode === 'grid'}
				<div class="grid grid-cols-3 gap-3">
					{#each filteredPages as page (page.id)}
						<button
							onclick={() => onSelectPage(page)}
							class="group flex flex-col p-4 bg-gray-50 dark:bg-gray-800/30 rounded-lg border border-transparent hover:border-gray-200 dark:hover:border-gray-700 hover:bg-white dark:hover:bg-gray-800/50 transition-all text-left"
						>
							<div class="w-8 h-8 rounded flex items-center justify-center mb-3 bg-gray-100 dark:bg-gray-800">
								{#if page.icon && getIconPath(page.icon)}
									<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getIconPath(page.icon)} />
									</svg>
								{:else}
									<svg class="w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								{/if}
							</div>
							<div class="text-sm font-medium text-gray-900 dark:text-white truncate mb-1">
								{page.name || 'New page'}
							</div>
							<div class="text-xs text-gray-400 dark:text-gray-500">
								{formatDate(page.updated_at)}
							</div>
						</button>
					{/each}
				</div>
			{:else}
				<div class="space-y-0.5">
					{#each filteredPages as page (page.id)}
						<button
							onclick={() => onSelectPage(page)}
							class="group w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800/50 transition-colors text-left"
						>
							<div class="w-6 h-6 rounded flex items-center justify-center flex-shrink-0 bg-gray-100 dark:bg-gray-800">
								{#if page.icon && getIconPath(page.icon)}
									<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getIconPath(page.icon)} />
									</svg>
								{:else}
									<svg class="w-4 h-4 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
								{/if}
							</div>
							<span class="flex-1 text-sm text-gray-700 dark:text-gray-300 truncate">{page.name || 'New page'}</span>
							<span class="text-xs text-gray-400 dark:text-gray-500">{formatDate(page.updated_at)}</span>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Tree Search Modal -->
{#if showTreeSearch}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={(e) => { if (e.target === e.currentTarget) showTreeSearch = false; }}>
		<div class="bg-white dark:bg-[#191919] rounded-xl shadow-2xl w-full max-w-4xl h-[80vh] flex flex-col" onclick={(e) => e.stopPropagation()}>
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Tree Search</h2>
				<button
					onclick={() => showTreeSearch = false}
					class="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
					aria-label="Close"
				>
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
			<div class="flex-1 overflow-hidden">
				<TreeSearchPanel onItemSelect={handleTreeSearchItemSelect} />
			</div>
		</div>
	</div>
{/if}
