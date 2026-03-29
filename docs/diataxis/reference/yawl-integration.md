# YAWL v6 Integration — Reference

BusinessOS proxies conformance checking, spec management, and workflow simulation to the YAWL v6 engine running at `YAWLV6_URL` (default: `http://localhost:8080`).

---

## Endpoints

All routes live under `/api/yawl`. Authentication: JWT bearer token or static bearer token (`YAWLV6_API_TOKEN`).

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/yawl/health` | Check YAWL engine status |
| POST | `/api/yawl/conformance` | Run conformance check via OSA → YAWL |
| POST | `/api/yawl/spec` | Build YAWL specification XML from pattern params |
| GET | `/api/yawl/spec/load` | Load a WCP pattern spec file from disk |
| GET | `/api/yawl/specs` | List available spec files |
| GET | `/api/yawl/real-data` | List available real event log datasets |
| GET | `/api/yawl/real-data/:name` | Get a specific real event log dataset |
| POST | `/api/yawl/simulate` | Simulate concurrent workflow execution |

### Authentication

Two modes, controlled by `YAWLV6_API_TOKEN` env var:

- **Token set** (`YAWLV6_API_TOKEN=...`): static bearer auth for service-to-service calls
- **Token empty** (default): JWT auth via standard middleware (browser / session callers)

---

## Configuration

| Env Var | Default | Purpose |
|---------|---------|---------|
| `YAWLV6_URL` | `http://localhost:8080` | YAWL v6 engine base URL |
| `YAWLV6_SPECS_PATH` | `~/yawlv6/exampleSpecs` | Path to WCP pattern `.yawl` files |
| `YAWLV6_API_TOKEN` | _(empty)_ | Static bearer token for service-to-service auth |

---

## API Examples

### Health Check

```bash
curl -s http://localhost:8001/api/yawl/health \
  -H "Authorization: Bearer $JWT"
```

Success:
```json
{"status": "ok", "engine": "yawl-v6"}
```

Engine unreachable:
```json
{"status": "unreachable", "error": "connection refused"}
```

### Conformance Check

```bash
curl -s -X POST http://localhost:8001/api/yawl/conformance \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{"spec_xml": "<net>...</net>", "event_log": [...]}' \
```

Response:
```json
{"fitness": 0.95, "violations": []}
```

### Simulate Workflows

```bash
curl -s -X POST http://localhost:8001/api/yawl/simulate \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{"spec_set": "basic_wcp", "user_count": 3}'
```

### Load Spec File

```bash
curl -s "http://localhost:8001/api/yawl/spec/load?pattern=WCP1" \
  -H "Authorization: Bearer $JWT"
```

---

## Key Implementation

| Component | Path |
|-----------|------|
| Route registration | `desktop/backend-go/internal/handlers/routes_yawl.go` |
| Handler logic | `desktop/backend-go/internal/handlers/yawl_handler.go` |
| OSA YAWL client | `OSA/lib/optimal_system_agent/yawl/client.ex` |
| OSA spec builder | `OSA/lib/optimal_system_agent/yawl/spec_builder.ex` |
| OSA simulator | `OSA/lib/optimal_system_agent/yawl/simulator.ex` |

---

## Health.jsp Note

YAWL's `health.jsp` returns JSON with a trailing comma (non-standard). The OSA client uses `decode_body: false` on the `Req.get` call to avoid Jason parse errors — only the HTTP status code (200) is checked. This prevents false 503 responses when the engine is healthy.

```elixir
# OSA/lib/optimal_system_agent/yawl/client.ex
case Req.get(url, receive_timeout: @timeout_ms, decode_body: false) do
  {:ok, %{status: status}} when status in 200..299 -> :ok
  {:ok, _} -> {:error, :unreachable}
  {:error, _} -> {:error, :unreachable}
end
```

---

## Docker

YAWL runs in Docker. Start with:

```bash
cd /path/to/yawlv6/docker
docker compose -f docker-compose.simple.yml up -d yawl-engine
```

The container uses `restart: unless-stopped` and exposes port 8080. Health check at `http://localhost:8080/health.jsp`.

---

*BusinessOS YAWL Integration Reference — Part of the ChatmanGPT Knowledge System*
