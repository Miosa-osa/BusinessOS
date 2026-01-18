# Reset Onboarding for Existing Users

This script resets the `onboarding_completed` flag for all existing users, forcing them through the new OSA Build onboarding flow.

## Why Reset Onboarding?

The new OSA Build onboarding flow (13 screens) is completely different from the previous onboarding. To ensure all users experience the new flow and complete all required steps (username, Gmail connection, AI analysis, starter apps), we reset their onboarding status.

## What Does It Do?

- Sets `onboarding_completed = FALSE` for all users in the database
- Updates the `updatedAt` timestamp
- Logs the number of users affected

## How to Run

### Option 1: Direct psql Command (Production)
```bash
psql $DATABASE_URL -f reset_onboarding.sql
```

### Option 2: Local Development
```bash
# Make sure your DATABASE_URL is set in .env
source .env
psql $DATABASE_URL -f reset_onboarding.sql
```

### Option 3: Run from Go Code
```bash
# From backend-go directory
go run ./cmd/server -migrate-reset-onboarding
```
(Note: This requires implementing the flag in cmd/server/main.go if needed)

## After Running

1. All existing users will be redirected to `/onboarding` when they next log in
2. They must complete all 13 onboarding screens:
   - Welcome
   - Meet OSA
   - Sign in with Google
   - Connect Gmail (optional)
   - Choose username
   - AI analysis (3 screens with insights)
   - Starter apps (4 personalized apps)
   - Ready screen

3. After completion, they can access `/window` (main app)

## Verification

Check affected users:
```sql
SELECT COUNT(*) FROM "user" WHERE onboarding_completed = FALSE;
```

Check specific user:
```sql
SELECT id, name, email, onboarding_completed FROM "user" WHERE email = 'user@example.com';
```

## Frontend Behavior

The frontend automatically handles this:
- Root `/` checks `onboardingCompleted` status
- `/auth/callback` checks after OAuth
- Users with `onboardingCompleted = false` are redirected to `/onboarding`
- After completing onboarding, backend sets `onboarding_completed = TRUE`

## Rollback (if needed)

To mark all users as having completed onboarding (emergency rollback):
```sql
UPDATE "user" SET onboarding_completed = TRUE, "updatedAt" = NOW();
```
