<script lang="ts">
	import type { Project } from '$lib/api';
	import { api } from '$lib/api';
	import { formatDate, formatTime } from '$lib/utils/project';

	interface Props {
		project: Project;
		onProjectUpdate: () => Promise<void>;
	}

	let { project, onProjectUpdate }: Props = $props();

	let newNote = $state('');
	let isAddingNote = $state(false);

	async function handleAddNote() {
		if (!newNote.trim()) return;
		isAddingNote = true;
		try {
			await api.addProjectNote(project.id, newNote);
			await onProjectUpdate();
			newNote = '';
		} catch (err) {
			console.error('Error adding note:', err);
		} finally {
			isAddingNote = false;
		}
	}
</script>

<div class="bg-white rounded-xl border border-gray-200 p-6">
	<h2 class="text-lg font-medium text-gray-900 mb-4">Notes</h2>

	<!-- Add Note -->
	<div class="mb-6">
		<textarea
			bind:value={newNote}
			placeholder="Add a note..."
			class="input input-square resize-none w-full"
			rows="3"
		></textarea>
		<div class="flex justify-end mt-2">
			<button
				onclick={handleAddNote}
				disabled={!newNote.trim() || isAddingNote}
				class="btn btn-primary text-sm disabled:opacity-50"
			>
				{isAddingNote ? 'Adding...' : 'Add Note'}
			</button>
		</div>
	</div>

	<!-- Notes List -->
	{#if project.notes.length === 0}
		<div class="text-center py-8 text-gray-400">
			<svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
			</svg>
			<p>No notes yet. Add your first note above.</p>
		</div>
	{:else}
		<div class="space-y-4">
			{#each project.notes as note}
				<div class="p-4 bg-gray-50 rounded-xl">
					<p class="text-gray-700 whitespace-pre-wrap">{note.content}</p>
					<p class="text-xs text-gray-400 mt-3">
						{formatDate(note.created_at)} at {formatTime(note.created_at)}
					</p>
				</div>
			{/each}
		</div>
	{/if}
</div>
