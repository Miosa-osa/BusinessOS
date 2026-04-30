# OptimalEngine — Self-Hosted Docker Deployment

Run the entire OptimalEngine stack with a single command. The compose file starts four services:

| Service | Default URL | Description |
|---------|------------|-------------|
| engine  | `http://localhost:4200` | Elixir/OTP API + SQLite knowledge store |
| desktop | `http://localhost:1420` | Graph explorer UI (SvelteKit static) |
| site    | `http://localhost:1421` | Marketing site (SvelteKit static) |
| docs    | `http://localhost:1422` | Documentation site (SvelteKit static) |

---

## Prerequisites

- Docker Engine 24+ with Compose v2 (`docker compose` not `docker-compose`)
- BuildKit enabled (default in Docker 23+; set `DOCKER_BUILDKIT=1` if needed)
- ~2 GB free disk for images; ~100 MB per rebuild with layer caching

---

## Quickstart (development)

```bash
# 1. Clone the repo
git clone <repo-url> OptimalEngine
cd OptimalEngine/deploy

# 2. (Optional) Copy env defaults
cp env.example .env
# Edit .env if you need non-default ports or a custom Ollama URL

# 3. Build and start everything
docker compose up --build

# Wait for the engine health check to pass (~20 s), then open:
#   http://localhost:1420  — desktop UI
#   http://localhost:4200/api/health  — engine health JSON
```

To start in the background:

```bash
docker compose up --build -d
docker compose logs -f engine   # tail engine logs
```

To stop without losing data:

```bash
docker compose down        # stops containers, keeps volumes
docker compose down -v     # ALSO deletes the optimal_data volume (destructive)
```

---

## Production Deployment

### 1. Server setup

A single Linux VM with at least 2 vCPU / 2 GB RAM is sufficient for small teams.
Ports 80 and 443 must be reachable from the internet for Let's Encrypt ACME challenges.

### 2. DNS

Point the following A/AAAA records at the server's IP **before** starting Caddy:

```
optimal.example.com        → <server IP>
www.optimal.example.com    → <server IP>
docs.optimal.example.com   → <server IP>
```

### 3. Configure .env.prod

```bash
cd OptimalEngine/deploy
cp env.example .env.prod
```

Edit `.env.prod` and set **at minimum**:

```dotenv
OPTIMAL_AUTH_REQUIRED=true
DOMAIN=optimal.example.com
LETSENCRYPT_EMAIL=admin@example.com
```

### 4. Start the stack

```bash
docker compose \
  -f docker-compose.yml \
  -f docker-compose.prod.yml \
  up -d --build
```

Caddy will automatically obtain and renew TLS certificates. First startup may
take 30–60 s while the ACME challenge completes.

Check status:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
docker compose -f docker-compose.yml -f docker-compose.prod.yml logs caddy
```

---

## Volume Layout

All persistent state lives in the `optimal_data` Docker named volume, mounted at `/data` inside the engine container:

```
/data/
├── .optimal/
│   ├── index.db        — SQLite knowledge store (WAL mode)
│   ├── index.db-wal    — WAL journal (normal during operation)
│   ├── index.db-shm    — Shared memory file
│   └── cache/          — Embedding and parse caches
├── sample-workspace/   — Read-only sample signals (bind-mounted, not in volume)
└── <your-workspace>/   — Any workspace directories mounted additionally
```

### Backup

```bash
# Create a timestamped archive of the entire data volume
docker run --rm \
  -v optimal-engine_optimal_data:/data:ro \
  -v "$(pwd)":/backup \
  alpine \
  tar czf /backup/optimal_data_$(date +%Y%m%d_%H%M%S).tar.gz /data
```

### Restore

```bash
# Stop the stack first
docker compose down

# Restore from archive
docker run --rm \
  -v optimal-engine_optimal_data:/data \
  -v "$(pwd)":/backup \
  alpine \
  sh -c "cd / && tar xzf /backup/optimal_data_<timestamp>.tar.gz"

# Restart
docker compose up -d
```

---

## Updating

Pull the latest source, then rebuild:

```bash
git pull origin main

# Dev
docker compose up --build -d

# Production
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build -d
```

Compose will rebuild only the layers that changed. The `optimal_data` volume is
unaffected; the engine runs any pending SQLite migrations on startup.

---

## Environment Variables

See `env.example` for a fully annotated reference. Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `OPTIMAL_API_PORT` | `4200` | Engine HTTP port (dev host binding) |
| `OPTIMAL_AUTH_REQUIRED` | `false` | Enable API auth — **set `true` in production** |
| `OPTIMAL_DATA_DIR` | `/data` | Engine data root inside container |
| `OPTIMAL_OLLAMA_URL` | `http://host.docker.internal:11434` | Ollama endpoint for embeddings |
| `DESKTOP_PORT` | `1420` | Desktop UI host port (dev only) |
| `SITE_PORT` | `1421` | Marketing site host port (dev only) |
| `DOCS_PORT` | `1422` | Docs site host port (dev only) |
| `DOMAIN` | — | FQDN for production TLS (required in prod) |
| `LETSENCRYPT_EMAIL` | — | ACME contact email (required in prod) |

---

## Troubleshooting

### Port already in use

```
Error: address already in use :::4200
```

Another process is using that port. Change the host port in `.env`:

```dotenv
OPTIMAL_API_PORT=4201
DESKTOP_PORT=1430
```

Or stop the conflicting process: `lsof -i :4200 | grep LISTEN`

### Ollama not reachable

The engine logs will show connection refused on embedding calls. Check:

1. Ollama is running on the host: `ollama list`
2. On Linux, `host.docker.internal` requires Docker 20.10+. If it doesn't
   resolve, use the host's LAN IP instead:
   ```dotenv
   OPTIMAL_OLLAMA_URL=http://192.168.1.100:11434
   ```
3. Ollama is bound to `0.0.0.0` (not just `127.0.0.1`):
   `OLLAMA_HOST=0.0.0.0 ollama serve`

Embeddings are optional — search and graph features work without them;
only vector similarity ranking is degraded.

### SQLite WAL mode / locked database

If the engine crashes uncleanly, the WAL file may need a checkpoint:

```bash
# Run a WAL checkpoint from outside the container
docker run --rm \
  -v optimal-engine_optimal_data:/data \
  kevinlawson/sqlite3 \
  sqlite3 /data/.optimal/index.db "PRAGMA wal_checkpoint(FULL);"
```

In normal operation the engine handles checkpointing automatically.

### Engine fails to start — missing API config

The HTTP API is opt-in. In Docker mode the compose file sets:
```
OPTIMAL_API_ENABLED=true
OPTIMAL_API_INTERFACE=0.0.0.0
```
If you run the engine binary directly, set these env vars or pass them in `config/prod.exs`.

### docs service fails to build

`apps/docs/` is under active development. If it doesn't exist yet, start the
other three services only:

```bash
docker compose up --build engine desktop site
```

Add `docs` to the command once `apps/docs/` is present and has a working
`npm run build`.

### Checking image sizes

```bash
docker images | grep optimal
```

Target sizes (compressed):
- `engine` runtime stage: < 80 MB
- `desktop`, `site`, `docs` serve stage: < 40 MB each

---

## Assumptions

1. **`apps/docs/` will exist** with a standard SvelteKit + `@sveltejs/adapter-static` setup where `npm run build` outputs to `build/`. The `docs` service will fail to build until this directory is present.
2. **Ollama is not containerised** in this compose file. It is expected to run on the host machine and be reachable at `OPTIMAL_OLLAMA_URL`. Add an `ollama` service to the compose file if you need a fully self-contained stack.
3. **The engine release name is `optimal`** — matching the `releases:` key in `mix.exs`. The CMD in `Dockerfile.engine` is `/app/bin/optimal start`.
4. **`OPTIMAL_API_ENABLED` / `OPTIMAL_API_INTERFACE`** are consumed by the engine via its `Application` module reading `config :optimal_engine, :api`. If these env vars are not wired into `config/runtime.exs`, add them there. See `lib/optimal_engine/api/endpoint.ex` for the relevant config keys.
5. **The desktop app uses `@sveltejs/adapter-static`** with `fallback: "index.html"`. The Dockerfile expects `npm run build` to emit files into `build/`.
