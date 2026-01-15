<script lang="ts">
	import { T } from '@threlte/core';
	import { HTML } from '@threlte/extras';
	import { spring } from 'svelte/motion';
	import { desktop3dStore, type Window3DState, type ViewMode } from '$lib/stores/desktop3dStore';

	interface Props {
		window: Window3DState;
		windowIndex: number; // Index for staggered loading
		isFocused: boolean;
		isNextWindow: boolean;
		isPrevWindow: boolean;
		isHovered: boolean;
		viewMode: ViewMode;
		onClick?: () => void;
		onResize?: (widthDelta: number, heightDelta: number) => void;
		onHover?: (hovered: boolean) => void;
	}

	let {
		window,
		windowIndex = 0,
		isFocused = false,
		isNextWindow = false,
		isPrevWindow = false,
		isHovered = false,
		viewMode = 'orb',
		onClick,
		onResize,
		onHover
	}: Props = $props();

	// STAGGERED LOADING: Load iframes sequentially to prevent ERR_INSUFFICIENT_RESOURCES
	// Each window loads after a delay based on its index (300ms apart)
	let iframeLoaded = $state(false);

	$effect(() => {
		// Focused window loads immediately
		if (isFocused) {
			iframeLoaded = true;
			return;
		}

		// Other windows load with staggered delay (300ms per window)
		const delay = windowIndex * 300;
		const timer = setTimeout(() => {
			iframeLoaded = true;
		}, delay);

		return () => clearTimeout(timer);
	});

	// Track pointer down position to distinguish click vs drag
	let pointerDownPos: { x: number; y: number } | null = null;
	const DRAG_THRESHOLD = 15; // pixels - increased for better click detection

	// Spring animation for smooth position transitions
	type Vec3 = [number, number, number];
	let animatedPosition = spring<Vec3>(window.position, {
		stiffness: 0.08,
		damping: 0.7
	});
	// Separate springs for Y rotation (facing) and X rotation (tilt)
	// This allows proper rotation order: Y first, then X
	let animatedYRotation = spring(0, {
		stiffness: 0.08,
		damping: 0.7
	});
	let animatedXRotation = spring(0, {
		stiffness: 0.08,
		damping: 0.7
	});
	let animatedScale = spring(1, {
		stiffness: 0.1,
		damping: 0.6
	});

	// Handle resize button clicks - call store DIRECTLY to bypass prop chain issues in 3D HTML
	function handleSizeIncrease(e: MouseEvent | Event) {
		e.stopPropagation();
		if ('stopImmediatePropagation' in e) e.stopImmediatePropagation();
		e.preventDefault();
		desktop3dStore.resizeFocusedWindow(100, 75);
	}
	function handleSizeDecrease(e: MouseEvent | Event) {
		e.stopPropagation();
		if ('stopImmediatePropagation' in e) e.stopImmediatePropagation();
		e.preventDefault();
		desktop3dStore.resizeFocusedWindow(-100, -75);
	}

	// Current opacity (not animated)
	// All windows always visible, just dimmed in focus mode
	let currentOpacity = $derived.by(() => {
		if (isFocused) return 1;
		if (isNextWindow || isPrevWindow) return 0.85; // Side previews clearly visible
		if (viewMode === 'focused') return 0.15; // Others very faded, orb in background
		return 1;
	});

	// Always render all windows
	let shouldRender = true;

	// Track if we're dragging to prevent click on drag end
	let isDragging = $state(false);
	let pointerDownTime = 0;

	// Handle pointer down - track start position
	function handlePointerDown(e: any) {
		isDragging = false;
		pointerDownTime = Date.now();
		// Store the pointer position when pressed
		const event = e.nativeEvent || e;
		if (event.clientX !== undefined) {
			pointerDownPos = { x: event.clientX, y: event.clientY };
		}
	}

	// Handle pointer move - detect if dragging
	function handlePointerMove(e: any) {
		if (!pointerDownPos) return;
		const event = e.nativeEvent || e;
		if (event.clientX !== undefined) {
			const dx = Math.abs(event.clientX - pointerDownPos.x);
			const dy = Math.abs(event.clientY - pointerDownPos.y);
			if (dx > DRAG_THRESHOLD || dy > DRAG_THRESHOLD) {
				isDragging = true;
			}
		}
	}

	// Handle click - only trigger if it wasn't a drag
	function handleClick(e: any) {
		const clickDuration = Date.now() - pointerDownTime;

		// Don't trigger if: was dragging, held too long, or no pointer down recorded
		if (isDragging || clickDuration > 500 || !pointerDownPos) {
			pointerDownPos = null;
			isDragging = false;
			return;
		}

		pointerDownPos = null;
		isDragging = false;
		e.stopPropagation();
		onClick?.();
	}

	// Calculate target position based on state
	function getTargetPosition(): Vec3 {
		if (isFocused) {
			// Focused window: always come to front center, facing camera
			return [0, 10, 200];
		}
		if (isPrevWindow) {
			// Previous window: to the left
			return [-200, 10, 150];
		}
		if (isNextWindow) {
			// Next window: to the right
			return [200, 10, 150];
		}
		// Other windows: stay in orb positions
		return window.position;
	}

	// Calculate Y rotation - face outward from sphere center
	// This is applied FIRST (outer group) so tilt works correctly
	function getTargetYRotation(): number {
		const [x, , z] = window.position;

		// When focused or adjacent, face the camera (no Y rotation)
		if (isFocused || isNextWindow || isPrevWindow) {
			return 0;
		}

		// Face outward from sphere center
		return Math.atan2(x, z);
	}

	// Calculate X rotation (tilt) - applied AFTER Y rotation
	// Top windows tilt UP (face upward), bottom windows tilt DOWN (face downward)
	function getTargetXRotation(): number {
		const [, y] = window.position;

		// When focused or adjacent, no tilt
		if (isFocused || isNextWindow || isPrevWindow) {
			return 0;
		}

		// Normalize Y position based on sphere radius
		const normalizedY = y / 95;

		// Top windows (positive Y) = negative tilt = face DOWNWARD ~45 degrees
		// Bottom windows (negative Y) = positive tilt = face UPWARD ~45 degrees
		// Middle windows (y near 0) = minimal tilt
		// 0.785 radians = 45 degrees
		return -normalizedY * 0.785;
	}

	// Calculate target scale
	function getTargetScale(): number {
		if (isFocused) return 2.2; // Large focused window
		if (isNextWindow || isPrevWindow) return 0.9; // Side previews visible
		if (viewMode === 'focused') return 0.5; // Background windows much smaller
		return 1;
	}

	// Update springs when state changes
	$effect(() => {
		const targetPos = getTargetPosition();
		animatedPosition.set(targetPos);
	});

	$effect(() => {
		const targetY = getTargetYRotation();
		animatedYRotation.set(targetY);
	});

	$effect(() => {
		const targetX = getTargetXRotation();
		animatedXRotation.set(targetX);
	});

	$effect(() => {
		const targetScale = getTargetScale();
		animatedScale.set(targetScale);
	});

	// HUGE scale for visible windows - BIGGER for better visibility
	const htmlScale = 2.0;

	// Route mapping for modules where route differs from module name
	function getModuleRoute(module: string): string {
		// Handle core modules with special routes
		switch (module) {
			case 'pages':
				return '/knowledge-v2';  // pages redirects to knowledge-v2
			case 'contexts':
				return '/knowledge-v2';  // contexts also redirects to knowledge-v2
			case 'communication':
				return '/communication/calendar';  // communication redirects to calendar tab
			default:
				return `/${module}`;
		}
	}

	// Detect if running in Electron (desktop app)
	// In Electron, we can use <webview> tag which bypasses X-Frame-Options restrictions
	// The preload script exposes window.electron API via contextBridge
	// Use globalThis to avoid conflict with the 'window' prop (Window3DState)
	function checkIsElectron(): boolean {
		if (typeof globalThis === 'undefined' || typeof globalThis.window === 'undefined') return false;
		const browserWindow = globalThis.window as any;
		return browserWindow && 'electron' in browserWindow;
	}

	// Check immediately on client-side and make reactive
	// Initialize with client-side check to avoid flash of wrong content
	const browserWindow = typeof globalThis !== 'undefined' ? (globalThis as any).window : undefined;
	let isElectron = $state(browserWindow && 'electron' in browserWindow);

	$effect(() => {
		// Re-check when component mounts (client-side) to ensure we have the latest value
		const checked = checkIsElectron();
		if (checked !== isElectron) {
			isElectron = checked;
		}
	});

	// Get the iframe src for user apps - use proxy for external URLs
	function getUserAppIframeSrc(url: string): string {
		// Native bundle IDs can't be proxied (e.g., com.apple.finder)
		if (url.startsWith('com.') || url.startsWith('org.') || url.startsWith('io.')) {
			return ''; // Will show native app UI instead
		}

		// In Electron, load URLs directly - webview bypasses X-Frame-Options
		if (isElectron) {
			return url;
		}

		// In browser, use proxy endpoint to strip X-Frame-Options/CSP headers
		return `/api/proxy?url=${encodeURIComponent(url)}`;
	}

	// Check if this is a native bundle ID (not a URL)
	function isNativeBundleId(url: string): boolean {
		return url.startsWith('com.') || url.startsWith('org.') || url.startsWith('io.');
	}

	// List of domains known to block iframe embedding
	// These sites use X-Frame-Options: DENY or strict CSP that proxying can't reliably fix
	const KNOWN_BLOCKED_DOMAINS = [
		'claude.ai',
		'chat.openai.com',
		'notion.so',
		'notion.com',
		'linear.app',
		'slack.com',
		'app.slack.com',
		'figma.com',
		'github.com',
		'twitter.com',
		'x.com',
		'facebook.com',
		'instagram.com',
		'linkedin.com',
		'mail.google.com',
		'drive.google.com',
		'docs.google.com',
		'app.asana.com',
		'trello.com',
		'monday.com',
		'clickup.com',
		'airtable.com',
		'hubspot.com',
		'salesforce.com',
		'dropbox.com',
		'box.com'
	];

	// Check if URL is from a known blocked domain
	function isKnownBlockedDomain(url: string): boolean {
		try {
			const parsedUrl = new URL(url);
			const hostname = parsedUrl.hostname.toLowerCase();
			return KNOWN_BLOCKED_DOMAINS.some(domain =>
				hostname === domain || hostname.endsWith('.' + domain)
			);
		} catch {
			return false;
		}
	}

	// Determine if we should show launcher instead of iframe (in browser mode)
	// This gives better UX by not showing a broken/loading iframe
	function shouldShowLauncher(url: string): boolean {
		if (isElectron) return false; // Electron can embed anything via webview
		if (isNativeBundleId(url)) return false; // Handled separately
		return isKnownBlockedDomain(url);
	}

	// Track iframe element for focus management
	let iframeElement = $state<HTMLIFrameElement | null>(null);

	// Open user app in new browser window (popup Chromium window)
	function openInBrowser() {
		if (window.userAppUrl) {
			// Open as a popup window with specific size - like a mini browser
			const width = 1400;
			const height = 900;
			const left = (screen.width - width) / 2;
			const top = (screen.height - height) / 2;
			globalThis.open(
				window.userAppUrl,
				window.title,
				`width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no,location=yes,status=no,resizable=yes,scrollbars=yes`
			);
		}
	}

	// Auto-open popup when user app window becomes focused (for blocked sites in browser mode)
	let hasAutoOpened = $state(false);
	$effect(() => {
		// Only auto-open for user apps with blocked domains, when focused, and only once
		if (
			isFocused &&
			!hasAutoOpened &&
			window.isUserApp &&
			window.userAppUrl &&
			!isElectron &&
			shouldShowLauncher(window.userAppUrl)
		) {
			hasAutoOpened = true;
			// Small delay so user sees the window first
			setTimeout(() => {
				openInBrowser();
			}, 300);
		}
		// Reset when unfocused so it can open again next time
		if (!isFocused) {
			hasAutoOpened = false;
		}
	});

	// Track if proxy failed to load content
	let proxyLoadFailed = $state(false);

	// Handle iframe load error
	function handleIframeError() {
		// Proxy failed - will show error state
		proxyLoadFailed = true;
	}

	// Track previous focus state to only focus on transition
	let wasFocused = $state(false);

	// CRITICAL: Focus iframe content when window BECOMES focused (transition from unfocused to focused)
	// This ensures keyboard input (arrow keys, Enter, etc.) reaches the iframe content
	// BUT only runs ONCE when focus changes, not continuously
	$effect(() => {
		// Only focus when transitioning from unfocused to focused
		if (isFocused && !wasFocused && iframeElement) {
			// Focus iframe when window becomes focused
			// Focus the iframe element itself
			iframeElement.focus();

			// Try to focus content inside iframe (works for same-origin iframes)
			try {
				iframeElement.contentWindow?.focus();
			} catch (e) {
				// Cross-origin iframe - can't access contentWindow
				// Expected for cross-origin iframes - focus on iframe element is sufficient
			}
		}
		// Update previous state
		wasFocused = isFocused;
	});
</script>

<!-- Window using HTML component for DOM content in 3D space -->
<!-- Only render if should be visible -->
{#if shouldRender}
	<!-- Invisible click mesh - ONLY for non-focused, non-adjacent windows in orb mode -->
	{#if !isFocused && !isNextWindow && !isPrevWindow}
		<!-- Click mesh - positioned at window center, faces camera (no rotation) -->
		<T.Mesh
			position={$animatedPosition}
			onpointerdown={(e: { stopPropagation: () => void }) => { e.stopPropagation(); handlePointerDown(e); }}
			onpointermove={(e: unknown) => { handlePointerMove(e); }}
			onclick={(e: { stopPropagation: () => void }) => { e.stopPropagation(); handleClick(e); }}
			onpointerenter={(e: { stopPropagation: () => void }) => { e.stopPropagation(); onHover?.(true); }}
			onpointerleave={(e: { stopPropagation: () => void }) => { e.stopPropagation(); onHover?.(false); }}
		>
			<T.SphereGeometry args={[50, 12, 12]} />
			<T.MeshBasicMaterial visible={false} transparent opacity={0} />
		</T.Mesh>
	{:else if isFocused}
		<!-- Blocker mesh when focused - prevents background clicks -->
		<T.Mesh
			position={$animatedPosition}
			onclick={(e: { stopPropagation: () => void }) => e.stopPropagation()}
		>
			<T.PlaneGeometry args={[window.width * 0.6, window.height * 0.6]} />
			<T.MeshBasicMaterial visible={false} transparent opacity={0} />
		</T.Mesh>
	{/if}
	<!-- NOTE: No click mesh for prev/next windows - they use HTML pointer events -->

	<!-- Nested groups for proper rotation order: Y first (face outward), then X (tilt) -->
	<T.Group position={$animatedPosition}>
	<T.Group rotation={[0, $animatedYRotation, 0]}>
	<T.Group rotation={[$animatedXRotation, 0, 0]}>
	<HTML
		transform
		scale={htmlScale * $animatedScale}
		pointerEvents={isFocused || isNextWindow || isPrevWindow ? 'auto' : 'none'}
		zIndexRange={[100, 0]}
	>
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="window-wrapper"
			class:focused={isFocused}
			class:dimmed={!isFocused && viewMode === 'focused'}
			class:clickable={isNextWindow || isPrevWindow}
			class:hovered={isHovered}
			style="opacity: {currentOpacity};"
			onclick={(isNextWindow || isPrevWindow) ? onClick : undefined}
		>
			<div
				class="window-3d"
				style="--window-width: {window.width}px; --window-height: {window.height}px;"
			>

				<!-- Title bar -->
				<div class="window-titlebar" style="border-left: 4px solid {window.color};">
					<span class="module-title">{window.title}</span>
					{#if window.isUserApp && window.userAppUrl}
						<button
							class="open-external-btn"
							onclick={(e) => { e.stopPropagation(); openInBrowser(); }}
							title="Open in browser"
						>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6M15 3h6v6M10 14L21 3" />
							</svg>
						</button>
					{/if}
					<div class="titlebar-right">
						{#if isFocused}
							<!-- Size controls when focused -->
							<!-- svelte-ignore a11y_no_static_element_interactions -->
							<!-- svelte-ignore a11y_click_events_have_key_events -->
							<div
								class="size-controls"
								onmousedown={(e) => e.stopPropagation()}
								onclick={(e) => e.stopPropagation()}
							>
								<button
									type="button"
									class="size-btn decrease-btn"
									onclick={(e) => { e.stopPropagation(); e.preventDefault(); handleSizeDecrease(e); }}
									onpointerdown={(e) => { e.stopPropagation(); e.preventDefault(); }}
									onmousedown={(e) => { e.stopPropagation(); e.preventDefault(); }}
									title="Click to make smaller"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
										<path d="M5 12h14" />
									</svg>
								</button>
								<span class="size-label">{window.width}x{window.height}</span>
								<button
									type="button"
									class="size-btn increase-btn"
									onclick={(e) => { e.stopPropagation(); e.preventDefault(); handleSizeIncrease(e); }}
									onpointerdown={(e) => { e.stopPropagation(); e.preventDefault(); }}
									onmousedown={(e) => { e.stopPropagation(); e.preventDefault(); }}
									title="Click to make larger"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
										<path d="M12 5v14M5 12h14" />
									</svg>
								</button>
							</div>
						{/if}
						<div class="titlebar-controls">
							<span class="control-dot" style="background: #febc2e;"></span>
							<span class="control-dot" style="background: #28c840;"></span>
							<span class="control-dot" style="background: #ff5f57;"></span>
						</div>
					</div>
				</div>

				<!-- LIVE Content - STAGGERED LOADING to prevent ERR_INSUFFICIENT_RESOURCES -->
				<!-- Windows load sequentially (300ms apart) to prevent browser resource exhaustion -->
				<div class="window-content">
					{#if iframeLoaded}
						<!-- Window content loads after staggered delay -->
						{#if window.isUserApp && window.userAppUrl && isNativeBundleId(window.userAppUrl)}
							<!-- Native macOS App - Show launch button -->
							<div class="native-app-container">
								<div class="native-app-icon">
									{#if window.userAppLogoUrl}
										<img src={window.userAppLogoUrl} alt={window.title} />
									{:else}
										<div class="icon-placeholder" style="background-color: {window.color};">
											{window.title.charAt(0)}
										</div>
									{/if}
								</div>
								<h3>{window.title}</h3>
								<p class="native-app-info">This is a native macOS application</p>
								<button
									class="launch-button"
									onclick={() => {
										fetch(`/api/system/launch-app?bundle_id=${encodeURIComponent(window.userAppUrl || '')}`, { method: 'POST' })
											.catch(() => { /* App launch failed silently */ });
									}}
								>
									Launch {window.title}
								</button>
							</div>
						{:else if window.isUserApp && window.userAppUrl && (shouldShowLauncher(window.userAppUrl) || proxyLoadFailed)}
							<!-- Known blocked site OR proxy failed - Show launcher card -->
							<div class="webapp-launcher-container" style="--app-color: {window.color};">
								<div class="webapp-launcher-icon">
									{#if window.userAppLogoUrl}
										<img src={window.userAppLogoUrl} alt={window.title} />
									{:else}
										<div class="icon-placeholder" style="background-color: {window.color};">
											{window.title.charAt(0)}
										</div>
									{/if}
								</div>
								<h3 class="webapp-launcher-title">{window.title}</h3>
								<p class="webapp-launcher-info">
									{#if isElectron}
										Web application ready to launch
									{:else}
										This app opens in a new window for the best experience
									{/if}
								</p>
								<div class="webapp-launcher-actions">
									<button class="webapp-launch-button primary" onclick={openInBrowser}>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6M15 3h6v6M10 14L21 3" />
										</svg>
										Open {window.title}
									</button>
								</div>
								<p class="webapp-launcher-hint">
									Tip: In the desktop app, this will embed directly
								</p>
							</div>
						{:else if window.isUserApp && window.userAppUrl}
							<!-- User Web App -->
							{#if isElectron}
								<!-- In Electron: Use webview for unrestricted embedding -->
								<!-- svelte-ignore element_invalid_self_closing_tag -->
								<webview
									src={window.userAppUrl}
									class="window-iframe electron-webview"
									style="width: 100%; height: 100%;"
									allowpopups={true}
								></webview>
							{:else}
								<!-- In Browser: Use proxy to strip X-Frame-Options/CSP headers -->
								<iframe
									bind:this={iframeElement}
									src={getUserAppIframeSrc(window.userAppUrl)}
									title={window.title}
									class="window-iframe"
									sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-modals allow-top-navigation"
									loading="eager"
									tabindex="0"
									onerror={handleIframeError}
								></iframe>
							{/if}
						{:else}
							<!-- Core Module - Direct embedding (no sandbox for same-origin content) -->
							<iframe
								bind:this={iframeElement}
								src="{getModuleRoute(window.module)}?embed=true"
								title={window.title}
								class="window-iframe"
								loading="eager"
								tabindex="0"
							></iframe>
						{/if}
					{:else}
						<!-- Loading placeholder - will auto-load after staggered delay -->
						<div class="lazy-placeholder">
							<div class="lazy-placeholder-icon" style="background-color: {window.color}20; border-color: {window.color};">
								<span style="color: {window.color};">{window.title.charAt(0).toUpperCase()}</span>
							</div>
							<p class="lazy-placeholder-text">{window.title}</p>
							<p class="lazy-placeholder-hint">Loading...</p>
						</div>
					{/if}
				</div>
			</div>

			<!-- Label under window (hidden when focused) -->
			{#if !isFocused}
				<div class="window-label" style="background-color: {window.color};">
					{window.title}
				</div>
			{/if}
		</div>
	</HTML>
	</T.Group>
	</T.Group>
</T.Group>
{/if}

<style>
	.window-wrapper {
		display: flex;
		flex-direction: column;
		align-items: center;
		cursor: pointer;
		backface-visibility: visible;
		-webkit-backface-visibility: visible;
	}

	.window-wrapper.dimmed {
		opacity: 0.5;
	}

	.window-wrapper.clickable {
		cursor: pointer;
	}

	.window-wrapper.clickable:hover {
		transform: scale(1.02);
	}

	/* Hover effect - shows which window you're about to click */
	.window-wrapper.hovered {
		transform: scale(1.08);
	}

	.window-wrapper.hovered .window-3d {
		box-shadow:
			0 0 0 4px rgba(74, 158, 255, 0.6),
			0 12px 48px rgba(74, 158, 255, 0.3),
			0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.window-wrapper.hovered .window-label {
		transform: scale(1.1);
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.3);
	}

	.window-3d {
		position: relative;
		/* Width/height set dynamically via style attribute */
		width: var(--window-width, 700px);
		height: var(--window-height, 450px);
		background: white;
		border-radius: 10px;
		overflow: hidden;
		box-shadow:
			0 8px 32px rgba(0, 0, 0, 0.15),
			0 2px 8px rgba(0, 0, 0, 0.1);
		display: flex;
		flex-direction: column;
		transition: box-shadow 0.3s ease, transform 0.3s ease;
		border: 1px solid rgba(0, 0, 0, 0.1);
		backface-visibility: visible;
		-webkit-backface-visibility: visible;
	}

	.window-label {
		margin-top: 12px;
		padding: 8px 20px;
		border-radius: 20px;
		color: white;
		font-size: 16px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 1px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
	}

	.window-3d:hover {
		box-shadow:
			0 12px 48px rgba(0, 0, 0, 0.2),
			0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.window-wrapper.focused .window-3d {
		cursor: default;
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.25),
			0 8px 24px rgba(0, 0, 0, 0.2);
	}

	.window-titlebar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 8px 12px;
		background: #f5f5f5;
		border-bottom: 1px solid #e0e0e0;
		user-select: none;
		flex-shrink: 0;
		pointer-events: auto !important;
		position: relative;
		z-index: 10;
	}

	.module-title {
		font-size: 13px;
		font-weight: 600;
		color: #333;
	}

	.open-external-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		padding: 4px;
		background: rgba(59, 130, 246, 0.1);
		border: 1px solid rgba(59, 130, 246, 0.3);
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.15s;
		margin-left: 8px;
	}

	.open-external-btn:hover {
		background: rgba(59, 130, 246, 0.2);
		border-color: rgba(59, 130, 246, 0.5);
	}

	.open-external-btn svg {
		width: 14px;
		height: 14px;
		stroke: #3b82f6;
	}

	.titlebar-right {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.titlebar-controls {
		display: flex;
		gap: 6px;
	}

	.control-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
	}

	/* Size controls in titlebar */
	.size-controls {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 4px 12px;
		background: rgba(74, 158, 255, 0.1);
		border: 1px solid rgba(74, 158, 255, 0.2);
		border-radius: 8px;
		position: relative;
		z-index: 10000;
		pointer-events: auto !important;
		touch-action: manipulation;
		user-select: none;
	}

	.size-btn {
		width: 28px;
		height: 28px;
		padding: 4px;
		background: rgba(74, 158, 255, 0.2);
		border: 1px solid rgba(74, 158, 255, 0.4);
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s;
		pointer-events: auto !important;
		position: relative;
		z-index: 10000;
		touch-action: manipulation;
		user-select: none;
		-webkit-user-select: none;
	}

	.size-btn:hover {
		background: rgba(74, 158, 255, 0.4);
		border-color: rgba(74, 158, 255, 0.6);
		transform: scale(1.1);
	}

	.size-btn:active {
		background: rgba(74, 158, 255, 0.6);
		transform: scale(0.95);
	}

	.size-btn svg {
		width: 12px;
		height: 12px;
		stroke: #333;
	}

	.size-label {
		font-size: 10px;
		color: #666;
		min-width: 60px;
		text-align: center;
	}

	.window-content {
		flex: 1;
		overflow: hidden;
		background: white;
	}

	.window-iframe {
		width: 100%;
		height: 100%;
		border: none;
		pointer-events: none;
		outline: none;
	}

	.window-wrapper.focused .window-iframe {
		pointer-events: auto;
		/* Ensure iframe can receive keyboard events */
		user-select: auto;
	}

	/* Focus indicator for iframe */
	.window-iframe:focus {
		outline: none;
	}

	/* Native App Container */
	.native-app-container {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1.5rem;
		padding: 2rem;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	}

	.native-app-icon {
		width: 128px;
		height: 128px;
		border-radius: 24px;
		overflow: hidden;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
	}

	.native-app-icon img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.icon-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 64px;
		font-weight: bold;
		color: white;
		text-transform: uppercase;
	}

	.native-app-container h3 {
		color: white;
		font-size: 28px;
		font-weight: 700;
		margin: 0;
		text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
	}

	.native-app-info {
		color: rgba(255, 255, 255, 0.9);
		font-size: 16px;
		margin: 0;
	}

	.launch-button {
		background: white;
		color: #667eea;
		border: none;
		padding: 12px 32px;
		font-size: 16px;
		font-weight: 600;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.launch-button:hover {
		transform: translateY(-2px);
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
	}

	.launch-button:active {
		transform: translateY(0);
	}

	/* Lazy Loading Placeholder Styles */
	.lazy-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 12px;
		background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
		cursor: pointer;
	}

	.lazy-placeholder-icon {
		width: 80px;
		height: 80px;
		border-radius: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		border: 2px solid;
		transition: transform 0.2s ease;
	}

	.lazy-placeholder-icon span {
		font-size: 36px;
		font-weight: 700;
		text-transform: uppercase;
	}

	.lazy-placeholder-text {
		font-size: 16px;
		font-weight: 600;
		color: #334155;
		margin: 0;
	}

	.lazy-placeholder-hint {
		font-size: 12px;
		color: #94a3b8;
		margin: 0;
	}

	.window-wrapper:hover .lazy-placeholder-icon {
		transform: scale(1.05);
	}

	/* Web App Launcher Container - Modern card design for external apps */
	.webapp-launcher-container {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1.25rem;
		padding: 2rem;
		background: linear-gradient(
			135deg,
			color-mix(in srgb, var(--app-color, #6366f1) 8%, white) 0%,
			color-mix(in srgb, var(--app-color, #6366f1) 15%, #f8fafc) 100%
		);
		position: relative;
		overflow: hidden;
	}

	/* Subtle background pattern */
	.webapp-launcher-container::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image: radial-gradient(
			circle at 50% 50%,
			color-mix(in srgb, var(--app-color, #6366f1) 10%, transparent) 0%,
			transparent 50%
		);
		pointer-events: none;
	}

	.webapp-launcher-icon {
		width: 96px;
		height: 96px;
		border-radius: 24px;
		overflow: hidden;
		box-shadow:
			0 10px 30px color-mix(in srgb, var(--app-color, #6366f1) 30%, transparent),
			0 4px 12px rgba(0, 0, 0, 0.1);
		position: relative;
		z-index: 1;
	}

	.webapp-launcher-icon img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.webapp-launcher-title {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 700;
		color: #1e293b;
		text-align: center;
		position: relative;
		z-index: 1;
	}

	.webapp-launcher-info {
		margin: 0;
		font-size: 0.9rem;
		color: #64748b;
		text-align: center;
		max-width: 280px;
		line-height: 1.5;
		position: relative;
		z-index: 1;
	}

	.webapp-launcher-actions {
		display: flex;
		gap: 0.75rem;
		position: relative;
		z-index: 1;
	}

	.webapp-launch-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.875rem 1.75rem;
		border: none;
		border-radius: 12px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.webapp-launch-button.primary {
		background: var(--app-color, #6366f1);
		color: white;
		box-shadow:
			0 4px 14px color-mix(in srgb, var(--app-color, #6366f1) 40%, transparent),
			0 2px 4px rgba(0, 0, 0, 0.1);
	}

	.webapp-launch-button.primary:hover {
		transform: translateY(-2px);
		box-shadow:
			0 6px 20px color-mix(in srgb, var(--app-color, #6366f1) 50%, transparent),
			0 4px 8px rgba(0, 0, 0, 0.15);
	}

	.webapp-launch-button.primary:active {
		transform: translateY(0);
	}

	.webapp-launch-button svg {
		width: 18px;
		height: 18px;
	}

	.webapp-launcher-hint {
		margin: 0;
		font-size: 0.75rem;
		color: #94a3b8;
		text-align: center;
		position: relative;
		z-index: 1;
		margin-top: 0.5rem;
	}
</style>
