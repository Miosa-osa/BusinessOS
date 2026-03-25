# Authentication Quick Reference

**For developers adding new endpoints or fixing auth issues.**

---

## One-Minute Summary

**Every `/api/` endpoint must have authentication.**

```go
// ✅ CORRECT: Protected endpoint
protectedGroup := api.Group("/my-endpoint")
protectedGroup.Use(auth, middleware.RequireAuth())
{
    protectedGroup.GET("", handler.GetData)  // 401 if not authenticated
}

// ❌ WRONG: Unprotected endpoint (SECURITY RISK)
api.GET("/my-endpoint", handler.GetData)  // Anyone can access!
```

---

## How to Add a New Protected Endpoint

### Step 1: Choose Authentication Type

**Session-Based (Browsers/SPA):**
```go
h.registerMyRoutes(api, auth)  // Use 'auth' parameter

func (h *Handlers) registerMyRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
    protected := api.Group("/my-endpoints")
    protected.Use(auth, middleware.RequireAuth())
    {
        protected.GET("", h.GetMyData)
    }
}
```

**JWT Bearer (API-to-API):**
```go
jwtAuth := middleware.JWTAuth(h.cfg.SecretKey)

protected := api.Group("/webhook")
protected.Use(jwtAuth, middleware.RequireAuth())
{
    protected.POST("/receive", h.HandleWebhook)
}
```

**Optional Auth (Works with or without auth):**
```go
optional := api.Group("/browse")
optional.Use(optionalAuth)  // User in context if authenticated
{
    optional.GET("/providers", h.ListProviders)  // GetCurrentUser(c) may be nil
}
```

### Step 2: Use Authenticated User in Handler

**Session-Based:**
```go
func (h *MyHandler) GetMyData(c *gin.Context) {
    user := middleware.MustGetCurrentUser(c)  // Panics if not set (shouldn't happen)

    // Use user.ID, user.Email, user.Name
    data, err := h.repository.GetData(c.Request.Context(), user.ID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Internal server error"})
        return
    }

    c.JSON(200, data)
}
```

**JWT Bearer:**
```go
func (h *MyHandler) HandleWebhook(c *gin.Context) {
    claims := middleware.GetJWTClaims(c)  // Returns nil if not authenticated
    if claims == nil {
        c.JSON(401, gin.H{"error": "Not authenticated"})
        return
    }

    // Use claims.UserID, claims.Email
    result, err := h.service.Process(c.Request.Context(), claims.UserID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Internal server error"})
        return
    }

    c.JSON(200, result)
}
```

**Optional Auth:**
```go
func (h *MyHandler) ListProviders(c *gin.Context) {
    user := middleware.GetCurrentUser(c)

    if user != nil {
        // User authenticated — show personalized providers
        return h.getPersonalizedProviders(c, user.ID)
    } else {
        // User not authenticated — show public providers
        return h.getPublicProviders(c)
    }
}
```

---

## Public Endpoints (Exception Cases)

Only these endpoint patterns are public (no auth required):

```
/health                           ← Liveness probe
/ready                            ← Readiness probe
/health/detailed                  ← Detailed health
/healthz, /readyz                 ← Kubernetes probes
/api/auth/*                       ← Auth endpoints (login, signup, oauth)
/api/osa/health                   ← OSA health
/api/integrations/providers       ← Browse providers (no user data exposed)
/api/sorx/skills                  ← Browse skills (public catalog)
/api/sorx/callback                ← Webhook callback (validates signature internally)
```

**DO NOT add new public endpoints without explicit security review.**

---

## Testing Your Endpoint

### Unit Test Template

```go
func TestMyHandler_GetData_WithAuth(t *testing.T) {
    // Setup context with authenticated user
    c, w := setupTestContext()
    user := &middleware.BetterAuthUser{
        ID:    "test-user-123",
        Email: "test@example.com",
    }
    c.Set(middleware.UserContextKey, user)

    // Create handler and call it
    handler := NewMyHandler(mockRepo)
    handler.GetData(c)

    // Assert: Returns 200, not 401
    assert.Equal(t, 200, w.Code)
}

func TestMyHandler_GetData_WithoutAuth(t *testing.T) {
    // Setup context WITHOUT user (unauthenticated)
    c, w := setupTestContext()
    // DO NOT set UserContextKey

    // Create handler and call it
    handler := NewMyHandler(mockRepo)
    handler.GetData(c)

    // Assert: Returns 401 because RequireAuth() middleware rejects it
    assert.Equal(t, 401, w.Code)
}
```

### Manual Test (cURL)

**With Session Cookie:**
```bash
curl -H "Cookie: better-auth.session_token=YOUR_TOKEN" \
  http://localhost:8001/api/my-endpoint
```

**With JWT Bearer:**
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8001/api/my-endpoint
```

**Without Auth (should get 401):**
```bash
curl http://localhost:8001/api/my-endpoint
# Response: 401 Unauthorized
# {"error": "Authentication required", "code": "UNAUTHENTICATED"}
```

---

## Common Mistakes

### ❌ Mistake 1: Forgetting RequireAuth()
```go
// ❌ WRONG: Has auth middleware but no RequireAuth()
protected := api.Group("/data")
protected.Use(auth)  // Missing RequireAuth()!
{
    protected.GET("", handler.GetData)  // Anyone can access!
}

// ✅ CORRECT: Has both auth and RequireAuth()
protected := api.Group("/data")
protected.Use(auth, middleware.RequireAuth())
{
    protected.GET("", handler.GetData)  // Only authenticated users
}
```

### ❌ Mistake 2: Not Handling Nil User
```go
// ❌ WRONG: Assumes user always exists (panics if nil)
func (h *Handler) GetData(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    fmt.Println(user.ID)  // Panic if user is nil!
}

// ✅ CORRECT: Checks for nil before using
func (h *Handler) GetData(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    if user == nil {
        c.JSON(401, gin.H{"error": "Not authenticated"})
        return
    }
    fmt.Println(user.ID)  // Safe
}

// ✅ BETTER: Use MustGetCurrentUser if RequireAuth() is middleware
func (h *Handler) GetData(c *gin.Context) {
    user := middleware.MustGetCurrentUser(c)  // Safe because RequireAuth() guarantees user
    fmt.Println(user.ID)
}
```

### ❌ Mistake 3: Mixing Auth Types
```go
// ❌ WRONG: Using session auth on API endpoint
func setupRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
    api.POST("/webhook", auth, middleware.RequireAuth(), handleWebhook)
    // Frontend sends session cookie, but webhook from service sends JWT!
}

// ✅ CORRECT: Choose appropriate auth for endpoint
// For webhooks from external services:
jwtAuth := middleware.JWTAuth(secretKey)
api.POST("/webhook", jwtAuth, middleware.RequireAuth(), handleWebhook)
```

### ❌ Mistake 4: Hardcoding Secrets
```go
// ❌ WRONG: Secret in code
const Secret = "my-secret-key"
jwtAuth := middleware.JWTAuth(Secret)

// ✅ CORRECT: Secret from environment
jwtAuth := middleware.JWTAuth(cfg.SecretKey)  // Loaded from env
```

---

## HTTP Status Codes

| Code | Scenario | Response |
|------|----------|----------|
| **200** | Authenticated, authorized, success | `{"data": {...}}` |
| **400** | Bad request (validation error) | `{"error": "...", "code": "..."}` |
| **401** | Not authenticated (no token, expired, invalid) | `{"error": "...", "code": "JWT_INVALID"}` |
| **403** | Authenticated but not authorized (insufficient permissions) | `{"error": "...", "code": "FORBIDDEN"}` |
| **404** | Resource not found | `{"error": "..."}` |
| **429** | Rate limited (too many requests) | `{"error": "Rate limited"}` |
| **500** | Internal server error | `{"error": "Internal server error"}` |

**Remember:** 401 = "Who are you?" (authentication failed)
**Remember:** 403 = "I know who you are, but you can't do that" (authorization failed)

---

## Middleware Behavior

### AuthMiddleware (Session-Based)
```
Request
  ↓
Extract cookie: better-auth.session_token
  ↓
Query DB for session + user
  ↓
Check expiry: if expired → 401
  ↓
Check absolute max (30d): if exceeded → 401
  ↓
Refresh if < 24h remaining: UPDATE session SET expiresAt = NOW() + 7d
  ↓
Set user in context: c.Set("user", &BetterAuthUser{...})
  ↓
Next handler
```

### JWTAuth (Bearer Token)
```
Request
  ↓
Extract header: Authorization: Bearer <TOKEN>
  ↓
Parse JWT: header.payload.signature
  ↓
Verify signature: HMAC-SHA256(header.payload, secretKey)
  ↓
Validate claims: user_id, email, exp, iat, nbf
  ↓
If invalid/expired → 401
  ↓
Set claims in context: c.Set("jwt_claims", &JWTClaims{...})
  ↓
Next handler
```

### RequireAuth (Guard)
```
After auth middleware, checks if user/claims exist
  ↓
If user is nil: c.AbortWithStatusJSON(401, {"error": "..."})
  ↓
If user exists: c.Next()
```

---

## Debugging Auth Issues

### Check logs (debug level)
```bash
RUST_LOG=debug go run ./cmd/server
# Look for: "AuthMiddleware: user authenticated" or "JWT: token validated"
```

### Test with curl
```bash
# 1. Get current session
curl -i http://localhost:8001/api/auth/session

# 2. Get CSRF token
curl -i http://localhost:8001/api/auth/csrf

# 3. Try protected endpoint without auth
curl -i http://localhost:8001/api/chat  # Should return 401

# 4. Try protected endpoint with invalid token
curl -i -H "Authorization: Bearer invalid" \
  http://localhost:8001/api/chat  # Should return 401
```

### Check database (Better Auth)
```sql
-- List active sessions
SELECT u.id, u.email, s.token, s."expiresAt"
FROM session s
JOIN "user" u ON s."userId" = u.id
WHERE s."expiresAt" > NOW()
ORDER BY s."expiresAt" DESC;

-- Check user
SELECT id, email, name, "emailVerified", "createdAt"
FROM "user"
WHERE email = 'test@example.com';
```

---

## See Also

- Full docs: `SECURITY_AUDIT_ENDPOINT_PROTECTION.md`
- Summary: `SECURITY_AUDIT_SUMMARY.md`
- Middleware source: `internal/middleware/auth.go`, `internal/middleware/jwt_auth.go`
- Route examples: `internal/handlers/routes_*.go`
- Tests: `internal/middleware/*_test.go`

---

## Checklist for New Endpoints

- [ ] Endpoint uses appropriate auth middleware (session or JWT)
- [ ] `middleware.RequireAuth()` is in middleware chain
- [ ] Handler checks for user/claims (doesn't assume they exist)
- [ ] Error responses return 401 for auth failures
- [ ] Rate limiting applied if sensitive (10 req/min for auth)
- [ ] Tests verify 401 when unauthenticated
- [ ] Tests verify success when authenticated
- [ ] Database migration (if storing user-specific data)
- [ ] Security review before merge

---

**Questions?** See `SECURITY_AUDIT_ENDPOINT_PROTECTION.md` for comprehensive documentation.
