// Window Store - State management for the desktop environment
import { writable, derived, get } from 'svelte/store';

export type SnapZone = 'left' | 'right' | 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | null;

export interface WindowState {
	id: string;
	module: string;
	title: string;
	x: number;
	y: number;
	width: number;
	height: number;
	minWidth: number;
	minHeight: number;
	minimized: boolean;
	maximized: boolean;
	snapped?: SnapZone;
	previousBounds?: { x: number; y: number; width: number; height: number };
}

export interface DesktopIcon {
	id: string;
	module: string;
	label: string;
	x: number;
	y: number;
}

interface WindowStore {
	windows: WindowState[];
	focusedWindowId: string | null;
	windowOrder: string[]; // Z-index order, last is on top
	dockPinnedItems: string[];
	desktopIcons: DesktopIcon[];
	selectedIconIds: string[];
}

// Default module configurations
const moduleDefaults: Record<string, { title: string; width: number; height: number; minWidth: number; minHeight: number }> = {
	platform: { title: 'Business OS', width: 1200, height: 800, minWidth: 800, minHeight: 600 },
	dashboard: { title: 'Dashboard', width: 1000, height: 700, minWidth: 600, minHeight: 400 },
	chat: { title: 'Chat', width: 900, height: 650, minWidth: 400, minHeight: 300 },
	tasks: { title: 'Tasks', width: 800, height: 600, minWidth: 400, minHeight: 300 },
	projects: { title: 'Projects', width: 900, height: 650, minWidth: 500, minHeight: 400 },
	team: { title: 'Team', width: 850, height: 600, minWidth: 400, minHeight: 300 },
	clients: { title: 'Clients', width: 1000, height: 700, minWidth: 600, minHeight: 400 },
	contexts: { title: 'Contexts', width: 800, height: 600, minWidth: 400, minHeight: 300 },
	nodes: { title: 'Nodes', width: 1000, height: 700, minWidth: 600, minHeight: 400 },
	daily: { title: 'Daily Log', width: 700, height: 550, minWidth: 350, minHeight: 300 },
	settings: { title: 'Settings', width: 700, height: 550, minWidth: 400, minHeight: 350 },
	trash: { title: 'Trash', width: 600, height: 450, minWidth: 300, minHeight: 250 },
	terminal: { title: 'Terminal - OS Agent', width: 700, height: 500, minWidth: 400, minHeight: 300 },
	'desktop-settings': { title: 'Desktop Settings', width: 550, height: 500, minWidth: 450, minHeight: 400 },
};

// Initial desktop icon positions (right side, top to bottom)
const initialDesktopIcons: DesktopIcon[] = [
	{ id: 'icon-platform', module: 'platform', label: 'Business OS', x: 0, y: 0 }, // Top left - full platform
	{ id: 'icon-terminal', module: 'terminal', label: 'Terminal', x: -1, y: 0 },
	{ id: 'icon-dashboard', module: 'dashboard', label: 'Dashboard', x: -1, y: 1 },
	{ id: 'icon-chat', module: 'chat', label: 'Chat', x: -1, y: 2 },
	{ id: 'icon-tasks', module: 'tasks', label: 'Tasks', x: -1, y: 3 },
	{ id: 'icon-projects', module: 'projects', label: 'Projects', x: -1, y: 4 },
	{ id: 'icon-team', module: 'team', label: 'Team', x: -1, y: 5 },
	{ id: 'icon-clients', module: 'clients', label: 'Clients', x: -1, y: 6 },
	{ id: 'icon-contexts', module: 'contexts', label: 'Contexts', x: -2, y: 0 },
	{ id: 'icon-nodes', module: 'nodes', label: 'Nodes', x: -2, y: 1 },
	{ id: 'icon-daily', module: 'daily', label: 'Daily Log', x: -2, y: 2 },
	{ id: 'icon-settings', module: 'settings', label: 'Settings', x: -2, y: 3 },
	{ id: 'icon-trash', module: 'trash', label: 'Trash', x: -1, y: -1 }, // Bottom right
];

const initialState: WindowStore = {
	windows: [],
	focusedWindowId: null,
	windowOrder: [],
	dockPinnedItems: ['dashboard', 'chat', 'projects', 'tasks', 'clients'],
	desktopIcons: initialDesktopIcons,
	selectedIconIds: [],
};

function createWindowStore() {
	const { subscribe, set, update } = writable<WindowStore>(initialState);

	let cascadeOffset = 0;

	return {
		subscribe,

		// Open a new window for a module
		openWindow: (module: string, customTitle?: string) => {
			update(state => {
				// Check if window already exists for this module
				const existingWindow = state.windows.find(w => w.module === module && !w.minimized);
				if (existingWindow) {
					// Just focus it
					return {
						...state,
						focusedWindowId: existingWindow.id,
						windowOrder: [...state.windowOrder.filter(id => id !== existingWindow.id), existingWindow.id]
					};
				}

				// Check if minimized window exists
				const minimizedWindow = state.windows.find(w => w.module === module && w.minimized);
				if (minimizedWindow) {
					// Restore it
					return {
						...state,
						windows: state.windows.map(w =>
							w.id === minimizedWindow.id ? { ...w, minimized: false } : w
						),
						focusedWindowId: minimizedWindow.id,
						windowOrder: [...state.windowOrder.filter(id => id !== minimizedWindow.id), minimizedWindow.id]
					};
				}

				const defaults = moduleDefaults[module] || { title: module, width: 800, height: 600, minWidth: 400, minHeight: 300 };
				const id = `${module}-${Date.now()}`;

				// Calculate cascade position
				const baseX = 100 + (cascadeOffset * 30);
				const baseY = 50 + (cascadeOffset * 30);
				cascadeOffset = (cascadeOffset + 1) % 10;

				const newWindow: WindowState = {
					id,
					module,
					title: customTitle || defaults.title,
					x: baseX,
					y: baseY,
					width: defaults.width,
					height: defaults.height,
					minWidth: defaults.minWidth,
					minHeight: defaults.minHeight,
					minimized: false,
					maximized: false,
				};

				return {
					...state,
					windows: [...state.windows, newWindow],
					focusedWindowId: id,
					windowOrder: [...state.windowOrder, id]
				};
			});
		},

		// Close a window
		closeWindow: (windowId: string) => {
			update(state => {
				const newWindows = state.windows.filter(w => w.id !== windowId);
				const newOrder = state.windowOrder.filter(id => id !== windowId);
				const newFocused = state.focusedWindowId === windowId
					? (newOrder.length > 0 ? newOrder[newOrder.length - 1] : null)
					: state.focusedWindowId;

				return {
					...state,
					windows: newWindows,
					windowOrder: newOrder,
					focusedWindowId: newFocused
				};
			});
		},

		// Minimize a window
		minimizeWindow: (windowId: string) => {
			update(state => {
				const newOrder = state.windowOrder.filter(id => id !== windowId);
				const newFocused = state.focusedWindowId === windowId
					? (newOrder.length > 0 ? newOrder[newOrder.length - 1] : null)
					: state.focusedWindowId;

				return {
					...state,
					windows: state.windows.map(w =>
						w.id === windowId ? { ...w, minimized: true } : w
					),
					windowOrder: newOrder,
					focusedWindowId: newFocused
				};
			});
		},

		// Restore a minimized window
		restoreWindow: (windowId: string) => {
			update(state => ({
				...state,
				windows: state.windows.map(w =>
					w.id === windowId ? { ...w, minimized: false } : w
				),
				focusedWindowId: windowId,
				windowOrder: [...state.windowOrder.filter(id => id !== windowId), windowId]
			}));
		},

		// Toggle maximize state
		toggleMaximize: (windowId: string) => {
			update(state => ({
				...state,
				windows: state.windows.map(w => {
					if (w.id !== windowId) return w;

					if (w.maximized) {
						// Restore to previous bounds
						return {
							...w,
							maximized: false,
							snapped: null,
							x: w.previousBounds?.x ?? w.x,
							y: w.previousBounds?.y ?? w.y,
							width: w.previousBounds?.width ?? w.width,
							height: w.previousBounds?.height ?? w.height,
							previousBounds: undefined
						};
					} else {
						// Store current bounds and maximize
						return {
							...w,
							maximized: true,
							snapped: null,
							previousBounds: { x: w.x, y: w.y, width: w.width, height: w.height }
						};
					}
				})
			}));
		},

		// Snap window to a zone (split screen / quadrants)
		snapWindow: (windowId: string, zone: SnapZone, workspaceWidth: number, workspaceHeight: number) => {
			update(state => ({
				...state,
				windows: state.windows.map(w => {
					if (w.id !== windowId) return w;

					// If unsnapping, restore previous bounds
					if (!zone) {
						return {
							...w,
							snapped: null,
							maximized: false,
							x: w.previousBounds?.x ?? w.x,
							y: w.previousBounds?.y ?? w.y,
							width: w.previousBounds?.width ?? w.width,
							height: w.previousBounds?.height ?? w.height,
							previousBounds: undefined
						};
					}

					// Store current bounds if not already snapped
					const prevBounds = w.snapped ? w.previousBounds : { x: w.x, y: w.y, width: w.width, height: w.height };

					// Calculate new bounds based on zone
					let newBounds = { x: 0, y: 0, width: workspaceWidth, height: workspaceHeight };

					switch (zone) {
						case 'left':
							newBounds = { x: 0, y: 0, width: workspaceWidth / 2, height: workspaceHeight };
							break;
						case 'right':
							newBounds = { x: workspaceWidth / 2, y: 0, width: workspaceWidth / 2, height: workspaceHeight };
							break;
						case 'top-left':
							newBounds = { x: 0, y: 0, width: workspaceWidth / 2, height: workspaceHeight / 2 };
							break;
						case 'top-right':
							newBounds = { x: workspaceWidth / 2, y: 0, width: workspaceWidth / 2, height: workspaceHeight / 2 };
							break;
						case 'bottom-left':
							newBounds = { x: 0, y: workspaceHeight / 2, width: workspaceWidth / 2, height: workspaceHeight / 2 };
							break;
						case 'bottom-right':
							newBounds = { x: workspaceWidth / 2, y: workspaceHeight / 2, width: workspaceWidth / 2, height: workspaceHeight / 2 };
							break;
					}

					return {
						...w,
						snapped: zone,
						maximized: false,
						x: newBounds.x,
						y: newBounds.y,
						width: newBounds.width,
						height: newBounds.height,
						previousBounds: prevBounds
					};
				})
			}));
		},

		// Focus a window (bring to front)
		focusWindow: (windowId: string) => {
			update(state => ({
				...state,
				focusedWindowId: windowId,
				windowOrder: [...state.windowOrder.filter(id => id !== windowId), windowId]
			}));
		},

		// Update window position
		updateWindowPosition: (windowId: string, x: number, y: number) => {
			update(state => ({
				...state,
				windows: state.windows.map(w =>
					w.id === windowId ? { ...w, x, y, maximized: false } : w
				)
			}));
		},

		// Update window size
		updateWindowSize: (windowId: string, width: number, height: number) => {
			update(state => ({
				...state,
				windows: state.windows.map(w =>
					w.id === windowId ? {
						...w,
						width: Math.max(width, w.minWidth),
						height: Math.max(height, w.minHeight),
						maximized: false
					} : w
				)
			}));
		},

		// Update window bounds (position and size)
		updateWindowBounds: (windowId: string, x: number, y: number, width: number, height: number) => {
			update(state => ({
				...state,
				windows: state.windows.map(w =>
					w.id === windowId ? {
						...w,
						x,
						y,
						width: Math.max(width, w.minWidth),
						height: Math.max(height, w.minHeight),
						maximized: false
					} : w
				)
			}));
		},

		// Select desktop icon
		selectIcon: (iconId: string, additive: boolean = false) => {
			update(state => ({
				...state,
				selectedIconIds: additive
					? (state.selectedIconIds.includes(iconId)
						? state.selectedIconIds.filter(id => id !== iconId)
						: [...state.selectedIconIds, iconId])
					: [iconId]
			}));
		},

		// Clear icon selection
		clearIconSelection: () => {
			update(state => ({
				...state,
				selectedIconIds: []
			}));
		},

		// Set selected icons (for lasso selection)
		setSelectedIcons: (iconIds: string[]) => {
			update(state => ({
				...state,
				selectedIconIds: iconIds
			}));
		},

		// Update desktop icon position
		updateIconPosition: (iconId: string, x: number, y: number) => {
			update(state => ({
				...state,
				desktopIcons: state.desktopIcons.map(icon =>
					icon.id === iconId ? { ...icon, x, y } : icon
				)
			}));
		},

		// Add item to dock
		addToDock: (module: string) => {
			update(state => {
				if (state.dockPinnedItems.includes(module)) return state;
				return {
					...state,
					dockPinnedItems: [...state.dockPinnedItems, module]
				};
			});
		},

		// Remove item from dock
		removeFromDock: (module: string) => {
			update(state => ({
				...state,
				dockPinnedItems: state.dockPinnedItems.filter(m => m !== module)
			}));
		},

		// Cycle to next window (for Cmd+`)
		cycleWindows: () => {
			update(state => {
				const visibleWindows = state.windows.filter(w => !w.minimized);
				if (visibleWindows.length < 2) return state;

				const currentIndex = state.windowOrder.indexOf(state.focusedWindowId || '');
				const visibleOrder = state.windowOrder.filter(id =>
					state.windows.find(w => w.id === id && !w.minimized)
				);

				if (visibleOrder.length < 2) return state;

				const currentVisibleIndex = visibleOrder.indexOf(state.focusedWindowId || '');
				const nextIndex = (currentVisibleIndex + 1) % visibleOrder.length;
				const nextWindowId = visibleOrder[nextIndex];

				return {
					...state,
					focusedWindowId: nextWindowId,
					windowOrder: [...state.windowOrder.filter(id => id !== nextWindowId), nextWindowId]
				};
			});
		},

		// Get windows for a specific module (for dock indicator)
		getWindowsForModule: (module: string) => {
			let result: WindowState[] = [];
			const unsubscribe = subscribe(state => {
				result = state.windows.filter(w => w.module === module);
			});
			unsubscribe();
			return result;
		},

		// Reset store
		reset: () => set(initialState)
	};
}

export const windowStore = createWindowStore();

// Derived stores
export const focusedWindow = derived(windowStore, $store =>
	$store.windows.find(w => w.id === $store.focusedWindowId) || null
);

export const visibleWindows = derived(windowStore, $store =>
	$store.windows.filter(w => !w.minimized)
);

export const minimizedWindows = derived(windowStore, $store =>
	$store.windows.filter(w => w.minimized)
);

export const openModules = derived(windowStore, $store =>
	[...new Set($store.windows.map(w => w.module))]
);
