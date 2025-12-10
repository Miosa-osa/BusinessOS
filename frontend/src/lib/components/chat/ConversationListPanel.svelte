<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import ConversationListItem from './ConversationListItem.svelte';

	interface Conversation {
		id: string;
		title: string;
		preview?: string;
		timestamp: string;
		projectName?: string;
		pinned?: boolean;
		messageCount?: number;
	}

	interface Props {
		conversations?: Conversation[];
		activeId?: string;
		searchQuery?: string;
		onSelect?: (id: string) => void;
		onNewChat?: () => void;
		onSearch?: (query: string) => void;
		onRename?: (id: string) => void;
		onPin?: (id: string) => void;
		onDelete?: (id: string) => void;
		onArchive?: (id: string) => void;
		onExport?: (id: string) => void;
		onLinkProject?: (id: string) => void;
	}

	let {
		conversations = [],
		activeId = '',
		searchQuery = $bindable(''),
		onSelect,
		onNewChat,
		onSearch,
		onRename,
		onPin,
		onDelete,
		onArchive,
		onExport,
		onLinkProject
	}: Props = $props();

	let filterTab: 'all' | 'pinned' | 'recent' = $state('all');

	const filteredConversations = $derived(() => {
		let filtered = conversations;

		// Apply search filter
		if (searchQuery) {
			const query = searchQuery.toLowerCase();
			filtered = filtered.filter(c =>
				c.title.toLowerCase().includes(query) ||
				c.preview?.toLowerCase().includes(query)
			);
		}

		// Apply tab filter
		if (filterTab === 'pinned') {
			filtered = filtered.filter(c => c.pinned);
		} else if (filterTab === 'recent') {
			// Sort by timestamp, most recent first
			filtered = [...filtered].sort((a, b) =>
				new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
			).slice(0, 10);
		}

		// Always show pinned at top for 'all' tab
		if (filterTab === 'all') {
			const pinned = filtered.filter(c => c.pinned);
			const unpinned = filtered.filter(c => !c.pinned);
			filtered = [...pinned, ...unpinned];
		}

		return filtered;
	});

	function handleSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		searchQuery = target.value;
		onSearch?.(searchQuery);
	}
</script>

<div class="h-full flex flex-col bg-card">
	<!-- Header -->
	<div class="flex-shrink-0 p-4 border-b border-border">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold text-foreground">Chats</h2>
			<button
				onclick={onNewChat}
				class="w-8 h-8 flex items-center justify-center bg-gray-900 text-white rounded-lg hover:bg-gray-800 transition-colors"
				title="New chat"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
			</button>
		</div>

		<!-- Search -->
		<div class="relative">
			<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				placeholder="Search conversations..."
				value={searchQuery}
				oninput={handleSearchInput}
				class="w-full pl-10 pr-4 py-2 text-sm bg-gray-50 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent transition-all"
			/>
		</div>

		<!-- Filter Tabs -->
		<div class="flex items-center gap-1 mt-3">
			<button
				onclick={() => filterTab = 'all'}
				class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors
					{filterTab === 'all' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
			>
				All
			</button>
			<button
				onclick={() => filterTab = 'pinned'}
				class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors
					{filterTab === 'pinned' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
			>
				Pinned
			</button>
			<button
				onclick={() => filterTab = 'recent'}
				class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors
					{filterTab === 'recent' ? 'bg-gray-900 text-white' : 'text-gray-600 hover:bg-gray-100'}"
			>
				Recent
			</button>
		</div>
	</div>

	<!-- Conversation List -->
	<div class="flex-1 overflow-y-auto p-2">
		{#if filteredConversations().length === 0}
			<div class="flex flex-col items-center justify-center py-12 text-center" in:fade={{ duration: 200 }}>
				<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
					<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
					</svg>
				</div>
				<p class="text-sm text-gray-500 mb-1">
					{searchQuery ? 'No conversations found' : 'No conversations yet'}
				</p>
				<p class="text-xs text-gray-400">
					{searchQuery ? 'Try a different search term' : 'Start a new chat to begin'}
				</p>
			</div>
		{:else}
			<div class="space-y-1">
				{#each filteredConversations() as conversation (conversation.id)}
					<div in:fly={{ y: 10, duration: 200 }}>
						<ConversationListItem
							{...conversation}
							active={conversation.id === activeId}
							onclick={() => onSelect?.(conversation.id)}
							onRename={() => onRename?.(conversation.id)}
							onPin={() => onPin?.(conversation.id)}
							onDelete={() => onDelete?.(conversation.id)}
							onArchive={() => onArchive?.(conversation.id)}
							onExport={() => onExport?.(conversation.id)}
							onLinkProject={() => onLinkProject?.(conversation.id)}
						/>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Footer -->
	<div class="flex-shrink-0 p-3 border-t border-border">
		<button
			class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
			</svg>
			View archived
		</button>
	</div>
</div>
