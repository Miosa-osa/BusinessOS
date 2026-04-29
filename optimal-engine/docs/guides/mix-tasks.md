# Mix Tasks Reference

All commands are prefixed `mix optimal.*`. Run any command with `--help` for
argument details.

## Ingestion

| Task                  | Purpose                                                                 |
|-----------------------|-------------------------------------------------------------------------|
| `mix optimal.ingest`  | Classify, route, write signal files, and index content                  |
| `mix optimal.intake`  | Interactive multi-line intake from stdin                                |
| `mix optimal.index`   | Full reindex of all markdown files under the root                       |

**Ingest examples:**

```bash
mix optimal.ingest "Customer called about pricing" --genre note
mix optimal.ingest --file notes.md --genre transcript --title "Team sync"
```

## Retrieval

| Task                   | Purpose                                                                 |
|------------------------|-------------------------------------------------------------------------|
| `mix optimal.search`   | Hybrid BM25 + temporal + graph-boosted search                           |
| `mix optimal.read`     | Read a context by `optimal://` URI at a given tier                      |
| `mix optimal.ls`       | List contexts under an `optimal://` URI                                 |
| `mix optimal.l0`       | Print the always-loaded L0 context                                      |
| `mix optimal.assemble` | Build a tiered (L0/L1/L2) context bundle for a topic                    |

**Tier semantics:**
- `l0` — ~100 tokens per context. Abstract. Use for discovery.
- `l1` — ~2K tokens. Summary. The default working tier.
- `full` — complete content. Use only when `l1` is insufficient.

**Examples:**

```bash
mix optimal.search "pricing decision" --limit 5
mix optimal.search "Alice" --node ai-masters
mix optimal.read "optimal://nodes/ai-masters/signals/2026-04-17-ed-pricing.md" --tier l1
mix optimal.assemble "AI Masters pricing"
```

## Graph analysis

| Task                  | Purpose                                                                 |
|-----------------------|-------------------------------------------------------------------------|
| `mix optimal.graph`   | Knowledge graph stats, triangles, clusters, hubs                        |
| `mix optimal.reflect` | Find missing edges from entity co-occurrences                           |
| `mix optimal.reweave` | Find stale contexts on a topic + suggest updates                        |
| `mix optimal.simulate`| Run a "what if" scenario through the graph                              |
| `mix optimal.impact`  | Impact analysis for an entity or node                                   |

**Examples:**

```bash
mix optimal.graph triangles       # A→B, A→C exist but B→C missing
mix optimal.graph clusters        # isolated knowledge islands
mix optimal.graph hubs            # most-connected entities (>2σ degree)
mix optimal.reflect --min 3       # require 3+ co-occurrences
mix optimal.reweave "pricing" --days 60
```

## Learning loop

| Task                   | Purpose                                                                 |
|------------------------|-------------------------------------------------------------------------|
| `mix optimal.remember` | Store observations; mine friction patterns                              |
| `mix optimal.rethink`  | Synthesize observations into actionable knowledge                       |
| `mix optimal.knowledge`| Knowledge graph + SICA learning operations                              |

**Examples:**

```bash
mix optimal.remember "always check duplicates before inserting"
mix optimal.remember --contextual   # scan recent signals for friction patterns
mix optimal.remember --list         # see all stored observations
mix optimal.remember --escalations  # categories ready for rethink
mix optimal.rethink "process"
mix optimal.knowledge metrics
```

## Health & verification

| Task                  | Purpose                                                                 |
|-----------------------|-------------------------------------------------------------------------|
| `mix optimal.health`  | 10 diagnostic checks (orphans, drift, dupes, fidelity, etc.)            |
| `mix optimal.verify`  | Cold-read test of L0 abstract fidelity                                  |
| `mix optimal.stats`   | Store statistics (counts, sizes, token budgets)                         |

**Examples:**

```bash
mix optimal.health
mix optimal.health --quick     # critical alerts only
mix optimal.verify --sample 20
mix optimal.stats
```

## HTTP API & visualizer

| Task                   | Purpose                                                                 |
|------------------------|-------------------------------------------------------------------------|
| `mix optimal.api`      | Start the HTTP API on port 4200                                         |
| `mix optimal.graph_ui` | Launch the graph visualizer against a running API                       |

The API exposes `/api/graph`, `/api/search`, `/api/l0`, `/api/health`, and
per-node subgraph endpoints. The visualizer is `priv/static/graph.html`.

## Spec tooling

| Task                     | Purpose                                                               |
|--------------------------|-----------------------------------------------------------------------|
| `mix optimal.spec.init`  | Scaffold the `.spec/` directory with templates and a starter spec     |
| `mix optimal.spec.check` | Validate spec files                                                   |
| `mix optimal.spec.drift` | Detect code changes without corresponding spec updates                |
| `mix optimal.spec.report`| Coverage and verification summary for all specs                       |

## Configuration

Runtime paths come from `config/config.exs` or environment variables:

- `OPTIMAL_ENGINE_ROOT` — root directory for nodes (default: `cwd()`)
- `OPTIMAL_ENGINE_DB` — SQLite database path (default: `<root>/.optimal/index.db`)
- `OPTIMAL_ENGINE_CACHE` — cache directory
- `OPTIMAL_ENGINE_TOPOLOGY` — routing rules YAML
- `OLLAMA_HOST` — embedding + generation backend (default: `http://localhost:11434`)
