<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { team } from '$lib/stores/team';
	import { goto } from '$app/navigation';
	import type { TeamMemberDetailResponse } from '$lib/api';
	import StatusBadge from '$lib/components/team/StatusBadge.svelte';
	import CapacityBar from '$lib/components/team/CapacityBar.svelte';
	import { ArrowLeft, Mail, Calendar, Briefcase, CheckSquare, Clock } from 'lucide-svelte';

	let member = $state<TeamMemberDetailResponse | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	const memberId = $derived($page.params.memberId);

	onMount(async () => {
		await loadMember();
	});

	async function loadMember() {
		loading = true;
		error = null;
		try {
			const id = memberId;
			if (!id) {
				error = 'Member not found';
				return;
			}
			const data = await team.loadMember(id);
			if (data) {
				member = data;
			} else {
				error = 'Member not found';
			}
		} catch {
			error = 'Failed to load member details';
		} finally {
			loading = false;
		}
	}

	function getInitials(name: string) {
		return name
			.split(' ')
			.map((n) => n.charAt(0))
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}

	function formatDate(dateStr: string | undefined) {
		if (!dateStr) return 'N/A';
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'long',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function capacityColor(pct: number): string {
		if (pct >= 90) return '#ef4444';
		if (pct >= 65) return '#f59e0b';
		return '#22c55e';
	}
</script>

<div class="td-profile">
	<!-- Header -->
	<div class="td-profile__header">
		<div class="td-profile__header-inner">
			<button
				onclick={() => goto('/team')}
				class="td-profile__back"
			>
				<ArrowLeft class="w-4 h-4" />
				Back to Team
			</button>

			{#if loading}
				<div class="td-profile__skeleton">
					<div class="td-profile__skeleton-avatar"></div>
					<div class="td-profile__skeleton-lines">
						<div class="td-profile__skeleton-line td-profile__skeleton-line--lg"></div>
						<div class="td-profile__skeleton-line td-profile__skeleton-line--sm"></div>
					</div>
				</div>
			{:else if error}
				<div class="td-profile__error">
					<p class="td-profile__error-text">{error}</p>
					<button onclick={loadMember} class="btn-pill btn-pill-primary btn-pill-sm">
						Try Again
					</button>
				</div>
			{:else if member}
				<div class="td-profile__identity">
					<!-- Avatar -->
					{#if member.avatar_url}
						<img
							src={member.avatar_url}
							alt={member.name}
							class="td-profile__avatar-img"
						/>
					{:else}
						<div class="td-profile__avatar-initials">
							{getInitials(member.name)}
						</div>
					{/if}

					<!-- Info -->
					<div class="td-profile__identity-info">
						<div class="td-profile__name-row">
							<h1 class="td-profile__name">{member.name}</h1>
							<StatusBadge status={member.status as 'available' | 'busy' | 'overloaded' | 'ooo'} />
						</div>
						<p class="td-profile__role">{member.role}</p>
						<div class="td-profile__meta-row">
							<span class="td-profile__meta-item">
								<Mail class="w-4 h-4" />
								{member.email}
							</span>
							<span class="td-profile__meta-item">
								<Calendar class="w-4 h-4" />
								Joined {formatDate(member.joined_at)}
							</span>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>

	{#if member && !loading}
		<div class="td-profile__body">
			<div class="td-profile__grid">
				<!-- Stats Cards -->
				<div class="td-profile__stats-col">
					<div class="td-profile__stats-row">
						<div class="td-profile__stat-card">
							<div class="td-profile__stat-header">
								<div class="td-profile__stat-icon td-profile__stat-icon--blue">
									<Briefcase class="w-5 h-5" />
								</div>
								<span class="td-profile__stat-label">Active Projects</span>
							</div>
							<p class="td-profile__stat-value">{member.active_projects}</p>
						</div>

						<div class="td-profile__stat-card">
							<div class="td-profile__stat-header">
								<div class="td-profile__stat-icon td-profile__stat-icon--green">
									<CheckSquare class="w-5 h-5" />
								</div>
								<span class="td-profile__stat-label">Open Tasks</span>
							</div>
							<p class="td-profile__stat-value">{member.open_tasks}</p>
						</div>
					</div>

					<!-- Capacity -->
					<div class="td-profile__card">
						<div class="td-profile__capacity-header">
							<span class="td-profile__stat-label">Current Capacity</span>
							<span class="td-profile__capacity-pct" style="color:{capacityColor(member.capacity)}">
								{member.capacity}%
							</span>
						</div>
						<CapacityBar capacity={member.capacity} size="lg" />
					</div>
				</div>

				<!-- Skills -->
				<div class="td-profile__card">
					<h3 class="td-profile__section-title">Skills</h3>
					{#if member.skills && member.skills.length > 0}
						<div class="td-profile__skills">
							{#each member.skills as skill}
								<span class="td-profile__skill-tag">{skill}</span>
							{/each}
						</div>
					{:else}
						<p class="td-profile__empty-text">No skills listed</p>
					{/if}
				</div>

				<!-- Recent Activity -->
				<div class="td-profile__card td-profile__card--full">
					<h3 class="td-profile__section-title">
						<Clock class="w-4 h-4" />
						Recent Activity
					</h3>
					{#if member.activities && member.activities.length > 0}
						<div class="td-profile__activity-list">
							{#each member.activities.slice(0, 5) as activity}
								<div class="td-profile__activity-item">
									<span class="td-profile__activity-dot"></span>
									<div class="td-profile__activity-body">
										<p class="td-profile__activity-text">{activity.description}</p>
										<p class="td-profile__activity-time">{formatDate(activity.created_at)}</p>
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="td-profile__empty-text">No recent activity</p>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	/* ─── Team Profile Page ─ Foundation Tokens ─────────────────── */
	.td-profile {
		min-height: 100%;
	}

	/* Header */
	.td-profile__header {
		background: var(--dbg, #fff);
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}
	.td-profile__header-inner {
		max-width: 64rem;
		margin: 0 auto;
		padding: 1rem 1.5rem;
	}
	.td-profile__back {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: var(--dt3, #888);
		margin-bottom: 1rem;
		transition: color 0.15s;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}
	.td-profile__back:hover {
		color: var(--dt, #111);
	}

	/* Skeleton */
	.td-profile__skeleton {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}
	.td-profile__skeleton-avatar {
		width: 5rem;
		height: 5rem;
		border-radius: 9999px;
		background: var(--dbg2, #f5f5f5);
		animation: pulse 1.5s ease-in-out infinite;
	}
	.td-profile__skeleton-lines {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.td-profile__skeleton-line {
		height: 1rem;
		border-radius: 0.25rem;
		background: var(--dbg2, #f5f5f5);
		animation: pulse 1.5s ease-in-out infinite;
	}
	.td-profile__skeleton-line--lg { width: 12rem; height: 1.5rem; }
	.td-profile__skeleton-line--sm { width: 8rem; }

	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.5; }
	}

	/* Error */
	.td-profile__error {
		text-align: center;
		padding: 3rem 0;
	}
	.td-profile__error-text {
		color: #ef4444;
		margin-bottom: 1rem;
	}

	/* Identity */
	.td-profile__identity {
		display: flex;
		align-items: flex-start;
		gap: 1.5rem;
	}
	.td-profile__avatar-img {
		width: 5rem;
		height: 5rem;
		border-radius: 9999px;
		object-fit: cover;
		border: 3px solid var(--dbd, #e0e0e0);
		box-shadow: 0 4px 12px rgba(0,0,0,0.08);
	}
	.td-profile__avatar-initials {
		width: 5rem;
		height: 5rem;
		border-radius: 9999px;
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		display: flex;
		align-items: center;
		justify-content: center;
		color: #fff;
		font-size: 1.5rem;
		font-weight: 800;
		border: 3px solid var(--dbd, #e0e0e0);
		box-shadow: 0 4px 12px rgba(0,0,0,0.08);
		letter-spacing: -0.02em;
	}
	.td-profile__identity-info {
		flex: 1;
	}
	.td-profile__name-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.25rem;
	}
	.td-profile__name {
		font-size: 1.5rem;
		font-weight: 800;
		color: var(--dt, #111);
		letter-spacing: -0.03em;
		line-height: 1.2;
	}
	.td-profile__role {
		font-size: 1rem;
		color: var(--dt3, #888);
		font-weight: 500;
	}
	.td-profile__meta-row {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-top: 0.75rem;
	}
	.td-profile__meta-item {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.8125rem;
		color: var(--dt3, #888);
	}

	/* Body */
	.td-profile__body {
		max-width: 64rem;
		margin: 0 auto;
		padding: 2rem 1.5rem;
	}
	.td-profile__grid {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 1rem;
	}
	@media (max-width: 768px) {
		.td-profile__grid { grid-template-columns: 1fr; }
	}

	/* Stats row */
	.td-profile__stats-col {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.td-profile__stats-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	/* Cards */
	.td-profile__card {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 0.875rem;
		padding: 1.25rem;
	}
	.td-profile__card--full {
		grid-column: 1 / -1;
	}
	.td-profile__stat-card {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 0.875rem;
		padding: 1.25rem;
	}
	.td-profile__stat-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.75rem;
	}
	.td-profile__stat-icon {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.td-profile__stat-icon--blue {
		background: color-mix(in srgb, #3b82f6 12%, transparent);
		color: #3b82f6;
	}
	.td-profile__stat-icon--green {
		background: color-mix(in srgb, #22c55e 12%, transparent);
		color: #22c55e;
	}
	.td-profile__stat-label {
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--dt3, #888);
	}
	.td-profile__stat-value {
		font-size: 2rem;
		font-weight: 800;
		color: var(--dt, #111);
		letter-spacing: -0.03em;
		line-height: 1;
	}

	/* Capacity */
	.td-profile__capacity-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}
	.td-profile__capacity-pct {
		font-size: 0.875rem;
		font-weight: 700;
	}

	/* Section titles */
	.td-profile__section-title {
		font-size: 0.8125rem;
		font-weight: 700;
		color: var(--dt, #111);
		margin-bottom: 1rem;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	/* Skills */
	.td-profile__skills {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
	}
	.td-profile__skill-tag {
		display: inline-flex;
		align-items: center;
		height: 1.625rem;
		padding: 0 0.75rem;
		border-radius: 9999px;
		border: 1px solid var(--dbd, #e0e0e0);
		background: var(--dbg2, #f5f5f5);
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--dt2, #555);
	}
	.td-profile__empty-text {
		font-size: 0.8125rem;
		color: var(--dt4, #bbb);
	}

	/* Activity */
	.td-profile__activity-list {
		display: flex;
		flex-direction: column;
	}
	.td-profile__activity-item {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.625rem 0;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}
	.td-profile__activity-item:last-child {
		border-bottom: none;
	}
	.td-profile__activity-dot {
		width: 0.4375rem;
		height: 0.4375rem;
		border-radius: 9999px;
		background: var(--dbd2, #ccc);
		flex-shrink: 0;
		margin-top: 0.375rem;
	}
	.td-profile__activity-body {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
		flex: 1;
	}
	.td-profile__activity-text {
		font-size: 0.8125rem;
		color: var(--dt, #111);
		font-weight: 500;
		line-height: 1.4;
	}
	.td-profile__activity-time {
		font-size: 0.6875rem;
		color: var(--dt4, #bbb);
		font-weight: 500;
	}
</style>
