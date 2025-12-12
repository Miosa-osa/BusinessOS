<script lang="ts">
	import { desktopSettings, type IconStyle } from '$lib/stores/desktopStore';

	interface Props {
		id: string;
		module: string;
		label: string;
		selected?: boolean;
		posX: number;
		posY: number;
		darkBackground?: boolean;
		onSelect?: (id: string, additive: boolean) => void;
		onOpen?: (module: string) => void;
		onDragStart?: (id: string) => void;
		onDragMove?: (id: string, newX: number, newY: number) => void;
		onDragEnd?: (id: string, finalX: number, finalY: number) => void;
	}

	let {
		id,
		module,
		label,
		selected = false,
		posX,
		posY,
		darkBackground = false,
		onSelect,
		onOpen,
		onDragStart,
		onDragMove,
		onDragEnd
	}: Props = $props();

	const iconStyle = $derived($desktopSettings.iconStyle);
	const iconSize = $derived($desktopSettings.iconSize);
	const showIconLabels = $derived($desktopSettings.showIconLabels);

	// Calculate dimensions based on icon size
	const containerWidth = $derived(Math.max(iconSize + 24, 64));
	const imageSize = $derived(iconSize * 0.875); // Icon image is 87.5% of icon size
	const svgSize = $derived(iconSize * 0.4375); // SVG is about 50% of image
	const labelSize = $derived(Math.max(9, Math.min(13, iconSize * 0.17)));

	let clickCount = $state(0);
	let clickTimer: ReturnType<typeof setTimeout> | null = null;
	let isDragging = $state(false);
	let dragStartPos = { x: 0, y: 0 };
	let iconStartPos = { x: 0, y: 0 };
	let hasMoved = $state(false);

	function handleMouseDown(event: MouseEvent) {
		if (event.button !== 0) return; // Only left click
		event.preventDefault();

		dragStartPos = { x: event.clientX, y: event.clientY };
		iconStartPos = { x: posX, y: posY };
		hasMoved = false;

		// Start listening for drag
		document.addEventListener('mousemove', handleMouseMove);
		document.addEventListener('mouseup', handleMouseUp);
	}

	function handleMouseMove(event: MouseEvent) {
		const deltaX = event.clientX - dragStartPos.x;
		const deltaY = event.clientY - dragStartPos.y;

		// Only start dragging if moved more than 5px
		if (!isDragging && (Math.abs(deltaX) > 5 || Math.abs(deltaY) > 5)) {
			isDragging = true;
			hasMoved = true;
			onSelect?.(id, false); // Select when starting drag
			onDragStart?.(id);
		}

		if (isDragging) {
			onDragMove?.(id, iconStartPos.x + deltaX, iconStartPos.y + deltaY);
		}
	}

	function handleMouseUp(event: MouseEvent) {
		document.removeEventListener('mousemove', handleMouseMove);
		document.removeEventListener('mouseup', handleMouseUp);

		if (isDragging) {
			const deltaX = event.clientX - dragStartPos.x;
			const deltaY = event.clientY - dragStartPos.y;
			const finalX = iconStartPos.x + deltaX;
			const finalY = iconStartPos.y + deltaY;
			onDragEnd?.(id, finalX, finalY);
			isDragging = false;
		} else if (!hasMoved) {
			// Handle click
			handleClick(event);
		}
	}

	function handleClick(event: MouseEvent) {
		clickCount++;

		if (clickCount === 1) {
			// Single click - select
			onSelect?.(id, event.metaKey || event.ctrlKey);
			clickTimer = setTimeout(() => {
				clickCount = 0;
			}, 300);
		} else if (clickCount === 2) {
			// Double click - open
			if (clickTimer) clearTimeout(clickTimer);
			clickCount = 0;
			onOpen?.(module);
		}
	}

	// Icon SVG paths for each module
	const iconPaths: Record<string, { path: string; color: string; bgColor: string }> = {
		platform: {
			path: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5',
			color: '#333333',
			bgColor: '#F5F5F5'
		},
		terminal: {
			path: 'M4 17l6-6-6-6M12 19h8',
			color: '#1E1E1E',
			bgColor: '#2D2D2D'
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

	const iconData = $derived(iconPaths[module] || iconPaths.dashboard);
	const isTerminal = module === 'terminal';
	const isPlatform = module === 'platform';

	// HTML5 drag start for dock pinning
	function handleDragStart(event: DragEvent) {
		if (event.dataTransfer) {
			event.dataTransfer.setData('text/plain', module);
			event.dataTransfer.effectAllowed = 'copyMove';
		}
	}
</script>

<div
	class="desktop-icon style-{iconStyle}"
	class:selected
	class:dragging={isDragging}
	class:dark-bg={darkBackground}
	style="width: {containerWidth}px;"
	onmousedown={handleMouseDown}
	ondragstart={handleDragStart}
	draggable="true"
	role="button"
	tabindex="0"
	aria-label={label}
>
	<div
		class="icon-image"
		class:terminal={isTerminal}
		style="
			width: {imageSize}px;
			height: {imageSize}px;
			border-radius: {Math.max(8, imageSize * 0.2)}px;
			background-color: {iconStyle === 'minimal' ? 'transparent' : iconData.bgColor};
			{iconStyle === 'outlined' ? `border: 2px solid ${iconData.color}; background-color: transparent;` : ''}
			{iconStyle === 'neon' ? `color: ${iconData.color};` : ''}
			{iconStyle === 'gradient' ? `--gradient-start: ${iconData.color}; --gradient-end: ${iconData.bgColor};` : ''}
		"
	>
		{#if isTerminal}
			<div class="terminal-icon">
				<span class="terminal-prompt" style="font-size: {svgSize * 0.65}px;">&gt;_</span>
			</div>
		{:else}
			<svg
				class="icon-svg"
				viewBox="0 0 24 24"
				fill="none"
				stroke={iconData.color}
				stroke-width="1.5"
				stroke-linecap="round"
				stroke-linejoin="round"
				style="width: {svgSize}px; height: {svgSize}px;"
			>
				<path d={iconData.path} />
			</svg>
		{/if}
	</div>
	{#if showIconLabels}
		<span class="icon-label" class:selected style="font-size: {labelSize}px; max-width: {containerWidth - 8}px;">{label}</span>
	{/if}
</div>

<style>
	.desktop-icon {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		padding: 8px;
		border: none;
		background: transparent;
		cursor: pointer;
		border-radius: 8px;
		transition: transform 0.15s ease;
		width: 80px;
		user-select: none;
	}

	.desktop-icon:hover:not(.dragging) {
		transform: scale(1.05);
	}

	.desktop-icon.dragging {
		opacity: 0.8;
		cursor: grabbing;
		transition: none;
	}

	.desktop-icon:hover .icon-image {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.desktop-icon.selected .icon-image {
		box-shadow: 0 0 0 2px #0066FF;
	}

	.icon-image {
		width: 56px;
		height: 56px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		transition: box-shadow 0.15s ease;
	}

	.icon-image.terminal {
		background: #1E1E1E !important;
	}

	.terminal-icon {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.terminal-prompt {
		font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Code', 'Courier New', monospace;
		font-size: 18px;
		font-weight: bold;
		color: #00FF00;
		text-shadow: 0 0 10px rgba(0, 255, 0, 0.5);
	}

	.icon-svg {
		width: 28px;
		height: 28px;
	}

	.icon-label {
		font-size: 11px;
		font-weight: 500;
		color: #333;
		text-align: center;
		max-width: 72px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		padding: 2px 6px;
		border-radius: 4px;
		text-shadow: 0 1px 2px rgba(255, 255, 255, 0.8);
	}

	/* Icon Style Variants */

	/* Minimal - no backgrounds, just icons */
	.desktop-icon.style-minimal .icon-image {
		box-shadow: none;
		background: transparent !important;
	}

	.desktop-icon.style-minimal:hover .icon-image {
		box-shadow: none;
		background: rgba(0, 0, 0, 0.05) !important;
	}

	.desktop-icon.style-minimal.selected .icon-image {
		box-shadow: none;
		background: rgba(0, 102, 255, 0.1) !important;
	}

	.desktop-icon.style-minimal .icon-svg {
		width: 36px;
		height: 36px;
	}

	/* Rounded - circular backgrounds */
	.desktop-icon.style-rounded .icon-image {
		border-radius: 50%;
	}

	/* Square - sharp corners */
	.desktop-icon.style-square .icon-image {
		border-radius: 4px;
	}

	/* macOS - squircle style */
	.desktop-icon.style-macos .icon-image {
		border-radius: 22%;
		width: 60px;
		height: 60px;
	}

	.desktop-icon.style-macos .icon-svg {
		width: 32px;
		height: 32px;
	}

	/* Outlined - border outline style */
	.desktop-icon.style-outlined .icon-image {
		box-shadow: none;
	}

	.desktop-icon.style-outlined:hover .icon-image {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.desktop-icon.style-outlined.selected .icon-image {
		box-shadow: 0 0 0 2px #0066FF;
	}

	/* Retro - classic pixelated computer style */
	.desktop-icon.style-retro .icon-image {
		border-radius: 0;
		box-shadow:
			4px 4px 0 rgba(0, 0, 0, 0.3),
			inset -2px -2px 0 rgba(0, 0, 0, 0.2),
			inset 2px 2px 0 rgba(255, 255, 255, 0.3);
		image-rendering: pixelated;
	}

	.desktop-icon.style-retro .icon-label {
		font-family: 'Courier New', monospace;
		text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.3);
	}

	.desktop-icon.style-retro.selected .icon-image {
		box-shadow:
			4px 4px 0 rgba(0, 0, 0, 0.3),
			0 0 0 2px #0066FF;
	}

	/* Win95 - Windows 95 style 3D borders */
	.desktop-icon.style-win95 .icon-image {
		border-radius: 0;
		box-shadow: none;
		border: 2px solid;
		border-color: #DFDFDF #808080 #808080 #DFDFDF;
		background: #C0C0C0 !important;
	}

	.desktop-icon.style-win95:hover .icon-image {
		border-color: #808080 #DFDFDF #DFDFDF #808080;
	}

	.desktop-icon.style-win95.selected .icon-image {
		border-color: #808080 #DFDFDF #DFDFDF #808080;
		background: #000080 !important;
	}

	.desktop-icon.style-win95 .icon-label {
		font-family: 'MS Sans Serif', 'Segoe UI', sans-serif;
		font-size: 11px;
	}

	.desktop-icon.style-win95.selected .icon-label {
		background: #000080;
		color: white;
		text-shadow: none;
	}

	/* Glassmorphism - frosted glass effect */
	.desktop-icon.style-glassmorphism .icon-image {
		background: rgba(255, 255, 255, 0.2) !important;
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		border: 1px solid rgba(255, 255, 255, 0.3);
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
	}

	.desktop-icon.style-glassmorphism:hover .icon-image {
		background: rgba(255, 255, 255, 0.3) !important;
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
	}

	.desktop-icon.style-glassmorphism.selected .icon-image {
		border-color: #0066FF;
		box-shadow: 0 0 0 2px rgba(0, 102, 255, 0.3), 0 8px 32px rgba(0, 0, 0, 0.15);
	}

	/* Neon - glowing neon style */
	.desktop-icon.style-neon .icon-image {
		background: #1a1a2e !important;
		border-radius: 12px;
		box-shadow:
			0 0 10px currentColor,
			0 0 20px currentColor,
			inset 0 0 10px rgba(255, 255, 255, 0.1);
		border: 1px solid currentColor;
	}

	.desktop-icon.style-neon:hover .icon-image {
		box-shadow:
			0 0 15px currentColor,
			0 0 30px currentColor,
			0 0 45px currentColor,
			inset 0 0 10px rgba(255, 255, 255, 0.1);
	}

	.desktop-icon.style-neon .icon-svg {
		filter: drop-shadow(0 0 3px currentColor);
	}

	.desktop-icon.style-neon .icon-label {
		color: #fff;
		text-shadow: 0 0 10px currentColor;
	}

	.desktop-icon.style-neon.selected .icon-image {
		box-shadow:
			0 0 10px #0066FF,
			0 0 20px #0066FF,
			0 0 30px #0066FF,
			inset 0 0 10px rgba(255, 255, 255, 0.1);
		border-color: #0066FF;
	}

	/* Flat - flat design with no shadows */
	.desktop-icon.style-flat .icon-image {
		box-shadow: none;
		border-radius: 8px;
	}

	.desktop-icon.style-flat:hover .icon-image {
		box-shadow: none;
		filter: brightness(0.95);
	}

	.desktop-icon.style-flat.selected .icon-image {
		box-shadow: none;
		outline: 2px solid #0066FF;
		outline-offset: 2px;
	}

	/* Gradient - gradient background style */
	.desktop-icon.style-gradient .icon-image {
		background: linear-gradient(135deg, var(--gradient-start, #667eea) 0%, var(--gradient-end, #764ba2) 100%) !important;
		box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
	}

	.desktop-icon.style-gradient .icon-svg {
		stroke: white !important;
		filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.2));
	}

	.desktop-icon.style-gradient:hover .icon-image {
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.25);
		transform: translateY(-2px);
	}

	.desktop-icon.style-gradient.selected .icon-image {
		box-shadow: 0 0 0 2px #0066FF, 0 4px 15px rgba(0, 0, 0, 0.2);
	}

	.icon-label.selected {
		background: #0066FF;
		color: white;
		text-shadow: none;
	}

	/* macOS Classic - Mac OS 9 platinum style */
	.desktop-icon.style-macos-classic .icon-image {
		border-radius: 4px;
		background: linear-gradient(180deg, #EAEAEA 0%, #D4D4D4 50%, #C4C4C4 100%) !important;
		border: 1px solid;
		border-color: #FFFFFF #888888 #888888 #FFFFFF;
		box-shadow:
			1px 1px 0 #666666,
			inset 1px 1px 0 rgba(255, 255, 255, 0.8);
	}

	.desktop-icon.style-macos-classic:hover .icon-image {
		background: linear-gradient(180deg, #F0F0F0 0%, #E0E0E0 50%, #D0D0D0 100%) !important;
	}

	.desktop-icon.style-macos-classic.selected .icon-image {
		background: linear-gradient(180deg, #3366CC 0%, #2255BB 50%, #1144AA 100%) !important;
		border-color: #1144AA #000033 #000033 #1144AA;
	}

	.desktop-icon.style-macos-classic.selected .icon-svg {
		stroke: white !important;
	}

	.desktop-icon.style-macos-classic .icon-label {
		font-family: 'Chicago', 'Charcoal', 'Geneva', 'Helvetica', sans-serif;
		font-size: 10px;
		font-weight: normal;
		text-shadow: none;
		color: #000;
	}

	.desktop-icon.style-macos-classic.selected .icon-label {
		background: #3366CC;
		color: white;
		text-shadow: none;
	}

	/* Paper - card style with soft shadows */
	.desktop-icon.style-paper .icon-image {
		background: #FFFFFF !important;
		border-radius: 8px;
		box-shadow:
			0 1px 3px rgba(0, 0, 0, 0.08),
			0 4px 12px rgba(0, 0, 0, 0.05);
		border: 1px solid rgba(0, 0, 0, 0.06);
	}

	.desktop-icon.style-paper:hover .icon-image {
		box-shadow:
			0 2px 8px rgba(0, 0, 0, 0.1),
			0 8px 24px rgba(0, 0, 0, 0.08);
		transform: translateY(-2px);
	}

	.desktop-icon.style-paper.selected .icon-image {
		box-shadow:
			0 0 0 2px #0066FF,
			0 2px 8px rgba(0, 0, 0, 0.1);
	}

	.desktop-icon.style-paper .icon-label {
		background: rgba(255, 255, 255, 0.9);
		padding: 3px 8px;
		border-radius: 4px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
	}

	/* Pixel - 8-bit pixel art style */
	.desktop-icon.style-pixel .icon-image {
		border-radius: 0;
		image-rendering: pixelated;
		box-shadow:
			4px 0 0 #000,
			-4px 0 0 #000,
			0 4px 0 #000,
			0 -4px 0 #000;
		border: none;
	}

	.desktop-icon.style-pixel:hover .icon-image {
		box-shadow:
			4px 0 0 #333,
			-4px 0 0 #333,
			0 4px 0 #333,
			0 -4px 0 #333;
		filter: brightness(1.1);
	}

	.desktop-icon.style-pixel.selected .icon-image {
		box-shadow:
			4px 0 0 #0066FF,
			-4px 0 0 #0066FF,
			0 4px 0 #0066FF,
			0 -4px 0 #0066FF;
	}

	.desktop-icon.style-pixel .icon-svg {
		image-rendering: pixelated;
	}

	.desktop-icon.style-pixel .icon-label {
		font-family: 'Press Start 2P', 'Courier New', monospace;
		font-size: 8px;
		letter-spacing: 0.5px;
		text-transform: uppercase;
	}

	/* Dark background mode - light text */
	.desktop-icon.dark-bg .icon-label {
		color: #FFFFFF;
		text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8), 0 0 8px rgba(0, 0, 0, 0.5);
	}

	.desktop-icon.dark-bg.selected .icon-label,
	.desktop-icon.dark-bg .icon-label.selected {
		background: rgba(0, 102, 255, 0.9);
		color: white;
		text-shadow: none;
	}

	/* Dark background specific style overrides */
	.desktop-icon.dark-bg.style-win95.selected .icon-label {
		background: #000080;
		color: white;
	}

	.desktop-icon.dark-bg.style-macos-classic .icon-label {
		color: #FFFFFF;
		text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
	}

	.desktop-icon.dark-bg.style-macos-classic.selected .icon-label {
		background: #3366CC;
		color: white;
		text-shadow: none;
	}

	.desktop-icon.dark-bg.style-paper .icon-label {
		background: rgba(0, 0, 0, 0.6);
		color: white;
		text-shadow: none;
	}

	.desktop-icon.dark-bg.style-retro .icon-label {
		color: #00FF00;
		text-shadow: 0 0 10px rgba(0, 255, 0, 0.5), 1px 1px 0 rgba(0, 0, 0, 0.5);
	}
</style>
