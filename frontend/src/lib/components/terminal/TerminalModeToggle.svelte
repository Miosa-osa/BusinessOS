<script lang="ts">
	import { terminalPreferences, type TerminalMode } from '$lib/stores/terminalPreferences';
	import { AlertTriangle } from 'lucide-svelte';

	interface Props {
		currentMode: TerminalMode;
		onModeChange?: (mode: TerminalMode) => void;
	}

	let { currentMode = $bindable(), onModeChange }: Props = $props();

	let showWarning = $state(false);
	let pendingMode: TerminalMode | null = $state(null);

	function handleToggle() {
		const newMode: TerminalMode = currentMode === 'docker' ? 'local' : 'docker';

		if (newMode === 'local' && !$terminalPreferences.hasSeenLocalWarning) {
			pendingMode = newMode;
			showWarning = true;
		} else {
			switchMode(newMode);
		}
	}

	function confirmLocalMode() {
		terminalPreferences.markWarningShown();
		if (pendingMode) {
			switchMode(pendingMode);
		}
		showWarning = false;
	}

	function switchMode(mode: TerminalMode) {
		terminalPreferences.setDefaultMode(mode);
		currentMode = mode;
		onModeChange?.(mode);
	}
</script>

<div class="mode-toggle">
	<button onclick={handleToggle} class="toggle-button" class:local={currentMode === 'local'}>
		{#if currentMode === 'docker'}
			<span class="icon">🐳</span>
			<span>Docker</span>
		{:else}
			<span class="icon">💻</span>
			<span>Glimpse</span>
		{/if}
	</button>
</div>

{#if showWarning}
	<div class="modal-overlay" onclick={() => (showWarning = false)}>
		<div class="modal" onclick={(e) => e.stopPropagation()}>
			<AlertTriangle size={48} color="#f59e0b" />
			<h2>Switch to Local Mode?</h2>
			<p>Local mode provides <strong>full access to your Mac</strong>.</p>
			<p>Commands run directly on your system, not in a sandbox.</p>
			<div class="actions">
				<button class="cancel" onclick={() => (showWarning = false)}>Cancel</button>
				<button class="confirm" onclick={confirmLocalMode}>Enable Local Mode</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.mode-toggle {
		display: flex;
		align-items: center;
	}

	.toggle-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.25rem 0.75rem;
		border-radius: 0.375rem;
		background: rgba(74, 222, 128, 0.1);
		color: #4ade80;
		border: 1px solid #4ade80;
		cursor: pointer;
		transition: all 0.2s;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.toggle-button:hover {
		background: rgba(74, 222, 128, 0.2);
	}

	.toggle-button.local {
		background: rgba(251, 146, 60, 0.1);
		color: #fb923c;
		border-color: #fb923c;
	}

	.toggle-button.local:hover {
		background: rgba(251, 146, 60, 0.2);
	}

	.icon {
		font-size: 1rem;
	}

	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 9999;
	}

	.modal {
		background: white;
		padding: 2rem;
		border-radius: 0.5rem;
		max-width: 400px;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
	}

	.modal h2 {
		margin: 0;
		font-size: 1.5rem;
		color: #111827;
	}

	.modal p {
		margin: 0;
		color: #6b7280;
		text-align: center;
	}

	.actions {
		display: flex;
		gap: 1rem;
		margin-top: 1rem;
		width: 100%;
	}

	.actions button {
		flex: 1;
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.cancel {
		background: #f3f4f6;
		color: #374151;
		border: none;
	}

	.cancel:hover {
		background: #e5e7eb;
	}

	.confirm {
		background: #f59e0b;
		color: white;
		border: none;
	}

	.confirm:hover {
		background: #d97706;
	}
</style>
