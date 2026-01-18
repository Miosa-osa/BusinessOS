<script lang="ts">
	import { fade, slide } from 'svelte/transition';
	import TaskRow from './TaskRow.svelte';
	import TaskGroupHeader from './TaskGroupHeader.svelte';
	import TaskInlineAdd from './TaskInlineAdd.svelte';

	type TaskStatus = 'todo' | 'in_progress' | 'in_review' | 'done' | 'blocked';
	type Priority = 'critical' | 'high' | 'medium' | 'low';
	type GroupBy = 'status' | 'priority' | 'project' | 'assignee' | 'none';

	interface Assignee {
		id: string;
		name: string;
		avatar?: string;
	}

	interface Task {
		id: string;
		title: string;
		status: TaskStatus;
		priority: Priority;
		projectId?: string;
		projectName?: string;
		projectColor?: string;
		assignee?: Assignee;
		dueDate?: string;
		tags?: string[];
		// Allow TaskRow to receive projectId for navigation
	}

	interface Props {
		tasks: Task[];
		groupBy?: GroupBy;
		showDoneTasks?: boolean;
		onTaskClick?: (taskId: string) => void;
		onTaskStatusChange?: (taskId: string, status: TaskStatus) => void;
		onTaskEdit?: (taskId: string) => void;
		onTaskDuplicate?: (taskId: string) => void;
		onTaskDelete?: (taskId: string) => void;
		onAddTask?: (task: { title: string; priority: Priority; status?: string }) => void;
	}

	let {
		tasks,
		groupBy = 'status',
		showDoneTasks = true,
		onTaskClick,
		onTaskStatusChange,
		onTaskEdit,
		onTaskDuplicate,
		onTaskDelete,
		onAddTask
	}: Props = $props();

	let collapsedGroups = $state<Set<string>>(new Set());
	let showInlineAdd = $state<string | null>(null);

	const statusConfig: Record<TaskStatus, { label: string; color: string; order: number }> = {
		todo: { label: 'To Do', color: '#6B7280', order: 0 },
		in_progress: { label: 'In Progress', color: '#3B82F6', order: 1 },
		in_review: { label: 'In Review', color: '#8B5CF6', order: 2 },
		done: { label: 'Done', color: '#10B981', order: 3 },
		blocked: { label: 'Blocked', color: '#F59E0B', order: 4 }
	};

	const priorityConfig: Record<Priority, { label: string; color: string; order: number }> = {
		critical: { label: 'Critical', color: '#EF4444', order: 0 },
		high: { label: 'High', color: '#F97316', order: 1 },
		medium: { label: 'Medium', color: '#EAB308', order: 2 },
		low: { label: 'Low', color: '#6B7280', order: 3 }
	};

	interface GroupedTasks {
		key: string;
		label: string;
		color: string;
		tasks: Task[];
		order: number;
	}

	const groupedTasks = $derived((): GroupedTasks[] => {
		let filtered = showDoneTasks ? tasks : tasks.filter(t => t.status !== 'done');

		if (groupBy === 'none') {
			return [{ key: 'all', label: 'All Tasks', color: '#6B7280', tasks: filtered, order: 0 }];
		}

		const groups: Map<string, GroupedTasks> = new Map();

		filtered.forEach(task => {
			let key: string;
			let label: string;
			let color: string;
			let order: number;

			switch (groupBy) {
				case 'status':
					key = task.status;
					label = statusConfig[task.status].label;
					color = statusConfig[task.status].color;
					order = statusConfig[task.status].order;
					break;
				case 'priority':
					key = task.priority;
					label = priorityConfig[task.priority].label;
					color = priorityConfig[task.priority].color;
					order = priorityConfig[task.priority].order;
					break;
				case 'project':
					key = task.projectId || 'no-project';
					label = task.projectName || 'No Project';
					color = task.projectColor || '#6B7280';
					order = task.projectName ? 0 : 999;
					break;
				case 'assignee':
					key = task.assignee?.id || 'unassigned';
					label = task.assignee?.name || 'Unassigned';
					color = '#6B7280';
					order = task.assignee ? 0 : 999;
					break;
				default:
					key = 'all';
					label = 'All Tasks';
					color = '#6B7280';
					order = 0;
			}

			if (!groups.has(key)) {
				groups.set(key, { key, label, color, tasks: [], order });
			}
			groups.get(key)!.tasks.push(task);
		});

		// Sort groups by order
		return Array.from(groups.values()).sort((a, b) => a.order - b.order);
	});

	function toggleGroup(key: string) {
		if (collapsedGroups.has(key)) {
			collapsedGroups.delete(key);
		} else {
			collapsedGroups.add(key);
		}
		collapsedGroups = new Set(collapsedGroups);
	}

	function handleAddTask(groupKey: string, task: { title: string; priority: Priority }) {
		onAddTask?.({ ...task, status: groupBy === 'status' ? groupKey : undefined });
		showInlineAdd = null;
	}
</script>

<div class="flex-1 overflow-y-auto">
	{#each groupedTasks() as group (group.key)}
		<div in:fade={{ duration: 150 }}>
			{#if groupBy !== 'none'}
				<TaskGroupHeader
					title={group.label}
					count={group.tasks.length}
					color={group.color}
					collapsed={collapsedGroups.has(group.key)}
					showAddButton={group.key !== 'done'}
					onToggle={() => toggleGroup(group.key)}
					onAdd={() => showInlineAdd = group.key}
				/>
			{/if}

			{#if !collapsedGroups.has(group.key)}
				<div transition:slide={{ duration: 200 }}>
					{#each group.tasks as task (task.id)}
						<TaskRow
							{...task}
							onClick={() => onTaskClick?.(task.id)}
							onStatusChange={(status) => onTaskStatusChange?.(task.id, status)}
							onEdit={() => onTaskEdit?.(task.id)}
							onDuplicate={() => onTaskDuplicate?.(task.id)}
							onDelete={() => onTaskDelete?.(task.id)}
						/>
					{/each}

					{#if showInlineAdd === group.key}
						<TaskInlineAdd
							status={groupBy === 'status' ? group.key : undefined}
							onAdd={(task) => handleAddTask(group.key, task)}
							onCancel={() => showInlineAdd = null}
						/>
					{/if}
				</div>
			{/if}
		</div>
	{/each}

	{#if tasks.length === 0}
		<div class="flex flex-col items-center justify-center py-16" in:fade={{ duration: 200 }}>
			{#if showInlineAdd !== 'all'}
				<div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
					</svg>
				</div>
				<h3 class="text-lg font-medium text-gray-900 mb-1">No tasks yet</h3>
				<p class="text-gray-500 mb-4">Create your first task to start tracking work</p>
				<button
					onclick={() => showInlineAdd = 'all'}
					class="btn-pill btn-pill-primary btn-pill-sm"
				>
					+ Create your first task
				</button>
			{:else}
				<div class="w-full max-w-2xl">
					<TaskInlineAdd
						onAdd={(task) => handleAddTask('all', task)}
						onCancel={() => showInlineAdd = null}
					/>
				</div>
			{/if}
		</div>
	{/if}
</div>
