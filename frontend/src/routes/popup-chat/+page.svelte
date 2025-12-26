<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { apiClient } from '$lib/api';

	// Types
	interface LLMModel {
		id: string;
		name: string;
		provider: string;
		description?: string;
		size?: string;
	}

	interface PullProgress {
		status: string;
		total?: number;
		completed?: number;
	}

	// State
	let inputValue = $state('');
	let messages = $state<Array<{ role: 'user' | 'assistant'; content: string }>>([]);
	let isLoading = $state(false);
	let isRecording = $state(false);
	let isMeetingMode = $state(false);
	let inputElement: HTMLTextAreaElement | undefined = $state(undefined);
	let messagesContainer: HTMLDivElement | undefined = $state(undefined);

	// Model selection
	let availableModels = $state<LLMModel[]>([]);
	let localModels = $state<LLMModel[]>([]);
	let selectedModel = $state('');
	let activeProvider = $state('ollama_local');
	let showModelSelector = $state(false);
	let configuredProviders = $state<string[]>([]);

	// Cloud models available per provider
	const cloudModelsByProvider: Record<string, { id: string; name: string; description: string }[]> = {
		groq: [
			{ id: 'llama-3.3-70b-versatile', name: 'Llama 3.3 70B', description: 'Fast 70B' },
			{ id: 'llama-3.1-8b-instant', name: 'Llama 3.1 8B', description: 'Ultra-fast' },
			{ id: 'mixtral-8x7b-32768', name: 'Mixtral 8x7B', description: '32k context' },
		],
		anthropic: [
			{ id: 'claude-sonnet-4-20250514', name: 'Claude Sonnet 4', description: 'Best for most tasks' },
			{ id: 'claude-opus-4-20250514', name: 'Claude Opus 4', description: 'Most capable' },
		],
		ollama_cloud: [
			{ id: 'qwen3:480b', name: 'Qwen3 480B', description: 'Largest Qwen model' },
			{ id: 'qwen3:235b', name: 'Qwen3 235B', description: 'Large Qwen model' },
			{ id: 'qwen3:32b', name: 'Qwen3 32B', description: 'Qwen3 32B' },
			{ id: 'llama3.3:70b', name: 'Llama 3.3 70B', description: 'Latest Llama' },
			{ id: 'deepseek-r1:671b', name: 'DeepSeek R1 671B', description: 'Reasoning model' },
			{ id: 'deepseek-r1:70b', name: 'DeepSeek R1 70B', description: 'Compact reasoning' },
			{ id: 'command-a:111b', name: 'Command A 111B', description: 'Cohere Command' },
		]
	};

	// Model pulling state
	let isPulling = $state(false);
	let pullingModel = $state('');
	let pullProgress = $state<PullProgress | null>(null);

	// Audio recording state
	let mediaRecorder: MediaRecorder | null = null;
	let audioChunks: Blob[] = [];
	let whisperAvailable = $state(false);
	let recordingDuration = $state(0);
	let recordingInterval: ReturnType<typeof setInterval> | null = null;

	// Live transcript state (using Web Speech API)
	let liveTranscript = $state('');
	let speechRecognition: SpeechRecognition | null = null;

	// Audio visualization state
	let audioContext: AudioContext | null = null;
	let analyser: AnalyserNode | null = null;
	let audioDataArray: Uint8Array | null = null;
	let waveformBars = $state<number[]>(Array(30).fill(2));
	let animationFrameId: number | null = null;

	// Clipboard state
	let copiedMessageId = $state<number | null>(null);

	// Screenshot state
	let pendingScreenshot = $state<string | null>(null);
	let isCapturingScreenshot = $state(false);

	// Size state
	type PopupSize = 'small' | 'medium' | 'large' | 'full';
	let currentSize = $state<PopupSize>('small');

	// Format recording duration as MM:SS
	function formatDuration(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	// Upcoming meetings from calendar
	let upcomingMeeting = $state<{ id: string; title: string; start: string } | null>(null);

	onMount(() => {
		// Focus input on mount
		inputElement?.focus();

		// Listen for focus event from Electron
		if (browser && 'electron' in window) {
			const electron = (window as any).electron;

			electron?.on?.('popup:focus-input', () => {
				inputElement?.focus();
			});

			electron?.on?.('popup:start-meeting-recording', () => {
				startMeetingRecording();
			});

			electron?.on?.('popup:size-changed', (size: PopupSize) => {
				currentSize = size;
			});

			// Get initial size
			electron?.popup?.getSize?.().then((size: PopupSize) => {
				currentSize = size;
			});
		}

		// Load available models
		loadModels();

		// Check whisper status
		checkWhisperStatus();

		// Load upcoming meeting
		loadUpcomingMeeting();

		// Handle escape key to close
		const handleKeyDown = (e: KeyboardEvent) => {
			if (e.key === 'Escape') {
				hidePopup();
			}
		};
		window.addEventListener('keydown', handleKeyDown);

		return () => {
			window.removeEventListener('keydown', handleKeyDown);
		};
	});

	async function loadModels() {
		try {
			// Get provider info first
			const providersRes = await apiClient.get('/ai/providers');
			if (providersRes.ok) {
				const data = await providersRes.json();
				activeProvider = data.active_provider || 'ollama_local';
				configuredProviders = (data.providers || [])
					.filter((p: any) => p.configured)
					.map((p: any) => p.id);
				if (data.default_model && !selectedModel) {
					selectedModel = data.default_model;
				}
			}

			// Get all available models
			const response = await apiClient.get('/ai/models');
			if (response.ok) {
				const data = await response.json();
				availableModels = data.models || [];
				if (!selectedModel && availableModels.length > 0) {
					selectedModel = availableModels[0].id;
				}
			}

			// Get local models separately to know what's pulled
			const localResponse = await apiClient.get('/ai/models/local');
			if (localResponse.ok) {
				const data = await localResponse.json();
				localModels = data.models || [];
				// If using local provider and no model selected, pick first local model
				if (activeProvider === 'ollama_local' && !selectedModel && localModels.length > 0) {
					selectedModel = localModels[0].id;
				}
			}
		} catch (error) {
			console.error('Failed to load models:', error);
		}
	}

	// Check if a model is already pulled locally
	function isModelPulled(modelId: string): boolean {
		return localModels.some(m => m.id === modelId || m.id.startsWith(modelId));
	}

	// Get provider for a model
	function getModelProvider(modelId: string): string {
		const model = availableModels.find(m => m.id === modelId);
		return model?.provider || 'ollama';
	}

	// Check if model is cloud-based
	function isCloudModel(modelId: string): boolean {
		const provider = getModelProvider(modelId);
		return provider === 'groq' || provider === 'anthropic' || provider === 'ollama_cloud';
	}

	// Pull a local model
	async function pullModel(modelId: string) {
		if (isPulling) return;

		isPulling = true;
		pullingModel = modelId;
		pullProgress = { status: 'Starting...' };

		try {
			const response = await fetch('/api/ai/models/pull', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ model: modelId }),
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error('Failed to pull model');
			}

			const reader = response.body?.getReader();
			if (!reader) throw new Error('No response body');

			const decoder = new TextDecoder();
			let buffer = '';

			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				buffer += decoder.decode(value, { stream: true });
				const lines = buffer.split('\n');
				buffer = lines.pop() || '';

				for (const line of lines) {
					if (line.startsWith('data: ')) {
						try {
							const data = JSON.parse(line.slice(6));
							pullProgress = data;

							if (data.status === 'complete' || data.status === 'success') {
								selectedModel = modelId;
								showModelSelector = false;
								await loadModels();
							}
						} catch {}
					}
				}
			}
		} catch (error) {
			console.error('Pull error:', error);
			pullProgress = { status: 'Failed to pull model' };
		} finally {
			isPulling = false;
			pullingModel = '';
			setTimeout(() => pullProgress = null, 2000);
		}
	}

	// Select model - pull if needed
	function selectModel(model: LLMModel) {
		if (model.provider === 'ollama' && !isModelPulled(model.id)) {
			// Need to pull local model first
			pullModel(model.id);
		} else {
			selectedModel = model.id;
			showModelSelector = false;
		}
	}

	async function checkWhisperStatus() {
		try {
			const response = await apiClient.get('/transcribe/status');
			if (response.ok) {
				const data = await response.json();
				whisperAvailable = data.available;
			}
		} catch (error) {
			console.error('Failed to check whisper status:', error);
			whisperAvailable = false;
		}
	}

	function hidePopup() {
		if (browser && 'electron' in window) {
			(window as any).electron?.popup?.hide?.() || (window as any).electron?.send?.('popup:hide');
		}
	}

	function openMainWindow() {
		if (browser && 'electron' in window) {
			(window as any).electron?.popup?.openMain?.() || (window as any).electron?.send?.('popup:open-main');
		}
	}

	function setSize(size: PopupSize) {
		if (browser && 'electron' in window) {
			(window as any).electron?.popup?.setSize?.(size);
			currentSize = size;
		}
	}

	function cycleSize() {
		const sizes: PopupSize[] = ['small', 'medium', 'large', 'full'];
		const currentIndex = sizes.indexOf(currentSize);
		const nextIndex = (currentIndex + 1) % sizes.length;
		setSize(sizes[nextIndex]);
	}

	function expandToFull() {
		if (browser && 'electron' in window) {
			(window as any).electron?.popup?.expandToFull?.();
		}
	}

	async function loadUpcomingMeeting() {
		try {
			const response = await apiClient.get('/calendar/upcoming');
			if (response.ok) {
				const data = await response.json();
				if (data.events && data.events.length > 0) {
					const nextMeeting = data.events[0];
					const meetingStart = new Date(nextMeeting.start_time);
					const now = new Date();
					const diffMinutes = (meetingStart.getTime() - now.getTime()) / (1000 * 60);

					// Only show if meeting is within next 30 minutes
					if (diffMinutes <= 30 && diffMinutes > -60) {
						upcomingMeeting = {
							id: nextMeeting.id,
							title: nextMeeting.title,
							start: nextMeeting.start_time
						};
					}
				}
			}
		} catch (error) {
			console.error('Failed to load upcoming meeting:', error);
		}
	}

	async function handleSubmit() {
		if (!inputValue.trim() || isLoading) return;

		const userMessage = inputValue.trim();
		inputValue = '';

		// Add user message
		messages = [...messages, { role: 'user', content: userMessage }];
		scrollToBottom();

		isLoading = true;

		try {
			// Send to AI backend
			const response = await apiClient.post('/api/chat/message', {
				message: userMessage,
				model: selectedModel || undefined,
				context: isMeetingMode ? 'meeting_assistant' : 'quick_chat'
			});

			if (response.ok) {
				const data = await response.json();
				messages = [...messages, { role: 'assistant', content: data.response || data.content }];
			} else {
				const error = await response.json();
				messages = [...messages, { role: 'assistant', content: `Error: ${error.error || 'Unknown error'}` }];
			}
		} catch (error) {
			console.error('Chat error:', error);
			messages = [...messages, { role: 'assistant', content: 'Connection error. Please check your network.' }];
		} finally {
			isLoading = false;
			scrollToBottom();
		}
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSubmit();
		}
	}

	function scrollToBottom() {
		setTimeout(() => {
			messagesContainer?.scrollTo({
				top: messagesContainer.scrollHeight,
				behavior: 'smooth'
			});
		}, 100);
	}

	// Voice recording functions
	async function toggleRecording() {
		if (isRecording) {
			stopRecording();
		} else {
			await startRecording();
		}
	}

	async function startRecording() {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
			mediaRecorder = new MediaRecorder(stream);
			audioChunks = [];
			liveTranscript = '';
			recordingDuration = 0;

			// Set up audio analyzer for waveform visualization
			audioContext = new AudioContext();
			analyser = audioContext.createAnalyser();
			analyser.fftSize = 256;
			analyser.smoothingTimeConstant = 0.3;
			const source = audioContext.createMediaStreamSource(stream);
			source.connect(analyser);
			audioDataArray = new Uint8Array(analyser.fftSize);

			// Start waveform animation
			function updateWaveform() {
				if (!analyser || !audioDataArray) {
					animationFrameId = requestAnimationFrame(updateWaveform);
					return;
				}
				analyser.getByteTimeDomainData(audioDataArray as Uint8Array<ArrayBuffer>);
				const bars: number[] = [];
				const step = Math.floor(audioDataArray.length / 30);
				for (let i = 0; i < 30; i++) {
					const value = audioDataArray[i * step] || 128;
					const deviation = Math.abs(value - 128);
					const height = Math.max(2, Math.min(20, 2 + (deviation / 128) * 36));
					bars.push(height);
				}
				waveformBars = bars;
				animationFrameId = requestAnimationFrame(updateWaveform);
			}
			animationFrameId = requestAnimationFrame(updateWaveform);

			// Start duration timer
			recordingInterval = setInterval(() => {
				recordingDuration++;
			}, 1000);

			// Set up Web Speech API for live transcription
			const SpeechRecognitionAPI = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
			if (SpeechRecognitionAPI) {
				const recognition = new SpeechRecognitionAPI() as SpeechRecognition;
				recognition.continuous = true;
				recognition.interimResults = true;
				recognition.lang = 'en-US';

				recognition.onresult = (event: SpeechRecognitionEvent) => {
					let interimTranscript = '';
					let finalTranscript = '';
					for (let i = event.resultIndex; i < event.results.length; i++) {
						const transcript = event.results[i][0].transcript;
						if (event.results[i].isFinal) {
							finalTranscript += transcript;
						} else {
							interimTranscript += transcript;
						}
					}
					liveTranscript = finalTranscript || interimTranscript;
				};

				recognition.onerror = (event: SpeechRecognitionErrorEvent) => {
					console.log('Speech recognition error:', event.error);
				};

				recognition.start();
				speechRecognition = recognition;
			}

			mediaRecorder.ondataavailable = (event) => {
				audioChunks.push(event.data);
			};

			mediaRecorder.onstop = async () => {
				const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
				await transcribeAudio(audioBlob);
				stream.getTracks().forEach(track => track.stop());
			};

			mediaRecorder.start();
			isRecording = true;
		} catch (error) {
			console.error('Failed to start recording:', error);
			alert('Microphone access denied. Please enable microphone permissions.');
		}
	}

	function stopRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.stop();
			isRecording = false;
		}
		if (recordingInterval) {
			clearInterval(recordingInterval);
			recordingInterval = null;
		}
		if (speechRecognition) {
			speechRecognition.stop();
			speechRecognition = null;
		}
		if (animationFrameId) {
			cancelAnimationFrame(animationFrameId);
			animationFrameId = null;
		}
		if (audioContext) {
			audioContext.close();
			audioContext = null;
			analyser = null;
			audioDataArray = null;
		}
		waveformBars = Array(30).fill(2);
		recordingDuration = 0;
		liveTranscript = '';
	}

	function cancelRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.ondataavailable = null;
			mediaRecorder.onstop = null;
			mediaRecorder.stop();
		}
		stopRecording();
		isRecording = false;
	}

	async function copyToClipboard(text: string, index: number) {
		try {
			await navigator.clipboard.writeText(text);
			copiedMessageId = index;
			setTimeout(() => copiedMessageId = null, 2000);
		} catch (e) {
			console.error('Failed to copy:', e);
		}
	}

	async function transcribeAudio(audioBlob: Blob) {
		isLoading = true;

		try {
			const formData = new FormData();
			formData.append('audio', audioBlob, 'recording.webm');

			const response = await apiClient.postFormData('/transcribe', formData);

			if (response.ok) {
				const data = await response.json();
				if (data.text) {
					inputValue = data.text;
					// Auto-submit after transcription
					await handleSubmit();
				}
			} else {
				const error = await response.json();
				console.error('Transcription failed:', error.message);
				messages = [...messages, {
					role: 'assistant',
					content: `Transcription not available: ${error.message || 'Unknown error'}`
				}];
			}
		} catch (error) {
			console.error('Transcription error:', error);
			messages = [...messages, {
				role: 'assistant',
				content: 'Voice transcription requires whisper.cpp to be installed locally.'
			}];
		} finally {
			isLoading = false;
		}
	}

	// Meeting recording state
	let meetingSession: any = null;
	let systemMediaRecorder: MediaRecorder | null = null;
	let systemAudioChunks: Blob[] = [];

	// Meeting recording
	async function startMeetingRecording() {
		isMeetingMode = true;

		// Check if we're in Electron
		if (browser && 'electron' in window) {
			const electron = (window as any).electron;

			try {
				// Start meeting session in main process
				meetingSession = await electron.meeting?.start({
					title: upcomingMeeting?.title || 'Meeting Recording',
					calendarEventId: upcomingMeeting?.id
				});

				// Request system audio capture using desktopCapturer
				// The renderer needs to capture audio via getUserMedia with specific constraints
				const constraints: MediaStreamConstraints = {
					audio: {
						// @ts-ignore - Chrome-specific constraints for system audio
						mandatory: {
							chromeMediaSource: 'desktop'
						}
					},
					video: {
						// @ts-ignore
						mandatory: {
							chromeMediaSource: 'desktop',
							maxWidth: 1,
							maxHeight: 1
						}
					}
				};

				const stream = await navigator.mediaDevices.getUserMedia(constraints);

				// Stop video track, we only need audio
				stream.getVideoTracks().forEach(track => track.stop());

				// Set up MediaRecorder for system audio
				systemMediaRecorder = new MediaRecorder(stream, {
					mimeType: 'audio/webm;codecs=opus'
				});
				systemAudioChunks = [];

				systemMediaRecorder.ondataavailable = (event) => {
					if (event.data.size > 0) {
						systemAudioChunks.push(event.data);
					}
				};

				systemMediaRecorder.onstop = async () => {
					// Combine all chunks and send to main process
					const audioBlob = new Blob(systemAudioChunks, { type: 'audio/webm' });
					const arrayBuffer = await audioBlob.arrayBuffer();

					if (meetingSession?.id) {
						await electron.meeting?.saveAudioChunk({
							sessionId: meetingSession.id,
							chunk: arrayBuffer,
							isLast: true
						});
					}

					stream.getTracks().forEach(track => track.stop());
				};

				systemMediaRecorder.start(10000); // Chunk every 10 seconds

				messages = [...messages, {
					role: 'assistant',
					content: `Meeting recording started${upcomingMeeting ? ` for "${upcomingMeeting.title}"` : ''}. I'm capturing system audio and will transcribe when you stop. Click the mic to add voice notes, or type questions.`
				}];
			} catch (error) {
				console.error('Failed to start meeting recording:', error);
				messages = [...messages, {
					role: 'assistant',
					content: 'Could not start system audio capture. Please grant screen/audio permissions in System Preferences > Security & Privacy > Privacy > Screen Recording.'
				}];
				isMeetingMode = false;
			}
		} else {
			// Web fallback - just mic recording
			messages = [...messages, {
				role: 'assistant',
				content: `Meeting mode started${upcomingMeeting ? ` for "${upcomingMeeting.title}"` : ''}. System audio capture requires the desktop app. Using microphone only.`
			}];
		}
	}

	async function stopMeetingRecording() {
		isMeetingMode = false;

		// Stop system audio recording
		if (systemMediaRecorder && systemMediaRecorder.state !== 'inactive') {
			systemMediaRecorder.stop();
		}

		// Stop meeting session in main process
		if (browser && 'electron' in window && meetingSession) {
			const electron = (window as any).electron;
			await electron.meeting?.stop();
		}

		messages = [...messages, {
			role: 'assistant',
			content: 'Meeting recording stopped. Processing audio and generating transcription...'
		}];

		// TODO: Trigger transcription and AI summarization
		// This would involve:
		// 1. Send audio to transcription endpoint
		// 2. Send transcription to AI for summary
		// 3. Extract action items and create tasks

		meetingSession = null;
		systemMediaRecorder = null;
		systemAudioChunks = [];

		setTimeout(() => {
			messages = [...messages, {
				role: 'assistant',
				content: 'Transcription and summary will be available in the full app. Open the main window to view meeting notes.'
			}];
		}, 2000);
	}

	function formatTime(dateStr: string): string {
		return new Date(dateStr).toLocaleTimeString('en-US', {
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	// Screenshot capture
	async function captureScreenshot() {
		if (isCapturingScreenshot) return;

		isCapturingScreenshot = true;

		try {
			if (browser && 'electron' in window) {
				const electron = (window as any).electron;

				// Hide popup temporarily for clean screenshot
				hidePopup();
				await new Promise(resolve => setTimeout(resolve, 200));

				const result = await electron.screenshot?.capture();

				// Show popup again
				if (browser && 'electron' in window) {
					// Re-show by toggling (popup will be shown again via the global shortcut or tray)
				}

				if (result?.success && result.dataUrl) {
					pendingScreenshot = result.dataUrl;
					// Add screenshot to input context
					messages = [...messages, {
						role: 'user',
						content: `[Screenshot captured - ${result.size.width}x${result.size.height}]`
					}];
					messages = [...messages, {
						role: 'assistant',
						content: 'Screenshot captured! You can describe what you want me to help with regarding this image.'
					}];
				} else {
					messages = [...messages, {
						role: 'assistant',
						content: `Screenshot capture failed: ${result?.error || 'Unknown error'}. Make sure BusinessOS has screen recording permission in System Preferences → Privacy & Security → Screen Recording.`
					}];
				}
			} else {
				// Web fallback - use clipboard API if available
				try {
					const items = await navigator.clipboard.read();
					for (const item of items) {
						if (item.types.includes('image/png')) {
							const blob = await item.getType('image/png');
							const reader = new FileReader();
							reader.onload = () => {
								pendingScreenshot = reader.result as string;
								messages = [...messages, {
									role: 'user',
									content: '[Image pasted from clipboard]'
								}];
							};
							reader.readAsDataURL(blob);
							return;
						}
					}
					messages = [...messages, {
						role: 'assistant',
						content: 'No image found in clipboard. Take a screenshot (Cmd+Shift+4 on Mac) and paste it here (Cmd+V).'
					}];
				} catch (e) {
					messages = [...messages, {
						role: 'assistant',
						content: 'Clipboard access requires permission. Take a screenshot and paste it, or use the desktop app for direct screenshot capture.'
					}];
				}
			}
		} catch (error) {
			console.error('Screenshot capture error:', error);
			messages = [...messages, {
				role: 'assistant',
				content: 'Screenshot capture failed. Please check screen recording permissions.'
			}];
		} finally {
			isCapturingScreenshot = false;
			scrollToBottom();
		}
	}

	// Clear pending screenshot
	function clearScreenshot() {
		pendingScreenshot = null;
	}
</script>

<div class="popup-container">
	<!-- Header -->
	<div class="popup-header">
		<div class="header-drag-region"></div>
		<div class="header-content">
			<div class="header-title">
				{#if isMeetingMode}
					<span class="recording-indicator"></span>
					<span>Meeting Mode</span>
				{:else}
					<svg class="header-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
					</svg>
					<span>Quick Chat</span>
				{/if}
			</div>
			<div class="header-actions">
				<!-- Size toggle -->
				<div class="size-toggle">
					<button
						class="size-btn"
						class:active={currentSize === 'small'}
						onclick={() => setSize('small')}
						title="Small"
					>S</button>
					<button
						class="size-btn"
						class:active={currentSize === 'medium'}
						onclick={() => setSize('medium')}
						title="Medium"
					>M</button>
					<button
						class="size-btn"
						class:active={currentSize === 'large'}
						onclick={() => setSize('large')}
						title="Large"
					>L</button>
				</div>
				<button class="header-btn" onclick={openMainWindow} title="Open full app (Cmd+Enter)">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
						<polyline points="15 3 21 3 21 9"/>
						<line x1="10" y1="14" x2="21" y2="3"/>
					</svg>
				</button>
				<button class="header-btn" onclick={hidePopup} title="Close (Esc)">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"/>
						<line x1="6" y1="6" x2="18" y2="18"/>
					</svg>
				</button>
			</div>
		</div>
	</div>

	<!-- Pull progress banner -->
	{#if isPulling && pullProgress}
		<div class="pull-banner">
			<div class="pull-info">
				<div class="mini-spinner"></div>
				<span>Pulling {pullingModel}...</span>
			</div>
			<span class="pull-status">{pullProgress.status}</span>
			{#if pullProgress.total && pullProgress.completed}
				<div class="pull-bar">
					<div class="pull-bar-fill" style="width: {Math.round((pullProgress.completed / pullProgress.total) * 100)}%"></div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Upcoming meeting banner -->
	{#if upcomingMeeting && !isMeetingMode}
		<div class="meeting-banner">
			<div class="meeting-info">
				<span class="meeting-time">{formatTime(upcomingMeeting.start)}</span>
				<span class="meeting-title">{upcomingMeeting.title}</span>
			</div>
			<button class="meeting-record-btn" onclick={startMeetingRecording}>
				<svg viewBox="0 0 24 24" fill="currentColor">
					<circle cx="12" cy="12" r="6"/>
				</svg>
				Record
			</button>
		</div>
	{/if}

	<!-- Meeting mode controls -->
	{#if isMeetingMode}
		<div class="meeting-controls">
			<button class="stop-recording-btn" onclick={stopMeetingRecording}>
				<svg viewBox="0 0 24 24" fill="currentColor">
					<rect x="6" y="6" width="12" height="12" rx="2"/>
				</svg>
				Stop Recording
			</button>
		</div>
	{/if}

	<!-- Messages -->
	<div class="messages-container" bind:this={messagesContainer}>
		{#if messages.length === 0}
			<div class="empty-state">
				<p>Ask me anything or start a meeting recording</p>
				<p class="shortcut-hint">Press <kbd>Cmd+Shift+Space</kbd> to toggle</p>
			</div>
		{:else}
			{#each messages as message, i}
				<div class="message {message.role}">
					<div class="message-content">
						{message.content}
					</div>
					{#if message.role === 'assistant'}
						<button
							class="copy-btn"
							class:copied={copiedMessageId === i}
							onclick={() => copyToClipboard(message.content, i)}
							title="Copy to clipboard"
						>
							{#if copiedMessageId === i}
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polyline points="20 6 9 17 4 12"/>
								</svg>
							{:else}
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
									<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
								</svg>
							{/if}
						</button>
					{/if}
				</div>
			{/each}
			{#if isLoading}
				<div class="message assistant">
					<div class="message-content loading">
						<span class="dot"></span>
						<span class="dot"></span>
						<span class="dot"></span>
					</div>
				</div>
			{/if}
		{/if}
	</div>

	<!-- Screenshot preview -->
	{#if pendingScreenshot}
		<div class="screenshot-preview">
			<img src={pendingScreenshot} alt="Screenshot preview" />
			<button class="remove-screenshot" onclick={clearScreenshot} title="Remove screenshot">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
				</svg>
			</button>
		</div>
	{/if}

	<!-- Input area -->
	<div class="input-area">
		{#if isRecording}
			<!-- Recording UI with waveform -->
			<div class="recording-ui">
				<!-- Live transcript -->
				{#if liveTranscript}
					<div class="live-transcript">{liveTranscript}</div>
				{:else}
					<div class="live-transcript placeholder">Listening...</div>
				{/if}
				<!-- Waveform bar -->
				<div class="waveform-bar">
					<button class="cancel-btn" onclick={cancelRecording} title="Cancel">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
						</svg>
					</button>
					<div class="waveform">
						{#each waveformBars as height}
							<div class="bar" style="height: {height}px"></div>
						{/each}
					</div>
					<span class="duration">{formatDuration(recordingDuration)}</span>
					<button class="confirm-btn" onclick={stopRecording} title="Done">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="20 6 9 17 4 12"/>
						</svg>
					</button>
				</div>
			</div>
		{:else}
			<button
				class="screenshot-btn"
				onclick={captureScreenshot}
				disabled={isCapturingScreenshot}
				title="Capture screenshot (Cmd+Shift+S)"
			>
				{#if isCapturingScreenshot}
					<div class="mini-spinner"></div>
				{:else}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
						<circle cx="8.5" cy="8.5" r="1.5"/>
						<polyline points="21 15 16 10 5 21"/>
					</svg>
				{/if}
			</button>
			<button
				class="mic-btn"
				onclick={toggleRecording}
				title="Voice input (Cmd+D)"
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
					<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
					<line x1="12" y1="19" x2="12" y2="23"/>
					<line x1="8" y1="23" x2="16" y2="23"/>
				</svg>
			</button>
			<textarea
				bind:this={inputElement}
				bind:value={inputValue}
				onkeydown={handleKeyDown}
				placeholder="Ask anything..."
				rows="1"
			></textarea>
			<button
				class="send-btn"
				onclick={handleSubmit}
				disabled={!inputValue.trim() || isLoading}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="22" y1="2" x2="11" y2="13"/>
					<polygon points="22 2 15 22 11 13 2 9 22 2"/>
				</svg>
			</button>
		{/if}
	</div>
</div>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
		background: transparent;
		overflow: hidden;
	}

	.popup-container {
		width: 100%;
		height: 100vh;
		display: flex;
		flex-direction: column;
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-radius: 12px;
		border: 1px solid rgba(0, 0, 0, 0.1);
		overflow: hidden;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	}

	/* Header */
	.popup-header {
		position: relative;
		background: rgba(249, 250, 251, 0.9);
		border-bottom: 1px solid rgba(0, 0, 0, 0.08);
	}

	.header-drag-region {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 32px;
		-webkit-app-region: drag;
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		-webkit-app-region: no-drag;
	}

	.header-title {
		display: flex;
		align-items: center;
		gap: 8px;
		font-weight: 600;
		font-size: 14px;
		color: #111;
	}

	.header-icon {
		width: 18px;
		height: 18px;
	}

	.recording-indicator {
		width: 8px;
		height: 8px;
		background: #ef4444;
		border-radius: 50%;
		animation: pulse 1.5s infinite;
	}

	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.5; }
	}

	.header-actions {
		display: flex;
		gap: 4px;
	}

	.header-btn {
		width: 28px;
		height: 28px;
		border: none;
		background: none;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #666;
		transition: all 0.15s;
	}

	.header-btn:hover {
		background: rgba(0, 0, 0, 0.08);
		color: #111;
	}

	.header-btn svg {
		width: 16px;
		height: 16px;
	}

	/* Size toggle */
	.size-toggle {
		display: flex;
		background: rgba(0, 0, 0, 0.05);
		border-radius: 6px;
		padding: 2px;
		gap: 1px;
	}

	.size-btn {
		width: 22px;
		height: 22px;
		border: none;
		background: none;
		border-radius: 4px;
		cursor: pointer;
		font-size: 10px;
		font-weight: 600;
		color: #999;
		transition: all 0.15s;
	}

	.size-btn:hover {
		color: #666;
		background: rgba(0, 0, 0, 0.05);
	}

	.size-btn.active {
		background: white;
		color: #111;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
	}

	/* Model selector */
	.model-selector {
		position: relative;
	}

	.model-btn {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 4px 8px;
		border: 1px solid rgba(0, 0, 0, 0.1);
		background: white;
		border-radius: 6px;
		cursor: pointer;
		font-size: 11px;
		color: #666;
		transition: all 0.15s;
	}

	.model-btn:hover {
		border-color: rgba(0, 0, 0, 0.2);
		color: #111;
	}

	.model-btn svg {
		width: 14px;
		height: 14px;
	}

	.model-name {
		max-width: 60px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.model-dropdown {
		position: absolute;
		top: 100%;
		right: 0;
		margin-top: 4px;
		min-width: 180px;
		max-height: 240px;
		overflow-y: auto;
		background: white;
		border: 1px solid rgba(0, 0, 0, 0.1);
		border-radius: 8px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		z-index: 100;
	}

	.model-option {
		display: flex;
		flex-direction: row;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		padding: 8px 12px;
		border: none;
		background: none;
		cursor: pointer;
		text-align: left;
		transition: background 0.1s;
	}

	.model-option:hover {
		background: #f3f4f6;
	}

	.model-option.selected {
		background: #e5e7eb;
	}

	.model-option-name {
		font-size: 13px;
		font-weight: 500;
		color: #111;
	}

	.model-option-provider {
		font-size: 11px;
		color: #666;
	}

	.model-option-empty {
		padding: 12px;
		text-align: center;
		color: #999;
		font-size: 12px;
	}

	.model-btn.pulling {
		background: #f3f4f6;
	}

	.mini-spinner {
		width: 14px;
		height: 14px;
		border: 2px solid #e5e7eb;
		border-top-color: #111;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.provider-indicator {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		background: #f9fafb;
		border-bottom: 1px solid #e5e7eb;
		font-size: 11px;
	}

	.provider-label {
		color: #666;
	}

	.provider-name {
		font-weight: 600;
		color: #111;
	}

	.dropdown-section {
		padding: 4px 0;
	}

	.dropdown-section:not(:last-child) {
		border-bottom: 1px solid #e5e7eb;
	}

	.dropdown-section-title {
		padding: 6px 12px;
		font-size: 10px;
		font-weight: 600;
		color: #666;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.model-option-info {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 2px;
	}

	.model-option-size {
		font-size: 10px;
		color: #999;
	}

	.model-status {
		font-size: 10px;
		padding: 2px 6px;
		border-radius: 4px;
		font-weight: 500;
	}

	.model-status.ready {
		background: #dcfce7;
		color: #166534;
	}

	.model-status.cloud {
		background: #dbeafe;
		color: #1e40af;
	}

	.model-status.download {
		background: #fef3c7;
		color: #92400e;
	}

	.model-settings-link {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 12px;
		font-size: 12px;
		color: #666;
		text-decoration: none;
		border-top: 1px solid #e5e7eb;
		margin-top: 4px;
	}

	.model-settings-link:hover {
		background: #f3f4f6;
		color: #111;
	}

	.model-settings-link svg {
		width: 14px;
		height: 14px;
	}

	/* Pull progress banner */
	.pull-banner {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 10px 16px;
		background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
		border-bottom: 1px solid rgba(0, 0, 0, 0.08);
	}

	.pull-info {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		font-weight: 500;
	}

	.pull-status {
		font-size: 11px;
		color: #666;
	}

	.pull-bar {
		height: 4px;
		background: #d1d5db;
		border-radius: 2px;
		overflow: hidden;
	}

	.pull-bar-fill {
		height: 100%;
		background: linear-gradient(90deg, #111 0%, #444 100%);
		transition: width 0.3s;
	}

	/* Meeting banner */
	.meeting-banner {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 16px;
		background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
		color: white;
	}

	.meeting-info {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.meeting-time {
		font-weight: 600;
		font-size: 13px;
	}

	.meeting-title {
		font-size: 13px;
		opacity: 0.9;
		max-width: 180px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.meeting-record-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		background: rgba(255, 255, 255, 0.2);
		border: none;
		border-radius: 6px;
		color: white;
		font-size: 12px;
		font-weight: 500;
		cursor: pointer;
		transition: background 0.15s;
	}

	.meeting-record-btn:hover {
		background: rgba(255, 255, 255, 0.3);
	}

	.meeting-record-btn svg {
		width: 12px;
		height: 12px;
	}

	/* Meeting controls */
	.meeting-controls {
		padding: 10px 16px;
		background: #fef2f2;
		border-bottom: 1px solid #fecaca;
	}

	.stop-recording-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 16px;
		background: #ef4444;
		border: none;
		border-radius: 6px;
		color: white;
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		width: 100%;
		justify-content: center;
	}

	.stop-recording-btn:hover {
		background: #dc2626;
	}

	.stop-recording-btn svg {
		width: 14px;
		height: 14px;
	}

	/* Messages */
	.messages-container {
		flex: 1;
		overflow-y: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.empty-state {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		color: #666;
		text-align: center;
		padding: 40px 20px;
	}

	.empty-state p {
		margin: 0;
		font-size: 14px;
	}

	.shortcut-hint {
		margin-top: 12px !important;
		font-size: 12px !important;
		color: #999;
	}

	.shortcut-hint kbd {
		background: #f3f4f6;
		padding: 2px 6px;
		border-radius: 4px;
		font-family: inherit;
		font-size: 11px;
	}

	.message {
		display: flex;
		max-width: 90%;
	}

	.message.user {
		align-self: flex-end;
	}

	.message.assistant {
		align-self: flex-start;
	}

	.message-content {
		padding: 10px 14px;
		border-radius: 16px;
		font-size: 14px;
		line-height: 1.5;
		white-space: pre-wrap;
	}

	.message.user .message-content {
		background: #111;
		color: white;
		border-bottom-right-radius: 4px;
	}

	.message.assistant .message-content {
		background: #f3f4f6;
		color: #111;
		border-bottom-left-radius: 4px;
	}

	.message-content.loading {
		display: flex;
		gap: 4px;
		padding: 14px 18px;
	}

	.dot {
		width: 8px;
		height: 8px;
		background: #999;
		border-radius: 50%;
		animation: bounce 1.4s infinite ease-in-out both;
	}

	.dot:nth-child(1) { animation-delay: -0.32s; }
	.dot:nth-child(2) { animation-delay: -0.16s; }

	@keyframes bounce {
		0%, 80%, 100% { transform: scale(0.8); }
		40% { transform: scale(1); }
	}

	/* Input area */
	.input-area {
		display: flex;
		align-items: flex-end;
		gap: 8px;
		padding: 12px 16px;
		border-top: 1px solid rgba(0, 0, 0, 0.08);
		background: rgba(249, 250, 251, 0.9);
	}

	.mic-btn {
		width: 40px;
		height: 40px;
		border: none;
		background: #f3f4f6;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #666;
		transition: all 0.15s;
		flex-shrink: 0;
	}

	.mic-btn:hover {
		background: #e5e7eb;
		color: #111;
	}

	.mic-btn.recording {
		background: #ef4444;
		color: white;
		animation: pulse 1.5s infinite;
	}

	.mic-btn svg {
		width: 20px;
		height: 20px;
	}

	textarea {
		flex: 1;
		padding: 10px 14px;
		border: 1px solid rgba(0, 0, 0, 0.1);
		border-radius: 20px;
		font-size: 14px;
		font-family: inherit;
		resize: none;
		outline: none;
		background: white;
		max-height: 120px;
		line-height: 1.4;
	}

	textarea:focus {
		border-color: #111;
	}

	textarea:disabled {
		background: #f9fafb;
		color: #999;
	}

	.send-btn {
		width: 40px;
		height: 40px;
		border: none;
		background: #111;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		transition: all 0.15s;
		flex-shrink: 0;
	}

	.send-btn:hover:not(:disabled) {
		background: #333;
	}

	.send-btn:disabled {
		background: #d1d5db;
		cursor: not-allowed;
	}

	.send-btn svg {
		width: 18px;
		height: 18px;
	}

	/* Recording UI */
	.recording-ui {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.live-transcript {
		font-size: 13px;
		color: #111;
		min-height: 20px;
		animation: pulse 2s infinite;
	}

	.live-transcript.placeholder {
		color: #999;
	}

	.waveform-bar {
		display: flex;
		align-items: center;
		gap: 8px;
		background: #1f2937;
		border-radius: 24px;
		padding: 6px 12px;
	}

	.waveform {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 2px;
		height: 24px;
	}

	.waveform .bar {
		width: 2px;
		background: white;
		border-radius: 1px;
	}

	.duration {
		font-size: 12px;
		font-family: monospace;
		color: white;
		min-width: 32px;
		text-align: right;
	}

	.cancel-btn, .confirm-btn {
		width: 28px;
		height: 28px;
		border: none;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.15s;
	}

	.cancel-btn {
		background: transparent;
		color: #9ca3af;
	}

	.cancel-btn:hover {
		color: white;
	}

	.confirm-btn {
		background: white;
		color: #1f2937;
	}

	.confirm-btn:hover {
		background: #e5e7eb;
	}

	.cancel-btn svg, .confirm-btn svg {
		width: 16px;
		height: 16px;
	}

	/* Copy button */
	.message {
		position: relative;
	}

	.copy-btn {
		position: absolute;
		bottom: -4px;
		right: 8px;
		width: 24px;
		height: 24px;
		border: none;
		background: #f3f4f6;
		border-radius: 4px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #666;
		opacity: 0;
		transition: all 0.15s;
	}

	.message:hover .copy-btn {
		opacity: 1;
	}

	.copy-btn:hover {
		background: #e5e7eb;
		color: #111;
	}

	.copy-btn.copied {
		color: #22c55e;
		opacity: 1;
	}

	.copy-btn svg {
		width: 14px;
		height: 14px;
	}

	/* Screenshot button */
	.screenshot-btn {
		width: 36px;
		height: 36px;
		border: none;
		background: #f3f4f6;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #666;
		transition: all 0.15s;
		flex-shrink: 0;
	}

	.screenshot-btn:hover:not(:disabled) {
		background: #e5e7eb;
		color: #111;
	}

	.screenshot-btn:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.screenshot-btn svg {
		width: 18px;
		height: 18px;
	}

	.screenshot-btn .mini-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid #e5e7eb;
		border-top-color: #111;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	/* Screenshot preview */
	.screenshot-preview {
		position: relative;
		margin: 0 16px 8px;
		border-radius: 8px;
		overflow: hidden;
		border: 1px solid rgba(0, 0, 0, 0.1);
		max-height: 120px;
	}

	.screenshot-preview img {
		width: 100%;
		height: auto;
		max-height: 120px;
		object-fit: cover;
		display: block;
	}

	.remove-screenshot {
		position: absolute;
		top: 6px;
		right: 6px;
		width: 24px;
		height: 24px;
		border: none;
		background: rgba(0, 0, 0, 0.6);
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		transition: all 0.15s;
	}

	.remove-screenshot:hover {
		background: rgba(0, 0, 0, 0.8);
	}

	.remove-screenshot svg {
		width: 14px;
		height: 14px;
	}

	/* ===== DARK MODE FOR POPUP CHAT ===== */
	:global(.dark) .popup-container {
		background: rgba(28, 28, 30, 0.95);
		border-color: rgba(255, 255, 255, 0.12);
	}

	:global(.dark) .popup-header {
		background: rgba(44, 44, 46, 0.9);
		border-bottom-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .header-title {
		color: #f5f5f7;
	}

	:global(.dark) .header-btn {
		color: #a1a1a6;
	}

	:global(.dark) .header-btn:hover {
		background: rgba(255, 255, 255, 0.1);
		color: #f5f5f7;
	}

	:global(.dark) .size-toggle {
		background: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .size-btn {
		color: #6e6e73;
	}

	:global(.dark) .size-btn:hover {
		color: #a1a1a6;
		background: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .size-btn.active {
		background: #3a3a3c;
		color: #f5f5f7;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
	}

	:global(.dark) .model-btn {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		color: #a1a1a6;
	}

	:global(.dark) .model-btn:hover {
		border-color: rgba(255, 255, 255, 0.2);
		color: #f5f5f7;
	}

	:global(.dark) .model-dropdown {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .model-option {
		color: #f5f5f7;
	}

	:global(.dark) .model-option:hover {
		background: #3a3a3c;
	}

	:global(.dark) .model-option.selected {
		background: rgba(10, 132, 255, 0.2);
	}

	:global(.dark) .messages-container {
		background: #1c1c1e;
	}

	:global(.dark) .user-message {
		background: #0A84FF;
		color: white;
	}

	:global(.dark) .assistant-message {
		background: #2c2c2e;
		color: #f5f5f7;
	}

	:global(.dark) .input-container {
		background: rgba(44, 44, 46, 0.9);
		border-top-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .chat-input {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.12);
		color: #f5f5f7;
	}

	:global(.dark) .chat-input:focus {
		border-color: #0A84FF;
	}

	:global(.dark) .chat-input::placeholder {
		color: #6e6e73;
	}

	:global(.dark) .action-btn {
		color: #6e6e73;
	}

	:global(.dark) .action-btn:hover {
		background: #3a3a3c;
		color: #f5f5f7;
	}

	:global(.dark) .send-btn {
		background: #0A84FF;
		color: white;
	}

	:global(.dark) .send-btn:disabled {
		background: #3a3a3c;
		color: #6e6e73;
	}

	:global(.dark) .category-title {
		color: #6e6e73;
	}

	:global(.dark) .empty-state {
		color: #6e6e73;
	}

	:global(.dark) .provider-section {
		background: #2c2c2e;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .provider-header {
		color: #a1a1a6;
	}

	:global(.dark) .pull-btn {
		background: #0A84FF;
		color: white;
	}
</style>
