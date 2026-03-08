<script lang="ts">
	import { currentWorkspaceId, currentWorkspaceMembers } from '$lib/stores/workspaces';
	import type { WorkspaceMemoryListItem } from '$lib/api/workspaces/memory';
	import { shareMemory, unshareMemory, updateWorkspaceMemory } from '$lib/api/workspaces/memory';

	interface Props {
		memory: WorkspaceMemoryListItem | null;
		onClose: () => void;
		onComplete: () => void;
	}

	let { memory, onClose, onComplete }: Props = $props();

	let selectedUserIds = $state<string[]>([]);
	let newVisibility = $state<'private' | 'workspace' | 'shared'>('private');
	let loading = $state(false);
	let error = $state<string | null>(null);

	// Initialize selected users when memory changes
	$effect(() => {
		if (memory) {
			selectedUserIds = memory.shared_with_user_ids || [];
			newVisibility = memory.visibility;
		}
	});

	// Filter out current user from members list
	let availableMembers = $derived(() => {
		return $currentWorkspaceMembers.filter((member) => {
			// Filter out the memory creator (they always have access)
			return member.user_id !== memory?.created_by;
		});
	});

	function toggleUser(userId: string) {
		if (selectedUserIds.includes(userId)) {
			selectedUserIds = selectedUserIds.filter((id) => id !== userId);
		} else {
			selectedUserIds = [...selectedUserIds, userId];
		}
	}

	async function handleSave() {
		if (!memory || !$currentWorkspaceId) return;

		loading = true;
		error = null;

		try {
			// First, update visibility if changed
			if (newVisibility !== memory.visibility) {
				await updateWorkspaceMemory($currentWorkspaceId, memory.id, {
					visibility: newVisibility
				});
			}

			// Then handle sharing changes
			if (newVisibility === 'shared') {
				// Determine which users to add and remove
				const currentShared = memory.shared_with_user_ids || [];
				const toAdd = selectedUserIds.filter((id) => !currentShared.includes(id));
				const toRemove = currentShared.filter((id) => !selectedUserIds.includes(id));

				// Add new shares
				if (toAdd.length > 0) {
					await shareMemory($currentWorkspaceId, memory.id, { user_ids: toAdd });
				}

				// Remove shares
				if (toRemove.length > 0) {
					await unshareMemory($currentWorkspaceId, memory.id, { user_ids: toRemove });
				}
			}

			onComplete();
			onClose();
		} catch (err) {
			console.error('Failed to update memory sharing:', err);
			error = err instanceof Error ? err.message : 'Failed to update sharing settings';
		} finally {
			loading = false;
		}
	}

	function handleCancel() {
		onClose();
	}
</script>

{#if memory}
	<div class="modal-overlay" onclick={handleCancel}>
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h3 class="modal-title">Share Memory</h3>
				<button class="btn-pill btn-pill-ghost close-btn" onclick={handleCancel} aria-label="Close">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						width="20"
						height="20"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="modal-body">
				<div class="memory-info">
					<h4 class="memory-title">{memory.title}</h4>
					<p class="memory-summary">{memory.summary}</p>
				</div>

				<div class="visibility-section">
					<label class="section-label">Visibility</label>
					<div class="visibility-options">
						<button
							class="btn-pill btn-pill-ghost visibility-option"
							class:selected={newVisibility === 'private'}
							onclick={() => (newVisibility = 'private')}
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
								stroke-width="1.5"
								stroke="currentColor"
								width="16"
								height="16"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z"
								/>
							</svg>
							<div class="option-text">
								<span class="option-title">Private</span>
								<span class="option-desc">Only you can see this</span>
							</div>
						</button>

						<button
							class="btn-pill btn-pill-ghost visibility-option"
							class:selected={newVisibility === 'workspace'}
							onclick={() => (newVisibility = 'workspace')}
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
								stroke-width="1.5"
								stroke="currentColor"
								width="16"
								height="16"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M18 18.72a9.094 9.094 0 0 0 3.741-.479 3 3 0 0 0-4.682-2.72m.94 3.198.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0 1 12 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 0 1 6 18.719m12 0a5.971 5.971 0 0 0-.941-3.197m0 0A5.995 5.995 0 0 0 12 12.75a5.995 5.995 0 0 0-5.058 2.772m0 0a3 3 0 0 0-4.681 2.72 8.986 8.986 0 0 0 3.74.477m.94-3.197a5.971 5.971 0 0 0-.94 3.197M15 6.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm6 3a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Zm-13.5 0a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Z"
								/>
							</svg>
							<div class="option-text">
								<span class="option-title">Workspace</span>
								<span class="option-desc">All workspace members</span>
							</div>
						</button>

						<button
							class="btn-pill btn-pill-ghost visibility-option"
							class:selected={newVisibility === 'shared'}
							onclick={() => (newVisibility = 'shared')}
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
								stroke-width="1.5"
								stroke="currentColor"
								width="16"
								height="16"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0 0a2.25 2.25 0 1 0 3.935 2.186 2.25 2.25 0 0 0-3.935-2.186Zm0-12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z"
								/>
							</svg>
							<div class="option-text">
								<span class="option-title">Shared</span>
								<span class="option-desc">Specific team members</span>
							</div>
						</button>
					</div>
				</div>

				{#if newVisibility === 'shared'}
					<div class="members-section">
						<label class="section-label"
							>Share with ({selectedUserIds.length} selected)</label
						>
						<div class="members-list">
							{#each availableMembers() as member}
								<button
									class="btn-pill btn-pill-ghost member-item"
									class:selected={selectedUserIds.includes(member.user_id)}
									onclick={() => toggleUser(member.user_id)}
								>
									<div class="member-info">
										<div class="member-avatar">
											{member.user_id.charAt(0).toUpperCase()}
										</div>
										<div class="member-details">
											<span class="member-name">{member.user_id}</span>
											<span class="member-role">{member.role}</span>
										</div>
									</div>
									<div class="member-checkbox">
										{#if selectedUserIds.includes(member.user_id)}
											<svg
												xmlns="http://www.w3.org/2000/svg"
												fill="none"
												viewBox="0 0 24 24"
												stroke-width="2"
												stroke="currentColor"
												width="16"
												height="16"
											>
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													d="m4.5 12.75 6 6 9-13.5"
												/>
											</svg>
										{/if}
									</div>
								</button>
							{/each}
						</div>
					</div>
				{/if}

				{#if error}
					<div class="error-message">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							fill="none"
							viewBox="0 0 24 24"
							stroke-width="1.5"
							stroke="currentColor"
							width="16"
							height="16"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"
							/>
						</svg>
						{error}
					</div>
				{/if}
			</div>

			<div class="modal-footer">
				<button class="btn-pill btn-pill-ghost btn btn-secondary" onclick={handleCancel} disabled={loading}>
					Cancel
				</button>
				<button class="btn-pill btn-pill-ghost btn btn-primary" onclick={handleSave} disabled={loading}>
					{#if loading}
						<span class="btn-spinner"></span>
						Saving...
					{:else}
						Save Changes
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 16px;
	}

	:global(.dark) .modal-overlay {
		background: rgba(0, 0, 0, 0.7);
	}

	.modal-content {
		background: var(--color-bg);
		border-radius: 12px;
		width: 100%;
		max-width: 500px;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
	}

	:global(.dark) .modal-content {
		background: #1c1c1e;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3), 0 10px 10px -5px rgba(0, 0, 0, 0.2);
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20px;
		border-bottom: 1px solid var(--color-border);
	}

	:global(.dark) .modal-header {
		border-bottom-color: rgba(255, 255, 255, 0.1);
	}

	.modal-title {
		font-size: 16px;
		font-weight: 600;
		color: var(--color-text);
		margin: 0;
	}

	:global(.dark) .modal-title {
		color: #f5f5f7;
	}

	.close-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border: none;
		background: transparent;
		color: var(--color-text-muted);
		cursor: pointer;
		border-radius: 6px;
		transition: all 0.15s ease;
	}

	.close-btn:hover {
		background: var(--color-bg-secondary);
		color: var(--color-text);
	}

	:global(.dark) .close-btn {
		color: #6e6e73;
	}

	:global(.dark) .close-btn:hover {
		background: #3a3a3c;
		color: #f5f5f7;
	}

	.modal-body {
		flex: 1;
		overflow-y: auto;
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.memory-info {
		padding: 12px;
		background: var(--color-bg-secondary);
		border-radius: 8px;
	}

	:global(.dark) .memory-info {
		background: #2c2c2e;
	}

	.memory-title {
		font-size: 13px;
		font-weight: 600;
		color: var(--color-text);
		margin: 0 0 4px 0;
	}

	:global(.dark) .memory-title {
		color: #f5f5f7;
	}

	.memory-summary {
		font-size: 12px;
		color: var(--color-text-muted);
		margin: 0;
		line-height: 1.5;
	}

	:global(.dark) .memory-summary {
		color: #a1a1a6;
	}

	.visibility-section,
	.members-section {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.section-label {
		font-size: 12px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--color-text-muted);
	}

	:global(.dark) .section-label {
		color: #a1a1a6;
	}

	.visibility-options {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.visibility-option {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		background: transparent;
		border: 1px solid var(--color-border);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.15s ease;
		text-align: left;
	}

	.visibility-option:hover {
		background: var(--color-bg-secondary);
		border-color: var(--color-border);
	}

	.visibility-option.selected {
		background: rgba(59, 130, 246, 0.1);
		border-color: #3b82f6;
	}

	:global(.dark) .visibility-option {
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .visibility-option:hover {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.15);
	}

	:global(.dark) .visibility-option.selected {
		background: rgba(59, 130, 246, 0.15);
		border-color: #3b82f6;
	}

	.visibility-option svg {
		color: var(--color-text-muted);
		flex-shrink: 0;
	}

	.visibility-option.selected svg {
		color: #3b82f6;
	}

	:global(.dark) .visibility-option svg {
		color: #6e6e73;
	}

	:global(.dark) .visibility-option.selected svg {
		color: #3b82f6;
	}

	.option-text {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.option-title {
		font-size: 13px;
		font-weight: 600;
		color: var(--color-text);
	}

	.visibility-option.selected .option-title {
		color: #3b82f6;
	}

	:global(.dark) .option-title {
		color: #f5f5f7;
	}

	.option-desc {
		font-size: 11px;
		color: var(--color-text-muted);
	}

	:global(.dark) .option-desc {
		color: #a1a1a6;
	}

	.members-list {
		display: flex;
		flex-direction: column;
		gap: 6px;
		max-height: 300px;
		overflow-y: auto;
	}

	.member-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px;
		background: transparent;
		border: 1px solid var(--color-border);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.member-item:hover {
		background: var(--color-bg-secondary);
	}

	.member-item.selected {
		background: rgba(59, 130, 246, 0.1);
		border-color: #3b82f6;
	}

	:global(.dark) .member-item {
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .member-item:hover {
		background: #2c2c2e;
	}

	:global(.dark) .member-item.selected {
		background: rgba(59, 130, 246, 0.15);
		border-color: #3b82f6;
	}

	.member-info {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.member-avatar {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: var(--color-bg-tertiary);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
		font-weight: 600;
		color: var(--color-text);
	}

	:global(.dark) .member-avatar {
		background: #3a3a3c;
		color: #f5f5f7;
	}

	.member-details {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.member-name {
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text);
	}

	:global(.dark) .member-name {
		color: #f5f5f7;
	}

	.member-role {
		font-size: 11px;
		color: var(--color-text-muted);
		text-transform: capitalize;
	}

	:global(.dark) .member-role {
		color: #a1a1a6;
	}

	.member-checkbox {
		width: 20px;
		height: 20px;
		border: 2px solid var(--color-border);
		border-radius: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.15s ease;
	}

	.member-item.selected .member-checkbox {
		background: #3b82f6;
		border-color: #3b82f6;
	}

	.member-item.selected .member-checkbox svg {
		color: white;
	}

	:global(.dark) .member-checkbox {
		border-color: rgba(255, 255, 255, 0.2);
	}

	.error-message {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px;
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
		border-radius: 8px;
		font-size: 12px;
	}

	:global(.dark) .error-message {
		background: rgba(239, 68, 68, 0.15);
	}

	.modal-footer {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 12px;
		padding: 16px 20px;
		border-top: 1px solid var(--color-border);
	}

	:global(.dark) .modal-footer {
		border-top-color: rgba(255, 255, 255, 0.1);
	}

	.btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		padding: 8px 16px;
		font-size: 13px;
		font-weight: 500;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-secondary {
		color: var(--color-text);
		background: var(--color-bg-secondary);
	}

	.btn-secondary:hover:not(:disabled) {
		background: var(--color-bg-tertiary);
	}

	:global(.dark) .btn-secondary {
		color: #f5f5f7;
		background: #3a3a3c;
	}

	:global(.dark) .btn-secondary:hover:not(:disabled) {
		background: #4a4a4c;
	}

	.btn-primary {
		color: white;
		background: #3b82f6;
	}

	.btn-primary:hover:not(:disabled) {
		background: #2563eb;
	}

	.btn-spinner {
		width: 14px;
		height: 14px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
