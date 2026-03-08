// Desktop Persistence - localStorage save/load and JSON config export/import
import { browser } from "$app/environment";
import type {
  WindowStoreShape,
  DesktopIcon,
  DesktopConfig,
} from "./desktopTypes";

// Storage key for persisting desktop settings
export const STORAGE_KEY = "businessos_desktop_settings";

// Desktop config schema version for backwards compatibility
export const CONFIG_VERSION = "1.0.0";

// Initial desktop icon positions (right side, top to bottom)
export const initialDesktopIcons: DesktopIcon[] = [
  { id: "icon-platform", module: "platform", label: "Business OS", x: 0, y: 0 }, // Top left - full platform
  { id: "icon-terminal", module: "terminal", label: "Terminal", x: -1, y: 0 },
  {
    id: "icon-dashboard",
    module: "dashboard",
    label: "Dashboard",
    x: -1,
    y: 1,
  },
  { id: "icon-chat", module: "chat", label: "Chat", x: -1, y: 2 },
  { id: "icon-tasks", module: "tasks", label: "Tasks", x: -1, y: 3 },
  { id: "icon-projects", module: "projects", label: "Projects", x: -1, y: 4 },
  { id: "icon-team", module: "team", label: "Team", x: -1, y: 5 },
  { id: "icon-clients", module: "clients", label: "Clients", x: -1, y: 6 },
  { id: "icon-tables", module: "tables", label: "Tables", x: -1, y: 7 },
  {
    id: "icon-communication",
    module: "communication",
    label: "Communication",
    x: -1,
    y: 8,
  },
  { id: "icon-files", module: "files", label: "Files", x: -2, y: 0 },
  { id: "icon-pages", module: "pages", label: "Pages", x: -2, y: 1 },
  { id: "icon-nodes", module: "nodes", label: "Nodes", x: -2, y: 2 },
  { id: "icon-daily", module: "daily", label: "Daily Log", x: -2, y: 3 },
  { id: "icon-settings", module: "settings", label: "Settings", x: -2, y: 4 },
  {
    id: "icon-ai-settings",
    module: "ai-settings",
    label: "AI Settings",
    x: -2,
    y: 5,
  },
  {
    id: "icon-integrations",
    module: "integrations",
    label: "Integrations",
    x: -2,
    y: 6,
  },
  { id: "icon-trash", module: "trash", label: "Trash", x: -1, y: -1 }, // Bottom right
];

// Load saved desktop settings from localStorage
export function loadSavedSettings(
  initialDockPinnedItems: string[],
): Partial<WindowStoreShape> {
  if (!browser) return {};

  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      // Ensure we have valid arrays
      let desktopIcons =
        Array.isArray(parsed.desktopIcons) && parsed.desktopIcons.length > 0
          ? parsed.desktopIcons
          : initialDesktopIcons;
      const dockPinnedItems =
        Array.isArray(parsed.dockPinnedItems) &&
        parsed.dockPinnedItems.length > 0
          ? parsed.dockPinnedItems
          : initialDockPinnedItems;
      const folders = Array.isArray(parsed.folders) ? parsed.folders : [];

      // Merge in any new default icons that were added since last save
      // This ensures new modules (like integrations) appear on existing users' desktops
      const savedIconIds = new Set(desktopIcons.map((i: DesktopIcon) => i.id));
      const newIcons = initialDesktopIcons.filter(
        (icon) => !savedIconIds.has(icon.id),
      );
      if (newIcons.length > 0) {
        desktopIcons = [...desktopIcons, ...newIcons];
      }

      // Strip icons with unknown modules (stale integration icons like
      // "Figma", "Slack", "Notion", "GitHub" or osa-app-* from previous sessions)
      const knownModules = new Set(initialDesktopIcons.map((i) => i.module));
      desktopIcons = desktopIcons.filter(
        (icon: DesktopIcon) =>
          knownModules.has(icon.module) || icon.module === "finder",
      );

      // Update labels for existing icons to match defaults (preserve user positions)
      // This ensures label renames like "Contexts" -> "Knowledge" are applied
      const defaultLabels = new Map(
        initialDesktopIcons.map((i) => [i.id, i.label]),
      );
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
    console.error("Failed to load desktop settings:", e);
  }
  return {};
}

// Save desktop settings to localStorage
export function saveSettings(state: WindowStoreShape) {
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
    console.error("Failed to save desktop settings:", e);
  }
}

// Build an exportable DesktopConfig snapshot from current state
export function buildExportConfig(state: WindowStoreShape): DesktopConfig {
  return {
    version: CONFIG_VERSION,
    exportedAt: new Date().toISOString(),
    desktopIcons: state.desktopIcons,
    dockPinnedItems: state.dockPinnedItems,
    folders: state.folders,
  };
}

// Validate an imported DesktopConfig object
export function validateImportConfig(
  config: unknown,
): { valid: true; data: DesktopConfig } | { valid: false; error: string } {
  if (!config || typeof config !== "object") {
    return { valid: false, error: "Invalid configuration format" };
  }
  const c = config as Record<string, unknown>;
  if (!Array.isArray(c.desktopIcons)) {
    return { valid: false, error: "Missing or invalid desktopIcons" };
  }
  if (!Array.isArray(c.dockPinnedItems)) {
    return { valid: false, error: "Missing or invalid dockPinnedItems" };
  }
  // Validate each icon has required fields
  for (const icon of c.desktopIcons as Record<string, unknown>[]) {
    if (
      !icon.id ||
      !icon.module ||
      !icon.label ||
      typeof icon.x !== "number" ||
      typeof icon.y !== "number"
    ) {
      return { valid: false, error: "Invalid icon structure" };
    }
  }
  return { valid: true, data: c as unknown as DesktopConfig };
}

// JSON schema for desktop config (for documentation/tooling)
export const desktopConfigSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  title: "BusinessOS Desktop Configuration",
  type: "object",
  required: ["version", "desktopIcons", "dockPinnedItems"],
  properties: {
    version: { type: "string", description: "Config version" },
    exportedAt: {
      type: "string",
      format: "date-time",
      description: "Export timestamp",
    },
    desktopIcons: {
      type: "array",
      items: {
        type: "object",
        required: ["id", "module", "label", "x", "y"],
        properties: {
          id: { type: "string" },
          module: { type: "string" },
          label: { type: "string" },
          x: { type: "number" },
          y: { type: "number" },
          type: { type: "string", enum: ["app", "folder"] },
          folderId: { type: "string" },
          folderColor: { type: "string" },
        },
      },
    },
    dockPinnedItems: {
      type: "array",
      items: { type: "string" },
    },
    folders: {
      type: "array",
      items: {
        type: "object",
        required: ["id", "name", "color", "iconIds"],
        properties: {
          id: { type: "string" },
          name: { type: "string" },
          color: { type: "string" },
          iconIds: { type: "array", items: { type: "string" } },
        },
      },
    },
  },
};
