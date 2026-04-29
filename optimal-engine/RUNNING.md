# Running the OptimalEngine for BusinessOS

The BusinessOS Pages → Knowledge Graph view is powered by the Elixir
OptimalEngine in this directory. To make it light up:

## 1. Start the engine

```bash
make optimal-engine
```

This:
1. Installs Elixir deps if needed (`mix deps.get`)
2. Starts the engine on **http://localhost:4200**
3. Configures it to read from `sample-workspace/.system/index.db` —
   the same SQLite that BusinessOS writes to when users save Pages.

First boot takes ~1-2 minutes (cowboy + plug + erlex compile from
source). Subsequent boots are near-instant.

## 2. Confirm `.env` is wired

`desktop/backend-go/.env` should have:

```
OPTIMAL_ENGINE_URL=http://localhost:4200
OPTIMAL_OS_ROOT=../../optimal-engine/sample-workspace
OPTIMAL_NODES_ROOT=../../optimal-engine/sample-workspace/nodes
OPTIMAL_ENGINE_PATH=../../optimal-engine
OPTIMAL_DB_PATH=../../optimal-engine/sample-workspace/.system/index.db
```

When `OPTIMAL_ENGINE_URL` is set, every `/api/optimal/*` request the
BusinessOS Go server receives is reverse-proxied to the engine. When
unset, the Go fallback handles requests in-process (less feature-rich).

## 3. Verify

```bash
# Direct hit on the engine
curl http://localhost:4200/api/optimal/nodes

# Through the BusinessOS proxy (after `make dev-local`)
curl http://localhost:8801/api/optimal/nodes
```

Both should return the same JSON. The Pages → Knowledge Graph view in
the Electron desktop app will render the engine's graph.

## 4. Stop it

```bash
make optimal-engine-stop
```

## What happens when a user saves a Page

```
1. POST /api/contexts (BusinessOS frontend → Go server)
2. Save to Postgres contexts table  (existing behaviour)
3. PagesEngineSync writes
     OPTIMAL_NODES_ROOT/10-businessos/<page-id>.md
4. PagesEngineSync calls optimal.Reindex (Go) — updates the SQLite
   the Elixir engine reads from.
5. Pages → Knowledge Graph view re-fetches /api/optimal/graph
   (proxied to Elixir) — the new Page appears as a signal under the
   `10-businessos` node alongside the seed corpus.
```

## Troubleshooting

**"Pages graph is empty"**
Make sure the engine is running (`lsof -i :4200`), and that the
`OPTIMAL_ENGINE_URL` is set in `.env`. Check
`desktop/backend-go/.env` and the BusinessOS startup log for:

> `OptimalOS routes proxied to live Elixir engine engine_url=http://localhost:4200`

**"Engine sees old data after I edit MD files"**
Run `make optimal-reindex` to force a re-walk.

**"Engine and BusinessOS disagree on what data exists"**
They're hitting different SQLite files. Confirm both point at
`optimal-engine/sample-workspace/.system/index.db`:

```bash
lsof -p $(lsof -ti :4200) | grep '\.db$'
```

The path printed must match `OPTIMAL_DB_PATH` in `.env`.

## Why this layout

The Elixir engine is the **canonical runtime**. The Go port
(`Miosa-osa/OptimalEngine-go`, consumed via `go.mod replace`) is a
zero-dependency fallback for environments without an Erlang runtime.
The HTTP API surface is identical between them, so the BusinessOS
frontend doesn't care which one is on the other side of the proxy.
