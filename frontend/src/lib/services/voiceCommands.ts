/**
 * Voice Commands Parser
 *
 * Parses natural language voice input into structured commands
 * for controlling the 3D Desktop.
 *
 * Supported Commands:
 * - Layout Management: "edit layout", "save layout as X", "load layout X"
 * - Module Navigation: "open chat", "focus dashboard", "show tasks"
 * - View Control: "switch to grid", "switch to orb", "toggle auto-rotate"
 * - General: "help", "what can I say"
 */

import type { ModuleId } from '$lib/stores/desktop3dStore';

export type VoiceCommand =
	| { type: 'enter_edit_mode' }
	| { type: 'exit_edit_mode' }
	| { type: 'save_layout'; name: string }
	| { type: 'load_layout'; name: string }
	| { type: 'delete_layout'; name: string }
	| { type: 'reset_layout' }
	| { type: 'open_layout_manager' }
	| { type: 'focus_module'; module: ModuleId }
	| { type: 'open_module'; module: ModuleId }
	| { type: 'close_module'; module: ModuleId }
	| { type: 'close_all_windows' }
	| { type: 'minimize_window' }
	| { type: 'maximize_window' }
	| { type: 'unfocus' }
	| { type: 'resize_window'; direction: 'wider' | 'narrower' | 'taller' | 'shorter' }
	| { type: 'switch_view'; view: 'orb' | 'grid' }
	| { type: 'toggle_auto_rotate' }
	| { type: 'rotate_left' }
	| { type: 'rotate_right' }
	| { type: 'stop_rotation' }
	| { type: 'rotate_faster' }
	| { type: 'rotate_slower' }
	| { type: 'zoom_in' }
	| { type: 'zoom_out' }
	| { type: 'reset_zoom' }
	| { type: 'expand_orb' }
	| { type: 'contract_orb' }
	| { type: 'increase_grid_spacing' }
	| { type: 'decrease_grid_spacing' }
	| { type: 'more_grid_columns' }
	| { type: 'less_grid_columns' }
	| { type: 'next_window' }
	| { type: 'previous_window' }
	| { type: 'help' }
	| { type: 'unknown'; text: string };

export class VoiceCommandParser {
	/**
	 * Detect and strip wake word ("OSA")
	 */
	private stripWakeWord(text: string): { stripped: string; hadWakeWord: boolean } {
		const wakeWords = ['osa', 'hey osa', 'ok osa'];

		for (const wake of wakeWords) {
			const pattern = new RegExp(`^${wake}[,.]?\\s+`, 'i');
			if (pattern.test(text)) {
				return {
					stripped: text.replace(pattern, '').trim(),
					hadWakeWord: true
				};
			}
		}

		return { stripped: text, hadWakeWord: false };
	}

	/**
	 * Extract core command from conversational wrappers
	 */
	private extractCommand(text: string): { extracted: string; confidence: number } {
		const wrappers = [
			{ pattern: /^(can you|could you|please|can i|i want to)\s+/i, confidence: 0.9 },
			{ pattern: /^(hey|hi|hello),?\s+/i, confidence: 0.8 },
			{ pattern: /^(would you|will you)\s+/i, confidence: 0.85 }
		];

		let extracted = text;
		let confidence = 1.0;

		for (const wrapper of wrappers) {
			if (wrapper.pattern.test(extracted)) {
				extracted = extracted.replace(wrapper.pattern, '');
				confidence *= wrapper.confidence;
			}
		}

		// Strip trailing politeness
		extracted = extracted.replace(/\s+(please|thanks|thank you)[\.\?]?$/i, '');

		return { extracted: extracted.trim(), confidence };
	}

	/**
	 * Detect if text contains a module name (fuzzy)
	 */
	private detectModule(text: string): { module: ModuleId | null; confidence: number } {
		const modules: ModuleId[] = [
			'dashboard', 'chat', 'tasks', 'projects', 'team', 'clients',
			'tables', 'communication', 'pages', 'nodes', 'daily',
			'settings', 'terminal', 'help', 'agents', 'crm',
			'integrations', 'knowledge-v2', 'notifications', 'profile',
			'voice-notes', 'usage'
		];

		// Check for exact word boundary match
		for (const module of modules) {
			const moduleName = module.replace(/-/g, ' '); // "knowledge-v2" → "knowledge v2"
			const regex = new RegExp(`\\b${this.escapeRegex(moduleName)}\\b`, 'i');
			if (regex.test(text)) {
				return { module, confidence: 0.95 };
			}
		}

		// Check for partial match (less confident)
		for (const module of modules) {
			if (text.includes(module)) {
				return { module, confidence: 0.7 };
			}
		}

		return { module: null, confidence: 0 };
	}

	/**
	 * Try matching against all pattern categories
	 */
	private tryAllPatterns(text: string): VoiceCommand | null {
		// Try each category in order
		let result = this.parseLayoutCommand(text);
		if (result) return result;

		result = this.parseModuleCommand(text);
		if (result) return result;

		result = this.parseResizeCommand(text);
		if (result) return result;

		result = this.parseViewCommand(text);
		if (result) return result;

		result = this.parseNavigationCommand(text);
		if (result) return result;

		return null;
	}

	/**
	 * Parse transcript text into a voice command
	 */
	parse(transcript: string): VoiceCommand {
		const lower = transcript.toLowerCase().trim();
		console.log('[Parser] 🔍 ANALYZING:', transcript);

		// LAYER 1: Strip wake word
		const { stripped: afterWake, hadWakeWord } = this.stripWakeWord(lower);
		if (hadWakeWord) {
			console.log('[Parser] 👂 Wake word detected, stripped to:', afterWake);
		}

		const normalized = this.normalize(afterWake);
		console.log('[Parser] 📝 Normalized:', normalized);

		// LAYER 2: Try exact pattern matching first (highest confidence)
		const exactMatch = this.tryAllPatterns(normalized);
		if (exactMatch) {
			console.log('[Parser] ✅ EXACT MATCH:', exactMatch.type);
			return exactMatch;
		}

		// LAYER 3: Extract command from conversational wrapper
		const { extracted, confidence } = this.extractCommand(normalized);
		console.log('[Parser] 🧠 Extracted command:', { extracted, confidence });

		if (extracted !== normalized && confidence > 0.7) {
			const extractedMatch = this.tryAllPatterns(extracted);
			if (extractedMatch) {
				console.log('[Parser] ✅ EXTRACTED MATCH:', extractedMatch.type);
				return extractedMatch;
			}
		}

		// LAYER 4: Fuzzy module detection - DISABLED
		// This was too aggressive - mentioning "dashboard" in conversation shouldn't open it
		// Module commands now require explicit action words (open, show, etc.)
		// const { module, confidence: moduleConfidence } = this.detectModule(normalized);
		// console.log('[Parser] 🔎 Module detection:', { module, confidence: moduleConfidence });
		// if (module && moduleConfidence > 0.7) {
		// 	console.log('[Parser] ✅ FUZZY MODULE MATCH → open:', module);
		// 	return { type: 'focus_module', module };
		// }

		// LAYER 5: Help intent detection - STRICT command patterns only
		// Only trigger for explicit help commands, NOT conversational questions
		const helpCommandPatterns = [
			/^help$/i,                    // Just "help"
			/^show\s+help$/i,             // "show help"
			/^open\s+help$/i,             // "open help"
			/^display\s+help$/i,          // "display help"
			/^what\s+can\s+i\s+say$/i,    // "what can I say" (exact)
			/^show\s+commands$/i,         // "show commands"
			/^list\s+commands$/i          // "list commands"
		];

		const isHelpCommand = helpCommandPatterns.some(pattern => pattern.test(normalized.trim()));
		if (isHelpCommand) {
			console.log('[Parser] ✅ HELP COMMAND (explicit)');
			return { type: 'help' };
		}

		// LAYER 6: Route to conversation (default for questions and long phrases)
		const isQuestion = transcript.includes('?');
		const wordCount = transcript.trim().split(/\s+/).length;
		const isConv = this.isConversational(normalized);

		console.log('[Parser] 🤔 Routing decision:', {
			wordCount,
			isQuestion,
			isConversational: isConv,
			hasModule: !!module
		});

		// Default to AI conversation for:
		// - Questions (contains ?)
		// - Long phrases (7+ words)
		// - Conversational language
		if (isQuestion || wordCount > 7 || isConv) {
			console.log('[Parser] 💬 ROUTING TO AI (conversational)');
			return { type: 'unknown', text: transcript };
		}

		// If short phrase and not matched, might be unclear command
		console.log('[Parser] ❓ UNKNOWN (possible unclear command)');
		return { type: 'unknown', text: transcript };
	}

	/**
	 * Check if text is conversational (not a command)
	 * Updated to be less aggressive - checks for action words first
	 */
	private isConversational(text: string): boolean {
		// Only flag as conversational if it contains greeting/question phrases
		// WITHOUT action words
		const conversationalPhrases = [
			'how are you', 'how are', 'what are you',
			'tell me about', 'explain',
			'i need help', 'i have a question',
			'good morning', 'good afternoon', 'good evening'
		];

		// Don't flag simple polite prefixes as conversational
		// (they're handled by extractCommand)
		const actionWords = ['open', 'close', 'show', 'hide', 'zoom', 'expand', 'contract', 'switch', 'go', 'focus', 'load', 'save'];
		const hasActionWord = actionWords.some(word => text.includes(word));

		if (hasActionWord) {
			return false; // It's a command, not conversation
		}

		return conversationalPhrases.some(phrase => text.includes(phrase));
	}

	/**
	 * Normalize transcript (remove filler words, fix common transcription errors)
	 */
	private normalize(text: string): string {
		// Remove filler words
		const fillers = ['um', 'uh', 'like', 'you know', 'i mean', 'basically'];
		let normalized = text;
		for (const filler of fillers) {
			normalized = normalized.replace(new RegExp(`\\b${filler}\\b`, 'gi'), '');
		}

		// Fix common transcription errors
		const corrections: Record<string, string> = {
			'lay out': 'layout',
			'edit mode': 'edit mode',
			'auto rotate': 'auto-rotate',
			'auto rotation': 'auto-rotate'
		};

		for (const [wrong, right] of Object.entries(corrections)) {
			normalized = normalized.replace(new RegExp(wrong, 'gi'), right);
		}

		return normalized.trim();
	}

	/**
	 * Parse layout management commands
	 */
	private parseLayoutCommand(text: string): VoiceCommand | null {
		// Edit mode
		if (this.matchesPattern(text, ['edit layout', 'edit mode', 'start editing', 'enter edit'])) {
			return { type: 'enter_edit_mode' };
		}

		// Exit edit mode
		if (
			this.matchesPattern(text, [
				'exit edit',
				'stop editing',
				'done editing',
				'cancel edit',
				'leave edit'
			])
		) {
			return { type: 'exit_edit_mode' };
		}

		// Save layout
		const saveMatch = text.match(
			/save\s+(?:the\s+)?layout(?:\s+as)?(?:\s+called)?(?:\s+named)?\s+(.+)/i
		);
		if (saveMatch) {
			const name = this.cleanLayoutName(saveMatch[1]);
			if (name) {
				return { type: 'save_layout', name };
			}
		}

		// Simple "save layout" without name
		if (this.matchesPattern(text, ['save layout', 'save this layout'])) {
			return { type: 'save_layout', name: `Layout ${new Date().toLocaleDateString()}` };
		}

		// Load layout
		const loadMatch = text.match(/(?:load|switch\s+to|open)\s+(?:the\s+)?layout\s+(.+)/i);
		if (loadMatch) {
			const name = this.cleanLayoutName(loadMatch[1]);
			if (name) {
				return { type: 'load_layout', name };
			}
		}

		// Delete layout
		const deleteMatch = text.match(/delete\s+(?:the\s+)?layout\s+(.+)/i);
		if (deleteMatch) {
			const name = this.cleanLayoutName(deleteMatch[1]);
			if (name) {
				return { type: 'delete_layout', name };
			}
		}

		// Open layout manager
		if (
			this.matchesPattern(text, [
				'manage layouts',
				'layout manager',
				'show layouts',
				'open layout manager'
			])
		) {
			return { type: 'open_layout_manager' };
		}

		// Reset layout
		if (
			this.matchesPattern(text, [
				'reset layout',
				'default layout',
				'restore default',
				'reset to default'
			])
		) {
			return { type: 'reset_layout' };
		}

		return null;
	}

	/**
	 * Parse module navigation commands
	 */
	private parseModuleCommand(text: string): VoiceCommand | null {
		const modules: ModuleId[] = [
			'dashboard',
			'chat',
			'tasks',
			'projects',
			'team',
			'clients',
			'tables',
			'communication',
			'pages',
			'nodes',
			'daily',
			'settings',
			'terminal',
			'help',
			'agents',
			'crm',
			'integrations',
			'knowledge-v2',
			'notifications',
			'profile',
			'voice-notes',
			'usage'
		];

		console.log('[Parser] Checking modules for:', text);

		// Open/focus module - ALL PERMUTATIONS
		for (const module of modules) {
			const moduleName = module.replace(/-/g, ' '); // "knowledge-v2" -> "knowledge v2"
			const patterns = [
				// Direct opens
				`open ${moduleName}`,
				`open up ${moduleName}`,
				`open the ${moduleName}`,

				// Show
				`show ${moduleName}`,
				`show me ${moduleName}`,
				`show the ${moduleName}`,

				// Focus
				`focus ${moduleName}`,
				`focus on ${moduleName}`,

				// Go to
				`go ${moduleName}`,
				`go to ${moduleName}`,
				`go to the ${moduleName}`,

				// Switch/Change
				`switch to ${moduleName}`,
				`switch to the ${moduleName}`,
				`change to ${moduleName}`,
				`change to the ${moduleName}`,
				`switch me to ${moduleName}`,

				// Pull up / Bring up
				`pull up ${moduleName}`,
				`pull up the ${moduleName}`,
				`bring up ${moduleName}`,
				`bring up the ${moduleName}`,

				// Load
				`load ${moduleName}`,
				`load the ${moduleName}`,

				// Start
				`start ${moduleName}`,
				`launch ${moduleName}`
			];

			if (this.matchesPattern(text, patterns)) {
				console.log(`[Parser] ✅ Module matched: "${module}"`);
				return { type: 'focus_module', module };
			}
		}

		// Close module - ALL PERMUTATIONS
		for (const module of modules) {
			const moduleName = module.replace(/-/g, ' ');
			const patterns = [
				`close ${moduleName}`,
				`close the ${moduleName}`,
				`hide ${moduleName}`,
				`hide the ${moduleName}`,
				`exit ${moduleName}`,
				`quit ${moduleName}`,
				`shut ${moduleName}`,
				`shut down ${moduleName}`,
				`close down ${moduleName}`
			];

			if (this.matchesPattern(text, patterns)) {
				console.log(`[Parser] ✅ Close module: "${module}"`);
				return { type: 'close_module', module };
			}
		}

		// Close all windows
		if (this.matchesPattern(text, [
			'close all',
			'close all windows',
			'close everything',
			'hide all',
			'hide everything',
			'clear all',
			'clear workspace',
			'clear desktop'
		])) {
			return { type: 'close_all_windows' };
		}

		// Minimize current window
		if (this.matchesPattern(text, [
			'minimize',
			'minimize window',
			'minimize this',
			'hide this window',
			'minimize current'
		])) {
			return { type: 'minimize_window' };
		}

		// Maximize current window
		if (this.matchesPattern(text, [
			'maximize',
			'maximize window',
			'maximize this',
			'full screen',
			'fullscreen',
			'maximize current',
			'make full screen'
		])) {
			return { type: 'maximize_window' };
		}

		return null;
	}

	/**
	 * Parse view control commands
	 */
	private parseViewCommand(text: string): VoiceCommand | null {
		// Switch to orb view
		if (this.matchesPattern(text, ['orb view', 'switch to orb', 'sphere view', 'circular view'])) {
			return { type: 'switch_view', view: 'orb' };
		}

		// Switch to grid view
		if (this.matchesPattern(text, ['grid view', 'switch to grid', 'table view'])) {
			return { type: 'switch_view', view: 'grid' };
		}

		// Toggle auto-rotate
		if (
			this.matchesPattern(text, [
				'toggle auto-rotate',
				'auto rotate',
				'start rotating',
				'toggle rotation'
			])
		) {
			return { type: 'toggle_auto_rotate' };
		}

		// Rotate left
		if (this.matchesPattern(text, [
			'rotate left',
			'turn left',
			'spin left',
			'rotate counter clockwise',
			'counterclockwise'
		])) {
			return { type: 'rotate_left' };
		}

		// Rotate right
		if (this.matchesPattern(text, [
			'rotate right',
			'turn right',
			'spin right',
			'rotate clockwise',
			'clockwise'
		])) {
			return { type: 'rotate_right' };
		}

		// Stop rotation
		if (this.matchesPattern(text, [
			'stop rotating',
			'stop rotation',
			'freeze',
			'pause rotation',
			'halt rotation'
		])) {
			return { type: 'stop_rotation' };
		}

		// Rotate faster
		if (this.matchesPattern(text, [
			'rotate faster',
			'speed up',
			'faster rotation',
			'increase speed'
		])) {
			return { type: 'rotate_faster' };
		}

		// Rotate slower
		if (this.matchesPattern(text, [
			'rotate slower',
			'slow down',
			'slower rotation',
			'decrease speed'
		])) {
			return { type: 'rotate_slower' };
		}

		// Camera Zoom (brings objects closer/farther by adjusting sphere)
		if (this.matchesPattern(text, [
			'zoom in',
			'closer',
			'move closer',
			'come closer',
			'get closer',
			'bring closer'
		])) {
			return { type: 'zoom_in' };
		}

		if (this.matchesPattern(text, [
			'zoom out',
			'farther',
			'move back',
			'go back',
			'back up',
			'move away',
			'pull back'
		])) {
			return { type: 'zoom_out' };
		}

		// Reset zoom
		if (this.matchesPattern(text, [
			'reset zoom',
			'default zoom',
			'normal zoom',
			'reset view',
			'normal view'
		])) {
			return { type: 'reset_zoom' };
		}

		// Orb Expansion (makes orb bigger - windows spread out)
		if (this.matchesPattern(text, [
			'expand',
			'expand orb',
			'expand out',
			'make bigger',
			'bigger',
			'enlarge',
			'spread out',
			'open up'
		])) {
			return { type: 'expand_orb' };
		}

		// Orb Contraction (makes orb smaller - windows come together)
		if (this.matchesPattern(text, [
			'contract',
			'contract orb',
			'unexpand',
			'expand in',
			'expand back',
			'go back',
			'undo expand',
			'make smaller',
			'smaller',
			'shrink',
			'shrink orb',
			'close up',
			'bring together'
		])) {
			return { type: 'contract_orb' };
		}

		// Unfocus window
		if (this.matchesPattern(text, ['unfocus', 'exit focus', 'back to orb', 'back to desktop', 'show all'])) {
			return { type: 'unfocus' };
		}

		// Grid spacing controls
		if (this.matchesPattern(text, [
			'increase spacing',
			'more spacing',
			'spread apart',
			'more space',
			'looser grid'
		])) {
			return { type: 'increase_grid_spacing' };
		}

		if (this.matchesPattern(text, [
			'decrease spacing',
			'less spacing',
			'bring closer',
			'less space',
			'tighter grid',
			'compact grid'
		])) {
			return { type: 'decrease_grid_spacing' };
		}

		// Grid column controls
		if (this.matchesPattern(text, [
			'more columns',
			'increase columns',
			'add columns',
			'more per row'
		])) {
			return { type: 'more_grid_columns' };
		}

		if (this.matchesPattern(text, [
			'less columns',
			'fewer columns',
			'decrease columns',
			'remove columns',
			'less per row'
		])) {
			return { type: 'less_grid_columns' };
		}

		return null;
	}

	/**
	 * Parse window resize commands
	 */
	private parseResizeCommand(text: string): VoiceCommand | null {
		if (this.matchesPattern(text, ['make wider', 'expand width', 'wider', 'widen'])) {
			return { type: 'resize_window', direction: 'wider' };
		}

		if (this.matchesPattern(text, ['make narrower', 'reduce width', 'narrower', 'shrink width'])) {
			return { type: 'resize_window', direction: 'narrower' };
		}

		if (this.matchesPattern(text, ['make taller', 'expand height', 'taller', 'higher', 'increase height'])) {
			return { type: 'resize_window', direction: 'taller' };
		}

		if (this.matchesPattern(text, ['make shorter', 'reduce height', 'shorter', 'lower', 'decrease height'])) {
			return { type: 'resize_window', direction: 'shorter' };
		}

		return null;
	}

	/**
	 * Parse navigation commands
	 */
	private parseNavigationCommand(text: string): VoiceCommand | null {
		// Next window - expanded with ALL natural variations
		if (this.matchesPattern(text, [
			// Window variations
			'next window',
			'next one',
			'next page',
			'next thing',
			'next module',
			'next app',

			// "Go to" variations
			'go to next',
			'go to the next',
			'go to next window',
			'go to the next window',
			'go to next page',
			'go to the next page',
			'go to next one',
			'go to the next one',

			// "Move to" variations
			'move to next',
			'move to the next',
			'move to next window',
			'move to next page',

			// "Switch to" variations
			'switch to next',
			'switch to the next',
			'switch to next window',
			'switch to next page',

			// Short forms
			'next',
			'forward',
			'go forward',
			'move forward',
			'right'
		])) {
			return { type: 'next_window' };
		}

		// Previous window - expanded with ALL natural variations
		if (this.matchesPattern(text, [
			// Window variations
			'previous window',
			'previous one',
			'previous page',
			'previous thing',
			'previous module',
			'previous app',
			'last window',
			'last one',
			'last page',

			// "Go to" variations
			'go to previous',
			'go to the previous',
			'go to previous window',
			'go to the previous window',
			'go to previous page',
			'go to the previous page',
			'go to previous one',
			'go to the previous one',
			'go to last',
			'go to the last',
			'go to last window',
			'go to last page',

			// "Move to" variations
			'move to previous',
			'move to the previous',
			'move to previous window',
			'move to previous page',
			'move to last',
			'move to last window',

			// "Switch to" variations
			'switch to previous',
			'switch to the previous',
			'switch to previous window',
			'switch to previous page',
			'switch to last',

			// "Go back" variations
			'go back',
			'back',
			'previous',
			'go backward',
			'move back',
			'left'
		])) {
			return { type: 'previous_window' };
		}

		return null;
	}

	/**
	 * Check if text matches any of the patterns
	 */
	private matchesPattern(text: string, patterns: string[]): boolean {
		for (const pattern of patterns) {
			// Exact match or word boundary match
			const regex = new RegExp(`\\b${this.escapeRegex(pattern)}\\b`, 'i');
			if (regex.test(text)) {
				return true;
			}
		}
		return false;
	}

	/**
	 * Escape special regex characters
	 */
	private escapeRegex(str: string): string {
		return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
	}

	/**
	 * Clean up layout name from transcript
	 */
	private cleanLayoutName(name: string): string {
		return name
			.replace(/\b(called|named|as)\b/gi, '')
			.trim()
			.replace(/\s+/g, ' '); // Normalize whitespace
	}

	/**
	 * Get help text for available commands
	 */
	getHelpText(): string {
		return `
**Available Voice Commands:**

**Layout Management:**
- "Edit layout" - Enter edit mode
- "Exit edit" - Exit edit mode
- "Save layout as [name]" - Save current layout
- "Load layout [name]" - Switch to a saved layout
- "Manage layouts" - Open layout manager

**Module Navigation:**
- "Open [module]" - Open and focus a module
  Examples: "open chat", "show dashboard", "focus tasks"
- "Close [module]" - Close a module

**View Control:**
- "Switch to orb" / "Switch to grid" - Change view mode
- "Toggle auto-rotate" - Start/stop rotation
- "Zoom in" / "Zoom out" - Adjust zoom level

**Navigation:**
- "Next window" / "Previous window" - Navigate between windows

Say "help" anytime to see this list!
		`.trim();
	}
}

// Export singleton instance
export const voiceCommandParser = new VoiceCommandParser();
