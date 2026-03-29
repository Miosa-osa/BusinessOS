<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { useSession } from '$lib/auth-client';
	import { moduleEvents } from '$lib/stores/events';
	import {
		TaskListView,
		TaskBoardView,
		TasksToolbar,
		TaskQuickFilters,
		NewTaskModal
		// TaskDetailSlideOver is deprecated — task detail is now at /tasks/[id]
		// TaskCalendarView is disconnected — calendar view removed (#57)
	} from '$lib/components/tasks';
	import { api, type Task as APITask, type Project as APIProject, type TaskStatus as APITaskStatus } from '$lib/api';

	const session = useSession();

	type TaskStatus = 'todo' | 'in_progress' | 'in_review' | 'done' | 'blocked';
	type Priority = 'critical' | 'high' | 'medium' | 'low';
	type ViewMode = 'list' | 'board';
	type GroupBy = 'status' | 'priority' | 'project' | 'assignee' | 'none';
	type QuickFilter = 'my-tasks' | 'all' | 'overdue' | 'today' | 'this-week' | 'blocked' | 'unassigned';

	interface Assignee {
		id: string;
		name: string;
		avatar?: string;
	}

	interface Task {
		id: string;
		title: string;
		description?: string;
		status: TaskStatus;
		priority: Priority;
		projectId?: string;
		projectName?: string;
		projectColor?: string;
		assignee?: Assignee;
		dueDate?: string;
		tags?: string[];
		subtasks?: { id: string; title: string; completed: boolean }[];
		comments?: { id: string; authorId: string; authorName: string; content: string; createdAt: string }[];
		activity?: { id: string; type: string; description: string; createdAt: string }[];
		createdAt?: string;
	}

	interface Project {
		id: string;
		name: string;
		color: string;
	}

	// State
	let viewMode = $state<ViewMode>('list');
	let groupBy = $state<GroupBy>('status');
	let searchQuery = $state('');
	let quickFilter = $state<QuickFilter>('all');
	let showNewTaskModal = $state(false);
	let isLoading = $state(true);

	// Data from API
	let projects = $state<Project[]>([]);
	let teamMembers = $state<Assignee[]>([]);
	let tasks = $state<Task[]>([]);

	// Project colors for display
	const projectColors = ['var(--bos-status-info)', 'var(--bos-status-success)', 'var(--bos-category-ai)', 'var(--bos-status-warning)', 'var(--bos-status-error)', 'var(--bos-category-automation)'];

	// Load data from API
	async function loadData() {
		isLoading = true;
		try {
			const [apiTasks, apiProjects, apiTeam] = await Promise.all([
				api.getTasks().catch(() => []),
				api.getProjects().catch(() => []),
				api.getTeamMembers().catch(() => [])
			]);

			// Transform projects
			projects = apiProjects.map((p, i) => ({
					id: p.id,
					name: p.name,
					color: projectColors[i % projectColors.length]
				}));

			// Transform team members
			teamMembers = apiTeam.map(m => ({
					id: m.id,
					name: m.name,
					avatar: m.avatar_url || undefined
				}));

			// Transform tasks
			tasks = apiTasks.map(t => {
					const project = projects.find(p => p.id === t.project_id);
					const assignee = teamMembers.find(m => m.id === t.assignee_id);
					return {
						id: t.id,
						title: t.title,
						description: t.description || undefined,
						status: mapStatus(t.status),
						priority: t.priority as Priority,
						projectId: t.project_id || undefined,
						projectName: project?.name,
						projectColor: project?.color,
						assignee: assignee,
						dueDate: t.due_date || undefined,
						createdAt: t.created_at
					};
				});
		} catch (error) {
			console.error('Failed to load tasks:', error);
		} finally {
			isLoading = false;
		}
	}

	function mapStatus(status: string): TaskStatus {
		// Map API status to UI status
		switch (status) {
			case 'todo': return 'todo';
			case 'in_progress': return 'in_progress';
			case 'done': return 'done';
			case 'cancelled': return 'blocked';
			default: return 'todo';
		}
	}

	function mapStatusToApi(status: TaskStatus): APITaskStatus {
		switch (status) {
			case 'todo': return 'todo';
			case 'in_progress': return 'in_progress';
			case 'in_review': return 'in_progress';
			case 'done': return 'done';
			case 'blocked': return 'todo';
			default: return 'todo';
		}
	}

	// Load on mount
	onMount(() => {
		loadData();
	});

	// Reload when a task is created from the chat slash command
	$effect(() => {
		const event = $moduleEvents;
		if (event?.type === 'task:created') {
			loadData();
		}
	});

	// Quick filter counts
	const currentUserId = $derived($session.data?.user?.id);
	const filterCounts = $derived({
		'my-tasks': tasks.filter(t => t.assignee?.id === currentUserId).length,
		'all': tasks.length,
		'overdue': tasks.filter(t => {
			if (!t.dueDate || t.status === 'done') return false;
			return new Date(t.dueDate) < new Date();
		}).length,
		'today': tasks.filter(t => {
			if (!t.dueDate) return false;
			const today = new Date().toISOString().split('T')[0];
			return t.dueDate.split('T')[0] === today;
		}).length,
		'this-week': tasks.filter(t => {
			if (!t.dueDate) return false;
			const due = new Date(t.dueDate);
			const now = new Date();
			const weekFromNow = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);
			return due >= now && due <= weekFromNow;
		}).length,
		'blocked': tasks.filter(t => t.status === 'blocked').length,
		'unassigned': tasks.filter(t => !t.assignee).length
	});

	// Filtered tasks based on quick filter
	const filteredTasks = $derived(() => {
		let filtered = tasks;

		// Apply search
		if (searchQuery) {
			const query = searchQuery.toLowerCase();
			filtered = filtered.filter(t =>
				t.title.toLowerCase().includes(query) ||
				t.projectName?.toLowerCase().includes(query) ||
				t.tags?.some(tag => tag.toLowerCase().includes(query))
			);
		}

		// Apply quick filter
		switch (quickFilter) {
			case 'my-tasks':
				filtered = filtered.filter(t => t.assignee?.id === 'user-1');
				break;
			case 'overdue':
				filtered = filtered.filter(t => {
					if (!t.dueDate || t.status === 'done') return false;
					return new Date(t.dueDate) < new Date();
				});
				break;
			case 'today':
				filtered = filtered.filter(t => {
					if (!t.dueDate) return false;
					const today = new Date().toISOString().split('T')[0];
					return t.dueDate.split('T')[0] === today;
				});
				break;
			case 'this-week':
				filtered = filtered.filter(t => {
					if (!t.dueDate) return false;
					const due = new Date(t.dueDate);
					const now = new Date();
					const weekFromNow = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);
					return due >= now && due <= weekFromNow;
				});
				break;
			case 'blocked':
				filtered = filtered.filter(t => t.status === 'blocked');
				break;
			case 'unassigned':
				filtered = filtered.filter(t => !t.assignee);
				break;
		}

		return filtered;
	});

	function handleTaskClick(taskId: string) {
		// Navigate to the full task detail page instead of opening slide-over
		const embedSuffix = $page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '';
		goto(`/tasks/${taskId}${embedSuffix}`);
	}

	async function handleTaskStatusChange(taskId: string, status: TaskStatus) {
		try {
			await api.updateTask(taskId, { status: mapStatusToApi(status) });
			tasks = tasks.map(t => t.id === taskId ? { ...t, status } : t);
		} catch (error) {
			console.error('Failed to update task status:', error);
		}
	}

	function handleTaskEdit(taskId: string) {
		handleTaskClick(taskId);
	}

	async function handleTaskDuplicate(taskId: string) {
		const task = tasks.find(t => t.id === taskId);
		if (task) {
			try {
				const newApiTask = await api.createTask({
					title: `${task.title} (Copy)`,
					description: task.description,
					priority: task.priority,
					project_id: task.projectId,
					assignee_id: task.assignee?.id,
					due_date: task.dueDate
				});
				const project = projects.find(p => p.id === newApiTask.project_id);
				const assignee = teamMembers.find(m => m.id === newApiTask.assignee_id);
				const newTask: Task = {
					id: newApiTask.id,
					title: newApiTask.title,
					description: newApiTask.description || undefined,
					status: mapStatus(newApiTask.status),
					priority: newApiTask.priority as Priority,
					projectId: newApiTask.project_id || undefined,
					projectName: project?.name,
					projectColor: project?.color,
					assignee: assignee,
					dueDate: newApiTask.due_date || undefined,
					createdAt: newApiTask.created_at
				};
				tasks = [...tasks, newTask];
			} catch (error) {
				console.error('Failed to duplicate task:', error);
			}
		}
	}

	async function handleTaskDelete(taskId: string) {
		try {
			await api.deleteTask(taskId);
			tasks = tasks.filter(t => t.id !== taskId);
		} catch (error) {
			console.error('Failed to delete task:', error);
		}
	}

	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	function handleAddTask(_arg?: any) {
		showNewTaskModal = true;
	}

	async function handleCreateTask(taskData: any) {
		try {
			const newApiTask = await api.createTask({
				title: taskData.title,
				description: taskData.description,
				priority: taskData.priority,
				project_id: taskData.projectId,
				assignee_id: taskData.assigneeId,
				due_date: taskData.dueDate
			});
			const project = projects.find(p => p.id === newApiTask.project_id);
			const assignee = teamMembers.find(m => m.id === newApiTask.assignee_id);
			const newTask: Task = {
				id: newApiTask.id,
				title: newApiTask.title,
				description: newApiTask.description || undefined,
				status: mapStatus(newApiTask.status),
				priority: newApiTask.priority as Priority,
				projectId: newApiTask.project_id || undefined,
				projectName: project?.name,
				projectColor: project?.color,
				assignee: assignee,
				dueDate: newApiTask.due_date || undefined,
				createdAt: newApiTask.created_at
			};
			tasks = [...tasks, newTask];
			showNewTaskModal = false;
		} catch (error) {
			console.error('Failed to create task:', error);
		}
	}

	// Keyboard shortcuts
	onMount(() => {
		const handleKeydown = (e: KeyboardEvent) => {
			if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

			if (e.key === 'n' && !e.metaKey && !e.ctrlKey) {
				e.preventDefault();
				showNewTaskModal = true;
			}
		};

		document.addEventListener('keydown', handleKeydown);
		return () => document.removeEventListener('keydown', handleKeydown);
	});
</script>

<div class="tb-page flex flex-col h-full overflow-hidden">
	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 tb-page-spinner border-t-transparent rounded-full"></div>
		</div>
	{:else}
		<!-- Toolbar (now includes New Task button) -->
		<TasksToolbar
			bind:view={viewMode}
			bind:groupBy
			bind:searchQuery
			onViewChange={(v) => viewMode = v}
			onGroupByChange={(g) => groupBy = g}
			onSearchChange={(q) => searchQuery = q}
			onNewTask={() => showNewTaskModal = true}
		/>

		<!-- Quick Filters -->
		<TaskQuickFilters
			activeFilter={quickFilter}
			counts={filterCounts}
			onFilterChange={(f) => quickFilter = f}
		/>

		<!-- Content -->
		{#if viewMode === 'list'}
			<TaskListView
				tasks={filteredTasks()}
				{groupBy}
				onTaskClick={handleTaskClick}
				onTaskStatusChange={handleTaskStatusChange}
				onTaskEdit={handleTaskEdit}
				onTaskDuplicate={handleTaskDuplicate}
				onTaskDelete={handleTaskDelete}
				onAddTask={handleAddTask}
			/>
		{:else if viewMode === 'board'}
			<TaskBoardView
				tasks={filteredTasks()}
				onTaskClick={handleTaskClick}
				onTaskStatusChange={handleTaskStatusChange}
				onAddTask={handleAddTask}
			/>
		{/if}
	{/if}
</div>

<!-- New Task Modal -->
<NewTaskModal
	bind:open={showNewTaskModal}
	{projects}
	{teamMembers}
	onCreate={handleCreateTask}
/>

<!-- Task Detail: navigates to /tasks/[id] (full page) — slide-over removed -->

<style>
	.tb-page {
		background: var(--dbg);
	}
	.tb-page-spinner {
		border-color: var(--dt);
	}
</style>
