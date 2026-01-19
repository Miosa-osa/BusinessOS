# Supabase Setup Guide for BusinessOS Backend

This guide walks you through setting up Supabase as the database provider for BusinessOS (standalone/local setup).

## Prerequisites

- Supabase account with project created
- Project ID: `fuqhjbgbjamtxcdphjpp`
- Project URL: `https://fuqhjbgbjamtxcdphjpp.supabase.co`
- Go 1.21+ installed

## Step 1: Run the Database Migration

1. **Open Supabase SQL Editor**
   - Navigate to: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/editor
   - Click "SQL Editor" in the left sidebar
   - Click "+ New query"

2. **Load the Migration Script**
   - Open the migration file: `internal/database/migrations/supabase_migration.sql`
   - Copy the entire contents (759 lines)
   - Paste into the Supabase SQL Editor

3. **Execute the Migration**
   - Click "Run" button (or press Ctrl/Cmd + Enter)
   - Wait for execution to complete
   - You should see: "Success. No rows returned"

4. **Verify Tables Created**
   - Click "Table Editor" in the left sidebar
   - You should see 40+ tables including:
     - `contexts`
     - `conversations`
     - `messages`
     - `projects`
     - `clients`
     - `tasks`
     - `team_members`
     - etc.

## Step 2: Get Your Database Connection String

1. **Navigate to Database Settings**
   - Go to: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/settings/database
   - Scroll to "Connection string" section

2. **Copy the Pooled Connection String**
   - **IMPORTANT**: Select "Connection pooling" mode
   - Select "Transaction" mode
   - The connection string should look like:
     ```
     postgres://postgres.fuqhjbgbjamtxcdphjpp:[YOUR-PASSWORD]@aws-0-us-east-1.pooler.supabase.com:6543/postgres?pgbouncer=true
     ```
   - **Critical**: Port must be **6543** (pooled), NOT 5432 (direct)
   - Copy this string and save it securely

3. **Find Your Database Password**
   - If you don't remember your database password, click "Reset database password"
   - Save the new password immediately
   - Replace `[YOUR-PASSWORD]` in the connection string with your actual password

## Step 3: Configure Local Environment

1. **Create .env File**
   ```bash
   cd /Users/ososerious/BusinessOS-1/desktop/backend-go
   cp .env.example .env
   ```

2. **Edit .env File**
   - Open `.env` in your editor
   - Replace `[YOUR-PASSWORD]` with your actual database password
   - Example:
     ```
     DATABASE_URL=postgres://postgres.fuqhjbgbjamtxcdphjpp:your-actual-password@aws-0-us-east-1.pooler.supabase.com:6543/postgres?pgbouncer=true
     ```

3. **Verify .env is in .gitignore**
   - Ensure `.env` is listed in your `.gitignore` file
   - NEVER commit `.env` with real credentials

## Step 4: Test the Connection

1. **Build the Application**
   ```bash
   go mod download
   go build -o bin/server ./cmd/server
   ```

2. **Run the Server**
   ```bash
   ./bin/server
   ```

3. **Check the Logs**
   - You should see successful database connection messages
   - Look for: "Successfully connected to database" or similar
   - If you see connection errors, proceed to troubleshooting

4. **Test a Query (Optional)**
   ```bash
   # If you have a health check endpoint
   curl http://localhost:8080/health
   ```

## Understanding the Connection Setup

### Why Port 6543 (Pooled) vs Port 5432 (Direct)?

- **Port 5432 (Direct)**: Limited connections (60 free, 200 pro)
- **Port 6543 (Pooled)**: Uses Supavisor (PgBouncer) to handle 1000+ client connections
- Our app uses **port 6543** to prevent connection exhaustion

### Connection Pool Settings

The connection pool is configured in `internal/database/postgres.go`:

```go
MaxConns:           10    // Conservative for cross-cloud latency
MinConns:           2     // Keep some connections warm
MaxConnLifetime:    15min // Supabase closes stale connections
MaxConnIdleTime:    5min  // Release idle connections faster
HealthCheckPeriod:  30sec // More frequent health checks
```

These settings are optimized for:
- **Cross-cloud latency** (GCP → AWS Supabase)
- **Connection stability** (prevent stale connections)
- **Resource efficiency** (use pooler effectively)

## Troubleshooting

### Error: "connection refused"

**Cause**: Wrong connection string or network issues

**Solutions**:
1. Verify you're using port **6543** (pooled), not 5432
2. Check your database password is correct
3. Ensure the connection string includes `?pgbouncer=true`
4. Verify your IP is allowed in Supabase network settings

### Error: "too many connections"

**Cause**: Using direct connection (port 5432) instead of pooled

**Solutions**:
1. Change port from 5432 → 6543
2. Add `?pgbouncer=true` to connection string
3. Restart your application

### Error: "password authentication failed"

**Cause**: Incorrect database password

**Solutions**:
1. Go to Supabase dashboard → Settings → Database
2. Click "Reset database password"
3. Update your `.env` file with the new password
4. Restart your application

### Error: "SSL connection required"

**Cause**: Supabase requires SSL connections

**Solutions**:
1. Add `?sslmode=require` to your connection string
2. Or use `?sslmode=disable` for local testing only (NOT recommended for production)

### Slow Query Performance

**Cause**: Cross-cloud latency (GCP → AWS)

**Solutions**:
1. This is expected with cross-cloud setup (75-150ms added latency)
2. Consider optimizing queries with indexes
3. Use connection pooling effectively (already configured)
4. For production, consider deploying to AWS to reduce latency

## Next Steps

After successful connection:

1. **Run Tests**: `go test ./...`
2. **Verify SQLC**: If using SQLC, ensure queries work with Supabase
3. **Check Indexes**: Verify indexes are created properly
4. **Monitor Performance**: Use Supabase dashboard to monitor queries

## Security Checklist

- [ ] `.env` file is in `.gitignore`
- [ ] Database password is strong (20+ characters)
- [ ] Never commit real credentials to Git
- [ ] Use environment variables for all secrets
- [ ] Enable Row Level Security (RLS) in Supabase if needed
- [ ] Review Supabase network settings/allowed IPs

## Support

- **Supabase Docs**: https://supabase.com/docs
- **Supabase Dashboard**: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp
- **pgx Documentation**: https://pkg.go.dev/github.com/jackc/pgx/v5

## Migration Script Location

The complete migration script is located at:
```
/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/database/migrations/supabase_migration.sql
```

This script includes:
- 15+ custom ENUM types
- 40+ tables (contexts, conversations, messages, projects, clients, tasks, etc.)
- Comprehensive indexes for performance
- Foreign key constraints for data integrity
