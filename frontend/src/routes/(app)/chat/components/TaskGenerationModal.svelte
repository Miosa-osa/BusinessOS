<script lang="ts">
	interface GeneratedTask {
		title: string;
		description?: string;
		priority: 'high' | 'medium' | 'low';
		assignee_id?: string;
	}

	interface Project {
		id: string;
		name: string;
	}

	interface TeamMember {
		id: string;
		name: string;
		role: string;
	}

	interface ArtifactRef {
		title: string;
		type: string;
	}

	interface Props {
		show: boolean;
		generatingTasks: boolean;
		generatedTasks: GeneratedTask[];
		selectedProjectForTasks: string;
		availableProjects: Project[];
		availableTeamMembers: TeamMember[];
		taskGenerationArtifact: ArtifactRef | null;
		onClose: () => void;
		onSelectProject: (id: string) => void;
		onRemoveTask: (index: number) => void;
		onUpdateTaskAssignee: (index: number, assigneeId: string) => void;
		onConfirm: () => void;
	}

	let {
		show,
		generatingTasks,
		generatedTasks,
		selectedProjectForTasks,
		availableProjects,
		availableTeamMembers,
		taskGenerationArtifact,
		onClose,
		onSelectProject,
		onRemoveTask,
		onUpdateTaskAssignee,
		onConfirm,
	}: Props = $props();
</script>

{#if show}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" onclick={onClose}>
		<div class="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[85vh] flex flex-col" onclick={(e) => e.stopPropagation()}>
			<!-- Header -->
			<div class="p-5 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
				<div>
					<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Generate Tasks from Plan</h3>
					<p class="text-sm text-gray-500 mt-0.5">Review and assign tasks extracted from "{taskGenerationArtifact?.title}"</p>
				</div>
				<button
					onclick={onClose}
					class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-sm"
					aria-label="Close modal"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Project Selection -->
			<div class="px-5 py-3 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
				<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Assign to Project</label>
				<select
					value={selectedProjectForTasks}
					onchange={(e) => onSelectProject((e.target as HTMLSelectElement).value)}
					class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-100 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				>
					<option value="">Select a project...</option>
					{#each availableProjects as project}
						<option value={project.id}>{project.name}</option>
					{/each}
				</select>
			</div>

			<!-- Tasks List -->
			<div class="flex-1 overflow-y-auto p-5">
				{#if generatingTasks}
					<div class="flex flex-col items-center justify-center py-12">
						<div class="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center mb-4">
							<svg class="w-6 h-6 text-blue-600 animate-spin" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
						</div>
						<p class="text-sm font-medium text-gray-900">Analyzing plan...</p>
						<p class="text-xs text-gray-500 mt-1">Extracting actionable tasks from your artifact</p>
					</div>
				{:else if generatedTasks.length === 0}
					<div class="flex flex-col items-center justify-center py-12">
						<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-4">
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
							</svg>
						</div>
						<p class="text-sm font-medium text-gray-900">No tasks extracted</p>
						<p class="text-xs text-gray-500 mt-1">Try with a different artifact or add tasks manually</p>
					</div>
				{:else}
					<div class="space-y-3">
						{#each generatedTasks as task, index}
							<div class="border border-gray-200 rounded-xl p-4 hover:border-gray-300 transition-colors">
								<div class="flex items-start justify-between gap-3 mb-2">
									<div class="flex-1 min-w-0">
										<h4 class="font-medium text-gray-900 text-sm">{task.title}</h4>
										{#if task.description}
											<p class="text-xs text-gray-500 mt-1 line-clamp-2">{task.description}</p>
										{/if}
									</div>
									<button
										onclick={() => onRemoveTask(index)}
										class="btn-pill btn-pill-danger btn-pill-icon btn-pill-sm flex-shrink-0"
										aria-label="Remove task"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
										</svg>
									</button>
								</div>
								<div class="flex items-center gap-3 mt-3">
									<div class="flex items-center gap-2">
										<span class="text-xs text-gray-500">Priority:</span>
										<span class="px-2 py-0.5 text-xs font-medium rounded-full {task.priority === 'high' ? 'bg-red-100 text-red-700' : task.priority === 'medium' ? 'bg-yellow-100 text-yellow-700' : 'bg-gray-100 text-gray-700'}">{task.priority}</span>
									</div>
									<div class="flex items-center gap-2 flex-1 min-w-0">
										<span class="text-xs text-gray-500 flex-shrink-0">Assign to:</span>
										<select
											value={task.assignee_id || ''}
											onchange={(e) => onUpdateTaskAssignee(index, (e.target as HTMLSelectElement).value)}
											class="flex-1 min-w-0 px-2 py-1 text-xs border border-gray-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-blue-500"
										>
											<option value="">Unassigned</option>
											{#each availableTeamMembers as member}
												<option value={member.id}>{member.name} ({member.role})</option>
											{/each}
										</select>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="p-4 border-t border-gray-100 flex items-center justify-between">
				<div class="text-sm text-gray-500">
					{generatedTasks.length} task{generatedTasks.length !== 1 ? 's' : ''} ready
				</div>
				<div class="flex gap-3">
					<button
						onclick={onClose}
						class="btn-pill btn-pill-soft btn-pill-sm"
					>
						Cancel
					</button>
					<button
						onclick={onConfirm}
						disabled={!selectedProjectForTasks || generatedTasks.length === 0}
						class="btn-pill btn-pill-success btn-pill-sm"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						Create {generatedTasks.length} Task{generatedTasks.length !== 1 ? 's' : ''}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
