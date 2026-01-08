# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build the server
go build -o bin/server ./cmd/server

# Run the server (loads .env automatically)
go run ./cmd/server

# Run tests
go test ./...

# Run a single test file
go test ./internal/agents/agent_v2_test.go

# Run tests with verbose output
go test -v ./internal/handlers/...

# Generate SQLC code (after modifying queries or schema)
sqlc generate

# Run database migrations
go run ./cmd/migrate

# Re-embed documents (regenerate vector embeddings)
go run ./cmd/reembed

# Summarize conversations
go run ./cmd/summarize

# Build for Docker/Cloud Run
docker build -t businessos-server .
```

## Architecture Overview

This is a Go backend for BusinessOS, an AI-powered business operating system. It uses:
- **Gin** for HTTP routing
- **pgx/v5** for PostgreSQL (with pgvector for embeddings)
- **SQLC** for type-safe SQL query generation
- **Redis** for session caching and pub/sub (optional, for horizontal scaling)
- **Docker** for containerized terminal sessions (optional)

### Package Structure

```
cmd/
├── server/      # Main API server entrypoint
├── migrate/     # Database migration runner
├── reembed/     # Vector embedding regeneration
└── summarize/   # Conversation summarization

internal/
├── agents/      # AI agent system (orchestrator, document, project, task, client, analyst)
├── config/      # Configuration loading (Viper + .env)
├── container/   # Docker container management for terminal isolation
├── database/    # PostgreSQL connection, schema, and SQLC-generated code
│   ├── migrations/  # SQL migration files (numbered)
│   ├── queries/     # SQLC query files
│   ├── sqlc/        # Generated Go code (DO NOT EDIT)
│   └── schema.sql   # Full schema for SQLC
├── handlers/    # HTTP handlers (one file per domain)
├── integrations/  # Third-party integrations (Google, Slack, Notion, etc.)
├── logging/     # Structured logging with sanitization
├── middleware/  # Auth, CORS, rate limiting
├── prompts/     # AI prompt templates (agents and core)
├── redis/       # Redis client, pub/sub, session management
├── security/    # Encryption, validation utilities
├── services/    # Business logic layer
├── sorx/        # Sorx skill execution engine
├── streaming/   # SSE streaming for real-time AI responses
├── terminal/    # WebSocket terminal with PTY support
└── tools/       # AI tool definitions
```

### Key Architectural Patterns

**Layered Architecture**: Handler → Service → Database (SQLC)
- Handlers validate input and call services
- Services contain business logic
- SQLC provides type-safe database queries

**Agent System (V2)**:
- Agents implement `AgentV2` interface with `Run()` and `RunWithTools()`
- Return `<-chan streaming.StreamEvent` for real-time SSE
- Orchestrator routes to specialized agents (document, project, task, client, analyst)

**Integration Providers**:
- Each integration (Google, Slack, Notion, etc.) is a separate package under `integrations/`
- OAuth flows handled via `oauth.go` with provider-specific callbacks
- Unified registry pattern via `registry.go`

### Database / SQLC Workflow

1. Add migration to `internal/database/migrations/XXX_name.sql`
2. Update `internal/database/schema.sql` with table definitions (SQLC reads this)
3. Add queries to `internal/database/queries/name.sql`
4. Run `sqlc generate` to regenerate `internal/database/sqlc/`

**SQLC Query Naming**: Query names must be unique across all `.sql` files. Prefix with domain if conflicts arise (e.g., `ListCRMDeals` vs `ListDeals`).

### Authentication

Two modes supported:
- **Redis-cached auth**: For horizontal scaling (production)
- **Direct DB auth**: Single instance fallback

Auth middleware extracts user from session token, validates against `sessions` table, caches in Redis if available.

### Environment Configuration

Configuration via `.env` file or environment variables. Key variables:
- `DATABASE_URL` - PostgreSQL connection string (use pooled connection for Supabase)
- `REDIS_URL` - Redis for sessions/pub-sub (optional)
- `AI_PROVIDER` - `ollama_local`, `ollama_cloud`, `anthropic`, or `groq`
- `ANTHROPIC_API_KEY`, `GROQ_API_KEY` - API keys for AI providers
- `SECRET_KEY` - JWT signing key (MUST change in production)
- `TOKEN_ENCRYPTION_KEY` - AES encryption for OAuth tokens in DB

### Testing

Tests use `testify` for assertions. Mock services where needed:

```go
func TestSomething(t *testing.T) {
    // Setup
    pool := setupTestDB(t)

    // Test
    result, err := someFunction(pool, input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### SSE Streaming

AI responses stream via Server-Sent Events:
- Events: `content`, `thinking`, `artifact`, `tool_call`, `done`, `error`
- Use `streaming.SSEWriter` for consistent event formatting
- Clients consume via EventSource API

### Background Jobs

Optional background jobs (disabled by default):
- `ConversationSummaryJob` - Backfill conversation summaries
- `BehaviorPatternsJob` - Detect user behavior patterns
- `AppProfilerSyncJob` - Auto-sync application profiles

Enable via config: `CONVERSATION_SUMMARY_JOB_ENABLED=true`
