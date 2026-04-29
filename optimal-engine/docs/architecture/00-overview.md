# The Optimal System: 7-Layer Context Architecture

> The minimum complete architecture for autonomous agent operation.
> Any missing layer produces catalogued failure modes.

## The Onion

```
         ┌─────────────────────────────────────────────────┐
         │  L7  GOVERNANCE  (VSM: viability + identity)    │
         │  ┌─────────────────────────────────────────────┐│
         │  │ L6  FEEDBACK  (single/double/triple loop)   ││
         │  │ ┌─────────────────────────────────────────┐ ││
         │  │ │ L5  DATA  (DIKW + SignalGraph + search) │ ││
         │  │ │ ┌─────────────────────────────────────┐ │ ││
         │  │ │ │ L4  INTERFACE  (tiered disclosure)  │ │ ││
         │  │ │ │ ┌─────────────────────────────────┐ │ │ ││
         │  │ │ │ │ L3  COMPOSITION (genre skeletons)│ │ │ ││
         │  │ │ │ │ ┌─────────────────────────────┐ │ │ │ ││
         │  │ │ │ │ │ L2  SIGNAL  S=(M,G,T,F,W)  │ │ │ │ ││
         │  │ │ │ │ │ ┌─────────────────────────┐ │ │ │ │ ││
         │  │ │ │ │ │ │ L1  NETWORK  (topology) │ │ │ │ │ ││
         │  │ │ │ │ │ └─────────────────────────┘ │ │ │ │ ││
         │  │ │ │ │ └─────────────────────────────┘ │ │ │ ││
         │  │ │ │ └─────────────────────────────────┘ │ │ ││
         │  │ │ └─────────────────────────────────────┘ │ ││
         │  │ └─────────────────────────────────────────┘ ││
         │  └─────────────────────────────────────────────┘│
         └─────────────────────────────────────────────────┘
```

Concentric, not stacked. L1 is the core. Each outer layer wraps and depends on the inner. Governance wraps everything.

---

## Layer Summary

| Layer | Name | Governing Constraint | One-Line Purpose | Existing MIOSA Code |
|-------|------|---------------------|------------------|-------------------|
| **L1** | Network | Shannon (channel capacity) | Who are the nodes, how are they connected, how do signals route between them | `miosa_knowledge` (SPARQL triples for topology) |
| **L2** | Signal | Ashby (requisite variety) | Every piece of data classified as S=(M,G,T,F,W) — no unclassified data exists | `miosa_signal` (CloudEvents + classifier + S/N ratio) |
| **L3** | Composition | Ashby (genre repertoire) | Internal skeleton per genre — what sections, what structure, what granularity | `miosa_signal.Classifier` (genre→structure mapping) |
| **L4** | Interface | Shannon (bandwidth matching) | Progressive disclosure — show L0 first, drill to L1/L2/L3 as needed | `miosa_memory.Injector` + `miosa_knowledge.Context` |
| **L5** | Data | All four constraints | DIKW hierarchy with temporal versioning, hybrid search, decision traces | `miosa_memory` + `miosa_knowledge` + SignalGraph SQLite |
| **L6** | Feedback | Wiener (closed loops) | Did it happen? Was it right? Are we asking the right questions? | `miosa_memory.Learning` (SICA) + `miosa_memory.Episodic` |
| **L7** | Governance | Beer (viable structure) | System 1-5, agent autonomy levels, algedonic bypass, autopoiesis | New (`miosa_context.Governance`) |

---

## How Signals Flow Through All 7 Layers

### INTAKE (signal enters the system)

```
RAW INPUT (transcript, message, document, voice note, data)
    │
    ▼
L1: NETWORK — Identify source node, determine routing path
    │
    ▼
L2: SIGNAL — Classify: Mode? Genre? Type? Format? Structure?
    │         Measure S/N ratio. Detect failure modes.
    ▼
L3: COMPOSITION — Validate structure against genre skeleton.
    │               Extract facts (SPO triples). Extract entities.
    ▼
L5: DATA — Store in SignalGraph (SQLite + FTS5 + graph).
    │        Create temporal version. Log decision trace.
    │        Index for search. Compute tier summaries (L0/L1/L2).
    ▼
L6: FEEDBACK — Did the intake succeed? Single-loop check.
    │            SICA observes the mutation for pattern learning.
    ▼
L7: GOVERNANCE — Does this signal affect viability?
                  If algedonic trigger → bypass to System 5.
```

### RETRIEVAL (agent or human needs context)

```
QUERY ("I need context about AI Masters sales funnel")
    │
    ▼
L7: GOVERNANCE — Is this query authorized? Agent autonomy check.
    │
    ▼
L4: INTERFACE — Determine token budget. Start with L0 tier.
    │
    ▼
L5: DATA — Hybrid search:
    │        1. BM25 via FTS5 (lexical match)
    │        2. SPARQL graph traversal (relationship match)
    │        3. MCTS tree search (optimal expansion)
    │        4. Temporal decay scoring (recency)
    │        5. Reciprocal Rank Fusion across all modes
    │
    ▼
L4: INTERFACE — Assemble context within budget.
    │            L0 (2K tokens) → L1 (10K) → L2 (50K) as needed.
    ▼
L3: COMPOSITION — Format output for receiver bandwidth.
    │               Genre-appropriate structure.
    ▼
L2: SIGNAL — Classify the OUTPUT signal. Measure S/N.
    │
    ▼
L1: NETWORK — Route to destination (human terminal, agent context, etc.)
    │
    ▼
L6: FEEDBACK — Was the context useful? Close the loop.
```

### CREATION (operator needs to produce a signal)

```
INTENT ("I need a sales doc + video for Robert about AI Masters")
    │
    ▼
L2: SIGNAL — Classify the needed output:
    │         Mode: linguistic + visual
    │         Genre: brief (sales)
    │         Type: direct (compels action)
    │         Format: document + video
    │         Structure: sales_brief skeleton
    │
    ▼
L5: DATA — Pull relevant context:
    │        - AI Masters operation context
    │        - Bob's role and relationship
    │        - Sales funnel status and targets
    │        - Past briefs (genre: brief) for pattern
    │        - Revenue targets from finance
    │
    ▼
L3: COMPOSITION — Apply sales_brief skeleton:
    │               1. Objective
    │               2. Audience (Robert's bandwidth + genre competence)
    │               3. Key Messages
    │               4. Call to Action
    │               5. Supporting Materials (video companion)
    │
    ▼
L4: INTERFACE — Match output to Robert's decoding capacity.
    │            Robert = salesperson → brief genre, not spec genre.
    │
    ▼
L1: NETWORK — Route: Alice → Robert (channel: email + Slack)
    │
    ▼
L6: FEEDBACK — Did Robert act on it? Track response.
    │            Single-loop: Was it received?
    │            Double-loop: Did it advance the funnel?
    ▼
L7: GOVERNANCE — Log decision trace. Update funnel status.
```

---

## Detailed Layer Specifications

See individual docs:
- [Layer 1: Network](01-network.md)
- [Layer 2: Signal](02-signal.md)
- [Layer 3: Composition](03-composition.md)
- [Layer 4: Interface](04-interface.md)
- [Layer 5: Data](05-data.md)
- [Layer 6: Feedback](06-feedback.md)
- [Layer 7: Governance](07-governance.md)

## Supporting Documentation

- [Taxonomy: Hierarchy Primitives](../taxonomy/hierarchy.md)
- [Taxonomy: Glossary](../taxonomy/glossary.md)
- [Taxonomy: Genre Catalogue](../taxonomy/genres.md)
- [Operations: Intake Pipeline](../operations/intake-pipeline.md)
- [Operations: Search & Retrieval](../operations/search-retrieval.md)
- [Operations: Auto-Routing](../operations/auto-routing.md)
- [Guide: Quick Start](../guides/quick-start.md)

## Existing Code Assets

| MIOSA App | What It Does | Layers It Serves |
|-----------|-------------|-----------------|
| `miosa_signal` | Signal envelope, auto-classifier, S/N measurement, 11 failure modes | L2, L3 |
| `miosa_knowledge` | SPARQL engine, OWL 2 RL reasoner, dictionary encoding, agent context injection | L1, L5 |
| `miosa_memory` | Three-store memory, Cortex synthesis, SICA learning, Injector, Taxonomy, Search | L4, L5, L6 |
| `miosa_context` (NEW) | Composition layer — SQLite + FTS5, temporal versioning, tiered loading, MCTS | L3, L4, L5, L7 |
