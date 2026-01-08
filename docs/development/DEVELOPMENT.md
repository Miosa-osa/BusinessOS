# Development Guide

This document covers recent development work, architecture decisions, and technical implementation details.

---

## Recent Changes (December 2024)

### Terminal WebSocket Integration

Replaced the mock terminal with a real PTY-based terminal that executes actual shell commands.

#### Architecture

```text
┌──────────────────────────────────────────────────────────────────┐
│                         FRONTEND                                  │
│  Terminal.svelte → TerminalService → WebSocket → Vite Proxy      │
│      (xterm.js)      (TypeScript)     (ws://)     (/api/*)       │
└──────────────────────────────────────────────────────────────────┘
                                │
                    ws://localhost:5173/api/terminal/ws
                    (Vite proxies to backend:8001)
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│                          BACKEND                                  │
│  WebSocketHandler → Manager → Session → PTY                       │
│   (gorilla/ws)       (Go)     (Go)   (creack/pty)                │
└──────────────────────────────────────────────────────────────────┘
```

#### Key Files

| File | Purpose |
|------|---------|
| `frontend/src/lib/components/desktop/Terminal.svelte` | xterm.js UI component |
| `frontend/src/lib/services/terminal.service.ts` | WebSocket client service |
| `frontend/vite.config.ts` | Proxy config with `ws: true` |
| `desktop/backend-go/internal/handlers/terminal.go` | HTTP/WebSocket handler |
| `desktop/backend-go/internal/terminal/manager.go` | Session management |
| `desktop/backend-go/internal/terminal/websocket.go` | WebSocket protocol |
| `desktop/backend-go/internal/terminal/pty.go` | PTY spawning |
| `desktop/backend-go/internal/terminal/session.go` | Session state |

#### WebSocket Protocol

Messages are JSON with this structure:

```typescript
interface TerminalMessage {
  type: 'input' | 'output' | 'resize' | 'heartbeat' | 'error' | 'status';
  session_id?: string;
  data?: string;
  metadata?: Record<string, unknown>;
}
```

**Message Types:**
- `input` — User keystrokes sent to PTY
- `output` — PTY output sent to xterm.js
- `resize` — Terminal dimensions changed
- `heartbeat` — Keep-alive ping (every 30s)
- `status` — Connection status (e.g., `connected`)
- `error` — Error messages

---

### Bug Fix: Authentication Context Key

**Problem:** Terminal WebSocket returned 401 even for authenticated users.

**Root Cause:** The auth middleware stored the user under key `"user"`, but the terminal handler looked for `"user_id"`.

```go
// middleware/auth.go (line 80)
c.Set(UserContextKey, &user)  // UserContextKey = "user"

// handlers/terminal.go (BEFORE - broken)
userID, exists := c.Get("user_id")  // Wrong key!
```

**Fix:** Updated terminal handler to use `middleware.GetCurrentUser(c)`:

```go
// handlers/terminal.go (AFTER - fixed)
user := middleware.GetCurrentUser(c)
if user == nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
    return
}
h.wsHandler.HandleConnection(c.Writer, c.Request, user.ID)
```

---

### Port Migration: 8000 → 8001

The backend now runs on **port 8001** (Docker uses 8000).

**Files Updated:**
- `frontend/vite.config.ts` — All proxy targets
- `frontend/src/lib/api/client.ts` — API base URL
- `frontend/src/lib/auth-client.ts` — Auth API URL
- `frontend/src/lib/services/terminal.service.ts` — WebSocket fallback URL

**Vite Proxy Configuration:**

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api/terminal': {
      target: 'http://localhost:8001',
      changeOrigin: true,
      ws: true,  // Enable WebSocket proxying
    },
    '/api/auth': { target: 'http://localhost:8001', changeOrigin: true },
    '/api/chat': { target: 'http://localhost:8001', changeOrigin: true },
    // ... all other /api/* routes
  }
}
```

---

## Architecture Decisions

### Why WebSocket for Terminal?

1. **Bidirectional** — Real-time I/O between browser and PTY
2. **Low latency** — No HTTP overhead per keystroke
3. **Persistent** — Single connection for entire session
4. **Standard** — Works through Vite proxy with `ws: true`

### Why creack/pty?

The `github.com/creack/pty` package provides cross-platform PTY support:
- Spawns real shell processes (zsh, bash, sh)
- Handles terminal resize (SIGWINCH)
- Works on macOS and Linux

### Session Management

Each terminal connection creates a `Session` with:
- Unique session ID (UUID)
- User ID (from auth)
- PTY file descriptor
- Created timestamp
- Shell type and working directory

Sessions are tracked by the `Manager` for cleanup on disconnect.

---

## Development Workflow

### Starting Services

```bash
# Quick start (recommended)
./startup.sh

# Manual start
# Terminal 1: Database
brew services start postgresql@14

# Terminal 2: Backend
cd desktop/backend-go
DATABASE_URL="postgresql://user@localhost:5432/business_os?sslmode=disable" \
SERVER_PORT=8001 go run ./cmd/server

# Terminal 3: Frontend
cd frontend
npm run dev
```

### Testing Terminal

1. Open http://localhost:5173
2. Sign in
3. Open Desktop mode → Terminal window
4. Should see "Connected" status and working shell

### Debugging WebSocket

```bash
# Check if backend terminal endpoint is accessible
curl -s http://localhost:8001/api/terminal/sessions
# Should return 401 (auth required) or 200 with session list

# Check Vite is proxying correctly
# In browser console: ws://localhost:5173/api/terminal/ws should connect
```

---

## Common Issues

### Terminal shows "Connecting..." forever

**Cause:** Backend not running or WebSocket proxy misconfigured.

**Fix:**
1. Ensure backend is running: `curl http://localhost:8001/health`
2. Check Vite config has `ws: true` for `/api/terminal`

### Terminal shows 401 Unauthorized

**Cause:** Not logged in, or session cookie not sent.

**Fix:**
1. Sign in at http://localhost:5173
2. Refresh and try terminal again

### WebSocket connection refused

**Cause:** Backend crashed or wrong port.

**Fix:**
1. Check backend logs: `tail -f /tmp/backend.log`
2. Restart backend with correct `SERVER_PORT=8001`

---

## File Structure Reference

```text
desktop/backend-go/internal/terminal/
├── manager.go      # Session lifecycle management
├── pty.go          # PTY spawning and I/O
├── session.go      # Session state struct
└── websocket.go    # WebSocket upgrade and protocol

frontend/src/lib/services/
└── terminal.service.ts  # WebSocket client class
```

---

## API Endpoints

### Terminal Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/terminal/ws` | WebSocket upgrade for terminal I/O |
| GET | `/api/terminal/sessions` | List active sessions for user |
| DELETE | `/api/terminal/sessions/:id` | Close a terminal session |

All terminal routes require authentication via `better-auth.session_token` cookie.

---

## Dependencies Added

### Frontend

```json
{
  "@xterm/xterm": "^5.5.0",
  "@xterm/addon-fit": "^0.10.0",
  "@xterm/addon-search": "^0.15.0",
  "@xterm/addon-web-links": "^0.11.0"
}
```

### Backend

```go
require (
    github.com/creack/pty v1.1.21
    github.com/gorilla/websocket v1.5.1
)
```
