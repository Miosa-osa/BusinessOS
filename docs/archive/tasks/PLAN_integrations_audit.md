# BusinessOS Integrations & Sorx Engine - Comprehensive Audit & Plan

## Executive Summary

**Audit Date:** 2026-01-05
**Branch:** main-dev
**Total Changes:** 25 files modified, 10+ new files, ~10,500 lines added

### Critical Finding: NOT READY FOR PRODUCTION

The integrations and Sorx engine implementation has a **solid foundation** but contains:
- **15 CRITICAL security vulnerabilities**
- **18 HIGH-severity security issues**
- **Major architectural gaps** (45% of documented vision implemented)
- **All integration actions are mock/placeholder implementations**

---

## Part 1: What Was Built

### 1.1 Backend Changes (Go)

| Component | Lines | Status |
|-----------|-------|--------|
| Sorx Engine (`internal/sorx/`) | 2,634 | Core working, actions mocked |
| Integration Handlers | 1,593 | API complete, logic partial |
| Sorx Service | 703 | Credential system working |
| Database Migrations | 639 | 9 new tables |
| SQLC Queries | 274 | Complete |

**New Files Created:**
```
desktop/backend-go/internal/sorx/
├── engine.go           (755 lines) - Core execution engine
├── types.go            (422 lines) - Type definitions
├── actions.go          (436 lines) - Action handlers (MOCKED)
├── agent_bridge.go     (299 lines) - BusinessOS agent integration
├── command_skills.go   (572 lines) - Command-based skills
└── skill_commands.go   (150 lines) - Skill command definitions

desktop/backend-go/internal/handlers/
├── integrations.go     (1,095 lines) - Integration management
└── sorx.go             (498 lines) - Sorx HTTP handlers

desktop/backend-go/internal/services/
└── sorx.go             (703 lines) - Credential & callback handling

desktop/backend-go/internal/database/migrations/
├── 025_integrations_module.sql (365 lines) - 9 tables
└── 026_oauth_tokens_tables.sql - OAuth token schema
```

**44 New API Routes Added:**
- `/api/sorx/*` - 22 routes for skill execution
- `/api/integrations/*` - 22 routes for provider management

### 1.2 Frontend Changes (Svelte)

| Component | Size | Status |
|-----------|------|--------|
| API Types | 147 lines | Complete |
| API Client | 206 lines | Complete |
| Components | ~9KB | Basic UI |

**New Files:**
```
frontend/src/lib/api/integrations/
├── index.ts
├── integrations.ts     (206 lines)
└── types.ts            (147 lines)

frontend/src/lib/components/integrations/
├── index.ts
└── ModuleIntegrations.svelte (8,821 bytes)

frontend/src/routes/(app)/integrations/  (new route)
```

### 1.3 Database Schema (9 New Tables)

```sql
1. integration_providers    - Catalog of 40+ providers
2. user_integrations        - User's connected integrations
3. module_integration_settings - Per-module customization
4. user_model_preferences   - AI model tier selection
5. pending_decisions        - Human-in-the-loop queue
6. integration_sync_log     - Sync operation tracking
7. skill_executions         - Sorx skill run tracking
8. google_oauth_tokens      - Google OAuth (updated)
9. slack_oauth_tokens       - Slack OAuth (new)
10. notion_oauth_tokens     - Notion OAuth (new)
```

---

## Part 2: Security Audit Results

### CRITICAL VULNERABILITIES (Must Fix Before ANY Deployment)

| # | Issue | Location | Risk |
|---|-------|----------|------|
| 1 | **Plaintext OAuth tokens in database** | `026_oauth_tokens_tables.sql` | Full account takeover in DB breach |
| 2 | **"Encrypted" columns not actually encrypted** | `025_integrations_module.sql:45-46` | False security, data exposed |
| 3 | **Insecure OAuth state cookies** | `google_oauth.go:38-39` | CSRF attacks via MitM |
| 4 | **State generation ignores crypto errors** | `google_oauth.go:160-163` | Predictable tokens |
| 5 | **HMAC secret auto-generation** | `redis_auth.go:71-79` | Breaks horizontal scaling |
| 6 | **Weak default SECRET_KEY** | `config.go:121` | JWT signature compromise |
| 7 | **Empty HMAC secret default** | `config.go:152` | Session security broken |
| 8 | **Access token in URL (not header)** | `google_oauth.go:166` | Token logged everywhere |
| 9 | **No provider ID input validation** | `integrations.go:243-288` | SQL injection risk |
| 10 | **Same issues in Slack OAuth** | `slack_oauth.go` | Duplicate vulnerabilities |
| 11 | **Same issues in Notion OAuth** | `notion_oauth.go` | Duplicate vulnerabilities |

### HIGH-SEVERITY ISSUES

| # | Issue | Location |
|---|-------|----------|
| 1 | Missing scope validation in OAuth | `slack_oauth.go:68-82` |
| 2 | Race condition in token storage | `google_oauth.go:92-98` |
| 3 | No rate limiting on OAuth endpoints | All OAuth handlers |
| 4 | Missing security headers | No middleware |
| 5 | CSRF tokens missing in API | `integrations.ts` |
| 6 | OAuth URL not validated in frontend | `integrations.ts:231-234` |
| 7 | Session cache TTL mismatch | `redis_auth.go:54` |
| 8 | Hardcoded provider mapping | `integrations.go:146-178` |

---

## Part 3: Architecture Gap Analysis

### Sorx 2.0: Documented Vision vs Implementation

| Feature | Documented | Implemented | Gap |
|---------|------------|-------------|-----|
| Core Execution Engine | Yes | Yes | None |
| Tier 1 Skills (Deterministic) | Yes | Working | None |
| Tier 2 Skills (Haiku AI) | Yes | **None** | 100% |
| Tier 3 Skills (Sonnet AI) | Yes | ~9 skills | 80% |
| Tier 4 Skills (Opus AI) | Yes | **None** | 100% |
| Skill Auto-Splitting | Detailed spec | **Not built** | 100% |
| Skill Health Monitoring | Designed | **Not built** | 100% |
| Objective Database | Core concept | **Not built** | 100% |
| 50+ Skill Catalog | Documented | ~9 built | 82% |
| Temperature Control | Defined | Types only | 80% |
| Human-in-the-Loop | Designed | Working | None |

**Implementation Status: ~45% of documented vision**

### Integration Auth Patterns - Current State

| Provider | Auth Type | Status | Issues |
|----------|-----------|--------|--------|
| Google (Login) | OAuth 2.0 | Working | Insecure cookies |
| Google (Calendar/Drive) | OAuth 2.0 | Partial | Dual redirect needed |
| Slack | OAuth 2.0 | Working | Scope validation missing |
| Notion | OAuth 2.0 | Working | Same cookie issues |
| HubSpot | OAuth 2.0 | **Not implemented** | - |
| Linear | API Key | **Not implemented** | - |
| OpenAI/Claude | API Key | **Not implemented** | - |
| Other 30+ providers | Various | **Not implemented** | - |

### What's Actually Working vs Mocked

**Working:**
- OAuth flow for Google/Slack/Notion
- Credential ticket system (encrypt/decrypt)
- Human-in-the-loop decision flow
- Skill execution engine (core loop)
- Event-driven callbacks
- Frontend API layer

**Mocked/Placeholder:**
- ALL integration actions (Gmail, Calendar, HubSpot, Linear, Slack, Notion)
- AI actions (extract, summarize, classify)
- No actual API calls to external services
- BusinessOS data handlers return empty results

---

## Part 4: Action Plan

### Phase 0: Pre-Deployment Security Fixes (BLOCKING)

**Must complete before ANY code goes to production:**

1. **Implement Token Encryption**
   ```go
   // Create: internal/security/token_encryption.go
   type TokenEncryption struct {
       cipher cipher.AEAD
   }
   func (te *TokenEncryption) Encrypt(plaintext string) ([]byte, error)
   func (te *TokenEncryption) Decrypt(ciphertext []byte) (string, error)
   ```

2. **Fix OAuth Cookie Security**
   ```go
   // All OAuth handlers - set Secure=true
   c.SetCookie("oauth_state", state, 600, "/", "", true, true)
   //                                              ^^^^ Secure=true
   ```

3. **Validate Production Secrets**
   ```go
   // config.go - Add validation
   func (c *Config) Validate() error {
       if c.IsProduction() {
           if c.SecretKey == "your-secret-key-change-this-in-production" {
               return errors.New("SECRET_KEY must be changed in production")
           }
           if c.RedisKeyHMACSecret == "" {
               return errors.New("REDIS_KEY_HMAC_SECRET required in production")
           }
       }
       return nil
   }
   ```

4. **Fix State Generation**
   ```go
   func generateRandomState() (string, error) {
       b := make([]byte, 32)
       if _, err := rand.Read(b); err != nil {
           return "", fmt.Errorf("crypto/rand failed: %w", err)
       }
       return base64.URLEncoding.EncodeToString(b), nil
   }
   ```

5. **Add Input Validation**
   ```go
   // Provider ID whitelist
   var validProviders = map[string]bool{"gmail": true, "slack": true, ...}
   ```

### Phase 1: Sorx Engine Repository Extraction

**Create separate `sorx-engine` repository:**

```
sorx-engine/
├── cmd/
│   └── sorx/main.go
├── internal/
│   ├── engine/
│   │   ├── engine.go
│   │   ├── types.go
│   │   └── actions.go
│   ├── skills/
│   │   ├── registry.go
│   │   └── builtin/
│   ├── credentials/
│   │   └── manager.go
│   └── decisions/
│       └── handler.go
├── pkg/
│   └── sorx/
│       └── client.go      (SDK for BusinessOS)
├── api/
│   └── proto/             (gRPC definitions)
├── docs/
│   └── SORX_2.0_SPECIFICATION.md
├── go.mod
└── README.md
```

**Files to Extract from BusinessOS:**
- `internal/sorx/*` → `sorx-engine/internal/engine/`
- `internal/handlers/sorx.go` → Keep in BusinessOS (calls Sorx SDK)
- `internal/services/sorx.go` → Split between both repos
- `docs/sorxdocs/*` → `sorx-engine/docs/`

### Phase 2: Integration Auth Architecture Fix

**Unified OAuth Handler Pattern:**

```go
// internal/integrations/oauth/handler.go
type UnifiedOAuthHandler struct {
    config       *OAuthConfig
    encryptor    security.TokenEncryption
    stateManager security.StateManager
    providers    map[string]OAuthProvider
}

// Per-provider config
type OAuthProvider interface {
    GetAuthURL(state string) string
    ExchangeCode(code string) (*TokenResponse, error)
    RefreshToken(refreshToken string) (*TokenResponse, error)
    GetScopes() []string
}
```

**Auth Type Registry:**

| Provider | Auth Type | Implementation |
|----------|-----------|----------------|
| Google | OAuth 2.0 | `UnifiedOAuthHandler` |
| Slack | OAuth 2.0 | `UnifiedOAuthHandler` |
| Notion | OAuth 2.0 | `UnifiedOAuthHandler` |
| HubSpot | OAuth 2.0 | `UnifiedOAuthHandler` |
| Linear | API Key | `APIKeyHandler` |
| OpenAI | API Key | `APIKeyHandler` |
| Anthropic | API Key | `APIKeyHandler` |
| Custom | Varies | `CustomAuthHandler` |

### Phase 3: Push to Integrations Branch

**Steps:**
1. Complete Phase 0 security fixes
2. Run all tests
3. Create `integrations` branch from `main-dev`
4. Push changes
5. Create PR to `main` with security review

---

## Part 5: What's Bullshit / Needs Cleanup

### Code Issues Found

1. **Hardcoded 46-provider fallback** in `integrations.go:181-239`
   - Should fail if database empty, not load hardcoded data
   - Creates maintenance nightmare

2. **Duplicate OAuth logic** across 3 files
   - `google_oauth.go`, `slack_oauth.go`, `notion_oauth.go`
   - Same vulnerabilities repeated 3 times

3. **Binary files in repo**
   - `main`, `server`, `businessos-backend` binaries checked in
   - Should be in `.gitignore`

4. **Mock implementations pretending to be real**
   - `actions.go` has handlers like `handleGmailAction()` that return empty data
   - No indication these are stubs
   - Could cause confusion

5. **"Encrypted" columns that aren't encrypted**
   - `access_token_encrypted BYTEA` - misleading name
   - No encryption code exists

6. **Incomplete frontend integration page**
   - Route exists but functionality minimal
   - No error handling for failed OAuth

### Documentation Issues

1. **SORX_2.0_SPECIFICATION.md** (388KB) describes features that don't exist
   - Skill auto-splitting: not implemented
   - Objective database: not implemented
   - 50+ skills: only 9 exist

2. **No API documentation**
   - 44 new endpoints with no OpenAPI spec
   - No request/response examples

3. **No security documentation**
   - Token encryption not documented
   - Credential flow not documented

---

## Part 6: Recommended Next Steps

### Immediate (This Session)

1. [ ] Fix critical security issues (Phase 0)
2. [ ] Remove binary files from repo
3. [ ] Add `.gitignore` entries
4. [ ] Create `integrations` branch
5. [ ] Push with security fixes

### Short-Term (Next Sprint)

1. [ ] Extract Sorx engine to separate repo
2. [ ] Implement unified OAuth handler
3. [ ] Add real integration actions (start with Gmail, Calendar)
4. [ ] Add rate limiting
5. [ ] Add security headers middleware

### Medium-Term (Following Sprints)

1. [ ] Implement remaining OAuth providers
2. [ ] Add API key auth pattern
3. [ ] Implement skill health monitoring
4. [ ] Build Tier 2/4 skills
5. [ ] Add OpenAPI documentation

---

## Appendix: Files Modified Summary

### Modified Files (23)
```
desktop/backend-go/internal/config/config.go
desktop/backend-go/internal/handlers/auth_google.go
desktop/backend-go/internal/handlers/delegation.go
desktop/backend-go/internal/handlers/google_oauth.go
desktop/backend-go/internal/handlers/handlers.go
desktop/backend-go/internal/handlers/notion_oauth.go
desktop/backend-go/internal/handlers/slack_oauth.go
desktop/backend-go/internal/handlers/sync.go
desktop/backend-go/internal/handlers/thinking.go
desktop/backend-go/internal/middleware/auth.go
desktop/backend-go/internal/middleware/redis_auth.go
desktop/backend-go/internal/services/google_calendar.go
desktop/backend-go/internal/services/notion.go
desktop/backend-go/internal/services/slack.go
frontend/src/lib/api/integrations/index.ts
frontend/src/lib/api/integrations/integrations.ts
frontend/src/lib/api/integrations/types.ts
frontend/src/lib/auth-client.ts
frontend/src/lib/stores/windowStore.ts
frontend/src/routes/(app)/+layout.svelte
frontend/src/routes/window/+page.svelte
```

### New Files (10+)
```
desktop/backend-go/internal/database/migrations/025_integrations_module.sql
desktop/backend-go/internal/database/migrations/026_oauth_tokens_tables.sql
desktop/backend-go/internal/database/queries/integrations.sql
desktop/backend-go/internal/handlers/integrations.go
desktop/backend-go/internal/handlers/sorx.go
desktop/backend-go/internal/integrations/ (directory)
desktop/backend-go/internal/sorx/ (6 files)
desktop/backend-go/internal/services/sorx.go
frontend/src/lib/components/integrations/ (2 files)
frontend/src/routes/(app)/integrations/ (directory)
```

---

**Plan Created:** 2026-01-05
**Author:** @architect + @security-auditor + @explorer
**Status:** Awaiting Approval
