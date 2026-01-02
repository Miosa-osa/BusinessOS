# Business OS

**Your business operating system for the agentic era.**

AI-native. Self-hosted. Built for fast software.

---

## Quick Start

```bash
# Clone and enter directory
git clone https://github.com/robertohluna/BusinessOS.git
cd BusinessOS

# Start all services
./startup.sh

# Open browser
open http://localhost:5173
```

### No Docker (UI / Dev Plumbing)

If you don't want Docker, you can run the frontend plus a **degraded backend** (no DB). This is useful for UI iteration and basic wiring.

```bash
# Frontend
npm --prefix frontend install
npm --prefix frontend run dev

# Backend (degraded mode)
go -C desktop/backend-go run ./cmd/server
```

Backend status endpoint: `http://localhost:8001/api/status`

**Requirements:** Node.js 18+, Go 1.21+, PostgreSQL 15+

---

## Overview

Business OS is a foundational operating system for the agentic era. Built for fast software creation where you own your data, control your AI, and customize everything.

### Key Principles

- **The Agentic Era** — AI agents that work FOR you, connecting your tools and operating on your terms
- **Fast Software** — Build and customize faster than ever. Ship changes in hours, not months
- **Your Data, Your Control** — Self-hosted by default. Nothing leaves without your permission

---

## Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│                         FRONTEND                            │
│              SvelteKit 2.0 + TypeScript + Tailwind          │
│                      http://localhost:5173                  │
└─────────────────────────────────────────────────────────────┘
                              │
                    Vite Proxy (/api/*)
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        GO BACKEND                           │
│              Gin + pgx/v5 + Better Auth                     │
│                      http://localhost:8001                  │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  REST API   │  │  WebSocket  │  │  Terminal   │        │
│  │  Handlers   │  │  Terminal   │  │  PTY (pty)  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       POSTGRESQL                            │
│                      localhost:5432                         │
│                    Database: business_os                    │
└─────────────────────────────────────────────────────────────┘
```
---

## Features

### Core Modules

| Module | Description |
|--------|-------------|
| **Dashboard** | Command center with widgets for tasks, projects, and activity |
| **Projects** | Track work with status, deadlines, and team assignments |
| **Tasks** | Kanban boards, list views, calendar views |
| **Team** | Org chart, capacity planning, workload management |
| **AI Chat** | Chat with AI using Focus Modes for specialized assistance |
| **Contexts** | Store business knowledge that AI can reference |
| **Documents** | Notion-like block editor with properties and relations |
| **Clients** | Full CRM with deals pipeline and interaction tracking |
| **Nodes** | Hierarchical business structure - your cognitive OS |
| **Calendar** | Event management with Google Calendar integration |
| **Terminal** | Real terminal with WebSocket backend (xterm.js + PTY) |
| **Desktop Mode** | macOS-inspired multi-window interface |

### Desktop Mode

A macOS-inspired desktop environment with:

- **Multi-Window Management** — Draggable, resizable windows with snap zones
- **Dock** — Quick access to apps with window indicators
- **Real Terminal** — Full PTY terminal via WebSocket (not a mock!)
- **Spotlight Search** — ⌘+Space for instant search
- **Custom Backgrounds** — 50+ options + custom upload

### Real Terminal Integration

The terminal is a **real shell** powered by:

- **Frontend**: xterm.js with fit, search, and web-links addons
- **Backend**: Go WebSocket handler with PTY (pseudo-terminal)
- **Features**: Resize support, session management, heartbeat keepalive
- **Auth**: Session-cookie authenticated via Better Auth

**🔒 Security Features (Phase 2 Complete):**
- **Container Hardening**: Read-only root filesystem with tmpfs, custom Seccomp profile blocking 15+ escape syscalls
- **Input Sanitization**: 28+ dangerous command patterns blocked (fork bombs, rm -rf /, container escapes)
- **Rate Limiting**: Token bucket algorithm with 100 msg/sec limit, 5 concurrent connections per user
- **Session Security**: IP binding, 8-hour max + 30-minute idle timeout, WebSocket origin validation

---

## Tech Stack

### Frontend (Port 5173)

| Component | Technology |
|-----------|------------|
| Framework | SvelteKit 2.0 |
| Language | TypeScript |
| Styling | TailwindCSS 4.x |
| State | Svelte 5 Runes |
| Terminal | xterm.js |
| Auth | Better Auth |

### Backend (Port 8001)

| Component | Technology |
|-----------|------------|
| Language | Go 1.25 |
| Framework | Gin Gonic |
| Database | PostgreSQL + pgx/v5 |
| WebSocket | gorilla/websocket |
| Terminal | creack/pty |
| Config | Viper |
| **Security** | **Docker containers + hardening** |
| **Testing** | **80 tests, 2,097 lines, 15+ benchmarks** |

### AI Integration

| Provider | Type |
|----------|------|
| Ollama (Local) | Local LLMs (Qwen, Llama, Mistral) |
| Groq | Fast cloud inference |
| Anthropic | Claude models |

---

## Project Structure

```text
BusinessOS/
├── desktop/
│   ├── backend-go/              # Go backend
│   │   ├── cmd/server/          # Entry point
│   │   ├── internal/
│   │   │   ├── config/          # Viper configuration
│   │   │   ├── container/       # Docker management + security hardening
│   │   │   ├── database/        # PostgreSQL + SQLC
│   │   │   ├── handlers/        # HTTP/WebSocket handlers
│   │   │   ├── logging/         # Sanitized logger with PII masking
│   │   │   ├── middleware/      # Auth middleware
│   │   │   ├── services/        # LLM + MCP services
│   │   │   └── terminal/        # PTY + rate limiting + input sanitization
│   │   └── go.mod
│   └── src/                     # Electron app (optional)
│
├── frontend/                    # SvelteKit frontend
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/             # API client
│   │   │   ├── components/      # 140+ components
│   │   │   ├── services/        # Terminal WebSocket service
│   │   │   └── stores/          # State management
│   │   └── routes/              # SvelteKit routes
│   ├── vite.config.ts           # Vite + proxy config
│   └── package.json
│
├── docs/                        # Documentation
├── startup.sh                   # Quick start script
└── README.md
```
---

## Configuration

### Environment Variables

```env
# Server
SERVER_PORT=8001

# Database
DATABASE_URL=postgresql://user@localhost:5432/business_os?sslmode=disable

# Auth
SECRET_KEY=your-secret-key

# AI (choose one or more)
OLLAMA_LOCAL_URL=http://localhost:11434
GROQ_API_KEY=your-groq-key
ANTHROPIC_API_KEY=your-anthropic-key

# Google Calendar (optional)
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
```

### Ports

| Service | Port | Description |
|---------|------|-------------|
| Frontend | 5173 | SvelteKit dev server |
| Backend | 8001 | Go API + WebSocket |
| PostgreSQL | 5432 | Database |
| Ollama | 11434 | Local LLM (optional) |

---

## Security Implementation

**Phase 2 Security Hardening Complete** — Production-ready container isolation and input validation.

### Container Security

| Feature | Implementation |
|---------|----------------|
| **Isolation** | Read-only root filesystem with tmpfs for /tmp, /var/tmp, /run |
| **Capabilities** | ALL capabilities dropped, only CHOWN + FOWNER allowed |
| **Syscalls** | Custom Seccomp profile blocks 15+ escape vectors (mount, setns, ptrace, bpf) |
| **Privileges** | no-new-privileges prevents setuid escalation |
| **Resources** | Memory: 512MB, CPU: 50%, PIDs: 100 max |
| **Network** | None (network isolation enabled) |

### Input Protection

| Layer | Coverage |
|-------|----------|
| **Command Filtering** | 28+ dangerous patterns: fork bombs, rm -rf /, container escapes |
| **Escape Sequences** | OSC 8 hyperlink injection, clipboard access, cursor manipulation |
| **Injection Prevention** | Null byte detection, length limits (4KB default) |
| **Rate Limiting** | Token bucket: 100 msg/sec, 20 burst, 5 connections/user |
| **Session Security** | IP binding, 8-hour max + 30-min idle timeout |

### Test Coverage

- **80 tests** across 8 test files (2,097 lines total)
- **15+ benchmarks** for performance validation
- **Container integration tests** with real Docker
- **Concurrency testing** (20-100 parallel operations)
- **Security pattern validation** for all 28+ dangerous commands

---

## API Overview

The backend exposes 100+ routes:

| Domain | Endpoints | Description |
|--------|-----------|-------------|
| `/api/chat` | 12 | Conversations, messages, AI |
| `/api/terminal` | 3 | WebSocket + session management |
| `/api/contexts` | 15 | Documents and sharing |
| `/api/clients` | 20+ | CRM with deals pipeline |
| `/api/projects` | 5 | Project CRUD |
| `/api/nodes` | 12 | Business structure tree |
| `/api/calendar` | 8 | Events and sync |
| `/api/auth` | 6 | Login, logout, session |

---

## Development

### Start Development Servers

```bash
# Option 1: Use startup script
./startup.sh

# Option 2: Manual start
# Terminal 1 - Backend
cd desktop/backend-go
DATABASE_URL="postgresql://user@localhost:5432/business_os?sslmode=disable" \
SERVER_PORT=8001 go run ./cmd/server

# Terminal 2 - Frontend
cd frontend
npm run dev
```

### Database Setup

```bash
createdb business_os
psql business_os < desktop/backend-go/internal/database/schema.sql
```

---

## Documentation

- **[DEVELOPMENT.md](DEVELOPMENT.md)** — Development guide and recent changes
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** — System architecture
- **[docs/FRONTEND.md](docs/FRONTEND.md)** — Frontend architecture
- **[docs/BACKEND.md](docs/BACKEND.md)** — Backend API reference

---

## License

MIT License — See [LICENSE](LICENSE) for details.

---

<p align="center">
  <strong>Business OS</strong> — Your business operating system for the agentic era.
</p>
