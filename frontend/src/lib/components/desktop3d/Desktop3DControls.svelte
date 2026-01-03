<script lang="ts">
	import type { ViewMode } from '$lib/stores/desktop3dStore';

	interface Props {
		viewMode: ViewMode;
		autoRotate: boolean;
		hasFocusedWindow: boolean;
		onToggleView?: () => void;
		onToggleAutoRotate?: () => void;
		onExit?: () => void;
	}

	let {
		viewMode = 'orb',
		autoRotate = true,
		hasFocusedWindow = false,
		onToggleView,
		onToggleAutoRotate,
		onExit
	}: Props = $props();
</script>

<div class="controls-overlay">
	<!-- Top Left: Exit Button -->
	<div class="controls-top-left">
		<button class="control-btn exit-btn" onclick={onExit} title="Exit 3D Desktop (Esc)">
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
			</svg>
			<span>Exit</span>
		</button>
	</div>

	<!-- Top Right: View Controls -->
	<div class="controls-top-right">
		<!-- Auto-Rotate Toggle -->
		<button
			class="control-btn"
			class:active={autoRotate}
			onclick={onToggleAutoRotate}
			title="Toggle Auto-Rotate"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
			</svg>
		</button>

		<!-- View Mode Toggle -->
		<button
			class="control-btn view-toggle"
			onclick={onToggleView}
			disabled={hasFocusedWindow}
			title={viewMode === 'orb' ? 'Spread to Grid (Space)' : 'Collapse to Orb (Space)'}
		>
			{#if viewMode === 'orb'}
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
				</svg>
				<span>Grid</span>
			{:else}
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
				</svg>
				<span>Orb</span>
			{/if}
		</button>
	</div>

	<!-- Bottom Center: Instructions -->
	{#if !hasFocusedWindow}
		<div class="controls-bottom">
			<div class="instructions">
				<span class="instruction"><kbd>Drag</kbd> Rotate view</span>
				<span class="instruction"><kbd>Scroll</kbd> Zoom</span>
				<span class="instruction"><kbd>Click</kbd> Focus window</span>
				<span class="instruction"><kbd>Space</kbd> Toggle view</span>
				<span class="instruction"><kbd>Esc</kbd> Exit</span>
			</div>
		</div>
	{:else}
		<div class="controls-bottom">
			<div class="instructions focused-instructions">
				<span class="instruction"><kbd>Click outside</kbd> or <kbd>Esc</kbd> to unfocus</span>
			</div>
		</div>
	{/if}
</div>

<style>
	.controls-overlay {
		position: fixed;
		inset: 0;
		pointer-events: none;
		z-index: 50;
	}

	.controls-top-left {
		position: absolute;
		top: 60px; /* Below MenuBar */
		left: 20px;
		display: flex;
		gap: 10px;
		pointer-events: auto;
	}

	.controls-top-right {
		position: absolute;
		top: 60px; /* Below MenuBar */
		right: 20px;
		display: flex;
		gap: 10px;
		pointer-events: auto;
	}

	.controls-bottom {
		position: absolute;
		bottom: 100px;
		left: 50%;
		transform: translateX(-50%);
		pointer-events: none;
	}

	.control-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		background: rgba(255, 255, 255, 0.85);
		backdrop-filter: blur(12px);
		border: 1px solid rgba(0, 0, 0, 0.08);
		border-radius: 10px;
		color: #333333;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
	}

	.control-btn:hover {
		background: rgba(255, 255, 255, 0.95);
		border-color: rgba(0, 0, 0, 0.12);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.control-btn.active {
		background: rgba(74, 158, 255, 0.15);
		border-color: rgba(74, 158, 255, 0.3);
		color: #1a73e8;
	}

	.control-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.exit-btn {
		background: rgba(255, 240, 240, 0.9);
		border-color: rgba(200, 100, 100, 0.2);
		color: #c53030;
	}

	.exit-btn:hover {
		background: rgba(255, 230, 230, 0.95);
		border-color: rgba(200, 100, 100, 0.3);
	}

	.instructions {
		display: flex;
		gap: 20px;
		padding: 12px 20px;
		background: rgba(255, 255, 255, 0.85);
		backdrop-filter: blur(12px);
		border-radius: 12px;
		border: 1px solid rgba(0, 0, 0, 0.08);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
	}

	.focused-instructions {
		background: rgba(74, 158, 255, 0.1);
		border-color: rgba(74, 158, 255, 0.2);
	}

	.instruction {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		color: #666666;
	}

	.instruction kbd {
		padding: 3px 8px;
		background: rgba(0, 0, 0, 0.05);
		border: 1px solid rgba(0, 0, 0, 0.1);
		border-radius: 4px;
		font-family: inherit;
		font-size: 11px;
		color: #333333;
	}

	/* ===== DARK MODE STYLES ===== */
	:global(.dark) .control-btn {
		background: rgba(44, 44, 46, 0.85);
		border-color: rgba(255, 255, 255, 0.12);
		color: #ffffff;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
	}

	:global(.dark) .control-btn:hover {
		background: rgba(58, 58, 60, 0.95);
		border-color: rgba(255, 255, 255, 0.18);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .control-btn.active {
		background: rgba(74, 158, 255, 0.25);
		border-color: rgba(74, 158, 255, 0.5);
		color: #6eb5ff;
	}

	:global(.dark) .control-btn svg {
		stroke: #ffffff;
	}

	:global(.dark) .control-btn.active svg {
		stroke: #6eb5ff;
	}

	:global(.dark) .exit-btn {
		background: rgba(80, 30, 30, 0.85);
		border-color: rgba(255, 100, 100, 0.3);
		color: #ff8888;
	}

	:global(.dark) .exit-btn:hover {
		background: rgba(100, 40, 40, 0.95);
		border-color: rgba(255, 100, 100, 0.5);
	}

	:global(.dark) .exit-btn svg {
		stroke: #ff8888;
	}

	:global(.dark) .instructions {
		background: rgba(44, 44, 46, 0.85);
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
	}

	:global(.dark) .focused-instructions {
		background: rgba(74, 158, 255, 0.15);
		border-color: rgba(74, 158, 255, 0.3);
	}

	:global(.dark) .instruction {
		color: #aaaaaa;
	}

	:global(.dark) .instruction kbd {
		background: rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.15);
		color: #ffffff;
	}
</style>
