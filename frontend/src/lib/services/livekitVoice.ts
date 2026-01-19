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
	onError?: (error: string) => void;
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
			const response = await fetch('/api/livekit/token', {
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
					document.body.appendChild(audioElement);

					// Force play to bypass autoplay restrictions
					audioElement.play().then(() => {
						console.log('[LiveKit] ✅ Audio playing successfully!');
					}).catch((err) => {
						console.error('[LiveKit] ❌ Audio play failed:', err);
						console.error('[LiveKit] This is likely a browser autoplay policy restriction');
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
