# Layer 5: Data

> Store all signals with full DIKW transformation traceability.
> Every piece of data can be traced from raw observation through to wisdom.

---

## Purpose

Layer 5 is the persistence substrate for the entire system. It answers one question at every access point: "What do we know, how certain are we, and when was it true?"

No signal enters the system without being stored. No agent operates without drawing from this layer. No decision is made without its trace being written here. The three stores — Markdown files, SignalGraph SQLite database, and MIOSA Knowledge Graph — compose into a single coherent memory that is human-readable, machine-searchable, and temporally complete.

---

## Governing Constraints

All four Signal Theory constraints apply simultaneously at this layer:

| Constraint | Application |
|------------|-------------|
| **Shannon (channel capacity)** | Storage must not exceed the retrieval channel's bandwidth. Tiered summaries (L0/L1/L2) pre-compute compressed representations so agents never load more than their token budget allows. |
| **Ashby (requisite variety)** | The storage schema must have enough structural variety to represent every signal type the system will encounter — facts, decisions, entities, relationships, sessions, patterns, temporal chains. |
| **Beer (viable structure)** | Every store must be coherently structured at every scale. Documents, entities, facts, and edges are all first-class citizens with defined relationships. No orphaned data. |
| **Wiener (feedback loop)** | Data written by agents must feed back into future agent context. The decision log is the feedback mechanism: every mutation records why it happened, closing the write→read loop. |

---

## DIKW Hierarchy with Explicit Transformations

```
DATA (raw observations — meeting transcripts, chat messages, voice notes)
│
│  Transform: Classification (Layer 2 auto-classifier)
│  Input:  Unstructured text / audio / events
│  Output: Signal envelope S=(M,G,T,F,W) + intent field
│
▼
INFORMATION (classified signals with context)
│
│  Transform: Extraction (Layer 3 fact extractor → SPO triples)
│  Input:  Classified document with entities list
│  Output: Subject-Predicate-Object triples → facts table + edges table
│
▼
KNOWLEDGE (connected facts in the graph)
│
│  Transform: Inference (OWL 2 RL reasoner materializes implicit relationships)
│  Input:  Explicit SPO triples in the knowledge graph
│  Output: Derived relationships, transitive closures, class memberships
│
▼
WISDOM (synthesized understanding with decision traces)

   Transform: Cortex synthesis + SICA learning + decision patterns
   Input:  Knowledge graph + decision_log + temporal chains
   Output: Actionable understanding — what to do next, what patterns hold,
           what decisions are due for review
```

Each transformation is traceable. A wisdom output (Cortex synthesis) can be walked back to the knowledge it drew from, the facts that composed it, the documents that sourced those facts, and the raw observation that started the chain.

---

## Storage Architecture

Three stores that compose into a single logical memory. Each store has a distinct role; none is redundant.

### 1. Markdown Files (Source of Truth, Human-Readable)

```
~/.context/
├── entities/       — Entity profiles (people, orgs, tools, concepts)
├── operations/     — Operation context (projects, programs, ventures)
├── decisions/      — ADRs and decision logs
├── patterns/       — Learned behavioral and architectural patterns
├── sessions/       — Session handoff notes
└── knowledge/      — Domain knowledge articles
```

Every file is Signal-encoded Markdown with YAML frontmatter carrying all five S=(M,G,T,F,W) dimensions plus temporal validity fields:

```markdown
---
id: doc_a1b2c3d4
type: decision                    # Genre (G)
mode: linguistic                  # Mode (M) — linguistic/visual/code/data
act: commit                       # Type (T) — direct/inform/commit/decide/express
format: markdown                  # Format (F)
structure: adr                    # Structure (W) — adr/spec/note/log/profile

created: 2026-03-16T10:00:00Z
modified: 2026-03-16T14:30:00Z
valid_from: 2026-03-16T10:00:00Z
valid_until: null                 # null = still valid
supersedes: doc_x9y8z7w6          # temporal chain link

entities:
  - "[[Alice]]"
  - "[[OptimalOS]]"
  - "[[Context System]]"

tags: [architecture, memory, decision]

signal:
  intent: "Choose storage architecture for personal context OS"
  confidence: 0.9
  noise_level: low
  audience: [agent, human]
---
```

These files are the canonical record. The SQLite database is an index built from them. The SPARQL store is a relationship layer extracted from them. If either derived store is lost, it can be rebuilt from the Markdown files alone.

**Properties:**
- Git-versionable: `git add ~/.context && git commit`
- Obsidian-compatible: wiki-links render as graph edges natively
- Editor-agnostic: readable in any text editor without tooling
- Agent-writable: agents produce Signal-encoded Markdown as their output format

---

### 2. SignalGraph DB (SQLite + Extensions)

The SQLite database indexes the Markdown files for machine-speed retrieval, stores the extracted knowledge graph (entities + edges + facts), and maintains the append-only decision audit trail.

#### Full Schema

```sql
-- Core document storage (markdown files indexed)
CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,          -- filesystem path
    title TEXT,
    content TEXT,                       -- full markdown content

    -- Signal dimensions (S = M, G, T, F, W)
    mode TEXT DEFAULT 'linguistic',     -- M: linguistic/visual/code/data
    genre TEXT,                         -- G: decision/pattern/entity/session/note
    act TEXT,                           -- T: direct/inform/commit/decide/express
    format TEXT DEFAULT 'markdown',     -- F: markdown/json/yaml/code
    structure TEXT,                     -- W: adr/spec/note/log/profile

    -- Temporal validity (bi-temporal model)
    created_at TEXT NOT NULL,
    modified_at TEXT NOT NULL,
    valid_from TEXT,
    valid_until TEXT,                   -- NULL = still valid
    supersedes TEXT,                    -- FK to documents.id (temporal chain)

    -- Intent and signal quality
    intent TEXT,                        -- WHY this document exists
    confidence REAL DEFAULT 1.0,
    audience TEXT DEFAULT 'both',       -- human/agent/both

    -- Tiered loading (pre-computed compressed representations)
    l0_summary TEXT,                    -- ~100 tokens: one-liner
    l1_description TEXT,                -- ~500 tokens: key facts + links
    l2_excerpt TEXT                     -- ~2000 tokens: actionable detail
);

-- FTS5 virtual table for BM25 full-text search
CREATE VIRTUAL TABLE documents_fts USING fts5(
    title, content, intent, l0_summary, l1_description,
    content='documents',
    content_rowid='rowid',
    tokenize='porter unicode61'
);

-- Entities: people, projects, concepts, tools, organizations
CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,                 -- person/project/concept/tool/org
    canonical_name TEXT,                -- normalized/deduplicated form
    description TEXT,
    first_seen TEXT NOT NULL,
    last_seen TEXT NOT NULL,
    mention_count INTEGER DEFAULT 1,
    properties JSON                     -- flexible key-value attributes
);

-- Graph edges: typed relationships between entities and documents
CREATE TABLE edges (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,            -- entity or document id
    source_type TEXT NOT NULL,          -- 'entity' or 'document'
    target_id TEXT NOT NULL,
    target_type TEXT NOT NULL,
    relation TEXT NOT NULL,             -- owns/depends_on/supersedes/mentions/etc
    weight REAL DEFAULT 1.0,

    -- Temporal validity
    valid_from TEXT NOT NULL,
    valid_until TEXT,                   -- NULL = still valid

    -- Decision trace on the relationship itself
    reason TEXT,                        -- WHY this relationship exists
    context TEXT,                       -- situational context at creation time

    metadata JSON,
    created_at TEXT NOT NULL
);

-- Atomic facts: extracted Subject-Predicate-Object triples
CREATE TABLE facts (
    id TEXT PRIMARY KEY,
    subject_entity TEXT NOT NULL,       -- FK to entities.id
    predicate TEXT NOT NULL,            -- "works_at" / "prefers" / "decided"
    object TEXT NOT NULL,               -- literal value or FK to entities.id
    object_type TEXT,                   -- literal/entity/date/number

    source_document TEXT NOT NULL,      -- FK to documents.id (provenance)
    confidence REAL DEFAULT 1.0,

    -- Temporal validity
    valid_from TEXT NOT NULL,
    valid_until TEXT,                   -- NULL = still valid

    -- Decision trace
    reason TEXT,                        -- WHY this fact was recorded
    supersedes TEXT                     -- FK to facts.id (temporal chain)
);

-- Decision log: append-only audit trail for all mutations
CREATE TABLE decision_log (
    id TEXT PRIMARY KEY,
    timestamp TEXT NOT NULL,
    actor TEXT NOT NULL,                -- 'human' or agent name
    action TEXT NOT NULL,               -- create/update/supersede/invalidate
    target_type TEXT NOT NULL,          -- document/entity/edge/fact
    target_id TEXT NOT NULL,

    -- The decision trace
    what_changed TEXT NOT NULL,
    why TEXT,                           -- reasoning behind the change
    alternatives_considered JSON,       -- what else was evaluated
    context JSON,                       -- situational factors at decision time
    outcome_optimized TEXT              -- what goal this mutation served
);

-- Phase 3: Vector embeddings via sqlite-vec extension
-- Commented out until Phase 3 implementation
-- CREATE VIRTUAL TABLE vec_documents USING vec0(
--     id TEXT PRIMARY KEY,
--     embedding FLOAT[384]            -- all-MiniLM-L6-v2 dimensions
-- );

-- Tiered loading cache: pre-computed context packages per scope
CREATE TABLE context_tiers (
    scope TEXT NOT NULL,                -- 'global', project name, entity name
    tier INTEGER NOT NULL,              -- 0, 1, or 2
    content TEXT NOT NULL,              -- pre-assembled context string
    token_count INTEGER NOT NULL,
    last_computed TEXT NOT NULL,
    PRIMARY KEY (scope, tier)
);
```

#### Key Design Decisions

**Why SQLite, not PostgreSQL or Neo4j.** Zero deployment overhead. Single-file database — backup is `cp ~/.context/signal.db ~/backup/`. FTS5 gives BM25 for free. sqlite-vec adds vectors as an extension in Phase 3. The graph traversal requirement is met by the `edges` table with recursive SQL queries, not a dedicated graph engine.

**Why append-only facts and edges.** Following the bi-temporal model from Zep/Graphiti: `valid_from` + `valid_until` intervals mean no fact is ever destroyed. Historical state is always queryable. `valid_until = NULL` means currently valid.

**Why decision_log is separate.** The decision log is write-only by design. It answers the question HydraDB's Cortex paper identifies as critical: not just WHAT changed, but WHY, WHAT was considered, and WHAT goal it served. This is the Wiener constraint applied to storage — every mutation closes a loop by recording its own rationale.

---

### 3. MIOSA Knowledge Graph (SPARQL + OWL)

The knowledge graph stores relationships as RDF triples and runs OWL 2 RL inference to materialize implicit facts. It serves as the semantic reasoning layer above the SQLite structural layer.

#### Triple Structure

Every relationship is a triple: `(subject, predicate, object)`

```
("Alice", "works_on", "OptimalOS")
("OptimalOS", "uses", "SignalGraph")
("SignalGraph", "implements", "DIKW hierarchy")
("Alice", "decided", "ADR-001")
("ADR-001", "supersedes", "ADR-000")
```

#### OWL 2 RL Inference

The OWL 2 RL profile runs on the materialized triple store to derive implicit relationships:

```
EXPLICIT:  Alice works_on OptimalOS
EXPLICIT:  OptimalOS is_a Project
INFERRED:  Alice works_on Project   (class inference)

EXPLICIT:  ADR-001 supersedes ADR-000
EXPLICIT:  supersedes is transitive
INFERRED:  If ADR-002 supersedes ADR-001, then ADR-002 supersedes ADR-000
```

This eliminates the need to explicitly record every derived relationship. The reasoner materializes the transitive closure at query time.

#### Agent Context Injection

The `for_agent/2` predicate scopes knowledge to a specific agent:

```prolog
for_agent(agent_orchestrator, context_block_id).
for_agent(agent_debugger, context_block_id).
```

When an agent requests context, the knowledge graph returns only triples scoped to that agent plus globally-scoped triples. This enforces the Layer 7 Governance principle of need-to-know access without a separate permission system.

#### SPARQL Retrieval

The `miosa_knowledge` app exposes a SPARQL endpoint for complex graph queries that SQLite's edge table cannot efficiently express:

```sparql
# All entities Alice is connected to within 2 hops
SELECT DISTINCT ?connected_entity WHERE {
  <entity:roberto_luna> ?rel1 ?hop1 .
  ?hop1 ?rel2 ?connected_entity .
  FILTER(?connected_entity != <entity:roberto_luna>)
}

# Everything that happened in the last 30 days
SELECT ?subject ?predicate ?object WHERE {
  ?subject ?predicate ?object .
  ?subject :valid_from ?t .
  FILTER(?t > "2026-02-17"^^xsd:dateTime)
}
```

---

## Temporal Versioning (Append-Only)

Every mutation to the knowledge base follows a four-step protocol:

```
1. APPEND   — Write new state as a new record (never UPDATE in place)
2. INVALIDATE — Set valid_until on the old record to now()
3. LINK     — Set supersedes on the new record to point to the old record
4. LOG      — Write to decision_log: what changed, why, alternatives, context
```

#### Example: Preference Change

```
BEFORE:
  fact_001: Alice → prefers → Elixir/Phoenix
            valid_from: 2025-06-01, valid_until: NULL

MUTATION EVENT:
  New requirement: compute-engine needs syscall-heavy performance work.

AFTER:
  fact_001: Alice → prefers → Elixir/Phoenix
            valid_from: 2025-06-01, valid_until: 2026-03-16T09:00:00Z

  fact_002: Alice → prefers → Elixir/Phoenix + Go
            valid_from: 2026-03-16T09:00:00Z, valid_until: NULL
            supersedes: fact_001
            reason: "Added Go for compute-engine performance requirements"
            context: "Building VM orchestration layer; Go better for syscall-heavy work"

  decision_log entry:
    actor: human
    action: supersede
    target_type: fact
    target_id: fact_001
    what_changed: "Preferred stack updated from Elixir-only to Elixir+Go"
    why: "compute-engine architecture requires low-level syscall control"
    alternatives_considered: ["Rust", "C", "keep Elixir with NIFs"]
    outcome_optimized: "VM orchestration performance"
```

#### Time-Travel Queries

```sql
-- What did Alice prefer in January 2026?
SELECT * FROM facts
WHERE subject_entity = 'roberto_luna'
  AND predicate = 'prefers'
  AND valid_from <= '2026-01-15'
  AND (valid_until IS NULL OR valid_until > '2026-01-15');

-- Temporal chain of a decision
WITH RECURSIVE chain AS (
    SELECT id, title, created_at, supersedes
    FROM documents WHERE id = 'current_decision_id'
    UNION ALL
    SELECT d.id, d.title, d.created_at, d.supersedes
    FROM documents d JOIN chain c ON d.id = c.supersedes
)
SELECT * FROM chain ORDER BY created_at DESC;

-- All changes made by a specific agent in a session
SELECT * FROM decision_log
WHERE actor = 'agent_orchestrator'
  AND timestamp >= '2026-03-16T00:00:00Z'
ORDER BY timestamp DESC;
```

---

## Hybrid Search

Five search modes fuse into a single ranked result set. No single mode is sufficient; the fusion is what produces high-accuracy retrieval without LLM calls at query time.

| Mode | Technique | Weight | Implementation |
|------|-----------|--------|----------------|
| Lexical | BM25 via FTS5 | 0.25 | SQLite FTS5 (`documents_fts`) |
| Graph | SPARQL traversal + 1-hop neighbor boost | 0.30 | `miosa_knowledge` |
| Semantic | Vector cosine similarity | 0.20 | `sqlite-vec` (Phase 3) |
| MCTS | Monte Carlo tree search over knowledge graph | 0.15 | `miosa_context` (new) |
| Temporal | Recency decay + validity filter | 0.10 | `miosa_memory.Episodic` |

**Graph gets the highest weight** because relationships outperform similarity. Knowing that Document A is linked to Entity B which is linked to the query entity is more precise than knowing Document A is semantically similar to the query string.

#### Fusion Formula

Reciprocal Rank Fusion across all active search modes:

```
score(doc) = sum over all modes i of: 1 / (60 + rank_i(doc))
```

Where `rank_i(doc)` is the position of `doc` in the ranked list produced by mode `i`. The constant 60 dampens the influence of high ranks without eliminating low-ranked results entirely (standard RRF parameter from the Cormack et al. 2009 paper).

#### Phase Implementation

- **Phase 1 (now):** Lexical + Temporal. BM25 via FTS5 + recency scoring from `miosa_memory.Episodic`.
- **Phase 2:** Add Graph. SPARQL traversal from `miosa_knowledge` + 1-hop neighbor boost.
- **Phase 3:** Add Semantic. sqlite-vec extension + local embeddings via `all-MiniLM-L6-v2` running on llama.cpp.
- **Phase 4:** Add MCTS. Monte Carlo tree search over the knowledge graph for exploratory queries that benefit from structured expansion.

#### BM25 Query (Phase 1)

```sql
SELECT d.id, d.title, d.l1_description,
       bm25(documents_fts) AS score
FROM documents_fts
JOIN documents d ON d.rowid = documents_fts.rowid
WHERE documents_fts MATCH 'context memory architecture'
  AND d.valid_until IS NULL          -- only current documents
ORDER BY score
LIMIT 20;
```

#### Graph-Boosted Query (Phase 2)

```
1. BM25 search → candidate set C
2. For each doc in C: extract entity mentions
3. For each entity: traverse 1-hop graph neighbors via edges table
4. Boost score for candidates that share entity neighbors with the query entities
5. RRF merge of BM25 rank + graph-boosted rank
```

---

## Existing Code Assets

| Module | Role in Layer 5 |
|--------|----------------|
| `miosa_knowledge` | SPARQL engine, OWL 2 RL reasoner, dictionary encoding, `for_agent/2` context injection |
| `miosa_memory.Store` | Three-store GenServer — coordinates Markdown, SQLite, and SPARQL stores |
| `miosa_memory.Search` | Keyword + recency + importance scoring (Phase 1 search pipeline) |
| `miosa_memory.Index` | ETS inverted keyword index (in-memory layer over SQLite for hot queries) |
| `miosa_memory.Parser` | Parses MEMORY.md entry format into structured records |
| `miosa_memory.Episodic` | Temporal decay scoring — recency weighting for the temporal search mode |
| `miosa_memory.Cortex` | LLM synthesis over retrieved knowledge — the Wisdom layer output |

The `miosa_context` app (new, being built) owns the SQLite schema, FTS5 indexing, temporal versioning protocol, MCTS search, and tiered context assembly. It wraps the existing MIOSA apps and adds the structured persistence layer they currently lack.

---

## Layer Interfaces

**Consumed by:**
- L4 (Interface) — requests assembled context packages at specified token budgets
- L6 (Feedback) — reads decision logs and fact validity to detect contradictions and stale knowledge
- L7 (Governance) — reads decision traces to evaluate agent behavior and system viability

**Produces for:**
- L4: Tiered context packages (L0/L1/L2/L3) pre-computed in `context_tiers` table
- L6: Full mutation history via `decision_log` for SICA pattern learning
- L7: Decision traces and fact temporal chains for viability assessment

**Writes from:**
- L3 (Composition) — extracted SPO triples land in `facts` and `edges`
- L2 (Signal) — classified signal envelope lands in `documents` with Signal dimensions
- L6 (Feedback) — SICA learning outputs land in `patterns/` Markdown files

---

## Related Specifications

- [Layer 4: Interface](04-interface.md) — tiered loading protocol (L0/L1/L2/L3 budgets)
- [Layer 3: Composition](03-composition.md) — fact extraction pipeline that populates this layer
- [Layer 6: Feedback](06-feedback.md) — SICA learning that reads from and writes to this layer
- [Operations: Search and Retrieval](../operations/search-retrieval.md) — hybrid search pipeline detail
- [Source Research: SignalGraph Architecture](../../tasks/context-os-architecture.md)
