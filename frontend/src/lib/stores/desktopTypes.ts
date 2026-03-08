// Desktop Types - Shared types for the desktop environment
// Imported by windowStore.ts, desktopStore.ts (settings), and desktopPersistence.ts

export type SnapZone =
  | "left"
  | "right"
  | "top-left"
  | "top-right"
  | "bottom-left"
  | "bottom-right"
  | null;

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
  type: "lucide" | "custom";
  lucideName?: string; // e.g., 'Home', 'Settings' - name from lucide-svelte
  customSvg?: string; // Base64 or raw SVG string for custom icons
  foregroundColor?: string; // Override icon color
  backgroundColor?: string; // Override background color
}

export interface DesktopIcon {
  id: string;
  module: string;
  label: string;
  x: number;
  y: number;
  type?: "app" | "folder";
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

// Internal store shape — not exported to consumers directly
export interface WindowStoreShape {
  windows: WindowState[];
  focusedWindowId: string | null;
  windowOrder: string[]; // Z-index order, last is on top
  dockPinnedItems: string[];
  desktopIcons: DesktopIcon[];
  selectedIconIds: string[];
  folders: DesktopFolder[];
}

// Export config type for JSON export/import
export interface DesktopConfig {
  version: string;
  exportedAt: string;
  desktopIcons: DesktopIcon[];
  dockPinnedItems: string[];
  folders: DesktopFolder[];
}
