<script lang="ts">
	import { fly, scale } from 'svelte/transition';
	import { goto } from '$app/navigation';

	type TaskPriority = 'critical' | 'high' | 'medium' | 'low';

	interface DashboardTask {
		id: string;
		title: string;
		projectName?: string;
		dueDate?: string;
		priority: TaskPriority;
		completed: boolean;
	}

	interface Props {
		tasks?: DashboardTask[];
		onToggle?: (id: string) => void;
		onViewAll?: () => void;
	}

	let { tasks = [], onToggle, onViewAll }: Props = $props();

	// Categorize tasks
	const categorizedTasks = $derived(() => {
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const nextWeek = new Date(today);
		nextWeek.setDate(nextWeek.getDate() + 7);

		const dueToday: DashboardTask[] = [];
		const upcoming: DashboardTask[] = [];
		const overdue: DashboardTask[] = [];

		for (const task of tasks.filter((t) => !t.completed)) {
			if (!task.dueDate) {
				upcoming.push(task);
				continue;
			}
			const due = new Date(task.dueDate);
			due.setHours(0, 0, 0, 0);

			if (due < today) {
				overdue.push(task);
			} else if (due.getTime() === today.getTime()) {
				dueToday.push(task);
			} else if (due <= nextWeek) {
				upcoming.push(task);
			}
		}

		return { dueToday, upcoming, overdue };
	});

	const priorityColors: Record<TaskPriority, string> = {
		critical: 'text-red-600 bg-red-50',
		high: 'text-orange-600 bg-orange-50',
		medium: 'text-yellow-600 bg-yellow-50',
		low: 'text-gray-600 bg-gray-50'
	};

	function formatDueDate(dueDate?: string): string {
		if (!dueDate) return '';
		const due = new Date(dueDate);
		const today = new Date();
		today.setHours(0, 0, 0, 0);
		const diff = Math.ceil((due.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

		if (diff < 0) return `${Math.abs(diff)}d overdue`;
		if (diff === 0) return 'Today';
		if (diff === 1) return 'Tomorrow';
		return due.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function handleToggle(id: string) {
		onToggle?.(id);
	}
</script>

<div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm hover:shadow-md transition-shadow duration-300">
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-green-500 to-green-600 flex items-center justify-center shadow-sm">
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
				</svg>
			</div>
			<h2 class="text-base font-semibold text-gray-900">My Tasks</h2>
		</div>
		{#if tasks.length > 0}
			<button
				onclick={() => onViewAll?.()}
				class="text-xs text-gray-500 hover:text-gray-700 transition-colors flex items-center gap-1 px-2 py-1 rounded-md hover:bg-gray-100"
			>
				View All
				<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
			</button>
		{/if}
	</div>

	{#if tasks.length === 0}
		<div class="text-center py-8">
			<div class="w-14 h-14 bg-gradient-to-br from-green-100 to-green-50 rounded-xl flex items-center justify-center mx-auto mb-3 shadow-sm">
				<svg class="w-7 h-7 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 13l4 4L19 7" />
				</svg>
			</div>
			<p class="text-sm font-medium text-gray-900 mb-1">All caught up!</p>
			<p class="text-xs text-gray-500 mb-3">No tasks due soon.</p>
			<button
				onclick={() => goto('/tasks')}
				class="inline-flex items-center gap-1.5 text-sm text-gray-600 hover:text-gray-900 font-medium px-3 py-1.5 rounded-lg hover:bg-gray-50 transition-colors"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Add a task
			</button>
		</div>
	{:else}
		<div class="space-y-4 max-h-80 overflow-y-auto">
			<!-- Overdue -->
			{#if categorizedTasks().overdue.length > 0}
				<div>
					<h3 class="text-xs font-medium text-red-600 uppercase tracking-wide mb-2">
						Overdue ({categorizedTasks().overdue.length})
					</h3>
					<div class="space-y-1">
						{#each categorizedTasks().overdue.slice(0, 3) as task (task.id)}
							<div
								class="flex items-center gap-3 p-2 rounded-lg bg-red-50 border border-red-100"
								in:fly={{ x: -10, duration: 200 }}
							>
								<button
									onclick={() => handleToggle(task.id)}
									class="flex-shrink-0 w-4 h-4 rounded border-2 border-red-300 hover:border-red-400 transition-colors"
								></button>
								<div class="flex-1 min-w-0">
									<p class="text-sm text-gray-900 truncate">{task.title}</p>
									{#if task.projectName}
										<p class="text-xs text-gray-500">{task.projectName}</p>
									{/if}
								</div>
								<span class="text-xs text-red-600 whitespace-nowrap flex items-center gap-1">
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
									</svg>
									{formatDueDate(task.dueDate)}
								</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Due Today -->
			{#if categorizedTasks().dueToday.length > 0}
				<div>
					<h3 class="text-xs font-medium text-gray-600 uppercase tracking-wide mb-2">
						Due Today ({categorizedTasks().dueToday.length})
					</h3>
					<div class="space-y-1">
						{#each categorizedTasks().dueToday.slice(0, 3) as task (task.id)}
							<div
								class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 transition-colors"
								in:fly={{ x: -10, duration: 200 }}
							>
								<button
									onclick={() => handleToggle(task.id)}
									class="flex-shrink-0 w-4 h-4 rounded border-2 border-gray-300 hover:border-gray-400 transition-colors"
								></button>
								<div class="flex-1 min-w-0">
									<p class="text-sm text-gray-900 truncate">{task.title}</p>
									{#if task.projectName}
										<p class="text-xs text-gray-500">{task.projectName}</p>
									{/if}
								</div>
								<span
									class="text-xs px-2 py-0.5 rounded {priorityColors[task.priority]} capitalize"
								>
									{task.priority}
								</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Upcoming -->
			{#if categorizedTasks().upcoming.length > 0}
				<div>
					<h3 class="text-xs font-medium text-gray-600 uppercase tracking-wide mb-2">
						Upcoming ({categorizedTasks().upcoming.length})
					</h3>
					<div class="space-y-1">
						{#each categorizedTasks().upcoming.slice(0, 3) as task (task.id)}
							<div
								class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 transition-colors"
								in:fly={{ x: -10, duration: 200 }}
							>
								<button
									onclick={() => handleToggle(task.id)}
									class="flex-shrink-0 w-4 h-4 rounded border-2 border-gray-300 hover:border-gray-400 transition-colors"
								></button>
								<div class="flex-1 min-w-0">
									<p class="text-sm text-gray-900 truncate">{task.title}</p>
									{#if task.projectName}
										<p class="text-xs text-gray-500">{task.projectName}</p>
									{/if}
								</div>
								<span class="text-xs text-gray-500 whitespace-nowrap">
									{formatDueDate(task.dueDate)}
								</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
