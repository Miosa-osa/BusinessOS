<script lang="ts">
	import { onMount } from 'svelte';
	import { projects } from '$lib/stores/projects';
	import { Dialog } from 'bits-ui';

	let showNewProject = $state(false);
	let newProject = $state({
		name: '',
		description: '',
		client_name: '',
		project_type: 'internal',
		priority: 'medium' as const
	});
	let statusFilter = $state('');

	onMount(() => {
		projects.loadProjects();
	});

	let createError = $state('');

	async function handleCreateProject(e: Event) {
		e.preventDefault();
		createError = '';
		try {
			await projects.createProject(newProject);
			showNewProject = false;
			newProject = { name: '', description: '', client_name: '', project_type: 'internal', priority: 'medium' };
		} catch (error) {
			createError = (error as Error).message || 'Failed to create project';
			console.error('Failed to create project:', error);
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
		return new Date(dateStr).toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
	}
</script>

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
		<div>
			<h1 class="text-xl font-semibold text-gray-900">Projects</h1>
			<p class="text-sm text-gray-500 mt-0.5">Manage your work and track progress</p>
		</div>
		<button onclick={() => showNewProject = true} class="btn btn-primary">
			+ New Project
		</button>
	</div>

	<!-- Filters -->
	<div class="px-6 py-3 border-b border-gray-100 flex gap-2">
		<button
			onclick={() => { statusFilter = ''; projects.loadProjects(); }}
			class="btn {statusFilter === '' ? 'btn-primary' : 'btn-secondary'} text-xs px-3 py-1.5"
		>
			All
		</button>
		<button
			onclick={() => { statusFilter = 'active'; projects.loadProjects('active'); }}
			class="btn {statusFilter === 'active' ? 'btn-primary' : 'btn-secondary'} text-xs px-3 py-1.5"
		>
			Active
		</button>
		<button
			onclick={() => { statusFilter = 'paused'; projects.loadProjects('paused'); }}
			class="btn {statusFilter === 'paused' ? 'btn-primary' : 'btn-secondary'} text-xs px-3 py-1.5"
		>
			Paused
		</button>
		<button
			onclick={() => { statusFilter = 'completed'; projects.loadProjects('completed'); }}
			class="btn {statusFilter === 'completed' ? 'btn-primary' : 'btn-secondary'} text-xs px-3 py-1.5"
		>
			Completed
		</button>
	</div>

	<!-- Projects Grid -->
	<div class="flex-1 overflow-y-auto p-6">
		{#if $projects.loading}
			<div class="flex items-center justify-center h-48">
				<div class="animate-spin h-8 w-8 border-2 border-gray-900 border-t-transparent rounded-full"></div>
			</div>
		{:else if $projects.projects.length === 0}
			<div class="flex flex-col items-center justify-center h-48 text-center">
				<div class="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-3">
					<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
					</svg>
				</div>
				<p class="text-gray-500">No projects yet</p>
				<button onclick={() => showNewProject = true} class="btn btn-secondary text-sm mt-3">
					Create your first project
				</button>
			</div>
		{:else}
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each $projects.projects as project}
					<a href="/projects/{project.id}" class="card hover:shadow-md transition-shadow duration-200 cursor-pointer group">
						<div class="flex items-start justify-between mb-3">
							<span class="text-xs font-medium px-2.5 py-1 rounded-full {getStatusColor(project.status)}">
								{project.status}
							</span>
							<span class="text-xs {getPriorityColor(project.priority)} font-medium">
								{project.priority}
							</span>
						</div>
						<h3 class="font-medium text-gray-900 group-hover:text-gray-700 mb-1">{project.name}</h3>
						{#if project.client_name}
							<p class="text-sm text-gray-500 mb-2">{project.client_name}</p>
						{/if}
						{#if project.description}
							<p class="text-sm text-gray-500 line-clamp-2">{project.description}</p>
						{/if}
						<div class="mt-4 pt-3 border-t border-gray-100 flex items-center justify-between text-xs text-gray-400">
							<span>{project.project_type}</span>
							<span>Updated {formatDate(project.updated_at)}</span>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- New Project Dialog -->
<Dialog.Root bind:open={showNewProject}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
		<Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white rounded-2xl shadow-xl p-6 w-full max-w-md z-50">
			<Dialog.Title class="text-lg font-semibold text-gray-900 mb-4">New Project</Dialog.Title>

			<form onsubmit={handleCreateProject} class="space-y-4">
				<div>
					<label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
					<input
						id="name"
						type="text"
						bind:value={newProject.name}
						class="input input-square"
						placeholder="Project name"
						required
					/>
				</div>

				<div>
					<label for="client" class="block text-sm font-medium text-gray-700 mb-1">Client (optional)</label>
					<input
						id="client"
						type="text"
						bind:value={newProject.client_name}
						class="input input-square"
						placeholder="Client name"
					/>
				</div>

				<div>
					<label for="description" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
					<textarea
						id="description"
						bind:value={newProject.description}
						class="input input-square resize-none"
						rows="3"
						placeholder="What's this project about?"
					></textarea>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="type" class="block text-sm font-medium text-gray-700 mb-1">Type</label>
						<select id="type" bind:value={newProject.project_type} class="input input-square">
							<option value="internal">Internal</option>
							<option value="client_work">Client Work</option>
							<option value="learning">Learning</option>
						</select>
					</div>
					<div>
						<label for="priority" class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
						<select id="priority" bind:value={newProject.priority} class="input input-square">
							<option value="low">Low</option>
							<option value="medium">Medium</option>
							<option value="high">High</option>
							<option value="critical">Critical</option>
						</select>
					</div>
				</div>

				{#if createError}
					<div class="text-sm text-red-600 bg-red-50 px-3 py-2 rounded-xl">
						{createError}
					</div>
				{/if}

				<div class="flex gap-3 pt-2">
					<button type="button" onclick={() => showNewProject = false} class="btn btn-secondary flex-1">
						Cancel
					</button>
					<button type="submit" class="btn btn-primary flex-1">
						Create Project
					</button>
				</div>
			</form>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
