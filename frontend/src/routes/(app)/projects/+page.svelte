<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { projects } from '$lib/stores/projects';
	import { api, type ClientListResponse } from '$lib/api';
	import { Dialog, Popover } from 'bits-ui';
	import type { Project } from '$lib/api';

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

	// Clients for dropdown
	let clients = $state<ClientListResponse[]>([]);
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
		await Promise.all([
			projects.loadProjects(),
			loadClients()
		]);
	});

	async function loadClients() {
		try {
			clients = await api.getClients();
		} catch (err) {
			console.error('Error loading clients:', err);
		}
	}

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
			case 'active': return 'bg-emerald-100 text-emerald-700';
			case 'paused': return 'bg-amber-100 text-amber-700';
			case 'completed': return 'bg-blue-100 text-blue-700';
			case 'archived': return 'bg-gray-100 text-gray-600';
			default: return 'bg-gray-100 text-gray-600';
		}
	}

	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'critical': return 'text-red-600';
			case 'high': return 'text-orange-500';
			case 'medium': return 'text-yellow-500';
			case 'low': return 'text-green-500';
			default: return 'text-gray-400';
		}
	}

	function getPriorityIcon(priority: string) {
		const count = priority === 'critical' ? 3 : priority === 'high' ? 2 : priority === 'medium' ? 1 : 0;
		return count;
	}

	function getTypeColor(type: string) {
		switch (type) {
			case 'internal': return 'text-purple-600 bg-purple-50';
			case 'client_work': return 'text-blue-600 bg-blue-50';
			case 'learning': return 'text-teal-600 bg-teal-50';
			default: return 'text-gray-600 bg-gray-50';
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

	let hasActiveFilters = $derived(statusFilter || typeFilter || priorityFilter || searchQuery);
</script>

<div class="h-full flex flex-col bg-gray-50/50">
	<!-- Header -->
	<div class="px-6 py-4 bg-white border-b border-gray-200">
		<div class="flex items-center justify-between mb-4">
			<div>
				<h1 class="text-xl font-semibold text-gray-900">Projects</h1>
				<p class="text-sm text-gray-500 mt-0.5">Manage your work and track progress</p>
			</div>
			<button onclick={() => showNewProject = true} class="btn-pill btn-pill-primary flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				New Project
			</button>
		</div>

		<!-- Stats Row -->
		<div class="flex gap-6 text-sm">
			<div class="flex items-center gap-2">
				<span class="text-gray-500">Total:</span>
				<span class="font-medium text-gray-900">{stats.total}</span>
			</div>
			<div class="flex items-center gap-2">
				<span class="w-2 h-2 rounded-full bg-emerald-500"></span>
				<span class="text-gray-500">Active:</span>
				<span class="font-medium text-gray-900">{stats.active}</span>
			</div>
			<div class="flex items-center gap-2">
				<span class="w-2 h-2 rounded-full bg-amber-500"></span>
				<span class="text-gray-500">Paused:</span>
				<span class="font-medium text-gray-900">{stats.paused}</span>
			</div>
			<div class="flex items-center gap-2">
				<span class="w-2 h-2 rounded-full bg-blue-500"></span>
				<span class="text-gray-500">Completed:</span>
				<span class="font-medium text-gray-900">{stats.completed}</span>
			</div>
		</div>
	</div>

	<!-- Filters & Controls Bar -->
	<div class="px-6 py-3 bg-white border-b border-gray-200 flex items-center gap-3 flex-wrap">
		<!-- Search -->
		<div class="relative flex-1 min-w-[200px] max-w-xs">
			<svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Search projects..."
				class="w-full pl-9 pr-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent"
			/>
		</div>

		<!-- Status Filter Pills -->
		<div class="btn-pill-group border-l border-gray-200 pl-3">
			<button
				onclick={() => { statusFilter = ''; projects.loadProjects(); }}
				class="btn-pill btn-pill-sm {statusFilter === '' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
			>
				All
			</button>
			<button
				onclick={() => { statusFilter = 'active'; projects.loadProjects('active'); }}
				class="btn-pill btn-pill-sm {statusFilter === 'active' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
			>
				Active
			</button>
			<button
				onclick={() => { statusFilter = 'paused'; projects.loadProjects('paused'); }}
				class="btn-pill btn-pill-sm {statusFilter === 'paused' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
			>
				Paused
			</button>
			<button
				onclick={() => { statusFilter = 'completed'; projects.loadProjects('completed'); }}
				class="btn-pill btn-pill-sm {statusFilter === 'completed' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
			>
				Completed
			</button>
		</div>

		<!-- Type Filter -->
		<select
			bind:value={typeFilter}
			class="btn-pill btn-pill-secondary btn-pill-sm"
		>
			<option value="">All Types</option>
			<option value="internal">Internal</option>
			<option value="client_work">Client Work</option>
			<option value="learning">Learning</option>
		</select>

		<!-- Priority Filter -->
		<select
			bind:value={priorityFilter}
			class="btn-pill btn-pill-secondary btn-pill-sm"
		>
			<option value="">All Priorities</option>
			<option value="critical">Critical</option>
			<option value="high">High</option>
			<option value="medium">Medium</option>
			<option value="low">Low</option>
		</select>

		<!-- Clear Filters -->
		{#if hasActiveFilters}
			<button onclick={clearFilters} class="text-xs text-gray-500 hover:text-gray-700 underline">
				Clear filters
			</button>
		{/if}

		<!-- Spacer -->
		<div class="flex-1"></div>

		<!-- Group by Type Toggle -->
		<label class="flex items-center gap-2 text-xs text-gray-600 cursor-pointer">
			<input
				type="checkbox"
				bind:checked={groupByType}
				class="w-3.5 h-3.5 rounded border-gray-300 text-gray-900 focus:ring-gray-900"
			/>
			Group by type
		</label>

		<!-- View Mode Toggle -->
		<div class="btn-pill-group">
			<button
				onclick={() => viewMode = 'grid'}
				class="btn-pill btn-pill-icon btn-pill-sm {viewMode === 'grid' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="Grid view"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
				</svg>
			</button>
			<button
				onclick={() => viewMode = 'list'}
				class="btn-pill btn-pill-icon btn-pill-sm {viewMode === 'list' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="List view"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
			<button
				onclick={() => viewMode = 'kanban'}
				class="btn-pill btn-pill-icon btn-pill-sm {viewMode === 'kanban' ? 'btn-pill-primary' : 'btn-pill-ghost'}"
				title="Kanban view"
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
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
			</div>
		{:else if $projects.projects.length === 0}
			<div class="flex flex-col items-center justify-center h-64 text-center">
				<div class="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
				</div>
				<h3 class="text-lg font-medium text-gray-900 mb-1">No projects yet</h3>
				<p class="text-sm text-gray-500 mb-4">Get started by creating your first project</p>
				<button onclick={() => showNewProject = true} class="btn-pill btn-pill-primary">
					Create Project
				</button>
			</div>
		{:else if filteredProjects().length === 0}
			<div class="flex flex-col items-center justify-center h-48 text-center">
				<p class="text-gray-500 mb-2">No projects match your filters</p>
				<button onclick={clearFilters} class="text-sm text-gray-900 underline">
					Clear all filters
				</button>
			</div>
		{:else if viewMode === 'kanban'}
			<!-- Kanban View -->
			<div class="flex gap-4 h-full min-h-0">
				<!-- Active Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col bg-white rounded-xl border border-gray-200 overflow-hidden">
					<div class="px-4 py-3 border-b border-gray-100 flex items-center gap-2">
						<span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
						<span class="font-medium text-gray-900">Active</span>
						<span class="text-xs text-gray-400 ml-auto">{groupedByStatus.active.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.active as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium text-gray-900 line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{getTypeIcon(project.project_type)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs text-gray-500 mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs text-gray-400">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.active.length === 0}
							<p class="text-xs text-gray-400 text-center py-4">No active projects</p>
						{/if}
					</div>
				</div>

				<!-- Paused Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col bg-white rounded-xl border border-gray-200 overflow-hidden">
					<div class="px-4 py-3 border-b border-gray-100 flex items-center gap-2">
						<span class="w-2.5 h-2.5 rounded-full bg-amber-500"></span>
						<span class="font-medium text-gray-900">Paused</span>
						<span class="text-xs text-gray-400 ml-auto">{groupedByStatus.paused.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.paused as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium text-gray-900 line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{getTypeIcon(project.project_type)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs text-gray-500 mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs text-gray-400">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.paused.length === 0}
							<p class="text-xs text-gray-400 text-center py-4">No paused projects</p>
						{/if}
					</div>
				</div>

				<!-- Completed Column -->
				<div class="flex-1 min-w-[280px] max-w-[350px] flex flex-col bg-white rounded-xl border border-gray-200 overflow-hidden">
					<div class="px-4 py-3 border-b border-gray-100 flex items-center gap-2">
						<span class="w-2.5 h-2.5 rounded-full bg-blue-500"></span>
						<span class="font-medium text-gray-900">Completed</span>
						<span class="text-xs text-gray-400 ml-auto">{groupedByStatus.completed.length}</span>
					</div>
					<div class="flex-1 overflow-y-auto p-2 space-y-2">
						{#each groupedByStatus.completed as project}
							<a href="/projects/{project.id}{embedSuffix}" class="block p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
								<div class="flex items-start justify-between mb-1">
									<span class="text-sm font-medium text-gray-900 line-clamp-1">{project.name}</span>
									<span class="text-xs px-1.5 py-0.5 rounded {getTypeColor(project.project_type)}">{getTypeIcon(project.project_type)}</span>
								</div>
								{#if project.client_name}
									<p class="text-xs text-gray-500 mb-1">{project.client_name}</p>
								{/if}
								<div class="flex items-center justify-between text-xs text-gray-400">
									<span class={getPriorityColor(project.priority)}>
										{'●'.repeat(getPriorityIcon(project.priority) + 1)}
									</span>
									<span>{formatDate(project.updated_at)}</span>
								</div>
							</a>
						{/each}
						{#if groupedByStatus.completed.length === 0}
							<p class="text-xs text-gray-400 text-center py-4">No completed projects</p>
						{/if}
					</div>
				</div>
			</div>
		{:else if viewMode === 'list'}
			<!-- List View -->
			<div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
				<table class="w-full">
					<thead class="bg-gray-50 border-b border-gray-200">
						<tr>
							<th class="text-left text-xs font-medium text-gray-500 uppercase tracking-wider px-4 py-3">Project</th>
							<th class="text-left text-xs font-medium text-gray-500 uppercase tracking-wider px-4 py-3">Type</th>
							<th class="text-left text-xs font-medium text-gray-500 uppercase tracking-wider px-4 py-3">Status</th>
							<th class="text-left text-xs font-medium text-gray-500 uppercase tracking-wider px-4 py-3">Priority</th>
							<th class="text-left text-xs font-medium text-gray-500 uppercase tracking-wider px-4 py-3">Updated</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each filteredProjects() as project}
							<tr class="hover:bg-gray-50 transition-colors cursor-pointer" onclick={() => window.location.href = `/projects/${project.id}${embedSuffix}`}>
								<td class="px-4 py-3">
									<div>
										<span class="font-medium text-gray-900">{project.name}</span>
										{#if project.client_name}
											<span class="text-gray-400 ml-2">· {project.client_name}</span>
										{/if}
									</div>
									{#if project.description}
										<p class="text-sm text-gray-500 line-clamp-1 mt-0.5">{project.description}</p>
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
								<td class="px-4 py-3 text-sm text-gray-500">
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
							<h2 class="font-semibold text-gray-900">Internal Projects</h2>
							<span class="text-xs text-gray-400">({groupedByType.internal.length})</span>
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
							<h2 class="font-semibold text-gray-900">Client Work</h2>
							<span class="text-xs text-gray-400">({groupedByType.client_work.length})</span>
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
							<h2 class="font-semibold text-gray-900">Learning</h2>
							<span class="text-xs text-gray-400">({groupedByType.learning.length})</span>
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
							<h2 class="font-semibold text-gray-900">Other</h2>
							<span class="text-xs text-gray-400">({groupedByType.other.length})</span>
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
	<a href="/projects/{project.id}{embedSuffix}" class="block bg-white rounded-xl border border-gray-200 p-4 hover:shadow-md hover:border-gray-300 transition-all duration-200 cursor-pointer group">
		<div class="flex items-start justify-between mb-3">
			<span class="text-xs font-medium px-2.5 py-1 rounded-full {getStatusColor(project.status)}">
				{project.status}
			</span>
			<span class="text-xs px-2 py-1 rounded-full {getTypeColor(project.project_type)}">
				{getTypeIcon(project.project_type)} {getTypeLabel(project.project_type)}
			</span>
		</div>
		<h3 class="font-medium text-gray-900 group-hover:text-gray-700 mb-1 line-clamp-1">{project.name}</h3>
		{#if project.client_name}
			<p class="text-sm text-gray-500 mb-1">{project.client_name}</p>
		{/if}
		{#if project.description}
			<p class="text-sm text-gray-400 line-clamp-2 mb-3">{project.description}</p>
		{:else}
			<div class="mb-3"></div>
		{/if}
		<div class="flex items-center justify-between text-xs pt-3 border-t border-gray-100">
			<span class="font-medium {getPriorityColor(project.priority)} capitalize">
				{'●'.repeat(getPriorityIcon(project.priority) + 1)} {project.priority}
			</span>
			<span class="text-gray-400">Updated {formatDate(project.updated_at)}</span>
		</div>
	</a>
{/snippet}

<!-- New Project Dialog -->
<Dialog.Root bind:open={showNewProject}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-lg z-50 max-h-[90vh] overflow-y-auto">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">New Project</Dialog.Title>

			<form onsubmit={handleCreateProject} class="space-y-4">
				<!-- Icon + Name Row -->
				<div class="flex gap-3">
					<!-- Icon Picker -->
					<Popover.Root bind:open={showIconPicker}>
						<Popover.Trigger class="w-14 h-14 rounded-xl bg-gray-100 hover:bg-gray-200 transition-colors flex items-center justify-center text-2xl flex-shrink-0 border-2 border-transparent hover:border-gray-300">
							{newProject.icon}
						</Popover.Trigger>
						<Popover.Content class="z-[60] bg-white rounded-xl shadow-lg border border-gray-200 p-3 w-64">
							<p class="text-xs font-medium text-gray-500 mb-2">Choose Icon</p>
							<div class="grid grid-cols-8 gap-1">
								{#each projectIcons as icon}
									<button
										type="button"
										onclick={() => { newProject.icon = icon; showIconPicker = false; }}
										class="w-7 h-7 rounded hover:bg-gray-100 flex items-center justify-center text-lg transition-colors {newProject.icon === icon ? 'bg-purple-100 ring-2 ring-purple-500' : ''}"
									>
										{icon}
									</button>
								{/each}
							</div>
						</Popover.Content>
					</Popover.Root>

					<div class="flex-1">
						<label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
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
					<label class="block text-sm font-medium text-gray-700 mb-2">Type</label>
					<div class="grid grid-cols-3 gap-2">
						<button
							type="button"
							onclick={() => newProject.project_type = 'internal'}
							class="p-3 rounded-xl border-2 transition-all text-center {newProject.project_type === 'internal' ? 'border-purple-500 bg-purple-50' : 'border-gray-200 hover:border-gray-300'}"
						>
							<span class="text-xl block mb-1">{getTypeEmoji('internal')}</span>
							<span class="text-xs font-medium {newProject.project_type === 'internal' ? 'text-purple-700' : 'text-gray-600'}">Internal</span>
						</button>
						<button
							type="button"
							onclick={() => newProject.project_type = 'client_work'}
							class="p-3 rounded-xl border-2 transition-all text-center {newProject.project_type === 'client_work' ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
						>
							<span class="text-xl block mb-1">{getTypeEmoji('client_work')}</span>
							<span class="text-xs font-medium {newProject.project_type === 'client_work' ? 'text-blue-700' : 'text-gray-600'}">Client Work</span>
						</button>
						<button
							type="button"
							onclick={() => newProject.project_type = 'learning'}
							class="p-3 rounded-xl border-2 transition-all text-center {newProject.project_type === 'learning' ? 'border-teal-500 bg-teal-50' : 'border-gray-200 hover:border-gray-300'}"
						>
							<span class="text-xl block mb-1">{getTypeEmoji('learning')}</span>
							<span class="text-xs font-medium {newProject.project_type === 'learning' ? 'text-teal-700' : 'text-gray-600'}">Learning</span>
						</button>
					</div>
				</div>

				<!-- Priority Selection with Visual Indicators -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">Priority</label>
					<div class="flex gap-2">
						{#each ['low', 'medium', 'high', 'critical'] as priority}
							<button
								type="button"
								onclick={() => newProject.priority = priority as 'low' | 'medium' | 'high' | 'critical'}
								class="flex-1 py-2 px-3 rounded-lg border-2 transition-all text-center text-sm font-medium {newProject.priority === priority ? 'border-gray-900 bg-gray-900 text-white' : 'border-gray-200 hover:border-gray-300 text-gray-600'}"
							>
								<span class="mr-1">{getPriorityEmoji(priority)}</span>
								<span class="capitalize">{priority}</span>
							</button>
						{/each}
					</div>
				</div>

				<!-- Client Dropdown -->
				<div>
					<label for="client" class="block text-sm font-medium text-gray-700 mb-1">Client (optional)</label>
					<select
						id="client"
						bind:value={newProject.client_name}
						class="input input-square"
					>
						<option value="">No client</option>
						{#each clients as client}
							<option value={client.name}>{client.name}</option>
						{/each}
					</select>
					{#if clients.length === 0}
						<p class="text-xs text-gray-400 mt-1">No clients yet. Create one in the Clients section.</p>
					{/if}
				</div>

				<!-- Description -->
				<div>
					<label for="description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
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
					class="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700"
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
					<div class="space-y-4 pl-4 border-l-2 border-gray-100">
						<p class="text-xs text-gray-400">Additional options like team assignment and tags will be available after creating the project.</p>
					</div>
				{/if}

				{#if createError}
					<div class="text-sm text-red-600 bg-red-50 px-3 py-2 rounded-xl">
						{createError}
					</div>
				{/if}

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => { showNewProject = false; showAdvancedOptions = false; }} class="btn-pill btn-pill-secondary flex-1">
						Cancel
					</button>
					<button type="submit" class="btn-pill btn-pill-primary flex-1">
						Create Project
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
