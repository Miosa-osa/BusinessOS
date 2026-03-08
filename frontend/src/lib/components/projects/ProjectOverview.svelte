<script lang="ts">
	import type { Project, Task, TeamMemberListResponse, ClientListResponse } from '$lib/api';
	import { api } from '$lib/api';
	import { getPriorityColor, getTypeLabel, getTypeIcon, formatDate } from '$lib/utils/project';

	interface Props {
		project: Project;
		tasks: Task[];
		teamMembers: TeamMemberListResponse[];
		clients: ClientListResponse[];
		embedSuffix: string;
		onProjectUpdate: () => Promise<void>;
		onNavigateToTasks: () => void;
		onShowAddTask: () => void;
		onShowAssignTeam: () => void;
	}

	let {
		project,
		tasks,
		teamMembers,
		clients,
		embedSuffix,
		onProjectUpdate,
		onNavigateToTasks,
		onShowAddTask,
		onShowAssignTeam
	}: Props = $props();

	let completedTasks = $derived(tasks.filter((t) => t.status === 'done').length);
	let totalTasks = $derived(tasks.length);
	let completionPct = $derived(totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0);

	// Metrics
	let todoCount = $derived(tasks.filter(t => t.status === 'todo').length);
	let inProgressCount = $derived(tasks.filter(t => t.status === 'in_progress').length);
	let cancelledCount = $derived(tasks.filter(t => t.status === 'cancelled').length);

	let criticalCount = $derived(tasks.filter(t => t.priority === 'critical').length);
	let highCount = $derived(tasks.filter(t => t.priority === 'high').length);
	let mediumCount = $derived(tasks.filter(t => t.priority === 'medium').length);
	let lowCount = $derived(tasks.filter(t => t.priority === 'low').length);

	// Project health indicator
	let healthScore = $derived((() => {
		if (totalTasks === 0) return 'new';
		if (criticalCount > 0) return 'at-risk';
		if (completionPct >= 75) return 'on-track';
		if (highCount > completedTasks) return 'needs-attention';
		return 'on-track';
	})());
	let healthLabel = $derived((() => {
		if (healthScore === 'new') return 'New Project';
		if (healthScore === 'at-risk') return 'At Risk';
		if (healthScore === 'needs-attention') return 'Needs Attention';
		return 'On Track';
	})());

	// Overdue tasks
	let overdueTasks = $derived(tasks.filter(t => {
		if (t.status === 'done' || t.status === 'cancelled' || !t.due_date) return false;
		return new Date(t.due_date) < new Date();
	}));

	// Upcoming tasks (due within 7 days)
	let upcomingTasksList = $derived((() => {
		const now = new Date();
		const weekOut = new Date();
		weekOut.setDate(weekOut.getDate() + 7);
		return tasks.filter(t => {
			if (t.status === 'done' || t.status === 'cancelled' || !t.due_date) return false;
			const d = new Date(t.due_date);
			return d >= now && d <= weekOut;
		});
	})());

	// SVG donut for completion ring
	let ringRadius = 42;
	let ringCircumference = $derived(2 * Math.PI * ringRadius);
	let ringOffset = $derived(ringCircumference - (completionPct / 100) * ringCircumference);

	function handleToggleTask(taskId: string) {
		api.toggleTask(taskId).then(() => {
			onProjectUpdate();
		}).catch((err: unknown) => {
			console.error('Failed to toggle task:', err);
		});
	}

	function daysAgo(dateStr: string): string {
		const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 86400000);
		if (diff === 0) return 'Today';
		if (diff === 1) return 'Yesterday';
		return `${diff} days ago`;
	}

	function statusColor(status: string): string {
		switch (status) {
			case 'active': return '#22c55e';
			case 'paused': return '#f59e0b';
			case 'completed': return '#8b5cf6';
			case 'archived': return '#6b7280';
			default: return '#3b82f6';
		}
	}
</script>

<!-- Progress Header -->
<div class="prm-ov-progress-header">
	<div class="prm-ov-progress-inner">
		<!-- Health badge -->
		<div class="prm-ov-health">
			<span class="prm-ov-health-dot" style="background:{statusColor(project.status)}"></span>
			<span class="prm-ov-health-label">{healthLabel}</span>
		</div>

		<!-- Stats row -->
		<div class="prm-ov-kpi-row">
			<div class="prm-ov-kpi">
				<span class="prm-ov-kpi-val">{completedTasks}</span>
				<span class="prm-ov-kpi-label">Done</span>
			</div>
			<div class="prm-ov-kpi-divider"></div>
			<div class="prm-ov-kpi">
				<span class="prm-ov-kpi-val">{inProgressCount}</span>
				<span class="prm-ov-kpi-label">In Progress</span>
			</div>
			<div class="prm-ov-kpi-divider"></div>
			<div class="prm-ov-kpi">
				<span class="prm-ov-kpi-val">{todoCount}</span>
				<span class="prm-ov-kpi-label">To Do</span>
			</div>
			<div class="prm-ov-kpi-divider"></div>
			<div class="prm-ov-kpi">
				<span class="prm-ov-kpi-val prm-ov-kpi-val--accent">{completionPct}%</span>
				<span class="prm-ov-kpi-label">Complete</span>
			</div>
		</div>

		<!-- Progress bar -->
		<div class="prm-ov-progress-track">
			{#if completedTasks > 0}
				<div class="prm-ov-progress-seg prm-ov-progress-seg--done" style="width:{(completedTasks/Math.max(totalTasks,1))*100}%"></div>
			{/if}
			{#if inProgressCount > 0}
				<div class="prm-ov-progress-seg prm-ov-progress-seg--progress" style="width:{(inProgressCount/Math.max(totalTasks,1))*100}%"></div>
			{/if}
			{#if todoCount > 0}
				<div class="prm-ov-progress-seg prm-ov-progress-seg--todo" style="width:{(todoCount/Math.max(totalTasks,1))*100}%"></div>
			{/if}
		</div>
	</div>
</div>

<div class="prm-ov-layout">
	<!-- Main Content -->
	<div class="prm-ov-main">
		<!-- Alerts: overdue tasks -->
		{#if overdueTasks.length > 0}
			<div class="prm-ov-alert">
				<svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4.5c-.77-.833-2.694-.833-3.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
				</svg>
				<span>{overdueTasks.length} overdue task{overdueTasks.length > 1 ? 's' : ''} need attention</span>
			</div>
		{/if}

		<!-- Description -->
		<div class="prm-ov-card">
			<h2 class="prm-ov-heading">
				<svg class="w-4 h-4 prm-ov-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
				</svg>
				About
			</h2>
			{#if project.description}
				<p class="prm-ov-desc">{project.description}</p>
			{:else}
				<p class="prm-ov-empty">No description added yet. Click Edit to add one.</p>
			{/if}
			<!-- Inline details for quick scanning -->
			<div class="prm-ov-inline-details">
				<div class="prm-ov-detail-chip">
					<span class="prm-ov-detail-chip-dot" style="background:{statusColor(project.status)}"></span>
					<span class="capitalize">{project.status}</span>
				</div>
				<div class="prm-ov-detail-chip">{getTypeLabel(project.project_type)}</div>
				<div class="prm-ov-detail-chip capitalize">{project.priority} priority</div>
				{#if project.client_name}
					<div class="prm-ov-detail-chip">{project.client_name}</div>
				{/if}
				<div class="prm-ov-detail-chip">Updated {daysAgo(project.updated_at)}</div>
			</div>
		</div>

		<!-- Tasks -->
		<div class="prm-ov-card">
			<div class="prm-ov-card-header">
				<h2 class="prm-ov-heading" style="margin-bottom:0">
					<svg class="w-4 h-4 prm-ov-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
					Tasks
					{#if totalTasks > 0}
						<span class="prm-ov-count-badge">{totalTasks}</span>
					{/if}
				</h2>
				<button
					onclick={() => { onNavigateToTasks(); onShowAddTask(); }}
					class="btn-pill btn-pill-ghost btn-pill-sm"
				>
					+ Add Task
				</button>
			</div>
			{#if tasks.length === 0}
				<div class="prm-ov-empty-state">
					<div class="prm-ov-empty-circle">
						<svg class="w-6 h-6 prm-ov-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
						</svg>
					</div>
					<p class="prm-ov-empty-title">No tasks yet</p>
					<p class="prm-ov-empty-sub">Create your first task to start tracking progress</p>
					<button
						onclick={() => { onNavigateToTasks(); onShowAddTask(); }}
						class="btn-pill btn-pill-primary btn-pill-sm"
					>
						Add First Task
					</button>
				</div>
			{:else}
				<div class="prm-ov-task-list">
					{#each tasks.slice(0, 6) as task}
						<div class="prm-ov-task-row">
							<button
								onclick={() => handleToggleTask(task.id)}
								class="prm-ov-checkbox {task.status === 'done' ? 'prm-ov-checkbox--done' : ''}"
								aria-label="Toggle task complete"
							>
								{#if task.status === 'done'}
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
									</svg>
								{/if}
							</button>
							<div class="prm-ov-task-body">
								<p class="prm-ov-task-title {task.status === 'done' ? 'prm-ov-done' : ''}">{task.title}</p>
								<div class="prm-ov-task-meta">
									{#if task.due_date}
										{@const isOverdue = new Date(task.due_date) < new Date() && task.status !== 'done'}
										<span class="prm-ov-task-due {isOverdue ? 'prm-ov-task-due--overdue' : ''}">
											Due {formatDate(task.due_date)}
										</span>
									{/if}
									{#if task.status === 'in_progress'}
										<span class="prm-ov-task-status-chip">In Progress</span>
									{/if}
								</div>
							</div>
							<span class="prm-ov-priority-badge prm-ov-priority-badge--{task.priority}">{task.priority}</span>
						</div>
					{/each}
					{#if tasks.length > 6}
						<button
							onclick={onNavigateToTasks}
							class="btn-pill btn-pill-ghost btn-pill-sm prm-ov-view-all"
						>
							View all {tasks.length} tasks →
						</button>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Metrics Dashboard -->
		{#if totalTasks > 0}
			<div class="prm-ov-card">
				<h2 class="prm-ov-heading">
					<svg class="w-4 h-4 prm-ov-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
					</svg>
					Analytics
				</h2>
				<div class="prm-ov-analytics">
					<!-- Ring -->
					<div class="prm-ov-ring-wrap">
						<svg viewBox="0 0 100 100" class="prm-ov-ring-svg">
							<circle cx="50" cy="50" r={ringRadius} fill="none" stroke="var(--dbg3, #e5e7eb)" stroke-width="8" />
							<circle cx="50" cy="50" r={ringRadius} fill="none" stroke="#8b5cf6" stroke-width="8"
								stroke-dasharray={ringCircumference} stroke-dashoffset={ringOffset}
								stroke-linecap="round" transform="rotate(-90 50 50)"
								style="transition: stroke-dashoffset 0.6s ease"
							/>
							<text x="50" y="46" text-anchor="middle" class="prm-ov-ring-pct">{completionPct}%</text>
							<text x="50" y="60" text-anchor="middle" class="prm-ov-ring-label">complete</text>
						</svg>
						<p class="prm-ov-ring-sub">{completedTasks} of {totalTasks} tasks</p>
					</div>

					<!-- Breakdowns -->
					<div class="prm-ov-breakdowns">
						<!-- Status -->
						<div class="prm-ov-breakdown">
							<span class="prm-ov-breakdown-title">By Status</span>
							<div class="prm-ov-breakdown-bars">
								<div class="prm-ov-bar-row">
									<span class="prm-ov-bar-label">Done</span>
									<div class="prm-ov-bar-track"><div class="prm-ov-bar-fill" style="width:{(completedTasks/totalTasks)*100}%;background:#22c55e"></div></div>
									<span class="prm-ov-bar-val">{completedTasks}</span>
								</div>
								<div class="prm-ov-bar-row">
									<span class="prm-ov-bar-label">Active</span>
									<div class="prm-ov-bar-track"><div class="prm-ov-bar-fill" style="width:{(inProgressCount/totalTasks)*100}%;background:#3b82f6"></div></div>
									<span class="prm-ov-bar-val">{inProgressCount}</span>
								</div>
								<div class="prm-ov-bar-row">
									<span class="prm-ov-bar-label">To Do</span>
									<div class="prm-ov-bar-track"><div class="prm-ov-bar-fill" style="width:{(todoCount/totalTasks)*100}%;background:#9ca3af"></div></div>
									<span class="prm-ov-bar-val">{todoCount}</span>
								</div>
								{#if cancelledCount > 0}
									<div class="prm-ov-bar-row">
										<span class="prm-ov-bar-label">Cancelled</span>
										<div class="prm-ov-bar-track"><div class="prm-ov-bar-fill" style="width:{(cancelledCount/totalTasks)*100}%;background:#ef4444"></div></div>
										<span class="prm-ov-bar-val">{cancelledCount}</span>
									</div>
								{/if}
							</div>
						</div>

						<!-- Priority -->
						<div class="prm-ov-breakdown">
							<span class="prm-ov-breakdown-title">By Priority</span>
							<div class="prm-ov-breakdown-bars">
								{#if criticalCount > 0}
									<div class="prm-ov-bar-row">
										<span class="prm-ov-bar-label">Critical</span>
										<div class="prm-ov-bar-track"><div class="prm-ov-bar-fill" style="width:{(criticalCount/totalTasks)*100}%;background:#ef4444"></div></div>
										<span class="prm-ov-bar-val">{criticalCount}</span>
									</div>
								{/if}
								{#if highCount > 0}
									<div class="prm-ov-bar-row">
										<span class="prm-ov-bar-label">High</span>
										<div class="prm-ov-bar-track"><div class="prm-ov-bar-fill" style="width:{(highCount/totalTasks)*100}%;background:#f97316"></div></div>
										<span class="prm-ov-bar-val">{highCount}</span>
									</div>
								{/if}
								<div class="prm-ov-bar-row">
									<span class="prm-ov-bar-label">Medium</span>
									<div class="prm-ov-bar-track"><div class="prm-ov-bar-fill" style="width:{(mediumCount/totalTasks)*100}%;background:#eab308"></div></div>
									<span class="prm-ov-bar-val">{mediumCount}</span>
								</div>
								<div class="prm-ov-bar-row">
									<span class="prm-ov-bar-label">Low</span>
									<div class="prm-ov-bar-track"><div class="prm-ov-bar-fill" style="width:{(lowCount/totalTasks)*100}%;background:#22c55e"></div></div>
									<span class="prm-ov-bar-val">{lowCount}</span>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>

	<!-- Sidebar -->
	<div class="prm-ov-sidebar">
		<!-- Quick Actions -->
		<div class="prm-ov-card">
			<h2 class="prm-ov-heading">Quick Actions</h2>
			<div class="prm-ov-actions">
				{#if project.status !== 'completed'}
					<button
						onclick={async () => {
							await api.updateProject(project.id, { status: 'completed' });
							await onProjectUpdate();
						}}
						class="btn-pill btn-pill-ghost btn-pill-sm w-full justify-start"
					>
						<svg class="w-4 h-4 mr-2 prm-ov-action-icon--green" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						Mark Complete
					</button>
				{/if}
				{#if project.status === 'active'}
					<button
						onclick={async () => {
							await api.updateProject(project.id, { status: 'paused' });
							await onProjectUpdate();
						}}
						class="btn-pill btn-pill-ghost btn-pill-sm w-full justify-start"
					>
						<svg class="w-4 h-4 mr-2 prm-ov-action-icon--amber" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						Pause Project
					</button>
				{:else if project.status === 'paused'}
					<button
						onclick={async () => {
							await api.updateProject(project.id, { status: 'active' });
							await onProjectUpdate();
						}}
						class="btn-pill btn-pill-ghost btn-pill-sm w-full justify-start"
					>
						<svg class="w-4 h-4 mr-2 prm-ov-action-icon--green" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						Resume Project
					</button>
				{/if}
				<button
					onclick={() => { onNavigateToTasks(); onShowAddTask(); }}
					class="btn-pill btn-pill-ghost btn-pill-sm w-full justify-start"
				>
					<svg class="w-4 h-4 mr-2 prm-ov-action-icon--purple" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Add Task
				</button>
				<a href="/knowledge{embedSuffix}" class="btn-pill btn-pill-ghost btn-pill-sm w-full justify-start">
					<svg class="w-4 h-4 mr-2 prm-ov-action-icon--blue" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
					</svg>
					View Documents
				</a>
				{#if project.status !== 'archived'}
					<button
						onclick={async () => {
							await api.updateProject(project.id, { status: 'archived' });
							await onProjectUpdate();
						}}
						class="btn-pill btn-pill-ghost btn-pill-sm w-full justify-start prm-ov-muted"
					>
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
						</svg>
						Archive
					</button>
				{/if}
			</div>
		</div>

		<!-- Details -->
		<div class="prm-ov-card">
			<h2 class="prm-ov-heading">Details</h2>
			<dl class="prm-ov-details-list">
				<div class="prm-ov-detail-row">
					<dt class="prm-ov-dt">Status</dt>
					<dd class="prm-ov-dd capitalize">{project.status}</dd>
				</div>
				<div class="prm-ov-detail-row">
					<dt class="prm-ov-dt">Priority</dt>
					<dd><span class="prm-ov-priority-badge prm-ov-priority-badge--{project.priority}">{project.priority}</span></dd>
				</div>
				<div class="prm-ov-detail-row">
					<dt class="prm-ov-dt">Type</dt>
					<dd class="prm-ov-dd">{getTypeLabel(project.project_type)}</dd>
				</div>
				{#if project.client_name}
					<div class="prm-ov-detail-row">
						<dt class="prm-ov-dt">Client</dt>
						<dd class="prm-ov-dd">{project.client_name}</dd>
					</div>
				{/if}
				<div class="prm-ov-detail-row">
					<dt class="prm-ov-dt">Created</dt>
					<dd class="prm-ov-dd">{formatDate(project.created_at)}</dd>
				</div>
				<div class="prm-ov-detail-row">
					<dt class="prm-ov-dt">Updated</dt>
					<dd class="prm-ov-dd">{formatDate(project.updated_at)}</dd>
				</div>
			</dl>
		</div>

		<!-- Team Members -->
		<div class="prm-ov-card">
			<div class="prm-ov-card-header">
				<h2 class="prm-ov-heading" style="margin-bottom:0">
					Team
					{#if teamMembers.length > 0}
						<span class="prm-ov-count-badge">{teamMembers.length}</span>
					{/if}
				</h2>
				<button onclick={onShowAssignTeam} class="btn-pill btn-pill-ghost btn-pill-sm">
					+ Assign
				</button>
			</div>
			{#if teamMembers.length > 0}
				<div class="prm-ov-team-list">
					{#each teamMembers.slice(0, 4) as member}
						<div class="prm-ov-team-member">
							<div class="prm-ov-team-avatar">
								{member.name.split(' ').map((n: string) => n[0]).join('').slice(0, 2)}
							</div>
							<div class="prm-ov-team-info">
								<p class="prm-ov-team-name">{member.name}</p>
								<p class="prm-ov-team-role">{member.role}</p>
							</div>
						</div>
					{/each}
					{#if teamMembers.length > 4}
						<p class="prm-ov-team-more">+{teamMembers.length - 4} more member{teamMembers.length - 4 > 1 ? 's' : ''}</p>
					{/if}
				</div>
			{:else}
				<p class="prm-ov-empty" style="margin-top:0.75rem">No team members assigned</p>
			{/if}
		</div>

		<!-- Upcoming Deadlines -->
		{#if upcomingTasksList.length > 0}
			<div class="prm-ov-card">
				<h2 class="prm-ov-heading">
					<svg class="w-4 h-4 prm-ov-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					Due This Week
				</h2>
				<div class="prm-ov-upcoming">
					{#each upcomingTasksList.slice(0, 3) as task}
						<div class="prm-ov-upcoming-item">
							<p class="prm-ov-upcoming-title">{task.title}</p>
							<p class="prm-ov-upcoming-date">{formatDate(task.due_date ?? '')}</p>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	/* ─── Project Overview ─ Foundation Tokens ─────────────────── */

	/* Progress Header */
	.prm-ov-progress-header {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 0.875rem;
		padding: 1.25rem 1.5rem;
		margin-bottom: 1.5rem;
	}
	.prm-ov-progress-inner {
		display: flex;
		flex-direction: column;
		gap: 0.875rem;
	}
	.prm-ov-health {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.prm-ov-health-dot {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
	}
	.prm-ov-health-label {
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--dt2, #555);
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	/* KPI row */
	.prm-ov-kpi-row {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}
	.prm-ov-kpi {
		display: flex;
		flex-direction: column;
	}
	.prm-ov-kpi-val {
		font-size: 1.5rem;
		font-weight: 800;
		color: var(--dt, #111);
		letter-spacing: -0.03em;
		line-height: 1;
	}
	.prm-ov-kpi-val--accent { color: #8b5cf6; }
	.prm-ov-kpi-label {
		font-size: 0.6875rem;
		font-weight: 500;
		color: var(--dt3, #888);
		margin-top: 0.125rem;
	}
	.prm-ov-kpi-divider {
		width: 1px;
		height: 2rem;
		background: var(--dbd, #e0e0e0);
	}

	/* Progress bar */
	.prm-ov-progress-track {
		display: flex;
		height: 0.375rem;
		border-radius: 9999px;
		overflow: hidden;
		background: var(--dbg2, #f5f5f5);
	}
	.prm-ov-progress-seg { height: 100%; transition: width 0.4s ease; }
	.prm-ov-progress-seg--done { background: #22c55e; }
	.prm-ov-progress-seg--progress { background: #3b82f6; }
	.prm-ov-progress-seg--todo { background: transparent; }

	/* Layout */
	.prm-ov-layout {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 1.5rem;
	}
	@media (max-width: 768px) {
		.prm-ov-layout { grid-template-columns: 1fr; }
	}
	.prm-ov-main { display: flex; flex-direction: column; gap: 1rem; }
	.prm-ov-sidebar { display: flex; flex-direction: column; gap: 1rem; }

	/* Alert */
	.prm-ov-alert {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.625rem 1rem;
		border-radius: 0.5rem;
		background: color-mix(in srgb, #ef4444 8%, transparent);
		border: 1px solid color-mix(in srgb, #ef4444 20%, transparent);
		color: #ef4444;
		font-size: 0.8125rem;
		font-weight: 600;
	}

	/* Cards */
	.prm-ov-card {
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 0.875rem;
		padding: 1.25rem;
	}
	.prm-ov-card-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.875rem;
	}
	.prm-ov-heading {
		font-size: 0.8125rem;
		font-weight: 700;
		color: var(--dt, #111);
		margin-bottom: 0.75rem;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.prm-ov-count-badge {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt2, #555);
		font-size: 0.6875rem;
		font-weight: 700;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
		letter-spacing: 0;
	}
	.prm-ov-icon { color: var(--dt3, #888); }
	.prm-ov-desc { color: var(--dt2, #555); white-space: pre-wrap; line-height: 1.6; font-size: 0.875rem; }
	.prm-ov-empty { color: var(--dt3, #888); font-style: italic; font-size: 0.8125rem; }
	.prm-ov-muted { color: var(--dt2, #555); }

	/* Inline details chips */
	.prm-ov-inline-details {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--dbd, #e0e0e0);
	}
	.prm-ov-detail-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		height: 1.5rem;
		padding: 0 0.625rem;
		border-radius: 9999px;
		background: var(--dbg2, #f5f5f5);
		font-size: 0.6875rem;
		font-weight: 600;
		color: var(--dt2, #555);
	}
	.prm-ov-detail-chip-dot {
		width: 0.375rem;
		height: 0.375rem;
		border-radius: 50%;
	}

	/* Empty state */
	.prm-ov-empty-state {
		text-align: center;
		padding: 2.5rem 0;
	}
	.prm-ov-empty-circle {
		width: 3rem;
		height: 3rem;
		border-radius: 9999px;
		background: var(--dbg2, #f5f5f5);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto 0.75rem;
	}
	.prm-ov-empty-title {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--dt, #111);
		margin-bottom: 0.25rem;
	}
	.prm-ov-empty-sub {
		font-size: 0.75rem;
		color: var(--dt3, #888);
		margin-bottom: 1rem;
	}

	/* Tasks */
	.prm-ov-task-list {
		display: flex;
		flex-direction: column;
	}
	.prm-ov-task-row {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.625rem 0.5rem;
		border-radius: 0.5rem;
		transition: background 0.1s;
	}
	.prm-ov-task-row:hover { background: var(--dbg2, #f5f5f5); }
	.prm-ov-checkbox {
		width: 1.125rem;
		height: 1.125rem;
		border-radius: 0.25rem;
		border: 2px solid var(--dbd, #ccc);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		background: none;
		cursor: pointer;
		margin-top: 0.0625rem;
		transition: border-color 0.15s;
	}
	.prm-ov-checkbox:hover { border-color: #9333ea; }
	.prm-ov-checkbox--done { background: #9333ea; border-color: #9333ea; color: #fff; }
	.prm-ov-task-body { flex: 1; min-width: 0; }
	.prm-ov-task-title {
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--dt, #111);
		line-height: 1.3;
	}
	.prm-ov-done { color: var(--dt3, #888); text-decoration: line-through; }
	.prm-ov-task-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.125rem;
	}
	.prm-ov-task-due { font-size: 0.6875rem; color: var(--dt3, #888); font-weight: 500; }
	.prm-ov-task-due--overdue { color: #ef4444; }
	.prm-ov-task-status-chip {
		font-size: 0.625rem;
		font-weight: 700;
		padding: 0.0625rem 0.375rem;
		border-radius: 9999px;
		background: color-mix(in srgb, #3b82f6 12%, transparent);
		color: #3b82f6;
		text-transform: uppercase;
	}
	.prm-ov-view-all {
		margin-top: 0.25rem;
		width: 100%;
		text-align: center;
	}

	/* Priority badges */
	.prm-ov-priority-badge {
		font-size: 0.625rem;
		font-weight: 700;
		padding: 0.125rem 0.5rem;
		border-radius: 9999px;
		text-transform: uppercase;
		flex-shrink: 0;
		letter-spacing: 0.02em;
	}
	.prm-ov-priority-badge--critical { background: color-mix(in srgb, #ef4444 12%, transparent); color: #ef4444; }
	.prm-ov-priority-badge--high { background: color-mix(in srgb, #f97316 12%, transparent); color: #f97316; }
	.prm-ov-priority-badge--medium { background: color-mix(in srgb, #eab308 12%, transparent); color: #eab308; }
	.prm-ov-priority-badge--low { background: color-mix(in srgb, #22c55e 12%, transparent); color: #22c55e; }

	/* Actions */
	.prm-ov-actions {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.prm-ov-action-icon--green { color: #22c55e; }
	.prm-ov-action-icon--amber { color: #f59e0b; }
	.prm-ov-action-icon--purple { color: #9333ea; }
	.prm-ov-action-icon--blue { color: #3b82f6; }

	/* Details list */
	.prm-ov-details-list {
		display: flex;
		flex-direction: column;
		gap: 0.625rem;
	}
	.prm-ov-detail-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.prm-ov-dt {
		font-size: 0.75rem;
		color: var(--dt3, #888);
		font-weight: 500;
	}
	.prm-ov-dd {
		font-size: 0.8125rem;
		color: var(--dt, #111);
		font-weight: 500;
	}

	/* Team list */
	.prm-ov-team-list {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
		margin-top: 0.75rem;
	}
	.prm-ov-team-member {
		display: flex;
		align-items: center;
		gap: 0.625rem;
		padding: 0.375rem;
		border-radius: 0.5rem;
		transition: background 0.1s;
	}
	.prm-ov-team-member:hover { background: var(--dbg2, #f5f5f5); }
	.prm-ov-team-avatar {
		width: 2rem;
		height: 2rem;
		border-radius: 9999px;
		background: linear-gradient(135deg, #a855f7, #6366f1);
		display: flex;
		align-items: center;
		justify-content: center;
		color: #fff;
		font-size: 0.6875rem;
		font-weight: 700;
		flex-shrink: 0;
	}
	.prm-ov-team-info { flex: 1; min-width: 0; }
	.prm-ov-team-name {
		font-size: 0.8125rem;
		font-weight: 600;
		color: var(--dt, #111);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.prm-ov-team-role {
		font-size: 0.6875rem;
		color: var(--dt3, #888);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.prm-ov-team-more {
		font-size: 0.6875rem;
		color: var(--dt3, #888);
		text-align: center;
		padding-top: 0.25rem;
	}

	/* Analytics */
	.prm-ov-analytics {
		display: flex;
		gap: 1.5rem;
		align-items: flex-start;
	}
	@media (max-width: 640px) {
		.prm-ov-analytics { flex-direction: column; align-items: center; }
	}
	.prm-ov-ring-wrap {
		display: flex;
		flex-direction: column;
		align-items: center;
		flex-shrink: 0;
	}
	.prm-ov-ring-svg { width: 6.5rem; height: 6.5rem; }
	.prm-ov-ring-pct { font-size: 1rem; font-weight: 800; fill: var(--dt, #111); }
	.prm-ov-ring-label { font-size: 0.5rem; fill: var(--dt3, #6b7280); text-transform: uppercase; letter-spacing: 0.05em; }
	.prm-ov-ring-sub {
		font-size: 0.6875rem;
		color: var(--dt3, #888);
		margin-top: 0.375rem;
	}

	.prm-ov-breakdowns {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}
	.prm-ov-breakdown-title {
		display: block;
		font-size: 0.6875rem;
		font-weight: 700;
		color: var(--dt3, #888);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		margin-bottom: 0.5rem;
	}
	.prm-ov-breakdown-bars {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}
	.prm-ov-bar-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.75rem;
	}
	.prm-ov-bar-label {
		width: 3.5rem;
		color: var(--dt2, #555);
		font-weight: 500;
		flex-shrink: 0;
	}
	.prm-ov-bar-track {
		flex: 1;
		height: 0.375rem;
		border-radius: 9999px;
		background: var(--dbg2, #f5f5f5);
		overflow: hidden;
	}
	.prm-ov-bar-fill {
		height: 100%;
		border-radius: 9999px;
		transition: width 0.4s ease;
	}
	.prm-ov-bar-val {
		width: 1.5rem;
		text-align: right;
		font-weight: 700;
		color: var(--dt, #111);
		flex-shrink: 0;
	}

	/* Upcoming deadlines */
	.prm-ov-upcoming {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.prm-ov-upcoming-item {
		padding: 0.5rem;
		border-radius: 0.375rem;
		background: var(--dbg2, #f5f5f5);
	}
	.prm-ov-upcoming-title {
		font-size: 0.8125rem;
		font-weight: 500;
		color: var(--dt, #111);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.prm-ov-upcoming-date {
		font-size: 0.6875rem;
		color: var(--dt3, #888);
		margin-top: 0.125rem;
	}
</style>
