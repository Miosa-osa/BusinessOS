/**
 * Desktop 3D Layout Store
 *
 * Manages custom 3D Desktop layouts with backend persistence.
 *
 * Features:
 * - Default layout (5-ring geodesic) always available and immutable
 * - Create/save custom layouts with module positions
 * - Load saved layouts from backend
 * - Activate layouts (marks as active in database)
 * - Delete custom layouts
 * - Edit mode for dragging modules
 *
 * Architecture:
 * - Communicates with /api/desktop3d/layouts endpoints
 * - Syncs with desktop3dStore for module positions
 * - Persists to PostgreSQL via backend API
 */

import { writable, derived, get } from 'svelte/store';
import type { ModuleId } from './desktop3dStore';
import { desktop3dStore } from './desktop3dStore';

export type LayoutType = 'default' | 'custom';

export interface ModulePosition {
	module_id: ModuleId;
	position: { x: number; y: number; z: number };
	rotation: { x: number; y: number; z: number };
	scale: number;
}

export interface Layout {
	id: string;
	name: string;
	type: LayoutType;
	created_at: Date;
	updated_at: Date;
	is_active: boolean;
	user_id: string;
	modules: ModulePosition[];
}

interface LayoutState {
	layouts: Layout[];
	activeLayoutId: string;
	editMode: boolean;
	loading: boolean;
	error: string | null;
}

const initialState: LayoutState = {
	layouts: [],
	activeLayoutId: 'default',
	editMode: false,
	loading: false,
	error: null
};

/**
 * Get API base URL
 */
function getApiBase(): string {
	if (typeof window === 'undefined') return '';

	const isElectron = 'electron' in window;
	if (isElectron) {
		const mode = localStorage.getItem('businessos_mode');
		const cloudUrl = localStorage.getItem('businessos_cloud_url');
		if (mode === 'cloud' && cloudUrl) return `${cloudUrl}/api`;
		if (mode === 'local') return 'http://localhost:18080/api';
		return 'http://localhost:8001/api';
	}

	return import.meta.env.VITE_API_URL || '/api';
}

function createLayoutStore() {
	const { subscribe, set, update } = writable<LayoutState>(initialState);

	return {
		subscribe,

		/**
		 * Get default layout (current 5-ring geodesic layout)
		 * This is generated from the current desktop3dStore state
		 */
		getDefaultLayout: (): Layout => {
			const store = get(desktop3dStore);
			const modules: ModulePosition[] = (store.windows || []).map((win) => ({
				module_id: win.module,
				position: win.position,
				rotation: win.rotation || { x: 0, y: 0, z: 0 },
				scale: win.targetScale || 1
			}));

			return {
				id: 'default',
				name: 'Default (5-Ring Geodesic)',
				type: 'default',
				created_at: new Date(),
				updated_at: new Date(),
				is_active: true,
				user_id: '',
				modules
			};
		},

		/**
		 * Initialize and load all layouts from backend
		 */
		initialize: async () => {
			console.log('[Layout Store] Initializing...');
			await desktop3dLayoutStore.loadLayouts();
		},

		/**
		 * Load all layouts from backend
		 */
		loadLayouts: async () => {
			update((s) => ({ ...s, loading: true, error: null }));

			try {
				const baseUrl = getApiBase();
				const response = await fetch(`${baseUrl}/desktop3d/layouts`, {
					credentials: 'include'
				});

				if (response.ok) {
					const layouts: Layout[] = await response.json();

					// Always include default layout at the beginning
					const defaultLayout = desktop3dLayoutStore.getDefaultLayout();

					// Get active layout from backend
					const activeResponse = await fetch(`${baseUrl}/desktop3d/layouts/active`, {
						credentials: 'include'
					});

					let activeLayoutId = 'default';
					if (activeResponse.ok) {
						const activeLayout = await activeResponse.json();
						if (activeLayout?.id) {
							activeLayoutId = activeLayout.id;
						}
					}

					update((s) => ({
						...s,
						layouts: [defaultLayout, ...layouts],
						activeLayoutId,
						loading: false,
						error: null
					}));

					console.log('[Layout Store] ✅ Loaded layouts', {
						count: layouts.length + 1,
						active: activeLayoutId
					});
				} else {
					throw new Error(`HTTP ${response.status}`);
				}
			} catch (err) {
				// On error (e.g. 404 when endpoint doesn't exist), silently fall back to default layout
				// Only log if it's NOT a 404
				if (err instanceof Error && !err.message.includes('404')) {
					console.warn('[Layout Store] Failed to load layouts, using default:', err.message);
				}

				// On error, just show default layout
				const defaultLayout = desktop3dLayoutStore.getDefaultLayout();
				update((s) => ({
					...s,
					layouts: [defaultLayout],
					activeLayoutId: 'default',
					loading: false,
					error: err instanceof Error ? err.message : 'Failed to load layouts'
				}));
			}
		},

		/**
		 * Save current positions as new custom layout
		 */
		saveLayout: async (name: string) => {
			if (!name || name.trim().length === 0) {
				console.error('[Layout Store] Cannot save layout with empty name');
				return false;
			}

			console.log('[Layout Store] Saving layout...', { name });

			const store = get(desktop3dStore);
			const modules: ModulePosition[] = (store.windows || []).map((win) => ({
				module_id: win.module,
				position: win.position,
				rotation: win.rotation || { x: 0, y: 0, z: 0 },
				scale: win.targetScale || 1
			}));

			try {
				const baseUrl = getApiBase();
				const response = await fetch(`${baseUrl}/desktop3d/layouts`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					credentials: 'include',
					body: JSON.stringify({
						name: name.trim(),
						modules
					})
				});

				if (response.ok) {
					const newLayout: Layout = await response.json();

					update((s) => ({
						...s,
						layouts: [...s.layouts, newLayout],
						activeLayoutId: newLayout.id,
						error: null
					}));

					console.log('[Layout Store] ✅ Saved layout', { id: newLayout.id, name });
					return true;
				} else {
					const errorData = await response.json().catch(() => ({}));
					throw new Error(errorData.error || `HTTP ${response.status}`);
				}
			} catch (err) {
				const error = err instanceof Error ? err.message : 'Failed to save layout';
				console.error('[Layout Store] ❌ Failed to save layout:', error);
				update((s) => ({ ...s, error }));
				return false;
			}
		},

		/**
		 * Load a specific layout and apply its module positions
		 */
		loadLayout: async (layoutId: string) => {
			console.log('[Layout Store] Loading layout...', { id: layoutId });

			const state = get(desktop3dLayoutStore);
			const layout = state.layouts.find((l) => l.id === layoutId);

			if (!layout) {
				console.error('[Layout Store] Layout not found:', layoutId);
				return;
			}

			// Apply positions to desktop3dStore
			layout.modules.forEach((modulePos) => {
				desktop3dStore.updateWindowPosition(
					modulePos.module_id,
					modulePos.position,
					modulePos.rotation,
					modulePos.scale
				);
			});

			// Mark as active in backend (if custom layout)
			if (layout.type === 'custom') {
				try {
					const baseUrl = getApiBase();
					await fetch(`${baseUrl}/desktop3d/layouts/${layoutId}/activate`, {
						method: 'POST',
						credentials: 'include'
					});
					console.log('[Layout Store] ✅ Activated layout in backend', { id: layoutId });
				} catch (err) {
					console.error('[Layout Store] Failed to activate layout in backend:', err);
				}
			}

			update((s) => ({ ...s, activeLayoutId: layoutId, error: null }));
			console.log('[Layout Store] ✅ Loaded layout', { id: layoutId, name: layout.name });
		},

		/**
		 * Delete a custom layout
		 */
		deleteLayout: async (layoutId: string) => {
			if (layoutId === 'default') {
				console.warn('[Layout Store] Cannot delete default layout');
				return false;
			}

			console.log('[Layout Store] Deleting layout...', { id: layoutId });

			try {
				const baseUrl = getApiBase();
				const response = await fetch(`${baseUrl}/desktop3d/layouts/${layoutId}`, {
					method: 'DELETE',
					credentials: 'include'
				});

				if (response.ok) {
					update((s) => {
						const newLayouts = s.layouts.filter((l) => l.id !== layoutId);
						const newActiveId = s.activeLayoutId === layoutId ? 'default' : s.activeLayoutId;

						return {
							...s,
							layouts: newLayouts,
							activeLayoutId: newActiveId,
							error: null
						};
					});

					console.log('[Layout Store] ✅ Deleted layout', { id: layoutId });
					return true;
				} else {
					const errorData = await response.json().catch(() => ({}));
					throw new Error(errorData.error || `HTTP ${response.status}`);
				}
			} catch (err) {
				const error = err instanceof Error ? err.message : 'Failed to delete layout';
				console.error('[Layout Store] ❌ Failed to delete layout:', error);
				update((s) => ({ ...s, error }));
				return false;
			}
		},

		/**
		 * Toggle edit mode
		 */
		toggleEditMode: () => {
			update((s) => ({ ...s, editMode: !s.editMode }));
			console.log('[Layout Store] Edit mode toggled');
		},

		/**
		 * Enter edit mode
		 */
		enterEditMode: () => {
			update((s) => ({ ...s, editMode: true }));
			console.log('[Layout Store] ✏️ Entered edit mode');
		},

		/**
		 * Exit edit mode
		 */
		exitEditMode: () => {
			update((s) => ({ ...s, editMode: false }));
			console.log('[Layout Store] ✅ Exited edit mode');
		},

		/**
		 * Reset error state
		 */
		clearError: () => {
			update((s) => ({ ...s, error: null }));
		}
	};
}

export const desktop3dLayoutStore = createLayoutStore();

// Derived stores for convenience
export const activeLayout = derived(desktop3dLayoutStore, ($store) =>
	$store.layouts.find((l) => l.id === $store.activeLayoutId)
);

export const customLayouts = derived(desktop3dLayoutStore, ($store) =>
	$store.layouts.filter((l) => l.type === 'custom')
);

export const isEditMode = derived(desktop3dLayoutStore, ($store) => $store.editMode);

export const layoutsLoading = derived(desktop3dLayoutStore, ($store) => $store.loading);

export const layoutsError = derived(desktop3dLayoutStore, ($store) => $store.error);

// Initialize on import (will be called when entering 3D Desktop)
// Note: Must be called manually via desktop3dLayoutStore.initialize() when ready
