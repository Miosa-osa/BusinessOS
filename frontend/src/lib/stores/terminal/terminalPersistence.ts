/**
 * Terminal Persistence — save/restore tab layout and config to localStorage
 */

import type {
  TerminalTab,
  TerminalConfig,
  PaneNode,
  TerminalProvider,
  PaneMode,
  SplitDirection,
} from "./terminalTypes";
import { DEFAULT_CONFIG, PROVIDER_CONFIGS } from "./terminalTypes";

const STORAGE_KEY = "bos-terminal-state";
const CONFIG_KEY = "bos-terminal-config";

interface PersistedState {
  tabs: Array<{
    id: string;
    title: string;
    provider: string;
    paneMode: string;
    rootPaneId: string;
  }>;
  activeTabId: string | null;
  panes: Record<string, unknown>;
}

const VALID_PROVIDERS = new Set<string>(PROVIDER_CONFIGS.map((p) => p.id));
const VALID_MODES = new Set<string>(["shell", "ai", "monaco"]);
const VALID_DIRECTIONS = new Set<string>(["horizontal", "vertical"]);

/**
 * Save terminal layout (tabs + active tab + pane tree structure)
 * Note: WebSocket sessions are NOT persisted — shells reconnect on restore
 */
export function saveTerminalLayout(
  tabs: TerminalTab[],
  activeTabId: string | null,
  panes: Record<string, PaneNode>,
): void {
  try {
    const state: PersistedState = {
      tabs: tabs.map((t) => ({
        id: t.id,
        title: t.title,
        provider: t.provider,
        paneMode: t.paneMode,
        rootPaneId: t.rootPaneId,
      })),
      activeTabId,
      panes: serializePanes(panes),
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } catch {
    // localStorage full or unavailable — silently ignore
  }
}

/**
 * Restore terminal layout from localStorage
 * Returns null if no saved state or invalid data
 */
export function restoreTerminalLayout(): PersistedState | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as PersistedState;

    // Validate structure
    if (!parsed || typeof parsed !== "object") return null;
    if (!Array.isArray(parsed.tabs) || parsed.tabs.length === 0) return null;

    // Validate each tab
    for (const tab of parsed.tabs) {
      if (!tab.id || typeof tab.id !== "string") return null;
      if (!tab.rootPaneId || typeof tab.rootPaneId !== "string") return null;
      if (!VALID_PROVIDERS.has(tab.provider)) return null;
      if (!VALID_MODES.has(tab.paneMode)) return null;
    }

    // Validate activeTabId references a real tab
    if (
      parsed.activeTabId &&
      !parsed.tabs.some((t) => t.id === parsed.activeTabId)
    ) {
      parsed.activeTabId = parsed.tabs[0].id;
    }

    // Validate panes exist for each tab
    if (!parsed.panes || typeof parsed.panes !== "object") return null;
    for (const tab of parsed.tabs) {
      if (!(tab.rootPaneId in (parsed.panes as Record<string, unknown>)))
        return null;
    }

    return parsed;
  } catch {
    return null;
  }
}

/**
 * Save terminal config (font, theme, cursor)
 */
export function saveTerminalConfig(config: TerminalConfig): void {
  try {
    localStorage.setItem(CONFIG_KEY, JSON.stringify(config));
  } catch {
    // silently ignore
  }
}

/**
 * Restore terminal config from localStorage
 */
export function restoreTerminalConfig(): TerminalConfig {
  try {
    const raw = localStorage.getItem(CONFIG_KEY);
    if (!raw) return DEFAULT_CONFIG;
    const parsed = JSON.parse(raw) as Partial<TerminalConfig>;

    // Validate individual fields
    const config = { ...DEFAULT_CONFIG };

    if (typeof parsed.fontFamily === "string") {
      config.fontFamily = parsed.fontFamily;
    }
    if (
      typeof parsed.fontSize === "number" &&
      parsed.fontSize >= 8 &&
      parsed.fontSize <= 32
    ) {
      config.fontSize = parsed.fontSize;
    }
    if (typeof parsed.theme === "string") {
      config.theme = parsed.theme;
    }
    if (
      parsed.cursorStyle === "block" ||
      parsed.cursorStyle === "underline" ||
      parsed.cursorStyle === "bar"
    ) {
      config.cursorStyle = parsed.cursorStyle;
    }
    if (typeof parsed.cursorBlink === "boolean") {
      config.cursorBlink = parsed.cursorBlink;
    }

    return config;
  } catch {
    return DEFAULT_CONFIG;
  }
}

/**
 * Clear all terminal persistence
 */
export function clearTerminalPersistence(): void {
  localStorage.removeItem(STORAGE_KEY);
  localStorage.removeItem(CONFIG_KEY);
}

// Serialize pane tree — strip non-serializable data (service instances, xterm refs)
function serializePanes(
  panes: Record<string, PaneNode>,
): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const [key, node] of Object.entries(panes)) {
    result[key] = serializePaneNode(node);
  }
  return result;
}

function serializePaneNode(node: PaneNode): unknown {
  if (node.type === "leaf") {
    return {
      type: "leaf",
      id: node.id,
      mode: node.mode,
      provider: node.provider,
      sessionId: null, // sessions don't survive reload
      filePath: node.filePath,
      fileContent: undefined, // don't persist file contents
    };
  }
  return {
    type: "split",
    id: node.id,
    direction: node.direction,
    ratio: node.ratio,
    children: [
      serializePaneNode(node.children[0]),
      serializePaneNode(node.children[1]),
    ],
  };
}
