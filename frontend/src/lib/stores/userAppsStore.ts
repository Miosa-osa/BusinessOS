import { writable } from 'svelte/store';
import { windowStore } from './windowStore';

export interface UserApp {
	id: string;
	user_id: string;
	workspace_id: string;
	name: string;
	url: string;
	icon: string; // Lucide icon name (deprecated - use logo_url)
	color: string; // Hex color
	logo_url?: string | null; // URL to actual app logo/favicon
	category?: string | null;
	description?: string | null;
	position_x?: number | null;
	position_y?: number | null;
	position_z?: number | null;
	iframe_config: Record<string, any>;
	is_active?: boolean | null;
	open_on_startup?: boolean | null;
	app_type: string; // 'web' or 'native'
	created_at: string;
	updated_at: string;
	last_opened_at?: string | null;
}

export interface CreateUserAppParams {
	workspace_id: string;
	name: string;
	url: string;
	icon?: string;
	color?: string;
	logo_url?: string; // Optional - will be auto-fetched if not provided
	category?: string;
	description?: string;
	iframe_config?: Record<string, any>;
	open_on_startup?: boolean;
	app_type?: string;
}

export interface UpdateUserAppParams {
	name?: string;
	url?: string;
	icon?: string;
	color?: string;
	logo_url?: string;
	category?: string;
	description?: string;
	position_x?: number;
	position_y?: number;
	position_z?: number;
	iframe_config?: Record<string, any>;
	is_active?: boolean;
	open_on_startup?: boolean;
}

interface UserAppsState {
	apps: UserApp[];
	loading: boolean;
	error: string | null;
}

function createUserAppsStore() {
	const { subscribe, set, update } = writable<UserAppsState>({
		apps: [],
		loading: false,
		error: null
	});

	const API_BASE = '/api/user-apps';

	return {
		subscribe,

		/**
		 * Fetch all user apps for a workspace
		 */
		async fetch(workspaceId: string, includeInactive = false): Promise<void> {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const params = new URLSearchParams({
					workspace_id: workspaceId
				});
				if (includeInactive) {
					params.set('include_inactive', 'true');
				}

				const response = await fetch(`${API_BASE}?${params}`, {
					credentials: 'include'
				});
				if (!response.ok) {
					throw new Error(`Failed to fetch apps: ${response.statusText}`);
				}

				const data = await response.json();
				const apps = data.apps || [];

				update((state) => ({
					...state,
					apps,
					loading: false
				}));

				// Register all apps with windowStore for desktop icons
				apps.forEach((app: UserApp) => {
					windowStore.registerUserApp({
						id: app.id,
						name: app.name,
						url: app.url,
						icon: app.icon,
						color: app.color,
						logo_url: app.logo_url
					});
				});
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : 'Failed to fetch apps';

				// In dev mode, fall back to mock apps
				const isDev = typeof window !== 'undefined' &&
					(window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');

				if (isDev) {
					console.log('[UserApps] API failed in dev mode, loading mock apps');
					const mockApps = getMockUserApps(workspaceId);
					update((state) => ({
						...state,
						apps: mockApps,
						loading: false,
						error: null
					}));

					// Register mock apps with windowStore
					mockApps.forEach((app: UserApp) => {
						windowStore.registerUserApp({
							id: app.id,
							name: app.name,
							url: app.url,
							icon: app.icon,
							color: app.color,
							logo_url: app.logo_url
						});
					});
					return;
				}

				update((state) => ({
					...state,
					loading: false,
					error: errorMessage
				}));
				throw error;
			}
		},

		/**
		 * Get a specific user app
		 */
		async get(appId: string, workspaceId: string): Promise<UserApp> {
			const params = new URLSearchParams({ workspace_id: workspaceId });
			const response = await fetch(`${API_BASE}/${appId}?${params}`, {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`Failed to fetch app: ${response.statusText}`);
			}

			const data = await response.json();
			return data.app;
		},

		/**
		 * Create a new user app
		 */
		async create(params: CreateUserAppParams): Promise<UserApp> {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const response = await fetch(API_BASE, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					credentials: 'include',
					body: JSON.stringify(params)
				});

				if (!response.ok) {
					throw new Error(`Failed to create app: ${response.statusText}`);
				}

				const data = await response.json();
				const newApp = data.app;

				update((state) => ({
					...state,
					apps: [...state.apps, newApp],
					loading: false
				}));

				// Register app with windowStore for desktop icon
				windowStore.registerUserApp({
					id: newApp.id,
					name: newApp.name,
					url: newApp.url,
					icon: newApp.icon,
					color: newApp.color,
					logo_url: newApp.logo_url
				});

				return newApp;
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : 'Failed to create app';
				update((state) => ({
					...state,
					loading: false,
					error: errorMessage
				}));
				throw error;
			}
		},

		/**
		 * Update an existing user app
		 */
		async update(appId: string, workspaceId: string, params: UpdateUserAppParams): Promise<UserApp> {
			const queryParams = new URLSearchParams({ workspace_id: workspaceId });

			const response = await fetch(`${API_BASE}/${appId}?${queryParams}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify(params)
			});

			if (!response.ok) {
				throw new Error(`Failed to update app: ${response.statusText}`);
			}

			const data = await response.json();
			const updatedApp = data.app;

			// Update in store
			update((state) => ({
				...state,
				apps: state.apps.map((app) => (app.id === appId ? updatedApp : app))
			}));

			return updatedApp;
		},

		/**
		 * Delete a user app
		 */
		async delete(appId: string, workspaceId: string): Promise<void> {
			const params = new URLSearchParams({ workspace_id: workspaceId });

			const response = await fetch(`${API_BASE}/${appId}?${params}`, {
				method: 'DELETE',
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`Failed to delete app: ${response.statusText}`);
			}

			// Remove from store
			update((state) => ({
				...state,
				apps: state.apps.filter((app) => app.id !== appId)
			}));

			// Unregister from windowStore
			windowStore.unregisterUserApp(appId);
		},

		/**
		 * Update app position in 3D desktop
		 */
		async updatePosition(
			appId: string,
			position: { position_x: number; position_y: number; position_z: number }
		): Promise<void> {
			const response = await fetch(`${API_BASE}/${appId}/position`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify(position)
			});

			if (!response.ok) {
				throw new Error(`Failed to update position: ${response.statusText}`);
			}

			// Update in store (optimistic)
			update((state) => ({
				...state,
				apps: state.apps.map((app) =>
					app.id === appId
						? {
								...app,
								position_x: position.position_x,
								position_y: position_y,
								position_z: position.position_z
							}
						: app
				)
			}));
		},

		/**
		 * Record that an app was opened (updates last_opened_at)
		 */
		async recordOpened(appId: string): Promise<void> {
			const response = await fetch(`${API_BASE}/${appId}/open`, {
				method: 'POST',
				credentials: 'include'
			});

			if (!response.ok) {
				console.warn(`Failed to record app opened: ${response.statusText}`);
			}

			// Update in store (optimistic)
			update((state) => ({
				...state,
				apps: state.apps.map((app) =>
					app.id === appId
						? {
								...app,
								last_opened_at: new Date().toISOString()
							}
						: app
				)
			}));
		},

		/**
		 * Get apps configured to open on startup
		 */
		async fetchStartupApps(workspaceId: string): Promise<UserApp[]> {
			const params = new URLSearchParams({ workspace_id: workspaceId });
			const response = await fetch(`${API_BASE}/startup?${params}`, {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`Failed to fetch startup apps: ${response.statusText}`);
			}

			const data = await response.json();
			return data.apps || [];
		},

		/**
		 * Clear error state
		 */
		clearError() {
			update((state) => ({ ...state, error: null }));
		},

		/**
		 * Reset store to initial state
		 */
		reset() {
			set({
				apps: [],
				loading: false,
				error: null
			});
		}
	};
}

export const userAppsStore = createUserAppsStore();

/**
 * Mock user apps for dev mode when API is unavailable
 */
function getMockUserApps(workspaceId: string): UserApp[] {
	const now = new Date().toISOString();
	return [
		{
			id: 'mock-app-1',
			user_id: 'mock-user',
			workspace_id: workspaceId,
			name: 'Notion',
			url: 'https://notion.so',
			icon: 'FileText',
			color: '#000000',
			logo_url: 'https://www.notion.so/images/favicon.ico',
			category: 'productivity',
			description: 'Notes and documentation',
			iframe_config: {},
			is_active: true,
			app_type: 'web',
			created_at: now,
			updated_at: now
		},
		{
			id: 'mock-app-2',
			user_id: 'mock-user',
			workspace_id: workspaceId,
			name: 'Figma',
			url: 'https://figma.com',
			icon: 'Pen',
			color: '#F24E1E',
			logo_url: 'https://static.figma.com/app/icon/1/favicon.ico',
			category: 'design',
			description: 'Design and prototyping',
			iframe_config: {},
			is_active: true,
			app_type: 'web',
			created_at: now,
			updated_at: now
		},
		{
			id: 'mock-app-3',
			user_id: 'mock-user',
			workspace_id: workspaceId,
			name: 'Linear',
			url: 'https://linear.app',
			icon: 'CheckSquare',
			color: '#5E6AD2',
			logo_url: 'https://linear.app/favicon.ico',
			category: 'project',
			description: 'Issue tracking',
			iframe_config: {},
			is_active: true,
			app_type: 'web',
			created_at: now,
			updated_at: now
		},
		{
			id: 'mock-app-4',
			user_id: 'mock-user',
			workspace_id: workspaceId,
			name: 'Slack',
			url: 'https://slack.com',
			icon: 'MessageSquare',
			color: '#4A154B',
			logo_url: 'https://a.slack-edge.com/80588/marketing/img/meta/favicon-32.png',
			category: 'communication',
			description: 'Team communication',
			iframe_config: {},
			is_active: true,
			app_type: 'web',
			created_at: now,
			updated_at: now
		},
		{
			id: 'mock-app-5',
			user_id: 'mock-user',
			workspace_id: workspaceId,
			name: 'GitHub',
			url: 'https://github.com',
			icon: 'GitBranch',
			color: '#181717',
			logo_url: 'https://github.com/favicon.ico',
			category: 'development',
			description: 'Code repository',
			iframe_config: {},
			is_active: true,
			app_type: 'web',
			created_at: now,
			updated_at: now
		}
	];
}
