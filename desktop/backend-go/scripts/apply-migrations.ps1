# Apply migrations to local PostgreSQL
# Usage: .\apply-migrations.ps1

$env:PGPASSWORD = "yasdas230321*"
$psqlPath = "C:\Program Files\PostgreSQL\18\bin\psql.exe"
$migrationFile = "C:\Users\Pichau\Desktop\BusinessOS-main-dev\supabase-migrations-combined.sql"

Write-Host "Applying migrations to local PostgreSQL..." -ForegroundColor Cyan

# Check if migration file exists
if (-not (Test-Path $migrationFile)) {
    Write-Host "Migration file not found: $migrationFile" -ForegroundColor Red
    exit 1
}

# Apply migrations
& $psqlPath -U postgres -d postgres -f $migrationFile

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Migrations applied successfully!" -ForegroundColor Green

    # List tables
    Write-Host ""
    Write-Host "Listing tables in database:" -ForegroundColor Cyan
    & $psqlPath -U postgres -d postgres -c "\dt"
} else {
    Write-Host ""
    Write-Host "Error applying migrations. Exit code: $LASTEXITCODE" -ForegroundColor Red
    exit 1
}

# Clean up password from environment
Remove-Item Env:\PGPASSWORD
