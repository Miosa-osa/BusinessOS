# Backend Scripts

PowerShell scripts for database setup and maintenance.

## Scripts

### apply-migrations.ps1

Applies all database migrations to local PostgreSQL.

**Usage:**
```powershell
.\apply-migrations.ps1
```

**What it does:**
1. Sets PGPASSWORD environment variable
2. Runs `supabase-migrations-combined.sql`
3. Creates 26 database tables
4. Lists all created tables
5. Cleans up environment variables

**Output:**
- Migration SQL statements (CREATE TABLE, CREATE INDEX, etc.)
- List of 26 tables created
- Success confirmation

**Prerequisites:**
- PostgreSQL 17 or 18 installed
- Password: `yasdas230321*` (configured in script)
- Combined migration file at root: `../../supabase-migrations-combined.sql`

**Tables Created:**
- memories, user_facts, memory_associations
- uploaded_documents, document_chunks
- conversation_summaries, learning_events
- application_profiles, context_profiles
- And 17 more tables

**Created:** January 2, 2026
**Last Updated:** January 2, 2026

## Environment

All scripts expect:
- PostgreSQL installed at: `C:\Program Files\PostgreSQL\18\bin\psql.exe`
- Database: `postgres`
- User: `postgres`
- Password: Configured in each script

## Troubleshooting

### Script Not Found
```powershell
# Run from correct directory
cd C:\Users\Pichau\Desktop\BusinessOS-main-dev\desktop\backend-go\scripts
```

### Permission Denied
```powershell
# Set execution policy
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
.\apply-migrations.ps1
```

### Migration File Not Found
```bash
# Verify file exists
Test-Path "..\..\supabase-migrations-combined.sql"
# Should return: True
```

### Password Authentication Failed
```powershell
# Update password in script
# Edit apply-migrations.ps1, line 5:
$env:PGPASSWORD = "your_actual_password"
```

## Related Files

**Root Directory:**
- `supabase-migrations-combined.sql` - Combined migration file (65 KB)
- `test-user-setup.sql` - Test user creation SQL
- `setup-test-user.ps1` - Test user setup script (deprecated)
- `run-test-setup.ps1` - Test user runner

**Documentation:**
- `docs/DATABASE_SETUP.md` - Complete database setup guide
- `docs/DEVELOPER_QUICKSTART.md` - Quick start guide

## See Also

- [DATABASE_SETUP.md](../../../docs/DATABASE_SETUP.md) - Full database documentation
- [DEVELOPER_QUICKSTART.md](../../../docs/DEVELOPER_QUICKSTART.md) - Development guide
