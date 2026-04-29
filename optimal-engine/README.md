# Optimal Engine

**The second brain of a company.** An Elixir application that ingests every
signal flowing through an organization — chat, email, docs, meetings, tickets,
CRM, code, voice, video — decodes the intent behind each one, embeds it into a
multi-modal aligned vector space, clusters it wide by theme, and delivers the
right chunks at the right grain to any agent runtime that queries it, scoped
to what the caller is allowed to see.

> **Positioning:** Classical RAG re-discovers the same facts on every query.
> The Optimal Engine discovers once, curates forever, and delivers
> permission-gated, receiver-matched context to any agent runtime in under
> 200ms.

---

## Status

Active development. Per-phase progress against [`PLAN.md`](PLAN.md):

| Phase | Scope | Status |
|---|---|---|
| 0    | Extract engine into standalone repo, absorb 3 MIOSA subsystems | ✅ |
| 0.5  | Restructure `lib/` to canonical layout (pipeline / retrieval / store / wiki / insight / graph / embed / tenancy / identity / audit / connectors) | ✅ |
| 1    | Schema + tenancy + identity + ACLs + audit foundation (20 new tables, 16 versioned migrations, principal-scoped search) | ✅ |
| 3.5  | Workspace — nodes + members + skills (6 new tables, node tree + skill levels + internal/external memberships) | ✅ |
| 2    | Parser backends — 10 formats (md, yaml/json, csv, html, code, pdf, office, image-OCR, audio-whisper, video) | ✅ |
| 3    | Decomposer — hierarchical chunking at 4 scales | ✅ |
| 4    | Per-chunk classify + intent extract (10-value enum) | ✅ |
| 5    | Multi-modal embedder — nomic-embed-text + nomic-embed-vision + whisper.cpp, all 768-dim aligned | ✅ |
| 6    | Clusterer — incremental greedy theme grouping with weighted similarity | ✅ |
| 7    | Wiki Layer — LLM-maintained curated top layer with hot citations + executable directives | ✅ |
| 8    | Scale-aware Deliver + Composer + RAG | ✅ |
| 9    | 14 enterprise connectors (Slack / Gmail / Drive / Notion / Jira / Linear / GitHub / Zoom / …) | ✅ |
| 10   | Production hardening — perf targets, release, backup/restore | ✅ |
| 11   | Compliance — SOC 2, GDPR, HIPAA | ✅ |
| 12   | Desktop UI — Tauri + SvelteKit + HTTP API | ✅ |
| 13   | v0.1.0 tag | ✅ |
| 14   | DataArchitecture layer — model-agnostic data-point schema + processor registry; workspace / activity / architectures UI | ✅ |

**Current suite:** 1,075 tests passing, 29 excluded (RocksDB NIF, optional backend).

---

## The three tiers

```
TIER 3 — THE WIKI              LLM-maintained. Read first. Audience-aware.
Path: .wiki/                   Hot citations + executable directives.
           ▲ CURATE ▼
TIER 2 — DERIVATIVES           Machine-maintained. Rebuildable.
Path: .optimal/index.db        SQLite + FTS5 + vectors + graph + clusters.
           ▲ DERIVE ▼
TIER 1 — RAW SOURCES           Immutable. Append-only. Hash-addressed.
Path: nodes/**/signals/*        Signal files, PDFs, images, audio, video.
      assets/                   The engine NEVER rewrites them.
```

## The 9-stage ingestion pipeline

```
1. INTAKE → 2. PARSE → 3. DECOMPOSE → 4. CLASSIFY → 5. EMBED → 6. ROUTE
                                                                    │
9. CURATE ← 8. CLUSTER ← 7. STORE ←────────────────────────────────┘
```

Each stage has one responsibility and a typed contract with the next. See
[`docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md) for
per-stage detail.

---

## Quick start

Requires Elixir `~> 1.17`, Erlang/OTP 26+, a C toolchain (for the exqlite NIF).

```bash
git clone git@github.com:robertohluna/OptimalEngine.git
cd OptimalEngine
mix deps.get
mix compile
mix test                    # expect 1,075/1,075 passing
```

First ingest + search:

```bash
# Walk the sample-workspace/ directory — 6 nodes, 13 signals, 2
# curated wiki pages — and run them through the full pipeline.
mix optimal.ingest_workspace sample-workspace/

mix optimal.search "platform"
mix optimal.rag "healthtech pricing decision" --trace
mix optimal.wiki list
```

See [`sample-workspace/README.md`](sample-workspace/README.md) for the
on-disk convention (nodes / signals / wiki / architectures) with
filled-in example files. Scaffold a fresh workspace from it with:

```bash
mix optimal.init ~/my-engine
```

Browse the engine in a desktop window:

```bash
# In config/dev.exs enable the HTTP API
#   config :optimal_engine, :api, enabled: true, port: 4200
iex -S mix                  # starts the engine + API

# in a second terminal
cd desktop
npm install
npm run dev                 # browser preview at http://localhost:1420
npm run tauri:dev           # native window (once Rust + Tauri are set up)
```

Check migration state:

```bash
mix optimal.migrate --status   # lists applied + pending migrations
mix optimal.stats              # row counts across all tables
```

---

## Positioning

The typical AI-context stack looks like:

```
Storage → Retrieval → Dump to LLM
```

The Optimal Engine puts a full pipeline under it:

```
Classification → Routing → Composition → Tiered Assembly → Delivery → Feedback
                          (with any storage engine underneath)
```

### Capability grid

Reference points across the enterprise-RAG and memory-layer space — framed by what ships today, not who got there first.

| Capability | Glean | Dust | NotebookLM | OpenViking | Notion AI | Mem0 / Letta / Zep | Pinecone / Weaviate / LlamaIndex | **Optimal Engine** |
|---|---|---|---|---|---|---|---|---|
| Enterprise connectors | ✅ | ✅ | ⚠️ | ❌ | ⚠️ | ❌ | ❌ | ✅ 14 adapters (Phase 9) |
| Permission-aware RAG | ✅ | ✅ | ⚠️ | ❌ | ✅ | ❌ | ❌ | ✅ chunk-level + intersection propagation |
| Signal classification `S=(M,G,T,F,W)` | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Intent extraction per chunk | ❌ | ❌ | ❌ | ❌ | ❌ | ⚠️ | ❌ | ✅ 10-value enum |
| Multi-modal aligned embeddings | ⚠️ | ⚠️ | ✅ | ❌ | ❌ | ❌ | ⚠️ | ✅ nomic 768-dim text + vision |
| Cross-modal retrieval (text ↔ image) | ❌ | ❌ | ⚠️ | ❌ | ❌ | ❌ | ⚠️ | ✅ |
| Audience-aware curation variants | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| Hierarchical chunking (4 scales) | ⚠️ | ⚠️ | ⚠️ | ✅ | ❌ | ❌ | ⚠️ | ✅ doc / sec / para / sent |
| Incremental LLM-curated wiki | ❌ | ❌ | ❌ | ✅ single-loop | ❌ | ❌ | ❌ | ✅ triple-loop SICA + citations |
| Episodic → semantic memory consolidation | ❌ | ⚠️ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ Memory.Cortex |
| OWL / neuro-symbolic reasoner | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ Knowledge module |
| Hybrid retrieval (BM25 + vector + graph) | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ✅ wiki-first → hybrid fallback |
| Model-agnostic processor registry | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ DataArchitecture (Phase 14) |
| GDPR Art. 17 erasure + legal hold | ⚠️ | ⚠️ | ❌ | ❌ | ⚠️ | ❌ | ❌ | ✅ Phase 11 |
| Agent-runtime integration (any lang) | API | API | Workspace | API | API | SDK | SDK | ✅ CLI + HTTP + MCP + Elixir |
| Local-first / self-hosted | ❌ | ❌ | ❌ | ✅ | ❌ | ⚠️ | ⚠️ | ✅ |

Reads: ✅ ships, ⚠️ partial / paid tier / beta, ❌ not in the shipping product.

---

## Mix tasks

**Core ingest + query**

| Task | Purpose |
|---|---|
| `mix optimal.ingest <text\|--file path>` | Classify, route, persist, and index a signal |
| `mix optimal.search <query>` | Hybrid BM25 + vector + graph + temporal. `--principal` enforces ACLs. |
| `mix optimal.read <optimal://uri>` | Read a context at a given tier |
| `mix optimal.assemble <topic>` | Build a tiered (L0/L1/L2) context bundle |
| `mix optimal.rag <query>` | Wiki-first retrieval envelope (`--format plain\|markdown\|claude\|openai`, `--audience`, `--bandwidth`, `--trace`) |
| `mix optimal.l0` | Print the always-loaded L0 context |
| `mix optimal.ls <optimal://uri-prefix>` | List contexts under a URI |
| `mix optimal.seed [--reset]` | Load the realistic demo dataset (nodes / signals / chunks / entities / clusters / wiki / events) |

**Graph + insight**

| Task | Purpose |
|---|---|
| `mix optimal.graph [triangles\|clusters\|hubs]` | Knowledge-graph analysis |
| `mix optimal.reflect [--min N]` | Find missing edges from entity co-occurrence |
| `mix optimal.reweave <topic>` | Find stale contexts on a topic |
| `mix optimal.simulate <scenario>` | "What if" scenario through the graph |
| `mix optimal.impact <entity-or-node>` | Impact analysis |
| `mix optimal.remember <observation>` | Store an observation for the SICA learning loop |
| `mix optimal.rethink <topic>` | Synthesize observations into actionable knowledge |

**Wiki + architectures**

| Task | Purpose |
|---|---|
| `mix optimal.wiki [list\|view\|verify\|verify-all]` | Operate on Tier-3 wiki pages |
| `mix optimal.architectures [show <name>\|processors]` | Inspect the DataArchitecture catalog + processor registry |

**Ops + compliance**

| Task | Purpose |
|---|---|
| `mix optimal.status [--json\|--quick\|--metrics]` | Live readiness (supervisor / store / migrations / credentials / embedder) |
| `mix optimal.backup <path> [--verify]` | Online SQLite backup via `VACUUM INTO`, optional integrity check |
| `mix optimal.health [--quick]` | Knowledge-base diagnostic checks (orphans, drift, duplicates) |
| `mix optimal.verify [--sample N]` | Cold-read test of L0 abstract fidelity |
| `mix optimal.stats` | Row counts + Phase 1 table counts |
| `mix optimal.migrate [--status]` | Apply pending migrations / show version map |
| `mix optimal.connector {list\|register\|run}` | Manage enterprise connectors |
| `mix optimal.compliance {dsar\|erase\|hold\|retention}` | GDPR / CCPA / HIPAA workflows |
| `mix optimal.knowledge [metrics]` | Knowledge-graph stats + SICA patterns |
| `mix optimal.graph_ui` | Launch the standalone graph visualizer |
| `mix optimal.spec.{init,drift,report}` | Spec-led-dev tooling |

Full reference: [`docs/guides/mix-tasks.md`](docs/guides/mix-tasks.md).

---

## Agent runtime integration

Any runtime that can shell out to a command can use the engine:

```python
# Python
import subprocess, json
ctx = json.loads(subprocess.check_output([
    "mix", "optimal.rag", "current state of Ed's pricing negotiation",
    "--format", "json", "--limit", "6"
]))
```

```javascript
// Claude Agent SDK / Node
const ctx = await exec(`mix optimal.rag "${query}" --format claude`);
const response = await anthropic.messages.create({
  model: "claude-opus-4-7",
  system: ctx,
  messages: [{role: "user", content: userAsk}]
});
```

```go
// Go
ctx := exec.Command("mix", "optimal.rag", query, "--format", "openai").Output()
```

Same data. Same guarantees. Same HTTP API behind it.

---

## HTTP API

Enable via `config :optimal_engine, :api, enabled: true, port: 4200`.

| Endpoint | Purpose |
|---|---|
| `GET  /api/status` | Liveness + readiness (store / migrations / credentials / embedder) |
| `GET  /api/metrics` | Telemetry snapshot (counters, histograms, uptime) |
| `POST /api/rag` | Wiki-first retrieval envelope with receiver profile |
| `GET  /api/search?q=…` | Hybrid search hits |
| `GET  /api/l0` | L0 context cache |
| `GET  /api/wiki` · `/api/wiki/:slug` | List + render Tier-3 pages |
| `GET  /api/workspace` | Node forest (parent/child + signal counts) |
| `GET  /api/signals/:id` | Full granularity for one data point — dimensions, 4-scale chunks, entities by type, classification, intent, cluster membership, citations, architecture binding |
| `GET  /api/activity?kind=…` | Append-only events feed |
| `GET  /api/architectures` · `/:id` | DataArchitecture catalog + processor registry |
| `GET  /api/graph` · `/graph/hubs` · `/triangles` · `/clusters` · `/reflect` | Graph analysis |
| `GET  /api/optimal/graph` · `/optimal/nodes` · `/optimal/nodes/:slug/files` | Desktop knowledge-graph feed |
| `GET  /api/node/:node_id` | Subgraph for one node |
| `GET  /api/health` | Knowledge-base diagnostic checks |

---

## Desktop UI

`desktop/` is a Tauri + SvelteKit shell that talks to the HTTP API above.

```bash
cd desktop
npm install
npm run dev        # browser preview at http://localhost:1420
npm run tauri:dev  # native window
```

| Route | What it shows |
|---|---|
| `/` | Ask — wiki-first query over `POST /api/rag` |
| `/workspace` | 3-pane granularity explorer: node tree → signals → S=(M,G,T,F,W) dims, 4-scale chunks, entities, classification, intent, clusters, citations |
| `/graph` | Three.js 3D force-directed entity graph + node drill-down (ported from BusinessOS-5) |
| `/wiki` | Tier-3 page list + render |
| `/architectures` | DataArchitecture catalog + processor registry |
| `/activity` | Append-only events feed with 5s auto-refresh + kind filter |
| `/status` | Liveness / readiness / degraded-checks report |

---

## Architecture

```
lib/optimal_engine/
├── optimal_engine.ex           # top-level facade (defdelegate everything)
├── application.ex              # unified supervision tree
├── cli.ex                      # escript entry
│
├── pipeline/                   # Stages 1–9 (ingest path)
├── retrieval/                  # Serving path (the "read" side)
├── store/                      # SQLite + FTS + vectors + migrations
├── wiki/                       # Tier 3 — LLM-maintained (Phase 7)
├── architecture/               # DataArchitecture layer (Phase 14) — fields,
│                               # processors, registry, built-in catalog
│
├── knowledge/ memory/ signal/  # Absorbed subsystems (OWL, episodic, CloudEvents)
│
├── graph/                      # Graph analysis (not storage)
├── insight/                    # Learning + synthesis (verify/health/rethink/…)
├── embed/                      # Embedding providers (Ollama; Phase 5 adds whisper)
│
├── tenancy/                    # Multi-tenant primitives (Phase 1)
├── identity/                   # Principal / Group / Role / ACL (Phase 1)
├── audit/                      # Event + Logger (Phase 1)
├── connectors/                 # Phase 9 — enterprise integrations
├── compliance/                 # Phase 11 — PII / DSAR / Erasure / LegalHold / Retention
│
├── telemetry.ex health.ex      # Phase 10 — observability + readiness
├── backup.ex release.ex        # Phase 10 — VACUUM INTO + release entry points
│
├── api/                        # HTTP surface (Plug + Cowboy, opt-in)
└── spec/                       # Spec-led-dev tooling
```

Stable root primitives: `context.ex` `session.ex` `session_compressor.ex`
`signal.ex` `topology.ex` `uri.ex` `graph.ex` `knowledge.ex` `memory.ex`
`architecture.ex` `compliance.ex`.

---

## Documentation map

| Doc | Purpose |
|---|---|
| [`PLAN.md`](PLAN.md) | **Master plan** — every phase, every decision, every open question |
| [`docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md) | Canonical 9-stage pipeline + 3 tiers + data contracts |
| [`docs/architecture/DATA_ARCHITECTURE.md`](docs/architecture/DATA_ARCHITECTURE.md) | **Phase 14** — universal data-point layer: `Field`, `Architecture`, `Processor`, SOTA inspirations |
| [`docs/architecture/ENTERPRISE.md`](docs/architecture/ENTERPRISE.md) | Tenancy, ACLs, connectors, retention, audit, performance targets |
| [`docs/architecture/WIKI-LAYER.md`](docs/architecture/WIKI-LAYER.md) | Tier 3 deep dive: directives, curator, integrity, schema governance |
| [`docs/architecture/UI.md`](docs/architecture/UI.md) | Desktop app spec — Tauri + SvelteKit, 7 views |
| [`docs/concepts/signal-theory.md`](docs/concepts/signal-theory.md) | `S=(M,G,T,F,W)` + 4 constraints + 6 principles + 11 failure modes |
| [`docs/guides/getting-started.md`](docs/guides/getting-started.md) | Clone → compile → first ingest + search |
| [`docs/guides/mix-tasks.md`](docs/guides/mix-tasks.md) | Every `mix optimal.*` command |
| [`docs/architecture/00-overview.md`](docs/architecture/00-overview.md) … `07-governance.md` | Original 7-layer system architecture |

---

## Development

```bash
mix test                  # full suite (1,075 tests, ~90s)
mix format                # after edits
mix credo                 # lints
mix dialyzer              # optional — slow first run
```

Optional shell tools enhance text-extraction for binary formats (installed on
demand; missing tools cause graceful degradation, never crashes):

```bash
brew install pdftotext tesseract ffmpeg
# whisper.cpp via separate install or use Ollama's whisper support
```

---

## Status

Private. Active development. Not yet published. Not yet accepting external
contributions. See [`PLAN.md`](PLAN.md) for the per-phase build sequence.
