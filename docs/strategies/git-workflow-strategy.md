# ChatmanGPT Git Workflow Strategy
## Monorepo Integration Across 5 Projects

**Version:** 1.0.0
**Date:** 2026-03-28
**Status:** Ready for Implementation
**Applies To:** `/Users/sac/chatmangpt` monorepo + 4 submodules

---

## Executive Summary

ChatmanGPT is a **5-project monorepo** with a **linear integration chain**:
```
pm4py-rust (8090) → BusinessOS (8001) → canopy (9089) → OSA (8089) → yawlv6 (8080)
```

**Key Challenges:**
1. **4 submodules** with separate GitHub repos (BusinessOS, canopy, OSA, yawlv6)
2. **1 root project** (pm4py-rust) embedded in monorepo
3. **A2A cross-stack communication** requiring coordinated changes
4. **Integration chain dependencies** (breaking changes cascade downstream)
5. **Weaver automation** requiring synchronized semconv updates

**This strategy addresses:**
- Branch naming and lifecycle
- Submodule pointer synchronization
- Cross-cutting change coordination
- Hotfix procedures
- Release management
- CI/CD integration points

---

## Table of Contents

1. [Branch Strategy](#branch-strategy)
2. [Submodule Workflow](#submodule-workflow)
3. [Cross-Cutting Changes](#cross-cutting-changes)
4. [Hotfix Procedure](#hotfix-procedure)
5. [Release Management](#release-management)
6. [CI/CD Integration](#cicd-integration)
7. [Emergency Procedures](#emergency-procedures)
8. [Migration Checklist](#migration-checklist)

---

## Branch Strategy

### Branch Naming Convention

**Format:** `<type>/<scope>-<short-description>`

| Type | When to Use | Examples |
|------|-------------|----------|
| `feat/*` | New feature | `feat/a2a-bidirectional-stream` |
| `fix/*` | Bug fix | `fix/otel-trace-threading` |
| `refactor/*` | Code restructuring | `refactor/armstrong-supervision` |
| `test/*` | Test-only changes | `test/chicago-tdd-board` |
| `docs/*` | Documentation only | `docs/weaver-automation` |
| `chore/*` | Build/config/dependencies | `chore/submodule-pointers` |
| `claude/*` | Claude Code agent work | `claude/weaver-automation` |

**Scope Options:**
- Single project: `osa`, `bos`, `canopy`, `pm4py`, `yawl`
- Integration: `a2a`, `cross-stack`, `weaver`, `semconv`
- System: `ci`, `docs`, `deps`

### Branch Lifecycle

```
1. Create feature branch from main
   git checkout -b feat/osa-healing-diagnosis

2. Work in submodule(s) + update pointer
   cd OSA && git checkout -b feat/osa-healing-diagnosis
   # ... commits ...
   cd ../ && git add OSA

3. Commit pointer update in monorepo
   git commit -m "chore(submodule): sync OSA pointer for healing diagnosis"

4. Push to origin
   git push -u origin feat/osa-healing-diagnosis

5. Create PR (one per feature, not per submodule)
   # PR includes all submodule pointer changes

6. Merge via squash-merge to main
   # NEVER rebase
   # NEVER merge to main directly
```

### Branch Protection Rules (GitHub Settings)

**Protected Branches:** `main`

**Required:**
- [ ] Pull request reviews (1 approval)
- [ ] Status checks pass (CI/CD)
- [ ] Do not allow bypassing settings

**Allowed:**
- [ ] Force pushes to feature branches
- [ ] Squash merges to main
- [ ] Merge commits to main

**Blocked:**
- [ ] Force pushes to main
- [ ] Deletions of main
- [ ] Rebasing main

---

## Submodule Workflow

### Submodule Pointer Updates

**Golden Rule:** **Every submodule change MUST have a corresponding monorepo pointer commit.**

**Pattern: Single-Project Feature**

```bash
# 1. Create feature branch in monorepo
git checkout -b feat/osa-healing-diagnosis

# 2. Create matching branch in submodule
cd OSA
git checkout -b feat/osa-healing-diagnosis

# 3. Make changes and commit in submodule
git commit -am "feat(healing): add diagnosis classify for 11 failure modes"
git push -u origin feat/osa-healing-diagnosis

# 4. Update pointer in monorepo
cd ..
git add OSA
git commit -m "chore(submodule): sync OSA pointer for healing diagnosis

Submodule: OSA@<commit-sha>
Feature: feat/osa-healing-diagnosis
Tests: mix test (121 passing)
OTEL: healing.diagnosis spans verified"

# 5. Push monorepo branch
git push -u origin feat/osa-healing-diagnosis

# 6. Create PR in monorepo (includes pointer update)
# PR body references submodule PR
```

**Pattern: Cross-Cutting Feature (2+ Submodules)**

```bash
# 1. Create feature branch in monorepo
git checkout -b feat/a2a-bidirectional-stream

# 2. Create matching branches in ALL affected submodules
cd OSA && git checkout -b feat/a2a-bidirectional-stream
cd ../canopy && git checkout -b feat/a2a-bidirectional-stream
cd ../BusinessOS && git checkout -b feat/a2a-bidirectional-stream

# 3. Make changes in dependency order (pm4py → bos → canopy → osa)
cd ../../pm4py-rust && git commit -am "feat(a2a): add response streaming"
cd ../BusinessOS && git commit -am "feat(a2a): wire response streaming"
cd ../canopy && git commit -am "feat(a2a): support response streaming"
cd ../OSA && git commit -am "feat(a2a): consume response streaming"

# 4. Push all submodule branches
git push -u origin feat/a2a-bidirectional-stream # repeat for each

# 5. Update ALL pointers in monorepo
cd /Users/sac/chatmangpt
git add OSA canopy BusinessOS pm4py-rust
git commit -m "chore(submodule): sync pointers for a2a bidirectional streaming

Submodules:
  - pm4py-rust@<sha>
  - BusinessOS@<sha>
  - canopy@<sha>
  - OSA@<sha>

Feature: feat/a2a-bidirectional-stream
Tests:
  - pm4py-rust: cargo test (30 passing)
  - BusinessOS: go test (48 passing)
  - canopy: mix test (85 passing)
  - OSA: mix test (121 passing)
  - E2E: bash scripts/mcp-a2a-smoke-test.sh (10/10 passing)

Integration chain verified: pm4py → bos → canopy → osa"

# 6. Create monorepo PR with all pointer updates
```

### Submodule Pointer Verification

**Before every commit to main:**

```bash
# 1. Check submodule status
git submodule status

# Expected output:
#  <commit-sha> OSA (<branch>)
#  <commit-sha> canopy (<branch>)
#  <commit-sha> BusinessOS (<branch>)
#  <commit-sha> yawlv6 (<branch>)

# 2. Verify no detached HEAD states
cd OSA && git status
cd ../canopy && git status
cd ../BusinessOS && git status
cd ../yawlv6 && git status

# 3. Verify integration chain
bash scripts/mcp-a2a-smoke-test.sh
bash scripts/vision2030-smoke-test.sh
```

### Submodule Initialization for New Developers

```bash
# Clone monorepo
git clone https://github.com/seanchatmangpt/chatmangpt.git
cd chatmangpt

# Initialize submodules
git submodule update --init --recursive

# Verify all submodules checked out
git submodule status
```

---

## Cross-Cutting Changes

### When A Change Affects Multiple Projects

**Scenario 1: Shared Protocol Change (A2A, MCP, OTEL)**

**Example:** Adding a new A2A action `tasks_send`

```
Affected Projects: OSA (defines), canopy (consumes), BusinessOS (consumes)

Workflow:
1. Design phase: Create feat/a2a-tasks-send in monorepo
2. Implement in OSA first (upstream dependency)
   - Define protocol in lib/osa/a2a/protocol.ex
   - Add semconv span: a2a.tasks_send
   - Test: test/a2a/task_stream_test.exs
3. Update OSA pointer in monorepo
4. Implement in canopy (downstream consumer)
   - Wire protocol in adapters/mcp.ex
   - Test: test/a2a/canopy_tasks_send_test.exs
5. Update canopy pointer in monorepo
6. Implement in BusinessOS (downstream consumer)
   - Add Go handler in handlers/a2a.go
   - Test: handlers/a2a_test.go
7. Update BusinessOS pointer in monorepo
8. E2E verification: scripts/mcp-a2a-smoke-test.sh
9. Single PR to monorepo with all pointer updates
```

**Scenario 2: Weaver Semconv Update**

```
Affected Projects: ALL (semconv/ copied to each)

Workflow:
1. Update semconv in OSA (canonical source)
   - Edit OSA/semconv/model/<domain>/spans.yaml
   - Run: weaver registry check -r ./semconv/model -p ./semconv/policies
   - Commit: feat(weaver): add healing.diagnosis span
2. Copy semconv to other projects
   - cp -r OSA/semconv canopy/semconv
   - cp -r OSA/semconv BusinessOS/semconv
   - cp -r OSA/semconv pm4py-rust/semconv
3. Update all pointers in monorepo
4. Single PR: chore(semconv): sync weaver schemas across all projects
```

**Scenario 3: Dependency Upgrade (e.g., Elixir 1.17 → 1.18)**

```
Affected Projects: OSA, canopy (both Elixir)

Workflow:
1. Test upgrade in isolation (OSA first)
   - cd OSA && git checkout -b chore/elixir-1.18-upgrade
   - Update .tool-versions: elixir 1.18.0
   - Run: mix deps.get && mix test
   - If tests pass, push and update pointer
2. Test upgrade in canopy
   - cd canopy && git checkout -b chore/elixir-1.18-upgrade
   - Update .tool-versions
   - Run: mix deps.get && mix test
   - If tests pass, push and update pointer
3. E2E verification: bash scripts/vision2030-smoke-test.sh
4. Single PR: chore(deps): upgrade Elixir 1.17 → 1.18
```

### Cross-Project PR Template

```markdown
## Summary
- [ ] feat(a2a): add tasks_send action to A2A protocol
- [ ] Implements: [GitHub Issue #XX]

## Changes by Project

### OSA (Protocol Definition)
- Commits: 3
- SHA: `abc123def`
- Tests: 12 passing
- PR: seanchatmangpt/OSA#XX

### canopy (Protocol Consumer)
- Commits: 2
- SHA: `def456ghi`
- Tests: 8 passing
- PR: seanchatmangpt/canopy#XX

### BusinessOS (Protocol Consumer)
- Commits: 4
- SHA: `ghi789jkl`
- Tests: 15 passing
- PR: seanchatmangpt/BusinessOS#XX

## Integration Chain Verification
- [ ] pm4py-rust health: curl http://localhost:8090/api/health
- [ ] BusinessOS health: curl http://localhost:8001/healthz
- [ ] canopy health: curl http://localhost:9089/health
- [ ] OSA health: curl http://localhost:8089/health
- [ ] A2A smoke test: bash scripts/mcp-a2a-smoke-test.sh (10/10 passing)
- [ ] Vision 2030 smoke test: bash scripts/vision2030-smoke-test.sh (16/16 passing)

## Submodule Pointers
- OSA: abc123def (feat/a2a-tasks-send)
- canopy: def456ghi (feat/a2a-tasks-send)
- BusinessOS: ghi789jkl (feat/a2a-tasks-send)
```

---

## Hotfix Procedure

### Hotfix Branch Naming

**Format:** `hotfix/<version>-<critical-issue>`

**Examples:**
- `hotfix/v1.2.3-otel-trace-leak`
- `hotfix/v1.2.3-a2a-deadlock`

### Hotfix Workflow

**Scenario:** Critical bug in production affecting all projects

```bash
# 1. Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix/v1.2.3-otel-trace-leak

# 2. Create matching hotfix branches in affected submodules
cd OSA && git checkout -b hotfix/v1.2.3-otel-trace-leak
cd ../canopy && git checkout -b hotfix/v1.2.3-otel-trace-leak

# 3. Apply fix in submodule(s)
cd OSA
# Edit fix
git commit -am "fix(otel): prevent trace_id leak in async_stream_nolink"
git push -u origin hotfix/v1.2.3-otel-trace-leak

# 4. Update pointer in monorepo
cd ..
git add OSA
git commit -m "chore(submodule): sync OSA pointer for hotfix v1.2.3-otel-trace-leak

Fix: OSA@<commit-sha>
Tests: mix test (8433 passing)
OTEL: Verified no trace leaks in Jaeger"

# 5. Create PR (expedited review)
# PR title: [HOTFIX] v1.2.3 - Fix OTEL trace leak in async_stream_nolink

# 6. Merge to main via squash-merge (bypass CI if critical)
# 7. Tag release: git tag -a v1.2.4 -m "Hotfix: OTEL trace leak"
# 8. Push tag: git push origin v1.2.4

# 9. Cherry-pick to release branch if exists
git checkout release/v1.2.x
git cherry-pix -x <commit-sha>
git push origin release/v1.2.x
```

### Hotfix Checklist

- [ ] Bug reproduced locally
- [ ] Fix tested in isolation (unit tests)
- [ ] Fix tested in integration (E2E smoke tests)
- [ ] Submodule pointers updated
- [ ] PR created with [HOTFIX] prefix
- [ ] Expedited review requested
- [ ] Merge to main approved
- [ ] Release tag created
- [ ] Deployment to production verified
- [ ] Post-deployment smoke tests pass

---

## Release Management

### Versioning Strategy

**Format:** Semantic Versioning 2.0.0

```
MAJOR.MINOR.PATCH

Examples:
  v1.2.3  → Major 1, Minor 2, Patch 3
  v2.0.0  → Breaking changes
  v1.3.0  → New features (backward compatible)
  v1.2.4  → Bug fixes (backward compatible)
```

**Monorepo Version:** Single version for all projects (synchronized)

**Submodule Versions:** Track independently, but aligned for releases

### Release Branches

**Create Release Branch:**

```bash
# 1. Create release branch from main
git checkout main
git pull origin main
git checkout -b release/v1.3.0

# 2. Update version files
cd OSA && echo "1.3.0" > VERSION && git add VERSION
cd ../canopy && echo "1.3.0" > VERSION && git add VERSION
cd ../BusinessOS && echo "1.3.0" > VERSION && git add VERSION

# 3. Update submodule pointers to release candidates
cd ..
git add OSA canopy BusinessOS
git commit -m "chore(release): v1.3.0 release candidates

Submodules:
  - OSA@<sha> (v1.3.0-rc1)
  - canopy@<sha> (v1.3.0-rc1)
  - BusinessOS@<sha> (v1.3.0-rc1)"

# 4. Push release branch
git push -u origin release/v1.3.0

# 5. Create PR for release
# PR title: [RELEASE] v1.3.0 - Release Candidates
```

**Release Branch Lifecycle:**

```
release/v1.3.0 (branch)
  ├─ rc1 commits (bug fixes)
  ├─ rc2 commits (bug fixes)
  └─ final release commit (merged to main)

When stable:
  1. Merge release/v1.3.0 to main (squash merge)
  2. Tag main: git tag -a v1.3.0 -m "Release v1.3.0"
  3. Push tag: git push origin v1.3.0
  4. Delete release branch: git branch -d release/v1.3.0
```

### Release Checklist

**Pre-Release:**
- [ ] All tests passing (mix test, go test, cargo test)
- [ ] E2E smoke tests passing (A2A, Vision 2030)
- [ ] Compiler warnings: 0 (mix compile --warnings-as-errors, go vet, cargo clippy)
- [ ] Submodule pointers synchronized
- [ ] Documentation updated (CHANGELOG.md, migration guides)
- [ ] Breaking changes documented
- [ ] Performance benchmarks pass

**Release:**
- [ ] Release branch created from main
- [ ] Version numbers updated in all projects
- [ ] CHANGELOG.md updated
- [ ] Release notes drafted
- [ ] Release PR created
- [ ] Release approved and merged to main
- [ ] Git tag created and pushed
- [ ] Submodule releases tagged (OSA v1.3.0, canopy v1.3.0, etc.)

**Post-Release:**
- [ ] Deployed to production
- [ ] Post-deployment smoke tests pass
- [ ] Monitoring dashboards green
- [ ] Release announcement sent
- [ ] Next version planning started

---

## CI/CD Integration

### GitHub Actions Workflow

**Triggered On:** Pull requests to main, pushes to main

**Jobs:**

```yaml
name: ChatmanGPT CI

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  # Job 1: Verify submodule pointers
  verify-submodules:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive
      - name: Check submodule status
        run: |
          git submodule status
          if git submodule status | grep -q "^-"; then
            echo "ERROR: Submodules not initialized"
            exit 1
          fi
          if git submodule status | grep -q "^+"; then
            echo "WARNING: Submodule pointers ahead of remote"
          fi

  # Job 2: Run pm4py-rust tests
  test-pm4py:
    runs-on: ubuntu-latest
    needs: verify-submodules
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive
      - name: Install Rust
        run: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
      - name: Run tests
        run: |
          cd pm4py-rust
          cargo test --verbose

  # Job 3: Run BusinessOS tests
  test-businessos:
    runs-on: ubuntu-latest
    needs: verify-submodules
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive
      - name: Install Go
        run: wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
      - name: Run tests
        run: |
          cd BusinessOS/desktop/backend-go
          go test ./...

  # Job 4: Run canopy tests
  test-canopy:
    runs-on: ubuntu-latest
    needs: verify-submodules
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive
      - name: Install Erlang/Elixir
        run: |
          wget https://packages.erlang-solutions.com/erlang-solutions_2.0_all.deb
          sudo dpkg -i erlang-solutions_2.0_all.deb
          sudo apt-get install elixir
      - name: Run tests
        run: |
          cd canopy
          mix test

  # Job 5: Run OSA tests
  test-osa:
    runs-on: ubuntu-latest
    needs: verify-submodules
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive
      - name: Install Erlang/Elixir
        run: |
          wget https://packages.erlang-solutions.com/erlang-solutions_2.0_all.deb
          sudo dpkg -i erlang-solutions_2.0_all.deb
          sudo apt-get install elixir
      - name: Run tests
        run: |
          cd OSA
          mix test

  # Job 6: Verify integration chain
  verify-integration:
    runs-on: ubuntu-latest
    needs: [test-pm4py, test-businessos, test-canopy, test-osa]
    steps:
      - uses: actions/checkout@v3
        with:
          submodules: recursive
      - name: Boot services
        run: |
          cd BusinessOS
          make dev
      - name: Run A2A smoke tests
        run: bash scripts/mcp-a2a-smoke-test.sh
      - name: Run Vision 2030 smoke tests
        run: bash scripts/vision2030-smoke-test.sh
```

### Pre-Commit Hook (Weaver Verification)

**File:** `.git/hooks/pre-commit`

```bash
#!/bin/bash
# Pre-commit hook: Verify weaver schemas before commit

echo "🔍 Running weaver registry check..."

# Check if semconv changes present
if git diff --cached --name-only | grep -q "semconv/"; then
  # Run weaver check in each project with semconv
  for project in OSA canopy BusinessOS pm4py-rust; do
    if [ -d "$project/semconv" ]; then
      echo "Checking $project/semconv..."
      cd "$project"
      weaver registry check -r ./semconv/model -p ./semconv/policies --quiet
      if [ $? -ne 0 ]; then
        echo "❌ Weaver registry check failed in $project"
        exit 1
      fi
      cd ..
    fi
  done
  echo "✅ Weaver registry checks passed"
fi

exit 0
```

**Install:**

```bash
cp .git/hooks/pre-commit.sample .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Pre-Push Hook (Submodule Verification)

**File:** `.git/hooks/pre-push`

```bash
#!/bin/bash
# Pre-push hook: Verify submodule pointers and integration chain

echo "🔍 Verifying submodule pointers..."

# Check for detached HEAD states
for submodule in OSA canopy BusinessOS yawlv6; do
  cd "$submodule"
  if git status | grep -q "detached"; then
    echo "❌ ERROR: $submodule is in detached HEAD state"
    echo "Run: cd $submodule && git checkout main"
    exit 1
  fi
  cd ..
done

echo "✅ Submodule pointers verified"

exit 0
```

**Install:**

```bash
cp .git/hooks/pre-push.sample .git/hooks/pre-push
chmod +x .git/hooks/pre-push
```

---

## Emergency Procedures

### Scenario: Accidental Force Push to main

**Symptom:** `git reflog` shows force push to main

**Recovery:**

```bash
# 1. Identify lost commit
git reflog | grep "before force push"

# 2. Reset main to good commit
git checkout main
git reset --hard <good-commit-sha>

# 3. Force push (careful!)
git push origin main --force

# 4. Verify all submodules
git submodule update --recursive
```

**Prevention:**

- GitHub branch protection rules (block force pushes to main)
- Pre-push hook (warn on force push)
- Team training (never force push to main)

### Scenario: Submodule Pointer Desync

**Symptom:** `git submodule status` shows mixed branches

**Recovery:**

```bash
# 1. Check each submodule state
git submodule status

# 2. For each submodule in wrong branch:
cd OSA
git checkout main  # or correct feature branch
git pull origin main
cd ..

# 3. Update pointer
git add OSA
git commit -m "chore(submodule): fix OSA pointer desync"
git push origin <branch>
```

**Prevention:**

- Always commit pointer updates immediately after submodule commits
- Pre-commit hook checks for detached HEAD states
- CI/CD verifies submodule status

### Scenario: Integration Chain Test Failure

**Symptom:** `bash scripts/mcp-a2a-smoke-test.sh` fails

**Recovery:**

```bash
# 1. Identify failing project
for service in pm4py-rust BusinessOS canopy OSA; do
  curl http://localhost:${PORTS[$service]}/health
done

# 2. Check logs for failing service
cd BusinessOS && make dev-logs | grep ERROR

# 3. Fix bug in appropriate project
cd OSA && git checkout -b fix/a2a-timeout
# ... fix and test ...

# 4. Update pointer and re-run integration tests
cd .. && git add OSA && git commit -m "fix(a2a): add timeout to downstream calls"
bash scripts/mcp-a2a-smoke-test.sh
```

**Prevention:**

- Run integration tests before every commit to main
- CI/CD runs integration tests on every PR
- Feature flags for breaking changes

---

## Migration Checklist

### Phase 1: Preparation (Week 1)

- [ ] Review this workflow strategy with team
- [ ] Configure GitHub branch protection rules (main)
- [ ] Install pre-commit and pre-push hooks
- [ ] Create `.github/ISSUE_TEMPLATE/` templates
- [ ] Create `.github/PULL_REQUEST_TEMPLATE.md`
- [ ] Set up GitHub Actions CI/CD pipeline
- [ ] Document integration chain smoke tests

### Phase 2: Pilot (Week 2)

- [ ] Create `feat/pilot-git-workflow` branch
- [ ] Practice submodule pointer updates (single project)
- [ ] Practice cross-cutting changes (2+ projects)
- [ ] Test hotfix procedure (non-critical)
- [ ] Verify CI/CD pipeline
- [ ] Run integration smoke tests
- [ ] Document lessons learned

### Phase 3: Rollout (Week 3)

- [ ] Merge pilot workflow to main
- [ ] Train team on new workflow
- [ ] Update CLAUDE.md with workflow references
- [ ] Create workflow cheat sheet
- [ ] Enable all branch protection rules
- [ ] Monitor CI/CD pipeline for issues

### Phase 4: Optimization (Week 4+)

- [ ] Collect metrics on merge times, failure rates
- [ ] Optimize CI/CD pipeline (parallel tests, caching)
- [ ] Automate submodule pointer updates (script)
- [ ] Add integration tests to CI/CD
- [ ] Create release automation scripts

---

## Appendix

### A. Git Aliases (Recommended)

**File:** `.gitconfig`

```ini
[alias]
  # Submodule helpers
  sync-submodules = "!f() { git submodule update --recursive --remote && git add $(git submodule foreach -q 'echo $path'); }; f"
  sub-status = submodule status
  sub-pull = "!f() { git submodule foreach -q 'git pull origin $(git rev-parse --abbrev-ref HEAD)'; }; f"

  # Integration chain helpers
  verify-integration = "!bash scripts/mcp-a2a-smoke-test.sh && bash scripts/vision2030-smoke-test.sh"
  boot-services = "!cd BusinessOS && make dev"

  # Workflow helpers
  feature = "!f() { git checkout -b feat/$1; }; f"
  hotfix = "!f() { git checkout -b hotfix/$1; }; f"
  release = "!f() { git checkout -b release/$1; }; f"
```

**Usage:**

```bash
git feature osa-healing-diagnosis
git sync-submodules
git verify-integration
```

### B. Submodule Helper Script

**File:** `scripts/sync-submodule-pointers.sh`

```bash
#!/bin/bash
# Sync all submodule pointers to current branch

set -e

SUBMODULES="OSA canopy BusinessOS yawlv6"
BRANCH=$(git rev-parse --abbrev-ref HEAD)

echo "🔄 Syncing submodule pointers for branch: $BRANCH"

for submodule in $SUBMODULES; do
  if [ -d "$submodule" ]; then
    echo "📦 Checking $submodule..."

    cd "$submodule"

    # Check if branch exists
    if git rev-parse --verify "$BRANCH" >/dev/null 2>&1; then
      git checkout "$BRANCH"
      git pull origin "$BRANCH"
    else
      echo "⚠️  Branch $BRANCH does not exist in $submodule"
      echo "Creating from main..."
      git checkout main
      git checkout -b "$BRANCH"
    fi

    cd ..
  fi
done

# Update pointers
git add $SUBMODULES
git commit -m "chore(submodule): sync pointers for $BRANCH"

echo "✅ Submodule pointers synced"
```

**Install:**

```bash
chmod +x scripts/sync-submodule-pointers.sh
```

### C. Quick Reference Card

```
┌─────────────────────────────────────────────────────────────┐
│ ChatmanGPT Git Workflow - Quick Reference                    │
├─────────────────────────────────────────────────────────────┤
│ Branch Naming: feat/*, fix/*, hotfix/*, chore/*             │
│ Commit Format: type(scope): description                      │
│                                                             │
│ Single Project:                                             │
│   1. git checkout -b feat/osa-healing                       │
│   2. cd OSA && git checkout -b feat/osa-healing             │
│   3. Make changes, commit, push                             │
│   4. cd .. && git add OSA && git commit -m "chore(submodule):..."│
│   5. git push && create PR                                  │
│                                                             │
│ Cross-Cutting:                                              │
│   1. Create matching branches in ALL affected submodules    │
│   2. Implement in dependency order (pm4py → bos → canopy → osa)│
│   3. Update ALL pointers in monorepo                        │
│   4. Single PR with all pointer updates                     │
│                                                             │
│ Hotfix:                                                     │
│   1. git checkout hotfix/v1.2.3-critical-bug                │
│   2. Fix in submodules, update pointers                     │
│   3. Expedited PR, merge to main                           │
│   4. Tag release: git tag -a v1.2.4                        │
│                                                             │
│ Verify Integration:                                         │
│   bash scripts/mcp-a2a-smoke-test.sh                        │
│   bash scripts/vision2030-smoke-test.sh                     │
│                                                             │
│ NEVER:                                                      │
│   ❌ Rebase main                                            │
│   ❌ Force push to main                                     │
│   ❌ Commit submodule changes without pointer update        │
│                                                             │
│ ALWAYS:                                                     │
│   ✅ Merge commits only (no squash to feature branches)     │
│   ✅ Update submodule pointers immediately                  │
│   ✅ Run integration tests before merge                     │
└─────────────────────────────────────────────────────────────┘
```

---

## Change Log

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-28 | Initial strategy document |

---

**Next Steps:**

1. Review this strategy with team
2. Configure GitHub branch protection rules
3. Install pre-commit/pre-push hooks
4. Set up CI/CD pipeline
5. Begin pilot phase

**Questions?** Contact: Sean Chatman <info@chatmangpt.com>
