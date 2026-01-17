/**
 * Voice Commands - Stub for local command parsing
 *
 * Most voice processing is now handled by the Python LiveKit agent.
 * This provides minimal local parsing for quick actions.
 */

export interface VoiceCommand {
	type: string;
	text: string;
	module?: string;
	view?: string;
	name?: string;
	direction?: 'wider' | 'narrower' | 'taller' | 'shorter';
}

/**
 * Simple voice command parser for local actions
 */
class VoiceCommandParser {
	parse(text: string): VoiceCommand {
		const normalized = text.toLowerCase().trim();

		// Module commands
		const moduleMatch = normalized.match(/^(?:open|show|go to|launch)\s+(?:the\s+)?(\w+)/);
		if (moduleMatch) {
			const module = this.normalizeModuleName(moduleMatch[1]);
			if (module) {
				return { type: 'focus_module', text, module };
			}
		}

		const closeMatch = normalized.match(/^close\s+(?:the\s+)?(\w+)/);
		if (closeMatch) {
			const module = this.normalizeModuleName(closeMatch[1]);
			if (module) {
				return { type: 'close_module', text, module };
			}
		}

		// View commands
		if (normalized.includes('orb view') || normalized.includes('sphere view')) {
			return { type: 'switch_view', text, view: 'orb' };
		}
		if (normalized.includes('grid view') || normalized.includes('flat view')) {
			return { type: 'switch_view', text, view: 'grid' };
		}

		// Navigation
		if (normalized.includes('next window') || normalized.includes('next')) {
			return { type: 'next_window', text };
		}
		if (normalized.includes('previous window') || normalized.includes('previous') || normalized.includes('back')) {
			return { type: 'previous_window', text };
		}

		// Camera
		if (normalized.includes('zoom in') || normalized.includes('closer')) {
			return { type: 'zoom_in', text };
		}
		if (normalized.includes('zoom out') || normalized.includes('farther')) {
			return { type: 'zoom_out', text };
		}
		if (normalized.includes('reset zoom')) {
			return { type: 'reset_zoom', text };
		}

		// Rotation
		if (normalized.includes('start rotating') || normalized.includes('auto rotate')) {
			return { type: 'toggle_auto_rotate', text };
		}
		if (normalized.includes('stop rotating') || normalized.includes('stop rotation')) {
			return { type: 'stop_rotation', text };
		}

		// Close all
		if (normalized.includes('close all') || normalized.includes('close everything')) {
			return { type: 'close_all_windows', text };
		}

		// Help
		if (normalized === 'help' || normalized.includes('what can you do')) {
			return { type: 'help', text };
		}

		// Unknown - route to AI
		return { type: 'unknown', text };
	}

	private normalizeModuleName(name: string): string | null {
		const moduleMap: Record<string, string> = {
			'dashboard': 'dashboard',
			'chat': 'chat',
			'tasks': 'tasks',
			'task': 'tasks',
			'projects': 'projects',
			'project': 'projects',
			'team': 'team',
			'clients': 'clients',
			'client': 'clients',
			'crm': 'crm',
			'tables': 'tables',
			'table': 'tables',
			'pages': 'pages',
			'page': 'pages',
			'agents': 'agents',
			'agent': 'agents',
			'nodes': 'nodes',
			'node': 'nodes',
			'settings': 'settings',
			'setting': 'settings',
			'daily': 'daily-log',
			'log': 'daily-log',
			'usage': 'usage',
			'integrations': 'integrations',
			'integration': 'integrations',
			'communication': 'communication',
			'messages': 'communication',
		};

		return moduleMap[name] || null;
	}
}

export const voiceCommandParser = new VoiceCommandParser();
