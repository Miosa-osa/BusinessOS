/**
 * Voice Transcription Service
 *
 * Handles real-time speech-to-text transcription using Deepgram's Nova-2 model.
 * Features:
 * - WebSocket-based streaming transcription
 * - Keyword boosting for accurate recognition of "OSA" and "BusinessOS"
 * - Interim and final results for responsive UI
 */

import { createClient, LiveTranscriptionEvents } from '@deepgram/sdk';
import { get } from 'svelte/store';
import { microphoneStream } from './desktop3dPermissions';

export type TranscriptCallback = (text: string, isFinal: boolean) => void;

class VoiceTranscriptionService {
	private deepgram: any = null;
	private connection: any = null;
	private recorder: MediaRecorder | null = null;
	private callback: TranscriptCallback | null = null;
	private isActive = false;

	async start(onTranscript: TranscriptCallback): Promise<boolean> {
		if (this.isActive) return false;

		// Get mic stream
		const stream = get(microphoneStream);
		if (!stream) {
			console.error('[Voice] No microphone');
			return false;
		}

		try {
			// Get API key
			const key = import.meta.env.VITE_DEEPGRAM_API_KEY;
			if (!key) {
				console.error('[Voice] No API key');
				return false;
			}

			console.log('[Voice] Starting...');
			this.callback = onTranscript;

			// Connect to Deepgram
			this.deepgram = createClient(key);
			this.connection = this.deepgram.listen.live({
				model: 'nova-2',
				language: 'en-US',
				punctuate: true,
				interim_results: true,
				keywords: ['OSA:2', 'BusinessOS:1.5']  // Boost recognition of "OSA" and "BusinessOS"
			});

			// Handle transcriptions
			this.connection.on(LiveTranscriptionEvents.Transcript, (data: any) => {
				const text = data.channel?.alternatives?.[0]?.transcript || '';
				if (text.trim()) {
					const isFinal = data.is_final || data.speech_final || false;
					this.callback?.(text.trim(), isFinal);
				}
			});

			// Handle errors
			this.connection.on(LiveTranscriptionEvents.Error, (err: any) => {
				console.error('[Voice] Deepgram error:', err);
			});

			// Wait for connection
			await new Promise<void>((resolve) => {
				this.connection.on(LiveTranscriptionEvents.Open, () => {
					console.log('[Voice] Connected');
					resolve();
				});
			});

			// Start recording
			this.recorder = new MediaRecorder(stream, {
				mimeType: 'audio/webm;codecs=opus'
			});

			this.recorder.ondataavailable = (event) => {
				if (event.data.size > 0 && this.connection) {
					this.connection.send(event.data);
				}
			};

			this.recorder.start(250);
			this.isActive = true;

			console.log('[Voice] Ready');
			return true;
		} catch (err) {
			console.error('[Voice] Start failed:', err);
			this.stop();
			return false;
		}
	}

	stop() {
		console.log('[Voice] Stopping');

		if (this.recorder) {
			this.recorder.stop();
			this.recorder = null;
		}

		if (this.connection) {
			this.connection.finish();
			this.connection = null;
		}

		this.deepgram = null;
		this.callback = null;
		this.isActive = false;
	}

	isListening(): boolean {
		return this.isActive;
	}
}

export const voiceTranscription = new VoiceTranscriptionService();
