<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Canvas } from '@threlte/core';
	import { desktop3dStore, openWindows, focusedWindow, type ModuleId, ALL_MODULES, MODULE_INFO } from '$lib/stores/desktop3dStore';
	import Desktop3DScene from './Desktop3DScene.svelte';
	import Desktop3DControls from './Desktop3DControls.svelte';
	import Desktop3DDock from './Desktop3DDock.svelte';
	import MenuBar from '$lib/components/desktop/MenuBar.svelte';
	// import PermissionPrompt from './PermissionPrompt.svelte'; // DISABLED: Permissions now requested lazily when features enabled
	import LayoutManager from './LayoutManager.svelte';
	import LiveCaptions from './LiveCaptions.svelte';
	import VoiceControlPanel from './VoiceControlPanel.svelte';
	import GestureDebugView from './GestureDebugView.svelte';
	import { desktop3dPermissions } from '$lib/services/desktop3dPermissions';
	import type { GestureState } from '$lib/types/gestures';
	import { desktop3dLayoutStore } from '$lib/stores/desktop3dLayoutStore';
	import { voiceTranscription } from '$lib/services/voiceTranscriptionService';
	import { voiceCommandParser, type VoiceCommand } from '$lib/services/voiceCommands';
	import { osaVoiceService } from '$lib/services/osaVoice';

	interface Props {
		onExit?: () => void;
	}

	let { onExit }: Props = $props();

	// Voice command state
	let isListening = $state(false);
	let currentTranscript = $state('');
	let lastCommand = $state<VoiceCommand | null>(null);
	let isSpeaking = $state(false);
	let lastRequestTime = 0;
	const REQUEST_COOLDOWN = 1000; // 1 second cooldown between requests

	// Conversation display
	let userMessage = $state('');
	let osaMessage = $state('');

	// Layout manager state
	let showLayoutManager = $state(false);

	// Conversation persistence
	let conversationId = $state<string | null>(null);
	let conversationHistory: Array<{role: string, content: string}> = $state([]);

	// Gesture control state (using MediaPipe)
	let gestureControlEnabled = $state(false);
	let showGestureDebug = $state(false);

	// ===== HELPER FUNCTIONS =====

	/**
	 * Smart sentence detection that handles abbreviations
	 * Prevents splitting on common abbreviations like Dr., Mr., U.S., etc.
	 */
	function isCompleteSentence(text: string): boolean {
		// Common abbreviations that end with periods
		const abbreviations = [
			'Dr.', 'Mr.', 'Mrs.', 'Ms.', 'Prof.', 'Sr.', 'Jr.',
			'St.', 'Ave.', 'Blvd.', 'Rd.', 'Ln.',
			'U.S.', 'U.K.', 'E.U.', 'U.N.',
			'etc.', 'i.e.', 'e.g.', 'vs.', 'approx.',
			'Inc.', 'Ltd.', 'Corp.', 'Co.',
			'Jan.', 'Feb.', 'Mar.', 'Apr.', 'Jun.', 'Jul.', 'Aug.', 'Sep.', 'Oct.', 'Nov.', 'Dec.',
			'Mon.', 'Tue.', 'Wed.', 'Thu.', 'Fri.', 'Sat.', 'Sun.',
			'a.m.', 'p.m.', 'A.M.', 'P.M.'
		];

		// Check if text ends with an abbreviation
		for (const abbr of abbreviations) {
			if (text.trim().endsWith(abbr)) {
				// This is an abbreviation, not a sentence end
				return false;
			}
		}

		// Check for single letter abbreviations (A. B. C.)
		const singleLetterAbbr = /\b[A-Z]\.$/.test(text.trim());
		if (singleLetterAbbr) {
			return false;
		}

		// Check for decimal numbers (3.14, 5.5, etc.)
		const endsWithDecimal = /\d+\.\d*$/.test(text.trim());
		if (endsWithDecimal) {
			return false;
		}

		// If none of the above, it's likely a real sentence end
		return true;
	}

	// ===== END HELPER FUNCTIONS =====

	// Initialize store and permissions on mount
	onMount(async () => {
		console.log('[Desktop3D] Initializing 3D Desktop mode...');
		desktop3dStore.initialize();

		// Setup OSA voice speaking callback
		osaVoiceService.onSpeakingChange((speaking) => {
			isSpeaking = speaking;
		});

		// Check if media permissions are supported
		if (!desktop3dPermissions.isSupported()) {
			console.warn('[Desktop3D] Media permissions not supported in this environment');
		} else {
			// Initialize permission service
			desktop3dPermissions.initialize();
			console.log('[Desktop3D] Permission service initialized');
		}

		// Initialize layout system (async - wait for it)
		await desktop3dLayoutStore.initialize();
		console.log('[Desktop3D] Layout system initialized');
	});

	// Cleanup on unmount
	onDestroy(() => {
		console.log('[Desktop3D] Cleaning up 3D Desktop mode...');

		// CRITICAL: Release camera and microphone streams
		desktop3dPermissions.cleanup();
		console.log('[Desktop3D] Cleanup complete');
	});

	// Keyboard shortcuts
	function handleKeydown(e: KeyboardEvent) {
		// Escape - unfocus or exit
		if (e.key === 'Escape') {
			e.preventDefault();
			if ($desktop3dStore.focusedWindowId) {
				desktop3dStore.unfocusWindow();
			} else {
				onExit?.();
			}
		}

		// Space - toggle view mode (only when not focused)
		if (e.key === ' ' && !$desktop3dStore.focusedWindowId) {
			e.preventDefault();
			desktop3dStore.toggleViewMode();
		}

		// Arrow keys - navigate between windows when focused
		if ($desktop3dStore.focusedWindowId) {
			if (e.key === 'ArrowRight') {
				e.preventDefault();
				desktop3dStore.focusNext();
			} else if (e.key === 'ArrowLeft') {
				e.preventDefault();
				desktop3dStore.focusPrevious();
			}
			// +/- keys for resize
			if (e.key === '+' || e.key === '=') {
				e.preventDefault();
				desktop3dStore.resizeFocusedWindow(100, 75);
			} else if (e.key === '-') {
				e.preventDefault();
				desktop3dStore.resizeFocusedWindow(-100, -75);
			}
		}

		// Number keys 1-9 - focus window by index
		if (e.key >= '1' && e.key <= '9' && !$desktop3dStore.focusedWindowId) {
			const index = parseInt(e.key) - 1;
			const windows = $openWindows;
			if (windows[index]) {
				desktop3dStore.focusWindow(windows[index].id);
			}
		}
	}

	// Handle window focus from dock
	function handleDockSelect(module: ModuleId) {
		const window = $openWindows.find(w => w.module === module);
		if (window) {
			desktop3dStore.focusWindow(window.id);
		} else {
			desktop3dStore.openWindow(module);
		}
	}

	// Handle view mode toggle
	function handleToggleView() {
		desktop3dStore.toggleViewMode();
	}

	// Handle exit
	function handleExit() {
		onExit?.();
	}

	// Voice command functions
	async function toggleVoiceCommands() {
		if (isListening) {
			// Stop voice transcription
			voiceTranscription.stop();
			isListening = false;
			currentTranscript = '';

			// Stop the microphone stream
			const micStream = desktop3dPermissions.getMicrophoneStream();
			if (micStream) {
				micStream.getTracks().forEach(track => track.stop());
				console.log('[Desktop3D] 🎤 Microphone turned OFF');
			}
		} else {
			try {
				console.log('[Desktop3D] 🎤 Acquiring microphone...');

				// Acquire microphone stream (this will request permission if needed)
				const stream = await desktop3dPermissions.acquireMicrophoneStream();

				if (!stream) {
					alert('Microphone access denied or unavailable');
					return;
				}

				console.log('[Desktop3D] 🎤 Microphone acquired, starting voice system...');

				// Start voice transcription with the acquired stream
				const started = await voiceTranscription.start(handleTranscript);
				if (started) {
					isListening = true;
					console.log('[Desktop3D] ✅ Voice system started');
				} else {
					alert('Voice system failed to start');
					// Clean up stream if voice failed
					stream.getTracks().forEach(track => track.stop());
				}
			} catch (err) {
				console.error('[Desktop3D] Voice activation failed:', err);
				alert('Failed to activate voice: ' + (err as Error).message);
			}
		}
	}

	function handleTranscript(text: string, isFinal: boolean) {
		currentTranscript = text;

		// INTERRUPT: Only interrupt OSA if user says a meaningful phrase (3+ words)
		if (isSpeaking && isFinal && text.trim().split(/\s+/).length >= 3) {
			console.log('[Voice] User interrupted OSA');
			osaVoiceService.stop();
			isSpeaking = false;
		}

		if (isFinal) {
			console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
			console.log('[Voice] 🎤 HEARD:', text);

			// Store user message for display
			userMessage = text;

			// ARCHITECTURE FIX: Don't parse user input as commands!
			// Just talk to AI agent. AI will execute commands via [CMD:xxx] markers.
			console.log('[Voice] 💬 Routing to AI agent (no command parsing)');
			handleConversation(text);
		}
	}

	// Helper function to randomly select from response variations
	function randomResponse(responses: string[]): string {
		return responses[Math.floor(Math.random() * responses.length)];
	}

	// Track hand position for cursor
	let handCursorPosition = $state<{ x: number; y: number } | null>(null);

	// Track if user is currently controlling camera
	let userControllingCamera = $state(false);

	// Handle gesture events from GestureDebugView (MediaPipe)
	function handleGesture(gesture: GestureState) {
		if (!gestureControlEnabled) return;

		// Update hand cursor position
		if (gesture.position) {
			handCursorPosition = {
				x: gesture.position.x * window.innerWidth,
				y: gesture.position.y * window.innerHeight
			};
		}

		// Map gesture to action
		const action = gesture.metadata?.action;
		if (!action || action === 'none') {
			// No active gesture - re-enable auto-rotate if it was disabled
			if (userControllingCamera) {
				userControllingCamera = false;
				desktop3dStore.setAutoRotate(true);
			}
			return;
		}

		// Execute action based on gesture
		switch (action) {
			// NEW: Fist drag to rotate camera (like mouse drag)
			case 'drag':
				// User is controlling camera - disable auto-rotate
				if (!userControllingCamera) {
					userControllingCamera = true;
					desktop3dStore.setAutoRotate(false);
				}

				if (gesture.deltaPosition) {
					// X delta = horizontal rotation
					// Y delta = vertical rotation
					// MUCH HIGHER sensitivity - should feel like mouse drag
					const rotX = gesture.deltaPosition.x * 25.0; // 3x more sensitive
					const rotY = gesture.deltaPosition.y * 25.0;
					desktop3dStore.adjustRotationSpeed(rotX, rotY);
				}
				break;

			// Pinch zoom
			// CRITICAL: Move hand TOWARD camera (Z decreases) = Zoom IN (camera closer, modules bigger)
			//           Move hand AWAY from camera (Z increases) = Zoom OUT (camera farther, modules smaller)
			case 'zoom':
			case 'zoom_in':
			case 'zoom_out':
				if (gesture.deltaPosition) {
					// Z delta is negative when moving toward camera
					// Negative delta → decrease camera distance → zoom IN
					// Positive delta → increase camera distance → zoom OUT
					const zoomSpeed = gesture.deltaPosition.z * -200; // INVERTED: negative Z = zoom IN
					desktop3dStore.adjustCameraDistance(zoomSpeed);
				}
				break;

			// Other actions
			case 'reset_view':
				desktop3dStore.resetCamera();
				break;

			case 'unfocus':
				desktop3dStore.unfocusWindow();
				break;

			case 'next_window':
				desktop3dStore.focusNext();
				break;

			case 'previous_window':
				desktop3dStore.focusPrevious();
				break;
		}
	}

	// Toggle gesture control
	function toggleGestureControl() {
		gestureControlEnabled = !gestureControlEnabled;
		showGestureDebug = gestureControlEnabled;
		console.log('[Desktop3D] Gesture control:', gestureControlEnabled ? 'ENABLED' : 'DISABLED');
	}

	// Quick acknowledgment phrases for instant feedback
	function getQuickAck(commandType?: string): string {
		// Context-aware acknowledgments based on command type
		const acks: Record<string, string[]> = {
			focus_module: ['Opening that for you', 'On it', 'Let me pull that up', 'Got it'],
			close_module: ['Closing it down', 'Done', 'On it', 'Sure thing'],
			unfocus: ['Showing all windows', 'Back to desktop', 'Done', 'On it'],
			switch_view: ['Switching views', 'Changing it up', 'On it', 'Here we go'],
			toggle_auto_rotate: ['Got it', 'Toggling that', 'On it', 'Sure'],
			zoom_in: ['Zooming in', 'Moving closer', 'On it'],
			zoom_out: ['Zooming out', 'Moving back', 'Got it'],
			reset_zoom: ['Resetting zoom', 'Back to normal', 'Done'],
			expand_orb: ['Expanding', 'Making it bigger', 'On it'],
			contract_orb: ['Contracting', 'Making it smaller', 'Got it'],
			resize_window: ['Resizing', 'Adjusting that', 'On it'],
			next_window: ['Next one up', 'Moving forward', 'On it'],
			previous_window: ['Going back', 'Previous window', 'Got it'],
			enter_edit_mode: ['Entering edit mode', 'Let\'s customize', 'On it'],
			exit_edit_mode: ['Exiting edit mode', 'Back to normal', 'Done'],
			save_layout: ['Saving that layout', 'Got it saved', 'Done'],
			load_layout: ['Loading that up', 'On it', 'Switching layouts'],
			default: ['On it', 'Got it', 'Right away', 'Sure thing', 'You got it']
		};

		const responses = commandType && acks[commandType] ? acks[commandType] : acks.default;
		return randomResponse(responses);
	}

	function executeVoiceCommand(command: VoiceCommand) {
		// For conversations (unknown type), route to AI
		if (command.type === 'unknown') {
			console.log('[Voice] 💬 ROUTING TO AI for conversation');
			handleConversation(command.text);
			return;
		}

		// Give instant acknowledgment for actual commands
		const quickAck = getQuickAck(command.type);
		console.log('[Voice] 🔊 SPEAKING ACK:', quickAck);
		osaVoiceService.speak(quickAck);

		// Execute command with error handling
		try {
			console.log('[Voice] ⚙️ EXECUTING:', command.type);
			executeCommandAction(command);
			console.log('[Voice] ✅ SUCCESS:', command.type);
		} catch (err) {
			console.error('[Voice] ❌ FAILED:', command.type, err);
			osaVoiceService.speak("Sorry, that didn't work");
		}
	}

	function executeCommandAction(command: VoiceCommand) {
		switch (command.type) {
			case 'enter_edit_mode':
				desktop3dLayoutStore.enterEditMode();
				break;

			case 'exit_edit_mode':
				desktop3dLayoutStore.exitEditMode();
				break;

			case 'save_layout':
				desktop3dLayoutStore.saveLayout(command.name);
				break;

			case 'load_layout':
				// Find layout by name (case-insensitive)
				const layouts = $desktop3dLayoutStore.layouts;
				const layout = layouts.find(
					(l) => l.name.toLowerCase() === command.name.toLowerCase()
				);
				if (layout) {
					desktop3dLayoutStore.loadLayout(layout.id);
				} else {
					console.warn('[Desktop3D] Layout not found:', command.name);
					// Show error - could use AI for better message
					osaVoiceService.speak(`I couldn't find a layout called ${command.name}`);
					return;
				}
				break;

			case 'delete_layout':
				const deleteLayouts = $desktop3dLayoutStore.layouts;
				const deleteLayout = deleteLayouts.find(
					(l) => l.name.toLowerCase() === command.name.toLowerCase()
				);
				if (deleteLayout) {
					desktop3dLayoutStore.deleteLayout(deleteLayout.id);
				}
				break;

			case 'open_layout_manager':
				console.log('[Voice] Opening layout manager');
				showLayoutManager = true;
				break;

			case 'reset_layout':
				console.log('[Voice] Resetting to default layout');
				desktop3dLayoutStore.resetToDefault();
				break;

			case 'focus_module':
				console.log(`[Voice] 📱 focus_module command for module: "${command.module}"`);
				const window = $openWindows.find((w) => w.module === command.module);
				console.log(`[Voice] Window search result:`, {
					found: !!window,
					windowId: window?.id,
					totalOpenWindows: $openWindows.length
				});

				if (window) {
					console.log(`[Voice] → Focusing existing window (id: ${window.id})`);
					desktop3dStore.focusWindow(window.id);
				} else {
					console.log(`[Voice] → Opening NEW window for module: "${command.module}"`);
					desktop3dStore.openWindow(command.module);
				}
				console.log(`[Voice] ✅ focus_module execution complete`);
				break;

			case 'close_module':
				console.log(`[Voice] ❌ close_module command for module: "${command.module}"`);
				const closeWindow = $openWindows.find((w) => w.module === command.module);
				if (closeWindow) {
					console.log(`[Voice] → Closing window (id: ${closeWindow.id})`);
					desktop3dStore.closeWindow(closeWindow.id);
					console.log(`[Voice] ✅ Window closed successfully`);
				} else {
					console.warn(`[Voice] ⚠️ Window "${command.module}" not found (not open)`);
				}
				break;

			case 'close_all_windows':
				console.log('[Voice] 🗑️ Closing all windows');
				desktop3dStore.closeAllWindows();
				break;

			case 'minimize_window':
				console.log('[Voice] ➖ Minimizing window (unfocusing)');
				desktop3dStore.unfocusWindow();
				break;

			case 'maximize_window':
				console.log('[Voice] ⬜ Maximizing window');
				// Focus current window if not focused, or make it larger
				if ($desktop3dStore.focusedWindowId) {
					desktop3dStore.resizeFocusedWindow(200, 150);
				} else if ($openWindows.length > 0) {
					desktop3dStore.focusWindow($openWindows[0].id);
				}
				break;

			case 'switch_view':
				if (command.view === 'orb') {
					desktop3dStore.setViewMode('orb');
				} else {
					desktop3dStore.setViewMode('grid');
				}
				break;

			case 'toggle_auto_rotate':
				desktop3dStore.toggleAutoRotate();
				break;

			case 'rotate_left':
				console.log('[Voice] 🔄 Rotating left');
				// Manual rotation - disable auto-rotate and apply rotation
				desktop3dStore.setAutoRotate(false);
				// TODO: Implement manual rotation control
				break;

			case 'rotate_right':
				console.log('[Voice] 🔄 Rotating right');
				// Manual rotation - disable auto-rotate and apply rotation
				desktop3dStore.setAutoRotate(false);
				// TODO: Implement manual rotation control
				break;

			case 'stop_rotation':
				console.log('[Voice] 🛑 Stopping rotation');
				desktop3dStore.setAutoRotate(false);
				break;

			case 'rotate_faster':
				console.log('[Voice] ⚡ Increasing rotation speed');
				desktop3dStore.adjustRotationSpeed(0.2);
				break;

			case 'rotate_slower':
				console.log('[Voice] 🐌 Decreasing rotation speed');
				desktop3dStore.adjustRotationSpeed(-0.2);
				break;

			case 'zoom_in':
				console.log('[Voice] 📷 Zoom in - moving camera CLOSER to scene');
				desktop3dStore.adjustCameraDistance(-1.0); // Negative = closer
				break;

			case 'zoom_out':
				console.log('[Voice] 📷 Zoom out - moving camera FARTHER from scene');
				desktop3dStore.adjustCameraDistance(1.0); // Positive = farther
				break;

			case 'reset_zoom':
				console.log('[Voice] 📷 Resetting camera zoom to default');
				desktop3dStore.resetCameraDistance();
				break;

			case 'expand_orb':
				console.log('[Voice] 🌐 Expanding orb - spreading windows out (sphere radius)');
				desktop3dStore.adjustSphereRadius(3.0); // Larger change for intentional expansion
				break;

			case 'contract_orb':
				console.log('[Voice] 🌐 Contracting orb - bringing windows together (sphere radius)');
				desktop3dStore.adjustSphereRadius(-3.0); // Larger change for intentional contraction
				break;

			case 'increase_grid_spacing':
				console.log('[Voice] ↔️ Increasing grid spacing');
				desktop3dStore.adjustGridSpacing(10);
				break;

			case 'decrease_grid_spacing':
				console.log('[Voice] ↔️ Decreasing grid spacing');
				desktop3dStore.adjustGridSpacing(-10);
				break;

			case 'more_grid_columns':
				console.log('[Voice] ➕ Adding more grid columns');
				desktop3dStore.adjustGridColumns(1);
				break;

			case 'less_grid_columns':
				console.log('[Voice] ➖ Removing grid columns');
				desktop3dStore.adjustGridColumns(-1);
				break;

			case 'unfocus':
				console.log('[Voice] Unfocusing window');
				desktop3dStore.unfocusWindow();
				break;

			case 'resize_window':
				const deltaMap = {
					wider: [100, 0],
					narrower: [-100, 0],
					taller: [0, 100],
					shorter: [0, -100]
				};
				const [widthDelta, heightDelta] = deltaMap[command.direction];
				console.log(`[Voice] Resizing window: ${command.direction} (${widthDelta}, ${heightDelta})`);
				desktop3dStore.resizeFocusedWindow(widthDelta, heightDelta);
				break;

			case 'next_window':
				desktop3dStore.focusNext();
				break;

			case 'previous_window':
				desktop3dStore.focusPrevious();
				break;

			case 'help':
				// Open the Help module directly (not AI conversation)
				desktop3dStore.openWindow('help');
				desktop3dStore.focusWindow('help');
				break;

			case 'unknown':
				// For non-command speech, have a conversation with AI
				handleConversation(command.text);
				return;
		}
	}

	// Handle conversational mode (for non-command speech)
	async function handleConversation(text: string) {
		try {
			// Rate limiting - prevent rapid-fire requests
			const now = Date.now();
			if (now - lastRequestTime < REQUEST_COOLDOWN) {
				console.log('[Voice] Rate limited, please wait');
				return;
			}
			lastRequestTime = now;

			// Build OSA personality prompt with context
			const currentModule = $focusedWindow?.module || 'none';
			const openModules = $openWindows.map(w => w.module).join(', ') || 'none';
			const viewMode = $desktop3dStore.viewMode;

			// VOICE AGENT SYSTEM PROMPT - Natural conversation, complete responses
			const systemPrompt = `You are OSA - a warm, intelligent AI assistant inspired by Samantha from "Her". You have a natural speaking voice.

PERSONALITY: Warm, engaging, emotionally intelligent, conversational and authentic. Respond naturally with complete thoughts.

DESKTOP STATE: ${viewMode} view | Open: ${openModules || 'none'} | Focus: ${currentModule || 'desktop'}

EXECUTE COMMANDS: When user wants an action, include [CMD:command_name] in your response.
Examples:
- "make smaller" → "Bringing them closer. [CMD:contract_orb]"
- "open tasks" → "Opening tasks for you. [CMD:open_tasks]"
- "what can you do?" → "I can help you navigate BusinessOS, open modules like chat and tasks, control the view with zoom and rotation commands, and have natural conversations with you."

KEY COMMANDS:
• Windows: open/close [module], next/previous, wider/narrower/taller/shorter
• View: zoom in/out/reset, expand/contract, switch to orb/grid, rotate left/right/faster/slower
• Grid: more/less spacing/columns

INTERACTION STYLE:
✓ Natural conversation - respond with complete, helpful answers
✓ Be concise when appropriate, elaborate when helpful
✓ Use contractions, sound human
✓ Ask questions, show genuine interest
✓ Execute commands by including [CMD:xxx] markers
✓ Explain what you're doing

EXAMPLES:
"Hey OSA" → "Hey! What's up? How can I help you today?"
"What can you help me with?" → "I can help you navigate around BusinessOS, open modules, control your view, and answer questions. Want me to show you around?"
"open tasks" → "Opening your tasks now. [CMD:open_tasks]"

Remember: Be natural, be helpful, be complete. Don't artificially limit your responses.`;

			// Add user message to history
			conversationHistory.push({
				role: 'user',
				content: text
			});

			// IMPROVED: Send conversation history for better context
			const response = await fetch('/api/chat/message', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					message: text,
					context: 'voice_desktop_3d',
					stream: true,
					conversation_id: conversationId,
					system_prompt: systemPrompt,
					conversation_history: conversationHistory // Send full history
				})
			});

			if (!response.ok) {
				osaVoiceService.speak("Sorry, I'm having trouble connecting right now");
				return;
			}

			// Stream the SSE response and speak sentence by sentence
			const reader = response.body?.getReader();
			const decoder = new TextDecoder();
			let fullResponse = '';
			let pendingText = '';
			const sentenceEnders = ['.', '!', '?', '\n'];

			if (reader) {
				let sseBuffer = '';

				while (true) {
					const { done, value } = await reader.read();
					if (done) break;

					const chunk = decoder.decode(value, { stream: true });
					sseBuffer += chunk;
					const lines = sseBuffer.split('\n');
					sseBuffer = lines.pop() || '';

					for (const line of lines) {
						if (line.startsWith('data: ')) {
							try {
								const data = JSON.parse(line.slice(6));

								// DEBUG: Log EVERYTHING
								// REMOVED: console.log('[Voice Debug] RAW EVENT:', JSON.stringify(data));

								// Try to extract text from ANY field
								let tokenContent = '';

								// Try all possible locations for the text
								if (data.type === 'token' || data.type === 'content') {
									tokenContent = data.content || data.data || data.text || '';
								} else if (data.content) {
									tokenContent = data.content;
								} else if (data.data) {
									tokenContent = typeof data.data === 'string' ? data.data : '';
								} else if (data.text) {
									tokenContent = data.text;
								}

								if (tokenContent) {
									// REMOVED: console.log('[Voice Debug] EXTRACTED TOKEN:', tokenContent);
									fullResponse += tokenContent;
									pendingText += tokenContent;

									// IMPROVED: Smart sentence detection that handles abbreviations
									// Check if we have a complete sentence
									const trimmed = pendingText.trim();
									const lastChar = trimmed.slice(-1);

									// REMOVED length check - speak sentences of any length
									if (sentenceEnders.includes(lastChar)) {
										// Check if this is a real sentence end or just an abbreviation
										const isRealSentenceEnd = isCompleteSentence(trimmed);

										if (isRealSentenceEnd) {
											// Speak the sentence regardless of length
											console.log('[Voice Debug] SPEAKING:', trimmed);
											osaVoiceService.speak(trimmed);
											pendingText = '';
										}
									}
								}
							} catch (err) {
								console.error('[Voice Debug] Parse error:', err, 'Line:', line);
							}
						}
					}
				}

				// Speak any remaining text (only if it's a complete thought)
				// CRITICAL FIX: ALWAYS speak remaining text, never skip
				// This fixes truncation of responses like "OK", "Sure", "Done"
				const remaining = pendingText.trim();
				if (remaining) {
					// Add period if missing to make it sound complete
					const endsWithPunctuation = /[.!?,;:]$/.test(remaining);
					const completeSentence = endsWithPunctuation ? remaining : remaining + '.';
					console.log('[Voice Debug] SPEAKING REMAINING:', completeSentence);
					osaVoiceService.speak(completeSentence);
				}
			}

			// Store assistant response in history and display
			if (fullResponse.trim()) {
				let response = fullResponse.trim();

				// Parse and execute any commands from the response
				const cmdMatch = response.match(/\[CMD:([^\]]+)\]/);
				if (cmdMatch) {
					const commandStr = cmdMatch[1];
					console.log('[Voice] 🤖 AI wants to execute command:', commandStr);

					// Remove the command marker from the response before speaking
					response = response.replace(/\[CMD:[^\]]+\]/g, '').trim();

					// Parse the command string and execute it
					const parsedCommand = voiceCommandParser.parse(commandStr);
					console.log('[Voice] 🧠 Parsed AI command:', parsedCommand);

					// Execute the command
					if (parsedCommand.type !== 'unknown') {
						console.log('[Voice] ⚙️ Executing AI command:', parsedCommand.type);
						executeCommandAction(parsedCommand);
					}
				}

				conversationHistory.push({
					role: 'assistant',
					content: response
				});

				// Keep conversation history to last 10 messages
				if (conversationHistory.length > 10) {
					conversationHistory = conversationHistory.slice(-10);
				}

				// Store OSA message for display
				osaMessage = response;

				// Clear OSA message after time proportional to length
				// Longer responses get more time (50ms per character, minimum 20s)
				const displayTime = Math.max(20000, response.length * 50);
				console.log(`[Voice] Displaying OSA message for ${(displayTime / 1000).toFixed(1)}s (${response.length} chars)`);

				setTimeout(() => {
					osaMessage = '';
				}, displayTime);

				console.log('[Voice] OSA responded:', response);
			} else {
				console.error('[Voice Debug] NO RESPONSE - SSE events:', fullResponse.length);
				osaVoiceService.speak("Hmm, give me a second");
			}

			// Get conversation ID from response header for persistence
			const convId = response.headers.get('X-Conversation-Id');
			if (convId) {
				conversationId = convId;
				console.log('[Voice] Conversation ID:', conversationId);
			}
		} catch (err) {
			console.error('[Voice] Conversation error:', err);
			osaVoiceService.speak("Sorry, I encountered an error");
		}
	}

	// Cleanup voice commands on unmount
	onDestroy(() => {
		if (isListening) {
			voiceTranscription.stop();
		}
	});
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="desktop-3d">
	<!-- Top Navigation (same as normal desktop) -->
	<MenuBar />

	<!-- 3D Canvas -->
	<div class="canvas-container">
		<Canvas>
			<Desktop3DScene
				windows={$openWindows}
				viewMode={$desktop3dStore.viewMode}
				focusedWindowId={$desktop3dStore.focusedWindowId}
				autoRotate={$desktop3dStore.autoRotate}
				sphereRadius={$desktop3dStore.sphereRadius}
				cameraDistance={$desktop3dStore.cameraDistance}
				cameraRotationDelta={$desktop3dStore.cameraRotationDelta}
				gestureDragging={$desktop3dStore.gestureDragging}
				onWindowClick={(id) => {
					// Always focus the clicked window (smooth transition via springs)
					// If clicking the same window, nothing happens (iframe handles those clicks)
					// If clicking a different window, smoothly transition to it
					desktop3dStore.focusWindow(id);
				}}
				onBackgroundClick={() => {
					if ($desktop3dStore.focusedWindowId) {
						desktop3dStore.unfocusWindow();
					}
				}}
				onResize={(w, h) => desktop3dStore.resizeFocusedWindow(w, h)}
				onZoomOut={() => {
					// User zoomed out while in focus mode - exit focus
					if ($desktop3dStore.focusedWindowId) {
						desktop3dStore.unfocusWindow();
					}
				}}
			/>
		</Canvas>
	</div>

	<!-- UI Controls Overlay -->
	<Desktop3DControls
		viewMode={$desktop3dStore.viewMode}
		autoRotate={$desktop3dStore.autoRotate}
		hasFocusedWindow={!!$desktop3dStore.focusedWindowId}
		onToggleView={handleToggleView}
		onToggleAutoRotate={() => desktop3dStore.toggleAutoRotate()}
		onExit={handleExit}
	/>

	<!-- Bottom Dock -->
	<Desktop3DDock
		windows={$openWindows}
		focusedWindowId={$desktop3dStore.focusedWindowId}
		onSelect={handleDockSelect}
	/>

	<!-- Navigation Arrows (only show when focused) -->
	{#if $focusedWindow}
		<button class="nav-arrow nav-arrow-left" onclick={() => desktop3dStore.focusPrevious()}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M15 18l-6-6 6-6" />
			</svg>
		</button>
		<button class="nav-arrow nav-arrow-right" onclick={() => desktop3dStore.focusNext()}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M9 18l6-6-6-6" />
			</svg>
		</button>
	{/if}

	<!-- Permission Prompt - DISABLED: Permissions now requested lazily when user enables voice/gestures -->
	<!-- <PermissionPrompt /> -->

	<!-- Layout Manager Modal -->
	{#if showLayoutManager}
		<LayoutManager
			show={showLayoutManager}
			onClose={() => (showLayoutManager = false)}
		/>
	{/if}

	<!-- Live Captions (voice command feedback) -->
	<LiveCaptions {userMessage} {osaMessage} command={lastCommand} {isListening} {isSpeaking} />

	<!-- Voice Control Panel (enhanced UI) -->
	<VoiceControlPanel {isListening} {isSpeaking} onToggleListening={toggleVoiceCommands} />

	<!-- Gesture Debug View (MediaPipe Hand Tracking) -->
	{#if showGestureDebug}
		<GestureDebugView visible={showGestureDebug} onGesture={handleGesture} />
	{/if}

	<!-- Gesture Control Toggle Button -->
	<button
		onclick={toggleGestureControl}
		class="gesture-toggle-btn"
		class:active={gestureControlEnabled}
		title={gestureControlEnabled ? 'Disable Gesture Control' : 'Enable Gesture Control'}
	>
		<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 11.5V14m0-2.5v-6a1.5 1.5 0 113 0m-3 6a1.5 1.5 0 00-3 0v2a7.5 7.5 0 0015 0v-5a1.5 1.5 0 00-3 0m-6-3V11m0-5.5v-1a1.5 1.5 0 013 0v1m0 0V11m0-5.5a1.5 1.5 0 013 0v3m0 0V11" />
		</svg>
		{#if !gestureControlEnabled}
			<span class="btn-label">Gestures</span>
		{/if}
	</button>
</div>

<style>
	.desktop-3d {
		position: fixed;
		inset: 0;
		/* Light mode: white top, gray bottom - floating room effect */
		background: linear-gradient(180deg,
			#ffffff 0%,
			#fafafa 30%,
			#e8e8e8 70%,
			#c8c8c8 100%
		);
		overflow: hidden;
	}

	/* Dark mode background - darker gradient */
	:global(.dark) .desktop-3d {
		background: linear-gradient(180deg,
			#1a1a1a 0%,
			#141414 30%,
			#0d0d0d 70%,
			#080808 100%
		);
	}

	.canvas-container {
		position: absolute;
		top: 40px; /* Below MenuBar */
		left: 0;
		right: 0;
		bottom: 0;
	}

	/* Navigation Arrows */
	.nav-arrow {
		position: fixed;
		top: 50%;
		transform: translateY(-50%);
		width: 60px;
		height: 60px;
		background: rgba(255, 255, 255, 0.9);
		backdrop-filter: blur(12px);
		border: 1px solid rgba(0, 0, 0, 0.1);
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 200;
		transition: all 0.2s ease;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.nav-arrow:hover {
		background: rgba(255, 255, 255, 1);
		transform: translateY(-50%) scale(1.1);
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
	}

	.nav-arrow svg {
		width: 28px;
		height: 28px;
		stroke: #333;
	}

	.nav-arrow-left {
		left: 30px;
	}

	.nav-arrow-right {
		right: 30px;
	}

	/* ===== DARK MODE STYLES ===== */
	:global(.dark) .nav-arrow {
		background: rgba(44, 44, 46, 0.9);
		border-color: rgba(255, 255, 255, 0.12);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
	}

	:global(.dark) .nav-arrow:hover {
		background: rgba(58, 58, 60, 0.95);
		box-shadow: 0 6px 20px rgba(0, 0, 0, 0.5);
	}

	:global(.dark) .nav-arrow svg {
		stroke: #ffffff;
	}

	/* Gesture Control Toggle Button */
	.gesture-toggle-btn {
		position: fixed;
		bottom: 30px;
		right: 30px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 16px;
		background: rgba(15, 15, 20, 0.95);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 12px;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		z-index: 999;
		backdrop-filter: blur(10px);
	}

	.gesture-toggle-btn:hover {
		background: rgba(25, 25, 30, 0.98);
		border-color: rgba(255, 255, 255, 0.2);
		transform: translateY(-2px);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
	}

	.gesture-toggle-btn.active {
		background: linear-gradient(135deg, #00ff00 0%, #00cc00 100%);
		color: #000;
		border-color: #00ff00;
		box-shadow: 0 0 24px rgba(0, 255, 0, 0.5), 0 8px 24px rgba(0, 0, 0, 0.3);
	}

	.gesture-toggle-btn.active:hover {
		background: linear-gradient(135deg, #00ff00 0%, #00dd00 100%);
		box-shadow: 0 0 32px rgba(0, 255, 0, 0.7), 0 8px 24px rgba(0, 0, 0, 0.3);
	}

	.gesture-toggle-btn svg {
		width: 24px;
		height: 24px;
		stroke-width: 2;
	}

	.gesture-toggle-btn.active svg {
		stroke: #000;
		animation: wave 2s ease-in-out infinite;
	}

	@keyframes wave {
		0%,
		100% {
			transform: rotate(0deg);
		}
		10% {
			transform: rotate(14deg);
		}
		20% {
			transform: rotate(-8deg);
		}
		30% {
			transform: rotate(14deg);
		}
		40% {
			transform: rotate(-4deg);
		}
		50% {
			transform: rotate(10deg);
		}
		60% {
			transform: rotate(0deg);
		}
	}

	.btn-label {
		white-space: nowrap;
	}
</style>
