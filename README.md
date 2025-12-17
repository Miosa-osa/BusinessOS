# Business OS

**Your business operating system for the agentic era.**

AI-native. Self-hosted. Built for fast software.

---

## Overview

Business OS is a foundational operating system for the agentic era. Built for fast software creation where you own your data, control your AI, and customize everything.

### Key Principles

- **The Agentic Era** — AI agents that work FOR you, connecting your tools and operating on your terms
- **Fast Software** — Build and customize faster than ever. Ship changes in hours, not months
- **Your Data, Your Control** — Self-hosted by default. Nothing leaves without your permission

---

## Features

### Core Modules

| Module | Description |
|--------|-------------|
| **Dashboard** | Your daily command center with widgets for tasks, projects, and activity |
| **Projects** | Track work across your business with status, deadlines, and team assignments |
| **Tasks** | Kanban boards, list views, calendar views with priorities and due dates |
| **Team** | Org chart, capacity planning, workload management, and member profiles |
| **AI Chat** | Chat with AI using Focus Modes for specialized assistance |
| **Contexts** | Store business knowledge (people, businesses, projects) that AI can reference |
| **Documents** | Notion-like document editor with blocks, properties, and relations |
| **Clients** | Full CRM with profiles, contacts, deals pipeline, and interaction tracking |
| **Nodes** | Hierarchical business structure - your cognitive operating system |
| **Calendar** | Event management with Google Calendar integration |
| **Daily Log** | Track your day, patterns, and reflections |
| **Artifacts** | AI-generated documents: proposals, SOPs, frameworks, reports, code |
| **Desktop Mode** | macOS-inspired multi-window interface with customizable backgrounds |
| **Voice Notes** | Record and transcribe voice memos with AI |
| **Settings** | Configure AI models, providers, and integrations |

### AI Focus Modes

Five specialized modes for different types of work:

| Mode | Agent | Purpose | Options |
|------|-------|---------|---------|
| **Research** | Analysis Agent | Web/docs search | Scope (Web/Docs/All), Depth (Quick/Thorough), Output (Summary/Report) |
| **Analyze** | Analysis Agent | Data analysis | Approach (Validate/Compare/Forecast), Depth, Output (Findings/Dashboard) |
| **Write** | Document Agent | Content creation | Format (Doc/Slides/Spreadsheet), Mode (Step-by-step/First Draft) |
| **Build** | Planning Agent | Strategy & planning | Create (Framework/SOP/Plan), Detail (Outline/Detailed) |
| **Do More** | Orchestrator | General assistance | Mode (Chat/Learn/Brainstorm) |

### Nodes System

Nodes are the core organizational structure — a hierarchical system to manage different areas of your business:

- **Node Types**: Business, Project, Learning, Operational
- **Health Tracking**: Healthy, Needs Attention, Critical, Not Started
- **Features**: Purpose, Current Status, Weekly Focus, Decision Queue, Delegation Ready
- **Views**: Tree, List, Grid
- **Activation**: Set an active node to focus AI context on that area

### AI Agent System

Business OS includes a multi-agent architecture:

| Agent | Role |
|-------|------|
| **Orchestrator** | Main coordinator that handles requests and delegates to sub-agents |
| **DocumentAgent** | Creates business documents: proposals, SOPs, frameworks, agendas, reports |
| **AnalysisAgent** | Analyzes data, situations, and provides insights |
| **PlanningAgent** | Helps with planning, prioritization, and strategy |

### Desktop Mode

A macOS-inspired desktop environment:

- **Multi-Window Management**: Draggable, resizable windows with snap zones
- **Dock**: Quick access to apps with pinned items and window indicators
- **Menu Bar**: File, Edit, View, Window menus with keyboard shortcuts
- **Spotlight Search**: ⌘+Space for instant search and AI chat
- **Desktop Icons**: 15+ icon styles (macOS, Windows 95, neon, glassmorphism, etc.)
- **Customizable Backgrounds**: 50+ backgrounds (solid, gradient, pattern) + custom upload
- **Folders**: Create and organize desktop folders
- **Quick Chat**: Voice input from dock with auto-transcription

### Document Editor

A full-featured Notion-like block editor:

- **Block Types**: Paragraph, headings (H1-H3), bullet/numbered lists, todo, quote, code, divider, image, callout, table, embed, artifact
- **Slash Commands**: Type `/` to insert any block type
- **Properties**: Custom fields (text, select, multi-select, date, number, checkbox, URL, email, relation)
- **Relations**: Link documents to other contexts, projects, or clients
- **Templates**: Create reusable document templates
- **Public Sharing**: Generate shareable links with unique IDs
- **Auto-save**: Changes save automatically with debouncing

### Client Management (CRM)

Full client relationship management:

- **Client Profiles**: Company/individual info, status tracking (lead → prospect → active → churned)
- **Contacts**: Multiple contacts per client with primary designation
- **Deals Pipeline**: Kanban board with stages (Qualification → Proposal → Negotiation → Closed)
- **Interactions**: Log calls, emails, meetings, and notes
- **Project Linking**: Associate projects and contexts with clients

### Voice Features

- **Voice Notes**: Record voice memos with automatic transcription
- **Quick Chat Voice**: Record from dock, auto-transcribes and opens chat with transcript
- **Voice Input**: Microphone button in chat input for voice messages

---

## Tech Stack

### Frontend

| Component | Technology |
|-----------|------------|
| Framework | SvelteKit 2.0 |
| Language | TypeScript |
| Styling | TailwindCSS 4.x |
| State | Svelte 5 Runes (`$state`, `$derived`, `$effect`) |
| UI Components | Radix UI + bits-ui |
| Animation | Motion + svelte/transition |
| Auth | Better Auth |

### Backend

| Component | Technology |
|-----------|------------|
| Language | Go 1.25 |
| Framework | Gin Gonic |
| Database | PostgreSQL |
| Driver | pgx/v5 |
| Code Generation | SQLC |
| Auth | Better Auth integration |
| Config | Viper |

### AI Integration

| Provider | Type |
|----------|------|
| Ollama (Local) | Local LLMs (Qwen, Llama, Mistral, etc.) |
| Ollama Cloud | Cloud-hosted Ollama |
| Groq | Fast inference cloud |
| Anthropic | Claude models |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      YOUR SERVERS                           │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  PostgreSQL │  │   Ollama    │  │  Go Backend │        │
│  │  (Data)     │  │   (LLMs)    │  │  (Gin API)  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│                                                             │
│          🔒 Nothing leaves without your permission          │
└─────────────────────────────────────────────────────────────┘
                            ↕
                     [YOU CONTROL]
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                   EXTERNAL (Optional)                       │
│    Cloud LLMs (Groq/Anthropic)  •  Google Calendar  •  MCP │
└─────────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
BusinessOS/
├── backend-go/                  # Go backend (Gin + SQLC)
│   ├── cmd/server/
│   │   └── main.go              # Server entry point
│   ├── internal/
│   │   ├── config/              # Configuration (Viper)
│   │   ├── database/
│   │   │   ├── postgres.go      # Connection pool
│   │   │   ├── schema.sql       # Database schema
│   │   │   ├── queries/         # SQLC query definitions
│   │   │   └── sqlc/            # Generated Go code
│   │   ├── handlers/            # HTTP handlers (313 routes)
│   │   │   ├── chat.go          # Chat & AI endpoints
│   │   │   ├── contexts.go      # Documents & contexts
│   │   │   ├── clients.go       # CRM endpoints
│   │   │   ├── projects.go      # Project management
│   │   │   ├── commands.go      # Slash commands
│   │   │   └── ...
│   │   ├── middleware/          # Auth & CORS
│   │   └── services/            # LLM & integrations
│   ├── sqlc.yaml                # SQLC config
│   └── go.mod
│
├── frontend/                    # SvelteKit frontend
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/
│   │   │   │   └── client.ts    # API client (100+ methods)
│   │   │   ├── components/      # 140+ Svelte components
│   │   │   │   ├── ai-elements/ # AI UI (Message, Artifact, etc.)
│   │   │   │   ├── chat/        # Chat & Focus Modes
│   │   │   │   ├── desktop/     # Desktop mode (Window, Dock, etc.)
│   │   │   │   ├── editor/      # Block-based document editor
│   │   │   │   ├── clients/     # CRM components
│   │   │   │   └── ...
│   │   │   └── stores/          # State management
│   │   │       ├── windowStore.ts    # Desktop windows
│   │   │       ├── desktopStore.ts   # Desktop customization
│   │   │       ├── chat.ts           # Chat state
│   │   │       └── ...
│   │   └── routes/
│   │       ├── (app)/           # Protected routes
│   │       │   ├── chat/        # AI Chat with Focus Modes
│   │       │   ├── dashboard/   # Dashboard
│   │       │   ├── projects/    # Projects
│   │       │   ├── clients/     # CRM
│   │       │   ├── contexts/    # Documents
│   │       │   ├── nodes/       # Node tree
│   │       │   └── ...
│   │       ├── window/          # Desktop mode
│   │       └── popup-chat/      # Embedded chat widget
│   └── package.json
│
├── docs/
│   ├── ARCHITECTURE.md          # System architecture
│   ├── FRONTEND.md              # Frontend architecture
│   ├── BACKEND.md               # Backend architecture
│   └── FEATURES.md              # Feature documentation
│
└── README.md
```

---

## Getting Started

### Prerequisites

- Node.js 18+
- Go 1.21+
- PostgreSQL 15+
- Ollama (for local LLMs)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/robertohluna/BusinessOS.git
   cd BusinessOS
   ```

2. **Setup the database**
   ```bash
   # Create PostgreSQL database
   createdb business_os

   # Apply schema
   psql business_os < backend-go/internal/database/schema.sql
   ```

3. **Setup the backend**
   ```bash
   cd backend-go
   cp .env.example .env
   # Edit .env with your configuration

   # Generate SQLC code (if needed)
   sqlc generate

   # Run the server
   go run cmd/server/main.go
   ```

4. **Setup the frontend**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

5. **Start Ollama** (for local AI)
   ```bash
   ollama serve
   ollama pull qwen3:4b  # or your preferred model
   ```

6. **Open your browser**
   ```
   http://localhost:5173
   ```

---

## Configuration

### Backend Environment Variables

```env
# Server
PORT=8000
GIN_MODE=debug

# Database
DATABASE_URL=postgres://user:password@localhost:5432/business_os

# Better Auth
BETTER_AUTH_SECRET=your-secret-key

# Ollama (Local)
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=qwen3:4b

# Cloud Providers (Optional)
GROQ_API_KEY=your-groq-key
ANTHROPIC_API_KEY=your-anthropic-key
OLLAMA_CLOUD_API_KEY=your-ollama-cloud-key

# Google Calendar (Optional)
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
```

### Supported AI Models

**Local (Ollama):**
- Qwen 3 (4b, 8b, 14b, 30b, 32b)
- Llama 3.2/3.1 (various sizes)
- Mistral/Mixtral
- DeepSeek Coder
- Phi-3
- Gemma 2
- LLaVA (vision)

**Cloud:**
- Groq: Llama 3.3 70B, Llama 3.1 8B, Mixtral
- Anthropic: Claude Sonnet 4, Claude Opus 4
- Ollama Cloud: Qwen3 (235b, 480b)

---

## API Overview

The backend exposes 313 routes across these domains:

| Domain | Endpoints | Description |
|--------|-----------|-------------|
| `/api/chat` | 12 | Conversations, messages, AI analysis |
| `/api/artifacts` | 6 | Generated content management |
| `/api/contexts` | 15 | Documents, sharing, aggregation |
| `/api/projects` | 5 | Project CRUD |
| `/api/clients` | 20+ | CRM with contacts, deals, interactions |
| `/api/dashboard` | 10 | Focus items, tasks, summary |
| `/api/team` | 8 | Team member management |
| `/api/nodes` | 12 | Business structure tree |
| `/api/daily/logs` | 6 | Daily journal entries |
| `/api/calendar` | 8 | Events and sync |
| `/api/ai` | 15 | Models, providers, commands |
| `/api/voice-notes` | 6 | Recording and transcription |
| `/api/settings` | 3 | User preferences |
| `/api/usage` | 8 | Analytics and tracking |
| `/api/mcp` | 3 | MCP tool integration |

See [docs/BACKEND.md](docs/BACKEND.md) for full API documentation.

---

## Data Privacy

- **Self-hosted** — Runs entirely on your servers
- **Local LLMs** — Use AI without Big Tech watching
- **You decide** — Choose what connects externally
- **No data harvesting** — Your business data stays yours

---

## The Fast Software Philosophy

| Traditional SaaS | Fast Software |
|------------------|---------------|
| Request a feature | Need a feature? |
| Wait 6 months | Build it today |
| Maybe get it | Ship it tonight |

- Modify anything — it's your codebase
- Extend with your own features
- Integrate your existing tools
- Ship changes in hours, not months

---

## Documentation

- **[Architecture](docs/ARCHITECTURE.md)** — System architecture overview
- **[Frontend](docs/FRONTEND.md)** — Frontend architecture and components
- **[Backend](docs/BACKEND.md)** — Backend architecture and API reference
- **[Features](docs/FEATURES.md)** — Detailed feature documentation

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## License

MIT License — See [LICENSE](LICENSE) for details.

---

## Links

- **Repository**: [github.com/robertohluna/BusinessOS](https://github.com/robertohluna/BusinessOS)
- **OSA (OS Agent)**: [osa.dev](https://osa.dev)

---

<p align="center">
  <strong>Business OS</strong> — Your business operating system for the agentic era.
</p>
