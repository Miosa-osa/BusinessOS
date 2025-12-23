# BusinessOS Go Backend

Production-ready Go backend with real PTY terminal, container isolation, and security hardening.

---

## Quick Start

```bash
# Run with environment variables
DATABASE_URL="postgresql://user@localhost:5432/business_os?sslmode=disable" \
SERVER_PORT=8001 \
go run ./cmd/server
```

---

## Architecture

```text
cmd/server/           Entry point
internal/
  config/             Viper configuration
  container/          Docker management + security hardening
  database/           PostgreSQL + SQLC
  handlers/           HTTP/WebSocket handlers (145 endpoints)
  logging/            Sanitized logger with PII masking
  middleware/         Auth middleware
  services/           LLM + MCP services
  terminal/           PTY + rate limiting + input sanitization
```

---

## Security Implementation (Phase 2)

### Container Hardening

| Feature | Implementation |
|---------|----------------|
| Root filesystem | Read-only with tmpfs |
| Capabilities | ALL dropped, only CHOWN + FOWNER |
| Syscalls | Custom Seccomp blocks 15+ escape vectors |
| Privileges | no-new-privileges prevents setuid |
| Resources | 512MB RAM, 50% CPU, 100 PIDs |
| Network | Isolated (NetworkMode: none) |

### Input Sanitization

28+ dangerous command patterns blocked:

| Category | Examples |
|----------|----------|
| Fork bombs | `:(){ :|:& };:` |
| Destructive | `rm -rf /`, `dd if=/dev/zero` |
| Escape attempts | `nsenter`, `unshare`, `chroot` |
| Privilege escalation | `sudo`, `su root` |
| Reverse shells | `bash -i`, `/dev/tcp` |
| ANSI injection | OSC 8, clipboard, cursor |

### Rate Limiting

| Parameter | Value |
|-----------|-------|
| Messages/second | 100 |
| Burst capacity | 20 |
| Connections/user | 5 |
| Algorithm | Token bucket |

### Session Security

| Feature | Value |
|---------|-------|
| IP binding | Enabled |
| Max session | 8 hours |
| Idle timeout | 30 minutes |
| Origin validation | Configured origins |

---

## Test Coverage

| File | Tests | Lines |
|------|-------|-------|
| sanitizer_test.go | 12+ | 478 |
| ratelimit_test.go | 20+ | 423 |
| session_test.go | 48+ | 1196 |
| **Total** | **80** | **2,097** |

### Benchmarks

| Operation | Time | Memory |
|-----------|------|--------|
| AllowMessage | 285 ns | 0 B |
| QuickValidate | 468 ns | 0 B |
| FullValidation | 6.6 us | 725 B |
| ValidateIP | 162 ns | 80 B |

Run tests:
```bash
go test -v ./internal/terminal/...
go test -bench=. -benchmem ./internal/terminal/...
```

---

## Terminal API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/terminal/ws` | WebSocket | PTY connection |
| `/api/terminal/sessions` | GET | List sessions |
| `/api/terminal/sessions/:id` | DELETE | Close session |

### WebSocket Parameters

| Param | Default | Description |
|-------|---------|-------------|
| cols | 80 | Terminal columns |
| rows | 24 | Terminal rows |
| shell | zsh | Shell to spawn |
| cwd | ~ | Working directory |

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_PORT | 8001 | HTTP port |
| DATABASE_URL | required | PostgreSQL connection |
| SECRET_KEY | required | Session encryption |
| ALLOWED_ORIGINS | localhost | CORS origins |
| GIN_MODE | debug | Gin mode (release for prod) |

---

## Dependencies

| Package | Purpose |
|---------|---------|
| gin-gonic/gin | HTTP framework |
| jackc/pgx/v5 | PostgreSQL driver |
| gorilla/websocket | WebSocket handling |
| creack/pty | PTY management |
| docker/docker | Container API |
| spf13/viper | Configuration |

---

## Documentation

| Document | Location |
|----------|----------|
| API Reference | `/docs/API_REFERENCE.md` |
| Deployment Guide | `/docs/DEPLOYMENT_GUIDE.md` |
| Full Security Docs | `/docs/BUSINESSOS_TERMINAL_SYSTEM_DOCUMENTATION.md` |
| Seccomp Profile | `/docker/seccomp-terminal.json` |
