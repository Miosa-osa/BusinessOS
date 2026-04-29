---
signal:
  mode: visual
  genre: spec
  type: inform
  format: markdown
  structure: architecture-overview
---

# OptimalOS — Full System Architecture

> The synthesized optimal cognitive operating system.
> 42+ modules. 20+ CLI commands. 10-layer architecture. Zero competitors cover the full stack.

---

## What This Is

A **cognitive operating system** — not a note app, not a wiki, not a knowledge base. OptimalOS is an externalized decision tree library with an intelligent engine that classifies, routes, stores, searches, learns from, and maintains every piece of information that enters the system.

Built on **Signal Theory** — the principle that every output is a Signal `S = (Mode, Genre, Type, Format, Structure)` and the goal is always to maximize signal-to-noise ratio.

Built on **Elixir/OTP** — fault-tolerant, concurrent, hot-reloadable. Each component is independently supervised. The system self-heals.

---

## The 30-Second Version

```
INPUT: "Customer called about pricing, wants $2K/seat"
  │
  ▼
┌──────────────────────────────────────────────────────┐
│              OPTIMAL ENGINE (Elixir/OTP)              │
│                                                       │
│  1. CLASSIFY  →  S=(linguistic, transcript, decide,   │
│                    markdown, meeting-notes)            │
│  2. ROUTE     →  ai-masters + money-revenue + team    │
│  3. STORE     →  Markdown files + SQLite + FTS5       │
│  4. GRAPH     →  Ed→mentioned_in→pricing-call         │
│  5. LEARN     →  SICA records pattern, Cortex synth.  │
│  6. EMBED     →  nomic-embed-text → vector store      │
│                                                       │
│  Later: "What pricing decisions were made?"           │
│                                                       │
│  7. SEARCH    →  BM25 + vector + RRF + graph boost    │
│  8. ASSEMBLE  →  L0(summary) + L1(detail) + L2(full)  │
│  9. COMPOSE   →  Shaped for the receiver's genre       │
│ 10. VERIFY    →  Health checks, reweave, fidelity     │
│                                                       │
│  Returns: Tiered context within token budget           │
└──────────────────────────────────────────────────────┘
```

---

## 10-Layer Architecture

### Layer 0: Data Storage

**What it does:** Persistent storage — SQLite, ETS hot cache, filesystem, vector embeddings.

| Module | Type | Responsibility |
|--------|------|---------------|
| `Store` | GenServer | SQLite connection, CRUD operations, ETS cache (500 items), schema migrations |
| `VectorStore` | Stateless | Vector embedding storage/retrieval, cosine similarity search via SQLite `vectors` table |
| `L0Cache` | GenServer | Always-loaded ~2K token context, auto-refresh every 5 minutes |

**Tables:** `contexts`, `contexts_fts` (FTS5), `entities`, `edges`, `decisions`, `sessions`, `vectors`, `observations`

**Storage hierarchy:**
```
Markdown files  →  Source of truth (human-readable, git-trackable)
SQLite + FTS5   →  Search index (regenerated from files via mix optimal.index)
ETS             →  Hot cache (sub-millisecond, lost on restart)
Vectors         →  Embedding store (768-dim nomic-embed-text)
```

---

### Layer 1: Classification

**What it does:** Auto-classifies every piece of content on 5 Signal Theory dimensions + quality gates.

| Module | Type | Responsibility |
|--------|------|---------------|
| `Classifier` | Stateless | S=(M,G,T,F,W) detection from content + YAML frontmatter. 16 genre patterns, 5 mode patterns. Entity extraction. L0/L1 generation. |
| `Bridge.Signal` | Stateless | Classification augmentation via MiosaSignal. Quality audit via 11 failure mode detectors (Shannon/Ashby/Beer/Wiener violations). S/N ratio scoring. |

**Classification pipeline:**
```
Raw text → Parse YAML frontmatter → Detect Mode → Detect Genre → Detect Type
         → Determine Format → Infer Structure → Extract Entities → Generate L0/L1
         → MiosaSignal augmentation → FailureModes audit → S/N scoring
```

**Quality gates:** Content with S/N ratio < 0.3 is rejected by the intake pipeline.

---

### Layer 2: Routing

**What it does:** Auto-routes content to the correct organizational node(s) with cross-referencing.

| Module | Type | Responsibility |
|--------|------|---------------|
| `Router` | GenServer | 13 topology-driven routing rules. Multi-destination routing. Keyword matching, entity matching, financial detection. |
| `Topology` | Stateless | Loads `topology.yaml` — nodes, endpoints (people), routing rules, genre half-lives |
| `URI` | Stateless | `optimal://` addressing scheme. Parse, resolve, construct URIs for all context types. |

**Routing rules:**
```
Keyword match     →  Primary node (e.g., "AI Masters" → ai-masters)
Financial data    →  Always cross-ref to money-revenue
Entity mentions   →  Cross-ref to team registry
Domain keywords   →  Route to domain node
Ambiguous         →  Route to inbox (09-new-stuff)
```

---

### Layer 3: Search & Retrieval

**What it does:** Hybrid search combining FTS5 BM25, vector similarity, temporal decay, S/N ratio, and graph boost. Tiered context assembly.

| Module | Type | Responsibility |
|--------|------|---------------|
| `SearchEngine` | GenServer | Hybrid search orchestrator. BM25 + vector + RRF fusion. Temporal decay with genre-specific half-lives. Intent analysis. Graph boost. |
| `ContextAssembler` | Stateless | Tiered loading: L0 (~2K tokens always), L1 (~10K task-relevant), L2 (~50K deep). Token budget management. |
| `IntentAnalyzer` | Stateless | Query intent classification. Expands queries with synonyms. Infers node hints. Detects question type. |

**Search pipeline:**
```
Query → IntentAnalyzer (expand, classify)
      → FTS5 BM25 search (keyword relevance)
      → Vector search (semantic similarity via Ollama embeddings)
      → RRF fusion (alpha * BM25_norm + (1-alpha) * vector_sim)
      → Temporal decay (genre-specific half-lives)
      → S/N ratio boost
      → Graph boost (Knowledge Bridge)
      → Ranked results
```

---

### Layer 4: Knowledge Graph

**What it does:** Entity-context-node graph with triangle detection, cluster analysis, hub identification, and missing edge reflection.

| Module | Type | Responsibility |
|--------|------|---------------|
| `Graph` | Stateless | Edge CRUD. 5 edge types: `mentioned_in`, `lives_in`, `works_on`, `cross_ref`, `supersedes`. Subgraph queries. Topology seeding. |
| `Bridge.Knowledge` | Stateless | SPARQL triple store sync. OWL 2 RL reasoning. Graph boost for search. Materialization. |
| **`GraphAnalyzer`** | Stateless | **NEW** — Triangle detection (synthesis opportunities), BFS connected components, hub entity detection (>2σ degree). |
| **`Reflector`** | Stateless | **NEW** — Entity co-occurrence scanning. Finds pairs that appear together but lack edges. LLM-enhanced relationship classification. |

**Edge types:**
```
entity  --mentioned_in-->  context     (person appears in document)
context --lives_in-->      node        (document belongs to folder)
entity  --works_on-->      node        (person works on project)
context --cross_ref-->     node        (document cross-referenced)
context --supersedes-->    context     (newer replaces older)
```

**Graph analytics (NEW):**
```
Triangles:  A→B and A→C exist but B→C doesn't → synthesis opportunity
Clusters:   BFS connected components → isolated knowledge islands
Hubs:       Entities with degree > mean + 2σ → key connectors
Reflection: Entity co-occurrence without edges → missing relationships
```

---

### Layer 5: Intelligence

**What it does:** Simulation, planning, semantic processing, and local LLM integration.

| Module | Type | Responsibility |
|--------|------|---------------|
| `MCTS` | Stateless | Monte Carlo Tree Search for decision planning. UCB1 selection, rollout simulation, backpropagation. |
| `MonteCarlo` | Stateless | Monte Carlo probability estimation. Random sampling with Bayesian updating. |
| `Simulator` | GenServer | Scenario planning and impact analysis. "What if X happens?" simulation across the knowledge graph. |
| `SemanticProcessor` | Stateless | Semantic similarity, concept extraction, relationship inference via local LLM. |
| `Ollama` | Stateless | Thin HTTP wrapper for local Ollama API. Embedding (nomic-embed-text), generation (qwen3:8b). 60s availability cache. |

---

### Layer 6: Processing Pipeline

**What it does:** End-to-end ingestion pipeline — raw input to classified, routed, stored, indexed, graph-connected context.

| Module | Type | Responsibility |
|--------|------|---------------|
| `Intake` | GenServer | Full ingestion pipeline: classify → route → write files → index → create edges → record in memory systems |
| `Indexer` | GenServer | File crawler. Walks the OptimalOS directory tree, classifies each file, stores in SQLite, creates FTS5 entries. |
| `MemoryExtractor` | Stateless | 6-category memory extraction from text (fact, preference, decision, relationship, skill, context). LLM or regex fallback. |
| `SessionCompressor` | Stateless | Compresses session transcripts preserving key information. Extractive then abstractive summarization. |

---

### Layer 7: Agent Interface

**What it does:** Signal Theory encoding, genre skeleton system, receiver-adaptive composition.

| Module | Type | Responsibility |
|--------|------|---------------|
| `Composer` | Stateless | Renders signals for specific receivers. Applies genre skeletons. Adapts formality, detail level, vocabulary per person. |
| `CortexFeed` | Stateless | Feeds knowledge graph data to Cortex for synthesis bulletins. |
| `Context` | Struct | Universal context struct. 4 types (resource, memory, skill, signal). Score field for search ranking. |
| `Signal` | Struct | Signal struct with full S=(M,G,T,F,W) dimensions. |

**Genre skeletons (10 built-in):**
```
transcript  →  Participants, Key Points, Decisions, Action Items, Open Questions
brief       →  Objective, Key Messages, Call to Action, Supporting Materials
spec        →  Goal, Requirements, Constraints, Architecture, Acceptance Criteria
plan        →  Objective, Non-Negotiables, Time Blocks, Dependencies, Success Criteria
note        →  Context, Content, Route
proposal    →  Problem, Solution, Investment, Timeline, Deliverables
pitch       →  Hook, Problem, Solution, Proof, Ask
report      →  Summary, Findings, Analysis, Recommendations
postmortem  →  Timeline, Impact, Root Cause, Action Items, Lessons
standup     →  Yesterday, Today, Blockers
```

---

### Layer 8: Learning & Feedback

**What it does:** Self-improvement loop — SICA learning, episodic memory, friction capture, evidence synthesis.

| Module | Type | Responsibility |
|--------|------|---------------|
| `Bridge.Memory` | Stateless | Episodic event recording, SICA observation, Cortex injection, learning metrics. |
| `MiosaMemory.Episodic` | GenServer | Temporal event recording. Every search, intake, and mutation is logged. |
| `MiosaMemory.Learning` | GenServer | SICA self-improvement: observe → classify → detect patterns → generate skills → error recovery. |
| `MiosaMemory.Cortex` | GenServer | LLM-powered synthesis bulletins. Periodically digests events into actionable summaries. |
| **`RememberLoop`** | Stateless | **NEW** — 3-mode friction capture: explicit observations, contextual scanning, session mining. Escalation when category reaches 3+ observations. |
| **`RethinkEngine`** | Stateless | **NEW** — Evidence synthesis when cumulative confidence ≥ 1.5. Gathers observations + search results → generates structured rethink reports with proposed context.md updates. |

**SICA learning loop:**
```
1. OBSERVE  →  Record every mutation (intake, search, index)
2. CLASSIFY →  Categorize the pattern (success, failure, novel)
3. DETECT   →  Find recurring patterns across observations
4. GENERATE →  Create skills from repeated successful patterns
5. RECOVER  →  Auto-classify and handle errors
```

**Friction capture (NEW):**
```
Explicit:   "always check duplicates" → classify → store in observations
Contextual: Scan recent contexts for "no, not that", "wrong", "should have"
Mining:     Bulk extract patterns from sessions via MemoryExtractor
Escalation: 3+ observations in same category → flag for RethinkEngine
Rethink:    confidence ≥ 1.5 → synthesize → propose context.md updates
```

---

### Layer 9: Operating System

**What it does:** Health monitoring, backward-pass reweaving, L0 fidelity verification, daily rhythm support.

| Module | Type | Responsibility |
|--------|------|---------------|
| **`HealthDiagnostics`** | Stateless | **NEW** — 10 diagnostic checks: orphaned contexts, stale signals, FTS drift, entity merge candidates, node imbalance, duplicates, broken references, embedding coverage, quality distribution. |
| **`Reweaver`** | Stateless | **NEW** — Backward pass: given topic → search + graph → find stale contexts → generate update suggestions. LLM-enhanced diff suggestions when Ollama available. |
| **`VerifyEngine`** | Stateless | **NEW** — Cold-read L0 fidelity test. Samples contexts, evaluates how well L0 abstracts represent full content. Jaccard similarity fallback, LLM scoring when available. |

**Health checks (10):**
```
 1. Orphaned contexts     — contexts with zero edges
 2. Stale signals         — modified_at > 30 days ago
 3. Missing cross-refs    — multi-node contexts without cross_ref edges
 4. FTS/index drift       — count mismatch contexts vs contexts_fts
 5. Entity merge candidates — duplicate entities by lowercase name
 6. Node imbalance        — nodes with >3x mean context count
 7. Duplicate detection   — identical titles within same node
 8. Broken references     — supersedes pointing to nonexistent IDs
 9. Embedding coverage    — ratio of vectors to total contexts
10. Quality distribution  — flag if >20% have sn_ratio < 0.4
```

---

## Complete Module-to-Layer Map

| # | Module | Layer | Type | Process? |
|---|--------|-------|------|----------|
| 1 | `OptimalEngine` | API | Facade | No |
| 2 | `Store` | L0: Data | GenServer | Yes (supervised) |
| 3 | `VectorStore` | L0: Data | Stateless | No |
| 4 | `L0Cache` | L0: Data | GenServer | Yes (supervised) |
| 5 | `Classifier` | L1: Classification | Stateless | No |
| 6 | `Bridge.Signal` | L1: Classification | Stateless | No |
| 7 | `Router` | L2: Routing | GenServer | Yes (supervised) |
| 8 | `Topology` | L2: Routing | Stateless | No |
| 9 | `URI` | L2: Routing | Stateless | No |
| 10 | `SearchEngine` | L3: Search | GenServer | Yes (supervised) |
| 11 | `ContextAssembler` | L3: Search | Stateless | No |
| 12 | `IntentAnalyzer` | L3: Search | Stateless | No |
| 13 | `Graph` | L4: Knowledge | Stateless | No |
| 14 | `Bridge.Knowledge` | L4: Knowledge | Stateless | No |
| 15 | `GraphAnalyzer` | L4: Knowledge | Stateless | No |
| 16 | `Reflector` | L4: Knowledge | Stateless | No |
| 17 | `MCTS` | L5: Intelligence | Stateless | No |
| 18 | `MonteCarlo` | L5: Intelligence | Stateless | No |
| 19 | `Simulator` | L5: Intelligence | GenServer | Yes (supervised) |
| 20 | `SemanticProcessor` | L5: Intelligence | Stateless | No |
| 21 | `Ollama` | L5: Intelligence | Stateless | No |
| 22 | `Intake` | L6: Pipeline | GenServer | Yes (supervised) |
| 23 | `Indexer` | L6: Pipeline | GenServer | Yes (supervised) |
| 24 | `MemoryExtractor` | L6: Pipeline | Stateless | No |
| 25 | `SessionCompressor` | L6: Pipeline | Stateless | No |
| 26 | `Composer` | L7: Agent | Stateless | No |
| 27 | `CortexFeed` | L7: Agent | Stateless | No |
| 28 | `Context` | L7: Agent | Struct | No |
| 29 | `Signal` | L7: Agent | Struct | No |
| 30 | `Bridge.Memory` | L8: Learning | Stateless | No |
| 31 | `MiosaMemory.Episodic` | L8: Learning | GenServer | Yes (supervised) |
| 32 | `MiosaMemory.Learning` | L8: Learning | GenServer | Yes (supervised) |
| 33 | `MiosaMemory.Cortex` | L8: Learning | GenServer | Yes (supervised) |
| 34 | `RememberLoop` | L8: Learning | Stateless | No |
| 35 | `RethinkEngine` | L8: Learning | Stateless | No |
| 36 | `HealthDiagnostics` | L9: Operating System | Stateless | No |
| 37 | `Reweaver` | L9: Operating System | Stateless | No |
| 38 | `VerifyEngine` | L9: Operating System | Stateless | No |
| 39 | `Session` | Sessions | GenServer | Yes (dynamic) |

**Process count:** 12 supervised + N dynamic sessions

---

## OTP Supervision Tree

```
OptimalEngine.Application (one_for_one)
│
├── OptimalEngine.Store              ← SQLite + ETS cache
├── OptimalEngine.Router             ← 13 routing rules from topology
├── OptimalEngine.Indexer            ← File crawler + classification pipeline
├── OptimalEngine.SearchEngine       ← Hybrid BM25 + vector + RRF + graph
├── OptimalEngine.L0Cache            ← Always-loaded context, auto-refresh
├── OptimalEngine.Intake             ← Raw text → classified → stored → indexed
├── OptimalEngine.Simulator          ← Scenario planning + MCTS
│
├── Registry (SessionRegistry)       ← Session name lookup
├── DynamicSupervisor (Sessions)     ← Per-agent conversation state
│   └── Session (per-agent)          ← Temporary restart
│
├── MiosaMemory.Episodic             ← Temporal event recording
├── MiosaMemory.Cortex               ← LLM synthesis bulletins
└── MiosaMemory.Learning             ← SICA self-improvement loop
```

---

## Data Flow Diagrams

### Write Path (Ingestion)

```
Raw Input
    │
    ▼
CLASSIFIER ──────────────────────────────────────────────────
  │ Parse YAML frontmatter (if present)
  │ Detect S=(Mode, Genre, Type, Format, Structure)
  │ Extract entities from topology roster
  │ Generate L0 abstract + L1 overview
  │ MiosaSignal augmentation + FailureModes audit
  │ Reject if S/N < 0.3
    │
    ▼
ROUTER ──────────────────────────────────────────────────────
  │ Keyword match → primary node
  │ Financial data → money-revenue (always)
  │ Entity mentions → team registry
  │ Multi-destination → cross-ref edges
    │
    ▼
WRITER ──────────────────────────────────────────────────────
  │ Apply genre skeleton template
  │ Write YAML-frontmatter markdown files
  │ Write to primary + cross-ref destinations
    │
    ▼
STORE + INDEX ───────────────────────────────────────────────
  │ Insert into contexts table
  │ FTS5 index (BM25)
  │ Create entities
  │ Create edges (mentioned_in, lives_in, cross_ref, supersedes)
  │ Generate embedding → vectors table (if Ollama available)
    │
    ▼
MEMORY SYSTEMS ──────────────────────────────────────────────
  │ Episodic: record temporal event
  │ SICA: observe mutation pattern
  │ Cortex: queue for synthesis
    │
    ▼
DONE → {:ok, %{context, files_written, routed_to, uri}}
```

### Read Path (Retrieval)

```
Query: "What pricing decisions were made?"
    │
    ▼
INTENT ANALYZER ─────────────────────────────────────────────
  │ Classify question type
  │ Expand query with synonyms
  │ Infer node hints
    │
    ▼
HYBRID SEARCH ───────────────────────────────────────────────
  │ FTS5 BM25 (keyword relevance)
  │ Vector search (semantic similarity)
  │ RRF fusion: α * BM25_norm + (1-α) * vector_sim
  │ Temporal decay (genre-specific half-lives)
  │ S/N ratio boost
  │ Graph boost (Knowledge Bridge traversal)
    │
    ▼
CONTEXT ASSEMBLER ───────────────────────────────────────────
  │ L0: Always-loaded (~2K tokens) — Cortex bulletin + active ops
  │ L1: Task-relevant (~10K tokens) — top search results, L1 overviews
  │ L2: Deep retrieval (~50K tokens) — full content, decision history
    │
    ▼
COMPOSER (optional) ─────────────────────────────────────────
  │ Shape output for specific receiver
  │ Apply genre competence (brief for salespeople, spec for devs)
    │
    ▼
RETURN → {:ok, %{l0, l1, l2, total_tokens, sources, scores}}
```

### Learning Loop

```
Every mutation (intake, search, index)
    │
    ▼
EPISODIC ────────── Record temporal event with metadata
    │
    ▼
SICA ───────────── Observe → Classify → Detect patterns → Generate skills
    │
    ▼
CORTEX ─────────── Periodically synthesize bulletin from recent events
    │
    ▼
L0 CACHE ───────── Inject synthesized awareness into always-loaded context
```

### Reweave Flow (NEW)

```
Topic: "Alice"
    │
    ▼
SEARCH ─────────── Find all contexts mentioning topic
    │
    ▼
GRAPH ──────────── Find entity → context edges
    │
    ▼
SCORE STALENESS ── Compute days since last update
    │
    ▼
SUGGEST ────────── LLM generates specific update suggestions
                   (or flag as "potentially outdated" without LLM)
```

### Remember → Rethink Flow (NEW)

```
Observation: "always check duplicates"
    │
    ▼
REMEMBER ──────── Classify category → Store in observations table
    │
    ▼
ACCUMULATE ────── 3+ observations in same category?
    │
    ▼
RETHINK ───────── confidence ≥ 1.5?
    │              → Gather all evidence (observations + search)
    │              → LLM synthesizes into actionable knowledge
    │              → Propose context.md updates
    │
    ▼
REPORT ────────── Structured synthesis with evidence + proposals
                  (never auto-applies by default)
```

---

## CLI Command Reference

| Command | Description |
|---------|------------|
| `mix optimal.index` | Full reindex of all files |
| `mix optimal.search "query"` | Hybrid search with scoring |
| `mix optimal.ingest "text"` | Auto-classify + route + store |
| `mix optimal.intake "text"` | Full intake pipeline with file writing |
| `mix optimal.assemble "topic"` | Tiered context assembly (L0/L1/L2) |
| `mix optimal.l0` | Show always-loaded L0 context |
| `mix optimal.ls "optimal://..."` | Browse contexts by URI |
| `mix optimal.read "optimal://..."` | Read a specific context |
| `mix optimal.stats` | Store statistics |
| `mix optimal.graph` | Knowledge graph stats + sample edges |
| `mix optimal.graph triangles` | Find synthesis opportunities (missing edges) |
| `mix optimal.graph clusters` | Find isolated knowledge clusters |
| `mix optimal.graph hubs` | Find hub entities (>2σ connections) |
| `mix optimal.knowledge` | SPARQL knowledge graph operations |
| `mix optimal.simulate` | Scenario simulation |
| `mix optimal.impact` | Impact analysis |
| `mix optimal.health` | 10 diagnostic health checks |
| `mix optimal.reflect` | Find missing edges from co-occurrences |
| `mix optimal.reweave "topic"` | Find stale contexts + suggest updates |
| `mix optimal.verify` | L0 fidelity cold-read test |
| `mix optimal.remember "text"` | Store observation / mine friction |
| `mix optimal.rethink "topic"` | Synthesize accumulated observations |

---

## Competitive Positioning

### OptimalOS vs 8 Competitors

| Capability | OptimalOS | Ars Contexta | Obsidian | Notion | Mem.ai | Roam | Logseq | Reflect |
|-----------|-----------|-------------|----------|--------|--------|------|--------|---------|
| Auto-classification (5 dim) | **S=(M,G,T,F,W)** | Partial (3 dim) | Manual tags | Manual | AI tags | Manual | Manual | AI tags |
| Auto-routing (multi-dest) | **13 rules** | None | Manual | Manual | Auto-folder | Manual | Manual | None |
| Hybrid search (BM25+vector+RRF) | **Yes** | BM25 only | Plugin | Basic | Vector only | Basic | Basic | AI search |
| Knowledge graph | **SQLite+SPARQL+OWL** | None | Plugin | None | None | Backlinks | Graph | None |
| Graph analytics (triangles/hubs) | **Yes** | None | None | None | None | None | None | None |
| Simulation (MCTS/Monte Carlo) | **Yes** | None | None | None | None | None | None | None |
| Quality gates (S/N ratio) | **11 failure modes** | None | None | None | None | None | None | None |
| Self-improvement (SICA) | **Learning loop** | None | None | None | None | None | None | None |
| Genre system (10 skeletons) | **Receiver-adapted** | None | Templates | Templates | None | None | Templates | None |
| Health diagnostics (10 checks) | **Yes** | Partial | None | None | None | None | None | None |
| Backward-pass reweave | **Yes** | Yes | None | None | None | None | None | None |
| Friction capture + rethink | **Yes** | None | None | None | None | None | None | None |
| L0 fidelity verification | **Yes** | None | None | None | None | None | None | None |
| Temporal decay (per-genre) | **Yes** | None | None | None | None | None | None | None |
| Tiered loading (L0/L1/L2) | **Token-budgeted** | None | None | None | None | None | None | None |
| Session management (OTP) | **Supervised** | None | None | None | None | None | None | None |
| Local-first (no cloud) | **Yes** | Plugin | Yes | No | No | No | Yes | No |
| Fault-tolerant runtime | **Elixir/OTP** | VSCode ext | Electron | Cloud | Cloud | Electron | Electron | Cloud |

### What We Took From Each

| Competitor | What They Do Well | What We Stole | What We Do Better |
|-----------|-------------------|---------------|-------------------|
| **Ars Contexta** | 6R pipeline (read, reflect, remember, rethink, reweave, verify). Elegant conceptual framework. | Reweave (backward pass), Health diagnostics, Reflect (co-occurrence), Remember (friction capture), Rethink (evidence synthesis), Verify (fidelity test), Triangle detection. | Full engine implementation vs VSCode plugin. Auto-classification. Simulation. Hybrid search. Routing. Quality gates. |
| **OpenViking/Sagacity** | Resource/memory/skill typology. Tiered loading. Clean URI scheme. | Tiered loading (L0/L1/L2). Vector search. Memory extraction. Session compression. URI system. Context type system. | Signal Theory classification. Simulation. Genre system. Knowledge graph. Learning loop. |
| **Obsidian** | Plugin ecosystem. UI. Community. Backlinks. | Graph-first thinking (we build a real graph, not just backlinks). | Automation. Search quality. No manual tagging. Classification. Simulation. |
| **Mem.ai** | AI-first memory. Auto-organization. | Memory extraction patterns. Auto-categorization concept. | 5-dimension classification vs flat tags. Simulation. Quality gates. Full OS vs note app. |
| **Roam Research** | Bidirectional linking. Block-level outliner. | Graph-first philosophy. | Engine intelligence. Hybrid search. Classification. Not just links — typed, weighted, reasoned graph. |
| **Logseq** | Open-source. Local-first. Graph view. | Local-first philosophy. | Engine intelligence. Hybrid search. Not just storage — classification + routing + learning. |
| **Notion** | Collaboration UX. Database views. AI features. | Nothing. Different paradigm entirely. | Privacy. Classification depth. Simulation. No vendor lock-in. |
| **Reflect** | Clean AI-enhanced notes. | Nothing. Minimal overlap. | Full cognitive OS vs note app. Graph. Simulation. Genre system. |

### Positioning Statement

> **OptimalOS is NOT a note-taking app.**
>
> It is a cognitive operating system — an externalized decision tree library with an intelligent engine that classifies, routes, stores, searches, learns from, and maintains every piece of information.
>
> No other system combines: auto-classification on 5 dimensions + simulation via MCTS + hybrid search (BM25 + vector + RRF + graph) + 11 quality failure mode detectors + knowledge graph with OWL reasoning + genre-adaptive composition + SICA self-improvement loop + health diagnostics + backward-pass reweaving + friction capture + evidence synthesis.
>
> Together. In a single, fault-tolerant Elixir/OTP system. Running locally. No cloud. No vendor lock-in.
>
> The synthesized optimal system.

---

## Technical Reference

### Configuration

```elixir
# config/config.exs
config :optimal_engine,
  root: "/path/to/OptimalOS",
  db_path: "/path/to/OptimalOS/.system/index.db",
  topology_path: "/path/to/OptimalOS/engine/topology.yaml"

config :optimal_engine, :ollama,
  host: "http://localhost:11434",
  embed_model: "nomic-embed-text",
  generate_model: "qwen3:8b",
  timeout_ms: 30_000

config :optimal_engine, :hybrid_search,
  vector_enabled: true,
  alpha: 0.6  # BM25 weight (1-alpha = vector weight)
```

### Graceful Degradation

Every module that uses Ollama checks `Ollama.available?()` first and falls back:

| Module | With Ollama | Without Ollama |
|--------|------------|----------------|
| SearchEngine | BM25 + vector + RRF | BM25 + temporal + graph |
| MemoryExtractor | LLM 6-category extraction | Regex pattern matching |
| SemanticProcessor | LLM semantic analysis | Keyword overlap |
| IntentAnalyzer | LLM query expansion | Rule-based expansion |
| GraphAnalyzer | LLM synthesis suggestions | Triangle/hub/cluster only |
| Reflector | LLM relationship classification | Default "related" type |
| Reweaver | LLM diff suggestions | "Potentially outdated" flags |
| VerifyEngine | LLM prediction scoring | Jaccard keyword similarity |
| RememberLoop | LLM category classification | Regex keyword matching |
| RethinkEngine | LLM evidence synthesis | Rule-based grouping |

### Performance Characteristics

```
Search latency:     ~50ms (FTS5 only), ~200ms (hybrid with vectors)
Indexing throughput: ~100 files/second
L0 cache refresh:   Every 5 minutes (configurable)
ETS cache:          500 items, LRU eviction, sub-millisecond reads
SQLite WAL mode:    Concurrent reads during writes
OTP supervision:    Auto-restart on crash, ~100ms recovery
```

---

*Architecture version: 2.0 — Post AC-steal, 42+ modules, 10-layer architecture*
*Generated: 2026-03-19*
