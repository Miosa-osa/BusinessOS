$env:PGPASSWORD = "yasdas230321*"
& 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -U postgres -d postgres -f "C:\Users\Pichau\Desktop\BusinessOS-main-dev\test-user-setup.sql"
Remove-Item Env:\PGPASSWORD

Write-Host ""
Write-Host "======================================" -ForegroundColor Green
Write-Host "Test Session Token: test-token-businessos-123" -ForegroundColor Yellow  
Write-Host "Use in Cookie: better-auth.session_token=test-token-businessos-123" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Green
