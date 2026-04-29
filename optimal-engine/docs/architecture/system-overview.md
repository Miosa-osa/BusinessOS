---
signal:
  mode: linguistic
  genre: spec
  type: inform
  format: markdown
  structure: architecture_overview
  audience: [roberto, javaris, pedram]
  intent: "Full system architecture as concentric layers — from raw files to conversation interface"
  sn_ratio: 1.0
---

# OptimalOS System Architecture

OptimalOS is a signal-classified context operating system. Every piece of information that enters it is classified across 5 dimensions, routed to the right organizational nodes, stored with pre-computed summaries, and retrieved using hybrid scoring. The interface is a conversation. The engine is Elixir/OTP.

The architecture is best understood as concentric layers — each one wraps the previous, each one adds intelligence.

---

## The Onion

```
┌─────────────────────────────────────────────────────────────────┐
│  Layer 9 — THE INTERFACE (Claude Code TUI + CLAUDE.md)          │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Layer 8 — SESSION & FEEDBACK (single/double/triple loop) │  │
│  │  ┌─────────────────────────────────────────────────────┐  │  │
│  │  │  Layer 7 — COMPOSITION & RECEIVER MODELING          │  │  │
│  │  │  ┌───────────────────────────────────────────────┐  │  │  │
│  │  │  │  Layer 6 — INTAKE PIPELINE (classify→route)   │  │  │  │
│  │  │  │  ┌─────────────────────────────────────────┐  │  │  │  │
│  │  │  │  │  Layer 5 — ROUTING & TOPOLOGY            │  │  │  │  │
│  │  │  │  │  ┌───────────────────────────────────┐   │  │  │  │  │
│  │  │  │  │  │  Layer 4 — SEARCH & RETRIEVAL      │   │  │  │  │  │
│  │  │  │  │  │  ┌─────────────────────────────┐   │   │  │  │  │  │
│  │  │  │  │  │  │  Layer 3 — TIERED LOADING    │   │   │  │  │  │  │
│  │  │  │  │  │  │  ┌───────────────────────┐   │   │   │  │  │  │  │
│  │  │  │  │  │  │  │  Layer 2 — STORAGE     │   │   │   │  │  │  │  │
│  │  │  │  │  │  │  │  ┌─────────────────┐   │   │   │   │  │  │  │  │
│  │  │  │  │  │  │  │  │  Layer 1 — DATA  │   │   │   │   │  │  │  │  │
│  │  │  │  │  │  │  │  │  (The Core)      │   │   │   │   │  │  │  │  │
│  │  │  │  │  │  │  │  └─────────────────┘   │   │   │   │  │  │  │  │
│  │  │  │  │  │  │  └───────────────────────┘   │   │   │  │  │  │  │
│  │  │  │  │  │  └─────────────────────────────┘   │   │  │  │  │  │
│  │  │  │  │  └───────────────────────────────────┘   │  │  │  │  │
│  │  │  │  └─────────────────────────────────────────┘  │  │  │  │
│  │  │  └───────────────────────────────────────────────┘  │  │  │
│  │  └─────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

Concentric, not stacked. Layer 1 is the core. Each outer layer depends on everything inside it. You can pull out any outer layer and the inner layers still function.

---

## Layer 1 — The Data (Core)

**What it is:** Markdown files with YAML frontmatter, organized in 12 numbered folders.

**Structure:**
```
OptimalOS/
├── 01-roberto/          # Personal context, tasks, goals
├── 02-miosa/            # MIOSA LLC platform
├── 03-lunivate/         # Lunivate LLC — agency entity
├── 04-ai-masters/       # AI Masters course
├── 05-os-architect/     # OS Architect channel + research
├── 06-agency-accelerants/ # AA network + ClinicIQ
├── 07-accelerants-community/ # AA school group
├── 08-content-creators/ # Mosaic Effect + ContentOS
├── 09-new-stuff/        # Unrouted inbox
├── 10-team/             # People registry
├── 11-money-revenue/    # Cross-cutting financial node
├── 12-os-accelerator/   # OS Accelerator course
└── docs/                # System documentation
```

Each numbered folder contains:
- `context.md` — Persistent facts. Never overwritten. Grows over time.
- `signal.md` — Weekly status. Replaced each Monday.
- `signals/` — Individual classified signal files.

**Every file is a signal.** Every `.md` file in the system carries YAML frontmatter that classifies it:

```yaml
---
id: sig_20260318_001
signal:
  mode: linguistic
  genre: decision-log
  type: decide
  format: markdown
  structure: decision_log_skeleton
node: 04-ai-masters
entities: [ed-honour, robert-potter, roberto]
created_at: 2026-03-18T01:00:00Z
valid_until: null
sn_ratio: 0.95
tiers:
  l0: "AI Masters pricing set at $99/mo community, $8-9K/yr premium"
  l1: "Ed and Alice agreed on two-tier pricing..."
---
```

**Source of truth is always the filesystem.** The database is an index of the files, not a replacement for them.

---

## Layer 2 — Storage and Index

**What it is:** SQLite + FTS5 search index at `.system/index.db`. ETS hot cache for fast reads.

**Three storage layers:**

```
┌─────────────────────────────────────────┐
│  Markdown Files (source of truth)        │
│  Git-versionable. Human-readable.        │
│  Every file = classified signal.         │
└──────────────────┬──────────────────────┘
                   │ indexed by
┌──────────────────▼──────────────────────┐
│  SQLite + FTS5 (fast search)             │
│  BM25 full-text. Temporal queries.       │
│  Pre-computed tier content.              │
└──────────────────┬──────────────────────┘
                   │ cached in
┌──────────────────▼──────────────────────┐
│  ETS Hot Cache (fast reads)              │
│  Frequently accessed signals.            │
│  Invalidated on index rebuild.           │
└─────────────────────────────────────────┘
```

**Key tables:**
- `contexts` — Every indexed signal with all 5 signal dimensions, tier content, temporal fields
- `signals_fts` — FTS5 virtual table for BM25 full-text search
- `entities` — People, orgs, tools extracted from signal content
- `edges` — Relationships between entities (manages, works_on, depends_on)
- `decisions` — Append-only audit trail of every system action

**The index is rebuilt with `mix optimal.index`.** The source files are always the canonical record.

---

## Layer 3 — Tiered Loading (L0/L1/L2)

**What it is:** Progressive disclosure of context. Pre-computed at index time. Stored in signal frontmatter.

**The four tiers:**

| Tier | Name | Budget | Content | When Loaded |
|------|------|--------|---------|-------------|
| L0 | HOT | ~2K tokens | Headlines. Always present. | Every session |
| L1 | WARM | ~10K tokens | Summaries of active context | When task scope is identified |
| L2 | COLD | ~50K tokens | Full detail | On explicit drill-down |
| L3 | ARCHIVE | fragments | Historical, superseded | Never auto-loaded — search only |

**L0 cache** lives at `.system/cache/l0.md`. It contains system identity, active priorities, recent decisions, and key people. Regenerated automatically when decisions change or weekly priorities update.

**Tier summaries are pre-computed** in each file's YAML frontmatter (`tiers.l0`, `tiers.l1`). The engine loads headlines without reading file bodies. At 1,000 signals, that means reading ~200K tokens of frontmatter to surface the right ~2K token context window.

**Promotion and demotion are automatic:**
- Signal accessed 3+ consecutive sessions for the same scope → promoted L2 → L1
- Signal not accessed in 30 days → demoted L1 → L2
- Every promotion/demotion is logged to the `decisions` audit table

---

## Layer 4 — Search and Retrieval

**What it is:** Hybrid scoring engine that runs multiple search modes in parallel and fuses results.

**Search modes:**

```
Query: "What did Ed say about pricing?"
             │
             ▼
┌────────────────────────────────────────────┐
│           PARALLEL SEARCH                   │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  BM25    │  │ Temporal  │  │   S/N    │  │
│  │  (FTS5)  │  │ (decay)   │  │  ratio   │  │
│  │  0.60    │  │  0.30     │  │  0.10    │  │
│  └────┬─────┘  └────┬──────┘  └────┬─────┘  │
│       │              │              │         │
│  ┌────▼──────────────▼──────────────▼──────┐  │
│  │    Hybrid Score (weighted sum)          │  │
│  │    score = bm25×0.6 + decay×0.3 + sn×0.1│  │
│  └─────────────────┬───────────────────────┘  │
└───────────────────┼────────────────────────┘
                    ▼
          Tier-Aware Assembly
          (fill L0 → L1 → L2 within budget)
```

**Temporal half-lives by genre** — older signals decay faster or slower depending on type:

| Genre | Half-Life | Rationale |
|-------|-----------|-----------|
| message, transcript | 7 days | Ephemeral |
| standup, status-report | 14 days | Weekly cycle |
| plan | 30 days | Monthly planning |
| note | 60 days | Processing buffer |
| spec, prd | 180 days | Build artifacts |
| decision-log, adr | 365 days | Long-lived |
| pattern | 730 days | Institutional knowledge |
| entity | no decay | People don't expire |

**Performance targets:** BM25 <10ms, temporal <5ms, hybrid assembly <80ms end-to-end.

**Every retrieval produces an observable trace** showing which signals were loaded, why they ranked where they did, and how many tokens were used. Debuggable by design.

---

## Layer 5 — Routing and Topology

**What it is:** The organizational map. Nodes, people, routing rules — all configuration-driven from `topology.yaml`.

**The topology defines:**
- 12 nodes with their types and folder mappings
- People (endpoints) with genre competence and preferred channels
- Routing rules: keyword patterns → destination node(s)
- Cross-cutting rules: financial signals ALWAYS also route to `nodes/11-money-revenue`

**Routing is signal-property-based, not session-based.** When "Dan closed a deal for $1,500" enters the system, the router evaluates:
1. Keywords match (`deal`, `$`) → fires financial routing rule → routes to `nodes/06-agency-accelerants` AND `nodes/11-money-revenue`
2. Entity mention (`Dan`) → routes to `nodes/10-team` for relationship update

One signal, three destinations. Fully automatic.

**No routing logic lives in code.** All rules live in `topology.yaml`. Alice adds a new node or person, edits the YAML, and the router picks it up on next run.

---

## Layer 6 — Intake Pipeline

**What it is:** The 5-step pipeline that turns raw text into classified, indexed, routed signals.

```
Raw text (voice, message, brain dump)
    │
    ▼
┌──────────┐   ┌───────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ CLASSIFY  │──▶│  COMPOSE  │──▶│ SKELETON │──▶│  WRITE   │──▶│  INDEX   │
│ S=(M,G,   │   │ Extract   │   │ Validate │   │ Markdown │   │ SQLite + │
│  T,F,W)   │   │ entities  │   │ genre    │   │ file to  │   │ FTS5 +   │
│ S/N ratio │   │ triples   │   │ skeleton │   │ right    │   │ route    │
│ measured  │   │ tiers     │   │ present? │   │ node(s)  │   │ complete │
└──────────┘   └───────────┘   └──────────┘   └──────────┘   └──────────┘
```

**Genre skeletons** ensure structural consistency. 10 built-in skeletons, extensible:

| Skeleton | Genre | Required Sections |
|----------|-------|-------------------|
| `transcript` | Meeting/call record | Participants, Key Points, Decisions, Action Items, Open Questions |
| `brief` | Sales/exec comms | Objective, Key Messages, CTA, Supporting Materials |
| `spec` | Developer docs | Goal, Requirements, Constraints, Architecture, Acceptance Criteria |
| `plan` | Personal planning | Objective, Non-Negotiables, Time Blocks, Dependencies, Success Criteria |
| `note` | Quick capture | Context, Content, Route |
| `decision-log` | Permanent decisions | Date, Decision, Rationale, Alternatives Considered, Made By |
| `standup` | Status update | Done, Doing, Blockers, Next |
| `review` | Feedback loops | What Happened, What Worked, What Didn't, Adjustments |
| `report` | Formal reporting | Executive Summary, Data, Analysis, Recommendations |
| `pitch` | Sales/proposals | Hook, Problem, Solution, Proof, Ask |

**The full catalogue has 143 genres across 18 categories.** See `docs/taxonomy/genres.md`.

---

## Layer 7 — Composition and Receiver Modeling

**What it is:** The output layer. Shapes context and documents FOR a specific receiver in their preferred genre.

**Same data. Different output.**

When Alice asks "create something for Robert about AI Masters pricing":
1. System identifies receiver: `robert-potter`
2. Looks up genre competence: `[brief]` — Robert receives briefs only
3. Pulls relevant context from `04-ai-masters/`
4. Composes in brief skeleton: Objective → Key Messages → CTA → Supporting Materials
5. Outputs a brief. Not a spec. Not a report. A brief.

When Alice asks "create a spec for Frank for the dashboard feature":
1. System identifies receiver: `nejd`
2. Looks up notes: "explicit constraints required — over-engineers on open-ended"
3. Pulls relevant context from `02-miosa/`
4. Composes spec with a mandatory Constraints section listing hard limits
5. Outputs a constrained spec.

**No equivalent exists in any competitor.** OpenViking, Mem0, Letta, Zep — all produce generic output. OptimalOS encodes FOR the receiver.

---

## Layer 8 — Session and Feedback

**What it is:** Three feedback loops that keep the system calibrated over time.

| Loop | Question | Frequency | Output |
|------|----------|-----------|--------|
| Single | Did the signal arrive and get decoded? | Per signal | Confirmation, retry, escalate |
| Double | Were these the right priorities this week? | Weekly (Friday) | Updated week-plan, alignment scores |
| Triple | Are we asking the right questions at all? | Monthly | Node structure review, routing rule updates |

**Session lifecycle:**
1. Session starts → L0 cache loaded (~2K tokens always present)
2. Work happens → signals classified and routed in real time
3. Session ends → decisions extracted → context.md updated → next L0 cache queued

**The SICA learning loop** (Observe → Reflect → Propose → Test → Integrate):
- Agent notices a pattern (e.g., Alice always adds context when a brief is too short)
- Reflects: "Bob briefs should include a metrics section by default"
- Proposes an update to the `brief` skeleton
- Alice confirms or rejects
- On confirm: skeleton updated in templates

---

## Layer 9 — The Interface (Outermost)

**What it is:** Claude Code TUI. Alice talks. The system processes.

**The conversation IS the interface.** There is no separate UI. No web app. No dashboard. The terminal window where Alice talks to Claude Code is the operating system's control surface.

**CLAUDE.md is the engine config.** Claude reads it every session. It contains:
- The routing table (which keywords go where)
- The people table (who gets what genre)
- The workflow definitions (how to process a brain dump, how to run a Friday review)
- The rules (financial data always to node 11, Bob always gets briefs)

**Mix commands** give direct engine access when needed. Claude runs them; Alice doesn't have to.

---

## vs Competitors

What makes OptimalOS different is not any single feature — it's the combination of all layers working together.

| Layer | Capability | OpenViking | Mem0 | Letta | Zep/Graphiti | Dust | OptimalOS |
|-------|-----------|-----------|------|-------|--------------|------|-----------|
| 1 | Filesystem hierarchy | Yes | No | No | No | Partial | Yes |
| 1 | Signal classification (5 dims) | No | No | No | No | No | Yes |
| 2 | SQLite + FTS5 index | No | No | No | No | No | Yes |
| 3 | Tiered loading (L0/L1/L2) | Yes (3 tiers) | No | Yes | No | No | Yes (4 tiers + budgets) |
| 3 | Token budget control | No | No | No | No | No | Yes |
| 4 | Temporal decay by genre | No | No | No | Partial | No | Yes |
| 4 | Hybrid search | No | No | No | Yes (BM25+vec+graph) | No | Yes |
| 4 | Observable retrieval trace | Yes | No | No | No | No | Yes |
| 5 | Org topology + routing | No | No | No | No | No | Yes |
| 5 | Cross-node routing | No | No | No | No | No | Yes |
| 6 | 143-genre catalogue | No | No | No | No | No | Yes |
| 6 | Genre skeleton validation | No | No | No | No | No | Yes |
| 7 | Receiver modeling | No | No | No | No | No | Yes |
| 7 | Genre re-encoding for receiver | No | No | No | No | No | Yes |
| 8 | Triple-loop feedback | No (single) | No | No | No | No | Yes |
| 8 | SICA self-improvement | No | No | No | No | No | Yes |

**The fundamental difference:** every competitor treats context as a retrieval problem (how do I find the right chunk of text?). OptimalOS treats context as a communication problem (how do I send the right Signal to the right receiver in the right genre at the right time?).

---

## Signal Flow (End to End)

**Intake:**
```
Alice: "Customer called — we're pushing filming to next week"
    │
    ▼ Classify
    mode=linguistic, genre=transcript, type=inform
    entities=[ed-honour], node=04-ai-masters
    sn_ratio=0.8
    │
    ▼ Route
    04-ai-masters/signal.md (primary)
    10-team/context.md (entity update for ed-honour)
    │
    ▼ Write
    04-ai-masters/signals/2026-03-18-ed-filming-delay.md
    │
    ▼ Index
    SQLite contexts table + FTS5 + L0 cache invalidated
```

**Retrieval:**
```
Alice: "What's the status on Ed's filming?"
    │
    ▼ Search
    BM25: "ed filming" → 3 results
    Temporal: recent signals in 04-ai-masters → 2 results
    Hybrid: ranked list
    │
    ▼ Assemble
    L0: "Ed filming delayed — new date TBD" (20 tokens)
    L1: "Alice pushed filming session from March 20 to next week..." (50 tokens)
    │
    ▼ Respond
    Within 80ms from query to answer
```

---

## See Also

- [Context System Spec](context-system.md) — Deep technical detail on storage, retrieval, and evolution
- [ADR-002: Optimal Context Engine Architecture](ADR-002-optimal-context-engine-architecture.md) — Decision record for Elixir package structure
- [Engine Package Spec](optimal-engine-package-spec.md) — Module-by-module specification
- [Layer 1: Network](01-network.md) through [Layer 7: Governance](07-governance.md) — Original 7-layer reference architecture
- [Genre Catalogue](../taxonomy/genres.md) — All 143 genres across 18 categories
- [SOP](../SOP.md) — How to operate this system day-to-day
