<script lang="ts">
	import { slide } from 'svelte/transition';
	import SidebarPageItem from './SidebarPageItem.svelte';
	import type { ContextListItem } from '$lib/api/client';

	interface Props {
		// Data
		pages: ContextListItem[];
		favorites: ContextListItem[];
		recentPages: ContextListItem[];
		memories?: any[]; // Learning items from the store

		// State
		selectedPageId?: string | null;
		expandedSections?: Set<string>;
		expandedPages?: string[]; // Array for better Svelte 5 reactivity
		favoriteIds?: string[]; // Array for better Svelte 5 reactivity
		searchQuery?: string;
		width?: number;
		isCollapsed?: boolean;

		// Callbacks
		onSectionToggle: (sectionId: string) => void;
		onPageSelect: (page: ContextListItem) => void;
		onPageExpand: (pageId: string) => void;
		onPageAddChild: (page: ContextListItem) => void;
		onPageOpenPeek: (page: ContextListItem) => void;
		onPageOpenCenterPeek?: (page: ContextListItem) => void;
		onPageDuplicate: (page: ContextListItem) => void;
		onPageRename: (page: ContextListItem, newName: string) => void;
		onPageMove: (page: ContextListItem) => void;
		onPageDelete: (page: ContextListItem) => void;
		onPageToggleFavorite: (page: ContextListItem) => void;
		onPageCopyLink: (page: ContextListItem) => void;
		onAddPage: (parentId?: string) => void;
		onAddProfile?: (type: 'business' | 'person' | 'project' | 'custom') => void;
		onSearch: (query: string) => void;
		onOpenCommandPalette?: () => void;
		onGoHome?: () => void;
		onOpenGraph?: () => void;
		isGraphView?: boolean;
		onWidthChange?: (width: number) => void;
		onToggleCollapse?: () => void;
		onSearchInputRef?: (input: HTMLInputElement) => void;
	}

	let {
		pages,
		favorites,
		recentPages,
		memories = [],
		selectedPageId = null,
		expandedSections = new Set(['favorites', 'context-profiles', 'documents']),
		expandedPages = [],
		favoriteIds = [],
		searchQuery = '',
		width = 280,
		isCollapsed = false,
		onSectionToggle,
		onPageSelect,
		onPageExpand,
		onPageAddChild,
		onPageOpenPeek,
		onPageOpenCenterPeek,
		onPageDuplicate,
		onPageRename,
		onPageMove,
		onPageDelete,
		onPageToggleFavorite,
		onPageCopyLink,
		onAddPage,
		onAddProfile,
		onSearch,
		onOpenCommandPalette,
		onGoHome,
		onOpenGraph,
		isGraphView = false,
		onWidthChange,
		onToggleCollapse,
		onSearchInputRef
	}: Props = $props();

	// Search input binding
	let searchInputElement: HTMLInputElement | null = $state(null);

	// New page menu state
	let newPageMenuOpen = $state(false);

	$effect(() => {
		if (searchInputElement && onSearchInputRef) {
			onSearchInputRef(searchInputElement);
		}
	});

	// CONTEXT PROFILES vs DOCUMENTS
	// Context Profiles: business, person, project, custom types (top-level only)
	// Documents: document type or no type (top-level only)

	// All Context Profiles (any type except document, no parent)
	const contextProfiles = $derived(pages.filter(p =>
		p.type && p.type !== 'document' && !p.parent_id
	));

	// All Documents (document type or no type, no parent)
	const documents = $derived(pages.filter(p =>
		(!p.type || p.type === 'document') && !p.parent_id
	));

	// Get child pages for a given parent
	function getChildPages(parentId: string): ContextListItem[] {
		return pages.filter(p => p.parent_id === parentId);
	}

	// Check if page has children
	function hasChildren(page: ContextListItem): boolean {
		return getChildPages(page.id).length > 0;
	}

	// Search filtering
	const filteredPages = $derived.by(() => {
		if (!searchQuery.trim()) return null;
		const query = searchQuery.toLowerCase();
		return pages.filter(p => p.name.toLowerCase().includes(query));
	});

	// Section expand states
	const showContextProfiles = $derived(expandedSections.has('context-profiles'));
	const showDocuments = $derived(expandedSections.has('documents'));
	const showMemories = $derived(expandedSections.has('memories'));


	// Resize handling
	let isResizing = $state(false);

	function startResize(e: MouseEvent) {
		isResizing = true;
		e.preventDefault();
		document.addEventListener('mousemove', handleResize);
		document.addEventListener('mouseup', stopResize);
	}

	function handleResize(e: MouseEvent) {
		if (!isResizing) return;
		const newWidth = e.clientX;
		const clampedWidth = Math.max(200, Math.min(400, newWidth));
		onWidthChange?.(clampedWidth);
	}

	function stopResize() {
		isResizing = false;
		document.removeEventListener('mousemove', handleResize);
		document.removeEventListener('mouseup', stopResize);
	}

	// Sections expand state
	const showFavorites = $derived(expandedSections.has('favorites'));

	// Click outside handler for new page menu
	function handleClickOutside(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (newPageMenuOpen && !target.closest('.new-page-menu-container')) {
			newPageMenuOpen = false;
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

<div
	class="h-full flex flex-col bg-gray-50 dark:bg-[#1a1a1c] border-r border-gray-200 dark:border-gray-800 relative"
	style="width: {isCollapsed ? 48 : width}px;"
>
	{#if !isCollapsed}
		<!-- Top Quick Actions (Notion-style) -->
		<div class="flex-shrink-0 px-2 pt-3 pb-1">
			<!-- Search Button -->
			<button
				onclick={() => onOpenCommandPalette?.()}
				class="w-full flex items-center gap-3 px-2.5 py-1.5 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
				<span>Search</span>
				<kbd class="ml-auto text-[10px] text-gray-400 bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded">⌘K</kbd>
			</button>

			<!-- Home Button -->
			<button
				onclick={() => onGoHome?.()}
				class="w-full flex items-center gap-3 px-2.5 py-1.5 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
				</svg>
				<span>Home</span>
			</button>

			<!-- Knowledge Graph Button -->
			<button
				onclick={() => onOpenGraph?.()}
				class="w-full flex items-center gap-3 px-2.5 py-1.5 text-sm transition-colors rounded-md {isGraphView ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
				</svg>
				<span>Knowledge Graph</span>
			</button>

				<!-- New Page Button with Native Dropdown -->
			<div class="relative new-page-menu-container">
				<button
					onclick={() => newPageMenuOpen = !newPageMenuOpen}
					class="w-full flex items-center gap-3 px-2.5 py-1.5 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					<span class="flex-1 text-left">New page</span>
					<svg class="w-3 h-3 text-gray-400 transition-transform {newPageMenuOpen ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
					</svg>
				</button>

				{#if newPageMenuOpen}
					<div
						class="absolute left-0 top-full mt-1 w-full min-w-[200px] bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-lg p-1 z-[100]"
						transition:slide={{ duration: 100 }}
					>
						<!-- Context Profiles Section -->
						<div class="px-3 py-1.5 text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">
							Context Profile
						</div>

						<button
							class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer text-left"
							onclick={() => { newPageMenuOpen = false; onAddProfile?.('business'); }}
						>
							<svg class="w-4 h-4 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
							</svg>
							<span>Business</span>
						</button>

						<button
							class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer text-left"
							onclick={() => { newPageMenuOpen = false; onAddProfile?.('person'); }}
						>
							<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
							<span>Person</span>
						</button>

						<button
							class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer text-left"
							onclick={() => { newPageMenuOpen = false; onAddProfile?.('project'); }}
						>
							<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
							<span>Project</span>
						</button>

						<button
							class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer text-left"
							onclick={() => { newPageMenuOpen = false; onAddProfile?.('custom'); }}
						>
							<svg class="w-4 h-4 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
							</svg>
							<span>Custom</span>
						</button>

						<div class="my-1 h-px bg-gray-200 dark:bg-gray-700"></div>

						<div class="px-3 py-1.5 text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">
							Document
						</div>

						<button
							class="w-full flex items-center gap-3 px-3 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg cursor-pointer text-left"
							onclick={() => { newPageMenuOpen = false; onAddPage(); }}
						>
							<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
							</svg>
							<span>Blank Document</span>
						</button>
					</div>
				{/if}
			</div>
		</div>

		<!-- Scrollable Content -->
		<div class="flex-1 overflow-y-auto px-2 py-1">
			{#if filteredPages}
				<!-- Search Results -->
				<div class="mb-2">
					<div class="px-2 py-1.5 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Search Results ({filteredPages?.length})
					</div>
					{#each filteredPages || [] as page (page.id)}
						<SidebarPageItem
							{page}
							depth={0}
							isSelected={selectedPageId === page.id}
							isExpanded={expandedPages.includes(page.id)}
							hasChildren={hasChildren(page)}
							isFavorite={favoriteIds.includes(page.id)}
							onSelect={() => onPageSelect(page)}
							onExpand={() => onPageExpand(page.id)}
							onAddChild={() => onPageAddChild(page)}
							onOpenPeek={() => onPageOpenPeek(page)}
							onOpenCenterPeek={() => onPageOpenCenterPeek?.(page)}
							onDuplicate={() => onPageDuplicate(page)}
							onRename={() => onPageRename(page, page.name)}
							onMove={() => onPageMove(page)}
							onDelete={() => onPageDelete(page)}
							onToggleFavorite={() => onPageToggleFavorite(page)}
							onCopyLink={() => onPageCopyLink(page)}
						/>
					{/each}
				</div>
			{:else}
				<!-- Favorites Section -->
				{#if favorites.length > 0}
					<div class="mb-2">
						<button
							onclick={() => onSectionToggle('favorites')}
							class="w-full flex items-center gap-2 px-2 py-1.5 text-xs font-medium text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
						>
							<svg
								class="w-3 h-3 text-gray-400 transition-transform {showFavorites ? '' : '-rotate-90'}"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
							</svg>
							<svg class="w-3.5 h-3.5 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
								<path d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
							</svg>
							<span class="uppercase tracking-wider">Favorites</span>
						</button>
						{#if showFavorites}
							<div transition:slide={{ duration: 150 }}>
								{#each favorites as page (page.id)}
									<SidebarPageItem
										{page}
										depth={0}
										isSelected={selectedPageId === page.id}
										isExpanded={expandedPages.includes(page.id)}
										hasChildren={hasChildren(page)}
										isFavorite={true}
										onSelect={() => onPageSelect(page)}
										onExpand={() => onPageExpand(page.id)}
										onAddChild={() => onPageAddChild(page)}
										onOpenPeek={() => onPageOpenPeek(page)}
										onOpenCenterPeek={() => onPageOpenCenterPeek?.(page)}
										onDuplicate={() => onPageDuplicate(page)}
										onRename={() => onPageRename(page, page.name)}
										onMove={() => onPageMove(page)}
										onDelete={() => onPageDelete(page)}
										onToggleFavorite={() => onPageToggleFavorite(page)}
										onCopyLink={() => onPageCopyLink(page)}
									/>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

				<!-- CONTEXT PROFILES Section (flat list) -->
				<div class="mb-2">
					<button
						onclick={() => onSectionToggle('context-profiles')}
						class="w-full flex items-center gap-2 px-2 py-1.5 text-xs font-medium text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
					>
						<svg class="w-3 h-3 text-gray-400 transition-transform {showContextProfiles ? '' : '-rotate-90'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
						<svg class="w-3.5 h-3.5 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
						</svg>
						<span class="uppercase tracking-wider">Context Profiles</span>
						<span class="ml-auto text-gray-400 text-[10px]">{contextProfiles.length}</span>
					</button>
					{#if showContextProfiles}
						<div transition:slide={{ duration: 150 }}>
							{#each contextProfiles as profile (profile.id)}
								<SidebarPageItem
									page={profile}
									depth={0}
									isSelected={selectedPageId === profile.id}
									isExpanded={expandedPages.includes(profile.id)}
									hasChildren={hasChildren(profile)}
									isFavorite={favoriteIds.includes(profile.id)}
									onSelect={() => onPageSelect(profile)}
									onExpand={() => onPageExpand(profile.id)}
									onAddChild={() => onPageAddChild(profile)}
									onOpenPeek={() => onPageOpenPeek(profile)}
									onOpenCenterPeek={() => onPageOpenCenterPeek?.(profile)}
									onDuplicate={() => onPageDuplicate(profile)}
									onRename={() => onPageRename(profile, profile.name)}
									onMove={() => onPageMove(profile)}
									onDelete={() => onPageDelete(profile)}
									onToggleFavorite={() => onPageToggleFavorite(profile)}
									onCopyLink={() => onPageCopyLink(profile)}
								/>
								<!-- Children of profile (documents inside) -->
								{#if expandedPages.includes(profile.id)}
									{@const profileChildren = getChildPages(profile.id)}
									{#if profileChildren.length === 0}
										<div class="py-1.5 text-xs text-gray-400 dark:text-gray-500 italic" style="padding-left: 40px;">
											No pages inside
										</div>
									{:else}
										{#each profileChildren as child (child.id)}
										<SidebarPageItem
											page={child}
											depth={1}
											isSelected={selectedPageId === child.id}
											isExpanded={expandedPages.includes(child.id)}
											hasChildren={hasChildren(child)}
											isFavorite={favoriteIds.includes(child.id)}
											onSelect={() => onPageSelect(child)}
											onExpand={() => onPageExpand(child.id)}
											onAddChild={() => onPageAddChild(child)}
											onOpenPeek={() => onPageOpenPeek(child)}
											onOpenCenterPeek={() => onPageOpenCenterPeek?.(child)}
											onDuplicate={() => onPageDuplicate(child)}
											onRename={() => onPageRename(child, child.name)}
											onMove={() => onPageMove(child)}
											onDelete={() => onPageDelete(child)}
											onToggleFavorite={() => onPageToggleFavorite(child)}
											onCopyLink={() => onPageCopyLink(child)}
										/>
										{/each}
									{/if}
								{/if}
							{/each}
							{#if contextProfiles.length === 0}
								<div class="px-4 py-3 text-xs text-gray-400 dark:text-gray-500 italic">
									No context profiles yet. Click "New page" to create one.
								</div>
							{/if}
						</div>
					{/if}
				</div>

				<!-- DOCUMENTS Section -->
				<div class="mb-2">
					<button
						onclick={() => onSectionToggle('documents')}
						class="w-full flex items-center gap-2 px-2 py-1.5 text-xs font-medium text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
					>
						<svg class="w-3 h-3 text-gray-400 transition-transform {showDocuments ? '' : '-rotate-90'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
						<svg class="w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
						<span class="uppercase tracking-wider">Documents</span>
						<span class="ml-auto text-gray-400 text-[10px]">{documents.length}</span>
					</button>
					{#if showDocuments}
						<div transition:slide={{ duration: 150 }}>
							{#each documents as doc (doc.id)}
								<SidebarPageItem
									page={doc}
									depth={0}
									isSelected={selectedPageId === doc.id}
									isExpanded={expandedPages.includes(doc.id)}
									hasChildren={hasChildren(doc)}
									isFavorite={favoriteIds.includes(doc.id)}
									onSelect={() => onPageSelect(doc)}
									onExpand={() => onPageExpand(doc.id)}
									onAddChild={() => onPageAddChild(doc)}
									onOpenPeek={() => onPageOpenPeek(doc)}
									onOpenCenterPeek={() => onPageOpenCenterPeek?.(doc)}
									onDuplicate={() => onPageDuplicate(doc)}
									onRename={() => onPageRename(doc, doc.name)}
									onMove={() => onPageMove(doc)}
									onDelete={() => onPageDelete(doc)}
									onToggleFavorite={() => onPageToggleFavorite(doc)}
									onCopyLink={() => onPageCopyLink(doc)}
								/>
								<!-- Nested children of document (recursive) -->
								{#if expandedPages.includes(doc.id)}
									{@const docChildren = getChildPages(doc.id)}
									{#if docChildren.length === 0}
										<div class="py-1.5 text-xs text-gray-400 dark:text-gray-500 italic" style="padding-left: 40px;">
											No pages inside
										</div>
									{:else}
										{#each docChildren as child (child.id)}
											<SidebarPageItem
												page={child}
												depth={1}
												isSelected={selectedPageId === child.id}
												isExpanded={expandedPages.includes(child.id)}
												hasChildren={hasChildren(child)}
												isFavorite={favoriteIds.includes(child.id)}
												onSelect={() => onPageSelect(child)}
												onExpand={() => onPageExpand(child.id)}
												onAddChild={() => onPageAddChild(child)}
												onOpenPeek={() => onPageOpenPeek(child)}
												onOpenCenterPeek={() => onPageOpenCenterPeek?.(child)}
												onDuplicate={() => onPageDuplicate(child)}
												onRename={() => onPageRename(child, child.name)}
												onMove={() => onPageMove(child)}
												onDelete={() => onPageDelete(child)}
												onToggleFavorite={() => onPageToggleFavorite(child)}
												onCopyLink={() => onPageCopyLink(child)}
											/>
											<!-- Level 2 children -->
											{#if expandedPages.includes(child.id)}
												{@const grandChildren = getChildPages(child.id)}
												{#if grandChildren.length === 0}
													<div class="py-1.5 text-xs text-gray-400 dark:text-gray-500 italic" style="padding-left: 56px;">
														No pages inside
													</div>
												{:else}
													{#each grandChildren as grandChild (grandChild.id)}
														<SidebarPageItem
															page={grandChild}
															depth={2}
															isSelected={selectedPageId === grandChild.id}
															isExpanded={expandedPages.includes(grandChild.id)}
															hasChildren={hasChildren(grandChild)}
															isFavorite={favoriteIds.includes(grandChild.id)}
															onSelect={() => onPageSelect(grandChild)}
															onExpand={() => onPageExpand(grandChild.id)}
															onAddChild={() => onPageAddChild(grandChild)}
															onOpenPeek={() => onPageOpenPeek(grandChild)}
															onOpenCenterPeek={() => onPageOpenCenterPeek?.(grandChild)}
															onDuplicate={() => onPageDuplicate(grandChild)}
															onRename={() => onPageRename(grandChild, grandChild.name)}
															onMove={() => onPageMove(grandChild)}
															onDelete={() => onPageDelete(grandChild)}
															onToggleFavorite={() => onPageToggleFavorite(grandChild)}
															onCopyLink={() => onPageCopyLink(grandChild)}
														/>
													{/each}
												{/if}
											{/if}
										{/each}
									{/if}
								{/if}
							{/each}
							{#if documents.length === 0}
								<div class="px-4 py-3 text-xs text-gray-400 dark:text-gray-500 italic">
									No documents yet
								</div>
							{/if}
						</div>
					{/if}
				</div>

				<!-- MEMORIES Section -->
				<div class="mb-2">
					<button
						onclick={() => onSectionToggle('memories')}
						class="w-full flex items-center gap-2 px-2 py-1.5 text-xs font-medium text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
					>
						<svg class="w-3 h-3 text-gray-400 transition-transform {showMemories ? '' : '-rotate-90'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
						<svg class="w-3.5 h-3.5 text-pink-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
						</svg>
						<span class="uppercase tracking-wider">Memories</span>
						<span class="ml-auto text-gray-400 text-[10px]">{memories?.length || 0}</span>
					</button>
					{#if showMemories}
						<div transition:slide={{ duration: 150 }}>
							{#each memories || [] as memory (memory.id)}
								<button
									onclick={() => onPageSelect({
										id: memory.id,
										name: memory.learning_summary || memory.learning_content,
										type: 'memory' as any,
										updated_at: memory.updated_at
									} as any)}
									class="w-full flex items-center gap-3 px-2.5 py-1.5 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors {selectedPageId === memory.id ? 'bg-gray-100 dark:bg-gray-800' : ''}"
								>
									<svg class="w-4 h-4 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
									</svg>
									<span class="truncate">{memory.learning_summary || memory.learning_content}</span>
								</button>
							{/each}
							{#if !memories || memories.length === 0}
								<div class="px-4 py-3 text-xs text-gray-400 dark:text-gray-500 italic">
									No memories yet
								</div>
							{/if}
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Bottom Actions -->
		<div class="flex-shrink-0 px-2 pb-3 pt-1 border-t border-gray-200 dark:border-gray-800">
			<button
				class="w-full flex items-center gap-3 px-2.5 py-1.5 text-sm text-gray-500 dark:text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
				</svg>
				<span>Trash</span>
			</button>
		</div>

		<!-- Resize Handle -->
		<button
			class="absolute right-0 top-0 bottom-0 w-1 cursor-ew-resize hover:bg-blue-500/50 active:bg-blue-500 transition-colors {isResizing ? 'bg-blue-500' : ''}"
			onmousedown={startResize}
			aria-label="Resize sidebar"
		></button>
	{:else}
		<!-- Collapsed State -->
		<div class="flex flex-col items-center py-3 gap-2">
			<!-- Expand Button -->
			<button
				onclick={onToggleCollapse}
				class="w-9 h-9 flex items-center justify-center rounded-lg hover:bg-gray-200 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 transition-colors"
				title="Expand sidebar"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
				</svg>
			</button>

			<!-- Search -->
			<button
				onclick={() => onOpenCommandPalette?.()}
				class="w-9 h-9 flex items-center justify-center rounded-lg hover:bg-gray-200 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 transition-colors"
				title="Search (⌘K)"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
			</button>

			<!-- Home -->
			<button
				onclick={() => onGoHome?.()}
				class="w-9 h-9 flex items-center justify-center rounded-lg hover:bg-gray-200 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 transition-colors"
				title="Home"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
				</svg>
			</button>

			<!-- New Page -->
			<button
				onclick={() => onAddPage()}
				class="w-9 h-9 flex items-center justify-center rounded-lg hover:bg-gray-200 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400 transition-colors"
				title="New page"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
			</button>
		</div>
	{/if}
</div>
