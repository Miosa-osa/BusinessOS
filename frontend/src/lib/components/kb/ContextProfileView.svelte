<script lang="ts">
	import { slide } from 'svelte/transition';
	import type { ContextListItem } from '$lib/api/client';

	interface Props {
		profile: ContextListItem;
		children: ContextListItem[];
		allPages: ContextListItem[];
		onAddPage?: () => void;
		onSelectPage?: (page: ContextListItem) => void;
		onPageAction?: (action: string, page: ContextListItem) => void;
		onUpdateProfile?: (updates: { name?: string; icon?: string }) => void;
	}

	let {
		profile,
		children,
		allPages,
		onAddPage,
		onSelectPage,
		onPageAction,
		onUpdateProfile
	}: Props = $props();

	// Icon presets for rendering icon IDs as SVGs
	const iconPresets = [
		{ id: 'document', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		{ id: 'folder', path: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
		{ id: 'clipboard', path: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2' },
		{ id: 'chart', path: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' },
		{ id: 'user', path: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
		{ id: 'users', path: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
		{ id: 'building', path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4' },
		{ id: 'briefcase', path: 'M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ id: 'lightbulb', path: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z' },
		{ id: 'star', path: 'M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z' },
	];

	function getIconPath(iconId: string | null): string | null {
		if (!iconId) return null;
		const found = iconPresets.find(i => i.id === iconId);
		return found?.path || null;
	}

	// Editable title state
	let isEditingTitle = $state(false);
	let editedTitle = $state(profile.name);

	// Update editedTitle when profile changes
	$effect(() => {
		editedTitle = profile.name;
	});

	function startEditingTitle() {
		editedTitle = profile.name;
		isEditingTitle = true;
	}

	function saveTitle() {
		if (editedTitle.trim() && editedTitle !== profile.name) {
			onUpdateProfile?.({ name: editedTitle.trim() });
		}
		isEditingTitle = false;
	}

	function cancelEditTitle() {
		editedTitle = profile.name;
		isEditingTitle = false;
	}

	// Tabs definition
	type TabId = 'overview' | 'pages' | 'conversations' | 'voice' | 'artifacts' | 'calendar' | 'activity';

	interface Tab {
		id: TabId;
		label: string;
		icon: string;
		count?: number;
	}

	const tabs: Tab[] = [
		{ id: 'overview', label: 'Overview', icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'pages', label: 'Pages', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		{ id: 'conversations', label: 'Chats', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' },
		{ id: 'voice', label: 'Voice', icon: 'M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z' },
		{ id: 'artifacts', label: 'Artifacts', icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z' },
		{ id: 'calendar', label: 'Calendar', icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'activity', label: 'Activity', icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' }
	];

	let activeTab = $state<TabId>('overview');

	// View modes for the pages tab
	let viewMode = $state<'table' | 'grid' | 'list'>('table');
	let sortBy = $state<'name' | 'updated_at' | 'type'>('updated_at');
	let sortOrder = $state<'asc' | 'desc'>('desc');

	// Sorted children for pages tab
	const sortedChildren = $derived.by(() => {
		const sorted = [...children].sort((a, b) => {
			if (sortBy === 'name') {
				return a.name.localeCompare(b.name);
			} else if (sortBy === 'updated_at') {
				return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
			} else if (sortBy === 'type') {
				return (a.type || '').localeCompare(b.type || '');
			}
			return 0;
		});
		return sortOrder === 'desc' ? sorted : sorted.reverse();
	});

	// Helper functions
	function getTypeIcon(type: string | null): string {
		switch (type) {
			case 'document':
				return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			case 'project':
				return 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01';
			case 'person':
				return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
			case 'business':
				return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
			default:
				return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
		}
	}

	function getProfileIcon(type: string | null): string {
		return getTypeIcon(type);
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));

		if (days === 0) return 'Today';
		if (days === 1) return 'Yesterday';
		if (days < 7) return `${days} days ago`;
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function getChildCount(pageId: string): number {
		return allPages.filter(p => p.parent_id === pageId).length;
	}

	function getProfileColor(type: string | null): string {
		switch (type) {
			case 'business': return 'from-blue-500 to-blue-600';
			case 'person': return 'from-green-500 to-green-600';
			case 'project': return 'from-amber-500 to-amber-600';
			default: return 'from-purple-500 to-purple-600';
		}
	}
</script>

<div class="flex flex-col h-full bg-white dark:bg-[#1c1c1e]">
	<!-- Profile Header -->
	<div class="flex-shrink-0 px-6 pt-6 pb-4 border-b border-gray-200 dark:border-gray-700">
		<div class="flex items-center gap-4">
			<!-- Profile Icon -->
			<div class="w-14 h-14 rounded-xl flex items-center justify-center bg-gradient-to-br {getProfileColor(profile.type)}">
				{#if profile.icon}
					<span class="text-2xl">{profile.icon}</span>
				{:else}
					<svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getProfileIcon(profile.type)} />
					</svg>
				{/if}
			</div>
			<!-- Profile Info -->
			<div class="flex-1 min-w-0">
				{#if isEditingTitle}
					<input
						type="text"
						bind:value={editedTitle}
						onblur={saveTitle}
						onkeydown={(e) => {
							if (e.key === 'Enter') saveTitle();
							if (e.key === 'Escape') cancelEditTitle();
						}}
						class="text-2xl font-semibold text-gray-900 dark:text-white bg-transparent border-b-2 border-blue-500 outline-none w-full"
						autofocus
					/>
				{:else}
					<h1
						onclick={startEditingTitle}
						class="text-2xl font-semibold text-gray-900 dark:text-white truncate cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 rounded px-1 -mx-1 transition-colors"
						title="Click to edit"
					>
						{profile.name || 'New page'}
					</h1>
				{/if}
				<p class="text-sm text-gray-500 dark:text-gray-400 capitalize">
					{profile.type || 'Document'} &bull; Updated {formatDate(profile.updated_at)}
				</p>
			</div>
			<!-- Actions -->
			<button
				onclick={() => onAddPage?.()}
				class="btn-pill btn-pill-primary flex items-center gap-2"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				<span>New</span>
			</button>
		</div>
	</div>

	<!-- Tabs -->
	<div class="flex-shrink-0 px-6 border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
		<div class="flex items-center gap-1 -mb-px">
			{#each tabs as tab}
				<button
					onclick={() => activeTab = tab.id}
					class="flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors whitespace-nowrap
						{activeTab === tab.id
							? 'border-blue-500 text-blue-600 dark:text-blue-400'
							: 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={tab.icon} />
					</svg>
					<span>{tab.label}</span>
					{#if tab.id === 'pages'}
						<span class="ml-1 px-1.5 py-0.5 text-xs bg-gray-100 dark:bg-gray-700 rounded">
							{children.length}
						</span>
					{/if}
				</button>
			{/each}
		</div>
	</div>

	<!-- Tab Content -->
	<div class="flex-1 overflow-auto">
		{#if activeTab === 'overview'}
			<!-- Overview Tab -->
			<div class="p-6 space-y-6">
				<!-- Stats -->
				<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
					<div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-xl">
						<div class="text-2xl font-bold text-gray-900 dark:text-white">{children.length}</div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Pages</div>
					</div>
					<div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-xl">
						<div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Conversations</div>
					</div>
					<div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-xl">
						<div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Voice Notes</div>
					</div>
					<div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-xl">
						<div class="text-2xl font-bold text-gray-900 dark:text-white">0</div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Artifacts</div>
					</div>
				</div>

				<!-- About -->
				<div>
					<h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">About</h3>
					<div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-xl text-gray-600 dark:text-gray-300">
						<span class="italic text-gray-400">No description available</span>
					</div>
				</div>

				<!-- Properties -->
				<div>
					<h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Properties</h3>
					<div class="p-4 bg-gray-50 dark:bg-gray-800 rounded-xl space-y-3">
						<div class="flex items-center justify-between">
							<span class="text-sm text-gray-500 dark:text-gray-400">Type</span>
							<span class="text-sm text-gray-900 dark:text-white capitalize">{profile.type || 'Document'}</span>
						</div>
						<div class="flex items-center justify-between">
							<span class="text-sm text-gray-500 dark:text-gray-400">Last Updated</span>
							<span class="text-sm text-gray-900 dark:text-white">{new Date(profile.updated_at).toLocaleDateString()}</span>
						</div>
						{#if profile.word_count}
							<div class="flex items-center justify-between">
								<span class="text-sm text-gray-500 dark:text-gray-400">Word Count</span>
								<span class="text-sm text-gray-900 dark:text-white">{profile.word_count.toLocaleString()}</span>
							</div>
						{/if}
					</div>
				</div>

				<!-- Recent Activity -->
				<div>
					<h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Recent Pages</h3>
					{#if children.length > 0}
						<div class="space-y-1">
							{#each children.slice(0, 5) as page}
								<button
									onclick={() => onSelectPage?.(page)}
									class="w-full flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-xl transition-colors text-left"
								>
									<div class="w-8 h-8 rounded-lg bg-white dark:bg-gray-700 flex items-center justify-center flex-shrink-0">
										{#if page.icon && getIconPath(page.icon)}
											<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getIconPath(page.icon)} />
											</svg>
										{:else}
											<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeIcon(page.type)} />
											</svg>
										{/if}
									</div>
									<div class="flex-1 min-w-0">
										<div class="text-sm font-medium text-gray-900 dark:text-white truncate">{page.name || 'New page'}</div>
										<div class="text-xs text-gray-500 dark:text-gray-400">{formatDate(page.updated_at)}</div>
									</div>
								</button>
							{/each}
						</div>
					{:else}
						<div class="p-8 bg-gray-50 dark:bg-gray-800 rounded-xl text-center">
							<p class="text-sm text-gray-500 dark:text-gray-400">No pages yet</p>
						</div>
					{/if}
				</div>
			</div>

		{:else if activeTab === 'pages'}
			<!-- Pages Tab -->
			<div class="p-4">
				<!-- View Controls -->
				<div class="flex items-center justify-between mb-4">
					<div class="flex items-center gap-1 p-1 bg-gray-100 dark:bg-gray-800 rounded-lg">
						<button
							onclick={() => viewMode = 'table'}
							class="p-1.5 rounded-md transition-colors {viewMode === 'table' ? 'bg-white dark:bg-gray-700 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M3 14h18M3 18h18M3 6h18" />
							</svg>
						</button>
						<button
							onclick={() => viewMode = 'grid'}
							class="p-1.5 rounded-md transition-colors {viewMode === 'grid' ? 'bg-white dark:bg-gray-700 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
							</svg>
						</button>
						<button
							onclick={() => viewMode = 'list'}
							class="p-1.5 rounded-md transition-colors {viewMode === 'list' ? 'bg-white dark:bg-gray-700 shadow-sm' : 'text-gray-500 hover:text-gray-700'}"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
							</svg>
						</button>
					</div>

					<div class="flex items-center gap-2 text-sm text-gray-500">
						<span>Sort by</span>
						<select
							bind:value={sortBy}
							class="bg-transparent border border-gray-200 dark:border-gray-700 rounded px-2 py-1 text-gray-700 dark:text-gray-300"
						>
							<option value="updated_at">Last updated</option>
							<option value="name">Name</option>
							<option value="type">Type</option>
						</select>
						<button
							onclick={() => sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'}
							class="p-1 hover:bg-gray-100 dark:hover:bg-gray-800 rounded"
						>
							<svg class="w-4 h-4 transition-transform {sortOrder === 'asc' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
							</svg>
						</button>
					</div>
				</div>

				{#if children.length === 0}
					<div class="flex flex-col items-center justify-center py-16 text-center">
						<div class="w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
							<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
							</svg>
						</div>
						<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No pages yet</h3>
						<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
							Create your first page in {profile.name}
						</p>
						<button
							onclick={() => onAddPage?.()}
							class="btn-pill btn-pill-primary flex items-center gap-2"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
							<span>New page</span>
						</button>
					</div>
				{:else if viewMode === 'table'}
					<table class="w-full">
						<thead>
							<tr class="border-b border-gray-200 dark:border-gray-700">
								<th class="text-left py-2 px-3 text-sm font-medium text-gray-500">Name</th>
								<th class="text-left py-2 px-3 text-sm font-medium text-gray-500 hidden sm:table-cell">Type</th>
								<th class="text-left py-2 px-3 text-sm font-medium text-gray-500 hidden md:table-cell">Updated</th>
								<th class="w-10"></th>
							</tr>
						</thead>
						<tbody>
							{#each sortedChildren as child (child.id)}
								<tr
									class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer transition-colors"
									onclick={() => onSelectPage?.(child)}
								>
									<td class="py-3 px-3">
										<div class="flex items-center gap-3">
											<div class="w-8 h-8 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center flex-shrink-0">
												{#if child.icon}
													<span class="text-lg">{child.icon}</span>
												{:else}
													<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeIcon(child.type)} />
													</svg>
												{/if}
											</div>
											<div class="min-w-0">
												<div class="font-medium text-gray-900 dark:text-white truncate">
													{child.name || 'New page'}
												</div>
												{#if getChildCount(child.id) > 0}
													<div class="text-xs text-gray-400">
														{getChildCount(child.id)} nested
													</div>
												{/if}
											</div>
										</div>
									</td>
									<td class="py-3 px-3 hidden sm:table-cell">
										<span class="text-sm text-gray-500 capitalize">{child.type || 'Document'}</span>
									</td>
									<td class="py-3 px-3 hidden md:table-cell">
										<span class="text-sm text-gray-500">{formatDate(child.updated_at)}</span>
									</td>
									<td class="py-3 px-3">
										<button
											onclick={(e) => { e.stopPropagation(); onPageAction?.('menu', child); }}
											class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400 hover:text-gray-600"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
											</svg>
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				{:else if viewMode === 'grid'}
					<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
						{#each sortedChildren as child (child.id)}
							<button
								onclick={() => onSelectPage?.(child)}
								class="group flex flex-col items-center p-4 rounded-xl border border-gray-200 dark:border-gray-700 hover:border-blue-500 hover:bg-blue-50/50 dark:hover:bg-blue-900/20 transition-all text-left"
							>
								<div class="w-12 h-12 rounded-xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-3 group-hover:bg-blue-100 dark:group-hover:bg-blue-900/50">
									{#if child.icon}
										<span class="text-2xl">{child.icon}</span>
									{:else}
										<svg class="w-6 h-6 text-gray-400 group-hover:text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeIcon(child.type)} />
										</svg>
									{/if}
								</div>
								<div class="text-sm font-medium text-gray-900 dark:text-white text-center truncate w-full">
									{child.name || 'New page'}
								</div>
								<div class="text-xs text-gray-400 mt-1">
									{formatDate(child.updated_at)}
								</div>
							</button>
						{/each}
					</div>
				{:else}
					<div class="space-y-1">
						{#each sortedChildren as child (child.id)}
							<button
								onclick={() => onSelectPage?.(child)}
								class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-left"
							>
								<div class="w-6 h-6 rounded flex items-center justify-center flex-shrink-0">
									{#if child.icon}
										<span class="text-base">{child.icon}</span>
									{:else}
										<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getTypeIcon(child.type)} />
										</svg>
									{/if}
								</div>
								<div class="flex-1 min-w-0">
									<span class="text-sm text-gray-900 dark:text-white truncate block">
										{child.name || 'New page'}
									</span>
								</div>
								<span class="text-xs text-gray-400 flex-shrink-0">
									{formatDate(child.updated_at)}
								</span>
							</button>
						{/each}
					</div>
				{/if}
			</div>

		{:else if activeTab === 'conversations'}
			<!-- Conversations Tab -->
			<div class="p-6">
				<div class="flex flex-col items-center justify-center py-16 text-center">
					<div class="w-16 h-16 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center mb-4">
						<svg class="w-8 h-8 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
						</svg>
					</div>
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No conversations yet</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
						Start a chat with this context selected to link conversations here. All AI interactions related to {profile.name} will appear in this tab.
					</p>
				</div>
			</div>

		{:else if activeTab === 'voice'}
			<!-- Voice Notes Tab -->
			<div class="p-6">
				<div class="flex flex-col items-center justify-center py-16 text-center">
					<div class="w-16 h-16 rounded-full bg-purple-50 dark:bg-purple-900/20 flex items-center justify-center mb-4">
						<svg class="w-8 h-8 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
						</svg>
					</div>
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No voice notes</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
						Voice recordings linked to {profile.name} will appear here along with their transcriptions.
					</p>
				</div>
			</div>

		{:else if activeTab === 'artifacts'}
			<!-- Artifacts Tab -->
			<div class="p-6">
				<div class="flex flex-col items-center justify-center py-16 text-center">
					<div class="w-16 h-16 rounded-full bg-amber-50 dark:bg-amber-900/20 flex items-center justify-center mb-4">
						<svg class="w-8 h-8 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
						</svg>
					</div>
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No artifacts</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
						AI-generated content like proposals, SOPs, frameworks, and reports for {profile.name} will appear here.
					</p>
				</div>
			</div>

		{:else if activeTab === 'calendar'}
			<!-- Calendar Tab -->
			<div class="p-6">
				<div class="flex flex-col items-center justify-center py-16 text-center">
					<div class="w-16 h-16 rounded-full bg-green-50 dark:bg-green-900/20 flex items-center justify-center mb-4">
						<svg class="w-8 h-8 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
					</div>
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No calendar events</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
						Meetings, deadlines, and events linked to {profile.name} will appear here.
					</p>
				</div>
			</div>

		{:else if activeTab === 'activity'}
			<!-- Activity Tab -->
			<div class="p-6">
				<div class="flex flex-col items-center justify-center py-16 text-center">
					<div class="w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
						<svg class="w-8 h-8 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">Activity timeline</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400 max-w-sm">
						A unified timeline of all changes, interactions, and updates related to {profile.name} will be shown here.
					</p>
				</div>
			</div>
		{/if}
	</div>
</div>
