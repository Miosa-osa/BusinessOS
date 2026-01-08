# MCP Security Assessment

> **Document Version:** 2.0
> **Last Updated:** December 31, 2025
> **Status:** Action Required - Critical vulnerabilities identified

---

## Table of Contents

- [Executive Summary](#executive-summary)
- [Current Architecture](#current-architecture)
- [Security Analysis](#security-analysis)
  - [1. Unauthorized Tool Exposure](#1-unauthorized-tool-exposure)
  - [2. Session Hijacking](#2-session-hijacking)
  - [3. Tool Shadowing / Shadow MCP](#3-tool-shadowing--shadow-mcp)
  - [4. Sensitive Data Exposure & Token Theft](#4-sensitive-data-exposure--token-theft)
  - [5. Authentication Bypass](#5-authentication-bypass)
- [Priority Fix List](#priority-fix-list)
- [Implementation Recommendations](#implementation-recommendations)
- [Appendix: Key Files](#appendix-key-files)

---

## Executive Summary

| Category | Status | Risk Level |
|----------|--------|------------|
| Unauthorized Tool Exposure | Partial Protection | HIGH |
| Session Hijacking | Moderate Protection | MEDIUM |
| Tool Shadowing | Not Addressed | MEDIUM |
| Token Theft / Data Exposure | Critical Vulnerability | CRITICAL |
| Authentication Bypass | Needs Review | MEDIUM |

**Immediate Actions Required:**
1. Encrypt OAuth tokens at rest
2. Add Secure + SameSite flags to session cookies
3. Implement per-tool authorization checks

---

## Current Architecture

### MCP Routes

```
POST   /api/mcp/execute     → ExecuteMCPTool handler
GET    /api/mcp/tools       → ListMCPTools handler
GET    /api/mcp/health      → MCPHealth handler
```

All routes protected by `auth` middleware (session cookie validation).

### Available Tools (23 total)

| Category | Tools | Count |
|----------|-------|-------|
| Builtin | search_conversations, get_project_context, create_artifact, add_to_daily_log, get_context_profile, list_resources | 6 |
| Google Calendar | calendar_list_events, calendar_create_event, calendar_update_event, calendar_delete_event | 4 |
| Slack | slack_list_channels, slack_send_message, slack_get_channel_history, slack_search_messages, slack_list_users, slack_get_user_info | 6 |
| Notion | notion_list_databases, notion_get_database, notion_query_database, notion_get_page, notion_create_page, notion_update_page, notion_search | 7 |

### OAuth Token Storage

```sql
-- All tokens stored in PLAINTEXT
google_oauth_tokens.access_token   TEXT NOT NULL
google_oauth_tokens.refresh_token  TEXT NOT NULL
slack_oauth_tokens.bot_token       TEXT NOT NULL
slack_oauth_tokens.user_token      TEXT
notion_oauth_tokens.access_token   TEXT NOT NULL
```

---

## Security Analysis

### 1. Unauthorized Tool Exposure

**Current State:** PARTIAL PROTECTION

**What's Protected:**
- All MCP endpoints require valid session
- Session validated against PostgreSQL

**What's Missing:**

| Problem | Severity | Impact |
|---------|----------|--------|
| No per-tool authorization | HIGH | Any authenticated user can execute ANY tool |
| No integration connection check | HIGH | User can attempt tools for disconnected integrations |
| No tool allowlisting | MEDIUM | All 23 tools exposed to all users |
| No per-tool rate limiting | MEDIUM | Abuse of external API quotas |

**Current Flow:**
```go
func (h *Handlers) ExecuteMCPTool(c *gin.Context) {
    user := middleware.GetCurrentUser(c)  // Only checks: is user logged in?

    // MISSING: Does user have this integration connected?
    // MISSING: Is this tool enabled for this user?
    // MISSING: Per-tool rate limiting

    result, err := mcpService.ExecuteTool(ctx, req.Tool, req.Arguments)
}
```

**Recommended Fix:**
```go
func (h *Handlers) ExecuteMCPTool(c *gin.Context) {
    user := middleware.GetCurrentUser(c)

    // ADD: Pre-execution authorization
    if err := h.authorizeTool(user.ID, req.Tool); err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }

    // ADD: Per-tool rate limiting
    if !h.rateLimiter.AllowTool(user.ID, req.Tool) {
        c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
        return
    }

    result, err := mcpService.ExecuteTool(ctx, req.Tool, req.Arguments)
}

func (h *Handlers) authorizeTool(userID, toolName string) error {
    if IsSlackTool(toolName) {
        connected, _ := h.slackService.GetConnectionStatus(userID)
        if !connected {
            return errors.New("slack not connected")
        }
    }
    // Similar for Notion, Calendar...
    return nil
}
```

---

### 2. Session Hijacking

**Current State:** MODERATE PROTECTION

**What's Good:**
- HMAC-signed session tokens (`{token}.{signature}`)
- HttpOnly cookies (prevents XSS token theft)
- Session expiration in database
- Redis cache with 15-minute TTL
- Token hashing for Redis keys (HMAC-SHA256)

**What's Missing:**

| Problem | Severity | Impact |
|---------|----------|--------|
| No `Secure` flag enforcement | HIGH | Cookies transmit over HTTP |
| No `SameSite` attribute | HIGH | Vulnerable to CSRF attacks |
| Session not bound to IP/UA | MEDIUM | Stolen token works from anywhere |
| No session invalidation on password change | MEDIUM | Old sessions persist |
| No concurrent session limits | LOW | Unlimited active sessions |

**Current Cookie Handling:**
```go
// middleware/auth.go - NOT setting Secure or SameSite
sessionToken, err := c.Cookie("better-auth.session_token")
```

**Recommended Fix:**
```go
// When setting session cookies (in Better Auth config or custom handler)
c.SetSameSite(http.SameSiteStrictMode)
c.SetCookie(
    "better-auth.session_token",
    token,
    maxAge,
    "/",
    domain,
    true,   // Secure - HTTPS only
    true,   // HttpOnly - no JS access
)
```

**Additional Hardening:**
```go
type SessionMetadata struct {
    UserID      string
    Token       string
    IPAddress   string    // Bind to IP
    UserAgent   string    // Bind to browser fingerprint
    CreatedAt   time.Time
    LastUsedAt  time.Time
}

func ValidateSession(session SessionMetadata, req *http.Request) error {
    if session.IPAddress != getClientIP(req) {
        return errors.New("session IP mismatch")
    }
    if session.UserAgent != req.UserAgent() {
        return errors.New("session user agent mismatch")
    }
    return nil
}
```

---

### 3. Tool Shadowing / Shadow MCP

**Current State:** NOT ADDRESSED

**What is Tool Shadowing?**
Malicious tool definitions that impersonate legitimate tools to intercept data. An attacker could:
1. Inject fake tool definitions via compromised server
2. MITM the `/api/mcp/tools` response to add malicious tools
3. Modify tool behavior to exfiltrate data

**Current Tool Registration:**
```go
// services/mcp.go - Tools are hardcoded
var builtinTools = []Tool{
    {
        Name:        "search_conversations",
        Description: "Search through past conversations",
        // No signature, no hash, no version
    },
}
```

**What's Missing:**

| Problem | Severity | Impact |
|---------|----------|--------|
| No tool integrity validation | HIGH | Can't detect modified tools |
| No tool version tracking | MEDIUM | Can't detect if tool changed |
| No cryptographic signatures | MEDIUM | Can't verify tool authenticity |
| Frontend trusts backend blindly | MEDIUM | MITM could inject tools |

**Recommended Fix:**
```go
type SecureTool struct {
    Name        string
    Description string
    Version     string
    Hash        string    // SHA-256 of tool definition
    Signature   string    // Signed by server key
}

func (m *MCPService) GetAllTools() []SecureTool {
    tools := []SecureTool{}
    for _, t := range builtinTools {
        secure := SecureTool{
            Name:        t.Name,
            Description: t.Description,
            Version:     "1.0.0",
            Hash:        computeHash(t),
            Signature:   signTool(t, m.serverKey),
        }
        tools = append(tools, secure)
    }
    return tools
}
```

---

### 4. Sensitive Data Exposure & Token Theft

**Current State:** CRITICAL VULNERABILITY

**OAuth Tokens Stored in PLAINTEXT:**

```sql
-- google_oauth_tokens
access_token TEXT NOT NULL,     -- PLAINTEXT
refresh_token TEXT NOT NULL,    -- PLAINTEXT

-- slack_oauth_tokens
bot_token TEXT NOT NULL,        -- PLAINTEXT
user_token TEXT,                -- PLAINTEXT
incoming_webhook_url TEXT,      -- PLAINTEXT

-- notion_oauth_tokens
access_token TEXT NOT NULL,     -- PLAINTEXT
```

**Attack Vectors:**

| Vector | Likelihood | Impact |
|--------|------------|--------|
| SQL Injection | Medium | Full token theft for all users |
| Database backup exposure | High | All integration tokens compromised |
| Database admin access | Medium | Read all user tokens |
| Log file exposure | Medium | Tokens may appear in error logs |
| Memory dump | Low | Tokens in process memory |

**If Database Compromised, Attacker Gets:**
- Full Google Calendar access (read/write events)
- Full Slack workspace access (read messages, send as bot)
- Full Notion workspace access (read/write all pages)

**Recommended Fix - Envelope Encryption:**

```go
// services/encryption.go
type EncryptionService struct {
    masterKey []byte              // From env/KMS
    keyCache  map[string][]byte   // DEK cache
}

type EncryptedField struct {
    Ciphertext []byte `json:"ciphertext"`
    KeyID      string `json:"key_id"`
    Nonce      []byte `json:"nonce"`
    Algorithm  string `json:"algorithm"` // "AES-256-GCM"
}

func (e *EncryptionService) Encrypt(plaintext string) (*EncryptedField, error) {
    // Generate data encryption key (DEK)
    dek := generateRandomKey(32)
    keyID := uuid.New().String()

    // Encrypt plaintext with DEK
    block, _ := aes.NewCipher(dek)
    gcm, _ := cipher.NewGCM(block)
    nonce := generateRandomNonce(gcm.NonceSize())
    ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

    // Encrypt DEK with master key (envelope encryption)
    encryptedDEK := encryptWithMasterKey(dek, e.masterKey)

    // Store encrypted DEK separately (key table)
    storeEncryptedDEK(keyID, encryptedDEK)

    return &EncryptedField{
        Ciphertext: ciphertext,
        KeyID:      keyID,
        Nonce:      nonce,
        Algorithm:  "AES-256-GCM",
    }, nil
}

func (e *EncryptionService) Decrypt(field *EncryptedField) (string, error) {
    // Retrieve and decrypt DEK
    encryptedDEK := getEncryptedDEK(field.KeyID)
    dek := decryptWithMasterKey(encryptedDEK, e.masterKey)

    // Decrypt ciphertext with DEK
    block, _ := aes.NewCipher(dek)
    gcm, _ := cipher.NewGCM(block)
    plaintext, err := gcm.Open(nil, field.Nonce, field.Ciphertext, nil)

    return string(plaintext), err
}
```

**Database Schema Changes:**
```sql
-- Add encrypted columns
ALTER TABLE google_oauth_tokens
ADD COLUMN access_token_encrypted JSONB,
ADD COLUMN refresh_token_encrypted JSONB;

-- Add encryption key table
CREATE TABLE encryption_keys (
    id UUID PRIMARY KEY,
    encrypted_dek BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    rotated_at TIMESTAMP
);

-- Migration: encrypt existing tokens, then drop plaintext columns
```

---

### 5. Authentication Bypass

**Current State:** NEEDS REVIEW

**Auth Middleware Flow:**
```go
func AuthMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Get session cookie
        sessionToken, err := c.Cookie("better-auth.session_token")
        if err != nil {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        // 2. URL decode token
        decoded, _ := url.QueryUnescape(sessionToken)

        // 3. Split token from signature
        parts := strings.SplitN(decoded, ".", 2)
        if len(parts) != 2 {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        // 4. Query database for session
        row := pool.QueryRow(ctx, query, parts[0])
        // ...
    }
}
```

**Potential Issues:**

| Area | Risk | Details |
|------|------|---------|
| URL decode before split | LOW | Double-encoding could bypass validation |
| No rate limiting on auth | MEDIUM | Brute force session tokens |
| Token format validation | LOW | Only checks for "." separator |
| Signature verification | UNKNOWN | Need to verify HMAC check implementation |

**Recommended Hardening:**
```go
func AuthMiddleware(pool *pgxpool.Pool, limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()

        // ADD: Rate limit auth attempts
        if !limiter.AllowAuth(clientIP) {
            c.AbortWithStatus(http.StatusTooManyRequests)
            return
        }

        sessionToken, err := c.Cookie("better-auth.session_token")
        if err != nil {
            limiter.RecordFailure(clientIP)
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        // ADD: Validate token format before processing
        if !isValidTokenFormat(sessionToken) {
            limiter.RecordFailure(clientIP)
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        // ADD: Verify HMAC signature before DB lookup
        if !verifyTokenSignature(sessionToken, signingKey) {
            limiter.RecordFailure(clientIP)
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        // ... rest of validation
    }
}
```

---

## Priority Fix List

### CRITICAL - Fix Immediately

| # | Issue | Effort | Files Affected |
|---|-------|--------|----------------|
| 1 | Encrypt OAuth tokens at rest | High | `schema.sql`, all OAuth services, new encryption service |
| 2 | Add Secure + SameSite to cookies | Low | `middleware/auth.go`, Better Auth config |

### HIGH - Fix This Week

| # | Issue | Effort | Files Affected |
|---|-------|--------|----------------|
| 3 | Per-tool authorization | Medium | `handlers/mcp.go`, `services/mcp.go` |
| 4 | Audit logging for MCP | Medium | New `services/audit.go`, `handlers/mcp.go` |
| 5 | Token validation before use | Medium | All MCP service files |
| 6 | Rate limiting on auth middleware | Low | `middleware/auth.go` |

### MEDIUM - Plan for Next Sprint

| # | Issue | Effort | Files Affected |
|---|-------|--------|----------------|
| 7 | Session binding (IP/UA) | Medium | `middleware/auth.go`, session table |
| 8 | Tool integrity signatures | Medium | `services/mcp.go` |
| 9 | Per-tool rate limiting | Medium | New rate limiter, `handlers/mcp.go` |
| 10 | Concurrent session limits | Low | `middleware/auth.go` |

### LOW - Backlog

| # | Issue | Effort | Files Affected |
|---|-------|--------|----------------|
| 11 | Token rotation mechanism | High | All OAuth services, background job |
| 12 | Session invalidation on password change | Medium | Auth handlers |
| 13 | Tool version tracking | Low | `services/mcp.go` |

---

## Implementation Recommendations

### Phase 1: Critical Security (Week 1)

1. **Token Encryption**
   - Create `EncryptionService` with envelope encryption
   - Add encrypted columns to OAuth tables
   - Migrate existing tokens (encrypt in place)
   - Update all OAuth services to use encryption
   - Remove plaintext columns after verification

2. **Cookie Security**
   - Add `Secure` flag (HTTPS only)
   - Add `SameSite=Strict`
   - Verify in staging before production

### Phase 2: Authorization & Audit (Week 2)

3. **Per-Tool Authorization**
   - Add `authorizeTool()` check before execution
   - Verify integration connection status
   - Return clear error messages

4. **Audit Logging**
   - Create `mcp_audit_logs` table
   - Log all tool executions with:
     - User ID, Tool name, Arguments (sanitized)
     - Success/failure, Error message
     - Timestamp, IP address, Duration

### Phase 3: Hardening (Week 3-4)

5. **Session Hardening**
   - Bind sessions to IP + User-Agent hash
   - Add rate limiting to auth middleware
   - Implement concurrent session limits

6. **Tool Integrity**
   - Add version to tool definitions
   - Compute hash of tool schema
   - Consider signing for production

---

## Appendix: Key Files

| File | Purpose |
|------|---------|
| `internal/handlers/mcp.go` | MCP HTTP endpoints |
| `internal/services/mcp.go` | Core MCP logic, tool definitions |
| `internal/services/mcp_notion.go` | Notion integration tools |
| `internal/services/mcp_slack.go` | Slack integration tools |
| `internal/services/mcp_calendar.go` | Google Calendar tools |
| `internal/handlers/notion_oauth.go` | Notion OAuth flow |
| `internal/handlers/slack_oauth.go` | Slack OAuth flow |
| `internal/middleware/auth.go` | Session validation |
| `internal/middleware/redis_auth.go` | Redis session cache |
| `internal/database/schema.sql` | OAuth token tables |

---

*Document maintained by the BusinessOS Security Team*
