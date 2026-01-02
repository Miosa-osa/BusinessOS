$env:PGPASSWORD = "yasdas230321*"
$psqlPath = "C:\Program Files\PostgreSQL\18\bin\psql.exe"

Write-Host "Creating test user and session..." -ForegroundColor Cyan

$sql = @"
-- Find or create test user
DO \$\$
DECLARE
    v_user_id TEXT;
BEGIN
    -- Try to get existing user with test email
    SELECT id INTO v_user_id FROM "user" WHERE email = 'test@example.com';
    
    -- If not found, create new user
    IF v_user_id IS NULL THEN
        INSERT INTO "user" (id, name, email, "emailVerified")
        VALUES ('test-user-' || gen_random_uuid()::text, 'Test User', 'test@example.com', true)
        RETURNING id INTO v_user_id;
    END IF;
    
    -- Delete old test session if exists
    DELETE FROM session WHERE token = 'test-token-123';
    
    -- Create new session
    INSERT INTO session (id, "userId", token, "expiresAt")
    VALUES ('test-session-' || gen_random_uuid()::text, v_user_id, 'test-token-123', NOW() + INTERVAL '30 days');
    
    RAISE NOTICE 'User ID: %', v_user_id;
END \$\$;

-- Show result
SELECT u.id, u.name, u.email, s.token, s."expiresAt"
FROM "user" u
JOIN session s ON s."userId" = u.id
WHERE s.token = 'test-token-123';
"@

echo $sql | & $psqlPath -U postgres -d postgres

Write-Host ""
Write-Host "Test credentials ready!" -ForegroundColor Green
Write-Host "Session Token: test-token-123" -ForegroundColor Yellow

Remove-Item Env:\PGPASSWORD
