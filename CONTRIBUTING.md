# Contributing to BusinessOS

> **Status:** SHELL — fill in sections marked TODO
> **Owner:** Team
> **Priority:** P1 (should have within 1 week of launch)

---

## Getting Started

### Prerequisites

- Go 1.24.1+
- Node.js 20+
- PostgreSQL 15+
- Redis 7+ (optional, for multi-instance mode)
- Docker (for terminal system)

### Setup

```bash
# Clone
git clone https://github.com/Miosa-osa/BusinessOS.git
cd BOS

# Backend
cd desktop/backend-go
cp .env.production.example .env  # Edit with your config
go mod download
go build ./cmd/server

# Frontend
cd frontend
npm install
npm run dev

# Desktop (Electron)
cd desktop
npm install
npm start
```

See [docs/development/DEVELOPER_QUICKSTART.md](docs/development/DEVELOPER_QUICKSTART.md) for full setup.

## Development Workflow

### Branches

```
main              Production-ready code
feature/*         New features
fix/*             Bug fixes
```

### Making Changes

1. Create a branch from `main`
2. Make your changes
3. Run validation:
   ```bash
   cd desktop/backend-go && go build ./cmd/server && go test ./...
   cd frontend && npm run build && npm run check
   ```
4. Commit with conventional format: `type(scope): description`
5. Open a PR using the PR template
6. Get review approval
7. Merge to `main`

### Commit Format

```
feat(auth): add Google OAuth callback
fix(osa): resolve FK violation in app generation
refactor(sorx): extract skill matcher into separate module
docs(architecture): update OSA system status
test(orchestration): add mode classification edge cases
```

### Code Standards

- **Go:** Follow existing patterns. Use `slog` for logging. Handler → Service → Repository layers.
- **TypeScript:** Strict mode. Avoid `any`. Follow SvelteKit conventions.
- **Tests:** Write tests for new functionality. Maintain 80%+ coverage on critical paths.
- **Security:** No hardcoded credentials. Parameterized queries. Input validation on boundaries.

See [CLAUDE.md](CLAUDE.md) for full coding conventions.

## Agent Dispatch Workflow

If you're dispatching Claude Code agents (Sprint 1+):

1. Read the [TEAM-BRIEFING.md](docs/agent-dispatch/TEAM-BRIEFING.md)
2. Read YOUR operator's `DISPATCH.md` in `docs/agent-dispatch/sprints/`
3. Follow the human oversight checkpoints
4. Validate after every agent completes

## Architecture

- [OSA-SYSTEM-STATUS.md](docs/architecture/OSA-SYSTEM-STATUS.md) — What's built
- [MIOSA-PRODUCTION-ROADMAP.md](docs/architecture/MIOSA-PRODUCTION-ROADMAP.md) — Where we're going
- [PRODUCT-READINESS-ARCHITECTURE.md](docs/architecture/PRODUCT-READINESS-ARCHITECTURE.md) — Full file tree

## Reporting Issues

- **Bugs:** Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md)
- **Features:** Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md)
- **Security:** See [SECURITY.md](SECURITY.md) — do NOT open public issues for vulnerabilities

---

**Last Updated:** 2026-02-23
