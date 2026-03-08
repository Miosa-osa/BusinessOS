// Window State Store - Window lifecycle, positioning, z-order, focus, minimize/maximize
// This is the primary store that composes all domain-specific methods.
// It exports windowStore, focusedWindow, visibleWindows, minimizedWindows, openModules.

import { writable, derived } from "svelte/store";
import { soundStore } from "./soundStore";
import {
  loadSavedSettings,
  saveSettings,
  initialDesktopIcons,
} from "./desktopPersistence";
import type { WindowState, WindowStoreShape } from "./desktopTypes";
import { moduleDefaults, createModuleMethods } from "./windowModuleStore";
import { createSnapMethods } from "./windowSnapStore";
import { createIconMethods } from "./windowIconStore";
import { createSerializationMethods } from "./windowSerializationStore";

const initialState: WindowStoreShape = {
  windows: [],
  focusedWindowId: null,
  windowOrder: [],
  dockPinnedItems: [
    "finder",
    "dashboard",
    "chat",
    "projects",
    "tasks",
    "clients",
  ],
  desktopIcons: initialDesktopIcons,
  selectedIconIds: [],
  folders: [],
};

function createWindowStore() {
  // Merge initial state with any saved settings
  const savedSettings = loadSavedSettings(initialState.dockPinnedItems);
  const mergedInitial: WindowStoreShape = {
    ...initialState,
    ...savedSettings,
  };

  const { subscribe, set, update } = writable<WindowStoreShape>(mergedInitial);

  let cascadeOffset = 0;
  let initialized = false;

  // Compose domain-specific method groups
  const snapMethods = createSnapMethods(update);
  const iconMethods = createIconMethods(update, subscribe);
  const moduleMethods = createModuleMethods(update);
  const serializationMethods = createSerializationMethods(
    update,
    set,
    subscribe,
    initialState,
  );

  return {
    subscribe,

    // Initialize store on client side (call this in onMount)
    initialize: () => {
      if (initialized || typeof window === "undefined") return;
      initialized = true;

      const saved = loadSavedSettings(initialState.dockPinnedItems);

      update((state) => {
        // Get dock items from saved settings or current state
        let dockItems =
          Object.keys(saved).length > 0 && saved.dockPinnedItems
            ? saved.dockPinnedItems
            : state.dockPinnedItems;

        // Filter out stale dock items that don't match any known module
        // This prevents ghost entries (e.g. "app-store") from persisting in localStorage
        const knownModules = new Set(Object.keys(moduleDefaults));
        dockItems = dockItems.filter((item) => knownModules.has(item));

        // ALWAYS ensure Finder is at the beginning of dock
        if (!dockItems.includes("finder")) {
          dockItems = ["finder", ...dockItems];
        } else if (dockItems[0] !== "finder") {
          // Move finder to first position
          dockItems = [
            "finder",
            ...dockItems.filter((item) => item !== "finder"),
          ];
        }

        // Get desktop icons - use saved if available, otherwise keep current state (initial)
        // Filter out stale osa-app-* icons — they are re-registered dynamically
        // by deployedAppsStore.startDiscovery() from the live backend API
        const rawDesktopIcons =
          Object.keys(saved).length > 0 &&
          saved.desktopIcons &&
          saved.desktopIcons.length > 0
            ? saved.desktopIcons
            : state.desktopIcons;
        const desktopIcons = rawDesktopIcons.filter(
          (icon) => !icon.module.startsWith("osa-app-"),
        );

        // Get folders
        const folders =
          Object.keys(saved).length > 0 && saved.folders
            ? saved.folders
            : state.folders;

        // Filter out stale windows for unknown modules
        const windows = state.windows.filter(
          (w) =>
            knownModules.has(w.module) ||
            w.module.startsWith("folder-") ||
            w.module.startsWith("osa-app-"),
        );
        const windowIds = new Set(windows.map((w) => w.id));
        const windowOrder = state.windowOrder.filter((id) => windowIds.has(id));

        const newState = {
          ...state,
          dockPinnedItems: dockItems,
          desktopIcons,
          folders,
          windows,
          windowOrder,
        };

        // Save the updated settings
        saveSettings(newState);

        return newState;
      });
    },

    // Open a new window for a module
    openWindow: (
      module: string,
      options?: string | { title?: string; data?: Record<string, unknown> },
    ) => {
      // Handle legacy string parameter (custom title)
      const customTitle =
        typeof options === "string" ? options : options?.title;
      const windowData =
        typeof options === "object" ? options?.data : undefined;

      update((state) => {
        // Check if window already exists for this module
        const existingWindow = state.windows.find(
          (w) => w.module === module && !w.minimized,
        );
        if (existingWindow) {
          // Update data if provided and focus window
          const updatedWindows = windowData
            ? state.windows.map((w) =>
                w.id === existingWindow.id
                  ? { ...w, data: { ...w.data, ...windowData } }
                  : w,
              )
            : state.windows;
          return {
            ...state,
            windows: updatedWindows,
            focusedWindowId: existingWindow.id,
            windowOrder: [
              ...state.windowOrder.filter((id) => id !== existingWindow.id),
              existingWindow.id,
            ],
          };
        }

        // Check if minimized window exists
        const minimizedWindow = state.windows.find(
          (w) => w.module === module && w.minimized,
        );
        if (minimizedWindow) {
          // Restore it and update data if provided
          return {
            ...state,
            windows: state.windows.map((w) =>
              w.id === minimizedWindow.id
                ? {
                    ...w,
                    minimized: false,
                    data: windowData ? { ...w.data, ...windowData } : w.data,
                  }
                : w,
            ),
            focusedWindowId: minimizedWindow.id,
            windowOrder: [
              ...state.windowOrder.filter((id) => id !== minimizedWindow.id),
              minimizedWindow.id,
            ],
          };
        }

        const defaults = moduleDefaults[module] || {
          title: module,
          width: 800,
          height: 600,
          minWidth: 400,
          minHeight: 300,
        };
        const id = `${module}-${Date.now()}`;

        // Calculate cascade position
        const baseX = 100 + cascadeOffset * 30;
        const baseY = 50 + cascadeOffset * 30;
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
        soundStore.playSound("windowOpen");

        return {
          ...state,
          windows: [...state.windows, newWindow],
          focusedWindowId: id,
          windowOrder: [...state.windowOrder, id],
        };
      });
    },

    // Close a window
    closeWindow: (windowId: string) => {
      // Play window close sound
      soundStore.playSound("windowClose");
      update((state) => {
        const newWindows = state.windows.filter((w) => w.id !== windowId);
        const newOrder = state.windowOrder.filter((id) => id !== windowId);
        const newFocused =
          state.focusedWindowId === windowId
            ? newOrder.length > 0
              ? newOrder[newOrder.length - 1]
              : null
            : state.focusedWindowId;

        return {
          ...state,
          windows: newWindows,
          windowOrder: newOrder,
          focusedWindowId: newFocused,
        };
      });
    },

    // Minimize a window
    minimizeWindow: (windowId: string) => {
      // Play minimize sound
      soundStore.playSound("windowMinimize");
      update((state) => {
        const newOrder = state.windowOrder.filter((id) => id !== windowId);
        const newFocused =
          state.focusedWindowId === windowId
            ? newOrder.length > 0
              ? newOrder[newOrder.length - 1]
              : null
            : state.focusedWindowId;

        return {
          ...state,
          windows: state.windows.map((w) =>
            w.id === windowId ? { ...w, minimized: true } : w,
          ),
          windowOrder: newOrder,
          focusedWindowId: newFocused,
        };
      });
    },

    // Restore a minimized window
    restoreWindow: (windowId: string) => {
      update((state) => ({
        ...state,
        windows: state.windows.map((w) =>
          w.id === windowId ? { ...w, minimized: false } : w,
        ),
        focusedWindowId: windowId,
        windowOrder: [
          ...state.windowOrder.filter((id) => id !== windowId),
          windowId,
        ],
      }));
    },

    // Toggle maximize state
    toggleMaximize: (windowId: string) => {
      // Play maximize sound
      soundStore.playSound("windowMaximize");
      update((state) => ({
        ...state,
        windows: state.windows.map((w) => {
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
              previousBounds: undefined,
            };
          } else {
            // Store current bounds and maximize
            return {
              ...w,
              maximized: true,
              snapped: null,
              previousBounds: {
                x: w.x,
                y: w.y,
                width: w.width,
                height: w.height,
              },
            };
          }
        }),
      }));
    },

    // Focus a window (bring to front)
    focusWindow: (windowId: string) => {
      update((state) => ({
        ...state,
        focusedWindowId: windowId,
        windowOrder: [
          ...state.windowOrder.filter((id) => id !== windowId),
          windowId,
        ],
      }));
    },

    // Update window position
    updateWindowPosition: (windowId: string, x: number, y: number) => {
      update((state) => ({
        ...state,
        windows: state.windows.map((w) =>
          w.id === windowId ? { ...w, x, y, maximized: false } : w,
        ),
      }));
    },

    // Update window size
    updateWindowSize: (windowId: string, width: number, height: number) => {
      update((state) => ({
        ...state,
        windows: state.windows.map((w) =>
          w.id === windowId
            ? {
                ...w,
                width: Math.max(width, w.minWidth),
                height: Math.max(height, w.minHeight),
                maximized: false,
              }
            : w,
        ),
      }));
    },

    // Update window bounds (position and size)
    updateWindowBounds: (
      windowId: string,
      x: number,
      y: number,
      width: number,
      height: number,
    ) => {
      update((state) => ({
        ...state,
        windows: state.windows.map((w) =>
          w.id === windowId
            ? {
                ...w,
                x,
                y,
                width: Math.max(width, w.minWidth),
                height: Math.max(height, w.minHeight),
                maximized: false,
              }
            : w,
        ),
      }));
    },

    // Cycle to next window (for Cmd+`)
    cycleWindows: () => {
      update((state) => {
        const visibleWindows = state.windows.filter((w) => !w.minimized);
        if (visibleWindows.length < 2) return state;

        const visibleOrder = state.windowOrder.filter((id) =>
          state.windows.find((w) => w.id === id && !w.minimized),
        );

        if (visibleOrder.length < 2) return state;

        const currentVisibleIndex = visibleOrder.indexOf(
          state.focusedWindowId || "",
        );
        const nextIndex = (currentVisibleIndex + 1) % visibleOrder.length;
        const nextWindowId = visibleOrder[nextIndex];

        return {
          ...state,
          focusedWindowId: nextWindowId,
          windowOrder: [
            ...state.windowOrder.filter((id) => id !== nextWindowId),
            nextWindowId,
          ],
        };
      });
    },

    // Get windows for a specific module (for dock indicator)
    getWindowsForModule: (module: string) => {
      let result: WindowState[] = [];
      const unsubscribe = subscribe((state) => {
        result = state.windows.filter((w) => w.module === module);
      });
      unsubscribe();
      return result;
    },

    // --- Snap methods (from windowSnapStore) ---
    ...snapMethods,

    // --- Icon and folder methods (from windowIconStore) ---
    ...iconMethods,

    // --- Module registry methods (from windowModuleStore) ---
    ...moduleMethods,

    // --- Serialization methods (from windowSerializationStore) ---
    ...serializationMethods,
  };
}

export const windowStore = createWindowStore();

// Derived stores
export const focusedWindow = derived(
  windowStore,
  ($store) =>
    $store.windows.find((w) => w.id === $store.focusedWindowId) || null,
);

export const visibleWindows = derived(windowStore, ($store) =>
  $store.windows.filter((w) => !w.minimized),
);

export const minimizedWindows = derived(windowStore, ($store) =>
  $store.windows.filter((w) => w.minimized),
);

export const openModules = derived(windowStore, ($store) => [
  ...new Set($store.windows.map((w) => w.module)),
]);
