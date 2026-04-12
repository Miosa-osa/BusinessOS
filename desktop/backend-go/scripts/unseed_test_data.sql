-- ============================================
-- REMOVE SEED TEST DATA
-- Run: psql $DATABASE_URL -f unseed_test_data.sql
-- ============================================

DO $$
DECLARE
    test_user_id VARCHAR(255) := 'test-user-seed-001';
    deleted_count INT;
BEGIN

-- Delete in order to respect foreign key constraints

-- Delete emails
DELETE FROM emails WHERE user_id = test_user_id;
GET DIAGNOSTICS deleted_count = ROW_COUNT;
RAISE NOTICE 'Deleted % emails', deleted_count;

-- Delete calendar events
DELETE FROM calendar_events WHERE user_id = test_user_id;
GET DIAGNOSTICS deleted_count = ROW_COUNT;
RAISE NOTICE 'Deleted % calendar events', deleted_count;

-- Delete tasks (will cascade to subtasks)
DELETE FROM tasks WHERE user_id = test_user_id;
GET DIAGNOSTICS deleted_count = ROW_COUNT;
RAISE NOTICE 'Deleted % tasks', deleted_count;

-- Delete projects (after tasks that reference them)
DELETE FROM projects WHERE user_id = test_user_id;
GET DIAGNOSTICS deleted_count = ROW_COUNT;
RAISE NOTICE 'Deleted % projects', deleted_count;

-- Delete clients (after projects that reference them)
DELETE FROM clients WHERE user_id = test_user_id;
GET DIAGNOSTICS deleted_count = ROW_COUNT;
RAISE NOTICE 'Deleted % clients', deleted_count;

-- Delete team members (after tasks that reference them)
DELETE FROM team_members WHERE user_id = test_user_id;
GET DIAGNOSTICS deleted_count = ROW_COUNT;
RAISE NOTICE 'Deleted % team members', deleted_count;

RAISE NOTICE '✓ All seed data removed for user: %', test_user_id;

END $$;
