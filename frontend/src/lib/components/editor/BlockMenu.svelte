<script lang="ts">
	import { editor, blockTypes, filterBlockTypes, getBlockTypesBySection, type BlockType, type BlockTypeDefinition } from '$lib/stores/editor';
	import { fly } from 'svelte/transition';

	let selectedIndex = $state(0);

	// Filter block types based on query using priority-based filtering
	let filteredTypes = $derived(
		$editor.slashMenuQuery
			? filterBlockTypes($editor.slashMenuQuery)
			: blockTypes
	);

	// Get sections for display (only when not filtering)
	let sections = $derived(
		$editor.slashMenuQuery
			? null
			: getBlockTypesBySection(filteredTypes)
	);

	// Flat list for keyboard navigation
	let flatList = $derived(
		$editor.slashMenuQuery
			? filteredTypes
			: [...(sections?.suggested || []), ...(sections?.basic || [])]
	);

	// Reset selection when query changes
	$effect(() => {
		if ($editor.slashMenuQuery !== undefined) {
			selectedIndex = 0;
		}
	});

	function selectBlockType(type: BlockType) {
		// Use the store's selectBlockType which sets pendingBlockTypeSelection
		// Block.svelte will watch this and handle the actual selection (including page creation)
		editor.selectBlockType(type);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!$editor.showSlashMenu) return;

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, flatList.length - 1);
			scrollSelectedIntoView();
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, 0);
			scrollSelectedIntoView();
		} else if (e.key === 'Enter' || e.key === 'Tab') {
			e.preventDefault();
			if (flatList[selectedIndex]) {
				selectBlockType(flatList[selectedIndex].type);
			}
		} else if (e.key === 'Escape') {
			e.preventDefault();
			editor.hideSlashMenu();
		}
	}

	function scrollSelectedIntoView() {
		// Use requestAnimationFrame to wait for DOM update
		requestAnimationFrame(() => {
			const menuEl = document.querySelector('[data-slash-menu]');
			const selectedEl = menuEl?.querySelector('.menu-item.bg-gray-100');
			if (selectedEl) {
				selectedEl.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
			}
		});
	}

	// Icon mapping to SVG paths
	function getIconSvg(iconName: string): string {
		const icons: Record<string, string> = {
			'file-text': 'M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z M14 2v6h6 M16 13H8 M16 17H8 M10 9H8',
			'minus': 'M5 12h14',
			'alert-circle': 'M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z M12 8v4 M12 16h.01',
			'type': 'M4 7V4h16v3 M9 20h6 M12 4v16',
			'heading-1': 'M4 12h8 M4 18V6 M12 18V6 M17 10v8 M17 10l3-2',
			'heading-2': 'M4 12h8 M4 18V6 M12 18V6 M21 18h-4c0-4 4-3 4-6 0-1.5-2-2.5-4-1',
			'heading-3': 'M4 12h8 M4 18V6 M12 18V6 M17.5 10.5c1.7-1 3.5 0 3.5 1.5a2 2 0 0 1-2 2 M17.5 17.5c1.7 1 3.5 0 3.5-1.5a2 2 0 0 0-2-2',
			'list': 'M8 6h13 M8 12h13 M8 18h13 M3 6h.01 M3 12h.01 M3 18h.01',
			'list-ordered': 'M10 6h11 M10 12h11 M10 18h11 M4 6h1v4 M4 10h2 M6 18H4c0-1 2-2 2-3s-1-1.5-2-1',
			'check-square': 'M9 11l3 3L22 4 M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11',
			'chevron-right': 'M9 18l6-6-6-6',
			'quote': 'M3 21c3 0 7-1 7-8V5c0-1.25-.756-2.017-2-2H4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2 1 0 1 0 1 1v1c0 1-1 2-2 2s-1 .008-1 1.031V21z M15 21c3 0 7-1 7-8V5c0-1.25-.757-2.017-2-2h-4c-1.25 0-2 .75-2 1.972V11c0 1.25.75 2 2 2h.75c0 2.25.25 4-2.75 4v3z',
			'code': 'M16 18l6-6-6-6 M8 6l-6 6 6 6',
			'columns': 'M9 4H5a1 1 0 0 0-1 1v14a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1V5a1 1 0 0 0-1-1z M19 4h-4a1 1 0 0 0-1 1v14a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1V5a1 1 0 0 0-1-1z',
			'link': 'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71 M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71'
		};
		return icons[iconName] || icons['type'];
	}

	function getGlobalIndex(section: 'suggested' | 'basic', localIndex: number): number {
		if (section === 'suggested') return localIndex;
		return (sections?.suggested?.length || 0) + localIndex;
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if $editor.slashMenuPosition}
	<div
		data-slash-menu
		class="slash-menu fixed z-50 w-80 bg-white dark:bg-[#252525] rounded-xl shadow-2xl border border-gray-200 dark:border-[#3d3d3d] overflow-hidden"
		style="left: {$editor.slashMenuPosition.x}px; top: {$editor.slashMenuPosition.y}px;"
		transition:fly={{ y: -8, duration: 150 }}
	>
		<!-- Menu Content -->
		<div class="max-h-96 overflow-y-auto">
			{#if !$editor.slashMenuQuery && sections}
				<!-- SUGGESTED SECTION -->
				{#if sections.suggested.length > 0}
					<div class="px-3 pt-3 pb-1.5">
						<p class="text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Suggested</p>
					</div>
					{#each sections.suggested as blockType, idx}
						{@const globalIdx = getGlobalIndex('suggested', idx)}
						<button
							onclick={() => selectBlockType(blockType.type)}
							onmouseenter={() => selectedIndex = globalIdx}
							class="menu-item w-full px-3 py-2 flex items-center gap-3 text-left transition-colors
								{globalIdx === selectedIndex ? 'bg-gray-100 dark:bg-[#3d3d3d]' : 'hover:bg-gray-50 dark:hover:bg-[#2f2f2f]'}"
						>
							<div class="w-10 h-10 rounded-lg bg-gray-100 dark:bg-[#2f2f2f] border border-gray-200 dark:border-[#3d3d3d] flex items-center justify-center text-gray-500 dark:text-gray-400">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" d={getIconSvg(blockType.icon)} />
								</svg>
							</div>
							<div class="flex-1 min-w-0">
								<div class="text-sm font-medium text-gray-800 dark:text-gray-200">{blockType.label}</div>
								<div class="text-xs text-gray-500">{blockType.description}</div>
							</div>
							{#if blockType.keyboardShortcut}
								<kbd class="px-1.5 py-0.5 text-[10px] font-mono bg-gray-100 dark:bg-[#3d3d3d] border border-gray-200 dark:border-[#4d4d4d] rounded text-gray-500 dark:text-gray-400">
									{blockType.keyboardShortcut}
								</kbd>
							{/if}
						</button>
					{/each}
				{/if}

				<!-- BASIC BLOCKS SECTION -->
				{#if sections.basic.length > 0}
					<div class="px-3 pt-3 pb-1.5 {sections.suggested.length > 0 ? 'border-t border-gray-200 dark:border-[#3d3d3d] mt-1' : ''}">
						<p class="text-[11px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Basic blocks</p>
					</div>
					{#each sections.basic as blockType, idx}
						{@const globalIdx = getGlobalIndex('basic', idx)}
						<button
							onclick={() => selectBlockType(blockType.type)}
							onmouseenter={() => selectedIndex = globalIdx}
							class="menu-item w-full px-3 py-2 flex items-center gap-3 text-left transition-colors
								{globalIdx === selectedIndex ? 'bg-gray-100 dark:bg-[#3d3d3d]' : 'hover:bg-gray-50 dark:hover:bg-[#2f2f2f]'}"
						>
							<div class="w-10 h-10 rounded-lg bg-gray-100 dark:bg-[#2f2f2f] border border-gray-200 dark:border-[#3d3d3d] flex items-center justify-center text-gray-500 dark:text-gray-400">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" d={getIconSvg(blockType.icon)} />
								</svg>
							</div>
							<div class="flex-1 min-w-0">
								<div class="text-sm font-medium text-gray-800 dark:text-gray-200">{blockType.label}</div>
								<div class="text-xs text-gray-500">{blockType.description}</div>
							</div>
							{#if blockType.keyboardShortcut}
								<kbd class="px-1.5 py-0.5 text-[10px] font-mono bg-gray-100 dark:bg-[#3d3d3d] border border-gray-200 dark:border-[#4d4d4d] rounded text-gray-500 dark:text-gray-400">
									{blockType.keyboardShortcut}
								</kbd>
							{/if}
						</button>
					{/each}
				{/if}
			{:else}
				<!-- FILTERED RESULTS (flat list) -->
				{#if filteredTypes.length > 0}
					{#each filteredTypes as blockType, idx}
						<button
							onclick={() => selectBlockType(blockType.type)}
							onmouseenter={() => selectedIndex = idx}
							class="menu-item w-full px-3 py-2 flex items-center gap-3 text-left transition-colors
								{idx === selectedIndex ? 'bg-gray-100 dark:bg-[#3d3d3d]' : 'hover:bg-gray-50 dark:hover:bg-[#2f2f2f]'}"
						>
							<div class="w-10 h-10 rounded-lg bg-gray-100 dark:bg-[#2f2f2f] border border-gray-200 dark:border-[#3d3d3d] flex items-center justify-center text-gray-500 dark:text-gray-400">
								<svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" d={getIconSvg(blockType.icon)} />
								</svg>
							</div>
							<div class="flex-1 min-w-0">
								<div class="text-sm font-medium text-gray-800 dark:text-gray-200">{blockType.label}</div>
								<div class="text-xs text-gray-500">{blockType.description}</div>
							</div>
							{#if blockType.keyboardShortcut}
								<kbd class="px-1.5 py-0.5 text-[10px] font-mono bg-gray-100 dark:bg-[#3d3d3d] border border-gray-200 dark:border-[#4d4d4d] rounded text-gray-500 dark:text-gray-400">
									{blockType.keyboardShortcut}
								</kbd>
							{/if}
						</button>
					{/each}
				{:else}
					<!-- Empty state -->
					<div class="px-4 py-8 text-center">
						<p class="text-sm text-gray-500">No blocks found</p>
						<p class="text-xs text-gray-400 dark:text-gray-600 mt-1">Try a different search term</p>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Filter Footer -->
		<div class="px-3 py-2 border-t border-gray-200 dark:border-[#3d3d3d] bg-gray-50 dark:bg-[#2a2a2a] flex items-center gap-2">
			<span class="text-gray-400 dark:text-gray-500 text-sm">/</span>
			<span class="flex-1 text-sm {$editor.slashMenuQuery ? 'text-gray-700 dark:text-gray-300' : 'text-gray-400 dark:text-gray-500'}">
				{$editor.slashMenuQuery || 'Filter...'}
			</span>
			<kbd class="px-1.5 py-0.5 text-[10px] font-mono bg-gray-100 dark:bg-[#3d3d3d] rounded text-gray-500 dark:text-gray-400">esc</kbd>
		</div>
	</div>
{/if}

<style>
	.slash-menu {
		/* Subtle shadow for light theme, darker for dark theme */
		box-shadow:
			0 0 0 1px rgba(0, 0, 0, 0.05),
			0 4px 8px rgba(0, 0, 0, 0.08),
			0 16px 24px rgba(0, 0, 0, 0.1),
			0 24px 32px rgba(0, 0, 0, 0.08);
	}

	:global(.dark) .slash-menu {
		box-shadow:
			0 0 0 1px rgba(0, 0, 0, 0.2),
			0 4px 8px rgba(0, 0, 0, 0.15),
			0 16px 24px rgba(0, 0, 0, 0.2),
			0 24px 32px rgba(0, 0, 0, 0.15);
	}

	/* Smooth scrollbar - light theme */
	.slash-menu > div:first-child::-webkit-scrollbar {
		width: 6px;
	}

	.slash-menu > div:first-child::-webkit-scrollbar-track {
		background: transparent;
	}

	.slash-menu > div:first-child::-webkit-scrollbar-thumb {
		background: #d1d5db;
		border-radius: 3px;
	}

	.slash-menu > div:first-child::-webkit-scrollbar-thumb:hover {
		background: #9ca3af;
	}

	/* Dark theme scrollbar */
	:global(.dark) .slash-menu > div:first-child::-webkit-scrollbar-thumb {
		background: #4d4d4d;
	}

	:global(.dark) .slash-menu > div:first-child::-webkit-scrollbar-thumb:hover {
		background: #5d5d5d;
	}
</style>
