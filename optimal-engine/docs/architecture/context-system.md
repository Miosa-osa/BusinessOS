---
signal:
  mode: linguistic
  genre: spec
  type: inform
  format: markdown
  structure: spec_template
  intent: "Technical spec for how the Optimal Context System stores, loads, retrieves, and evolves context"
---

# Context System — Technical Architecture

> How context is organized, stored, loaded, retrieved, and evolved.
> This is the ENGINE. Signal Theory is the THEORY. This is the IMPLEMENTATION.

## Why This Exists — And Why It's Not OpenViking

OpenViking is the closest system to what we're building. They got the filesystem paradigm right.
Credit where it's due. But there's a FUNDAMENTAL difference:

**OpenViking treats context as FILES. We treat context as SIGNALS.**

A file is dead. It sits in a directory. You find it by name or search. That's it.

A signal is alive. It has:
- **Identity** — WHAT it is (classified across 5 dimensions)
- **Intent** — what it's trying to DO (direct, inform, commit, decide, express)
- **Audience** — WHO should receive it, in WHAT form
- **Temporality** — WHEN it's valid, what it REPLACED
- **Depth** — pre-computed summaries at every tier (headline → summary → full)
- **Relationships** — what ENTITIES and NODES it connects to
- **Routing** — WHERE it should go automatically

When OpenViking stores "we decided to use $99 pricing," it's a memory file in a folder.
When we store the same thing, the system KNOWS:
- It's a **decision** (type=decide, genre=decision-log)
- It affects **AI Masters** and **money-revenue** nodes (auto-routed to both)
- It involves **Ed, Alice, Bob** (entities extracted)
- **Bob** should get a **brief** about it, not the raw decision log (receiver modeling)
- The L0 headline is 10 words, L1 summary is 50 words, without reading the file
- It **supersedes** the old pricing discussion (temporal chain)
- The OWL reasoner can infer that anything affecting AI Masters pricing affects Ed's revenue targets

That's not "files with extra metadata." That's a fundamentally different paradigm.

### The 5 Things We Have That Nobody Else Does

| # | Innovation | What It Means | OpenViking Equivalent |
|---|-----------|--------------|----------------------|
| 1 | **Signal Classification** — every context item classified on 5 dimensions S=(M,G,T,F,W) | System understands WHAT something IS, not just WHERE it lives. Filter by genre, type, mode. Retrieval is 10x smarter. | Nothing. Files are untyped. |
| 2 | **Receiver Modeling** — output shaped FOR the specific human/agent receiving it | Same data produces a brief for a salesperson, a spec for an engineer, an L0 for a busy exec. Context isn't just retrieved — it's ENCODED for the receiver. | Nothing. All output is generic. |
| 3 | **OWL Reasoning** — 16 forward-chaining rules that auto-derive facts | "Alice manages MIOSA" + "Erin works on MIOSA" → system INFERS "Alice indirectly manages Erin" without anyone writing that. Relationships compound automatically. | Nothing. No inference at all. |
| 4 | **Genre Composition** — 143 genre skeletons with required/optional/auto sections | Agent doesn't just "write a response." It composes in the CORRECT genre with the CORRECT skeleton. A postmortem has a root cause section. A brief has a CTA. Structure is imposed, not hoped for. | Nothing. Raw text output. |
| 5 | **Cross-Node Routing** — signals automatically flow to every relevant node | "Customer called about AI Masters pricing" → routes to `ai-masters/` AND `money-revenue/` AND updates Ed's entity profile. One signal, multiple destinations. Information doesn't get siloed. | Nothing. Single agent, single directory. |

### What OpenViking Got RIGHT (That We Keep)

- Filesystem paradigm (directories, not databases, as primary interface)
- Tiered loading (L0/L1/L2 — don't load what you don't need)
- Recursive retrieval (navigate hierarchy → search locally → expand outward)
- Observable retrieval (trace WHY something was loaded — debuggable)
- Self-evolution (agent improves its own context over time)

We keep ALL of these. They're proven. We don't reinvent what works.

---

## 1. Context Organization — The Filesystem

## The Filesystem — Side by Side

### OpenViking (theirs)

```
viking://
├── resources/                  # Shared docs, reference material
│   ├── company-wiki.md
│   └── api-docs.md
├── user/
│   └── memories/               # Flat list of user memories
│       ├── memory-001.md
│       └── memory-002.md
├── agent/
│   ├── skills/                 # Agent capabilities (flat)
│   │   ├── skill-001.md
│   │   └── skill-002.md
│   └── memories/               # Agent memories (flat)
│       ├── memory-001.md
│       └── memory-002.md
└── session/                    # Conversation logs
    └── session-001.jsonl
```

Simple. Clean. Works. But every file is just... a file. No classification, no relationships,
no routing, no tier summaries. The agent has to READ every file to know what it is.

### Optimal System (ours)

```
optimal://
│
├── .system/                            # ── INFRASTRUCTURE (not context) ──
│   ├── config.yaml                     # System config (paths, budgets, thresholds)
│   ├── topology.yaml                   # Node types, routing rules, endpoint definitions
│   ├── index.db                        # SQLite + FTS5 (indexes everything below)
│   └── cache/
│       └── l0.md                       # Pre-computed L0 (always-loaded context, ~2K tokens)
│
├── resources/                          # ── SHARED KNOWLEDGE ──
│   │                                   # (= OpenViking resources/ but TYPED)
│   ├── knowledge/                      # Domain knowledge, reference docs
│   │   ├── signal-theory.md            #   ← classified: genre=white-paper, type=inform
│   │   ├── sparql-owl.md
│   │   └── firecracker-vms.md
│   ├── patterns/                       # Proven patterns (promoted from agent/memories)
│   │   ├── wyatt-carson-principle.md   #   ← classified: genre=pattern, type=inform
│   │   └── noise-filter-positioning.md
│   ├── templates/                      # Genre skeletons, output templates
│   │   ├── brief.md                    #   ← the skeleton for composing briefs
│   │   ├── spec.md
│   │   └── decision-log.md
│   └── tools/                          # Tool configs, integration definitions
│       ├── slack.yaml
│       └── github.yaml
│
├── user/                               # ── USER CONTEXT ──
│   │                                   # (= OpenViking user/ but with TOPOLOGY)
│   ├── profile.md                      # Who this user is, role, bandwidth, preferences
│   ├── memories/                       # User's stated preferences and corrections
│   │   ├── dont-send-robert-specs.md   #   ← classified: genre=feedback, type=direct
│   │   ├── ed-prefers-voice.md
│   │   └── revenue-is-critical.md
│   └── nodes/                          # ── THE USER'S WORLD (organizational topology) ──
│       ├── entities/                   # People, orgs, tools
│       │   ├── roberto.md              #   ← type=person, bandwidth=high, receives=[spec,plan]
│       │   ├── ed-honour.md            #   ← type=person, bandwidth=medium, receives=[brief]
│       │   ├── robert-potter.md        #   ← type=person, bandwidth=low, receives=[brief ONLY]
│       │   ├── pedro.md
│       │   └── miosa-llc.md            #   ← type=org
│       └── operations/                 # Projects, programs, products
│           ├── miosa-platform/
│           │   ├── context.md          #   ← persistent facts (L2 detail)
│           │   ├── signal.md           #   ← weekly status (L1 summary)
│           │   └── decisions/
│           │       ├── 001-elixir-over-node.md
│           │       └── 002-firecracker-vms.md
│           ├── ai-masters/
│           │   ├── context.md
│           │   ├── signal.md
│           │   └── decisions/
│           │       └── 001-pricing-99.md
│           ├── agency-accelerants/
│           ├── content-creators/
│           └── money-revenue/          # Cross-cutting node — financial signals ALWAYS route here too
│               ├── context.md
│               └── signal.md
│
├── agent/                              # ── AGENT CONTEXT ──
│   │                                   # (= OpenViking agent/ but with SICA LEARNING)
│   ├── identity.md                     # Who this agent is, capabilities, constraints
│   ├── skills/                         # Learned executable workflows
│   │   ├── classify-signal.md          #   ← HOW to resolve S=(M,G,T,F,W)
│   │   ├── compose-genre.md            #   ← HOW to compose in each genre skeleton
│   │   ├── route-signal.md             #   ← HOW to determine which nodes get a signal
│   │   ├── assemble-context.md         #   ← HOW to build tiered context for a query
│   │   ├── weekly-review.md            #   ← HOW to run Alice's Friday review
│   │   └── brain-dump-intake.md        #   ← HOW to process Monday brain dumps
│   └── memories/                       # Agent's learning pipeline
│       ├── observations/               # SICA step 1: what the agent noticed
│       │   └── 2026-03-18-classification-error.md
│       ├── reflections/                # SICA step 2: what the agent concluded
│       │   └── financial-signals-need-cross-routing.md
│       └── corrections/                # SICA step 3: mistakes + lessons learned
│           └── robert-got-a-spec-should-have-been-brief.md
│
├── sessions/                           # ── CONVERSATION HISTORY ──
│   │                                   # (= OpenViking session/ but with EXTRACTED SIGNALS)
│   ├── current.jsonl                   # Active session (raw log)
│   └── history/
│       ├── 2026-03-18-brain-dump/
│       │   ├── summary.md              #   ← auto-generated L0/L1 summary
│       │   └── extracted/              #   ← signals extracted FROM the session
│       │       ├── decision-pricing.md #     → auto-routed to ai-masters/decisions/
│       │       └── task-ed-filming.md  #     → auto-routed to ai-masters/signal.md
│       └── 2026-03-17-ai-masters/
│           ├── summary.md
│           └── extracted/
│
└── inbox/                              # ── UNROUTED SIGNALS ──
    └── pending-001.md                  # Couldn't classify → human routes it later
```

### What Makes This BETTER (Not Just Different)

| What OpenViking Does | What We Do | WHY Ours Is Better |
|---------------------|-----------|-------------------|
| Agent reads `memories/memory-001.md` — has to parse it to know what it is | Agent reads frontmatter `genre: decision-log, type: decide` — KNOWS what it is before reading the body | **Retrieval is instant.** Filter 1000 files by genre in <1ms without reading content. OpenViking reads every file. |
| Agent stores a memory — goes into flat `memories/` folder | Agent ingests a signal — auto-classified, auto-routed to `ai-masters/decisions/` AND `money-revenue/` | **Information goes to the RIGHT PLACE automatically.** OpenViking's agent has to manually decide where to put things. |
| Agent retrieves context — directory search + semantic | Agent assembles context — BM25 + graph traversal + temporal decay + RRF fusion, within a token budget, shaped for the receiver | **Retrieval uses 4 search modes fused together.** OpenViking uses 2. And we respect token budgets — they don't. |
| Agent has skills — flat files describing capabilities | Agent has skills — each is a classified, versioned, executable workflow with SICA feedback loop | **Skills improve themselves.** Agent makes a mistake → logs correction → updates skill. OpenViking skills are static. |
| Agent learns — end-of-session memory extraction | Agent learns — OBSERVE → REFLECT → PROPOSE → TEST → INTEGRATE (triple-loop) | **Three levels of learning.** Did it work? Was it the right thing? Are we asking the right questions? OpenViking only asks "did it work?" |
| Session history — raw logs | Session history — auto-extracted signals routed to the right nodes + searchable summaries | **Past sessions are MINED for signals.** A decision from last Tuesday's call is already in the right operation folder. OpenViking leaves it buried in a log. |
| No organizational awareness | Full topology — nodes, entities, operations, routing rules, endpoint bandwidth profiles | **The agent understands the ORGANIZATION.** It knows Bob gets briefs, Frank gets explicit constraints, financial data always goes to money-revenue. OpenViking knows nothing about organizational structure. |
| `l0` / `l1` / `l2` tiers | `l0` / `l1` / `l2` / `l3` + token budgets + adaptive scaling + promotion/demotion | **Budget-controlled tiered loading.** We never blow the context window. Content auto-promotes (accessed often → higher tier) and auto-demotes (stale → lower tier). OpenViking loads tiers but has no budget system. |

### Why This Structure Works for Any Agent

OpenViking proved the filesystem paradigm works — agents navigate directories, not databases.
We keep that. But their files are dumb — just content in a folder. Ours are SMART:

1. **Every file is classified** — the agent doesn't have to figure out what it's looking at
2. **Every file has tier summaries** — the agent can load the headline without reading the whole thing
3. **Skills are real skills** — not just saved text, but typed workflows the agent can execute
4. **Memories feed back** — corrections become lessons, observations become patterns, patterns become skills
5. **User context is structured** — not "here's some memories about the user" but a full organizational map

An agent using our system can answer "What did Ed say about pricing?" in <80ms by:
1. Navigating to `user/nodes/operations/ai-masters/`
2. Searching `decisions/` with BM25
3. Loading the L0 summary from frontmatter (no need to read the full file)
4. Drilling into L2 only if the user asks for detail

Same filesystem paradigm. 10x more intelligence per file.

### File Format — Every File is a Signal

Every `.md` file in the context store carries YAML frontmatter:

```yaml
---
id: sig_20260318_001
signal:
  mode: linguistic
  genre: decision-log
  type: decide
  format: markdown
  structure: decision_log_skeleton
node: operations/ai-masters
entities: [ed-honour, robert-potter, roberto]
created_at: 2026-03-18T01:00:00Z
valid_from: 2026-03-18T01:00:00Z
valid_until: null          # null = still current
supersedes: null           # or ID of previous version
sn_ratio: 0.95
tiers:
  l0: "AI Masters pricing set at $99/mo community, $8-9K/yr premium"
  l1: "Ed and Alice agreed on two-tier pricing for AI Masters course..."
  # l2 = the full file body
---

# AI Masters Pricing Decision

[full content here — this IS the L2]
```

**This is the key innovation.** OpenViking stores files. We store CLASSIFIED SIGNALS with pre-computed tier summaries. The system knows:
- What this IS (genre: decision-log)
- Who it's ABOUT (entities)
- Where it BELONGS (node)
- How CURRENT it is (valid_from/valid_until)
- What it REPLACED (supersedes)
- The HEADLINE (l0) and SUMMARY (l1) without reading the whole file

---

## 2. Context Storage — Three Layers

```
┌─────────────────────────────────────────┐
│  Layer 1: Markdown Files (source of truth) │
│  Git-versionable. Human-readable.        │
│  Every file = classified signal.         │
└──────────────────┬──────────────────────┘
                   │ indexes
┌──────────────────▼──────────────────────┐
│  Layer 2: SQLite + FTS5 (fast search)    │
│  BM25 full-text. Temporal queries.       │
│  Pre-computed tier content.              │
└──────────────────┬──────────────────────┘
                   │ enriches
┌──────────────────▼──────────────────────┐
│  Layer 3: Knowledge Graph (reasoning)    │
│  SPARQL + OWL 2 RL. Inferred facts.     │
│  Relationship traversal.                 │
└─────────────────────────────────────────┘
```

### Layer 1: Markdown Files

- Source of truth. Always readable by humans and agents.
- Git-versionable (every change is a commit).
- Stored at `~/.optimal/context/` (or configured path).
- Organized in the filesystem hierarchy above.

### Layer 2: SQLite + FTS5

Indexes the markdown files for fast search. Schema:

```sql
-- Every signal in the system
CREATE TABLE signals (
  id TEXT PRIMARY KEY,
  path TEXT NOT NULL,               -- filesystem path
  -- Signal dimensions
  mode TEXT NOT NULL,
  genre TEXT NOT NULL,
  type TEXT NOT NULL,
  format TEXT NOT NULL,
  structure TEXT NOT NULL,
  -- Temporal
  created_at TEXT NOT NULL,         -- ISO 8601
  valid_from TEXT NOT NULL,
  valid_until TEXT,                 -- null = current
  supersedes TEXT,                  -- previous signal ID
  -- Classification
  node TEXT NOT NULL,               -- which operation/entity this belongs to
  sn_ratio REAL NOT NULL,
  -- Pre-computed tiers
  l0_summary TEXT NOT NULL,         -- ~10 words, always loaded
  l1_description TEXT NOT NULL,     -- ~50 words, loaded when relevant
  content TEXT NOT NULL             -- full L2 body
);

-- Full-text search
CREATE VIRTUAL TABLE signals_fts USING fts5(
  l0_summary, l1_description, content,
  content=signals, content_rowid=rowid,
  tokenize='porter unicode61'
);

-- People, projects, tools, orgs
CREATE TABLE entities (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,               -- person/project/tool/org
  node TEXT,                        -- home node
  properties TEXT,                  -- JSON blob
  last_seen TEXT NOT NULL
);

-- Relationships between anything
CREATE TABLE edges (
  source_id TEXT NOT NULL,
  target_id TEXT NOT NULL,
  relation TEXT NOT NULL,           -- manages/works_on/depends_on/etc
  weight REAL DEFAULT 1.0,
  valid_from TEXT NOT NULL,
  valid_until TEXT,
  UNIQUE(source_id, target_id, relation, valid_from)
);

-- Append-only decision audit trail
CREATE TABLE decisions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp TEXT NOT NULL,
  actor TEXT NOT NULL,              -- who or what made this decision
  action TEXT NOT NULL,             -- created/updated/promoted/demoted/archived
  signal_id TEXT,
  what_changed TEXT NOT NULL,
  why TEXT NOT NULL,
  context TEXT
);
```

### Layer 3: Knowledge Graph (SPARQL + OWL)

Uses the existing `miosa_knowledge` engine. Same data as `edges` table but queryable via SPARQL with OWL 2 RL reasoning (16 forward-chaining rules).

What the graph gives us that SQLite can't:
- **Inference**: "Alice manages MIOSA" + "Erin works on MIOSA" → auto-derives "Alice indirectly manages Erin"
- **Transitive closure**: Decision chains, supersession chains
- **2-hop traversal**: "Give me everything related to things Erin works on"

---

## 3. Context Loading — Tiered Disclosure

Like OpenViking's L0/L1/L2 but with **token budgets** and **receiver bandwidth matching**.

### The Four Tiers

| Tier | Name | Budget | What's In It | When Loaded |
|------|------|--------|-------------|-------------|
| **L0** | HOT | ~2K tokens | Headlines. Always present. | Every session, non-negotiable |
| **L1** | WARM | ~10K tokens | Summaries of relevant stuff | When task context identified |
| **L2** | COLD | ~50K tokens | Full detail | On explicit drill-down or L1 miss |
| **L3** | ARCHIVE | fragments only | Historical, superseded | Never auto-loaded. Search only. |

### L0 — Always Loaded (~2K tokens)

```markdown
## System Identity
You are the Optimal System operating for Alice H. Luna.
Active operations: MIOSA Platform, AI Masters, Agency Accelerants, OS Accelerator, Mosaic Effect.

## Top Priorities (this week)
1. AI Masters: Ed filming technical modules, Robert filming sales modules
2. Revenue: Dan pipeline, Ed $20K/mo target
3. MIOSA: Erin PR review, Carol audit

## Recent Decisions
- 2026-03-17: AI Masters pricing — $99/mo community, $8-9K premium
- 2026-03-16: OpenClaw rejected, use NanoClaw/IronClaw
- 2026-03-15: Course needs monthly refresh

## Key People (by recent mention)
Alice (CEO), Alice (course partner), Bob (sales), Dan (AA/content), Erin (frontend)
```

This is ~400 tokens. Loaded EVERY session. Never skipped. Regenerated when:
- A new decision is logged
- Weekly priorities change
- An entity's status changes

### L1 — Task-Relevant (~10K tokens)

When a task is identified (e.g., "work on AI Masters pricing"), load:
- Full `signal.md` for `operations/ai-masters/`
- Entity profiles for mentioned people (Ed, Bob)
- Recent decisions tagged to this node
- Relevant patterns from `knowledge/patterns/`

**Selection scoring:**
```
score = keyword_match × 0.40
      + node_match × 0.30        # signal belongs to same node as query
      + recency × 0.20           # exponential decay from created_at
      + importance × 0.10        # genre-based weight (decision > note)
```

Load signals in score order until budget is filled. Never truncate a signal mid-content — skip to next if it doesn't fit.

### L2 — On Demand (~50K tokens)

Loaded when:
- L1 doesn't answer the question
- User explicitly asks for detail ("show me the full spec")
- Entity relationship traversal needed

Contains full `context.md` files, complete decision chains, deep relationship graphs.

### L3 — Archive (search only)

All historical sessions, superseded signals, archived operations. Never auto-injected.
Retrieved as fragments with provenance: `[L3 archive — sig_20260215_003, retrieved: 2026-03-18]`

### Promotion / Demotion (Automatic)

- **Promote L2 → L1**: Signal accessed 3+ consecutive sessions for same scope
- **Demote L1 → L2**: Signal not accessed in 30 days
- Every promotion/demotion logged to `decisions` table

---

## 4. Context Retrieval — Hybrid Search

Query comes in. Multiple search modes fire **in parallel**. Results merged via RRF.

```
Query: "What did Ed say about pricing?"
         │
         ▼
┌─────────────────────────────────────────────┐
│              PARALLEL SEARCH                 │
│                                              │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐ │
│  │  BM25    │  │  Graph   │  │ Temporal   │ │
│  │  (FTS5)  │  │ (SPARQL) │  │  (decay)   │ │
│  │  0.30    │  │  0.35    │  │  0.10      │ │
│  └────┬─────┘  └────┬─────┘  └─────┬─────┘ │
│       │              │              │        │
│  ┌────▼──────────────▼──────────────▼─────┐ │
│  │    Reciprocal Rank Fusion (RRF)        │ │
│  │    score = Σ 1/(60 + rank_i)           │ │
│  └────────────────┬───────────────────────┘ │
└───────────────────┼─────────────────────────┘
                    ▼
          Tier-Aware Assembly
          (fill L0 → L1 → L2 within budget)
```

### Search Modes

| Mode | What It Does | Weight | Speed |
|------|-------------|--------|-------|
| **BM25** | Full-text keyword search via FTS5 | 0.30 | <10ms |
| **Graph** | SPARQL traversal + OWL inference | 0.35 | <50ms |
| **Temporal** | Exponential decay by age + genre half-life | 0.10 | <5ms |
| **Vector** | Semantic similarity (Phase 3 — sqlite-vec) | 0.25 | <30ms |

**RRF fusion**: `score(signal) = Σ 1/(60 + rank_i(signal))` per mode. Modes that don't return a signal contribute 0.

### Temporal Half-Lives by Genre

| Genre | Half-Life | Why |
|-------|-----------|-----|
| message, transcript | 7 days | Ephemeral communication |
| standup, status-report | 14 days | Weekly cycle |
| plan | 30 days | Monthly planning |
| note | 60 days | Processing buffer |
| spec, prd | 180 days | Build artifacts |
| decision-log, adr | 365 days | Long-lived |
| pattern | 730 days | Institutional knowledge |
| entity | ∞ | People don't expire |

### Performance Targets

- BM25: <10ms
- Graph: <50ms
- Temporal: <5ms
- All parallel → total search: <60ms
- Assembly: <20ms
- **End-to-end: <80ms**

---

## 5. Context Intake — How Signals Enter the System

```
Raw input (voice, text, file, message)
    │
    ▼
┌──────────┐   ┌───────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ NORMALIZE │──▶│ CLASSIFY  │──▶│ COMPOSE  │──▶│  INDEX   │──▶│  ROUTE   │
│           │   │ S=(M,G,   │   │ Extract  │   │ SQLite + │   │ Which    │
│ HTML→md   │   │  T,F,W)   │   │ entities │   │ FTS5 +   │   │ node(s)  │
│ PDF→text  │   │ Measure   │   │ triples  │   │ SPARQL   │   │ gets     │
│ audio→txt │   │ S/N ratio │   │ tiers    │   │          │   │ this?    │
└──────────┘   └───────────┘   └──────────┘   └──────────┘   └──────────┘
```

### Step 1: NORMALIZE
Convert any input to markdown. Extract metadata (author, timestamp, source).

### Step 2: CLASSIFY
Resolve all 5 dimensions: `S=(M, G, T, F, W)`. Order: F→M→G→T→W.
Measure S/N ratio: 1.0 = fully classified, <0.6 = quarantine, <0.3 = reject.

### Step 3: COMPOSE
- Validate against genre skeleton ([R] required sections present?)
- Extract entities → map to canonical IDs
- Extract SPO triples (subject-predicate-object)
- Generate tier content: L0 (~10 words), L1 (~50 words), L2 = full body

### Step 4: INDEX
Single transaction writes to:
1. Markdown file to filesystem
2. SQLite `signals` table + `signals_fts`
3. SPARQL graph store (SPO triples)
4. `decisions` table if genre is decision-log/adr

### Step 5: ROUTE
1. Identify source node (where did this come from?)
2. Genre-based routing (which node types receive this genre?)
3. Entity mention routing (mentioned people/operations → their nodes)
4. Cross-cutting rules (financial data → ALWAYS also money-revenue node)

---

## 6. Context Evolution — Self-Improvement

### Session End → Memory Extraction

Like OpenViking's self-evolution, but with classification:

```
Session ends
    │
    ▼
Extract from conversation:
    ├── New decisions → decision-log signals (classified, routed)
    ├── New facts → entity updates or context.md updates
    ├── New patterns → knowledge/patterns/ (SICA: OBSERVE→REFLECT→PROPOSE)
    └── Session summary → sessions/ (auto-generated L0/L1)
```

### Three Feedback Loops

| Loop | Question | Frequency | What Changes |
|------|----------|-----------|--------------|
| **Single** | Did the signal arrive? Was it decoded? | Per signal | Confirmation, retry, escalate |
| **Double** | Were these the right priorities? | Weekly (Friday review) | Next week's L0, priorities |
| **Triple** | Are we asking the right questions? | Monthly | Node structure, routing rules |

### Observable Retrieval (OpenViking-inspired)

Every context assembly includes a retrieval trace:

```markdown
## Retrieval Trace
- Query: "AI Masters pricing"
- Modes: BM25 (4 results), Graph (2 results), Temporal (6 results)
- Top result: sig_20260317_pricing (RRF score: 0.089)
- Signals loaded: 8 (3.2K tokens)
- Tier: L1 (within 10K budget)
- Assembly time: 47ms
```

This makes context retrieval **debuggable**. If the wrong context shows up, you can see WHY.

---

## 7. Pluggability — How to Use This in Any App

### Minimal Integration (Tier 0 — any app)

```
1. Store markdown files in a directory
2. Add YAML frontmatter with signal dimensions
3. Pre-compute L0/L1 summaries in frontmatter
4. Read files with glob + grep (no database needed)
```

That's it. This is what OptimalOS does TODAY with Claude Code. It works.

### Standard Integration (Tier 1 — SQLite)

```
1. Everything from Tier 0
2. Add SQLite database indexing the files
3. FTS5 for full-text search
4. Temporal queries for "what's current?"
5. Pre-computed tier table for fast loading
```

Single binary (SQLite). No servers. Drop-in.

### Full Integration (Tier 2 — SignalGraph)

```
1. Everything from Tier 1
2. Add SPARQL graph for relationship reasoning
3. OWL inference for derived facts
4. Hybrid search with RRF fusion
5. Auto-routing engine
```

Requires Elixir runtime (miosa_knowledge). Or any SPARQL engine.

### Enterprise Integration (Tier 3 — MIOSA)

```
1. Everything from Tier 2
2. Add vector embeddings (semantic search)
3. Multi-tenant context isolation
4. L7 Governance (VSM autonomy levels)
5. Full 4-mode RRF with MCTS
```

Full platform deployment.

### The Plugin Interface

Any agent framework integrates via 6 functions — matching the base OpenViking operations + our additions:

```
# ─── BASE (matches OpenViking's core operations) ───

# Store something in the context filesystem
context.store(path, content, metadata?) → signal_id
# Example: context.store("user/memories/feedback/no-specs-for-robert.md", content)

# Read from a specific path (with tier control)
context.read(path, tier?) → content
# Example: context.read("user/nodes/operations/ai-masters/", tier="l0")
# Returns just the L0 summaries of everything in that directory

# Search across the whole filesystem
context.search(query, scope?, limit?) → ranked_results
# Example: context.search("Ed pricing", scope="user/nodes/operations/ai-masters/")

# ─── OPTIMAL ADDITIONS (what OpenViking doesn't have) ───

# Ingest + auto-classify + auto-route (the smart intake)
context.ingest(content) → {signal_id, classification, routed_to[]}
# Example: context.ingest("Customer called, $99 for community tier")
# → classifies as decision-log, routes to ai-masters + money-revenue

# Get assembled context for a task (tiered, budget-aware)
context.assemble(query, budget?) → {l0, l1, l2?, trace}
# Example: context.assemble("prepare AI Masters pricing brief for Bob", budget=10000)
# → loads relevant context, shapes for Robert's bandwidth (brief only, no specs)

# Get agent's learned skill for a task type
context.skill(task_type) → skill_definition
# Example: context.skill("weekly-review")
# → returns the agent's learned workflow for running Friday reviews
```

That's it. 3 base functions (store/read/search) that any filesystem agent can use.
3 smart functions (ingest/assemble/skill) that add Signal Theory intelligence.

An agent framework that ONLY uses store/read/search still gets:
- Organized filesystem hierarchy
- Pre-computed tier summaries in frontmatter
- Classified files it can filter by genre/type/mode

An agent framework that ALSO uses ingest/assemble/skill gets:
- Auto-classification and routing
- Budget-aware tiered context assembly
- Receiver-modeled output shaping
- Learned skills that improve over time

---

## 8. Build Sequence

### Phase 0: Filesystem + SQLite (NOW)

What we build first. No SPARQL, no vectors, no fancy shit.

1. Define the directory structure (`~/.optimal/context/`)
2. YAML frontmatter schema for every file
3. SQLite schema (signals, entities, edges, decisions)
4. FTS5 indexing
5. Basic intake: normalize → classify → write file + index
6. Basic retrieval: BM25 + temporal decay → tier assembly
7. L0 generator (auto-generate from current signals)

**This gets us a working system.** Everything after this is optimization.

### Phase 1: Wire Existing MIOSA Systems

1. miosa_signal → auto-classify on intake
2. miosa_knowledge → SPO triple extraction + graph queries
3. miosa_memory → episodic events, SICA observations
4. Hybrid search: BM25 + SPARQL + temporal + RRF

### Phase 2: Context Assembly Engine

1. Tier-aware loading orchestrator (L0→L1→L2 state machine)
2. Receiver bandwidth profiles (who gets what depth)
3. Genre conversion on delivery (spec→brief for salespeople)
4. Observable retrieval traces

### Phase 3: Vectors + Full Fusion

1. Vector embeddings (sqlite-vec, all-MiniLM-L6-v2)
2. 4-mode RRF (BM25 + graph + temporal + vector)
3. Auto-routing engine
4. MCTS for exploration queries

---

## 9. How This Compares

| Capability | OpenViking | Mem0 | Letta | Zep | **Optimal** |
|-----------|-----------|------|-------|-----|------------|
| Filesystem hierarchy | ✓ | — | — | — | **✓** |
| Signal classification | — | — | — | — | **✓ (5 dimensions)** |
| Tiered loading | ✓ (L0/L1/L2) | — | ✓ (3 tiers) | — | **✓ (L0/L1/L2/L3 + budgets)** |
| Token budget control | — | — | — | — | **✓** |
| Receiver modeling | — | — | — | — | **✓** |
| Genre composition | — | — | — | — | **✓ (143 genres)** |
| OWL reasoning | — | — | — | — | **✓ (16 rules)** |
| Hybrid search | — | — | — | ✓ (BM25+vec+graph) | **✓ (BM25+graph+temporal+vec)** |
| Temporal versioning | — | — | — | ✓ (bi-temporal) | **✓ (append-only + supersedes)** |
| Self-evolution | ✓ (single-loop) | — | — | — | **✓ (triple-loop SICA)** |
| Observable retrieval | ✓ | — | — | — | **✓** |
| Multi-node topology | — | — | — | — | **✓** |
| Pluggable tiers (0-3) | — | — | — | — | **✓** |
| Cross-org routing | — | — | — | — | **✓** |

---

## References

- [OpenViking](https://github.com/volcengine/OpenViking) — filesystem paradigm, tiered loading
- [Signal Theory](../../docs/taxonomy/glossary.md) — S=(M,G,T,F,W) classification
- [Layer 4: Interface](04-interface.md) — tiered disclosure spec
- [Layer 5: Data](05-data.md) — storage and retrieval spec
- [Intake Pipeline](../operations/intake-pipeline.md) — 6-stage intake
- [Search & Retrieval](../operations/search-retrieval.md) — hybrid search spec
- Luna, R.H. — *Signal Theory: The Architecture of Optimal Intent Encoding* (MIOSA Research, Feb 2026)
