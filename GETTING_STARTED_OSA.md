# Getting Started with OSA Integration

**5-minute setup guide** for connecting BusinessOS with OSA-5.

---

## What You'll Build

Run a single command in BusinessOS to generate complete features.

```bash
osa generate "expense tracker with receipts"
```

OSA-5's 21 AI agents write the code. BusinessOS integrates it automatically.

---

## Prerequisites

Install these before starting.

### Required

- **Go 1.22+** - [Download](https://go.dev/dl/)
- **Node.js 20+** - [Download](https://nodejs.org/)
- **PostgreSQL 15+** - [Download](https://www.postgresql.org/download/)
- **Redis 7+** - Run `brew install redis` (Mac) or use Docker

### Optional

- **Docker** - For containerized terminals
- **jq** - For pretty JSON output: `brew install jq`

### Verify Installation

```bash
go version        # Should show 1.22 or higher
node --version    # Should show v20 or higher
psql --version    # Should show 15 or higher
redis-cli ping    # Should return PONG
```

---

## Step 1: Clone Repositories

Get both BusinessOS and OSA-5 code.

```bash
# BusinessOS (if not already cloned)
git clone https://github.com/your-org/BusinessOS-1.git
cd BusinessOS-1

# OSA-5 (separate terminal)
git clone https://github.com/your-org/OSA-5.git
cd OSA-5
```

---

## Step 2: Setup Database

Create PostgreSQL database and run migrations.

### Create Database

```bash
# Connect to PostgreSQL
psql postgres

# Create database
CREATE DATABASE businessos;

# Create user (optional)
CREATE USER businessos_user WITH PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE businessos TO businessos_user;

# Exit
\q
```

### Run Migrations

```bash
cd BusinessOS-1/desktop/backend-go

# Run all migrations
go run ./cmd/migrate
```

This creates all tables including OSA integration tables.

---

## Step 3: Configure Environment

Setup environment variables for both projects.

### BusinessOS Backend

Create `.env` in `desktop/backend-go/`:

```bash
cd /Users/ososerious/BusinessOS-1/desktop/backend-go
cp .env.example .env
```

Edit `.env` and add:

```bash
# Database (REQUIRED)
DATABASE_URL=postgresql://postgres:password@localhost:5432/businessos

# Server
SERVER_PORT=8001
ENVIRONMENT=development

# Security (REQUIRED - change these!)
SECRET_KEY=your-secret-key-change-this
TOKEN_ENCRYPTION_KEY=generate-with-openssl-rand-base64-32

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=your-redis-password

# AI Provider (use local Ollama for development)
AI_PROVIDER=ollama_local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=llama3.2:3b

# OSA Integration (REQUIRED for OSA)
OSA_ENABLED=true
OSA_BASE_URL=http://localhost:3003
OSA_TIMEOUT=30s
OSA_MAX_RETRIES=3
OSA_WEBHOOK_SECRET=your-webhook-secret-change-this
```

**What each variable does:**

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | Where your data lives |
| `SECRET_KEY` | Signs JWT tokens for auth |
| `TOKEN_ENCRYPTION_KEY` | Encrypts OAuth tokens |
| `OSA_ENABLED` | Turns on OSA features |
| `OSA_BASE_URL` | Where OSA-5 server runs |
| `OSA_WEBHOOK_SECRET` | Secures webhooks from OSA |

### BusinessOS Frontend

Create `.env` in `frontend/`:

```bash
cd /Users/ososerious/BusinessOS-1/frontend
echo "PUBLIC_API_URL=http://localhost:8001" > .env
```

### OSA-5

Create `.env` in OSA-5 root:

```bash
cd /path/to/OSA-5
cp .env.example .env
```

Edit `.env` and add:

```bash
# OSA Server
PORT=3003

# BusinessOS Integration
BUSINESSOS_URL=http://localhost:8001
WEBHOOK_SECRET=your-webhook-secret-change-this

# AI Provider (use same as BusinessOS)
AI_PROVIDER=ollama_local
OLLAMA_URL=http://localhost:11434
```

**Keep `WEBHOOK_SECRET` identical in both projects.**

---

## Step 4: Install Dependencies

Install packages for both projects.

### BusinessOS Backend

```bash
cd /Users/ososerious/BusinessOS-1/desktop/backend-go
go mod download
```

### BusinessOS Frontend

```bash
cd /Users/ososerious/BusinessOS-1/frontend
npm install
```

### OSA-5

```bash
cd /path/to/OSA-5
npm install
```

---

## Step 5: Start Services

Start all three services in separate terminals.

### Terminal 1: Redis

```bash
redis-server
```

Leave running. Output shows `Ready to accept connections`.

### Terminal 2: BusinessOS Backend

```bash
cd /Users/ososerious/BusinessOS-1/desktop/backend-go

# Build
go build -o bin/server ./cmd/server

# Run
./bin/server
```

Server starts on `http://localhost:8001`.

### Terminal 3: BusinessOS Frontend

```bash
cd /Users/ososerious/BusinessOS-1/frontend
npm run dev
```

Frontend starts on `http://localhost:5173`.

### Terminal 4: OSA-5 Server

```bash
cd /path/to/OSA-5
npm start
```

OSA-5 starts on `http://localhost:3003`.

---

## Step 6: Test Integration

Verify everything works.

### Check OSA Health

Open BusinessOS terminal at `http://localhost:5173` and run:

```bash
osa health
```

**Expected output:**

```
✅ OSA-5 is healthy
Version: 1.0.0
Status: Ready
```

If you see `command not found`, restart the terminal session.

### Generate Test Module

```bash
osa generate "simple todo list"
```

**Expected output:**

```
🚀 Starting generation...
App ID: app-abc-123
Track progress: osa status app-abc-123
```

### Check Status

```bash
osa status app-abc-123
```

**Expected output:**

```
App: simple todo list
Status: in_progress
Progress: 45%
Phase: testing
```

---

## Step 7: View Generated Code

OSA generates files in BusinessOS directories.

### Frontend Components

```bash
ls -la frontend/src/routes/(app)/todo-list/
```

You'll see:
- `+page.svelte` - UI component
- `+page.server.ts` - Server data loading
- `+page.ts` - Client-side logic

### Backend Handlers

```bash
ls -la desktop/backend-go/internal/handlers/todo_*.go
```

### Database Migrations

```bash
ls -la desktop/backend-go/internal/database/migrations/*todo*
```

---

## Troubleshooting

### `osa: command not found`

**Problem:** Terminal didn't load init script.

**Solution:** Restart terminal or run:

```bash
source /app/.bashrc
```

### `Connection refused` from OSA

**Problem:** OSA-5 not running.

**Solution:** Start OSA-5:

```bash
cd /path/to/OSA-5
npm start
```

### `{"enabled":false}` from health check

**Problem:** `OSA_ENABLED` not set.

**Solution:** Add to `.env`:

```bash
OSA_ENABLED=true
```

Restart backend.

### Database Connection Fails

**Problem:** PostgreSQL not running or wrong credentials.

**Solution:** Check PostgreSQL:

```bash
# Start PostgreSQL (Mac)
brew services start postgresql

# Verify connection
psql postgresql://postgres:password@localhost:5432/businessos
```

### Redis Connection Fails

**Problem:** Redis not running.

**Solution:** Start Redis:

```bash
# Mac with Homebrew
brew services start redis

# Or run directly
redis-server
```

### Port Already in Use

**Problem:** Another service on port 8001, 5173, or 3003.

**Solution:** Find and kill process:

```bash
# Find process on port
lsof -i :8001

# Kill process
kill -9 <PID>
```

---

## Next Steps

### Try More Commands

```bash
# List all workspaces
osa list

# Get help
osa help

# Generate complex module
osa generate "CRM with contacts, deals, and email integration"
```

### Customize OSA

Edit `OSA-5/.env` to:
- Change AI model
- Adjust timeout values
- Configure code generation style

### Build Production Features

Use OSA to generate:
- Admin dashboards
- Data import/export tools
- Custom reports
- Integration modules

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│  BusinessOS Terminal (Port 5173)                        │
│                                                         │
│  User types: osa generate "expense tracker"            │
│       ↓                                                 │
│  Shell function calls API                              │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│  BusinessOS Backend (Port 8001)                         │
│                                                         │
│  POST /api/internal/osa/generate                       │
│       ↓                                                 │
│  Circuit Breaker checks OSA health                     │
│       ↓                                                 │
│  Sends request to OSA-5                                │
│       ↓                                                 │
│  Stores workflow in PostgreSQL                         │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│  OSA-5 Server (Port 3003)                               │
│                                                         │
│  21 AI Agents work in parallel:                        │
│  • Architect Agent designs structure                   │
│  • Code Agent writes Svelte + Go                       │
│  • Test Agent creates tests                            │
│  • Integration Agent merges code                       │
│       ↓                                                 │
│  Sends webhooks back to BusinessOS                     │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│  Generated Files (Integrated into BusinessOS)          │
│                                                         │
│  frontend/src/routes/(app)/expense-tracker/            │
│  internal/handlers/expense_handler.go                  │
│  internal/database/migrations/042_expenses.sql         │
└─────────────────────────────────────────────────────────┘
```

---

## Support

### Check Logs

**Backend logs:**

```bash
tail -f desktop/backend-go/logs/server.log
```

**Frontend logs:**

Check browser console (F12).

**OSA-5 logs:**

Check terminal where OSA-5 is running.

### Common Issues

- Database not migrated → Run `go run ./cmd/migrate`
- Redis timeout → Check `REDIS_PASSWORD` matches
- OSA unreachable → Verify `OSA_BASE_URL` is correct
- Webhook auth fails → Match `OSA_WEBHOOK_SECRET` in both `.env` files

### Get Help

- Check full guide: `OSA_INTEGRATION_GUIDE.md`
- Review architecture: `docs/architecture/OSA_PHASE3_SUMMARY.md`
- Test manually: `curl http://localhost:8001/api/osa/health`

---

**You're ready!** Type `osa generate "your feature idea"` and watch AI build it.
