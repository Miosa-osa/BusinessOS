# Getting Started

This walks you from a fresh clone to your first ingest-and-search round trip in
about two minutes.

## 1. Prerequisites

- **Erlang/OTP 26+** and **Elixir ~> 1.17**. Install via `asdf` or `homebrew`.
- **C toolchain** — required for the `:exqlite` SQLite NIF. macOS: Xcode
  Command Line Tools (`xcode-select --install`). Linux: `build-essential`.
- **Git**.

Optional:
- **Ollama** — for embeddings (`nomic-embed-text`) and local generation. Without
  Ollama the engine runs fine; hybrid vector search just degrades to BM25.
- **RocksDB** — alternative knowledge backend. Not needed for the default ETS
  path. Tests that require it are skipped unless the NIF is loadable.

## 2. Clone and compile

```bash
git clone git@github.com:robertohluna/OptimalEngine.git
cd OptimalEngine
mix deps.get
mix compile
```

The first compile pulls ~30 hex packages and builds the SQLite NIF. Expect
60–120 seconds on a fresh machine.

## 3. Run the test suite

```bash
mix test
```

Expect `689 tests, 0 failures (29 excluded)`. The 29 excluded tests require the
RocksDB backend.

## 4. First ingest

The engine stores signals and contexts under a root directory controlled by
`config/config.exs` (defaults to `.optimal/` under the current working
directory). Point it somewhere you control:

```bash
export OPTIMAL_ENGINE_ROOT=$HOME/tmp/optimal-demo
mkdir -p "$OPTIMAL_ENGINE_ROOT"
mix optimal.ingest "Customer called about pricing, wants $2K per seat" --genre note
```

What happens under the hood:
1. The signal is classified on `S=(Mode, Genre, Type, Format, Structure)`.
2. The router picks a destination node based on rules in `config.yaml` (falls
   back to `09-new-stuff` when routing is ambiguous).
3. A signal file is written under `nodes/<routed-node>/signals/YYYY-MM-DD-slug.md`.
4. The file is indexed into SQLite (`FTS5` full-text + BM25) and, if Ollama is
   up, into the vector store.
5. Cross-references are recorded (financial data and decisions get automatic
   secondary routes).

## 5. First search

```bash
mix optimal.search "pricing"
```

Returns L0 abstracts (~100 tokens each) by default — the tier designed for
"search, then decide what to load fully." To open one:

```bash
mix optimal.read "optimal://nodes/ai-masters/signals/2026-04-17-ed-pricing.md" --tier l1
```

Tiers:
- `l0` — ~100 tokens. Abstract. Cheap to load many.
- `l1` — ~2K tokens. Summary with key facts. The default working tier.
- `full` — complete content. Use only when `l1` is insufficient.

## 6. L0 — the always-loaded context

```bash
mix optimal.l0
```

Prints every node's abstract. This is what agents load at session start — the
whole knowledge base in roughly 2K tokens.

## 7. Assemble tiered context for a topic

```bash
mix optimal.assemble "pricing decisions"
```

Returns an L0 + L1 + L2 bundle with token counts so you can fit a context
window deterministically.

## 8. Keep going

- `mix optimal.graph` — knowledge graph statistics and analysis
- `mix optimal.reflect` — find missing edges between entities that co-occur
- `mix optimal.reweave "ed honour"` — find stale contexts about a topic
- `mix optimal.health` — 10 diagnostic checks on the knowledge base
- `mix optimal.remember "always check duplicates before insert"` — store a
  learned observation for future sessions

Full command reference: [mix-tasks.md](mix-tasks.md).

## 9. Architecture deep-dive

When you want to understand what's actually happening under the calls:

- [System overview](../architecture/system-overview.md) — the 30,000-foot view
- [FULL-SYSTEM-ARCHITECTURE](../architecture/FULL-SYSTEM-ARCHITECTURE.md) — the walkthrough
- [Signal Theory](../concepts/signal-theory.md) — the framework that governs classification
