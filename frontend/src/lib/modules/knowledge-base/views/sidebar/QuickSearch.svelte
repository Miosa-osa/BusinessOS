<script lang="ts">
	import { Modal, Input, ScrollArea, Separator } from '$lib/ui';
	import { Search, FileText, Clock, ArrowRight, Hash } from 'lucide-svelte';
	import { searchDocuments } from '../../services/documents.service';
	import { recentDocuments } from '../../stores/documents';
	import type { DocumentMeta } from '../../entities/types';
	import { debounce } from '$lib/utils';

	interface Props {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		onSelectDocument?: (id: string) => void;
	}

	let { open = $bindable(false), onOpenChange, onSelectDocument }: Props = $props();

	let searchQuery = $state('');
	let searchResults = $state<DocumentMeta[]>([]);
	let isSearching = $state(false);
	let selectedIndex = $state(0);

	let recent = $derived($recentDocuments);

	// Displayed items (recent when no query, search results when querying)
	let displayedItems = $derived(searchQuery.trim() ? searchResults : recent.slice(0, 5));

	const debouncedSearch = debounce(async (query: string) => {
		if (!query.trim()) {
			searchResults = [];
			isSearching = false;
			return;
		}

		isSearching = true;
		try {
			searchResults = await searchDocuments(query);
		} catch {
			searchResults = [];
		} finally {
			isSearching = false;
		}
	}, 200);

	function handleSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		searchQuery = target.value;
		selectedIndex = 0;
		debouncedSearch(searchQuery);
	}

	function handleKeydown(e: KeyboardEvent) {
		const itemCount = displayedItems.length;
		if (itemCount === 0) return;

		switch (e.key) {
			case 'ArrowDown':
				e.preventDefault();
				selectedIndex = (selectedIndex + 1) % itemCount;
				break;
			case 'ArrowUp':
				e.preventDefault();
				selectedIndex = (selectedIndex - 1 + itemCount) % itemCount;
				break;
			case 'Enter':
				e.preventDefault();
				if (displayedItems[selectedIndex]) {
					selectDocument(displayedItems[selectedIndex].id);
				}
				break;
			case 'Escape':
				e.preventDefault();
				closeSearch();
				break;
		}
	}

	function selectDocument(id: string) {
		onSelectDocument?.(id);
		closeSearch();
	}

	function closeSearch() {
		open = false;
		searchQuery = '';
		searchResults = [];
		selectedIndex = 0;
		onOpenChange?.(false);
	}

	function handleOpenChange(isOpen: boolean) {
		open = isOpen;
		if (!isOpen) {
			searchQuery = '';
			searchResults = [];
			selectedIndex = 0;
		}
		onOpenChange?.(isOpen);
	}

	// Global keyboard shortcut
	function handleGlobalKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
			e.preventDefault();
			open = !open;
		}
	}
</script>

<svelte:window onkeydown={handleGlobalKeydown} />

<Modal
	bind:open
	onOpenChange={handleOpenChange}
	size="default"
	showClose={false}
	class="quick-search-modal"
>
	<div class="quick-search">
		<div class="quick-search__input-wrapper">
			<Search class="quick-search__search-icon h-4 w-4" />
			<input
				type="text"
				class="quick-search__input"
				placeholder="Search pages..."
				value={searchQuery}
				oninput={handleSearchInput}
				onkeydown={handleKeydown}
				autofocus
			/>
			{#if isSearching}
				<div class="quick-search__spinner"></div>
			{/if}
		</div>

		<Separator />

		<ScrollArea class="quick-search__results">
			{#if displayedItems.length > 0}
				<div class="quick-search__section">
					<span class="quick-search__section-title">
						{searchQuery.trim() ? 'Search Results' : 'Recent'}
					</span>
				</div>
				{#each displayedItems as item, index}
					<button
						class="quick-search__item"
						class:quick-search__item--selected={index === selectedIndex}
						onclick={() => selectDocument(item.id)}
						onmouseenter={() => (selectedIndex = index)}
					>
						<span class="quick-search__item-icon">
							{#if item.icon && typeof item.icon === 'string'}
								{item.icon}
							{:else}
								<FileText class="h-4 w-4" />
							{/if}
						</span>
						<span class="quick-search__item-title">
							{item.title || 'Untitled'}
						</span>
						{#if index === selectedIndex}
							<ArrowRight class="quick-search__item-arrow h-3 w-3" />
						{/if}
					</button>
				{/each}
			{:else if searchQuery.trim() && !isSearching}
				<div class="quick-search__empty">
					<p>No results found for "{searchQuery}"</p>
				</div>
			{:else if !searchQuery.trim()}
				<div class="quick-search__empty">
					<Clock class="h-8 w-8 opacity-50" />
					<p>No recent pages</p>
				</div>
			{/if}
		</ScrollArea>

		<Separator />

		<div class="quick-search__footer">
			<div class="quick-search__hint">
				<kbd>↑↓</kbd> navigate
				<kbd>↵</kbd> select
				<kbd>esc</kbd> close
			</div>
		</div>
	</div>
</Modal>

<style>
	.quick-search {
		display: flex;
		flex-direction: column;
		margin: -1.5rem;
	}

	.quick-search__input-wrapper {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem;
	}

	.quick-search__search-icon {
		color: hsl(var(--muted-foreground));
	}

	.quick-search__input {
		flex: 1;
		background: transparent;
		border: none;
		outline: none;
		font-size: 1rem;
		color: hsl(var(--foreground));
	}

	.quick-search__input::placeholder {
		color: hsl(var(--muted-foreground));
	}

	.quick-search__spinner {
		width: 16px;
		height: 16px;
		border: 2px solid hsl(var(--muted-foreground) / 0.3);
		border-top-color: hsl(var(--muted-foreground));
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.quick-search__results {
		max-height: 300px;
		padding: 0.5rem;
	}

	.quick-search__section {
		padding: 0.5rem 0.75rem;
	}

	.quick-search__section-title {
		font-size: 0.75rem;
		font-weight: 500;
		color: hsl(var(--muted-foreground));
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.quick-search__item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		width: 100%;
		padding: 0.625rem 0.75rem;
		background: transparent;
		border: none;
		border-radius: 0.375rem;
		cursor: pointer;
		text-align: left;
		transition: background-color 0.1s;
	}

	.quick-search__item:hover,
	.quick-search__item--selected {
		background-color: hsl(var(--accent));
	}

	.quick-search__item-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		font-size: 14px;
		color: hsl(var(--muted-foreground));
	}

	.quick-search__item-title {
		flex: 1;
		font-size: 0.875rem;
		color: hsl(var(--foreground));
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.quick-search__item-arrow {
		color: hsl(var(--muted-foreground));
	}

	.quick-search__empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 2rem;
		color: hsl(var(--muted-foreground));
		text-align: center;
	}

	.quick-search__footer {
		display: flex;
		justify-content: flex-end;
		padding: 0.75rem 1rem;
	}

	.quick-search__hint {
		display: flex;
		gap: 1rem;
		font-size: 0.75rem;
		color: hsl(var(--muted-foreground));
	}

	.quick-search__hint kbd {
		display: inline-flex;
		align-items: center;
		padding: 0.125rem 0.375rem;
		margin-right: 0.25rem;
		border-radius: 0.25rem;
		background-color: hsl(var(--muted));
		font-family: ui-monospace, monospace;
		font-size: 0.625rem;
	}
</style>
