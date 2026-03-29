# Changelog

All notable changes to BusinessOS will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] — Process Mining Dashboard & CLI (feat/weaver-automation)

### Added
- **ProcessMapViewer.svelte** — BFS DAG layout, SVG Petri net rendering, HSL performance overlay, bottleneck pulse animation, pan/zoom
- **KPI Dashboard Widgets** — ConformanceScore, VariantDistribution, BottleneckHeatmap, CycleTimeTrend — wired to live pm4py-rust API
- **bos-commands crate** (`BusinessOS/bos/src/`) — proper Cargo workspace member with zero file moves
  - `bosctl` binary — 20 clap subcommands
  - 3 gateway commands (discover/conformance/statistics) via async-to-sync Tokio bridge
  - 16 local commands; 3 deferred (require model registry)
  - 11 unit tests pass

## [Unreleased] — Signal Theory 7-Layer Optimal System (d33a890, feat/sprint-3b-4-osa-modules)

### Added
- **L1 FastClassifier** (`internal/signal/classifier.go`): Zero-LLM keyword-based signal classification producing SignalEnvelope{Mode, Genre, DocType, Weight, Confidence} in <1ms
- **L2 CompetenceRegistry** (`internal/signal/genre_competence.go`): Maps (AgentType, Genre) → ContextHints + eligible DocTypes
- **L3 GenreEnricher** (`internal/agents/genre_enricher.go`): Model-agnostic signal annotation with context inventory, structure templates, and writing style guidance — works with any LLM (local 7B, Groq, Anthropic, OpenAI)
- **L3 Genre Templates** (`internal/agents/genre_templates.go`): Structure skeletons for 7 document types (proposal, sop, report, brief, framework, guide, plan) + per-genre output hints
- **L4 TieredContext Injection**: Context now injected into authoritative 13-layer system prompt with adaptive token budget based on signal weight
- **L5 DIKW Classifier** (`internal/services/dikw_classifier.go`): Data/Information/Knowledge/Wisdom hierarchy classification for workspace memories
- **L6 TimeToDecide Metric** (`internal/subconscious/time_to_decide_impl.go`): Sliding window latency metric (p50/p99) for message-to-first-token timing
- **L6 Failure Detectors** (`internal/signal/failure_detectors.go`): 5 Shannon/Ashby/Beer/Wiener failure detectors (BandwidthOverload, FeedbackFailure, RoutingFailure, GenreMismatch, BridgeFailure)
- **L6 Double-Loop Controller** (`internal/feedback/double_loop.go`): Argyris double-loop feedback — auto-relaxes setpoint tolerances after 3 consecutive failures
- **L7 Algedonic Channel** (`internal/governance/algedonic.go`): Beer's VSM emergency bypass S1→S5 with logging + postgres audit trail
- **Frontend Signal Display**: SSE `signal_classified` event, SignalBadge with genre/mode/docType, SignalDetailPanel, SignalHealthWidget dashboard widget
- **Signal Health API**: GET `/api/v1/signal/health` endpoint for system observability
- SQL migrations: `099_dikw_classification.sql`, `100_governance_events.sql`

### Changed
- `chat_streaming.go`: Signal classification runs on every message (no gating), annotation is purely additive metadata
- `base_agent.go`: Genre context and TieredContext injected into system prompt assembly
- `homeostatic_loop.go`: Algedonic callback fires on persistent correction failures (3+ consecutive)
- `metric_emitter.go`: TimeToDecide metric wired into emission pipeline

### Architecture
- **No-gate principle**: Signal classification is additive metadata only — no middleware filtering between user message and LLM API call
- **Model-agnostic**: Signal annotations include explicit structure, plain-English genre descriptions, and writing style guidance that any model can follow
- **40 files total**: 21 new files, 19 modified across backend and frontend

---

## [Unreleased] — MIOSA Platform Sprint 4

### Architecture Decision (2026-02-27 — D16/D17)
- **Build order confirmed:** Open source OSA first → closed source BOS → templates
- Open source OSA (github.com/Miosa-osa/OSA) = terminal-first agent foundation
- Closed source BOS = premium layer (template context, module awareness, SORX, autonomy)
- Templates (BusinessOS, CustomOS, etc.) embed agent with per-template context
- Fixing OSA agent is #1 priority (D11) — everything else downstream

### Added (Sprint 2.5 — 2026-02-28, direct commits)

**BOS ↔ OSA Full Integration (Roberto)**
- Stripped ~7,950 LOC of orchestration from BOS, rewired through OSA SDK
- `chatSSEParser.ts` — 470-line typed SSE parser (12 event types, async generator)
- `SignalBadge.svelte` — 5-mode color pill component
- `ToolCallCard.svelte` — 3-state collapsible tool invocation card
- 7 proxy routes: swarm CRUD, fleet dispatch, tools
- `OSA_ENABLED` defaults to true (G1 blocker FIXED)
- Boot health check + graceful fallback logging

**Cognitive Session Services (Roberto)**
- Context tracker, mode transitions, session health wiring

### Added (Sprint 4 — 2026-02-27/28, PRs #17-#22)

**Module System (PR #17 — Pedro)**
- Module lifecycle test suite — single install, multi install, uninstall, add/remove features, version upgrade, protection enforcement

**Module Management UI (PR #18 — Javaris)**
- `/settings/modules` page — installed modules list, install/uninstall flows, confirmation dialogs, error handling (`MANIFEST_INVALID`, `PROTECTION_VIOLATION`, `SQL_UNSAFE`)
- Skills page (`/skills`) — skill catalog with API wiring, inline decision cards (~21K lines)
- `syncDynamicModules()` — parallel module fetch on startup with `Promise.allSettled`
- BUILD mode "Module created — View in Settings" link in `ResponseStream`
- Wire install/uninstall to `desktop3dStore.addModule()`/`removeModule()` + dock icons
- OSA store `destroy()` lifecycle method

**OSA Glassmorphism Redesign (PR #19 — Javaris)**
- `GlassCard.svelte` — reusable glass-effect container component
- `PillButton.svelte` — glassmorphism pill button with hover states
- ChatGPT-style message bubbles with avatars and timestamps in `ResponseStream`
- Keyboard accessibility (Enter/Space) on collapsed `OsaPill`
- Lucide SVG icons replacing emoji in `ModeSelector`
- DOMPurify SSR guard (`sanitizeHtml` wrapper with browser check)

**Pedro Blockers Fix (PR #20 — Javaris)**
- Table name inconsistency fix (`custom_modules` vs `modules`)
- Module fixtures + lifecycle tests alignment

**OSA Pill Redesign + Store Quick Wins (PR #21 — Javaris, OPEN)**
- OsaPill glassmorphism redesign with pointer-events passthrough
- MODE_COLORS centralized to osa.ts store (single source of truth)
- `loadModes()` — browser-only, idempotent, 3-second abort timeout
- `reinstallModule` refactored to use typed `updateInstallation()` API
- knowledge-v2 page with retry/error states
- Fix: compact mode label abbreviation, a11y combobox on button, error state wiring

**Module System Backend — Sprint 3B Cherry-Pick (PR #22 — Pedro, OPEN)**
- Migration 097: `module_docs` table + `module_migration_history` + `module_route_registrations`
- `ModuleInstallService` — full install lifecycle (validate → pre-check → install → migrate → register routes)
- `ResolveDependencies` — batch dependency resolution (single query, no N+1)
- `ModuleDocsService` — index, search, delete module documentation with pgvector
- E2B sandbox client Windows path fix (`filepath.ToSlash`)
- Fix: unique constraint for module_docs upsert, transaction wrapping for Install, HTTP method validation, checksum verification

### Codebase Maintenance (2026-02-27/28 — direct commits)
- Removed 22K lines of dead code across backend and frontend
- Stripped V2 naming from agents, knowledge route, and chat handler
- Resolved 3 type errors in `templates.test.ts`

### Documentation (2026-02-27/03-03 — direct commits)
- `CODEBASE.md` — Complete architecture doc: request lifecycle, all 450+ endpoints, 280+ components, ~130 tables, dead code analysis, tiered from most critical to least
- `COMPONENTS.md` — Full 280+ Svelte component reference with props, descriptions, consumer routes
- `SIGNAL-REPORT.md` — Per-person signal dashboards (Roberto/Javaris/Pedro) with build pipeline context
- `README.md` — Enterprise-grade doc index with status markers and quick access table

---

### Added (Sprint 3C — 2026-02-26/27, commits f6b8dd6, ef380af, cd0107d)

**CARRIER Bridge (BOS ↔ Optimal)**
- CARRIER AMQP 0-9-1 bridge — full protocol implementation in Go (`internal/carrier/`) connecting BOS to Optimal (Elixir/OTP) over RabbitMQ
- Full MessageContext propagation — conversation history, tiered context (L1-L3), RAG results, user preferences, installed modules all flow through CARRIER
- Proactive command consumer — handles `request_decision`, `proactive_signal`, `execute_action` from Optimal
- Registration + heartbeat system — BOS auto-registers with Optimal, sends heartbeats on configurable interval
- Graceful degradation — CARRIER failures fall through to local agent execution transparently

**Optimal (Elixir/OTP) — New Repository: Miosa-osa/Optimal**
- STEERSMAN reasoning engine (1050 lines) — unified cybernetic planning: auto-complexity classification (1-10), MCTS budget selection, Boardroom multi-perspective deliberation, Critic triple-layer verification (formal 40% + semantic 35% + info-theoretic 25%), SHA-256 plan caching (5-min TTL)
- HOMEOSTAT viability monitor (820 lines) — VSM-based system health: 15s polling, adjustment loops, ultrastability, critical breach notification via ProactivePublisher
- Engine upgrades — MCTS (LLM-backed expand/simulate, UCB1), Boardroom (5-perspective parallel deliberation), Critic (triple-layer verification), Dispatcher (new routes: steersman, deliberate, verify, homeostat)

**OSA Mode Classification Upgrade**
- LLM fallback for ambiguous messages — keyword scores in [0.4, 0.7) zone routed to LLM with structured JSON schema, falls back to ASSIST on failure
- 45 E2E tests — realistic messages covering all 5 modes, priority order collisions, keyword gap documentation
- Full pipeline integration tests — classify → dispatch verification with 25-message coverage
- 4 nil-guard bug fixes — PACT, BMAD, and orchestrator handle nil agent registry gracefully instead of panicking

**OSA Module Protection**
- `validateBuildIntent()` in orchestrator — checks module protection manifests before PACT/BMAD dispatch
- Module detection from natural language messages (`detectModuleFromMessage`)

**SORX Enhancements**
- AgentBridge `buildMessageContext` — full context extraction for CARRIER routing
- Proactive scheduler — SORX skills can be scheduled via CARRIER signals from Optimal
- Tier-based CARRIER routing with priority mapping

### Repository Structure Update
- **BOS** (`robertohluna/BOS`, private) — Full system: desktop shell + OSA + SORX + all business modules. Active development.
- **Optimal** (`Miosa-osa/Optimal`, private) — Elixir/OTP reasoning engine: STEERSMAN, HOMEOSTAT, MCTS, Boardroom, Critic. Connected to BOS via CARRIER.
- **BusinessOS** (`robertohluna/BusinessOS`, private → public) — Open-source version. Full system MINUS proprietary SORX/Signal Theory/advanced learning.
- **CustomOS** (future) — Bare desktop shell, no modules, no business logic. Build anything on it.

### Added (Sprint 3 — 2026-02-26/27, commits ef380af, f6b8dd6, cd0107d)
- **CARRIER AMQP bridge** — Full Go client (`internal/carrier/`) implementing AMQP 0-9-1 protocol for BOS ↔ Optimal (Elixir) communication. Includes `carrier_client.go` (connection, reconnection, RPC), `messages.go` (MessageContext with tiered context L1-L3, conversation history, RAG, preferences), `proactive_consumer.go` (handles `request_decision`, `proactive_signal`, `execute_action` from Optimal), `registration.go` (heartbeat + registration with Optimal).
- **SORX proactive scheduler** — `internal/sorx/scheduler.go` adds cron-like skill scheduling with viability monitoring. Integrates with CARRIER for Optimal-driven scheduling.
- **SORX CARRIER routing** — `internal/sorx/carrier_routing.go` routes Tier 3-4 skill invocations through CARRIER to Optimal's SorxMain engine with transparent fallback to local execution.
- **Full MessageContext bridge** — `internal/sorx/agent_bridge.go` now populates complete MessageContext (workspace_id, mode, temperature, user_role, installed_modules, connected_integrations, conversation_history) for CARRIER requests.
- **LLM fallback for mode classification** — Mode classifier now calls LLM for messages scoring [0.4, 0.7) on keyword matching (the ambiguous zone). Returns structured JSON with mode, confidence, reasoning. Falls back to ASSIST on any failure. Existing keyword fast-path unchanged.
- **OSA module protection in BUILD flow** — `osa_orchestrator.go` validates BUILD intent against module protection manifests before dispatching to PACT/BMAD. Prevents unauthorized modifications to protected modules.
- **Proactive command handling** — BOS can now receive and execute commands from Optimal: decision requests, proactive signals, and action execution.
- **Optimal config** — `internal/config/optimal.go` adds OptimalConfig (Enabled, Mode, InstalledModules, Capabilities, HeartbeatInterval, TemplateType).
- **E2E mode classification tests** — 45 subtests covering all 5 modes, priority order collisions, keyword gaps, and ambiguous messages.
- **Integration pipeline tests** — Full classify → dispatch pipeline tests verifying events are emitted correctly for all 5 modes.

### Fixed (Sprint 3 — 2026-02-27)
- **4 nil-registry panics** — `pact.go`, `bmad.go`, `osa_orchestrator.go` (handleAssistMode, handleAnalyzeMode) now gracefully degrade instead of panicking when agent registry is nil (degraded/test deployments).

### New Repositories Created
- **Optimal** (`Miosa-osa/Optimal`) — Elixir/OTP engine for MIOSA's premium AI reasoning. Contains STEERSMAN (unified cybernetic reasoning: MCTS → Boardroom → Critic), HOMEOSTAT (VSM viability monitor), CARRIER consumer, Engine upgrades (MCTS, Boardroom, Critic, Dispatcher). This is the "brain" that BOS connects to via CARRIER.

### Added (Sprint 3 — 2026-02-25/26, PRs #8, #9, #13-#15)
- **OSA mode classification API** — `POST /api/osa/modes` classifies messages into BUILD/ASSIST/ANALYZE/EXECUTE/MAINTAIN with confidence scores. `GET /api/osa/modes` returns all modes with metadata for frontend dropdown.
- **DB-backed SORX skills** — Migration 096 creates `sorx_skills` table with pgvector embeddings. Skills loaded from DB instead of hardcoded Go maps. Admin can add skills via INSERT without redeploy.
- **Semantic skill matching** — Replaced 11 `strings.Contains` patterns with pgvector cosine similarity over skill embeddings. Configurable confidence threshold (default 0.72). Graceful fallback to hardcoded patterns.
- **Real ANALYZE mode** — Queries actual workspace data (app counts, queue stats, active modules) before calling analyst agent. LLM gets real data, not hallucinations.
- **Real MAINTAIN mode** — Returns real DB health metrics (module installations, queue failures, embedding coverage) instead of hardcoded stubs.
- **Signal Theory** — Proxy metrics, failure modes, triple-layer verification (PRs #13, #14)
- **E2B sandbox integration** — Sandbox edit lifecycle, 42 comprehensive tests (PR #15)
- **Mode feedback wiring** (D1 task)
- **Skill loader service** — Redis-cached skill loading with 5-minute TTL, automatic embedding backfill

### Fixed (Sprint 3)
- **SkillID field bug** — `semantic_skill_matcher.go` returned `skillName` instead of `skillID` for `SkillMatch.SkillID`. SORX expects UUID, not dotted name.

### Security (Sprint 2-3 overlap — PRs #10-#12)
- **OWASP Top 10 hardening** — Comprehensive security fixes across backend
- **Credential cleanup** — Hardcoded credentials removed from 20 files
- **HMAC auth middleware** — Internal service-to-service authentication
- **Production hardening + GDPR compliance**
- **Route auth hardening** — All endpoints verified for auth requirements

## [1.1.0-sprint2] — MIOSA Platform Sprint 2 (2026-02-24)

### Fixed (Sprint 2 — 2026-02-24, PR #5 + PR #6)
- **Streaming race condition (P0):** `base_agent_v2.go` Run() exited prematurely when error channel closed before chunks channel (Go defer LIFO). Chat was returning 0 output tokens. Fixed with ok-check and channel nil-out pattern.
- **Supabase Transaction Pooler compatibility (P0):** PgBouncer/Supavisor doesn't support extended query protocol. Added auto-detection (port 6543) + SimpleProtocol mode in `postgres.go`. Fixed 41 `[]byte("{}")` → `nil` across 17 files to prevent hex-encoding as bytea. Added `::jsonb` casts to 14 sqlc query files.
- **Wrong model sent to LLM (P0):** `chat_v2.go` used `DefaultModel` (llama3.2:3b) for all providers. Added `GetModelForProvider()` in `config.go` returning correct model per active provider (Ollama Cloud → glm-5:cloud, Anthropic → claude-sonnet, Groq → llama-3.3-70b).
- **FK violation in file generation (P0):** `osa_generated_files.workflow_id` NOT NULL but queue passes NULL. Migration 092 makes it nullable, fixes FK reference to correct table.
- **Prompt overwrite (P0):** `buildPrompt()` in `osa_queue_worker.go` never read `config["prompt"]`. Fixed with priority chain. Added 8 tests.
- **Gin route panic:** `:v1` conflicted with `:version` at same path level. Changed to `/versions/compare/:v1/:v2`.
- **Task toggle enum:** `::task_status` → `::taskstatus` (correct PostgreSQL enum name).
- **Workspace invitations schema:** Rewrote 9 queries from phantom `workspace_invitations` table to actual `workspace_invites` with correct columns.
- **Missing project columns:** Migration 093 adds `start_date`, `due_date`, `completed_at`, `visibility`, `owner_id` to projects table.
- **Missing sandbox infrastructure:** Migration 094 creates `sandbox_events` table.

### Added (Sprint 2)
- Ollama Cloud native API integration — rewrote from OpenAI-compatible `/v1/chat/completions` to native `/api/chat` with NDJSON streaming
- Email/password auth flow — `auth_email.go` with bcrypt + session cookies
- Provider-aware model selection — `config.GetModelForProvider()`

### Verified Working
- Full E2E chat pipeline: Auth → CSRF → DB → Agent classification → LLM (Ollama Cloud GLM-5:cloud) → SSE streaming
- `GET /projects` → 200, `POST /chat/v2/message` → streaming tokens

## [1.0.0-sprint1] — MIOSA Platform Sprint 1 (2026-02-23)

### Added
- OSA 5-Mode Router (BUILD/ASSIST/ANALYZE/EXECUTE/MAINTAIN) — `internal/orchestration/osa_modes.go` (349 lines, 56 tests)
- Module Protection Service (manifest validation, 4 pattern types) — `internal/services/module_protection.go` (306 lines, 48 tests)
- SORX Engine wired into OSA Orchestrator via `cmd/server/main.go`, `handlers.go`, `chat_v2.go`
- Tenant/Organization foundation — `supabase/migrations/090_tenant_org_foundation.sql` (155 lines)
- Agent Dispatch system for 14 parallel Claude Code agents across 3 operators
- Team briefing document with human oversight checkpoints
- Electron packaging guide (DMG, EXE, DEB, code signing, auto-updates)
- Production go-live checklist (10 phases)
- Enterprise documentation gap analysis
- GitHub PR template and issue templates (bug report, feature request)
- Operations shell docs (incident response, release process, SLO/SLI, runbook, secrets management)
- SECURITY.md and CONTRIBUTING.md at repo root

### Changed
- OSA Orchestrator refactored with 5 mode handlers — `internal/orchestration/osa_orchestrator.go`
- Docs directory reorganized: 146 old docs archived to `docs/_archive/`, 68 active docs remain
- Architecture docs updated with implementation reality badges
- SORX-CARRIER architecture doc updated with existence/non-existence reality check

## [0.9.0] - 2026-01-18 (Pre-MIOSA)

> Everything below is from the pre-pivot era. Preserved for history.

### Added
- Memory hierarchy system (workspace/project/agent)
- Security hardening (auth, sessions, CORS, rate limiting)
- Background jobs system with retry logic
- Chain-of-Thought orchestration
- SSE streaming for AI responses
- CI/CD pipelines (9 GitHub Actions workflows)
- 93 SQL migrations
- 11 integration providers
- Desktop shell (Electron + SvelteKit + Go backend)
- Docker terminal system

## [2.1.0] - 2026-01-11

### Added
- Thinking/COT System verification and documentation
- ThinkingPanel component in production
- 4 built-in reasoning templates in database
- Settings pages for AI configuration
- Real-time SSE streaming confirmation

### Fixed
- Background jobs system verification
- Migration 036 applied to Supabase
- 3 workers + scheduler auto-start
- 12 REST API endpoints for job management

## [2.0.0] - 2026-01-08

### Added
- Background Jobs System with complete documentation
- Retry logic with exponential backoff
- Cron scheduling with timezone support
- Job priority and dependency management

### Changed
- Merged pedro-dev branch with main-dev
- Updated database schema with new tables
- Enhanced API with job management endpoints

## Previous Releases

See git history for releases prior to 2.0.0.

---

**Note:** This changelog focuses on changes from Q1 2026. For older changes, please refer to the git commit history.
