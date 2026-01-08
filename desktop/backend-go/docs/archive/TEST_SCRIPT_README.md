# Comprehensive API Test Script

Complete test suite for all BusinessOS backend endpoints.

## Overview

The `test_all_endpoints.sh` script provides comprehensive testing of all critical API endpoints across the BusinessOS backend. It validates:

- Public health endpoints
- Chat system endpoints
- Context management
- Workspace & collaboration features
- Memory systems (CUS-25)
- RAG/search capabilities
- Multimodal search support
- Project & client management
- User settings & preferences
- Dashboard features
- Team management
- And more...

## Quick Start

### Prerequisites

```bash
# Install required tools
sudo apt-get install curl jq    # Linux
brew install curl jq             # macOS

# Ensure server is running
cd desktop/backend-go
go run ./cmd/server/main.go
```

### Basic Usage

```bash
# Run all tests (no auth required for public endpoints)
./test_all_endpoints.sh

# Run with specific server
./test_all_endpoints.sh http://localhost:8001

# Run with authentication token
./test_all_endpoints.sh http://localhost:8001 "Bearer YOUR_TOKEN_HERE"
```

## Usage Examples

### 1. Test Local Development Server (No Auth)

```bash
./test_all_endpoints.sh http://localhost:8001
```

This will:
- Test all public endpoints (health, readiness)
- Skip authenticated endpoints
- Show which endpoints require authentication

### 2. Test with Bearer Token

```bash
# Get your token from environment or auth system
TOKEN="Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

./test_all_endpoints.sh http://localhost:8001 "$TOKEN"
```

This will:
- Test all authenticated endpoints
- Create test workspace, project, client
- Test memory, search, and other features
- Display created resource IDs for further testing

### 3. Test Production Endpoint

```bash
./test_all_endpoints.sh https://api.yourdomain.com "Bearer PROD_TOKEN"
```

### 4. Quick Health Check

```bash
# Just check server is running
curl -s http://localhost:8001/health | jq .
```

## Test Coverage

### Public Endpoints (No Auth Required)

- `GET /health` - Server health check
- `GET /ready` - Readiness check with dependencies
- `GET /health/detailed` - Detailed component health

### Chat Endpoints

- `POST /api/chat/conversations` - Create conversation
- `GET /api/chat/conversations` - List conversations
- `GET /api/chat/conversations/:id` - Get specific conversation

### Context Endpoints

- `POST /api/contexts` - Create context
- `GET /api/contexts` - List contexts
- `GET /api/contexts/:id` - Get context details

### Workspace Endpoints

- `POST /api/workspaces` - Create workspace
- `GET /api/workspaces` - List workspaces
- `GET /api/workspaces/:id` - Get workspace details
- `GET /api/workspaces/:id/role-context` - Get user role & permissions
- `GET /api/workspaces/:id/members` - List workspace members

### Workspace Memory (CUS-25)

- `POST /api/workspaces/:id/memories` - Create workspace memory
- `GET /api/workspaces/:id/memories` - List workspace memories
- `GET /api/workspaces/:id/memories/private` - List private memories

### RAG Search Endpoints

- `POST /api/search/hybrid` - Hybrid search (semantic + keyword)
- `GET /api/search/explain?query=test` - Search scoring explanation

### Multimodal Search

- `GET /api/search/modalities` - Supported search modalities

### Projects

- `POST /api/projects` - Create project
- `GET /api/projects` - List projects
- `GET /api/projects/stats` - Project statistics

### Clients

- `POST /api/clients` - Create client
- `GET /api/clients` - List clients

### Search

- `GET /api/search/web?q=test` - Web search
- `GET /api/search/history` - Search history

### Settings

- `GET /api/settings` - User settings
- `GET /api/settings/system` - System settings
- `GET /api/settings/full-state` - Complete UI state

### Thinking/Reasoning

- `GET /api/thinking/traces/:conversationId` - COT traces
- `GET /api/thinking/settings` - Thinking settings
- `GET /api/reasoning/templates` - Reasoning templates

### Dashboard

- `GET /api/dashboard/summary` - Dashboard overview
- `GET /api/dashboard/focus` - Focus items
- `GET /api/dashboard/tasks` - Tasks

### Team

- `GET /api/team` - Team members

### Artifacts

- `GET /api/artifacts` - Artifacts

### Nodes

- `GET /api/nodes` - List nodes
- `GET /api/nodes/tree` - Node hierarchy

## Output Format

### Test Results

```
✓ PASS: Create Workspace (HTTP 200)
   Workspace ID: 550e8400-e29b-41d4-a716-446655440000

✗ FAIL: Hybrid Search (Expected 200, got 500)
   Response: {"error":"embedding service unavailable"}

⊘ SKIP: Create Memory (No auth token provided)
```

### Summary Report

```
╔════════════════════════════════════════════════════════════════╗
║ TEST SUMMARY                                                   ║
╚════════════════════════════════════════════════════════════════╝

Total Tests:   42
Passed:        41
Failed:        1
Skipped:       0

Pass Rate:     97%

Generated Test IDs:
  Workspace:     550e8400-e29b-41d4-a716-446655440000
  Conversation:  660f9511-f30c-52e5-b827-557766551111
  Context:       770g0622-g41d-63f6-c938-668877662222
  Project:       880h1733-h52e-74g7-d049-779988773333
  Client:        990i2844-i63f-85h8-e15a-88a099884444
```

## Interpreting Results

### Green (✓ PASS)

Endpoint responded with expected HTTP status code and valid JSON. The feature is working correctly.

```bash
✓ PASS: Create Workspace (HTTP 200)
```

### Red (✗ FAIL)

Endpoint either:
- Returned unexpected HTTP status code
- Returned invalid JSON
- Had an error in response body

```bash
✗ FAIL: Hybrid Search (Expected 200, got 500)
```

### Yellow (⊘ SKIP)

Endpoint was skipped because:
- Required authentication not provided
- Required dependencies not available
- Prerequisite test failed

```bash
⊘ SKIP: Create Memory (No auth token provided)
```

## Troubleshooting

### "Server is not responding"

```bash
# Start the server
cd desktop/backend-go
go run ./cmd/server/main.go

# Or run compiled binary
./bin/server.exe
```

### "No database connected"

```bash
# Check database connection
psql $DATABASE_URL -c "SELECT version();"

# View server logs
tail -f logs/server.log

# Check configuration
echo $DATABASE_URL
echo $REDIS_URL
```

### "Unauthorized" on authenticated endpoints

```bash
# Get a valid token from your auth system
# For development, you might create a test token

# Then run with token
./test_all_endpoints.sh http://localhost:8001 "Bearer YOUR_TOKEN"
```

### "Embedding service unavailable"

```bash
# Check if Ollama is running
ollama list

# Pull the embedding model
ollama pull nomic-embed-text

# Verify it's running
curl http://localhost:11434/api/tags
```

## Advanced Usage

### Test Specific Sections

You can modify the script to test specific sections:

```bash
# Edit script and comment out unwanted sections, then run
nano test_all_endpoints.sh
./test_all_endpoints.sh
```

### Extract Test IDs for Further Testing

```bash
# Run script and capture output
./test_all_endpoints.sh > test_results.log 2>&1

# Extract workspace ID
WORKSPACE_ID=$(grep "Workspace ID:" test_results.log | awk '{print $NF}')
echo $WORKSPACE_ID

# Use for manual testing
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/workspaces/$WORKSPACE_ID
```

### Continuous Integration

```bash
# Add to CI/CD pipeline
# .github/workflows/api-tests.yml

- name: Run API Tests
  run: |
    cd desktop/backend-go
    chmod +x test_all_endpoints.sh
    ./test_all_endpoints.sh http://localhost:8001 "${{ secrets.API_TOKEN }}"
  continue-on-error: true

- name: Check Results
  if: ${{ failure() }}
  run: echo "Some API tests failed - review logs"
```

## Performance Considerations

- **Execution Time**: ~30-60 seconds for full suite
- **Network**: Uses keep-alive connections
- **Rate Limiting**: Respects server rate limits (100-200 req/s)
- **Database**: Creates test records (can be cleaned up manually)

## Cleanup After Testing

Created test records are tagged with "API test" for easy cleanup:

```bash
# Find test records in database
SELECT * FROM workspaces WHERE description LIKE '%API test%';
SELECT * FROM projects WHERE description LIKE '%API test%';
SELECT * FROM clients WHERE name LIKE '%Test%';

# Delete if needed (be careful!)
DELETE FROM workspaces WHERE description LIKE '%API test%';
DELETE FROM projects WHERE description LIKE '%API test%';
```

## Environment Variables

The script respects these environment variables:

```bash
# Optional: Set before running
export BASE_URL="http://localhost:8001"
export AUTH_TOKEN="Bearer YOUR_TOKEN"

# Then run
./test_all_endpoints.sh
```

Or pass as arguments:

```bash
./test_all_endpoints.sh http://localhost:8001 "Bearer TOKEN"
```

## Integration Testing

### With Docker

```bash
# Run backend in container
docker run -p 8001:8001 businessos-backend:latest

# Run tests
./test_all_endpoints.sh http://localhost:8001
```

### With Database

```bash
# Ensure database is running
docker run -d -p 5432:5432 postgres:15

# Set connection string
export DATABASE_URL="postgresql://user:pass@localhost:5432/businessos"

# Run tests
./test_all_endpoints.sh
```

## Monitoring

### Real-time Monitoring

```bash
# In one terminal, watch logs
tail -f logs/server.log | grep -i "error\|warning"

# In another, run tests
./test_all_endpoints.sh http://localhost:8001
```

### Metrics Tracking

```bash
# Before tests
curl http://localhost:8001/health/detailed | jq '.components'

# After tests
curl http://localhost:8001/health/detailed | jq '.components'

# Compare database sizes
psql $DATABASE_URL -c "SELECT schemaname, SUM(heap_blks) FROM pg_statio_user_tables GROUP BY schemaname;"
```

## Support

For issues with the test script:

1. Check that all prerequisites are installed (curl, jq)
2. Verify server is running and accessible
3. Ensure correct authentication token
4. Check server logs for error messages
5. Review endpoint documentation for parameter requirements

## Related Documentation

- API Documentation: `/docs/api_rag_endpoints.md`
- Workspace Implementation: `/docs/feature1_workspace_status.md`
- Role-Based Access: `/docs/feature1_role_based_agents_status.md`
- RAG Features: `/docs/multimodal_search_integration.md`

---

**Last Updated**: 2026-01-06
**Script Version**: 1.0
**Maintained By**: BusinessOS Development Team
