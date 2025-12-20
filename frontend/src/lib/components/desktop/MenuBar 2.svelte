<script lang="ts">
	import { windowStore, focusedWindow } from '$lib/stores/windowStore';
	import { desktopSettings, desktopBackgrounds, iconStyles } from '$lib/stores/desktopStore';
	import { useSession, signOut } from '$lib/auth-client';
	import { goto } from '$app/navigation';

	const session = useSession();

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
			case 'toggle-noise':
				desktopSettings.toggleNoise();
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
		}
	}

	function handleBackgroundChange(backgroundId: string) {
		desktopSettings.setBackground(backgroundId);
		closeMenus();
	}

	let fileInput: HTMLInputElement | undefined = $state(undefined);

	function handleCustomBackgroundClick() {
		fileInput?.click();
	}

	async function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;

		// Check file type
		if (!file.type.startsWith('image/')) {
			alert('Please select an image file');
			return;
		}

		// Check file size (max 5MB)
		if (file.size > 5 * 1024 * 1024) {
			alert('Image must be less than 5MB');
			return;
		}

		// Upload to backend
		const formData = new FormData();
		formData.append('file', file);

		try {
			const response = await fetch('/api/profile/background', {
				method: 'POST',
				body: formData
			});

			if (response.ok) {
				const data = await response.json();
				desktopSettings.setCustomBackground(data.url);
				closeMenus();
			} else {
				// Fallback: use local data URL for now
				const reader = new FileReader();
				reader.onload = (e) => {
					const dataUrl = e.target?.result as string;
					desktopSettings.setCustomBackground(dataUrl);
					closeMenus();
				};
				reader.readAsDataURL(file);
			}
		} catch {
			// Fallback: use local data URL
			const reader = new FileReader();
			reader.onload = (e) => {
				const dataUrl = e.target?.result as string;
				desktopSettings.setCustomBackground(dataUrl);
				closeMenus();
			};
			reader.readAsDataURL(file);
		}

		// Reset input
		target.value = '';
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
			items: [
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
				{ label: 'Documentation', action: 'docs', disabled: true },
				{ type: 'separator' },
				{ label: 'About Business OS', action: 'about', disabled: true },
			]
		},
	]);

	// Group backgrounds by type
	const solidBackgrounds = desktopBackgrounds.filter(b => b.type === 'solid');
	const gradientBackgrounds = desktopBackgrounds.filter(b => b.type === 'gradient');
	const patternBackgrounds = desktopBackgrounds.filter(b => b.type === 'pattern');
</script>

<div class="menu-bar">
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
				<div class="menu-dropdown desktop-menu">
					<div class="menu-section-title">Desktop Background</div>

					<div class="menu-subsection-title">Solid Colors</div>
					<div class="background-grid">
						{#each solidBackgrounds as bg}
							<button
								class="background-swatch"
								class:selected={$desktopSettings.backgroundId === bg.id}
								style="background: {bg.preview}"
								onclick={() => handleBackgroundChange(bg.id)}
								title={bg.name}
								aria-label={bg.name}
							></button>
						{/each}
					</div>

					<div class="menu-subsection-title">Gradients</div>
					<div class="background-grid">
						{#each gradientBackgrounds as bg}
							<button
								class="background-swatch"
								class:selected={$desktopSettings.backgroundId === bg.id}
								style="background: {bg.preview}"
								onclick={() => handleBackgroundChange(bg.id)}
								title={bg.name}
								aria-label={bg.name}
							></button>
						{/each}
					</div>

					<div class="menu-subsection-title">Patterns</div>
					<div class="background-grid">
						{#each patternBackgrounds as bg}
							<button
								class="background-swatch pattern-swatch"
								class:selected={$desktopSettings.backgroundId === bg.id}
								style="background: {bg.preview}; background-size: 10px 10px;"
								onclick={() => handleBackgroundChange(bg.id)}
								title={bg.name}
								aria-label={bg.name}
							></button>
						{/each}
					</div>

					<div class="menu-subsection-title">Custom</div>
					<div class="custom-background-section">
						<button class="upload-background-btn" onclick={handleCustomBackgroundClick}>
							<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
							</svg>
							<span>Upload Image</span>
						</button>
						{#if $desktopSettings.backgroundId === 'custom' && $desktopSettings.customBackgroundUrl}
							<div
								class="custom-preview"
								style="background-image: url({$desktopSettings.customBackgroundUrl})"
							></div>
						{/if}
					</div>

					<input
						bind:this={fileInput}
						type="file"
						accept="image/*"
						onchange={handleFileSelect}
						style="display: none;"
					/>

					<div class="menu-separator"></div>

					<button class="menu-item" onclick={() => handleMenuAction('toggle-noise')}>
						<span class="menu-item-check">
							{#if $desktopSettings.showNoise}
								<svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
									<polyline points="20 6 9 17 4 12"></polyline>
								</svg>
							{/if}
						</span>
						<span class="menu-item-label">Show Texture Overlay</span>
					</button>

					<div class="menu-separator"></div>

					<div class="menu-section-title">Icon Style</div>
					<div class="icon-style-grid">
						{#each iconStyles as style}
							<button
								class="icon-style-option"
								class:selected={$desktopSettings.iconStyle === style.id}
								onclick={() => desktopSettings.setIconStyle(style.id)}
								title={style.description}
							>
								<span class="icon-style-name">{style.name}</span>
							</button>
						{/each}
					</div>

					<div class="menu-separator"></div>

					<button class="menu-item" onclick={() => handleMenuAction('open-terminal')}>
						<span class="menu-item-check">
							<svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M4 17l6-6-6-6M12 19h8"/>
							</svg>
						</span>
						<span class="menu-item-label">Open Terminal (OSA)</span>
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
							{#if item.type === 'separator'}
								<div class="menu-separator"></div>
							{:else}
								<button
									class="menu-item"
									class:disabled={item.disabled}
									class:checked={item.checked}
									disabled={item.disabled}
									onclick={() => {
										if (item.action?.startsWith('window:')) {
											handleWindowSelect(item.action.replace('window:', ''));
										} else {
											handleMenuAction(item.action);
										}
									}}
								>
									<span class="menu-item-check">
										{#if item.checked}
											<svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
												<polyline points="20 6 9 17 4 12"></polyline>
											</svg>
										{/if}
									</span>
									<span class="menu-item-label">{item.label}</span>
									{#if item.shortcut}
										<span class="menu-item-shortcut">{item.shortcut}</span>
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

	.menu-bar-left {
		display: flex;
		align-items: center;
		gap: 0;
	}

	.menu-bar-right {
		display: flex;
		align-items: center;
		gap: 12px;
		position: relative;
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

	.menu-dropdown.desktop-menu {
		min-width: 280px;
		padding: 8px;
	}

	.menu-section-title {
		font-size: 11px;
		font-weight: 600;
		color: #666;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		padding: 4px 8px 8px;
	}

	.menu-subsection-title {
		font-size: 10px;
		font-weight: 500;
		color: #999;
		padding: 8px 8px 4px;
	}

	.background-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 8px;
		padding: 0 8px 8px;
	}

	.background-swatch {
		width: 100%;
		aspect-ratio: 1;
		border-radius: 6px;
		border: 2px solid transparent;
		cursor: pointer;
		transition: all 0.15s ease;
		position: relative;
	}

	.background-swatch:hover {
		transform: scale(1.05);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
	}

	.background-swatch.selected {
		border-color: #0066FF;
		box-shadow: 0 0 0 2px rgba(0, 102, 255, 0.3);
	}

	.background-swatch::after {
		content: attr(title);
		position: absolute;
		bottom: calc(100% + 6px);
		left: 50%;
		transform: translateX(-50%);
		background: rgba(0, 0, 0, 0.85);
		color: white;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 11px;
		white-space: nowrap;
		opacity: 0;
		pointer-events: none;
		transition: opacity 0.15s ease;
		z-index: 10;
	}

	.background-swatch:hover::after {
		opacity: 1;
	}

	.custom-background-section {
		padding: 0 8px 8px;
		display: flex;
		gap: 8px;
		align-items: center;
	}

	.upload-background-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		background: #f0f0f0;
		border: 1px dashed #ccc;
		border-radius: 6px;
		cursor: pointer;
		font-size: 12px;
		color: #666;
		transition: all 0.15s ease;
		flex: 1;
	}

	.upload-background-btn:hover {
		background: #e8e8e8;
		border-color: #999;
		color: #333;
	}

	.upload-background-btn svg {
		width: 16px;
		height: 16px;
	}

	.custom-preview {
		width: 48px;
		height: 48px;
		border-radius: 6px;
		background-size: cover;
		background-position: center;
		border: 2px solid #0066FF;
		box-shadow: 0 0 0 2px rgba(0, 102, 255, 0.3);
		flex-shrink: 0;
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

	/* Icon Style Selector */
	.icon-style-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 6px;
		padding: 8px 12px;
	}

	.icon-style-option {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 6px 8px;
		border-radius: 6px;
		background: rgba(0, 0, 0, 0.04);
		border: 1px solid transparent;
		cursor: pointer;
		font-size: 11px;
		color: #666;
		transition: all 0.15s ease;
	}

	.icon-style-option:hover {
		background: rgba(0, 0, 0, 0.08);
		color: #333;
	}

	.icon-style-option.selected {
		background: rgba(0, 102, 255, 0.1);
		border-color: #0066FF;
		color: #0066FF;
	}

	.icon-style-name {
		font-weight: 500;
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
</style>
