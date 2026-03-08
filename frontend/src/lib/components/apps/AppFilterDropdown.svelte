<script lang="ts">
	import type { AppStatus } from '$lib/types/apps';

	type FilterValue = AppStatus | 'all';

	interface Props {
		value: FilterValue;
		onChange: (value: FilterValue) => void;
	}

	let { value = 'all', onChange }: Props = $props();

	let isOpen = $state(false);

	const options: { value: FilterValue; label: string }[] = [
		{ value: 'all', label: 'All Apps' },
		{ value: 'active', label: 'Active' },
		{ value: 'draft', label: 'Draft' },
		{ value: 'generating', label: 'Generating' },
		{ value: 'archived', label: 'Archived' }
	];

	const selectedLabel = $derived(options.find((o) => o.value === value)?.label || 'All Apps');

	function handleSelect(newValue: FilterValue) {
		onChange(newValue);
		isOpen = false;
	}

	function handleClickOutside(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (!target.closest('.filter-dropdown')) {
			isOpen = false;
		}
	}
</script>

<svelte:window onclick={handleClickOutside} />

<div class="relative filter-dropdown">
	<!-- Trigger -->
	<button
		onclick={() => (isOpen = !isOpen)}
		aria-label="Filter apps by status"
		aria-expanded={isOpen}
		aria-haspopup="listbox"
		class="inline-flex items-center gap-2 h-10 px-4 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700
			rounded-xl text-sm font-medium text-gray-700 dark:text-gray-300
			transition-all duration-150
			hover:border-gray-300 dark:hover:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-750
			{isOpen ? 'border-gray-300 dark:border-gray-600 shadow-sm' : ''}"
	>
		{selectedLabel}
		<svg
			class="w-4 h-4 text-gray-500 transition-transform duration-150 {isOpen ? 'rotate-180' : ''}"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
		</svg>
	</button>

	<!-- Dropdown -->
	{#if isOpen}
		<div
			class="absolute right-0 mt-2 w-44 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700
				rounded-xl shadow-lg shadow-gray-200/50 dark:shadow-gray-900/30 overflow-hidden z-50
				animate-in fade-in-0 zoom-in-95 duration-150"
			role="listbox"
			aria-label="App status filter options"
		>
			{#each options as option}
				<button
					onclick={() => handleSelect(option.value)}
					role="option"
					aria-selected={value === option.value}
					class="w-full flex items-center gap-2 px-4 py-2.5 text-sm text-left transition-colors
						{value === option.value
						? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white font-medium'
						: 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-750'}"
				>
					{#if value === option.value}
						<svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
						</svg>
					{:else}
						<span class="w-4"></span>
					{/if}
					{option.label}
				</button>
			{/each}
		</div>
	{/if}
</div>
