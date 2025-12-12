<script lang="ts">
	import { windowStore } from '$lib/stores/windowStore';
	import { desktopSettings } from '$lib/stores/desktopStore';

	const iconStyle = $derived($desktopSettings.iconStyle);

	interface DockItem {
		id: string;
		module: string;
		label: string;
		isOpen: boolean;
		isMinimized: boolean;
		windowId?: string;
	}

	// Icon data for each module
	const moduleIcons: Record<string, { path: string; color: string; bgColor: string; isTerminal?: boolean }> = {
		platform: {
			path: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5',
			color: '#333333',
			bgColor: '#F5F5F5'
		},
		terminal: {
			path: 'M4 17l6-6-6-6M12 19h8',
			color: '#00FF00',
			bgColor: '#1E1E1E',
			isTerminal: true
		},
		dashboard: {
			path: 'M4 5a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm10 0a1 1 0 011-1h4a1 1 0 011 1v2a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h4a1 1 0 011 1v5a1 1 0 01-1 1h-4a1 1 0 01-1-1v-5zm-10 1a1 1 0 011-1h4a1 1 0 011 1v3a1 1 0 01-1 1H5a1 1 0 01-1-1v-3z',
			color: '#1E88E5',
			bgColor: '#E3F2FD'
		},
		chat: {
			path: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z',
			color: '#43A047',
			bgColor: '#E8F5E9'
		},
		tasks: {
			path: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4',
			color: '#FB8C00',
			bgColor: '#FFF3E0'
		},
		projects: {
			path: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z',
			color: '#8E24AA',
			bgColor: '#F3E5F5'
		},
		team: {
			path: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z',
			color: '#00ACC1',
			bgColor: '#E0F7FA'
		},
		clients: {
			path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4',
			color: '#7B1FA2',
			bgColor: '#F3E5F5'
		},
		contexts: {
			path: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
			color: '#5E35B1',
			bgColor: '#EDE7F6'
		},
		nodes: {
			path: 'M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z',
			color: '#E53935',
			bgColor: '#FFEBEE'
		},
		daily: {
			path: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
			color: '#039BE5',
			bgColor: '#E1F5FE'
		},
		settings: {
			path: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
			color: '#546E7A',
			bgColor: '#ECEFF1'
		},
		trash: {
			path: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16',
			color: '#78909C',
			bgColor: '#ECEFF1'
		}
	};

	const moduleLabels: Record<string, string> = {
		platform: 'Business OS',
		terminal: 'Terminal',
		dashboard: 'Dashboard',
		chat: 'Chat',
		tasks: 'Tasks',
		projects: 'Projects',
		team: 'Team',
		clients: 'Clients',
		contexts: 'Contexts',
		nodes: 'Nodes',
		daily: 'Daily Log',
		settings: 'Settings',
		trash: 'Trash'
	};

	let hoveredIndex = $state<number | null>(null);
	let isDragOver = $state(false);

	// Handle drag over for pinning apps
	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		if (event.dataTransfer) {
			event.dataTransfer.dropEffect = 'copy';
		}
		isDragOver = true;
	}

	function handleDragLeave() {
		isDragOver = false;
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		isDragOver = false;
		const module = event.dataTransfer?.getData('text/plain');
		if (module && module !== 'trash' && module !== 'separator') {
			// Add to pinned items if not already there
			if (!$windowStore.dockPinnedItems.includes(module)) {
				windowStore.addToDock(module);
			}
		}
	}

	// Build dock items from pinned items and open windows
	const dockItems = $derived(() => {
		const items: DockItem[] = [];

		// Add pinned items
		for (const module of $windowStore.dockPinnedItems) {
			const windows = $windowStore.windows.filter(w => w.module === module);
			const isOpen = windows.length > 0;
			const minimizedWindow = windows.find(w => w.minimized);

			items.push({
				id: `pinned-${module}`,
				module,
				label: moduleLabels[module] || module,
				isOpen,
				isMinimized: !!minimizedWindow,
				windowId: minimizedWindow?.id || windows[0]?.id
			});
		}

		// Add separator marker
		items.push({
			id: 'separator',
			module: 'separator',
			label: '',
			isOpen: false,
			isMinimized: false
		});

		// Add open windows that aren't pinned
		const pinnedModules = new Set($windowStore.dockPinnedItems);
		const openWindowModules = [...new Set($windowStore.windows.map(w => w.module))];

		for (const module of openWindowModules) {
			if (!pinnedModules.has(module)) {
				const windows = $windowStore.windows.filter(w => w.module === module);
				const minimizedWindow = windows.find(w => w.minimized);

				items.push({
					id: `open-${module}`,
					module,
					label: moduleLabels[module] || module,
					isOpen: true,
					isMinimized: !!minimizedWindow,
					windowId: minimizedWindow?.id || windows[0]?.id
				});
			}
		}

		// Add trash at the end
		if (!pinnedModules.has('trash')) {
			items.push({
				id: 'separator-2',
				module: 'separator',
				label: '',
				isOpen: false,
				isMinimized: false
			});
			items.push({
				id: 'trash',
				module: 'trash',
				label: 'Trash',
				isOpen: $windowStore.windows.some(w => w.module === 'trash'),
				isMinimized: $windowStore.windows.some(w => w.module === 'trash' && w.minimized),
				windowId: $windowStore.windows.find(w => w.module === 'trash')?.id
			});
		}

		return items;
	});

	function handleItemClick(item: DockItem) {
		if (item.module === 'separator') return;

		if (item.isMinimized && item.windowId) {
			windowStore.restoreWindow(item.windowId);
		} else if (item.isOpen && item.windowId) {
			windowStore.focusWindow(item.windowId);
		} else {
			windowStore.openWindow(item.module);
		}
	}

	function getScale(index: number): number {
		if (hoveredIndex === null) return 1;

		const distance = Math.abs(index - hoveredIndex);
		if (distance === 0) return 1.4;
		if (distance === 1) return 1.2;
		if (distance === 2) return 1.1;
		return 1;
	}

	function getTranslateY(index: number): number {
		if (hoveredIndex === null) return 0;

		const distance = Math.abs(index - hoveredIndex);
		if (distance === 0) return -12;
		if (distance === 1) return -6;
		if (distance === 2) return -2;
		return 0;
	}
</script>

<div class="dock-container">
	<div
		class="dock"
		class:drag-over={isDragOver}
		ondragover={handleDragOver}
		ondragleave={handleDragLeave}
		ondrop={handleDrop}
		role="toolbar"
		aria-label="Application dock"
		tabindex="0"
	>
		{#each dockItems() as item, index (item.id)}
			{#if item.module === 'separator'}
				<div class="dock-separator"></div>
			{:else}
				{@const icon = moduleIcons[item.module] || moduleIcons.dashboard}
				<button
					class="dock-item style-{iconStyle}"
					class:has-indicator={item.isOpen}
					style="
						transform: scale({getScale(index)}) translateY({getTranslateY(index)}px);
					"
					onmouseenter={() => hoveredIndex = index}
					onmouseleave={() => hoveredIndex = null}
					onclick={() => handleItemClick(item)}
					aria-label={item.label}
				>
					<div
						class="dock-icon"
						class:terminal={icon.isTerminal}
						style="
							background-color: {iconStyle === 'minimal' ? 'transparent' : icon.bgColor};
							{iconStyle === 'outlined' ? `border: 2px solid ${icon.color}; background-color: transparent;` : ''}
							{iconStyle === 'neon' ? `color: ${icon.color};` : ''}
							{iconStyle === 'gradient' ? `--gradient-start: ${icon.color}; --gradient-end: ${icon.bgColor};` : ''}
						"
					>
						{#if icon.isTerminal}
							<span class="terminal-prompt">&gt;_</span>
						{:else}
							<svg
								class="dock-icon-svg"
								viewBox="0 0 24 24"
								fill="none"
								stroke={icon.color}
								stroke-width="1.5"
								stroke-linecap="round"
								stroke-linejoin="round"
							>
								<path d={icon.path} />
							</svg>
						{/if}
					</div>

					<!-- Open indicator dot -->
					{#if item.isOpen}
						<div class="dock-indicator"></div>
					{/if}

					<!-- Tooltip -->
					{#if hoveredIndex === index}
						<div class="dock-tooltip">{item.label}</div>
					{/if}
				</button>
			{/if}
		{/each}
	</div>
</div>

<style>
	.dock-container {
		position: fixed;
		bottom: 8px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 9999;
	}

	.dock {
		display: flex;
		align-items: flex-end;
		gap: 4px;
		padding: 6px 10px 8px;
		background: rgba(255, 255, 255, 0.75);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.6);
		border-radius: 16px;
		box-shadow:
			0 0 0 0.5px rgba(0, 0, 0, 0.1),
			0 8px 32px rgba(0, 0, 0, 0.12);
		transition: all 0.2s ease;
	}

	.dock.drag-over {
		background: rgba(0, 102, 255, 0.15);
		border-color: rgba(0, 102, 255, 0.5);
		box-shadow:
			0 0 0 2px rgba(0, 102, 255, 0.3),
			0 8px 32px rgba(0, 0, 0, 0.15);
	}

	.dock-separator {
		width: 1px;
		height: 40px;
		background: rgba(0, 0, 0, 0.15);
		margin: 0 4px;
		align-self: center;
	}

	.dock-item {
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		transition: transform 0.15s ease-out;
		transform-origin: bottom center;
	}

	.dock-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: box-shadow 0.15s ease;
	}

	.dock-item:hover .dock-icon {
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
	}

	.dock-icon-svg {
		width: 24px;
		height: 24px;
	}

	.dock-icon.terminal {
		background: #1E1E1E !important;
	}

	.terminal-prompt {
		font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Code', 'Courier New', monospace;
		font-size: 16px;
		font-weight: bold;
		color: #00FF00;
		text-shadow: 0 0 8px rgba(0, 255, 0, 0.5);
	}

	.dock-indicator {
		width: 4px;
		height: 4px;
		border-radius: 50%;
		background: #333;
		margin-top: 4px;
	}

	.dock-tooltip {
		position: absolute;
		bottom: 100%;
		left: 50%;
		transform: translateX(-50%);
		margin-bottom: 8px;
		padding: 6px 12px;
		background: rgba(30, 30, 30, 0.9);
		color: white;
		font-size: 12px;
		font-weight: 500;
		border-radius: 6px;
		white-space: nowrap;
		pointer-events: none;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
	}

	.dock-tooltip::after {
		content: '';
		position: absolute;
		top: 100%;
		left: 50%;
		transform: translateX(-50%);
		border: 6px solid transparent;
		border-top-color: rgba(30, 30, 30, 0.9);
	}

	/* Icon Style Variants for Dock */

	/* Minimal */
	.dock-item.style-minimal .dock-icon {
		box-shadow: none;
		background: transparent !important;
	}

	.dock-item.style-minimal:hover .dock-icon {
		background: rgba(0, 0, 0, 0.08) !important;
	}

	/* Rounded */
	.dock-item.style-rounded .dock-icon {
		border-radius: 50%;
	}

	/* Square */
	.dock-item.style-square .dock-icon {
		border-radius: 4px;
	}

	/* macOS */
	.dock-item.style-macos .dock-icon {
		border-radius: 22%;
		width: 52px;
		height: 52px;
	}

	.dock-item.style-macos .dock-icon-svg {
		width: 28px;
		height: 28px;
	}

	/* Outlined */
	.dock-item.style-outlined .dock-icon {
		box-shadow: none;
	}

	.dock-item.style-outlined:hover .dock-icon {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	/* Retro */
	.dock-item.style-retro .dock-icon {
		border-radius: 0;
		box-shadow:
			3px 3px 0 rgba(0, 0, 0, 0.3),
			inset -2px -2px 0 rgba(0, 0, 0, 0.2),
			inset 2px 2px 0 rgba(255, 255, 255, 0.3);
	}

	/* Win95 */
	.dock-item.style-win95 .dock-icon {
		border-radius: 0;
		box-shadow: none;
		border: 2px solid;
		border-color: #DFDFDF #808080 #808080 #DFDFDF;
		background: #C0C0C0 !important;
	}

	.dock-item.style-win95:hover .dock-icon {
		border-color: #808080 #DFDFDF #DFDFDF #808080;
	}

	/* Glassmorphism */
	.dock-item.style-glassmorphism .dock-icon {
		background: rgba(255, 255, 255, 0.2) !important;
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		border: 1px solid rgba(255, 255, 255, 0.3);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
	}

	.dock-item.style-glassmorphism:hover .dock-icon {
		background: rgba(255, 255, 255, 0.3) !important;
	}

	/* Neon */
	.dock-item.style-neon .dock-icon {
		background: #1a1a2e !important;
		box-shadow:
			0 0 8px currentColor,
			0 0 16px currentColor,
			inset 0 0 8px rgba(255, 255, 255, 0.1);
		border: 1px solid currentColor;
	}

	.dock-item.style-neon:hover .dock-icon {
		box-shadow:
			0 0 12px currentColor,
			0 0 24px currentColor,
			0 0 36px currentColor,
			inset 0 0 8px rgba(255, 255, 255, 0.1);
	}

	.dock-item.style-neon .dock-icon-svg {
		filter: drop-shadow(0 0 3px currentColor);
	}

	/* Flat */
	.dock-item.style-flat .dock-icon {
		box-shadow: none;
		border-radius: 8px;
	}

	.dock-item.style-flat:hover .dock-icon {
		filter: brightness(0.95);
	}

	/* Gradient */
	.dock-item.style-gradient .dock-icon {
		background: linear-gradient(135deg, var(--gradient-start, #667eea) 0%, var(--gradient-end, #764ba2) 100%) !important;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
	}

	.dock-item.style-gradient .dock-icon-svg {
		stroke: white !important;
		filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.2));
	}

	.dock-item.style-gradient:hover .dock-icon {
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.25);
	}

	/* macOS Classic */
	.dock-item.style-macos-classic .dock-icon {
		border-radius: 4px;
		background: linear-gradient(180deg, #EAEAEA 0%, #D4D4D4 50%, #C4C4C4 100%) !important;
		border: 1px solid;
		border-color: #FFFFFF #888888 #888888 #FFFFFF;
		box-shadow:
			1px 1px 0 #666666,
			inset 1px 1px 0 rgba(255, 255, 255, 0.8);
	}

	.dock-item.style-macos-classic:hover .dock-icon {
		background: linear-gradient(180deg, #F0F0F0 0%, #E0E0E0 50%, #D0D0D0 100%) !important;
	}

	/* Paper */
	.dock-item.style-paper .dock-icon {
		background: #FFFFFF !important;
		border-radius: 8px;
		box-shadow:
			0 1px 3px rgba(0, 0, 0, 0.08),
			0 4px 12px rgba(0, 0, 0, 0.05);
		border: 1px solid rgba(0, 0, 0, 0.06);
	}

	.dock-item.style-paper:hover .dock-icon {
		box-shadow:
			0 2px 8px rgba(0, 0, 0, 0.1),
			0 8px 24px rgba(0, 0, 0, 0.08);
		transform: translateY(-2px);
	}

	/* Pixel */
	.dock-item.style-pixel .dock-icon {
		border-radius: 0;
		image-rendering: pixelated;
		box-shadow:
			3px 0 0 #000,
			-3px 0 0 #000,
			0 3px 0 #000,
			0 -3px 0 #000;
		border: none;
	}

	.dock-item.style-pixel:hover .dock-icon {
		box-shadow:
			3px 0 0 #333,
			-3px 0 0 #333,
			0 3px 0 #333,
			0 -3px 0 #333;
		filter: brightness(1.1);
	}
</style>
