# OSA Build User Flow - Complete Guide

## 🔄 HOW THE ONBOARDING FLOW WORKS

### Current Implementation Status

✅ **WORKING:**
- New user registration → Onboarding flow
- 13 onboarding screens fully built
- Frontend UI complete
- Backend APIs ready
- Username system ready

⚠️ **NEEDS FIXING:**
- Existing user login → Should check if onboarding completed
- Currently existing users bypass onboarding and go to `/window`

---

## 📋 COMPLETE USER FLOW DIAGRAM

### Scenario 1: NEW USER SIGNUP

```
User visits /register
      ↓
Fills out form:
  - Name
  - Email
  - Password
  - Agrees to terms
      ↓
Clicks "Create account"
      ↓
signUpWithEmail() called
      ↓
SUCCESS → goto('/onboarding') ✅
      ↓
╔══════════════════════════════════════════════════════════════╗
║ ONBOARDING FLOW (13 Screens)                                ║
╠══════════════════════════════════════════════════════════════╣
║ Screen 1:  /onboarding                                       ║
║    → Welcome to OSA Build                                    ║
║    → 3 feature cards                                         ║
║    → "Get Started" button                                    ║
║                                                              ║
║ Screen 2:  /onboarding/meet-osa                              ║
║    → Meet OSA (AI Agent intro)                               ║
║    → Pulsing gradient orb                                    ║
║    → 4 capability cards                                      ║
║                                                              ║
║ Screen 3:  /onboarding/signin                                ║
║    → OAuth options (Google, Apple, Email)                    ║
║    → NOTE: Already signed in from /register                  ║
║    → This is about connecting OAuth for personalization      ║
║                                                              ║
║ Screen 4:  /onboarding/gmail                                 ║
║    → Connect Gmail (OPTIONAL)                                ║
║    → Can skip                                                ║
║    → Used for AI analysis                                    ║
║                                                              ║
║ Screen 5:  /onboarding/username                              ║
║    → Claim unique username                                   ║
║    → Real-time availability check                            ║
║    → API: GET /api/users/check-username/:username           ║
║    → API: PATCH /api/users/me/username                       ║
║                                                              ║
║ Screen 6:  /onboarding/analyzing                             ║
║    → "OSA is analyzing your data..."                         ║
║    → Pulsing OSA orb                                         ║
║    → Shows 1st insight                                       ║
║    → API: POST /api/osa-onboarding/analyze                   ║
║    → Auto-advance after 2s                                   ║
║                                                              ║
║ Screen 7:  /onboarding/analyzing-2                           ║
║    → Shows 2nd insight                                       ║
║    → Auto-advance after 2s                                   ║
║                                                              ║
║ Screen 8:  /onboarding/analyzing-3                           ║
║    → Shows 3rd insight                                       ║
║    → Auto-advance after 2s                                   ║
║                                                              ║
║ Screens 9-12:  /onboarding/starter-apps (CAROUSEL)           ║
║    → "Here are the apps we built for you"                    ║
║    → Shows 1 of 4 apps at a time                             ║
║    → Left/Right navigation                                   ║
║    → Progress: "1 of 4", "2 of 4", etc.                      ║
║    → Must view all 4 before continuing                       ║
║    → API: POST /api/osa-onboarding/generate-apps             ║
║                                                              ║
║ Screen 13: /onboarding/ready                                 ║
║    → "Your OS is ready! 🎉"                                  ║
║    → Animated success checkmark                              ║
║    → "Enter Your OS" button                                  ║
║    → Calls: onboardingStore.complete()                       ║
║    → Saves: { completed: true } to localStorage              ║
╚══════════════════════════════════════════════════════════════╝
      ↓
goto('/window') → MAIN APP ✅
```

### Scenario 2: EXISTING USER LOGIN (CURRENT - NEEDS FIX)

```
User visits /login
      ↓
Enters email + password
      ↓
Clicks "Sign in"
      ↓
signInWithEmail() called
      ↓
SUCCESS → goto('/window') ⚠️
      ↓
PROBLEM: No check if onboarding completed!
```

### Scenario 3: EXISTING USER LOGIN (SHOULD BE)

```
User visits /login
      ↓
Enters email + password
      ↓
Clicks "Sign in"
      ↓
signInWithEmail() called
      ↓
SUCCESS → Check onboarding status
      ↓
Has user completed onboarding?
      ├─ YES → goto('/window') ✅
      └─ NO  → goto('/onboarding') ✅
```

---

## 🧪 HOW TO TEST AS A NEW USER

### Prerequisites

1. **Start Backend:**
```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run cmd/server/main.go
```

2. **Start Frontend:**
```bash
cd /Users/rhl/Desktop/BusinessOS2/frontend
npm run dev
```

3. **Apply Migrations:**
```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go

# Migration 050: Onboarding profiles
psql $DATABASE_URL < internal/database/migrations/050_osa_build_onboarding.sql

# Migration 051: Username system
psql $DATABASE_URL < internal/database/migrations/051_add_username_system.sql
```

### Step-by-Step Testing

#### Part 1: New User Signup

1. **Open browser to:** `http://localhost:5173/register`

2. **Fill out registration form:**
   - Name: `Test User`
   - Email: `test@example.com` (use a unique email)
   - Password: `password123` (at least 8 chars)
   - Check "I agree to the Terms"

3. **Click "Create account"**

4. **Expected:** Redirected to `/onboarding` (Screen 1)

#### Part 2: Walk Through Onboarding

**Screen 1: Welcome** (`/onboarding`)
- Should see: "Welcome to OSA Build" with gradient background
- 3 feature cards: AI-Powered, Day 1 Apps, Fully Customizable
- Click "Get Started"

**Screen 2: Meet OSA** (`/onboarding/meet-osa`)
- Should see: Pulsing gradient orb
- 4 capability cards
- Click "Continue"

**Screen 3: Sign In** (`/onboarding/signin`)
- Should see: OAuth options (Google, Apple, Email)
- Note: You're already authenticated from registration
- This is for connecting OAuth for personalization
- Click "Skip for now" or choose an option

**Screen 4: Connect Gmail** (`/onboarding/gmail`)
- Should see: "Make Your OS Truly Yours"
- 3 benefits listed
- Click "Connect Gmail" OR "Skip for now"

**Screen 5: Claim Username** (`/onboarding/username`)
- Type a username (e.g., `testuser123`)
- Should see real-time availability check
- Check DevTools Network tab:
  - `GET /api/users/check-username/testuser123`
  - Response: `{"available": true}` or `{"available": false, "reason": "..."}`
- Click "Continue" when available

**Screen 6-8: AI Analysis** (`/onboarding/analyzing`, `/analyzing-2`, `/analyzing-3`)
- Each screen shows for ~2 seconds
- Pulsing OSA orb animation
- 3 insights displayed one by one
- Auto-advances
- Check DevTools Network tab:
  - `POST /api/osa-onboarding/analyze`

**Screens 9-12: Starter Apps Carousel** (`/onboarding/starter-apps`)
- Shows 1 of 4 apps at a time
- Use ← → buttons or arrow keys to navigate
- Progress shows "1 of 4", "2 of 4", etc.
- Must view all 4 apps
- "Continue to Your OS" button appears after viewing all
- Check DevTools Network tab:
  - `POST /api/osa-onboarding/generate-apps`

**Screen 13: Your OS is Ready** (`/onboarding/ready`)
- Success animation
- "Your OS is ready! 🎉"
- Click "Enter Your OS"
- Redirected to `/window` (main app)

#### Part 3: Verify State Persistence

1. **Check localStorage:**
   - Open DevTools → Application → Local Storage
   - Look for `osa_onboarding_state`
   - Should show: `{ "completed": true, ... }`

2. **Refresh the page:**
   - Should stay on `/window`
   - Should NOT go back to onboarding

3. **Check database:**
```bash
psql $DATABASE_URL -c "SELECT * FROM workspace_onboarding_profiles ORDER BY created_at DESC LIMIT 5;"
```
   - Should see your onboarding profile with analysis_data and starter_apps_data

#### Part 4: Test Existing User Login (Current Behavior)

1. **Log out** (if you have a logout button) OR **open incognito window**

2. **Go to:** `http://localhost:5173/login`

3. **Sign in** with the same email/password

4. **Current behavior:** Goes directly to `/window` (bypasses onboarding check)

5. **Expected behavior (after fix):** Should check if onboarding completed:
   - If completed → go to `/window` ✅
   - If not completed → go to `/onboarding` ✅

---

## 🐛 WHAT NEEDS TO BE FIXED

### Issue: Login Doesn't Check Onboarding Status

**File:** `frontend/src/routes/login/+page.svelte`

**Current code (line 59-60):**
```typescript
loading = false;
goto('/window');
```

**Should be:**
```typescript
loading = false;

// Check if onboarding is completed
const state = get(onboardingStore);
if (state.completed) {
  goto('/window');
} else {
  goto('/onboarding');
}
```

**But this has a problem:** It checks localStorage, not the backend.

### Better Solution: Check Backend for Onboarding Status

**Option 1: Add onboarding_completed field to user table**

Migration:
```sql
ALTER TABLE users ADD COLUMN onboarding_completed BOOLEAN DEFAULT FALSE;
```

When user completes onboarding (screen 13), call API:
```typescript
POST /api/users/me/complete-onboarding
```

On login, backend returns user object:
```json
{
  "id": "...",
  "email": "...",
  "username": "...",
  "onboarding_completed": true
}
```

Frontend checks this field:
```typescript
if (result.user.onboarding_completed) {
  goto('/window');
} else {
  goto('/onboarding');
}
```

**Option 2: Check workspace_onboarding_profiles table**

On login, call:
```typescript
GET /api/osa-onboarding/status
// Returns: { completed: boolean, currentStep: number }
```

Use this to decide where to redirect.

---

## 🎯 TESTING CHECKLIST

### New User Flow
- [ ] Register new account
- [ ] Redirected to /onboarding
- [ ] Walk through all 13 screens
- [ ] Username availability check works
- [ ] AI analysis API called
- [ ] Starter apps generated
- [ ] Redirected to /window at end
- [ ] localStorage shows completed: true
- [ ] Database has onboarding profile

### Existing User Flow (After Fix)
- [ ] User with completed onboarding → goes to /window
- [ ] User with incomplete onboarding → goes to /onboarding
- [ ] Can resume onboarding from where they left off

### Edge Cases
- [ ] Back button works (state preserved)
- [ ] Refresh page during onboarding (state preserved)
- [ ] Close browser and reopen (state preserved)
- [ ] Network errors handled gracefully

### API Endpoints
- [ ] `GET /api/users/check-username/:username` works
- [ ] `PATCH /api/users/me/username` works
- [ ] `POST /api/osa-onboarding/analyze` works
- [ ] `POST /api/osa-onboarding/generate-apps` works
- [ ] `GET /api/osa-onboarding/apps-status` works

---

## 📝 CURRENT STATUS SUMMARY

✅ **Ready to Test:**
- All 13 onboarding screens
- Username system backend
- OSA analysis backend
- Frontend UI complete
- localStorage persistence

⚠️ **Needs Work:**
- Login flow doesn't check onboarding status
- Need backend field to track onboarding completion
- Need API endpoint to check/update onboarding status

🎯 **Next Steps:**
1. Test the flow manually with a new user
2. Fix login redirect logic
3. Add backend onboarding status tracking
4. Test complete flow end-to-end

---

## 🚀 QUICK START TEST COMMANDS

```bash
# Terminal 1: Start backend
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go run cmd/server/main.go

# Terminal 2: Start frontend
cd /Users/rhl/Desktop/BusinessOS2/frontend
npm run dev

# Terminal 3: Apply migrations (one-time)
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
psql $DATABASE_URL < internal/database/migrations/050_osa_build_onboarding.sql
psql $DATABASE_URL < internal/database/migrations/051_add_username_system.sql

# Browser
# Open: http://localhost:5173/register
# Create test account and walk through flow
```

---

## 🔍 DEBUGGING TIPS

### Check onboarding state:
```javascript
// In browser console
JSON.parse(localStorage.getItem('osa_onboarding_state'))
```

### Clear onboarding state (to test again):
```javascript
// In browser console
localStorage.removeItem('osa_onboarding_state')
```

### Check database:
```bash
# View onboarding profiles
psql $DATABASE_URL -c "SELECT * FROM workspace_onboarding_profiles ORDER BY created_at DESC LIMIT 5;"

# View users with usernames
psql $DATABASE_URL -c "SELECT id, email, username, username_claimed_at FROM users WHERE username IS NOT NULL ORDER BY created_at DESC LIMIT 5;"

# View reserved usernames
psql $DATABASE_URL -c "SELECT * FROM reserved_usernames ORDER BY username;"
```

### Check API calls:
- Open DevTools → Network tab
- Filter: `XHR`
- Look for: `/api/users/check-username`, `/api/osa-onboarding/analyze`, etc.
- Inspect request/response payloads
