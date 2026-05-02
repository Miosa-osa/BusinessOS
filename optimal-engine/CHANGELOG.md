# Changelog

All notable changes to the Optimal Engine are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/).

## [0.2.0] — 2026-04-30

### Added
- **Multi-workspace isolation** — organizations contain workspaces; each workspace gets its own filesystem, wiki, config, and scoped queries (Phases 1.5–1.7)
- **Memory primitive** — versioned memories with 5 typed relations (updates / extends / derives / contradicts / cites), soft-forget with reason, content-hash dedup, citation_uri on every memory (Phase 17)
- **Proactive surfacing** — background Surfacer GenServer with 14-category subscription model, SSE push channel, webhook callbacks (Phases 15, 18)
- **Memory ↔ Wiki bridge** — bidirectional integration: memories promote to wiki, wiki claims extract as memories (Phase 18)
- **API authentication** — bearer-token API keys with bcrypt-hashed secrets, scoped permissions (Phase 18)
- **Rate limiting** — ETS-backed token bucket per API key/IP, X-RateLimit-* headers (Phase 18)
- **Semantic grep** — `mix optimal.grep` + `GET /api/grep` with intent/scale/modality filters (Phase 16)
- **4-tier workspace profile** — `GET /api/profile` returns static + dynamic + curated + activity in one call (Phase 16)
- **5 typed recall endpoints** — `GET /api/recall/{actions,who,when,where,owns}` for cued memory recovery (Phase 15)
- **Streaming RAG** — `GET /api/rag/stream` SSE endpoint streaming pipeline stages (Phase 18)
- **Wiki contradiction detection** — 3 policies: flag_for_review / silent_resolve / reject (Phase 16)
- **Per-workspace YAML config** — `.optimal/config.yaml` with visualizations, profile, grep, contradictions, memory, rate_limit sections (Phase 16)
- **TypeScript SDK** — `@optimal-engine/client` with Vercel AI SDK + OpenAI Agents adapters (Phase 17)
- **Python SDK** — `optimal-engine` with sync/async clients, LangChain + OpenAI Agents adapters (Phase 17)
- **MCP server** — first-party stdio server with 9 tools for Claude Desktop / Cursor / Windsurf (Phase 17)
- **Chrome browser extension** — MV3 web clipper with popup, options, context menu (Phase 17)
- **Raycast extension** — 3 commands: search-memory, add-memory, ask-engine (Phase 18)
- **Claude Code Skill** — `skills/optimal-engine/` with SKILL.md + 6 references + bootstrap.sh (Phase 17)
- **Docs site** — SvelteKit static docs at `apps/docs/` with 21 pages (Phase 18)
- **Docker deployment** — `deploy/` with multi-stage Dockerfiles + compose + prod overlay + auto-TLS (Phase 18)
- **Desktop UI** — Foundation tokens, Org/Workspace switcher, 12 routes including /memory, /surface, /timeline, /heatmap
- **Marketing site** — landing with onion visualization, capabilities matrix, memory-failure section, on-disk convention
- **Schema migrations 026–030** — workspaces, surfacing, memories, content_hash, api_keys

### Changed
- Module rename: `OptimalEngine.Workspace` (nodes/skills) → `OptimalEngine.Topology`
- Module rename: `OptimalEngine.Topology` (YAML routing) → `OptimalEngine.Routing`
- `OptimalEngine.Workspace` now refers to the knowledge-base layer (multi-workspace)
- CI workflow optimized: paths-ignore for non-Elixir files, concurrency groups, single matrix entry

### Fixed
- FTS search workspace_id scoping (rows without workspace_id in SELECT caused MatchError)
- Grep SQL placeholder numbering bug (?1 then ?3 skipping ?2)

## [0.1.0] — 2026-04-18

### Added
- Initial release: 9-stage pipeline, 3-tier memory, 14 enterprise connectors, wiki layer, desktop UI, 1,075 tests passing
