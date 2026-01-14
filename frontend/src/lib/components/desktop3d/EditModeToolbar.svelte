<script lang="ts">
	import { desktop3dLayoutStore, isEditMode, activeLayout } from '$lib/stores/desktop3dLayoutStore';
	import LayoutManager from './LayoutManager.svelte';

	// State for save layout modal
	let showSaveModal = $state(false);
	let layoutName = $state('');
	let saving = $state(false);
	let saveError = $state<string | null>(null);

	// State for layout manager
	let showLayoutManager = $state(false);

	function handleEnterEditMode() {
		desktop3dLayoutStore.enterEditMode();
		console.log('[EditModeToolbar] Entered edit mode');
	}

	function handleExitEditMode() {
		desktop3dLayoutStore.exitEditMode();
		console.log('[EditModeToolbar] Exited edit mode');
	}

	function handleOpenSaveModal() {
		showSaveModal = true;
		layoutName = '';
		saveError = null;
	}

	function handleCloseSaveModal() {
		showSaveModal = false;
		layoutName = '';
		saveError = null;
	}

	async function handleSaveLayout() {
		if (!layoutName || layoutName.trim().length === 0) {
			saveError = 'Please enter a layout name';
			return;
		}

		if (layoutName.trim().length > 255) {
			saveError = 'Layout name must be 255 characters or less';
			return;
		}

		saving = true;
		saveError = null;

		try {
			const success = await desktop3dLayoutStore.saveLayout(layoutName.trim());

			if (success) {
				console.log('[EditModeToolbar] ✅ Layout saved:', layoutName);
				handleCloseSaveModal();
				desktop3dLayoutStore.exitEditMode();
			} else {
				saveError = 'Failed to save layout. Please try again.';
			}
		} catch (err) {
			console.error('[EditModeToolbar] Save error:', err);
			saveError = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			saving = false;
		}
	}
</script>

<!-- Toolbar container -->
<div class="edit-mode-toolbar">
	{#if !$isEditMode}
		<!-- View Mode: Show Enter Edit Mode button -->
		<div class="toolbar-content">
			<div class="layout-info">
				<span class="layout-label">Current Layout:</span>
				<span class="layout-name">{$activeLayout?.name || 'Default'}</span>
			</div>

			<button class="btn btn-secondary" onclick={() => (showLayoutManager = true)}>
				<svg
					class="icon"
					width="16"
					height="16"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<rect x="3" y="3" width="7" height="7" />
					<rect x="14" y="3" width="7" height="7" />
					<rect x="14" y="14" width="7" height="7" />
					<rect x="3" y="14" width="7" height="7" />
				</svg>
				Manage Layouts
			</button>

			<button class="btn btn-primary" onclick={handleEnterEditMode}>
				<svg
					class="icon"
					width="16"
					height="16"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
					<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
				</svg>
				Edit Layout
			</button>
		</div>
	{:else}
		<!-- Edit Mode: Show Save and Cancel buttons -->
		<div class="toolbar-content edit-mode-active">
			<div class="edit-indicator">
				<div class="edit-pulse"></div>
				<span class="edit-text">Edit Mode Active</span>
			</div>

			<div class="button-group">
				<button class="btn btn-secondary" onclick={handleExitEditMode}>
					<svg
						class="icon"
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<line x1="18" y1="6" x2="6" y2="18" />
						<line x1="6" y1="6" x2="18" y2="18" />
					</svg>
					Cancel
				</button>

				<button class="btn btn-success" onclick={handleOpenSaveModal}>
					<svg
						class="icon"
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
						<polyline points="17 21 17 13 7 13 7 21" />
						<polyline points="7 3 7 8 15 8" />
					</svg>
					Save Layout
				</button>
			</div>
		</div>
	{/if}
</div>

<!-- Save Layout Modal -->
{#if showSaveModal}
	<div class="modal-overlay" onclick={handleCloseSaveModal}>
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h3>Save Custom Layout</h3>
				<button class="close-btn" onclick={handleCloseSaveModal}>
					<svg
						width="20"
						height="20"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<line x1="18" y1="6" x2="6" y2="18" />
						<line x1="6" y1="6" x2="18" y2="18" />
					</svg>
				</button>
			</div>

			<div class="modal-body">
				<p class="modal-description">Enter a name for your custom 3D Desktop layout:</p>

				<input
					type="text"
					class="layout-name-input"
					placeholder="e.g., My Workspace, Development Setup, etc."
					bind:value={layoutName}
					maxlength="255"
					disabled={saving}
					onkeydown={(e) => {
						if (e.key === 'Enter' && !saving) {
							handleSaveLayout();
						}
					}}
				/>

				{#if saveError}
					<div class="error-message">{saveError}</div>
				{/if}
			</div>

			<div class="modal-footer">
				<button class="btn btn-secondary" onclick={handleCloseSaveModal} disabled={saving}>
					Cancel
				</button>

				<button class="btn btn-primary" onclick={handleSaveLayout} disabled={saving || !layoutName}>
					{#if saving}
						<span class="spinner"></span>
						Saving...
					{:else}
						Save
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Layout Manager Modal -->
<LayoutManager show={showLayoutManager} onClose={() => (showLayoutManager = false)} />

<style>
	.edit-mode-toolbar {
		position: fixed;
		top: 20px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 1000;
		animation: slideDown 0.3s ease-out;
	}

	@keyframes slideDown {
		from {
			opacity: 0;
			transform: translateX(-50%) translateY(-20px);
		}
		to {
			opacity: 1;
			transform: translateX(-50%) translateY(0);
		}
	}

	.toolbar-content {
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 12px 20px;
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(10px);
		border-radius: 12px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.toolbar-content.edit-mode-active {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.95), rgba(147, 51, 234, 0.95));
		color: white;
	}

	.layout-info {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
	}

	.layout-label {
		color: #64748b;
		font-weight: 500;
	}

	.layout-name {
		color: #1e293b;
		font-weight: 600;
	}

	.edit-indicator {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.edit-pulse {
		width: 8px;
		height: 8px;
		background: #22c55e;
		border-radius: 50%;
		animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	.edit-text {
		font-weight: 600;
		font-size: 14px;
	}

	.button-group {
		display: flex;
		gap: 12px;
	}

	.btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		border: none;
		border-radius: 8px;
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-primary {
		background: linear-gradient(135deg, #3b82f6, #2563eb);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
	}

	.btn-secondary {
		background: rgba(255, 255, 255, 0.9);
		color: #475569;
		border: 1px solid #e2e8f0;
	}

	.btn-secondary:hover:not(:disabled) {
		background: rgba(255, 255, 255, 1);
		border-color: #cbd5e1;
	}

	.btn-success {
		background: linear-gradient(135deg, #22c55e, #16a34a);
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(34, 197, 94, 0.4);
	}

	.icon {
		flex-shrink: 0;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(4px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 2000;
		animation: fadeIn 0.2s ease-out;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	.modal-content {
		background: white;
		border-radius: 16px;
		width: 90%;
		max-width: 500px;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
		animation: slideUp 0.3s ease-out;
	}

	@keyframes slideUp {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 24px;
		border-bottom: 1px solid #e2e8f0;
	}

	.modal-header h3 {
		margin: 0;
		font-size: 20px;
		font-weight: 600;
		color: #1e293b;
	}

	.close-btn {
		background: none;
		border: none;
		padding: 4px;
		cursor: pointer;
		color: #64748b;
		transition: color 0.2s;
	}

	.close-btn:hover {
		color: #1e293b;
	}

	.modal-body {
		padding: 24px;
	}

	.modal-description {
		margin: 0 0 16px 0;
		color: #64748b;
		font-size: 14px;
	}

	.layout-name-input {
		width: 100%;
		padding: 12px 16px;
		border: 2px solid #e2e8f0;
		border-radius: 8px;
		font-size: 14px;
		transition: all 0.2s;
	}

	.layout-name-input:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.layout-name-input:disabled {
		background: #f8fafc;
		cursor: not-allowed;
	}

	.error-message {
		margin-top: 12px;
		padding: 12px;
		background: #fef2f2;
		border: 1px solid #fecaca;
		border-radius: 8px;
		color: #dc2626;
		font-size: 13px;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 12px;
		padding: 24px;
		border-top: 1px solid #e2e8f0;
	}

	.spinner {
		display: inline-block;
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
