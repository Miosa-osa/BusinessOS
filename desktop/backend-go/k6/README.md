# BusinessOS k6 Load Testing Scripts

Performance testing suite for BusinessOS endpoints using k6 load tester.

## Prerequisites

```bash
brew install k6
```

## Quick Start

```bash
# From BusinessOS/desktop/backend-go/
k6 run k6/load_test_osa.js          # Progressive load test (15 min)
k6 run k6/spike_test.js              # Spike test (3 min)
k6 run k6/endurance_test.js          # Endurance test (2 hours)
k6 run k6/load_test_hybrid.js        # Hybrid load test (17 min)
```

## Test Scripts

### 1. load_test_osa.js — Progressive Load Test
**Purpose:** Measure performance under increasing load
**Duration:** 15 minutes
**Stages:**
- 5 min @ 100 req/s
- 5 min @ 500 req/s
- 5 min @ 1000 req/s

**Thresholds (RED phase - will FAIL initially):**
- Error rate < 1%
- P95 latency < 500ms
- P99 latency < 1000ms

**Endpoints tested:**
- GET /healthz (BusinessOS liveness)
- GET /api/yawl/health (YAWL engine)
- GET /api/pm4py/health (pm4py-rust process mining)

---

### 2. spike_test.js — Spike Test
**Purpose:** Test system resilience to sudden traffic spikes
**Duration:** 3 minutes
**Stages:**
- 1 min @ 100 req/s (baseline)
- 30 sec @ 5000 req/s (50x spike)
- 1 min @ 100 req/s (recovery)

**Thresholds (RED phase - will FAIL initially):**
- Error rate < 5% during spike
- P95 latency < 1000ms
- P99 latency < 2000ms

**Endpoints tested:** Same as load_test_osa.js

---

### 3. endurance_test.js — Endurance Test
**Purpose:** Detect memory leaks and performance degradation over time
**Duration:** 2 hours
**Stages:**
- 2 hours @ 200 req/s (sustained load)

**Thresholds (RED phase - will FAIL initially):**
- Error rate < 1%
- P95 latency < 500ms
- P99 latency < 1000ms
- Memory leak detection: response time degradation < 50% over 2h

**Endpoints tested:** Same as load_test_osa.js

---

### 4. load_test_hybrid.js — Hybrid Load Test
**Purpose:** Test mixed traffic patterns across all BusinessOS subsystems
**Duration:** 17 minutes
**Stages:**
- 5 min @ 200 req/s
- 5 min @ 400 req/s
- 5 min @ 600 req/s

**Traffic split:**
- 40% /healthz (BusinessOS core)
- 30% /api/yawl/health (YAWL workflow engine)
- 30% /api/pm4py/health (pm4py-rust process mining)

**Thresholds (RED phase - will FAIL initially):**
- Error rate < 1% (per endpoint)
- P95 latency < 600ms
- P99 latency < 1000ms

---

## Chicago TDD Workflow

### RED Phase (Current)
All scripts have thresholds that WILL FAIL because endpoints aren't optimized yet.

```bash
k6 run k6/load_test_osa.js
# Expected: ✗ thresholds on metrics 'errors' have been crossed
#          Error rate > 1% (threshold requires < 1%)
```

### GREEN Phase (Next)
After implementing optimizations:

```bash
k6 run k6/load_test_osa.js
# Expected: ✓ All thresholds passed
#          Error rate < 1%
#          P95 latency < 500ms
```

### Verify OTEL Spans
After tests pass, verify spans in Jaeger:
1. Open http://localhost:16686
2. Search by service: `businessos`
3. Look for span names: `http.request`, `GET /healthz`, etc.
4. Verify attributes: `http.status_code`, `http.method`, `http.url`

---

## Environment Variables

```bash
export BASE_URL=http://localhost:8001  # Default
k6 run k6/load_test_osa.js

# Or override per test
BASE_URL=http://staging.example.com k6 run k6/load_test_osa.js
```

---

## Output Metrics

Each test outputs:

**Custom metrics:**
- `errors` — Error rate (target: < 1%)
- `osa_response_time` — Per-endpoint response time distribution

**HTTP metrics:**
- `http_req_duration` — Request duration (P50, P95, P99)
- `http_req_failed` — Failed request percentage
- `http_reqs` — Total requests per second

**Network metrics:**
- `data_received` — Bytes received
- `data_sent` — Bytes sent

---

## Continuous Integration

Add to CI/CD pipeline:

```yaml
# .github/workflows/performance.yml
- name: Run k6 load tests
  run: |
    k6 run --summary-export=results.json k6/load_test_osa.js
    # Assert thresholds passed
    if [ $? -ne 0 ]; then
      echo "Performance thresholds not met"
      exit 1
    fi
```

---

## Troubleshooting

### Test fails with "endpoint not accessible"
```bash
curl http://localhost:8001/healthz
# Should return: {"status":"ok"}
```

### High error rate (> 1%)
- Check BusinessOS is running: `make dev` (from repo root)
- Check port 8001 is not in use: `lsof -i :8001`
- Review logs: `make dev-logs`

### P95 latency exceeds threshold
- Profile Go backend: `go tool pprof`
- Check database query performance
- Verify Redis cache is working

---

## Performance Baseline (Target)

After GREEN phase, expected baseline:

| Metric | Target | Current (RED) |
|--------|--------|---------------|
| Error rate | < 1% | 70% (unoptimized) |
| P95 latency | < 500ms | 7.48ms (good) |
| P99 latency | < 1000ms | 16.77ms (good) |
| Throughput | 1000 req/s | 2.5 req/s (test limit) |

---

## References

- k6 Documentation: https://grafana.com/docs/k6/latest/
- BusinessOS CLAUDE.md: `/Users/sac/chatmangpt/BusinessOS/CLAUDE.md`
- PERFORMANCE_TESTING.md: `/Users/sac/chatmangpt/BusinessOS/docs/diataxis/how-to/PERFORMANCE_TESTING.md`
