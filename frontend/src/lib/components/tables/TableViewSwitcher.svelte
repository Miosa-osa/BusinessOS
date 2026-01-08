<script lang="ts">
	/**
	 * TableViewSwitcher - Toggle between list and grid view
	 */
	import { List, LayoutGrid } from 'lucide-svelte';
	import type { TableViewMode } from '$lib/stores/tables';

	interface Props {
		viewMode: TableViewMode;
		onChange: (mode: TableViewMode) => void;
	}

	let { viewMode, onChange }: Props = $props();

	const views: { mode: TableViewMode; icon: typeof List; label: string }[] = [
		{ mode: 'list', icon: List, label: 'List view' },
		{ mode: 'grid', icon: LayoutGrid, label: 'Grid view' }
	];
</script>

<div class="flex items-center rounded-lg border border-gray-200 bg-white p-1">
	{#each views as view}
		<button
			type="button"
			class="rounded-md p-1.5 transition-colors {viewMode === view.mode
				? 'bg-gray-100 text-gray-900'
				: 'text-gray-400 hover:text-gray-600'}"
			title={view.label}
			onclick={() => onChange(view.mode)}
		>
			<view.icon class="h-5 w-5" />
		</button>
	{/each}
</div>
