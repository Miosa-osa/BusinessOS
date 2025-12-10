<script lang="ts">
	import { fade, fly } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	import TaskBoardCard from './TaskBoardCard.svelte';

	type TaskStatus = 'todo' | 'in_progress' | 'in_review' | 'done' | 'blocked';
	type Priority = 'critical' | 'high' | 'medium' | 'low';

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
		subtaskCount?: number;
		subtaskCompleted?: number;
		commentCount?: number;
	}

	interface Props {
		tasks: Task[];
		onTaskClick?: (taskId: string) => void;
		onTaskStatusChange?: (taskId: string, status: TaskStatus) => void;
		onAddTask?: (status: TaskStatus) => void;
	}

	let { tasks, onTaskClick, onTaskStatusChange, onAddTask }: Props = $props();

	const columns: { status: TaskStatus; label: string; color: string }[] = [
		{ status: 'todo', label: 'To Do', color: '#6B7280' },
		{ status: 'in_progress', label: 'In Progress', color: '#3B82F6' },
		{ status: 'in_review', label: 'In Review', color: '#8B5CF6' },
		{ status: 'done', label: 'Done', color: '#10B981' },
		{ status: 'blocked', label: 'Blocked', color: '#F59E0B' }
	];

	let draggedTask = $state<Task | null>(null);
	let dragOverColumn = $state<TaskStatus | null>(null);

	const tasksByStatus = $derived((): Record<TaskStatus, Task[]> => {
		const grouped: Record<TaskStatus, Task[]> = {
			todo: [],
			in_progress: [],
			in_review: [],
			done: [],
			blocked: []
		};

		tasks.forEach(task => {
			grouped[task.status].push(task);
		});

		return grouped;
	});

	function handleDragStart(e: DragEvent, task: Task) {
		draggedTask = task;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
			e.dataTransfer.setData('text/plain', task.id);
		}
	}

	function handleDragEnd() {
		draggedTask = null;
		dragOverColumn = null;
	}

	function handleDragOver(e: DragEvent, status: TaskStatus) {
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}
		dragOverColumn = status;
	}

	function handleDragLeave() {
		dragOverColumn = null;
	}

	function handleDrop(e: DragEvent, status: TaskStatus) {
		e.preventDefault();
		if (draggedTask && draggedTask.status !== status) {
			onTaskStatusChange?.(draggedTask.id, status);
		}
		draggedTask = null;
		dragOverColumn = null;
	}
</script>

<div class="flex-1 overflow-x-auto p-6">
	<div class="flex gap-4 min-w-max">
		{#each columns as column (column.status)}
			<div
				class="w-72 flex-shrink-0 flex flex-col bg-gray-50 rounded-xl"
				ondragover={(e) => handleDragOver(e, column.status)}
				ondragleave={handleDragLeave}
				ondrop={(e) => handleDrop(e, column.status)}
			>
				<!-- Column Header -->
				<div class="flex items-center justify-between px-3 py-3">
					<div class="flex items-center gap-2">
						<span class="w-2.5 h-2.5 rounded-full" style="background-color: {column.color}"></span>
						<span class="font-medium text-sm text-gray-700">{column.label}</span>
						<span class="text-xs text-gray-400 bg-gray-200 px-1.5 py-0.5 rounded-full">
							{tasksByStatus()[column.status].length}
						</span>
					</div>
					{#if column.status !== 'done'}
						<button
							onclick={() => onAddTask?.(column.status)}
							class="w-6 h-6 flex items-center justify-center text-gray-400 hover:text-gray-600 hover:bg-gray-200 rounded transition-colors"
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
						</button>
					{/if}
				</div>

				<!-- Column Content -->
				<div
					class="flex-1 px-2 pb-2 space-y-2 min-h-[200px] rounded-lg transition-colors
						{dragOverColumn === column.status ? 'bg-gray-100 ring-2 ring-inset ring-gray-300' : ''}"
				>
					{#each tasksByStatus()[column.status] as task (task.id)}
						<div
							draggable="true"
							ondragstart={(e) => handleDragStart(e, task)}
							ondragend={handleDragEnd}
							class="cursor-grab active:cursor-grabbing {draggedTask?.id === task.id ? 'opacity-50' : ''}"
							animate:flip={{ duration: 200 }}
						>
							<TaskBoardCard
								{...task}
								onClick={() => onTaskClick?.(task.id)}
							/>
						</div>
					{/each}

					{#if tasksByStatus()[column.status].length === 0 && !dragOverColumn}
						<div class="flex items-center justify-center h-24 text-sm text-gray-400" in:fade={{ duration: 150 }}>
							No tasks
						</div>
					{/if}
				</div>
			</div>
		{/each}
	</div>
</div>
