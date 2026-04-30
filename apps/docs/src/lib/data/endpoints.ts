export interface Param {
  name: string;
  type: string;
  required: boolean;
  default?: string;
  description: string;
}

export interface Endpoint {
  method: "GET" | "POST" | "PATCH" | "DELETE";
  path: string;
  summary: string;
  description: string;
  params?: Param[];
  body?: Param[];
  returns: string;
  useWhen: string;
  example: string;
}

export const retrievalEndpoints: Endpoint[] = [
  {
    method: "POST",
    path: "/api/rag",
    summary: "Wiki-first open question",
    description:
      "The primary endpoint for agent consumption. Tries the wiki (Tier 3) first; falls back to hybrid retrieval (BM25 + vector + graph) only on miss. Returns answer with inline citations and a wiki_hit flag.",
    body: [
      {
        name: "query",
        type: "string",
        required: true,
        description: "Natural-language question",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        default: '"default"',
        description: "Workspace slug or ID",
      },
      {
        name: "format",
        type: "string",
        required: false,
        default: '"markdown"',
        description: "markdown | text | claude | openai | json",
      },
      {
        name: "bandwidth",
        type: "string",
        required: false,
        default: '"medium"',
        description: "l0 (~100 tok) | medium (~2K tok) | full",
      },
      {
        name: "audience",
        type: "string",
        required: false,
        default: '"default"',
        description: "Audience tag for wiki variant selection",
      },
    ],
    returns: "{ answer, sources, wiki_hit, citations }",
    useWhen: "Agent asks an open question about organizational knowledge",
    example: `curl -X POST http://localhost:4200/api/rag \\
  -H 'Content-Type: application/json' \\
  -d '{"query":"current pricing strategy","workspace":"sales","format":"claude","audience":"sales"}'`,
  },
  {
    method: "GET",
    path: "/api/search",
    summary: "Hybrid document search",
    description:
      "BM25 + vector + graph search returning signal-level metadata. Applies graph_boost (connected entities score higher) and temporal_decay (recent signals score higher).",
    params: [
      {
        name: "q",
        type: "string",
        required: true,
        description: "Search terms",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        default: '"default"',
        description: "Workspace scope",
      },
      {
        name: "limit",
        type: "integer",
        required: false,
        default: "10",
        description: "Max results",
      },
    ],
    returns: "{ query, results: [{ id, title, slug, score, snippet, node }] }",
    useWhen:
      "User wants to find documents or signals by keyword or semantic similarity",
    example: `curl 'http://localhost:4200/api/search?q=pricing+negotiation&workspace=sales&limit=10'`,
  },
  {
    method: "GET",
    path: "/api/grep",
    summary: "Chunk-level filtered search",
    description:
      "Hybrid semantic + literal grep at chunk level. Returns chunk-level matches with full signal trace: slug, scale, intent, sn_ratio, modality, snippet, score. Use for debugging pipeline output, building filtered context, or auditing intent distribution.",
    params: [
      {
        name: "q",
        type: "string",
        required: true,
        description: "Search terms",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        default: '"default"',
        description: "Workspace scope",
      },
      {
        name: "intent",
        type: "string",
        required: false,
        description:
          "Filter by intent atom: request_info | propose_decision | record_fact | express_concern | commit_action | reference | narrate | reflect | specify | measure",
      },
      {
        name: "scale",
        type: "string",
        required: false,
        description: "document | section | paragraph | chunk",
      },
      {
        name: "modality",
        type: "string",
        required: false,
        description: "text | image | audio | code",
      },
      {
        name: "limit",
        type: "integer",
        required: false,
        default: "25",
        description: "Max results",
      },
      {
        name: "literal",
        type: "boolean",
        required: false,
        default: "false",
        description: "Force FTS literal match, skip vector",
      },
      {
        name: "path",
        type: "string",
        required: false,
        description: "Restrict to node slug prefix",
      },
    ],
    returns:
      "{ query, workspace_id, results: [{ slug, scale, intent, sn_ratio, modality, snippet, score }] }",
    useWhen:
      "Debugging pipeline output, building filtered context, auditing intent distribution",
    example: `curl 'http://localhost:4200/api/grep?q=pricing&workspace=sales&intent=record_fact&scale=paragraph'`,
  },
  {
    method: "GET",
    path: "/api/profile",
    summary: "4-tier workspace snapshot",
    description:
      "Returns static facts + dynamic signals + curated wiki + recent activity in one call. The recommended way to seed an agent system prompt.",
    params: [
      {
        name: "workspace",
        type: "string",
        required: false,
        default: '"default"',
        description: "Workspace slug or ID",
      },
      {
        name: "audience",
        type: "string",
        required: false,
        default: '"default"',
        description: "Audience tag",
      },
      {
        name: "bandwidth",
        type: "string",
        required: false,
        default: '"l1"',
        description: "l0 | l1 | full",
      },
      {
        name: "node",
        type: "string",
        required: false,
        description: "Restrict to one node slug",
      },
      {
        name: "tenant",
        type: "string",
        required: false,
        default: '"default"',
        description: "Tenant ID",
      },
    ],
    returns:
      "{ workspace_id, tenant_id, audience, static, dynamic, curated, activity, entities, generated_at }",
    useWhen:
      "Seeding an agent system prompt with workspace context before a session starts",
    example: `curl 'http://localhost:4200/api/profile?workspace=sales&audience=exec&bandwidth=l1'`,
  },
];

export const memoryEndpoints: Endpoint[] = [
  {
    method: "POST",
    path: "/api/memory",
    summary: "Create a memory",
    description:
      "Create a new first-class memory. Not a key-value store — every memory carries version lineage, typed relations, audience scoping, and an optional citation URI.",
    body: [
      {
        name: "content",
        type: "string",
        required: true,
        description: "The memory text",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        default: '"default"',
        description: "Workspace scope",
      },
      {
        name: "is_static",
        type: "boolean",
        required: false,
        description: "Static (rarely changes) vs dynamic (rolling)",
      },
      {
        name: "audience",
        type: "string",
        required: false,
        description: "Audience tag — restricts visibility in profile/rag",
      },
      {
        name: "citation_uri",
        type: "string",
        required: false,
        description: "optimal:// URI pointing to source",
      },
      {
        name: "source_chunk_id",
        type: "string",
        required: false,
        description: "Chunk ID this memory derives from",
      },
      {
        name: "metadata",
        type: "object",
        required: false,
        description: "Arbitrary key-value metadata",
      },
    ],
    returns: "201 + full memory struct",
    useWhen:
      "An agent observes a fact during a session and needs it to persist across sessions",
    example: `curl -X POST http://localhost:4200/api/memory \\
  -H 'Content-Type: application/json' \\
  -d '{"content":"Alice owns the pricing negotiation","workspace":"sales","is_static":true}'`,
  },
  {
    method: "GET",
    path: "/api/memory",
    summary: "List memories",
    description:
      "List memories for a workspace. By default returns only the latest version of each memory.",
    params: [
      {
        name: "workspace",
        type: "string",
        required: false,
        description: "Workspace scope",
      },
      {
        name: "audience",
        type: "string",
        required: false,
        description: "Filter by audience tag",
      },
      {
        name: "include_forgotten",
        type: "boolean",
        required: false,
        description: "Include soft-deleted memories",
      },
      {
        name: "include_old_versions",
        type: "boolean",
        required: false,
        description: "Include superseded versions",
      },
      {
        name: "limit",
        type: "integer",
        required: false,
        default: "50",
        description: "Max results",
      },
    ],
    returns: "{ workspace_id, count, memories: [...] }",
    useWhen: "Listing all active memories in a workspace",
    example: `curl 'http://localhost:4200/api/memory?workspace=sales&limit=20'`,
  },
  {
    method: "POST",
    path: "/api/memory/:id/update",
    summary: "Version a memory",
    description:
      "Create a new version of a memory (relation type: updates). The original is marked is_latest: false. Version chain navigable via GET /api/memory/:id/versions.",
    body: [
      {
        name: "content",
        type: "string",
        required: true,
        description: "Updated memory text",
      },
      {
        name: "audience",
        type: "string",
        required: false,
        description: "New audience tag",
      },
      {
        name: "citation_uri",
        type: "string",
        required: false,
        description: "New citation URI",
      },
      {
        name: "metadata",
        type: "object",
        required: false,
        description: "Updated metadata",
      },
    ],
    returns: "201 + new memory struct",
    useWhen: "A fact has changed and the old version is now stale",
    example: `curl -X POST http://localhost:4200/api/memory/mem:abc123/update \\
  -H 'Content-Type: application/json' \\
  -d '{"content":"Alice and Bob co-own pricing as of 2026-04-28"}'`,
  },
  {
    method: "POST",
    path: "/api/memory/:id/extend",
    summary: "Extend a memory",
    description:
      "Create a child memory linked via the extends relation. The parent is not superseded — both exist. Use for addenda, clarifications, or specializations.",
    body: [
      {
        name: "content",
        type: "string",
        required: true,
        description: "Extension text",
      },
      {
        name: "audience",
        type: "string",
        required: false,
        description: "Audience tag",
      },
      {
        name: "citation_uri",
        type: "string",
        required: false,
        description: "Citation URI",
      },
      {
        name: "metadata",
        type: "object",
        required: false,
        description: "Metadata",
      },
    ],
    returns: "201 + new memory struct",
    useWhen: "Adding detail to a memory without superseding it",
    example: `curl -X POST http://localhost:4200/api/memory/mem:abc123/extend \\
  -H 'Content-Type: application/json' \\
  -d '{"content":"Bob is the fallback contact when Alice is OOO"}'`,
  },
  {
    method: "POST",
    path: "/api/memory/:id/forget",
    summary: "Soft-delete a memory",
    description:
      "Marks the memory is_forgotten: true. Data is preserved for audit and GDPR erasure workflows. Pass forget_after for scheduled forgetting.",
    body: [
      {
        name: "reason",
        type: "string",
        required: false,
        description: "Human-readable reason",
      },
      {
        name: "forget_after",
        type: "string",
        required: false,
        description: "ISO-8601 datetime — schedule future forgetting",
      },
    ],
    returns: "204",
    useWhen: "A fact is no longer true or is subject to retention/GDPR policy",
    example: `curl -X POST http://localhost:4200/api/memory/mem:abc123/forget \\
  -H 'Content-Type: application/json' \\
  -d '{"reason":"decision reversed"}'`,
  },
];

export const recallEndpoints: Endpoint[] = [
  {
    method: "GET",
    path: "/api/recall/actions",
    summary: "What happened / who decided what",
    description:
      "Retrieves past actions, decisions, and commitments. Builds an intent-optimized query targeting commit_action and propose_decision chunks, then routes through /api/rag.",
    params: [
      {
        name: "actor",
        type: "string",
        required: false,
        description: "Person who acted",
      },
      {
        name: "topic",
        type: "string",
        required: false,
        description: "Subject matter",
      },
      {
        name: "since",
        type: "string",
        required: false,
        description: "ISO-8601 date lower bound",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        description: "Workspace scope",
      },
    ],
    returns: "Same envelope as /api/rag + recall_query field",
    useWhen: "Recovering what decisions were made or what someone committed to",
    example: `curl 'http://localhost:4200/api/recall/actions?actor=Alice&topic=pricing&since=2026-01-01&workspace=sales'`,
  },
  {
    method: "GET",
    path: "/api/recall/who",
    summary: "Ownership / contact lookup",
    description:
      "Contact and ownership lookup. Targets record_fact chunks with entity matching.",
    params: [
      {
        name: "topic",
        type: "string",
        required: true,
        description: "Subject or object to look up",
      },
      {
        name: "role",
        type: "string",
        required: false,
        default: '"owner"',
        description: "owner | lead | contact | …",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        description: "Workspace scope",
      },
    ],
    returns: "Same envelope as /api/rag + recall_query field",
    useWhen: "Recovering who owns or is responsible for something",
    example: `curl 'http://localhost:4200/api/recall/who?topic=pricing&role=owner&workspace=sales'`,
  },
  {
    method: "GET",
    path: "/api/recall/when",
    summary: "Schedule / temporal lookup",
    description:
      "Temporal lookup. Targets measure and record_fact chunks with date/time entities.",
    params: [
      {
        name: "event",
        type: "string",
        required: true,
        description: "Event name or description",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        description: "Workspace scope",
      },
    ],
    returns: "Same envelope as /api/rag + recall_query field",
    useWhen:
      "Recovering when something is scheduled or when something happened",
    example: `curl 'http://localhost:4200/api/recall/when?event=Q4+board+review&workspace=default'`,
  },
  {
    method: "GET",
    path: "/api/recall/where",
    summary: "Object-location lookup",
    description:
      "Locates which node, file, or path holds something. Targets reference chunks.",
    params: [
      {
        name: "thing",
        type: "string",
        required: true,
        description: "Item to locate",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        description: "Workspace scope",
      },
    ],
    returns: "Same envelope as /api/rag + recall_query field",
    useWhen: "Recovering where a document, file, or artifact lives",
    example: `curl 'http://localhost:4200/api/recall/where?thing=pricing+deck&workspace=sales'`,
  },
  {
    method: "GET",
    path: "/api/recall/owns",
    summary: "Actor's current commitments",
    description:
      "Open-task and current-commitment lookup for an actor. Targets commit_action chunks that are not yet resolved.",
    params: [
      {
        name: "actor",
        type: "string",
        required: true,
        description: "Person whose commitments to surface",
      },
      {
        name: "workspace",
        type: "string",
        required: false,
        description: "Workspace scope",
      },
    ],
    returns: "Same envelope as /api/rag + recall_query field",
    useWhen: "Recovering what someone is currently responsible for",
    example: `curl 'http://localhost:4200/api/recall/owns?actor=Alice&workspace=sales'`,
  },
];
