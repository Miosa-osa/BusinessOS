/**
 * appsStore.ts
 * Manages user's mini-apps (their Operating System modules)
 */

import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

export interface MiniApp {
	id: string;
	title: string;
	description: string;
	iconUrl?: string;
	iconPrompt?: string;
	visibility: 'public' | 'private';
	components: AppComponent[];
	usagePercentage: number;
	remixedFrom?: string; // App ID if remixed from another
	isStarterApp: boolean;
	stats: {
		likes: number;
		comments: number;
		remixes: number;
		gets: number;
	};
	createdAt: Date;
	updatedAt: Date;
}

export interface AppComponent {
	id: string;
	name: string;
	type: 'text' | 'image' | 'code' | 'data';
	prompt: string;
	model: {
		provider: 'openai' | 'anthropic' | 'google' | 'auto';
		name: string;
		config?: {
			temperature?: number;
			maxTokens?: number;
		};
	};
}

export interface AppsState {
	apps: MiniApp[];
	loading: boolean;
	error: string | null;
	selectedAppId: string | null;
}

// Default state
const defaultState: AppsState = {
	apps: [],
	loading: false,
	error: null,
	selectedAppId: null
};

// Load initial state from localStorage
function loadState(): AppsState {
	if (browser) {
		try {
			const stored = localStorage.getItem('osa_apps_state');
			if (stored) {
				const parsed = JSON.parse(stored);
				// Convert date strings back to Date objects
				if (parsed.apps) {
					parsed.apps = parsed.apps.map((app: any) => ({
						...app,
						createdAt: new Date(app.createdAt),
						updatedAt: new Date(app.updatedAt)
					}));
				}
				return parsed;
			}
		} catch (e) {
			console.error('Error loading apps state:', e);
		}
	}
	return defaultState;
}

// Create the store
function createAppsStore() {
	const { subscribe, set, update } = writable<AppsState>(loadState());

	return {
		subscribe,

		// Set all apps
		setApps: (apps: MiniApp[]) => update(state => {
			const newState = { ...state, apps, error: null };
			saveState(newState);
			return newState;
		}),

		// Add a new app
		addApp: (app: MiniApp) => update(state => {
			const newState = {
				...state,
				apps: [...state.apps, app],
				error: null
			};
			saveState(newState);
			return newState;
		}),

		// Update an existing app
		updateApp: (id: string, updates: Partial<MiniApp>) => update(state => {
			const newState = {
				...state,
				apps: state.apps.map(app =>
					app.id === id
						? { ...app, ...updates, updatedAt: new Date() }
						: app
				),
				error: null
			};
			saveState(newState);
			return newState;
		}),

		// Delete an app
		deleteApp: (id: string) => update(state => {
			const newState = {
				...state,
				apps: state.apps.filter(app => app.id !== id),
				selectedAppId: state.selectedAppId === id ? null : state.selectedAppId,
				error: null
			};
			saveState(newState);
			return newState;
		}),

		// Select an app
		selectApp: (id: string | null) => update(state => {
			const newState = { ...state, selectedAppId: id };
			saveState(newState);
			return newState;
		}),

		// Set loading state
		setLoading: (loading: boolean) => update(state => ({ ...state, loading })),

		// Set error
		setError: (error: string | null) => update(state => ({ ...state, error })),

		// Clear all apps (for logout)
		clear: () => {
			set(defaultState);
			if (browser) {
				localStorage.removeItem('osa_apps_state');
			}
		}
	};
}

// Save state to localStorage
function saveState(state: AppsState) {
	if (browser) {
		try {
			localStorage.setItem('osa_apps_state', JSON.stringify(state));
		} catch (e) {
			console.error('Error saving apps state:', e);
		}
	}
}

export const appsStore = createAppsStore();

// Derived store: Public apps only
export const publicApps = derived(
	appsStore,
	$apps => $apps.apps.filter(app => app.visibility === 'public')
);

// Derived store: Private apps only
export const privateApps = derived(
	appsStore,
	$apps => $apps.apps.filter(app => app.visibility === 'private')
);

// Derived store: Starter apps
export const starterApps = derived(
	appsStore,
	$apps => $apps.apps.filter(app => app.isStarterApp)
);

// Derived store: Currently selected app
export const selectedApp = derived(
	appsStore,
	$apps => $apps.apps.find(app => app.id === $apps.selectedAppId) || null
);

// Derived store: App count
export const appCount = derived(
	appsStore,
	$apps => $apps.apps.length
);

// Derived store: Most used apps (top 5)
export const mostUsedApps = derived(
	appsStore,
	$apps => [...$apps.apps]
		.sort((a, b) => b.usagePercentage - a.usagePercentage)
		.slice(0, 5)
);
