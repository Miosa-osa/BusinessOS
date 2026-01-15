<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		createWindowCaptureService,
		resetWindowCaptureConnection,
		type WindowInfo,
		type WindowCaptureService
	} from '$lib/services/windowCapture.service';

	// Module-level tracking to prevent spam across component remounts
	// This persists even when the component is destroyed and recreated
	const failedBundleIds = new Set<string>();
	// Track if we've already logged an error for this bundle (reduces spam)
	const loggedBundleIds = new Set<string>();
	// Track bundles we've already tried to initialize (prevents $effect re-runs)
	const initializedBundles = new Set<string>();

	interface Props {
		bundleId: string;
		appName: string;
		logoUrl?: string;
		color?: string;
	}

	let { bundleId, appName, logoUrl, color = '#6366F1' }: Props = $props();

	// Check if already failed BEFORE creating state (avoids setting state in effect)
	const alreadyFailed = failedBundleIds.has(bundleId);

	// State - initialize based on whether already failed
	let captureService: WindowCaptureService | null = $state(null);
	let status = $state<'connecting' | 'permission' | 'selecting' | 'capturing' | 'error' | 'stopped'>(
		alreadyFailed ? 'error' : 'connecting'
	);
	let errorMessage = $state<string | null>(
		alreadyFailed ? 'Window capture service is not available' : null
	);
	let windows = $state<WindowInfo[]>([]);
	let currentFrame = $state<string | null>(null);
	let isLaunching = $state(false);
	let permissionGranted = $state(false);

	// Canvas reference for rendering frames
	let canvasRef: HTMLCanvasElement | undefined = $state();
	let canvasCtx: CanvasRenderingContext2D | null = $state(null);

	// Frame dimensions
	let frameWidth = $state(0);
	let frameHeight = $state(0);

	let mounted = false;

	onMount(() => {
		mounted = true;
		initializeCapture();
	});

	onDestroy(() => {
		mounted = false;
		captureService?.disconnect();
	});

	// Render frame to canvas (lazily initializes canvas context)
	function renderFrame(base64Data: string) {
		if (!canvasRef) return;
		// Lazily initialize canvas context
		if (!canvasCtx) {
			canvasCtx = canvasRef.getContext('2d');
		}
		if (!canvasCtx) return;

		const img = new Image();
		img.onload = () => {
			// Update canvas size if needed
			if (frameWidth !== img.width || frameHeight !== img.height) {
				frameWidth = img.width;
				frameHeight = img.height;
				canvasRef!.width = img.width;
				canvasRef!.height = img.height;
			}

			// Draw the frame
			canvasCtx!.drawImage(img, 0, 0);
		};
		img.src = `data:image/jpeg;base64,${base64Data}`;
	}

	// Launch the app
	async function launchApp() {
		isLaunching = true;
		try {
			const response = await fetch(`/api/system/launch-app?bundle_id=${encodeURIComponent(bundleId)}`, {
				method: 'POST',
				credentials: 'include'
			});

			if (response.ok) {
				// Wait a bit for the app to start, then refresh windows
				setTimeout(() => {
					captureService?.listWindows(bundleId);
				}, 2000);
			}
		} catch (error) {
			console.error('[NativeAppCapture] Failed to launch app:', error);
		} finally {
			isLaunching = false;
		}
	}

	// Select a window to capture
	function selectWindow(windowId: number) {
		captureService?.selectWindow(windowId);
	}

	// Stop capture
	function stopCapture() {
		captureService?.stopCapture();
	}

	// Retry after error - manual retry triggered by user
	function retry() {
		errorMessage = null;
		status = 'connecting';
		// Clear this bundle ID from the failed sets to allow retry
		failedBundleIds.delete(bundleId);
		loggedBundleIds.delete(bundleId);
		initializedBundles.delete(bundleId);
		// Reset global connection state (clears localStorage)
		resetWindowCaptureConnection();
		// Disconnect existing service
		captureService?.disconnect();
		captureService = null;

		// Manually re-initialize (since we're using onMount, not $effect)
		initializeCapture();
	}

	// Initialize capture service - called from onMount and retry
	function initializeCapture() {
		// Skip if already failed
		if (failedBundleIds.has(bundleId)) {
			status = 'error';
			errorMessage = 'Window capture service is not available';
			return;
		}

		// Skip if already initialized for this bundle
		if (initializedBundles.has(bundleId)) {
			return;
		}
		initializedBundles.add(bundleId);

		captureService = createWindowCaptureService({
			onConnect: () => {
				if (!mounted) return;
				console.log('[NativeAppCapture] Connected');
				status = 'permission';
			},
			onDisconnect: () => {
				if (!mounted) return;
				console.log('[NativeAppCapture] Disconnected');
				if (status !== 'error') {
					status = 'stopped';
				}
			},
			onPermission: (granted) => {
				if (!mounted) return;
				console.log('[NativeAppCapture] Permission:', granted);
				permissionGranted = granted;
				if (granted) {
					captureService?.startCapture({ bundleId, quality: 0.7, fps: 30 });
					status = 'selecting';
				} else {
					status = 'permission';
				}
			},
			onWindows: (windowList) => {
				if (!mounted) return;
				console.log('[NativeAppCapture] Windows:', windowList);
				windows = windowList;
				if (windowList.length === 0) {
					status = 'selecting';
				}
			},
			onStarted: (windowId, fps, quality) => {
				if (!mounted) return;
				console.log('[NativeAppCapture] Capture started:', windowId, fps, quality);
				status = 'capturing';
			},
			onFrame: (data, windowId) => {
				if (!mounted) return;
				currentFrame = data;
				renderFrame(data);
			},
			onStopped: () => {
				if (!mounted) return;
				console.log('[NativeAppCapture] Capture stopped');
				status = 'stopped';
				currentFrame = null;
			},
			onError: (error) => {
				if (!mounted) return;
				// Only log once per bundle ID to prevent spam
				if (!loggedBundleIds.has(bundleId)) {
					console.error('[NativeAppCapture] Error:', error);
					loggedBundleIds.add(bundleId);
				}
				// Mark this bundle ID as failed to prevent future spam
				failedBundleIds.add(bundleId);
				errorMessage = error;
				status = 'error';
			}
		});

		captureService.connect();
	}

	// Input event handlers
	function getCanvasCoordinates(event: MouseEvent): { x: number; y: number } {
		if (!canvasRef) return { x: 0, y: 0 };

		const rect = canvasRef.getBoundingClientRect();
		const scaleX = frameWidth / rect.width;
		const scaleY = frameHeight / rect.height;

		return {
			x: Math.round((event.clientX - rect.left) * scaleX),
			y: Math.round((event.clientY - rect.top) * scaleY)
		};
	}

	function getModifiers(event: MouseEvent | KeyboardEvent): number {
		let modifiers = 0;
		if (event.shiftKey) modifiers |= 1;
		if (event.ctrlKey) modifiers |= 2;
		if (event.altKey) modifiers |= 4;
		if (event.metaKey) modifiers |= 8;
		return modifiers;
	}

	function sendInputEvent(type: string, data: Record<string, unknown>) {
		if (!captureService?.isConnected()) return;

		const ws = (captureService as unknown as { ws: WebSocket | null }).ws;
		if (!ws) return;

		ws.send(JSON.stringify({
			type: 'input',
			payload: { type, ...data }
		}));
	}

	// NOTE: Mouse move events are intentionally NOT forwarded to avoid hijacking the system cursor.
	// The backend ignores mousemove events anyway. We only forward clicks, scrolls, and keyboard.
	// If we need visual hover effects in the future, handle them purely client-side here.
	function handleMouseMove(_event: MouseEvent) {
		// No-op: Don't forward mouse moves to backend
	}

	function handleMouseDown(event: MouseEvent) {
		const { x, y } = getCanvasCoordinates(event);
		sendInputEvent('mousedown', { x, y, button: event.button, modifiers: getModifiers(event) });
	}

	function handleMouseUp(event: MouseEvent) {
		const { x, y } = getCanvasCoordinates(event);
		sendInputEvent('mouseup', { x, y, button: event.button, modifiers: getModifiers(event) });
	}

	function handleClick(event: MouseEvent) {
		const { x, y } = getCanvasCoordinates(event);
		sendInputEvent('click', { x, y, button: event.button, modifiers: getModifiers(event) });
	}

	function handleDblClick(event: MouseEvent) {
		const { x, y } = getCanvasCoordinates(event);
		sendInputEvent('dblclick', { x, y });
	}

	function handleWheel(event: WheelEvent) {
		event.preventDefault();
		const { x, y } = getCanvasCoordinates(event);
		sendInputEvent('scroll', { x, y, deltaX: Math.round(event.deltaX), deltaY: Math.round(event.deltaY) });
	}

	function handleKeyDown(event: KeyboardEvent) {
		event.preventDefault();
		sendInputEvent('keydown', { keyCode: event.keyCode, modifiers: getModifiers(event) });

		// Also send character for text input
		if (event.key.length === 1) {
			sendInputEvent('char', { char: event.key });
		}
	}

	function handleKeyUp(event: KeyboardEvent) {
		event.preventDefault();
		sendInputEvent('keyup', { keyCode: event.keyCode, modifiers: getModifiers(event) });
	}

	function handleContextMenu(event: MouseEvent) {
		event.preventDefault();
		const { x, y } = getCanvasCoordinates(event);
		sendInputEvent('click', { x, y, button: 2, modifiers: getModifiers(event) });
	}
</script>

<div class="native-capture-container">
	{#if status === 'connecting'}
		<div class="capture-status">
			<div class="loading-spinner"></div>
			<p>Connecting to capture service...</p>
		</div>
	{:else if status === 'permission' && !permissionGranted}
		<div class="capture-permission">
			<div class="app-icon">
				{#if logoUrl}
					<img src={logoUrl} alt={appName} />
				{:else}
					<div class="icon-placeholder" style="background-color: {color};">
						{appName.charAt(0)}
					</div>
				{/if}
			</div>
			<h3>Screen Recording Permission Required</h3>
			<p>To display {appName} inside BusinessOS, you need to grant screen recording permission.</p>
			<p class="permission-instructions">
				Go to System Settings &gt; Privacy &amp; Security &gt; Screen Recording and enable BusinessOS.
			</p>
			<button class="retry-button" onclick={retry}>
				Check Again
			</button>
		</div>
	{:else if status === 'selecting'}
		<div class="capture-selecting">
			<div class="app-icon">
				{#if logoUrl}
					<img src={logoUrl} alt={appName} />
				{:else}
					<div class="icon-placeholder" style="background-color: {color};">
						{appName.charAt(0)}
					</div>
				{/if}
			</div>
			{#if windows.length === 0}
				<h3>{appName} is not running</h3>
				<p>Launch the app to view it inside BusinessOS.</p>
				<button class="launch-button" onclick={launchApp} disabled={isLaunching}>
					{isLaunching ? 'Launching...' : `Launch ${appName}`}
				</button>
			{:else}
				<h3>Select a Window</h3>
				<p>Choose which {appName} window to display:</p>
				<div class="window-list">
					{#each windows as win (win.window_id)}
						<button class="window-option" onclick={() => selectWindow(win.window_id)}>
							<span class="window-name">{win.window_name || appName}</span>
							<span class="window-size">{win.width} x {win.height}</span>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	{:else if status === 'capturing'}
		<div class="capture-view">
			<!-- svelte-ignore a11y_positive_tabindex -->
			<canvas
				bind:this={canvasRef}
				class="capture-canvas"
				tabindex="1"
				onmousemove={handleMouseMove}
				onmousedown={handleMouseDown}
				onmouseup={handleMouseUp}
				onclick={handleClick}
				ondblclick={handleDblClick}
				onwheel={handleWheel}
				onkeydown={handleKeyDown}
				onkeyup={handleKeyUp}
				oncontextmenu={handleContextMenu}
			></canvas>
			<div class="capture-controls">
				<button class="stop-button" onclick={stopCapture}>
					Stop Capture
				</button>
			</div>
		</div>
	{:else if status === 'error'}
		<div class="capture-error">
			<div class="error-icon">!</div>
			<h3>Capture Error</h3>
			<p>{errorMessage || 'An unknown error occurred'}</p>
			<button class="retry-button" onclick={retry}>
				Retry
			</button>
		</div>
	{:else if status === 'stopped'}
		<div class="capture-stopped">
			<div class="app-icon">
				{#if logoUrl}
					<img src={logoUrl} alt={appName} />
				{:else}
					<div class="icon-placeholder" style="background-color: {color};">
						{appName.charAt(0)}
					</div>
				{/if}
			</div>
			<h3>Capture Stopped</h3>
			<button class="retry-button" onclick={retry}>
				Start Again
			</button>
		</div>
	{/if}
</div>

<style>
	.native-capture-container {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
		color: white;
	}

	.capture-status,
	.capture-permission,
	.capture-selecting,
	.capture-error,
	.capture-stopped {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		padding: 2rem;
		max-width: 400px;
	}

	.capture-view {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		position: relative;
	}

	.capture-canvas {
		flex: 1;
		width: 100%;
		cursor: default;
		outline: none;
		object-fit: contain;
		background: #000;
	}

	.capture-controls {
		position: absolute;
		top: 8px;
		right: 8px;
		opacity: 0;
		transition: opacity 0.2s;
	}

	.capture-view:hover .capture-controls {
		opacity: 1;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid rgba(255, 255, 255, 0.2);
		border-top-color: #fff;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.app-icon {
		width: 80px;
		height: 80px;
		margin-bottom: 1.5rem;
	}

	.app-icon img {
		width: 100%;
		height: 100%;
		object-fit: contain;
		border-radius: 16px;
	}

	.icon-placeholder {
		width: 100%;
		height: 100%;
		border-radius: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		font-weight: 600;
		color: white;
	}

	h3 {
		font-size: 1.25rem;
		font-weight: 600;
		margin: 0 0 0.5rem 0;
	}

	p {
		color: rgba(255, 255, 255, 0.7);
		margin: 0 0 1rem 0;
		line-height: 1.5;
	}

	.permission-instructions {
		font-size: 0.875rem;
		padding: 0.75rem;
		background: rgba(255, 255, 255, 0.1);
		border-radius: 8px;
		margin-bottom: 1.5rem;
	}

	.launch-button,
	.retry-button,
	.stop-button {
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		border: none;
	}

	.launch-button {
		background: #4f46e5;
		color: white;
	}

	.launch-button:hover:not(:disabled) {
		background: #4338ca;
	}

	.launch-button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.retry-button {
		background: #3b82f6;
		color: white;
	}

	.retry-button:hover {
		background: #2563eb;
	}

	.stop-button {
		background: rgba(239, 68, 68, 0.9);
		color: white;
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
	}

	.stop-button:hover {
		background: rgba(220, 38, 38, 0.9);
	}

	.window-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		width: 100%;
		max-width: 300px;
	}

	.window-option {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 1rem;
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
		border-radius: 8px;
		color: white;
		cursor: pointer;
		transition: all 0.2s;
	}

	.window-option:hover {
		background: rgba(255, 255, 255, 0.2);
		border-color: rgba(255, 255, 255, 0.3);
	}

	.window-name {
		font-weight: 500;
	}

	.window-size {
		font-size: 0.75rem;
		color: rgba(255, 255, 255, 0.5);
	}

	.error-icon {
		width: 60px;
		height: 60px;
		border-radius: 50%;
		background: rgba(239, 68, 68, 0.2);
		border: 2px solid #ef4444;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		font-weight: bold;
		color: #ef4444;
		margin-bottom: 1rem;
	}
</style>
