<script lang="ts">
	interface Props {
		title: string;
		count: number;
		color?: string;
		collapsed?: boolean;
		showAddButton?: boolean;
		onToggle?: () => void;
		onAdd?: () => void;
	}

	let {
		title,
		count,
		color = '#6B7280',
		collapsed = false,
		showAddButton = true,
		onToggle,
		onAdd
	}: Props = $props();
</script>

<div class="flex items-center justify-between py-2 px-4 bg-gray-50 border-b border-gray-200 sticky top-0 z-10">
	<button
		onclick={onToggle}
		class="flex items-center gap-2 text-sm font-medium text-gray-700 hover:text-gray-900 transition-colors"
	>
		<svg
			class="w-4 h-4 text-gray-400 transition-transform duration-200 {collapsed ? '-rotate-90' : ''}"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
		>
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
		</svg>
		<span class="w-2.5 h-2.5 rounded-full" style="background-color: {color}"></span>
		<span>{title}</span>
		<span class="text-gray-400 font-normal">({count})</span>
	</button>

	{#if showAddButton && !collapsed}
		<button
			onclick={onAdd}
			class="flex items-center gap-1 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-100 px-2 py-1 rounded-lg transition-colors"
		>
			<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
			Add
		</button>
	{/if}
</div>
