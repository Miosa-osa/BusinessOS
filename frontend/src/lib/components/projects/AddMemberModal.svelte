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
			class="prm-dialog fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-lg rounded-2xl shadow-xl animate-in fade-in-0 zoom-in-95"
		>
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 prm-dialog__border-b">
				<div class="flex items-center gap-3">
					<div class="p-2 prm-dialog__icon-bg rounded-lg">
						<UserPlus class="w-5 h-5 prm-dialog__accent-icon" />
					</div>
					<Dialog.Title class="prm-dialog__title" style="margin-bottom:0">
						Add Project Member
					</Dialog.Title>
				</div>
				<Dialog.Close
					class="btn-pill btn-pill-ghost btn-pill-icon"
					onclick={handleClose}
				>
					<X class="w-5 h-5 prm-dialog__icon-muted" />
				</Dialog.Close>
			</div>

			<!-- Body -->
			<div class="px-6 py-4 space-y-4">
				{#if error}
					<div class="flex items-start gap-2 p-3 prm-dialog__error rounded-lg">
						<AlertCircle class="w-5 h-5 prm-dialog__error-icon flex-shrink-0 mt-0.5" />
						<p class="text-sm prm-dialog__error-text">{error}</p>
					</div>
				{/if}

				<!-- User ID -->
				<div>
					<label for="user-id" class="prm-dialog__label">
						User ID <span class="text-red-500">*</span>
					</label>
					<input
						id="user-id"
						type="text"
						bind:value={userId}
						placeholder="e.g., user_123abc or user@example.com"
						class="prm-dialog__input"
					/>
					<p class="prm-dialog__hint">
						Enter the user's ID from your workspace or their email address
					</p>
				</div>

				<!-- Email (optional alternative) -->
				<div>
					<label for="user-email" class="prm-dialog__label">
						Or Email Address
					</label>
					<input
						id="user-email"
						type="email"
						bind:value={userEmail}
						placeholder="user@example.com"
						disabled={userId.trim().length > 0}
						class="prm-dialog__input"
					/>
					<p class="prm-dialog__hint">
						Alternative: enter email if you don't know the user ID
					</p>
				</div>

				<!-- Role Selection -->
				<div>
					<label class="prm-dialog__label" style="margin-bottom:0.5rem">
						Role <span class="text-red-500">*</span>
					</label>
					<RoleSelector bind:value={selectedRole} />
					<p class="prm-dialog__hint" style="margin-top:0.5rem">
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
				<div class="p-3 prm-dialog__info-box rounded-lg">
					<div class="flex items-start gap-2">
						<svg
							class="w-5 h-5 prm-dialog__accent-icon flex-shrink-0 mt-0.5"
							fill="currentColor"
							viewBox="0 0 20 20"
						>
							<path
								fill-rule="evenodd"
								d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
								clip-rule="evenodd"
							/>
						</svg>
						<div class="text-xs prm-dialog__info-text">
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
			<div class="flex items-center justify-end gap-3 px-6 py-4 prm-dialog__border-t">
				<button
					onclick={handleClose}
					class="btn-pill btn-pill-ghost btn-pill-sm"
				>
					Cancel
				</button>
				<button
					onclick={handleSubmit}
					disabled={!userId.trim() && !userEmail.trim()}
					class="btn-pill btn-pill-primary btn-pill-sm flex items-center gap-2"
				>
					<UserPlus class="w-4 h-4" />
					Add Member
				</button>
			</div>
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
	.prm-dialog__input {
		width: 100%;
		padding: 0.625rem 1rem;
		font-size: 0.875rem;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 0.75rem;
		background: var(--dbg, #fff);
		color: var(--dt, #111);
		transition: all 0.15s;
	}
	.prm-dialog__input:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.25);
	}
	.prm-dialog__input:disabled {
		background: var(--dbg2, #f5f5f5);
		color: var(--dt3, #888);
	}
	.prm-dialog__hint {
		margin-top: 0.25rem;
		font-size: 0.75rem;
		color: var(--dt2, #555);
	}
	.prm-dialog__icon-muted {
		color: var(--dt2, #555);
	}
	.prm-dialog__border-b {
		border-bottom: 1px solid var(--dbd2, #f0f0f0);
	}
	.prm-dialog__border-t {
		border-top: 1px solid var(--dbd2, #f0f0f0);
	}
	.prm-dialog__icon-bg { background: color-mix(in srgb, #3b82f6 12%, var(--dbg)); }
	.prm-dialog__accent-icon { color: #3b82f6; }
	.prm-dialog__error { background: color-mix(in srgb, #ef4444 10%, var(--dbg)); border: 1px solid color-mix(in srgb, #ef4444 25%, var(--dbd)); }
	.prm-dialog__error-icon { color: #ef4444; }
	.prm-dialog__error-text { color: #ef4444; }
	.prm-dialog__info-box { background: color-mix(in srgb, #3b82f6 10%, var(--dbg)); border: 1px solid color-mix(in srgb, #3b82f6 25%, var(--dbd)); }
	.prm-dialog__info-text { color: var(--dt2, #555); }
</style>
