# Workspace Integration Tests Runner (PowerShell)
# This script runs all workspace-related integration tests

$ErrorActionPreference = "Stop"

Write-Host "======================================================================"
Write-Host "Workspace Feature Integration Tests"
Write-Host "======================================================================"
Write-Host ""

# Check if database is available
Write-Host "Checking database connection..." -ForegroundColor Yellow
$dbUrl = $env:DATABASE_URL
if (-not $dbUrl) {
    $dbUrl = "postgresql://localhost:5432/businessos"
}

# Try to connect to database (you may need to adjust this based on your setup)
Write-Host "Database URL: $dbUrl" -ForegroundColor Cyan
Write-Host ""

# Run migrations
Write-Host "Running database migrations..." -ForegroundColor Yellow
go run cmd/migrations/main.go up
Write-Host "Migrations completed" -ForegroundColor Green
Write-Host ""

# Test categories
Write-Host "======================================================================"
Write-Host "Running Test Suites"
Write-Host "======================================================================"
Write-Host ""

# 1. Workspace Service Tests
Write-Host "1. Workspace Service Tests" -ForegroundColor Yellow
Write-Host "   Testing workspace creation, CRUD, roles, and members"
go test -v ./internal/services -run TestWorkspace -timeout 5m
Write-Host ""

# 2. Memory Hierarchy Tests
Write-Host "2. Memory Hierarchy Service Tests" -ForegroundColor Yellow
Write-Host "   Testing workspace vs private memory isolation"
go test -v ./internal/services -run TestMemoryHierarchy -timeout 5m
Write-Host ""

# 3. Workspace Handler Tests
Write-Host "3. Workspace Handler Tests" -ForegroundColor Yellow
Write-Host "   Testing HTTP API endpoints"
go test -v ./internal/handlers -run TestWorkspace -timeout 5m
go test -v ./internal/handlers -run TestAddMember -timeout 5m
go test -v ./internal/handlers -run TestUpdateMemberRole -timeout 5m
go test -v ./internal/handlers -run TestRemoveMember -timeout 5m
Write-Host ""

# 4. Role Context Tests
Write-Host "4. Role Context Tests" -ForegroundColor Yellow
Write-Host "   Testing permission checking and hierarchy"
go test -v ./internal/services -run TestRoleContext -timeout 5m
Write-Host ""

# 5. Access Control Tests
Write-Host "5. Memory Access Control Tests" -ForegroundColor Yellow
Write-Host "   Testing memory visibility and sharing"
go test -v ./internal/services -run TestMemoryAccessControl -timeout 5m
Write-Host ""

# Summary
Write-Host "======================================================================"
Write-Host "All Workspace Integration Tests Completed" -ForegroundColor Green
Write-Host "======================================================================"
Write-Host ""
Write-Host "Test Coverage:"
Write-Host "  ✓ Workspace creation with 6 default roles"
Write-Host "  ✓ Owner automatically added as first member"
Write-Host "  ✓ Member management (add, update role, remove)"
Write-Host "  ✓ Role hierarchy enforcement"
Write-Host "  ✓ Permission checking"
Write-Host "  ✓ Workspace vs private memory isolation"
Write-Host "  ✓ Memory sharing with specific users"
Write-Host "  ✓ Access control enforcement"
Write-Host "  ✓ HTTP API endpoints"
Write-Host "  ✓ Role context for agents"
Write-Host ""
Write-Host "Next Steps:"
Write-Host "  - Review test output for any failures"
Write-Host "  - Check test coverage: go test -cover ./internal/services ./internal/handlers"
Write-Host "  - Run specific test: go test -v -run TestSpecificName ./internal/services"
Write-Host ""
