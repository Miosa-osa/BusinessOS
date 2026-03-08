<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { projects } from '$lib/stores/projects';
	import { api, type ClientListResponse } from '$lib/api';
	import { Dialog, Popover, DropdownMenu } from 'bits-ui';
	import type { Project } from '$lib/api';

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
		icon: '📁'
	});
	let statusFilter = $state('');
	let typeFilter = $state('');
	let priorityFilter = $state('');
	let searchQuery = $state('');
	let viewMode = $state<'grid' | 'list' | 'kanban'>('grid');
	let groupByType = $state(false);
	let createError = $state('');

	// Clients for dropdown — seed from preloaded data
	let clients = $state<ClientListResponse[]>(data?.clients ?? []);
	let showIconPicker = $state(false);
	let showAdvancedOptions = $state(false);

	// Project icons
	const projectIcons = [
		'📁', '📂', '🗂️', '📊', '📈', '📉', '💼', '🏢', '🏠', '🏭',
		'💡', '🎯', '⭐', '🌟', '✨', '🔥', '💎', '🎨', '🎬', '📸',
		'🛠️', '⚙️', '🔧', '🔨', '🧰', '💻', '🖥️', '📱', '🌐', '🔌',
		'📦', '🚀', '🛸', '✈️', '🚗', '🏆', '🎓', '📚', '📖', '✏️'
	];

	onMount(async () => {
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

	// Filtered projects
	let filteredProjects = $derived((() => {
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

		return result;
	})());

	// Stats
	let stats = $derived({
		total: $projects.projects.length,
		active: $projects.projects.filter(p => p.status === 'active').length,
		paused: $projects.projects.filter(p => p.status === 'paused').length,
		completed: $projects.projects.filter(p => p.status === 'completed').length
	});

	// Grouped by type
	let groupedByType = $derived({
		internal: filteredProjects.filter(p => p.project_type === 'internal'),
		client_work: filteredProjects.filter(p => p.project_type === 'client_work'),
		learning: filteredProjects.filter(p => p.project_type === 'learning'),
		other: filteredProjects.filter(p => !['internal', 'client_work', 'learning'].includes(p.project_type))
	});

	// Grouped by status for kanban
	let groupedByStatus = $derived({
		active: filteredProjects.filter(p => p.status === 'active'),
		paused: filteredProjects.filter(p => p.status === 'paused'),
		completed: filteredProjects.filter(p => p.status === 'completed')
	});

	async function handleCreateProject(e: Event) {
		e.preventDefault();
		createError = '';
		try {
			await projects.createProject(newProject);
			showNewProject = false;
			newProject = { name: '', description: '', client_name: '', project_type: 'internal', priority: 'medium', icon: '📁' };
			showAdvancedOptions = false;
		} catch (error) {
			createError = (error as Error).message || 'Failed to create project';
		}
	}

	function getTypeEmoji(type: string) {
		switch (type) {
			case 'internal': return '🏢';
			case 'client_work': return '👥';
			case 'learning': return '📚';
			default: return '📁';
		}
	}

	function getPriorityEmoji(priority: string) {
		switch (priority) {
			case 'critical': return '🔴';
			case 'high': return '🟠';
			case 'medium': return '🟡';
			case 'low': return '🟢';
			default: return '⚪';
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'active': return 'prm-status prm-status--active';
			case 'paused': return 'prm-status prm-status--paused';
			case 'completed': return 'prm-status prm-status--completed';
			case 'archived': return 'prm-ls-status-default';
			default: return 'prm-ls-status-default';
		}
	}

	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'critical': return 'prm-priority prm-priority--critical';
			case 'high': return 'prm-priority prm-priority--high';
			case 'medium': return 'prm-priority prm-priority--medium';
			case 'low': return 'prm-priority prm-priority--low';
			default: return 'prm-ls-priority-default';
		}
	}

	function getPriorityIcon(priority: string) {
		const count = priority === 'critical' ? 3 : priority === 'high' ? 2 : priority === 'medium' ? 1 : 0;
		return count;
	}

	function getTypeColor(type: string) {
		switch (type) {
			case 'internal': return 'prm-type prm-type--internal';
			case 'client_work': return 'prm-type prm-type--client';
			case 'learning': return 'prm-type prm-type--learning';
			default: return 'prm-ls-type-default';
		}
	}

	function getTypeLabel(type: string) {
		switch (type) {
			case 'internal': return 'Internal';
			case 'client_work': return 'Client Work';
			case 'learning': return 'Learning';
			default: return type;
		}
	}

	function getTypeIcon(type: string) {
		switch (type) {
			case 'internal': return '🏢';
			case 'client_work': return '👥';
			case 'learning': return '📚';
			default: return '📁';
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
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
			<button onclick={() => showNewProject = true} class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				New Project
			</button>
		</div>

		<!-- Stats Row -->
		<div class="grid grid-cols-4 gap-3 mt-3">
			<div class="prm-ls-stat">
				<div class="prm-ls-stat__icon prm-ls-stat__icon--total">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
				</div>
				<div class="prm-ls-stat__body">
					<span class="prm-ls-stat__value">{stats.total}</span>
					<span class="prm-ls-stat__label">Total</span>
				</div>
			</div>
			<div class="prm-ls-stat">
				<div class="prm-ls-stat__icon prm-ls-stat__icon--active">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
					</svg>
				</div>
				<div class="prm-ls-stat__body">
					<span class="prm-ls-stat__value">{stats.active}</span>
					<span class="prm-ls-stat__label">Active</span>
				</div>
			</div>
			<div class="prm-ls-stat">
				<div class="prm-ls-stat__icon prm-ls-stat__icon--paused">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="prm-ls-stat__body">
					<span class="prm-ls-stat__value">{stats.paused}</span>
					<span class="prm-ls-stat__label">Paused</span>
				</div>
			</div>
			<div class="prm-ls-stat">
				<div class="prm-ls-stat__icon prm-ls-stat__icon--completed">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="prm-ls-stat__body">
					<span class="prm-ls-stat__value">{stats.completed}</span>
					<span class="prm-ls-stat__label">Completed</span>
				</div>
			</div>
		</div>
	</div>

	<!-- Filters & Controls Bar -->
	<div class="px-6 py-3 prm-ls-bar flex items-center gap-3 flex-wrap">
		<!-- Search -->
		<div class="relative flex-1 min-w-[200px] max-w-xs">
			<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 prm-ls-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search projects..."
				class="w-full pl-9 pr-3 py-1.5 text-sm prm-ls-search rounded-lg focus:outline-none focus:ring-2 focus:border-transparent"
			/>
		</div>

		<!-- Status Filter Pills -->
		<div class="flex items-center gap-1 prm-ls-divider-l pl-3">
			<button
				onclick={() => { statusFilter = ''; projects.loadProjects(); }}
				class="btn-pill btn-pill-xs {statusFilter === '' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
			>
				All
			</button>
			<button
				onclick={() => { statusFilter = 'active'; projects.loadProjects('active'); }}
				class="btn-pill btn-pill-xs {statusFilter === 'active' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
			>
				Active
			</button>
			<button
				onclick={() => { statusFilter = 'paused'; projects.loadProjects('paused'); }}
				class="btn-pill btn-pill-xs {statusFilter === 'paused' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
			>
				Paused
			</button>
			<button
				onclick={() => { statusFilter = 'completed'; projects.loadProjects('completed'); }}
				class="btn-pill btn-pill-xs {statusFilter === 'completed' ? 'btn-pill-soft' : 'btn-pill-ghost'}"
			>
				Completed
			</button>
		</div>

		<!-- Type Filter -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2 {typeFilter ? 'prm-filter--active' : ''}">
				<span>{typeOptions.find(opt => opt.value === typeFilter)?.label || 'All Types'}</span>
				<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content class="z-50 min-w-[160px] prm-ls-dropdown rounded-lg shadow-lg p-1" sideOffset={4}>
					{#each typeOptions as option}
						<DropdownMenu.Item
							class="px-3 py-2 text-xs rounded prm-ls-dropdown-item cursor-pointer transition-colors {typeFilter === option.value ? 'prm-dropdown-item--active' : ''}"
							onclick={() => typeFilter = option.value}
						>
							{option.label}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Priority Filter -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-2 {priorityFilter ? 'prm-filter--active' : ''}">
				<span>{priorityOptions.find(opt => opt.value === priorityFilter)?.label || 'All Priorities'}</span>
				<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content class="z-50 min-w-[160px] prm-ls-dropdown rounded-lg shadow-lg p-1" sideOffset={4}>
					{#each priorityOptions as option}
						<DropdownMenu.Item
							class="px-3 py-2 text-xs rounded prm-ls-dropdown-item cursor-pointer transition-colors {priorityFilter === option.value ? 'prm-dropdown-item--active' : ''}"
							onclick={() => priorityFilter = option.value}
						>
							{option.label}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>

		<!-- Clear Filters -->
		{#if hasActiveFilters}
			<button onclick={clearFilters} class="btn-pill btn-pill-ghost btn-pill-xs">
				Clear filters
			</button>
		{/if}

		<!-- Spacer -->
		<div class="flex-1"></div>

		<!-- Group by Type Toggle -->
		<label class="flex items-center gap-2 text-xs prm-ls-label cursor-pointer">
			<input
				type="checkbox"
				bind:checked={groupByType}
				class="w-3.5 h-3.5 rounded prm-ls-checkbox"
			/>
			Group by type
		</label>

		<!-- View Mode Toggle -->
		<div class="flex items-center gap-0.5 prm-ls-toggle-border rounded-lg overflow-hidden p-0.5">
			<button
				onclick={() => viewMode = 'grid'}
				class="btn-pill btn-pill-icon btn-pill-xs {viewMode === 'grid' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="Grid view"
				aria-label="Grid view"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
				</svg>
			</button>
			<button
				onclick={() => viewMode = 'list'}
				class="btn-pill btn-pill-icon btn-pill-xs {viewMode === 'list' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="List view"
				aria-label="List view"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
			<button
				onclick={() => viewMode = 'kanban'}
				class="btn-pill btn-pill-icon btn-pill-xs {viewMode === 'kanban' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="Kanban view"
				aria-label="Kanban view"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
				</svg>
			</button>
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto p-6">
		{#if $projects.loading}
			<div class="flex items-center justify-center h-48">
				<div class="animate-spin h-8 w-8 border-2 prm-ls-spinner rounded-full"></div>
			</div>
		{:else if $projects.projects.length === 0}
			<div class="flex flex-col items-center justify-center h-64 text-center">
				<div class="w-16 h-16 rounded-2xl prm-ls-empty-bg flex items-center justify-center mb-4">
					<svg class="w-8 h-8 prm-ls-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
				</div>
				<h3 class="text-lg font-medium prm-ls-title mb-1">No projects yet</h3>
				<p class="text-sm prm-ls-muted mb-4">Get started by creating your first project</p>
				<button onclick={() => showNewProject = true} class="btn-pill btn-pill-primary btn-pill-sm">
					Create Project
				</button>
			</div>
		{:else if filteredProjects.length === 0}
			<div class="flex flex-col items-center justify-center h-48 text-center">
				<p class="prm-ls-muted mb-2">No projects match your filters</p>
				<button onclick={clearFilters} class="btn-pill btn-pill-ghost btn-pill-sm underline">
					Clear all filters
				</button>
			</div>
		{:else if viewMode === 'kanban'}
			<!-- Kanban View -->
			<div class="flex gap-4 h-full min-h-0">
				<!-- Active Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col prm-ls-column overflow-hidden">
					<div class="px-4 py-3 prm-ls-column__header flex items-center gap-2">
						<span class="prm-dot prm-dot--active"></span>
						<span class="font-medium prm-ls-title">Active</span>
						<span class="prm-ls-kanban-count ml-auto">{groupedByStatus.active.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.active as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 prm-ls-kanban-card rounded-lg transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium prm-ls-title line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{getTypeIcon(project.project_type)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs prm-ls-muted mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs prm-ls-icon">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.active.length === 0}
							<p class="text-xs prm-ls-icon text-center py-4">No active projects</p>
						{/if}
					</div>
				</div>

				<!-- Paused Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col prm-ls-column overflow-hidden">
					<div class="px-4 py-3 prm-ls-column__header flex items-center gap-2">
						<span class="prm-dot prm-dot--paused"></span>
						<span class="font-medium prm-ls-title">Paused</span>
						<span class="prm-ls-kanban-count ml-auto">{groupedByStatus.paused.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.paused as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 prm-ls-kanban-card rounded-lg transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium prm-ls-title line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{getTypeIcon(project.project_type)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs prm-ls-muted mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs prm-ls-icon">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.paused.length === 0}
							<p class="text-xs prm-ls-icon text-center py-4">No paused projects</p>
						{/if}
					</div>
				</div>

				<!-- Completed Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col prm-ls-column overflow-hidden">
					<div class="px-4 py-3 prm-ls-column__header flex items-center gap-2">
						<span class="prm-dot prm-dot--completed"></span>
						<span class="font-medium prm-ls-title">Completed</span>
						<span class="prm-ls-kanban-count ml-auto">{groupedByStatus.completed.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.completed as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 prm-ls-kanban-card rounded-lg transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium prm-ls-title line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{getTypeIcon(project.project_type)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs prm-ls-muted mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs prm-ls-icon">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.completed.length === 0}
							<p class="text-xs prm-ls-icon text-center py-4">No completed projects</p>
						{/if}
					</div>
				</div>
			</div>
		{:else if viewMode === 'list'}
			<!-- List View -->
			<div class="prm-ls-column overflow-hidden">
				<table class="w-full">
					<thead class="prm-ls-table-head">
						<tr>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Project</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Type</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Status</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Priority</th>
							<th class="text-left text-xs font-medium prm-ls-muted uppercase tracking-wider px-4 py-3">Updated</th>
						</tr>
					</thead>
					<tbody class="prm-ls-table-body">
						{#each filteredProjects as project}
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
									<span class="text-xs px-2 py-1 rounded-full {getTypeColor(project.project_type)}">
										{getTypeIcon(project.project_type)} {getTypeLabel(project.project_type)}
									</span>
								</td>
								<td class="px-4 py-3">
									<span class="text-xs font-medium px-2.5 py-1 rounded-full {getStatusColor(project.status)}">
										{project.status}
									</span>
								</td>
								<td class="px-4 py-3">
									<span class="text-xs font-medium {getPriorityColor(project.priority)} capitalize">
										{'●'.repeat(getPriorityIcon(project.priority) + 1)} {project.priority}
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
							<span class="text-lg">{getTypeIcon('internal')}</span>
						<h2 class="font-semibold prm-ls-title">Internal Projects</h2>
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
							<span class="text-lg">{getTypeIcon('client_work')}</span>
						<h2 class="font-semibold prm-ls-title">Client Work</h2>
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
							<span class="text-lg">{getTypeIcon('learning')}</span>
						<h2 class="font-semibold prm-ls-title">Learning</h2>
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
							<span class="text-lg">{getTypeIcon('other')}</span>
						<h2 class="font-semibold prm-ls-title">Other</h2>
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
				{#each filteredProjects as project}
					{@render projectCard(project)}
				{/each}
			</div>
		{/if}
	</div>
</div>

{#snippet projectCard(project: Project)}
	<a href="/projects/{project.id}{embedSuffix}" class="block prm-ls-card prm-ls-card--{project.project_type} transition-all duration-200 cursor-pointer group">
		<div class="p-4">
			<div class="flex items-start justify-between mb-2">
				<div class="flex items-center gap-2 min-w-0">
					<span class="text-base flex-shrink-0">{getTypeIcon(project.project_type)}</span>
					<h3 class="font-medium prm-ls-title line-clamp-1 text-sm">{project.name}</h3>
				</div>
				<span class="text-xs font-medium px-2 py-0.5 rounded-full flex-shrink-0 ml-2 {getStatusColor(project.status)}">
					{project.status}
				</span>
			</div>
			{#if project.client_name}
				<p class="text-xs prm-ls-muted mb-1 ml-7">{project.client_name}</p>
			{/if}
			{#if project.description}
				<p class="text-xs prm-ls-icon line-clamp-2 mb-3 ml-7">{project.description}</p>
			{:else}
				<div class="mb-3"></div>
			{/if}
			<div class="flex items-center justify-between text-xs pt-3 prm-ls-card__footer">
				<div class="flex items-center gap-2">
					<span class="font-medium {getPriorityColor(project.priority)} capitalize">
						{'●'.repeat(getPriorityIcon(project.priority) + 1)} {project.priority}
					</span>
					<span class="px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">
						{getTypeLabel(project.project_type)}
					</span>
				</div>
				<span class="prm-ls-icon">{formatDate(project.updated_at)}</span>
			</div>
		</div>
	</a>
{/snippet}

<!-- New Project Dialog -->
<Dialog.Root bind:open={showNewProject}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 prm-ls-dialog shadow-xl p-6 w-full max-w-lg z-50 max-h-[90vh] overflow-y-auto">
			<Dialog.Title class="text-lg font-semibold prm-ls-title mb-4">New Project</Dialog.Title>

			<form onsubmit={handleCreateProject} class="space-y-4">
				<!-- Icon + Name Row -->
				<div class="flex gap-3">
					<!-- Icon Picker -->
					<Popover.Root bind:open={showIconPicker}>
						<Popover.Trigger class="w-14 h-14 rounded-xl prm-ls-icon-picker transition-colors flex items-center justify-center text-2xl flex-shrink-0 border-2 border-transparent">
							{newProject.icon}
						</Popover.Trigger>
						<Popover.Content class="z-[60] prm-ls-popover rounded-xl shadow-lg p-3 w-64">
							<p class="text-xs font-medium prm-ls-muted mb-2">Choose Icon</p>
							<div class="grid grid-cols-8 gap-1">
								{#each projectIcons as icon}
									<button
										type="button"
										onclick={() => { newProject.icon = icon; showIconPicker = false; }}
										class="w-7 h-7 rounded prm-ls-icon-btn flex items-center justify-center text-lg transition-colors {newProject.icon === icon ? 'prm-icon-picker--active' : ''}"
									>
										{icon}
									</button>
								{/each}
							</div>
						</Popover.Content>
					</Popover.Root>

					<div class="flex-1">
						<label for="name" class="block text-sm font-medium prm-ls-label mb-1">Name</label>
						<input
							id="name"
							type="text"
							bind:value={newProject.name}
							class="input input-square"
							placeholder="Project name"
							required
						/>
					</div>
				</div>

				<!-- Type Selection with Visual Cards -->
				<div>
					<label class="block text-sm font-medium prm-ls-label mb-2">Type</label>
					<div class="grid grid-cols-3 gap-2">
						<button
							type="button"
							onclick={() => newProject.project_type = 'internal'}
							class="p-3 rounded-xl border-2 transition-all text-center {newProject.project_type === 'internal' ? 'prm-type-card--active-internal' : 'prm-ls-type-btn'}"
						>
							<span class="text-xl block mb-1">{getTypeEmoji('internal')}</span>
							<span class="text-xs font-medium {newProject.project_type === 'internal' ? 'prm-type-card__label--active' : 'prm-ls-label'}">Internal</span>
						</button>
						<button
							type="button"
							onclick={() => newProject.project_type = 'client_work'}
							class="p-3 rounded-xl border-2 transition-all text-center {newProject.project_type === 'client_work' ? 'prm-type-card--active-client' : 'prm-ls-type-btn'}"
						>
							<span class="text-xl block mb-1">{getTypeEmoji('client_work')}</span>
							<span class="text-xs font-medium {newProject.project_type === 'client_work' ? 'prm-type-card__label--active' : 'prm-ls-label'}">Client Work</span>
						</button>
						<button
							type="button"
							onclick={() => newProject.project_type = 'learning'}
							class="p-3 rounded-xl border-2 transition-all text-center {newProject.project_type === 'learning' ? 'prm-type-card--active-learning' : 'prm-ls-type-btn'}"
						>
							<span class="text-xl block mb-1">{getTypeEmoji('learning')}</span>
							<span class="text-xs font-medium {newProject.project_type === 'learning' ? 'prm-type-card__label--active' : 'prm-ls-label'}">Learning</span>
						</button>
					</div>
				</div>

				<!-- Priority Selection with Visual Indicators -->
				<div>
					<label class="block text-sm font-medium prm-ls-label mb-2">Priority</label>
					<div class="flex gap-2">
						{#each ['low', 'medium', 'high', 'critical'] as priority}
							<button
								type="button"
								onclick={() => newProject.priority = priority as 'low' | 'medium' | 'high' | 'critical'}
								class="flex-1 py-2 px-3 rounded-lg border-2 transition-all text-center text-sm font-medium {newProject.priority === priority ? 'prm-priority-btn--active' : 'prm-ls-type-btn prm-ls-label'}"
							>
								<span class="mr-1">{getPriorityEmoji(priority)}</span>
								<span class="capitalize">{priority}</span>
							</button>
						{/each}
					</div>
				</div>

				<!-- Client Dropdown -->
				<div>
					<label class="block text-sm font-medium prm-ls-label mb-1">Client (optional)</label>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger class="w-full flex items-center justify-between gap-2 px-3 py-2 text-sm prm-ls-client-trigger transition-colors">
							<span class={newProject.client_name ? 'prm-ls-title' : 'prm-ls-icon'}>
								{newProject.client_name || 'No client'}
							</span>
							<svg class="w-4 h-4 prm-ls-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
							</svg>
						</DropdownMenu.Trigger>
						<DropdownMenu.Portal>
							<DropdownMenu.Content class="z-[70] w-[var(--bits-dropdown-trigger-width)] prm-ls-dropdown rounded-lg shadow-lg p-1 max-h-[200px] overflow-y-auto" sideOffset={4}>
								<DropdownMenu.Item
									class="px-3 py-2 text-sm rounded prm-ls-dropdown-item cursor-pointer transition-colors {!newProject.client_name ? 'prm-ls-empty-bg font-medium' : ''}"
									onclick={() => newProject.client_name = ''}
								>
									No client
								</DropdownMenu.Item>
								{#each clients as client}
									<DropdownMenu.Item
										class="px-3 py-2 text-sm rounded prm-ls-dropdown-item cursor-pointer transition-colors {newProject.client_name === client.name ? 'prm-ls-empty-bg font-medium' : ''}"
										onclick={() => newProject.client_name = client.name}
									>
										{client.name}
									</DropdownMenu.Item>
								{/each}
							</DropdownMenu.Content>
						</DropdownMenu.Portal>
					</DropdownMenu.Root>
					{#if clients.length === 0}
						<p class="text-xs prm-ls-icon mt-1">No clients yet. Create one in the Clients section.</p>
					{/if}
				</div>

				<!-- Description -->
				<div>
					<label for="description" class="block text-sm font-medium prm-ls-label mb-1">Description</label>
					<textarea
						id="description"
						bind:value={newProject.description}
						class="input input-square resize-none"
						rows="3"
						placeholder="What's this project about?"
					></textarea>
				</div>

				<!-- Advanced Options Toggle -->
				<button
					type="button"
					onclick={() => showAdvancedOptions = !showAdvancedOptions}
					class="flex items-center gap-2 text-sm prm-ls-muted prm-ls-adv-toggle"
				>
					<svg
						class="w-4 h-4 transition-transform {showAdvancedOptions ? 'rotate-90' : ''}"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
					Advanced Options
				</button>

				{#if showAdvancedOptions}
					<div class="space-y-4 pl-4 prm-ls-adv-border">
						<p class="text-xs prm-ls-icon">Additional options like team assignment and tags will be available after creating the project.</p>
					</div>
				{/if}

				{#if createError}
					<div class="prm-create-error">
						{createError}
					</div>
				{/if}

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => { showNewProject = false; showAdvancedOptions = false; }} class="btn-pill btn-pill-secondary flex-1">
						Cancel
					</button>
					<button type="submit" class="btn-pill btn-pill-primary btn-pill-sm flex-1">
						Create Project
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	/* Page & Layout */
	.prm-ls-page { background: var(--dbg2, rgba(249,250,251,.5)); }
	.prm-ls-bar { background: var(--dbg, #fff); border-bottom: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-title { color: var(--dt, #111); }
	.prm-ls-muted { color: var(--dt3, #6b7280); }
	.prm-ls-icon { color: var(--dt4, #9ca3af); }
	.prm-ls-label { color: var(--dt2, #4b5563); }
	.prm-ls-spinner { border-color: var(--dt, #111); border-top-color: transparent; }

	/* Search */
	.prm-ls-search { border: 1px solid var(--dbd, #e5e7eb); background: var(--dbg, #fff); color: var(--dt, #111); }
	.prm-ls-search:focus { box-shadow: 0 0 0 2px var(--dt, #111); }
	.prm-ls-divider-l { border-left: 1px solid var(--dbd, #e5e7eb); }

	/* Dropdowns */
	.prm-ls-dropdown { background: var(--dbg, #fff); border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-dropdown-item:hover { background: var(--dbg3, #f3f4f6); }

	/* Controls */
	.prm-ls-checkbox { border-color: var(--dbd, #d1d5db); color: var(--dt, #111); }
	.prm-ls-toggle-border { border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-empty-bg { background: var(--dbg3, #f3f4f6); }

	/* Stat Cards */
	.prm-ls-stat { display: flex; align-items: center; gap: 0.75rem; padding: 0.625rem 0.75rem; background: var(--dbg2, #f9fafb); border-radius: 0.5rem; border: 1px solid var(--dbd2, #f3f4f6); }
	.prm-ls-stat__icon { width: 2rem; height: 2rem; border-radius: 0.375rem; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
	.prm-ls-stat__icon--total { background: color-mix(in srgb, #8b5cf6 12%, var(--dbg, #fff)); color: #8b5cf6; }
	.prm-ls-stat__icon--active { background: color-mix(in srgb, #22c55e 12%, var(--dbg, #fff)); color: #22c55e; }
	.prm-ls-stat__icon--paused { background: color-mix(in srgb, #f59e0b 12%, var(--dbg, #fff)); color: #f59e0b; }
	.prm-ls-stat__icon--completed { background: color-mix(in srgb, #3b82f6 12%, var(--dbg, #fff)); color: #3b82f6; }
	.prm-ls-stat__body { display: flex; flex-direction: column; }
	.prm-ls-stat__value { font-size: 1.125rem; font-weight: 700; line-height: 1.2; color: var(--dt, #111); }
	.prm-ls-stat__label { font-size: 0.6875rem; color: var(--dt3, #6b7280); }

	/* Kanban & Cards */
	.prm-ls-column { background: var(--dbg, #fff); border-radius: 0.75rem; border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-column__header { border-bottom: 1px solid var(--dbd2, #f3f4f6); }
	.prm-ls-kanban-card { background: var(--dbg2, #f9fafb); }
	.prm-ls-kanban-card:hover { background: var(--dbg3, #f3f4f6); }
	.prm-ls-kanban-count { display: inline-flex; align-items: center; justify-content: center; min-width: 1.25rem; height: 1.25rem; padding: 0 0.375rem; font-size: 0.6875rem; font-weight: 600; border-radius: 9999px; background: var(--dbg3, #f3f4f6); color: var(--dt3, #6b7280); }
	.prm-ls-card { background: var(--dbg, #fff); border-radius: 0.75rem; border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-card:hover { box-shadow: 0 4px 6px rgba(0,0,0,.1); transform: scale(1.02); }
	.prm-ls-card--internal { border-top: 3px solid #8b5cf6; }
	.prm-ls-card--client_work { border-top: 3px solid #3b82f6; }
	.prm-ls-card--learning { border-top: 3px solid #14b8a6; }
	.prm-ls-card__footer { border-top: 1px solid var(--dbd2, #f3f4f6); }

	/* List view */
	.prm-ls-table-head { background: var(--dbg2, #f9fafb); border-bottom: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-table-body { }
	.prm-ls-table-body > :global(tr + tr) { border-top: 1px solid var(--dbd2, #f3f4f6); }
	.prm-ls-table-row:hover { background: var(--dbg2, #f9fafb); }

	/* Status/Priority/Type defaults */
	.prm-ls-status-default { background: var(--dbg3, #f3f4f6); color: var(--dt2, #4b5563); }
	.prm-ls-priority-default { color: var(--dt4, #9ca3af); }
	.prm-ls-type-default { color: var(--dt2, #4b5563); background: var(--dbg2, #f9fafb); }

	/* Dialog */
	.prm-ls-dialog { background: var(--dbg, #fff); border-radius: 1rem; }
	.prm-ls-popover { background: var(--dbg, #fff); border: 1px solid var(--dbd, #e5e7eb); }
	.prm-ls-icon-picker { background: var(--dbg3, #f3f4f6); }
	.prm-ls-icon-picker:hover { background: var(--dbg3, #e5e7eb); border-color: var(--dbd, #d1d5db); }
	.prm-ls-icon-btn:hover { background: var(--dbg3, #f3f4f6); }
	.prm-ls-type-btn { border-color: var(--dbd, #e5e7eb); }
	.prm-ls-type-btn:hover { border-color: var(--dbd, #d1d5db); }
	.prm-ls-client-trigger { border: 1px solid var(--dbd, #e5e7eb); border-radius: 0.5rem; background: var(--dbg, #fff); }
	.prm-ls-client-trigger:hover { background: var(--dbg2, #f9fafb); }
	.prm-ls-adv-toggle:hover { color: var(--dt2, #374151); }
	.prm-ls-adv-border { border-left: 2px solid var(--dbd2, #f3f4f6); }

	/* Status Pills (global — used by utility import) */
	:global(.prm-status) { display: inline-block; padding: 0.125rem 0.5rem; font-size: 0.6875rem; font-weight: 600; border-radius: 9999px; border: 1px solid transparent; }
	:global(.prm-status--active) { background: color-mix(in srgb, #22c55e 15%, var(--dbg)); color: #22c55e; border-color: color-mix(in srgb, #22c55e 25%, var(--dbd)); }
	:global(.prm-status--paused) { background: color-mix(in srgb, #f59e0b 15%, var(--dbg)); color: #f59e0b; border-color: color-mix(in srgb, #f59e0b 25%, var(--dbd)); }
	:global(.prm-status--completed) { background: color-mix(in srgb, #3b82f6 15%, var(--dbg)); color: #3b82f6; border-color: color-mix(in srgb, #3b82f6 25%, var(--dbd)); }
	:global(.prm-status--archived) { background: var(--dbg3); color: var(--dt3); border-color: var(--dbd); }

	/* Priority Pills (global) */
	:global(.prm-priority) { display: inline-block; padding: 0.125rem 0.5rem; font-size: 0.6875rem; font-weight: 600; border-radius: 9999px; }
	:global(.prm-priority--critical) { color: #ef4444; background: color-mix(in srgb, #ef4444 10%, var(--dbg)); }
	:global(.prm-priority--high) { color: #f97316; background: color-mix(in srgb, #f97316 10%, var(--dbg)); }
	:global(.prm-priority--medium) { color: #eab308; background: color-mix(in srgb, #eab308 10%, var(--dbg)); }
	:global(.prm-priority--low) { color: #22c55e; background: color-mix(in srgb, #22c55e 10%, var(--dbg)); }
	:global(.prm-priority--default) { color: var(--dt3); background: var(--dbg2); }

	/* Type Pills (global) */
	:global(.prm-type) { display: inline-block; padding: 0.125rem 0.5rem; font-size: 0.6875rem; font-weight: 600; border-radius: 6px; }
	:global(.prm-type--internal) { color: #8b5cf6; background: color-mix(in srgb, #8b5cf6 12%, var(--dbg)); }
	:global(.prm-type--client) { color: #3b82f6; background: color-mix(in srgb, #3b82f6 12%, var(--dbg)); }
	:global(.prm-type--learning) { color: #14b8a6; background: color-mix(in srgb, #14b8a6 12%, var(--dbg)); }

	/* Status Dots */
	.prm-dot { width: 0.625rem; height: 0.625rem; border-radius: 50%; flex-shrink: 0; }
	.prm-dot--active { background: #22c55e; }
	.prm-dot--paused { background: #f59e0b; }
	.prm-dot--completed { background: #3b82f6; }

	/* Filter active state */
	.prm-filter--active { box-shadow: 0 0 0 2px var(--dt3); }
	.prm-dropdown-item--active { background: var(--dbg3); font-weight: 600; }

	/* Icon picker active */
	.prm-icon-picker--active { background: var(--dbg3); box-shadow: 0 0 0 2px var(--dt); }

	/* Type card active states */
	.prm-type-card--active-internal { border-color: #8b5cf6; background: color-mix(in srgb, #8b5cf6 10%, var(--dbg)); }
	.prm-type-card--active-client { border-color: #3b82f6; background: color-mix(in srgb, #3b82f6 10%, var(--dbg)); }
	.prm-type-card--active-learning { border-color: #14b8a6; background: color-mix(in srgb, #14b8a6 10%, var(--dbg)); }
	.prm-type-card__label--active { color: var(--dt); }

	/* Priority button active */
	.prm-priority-btn--active { border-color: var(--dt); background: var(--dt); color: var(--dbg); }

	/* Create error */
	.prm-create-error { font-size: 0.875rem; color: #ef4444; background: color-mix(in srgb, #ef4444 10%, var(--dbg)); padding: 0.5rem 0.75rem; border-radius: 0.75rem; }
</style>
