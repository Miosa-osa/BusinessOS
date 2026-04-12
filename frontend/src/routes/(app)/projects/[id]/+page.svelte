<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api, type Project, type Task, type ClientListResponse, type TeamMemberListResponse } from '$lib/api';
	import { onMount } from 'svelte';
	import { Dialog, Popover } from 'bits-ui';
	import {
		ProjectTimeline,
		ProjectNotes,
		ProjectEditDialog,
		ProjectAddTaskDialog,
		ProjectEditTaskDialog,
		ProjectMembersPanel
	} from '$lib/components/projects';
	import { currentWorkspace } from '$lib/stores/workspaces';
	import { useSession } from '$lib/auth-client';
	import { getPriorityColor, getTypeLabel, formatDate } from '$lib/utils/project';
	import { getBackendUrl } from '$lib/api/base';
	import {
		ChevronRight, MapPin, Zap, Clock, CalendarDays, Paperclip,
		FileText, Archive, Users, Plus, CheckCircle2, PauseCircle,
		PlayCircle, RefreshCw, User, X, ListTodo, XCircle, Pencil,
		Trash2, Upload, Download, File, FileImage, FileArchive
	} from 'lucide-svelte';

	// ── Types ────────────────────────────────────────────────────
	interface ProjectMetadata {
		in_scope?: string[];
		out_of_scope?: string[];
		expected_outcomes?: string[];
		key_features?: { p0?: string[]; p1?: string[]; p2?: string[] };
		quick_links?: { name: string; size: string; url?: string }[];
		estimate?: string;
		location?: string;
		sprint?: string;
	}

	type ActiveTab = 'overview' | 'tasks' | 'notes' | 'files';

	// ── Session & Routing ─────────────────────────────────────────
	const session = useSession();
	const currentUserId = $derived($session.data?.user?.id ?? '');
	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');
	const projectId = $derived($page.params.id);

	// ── State ─────────────────────────────────────────────────────
	let project = $state<Project | null>(null);
	let tasks = $state<Task[]>([]);
	let clients = $state<ClientListResponse[]>([]);
	let teamMembers = $state<TeamMemberListResponse[]>([]);
	let isLoading = $state(true);
	let error = $state('');

	// Dialog state
	let showEditDialog = $state(false);
	let showDeleteConfirm = $state(false);
	let showAddTask = $state(false);
	let showEditTask = $state(false);
	let showLinkClient = $state(false);
	let showAssignTeam = $state(false);
	let editingTask = $state<Task | null>(null);

	// Inline title editing
	let editingTitle = $state(false);
	let titleDraft = $state('');

	// Tab navigation
	let activeTab = $state<ActiveTab>('overview');

	// File upload
	let uploadingFile = $state(false);
	let fileInputEl = $state<HTMLInputElement | null>(null);

	// ── Derived ───────────────────────────────────────────────────
	let completedTasks = $derived(tasks.filter((t) => t.status === 'done').length);
	let totalTasks = $derived(tasks.length);
	let taskProgress = $derived(totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0);

	let meta = $derived<ProjectMetadata>(
		(project?.project_metadata as ProjectMetadata) ?? {}
	);

	let shortId = $derived(project ? `#${project.id.slice(0, 6).toUpperCase()}` : '');

	let isAssignedToMe = $derived(
		teamMembers.some(m => m.id === currentUserId)
	);

	let daysRemaining = $derived(() => {
		if (!project?.due_date) return null;
		const diff = new Date(project.due_date).getTime() - Date.now();
		return Math.max(0, Math.ceil(diff / 86400000));
	});

	let daysProgress = $derived(() => {
		if (!project?.due_date || !project?.created_at) return 0;
		const total = new Date(project.due_date).getTime() - new Date(project.created_at).getTime();
		const elapsed = Date.now() - new Date(project.created_at).getTime();
		if (total <= 0) return 100;
		return Math.min(100, Math.max(0, Math.round((elapsed / total) * 100)));
	});

	// ── Data Loading ──────────────────────────────────────────────
	onMount(async () => {
		await Promise.all([loadProject(), loadTasks(), loadClients(), loadTeamMembers()]);
	});

	async function loadProject() {
		if (!projectId) { error = 'No project ID provided'; return; }
		isLoading = true;
		error = '';
		try {
			project = await api.getProject(projectId);
			if (project && !project.notes) project.notes = [];
		} catch (err) {
			error = 'Failed to load project';
			console.error('Error loading project:', err);
		} finally {
			isLoading = false;
		}
	}

	async function loadTasks() {
		try { tasks = await api.getTasks({ projectId }); }
		catch (err) { console.error('Error loading tasks:', err); }
	}

	async function loadClients() {
		try { clients = await api.getClients(); }
		catch (err) { console.error('Error loading clients:', err); }
	}

	async function loadTeamMembers() {
		try { teamMembers = await api.getTeamMembers(); }
		catch (err) { console.error('Error loading team:', err); }
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

	// ── File Upload ───────────────────────────────────────────────
	async function handleFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file || !project) return;
		uploadingFile = true;
		try {
			const formData = new FormData();
			formData.append('file', file);
			const res = await fetch(`${getBackendUrl()}/projects/${project.id}/files`, {
				method: 'POST',
				credentials: 'include',
				body: formData
			});
			if (!res.ok) throw new Error('Upload failed');
			const { url } = await res.json() as { url: string };
			const existingLinks = meta.quick_links ?? [];
			const sizeMB = (file.size / 1024 / 1024).toFixed(1);
			const updated: ProjectMetadata['quick_links'] = [
				...existingLinks,
				{ name: file.name, size: `${sizeMB} MB`, url }
			];
			await api.updateProject(project.id, {
				project_metadata: { ...(project.project_metadata as object), quick_links: updated }
			});
			await loadProject();
		} catch (err) {
			console.error('File upload error:', err);
		} finally {
			uploadingFile = false;
			if (input) input.value = '';
		}
	}

	async function removeFile(index: number) {
		if (!project) return;
		const existing = meta.quick_links ?? [];
		const updated = existing.filter((_, i) => i !== index);
		await api.updateProject(project.id, {
			project_metadata: { ...(project.project_metadata as object), quick_links: updated }
		});
		await loadProject();
	}

	// ── Utilities ─────────────────────────────────────────────────
	function getFileIcon(name: string) {
		const ext = name.split('.').pop()?.toLowerCase() ?? '';
		if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].includes(ext)) return 'img';
		if (['zip', 'rar', 'tar', 'gz'].includes(ext)) return 'zip';
		return 'file';
	}

	function getRelativeTime(dateStr: string): string {
		const diff = Date.now() - new Date(dateStr).getTime();
		const mins = Math.floor(diff / 60000);
		if (mins < 1) return 'just now';
		if (mins < 60) return `${mins}m ago`;
		const hours = Math.floor(mins / 60);
		if (hours < 24) return `${hours}h ago`;
		const days = Math.floor(hours / 24);
		if (days < 7) return `${days}d ago`;
		return formatDate(dateStr);
	}

	function getPriorityDot(priority: string): string {
		switch (priority) {
			case 'critical': return 'var(--bos-priority-critical)';
			case 'high':     return 'var(--bos-priority-high)';
			case 'medium':   return 'var(--bos-priority-medium)';
			default:         return 'var(--bos-priority-low)';
		}
	}
</script>

<!-- ═══════════════════════════════════════════════ PAGE ROOT -->
<div class="pm-page">

	{#if isLoading}
		<div class="pm-loading">
			<div class="pm-spinner"></div>
		</div>

	{:else if error || !project}
		<div class="pm-error">
			<p class="pm-error__msg">{error || 'Project not found'}</p>
			<a href="/projects{embedSuffix}" class="btn-pill btn-pill-soft btn-pill-sm">Back to Projects</a>
		</div>

	{:else}

		<!-- ── HEADER ────────────────────────────────────────────── -->
		<div class="pm-header">

			<!-- Breadcrumb -->
			<div class="pm-breadcrumb">
				<a href="/projects{embedSuffix}" class="pm-breadcrumb__back" aria-label="Back">
					<svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</a>
				<a href="/projects{embedSuffix}" class="pm-breadcrumb__link">Projects</a>
				<ChevronRight size={12} class="pm-breadcrumb__sep" />
				<span class="pm-breadcrumb__current">{project.name}</span>
			</div>

			<!-- Title row -->
			<div class="pm-title-row">
				<div class="pm-title-left">
					{#if editingTitle}
						<input
							class="pm-title-input"
							bind:value={titleDraft}
							onblur={async () => {
								if (titleDraft.trim() && titleDraft !== project!.name) {
									await api.updateProject(project!.id, { name: titleDraft });
									await loadProject();
								}
								editingTitle = false;
							}}
							onkeydown={(e) => {
								if (e.key === 'Enter') e.currentTarget.blur();
								if (e.key === 'Escape') editingTitle = false;
							}}
							autofocus
						/>
					{:else}
						<h1
							class="pm-title"
							role="button"
							tabindex="0"
							onclick={() => { titleDraft = project!.name; editingTitle = true; }}
							onkeydown={(e) => e.key === 'Enter' && (() => { titleDraft = project!.name; editingTitle = true; })()}
						>
							{project.name}
						</h1>
					{/if}
					<div class="pm-title-badges">
						<span class="pm-badge pm-badge--{project.status}">
							<span class="pm-badge__dot"></span>
							{project.status.charAt(0).toUpperCase() + project.status.slice(1)}
						</span>
						{#if isAssignedToMe}
							<span class="pm-badge pm-badge--assigned">
								<User size={11} />
								Assigned to me
							</span>
						{/if}
					</div>
				</div>
				<div class="pm-title-actions">
					<button onclick={() => showEditDialog = true} class="pm-action-btn" aria-label="Edit project">
						<Pencil size={13} />
						Edit
					</button>
					<button onclick={() => showDeleteConfirm = true} class="pm-action-btn pm-action-btn--danger" aria-label="Delete project">
						<Trash2 size={13} />
					</button>
				</div>
			</div>

			<!-- Meta chips row -->
			<div class="pm-meta-row">
				<span class="pm-chip pm-chip--id">{shortId}</span>

				<span class="pm-chip">
					<span class="pm-chip__dot" style="background: {getPriorityDot(project.priority)}"></span>
					{project.priority.charAt(0).toUpperCase() + project.priority.slice(1)}
				</span>

				{#if meta.location || project.client_name}
					<Popover.Root bind:open={showLinkClient}>
						<Popover.Trigger class="pm-chip pm-chip--btn" aria-label="Link client">
							<MapPin size={11} />
							{project.client_name || meta.location}
						</Popover.Trigger>
						<Popover.Content class="pm-popover" sideOffset={6}>
							<div class="pm-popover__header">Link to Client</div>
							{#if project.client_name}
								<button onclick={() => updateClientLink(null)} class="btn-pill btn-pill-danger btn-pill-sm w-full text-left mb-1">
									Remove link
								</button>
								<div class="pm-popover__divider"></div>
							{/if}
							{#each clients as client}
								<button
									onclick={() => updateClientLink(client.id)}
									class="pm-popover__item {project.client_name === client.name ? 'pm-popover__item--active' : ''}"
								>
									<span class="pm-popover__avatar">{client.name.charAt(0)}</span>
									{client.name}
								</button>
							{/each}
							{#if clients.length === 0}
								<p class="pm-popover__empty">No clients yet</p>
							{/if}
						</Popover.Content>
					</Popover.Root>
				{:else}
					<Popover.Root bind:open={showLinkClient}>
						<Popover.Trigger class="pm-chip pm-chip--btn pm-chip--muted" aria-label="Link client">
							<MapPin size={11} />
							Link client
						</Popover.Trigger>
						<Popover.Content class="pm-popover" sideOffset={6}>
							<div class="pm-popover__header">Link to Client</div>
							{#each clients as client}
								<button onclick={() => updateClientLink(client.id)} class="pm-popover__item">
									<span class="pm-popover__avatar">{client.name.charAt(0)}</span>
									{client.name}
								</button>
							{/each}
							{#if clients.length === 0}
								<p class="pm-popover__empty">No clients yet</p>
							{/if}
						</Popover.Content>
					</Popover.Root>
				{/if}

				{#if meta.sprint}
					<span class="pm-chip">
						<Zap size={11} />
						{meta.sprint}
					</span>
				{/if}

				<span class="pm-chip pm-chip--muted">
					<RefreshCw size={10} />
					{getRelativeTime(project.updated_at)}
				</span>
			</div>

			<!-- Tab bar -->
			<div class="pm-tabs" role="tablist">
				{#each (['overview', 'tasks', 'notes', 'files'] as const) as tab}
					<button
						role="tab"
						aria-selected={activeTab === tab}
						class="pm-tab {activeTab === tab ? 'pm-tab--active' : ''}"
						onclick={() => activeTab = tab}
					>
						{tab.charAt(0).toUpperCase() + tab.slice(1)}
						{#if tab === 'tasks' && totalTasks > 0}
							<span class="pm-tab-count">{totalTasks}</span>
						{/if}
						{#if tab === 'notes' && project.notes && project.notes.length > 0}
							<span class="pm-tab-count">{project.notes.length}</span>
						{/if}
						{#if tab === 'files' && meta.quick_links && meta.quick_links.length > 0}
							<span class="pm-tab-count">{meta.quick_links.length}</span>
						{/if}
					</button>
				{/each}
			</div>

		</div>
		<!-- end pm-header -->

		<!-- ── BODY ─────────────────────────────────────────────── -->
		<div class="pm-body">

			<!-- ── Main content ── -->
			<div class="pm-main" role="tabpanel">

				<!-- ══ OVERVIEW TAB ══ -->
				{#if activeTab === 'overview'}

					<!-- Goals -->
					{#if project.description}
						<div class="pm-section">
							<h2 class="pm-section-heading">Goals</h2>
							<p class="pm-description">{project.description}</p>
						</div>
					{/if}

					<!-- In scope / Out of scope -->
					{#if (meta.in_scope && meta.in_scope.length > 0) || (meta.out_of_scope && meta.out_of_scope.length > 0)}
						<div class="pm-section pm-section--scope">
							<div class="pm-scope-col">
								<h2 class="pm-section-heading pm-section-heading--green">
									<CheckCircle2 size={14} />
									In scope
								</h2>
								{#if meta.in_scope && meta.in_scope.length > 0}
									<ul class="pm-scope-list pm-scope-list--in">
										{#each meta.in_scope as item}<li>{item}</li>{/each}
									</ul>
								{:else}
									<p class="pm-empty-hint">Not defined</p>
								{/if}
							</div>
							<div class="pm-scope-col">
								<h2 class="pm-section-heading pm-section-heading--red">
									<XCircle size={14} />
									Out of scope
								</h2>
								{#if meta.out_of_scope && meta.out_of_scope.length > 0}
									<ul class="pm-scope-list pm-scope-list--out">
										{#each meta.out_of_scope as item}<li>{item}</li>{/each}
									</ul>
								{:else}
									<p class="pm-empty-hint">Not defined</p>
								{/if}
							</div>
						</div>
					{/if}

					<!-- Expected Outcomes -->
					{#if meta.expected_outcomes && meta.expected_outcomes.length > 0}
						<div class="pm-section">
							<h2 class="pm-section-heading">Expected Outcomes</h2>
							<ul class="pm-outcome-list">
								{#each meta.expected_outcomes as outcome}
									<li class="pm-outcome-item">
										<span class="pm-outcome-dot"></span>
										{outcome}
									</li>
								{/each}
							</ul>
						</div>
					{/if}

					<!-- Timeline -->
					{#if tasks.length > 0}
						<div class="pm-section">
							<h2 class="pm-section-heading">Expected Timeline</h2>
							<ProjectTimeline {tasks} {teamMembers} />
						</div>
					{/if}

					<!-- Empty state for overview -->
					{#if !project.description && !meta.in_scope?.length && !meta.expected_outcomes?.length}
						<div class="pm-empty-overview">
							<p class="pm-empty-overview__hint">No overview content yet.</p>
							<button onclick={() => showEditDialog = true} class="btn-pill btn-pill-soft btn-pill-sm">
								<Pencil size={13} /> Fill in project details
							</button>
						</div>
					{/if}

				<!-- ══ TASKS TAB ══ -->
				{:else if activeTab === 'tasks'}
					<div class="pm-section">
						<!-- Progress -->
						{#if totalTasks > 0}
							<div class="pm-progress-bar-wrap">
								<div class="pm-progress-bar-track">
									<div class="pm-progress-bar-fill" style="width:{taskProgress}%"></div>
								</div>
								<span class="pm-progress-label">{completedTasks}/{totalTasks} &middot; {taskProgress}%</span>
							</div>
						{/if}

						<div class="pm-section-header-row">
							<h2 class="pm-section-heading" style="margin-bottom:0">
								<ListTodo size={14} />
								Tasks
								{#if totalTasks > 0}<span class="pm-count-badge">{totalTasks}</span>{/if}
							</h2>
							<button onclick={() => showAddTask = true} class="btn-pill btn-pill-ghost btn-pill-sm" aria-label="Add task">
								<Plus size={13} /> Add Task
							</button>
						</div>

						{#if tasks.length === 0}
							<div class="pm-empty-tasks">
								<p class="pm-empty-hint">No tasks yet</p>
								<button onclick={() => showAddTask = true} class="btn-pill btn-pill-soft btn-pill-sm">
									<Plus size={13} /> Add First Task
								</button>
							</div>
						{:else}
							<div class="pm-task-list">
								{#each tasks as task}
									<div
										class="pm-task-row"
										role="button"
										tabindex="0"
										onclick={() => handleEditTask(task)}
										onkeydown={(e) => e.key === 'Enter' && handleEditTask(task)}
									>
										<span class="pm-task-dot" style="background:{getPriorityDot(task.priority)}"></span>
										<span class="pm-task-title {task.status === 'done' ? 'pm-task-title--done' : ''}">{task.title}</span>
										{#if task.due_date}
											<span class="pm-task-due">{formatDate(task.due_date)}</span>
										{/if}
										<span class="pm-task-status pm-task-status--{task.status.replace('_','-')}">{task.status.replace('_',' ')}</span>
									</div>
								{/each}
							</div>
						{/if}
					</div>

				<!-- ══ NOTES TAB ══ -->
				{:else if activeTab === 'notes'}
					<div class="pm-section">
						<ProjectNotes {project} onProjectUpdate={loadProject} />
					</div>

				<!-- ══ FILES TAB ══ -->
				{:else if activeTab === 'files'}
					<div class="pm-section">
						<div class="pm-section-header-row">
							<h2 class="pm-section-heading" style="margin-bottom:0">
								<Paperclip size={14} />
								Files
								{#if meta.quick_links?.length}<span class="pm-count-badge">{meta.quick_links.length}</span>{/if}
							</h2>
							<button
								onclick={() => fileInputEl?.click()}
								class="btn-pill btn-pill-ghost btn-pill-sm"
								disabled={uploadingFile}
								aria-label="Upload file"
							>
								{#if uploadingFile}
									<div class="pm-spinner-sm"></div>
									Uploading…
								{:else}
									<Upload size={13} /> Upload File
								{/if}
							</button>
							<input
								bind:this={fileInputEl}
								type="file"
								class="sr-only"
								onchange={handleFileSelect}
								aria-label="Select file to upload"
							/>
						</div>

						{#if !meta.quick_links || meta.quick_links.length === 0}
							<div class="pm-empty-tasks">
								<p class="pm-empty-hint">No files uploaded yet</p>
								<button onclick={() => fileInputEl?.click()} class="btn-pill btn-pill-soft btn-pill-sm">
									<Upload size={13} /> Upload First File
								</button>
							</div>
						{:else}
							<div class="pm-file-list">
								{#each meta.quick_links as link, i}
									<div class="pm-file-row">
										<span class="pm-file-icon pm-file-icon--{getFileIcon(link.name)}">
											{#if getFileIcon(link.name) === 'img'}
												<FileImage size={16} />
											{:else if getFileIcon(link.name) === 'zip'}
												<FileArchive size={16} />
											{:else}
												<FileText size={16} />
											{/if}
										</span>
										<div class="pm-file-info">
											<span class="pm-file-name">{link.name}</span>
											<span class="pm-file-size">{link.size}</span>
										</div>
										<div class="pm-file-actions">
											{#if link.url}
												<a href={link.url} target="_blank" rel="noopener noreferrer" class="pm-file-btn" aria-label="Download {link.name}">
													<Download size={14} />
												</a>
											{/if}
											<button
												onclick={() => removeFile(i)}
												class="pm-file-btn pm-file-btn--danger"
												aria-label="Remove {link.name}"
											>
												<X size={14} />
											</button>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/if}

			</div>
			<!-- end pm-main -->

			<!-- ── Right Sidebar ── -->
			<aside class="pm-sidebar">

				<!-- Time section -->
				<div class="pm-sb-card">
					<h3 class="pm-sb-heading">Time</h3>

					<div class="pm-sb-row">
						<span class="pm-sb-label"><Clock size={12} class="pm-sb-icon" /> Estimate</span>
						<span class="pm-sb-value">{meta.estimate ?? '—'}</span>
					</div>

					<div class="pm-sb-row">
						<span class="pm-sb-label"><CalendarDays size={12} class="pm-sb-icon" /> Due Date</span>
						<span class="pm-sb-value">{project.due_date ? formatDate(project.due_date) : '—'}</span>
					</div>

					{#if project.due_date}
						<div class="pm-sb-progress">
							<div class="pm-sb-progress-header">
								<span class="pm-sb-progress-label">
									{#if daysRemaining() !== null && daysRemaining()! > 0}
										{daysRemaining()} days to go
									{:else if daysRemaining() === 0}
										Due today
									{:else}
										Overdue
									{/if}
								</span>
								<span class="pm-sb-progress-pct">{daysProgress()}%</span>
							</div>
							<div class="pm-sb-progress-track">
								<div
									class="pm-sb-progress-fill {daysProgress() >= 85 ? 'pm-sb-progress-fill--danger' : ''}"
									style="width:{daysProgress()}%"
								></div>
							</div>
						</div>
					{/if}
				</div>

				<!-- Details section -->
				<div class="pm-sb-card">
					<h3 class="pm-sb-heading">Details</h3>

					<div class="pm-sb-row">
						<span class="pm-sb-label">Status</span>
						<span class="pm-sb-badge pm-sb-badge--{project.status}">
							{project.status.charAt(0).toUpperCase() + project.status.slice(1)}
						</span>
					</div>

					<div class="pm-sb-row">
						<span class="pm-sb-label">Priority</span>
						<div class="pm-sb-priority">
							<span class="pm-sb-priority-dot" style="background:{getPriorityDot(project.priority)}"></span>
							<span class="pm-sb-value">{project.priority.charAt(0).toUpperCase() + project.priority.slice(1)}</span>
						</div>
					</div>

					<div class="pm-sb-row">
						<span class="pm-sb-label">Type</span>
						<span class="pm-sb-value">{getTypeLabel(project.project_type)}</span>
					</div>

					{#if project.client_name}
						<div class="pm-sb-row">
							<span class="pm-sb-label"><User size={12} class="pm-sb-icon" /> Client</span>
							<span class="pm-sb-value">{project.client_name}</span>
						</div>
					{/if}

					<div class="pm-sb-divider"></div>

					<div class="pm-sb-row pm-sb-row--top">
						<span class="pm-sb-label"><Users size={12} class="pm-sb-icon" /> Team</span>
						{#if teamMembers.length > 0}
							<div class="pm-sb-avatars">
								{#each teamMembers.slice(0, 4) as member}
									<span class="pm-sb-avatar" title={member.name}>
										{member.name.split(' ').map((n: string) => n[0]).join('').slice(0, 2)}
									</span>
								{/each}
								{#if teamMembers.length > 4}
									<span class="pm-sb-avatar pm-sb-avatar--more">+{teamMembers.length - 4}</span>
								{/if}
								<button onclick={() => showAssignTeam = true} class="pm-sb-avatar pm-sb-avatar--add" aria-label="Manage team">
									<Plus size={10} />
								</button>
							</div>
						{:else}
							<button onclick={() => showAssignTeam = true} class="btn-pill btn-pill-ghost btn-pill-sm" aria-label="Assign team">
								<Plus size={11} /> Assign
							</button>
						{/if}
					</div>
				</div>

				<!-- Quick actions -->
				<div class="pm-sb-actions">
					{#if project.status === 'active'}
						<button
							onclick={async () => { await api.updateProject(project!.id, { status: 'completed' }); await loadProject(); }}
							class="pm-sb-primary-btn"
						>
							<CheckCircle2 size={14} /> Mark Complete
						</button>
						<button
							onclick={async () => { await api.updateProject(project!.id, { status: 'paused' }); await loadProject(); }}
							class="pm-sb-secondary-btn"
						>
							<PauseCircle size={13} /> Pause
						</button>
					{:else if project.status === 'paused'}
						<button
							onclick={async () => { await api.updateProject(project!.id, { status: 'active' }); await loadProject(); }}
							class="pm-sb-primary-btn"
						>
							<PlayCircle size={14} /> Resume
						</button>
					{:else if project.status === 'completed'}
						<button
							onclick={async () => { await api.updateProject(project!.id, { status: 'active' }); await loadProject(); }}
							class="pm-sb-primary-btn"
						>
							<PlayCircle size={14} /> Reopen
						</button>
					{/if}
					{#if project.status !== 'archived'}
						<button
							onclick={async () => { await api.updateProject(project!.id, { status: 'archived' }); await loadProject(); }}
							class="pm-sb-ghost-btn"
						>
							<Archive size={13} /> Archive
						</button>
					{/if}
				</div>

				<!-- Timestamps -->
				<div class="pm-sb-timestamps">
					<span>Created {formatDate(project.created_at)}</span>
					<span>Updated {getRelativeTime(project.updated_at)}</span>
				</div>

			</aside>
			<!-- end pm-sidebar -->

		</div>
		<!-- end pm-body -->

	{/if}
</div>
<!-- end pm-page -->

<!-- ═══════════════════════════════════════════════ DIALOGS -->
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

<!-- Team Dialog -->
<Dialog.Root bind:open={showAssignTeam}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 pm-dialog shadow-xl p-6 w-full max-w-3xl max-h-[80vh] overflow-y-auto z-50">
			<div class="flex items-center justify-between mb-4">
				<Dialog.Title class="text-lg font-semibold pm-dialog__title">Manage Team</Dialog.Title>
				<button onclick={() => showAssignTeam = false} class="btn-pill btn-pill-ghost btn-pill-icon" aria-label="Close">
					<X size={20} />
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
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 pm-dialog shadow-xl p-6 w-full max-w-sm z-50">
			<Dialog.Title class="text-lg font-semibold pm-dialog__title mb-2">Delete Project</Dialog.Title>
			<p class="text-sm pm-dialog__body mb-6">
				Are you sure you want to delete "{project?.name}"? This cannot be undone.
			</p>
			<div class="flex gap-3">
				<button onclick={() => showDeleteConfirm = false} class="btn-pill btn-pill-soft btn-pill-sm flex-1">Cancel</button>
				<button onclick={handleDelete} class="btn-pill btn-pill-danger btn-pill-sm flex-1">Delete</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	/* ── Page shell ─────────────────────────────────────────────── */
	.pm-page {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: var(--dbg, #fff);
		overflow: hidden;
	}

	.pm-loading,
	.pm-error {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
	}
	.pm-error__msg {
		font-size: 0.875rem;
		color: var(--dt2, #555);
	}
	.pm-spinner {
		width: 28px;
		height: 28px;
		border: 2px solid var(--dbd, #e0e0e0);
		border-top-color: var(--dt2, #555);
		border-radius: 50%;
		animation: spin 0.7s linear infinite;
	}
	.pm-spinner-sm {
		width: 14px;
		height: 14px;
		border: 1.5px solid var(--dbd, #e0e0e0);
		border-top-color: var(--dt2, #555);
		border-radius: 50%;
		animation: spin 0.7s linear infinite;
		display: inline-block;
	}
	@keyframes spin { to { transform: rotate(360deg); } }

	/* ── Header ─────────────────────────────────────────────────── */
	.pm-header {
		flex-shrink: 0;
		background: var(--dbg, #fff);
		border-bottom: 1px solid var(--dbd2, #f0f0f0);
		padding: 14px 24px 0;
	}

	/* Breadcrumb */
	.pm-breadcrumb {
		display: flex;
		align-items: center;
		gap: 6px;
		margin-bottom: 12px;
	}
	.pm-breadcrumb__back {
		display: flex;
		align-items: center;
		color: var(--dt3, #888);
		transition: color 150ms;
	}
	.pm-breadcrumb__back:hover { color: var(--dt, #111); }
	.pm-breadcrumb__link {
		font-size: 0.75rem;
		color: var(--dt3, #888);
		text-decoration: none;
		transition: color 150ms;
	}
	.pm-breadcrumb__link:hover { color: var(--dt, #111); }
	.pm-breadcrumb__sep { color: var(--dt4, #bbb); }
	.pm-breadcrumb__current {
		font-size: 0.75rem;
		color: var(--dt2, #555);
		font-weight: 500;
		max-width: 260px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* Title row */
	.pm-title-row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 10px;
	}
	.pm-title-left {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
		min-width: 0;
	}
	.pm-title {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--dt, #111);
		cursor: text;
		margin: 0;
		line-height: 1.3;
	}
	.pm-title:hover { opacity: 0.8; }
	.pm-title-input {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--dt, #111);
		background: var(--dbg2, #f5f5f5);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 6px;
		padding: 2px 8px;
		outline: none;
		width: 400px;
		max-width: 100%;
	}
	.pm-title-badges {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-shrink: 0;
	}

	/* Badges */
	.pm-badge {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-size: 0.7rem;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 20px;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt2, #555);
		border: 1px solid var(--dbd2, #f0f0f0);
	}
	.pm-badge__dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--dt3, #888);
	}
	.pm-badge--active   .pm-badge__dot { background: #22c55e; }
	.pm-badge--paused   .pm-badge__dot { background: #f59e0b; }
	.pm-badge--completed .pm-badge__dot { background: #3b82f6; }
	.pm-badge--archived .pm-badge__dot { background: var(--dt4, #bbb); }
	.pm-badge--assigned {
		background: color-mix(in srgb, #f97316 12%, transparent);
		color: #ea580c;
		border-color: color-mix(in srgb, #f97316 20%, transparent);
	}

	/* Action buttons */
	.pm-title-actions {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-shrink: 0;
	}
	.pm-action-btn {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-size: 0.75rem;
		font-weight: 500;
		padding: 5px 10px;
		border-radius: 8px;
		border: 1px solid var(--dbd, #e0e0e0);
		background: var(--dbg, #fff);
		color: var(--dt2, #555);
		cursor: pointer;
		transition: all 150ms;
	}
	.pm-action-btn:hover { background: var(--dbg2, #f5f5f5); color: var(--dt, #111); }
	.pm-action-btn--danger { color: #ef4444; }
	.pm-action-btn--danger:hover { background: color-mix(in srgb, #ef4444 8%, transparent); border-color: color-mix(in srgb, #ef4444 30%, transparent); }

	/* Meta chips row */
	.pm-meta-row {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
		margin-bottom: 14px;
	}
	.pm-chip {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-size: 0.7rem;
		font-weight: 500;
		color: var(--dt2, #555);
		background: var(--dbg2, #f5f5f5);
		border: 1px solid var(--dbd2, #f0f0f0);
		border-radius: 6px;
		padding: 3px 8px;
		white-space: nowrap;
	}
	.pm-chip--id {
		font-family: monospace;
		font-size: 0.68rem;
		color: var(--dt3, #888);
	}
	.pm-chip__dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.pm-chip--btn {
		cursor: pointer;
		background: none;
		border: none;
		transition: background 150ms;
	}
	.pm-chip--btn:hover { background: var(--dbg2, #f5f5f5); }
	.pm-chip--muted { color: var(--dt3, #888); }

	/* Tab bar */
	.pm-tabs {
		display: flex;
		gap: 0;
	}
	.pm-tab {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--dt3, #888);
		padding: 8px 14px;
		border: none;
		background: none;
		cursor: pointer;
		border-bottom: 2px solid transparent;
		transition: color 150ms, border-color 150ms;
		white-space: nowrap;
	}
	.pm-tab:hover { color: var(--dt, #111); }
	.pm-tab--active {
		color: var(--dt, #111);
		border-bottom-color: var(--dt, #111);
		font-weight: 600;
	}
	.pm-tab-count {
		font-size: 0.65rem;
		font-weight: 600;
		background: var(--dbg3, #eee);
		color: var(--dt3, #888);
		border-radius: 10px;
		padding: 1px 6px;
		line-height: 1.4;
	}
	.pm-tab--active .pm-tab-count {
		background: var(--dt, #111);
		color: var(--dbg, #fff);
	}

	/* ── Body ───────────────────────────────────────────────────── */
	.pm-body {
		flex: 1;
		display: flex;
		overflow: hidden;
	}

	/* Main content */
	.pm-main {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
		scrollbar-width: none;
	}
	.pm-main::-webkit-scrollbar { display: none; }

	/* ── Sections ───────────────────────────────────────────────── */
	.pm-section {
		margin-bottom: 28px;
	}
	.pm-section-heading {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 0.8rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--dt3, #888);
		margin: 0 0 12px;
	}
	.pm-section-heading--green { color: #16a34a; }
	.pm-section-heading--red   { color: #dc2626; }

	.pm-section-header-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 12px;
	}

	.pm-description {
		font-size: 0.875rem;
		line-height: 1.7;
		color: var(--dt, #111);
		margin: 0;
	}

	/* In scope / Out of scope */
	.pm-section--scope {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 20px;
	}
	.pm-scope-col {}
	.pm-scope-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.pm-scope-list li {
		font-size: 0.825rem;
		color: var(--dt2, #555);
		padding-left: 14px;
		position: relative;
		line-height: 1.5;
	}
	.pm-scope-list li::before {
		content: '–';
		position: absolute;
		left: 0;
		color: var(--dt4, #bbb);
	}

	/* Outcomes */
	.pm-outcome-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.pm-outcome-item {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		font-size: 0.825rem;
		color: var(--dt2, #555);
		line-height: 1.5;
	}
	.pm-outcome-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--dt4, #bbb);
		flex-shrink: 0;
		margin-top: 6px;
	}

	/* Progress bar (tasks tab) */
	.pm-progress-bar-wrap {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 16px;
	}
	.pm-progress-bar-track {
		flex: 1;
		height: 4px;
		background: var(--dbd2, #f0f0f0);
		border-radius: 2px;
		overflow: hidden;
	}
	.pm-progress-bar-fill {
		height: 100%;
		background: #22c55e;
		border-radius: 2px;
		transition: width 400ms ease;
	}
	.pm-progress-label {
		font-size: 0.7rem;
		color: var(--dt3, #888);
		white-space: nowrap;
	}

	/* Task list */
	.pm-task-list {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.pm-task-row {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 10px;
		border-radius: 8px;
		cursor: pointer;
		transition: background 150ms;
	}
	.pm-task-row:hover { background: var(--dbg2, #f5f5f5); }
	.pm-task-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.pm-task-title {
		flex: 1;
		font-size: 0.825rem;
		color: var(--dt, #111);
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.pm-task-title--done {
		text-decoration: line-through;
		color: var(--dt3, #888);
	}
	.pm-task-due {
		font-size: 0.7rem;
		color: var(--dt3, #888);
		flex-shrink: 0;
	}
	.pm-task-status {
		font-size: 0.65rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		padding: 2px 7px;
		border-radius: 10px;
		flex-shrink: 0;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt3, #888);
	}
	.pm-task-status--done        { background: color-mix(in srgb, #22c55e 12%, transparent); color: #16a34a; }
	.pm-task-status--in-progress { background: color-mix(in srgb, #3b82f6 12%, transparent); color: #2563eb; }
	.pm-task-status--blocked     { background: color-mix(in srgb, #ef4444 12%, transparent); color: #dc2626; }

	/* Count badge */
	.pm-count-badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-size: 0.65rem;
		font-weight: 600;
		background: var(--dbg3, #eee);
		color: var(--dt2, #555);
		border-radius: 10px;
		padding: 1px 6px;
		line-height: 1.4;
	}

	/* Empty states */
	.pm-empty-tasks {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		padding: 40px 20px;
		border: 1px dashed var(--dbd, #e0e0e0);
		border-radius: 12px;
	}
	.pm-empty-hint {
		font-size: 0.825rem;
		color: var(--dt3, #888);
		margin: 0;
	}
	.pm-empty-overview {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 60px 20px;
	}
	.pm-empty-overview__hint {
		font-size: 0.875rem;
		color: var(--dt3, #888);
		margin: 0;
	}

	/* File list */
	.pm-file-list {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.pm-file-row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
		border-radius: 10px;
		border: 1px solid var(--dbd2, #f0f0f0);
		transition: border-color 150ms, background 150ms;
	}
	.pm-file-row:hover { background: var(--dbg2, #f5f5f5); border-color: var(--dbd, #e0e0e0); }
	.pm-file-icon {
		width: 34px;
		height: 34px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt2, #555);
		flex-shrink: 0;
	}
	.pm-file-icon--img { background: color-mix(in srgb, #8b5cf6 12%, transparent); color: #7c3aed; }
	.pm-file-icon--zip { background: color-mix(in srgb, #f59e0b 12%, transparent); color: #d97706; }
	.pm-file-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.pm-file-name {
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--dt, #111);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.pm-file-size {
		font-size: 0.7rem;
		color: var(--dt3, #888);
	}
	.pm-file-actions {
		display: flex;
		align-items: center;
		gap: 4px;
		opacity: 0;
		transition: opacity 150ms;
	}
	.pm-file-row:hover .pm-file-actions { opacity: 1; }
	.pm-file-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: 6px;
		border: none;
		background: none;
		color: var(--dt3, #888);
		cursor: pointer;
		transition: background 150ms, color 150ms;
		text-decoration: none;
	}
	.pm-file-btn:hover { background: var(--dbg3, #eee); color: var(--dt, #111); }
	.pm-file-btn--danger:hover { background: color-mix(in srgb, #ef4444 10%, transparent); color: #ef4444; }

	/* ── Right Sidebar ──────────────────────────────────────────── */
	.pm-sidebar {
		width: 272px;
		flex-shrink: 0;
		border-left: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg, #fff);
		overflow-y: auto;
		padding: 20px 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		scrollbar-width: none;
	}
	.pm-sidebar::-webkit-scrollbar { display: none; }

	.pm-sb-card {
		background: var(--dbg2, #f5f5f5);
		border: 1px solid var(--dbd2, #f0f0f0);
		border-radius: 12px;
		padding: 14px;
	}
	.pm-sb-heading {
		font-size: 0.65rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--dt4, #bbb);
		margin: 0 0 12px;
	}
	.pm-sb-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		padding: 5px 0;
	}
	.pm-sb-row--top { align-items: flex-start; }
	.pm-sb-label {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 0.72rem;
		color: var(--dt3, #888);
		flex-shrink: 0;
	}
	.pm-sb-icon { opacity: 0.7; }
	.pm-sb-value {
		font-size: 0.75rem;
		color: var(--dt, #111);
		font-weight: 500;
		text-align: right;
	}

	/* Status badge */
	.pm-sb-badge {
		font-size: 0.68rem;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 20px;
		background: var(--dbg3, #eee);
		color: var(--dt2, #555);
	}
	.pm-sb-badge--active    { background: color-mix(in srgb, #22c55e 15%, transparent); color: #16a34a; }
	.pm-sb-badge--paused    { background: color-mix(in srgb, #f59e0b 15%, transparent); color: #d97706; }
	.pm-sb-badge--completed { background: color-mix(in srgb, #3b82f6 15%, transparent); color: #2563eb; }
	.pm-sb-badge--archived  { background: var(--dbg3, #eee); color: var(--dt3, #888); }

	/* Priority */
	.pm-sb-priority {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.pm-sb-priority-dot {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	/* Avatars */
	.pm-sb-avatars {
		display: flex;
		align-items: center;
	}
	.pm-sb-avatar {
		width: 26px;
		height: 26px;
		border-radius: 50%;
		background: var(--dbg3, #eee);
		color: var(--dt2, #555);
		font-size: 0.6rem;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
		border: 2px solid var(--dbg, #fff);
		margin-left: -6px;
	}
	.pm-sb-avatars :first-child { margin-left: 0; }
	.pm-sb-avatar--more { background: var(--dbg3, #eee); color: var(--dt3, #888); font-size: 0.55rem; }
	.pm-sb-avatar--add {
		background: none;
		border: 2px dashed var(--dbd, #e0e0e0);
		color: var(--dt3, #888);
		cursor: pointer;
		transition: border-color 150ms, color 150ms;
	}
	.pm-sb-avatar--add:hover { border-color: var(--dt2, #555); color: var(--dt, #111); }

	/* Divider */
	.pm-sb-divider {
		height: 1px;
		background: var(--dbd2, #f0f0f0);
		margin: 6px 0;
	}

	/* Days remaining progress */
	.pm-sb-progress {
		margin-top: 10px;
	}
	.pm-sb-progress-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 6px;
	}
	.pm-sb-progress-label {
		font-size: 0.7rem;
		color: var(--dt2, #555);
		font-weight: 500;
	}
	.pm-sb-progress-pct {
		font-size: 0.68rem;
		color: var(--dt3, #888);
	}
	.pm-sb-progress-track {
		height: 5px;
		background: var(--dbd2, #f0f0f0);
		border-radius: 3px;
		overflow: hidden;
	}
	.pm-sb-progress-fill {
		height: 100%;
		background: #3b82f6;
		border-radius: 3px;
		transition: width 400ms ease;
	}
	.pm-sb-progress-fill--danger { background: #ef4444; }

	/* Quick actions */
	.pm-sb-actions {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.pm-sb-primary-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		width: 100%;
		padding: 8px 14px;
		font-size: 0.78rem;
		font-weight: 600;
		background: var(--dt, #111);
		color: var(--dbg, #fff);
		border: none;
		border-radius: 10px;
		cursor: pointer;
		transition: opacity 150ms;
	}
	.pm-sb-primary-btn:hover { opacity: 0.85; }
	.pm-sb-secondary-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		width: 100%;
		padding: 7px 14px;
		font-size: 0.78rem;
		font-weight: 500;
		background: var(--dbg2, #f5f5f5);
		color: var(--dt, #111);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		cursor: pointer;
		transition: background 150ms;
	}
	.pm-sb-secondary-btn:hover { background: var(--dbg3, #eee); }
	.pm-sb-ghost-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		width: 100%;
		padding: 7px 14px;
		font-size: 0.75rem;
		font-weight: 500;
		background: none;
		color: var(--dt3, #888);
		border: none;
		border-radius: 10px;
		cursor: pointer;
		transition: background 150ms, color 150ms;
	}
	.pm-sb-ghost-btn:hover { background: var(--dbg2, #f5f5f5); color: var(--dt, #111); }

	/* Timestamps */
	.pm-sb-timestamps {
		display: flex;
		flex-direction: column;
		gap: 3px;
		padding: 0 2px;
	}
	.pm-sb-timestamps span {
		font-size: 0.65rem;
		color: var(--dt4, #bbb);
	}

	/* Popover */
	:global(.pm-popover) {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		padding: 6px;
		min-width: 180px;
		box-shadow: 0 8px 24px rgba(0,0,0,0.08);
		z-index: 100;
	}
	:global(.pm-popover__header) {
		font-size: 0.65rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--dt3, #888);
		padding: 4px 8px 8px;
	}
	:global(.pm-popover__divider) {
		height: 1px;
		background: var(--dbd2, #f0f0f0);
		margin: 4px 0;
	}
	:global(.pm-popover__item) {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 7px 8px;
		font-size: 0.8rem;
		color: var(--dt, #111);
		background: none;
		border: none;
		border-radius: 7px;
		cursor: pointer;
		transition: background 150ms;
		text-align: left;
	}
	:global(.pm-popover__item:hover) { background: var(--dbg2, #f5f5f5); }
	:global(.pm-popover__item--active) { background: var(--dbg2, #f5f5f5); font-weight: 600; }
	:global(.pm-popover__avatar) {
		width: 22px;
		height: 22px;
		border-radius: 50%;
		background: var(--dbg3, #eee);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-size: 0.65rem;
		font-weight: 700;
		color: var(--dt2, #555);
	}
	:global(.pm-popover__empty) {
		font-size: 0.78rem;
		color: var(--dt3, #888);
		padding: 8px;
		margin: 0;
	}

	/* Dialog */
	:global(.pm-dialog) {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 16px;
	}
	:global(.pm-dialog__title) { color: var(--dt, #111); }
	:global(.pm-dialog__body)  { color: var(--dt2, #555); }

	/* Screen reader only */
	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0,0,0,0);
		white-space: nowrap;
		border: 0;
	}
</style>
