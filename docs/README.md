# Optimal Engine — Documentation

This index is the entry point. Follow the path that matches what you need.

## If you are new here

1. [Getting started](guides/getting-started.md) — clone, compile, first ingest + search.
2. [Signal Theory](concepts/signal-theory.md) — the theoretical foundation. 10-minute read.
3. [System overview](architecture/system-overview.md) — how the pieces fit together.

## Architecture (how it's built)

- [00 — Overview](architecture/00-overview.md)
- [01 — Network layer](architecture/01-network.md)
- [02 — Signal layer](architecture/02-signal.md)
- [03 — Composition layer](architecture/03-composition.md)
- [04 — Interface layer](architecture/04-interface.md)
- [05 — Data layer](architecture/05-data.md)
- [06 — Feedback layer](architecture/06-feedback.md)
- [07 — Governance layer](architecture/07-governance.md)
- [FULL-SYSTEM-ARCHITECTURE](architecture/FULL-SYSTEM-ARCHITECTURE.md) — end-to-end walkthrough
- [Optimal Engine package spec](architecture/optimal-engine-package-spec.md)
- [ADR-002: Optimal Context Engine architecture](architecture/ADR-002-optimal-context-engine-architecture.md)
- [Context system](architecture/context-system.md)
- [Layer bridge](architecture/layer-bridge.md)
- [Workspace protocol](architecture/workspace-protocol.md)
- [System overview](architecture/system-overview.md)
- [Oscar / Atlas system](architecture/jordan-jarvis-system.md) — cognitive model that informed the design
- [arscontexta extraction](architecture/arscontexta-extraction.md)

## Concepts (why it's built this way)

- [Signal Theory](concepts/signal-theory.md) — `S=(Mode, Genre, Type, Format, Structure)` with four constraints
- [Methodology](concepts/methodology.md) — the working loops (boot, operate, build, review)
- [Three spaces](concepts/three-spaces.md) — input / signal / persistence separation
- [Failure modes](concepts/failure-modes.md) — the 11 Shannon / Ashby / Beer / Wiener violations the engine guards against
- [Infinite context framework](concepts/infinite-context-framework.md) — tiered loading, compaction, recall
- [Kernel primitives](concepts/kernel.yaml) — the invariant vocabulary

## Reference (what each piece does)

- [Data model](reference/data-model.md) — schemas, relationships, URI conventions
- [Search architecture](reference/search-architecture.md) — hybrid BM25 + temporal + graph boost
- [Session lifecycle](reference/session-lifecycle.md) — boot → operate → shutdown
- [Hooks](reference/hooks.md) — extension points
- [Operations spec](reference/operations-spec.md) — the full operational contract
- [Vocabulary](reference/vocabulary.md) — terminology reference
- [Components](reference/components.md) — catalog of engine components
- [Node template](reference/node-template.md) — the anatomy of a node

## Guides (how to use it)

- [Getting started](guides/getting-started.md)
- [Mix tasks](guides/mix-tasks.md) — the 25 `mix optimal.*` commands, annotated
- [Writing guide](guides/writing-guide.md) — conventions for Signal-correct output

## Research

- [Signal composition layer](research/signal-composition-layer.md) — exploratory work on the composition primitives
