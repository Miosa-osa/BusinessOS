# Developer Quick Start Guide

Get BusinessOS running locally in 15 minutes.

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.21+ | `brew install go` |
| Node.js | 20+ | `brew install node` |
| PostgreSQL | 14+ | `brew install postgresql@14` |
| Docker | Latest | Docker Desktop (optional) |
| Redis | 7+ | Optional |

## Quick Start (3 Steps)

### 1. Clone and Setup
```bash
git clone https://github.com/your-org/BusinessOS-1.git
cd BusinessOS-1
```

### 2. Configure Environment
```bash
# Backend
cp desktop/backend-go/.env.example desktop/backend-go/.env
# Edit: Set DATABASE_URL

# Frontend (optional)
cp frontend/.env.production.example frontend/.env
```

### 3. Start Everything
```bash
./dev.sh start
```

This starts:
- PostgreSQL (via Homebrew)
- Redis (optional)
- Go Backend at http://localhost:8001
- SvelteKit Frontend at http://localhost:5173

## Local Development (No Docker)

If you don't want Docker, you can run the frontend plus a **degraded backend** (no DB). This boots the server and exposes health/status endpoints so you can iterate UI and basic plumbing.

### Backend (degraded mode)

In `desktop/backend-go/.env`:

- Set `DATABASE_REQUIRED=false`
- Optionally disable Redis noise: set `REDIS_URL=`

Run:

```bash
go -C desktop/backend-go run ./cmd/server
```

Check:

```bash
curl -s http://localhost:8001/api/status
```

Note: In degraded mode, DB/auth-dependent APIs are not registered.

## Development Commands

| Command | Description |
|---------|-------------|
| `./dev.sh start` | Start all services |
| `./dev.sh stop` | Stop all services |
| `./dev.sh status` | Check service health |
| `./dev.sh logs` | Tail all logs |
| `./dev.sh restart` | Restart everything |

## Project Structure

```
BusinessOS-1/
├── desktop/backend-go/        # Go backend (Gin framework)
│   ├── cmd/server/main.go     # Entry point
│   ├── internal/
│   │   ├── handlers/          # API endpoints
│   │   ├── terminal/          # WebSocket PTY terminal
│   │   ├── container/         # Docker isolation
│   │   ├── middleware/        # Auth, CORS, rate limiting
│   │   └── database/          # PostgreSQL + SQLC
│   └── .env                   # Backend config
│
├── frontend/                  # SvelteKit 2.0 + Svelte 5
│   ├── src/routes/            # Pages and API routes
│   ├── src/lib/components/    # UI components
│   └── src/lib/api/           # Backend API client
│
├── docs/                      # Documentation
├── dev.sh                     # Development startup script
└── startup.sh                 # Legacy startup (use dev.sh)
```

## Key Environment Variables

### Backend (`desktop/backend-go/.env`)
```bash
# Database (required)
DATABASE_URL=postgres://user:pass@localhost:5432/business_os?sslmode=disable

# Server
SERVER_PORT=8001
ENVIRONMENT=development

# Redis (optional for dev, required for production)
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=dev-password

# OAuth (for Google login)
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8001/api/auth/google/callback

# AI (optional)
AI_PROVIDER=ollama_local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=llama3.2:3b
```

## Common Tasks

### Add New API Endpoint

1. Create SQL query in `internal/database/queries/`
2. Run SQLC: `cd desktop/backend-go && sqlc generate`
3. Create handler in `internal/handlers/`
4. Register route in the handler file

### Database Migrations

**Windows (PowerShell):**
```powershell
# Apply all migrations to local PostgreSQL
cd desktop/backend-go/scripts
.\apply-migrations.ps1
```

**Manual (any platform):**
```bash
cd desktop/backend-go
psql -U postgres -d postgres -f ../../supabase-migrations-combined.sql
```

This creates **26 tables** including:
- `memories` - Semantic memory storage with 768D embeddings
- `uploaded_documents` - Document management
- `document_chunks` - Chunked documents with embeddings
- `conversation_summaries` - AI conversation history
- `learning_events` - User feedback and personalization
- `application_profiles` - App-specific context
- And 20 more tables for the complete system

### Testing & Verification

**Setup Test User (for API testing):**
```powershell
# Windows
.\setup-test-user.ps1

# Or manually via SQL
psql -U postgres -d postgres -f test-user-setup.sql
```

This creates:
- Test User: `testuser@businessos.dev`
- Session Token: `test-token-businessos-123`
- Cookie: `better-auth.session_token=test-token-businessos-123`

**Test API Endpoints:**
```bash
# Health check
curl http://localhost:8001/health/detailed

# Send chat message (requires auth)
curl -X POST http://localhost:8001/api/chat/message \
  -H "Content-Type: application/json" \
  -H "Cookie: better-auth.session_token=test-token-businessos-123" \
  -d '{"message": "Hello", "conversation_id": null}'

# Upload document
curl -X POST http://localhost:8001/api/documents \
  -H "Cookie: better-auth.session_token=test-token-businessos-123" \
  -F "file=@test.txt" \
  -F "title=Test Doc" \
  -F "document_type=text"
```

**Run Unit Tests:**
```bash
# Backend
cd desktop/backend-go && go test ./...

# Frontend
cd frontend && npm test
```

**Verified Features (Jan 2, 2026):**
- ✅ Conversation system with SSE streaming
- ✅ Memory creation with 768D embeddings
- ✅ Document upload with automatic chunking
- ✅ Semantic search ready (pgvector + HNSW indexes)
- ✅ 337 API endpoints registered

### Build for Production

```bash
# Backend binary
cd desktop/backend-go && go build -o server cmd/server/main.go

# Frontend static
cd frontend && npm run build
```

## Architecture Overview

```
                    Frontend (SvelteKit)
                    http://localhost:5173
                            │
                            ▼
                    Backend (Go/Gin)
                    http://localhost:8001
                            │
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
        PostgreSQL      Redis           Docker
        (data store)    (sessions,      (terminal &
                        pub/sub)        file isolation)
```

## Key Features Implemented

| Feature | Status | Documentation |
|---------|--------|---------------|
| Google OAuth | Done | `internal/handlers/auth_google.go` |
| Terminal (PTY) | Done | `docs/TERMINAL_SYSTEM.md` |
| File Browser | Done | Container isolation via Docker |
| Rate Limiting | Done | 100 msg/sec, 16KB max message |
| Session Security | Done | 30 min timeout, crypto/rand IDs |
| Redis Pub/Sub | Done | Horizontal scaling ready |

## Troubleshooting

### Database Connection Failed
```bash
# Ensure PostgreSQL running
brew services start postgresql@14

# Create database
psql -U postgres -c "CREATE DATABASE business_os;"
```

### Redis Unavailable
```bash
# Check Docker
docker ps | grep businessos-redis

# Restart Redis
docker-compose -f desktop/backend-go/docker-compose.yml up -d redis
```

If you are running with `REDIS_URL=` (no Redis), you can ignore this.

### Port Already in Use
```bash
# Check what's using port 8001
lsof -i :8001

# Kill process
kill -9 <PID>
```

### CORS Errors
Ensure `ALLOWED_ORIGINS` in backend `.env` includes frontend URL:
```bash
ALLOWED_ORIGINS=http://localhost:5173
```

## Next Steps

- [ARCHITECTURE.md](./ARCHITECTURE.md) - Full system architecture
- [FRONTEND.md](./FRONTEND.md) - SvelteKit patterns and components
- [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) - Production deployment
- [API_REFERENCE.md](./API_REFERENCE.md) - All 145+ API endpoints
