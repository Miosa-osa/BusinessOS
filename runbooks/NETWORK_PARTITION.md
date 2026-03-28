# Network Partition Runbook

**Version:** 1.0.0
**Last Updated:** 2026-03-27
**Status:** ACTIVE

---

## Executive Summary

This runbook provides step-by-step instructions for handling network partitions between services. Network partitions occur when services cannot communicate due to network failures, causing timeouts and cascading failures.

**Risk Level:** HIGH - Network partition disrupts all inter-service communication (BusinessOS ↔ OSA ↔ Canopy ↔ pm4py-rust).

**Business Impact:** Agent orchestration fails, cross-system integration broken, board intelligence unavailable. Partial functionality may remain if services cache data locally.

---

## Table of Contents

1. [Detection Symptoms](#detection-symptoms)
2. [Immediate Actions](#immediate-actions)
3. [Timeout Adjustment Procedures](#timeout-adjustment-procedures)
4. [Resolution Steps](#resolution-steps)
5. [Verification](#verification)
6. [Escalation Path](#escalation-path)

---

## Detection Symptoms

### Automatic Detection

**Symptoms:**
- HTTP timeout errors spike
- Circuit breakers trip across services
- Health checks fail for dependent services
- OpenTelemetry traces show "timeout" spans
- Connection refused errors

**Detection Time:** <30 seconds

### Manual Detection

**Symptoms:**
- API requests hang or timeout
- Agent workflows fail at integration points
- "Service unavailable" errors in UI
- Intermittent connection failures

**Check Commands:**
```bash
# Test connectivity between services
curl -v http://localhost:8001/healthz      # BusinessOS
curl -v http://localhost:8089/health        # OSA
curl -v http://localhost:9089/health        # Canopy
curl -v http://localhost:8090/api/health    # pm4py-rust

# Check Docker network
docker network inspect businessos_default

# Check for container IP conflicts
docker inspect businessos-backend | grep IPAddress
docker inspect businessos-osa | grep IPAddress
```

---

## Immediate Actions

### 🚨 First 5 Minutes

- [ ] **Confirm Partition:** Identify which services cannot communicate
- [ ] **Check Docker Network:** Verify network bridge is healthy
- [ ] **Review Timeouts:** Check if timeouts are too aggressive
- [ ] **Enable Circuit Breakers:** Prevent cascading failures
- [ ] **Notify Team:** Alert backend team via Slack #backend-dev

### Network Diagnostics

```bash
# Check Docker network status
docker network ls

# Inspect network configuration
docker network inspect businessos_default

# Check container connectivity
docker exec businessos-backend ping -c 3 businessos-osa
docker exec businessos-osa ping -c 3 businessos-canopy

# Check port bindings
docker ps --format "table {{.Names}}\t{{.Ports}}"
```

---

## Timeout Adjustment Procedures

### Option A: Increase HTTP Timeouts (Go Backend)

**Use when:** Network is slow but reliable, timeouts are too aggressive.

**In BusinessOS (`internal/integrations/`):**

```go
// Before (too aggressive)
client := &http.Client{
    Timeout: 2 * time.Second,
}

// After (adjusted for network conditions)
client := &http.Client{
    Timeout: 30 * time.Second, // Increased from 2s
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second, // Connection timeout
            KeepAlive: 30 * time.Second,
        }).DialContext,
        MaxIdleConns:        100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    },
}
```

**Apply changes:**
```bash
# 1. Edit timeout configuration in code
# 2. Rebuild BusinessOS backend
docker compose build backend

# 3. Restart backend
docker compose restart backend
```

**Recovery Time:** 30-45 seconds

---

### Option B: Adjust OSA HTTP Timeouts

**Use when:** OSA cannot reach BusinessOS or Canopy due to slow network.

**In OSA (`config/dev.exs` or `config/prod.exs`):**

```elixir
# Before
config :osa, BusinessOSClient,
  timeout: 2000,
  recv_timeout: 2000

# After (network partition mode)
config :osa, BusinessOSClient,
  timeout: 30_000,      # 30 seconds
  recv_timeout: 30_000

config :osa, CanopyClient,
  timeout: 30_000,
  recv_timeout: 30_000
```

**Apply changes:**
```bash
# 1. Edit configuration file
# 2. Rebuild OSA
cd OSA && mix compile

# 3. Restart OSA container
docker compose restart osa
```

**Recovery Time:** 60-90 seconds (OSA rebuild)

---

### Option C: Adjust Circuit Breaker Thresholds

**Use when:** Circuit breakers tripping too aggressively during network degradation.

**In BusinessOS (`internal/circuitbreaker/`):**

```go
// Before (sensitive)
breaker := NewCircuitBreaker(
    WithMaxRequests(1),
    WithInterval(time.Second),
    WithTimeout(5*time.Second),
    WithReadyToTrip(func(counts Counts) bool {
        return counts.ConsecutiveFailures > 3
    }),
)

// After (tolerant)
breaker := NewCircuitBreaker(
    WithMaxRequests(5),
    WithInterval(10*time.Second),
    WithTimeout(60*time.Second),  // Give network time to recover
    WithReadyToTrip(func(counts Counts) bool {
        return counts.ConsecutiveFailures > 10  // More failures before trip
    }),
)
```

**Apply changes:**
```bash
# 1. Edit circuit breaker configuration
# 2. Rebuild backend
docker compose build backend && docker compose restart backend
```

**Recovery Time:** 30-45 seconds

---

### Option D: Enable Retry with Exponential Backoff

**Use when:** Network is unstable but requests succeed on retry.

**In BusinessOS (`internal/httpclient/`):**

```go
// Add retry middleware
func RetryMiddleware(maxRetries int) http Middleware {
    return func(next http.RoundTripper) http.RoundTripper {
        return &retryTransport{
            next:       next,
            maxRetries: maxRetries,
        }
    }
}

type retryTransport struct {
    next       http.RoundTripper
    maxRetries int
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    var lastErr error
    for attempt := 0; attempt <= t.maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff: 1s, 2s, 4s, 8s
            wait := time.Duration(1<<uint(attempt-1)) * time.Second
            time.Sleep(wait)
        }

        resp, err := t.next.RoundTrip(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        lastErr = err
    }
    return nil, lastErr
}

// Usage
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: RetryMiddleware(4)(http.DefaultTransport),
}
```

**Recovery Time:** N/A (proactive improvement)

---

## Resolution Steps

### Option A: Docker Network Repair

**Use when:** Docker network bridge is corrupted or misconfigured.

```bash
# 1. Stop all services
docker compose down

# 2. Remove old network
docker network rm businessos_default

# 3. Recreate network with explicit configuration
docker network create \
  --driver bridge \
  --subnet 172.20.0.0/16 \
  --gateway 172.20.0.1 \
  businessos_default

# 4. Restart services
docker compose up -d

# 5. Verify connectivity
docker exec businessos-backend ping -c 3 businessos-osa
```

**Recovery Time:** 30-60 seconds

---

### Option B: Host Network Mode (Emergency)

**Use when:** Docker network completely broken, emergency bypass needed.

**In `docker-compose.yml` (temporary):**
```yaml
services:
  backend:
    network_mode: "host"  # Bypass Docker network

  osa:
    network_mode: "host"
```

**Apply changes:**
```bash
# 1. Edit docker-compose.yml
# 2. Restart services
docker compose down && docker compose up -d

# 3. Revert to bridge network after fixing
```

**Warning:** Host mode bypasses Docker networking isolation. Use only as emergency measure.

**Recovery Time:** 20-30 seconds

---

### Option C: Service Reordering (Dependency Fix)

**Use when:** Services starting before dependencies are ready.

**In `docker-compose.yml`:**
```yaml
services:
  backend:
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: on-failure

  osa:
    depends_on:
      backend:
        condition: service_started
    restart: on-failure
```

**Apply changes:**
```bash
# 1. Edit docker-compose.yml
# 2. Restart with dependency order
docker compose up -d --force-recreate
```

**Recovery Time:** 30-45 seconds

---

### Option D: Firewall/Port Conflict Resolution

**Use when:** Ports blocked or conflicting.

```bash
# 1. Check for port conflicts
lsof -i :8001  # BusinessOS
lsof -i :8089  # OSA
lsof -i :9089  # Canopy
lsof -i :8090  # pm4py-rust

# 2. Kill conflicting processes
kill -9 <PID>

# 3. Check firewall rules
sudo iptables -L -n | grep 8001
sudo iptables -L -n | grep 8089

# 4. Add firewall rules if needed
sudo iptables -A INPUT -p tcp --dport 8001 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 8089 -j ACCEPT

# 5. Restart services
docker compose restart
```

**Recovery Time:** 15-30 seconds

---

## Verification

### Step 1: Service Health

```bash
# All services should be healthy
curl -f http://localhost:8001/healthz      # BusinessOS
curl -f http://localhost:8089/health        # OSA
curl -f http://localhost:9089/health        # Canopy
curl -f http://localhost:8090/api/health    # pm4py-rust
```

### Step 2: Inter-Service Connectivity

```bash
# Test BusinessOS -> OSA
docker exec businessos-backend curl -f http://businessos-osa:8089/health

# Test OSA -> Canopy
docker exec businessos-osa curl -f http://businessos-canopy:9089/health

# Test BusinessOS -> pm4py-rust
docker exec businessos-backend curl -f http://businessos-pm4py-rust:8090/api/health
```

### Step 3: End-to-End Integration

```bash
# Test full chain: Frontend -> BusinessOS -> OSA -> Canopy
curl -X POST http://localhost:8001/api/integrations/a2a/agents/canopy/status \
  -H "Content-Type: application/json"

# Expected: Agent status returned successfully
```

### Step 4: Timeout Configuration

```bash
# Verify timeouts are appropriate
# Check Go backend logs for timeout duration
docker logs businessos-backend | grep "timeout"

# Check OSA configuration
docker exec businessos-osa cat /app/config/dev.exs | grep timeout
```

### Step 5: Circuit Breaker State

```bash
# Check circuit breaker status (should be "closed")
curl http://localhost:8001/api/circuit-breaker/status

# Expected: {"state":"closed","failures":0,"successes":0}
```

---

## Escalation Path

| Level | Contact | When to Escalate | Response Time |
|-------|---------|------------------|---------------|
| **L1** | Backend Team | Timeout adjustments, network repair | 15 minutes |
| **L2** | Tech Lead | Docker network corruption, firewall issues | 30 minutes |
| **L3** | DevOps Engineer | Host networking, kernel problems | 1 hour |
| **L4** | CTO | Data center network issues, extended outage | Immediate |

### Support Channels

- **Slack:** #backend-dev, #incident-response, #networking
- **Email:** backend-team@businessos.com
- **On-Call:** Check PagerDuty for current on-call engineer

---

## Post-Incident Actions

### Immediate (Within 1 Hour)

- [ ] **Verify Connectivity:** All services communicating
- [ ] **Monitor Timeouts:** Check if adjustments are working
- [ ] **Close Circuit Breakers:** Reset if still open
- [ ] **Update Team:** Post resolution summary
- [ ] **Close Incident:** Mark as resolved

### Short-Term (Within 24 Hours)

- [ ] **Analyze Root Cause:** Why did partition occur?
- [ ] **Review Timeout Settings:** Are they appropriate?
- [ ] **Update Monitoring:** Add network latency alerts
- [ ] **Document Learnings:** Update runbook

### Long-Term (Within 1 Week)

- [ ] **Implement Service Mesh:** Consider Istio/Linkerd for resilience
- [ ] **Add Network Tests:** Chaos engineering for network partitions
- [ ] **Improve Retry Logic:** Exponential backoff everywhere
- [ ] **Circuit Breaker Tuning:** Adjust thresholds based on metrics

---

## Common Issues and Solutions

### Issue 1: "Connection Refused" Intermittently

**Symptoms:**
```
Error: dial tcp 172.20.0.5:8089: connect: connection refused
```

**Solution:**
```bash
# Check if target service is running
docker ps | grep businessos-osa

# Check service logs for crashes
docker logs --tail 50 businessos-osa

# Restart affected service
docker compose restart osa
```

---

### Issue 2: "i/o timeout" on All Requests

**Symptoms:**
- All requests timeout after N seconds
- Network appears frozen

**Solution:**
```bash
# Check Docker network status
docker network inspect businessos_default

# Restart Docker network
docker compose down
docker network rm businessos_default
docker compose up -d

# Or use host network mode (emergency)
# Edit docker-compose.yml: network_mode: "host"
```

---

### Issue 3: Circuit Breaker Won't Close

**Symptoms:**
- Circuit breaker stuck in "open" state
- Requests failing immediately

**Solution:**
```bash
# Manually reset circuit breaker
curl -X POST http://localhost:8001/api/circuit-breaker/reset

# Check configuration
curl http://localhost:8001/api/circuit-breaker/config

# If stuck, restart service
docker compose restart backend
```

---

### Issue 4: High Network Latency

**Symptoms:**
- Requests succeed but very slow (>10 seconds)
- Intermittent timeouts

**Solution:**
```bash
# Check system load
docker stats --all --no-stream

# Check for bandwidth saturation
iftop -i docker0

# Check for DNS issues
docker exec businessos-backend nslookup businessos-osa

# Increase timeouts (see Timeout Adjustment Procedures)
```

---

### Issue 5: Container IP Conflicts

**Symptoms:**
- Containers getting same IP address
- Connectivity failures

**Solution:**
```bash
# Check current IPs
docker inspect businessos-backend | grep IPAddress
docker inspect businessos-osa | grep IPAddress

# Restart services with IP allocation
docker compose down
docker compose up -d

# Or assign static IPs in docker-compose.yml
# services:
#   backend:
#     networks:
#       default:
#         ipv4_address: 172.20.0.10
```

---

## Quick Reference Commands

```bash
# Check service health
curl -f http://localhost:8001/healthz
curl -f http://localhost:8089/health

# Check Docker network
docker network inspect businessos_default

# Test connectivity between containers
docker exec businessos-backend ping -c 3 businessos-osa

# Restart Docker network
docker compose down && docker network rm businessos_default && docker compose up -d

# Check circuit breaker status
curl http://localhost:8001/api/circuit-breaker/status

# Reset circuit breaker
curl -X POST http://localhost:8001/api/circuit-breaker/reset

# Check port conflicts
lsof -i :8001

# Check firewall rules
sudo iptables -L -n

# View service logs
docker logs -f businessos-backend
docker logs -f businessos-osa

# Restart all services
docker compose restart
```

---

## Prevention Measures

### Monitoring Alerts

Configure alerts for:
- HTTP timeout rate >10% (threshold: warning)
- Circuit breaker open (threshold: immediate)
- Network latency >5s (threshold: warning)
- Connection refused errors (threshold: any)
- Docker network errors (threshold: any)

### Timeout Best Practices

**Service-Specific Timeouts:**
- Database queries: 5-10 seconds
- HTTP inter-service calls: 30 seconds
- External APIs: 60 seconds
- Circuit breaker half-open timeout: 60 seconds

**Retry Configuration:**
- Max retries: 3-4
- Backoff strategy: Exponential (1s, 2s, 4s, 8s)
- Jitter: Add random ±20% to avoid thundering herd

### Circuit Breaker Configuration

```go
// Recommended thresholds for production
breaker := NewCircuitBreaker(
    WithMaxRequests(5),
    WithInterval(30*time.Second),
    WithTimeout(60*time.Second),
    WithReadyToTrip(func(counts Counts) bool {
        // Trip if 50% failure rate over 30 seconds
        return counts.Requests > 10 && counts.ConsecutiveFailures > 5
    }),
)
```

### Service Discovery

Consider adding service discovery (Consul, etcd) for dynamic endpoint management and health-based routing.

---

**Document Version:** 1.0.0
**Last Reviewed:** 2026-03-27
**Next Review:** 2026-04-27
**Maintained By:** Backend Team
