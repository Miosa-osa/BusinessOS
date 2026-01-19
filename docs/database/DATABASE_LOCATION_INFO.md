# 🗄️ Database Location & Access Info

**Date**: 2026-01-06

---

## 📍 Database Location

Your database is hosted on **Supabase** (PostgreSQL as a Service).

```
Provider:    Supabase
Type:        PostgreSQL 15+
Region:      US East (AWS)
Status:      ✅ ACTIVE
```

---

## 🔗 Connection Details

### Primary Connection (Direct)
```
Host:        db.fuqhjbgbjamtxcdphjpp.supabase.co
Port:        5432
Database:    postgres
User:        postgres
Password:    Lunivate69420
```

**Connection String**:
```
postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30
```

### Pooler Connection (Alternative - COMMENTED OUT)
```
Host:        aws-0-us-east-1.pooler.supabase.com
Port:        6543
Database:    postgres
```

**Pooler Connection String** (currently disabled):
```
postgres://postgres.fuqhjbgbjamtxcdphjpp:Lunivate69420@aws-0-us-east-1.pooler.supabase.com:6543/postgres?connect_timeout=30
```

---

## 🌐 Supabase Dashboard

### Project URL
```
https://fuqhjbgbjamtxcdphjpp.supabase.co
```

### Dashboard Access
```
https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp
```

**Sections**:
- **Table Editor**: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/editor
- **SQL Editor**: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/sql
- **Database**: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/database/tables
- **API Docs**: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/api
- **Logs**: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/logs/explorer

---

## 📊 Database Schema Overview

Your database contains **50+ tables** for the BusinessOS platform:

### Core Tables
- `workspaces` - Multi-tenant containers (✅ Migration 026)
- `workspace_roles` - Role definitions (✅ Migration 026)
- `workspace_members` - User-workspace assignments (✅ Migration 026)
- `user_workspace_profiles` - Per-workspace profiles (✅ Migration 026)
- `workspace_memories` - Shared knowledge base (✅ Migration 026)

### Business Objects
- `projects` - Project management
- `clients` - Client/CRM data
- `tasks` - Task management
- `conversations` - Chat conversations
- `contexts` - Knowledge Base articles

### AI/Agent System
- `custom_agents` - User-defined agents (✅ Migration 009)
- `agent_presets` - Built-in agent templates (✅ Migration 009)
- `thinking_traces` - Chain of thought logs (✅ Migration 008)
- `focus_mode_templates` - Focus configurations (✅ Migration 013)

### Embeddings & RAG
- `documents` - Uploaded documents
- `document_chunks` - Chunked content with embeddings
- `memories` - User-specific memories

### Team & Collaboration
- `team_members` - Team roster
- `project_members` - Project assignments (✅ Enhanced in Migration 026)
- `calendar_events` - Calendar integration

### Settings & Preferences
- `user_settings` - User preferences (✅ Migration 027 added thinking columns)
- `user_commands` - Custom slash commands
- `output_styles` - AI output preferences

---

## 🔧 Environment Configuration

The database connection is configured in:

**File**: `desktop/backend-go/.env`

```env
# PostgreSQL Database (Supabase)
DATABASE_URL=postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30

# Alternative: Connection Pooler (currently disabled)
# DATABASE_URL=postgres://postgres.fuqhjbgbjamtxcdphjpp:Lunivate69420@aws-0-us-east-1.pooler.supabase.com:6543/postgres?connect_timeout=30

# Supabase Project URL
SUPABASE_URL=https://fuqhjbgbjamtxcdphjpp.supabase.co
```

---

## 🛠️ How to Access the Database

### Option 1: Supabase Dashboard (Easiest)
1. Go to https://supabase.com
2. Login to your account
3. Navigate to project `fuqhjbgbjamtxcdphjpp`
4. Use **Table Editor** or **SQL Editor**

### Option 2: psql (Command Line)
```bash
psql "postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30"
```

### Option 3: pgAdmin / DBeaver (GUI Tools)
**Connection Settings**:
- Host: `db.fuqhjbgbjamtxcdphjpp.supabase.co`
- Port: `5432`
- Database: `postgres`
- Username: `postgres`
- Password: `Lunivate69420`
- SSL: Prefer

### Option 4: Via Backend Code
The Go backend automatically connects using `DATABASE_URL` from `.env`:

```bash
cd desktop/backend-go
export DATABASE_URL="postgres://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres?connect_timeout=30"
go run cmd/server/main.go
```

---

## 📋 Recent Migrations Applied

| Migration | Description | Status |
|-----------|-------------|--------|
| 026_workspaces_and_roles.sql | Complete workspace system | ✅ Applied |
| 027_add_thinking_enabled_to_user_settings.sql | Thinking settings | ✅ Applied |

---

## 🔍 Quick Database Queries

### Check Tables
```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

### Count Workspaces
```sql
SELECT COUNT(*) FROM workspaces;
```

### List Your Workspaces
```sql
SELECT id, name, slug, plan_type, created_at
FROM workspaces
ORDER BY created_at DESC;
```

### Check Workspace Members
```sql
SELECT
  w.name as workspace_name,
  wm.user_id,
  wm.role,
  wm.status,
  wm.joined_at
FROM workspace_members wm
JOIN workspaces w ON w.id = wm.workspace_id
WHERE wm.status = 'active'
ORDER BY w.name, wm.joined_at;
```

---

## 🚀 Database Features Enabled

### Vector Embeddings (pgvector)
- ✅ Enabled for semantic search
- Used in: `workspace_memories`, `document_chunks`, `memories`
- Embedding dimension: 768 (Nomic) or 1536 (OpenAI)

### JSONB Support
- ✅ Enabled for flexible schema
- Used in: `settings`, `permissions`, `metadata`, `custom_settings`

### Full-Text Search
- ✅ Available for content search
- Can be used on `TEXT` columns

### Triggers & Functions
- ✅ Auto-update `updated_at` timestamps
- ✅ `seed_default_workspace_roles()` function

---

## 📊 Database Storage & Limits

**Current Plan**: (Check in Supabase Dashboard)

Typical Supabase limits:
- **Free Plan**: 500MB database, 2GB bandwidth
- **Pro Plan**: 8GB database, 50GB bandwidth
- **Team Plan**: Unlimited database, 250GB bandwidth

Check your current usage:
```
https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/settings/billing
```

---

## 🔒 Security Notes

### Connection Security
- ✅ SSL/TLS encryption enabled
- ✅ Row Level Security (RLS) available
- ✅ Connection pooling available

### Access Control
- Database password stored in `.env` file
- ⚠️ **NEVER commit `.env` to git** (already in `.gitignore`)
- Backend uses environment variables for security

---

## 🆘 Troubleshooting

### Connection Issues
If you get connection errors, check:
1. **Network**: Is Supabase accessible? (check status.supabase.com)
2. **Password**: Correct password in `.env`?
3. **Firewall**: No firewall blocking port 5432?

### Slow Queries
If database is slow:
1. Use connection pooler (port 6543) instead of direct connection
2. Add indexes to frequently queried columns
3. Check query performance in Supabase Dashboard

### Migration Issues
If migrations fail:
1. Check which migrations are applied:
   ```sql
   SELECT * FROM schema_migrations ORDER BY version;
   ```
2. Roll back if needed
3. Re-run migration scripts

---

## 📞 Support

- **Supabase Docs**: https://supabase.com/docs
- **Supabase Support**: https://supabase.com/dashboard/support
- **Community**: https://github.com/supabase/supabase/discussions

---

## ✅ Summary

```
┌─────────────────────────────────────────────────────────────────┐
│ YOUR DATABASE LOCATION                                          │
├─────────────────────────────────────────────────────────────────┤
│ Provider:   Supabase                                            │
│ Type:       PostgreSQL 15+                                      │
│ Host:       db.fuqhjbgbjamtxcdphjpp.supabase.co                 │
│ Database:   postgres                                            │
│ User:       postgres                                            │
│                                                                 │
│ Dashboard:  https://supabase.com/dashboard/project/...         │
│                                                                 │
│ Tables:     50+ tables                                          │
│ Migrations: 27 applied                                          │
│ Features:   Vector embeddings, JSONB, Full-text search         │
│ Status:     ✅ ONLINE & WORKING                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Your database is fully operational and accessible via Supabase!** 🚀
