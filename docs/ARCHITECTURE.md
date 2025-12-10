# Business OS - Architecture Documentation

> Generated: December 10, 2025

## Overview

**Business OS** is a personal command center / internal operations platform. It's a full-stack application with:
- **Backend**: Python FastAPI with async PostgreSQL
- **Frontend**: SvelteKit 5 with TailwindCSS 4
- **Auth**: Better Auth (cookie-based sessions)
- **AI**: Ollama integration (local or cloud) with MCP (Model Context Protocol) support

---

## Tech Stack

### Backend (`/backend`)
| Component | Technology |
|-----------|------------|
| Framework | FastAPI 0.124 |
| Database | PostgreSQL + asyncpg |
| ORM | SQLAlchemy 2.0 (async) |
| Migrations | Alembic |
| Task Queue | Celery + Redis |
| AI/LLM | Ollama (local/cloud), MCP |
| Memory | Supermemory API |

### Frontend (`/frontend`)
| Component | Technology |
|-----------|------------|
| Framework | SvelteKit 2.48 / Svelte 5 |
| Styling | TailwindCSS 4 |
| UI Components | bits-ui |
| Auth Client | better-auth/svelte |
| AI SDK | @ai-sdk/svelte, ai |

---

## Project Structure

```
BusinessOS/
├── backend/
│   ├── app/
│   │   ├── main.py           # FastAPI app entry point
│   │   ├── config.py         # Pydantic settings
│   │   ├── database.py       # SQLAlchemy async setup
│   │   ├── models/           # SQLAlchemy models
│   │   ├── routers/          # API route handlers
│   │   ├── schemas/          # Pydantic schemas
│   │   ├── services/         # Business logic (Ollama, MCP)
│   │   └── utils/            # Auth utilities
│   ├── alembic/              # Database migrations
│   ├── requirements.txt
│   └── .env.example
│
├── frontend/
│   ├── src/
│   │   ├── routes/           # SvelteKit pages
│   │   │   ├── (app)/        # Authenticated routes
│   │   │   │   ├── dashboard/
│   │   │   │   ├── chat/
│   │   │   │   ├── tasks/
│   │   │   │   ├── projects/
│   │   │   │   ├── team/
│   │   │   │   ├── contexts/
│   │   │   │   └── daily/
│   │   │   ├── login/
│   │   │   ├── register/
│   │   │   └── api/
│   │   └── lib/
│   │       ├── components/   # Svelte components
│   │       ├── stores/       # Svelte stores
│   │       ├── auth-client.ts
│   │       └── server/
│   ├── package.json
│   └── svelte.config.js
│
└── docs/                     # This documentation
```

---

## Backend Architecture

### API Routers (`/backend/app/routers/`)

| Router | Prefix | Purpose |
|--------|--------|---------|
| `chat.py` | `/api/chat` | AI conversations, streaming responses |
| `dashboard.py` | `/api/dashboard` | Focus items, tasks, summary data |
| `projects.py` | `/api/projects` | Project CRUD |
| `contexts.py` | `/api/contexts` | Context profiles (people, businesses) |
| `team.py` | `/api/team` | Team member management |
| `mcp.py` | `/api/mcp` | MCP tool listing/execution |

### Data Models (`/backend/app/models/`)

| Model | Description |
|-------|-------------|
| `Conversation` | Chat conversations with messages |
| `Message` | Individual chat messages (user/assistant) |
| `Project` | Projects with status, priority, metadata |
| `ProjectNote` | Notes attached to projects |
| `Task` | Tasks with due dates, priorities |
| `FocusItem` | Daily focus items |
| `Context` | Context profiles (system prompts, content) |
| `TeamMember` | Team members with activities |
| `Artifact` | Generated artifacts (code, docs) |
| `Node` | Graph nodes with metrics |
| `DailyLog` | Daily log entries |

**Note**: User model is managed by Better Auth (external `user` and `session` tables).

### Services (`/backend/app/services/`)

#### OllamaService (`ollama.py`)
- Supports **local** (`http://localhost:11434`) or **cloud** mode
- Streaming chat completions
- Default model: `deepseek-r1:70b`
- Pre-defined system prompts for different contexts:
  - `default` - General assistant
  - `daily_planning` - Daily planning helper
  - `project_analysis` - Project context analysis
  - `strategic_thinking` - Strategic thinking partner
  - `code_review` - Code review assistant

#### MCP Server (`mcp_server.py`)
- Model Context Protocol implementation
- Built-in tools:
  - `search_conversations` - Search past conversations
  - `get_project_context` - Get project details
  - `create_artifact` - Create code/document artifacts
  - `add_to_daily_log` - Add daily log entries
  - `get_context_profile` - Get context profiles
- Custom tool registration support

### Authentication (`/backend/app/utils/auth.py`)
- Session validation via Better Auth cookies
- Cookie name: `better-auth.session_token`
- Validates against `session` table, joins with `user` table
- Returns `BetterAuthUser` dataclass with user info

---

## Frontend Architecture

### Routes

| Route | Purpose |
|-------|---------|
| `/` | Landing page |
| `/login` | Login page |
| `/register` | Registration page |
| `/dashboard` | Main dashboard (focus, tasks, projects) |
| `/chat` | AI chat interface |
| `/tasks` | Task management |
| `/projects` | Project management |
| `/team` | Team management |
| `/contexts` | Context profiles |
| `/daily` | Daily log |

### Component Organization (`/frontend/src/lib/components/`)

```
components/
├── auth/        # Login, register forms
├── chat/        # Chat UI components
├── dashboard/   # Dashboard widgets
├── onboarding/  # Onboarding flow
├── tasks/       # Task components
└── team/        # Team components
```

### Authentication Flow
1. Frontend uses `better-auth/svelte` client
2. Session stored in cookies with `credentials: 'include'`
3. Protected routes check session in `(app)/+layout.svelte`
4. Redirects to `/login` if no session

---

## Configuration

### Backend Environment Variables (`.env`)

```env
# Database
DATABASE_URL=postgresql+asyncpg://postgres:password@localhost:5432/business_os

# JWT (legacy, Better Auth handles auth now)
SECRET_KEY=your-secret-key
ALGORITHM=HS256
ACCESS_TOKEN_EXPIRE_MINUTES=1440

# Ollama
OLLAMA_MODE=local              # "local" or "cloud"
OLLAMA_LOCAL_URL=http://localhost:11434
OLLAMA_CLOUD_URL=https://api.ollama.com/v1
OLLAMA_CLOUD_API_KEY=

# Default Model
DEFAULT_MODEL=deepseek-r1:70b

# Redis (for Celery)
REDIS_URL=redis://localhost:6379/0

# Supermemory
SUPERMEMORY_API_KEY=
```

### CORS Configuration
Allowed origins:
- `http://localhost:5173` (SvelteKit dev)
- `http://localhost:5174`
- `http://localhost:3000`

---

## Key Features

### 1. AI Chat with Streaming
- Real-time streaming responses from Ollama
- Conversation history persistence
- Context-aware system prompts
- Conversation search

### 2. Dashboard
- **Focus Items**: Daily priorities (3 items)
- **Tasks**: Due dates, priorities, project links
- **Projects**: Active projects with health status
- **Activities**: Recent conversation/team activities

### 3. Project Management
- Status tracking (active, paused, completed, archived)
- Priority levels (critical, high, medium, low)
- Client association
- Notes and conversation linking

### 4. Team Management
- Team member profiles
- Activity tracking
- Role management

### 5. Context Profiles
- Custom system prompts
- Context content for AI
- Reusable across conversations

### 6. MCP Integration
- Tool-use capabilities for AI
- Extensible tool system
- Built-in business tools

---

## Database Schema

The app uses PostgreSQL with the following key tables:

- `user` - Managed by Better Auth
- `session` - Managed by Better Auth
- `conversations` - Chat conversations
- `messages` - Chat messages
- `projects` - Projects
- `project_notes` - Project notes
- `project_conversations` - Project-conversation links
- `tasks` - Tasks
- `focus_items` - Daily focus items
- `contexts` - Context profiles
- `team_members` - Team members
- `team_member_activities` - Team activities
- `artifacts` - Generated artifacts
- `artifact_versions` - Artifact version history
- `nodes` - Graph nodes
- `node_metrics` - Node metrics
- `daily_logs` - Daily log entries

---

## Running the Application

### Backend
```bash
cd backend
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
cp .env.example .env  # Configure your settings
uvicorn app.main:app --reload
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Prerequisites
- PostgreSQL database
- Redis (for Celery tasks)
- Ollama (local) or Ollama Cloud API key

---

## TODOs / Incomplete Features

Based on code comments:
- MCP tool handlers are stubbed (TODO implementations)
- Full-text search for conversations (currently using ILIKE)
- Dynamic MCP resource listing
- Team member count per project (hardcoded to 1)

---

## Notes

- The app is designed for a single user ("Roberto") based on system prompts
- Better Auth handles all authentication - no custom JWT implementation
- Streaming responses use FastAPI's `StreamingResponse`
- Frontend uses Svelte 5 runes (`$state`, `$effect`, `$props`)
