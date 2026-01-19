---
title: Backend Refactoring Priority Quick Reference
author: Roberto Luna (with Claude Code)
created: 2026-01-19
updated: 2026-01-19
category: Backend
type: Guide
status: Active
part_of: Codebase Cleanup Initiative
relevance: Recent
---

# Backend Refactoring Priority Quick Reference

**Generated:** 2026-01-19
**Full Report:** See `DUPLICATE_CODE_ANALYSIS.md`

## Top 5 Critical Duplicates to Fix IMMEDIATELY

### 1. User Authentication Check (392 occurrences)
**Problem:** Every handler has this:
```go
user := middleware.GetCurrentUser(c)
if user == nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
    return
}
```

**Solution:** Use middleware or helper
```go
// Option 1: Middleware (RECOMMENDED)
authenticated := router.Group("/api")
authenticated.Use(middleware.RequireAuth())

// Option 2: Helper function
user, ok := RequireUser(c)
if !ok {
    return
}
```

**Impact:** Removes ~1,560 lines of duplicate code
**Effort:** 2-3 days (mostly automated find-replace)
**Files affected:** 50+ handler files

---

### 2. Session Cookie Configuration (5+ occurrences)
**Problem:** Same 20-line cookie setup in multiple places

**Solution:** Extract to helper
```go
middleware.SetAuthSessionCookie(c, sessionToken)
middleware.ClearAuthSessionCookie(c)
```

**Impact:** Removes ~120 lines
**Effort:** 1 day
**Files affected:** auth_google.go, auth_email.go

---

### 3. Random ID Generation (10+ duplicate functions)
**Problem:** Same functions duplicated across files:
- `generateSessionToken()`
- `generateUserID()`
- `generateSessionID()`
- `generateRandomState()`

**Solution:** Centralize in utils
```go
utils.GenerateSessionToken()
utils.GenerateUserID()
utils.GenerateSessionID()
utils.GenerateOAuthState()
```

**Impact:** Removes ~40 lines
**Effort:** 1 day
**Files affected:** auth_google.go, auth_email.go

---

### 4. Session Creation (2+ identical functions)
**Problem:** Duplicate `createSession()` in auth handlers

**Solution:** Create session service
```go
sessionService.CreateSession(ctx, userID)
sessionService.InvalidateSession(ctx, token)
sessionService.InvalidateUserSessions(ctx, userID)
```

**Impact:** Removes ~32 lines
**Effort:** 1 day
**Files affected:** auth_google.go, auth_email.go

---

### 5. Integration Helpers (3 packages, identical code)
**Problem:** Same helper functions in fathom, hubspot, linear packages:
- `syncStats` struct
- `containsString()` function

**Solution:** Extract to common package
```go
common.SyncStats
common.Contains[T](slice []T, item T) bool  // Generic version
```

**Impact:** Removes ~45 lines
**Effort:** 1 day
**Files affected:** 4 integration packages

---

## Quick Win Checklist

### Week 1 (Critical Security & Auth)
- [ ] Create `internal/utils/random.go` with ID generators
- [ ] Create `internal/middleware/session_cookie.go` with cookie helpers
- [ ] Create `internal/services/session_service.go`
- [ ] Update auth handlers to use new utilities
- [ ] Test auth flows thoroughly

**Estimated reduction:** ~250 lines
**Risk:** HIGH - auth changes require careful testing

---

### Week 2 (Biggest Impact)
- [ ] Implement `middleware.RequireAuth()` OR `RequireUser()` helper
- [ ] Find-replace all 392 manual auth checks
- [ ] Staged rollout (10% → 50% → 100%)
- [ ] Monitor for auth failures

**Estimated reduction:** ~1,560 lines
**Risk:** HIGH - could break all protected routes if done wrong

---

### Week 3 (Services & Integrations)
- [ ] Extract integration helpers to `internal/integrations/common/`
- [ ] Refactor LLM services to use `BaseLLMService`
- [ ] Consolidate onboarding services

**Estimated reduction:** ~400 lines
**Risk:** MEDIUM - service changes are isolated

---

### Week 4 (Polish & Documentation)
- [ ] Reorganize context services
- [ ] Update documentation
- [ ] Add CI linter rules to prevent future duplicates
- [ ] Final testing

**Estimated reduction:** ~200 lines
**Risk:** LOW - organizational changes

---

## Testing Requirements

### Critical Path Tests (Week 1-2)
```go
// Must pass before deploying auth changes
func TestGoogleOAuthFlow(t *testing.T)
func TestEmailSignupFlow(t *testing.T)
func TestSessionCreation(t *testing.T)
func TestSessionCookieBehavior(t *testing.T)
func TestMiddlewareAuth(t *testing.T)
func TestLogoutAllDevices(t *testing.T)
```

### Integration Tests (Week 3)
```go
func TestFathomSync(t *testing.T)
func TestHubspotSync(t *testing.T)
func TestLinearSync(t *testing.T)
func TestAnthropicLLM(t *testing.T)
func TestGroqLLM(t *testing.T)
```

---

## Rollback Plan

If auth changes break production:

1. **Immediate rollback:**
   ```bash
   git revert <commit-hash>
   docker build -t businessos-server .
   gcloud run deploy --image businessos-server
   ```

2. **Session cookie issues:**
   - Existing sessions may be invalidated
   - Users will need to log in again
   - NOT a security issue, just inconvenience

3. **Monitoring:**
   - Watch for 401 errors spike
   - Monitor session creation rate
   - Check login success rate

---

## Success Metrics

### Code Quality
- [ ] Total lines reduced: 3,000-5,000
- [ ] Duplicate auth checks: 392 → 0
- [ ] Single source of truth for security code
- [ ] Improved test coverage

### Developer Experience
- [ ] Faster onboarding (fewer patterns to learn)
- [ ] Easier to find reusable code
- [ ] Less copy-paste between handlers

### Maintenance
- [ ] Bug fixes in one place affect all code
- [ ] Security audits are easier
- [ ] CI/CD catches new duplicates

---

## Commands for Analysis

```bash
# Find all manual auth checks
grep -r "middleware.GetCurrentUser.*nil" internal/handlers/ | wc -l

# Find all session cookie setups
grep -r "better-auth.session_token" internal/handlers/ | wc -l

# Find all ID generation functions
grep -r "generateSessionToken\|generateUserID" internal/handlers/

# Find all integration helpers
find internal/integrations -name "helpers.go"
```

---

## Files to Create (Week 1)

```
internal/
├── utils/
│   └── random.go              # NEW: ID generation utilities
├── middleware/
│   └── session_cookie.go      # NEW: Cookie helpers
└── services/
    └── session_service.go     # NEW: Session management
```

---

## Files to Modify (Week 2)

**High priority:**
- `internal/handlers/auth_google.go` - Use new utilities
- `internal/handlers/auth_email.go` - Use new utilities
- `internal/handlers/*.go` (50+ files) - Replace manual auth checks

**Medium priority:**
- `internal/integrations/fathom/helpers.go`
- `internal/integrations/hubspot/helpers.go`
- `internal/integrations/linear/helpers.go`

---

## Next Steps

1. Review this plan with team
2. Get approval for Week 1 changes
3. Create feature branch: `refactor/auth-deduplication`
4. Implement Phase 1 (auth utilities)
5. Comprehensive testing
6. Code review
7. Staged deployment

**Estimated total time: 3-4 weeks**
**Estimated LOC reduction: 3,000-5,000 lines**
**Risk level: HIGH initially, decreases each week**

---

**Full details:** See `DUPLICATE_CODE_ANALYSIS.md`
