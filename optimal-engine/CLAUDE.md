# Optimal Engine Claude Boot

This file is the Claude Code entrypoint for the public Optimal Engine repo.
Read `AGENTS.md` first when you need the full contract.
This file keeps the shortest possible rules in front of Claude users and coding agents.
The canonical upstream is `https://github.com/Miosa-osa/OptimalEngine.git` on `main`.
Never push Engine changes to `robertohluna/OptimalEngine`.

## What This Repo Is

Optimal Engine is the memory, retrieval, graph, RAG, claim, fact, and context package layer for OptimalOS, BusinessOS, and other workspace products.
It is the company second brain runtime, not a side notes folder.

## Start Every Session

Run this from the repo root:

```bash
git status --short --branch
bin/optimal doctor
```

If you need local engine context, boot the CLI loop:

```bash
bin/optimal boot
bin/optimal topology --workspace default:my-workspace
bin/optimal find "current state" --workspace default:my-workspace
bin/optimal rag "what context should I know?" --workspace default:my-workspace
```

Use the actual workspace ID for the task.
Never guess a workspace when the user has multiple workspaces with similar names.

## Agent Memory Loop

Use `bin/optimal` for normal agent memory work:

```bash
bin/optimal find "query" --workspace default:my-workspace
bin/optimal capture "raw signal or evidence" --workspace default:my-workspace
bin/optimal aware "important correction or durable signal" --workspace default:my-workspace
bin/optimal note "short note to remember"
bin/optimal decision "decision made"
bin/optimal task "action item"
bin/optimal close "what changed and how verified"
```

Use direct `mix optimal.*` commands only when developing or debugging the engine task itself.

## Data Boundary

Do not commit runtime data.
Do not commit private stores, connector keys, imported workspace data, user memory, local `.optimal/` files, databases, cache files, RocksDB data, or private machine paths.

Runtime data is local and ignored by git:

```text
.optimal/index.db
.optimal/index.db-wal
.optimal/index.db-shm
.optimal/cache/
.optimal/connector_key
.optimal/knowledge-rocksdb/
.optimal/workspaces/
```

The repo should contain code, docs, schemas, setup scripts, examples, and public sample fixtures only.

## Store Reality

Use these words precisely:

| Layer | Meaning |
| --- | --- |
| SQLite | Durable local canonical runtime store today. |
| Postgres | Production canonical runtime target. |
| RocksDB | Default local persistent knowledge graph backend when installed. |
| ETS | In-memory fallback graph backend. |
| Mnesia | Optional distributed graph backend. |
| Vectors, chunks, FTS, caches | Rebuildable retrieval projections. |
| RAG | An answer path over retrieved context, not a database. |

The practical flow is:

```text
raw inputs
  -> ingestion
  -> SQLite canonical store
  -> RocksDB graph runtime plus retrieval projections
  -> search, graph, and RAG
  -> API, CLI, wiki, and app surfaces
```

## BusinessOS Integration Rule

BusinessOS owns app state, windows, sessions, users, roles, workspace records, and operational records.
Optimal Engine owns knowledge, memory, source packages, claims, facts, retrieval, graph, RAG, and context packages.

Every BusinessOS call into Optimal Engine must include explicit tenant, organization, and workspace scope.
Downloaded users run their own local engine and must not read Roberto's private engine data.

## Verification

For docs-only changes:

```bash
git diff --check
bash -n scripts/run-engine.sh
```

For setup, backend, store, CLI, or runtime changes:

```bash
bash -n scripts/run-engine.sh
mix compile
mix optimal.reality_check
curl http://localhost:4200/api/stores/audit
```

If port `4200` is already in use, identify the process before touching it.
Do not kill a live user session without explicit permission.
