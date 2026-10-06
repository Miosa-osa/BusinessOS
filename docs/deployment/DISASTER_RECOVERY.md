# Disaster Recovery Plan

## Overview

This document outlines the disaster recovery procedures for BusinessOS, including backup strategies, restore procedures, incident response, and recovery time objectives (RTO/RPO).

BusinessOS production runs on Railway (backend + Postgres) and Cloudflare (Pages frontend,
R2 downloads).

## Table of Contents

1. [Recovery Objectives](#recovery-objectives)
2. [Backup Strategy](#backup-strategy)
3. [Database Recovery](#database-recovery)
4. [Application Recovery](#application-recovery)
5. [Incident Response](#incident-response)
6. [Testing & Validation](#testing--validation)

---

## Recovery Objectives

### Definitions

**RTO (Recovery Time Objective):** Maximum acceptable time to restore service after an outage.

**RPO (Recovery Point Objective):** Maximum acceptable data loss measured in time.

### BusinessOS Objectives

| Service | RTO | RPO | Priority |
|---------|-----|-----|----------|
| Database (PostgreSQL) | 1 hour | 1 hour | Critical |
| Cache (Redis) | 30 minutes | None (ephemeral) | High |
| Backend API | 15 minutes | N/A | Critical |
| Frontend | 5 minutes | N/A | High |
| Background Jobs | 1 hour | N/A | Medium |

---

## Backup Strategy

### 1. Database Backups (Railway Postgres)

Railway's managed Postgres provides automated backups, and point-in-time recovery (PITR) on
eligible plans. Confirm the current backup settings in the Railway dashboard
(project `BusinessOS` > Postgres service > Backups).

```bash
# Confirm Postgres service status
railway status --service Postgres

# Take a manual logical backup and store it encrypted off-platform
railway connect Postgres --environment production
#   inside psql:  \! pg_dump "$DATABASE_URL" -Fc -f backup-$(date +%Y%m%d).dump
# or, from a shell with DATABASE_URL exported:
pg_dump "$DATABASE_URL" -Fc -f businessos-$(date +%Y%m%d).dump
```

Store manual dumps encrypted in Cloudflare R2 (see "Backup Storage Locations"):

```bash
gpg --symmetric --cipher-algo AES256 businessos-$(date +%Y%m%d).dump
npx wrangler r2 object put businessos-backups/db/businessos-$(date +%Y%m%d).dump.gpg \
  --file businessos-$(date +%Y%m%d).dump.gpg --remote
```

### 2. Redis Backups

Redis is optional and used for cache/rate limits; losing it is a non-event (RPO: none).
If a Railway Redis service is in use, enable persistence in its settings and verify with
`redis-cli INFO persistence`.

### 3. Application Configuration Backups

```bash
# Export the service variables (review before storing; contains secrets)
railway variables --service businessos-api --environment production

# Backup GitHub Actions secrets inventory (names only) from the repo settings
# Backup Railway project settings (service config, domains) from the dashboard
```

Store secret material encrypted in R2, never in the repo.

---

## Database Recovery

### Scenario 1: Restore from a Railway Backup

```bash
# 1. Open the Railway dashboard > project BusinessOS > Postgres > Backups.
# 2. Choose a backup / PITR point and restore it, OR clone the database.

# Alternatively, restore a manual dump into a fresh database:
pg_restore --clean --if-exists -d "$DATABASE_URL" businessos-YYYYMMDD.dump
```

### Scenario 2: Point-in-Time Recovery

Use Railway Postgres PITR (dashboard > Postgres > Backups > Point-in-time recovery) to roll
forward to a timestamp before the data-loss event.

```bash
# Confirm PITR availability
railway postgres --help
```

### Scenario 3: Restore from a Manual Export (R2)

```bash
# 1. Download the encrypted dump from R2
npx wrangler r2 object get businessos-backups/db/businessos-YYYYMMDD.dump.gpg \
  --file businessos-YYYYMMDD.dump.gpg --remote

# 2. Decrypt
gpg --output businessos-YYYYMMDD.dump --decrypt businessos-YYYYMMDD.dump.gpg

# 3. Restore into the target database
pg_restore --clean --if-exists -d "$DATABASE_URL" businessos-YYYYMMDD.dump
```

### Database Recovery Checklist

- [ ] Backup or PITR point identified
- [ ] Restore completed without errors
- [ ] Row counts match expectations (users, workspaces, module records)
- [ ] `pgvector` extension present
- [ ] Backend reconnects (`/health/detailed` reports `database: connected`)
- [ ] Read/write smoke test passes

---

## Application Recovery

### Scenario 1: Backend Service Down

```bash
# 1. Check service and recent deployments
railway status --service businessos-api
railway deployment list --service businessos-api
railway logs --service businessos-api

# 2. Roll back to the previous good deployment
railway down -y --service businessos-api
# or: Railway dashboard > Deployments > select a good deployment > Redeploy

# 3. If the service will not build at all, redeploy from a known-good commit:
STAGE="$(mktemp -d)/businessos-backend"
mkdir -p "$STAGE" && cp -R desktop/backend-go/. "$STAGE"/
railway up "$STAGE" --path-as-root --no-gitignore --service businessos-api --ci

# 4. Verify
curl -fsS https://businessos-api-production.up.railway.app/health
curl -fsS https://businessos-api-production.up.railway.app/health/detailed
```

### Scenario 2: Frontend Service Down

```bash
# 1. Check the Pages deployment
npx wrangler pages deployment list --project-name=businessos-5

# 2. Roll back in the Cloudflare Pages dashboard
#    Workers & Pages > businessos-5 > Deployments > Rollback to a good deployment

# 3. Or rebuild and redeploy
cd frontend && CLOUDFLARE_BUILD=true npm run build
npx wrangler pages deploy build --project-name=businessos-5 --branch=main --commit-dirty=true
```

### Scenario 3: Complete Regional Outage

Railway deploys into a single region per service; Cloudflare Pages is globally distributed.

```bash
# 1. Confirm the outage scope at https://status.railway.app and https://www.cloudflarestatus.com
# 2. If Railway is degraded globally, stand the backend up in another Railway
#    region/workspace from the same staged source and DATABASE_URL, then point the
#    Cloudflare Pages Function at it via the BUSINESSOS_BACKEND_URL env var
#    (Pages > Settings > Environment variables).
# 3. Frontend stays up (Cloudflare Pages), so users get the app shell with a clear error
#    from the API proxy while the backend is unavailable.
```

---

## Incident Response

### Incident Severity Levels

| Severity | Definition | Response Time | Example |
|----------|------------|---------------|---------|
| **P0 - Critical** | Complete service outage | Immediate | Database down, backend crashed |
| **P1 - High** | Major feature broken, data loss risk | < 15 min | Payment processing failing |
| **P2 - Medium** | Degraded performance, non-critical feature down | < 1 hour | Slow API responses, notifications delayed |
| **P3 - Low** | Minor issue, cosmetic bug | < 4 hours | UI glitch, typo |

### Incident Response Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                     INCIDENT DETECTED                           │
│                   (Alert, User Report)                          │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. ACKNOWLEDGE                                                 │
│     • Assign incident commander                                 │
│     • Create incident channel (#incident-YYYYMMDD-001)          │
│     • Update status page                                        │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. ASSESS                                                      │
│     • Determine severity (P0-P3)                                │
│     • Identify affected services                                │
│     • Estimate user impact                                      │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. MITIGATE                                                    │
│     • Execute runbook procedure                                 │
│     • Rollback if needed                                        │
│     • Engage escalation path                                    │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  4. MONITOR                                                     │
│     • Verify service restored                                   │
│     • Check error rates normalized                              │
│     • Confirm user impact resolved                              │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  5. COMMUNICATE                                                 │
│     • Update status page                                        │
│     • Notify affected users                                     │
│     • Document timeline                                         │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  6. POST-MORTEM                                                 │
│     • Root cause analysis                                       │
│     • Action items                                              │
│     • Update runbooks                                           │
└─────────────────────────────────────────────────────────────────┘
```

### Incident Commander Responsibilities

1. **Own the incident** from detection to resolution
2. **Coordinate responders** and delegate tasks
3. **Make decisions** on mitigation strategies
4. **Communicate** with stakeholders
5. **Document** actions and timeline
6. **Lead post-mortem** review

### On-Call Rotation

Recommended setup:
- **Primary on-call:** 24/7 coverage, 1-week shifts
- **Secondary on-call:** Backup escalation
- **Manager on-call:** Escalation for P0 incidents

**Tools:**
- PagerDuty / Opsgenie for alert routing
- Slack for incident coordination
- Statuspage.io for customer communication

See also `docs/operations/INCIDENT-RESPONSE-PLAYBOOK.md`.

---

## Testing & Validation

### Quarterly Disaster Recovery Drills

**Test 1: Database Restore (Q1, Q3)**

```bash
# 1. Create a scratch database (Railway dashboard, or a throwaway Postgres)
# 2. Restore the latest backup / dump into it
pg_restore --clean --if-exists -d "$SCRATCH_DATABASE_URL" businessos-YYYYMMDD.dump
# 3. Validate data
psql "$SCRATCH_DATABASE_URL" -c "SELECT COUNT(*) FROM users;"
# 4. Cleanup
#    Remove the scratch database in the Railway dashboard
```

**Test 2: Application Rollback (Q2, Q4)**

```bash
# 1. Note the current good deployment
railway deployment list --service businessos-api
# 2. Trigger a deploy (any small change) so there is a "newest" deployment
# 3. Execute rollback
railway down -y --service businessos-api
# 4. Measure rollback time (target: < 5 minutes)
# 5. Verify health
curl -fsS https://businessos-api-production.up.railway.app/health
```

**Test 3: Regional Failover (Annual)**

Full simulation of a regional outage:
- Stand the backend up in an alternate region from the same source + database
- Point the Pages Function at it (`BUSINESSOS_BACKEND_URL`)
- Verify application functionality
- Measure total recovery time
- Document lessons learned

### Validation Checklist

After any recovery:

- [ ] Database accessible and responding
- [ ] All critical tables present
- [ ] User authentication working
- [ ] API endpoints responding
- [ ] Frontend loading correctly
- [ ] Background jobs processing
- [ ] Monitoring dashboards show green
- [ ] No data corruption detected
- [ ] Sample user workflows tested
- [ ] Incident documented in runbook

---

## Contact Information

### Escalation Path

| Role | Contact | Escalation Level |
|------|---------|------------------|
| Primary On-Call | on-call@businessos.com | L1 |
| Engineering Manager | manager@businessos.com | L2 |
| CTO | cto@businessos.com | L3 |
| External Support | Railway, Cloudflare support | L3 |

### Vendor Support

**Railway:**
- Support: https://station.railway.com/
- Status: https://status.railway.app
- Docs: https://docs.railway.com

**Cloudflare:**
- Support: https://dash.cloudflare.com (Support > Help)
- Status: https://www.cloudflarestatus.com
- Docs: https://developers.cloudflare.com

---

## Appendix

### Backup Storage Locations

| Backup Type | Primary Storage | Secondary Storage | Retention |
|-------------|-----------------|-------------------|-----------|
| Database snapshots | Railway Postgres backups | R2 (`businessos-backups/db/`) | 7 days |
| Manual exports | R2 (`businessos-backups/db/`) | Local + offsite copy | 90 days |
| Application configs | GitHub repo (names only) | R2 (`businessos-backups/config/`, encrypted) | Indefinite |
| Downloads (installers) | R2 (`businessos-downloads`) | n/a (rebuildable from source) | Indefinite |

### Recovery Time Estimates

Based on production data size:

| Database Size | Backup Time | Restore Time |
|---------------|-------------|--------------|
| < 1 GB | 2 minutes | 5 minutes |
| 1-10 GB | 10 minutes | 20 minutes |
| 10-100 GB | 30 minutes | 60 minutes |
| 100+ GB | 60+ minutes | 2+ hours |

### Useful Commands Reference

```bash
# Railway (backend + Postgres)
railway status --service businessos-api
railway deployment list --service businessos-api
railway logs --service businessos-api
railway down -y --service businessos-api
railway variables --service businessos-api --environment production
railway connect Postgres --environment production

# Cloudflare (frontend + downloads)
npx wrangler pages deployment list --project-name=businessos-5
npx wrangler pages deploy build --project-name=businessos-5 --branch=main
npx wrangler r2 object list businessos-backups
npx wrangler r2 object list businessos-downloads
```

---

## Document Maintenance

This document should be reviewed and updated:
- **Quarterly:** After disaster recovery drills
- **After incidents:** Update based on lessons learned
- **Annually:** Full review of all procedures

**Last Reviewed:** 2026-10-06
**Next Review:** 2027-01-06
**Owner:** DevOps Team
