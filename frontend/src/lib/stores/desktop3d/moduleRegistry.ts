/**
 * Desktop 3D Module Registry
 * Static lookup tables for module metadata and dynamic module colors.
 * These are not reactive — they are plain records mutated only by addModule/removeModule.
 */

// Module metadata for built-in modules
export const MODULE_INFO: Record<
  string,
  { title: string; color: string; icon: string }
> = {
  dashboard: { title: "Dashboard", color: "#1E88E5", icon: "grid" },
  chat: { title: "Chat", color: "#43A047", icon: "message-circle" },
  tasks: { title: "Tasks", color: "#FB8C00", icon: "check-square" },
  projects: { title: "Projects", color: "#8E24AA", icon: "folder" },
  team: { title: "Team", color: "#00ACC1", icon: "users" },
  clients: { title: "Clients", color: "#5C6BC0", icon: "briefcase" },
  tables: { title: "Tables", color: "#6366F1", icon: "table" },
  communication: { title: "Communication", color: "#E53935", icon: "mail" },
  pages: { title: "Pages", color: "#7CB342", icon: "book" },
  nodes: { title: "Nodes", color: "#FF7043", icon: "share-2" },
  daily: { title: "Daily Log", color: "#26A69A", icon: "edit" },
  settings: { title: "Settings", color: "#78909C", icon: "settings" },
  terminal: { title: "Terminal", color: "#37474F", icon: "terminal" },
  help: { title: "Help", color: "#607D8B", icon: "help-circle" },
  agents: { title: "Agents", color: "#9C27B0", icon: "bot" },
  crm: { title: "CRM", color: "#00897B", icon: "building" },
  integrations: { title: "Integrations", color: "#3F51B5", icon: "plug" },
  knowledge: { title: "Knowledge", color: "#FF6F00", icon: "book-open" },
  notifications: { title: "Notifications", color: "#D32F2F", icon: "bell" },
  profile: { title: "Profile", color: "#0288D1", icon: "user" },
  "voice-notes": { title: "Voice Notes", color: "#C2185B", icon: "mic" },
  usage: { title: "Usage", color: "#455A64", icon: "bar-chart" },
};

// Category-based colors for dynamically added modules
export const DYNAMIC_MODULE_COLORS: Record<string, string> = {
  finance: "#10b981",
  communication: "#3b82f6",
  productivity: "#a855f7",
  analytics: "#f97316",
  ecommerce: "#ec4899",
  crm: "#06b6d4",
  hr: "#6366f1",
  inventory: "#f59e0b",
  marketing: "#f43f5e",
  project: "#14b8a6",
  general: "#8B5CF6",
};
