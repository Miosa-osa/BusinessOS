// Window Module Store - Module configurations and module registry
// Contains moduleDefaults (the known module dimension/title configs) and
// the registerDeployedApp / unregisterDeployedApp methods for dynamic OSA apps.

import type { WindowStoreShape, DesktopIcon } from "./desktopTypes";
import { saveSettings } from "./desktopPersistence";

// Default module configurations
export const moduleDefaults: Record<
  string,
  {
    title: string;
    width: number;
    height: number;
    minWidth: number;
    minHeight: number;
  }
> = {
  platform: {
    title: "Business OS",
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
  },
  dashboard: {
    title: "Dashboard",
    width: 1000,
    height: 700,
    minWidth: 600,
    minHeight: 400,
  },
  chat: {
    title: "Chat",
    width: 900,
    height: 650,
    minWidth: 400,
    minHeight: 300,
  },
  tasks: {
    title: "Tasks",
    width: 800,
    height: 600,
    minWidth: 400,
    minHeight: 300,
  },
  projects: {
    title: "Projects",
    width: 900,
    height: 650,
    minWidth: 500,
    minHeight: 400,
  },
  team: {
    title: "Team",
    width: 850,
    height: 600,
    minWidth: 400,
    minHeight: 300,
  },
  clients: {
    title: "Clients",
    width: 1000,
    height: 700,
    minWidth: 600,
    minHeight: 400,
  },
  tables: {
    title: "Tables",
    width: 1100,
    height: 750,
    minWidth: 700,
    minHeight: 500,
  },
  pages: {
    title: "Pages",
    width: 900,
    height: 650,
    minWidth: 500,
    minHeight: 400,
  },
  contexts: {
    title: "Pages",
    width: 900,
    height: 650,
    minWidth: 500,
    minHeight: 400,
  }, // Legacy alias
  nodes: {
    title: "Nodes",
    width: 1000,
    height: 700,
    minWidth: 600,
    minHeight: 400,
  },
  daily: {
    title: "Daily Log",
    width: 700,
    height: 550,
    minWidth: 350,
    minHeight: 300,
  },
  settings: {
    title: "Settings",
    width: 700,
    height: 550,
    minWidth: 400,
    minHeight: 350,
  },
  communication: {
    title: "Communication",
    width: 1000,
    height: 700,
    minWidth: 600,
    minHeight: 450,
  },
  "ai-settings": {
    title: "AI Settings",
    width: 800,
    height: 600,
    minWidth: 500,
    minHeight: 400,
  },
  integrations: {
    title: "Integrations",
    width: 950,
    height: 700,
    minWidth: 600,
    minHeight: 500,
  },
  trash: {
    title: "Trash",
    width: 600,
    height: 450,
    minWidth: 300,
    minHeight: 250,
  },
  terminal: {
    title: "Terminal - OS Agent",
    width: 700,
    height: 500,
    minWidth: 400,
    minHeight: 300,
  },
  "desktop-settings": {
    title: "Desktop Settings",
    width: 550,
    height: 500,
    minWidth: 450,
    minHeight: 400,
  },
  folder: {
    title: "Folder",
    width: 600,
    height: 450,
    minWidth: 300,
    minHeight: 250,
  },
  files: {
    title: "Files",
    width: 900,
    height: 600,
    minWidth: 500,
    minHeight: 400,
  },
  finder: {
    title: "Finder",
    width: 900,
    height: 600,
    minWidth: 500,
    minHeight: 400,
  },
  help: {
    title: "Help",
    width: 900,
    height: 650,
    minWidth: 600,
    minHeight: 450,
  },
};

// Category-specific colors for deployed OSA apps
export const categoryColors: Record<string, { fg: string; bg: string }> = {
  finance: { fg: "#10b981", bg: "#d1fae5" },
  communication: { fg: "#3b82f6", bg: "#dbeafe" },
  productivity: { fg: "#a855f7", bg: "#f3e8ff" },
  analytics: { fg: "#f97316", bg: "#fed7aa" },
  ecommerce: { fg: "#ec4899", bg: "#fce7f3" },
  crm: { fg: "#06b6d4", bg: "#cffafe" },
  hr: { fg: "#6366f1", bg: "#e0e7ff" },
  inventory: { fg: "#f59e0b", bg: "#fef3c7" },
  marketing: { fg: "#f43f5e", bg: "#ffe4e6" },
  project: { fg: "#14b8a6", bg: "#ccfbf1" },
  general: { fg: "#8B5CF6", bg: "#F3E8FF" },
};

// Creates the module registry methods that operate on the window store
export function createModuleMethods(
  update: (fn: (state: WindowStoreShape) => WindowStoreShape) => void,
) {
  return {
    // Register a deployed OSA app dynamically
    registerDeployedApp: (app: {
      id: string;
      name: string;
      url: string;
      port: number;
      metadata?: {
        name: string;
        description: string;
        category: string;
        icon: string;
        keywords: string[];
      };
    }) => {
      update((state) => {
        const moduleId = `osa-app-${app.id}`;

        // Check if already registered
        if (moduleDefaults[moduleId]) {
          return state;
        }

        // Use metadata name or fallback to app name
        const displayName = app.metadata?.name || app.name;

        // Add to moduleDefaults
        moduleDefaults[moduleId] = {
          title: displayName,
          width: 1000,
          height: 700,
          minWidth: 600,
          minHeight: 400,
        };

        // Check if desktop icon already exists
        const iconExists = state.desktopIcons.some(
          (icon) => icon.module === moduleId,
        );
        if (iconExists) {
          return state;
        }

        // Find next available position on the right side
        const rightSideIcons = state.desktopIcons.filter(
          (icon) => icon.x === -1,
        );
        const nextY =
          rightSideIcons.length > 0
            ? Math.max(...rightSideIcons.map((icon) => icon.y)) + 1
            : 0;

        const category = app.metadata?.category || "general";
        const colors = categoryColors[category] || categoryColors.general;

        // Create desktop icon with metadata
        const newIcon: DesktopIcon = {
          id: `icon-${moduleId}`,
          module: moduleId,
          label: displayName,
          x: -1, // Right side
          y: nextY,
          type: "app",
          customIcon: {
            type: "lucide",
            lucideName: app.metadata?.icon || "AppWindow",
            foregroundColor: colors.fg,
            backgroundColor: colors.bg,
          },
        };

        const newState = {
          ...state,
          desktopIcons: [...state.desktopIcons, newIcon],
        };

        // Save to localStorage
        saveSettings(newState);

        if (import.meta.env.DEV) {
          console.log(
            `[windowStore] Registered OSA app: ${displayName} (${moduleId})`,
          );
        }

        return newState;
      });
    },

    // Unregister a deployed app when it stops
    unregisterDeployedApp: (appId: string) => {
      const moduleId = `osa-app-${appId}`;

      update((state) => {
        // Remove from moduleDefaults
        delete moduleDefaults[moduleId];

        // Remove desktop icon
        const newState = {
          ...state,
          desktopIcons: state.desktopIcons.filter(
            (icon) => icon.module !== moduleId,
          ),
          // Close any open windows for this app
          windows: state.windows.filter((w) => w.module !== moduleId),
          windowOrder: state.windowOrder.filter((id) => {
            const win = state.windows.find((w) => w.id === id);
            return win?.module !== moduleId;
          }),
        };

        saveSettings(newState);

        if (import.meta.env.DEV)
          console.log(`[windowStore] Unregistered OSA app: ${moduleId}`);

        return newState;
      });
    },
  };
}
