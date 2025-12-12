<script lang="ts">
	import { editor, blockTypes, type BlockType } from '$lib/stores/editor';

	let selectedIndex = $state(0);

	// Filter block types based on query
	let filteredTypes = $derived(
		$editor.slashMenuQuery
			? blockTypes.filter(bt =>
				bt.label.toLowerCase().includes($editor.slashMenuQuery.toLowerCase()) ||
				bt.type.toLowerCase().includes($editor.slashMenuQuery.toLowerCase())
			)
			: blockTypes
	);

	function selectBlockType(type: BlockType) {
		if ($editor.focusedBlockId) {
			editor.changeBlockType($editor.focusedBlockId, type);
			editor.updateBlock($editor.focusedBlockId, '');
		}
		editor.hideSlashMenu();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = Math.min(selectedIndex + 1, filteredTypes.length - 1);
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = Math.max(selectedIndex - 1, 0);
		} else if (e.key === 'Enter' || e.key === 'Tab') {
			e.preventDefault();
			if (filteredTypes[selectedIndex]) {
				selectBlockType(filteredTypes[selectedIndex].type);
			}
		} else if (e.key === 'Escape') {
			e.preventDefault();
			editor.hideSlashMenu();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if $editor.slashMenuPosition}
	<div
		class="fixed bg-white rounded-xl shadow-2xl border border-gray-200 z-50 overflow-hidden w-72"
		style="left: {$editor.slashMenuPosition.x}px; top: {$editor.slashMenuPosition.y}px;"
	>
		<div class="px-3 py-2 border-b border-gray-100">
			<p class="text-xs text-gray-400 uppercase tracking-wider font-medium">Basic blocks</p>
		</div>
		<div class="py-1 max-h-80 overflow-auto">
			{#each filteredTypes as blockType, idx}
				<button
					onclick={() => selectBlockType(blockType.type)}
					onmouseenter={() => selectedIndex = idx}
					class="w-full px-3 py-2.5 flex items-center gap-3 text-left transition-colors
						{idx === selectedIndex ? 'bg-gray-100' : 'hover:bg-gray-50'}"
				>
					<span class="w-10 h-10 rounded-lg bg-gray-50 border border-gray-200 flex items-center justify-center text-base font-medium text-gray-600">
						{blockType.icon}
					</span>
					<div class="flex-1 min-w-0">
						<div class="text-sm font-medium text-gray-900">{blockType.label}</div>
						<div class="text-xs text-gray-500">{blockType.description}</div>
					</div>
					{#if idx === selectedIndex}
						<kbd class="px-1.5 py-0.5 text-xs bg-gray-200 rounded text-gray-500">Enter</kbd>
					{/if}
				</button>
			{/each}
			{#if filteredTypes.length === 0}
				<div class="px-3 py-6 text-sm text-gray-400 text-center">
					No matching blocks found
				</div>
			{/if}
		</div>
	</div>
{/if}
