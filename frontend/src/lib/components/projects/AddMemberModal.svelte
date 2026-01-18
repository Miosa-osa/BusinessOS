<script lang="ts">
	import { Dialog } from 'bits-ui';
	import type { ProjectRole } from '$lib/api/projects/types';
	import RoleSelector from './RoleSelector.svelte';
	import { UserPlus, X, AlertCircle } from 'lucide-svelte';

	interface Props {
		open?: boolean;
		workspaceId: string;
		onClose?: () => void;
		onAdd?: (data: { user_id: string; role: ProjectRole; workspace_id: string }) => void;
	}

	let { open = $bindable(false), workspaceId, onClose, onAdd }: Props = $props();

	let userId = $state('');
	let userEmail = $state('');
	let selectedRole = $state<ProjectRole>('viewer');
	let error = $state('');

	function handleSubmit() {
		// Validate
		error = '';

		if (!userId.trim() && !userEmail.trim()) {
			error = 'Please enter a user ID or email address';
			return;
		}

		// In a real implementation, you might want to look up the user by email
		// For now, we'll use the userId if provided, otherwise use email as userId
		const finalUserId = userId.trim() || userEmail.trim();

		onAdd?.({
			user_id: finalUserId,
			role: selectedRole,
			workspace_id: workspaceId
		});

		resetForm();
		open = false;
	}

	function resetForm() {
		userId = '';
		userEmail = '';
		selectedRole = 'viewer';
		error = '';
	}

	function handleClose() {
		resetForm();
		open = false;
		onClose?.();
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 bg-black/50 z-50 animate-in fade-in-0" />
		<Dialog.Content
			class="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-lg bg-white rounded-2xl shadow-xl animate-in fade-in-0 zoom-in-95"
		>
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-100">
				<div class="flex items-center gap-3">
					<div class="p-2 bg-blue-50 rounded-lg">
						<UserPlus class="w-5 h-5 text-blue-600" />
					</div>
					<Dialog.Title class="text-lg font-semibold text-gray-900">
						Add Project Member
					</Dialog.Title>
				</div>
				<Dialog.Close
					class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 transition-colors"
					onclick={handleClose}
				>
					<X class="w-5 h-5 text-gray-500" />
				</Dialog.Close>
			</div>

			<!-- Body -->
			<div class="px-6 py-4 space-y-4">
				{#if error}
					<div class="flex items-start gap-2 p-3 bg-red-50 border border-red-200 rounded-lg">
						<AlertCircle class="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
						<p class="text-sm text-red-800">{error}</p>
					</div>
				{/if}

				<!-- User ID -->
				<div>
					<label for="user-id" class="block text-sm font-medium text-gray-700 mb-1">
						User ID <span class="text-red-500">*</span>
					</label>
					<input
						id="user-id"
						type="text"
						bind:value={userId}
						placeholder="e.g., user_123abc or user@example.com"
						class="w-full px-4 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
					/>
					<p class="mt-1 text-xs text-gray-500">
						Enter the user's ID from your workspace or their email address
					</p>
				</div>

				<!-- Email (optional alternative) -->
				<div>
					<label for="user-email" class="block text-sm font-medium text-gray-700 mb-1">
						Or Email Address
					</label>
					<input
						id="user-email"
						type="email"
						bind:value={userEmail}
						placeholder="user@example.com"
						disabled={userId.trim().length > 0}
						class="w-full px-4 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all disabled:bg-gray-50 disabled:text-gray-400"
					/>
					<p class="mt-1 text-xs text-gray-500">
						Alternative: enter email if you don't know the user ID
					</p>
				</div>

				<!-- Role Selection -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-2">
						Role <span class="text-red-500">*</span>
					</label>
					<RoleSelector bind:value={selectedRole} />
					<p class="mt-2 text-xs text-gray-500">
						{#if selectedRole === 'lead'}
							Full project control - can manage members, edit, and delete
						{:else if selectedRole === 'contributor'}
							Can edit and contribute to the project
						{:else if selectedRole === 'reviewer'}
							Can review and comment on the project
						{:else}
							Read-only access to the project
						{/if}
					</p>
				</div>

				<!-- Info Box -->
				<div class="p-3 bg-blue-50 border border-blue-200 rounded-lg">
					<div class="flex items-start gap-2">
						<svg
							class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5"
							fill="currentColor"
							viewBox="0 0 20 20"
						>
							<path
								fill-rule="evenodd"
								d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
								clip-rule="evenodd"
							/>
						</svg>
						<div class="text-xs text-blue-800">
							<p class="font-medium mb-1">About Project Roles</p>
							<ul class="space-y-1 list-disc list-inside">
								<li>Lead: Full control (edit, delete, invite members)</li>
								<li>Contributor: Can edit project content</li>
								<li>Reviewer: Can review and comment</li>
								<li>Viewer: Read-only access</li>
							</ul>
						</div>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
				<button
					onclick={handleClose}
					class="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={handleSubmit}
					disabled={!userId.trim() && !userEmail.trim()}
					class="btn-pill btn-pill-primary flex items-center gap-2"
				>
					<UserPlus class="w-4 h-4" />
					Add Member
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
