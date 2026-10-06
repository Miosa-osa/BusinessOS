# Operations Runbook

> **Status:** ACTIVE — living document, add new scenarios as they are discovered in production
> **Owner:** Roberto / Pedro
> **Priority:** P1 (should have within 1 week of launch)

---

## Quick Reference

| Problem | Section |
|---------|---------|
| Backend won't start | [Backend Startup](#backend-startup-issues) |
| Database connection failed | [Database Issues](#database-issues) |
| Redis connection failed | [Redis Issues](#redis-issues) |
| Frontend blank page | [Frontend Issues](#frontend-issues) |
| SSE not streaming | [SSE Issues](#sse-issues) |
| App generation fails | [Generation Pipeline](#generation-pipeline-issues) |
| Auth not working | [Auth Issues](#auth-issues) |
| High memory / OOM | [Resource Issues](#resource-issues) |
| Deployment failed | [Deployment Issues](#deployment-issues) |

---

## Backend Startup Issues

### Backend crashes on startup

```bash
# Check logs
railway logs --service businessos-api

# Common causes:
# 1. Missing environment variable → check the Railway service Variables tab
# 2. Database connection refused → check the Railway Postgres service / DATABASE_URL
# 3. Port conflict → check if another service uses port 8080
```

### Backend healthy but returning 500s

```bash
# Check recent logs for panic or error patterns
railway logs --service businessos-api | grep -i "error\|panic"

# Check if it's a specific endpoint
curl -v $BACKEND_URL/health
curl -v $BACKEND_URL/health/detailed
```

---

## Database Issues

### Connection refused

```bash
# Verify the Railway Postgres service is running (dashboard: project BusinessOS)
railway status --service Postgres

# Open a psql shell against it (proxied by the CLI)
railway connect Postgres --environment production
# then: SELECT 1;
```

### Too many connections

```sql
-- Check current connections
SELECT count(*) FROM pg_stat_activity;

-- See who's connected
SELECT pid, usename, application_name, state, query_start
FROM pg_stat_activity ORDER BY query_start;

-- Kill idle connections older than 10 min
SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE state = 'idle' AND query_start < now() - interval '10 minutes';
```

### Slow queries

```sql
-- Find slow queries
SELECT pid, now() - query_start AS duration, query
FROM pg_stat_activity
WHERE state = 'active' AND now() - query_start > interval '5 seconds';

-- Analyze a specific query
EXPLAIN ANALYZE <your query here>;
```

### Migration failed

```bash
# Check which migrations have been applied
psql "$DATABASE_URL" -c "SELECT * FROM schema_migrations ORDER BY version;"

# Re-apply a specific migration
psql "$DATABASE_URL" -f supabase/migrations/XXX_migration_name.sql

# Rollback (if rollback file exists)
psql "$DATABASE_URL" -f supabase/migrations/XXX_rollback.sql
```

---

## Redis Issues

### Connection refused

```bash
# Check if Redis is running
redis-cli -h HOST -p PORT PING

# Check Cloud Run can reach Redis
# (Redis must be in same VPC or accessible via public IP)
```

### Redis full / OOM

```bash
redis-cli -h HOST INFO memory
redis-cli -h HOST DBSIZE

# Flush if necessary (WARNING: clears all sessions)
redis-cli -h HOST FLUSHDB
```

---

## Frontend Issues

### Blank page after deploy

```bash
# Check the Cloudflare Pages deployment + build logs
npx wrangler pages deployment list --project-name=businessos-5

# Common causes:
# 1. PUBLIC_API_URL not set → frontend can't reach backend
# 2. CORS blocking → backend doesn't allow frontend origin
# 3. Build error → check npm run build output
```

### API calls failing (CORS)

```bash
# Check CORS headers
curl -I -H "Origin: https://businessos.dev" $BACKEND_URL/health

# Expected:
# Access-Control-Allow-Origin: https://businessos.dev
# Access-Control-Allow-Credentials: true
```

---

## SSE Issues

### SSE stream not connecting

```bash
# Test SSE endpoint directly
curl -N -H "Accept: text/event-stream" "$BACKEND_URL/api/osa/apps/generate/1/stream"

# Check the Railway service request timeout (SSE needs > 60s)
# Railway does not cap HTTP request duration by default; if a proxy timeout is
# suspected, check the service's Networking settings in the Railway dashboard.
```

### Events being dropped

```
Check BuildEventBus buffer size (default: 100 events)
If subscriber is slow, events are silently dropped
TODO: Fix in Sprint 1 (Pedro Agent D)
```

---

## Generation Pipeline Issues

### FK violation on file save

```
This is the Sprint 1 P0 bug. Files don't persist because
osa_generated_apps row isn't created before osa_generated_files INSERT.
TODO: Fixed by Roberto Agent A
```

### Prompt is truncated

```
This is the Sprint 1 P0 bug. Rich prompt overwritten by short description.
TODO: Fixed by Roberto Agent A
```

### Generation stuck / no progress

```bash
# Check if queue worker is running
# Check backend logs for queue processing errors
railway logs --service businessos-api | grep "queue"

# Check if Anthropic API is reachable
curl -H "x-api-key: $ANTHROPIC_API_KEY" https://api.anthropic.com/v1/messages
```

---

## Auth Issues

### Google OAuth callback fails

```bash
# Verify redirect URI matches
# Google Cloud Console → APIs & Services → Credentials
# Authorized redirect URIs must include:
#   https://businessos.dev/api/v1/auth/oauth/google/callback

# Check GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are set
```

### Session expired / can't stay logged in

```bash
# Check BETTER_AUTH_SECRET is consistent across deploys
# If secret changed, all existing sessions are invalidated

# Check cookie settings (domain, secure, httpOnly)
```

---

## Resource Issues

### High memory (OOM kills)

```bash
# Check service metrics
# railway metrics --service businessos-api

# If consistently high: raise the memory limit in the Railway service Settings

# Check for memory leaks: goroutine count, connection pool
```

### High CPU

```bash
# Check for hot loops, unoptimized queries
# Profile if needed: add pprof endpoint (dev only)
```

---

## Deployment Issues

### Railway deploy fails

```bash
# Check the most recent deployments and their status
railway deployment list --service businessos-api

# Stream the failing build/deploy logs
railway logs --service businessos-api

# Redeploy the latest, or roll back by removing the newest deployment
railway deployment redeploy --service businessos-api
railway down -y --service businessos-api
```

### CI/CD pipeline stuck

```bash
# Check GitHub Actions
# https://github.com/Miosa-osa/businessos-5/actions

# Re-run failed job from GitHub UI
```

---

<!-- Add new scenarios as they are discovered in production -->

**Last Updated:** 2026-02-23
