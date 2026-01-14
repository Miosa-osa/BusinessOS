<script lang="ts">
	import { windowStore, focusedWindow } from '$lib/stores/windowStore';
	import { desktopSettings } from '$lib/stores/desktopStore';
	import { desktop3dLayoutStore, isEditMode, activeLayout } from '$lib/stores/desktop3dLayoutStore';
	import { themeStore } from '$lib/stores/themeStore';
	import { useSession, signOut } from '$lib/auth-client';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { isElectron, isMacOS } from '$lib/utils/platform';
	import LayoutManager from '$lib/components/desktop3d/LayoutManager.svelte';

	const session = useSession();

	// State for layout management (3D Desktop)
	let showLayoutManager = $state(false);
	let showSaveLayoutModal = $state(false);
	let layoutNameInput = $state('');

	// Theme state
	const isDarkMode = $derived($themeStore.resolvedTheme === 'dark');
	const themeMode = $derived($themeStore.theme);

	function toggleDarkMode() {
		if ($themeStore.theme === 'dark') {
			themeStore.setTheme('light');
		} else {
			themeStore.setTheme('dark');
		}
	}

	function setThemeMode(mode: 'light' | 'dark' | 'system') {
		themeStore.setTheme(mode);
	}

	// Detect Electron and macOS for traffic light handling
	const inElectron = $derived(browser && isElectron());
	const onMac = $derived(browser && isMacOS());
	const needsTrafficLightSpace = $derived(inElectron && onMac);

	let activeMenu: string | null = $state(null);
	let currentTime = $state(new Date());
	let calendarMonth = $state(new Date());

	// Update clock every second
	$effect(() => {
		const interval = setInterval(() => {
			currentTime = new Date();
		}, 1000);
		return () => clearInterval(interval);
	});

	function formatTime(date: Date): string {
		return date.toLocaleDateString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit',
			hour12: true
		});
	}

	// Calendar helpers
	function getDaysInMonth(date: Date): number {
		return new Date(date.getFullYear(), date.getMonth() + 1, 0).getDate();
	}

	function getFirstDayOfMonth(date: Date): number {
		return new Date(date.getFullYear(), date.getMonth(), 1).getDay();
	}

	function getCalendarDays(date: Date): (number | null)[] {
		const daysInMonth = getDaysInMonth(date);
		const firstDay = getFirstDayOfMonth(date);
		const days: (number | null)[] = [];

		// Add empty slots for days before the first day
		for (let i = 0; i < firstDay; i++) {
			days.push(null);
		}

		// Add the days of the month
		for (let i = 1; i <= daysInMonth; i++) {
			days.push(i);
		}

		return days;
	}

	function prevMonth() {
		calendarMonth = new Date(calendarMonth.getFullYear(), calendarMonth.getMonth() - 1, 1);
	}

	function nextMonth() {
		calendarMonth = new Date(calendarMonth.getFullYear(), calendarMonth.getMonth() + 1, 1);
	}

	function isToday(day: number | null): boolean {
		if (day === null) return false;
		const today = new Date();
		return (
			day === today.getDate() &&
			calendarMonth.getMonth() === today.getMonth() &&
			calendarMonth.getFullYear() === today.getFullYear()
		);
	}

	const calendarDays = $derived(getCalendarDays(calendarMonth));
	const monthYearLabel = $derived(calendarMonth.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }));

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
				if ($focusedWindow) {
					windowStore.openWindow($focusedWindow.module);
				}
				break;
			case 'close-window':
				if ($focusedWindow) {
					windowStore.closeWindow($focusedWindow.id);
				}
				break;
			case 'close-all':
				$windowStore.windows.forEach(w => windowStore.closeWindow(w.id));
				break;
			case 'minimize':
				if ($focusedWindow) {
					windowStore.minimizeWindow($focusedWindow.id);
				}
				break;
			case 'maximize':
				if ($focusedWindow) {
					windowStore.toggleMaximize($focusedWindow.id);
				}
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
				layoutNameInput = '';
				break;
			case 'cancel-edit':
				desktop3dLayoutStore.exitEditMode();
				break;
			case 'manage-layouts':
				showLayoutManager = true;
				break;
			case 'reset-layout':
				if (confirm('Reset to default layout? This will discard unsaved changes.')) {
					desktop3dLayoutStore.resetToDefault();
				}
				break;
		}
	}

	function handleWindowSelect(windowId: string) {
		closeMenus();
		const window = $windowStore.windows.find(w => w.id === windowId);
		if (window?.minimized) {
			windowStore.restoreWindow(windowId);
		} else {
			windowStore.focusWindow(windowId);
		}
	}

	// Click outside handler
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.menu-bar-item') && !target.closest('.menu-dropdown') && !target.closest('.menu-bar-logo') && !target.closest('.menu-bar-avatar') && !target.closest('.menu-bar-clock')) {
			closeMenus();
		}
	}

	$effect(() => {
		if (activeMenu) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});

	// Menu item type definitions - cast items as any to avoid complex union type issues
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
						// 3D Desktop Layout Controls
						{ label: `Current: ${$activeLayout?.name || 'Default'}`, action: '', disabled: true },
						{ type: 'separator' },
						{ label: $isEditMode ? 'Cancel Edit Mode' : 'Edit Layout', action: $isEditMode ? 'cancel-edit' : 'edit-layout' },
						{ label: 'Save Layout As...', action: 'save-layout', disabled: !$isEditMode },
						{ label: 'Manage Layouts', action: 'manage-layouts' },
						{ type: 'separator' },
						{ label: 'Reset Layout', action: 'reset-layout' },
				  ]
				: [
						// Regular Desktop View Controls
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
>
	<!-- Left side: Logo and menus -->
	<div class="menu-bar-left">
		<!-- Logo / Desktop Settings -->
		<div class="menu-bar-item-wrapper">
			<button class="menu-bar-logo" onclick={() => toggleMenu('desktop')}>
				<svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
					<path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="2" fill="none"/>
				</svg>
			</button>

			{#if activeMenu === 'desktop'}
				<div class="menu-dropdown">
					<button class="menu-item" onclick={() => { handleMenuAction('desktop-settings'); closeMenus(); }}>
						<span class="menu-item-check">
							<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<circle cx="12" cy="12" r="3"/>
								<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
							</svg>
						</span>
						<span class="menu-item-label">Desktop Settings...</span>
					</button>
					<div class="menu-separator"></div>
					<button class="menu-item experimental-3d" onclick={() => { desktopSettings.toggle3DDesktop(); closeMenus(); }}>
						<span class="menu-item-check">
							{#if $desktopSettings.enable3DDesktop}
								<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polyline points="20 6 9 17 4 12"/>
								</svg>
							{:else}
								<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
									<polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
									<line x1="12" y1="22.08" x2="12" y2="12"/>
								</svg>
							{/if}
						</span>
						<span class="menu-item-label">
							{#if $desktopSettings.enable3DDesktop}
								Exit 3D Desktop
							{:else}
								Experimental 3D Desktop
							{/if}
						</span>
						<span class="menu-item-shortcut beta-badge">Beta</span>
					</button>
					<div class="menu-separator"></div>
					<button class="menu-item" onclick={() => handleMenuAction('open-terminal')}>
						<span class="menu-item-check">
							<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M4 17l6-6-6-6M12 19h8"/>
							</svg>
						</span>
						<span class="menu-item-label">Open Terminal</span>
					</button>
					<div class="menu-separator"></div>
					<button class="menu-item" onclick={() => handleMenuAction('exit-desktop')}>
						<span class="menu-item-check"></span>
						<span class="menu-item-label">Exit Desktop View</span>
					</button>
				</div>
			{/if}
		</div>

		<!-- App name (focused window) -->
		<span class="menu-bar-app-name">
			{$focusedWindow?.title || 'Business OS'}
		</span>

		<!-- Menus -->
		{#each menus as menu}
			<div class="menu-bar-item-wrapper">
				<button
					class="menu-bar-item"
					class:active={activeMenu === menu.id}
					onclick={() => toggleMenu(menu.id)}
					onmouseenter={() => activeMenu && activeMenu !== 'desktop' && activeMenu !== 'user' && (activeMenu = menu.id)}
				>
					{menu.label}
				</button>

				{#if activeMenu === menu.id}
					<div class="menu-dropdown">
						{#each menu.items as item}
							{@const menuItem = item as { type?: string; label?: string; shortcut?: string; action?: string; disabled?: boolean; checked?: boolean }}
							{#if menuItem.type === 'separator'}
								<div class="menu-separator"></div>
							{:else}
								<button
									class="menu-item"
									class:disabled={menuItem.disabled}
									class:checked={menuItem.checked}
									disabled={menuItem.disabled}
									onclick={() => {
										if (menuItem.action?.startsWith('window:')) {
											handleWindowSelect(menuItem.action.replace('window:', ''));
										} else if (menuItem.action) {
											handleMenuAction(menuItem.action);
										}
									}}
								>
									<span class="menu-item-check">
										{#if menuItem.checked}
											<svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
												<polyline points="20 6 9 17 4 12"></polyline>
											</svg>
										{/if}
									</span>
									<span class="menu-item-label">{menuItem.label}</span>
									{#if menuItem.shortcut}
										<span class="menu-item-shortcut">{menuItem.shortcut}</span>
									{/if}
								</button>
							{/if}
						{/each}
					</div>
				{/if}
			</div>
		{/each}
	</div>

	<!-- Right side: Status items -->
	<div class="menu-bar-right">
		<!-- Clock with calendar dropdown -->
		<div class="menu-bar-item-wrapper">
			<button class="menu-bar-clock" onclick={() => toggleMenu('calendar')}>
				{formatTime(currentTime)}
			</button>

			{#if activeMenu === 'calendar'}
				<div class="menu-dropdown calendar-menu">
					<div class="calendar-header">
						<button class="calendar-nav" onclick={prevMonth}>
							<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M15 18l-6-6 6-6" />
							</svg>
						</button>
						<span class="calendar-month-year">{monthYearLabel}</span>
						<button class="calendar-nav" onclick={nextMonth}>
							<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 18l6-6-6-6" />
							</svg>
						</button>
					</div>
					<div class="calendar-weekdays">
						<span>Su</span><span>Mo</span><span>Tu</span><span>We</span><span>Th</span><span>Fr</span><span>Sa</span>
					</div>
					<div class="calendar-days">
						{#each calendarDays as day}
							<span class="calendar-day" class:empty={day === null} class:today={isToday(day)}>
								{day ?? ''}
							</span>
						{/each}
					</div>
					<div class="calendar-today-btn-wrapper">
						<button class="calendar-today-btn" onclick={() => { calendarMonth = new Date(); }}>
							Today
						</button>
					</div>
				</div>
			{/if}
		</div>

		<!-- User avatar -->
		<button class="menu-bar-avatar" onclick={() => toggleMenu('user')}>
			<span class="avatar-initials">
				{$session.data?.user?.name?.charAt(0).toUpperCase() || 'U'}
			</span>
		</button>

		{#if activeMenu === 'user'}
			<div class="menu-dropdown user-menu">
				<div class="menu-user-info">
					<span class="menu-user-name">{$session.data?.user?.name || 'User'}</span>
					<span class="menu-user-email">{$session.data?.user?.email || ''}</span>
				</div>
				<div class="menu-separator"></div>

				<!-- Appearance Section -->
				<div class="menu-section-label">Appearance</div>
				<div class="theme-toggle-group">
					<button
						class="theme-toggle-btn"
						class:active={themeMode === 'light'}
						onclick={() => setThemeMode('light')}
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="5"/>
							<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
						</svg>
						<span>Light</span>
					</button>
					<button
						class="theme-toggle-btn"
						class:active={themeMode === 'dark'}
						onclick={() => setThemeMode('dark')}
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
						</svg>
						<span>Dark</span>
					</button>
					<button
						class="theme-toggle-btn"
						class:active={themeMode === 'system'}
						onclick={() => setThemeMode('system')}
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="2" y="3" width="20" height="14" rx="2"/>
							<path d="M8 21h8M12 17v4"/>
						</svg>
						<span>Auto</span>
					</button>
				</div>

				<div class="menu-separator"></div>
				<button class="menu-item" onclick={() => { closeMenus(); goto('/profile'); }}>
					<span class="menu-item-check"></span>
					<span class="menu-item-label">Profile</span>
				</button>
				<button class="menu-item" onclick={() => { closeMenus(); goto('/settings'); }}>
					<span class="menu-item-check"></span>
					<span class="menu-item-label">Settings</span>
				</button>
				<div class="menu-separator"></div>
				<button class="menu-item" onclick={() => handleMenuAction('logout')}>
					<span class="menu-item-check"></span>
					<span class="menu-item-label">Log Out</span>
				</button>
			</div>
		{/if}
	</div>
</div>

<!-- Save Layout Modal (3D Desktop) -->
{#if showSaveLayoutModal}
	<div class="modal-overlay" onclick={() => { showSaveLayoutModal = false; layoutNameInput = ''; }}>
		<div class="modal-content" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h3>Save Layout</h3>
				<button class="close-btn" onclick={() => { showSaveLayoutModal = false; layoutNameInput = ''; }}>
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18" />
						<line x1="6" y1="6" x2="18" y2="18" />
					</svg>
				</button>
			</div>
			<div class="modal-body">
				<p class="modal-description">Enter a name for your custom layout:</p>
				<input
					type="text"
					class="layout-name-input"
					placeholder="e.g., My Workspace, Development Setup"
					bind:value={layoutNameInput}
					maxlength="255"
					onkeydown={(e) => {
						if (e.key === 'Enter' && layoutNameInput.trim()) {
							desktop3dLayoutStore.saveLayout(layoutNameInput.trim());
							desktop3dLayoutStore.exitEditMode();
							showSaveLayoutModal = false;
							layoutNameInput = '';
							closeMenus();
						}
					}}
				/>
			</div>
			<div class="modal-footer">
				<button class="btn btn-secondary" onclick={() => { showSaveLayoutModal = false; layoutNameInput = ''; }}>
					Cancel
				</button>
				<button
					class="btn btn-primary"
					disabled={!layoutNameInput.trim()}
					onclick={() => {
						if (layoutNameInput.trim()) {
							desktop3dLayoutStore.saveLayout(layoutNameInput.trim());
							desktop3dLayoutStore.exitEditMode();
							showSaveLayoutModal = false;
							layoutNameInput = '';
							closeMenus();
						}
					}}
				>
					Save
				</button>
			</div>
		</div>
	</div>
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

	/* Electron on macOS: taller menu bar with traffic light space */
	.menu-bar.traffic-light-space {
		height: 52px;
		padding-left: 80px; /* Space for traffic lights */
		padding-top: 26px; /* Push content below traffic lights */
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

	.menu-bar-logo {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 4px 10px;
		border-radius: 4px;
		background: none;
		border: none;
		cursor: pointer;
		color: #333;
	}

	.menu-bar-logo:hover {
		background: rgba(0, 0, 0, 0.08);
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

	.menu-bar-item-wrapper {
		position: relative;
	}

	.menu-bar-item {
		padding: 4px 10px;
		border-radius: 4px;
		background: none;
		border: none;
		cursor: pointer;
		font-size: 13px;
		font-weight: 500;
		color: #333;
	}

	.menu-bar-item:hover,
	.menu-bar-item.active {
		background: rgba(0, 0, 0, 0.08);
	}

	.menu-dropdown {
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

	.menu-dropdown.user-menu {
		right: 0;
		left: auto;
	}

	.menu-item {
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

	.menu-item:hover:not(.disabled) {
		background: #0066FF;
		color: white;
	}

	.menu-item:hover:not(.disabled) .menu-item-shortcut {
		color: rgba(255, 255, 255, 0.7);
	}

	.menu-item.disabled {
		color: #999;
		cursor: default;
	}

	.menu-item-check {
		width: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.menu-item-label {
		flex: 1;
	}

	.menu-item-shortcut {
		color: #999;
		font-size: 12px;
		margin-left: auto;
	}

	.menu-item-shortcut.beta-badge {
		background: linear-gradient(135deg, #6366f1, #8b5cf6);
		color: white;
		padding: 2px 6px;
		border-radius: 4px;
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
	}

	.menu-item.experimental-3d:hover .menu-item-shortcut.beta-badge {
		background: linear-gradient(135deg, #818cf8, #a78bfa);
	}

	.menu-separator {
		height: 1px;
		background: rgba(0, 0, 0, 0.1);
		margin: 4px 8px;
	}

	.menu-user-info {
		padding: 8px 12px;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.menu-user-name {
		font-weight: 600;
		color: #111;
	}

	.menu-user-email {
		font-size: 11px;
		color: #666;
	}

	.menu-bar-clock {
		color: #333;
		font-size: 13px;
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px 8px;
		border-radius: 4px;
	}

	.menu-bar-clock:hover {
		background: rgba(0, 0, 0, 0.08);
	}

	/* Calendar dropdown */
	.calendar-menu {
		right: 0;
		left: auto;
		min-width: 260px;
		padding: 12px;
	}

	.calendar-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 12px;
	}

	.calendar-nav {
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		border-radius: 4px;
		color: #666;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.calendar-nav:hover {
		background: rgba(0, 0, 0, 0.08);
		color: #333;
	}

	.calendar-nav svg {
		width: 16px;
		height: 16px;
	}

	.calendar-month-year {
		font-weight: 600;
		font-size: 14px;
		color: #333;
	}

	.calendar-weekdays {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 2px;
		margin-bottom: 4px;
	}

	.calendar-weekdays span {
		text-align: center;
		font-size: 10px;
		font-weight: 600;
		color: #999;
		padding: 4px;
	}

	.calendar-days {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 2px;
	}

	.calendar-day {
		text-align: center;
		font-size: 12px;
		padding: 6px;
		border-radius: 4px;
		cursor: pointer;
		color: #333;
	}

	.calendar-day:hover:not(.empty) {
		background: rgba(0, 0, 0, 0.08);
	}

	.calendar-day.empty {
		cursor: default;
	}

	.calendar-day.today {
		background: #0066FF;
		color: white;
		font-weight: 600;
	}

	.calendar-day.today:hover {
		background: #0055DD;
	}

	.calendar-today-btn-wrapper {
		margin-top: 12px;
		padding-top: 8px;
		border-top: 1px solid rgba(0, 0, 0, 0.1);
		display: flex;
		justify-content: center;
	}

	.calendar-today-btn {
		background: none;
		border: 1px solid rgba(0, 0, 0, 0.15);
		padding: 4px 16px;
		border-radius: 4px;
		font-size: 12px;
		cursor: pointer;
		color: #0066FF;
		font-weight: 500;
	}

	.calendar-today-btn:hover {
		background: rgba(0, 102, 255, 0.08);
	}

	.menu-bar-avatar {
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: #333;
		border: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.avatar-initials {
		color: white;
		font-size: 10px;
		font-weight: 600;
	}

	/* Theme toggle group */
	.menu-section-label {
		padding: 4px 12px 2px;
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: #999;
	}

	.theme-toggle-group {
		display: flex;
		gap: 4px;
		padding: 6px 8px;
	}

	.theme-toggle-btn {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		padding: 8px 4px;
		border: 1px solid rgba(0, 0, 0, 0.1);
		background: transparent;
		border-radius: 8px;
		cursor: pointer;
		font-size: 10px;
		color: #666;
		transition: all 0.15s;
	}

	.theme-toggle-btn svg {
		width: 16px;
		height: 16px;
	}

	.theme-toggle-btn:hover {
		background: rgba(0, 0, 0, 0.05);
		border-color: rgba(0, 0, 0, 0.15);
	}

	.theme-toggle-btn.active {
		background: #0066FF;
		border-color: #0066FF;
		color: white;
	}

	/* ===== DARK MODE - MODERN APPLE STYLE ===== */
	:global(.dark) .menu-bar {
		background: rgba(28, 28, 30, 0.85);
		border-bottom-color: rgba(255, 255, 255, 0.12);
		color: #f5f5f7;
	}

	:global(.dark) .menu-bar-logo {
		color: #f5f5f7;
	}

	:global(.dark) .menu-bar-logo:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .menu-bar-app-name {
		color: #f5f5f7;
	}

	:global(.dark) .menu-bar-item {
		color: #f5f5f7;
	}

	:global(.dark) .menu-bar-item:hover,
	:global(.dark) .menu-bar-item.active {
		background: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .menu-dropdown {
		background: rgba(44, 44, 46, 0.98);
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .menu-item {
		color: #f5f5f7;
	}

	:global(.dark) .menu-item:hover:not(.disabled) {
		background: #0A84FF;
	}

	:global(.dark) .menu-item.disabled {
		color: #6e6e73;
	}

	:global(.dark) .menu-item-shortcut {
		color: #6e6e73;
	}

	:global(.dark) .menu-separator {
		background: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .menu-user-name {
		color: #f5f5f7;
	}

	:global(.dark) .menu-user-email {
		color: #a1a1a6;
	}

	:global(.dark) .menu-bar-clock {
		color: #f5f5f7;
	}

	:global(.dark) .menu-bar-clock:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .calendar-nav {
		color: #a1a1a6;
	}

	:global(.dark) .calendar-nav:hover {
		background: rgba(255, 255, 255, 0.1);
		color: #f5f5f7;
	}

	:global(.dark) .calendar-month-year {
		color: #f5f5f7;
	}

	:global(.dark) .calendar-weekdays span {
		color: #6e6e73;
	}

	:global(.dark) .calendar-day {
		color: #f5f5f7;
	}

	:global(.dark) .calendar-day:hover:not(.empty) {
		background: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .calendar-today-btn-wrapper {
		border-top-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .calendar-today-btn {
		border-color: rgba(255, 255, 255, 0.15);
		color: #0A84FF;
	}

	:global(.dark) .calendar-today-btn:hover {
		background: rgba(10, 132, 255, 0.15);
	}

	:global(.dark) .menu-bar-avatar {
		background: #48484a;
	}

	:global(.dark) .menu-section-label {
		color: #6e6e73;
	}

	:global(.dark) .theme-toggle-btn {
		border-color: rgba(255, 255, 255, 0.12);
		color: #a1a1a6;
	}

	:global(.dark) .theme-toggle-btn:hover {
		background: rgba(255, 255, 255, 0.08);
		border-color: rgba(255, 255, 255, 0.2);
	}

	:global(.dark) .theme-toggle-btn.active {
		background: #0A84FF;
		border-color: #0A84FF;
		color: white;
	}

	/* Modal Styles (for 3D Desktop Layout Save) */
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
		z-index: 20000;
		animation: fadeIn 0.2s ease-out;
	}

	@keyframes fadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}

	.modal-content {
		background: white;
		border-radius: 12px;
		width: 90%;
		max-width: 400px;
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
		padding: 20px;
		border-bottom: 1px solid #e2e8f0;
	}

	.modal-header h3 {
		margin: 0;
		font-size: 18px;
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
		display: flex;
		align-items: center;
	}

	.close-btn:hover {
		color: #1e293b;
	}

	.modal-body {
		padding: 20px;
	}

	.modal-description {
		margin: 0 0 12px 0;
		color: #64748b;
		font-size: 14px;
	}

	.layout-name-input {
		width: 100%;
		padding: 10px 14px;
		border: 2px solid #e2e8f0;
		border-radius: 6px;
		font-size: 14px;
		transition: all 0.2s;
	}

	.layout-name-input:focus {
		outline: none;
		border-color: #0066FF;
		box-shadow: 0 0 0 3px rgba(0, 102, 255, 0.1);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
		padding: 20px;
		border-top: 1px solid #e2e8f0;
	}

	.btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		border: none;
		border-radius: 6px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-primary {
		background: #0066FF;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #0055DD;
	}

	.btn-secondary {
		background: white;
		color: #64748b;
		border: 1px solid #e2e8f0;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #f8fafc;
		border-color: #cbd5e1;
	}

	/* Dark mode for modal */
	:global(.dark) .modal-content {
		background: #1e293b;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .modal-header h3 {
		color: #f5f5f7;
	}

	:global(.dark) .modal-header {
		border-bottom-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .close-btn {
		color: #a1a1a6;
	}

	:global(.dark) .close-btn:hover {
		color: #f5f5f7;
	}

	:global(.dark) .modal-description {
		color: #a1a1a6;
	}

	:global(.dark) .layout-name-input {
		background: #0f172a;
		border-color: rgba(255, 255, 255, 0.1);
		color: #f5f5f7;
	}

	:global(.dark) .layout-name-input:focus {
		border-color: #0A84FF;
		box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.2);
	}

	:global(.dark) .modal-footer {
		border-top-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .btn-primary {
		background: #0A84FF;
	}

	:global(.dark) .btn-primary:hover:not(:disabled) {
		background: #0077FF;
	}

	:global(.dark) .btn-secondary {
		background: #334155;
		border-color: rgba(255, 255, 255, 0.1);
		color: #f5f5f7;
	}

	:global(.dark) .btn-secondary:hover:not(:disabled) {
		background: #475569;
		border-color: rgba(255, 255, 255, 0.15);
	}
</style>
