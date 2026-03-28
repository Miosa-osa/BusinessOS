# OSA Outage Runbook

**Version:** 1.0.0
**Last Updated:** 2026-03-27
**Status:** ACTIVE

---

## Executive Summary

This runbook provides step-by-step instructions for handling OSA (Operations System Architecture) service outages. OSA is the multi-agent orchestration layer running on port 8089.

**Risk Level:** MEDIUM - OSA outage disables agent coordination but BusinessOS frontend continues functioning.

**Business Impact:** Agent orchestration, healing reflexes, and board intelligence unavailable during outage. Existing conversations and tasks remain accessible.

---

## Table of Contents

1. [Detection Symptoms](#detection-symptoms)
2. [Immediate Actions](#immediate-actions)
3. [Resolution Steps](#resolution-steps)
4. [Verification](#verification)
5. [Escalation Path](#escalation-path)

---

## Detection Symptoms

### Automatic Detection (Circuit Breaker)

**Symptoms:**
- Health endpoint `/health` returns non-200 status
- Circuit breaker trips after 5 consecutive failures
- Prometheus alert `OSADown` fires
- Grafana dashboard shows OSA as "DOWN"

**Detection Time:** <5 seconds (circuit breaker polling interval)

### Manual Detection

**Symptoms:**
- Agent workflows hang or fail
- Board intelligence briefings unavailable
- Healing reflexes not executing
- WebSocket connections to OSA drop

**Check Commands:**
```bash
# Health check
curl -f http://localhost:8089/health || echo "OSA DOWN"

# Container status
docker ps | grep businessos-osa

# Container logs (last 50 lines)
docker logs --tail 50 businessos-osa
```

---

## Immediate Actions

### 🚨 First 5 Minutes

- [ ] **Confirm Outage:** Verify OSA container is not responding
- [ ] **Check Logs:** Review recent logs for crash reasons
- [ ] **Notify Team:** Alert backend team via Slack #backend-dev
- [ ] **Check Resources:** Verify CPU/memory not exhausted
- [ ] **Document Start Time:** Record when outage was detected

### Resource Check Commands

```bash
# Check container resource usage
docker stats businessos-osa --no-stream

# Check system resources
docker stats --all --no-stream

# Check disk space
df -h

# Check memory
free -h
```

---

## Resolution Steps

### Option A: Container Restart (Soft Recovery)

**Use when:** Container is running but not responding, or crashed due to transient error.

```bash
# 1. Navigate to BusinessOS directory
cd /Users/sac/chatmangpt/BusinessOS

# 2. Restart OSA container
docker compose restart osa

# 3. Monitor logs for startup
docker logs -f businessos-osa

# Expected output:
# [info] OSA starting...
# [info] Application: optimal_system_agent
# [info] Listening on http://0.0.0.0:8089
```

**Recovery Time:** 30-60 seconds

---

### Option B: Full Container Rebuild (Hard Recovery)

**Use when:** Container corrupted, dependency issues, or soft restart fails.

```bash
# 1. Stop all services
docker compose down

# 2. Rebuild OSA image
docker compose build osa

# 3. Start all services
docker compose up -d

# 4. Monitor OSA startup
docker logs -f businessos-osa

# Expected: Full startup sequence, all modules loaded
```

**Recovery Time:** 2-3 minutes

---

### Option C: Queue Drain Procedure

**Use when:** OSA crashed with in-flight messages in Redis queue.

```bash
# 1. Check queue depth (should be 0 normally)
redis-cli -h localhost -p 6379 LLEN osa:task_queue

# 2. If queue has items, drain to backup
redis-cli -h localhost -p 6379 LRANGE osa:task_queue 0 -1 > /tmp/osa_queue_backup.txt

# 3. Clear queue (after backup)
redis-cli -h localhost -p 6379 DEL osa:task_queue

# 4. Restart OSA (see Option A or B)

# 5. Replay queued tasks after recovery
while IFS= read -r task; do
  redis-cli -h localhost -p 6379 RPUSH osa:task_queue "$task"
done < /tmp/osa_queue_backup.txt
```

**Recovery Time:** 3-5 minutes (includes queue replay)

---

## Verification

### Step 1: Health Endpoint

```bash
curl -f http://localhost:8089/health
# Expected: {"status":"ok","timestamp":"..."}

# Detailed health
curl -f http://localhost:8089/health/detailed
# Expected: All modules reported as "started"
```

### Step 2: Agent Coordination

```bash
# Test agent endpoint
curl -X POST http://localhost:8089/api/agents/test-agent/execute \
  -H "Content-Type: application/json" \
  -d '{"task":"test"}'

# Expected: Agent response < 2 seconds
```

### Step 3: WebSocket Connection

```bash
# Test WebSocket (requires websocat or similar)
websocat ws://localhost:8089/socket/websocket

# Expected: Connection established, Phoenix heartbeat received
```

### Step 4: Circuit Breaker Reset

```bash
# Check circuit breaker state
curl http://localhost:8001/api/circuit-breaker/status

# Expected: State="closed", FailureCount=0

# Manually reset if stuck open
curl -X POST http://localhost:8001/api/circuit-breaker/reset
```

### Step 5: Integration Chain

```bash
# Test full chain: BusinessOS -> OSA -> Canopy
curl -X POST http://localhost:8001/api/integrations/a2a/agents/canopy/status \
  -H "Authorization: Bearer $TOKEN"

# Expected: Canopy agent status returned via OSA
```

---

## Escalation Path

| Level | Contact | When to Escalate | Response Time |
|-------|---------|------------------|---------------|
| **L1** | Backend Team | Container restart fails | 15 minutes |
| **L2** | Tech Lead | Queue corruption, data loss | 30 minutes |
| **L3** | DevOps Engineer | Docker daemon issues, disk full | 1 hour |
| **L4** | CTO | Extended outage >1 hour, production impact | Immediate |

### Support Channels

- **Slack:** #backend-dev, #incident-response
- **Email:** backend-team@businessos.com
- **On-Call:** Check PagerDuty for current on-call engineer

---

## Post-Incident Actions

### Immediate (Within 1 Hour)

- [ ] **Verify Full Recovery:** All health checks passing
- [ ] **Monitor Logs:** Watch for recurring errors
- [ ] **Update Team:** Post incident summary to Slack
- [ ] **Close Incident:** Mark incident as resolved in tracking system

### Short-Term (Within 24 Hours)

- [ ] **Root Cause Analysis:** Investigate why outage occurred
- [ ] **Update Monitoring:** Add alerts for detected symptoms
- [ ] **Review Logs:** Check for patterns or precursors
- [ ] **Document Learnings:** Update this runbook with findings

### Long-Term (Within 1 Week)

- [ ] **Fix Root Cause:** Address underlying issue
- [ ] **Improve Resilience:** Add auto-restart or redundancy
- [ ] **Team Training:** Review runbook with team
- [ ] **Chaos Test:** Add test scenario to chaos suite

---

## Common Issues and Solutions

### Issue 1: Container Won't Start

**Symptoms:**
```
Error: Cannot start service osa: ... address already in use
```

**Solution:**
```bash
# Find process using port 8089
lsof -i :8089

# Kill conflicting process
kill -9 <PID>

# Retry container start
docker compose up -d osa
```

---

### Issue 2: Database Connection Refused

**Symptoms:**
```
[error] Postgrex.Protocol connection refused
```

**Solution:**
```bash
# Verify PostgreSQL is running
docker ps | grep postgres

# Restart PostgreSQL if needed
docker compose restart postgres

# Verify database exists
docker exec -it businessos-postgres psql -U postgres -c "\l"

# Restart OSA after DB is ready
docker compose restart osa
```

---

### Issue 3: Redis Connection Timeout

**Symptoms:**
```
[error] Redix connection timeout
```

**Solution:**
```bash
# Verify Redis is running
docker ps | grep redis

# Test Redis connection
redis-cli -h localhost -p 6379 ping

# Restart Redis if needed
docker compose restart redis

# Flush any stale data
redis-cli -h localhost -p 6379 FLUSHDB

# Restart OSA
docker compose restart osa
```

---

### Issue 4: Memory Exhaustion

**Symptoms:**
```
[error] Cannot allocate memory
Container killed by OOMKiller
```

**Solution:**
```bash
# Check system memory
free -h

# Increase Docker memory limit (Docker Desktop settings)
# Or add memory limit to docker-compose.yml:
# services:
#   osa:
#     mem_limit: 2g

# Restart with new limits
docker compose down
docker compose up -d
```

---

### Issue 5: Circuit Breaker Stuck Open

**Symptoms:**
- OSA is healthy but circuit breaker remains open
- All requests fail with "circuit open" error

**Solution:**
```bash
# Manually reset circuit breaker
curl -X POST http://localhost:8001/api/circuit-breaker/reset

# Check configuration
curl http://localhost:8001/api/circuit-breaker/config

# If stuck, restart BusinessOS backend
docker compose restart backend
```

---

## Quick Reference Commands

```bash
# Health check
curl -f http://localhost:8089/health

# View logs (live)
docker logs -f businessos-osa

# View logs (last 100 lines)
docker logs --tail 100 businessos-osa

# Restart container
docker compose restart osa

# Full rebuild
docker compose down && docker compose build osa && docker compose up -d

# Check queue depth
redis-cli -h localhost LLEN osa:task_queue

# Drain queue
redis-cli -h localhost LRANGE osa:task_queue 0 -1 > backup.txt
redis-cli -h localhost DEL osa:task_queue

# Circuit breaker status
curl http://localhost:8001/api/circuit-breaker/status

# Reset circuit breaker
curl -X POST http://localhost:8001/api/circuit-breaker/reset
```

---

## Prevention Measures

### Monitoring Alerts

Configure alerts for:
- OSA health check failures (threshold: 3 consecutive)
- Container crash/restart (threshold: any)
- Queue depth >100 (threshold: warning)
- Circuit breaker open (threshold: immediate)
- Memory usage >80% (threshold: warning)

### Proactive Health Checks

```bash
# Add to crontab for every-minute checks
* * * * * curl -f http://localhost:8089/health || echo "OSA DOWN at $(date)" >> /var/log/osa_health.log
```

### Auto-Restart Policy

Ensure `docker-compose.yml` has:
```yaml
services:
  osa:
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8089/health"]
      interval: 10s
      timeout: 5s
      retries: 3
```

---

**Document Version:** 1.0.0
**Last Reviewed:** 2026-03-27
**Next Review:** 2026-04-27
**Maintained By:** Backend Team
