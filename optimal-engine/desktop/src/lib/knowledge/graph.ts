import type { WorkspaceNode } from "$lib/api/workspace";

export type GKind = "node" | "signal" | "entity";

export interface GNode {
  id: string;
  label: string;
  kind: GKind;
  sub: string;
  connections: number;
  node_slug?: string;
}

export interface GEdge {
  source: string;
  target: string;
  relation: "contains" | "mentions" | "parent";
}

export interface GraphData {
  nodes: GNode[];
  edges: GEdge[];
}

interface EngineGraphNode {
  node: string;
  context_count: number;
  last_modified: string;
}

interface EngineEdge {
  source: string;
  target: string;
  relation: string;
  weight: number;
}

interface EngineNodeDetail {
  node: string;
  contexts: {
    id: string;
    title: string;
    genre: string | null;
    sn_ratio: number | null;
    type: string | null;
  }[];
  edges: EngineEdge[];
}

const KIND_MAP: Record<string, string> = {
  founder: "person",
  platform: "org",
  services: "org",
  academy: "operation",
  partners: "org",
  media: "org",
  "getting-started": "concept",
  "new-stuff": "concept",
  businessos: "product",
  tasks: "operation",
  crm: "product",
  chat: "product",
  team: "org",
  clients: "org",
  "integration-test": "concept",
  inbox: "operation",
  "platform-core": "project",
  "platform-services": "project",
  "platform-investors": "project",
  "academy-beginner": "project",
  "academy-advanced": "project",
  "partners-healthtech": "project",
};

function slugToName(slug: string): string {
  return slug
    .replace(/^\d+-/, "")
    .split("-")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}

export async function loadGraph(): Promise<GraphData> {
  // 1. Fetch workspace graph (nodes + edges)
  const graphRes = (await fetch("/api/graph").then((r) => r.json())) as {
    nodes: EngineGraphNode[];
    edges: EngineEdge[];
    stats: { entity_count: number; edge_count: number };
  };

  // 2. Fetch context→node mapping for edge resolution
  const contextToNode = new Map<string, string>();
  try {
    const mapRes = await fetch("/api/contexts/map");
    if (mapRes.ok) {
      const mapData = (await mapRes.json()) as {
        contexts: { id: string; node: string }[];
      };
      for (const c of mapData.contexts) {
        contextToNode.set(c.id, c.node);
      }
    }
  } catch {
    /* degrade gracefully */
  }

  // 3. Fetch per-node contexts (signals) in parallel
  const nodeDetails = await Promise.all(
    graphRes.nodes.map(async (n) => {
      try {
        const res = await fetch(`/api/node/${encodeURIComponent(n.node)}`);
        if (!res.ok) return null;
        return (await res.json()) as EngineNodeDetail;
      } catch {
        return null;
      }
    }),
  );

  const nodes: GNode[] = [];
  const edges: GEdge[] = [];

  // 4. Build workspace nodes
  for (const n of graphRes.nodes) {
    const kindKey = n.node.replace(/^\d+-/, "");
    nodes.push({
      id: `node:${n.node}`,
      label: slugToName(n.node),
      kind: "node",
      sub: KIND_MAP[kindKey] ?? "default",
      connections: n.context_count,
    });
  }

  // 5. Build signal nodes from per-node contexts
  const seenEntities = new Set<string>();
  for (const detail of nodeDetails) {
    if (!detail) continue;
    for (const ctx of detail.contexts) {
      const sigId = `signal:${ctx.id}`;
      nodes.push({
        id: sigId,
        label: ctx.title || ctx.id,
        kind: "signal",
        sub: ctx.genre ?? "note",
        connections: 0,
        node_slug: detail.node,
      });
      edges.push({
        source: `node:${detail.node}`,
        target: sigId,
        relation: "contains",
      });
    }

    // Build entity nodes from edge data (entities appear in node edges)
    for (const e of detail.edges || []) {
      if (e.relation === "mentions" || e.relation === "shared_entity") {
        const eid = `entity:${e.target}`;
        if (!seenEntities.has(eid)) {
          seenEntities.add(eid);
          nodes.push({
            id: eid,
            label: e.target,
            kind: "entity",
            sub: "person",
            connections: 1,
          });
        }
      }
    }
  }

  // 6. Build node-to-node edges from context cross-references
  const nodeEdgeSet = new Set<string>();
  for (const e of graphRes.edges) {
    const srcNode = contextToNode.get(e.source);
    const tgtNode = contextToNode.get(e.target);
    if (srcNode && tgtNode && srcNode !== tgtNode) {
      const key = [srcNode, tgtNode].sort().join("|");
      if (!nodeEdgeSet.has(key)) {
        nodeEdgeSet.add(key);
        edges.push({
          source: `node:${srcNode}`,
          target: `node:${tgtNode}`,
          relation: "mentions",
        });
      }
    }
  }

  return { nodes, edges };
}
