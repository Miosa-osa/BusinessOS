/**
 * Window Capture WebSocket Service
 * Handles streaming of native macOS app windows to the frontend
 */

export interface WindowInfo {
	window_id: number;
	owner_pid: number;
	owner_name: string;
	window_name: string;
	x: number;
	y: number;
	width: number;
	height: number;
	layer: number;
	is_onscreen: boolean;
}

export interface CaptureMessage {
	type: 'permission' | 'windows' | 'started' | 'frame' | 'stopped' | 'error';
	payload?: unknown;
	error?: string;
}

export interface FramePayload {
	window_id: number;
	width?: number;
	height?: number;
	data: string; // Base64 encoded JPEG
}

export interface WindowListPayload {
	windows: WindowInfo[];
}

export interface PermissionPayload {
	granted: boolean;
}

export interface CaptureConfig {
	bundleId: string;
	quality?: number; // 0.0 to 1.0, default 0.7
	fps?: number; // Frames per second, default 30
}

export type CaptureEventHandler = {
	onPermission: (granted: boolean) => void;
	onWindows: (windows: WindowInfo[]) => void;
	onStarted: (windowId: number, fps: number, quality: number) => void;
	onFrame: (data: string, windowId: number) => void;
	onStopped: () => void;
	onError: (error: string) => void;
	onConnect: () => void;
	onDisconnect: () => void;
};

export class WindowCaptureService {
	private ws: WebSocket | null = null;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 3;
	private config: CaptureConfig | null = null;
	private isConnecting = false;
	private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
	private permanentlyFailed = false; // Prevent infinite reconnection loops

	constructor(
		private baseUrl: string,
		private handlers: CaptureEventHandler
	) {}

	/**
	 * Connect to the window capture WebSocket
	 */
	connect(): void {
		// Prevent connection if no base URL (dummy service) - silently skip
		if (!this.baseUrl) {
			return;
		}

		// Prevent connection if permanently failed - silently skip
		if (this.permanentlyFailed) {
			return;
		}

		// Prevent multiple simultaneous connections
		if (this.isConnecting || this.isConnected()) {
			console.log('[WindowCapture] Already connected or connecting, skipping');
			return;
		}

		this.isConnecting = true;

		const wsProtocol = this.baseUrl.startsWith('https') ? 'wss' : 'ws';
		const wsBase = this.baseUrl.replace(/^https?/, wsProtocol);
		const wsUrl = `${wsBase}/api/window-capture/stream`;

		console.log('[WindowCapture] Connecting to:', wsUrl);

		try {
			this.ws = new WebSocket(wsUrl);
			this.setupEventHandlers();
		} catch (error) {
			this.isConnecting = false;
			this.permanentlyFailed = true; // Mark as permanently failed on immediate error
			console.error('[WindowCapture] Failed to create WebSocket:', error);
			this.handlers.onError('Failed to connect to window capture service');
		}
	}

	private setupEventHandlers(): void {
		if (!this.ws) return;

		this.ws.onopen = () => {
			console.log('[WindowCapture] WebSocket connected');
			this.isConnecting = false;
			this.reconnectAttempts = 0;
			this.handlers.onConnect();
		};

		this.ws.onmessage = (event) => {
			try {
				const message: CaptureMessage = JSON.parse(event.data);
				this.handleMessage(message);
			} catch (error) {
				console.error('[WindowCapture] Failed to parse message:', error);
			}
		};

		this.ws.onerror = (error) => {
			console.error('[WindowCapture] WebSocket error:', error);
			this.isConnecting = false;
			// Don't call onError here - let onclose handle it to avoid double-triggering
		};

		this.ws.onclose = (event) => {
			console.log('[WindowCapture] WebSocket closed:', event.code, event.reason);
			this.isConnecting = false;

			// Only attempt reconnect for unexpected closures, not clean disconnects
			if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
				this.attemptReconnect();
			} else if (!event.wasClean) {
				// Max reconnect attempts reached - mark as permanently failed
				this.permanentlyFailed = true;
				console.log('[WindowCapture] Max reconnect attempts reached, giving up');
				this.handlers.onError('Failed to connect after multiple attempts. Window capture service may not be running.');
			}

			this.handlers.onDisconnect();
		};
	}

	private handleMessage(message: CaptureMessage): void {
		switch (message.type) {
			case 'permission': {
				const payload = message.payload as PermissionPayload;
				this.handlers.onPermission(payload.granted);
				break;
			}

			case 'windows': {
				const payload = message.payload as WindowListPayload;
				this.handlers.onWindows(payload.windows);
				break;
			}

			case 'started': {
				const payload = message.payload as { window_id: number; fps: number; quality: number };
				this.handlers.onStarted(payload.window_id, payload.fps, payload.quality);
				break;
			}

			case 'frame': {
				const payload = message.payload as FramePayload;
				this.handlers.onFrame(payload.data, payload.window_id);
				break;
			}

			case 'stopped':
				this.handlers.onStopped();
				break;

			case 'error':
				this.handlers.onError(message.error || 'Unknown error');
				break;
		}
	}

	/**
	 * Start capturing a window for the given bundle ID
	 */
	startCapture(config: CaptureConfig): void {
		if (!this.isConnected()) {
			console.warn('[WindowCapture] Not connected');
			return;
		}

		this.config = config;

		const message = {
			type: 'start',
			payload: {
				bundle_id: config.bundleId,
				quality: config.quality ?? 0.7,
				fps: config.fps ?? 30
			}
		};

		this.ws!.send(JSON.stringify(message));
	}

	/**
	 * Select a specific window to capture
	 */
	selectWindow(windowId: number): void {
		if (!this.isConnected()) {
			console.warn('[WindowCapture] Not connected');
			return;
		}

		const message = {
			type: 'select_window',
			payload: {
				window_id: windowId
			}
		};

		this.ws!.send(JSON.stringify(message));
	}

	/**
	 * List windows for a bundle ID
	 */
	listWindows(bundleId: string): void {
		if (!this.isConnected()) {
			console.warn('[WindowCapture] Not connected');
			return;
		}

		const message = {
			type: 'list_windows',
			payload: {
				bundle_id: bundleId
			}
		};

		this.ws!.send(JSON.stringify(message));
	}

	/**
	 * Stop the current capture
	 */
	stopCapture(): void {
		if (!this.isConnected()) return;

		const message = {
			type: 'stop'
		};

		this.ws!.send(JSON.stringify(message));
	}

	private attemptReconnect(): void {
		// Clear any pending reconnect timeout
		if (this.reconnectTimeoutId) {
			clearTimeout(this.reconnectTimeoutId);
			this.reconnectTimeoutId = null;
		}

		this.reconnectAttempts++;
		const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 10000);

		console.log(`[WindowCapture] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

		this.reconnectTimeoutId = setTimeout(() => {
			this.reconnectTimeoutId = null;
			this.connect();
		}, delay);
	}

	/**
	 * Disconnect from the window capture service
	 */
	disconnect(): void {
		// Clear any pending reconnect timeout
		if (this.reconnectTimeoutId) {
			clearTimeout(this.reconnectTimeoutId);
			this.reconnectTimeoutId = null;
		}

		// Reset state
		this.isConnecting = false;
		this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection attempts

		if (this.ws) {
			this.ws.close(1000, 'Client disconnect');
			this.ws = null;
		}
	}

	/**
	 * Check if WebSocket is connected
	 */
	isConnected(): boolean {
		return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
	}
}

// Global flag to prevent creating services if connection is known to fail
// Using localStorage to survive HMR and page refreshes
const STORAGE_KEY = 'windowCapture_connectionFailed';
const STORAGE_TIME_KEY = 'windowCapture_failedAt';
const FAILURE_EXPIRY_MS = 5 * 60 * 1000; // Reset after 5 minutes

function isConnectionFailed(): boolean {
	if (typeof window === 'undefined') return false;
	const failed = localStorage.getItem(STORAGE_KEY) === 'true';
	if (failed) {
		// Check if the failure has expired
		const failedAt = parseInt(localStorage.getItem(STORAGE_TIME_KEY) || '0', 10);
		if (Date.now() - failedAt > FAILURE_EXPIRY_MS) {
			localStorage.removeItem(STORAGE_KEY);
			localStorage.removeItem(STORAGE_TIME_KEY);
			return false;
		}
	}
	return failed;
}

function markConnectionFailed(): void {
	if (typeof window === 'undefined') return;
	localStorage.setItem(STORAGE_KEY, 'true');
	localStorage.setItem(STORAGE_TIME_KEY, String(Date.now()));
}

// Module-level tracking (also persisted in localStorage)
let globalConnectionAttempts = 0;
const MAX_GLOBAL_ATTEMPTS = 3;
// Log suppression - only log once per session
let hasLoggedFailure = false;

/**
 * Create a window capture service instance
 */
export function createWindowCaptureService(handlers: CaptureEventHandler): WindowCaptureService {
	// Check if we've already determined that connections will fail (persisted)
	if (isConnectionFailed()) {
		// Only log once to prevent spam
		if (!hasLoggedFailure) {
			console.log('[WindowCapture] Connection previously failed, returning dummy service');
			hasLoggedFailure = true;
		}
		// Return a dummy service that immediately reports error
		setTimeout(() => {
			handlers.onError('Window capture service is not available');
		}, 0);
		return new WindowCaptureService('', {
			...handlers,
			onConnect: () => {},
			onDisconnect: () => {},
		});
	}

	// Track global connection attempts
	globalConnectionAttempts++;
	if (globalConnectionAttempts > MAX_GLOBAL_ATTEMPTS) {
		markConnectionFailed();
		console.log('[WindowCapture] Too many connection attempts, marking as failed for 5 minutes');
		setTimeout(() => {
			handlers.onError('Window capture service is not available');
		}, 0);
		return new WindowCaptureService('', handlers);
	}

	// Check for Electron or custom API URL
	const customUrl =
		typeof window !== 'undefined'
			? (window as unknown as { __API_URL__?: string }).__API_URL__
			: undefined;

	// In dev mode, use Vite proxy (same origin) so cookies are sent automatically
	const isDev =
		typeof window !== 'undefined' &&
		(window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') &&
		/^517[0-9]$/.test(window.location.port);

	const apiUrl = customUrl || (isDev ? window.location.origin : 'http://localhost:8001');

	console.log('[WindowCapture] Connecting to backend:', apiUrl, isDev ? '(via Vite proxy)' : '(direct)', `(attempt ${globalConnectionAttempts})`);

	return new WindowCaptureService(apiUrl, handlers);
}

/**
 * Reset global connection state (call when user explicitly wants to retry)
 */
export function resetWindowCaptureConnection(): void {
	// Clear localStorage
	if (typeof window !== 'undefined') {
		localStorage.removeItem(STORAGE_KEY);
		localStorage.removeItem(STORAGE_TIME_KEY);
	}
	// Reset module-level state
	globalConnectionAttempts = 0;
	hasLoggedFailure = false;
}

/**
 * Check screen capture permission via REST API
 */
export async function checkCapturePermission(): Promise<boolean> {
	try {
		const response = await fetch('/api/window-capture/permission', {
			credentials: 'include'
		});
		const data = await response.json();
		return data.granted ?? false;
	} catch (error) {
		console.error('[WindowCapture] Failed to check permission:', error);
		return false;
	}
}

/**
 * Request screen capture permission via REST API
 */
export async function requestCapturePermission(): Promise<void> {
	try {
		await fetch('/api/window-capture/permission', {
			method: 'POST',
			credentials: 'include'
		});
	} catch (error) {
		console.error('[WindowCapture] Failed to request permission:', error);
	}
}

/**
 * List windows for a bundle ID via REST API
 */
export async function listWindowsForApp(bundleId: string): Promise<WindowInfo[]> {
	try {
		const response = await fetch(`/api/window-capture/windows?bundle_id=${encodeURIComponent(bundleId)}`, {
			credentials: 'include'
		});
		const data = await response.json();
		return data.windows ?? [];
	} catch (error) {
		console.error('[WindowCapture] Failed to list windows:', error);
		return [];
	}
}
