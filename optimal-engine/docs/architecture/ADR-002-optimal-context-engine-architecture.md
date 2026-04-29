---
signal:
  mode: linguistic
  genre: adr
  type: decide
  format: markdown
  structure: adr_template
  sn_ratio: 1.0
  audience: [roberto, javaris, osa-agent]
  intent: "Architecture Decision Record — Optimal Context Engine package structure"
---

# ADR-002: Optimal Context Engine — Package Architecture

## Status: Accepted
## Date: 2026-03-18

---

## Context

`tools/optimal.py` is a 671-line monolithic script that mixes six distinct concerns:
YAML parsing, SQLite management, signal classification, entity extraction, hybrid search,
tier generation, and CLI dispatch. It works for Phase 0 (single user, filesystem only)
but cannot be extended to Phase 1 (vector embeddings) or Phase 2 (graph/OWL) without
rewriting the whole file.

The system has a clear conceptual architecture (7 layers, Signal Theory, topology.yaml,
143 genre catalogue) but the implementation does not reflect it. The gap between the
documented architecture and the code is a maintenance tax that compounds with every
feature added to the monolith.

Three constraints govern the design:
1. No external dependencies for core (stdlib only: `sqlite3`, `pathlib`, `json`, `hashlib`, `re`, `math`)
2. Must run as CLI (`python3 optimal/`) AND be importable as a library (`from optimal import Engine`)
3. Must stay simple enough that Alice can read the directory structure and understand
   what each folder does without a guided tour

---

## Decision

Decompose `optimal.py` into a Python package at `tools/optimal/` organized by **capability**
(what each module does), not by layer (the conceptual 7-layer model is the architecture
of the *system*, not the architecture of this *tool*). Each capability module maps to exactly
one port in a hexagonal/ports-and-adapters layout.

The package has eight modules plus a CLI entry point. This is the bicycle — not the 747.

---

## Consequences

### Positive
- Each module can be tested in isolation (no more "test the whole 671-line script")
- Phase 1 (embeddings) = add `search/vector.py` and plug into `SearchEngine` without touching anything else
- Phase 2 (graph) = add `graph/` package and plug into `Engine` via the same port interface
- `topology.yaml` becomes the single source of truth consumed by `config/topology.py` — no more hardcoded `NODE_MAP` dicts
- A developer can open the package and immediately understand where to look for any feature
- `Engine` class provides a stable API surface; CLI is just a thin wrapper over it

### Negative
- Initial migration cost: ~4 hours to split the existing file correctly
- Eight files instead of one (acceptable complexity trade-off for the extensibility gained)
- Must maintain `__init__.py` re-exports carefully so `from optimal import Engine` keeps working

### Neutral
- SQLite + FTS5 stays as-is. No schema changes in this ADR.
- The existing YAML frontmatter format is unchanged.
- All existing CLI commands (`index`, `search`, `l0`, `stats`, `ingest`) continue to work.

---

## Alternatives Considered

### Option A: Keep the monolith, add docstring section headers
Split into clearly labeled regions within the single file.
Rejected because: No import boundaries means no testability. Adding Phase 1 embeddings
means a new 200-line section in an already-large file. The problem recurs.

### Option B: Mirror the 7-layer architecture docs as Python packages
`network/`, `signal/`, `composition/`, `interface/`, `data/`, `feedback/`, `governance/`
Rejected because: The 7 layers are a conceptual system model, not deployment boundaries
for a CLI tool. A developer looking for "how does search work?" would have to understand
which layer search belongs to before finding the code. Capability organization maps more
directly to the user's mental model of the tool.

### Option C: Microservices / FastAPI server from day one
Each capability as a separate HTTP service.
Rejected because: Massive over-engineering for Phase 0. Alice runs this as a local CLI.
Adding HTTP, serialization, and service discovery triples the complexity with zero benefit
until Phase 1+ when the engine needs to serve multiple clients.

---

## References

- `tools/optimal.py` — current monolith being replaced
- `topology.yaml` — single source of truth for nodes and routing rules
- `docs/architecture/02-signal.md` — Signal Theory classification pipeline
- `docs/taxonomy/genres.md` — 143 genre catalogue (Ashby constraint)
- Competitor analysis: OpenViking, Mem0, LangChain (see `docs/competitors/`)
