# Redis Outage Runbook

**Version:** 1.0.0
**Last Updated:** 2026-03-27
**Status:** ACTIVE

---

## Executive Summary

This runbook provides step-by-step instructions for handling Redis cache outages. Redis provides caching, pub/sub messaging, and session storage for BusinessOS.

**Risk Level:** MEDIUM - Redis outage degrades performance but core functionality remains available.

**Business Impact:** Slower response times (cache misses), lost pub/sub messages, temporary session issues. Data persists in PostgreSQL.

---

## Table of Contents

1. [Detection Symptoms](#detection-symptoms)
2. [Immediate Actions](#immediate-actions)
3. [Resolution Steps](#resolution-steps)
4. [Cache Invalidation Procedures](#cache-invalidation-procedures)
5. [Verification](#verification)
6. [Escalation Path](#escalation-path)

---

## Detection Symptoms

### Automatic Detection

**Symptoms:**
- Redis health check fails
- Cache error rate spikes >50%
- Response times increase 3-5x
- Pub/sub messages not delivered
- Session lookup failures

**Detection Time:** <10 seconds

### Manual Detection

**Symptoms:**
- Application slowdowns
- "Cache miss" logs flooding
- WebSocket connection failures
- User sessions lost

**Check Commands:**
```bash
# Redis ping
redis-cli -h localhost -p 6379 ping
# Expected: PONG

# Check container status
docker ps | grep businessos-redis

# Check Redis logs
docker logs --tail 50 businessos-redis

# Test connection
redis-cli -h localhost -p 6379 INFO server
```

---

## Immediate Actions

### 🚨 First 5 Minutes

- [ ] **Confirm Outage:** Verify Redis not responding
- [ ] **Enable Fallback Mode:** Configure app to bypass cache
- [ ] **Check Disk Space:** Redis may have stopped if disk full
- [ ] **Check Memory:** Redis requires sufficient RAM
- [ ] **Notify Team:** Alert backend team via Slack #backend-dev

### Cache Fallback Configuration

**Go Backend (BusinessOS):**
```go
// In internal/cache/redis.go, set fallback mode
cache.EnableFallbackMode(true) // Bypass cache, hit DB directly
```

**Environment Variable:**
```bash
export REDIS_FALLBACK=true
export CACHE_DISABLED=true
```

---

## Resolution Steps

### Option A: Container Restart (Soft Recovery)

**Use when:** Redis container crashed or stopped due to transient error.

```bash
# 1. Navigate to BusinessOS directory
cd /Users/sac/chatmangpt/BusinessOS

# 2. Restart Redis container
docker compose restart redis

# 3. Monitor startup
docker logs -f businessos-redis

# Expected output:
# Ready to accept connections tcp://0.0.0.0:6379
```

**Recovery Time:** 5-10 seconds

---

### Option B: Redis Data Recovery (With Persistence)

**Use when:** Redis stopped with data on disk, need to recover persisted cache.

```bash
# 1. Check if persistence files exist
docker exec businessos-redis ls -lh /data/dump.rdb
docker exec businessos-redis ls -lh /data/appendonly.aof

# 2. If dump.rdb exists, restore from snapshot
docker compose stop redis
docker run --rm \
  -v businessos_redis_data:/data \
  -v $(pwd)/tmp:/backup \
  alpine tar czf /backup/redis_backup.tar.gz /data

# 3. Start Redis with recovery mode
docker compose up -d redis

# 4. Verify data loaded
redis-cli -h localhost INFO persistence
# Check: rdb_last_cow_seconds, aof_last_cow_seconds
```

**Recovery Time:** 10-30 seconds (depending on data size)

---

### Option C: Full Redis Rebuild (Hard Recovery)

**Use when:** Container corrupted, disk issues, or data corrupted.

```bash
# 1. Stop Redis
docker compose stop redis

# 2. Backup existing data (if recoverable)
docker run --rm \
  -v businessos_redis_data:/data \
  -v $(pwd)/tmp:/backup \
  alpine tar czf /backup/redis_pre_rebuild_backup.tar.gz /data

# 3. Remove corrupted volume
docker volume rm businessos_redis_data

# 4. Rebuild and start
docker compose build redis
docker compose up -d redis

# 5. Verify clean start
redis-cli -h localhost INFO server
redis-cli -h localhost DBSIZE
# Expected: DBSIZE: 0 (empty cache)
```

**Recovery Time:** 30-60 seconds (cache will warm up over time)

---

## Cache Invalidation Procedures

### Scenario 1: Partial Data Corruption

**Use when:** Some keys are corrupted but Redis is mostly functional.

```bash
# 1. Identify corrupted keys
redis-cli --scan --pattern "user:*" | head -100

# 2. Backup good data
redis-cli --rdb /tmp/redis_backup.rdb

# 3. Delete corrupted keys
redis-cli --scan --pattern "corrupted:*" | xargs redis-cli DEL

# 4. Or flush specific database
redis-cli SELECT 1  # Switch to DB 1
redis-cli FLUSHDB  # Flush only DB 1
redis-cli SELECT 0  # Switch back to DB 0
```

---

### Scenario 2: Full Cache Flush Required

**Use when:** All cache data is suspect or needs refresh.

```bash
# WARNING: This clears ALL cached data
# Application will hit PostgreSQL until cache warms up

# Option 1: Flush current database
redis-cli FLUSHDB

# Option 2: Flush all databases
redis-cli FLUSHALL

# Option 3: Delete by pattern (safer)
redis-cli --scan --pattern "session:*" | xargs redis-cli DEL
redis-cli --scan --pattern "cache:*" | xargs redis-cli DEL
```

---

### Scenario 3: Cache Warming After Recovery

**Use when:** Redis was empty after recovery and needs priming.

```bash
# 1. Trigger cache warmup via API
curl -X POST http://localhost:8001/api/cache/warmup \
  -H "Content-Type: application/json" \
  -d '{"strategies": ["user_sessions", "frequent_queries", "config"]}'

# 2. Monitor cache hit rate
redis-cli INFO stats | grep hit_rate

# 3. Check cache size
redis-cli DBSIZE

# Expected: DBSIZE should increase as cache warms
```

---

## Verification

### Step 1: Redis Health

```bash
# Basic ping
redis-cli -h localhost -p 6379 ping
# Expected: PONG

# Detailed info
redis-cli -h localhost INFO server
# Expected: redis_version, uptime_in_days, connected_clients

# Memory check
redis-cli -h localhost INFO memory
# Expected: used_memory_human, maxmemory_human
```

### Step 2: Cache Operations

```bash
# Test SET/GET
redis-cli SET test_key "test_value"
redis-cli GET test_key
# Expected: "test_value"

# Test expiration
redis-cli SETEX temp_key 60 "expires_in_60s"
redis-cli TTL temp_key
# Expected: 60 (seconds remaining)

# Clean up test key
redis-cli DEL test_key temp_key
```

### Step 3: Application Integration

```bash
# Test cache write via BusinessOS API
curl -X POST http://localhost:8001/api/cache/set \
  -H "Content-Type: application/json" \
  -d '{"key":"test","value":"from_api","ttl":60}'

# Test cache read
curl http://localhost:8001/api/cache/get?key=test
# Expected: {"value":"from_api"}

# Clean up
curl -X DELETE http://localhost:8001/api/cache/delete?key=test
```

### Step 4: Pub/Sub Messaging

```bash
# Terminal 1: Subscribe to channel
redis-cli SUBSCRIBE updates

# Terminal 2: Publish message
redis-cli PUBLISH updates "test_message"

# Terminal 1: Should receive message
# Expected: "message" "updates" "test_message"
```

### Step 5: Performance Metrics

```bash
# Check cache hit rate (should be >80% after warmup)
redis-cli INFO stats | grep keyspace_hits
redis-cli INFO stats | grep keyspace_misses

# Calculate hit rate
# hit_rate = hits / (hits + misses)

# Check response time
redis-cli --latency localhost 6379
# Expected: <1ms avg latency
```

---

## Escalation Path

| Level | Contact | When to Escalate | Response Time |
|-------|---------|------------------|---------------|
| **L1** | Backend Team | Container restart, cache flush | 15 minutes |
| **L2** | Tech Lead | Data corruption, persistence issues | 30 minutes |
| **L3** | DevOps Engineer | Disk full, memory exhaustion | 1 hour |
| **L4** | CTO | Extended outage >30 min, data loss | Immediate |

### Support Channels

- **Slack:** #backend-dev, #incident-response
- **Email:** backend-team@businessos.com
- **On-Call:** Check PagerDuty for current on-call engineer

---

## Post-Incident Actions

### Immediate (Within 1 Hour)

- [ ] **Verify Cache Working:** All read/write operations succeed
- [ ] **Monitor Hit Rate:** Cache should warm up within 10-15 minutes
- [ ] **Check Application Logs:** No Redis errors
- [ ] **Disable Fallback Mode:** Switch back to cache-enabled
- [ ] **Notify Team:** Post resolution summary

### Short-Term (Within 24 Hours)

- [ ] **Analyze Root Cause:** Why did Redis fail?
- [ ] **Review Persistence:** Is RDB/AOF configured correctly?
- [ ] **Update Monitoring:** Add alerts for memory/disk usage
- [ ] **Document Learnings:** Update runbook with findings

### Long-Term (Within 1 Week)

- [ ] **Improve Persistence:** Enable AOF + RDB hybrid
- [ ] **Add Redis Sentinel:** For high availability
- [ ] **Implement Cache Stampede Protection:** Add locking
- [ ] **Chaos Test:** Add Redis outage scenario

---

## Common Issues and Solutions

### Issue 1: "Connection Refused"

**Symptoms:**
```
Error: Connection refused to Redis: dial tcp 127.0.0.1:6379
```

**Solution:**
```bash
# Check if Redis is running
docker ps | grep redis

# Start Redis if stopped
docker compose up -d redis

# Check if port is correct
redis-cli -h localhost -p 6379 ping
```

---

### Issue 2: "LOADING Redis is loading"

**Symptoms:**
```
LOADING Redis is loading the dataset in memory
```

**Cause:** Redis is recovering from RDB snapshot

**Solution:**
```bash
# Wait for loading to complete (monitor progress)
redis-cli INFO persistence
# Check: loading:0 (done) or loading:1 (in progress)

# If stuck, cancel and restart
docker compose restart redis
```

---

### Issue 3: Memory Exhaustion

**Symptoms:**
```
OOM command not allowed when used memory > 'maxmemory'
```

**Solution:**
```bash
# Check current memory usage
redis-cli INFO memory | grep used_memory_human

# Check maxmemory setting
redis-cli CONFIG GET maxmemory

# Set maxmemory and eviction policy
redis-cli CONFIG SET maxmemory 1gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru

# Make persistent in redis.conf
# maxmemory 1gb
# maxmemory-policy allkeys-lru
```

---

### Issue 4: Disk Full

**Symptoms:**
```
Failed to save RDB snapshot: No space left on device
```

**Solution:**
```bash
# Check disk space
df -h

# Clean up old Redis backups
docker exec businessos-redis ls -lh /data/*.rdb

# Remove old backups (keep latest 2)
docker exec businessos-redis sh -c "cd /data && ls -t *.rdb | tail -n +3 | xargs rm --"

# Or disable persistence temporarily
redis-cli CONFIG SET save ""
```

---

### Issue 5: Slow Cache Warmup

**Symptoms:**
- Hit rate <50% after 15 minutes
- Database under heavy load

**Solution:**
```bash
# Manually trigger cache warmup
curl -X POST http://localhost:8001/api/cache/warmup \
  -H "Content-Type: application/json" \
  -d '{"strategies": ["all"]}'

# Or preload common queries
redis-cli MSET key1 value1 key2 value2 key3 value3

# Monitor warmup progress
watch -n 5 'redis-cli INFO stats | grep keyspace'
```

---

## Quick Reference Commands

```bash
# Health check
redis-cli -h localhost ping

# View info
redis-cli INFO server
redis-cli INFO memory
redis-cli INFO persistence

# Monitor live commands
redis-cli MONITOR

# Check keys
redis-cli DBSIZE
redis-cli --scan --pattern "user:*" | wc -l

# Flush cache
redis-cli FLUSHDB      # Current DB
redis-cli FLUSHALL     # All DBs

# Backup to RDB
redis-cli --rdb /tmp/redis_backup.rdb

# Restore from RDB
docker cp /tmp/redis_backup.rdb businessos-redis:/data/dump.rdb
docker compose restart redis

# Check logs
docker logs -f businessos-redis

# Restart container
docker compose restart redis

# Full rebuild
docker compose down && docker compose build redis && docker compose up -d
```

---

## Prevention Measures

### Monitoring Alerts

Configure alerts for:
- Redis down (threshold: immediate)
- Memory usage >80% (threshold: warning)
- Disk usage >90% (threshold: critical)
- Cache hit rate <70% (threshold: warning)
- Connection errors >10/min (threshold: warning)

### Persistence Configuration

**In `docker-compose.yml`:**
```yaml
services:
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --save 60 1000
    volumes:
      - redis_data:/data
```

**Explanation:**
- `--appendonly yes`: AOF enabled (every write logged)
- `--save 60 1000`: RDB snapshot if 1000 writes in 60 seconds
- Hybrid approach: fast recovery from AOF, backup from RDB

### Auto-Restart Policy

```yaml
services:
  redis:
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 3
```

### Memory Limits

```yaml
services:
  redis:
    deploy:
      resources:
        limits:
          memory: 1G
        reservations:
          memory: 512M
```

---

**Document Version:** 1.0.0
**Last Reviewed:** 2026-03-27
**Next Review:** 2026-04-27
**Maintained By:** Backend Team
