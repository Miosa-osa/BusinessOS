<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { fly } from 'svelte/transition';

	type ViewMode = 'list' | 'board' | 'calendar';
	type GroupBy = 'status' | 'priority' | 'project' | 'assignee' | 'none';

	interface Props {
		view?: ViewMode;
		groupBy?: GroupBy;
		searchQuery?: string;
		onViewChange?: (view: ViewMode) => void;
		onGroupByChange?: (groupBy: GroupBy) => void;
		onSearchChange?: (query: string) => void;
		onFilterChange?: (filters: Record<string, string[]>) => void;
	}

	let {
		view = $bindable('list'),
		groupBy = $bindable('status'),
		searchQuery = $bindable(''),
		onViewChange,
		onGroupByChange,
		onSearchChange,
		onFilterChange
	}: Props = $props();

	let filterOpen = $state(false);
	let groupByOpen = $state(false);
	const dropdownTransitionProps = { transition: fly, transitionConfig: { y: -10, duration: 150 } } as any;

	const viewOptions: { value: ViewMode; label: string; icon: string }[] = [
		{
			value: 'list',
			label: 'List',
			icon: `<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
			</svg>`
		},
		{
			value: 'board',
			label: 'Board',
			icon: `<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
			</svg>`
		},
		{
			value: 'calendar',
			label: 'Calendar',
			icon: `<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
			</svg>`
		}
	];

	const groupByOptions: { value: GroupBy; label: string }[] = [
		{ value: 'none', label: 'None' },
		{ value: 'status', label: 'Status' },
		{ value: 'priority', label: 'Priority' },
		{ value: 'project', label: 'Project' },
		{ value: 'assignee', label: 'Assignee' }
	];

	function handleViewChange(newView: ViewMode) {
		view = newView;
		onViewChange?.(newView);
	}

	function handleGroupByChange(newGroupBy: GroupBy) {
		groupBy = newGroupBy;
		onGroupByChange?.(newGroupBy);
		groupByOpen = false;
	}

	function handleSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		searchQuery = target.value;
		onSearchChange?.(searchQuery);
	}
</script>

<div class="tb-toolbar flex flex-wrap items-center justify-between gap-3 px-4 sm:px-6 py-3">
	<!-- View Switcher -->
	<div class="flex items-center gap-1 tb-toolbar-seg rounded-lg p-1 flex-shrink-0">
		{#each viewOptions as option}
			<button
				onclick={() => handleViewChange(option.value)}
				class="flex items-center gap-1.5 sm:gap-2 px-2 sm:px-3 py-1.5 rounded-md text-sm transition-colors
					{view === option.value ? 'tb-toolbar-seg-active font-medium' : 'tb-toolbar-seg-inactive'}"
			>
				{@html option.icon}
				<span>{option.label}</span>
			</button>
		{/each}
	</div>

	<!-- Center Actions -->
	<div class="flex items-center gap-2 flex-shrink-0">
		<!-- Filter Dropdown -->
		<DropdownMenu.Root bind:open={filterOpen}>
			<DropdownMenu.Trigger
				class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
				</svg>
				Filter
				<svg class="w-3 h-3 tb-toolbar-chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content
					class="z-50 w-64 tb-toolbar-dropdown rounded-xl shadow-lg p-3"
					sideOffset={4}
					{...dropdownTransitionProps}
				>
					<div class="space-y-4">
						<!-- Status -->
						<div>
							<p class="text-xs font-medium tb-toolbar-meta uppercase mb-2">Status</p>
							<div class="space-y-1">
								{#each ['To Do', 'In Progress', 'In Review', 'Done', 'Blocked'] as status}
									<label class="flex items-center gap-2 px-2 py-1.5 tb-toolbar-dropdown-hover rounded cursor-pointer">
										<input type="checkbox" class="rounded tb-toolbar-check" checked />
										<span class="text-sm tb-toolbar-label">{status}</span>
									</label>
								{/each}
							</div>
						</div>

						<!-- Priority -->
						<div>
							<p class="text-xs font-medium tb-toolbar-meta uppercase mb-2">Priority</p>
							<div class="space-y-1">
								{#each ['Critical', 'High', 'Medium', 'Low'] as priority}
									<label class="flex items-center gap-2 px-2 py-1.5 tb-toolbar-dropdown-hover rounded cursor-pointer">
										<input type="checkbox" class="rounded tb-toolbar-check" checked />
										<span class="text-sm tb-toolbar-label">{priority}</span>
									</label>
								{/each}
							</div>
						</div>

						<div class="flex items-center justify-between pt-2 tb-toolbar-sep">
							<button class="btn-pill btn-pill-ghost btn-pill-xs">Clear All</button>
							<button class="btn-pill btn-pill-primary btn-pill-xs">
								Apply
							</button>
						</div>
					</div>
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Group By Dropdown -->
		{#if view === 'list'}
			<DropdownMenu.Root bind:open={groupByOpen}>
				<DropdownMenu.Trigger
					class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
					</svg>
					Group
					<svg class="w-3 h-3 tb-toolbar-chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
					</svg>
				</DropdownMenu.Trigger>
				<DropdownMenu.Portal>
					<DropdownMenu.Content
					class="z-50 min-w-[160px] tb-toolbar-dropdown rounded-xl shadow-lg p-1"
					sideOffset={4}
					{...dropdownTransitionProps}
				>
					{#each groupByOptions as option}
						<DropdownMenu.Item
							class="flex items-center gap-2 px-3 py-2 text-sm rounded-lg cursor-pointer transition-colors
								{groupBy === option.value ? 'tb-toolbar-group-active font-medium' : 'tb-toolbar-group-item'}"
							onclick={() => handleGroupByChange(option.value)}
						>
							{#if groupBy === option.value}
								<svg class="w-4 h-4 tb-toolbar-check-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
								{:else}
									<span class="w-4"></span>
								{/if}
								{option.label}
							</DropdownMenu.Item>
						{/each}
					</DropdownMenu.Content>
				</DropdownMenu.Portal>
			</DropdownMenu.Root>
		{/if}
	</div>

	<!-- Search -->
	<div class="relative flex-1 sm:flex-none min-w-0 sm:min-w-[200px]">
		<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 tb-toolbar-search-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
		</svg>
		<input
			type="text"
			placeholder="Search tasks..."
			value={searchQuery}
			oninput={handleSearchInput}
			class="w-full sm:w-64 pl-10 pr-4 py-2 text-sm tb-toolbar-input rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent transition-all"
		/>
	</div>
</div>

<style>
	.tb-toolbar {
		background: var(--dbg, #fff);
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}
	.tb-toolbar-seg {
		background: var(--dbg2, #f5f5f5);
	}
	.tb-toolbar-seg-active {
		background: var(--dbg, #fff);
		color: var(--dt, #111);
		box-shadow: var(--shadow-xs, 0 1px 2px rgba(0,0,0,0.05));
	}
	.tb-toolbar-seg-inactive {
		color: var(--dt2, #555);
	}
	.tb-toolbar-seg-inactive:hover {
		color: var(--dt, #111);
	}
	.tb-toolbar-dropdown {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
	}
	.tb-toolbar-dropdown-hover:hover {
		background: var(--dbg2, #f5f5f5);
	}
	.tb-toolbar-meta {
		color: var(--dt2, #555);
	}
	.tb-toolbar-label {
		color: var(--dt, #111);
	}
	.tb-toolbar-check {
		border-color: var(--dbd, #e0e0e0);
	}
	.tb-toolbar-chevron {
		color: var(--dt4, #bbb);
	}
	.tb-toolbar-sep {
		border-top: 1px solid var(--dbd2, #f0f0f0);
	}
	.tb-toolbar-group-active {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
	}
	.tb-toolbar-group-item {
		color: var(--dt2, #555);
	}
	.tb-toolbar-group-item:hover {
		background: var(--dbg2, #f5f5f5);
	}
	.tb-toolbar-check-icon {
		color: var(--dt, #111);
	}
	.tb-toolbar-search-icon {
		color: var(--dt4, #bbb);
	}
	.tb-toolbar-input {
		background: var(--dbg2, #f5f5f5);
		border: 1px solid var(--dbd, #e0e0e0);
		color: var(--dt, #111);
	}
</style>
