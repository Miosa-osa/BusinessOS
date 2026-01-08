# Database Setup Guide

Complete guide for setting up PostgreSQL database for BusinessOS.

## Table of Contents
- [Quick Start](#quick-start)
- [Database Schema](#database-schema)
- [Migrations](#migrations)
- [Test Data](#test-data)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## Quick Start

### 1. Install PostgreSQL

**Windows:**
```powershell
# Download from https://www.postgresql.org/download/windows/
# Or use Chocolatey
choco install postgresql
```

**macOS:**
```bash
brew install postgresql@18
brew services start postgresql@18
```

**Linux:**
```bash
sudo apt-get install postgresql-18
sudo systemctl start postgresql
```

### 2. Create Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database (or use default 'postgres')
CREATE DATABASE businessos;

# Exit
\q
```

### 3. Apply Migrations

**Windows (Recommended):**
```powershell
cd desktop/backend-go/scripts
.\apply-migrations.ps1
```

**Cross-platform:**
```bash
# Set password (Linux/macOS)
export PGPASSWORD='your_password'

# Apply migrations
psql -U postgres -d postgres -f supabase-migrations-combined.sql

# Unset password
unset PGPASSWORD
```

## Database Schema

### Core Tables (26 total)

#### Memory System
- `memories` - Semantic memory with 768D embeddings
- `memory_associations` - Relationships between memories
- `user_facts` - Extracted user facts
- `memory_access_log` - Memory usage tracking

#### Document System
- `uploaded_documents` - File metadata and storage
- `document_chunks` - Chunked content with embeddings
- `document_references` - Cross-document references

#### Context & Conversation
- `conversation_summaries` - AI chat history summaries
- `context_profiles` - Context management rules
- `context_profile_items` - Profile configuration items
- `context_loading_rules` - Dynamic loading strategies
- `context_retrieval_log` - Context usage analytics

#### Learning & Intelligence
- `learning_events` - User feedback tracking
- `user_behavior_patterns` - Behavioral analysis
- `personalization_profiles` - User preferences
- `feedback_log` - System feedback

#### Application Profiles
- `application_profiles` - App-specific configurations
- `application_components` - Component registry
- `application_api_endpoints` - API definitions

#### Consultation System
- `agent_context_sessions` - Agent session tracking
- `consultation_sessions` - Consultation metadata
- `consultation_messages` - Message history
- `consultation_contexts` - Context snapshots

#### Output & Formatting
- `output_styles` - Output formatting templates
- `user_output_preferences` - User-specific preferences

#### System
- `schema_migrations` - Migration version tracking

### Key Features

**Vector Embeddings (768D):**
- Uses pgvector extension
- HNSW indexes for fast similarity search
- Cosine distance for semantic matching

**Indexes:**
```sql
-- HNSW vector indexes
CREATE INDEX idx_memories_embedding ON memories USING hnsw (embedding vector_cosine_ops);
CREATE INDEX idx_doc_chunks_embedding ON document_chunks USING hnsw (embedding vector_cosine_ops);

-- B-tree indexes for fast lookups
CREATE INDEX idx_memories_user ON memories(user_id);
CREATE INDEX idx_memories_type ON memories(memory_type);
CREATE INDEX idx_doc_chunks_document ON document_chunks(document_id);
```

## Migrations

### Migration Files

Located in `desktop/backend-go/internal/database/migrations/`:

1. `016_memories.sql` - Memory system foundation
2. `017_context_system.sql` - Context management
3. `018_output_styles.sql` - Output formatting
4. `019_documents_no_vector.sql` - Document tables
5. `020_context_integration_no_vector.sql` - Context integration
6. `021_learning_system.sql` - Learning & feedback
7. `022_application_profiles_no_vector.sql` - App profiles
8. `023_pedro_tasks_schema_fix.sql` - Schema fixes
9. `024_embedding_dimensions_768.sql` - Update to 768D embeddings

### Combined Migration File

All migrations combined in: `supabase-migrations-combined.sql` (~65 KB)

### Verification After Migration

```sql
-- Check tables created
\dt

-- Should show 26 tables
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE';

-- Check pgvector extension
SELECT * FROM pg_extension WHERE extname = 'vector';

-- Check HNSW indexes
SELECT tablename, indexname FROM pg_indexes
WHERE schemaname = 'public'
AND indexname LIKE '%embedding%';
```

**Expected Output:**
```
 tablename       | indexname
-----------------+----------------------------
 memories        | idx_memories_embedding
 document_chunks | idx_doc_chunks_embedding
```

## Test Data

### Setup Test User

**Script:** `test-user-setup.sql`

Creates test user for API authentication:
- Email: `testuser@businessos.dev`
- Session Token: `test-token-businessos-123`
- Valid for 30 days

**Run:**
```powershell
# Windows
.\run-test-setup.ps1

# Or manually
psql -U postgres -d postgres -f test-user-setup.sql
```

### Using Test Credentials

**cURL:**
```bash
curl -H "Cookie: better-auth.session_token=test-token-businessos-123" \
  http://localhost:8001/api/memories
```

**JavaScript:**
```javascript
fetch('http://localhost:8001/api/chat/message', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Cookie': 'better-auth.session_token=test-token-businessos-123'
  },
  body: JSON.stringify({ message: 'Hello' })
});
```

## Verification

### Check Database Health

```sql
-- Connection test
SELECT version();

-- Table sizes
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Test Embeddings

```sql
-- Count memories with embeddings
SELECT COUNT(*) as memories_with_embeddings
FROM memories
WHERE embedding IS NOT NULL;

-- Count document chunks with embeddings
SELECT COUNT(*) as chunks_with_embeddings
FROM document_chunks
WHERE embedding IS NOT NULL;

-- Test vector dimensions
SELECT
    id,
    title,
    vector_dims(embedding) as dimensions
FROM memories
WHERE embedding IS NOT NULL
LIMIT 1;
-- Should return: dimensions = 768
```

### Backend Health Check

```bash
# Check backend database connection
curl http://localhost:8001/health/detailed
```

**Expected Response:**
```json
{
  "status": "healthy",
  "instance_id": "...",
  "components": {
    "database": {"status": "connected"},
    "containers": {"status": "unavailable"},
    "redis": {"status": "not_configured"}
  }
}
```

## Verified Features (Jan 2, 2026)

✅ **Conversation System:**
- SSE streaming responses
- Message history tracking
- Real-time AI chat

✅ **Memory System:**
- Semantic memory storage
- 768D vector embeddings
- PostgreSQL array tags

✅ **Document Processing:**
- File upload (text, PDF support)
- Automatic chunking
- Embedding generation
- 3 chunks created from 657-byte test file

✅ **Database Performance:**
- HNSW indexes for fast vector search
- B-tree indexes for user/type lookups
- Connection pooling ready

## Troubleshooting

### Connection Refused

```bash
# Check PostgreSQL is running
# Windows
Get-Service postgresql*

# Start if stopped
net start postgresql-x64-18
```

### Authentication Failed

```bash
# Edit pg_hba.conf to allow password auth
# Windows: C:\Program Files\PostgreSQL\18\data\pg_hba.conf
# Add line:
host    all             all             127.0.0.1/32            md5

# Reload configuration
pg_ctl reload -D "C:\Program Files\PostgreSQL\18\data"
```

### Migration Errors

**"Extension vector does not exist":**
```sql
CREATE EXTENSION vector;
```

**"Column already exists":**
- Migrations are idempotent
- Safe to re-run
- Check applied migrations:
```sql
SELECT * FROM schema_migrations ORDER BY applied_at DESC;
```

### Performance Issues

**Slow Vector Search:**
```sql
-- Verify HNSW index usage
EXPLAIN ANALYZE
SELECT * FROM memories
ORDER BY embedding <=> '[...]'::vector
LIMIT 10;
-- Should show: "Index Scan using idx_memories_embedding"
```

**Out of Memory During Index Build:**
```sql
-- Increase work memory temporarily
SET maintenance_work_mem = '1GB';
REINDEX INDEX idx_memories_embedding;
```

## Backup & Restore

### Backup

```bash
# Full database
pg_dump -U postgres -d postgres -F c -f backup.dump

# Schema only
pg_dump -U postgres -d postgres --schema-only -f schema.sql

# Data only
pg_dump -U postgres -d postgres --data-only -f data.sql
```

### Restore

```bash
# From dump file
pg_restore -U postgres -d postgres -c backup.dump

# From SQL file
psql -U postgres -d postgres -f backup.sql
```

## Environment Variables

### Backend Configuration

**Local Development:**
```bash
DATABASE_URL=postgres://postgres:password@localhost:5432/postgres?sslmode=disable
DATABASE_REQUIRED=true
```

**Supabase (Production):**
```bash
DATABASE_URL=postgres://postgres.PROJECT_ID:PASSWORD@aws-0-us-east-1.pooler.supabase.com:6543/postgres
SUPABASE_URL=https://PROJECT_ID.supabase.co
SUPABASE_ANON_KEY=your_anon_key
```

## Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [pgvector GitHub](https://github.com/pgvector/pgvector)
- [HNSW Algorithm](https://github.com/pgvector/pgvector#hnsw)
- [Supabase Database Guides](https://supabase.com/docs/guides/database)

---

**Last Updated:** January 2, 2026
**PostgreSQL Version:** 17/18
**Schema Version:** 24 migrations
**Total Tables:** 26
**Test Status:** All features verified ✅
