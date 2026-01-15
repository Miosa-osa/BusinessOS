<script lang="ts">
	import { Download, Check, Loader2, AlertCircle } from 'lucide-svelte';
	import { installModule } from '$lib/api/osa/files';
	import type { OSAWorkflow } from './types';

	interface Props {
		workflow: OSAWorkflow;
		selectedFileIds?: string[];
		disabled?: boolean;
		onSuccess?: (moduleId: string) => void;
		onError?: (error: string) => void;
	}

	let {
		workflow,
		selectedFileIds = [],
		disabled = false,
		onSuccess,
		onError
	}: Props = $props();

	let installing = $state(false);
	let installed = $state(false);
	let error = $state<string | null>(null);
	let showOptions = $state(false);

	// Install options
	let moduleName = $state(workflow.name);
	let installPath = $state('');

	async function handleInstall() {
		if (installing) return;

		installing = true;
		error = null;

		try {
			const response = await installModule(workflow.id, {
				module_name: moduleName.trim() || workflow.name,
				install_path: installPath.trim() || undefined,
				file_ids: selectedFileIds.length > 0 ? selectedFileIds : undefined
			});

			if (response.success) {
				installed = true;
				if (onSuccess) {
					onSuccess(response.module_id);
				}

				// Reset after 3 seconds
				setTimeout(() => {
					installed = false;
					showOptions = false;
				}, 3000);
			} else {
				throw new Error(response.message || 'Installation failed');
			}
		} catch (err: any) {
			console.error('Failed to install module:', err);
			error = err?.message || 'Failed to install module';
			if (onError) {
				onError(error);
			}
		} finally {
			installing = false;
		}
	}

	function toggleOptions() {
		showOptions = !showOptions;
		error = null;
	}

	function handleQuickInstall() {
		showOptions = false;
		handleInstall();
	}
</script>

<div class="install-button-wrapper">
	{#if showOptions}
		<!-- Installation options form -->
		<div class="install-options">
			<div class="options-header">
				<h4>Installation Options</h4>
				<button class="close-btn" onclick={toggleOptions}>×</button>
			</div>

			<div class="option-group">
				<label for="module-name">Module Name</label>
				<input
					id="module-name"
					type="text"
					bind:value={moduleName}
					placeholder={workflow.name}
					disabled={installing}
				/>
			</div>

			<div class="option-group">
				<label for="install-path">
					Install Path <span class="optional">(optional)</span>
				</label>
				<input
					id="install-path"
					type="text"
					bind:value={installPath}
					placeholder="./modules or leave empty for default"
					disabled={installing}
				/>
			</div>

			{#if selectedFileIds.length > 0}
				<p class="info-text">
					Installing {selectedFileIds.length} selected {selectedFileIds.length === 1
						? 'file'
						: 'files'}
				</p>
			{:else}
				<p class="info-text">Installing all files from this workflow</p>
			{/if}

			{#if error}
				<div class="error-message">
					<AlertCircle size={16} />
					<span>{error}</span>
				</div>
			{/if}

			<div class="option-actions">
				<button class="cancel-btn" onclick={toggleOptions} disabled={installing}>
					Cancel
				</button>
				<button class="install-btn" onclick={handleInstall} disabled={installing}>
					{#if installing}
						<Loader2 size={16} class="animate-spin" />
						<span>Installing...</span>
					{:else}
						<Download size={16} />
						<span>Install Module</span>
					{/if}
				</button>
			</div>
		</div>
	{:else}
		<!-- Quick install button -->
		<button
			class="quick-install-btn"
			class:installed
			onclick={handleQuickInstall}
			disabled={disabled || installing || installed}
		>
			{#if installed}
				<Check size={18} class="text-green-400" />
				<span>Installed!</span>
			{:else if installing}
				<Loader2 size={18} class="animate-spin" />
				<span>Installing...</span>
			{:else}
				<Download size={18} />
				<span>Install as Module</span>
			{/if}
		</button>

		<button
			class="options-toggle-btn"
			onclick={toggleOptions}
			disabled={disabled || installing || installed}
			title="Configure installation options"
		>
			⚙️
		</button>
	{/if}
</div>

<style>
	.install-button-wrapper {
		display: flex;
		gap: 8px;
		position: relative;
	}

	.quick-install-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		background: linear-gradient(135deg, #3b82f6, #2563eb);
		border: none;
		border-radius: 8px;
		color: white;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.quick-install-btn:hover:not(:disabled) {
		background: linear-gradient(135deg, #2563eb, #1d4ed8);
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
	}

	.quick-install-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	.quick-install-btn.installed {
		background: linear-gradient(135deg, #10b981, #059669);
	}

	.options-toggle-btn {
		padding: 10px 12px;
		background: #374151;
		border: 1px solid #4b5563;
		border-radius: 8px;
		color: white;
		font-size: 16px;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.options-toggle-btn:hover:not(:disabled) {
		background: #4b5563;
		border-color: #60a5fa;
	}

	.options-toggle-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.install-options {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		background: #1f2937;
		border: 1px solid #4b5563;
		border-radius: 8px;
		padding: 20px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
		z-index: 10;
		min-width: 400px;
	}

	.options-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.options-header h4 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
		color: #f9fafb;
	}

	.close-btn {
		background: transparent;
		border: none;
		color: #9ca3af;
		font-size: 24px;
		cursor: pointer;
		padding: 0;
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 4px;
		transition: all 0.15s ease;
	}

	.close-btn:hover {
		background: #374151;
		color: #f9fafb;
	}

	.option-group {
		margin-bottom: 16px;
	}

	.option-group label {
		display: block;
		font-size: 13px;
		font-weight: 500;
		color: #d1d5db;
		margin-bottom: 6px;
	}

	.optional {
		font-weight: 400;
		color: #9ca3af;
		font-size: 12px;
	}

	.option-group input {
		width: 100%;
		padding: 10px 12px;
		background: #111827;
		border: 1px solid #374151;
		border-radius: 6px;
		color: #f9fafb;
		font-size: 14px;
		transition: all 0.15s ease;
	}

	.option-group input:focus {
		outline: none;
		border-color: #60a5fa;
		box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
	}

	.option-group input:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.info-text {
		font-size: 13px;
		color: #9ca3af;
		margin: 12px 0;
		padding: 8px 12px;
		background: #374151;
		border-radius: 4px;
	}

	.error-message {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 12px;
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.3);
		border-radius: 6px;
		color: #ef4444;
		font-size: 13px;
		margin: 12px 0;
	}

	.option-actions {
		display: flex;
		gap: 8px;
		margin-top: 16px;
	}

	.cancel-btn,
	.install-btn {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 10px 16px;
		border: none;
		border-radius: 6px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.cancel-btn {
		background: #374151;
		color: #e5e7eb;
	}

	.cancel-btn:hover:not(:disabled) {
		background: #4b5563;
	}

	.install-btn {
		background: linear-gradient(135deg, #3b82f6, #2563eb);
		color: white;
	}

	.install-btn:hover:not(:disabled) {
		background: linear-gradient(135deg, #2563eb, #1d4ed8);
	}

	.cancel-btn:disabled,
	.install-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
</style>
