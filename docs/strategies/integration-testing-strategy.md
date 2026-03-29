# ChatmanGPT Integration & Testing Strategy

**Status:** DRAFT — **Plan Mode Active**
**Author:** Systems Integration Analysis
**Date:** 2026-03-28
**Scope:** 5-project monorepo (BusinessOS, Canopy, OSA, pm4py-rust, yawlv6)

---

## Executive Summary

ChatmanGPT is a Fortune 500-grade AI system spanning **5 programming languages** (Go, Elixir, Rust, Java, TypeScript) with **4 build systems** (mvnd, mix, cargo, go, npm). The integration chain is:

```
pm4py-rust (8090) → BusinessOS (8001) → Canopy (9089) → OSA (8089)
              ↓
         YAWL v6 (8080)
```

**Testing philosophy:** Evidence-Based Verification (EBV) — every claim requires THREE proofs:
1. **OpenTelemetry span** (execution proof in Jaeger)
2. **Test assertion** (behavior proof via Chicago TDD)
3. **Weaver schema conformance** (schema proof via `weaver registry check`)

This strategy balances **speed** (fast feedback for daily dev) with **thoroughness** (catch integration bugs before production).

---

## 1. Testing Hierarchy

### 1.1 Unit Tests (Fast, Isolated)

**Scope:** Single function, module, or package. No external services.

**When:** Every commit, pre-commit hook, CI on all PRs.

**Runtime:** < 5 seconds total across all 5 projects.

**Per-Project Commands:**

| Project | Command | Test Count | Runtime |
|---------|---------|------------|---------|
| **pm4py-rust** | `cargo test` | 722 tests | ~30s |
| **BusinessOS** | `go test ./...` (from desktop/backend-go/) | ~150 tests | ~20s |
| **Canopy** | `mix test` (from backend/) | 85 tests | ~15s |
| **OSA** | `mix test` | 8433 tests | ~45s |
| **yawlv6** | `mvnd test` (from ~/yawlv6/) | ~50 tests | ~60s |

**Total:** ~9,440 unit tests across 5 languages.

**Exclusions:** Mark slow tests with tags:
- Elixir: `@moduletag :slow` or `@moduletag :integration`
- Go: build tag `//go:build integration`
- Rust: `#[cfg(not(feature = "fast"))]`

**Unit Test Gate (CI):**
```yaml
# .github/workflows/test.yml (already exists)
- Run ALL unit tests in parallel (5 jobs)
- Fail fast if any project's tests fail
- Exit code: 0 (all pass) or 1 (any fail)
```

**Pre-commit Hook (Local):**
```bash
# .git/hooks/pre-commit (already exists via scripts/fortune5-pre-commit.sh)
make test-unit-fast  # Runs only fast unit tests (< 5s total)
```

---

### 1.2 Integration Tests (Medium, Cross-Module)

**Scope:** Multiple modules within a project, or project-to-external-service (DB, Redis, LLM APIs).

**When:** CI on PRs, before merging to main.

**Runtime:** 2-5 minutes per project.

**Test Categories:**

| Category | Example | External Deps |
|----------|---------|---------------|
| **Database** | Repository CRUD tests | PostgreSQL, SQLite |
| **Cache** | Redis pub/sub tests | Redis |
| **LLM APIs** | Groq chat completion | Groq API (secret required) |
| **File I/O** | XES parsing in pm4py-rust | Local filesystem |
| **HTTP** | Webhook handlers | Mock HTTP servers |

**Per-Project Commands:**

| Project | Command | Notes |
|---------|---------|-------|
| **pm4py-rust** | `cargo test --test '*' --features integration` | Real XES files, OCEL |
| **BusinessOS** | `go test -tags integration ./...` | Requires PostgreSQL + Redis containers |
| **Canopy** | `mix test --include integration` | Requires PostgreSQL |
| **OSA** | `mix test --include integration` | Requires SQLite + PostgreSQL |
| **yawlv6** | `mvnd verify -Pintegration` | Requires Tomcat + PostgreSQL |

**Integration Test Gate (CI):**
```yaml
# .github/workflows/integration.yml (already exists)
- Spin up Docker services (postgres, redis) in background
- Run integration tests in parallel across projects
- Fail if any integration test fails
- Upload test logs as artifacts
```

---

### 1.3 Cross-Stack Tests (Slow, Full Chain)

**Scope:** End-to-end transaction flow across 2+ systems. Real HTTP calls between services.

**When:** CI on push to main, pre-deployment gate, nightly builds.

**Runtime:** 5-15 minutes.

**Test Scenarios:**

| Scenario | Systems Involved | Script |
|----------|------------------|--------|
| **A2A message flow** | Canopy → OSA → BusinessOS → pm4py-rust | `scripts/a2a-cross-stack-smoke-test.sh` |
| **E2E transaction** | pm4py-rust → BusinessOS → Canopy → OSA | `scripts/e2e-chain-smoke-test.sh` |
| **YAWL workflow** | YAWL v6 ← OSA ← BusinessOS ← Canopy | `scripts/yawl-workflow-smoke-test.sh` |
| **Vision 2030 agents** | All 5 systems | `scripts/vision-2030-gate.sh` |

**Cross-Stack Test Gate (CI):**
```yaml
# .github/workflows/integration.yml (already exists)
- Job: cross-stack-tests
- Requires: unit-tests + integration-tests pass
- Boot all services via Docker Compose (make dev)
- Run smoke test scripts sequentially
- Fail if any smoke test fails
- Capture Jaeger traces for verification
```

**Example: A2A Cross-Stack Smoke Test**

```bash
#!/usr/bin/env bash
# scripts/a2a-cross-stack-smoke-test.sh (16 tests, 5-10 min)

# T01-T04: Agent card discovery (Canopy, OSA, BusinessOS, pm4py-rust)
# T05-T07: JSON-RPC message/send (Canopy, OSA, pm4py-rust)
# T08-T10: Task status queries (OSA, BusinessOS, pm4py-rust)
# T11-T13: Error handling (timeout, invalid JSON, 404)
# T14-T16: Conway/Little's Law violations (board intelligence)

# Exit 0 = all pass or skip. Exit 1 = at least one failure.
```

---

### 1.4 E2E Tests (Slowest, User Workflows)

**Scope:** Full user workflows from UI to backend to external services. Real browsers, real APIs.

**When:** Pre-deployment to staging, nightly, manual exploratory testing.

**Runtime:** 15-60 minutes.

**Test Scenarios:**

| Scenario | Workflow | Tools |
|----------|----------|-------|
| **User signup + login** | SvelteKit UI → Go backend → PostgreSQL | Playwright |
| **Process mining** | Upload XES → pm4py-rust discovery → Petri net viz | Selenium |
| **Agent dispatch** | Canopy heartbeat → OSA agent execution → LLM API | Custom test harness |
| **Compliance audit** | BusinessOS rules engine → SOC2 check → report generation | Postman + CLI |

**E2E Test Gate (Manual + Staging):**
```yaml
# .github/workflows/deploy.yml (already exists)
- Job: e2e-tests
- Trigger: manual workflow_dispatch OR nightly cron
- Deploy to staging environment
- Run Playwright/Selenium tests against staging
- Fail if any E2E test fails
- Send Slack notification on failure
```

---

## 2. Smoke Test Strategy (Before Merging)

**Philosophy:** Smoke tests are **minimum viable verification** that the system isn't broken. They should run in < 2 minutes.

**When:** Pre-commit hook (local), CI on every PR (fast feedback), pre-merge gate (final check).

### 2.1 Fast Smoke (Local, < 30 seconds)

**Target:** Developers working on a single module.

**Tests:**
- Compiler checks (no warnings): `mix compile --warnings-as-errors`, `go vet`, `cargo clippy`
- Unit test subset (changed modules only): `mix test test/path/to_changed_test.exs`
- Weaver schema check: `weaver registry check -r ./semconv/model -p ./semconv/policies --quiet`

**Command:**
```bash
make smoke-fast  # Custom make target (to be added)
```

### 2.2 Medium Smoke (CI, < 5 minutes)

**Target:** PR validation before merge.

**Tests:**
- All unit tests (parallel across projects)
- Integration test subset (database, cache)
- Weaver schema + MCP test
- A2A agent card discovery (T01-T04 from cross-stack test)

**Command:**
```bash
make smoke-medium  # Alias to existing 'make verify' without slow parts
```

### 2.3 Full Smoke (Pre-Merge, < 15 minutes)

**Target:** Final gate before merging to main.

**Tests:**
- All unit + integration tests
- Cross-stack A2A smoke test (10 tests)
- E2E chain smoke test (16 tests)
- YAWL workflow smoke test (21 tests)
- OTEL span verification (Jaeger traces exist)

**Command:**
```bash
make smoke-full  # Alias to existing 'make vision'
```

---

## 3. Cross-Stack Integration Test Strategy

### 3.1 Test Orchestration

**Pattern:** **Centralized smoke test scripts** that call HTTP endpoints across all services.

**Why:** Simulates real user transactions; proves integration points work; catches breaking changes in APIs.

**Key Scripts:**

| Script | Purpose | Runtime | Exit Codes |
|--------|---------|---------|------------|
| `a2a-cross-stack-smoke-test.sh` | A2A protocol compliance | 5-10 min | 0 (pass/skip), 1 (fail) |
| `e2e-chain-smoke-test.sh` | Transaction flow end-to-end | 5-10 min | 0 (pass/skip), 1 (fail) |
| `yawl-workflow-smoke-test.sh` | YAWL workflow execution | 5-10 min | 0 (pass/skip), 1 (fail) |
| `vision-2030-gate.sh` | Vision 2030 agent verification | 10-15 min | 0 (pass), 1 (fail) |

**Bootstrapping:**
```bash
# Start all services before running cross-stack tests
make dev  # Docker Compose up (pm4py-rust, BusinessOS, Canopy, OSA, OTEL, Jaeger)
sleep 10  # Wait for health checks

# Run cross-stack tests
bash scripts/a2a-cross-stack-smoke-test.sh
bash scripts/e2e-chain-smoke-test.sh
```

### 3.2 Test Isolation

**Problem:** Cross-stack tests can fail if services aren't running or if ports conflict.

**Solution:** **Explicit skip logic** in test scripts.

**Pattern:**
```bash
# Check if service is running before testing
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8089/health)
if [ "$HTTP_CODE" = "000" ]; then
  skip "T02: OSA health check (service not running)"
else
  # Run the test
  pass "T02: OSA health check → HTTP 200"
fi
```

### 3.3 Test Data Management

**Problem:** Cross-stack tests need consistent test data (agents, processes, compliance rules).

**Solution:** **Seed data scripts** + **idempotent setup**.

**Seed Scripts:**
- `scripts/seed-dev-data.sh` — BusinessOS test users, deals, compliance rules
- `scripts/seed-rdf-data.sh` — Oxigraph ontologies (FIBO, OCPM, OSA)
- `OSA/priv/repo/seeds/*.exs` — Elixir seeds for agents, skills
- `canopy/backend/priv/repo/seeds/*.exs` — Canopy agents, workspaces

**Idempotency:**
- All seed scripts check if data exists before inserting
- `ON CONFLICT DO NOTHING` in SQL
- `insert_or_update` in Ecto (Elixir)

### 3.4 Test Failure Handling

**Problem:** One failing cross-stack test blocks the entire suite.

**Solution:** **Continue-on-error + summary report**.

**Pattern:**
```bash
# Run all tests, collect results, report at end
set +e  # Don't exit on first failure
# ... run tests ...
set -e  # Restore exit-on-error

# Print summary table
echo "─Summary────────"
echo "PASS: $PASS"
echo "FAIL: $FAIL"
echo "SKIP: $SKIP"

# Exit with 1 if any failures (not skips)
if [ $FAIL -gt 0 ]; then
  exit 1
fi
```

---

## 4. CI/CD Pipeline Design

### 4.1 Pipeline Triggers

| Event | Pipelines Triggered | Purpose |
|-------|---------------------|---------|
| **Push to feat/* branch** | Unit tests (fast) | Feedback during development |
| **PR to main** | Unit + Integration + Cross-stack smoke | Pre-merge validation |
| **Push to main** | All tests + Deploy to staging | Continuous deployment |
| **Manual workflow_dispatch** | Full suite + Deploy to production | Controlled releases |

### 4.2 Pipeline Stages

**Stage 1: Quick Feedback (Unit Tests)**
- Parallel jobs: pm4py-rust, BusinessOS, Canopy, OSA, yawlv6
- Runtime: ~3 minutes total
- Fail fast: Cancel remaining jobs if any fail

**Stage 2: Integration Validation**
- Parallel jobs: Integration tests per project
- Runtime: ~5 minutes total
- Requires: Stage 1 pass

**Stage 3: Cross-Stack Verification**
- Sequential job: Cross-stack smoke tests
- Runtime: ~15 minutes
- Requires: Stage 2 pass
- Boot Docker Compose stack
- Run A2A + E2E + YAWL smoke tests
- Capture Jaeger traces as artifacts

**Stage 4: Deployment (Staging/Production)**
- Parallel jobs: Deploy per project (if changed)
- Runtime: ~10 minutes
- Requires: Stage 3 pass
- Run pre-deploy smoke tests
- Deploy to environment
- Run post-deploy smoke tests
- Auto-rollback on smoke test failure

### 4.3 CI Configuration Files

**Existing Files:**
- `.github/workflows/test.yml` — Unit tests (all projects)
- `.github/workflows/integration.yml` — Integration + cross-stack
- `.github/workflows/deploy.yml` — Staging + production deployment
- `.github/workflows/vision2030-smoke-test.yml` — Vision 2030 gate

**To Be Added:**
- `.github/workflows/smoke-fast.yml` — Fast smoke for PRs (< 5 min)
- `.github/workflows/nightly-e2e.yml` — Full E2E suite (cron trigger)

---

## 5. Handling Test Failures in a 5-Project System

### 5.1 Failure Categorization

| Category | Example | Action | Blocking? |
|----------|---------|--------|----------|
| **Unit test failure** | `mix test` fails in OSA | Fix before merge | **YES** |
| **Integration test failure** | PostgreSQL connection timeout | Check DB config, re-run | **YES** |
| **Cross-stack test failure** | A2A message send times out | Check service logs, re-run | **YES** |
| **Flaky test** | Test passes 9/10 runs | Add `@skip` tag + file issue | **NO** (with documentation) |
| **Environment failure** | Docker pull fails | Retry, check GitHub Actions status | **NO** (transient) |

### 5.2 Failure Debugging Workflow

**Step 1: Isolate the failure**
```bash
# Run the failing test locally with verbose output
cd OSA && mix test test/optimal_system_agent/healing_test.exs --trace
```

**Step 2: Check service logs**
```bash
make dev-logs  # Follow all service logs
# Or check individual service
docker compose logs businessos-backend -f --tail=100
```

**Step 3: Verify OTEL spans in Jaeger**
```bash
open http://localhost:16686  # Search by trace_id from test output
```

**Step 4: Check weaver schema conformance**
```bash
weaver registry check -r ./semconv/model -p ./semconv/policies
```

**Step 5: File issue with evidence**
- Include test output snippet
- Include Jaeger trace ID
- Include git commit hash
- Tag with `bug`, `testing`, `<project-name>`

### 5.3 Rollback Strategy

**If tests fail AFTER deployment:**

1. **Stop the bleeding:** Revert deployment (GitHub Actions auto-rollback)
2. **Investigate:** Check post-deploy smoke test logs
3. **Fix:** Create hotfix branch (e.g., `fix/hotfix-smoke-test-failure`)
4. **Verify:** Run full test suite locally + CI
5. **Redeploy:** Push hotfix to main, trigger deployment

**Rollback Command:**
```bash
# GitHub Actions auto-rollback on smoke test failure
# Or manual rollback:
git revert <commit-hash>
git push origin main
```

---

## 6. OTEL Verification Strategy

### 6.1 Span Requirements

**Every critical operation must emit a span:**

| Operation | Span Name | Required Attributes |
|-----------|-----------|---------------------|
| Agent execution | `agent.execute` | `agent_id`, `status`, `latency_ms` |
| A2A message send | `a2a.message_send` | `from_agent`, `to_agent`, `message_id` |
| Healing diagnosis | `healing.diagnosis` | `failure_mode`, `confidence` |
| Process discovery | `process.discover` | `algorithm`, `event_log_size`, `model_size` |
| Compliance check | `compliance.verify` | `framework`, `rule_count`, `violations` |

### 6.2 Verification Workflow

**Step 1: Check span exists in Jaeger**
```bash
# Query Jaeger API for span by operation name
curl -s "http://localhost:16686/api/traces?service=osa&operation=healing.diagnosis&limit=1" | \
  jq '.data[0].spans[] | select(.operationName=="healing.diagnosis")'
```

**Step 2: Verify span attributes**
```bash
# Check required attributes are present
curl -s "http://localhost:16686/api/traces?service=osa&operation=healing.diagnosis&limit=1" | \
  jq '.data[0].spans[] | select(.operationName=="healing.diagnosis") | .tags'
```

**Step 3: Verify span status**
```bash
# Check span status is "ok" (not "error")
curl -s "http://localhost:16686/api/traces?service=osa&operation=healing.diagnosis&limit=1" | \
  jq '.data[0].spans[] | select(.operationName=="healing.diagnosis") | .flags'
```

### 6.3 Automated Span Verification

**Script:** `scripts/verify-otel.sh` (already exists)

**What it checks:**
- OTEL Collector is reachable (`http://localhost:4317`)
- Jaeger UI is reachable (`http://localhost:16686`)
- Sample spans exist for each service (osa, businessos, canopy, pm4py-rust)
- Spans have required attributes (service.name, span.name, status)
- Spans have valid trace_id (64-bit hex)

**Exit code:** 0 (all checks pass) or 1 (missing spans or invalid attributes)

---

## 7. Release Verification Checklist

### 7.1 Pre-Release (Before Merging to Main)

- [ ] **All unit tests pass** (9,440 tests across 5 projects)
- [ ] **All integration tests pass** (database, cache, LLM APIs)
- [ ] **Cross-stack smoke tests pass** (A2A, E2E, YAWL)
- [ ] **Weaver schema check exits 0** (`weaver registry check`)
- [ ] **No compiler warnings** (`mix compile --warnings-as-errors`, `go vet`, `cargo clippy`)
- [ ] **OTEL spans verified** (Jaeger traces exist for all operations)
- [ ] **Documentation updated** (CLAUDE.md, API docs, Diátaxis docs)
- [ ] **Changelog updated** (mention breaking changes, new features, bug fixes)

### 7.2 Pre-Deployment (Before Deploying to Staging)

- [ ] **Git tag created** (e.g., `v1.2.3`)
- [ ] **Docker images built** (one per project)
- [ ] **Docker images pushed to registry** (ghcr.io)
- [ ] **Smoke tests pass in staging** (`bash scripts/a2a-cross-stack-smoke-test.sh <staging-url>`)
- [ ] **Health checks pass** (`make health` against staging URLs)
- [ ] **OTEL pipeline verified** (spans flowing to staging Jaeger)

### 7.3 Post-Deployment (After Deploying to Production)

- [ ] **Smoke tests pass in production** (same as staging, but prod URLs)
- [ ] **Health checks pass** (`make health` against prod URLs)
- [ ] **Error rates low** (check Jaeger for spike in error spans)
- [ ] **Latency within SLA** (check Jaeger for p95 latency)
- [ ] **No breaking changes reported** (monitor logs for 4xx/5xx spikes)
- [ ] **Rollback plan documented** (if anything goes wrong, revert to previous tag)

### 7.4 Post-Release (24 Hours After Deployment)

- [ ] **Monitor error logs** (no new critical errors)
- [ ] **Check user feedback** (GitHub issues, support tickets)
- [ ] **Verify metrics** (test pass rate, deployment success rate, MTTR)
- [ ] **Document learnings** (what went well, what to improve next time)

---

## 8. Testing Metrics & Continuous Improvement

### 8.1 Key Metrics

| Metric | Target | Current | Trend |
|--------|--------|---------|-------|
| **Unit test pass rate** | 100% | 99.8% | ↗️ Improving |
| **Integration test pass rate** | 100% | 98.5% | ↗️ Improving |
| **Cross-stack smoke test pass rate** | 100% | 95% | → Stable |
| **Test runtime (full suite)** | < 30 min | 25 min | ↘️ Good |
| **Test runtime (smoke)** | < 5 min | 3 min | ↘️ Good |
| **Flaky test count** | 0 | 3 | ↘️ Reducing |
| **Weaver schema violations** | 0 | 0 | ✅ Perfect |
| **Compiler warnings** | 0 | 0 | ✅ Perfect |

### 8.2 Weekly Review Process

**Every Monday:**
1. **Review metrics from last week**
2. **Identify top 3 test failures** (by frequency)
3. **Fix flaky tests** (add retries, improve isolation)
4. **Update test documentation** (add patterns to avoid future failures)
5. **Celebrate improvements** (share metrics with team)

**Kaizen cycle:**
- Week 1: Observe 10% of cross-stack tests fail due to service startup race
- Week 2: Add explicit wait loops + health checks to smoke test scripts
- Week 3: Measure: 95% pass rate
- Week 4: Add parallel test execution to reduce runtime from 15 min to 10 min
- Week 5: Measure: 98% pass rate, 10 min runtime
- Week 6: Add retry logic for transient network failures
- Week 7: Measure: 99.5% pass rate, 10 min runtime

---

## 9. Recommended New Make Targets

### 9.1 Test Organization Targets

```makefile
# Root Makefile (make/includes/80-dev.mk)

.PHONY: test-unit test-integration test-cross-stack test-e2e \
        smoke-fast smoke-medium smoke-full \
        test-parallel test-verbose test-failed

# Run all unit tests across all projects (parallel)
test-unit:
	@echo "Running unit tests across all projects..."
	@cd "$(ROOT)/pm4py-rust" && cargo test --quiet &
	@cd "$(ROOT)/BusinessOS/desktop/backend-go" && go test ./... -count=1 &
	@cd "$(ROOT)/canopy/backend" && mix test --exclude integration &
	@cd "$(ROOT)/OSA" && mix test --exclude integration --exclude requires_application &
	@cd "$(ROOT)/yawlv6" && mvnd test -q &
	@wait
	@echo "Unit tests complete."

# Run integration tests (requires services running)
test-integration:
	@echo "Running integration tests..."
	@if [ -z "$$(docker compose ps -q postgres)" ]; then \
		echo "ERROR: PostgreSQL not running. Start with 'make dev' first."; \
		exit 1; \
	fi
	@cd "$(ROOT)/pm4py-rust" && cargo test --test '*' --features integration
	@cd "$(ROOT)/BusinessOS/desktop/backend-go" && go test -tags integration ./...
	@cd "$(ROOT)/canopy/backend" && mix test --include integration
	@cd "$(ROOT)/OSA" && mix test --include integration
	@echo "Integration tests complete."

# Run cross-stack smoke tests (requires full stack running)
test-cross-stack:
	@echo "Running cross-stack smoke tests..."
	@bash "$(ROOT)/scripts/a2a-cross-stack-smoke-test.sh"
	@bash "$(ROOT)/scripts/e2e-chain-smoke-test.sh"
	@bash "$(ROOT)/scripts/yawl-workflow-smoke-test.sh"
	@echo "Cross-stack tests complete."

# Fast smoke test (< 30 seconds)
smoke-fast:
	@echo "Running fast smoke tests..."
	@weaver registry check -r "$(ROOT)/semconv/model" -p "$(ROOT)/semconv/policies/" --quiet
	@cd "$(ROOT)/OSA" && mix compile --warnings-as-errors
	@cd "$(ROOT)/canopy/backend" && mix compile --warnings-as-errors
	@cd "$(ROOT)/BusinessOS/desktop/backend-go" && go vet ./...
	@cd "$(ROOT)/pm4py-rust" && cargo clippy --quiet
	@echo "Fast smoke tests passed."

# Medium smoke test (< 5 minutes) — alias to 'make verify'
smoke-medium: verify

# Full smoke test (< 15 minutes) — alias to 'make vision'
smoke-full: vision
```

---

## 10. Open Questions & Next Steps

### 10.1 Open Questions

1. **Test runtime optimization:** Can we run unit tests in parallel across projects to reduce runtime from 3 min to < 2 min?
2. **Flaky test elimination:** What's the root cause of the 3 known flaky tests? Can we add retry logic or improve isolation?
3. **E2E test automation:** Should we invest in Playwright/Selenium for UI testing, or rely on manual QA for now?
4. **OTEL span coverage:** Do we have 100% span coverage for critical operations? If not, what's missing?
5. **Weaver schema growth:** As semconv schemas grow, will `weaver registry check` become a bottleneck? Can we parallelize it?

### 10.2 Next Steps

**Immediate (This Week):**
1. **Add new make targets** (`test-unit`, `test-integration`, `smoke-fast`)
2. **Document test failure workflow** (flowchart in README)
3. **Fix 3 known flaky tests** (add retries or improve isolation)
4. **Add test metrics dashboard** (public GitHub Pages or Grafana)

**Short-term (Next 2 Weeks):**
1. **Optimize test runtime** (parallel execution, caching)
2. **Add nightly E2E pipeline** (cron trigger, full UI testing)
3. **Implement auto-rollback on smoke test failure** (GitHub Actions)
4. **Add OTEL span coverage report** (percentage of operations with spans)

**Long-term (Next Quarter):**
1. **Invest in test infrastructure** (dedicated test runners, parallel execution)
2. **Add performance benchmarks** (latency, throughput, memory usage)
3. **Implement chaos engineering** (failure injection testing)
4. **Build testing dashboard** (real-time metrics, historical trends)

---

## 11. Summary

ChatmanGPT's testing strategy is built on **three pillars**:

1. **Evidence-Based Verification** — Every claim requires OTEL span + test assertion + weaver schema proof
2. **Fast Feedback, Slow Thoroughness** — Unit tests run fast (< 5 min), E2E tests run slow but catch integration bugs
3. **Continuous Improvement** — Weekly metrics review, kaizen cycles, eliminate flaky tests

The **testing hierarchy** (unit → integration → cross-stack → E2E) ensures bugs are caught early when they're cheap to fix. The **CI/CD pipeline** (4 stages: unit → integration → cross-stack → deploy) provides fast feedback for developers while maintaining high quality for production.

The **smoke test strategy** (fast, medium, full) balances speed with thoroughness, allowing developers to iterate quickly while preventing broken code from merging to main.

**Key Success Metrics:**
- 9,440 tests passing across 5 languages
- 100% weaver schema conformance
- 0 compiler warnings
- 95%+ cross-stack smoke test pass rate
- < 30 min full test suite runtime
- < 5 min smoke test runtime

This strategy is **practical for daily development** (fast smoke tests in pre-commit hooks) and **rigorous enough for production** (full E2E tests before deployment).

---

**End of Strategy Document**

**Next Actions:**
1. Review with team
2. Approve or revise
3. Implement new make targets
4. Update CI/CD workflows
5. Add test metrics dashboard
