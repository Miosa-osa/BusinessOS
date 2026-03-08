<script lang="ts">
	interface Props {
		show: boolean;
		newProjectName: string;
		creatingProject: boolean;
		onClose: () => void;
		onNameChange: (name: string) => void;
		onCreate: () => void;
	}

	let {
		show,
		newProjectName = $bindable(),
		creatingProject,
		onClose,
		onNameChange,
		onCreate,
	}: Props = $props();
</script>

{#if show}
	<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
			<div class="p-5 border-b border-gray-100">
				<h3 class="text-lg font-semibold text-gray-900">Create New Project</h3>
				<p class="text-sm text-gray-500 mt-1">Give your project a name to get started</p>
			</div>
			<div class="p-5">
				<input
					type="text"
					bind:value={newProjectName}
					placeholder="Project name..."
					class="w-full px-4 py-3 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
					onkeydown={(e) => { if (e.key === 'Enter') onCreate(); if (e.key === 'Escape') onClose(); }}
					autofocus
				/>
			</div>
			<div class="p-4 border-t border-gray-100 flex justify-end gap-3">
				<button
					onclick={onClose}
					class="btn-pill btn-pill-soft btn-pill-sm"
				>
					Cancel
				</button>
				<button
					onclick={onCreate}
					disabled={!newProjectName.trim() || creatingProject}
					class="btn-pill btn-pill-primary btn-pill-sm {creatingProject ? 'btn-pill-loading' : ''}"
				>
					Create & Start Chat
				</button>
			</div>
		</div>
	</div>
{/if}
