<script lang="ts">
	import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

<svelte:head>
	<title>Workspaces — Optimal Engine</title>
</svelte:head>

<h1>Workspaces</h1>
<p>
	Workspaces are isolated brains. Each workspace has its own node tree, signal files, derivative
	index, and wiki. Data in workspace A is never visible to queries in workspace B unless explicitly
	cross-linked.
</p>

<h2>Why Multi-Workspace</h2>

<p>
	A single namespace causes sales context to bleed into engineering answers, and M&A-sensitive
	material to leak into general queries. One workspace per isolated brain solves this at the storage
	and query layer.
</p>

<table>
	<thead>
		<tr><th>Workspace</th><th>Who uses it</th><th>What it holds</th></tr>
	</thead>
	<tbody>
		<tr><td><code>default</code></td><td>Everyone</td><td>Cross-cutting company knowledge</td></tr>
		<tr><td><code>engineering</code></td><td>Engineers</td><td>Architecture decisions, code context, post-mortems</td></tr>
		<tr><td><code>sales</code></td><td>Sales team</td><td>Deals, negotiations, customer signals</td></tr>
		<tr><td><code>legal</code></td><td>Legal + leadership</td><td>Contracts, regulatory analysis</td></tr>
		<tr><td><code>ma</code></td><td>Leadership only</td><td>M&A targets, due-diligence</td></tr>
	</tbody>
</table>

<p>
	<strong>Audience tags</strong> (<code>audience=sales</code>, <code>audience=exec</code>) let one
	workspace serve multiple receiver types. Workspaces are the isolation boundary; audience is the
	presentation layer.
</p>

<h2>Creating a Workspace</h2>

<CodeBlock code={`curl -X POST http://localhost:4200/api/workspaces \\
  -H 'Content-Type: application/json' \\
  -d '{"slug":"engineering","name":"Engineering Brain","tenant":"default"}'`} lang="bash" />

<p>Response (201):</p>
<CodeBlock code={`{
  "id": "ws:engineering",
  "tenant_id": "default",
  "slug": "engineering",
  "name": "Engineering Brain",
  "status": "active",
  "created_at": "2026-04-28T00:00:00Z"
}`} lang="json" />

<p>Or via Mix task:</p>
<CodeBlock code={`mix optimal.init ~/company-brain/engineering`} lang="bash" />

<h2>On-Disk Convention</h2>

<CodeBlock code={`workspace/
├── nodes/
│   └── <slug>/               e.g. 01-founder, 02-platform, 03-sales
│       ├── context.md        Persistent facts — edit in place as ground truth changes
│       ├── signal.md         Rolling weekly status — overwritten each cycle
│       └── signals/          Append-only dated signals
│           └── YYYY-MM-DD-slug.md
├── .wiki/
│   ├── SCHEMA.md             Governance rules the curator honors
│   └── <slug>.md             Curated pages with citations back to Tier 1
├── .optimal/
│   └── config.yaml           Engine config for this workspace
└── assets/                   Binary attachments`} lang="text" />

<h2>Signal File Frontmatter</h2>

<p>Every signal file carries YAML frontmatter followed by a markdown body:</p>

<CodeBlock code={`---
title: Customer pricing call — Q4
genre: transcript
mode: linguistic
node: 03-sales
sn_ratio: 0.75
entities:
  - { name: "Alice", type: person }
  - { name: "Healthtech Product", type: product }
authored_at: 2026-04-28T14:00:00Z
---

## Summary
One-sentence abstract. Engine pulls this for L0 (~100 tokens).

## Key points
- Each bullet is an atomic claim. Engine pulls this for L1 (~2K tokens).

## Detail
Full content. Prose is fine; the decomposer splits automatically.`} lang="yaml" />

<table>
	<thead>
		<tr><th>Field</th><th>Required</th><th>Used for</th></tr>
	</thead>
	<tbody>
		<tr><td><code>title</code></td><td>yes</td><td>Display + search ranking</td></tr>
		<tr><td><code>genre</code></td><td>yes</td><td>Classification + retrieval filter</td></tr>
		<tr><td><code>mode</code></td><td>no</td><td>linguistic / visual / code / data / mixed</td></tr>
		<tr><td><code>node</code></td><td>inferred</td><td>Routed from directory; override if needed</td></tr>
		<tr><td><code>sn_ratio</code></td><td>no</td><td>Signal/noise boost; defaults to 0.5</td></tr>
		<tr><td><code>entities</code></td><td>no</td><td>Pre-extracted (engine can derive them)</td></tr>
		<tr><td><code>authored_at</code></td><td>no</td><td>ISO-8601; defaults to file mtime</td></tr>
	</tbody>
</table>

<h2>Workspace Config</h2>

<CodeBlock code={`# .optimal/config.yaml
workspace:
  slug: engineering
  name: Engineering Brain
  tenant: default

ingestion:
  formats: [md, pdf, code, yaml, json]
  exclude_globs: ["*.lock", "node_modules/**", ".git/**"]
  min_sn_ratio: 0.3

retrieval:
  default_audience: engineering
  default_bandwidth: l1
  graph_boost: true

wiki:
  curation_interval: 3600
  audiences: [default, engineering, leadership]
  citation_ttl_days: 90

surfacing:
  categories:
    - recent_actions
    - blockers
    - contradictions
    - specifications
    - metrics`} lang="yaml" filename=".optimal/config.yaml" />

<CodeBlock code={`# Read config
curl http://localhost:4200/api/workspaces/engineering/config | jq '.config'

# Write (deep-merge)
curl -X PATCH http://localhost:4200/api/workspaces/engineering/config \\
  -H 'Content-Type: application/json' \\
  -d '{"ingestion": {"min_sn_ratio": 0.4}}'`} lang="bash" />

<h2>Isolation Guarantees</h2>

<ol>
	<li><strong>Filesystem isolation</strong> — each workspace resolves to its own directory tree.</li>
	<li><strong>Query isolation</strong> — all SQL queries include <code>workspace_id = ?</code> predicates. Cross-workspace joins don't exist.</li>
	<li><strong>Wiki isolation</strong> — <code>.wiki/</code> is per-workspace. A page in <code>engineering</code> is never served in response to a <code>sales</code> query.</li>
	<li><strong>Memory isolation</strong> — <code>POST /api/memory</code> scopes to workspace. <code>GET /api/memory</code> returns only memories matching the requested workspace.</li>
	<li><strong>Subscription isolation</strong> — surfacing events are scoped to <code>workspace_id</code>. A subscription in <code>sales</code> never receives events from <code>engineering</code>.</li>
</ol>

<h2>Ingesting a Workspace</h2>

<CodeBlock code={`# Full workspace ingest (walk the node tree, process all signals)
mix optimal.ingest_workspace ~/company-brain/engineering

# Ingest a single signal
mix optimal.ingest --file ~/company-brain/engineering/nodes/02-platform/signals/2026-04-28-api-design.md`} lang="bash" />

<h2>See Also</h2>
<ul>
	<li><a href="/api/workspaces">Workspaces API</a> — CRUD and config endpoints</li>
	<li><a href="/concepts/memory-primitive">Memory Primitive</a> — audience scoping on individual memories</li>
	<li><a href="/api/surfacing">Surfacing API</a> — subscription isolation guarantees</li>
</ul>
