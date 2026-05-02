# Optimal Engine

**The second brain of your company.** Open-source memory infrastructure that ingests every signal flowing through an organization — chat, email, docs, meetings, tickets, CRM, code, voice, video — decodes the intent behind each, embeds it into a multi-modal aligned vector space, clusters by theme, curates an LLM-maintained wiki, and delivers permission-scoped context to any agent in **under 200 ms**.

> Classical RAG re-discovers the same facts on every query. The Optimal Engine discovers once, curates forever, and delivers permission-gated, receiver-matched context — with hot citations and a fail-closed integrity gate.

```
PRINCIPALS         ORGANIZATION ▸ WORKSPACE ▸ NODE ▸ SIGNAL ▸ CHUNK
                    (tenant)      (brain)    (folder) (.md)  (4 scales)

INGEST   →   PARSE   →   DECOMPOSE   →   CLASSIFY   →   EMBED
                                                          │
CURATE   ←   CLUSTER  ←  STORE       ←   ROUTE       ←──┘

           Tier 1 raw .md           Tier 2 SQLite           Tier 3 .wiki/
           (immutable)              (rebuildable)           (LLM-curated)
                                                                ▲
AGENT  ──── /api/rag ────►  envelope  ◄── /api/profile ────────┤
                                                                │
                            citations  ────────────────────────┘
```

---

## What ships

**v0.2.0 — Phases 1.5 → 18 complete. 1,297 tests passing.** MIT licensed. Self-hosted. Local-first.

### Engine (Elixir / OTP)

| Layer | What's there |
|---|---|
| **9-stage pipeline** | Intake → Parse → Decompose → Classify → Embed → Route → Store → Cluster → Curate. Each stage has a typed contract with the next; swap any stage without breaking the rest. |
| **3 tiers** | Tier 1 raw markdown signals (immutable, append-only) → Tier 2 SQLite + FTS5 + sqlite-vec derivatives (rebuildable from Tier 1) → Tier 3 LLM-curated wiki with hot citations + 7 directives + audience variants. |
| **Multi-workspace** | Org → Workspace → Node hierarchy. Every workspace gets its own filesystem, its own wiki, its own config. Filesystem provisioned on create. Workspace_id flows through every retrieval path. |
| **Memory primitive** | First-class versioned memories with parent/root chains, 5 typed relations (`updates` / `extends` / `derives` / `contradicts` / `cites`), soft-forget with reason + scheduled expiry, content-hash dedup with three policies, `is_static` vs dynamic, audience scoping, citation_uri on every memory. |
| **Signal classification** | `S=(M,G,T,F,W)` per chunk: Mode × Genre × Type × Format × Structure. Plus 10-value intent enum: `request_info / propose_decision / record_fact / express_concern / commit_action / reference / narrate / reflect / specify / measure`. |
| **4 chunk scales** | document → section → paragraph → chunk. Retrieval returns the coarsest scale that answers. |
| **Multi-modal embeddings** | Text + image + audio in the same 768-dim aligned space (nomic-embed-text + nomic-embed-vision + whisper.cpp). A text query can retrieve an image. |
| **DataArchitecture layer** | Model-agnostic processor registry. 12 modalities, 7 built-in architectures, 6 processors. Bring your own model. |
| **Proactive surfacing** | Background `Surfacer` GenServer + 14-category subscription model + SSE push channel. The engine pushes relevant memories without an explicit query. |
| **Wiki curator** | LLM-maintained pages with hot citations, 7 executable directives (`{{cite}} {{include}} {{expand}} {{search}} {{table}} {{trace}} {{recent}}`), audience-aware variants, contradiction detection with 3 policies. |
| **Authentication** | Bearer-token API keys. Bcrypt-hashed secrets. Scoped permissions (`read:*` / `write:*` / `admin`). Workspace-scoped or tenant-wide. Opt-in via `OPTIMAL_AUTH_REQUIRED=true`. |
| **Rate limiting** | ETS-backed token bucket per API key (or IP for anonymous). 100 req/min default, configurable per workspace. `X-RateLimit-*` headers + `Retry-After` on 429. |
| **Compliance** | GDPR Article 17 erasure + legal hold + per-node retention TTL + audit log of every retrieval. |

### Distribution surface

```
.
├── apps/
│   ├── docs/            SvelteKit static docs site (port 1422)
│   └── mcp/             First-party MCP server (stdio, 9 tools)
│
├── desktop/             Tauri + SvelteKit shell (port 1420, 9 routes)
│
├── site/                Marketing site (port 1421)
│
├── sdks/
│   ├── typescript/      @optimal-engine/client + Vercel AI SDK + OpenAI Agents adapters
│   └── python/          optimal-engine — sync + async + LangChain + OpenAI Agents adapters
│
├── extensions/
│   ├── browser/         Chrome MV3 web clipper (popup + options + context menu)
│   └── raycast/         macOS Raycast extension (3 commands)
│
├── skills/
│   └── optimal-engine/  Claude Code Skill (SKILL.md + 6 references + bootstrap.sh)
│
├── deploy/              Docker compose stack (engine + 3 sites + Caddy auto-TLS)
│
├── lib/                 Engine source
├── test/                1,297 tests
└── sample-workspace/    Demo data (6 nodes, 13 signals, 2 wiki pages)
```

---

## Quick start

### 5-minute local

Requires Elixir `~> 1.17`, Erlang/OTP 26+, Node 20+, a C toolchain.

```bash
git clone git@github.com:robertohluna/OptimalEngine.git
cd OptimalEngine
make install              # mix deps.get + mix compile
make bootstrap            # ingest sample-workspace/
make dev                  # iex -S mix (engine + HTTP API on :4200)

# In separate terminals
cd desktop && npm install && npm run dev    # http://localhost:1420
cd site    && npm install && npm run dev    # http://localhost:1421
cd apps/docs && npm install && npm run dev  # http://localhost:1422
```

### One-command Docker

```bash
cd deploy
docker compose up
# engine on :4200, desktop on :1420, site on :1421, docs on :1422
```

For production, copy `deploy/env.example → .env.prod`, set `OPTIMAL_AUTH_REQUIRED=true`, and run with the prod overlay (auto-TLS via Let's Encrypt):

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### First queries

```bash
mix optimal.search "platform"
mix optimal.rag    "healthtech pricing decision" --trace
mix optimal.grep   "decision" --intent decide --scale paragraph
mix optimal.wiki   list

# HTTP equivalents
curl 'http://127.0.0.1:4200/api/profile?workspace=default&bandwidth=l1'
curl 'http://127.0.0.1:4200/api/recall/actions?topic=pricing&workspace=default'
```

---

## Status

| Phase | Scope | Status |
|---|---|---|
| 0 — 14 | Engine foundation through DataArchitecture layer | ✅ |
| 1.5 | **Workspaces** — multi-workspace isolation, per-workspace filesystem + config, workspace_id scoping through every retrieval path, two-step module rename | ✅ |
| 1.6 / 1.7 | Workspace-aware wiki + RAG + recall + search + ingest | ✅ |
| 15 | **Proactive surfacing** — Surfacer GenServer, 14-category subscription model, SSE push channel | ✅ |
| 16 | **Workspace config** — per-workspace `.optimal/config.yaml`. **Semantic grep**, **4-tier profile**, **wiki contradiction detection**, **timeline + heatmap visualizations** | ✅ |
| 17 | **Memory primitive** — versioned, 5 typed relations, soft-forget, citations. **TypeScript SDK**. **Python SDK**. **MCP server**. **Browser extension**. **Memory graph viz**. **Claude Code Skill** | ✅ |
| 18 | **Memory ↔ Wiki bridge**. **Content-hash dedup**. **API key auth**. **Rate limiting**. **Streaming `/api/rag/stream`**. **Docs site**. **Raycast extension**. **Docker deploy stack** | ✅ |

**Test suite:** 1,297 passing, 29 excluded (RocksDB NIF, optional backend).

---

## Architecture at a glance

### The three tiers

```
TIER 3 — THE WIKI                LLM-curated. Audience-aware. Hot-cited.
Path: <workspace>/.wiki/         7 executable directives. Read first.
                                 Contradiction-gated.
        ▲ CURATE                  ▼ DERIVE
TIER 2 — DERIVATIVES             SQLite + FTS5 + sqlite-vec + graph + clusters.
Path: .optimal/index.db          Workspace-scoped. Rebuildable from Tier 1.

        ▲ DERIVE                  ▼ INGEST
TIER 1 — RAW SIGNALS             Markdown files with YAML frontmatter.
Path: <workspace>/nodes/         Hash-addressed. Append-only. Git-friendly.
      <workspace>/assets/        The engine NEVER rewrites them.
```

### The 8-layer onion

When the agent reads, it peels from the outside in. Most queries terminate at the wiki layer.

```
Outside:    [Agent]
            ↓ ask
Layer 7:    Envelope            ACL-scoped, audience-shaped, bandwidth-matched
Layer 6:    Wiki                Curated, hot-cited, audience-aware
Layer 5:    Cluster             HDBSCAN themes
Layer 4:    Embed               768-dim aligned vector space
Layer 3:    Classify            S=(M,G,T,F,W) + 10-value intent
Layer 2:    Chunks              4 scales: doc / section / paragraph / chunk
Layer 1:    Parsed              Text + structure + assets
Layer 0:    Raw signal          The .md file on disk
```

### The 9-stage ingestion pipeline

```
1. INTAKE → 2. PARSE → 3. DECOMPOSE → 4. CLASSIFY → 5. EMBED → 6. ROUTE
                                                                    │
9. CURATE ← 8. CLUSTER ← 7. STORE ←────────────────────────────────┘
```

Each stage has one responsibility and a typed contract with the next. Full detail in [`docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md).

---

## Capabilities matrix

What ships today, no vendor names.

| Capability | This engine |
|---|---|
| Enterprise connectors | 14 in v1 |
| Permission-aware retrieval | chunk-level + intersection propagation |
| Signal classification `S=(M,G,T,F,W)` | every chunk classified at ingest |
| Per-chunk intent extraction | 10-value enum, deterministic |
| Tiered disclosure | L0 / L1 / full + bandwidth-matched |
| Hierarchical chunking | 4 scales: document / section / paragraph / chunk |
| Multi-modal aligned embeddings | text + image + audio in one 768-dim space |
| Cross-modal retrieval | text query → image chunk, no separate index |
| Hot citations + integrity gate | fail-closed; every claim cites |
| Wiki directive grammar | 7 directives: cite / include / expand / search / table / trace / recent |
| Audience-aware wiki variants | sales / legal / exec / engineering |
| Memory primitive | versioned, 5 typed relations, soft-forget, citations |
| Memory dedup | content-hash, three policies |
| Triggered incremental curation | triple-loop SICA learning |
| Proactive surfacing | 14 categories, SSE push, configurable per workspace |
| Contradiction detection | three policies: flag / silent-resolve / reject |
| API auth | bcrypt-hashed keys, scoped permissions |
| Rate limiting | ETS token bucket, per-key/IP |
| Streaming retrieval | SSE pipeline-stage events |
| Local-first / self-hosted | runs on a laptop; on-disk markdown |

---

## API surface

40+ HTTP endpoints. Full reference: [`apps/docs/`](apps/docs/) (run `cd apps/docs && npm run dev`) or [`skills/optimal-engine/references/api-reference.md`](skills/optimal-engine/references/api-reference.md).

### Retrieval
| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/rag` | Wiki-first retrieval envelope |
| `GET` | `/api/rag/stream` | **SSE** streaming retrieval (intent → wiki_hit → chunks → composing → envelope) |
| `GET` | `/api/search` | Hybrid BM25 + vector hits |
| `GET` | `/api/grep` | Typed semantic + literal grep with full signal trace |
| `GET` | `/api/profile` | 4-tier workspace snapshot (static / dynamic / curated / activity) |
| `GET` | `/api/recall/{actions,who,when,where,owns}` | 5 typed cued-recall verbs |

### Memory
| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/memory` | Create memory (cited, integrity-gated, content-hash dedup) |
| `GET` | `/api/memory/:id` | Fetch one |
| `GET` | `/api/memory` | List with filters (workspace, audience, include_forgotten, ...) |
| `POST` | `/api/memory/:id/forget` | Soft delete with reason |
| `POST` | `/api/memory/:id/{update,extend,derive}` | New version with typed relation |
| `GET` | `/api/memory/:id/{versions,relations}` | Version chain + typed edges |
| `POST` | `/api/memory/:id/promote` | Promote to wiki page |
| `DELETE` | `/api/memory/:id` | Hard delete |

### Wiki
| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/wiki` · `/api/wiki/:slug` | List + render Tier-3 pages |
| `GET` | `/api/wiki/contradictions` | Active contradiction surfacing events |

### Workspaces + auth
| Method | Path | Purpose |
|---|---|---|
| `GET` `POST` | `/api/workspaces` · `/:id` · `/:id/archive` | Workspace CRUD |
| `GET` `PATCH` | `/api/workspaces/:id/config` | Per-workspace YAML config |
| `POST` | `/api/auth/keys` · `/:id/revoke` | API key minting + revocation |

### Surfacing
| Method | Path | Purpose |
|---|---|---|
| `GET` `POST` `DELETE` | `/api/subscriptions` | Subscription CRUD |
| `GET` | `/api/surface/stream` | **SSE** push channel for matched events |

---

## Mix tasks

```
mix optimal.bootstrap                          ingest sample-workspace + run reality-check
mix optimal.ingest_workspace <path> --workspace <slug>
mix optimal.search <query> [--workspace <slug>]
mix optimal.rag <query> [--audience sales|legal|exec|engineering] [--trace]
mix optimal.grep <query> [--intent <atom>] [--scale <atom>] [--workspace <slug>]
mix optimal.wiki list|view|verify
mix optimal.topology [--tenant <id>]            # nodes / members / skills inspector
mix optimal.architectures                       # data-architecture catalog
mix optimal.connector list|register|run
mix optimal.compliance dsar|erase|hold|retention
mix optimal.status [--json|--quick|--metrics]
mix optimal.backup <path> [--verify]
mix optimal.migrate [--status]
mix optimal.reality_check --hard                # 50+ probes, all green target
```

---

## SDKs

### TypeScript (`@optimal-engine/client`)

```bash
npm install @optimal-engine/client
```

```typescript
import { OptimalEngine } from "@optimal-engine/client";
import { optimalEngineTools } from "@optimal-engine/client/adapters/ai-sdk";
import { generateText } from "ai";
import { openai } from "@ai-sdk/openai";

const client = new OptimalEngine({ baseUrl: "http://localhost:4200", workspace: "default" });

const profile = await client.profile({ audience: "sales", bandwidth: "l1" });
const result  = await client.ask("healthtech pricing", { audience: "sales" });
const memory  = await client.memory.create({ content: "Bob owns Q4 pricing", isStatic: true });

// Drop into Vercel AI SDK
const { text } = await generateText({
  model: openai("gpt-4o"),
  tools: optimalEngineTools(client),  // 8 tools auto-wired
  prompt: "Who owns the platform decision?",
});
```

### Python (`optimal-engine`)

```bash
pip install optimal-engine
```

```python
from optimal_engine import OptimalEngine
from optimal_engine.adapters.openai_agents import optimal_engine_tools
from agents import Agent

client  = OptimalEngine(base_url="http://localhost:4200", workspace="default")
profile = client.profile(audience="sales", bandwidth="l1")
result  = client.ask("healthtech pricing")
client.memory.create(content="Bob owns Q4 pricing", is_static=True)

agent = Agent(name="assistant", tools=optimal_engine_tools(client))
```

LangChain adapter:

```python
from optimal_engine.adapters.langchain import build_tools

tools = build_tools(client)  # 8 BaseTool subclasses
```

---

## MCP server (Claude Desktop, Cursor, Windsurf, Zed)

```json
// claude_desktop_config.json
{
  "mcpServers": {
    "optimal-engine": {
      "command": "npx",
      "args": ["-y", "@optimal-engine/mcp"],
      "env": {
        "OPTIMAL_ENGINE_URL": "http://localhost:4200",
        "OPTIMAL_WORKSPACE": "engineering"
      }
    }
  }
}
```

Exposed tools: `ask`, `search`, `grep`, `profile`, `add_memory`, `forget_memory`, `recall`, `wiki_get`, `workspaces`.

---

## The second brain for your agent

The Optimal Engine is designed to be the **memory backend for any agent harness** — whether that's OSA (Operating System Agent), BusinessOS, Claude Code, a custom LangChain/LangGraph pipeline, or a bare-bones script. The engine doesn't care what agent talks to it; it cares that every agent gets cited, permission-scoped, audience-shaped context.

### How agent integration works

```
┌─────────────────────────────────────────────────┐
│  YOUR AGENT HARNESS                             │
│  (OSA · BusinessOS · Claude Code · LangChain)   │
│                                                 │
│  ┌───────────┐  ┌──────────┐  ┌──────────────┐ │
│  │ CLI tools  │  │ MCP      │  │ SDK client   │ │
│  │ mix optimal│  │ 9 tools  │  │ TS / Python  │ │
│  └─────┬──────┘  └────┬─────┘  └──────┬───────┘ │
│        │              │               │          │
│        └──────────────┼───────────────┘          │
│                       ▼                          │
│               HTTP API (:4200)                   │
└───────────────────────┬─────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────┐
│  OPTIMAL ENGINE                                 │
│  9-stage pipeline → 3-tier memory → wiki        │
│  workspaces · memories · surfacing              │
└─────────────────────────────────────────────────┘
```

Three integration surfaces — use whichever fits your agent's runtime:

| Surface | Best for | Example |
|---|---|---|
| **CLI** (`mix optimal.*`) | Shell agents, scripts, CI pipelines | `mix optimal.rag "pricing" --workspace engineering --format claude` |
| **MCP** (stdio, 9 tools) | Claude Desktop, Cursor, Windsurf, Zed, any MCP-compatible agent | Add to `claude_desktop_config.json` and the agent gains memory |
| **SDK** (TypeScript / Python) | Custom agents, LangChain, Vercel AI SDK, OpenAI Agents SDK | `client.ask(query)` / `client.memory.create(...)` |

### Example: OSA agent with Optimal Engine as its brain

OSA (Operating System Agent) uses Optimal Engine as its long-term memory and organizational context layer. The agent runtime dispatches sub-agents; each reads from and writes to the engine.

```typescript
// OSA agent harness — wiring the engine as the second brain
import { OptimalEngine } from "@optimal-engine/client";
import { optimalEngineTools } from "@optimal-engine/client/adapters/ai-sdk";

const brain = new OptimalEngine({
  baseUrl: "http://localhost:4200",
  workspace: "engineering",
});

// 1. Load context at session start — the agent knows who it is
const profile = await brain.profile({ audience: "engineering", bandwidth: "l1" });
const systemPrompt = `
You are an engineering assistant.

## What you know (from the second brain)
${profile.static}

## Recent context
${profile.dynamic}

## Curated summary
${profile.curated}
`;

// 2. During the conversation — recall on demand
const decisions = await brain.recall.actions({ topic: "microvm isolation" });
const owner = await brain.recall.who({ topic: "compute engine" });

// 3. After the conversation — persist what was learned
await brain.memory.create({
  content: "Team decided to use Firecracker for tenant isolation",
  isStatic: true,
  citationUri: "optimal://nodes/02-platform/signals/2026-04-30-standup.md",
});

// 4. Tools auto-wired — agent can search/recall/remember on its own
const tools = optimalEngineTools(brain);
// Pass `tools` to your LLM call (Vercel AI SDK, OpenAI, Anthropic)
```

### Example: BusinessOS orchestrator with workspace-per-module

BusinessOS runs multiple domain modules (CRM, HR, Finance, Ops). Each module gets its own workspace in the engine — isolated data, isolated wiki, isolated surfacing.

```typescript
const crm    = new OptimalEngine({ baseUrl: engine, workspace: "crm" });
const hr     = new OptimalEngine({ baseUrl: engine, workspace: "hr" });
const finance = new OptimalEngine({ baseUrl: engine, workspace: "finance" });

// CRM agent asks its own brain
const deals = await crm.ask("renewal pipeline status", { audience: "sales" });

// HR agent asks its own brain — no CRM data leaks through
const policies = await hr.ask("remote work policy", { audience: "exec" });

// Cross-module: orchestrator can read from multiple workspaces
const crmProfile = await crm.profile({ bandwidth: "l0" });
const finProfile = await finance.profile({ bandwidth: "l0" });
```

### Example: Claude Code with the Claude Skill

Drop the skill into your Claude Code setup and the engine becomes ambient context:

```bash
# Install the skill
cp -r skills/optimal-engine/ ~/.claude/skills/optimal-engine/

# Claude Code now:
# - Loads engine context when relevant queries arise
# - Has access to 6 reference docs (API, concepts, patterns)
# - Can bootstrap workspaces via the included script
```

Or wire via MCP for tool-level integration:

```json
{
  "mcpServers": {
    "optimal-engine": {
      "command": "npx",
      "args": ["-y", "@optimal-engine/mcp"],
      "env": {
        "OPTIMAL_ENGINE_URL": "http://localhost:4200",
        "OPTIMAL_WORKSPACE": "default"
      }
    }
  }
}
```

Now Claude Code can `ask`, `search`, `grep`, `add_memory`, `forget_memory`, `recall`, and `get_profile` against the engine — your second brain is always in context.

### Example: Python agent with LangChain

```python
from optimal_engine import OptimalEngine
from optimal_engine.adapters.langchain import build_tools
from langchain.agents import AgentExecutor, create_openai_tools_agent
from langchain_openai import ChatOpenAI

brain = OptimalEngine(base_url="http://localhost:4200", workspace="research")

# 8 tools: ask, search, grep, profile, add_memory, forget_memory, recall, wiki_get
tools = build_tools(brain)

llm = ChatOpenAI(model="gpt-4o")
agent = create_openai_tools_agent(llm, tools, prompt)
executor = AgentExecutor(agent=agent, tools=tools)

result = executor.invoke({"input": "What decisions were made about pricing?"})
```

### Example: Batch migration from another system

Moving from Notion / Confluence / a custom wiki? Bulk-import via the batch API:

```bash
# Export from your existing system to JSON, then:
curl -X POST http://localhost:4200/api/batch/import/signals \
  -H 'content-type: application/json' \
  -d '{
    "workspace": "migrated",
    "signals": [
      {"content": "Q3 pricing decision: $2K/seat", "title": "Pricing", "genre": "decision", "node": "sales"},
      {"content": "Platform uses Firecracker for isolation", "title": "Infra", "genre": "spec", "node": "platform"},
      ...
    ]
  }'
# → {"imported": 142, "skipped": 3, "errors": 0}

# Export a full workspace snapshot (for backup or migration to another engine instance):
curl http://localhost:4200/v1/batch/export/workspace?workspace=engineering > snapshot.json
```

### Example: Proactive surfacing with webhooks

Subscribe to topics and get POSTed when something relevant surfaces — no polling, no query:

```bash
# Create a webhook subscription
curl -X POST http://localhost:4200/v1/subscriptions \
  -H 'content-type: application/json' \
  -d '{
    "workspace": "engineering",
    "scope": "topic",
    "scope_value": "pricing",
    "categories": ["recent_actions", "ownership"],
    "webhook_url": "https://your-agent.example.com/hooks/brain",
    "webhook_secret": "whsec_abc123"
  }'

# Your agent receives POSTs like:
# POST /hooks/brain
# X-Optimal-Signature: sha256=<hmac>
# {
#   "trigger": "wiki_updated",
#   "envelope": {"slug": "pricing-decision", "kind": "wiki_page"},
#   "category": "recent_actions",
#   "score": 0.95
# }
```

---

## Desktop UI

`desktop/` — SvelteKit shell with Foundation tokens. Glass header, two-dropdown workspace switcher, 12 routes:

| Route | What it shows |
|---|---|
| `/` | **Ask** — wiki-first query with seed prompts |
| `/surface` | **Surface** — proactive memory pushes via SSE |
| `/memory` | **Memory** — versioned graph with 5 typed relation edges |
| `/timeline` | **Timeline** — SVG signal timeline grouped by intent |
| `/heatmap` | **Heatmap** — node × week activity grid |
| `/workspace` | **Nodes** — granularity explorer: tree → signals → S=(M,G,T,F,W) → 4 scales → entities → citations |
| `/graph` | Three.js 3D entity graph |
| `/wiki` | Tier-3 page list + render with citation highlighting |
| `/architectures` | DataArchitecture catalog + processor registry |
| `/activity` | Append-only events feed with auto-refresh |
| `/status` | Liveness + readiness checks |

```bash
cd desktop
npm install
npm run dev          # browser preview
npm run tauri:dev    # native window (requires Rust + Tauri)
```

---

## Extensions

### Browser (Chrome MV3)

```bash
cd extensions/browser
npm install
npm run build
# Then load extensions/browser/dist/ as unpacked extension in Chrome
```

Popup with workspace switcher, search, "Clip page" button, recent clips list. Right-click selection → "Save to Optimal Engine".

### Raycast (macOS)

```bash
cd extensions/raycast
npm install
npm run dev          # requires Raycast.app installed
```

Three commands: `Search Memory`, `Add Memory`, `Ask Engine`.

---

## Documentation

| Doc | Purpose |
|---|---|
| [`PLAN.md`](PLAN.md) | **Master plan** — every phase, every decision |
| [`apps/docs/`](apps/docs/) | **Documentation site** — quickstart, concepts, API reference, SDK guides, self-host |
| [`docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md) | Canonical 9-stage pipeline + 3 tiers + data contracts |
| [`docs/architecture/DATA_ARCHITECTURE.md`](docs/architecture/DATA_ARCHITECTURE.md) | Universal data-point layer — 12 modalities, 7 architectures, 6 processors |
| [`docs/architecture/ENTERPRISE.md`](docs/architecture/ENTERPRISE.md) | Tenancy, ACLs, connectors, retention, audit, performance targets |
| [`docs/architecture/WIKI-LAYER.md`](docs/architecture/WIKI-LAYER.md) | Tier 3 deep dive: directives, curator, integrity, schema governance |
| [`docs/concepts/signal-theory.md`](docs/concepts/signal-theory.md) | `S=(M,G,T,F,W)` + 4 constraints + 6 principles + 11 failure modes |
| [`skills/optimal-engine/`](skills/optimal-engine/) | Claude Code Skill — drop into `~/.claude/skills/` for ambient context |
| [`sample-workspace/README.md`](sample-workspace/README.md) | On-disk convention reference with filled-in example files |
| [`deploy/README.md`](deploy/README.md) | Production deployment guide |

---

## Configuration

Per-workspace YAML at `<workspace>/.optimal/config.yaml`:

```yaml
visualizations:
  enabled: [timeline, heatmap, graph, contradictions]
  timeline:
    group_by: intent
    default_window_days: 30
  heatmap:
    granularity: week

profile:
  default_audience: default
  recent_chunks_limit: 20

grep:
  default_scale: paragraph
  literal_threshold: 0.8

contradictions:
  policy: flag_for_review        # flag_for_review | silent_resolve | reject

memory:
  extract_from_wiki: false
  auto_promote_to_wiki: false
  dedup_policy: return_existing  # return_existing | bump_version | always_insert

rate_limit:
  requests_per_minute: 100
  burst_capacity: 200

retention:
  default_ttl_days: null
  archive_after_days: 365
```

Engine-wide config in `config/runtime.exs`:

```elixir
config :optimal_engine, :api,
  enabled: System.get_env("OPTIMAL_API_ENABLED", "true") == "true",
  port:    String.to_integer(System.get_env("OPTIMAL_API_PORT", "4200")),
  interface: System.get_env("OPTIMAL_API_INTERFACE", "127.0.0.1")

config :optimal_engine, :auth,
  auth_required: System.get_env("OPTIMAL_AUTH_REQUIRED", "false") == "true",
  bcrypt_cost: 12
```

---

## Development

```bash
mix test                   # full suite (1,297 tests, ~10s)
mix format                 # after edits
mix credo                  # lints
mix dialyzer               # optional — slow first run
mix optimal.reality_check --hard   # 50+ end-to-end probes
```

Optional shell tools (graceful degradation when absent):

```bash
brew install pdftotext tesseract ffmpeg
# whisper.cpp via Ollama or separate install
```

---

## License

MIT. Use it, fork it, ship it.

## Contributing

Issues + PRs welcome. See [`CONTRIBUTING.md`](CONTRIBUTING.md) when it lands. Per-phase build sequence in [`PLAN.md`](PLAN.md).
