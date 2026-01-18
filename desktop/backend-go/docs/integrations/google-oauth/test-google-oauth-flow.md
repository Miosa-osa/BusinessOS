# Google OAuth Onboarding Flow - Testing Guide

## 🎯 CRITICAL: Restart Both Servers First!

The code changes won't work until you restart:

```bash
# Kill both servers (Ctrl+C in each terminal)

# Terminal 1: Restart Backend
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run cmd/server/main.go

# Terminal 2: Restart Frontend
cd /Users/rhl/Desktop/BusinessOS2/frontend
npm run dev
```

---

## 🧪 TESTING STEPS

### 1. Open Browser DevTools FIRST
- Open Chrome/Edge
- Press F12 to open DevTools
- Go to **Console** tab (keep it open)
- Go to **Network** tab (keep it open too)

### 2. Start the Flow
- Go to: `http://localhost:5173/login`
- Click "Continue with Google"
- Sign in with a **BRAND NEW** Google account (never used before)

### 3. Watch What Happens

**In the URL bar, you should see:**
```
http://localhost:5173/login
  ↓
https://accounts.google.com/... (Google login)
  ↓
http://localhost:5173/auth/callback
  ↓
http://localhost:5173/onboarding ← SHOULD GO HERE ✅
```

**If it goes to `/window` instead, that's the bug.**

---

## 🔍 DEBUGGING

### Check 1: Backend Logs

In your backend terminal, after Google OAuth, you should see:

```
[Something about creating new user]
```

Look for any log that says "new user" or similar.

### Check 2: Network Tab

In DevTools → Network tab:

1. Find the request to: `http://localhost:8001/api/auth/google/callback`
2. Click on it
3. Go to **Response Headers**
4. Look for: `Set-Cookie: new_user=true`

**If you DON'T see that cookie, the backend code didn't run.**

### Check 3: Console Tab

In DevTools → Console tab, you should see logs from the frontend:

```
[Auth] Session restored
```

Or similar auth-related logs.

### Check 4: Application Tab

DevTools → Application → Cookies → `http://localhost:5173`

Look for:
- `better-auth.session_token` (should exist)
- `new_user` (should exist briefly after Google OAuth, then disappear)

---

## 🐛 COMMON ISSUES

### Issue 1: Still Goes to /window

**Cause:** Backend server not restarted, so old code is running.

**Fix:**
```bash
# In backend terminal
# Press Ctrl+C
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run cmd/server/main.go
```

### Issue 2: No `new_user` cookie

**Cause:** Backend code changes didn't compile or there's an error.

**Fix:** Check backend terminal for compilation errors.

### Issue 3: Migration not applied

**Cause:** Database doesn't have `onboarding_completed` column.

**Fix:**
```bash
# Make sure DATABASE_URL is set
export DATABASE_URL='your-postgres-connection-string'

# Apply migration
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
psql $DATABASE_URL < internal/database/migrations/052_add_onboarding_completed.sql
```

### Issue 4: Using existing Google account

**Cause:** You used an account that already signed up before.

**Fix:** Use a completely new Google account, or delete the user from the database:

```sql
DELETE FROM "user" WHERE email = 'your-test-email@gmail.com';
```

---

## ✅ SUCCESS CRITERIA

After Google OAuth, you should:
1. ✅ See `/onboarding` in URL bar
2. ✅ See "Welcome to OSA Build" screen (Screen 1)
3. ✅ Be able to click "Get Started" and walk through all 13 screens
4. ✅ End up at `/window` after completing screen 13

---

## 📞 IF IT STILL DOESN'T WORK

Share:
1. Backend terminal logs (after Google OAuth)
2. Browser Network tab screenshot (showing the callback request)
3. Browser Console tab logs
4. What URL you end up at after Google OAuth
