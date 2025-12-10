<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api, type Project } from '$lib/api/client';
	import { onMount } from 'svelte';
	import { Dialog } from 'bits-ui';

	let project = $state<Project | null>(null);
	let isLoading = $state(true);
	let error = $state('');
	let showEditDialog = $state(false);
	let showDeleteConfirm = $state(false);
	let isSaving = $state(false);
	let newNote = $state('');
	let isAddingNote = $state(false);

	// Edit form state
	let editForm = $state({
		name: '',
		description: '',
		status: 'active' as const,
		priority: 'medium' as const,
		client_name: '',
		project_type: 'internal'
	});

	const projectId = $derived($page.params.id);

	onMount(async () => {
		await loadProject();
	});

	async function loadProject() {
		isLoading = true;
		error = '';
		try {
			project = await api.getProject(projectId);
			if (project) {
				editForm = {
					name: project.name,
					description: project.description || '',
					status: project.status,
					priority: project.priority,
					client_name: project.client_name || '',
					project_type: project.project_type
				};
			}
		} catch (err) {
			error = 'Failed to load project';
			console.error('Error loading project:', err);
		} finally {
			isLoading = false;
		}
	}

	async function handleSave() {
		if (!project) return;
		isSaving = true;
		try {
			await api.updateProject(project.id, editForm);
			await loadProject();
			showEditDialog = false;
		} catch (err) {
			console.error('Error saving project:', err);
		} finally {
			isSaving = false;
		}
	}

	async function handleDelete() {
		if (!project) return;
		try {
			await api.deleteProject(project.id);
			goto('/projects');
		} catch (err) {
			console.error('Error deleting project:', err);
		}
	}

	async function handleAddNote() {
		if (!project || !newNote.trim()) return;
		isAddingNote = true;
		try {
			await api.addProjectNote(project.id, newNote);
			await loadProject();
			newNote = '';
		} catch (err) {
			console.error('Error adding note:', err);
		} finally {
			isAddingNote = false;
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'active': return 'bg-emerald-50 text-emerald-700';
			case 'paused': return 'bg-amber-50 text-amber-700';
			case 'completed': return 'bg-blue-50 text-blue-700';
			case 'archived': return 'bg-gray-100 text-gray-600';
			default: return 'bg-gray-100 text-gray-600';
		}
	}

	function getPriorityColor(priority: string) {
		switch (priority) {
			case 'critical': return 'text-red-600';
			case 'high': return 'text-orange-600';
			case 'medium': return 'text-yellow-600';
			case 'low': return 'text-green-600';
			default: return 'text-gray-600';
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function formatTime(dateStr: string) {
		return new Date(dateStr).toLocaleTimeString(undefined, {
			hour: 'numeric',
			minute: '2-digit'
		});
	}
</script>

<div class="h-full flex flex-col">
	{#if isLoading}
		<div class="flex-1 flex items-center justify-center">
			<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
		</div>
	{:else if error || !project}
		<div class="flex-1 flex items-center justify-center">
			<div class="text-center">
				<p class="text-gray-500 mb-4">{error || 'Project not found'}</p>
				<a href="/projects" class="btn btn-secondary">Back to Projects</a>
			</div>
		</div>
	{:else}
		<!-- Header -->
		<div class="px-6 py-4 border-b border-gray-100">
			<div class="flex items-center gap-2 text-sm text-gray-500 mb-2">
				<a href="/projects" class="hover:text-gray-700">Projects</a>
				<span>/</span>
				<span class="text-gray-900">{project.name}</span>
			</div>
			<div class="flex items-start justify-between">
				<div>
					<div class="flex items-center gap-3 mb-1">
						<h1 class="text-xl font-semibold text-gray-900">{project.name}</h1>
						<span class="text-xs font-medium px-2.5 py-1 rounded-full {getStatusColor(project.status)}">
							{project.status}
						</span>
					</div>
					{#if project.client_name}
						<p class="text-sm text-gray-500">{project.client_name}</p>
					{/if}
				</div>
				<div class="flex gap-2">
					<button onclick={() => showEditDialog = true} class="btn btn-secondary text-sm">
						Edit
					</button>
					<button onclick={() => showDeleteConfirm = true} class="btn text-sm bg-red-50 text-red-600 hover:bg-red-100">
						Delete
					</button>
				</div>
			</div>
		</div>

		<!-- Content -->
		<div class="flex-1 overflow-y-auto p-6">
			<div class="max-w-4xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Main Content -->
				<div class="lg:col-span-2 space-y-6">
					<!-- Description -->
					<div class="card">
						<h2 class="text-lg font-medium text-gray-900 mb-3">Description</h2>
						{#if project.description}
							<p class="text-gray-600">{project.description}</p>
						{:else}
							<p class="text-gray-400 italic">No description</p>
						{/if}
					</div>

					<!-- Notes -->
					<div class="card">
						<h2 class="text-lg font-medium text-gray-900 mb-3">Notes</h2>

						<!-- Add Note -->
						<div class="mb-4">
							<textarea
								bind:value={newNote}
								placeholder="Add a note..."
								class="input input-square resize-none mb-2"
								rows="2"
							></textarea>
							<button
								onclick={handleAddNote}
								disabled={!newNote.trim() || isAddingNote}
								class="btn btn-primary text-sm disabled:opacity-50"
							>
								{isAddingNote ? 'Adding...' : 'Add Note'}
							</button>
						</div>

						<!-- Notes List -->
						{#if project.notes.length === 0}
							<p class="text-gray-400 text-sm">No notes yet</p>
						{:else}
							<div class="space-y-3">
								{#each project.notes as note}
									<div class="p-3 bg-gray-50 rounded-lg">
										<p class="text-sm text-gray-600">{note.content}</p>
										<p class="text-xs text-gray-400 mt-2">
											{formatDate(note.created_at)} at {formatTime(note.created_at)}
										</p>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				</div>

				<!-- Sidebar -->
				<div class="space-y-6">
					<!-- Details -->
					<div class="card">
						<h2 class="text-lg font-medium text-gray-900 mb-3">Details</h2>
						<dl class="space-y-3">
							<div>
								<dt class="text-xs text-gray-500 uppercase">Priority</dt>
								<dd class="text-sm font-medium {getPriorityColor(project.priority)}">{project.priority}</dd>
							</div>
							<div>
								<dt class="text-xs text-gray-500 uppercase">Type</dt>
								<dd class="text-sm text-gray-900">{project.project_type}</dd>
							</div>
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

					<!-- Quick Actions -->
					<div class="card">
						<h2 class="text-lg font-medium text-gray-900 mb-3">Quick Actions</h2>
						<div class="space-y-2">
							{#if project.status !== 'completed'}
								<button
									onclick={async () => {
										await api.updateProject(project!.id, { status: 'completed' });
										await loadProject();
									}}
									class="btn btn-secondary w-full text-sm justify-start"
								>
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									Mark Complete
								</button>
							{/if}
							{#if project.status === 'active'}
								<button
									onclick={async () => {
										await api.updateProject(project!.id, { status: 'paused' });
										await loadProject();
									}}
									class="btn btn-secondary w-full text-sm justify-start"
								>
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									Pause Project
								</button>
							{:else if project.status === 'paused'}
								<button
									onclick={async () => {
										await api.updateProject(project!.id, { status: 'active' });
										await loadProject();
									}}
									class="btn btn-secondary w-full text-sm justify-start"
								>
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									Resume Project
								</button>
							{/if}
							{#if project.status !== 'archived'}
								<button
									onclick={async () => {
										await api.updateProject(project!.id, { status: 'archived' });
										await loadProject();
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
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Edit Dialog -->
<Dialog.Root bind:open={showEditDialog}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-md z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">Edit Project</Dialog.Title>

			<form onsubmit={(e) => { e.preventDefault(); handleSave(); }} class="space-y-4">
				<div>
					<label for="edit-name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
					<input
						id="edit-name"
						type="text"
						bind:value={editForm.name}
						class="input input-square"
						required
					/>
				</div>

				<div>
					<label for="edit-client" class="block text-sm font-medium text-gray-700 mb-1">Client</label>
					<input
						id="edit-client"
						type="text"
						bind:value={editForm.client_name}
						class="input input-square"
					/>
				</div>

				<div>
					<label for="edit-description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
					<textarea
						id="edit-description"
						bind:value={editForm.description}
						class="input input-square resize-none"
						rows="3"
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="edit-status" class="block text-sm font-medium text-gray-700 mb-1">Status</label>
						<select id="edit-status" bind:value={editForm.status} class="input input-square">
							<option value="active">Active</option>
							<option value="paused">Paused</option>
							<option value="completed">Completed</option>
							<option value="archived">Archived</option>
						</select>
					</div>
					<div>
						<label for="edit-priority" class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
						<select id="edit-priority" bind:value={editForm.priority} class="input input-square">
							<option value="low">Low</option>
							<option value="medium">Medium</option>
							<option value="high">High</option>
							<option value="critical">Critical</option>
						</select>
					</div>
				</div>

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => showEditDialog = false} class="btn btn-secondary flex-1">
						Cancel
					</button>
					<button type="submit" disabled={isSaving} class="btn btn-primary flex-1">
						{isSaving ? 'Saving...' : 'Save Changes'}
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<!-- Delete Confirmation -->
<Dialog.Root bind:open={showDeleteConfirm}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-sm z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-2">Delete Project</Dialog.Title>
			<p class="text-sm text-gray-500 mb-6">
				Are you sure you want to delete "{project?.name}"? This action cannot be undone.
			</p>
			<div class="flex gap-3">
				<button onclick={() => showDeleteConfirm = false} class="btn btn-secondary flex-1">
					Cancel
				</button>
				<button onclick={handleDelete} class="btn flex-1 bg-red-600 text-white hover:bg-red-700">
					Delete
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
