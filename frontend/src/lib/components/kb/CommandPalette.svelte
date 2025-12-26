<script lang="ts">
	import { slide, fade } from 'svelte/transition';
	import type { ContextListItem } from '$lib/api/client';

	interface Props {
		pages: ContextListItem[];
		isOpen: boolean;
		onClose: () => void;
		onSelect: (page: ContextListItem) => void;
	}

	let { pages, isOpen, onClose, onSelect }: Props = $props();

	let query = $state('');
	let selectedIndex = $state(0);
	let inputElement: HTMLInputElement | null = $state(null);

	// Filter results based on query
	const results = $derived.by(() => {
		if (!query.trim()) {
			// Show recent pages when no query
			return pages.slice(0, 10);
		}
		const q = query.toLowerCase();
		return pages.filter(p =>
			p.name?.toLowerCase().includes(q) ||
			p.type?.toLowerCase().includes(q)
		).slice(0, 20);
	});

	// Group results by date
	const groupedResults = $derived.by(() => {
		const items = results;
		const groups: { label: string; pages: ContextListItem[] }[] = [];

		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const yesterday = new Date(today);
		yesterday.setDate(yesterday.getDate() - 1);
		const lastWeek = new Date(today);
		lastWeek.setDate(lastWeek.getDate() - 7);

		const todayItems: ContextListItem[] = [];
		const yesterdayItems: ContextListItem[] = [];
		const lastWeekItems: ContextListItem[] = [];
		const olderItems: ContextListItem[] = [];

		for (const page of items) {
			const updated = page.updated_at ? new Date(page.updated_at) : null;
			if (!updated) {
				olderItems.push(page);
			} else if (updated >= today) {
				todayItems.push(page);
			} else if (updated >= yesterday) {
				yesterdayItems.push(page);
			} else if (updated >= lastWeek) {
				lastWeekItems.push(page);
			} else {
				olderItems.push(page);
			}
		}

		if (todayItems.length > 0) groups.push({ label: 'Today', pages: todayItems });
		if (yesterdayItems.length > 0) groups.push({ label: 'Yesterday', pages: yesterdayItems });
		if (lastWeekItems.length > 0) groups.push({ label: 'Last 7 days', pages: lastWeekItems });
		if (olderItems.length > 0) groups.push({ label: 'Older', pages: olderItems });

		return groups;
	});

	// Flat list of all results for keyboard navigation
	const flatResults = $derived(groupedResults.flatMap(g => g.pages));

	// Focus input when opened
	$effect(() => {
		if (isOpen && inputElement) {
			setTimeout(() => inputElement?.focus(), 50);
		}
	});

	// Reset state when opened/closed
	$effect(() => {
		if (isOpen) {
			query = '';
			selectedIndex = 0;
		}
	});

	function handleKeydown(e: KeyboardEvent) {
		const items = flatResults;

		switch (e.key) {
			case 'Escape':
				e.preventDefault();
				onClose();
				break;
			case 'ArrowDown':
				e.preventDefault();
				selectedIndex = (selectedIndex + 1) % items.length;
				break;
			case 'ArrowUp':
				e.preventDefault();
				selectedIndex = (selectedIndex - 1 + items.length) % items.length;
				break;
			case 'Enter':
				e.preventDefault();
				if (items[selectedIndex]) {
					onSelect(items[selectedIndex]);
					onClose();
				}
				break;
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			onClose();
		}
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

	function getTypeIcon(type: string): string {
		switch (type) {
			case 'project':
				return 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z';
			case 'person':
				return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
			case 'business':
				return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
			default:
				return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
		}
	}
</script>

<svelte:window onkeydown={(e) => {
	// Global Cmd+K handler
	if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
		e.preventDefault();
		if (!isOpen) {
			// This won't work here, but the parent should handle opening
		}
	}
}} />

{#if isOpen}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
		transition:fade={{ duration: 150 }}
		onclick={handleBackdropClick}
		onkeydown={handleKeydown}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<!-- Modal -->
		<div
			class="fixed top-[15%] left-1/2 -translate-x-1/2 w-full max-w-2xl bg-white dark:bg-[#2c2c2e] rounded-xl shadow-2xl overflow-hidden"
			transition:slide={{ duration: 200 }}
		>
			<!-- Search Input -->
			<div class="flex items-center gap-3 px-4 py-3 border-b border-gray-200 dark:border-gray-700">
				<svg class="w-5 h-5 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
				<input
					bind:this={inputElement}
					type="text"
					bind:value={query}
					placeholder="Search pages..."
					class="flex-1 bg-transparent border-0 outline-none text-gray-900 dark:text-gray-100 placeholder:text-gray-400 dark:placeholder:text-gray-500 text-base"
					onkeydown={handleKeydown}
				/>
				<kbd class="px-2 py-0.5 text-xs text-gray-400 bg-gray-100 dark:bg-gray-700 rounded">esc</kbd>
			</div>

			<!-- Results -->
			<div class="max-h-[400px] overflow-y-auto">
				{#if flatResults.length === 0}
					<div class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
						{#if query}
							<p>No results found for "{query}"</p>
						{:else}
							<p>No pages yet</p>
						{/if}
					</div>
				{:else}
					{#each groupedResults as group}
						<div class="py-1">
							<div class="px-4 py-1.5 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
								{group.label}
							</div>
							{#each group.pages as page, i}
								{@const globalIndex = flatResults.indexOf(page)}
								<button
									onclick={() => { onSelect(page); onClose(); }}
									class="w-full flex items-center gap-3 px-4 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors
										{globalIndex === selectedIndex ? 'bg-blue-50 dark:bg-blue-900/30' : ''}"
								>
									<!-- Icon -->
									<span class="w-6 h-6 flex items-center justify-center flex-shrink-0">
										{#if page.icon && getIconPath(page.icon)}
											<svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getIconPath(page.icon)} />
											</svg>
										{:else}
											<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getTypeIcon(page.type || 'document')} />
											</svg>
										{/if}
									</span>

									<!-- Title & Type -->
									<div class="flex-1 min-w-0">
										<div class="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
											{page.name || 'New page'}
										</div>
										{#if page.type && page.type !== 'document'}
											<div class="text-xs text-gray-500 dark:text-gray-400 capitalize">
												{page.type}
											</div>
										{/if}
									</div>

									<!-- Selection indicator -->
									{#if globalIndex === selectedIndex}
										<svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
										</svg>
									{/if}
								</button>
							{/each}
						</div>
					{/each}
				{/if}
			</div>

			<!-- Footer with keyboard shortcuts -->
			<div class="px-4 py-2 border-t border-gray-200 dark:border-gray-700 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
				<span class="flex items-center gap-1">
					<kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">↑</kbd>
					<kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">↓</kbd>
					<span>to navigate</span>
				</span>
				<span class="flex items-center gap-1">
					<kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">↵</kbd>
					<span>to open</span>
				</span>
				<span class="flex items-center gap-1">
					<kbd class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">esc</kbd>
					<span>to close</span>
				</span>
			</div>
		</div>
	</div>
{/if}
