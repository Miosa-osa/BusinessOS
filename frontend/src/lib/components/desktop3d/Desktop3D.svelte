<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Canvas } from '@threlte/core';
	import { desktop3dStore, openWindows, focusedWindow, type ModuleId, ALL_MODULES, MODULE_INFO } from '$lib/stores/desktop3dStore';
	import { userAppsStore } from '$lib/stores/userAppsStore';
	import { currentWorkspaceId, loadSavedWorkspace } from '$lib/stores/workspaces';
	import Desktop3DScene from './Desktop3DScene.svelte';
	import Desktop3DControls from './Desktop3DControls.svelte';
	import Desktop3DDock from './Desktop3DDock.svelte';
	import MenuBar from '$lib/components/desktop/MenuBar.svelte';
	import { openAppRegistry } from '$lib/stores/appRegistryStore';
	// import PermissionPrompt from './PermissionPrompt.svelte'; // DISABLED: Permissions now requested lazily when features enabled
	import LayoutManager from './LayoutManager.svelte';
	// Voice is now GLOBAL - managed in +layout.svelte, not Desktop3D!
	import { SimpleGestureController } from '$lib/services/simpleGestureController';
	import * as THREE from 'three';
	import { desktop3dLayoutStore } from '$lib/stores/desktop3dLayoutStore';
	import { useSession } from '$lib/auth-client';

	// STUBS: Old voice services removed - LiveKit handles everything now
	// These stubs prevent runtime errors while we clean up legacy code
	const osaVoiceService = {
		speak: (text: string, emotion?: string) => console.log('[Stub] OSA speak (ignored):', text?.substring(0, 30)),
		stop: () => {},
		onSpeakingChange: (cb: (speaking: boolean) => void) => {}
	};
	const emotionalTTS = {
		detectEmotion: (text: string) => 'neutral' as const
	};
	const stateObserver = {
		currentState: { openModules: [] as string[] },
		trackModuleOpen: (module: string) => {},
		trackModuleClose: (module: string) => {}
	};
	const executionOrchestrator = null;
	// voiceTranscription stub (old Deepgram system)
	const voiceTranscription = {
		start: async (cb: any) => { console.log('[Stub] voiceTranscription.start (ignored)'); return false; },
		stop: () => console.log('[Stub] voiceTranscription.stop (ignored)')
	};

	interface Props {
		onExit?: () => void;
	}

	let { onExit }: Props = $props();

	const session = useSession();

	// Voice is now GLOBAL - no local state needed!
	let currentTranscript = $state('');
	let lastCommand = $state<VoiceCommand | null>(null);
	let lastRequestTime = 0;
	const REQUEST_COOLDOWN = 1000; // 1 second cooldown between requests

	// Layout manager state
	let showLayoutManager = $state(false);

	// Conversation persistence
	let conversationId = $state<string | null>(null);
	let conversationHistory: Array<{role: string, content: string}> = $state([]);

	// Store unsubscribe function for desktop state
	let unsubscribeDesktopState: (() => void) | null = null;

	// Store voice command event handlers for cleanup
	let handleVoiceOpenApp: EventListener | null = null;
	let handleVoiceActivateNode: EventListener | null = null;

	// Gesture control state (SIMPLE - using SimpleGestureController)
	let gestureControlEnabled = $state(false);
	let gestureControlLoading = $state(false);
	let gestureController: SimpleGestureController | null = $state(null);
	let gestureVideoElement: HTMLVideoElement | null = $state(null);

	// OrbitControls reference (needed for direct camera manipulation)
	let orbitControlsRef: any = $state(null);

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

		// Initialize workspace store
		loadSavedWorkspace();

		// Fetch user apps first if we have a workspace
		if ($currentWorkspaceId) {
			console.log('[Desktop3D] Fetching user apps for workspace:', $currentWorkspaceId);
			await userAppsStore.fetch($currentWorkspaceId);
		}

		// Initialize with user apps
		desktop3dStore.initialize($userAppsStore.apps);

		// Wait for OrbitControls to be ready
		setTimeout(() => {
			if (orbitControlsRef) {
				console.log('[Desktop3D] ✅ OrbitControls ready for gesture control');
			} else {
				console.warn('[Desktop3D] ⚠️ OrbitControls not yet available (might take a moment)');
			}
		}, 2000);

		// Voice is now GLOBAL - no callbacks needed here
		console.log('[Desktop3D] Voice managed globally by +layout.svelte');

		// Setup voice command event listeners for SSE integration
		handleVoiceOpenApp = ((event: CustomEvent) => {
			const { app } = event.detail;
			console.log('[Desktop3D] Voice command: open app', app);

			// Open app via app registry or desktop store
			if (app === 'app-store' || app === 'business-os') {
				openAppRegistry();
			} else {
				// Try to open as a module if it matches
				const moduleId = app as ModuleId;
				if (ALL_MODULES.includes(moduleId)) {
					desktop3dStore.openWindow(moduleId);
				} else {
					console.warn('[Desktop3D] Unknown app:', app);
				}
			}
		}) as EventListener;

		handleVoiceActivateNode = ((event: CustomEvent) => {
			const { nodeId } = event.detail;
			console.log('[Desktop3D] Voice command: activate node', nodeId);
			// The voiceCommands service already navigates to /nodes/{nodeId}
			// Here we could do additional UI updates if needed
		}) as EventListener;

		window.addEventListener('voice:open-app', handleVoiceOpenApp);
		window.addEventListener('voice:activate-node', handleVoiceActivateNode);

		console.log('[Desktop3D] ✅ Voice command event listeners registered');

		// Initialize layout system (async - wait for it)
		await desktop3dLayoutStore.initialize();
		console.log('[Desktop3D] Layout system initialized');
	});

	// Cleanup on unmount
	onDestroy(() => {
		console.log('[Desktop3D] Cleaning up 3D Desktop mode...');

		// Cleanup gesture controller
		if (gestureController) {
			gestureController.destroy();
		}

		// Cleanup desktop state subscription
		if (unsubscribeDesktopState) {
			unsubscribeDesktopState();
		}

		// Cleanup voice command event listeners
		if (handleVoiceOpenApp) {
			window.removeEventListener('voice:open-app', handleVoiceOpenApp);
		}
		if (handleVoiceActivateNode) {
			window.removeEventListener('voice:activate-node', handleVoiceActivateNode);
		}

		console.log('[Desktop3D] Cleanup complete');
	});

	// Keyboard shortcuts
	function handleKeydown(e: KeyboardEvent) {
		// CRITICAL: Don't intercept keys when user is typing in an input, textarea, or iframe (terminal!)
		// This allows terminal to receive ALL keyboard input including arrow keys, Enter, etc.
		const target = e.target as HTMLElement;
		const activeEl = document.activeElement;

		// DEBUG: Log keyboard events to diagnose focus issues
		console.log('[Desktop3D] Key pressed:', e.key, 'target:', target?.tagName, 'activeElement:', activeEl?.tagName);

		const isInteractiveElement =
			target?.tagName === 'INPUT' ||
			target?.tagName === 'TEXTAREA' ||
			target?.isContentEditable ||
			target?.closest('iframe') ||
			activeEl?.tagName === 'IFRAME';

		// Escape - unfocus or exit (ALWAYS allow this, even in terminal)
		if (e.key === 'Escape') {
			e.preventDefault();
			if ($desktop3dStore.focusedWindowId) {
				desktop3dStore.unfocusWindow();
			} else {
				onExit?.();
			}
			return; // Early return after handling Escape
		}

		// Don't handle any other shortcuts when user is interacting with terminal/inputs
		if (isInteractiveElement) {
			console.log('[Desktop3D] Skipping shortcut - interactive element has focus');
			return;
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
	// Voice toggle removed - handled globally by +layout.svelte

	/**
	 * Execute a command from the Voice Agent backend
	 * Maps ParsedCommand to desktop3dStore actions
	 */
	// REMOVED: executeVoiceAgentCommand - LiveKit handles commands directly via callbacks

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

			try {
				// VOICE AGENT: Try local command parsing first, then AI for conversation
				const command = voiceCommandParser.parse(text);
				console.log('[Voice] 🧠 Parsed command:', command.type);

				if (command.type !== 'unknown') {
					// Known command - execute directly and give voice feedback
					console.log('[Voice] ⚡ Executing command directly:', command.type);
					executeCommandAction(command);
					lastCommand = command;

					// Give contextual voice feedback
					const feedback = getCommandFeedback(command);
					osaMessage = feedback;
					console.log('[Voice] 🔊 Speaking feedback:', feedback);
					osaVoiceService.speak(feedback);
				} else {
					// Unknown/conversational - route to AI for response
					console.log('[Voice] 💬 Routing to AI conversation for:', text);
					handleConversation(text);
				}
			} catch (err) {
				console.error('[Voice] ❌ Error processing voice:', err);
				// Fallback: route to conversation
				console.log('[Voice] 💬 Fallback to AI conversation');
				handleConversation(text);
			}
		}
	}

	/**
	 * Get voice feedback for a command
	 */
	function getCommandFeedback(command: VoiceCommand): string {
		switch (command.type) {
			case 'focus_module':
			case 'open_module':
				return `Opening ${command.module}.`;
			case 'close_module':
				return `Closing ${command.module}.`;
			case 'close_all_windows':
				return 'Closing all windows.';
			case 'switch_view':
				return `Switching to ${command.view} view.`;
			case 'zoom_in':
				return 'Zooming in.';
			case 'zoom_out':
				return 'Zooming out.';
			case 'reset_zoom':
				return 'Resetting zoom.';
			case 'toggle_auto_rotate':
				return 'Toggling rotation.';
			case 'rotate_left':
				return 'Rotating left.';
			case 'rotate_right':
				return 'Rotating right.';
			case 'stop_rotation':
				return 'Stopping rotation.';
			case 'expand_orb':
				return 'Expanding.';
			case 'contract_orb':
				return 'Contracting.';
			case 'next_window':
				return 'Next window.';
			case 'previous_window':
				return 'Previous window.';
			case 'minimize_window':
				return 'Minimizing.';
			case 'maximize_window':
				return 'Maximizing.';
			case 'unfocus':
				return 'Going back.';
			case 'enter_edit_mode':
				return 'Entering edit mode.';
			case 'exit_edit_mode':
				return 'Exiting edit mode.';
			case 'save_layout':
				return `Saving layout as ${command.name}.`;
			case 'load_layout':
				return `Loading layout ${command.name}.`;
			case 'reset_layout':
				return 'Resetting to default layout.';
			case 'help':
				return 'I can open modules, navigate windows, control the camera, and switch views. What would you like to do?';
			default:
				return 'Got it.';
		}
	}

	// Helper function to randomly select from response variations
	function randomResponse(responses: string[]): string {
		return responses[Math.floor(Math.random() * responses.length)];
	}

	// ===== SIMPLE GESTURE CONTROL =====

	/**
	 * Enable gesture control - initialize SimpleGestureController
	 */
	async function enableGestureControl() {
		// Prevent double initialization
		if (gestureControlLoading || gestureControlEnabled) {
			console.log('[Desktop3D] Already initializing or enabled, skipping...');
			return;
		}

		if (!gestureVideoElement) {
			console.error('[Desktop3D] ❌ Video element not found');
			alert('Error: Video element not initialized');
			return;
		}

		if (!orbitControlsRef) {
			console.error('[Desktop3D] ❌ OrbitControls not ready yet. Please wait a moment and try again.');
			alert('3D scene is still loading. Please wait a moment and try again.');
			return;
		}

		gestureControlLoading = true;

		try {
			// Create controller
			gestureController = new SimpleGestureController();

			// Set direct camera control callbacks
			gestureController.setCallbacks({
				// ROTATE: Fist drag
				onRotate: (deltaX: number, deltaY: number) => {
					const controls = orbitControlsRef;
					if (!controls) return;

					// Disable auto-rotate while gesturing
					desktop3dStore.setAutoRotate(false);

					// Directly manipulate OrbitControls camera position
					const offset = new THREE.Vector3();
					offset.copy(controls.object.position).sub(controls.target);

					const spherical = new THREE.Spherical();
					spherical.setFromVector3(offset);

					// Apply rotation deltas
					spherical.theta -= deltaX * 1.0;
					spherical.phi -= deltaY * 1.0;

					// Clamp vertical rotation
					spherical.phi = Math.max(0.1, Math.min(Math.PI - 0.1, spherical.phi));

					// Update camera
					offset.setFromSpherical(spherical);
					controls.object.position.copy(controls.target).add(offset);
					controls.update();
				},

				// ZOOM: Pinch
				onZoom: (deltaZ: number) => {
					const controls = orbitControlsRef;
					if (!controls) return;

					// Adjust camera distance
					const currentDistance = controls.object.position.length();
					const newDistance = Math.max(200, Math.min(800, currentDistance + deltaZ));

					// Scale camera position to new distance
					controls.object.position.normalize().multiplyScalar(newDistance);
					controls.update();
				},

				// RESET: Open palm
				onReset: () => {
					const controls = orbitControlsRef;
					if (!controls) return;

					// Reset camera to default
					controls.object.position.set(0, 40, 400);
					controls.target.set(0, 0, 0);
					controls.update();

					// Re-enable auto-rotate
					desktop3dStore.setAutoRotate(true);
				}
			});

			// Initialize with camera permissions
			await gestureController.init(gestureVideoElement);

			gestureControlEnabled = true;
			gestureControlLoading = false;
			console.log('[Desktop3D] ✅ Gesture control enabled');
		} catch (error) {
			console.error('[Desktop3D] ❌ Failed to enable gesture control:', error);

			// Show user-friendly error
			const errorMsg = error instanceof Error ? error.message : 'Unknown error';
			if (errorMsg.includes('Permission denied') || errorMsg.includes('NotAllowedError')) {
				alert('Camera permission denied. Please allow camera access and try again.');
			} else if (errorMsg.includes('NotFoundError') || errorMsg.includes('not found')) {
				alert('No camera found. Please connect a webcam and try again.');
			} else {
				alert(`Failed to enable gestures: ${error instanceof Error ? error.message : 'Unknown error'}`);
			}

			gestureControlEnabled = false;
			gestureControlLoading = false;
		}
	}

	/**
	 * Disable gesture control - cleanup
	 */
	function disableGestureControl() {
		if (gestureController) {
			gestureController.destroy();
			gestureController = null;
		}
		gestureControlEnabled = false;
		console.log('[Desktop3D] Gesture control DISABLED');
	}

	/**
	 * Toggle gesture control on/off
	 */
	async function toggleGestureControl() {
		if (gestureControlEnabled) {
			disableGestureControl();
		} else {
			await enableGestureControl();
		}
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
		// Execute with orchestrator if available
		if (executionOrchestrator) {
			console.log('[Voice] 🤖 Executing with orchestrator');
			executeWithOrchestrator(command);
			return;
		}

		// Fallback to legacy path if orchestrator not initialized
		console.log('[Voice] ⚠️ Using legacy voice command system');

		// For conversations (unknown type), route to AI
		if (command.type === 'unknown') {
			console.log('[Voice] 💬 ROUTING TO AI for conversation');
			handleConversation(command.text);
			return;
		}

		// Give instant acknowledgment for actual commands
		const quickAck = getQuickAck(command.type);
		console.log('[Voice] 🔊 SPEAKING ACK:', quickAck);
		const emotion = emotionalTTS.detectEmotion(quickAck);
		osaVoiceService.speak(quickAck, emotion);

		// Execute command with error handling
		try {
			console.log('[Voice] ⚙️ EXECUTING:', command.type);
			executeCommandAction(command);
			console.log('[Voice] ✅ SUCCESS:', command.type);
		} catch (err) {
			console.error('[Voice] ❌ FAILED:', command.type, err);
			osaVoiceService.speak("Sorry, that didn't work", "empathetic");
		}
	}

	// Voice Agent - Execution-first voice command handler
	async function executeWithOrchestrator(command: VoiceCommand) {
		console.log('[VoiceAgent] 🚀 Executing command:', command.text);

		// Detect emotion from command text
		const emotion = emotionalTTS.detectEmotion(command.text);

		try {
			// Map old VoiceCommand to new VoiceIntent
			const intent = {
				intent: command.text,
				tools: mapCommandToTools(command),
				narration: '', // Orchestrator will generate
				confidence: 0.9
			};

			// Execute via orchestrator (parallel execution + narration)
			const results = await executionOrchestrator.execute(intent);

			console.log('[VoiceV2] ✅ Execution complete:', {
				total: results.length,
				success: results.filter(r => r.success).length,
				failed: results.filter(r => !r.success).length
			});
		} catch (err) {
			console.error('[VoiceV2] ❌ Execution failed:', err);
			osaVoiceService.speak("Sorry, I encountered an error", "empathetic");
		}
	}

	// Helper: Map legacy VoiceCommand to new ToolCall format
	function mapCommandToTools(command: VoiceCommand): Array<{tool: string, params: any}> {
		// Map command type to tool name
		const toolMap: Record<string, string> = {
			'focus_module': 'open_module',
			'close_module': 'close_module',
			'zoom_in': 'zoom_camera',
			'zoom_out': 'zoom_camera',
			'rotate_left': 'rotate_camera',
			'rotate_right': 'rotate_camera',
			'enter_edit_mode': 'layout_enter_edit_mode',
			'exit_edit_mode': 'layout_exit_edit_mode',
			'save_layout': 'layout_save',
			'load_layout': 'layout_load',
			'reset_layout': 'layout_reset',
			// Add more mappings as needed
		};

		const toolName = toolMap[command.type];
		if (!toolName) {
			console.warn('[VoiceV2] No tool mapping for command:', command.type);
			return [];
		}

		// Build params based on command
		const params: any = {};
		if (command.module) params.module_id = command.module;
		if (command.name) params.name = command.name;
		if (command.view) params.view = command.view;
		if (command.type === 'zoom_in') params.direction = 'in';
		if (command.type === 'zoom_out') params.direction = 'out';
		if (command.type === 'rotate_left') params.direction = 'left';
		if (command.type === 'rotate_right') params.direction = 'right';

		return [{ tool: toolName, params }];
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
		const startTime = performance.now();
		console.log('[Conversation] 🚀 START:', text);
		try {
			// Rate limiting - prevent rapid-fire requests
			const now = Date.now();
			if (now - lastRequestTime < REQUEST_COOLDOWN) {
				console.log('[Conversation] ⏱️ Rate limited');
				return;
			}
			lastRequestTime = now;

			// Build OSA personality prompt with context
			const currentModule = $focusedWindow?.module || 'none';
			const openModules = $openWindows.map(w => w.module).join(', ') || 'none';
			const viewMode = $desktop3dStore.viewMode;

			// VOICE AGENT: Prepend ultra-short instructions to the message itself
			// (Backend ignores system_prompt field, so we put it in the message)
			const voiceInstruction = `[VOICE MODE - Reply in 5-10 words ONLY. Be casual like texting. No formalities.]`;
			const messageWithInstruction = `${voiceInstruction}\n\nUser: ${text}`;

			// Only keep last 2 exchanges for speed (4 messages max)
			conversationHistory.push({ role: 'user', content: text });
			if (conversationHistory.length > 4) {
				conversationHistory = conversationHistory.slice(-4);
			}

			// Fast API call - minimal payload with FAST model for voice
			console.log('[Conversation] 📤 Sending with fast model...');
			const response = await fetch('/api/chat/message', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					message: messageWithInstruction,
					context: 'voice_desktop_3d',
					stream: true,
					conversation_id: conversationId,
					model: 'llama-3.1-8b-instant', // FAST model for voice (8B vs 70B)
					max_tokens: 30, // Very short responses only
					temperature: 0.5, // More consistent
					output_style: 'concise' // Request concise output
				})
			});

			const apiTime = performance.now() - startTime;
			console.log(`[Conversation] 📥 Response: ${response.status} (${Math.round(apiTime)}ms)`);
			if (!response.ok) {
				console.error('[Conversation] ❌ API error:', response.status);
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

				// VALIDATION: Detect if user requested an action but AI didn't include command marker
				const userText = text.toLowerCase();
				const actionKeywords = [
					'open', 'close', 'launch', 'start', 'stop', 'zoom', 'rotate',
					'switch', 'change', 'move', 'reset', 'expand', 'contract',
					'minimize', 'maximize', 'save', 'load', 'enter', 'exit'
				];
				const userRequestedAction = actionKeywords.some(keyword => userText.includes(keyword));

				// Parse and execute any commands from the response
				const cmdMatch = response.match(/\[CMD:([^\]]+)\]/);

				// CRITICAL FIX: If user requested action but AI didn't include command, force it
				if (userRequestedAction && !cmdMatch) {
					console.warn('[Voice] ⚠️ AI failed to include command marker for action request!');
					console.warn('[Voice] User said:', text);
					console.warn('[Voice] AI said:', response);

					// Try to parse the user's command directly
					const userCommand = voiceCommandParser.parse(text);
					if (userCommand.type !== 'unknown') {
						console.log('[Voice] 🔧 Auto-fixing: Executing user command directly:', userCommand.type);
						executeCommandAction(userCommand);

						// Simplify the AI's response to just acknowledgment
						response = "On it.";
					} else {
						console.error('[Voice] ❌ Could not parse user command:', text);
						response = "Sorry, I didn't catch that. Can you try again?";
					}
				} else if (cmdMatch) {
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

	// Voice cleanup removed - handled globally by +layout.svelte
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
				bind:orbitControlsRef={orbitControlsRef}
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
		onOpenAppRegistry={() => {
			console.log('[Desktop3D] onOpenAppRegistry callback triggered - using store');
			openAppRegistry();
		}}
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

	<!-- App Registry Modal is now rendered by MenuBar using global store -->

	<!-- Voice Orb + Captions are now GLOBAL - rendered in +layout.svelte -->

	<!-- Hidden video element for gesture camera (MediaPipe) -->
	<!-- svelte-ignore a11y-media-has-caption -->
	<video
		bind:this={gestureVideoElement}
		style="position: absolute; opacity: 0; pointer-events: none; width: 1px; height: 1px;"
		autoplay
		playsinline
		muted
	></video>

	<!-- Gesture Control Toggle Button -->
	<button
		onclick={toggleGestureControl}
		class="gesture-toggle-btn"
		class:active={gestureControlEnabled}
		class:loading={gestureControlLoading}
		disabled={gestureControlLoading}
		title={gestureControlLoading ? 'Initializing...' : gestureControlEnabled ? 'Disable Gesture Control' : 'Enable Gesture Control'}
	>
		{#if gestureControlLoading}
			<!-- Loading spinner -->
			<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" class="spinner">
				<circle cx="12" cy="12" r="10" stroke-width="3" stroke-dasharray="50" stroke-dashoffset="0" />
			</svg>
			<span class="btn-label">Loading...</span>
		{:else}
			<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 11.5V14m0-2.5v-6a1.5 1.5 0 113 0m-3 6a1.5 1.5 0 00-3 0v2a7.5 7.5 0 0015 0v-5a1.5 1.5 0 00-3 0m-6-3V11m0-5.5v-1a1.5 1.5 0 013 0v1m0 0V11m0-5.5a1.5 1.5 0 013 0v3m0 0V11" />
			</svg>
			{#if !gestureControlEnabled}
				<span class="btn-label">Gestures</span>
			{:else}
				<span class="btn-label">ON</span>
			{/if}
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

	.gesture-toggle-btn.loading {
		background: rgba(30, 30, 35, 0.95);
		cursor: wait;
		opacity: 0.8;
	}

	.gesture-toggle-btn .spinner {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
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
