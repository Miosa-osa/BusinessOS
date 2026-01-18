# 🧪 OAuth + Gmail Integration Test Guide

## ✅ System Reset Complete

**Test User:** roberto@healthbusiness.ai (ID: wXspJcOne7eFxZlXSsoBdQ)

**What Was Reset:**
- ✅ Cleared all user_integrations (Gmail tokens)
- ✅ Cleared onboarding_user_analysis data
- ✅ Cleared onboarding_email_metadata
- ✅ Deleted all sessions (force re-login)
- ✅ Reset onboarding_completed to FALSE
- ✅ Cleared username

**Backend Status:**
- ✅ Running on port 8001
- ✅ OAuth scopes configured correctly:
  - https://www.googleapis.com/auth/userinfo.email
  - https://www.googleapis.com/auth/userinfo.profile
  - https://mail.google.com/ (FULL Gmail access)

---

## 📋 Complete Test Flow

### Step 1: Access Signin Page
```
Navigate to: http://localhost:5173/onboarding/signin
```

**Expected:** See "Get your first AI apps" screen with Gmail icon

---

### Step 2: Click "Connect with Google"

**Expected:**
- Redirects to Google OAuth consent screen
- Shows permissions for:
  - ✅ Email address
  - ✅ Basic profile info
  - ✅ Read, compose, send, and permanently delete all Gmail emails

**⚠️ IMPORTANT:** If Google doesn't show Gmail permissions, there's a config issue.

---

### Step 3: Approve Google Consent

**Expected:**
- Redirects back to: http://localhost:5173/onboarding/username
- User is logged in
- Backend logs should show:

```
📧 [Gmail] Storing tokens and starting analysis for user: wXspJcOne7eFxZlXSsoBdQ
✅ [Gmail] Tokens stored successfully for user: wXspJcOne7eFxZlXSsoBdQ
🔍 [Gmail Analysis] Starting background analysis for user: wXspJcOne7eFxZlXSsoBdQ
```

---

### Step 4: Verify Data in Database

Run this query to check everything was stored:

```sql
-- Check tokens were stored
SELECT 
    user_id, 
    provider_id, 
    status, 
    connected_at,
    array_length(scopes, 1) as scope_count,
    LENGTH(access_token_encrypted::text) > 0 as has_access_token,
    LENGTH(refresh_token_encrypted::text) > 0 as has_refresh_token
FROM user_integrations 
WHERE user_id = 'wXspJcOne7eFxZlXSsoBdQ';
```

**Expected Output:**
```
user_id: wXspJcOne7eFxZlXSsoBdQ
provider_id: google_gmail
status: connected
connected_at: [timestamp]
scope_count: 3
has_access_token: true
has_refresh_token: true
```

---

### Step 5: Check Background Analysis Started

```sql
SELECT 
    user_id, 
    status, 
    total_emails_analyzed,
    analysis_model,
    created_at
FROM onboarding_user_analysis
WHERE user_id = 'wXspJcOne7eFxZlXSsoBdQ';
```

**Expected:**
- Row may not exist yet (analysis runs async in background)
- Once analysis starts, status will be 'analyzing'
- When complete, status will be 'completed'

---

### Step 6: Check Backend Logs

```bash
tail -f /tmp/backend-oauth.log | grep -E "(Gmail|Analysis)"
```

**Expected to see:**
```
📧 [Gmail] Storing tokens and starting analysis for user: wXspJcOne7eFxZlXSsoBdQ
✅ [Gmail] Tokens stored successfully for user: wXspJcOne7eFxZlXSsoBdQ
🔍 [Gmail Analysis] Starting background analysis for user: wXspJcOne7eFxZlXSsoBdQ
📊 [Gmail Analysis] Background analysis started for user: wXspJcOne7eFxZlXSsoBdQ (will run asynchronously)
```

---

## 🔍 Deep Dive Verification

### Check Scopes Were Saved Correctly

```sql
SELECT 
    user_id,
    scopes
FROM user_integrations 
WHERE user_id = 'wXspJcOne7eFxZlXSsoBdQ';
```

**Expected scopes array:**
```
{
  "https://www.googleapis.com/auth/userinfo.email",
  "https://www.googleapis.com/auth/userinfo.profile", 
  "https://mail.google.com/"
}
```

---

### Check Token Expiry

```sql
SELECT 
    user_id,
    token_expires_at,
    token_expires_at > NOW() as is_valid
FROM user_integrations 
WHERE user_id = 'wXspJcOne7eFxZlXSsoBdQ';
```

**Expected:**
- token_expires_at: ~1 hour from OAuth time
- is_valid: true

---

### Check Session Was Created

```sql
SELECT 
    "userId",
    token,
    "expiresAt",
    "createdAt"
FROM session
WHERE "userId" = 'wXspJcOne7eFxZlXSsoBdQ'
ORDER BY "createdAt" DESC
LIMIT 1;
```

**Expected:**
- Session exists with 30-day expiry
- Token is set (used in cookie)

---

## 🐛 Troubleshooting

### Issue: No tokens stored in database

**Check:**
```bash
grep "Failed to store Gmail tokens" /tmp/backend-oauth.log
```

**Possible causes:**
- Foreign key constraint error (provider_id mismatch)
- Database connection issue
- Token encryption error

---

### Issue: Google doesn't show Gmail permissions

**Check OAuth URL:**
```bash
curl -s "http://localhost:8001/api/auth/google?redirect=http://localhost:5173/onboarding/username" | grep -o "scope=[^&]*"
```

**Should include:** `scope=...https%3A%2F%2Fmail.google.com%2F`

---

### Issue: Analysis not starting

**Check logs:**
```bash
grep "background analysis" /tmp/backend-oauth.log
```

**Note:** Analysis launches in goroutine - errors may not surface immediately.

---

## 📊 Success Criteria

✅ **OAuth works:** User authenticates with Google  
✅ **Tokens stored:** user_integrations table has row with google_gmail provider  
✅ **Scopes correct:** Array includes https://mail.google.com/  
✅ **Session created:** User is logged in and redirected to username page  
✅ **Analysis started:** Background goroutine launches (logs show "Starting background analysis")  
✅ **Non-blocking:** User can continue through onboarding while analysis runs  

---

## 🎯 Next Steps After Successful Test

1. Implement actual Gmail email fetching in background goroutine
2. Call EmailAnalyzerService.AnalyzeRecentEmails()
3. Store results in onboarding_user_analysis table
4. Display analysis insights on analyzing screens
5. Generate starter apps based on analysis

---

**Test ready!** 🚀 You can now test the complete OAuth flow.
