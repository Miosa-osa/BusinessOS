<script lang="ts">
	import { slide } from 'svelte/transition';
	import SidebarPageItem from './SidebarPageItem.svelte';
	import type { ContextListItem } from '$lib/api/contexts/types';

	interface Props {
		title: string;
		icon?: string;
		pages: ContextListItem[];
		isExpanded?: boolean;
		maxVisible?: number;
		selectedPageId?: string | null;
		expandedPages?: Set<string>;
		favorites?: Set<string>;
		getChildPages?: (parentId: string) => ContextListItem[];
		onToggle: () => void;
		onAddPage?: () => void;
		onPageSelect: (page: ContextListItem) => void;
		onPageExpand?: (pageId: string) => void;
		onPageAddChild?: (page: ContextListItem) => void;
		onPageOpenPeek?: (page: ContextListItem) => void;
		onPageDuplicate?: (page: ContextListItem) => void;
		onPageRename?: (page: ContextListItem) => void;
		onPageMove?: (page: ContextListItem) => void;
		onPageDelete?: (page: ContextListItem) => void;
		onPageToggleFavorite?: (page: ContextListItem) => void;
		onPageCopyLink?: (page: ContextListItem) => void;
	}

	let {
		title,
		icon,
		pages,
		isExpanded = true,
		maxVisible = 10,
		selectedPageId = null,
		expandedPages = new Set(),
		favorites = new Set(),
		getChildPages,
		onToggle,
		onAddPage,
		onPageSelect,
		onPageExpand,
		onPageAddChild,
		onPageOpenPeek,
		onPageDuplicate,
		onPageRename,
		onPageMove,
		onPageDelete,
		onPageToggleFavorite,
		onPageCopyLink
	}: Props = $props();

	let showAll = $state(false);

	const visiblePages = $derived(
		showAll ? pages : pages.slice(0, maxVisible)
	);

	const hasMore = $derived(pages.length > maxVisible);

	function getSectionIcon(iconName?: string): string {
		switch (iconName) {
			case 'people':
				return 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z';
			case 'business':
				return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
			case 'projects':
				return 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z';
			case 'documents':
				return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
			case 'favorites':
				return 'M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z';
			case 'recent':
				return 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z';
			default:
				return 'M4 6h16M4 12h16M4 18h16';
		}
	}

	function hasChildren(page: ContextListItem): boolean {
		if (!getChildPages) return false;
		return getChildPages(page.id).length > 0;
	}

	function renderPageWithChildren(page: ContextListItem, depth: number = 0) {
		const children = getChildPages?.(page.id) || [];
		const isPageExpanded = expandedPages.has(page.id);
		return { page, children, isPageExpanded, depth };
	}
</script>

<div class="mb-1">
	<!-- Section Header -->
	<button
		onclick={onToggle}
		class="group w-full flex items-center gap-2 px-2 py-1.5 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
	>
		<!-- Expand Arrow -->
		<svg
			class="w-3 h-3 text-gray-400 transition-transform {isExpanded ? '' : '-rotate-90'}"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
		>
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
		</svg>

		<!-- Section Icon -->
		{#if icon}
			<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getSectionIcon(icon)} />
			</svg>
		{/if}

		<!-- Title & Count -->
		<span class="flex-1 text-left">{title}</span>
		<span class="text-gray-400 font-normal">({pages.length})</span>

		<!-- Add Button (on hover) -->
		{#if onAddPage}
			<button
				onclick={(e) => { e.stopPropagation(); onAddPage?.(); }}
				class="w-5 h-5 flex items-center justify-center rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 opacity-0 group-hover:opacity-100 transition-opacity"
				title="Add page"
			>
				<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
			</button>
		{/if}
	</button>

	<!-- Pages List -->
	{#if isExpanded}
		<div transition:slide={{ duration: 200 }} class="mt-0.5" role="group">
			{#each visiblePages as page (page.id)}
				{@const pageData = renderPageWithChildren(page, 0)}
				<SidebarPageItem
					page={pageData.page}
					depth={0}
					isSelected={selectedPageId === page.id}
					isExpanded={pageData.isPageExpanded}
					hasChildren={hasChildren(page)}
					isFavorite={favorites.has(page.id)}
					onSelect={() => onPageSelect(page)}
					onExpand={() => onPageExpand?.(page.id)}
					onAddChild={() => onPageAddChild?.(page)}
					onOpenPeek={() => onPageOpenPeek?.(page)}
					onDuplicate={() => onPageDuplicate?.(page)}
					onRename={() => onPageRename?.(page)}
					onMove={() => onPageMove?.(page)}
					onDelete={() => onPageDelete?.(page)}
					onToggleFavorite={() => onPageToggleFavorite?.(page)}
					onCopyLink={() => onPageCopyLink?.(page)}
				/>

				<!-- Nested Children -->
				{#if pageData.isPageExpanded && pageData.children.length > 0}
					<div transition:slide={{ duration: 150 }}>
						{#each pageData.children as childPage (childPage.id)}
							{@const childData = renderPageWithChildren(childPage, 1)}
							<SidebarPageItem
								page={childData.page}
								depth={1}
								isSelected={selectedPageId === childPage.id}
								isExpanded={childData.isPageExpanded}
								hasChildren={hasChildren(childPage)}
								isFavorite={favorites.has(childPage.id)}
								onSelect={() => onPageSelect(childPage)}
								onExpand={() => onPageExpand?.(childPage.id)}
								onAddChild={() => onPageAddChild?.(childPage)}
								onOpenPeek={() => onPageOpenPeek?.(childPage)}
								onDuplicate={() => onPageDuplicate?.(childPage)}
								onRename={() => onPageRename?.(childPage)}
								onMove={() => onPageMove?.(childPage)}
								onDelete={() => onPageDelete?.(childPage)}
								onToggleFavorite={() => onPageToggleFavorite?.(childPage)}
								onCopyLink={() => onPageCopyLink?.(childPage)}
							/>

							<!-- Level 2 nested children -->
							{#if childData.isPageExpanded && childData.children.length > 0}
								<div transition:slide={{ duration: 150 }}>
									{#each childData.children as grandchildPage (grandchildPage.id)}
										<SidebarPageItem
											page={grandchildPage}
											depth={2}
											isSelected={selectedPageId === grandchildPage.id}
											isExpanded={expandedPages.has(grandchildPage.id)}
											hasChildren={hasChildren(grandchildPage)}
											isFavorite={favorites.has(grandchildPage.id)}
											onSelect={() => onPageSelect(grandchildPage)}
											onExpand={() => onPageExpand?.(grandchildPage.id)}
											onAddChild={() => onPageAddChild?.(grandchildPage)}
											onOpenPeek={() => onPageOpenPeek?.(grandchildPage)}
											onDuplicate={() => onPageDuplicate?.(grandchildPage)}
											onRename={() => onPageRename?.(grandchildPage)}
											onMove={() => onPageMove?.(grandchildPage)}
											onDelete={() => onPageDelete?.(grandchildPage)}
											onToggleFavorite={() => onPageToggleFavorite?.(grandchildPage)}
											onCopyLink={() => onPageCopyLink?.(grandchildPage)}
										/>
									{/each}
								</div>
							{/if}
						{/each}
					</div>
				{/if}
			{/each}

			<!-- "More" link -->
			{#if hasMore && !showAll}
				<button
					onclick={() => showAll = true}
					class="w-full flex items-center gap-2 px-2 py-1.5 text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
				>
					<span class="w-5"></span>
					<span>... {pages.length - maxVisible} more</span>
				</button>
			{/if}

			<!-- Empty State -->
			{#if pages.length === 0}
				<div class="px-4 py-3 text-xs text-gray-400 dark:text-gray-500 italic">
					No pages yet
				</div>
			{/if}
		</div>
	{/if}
</div>
