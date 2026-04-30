export interface NavItem {
  title: string;
  href: string;
}

export interface NavSection {
  title: string;
  items: NavItem[];
}

export const nav: NavSection[] = [
  {
    title: "",
    items: [{ title: "Quickstart", href: "/quickstart" }],
  },
  {
    title: "Concepts",
    items: [
      { title: "Three Tiers", href: "/concepts/three-tiers" },
      { title: "Nine Stages", href: "/concepts/nine-stages" },
      { title: "Signal Theory", href: "/concepts/signal-theory" },
      { title: "Workspaces", href: "/concepts/workspaces" },
      { title: "Memory Primitive", href: "/concepts/memory-primitive" },
      { title: "Proactive Surfacing", href: "/concepts/proactive-surfacing" },
    ],
  },
  {
    title: "API",
    items: [
      { title: "Overview", href: "/api" },
      { title: "Retrieval", href: "/api/retrieval" },
      { title: "Memory", href: "/api/memory" },
      { title: "Recall", href: "/api/recall" },
      { title: "Workspaces", href: "/api/workspaces" },
      { title: "Wiki", href: "/api/wiki" },
      { title: "Surfacing", href: "/api/surfacing" },
    ],
  },
  {
    title: "SDKs",
    items: [
      { title: "TypeScript", href: "/sdks/typescript" },
      { title: "Python", href: "/sdks/python" },
      { title: "MCP Server", href: "/sdks/mcp" },
    ],
  },
  {
    title: "Extensions",
    items: [
      { title: "Browser", href: "/extensions/browser" },
      { title: "Raycast", href: "/extensions/raycast" },
    ],
  },
  {
    title: "",
    items: [{ title: "Self-host", href: "/self-host" }],
  },
];
