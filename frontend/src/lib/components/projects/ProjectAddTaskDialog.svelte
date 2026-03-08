<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Task, TeamMemberListResponse, CreateTaskData } from '$lib/api';
	import { api } from '$lib/api';

	interface Props {
		open: boolean;
		projectId: string;
		tasks: Task[];
		teamMembers: TeamMemberListResponse[];
		onClose: () => void;
		onTaskCreated: () => Promise<void>;
	}

	let { open = $bindable(), projectId, tasks, teamMembers, onClose, onTaskCreated }: Props = $props();

	let newTask = $state<CreateTaskData>({
		title: '',
		description: '',
		priority: 'medium',
		due_date: '',
		estimated_hours: undefined,
		start_date: undefined,
		parent_task_id: undefined,
		assignee_id: undefined
	});

	function resetForm() {
		newTask = {
			title: '',
			description: '',
			priority: 'medium',
			due_date: '',
			estimated_hours: undefined,
			start_date: undefined,
			parent_task_id: undefined,
			assignee_id: undefined
		};
	}

	async function handleCreateTask(e: Event) {
		e.preventDefault();
		if (!newTask.title.trim()) return;
		try {
			await api.createTask({ ...newTask, project_id: projectId });
			await onTaskCreated();
			resetForm();
			onClose();
		} catch (err) {
			console.error('Error creating task:', err);
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="prm-dialog fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-2xl shadow-xl p-6 w-full max-w-md z-50">
			<Dialog.Title class="prm-dialog__title">Add Task</Dialog.Title>

			<form onsubmit={handleCreateTask} class="space-y-4">
				<div>
					<label for="task-title" class="prm-dialog__label">Title</label>
					<input
						id="task-title"
						type="text"
						bind:value={newTask.title}
						class="input input-square"
						placeholder="What needs to be done?"
						required
					/>
				</div>

				<div>
					<label for="task-description" class="prm-dialog__label">Description (optional)</label>
					<textarea
						id="task-description"
						bind:value={newTask.description}
						class="input input-square resize-none"
						rows="2"
						placeholder="Add more details..."
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
							<label for="task-priority" class="prm-dialog__label">Priority</label>
						<select id="task-priority" bind:value={newTask.priority} class="input input-square">
							<option value="low">Low</option>
							<option value="medium">Medium</option>
							<option value="high">High</option>
							<option value="critical">Critical</option>
						</select>
					</div>
					<div>
							<label for="task-due" class="prm-dialog__label">Due Date</label>
						<input id="task-due" type="date" bind:value={newTask.due_date} class="input input-square" />
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
							<label for="task-estimated" class="prm-dialog__label">Estimated Hours</label>
						<input
							id="task-estimated"
							type="number"
							min="0"
							step="0.5"
							bind:value={newTask.estimated_hours}
							class="input input-square"
							placeholder="0.0"
						/>
					</div>
					<div>
							<label for="task-start" class="prm-dialog__label">Start Date</label>
						<input id="task-start" type="date" bind:value={newTask.start_date} class="input input-square" />
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="task-parent" class="prm-dialog__label">Parent Task (optional)</label>
						<select id="task-parent" bind:value={newTask.parent_task_id} class="input input-square">
							<option value="">None (Top-level)</option>
							{#each tasks.filter((t) => t.status !== 'done') as task}
								<option value={task.id}>{task.title}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="task-assignee" class="prm-dialog__label">Assignee (optional)</label>
						<select id="task-assignee" bind:value={newTask.assignee_id} class="input input-square">
							<option value="">Unassigned</option>
							{#each teamMembers as member}
								<option value={member.id}>{member.name}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={onClose} class="btn-pill btn-pill-ghost flex-1">Cancel</button>
					<button type="submit" class="btn-pill btn-pill-primary flex-1">Add Task</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	.prm-dialog {
		background: var(--dbg, #fff);
	}
	.prm-dialog__title {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--dt, #111);
		margin-bottom: 1rem;
	}
	.prm-dialog__label {
		display: block;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--dt, #111);
		margin-bottom: 0.25rem;
	}
</style>
