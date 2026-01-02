$env:PGPASSWORD = "yasdas230321*"
& 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -U postgres -d postgres -c "\d memories" | Select-String -Pattern "tags"
Remove-Item Env:\PGPASSWORD
