/**
 * Desktop 3D Store
 * Manages state for the experimental 3D desktop view
 */
import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';

// Core modules that are always visible - only modules with actual routes
export const CORE_MODULES = [
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
	'terminal',
	'settings',
	'help'
] as const;

// All available modules
export const ALL_MODULES = [
	'dashboard',
	'chat',
	'tasks',
	'projects',
	'team',
	'clients',
	'tables',
	'communication',
	'pages',
	'daily',
	'settings',
	'terminal',
	'nodes',
	'files',
	'help'
] as const;

export type ModuleId = (typeof ALL_MODULES)[number];
export type ViewMode = 'orb' | 'grid' | 'focused';

export interface Window3DState {
	id: string;
	module: ModuleId;
	title: string;
	position: [number, number, number];
	targetPosition: [number, number, number];
	rotation: [number, number, number];
	scale: number;
	targetScale: number;
	opacity: number;
	targetOpacity: number;
	isCore: boolean;
	isOpen: boolean;
	isFocused: boolean;
	lastFocused: number;
	color: string;
	// Window dimensions (resizable)
	width: number;
	height: number;
}

export interface Desktop3DState {
	viewMode: ViewMode;
	windows: Window3DState[];
	focusedWindowId: string | null;
	sphereRadius: number;
	gridColumns: number;
	gridSpacing: number;
	autoRotate: boolean;
	animating: boolean;
}

// Module metadata
export const MODULE_INFO: Record<
	ModuleId,
	{ title: string; color: string; icon: string }
> = {
	dashboard: { title: 'Dashboard', color: '#1E88E5', icon: 'grid' },
	chat: { title: 'Chat', color: '#43A047', icon: 'message-circle' },
	tasks: { title: 'Tasks', color: '#FB8C00', icon: 'check-square' },
	projects: { title: 'Projects', color: '#8E24AA', icon: 'folder' },
	team: { title: 'Team', color: '#00ACC1', icon: 'users' },
	clients: { title: 'Clients', color: '#5C6BC0', icon: 'briefcase' },
	tables: { title: 'Tables', color: '#6366F1', icon: 'table' },
	communication: { title: 'Communication', color: '#E53935', icon: 'mail' },
	pages: { title: 'Pages', color: '#7CB342', icon: 'book' },
	nodes: { title: 'Nodes', color: '#FF7043', icon: 'share-2' },
	daily: { title: 'Daily Log', color: '#26A69A', icon: 'edit' },
	settings: { title: 'Settings', color: '#78909C', icon: 'settings' },
	terminal: { title: 'Terminal', color: '#37474F', icon: 'terminal' },
	files: { title: 'Files', color: '#5D4037', icon: 'folder-open' },
	help: { title: 'Help', color: '#607D8B', icon: 'help-circle' }
};

// Default state
const defaultState: Desktop3DState = {
	viewMode: 'orb',
	windows: [],
	focusedWindowId: null,
	sphereRadius: 95,  // Expanded radius for more spacing between windows
	gridColumns: 4,
	gridSpacing: 130,
	autoRotate: true,
	animating: false
};

// RING-BASED SPHERE LAYOUT - structured layers like a geodesic dome
// Creates 3 rings: top (3), middle (6), bottom (3) = 12 total
// For different totals, distributes proportionally

interface RingConfig {
	y: number;        // Height on sphere (-1 to 1)
	count: number;    // Windows in this ring
	startIndex: number; // Global start index
}

function getRingLayout(total: number): RingConfig[] {
	// 3-RING SPHERE LAYOUT: top, middle, bottom
	if (total <= 3) {
		return [{ y: 0, count: total, startIndex: 0 }];
	}
	if (total <= 6) {
		const top = Math.floor(total / 2);
		const bottom = total - top;
		return [
			{ y: 0.5, count: top, startIndex: 0 },
			{ y: -0.5, count: bottom, startIndex: top }
		];
	}
	if (total <= 9) {
		const middle = Math.ceil(total / 2);
		const remaining = total - middle;
		const top = Math.ceil(remaining / 2);
		const bottom = remaining - top;
		return [
			{ y: 0.6, count: top, startIndex: 0 },
			{ y: 0, count: middle, startIndex: top },
			{ y: -0.6, count: bottom, startIndex: top + middle }
		];
	}
	// 10+ modules: 3 top, N middle, 3 bottom
	const top = 3;
	const bottom = 3;
	const middle = total - top - bottom;
	return [
		{ y: 0.6, count: top, startIndex: 0 },
		{ y: 0, count: middle, startIndex: top },
		{ y: -0.6, count: bottom, startIndex: top + middle }
	];
}

function getPositionInRing(
	indexInRing: number,
	ringCount: number,
	ringY: number,
	radius: number,
	ringIndex: number // Which ring (0, 1, 2) for offset
): [number, number, number] {
	// Offset each ring's starting angle so they don't align vertically
	const ringOffset = (ringIndex * Math.PI) / 3; // 60 degree offset per ring

	// Calculate angle for this position in the ring
	const angle = ringOffset + (indexInRing / ringCount) * Math.PI * 2;

	// Y position from ring config
	const y = ringY * radius;

	// Radius at this height (smaller at poles, larger at equator)
	// Use a gentler curve so top/bottom rings aren't too small
	const heightFactor = Math.abs(ringY);
	const radiusAtY = Math.sqrt(1 - heightFactor * heightFactor * 0.5) * radius;

	const x = Math.cos(angle) * radiusAtY;
	const z = Math.sin(angle) * radiusAtY;

	return [x, y, z];
}

// Calculate position using structured ring layout
function calculateOrbPosition(
	index: number,
	total: number,
	radius: number,
	moduleId: string
): [number, number, number] {
	const rings = getRingLayout(total);

	// Find which ring this index belongs to
	for (let ringIndex = 0; ringIndex < rings.length; ringIndex++) {
		const ring = rings[ringIndex];
		if (index < ring.startIndex + ring.count) {
			const indexInRing = index - ring.startIndex;
			return getPositionInRing(indexInRing, ring.count, ring.y, radius, ringIndex);
		}
	}

	// Fallback to last ring
	const lastRing = rings[rings.length - 1];
	return getPositionInRing(0, lastRing.count, lastRing.y, radius, rings.length - 1);
}

// Calculate grid position
function calculateGridPosition(
	index: number,
	total: number,
	columns: number,
	spacing: number
): [number, number, number] {
	const rows = Math.ceil(total / columns);
	const col = index % columns;
	const row = Math.floor(index / columns);

	const offsetX = ((Math.min(total, columns) - 1) * spacing) / 2;
	const offsetY = ((rows - 1) * spacing) / 2;

	const x = col * spacing - offsetX;
	const y = row * -spacing + offsetY;
	const z = 0;

	return [x, y, z];
}

// Create the store
function createDesktop3DStore() {
	const { subscribe, set, update } = writable<Desktop3DState>(defaultState);

	return {
		subscribe,

		// Initialize with core modules
		initialize: () => {
			update((state) => {
				const windows: Window3DState[] = CORE_MODULES.map((module, index) => {
					const info = MODULE_INFO[module];
					const position = calculateOrbPosition(
						index,
						CORE_MODULES.length,
						state.sphereRadius,
						module
					);

					return {
						id: `window-${module}`,
						module,
						title: info.title,
						position,
						targetPosition: position,
						rotation: [0, 0, 0],
						scale: 1,
						targetScale: 1,
						opacity: 1,
						targetOpacity: 1,
						isCore: true,
						isOpen: true,
						isFocused: false,
						lastFocused: Date.now(),
						color: info.color,
						width: 1300,
						height: 900
					};
				});

				return { ...state, windows };
			});
		},

		// Recalculate all positions based on current view mode
		recalculatePositions: () => {
			update((state) => {
				const openWindows = state.windows.filter((w) => w.isOpen);
				const windows = state.windows.map((window) => {
					if (!window.isOpen) return window;

					const index = openWindows.findIndex((w) => w.id === window.id);
					let newPosition: [number, number, number];

					if (state.viewMode === 'grid') {
						newPosition = calculateGridPosition(
							index,
							openWindows.length,
							state.gridColumns,
							state.gridSpacing
						);
					} else {
						newPosition = calculateOrbPosition(
							index,
							openWindows.length,
							state.sphereRadius,
							window.module
						);
					}

					// Update BOTH position and targetPosition so windows actually move
					return { ...window, position: newPosition, targetPosition: newPosition };
				});

				return { ...state, windows };
			});
		},

		// Set view mode
		setViewMode: (mode: ViewMode) => {
			update((state) => {
				if (mode === state.viewMode) return state;

				// If going to focused mode without a focused window, don't change
				if (mode === 'focused' && !state.focusedWindowId) return state;

				const newState = { ...state, viewMode: mode, animating: true };

				// Recalculate positions after state update
				setTimeout(() => {
					desktop3dStore.recalculatePositions();
					setTimeout(() => {
						update((s) => ({ ...s, animating: false }));
					}, 500);
				}, 0);

				return newState;
			});
		},

		// Toggle between orb and grid
		toggleViewMode: () => {
			update((state) => {
				const newMode = state.viewMode === 'orb' ? 'grid' : 'orb';
				return { ...state, viewMode: newMode };
			});
			desktop3dStore.recalculatePositions();
		},

		// Focus on a window
		focusWindow: (windowId: string) => {
			update((state) => {
				const windows = state.windows.map((w) => ({
					...w,
					isFocused: w.id === windowId,
					targetOpacity: w.id === windowId ? 1 : 0.3,
					targetScale: w.id === windowId ? 1.5 : 0.8,
					lastFocused: w.id === windowId ? Date.now() : w.lastFocused
				}));

				return {
					...state,
					windows,
					focusedWindowId: windowId,
					viewMode: 'focused',
					autoRotate: false
				};
			});
		},

		// Unfocus current window
		unfocusWindow: () => {
			update((state) => {
				const windows = state.windows.map((w) => ({
					...w,
					isFocused: false,
					targetOpacity: 1,
					targetScale: 1
				}));

				return {
					...state,
					windows,
					focusedWindowId: null,
					viewMode: 'orb'
				};
			});
			desktop3dStore.recalculatePositions();
		},

		// Open a module window
		openWindow: (module: ModuleId) => {
			update((state) => {
				// Check if already open
				const existing = state.windows.find((w) => w.module === module);
				if (existing?.isOpen) {
					// Focus it instead
					desktop3dStore.focusWindow(existing.id);
					return state;
				}

				const info = MODULE_INFO[module];
				const openWindows = state.windows.filter((w) => w.isOpen);
				const position = calculateOrbPosition(
					openWindows.length,
					openWindows.length + 1,
					state.sphereRadius,
					module
				);

				if (existing) {
					// Reopen existing window
					const windows = state.windows.map((w) =>
						w.module === module
							? {
									...w,
									isOpen: true,
									position,
									targetPosition: position,
									lastFocused: Date.now()
								}
							: w
					);
					return { ...state, windows };
				}

				// Create new window
				const newWindow: Window3DState = {
					id: `window-${module}-${Date.now()}`,
					module,
					title: info.title,
					position,
					targetPosition: position,
					rotation: [0, 0, 0],
					scale: 1,
					targetScale: 1,
					opacity: 1,
					targetOpacity: 1,
					isCore: CORE_MODULES.includes(module as any),
					isOpen: true,
					isFocused: false,
					lastFocused: Date.now(),
					color: info.color,
					width: 800,  // Default window width
					height: 600  // Default window height
				};

				return { ...state, windows: [...state.windows, newWindow] };
			});
			desktop3dStore.recalculatePositions();
		},

		// Close a window (core modules just minimize)
		closeWindow: (windowId: string) => {
			update((state) => {
				const window = state.windows.find((w) => w.id === windowId);
				if (!window) return state;

				// Core modules can't be closed
				if (window.isCore) return state;

				const windows = state.windows.map((w) =>
					w.id === windowId ? { ...w, isOpen: false } : w
				);

				return {
					...state,
					windows,
					focusedWindowId:
						state.focusedWindowId === windowId ? null : state.focusedWindowId
				};
			});
			desktop3dStore.recalculatePositions();
		},

		// Toggle auto-rotate
		toggleAutoRotate: () => {
			update((state) => ({ ...state, autoRotate: !state.autoRotate }));
		},

		// Set auto-rotate
		setAutoRotate: (enabled: boolean) => {
			update((state) => ({ ...state, autoRotate: enabled }));
		},

		// Reset to default state
		reset: () => {
			set(defaultState);
			desktop3dStore.initialize();
		},

		// Navigate to next window (when focused)
		focusNext: () => {
			update((state) => {
				if (!state.focusedWindowId) return state;

				const openWins = state.windows.filter((w) => w.isOpen);
				const currentIndex = openWins.findIndex((w) => w.id === state.focusedWindowId);
				const nextIndex = (currentIndex + 1) % openWins.length;
				const nextWindow = openWins[nextIndex];

				if (!nextWindow) return state;

				const windows = state.windows.map((w) => ({
					...w,
					isFocused: w.id === nextWindow.id,
					targetOpacity: w.id === nextWindow.id ? 1 : 0.3,
					targetScale: w.id === nextWindow.id ? 1.5 : 0.8,
					lastFocused: w.id === nextWindow.id ? Date.now() : w.lastFocused
				}));

				return { ...state, windows, focusedWindowId: nextWindow.id };
			});
		},

		// Navigate to previous window (when focused)
		focusPrevious: () => {
			update((state) => {
				if (!state.focusedWindowId) return state;

				const openWins = state.windows.filter((w) => w.isOpen);
				const currentIndex = openWins.findIndex((w) => w.id === state.focusedWindowId);
				const prevIndex = (currentIndex - 1 + openWins.length) % openWins.length;
				const prevWindow = openWins[prevIndex];

				if (!prevWindow) return state;

				const windows = state.windows.map((w) => ({
					...w,
					isFocused: w.id === prevWindow.id,
					targetOpacity: w.id === prevWindow.id ? 1 : 0.3,
					targetScale: w.id === prevWindow.id ? 1.5 : 0.8,
					lastFocused: w.id === prevWindow.id ? Date.now() : w.lastFocused
				}));

				return { ...state, windows, focusedWindowId: prevWindow.id };
			});
		},

		// Resize focused window
		resizeFocusedWindow: (widthDelta: number, heightDelta: number) => {
			update((state) => {
				if (!state.focusedWindowId) return state;

				const windows = state.windows.map((w) => {
					if (w.id !== state.focusedWindowId) return w;
					return {
						...w,
						width: Math.max(800, Math.min(1600, w.width + widthDelta)),
						height: Math.max(500, Math.min(1100, w.height + heightDelta))
					};
				});

				return { ...state, windows };
			});
		}
	};
}

export const desktop3dStore = createDesktop3DStore();

// Derived store for open windows only
export const openWindows = derived(desktop3dStore, ($store) =>
	$store.windows.filter((w) => w.isOpen)
);

// Derived store for focused window
export const focusedWindow = derived(desktop3dStore, ($store) =>
	$store.windows.find((w) => w.id === $store.focusedWindowId)
);
