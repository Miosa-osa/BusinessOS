# Production Go-Live Checklist

> **Every step required to take BusinessOS from local development to live production.**
> Covers: Web (Railway backend + Cloudflare Pages frontend), Desktop (Electron), environment rotation, database migration, secrets, monitoring, rollback.
>
> Last Updated: 2026-10-06
> Status: CHECKLIST (execute in order, top to bottom)

---

## Phase 0: Pre-Flight (Before Anything Else)

### Code Readiness

- [ ] All Sprint 1 bug fixes merged (FK violation, prompt overwrite, SSE routes, hardcoded URLs)
- [ ] `cd desktop/backend-go && go build ./cmd/server` -> EXIT 0
- [ ] `cd desktop/backend-go && go test ./...` -> ALL PASS
- [ ] `cd frontend && npm run build` -> EXIT 0
- [ ] No hardcoded credentials in codebase (run: `grep -r "sk-ant\|password.*=.*\"[^\"]*\"" --include="*.go" --include="*.ts" desktop/backend-go/ frontend/src/`)
- [ ] No `localhost` URLs in production code paths
- [ ] All environment variables use `os.Getenv()` or config struct (no hardcoded values)

### Tooling

- [ ] Railway CLI authenticated: `railway whoami`
- [ ] Railway project linked: `railway status --service businessos-api` (project `BusinessOS`, environment `production`)
- [ ] Wrangler authenticated: `npx wrangler whoami`

### Platform Provisioning (Railway)

- [ ] Railway project `BusinessOS` exists in the `Roberto Luna's Projects` workspace
- [ ] Service `businessos-api` exists (built from `desktop/backend-go/Dockerfile`)
- [ ] Railway Postgres service `Postgres` exists in the same project
- [ ] Billing / usage limits configured on the workspace

---

## Phase 1: Database (Railway Postgres)

### Provision

- [ ] Postgres service added to project `BusinessOS` (Railway dashboard > New > Database > PostgreSQL)
- [ ] `DATABASE_URL` is referenced by the `businessos-api` service (Railway provides it as a service variable)

### Run Migrations

```bash
# Open a psql shell against the Railway Postgres (proxied by the CLI)
railway connect Postgres --environment production

# Apply pending migrations from desktop/backend-go/internal/database/migrations/
# (the backend can also auto-apply on boot, but apply explicitly before a deploy
# so a bad migration fails before the new revision serves traffic)
```

### Checklist

- [ ] Railway Postgres running
- [ ] All migrations applied (including `090_tenant_org_foundation.sql`)
- [ ] `pgvector` extension enabled: `CREATE EXTENSION IF NOT EXISTS vector;`
- [ ] Superadmin row present: `select email, role from users where email = 'roberto@businessos.dev';`
- [ ] Connection string reachable from the backend (`/health/detailed` reports `database: connected`)

---

## Phase 2: Redis (Optional)

Redis is used for session caching and rate limiting. It is optional: `/health/detailed` reports
`redis: not_configured` when it is absent, and the backend runs without it.

### Options

- Railway Redis service in the same project (Railway dashboard > New > Database > Redis), or
- Managed Redis from Upstash.

### Checklist

- [ ] Redis instance created (if used)
- [ ] `REDIS_URL` / `REDIS_PASSWORD` set on the `businessos-api` service
- [ ] Verified connection: `redis-cli -h HOST -p PORT PING` -> `PONG`

---

## Phase 3: Backend Deployment (Railway)

### Deploy

Preferred path is CI (`.github/workflows/deploy-backend.yml`): push to `main` with changes
under `desktop/backend-go/**`; the workflow runs the Go tests then `railway up`. To deploy by
hand, stage a copy outside the git tree (Railway prunes uploads with the repo-root
`.gitignore`, which drops `cmd/server`) and upload it:

```bash
STAGE="$(mktemp -d)/businessos-backend"
mkdir -p "$STAGE" && cp -R desktop/backend-go/. "$STAGE"/
railway up "$STAGE" --path-as-root --no-gitignore --service businessos-api --ci
```

### Verify Backend

```bash
# Health check
curl -fsS https://businessos-api-production.up.railway.app/health
# Expected: 200

# Detailed health (database connectivity)
curl -fsS https://businessos-api-production.up.railway.app/health/detailed
# Expected: {"components":{"database":{"status":"connected"},...},"status":"healthy"}
```

### Checklist

- [ ] Backend deployed to Railway
- [ ] Database connection working (`/health/detailed`)
- [ ] Redis connection working (if configured)
- [ ] Health endpoint returns 200
- [ ] CORS (`ALLOWED_ORIGINS`) set to production domains only (no wildcard)
- [ ] Environment variables set in the Railway Variables tab (no hardcoded secrets)

---

## Phase 4: Frontend Deployment (Cloudflare Pages)

```bash
cd frontend
CLOUDFLARE_BUILD=true npm run build
npx wrangler pages deploy build --project-name=businessos-5 --branch=main --commit-dirty=true
```

The project is `businessos-5` (serves `businessos.dev` and `app.businessos.dev`). `VITE_API_URL`
and `VITE_BACKEND_URL` stay UNSET: the app calls same-origin `/api/*`, which the Pages Function
(`frontend/functions/api/[[path]].js`) proxies to Railway.

### Checklist

- [ ] Cloudflare Pages project `businessos-5` exists and is the production project
- [ ] Build output dir is `build`
- [ ] `BUSINESSOS_BACKEND_URL` set (optional override; defaults to the Railway URL)
- [ ] Frontend loads without console errors
- [ ] Proxy works: `curl -fsS https://businessos.dev/api/v1/health`
- [ ] Auth flow works (login -> redirect -> dashboard)

---

## Phase 5: DNS & Custom Domain (Cloudflare)

`businessos.dev` is on Cloudflare. The Pages project and the R2 downloads bucket attach as
custom domains in the same zone, so DNS records are created automatically.

| Host | Target |
|------|--------|
| `businessos.dev` | Cloudflare Pages project `businessos-5` |
| `app.businessos.dev` | Cloudflare Pages project `businessos-5` |
| `downloads.businessos.dev` | Cloudflare R2 bucket `businessos-downloads` |

### Checklist

- [ ] `https://businessos.dev` loads the frontend
- [ ] `https://app.businessos.dev` loads the frontend
- [ ] `https://businessos.dev/api/v1/health` proxies to the backend
- [ ] `api.businessos.dev` is intentionally NOT routed (the backend is reached through `/api/*`)
- [ ] SSL certificates provisioned by Cloudflare (automatic)

---

## Phase 6: Auth Configuration (Production)

### Google OAuth

In Google Cloud Console -> APIs & Services -> Credentials, the Web application OAuth client
must list the production redirect URI:

- `https://businessos.dev/api/v1/auth/oauth/google/callback`

(Localhost redirects are already authorized, so local dev + Google works.)

### Backend Variables

Set on the Railway `businessos-api` service:

| Variable | Value |
|----------|-------|
| `SECRET_KEY` | random 64-char hex string |
| `COOKIE_DOMAIN` | `.businessos.dev` |
| `GOOGLE_CLIENT_ID` | from the Google Cloud Console |
| `GOOGLE_CLIENT_SECRET` | from the Google Cloud Console |
| `GOOGLE_REDIRECT_URI` | `https://businessos.dev/api/v1/auth/oauth/google/callback` |

### Checklist

- [ ] Google OAuth client configured for the production redirect URI
- [ ] `SECRET_KEY` set (unique per environment)
- [ ] Login flow works: Google OAuth -> callback -> session -> dashboard
- [ ] Session cookies have `Secure` and `HttpOnly` flags and `Domain=.businessos.dev`
- [ ] Session expiration configured (default: 30 days)

---

## Phase 7: GitHub Secrets (CI/CD)

Set these in the repository (Settings -> Secrets and variables -> Actions):

### Required Secrets

| Secret | Value | Used By |
|--------|-------|---------|
| `RAILWAY_TOKEN` | Railway project token for project `BusinessOS`, environment `production` | `deploy-backend.yml` |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token (R2 + Pages) | desktop release uploads, Pages deploys |

> `RAILWAY_TOKEN` must be a **project token** (Railway dashboard > Project BusinessOS >
> Settings > Tokens > Create Token, select the production environment). It is used only for
> deploys and is revoked by deleting the token in the same panel.

### Desktop Build Secrets (optional)

| Secret | Value | Used By |
|--------|-------|---------|
| `APPLE_ID` | Apple Developer email | macOS code signing |
| `APPLE_PASSWORD` | App-specific password | macOS notarization |
| `APPLE_TEAM_ID` | Apple Team ID | macOS signing |
| `APPLE_IDENTITY` | Developer ID Application string | macOS signing |
| `WIN_CSC_LINK` | Windows code signing cert (base64) | Windows signing |
| `WIN_CSC_KEY_PASSWORD` | Cert password | Windows signing |

### Checklist

- [ ] `RAILWAY_TOKEN` set (project token)
- [ ] `CLOUDFLARE_API_TOKEN` set
- [ ] CI/CD deploy workflow updated (`.github/workflows/deploy-backend.yml`) -> Railway
- [ ] CI/CD deploy workflow updated (`.github/workflows/deploy-frontend.yml`) -> Cloudflare Pages
- [ ] Test deployment via CI: push to `main`, verify auto-deploy succeeds

---

## Phase 8: Desktop App (Electron Packaging)

> See `docs/deployment/ELECTRON-PACKAGING-GUIDE.md` and `docs/deployment/BUILD-DESKTOP.md`.

### Quick Steps

```bash
# 1. Build Go backend binaries for all platforms
cd desktop/backend-go
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" \
  -o ../desktop/resources/bin/darwin-arm64/businessos-server ./cmd/server

# 2. Install Electron dependencies
cd desktop && npm install

# 3. Build DMG (macOS)
npm run make -- --platform=darwin

# 4. Output: desktop/out/make/BusinessOS.dmg
```

### Checklist

- [ ] Go binaries built for target platforms
- [ ] `desktop/resources/icons/` has icon.icns, icon.ico, icon.png
- [x] `forge.config.ts` publisher defaults to `Miosa-osa/businessos-5`
- [ ] DMG builds without errors
- [ ] DMG installs and launches correctly
- [ ] Backend sidecar starts (check: `lsof -i :18080`)
- [ ] Frontend connects to sidecar backend
- [ ] SQLite database created at `~/Library/Application Support/BusinessOS/`
- [ ] Installers uploaded to the `businessos-downloads` R2 bucket (see `docs/deployment/BUILD-DESKTOP.md`)

---

## Phase 9: Monitoring & Observability

### Railway (Built-in)

```bash
# View build + runtime logs
railway logs --service businessos-api

# Metrics in the dashboard
# https://railway.com/dashboard (project BusinessOS > service businessos-api > Metrics)
```

### Error Tracking (Sentry)

See `docs/deployment/SENTRY_SETUP.md` for Sentry integration.

- [ ] Sentry DSN configured for backend
- [ ] Sentry DSN configured for frontend
- [ ] Error alerts configured (email or Slack)

### Uptime Monitoring

- [ ] Uptime check on `GET https://businessos-api-production.up.railway.app/health` every 5 minutes
- [ ] Uptime check on `GET https://businessos.dev/api/v1/health` (proxy path)
- [ ] Alert channels set up (email/Slack) if down for 2+ consecutive checks

### Checklist

- [ ] Railway logs accessible
- [ ] Sentry configured (optional but recommended)
- [ ] Uptime monitoring configured
- [ ] Alert channels set up (email/Slack)

---

## Phase 10: Smoke Test (E2E Verification)

### Web App

- [ ] Navigate to `https://businessos.dev`
- [ ] Sign in with Google OAuth
- [ ] Create a workspace
- [ ] Open chat interface
- [ ] Send: "Create a todo list app"
- [ ] Verify: SSE progress stream appears
- [ ] Verify: Generation completes (4 agent tasks)
- [ ] Verify: Files visible in generated apps list
- [ ] Verify: Can open app in Monaco editor
- [ ] Verify: Can view file contents

### Desktop App

- [ ] Install DMG / EXE / DEB
- [ ] Launch app
- [ ] Verify: Backend sidecar starts (system tray icon appears)
- [ ] Sign in with Google OAuth
- [ ] Create workspace
- [ ] Generate an app
- [ ] Verify: Same E2E flow as web
- [ ] Verify: Cmd+Shift+Space opens quick chat popup (macOS)
- [ ] Close and reopen app -> verify data persists (SQLite)

---

## Environment Rotation Plan

| Environment | Backend | Frontend | Database | Purpose |
|-------------|---------|----------|----------|---------|
| **Local** | `localhost:8801` | `localhost:5173` | Local PostgreSQL (`businessos_dev`) | Development |
| **Production** | Railway `businessos-api` | Cloudflare Pages `businessos-5` | Railway Postgres | Live users |

### Promotion Flow

```
LOCAL (developer machines)
    |
    | PR merge to main -> CI tests pass
    |
    v
MAIN (Railway + Cloudflare Pages auto-deploy on main)
    |
    | Smoke test + QA verification pass
    |
    v
PRODUCTION (businessos.dev)
```

There is currently a single cloud environment (`production`). Add a Railway
environment (and a Pages preview branch) if a separate staging tier is needed.

---

## Rollback Plan

### Backend Rollback

```bash
# List recent deployments
railway deployment list --service businessos-api

# Redeploy a known-good deployment from the dashboard (Deployments > Redeploy),
# or remove the newest one to fall back to the previous:
railway down -y --service businessos-api
```

### Frontend Rollback

Roll back in the Cloudflare Pages dashboard (project `businessos-5` >
Deployments > select a previous production deployment > Rollback), or redeploy
the previous build with Wrangler.

### Database Rollback

```bash
# Apply a rollback migration (if available)
psql "$DATABASE_URL" -f supabase/migrations/XXX_rollback.sql

# Or restore from a Railway Postgres backup
# Project BusinessOS > Postgres service > Backups (or point-in-time recovery).
```

---

## Post-Launch (Day 1-3)

- [ ] Monitor error rates (Sentry / Railway logs)
- [ ] Monitor response times (p50 < 200ms, p99 < 2s)
- [ ] Monitor database connection pool usage
- [ ] Check for any 5xx errors
- [ ] Verify auto-scaling works under load (Railway service scaling settings)
- [ ] Verify Redis session caching is working (if configured)
- [ ] Test auto-update for desktop app (publish a minor version bump)
- [ ] Document any issues found and create Sprint 2 tasks

---

**Related Docs:**
- `docs/DEPLOY.md` - Deploy runbook (Railway + Cloudflare Pages)
- `docs/deployment/GO-LIVE-SOP.md` - Per-release go-live SOP
- `docs/deployment/BUILD-DESKTOP.md` - Desktop packaging + downloads upload
- `docs/deployment/ELECTRON-PACKAGING-GUIDE.md` - Desktop app packaging
- `docs/deployment/DISASTER_RECOVERY.md` - DR procedures
- `docs/deployment/MONITORING_SETUP.md` - Monitoring configuration
