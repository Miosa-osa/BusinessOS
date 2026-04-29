# Layer Model Bridge — Theory vs Implementation

> Two models describe OptimalOS. One is the theory (WHY). One is the code (HOW).
> This document maps them to each other so they stop competing.

## The Two Models

### Conceptual Model (7 Layers — Beer's VSM)

The **concentric onion** from Signal Theory + Viable System Model.
L1 is the operational core. L7 wraps everything with governance.

```
L7  Governance ─── Viability, identity, policy (Beer)
L6  Feedback ───── Self-correction loops (Wiener)
L5  Data ───────── Storage, retrieval, knowledge graph (All)
L4  Interface ──── Tiered disclosure, bandwidth matching (Shannon)
L3  Composition ── Genre skeletons, structure imposition (Ashby)
L2  Signal ─────── S=(M,G,T,F,W) classification (Ashby)
L1  Network ────── Topology, routing, nodes (Shannon)
```

**Read this as:** "Operations (L1) happen inside a governance shell (L7)."
This is Beer's model — the inner layers produce, the outer layers regulate.

### Engine Model (10 Layers — Implementation Stack)

The **bottom-up stack** matching the Elixir/OTP codebase.
L0 is the data foundation. L9 is the OS surface.

```
L9  Operating System ── Health, reweaving, verification
L8  Learning ────────── SICA, episodic memory, cortex
L7  Agent Interface ─── Composer, genre encoding, receiver matching
L6  Processing ──────── Intake pipeline, indexing, extraction
L5  Intelligence ────── MCTS, simulation, semantic processing
L4  Knowledge Graph ─── Entities, edges, triangles, clusters
L3  Search ──────────── FTS5, vector, RRF fusion, temporal decay
L2  Routing ─────────── Topology rules, cross-references
L1  Classification ──── S=(M,G,T,F,W), entity extraction
L0  Data Storage ────── SQLite, ETS cache, vectors
```

**Read this as:** "Data at the bottom, OS at the top." Standard software stack.

## How They Map

| Conceptual | Engine | What It Does |
|-----------|--------|-------------|
| L1 Network | L2 Routing | Topology, node definitions, signal routing |
| L2 Signal | L1 Classification | S=(M,G,T,F,W) classification |
| L3 Composition | L7 Agent Interface | Genre skeletons, receiver-adaptive encoding |
| L4 Interface | L3 Search + L6 Processing | Tiered disclosure, context assembly |
| L5 Data | L0 Data + L4 Knowledge Graph | SQLite, FTS5, graph, vectors |
| L6 Feedback | L8 Learning | SICA, episodic memory, cortex synthesis |
| L7 Governance | L9 Operating System | Health checks, drift detection, spec verification |
| — | L5 Intelligence | MCTS, simulation (no direct conceptual equivalent) |

## ROM vs RAM

The system has two memory modes, matching Oscar's "pre-computed decision tree library" model:

### ROM (OM — Operational Memory)
Slow-changing. The library. What you've already figured out.

| Artifact | Decay | Examples |
|----------|-------|---------|
| context.md | 180+ days | Facts about people, entities, structures |
| kernel.yaml | Years | 15 primitives, 4 constraints, 6 principles |
| topology.yaml | Months | Node definitions, routing rules |
| Genre templates | Years | 191 templates across 18 categories |
| Playbooks | Months | Pre-computed decision trees |
| Architecture docs | Months | Layer specs, ADRs |
| Specs (.spec.md) | Months | Requirement contracts |

### RAM (AM — Active Memory)
Fast-changing. What's happening now. Decays quickly.

| Artifact | Decay | Examples |
|----------|-------|---------|
| signal.md | 7 days | Weekly status, priorities, blockers |
| signals/*.md | 7 days | Dated episodic events |
| today.md | 1 day | Daily cockpit |
| weekly-dump.md | 7 days | Monday brain dump |
| ETS cache | Minutes | Hot contexts, 500 items LRU |
| Session state | Hours | Active conversation context |
| L0 cache | 30 min | Always-loaded inventory |

### The Bridge
The engine's **half-life decay rates** already encode this distinction:
- `spec: 180d`, `decision: 365d` → ROM
- `signal: 7d`, `note: 3d`, `standup: 7d` → RAM

The intake pipeline auto-classifies: persistent fact → context.md (ROM).
Temporal event → signal file (RAM). The decay rate determines how long
it stays relevant in search results.

## When to Use Which Model

| Situation | Use This Model |
|-----------|---------------|
| Explaining the theory to someone | 7-layer conceptual |
| Writing architecture docs | 7-layer conceptual |
| Debugging the engine | 10-layer engine |
| Adding a new module | 10-layer engine (which layer does it belong to?) |
| Discussing with Oscar | 7-layer conceptual (he thinks in Beer's VSM) |
| Discussing with Carol | 10-layer engine (he thinks in code) |
| Writing specs | Either — but reference the engine layer for `surface` paths |
