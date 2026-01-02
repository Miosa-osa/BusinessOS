$env:PGPASSWORD = "yasdas230321*"
$psqlPath = "C:\Program Files\PostgreSQL\18\bin\psql.exe"

Write-Host "Checking auth tables..." -ForegroundColor Cyan

& $psqlPath -U postgres -d postgres -c "\d ""user"""
& $psqlPath -U postgres -d postgres -c "SELECT * FROM ""user"" LIMIT 5;"

Remove-Item Env:\PGPASSWORD
