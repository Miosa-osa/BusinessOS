# Workspace Audit Endpoints Test Script (PowerShell)
# This script tests all 6 audit log endpoints

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$WorkspaceId = "your-workspace-id",
    [string]$UserId = "your-user-id",
    [string]$Token = "your-auth-token"
)

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Workspace Audit Endpoints Test" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl"
Write-Host "Workspace ID: $WorkspaceId"
Write-Host ""

$headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

# Test 1: List Audit Logs (with pagination)
Write-Host "Test 1: List Audit Logs" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs?limit=10&offset=0" -Method Get -Headers $headers
    $response | ConvertTo-Json -Depth 10
    Write-Host "✓ Success" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 2: List Audit Logs (with filters)
Write-Host "Test 2: List Audit Logs (filtered by action)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs?action=create&limit=5" -Method Get -Headers $headers
    $response | ConvertTo-Json -Depth 10
    Write-Host "✓ Success" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 3: List Audit Logs (with date range)
$startDate = (Get-Date).AddDays(-30).ToString("yyyy-MM-ddTHH:mm:ssZ")
$endDate = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
Write-Host "Test 3: List Audit Logs (date range)" -ForegroundColor Yellow
Write-Host "Start: $startDate"
Write-Host "End: $endDate"
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs?start_date=$startDate&end_date=$endDate&limit=10" -Method Get -Headers $headers
    $response | ConvertTo-Json -Depth 10
    Write-Host "✓ Success" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 4: Get User Activity
Write-Host "Test 4: Get User Activity" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs/user/$UserId?limit=20" -Method Get -Headers $headers
    $response | ConvertTo-Json -Depth 10
    Write-Host "✓ Success" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 5: Get Resource History
Write-Host "Test 5: Get Resource History (workspace)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs/resource/workspace/$WorkspaceId?limit=10" -Method Get -Headers $headers
    $response | ConvertTo-Json -Depth 10
    Write-Host "✓ Success" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 6: Get Action Statistics
Write-Host "Test 6: Get Action Statistics" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs/stats/actions?start_date=$startDate&end_date=$endDate" -Method Get -Headers $headers
    $response | ConvertTo-Json -Depth 10
    Write-Host "✓ Success" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 7: Get Most Active Users
Write-Host "Test 7: Get Most Active Users (top 5)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs/stats/active-users?limit=5&start_date=$startDate&end_date=$endDate" -Method Get -Headers $headers
    $response | ConvertTo-Json -Depth 10
    Write-Host "✓ Success" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 8: Get Specific Audit Log
Write-Host "Test 8: Get Specific Audit Log" -ForegroundColor Yellow
try {
    $listResponse = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs?limit=1" -Method Get -Headers $headers
    if ($listResponse.logs -and $listResponse.logs.Count -gt 0) {
        $logId = $listResponse.logs[0].id
        Write-Host "Log ID: $logId"
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs/$logId" -Method Get -Headers $headers
        $response | ConvertTo-Json -Depth 10
        Write-Host "✓ Success" -ForegroundColor Green
    } else {
        Write-Host "Skipped (no audit logs found)" -ForegroundColor DarkYellow
    }
} catch {
    Write-Host "✗ Failed: $_" -ForegroundColor Red
}
Write-Host ""

# Test 9: Permission Check
Write-Host "Test 9: Permission Check (expect 403/401 without valid token)" -ForegroundColor Yellow
try {
    $invalidHeaders = @{
        "Authorization" = "Bearer invalid-token"
        "Content-Type" = "application/json"
    }
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/workspaces/$WorkspaceId/audit-logs" -Method Get -Headers $invalidHeaders -ErrorAction Stop
    Write-Host "✗ Unexpected success (should have failed)" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401 -or $_.Exception.Response.StatusCode -eq 403) {
        Write-Host "✓ Permission check working correctly (got 401/403)" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed with unexpected error: $_" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "All tests completed!" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan

# Usage example:
# .\test_audit_endpoints.ps1 -BaseUrl "http://localhost:8080" -WorkspaceId "abc-123" -UserId "user_xyz" -Token "your-jwt-token"
