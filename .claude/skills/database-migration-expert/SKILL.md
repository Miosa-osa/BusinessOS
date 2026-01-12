---
name: database-migration-expert
description: Expert in PostgreSQL migrations, schema design, and sqlc integration for BusinessOS. Use when working with database schema, migrations, or when modifying database structure.
allowed-tools: Read, Edit, Write, Bash, Grep, Glob
---

# BusinessOS Database Migration Expert

## Migration File Structure

```sql
-- +migrate Up
-- Description of changes

CREATE TABLE IF NOT EXISTS example_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS example_table CASCADE;
```

## Best Practices

1. Always include both UP and DOWN migrations
2. Test rollback before committing
3. Use CONCURRENTLY for large table indexes
4. Create sqlc queries after schema changes
5. Regenerate sqlc code: `cd desktop/backend-go && sqlc generate`
