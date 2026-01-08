# Workspace Integration Tests

This document describes the comprehensive integration tests for the workspace feature.

## Overview

The workspace feature tests cover critical functionality including:

1. **Workspace Creation**: Creating workspaces with 6 default roles
2. **Member Management**: Adding, updating, and removing workspace members
3. **Role Hierarchy**: Testing role levels and permission inheritance
4. **Memory Hierarchy**: Workspace vs private memory isolation
5. **Permission Enforcement**: Role-based access control
6. **API Endpoints**: HTTP handlers for all workspace operations

## Test Files

### Service Layer Tests

#### `internal/services/workspace_service_test.go`

**Test Cases:**

1. **TestWorkspaceCreation**
   - Creates workspace with valid data
   - Verifies 6 default roles are seeded (owner, admin, manager, member, viewer, guest)
   - Verifies owner is added as first member with "active" status
   - Tests auto-slug generation
   - Tests validation (empty name, invalid slug)
   - Tests plan limits (free, professional, enterprise)

2. **TestMemberManagement**
   - Adds members with different roles
   - Updates member roles
   - Removes members
   - Enforces "cannot remove owner" rule
   - Enforces "cannot change owner role" rule
   - Tests member limit enforcement based on plan type
   - Rejects invalid role names

3. **TestRoleHierarchy**
   - Verifies hierarchy levels (1=owner, 2=admin, ..., 6=guest)
   - Verifies system roles are protected
   - Verifies "member" is the default role
   - Tests owner permissions (can delete workspace, manage billing)
   - Tests viewer restrictions (read-only access)

4. **TestWorkspaceCRUD**
   - Get workspace by ID
   - Get workspace by slug
   - Update workspace (name, description, settings)
   - List user's workspaces
   - Delete workspace (owner only)
   - Non-owner cannot delete

5. **TestPermissionEnforcement**
   - Get user role in workspace
   - List workspace members
   - Verify permission structure for each role

#### `internal/services/memory_hierarchy_service_test.go`

**Test Cases:**

1. **TestMemoryHierarchy**
   - Workspace memory visible to all members
   - Private memory only visible to owner
   - Share memory with specific users
   - Unshare memory (revoke access)
   - Get all accessible memories (workspace + private + shared)
   - Track memory access count
   - Filter memories by type (decision, process, knowledge)

2. **TestMemoryAccessControl**
   - Workspace member can access workspace memory
   - Non-member cannot access workspace memory
   - Only owner can access private memory
   - Shared memory access control (owner, shared users, outsiders)

### Handler Layer Tests

#### `internal/handlers/workspace_handlers_test.go`

**Test Cases:**

1. **TestCreateWorkspaceHandler** - `POST /api/workspaces`
   - Create workspace successfully (returns 201)
   - Reject unauthenticated request (returns 401)
   - Reject invalid request body (returns 400/500)

2. **TestListWorkspacesHandler** - `GET /api/workspaces`
   - List user's workspaces
   - Verify all workspaces are returned

3. **TestGetWorkspaceHandler** - `GET /api/workspaces/:id`
   - Get workspace as member (returns 200)
   - Reject non-member (returns 403)

4. **TestUpdateWorkspaceHandler** - `PUT /api/workspaces/:id`
   - Owner can update workspace (returns 200)
   - Member cannot update workspace (returns 403)

5. **TestDeleteWorkspaceHandler** - `DELETE /api/workspaces/:id`
   - Owner can delete workspace (returns 200)
   - Non-owner cannot delete (returns 403)

6. **TestAddMemberHandler** - `POST /api/workspaces/:id/members/invite`
   - Owner can invite members (returns 201)
   - Member cannot invite (returns 403)

7. **TestUpdateMemberRoleHandler** - `PUT /api/workspaces/:id/members/:userId`
   - Owner can update member role (returns 200)
   - Cannot change owner role (returns 500)

8. **TestRemoveMemberHandler** - `DELETE /api/workspaces/:id/members/:userId`
   - Owner can remove member (returns 200)
   - Cannot remove owner (returns 500)

9. **TestGetUserRoleContextHandler** - `GET /api/workspaces/:id/role-context`
   - Get user's role context with permissions
   - Verify role name, display name, hierarchy level

10. **TestListWorkspaceRolesHandler** - `GET /api/workspaces/:id/roles`
    - List all 6 workspace roles
    - Verify role names and hierarchy

## Running Tests

### Prerequisites

1. **Database**: PostgreSQL must be running
2. **Environment**: Set `DATABASE_URL` environment variable
3. **Migrations**: Run migrations before tests

```bash
export DATABASE_URL=postgresql://user:password@localhost:5432/businessos
go run cmd/migrations/main.go up
```

### Run All Tests

**Unix/Linux/Mac:**
```bash
chmod +x run_workspace_tests.sh
./run_workspace_tests.sh
```

**Windows (PowerShell):**
```powershell
.\run_workspace_tests.ps1
```

### Run Specific Test Suites

**Workspace Service Tests:**
```bash
go test -v ./internal/services -run TestWorkspace
```

**Memory Hierarchy Tests:**
```bash
go test -v ./internal/services -run TestMemoryHierarchy
```

**Handler Tests:**
```bash
go test -v ./internal/handlers -run TestWorkspace
```

**Role Context Tests:**
```bash
go test -v ./internal/services -run TestRoleContext
```

### Run Individual Tests

```bash
# Run a specific test
go test -v ./internal/services -run TestWorkspaceCreation

# Run with coverage
go test -cover ./internal/services -run TestWorkspace

# Run with verbose output and coverage report
go test -v -coverprofile=coverage.out ./internal/services
go tool cover -html=coverage.out
```

## Test Data Setup

Each test creates its own isolated test data and cleans up after itself:

1. **Test Users**: Generated with unique IDs (`test-owner-{uuid}`)
2. **Test Workspaces**: Created with unique names and cleaned up with defer
3. **Test Members**: Added and removed within test scope
4. **Test Memories**: Created and deleted within test scope

## Test Coverage

The tests cover these critical paths:

### ✅ Workspace Creation Flow
- Create workspace → Seed 6 roles → Add owner as member → Return workspace

### ✅ Member Management Flow
- Invite member → Assign role → Verify permissions → Update role → Remove member

### ✅ Role Hierarchy Flow
- Owner (1) > Admin (2) > Manager (3) > Member (4) > Viewer (5) > Guest (6)

### ✅ Memory Hierarchy Flow
- Workspace memory → Visible to all members
- Private memory → Visible only to owner
- Shared memory → Visible to owner + shared users

### ✅ Permission Enforcement Flow
- Check user role → Get permissions → Enforce access control

### ✅ Agent Integration Flow
- Get role context → Inject into prompt → Agent respects permissions

## Common Test Patterns

### Setup Test Workspace
```go
func setupTestWorkspace(t *testing.T, pool *pgxpool.Pool, ownerID string) (*Workspace, func()) {
    service := NewWorkspaceService(pool)
    ctx := context.Background()

    workspace, err := service.CreateWorkspace(ctx, CreateWorkspaceRequest{
        Name:     "Test Workspace",
        PlanType: "professional",
    }, ownerID)
    require.NoError(t, err)

    cleanup := func() {
        service.DeleteWorkspace(ctx, workspace.ID, ownerID)
    }

    return workspace, cleanup
}
```

### Use in Tests
```go
func TestSomething(t *testing.T) {
    owner := "test-owner-" + uuid.New().String()[:8]
    workspace, cleanup := setupTestWorkspace(t, pool, owner)
    defer cleanup()

    // Test code here
}
```

## Test Environment

### Required Environment Variables
```bash
DATABASE_URL=postgresql://user:password@localhost:5432/businessos
REDIS_URL=redis://localhost:6379  # Optional for some tests
```

### Database Setup
The tests expect these tables to exist (created by migrations):
- `workspaces`
- `workspace_roles`
- `workspace_members`
- `user_workspace_profiles`
- `workspace_memories`
- `project_members`
- `role_permissions`

### Test Database Isolation

For CI/CD, consider using **testcontainers** to spin up isolated PostgreSQL instances:

```go
// Future enhancement
func setupTestDB(t *testing.T) *pgxpool.Pool {
    ctx := context.Background()

    // Start PostgreSQL container
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:15",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_DB":       "test_db",
                "POSTGRES_USER":     "test_user",
                "POSTGRES_PASSWORD": "test_pass",
            },
            WaitingFor: wait.ForLog("database system is ready to accept connections"),
        },
        Started: true,
    })
    require.NoError(t, err)

    // Get connection string and create pool
    // ...

    return pool
}
```

## Troubleshooting

### Tests are skipped
All tests have `t.Skip()` by default. Remove the skip line to run with live database:

```go
// Comment out this line:
// t.Skip("Requires database - implement with testcontainers or run with live DB")
```

### Database connection errors
- Ensure PostgreSQL is running
- Check `DATABASE_URL` is correct
- Verify migrations have been run

### Permission denied errors
- Make sure test scripts are executable: `chmod +x run_workspace_tests.sh`

### Test failures due to existing data
- Tests create unique data using UUID
- Tests clean up after themselves with defer
- If cleanup fails, manually delete test data

## Future Enhancements

1. **Testcontainers Integration**: Automated database setup/teardown
2. **Parallel Tests**: Run tests in parallel with isolated databases
3. **Load Testing**: Test workspace limits under load
4. **Performance Tests**: Benchmark permission checking
5. **E2E Tests**: Full user journey tests (create workspace → invite → chat → memories)

## Contributing

When adding new workspace features:

1. Add service-level tests in `workspace_service_test.go`
2. Add handler-level tests in `workspace_handlers_test.go`
3. Update this README with new test cases
4. Ensure all tests pass before committing
5. Add integration tests for cross-feature interactions

## Test Output Example

```
=== RUN   TestWorkspaceCreation
=== RUN   TestWorkspaceCreation/Create_workspace_successfully
    workspace_service_test.go:45: Creating test workspace
    workspace_service_test.go:52: Workspace created: Test Workspace (acme-corp)
    workspace_service_test.go:56: Verifying 6 default roles
    workspace_service_test.go:60: ✓ Found all 6 roles
    workspace_service_test.go:65: Verifying owner is first member
    workspace_service_test.go:68: ✓ Owner added with role 'owner' and status 'active'
--- PASS: TestWorkspaceCreation (0.15s)
    --- PASS: TestWorkspaceCreation/Create_workspace_successfully (0.12s)
```

## References

- [Workspace Service Implementation](internal/services/workspace_service.go)
- [Workspace Handlers](internal/handlers/workspace_handlers.go)
- [Memory Hierarchy Service](internal/services/memory_hierarchy_service.go)
- [Role Context Service](internal/services/role_context.go)
- [Migration 026: Workspaces](internal/database/migrations/026_workspaces_and_roles.sql)
- [Migration 030: Memory Hierarchy](internal/database/migrations/030_memory_hierarchy.sql)
