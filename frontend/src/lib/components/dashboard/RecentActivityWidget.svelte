<script lang="ts">
	import { fly } from 'svelte/transition';
	import { goto } from '$app/navigation';

	type ActivityType =
		| 'task_completed'
		| 'task_started'
		| 'project_created'
		| 'project_updated'
		| 'conversation'
		| 'team'
		| 'artifact';

	interface Activity {
		id: string;
		type: ActivityType;
		description: string;
		actorName?: string;
		actorAvatar?: string;
		targetId?: string;
		targetType?: string;
		createdAt: string;
	}

	interface Props {
		activities?: Activity[];
		onViewAll?: () => void;
	}

	let { activities = [], onViewAll }: Props = $props();

	const typeColors: Record<ActivityType, string> = {
		task_completed: 'bg-green-100 text-green-600',
		task_started: 'bg-blue-100 text-blue-600',
		project_created: 'bg-purple-100 text-purple-600',
		project_updated: 'bg-indigo-100 text-indigo-600',
		conversation: 'bg-cyan-100 text-cyan-600',
		team: 'bg-orange-100 text-orange-600',
		artifact: 'bg-gray-100 text-gray-600'
	};

	function formatRelativeTime(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(minutes / 60);
		const days = Math.floor(hours / 24);

		if (minutes < 1) return 'Just now';
		if (minutes < 60) return `${minutes} min ago`;
		if (hours < 24) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
		if (days === 1) return 'Yesterday';
		if (days < 7) return `${days} days ago`;
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function handleActivityClick(activity: Activity) {
		if (activity.targetId && activity.targetType) {
			switch (activity.targetType) {
				case 'project':
					goto(`/projects/${activity.targetId}`);
					break;
				case 'task':
					goto(`/tasks?id=${activity.targetId}`);
					break;
				case 'conversation':
					goto(`/chat?id=${activity.targetId}`);
					break;
			}
		}
	}
</script>

<!-- SVG Icon snippets for activity types -->
{#snippet taskCompletedIcon()}
	<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
	</svg>
{/snippet}

{#snippet taskStartedIcon()}
	<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
		<path d="M8 5v14l11-7z" />
	</svg>
{/snippet}

{#snippet projectCreatedIcon()}
	<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
	</svg>
{/snippet}

{#snippet projectUpdatedIcon()}
	<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
	</svg>
{/snippet}

{#snippet conversationIcon()}
	<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
	</svg>
{/snippet}

{#snippet teamIcon()}
	<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
	</svg>
{/snippet}

{#snippet artifactIcon()}
	<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
	</svg>
{/snippet}

<div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm hover:shadow-md transition-shadow duration-300">
	<div class="flex items-center justify-between mb-4">
		<div class="flex items-center gap-2">
			<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-cyan-500 to-cyan-600 flex items-center justify-center shadow-sm">
				<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</div>
			<h2 class="text-base font-semibold text-gray-900">Recent Activity</h2>
		</div>
		{#if activities.length > 0}
			<button
				onclick={() => onViewAll?.()}
				class="btn-pill-sm text-xs"
			>
				View All
				<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
				</svg>
			</button>
		{/if}
	</div>

	{#if activities.length === 0}
		<div class="text-center py-8">
			<div class="w-14 h-14 bg-gradient-to-br from-cyan-100 to-cyan-50 rounded-xl flex items-center justify-center mx-auto mb-3 shadow-sm">
				<svg class="w-7 h-7 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
					/>
				</svg>
			</div>
			<p class="text-sm text-gray-500">No recent activity</p>
			<p class="text-xs text-gray-400 mt-1">Activity will appear here as you work</p>
		</div>
	{:else}
		<div class="space-y-1 max-h-80 overflow-y-auto">
			{#each activities.slice(0, 10) as activity, index (activity.id)}
				<button
					onclick={() => handleActivityClick(activity)}
					class="btn-pill w-full flex items-start gap-3 text-left"
					in:fly={{ x: -10, duration: 300, delay: index * 30 }}
				>
					<!-- Avatar or Icon -->
					{#if activity.actorAvatar}
						<img
							src={activity.actorAvatar}
							alt=""
							class="w-8 h-8 rounded-full flex-shrink-0"
						/>
					{:else if activity.actorName}
						<div
							class="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center flex-shrink-0"
						>
							<span class="text-sm font-medium text-gray-600">
								{activity.actorName.charAt(0)}
							</span>
						</div>
					{:else}
						<div
							class="w-8 h-8 rounded-full {typeColors[activity.type]} flex items-center justify-center flex-shrink-0"
						>
							{#if activity.type === 'task_completed'}
								{@render taskCompletedIcon()}
							{:else if activity.type === 'task_started'}
								{@render taskStartedIcon()}
							{:else if activity.type === 'project_created'}
								{@render projectCreatedIcon()}
							{:else if activity.type === 'project_updated'}
								{@render projectUpdatedIcon()}
							{:else if activity.type === 'conversation'}
								{@render conversationIcon()}
							{:else if activity.type === 'team'}
								{@render teamIcon()}
							{:else if activity.type === 'artifact'}
								{@render artifactIcon()}
							{/if}
						</div>
					{/if}

					<!-- Content -->
					<div class="flex-1 min-w-0">
						<p class="text-sm text-gray-700 line-clamp-2">
							{#if activity.actorName}
								<span class="font-medium">{activity.actorName}</span>
							{/if}
							{activity.description}
						</p>
					</div>

					<!-- Time -->
					<span class="text-xs text-gray-400 whitespace-nowrap flex-shrink-0">
						{formatRelativeTime(activity.createdAt)}
					</span>
				</button>
			{/each}
		</div>
	{/if}
</div>
