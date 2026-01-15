/**
 * Terminal WebSocket Service
 * Handles bi-directional communication with the terminal backend
 */

export interface TerminalMessage {
	type: 'input' | 'output' | 'resize' | 'heartbeat' | 'error' | 'status';
	session_id?: string;
	data?: string;
	metadata?: Record<string, unknown>;
}

export interface TerminalConfig {
	cols?: number;
	rows?: number;
	shell?: string;
	cwd?: string;
	mode?: 'docker' | 'local'; // Terminal mode: docker (containerized) or local (Mac)
}

export type TerminalEventHandler = {
	onData: (data: string) => void;
	onConnect: (sessionId: string, metadata: Record<string, unknown>) => void;
	onDisconnect: () => void;
	onError: (error: string) => void;
};

export class TerminalService {
	private ws: WebSocket | null = null;
	private sessionId: string | null = null;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private heartbeatInterval: ReturnType<typeof setInterval> | null = null;
	private config: TerminalConfig;

	constructor(
		private baseUrl: string,
		private handlers: TerminalEventHandler,
		config: TerminalConfig = {}
	) {
		this.config = {
			cols: config.cols ?? 80,
			rows: config.rows ?? 24,
			shell: config.shell ?? 'zsh',
			cwd: config.cwd ?? '',
			mode: config.mode // CRITICAL: Copy mode from config parameter
		};
	}

	/**
	 * Connect to the terminal WebSocket
	 */
	connect(): void {
		// Build WebSocket URL with query params
		const wsProtocol = this.baseUrl.startsWith('https') ? 'wss' : 'ws';
		const wsBase = this.baseUrl.replace(/^https?/, wsProtocol);

		const params = new URLSearchParams({
			cols: String(this.config.cols),
			rows: String(this.config.rows),
			shell: this.config.shell || 'zsh'
		});

		if (this.config.cwd) {
			params.set('cwd', this.config.cwd);
		}

		// Add mode parameter if specified
		if (this.config.mode) {
			params.set('mode', this.config.mode);
		}

		const wsUrl = `${wsBase}/api/terminal/ws?${params.toString()}`;

		// DEBUG: Log WebSocket URL to verify mode is being sent
		console.log('[TerminalService] 🔧 Connecting with config:', this.config);
		console.log('[TerminalService] 🔧 WebSocket URL:', wsUrl);
		console.log('[TerminalService] 🔧 Mode parameter:', this.config.mode || '(not set)');

		try {
			this.ws = new WebSocket(wsUrl);
			this.setupEventHandlers();
		} catch (error) {
			console.error('Failed to create WebSocket:', error);
			this.handlers.onError('Failed to connect to terminal');
		}
	}

	private setupEventHandlers(): void {
		if (!this.ws) return;

		this.ws.onopen = () => {
			console.log('Terminal WebSocket connected');
			this.reconnectAttempts = 0;
			this.startHeartbeat();
		};

		this.ws.onmessage = (event) => {
			// REDUCED LOGGING: Only log errors, not every message
			try {
				const message: TerminalMessage = JSON.parse(event.data);
				this.handleMessage(message);
			} catch (error) {
				// If not JSON, treat as raw output
				this.handlers.onData(event.data);
			}
		};

		this.ws.onerror = (error) => {
			console.error('WebSocket error:', error);
			this.handlers.onError('Connection error');
		};

		this.ws.onclose = (event) => {
			console.log('WebSocket closed:', event.code, event.reason);
			this.stopHeartbeat();
			this.handlers.onDisconnect();

			if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
				this.attemptReconnect();
			}
		};
	}

	private handleMessage(message: TerminalMessage): void {
		// REDUCED LOGGING: Only log errors, not every message

		switch (message.type) {
			case 'output':
				if (message.data) {
					this.handlers.onData(message.data);
				}
				break;

			case 'status':
				console.log('[Terminal] Status message:', message.data, 'metadata:', message.metadata);
				if (message.data === 'connected' && message.metadata?.session_id) {
					this.sessionId = message.metadata.session_id as string;
					console.log('[Terminal] Calling onConnect with session:', this.sessionId);
					this.handlers.onConnect(this.sessionId, message.metadata);
				}
				break;

			case 'error':
				this.handlers.onError(message.data || 'Unknown error');
				break;
		}
	}

	/**
	 * Send user input to the terminal
	 */
	sendInput(data: string): void {
		if (!this.isConnected()) {
			console.warn('WebSocket not connected');
			return;
		}

		const message: TerminalMessage = {
			type: 'input',
			session_id: this.sessionId || undefined,
			data: data
		};

		this.ws!.send(JSON.stringify(message));
	}

	/**
	 * Send terminal resize event
	 */
	resize(cols: number, rows: number): void {
		if (!this.isConnected()) return;

		this.config.cols = cols;
		this.config.rows = rows;

		const message: TerminalMessage = {
			type: 'resize',
			session_id: this.sessionId || undefined,
			data: JSON.stringify({ cols, rows })
		};

		this.ws!.send(JSON.stringify(message));
	}

	/**
	 * Send heartbeat to keep connection alive
	 */
	private sendHeartbeat(): void {
		if (!this.isConnected()) return;

		const message: TerminalMessage = {
			type: 'heartbeat',
			session_id: this.sessionId || undefined
		};

		this.ws!.send(JSON.stringify(message));
	}

	private startHeartbeat(): void {
		this.stopHeartbeat();
		this.heartbeatInterval = setInterval(() => {
			this.sendHeartbeat();
		}, 30000); // Every 30 seconds
	}

	private stopHeartbeat(): void {
		if (this.heartbeatInterval) {
			clearInterval(this.heartbeatInterval);
			this.heartbeatInterval = null;
		}
	}

	/**
	 * Attempt to reconnect with exponential backoff
	 */
	private attemptReconnect(): void {
		this.reconnectAttempts++;
		const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 10000);

		console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

		setTimeout(() => {
			this.connect();
		}, delay);
	}

	/**
	 * Disconnect from the terminal
	 */
	disconnect(): void {
		this.stopHeartbeat();

		if (this.ws) {
			this.ws.close(1000, 'Client disconnect');
			this.ws = null;
		}

		this.sessionId = null;
	}

	/**
	 * Check if WebSocket is connected
	 */
	isConnected(): boolean {
		return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
	}

	/**
	 * Get current session ID
	 */
	getSessionId(): string | null {
		return this.sessionId;
	}
}

/**
 * Create a terminal service instance with the default API URL
 */
export function createTerminalService(
	handlers: TerminalEventHandler,
	config?: TerminalConfig
): TerminalService {
	// Check for Electron or custom API URL
	const customUrl = typeof window !== 'undefined'
		? (window as unknown as { __API_URL__?: string }).__API_URL__
		: undefined;

	// In dev mode, use Vite proxy (same origin) so cookies are sent automatically
	// In Electron, use the custom URL or direct backend
	// The Vite proxy at /api/terminal has ws: true for WebSocket support
	// Check for dev ports (5173, 5174, etc.) - Vite may use alternate ports if primary is busy
	const isDev = typeof window !== 'undefined' &&
		(window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') &&
		/^517[0-9]$/.test(window.location.port);

	const apiUrl = customUrl || (isDev ? window.location.origin : 'http://localhost:8001');

	console.log('[Terminal] Connecting to backend:', apiUrl, isDev ? '(via Vite proxy)' : '(direct)');

	return new TerminalService(apiUrl, handlers, config);
}
