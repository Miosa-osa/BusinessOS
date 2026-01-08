# BusinessOS Cloud Infrastructure

> Last Updated: December 21, 2025

This document describes the cloud infrastructure setup for BusinessOS, including GCP services, architecture decisions, and operational procedures.

---

## Overview

BusinessOS uses a hybrid local-first architecture with cloud synchronization:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           PRODUCTION ARCHITECTURE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐      ┌──────────────────┐      ┌──────────────────┐       │
│  │   Frontend   │─────▶│    Cloud Run     │─────▶│    Cloud SQL     │       │
│  │   (Vercel)   │      │   (Go Backend)   │      │   (PostgreSQL)   │       │
│  │   SvelteKit  │      │   Port 8080      │      │   PostgreSQL 15  │       │
│  └──────────────┘      └──────────────────┘      └──────────────────┘       │
│         │                      │                                             │
│         │               ┌──────┴──────┐                                      │
│         │               │             │                                      │
│         ▼               ▼             ▼                                      │
│  ┌──────────────┐ ┌──────────┐ ┌──────────────┐                             │
│  │   Electron   │ │   AI     │ │   Google     │                             │
│  │   Desktop    │ │ Providers│ │  Calendar    │                             │
│  │  + SQLite    │ │Anthropic │ │    OAuth     │                             │
│  └──────────────┘ └──────────┘ └──────────────┘                             │
│         │                                                                    │
│         ▼                                                                    │
│  ┌──────────────────────────────────────────┐                               │
│  │              SYNC ENGINE                  │                               │
│  │  Local SQLite ◀────────▶ Cloud SQL       │                               │
│  │  (Offline)      Sync     (Online)        │                               │
│  └──────────────────────────────────────────┘                               │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## GCP Services

### Project Details

| Property | Value |
|----------|-------|
| Project ID | `miosa-460433` |
| Region | `us-central1` |
| Billing | Enabled |

### Cloud SQL (PostgreSQL)

| Property | Value |
|----------|-------|
| Instance Name | `businessos-db` |
| Version | PostgreSQL 15 |
| Tier | `db-f1-micro` (dev) / `db-g1-small` (prod) |
| Region | `us-central1` |
| Connection Name | `miosa-460433:us-central1:businessos-db` |
| Database | `businessos` |

**Connection String Format:**
```
postgres://postgres:PASSWORD@/businessos?host=/cloudsql/miosa-460433:us-central1:businessos-db
```

### Cloud Run

| Property | Value |
|----------|-------|
| Service Name | `businessos-api` |
| URL | `https://businessos-api-460433387676.us-central1.run.app` |
| Region | `us-central1` |
| Memory | 512Mi |
| CPU | 1 |
| Min Instances | 0 |
| Max Instances | 10 |
| Timeout | 300s |

### Artifact Registry

| Property | Value |
|----------|-------|
| Repository | `businessos-repo` |
| Region | `us-central1` |
| Format | Docker |
| Image | `us-central1-docker.pkg.dev/miosa-460433/businessos-repo/businessos-api` |

---

## Authentication Architecture

### Cloud Run Authentication

The Cloud Run service uses **IAM-based authentication**. The organization policy blocks `allUsers` access, so:

1. **Browser requests** require an authenticated session (cookies from `/api/auth/*`)
2. **Server-to-server** requests require an identity token
3. **Local development** uses `localhost:8080` directly

### User Authentication Flow

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Browser   │─────▶│  Cloud Run  │─────▶│  Cloud SQL  │
│             │      │  /api/auth  │      │   users     │
└─────────────┘      └─────────────┘      └─────────────┘
       │                    │
       │  1. OAuth/Email    │
       │◀───────────────────│
       │                    │
       │  2. Session Cookie │
       │◀───────────────────│
       │                    │
       │  3. API Requests   │
       │   (with cookie)    │
       │───────────────────▶│
```

**Supported Auth Methods:**
- Google OAuth (`/api/auth/google`)
- Email/Password (`/api/auth/sign-in/email`, `/api/auth/sign-up/email`)

---

## Sync Architecture

### Sync Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/sync/status` | GET | Health check with server timestamp |
| `/api/sync/full` | GET | Full sync (all tables, all data) |
| `/api/sync/:table` | GET | Incremental sync for specific table |
| `/api/:table/sync` | GET | Alternative table-specific sync |

### Syncable Tables

```
contexts            conversations       messages
projects            artifacts           nodes
team_members        tasks              focus_items
daily_logs          user_settings      clients
client_contacts     client_interactions client_deals
calendar_events
```

### Sync Query Format

All sync queries accept a `since` parameter (RFC3339 timestamp):

```bash
# Get all contexts updated since a timestamp
GET /api/sync/contexts?since=2025-12-21T00:00:00Z

# Full sync (no since parameter = epoch)
GET /api/sync/full
```

---

## Frontend Configuration

### Auto-URL Detection

The frontend automatically selects the backend URL:

```typescript
// frontend/src/lib/api/base.ts
const LOCAL_BACKEND_URL = 'http://localhost:8080';
const CLOUD_RUN_URL = 'https://businessos-api-460433387676.us-central1.run.app';

function getApiBase(): string {
  const isDev = window.location.hostname === 'localhost';

  if (isElectron) {
    // Check localStorage for configured URL
    let cloudUrl = localStorage.getItem('businessos_cloud_url');
    if (!cloudUrl) {
      cloudUrl = isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
    }
    return `${cloudUrl}/api`;
  }

  // Web: auto-detect based on environment
  return isDev ? `${LOCAL_BACKEND_URL}/api` : `${CLOUD_RUN_URL}/api`;
}
```

### Mode Configuration

| Mode | Backend | Database | Use Case |
|------|---------|----------|----------|
| `cloud` | Cloud Run | Cloud SQL | Production, web app |
| `local` | localhost:8080 | Cloud SQL or local | Development |
| Electron `local` | localhost:18080 | Local SQLite | Offline desktop |

---

## Electron IPC Bridge

### Database Operations

The Electron app exposes database operations via IPC:

```typescript
// Available via window.electron.db
db.getAll(table, where?)     // Get all records
db.getById(table, id)        // Get by ID
db.create(table, data)       // Create record
db.update(table, id, data)   // Update record
db.delete(table, id)         // Delete record
db.query(sql, params?)       // Raw SQL query

// Domain-specific helpers
db.contexts.getWithChildren(parentId?)
db.conversations.getWithMessages(conversationId)
db.tasks.getByStatus(status?)
db.projects.getWithTasks(projectId)
db.calendar.getByRange(startDate, endDate)
db.clients.getWithDeals(clientId)
db.settings.get(userId)
db.settings.upsert(userId, settings)
```

### Sync Operations

```typescript
// Available via window.electron.sync
sync.getStatus()      // Get sync status
sync.trigger()        // Trigger manual sync
sync.getPending()     // Get pending changes count
```

---

## Deployment Commands

### Build and Deploy Backend

```bash
# Navigate to backend
cd /Users/rhl/Desktop/BusinessOS/desktop/backend-go

# Build Docker image
docker build -t us-central1-docker.pkg.dev/miosa-460433/businessos-repo/businessos-api:latest .

# Push to Artifact Registry
docker push us-central1-docker.pkg.dev/miosa-460433/businessos-repo/businessos-api:latest

# Deploy to Cloud Run
gcloud run deploy businessos-api \
  --image us-central1-docker.pkg.dev/miosa-460433/businessos-repo/businessos-api:latest \
  --region us-central1 \
  --add-cloudsql-instances miosa-460433:us-central1:businessos-db \
  --set-env-vars "DATABASE_URL=postgres://postgres:PASSWORD@/businessos?host=/cloudsql/miosa-460433:us-central1:businessos-db" \
  --set-env-vars "ENVIRONMENT=production" \
  --set-env-vars "ALLOWED_ORIGINS=https://businessos.app,http://localhost:5173,http://localhost:5174"
```

### Run Locally (Development)

```bash
# Start local backend connected to Cloud SQL
cd /Users/rhl/Desktop/BusinessOS/desktop/backend-go
go run ./cmd/server

# Or with Cloud SQL Proxy
cloud-sql-proxy miosa-460433:us-central1:businessos-db &
DATABASE_URL="postgres://postgres:PASSWORD@localhost:5432/businessos" go run ./cmd/server
```

---

## Environment Variables

### Cloud Run (Production)

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Cloud SQL connection string |
| `ENVIRONMENT` | Yes | `production` |
| `ALLOWED_ORIGINS` | Yes | CORS allowed origins |
| `GOOGLE_CLIENT_ID` | Yes | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Yes | Google OAuth secret |
| `SESSION_SECRET` | Yes | Session encryption key |
| `ANTHROPIC_API_KEY` | If using | AI provider key |

### Local Development

```bash
# .env file
DATABASE_URL=postgres://postgres:password@localhost:5432/businessos
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174,http://localhost:3000
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
SESSION_SECRET=dev-secret-key
```

---

## Monitoring & Debugging

### View Cloud Run Logs

```bash
# Recent logs
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=businessos-api" --limit 50

# Error logs only
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=businessos-api AND severity>=ERROR" --limit 20
```

### Check Service Health

```bash
# With identity token (required for IAM-protected services)
TOKEN=$(gcloud auth print-identity-token)
curl -H "Authorization: Bearer $TOKEN" https://businessos-api-460433387676.us-central1.run.app/health

# Expected response:
# {"status":"healthy"}
```

### View Metrics

- **Cloud Run**: [console.cloud.google.com/run](https://console.cloud.google.com/run/detail/us-central1/businessos-api)
- **Cloud SQL**: [console.cloud.google.com/sql](https://console.cloud.google.com/sql/instances/businessos-db)
- **Logs**: [console.cloud.google.com/logs](https://console.cloud.google.com/logs)

---

## Cost Estimates

| Service | Tier | Est. Monthly Cost |
|---------|------|-------------------|
| Cloud SQL | db-f1-micro | ~$10-15 |
| Cloud Run | Pay-per-use (min=0) | ~$0-20 |
| Artifact Registry | Storage | ~$1-5 |
| Networking | Egress | ~$1-10 |

**Total**: ~$12-50/month depending on usage

**Free Tier Benefits:**
- Cloud Run: 2M requests/month free
- Cloud SQL: No free tier, but micro instance is cheap
- Artifact Registry: 500MB free storage

---

## Troubleshooting

### "403 Forbidden" on Cloud Run

The service requires IAM authentication. For development:
1. Use local backend: `go run ./cmd/server`
2. Or get identity token: `gcloud auth print-identity-token`

### "CORS error"

1. Check `ALLOWED_ORIGINS` includes your frontend URL
2. Ensure no trailing slashes
3. Include both `http://` and `https://` if needed

### "Database connection failed"

1. Verify Cloud SQL instance is running
2. Check connection string format (use `/cloudsql/...` socket path)
3. Ensure Cloud Run has `roles/cloudsql.client` permission

### "Not authenticated" on API calls

1. User needs to log in first (session cookie required)
2. Check if cookies are being sent (`credentials: 'include'`)
3. Verify session is valid via `/api/auth/session`

---

## Security Considerations

1. **Cloud SQL**: Uses private IP + Cloud SQL Proxy socket
2. **Secrets**: Stored in environment variables (consider Secret Manager for production)
3. **CORS**: Restricted to known origins
4. **IAM**: Cloud Run requires authenticated invocations
5. **Sessions**: HTTP-only cookies, secure in production

---

## Files Modified/Created

| File | Purpose |
|------|---------|
| `desktop/backend-go/Dockerfile` | Multi-stage Docker build |
| `desktop/backend-go/internal/handlers/sync.go` | Sync endpoints |
| `desktop/backend-go/internal/middleware/cors.go` | CORS configuration |
| `desktop/src/main/ipc/database.ts` | IPC database handlers |
| `desktop/src/preload/index.ts` | Exposed db/sync APIs |
| `frontend/src/lib/auth-client.ts` | Auth with Cloud Run URL |
| `frontend/src/lib/api/base.ts` | API base with Cloud Run URL |
