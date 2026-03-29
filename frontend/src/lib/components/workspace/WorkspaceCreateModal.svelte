<script lang="ts">
	import { createWorkspace } from '$lib/api/workspaces';
	import { switchWorkspace, initializeWorkspaces } from '$lib/stores/workspaces';
	import { X, AlertCircle } from 'lucide-svelte';

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
	<div class="bos-modal-overlay" onclick={handleCancel} onkeydown={handleKeydown}>
		<div class="bos-modal bos-modal--md" onclick={(e) => e.stopPropagation()}>
			<div class="bos-modal-header">
				<h3 class="bos-modal-title">Create Workspace</h3>
				<button class="bos-modal-close" onclick={handleCancel} aria-label="Close">
					<X size={18} />
				</button>
			</div>

			<div class="bos-modal-body">
				<form onsubmit={(e) => e.preventDefault()}>
					<!-- Workspace Name -->
					<div class="bos-form-group">
						<label for="workspace-name" class="bos-label bos-label--required">
							Workspace Name
						</label>
						<input
							id="workspace-name"
							type="text"
							class="bos-input"
							class:bos-input--error={validationErrors.name}
							bind:value={name}
							placeholder="My Workspace"
							maxlength="50"
							autofocus
						/>
						{#if validationErrors.name}
							<span class="bos-field-error">{validationErrors.name}</span>
						{/if}
					</div>

					<!-- Description -->
					<div class="bos-form-group">
						<label for="workspace-description" class="bos-label">Description</label>
						<textarea
							id="workspace-description"
							class="bos-textarea"
							class:bos-textarea--error={validationErrors.description}
							bind:value={description}
							placeholder="A brief description of your workspace (optional)"
							rows="3"
							maxlength="500"
						></textarea>
						<div class="bos-char-count">
							{description.length}/500
						</div>
						{#if validationErrors.description}
							<span class="bos-field-error">{validationErrors.description}</span>
						{/if}
					</div>

					<!-- Logo URL -->
					<div class="bos-form-group">
						<label for="workspace-logo" class="bos-label">Logo URL</label>
						<input
							id="workspace-logo"
							type="text"
							class="bos-input"
							class:bos-input--error={validationErrors.logoUrl}
							bind:value={logoUrl}
							placeholder="https://example.com/logo.png (optional)"
						/>
						{#if validationErrors.logoUrl}
							<span class="bos-field-error">{validationErrors.logoUrl}</span>
						{/if}
					</div>

					<!-- Plan Type -->
					<div class="bos-form-group">
						<label class="bos-label">Plan Type</label>
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
					<div class="bos-error-banner">
						<AlertCircle size={16} />
						{error}
					</div>
				{/if}
			</div>

			<div class="bos-modal-footer">
				<button class="btn-pill btn-pill-ghost btn-pill-sm" onclick={handleCancel} disabled={loading}>
					Cancel
				</button>
				<button class="btn-pill btn-pill-primary btn-pill-sm" onclick={handleSubmit} disabled={loading}>
					{#if loading}
						<span class="btn-spinner"></span>
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
	/* Plan options grid */
	.plan-options {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 8px;
	}

	/* Plan option button — token-based, no dark overrides needed */
	.plan-option {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 12px;
		background: transparent;
		border: 1px solid var(--bos-border-color);
		border-radius: var(--radius-sm);
		cursor: pointer;
		transition: background var(--bos-transition-fast), border-color var(--bos-transition-fast);
		text-align: center;
		font-family: inherit;
	}

	.plan-option:hover {
		background: var(--bos-hover-color);
	}

	.plan-option.selected {
		background: color-mix(in srgb, var(--bos-primary-color) 10%, var(--bos-modal-bg));
		border-color: var(--bos-primary-color);
	}

	.plan-content {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.plan-title {
		font-size: 13px;
		font-weight: 600;
		color: var(--bos-text-primary-color);
	}

	.plan-option.selected .plan-title {
		color: var(--bos-primary-color);
	}

	.plan-desc {
		font-size: 11px;
		color: var(--bos-text-secondary-color);
	}

	/* Loading spinner (used inside btn-pill-primary) */
	.btn-spinner {
		width: 14px;
		height: 14px;
		border: 2px solid color-mix(in srgb, var(--bos-surface-on-color) 30%, transparent);
		border-top-color: var(--bos-surface-on-color);
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
		flex-shrink: 0;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}
</style>
