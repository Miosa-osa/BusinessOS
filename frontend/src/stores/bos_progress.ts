/**
 * BOS Progress Store — Real-time tracking of BusinessOS process mining operations
 *
 * This Svelte store manages real-time progress updates from BOS (BusinessOS Data Layer)
 * process discovery and conformance checking. It subscribes to Server-Sent Events (SSE)
 * and maintains reactive state for progress display in the UI.
 *
 * Architecture:
 *   BOS (Rust) -> SSE -> Go Handler -> store updates -> Svelte components
 *
 * Usage in components:
 *   import { bosProgress, subscribe } from './bos_progress';
 *
 *   onMount(() => {
 *     subscribe(sessionId);
 *   });
 *
 *   {@html `Progress: ${$bosProgress.percentComplete}%`}
 */

import { writable, derived } from 'svelte/store';
import type { Readable, Writable } from 'svelte/store';

/**
 * BOS Progress Event Type
 */
export enum BOSEventType {
	DiscoveryStarted = 'discovery_started',
	DiscoveryProgress = 'discovery_progress',
	ConformanceStarted = 'conformance_started',
	ConformanceProgress = 'conformance_progress',
	ProcessingComplete = 'processing_complete',
	PartialResults = 'partial_results',
	ErrorRecoverable = 'error_recoverable',
	ErrorFatal = 'error_fatal',
	Metrics = 'metrics',
	Heartbeat = 'heartbeat',
	Connected = 'connected'
}

/**
 * Progress update during discovery/conformance
 */
export interface BOSProgress {
	eventsProcessed: number;
	totalEvents?: number;
	percentComplete: number;
	currentStep: string;
	activeWorkers: number;
	throughputEps: number;
}

/**
 * Metrics snapshot
 */
export interface BOSMetrics {
	elapsedSecs: number;
	totalProcessed: number;
	avgThroughputEps: number;
	currentThroughputEps: number;
	peakThroughputEps: number;
	variantsFound: number;
	violationsFound: number;
}

/**
 * Error information
 */
export interface BOSError {
	code: string;
	message: string;
	recoverable: boolean;
	retryAttempt?: number;
	maxRetries?: number;
	details?: string;
}

/**
 * Partial result event
 */
export interface BOSPartialResult {
	resultType: string; // "top_variants", "traces_sample", "bottlenecks"
	data: Record<string, unknown>;
	itemsCount: number;
	isFinal: boolean;
}

/**
 * BOS streaming event received from server
 */
export interface BOSStreamEvent {
	id: string;
	eventType: string;
	sessionId: string;
	progress?: BOSProgress;
	metrics?: BOSMetrics;
	error?: BOSError;
	partialResult?: BOSPartialResult;
	timestampMs: number;
	estimatedRemainingSecs?: number;
}

/**
 * Store state for a BOS session
 */
export interface BOSProgressState {
	sessionId: string;
	isConnected: boolean;
	isProcessing: boolean;
	phase: 'idle' | 'discovery' | 'conformance' | 'complete';
	progress?: BOSProgress;
	metrics?: BOSMetrics;
	error?: BOSError;
	partialResults: BOSPartialResult[];
	lastEventTime: number;
	connectionStartTime?: number;
	reconnectCount: number;
	statusMessage: string;
}

/**
 * Create initial state
 */
function createInitialState(sessionId: string): BOSProgressState {
	return {
		sessionId,
		isConnected: false,
		isProcessing: false,
		phase: 'idle',
		partialResults: [],
		lastEventTime: Date.now(),
		reconnectCount: 0,
		statusMessage: 'Initializing...'
	};
}

/**
 * Create BOS progress store
 */
function createBOSProgressStore() {
	const { subscribe, set, update } = writable<BOSProgressState>(createInitialState(''));

	let eventSource: EventSource | null = null;
	let reconnectTimer: NodeJS.Timeout | null = null;
	const MAX_RECONNECT_ATTEMPTS = 5;
	const RECONNECT_DELAY_MS = 3000;

	/**
	 * Subscribe to SSE stream for a session
	 */
	function connectToSession(sessionId: string, token?: string): Promise<void> {
		return new Promise((resolve, reject) => {
			// Close existing connection
			if (eventSource) {
				eventSource.close();
				eventSource = null;
			}

			if (reconnectTimer) {
				clearTimeout(reconnectTimer);
				reconnectTimer = null;
			}

			// Build URL
			const url = `/api/bos/stream/discover/${sessionId}`;

			// Create headers
			const headers: Record<string, string> = {};
			if (token) {
				headers['Authorization'] = `Bearer ${token}`;
			}

			try {
				eventSource = new EventSource(url);

				// Handle connected event
				eventSource.addEventListener('connected', (event: Event) => {
					const evt = event as MessageEvent;
					const data = JSON.parse(evt.data);

					update((state) => ({
						...state,
						sessionId: data.session_id || sessionId,
						isConnected: true,
						isProcessing: true,
						connectionStartTime: Date.now(),
						statusMessage: 'Connected, starting discovery...'
					}));

					resolve();
				});

				// Handle discovery progress
				eventSource.addEventListener('discovery_started', (event: Event) => {
					const evt = event as MessageEvent;
					const streamEvent = JSON.parse(evt.data) as BOSStreamEvent;

					update((state) => ({
						...state,
						phase: 'discovery',
						isProcessing: true,
						lastEventTime: Date.now(),
						statusMessage: 'Starting discovery phase...'
					}));
				});

				eventSource.addEventListener('discovery_progress', (event: Event) => {
					const evt = event as MessageEvent;
					const streamEvent = JSON.parse(evt.data) as BOSStreamEvent;

					if (streamEvent.progress) {
						update((state) => ({
							...state,
							progress: streamEvent.progress,
							lastEventTime: Date.now(),
							statusMessage: `${streamEvent.progress.currentStep} (${streamEvent.progress.percentComplete}%)`
						}));
					}
				});

				// Handle conformance phase
				eventSource.addEventListener('conformance_started', (event: Event) => {
					update((state) => ({
						...state,
						phase: 'conformance',
						statusMessage: 'Starting conformance phase...'
					}));
				});

				eventSource.addEventListener('conformance_progress', (event: Event) => {
					const evt = event as MessageEvent;
					const streamEvent = JSON.parse(evt.data) as BOSStreamEvent;

					if (streamEvent.progress) {
						update((state) => ({
							...state,
							progress: streamEvent.progress,
							lastEventTime: Date.now(),
							statusMessage: `Conformance: ${streamEvent.progress.currentStep}`
						}));
					}
				});

				// Handle metrics updates
				eventSource.addEventListener('metrics', (event: Event) => {
					const evt = event as MessageEvent;
					const streamEvent = JSON.parse(evt.data) as BOSStreamEvent;

					if (streamEvent.metrics) {
						update((state) => ({
							...state,
							metrics: streamEvent.metrics,
							lastEventTime: Date.now()
						}));
					}
				});

				// Handle partial results
				eventSource.addEventListener('partial_results', (event: Event) => {
					const evt = event as MessageEvent;
					const streamEvent = JSON.parse(evt.data) as BOSStreamEvent;

					if (streamEvent.partialResult) {
						update((state) => {
							const results = [...state.partialResults];
							// Keep last 10 partial results
							if (results.length >= 10) {
								results.shift();
							}
							results.push(streamEvent.partialResult);

							return {
								...state,
								partialResults: results,
								lastEventTime: Date.now()
							};
						});
					}
				});

				// Handle errors
				eventSource.addEventListener('error_recoverable', (event: Event) => {
					const evt = event as MessageEvent;
					const streamEvent = JSON.parse(evt.data) as BOSStreamEvent;

					if (streamEvent.error) {
						update((state) => ({
							...state,
							error: streamEvent.error,
							lastEventTime: Date.now(),
							statusMessage: `Error (recoverable): ${streamEvent.error.message}`
						}));
					}
				});

				eventSource.addEventListener('error_fatal', (event: Event) => {
					const evt = event as MessageEvent;
					const streamEvent = JSON.parse(evt.data) as BOSStreamEvent;

					if (streamEvent.error) {
						update((state) => ({
							...state,
							error: streamEvent.error,
							isProcessing: false,
							lastEventTime: Date.now(),
							statusMessage: `Fatal Error: ${streamEvent.error.message}`
						}));
					}
				});

				// Handle processing complete
				eventSource.addEventListener('processing_complete', (event: Event) => {
					update((state) => ({
						...state,
						phase: 'complete',
						isProcessing: false,
						lastEventTime: Date.now(),
						statusMessage: 'Processing complete!'
					}));

					// Auto-close connection on completion
					setTimeout(() => {
						disconnect();
					}, 5000);
				});

				// Handle heartbeat
				eventSource.addEventListener('heartbeat', () => {
					update((state) => ({
						...state,
						lastEventTime: Date.now()
					}));
				});

				// Handle SSE errors
				eventSource.onerror = () => {
					update((state) => ({
						...state,
						isConnected: false,
						statusMessage: 'Connection error, attempting reconnect...'
					}));

					reconnectWithBackoff(sessionId, token);
				};
			} catch (error) {
				reject(error);
			}
		});
	}

	/**
	 * Reconnect with exponential backoff
	 */
	function reconnectWithBackoff(sessionId: string, token?: string) {
		update((state) => {
			if (state.reconnectCount >= MAX_RECONNECT_ATTEMPTS) {
				return {
					...state,
					isConnected: false,
					statusMessage: 'Failed to reconnect after multiple attempts'
				};
			}

			const delay = RECONNECT_DELAY_MS * Math.pow(2, state.reconnectCount);
			reconnectTimer = setTimeout(() => {
				connectToSession(sessionId, token).catch((err) => {
					console.error('Reconnection failed:', err);
				});
			}, delay);

			return {
				...state,
				reconnectCount: state.reconnectCount + 1
			};
		});
	}

	/**
	 * Disconnect from current session
	 */
	function disconnect() {
		if (eventSource) {
			eventSource.close();
			eventSource = null;
		}

		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
			reconnectTimer = null;
		}

		update((state) => ({
			...state,
			isConnected: false,
			isProcessing: false,
			statusMessage: 'Disconnected'
		}));
	}

	/**
	 * Reset store state
	 */
	function reset(sessionId?: string) {
		disconnect();
		set(createInitialState(sessionId || ''));
	}

	/**
	 * Get current progress percentage
	 */
	const percentComplete = derived(
		{ subscribe },
		(state) => state.progress?.percentComplete ?? 0
	);

	/**
	 * Get formatted time remaining
	 */
	const timeRemaining = derived(
		{ subscribe },
		(state) => {
			const secs = state.progress?.estimatedRemainingSecs;
			if (!secs) return 'Unknown';

			const hours = Math.floor(secs / 3600);
			const mins = Math.floor((secs % 3600) / 60);
			const s = secs % 60;

			if (hours > 0) {
				return `${hours}h ${mins}m`;
			} else if (mins > 0) {
				return `${mins}m ${s}s`;
			} else {
				return `${s}s`;
			}
		}
	);

	/**
	 * Get status message with metrics
	 */
	const statusWithMetrics = derived(
		{ subscribe },
		(state) => {
			let msg = state.statusMessage;

			if (state.metrics) {
				msg += ` | ${state.metrics.currentThroughputEps.toFixed(0)} eps`;
			}

			return msg;
		}
	);

	return {
		subscribe,
		connect: connectToSession,
		disconnect,
		reset,
		percentComplete,
		timeRemaining,
		statusWithMetrics
	};
}

export const bosProgress = createBOSProgressStore();

/**
 * Hook for easy component integration
 */
export function useBOSProgress(sessionId: string, token?: string) {
	return {
		store: bosProgress,
		connect: () => bosProgress.connect(sessionId, token),
		disconnect: () => bosProgress.disconnect(),
		reset: () => bosProgress.reset(sessionId)
	};
}
