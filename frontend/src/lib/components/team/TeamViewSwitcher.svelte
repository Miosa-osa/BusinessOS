<script lang="ts">
	type ViewMode = 'directory' | 'orgchart' | 'capacity';

	interface Props {
		view?: ViewMode;
		searchQuery?: string;
		onViewChange?: (view: ViewMode) => void;
		onSearchChange?: (query: string) => void;
	}

	let {
		view = $bindable('directory'),
		searchQuery = $bindable(''),
		onViewChange,
		onSearchChange
	}: Props = $props();

	const viewOptions: { value: ViewMode; label: string; icon: string }[] = [
		{
			value: 'directory',
			label: 'Directory',
			icon: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
			</svg>`
		},
		{
			value: 'orgchart',
			label: 'Org Chart',
			icon: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
			</svg>`
		},
		{
			value: 'capacity',
			label: 'Capacity',
			icon: `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
			</svg>`
		}
	];

	function handleViewChange(newView: ViewMode) {
		view = newView;
		onViewChange?.(newView);
	}

	function handleSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		searchQuery = target.value;
		onSearchChange?.(searchQuery);
	}
</script>

<div class="flex items-center justify-between px-6 py-3 border-b border-gray-200 bg-white">
	<!-- View Switcher -->
	<div class="flex items-center gap-1 bg-gray-100 rounded-lg p-1">
		{#each viewOptions as option}
			<button
				onclick={() => handleViewChange(option.value)}
				class="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm transition-colors
					{view === option.value ? 'bg-white shadow text-gray-900 font-medium' : 'text-gray-600 hover:text-gray-900'}"
			>
				{@html option.icon}
				<span>{option.label}</span>
			</button>
		{/each}
	</div>

	<!-- Search -->
	<div class="relative">
		<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
		</svg>
		<input
			type="text"
			placeholder="Search team..."
			value={searchQuery}
			oninput={handleSearchInput}
			class="w-64 pl-10 pr-4 py-2 text-sm bg-gray-50 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent transition-all"
		/>
	</div>
</div>
