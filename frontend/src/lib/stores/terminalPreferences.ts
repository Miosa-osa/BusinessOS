import { writable } from 'svelte/store';

export type TerminalMode = 'docker' | 'local';

interface TerminalPreferences {
	defaultMode: TerminalMode;
	hasSeenLocalWarning: boolean;
}

const DEFAULT_PREFS: TerminalPreferences = {
	defaultMode: 'docker',
	hasSeenLocalWarning: false
};

function createTerminalPreferences() {
	const stored = localStorage.getItem('terminal-preferences');
	const initial = stored ? JSON.parse(stored) : DEFAULT_PREFS;

	const { subscribe, set, update } = writable<TerminalPreferences>(initial);

	return {
		subscribe,
		setDefaultMode: (mode: TerminalMode) =>
			update((p) => {
				const updated = { ...p, defaultMode: mode };
				localStorage.setItem('terminal-preferences', JSON.stringify(updated));
				return updated;
			}),
		markWarningShown: () =>
			update((p) => {
				const updated = { ...p, hasSeenLocalWarning: true };
				localStorage.setItem('terminal-preferences', JSON.stringify(updated));
				return updated;
			}),
		reset: () => {
			localStorage.removeItem('terminal-preferences');
			set(DEFAULT_PREFS);
		}
	};
}

export const terminalPreferences = createTerminalPreferences();
