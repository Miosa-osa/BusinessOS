# Business OS - Architecture Documentation

> Generated: December 18, 2025

## Overview

**Business OS** is a personal command center / internal operations platform. It's a full-stack application with three deployment modes:

- **Web Mode**: SvelteKit frontend + Go backend deployed separately (Vercel + Cloud Run)
- **Desktop Mode**: Electron app bundling both frontend and backend
- **Development**: Local development with hot reload

---

## Tech Stack

### Backend (`/desktop/backend-go`)
| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Framework | Gin Gonic v1.9+ |
| Database | PostgreSQL + pgx/v5 |
| SQL | SQLC (type-safe code generation) |
| Config | Viper |
| Auth | Better Auth (cookie-based, frontend-driven) |

### Frontend (`/frontend`)
| Component | Technology |
|-----------|------------|
| Framework | SvelteKit 2.0 / Svelte 5 |
| Styling | TailwindCSS 4 |
| UI Components | bits-ui |
| Auth Client | better-auth/svelte |
| AI SDK | @ai-sdk/svelte |

### Desktop (`/desktop`)
| Component | Technology |
|-----------|------------|
| Framework | Electron Forge |
| Bundler | Vite |
| Backend | Embedded Go binary |
| Frontend | SvelteKit (built) |

---

## Project Structure

```
BusinessOS/
├── desktop/                    # Electron desktop app
│   ├── backend-go/             # Go backend (embedded in desktop)
│   │   ├── cmd/server/         # Server entry point
│   │   ├── internal/
│   │   │   ├── config/         # Configuration (Viper)
│   │   │   ├── database/       # PostgreSQL + SQLC
│   │   │   │   ├── schema.sql  # Database schema
│   │   │   │   ├── queries/    # SQLC query files
│   │   │   │   └── sqlc/       # Generated Go code
│   │   │   ├── handlers/       # HTTP handlers (22 files)
│   │   │   ├── middleware/     # Auth, CORS
│   │   │   ├── services/       # LLM, Google Calendar, etc.
│   │   │   ├── agents/         # Multi-agent system
│   │   │   ├── prompts/        # AI system prompts
│   │   │   └── tools/          # Artifact tools
│   │   ├── go.mod
│   │   └── sqlc.yaml
│   ├── src/
│   │   ├── main/               # Electron main process
│   │   ├── preload/            # Preload scripts
│   │   └── renderer/           # Built frontend (production)
│   ├── scripts/                # Build scripts
│   ├── forge.config.ts         # Electron Forge config
│   └── package.json
│
├── frontend/                   # SvelteKit frontend
│   ├── src/
│   │   ├── routes/             # SvelteKit pages
│   │   │   ├── (app)/          # Authenticated routes
│   │   │   ├── login/
│   │   │   ├── register/
│   │   │   └── api/
│   │   └── lib/
│   │       ├── api/            # API client (1600+ lines)
│   │       ├── components/     # 75+ Svelte components
│   │       ├── stores/         # State management
│   │       └── server/         # Server-side auth
│   ├── package.json
│   └── svelte.config.js
│
└── docs/                       # Documentation
```

---

## Deployment Modes

### 1. Web Deployment (Production)

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│     Vercel      │────▶│   Cloud Run     │────▶│   Cloud SQL     │
│   (Frontend)    │     │   (Go Backend)  │     │  (PostgreSQL)   │
│   SvelteKit     │     │   Port 8080     │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

- **Frontend**: Deployed to Vercel
- **Backend**: Deployed to Google Cloud Run
- **Database**: Google Cloud SQL (PostgreSQL)

### 2. Desktop App (Electron)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Electron Desktop App                          │
│  ┌────────────────────────┐  ┌────────────────────────────────┐ │
│  │   Renderer Process     │  │      Main Process              │ │
│  │   (SvelteKit Build)    │  │  ┌───────────────────────┐     │ │
│  │   - Desktop UI         │──│──│   Embedded Go Backend │     │ │
│  │   - Window management  │  │  │   - API server        │     │ │
│  │   - macOS-like UX      │  │  │   - SQLite/PostgreSQL │     │ │
│  └────────────────────────┘  │  └───────────────────────┘     │ │
│                              └────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

- Self-contained application
- Backend runs as subprocess
- Frontend served from local files

### 3. Development Mode

```bash
# Terminal 1: Backend
cd desktop/backend-go
go run cmd/server/main.go

# Terminal 2: Frontend
cd frontend
npm run dev
```

---

## Backend Architecture

### API Routers (`/desktop/backend-go/internal/handlers/`)

| Router | Prefix | Purpose |
|--------|--------|---------|
| `chat.go` | `/api/chat` | AI conversations, streaming |
| `contexts.go` | `/api/contexts` | Documents with blocks |
| `projects.go` | `/api/projects` | Project management |
| `clients.go` | `/api/clients` | CRM functionality |
| `deals.go` | `/api/deals` | Sales pipeline |
| `team.go` | `/api/team` | Team management |
| `nodes.go` | `/api/nodes` | Business structure |
| `dashboard.go` | `/api/dashboard` | Dashboard & tasks |
| `calendar.go` | `/api/calendar` | Calendar events |
| `ai_config.go` | `/api/ai` | AI provider config |
| `usage.go` | `/api/usage` | Usage analytics |
| `voice_notes.go` | `/api/voice-notes` | Voice transcription |

### Database (SQLC)

27 tables organized into domains:

**Core Tables**
- `contexts` - Documents with blocks, properties, sharing
- `conversations` - Chat conversations
- `messages` - Chat messages with metadata
- `projects` - Project management
- `tasks` - Task management
- `artifacts` - AI-generated content
- `nodes` - Business structure hierarchy

**CRM Tables**
- `clients` - Client profiles
- `client_contacts` - Client contacts
- `client_interactions` - Interaction history
- `client_deals` - Sales pipeline

**Team & User Tables**
- `team_members` - Team directory
- `user_settings` - User preferences
- `user_commands` - Custom slash commands

**Calendar & Logs**
- `calendar_events` - Calendar with Google sync
- `daily_logs` - Daily journal
- `voice_notes` - Voice transcriptions

### AI System

**Multi-provider support:**
- Ollama Local (for development/desktop)
- Ollama Cloud
- Anthropic (Claude)
- Groq

**Multi-agent system:**
- `OrchestratorAgent` - Coordinates tasks
- `DocumentAgent` - Creates documents
- `AnalysisAgent` - Data analysis
- `PlanningAgent` - Planning & prioritization

---

## Frontend Architecture

### Routes

| Route | Purpose |
|-------|---------|
| `/` | Landing page |
| `/login` | Login page |
| `/register` | Registration |
| `/dashboard` | Main dashboard |
| `/chat` | AI chat with focus modes |
| `/tasks` | Task management |
| `/projects` | Project management |
| `/clients` | CRM |
| `/contexts` | Documents |
| `/team` | Team management |
| `/nodes` | Business structure |
| `/calendar` | Calendar |
| `/settings` | User settings |
| `/window` | Desktop mode entry |

### Component Organization

```
components/
├── ai-elements/     # Chat message components
├── auth/            # Login, register forms
├── calendar/        # Calendar widgets
├── chat/            # Chat & Focus Modes
├── clients/         # CRM views
├── dashboard/       # Dashboard widgets
├── desktop/         # Desktop mode (Window, Dock, etc.)
├── editor/          # Block-based document editor
├── onboarding/      # Onboarding flows
├── tasks/           # Task components
├── team/            # Team components
└── ui/              # Shared UI primitives
```

### State Management

| Store | Purpose |
|-------|---------|
| `windowStore.ts` | Desktop window management |
| `desktopStore.ts` | Desktop customization |
| `chat.ts` | Chat & conversations |
| `auth.ts` | User session |
| `projects.ts` | Projects |
| `clients.ts` | CRM |
| `contexts.ts` | Documents |
| `team.ts` | Team members |
| `editor.ts` | Block editor |

---

## Desktop App Architecture

### Electron Forge Configuration

```typescript
// forge.config.ts
export default {
  packagerConfig: {
    asar: true,
    extraResource: ['./backend-go/server'],  // Embedded Go binary
  },
  makers: [
    { name: '@electron-forge/maker-squirrel' },  // Windows
    { name: '@electron-forge/maker-dmg' },       // macOS
    { name: '@electron-forge/maker-deb' },       // Linux
  ],
  plugins: [
    {
      name: '@electron-forge/plugin-vite',
      config: {
        main: './vite.main.config.ts',
        preload: './vite.preload.config.ts',
        renderer: './vite.renderer.config.ts',
      },
    },
  ],
};
```

### Build Process

```bash
# Build frontend
cd frontend && npm run build

# Copy to desktop renderer
cp -r build/* ../desktop/src/renderer/

# Build Go backend
cd desktop/backend-go
go build -o server cmd/server/main.go

# Package desktop app
cd desktop
npm run make
```

---

## Authentication

### Better Auth Flow

1. Frontend handles authentication via `better-auth/svelte`
2. Session stored in cookies (`better-auth.session_token`)
3. Backend validates session via middleware
4. User ID extracted and passed to handlers

```go
// middleware/auth.go
func AuthMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) {
        sessionToken, _ := c.Cookie("better-auth.session_token")
        session, _ := validateSession(pool, sessionToken)
        c.Set("user_id", session.UserID)
        c.Next()
    }
}
```

---

## Configuration

### Backend Environment Variables

```env
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/business_os

# Server
SERVER_PORT=8000

# AI Provider (ollama_local, ollama_cloud, anthropic, groq)
AI_PROVIDER=ollama_local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=llama3.2:3b

# Cloud AI
ANTHROPIC_API_KEY=sk-ant-xxx
GROQ_API_KEY=gsk_xxx

# Google OAuth
GOOGLE_CLIENT_ID=xxx
GOOGLE_CLIENT_SECRET=xxx

# CORS
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

### Frontend Environment Variables

```env
# API URL (development)
VITE_API_URL=http://localhost:8000/api

# Production (set in Vercel)
VITE_API_URL=https://api.businessos.app/api
```

---

## Key Features

### 1. AI Chat with Streaming
- Real-time streaming responses
- Multiple AI providers
- Focus modes (Research, Analyze, Write, Build)
- Conversation history

### 2. Document Editor
- Notion-like block editor
- Custom properties
- Document sharing
- Templates

### 3. Desktop Mode
- macOS-like windowing
- Dock with app shortcuts
- Spotlight search (⌘+Space)
- Window snapping

### 4. CRM
- Client profiles
- Contact management
- Deal pipeline
- Interaction tracking

### 5. Project Management
- Projects with status/priority
- Task boards
- Team assignment
- Notes and conversations

---

## Running the Application

### Development

```bash
# Backend
cd desktop/backend-go
go run cmd/server/main.go

# Frontend (separate terminal)
cd frontend
npm run dev
```

### Desktop App

```bash
# Build and run
cd desktop
npm start

# Package for distribution
npm run make
```

### Production

See [DEPLOYMENT.md](./DEPLOYMENT.md) for production deployment guide.

---

## Notes

- The app is designed for single-user or small team use
- Better Auth handles all authentication
- Streaming uses SSE (Server-Sent Events)
- Frontend uses Svelte 5 runes (`$state`, `$derived`, `$effect`)
- Desktop app embeds Go binary for self-contained deployment
