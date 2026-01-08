// Window Store - State management for the desktop environment
import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';
import { soundStore } from './soundStore';

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
	data?: Record<string, unknown>; // Custom data passed when opening window
}

export interface CustomIconConfig {
	type: 'lucide' | 'custom';
	lucideName?: string;        // e.g., 'Home', 'Settings' - name from lucide-svelte
	customSvg?: string;         // Base64 or raw SVG string for custom icons
	foregroundColor?: string;   // Override icon color
	backgroundColor?: string;   // Override background color
}

export interface DesktopIcon {
	id: string;
	module: string;
	label: string;
	x: number;
	y: number;
	type?: 'app' | 'folder';
	folderId?: string; // If icon is inside a folder
	folderColor?: string; // For folder icons
	customIcon?: CustomIconConfig; // Per-icon customization
}

export interface DesktopFolder {
	id: string;
	name: string;
	color: string;
	iconIds: string[]; // Icons inside this folder
}

interface WindowStore {
	windows: WindowState[];
	focusedWindowId: string | null;
	windowOrder: string[]; // Z-index order, last is on top
	dockPinnedItems: string[];
	desktopIcons: DesktopIcon[];
	selectedIconIds: string[];
	folders: DesktopFolder[];
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
	tables: { title: 'Tables', width: 1100, height: 750, minWidth: 700, minHeight: 500 },
	pages: { title: 'Pages', width: 900, height: 650, minWidth: 500, minHeight: 400 },
	contexts: { title: 'Pages', width: 900, height: 650, minWidth: 500, minHeight: 400 }, // Legacy alias
	nodes: { title: 'Nodes', width: 1000, height: 700, minWidth: 600, minHeight: 400 },
	daily: { title: 'Daily Log', width: 700, height: 550, minWidth: 350, minHeight: 300 },
	settings: { title: 'Settings', width: 700, height: 550, minWidth: 400, minHeight: 350 },
	communication: { title: 'Communication', width: 1000, height: 700, minWidth: 600, minHeight: 450 },
	'ai-settings': { title: 'AI Settings', width: 800, height: 600, minWidth: 500, minHeight: 400 },
	integrations: { title: 'Integrations', width: 950, height: 700, minWidth: 600, minHeight: 500 },
	trash: { title: 'Trash', width: 600, height: 450, minWidth: 300, minHeight: 250 },
	terminal: { title: 'Terminal - OS Agent', width: 700, height: 500, minWidth: 400, minHeight: 300 },
	'desktop-settings': { title: 'Desktop Settings', width: 550, height: 500, minWidth: 450, minHeight: 400 },
	folder: { title: 'Folder', width: 600, height: 450, minWidth: 300, minHeight: 250 },
	files: { title: 'Files', width: 900, height: 600, minWidth: 500, minHeight: 400 },
	finder: { title: 'Finder', width: 900, height: 600, minWidth: 500, minHeight: 400 },
	help: { title: 'Help', width: 900, height: 650, minWidth: 600, minHeight: 450 },
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
	{ id: 'icon-tables', module: 'tables', label: 'Tables', x: -1, y: 7 },
	{ id: 'icon-communication', module: 'communication', label: 'Communication', x: -1, y: 8 },
	{ id: 'icon-files', module: 'files', label: 'Files', x: -2, y: 0 },
	{ id: 'icon-pages', module: 'pages', label: 'Pages', x: -2, y: 1 },
	{ id: 'icon-nodes', module: 'nodes', label: 'Nodes', x: -2, y: 2 },
	{ id: 'icon-daily', module: 'daily', label: 'Daily Log', x: -2, y: 3 },
	{ id: 'icon-settings', module: 'settings', label: 'Settings', x: -2, y: 4 },
	{ id: 'icon-ai-settings', module: 'ai-settings', label: 'AI Settings', x: -2, y: 5 },
	{ id: 'icon-integrations', module: 'integrations', label: 'Integrations', x: -2, y: 6 },
	{ id: 'icon-trash', module: 'trash', label: 'Trash', x: -1, y: -1 }, // Bottom right
];

const initialState: WindowStore = {
	windows: [],
	focusedWindowId: null,
	windowOrder: [],
	dockPinnedItems: ['finder', 'dashboard', 'chat', 'projects', 'tasks', 'clients'],
	desktopIcons: initialDesktopIcons,
	selectedIconIds: [],
	folders: [],
};

// Storage key for persisting desktop settings
const STORAGE_KEY = 'businessos_desktop_settings';

// Load saved desktop settings from localStorage
function loadSavedSettings(): Partial<WindowStore> {
	if (!browser) return {};

	try {
		const saved = localStorage.getItem(STORAGE_KEY);
		if (saved) {
			const parsed = JSON.parse(saved);
			// Ensure we have valid arrays
			let desktopIcons = Array.isArray(parsed.desktopIcons) && parsed.desktopIcons.length > 0
				? parsed.desktopIcons
				: initialDesktopIcons;
			const dockPinnedItems = Array.isArray(parsed.dockPinnedItems) && parsed.dockPinnedItems.length > 0
				? parsed.dockPinnedItems
				: initialState.dockPinnedItems;
			const folders = Array.isArray(parsed.folders) ? parsed.folders : [];

			// Merge in any new default icons that were added since last save
			// This ensures new modules (like integrations) appear on existing users' desktops
			const savedIconIds = new Set(desktopIcons.map((i: DesktopIcon) => i.id));
			const newIcons = initialDesktopIcons.filter(icon => !savedIconIds.has(icon.id));
			if (newIcons.length > 0) {
				desktopIcons = [...desktopIcons, ...newIcons];
			}

			// Update labels for existing icons to match defaults (preserve user positions)
			// This ensures label renames like "Contexts" -> "Knowledge" are applied
			const defaultLabels = new Map(initialDesktopIcons.map(i => [i.id, i.label]));
			desktopIcons = desktopIcons.map((icon: DesktopIcon) => {
				const defaultLabel = defaultLabels.get(icon.id);
				if (defaultLabel && icon.label !== defaultLabel) {
					return { ...icon, label: defaultLabel };
				}
				return icon;
			});

			return { desktopIcons, dockPinnedItems, folders };
		}
	} catch (e) {
		console.error('Failed to load desktop settings:', e);
	}
	return {};
}

// Desktop config schema version for backwards compatibility
const CONFIG_VERSION = '1.0.0';

// Export config type for JSON export/import
export interface DesktopConfig {
	version: string;
	exportedAt: string;
	desktopIcons: DesktopIcon[];
	dockPinnedItems: string[];
	folders: DesktopFolder[];
}

// Save desktop settings to localStorage
function saveSettings(state: WindowStore) {
	if (!browser) return;

	try {
		const config = {
			version: CONFIG_VERSION,
			desktopIcons: state.desktopIcons,
			dockPinnedItems: state.dockPinnedItems,
			folders: state.folders,
		};
		localStorage.setItem(STORAGE_KEY, JSON.stringify(config));
	} catch (e) {
		console.error('Failed to save desktop settings:', e);
	}
}

function createWindowStore() {
	// Merge initial state with any saved settings
	const savedSettings = loadSavedSettings();
	const mergedInitial: WindowStore = {
		...initialState,
		...savedSettings,
	};

	const { subscribe, set, update } = writable<WindowStore>(mergedInitial);

	let cascadeOffset = 0;
	let initialized = false;

	return {
		subscribe,

		// Initialize store on client side (call this in onMount)
		initialize: () => {
			if (initialized || typeof window === 'undefined') return;
			initialized = true;

			const saved = loadSavedSettings();

			update(state => {
				// Get dock items from saved settings or current state
				let dockItems = (Object.keys(saved).length > 0 && saved.dockPinnedItems)
					? saved.dockPinnedItems
					: state.dockPinnedItems;

				// ALWAYS ensure Finder is at the beginning of dock
				if (!dockItems.includes('finder')) {
					dockItems = ['finder', ...dockItems];
				} else if (dockItems[0] !== 'finder') {
					// Move finder to first position
					dockItems = ['finder', ...dockItems.filter(item => item !== 'finder')];
				}

				// Get desktop icons - use saved if available, otherwise keep current state (initial)
				const desktopIcons = (Object.keys(saved).length > 0 && saved.desktopIcons && saved.desktopIcons.length > 0)
					? saved.desktopIcons
					: state.desktopIcons;

				// Get folders
				const folders = (Object.keys(saved).length > 0 && saved.folders)
					? saved.folders
					: state.folders;

				const newState = {
					...state,
					dockPinnedItems: dockItems,
					desktopIcons,
					folders,
				};

				// Save the updated settings
				saveSettings(newState);

				return newState;
			});
		},

		// Open a new window for a module
		openWindow: (module: string, options?: string | { title?: string; data?: Record<string, unknown> }) => {
			// Handle legacy string parameter (custom title)
			const customTitle = typeof options === 'string' ? options : options?.title;
			const windowData = typeof options === 'object' ? options?.data : undefined;

			update(state => {
				// Check if window already exists for this module
				const existingWindow = state.windows.find(w => w.module === module && !w.minimized);
				if (existingWindow) {
					// Update data if provided and focus window
					const updatedWindows = windowData
						? state.windows.map(w => w.id === existingWindow.id ? { ...w, data: { ...w.data, ...windowData } } : w)
						: state.windows;
					return {
						...state,
						windows: updatedWindows,
						focusedWindowId: existingWindow.id,
						windowOrder: [...state.windowOrder.filter(id => id !== existingWindow.id), existingWindow.id]
					};
				}

				// Check if minimized window exists
				const minimizedWindow = state.windows.find(w => w.module === module && w.minimized);
				if (minimizedWindow) {
					// Restore it and update data if provided
					return {
						...state,
						windows: state.windows.map(w =>
							w.id === minimizedWindow.id
								? { ...w, minimized: false, data: windowData ? { ...w.data, ...windowData } : w.data }
								: w
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
					data: windowData,
				};

				// Play window open sound
				soundStore.playSound('windowOpen');

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
			// Play window close sound
			soundStore.playSound('windowClose');
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
			// Play minimize sound
			soundStore.playSound('windowMinimize');
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
			// Play maximize sound
			soundStore.playSound('windowMaximize');
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
			update(state => {
				const newState = {
					...state,
					desktopIcons: state.desktopIcons.map(icon =>
						icon.id === iconId ? { ...icon, x, y } : icon
					)
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Add item to dock
		addToDock: (module: string) => {
			update(state => {
				if (state.dockPinnedItems.includes(module)) return state;
				const newState = {
					...state,
					dockPinnedItems: [...state.dockPinnedItems, module]
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Remove item from dock
		removeFromDock: (module: string) => {
			update(state => {
				const newState = {
					...state,
					dockPinnedItems: state.dockPinnedItems.filter(m => m !== module)
				};
				saveSettings(newState);
				return newState;
			});
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

		// Reset store (clears saved settings too)
		reset: () => {
			if (browser) {
				localStorage.removeItem(STORAGE_KEY);
			}
			set(initialState);
		},

		// Reset only desktop settings (icons, dock) to defaults
		resetDesktop: () => {
			if (browser) {
				localStorage.removeItem(STORAGE_KEY);
			}
			update(state => ({
				...state,
				desktopIcons: initialDesktopIcons,
				dockPinnedItems: initialState.dockPinnedItems,
				folders: [],
			}));
		},

		// Export desktop configuration as JSON
		exportConfig: (): DesktopConfig => {
			let config: DesktopConfig = {
				version: CONFIG_VERSION,
				exportedAt: new Date().toISOString(),
				desktopIcons: initialDesktopIcons,
				dockPinnedItems: initialState.dockPinnedItems,
				folders: [],
			};
			const unsubscribe = subscribe(state => {
				config = {
					version: CONFIG_VERSION,
					exportedAt: new Date().toISOString(),
					desktopIcons: state.desktopIcons,
					dockPinnedItems: state.dockPinnedItems,
					folders: state.folders,
				};
			});
			unsubscribe();
			return config;
		},

		// Import desktop configuration from JSON
		importConfig: (config: DesktopConfig): { success: boolean; error?: string } => {
			try {
				// Validate config structure
				if (!config || typeof config !== 'object') {
					return { success: false, error: 'Invalid configuration format' };
				}
				if (!Array.isArray(config.desktopIcons)) {
					return { success: false, error: 'Missing or invalid desktopIcons' };
				}
				if (!Array.isArray(config.dockPinnedItems)) {
					return { success: false, error: 'Missing or invalid dockPinnedItems' };
				}

				// Validate each icon has required fields
				for (const icon of config.desktopIcons) {
					if (!icon.id || !icon.module || !icon.label || typeof icon.x !== 'number' || typeof icon.y !== 'number') {
						return { success: false, error: 'Invalid icon structure' };
					}
				}

				update(state => {
					// Ensure Finder is in dock
					let dockItems = config.dockPinnedItems;
					if (!dockItems.includes('finder')) {
						dockItems = ['finder', ...dockItems];
					}

					const newState = {
						...state,
						desktopIcons: config.desktopIcons,
						dockPinnedItems: dockItems,
						folders: config.folders || [],
					};
					saveSettings(newState);
					return newState;
				});

				return { success: true };
			} catch (e) {
				return { success: false, error: 'Failed to import configuration' };
			}
		},

		// Get JSON schema for desktop config
		getConfigSchema: () => ({
			$schema: 'http://json-schema.org/draft-07/schema#',
			title: 'BusinessOS Desktop Configuration',
			type: 'object',
			required: ['version', 'desktopIcons', 'dockPinnedItems'],
			properties: {
				version: { type: 'string', description: 'Config version' },
				exportedAt: { type: 'string', format: 'date-time', description: 'Export timestamp' },
				desktopIcons: {
					type: 'array',
					items: {
						type: 'object',
						required: ['id', 'module', 'label', 'x', 'y'],
						properties: {
							id: { type: 'string' },
							module: { type: 'string' },
							label: { type: 'string' },
							x: { type: 'number' },
							y: { type: 'number' },
							type: { type: 'string', enum: ['app', 'folder'] },
							folderId: { type: 'string' },
							folderColor: { type: 'string' },
						},
					},
				},
				dockPinnedItems: {
					type: 'array',
					items: { type: 'string' },
				},
				folders: {
					type: 'array',
					items: {
						type: 'object',
						required: ['id', 'name', 'color', 'iconIds'],
						properties: {
							id: { type: 'string' },
							name: { type: 'string' },
							color: { type: 'string' },
							iconIds: { type: 'array', items: { type: 'string' } },
						},
					},
				},
			},
		}),

		// Create a new folder
		createFolder: (name: string, x: number, y: number, color: string = '#3B82F6') => {
			update(state => {
				const folderId = `folder-${Date.now()}`;
				const iconId = `icon-${folderId}`;

				const newFolder: DesktopFolder = {
					id: folderId,
					name,
					color,
					iconIds: [],
				};

				const newIcon: DesktopIcon = {
					id: iconId,
					module: 'folder',
					label: name,
					x,
					y,
					type: 'folder',
					folderId,
					folderColor: color,
				};

				const newState = {
					...state,
					folders: [...state.folders, newFolder],
					desktopIcons: [...state.desktopIcons, newIcon],
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Rename a folder
		renameFolder: (folderId: string, newName: string) => {
			update(state => {
				const newState = {
					...state,
					folders: state.folders.map(f =>
						f.id === folderId ? { ...f, name: newName } : f
					),
					desktopIcons: state.desktopIcons.map(icon =>
						icon.folderId === folderId ? { ...icon, label: newName } : icon
					),
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Change folder color
		setFolderColor: (folderId: string, color: string) => {
			update(state => {
				const newState = {
					...state,
					folders: state.folders.map(f =>
						f.id === folderId ? { ...f, color } : f
					),
					desktopIcons: state.desktopIcons.map(icon =>
						icon.folderId === folderId && icon.type === 'folder'
							? { ...icon, folderColor: color }
							: icon
					),
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Delete a folder (moves icons back to desktop)
		deleteFolder: (folderId: string) => {
			update(state => {
				const folder = state.folders.find(f => f.id === folderId);

				// Move icons out of folder back to desktop
				let updatedIcons = state.desktopIcons.map(icon => {
					if (folder?.iconIds.includes(icon.id)) {
						return { ...icon, folderId: undefined };
					}
					return icon;
				});

				// Remove the folder icon
				updatedIcons = updatedIcons.filter(icon =>
					!(icon.type === 'folder' && icon.folderId === folderId)
				);

				const newState = {
					...state,
					folders: state.folders.filter(f => f.id !== folderId),
					desktopIcons: updatedIcons,
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Move an icon into a folder
		moveIconToFolder: (iconId: string, folderId: string) => {
			update(state => {
				const newState = {
					...state,
					folders: state.folders.map(f => {
						if (f.id === folderId) {
							// Add icon to this folder
							return {
								...f,
								iconIds: f.iconIds.includes(iconId)
									? f.iconIds
									: [...f.iconIds, iconId],
							};
						}
						// Remove icon from other folders
						return {
							...f,
							iconIds: f.iconIds.filter(id => id !== iconId),
						};
					}),
					desktopIcons: state.desktopIcons.map(icon =>
						icon.id === iconId
							? { ...icon, folderId }
							: icon
					),
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Remove an icon from its folder (back to desktop)
		removeIconFromFolder: (iconId: string) => {
			update(state => {
				const newState = {
					...state,
					folders: state.folders.map(f => ({
						...f,
						iconIds: f.iconIds.filter(id => id !== iconId),
					})),
					desktopIcons: state.desktopIcons.map(icon =>
						icon.id === iconId
							? { ...icon, folderId: undefined }
							: icon
					),
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Update icon customization (Lucide icon or custom SVG)
		updateIconCustomization: (iconId: string, customIcon: CustomIconConfig | undefined) => {
			update(state => {
				const newState = {
					...state,
					desktopIcons: state.desktopIcons.map(icon =>
						icon.id === iconId
							? { ...icon, customIcon }
							: icon
					),
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Reset icon to default appearance
		resetIconCustomization: (iconId: string) => {
			update(state => {
				const newState = {
					...state,
					desktopIcons: state.desktopIcons.map(icon =>
						icon.id === iconId
							? { ...icon, customIcon: undefined }
							: icon
					),
				};
				saveSettings(newState);
				return newState;
			});
		},

		// Get folder by ID
		getFolder: (folderId: string): DesktopFolder | undefined => {
			let result: DesktopFolder | undefined;
			const unsubscribe = subscribe(state => {
				result = state.folders.find(f => f.id === folderId);
			});
			unsubscribe();
			return result;
		},

		// Get icons in a folder
		getIconsInFolder: (folderId: string): DesktopIcon[] => {
			let result: DesktopIcon[] = [];
			const unsubscribe = subscribe(state => {
				result = state.desktopIcons.filter(icon => icon.folderId === folderId && icon.type !== 'folder');
			});
			unsubscribe();
			return result;
		},

		// Open folder window
		openFolder: (folderId: string) => {
			update(state => {
				const folder = state.folders.find(f => f.id === folderId);
				if (!folder) return state;

				// Check if folder window already exists
				const existingWindow = state.windows.find(w => w.module === `folder-${folderId}` && !w.minimized);
				if (existingWindow) {
					return {
						...state,
						focusedWindowId: existingWindow.id,
						windowOrder: [...state.windowOrder.filter(id => id !== existingWindow.id), existingWindow.id]
					};
				}

				const defaults = moduleDefaults.folder;
				const id = `folder-${folderId}-${Date.now()}`;

				const baseX = 150 + Math.random() * 100;
				const baseY = 80 + Math.random() * 50;

				const newWindow: WindowState = {
					id,
					module: `folder-${folderId}`,
					title: folder.name,
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
