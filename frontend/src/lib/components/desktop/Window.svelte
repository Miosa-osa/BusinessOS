<script lang="ts">
	import { windowStore, type WindowState, type SnapZone } from '$lib/stores/windowStore';

	interface Props {
		window: WindowState;
		focused: boolean;
		zIndex: number;
		workspaceHeight: number;
		workspaceWidth: number;
		children?: import('svelte').Snippet;
		onsnapZoneChange?: (zone: SnapZone) => void;
	}

	let {
		window,
		focused,
		zIndex,
		workspaceHeight,
		workspaceWidth,
		children,
		onsnapZoneChange
	}: Props = $props();

	let windowElement: HTMLDivElement;
	let isDragging = $state(false);
	let isResizing = $state(false);
	let resizeDirection = $state('');
	let dragOffset = { x: 0, y: 0 };
	let startBounds = { x: 0, y: 0, width: 0, height: 0 };
	let startMouse = { x: 0, y: 0 };

	// Snap zone state
	let currentSnapZone = $state<SnapZone>(null);
	const SNAP_THRESHOLD = 20; // Pixels from edge to trigger snap zone

	// Close button hover state for showing X
	let closeHover = $state(false);
	let minimizeHover = $state(false);
	let maximizeHover = $state(false);

	// Detect which snap zone the mouse is in
	function detectSnapZone(mouseX: number, mouseY: number): SnapZone {
		const nearLeft = mouseX < SNAP_THRESHOLD;
		const nearRight = mouseX > workspaceWidth - SNAP_THRESHOLD;
		const nearTop = mouseY < SNAP_THRESHOLD;
		const nearBottom = mouseY > workspaceHeight - SNAP_THRESHOLD;

		// Corner zones (quadrants)
		if (nearLeft && nearTop) return 'top-left';
		if (nearRight && nearTop) return 'top-right';
		if (nearLeft && nearBottom) return 'bottom-left';
		if (nearRight && nearBottom) return 'bottom-right';

		// Edge zones (halves)
		if (nearLeft) return 'left';
		if (nearRight) return 'right';

		return null;
	}

	// Calculate actual bounds (for maximized state)
	const bounds = $derived(() => {
		if (window.maximized) {
			return {
				x: 0,
				y: 0,
				width: workspaceWidth,
				height: workspaceHeight
			};
		}
		return {
			x: window.x,
			y: window.y,
			width: window.width,
			height: window.height
		};
	});

	function handleTitleBarMouseDown(event: MouseEvent) {
		if (window.maximized) return;
		if ((event.target as HTMLElement).closest('.window-controls')) return;

		isDragging = true;
		dragOffset = {
			x: event.clientX - window.x,
			y: event.clientY - window.y
		};

		windowStore.focusWindow(window.id);

		document.addEventListener('mousemove', handleMouseMove);
		document.addEventListener('mouseup', handleMouseUp);
	}

	function handleResizeMouseDown(event: MouseEvent, direction: string) {
		if (window.maximized) return;

		isResizing = true;
		resizeDirection = direction;
		startBounds = {
			x: window.x,
			y: window.y,
			width: window.width,
			height: window.height
		};
		startMouse = { x: event.clientX, y: event.clientY };

		windowStore.focusWindow(window.id);

		document.addEventListener('mousemove', handleResizeMove);
		document.addEventListener('mouseup', handleResizeUp);
		event.preventDefault();
	}

	function handleMouseMove(event: MouseEvent) {
		if (!isDragging) return;

		const newX = Math.max(0, Math.min(event.clientX - dragOffset.x, workspaceWidth - 100));
		const newY = Math.max(0, Math.min(event.clientY - dragOffset.y, workspaceHeight - 50));

		windowStore.updateWindowPosition(window.id, newX, newY);

		// Detect snap zone while dragging
		const zone = detectSnapZone(event.clientX, event.clientY);
		if (zone !== currentSnapZone) {
			currentSnapZone = zone;
			onsnapZoneChange?.(zone);
		}
	}

	function handleMouseUp(event: MouseEvent) {
		// Check if we should snap the window
		if (currentSnapZone) {
			windowStore.snapWindow(window.id, currentSnapZone, workspaceWidth, workspaceHeight);
		}

		isDragging = false;
		currentSnapZone = null;
		onsnapZoneChange?.(null);
		document.removeEventListener('mousemove', handleMouseMove);
		document.removeEventListener('mouseup', handleMouseUp);
	}

	function handleResizeMove(event: MouseEvent) {
		if (!isResizing) return;

		const deltaX = event.clientX - startMouse.x;
		const deltaY = event.clientY - startMouse.y;

		let newX = startBounds.x;
		let newY = startBounds.y;
		let newWidth = startBounds.width;
		let newHeight = startBounds.height;

		if (resizeDirection.includes('e')) {
			newWidth = Math.max(window.minWidth, startBounds.width + deltaX);
		}
		if (resizeDirection.includes('w')) {
			const maxDelta = startBounds.width - window.minWidth;
			const actualDelta = Math.min(deltaX, maxDelta);
			newX = startBounds.x + actualDelta;
			newWidth = startBounds.width - actualDelta;
		}
		if (resizeDirection.includes('s')) {
			newHeight = Math.max(window.minHeight, startBounds.height + deltaY);
		}
		if (resizeDirection.includes('n')) {
			const maxDelta = startBounds.height - window.minHeight;
			const actualDelta = Math.min(deltaY, maxDelta);
			newY = startBounds.y + actualDelta;
			newHeight = startBounds.height - actualDelta;
		}

		windowStore.updateWindowBounds(window.id, newX, newY, newWidth, newHeight);
	}

	function handleResizeUp() {
		isResizing = false;
		document.removeEventListener('mousemove', handleResizeMove);
		document.removeEventListener('mouseup', handleResizeUp);
	}

	function handleWindowClick() {
		windowStore.focusWindow(window.id);
	}

	function handleWindowMouseDown(event: MouseEvent) {
		// Focus window on any mousedown within the window
		// This ensures focus even when clicking on interactive content
		windowStore.focusWindow(window.id);
	}

	function handleClose() {
		windowStore.closeWindow(window.id);
	}

	function handleMinimize() {
		windowStore.minimizeWindow(window.id);
	}

	function handleMaximize() {
		windowStore.toggleMaximize(window.id);
	}

	function handleTitleBarDoubleClick() {
		windowStore.toggleMaximize(window.id);
	}
</script>

<div
	bind:this={windowElement}
	class="window"
	class:focused
	class:maximized={window.maximized}
	class:dragging={isDragging}
	class:resizing={isResizing}
	style="
		left: {bounds().x}px;
		top: {bounds().y}px;
		width: {bounds().width}px;
		height: {bounds().height}px;
		z-index: {zIndex};
	"
	onmousedown={handleWindowMouseDown}
	onclick={handleWindowClick}
	onkeydown={(e) => e.key === 'Enter' && handleWindowClick()}
	role="dialog"
	aria-label={window.title}
	tabindex="0"
>
	<!-- Title Bar -->
	<div
		class="title-bar"
		class:focused
		onmousedown={handleTitleBarMouseDown}
		ondblclick={handleTitleBarDoubleClick}
		role="toolbar"
		tabindex="-1"
	>
		<!-- Window Controls (Traffic Lights) -->
		<div class="window-controls">
			<button
				class="control-button close"
				class:hover={closeHover}
				onmouseenter={() => closeHover = true}
				onmouseleave={() => closeHover = false}
				onclick={handleClose}
				aria-label="Close"
			>
				{#if closeHover || focused}
					<svg class="control-icon" viewBox="0 0 12 12">
						<path d="M3 3l6 6M9 3l-6 6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
					</svg>
				{/if}
			</button>
			<button
				class="control-button minimize"
				class:hover={minimizeHover}
				onmouseenter={() => minimizeHover = true}
				onmouseleave={() => minimizeHover = false}
				onclick={handleMinimize}
				aria-label="Minimize"
			>
				{#if minimizeHover || focused}
					<svg class="control-icon" viewBox="0 0 12 12">
						<path d="M2 6h8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
					</svg>
				{/if}
			</button>
			<button
				class="control-button maximize"
				class:hover={maximizeHover}
				onmouseenter={() => maximizeHover = true}
				onmouseleave={() => maximizeHover = false}
				onclick={handleMaximize}
				aria-label="Maximize"
			>
				{#if maximizeHover || focused}
					{#if window.maximized}
						<!-- Restore icon -->
						<svg class="control-icon" viewBox="0 0 12 12">
							<path d="M4 8V4h4M8 4v4H4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					{:else}
						<!-- Maximize icon -->
						<svg class="control-icon" viewBox="0 0 12 12">
							<path d="M2 4l4-2 4 2v4l-4 2-4-2z" stroke="currentColor" stroke-width="1" fill="none"/>
						</svg>
					{/if}
				{/if}
			</button>
		</div>

		<!-- Title -->
		<span class="title-text">{window.title}</span>

		<!-- Spacer to center title -->
		<div class="title-spacer"></div>
	</div>

	<!-- Content Area -->
	<div class="window-content">
		<!-- Focus overlay - captures clicks when window is not focused -->
		{#if !focused}
			<div
				class="focus-overlay"
				onmousedown={(e) => {
					windowStore.focusWindow(window.id);
					// Don't prevent default - let the click go through after focusing
				}}
			></div>
		{/if}
		{#if children}
			{@render children()}
		{:else}
			<div class="window-placeholder">
				<span>{window.module} content</span>
			</div>
		{/if}
	</div>

	<!-- Resize Handles -->
	{#if !window.maximized}
		<div class="resize-handle resize-n" role="separator" aria-orientation="horizontal" onmousedown={(e) => handleResizeMouseDown(e, 'n')}></div>
		<div class="resize-handle resize-s" role="separator" aria-orientation="horizontal" onmousedown={(e) => handleResizeMouseDown(e, 's')}></div>
		<div class="resize-handle resize-e" role="separator" aria-orientation="vertical" onmousedown={(e) => handleResizeMouseDown(e, 'e')}></div>
		<div class="resize-handle resize-w" role="separator" aria-orientation="vertical" onmousedown={(e) => handleResizeMouseDown(e, 'w')}></div>
		<div class="resize-handle resize-ne" role="separator" onmousedown={(e) => handleResizeMouseDown(e, 'ne')}></div>
		<div class="resize-handle resize-nw" role="separator" onmousedown={(e) => handleResizeMouseDown(e, 'nw')}></div>
		<div class="resize-handle resize-se" role="separator" onmousedown={(e) => handleResizeMouseDown(e, 'se')}></div>
		<div class="resize-handle resize-sw" role="separator" onmousedown={(e) => handleResizeMouseDown(e, 'sw')}></div>
	{/if}
</div>

<style>
	.window {
		position: absolute;
		background: white;
		border-radius: 10px;
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
		border: 1px solid rgba(0, 0, 0, 0.1);
		display: flex;
		flex-direction: column;
		overflow: hidden;
		transition: box-shadow 0.2s ease;
	}

	.window.focused {
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
	}

	.window.maximized {
		border-radius: 0;
	}

	.window.dragging,
	.window.resizing {
		transition: none;
		user-select: none;
	}

	.window.dragging {
		opacity: 0.95;
	}

	.title-bar {
		height: 36px;
		background: linear-gradient(to bottom, #f8f8f8, #e8e8e8);
		border-bottom: 1px solid #d0d0d0;
		display: flex;
		align-items: center;
		padding: 0 12px;
		cursor: grab;
		flex-shrink: 0;
		position: relative;
		z-index: 1;
		user-select: none;
		-webkit-user-select: none;
	}

	.window.dragging .title-bar {
		cursor: grabbing;
	}

	.window.maximized .title-bar {
		cursor: default;
	}

	.title-bar.focused {
		background: linear-gradient(to bottom, #ffffff, #f0f0f0);
	}

	.window-controls {
		display: flex;
		gap: 8px;
		flex-shrink: 0;
	}

	.control-button {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: filter 0.15s ease;
	}

	.control-button:hover {
		filter: brightness(0.9);
	}

	.control-button.close {
		background: #FF5F57;
	}

	.control-button.minimize {
		background: #FFBD2E;
	}

	.control-button.maximize {
		background: #28C840;
	}

	.window:not(.focused) .control-button {
		background: #CDCDCD;
	}

	.control-icon {
		width: 8px;
		height: 8px;
		color: rgba(0, 0, 0, 0.5);
	}

	.title-text {
		flex: 1;
		text-align: center;
		font-size: 13px;
		font-weight: 500;
		color: #333;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		padding: 0 12px;
	}

	.window:not(.focused) .title-text {
		color: #999;
	}

	.title-spacer {
		width: 60px;
		flex-shrink: 0;
	}

	.window-content {
		flex: 1;
		overflow: auto;
		background: white;
		position: relative;
	}

	.focus-overlay {
		position: absolute;
		inset: 0;
		z-index: 5;
		cursor: default;
		/* Transparent overlay to capture clicks */
	}

	.window-placeholder {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: #999;
		font-size: 14px;
	}

	/* Resize Handles */
	.resize-handle {
		position: absolute;
		z-index: 10;
	}

	.resize-n {
		top: 0;
		left: 10px;
		right: 10px;
		height: 4px;
		cursor: ns-resize;
	}

	.resize-s {
		bottom: 0;
		left: 10px;
		right: 10px;
		height: 4px;
		cursor: ns-resize;
	}

	.resize-e {
		right: 0;
		top: 10px;
		bottom: 10px;
		width: 4px;
		cursor: ew-resize;
	}

	.resize-w {
		left: 0;
		top: 10px;
		bottom: 10px;
		width: 4px;
		cursor: ew-resize;
	}

	.resize-ne {
		top: 0;
		right: 0;
		width: 10px;
		height: 10px;
		cursor: nesw-resize;
	}

	.resize-nw {
		top: 0;
		left: 0;
		width: 10px;
		height: 10px;
		cursor: nwse-resize;
	}

	.resize-se {
		bottom: 0;
		right: 0;
		width: 10px;
		height: 10px;
		cursor: nwse-resize;
	}

	.resize-sw {
		bottom: 0;
		left: 0;
		width: 10px;
		height: 10px;
		cursor: nesw-resize;
	}
</style>
