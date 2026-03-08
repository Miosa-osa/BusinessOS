<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api, type Project, type Task, type ContextListItem, type ClientListResponse, type TeamMemberListResponse } from '$lib/api';
	import { onMount } from 'svelte';
	import { Dialog, Popover } from 'bits-ui';
	import {
		ProjectOverview,
		ProjectTasks,
		ProjectTimeline,
		ProjectSprints,
		ProjectDocuments,
		ProjectNotes,
		ProjectEditDialog,
		ProjectAddTaskDialog,
		ProjectEditTaskDialog,
		ProjectMembersPanel
	} from '$lib/components/projects';
	import { currentWorkspace } from '$lib/stores/workspaces';
	import { useSession } from '$lib/auth-client';
	import { getStatusColor, getStatusIcon, getPriorityColor, getTypeIcon, getTypeLabel, formatDate } from '$lib/utils/project';

	const session = useSession();
	const currentUserId = $derived($session.data?.user?.id ?? '');

	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	let project = $state<Project | null>(null);
	let tasks = $state<Task[]>([]);
	let availableDocuments = $state<ContextListItem[]>([]);
	let clients = $state<ClientListResponse[]>([]);
	let teamMembers = $state<TeamMemberListResponse[]>([]);
	let isLoading = $state(true);
	let error = $state('');
	let showEditDialog = $state(false);
	let showDeleteConfirm = $state(false);
	let showAddTask = $state(false);
	let showEditTask = $state(false);
	let showLinkClient = $state(false);
	let showAssignTeam = $state(false);
	let loadingAvailable = $state(false);
	let activeTab = $state<'overview' | 'tasks' | 'timeline' | 'sprints' | 'documents' | 'notes'>('overview');
	let editingTask = $state<Task | null>(null);

	const projectId = $derived($page.params.id);

	// Derived task counts for header progress bar
	let completedTasks = $derived(tasks.filter((t) => t.status === 'done').length);
	let totalTasks = $derived(tasks.length);
	let taskProgress = $derived(totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0);

	onMount(async () => {
		await Promise.all([loadProject(), loadTasks(), loadClients(), loadTeamMembers()]);
	});

	async function loadProject() {
		if (!projectId) {
			error = 'No project ID provided';
			return;
		}
		isLoading = true;
		error = '';
		try {
			project = await api.getProject(projectId);
			if (project && !project.notes) {
				project.notes = [];
			}
		} catch (err) {
			error = 'Failed to load project';
			console.error('Error loading project:', err);
		} finally {
			isLoading = false;
		}
	}

	async function loadTasks() {
		try {
			tasks = await api.getTasks({ projectId });
		} catch (err) {
			console.error('Error loading tasks:', err);
		}
	}

	async function loadClients() {
		try {
			clients = await api.getClients();
		} catch (err) {
			console.error('Error loading clients:', err);
		}
	}

	async function loadTeamMembers() {
		try {
			teamMembers = await api.getTeamMembers();
		} catch (err) {
			console.error('Error loading team:', err);
		}
	}

	async function loadAvailableDocuments() {
		loadingAvailable = true;
		try {
			const allContexts = await api.getContexts();
			availableDocuments = allContexts.filter((c) => c.type === 'document');
		} catch (err) {
			console.error('Error loading documents:', err);
		} finally {
			loadingAvailable = false;
		}
	}

	async function handleDelete() {
		if (!project) return;
		try {
			await api.deleteProject(project.id);
			goto('/projects' + embedSuffix);
		} catch (err) {
			console.error('Error deleting project:', err);
		}
	}

	async function updateClientLink(clientId: string | null) {
		if (!project) return;
		try {
			const selectedClient = clientId ? clients.find((c) => c.id === clientId) : null;
			await api.updateProject(project.id, { client_name: selectedClient?.name || '' });
			await loadProject();
			showLinkClient = false;
		} catch (err) {
			console.error('Error updating client:', err);
		}
	}

	function handleEditTask(task: Task) {
		editingTask = task;
		showEditTask = true;
	}

	function handleCloseEditTask() {
		showEditTask = false;
		editingTask = null;
	}
</script>

<div class="h-full flex flex-col prm-det-page">
	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 prm-det-spinner rounded-full"></div>
		</div>
	{:else if error || !project}
		<div class="flex-1 flex items-center justify-center">
			<div class="text-center">
				<p class="prm-det-muted mb-4">{error || 'Project not found'}</p>
				<a href="/projects{embedSuffix}" class="btn-pill btn-pill-soft btn-pill-sm">Back to Projects</a>
			</div>
		</div>
	{:else}
		<!-- Header — Compact Info Bar -->
		<div class="prm-det-header">
			<!-- Row 1: Breadcrumb -->
			<div class="px-6 pt-3 pb-0">
				<div class="flex items-center gap-1.5 text-xs prm-det-muted">
					<a href="/projects{embedSuffix}" class="prm-det-breadcrumb flex items-center gap-1">
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
						Projects
					</a>
					<span class="prm-det-sep">/</span>
					<span>{getTypeLabel(project.project_type)}</span>
					<span class="prm-det-sep">/</span>
					<span class="prm-det-label truncate max-w-[200px]">{project.name}</span>
				</div>
			</div>

			<!-- Row 2: Icon + Title + Badges + Actions -->
			<div class="px-6 pt-2 pb-1 flex items-center justify-between gap-4">
				<div class="flex items-center gap-3 min-w-0">
					<div class="w-8 h-8 rounded-lg prm-det-icon-gradient flex items-center justify-center text-white text-sm flex-shrink-0">
						{getTypeIcon(project.project_type)}
					</div>
					<h1 class="text-lg font-semibold prm-det-title truncate">{project.name}</h1>
					<span class="text-xs font-medium px-2 py-0.5 rounded-full border {getStatusColor(project.status)} flex items-center gap-1 flex-shrink-0">
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getStatusIcon(project.status)} />
						</svg>
						{project.status}
					</span>
					<span class="text-xs font-medium px-1.5 py-0.5 rounded {getPriorityColor(project.priority)} flex-shrink-0">
						{project.priority}
					</span>
				</div>
				<div class="flex gap-2 flex-shrink-0">
					<button onclick={() => showEditDialog = true} class="btn-pill btn-pill-secondary btn-pill-sm flex items-center gap-1">
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
						</svg>
						Edit
					</button>
					<button onclick={() => showDeleteConfirm = true} class="btn-pill btn-pill-danger btn-pill-sm flex items-center gap-1">
						<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
						</svg>
						Delete
					</button>
				</div>
			</div>

			<!-- Row 3: Dense Metadata Strip + Mini Progress -->
			<div class="px-6 pb-2 flex items-center gap-4 text-xs prm-det-muted flex-wrap">
				<!-- Client -->
				<Popover.Root bind:open={showLinkClient}>
					<Popover.Trigger class="flex items-center gap-1 prm-det-breadcrumb cursor-pointer">
						{#if project.client_name}
							<svg class="w-3.5 h-3.5 prm-det-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
							<span class="prm-det-accent">{project.client_name}</span>
						{:else}
							<svg class="w-3.5 h-3.5 prm-det-hint" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
							</svg>
							<span class="prm-det-hint">Link client</span>
						{/if}
					</Popover.Trigger>
					<Popover.Content class="z-50 prm-det-popover p-2 w-64 max-h-64 overflow-y-auto">
						<div class="text-xs font-medium prm-det-hint uppercase px-2 py-1 mb-1">Link to Client</div>
						{#if project.client_name}
							<button
								onclick={() => updateClientLink(null)}
								class="w-full text-left btn-pill btn-pill-danger btn-pill-sm"
							>
								Remove link
							</button>
							<div class="prm-det-divider my-1"></div>
						{/if}
						{#each clients as client}
							<button
								onclick={() => updateClientLink(client.id)}
								class="w-full text-left btn-pill btn-pill-ghost btn-pill-sm flex items-center gap-2 {project.client_name === client.name ? 'btn-pill-soft' : ''}"
							>
								<span class="w-6 h-6 rounded-full prm-det-avatar flex items-center justify-center text-xs">
									{client.name.charAt(0)}
								</span>
								{client.name}
							</button>
						{/each}
						{#if clients.length === 0}
							<p class="text-xs prm-det-hint text-center py-4">No clients yet</p>
						{/if}
					</Popover.Content>
				</Popover.Root>

				<span class="prm-det-sep">·</span>
				<span>Created {formatDate(project.created_at)}</span>

				{#if project.due_date}
					<span class="prm-det-sep">·</span>
					<span>Due {formatDate(project.due_date)}</span>
				{/if}

				<!-- Mini Progress -->
				{#if totalTasks > 0}
					<span class="prm-det-sep">·</span>
					<div class="flex items-center gap-2">
						<div class="w-20 h-1.5 prm-det-progress-bg rounded-full overflow-hidden">
							<div
								class="h-full prm-det-progress-fill rounded-full transition-all duration-300"
								style="width: {taskProgress}%"
							></div>
						</div>
						<span class="font-medium prm-det-label">{completedTasks}/{totalTasks} · {taskProgress}%</span>
					</div>
				{/if}
			</div>

			<!-- Tabs -->
			<div class="px-6 flex gap-1 prm-det-divider-top">
				<button
					onclick={() => activeTab = 'overview'}
					class="prm-det-tab {activeTab === 'overview' ? 'prm-det-tab--active' : ''}"
				>
					Overview
				</button>
				<button
					onclick={() => activeTab = 'tasks'}
					class="prm-det-tab {activeTab === 'tasks' ? 'prm-det-tab--active' : ''}"
				>
					Tasks
					{#if totalTasks > 0}
						<span class="prm-det-tab-count">{totalTasks}</span>
					{/if}
				</button>
				<button
					onclick={() => activeTab = 'timeline'}
					class="prm-det-tab {activeTab === 'timeline' ? 'prm-det-tab--active' : ''}"
				>
					Timeline
				</button>
				<button
					onclick={() => activeTab = 'sprints'}
					class="prm-det-tab {activeTab === 'sprints' ? 'prm-det-tab--active' : ''}"
				>
					Sprints
				</button>
				<button
					onclick={() => { activeTab = 'documents'; loadAvailableDocuments(); }}
					class="prm-det-tab {activeTab === 'documents' ? 'prm-det-tab--active' : ''}"
				>
					Documents
					{#if availableDocuments.length > 0}
						<span class="prm-det-tab-count">{availableDocuments.length}</span>
					{/if}
				</button>
				<button
					onclick={() => activeTab = 'notes'}
					class="prm-det-tab {activeTab === 'notes' ? 'prm-det-tab--active' : ''}"
				>
					Notes
					{#if project.notes && project.notes.length > 0}
						<span class="prm-det-tab-count">{project.notes.length}</span>
					{/if}
				</button>
			</div>
		</div>

		<!-- Tab Content -->
		<div class="flex-1 overflow-y-auto p-6">
			<div class="max-w-5xl mx-auto">
				{#if activeTab === 'overview'}
					<ProjectOverview
						{project}
						{tasks}
						{teamMembers}
						{clients}
						{embedSuffix}
						onProjectUpdate={loadProject}
						onNavigateToTasks={() => activeTab = 'tasks'}
						onShowAddTask={() => showAddTask = true}
						onShowAssignTeam={() => showAssignTeam = true}
					/>
				{:else if activeTab === 'tasks'}
					<ProjectTasks
						{tasks}
						{teamMembers}
						{embedSuffix}
						onShowAddTask={() => showAddTask = true}
						onEditTask={handleEditTask}
						onTasksChanged={loadTasks}
					/>
				{:else if activeTab === 'timeline'}
					<ProjectTimeline
						{tasks}
						{teamMembers}
					/>
				{:else if activeTab === 'sprints'}
					<ProjectSprints
						{tasks}
					/>
				{:else if activeTab === 'documents'}
					<ProjectDocuments
						{availableDocuments}
						{loadingAvailable}
						{embedSuffix}
					/>
				{:else if activeTab === 'notes'}
					<ProjectNotes {project} onProjectUpdate={loadProject} />
				{/if}
			</div>
		</div>
	{/if}
</div>

<!-- Dialogs -->
{#if project}
	<ProjectEditDialog
		bind:open={showEditDialog}
		{project}
		{clients}
		onClose={() => showEditDialog = false}
		onProjectUpdate={loadProject}
	/>

	<ProjectAddTaskDialog
		bind:open={showAddTask}
		projectId={project.id}
		{tasks}
		{teamMembers}
		onClose={() => showAddTask = false}
		onTaskCreated={loadTasks}
	/>
{/if}

<ProjectEditTaskDialog
	bind:open={showEditTask}
	bind:task={editingTask}
	onClose={handleCloseEditTask}
	onTaskUpdated={loadTasks}
/>

<!-- Assign Team Members -->
<Dialog.Root bind:open={showAssignTeam}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 prm-det-dialog shadow-xl p-6 w-full max-w-3xl max-h-[80vh] overflow-y-auto z-50">
			<div class="flex items-center justify-between mb-4">
				<Dialog.Title class="text-lg font-semibold prm-det-title">Manage Team</Dialog.Title>
				<button onclick={() => showAssignTeam = false} class="btn-pill btn-pill-ghost btn-pill-icon">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
			{#if project}
				<ProjectMembersPanel
					projectId={project.id}
					workspaceId={$currentWorkspace?.id ?? ''}
					{currentUserId}
					userRole="lead"
					canInvite={true}
				/>
			{/if}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<!-- Delete Confirmation -->
<Dialog.Root bind:open={showDeleteConfirm}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 prm-det-dialog shadow-xl p-6 w-full max-w-sm z-50">
			<Dialog.Title class="text-lg font-semibold prm-det-title mb-2">Delete Project</Dialog.Title>
			<p class="text-sm prm-det-muted mb-6">
				Are you sure you want to delete "{project?.name}"? This action cannot be undone.
			</p>
			<div class="flex gap-3">
				<button onclick={() => showDeleteConfirm = false} class="btn-pill btn-pill-soft btn-pill-sm flex-1">Cancel</button>
				<button onclick={handleDelete} class="btn-pill btn-pill-danger btn-pill-sm flex-1">Delete</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	.prm-det-page { background: var(--dbg2, #f9fafb); }
	.prm-det-header { background: var(--dbg, #fff); border-bottom: 1px solid var(--dbd, #e5e7eb); }
	.prm-det-title { color: var(--dt, #111); }
	.prm-det-muted { color: var(--dt3, #6b7280); }
	.prm-det-hint { color: var(--dt4, #9ca3af); }
	.prm-det-label { color: var(--dt2, #4b5563); }
	.prm-det-sep { color: var(--dt4, #d1d5db); }
	.prm-det-breadcrumb { color: inherit; }
	.prm-det-breadcrumb:hover { color: var(--dt2, #374151); }
	.prm-det-spinner { border-color: var(--dt, #111); border-top-color: transparent; }
	.prm-det-popover { background: var(--dbg, #fff); border-radius: 0.75rem; box-shadow: 0 10px 15px rgba(0,0,0,.1); border: 1px solid var(--dbd, #e5e7eb); }
	.prm-det-divider { border-top: 1px solid var(--dbd2, #f3f4f6); }
	.prm-det-divider-top { border-top: 1px solid var(--dbd2, #f3f4f6); }
	.prm-det-avatar { background: var(--dbg3, #f3f4f6); }
	.prm-det-progress-bg { background: var(--dbg3, #e5e7eb); }
	.prm-det-badge { background: var(--dbg3, #f3f4f6); color: var(--dt2, #4b5563); }
	.prm-det-badge--active { background: var(--dt, #111); color: #fff; }
	.prm-det-icon-gradient { background: linear-gradient(135deg, #9333ea, #4f46e5); }
	.prm-det-accent { color: #3b82f6; }
	.prm-det-progress-fill { background: linear-gradient(to right, #9333ea, #4f46e5); }
	.prm-det-tab {
		position: relative;
		padding: 0.625rem 0.75rem;
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--dt3, #6b7280);
		background: none;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.375rem;
		transition: color 0.15s;
	}
	.prm-det-tab:hover { color: var(--dt, #111); }
	.prm-det-tab--active {
		color: var(--dt, #111);
		font-weight: 600;
	}
	.prm-det-tab--active::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 0.5rem;
		right: 0.5rem;
		height: 2px;
		background: #9333ea;
		border-radius: 1px;
	}
	.prm-det-tab-count {
		font-size: 0.6875rem;
		font-weight: 600;
		padding: 0.0625rem 0.375rem;
		border-radius: 9999px;
		background: var(--dbg3, #f3f4f6);
		color: var(--dt2, #4b5563);
	}
	.prm-det-dialog { background: var(--dbg, #fff); border-radius: 1rem; }
</style>
