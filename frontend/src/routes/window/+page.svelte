<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSession } from '$lib/auth-client';
	import { windowStore, visibleWindows, focusedWindow, type SnapZone } from '$lib/stores/windowStore';
	import { desktopSettings, getBackgroundCSS, isBackgroundDark } from '$lib/stores/desktopStore';
	import { deployedAppsStore } from '$lib/stores/deployedAppsStore';
	import { currentWorkspaceId } from '$lib/stores/workspaces';
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { isElectron, isMacOS } from '$lib/utils/platform';

	import MenuBar from '$lib/components/desktop/MenuBar.svelte';
	import DesktopIcon from '$lib/components/desktop/DesktopIcon.svelte';
	import Window from '$lib/components/desktop/Window.svelte';
	import Dock from '$lib/components/desktop/Dock.svelte';
	import SpotlightSearch from '$lib/components/desktop/SpotlightSearch.svelte';
	import IconPicker from '$lib/components/desktop/IconPicker.svelte';
	import AnimatedBackground from '$lib/components/desktop/AnimatedBackground.svelte';
	import Desktop3D from '$lib/components/desktop3d/Desktop3D.svelte';
	import OsaOrb from '$lib/components/osa/OsaOrb.svelte';
	import type { CustomIconConfig } from '$lib/stores/windowStore';

	import BootScreen from '$lib/components/window/BootScreen.svelte';
	import DesktopContextMenu from '$lib/components/window/DesktopContextMenu.svelte';
	import DesktopOnboarding from '$lib/components/window/DesktopOnboarding.svelte';
	import WindowContent from '$lib/components/window/WindowContent.svelte';

	const APP_VERSION = '0.0.1';
	const session = useSession();

	// Boot screen logic - show full loading on every visit
	let showBootScreen = $state(true);
	let bootComplete = $state(false);

	onMount(() => {
		// Initialize window store to load saved settings from localStorage
		windowStore.initialize();

		// Start discovering deployed OSA apps and user-generated apps
		const workspaceId = $currentWorkspaceId || undefined;
		deployedAppsStore.startDiscovery(workspaceId);

		// Show loading screen for consistent duration (matches CSS animation)
		setTimeout(() => {
			showBootScreen = false;
			bootComplete = true;
		}, 1000); // 1 second for boot animation
	});

	onDestroy(() => {
		// Stop discovery when component unmounts
		deployedAppsStore.stopDiscovery();
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

	// Icon picker state
	let showIconPicker = $state(false);
	let customizeIconId = $state<string | null>(null);
	let customizeIconCurrentConfig = $state<CustomIconConfig | undefined>(undefined);

	// Handler for icon customization
	function handleCustomizeIcon(iconId: string) {
		customizeIconId = iconId;
		const icon = $windowStore.desktopIcons.find(i => i.id === iconId);
		customizeIconCurrentConfig = icon?.customIcon;
		showIconPicker = true;
	}

	function handleIconPickerSelect(config: CustomIconConfig | undefined) {
		if (customizeIconId) {
			windowStore.updateIconCustomization(customizeIconId, config);
		}
		showIconPicker = false;
		customizeIconId = null;
		customizeIconCurrentConfig = undefined;
	}

	function handleIconPickerClose() {
		showIconPicker = false;
		customizeIconId = null;
		customizeIconCurrentConfig = undefined;
	}

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

		measureDimensions();

		const resizeObserver = new ResizeObserver((entries) => {
			for (const entry of entries) {
				workspaceWidth = entry.contentRect.width;
				workspaceHeight = entry.contentRect.height;
			}
		});
		resizeObserver.observe(workspaceElement);

		const handleResize = () => measureDimensions();
		window.addEventListener('resize', handleResize);

		requestAnimationFrame(measureDimensions);

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
			const fitMap: Record<string, string> = {
				'cover': 'cover',
				'contain': 'contain',
				'fill': '100% 100%',
				'center': 'auto'
			};
			const bgSize = fitMap[$desktopSettings.backgroundFit] || 'cover';

			return `
				background-image: ${bgCSS.background};
				background-size: ${bgSize};
				background-position: center center;
				background-repeat: no-repeat;
				background-attachment: fixed;
				background-color: var(--dbg);
			`;
		} else if (bgCSS.backgroundSize) {
			return `background: ${bgCSS.background}; background-size: ${bgCSS.backgroundSize};`;
		}
		return `background: ${bgCSS.background};`;
	});

	onMount(() => {
		function updateDimensions() {
			if (workspaceElement) {
				workspaceWidth = workspaceElement.clientWidth;
				workspaceHeight = workspaceElement.clientHeight;
			}
		}

		window.addEventListener('resize', updateDimensions);

		function handleKeyDown(event: KeyboardEvent) {
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
				event.preventDefault();
				showSpotlight = true;
			} else if (isMeta && event.key === 'w') {
				event.preventDefault();
				if ($focusedWindow) windowStore.closeWindow($focusedWindow.id);
			} else if (isMeta && event.key === 'm') {
				event.preventDefault();
				if ($focusedWindow) windowStore.minimizeWindow($focusedWindow.id);
			} else if (isMeta && isShift && event.key === 'F') {
				event.preventDefault();
				if ($focusedWindow) windowStore.toggleMaximize($focusedWindow.id);
			} else if (isMeta && event.key === '`' && !isShift) {
				event.preventDefault();
				windowStore.cycleWindows();
			} else if (isMeta && isShift && event.key === '`') {
				event.preventDefault();
				windowStore.openWindow('terminal');
			} else if (isCtrlAlt && event.key === 'ArrowLeft') {
				event.preventDefault();
				if ($focusedWindow) windowStore.snapWindow($focusedWindow.id, 'left', workspaceWidth, workspaceHeight);
			} else if (isCtrlAlt && event.key === 'ArrowRight') {
				event.preventDefault();
				if ($focusedWindow) windowStore.snapWindow($focusedWindow.id, 'right', workspaceWidth, workspaceHeight);
			} else if (isMeta && isShift && event.key === 'T') {
				event.preventDefault();
				windowStore.openWindow('tasks');
			} else if (isMeta && isShift && event.key === 'P') {
				event.preventDefault();
				windowStore.openWindow('projects');
			} else if (isMeta && isShift && event.key === 'N') {
				event.preventDefault();
				windowStore.openWindow('contexts');
			} else if (isMeta && event.key === '1') {
				event.preventDefault();
				windowStore.openWindow('dashboard');
			} else if (isMeta && event.key === '2') {
				event.preventDefault();
				windowStore.openWindow('chat');
			} else if (isMeta && event.key === '3') {
				event.preventDefault();
				windowStore.openWindow('tasks');
			} else if (isMeta && event.key === '4') {
				event.preventDefault();
				windowStore.openWindow('calendar');
			} else if (isMeta && event.key === '5') {
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
		if (iconPositions[icon.id]) {
			return iconPositions[icon.id];
		}

		let x: number;
		let y: number;

		if (icon.x < 0) {
			x = workspaceWidth + (icon.x * GRID_SIZE) - ICON_PADDING;
		} else {
			x = icon.x * GRID_SIZE + ICON_PADDING;
		}

		if (icon.y < 0) {
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
		const constrainedX = Math.max(0, Math.min(newX, workspaceWidth - 80));
		const constrainedY = Math.max(0, Math.min(newY, workspaceHeight - 100));

		iconPositions = {
			...iconPositions,
			[iconId]: { x: constrainedX, y: constrainedY }
		};
	}

	function handleDesktopClick(event: MouseEvent) {
		if (didSelectionDrag) {
			didSelectionDrag = false;
			return;
		}
		const target = event.target as HTMLElement;
		const isDesktopOrBackground = target.classList.contains('desktop-workspace') ||
			target.classList.contains('animated-background') ||
			target.tagName === 'CANVAS';
		const isNotIconOrWindow = !target.closest('.desktop-icon') &&
			!target.closest('.window') &&
			!target.closest('.context-menu');

		if (isDesktopOrBackground || (target.closest('.desktop-workspace') && isNotIconOrWindow)) {
			windowStore.clearIconSelection();
		}
	}

	// Selection box handlers
	function handleDesktopMouseDown(event: MouseEvent) {
		if (event.button !== 0) return;
		if (!(event.target as HTMLElement).classList.contains('desktop-workspace')) return;
		if (!workspaceElement) return;

		const rect = workspaceElement.getBoundingClientRect();
		const x = event.clientX - rect.left;
		const y = event.clientY - rect.top;

		selectionStart = { x, y };
		selectionEnd = { x, y };
		isSelecting = true;
		didSelectionDrag = false;

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
		const folderIcon = visibleDesktopIcons.find(icon => {
			if (icon.type !== 'folder' || icon.id === iconId) return false;
			const folderPos = getIconPosition(icon);
			const inX = finalX >= folderPos.x - 20 && finalX <= folderPos.x + 80;
			const inY = finalY >= folderPos.y - 20 && finalY <= folderPos.y + 100;
			return inX && inY;
		});

		if (folderIcon && folderIcon.folderId) {
			windowStore.moveIconToFolder(iconId, folderIcon.folderId);
			const { [iconId]: _, ...rest } = iconPositions;
			iconPositions = rest;
		} else {
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

	const contextMenuIcon = $derived(
		contextMenuIconId ? $windowStore.desktopIcons.find(i => i.id === contextMenuIconId) : null
	);
</script>

<svelte:head>
	<title>Business OS - Desktop</title>
</svelte:head>

{#if showBootScreen}
	<BootScreen
		companyName={$desktopSettings.companyName}
		appVersion={APP_VERSION}
	/>
{:else if $session.data}
	<!-- 3D Desktop Mode -->
	{#if $desktopSettings.enable3DDesktop}
		<Desktop3D onExit={() => desktopSettings.set3DDesktop(false)} />
	{:else}
	<div class="desktop-environment" style={backgroundStyle()}>
		<!-- Animated Background Effect -->
		{#if $desktopSettings.animatedBackground.effect !== 'none'}
			<AnimatedBackground
				effectType={$desktopSettings.animatedBackground.effect}
				intensity={$desktopSettings.animatedBackground.intensity}
				colors={$desktopSettings.animatedBackground.colors}
				speed={$desktopSettings.animatedBackground.speed}
			/>
		{/if}

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
								customIcon={icon.customIcon}
								onSelect={handleIconSelect}
								onOpen={handleIconOpen}
								onDragStart={handleIconDragStart}
								onDragMove={handleIconDragMove}
								onDragEnd={handleIconDragEnd}
								onCustomizeIcon={handleCustomizeIcon}
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
						<WindowContent
							module={win.module}
							windowTitle={win.title}
							deployedApps={$deployedAppsStore.apps}
						/>
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
			<DesktopContextMenu
				x={contextMenuPos.x}
				y={contextMenuPos.y}
				type={contextMenuType}
				icon={contextMenuIcon}
				onClose={closeContextMenu}
				onOpenIcon={openIcon}
				onRenameIcon={startRenameIcon}
				onPinToDock={pinToDock}
				onDeleteFolder={deleteFolder}
				onCreateNewFolder={createNewFolder}
				onArrangeIcons={arrangeIcons}
				onOpenDesktopSettings={openDesktopSettings}
			/>
		{/if}

		<!-- Dock -->
		<Dock />

		<!-- OSA Orb — floating, draggable on window desktop -->
		<OsaOrb />

		<!-- Spotlight Search -->
		<SpotlightSearch open={showSpotlight} onClose={() => showSpotlight = false} />

		<!-- Icon Picker Modal -->
		{#if showIconPicker}
			<div class="icon-picker-overlay">
				<button class="icon-picker-backdrop" onclick={handleIconPickerClose}></button>
				<IconPicker
					currentIcon={customizeIconCurrentConfig}
					onSelect={handleIconPickerSelect}
					onClose={handleIconPickerClose}
				/>
			</div>
		{/if}

		<!-- Onboarding Overlay -->
		{#if showOnboarding}
			<DesktopOnboarding
				step={onboardingStep}
				onNext={nextOnboardingStep}
				onSkip={skipOnboarding}
				onComplete={completeOnboarding}
			/>
		{/if}
	</div>
	{/if}
{/if}

<style>
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

	/* Icon Picker Modal */
	.icon-picker-overlay {
		position: fixed;
		inset: 0;
		z-index: 10000;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.icon-picker-backdrop {
		position: absolute;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(4px);
		-webkit-backdrop-filter: blur(4px);
		border: none;
		cursor: pointer;
	}
</style>
