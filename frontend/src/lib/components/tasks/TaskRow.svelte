<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { goto } from '$app/navigation';
	import TaskCheckbox from './TaskCheckbox.svelte';
	import PriorityBadge from './PriorityBadge.svelte';

	type TaskStatus = 'todo' | 'in_progress' | 'in_review' | 'done' | 'blocked';
	type Priority = 'critical' | 'high' | 'medium' | 'low';

	interface Assignee {
		id: string;
		name: string;
		avatar?: string;
	}

	interface Props {
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
		onClick?: () => void;
		onStatusChange?: (status: TaskStatus) => void;
		onEdit?: () => void;
		onDuplicate?: () => void;
		onDelete?: () => void;
		onAssign?: () => void;
		onSetDueDate?: () => void;
	}

	let {
		id,
		title,
		status,
		priority,
		projectId,
		projectName,
		projectColor = '#6B7280',
		assignee,
		dueDate,
		tags = [],
		onClick,
		onStatusChange,
		onEdit,
		onDuplicate,
		onDelete,
		onAssign,
		onSetDueDate
	}: Props = $props();

	function navigateToProject(e: MouseEvent) {
		if (projectId) {
			e.stopPropagation();
			goto(`/projects/${projectId}`);
		}
	}

	let menuOpen = $state(false);

	function formatDueDate(dateStr: string) {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = date.getTime() - now.getTime();
		const days = Math.ceil(diff / (1000 * 60 * 60 * 24));

		if (days < 0) return { text: `${Math.abs(days)}d overdue`, isOverdue: true };
		if (days === 0) return { text: 'Due today', isOverdue: false };
		if (days === 1) return { text: 'Due tomorrow', isOverdue: false };
		if (days < 7) return { text: `Due in ${days}d`, isOverdue: false };
		return { text: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }), isOverdue: false };
	}

	const dueDateInfo = $derived(dueDate ? formatDueDate(dueDate) : null);
	const isDone = $derived(status === 'done');
	const isBlocked = $derived(status === 'blocked');
</script>

<div
	class="group relative flex items-center gap-3 px-4 py-3 border-b border-gray-100 transition-all duration-200 hover:bg-gray-50
		{isBlocked ? 'bg-orange-50/50' : ''}
		{isDone ? 'opacity-60' : ''}"
>
	<!-- Drag Handle (visible on hover) -->
	<div class="opacity-0 group-hover:opacity-100 absolute left-1 cursor-grab text-gray-300 hover:text-gray-400">
		<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
			<circle cx="9" cy="6" r="1.5" />
			<circle cx="15" cy="6" r="1.5" />
			<circle cx="9" cy="12" r="1.5" />
			<circle cx="15" cy="12" r="1.5" />
			<circle cx="9" cy="18" r="1.5" />
			<circle cx="15" cy="18" r="1.5" />
		</svg>
	</div>

	<!-- Checkbox -->
	<TaskCheckbox {status} onStatusChange={onStatusChange} />

	<!-- Main Content -->
	<button
		onclick={onClick}
		class="flex-1 min-w-0 text-left"
	>
		<div class="flex items-center gap-3">
			<span class="font-medium text-sm text-gray-900 truncate {isDone ? 'line-through text-gray-500' : ''}">
				{title}
			</span>
		</div>

		<div class="flex items-center gap-2 mt-1 text-xs text-gray-500">
			{#if projectName}
				<button
					onclick={navigateToProject}
					class="flex items-center gap-1 hover:text-gray-700 transition-colors {projectId ? 'hover:underline cursor-pointer' : ''}"
				>
					<span class="w-2 h-2 rounded-full" style="background-color: {projectColor}"></span>
					{projectName}
				</button>
				<span class="text-gray-300">•</span>
			{/if}

			{#if dueDateInfo}
				<span class="{dueDateInfo.isOverdue ? 'text-red-600 font-medium' : ''}">
					{dueDateInfo.text}
				</span>
				{#if tags.length > 0 || assignee}
					<span class="text-gray-300">•</span>
				{/if}
			{/if}

			{#each tags.slice(0, 2) as tag}
				<span class="px-1.5 py-0.5 bg-gray-100 rounded text-gray-600">{tag}</span>
			{/each}
			{#if tags.length > 2}
				<span class="text-gray-400">+{tags.length - 2}</span>
			{/if}
		</div>
	</button>

	<!-- Priority -->
	<PriorityBadge {priority} />

	<!-- Assignee -->
	{#if assignee}
		<div class="flex-shrink-0" title={assignee.name}>
			{#if assignee.avatar}
				<img src={assignee.avatar} alt={assignee.name} class="w-7 h-7 rounded-full" />
			{:else}
				<div class="w-7 h-7 rounded-full bg-gray-200 flex items-center justify-center text-xs font-medium text-gray-600">
					{assignee.name.charAt(0).toUpperCase()}
				</div>
			{/if}
		</div>
	{:else}
		<button
			onclick={onAssign}
			class="w-7 h-7 rounded-full border-2 border-dashed border-gray-200 flex items-center justify-center text-gray-400 hover:border-gray-300 hover:text-gray-500 transition-colors"
			title="Assign"
		>
			<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
			</svg>
		</button>
	{/if}

	<!-- Menu -->
	<div class="opacity-0 group-hover:opacity-100 transition-opacity {menuOpen ? 'opacity-100' : ''}">
		<DropdownMenu.Root bind:open={menuOpen}>
			<DropdownMenu.Trigger
				class="p-1.5 rounded-lg hover:bg-gray-200 text-gray-400 hover:text-gray-600 transition-colors"
				onclick={(e: MouseEvent) => e.stopPropagation()}
			>
				<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
					<path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
				</svg>
			</DropdownMenu.Trigger>
			<DropdownMenu.Portal>
				<DropdownMenu.Content
					class="z-50 min-w-[180px] bg-white border border-gray-200 rounded-xl shadow-lg p-1 animate-in fade-in-0 zoom-in-95"
					sideOffset={4}
				>
					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						onclick={onClick}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
						</svg>
						View Details
					</DropdownMenu.Item>
					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						onclick={onEdit}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
						</svg>
						Edit
					</DropdownMenu.Item>
					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						onclick={onDuplicate}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
						</svg>
						Duplicate
					</DropdownMenu.Item>
					{#if projectId && projectName}
						<DropdownMenu.Item
							class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
							onclick={() => goto(`/projects/${projectId}`)}
						>
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
							</svg>
							Go to Project
						</DropdownMenu.Item>
					{/if}

					<DropdownMenu.Separator class="h-px bg-gray-200 my-1" />

					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						onclick={onAssign}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
						</svg>
						Assign to...
					</DropdownMenu.Item>
					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						onclick={onSetDueDate}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						Set due date...
					</DropdownMenu.Item>

					<DropdownMenu.Separator class="h-px bg-gray-200 my-1" />

					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						onclick={() => onStatusChange?.('blocked')}
					>
						<svg class="w-4 h-4 text-orange-500" fill="currentColor" viewBox="0 0 24 24">
							<rect x="6" y="10" width="12" height="4" rx="1" />
						</svg>
						Mark Blocked
					</DropdownMenu.Item>
					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer transition-colors"
						onclick={() => onStatusChange?.('done')}
					>
						<svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						Mark Done
					</DropdownMenu.Item>

					<DropdownMenu.Separator class="h-px bg-gray-200 my-1" />

					<DropdownMenu.Item
						class="flex items-center gap-3 px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg cursor-pointer transition-colors"
						onclick={onDelete}
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
						</svg>
						Delete
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Portal>
		</DropdownMenu.Root>
	</div>
</div>
