<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import { windowStore, visibleWindows, focusedWindow, type SnapZone } from '$lib/stores/windowStore';
	import { desktopSettings, getBackgroundCSS, isBackgroundDark } from '$lib/stores/desktopStore';
	import { onMount } from 'svelte';

	import MenuBar from '$lib/components/desktop/MenuBar.svelte';
	import DesktopIcon from '$lib/components/desktop/DesktopIcon.svelte';
	import Window from '$lib/components/desktop/Window.svelte';
	import Dock from '$lib/components/desktop/Dock.svelte';
	import Terminal from '$lib/components/desktop/Terminal.svelte';
	import DesktopSettingsContent from '$lib/components/desktop/DesktopSettingsContent.svelte';

	const session = useSession();

	// Workspace dimensions (excluding menu bar and dock)
	let workspaceElement: HTMLDivElement;
	let workspaceWidth = $state(0);
	let workspaceHeight = $state(0);

	// Grid settings for icons - dynamic based on icon size
	const ICON_PADDING = 16;
	// Grid size adjusts based on icon size to prevent overlap
	const GRID_SIZE = $derived(Math.max(96, $desktopSettings.iconSize + 40));

	// Check if current background is dark (needs light text)
	const darkBackground = $derived(isBackgroundDark($desktopSettings.backgroundId));

	// Track icon positions (pixel-based for dragging)
	let iconPositions = $state<Record<string, { x: number; y: number }>>({});

	// Track which icon is being dragged
	let draggingIconId = $state<string | null>(null);

	// Selection box (lasso) state
	let isSelecting = $state(false);
	let selectionStart = $state({ x: 0, y: 0 });
	let selectionEnd = $state({ x: 0, y: 0 });
	let didSelectionDrag = $state(false);

	// Snap zone preview state
	let currentSnapZone = $state<SnapZone>(null);

	// Selection box computed bounds
	const selectionBox = $derived(() => {
		if (!isSelecting) return null;
		return {
			x: Math.min(selectionStart.x, selectionEnd.x),
			y: Math.min(selectionStart.y, selectionEnd.y),
			width: Math.abs(selectionEnd.x - selectionStart.x),
			height: Math.abs(selectionEnd.y - selectionStart.y)
		};
	});

	// Snap zone preview bounds
	const snapZonePreview = $derived(() => {
		if (!currentSnapZone || workspaceWidth === 0 || workspaceHeight === 0) return null;

		switch (currentSnapZone) {
			case 'left':
				return { x: 0, y: 0, width: workspaceWidth / 2, height: workspaceHeight };
			case 'right':
				return { x: workspaceWidth / 2, y: 0, width: workspaceWidth / 2, height: workspaceHeight };
			case 'top-left':
				return { x: 0, y: 0, width: workspaceWidth / 2, height: workspaceHeight / 2 };
			case 'top-right':
				return { x: workspaceWidth / 2, y: 0, width: workspaceWidth / 2, height: workspaceHeight / 2 };
			case 'bottom-left':
				return { x: 0, y: workspaceHeight / 2, width: workspaceWidth / 2, height: workspaceHeight / 2 };
			case 'bottom-right':
				return { x: workspaceWidth / 2, y: workspaceHeight / 2, width: workspaceWidth / 2, height: workspaceHeight / 2 };
			default:
				return null;
		}
	});

	// Handle snap zone change from window dragging
	function handleSnapZoneChange(zone: SnapZone) {
		currentSnapZone = zone;
	}

	$effect(() => {
		if (!$session.isPending && !$session.data) {
			goto('/login');
		}
	});

	// Update workspace dimensions when element is available
	$effect(() => {
		if (workspaceElement) {
			workspaceWidth = workspaceElement.clientWidth;
			workspaceHeight = workspaceElement.clientHeight;
		}
	});

	// Get background style
	const backgroundStyle = $derived(() => {
		const bgCSS = getBackgroundCSS($desktopSettings.backgroundId, $desktopSettings.customBackgroundUrl);
		const isCustomImage = $desktopSettings.backgroundId === 'custom';

		if (isCustomImage && $desktopSettings.customBackgroundUrl) {
			// Map fit option to CSS
			const fitMap: Record<string, string> = {
				'cover': 'cover',
				'contain': 'contain',
				'fill': '100% 100%',
				'center': 'auto'
			};
			const bgSize = fitMap[$desktopSettings.backgroundFit] || 'cover';

			// For custom images, use separate properties to ensure proper fitting
			return `
				background-image: ${bgCSS.background};
				background-size: ${bgSize};
				background-position: center center;
				background-repeat: no-repeat;
				background-attachment: fixed;
				background-color: #1a1a1a;
			`;
		} else if (bgCSS.backgroundSize) {
			// For patterns
			return `background: ${bgCSS.background}; background-size: ${bgCSS.backgroundSize};`;
		}
		// For solid colors and gradients
		return `background: ${bgCSS.background};`;
	});

	onMount(() => {
		// Update workspace dimensions on resize
		function updateDimensions() {
			if (workspaceElement) {
				workspaceWidth = workspaceElement.clientWidth;
				workspaceHeight = workspaceElement.clientHeight;
			}
		}

		window.addEventListener('resize', updateDimensions);

		// Keyboard shortcuts
		function handleKeyDown(event: KeyboardEvent) {
			// Don't handle shortcuts when focus is inside an iframe or input
			const activeElement = document.activeElement;
			if (activeElement?.tagName === 'IFRAME' ||
				activeElement?.tagName === 'INPUT' ||
				activeElement?.tagName === 'TEXTAREA') {
				return;
			}

			const isMeta = event.metaKey || event.ctrlKey;

			if (isMeta && event.key === 'w') {
				event.preventDefault();
				if ($focusedWindow) {
					windowStore.closeWindow($focusedWindow.id);
				}
			} else if (isMeta && event.key === 'm') {
				event.preventDefault();
				if ($focusedWindow) {
					windowStore.minimizeWindow($focusedWindow.id);
				}
			} else if (isMeta && event.key === '`') {
				event.preventDefault();
				windowStore.cycleWindows();
			} else if (event.key === 'Escape') {
				windowStore.clearIconSelection();
			}
		}

		window.addEventListener('keydown', handleKeyDown);

		return () => {
			window.removeEventListener('resize', updateDimensions);
			window.removeEventListener('keydown', handleKeyDown);
		};
	});

	// Calculate icon positions - use stored pixel position or calculate from grid
	function getIconPosition(icon: { id: string; x: number; y: number }) {
		// Check if we have a custom position from dragging
		if (iconPositions[icon.id]) {
			return iconPositions[icon.id];
		}

		// Calculate from grid position
		let x: number;
		let y: number;

		if (icon.x < 0) {
			// Negative x means from right edge
			x = workspaceWidth + (icon.x * GRID_SIZE) - ICON_PADDING;
		} else {
			x = icon.x * GRID_SIZE + ICON_PADDING;
		}

		if (icon.y < 0) {
			// Negative y means from bottom
			y = workspaceHeight + (icon.y * GRID_SIZE) - ICON_PADDING;
		} else {
			y = icon.y * GRID_SIZE + ICON_PADDING;
		}

		return { x, y };
	}

	function handleIconDragStart(iconId: string) {
		draggingIconId = iconId;
	}

	function handleIconDragMove(iconId: string, newX: number, newY: number) {
		// Constrain to workspace bounds
		const constrainedX = Math.max(0, Math.min(newX, workspaceWidth - 80));
		const constrainedY = Math.max(0, Math.min(newY, workspaceHeight - 100));

		iconPositions = {
			...iconPositions,
			[iconId]: { x: constrainedX, y: constrainedY }
		};
	}

	function handleDesktopClick(event: MouseEvent) {
		// Skip if we just finished a selection drag
		if (didSelectionDrag) {
			didSelectionDrag = false;
			return;
		}
		// Only clear selection if clicking directly on desktop (not on icon or window)
		if ((event.target as HTMLElement).classList.contains('desktop-workspace')) {
			windowStore.clearIconSelection();
		}
	}

	// Selection box handlers
	function handleDesktopMouseDown(event: MouseEvent) {
		// Only start selection on left click directly on desktop
		if (event.button !== 0) return;
		if (!(event.target as HTMLElement).classList.contains('desktop-workspace')) return;

		const rect = workspaceElement.getBoundingClientRect();
		const x = event.clientX - rect.left;
		const y = event.clientY - rect.top;

		selectionStart = { x, y };
		selectionEnd = { x, y };
		isSelecting = true;
		didSelectionDrag = false;

		// Clear selection if not holding shift
		if (!event.shiftKey) {
			windowStore.clearIconSelection();
		}

		document.addEventListener('mousemove', handleSelectionMove);
		document.addEventListener('mouseup', handleSelectionEnd);
	}

	function handleSelectionMove(event: MouseEvent) {
		if (!isSelecting || !workspaceElement) return;

		const rect = workspaceElement.getBoundingClientRect();
		const x = Math.max(0, Math.min(event.clientX - rect.left, workspaceWidth));
		const y = Math.max(0, Math.min(event.clientY - rect.top, workspaceHeight));

		selectionEnd = { x, y };

		// Select icons that intersect with the selection box
		const box = selectionBox();
		if (box && box.width > 5 && box.height > 5) {
			didSelectionDrag = true;
			selectIconsInBox(box);
		}
	}

	function handleSelectionEnd() {
		isSelecting = false;
		document.removeEventListener('mousemove', handleSelectionMove);
		document.removeEventListener('mouseup', handleSelectionEnd);
	}

	function selectIconsInBox(box: { x: number; y: number; width: number; height: number }) {
		const selectedIds: string[] = [];

		for (const icon of $windowStore.desktopIcons) {
			const pos = getIconPosition(icon);
			const iconWidth = 80;
			const iconHeight = 90;

			// Check if icon intersects with selection box
			const iconRight = pos.x + iconWidth;
			const iconBottom = pos.y + iconHeight;
			const boxRight = box.x + box.width;
			const boxBottom = box.y + box.height;

			if (
				pos.x < boxRight &&
				iconRight > box.x &&
				pos.y < boxBottom &&
				iconBottom > box.y
			) {
				selectedIds.push(icon.id);
			}
		}

		// Update selection in store
		if (selectedIds.length > 0) {
			windowStore.setSelectedIcons(selectedIds);
		}
	}

	function handleIconSelect(iconId: string, additive: boolean) {
		windowStore.selectIcon(iconId, additive);
	}

	function handleIconOpen(module: string) {
		windowStore.openWindow(module);
	}

	function handleIconDragEnd(iconId: string, finalX: number, finalY: number) {
		// Store the final pixel position
		const constrainedX = Math.max(0, Math.min(finalX, workspaceWidth - 80));
		const constrainedY = Math.max(0, Math.min(finalY, workspaceHeight - 100));

		iconPositions = {
			...iconPositions,
			[iconId]: { x: constrainedX, y: constrainedY }
		};
		draggingIconId = null;
	}

	// Get z-index for a window based on its position in windowOrder
	function getWindowZIndex(windowId: string): number {
		const index = $windowStore.windowOrder.indexOf(windowId);
		return 100 + index;
	}
</script>

<svelte:head>
	<title>Business OS - Desktop</title>
</svelte:head>

{#if $session.isPending}
	<div class="loading-screen">
		<div class="loading-spinner"></div>
	</div>
{:else if $session.data}
	<div class="desktop-environment" style={backgroundStyle()}>
		<!-- Noise texture overlay -->
		{#if $desktopSettings.showNoise}
			<div class="noise-overlay"></div>
		{/if}

		<!-- Menu Bar -->
		<MenuBar />

		<!-- Desktop Workspace -->
		<div
			bind:this={workspaceElement}
			class="desktop-workspace"
			onclick={handleDesktopClick}
			onmousedown={handleDesktopMouseDown}
			role="application"
			aria-label="Desktop workspace"
		>
			<!-- Desktop Icons - only render when workspace dimensions are known -->
			{#if workspaceWidth > 0 && workspaceHeight > 0}
				{#each $windowStore.desktopIcons as icon (icon.id)}
					{@const pos = getIconPosition(icon)}
					<div
						class="desktop-icon-wrapper"
						class:dragging={draggingIconId === icon.id}
						style="position: absolute; left: {pos.x}px; top: {pos.y}px;"
					>
						<DesktopIcon
							id={icon.id}
							module={icon.module}
							label={icon.label}
							selected={$windowStore.selectedIconIds.includes(icon.id)}
							posX={pos.x}
							posY={pos.y}
							darkBackground={darkBackground}
							onSelect={handleIconSelect}
							onOpen={handleIconOpen}
							onDragStart={handleIconDragStart}
							onDragMove={handleIconDragMove}
							onDragEnd={handleIconDragEnd}
						/>
					</div>
				{/each}
			{/if}

			<!-- Snap Zone Preview Overlay -->
			{#if currentSnapZone}
				{@const preview = snapZonePreview()}
				{#if preview}
					<div
						class="snap-zone-preview"
						style="
							left: {preview.x}px;
							top: {preview.y}px;
							width: {preview.width}px;
							height: {preview.height}px;
						"
					></div>
				{/if}
			{/if}

			<!-- Windows -->
			{#each $visibleWindows as win (win.id)}
				<Window
					window={win}
					focused={$focusedWindow?.id === win.id}
					zIndex={getWindowZIndex(win.id)}
					workspaceHeight={workspaceHeight}
					workspaceWidth={workspaceWidth}
					onsnapZoneChange={handleSnapZoneChange}
				>
					{#snippet children()}
						<div class="window-module-content">
							{#if win.module === 'terminal'}
								<Terminal />
							{:else if win.module === 'desktop-settings'}
								<DesktopSettingsContent />
							{:else if win.module === 'platform'}
								<iframe src="/dashboard" title="Business OS" class="module-iframe"></iframe>
							{:else if win.module === 'dashboard'}
								<iframe src="/dashboard?embed=true" title="Dashboard" class="module-iframe"></iframe>
							{:else if win.module === 'chat'}
								<iframe src="/chat?embed=true" title="Chat" class="module-iframe"></iframe>
							{:else if win.module === 'tasks'}
								<iframe src="/tasks?embed=true" title="Tasks" class="module-iframe"></iframe>
							{:else if win.module === 'projects'}
								<iframe src="/projects?embed=true" title="Projects" class="module-iframe"></iframe>
							{:else if win.module === 'team'}
								<iframe src="/team?embed=true" title="Team" class="module-iframe"></iframe>
							{:else if win.module === 'contexts'}
								<iframe src="/contexts?embed=true" title="Contexts" class="module-iframe"></iframe>
							{:else if win.module === 'nodes'}
								<iframe src="/nodes?embed=true" title="Nodes" class="module-iframe"></iframe>
							{:else if win.module === 'daily'}
								<iframe src="/daily?embed=true" title="Daily Log" class="module-iframe"></iframe>
							{:else if win.module === 'settings'}
								<iframe src="/settings?embed=true" title="Settings" class="module-iframe"></iframe>
							{:else if win.module === 'clients'}
								<iframe src="/clients?embed=true" title="Clients" class="module-iframe"></iframe>
							{:else}
								<div class="module-placeholder">
									<span class="placeholder-icon">
										<svg class="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
										</svg>
									</span>
									<span class="placeholder-text">{win.title}</span>
								</div>
							{/if}
						</div>
					{/snippet}
				</Window>
			{/each}

			<!-- Selection box -->
			{#if isSelecting}
				{@const box = selectionBox()}
				{#if box && box.width > 2 && box.height > 2}
					<div
						class="selection-box"
						style="
							left: {box.x}px;
							top: {box.y}px;
							width: {box.width}px;
							height: {box.height}px;
						"
					></div>
				{/if}
			{/if}
		</div>

		<!-- Dock -->
		<Dock />
	</div>
{/if}

<style>
	.loading-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #E5E5E5;
	}

	.loading-spinner {
		width: 32px;
		height: 32px;
		border: 2px solid #333;
		border-top-color: transparent;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.desktop-environment {
		position: fixed;
		inset: 0;
		overflow: hidden;
	}

	/* Noise texture overlay */
	.noise-overlay {
		position: absolute;
		inset: 0;
		opacity: 0.03;
		pointer-events: none;
		z-index: 1;
		background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)'/%3E%3C/svg%3E");
	}

	.desktop-workspace {
		position: absolute;
		top: 26px; /* Menu bar height */
		left: 0;
		right: 0;
		bottom: 80px; /* Dock area */
		overflow: hidden;
		z-index: 2;
	}

	.window-module-content {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
	}

	.module-iframe {
		width: 100%;
		height: 100%;
		border: none;
		background: white;
	}

	.module-placeholder {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 16px;
		color: #999;
		background: #fafafa;
	}

	.placeholder-icon {
		color: #ccc;
	}

	.placeholder-text {
		font-size: 14px;
		font-weight: 500;
	}

	.desktop-icon-wrapper {
		pointer-events: auto;
	}

	.desktop-icon-wrapper.dragging {
		z-index: 9998;
	}

	/* Selection box (lasso) */
	.selection-box {
		position: absolute;
		background: rgba(0, 102, 255, 0.1);
		border: 1px solid rgba(0, 102, 255, 0.5);
		border-radius: 2px;
		pointer-events: none;
		z-index: 50;
	}

	/* Snap zone preview overlay */
	.snap-zone-preview {
		position: absolute;
		background: rgba(100, 150, 255, 0.15);
		border: 2px solid rgba(100, 150, 255, 0.5);
		border-radius: 8px;
		pointer-events: none;
		z-index: 99;
		transition: all 0.15s ease-out;
		box-shadow: inset 0 0 30px rgba(100, 150, 255, 0.1);
	}
</style>
