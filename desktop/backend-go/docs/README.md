# BusinessOS Backend Documentation

**Version:** 2.0
**Last Updated:** January 19, 2026
**Tech Stack:** Go 1.24.1 + PostgreSQL + Redis + pgvector

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [API Documentation](#api-documentation)
3. [Feature Documentation](#feature-documentation)
4. [Architecture](#architecture)
5. [Database](#database)
6. [Integrations](#integrations)
7. [Setup & Deployment](#setup--deployment)
8. [Team Review](#team-review)

---

## Quick Start

### Running the Server

```bash
# Install dependencies
go mod download

# Run development server (loads .env automatically)
go run ./cmd/server

# Build for production
go build -o bin/server ./cmd/server

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Essential Files

- **CLAUDE.md** - Main development guide for AI assistance
- **[ENVIRONMENT_SETUP.md](ENVIRONMENT_SETUP.md)** - Environment configuration
- **[QUICK_START_VALIDATION.md](QUICK_START_VALIDATION.md)** - Validation checklist

---

## API Documentation

Complete API reference and guides:

### Main References

| Document | Purpose |
|----------|---------|
| **[api/API_README.md](api/API_README.md)** | Complete API patterns documentation |
| **[api/API_CHEATSHEET.md](api/API_CHEATSHEET.md)** | Quick reference for API patterns |
| **[api/API_ENDPOINTS_REFERENCE.md](api/API_ENDPOINTS_REFERENCE.md)** | All endpoint details |
| **[api/API_VISUAL_GUIDE.md](api/API_VISUAL_GUIDE.md)** | Visual diagrams and flows |

### Specialized APIs

- **[api/OSA_BUILD_API_REFERENCE.md](api/OSA_BUILD_API_REFERENCE.md)** - OSA Build onboarding API
- **[api/MOBILE_API.md](api/MOBILE_API.md)** - Mobile app API reference
- **[api/api_rag_endpoints.md](api/api_rag_endpoints.md)** - RAG (Retrieval) endpoints
- **[api/osa-businessos-api.yaml](api/osa-businessos-api.yaml)** - OpenAPI spec

### SSE Streaming

- **[OSA_SSE_STREAMING.md](OSA_SSE_STREAMING.md)** - Server-Sent Events implementation
- **[OSA_SSE_QUICK_START.md](OSA_SSE_QUICK_START.md)** - Quick start guide

---

## Feature Documentation

### Agent System (V2)

**Location:** `docs/features/agents/`

The intelligent agent system that powers BusinessOS AI interactions.

| Document | Purpose |
|----------|---------|
| **[AGENT_SYSTEM.md](features/agents/AGENT_SYSTEM.md)** | Complete V2 architecture |
| **[CUSTOM_AGENTS_PRODUCTION_CHECKLIST.md](features/agents/CUSTOM_AGENTS_PRODUCTION_CHECKLIST.md)** | Launch checklist |
| **[CUSTOM_AGENTS_REVIEW_AND_IMPROVEMENTS.md](features/agents/CUSTOM_AGENTS_REVIEW_AND_IMPROVEMENTS.md)** | Review guide |
| **[CUSTOM_JOB_HANDLERS_GUIDE.md](features/agents/CUSTOM_JOB_HANDLERS_GUIDE.md)** | Background jobs |

**Key Components:**
- Agent V2 Interface
- Orchestrator (Intent Routing)
- Specialized Agents (Document, Project, Task, Client, Analyst)
- Tool Calling System
- Chain of Thought (COT)

---

### Voice System

**Location:** `docs/features/voice/`

Real-time voice conversations with OSA using LiveKit + gRPC.

| Document | Purpose |
|----------|---------|
| **[VOICE_SYSTEM.md](features/voice/VOICE_SYSTEM.md)** | Complete architecture |
| **[VOICE_SYSTEM_STATUS.md](features/voice/VOICE_SYSTEM_STATUS.md)** | Current status |
| **[VOICE_TESTING_GUIDE.md](features/voice/VOICE_TESTING_GUIDE.md)** | Testing guide |
| **[VAD_INTEGRATION.md](VAD_INTEGRATION.md)** | Voice Activity Detection |
| **[VOICE_ENVIRONMENT_CONFIG.md](VOICE_ENVIRONMENT_CONFIG.md)** | Configuration |

**Tech Stack:**
- LiveKit WebRTC (audio streaming)
- Python adapter (thin I/O bridge)
- Go backend (STT, TTS, Agent V2)
- gRPC bidirectional streaming
- Whisper (STT) + ElevenLabs (TTS)

---

### Workspace System

**Location:** `docs/features/workspace/`

Multi-user workspace collaboration and memory hierarchy.

| Document | Purpose |
|----------|---------|
| **[workspace_schema_analysis.md](features/workspace/workspace_schema_analysis.md)** | Database schema |
| **[workspace_implementation_status_complete.md](features/workspace/workspace_implementation_status_complete.md)** | Implementation status |
| **[workspace_frontend_integration_complete.md](features/workspace/workspace_frontend_integration_complete.md)** | Frontend integration |
| **[workspace_invite_audit_implementation.md](features/workspace/workspace_invite_audit_implementation.md)** | Invite system |
| **[workspace_memory_ui_guide.md](features/workspace/workspace_memory_ui_guide.md)** | Memory UI |

**Features:**
- Workspace-level permissions
- Invite system
- Memory hierarchy (workspace → project → agent)
- Role-based access control

---

### Background Jobs

**Location:** `docs/features/`

Async background processing system.

- **[BACKGROUND_JOBS_INTEGRATION_GUIDE.md](features/BACKGROUND_JOBS_INTEGRATION_GUIDE.md)** - Integration guide
- **[BACKGROUND_JOBS_QUICKSTART.md](features/BACKGROUND_JOBS_QUICKSTART.md)** - Quick start
- **[BACKGROUND_JOBS_README.md](features/BACKGROUND_JOBS_README.md)** - Overview

**Available Jobs:**
- Conversation summarization
- Behavior pattern detection
- App profiler sync

---

### Thinking System

**Location:** `docs/features/`

Chain-of-thought reasoning system for complex queries.

- **[THINKING_SYSTEM_INTEGRATION.md](features/THINKING_SYSTEM_INTEGRATION.md)** - Implementation guide

---

## Architecture

**Location:** `docs/architecture/`

System design and architectural decisions.

| Document | Purpose |
|----------|---------|
| **[BUSINESSOS_ARCHITECTURE.md](architecture/BUSINESSOS_ARCHITECTURE.md)** | Overall system architecture |
| **[BUSINESSOS_AGENT_ARCHITECTURE.md](architecture/BUSINESSOS_AGENT_ARCHITECTURE.md)** | Agent system design |
| **[CONTAINER_EXEC.md](architecture/CONTAINER_EXEC.md)** | Container execution |
| **[CONTAINER_MANAGER.md](architecture/CONTAINER_MANAGER.md)** | Container management |
| **[INTEGRATION_MODULE_MAPPING.md](architecture/INTEGRATION_MODULE_MAPPING.md)** | Integration architecture |

### Key Patterns

**Handler → Service → Repository (3-Layer)**

```go
// Handler validates input, calls service
func (h *Handler) CreateProject(c *gin.Context) {
    var req CreateProjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    project, err := h.service.CreateProject(c, req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, project)
}

// Service contains business logic
func (s *Service) CreateProject(ctx context.Context, req CreateProjectRequest) (*Project, error) {
    // Business logic
    return s.repo.Create(ctx, project)
}

// Repository handles database (SQLC)
func (r *Repository) Create(ctx context.Context, project *Project) error {
    return r.queries.CreateProject(ctx, database.CreateProjectParams{...})
}
```

---

## Database

**Location:** `docs/database/`

PostgreSQL schema, migrations, and SQLC patterns.

| Document | Purpose |
|----------|---------|
| **[database/DATABASE_SETUP.md](database/DATABASE_SETUP.md)** | Setup guide |
| **[database/DATABASE_LOCATION_INFO.md](database/DATABASE_LOCATION_INFO.md)** | Connection info |
| **[database/database_troubleshooting.md](database/database_troubleshooting.md)** | Troubleshooting |
| **[database/SUPABASE_MIGRATION.md](database/SUPABASE_MIGRATION.md)** | Migration guide |

### SQLC Workflow

1. Add migration: `internal/database/migrations/XXX_name.sql`
2. Update schema: `internal/database/schema.sql`
3. Add queries: `internal/database/queries/name.sql`
4. Generate code: `sqlc generate`

**Important:** SQLC-generated code is in `internal/database/sqlc/` - DO NOT EDIT directly.

---

## Integrations

**Location:** `docs/integrations/`

Third-party service integrations (Google, Slack, Notion, HubSpot, etc.)

| Document | Purpose |
|----------|---------|
| **[integrations/INTEGRATION_INFRASTRUCTURE.md](integrations/INTEGRATION_INFRASTRUCTURE.md)** | Overall architecture |
| **[integrations/INTEGRATION_SETUP_CHECKLIST.md](integrations/INTEGRATION_SETUP_CHECKLIST.md)** | Setup guide |
| **[integrations/INTEGRATIONS_MASTER_LIST.md](integrations/INTEGRATIONS_MASTER_LIST.md)** | All integrations |
| **[integrations/google-oauth/](integrations/google-oauth/)** | Google OAuth flow |
| **[integrations/livekit/](integrations/livekit/)** | LiveKit voice setup |

### OAuth Flow

All OAuth integrations follow this pattern:

1. User initiates OAuth from frontend
2. Backend redirects to provider
3. Provider redirects back with code
4. Backend exchanges code for tokens
5. Tokens encrypted and stored in DB
6. Access tokens cached in Redis

---

## Setup & Deployment

### Environment Configuration

**Required environment variables:**

```bash
# Database
DATABASE_URL=postgresql://user:pass@host:5432/businessos

# Redis (optional, for caching)
REDIS_URL=redis://localhost:6379

# AI Provider
AI_PROVIDER=anthropic  # or ollama_local, ollama_cloud, groq
ANTHROPIC_API_KEY=sk-ant-...
GROQ_API_KEY=gsk_...

# Security
SECRET_KEY=your-secret-key-change-in-production
TOKEN_ENCRYPTION_KEY=your-32-byte-key

# Voice (optional)
ELEVENLABS_API_KEY=your-elevenlabs-key
LIVEKIT_API_KEY=your-livekit-key
LIVEKIT_API_SECRET=your-livekit-secret

# Integrations (as needed)
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
```

See **[ENVIRONMENT_SETUP.md](ENVIRONMENT_SETUP.md)** for complete guide.

### Docker Deployment

```bash
# Build image
docker build -t businessos-backend .

# Run container
docker run -p 8080:8080 --env-file .env businessos-backend
```

### GCP Cloud Run

See deployment guide in main repo docs.

---

## Team Review

**Location:** `docs/team-review/`

Recent changes and improvements for team review.

### Recent Backend Changes (January 2026)

See **[team-review/RECENT_BACKEND_CHANGES.md](team-review/RECENT_BACKEND_CHANGES.md)** for:

1. **Voice System Improvements**
   - Agent V2 integration
   - Audio playback via SSE
   - LiveKit room management
   - VAD (Voice Activity Detection)

2. **OAuth Backend Implementation**
   - Google OAuth flow
   - Email analysis service
   - Token encryption
   - Refresh token handling

3. **Agent System Updates**
   - Agent V2 architecture
   - Intent routing improvements
   - Tool calling enhancements
   - COT orchestration

4. **Database Enhancements**
   - Workspace schema updates
   - Memory hierarchy improvements
   - pgvector optimizations

---

## Additional Resources

### Related Documentation

- **Main repo docs/** - Frontend, deployment, planning
- **CLAUDE.md** - AI assistant guidance
- **TECHNICAL_REFERENCE.md** - Technical deep-dive

### Code References

```
internal/
├── agents/          # Agent system
├── handlers/        # API handlers
├── services/        # Business logic
├── database/        # SQLC + migrations
├── integrations/    # OAuth providers
└── streaming/       # SSE streaming
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/agents/

# Run with coverage
go test -cover ./...

# Run integration tests
go test ./internal/handlers/ -v
```

---

## Contributing

When adding new features:

1. **Documentation first** - Update relevant docs
2. **SQLC workflow** - Add migrations, schema, queries
3. **3-layer pattern** - Handler → Service → Repository
4. **Tests** - Add unit tests for services
5. **API docs** - Update API reference

---

## Getting Help

1. Check **[API_CHEATSHEET.md](api/API_CHEATSHEET.md)** for quick patterns
2. Review **[TROUBLESHOOTING.md](database/database_troubleshooting.md)** for common issues
3. Search feature docs for specific components
4. Check **CLAUDE.md** for AI assistant guidance

---

**Last Updated:** January 19, 2026
**Maintained by:** Backend Team
**Status:** Production Ready
