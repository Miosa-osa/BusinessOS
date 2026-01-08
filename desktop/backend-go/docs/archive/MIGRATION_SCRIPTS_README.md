# Database Migration Scripts

## Overview
This directory contains Go scripts for managing BusinessOS database migrations.

## Quick Start

### Check Database Status
```bash
go run check_db_status.go
```
Shows current migration status, table counts, and database health.

### Verify All Migrations
```bash
go run verify_migrations.go
```
Comprehensive verification of all migrations, tables, functions, and indexes.

---

## Available Scripts

### 1. `check_db_status.go`
**Purpose:** Quick database health check

**What it shows:**
- Number of applied migrations
- Latest migration version
- Row counts for key tables
- Installed extensions
- Database size and stats
- Recent activity

**Use when:** You want a quick status overview

---

### 2. `verify_migrations.go`
**Purpose:** Comprehensive migration verification

**What it checks:**
- All applied migrations with timestamps
- All expected tables from each migration
- All database functions
- PostgreSQL extensions (pgvector)
- Vector similarity indexes
- Complete verification summary

**Use when:** After applying migrations or troubleshooting

---

### 3. `run_all_pending_migrations.go`
**Purpose:** Apply pending migrations

**What it does:**
- Checks which migrations are already applied
- Applies any pending migrations in order
- Uses transactions for safety (rollback on error)
- Records each migration in schema_migrations table
- Verifies tables after migration

**Use when:** You have new migration files to apply

---

### 4. `mark_existing_migrations.go`
**Purpose:** Mark pre-existing migrations

**What it does:**
- Checks if tables from migrations already exist
- Marks those migrations as applied
- Useful when migrations were run manually

**Use when:** Migrations were applied outside the tracking system

---

### 5. `fix_migration_029_functions.go`
**Purpose:** Fix missing functions from migration 029

**What it creates:**
- has_project_access()
- get_project_role()
- get_project_permissions()

**Use when:** Migration 029 functions are missing

---

## Migration Files Location

All migration SQL files are in:
```
desktop/backend-go/internal/database/migrations/
```

Current migrations:
- 025_image_embeddings.sql
- 026_workspaces_and_roles.sql
- 027_workspace_invites.sql
- 028_workspace_audit_logs.sql
- 029_project_members.sql
- 030_memory_hierarchy_v2.sql

---

## Database Connection

All scripts use the following connection:
```
Host: db.fuqhjbgbjamtxcdphjpp.supabase.co
Database: postgres
User: postgres
```

Connection string is hardcoded in each script for convenience.

---

## Migration Tracking

Migrations are tracked in the `schema_migrations` table:

```sql
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

To check manually:
```sql
SELECT * FROM schema_migrations ORDER BY version;
```

---

## Adding New Migrations

1. Create SQL file: `NNN_description.sql` in migrations folder
2. Update `run_all_pending_migrations.go`:
   ```go
   {"NNN", "Description", "internal/database/migrations/NNN_description.sql"},
   ```
3. Run: `go run run_all_pending_migrations.go`
4. Verify: `go run verify_migrations.go`

---

## Troubleshooting

### Migration fails with "already exists" error
The objects already exist. Run `mark_existing_migrations.go` to record them.

### Functions missing after migration
Run the specific fix script (e.g., `fix_migration_029_functions.go`)

### Can't connect to database
Check:
- Network connectivity
- Supabase project is running
- Connection string is correct
- Firewall allows outbound connections

### Migration partially applied
Migrations use transactions, so they should rollback on error. Check the error message and fix the SQL, then retry.

---

## Best Practices

1. **Always verify** after applying migrations
2. **Backup** before running migrations on production
3. **Test** migrations on a dev database first
4. **Review** migration SQL before running
5. **Track** all migrations in schema_migrations table

---

## Current Status (as of 2026-01-06)

✅ All 6 migrations applied successfully
✅ All tables created (15 total)
✅ All functions created (12 total)
✅ Vector extension enabled
✅ Vector indexes created

Database is ready for use!

---

For more details, see: `MIGRATION_SUMMARY.md` in project root
