/**
 * LiveKit Native Voice Service
 *
 * For use with Python LiveKit voice agent (python-voice-agent/)
 * The Python agent handles ALL voice processing (STT, LLM, TTS)
 * Frontend just connects to LiveKit room - everything else is automatic!
 *
 * Tech Stack:
 * - LiveKit: WebRTC audio transport (bidirectional)
 * - Python Agent: Handles STT (Groq Whisper), LLM (Groq Llama), TTS (ElevenLabs)
 * - This file: Just manages connection state
 */

import { Room, RoomEvent, Track } from 'livekit-client';

export type VoiceState = 'disconnected' | 'connecting' | 'connected' | 'speaking';

class SimpleVoiceService {
	private room: Room | null = null;
	private state: VoiceState = 'disconnected';

	// Callbacks
	private onStateChange: ((state: VoiceState) => void) | null = null;
	private onUserMessage: ((text: string) => void) | null = null;
	private onAgentMessage: ((text: string) => void) | null = null;

	/**
	 * Connect to LiveKit voice room
	 * Python agent will join automatically and handle all voice processing
	 */
	async connect() {
		if (this.room) {
			console.log('[Voice] Already connected');
			return;
		}

		this.setState('connecting');
		console.log('[Voice] 🎤 Connecting to LiveKit...');

		try {
			// Get token from Go backend
			const response = await fetch('http://localhost:8001/api/livekit/token', {
				method: 'POST',
				credentials: 'include',
			});

			if (!response.ok) {
				throw new Error(`Failed to get token: ${response.status}`);
			}

			const { token, url, room_name, identity } = await response.json();
			console.log('[Voice] ✅ Got LiveKit token');
			console.log('[Voice] Room:', room_name);
			console.log('[Voice] Identity:', identity);
			console.log('[Voice] URL:', url);

			// Connect to LiveKit room
			this.room = new Room();

			// Listen for participants (the Python voice agent will join)
			this.room.on(RoomEvent.ParticipantConnected, (participant) => {
				console.log('[Voice] 🤖 Participant connected:', participant.identity);
				if (participant.identity.includes('agent')) {
					console.log('[Voice] ✅ Voice agent joined the room!');
				}
			});

			// Listen for audio tracks (agent's voice responses)
			this.room.on(RoomEvent.TrackSubscribed, (track, publication, participant) => {
				console.log('[Voice] 🔊 Track subscribed:', track.kind, 'from', participant.identity);

				if (track.kind === Track.Kind.Audio) {
					// Agent is speaking
					this.setState('speaking');
					console.log('[Voice] 🗣️ Agent is speaking...');

					// Attach to an audio element for playback
					const audioElement = track.attach();
					document.body.appendChild(audioElement);
					audioElement.play().catch((e) => {
						console.error('[Voice] Audio playback failed:', e);
					});

					// When track ends, remove element
					track.on('ended', () => {
						console.log('[Voice] 🔇 Agent finished speaking');
						audioElement.remove();
						this.setState('connected');
					});
				}
			});

			// Listen for track unsubscribed
			this.room.on(RoomEvent.TrackUnsubscribed, (track, publication, participant) => {
				console.log('[Voice] Track unsubscribed:', track.kind);
			});

			// Listen for disconnection
			this.room.on(RoomEvent.Disconnected, () => {
				console.log('[Voice] ❌ Disconnected from room');
				this.setState('disconnected');
			});

			// Listen for data messages (transcripts from Python agent)
			this.room.on(RoomEvent.DataReceived, (payload, participant) => {
				try {
					const text = new TextDecoder().decode(payload);
					const data = JSON.parse(text);
					console.log('[Voice] Data received:', data);

					if (data.type === 'user_transcript') {
						console.log('[Voice] User said:', data.text);
						this.notifyUserMessage(data.text);
					} else if (data.type === 'agent_transcript') {
						console.log('[Voice] Agent said:', data.text);
						this.notifyAgentMessage(data.text);
					}
				} catch (e) {
					console.log('[Voice] Non-JSON data received');
				}
			});

			// Connect to the room
			await this.room.connect(url, token);
			console.log('[Voice] ✅ Connected to LiveKit room');

			// Enable microphone (user speaks to the agent)
			await this.room.localParticipant.setMicrophoneEnabled(true);
			console.log('[Voice] 🎤 Microphone enabled');

			this.setState('connected');
			console.log('[Voice] ✅ Ready! Python agent will join automatically.');
			console.log('[Voice] 💡 Just speak - the agent will respond via LiveKit WebRTC');

		} catch (error) {
			console.error('[Voice] ❌ Connection failed:', error);
			this.setState('disconnected');
			throw error;
		}
	}

	/**
	 * Disconnect from room
	 */
	async disconnect() {
		console.log('[Voice] 🔌 Disconnecting...');

		if (this.room) {
			this.room.disconnect();
			this.room = null;
		}

		this.setState('disconnected');
		console.log('[Voice] ✅ Disconnected');
	}

	/**
	 * Set state and notify callback
	 */
	private setState(state: VoiceState) {
		this.state = state;
		console.log('[Voice] State changed:', state);

		if (this.onStateChange) {
			this.onStateChange(state);
		}
	}

	/**
	 * Set callback for state changes
	 */
	setStateCallback(callback: (state: VoiceState) => void) {
		this.onStateChange = callback;
	}

	/**
	 * Get current state
	 */
	getState(): VoiceState {
		return this.state;
	}

	/**
	 * Check if connected
	 */
	isConnected(): boolean {
		return this.state !== 'disconnected';
	}

	/**
	 * Set callback for user transcription (what user said)
	 * Note: Requires Python agent to send transcripts via data channel
	 */
	setUserCallback(callback: (text: string) => void) {
		this.onUserMessage = callback;
	}

	/**
	 * Set callback for agent responses (what agent said)
	 * Note: Requires Python agent to send transcripts via data channel
	 */
	setAgentCallback(callback: (text: string) => void) {
		this.onAgentMessage = callback;
	}

	/**
	 * Notify user message callback
	 */
	notifyUserMessage(text: string) {
		if (this.onUserMessage) {
			this.onUserMessage(text);
		}
	}

	/**
	 * Notify agent message callback
	 */
	notifyAgentMessage(text: string) {
		if (this.onAgentMessage) {
			this.onAgentMessage(text);
		}
	}
}

export const simpleVoice = new SimpleVoiceService();
