# BusinessOS V2: Remaining Implementation Gaps

All V2 gaps tracked here are implemented.

This document tracks the tasks from the [taks_v2.md](taks_v2.md) master plan that were NOT yet fully operational.

## ✅ Status

- All checklist items below are completed.

## 🧪 Local Development (No Docker)

You can run the backend without Docker/Postgres using a **degraded mode** that boots the server and exposes health/status endpoints.

- Backend (degraded mode):
	- Set `DATABASE_REQUIRED=false` in `desktop/backend-go/.env`
	- Optionally disable Redis noise: set `REDIS_URL=`
	- Run: `go -C desktop/backend-go run ./cmd/server`
	- Check: `GET http://localhost:8001/api/status`

Note: In degraded mode, DB/auth-dependent APIs are not registered; this is intended for UI/dev plumbing without Docker.

## 🔴 CRITICAL GAPS

### 1. Output Styles & Block System (Part 3)
- [x] **Markdown-to-Block Translation**: Backend converts LLM Markdown into structured `Block` output when `structured_output=true` (via BlockMapper).
- [x] **Context-Specific Style Selection**: Backend auto-selects an output style from `focus_mode` / `agent_type`, with optional per-user overrides.
- [x] **User Style Overrides**: Backend persistence in `user_output_preferences` + Settings UI for default style.

### 2. Deep Context Integration (Part 4)
- [x] **Conversation Summarization**: Operational backfill + optional background job to generate/update `conversation_summaries` (with embeddings), enabling semantic indexing of past chat history.
- [x] **Voice Note Semantic Search**: Voice note transcripts are embedded on upload and included in semantic/text TreeSearch results.
- [x] **Node Hierarchy Inheritance**: TieredContext automatically includes selected node + ancestor node contexts when a child node is selected.

### 3. Self-Learning & Behavior Patterns (Part 6)
- [x] **Behavior Pattern Detection**: LearningService detects behavior patterns and persists them into `user_facts` (inactive by default) for explicit confirmation; optional background job keeps them fresh.
- [x] **Explicit Learning Validation**: Settings UI + backend endpoints to "Confirm" or "Reject" learned user facts (controls whether facts are injected).
- [x] **Automatic Context Injection**: ChatV2 always builds TieredContext and injects active "User Facts" into the system context automatically.

## 🟡 ENHANCEMENTS & REFINEMENT

### Context Management
- [x] **Context Tree API Refinement**: Dedicated stats endpoints for the tree visualization (`/api/context-tree/stats`).
- [x] **Token Budgeting (Priority-Based)**: Strict priority-based LRU eviction across context sections during prompt assembly (TieredContext formatting), enforcing per-agent `MaxContextTokens`.

### Developer Experience
- [x] **Application Profiler Sync**: Auto-syncing `ApplicationProfiles` with local file changes or Git events.
- [x] **Specialized Specialist Prompts**: Updating @coder, @analyst, and @researcher prompts to be as precise and recursive as the new Orchestrator V2.
