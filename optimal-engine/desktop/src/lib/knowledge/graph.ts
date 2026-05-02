/**
 * Builds the unified graph data shape for the /graph route:
 *   workspace nodes  (kind=node,   color=type-based)
 *   signals          (kind=signal, color=genre-based)
 *   entities         (kind=entity, color=entity-type-based)
 *
 * Edges:
 *   node → signal    (signal lives in this node)
 *   signal → entity  (signal mentions this entity)
 *   node → node      (parent → child in the workspace tree)
 *
 * The component uses PixiJS for rendering + d3-force for layout, matching
 * the visual language of the Pages module.
 */

import type { WorkspaceNode } from "$lib/api/workspace";

export type GKind = "node" | "signal" | "entity";

export interface GNode {
  id: string;
  label: string;
  kind: GKind;
  // Drives colour + sizing. For `node` this is the node.kind (org/person/…);
  // for `signal` it's the genre; for `entity` it's the entity type.
  sub: string;
  connections: number;
  // Signal-specific
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

export interface OptimalGraphResponse {
  entities: { name: string; type: string; connections: number }[];
  edges: {
    source: string;
    target: string;
    relation: string;
    weight: number;
  }[];
  stats: {
    entity_count: number;
    edge_count: number;
    edge_types: Record<string, number>;
  };
}

export interface NodeFilesResponse {
  files: {
    name: string;
    path: string;
    is_dir: boolean;
    size: number;
    genre: string | null;
    modified_at: string | null;
  }[];
}

/** Pull everything the graph needs in parallel and fold it into one shape. */
export async function loadGraph(): Promise<GraphData> {
  // The engine serves at /api/graph (not /api/optimal/graph).
  // /api/graph returns {nodes: [{node, context_count, last_modified}], edges: [...], stats: {...}}
  const graphRes = (await fetch("/api/graph").then((r) => r.json())) as {
    nodes: { node: string; context_count: number; last_modified: string }[];
    edges: {
      source: string;
      target: string;
      relation: string;
      weight: number;
    }[];
    stats: { entity_count: number; edge_count: number };
  };

  // Convert engine graph nodes to WorkspaceNode shape
  const kindGuess: Record<string, string> = {
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
  };

  const workspaceNodes: WorkspaceNode[] = graphRes.nodes.map((n) => {
    const slug = n.node;
    const name = slug
      .replace(/^\d+-/, "")
      .split("-")
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
      .join(" ");
    const kindKey = slug.replace(/^\d+-/, "");
    return {
      id: slug,
      slug,
      name,
      kind: kindGuess[kindKey] ?? "default",
      signal_count: n.context_count,
      parent_id: null,
    } as WorkspaceNode;
  });

  const entityData = (graphRes as any).entities ?? [];
  const entityEdges = graphRes.edges;

  // Fetch per-node signal lists in parallel (engine serves at /api/node/:slug).
  const signalLists = await Promise.all(
    workspaceNodes.map(async (n) => {
      try {
        const res = await fetch(
          `/api/node/${encodeURIComponent(n.slug)}/files`,
        );
        if (!res.ok) return { node: n, files: [] };
        const data = (await res.json()) as NodeFilesResponse;
        return { node: n, files: data.files ?? [] };
      } catch {
        return { node: n, files: [] };
      }
    }),
  );

  const nodes: GNode[] = [];
  const edges: GEdge[] = [];

  // Workspace nodes
  for (const n of workspaceNodes) {
    nodes.push({
      id: `node:${n.slug}`,
      label: n.name,
      kind: "node",
      sub: n.kind,
      connections: n.signal_count,
    });
    if (n.parent_id) {
      // parent_id is the nodes.id, but our node ids here use slugs — walk
      // the list to find the parent slug.
      const parent = workspaceNodes.find((p) => p.id === n.parent_id);
      if (parent) {
        edges.push({
          source: `node:${parent.slug}`,
          target: `node:${n.slug}`,
          relation: "parent",
        });
      }
    }
  }

  // Signals under each node
  for (const { node, files } of signalLists) {
    for (const f of files) {
      const sigId = `signal:${f.path}`;
      nodes.push({
        id: sigId,
        label: f.name,
        kind: "signal",
        sub: f.genre ?? "note",
        connections: 0,
        node_slug: node.slug,
      });
      edges.push({
        source: `node:${node.slug}`,
        target: sigId,
        relation: "contains",
      });
    }
  }

  // Entities (collapse into one node per entity name across the tenant)
  const entityIds = new Set<string>();
  for (const e of entityData) {
    const eid = `entity:${e.name}`;
    if (!entityIds.has(eid)) {
      entityIds.add(eid);
      nodes.push({
        id: eid,
        label: e.name,
        kind: "entity",
        sub: e.type,
        connections: e.connections,
      });
    }
  }

  // Build a context_id → node_slug lookup from the engine edges.
  // Engine edges use format: {source: context_hash, target: node_slug, relation: "lives_in"|"cross_ref"|"works_on"}
  const contextToNode = new Map<string, string>();
  for (const e of entityEdges) {
    if (
      e.relation === "lives_in" &&
      workspaceNodes.some((n) => n.slug === e.target)
    ) {
      contextToNode.set(e.source, e.target);
    }
  }

  // Cross-ref edges: context references another context → node-to-node edge
  const nodeEdgeSet = new Set<string>();
  for (const e of entityEdges) {
    if (e.relation === "cross_ref") {
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
    // works_on edges: node → node relationship from topology
    if (e.relation === "works_on") {
      const src = workspaceNodes.find((n) => n.slug === e.source);
      const tgt = workspaceNodes.find((n) => n.slug === e.target);
      if (src && tgt) {
        const key = [src.slug, tgt.slug].sort().join("|");
        if (!nodeEdgeSet.has(key)) {
          nodeEdgeSet.add(key);
          edges.push({
            source: `node:${src.slug}`,
            target: `node:${tgt.slug}`,
            relation: "mentions",
          });
        }
      }
    }
  }

  // Entity → node edges: connect entities to the nodes they appear in
  for (const e of entityData) {
    const eid = `entity:${e.name}`;
    // Find which nodes mention this entity
    for (const { node, files } of signalLists) {
      if (files.length > 0) {
        edges.push({
          source: `node:${node.slug}`,
          target: eid,
          relation: "mentions",
        });
        break;
      }
    }
  }

  return { nodes, edges };
}
