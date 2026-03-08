/**
 * Terminal App Type Definitions
 * Supports multi-tab, split-pane terminal with AI provider switching
 */

// AI providers available in terminal tabs
export type TerminalProvider = "shell" | "osa" | "claude" | "codex" | "ollama";

// What a pane renders
export type PaneMode = "shell" | "ai" | "monaco";

// Split direction
export type SplitDirection = "horizontal" | "vertical";

// Terminal cursor styles
export type CursorStyle = "block" | "underline" | "bar";

// Environment modes for terminal sessions
export type EnvironmentMode = "sandbox" | "production" | "local";

// Environment info for a terminal session
export interface EnvironmentInfo {
  mode: EnvironmentMode;
  os?: string; // "darwin" | "linux" | "windows" (local mode)
  containerized?: boolean;
  agentProcess?: string; // "claude" | "codex" | null (running agent CLI)
  changeCount?: number; // tracked edits in sandbox mode
  branchName?: string; // sandbox git branch
}

// Provider display config
export interface ProviderConfig {
  id: TerminalProvider;
  label: string;
  icon: string; // SVG path or icon name
  color: string;
  shortcut?: string;
}

// Terminal visual config
export interface TerminalConfig {
  fontFamily: string;
  fontSize: number;
  theme: string;
  cursorStyle: CursorStyle;
  cursorBlink: boolean;
}

// A terminal tab
export interface TerminalTab {
  id: string;
  title: string;
  provider: TerminalProvider;
  paneMode: PaneMode;
  rootPaneId: string;
  sessionId: string | null;
  isActive: boolean;
  environment?: EnvironmentInfo;
}

// Binary tree node for split panes
export type PaneNode = PaneLeaf | PaneSplit;

export interface PaneLeaf {
  type: "leaf";
  id: string;
  mode: PaneMode;
  provider: TerminalProvider;
  sessionId: string | null;
  // AI-specific
  focusMode?: string;
  // Monaco-specific
  filePath?: string;
  fileContent?: string;
  environment?: EnvironmentMode;
}

export interface PaneSplit {
  type: "split";
  id: string;
  direction: SplitDirection;
  ratio: number; // 0-1, position of divider
  children: [PaneNode, PaneNode];
}

// Full terminal state
export interface TerminalState {
  tabs: TerminalTab[];
  activeTabId: string | null;
  panes: Record<string, PaneNode>; // rootPaneId -> PaneNode tree
  config: TerminalConfig;
}

// Provider configs with display info
export const PROVIDER_CONFIGS: ProviderConfig[] = [
  {
    id: "shell",
    label: "Shell",
    icon: "terminal",
    color: "#00ff00",
    shortcut: "1",
  },
  {
    id: "osa",
    label: "OSA",
    icon: "brain",
    color: "#8b5cf6",
    shortcut: "2",
  },
  {
    id: "claude",
    label: "Claude",
    icon: "sparkles",
    color: "#d97706",
    shortcut: "3",
  },
  {
    id: "codex",
    label: "Codex",
    icon: "code",
    color: "#10b981",
    shortcut: "4",
  },
  {
    id: "ollama",
    label: "Ollama",
    icon: "cpu",
    color: "#3b82f6",
    shortcut: "5",
  },
];

// Default terminal config
export const DEFAULT_CONFIG: TerminalConfig = {
  fontFamily:
    '"SF Mono", "Monaco", "Inconsolata", "Fira Code", "Courier New", monospace',
  fontSize: 14,
  theme: "dark",
  cursorStyle: "block",
  cursorBlink: true,
};
