# Custom Agents - Production Deployment Checklist

**Feature**: Custom Agents System (v2.1.0)
**Target Environment**: GCP Cloud Run (Production)
**Deployment Date**: TBD
**Deployment Lead**: TBD

---

## Pre-Deployment Checklist

### 📋 Code Verification

- [ ] **All tests passing**
  ```bash
  # Backend
  cd desktop/backend-go && go test ./...

  # Frontend
  cd frontend && npm test
  ```
  - Expected: 100% pass rate
  - No flaky tests
  - Coverage > 70% for new code

- [ ] **Build succeeds**
  ```bash
  # Backend
  cd desktop/backend-go && go build -o bin/server ./cmd/server

  # Frontend
  cd frontend && npm run build
  ```
  - No compilation errors
  - No TypeScript errors
  - Build completes in < 5 minutes

- [ ] **Linting passes**
  ```bash
  # Backend
  cd desktop/backend-go && golangci-lint run

  # Frontend
  cd frontend && npm run lint
  ```
  - No critical linting errors
  - Warnings documented

---

### 🔐 Security Hardening

#### 1. Cookie Security Configuration

- [ ] **Update Cookie Settings for Production**

  **Files to modify**:
  - `desktop/backend-go/internal/handlers/auth_email.go`
  - `desktop/backend-go/internal/handlers/auth_google.go`

  **Required changes**:
  ```go
  // Add to .env or config
  ENVIRONMENT=production
  COOKIE_DOMAIN=.businessos.com  // Your production domain
  ALLOW_CROSS_ORIGIN=false       // Use SameSite=Lax if same domain
  ```

  **Update cookie code**:
  ```go
  isProduction := os.Getenv("ENVIRONMENT") == "production"
  domain := os.Getenv("COOKIE_DOMAIN")
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
      Secure:   isProduction,  // ✅ true in production
      SameSite: sameSite,       // ✅ Lax in production
  })
  ```

- [ ] **Verify HTTPS is enabled**
  - GCP Cloud Run enforces HTTPS automatically
  - Custom domain has valid SSL certificate
  - No HTTP-only endpoints

- [ ] **Verify Cookie Domain**
  ```bash
  # Test in production-like environment
  curl -v https://app.businessos.com/api/auth/login
  # Check Set-Cookie header has:
  # - Secure flag
  # - Domain=.businessos.com
  # - SameSite=Lax (or None if cross-origin)
  ```

#### 2. Input Validation

- [ ] **Add validation to CreateCustomAgent handler**

  **File**: `desktop/backend-go/internal/handlers/agents.go`

  **Add before SQLC call**:
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
  }

  // Validate welcome_message
  if len(req.WelcomeMessage) > 2000 {
      c.JSON(http.StatusBadRequest, gin.H{
          "error": "Welcome message cannot exceed 2000 characters"
      })
      return
  }

  // Validate category
  allowedCategories := map[string]bool{
      "general": true, "coding": true, "writing": true,
      "analysis": true, "research": true, "support": true,
      "sales": true, "marketing": true,
  }
  if req.Category != "" && !allowedCategories[req.Category] {
      c.JSON(http.StatusBadRequest, gin.H{
          "error": "Invalid category"
      })
      return
  }
  ```

- [ ] **Add same validation to UpdateCustomAgent handler**

- [ ] **Fix temperature edge case**
  ```go
  // Change from:
  if req.Temperature > 0 {

  // To:
  if req.Temperature >= 0 && req.Temperature <= 2.0 {
  ```

#### 3. Rate Limiting

- [ ] **Add rate limiting to agent creation**

  **Option A: Per-user limit** (Recommended)
  ```go
  // Add SQLC query
  -- name: CountUserAgents :one
  SELECT COUNT(*) FROM custom_agents WHERE user_id = $1;

  // In CreateCustomAgent handler
  count, err := queries.CountUserAgents(ctx, user.ID)
  if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check agent count"})
      return
  }
  if count >= 100 {
      c.JSON(http.StatusBadRequest, gin.H{
          "error": "Maximum 100 custom agents allowed per user"
      })
      return
  }
  ```

  **Option B: Time-based rate limit**
  ```go
  // In route registration
  r.POST("/api/ai/custom-agents",
      middleware.RateLimiter(10, time.Minute),
      h.CreateCustomAgent,
  )
  ```

- [ ] **Test rate limiting**
  ```bash
  # Try creating 101 agents
  # Should fail on 101st
  ```

---

### 🗄️ Database Migration

#### 1. Backup Current Database

- [ ] **Create database backup BEFORE migration**
  ```bash
  # For Supabase
  # Use Supabase Dashboard → Database → Backups → Create Backup

  # Or use pg_dump
  pg_dump -h db.xxx.supabase.co -U postgres -d postgres > backup_pre_migration_043.sql
  ```

- [ ] **Verify backup is downloadable and complete**
  ```bash
  # Check file size > 0
  ls -lh backup_pre_migration_043.sql
  ```

#### 2. Run Migration

- [ ] **Test migration on staging first**
  ```bash
  # Point to staging database
  export DATABASE_URL="postgresql://staging..."

  cd desktop/backend-go
  go run ./cmd/migrate
  ```

- [ ] **Verify migration output**
  ```
  Expected output:
  ✓ custom_agents.welcome_message column OK
  ✓ custom_agents.suggested_prompts column OK
  ✓ custom_agents.is_featured column OK
  ✓ idx_custom_agents_featured index OK
  Migration complete!
  ```

- [ ] **Run migration on production**
  ```bash
  # Point to production database
  export DATABASE_URL="postgresql://production..."

  cd desktop/backend-go
  go run ./cmd/migrate
  ```

#### 3. Migration Rollback Strategy

If migration fails or causes issues:

- [ ] **Document rollback SQL**
  ```sql
  -- Save this as rollback_migration_043.sql

  BEGIN;

  -- Drop index
  DROP INDEX IF EXISTS idx_custom_agents_featured;

  -- Drop columns (⚠️ DATA LOSS!)
  ALTER TABLE custom_agents
  DROP COLUMN IF EXISTS welcome_message,
  DROP COLUMN IF EXISTS suggested_prompts,
  DROP COLUMN IF EXISTS is_featured;

  COMMIT;
  ```

- [ ] **Test rollback on staging**
  ```bash
  psql $STAGING_DATABASE_URL < rollback_migration_043.sql
  ```

- [ ] **Keep rollback script accessible during deployment**
  - Store in Google Drive or secure location
  - Have database credentials ready

---

### 🚀 Backend Deployment

#### 1. Environment Variables

- [ ] **Verify all required env vars in GCP Cloud Run**
  ```bash
  # Check these are set:
  ENVIRONMENT=production
  COOKIE_DOMAIN=.businessos.com
  ALLOW_CROSS_ORIGIN=false  # or true if cross-origin needed

  # Existing vars (verify still present)
  DATABASE_URL=postgresql://...
  REDIS_URL=redis://...
  AI_PROVIDER=anthropic
  ANTHROPIC_API_KEY=...
  GROQ_API_KEY=...
  DEFAULT_MODEL=llama-3.3-70b-versatile
  SECRET_KEY=...
  TOKEN_ENCRYPTION_KEY=...
  ```

- [ ] **Verify secret keys are secure**
  - `SECRET_KEY` is at least 32 characters
  - `TOKEN_ENCRYPTION_KEY` is exactly 32 characters (AES-256)
  - Keys are NOT hard-coded in code

#### 2. Build & Deploy

- [ ] **Build Docker image**
  ```bash
  cd desktop/backend-go

  docker build -t gcr.io/your-project/businessos-backend:v2.1.0 .
  docker build -t gcr.io/your-project/businessos-backend:latest .
  ```

- [ ] **Push to GCR**
  ```bash
  docker push gcr.io/your-project/businessos-backend:v2.1.0
  docker push gcr.io/your-project/businessos-backend:latest
  ```

- [ ] **Deploy to Cloud Run**
  ```bash
  gcloud run deploy businessos-backend \
    --image gcr.io/your-project/businessos-backend:v2.1.0 \
    --region us-central1 \
    --platform managed \
    --allow-unauthenticated \
    --set-env-vars ENVIRONMENT=production,COOKIE_DOMAIN=.businessos.com
  ```

- [ ] **Verify deployment health**
  ```bash
  # Check health endpoint
  curl https://backend.businessos.com/health

  # Expected: {"status": "ok", "database": "connected"}
  ```

---

### 💻 Frontend Deployment

#### 1. Build Configuration

- [ ] **Update frontend .env for production**
  ```bash
  # frontend/.env.production
  VITE_API_URL=https://backend.businessos.com
  VITE_ENVIRONMENT=production
  ```

- [ ] **Build frontend**
  ```bash
  cd frontend
  npm run build
  ```

- [ ] **Verify build output**
  ```bash
  # Check dist/ folder exists and has content
  ls -lh dist/
  # Should see index.html, assets/, etc.
  ```

#### 2. Deploy Frontend

- [ ] **Deploy to hosting (Vercel/Netlify/GCP Storage)**
  ```bash
  # Example for Vercel
  vercel --prod

  # Example for GCP Storage + Cloud CDN
  gsutil -m rsync -r dist/ gs://businessos-frontend-prod/
  ```

- [ ] **Verify frontend loads**
  ```bash
  curl https://app.businessos.com
  # Should return HTML with no errors
  ```

---

### 🧪 Post-Deployment Testing

#### 1. Smoke Tests

- [ ] **Test agent creation**
  1. Navigate to /settings/ai
  2. Click "Create Custom Agent"
  3. Fill all fields including:
     - Welcome message (test: "Welcome to my agent!")
     - Suggested prompts (test: ["Hello", "Help me", "What can you do?"])
     - Category (test: "coding")
     - Toggle "Featured" and "Public"
  4. Click "Create"
  5. ✅ Agent appears in list

- [ ] **Test agent update**
  1. Click on created agent
  2. Edit welcome message
  3. Add/remove suggested prompts
  4. Save
  5. ✅ Changes persist after refresh

- [ ] **Test agent testing (sandbox)**
  1. Click "Test" on agent
  2. Send message: "Hello"
  3. ✅ Response streams in real-time (SSE)
  4. ✅ No console errors

- [ ] **Test agent deletion**
  1. Delete test agent
  2. ✅ Agent removed from list
  3. ✅ Database record deleted

#### 2. Edge Case Tests

- [ ] **Test with maximum values**
  - Welcome message: 2000 characters
  - Suggested prompts: 10 prompts
  - Each prompt: 500 characters
  - ✅ All save successfully

- [ ] **Test with over-limit values**
  - Welcome message: 2001 characters → ✅ Should fail with error
  - Suggested prompts: 11 prompts → ✅ Should fail with error

- [ ] **Test temperature edge case**
  - Set temperature to 0.0
  - ✅ Should save as 0.0 (not ignored)

#### 3. Security Tests

- [ ] **Test cookie security**
  1. Open browser DevTools → Application → Cookies
  2. Find `better-auth.session_token`
  3. ✅ Verify flags:
     - `Secure`: ✅
     - `HttpOnly`: ✅
     - `SameSite`: `Lax` or `None` (depending on config)
     - `Domain`: `.businessos.com`

- [ ] **Test authentication**
  1. Log out
  2. Try to create agent without login
  3. ✅ Should get 401 Unauthorized

- [ ] **Test authorization**
  1. Create agent as User A
  2. Log in as User B
  3. Try to edit User A's agent
  4. ✅ Should fail (user_id check prevents it)

#### 4. Performance Tests

- [ ] **Test list performance**
  ```bash
  # Create 50 agents
  # Measure list load time
  # Should be < 500ms
  ```

- [ ] **Test featured agents query**
  ```bash
  # Create 10 featured public agents
  # Query with filter is_featured=true&is_public=true
  # Should use partial index (fast)
  ```

---

### 📊 Monitoring Setup

#### 1. Logging

- [ ] **Verify structured logging works**
  ```bash
  # Check Cloud Run logs
  gcloud logging read "resource.type=cloud_run_revision" --limit 50

  # Should see logs like:
  # [CreateCustomAgent] User xxx created agent "my-agent"
  # [UpdateCustomAgent] Agent xxx updated by user yyy
  ```

- [ ] **Set up log-based alerts**
  - Alert on: `[ERROR]` in logs
  - Alert on: HTTP 500 responses > 1% of traffic
  - Alert on: Database connection failures

#### 2. Metrics

- [ ] **Track custom agent metrics**
  - Total agents created (counter)
  - Active agents per user (histogram)
  - Agent test requests (counter)
  - SSE stream duration (histogram)

- [ ] **Set up dashboards**
  - GCP Cloud Run dashboard
  - Custom Agents usage dashboard
  - Error rate dashboard

#### 3. Alerting

- [ ] **Configure alerts**
  - Error rate > 5% → page on-call
  - API latency > 2s → warn
  - Database CPU > 80% → warn
  - Memory usage > 90% → page

---

### 📚 Documentation

- [ ] **Update API documentation**
  - Document new fields in OpenAPI spec
  - Update Postman collection
  - Add examples for suggested_prompts, welcome_message

- [ ] **Update user documentation**
  - Add guide: "Creating Custom Agents"
  - Add guide: "Featured Agents"
  - Update FAQ

- [ ] **Update team runbook**
  - Add migration 043 rollback procedure
  - Document troubleshooting steps
  - Add monitoring dashboard links

---

### 🔄 Rollback Plan

If deployment causes critical issues:

#### Immediate Rollback (< 5 minutes)

1. **Rollback backend**
   ```bash
   # Deploy previous version
   gcloud run deploy businessos-backend \
     --image gcr.io/your-project/businessos-backend:v2.0.0 \
     --region us-central1
   ```

2. **Rollback frontend**
   ```bash
   # Redeploy previous version
   vercel rollback

   # Or for GCP
   gsutil -m rsync -r dist_backup/ gs://businessos-frontend-prod/
   ```

3. **Rollback database (if needed)**
   ```bash
   # Run rollback script
   psql $DATABASE_URL < rollback_migration_043.sql

   # Or restore from backup
   psql $DATABASE_URL < backup_pre_migration_043.sql
   ```

#### Post-Rollback

- [ ] **Investigate issue**
  - Check logs in GCP Console
  - Check error reports
  - Identify root cause

- [ ] **Fix in staging**
  - Reproduce issue in staging
  - Apply fix
  - Re-test thoroughly

- [ ] **Schedule re-deployment**
  - Communicate new deployment time
  - Run through checklist again

---

### ✅ Final Sign-Off

Before marking deployment as complete:

- [ ] **All smoke tests pass**
- [ ] **No critical errors in logs (first 30 minutes)**
- [ ] **Metrics look normal (latency, error rate)**
- [ ] **Team notified of successful deployment**
- [ ] **Deployment notes saved to project wiki**

**Deployed By**: _________________

**Deployment Date**: _________________

**Deployment Time**: _________________

**Rollback Plan Tested**: ☐ Yes ☐ No

**Sign-Off**: _________________

---

## Appendix A: Environment Variable Reference

### Required for Production

| Variable | Example | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `production` | Enables production security features |
| `COOKIE_DOMAIN` | `.businessos.com` | Cookie domain for auth |
| `DATABASE_URL` | `postgresql://...` | PostgreSQL connection string |
| `REDIS_URL` | `redis://...` | Redis connection string |
| `AI_PROVIDER` | `anthropic` | AI provider (anthropic/groq/ollama) |
| `ANTHROPIC_API_KEY` | `sk-ant-...` | Anthropic API key |
| `GROQ_API_KEY` | `gsk_...` | Groq API key |
| `DEFAULT_MODEL` | `llama-3.3-70b-versatile` | Default LLM model |
| `SECRET_KEY` | `<random-32-chars>` | JWT signing key |
| `TOKEN_ENCRYPTION_KEY` | `<random-32-chars>` | AES-256 encryption key |

### Optional

| Variable | Example | Description |
|----------|---------|-------------|
| `ALLOW_CROSS_ORIGIN` | `false` | Enable SameSite=None (cross-origin cookies) |
| `LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |
| `MAX_AGENTS_PER_USER` | `100` | Limit agents per user |

---

## Appendix B: Rollback SQL Script

**File**: `rollback_migration_043.sql`

```sql
-- ⚠️ WARNING: This will DELETE data in new columns!
-- Only run if absolutely necessary

BEGIN;

-- Verify we're in the right database
SELECT current_database();

-- Drop index first (can't drop columns with index)
DROP INDEX IF EXISTS idx_custom_agents_featured;

-- Drop new columns (DATA LOSS!)
ALTER TABLE custom_agents
DROP COLUMN IF EXISTS welcome_message,
DROP COLUMN IF EXISTS suggested_prompts,
DROP COLUMN IF EXISTS is_featured;

-- Verify columns are gone
SELECT column_name
FROM information_schema.columns
WHERE table_name = 'custom_agents'
  AND column_name IN ('welcome_message', 'suggested_prompts', 'is_featured');

-- Should return 0 rows

COMMIT;
```

---

## Appendix C: Monitoring Queries

### Check Custom Agents Usage

```sql
-- Total agents created
SELECT COUNT(*) AS total_agents FROM custom_agents;

-- Agents per user (top 10)
SELECT user_id, COUNT(*) AS agent_count
FROM custom_agents
GROUP BY user_id
ORDER BY agent_count DESC
LIMIT 10;

-- Featured public agents
SELECT COUNT(*) AS featured_count
FROM custom_agents
WHERE is_featured = TRUE AND is_public = TRUE;

-- Agents with suggested prompts
SELECT COUNT(*) AS agents_with_prompts
FROM custom_agents
WHERE suggested_prompts IS NOT NULL
  AND array_length(suggested_prompts, 1) > 0;

-- Average suggested prompts per agent
SELECT AVG(array_length(suggested_prompts, 1)) AS avg_prompts
FROM custom_agents
WHERE suggested_prompts IS NOT NULL;
```

### Check for Issues

```sql
-- Agents with overly long welcome messages (shouldn't exist if validation works)
SELECT id, user_id, LENGTH(welcome_message) AS msg_length
FROM custom_agents
WHERE LENGTH(welcome_message) > 2000;

-- Agents with too many suggested prompts (shouldn't exist)
SELECT id, user_id, array_length(suggested_prompts, 1) AS prompt_count
FROM custom_agents
WHERE array_length(suggested_prompts, 1) > 10;
```

---

**Document Version**: 1.0
**Last Updated**: 2026-01-11
**Maintained By**: DevOps Team
