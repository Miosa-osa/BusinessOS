/**
 * Active Listening Service - Deepgram Edition
 *
 * Provides real-time audio transcription for voice commands in 3D Desktop.
 * Uses Deepgram's WebSocket API for sub-300ms latency transcription.
 *
 * Features:
 * - Real-time streaming transcription (<300ms latency)
 * - Direct WebSocket connection to Deepgram
 * - No backend processing needed
 * - Automatic punctuation and formatting
 * - Interim and final results
 * - Proper resource cleanup
 */

import { browser } from '$app/environment';
import { get } from 'svelte/store';
import { desktop3dPermissions, microphoneStream } from './desktop3dPermissions';
import { createClient, LiveTranscriptionEvents, type LiveClient } from '@deepgram/sdk';

export type TranscriptCallback = (text: string, isFinal: boolean) => void;

export class ActiveListeningService {
	private static instance: ActiveListeningService | null = null;

	private deepgramClient: any = null;
	private liveConnection: any = null;
	private mediaRecorder: MediaRecorder | null = null;
	private isListening = false;
	private transcriptCallback: TranscriptCallback | null = null;

	// Track current session
	private sessionId: string | null = null;

	private constructor() {}

	static getInstance(): ActiveListeningService {
		if (!ActiveListeningService.instance) {
			ActiveListeningService.instance = new ActiveListeningService();
		}
		return ActiveListeningService.instance;
	}

	/**
	 * Check if active listening is supported
	 */
	isSupported(): boolean {
		if (!browser) return false;
		return !!(
			window.MediaRecorder &&
			MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
		);
	}

	/**
	 * Start listening for voice input using Deepgram
	 */
	async startListening(onTranscript: TranscriptCallback): Promise<boolean> {
		console.log('[ActiveListening] 🚀 startListening() called');

		if (!this.isSupported()) {
			console.error('[ActiveListening] ❌ MediaRecorder not supported');
			return false;
		}

		if (this.isListening) {
			console.warn('[ActiveListening] ⚠️ Already listening');
			return false;
		}

		// Get microphone stream from permissions service
		console.log('[ActiveListening] 🎤 Getting microphone stream from permissions service...');
		const stream = get(microphoneStream);

		if (!stream) {
			console.error('[ActiveListening] ❌ No microphone stream available');
			console.error('[ActiveListening] ❌ Did you request microphone permission first?');
			return false;
		}

		console.log('[ActiveListening] ✅ Microphone stream retrieved:', {
			id: stream.id,
			active: stream.active,
			tracks: stream.getTracks().length,
			audioTracks: stream.getAudioTracks().length
		});

		// Verify stream has active audio tracks
		const audioTracks = stream.getAudioTracks();
		if (audioTracks.length === 0) {
			console.error('[ActiveListening] ❌ No audio tracks in microphone stream');
			return false;
		}

		if (!audioTracks[0].enabled) {
			console.error('[ActiveListening] ❌ Audio track is not enabled');
			return false;
		}

		console.log('[ActiveListening] ✅ Audio track valid:', {
			label: audioTracks[0].label,
			enabled: audioTracks[0].enabled,
			muted: audioTracks[0].muted,
			readyState: audioTracks[0].readyState
		});

		try {
			// Get Deepgram API key from environment
			console.log('[ActiveListening] 🔑 Checking for Deepgram API key...');
			const apiKey = import.meta.env.VITE_DEEPGRAM_API_KEY;

			if (!apiKey) {
				console.error('[ActiveListening] ❌ VITE_DEEPGRAM_API_KEY not found in environment');
				alert('Deepgram API key not configured. Please add VITE_DEEPGRAM_API_KEY to your .env file.');
				return false;
			}

			console.log('[ActiveListening] ✅ Deepgram API key found (length:', apiKey.length, ')');
			console.log('[ActiveListening] 🔑 Initializing Deepgram client...');

			// Create Deepgram client
			this.deepgramClient = createClient(apiKey);

			// Open live transcription connection
			// IMPORTANT: Must specify mime_type for WebM/Opus audio from MediaRecorder
			this.liveConnection = this.deepgramClient.listen.live({
				model: 'nova-2', // Use nova-2 for better WebM support
				language: 'en-US',
				punctuate: true,
				interim_results: true,
				endpointing: 300,
				utterance_end_ms: 1000,
				vad_events: true
			});

			this.transcriptCallback = onTranscript;
			this.sessionId = crypto.randomUUID();

			console.log('[ActiveListening] 🌐 Connecting to Deepgram WebSocket...');

			// Add timeout warning if connection doesn't open
			const connectionTimeout = setTimeout(() => {
				console.warn('[ActiveListening] ⚠️ Deepgram WebSocket taking longer than expected to connect...');
			}, 3000);

			// Handle connection opened
			this.liveConnection.on(LiveTranscriptionEvents.Open, () => {
				clearTimeout(connectionTimeout);
				websocketReady = true;
				console.log('[ActiveListening] ✅ Deepgram WebSocket connected!');
				console.log('[ActiveListening] 🎤 Ready to receive audio - speak now!');
			});

			// Handle transcription results
			this.liveConnection.on(LiveTranscriptionEvents.Transcript, (data: any) => {
				console.log('[ActiveListening] 🎯 RAW Deepgram response:', JSON.stringify(data).substring(0, 200));

				try {
					const transcript = data.channel?.alternatives?.[0]?.transcript || '';
					const isFinal = data.is_final || false;
					const speechFinal = data.speech_final || false;
					const confidence = data.channel?.alternatives?.[0]?.confidence || 0;

					console.log('[ActiveListening] 📝 Transcription received:', {
						transcript,
						isFinal,
						speechFinal,
						confidence
					});

					if (transcript && transcript.trim().length > 0) {
						// Call transcript callback
						this.transcriptCallback?.(transcript.trim(), speechFinal || isFinal);

						if (speechFinal || isFinal) {
							console.log('[ActiveListening] ✅ Final transcript:', transcript.trim());
						}
					} else {
						console.warn('[ActiveListening] ⚠️ Empty transcript received');
					}
				} catch (err) {
					console.error('[ActiveListening] ❌ Error processing transcription:', err);
				}
			});

			// Handle metadata
			this.liveConnection.on(LiveTranscriptionEvents.Metadata, (data: any) => {
				console.log('[ActiveListening] 📊 Metadata:', data);
			});

			// Handle errors
			this.liveConnection.on(LiveTranscriptionEvents.Error, (error: any) => {
				console.error('[ActiveListening] ❌ Deepgram error:', error);
			});

			// Handle connection closed
			this.liveConnection.on(LiveTranscriptionEvents.Close, () => {
				console.log('[ActiveListening] 🔌 Deepgram connection closed');
			});

			// Handle warnings (helpful for debugging)
			this.liveConnection.on(LiveTranscriptionEvents.Warning, (warning: any) => {
				console.warn('[ActiveListening] ⚠️ Deepgram warning:', warning);
			});

			// Handle utterance end events
			this.liveConnection.on(LiveTranscriptionEvents.UtteranceEnd, (data: any) => {
				console.log('[ActiveListening] 🔚 Utterance ended:', data);
			});

			// Create MediaRecorder to capture audio
			this.mediaRecorder = new MediaRecorder(stream, {
				mimeType: 'audio/webm;codecs=opus'
			});

			// Track if WebSocket is ready
			let websocketReady = false;

			// Send audio data to Deepgram as it's recorded
			this.mediaRecorder.ondataavailable = (event) => {
				if (event.data.size > 0) {
					// Only send if WebSocket is ready
					if (!websocketReady) {
						console.warn('[ActiveListening] ⚠️ WebSocket not ready yet, buffering audio...');
						return;
					}

					console.log('[ActiveListening] 📤 Sending audio chunk:', {
						size: event.data.size,
						type: event.data.type
					});

					// Try to send to Deepgram
					try {
						if (this.liveConnection) {
							this.liveConnection.send(event.data);
							console.log('[ActiveListening] ✅ Sent to Deepgram');
						} else {
							console.error('[ActiveListening] ❌ No liveConnection available');
						}
					} catch (err) {
						console.error('[ActiveListening] ❌ Error sending to Deepgram:', err);
					}
				}
			};

			// Handle recorder errors
			this.mediaRecorder.onerror = (event) => {
				console.error('[ActiveListening] MediaRecorder error:', event);
				this.stopListening();
			};

			// Handle recorder start
			this.mediaRecorder.onstart = () => {
				console.log('[ActiveListening] 🎙️ MediaRecorder started - capturing audio now');
			};

			// Start recording in small chunks for low latency (250ms chunks)
			console.log('[ActiveListening] 📹 Starting MediaRecorder with 250ms chunks...');
			this.mediaRecorder.start(250);
			this.isListening = true;

			console.log('[ActiveListening] ✅ Started listening with Deepgram', {
				sessionId: this.sessionId,
				model: 'nova-3',
				language: 'en-US',
				mimeType: 'audio/webm;codecs=opus'
			});

			return true;
		} catch (err) {
			console.error('[ActiveListening] Failed to start listening:', err);
			return false;
		}
	}

	/**
	 * Stop listening
	 */
	stopListening(): void {
		if (!this.isListening) {
			return;
		}

		console.log('[ActiveListening] 🛑 Stopping listening...');

		// Stop MediaRecorder
		if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
			this.mediaRecorder.stop();
		}

		// Close Deepgram connection
		if (this.liveConnection) {
			try {
				this.liveConnection.finish();
				console.log('[ActiveListening] ✅ Deepgram connection finished');
			} catch (err) {
				console.error('[ActiveListening] Error closing Deepgram connection:', err);
			}
		}

		this.mediaRecorder = null;
		this.liveConnection = null;
		this.deepgramClient = null;
		this.transcriptCallback = null;
		this.isListening = false;
		this.sessionId = null;

		console.log('[ActiveListening] ✅ Stopped listening');
	}

	/**
	 * Check if currently listening
	 */
	getIsListening(): boolean {
		return this.isListening;
	}

	/**
	 * Get current session ID
	 */
	getSessionId(): string | null {
		return this.sessionId;
	}
}

// Export singleton instance
export const activeListeningService = ActiveListeningService.getInstance();
