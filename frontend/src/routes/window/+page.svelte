<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import { windowStore, visibleWindows, focusedWindow, type SnapZone } from '$lib/stores/windowStore';
	import { desktopSettings, getBackgroundCSS, isBackgroundDark } from '$lib/stores/desktopStore';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { isElectron, isMacOS } from '$lib/utils/platform';

	import MenuBar from '$lib/components/desktop/MenuBar.svelte';
	import DesktopIcon from '$lib/components/desktop/DesktopIcon.svelte';
	import Window from '$lib/components/desktop/Window.svelte';
	import Dock from '$lib/components/desktop/Dock.svelte';
	import Terminal from '$lib/components/desktop/Terminal.svelte';
	import DesktopSettingsContent from '$lib/components/desktop/DesktopSettingsContent.svelte';
	import FolderWindow from '$lib/components/desktop/FolderWindow.svelte';
	import SpotlightSearch from '$lib/components/desktop/SpotlightSearch.svelte';
	import FileBrowser from '$lib/components/desktop/FileBrowser.svelte';

	const APP_VERSION = '0.0.1';
	const session = useSession();

	// Boot screen logic - show full loading on every visit
	let showBootScreen = $state(true);
	let bootComplete = $state(false);

	onMount(() => {
		// Initialize window store to load saved settings from localStorage
		windowStore.initialize();

		// Show loading screen for consistent duration (matches CSS animation)
		setTimeout(() => {
			showBootScreen = false;
			bootComplete = true;
		}, 1000); // 1 second for boot animation
	});

	$effect(() => {
		if (!$session.isPending && bootComplete) {
			showBootScreen = false;
		}
	});

	// Onboarding state
	let showOnboarding = $state(false);
	let onboardingStep = $state(0);

	onMount(() => {
		const hasOnboarded = localStorage.getItem('businessos-onboarded');
		if (!hasOnboarded && !sessionStorage.getItem('businessos-booted')) {
			// Will show onboarding after boot completes
		}
	});

	$effect(() => {
		if (bootComplete && !showBootScreen) {
			const hasOnboarded = localStorage.getItem('businessos-onboarded');
			if (!hasOnboarded) {
				setTimeout(() => {
					showOnboarding = true;
				}, 500);
			}
		}
	});

	function completeOnboarding() {
		localStorage.setItem('businessos-onboarded', 'true');
		showOnboarding = false;
	}

	function nextOnboardingStep() {
		if (onboardingStep < 3) {
			onboardingStep++;
		} else {
			completeOnboarding();
		}
	}

	function skipOnboarding() {
		completeOnboarding();
	}

	// Detect Electron and macOS for traffic light handling
	const inElectron = $derived(browser && isElectron());
	const onMac = $derived(browser && isMacOS());
	const needsTrafficLightSpace = $derived(inElectron && onMac);
	// Menu bar height: 52px in Electron macOS, 26px otherwise
	const menuBarHeight = $derived(needsTrafficLightSpace ? 52 : 26);

	// Workspace dimensions (excluding menu bar and dock)
	let workspaceElement: HTMLDivElement | undefined = $state(undefined);
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

	// Context menu state
	let showContextMenu = $state(false);
	let contextMenuPos = $state({ x: 0, y: 0 });
	let contextMenuType = $state<'desktop' | 'icon'>('desktop');
	let contextMenuIconId = $state<string | null>(null);

	// Rename state
	let renamingIconId = $state<string | null>(null);
	let renameValue = $state('');

	// Spotlight search state
	let showSpotlight = $state(false);

	// Only show icons that are NOT inside a folder
	const visibleDesktopIcons = $derived(
		$windowStore.desktopIcons.filter(icon => !icon.folderId || icon.type === 'folder')
	);

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
		if (!workspaceElement) return;

		// Function to measure and set dimensions
		const measureDimensions = () => {
			if (workspaceElement) {
				const width = workspaceElement.clientWidth;
				const height = workspaceElement.clientHeight;
				if (width > 0 && height > 0) {
					workspaceWidth = width;
					workspaceHeight = height;
				}
			}
		};

		// Initial dimensions
		measureDimensions();

		// Use ResizeObserver for proper dimension tracking
		const resizeObserver = new ResizeObserver((entries) => {
			for (const entry of entries) {
				workspaceWidth = entry.contentRect.width;
				workspaceHeight = entry.contentRect.height;
			}
		});
		resizeObserver.observe(workspaceElement);

		// Also listen for window resize as fallback
		const handleResize = () => measureDimensions();
		window.addEventListener('resize', handleResize);

		// Request animation frame to get accurate initial dimensions
		requestAnimationFrame(measureDimensions);

		// Double-check after a short delay to catch any layout shifts
		const delayedMeasure = setTimeout(measureDimensions, 100);

		return () => {
			resizeObserver.disconnect();
			window.removeEventListener('resize', handleResize);
			clearTimeout(delayedMeasure);
		};
	});

	// Re-measure when boot completes and session is ready
	$effect(() => {
		if (bootComplete && $session.data && workspaceElement) {
			// Give DOM time to render the workspace
			requestAnimationFrame(() => {
				if (workspaceElement) {
					const width = workspaceElement.clientWidth;
					const height = workspaceElement.clientHeight;
					if (width > 0 && height > 0) {
						workspaceWidth = width;
						workspaceHeight = height;
					}
				}
			});
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
			const isShift = event.shiftKey;
			const isCtrlAlt = event.ctrlKey && event.altKey;

			if (isMeta && event.key === ' ') {
				// Cmd+Space - Open Spotlight
				event.preventDefault();
				showSpotlight = true;
			} else if (isMeta && event.key === 'w') {
				event.preventDefault();
				if ($focusedWindow) {
					windowStore.closeWindow($focusedWindow.id);
				}
			} else if (isMeta && event.key === 'm') {
				event.preventDefault();
				if ($focusedWindow) {
					windowStore.minimizeWindow($focusedWindow.id);
				}
			} else if (isMeta && isShift && event.key === 'F') {
				// Cmd+Shift+F - Maximize/Restore
				event.preventDefault();
				if ($focusedWindow) {
					windowStore.toggleMaximize($focusedWindow.id);
				}
			} else if (isMeta && event.key === '`' && !isShift) {
				event.preventDefault();
				windowStore.cycleWindows();
			} else if (isMeta && isShift && event.key === '`') {
				// Cmd+Shift+` - Toggle Terminal
				event.preventDefault();
				windowStore.openWindow('terminal');
			} else if (isCtrlAlt && event.key === 'ArrowLeft') {
				// Ctrl+Alt+Left - Snap left
				event.preventDefault();
				if ($focusedWindow) {
					windowStore.snapWindow($focusedWindow.id, 'left', workspaceWidth, workspaceHeight);
				}
			} else if (isCtrlAlt && event.key === 'ArrowRight') {
				// Ctrl+Alt+Right - Snap right
				event.preventDefault();
				if ($focusedWindow) {
					windowStore.snapWindow($focusedWindow.id, 'right', workspaceWidth, workspaceHeight);
				}
			} else if (isMeta && isShift && event.key === 'T') {
				// Cmd+Shift+T - New Task
				event.preventDefault();
				windowStore.openWindow('tasks');
			} else if (isMeta && isShift && event.key === 'P') {
				// Cmd+Shift+P - New Project
				event.preventDefault();
				windowStore.openWindow('projects');
			} else if (isMeta && isShift && event.key === 'N') {
				// Cmd+Shift+N - New Note
				event.preventDefault();
				windowStore.openWindow('contexts');
			} else if (isMeta && event.key === '1') {
				// Cmd+1 - Dashboard
				event.preventDefault();
				windowStore.openWindow('dashboard');
			} else if (isMeta && event.key === '2') {
				// Cmd+2 - Chat
				event.preventDefault();
				windowStore.openWindow('chat');
			} else if (isMeta && event.key === '3') {
				// Cmd+3 - Tasks
				event.preventDefault();
				windowStore.openWindow('tasks');
			} else if (isMeta && event.key === '4') {
				// Cmd+4 - Calendar
				event.preventDefault();
				windowStore.openWindow('calendar');
			} else if (isMeta && event.key === '5') {
				// Cmd+5 - Projects
				event.preventDefault();
				windowStore.openWindow('projects');
			} else if (event.key === 'Escape') {
				windowStore.clearIconSelection();
				showSpotlight = false;
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
		// Check if dropped on a folder
		const folderIcon = visibleDesktopIcons.find(icon => {
			if (icon.type !== 'folder' || icon.id === iconId) return false;
			const folderPos = getIconPosition(icon);
			// Check if finalX,finalY is within the folder icon bounds
			const inX = finalX >= folderPos.x - 20 && finalX <= folderPos.x + 80;
			const inY = finalY >= folderPos.y - 20 && finalY <= folderPos.y + 100;
			return inX && inY;
		});

		if (folderIcon && folderIcon.folderId) {
			// Move icon into folder
			windowStore.moveIconToFolder(iconId, folderIcon.folderId);
			// Clear from iconPositions since it's now in a folder
			const { [iconId]: _, ...rest } = iconPositions;
			iconPositions = rest;
		} else {
			// Store the final pixel position
			const constrainedX = Math.max(0, Math.min(finalX, workspaceWidth - 80));
			const constrainedY = Math.max(0, Math.min(finalY, workspaceHeight - 100));

			iconPositions = {
				...iconPositions,
				[iconId]: { x: constrainedX, y: constrainedY }
			};
		}
		draggingIconId = null;
	}

	// Get z-index for a window based on its position in windowOrder
	function getWindowZIndex(windowId: string): number {
		const index = $windowStore.windowOrder.indexOf(windowId);
		return 100 + index;
	}

	// Context menu handlers
	function handleContextMenu(event: MouseEvent) {
		// Only show on desktop workspace (not on icons)
		if (!(event.target as HTMLElement).classList.contains('desktop-workspace')) return;

		event.preventDefault();
		contextMenuPos = { x: event.clientX, y: event.clientY };
		contextMenuType = 'desktop';
		contextMenuIconId = null;
		showContextMenu = true;
	}

	function handleIconContextMenu(event: MouseEvent, iconId: string) {
		event.preventDefault();
		event.stopPropagation();
		contextMenuPos = { x: event.clientX, y: event.clientY };
		contextMenuType = 'icon';
		contextMenuIconId = iconId;
		showContextMenu = true;
	}

	function closeContextMenu() {
		showContextMenu = false;
		contextMenuIconId = null;
	}

	function createNewFolder() {
		// Calculate grid position from click position
		const relativeX = contextMenuPos.x;
		const relativeY = contextMenuPos.y - menuBarHeight;

		const gridX = Math.floor(relativeX / GRID_SIZE);
		const gridY = Math.floor(relativeY / GRID_SIZE);

		windowStore.createFolder('New Folder', gridX, gridY);
		closeContextMenu();
	}

	function openDesktopSettings() {
		windowStore.openWindow('desktop-settings');
		closeContextMenu();
	}

	function arrangeIcons() {
		// Reset icon positions to default grid
		iconPositions = {};
		closeContextMenu();
	}

	function startRenameIcon() {
		if (!contextMenuIconId) return;
		const icon = $windowStore.desktopIcons.find(i => i.id === contextMenuIconId);
		if (icon) {
			renameValue = icon.label;
			renamingIconId = contextMenuIconId;
		}
		closeContextMenu();
	}

	function finishRename() {
		if (!renamingIconId || !renameValue.trim()) {
			renamingIconId = null;
			return;
		}

		const icon = $windowStore.desktopIcons.find(i => i.id === renamingIconId);
		if (icon?.type === 'folder' && icon.folderId) {
			windowStore.renameFolder(icon.folderId, renameValue.trim());
		}
		renamingIconId = null;
	}

	function pinToDock() {
		if (!contextMenuIconId) return;
		const icon = $windowStore.desktopIcons.find(i => i.id === contextMenuIconId);
		if (icon) {
			// For folders, use the full folder module string
			if (icon.type === 'folder' && icon.folderId) {
				windowStore.addToDock(`folder-${icon.folderId}`);
			} else {
				windowStore.addToDock(icon.module);
			}
		}
		closeContextMenu();
	}

	function deleteFolder() {
		if (!contextMenuIconId) return;
		const icon = $windowStore.desktopIcons.find(i => i.id === contextMenuIconId);
		if (icon?.type === 'folder' && icon.folderId) {
			windowStore.deleteFolder(icon.folderId);
		}
		closeContextMenu();
	}

	function openIcon() {
		if (!contextMenuIconId) return;
		const icon = $windowStore.desktopIcons.find(i => i.id === contextMenuIconId);
		if (icon) {
			if (icon.type === 'folder' && icon.folderId) {
				windowStore.openFolder(icon.folderId);
			} else {
				windowStore.openWindow(icon.module);
			}
		}
		closeContextMenu();
	}

	// Get icon for context menu
	const contextMenuIcon = $derived(
		contextMenuIconId ? $windowStore.desktopIcons.find(i => i.id === contextMenuIconId) : null
	);
</script>

<svelte:head>
	<title>Business OS - Desktop</title>
</svelte:head>

{#if showBootScreen}
	<div class="boot-screen">
		<div class="grid-overlay"></div>
		<div class="boot-content">
			<div class="os-logo">
				<span class="logo-name">{$desktopSettings.companyName}</span>
				<span class="logo-os">OS</span>
			</div>
			<div class="boot-terminal">
				<div class="terminal-line line-1">
					<span class="prompt">$</span>
					<span class="cmd">init</span>
					<span class="args">--workspace</span>
				</div>
				<div class="terminal-line line-2">
					<span class="output">Loading modules</span>
					<span class="cursor">█</span>
				</div>
			</div>
			<div class="boot-loader">
				<div class="loader-segments">
					<div class="segment s1"></div>
					<div class="segment s2"></div>
					<div class="segment s3"></div>
					<div class="segment s4"></div>
					<div class="segment s5"></div>
				</div>
			</div>
		</div>
		<div class="boot-footer">
			<a href="https://osa.dev" target="_blank" rel="noopener noreferrer" class="osa-link">
				<img src="/osa-logo.png" alt="OSA" class="osa-logo" />
			</a>
			<span class="version">v{APP_VERSION}</span>
		</div>
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
			style="top: {menuBarHeight}px;"
			onclick={handleDesktopClick}
			onmousedown={handleDesktopMouseDown}
			oncontextmenu={handleContextMenu}
			role="application"
			aria-label="Desktop workspace"
		>
			<!-- Desktop Icons - only render when workspace dimensions are known -->
			{#if workspaceWidth > 0 && workspaceHeight > 0}
				{#each visibleDesktopIcons as icon (icon.id)}
					{@const pos = getIconPosition(icon)}
					<div
						class="desktop-icon-wrapper"
						class:dragging={draggingIconId === icon.id}
						style="position: absolute; left: {pos.x}px; top: {pos.y}px;"
						oncontextmenu={(e) => handleIconContextMenu(e, icon.id)}
					>
						{#if renamingIconId === icon.id}
							<!-- Rename input -->
							<div class="rename-container">
								<div class="rename-icon-preview" style="background: {icon.folderColor || '#3B82F6'}20">
									<svg viewBox="0 0 24 24" fill={icon.folderColor || '#3B82F6'}>
										<path d="M3 7V17C3 18.1046 3.89543 19 5 19H19C20.1046 19 21 18.1046 21 17V9C21 7.89543 20.1046 7 19 7H12L10 5H5C3.89543 5 3 5.89543 3 7Z"/>
									</svg>
								</div>
								<input
									type="text"
									class="rename-input"
									bind:value={renameValue}
									onblur={finishRename}
									onkeydown={(e) => {
										if (e.key === 'Enter') finishRename();
										if (e.key === 'Escape') { renamingIconId = null; }
									}}
									autofocus
								/>
							</div>
						{:else}
							<DesktopIcon
								id={icon.id}
								module={icon.module}
								label={icon.label}
								selected={$windowStore.selectedIconIds.includes(icon.id)}
								posX={pos.x}
								posY={pos.y}
								darkBackground={darkBackground}
								iconType={icon.type || 'app'}
								folderId={icon.type === 'folder' ? icon.folderId : undefined}
								folderColor={icon.folderColor}
								onSelect={handleIconSelect}
								onOpen={handleIconOpen}
								onDragStart={handleIconDragStart}
								onDragMove={handleIconDragMove}
								onDragEnd={handleIconDragEnd}
							/>
						{/if}
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
							{:else if win.module.startsWith('folder-')}
								<FolderWindow folderId={win.module.replace('folder-', '')} />
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
							{:else if win.module === 'calendar'}
								<iframe src="/calendar?embed=true" title="Calendar" class="module-iframe"></iframe>
							{:else if win.module === 'ai-settings'}
								<iframe src="/settings/ai?embed=true" title="AI Settings" class="module-iframe"></iframe>
							{:else if win.module === 'files'}
								<FileBrowser />
							{:else if win.module === 'finder'}
								<FileBrowser />
							{:else if win.module === 'help'}
								<iframe src="/help?embed=true" title="Help" class="module-iframe"></iframe>
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

		<!-- Context Menu -->
		{#if showContextMenu}
			<div
				class="context-menu-overlay"
				onclick={closeContextMenu}
				oncontextmenu={(e) => { e.preventDefault(); closeContextMenu(); }}
				role="presentation"
			></div>
			<div
				class="context-menu"
				style="left: {contextMenuPos.x}px; top: {contextMenuPos.y}px;"
			>
				{#if contextMenuType === 'icon' && contextMenuIcon}
					<!-- Icon Context Menu -->
					<button class="context-menu-item" onclick={openIcon}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
							<path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
							<path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
						</svg>
						Open
					</button>
					{#if contextMenuIcon.type === 'folder'}
						<button class="context-menu-item" onclick={startRenameIcon}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
								<path d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
							</svg>
							Rename
						</button>
					{/if}
					<div class="context-menu-separator"></div>
					<button class="context-menu-item" onclick={pinToDock}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
							<path d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"/>
						</svg>
						Add to Dock
					</button>
					{#if contextMenuIcon.type === 'folder'}
						<div class="context-menu-separator"></div>
						<button class="context-menu-item context-menu-item--danger" onclick={deleteFolder}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
								<path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
							</svg>
							Delete Folder
						</button>
					{/if}
				{:else}
					<!-- Desktop Context Menu -->
					<button class="context-menu-item" onclick={createNewFolder}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
							<path d="M3 7V17C3 18.1046 3.89543 19 5 19H19C20.1046 19 21 18.1046 21 17V9C21 7.89543 20.1046 7 19 7H12L10 5H5C3.89543 5 3 5.89543 3 7Z"/>
						</svg>
						New Folder
					</button>
					<div class="context-menu-separator"></div>
					<button class="context-menu-item" onclick={arrangeIcons}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
							<path d="M4 6h4M4 10h4M4 14h4M4 18h4M10 6h10M10 10h10M10 14h10M10 18h10"/>
						</svg>
						Arrange Icons
					</button>
					<button class="context-menu-item" onclick={openDesktopSettings}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
							<path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
							<circle cx="12" cy="12" r="3"/>
						</svg>
						Desktop Settings
					</button>
				{/if}
			</div>
		{/if}

		<!-- Dock -->
		<Dock />

		<!-- Spotlight Search -->
		<SpotlightSearch open={showSpotlight} onClose={() => showSpotlight = false} />

		<!-- Onboarding Overlay -->
		{#if showOnboarding}
			<div class="onboarding-overlay">
				<div class="onboarding-backdrop"></div>

				<!-- Spotlight highlights based on step -->
				{#if onboardingStep === 0}
					<!-- Highlight Dock -->
					<div class="spotlight-highlight dock-highlight"></div>
					<div class="onboarding-tooltip dock-tooltip">
						<div class="tooltip-content">
							<h3>Your Dock</h3>
							<p>Quick access to all your apps. Click any icon to open it in a window. You can also use the chat bubble for quick AI assistance.</p>
						</div>
						<div class="tooltip-footer">
							<span class="step-indicator">1 of 4</span>
							<div class="tooltip-actions">
								<button class="skip-btn" onclick={skipOnboarding}>Skip</button>
								<button class="next-btn" onclick={nextOnboardingStep}>Next</button>
							</div>
						</div>
					</div>
				{:else if onboardingStep === 1}
					<!-- Highlight Menu Bar -->
					<div class="spotlight-highlight menubar-highlight"></div>
					<div class="onboarding-tooltip menubar-tooltip">
						<div class="tooltip-content">
							<h3>Menu Bar</h3>
							<p>Access system functions, view time, and control your workspace from here. Click the company name for quick navigation.</p>
						</div>
						<div class="tooltip-footer">
							<span class="step-indicator">2 of 4</span>
							<div class="tooltip-actions">
								<button class="skip-btn" onclick={skipOnboarding}>Skip</button>
								<button class="next-btn" onclick={nextOnboardingStep}>Next</button>
							</div>
						</div>
					</div>
				{:else if onboardingStep === 2}
					<!-- Highlight Desktop Icons -->
					<div class="spotlight-highlight icons-highlight"></div>
					<div class="onboarding-tooltip icons-tooltip">
						<div class="tooltip-content">
							<h3>Desktop Icons</h3>
							<p>Double-click icons to open apps. Drag to rearrange. Right-click for more options like creating folders.</p>
						</div>
						<div class="tooltip-footer">
							<span class="step-indicator">3 of 4</span>
							<div class="tooltip-actions">
								<button class="skip-btn" onclick={skipOnboarding}>Skip</button>
								<button class="next-btn" onclick={nextOnboardingStep}>Next</button>
							</div>
						</div>
					</div>
				{:else if onboardingStep === 3}
					<!-- Final welcome -->
					<div class="onboarding-welcome">
						<div class="welcome-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
								<path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
							</svg>
						</div>
						<h2>You're All Set!</h2>
						<p>Press <kbd>⌘</kbd> + <kbd>Space</kbd> anytime to open Spotlight search and find anything quickly.</p>
						<button class="get-started-btn" onclick={completeOnboarding}>
							Get Started
						</button>
					</div>
				{/if}
			</div>
		{/if}
	</div>
{/if}

<style>
	.boot-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: #fafafa;
		position: relative;
		overflow: hidden;
	}

	.grid-overlay {
		position: absolute;
		inset: 0;
		background-image:
			linear-gradient(rgba(0,0,0,0.02) 1px, transparent 1px),
			linear-gradient(90deg, rgba(0,0,0,0.02) 1px, transparent 1px);
		background-size: 20px 20px;
		pointer-events: none;
	}

	.boot-content {
		text-align: center;
		z-index: 1;
	}

	.os-logo {
		font-family: 'SF Mono', 'Monaco', 'Fira Code', monospace;
		font-size: 42px;
		font-weight: 800;
		letter-spacing: 6px;
		margin-bottom: 40px;
		display: flex;
		align-items: baseline;
		justify-content: center;
		gap: 2px;
	}

	.logo-name {
		color: #111;
		animation: glitch-text 0.3s ease-out;
	}

	.logo-os {
		color: #111;
		font-weight: 400;
		opacity: 0.4;
		font-size: 36px;
	}

	@keyframes glitch-text {
		0% { opacity: 0; transform: translateX(-10px); }
		20% { opacity: 1; transform: translateX(2px); }
		40% { transform: translateX(-1px); }
		60% { transform: translateX(1px); }
		100% { transform: translateX(0); }
	}

	.boot-terminal {
		font-family: 'SF Mono', 'Monaco', monospace;
		font-size: 13px;
		margin-bottom: 32px;
		text-align: left;
		display: inline-block;
	}

	.terminal-line {
		display: flex;
		align-items: center;
		gap: 8px;
		height: 22px;
		opacity: 0;
		animation: type-line 0.2s ease-out forwards;
	}

	.line-1 { animation-delay: 0.1s; }
	.line-2 { animation-delay: 0.3s; }

	@keyframes type-line {
		from { opacity: 0; transform: translateY(5px); }
		to { opacity: 1; transform: translateY(0); }
	}

	.prompt {
		color: #999;
	}

	.cmd {
		color: #111;
		font-weight: 600;
	}

	.args {
		color: #666;
	}

	.output {
		color: #888;
	}

	.cursor {
		color: #111;
		animation: blink-cursor 0.6s step-end infinite;
		font-size: 12px;
	}

	@keyframes blink-cursor {
		0%, 100% { opacity: 1; }
		50% { opacity: 0; }
	}

	.boot-loader {
		display: flex;
		justify-content: center;
	}

	.loader-segments {
		display: flex;
		gap: 4px;
	}

	.segment {
		width: 24px;
		height: 3px;
		background: #e0e0e0;
		position: relative;
		overflow: hidden;
	}

	.segment::after {
		content: '';
		position: absolute;
		inset: 0;
		background: #111;
		transform: translateX(-100%);
		animation: segment-fill 0.5s ease-out forwards;
	}

	.s1::after { animation-delay: 0s; }
	.s2::after { animation-delay: 0.08s; }
	.s3::after { animation-delay: 0.16s; }
	.s4::after { animation-delay: 0.24s; }
	.s5::after { animation-delay: 0.32s; }

	@keyframes segment-fill {
		to { transform: translateX(0); }
	}

	.boot-footer {
		position: absolute;
		bottom: 32px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		opacity: 0;
		animation: fade-in 0.3s ease-out 0.5s forwards;
	}

	.osa-link {
		display: block;
		transition: opacity 0.15s ease, transform 0.15s ease;
	}

	.osa-link:hover {
		opacity: 0.8;
		transform: scale(1.05);
	}

	.osa-logo {
		height: 56px;
		width: auto;
		opacity: 0.5;
	}

	.version {
		font-family: 'SF Mono', 'Monaco', monospace;
		font-size: 10px;
		color: #aaa;
		letter-spacing: 1px;
	}

	@keyframes fade-in {
		from { opacity: 0; transform: translateX(-50%) translateY(5px); }
		to { opacity: 1; transform: translateX(-50%) translateY(0); }
	}

	/* ===== DARK MODE BOOT SCREEN ===== */
	:global(.dark) .boot-screen {
		background: #1c1c1e;
	}

	:global(.dark) .grid-overlay {
		background-image:
			linear-gradient(rgba(255,255,255,0.03) 1px, transparent 1px),
			linear-gradient(90deg, rgba(255,255,255,0.03) 1px, transparent 1px);
	}

	:global(.dark) .logo-name {
		color: #f5f5f7;
	}

	:global(.dark) .logo-os {
		color: #f5f5f7;
		opacity: 0.4;
	}

	:global(.dark) .prompt {
		color: #6e6e73;
	}

	:global(.dark) .cmd {
		color: #f5f5f7;
	}

	:global(.dark) .args {
		color: #a1a1a6;
	}

	:global(.dark) .output {
		color: #a1a1a6;
	}

	:global(.dark) .cursor {
		color: #0A84FF;
	}

	:global(.dark) .segment {
		background: #3a3a3c;
	}

	:global(.dark) .segment::after {
		background: #0A84FF;
	}

	:global(.dark) .version {
		color: #6e6e73;
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
		/* top is set dynamically via inline style */
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

	/* Context Menu */
	.context-menu-overlay {
		position: fixed;
		inset: 0;
		z-index: 9998;
	}

	.context-menu {
		position: fixed;
		z-index: 9999;
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-radius: 8px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(0, 0, 0, 0.05);
		padding: 4px;
		min-width: 180px;
	}

	.context-menu-item {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		padding: 8px 12px;
		border: none;
		background: none;
		cursor: pointer;
		font-size: 13px;
		color: #333;
		border-radius: 4px;
		text-align: left;
	}

	.context-menu-item:hover {
		background: rgba(0, 102, 255, 0.1);
		color: #0066FF;
	}

	.context-menu-item svg {
		width: 16px;
		height: 16px;
		flex-shrink: 0;
	}

	.context-menu-separator {
		height: 1px;
		background: rgba(0, 0, 0, 0.1);
		margin: 4px 8px;
	}

	.context-menu-item--danger {
		color: #DC2626;
	}

	.context-menu-item--danger:hover {
		background: rgba(220, 38, 38, 0.1);
		color: #B91C1C;
	}

	/* Rename container */
	.rename-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		padding: 8px;
		width: 100px;
	}

	.rename-icon-preview {
		width: 56px;
		height: 56px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}

	.rename-icon-preview svg {
		width: 28px;
		height: 28px;
	}

	.rename-input {
		width: 100%;
		padding: 4px 6px;
		font-size: 11px;
		border: 1px solid #3B82F6;
		border-radius: 4px;
		text-align: center;
		outline: none;
		background: white;
	}

	.rename-input:focus {
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.3);
	}

	/* Onboarding Overlay */
	.onboarding-overlay {
		position: fixed;
		inset: 0;
		z-index: 10000;
		pointer-events: auto;
	}

	.onboarding-backdrop {
		position: absolute;
		inset: 0;
		background: transparent;
		pointer-events: none;
	}

	.spotlight-highlight {
		position: fixed;
		background: transparent;
		box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.75);
		z-index: 10001;
		pointer-events: none;
		animation: pulse-highlight 2s ease-in-out infinite;
	}

	@keyframes pulse-highlight {
		0%, 100% {
			box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.75), 0 0 0 4px rgba(255, 255, 255, 0.3);
		}
		50% {
			box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.75), 0 0 0 8px rgba(255, 255, 255, 0.5);
		}
	}

	.dock-highlight {
		bottom: 8px;
		left: 50%;
		transform: translateX(-50%);
		width: calc(100% - 200px);
		max-width: 800px;
		height: 72px;
		border-radius: 20px;
	}

	.menubar-highlight {
		top: 0;
		left: 0;
		right: 0;
		height: 52px;
		border-radius: 0;
		background: transparent;
	}

	.icons-highlight {
		top: 80px;
		left: 20px;
		width: 120px;
		height: 350px;
		border-radius: 16px;
	}

	.onboarding-tooltip {
		position: absolute;
		background: white;
		border-radius: 16px;
		padding: 20px;
		width: 320px;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
		z-index: 10002;
		animation: tooltip-in 0.3s ease-out;
	}

	@keyframes tooltip-in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.dock-tooltip {
		bottom: 100px;
		left: 50%;
		transform: translateX(-50%);
	}

	.menubar-tooltip {
		top: 70px;
		left: 50%;
		transform: translateX(-50%);
	}

	.icons-tooltip {
		top: 120px;
		left: 160px;
	}

	.tooltip-content h3 {
		font-size: 18px;
		font-weight: 700;
		color: #111;
		margin: 0 0 8px 0;
	}

	.tooltip-content p {
		font-size: 14px;
		color: #666;
		line-height: 1.5;
		margin: 0;
	}

	.tooltip-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-top: 20px;
		padding-top: 16px;
		border-top: 1px solid #eee;
	}

	.step-indicator {
		font-size: 12px;
		color: #999;
		font-weight: 500;
	}

	.tooltip-actions {
		display: flex;
		gap: 10px;
	}

	.skip-btn {
		padding: 8px 16px;
		font-size: 13px;
		font-weight: 500;
		color: #666;
		background: none;
		border: none;
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.15s;
	}

	.skip-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	.next-btn {
		padding: 8px 20px;
		font-size: 13px;
		font-weight: 600;
		color: white;
		background: #111;
		border: none;
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.15s;
	}

	.next-btn:hover {
		background: #333;
	}

	/* Welcome screen (final step) */
	.onboarding-welcome {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: white;
		border-radius: 24px;
		padding: 48px;
		text-align: center;
		width: 400px;
		box-shadow: 0 30px 80px rgba(0, 0, 0, 0.3);
		z-index: 10002;
		animation: welcome-in 0.4s ease-out;
	}

	@keyframes welcome-in {
		from {
			opacity: 0;
			transform: translate(-50%, -50%) scale(0.9);
		}
		to {
			opacity: 1;
			transform: translate(-50%, -50%) scale(1);
		}
	}

	.welcome-icon {
		width: 64px;
		height: 64px;
		margin: 0 auto 24px;
		background: linear-gradient(135deg, #10B981 0%, #059669 100%);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.welcome-icon svg {
		width: 36px;
		height: 36px;
		color: white;
	}

	.onboarding-welcome h2 {
		font-size: 28px;
		font-weight: 700;
		color: #111;
		margin: 0 0 12px 0;
	}

	.onboarding-welcome p {
		font-size: 15px;
		color: #666;
		line-height: 1.6;
		margin: 0 0 28px 0;
	}

	.onboarding-welcome kbd {
		display: inline-block;
		padding: 4px 8px;
		font-family: 'SF Mono', Monaco, monospace;
		font-size: 12px;
		background: #f0f0f0;
		border: 1px solid #ddd;
		border-radius: 4px;
		box-shadow: 0 1px 2px rgba(0,0,0,0.1);
	}

	.get-started-btn {
		padding: 14px 32px;
		font-size: 15px;
		font-weight: 600;
		color: white;
		background: #111;
		border: none;
		cursor: pointer;
		border-radius: 12px;
		transition: all 0.15s;
	}

	.get-started-btn:hover {
		background: #333;
		transform: translateY(-1px);
	}
</style>
