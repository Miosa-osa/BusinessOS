# Onboarding System - Quick Reference

**Fast lookup for common tasks and troubleshooting**

---

## API Endpoints

### Authentication
```
GET  /api/v1/auth/google/login
GET  /api/v1/auth/google/callback
GET  /api/v1/auth/session
POST /api/v1/auth/logout
```

### Onboarding
```
GET  /api/osa-onboarding/user-analysis/:user_id
POST /api/v1/osa-onboarding/analyze
GET  /api/v1/osa-onboarding/analyze/:analysis_id
GET  /api/v1/osa-onboarding/analyze/:analysis_id/stream
POST /api/v1/osa-onboarding/generate-apps
```

### Users
```
GET   /api/v1/users/username/check/:username
PATCH /api/v1/users/username
```

---

## Frontend Routes

```
/onboarding              → Screen 1: Welcome
/onboarding/meet-osa     → Screen 2: Meet OSA
/onboarding/signin       → Screen 3: Sign In
/onboarding/gmail        → Screen 4: Gmail Connection
/onboarding/username     → Screen 5: Username
/onboarding/analyzing    → Screen 6: Analysis 1
/onboarding/analyzing-2  → Screen 7: Analysis 2
/onboarding/analyzing-3  → Screen 8: Analysis 3
/onboarding/starter-apps → Screen 9: Starter Apps
/onboarding/ready        → Screen 10: Ready
→ /dashboard             → Main app
```

---

## Database Quick Queries

### Check User Onboarding Status
```sql
SELECT
    id,
    email,
    username,
    onboarding_completed
FROM "user"
WHERE email = 'user@example.com';
```

### Check Analysis Status
```sql
SELECT
    user_id,
    status,
    insights,
    tools_used,
    total_emails_analyzed,
    error_message,
    completed_at
FROM onboarding_user_analysis
WHERE user_id = 'usr_xyz123';
```

### Check Gmail Token
```sql
SELECT
    user_id,
    provider_id,
    status,
    scopes,
    token_expires_at
FROM user_integrations
WHERE user_id = 'usr_xyz123'
  AND provider_id = 'google_gmail';
```

### Check Starter Apps
```sql
SELECT
    title,
    description,
    reasoning,
    display_order,
    status
FROM onboarding_starter_apps
WHERE user_id = 'usr_xyz123'
ORDER BY display_order;
```

### Reset User Onboarding
```sql
-- WARNING: This deletes all onboarding data
DELETE FROM onboarding_starter_apps WHERE user_id = 'usr_xyz123';
DELETE FROM onboarding_user_analysis WHERE user_id = 'usr_xyz123';
UPDATE "user" SET onboarding_completed = false WHERE id = 'usr_xyz123';
```

---

## Common Debugging Tasks

### 1. Check If Analysis Started
```bash
# Backend logs
tail -f logs/app.log | grep "Gmail Analysis"

# Database
psql -d businessos -c "SELECT status, created_at FROM onboarding_user_analysis WHERE user_id = 'usr_xyz123';"
```

### 2. Manually Trigger Analysis
```bash
curl -X POST http://localhost:8001/api/v1/osa-onboarding/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "usr_xyz123",
    "workspace_id": "00000000-0000-0000-0000-000000000000",
    "max_emails": 50
  }'
```

### 3. Check Gmail API Access
```bash
# In Go code (temporary debug):
// Add this to auth_google.go after storing tokens:
log.Printf("🔍 Testing Gmail access for user: %s", userID)
gmailService := integrations.NewGmailService(googleProvider)
emails, err := gmailService.GetEmails(ctx, userID, "INBOX", 10, 0)
log.Printf("✅ Fetched %d emails, err: %v", len(emails), err)
```

### 4. View Analysis Response
```sql
SELECT
    insights::text,
    interests::text,
    tools_used::text,
    profile_summary
FROM onboarding_user_analysis
WHERE user_id = 'usr_xyz123';
```

### 5. Check Groq API Call
```bash
# Backend logs
grep "ProfileAnalyzerAgent" logs/app.log

# Look for:
# "ProfileAnalyzerAgent starting analysis"
# "ProfileAnalyzerAgent analysis complete"
# "ProfileAnalyzerAgent LLM call failed" (if error)
```

---

## Configuration Checklist

### Local Development
- [ ] `GOOGLE_CLIENT_ID` set
- [ ] `GOOGLE_CLIENT_SECRET` set
- [ ] `GOOGLE_REDIRECT_URI=http://localhost:8001/api/v1/auth/google/callback`
- [ ] `GROQ_API_KEY` set
- [ ] `DATABASE_URL` points to local DB
- [ ] Gmail API enabled in Google Console
- [ ] Test user added to OAuth consent screen

### Production
- [ ] Production Google OAuth client created
- [ ] Redirect URI uses HTTPS
- [ ] `TOKEN_ENCRYPTION_KEY` set
- [ ] `ENVIRONMENT=production`
- [ ] `SECRET_KEY` rotated
- [ ] Secure cookies enabled
- [ ] Gmail API verified (for public users)

---

## Testing Scenarios

### Happy Path Test
1. Fresh user signs in with Google
2. Grants Gmail access
3. Claims username
4. Analysis completes successfully
5. 3 insights show on analyzing screens
6. 4 starter apps display
7. Completes onboarding
8. Redirects to dashboard

### Edge Cases to Test
- [ ] Analysis timeout (>2 minutes)
- [ ] Gmail API returns no emails
- [ ] Groq API error
- [ ] Username already taken
- [ ] OAuth state mismatch
- [ ] Token expiry during analysis
- [ ] Network interruption during polling

---

## Error Messages & Solutions

### "Analysis timeout"
**Cause:** Analysis took >2 minutes or failed silently
**Solution:**
1. Check backend logs for errors
2. Verify Groq API key is valid
3. Check Gmail API quota
4. Manually check analysis status in DB

### "Username already taken"
**Cause:** Username exists in DB
**Solution:**
1. Try different username
2. Check DB: `SELECT username FROM "user" WHERE username = 'test';`
3. If orphaned, delete old user

### "Gmail not connected"
**Cause:** OAuth flow didn't complete or tokens not stored
**Solution:**
1. Check `user_integrations` table
2. Re-run OAuth flow
3. Verify `GOOGLE_REDIRECT_URI` matches Google Console

### "Invalid state parameter"
**Cause:** CSRF token mismatch
**Solution:**
1. Clear browser cookies
2. Start OAuth flow fresh
3. Check cookie settings (SameSite)

---

## Key Files for Quick Edits

### Backend
```
handlers/auth_google.go              # OAuth + token storage
handlers/osa_onboarding.go           # Analysis endpoints
services/onboarding_email_analyzer.go # Gmail extraction
services/onboarding_profile_analyzer.go # AI analysis
```

### Frontend
```
stores/onboardingAnalysis.ts         # Polling logic
routes/onboarding/analyzing/+page.svelte # Screen 6
routes/onboarding/username/+page.svelte  # Username validation
```

### Database
```
migrations/054_onboarding_user_analysis.sql
migrations/055_onboarding_starter_apps.sql
migrations/056_onboarding_email_metadata.sql
```

---

## Performance Metrics

### Expected Timings
- OAuth flow: 2-5 seconds
- Gmail API fetch (100 emails): 5-10 seconds
- Email metadata extraction: 2-3 seconds
- Groq AI analysis: 3-5 seconds
- Total analysis time: 10-18 seconds

### Polling Behavior
- Interval: 2 seconds
- Timeout: 2 minutes (60 attempts)
- Auto-advance delay: 2 seconds after completion

---

## Environment Variables Quick Copy

```bash
# Backend (.env)
DATABASE_URL=postgresql://user:pass@localhost:5432/businessos
GOOGLE_CLIENT_ID=your-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-secret
GOOGLE_REDIRECT_URI=http://localhost:8001/api/v1/auth/google/callback
GROQ_API_KEY=your-groq-key
AI_PROVIDER=groq
SECRET_KEY=change-this-in-production
ENVIRONMENT=development

# Frontend (.env)
VITE_API_URL=http://localhost:8001
PUBLIC_APP_URL=http://localhost:5173
```

---

## Useful SQL Snippets

### Count Users by Onboarding Status
```sql
SELECT
    onboarding_completed,
    COUNT(*)
FROM "user"
GROUP BY onboarding_completed;
```

### Analysis Success Rate
```sql
SELECT
    status,
    COUNT(*),
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
FROM onboarding_user_analysis
GROUP BY status;
```

### Average Analysis Duration
```sql
SELECT
    AVG(analysis_duration_ms) as avg_ms,
    MIN(analysis_duration_ms) as min_ms,
    MAX(analysis_duration_ms) as max_ms
FROM onboarding_user_analysis
WHERE status = 'completed';
```

### Top Detected Tools
```sql
SELECT
    jsonb_array_elements_text(tools_used) as tool,
    COUNT(*) as mentions
FROM onboarding_user_analysis
WHERE status = 'completed'
GROUP BY tool
ORDER BY mentions DESC
LIMIT 10;
```

---

## Quick Fixes

### Reset Frontend State
```javascript
// In browser console
localStorage.removeItem('osa_onboarding_state');
document.cookie.split(";").forEach(c => {
    document.cookie = c.trim().split("=")[0] +
        "=;expires=Thu, 01 Jan 1970 00:00:00 UTC;path=/;";
});
location.reload();
```

### Force Analysis Completion (Debug)
```sql
UPDATE onboarding_user_analysis
SET
    status = 'completed',
    insights = '["No-code builder energy", "Design tools are your playground", "AI-curious, testing new platforms"]'::jsonb,
    completed_at = NOW()
WHERE user_id = 'usr_xyz123';
```

### Clear All Onboarding Data
```sql
TRUNCATE onboarding_starter_apps CASCADE;
TRUNCATE onboarding_email_metadata CASCADE;
TRUNCATE onboarding_user_analysis CASCADE;
UPDATE "user" SET onboarding_completed = false;
```

---

**Quick Help:** See full documentation in `ONBOARDING_SYSTEM.md`
