# Optimal Engine — Master Plan

> Single source of truth for what we're building, why, and the exact sequence
> of work to build it. If this plan and the code disagree, the plan wins until
> a decision log entry below says otherwise.
>
> **Status:** local repo green (689/689 tests), private remote at
> `robertohluna/OptimalEngine` created, **not pushed**. Holding the push until
> the build hits the production bar defined in Phase 11.

---

## Part 1 — Thesis

The Optimal Engine is **the second brain of a company**. It ingests every
signal flowing through an organization (chat, email, docs, meetings, tickets,
CRM, code, voice, video), breaks each down from large to small, decodes its
intent at every scale, clusters it wide by theme, and delivers the right
chunks at the right grain to any agent or human who queries it — scoped to
what that caller is allowed to see.

On top of raw storage sits an **LLM-maintained Wiki** — the front door every
agent reads first. Wiki pages carry **hot citations** back to source material
and **executable directives** that pull more context on demand. Agents don't
re-discover facts on every query; the wiki already knows them.

**Three promises:**

1. **Signal integrity** — nothing is stored without classification, intent, and citation lineage. The engine knows *why* each fact exists.
2. **Scale alignment** — every piece of content exists simultaneously as document / section / paragraph / chunk. Retrieval returns the coarsest scale that answers the query.
3. **Modality alignment** — text, image, audio embed into the same 768-dim space (nomic-embed-text / nomic-embed-vision / whisper.cpp). A text query can retrieve an image. One retrieval layer, not three.

**Positioning soundbite:** "Classical RAG re-discovers the same facts on every query. The Optimal Engine discovers once, curates forever, and delivers permission-gated, receiver-matched context to any agent runtime in under 200ms."

---

## Part 2 — The architecture in one glance

Full detail lives in `docs/architecture/`. This is the recap.

### The three tiers

```
TIER 3 — THE WIKI              LLM-maintained. Read first. Audience-aware.
Path: .wiki/                   Hot citations. Executable directives.
                               (See docs/architecture/WIKI-LAYER.md)
           ▲ CURATE ▼
TIER 2 — DERIVATIVES           Machine-maintained. Rebuildable.
Path: .optimal/index.db         SQLite + FTS5 + vectors + graph +
                               clusters + L0 abstracts. Produced by
                               the 9-stage pipeline.
           ▲ DERIVE ▼
TIER 1 — RAW SOURCES           Immutable. Append-only. Hash-addressed.
Path: nodes/**/signals/*        Signal files, PDFs, images, audio, video.
      assets/                  The engine NEVER rewrites them.
```

### The nine stages

```
1. INTAKE   → 2. PARSE    → 3. DECOMPOSE → 4. CLASSIFY → 5. EMBED
    any format    text+structure  large→small     S=(M,G,T,F,W)   multi-modal
    + hash        metadata        hierarchical    + intent        aligned 768-dim

6. ROUTE    → 7. STORE    → 8. CLUSTER   → 9. CURATE
    node +        atomic          HDBSCAN          affected wiki
    topology      SQLite txn      incremental      pages → Ollama
                                                   curator → commit
```

### Retrieval flow

```
AGENT QUERY
    ↓
IntentAnalyzer (decode query: lookup / compare / temporal / decide / explore)
    ↓
WIKI LOOKUP (Tier 3 — always tried first)
    ├─ sufficient → Composer formats output → return
    ↓ insufficient
HYBRID RETRIEVAL (Tier 2 — BM25 + vector + graph + intent-match + temporal + cluster-expand)
    ↓
ContextAssembler (scale-aware: returns coarsest chunk that answers)
    ↓
Composer (format: text / json / claude / openai / tool-result)
    ↓
PACKAGED CONTEXT → agent
```

### Five invariants (violate any → engine is broken)

1. Tier 1 is append-only. Nothing rewrites in place.
2. Tier 2 is fully derivable from Tier 1 (`mix optimal.rebuild` reconstructs).
3. Tier 3 is LLM-owned; humans write the schema, curator writes the pages.
4. Tenant isolation is absolute. No cross-tenant reads, ever.
5. Permissions propagate by **intersection**, never union, never inferred.

---

## Part 3 — Competitive position

### Category

Application-layer enterprise context platform. Direct rivals: **Glean**, **Dust**, **Google NotebookLM**. Infrastructure overlap with **OpenViking** (ByteDance), **MemOS**. Memory primitives differentiate us from **Mem0 / Letta / Zep / Cognee / HydraDB**.

### What everyone else built

```
Storage → Retrieval → Dump to LLM
```

### What we build

```
Classification → Routing → Composition → Tiered Assembly → Delivery → Feedback
                          (with any storage engine underneath)
```

### Feature comparison

| Capability                               | Glean | Dust | NotebookLM | OpenViking | LLM Wiki (nashsu) | **Optimal Engine** |
|-----------------------------------------|-------|------|-----------|-----------|-------------------|--------------------|
| Enterprise connectors                    | ✅ 100+ | ✅    | ⚠️ Workspace | ❌         | ❌ clipper only   | ✅ 14 in Phase 9   |
| Permission-aware RAG                     | ✅     | ✅    | ⚠️          | ❌         | ❌                | ✅ chunk-level + intersection propagation |
| Signal classification `S=(M,G,T,F,W)`    | ❌     | ❌    | ❌          | ❌         | ❌                | ✅                 |
| Intent extraction per chunk              | ❌     | ❌    | ❌          | ❌         | ❌                | ✅ 10-value enum   |
| Tiered disclosure L0/L1/L2               | ❌     | ❌    | ❌          | ✅         | ❌                | ✅ + receiver bandwidth |
| Hierarchical chunking (4 scales)         | ❌     | ❌    | ❌          | ❌         | ❌                | ✅                 |
| Multi-modal aligned embeddings           | ⚠️     | ⚠️    | ✅          | ❌         | ❌                | ✅ nomic 768-dim   |
| Cross-modal retrieval (text→image)       | ❌     | ❌    | ⚠️          | ❌         | ❌                | ✅                 |
| OWL 2 RL reasoning                       | ❌     | ❌    | ❌          | ❌         | ❌                | ✅                 |
| Hot citations with integrity check       | ⚠️     | ⚠️    | ✅          | ❌         | ✅                | ✅                 |
| Executable directives                    | ❌     | ❌    | ❌          | ❌         | `[[wikilink]]` only| `{{cite/include/expand/search/table/trace/recent}}` |
| Audience-aware wiki variants             | ❌     | ❌    | ❌          | ❌         | ❌                | ✅                 |
| Triggered incremental curation           | ❌     | ❌    | ❌          | ✅ single-loop | ✅ single-loop | ✅ **triple-loop SICA** |
| Agent-runtime integration (any lang)     | API   | API  | Workspace   | API       | desktop only      | ✅ CLI + HTTP + MCP + Elixir |
| Local-first / self-hosted                | ❌     | ❌    | ❌          | ✅         | ✅                | ✅                 |
| SIEM-exportable audit                    | ✅     | ✅    | ❌          | ❌         | ❌                | ✅                 |

### Asymmetric bets

1. **Classification is deterministic and cheap.** Every signal gets `S=(M,G,T,F,W) + intent` at ingest. No LLM call for classification (rules + heuristics, optional LLM augmentation). Rivals rely on expensive per-query LLM tagging that degrades at scale.
2. **Audience-aware wiki variants.** One signal → N curated views, each permission-gated. Filtered at **curation** time, not query time. Strictly safer than query-time filtering.
3. **Agents read wiki first.** At 10K employees × 50 queries/day, wiki-first retrieval saves ~20 engineer-hours of compute daily vs per-query RAG.

---

## Part 4 — Enterprise-scale requirements

Full detail: `docs/architecture/ENTERPRISE.md`.

### Dimensions

| Dimension        | Target for v1                                                           |
|------------------|-------------------------------------------------------------------------|
| Users per tenant | 100–10,000 humans × M agents per human × service accounts              |
| Signals          | 10M–1B+ per tenant, 10TB+ raw, millions ingested per day                |
| Sources          | 14 connectors in Phase 9 (Slack, Gmail, Drive, Docs, M365, Notion, Jira, Linear, GitHub, Zoom, Confluence, Salesforce/HubSpot, Intercom, webhook) |
| Tenancy          | Per-tenant DB + wiki. No cross-tenant access at DB level.              |
| Identity         | SAML 2.0, OIDC, SCIM 2.0 group sync                                     |
| ACLs             | Chunk-level `acl_read / acl_write / classification_level`               |
| Retention        | Per-node, per-genre, per-tenant TTL + cold archive + legal hold         |
| Compliance       | SOC 2 Type II, GDPR, HIPAA, CCPA, ISO 27001 (Phase 11)                  |
| Deployment       | SaaS / Docker / K8s / hybrid (metadata cloud + raw on-prem)             |

### Performance targets (Phase 10 acceptance)

| Operation                              | p50     | p99      |
|---------------------------------------|---------|----------|
| Wiki page read (cache hit)             | < 20ms  | < 100ms  |
| Wiki page read (cache miss)            | < 80ms  | < 300ms  |
| `optimal rag` query (wiki hit)         | < 50ms  | < 200ms  |
| `optimal rag` query (wiki miss → hybrid) | < 400ms | < 2s    |
| Ingest one signal (text)               | < 500ms | < 3s     |
| Ingest one signal (10-page PDF)        | < 3s    | < 15s    |
| Connector sync batch (1K items)        | < 30s   | < 2min   |
| Wiki curation (one page, incremental)  | < 10s   | < 60s    |
| Audit query (principal × time range)   | < 100ms | < 500ms  |

---

## Part 5 — The 12 build phases

Strict ordering. Each phase blocks the next via its acceptance criterion.

### Phase 0 — Extract & unify (DONE ✅)

- Extracted engine from `OptimalOS/engine/` into `~/Desktop/OptimalEngine/`.
- Absorbed three MIOSA subsystems (`miosa_knowledge`, `miosa_memory`, `miosa_signal`) into unified `OptimalEngine.*` namespace.
- 689 tests green. Private GitHub remote created (not pushed).

### Phase 1 — Schema & tenancy foundation

**Scope:**
- Add tables: `chunks`, `classifications`, `intents`, `assets`, `clusters`, `cluster_members`, `wiki_pages`, `citations`, `events`, `tenants`, `principals`, `groups`, `roles`, `role_grants`, `acls`, `connectors`, `connector_runs`, `retention_policies`, `legal_holds`, `audiences`.
- Every primary table gets `tenant_id` as first column; every index leads with it.
- `mix optimal.migrate` runnable.
- Backfill: every existing `contexts` row → `:document`-scale chunk + assigned to default tenant.
- Refactor `Store` GenServer for per-tenant SQLite routing.
- Add `OptimalEngine.Identity.Principal` + `OptimalEngine.Identity.ACL` primitives.
- Add `OptimalEngine.Audit.Event` + logger writing to `events`.

**Acceptance:** `mix test` green (689+ tests); `mix optimal.migrate` applies cleanly from current state; `mix optimal.stats` shows row counts for all new tables; `mix optimal.search` enforces principal-filtered WHERE clauses (added test).

**Modules touched / created:** ~18 new, 4 edited.

### Phase 3.5 — Workspace (organizational topology) ✅

**Why a .5 phase:** Alice's framing clarified that the engine isn't just a
pipeline — it's the second brain *of a company*. That means nodes, people-vs-
agents, skills, and memberships are first-class data, not implicit strings.
Phase 1 gave us tenants + principals + groups + roles. Phase 3.5 adds the
missing organizational-topology primitives so Phase 4's classifier can key off
"who owns this signal" and Phase 7's wiki curator can resolve audiences
properly.

**Scope:**
- Migrations 017–022: `nodes`, `node_members`, `skills`, `principal_skills`, tenant-first indexes, backfill distinct `contexts.node` values into `nodes` rows.
- `OptimalEngine.Workspace.Node` — CRUD, tree traversal (children, ancestors), list filtered by kind/status.
- `OptimalEngine.Workspace.NodeMember` — add/remove (owner/internal/external/observer), members_of, nodes_of, time-bounded (started_at/ended_at).
- `OptimalEngine.Workspace.Skill` — CRUD (technical/communication/strategic/domain/tool).
- `OptimalEngine.Workspace.PrincipalSkill` — grant/revoke at level (novice/intermediate/expert/lead), min_level filter, principals_with_skill returns humans + agents together.
- Top-level facade: `OptimalEngine.Workspace`.
- `mix optimal.workspace [--tenant X] [--nodes-only|--skills-only]`.
- Doc: `docs/architecture/WORKSPACE.md`.

**Acceptance:** 20 new tests; nodes form a tenant-scoped tree with correct
ancestor ordering; skill levels filter correctly; memberships are
time-bounded; default-tenant backfill runs. **Complete: 786/786 passing.**

### Phase 2 — Parser backends (format coverage)

**Scope:** implement `OptimalEngine.Pipeline.Parser` dispatch + 10 format backends in strict order:
1. `.md .txt .rst .adoc` (native)
2. `.yaml .yml .toml .json`
3. `.csv .tsv` (NimbleCSV)
4. `.html` (Floki)
5. source code (native / tree-sitter)
6. `.pdf` (pdftotext shell / pdf_extract)
7. `.docx .pptx .xlsx` (zip + OOXML)
8. `.png .jpg .jpeg .gif .webp` (tesseract OCR + asset kept)
9. `.mp3 .wav .m4a .ogg .flac` (whisper.cpp + asset kept)
10. `.mp4 .mov .webm` (ffmpeg → frames + audio → #8 + #9)

**Acceptance:** `mix optimal.parse <file>` returns `%ParsedDoc{text, structure, assets, modality}` for every format. Fixture test per format.

**Modules touched / created:** 11 new.

### Phase 3 — Decomposer

**Scope:** `OptimalEngine.Pipeline.Decomposer` — hierarchical chunking at 4 scales (`:document` / `:section` / `:paragraph` / `:chunk`) with parent_id links. Respects parser-reported structural boundaries.

**Acceptance:** a 5-page PDF produces 1 document + N section + M paragraph + K chunk-scale chunks, correct parent links. Reassembly by concat is byte-identical.

**Modules touched / created:** 1 new (Decomposer) + 1 refactored (Store to handle the `chunks` table).

### Phase 4 — Per-chunk classify + intent extract

**Scope:** run existing `Classifier` per chunk. Build `OptimalEngine.Pipeline.IntentExtractor` with heuristics-first + Ollama-augmented-if-available. 10-value intent enum: `:request_info :propose_decision :record_fact :express_concern :commit_action :reference :narrate :reflect :specify :measure`.

**Acceptance:** on 50-signal gold set, intent accuracy ≥ 80%. Every chunk ends up with `classifications` + `intents` row.

**Modules touched / created:** 1 new (IntentExtractor) + 1 edited (Classifier).

### Phase 5 — Embedder (multi-modal)

**Scope:**
- `OptimalEngine.Embed.Ollama.embed_text/1`, `.embed_image/1`.
- `OptimalEngine.Embed.Whisper` — whisper.cpp local HTTP client.
- `OptimalEngine.Pipeline.Embedder` — dispatches on modality.
- `embeddings` table per chunk, modality-tagged.

**Acceptance:** text-query-retrieves-image test: an image of "a pricing chart with $2K" embedded alongside a text chunk about "Ed's $2K pricing" both land in the retrieval top-5 for text query "pricing $2K".

**Modules touched / created:** 3 new (Whisper, Embedder, Embed.Provider).

### Phase 6 — Clusterer

**Scope:** `OptimalEngine.Pipeline.Clusterer` — HDBSCAN over `feature = 0.6·embedding + 0.2·entity_overlap + 0.15·intent_match + 0.05·node_affinity`. Theme auto-naming via Ollama over top-N chunks per cluster. Incremental add. `mix optimal.cluster.rebuild` for full rebuild.

**Acceptance:** on current corpus, ≥ 70% of chunks cluster coherently (human eval of 20 random clusters).

**Modules touched / created:** 1 new.

### Phase 7 — Wiki Layer (Tier 3)

**Scope:**
- `.wiki/SCHEMA.md` written.
- Page template defined (frontmatter + Summary + Open threads + Related + Incoming).
- `OptimalEngine.Wiki.Page` struct + (de)serialization.
- `OptimalEngine.Wiki.Directives` parser + renderer: `{{cite}} {{include}} {{expand}} {{search}} {{table}} {{trace}} {{recent}}` plus `[[wikilink]]`.
- `OptimalEngine.Wiki.Integrity` — citation + contradiction checker.
- `OptimalEngine.Wiki.Curator` — Ollama-driven rewriter.
- `OptimalEngine.Pipeline.CuratorTrigger` — Stage 9; on `store.chunk.indexed` events, compute affected wiki pages, enqueue curation.
- Audience-aware curation (one signal → N page variants, intersection-filtered).
- `mix optimal.wiki <view|edit|rebuild|verify> <slug-or-all>`.

**Acceptance:** ingesting a new signal about Alice updates `.wiki/accounts/acme.md` autonomously with new citations verified. Per-audience variants produced (sales / engineering / exec) with zero cross-leak (integrity test).

**Modules touched / created:** 8 new.

### Phase 8 — Scale-aware Deliver + Composer

**Scope:**
- `ContextAssembler` tries wiki first; falls through to hybrid retrieval only on wiki miss.
- Hybrid retrieval extended: permission-aware WHERE clauses, intent-match boost, cluster-sibling boost, temporal decay, scale preference (coarsest answering chunk).
- `OptimalEngine.Retrieval.Composer` — formats: `text / json / claude / openai / tool-result`.
- `mix optimal.rag <query>` wired to full pipeline.

**Acceptance:** `optimal rag "pricing with Ed"` returns wiki-sourced summary with inline citations on first call; zero retriever hits. On wiki miss, hybrid retrieval returns top-K within p99 < 2s.

**Modules touched / created:** 2 edited, 1 new (Composer).

### Phase 9 — Connectors (enterprise integrations)

**Scope:** `OptimalEngine.Connectors.Connector` behaviour + 14 implementations:

| # | Connector              | Mode                         |
|---|-----------------------|------------------------------|
| 1 | Filesystem             | filesystem watch             |
| 2 | Slack                  | events API + backfill        |
| 3 | Google Drive           | Changes API                  |
| 4 | Google Docs            | via Drive + Docs API         |
| 5 | Gmail                  | history API                  |
| 6 | Microsoft 365          | Graph API (Outlook + OneDrive + SharePoint) |
| 7 | Notion                 | polling + webhook            |
| 8 | Jira                   | webhook + polling            |
| 9 | Linear                 | webhook + GraphQL            |
| 10| GitHub                 | webhook + GraphQL            |
| 11| Zoom                   | transcript webhook           |
| 12| Confluence             | REST + CQL                   |
| 13| Salesforce / HubSpot   | Bulk API + streaming         |
| 14| Intercom / Zendesk     | webhook + polling            |
| 15| Custom HTTP webhook    | push                         |

Each connector: idempotent, resumable, honors rate limits, reports cursor state to `connector_runs`, emits structured errors for the Queue view.

**Acceptance:** each connector syncs a fixture dataset end-to-end; cursor state survives restart; deliberate errors surface as Queue items.

**Modules touched / created:** 15 new.

### Phase 10 — Production hardening

**Scope:**
- Structured JSON logging option.
- Telemetry spans for every pipeline stage + retrieval.
- Performance targets (Part 4) met on a 10M-signal benchmark dataset.
- Backup / restore scripts.
- Point-in-time recovery for SQLite.
- `mix release optimal` produces an OTP release tarball.
- Burrito single-binary build for cross-platform distribution (optional v1).

**Acceptance:** release tarball runs on a clean macOS or Linux VM with no Erlang installed and passes smoke tests. Performance targets hit on benchmark set.

**Modules touched / created:** telemetry sprinkled everywhere + release config.

### Phase 11 — Compliance

**Scope:**
- SOC 2 Type II audit trail surfaces (events → SIEM JSON export).
- GDPR right-to-be-forgotten flow: propagate delete through Tier 2 derivatives + queue wiki re-curation to strip citations.
- HIPAA-ready deployment mode (encrypted at rest + in transit + audit attestation).
- Admin console views (in UI) for access logs, retention rules, legal holds.

**Acceptance:** compliance posture documented + reviewable; all audit queries hit p99 < 500ms; delete flows verified end-to-end.

**Modules touched / created:** audit exporter, retention engine, delete propagator.

### Phase 12 — UI build (parallel with Phases 7–10)

**Scope:** desktop app per `docs/architecture/UI.md`. Seven views (Brief / Source / Probe / Atlas / Flow / Audit / Queue) + Ask bar + Sweep drawer + Clip extension. Tauri 2 + SvelteKit + Tailwind v4 + Foundation tokens + bits-ui + Milkdown + Sigma.js.

**Sub-phases:**

| UI Phase | Scope                                                      |
|----------|------------------------------------------------------------|
| U1       | Tauri + SvelteKit scaffold. Rail + tree + main panel shell.|
| U2       | Brief view (read-only wiki pages + directives resolve).    |
| U3       | Source view (modality-aware renderers).                    |
| U4       | Probe view (hybrid retrieval + Lens toggle).               |
| U5       | Atlas view (Sigma.js + Louvain + hubs + triangles).        |
| U6       | Flow + Audit views.                                        |
| U7       | Queue + Sweep + Clip extension.                            |
| U8       | Wiki editing (Milkdown) for human curator overrides.       |
| U9       | Packaging: Tauri bundle for macOS + Linux + Windows.       |

**Acceptance:** all seven views functional against live engine; keyboard shortcuts work; Tauri bundles produced for macOS + Linux + Windows.

### Phase 13 — Push

**Scope:** version `v0.1.0` tagged. Final docs pass. `git push -u origin main`.

**Acceptance:** Alice's explicit go.

---

## Part 6 — Module restructure plan

Current `lib/optimal_engine/` is 38 flat files. Target (from Part 2's architecture):

```
lib/optimal_engine/
├── optimal_engine.ex           # facade (defdelegate everything)
├── application.ex              # supervision tree
├── cli.ex
│
├── pipeline/                   # Stages 1–9
│   ├── intake.ex
│   ├── parser.ex + parser/{markdown,yaml,json,csv,html,code,pdf,office,image,audio,video}.ex
│   ├── decomposer.ex
│   ├── classifier.ex
│   ├── intent_extractor.ex
│   ├── embedder.ex
│   ├── router.ex
│   ├── indexer.ex
│   ├── clusterer.ex
│   └── curator_trigger.ex
│
├── store/
│   ├── store.ex + schema.ex + migrations.ex + fts.ex + vectors.ex + hot_cache.ex + tenancy.ex
│
├── retrieval/
│   ├── intent_analyzer.ex + search.ex + context_assembler.ex + composer.ex + l0_cache.ex
│
├── wiki/                       # Tier 3
│   ├── wiki.ex + page.ex + directives.ex + schema.ex + integrity.ex
│   └── curator/{curator,prompt,jobs}.ex
│
├── knowledge/  memory/  signal/  # absorbed subsystems, unchanged paths
│
├── graph/
│   ├── analyzer.ex + graph.ex + reflector.ex
│
├── insight/
│   ├── remember.ex + rethink.ex + reweave.ex + verify.ex + health.ex + simulate.ex
│
├── embed/
│   ├── ollama.ex + whisper.ex + provider.ex
│
├── tenancy/
│   ├── tenant.ex + scope.ex + config.ex
│
├── identity/
│   ├── principal.ex + group.ex + role.ex + acl.ex + saml.ex + oidc.ex + scim.ex
│
├── connectors/
│   ├── connector.ex + 14 per-source implementations
│
├── audit/
│   ├── event.ex + logger.ex + siem.ex
│
├── api/
│   ├── router.ex + auth_plug.ex + handlers/{rag,search,ingest,wiki,graph,health,stats,events}.ex
│
└── (flat primitives): context.ex  uri.ex  session.ex  session_compressor.ex  topology.ex  spec/
```

### Namespace convention

`OptimalEngine` is a thin facade using `defdelegate`. Everything else is
accessed via its folder-derived namespace (e.g.,
`OptimalEngine.Pipeline.Intake`, `OptimalEngine.Retrieval.Search`).

### Migration steps

1. Create empty target dirs.
2. `git mv` files into new locations per the moves table.
3. Scripted `sed` to rewrite `defmodule` + all call sites.
4. Split `store.ex` into `store/{store,schema,migrations,fts,hot_cache}.ex`.
5. Write `optimal_engine.ex` top-level facade.
6. `mix format && mix compile --warnings-as-errors && mix test`.
7. Grep + audit + delete dead modules: `MCTS`, `MonteCarlo`, `SemanticProcessor`, `MemoryExtractor`, `CortexFeed`.
8. Commit: "Restructure lib/ to match canonical architecture."

### Dead modules to delete

| Module | Why |
|---|---|
| `OptimalEngine.MCTS` | Half-built tree search, no callers. |
| `OptimalEngine.MonteCarlo` | Early simulator experiment. |
| `OptimalEngine.SemanticProcessor` | Purpose unclear; likely absorbed into Classifier. |
| `OptimalEngine.MemoryExtractor` | Absorbed into Memory.Episodic. |
| `OptimalEngine.CortexFeed` | Shim redundant after absorption. |

The restructure is **Phase 0.5** — happens before Phase 1. Confirm before each phase touches code that hasn't been relocated.

---

## Part 7 — Decisions locked (do not revisit unless explicitly re-opened)

### Engine tech stack

- **Elixir ~> 1.17 / OTP 26+** (unchanged)
- **SQLite + FTS5** for primary storage (no external vector DB; switch only if >50K vectors per tenant becomes a bottleneck)
- **Ollama** for local LLM calls (embeddings + generation) — OSS-first, OpenAI as one adapter later
- **`nomic-embed-text-v1.5` + `nomic-embed-vision-v1.5` + `whisper.cpp`** for multi-modal aligned 768-dim embeddings
- **exqlite** NIF (built-in to Phase 0)

### UI tech stack

- **Tauri 2** (Rust desktop shell)
- **SvelteKit** (matches MIOSA/BusinessOS)
- **Tailwind CSS v4 + Foundation tokens** (Alice's existing design system)
- **bits-ui + melt-ui** (Svelte-native headless primitives)
- **Milkdown** (markdown editor)
- **Sigma.js + graphology** (graph rendering)

### Explicitly rejected

- React, shadcn/ui, LanceDB, Electron, OpenAI-only SDKs, Google multi-modal embeddings (OSS-first until explicitly opened)

### Naming locked

- Stages: Intake / Parse / Decompose / Classify / Embed / Route / Store / Cluster / Curate
- Tiers: Raw Sources / Derivatives / Wiki
- UI views: Brief / Source / Probe / Atlas / Flow / Audit / Queue
- UI secondary: Ask / Sweep / Clip / Lens / Scaffold
- Engine terms: Signal / Scale / Intent / Modality / Cluster / Citation / Directive / Audience / Node / Entity / Principal / Tenant

### Explicitly not in v0.1

- Inline PDF annotation (Phase 11+)
- WebDAV / S3 sync (Phase 11+)
- Plugin system
- Mobile app
- Multi-language UI (i18n)
- Cloud LLM providers (Claude API, OpenAI, Google) as the default

### Git / release discipline

- No `Co-Authored-By: Claude` lines on commits
- No push to the remote until Phase 13 explicit go
- Commits use new commits, never amend
- No force-push to main ever
- All work in feature branches + merged via PR once CI is in place

---

## Part 8 — Open questions (genuine TBDs)

| # | Question | Blocking? |
|---|---|---|
| 1 | License: private, MIT, BUSL, AGPL — Alice's call | Phase 13 |
| 2 | Hosted offering strategy: do we offer SaaS, or strictly self-hosted with commercial support? | Phase 9 (impacts connector OAuth story) |
| 3 | Pricing model for a hosted tier (per-seat, per-signal, per-tenant, storage-based) | Post Phase 13 |
| 4 | Which agent runtime gets first-class integration beyond the CLI: Claude Agent SDK, LangChain, MCP server? | Phase 10 |
| 5 | Do we compete directly with Glean/Dust in sales, or anchor on MIOSA + expand outward? | Strategic; Alice to decide |
| 6 | Cloud-embedding backstop (Cohere / Voyage) for tenants without local Ollama capacity — in or out? | Phase 5 |
| 7 | Obsidian vault compatibility — nice-to-have or distraction? | Phase 7 |
| 8 | Authority model for multi-tenant deployments: Org-level superadmin vs per-principal-delegated? | Phase 1 |

---

## Part 9 — Today's next action

1. Clean up the task list — delete the old scattered tasks, create one task per phase.
2. Execute **Phase 0.5 (restructure)** — `lib/` reorganization per Part 6.
3. Execute **Phase 1 (schema + tenancy)** — all new tables, migration, principal + ACL primitives, audit logger.
4. Report at each phase completion with acceptance criteria verified.

---

## Appendix A — Reference docs (the deep context)

- `docs/architecture/ARCHITECTURE.md` — canonical 9-stage pipeline, data contracts, module tree, retrieval flow, invariants
- `docs/architecture/ENTERPRISE.md` — tenancy, identity, ACLs, connectors, retention, audit, performance targets
- `docs/architecture/WIKI-LAYER.md` — Tier-3 deep dive: directives, curator, integrity, schema governance
- `docs/architecture/UI.md` — desktop app spec: seven views, Ask / Sweep / Clip / Lens, tech stack
- `docs/architecture/00-overview.md` through `07-governance.md` — the original 7-layer system architecture
- `docs/architecture/FULL-SYSTEM-ARCHITECTURE.md` — end-to-end walkthrough
- `docs/concepts/signal-theory.md` — `S=(M,G,T,F,W)` + 4 constraints + 6 principles + 11 failure modes
- `docs/concepts/methodology.md`, `three-spaces.md`, `failure-modes.md`, `infinite-context-framework.md` — theoretical foundation

## Appendix B — Competitive intelligence

Source: `OptimalOS/miosa-sandbox/.canopy/reference/competitors/` — detailed per-competitor analysis of:
- **Application-layer rivals:** Glean, Dust, NotebookLM, Limitless, Pieces
- **Context databases:** OpenViking, MemOS
- **Agent memory:** Mem0, Letta, Zep/Graphiti, Cognee, HydraDB, LangMem
- **Retrieval frameworks:** LlamaIndex, LangChain, GraphRAG, RAGFlow
- **Vector DBs:** Pinecone, Weaviate, Qdrant, ChromaDB, Milvus, pgvector
- **PKM:** Obsidian, Notion, Logseq, Roam
- **Protocols:** MCP
- **Runtime:** Shannon (Kocoro)
- **Novel:** Nuggets (HRR memory)

Most-dangerous single competitor: **OpenViking (ByteDance)** — same filesystem paradigm, same L0/L1/L2 tiered loading, massive engineering resources. Lacks our classification / composition / receiver-modeling layers. If they add them, they are the top threat. Watch their roadmap.

## Appendix C — Theoretical foundation

- **Signal Theory** (R. Luna, Feb 2026) — `S=(M,G,T,F,W)` with 4 constraints (Shannon / Ashby / Beer / Wiener)
- **Three-spaces model** — input / signal / persistence separation
- **Tiered disclosure** — L0 (~100 tok abstract) / L1 (~2K tok overview) / L2 (full); bandwidth matched to receiver
- **Triple-loop learning** — single (did it work?) / double (was it right?) / triple (are we asking the right questions?)
- **VSM-inspired governance** — autonomy levels per agent, identity preservation across sessions

## Appendix D — Glossary (use these terms, don't invent new ones)

| Term | Meaning |
|---|---|
| Signal | a unit of meaning at any scale, classified |
| Scale | `:document` \| `:section` \| `:paragraph` \| `:chunk` |
| Intent | what the signal is trying to accomplish (10-value enum) |
| Modality | `:text` \| `:image` \| `:audio` \| `:video` \| `:code` \| `:data` \| `:mixed` |
| Cluster | theme-group of chunks across documents |
| Citation | URI pointer from wiki claim to source chunk |
| Directive | executable `{{verb: arg}}` in wiki pages |
| Audience | a role set that defines who a wiki page is curated for |
| Node | organizational folder (e.g., `04-ai-masters`) |
| Entity | a named thing mentioned in signals (e.g., "Alice") |
| Principal | an identity the engine knows about (human, agent, service account) |
| Tenant | an isolated org with its own DB + wiki |
| Tier 1 / 2 / 3 | Raw Sources / Derivatives / Wiki |
| Stage 1–9 | Intake → Parse → Decompose → Classify → Embed → Route → Store → Cluster → Curate |
| Ask / Sweep / Clip / Lens / Scaffold | UI secondary surfaces |
| Brief / Source / Probe / Atlas / Flow / Audit / Queue | UI primary views |

---

*End of master plan. Keep this doc updated with every architectural decision, phase completion, and open question resolution. When in doubt, this is the contract.*
