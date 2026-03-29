<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { DropdownMenu } from 'bits-ui';
	import {
		ArrowLeft,
		Pencil,
		Share2,
		Bot,
		ChevronRight,
		Check,
		ChevronDown,
		Plus,
		CircleCheck,
		Loader,
		AlertCircle,
		Clock,
		User,
		Calendar,
		Link
	} from 'lucide-svelte';
	import { api } from '$lib/api';
	import type { UpdateTaskData, TaskStatus as APITaskStatus } from '$lib/api';
	import { team } from '$lib/stores/team';

	type UIStatus = 'todo' | 'in_progress' | 'in_review' | 'done' | 'blocked';
	type Priority = 'critical' | 'high' | 'medium' | 'low';

	interface Assignee {
		id: string;
		name: string;
		avatar?: string;
	}

	interface SubtaskLocal {
		id: string;
		title: string;
		completed: boolean;
	}

	interface ActivityEntry {
		id: string;
		type: string;
		description: string;
		createdAt: string;
	}

	interface TaskDetail {
		id: string;
		title: string;
		description?: string;
		status: UIStatus;
		priority: Priority;
		projectId?: string;
		projectName?: string;
		projectColor?: string;
		assignee?: Assignee;
		dueDate?: string;
		parentTaskId?: string;
		parentTaskTitle?: string;
		tags?: string[];
		subtasks: SubtaskLocal[];
		activity: ActivityEntry[];
		estimatedHours?: number;
		createdAt?: string;
	}

	// ── State ────────────────────────────────────────────────────────────────
	let task = $state<TaskDetail | null>(null);
	let isLoading = $state(true);
	let error = $state('');
	let editingTitle = $state(false);
	let titleDraft = $state('');
	let editingDescription = $state(false);
	let descriptionDraft = $state('');
	let newSubtaskTitle = $state('');
	let addingSubtask = $state(false);
	let isSaving = $state(false);

	// Collapsible detail sections (design-spec style content blocks)
	let expandedSections = $state<Set<string>>(new Set(['overview']));

	const taskId = $derived($page.params.id);
	const isEmbedMode = $derived($page.url.searchParams.get('embed') === 'true');
	const embedSuffix = $derived(isEmbedMode ? '?embed=true' : '');

	// ── Constants ────────────────────────────────────────────────────────────
	const statusOptions: { value: UIStatus; label: string; color: string }[] = [
		{ value: 'todo', label: 'To Do', color: 'var(--bos-status-neutral)' },
		{ value: 'in_progress', label: 'In Progress', color: 'var(--bos-status-info)' },
		{ value: 'in_review', label: 'In Review', color: 'var(--bos-category-ai)' },
		{ value: 'done', label: 'Done', color: 'var(--bos-status-success)' },
		{ value: 'blocked', label: 'Blocked', color: 'var(--bos-status-warning)' }
	];

	const priorityOptions: { value: Priority; label: string; color: string }[] = [
		{ value: 'critical', label: 'Critical', color: 'var(--bos-priority-critical)' },
		{ value: 'high', label: 'High', color: 'var(--bos-priority-high)' },
		{ value: 'medium', label: 'Medium', color: 'var(--bos-priority-medium)' },
		{ value: 'low', label: 'Low', color: 'var(--bos-status-neutral)' }
	];

	const projectColors = ['var(--bos-status-info)', 'var(--bos-status-success)', 'var(--bos-category-ai)', 'var(--bos-status-warning)', 'var(--bos-status-error)', 'var(--bos-category-automation)'];

	// ── Helpers ───────────────────────────────────────────────────────────────
	function mapStatus(s: string): UIStatus {
		switch (s) {
			case 'todo': return 'todo';
			case 'in_progress': return 'in_progress';
			case 'done': return 'done';
			case 'cancelled': return 'blocked';
			default: return 'todo';
		}
	}

	function mapStatusToApi(s: UIStatus): APITaskStatus {
		switch (s) {
			case 'todo': return 'todo';
			case 'in_progress': return 'in_progress';
			case 'in_review': return 'in_progress';
			case 'done': return 'done';
			case 'blocked': return 'cancelled';
			default: return 'todo';
		}
	}

	function formatDate(dateStr?: string | null): string {
		if (!dateStr) return '';
		const d = new Date(dateStr);
		return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function formatRelativeTime(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(diff / 3600000);
		const days = Math.floor(diff / 86400000);
		if (minutes < 1) return 'Just now';
		if (minutes < 60) return `${minutes}m ago`;
		if (hours < 24) return `${hours}h ago`;
		if (days < 7) return `${days}d ago`;
		return date.toLocaleDateString();
	}

	function isToday(dateStr?: string | null): boolean {
		if (!dateStr) return false;
		const today = new Date().toISOString().split('T')[0];
		return dateStr.split('T')[0] === today;
	}

	function currentStatus() {
		return statusOptions.find(s => s.value === task?.status) ?? statusOptions[0];
	}

	function currentPriority() {
		return priorityOptions.find(p => p.value === task?.priority) ?? priorityOptions[2];
	}

	function toggleSection(id: string) {
		if (expandedSections.has(id)) {
			expandedSections.delete(id);
		} else {
			expandedSections.add(id);
		}
		expandedSections = new Set(expandedSections);
	}

	// ── Data loading ──────────────────────────────────────────────────────────
	onMount(async () => {
		team.loadMembers();
		await loadTask();
	});

	async function loadTask() {
		isLoading = true;
		error = '';
		try {
			const allTasks = await api.getTasks();
			const apiTask = allTasks.find(t => t.id === taskId);
			if (!apiTask) {
				error = 'Task not found';
				return;
			}

			// Load project info if available
			let projectName: string | undefined;
			let projectColor: string | undefined;
			if (apiTask.project_id) {
				try {
					const projects = await api.getProjects();
					const proj = projects.find(p => p.id === apiTask.project_id);
					if (proj) {
						const idx = projects.indexOf(proj);
						projectName = proj.name;
						projectColor = projectColors[idx % projectColors.length];
					}
				} catch {
					// project load failure is non-fatal
				}
			}

			// Load assignee info if available
			let assignee: Assignee | undefined;
			if (apiTask.assignee_id) {
				try {
					const members = await api.getTeamMembers();
					const member = members.find(m => m.id === apiTask.assignee_id);
					if (member) {
						assignee = { id: member.id, name: member.name, avatar: member.avatar_url ?? undefined };
					}
				} catch {
					// assignee load failure is non-fatal
				}
			}

			// Load subtasks (tasks with parent_task_id === this task's id)
			let subtasks: SubtaskLocal[] = [];
			try {
				const childTasks = allTasks.filter(t => t.parent_task_id === taskId);
				subtasks = childTasks.map(t => ({
					id: t.id,
					title: t.title,
					completed: t.status === 'done'
				}));
			} catch {
				// non-fatal
			}

			// Find parent task title if applicable
			let parentTaskTitle: string | undefined;
			if (apiTask.parent_task_id) {
				const parent = allTasks.find(t => t.id === apiTask.parent_task_id);
				if (parent) parentTaskTitle = parent.title;
			}

			task = {
				id: apiTask.id,
				title: apiTask.title,
				description: apiTask.description ?? undefined,
				status: mapStatus(apiTask.status),
				priority: apiTask.priority as Priority,
				projectId: apiTask.project_id ?? undefined,
				projectName,
				projectColor,
				assignee,
				dueDate: apiTask.due_date ?? undefined,
				parentTaskId: apiTask.parent_task_id ?? undefined,
				parentTaskTitle,
				subtasks,
				activity: [],
				estimatedHours: apiTask.estimated_hours ?? undefined,
				createdAt: apiTask.created_at
			};
		} catch (err) {
			error = 'Failed to load task';
			console.error('Error loading task:', err);
		} finally {
			isLoading = false;
		}
	}

	// ── Mutations ─────────────────────────────────────────────────────────────
	async function updateField(data: UpdateTaskData) {
		if (!task) return;
		isSaving = true;
		try {
			const updated = await api.updateTask(task.id, data);
			// Patch local state from API response
			task = {
				...task,
				status: mapStatus(updated.status),
				priority: updated.priority as Priority,
				description: updated.description ?? undefined,
				dueDate: updated.due_date ?? undefined
			};
		} catch (err) {
			console.error('Failed to update task:', err);
		} finally {
			isSaving = false;
		}
	}

	async function handleStatusChange(status: UIStatus) {
		if (!task) return;
		task = { ...task, status };
		await updateField({ status: mapStatusToApi(status) });
	}

	async function handlePriorityChange(priority: Priority) {
		if (!task) return;
		task = { ...task, priority };
		await updateField({ priority });
	}

	function startEditTitle() {
		if (!task) return;
		titleDraft = task.title;
		editingTitle = true;
	}

	async function saveTitle() {
		if (!task || !titleDraft.trim()) { editingTitle = false; return; }
		task = { ...task, title: titleDraft.trim() };
		editingTitle = false;
		await updateField({ title: titleDraft.trim() });
	}

	function startEditDescription() {
		if (!task) return;
		descriptionDraft = task.description ?? '';
		editingDescription = true;
	}

	async function saveDescription() {
		if (!task) return;
		task = { ...task, description: descriptionDraft };
		editingDescription = false;
		await updateField({ description: descriptionDraft });
	}

	async function handleSubtaskToggle(subtaskId: string) {
		if (!task) return;
		const subtask = task.subtasks.find(s => s.id === subtaskId);
		if (!subtask) return;
		const newCompleted = !subtask.completed;
		task = {
			...task,
			subtasks: task.subtasks.map(s =>
				s.id === subtaskId ? { ...s, completed: newCompleted } : s
			)
		};
		try {
			await api.updateTask(subtaskId, { status: newCompleted ? 'done' : 'todo' });
		} catch (err) {
			console.error('Failed to toggle subtask:', err);
			// Revert on failure
			task = {
				...task,
				subtasks: task.subtasks.map(s =>
					s.id === subtaskId ? { ...s, completed: !newCompleted } : s
				)
			};
		}
	}

	async function handleAddSubtask() {
		if (!task || !newSubtaskTitle.trim()) return;
		addingSubtask = true;
		try {
			const newApiTask = await api.createTask({
				title: newSubtaskTitle.trim(),
				parent_task_id: task.id,
				project_id: task.projectId
			});
			task = {
				...task,
				subtasks: [
					...task.subtasks,
					{ id: newApiTask.id, title: newApiTask.title, completed: false }
				]
			};
			newSubtaskTitle = '';
		} catch (err) {
			console.error('Failed to add subtask:', err);
		} finally {
			addingSubtask = false;
		}
	}

	async function handleDueDateChange(e: Event) {
		const input = e.target as HTMLInputElement;
		if (!task) return;
		task = { ...task, dueDate: input.value || undefined };
		await updateField({ due_date: input.value || undefined });
	}

	function handleShare() {
		if (typeof navigator !== 'undefined' && navigator.clipboard) {
			navigator.clipboard.writeText(window.location.href);
		}
	}

	function handleAskClaude() {
		goto(`/chat${embedSuffix}`);
	}

	// Derived
	const completedSubtasks = $derived(task?.subtasks.filter(s => s.completed).length ?? 0);
	const totalSubtasks = $derived(task?.subtasks.length ?? 0);
	const subtaskProgress = $derived(totalSubtasks > 0 ? Math.round((completedSubtasks / totalSubtasks) * 100) : 0);
</script>

<div class="tb-detail-page flex flex-col h-full overflow-hidden">
	{#if isLoading}
		<!-- Loading skeleton -->
		<div class="flex-1 flex items-center justify-center">
			<div class="tb-spinner animate-spin"></div>
		</div>

	{:else if error}
		<!-- Error state -->
		<div class="flex-1 flex flex-col items-center justify-center gap-4">
			<AlertCircle class="w-10 h-10 tb-text-muted" />
			<p class="tb-text-secondary text-sm">{error}</p>
			<a href="/tasks{embedSuffix}" class="btn-pill btn-pill-ghost btn-pill-sm tb-text-secondary">
				<ArrowLeft class="w-4 h-4 mr-1" />
				Back to Tasks
			</a>
		</div>

	{:else if task}
		<!-- ── Scrollable content ─────────────────────────────────────────────── -->
		<div class="flex-1 overflow-y-auto">
			<div class="tb-content-wrapper mx-auto px-6 py-6 pb-16">

				<!-- ── Breadcrumb & actions ─────────────────────────────────────── -->
				<div class="flex items-center justify-between mb-6 gap-4 flex-wrap">
					<!-- Breadcrumb -->
					<nav class="flex items-center gap-1.5 min-w-0" aria-label="Breadcrumb">
						<a
							href="/tasks{embedSuffix}"
							class="tb-breadcrumb-link flex items-center gap-1 text-sm font-medium hover:underline"
							aria-label="Back to Tasks"
						>
							<ArrowLeft class="w-3.5 h-3.5 flex-shrink-0" />
							Tasks
						</a>
						{#if task.projectName}
							<ChevronRight class="w-3.5 h-3.5 tb-text-muted flex-shrink-0" />
							<a
								href="/projects/{task.projectId}{embedSuffix}"
								class="tb-breadcrumb-link text-sm font-medium truncate max-w-[160px] hover:underline"
							>
								{task.projectName}
							</a>
						{/if}
						<ChevronRight class="w-3.5 h-3.5 tb-text-muted flex-shrink-0" />
						<span class="tb-text-muted text-sm truncate max-w-[200px]" aria-current="page">
							{task.title}
						</span>
					</nav>

					<!-- Action buttons -->
					<div class="flex items-center gap-2 flex-shrink-0">
						{#if isSaving}
							<span class="text-xs tb-text-muted flex items-center gap-1">
								<Loader class="w-3 h-3 animate-spin" />
								Saving…
							</span>
						{/if}
						<button
							onclick={handleShare}
							class="btn-pill btn-pill-soft btn-pill-sm flex items-center gap-1.5"
							aria-label="Copy link to this task"
						>
							<Share2 class="w-3.5 h-3.5" />
							Share
						</button>
						<button
							onclick={handleAskClaude}
							class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-1.5"
							aria-label="Ask Claude AI about this task"
						>
							<Bot class="w-3.5 h-3.5" />
							Ask Claude AI
						</button>
					</div>
				</div>

				<!-- ── Title row ────────────────────────────────────────────────── -->
				<div class="mb-5">
					{#if editingTitle}
						<div class="flex items-start gap-2">
							<input
								type="text"
								bind:value={titleDraft}
								onkeydown={(e) => { if (e.key === 'Enter') saveTitle(); if (e.key === 'Escape') editingTitle = false; }}
								class="tb-title-input flex-1 text-2xl font-semibold focus:outline-none"
								aria-label="Edit task title"
								autofocus
							/>
							<div class="flex gap-1.5 mt-1">
								<button onclick={saveTitle} class="btn-pill btn-pill-primary btn-pill-xs" aria-label="Save title">
									<Check class="w-3.5 h-3.5" />
								</button>
								<button onclick={() => editingTitle = false} class="btn-pill btn-pill-soft btn-pill-xs" aria-label="Cancel editing title">
									Cancel
								</button>
							</div>
						</div>
					{:else}
						<div class="flex items-start gap-3 group">
							<h1 class="tb-title flex-1 text-2xl font-semibold leading-tight {task.status === 'done' ? 'line-through tb-text-muted' : ''}">
								{task.title}
							</h1>
							<button
								onclick={startEditTitle}
								class="btn-pill btn-pill-icon btn-pill-ghost btn-pill-xs mt-1 opacity-0 group-hover:opacity-100 transition-opacity"
								aria-label="Edit task title"
							>
								<Pencil class="w-3.5 h-3.5" />
							</button>
						</div>
					{/if}
				</div>

				<!-- ── Properties strip ─────────────────────────────────────────── -->
				<div class="tb-properties-strip flex flex-wrap items-start gap-4 mb-7 pb-6">

					<!-- Status -->
					<div class="tb-property-block">
						<span class="tb-property-label">Status</span>
						<DropdownMenu.Root>
							<DropdownMenu.Trigger
								class="tb-property-value tb-status-trigger flex items-center gap-1.5 cursor-pointer"
								aria-label="Change task status"
							>
								<span class="w-2 h-2 rounded-full flex-shrink-0" style="background-color: {currentStatus().color}"></span>
								<span class="text-sm font-medium">{currentStatus().label}</span>
								<ChevronDown class="w-3.5 h-3.5 tb-text-muted ml-0.5" />
							</DropdownMenu.Trigger>
							<DropdownMenu.Portal>
								<DropdownMenu.Content
									class="tb-dropdown z-50 min-w-[160px] p-1 animate-in fade-in-0 zoom-in-95"
									sideOffset={4}
								>
									{#each statusOptions as opt}
										<DropdownMenu.Item
											class="tb-dropdown-item flex items-center gap-2 px-3 py-2 text-sm rounded-lg cursor-pointer"
											onclick={() => handleStatusChange(opt.value)}
										>
											<span class="w-2 h-2 rounded-full flex-shrink-0" style="background-color: {opt.color}"></span>
											{opt.label}
											{#if task.status === opt.value}
												<Check class="w-3.5 h-3.5 ml-auto tb-text-muted" />
											{/if}
										</DropdownMenu.Item>
									{/each}
								</DropdownMenu.Content>
							</DropdownMenu.Portal>
						</DropdownMenu.Root>
					</div>

					<!-- Priority -->
					<div class="tb-property-block">
						<span class="tb-property-label">Priority</span>
						<DropdownMenu.Root>
							<DropdownMenu.Trigger
								class="tb-property-value tb-status-trigger flex items-center gap-1.5 cursor-pointer"
								aria-label="Change task priority"
							>
								<span class="w-2 h-2 rounded-full flex-shrink-0" style="background-color: {currentPriority().color}"></span>
								<span class="text-sm font-medium">{currentPriority().label}</span>
								<ChevronDown class="w-3.5 h-3.5 tb-text-muted ml-0.5" />
							</DropdownMenu.Trigger>
							<DropdownMenu.Portal>
								<DropdownMenu.Content
									class="tb-dropdown z-50 min-w-[140px] p-1 animate-in fade-in-0 zoom-in-95"
									sideOffset={4}
								>
									{#each priorityOptions as opt}
										<DropdownMenu.Item
											class="tb-dropdown-item flex items-center gap-2 px-3 py-2 text-sm rounded-lg cursor-pointer"
											onclick={() => handlePriorityChange(opt.value)}
										>
											<span class="w-2 h-2 rounded-full flex-shrink-0" style="background-color: {opt.color}"></span>
											{opt.label}
											{#if task.priority === opt.value}
												<Check class="w-3.5 h-3.5 ml-auto tb-text-muted" />
											{/if}
										</DropdownMenu.Item>
									{/each}
								</DropdownMenu.Content>
							</DropdownMenu.Portal>
						</DropdownMenu.Root>
					</div>

					<!-- Due date -->
					<div class="tb-property-block">
						<span class="tb-property-label">
							<Calendar class="w-3 h-3 inline mr-0.5 relative -top-px" />
							Due Date
						</span>
						<div class="tb-property-value">
							<label class="flex items-center gap-1.5 cursor-pointer" aria-label="Set due date">
								<span class="text-sm" style="{task.dueDate && new Date(task.dueDate) < new Date() && task.status !== 'done' ? 'color: var(--bos-status-error)' : ''}">
									{#if task.dueDate}
										{isToday(task.dueDate) ? 'Today' : ''}
										{formatDate(task.dueDate)}
									{:else}
										<span class="tb-text-muted">No date</span>
									{/if}
								</span>
								<input
									type="date"
									value={task.dueDate?.split('T')[0] ?? ''}
									onchange={handleDueDateChange}
									class="tb-date-input sr-only"
									aria-label="Change due date"
								/>
							</label>
						</div>
					</div>

					<!-- Assignee -->
					<div class="tb-property-block">
						<span class="tb-property-label">
							<User class="w-3 h-3 inline mr-0.5 relative -top-px" />
							Assignee
						</span>
						<div class="tb-property-value">
							{#if task.assignee}
								<div class="flex items-center gap-1.5">
									<div class="w-5 h-5 rounded-full tb-avatar flex items-center justify-center text-xs font-medium flex-shrink-0">
										{task.assignee.name.charAt(0).toUpperCase()}
									</div>
									<span class="text-sm">{task.assignee.name}</span>
								</div>
							{:else}
								<span class="tb-text-muted text-sm">Unassigned</span>
							{/if}
						</div>
					</div>

					<!-- Project -->
					{#if task.projectName}
						<div class="tb-property-block">
							<span class="tb-property-label">Project</span>
							<a
								href="/projects/{task.projectId}{embedSuffix}"
								class="tb-property-value tb-project-link flex items-center gap-1.5 text-sm hover:underline"
							>
								<span class="w-2.5 h-2.5 rounded flex-shrink-0" style="background-color: {task.projectColor ?? 'var(--bos-status-neutral)'}"></span>
								{task.projectName}
							</a>
						</div>
					{/if}

					<!-- Parent task -->
					{#if task.parentTaskTitle}
						<div class="tb-property-block">
							<span class="tb-property-label">Subtask of</span>
							<a
								href="/tasks/{task.parentTaskId}{embedSuffix}"
								class="tb-property-value tb-project-link flex items-center gap-1.5 text-sm hover:underline"
							>
								<Link class="w-3 h-3 flex-shrink-0" />
								{task.parentTaskTitle}
							</a>
						</div>
					{/if}

					<!-- Estimated hours -->
					{#if task.estimatedHours}
						<div class="tb-property-block">
							<span class="tb-property-label">
								<Clock class="w-3 h-3 inline mr-0.5 relative -top-px" />
								Estimate
							</span>
							<div class="tb-property-value text-sm">{task.estimatedHours}h</div>
						</div>
					{/if}
				</div>

				<!-- ── Overview / Description ───────────────────────────────────── -->
				<section class="mb-2">
					<button
						onclick={() => toggleSection('overview')}
						class="tb-section-toggle w-full flex items-center gap-2 mb-3"
						aria-expanded={expandedSections.has('overview')}
					>
						<ChevronDown class="w-4 h-4 tb-text-muted transition-transform {expandedSections.has('overview') ? '' : '-rotate-90'}" />
						<h2 class="tb-section-heading text-sm font-semibold uppercase tracking-wide">Overview</h2>
					</button>

					{#if expandedSections.has('overview')}
						<div class="tb-section-body mb-6">
							{#if editingDescription}
								<textarea
									bind:value={descriptionDraft}
									rows={5}
									class="tb-textarea w-full px-3 py-2 text-sm resize-none focus:outline-none rounded-lg"
									placeholder="Add a description for this task…"
									aria-label="Task description"
								></textarea>
								<div class="flex gap-2 mt-2">
									<button onclick={saveDescription} class="btn-pill btn-pill-primary btn-pill-xs" aria-label="Save description">
										Save
									</button>
									<button onclick={() => editingDescription = false} class="btn-pill btn-pill-soft btn-pill-xs" aria-label="Cancel editing">
										Cancel
									</button>
								</div>
							{:else}
								<div
									class="tb-description-area group relative px-3 py-2 rounded-lg min-h-[60px] cursor-text"
									onclick={startEditDescription}
									onkeydown={(e) => e.key === 'Enter' && startEditDescription()}
									role="button"
									tabindex="0"
									aria-label="Click to edit description"
								>
									{#if task.description}
										<p class="text-sm tb-text leading-relaxed whitespace-pre-wrap">{task.description}</p>
									{:else}
										<p class="text-sm tb-text-muted italic">No description yet. Click to add one.</p>
									{/if}
									<span class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
										<Pencil class="w-3.5 h-3.5 tb-text-muted" />
									</span>
								</div>
							{/if}
						</div>
					{/if}
				</section>

				<!-- ── Subtasks ──────────────────────────────────────────────────── -->
				<section class="mb-2">
					<button
						onclick={() => toggleSection('subtasks')}
						class="tb-section-toggle w-full flex items-center gap-2 mb-3"
						aria-expanded={expandedSections.has('subtasks')}
					>
						<ChevronDown class="w-4 h-4 tb-text-muted transition-transform {expandedSections.has('subtasks') ? '' : '-rotate-90'}" />
						<h2 class="tb-section-heading text-sm font-semibold uppercase tracking-wide">
							Subtasks
						</h2>
						{#if totalSubtasks > 0}
							<span class="tb-badge ml-auto text-xs">{completedSubtasks}/{totalSubtasks}</span>
							<!-- Progress bar -->
							<div class="w-20 h-1 tb-progress-bg rounded-full overflow-hidden">
								<div
									class="h-full tb-progress-fill rounded-full transition-all"
									style="width: {subtaskProgress}%"
								></div>
							</div>
						{/if}
					</button>

					{#if expandedSections.has('subtasks')}
						<div class="tb-section-body mb-6">
							<!-- Subtask list -->
							{#if task.subtasks.length > 0}
								<ul class="space-y-1 mb-3" role="list">
									{#each task.subtasks as subtask}
										<li>
											<label
												class="tb-subtask-row flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer"
											>
												<input
													type="checkbox"
													checked={subtask.completed}
													onchange={() => handleSubtaskToggle(subtask.id)}
													class="tb-checkbox w-4 h-4 rounded"
													aria-label="Toggle subtask: {subtask.title}"
												/>
												<a
													href="/tasks/{subtask.id}{embedSuffix}"
													onclick={(e) => e.stopPropagation()}
													class="flex-1 text-sm {subtask.completed ? 'line-through tb-text-muted' : 'tb-text'} hover:underline"
												>
													{subtask.title}
												</a>
												{#if subtask.completed}
													<CircleCheck class="w-4 h-4 flex-shrink-0" style="color: var(--bos-status-success)" />
												{/if}
											</label>
										</li>
									{/each}
								</ul>
							{/if}

							<!-- Add subtask row -->
							<div class="flex items-center gap-2 px-3 py-1.5">
								<Plus class="w-4 h-4 tb-text-muted flex-shrink-0" />
								<input
									type="text"
									bind:value={newSubtaskTitle}
									onkeydown={(e) => e.key === 'Enter' && !addingSubtask && handleAddSubtask()}
									placeholder="Add a subtask…"
									class="tb-subtask-input flex-1 text-sm focus:outline-none bg-transparent"
									aria-label="New subtask title"
								/>
								{#if newSubtaskTitle.trim()}
									<button
										onclick={handleAddSubtask}
										disabled={addingSubtask}
										class="btn-pill btn-pill-primary btn-pill-xs disabled:opacity-50"
										aria-label="Add subtask"
									>
										{addingSubtask ? 'Adding…' : 'Add'}
									</button>
								{/if}
							</div>
						</div>
					{/if}
				</section>

				<!-- ── Activity ──────────────────────────────────────────────────── -->
				<section class="mb-2">
					<button
						onclick={() => toggleSection('activity')}
						class="tb-section-toggle w-full flex items-center gap-2 mb-3"
						aria-expanded={expandedSections.has('activity')}
					>
						<ChevronDown class="w-4 h-4 tb-text-muted transition-transform {expandedSections.has('activity') ? '' : '-rotate-90'}" />
						<h2 class="tb-section-heading text-sm font-semibold uppercase tracking-wide">Activity</h2>
					</button>

					{#if expandedSections.has('activity')}
						<div class="tb-section-body mb-6">
							{#if task.activity.length > 0}
								<ul class="space-y-3" role="list">
									{#each task.activity as entry}
										<li class="flex items-start gap-3 text-sm">
											<div class="w-1.5 h-1.5 rounded-full tb-activity-dot mt-2 flex-shrink-0"></div>
											<div>
												<span class="tb-text">{entry.description}</span>
												<span class="tb-text-muted"> — {formatRelativeTime(entry.createdAt)}</span>
											</div>
										</li>
									{/each}
								</ul>
							{:else}
								<!-- Created entry from metadata -->
								<ul class="space-y-3" role="list">
									{#if task.createdAt}
										<li class="flex items-start gap-3 text-sm">
											<div class="w-1.5 h-1.5 rounded-full tb-activity-dot mt-2 flex-shrink-0"></div>
											<div>
												<span class="tb-text">Task created</span>
												<span class="tb-text-muted"> — {formatRelativeTime(task.createdAt)}</span>
											</div>
										</li>
									{/if}
								</ul>
							{/if}
						</div>
					{/if}
				</section>

			</div>
		</div>
	{/if}
</div>

<style>
	/* ── Page shell ─────────────────────────────────────────────────────────── */
	.tb-detail-page {
		background: var(--dbg);
	}

	.tb-content-wrapper {
		max-width: 860px;
		width: 100%;
	}

	/* ── Typography ─────────────────────────────────────────────────────────── */
	.tb-title {
		color: var(--dt);
	}
	.tb-text {
		color: var(--dt);
	}
	.tb-text-muted {
		color: var(--dt3, #888);
	}
	.tb-text-secondary {
		color: var(--dt2);
	}

	/* ── Breadcrumb ─────────────────────────────────────────────────────────── */
	.tb-breadcrumb-link {
		color: var(--dt2);
		transition: color 150ms ease;
	}
	.tb-breadcrumb-link:hover {
		color: var(--dt);
	}

	/* ── Title input ────────────────────────────────────────────────────────── */
	.tb-title-input {
		color: var(--dt);
		background: transparent;
		border: none;
		border-bottom: 2px solid var(--dbd);
		padding: 0 0 4px;
		width: 100%;
	}
	.tb-title-input:focus {
		border-bottom-color: var(--bos-nav-active);
	}

	/* ── Properties strip ───────────────────────────────────────────────────── */
	.tb-properties-strip {
		border-bottom: 1px solid var(--dbd2, #f0f0f0);
	}

	.tb-property-block {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 80px;
	}

	.tb-property-label {
		font-size: 0.65rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--dt4, #bbb);
	}

	.tb-property-value {
		color: var(--dt);
	}

	.tb-status-trigger {
		padding: 4px 8px;
		border-radius: 6px;
		border: 1px solid var(--dbd);
		background: var(--dbg);
		transition: background 150ms ease;
	}
	.tb-status-trigger:hover {
		background: var(--dbg2, #f5f5f5);
	}

	.tb-project-link {
		color: var(--bos-nav-active);
	}

	/* ── Avatar ─────────────────────────────────────────────────────────────── */
	.tb-avatar {
		background: linear-gradient(135deg, var(--bos-nav-active) 0%, var(--bos-category-productivity) 100%);
		color: var(--bos-surface-on-color);
		font-size: 0.7rem;
	}

	/* ── Date input (visually hidden but accessible) ────────────────────────── */
	.tb-date-input {
		width: 0;
		height: 0;
		opacity: 0;
		position: absolute;
		pointer-events: none;
	}

	/* ── Sections ───────────────────────────────────────────────────────────── */
	.tb-section-toggle {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		text-align: left;
	}
	.tb-section-heading {
		color: var(--dt3, #888);
	}
	.tb-section-body {
		padding-left: 1.5rem;
	}

	/* ── Description ────────────────────────────────────────────────────────── */
	.tb-description-area {
		background: var(--dbg2, #f5f5f5);
		transition: background 150ms ease;
	}
	.tb-description-area:hover {
		background: var(--dbg3, #eee);
	}

	.tb-textarea {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt);
		border: 1px solid var(--dbd);
		transition: border-color 150ms ease;
	}
	.tb-textarea:focus {
		border-color: var(--bos-nav-active);
	}

	/* ── Subtasks ───────────────────────────────────────────────────────────── */
	.tb-subtask-row {
		transition: background 150ms ease;
	}
	.tb-subtask-row:hover {
		background: var(--dbg2, #f5f5f5);
	}

	.tb-checkbox {
		accent-color: var(--bos-nav-active);
		flex-shrink: 0;
	}

	.tb-subtask-input {
		color: var(--dt);
	}
	.tb-subtask-input::placeholder {
		color: var(--dt4, #bbb);
	}

	/* ── Progress ───────────────────────────────────────────────────────────── */
	.tb-progress-bg {
		background: var(--dbd);
	}
	.tb-progress-fill {
		background: var(--bos-status-success);
	}

	/* ── Badge ──────────────────────────────────────────────────────────────── */
	.tb-badge {
		color: var(--dt3, #888);
		background: var(--dbg2, #f5f5f5);
		padding: 1px 6px;
		border-radius: 999px;
		font-size: 0.65rem;
		font-weight: 600;
	}

	/* ── Activity ───────────────────────────────────────────────────────────── */
	.tb-activity-dot {
		background: var(--dt4, #bbb);
	}

	/* ── Dropdown ───────────────────────────────────────────────────────────── */
	.tb-dropdown {
		background: var(--dbg);
		border: 1px solid var(--dbd);
		border-radius: 12px;
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
	}
	.tb-dropdown-item {
		color: var(--dt);
		transition: background 120ms ease;
	}
	.tb-dropdown-item:hover {
		background: var(--dbg2, #f5f5f5);
	}

	/* ── Spinner ────────────────────────────────────────────────────────────── */
	.tb-spinner {
		width: 32px;
		height: 32px;
		border: 2px solid var(--dbd);
		border-top-color: var(--dt);
		border-radius: 50%;
	}
</style>
