<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { Task } from '$lib/api';
	import { api } from '$lib/api';

	interface Props {
		open: boolean;
		task: Task | null;
		onClose: () => void;
		onTaskUpdated: () => Promise<void>;
	}

	let { open = $bindable(), task = $bindable(), onClose, onTaskUpdated }: Props = $props();

	async function handleUpdateTask(e: Event) {
		e.preventDefault();
		if (!task) return;
		try {
			await api.updateTask(task.id, {
				title: task.title,
				description: task.description || '',
				priority: task.priority,
				status: task.status,
				due_date: task.due_date || ''
			});
			await onTaskUpdated();
			onClose();
		} catch (err) {
			console.error('Error updating task:', err);
		}
	}
</script>

{#if task}
	<Dialog.Root bind:open>
		<Dialog.Portal>
			<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
			<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-md z-50">
				<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">Edit Task</Dialog.Title>

				<form onsubmit={handleUpdateTask} class="space-y-4">
					<div>
						<label for="edit-task-title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
						<input
							id="edit-task-title"
							type="text"
							bind:value={task.title}
							class="input input-square"
							placeholder="Task title..."
							required
						/>
					</div>

					<div>
						<label for="edit-task-description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
						<textarea
							id="edit-task-description"
							bind:value={task.description}
							class="input input-square resize-none"
							rows="2"
							placeholder="Add more details..."
						></textarea>
					</div>

					<div class="grid grid-cols-2 gap-3">
						<div>
							<label for="edit-task-priority" class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
							<select id="edit-task-priority" bind:value={task.priority} class="input input-square">
								<option value="low">Low</option>
								<option value="medium">Medium</option>
								<option value="high">High</option>
								<option value="critical">Critical</option>
							</select>
						</div>
						<div>
							<label for="edit-task-due" class="block text-sm font-medium text-gray-700 mb-1">Due Date</label>
							<input id="edit-task-due" type="date" bind:value={task.due_date} class="input input-square" />
						</div>
					</div>

					<div class="grid grid-cols-2 gap-3">
						<div>
							<label for="edit-task-estimated" class="block text-sm font-medium text-gray-700 mb-1">Estimated Hours</label>
							<input
								id="edit-task-estimated"
								type="number"
								min="0"
								step="0.5"
								bind:value={task.estimated_hours}
								class="input input-square"
								placeholder="0.0"
							/>
						</div>
						<div>
							<label for="edit-task-start" class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
							<input id="edit-task-start" type="date" bind:value={task.start_date} class="input input-square" />
						</div>
					</div>

					<div>
						<label for="edit-task-status" class="block text-sm font-medium text-gray-700 mb-1">Status</label>
						<select id="edit-task-status" bind:value={task.status} class="input input-square">
							<option value="todo">To Do</option>
							<option value="in_progress">In Progress</option>
							<option value="done">Done</option>
							<option value="cancelled">Cancelled</option>
						</select>
					</div>

					<div class="flex gap-3 pt-2">
						<button type="button" onclick={onClose} class="btn btn-secondary flex-1">Cancel</button>
						<button type="submit" class="btn btn-primary flex-1">Save Changes</button>
					</div>
				</form>
			</Dialog.Content>
		</Dialog.Portal>
	</Dialog.Root>
{/if}
