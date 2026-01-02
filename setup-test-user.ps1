# Create test user and session for API testing
$env:PGPASSWORD = "yasdas230321*"
$psqlPath = "C:\Program Files\PostgreSQL\18\bin\psql.exe"

Write-Host "Setting up test user..." -ForegroundColor Cyan

# Create user and session tables if they don't exist (Better Auth schema)
$sql = @"
-- Create user table if not exists
CREATE TABLE IF NOT EXISTS "user" (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    "emailVerified" BOOLEAN DEFAULT FALSE,
    image TEXT,
    "createdAt" TIMESTAMP DEFAULT NOW(),
    "updatedAt" TIMESTAMP DEFAULT NOW()
);

-- Create session table if not exists
CREATE TABLE IF NOT EXISTS session (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    "userId" TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    "expiresAt" TIMESTAMP NOT NULL,
    "createdAt" TIMESTAMP DEFAULT NOW(),
    "updatedAt" TIMESTAMP DEFAULT NOW()
);

-- Insert test user
INSERT INTO "user" (id, name, email, "emailVerified", image)
VALUES ('test-user-123', 'Test User', 'test@example.com', true, null)
ON CONFLICT (email) DO UPDATE SET name = 'Test User';

-- Insert test session (expires in 30 days)
INSERT INTO session (id, "userId", token, "expiresAt")
VALUES ('test-session-123', 'test-user-123', 'test-token-123', NOW() + INTERVAL '30 days')
ON CONFLICT (token) DO UPDATE SET "expiresAt" = NOW() + INTERVAL '30 days';

-- Show created user and session
SELECT 'User created:' as status, id, name, email FROM "user" WHERE id = 'test-user-123';
SELECT 'Session created:' as status, token, "expiresAt" FROM session WHERE token = 'test-token-123';
"@

# Apply SQL
$sql | & $psqlPath -U postgres -d postgres

Write-Host ""
Write-Host "Test user setup complete!" -ForegroundColor Green
Write-Host "User ID: test-user-123" -ForegroundColor Yellow
Write-Host "Session Token: test-token-123" -ForegroundColor Yellow
Write-Host ""
Write-Host "Use this cookie for API testing:" -ForegroundColor Cyan
Write-Host 'Cookie: better-auth.session_token=test-token-123' -ForegroundColor White

# Clean up
Remove-Item Env:\PGPASSWORD
