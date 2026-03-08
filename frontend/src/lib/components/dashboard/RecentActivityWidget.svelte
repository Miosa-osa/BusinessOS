<script lang="ts">
	import { fly } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import WidgetWrapper from './WidgetWrapper.svelte';
	import WidgetHeader from './WidgetHeader.svelte';

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
{#snippet headerIcon()}
	<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
	</svg>
{/snippet}

{#snippet taskCompletedIcon()}
	<svg class="dw-feed-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
	</svg>
{/snippet}

{#snippet taskStartedIcon()}
	<svg class="dw-feed-icon" fill="currentColor" viewBox="0 0 24 24">
		<path d="M8 5v14l11-7z" />
	</svg>
{/snippet}

{#snippet projectCreatedIcon()}
	<svg class="dw-feed-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
	</svg>
{/snippet}

{#snippet projectUpdatedIcon()}
	<svg class="dw-feed-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
	</svg>
{/snippet}

{#snippet conversationIcon()}
	<svg class="dw-feed-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
	</svg>
{/snippet}

{#snippet teamIcon()}
	<svg class="dw-feed-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
	</svg>
{/snippet}

{#snippet artifactIcon()}
	<svg class="dw-feed-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
	</svg>
{/snippet}

<WidgetWrapper accent="cyan">
	<WidgetHeader
		title="Recent Activity"
		icon={headerIcon}
		iconColor="cyan"
		actionLabel={activities.length > 0 ? 'View All' : undefined}
		onAction={onViewAll}
	/>

	{#if activities.length === 0}
		<div class="dw-feed-empty">
			<div class="dw-feed-empty-icon">
				<svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
					/>
				</svg>
			</div>
			<p class="dw-feed-empty-title">No recent activity</p>
			<p class="dw-feed-empty-sub">Activity will appear here as you work</p>
		</div>
	{:else}
		<div class="dw-feed-list">
			{#each activities.slice(0, 10) as activity, index (activity.id)}
				<button
					onclick={() => handleActivityClick(activity)}
					class="dw-feed-item"
					aria-label="View {activity.type.replace('_', ' ')} activity"
					in:fly={{ x: -10, duration: 300, delay: index * 30 }}
				>
					<!-- Avatar or Icon -->
					{#if activity.actorAvatar}
						<img
							src={activity.actorAvatar}
							alt=""
							class="dw-feed-avatar"
						/>
					{:else if activity.actorName}
						<div class="dw-feed-avatar dw-feed-avatar--initials">
							<span>{activity.actorName.charAt(0)}</span>
						</div>
					{:else}
						<div class="dw-feed-avatar dw-feed-action--{activity.type.includes('completed') ? 'completed' : activity.type.includes('created') ? 'created' : activity.type.includes('updated') ? 'updated' : 'commented'}">
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
					<div class="dw-feed-content">
						<p class="dw-feed-text">
							{#if activity.actorName}
								<span class="dw-feed-actor">{activity.actorName}</span>
							{/if}
							{activity.description}
						</p>
					</div>

					<!-- Time -->
					<span class="dw-feed-time">
						{formatRelativeTime(activity.createdAt)}
					</span>
				</button>
			{/each}
		</div>
	{/if}
</WidgetWrapper>

<style>
	/* Foundation dw- Activity Feed patterns */
	.dw-feed-list {
		display: flex;
		flex-direction: column;
		max-height: 20rem;
		overflow-y: auto;
	}

	.dw-feed-item {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.5rem;
		border-radius: 0.5rem;
		transition: background 0.15s;
		text-align: left;
		background: none;
		border: none;
		cursor: pointer;
		width: 100%;
		font-family: inherit;
		color: inherit;
	}

	.dw-feed-item:hover {
		background: var(--dbg2, var(--bos-hover, rgba(0, 0, 0, 0.04)));
	}

	.dw-feed-item + .dw-feed-item {
		border-top: 1px solid var(--dbd2, var(--bos-border, rgba(0, 0, 0, 0.06)));
	}

	.dw-feed-avatar {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		font-weight: 700;
		flex-shrink: 0;
		object-fit: cover;
	}

	.dw-feed-avatar--initials {
		background: var(--dbg3, var(--bos-hover));
		color: var(--dt2, var(--bos-text-secondary));
	}

	.dw-feed-content {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	.dw-feed-text {
		font-size: 0.84rem;
		color: var(--dt, var(--bos-text-primary));
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.dw-feed-actor {
		font-weight: 600;
	}

	.dw-feed-time {
		font-size: 0.72rem;
		color: var(--dt4, var(--bos-text-tertiary));
		white-space: nowrap;
		flex-shrink: 0;
	}

	:global(.dw-feed-icon) {
		width: 1rem;
		height: 1rem;
	}

	/* Activity type colors — Foundation dw-feed-action patterns */
	.dw-feed-action--created {
		color: #22c55e;
		background: rgba(34, 197, 94, 0.1);
	}

	.dw-feed-action--updated {
		color: #3b82f6;
		background: rgba(59, 130, 246, 0.1);
	}

	.dw-feed-action--commented {
		color: #eab308;
		background: rgba(234, 179, 8, 0.1);
	}

	.dw-feed-action--completed {
		color: #a855f7;
		background: rgba(168, 85, 247, 0.1);
	}

	/* Empty state */
	.dw-feed-empty {
		text-align: center;
		padding: 2rem 0;
	}

	.dw-feed-empty-icon {
		width: 3.5rem;
		height: 3.5rem;
		background: linear-gradient(135deg, #06b6d4, #0891b2);
		border-radius: 0.75rem;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto 0.75rem;
		opacity: 0.3;
	}

	.dw-feed-empty-icon svg {
		width: 1.75rem;
		height: 1.75rem;
		color: #fff;
	}

	.dw-feed-empty-title {
		font-size: 0.875rem;
		color: var(--dt3, var(--bos-text-secondary));
	}

	.dw-feed-empty-sub {
		font-size: 0.75rem;
		color: var(--dt4, var(--bos-text-tertiary));
		margin-top: 0.25rem;
	}
</style>
