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

	function handleToggleTask(taskId: string) {
		api.toggleTask(taskId).then(() => {});
	}
</script>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Main Content -->
	<div class="lg:col-span-2 space-y-6">
		<!-- Description -->
		<div class="bg-white rounded-xl border border-gray-200 p-6">
			<h2 class="text-lg font-medium text-gray-900 mb-3 flex items-center gap-2">
				<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
				</svg>
				Description
			</h2>
			{#if project.description}
				<p class="text-gray-600 whitespace-pre-wrap">{project.description}</p>
			{:else}
				<p class="text-gray-400 italic">No description added yet. Click Edit to add one.</p>
			{/if}
		</div>

		<!-- Recent Tasks -->
		<div class="bg-white rounded-xl border border-gray-200 p-6">
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-medium text-gray-900 flex items-center gap-2">
					<svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
					Tasks
				</h2>
				<button
					onclick={() => { onNavigateToTasks(); onShowAddTask(); }}
					class="text-sm text-purple-600 hover:text-purple-700 font-medium"
				>
					+ Add Task
				</button>
			</div>
			{#if tasks.length === 0}
				<div class="text-center py-8">
					<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mx-auto mb-3">
						<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
						</svg>
					</div>
					<p class="text-gray-500 mb-2">No tasks yet</p>
					<button
						onclick={() => { onNavigateToTasks(); onShowAddTask(); }}
						class="btn btn-primary text-sm"
					>
						Add First Task
					</button>
				</div>
			{:else}
				<div class="space-y-2">
					{#each tasks.slice(0, 5) as task}
						<div class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 group">
							<button
								onclick={() => handleToggleTask(task.id)}
								class="w-5 h-5 rounded border-2 flex items-center justify-center flex-shrink-0 transition-colors {task.status === 'done' ? 'bg-purple-600 border-purple-600 text-white' : 'border-gray-300 hover:border-purple-600'}"
								aria-label="Toggle task complete"
							>
								{#if task.status === 'done'}
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
									</svg>
								{/if}
							</button>
							<div class="flex-1 min-w-0">
								<p class="text-sm {task.status === 'done' ? 'text-gray-400 line-through' : 'text-gray-900'}">{task.title}</p>
								{#if task.due_date}
									<p class="text-xs text-gray-400">Due {formatDate(task.due_date)}</p>
								{/if}
							</div>
							<span class="text-xs px-2 py-0.5 rounded {getPriorityColor(task.priority)}">{task.priority}</span>
						</div>
					{/each}
					{#if tasks.length > 5}
						<button
							onclick={onNavigateToTasks}
							class="text-sm text-purple-600 hover:text-purple-700 font-medium w-full text-center py-2"
						>
							View all {tasks.length} tasks
						</button>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<!-- Sidebar -->
	<div class="space-y-6">
		<!-- Quick Actions -->
		<div class="bg-white rounded-xl border border-gray-200 p-6">
			<h2 class="text-lg font-medium text-gray-900 mb-3">Quick Actions</h2>
			<div class="space-y-2">
				{#if project.status !== 'completed'}
					<button
						onclick={async () => {
							await api.updateProject(project.id, { status: 'completed' });
							await onProjectUpdate();
						}}
						class="btn btn-secondary w-full text-sm justify-start"
					>
						<svg class="w-4 h-4 mr-2 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
						class="btn btn-secondary w-full text-sm justify-start"
					>
						<svg class="w-4 h-4 mr-2 text-amber-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
						class="btn btn-secondary w-full text-sm justify-start"
					>
						<svg class="w-4 h-4 mr-2 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						Resume Project
					</button>
				{/if}
				<button
					onclick={() => { onNavigateToTasks(); onShowAddTask(); }}
					class="btn btn-secondary w-full text-sm justify-start"
				>
					<svg class="w-4 h-4 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Add Task
				</button>
				<a href="/knowledge{embedSuffix}" class="btn btn-secondary w-full text-sm justify-start">
					<svg class="w-4 h-4 mr-2 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
						class="btn btn-secondary w-full text-sm justify-start text-gray-500"
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
		<div class="bg-white rounded-xl border border-gray-200 p-6">
			<h2 class="text-lg font-medium text-gray-900 mb-3">Details</h2>
			<dl class="space-y-3">
				<div>
					<dt class="text-xs text-gray-500 uppercase">Status</dt>
					<dd class="text-sm font-medium capitalize">{project.status}</dd>
				</div>
				<div>
					<dt class="text-xs text-gray-500 uppercase">Priority</dt>
					<dd class="text-sm font-medium capitalize {getPriorityColor(project.priority)} inline-block px-2 py-0.5 rounded">{project.priority}</dd>
				</div>
				<div>
					<dt class="text-xs text-gray-500 uppercase">Type</dt>
					<dd class="text-sm text-gray-900">{getTypeLabel(project.project_type)}</dd>
				</div>
				{#if project.client_name}
					<div>
						<dt class="text-xs text-gray-500 uppercase">Client</dt>
						<dd class="text-sm text-gray-900">{project.client_name}</dd>
					</div>
				{/if}
				<div>
					<dt class="text-xs text-gray-500 uppercase">Created</dt>
					<dd class="text-sm text-gray-900">{formatDate(project.created_at)}</dd>
				</div>
				<div>
					<dt class="text-xs text-gray-500 uppercase">Last Updated</dt>
					<dd class="text-sm text-gray-900">{formatDate(project.updated_at)}</dd>
				</div>
			</dl>
		</div>

		<!-- Team Members -->
		{#if teamMembers.length > 0}
			<div class="bg-white rounded-xl border border-gray-200 p-6">
				<div class="flex items-center justify-between mb-3">
					<h2 class="text-lg font-medium text-gray-900">Team</h2>
					<button onclick={onShowAssignTeam} class="text-sm text-purple-600 hover:text-purple-700">
						+ Assign
					</button>
				</div>
				<div class="space-y-2">
					{#each teamMembers.slice(0, 3) as member}
						<div class="flex items-center gap-2 p-2 rounded-lg hover:bg-gray-50">
							<div class="w-8 h-8 rounded-full bg-gradient-to-br from-purple-400 to-indigo-500 flex items-center justify-center text-white text-xs font-medium">
								{member.name.split(' ').map((n: string) => n[0]).join('').slice(0, 2)}
							</div>
							<div class="flex-1 min-w-0">
								<p class="text-sm font-medium text-gray-900 truncate">{member.name}</p>
								<p class="text-xs text-gray-400 truncate">{member.role}</p>
							</div>
						</div>
					{/each}
					{#if teamMembers.length > 3}
						<p class="text-xs text-gray-400 text-center">+{teamMembers.length - 3} more</p>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>
