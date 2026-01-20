import { Room, RoomEvent, createLocalAudioTrack } from 'livekit-client';

export type VoiceState =
	| 'disconnected'
	| 'connecting'
	| 'connected'
	| 'transcribing'
	| 'speaking'
	| 'error';

interface Callbacks {
	onStateChange?: (state: VoiceState) => void;
	onUserTranscript?: (text: string) => void;
	onAgentResponse?: (text: string) => void;
	onNavigationCommand?: (module: string) => void;
	onError?: (error: string) => void;
}

/**
 * Get API base URL from environment
 */
function getApiUrl(): string {
	return import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8001' : '');
}

/**
 * LiveKit Voice Service
 * Replaces the old streamingVoice service with LiveKit WebRTC integration
 * Works with the existing VoiceOrbPanel and LiveCaptions components
 */
class LiveKitVoiceService {
	private room: Room | null = null;
	private callbacks: Callbacks = {};
	private currentState: VoiceState = 'disconnected';

	/**
	 * Set callback for state changes
	 */
	setStateCallback(cb: (state: VoiceState) => void) {
		this.callbacks.onStateChange = cb;
	}

	/**
	 * Set callback for user transcript updates
	 */
	setUserCallback(cb: (text: string) => void) {
		this.callbacks.onUserTranscript = cb;
	}

	/**
	 * Set callback for agent response updates
	 */
	setAgentCallback(cb: (text: string) => void) {
		this.callbacks.onAgentResponse = cb;
	}

	/**
	 * Set callback for navigation commands
	 */
	setNavigationCallback(cb: (module: string) => void) {
		this.callbacks.onNavigationCommand = cb;
	}

	/**
	 * Set callback for errors
	 */
	setErrorCallback(cb: (error: string) => void) {
		this.callbacks.onError = cb;
	}

	/**
	 * Connect to LiveKit room and start voice conversation
	 */
	async connect(): Promise<void> {
		try {
			this.updateState('connecting');

			// Get LiveKit token from backend
			const apiUrl = getApiUrl();
			const tokenUrl = `${apiUrl}/api/livekit/token`;
			console.log('[LiveKit] Fetching token from:', tokenUrl);

			const response = await fetch(tokenUrl, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({ agent_role: 'groq-agent' })
			});

			if (!response.ok) {
				const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
				throw new Error(errorData.error || 'Failed to get LiveKit token');
			}

			const { token, url } = await response.json();

			// Create LiveKit room
			this.room = new Room();

			// Setup event listeners - Handle transcript messages from Python agent
			this.room.on(RoomEvent.DataReceived, (payload: Uint8Array) => {
				try {
					const data = JSON.parse(new TextDecoder().decode(payload));

					// Handle transcript messages (from Python agent)
					if (data.type === 'transcript') {
						const source = data.source || 'unknown';
						const role = data.role || 'unknown';
						const text = data.text || '';

						// Log exactly like reference implementation
						console.log(`[${source.toUpperCase()}] ${role}: ${text}`);

						// Update UI based on role
						if (role === 'user') {
							this.callbacks.onUserTranscript?.(text);
							this.updateState('transcribing');
						} else if (role === 'agent') {
							this.callbacks.onAgentResponse?.(text);
							this.updateState('speaking');
						}

					// Parse navigation commands from agent response (multiple patterns)
					// Matches: "Opened X module", "X is now open", "opening X", "let me open X"
					const navPatterns = [
						/Opened (\w+) module/i,                            // "Opened tasks module"
						/(?:the )?(\w+) module is now open/i,              // "the pages module is now open"
						/(?:the )?(\w+) module (?:has been|is) opened/i,   // "the pages module has been opened"
						/taking you to (?:the )?(\w+)/i,                   // "taking you to tasks"
						/opening (?:the )?(\w+)(?: for you)?/i,            // "opening tasks for you"
						/let me open (?:the )?(\w+)/i,                     // "let me open tasks"
						/navigating to (?:the )?(\w+)/i,                   // "navigating to tasks"
						/showing you (?:the )?(\w+)/i,                     // "showing you tasks"
						/pulling up (?:the )?(\w+)/i,                      // "pulling up tasks"
						/(?:the )?(\w+) is now open/i,                     // "tasks is now open"
						/(?:the )?(\w+) (?:has been|is) opened/i,          // "the terminal has been opened"
						/opened (?:the )?(\w+)/i                           // "opened tasks"
					];

					for (const pattern of navPatterns) {
						const navMatch = text.match(pattern);
						if (navMatch && navMatch[1]) {
							const moduleName = navMatch[1].toLowerCase();
							console.log('[LiveKit] 🧭 Navigation command detected:', moduleName, 'from pattern:', pattern);
							this.callbacks.onNavigationCommand?.(moduleName);
							break; // Only trigger once per response
						}
					}
					}
				} catch (error) {
					console.error('[LiveKit] Error parsing data:', error);
				}
			});

			this.room.on(RoomEvent.TrackSubscribed, (track, publication, participant) => {
				console.log('[LiveKit] Track subscribed:', track.kind, 'from', participant?.identity);

				if (track.kind === 'audio') {
					console.log('[LiveKit] Agent audio track subscribed');
					this.updateState('speaking');

					// Attach audio track to an audio element (matches reference implementation)
					const audioElement = track.attach();
					audioElement.autoplay = true;
					audioElement.volume = 1.0;
					audioElement.muted = false; // Ensure not muted

					// Add to body for playback
					document.body.appendChild(audioElement);

					console.log('[LiveKit] Audio element state:', {
						paused: audioElement.paused,
						muted: audioElement.muted,
						volume: audioElement.volume,
						readyState: audioElement.readyState,
						src: audioElement.src
					});

					// Force play to bypass autoplay restrictions
					audioElement.play().then(() => {
						console.log('[LiveKit] ✅ Audio playing successfully!');
						console.log('[LiveKit] After play - paused:', audioElement.paused, 'currentTime:', audioElement.currentTime);
					}).catch((err) => {
						console.error('[LiveKit] ❌ Audio play failed:', err);
						console.error('[LiveKit] This is likely a browser autoplay policy restriction');
					});

					// Listen for audio events
					audioElement.addEventListener('playing', () => {
						console.log('[LiveKit] 🔊 Audio actually playing now!');
					});
					audioElement.addEventListener('ended', () => {
						console.log('[LiveKit] Audio playback ended');
					});
					audioElement.addEventListener('error', (e) => {
						console.error('[LiveKit] Audio error:', e);
					});
				}
			});

			this.room.on(RoomEvent.TrackUnsubscribed, () => {
				console.log('[LiveKit] Audio track unsubscribed');
				this.updateState('connected');
			});

			this.room.on(RoomEvent.Disconnected, () => {
				console.log('[LiveKit] Room disconnected');
				this.updateState('disconnected');
			});

			this.room.on(RoomEvent.ConnectionStateChanged, (state) => {
				console.log('[LiveKit] Connection state:', state);
			});

			// Connect to room
			await this.room.connect(url, token);
			console.log('[LiveKit] Connected to room successfully', {
				roomName: this.room.name,
				numParticipants: this.room.remoteParticipants.size,
				identity: this.room.localParticipant.identity
			});

			// Create and publish microphone track
			const audioTrack = await createLocalAudioTrack({
				echoCancellation: true,
				noiseSuppression: true,
				autoGainControl: true
			});

			await this.room.localParticipant.publishTrack(audioTrack);
			console.log('[LiveKit] Published audio track');

			this.updateState('connected');
		} catch (error) {
			console.error('[LiveKit] Connection error:', error);
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			this.callbacks.onError?.(errorMessage);
			this.updateState('error');
			throw error;
		}
	}

	/**
	 * Disconnect from LiveKit room
	 */
	async disconnect(): Promise<void> {
		console.trace('[LiveKit] disconnect() called - STACK TRACE:');
		if (this.room) {
			console.log('[LiveKit] Disconnecting from room');
			await this.room.disconnect();
			this.room = null;
		}
		this.updateState('disconnected');
	}

	/**
	 * Get current connection state
	 */
	getState(): VoiceState {
		return this.currentState;
	}

	/**
	 * Check if currently connected
	 */
	isConnected(): boolean {
		return this.currentState !== 'disconnected' && this.currentState !== 'error';
	}

	/**
	 * Update internal state and notify callback
	 */
	private updateState(state: VoiceState) {
		this.currentState = state;
		this.callbacks.onStateChange?.(state);
	}
}

// Export singleton instance
export const streamingVoice = new LiveKitVoiceService();
