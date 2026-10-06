# BusinessOS Backend - Railway Deployment Guide

Complete setup guide for deploying the Go backend to Railway with the Railway Postgres
service, optional Redis, and GitHub Actions CI/CD.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Railway Project Setup](#railway-project-setup)
3. [Postgres](#postgres)
4. [Redis (Optional)](#redis-optional)
5. [Service Configuration](#service-configuration)
6. [Variables and Secrets](#variables-and-secrets)
7. [GitHub Secrets Checklist](#github-secrets-checklist)
8. [Manual Deployment](#manual-deployment)
9. [Automated CI/CD (GitHub Actions)](#automated-cicd-github-actions)
10. [Post-Deployment Verification](#post-deployment-verification)
11. [Rollback Procedures](#rollback-procedures)
12. [Monitoring and Alerts](#monitoring-and-alerts)
13. [Troubleshooting](#troubleshooting)

---

## Prerequisites

- Railway CLI installed and authenticated (`npm install --global @railway/cli`, then `railway login`)
- Access to the `Roberto Luna's Projects` workspace / `BusinessOS` project
- `gh` CLI (for GitHub Secrets setup, optional)

```bash
railway whoami
railway list
```

---

## Railway Project Setup

The project already exists:

| Thing | Value |
|-------|-------|
| Workspace | `Roberto Luna's Projects` |
| Project | `BusinessOS` |
| Environment | `production` |
| Service | `businessos-api` |
| Public URL | `https://businessos-api-production.up.railway.app` |

```bash
# Link the repo to the project (interactive), or pass the id
railway link
railway status --service businessos-api
```

---

## Postgres

Railway provides the database as a service in the same project.

```bash
# Add if missing: Railway dashboard > project BusinessOS > New > Database > PostgreSQL
railway status --service Postgres
```

Referencing `DATABASE_URL` from the `businessos-api` service is the normal wiring
(Railway exposes the Postgres connection string as a variable you reference on the service).

### Enable pgvector and run migrations

```bash
railway connect Postgres --environment production
# inside psql:
#   CREATE EXTENSION IF NOT EXISTS vector;
```

Apply migrations from `desktop/backend-go/internal/database/migrations/`. The backend can
auto-apply on boot, but apply explicitly before a deploy so a bad migration fails before the
new revision serves traffic.

---

## Redis (Optional)

Redis powers session caching and rate limiting. It is optional; without it
`/health/detailed` reports `redis: not_configured` and the backend still runs.

```bash
# Add: Railway dashboard > project BusinessOS > New > Database > Redis
# Then reference REDIS_URL / REDIS_PASSWORD on the businessos-api service.
```

---

## Service Configuration

| Setting | Value |
|---------|-------|
| Source | `desktop/backend-go` (uploaded by `railway up`) |
| Builder | Dockerfile (`desktop/backend-go/Dockerfile`) |
| Port | `8080` (the backend binds `SERVER_PORT`) |
| Domain | `businessos-api-production.up.railway.app` (Railway-provided) |

Public traffic reaches the backend through the Cloudflare Pages Function proxy at
`https://businessos.dev/api/*`, not through a direct API subdomain.

---

## Variables and Secrets

Set these in the Railway `businessos-api` service Variables tab.

### Non-secret

| Var | Value |
|-----|-------|
| `ENVIRONMENT` | `production` |
| `SERVER_PORT` | `8080` |
| `COOKIE_DOMAIN` | `.businessos.dev` |
| `GOOGLE_REDIRECT_URI` | `https://businessos.dev/api/v1/auth/oauth/google/callback` |
| `ALLOWED_ORIGINS` | `https://businessos.dev,https://app.businessos.dev,app://localhost,http://localhost:5173` |
| `AI_PROVIDER` | `anthropic` |
| `ENABLE_LOCAL_MODELS` | `false` |

### Secrets

`DATABASE_URL` (from the Postgres service), `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`,
`SECRET_KEY` (>= 32 chars), `ANTHROPIC_API_KEY`, `TOKEN_ENCRYPTION_KEY` (>= 32 chars),
`REDIS_KEY_HMAC_SECRET` (>= 32 chars), `INTERNAL_API_SECRET` (>= 32 chars),
`WEBHOOK_SIGNING_SECRET` (>= 16 chars).

Production `config.Validate()` crashes the service at boot if a required value is missing.

```bash
# Inspect
railway variables --service businessos-api --environment production

# Set (repeat per key)
railway variables --service businessos-api --environment production --set KEY=VALUE
```

---

## GitHub Secrets Checklist

Set in the repository (Settings > Secrets and variables > Actions):

| Secret | Value | Used By |
|--------|-------|---------|
| `RAILWAY_TOKEN` | Railway project token (project `BusinessOS`, environment `production`) | `deploy-backend.yml` |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token | desktop release uploads, Pages deploys |

Create `RAILWAY_TOKEN` in the Railway dashboard (project `BusinessOS` > Settings > Tokens >
Create Token, select the production environment). It is used only for deploys and is revoked
by deleting the token in the same panel.

---

## Manual Deployment

```bash
# Stage a copy outside the git tree: Railway prunes uploads with the repo-root
# .gitignore, which drops cmd/server from the backend.
STAGE="$(mktemp -d)/businessos-backend"
mkdir -p "$STAGE" && cp -R desktop/backend-go/. "$STAGE"/

railway up "$STAGE" --path-as-root --no-gitignore --service businessos-api --ci
```

To deploy from a checkout that is not a git repository at all, `railway up . --path-as-root`
from the staged directory works the same way (there is no `.gitignore` to prune).

---

## Automated CI/CD (GitHub Actions)

`.github/workflows/deploy-backend.yml` runs on push to `main` when backend files change:

1. Validate `RAILWAY_TOKEN` is present.
2. `go test -short -count=1 ./internal/...` (gating).
3. Install the Railway CLI.
4. Stage `desktop/backend-go` to `$RUNNER_TEMP` and `railway up`.
5. Smoke test `/health` and `/health/detailed` (asserts `database: connected`).

---

## Post-Deployment Verification

```bash
curl -fsS https://businessos-api-production.up.railway.app/health            # 200
curl -fsS https://businessos-api-production.up.railway.app/health/detailed   # database: connected
curl -fsS https://businessos.dev/api/v1/health                               # proxy path
```

---

## Rollback Procedures

```bash
railway deployment list --service businessos-api
railway down -y --service businessos-api      # remove the newest deployment
# or: Railway dashboard > Deployments > Redeploy a known-good deployment
```

---

## Monitoring and Alerts

```bash
railway logs --service businessos-api
railway metrics --service businessos-api
```

See `docs/deployment/MONITORING_SETUP.md` for Sentry, uptime, and alert configuration.

---

## Troubleshooting

| Symptom | Check |
|---------|-------|
| Boot crash / panic | A required variable is missing. Check `railway logs`; fix in the Variables tab and redeploy. |
| `database: unavailable` in `/health/detailed` | `DATABASE_URL` reference on the service, or the Postgres service is down. |
| Login does not stick | `COOKIE_DOMAIN=.businessos.dev` must be set. |
| OAuth redirect_uri mismatch | `GOOGLE_REDIRECT_URI` must equal the Google console authorized URI. |
| Build drops `cmd/server` | You deployed from inside the repo without staging; use the staged copy + `--no-gitignore`. |
| CORS errors | `ALLOWED_ORIGINS` must include `https://businessos.dev`. |
