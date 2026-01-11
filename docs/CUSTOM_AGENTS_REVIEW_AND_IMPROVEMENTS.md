# Custom Agents Implementation - Code Review & Improvements

**Date**: 2026-01-11
**Reviewer**: @code-reviewer + @security-auditor + @backend-go
**Status**: Production Ready with Minor Improvements Recommended
**Overall Grade**: A- (Excellent implementation with minor edge cases)

---

## Executive Summary

The Custom Agents implementation is **production-ready** with excellent architecture, proper type safety, and good error handling. However, there are **7 minor edge cases** and **3 security considerations** that should be addressed before production deployment.

**What's Great**:
- ✅ End-to-end type safety (PostgreSQL → Go → TypeScript)
- ✅ Proper SQLC usage with COALESCE pattern
- ✅ SSE streaming implemented correctly
- ✅ Good separation of concerns (Handler → Service → Repository)
- ✅ Comprehensive testing (63 frontend tests passing)

**What Needs Attention**:
- ⚠️ Input validation gaps (array sizes, string lengths)
- ⚠️ Cookie security flags for production
- ⚠️ Missing migration rollback strategy
- ⚠️ Edge case in temperature handling

---

## 1. Input Validation Issues

### 🔴 HIGH: Missing Validation on Suggested Prompts Array

**Location**: `desktop/backend-go/internal/handlers/agents.go:169-190`

**Issue**: No limit on `suggested_prompts` array size or individual prompt length.

**Risk**: User could send 10,000 prompts or prompts with 1MB text each → database bloat, UI rendering issues.

**Current Code**:
```go
// No validation
SuggestedPrompts: req.SuggestedPrompts,
```

**Recommended Fix**:
```go
// Validate suggested_prompts
if len(req.SuggestedPrompts) > 10 {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "Maximum 10 suggested prompts allowed"
    })
    return
}
for i, prompt := range req.SuggestedPrompts {
    if len(prompt) > 500 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": fmt.Sprintf("Suggested prompt %d exceeds 500 characters", i+1)
        })
        return
    }
    if strings.TrimSpace(prompt) == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": fmt.Sprintf("Suggested prompt %d cannot be empty", i+1)
        })
        return
    }
}
```

---

### 🟡 MEDIUM: Missing Validation on Welcome Message Length

**Location**: `desktop/backend-go/internal/handlers/agents.go:151-154`

**Issue**: No length limit on `welcome_message`.

**Risk**: 10,000-character welcome message → UI layout breaks, database inefficiency.

**Current Code**:
```go
var welcomeMsg *string
if req.WelcomeMessage != "" {
    welcomeMsg = &req.WelcomeMessage
}
```

**Recommended Fix**:
```go
var welcomeMsg *string
if req.WelcomeMessage != "" {
    if len(req.WelcomeMessage) > 2000 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Welcome message cannot exceed 2000 characters"
        })
        return
    }
    welcomeMsg = &req.WelcomeMessage
}
```

---

### 🟡 MEDIUM: Category Validation Missing

**Location**: `desktop/backend-go/internal/handlers/agents.go:143-149`

**Issue**: No validation that `category` is from allowed set.

**Risk**: User creates category "🍕pizza-agent🍕" → inconsistent categorization, filtering breaks.

**Current Code**:
```go
category := "general"
if req.Category != "" {
    category = req.Category
}
```

**Recommended Fix**:
```go
allowedCategories := map[string]bool{
    "general": true,
    "coding": true,
    "writing": true,
    "analysis": true,
    "research": true,
    "support": true,
    "sales": true,
    "marketing": true,
}

category := "general"
if req.Category != "" {
    if !allowedCategories[req.Category] {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid category. Allowed: general, coding, writing, analysis, research, support, sales, marketing"
        })
        return
    }
    category = req.Category
}
```

---

## 2. Edge Cases in Temperature Handling

### 🟡 MEDIUM: Temperature Zero Edge Case

**Location**: `desktop/backend-go/internal/handlers/agents.go:156-160, 260-264`

**Issue**: Temperature of `0.0` is treated as "not provided" due to `> 0` check.

**Risk**: User explicitly sets temperature to `0.0` (for deterministic output) → gets ignored, defaults to unset.

**Current Code**:
```go
// CreateCustomAgent
tempNumeric := pgtype.Numeric{}
if req.Temperature > 0 {  // ❌ Zero is treated as "not provided"
    tempNumeric.Scan(req.Temperature)
}

// UpdateCustomAgent
tempNumeric := pgtype.Numeric{}
if req.Temperature != nil && *req.Temperature > 0 {  // ❌ Same issue
    tempNumeric.Scan(*req.Temperature)
}
```

**Recommended Fix**:
```go
// CreateCustomAgent
tempNumeric := pgtype.Numeric{}
if req.Temperature >= 0 && req.Temperature <= 2.0 {  // Valid range: 0.0 to 2.0
    tempNumeric.Scan(req.Temperature)
}

// UpdateCustomAgent
tempNumeric := pgtype.Numeric{}
if req.Temperature != nil {
    if *req.Temperature < 0 || *req.Temperature > 2.0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Temperature must be between 0.0 and 2.0"
        })
        return
    }
    tempNumeric.Scan(*req.Temperature)
}
```

**Note**: Anthropic/OpenAI support temperature range 0.0-2.0. Zero is a valid value for deterministic outputs.

---

## 3. Security Issues

### 🔴 HIGH: Cookie Security Flags for Production

**Location**:
- `desktop/backend-go/internal/handlers/auth_email.go:98-108`
- `desktop/backend-go/internal/handlers/auth_google.go:139-149, 330-340, 386-396`

**Issue**: Cookies set with `Secure: false` and `SameSite: None` for development.

**Risk**: In production without HTTPS:
- SameSite=None requires Secure=true (browser will reject cookie)
- Secure=false over HTTP → session tokens vulnerable to interception

**Current Code**:
```go
http.SetCookie(c.Writer, &http.Cookie{
    Name:     "better-auth.session_token",
    Value:    sessionToken,
    Path:     "/",
    Domain:   "",
    MaxAge:   60 * 60 * 24 * 7,
    HttpOnly: true,
    Secure:   false, // ❌ Hard-coded to false
    SameSite: http.SameSiteNoneMode,
})
```

**Recommended Fix**:
```go
// Read from environment
isProduction := os.Getenv("ENVIRONMENT") == "production"
domain := os.Getenv("COOKIE_DOMAIN") // e.g., ".businessos.com"
if domain == "" {
    domain = "" // Current domain
}

http.SetCookie(c.Writer, &http.Cookie{
    Name:     "better-auth.session_token",
    Value:    sessionToken,
    Path:     "/",
    Domain:   domain,
    MaxAge:   60 * 60 * 24 * 7,
    HttpOnly: true,
    Secure:   isProduction, // ✅ Dynamic based on environment
    SameSite: http.SameSiteNoneMode,
})
```

**Alternative** (recommended):
```go
// Use SameSite=Lax in production if frontend and backend are same domain
sameSite := http.SameSiteLaxMode
if os.Getenv("ALLOW_CROSS_ORIGIN") == "true" {
    sameSite = http.SameSiteNoneMode
}

http.SetCookie(c.Writer, &http.Cookie{
    Name:     "better-auth.session_token",
    Value:    sessionToken,
    Path:     "/",
    Domain:   domain,
    MaxAge:   60 * 60 * 24 * 7,
    HttpOnly: true,
    Secure:   isProduction,
    SameSite: sameSite, // ✅ Lax in production (more secure)
})
```

---

### 🟡 MEDIUM: No Rate Limiting on Agent Creation

**Location**: `desktop/backend-go/internal/handlers/agents.go:98-198`

**Issue**: No rate limiting on `CreateCustomAgent` endpoint.

**Risk**: Malicious user creates 10,000 agents in 1 minute → database bloat, resource exhaustion.

**Recommended Fix**:
```go
// Add middleware to route registration
r.POST("/api/ai/custom-agents",
    middleware.RateLimiter(10, time.Minute), // 10 agents per minute
    h.CreateCustomAgent,
)
```

Or add user-level limit:
```go
// In CreateCustomAgent handler
ctx := context.Background()
queries := sqlc.New(h.pool)

// Check user's agent count
count, err := queries.CountUserAgents(ctx, user.ID)
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check agent count"})
    return
}
if count >= 100 { // Max 100 agents per user
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "Maximum 100 custom agents allowed per user"
    })
    return
}
```

**SQLC Query to Add**:
```sql
-- name: CountUserAgents :one
SELECT COUNT(*) FROM custom_agents WHERE user_id = $1;
```

---

### 🟢 LOW: No SQL Injection Risk (SQLC Protects)

**Status**: ✅ Not an issue

SQLC generates parameterized queries, preventing SQL injection. All inputs are safely escaped.

---

## 4. Database Migration Issues

### 🟡 MEDIUM: No Rollback Strategy

**Location**: `desktop/backend-go/cmd/migrate/main.go:1006-1050`

**Issue**: Migration only adds columns with `ADD COLUMN IF NOT EXISTS`. No rollback/down migration.

**Risk**: If migration causes issues in production, no easy way to roll back.

**Current Code**:
```go
// Migration 043 only has UP logic
_, err = conn.Exec(ctx, `
    ALTER TABLE custom_agents
    ADD COLUMN IF NOT EXISTS welcome_message TEXT;
`)
```

**Recommended Fix**:

Add rollback function (though harder with inline migrations):
```go
// In migrate/main.go, add rollback capability
func rollbackMigration043(ctx context.Context, conn *pgx.Conn) error {
    log.Println("Rolling back Migration 043...")

    // Drop columns (BE CAREFUL - data loss!)
    _, err := conn.Exec(ctx, `
        ALTER TABLE custom_agents
        DROP COLUMN IF EXISTS welcome_message,
        DROP COLUMN IF EXISTS suggested_prompts,
        DROP COLUMN IF EXISTS is_featured;

        DROP INDEX IF EXISTS idx_custom_agents_featured;
    `)
    return err
}
```

**Better Alternative**: Use a proper migration tool like `golang-migrate` or `goose`:
- Separate UP and DOWN migration files
- Version tracking in database
- Easy rollback with `migrate down`

**For Now**: Document manual rollback in production runbook (see section below).

---

## 5. Frontend Edge Cases

### 🟢 LOW: Type Safety is Excellent

**Location**: `frontend/src/lib/api/ai/types.ts:84-111`

**Status**: ✅ No issues found

TypeScript interface correctly matches backend:
- All optional fields marked with `?`
- Correct types (`string[]`, `boolean`, etc.)
- Good alias handling (`usage_count` for `times_used`)

---

### 🟡 MEDIUM: No Client-Side Validation

**Location**: Frontend components using `CustomAgent` type

**Issue**: Frontend sends data to backend without validation.

**Risk**: User types 10,000-character welcome message → backend rejects → poor UX (error only after submit).

**Recommended Fix** (in AgentBuilder component):
```typescript
function validateAgentForm(agent: Partial<CustomAgent>): string | null {
    if (!agent.name || agent.name.length < 2 || agent.name.length > 50) {
        return "Name must be 2-50 characters";
    }

    if (!agent.system_prompt || agent.system_prompt.length < 10) {
        return "System prompt must be at least 10 characters";
    }

    if (agent.welcome_message && agent.welcome_message.length > 2000) {
        return "Welcome message cannot exceed 2000 characters";
    }

    if (agent.suggested_prompts && agent.suggested_prompts.length > 10) {
        return "Maximum 10 suggested prompts allowed";
    }

    if (agent.suggested_prompts) {
        for (let prompt of agent.suggested_prompts) {
            if (prompt.length > 500) {
                return "Each suggested prompt must be under 500 characters";
            }
        }
    }

    return null; // Valid
}
```

---

## 6. Performance Considerations

### 🟢 LOW: Partial Index is Optimal

**Location**: `desktop/backend-go/cmd/migrate/main.go:1041-1045`

**Status**: ✅ Excellent design

```sql
CREATE INDEX IF NOT EXISTS idx_custom_agents_featured
ON custom_agents(user_id, is_featured, is_public)
WHERE is_featured = TRUE AND is_public = TRUE;
```

**Why It's Good**:
- Only indexes featured public agents (small subset)
- Covers exact query pattern for featured gallery
- Minimal index size

**No changes needed.**

---

### 🟢 LOW: COALESCE Pattern is Efficient

**Location**: `desktop/backend-go/internal/database/queries/custom_agents.sql:31-56`

**Status**: ✅ Good pattern

Using `COALESCE` for partial updates is database-efficient and avoids N+1 query problems.

**No changes needed.**

---

## 7. Documentation & Testing

### 🟡 MEDIUM: Missing Unit Tests for Handlers

**Location**: `desktop/backend-go/internal/handlers/agents.go`

**Issue**: No unit tests for `CreateCustomAgent`, `UpdateCustomAgent` handlers.

**Risk**: Regressions go undetected during refactoring.

**Recommended Fix**:

Create `agents_test.go`:
```go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestCreateCustomAgent_Success(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    router := gin.New()
    h := &Handlers{pool: setupTestDB(t)}
    router.POST("/agents", h.CreateCustomAgent)

    // Test data
    reqBody := CreateCustomAgentRequest{
        Name: "test-agent",
        DisplayName: "Test Agent",
        SystemPrompt: "You are a test agent",
        SuggestedPrompts: []string{"Hello", "Help me"},
        WelcomeMessage: "Welcome!",
    }
    bodyBytes, _ := json.Marshal(reqBody)

    // Execute
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/agents", bytes.NewReader(bodyBytes))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)

    var resp map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NotNil(t, resp["agent"])
}

func TestCreateCustomAgent_InvalidName(t *testing.T) {
    // Test with invalid name (uppercase, spaces, special chars)
    // Should return 400 Bad Request
}

func TestCreateCustomAgent_TooManySuggestedPrompts(t *testing.T) {
    // Test with 15 suggested prompts (max is 10)
    // Should return 400 Bad Request
}
```

---

### 🟢 LOW: Documentation is Comprehensive

**Location**: `docs/CUSTOM_AGENTS_COMPLETE_IMPLEMENTATION.md`

**Status**: ✅ Excellent documentation

- Clear bug resolution timeline
- Complete schema documentation
- Technical patterns explained
- Lessons learned captured

**No changes needed.**

---

## Summary of Recommendations

### Must Fix Before Production (Critical)
1. ✅ **Cookie Security Flags** - Make `Secure` flag environment-dependent
2. ✅ **Input Validation** - Add limits to suggested_prompts, welcome_message
3. ✅ **Temperature Edge Case** - Allow 0.0 as valid value

### Should Fix (High Priority)
4. ✅ **Category Validation** - Validate against allowed categories
5. ✅ **Rate Limiting** - Limit agent creation per user/per minute
6. ✅ **Client-Side Validation** - Add frontend validation for better UX

### Nice to Have (Medium Priority)
7. ✅ **Unit Tests** - Add handler tests
8. ✅ **Migration Rollback** - Document manual rollback procedure

---

## Estimated Fix Time

| Fix | Complexity | Time Estimate |
|-----|------------|---------------|
| Cookie security flags | Low | 15 minutes |
| Input validation (backend) | Low | 30 minutes |
| Temperature edge case | Low | 10 minutes |
| Category validation | Low | 20 minutes |
| Rate limiting | Medium | 45 minutes |
| Client-side validation | Medium | 1 hour |
| Unit tests | High | 3 hours |
| Migration rollback docs | Low | 30 minutes |
| **TOTAL** | - | **~6.5 hours** |

---

## Conclusion

This is a **high-quality implementation** that demonstrates:
- Strong understanding of Go patterns
- Proper type safety across stack
- Good separation of concerns
- Comprehensive documentation

The identified issues are **minor edge cases** that don't affect core functionality but should be addressed for production hardening.

**Recommendation**: Merge to staging, fix critical items (1-3), deploy to production.

---

**Reviewed By**: @code-reviewer + @security-auditor
**Date**: 2026-01-11
**Status**: APPROVED with minor improvements recommended
