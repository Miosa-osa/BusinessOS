<script lang="ts">
	import { windowStore, focusedWindow } from '$lib/stores/windowStore';
	import { desktopSettings } from '$lib/stores/desktopStore';
	import { desktop3dLayoutStore, isEditMode, activeLayout } from '$lib/stores/desktop3dLayoutStore';
	import { signOut } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { isElectron, isMacOS } from '$lib/utils/platform';
	import LayoutManager from '$lib/components/desktop3d/LayoutManager.svelte';
	import MenuBarDesktopMenu from '$lib/components/desktop/menubar/MenuBarDesktopMenu.svelte';
	import MenuBarMenus from '$lib/components/desktop/menubar/MenuBarMenus.svelte';
	import MenuBarClock from '$lib/components/desktop/menubar/MenuBarClock.svelte';
	import MenuBarUserMenu from '$lib/components/desktop/menubar/MenuBarUserMenu.svelte';
	import SaveLayoutModal from '$lib/components/desktop/menubar/SaveLayoutModal.svelte';

	// Electron / macOS detection
	const inElectron = $derived(browser && isElectron());
	const onMac = $derived(browser && isMacOS());
	const needsTrafficLightSpace = $derived(inElectron && onMac);

	// Menu open/close state
	let activeMenu: string | null = $state(null);

	// 3D Desktop layout management
	let showLayoutManager = $state(false);
	let showSaveLayoutModal = $state(false);

	function toggleMenu(menu: string) {
		activeMenu = activeMenu === menu ? null : menu;
	}

	function closeMenus() {
		activeMenu = null;
	}

	function handleMenuAction(action: string) {
		closeMenus();

		switch (action) {
			case 'new-window':
				if ($focusedWindow) windowStore.openWindow($focusedWindow.module);
				break;
			case 'close-window':
				if ($focusedWindow) windowStore.closeWindow($focusedWindow.id);
				break;
			case 'close-all':
				$windowStore.windows.forEach(w => windowStore.closeWindow(w.id));
				break;
			case 'minimize':
				if ($focusedWindow) windowStore.minimizeWindow($focusedWindow.id);
				break;
			case 'maximize':
				if ($focusedWindow) windowStore.toggleMaximize($focusedWindow.id);
				break;
			case 'desktop-settings':
				windowStore.openWindow('desktop-settings');
				break;
			case 'exit-desktop':
				goto('/dashboard');
				break;
			case 'logout':
				signOut();
				break;
			case 'open-terminal':
				windowStore.openWindow('terminal');
				break;
			case 'open-docs':
				goto('/docs');
				break;
			case 'edit-layout':
				desktop3dLayoutStore.enterEditMode();
				break;
			case 'save-layout':
				showSaveLayoutModal = true;
				break;
			case 'cancel-edit':
				desktop3dLayoutStore.exitEditMode();
				break;
			case 'manage-layouts':
				showLayoutManager = true;
				break;
			case 'reset-layout':
				if (confirm('Reset to default layout? This will discard unsaved changes.')) {
					desktop3dLayoutStore.loadLayout('default');
					desktop3dLayoutStore.exitEditMode();
				}
				break;
		}
	}

	function handleWindowSelect(windowId: string) {
		closeMenus();
		const win = $windowStore.windows.find(w => w.id === windowId);
		if (win?.minimized) {
			windowStore.restoreWindow(windowId);
		} else {
			windowStore.focusWindow(windowId);
		}
	}

	// Click outside — close all menus
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (
			!target.closest('.menu-bar-item') &&
			!target.closest('.menu-dropdown') &&
			!target.closest('.menu-bar-logo') &&
			!target.closest('.menu-bar-avatar') &&
			!target.closest('.menu-bar-clock')
		) {
			closeMenus();
		}
	}

	$effect(() => {
		if (activeMenu) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});

	// Derived menu definitions
	const menus = $derived([
		{
			id: 'file',
			label: 'File',
			items: [
				{ label: 'New Window', shortcut: 'Cmd+N', action: 'new-window', disabled: !$focusedWindow },
				{ type: 'separator' },
				{ label: 'Close Window', shortcut: 'Cmd+W', action: 'close-window', disabled: !$focusedWindow },
				{ label: 'Close All Windows', action: 'close-all', disabled: $windowStore.windows.length === 0 },
				{ type: 'separator' },
				{ label: 'Exit Desktop View', action: 'exit-desktop' },
			]
		},
		{
			id: 'edit',
			label: 'Edit',
			items: [
				{ label: 'Undo', shortcut: 'Cmd+Z', action: 'undo', disabled: true },
				{ label: 'Redo', shortcut: 'Cmd+Shift+Z', action: 'redo', disabled: true },
				{ type: 'separator' },
				{ label: 'Cut', shortcut: 'Cmd+X', action: 'cut', disabled: true },
				{ label: 'Copy', shortcut: 'Cmd+C', action: 'copy', disabled: true },
				{ label: 'Paste', shortcut: 'Cmd+V', action: 'paste', disabled: true },
				{ label: 'Select All', shortcut: 'Cmd+A', action: 'select-all', disabled: true },
			]
		},
		{
			id: 'view',
			label: 'View',
			items: $desktopSettings.enable3DDesktop
				? [
						{ label: `Current: ${$activeLayout?.name || 'Default'}`, action: '', disabled: true },
						{ type: 'separator' },
						{ label: $isEditMode ? 'Cancel Edit Mode' : 'Edit Layout', action: $isEditMode ? 'cancel-edit' : 'edit-layout' },
						{ label: 'Save Layout As...', action: 'save-layout', disabled: !$isEditMode },
						{ label: 'Manage Layouts', action: 'manage-layouts' },
						{ type: 'separator' },
						{ label: 'Reset Layout', action: 'reset-layout' },
				  ]
				: [
						{ label: 'Zoom In', shortcut: 'Cmd++', action: 'zoom-in', disabled: true },
						{ label: 'Zoom Out', shortcut: 'Cmd+-', action: 'zoom-out', disabled: true },
						{ label: 'Actual Size', shortcut: 'Cmd+0', action: 'zoom-reset', disabled: true },
						{ type: 'separator' },
						{ label: 'Arrange Windows', action: 'arrange', disabled: true },
						{ label: 'Tile Windows', action: 'tile', disabled: true },
				  ]
		},
		{
			id: 'window',
			label: 'Window',
			items: [
				{ label: 'Minimize', shortcut: 'Cmd+M', action: 'minimize', disabled: !$focusedWindow },
				{ label: $focusedWindow?.maximized ? 'Restore' : 'Maximize', action: 'maximize', disabled: !$focusedWindow },
				{ type: 'separator' },
				...$windowStore.windows.map(w => ({
					label: w.title + (w.minimized ? ' (minimized)' : ''),
					action: `window:${w.id}`,
					checked: w.id === $focusedWindow?.id
				})),
				...($windowStore.windows.length > 0 ? [{ type: 'separator' }] : []),
				{ label: 'Bring All to Front', action: 'bring-all-front', disabled: true },
			]
		},
		{
			id: 'help',
			label: 'Help',
			items: [
				{ label: 'Keyboard Shortcuts', action: 'shortcuts', disabled: true },
				{ label: 'Documentation', action: 'open-docs' },
				{ type: 'separator' },
				{ label: 'About Business OS', action: 'about', disabled: true },
			]
		},
	]);
</script>

<div
	class="menu-bar"
	class:electron={inElectron}
	class:traffic-light-space={needsTrafficLightSpace}
	style={inElectron ? '-webkit-app-region: drag;' : ''}
	role="menubar"
>
	<!-- Left: logo, app name, menus -->
	<div class="menu-bar-left">
		<MenuBarDesktopMenu
			isOpen={activeMenu === 'desktop'}
			onToggle={() => toggleMenu('desktop')}
			onAction={handleMenuAction}
		/>

		<span class="menu-bar-app-name">
			{$focusedWindow?.title || 'Business OS'}
		</span>

		<MenuBarMenus
			{menus}
			{activeMenu}
			onToggle={toggleMenu}
			onAction={handleMenuAction}
			onWindowSelect={handleWindowSelect}
		/>
	</div>

	<!-- Right: clock, user avatar -->
	<div class="menu-bar-right">
		<MenuBarClock
			isOpen={activeMenu === 'calendar'}
			onToggle={() => toggleMenu('calendar')}
		/>

		<MenuBarUserMenu
			isOpen={activeMenu === 'user'}
			onToggle={() => toggleMenu('user')}
			onClose={closeMenus}
			onAction={handleMenuAction}
		/>
	</div>
</div>

<!-- Save Layout Modal (3D Desktop) -->
{#if showSaveLayoutModal}
	<SaveLayoutModal
		onClose={() => { showSaveLayoutModal = false; }}
		onSaved={() => { showSaveLayoutModal = false; closeMenus(); }}
	/>
{/if}

<!-- Layout Manager Modal (3D Desktop) -->
<LayoutManager show={showLayoutManager} onClose={() => showLayoutManager = false} />

<style>
	.menu-bar {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 26px;
		background: rgba(255, 255, 255, 0.85);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-bottom: 1px solid rgba(0, 0, 0, 0.1);
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 8px;
		z-index: 10000;
		font-size: 13px;
		font-weight: 500;
		color: #333;
		user-select: none;
	}

	.menu-bar.traffic-light-space {
		height: 52px;
		padding-left: 80px;
		padding-top: 26px;
		padding-bottom: 8px;
		align-items: center;
		box-sizing: border-box;
	}

	.menu-bar-left {
		display: flex;
		align-items: center;
		gap: 0;
		-webkit-app-region: no-drag;
	}

	.menu-bar-right {
		display: flex;
		align-items: center;
		gap: 12px;
		position: relative;
		-webkit-app-region: no-drag;
	}

	.menu-bar-app-name {
		font-weight: 600;
		padding: 0 12px 0 4px;
		color: #111;
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* Shared dropdown styles used by all child components */
	:global(.menu-dropdown) {
		position: absolute;
		top: 100%;
		left: 0;
		margin-top: 2px;
		min-width: 220px;
		background: rgba(255, 255, 255, 0.98);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(0, 0, 0, 0.15);
		border-radius: 6px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
		padding: 4px 0;
		z-index: 10001;
	}

	:global(.menu-dropdown.user-menu) {
		right: 0;
		left: auto;
	}

	:global(.menu-item) {
		display: flex;
		align-items: center;
		width: 100%;
		padding: 6px 12px;
		background: none;
		border: none;
		cursor: pointer;
		font-size: 13px;
		color: #333;
		text-align: left;
		gap: 8px;
		border-radius: 4px;
		margin: 0 4px;
		width: calc(100% - 8px);
	}

	:global(.menu-item:hover:not(.disabled)) {
		background: #0066FF;
		color: white;
	}

	:global(.menu-item:hover:not(.disabled) .menu-item-shortcut) {
		color: rgba(255, 255, 255, 0.7);
	}

	:global(.menu-item.disabled) {
		color: #999;
		cursor: default;
	}

	:global(.menu-item-check) {
		width: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	:global(.menu-item-label) {
		flex: 1;
	}

	:global(.menu-item-shortcut) {
		color: #999;
		font-size: 12px;
		margin-left: auto;
	}

	:global(.menu-item-shortcut.beta-badge) {
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		color: white;
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
	}

	:global(.menu-item.experimental-3d:hover .menu-item-shortcut.beta-badge) {
		background: linear-gradient(135deg, #818cf8, #a78bfa);
	}

	:global(.menu-separator) {
		height: 1px;
		background: rgba(0, 0, 0, 0.1);
		margin: 4px 8px;
	}

	/* ===== DARK MODE ===== */
	:global(.dark) .menu-bar {
		background: rgba(28, 28, 30, 0.85);
		border-bottom-color: rgba(255, 255, 255, 0.12);
		color: #f5f5f7;
	}

	:global(.dark) .menu-bar-app-name {
		color: #f5f5f7;
	}

	:global(.dark) :global(.menu-dropdown) {
		background: rgba(44, 44, 46, 0.98);
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) :global(.menu-item) {
		color: #f5f5f7;
	}

	:global(.dark) :global(.menu-item:hover:not(.disabled)) {
		background: #0A84FF;
	}

	:global(.dark) :global(.menu-item.disabled) {
		color: #6e6e73;
	}

	:global(.dark) :global(.menu-item-shortcut) {
		color: #6e6e73;
	}

	:global(.dark) :global(.menu-separator) {
		background: rgba(255, 255, 255, 0.1);
	}
</style>
