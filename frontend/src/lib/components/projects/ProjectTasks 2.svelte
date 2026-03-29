<script lang="ts">
	import type { Task, TeamMemberListResponse } from '$lib/api';
	import { api } from '$lib/api';
	import { getPriorityColor, formatDate } from '$lib/utils/project';

	interface Props {
		tasks: Task[];
		teamMembers: TeamMemberListResponse[];
		embedSuffix: string;
		onShowAddTask: () => void;
		onEditTask: (task: Task) => void;
		onTasksChanged: () => Promise<void>;
	}

	let { tasks, teamMembers, embedSuffix, onShowAddTask, onEditTask, onTasksChanged }: Props = $props();

	let completedTasks = $derived(tasks.filter((t) => t.status === 'done').length);
	let totalTasks = $derived(tasks.length);

	// Drag & Drop state
	let draggedTask = $state<Task | null>(null);
	let dragOverTask = $state<Task | null>(null);

	async function handleToggleTask(taskId: string) {
		try {
			await api.toggleTask(taskId);
			await onTasksChanged();
		} catch (err) {
			console.error('Error toggling task:', err);
		}
	}

	async function handleDeleteTask(taskId: string) {
		try {
			await api.deleteTask(taskId);
			await onTasksChanged();
		} catch (err) {
			console.error('Error deleting task:', err);
		}
	}

	function handleDragStart(e: DragEvent, task: Task) {
		draggedTask = task;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
		}
	}

	function handleDragOver(e: DragEvent, task: Task) {
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}
		if (draggedTask && draggedTask.id !== task.id && draggedTask.status === task.status) {
			dragOverTask = task;
		}
	}

	function handleDragLeave() {
		dragOverTask = null;
	}

	async function handleDrop(e: DragEvent, targetTask: Task) {
		e.preventDefault();
		dragOverTask = null;

		const currentDraggedTask = draggedTask;
		if (
			!currentDraggedTask ||
			currentDraggedTask.id === targetTask.id ||
			currentDraggedTask.status !== targetTask.status
		) {
			return;
		}

		try {
			const statusTasks = tasks.filter((t) => t.status === targetTask.status);
			const draggedIndex = statusTasks.findIndex((t) => t.id === currentDraggedTask.id);
			const targetIndex = statusTasks.findIndex((t) => t.id === targetTask.id);

			if (draggedIndex === -1 || targetIndex === -1) return;

			const newStatusTasks = [...statusTasks];
			const [removed] = newStatusTasks.splice(draggedIndex, 1);
			newStatusTasks.splice(targetIndex, 0, removed);

			const updatePromises = newStatusTasks.map((task, index) =>
				api.updateTask(task.id, { position: index })
			);

			await Promise.all(updatePromises);
			await onTasksChanged();
		} catch (err) {
			console.error('Error reordering tasks:', err);
			await onTasksChanged();
		}
	}

	function handleDragEnd() {
		draggedTask = null;
		dragOverTask = null;
	}
</script>

<div class="space-y-4">
	<!-- Task Stats Cards -->
	<div class="grid grid-cols-4 gap-3">
		<div class="bg-white rounded-xl border border-gray-200 p-4">
			<p class="text-xs font-medium text-gray-500 uppercase tracking-wider">To Do</p>
			<p class="text-2xl font-bold text-gray-900 mt-1">{tasks.filter((t) => t.status === 'todo').length}</p>
		</div>
		<div class="bg-white rounded-xl border border-gray-200 p-4">
			<p class="text-xs font-medium text-gray-500 uppercase tracking-wider">In Progress</p>
			<p class="text-2xl font-bold text-blue-600 mt-1">{tasks.filter((t) => t.status === 'in_progress').length}</p>
		</div>
		<div class="bg-white rounded-xl border border-gray-200 p-4">
			<p class="text-xs font-medium text-gray-500 uppercase tracking-wider">Done</p>
			<p class="text-2xl font-bold text-green-600 mt-1">{tasks.filter((t) => t.status === 'done').length}</p>
		</div>
		<div class="bg-white rounded-xl border border-gray-200 p-4">
			<p class="text-xs font-medium text-gray-500 uppercase tracking-wider">Completion</p>
			<p class="text-2xl font-bold text-purple-600 mt-1">
				{totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0}%
			</p>
		</div>
	</div>

	<!-- Tasks List -->
	<div class="bg-white rounded-xl border border-gray-200">
		<div class="p-4 border-b border-gray-100 flex items-center justify-between">
			<div class="flex items-center gap-2">
				<h2 class="text-lg font-medium text-gray-900">All Tasks</h2>
				<span class="text-sm text-gray-400">({totalTasks})</span>
			</div>
			<div class="flex items-center gap-2">
				<a href="/tasks{embedSuffix}" class="btn btn-secondary text-sm">
					<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
					</svg>
					Open Tasks
				</a>
				<button onclick={onShowAddTask} class="btn btn-primary text-sm">
					<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Add Task
				</button>
			</div>
		</div>

		{#if tasks.length === 0}
			<div class="text-center py-16">
				<div class="w-16 h-16 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-4">
					<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
				</div>
				<h3 class="text-lg font-medium text-gray-900 mb-1">No tasks yet</h3>
				<p class="text-gray-500 mb-4">Break down your project into manageable tasks</p>
				<button onclick={onShowAddTask} class="btn btn-primary">Add First Task</button>
			</div>
		{:else}
			{#each [
				{ status: 'todo', label: 'To Do', color: 'gray', items: tasks.filter((t) => t.status === 'todo') },
				{ status: 'in_progress', label: 'In Progress', color: 'blue', items: tasks.filter((t) => t.status === 'in_progress') },
				{ status: 'done', label: 'Done', color: 'green', items: tasks.filter((t) => t.status === 'done') }
			].filter((g) => g.items.length > 0) as group}
				<div class="border-b border-gray-100 last:border-b-0">
					<div class="px-4 py-2 bg-gray-50/50 flex items-center gap-2">
						<span class="w-2 h-2 rounded-full bg-{group.color}-500"></span>
						<span class="text-xs font-medium text-gray-600">{group.label}</span>
						<span class="text-xs text-gray-400">({group.items.length})</span>
					</div>
					<div class="divide-y divide-gray-100">
						{#each group.items as task}
							{@const isSubtask = !!task.parent_task_id}
							{@const assignee = task.assignee_id ? teamMembers.find((m) => m.id === task.assignee_id) : null}
							<div
								class="flex items-center gap-4 p-4 hover:bg-gray-50 group/task transition-all {dragOverTask?.id === task.id ? 'border-t-2 border-purple-600' : ''} {draggedTask?.id === task.id ? 'opacity-50' : 'opacity-100'} {isSubtask ? 'pl-12 bg-gray-50/50' : ''}"
								draggable="true"
								ondragstart={(e) => handleDragStart(e, task)}
								ondragover={(e) => handleDragOver(e, task)}
								ondragleave={handleDragLeave}
								ondrop={(e) => handleDrop(e, task)}
								ondragend={handleDragEnd}
							>
								<!-- Subtask Indicator -->
								{#if isSubtask}
									<div class="absolute left-6 w-4 h-4 flex items-center justify-center">
										<svg class="w-3 h-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
										</svg>
									</div>
								{/if}

								<!-- Drag Handle -->
								<div class="flex-shrink-0 cursor-move text-gray-400 hover:text-gray-600 opacity-0 group-hover/task:opacity-100 transition-opacity">
									<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8h16M4 16h16" />
									</svg>
								</div>

								<button
									onclick={() => handleToggleTask(task.id)}
									class="w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-colors {task.status === 'done' ? 'bg-green-600 border-green-600 text-white' : 'border-gray-300 hover:border-purple-600'}"
									aria-label="Toggle task complete"
								>
									{#if task.status === 'done'}
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
										</svg>
									{/if}
								</button>

								<div class="flex-1 min-w-0">
									<p class="font-medium {task.status === 'done' ? 'text-gray-400 line-through' : 'text-gray-900'}">{task.title}</p>
									{#if task.description}
										<p class="text-sm text-gray-500 mt-0.5 line-clamp-1">{task.description}</p>
									{/if}
									<div class="flex items-center gap-3 mt-1 text-xs text-gray-400">
										{#if task.start_date}
											<span class="flex items-center gap-1">
												<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
												</svg>
												Start: {formatDate(task.start_date)}
											</span>
										{/if}
										{#if task.due_date}
											<span class="flex items-center gap-1 {new Date(task.due_date) < new Date() && task.status !== 'done' ? 'text-red-500' : ''}">
												<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
												</svg>
												Due: {formatDate(task.due_date)}
											</span>
										{/if}
										{#if task.estimated_hours}
											<span class="flex items-center gap-1">
												<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
												</svg>
												{task.estimated_hours}h
											</span>
										{/if}
									</div>
								</div>

								<div class="flex items-center gap-2">
									<span class="text-xs px-2 py-1 rounded font-medium {getPriorityColor(task.priority)}">{task.priority}</span>
									{#if assignee}
										<span class="text-xs px-2 py-1 rounded bg-blue-100 text-blue-700 font-medium flex items-center gap-1">
											<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
											</svg>
											{assignee.name}
										</span>
									{/if}
								</div>

								<div class="flex items-center gap-1 opacity-0 group-hover/task:opacity-100 transition-opacity">
									<button
										onclick={() => onEditTask(task)}
										class="p-2 text-gray-400 hover:text-blue-600 rounded-lg hover:bg-blue-50 transition-colors"
										aria-label="Edit task"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
										</svg>
									</button>
									<button
										onclick={() => handleDeleteTask(task.id)}
										class="p-2 text-gray-400 hover:text-red-600 rounded-lg hover:bg-red-50 transition-colors"
										aria-label="Delete task"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
										</svg>
									</button>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/each}
		{/if}
	</div>
</div>
