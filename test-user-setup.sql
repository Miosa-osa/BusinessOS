-- Setup test user for API testing

-- Get or reuse existing test user
DO $$
DECLARE
    v_user_id TEXT;
BEGIN
    -- Try to find existing user
    SELECT id INTO v_user_id FROM "user" WHERE email = 'testuser@businessos.dev';
    
    -- Create if doesn't exist
    IF v_user_id IS NULL THEN
        v_user_id := 'test-user-' || replace(gen_random_uuid()::text, '-', '');
        INSERT INTO "user" (id, name, email, "emailVerified")
        VALUES (v_user_id, 'Test User BusinessOS', 'testuser@businessos.dev', true);
        RAISE NOTICE 'Created new user: %', v_user_id;
    ELSE
        RAISE NOTICE 'Using existing user: %', v_user_id;
    END IF;
    
    -- Clean up old test sessions
    DELETE FROM session WHERE token = 'test-token-businessos-123';
    
    -- Create fresh session
    INSERT INTO session (id, "userId", token, "expiresAt")
    VALUES (
        'test-session-' || replace(gen_random_uuid()::text, '-', ''),
        v_user_id,
        'test-token-businessos-123',
        NOW() + INTERVAL '30 days'
    );
    
    RAISE NOTICE 'Created session token: test-token-businessos-123';
END $$;

-- Verify setup
\echo '\nTest user and session created:'
SELECT 
    u.id as user_id,
    u.name,
    u.email,
    s.token,
    s."expiresAt"
FROM "user" u
JOIN session s ON s."userId" = u.id
WHERE s.token = 'test-token-businessos-123';
