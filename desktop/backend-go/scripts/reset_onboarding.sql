-- Reset onboarding status for all users
-- This forces existing users through the new OSA Build onboarding flow
-- Run this migration to reset all users: psql $DATABASE_URL -f reset_onboarding.sql

BEGIN;

-- Reset onboarding_completed to FALSE for all users
UPDATE "user"
SET onboarding_completed = FALSE,
    "updatedAt" = NOW()
WHERE onboarding_completed = TRUE;

-- Log the number of users reset
DO $$
DECLARE
    reset_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO reset_count
    FROM "user"
    WHERE onboarding_completed = FALSE;

    RAISE NOTICE 'Reset onboarding for users. Total users with onboarding_completed = FALSE: %', reset_count;
END $$;

COMMIT;
