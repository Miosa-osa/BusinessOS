<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { projects } from '$lib/stores/projects';
	import { moduleEvents } from '$lib/stores/events';
	import { api, type ClientListResponse } from '$lib/api';
	import { Dialog, Popover, DropdownMenu } from 'bits-ui';
	import { X, FolderPlus, Briefcase, Users, GraduationCap, ChevronDown, ChevronRight, AlertCircle, Search, Building2, BookOpen, FolderOpen, LayoutGrid, List, Columns3, Plus } from 'lucide-svelte';
	import type { Project } from '$lib/api';
	import { getTypeLabel, formatDate } from '$lib/utils/project';

	// Preloaded data from +page.ts load function (prefetched on hover)
	let { data } = $props();

	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	let showNewProject = $state(false);
	let newProject = $state({
		name: '',
		description: '',
		client_name: '',
		project_type: 'internal',
		priority: 'medium' as 'low' | 'medium' | 'high' | 'critical',
		icon: ''
	});
	let statusFilter = $state('');
	let typeFilter = $state('');
	let priorityFilter = $state('');
	let searchQuery = $state('');
	let viewMode = $state<'grid' | 'list' | 'kanban'>('grid');
	let groupByType = $state(false);
	let createError = $state('');

	$effect(() => {
		localStorage.setItem('bos-projects-view', viewMode);
	});

	let sortField = $state<'name' | 'type' | 'status' | 'priority' | 'updated'>('updated');
	let sortDir = $state<'asc' | 'desc'>('desc');

	// Clients for dropdown — seed from preloaded data
	let clients = $state<ClientListResponse[]>(data?.clients ?? []);
	let showAdvancedOptions = $state(false);

	onMount(async () => {
		const savedView = localStorage.getItem('bos-projects-view');
		if (savedView === 'grid' || savedView === 'list' || savedView === 'kanban') viewMode = savedView;

		// Seed the store with preloaded data if available (avoids redundant fetch)
		if (data?.projects?.length) {
			projects.setProjects(data.projects);
		} else {
			try {
				await projects.loadProjects();
			} catch {
				// Backend unavailable — empty state will show
			}
		}

		// Clients already seeded from preloaded data; refresh if empty
		if (!clients.length) {
			await loadClients();
		}
	});

	async function loadClients() {
		try {
			clients = await api.getClients();
		} catch (err) {
			console.error('Error loading clients:', err);
		}
	}

	// Reload when a project is created from the chat slash command
	$effect(() => {
		const event = $moduleEvents;
		if (event?.type === 'project:created') {
			projects.loadProjects();
		}
	});

	// Filtered projects
	let filteredProjects = $derived(() => {
		let result = $projects.projects;

		if (typeFilter) {
			result = result.filter(p => p.project_type === typeFilter);
		}

		if (priorityFilter) {
			result = result.filter(p => p.priority === priorityFilter);
		}

		if (searchQuery) {
			const query = searchQuery.toLowerCase();
			result = result.filter(p =>
				p.name.toLowerCase().includes(query) ||
				(p.description && p.description.toLowerCase().includes(query)) ||
				(p.client_name && p.client_name.toLowerCase().includes(query))
			);
		}

		if (viewMode === 'list') {
			const priorityOrder: Record<string, number> = { critical: 0, high: 1, medium: 2, low: 3 };
			result = [...result].sort((a, b) => {
				let cmp = 0;
				if (sortField === 'name') {
					cmp = a.name.localeCompare(b.name);
				} else if (sortField === 'type') {
					cmp = a.project_type.localeCompare(b.project_type);
				} else if (sortField === 'status') {
					cmp = a.status.localeCompare(b.status);
				} else if (sortField === 'priority') {
					cmp = (priorityOrder[a.priority] ?? 4) - (priorityOrder[b.priority] ?? 4);
				} else {
					cmp = new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
				}
				return sortDir === 'asc' ? cmp : -cmp;
			});
		}

		return result;
	});

	// Stats
	let stats = $derived({
		total: $projects.projects.length,
		active: $projects.projects.filter(p => p.status === 'active').length,
		paused: $projects.projects.filter(p => p.status === 'paused').length,
		completed: $projects.projects.filter(p => p.status === 'completed').length
	});

	// Grouped by type
	let groupedByType = $derived({
		internal: filteredProjects().filter(p => p.project_type === 'internal'),
		client_work: filteredProjects().filter(p => p.project_type === 'client_work'),
		learning: filteredProjects().filter(p => p.project_type === 'learning'),
		other: filteredProjects().filter(p => !['internal', 'client_work', 'learning'].includes(p.project_type))
	});

	// Grouped by status for kanban
	let groupedByStatus = $derived({
		active: filteredProjects().filter(p => p.status === 'active'),
		paused: filteredProjects().filter(p => p.status === 'paused'),
		completed: filteredProjects().filter(p => p.status === 'completed')
	});

	async function handleCreateProject(e: Event) {
		e.preventDefault();
		createError = '';
		try {
			await projects.createProject(newProject);
			showNewProject = false;
			newProject = { name: '', description: '', client_name: '', project_type: 'internal', priority: 'medium', icon: '' };
			showAdvancedOptions = false;
		} catch (error) {
			createError = (error as Error).message || 'Failed to create project';
		}
	}

	function getPriorityDot(priority: string): string {
		switch (priority) {
			case 'critical': return 'var(--bos-priority-critical)';
			case 'high': return 'var(--bos-priority-high)';
			case 'medium': return 'var(--bos-priority-medium)';
			default: return 'var(--bos-priority-low)';
		}
	}

	function clearFilters() {
		statusFilter = '';
		typeFilter = '';
		priorityFilter = '';
		searchQuery = '';
		projects.loadProjects();
	}

	// Type and Priority filter options
	const typeOptions = [
		{ value: '', label: 'All Types' },
		{ value: 'internal', label: 'Internal' },
		{ value: 'client_work', label: 'Client Work' },
		{ value: 'learning', label: 'Learning' }
	];

	const priorityOptions = [
		{ value: '', label: 'All Priorities' },
		{ value: 'critical', label: 'Critical' },
		{ value: 'high', label: 'High' },
		{ value: 'medium', label: 'Medium' },
		{ value: 'low', label: 'Low' }
	];

	let hasActiveFilters = $derived(statusFilter || typeFilter || priorityFilter || searchQuery);
</script>

<div class="h-full flex flex-col prm-ls-page">
	<!-- Header -->
	<div class="px-6 py-4 prm-ls-bar">
		<div class="flex items-center justify-between mb-4">
			<div>
				<h1 class="text-xl font-semibold prm-ls-title">Projects</h1>
				<p class="text-sm prm-ls-muted mt-0.5">Manage your work and track progress</p>
			</div>
			<button onclick={() => showNewProject = true} class="btn-cta">
				<Plus size={16} />
				New Project
			</button>
		</div>

		<!-- Stats Strip -->
		<div class="prm-stat-strip">
			<button class="prm-stat-item {statusFilter === '' ? 'prm-stat-item--active' : ''}" onclick={() => { statusFilter = ''; projects.loadProjects(); }}>
				<span class="prm-stat-val">{stats.total}</span> Total
			</button>
			<span class="prm-stat-sep"></span>
			<button class="prm-stat-item {statusFilter === 'active' ? 'prm-stat-item--active' : ''}" onclick={() => { statusFilter = 'active'; projects.loadProjects('active'); }}>
				<span class="prm-stat-val">{stats.active}</span> Active
			</button>
			<span class="prm-stat-sep"></span>
			<button class="prm-stat-item {statusFilter === 'paused' ? 'prm-stat-item--active' : ''}" onclick={() => { statusFilter = 'paused'; projects.loadProjects('paused'); }}>
				<span class="prm-stat-val">{stats.paused}</span> Paused
			</button>
			<span class="prm-stat-sep"></span>
			<button class="prm-stat-item {statusFilter === 'completed' ? 'prm-stat-item--active' : ''}" onclick={() => { statusFilter = 'completed'; projects.loadProjects('completed'); }}>
				<span class="prm-stat-val">{stats.completed}</span> Done
			</button>
		</div>
	</div>

	<!-- Filters & Controls Bar -->
	<div class="px-6 py-2 prm-ls-bar flex items-center gap-3 flex-wrap">
		<!-- Search -->
		<div class="relative flex-1 min-w-[200px] max-w-xs">
			<Search size={16} class="absolute left-3 top-1/2 -translate-y-1/2 prm-ls-icon" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search projects..."
				class="w-full pl-9 pr-3 py-1 text-sm prm-ls-search rounded-lg focus:outline-none focus:ring-2 focus:border-transparent"
			/>
		</div>

		<!-- Status Filter Pills -->
		<div class="flex items-center gap-1 prm-ls-divider-l pl-3">
			<button
				onclick={() => { statusFilter = ''; projects.loadProjects(); }}
				class="prm-filter-pill {statusFilter === '' ? 'prm-filter-pill--active' : ''}"
			>
				All
			</button>
			<button
				onclick={() => { statusFilter = 'active'; projects.loadProjects('active'); }}
				class="prm-filter-pill {statusFilter === 'active' ? 'prm-filter-pill--active' : ''}"
			>
				Active
			</button>
			<button
				onclick={() => { statusFilter = 'paused'; projects.loadProjects('paused'); }}
				class="prm-filter-pill {statusFilter === 'paused' ? 'prm-filter-pill--active' : ''}"
			>
				Paused
			</button>
			<button
				onclick={() => { statusFilter = 'completed'; projects.loadProjects('completed'); }}
				class="prm-filter-pill {statusFilter === 'completed' ? 'prm-filter-pill--active' : ''}"
			>
				Completed
			</button>
		</div>

		<!-- Type Filter -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="prm-dropdown-trigger {typeFilter ? 'prm-dropdown-trigger--active' : ''}" aria-label="Filter by type">
				<span class="prm-dropdown-trigger__label">{typeOptions.find(opt => opt.value === typeFilter)?.label || 'All Types'}</span>
				<ChevronDown size={12} class="prm-dropdown-trigger__chevron" />
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content class="prm-dropdown-menu" sideOffset={6} align="start">
					<div class="prm-dropdown-menu__header">Type</div>
					{#each typeOptions as option}
						<DropdownMenu.Item
							class="prm-dropdown-menu__item {typeFilter === option.value ? 'prm-dropdown-menu__item--active' : ''}"
							onclick={() => typeFilter = option.value}
						>
							{#if typeFilter === option.value}
								<span class="prm-dropdown-menu__check">
									<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M5 13l4 4L19 7"/></svg>
								</span>
							{:else}
								<span class="prm-dropdown-menu__check"></span>
							{/if}
							{option.label}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Priority Filter -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="prm-dropdown-trigger {priorityFilter ? 'prm-dropdown-trigger--active' : ''}" aria-label="Filter by priority">
				<span class="prm-dropdown-trigger__label">{priorityOptions.find(opt => opt.value === priorityFilter)?.label || 'All Priorities'}</span>
				<ChevronDown size={12} class="prm-dropdown-trigger__chevron" />
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content class="prm-dropdown-menu" sideOffset={6} align="start">
					<div class="prm-dropdown-menu__header">Priority</div>
					{#each priorityOptions as option}
						<DropdownMenu.Item
							class="prm-dropdown-menu__item {priorityFilter === option.value ? 'prm-dropdown-menu__item--active' : ''}"
							onclick={() => priorityFilter = option.value}
						>
							{#if option.value}
								<span class="prm-priority-dot" style="background: {getPriorityDot(option.value)};"></span>
							{:else}
								<span class="prm-dropdown-menu__check">
									{#if !priorityFilter}
										<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M5 13l4 4L19 7"/></svg>
									{/if}
								</span>
							{/if}
							{option.label}
							{#if priorityFilter === option.value && option.value}
								<svg class="ml-auto" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M5 13l4 4L19 7"/></svg>
							{/if}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Clear Filters -->
		{#if hasActiveFilters}
			<button onclick={clearFilters} class="prm-filter-pill prm-filter-clear">
				Clear filters
			</button>
		{/if}

		<!-- Spacer -->
		<div class="flex-1"></div>

		<!-- Group by Type Toggle -->
		<label class="prm-group-toggle">
			<span class="prm-toggle-box" class:prm-toggle-box--checked={groupByType}>
				{#if groupByType}
					<svg class="w-2.5 h-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
					</svg>
				{/if}
			</span>
			<input
				type="checkbox"
				bind:checked={groupByType}
				class="sr-only"
			/>
			Group by type
		</label>

		<!-- View Mode Toggle -->
		<div class="prm-view-toggle">
			<button
				onclick={() => viewMode = 'grid'}
				class="prm-view-btn {viewMode === 'grid' ? 'prm-view-btn--active' : ''}"
				title="Grid view"
				aria-label="Grid view"
			>
				<LayoutGrid size={16} />
			</button>
			<button
				onclick={() => viewMode = 'list'}
				class="prm-view-btn {viewMode === 'list' ? 'prm-view-btn--active' : ''}"
				title="List view"
				aria-label="List view"
			>
				<List size={16} />
			</button>
			<button
				onclick={() => viewMode = 'kanban'}
				class="prm-view-btn {viewMode === 'kanban' ? 'prm-view-btn--active' : ''}"
				title="Kanban view"
				aria-label="Kanban view"
			>
				<Columns3 size={16} />
			</button>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto px-6 pt-4 pb-6">
		{#if $projects.loading}
			<div class="prm-skeleton-grid">
				{#each Array(6) as _}
					<div class="prm-skeleton-card">
						<div class="prm-skeleton-line prm-skeleton-line--title"></div>
						<div class="prm-skeleton-line prm-skeleton-line--short"></div>
						<div class="prm-skeleton-line prm-skeleton-line--full"></div>
						<div class="prm-skeleton-footer">
							<div class="prm-skeleton-dot"></div>
							<div class="prm-skeleton-line prm-skeleton-line--tiny"></div>
						</div>
					</div>
				{/each}
			</div>
		{:else if $projects.projects.length === 0}
			<div class="flex flex-col items-center justify-center h-64 text-center">
				<div class="prm-empty-card">
					<h3 class="text-base font-semibold prm-ls-title mb-2">Get started</h3>
					<div class="prm-empty-steps">
						<div class="prm-empty-step">
							<span class="prm-empty-step-num">1</span>
							<span class="prm-ls-muted text-sm">Create a project</span>
						</div>
						<div class="prm-empty-step">
							<span class="prm-empty-step-num">2</span>
							<span class="prm-ls-muted text-sm">Add tasks & team</span>
						</div>
						<div class="prm-empty-step">
							<span class="prm-empty-step-num">3</span>
							<span class="prm-ls-muted text-sm">Track progress</span>
						</div>
					</div>
					<button onclick={() => showNewProject = true} class="btn-cta mt-4">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
						Create Project
					</button>
				</div>
			</div>
		{:else if filteredProjects().length === 0}
			<div class="flex flex-col items-center justify-center h-48 text-center">
				<p class="prm-ls-muted mb-2">No projects match your filters</p>
				<button onclick={clearFilters} class="btn-pill btn-pill-ghost btn-pill-sm underline">
					Clear all filters
				</button>
			</div>
		{:else if viewMode === 'kanban'}
			<!-- Kanban View -->
			<div class="flex gap-4 h-full min-h-0">
				{#each [
					{ key: 'active', label: 'Active', items: groupedByStatus.active },
					{ key: 'paused', label: 'Paused', items: groupedByStatus.paused },
					{ key: 'completed', label: 'Completed', items: groupedByStatus.completed }
				] as col}
					<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col prm-kanban-col overflow-hidden">
						<div class="px-4 py-2.5 prm-kanban-col__header flex items-center gap-2">
							<span class="prm-status-dot prm-status-dot--{col.key}"></span>
							<span class="text-xs font-medium prm-ls-title uppercase tracking-wider">{col.label}</span>
							<span class="prm-ls-kanban-count ml-auto">{col.items.length}</span>
						</div>
						<div class="flex-1 overflow-y-auto p-2 space-y-2">
							{#each col.items as project}
								<a href="/projects/{project.id}{embedSuffix}" class="block p-3 prm-kanban-card rounded-lg transition-colors">
									<span class="text-sm font-medium prm-ls-title line-clamp-1 block mb-1">{project.name}</span>
									{#if project.client_name}
										<p class="text-xs prm-ls-muted mb-1">{project.client_name}</p>
									{/if}
									<div class="flex items-center justify-between text-xs prm-ls-icon mt-2">
										<span class="flex items-center gap-1">
											<span class="prm-priority-dot" style="background: {getPriorityDot(project.priority)}"></span>
											<span class="prm-ls-muted capitalize">{project.priority}</span>
										</span>
										<span>{formatDate(project.updated_at)}</span>
									</div>
								</a>
							{/each}
							{#if col.items.length === 0}
								<p class="text-xs prm-ls-icon text-center py-4">No {col.label.toLowerCase()} projects</p>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{:else if viewMode === 'list'}
			<!-- List View -->
			<div class="prm-ls-column overflow-hidden">
				<table class="w-full">
					<thead class="prm-ls-table-head">
						<tr>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">
								<button class="prm-sort-btn" onclick={() => { if (sortField === 'name') sortDir = sortDir === 'asc' ? 'desc' : 'asc'; else { sortField = 'name'; sortDir = 'asc'; } }}>
									Project
									{#if sortField === 'name'}
										<svg class="w-3 h-3 {sortDir === 'desc' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 15l7-7 7 7"/></svg>
									{/if}
								</button>
							</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">
								<button class="prm-sort-btn" onclick={() => { if (sortField === 'type') sortDir = sortDir === 'asc' ? 'desc' : 'asc'; else { sortField = 'type'; sortDir = 'asc'; } }}>
									Type
									{#if sortField === 'type'}
										<svg class="w-3 h-3 {sortDir === 'desc' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 15l7-7 7 7"/></svg>
									{/if}
								</button>
							</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">
								<button class="prm-sort-btn" onclick={() => { if (sortField === 'status') sortDir = sortDir === 'asc' ? 'desc' : 'asc'; else { sortField = 'status'; sortDir = 'asc'; } }}>
									Status
									{#if sortField === 'status'}
										<svg class="w-3 h-3 {sortDir === 'desc' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 15l7-7 7 7"/></svg>
									{/if}
								</button>
							</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">
								<button class="prm-sort-btn" onclick={() => { if (sortField === 'priority') sortDir = sortDir === 'asc' ? 'desc' : 'asc'; else { sortField = 'priority'; sortDir = 'asc'; } }}>
									Priority
									{#if sortField === 'priority'}
										<svg class="w-3 h-3 {sortDir === 'desc' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 15l7-7 7 7"/></svg>
									{/if}
								</button>
							</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">
								<button class="prm-sort-btn" onclick={() => { if (sortField === 'updated') sortDir = sortDir === 'asc' ? 'desc' : 'asc'; else { sortField = 'updated'; sortDir = 'desc'; } }}>
									Updated
									{#if sortField === 'updated'}
										<svg class="w-3 h-3 {sortDir === 'desc' ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 15l7-7 7 7"/></svg>
									{/if}
								</button>
							</th>
						</tr>
					</thead>
					<tbody class="prm-ls-table-body">
						{#each filteredProjects() as project}
							<tr class="prm-ls-table-row transition-colors cursor-pointer" onclick={() => goto(`/projects/${project.id}${embedSuffix}`)}>
								<td class="px-4 py-3">
									<div>
										<span class="font-medium prm-ls-title">{project.name}</span>
										{#if project.client_name}
											<span class="prm-ls-icon ml-2">· {project.client_name}</span>
										{/if}
									</div>
									{#if project.description}
										<p class="text-sm prm-ls-muted line-clamp-1 mt-0.5">{project.description}</p>
									{/if}
								</td>
								<td class="px-4 py-3">
									<span class="text-xs prm-ls-muted">{getTypeLabel(project.project_type)}</span>
								</td>
								<td class="px-4 py-3">
									<span class="flex items-center gap-1.5 text-xs prm-ls-muted capitalize">
										<span class="prm-status-dot prm-status-dot--{project.status}"></span>
										{project.status}
									</span>
								</td>
								<td class="px-4 py-3">
									<span class="flex items-center gap-1.5 text-xs prm-ls-muted capitalize">
										<span class="prm-priority-dot" style="background: {getPriorityDot(project.priority)}"></span>
										{project.priority}
									</span>
								</td>
								<td class="px-4 py-3 text-sm prm-ls-muted">
									{formatDate(project.updated_at)}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else if groupByType}
			<!-- Grid View Grouped by Type -->
			<div class="space-y-8">
				{#if groupedByType.internal.length > 0}
					<div>
						<div class="flex items-center gap-2 mb-3">
						<h2 class="text-xs font-semibold prm-ls-muted uppercase tracking-wider">Internal Projects</h2>
						<span class="text-xs prm-ls-icon">({groupedByType.internal.length})</span>
						</div>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each groupedByType.internal as project}
								{@render projectCard(project)}
							{/each}
						</div>
					</div>
				{/if}

				{#if groupedByType.client_work.length > 0}
					<div>
						<div class="flex items-center gap-2 mb-3">
						<h2 class="text-xs font-semibold prm-ls-muted uppercase tracking-wider">Client Work</h2>
						<span class="text-xs prm-ls-icon">({groupedByType.client_work.length})</span>
						</div>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each groupedByType.client_work as project}
								{@render projectCard(project)}
							{/each}
						</div>
					</div>
				{/if}

				{#if groupedByType.learning.length > 0}
					<div>
						<div class="flex items-center gap-2 mb-3">
						<h2 class="text-xs font-semibold prm-ls-muted uppercase tracking-wider">Learning</h2>
						<span class="text-xs prm-ls-icon">({groupedByType.learning.length})</span>
						</div>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each groupedByType.learning as project}
								{@render projectCard(project)}
							{/each}
						</div>
					</div>
				{/if}

				{#if groupedByType.other.length > 0}
					<div>
						<div class="flex items-center gap-2 mb-3">
						<h2 class="text-xs font-semibold prm-ls-muted uppercase tracking-wider">Other</h2>
						<span class="text-xs prm-ls-icon">({groupedByType.other.length})</span>
						</div>
						<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
							{#each groupedByType.other as project}
								{@render projectCard(project)}
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<!-- Standard Grid View -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each filteredProjects() as project}
					{@render projectCard(project)}
				{/each}
			</div>
		{/if}
	</div>
</div>

{#snippet projectCard(project: Project)}
	<a href="/projects/{project.id}{embedSuffix}" class="block prm-card transition-all duration-200 cursor-pointer group">
		<div class="p-3">
			<div class="flex items-start justify-between mb-1.5">
				<h3 class="font-medium prm-ls-title line-clamp-1 text-sm">{project.name}</h3>
				<div class="flex items-center gap-1.5 shrink-0 ml-2">
					<span class="prm-status-dot prm-status-dot--{project.status}"></span>
					<span class="text-xs prm-ls-muted capitalize">{project.status}</span>
				</div>
			</div>
			{#if project.client_name}
				<p class="text-xs prm-ls-muted mb-1">{project.client_name}</p>
			{/if}
			{#if project.description}
				<p class="text-xs prm-ls-icon line-clamp-1 mb-2">{project.description}</p>
			{:else}
				<div class="mb-2"></div>
			{/if}
			<div class="flex items-center justify-between text-xs pt-2 prm-card__footer">
				<div class="flex items-center gap-3">
					<span class="flex items-center gap-1">
						<span class="prm-priority-dot" style="background: {getPriorityDot(project.priority)}"></span>
						<span class="prm-ls-muted capitalize">{project.priority}</span>
					</span>
					<span class="prm-ls-icon">{getTypeLabel(project.project_type)}</span>
				</div>
				<span class="prm-ls-icon">{formatDate(project.updated_at)}</span>
			</div>
		</div>
	</a>
{/snippet}

<!-- New Project Dialog -->
<Dialog.Root bind:open={showNewProject}>
	<Dialog.Portal>
		<Dialog.Overlay class="bos-modal-overlay" style="position:fixed;padding:0;" />
		<Dialog.Content class="prm-modal" aria-describedby={undefined}>
			<form onsubmit={handleCreateProject} class="prm-modal__form">
				<!-- Header -->
				<div class="prm-modal__header">
					<Dialog.Title class="prm-modal__title">New Project</Dialog.Title>
					<Dialog.Close class="prm-modal__close" onclick={() => { showNewProject = false; showAdvancedOptions = false; }} aria-label="Close">
						<X size={16} />
					</Dialog.Close>
				</div>

				<!-- Body -->
				<div class="prm-modal__body">
					<!-- Name -->
					<div class="prm-modal__field">
						<label for="prm-name" class="prm-modal__label">Name</label>
						<input
							id="prm-name"
							type="text"
							bind:value={newProject.name}
							class="prm-modal__input"
							placeholder="e.g. Website Redesign"
							required
						/>
					</div>

					<!-- Type + Priority side by side -->
					<div class="prm-modal__row">
						<!-- Type -->
						<div class="prm-modal__field prm-modal__field--flex1">
							<label class="prm-modal__label">Type</label>
							<div class="prm-modal__type-group">
								{#each [
									{ value: 'internal', label: 'Internal', Icon: Briefcase },
									{ value: 'client_work', label: 'Client', Icon: Users },
									{ value: 'learning', label: 'Learning', Icon: GraduationCap }
								] as opt}
									<button
										type="button"
										onclick={() => newProject.project_type = opt.value}
										class="prm-modal__type-btn {newProject.project_type === opt.value ? 'prm-modal__type-btn--active' : ''}"
										aria-label="Set type to {opt.label}"
									>
										<opt.Icon size={14} />
										<span>{opt.label}</span>
									</button>
								{/each}
							</div>
						</div>

						<!-- Priority -->
						<div class="prm-modal__field prm-modal__field--flex1">
							<label class="prm-modal__label">Priority</label>
							<div class="prm-modal__priority-group">
								{#each [
									{ value: 'low', label: 'Low' },
									{ value: 'medium', label: 'Med' },
									{ value: 'high', label: 'High' },
									{ value: 'critical', label: 'Crit' }
								] as opt}
									<button
										type="button"
										onclick={() => newProject.priority = opt.value as 'low' | 'medium' | 'high' | 'critical'}
										class="prm-modal__priority-btn {newProject.priority === opt.value ? 'prm-modal__priority-btn--active' : ''}"
										aria-label="Set priority to {opt.label}"
									>
										<span class="prm-modal__priority-dot" style="background: {getPriorityDot(opt.value)};"></span>
										{opt.label}
									</button>
								{/each}
							</div>
						</div>
					</div>

					<!-- Client -->
					<div class="prm-modal__field">
						<label class="prm-modal__label">Client <span class="prm-modal__optional">(optional)</span></label>
						<DropdownMenu.Root>
							<DropdownMenu.Trigger class="prm-modal__select" aria-label="Select client">
								<span class={newProject.client_name ? 'prm-modal__select-value' : 'prm-modal__select-placeholder'}>
									{newProject.client_name || 'None'}
								</span>
								<ChevronDown size={14} class="prm-modal__select-chevron" />
							</DropdownMenu.Trigger>
							<DropdownMenu.Portal>
								<DropdownMenu.Content class="prm-dropdown-menu" style="z-index:1100; width: var(--bits-dropdown-trigger-width);" sideOffset={4}>
									<DropdownMenu.Item
										class="prm-dropdown-menu__item {!newProject.client_name ? 'prm-dropdown-menu__item--active' : ''}"
										onclick={() => newProject.client_name = ''}
									>
										None
									</DropdownMenu.Item>
									{#each clients as client}
										<DropdownMenu.Item
											class="prm-dropdown-menu__item {newProject.client_name === client.name ? 'prm-dropdown-menu__item--active' : ''}"
											onclick={() => newProject.client_name = client.name}
										>
											{client.name}
										</DropdownMenu.Item>
									{/each}
								</DropdownMenu.Content>
							</DropdownMenu.Portal>
						</DropdownMenu.Root>
					</div>

					<!-- Description -->
					<div class="prm-modal__field">
						<label for="prm-desc" class="prm-modal__label">Description <span class="prm-modal__optional">(optional)</span></label>
						<textarea
							id="prm-desc"
							bind:value={newProject.description}
							class="prm-modal__textarea"
							rows="3"
							placeholder="What is this project about?"
						></textarea>
						{#if newProject.description}
							<span class="prm-modal__charcount">{newProject.description.length}/500</span>
						{/if}
					</div>

					{#if createError}
						<div class="prm-modal__error">
							<AlertCircle size={14} />
							{createError}
						</div>
					{/if}
				</div>

				<!-- Footer -->
				<div class="prm-modal__footer">
					<button type="button" onclick={() => { showNewProject = false; showAdvancedOptions = false; }} class="prm-modal__btn prm-modal__btn--ghost" aria-label="Cancel">
						Cancel
					</button>
					<button type="submit" disabled={!newProject.name.trim()} class="prm-modal__btn prm-modal__btn--primary" aria-label="Create project">
						Create Project
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	/* ── Page & Layout ── */
	.prm-ls-page { background: var(--dbg2, rgba(249,250,251,.5)); }
	.prm-ls-bar { background: var(--dbg, #fff); border-bottom: 1px solid var(--dbd, #e5e7eb); }
	:global(.prm-ls-title) { color: var(--bos-text-primary-color); }
	:global(.prm-ls-muted) { color: var(--bos-text-secondary-color); }
	:global(.prm-ls-icon) { color: var(--bos-text-tertiary-color); }
	:global(.prm-ls-label) { color: var(--bos-text-secondary-color); }

	/* ── Search ── */
	.prm-ls-search { border: 1px solid var(--dbd, #e5e7eb); background: var(--dbg, #fff); color: var(--dt, #111); }
	.prm-ls-search:focus { box-shadow: 0 0 0 2px var(--dt, #111); }
	.prm-ls-divider-l { border-left: 1px solid var(--dbd, #e5e7eb); }

	/* ── Dropdown Trigger ── */
	:global(.prm-dropdown-trigger) {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.3125rem 0.625rem;
		font-size: 0.75rem;
		font-weight: 500;
		border-radius: 0.375rem;
		border: 1px solid var(--dbd, #e0e0e0);
		background: var(--dbg, #fff);
		color: var(--dt2, #555);
		cursor: pointer;
		transition: all 0.15s;
		white-space: nowrap;
	}
	:global(.prm-dropdown-trigger:hover) {
		border-color: var(--dt3, #888);
		color: var(--dt, #111);
	}
	:global(.prm-dropdown-trigger--active) {
		border-color: var(--dt, #111);
		color: var(--dt, #111);
		font-weight: 600;
	}
	:global(.prm-dropdown-trigger__label) { line-height: 1; }
	:global(.prm-dropdown-trigger__chevron) { opacity: 0.5; transition: transform 0.15s; }
	:global([data-state="open"] .prm-dropdown-trigger__chevron) { transform: rotate(180deg); }

	/* ── Dropdown Menu ── */
	:global(.prm-dropdown-menu) {
		z-index: 50;
		min-width: 180px;
		background: var(--bos-modal-bg);
		border: 1px solid var(--bos-border-color);
		border-radius: 0.5rem;
		box-shadow: var(--bos-popover-shadow);
		padding: 0.25rem;
		overflow: hidden;
	}
	:global(.prm-dropdown-menu__header) {
		padding: 0.375rem 0.625rem 0.25rem;
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: var(--dt3, #888);
	}
	:global(.prm-dropdown-menu__item) {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.4375rem 0.625rem;
		font-size: 0.8125rem;
		border-radius: 0.3125rem;
		color: var(--bos-text-primary-color);
		cursor: pointer;
		transition: background 0.1s;
	}
	:global(.prm-dropdown-menu__item:hover) {
		background: var(--bos-hover-color);
	}
	:global(.prm-dropdown-menu__item--active) {
		font-weight: 600;
	}
	:global(.prm-dropdown-menu__check) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 14px;
		height: 14px;
		flex-shrink: 0;
		color: var(--dt, #111);
	}

	/* ── Controls ── */
	:global(.prm-ls-empty-bg) { background: var(--dbg3, #f3f4f6); }

	/* ── Filter Pills (monochrome) ── */
	.prm-filter-pill {
		display: inline-flex;
		align-items: center;
		padding: 0.25rem 0.625rem;
		font-size: 0.75rem;
		font-weight: 500;
		border-radius: 0.375rem;
		border: 1px solid transparent;
		background: transparent;
		color: var(--dt3, #888);
		cursor: pointer;
		transition: all 0.15s;
		white-space: nowrap;
	}
	.prm-filter-pill:hover {
		color: var(--dt, #111);
		background: var(--dbg2, #f5f5f5);
	}
	.prm-filter-pill--active {
		color: var(--dt, #111);
		background: var(--dt, #111);
		color: var(--dbg, #fff);
		font-weight: 600;
	}
	.prm-filter-pill--active:hover {
		background: var(--dt2, #333);
		color: var(--dbg, #fff);
	}
	.prm-filter-clear {
		text-decoration: underline;
		text-underline-offset: 2px;
	}

	/* ── Group Toggle (custom checkbox) ── */
	.prm-group-toggle {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.75rem;
		color: var(--dt2, #555);
		cursor: pointer;
		user-select: none;
	}
	.prm-toggle-box {
		width: 0.875rem;
		height: 0.875rem;
		border-radius: 0.1875rem;
		border: 1.5px solid var(--dbd, #d1d5db);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		transition: all 0.15s;
		background: transparent;
	}
	.prm-toggle-box--checked {
		background: var(--dt, #111);
		border-color: var(--dt, #111);
		color: #fff;
	}
	.prm-group-toggle:hover .prm-toggle-box:not(.prm-toggle-box--checked) {
		border-color: var(--dt3, #888);
	}

	/* ── View Toggle (compact segmented control) ── */
	.prm-view-toggle {
		display: inline-flex;
		align-items: center;
		gap: 1px;
		border: 1px solid var(--dbd, #e5e7eb);
		border-radius: 0.5rem;
		padding: 2px;
		background: var(--dbg2, #f9fafb);
	}
	.prm-view-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.25rem;
		border-radius: 0.375rem;
		color: var(--dt3, #888);
		background: transparent;
		border: none;
		cursor: pointer;
		transition: all 0.15s;
	}
	.prm-view-btn:hover { color: var(--dt, #111); background: var(--dbg, #fff); }
	.prm-view-btn--active {
		color: var(--dt, #111);
		background: var(--dbg, #fff);
		box-shadow: 0 1px 2px rgba(0,0,0,0.06);
	}

	/* ── Stat Strip (inline) ── */
	.prm-stat-strip {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-top: 0.75rem;
		font-size: 0.8125rem;
		color: var(--dt3, #6b7280);
	}
	.prm-stat-item { display: inline-flex; align-items: center; gap: 0.25rem; font-size: inherit; color: inherit; }
	.prm-stat-val { font-weight: 600; color: var(--dt, #111); }
	.prm-stat-sep { width: 3px; height: 3px; border-radius: 50%; background: var(--dbd, #d1d5db); flex-shrink: 0; }

	/* ── Project Cards (monochrome + blue accent) ── */
	.prm-card {
		background: var(--dbg, #fff);
		border-radius: 0.75rem;
		border: 1px solid var(--dbd, #e5e7eb);
		outline: none;
		box-shadow: none;
		transition: all 0.15s ease;
	}
	.prm-card:hover {
		box-shadow: 0 2px 8px rgba(0,0,0,0.12);
		transform: translateY(-1px);
	}
	.prm-card__footer { border-top: 1px solid var(--dbd2, #f3f4f6); }

	/* ── Priority Dot (tiny colored indicator) ── */
	.prm-priority-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
		display: inline-block;
	}

	/* ── Status Dot (tiny colored indicator) ── */
	.prm-status-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
		display: inline-block;
	}
	.prm-status-dot--active { background: #22c55e; }
	.prm-status-dot--paused { background: #f59e0b; }
	.prm-status-dot--completed { background: #9ca3af; }
	.prm-status-dot--archived { background: #d1d5db; }

	/* ── Kanban ── */
	.prm-kanban-col { background: var(--dbg, #fff); border-radius: 0.75rem; border: 1px solid var(--dbd, #e5e7eb); }
	.prm-kanban-col__header { border-bottom: 1px solid var(--dbd2, #f3f4f6); background: var(--dbg2, #f9fafb); border-radius: 0.75rem 0.75rem 0 0; }
	.prm-kanban-card { background: var(--dbg, #fff); border: 1px solid var(--dbd2, #f3f4f6); }
	.prm-kanban-card:hover { background: var(--dbg2, #f9fafb); }
	.prm-ls-kanban-count { display: inline-flex; align-items: center; justify-content: center; min-width: 1.25rem; height: 1.25rem; padding: 0 0.375rem; font-size: 0.6875rem; font-weight: 600; border-radius: 9999px; background: var(--dbg3, #f3f4f6); color: var(--dt3, #6b7280); }

	/* ── List view ── */
	.prm-ls-column { background: var(--dbg, #fff); border-radius: 0.75rem; border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-table-head { background: var(--dbg2, #f9fafb); border-bottom: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-table-body > :global(tr + tr) { border-top: 1px solid var(--dbd2, #f3f4f6); }
	.prm-ls-table-row:hover { background: var(--dbg2, #f9fafb); }

	/* ── Filter active state ── */
	.prm-filter--active { box-shadow: 0 0 0 2px var(--dt3); }
	:global(.prm-dropdown-item--active) { background: var(--dbg3); font-weight: 600; }

	/* ═══════════════════════════════════════════════════
	   NEW PROJECT MODAL — Premium Design
	   ═══════════════════════════════════════════════════ */
	:global(.prm-modal) {
		position: fixed;
		left: 50%;
		top: 50%;
		transform: translate(-50%, -52%);
		width: calc(100% - 2rem);
		max-width: 440px;
		max-height: calc(85vh - 60px);
		z-index: 1000;
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 14px;
		box-shadow: 0 20px 40px -8px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(0, 0, 0, 0.03);
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}
	:global(.dark .prm-modal),
	:global(.dark) :global(.prm-modal) {
		background: var(--dbg, #141414);
		border-color: var(--dbd, #1e1e1e);
		box-shadow: 0 20px 40px -8px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.04);
	}

	:global(.prm-modal__form) {
		display: flex;
		flex-direction: column;
		height: 100%;
		overflow: hidden;
	}

	:global(.prm-modal__header) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 18px 20px 0;
	}

	:global(.prm-modal__title) {
		font-size: 16px;
		font-weight: 700;
		color: var(--dt, #111);
		letter-spacing: -0.02em;
		margin: 0;
	}

	:global(.prm-modal__close) {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		border-radius: 6px;
		border: none;
		background: transparent;
		color: var(--dt3, #888);
		cursor: pointer;
		transition: all 0.15s;
	}
	:global(.prm-modal__close:hover) {
		background: var(--dbg3, #eee);
		color: var(--dt, #111);
	}

	:global(.prm-modal__body) {
		flex: 1;
		overflow-y: auto;
		padding: 16px 20px 20px;
		display: flex;
		flex-direction: column;
		gap: 16px;
		scrollbar-width: thin;
		scrollbar-color: var(--dbd, #e0e0e0) transparent;
	}

	:global(.prm-modal__field) {
		display: flex;
		flex-direction: column;
		gap: 5px;
	}
	:global(.prm-modal__field--flex1) { flex: 1; min-width: 0; }

	:global(.prm-modal__row) {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}

	:global(.prm-modal__label) {
		font-size: 11px;
		font-weight: 600;
		color: var(--dt2, #555);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	:global(.prm-modal__optional) {
		font-weight: 400;
		color: var(--dt4, #bbb);
		text-transform: none;
		letter-spacing: 0;
	}

	:global(.prm-modal__input) {
		width: 100%;
		padding: 8px 10px;
		font-size: 13px;
		color: var(--dt, #111);
		background: var(--dbg2, #f5f5f5);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 8px;
		outline: none;
		transition: border-color 0.15s, box-shadow 0.15s;
	}
	:global(.prm-modal__input::placeholder) { color: var(--dt4, #bbb); }
	:global(.prm-modal__input:focus) {
		border-color: var(--dt, #111);
		box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.04);
		background: var(--dbg, #fff);
	}
	:global(.dark .prm-modal__input:focus) {
		box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.06);
		background: var(--dbg, #141414);
	}

	:global(.prm-modal__textarea) {
		width: 100%;
		padding: 8px 10px;
		font-size: 13px;
		color: var(--dt, #111);
		background: var(--dbg2, #f5f5f5);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 8px;
		outline: none;
		resize: vertical;
		min-height: 64px;
		font-family: inherit;
		line-height: 1.5;
		transition: border-color 0.15s, box-shadow 0.15s;
	}
	:global(.prm-modal__textarea::placeholder) { color: var(--dt4, #bbb); }
	:global(.prm-modal__textarea:focus) {
		border-color: var(--dt, #111);
		box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.04);
		background: var(--dbg, #fff);
	}
	:global(.dark .prm-modal__textarea:focus) {
		box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.06);
		background: var(--dbg, #141414);
	}
	:global(.prm-modal__charcount) {
		font-size: 10px;
		color: var(--dt4, #bbb);
		text-align: right;
		margin-top: -2px;
	}

	/* Type selector — segmented control */
	:global(.prm-modal__type-group) {
		display: flex;
		gap: 0;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 8px;
		overflow: hidden;
		background: var(--dbg2, #f5f5f5);
		padding: 2px;
	}
	:global(.prm-modal__type-btn) {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 4px;
		padding: 6px 4px;
		font-size: 11px;
		font-weight: 500;
		color: var(--dt3, #888);
		background: transparent;
		border: none;
		cursor: pointer;
		transition: all 0.15s;
		white-space: nowrap;
		border-radius: 6px;
	}
	:global(.prm-modal__type-btn:hover) { color: var(--dt, #111); }
	:global(.prm-modal__type-btn--active) {
		background: var(--dbg, #fff);
		color: var(--dt, #111);
		font-weight: 600;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
	}
	:global(.dark .prm-modal__type-btn--active) {
		background: var(--dbg3, #1e1e1e);
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
	}

	/* Priority selector — pill toggle */
	:global(.prm-modal__priority-group) {
		display: flex;
		gap: 4px;
	}
	:global(.prm-modal__priority-btn) {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 4px;
		padding: 6px 4px;
		font-size: 11px;
		font-weight: 500;
		color: var(--dt3, #888);
		background: transparent;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.15s;
	}
	:global(.prm-modal__priority-btn:hover) {
		border-color: var(--dt3, #888);
		color: var(--dt, #111);
	}
	:global(.prm-modal__priority-btn--active) {
		border-color: var(--dt, #111);
		background: var(--dbg, #fff);
		color: var(--dt, #111);
		font-weight: 600;
	}
	:global(.dark .prm-modal__priority-btn--active) {
		background: var(--dbg3, #1e1e1e);
	}
	:global(.prm-modal__priority-dot) {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	/* Client select */
	:global(.prm-modal__select) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		padding: 8px 10px;
		font-size: 13px;
		background: var(--dbg2, #f5f5f5);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 8px;
		cursor: pointer;
		transition: border-color 0.15s;
	}
	:global(.prm-modal__select:hover) { border-color: var(--dt3, #888); }
	:global(.prm-modal__select-value) { color: var(--dt, #111); }
	:global(.prm-modal__select-placeholder) { color: var(--dt4, #bbb); }
	:global(.prm-modal__select-chevron) { color: var(--dt3, #888); flex-shrink: 0; }

	/* Error */
	:global(.prm-modal__error) {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 10px;
		font-size: 12px;
		color: var(--bos-status-error, #ef4444);
		background: var(--bos-status-error-bg);
		border-radius: 6px;
	}

	/* Footer */
	:global(.prm-modal__footer) {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 8px;
		padding: 12px 20px;
		border-top: 1px solid var(--dbd, #e0e0e0);
	}
	:global(.prm-modal__btn) {
		padding: 7px 16px;
		font-size: 12px;
		font-weight: 600;
		border-radius: 6px;
		border: none;
		cursor: pointer;
		transition: all 0.15s;
	}
	:global(.prm-modal__btn--ghost) {
		background: transparent;
		color: var(--dt2, #555);
	}
	:global(.prm-modal__btn--ghost:hover) {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
	}
	:global(.prm-modal__btn--primary) {
		background: var(--bos-btn-cta-bg, #111);
		color: var(--bos-btn-cta-text, #fff);
		box-shadow: var(--bos-btn-cta-glow);
		border: 1px solid var(--bos-btn-cta-border);
	}
	:global(.prm-modal__btn--primary:hover:not(:disabled)) {
		box-shadow: var(--bos-btn-cta-glow-hover);
		transform: translateY(-0.5px);
	}
	:global(.prm-modal__btn--primary:disabled) {
		opacity: 0.3;
		cursor: not-allowed;
		box-shadow: none;
		transform: none;
	}
	/* ── Skeleton loading ── */
	.prm-skeleton-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1rem;
	}
	.prm-skeleton-card {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e5e7eb);
		border-radius: 0.75rem;
		padding: 1rem;
		animation: prm-pulse 1.5s ease-in-out infinite;
	}
	.prm-skeleton-line {
		height: 0.75rem;
		background: var(--dbg3, #eee);
		border-radius: 0.25rem;
		margin-bottom: 0.5rem;
	}
	.prm-skeleton-line--title { width: 70%; height: 0.875rem; }
	.prm-skeleton-line--short { width: 40%; }
	.prm-skeleton-line--full { width: 100%; }
	.prm-skeleton-line--tiny { width: 3rem; }
	.prm-skeleton-footer {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.75rem;
		padding-top: 0.75rem;
		border-top: 1px solid var(--dbd2, #f3f4f6);
	}
	.prm-skeleton-dot {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		background: var(--dbg3, #eee);
	}
	@keyframes prm-pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.5; }
	}

	/* ── Modal sticky footer ── */
	:global(.prm-modal-sticky-footer) {
		position: sticky;
		bottom: 0;
		background: var(--bos-modal-bg);
		z-index: 2;
	}

	/* ── List view alternating rows ── */
	.prm-ls-table-body > :global(tr:nth-child(even)) { background: var(--dbg2, #f9fafb); }

	/* ── Empty state redesign ── */
	.prm-empty-card {
		border: 1px dashed var(--dbd, #e0e0e0);
		border-radius: 0.75rem;
		padding: 2rem 2.5rem;
		max-width: 20rem;
	}
	.prm-empty-steps {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-top: 1rem;
	}
	.prm-empty-step {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		text-align: left;
	}
	.prm-empty-step-num {
		width: 1.25rem;
		height: 1.25rem;
		border-radius: 50%;
		background: var(--dt, #111);
		color: #fff;
		font-size: 0.6875rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	/* ── Sort buttons (list view) ── */
	.prm-sort-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		background: none;
		border: none;
		cursor: pointer;
		font: inherit;
		color: inherit;
		text-transform: inherit;
		letter-spacing: inherit;
	}
	.prm-sort-btn:hover { color: var(--dt, #111); }

	/* ── Stat strip interactivity ── */
	.prm-stat-item {
		cursor: pointer;
		border: none;
		background: none;
		transition: color 0.15s;
	}
	.prm-stat-item:hover { color: var(--dt, #111); }
	.prm-stat-item--active {
		font-weight: 700;
		color: var(--dt, #111);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

</style>
