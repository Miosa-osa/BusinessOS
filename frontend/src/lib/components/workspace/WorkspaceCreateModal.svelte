<script lang="ts">
	import { createWorkspace } from '$lib/api/workspaces';
	import { switchWorkspace, initializeWorkspaces } from '$lib/stores/workspaces';

	interface Props {
		show: boolean;
		onClose: () => void;
	}

	let { show, onClose }: Props = $props();

	let name = $state('');
	let description = $state('');
	let logoUrl = $state('');
	let planType = $state<'free' | 'starter' | 'professional' | 'enterprise'>('free');
	let loading = $state(false);
	let error = $state<string | null>(null);
	let validationErrors = $state<Record<string, string>>({});

	function resetForm() {
		name = '';
		description = '';
		logoUrl = '';
		planType = 'free';
		error = null;
		validationErrors = {};
	}

	function validateForm(): boolean {
		const errors: Record<string, string> = {};

		if (!name || name.trim().length === 0) {
			errors.name = 'Workspace name is required';
		} else if (name.trim().length < 3) {
			errors.name = 'Workspace name must be at least 3 characters';
		} else if (name.trim().length > 50) {
			errors.name = 'Workspace name must be less than 50 characters';
		}

		if (description && description.length > 500) {
			errors.description = 'Description must be less than 500 characters';
		}

		if (logoUrl && logoUrl.trim().length > 0) {
			try {
				new URL(logoUrl);
			} catch {
				errors.logoUrl = 'Please enter a valid URL';
			}
		}

		validationErrors = errors;
		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		loading = true;
		error = null;

		try {
			const workspace = await createWorkspace({
				name: name.trim(),
				description: description.trim() || undefined,
				plan_type: planType,
			});

			// Update workspace after creation if logo_url was provided
			if (logoUrl && logoUrl.trim().length > 0) {
				const { updateWorkspace } = await import('$lib/api/workspaces');
				await updateWorkspace(workspace.id, {
					logo_url: logoUrl.trim(),
				});
			}

			// Refresh workspaces list
			await initializeWorkspaces();

			// Switch to the new workspace
			await switchWorkspace(workspace.id);

			// Reset and close
			resetForm();
			onClose();
		} catch (err) {
			console.error('Failed to create workspace:', err);
			error = err instanceof Error ? err.message : 'Failed to create workspace';
		} finally {
			loading = false;
		}
	}

	function handleCancel() {
		resetForm();
		onClose();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			handleCancel();
		}
	}
</script>

{#if show}
	<div class="modal-overlay" onclick={handleCancel} onkeydown={handleKeydown}>
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h3 class="modal-title">Create Workspace</h3>
				<button class="close-btn" onclick={handleCancel} aria-label="Close">
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
				<form onsubmit={(e) => e.preventDefault()}>
					<!-- Workspace Name -->
					<div class="form-group">
						<label for="workspace-name" class="form-label">
							Workspace Name <span class="required">*</span>
						</label>
						<input
							id="workspace-name"
							type="text"
							class="form-input"
							class:error={validationErrors.name}
							bind:value={name}
							placeholder="My Workspace"
							maxlength="50"
							autofocus
						/>
						{#if validationErrors.name}
							<span class="error-text">{validationErrors.name}</span>
						{/if}
					</div>

					<!-- Description -->
					<div class="form-group">
						<label for="workspace-description" class="form-label">Description</label>
						<textarea
							id="workspace-description"
							class="form-textarea"
							class:error={validationErrors.description}
							bind:value={description}
							placeholder="A brief description of your workspace (optional)"
							rows="3"
							maxlength="500"
						></textarea>
						<div class="char-count">
							{description.length}/500
						</div>
						{#if validationErrors.description}
							<span class="error-text">{validationErrors.description}</span>
						{/if}
					</div>

					<!-- Logo URL -->
					<div class="form-group">
						<label for="workspace-logo" class="form-label">Logo URL</label>
						<input
							id="workspace-logo"
							type="text"
							class="form-input"
							class:error={validationErrors.logoUrl}
							bind:value={logoUrl}
							placeholder="https://example.com/logo.png (optional)"
						/>
						{#if validationErrors.logoUrl}
							<span class="error-text">{validationErrors.logoUrl}</span>
						{/if}
					</div>

					<!-- Plan Type -->
					<div class="form-group">
						<label class="form-label">Plan Type</label>
						<div class="plan-options">
							<button
								type="button"
								class="plan-option"
								class:selected={planType === 'free'}
								onclick={() => (planType = 'free')}
							>
								<div class="plan-content">
									<span class="plan-title">Free</span>
									<span class="plan-desc">Basic features</span>
								</div>
							</button>

							<button
								type="button"
								class="plan-option"
								class:selected={planType === 'starter'}
								onclick={() => (planType = 'starter')}
							>
								<div class="plan-content">
									<span class="plan-title">Starter</span>
									<span class="plan-desc">More members & storage</span>
								</div>
							</button>

							<button
								type="button"
								class="plan-option"
								class:selected={planType === 'professional'}
								onclick={() => (planType = 'professional')}
							>
								<div class="plan-content">
									<span class="plan-title">Professional</span>
									<span class="plan-desc">Advanced features</span>
								</div>
							</button>

							<button
								type="button"
								class="plan-option"
								class:selected={planType === 'enterprise'}
								onclick={() => (planType = 'enterprise')}
							>
								<div class="plan-content">
									<span class="plan-title">Enterprise</span>
									<span class="plan-desc">Unlimited access</span>
								</div>
							</button>
						</div>
					</div>
				</form>

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
				<button class="btn-pill btn-pill-secondary flex-1" onclick={handleCancel} disabled={loading}>
					Cancel
				</button>
				<button class="btn-pill btn-pill-primary flex-1 {loading ? 'btn-pill-loading' : ''}" onclick={handleSubmit} disabled={loading}>
					{#if loading}
						Creating...
					{:else}
						Create Workspace
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
		max-height: 90vh;
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
	}

	.form-group {
		margin-bottom: 20px;
	}

	.form-group:last-child {
		margin-bottom: 0;
	}

	.form-label {
		display: block;
		font-size: 12px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: var(--color-text-muted);
		margin-bottom: 8px;
	}

	:global(.dark) .form-label {
		color: #a1a1a6;
	}

	.required {
		color: #ef4444;
	}

	.form-input,
	.form-textarea {
		width: 100%;
		padding: 10px 12px;
		font-size: 14px;
		color: var(--color-text);
		background: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: 8px;
		transition: all 0.15s ease;
		font-family: inherit;
	}

	.form-input:focus,
	.form-textarea:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.form-input.error,
	.form-textarea.error {
		border-color: #ef4444;
	}

	.form-input.error:focus,
	.form-textarea.error:focus {
		box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
	}

	:global(.dark) .form-input,
	:global(.dark) .form-textarea {
		color: #f5f5f7;
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .form-input:focus,
	:global(.dark) .form-textarea:focus {
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
	}

	.form-textarea {
		resize: vertical;
		min-height: 80px;
	}

	.char-count {
		text-align: right;
		font-size: 11px;
		color: var(--color-text-muted);
		margin-top: 4px;
	}

	:global(.dark) .char-count {
		color: #6e6e73;
	}

	.error-text {
		display: block;
		font-size: 12px;
		color: #ef4444;
		margin-top: 4px;
	}

	.plan-options {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 8px;
	}

	.plan-option {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 12px;
		background: transparent;
		border: 1px solid var(--color-border);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.15s ease;
		text-align: center;
	}

	.plan-option:hover {
		background: var(--color-bg-secondary);
		border-color: var(--color-border);
	}

	.plan-option.selected {
		background: rgba(59, 130, 246, 0.1);
		border-color: #3b82f6;
	}

	:global(.dark) .plan-option {
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .plan-option:hover {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.15);
	}

	:global(.dark) .plan-option.selected {
		background: rgba(59, 130, 246, 0.15);
		border-color: #3b82f6;
	}

	.plan-content {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.plan-title {
		font-size: 13px;
		font-weight: 600;
		color: var(--color-text);
	}

	.plan-option.selected .plan-title {
		color: #3b82f6;
	}

	:global(.dark) .plan-title {
		color: #f5f5f7;
	}

	.plan-desc {
		font-size: 11px;
		color: var(--color-text-muted);
	}

	:global(.dark) .plan-desc {
		color: #a1a1a6;
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
		margin-top: 16px;
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
