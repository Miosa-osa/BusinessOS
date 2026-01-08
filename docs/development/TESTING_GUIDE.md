# Testing Guide

**Version:** 1.0.0
**Last Updated:** January 6, 2026
**Target:** BusinessOS Backend Implementation

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Database Setup](#database-setup)
4. [Unit Tests](#unit-tests)
5. [Integration Tests](#integration-tests)
6. [API Testing](#api-testing)
7. [Performance Benchmarks](#performance-benchmarks)
8. [Manual Testing](#manual-testing)
9. [Test Data](#test-data)
10. [Troubleshooting](#troubleshooting)

---

## Overview

This guide provides comprehensive testing procedures for all implemented Linear issues (CUS-25, CUS-26, CUS-27, CUS-28, CUS-41).

### Test Coverage

- **Unit Tests:** Service-level logic testing
- **Integration Tests:** Database and API integration
- **Performance Benchmarks:** RAG system performance
- **Manual Tests:** End-to-end user workflows
- **API Tests:** HTTP endpoint validation

---

## Prerequisites

### Required Software

```bash
# Go 1.21 or higher
go version

# PostgreSQL 14 or higher
psql --version

# Redis (optional, for caching tests)
redis-cli --version

# Ollama (for embedding tests)
ollama --version

# CLIP server (for image embedding tests)
# Python 3.9+ with clip package
```

### Environment Variables

Create a `.env.test` file in the project root:

```bash
# Database
DATABASE_URL=postgres://postgres:password@localhost:5432/business_os_test

# Redis (optional)
REDIS_URL=localhost:6379
REDIS_PASSWORD=

# Embedding Services
OLLAMA_BASE_URL=http://localhost:11434
CLIP_SERVER_URL=http://localhost:8000

# Cache Configuration
EMBEDDING_CACHE_ENABLED=true
EMBEDDING_CACHE_TEXT_TTL=24h
EMBEDDING_CACHE_IMAGE_TTL=48h

# Test Configuration
TEST_WORKSPACE_ID=test-workspace-uuid
TEST_USER_ID=test-user-123
TEST_TOKEN=test-jwt-token
```

### Test Database Setup

```bash
# Create test database
createdb business_os_test

# Set connection
export DATABASE_URL=postgres://postgres:password@localhost:5432/business_os_test
```

---

## Database Setup

### Apply All Migrations

```bash
cd desktop/backend-go

# Apply migrations in order
psql $DATABASE_URL < internal/database/migrations/026_workspaces_and_roles.sql
psql $DATABASE_URL < internal/database/migrations/027_workspace_invites.sql
psql $DATABASE_URL < internal/database/migrations/028_workspace_audit_logs.sql
psql $DATABASE_URL < internal/database/migrations/029_project_members.sql
psql $DATABASE_URL < internal/database/migrations/030_memory_hierarchy_v2.sql
```

### Verify Migrations

```bash
# Check all tables exist
psql $DATABASE_URL -c "\dt"

# Expected tables:
# - workspaces
# - workspace_roles
# - workspace_members
# - workspace_invites
# - workspace_audit_logs
# - user_workspace_profiles
# - workspace_memories
# - project_members
# - project_role_definitions
```

### Verify Database Functions

```bash
# List all functions
psql $DATABASE_URL -c "\df"

# Expected functions:
# - can_access_memory
# - get_workspace_memories
# - get_user_memories
# - get_accessible_memories
# - share_memory
# - unshare_memory
# - track_memory_access
# - has_project_access
# - get_project_role
# - get_project_permissions
```

---

## Unit Tests

### Smart Chunking Service

**File:** `internal/services/smart_chunking_service_test.go`

**Run Tests:**
```bash
cd desktop/backend-go
go test -v ./internal/services -run TestSmartChunking
```

**Test Cases:**
- `TestChunkMarkdown` - Markdown document chunking
- `TestChunkCode` - Code file chunking (Go, Python, JavaScript)
- `TestChunkPlainText` - Plain text chunking
- `TestChunkStructured` - JSON/XML chunking
- `TestChunkOverlap` - Overlap strategy validation
- `TestChunkSizeValidation` - Size constraint validation
- `TestLanguageDetection` - Programming language detection

**Expected Output:**
```
=== RUN   TestChunkMarkdown
--- PASS: TestChunkMarkdown (0.01s)
=== RUN   TestChunkCode
--- PASS: TestChunkCode (0.02s)
...
PASS
ok      businessos/internal/services    0.156s
```

---

### Embedding Cache Service

**File:** `internal/services/embedding_cache_service_test.go`

**Run Tests:**
```bash
# Requires Redis running
redis-server &

# Run tests
go test -v ./internal/services -run TestEmbeddingCache
```

**Test Cases:**
- `TestCacheSetAndGet` - Basic cache operations
- `TestCacheMiss` - Cache miss handling
- `TestCacheExpiration` - TTL expiration
- `TestCacheInvalidation` - Manual invalidation
- `TestCacheStatistics` - Stats tracking
- `TestGracefulDegradation` - Behavior without Redis
- `TestFloat32Float64Adapter` - Type conversion
- `TestHealthCheck` - Connection health checking

**Expected Output:**
```
=== RUN   TestCacheSetAndGet
--- PASS: TestCacheSetAndGet (0.05s)
=== RUN   TestCacheMiss
--- PASS: TestCacheMiss (0.03s)
...
PASS
ok      businessos/internal/services    0.234s
```

**Without Redis:**
```
=== RUN   TestGracefulDegradation
--- PASS: TestGracefulDegradation (0.01s)
    cache_test.go:123: Cache disabled, graceful fallback working
```

---

### Memory Hierarchy Service

**Run Tests:**
```bash
go test -v ./internal/services -run TestMemoryHierarchy
```

**Test Cases:**
- Create workspace memory
- Create private memory
- Share private memory
- Unshare memory
- Access control validation
- Query filtering (type, category, tags)

---

### Role Context Service

**Run Tests:**
```bash
go test -v ./internal/services -run TestRoleContext
```

**Test Cases:**
- Get user role context
- Permission checking
- Role hierarchy validation
- Prompt generation

---

## Integration Tests

### Workspace API Integration

**File:** `test_workspace_api.go`

**Run Tests:**
```bash
cd desktop/backend-go

# Ensure database is clean
psql $DATABASE_URL -c "TRUNCATE workspaces, workspace_roles, workspace_members CASCADE;"

# Run tests
go run test_workspace_api.go
```

**Expected Output:**
```
Running Workspace API Integration Tests...
================================

Test 1: Create Workspace
✓ Workspace created successfully
  ID: 550e8400-e29b-41d4-a716-446655440000
  Name: Test Workspace
  Slug: test-workspace

Test 2: List Workspaces
✓ Found 1 workspace(s)

Test 3: Get Workspace Details
✓ Retrieved workspace details
  Members: 1
  Projects: 0

Test 4: Invite Member
✓ Invitation sent to newmember@example.com
  Token: abc123xyz

Test 5: Accept Invitation
✓ Invitation accepted successfully

Test 6: List Members
✓ Found 2 member(s)

Test 7: Update Member Role
✓ Role updated to 'admin'

Test 8: Create Custom Role
✓ Custom role created
  Name: custom-manager

Test 9: List Roles
✓ Found 7 role(s) (6 system + 1 custom)

Test 10: Update Workspace
✓ Workspace updated successfully

Test 11: Remove Member
✓ Member removed successfully

================================
All Tests Passed! (11/11)
```

---

### Invite and Audit System

**File:** `test_invite_audit_system.go`

**Run Tests:**
```bash
go run test_invite_audit_system.go
```

**Expected Output:**
```
Test 1: Create Workspace
✓ Workspace created

Test 2: Send Invitation
✓ Invitation sent

Test 3: List Pending Invites
✓ Found 1 pending invite(s)

Test 4: Accept Invitation
✓ Invitation accepted

Test 5: Verify Audit Logs
✓ Found 5 audit log entries:
  - create_workspace
  - invite_member
  - accept_invite
  - add_member
  - assign_role

Test 6: User-Specific Logs
✓ Found 3 actions by user

Test 7: Resource-Specific Logs
✓ Found 2 actions on workspace resource

Test 8: Action Statistics
✓ Action counts:
  create_workspace: 1
  invite_member: 1
  accept_invite: 1

Test 9: Active Users
✓ Found 2 active users

Test 10: Revoke Invitation
✓ New invitation sent
✓ Invitation revoked

Test 11: Expired Invitations
✓ Old invitation marked as expired

================================
All Tests Passed! (11/11)
```

---

### Embedding Cache Integration

**File:** `test_embedding_cache_integration.go`

**Run Tests:**
```bash
# Start Redis
redis-server &

# Run integration test
cd desktop/backend-go/internal/services
go run test_embedding_cache_integration.go
```

**Expected Output:**
```
Embedding Cache Integration Test
================================

1. Testing Cache Miss (First Request)
   Content: "machine learning algorithms"
   ✓ Cache miss (as expected)
   ✓ Embedding generated (1536 dimensions)
   ✓ Embedding cached
   Time: 523ms

2. Testing Cache Hit (Second Request)
   Content: "machine learning algorithms"
   ✓ Cache hit!
   ✓ Embedding retrieved (1536 dimensions)
   Time: 2ms

   Performance Improvement: 261x faster!

3. Testing Different Content
   Content: "neural network architectures"
   ✓ Cache miss (different content)
   Time: 498ms

4. Testing Cache Statistics
   ✓ Hits: 1
   ✓ Misses: 2
   ✓ Hit Rate: 33.33%

5. Testing Cache Invalidation
   ✓ Cache entry invalidated
   ✓ Next request is cache miss

6. Testing Graceful Degradation
   ✓ Stopping Redis...
   ✓ Cache disabled automatically
   ✓ Embedding generation still works
   Time: 512ms (no cache)

================================
All Integration Tests Passed!
```

---

## API Testing

### Manual API Testing with cURL

#### 1. Create Workspace

```bash
export TOKEN="your-jwt-token"
export BASE_URL="http://localhost:8080/api"

# Create workspace
curl -X POST $BASE_URL/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Workspace",
    "slug": "test-workspace",
    "description": "Testing workspace creation"
  }' | jq

# Save workspace ID
export WORKSPACE_ID="<workspace-id-from-response>"
```

#### 2. Create Workspace Memory

```bash
# Create workspace-level memory
curl -X POST $BASE_URL/workspaces/$WORKSPACE_ID/memories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "API Best Practices",
    "summary": "Guidelines for API design",
    "content": "Always use RESTful principles...",
    "memory_type": "pattern",
    "visibility": "workspace",
    "tags": ["api", "best-practices"]
  }' | jq
```

#### 3. Create Private Memory

```bash
# Create private memory
curl -X POST $BASE_URL/workspaces/$WORKSPACE_ID/memories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Personal Notes",
    "summary": "My thoughts on the project",
    "content": "I think we should...",
    "memory_type": "general",
    "visibility": "private",
    "tags": ["personal"]
  }' | jq

# Save memory ID
export MEMORY_ID="<memory-id-from-response>"
```

#### 4. Share Memory

```bash
# Get another user's ID
export USER2_ID="user-456"

# Share private memory
curl -X POST $BASE_URL/workspaces/$WORKSPACE_ID/memories/$MEMORY_ID/share \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_ids": ["'$USER2_ID'"]
  }' | jq
```

#### 5. List All Accessible Memories

```bash
# Get all memories accessible to user
curl -X GET "$BASE_URL/workspaces/$WORKSPACE_ID/memories/accessible?limit=50" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 6. Test Project Access Control

```bash
export PROJECT_ID="<your-project-id>"

# Add member to project
curl -X POST $BASE_URL/projects/$PROJECT_ID/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-789",
    "role": "contributor"
  }' | jq

# Check user's access
curl -X GET $BASE_URL/projects/$PROJECT_ID/access/user-789 \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 7. Test Hybrid Search

```bash
# Hybrid search
curl -X POST $BASE_URL/rag/search/hybrid \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "authentication security",
    "semantic_weight": 0.7,
    "keyword_weight": 0.3,
    "max_results": 10
  }' | jq
```

#### 8. Test Image Upload

```bash
# Upload image
curl -X POST $BASE_URL/images/upload-file \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/diagram.png" \
  -F "description=System architecture" | jq
```

---

### Automated API Tests with Test Scripts

**Create:** `test_api_endpoints.sh`

```bash
#!/bin/bash

# Configuration
BASE_URL="http://localhost:8080/api"
TOKEN="your-jwt-token"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0

# Test function
test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_code=$5

    TESTS_RUN=$((TESTS_RUN + 1))

    echo -n "Test $TESTS_RUN: $name... "

    if [ -z "$data" ]; then
        response=$(curl -s -w "%{http_code}" -X $method \
            -H "Authorization: Bearer $TOKEN" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "%{http_code}" -X $method \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi

    status_code="${response: -3}"

    if [ "$status_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}PASS${NC} ($status_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAIL${NC} (expected $expected_code, got $status_code)"
    fi
}

echo "Running API Endpoint Tests"
echo "=========================="
echo ""

# Test workspace creation
test_endpoint "Create Workspace" POST "/workspaces" \
    '{"name":"Test Workspace","slug":"test-ws"}' 201

# Test workspace list
test_endpoint "List Workspaces" GET "/workspaces" "" 200

# Test invalid endpoint
test_endpoint "Invalid Endpoint" GET "/invalid" "" 404

# Test unauthorized access
TOKEN="" test_endpoint "Unauthorized Access" GET "/workspaces" "" 401

echo ""
echo "=========================="
echo "Tests Run: $TESTS_RUN"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $((TESTS_RUN - TESTS_PASSED))"
```

**Run:**
```bash
chmod +x test_api_endpoints.sh
./test_api_endpoints.sh
```

---

## Performance Benchmarks

### RAG System Benchmarks

**File:** `internal/services/rag_benchmarks_test.go`

**Run All Benchmarks:**
```bash
cd desktop/backend-go

# Run all benchmarks with memory stats
go test -bench=. -benchmem ./internal/services/

# Run specific benchmark
go test -bench=BenchmarkCacheHit -benchmem ./internal/services/

# Run with custom duration
go test -bench=. -benchtime=10s ./internal/services/
```

**Using Benchmark Scripts:**

```bash
# Linux/Mac
./run_rag_benchmarks.sh

# Windows
.\run_rag_benchmarks.ps1

# Quick benchmark (specific component)
./quick_benchmark.sh cache
./quick_benchmark.sh chunking
./quick_benchmark.sh search
```

**Expected Output:**
```
Running RAG Performance Benchmarks...

BenchmarkTextEmbeddingSmall-8           100    523421 ns/op    4096 B/op    12 allocs/op
BenchmarkTextEmbeddingMedium-8           50   1045678 ns/op    8192 B/op    24 allocs/op
BenchmarkTextEmbeddingLarge-8            20   2134567 ns/op   16384 B/op    48 allocs/op

BenchmarkCacheHit-8                 500000      2345 ns/op     512 B/op     3 allocs/op
BenchmarkCacheMiss-8                    50    523421 ns/op    4096 B/op    12 allocs/op

BenchmarkHybridSearch-8                 20    567890 ns/op   32768 B/op    89 allocs/op

BenchmarkChunkMarkdown-8             10000    134567 ns/op    8192 B/op    45 allocs/op
BenchmarkChunkCode-8                  5000    223456 ns/op   12288 B/op    67 allocs/op

Performance Summary:
- Cache hit: 223x faster than cache miss
- Hybrid search: ~568ms average
- Smart chunking: ~135ms for typical document
```

**Interpret Results:**
- `ns/op` - Nanoseconds per operation (lower is better)
- `B/op` - Bytes allocated per operation (lower is better)
- `allocs/op` - Memory allocations per operation (lower is better)

**Performance Targets:**
- Cache hit: < 5ms
- Text embedding (cached): < 5ms
- Text embedding (uncached): < 600ms
- Hybrid search: < 1000ms
- Smart chunking: < 200ms per document

---

## Manual Testing

### Complete Workflow Tests

#### Test 1: Team Collaboration Workflow

**Scenario:** Create a workspace, invite members, assign roles, create shared knowledge

**Steps:**

1. **Create Workspace**
   ```bash
   # Use API or frontend
   POST /api/workspaces
   ```

2. **Invite Team Members**
   ```bash
   POST /api/workspaces/:id/members/invite
   # Email: alice@example.com, Role: Manager
   # Email: bob@example.com, Role: Member
   ```

3. **Members Accept Invitations**
   ```bash
   POST /api/workspaces/invites/accept
   # Use tokens from invitation emails
   ```

4. **Create Workspace Memory** (as owner)
   ```bash
   POST /api/workspaces/:id/memories
   # visibility: "workspace"
   ```

5. **Verify All Members Can Access** (as alice or bob)
   ```bash
   GET /api/workspaces/:id/memories
   # Should see the workspace memory
   ```

6. **Create Private Memory** (as alice)
   ```bash
   POST /api/workspaces/:id/memories
   # visibility: "private"
   ```

7. **Verify Owner Cannot See Alice's Private Memory**
   ```bash
   GET /api/workspaces/:id/memories/private
   # As owner, should NOT see alice's private memory
   ```

8. **Share Private Memory** (as alice)
   ```bash
   POST /api/workspaces/:id/memories/:memoryId/share
   # Share with owner
   ```

9. **Verify Owner Can Now See Shared Memory**
   ```bash
   GET /api/workspaces/:id/memories/accessible
   # Should now see alice's shared memory
   ```

10. **Check Audit Logs**
    ```bash
    GET /api/workspaces/:id/audit-logs
    # Should show all actions
    ```

**Expected Result:** Complete audit trail of all actions, proper access control throughout

---

#### Test 2: Project-Level Access Control

**Scenario:** Add users to projects with different roles

**Steps:**

1. **Create Project** (requires existing project)
   ```bash
   # Use your project creation endpoint
   ```

2. **Add User as Project Lead**
   ```bash
   POST /api/projects/:id/members
   # user_id: alice, role: lead
   ```

3. **Verify Lead Permissions**
   ```bash
   GET /api/projects/:id/access/alice
   # can_edit: true, can_delete: true, can_invite: true
   ```

4. **Add User as Contributor**
   ```bash
   POST /api/projects/:id/members
   # user_id: bob, role: contributor
   ```

5. **Verify Contributor Permissions**
   ```bash
   GET /api/projects/:id/access/bob
   # can_edit: true, can_delete: false, can_invite: false
   ```

6. **Try to Delete as Contributor** (should fail)
   ```bash
   # As bob, attempt to delete project
   # Should get 403 Forbidden
   ```

7. **Update Role to Reviewer**
   ```bash
   PUT /api/projects/:id/members/:memberId/role
   # role: reviewer
   ```

8. **Verify Updated Permissions**
   ```bash
   GET /api/projects/:id/access/bob
   # can_edit: false, can_delete: false, can_invite: false
   ```

**Expected Result:** Permissions correctly enforced at each role level

---

#### Test 3: Role-Based Agent Context

**Scenario:** Verify agent receives and respects role context

**Steps:**

1. **Chat as Viewer**
   ```bash
   POST /api/chat/message
   # workspace_id: test-workspace
   # message: "Create a new project"
   ```

   **Expected Response:** Agent explains viewer role cannot create projects

2. **Chat as Manager**
   ```bash
   POST /api/chat/message
   # workspace_id: test-workspace
   # message: "Create a new project"
   ```

   **Expected Response:** Agent offers to help create project

3. **Chat as Member**
   ```bash
   POST /api/chat/message
   # workspace_id: test-workspace
   # message: "Delete the workspace"
   ```

   **Expected Response:** Agent explains member role cannot delete workspace

4. **Check Agent Logs**
   ```bash
   # In server logs, should see:
   # [ChatV2] Injected role context: manager (level 4, 8 permissions)
   ```

**Expected Result:** Agent behavior matches user's role and permissions

---

#### Test 4: RAG Search Quality

**Scenario:** Test search accuracy and performance

**Steps:**

1. **Upload Test Documents**
   ```bash
   # Create several memories with distinct content
   # Topic 1: Authentication
   # Topic 2: Database design
   # Topic 3: Frontend development
   ```

2. **Test Semantic Search**
   ```bash
   POST /api/rag/search/hybrid
   # query: "user authentication"
   # semantic_weight: 1.0, keyword_weight: 0.0
   ```

   **Verify:** Authentication-related results ranked highest

3. **Test Keyword Search**
   ```bash
   POST /api/rag/search/hybrid
   # query: "JWT token"
   # semantic_weight: 0.0, keyword_weight: 1.0
   ```

   **Verify:** Documents containing "JWT" ranked highest

4. **Test Hybrid Search**
   ```bash
   POST /api/rag/search/hybrid
   # query: "secure login implementation"
   # semantic_weight: 0.7, keyword_weight: 0.3
   ```

   **Verify:** Balanced results from both strategies

5. **Test Agentic RAG**
   ```bash
   POST /api/rag/retrieve
   # query: "How do I implement authentication?"
   ```

   **Verify:** Intent classified as "procedural", quality score > 0.6

6. **Check Cache Statistics**
   ```bash
   # After multiple searches with same queries
   GET /api/rag/cache/stats
   ```

   **Verify:** Hit rate increasing over time

**Expected Result:** Relevant results, good quality scores, cache working

---

#### Test 5: Multi-Modal Search

**Scenario:** Upload images and search across modalities

**Steps:**

1. **Upload Diagram**
   ```bash
   POST /api/images/upload-file
   # File: architecture-diagram.png
   # Description: "System architecture overview"
   ```

2. **Upload Screenshots**
   ```bash
   # Upload 3-5 different screenshots
   # Vary descriptions
   ```

3. **Text-to-Image Search**
   ```bash
   POST /api/search/images-by-text
   # query: "architecture diagram"
   ```

   **Verify:** Architecture diagram ranked first

4. **Image-to-Image Search**
   ```bash
   POST /api/search/similar-images
   # image_id: <architecture diagram ID>
   ```

   **Verify:** Similar diagrams/screenshots ranked high

5. **Combined Search**
   ```bash
   POST /api/search/multimodal
   # text_query: "authentication flow"
   # image_id: <reference image>
   ```

   **Verify:** Results combine text and image similarity

**Expected Result:** Cross-modal search works, relevant images retrieved

---

## Test Data

### Sample Workspaces

```sql
-- Insert test workspace
INSERT INTO workspaces (id, name, slug, description, owner_id, plan_type)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Test Workspace', 'test-ws', 'Test workspace for development', 'test-user-123', 'professional');

-- Insert test members
INSERT INTO workspace_members (workspace_id, user_id, role_name, status)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'test-user-123', 'owner', 'active'),
    ('550e8400-e29b-41d4-a716-446655440000', 'test-user-456', 'admin', 'active'),
    ('550e8400-e29b-41d4-a716-446655440000', 'test-user-789', 'member', 'active');
```

### Sample Memories

```sql
-- Workspace memory
INSERT INTO workspace_memories (workspace_id, title, summary, content, memory_type, visibility, created_by)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000',
     'Authentication Best Practices',
     'Guidelines for secure authentication',
     'Always use JWT tokens with proper expiration...',
     'pattern',
     'workspace',
     'test-user-123');

-- Private memory
INSERT INTO workspace_memories (workspace_id, title, summary, content, memory_type, visibility, owner_user_id, created_by)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000',
     'Personal Notes',
     'My thoughts on the project',
     'I think we should refactor the auth module...',
     'general',
     'private',
     'test-user-456',
     'test-user-456');
```

### Cleanup Test Data

```bash
# Clean all test data
psql $DATABASE_URL -c "
DELETE FROM workspace_memories WHERE workspace_id IN (SELECT id FROM workspaces WHERE slug LIKE 'test-%');
DELETE FROM workspace_members WHERE workspace_id IN (SELECT id FROM workspaces WHERE slug LIKE 'test-%');
DELETE FROM workspaces WHERE slug LIKE 'test-%';
"
```

---

## Troubleshooting

### Common Issues

#### 1. Migrations Fail to Apply

**Error:** `relation already exists`

**Solution:**
```bash
# Check which migrations are applied
psql $DATABASE_URL -c "SELECT * FROM schema_migrations;"

# Drop and recreate test database
dropdb business_os_test
createdb business_os_test

# Reapply all migrations
./apply_migrations.sh
```

---

#### 2. Tests Fail with "Connection Refused"

**Error:** `dial tcp [::1]:5432: connect: connection refused`

**Solution:**
```bash
# Check PostgreSQL is running
pg_isready

# Start PostgreSQL
# Mac:
brew services start postgresql

# Linux:
sudo systemctl start postgresql

# Verify connection
psql -h localhost -U postgres -c "SELECT 1;"
```

---

#### 3. Redis Tests Fail

**Error:** `redis: connection refused`

**Solution:**
```bash
# Check Redis is running
redis-cli ping

# Start Redis
# Mac:
brew services start redis

# Linux:
sudo systemctl start redis

# Test connection
redis-cli ping
# Should return: PONG
```

---

#### 4. Embedding Tests Fail

**Error:** `Ollama connection failed`

**Solution:**
```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# Start Ollama
ollama serve

# Pull required model
ollama pull nomic-embed-text
```

---

#### 5. Slow Test Performance

**Issue:** Tests taking too long

**Solution:**
```bash
# Use test-specific lightweight config
# Reduce chunk sizes
# Use smaller embedding models for tests
# Mock external services

# Run tests in parallel
go test -parallel 4 ./internal/services/
```

---

#### 6. Permission Denied Errors

**Error:** `403 Forbidden`

**Solution:**
```bash
# Check token is valid
# Decode JWT token
echo "your-token" | cut -d'.' -f2 | base64 -d | jq

# Verify user is workspace member
psql $DATABASE_URL -c "
SELECT * FROM workspace_members
WHERE user_id = 'your-user-id'
AND workspace_id = 'workspace-id';
"

# Check role permissions
psql $DATABASE_URL -c "
SELECT permissions FROM workspace_roles
WHERE name = 'member';
"
```

---

## CI/CD Integration

### GitHub Actions Workflow

Create `.github/workflows/test.yml`:

```yaml
name: Run Tests

on:
  push:
    branches: [ main, dev ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: business_os_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Apply Migrations
      run: |
        cd desktop/backend-go
        for f in internal/database/migrations/*.sql; do
          psql -h localhost -U postgres -d business_os_test -f $f
        done
      env:
        PGPASSWORD: postgres

    - name: Run Unit Tests
      run: |
        cd desktop/backend-go
        go test -v ./internal/services/
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/business_os_test
        REDIS_URL: localhost:6379

    - name: Run Integration Tests
      run: |
        cd desktop/backend-go
        go run test_workspace_api.go
        go run test_invite_audit_system.go
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/business_os_test

    - name: Run Benchmarks
      run: |
        cd desktop/backend-go
        go test -bench=. -benchmem ./internal/services/ > benchmark_results.txt
        cat benchmark_results.txt
```

---

## Test Checklist

Use this checklist before marking features as complete:

### Memory Hierarchy
- [ ] Workspace memories visible to all members
- [ ] Private memories only visible to owner
- [ ] Shared memories visible to owner + shared users
- [ ] Access control functions work correctly
- [ ] Memory tracking (access count) works
- [ ] All database functions tested
- [ ] API endpoints return correct data
- [ ] Permission checks enforced

### Role-Based Agent Context
- [ ] Role context injected in chat handlers
- [ ] Different roles produce different agent behavior
- [ ] Permission-aware suggestions work
- [ ] Role context prompt generated correctly
- [ ] Chat works without workspace_id (no errors)
- [ ] Logging shows role context injection

### Project Access Control
- [ ] Four roles (lead, contributor, reviewer, viewer) work
- [ ] Permissions (can_edit, can_delete, can_invite) enforced
- [ ] Role updates work correctly
- [ ] Member removal works
- [ ] Access check endpoint works
- [ ] Audit logging captures all changes

### RAG System
- [ ] Smart chunking produces sensible chunks
- [ ] Embedding cache hits/misses work
- [ ] Cache statistics accurate
- [ ] Graceful degradation without Redis
- [ ] Hybrid search returns relevant results
- [ ] Re-ranking improves result quality
- [ ] Agentic RAG classifies intent correctly
- [ ] Performance benchmarks meet targets

### Multi-Modal Search
- [ ] Image upload (base64) works
- [ ] Image upload (multipart) works
- [ ] Text-to-image search works
- [ ] Image-to-image search works
- [ ] Combined search works
- [ ] CLIP embeddings generated
- [ ] Image cache works

---

**Document Version:** 1.0.0
**Last Updated:** January 6, 2026
**Maintainer:** Development Team
