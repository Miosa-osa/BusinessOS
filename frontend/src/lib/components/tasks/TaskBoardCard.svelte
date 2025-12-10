<script lang="ts">
	import PriorityBadge from './PriorityBadge.svelte';

	type Priority = 'critical' | 'high' | 'medium' | 'low';

	interface Assignee {
		id: string;
		name: string;
		avatar?: string;
	}

	interface Props {
		id: string;
		title: string;
		priority: Priority;
		projectName?: string;
		projectColor?: string;
		assignee?: Assignee;
		dueDate?: string;
		subtaskCount?: number;
		subtaskCompleted?: number;
		commentCount?: number;
		onClick?: () => void;
	}

	let {
		id,
		title,
		priority,
		projectName,
		projectColor = '#6B7280',
		assignee,
		dueDate,
		subtaskCount = 0,
		subtaskCompleted = 0,
		commentCount = 0,
		onClick
	}: Props = $props();

	function formatDueDate(dateStr: string) {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = date.getTime() - now.getTime();
		const days = Math.ceil(diff / (1000 * 60 * 60 * 24));

		if (days < 0) return { text: `${Math.abs(days)}d overdue`, isOverdue: true };
		if (days === 0) return { text: 'Today', isOverdue: false };
		if (days === 1) return { text: 'Tomorrow', isOverdue: false };
		return { text: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }), isOverdue: false };
	}

	const dueDateInfo = $derived(dueDate ? formatDueDate(dueDate) : null);
</script>

<button
	onclick={onClick}
	class="w-full text-left bg-white border border-gray-200 rounded-xl p-3 hover:shadow-md hover:border-gray-300 transition-all duration-200 cursor-pointer group"
>
	<!-- Title -->
	<h4 class="font-medium text-sm text-gray-900 mb-2 line-clamp-2 group-hover:text-gray-700">
		{title}
	</h4>

	<!-- Project -->
	{#if projectName}
		<div class="flex items-center gap-1.5 mb-2 text-xs text-gray-500">
			<span class="w-2 h-2 rounded-full flex-shrink-0" style="background-color: {projectColor}"></span>
			<span class="truncate">{projectName}</span>
		</div>
	{/if}

	<!-- Priority & Due Date -->
	<div class="flex items-center justify-between mb-3">
		<PriorityBadge {priority} size="sm" />
		{#if dueDateInfo}
			<span class="text-xs {dueDateInfo.isOverdue ? 'text-red-600 font-medium' : 'text-gray-500'}">
				{dueDateInfo.text}
			</span>
		{/if}
	</div>

	<!-- Footer -->
	<div class="flex items-center justify-between pt-2 border-t border-gray-100">
		<!-- Subtasks & Comments -->
		<div class="flex items-center gap-3 text-xs text-gray-400">
			{#if subtaskCount > 0}
				<span class="flex items-center gap-1">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
					</svg>
					{subtaskCompleted}/{subtaskCount}
				</span>
			{/if}
			{#if commentCount > 0}
				<span class="flex items-center gap-1">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
					</svg>
					{commentCount}
				</span>
			{/if}
		</div>

		<!-- Assignee -->
		{#if assignee}
			{#if assignee.avatar}
				<img src={assignee.avatar} alt={assignee.name} class="w-6 h-6 rounded-full" title={assignee.name} />
			{:else}
				<div class="w-6 h-6 rounded-full bg-gray-200 flex items-center justify-center text-xs font-medium text-gray-600" title={assignee.name}>
					{assignee.name.charAt(0).toUpperCase()}
				</div>
			{/if}
		{/if}
	</div>
</button>
