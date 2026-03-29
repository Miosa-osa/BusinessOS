# Sandbox/Deployment API Test Script
# Tests all OSA deployment endpoints for the BusinessOS system
# 
# Backend routes (osa_deployment.go):
#   POST /api/osa/deployment/:app_id/deploy
#   POST /api/osa/deployment/:app_id/stop
#   GET  /api/osa/deployment/:app_id/status
#   GET  /api/osa/deployments
#
# Run: .\test_sandbox_endpoints.ps1 -SessionCookie "your_session_cookie"

param(
    [Parameter(Mandatory=$false)]
    [string]$SessionCookie = "",
    
    [Parameter(Mandatory=$false)]
    [string]$BaseUrl = "http://localhost:8001",
    
    [Parameter(Mandatory=$false)]
    [string]$AppId = ""
)

$ErrorActionPreference = "Continue"

# Colors for output
function Write-Success { param($msg) Write-Host "[PASS] $msg" -ForegroundColor Green }
function Write-Fail { param($msg) Write-Host "[FAIL] $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Test { param($msg) Write-Host "`n=== $msg ===" -ForegroundColor Yellow }

# Test results tracking
$script:passed = 0
$script:failed = 0
$script:skipped = 0

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Endpoint,
        [int[]]$ExpectedStatus = @(200, 201),
        [string]$Body = "",
        [switch]$SkipIfNoAuth
    )
    
    Write-Info "Testing: $Name"
    Write-Info "  $Method $Endpoint"
    
    if ($SkipIfNoAuth -and [string]::IsNullOrEmpty($SessionCookie)) {
        Write-Host "[SKIP] No session cookie provided" -ForegroundColor DarkYellow
        $script:skipped++
        return $null
    }
    
    $headers = @{
        "Content-Type" = "application/json"
    }
    
    if (-not [string]::IsNullOrEmpty($SessionCookie)) {
        $headers["Cookie"] = "better-auth.session_token=$SessionCookie"
    }
    
    $uri = "$BaseUrl$Endpoint"
    
    try {
        $params = @{
            Uri = $uri
            Method = $Method
            Headers = $headers
            ErrorAction = "Stop"
        }
        
        if ($Body -and $Method -ne "GET") {
            $params["Body"] = $Body
        }
        
        $response = Invoke-WebRequest @params
        $statusCode = $response.StatusCode
        
        if ($ExpectedStatus -contains $statusCode) {
            Write-Success "$Name - Status: $statusCode"
            $script:passed++
            
            # Try to parse JSON response
            try {
                $json = $response.Content | ConvertFrom-Json
                return $json
            } catch {
                return $response.Content
            }
        } else {
            Write-Fail "$Name - Expected: $($ExpectedStatus -join ', '), Got: $statusCode"
            $script:failed++
            return $null
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($ExpectedStatus -contains $statusCode) {
            Write-Success "$Name - Status: $statusCode (expected error response)"
            $script:passed++
        } else {
            Write-Fail "$Name - Error: $($_.Exception.Message)"
            $script:failed++
        }
        return $null
    }
}

# =====================================================================
# TESTS START HERE
# =====================================================================

Write-Host "`n" + "="*60 -ForegroundColor Magenta
Write-Host "  SANDBOX API TEST SUITE" -ForegroundColor Magenta
Write-Host "  Base URL: $BaseUrl" -ForegroundColor Magenta
Write-Host "="*60 -ForegroundColor Magenta

# ---------------------------------------------------------------------
Write-Test "1. Health Check"
# ---------------------------------------------------------------------
Test-Endpoint -Name "API Status" -Method "GET" -Endpoint "/api/status"

# ---------------------------------------------------------------------
Write-Test "2. List All Deployments (requires auth)"
# ---------------------------------------------------------------------
$deployments = Test-Endpoint -Name "List Deployments" -Method "GET" -Endpoint "/api/osa/deployments" -SkipIfNoAuth

if ($deployments -and $deployments.deployments) {
    Write-Info "  Found $($deployments.deployments.Count) deployment(s)"
    foreach ($dep in $deployments.deployments) {
        Write-Info "    - App: $($dep.name), Status: $($dep.status)"
    }
}

# ---------------------------------------------------------------------
Write-Test "3. Get Generated Apps (to find app ID)"
# ---------------------------------------------------------------------
$apps = Test-Endpoint -Name "List Generated Apps" -Method "GET" -Endpoint "/api/v1/osa/apps" -SkipIfNoAuth

if ($apps -and $apps.apps -and $apps.apps.Count -gt 0) {
    $testAppId = $apps.apps[0].id
    Write-Info "  Found $($apps.apps.Count) app(s)"
    Write-Info "  Using first app for tests: $testAppId"
} elseif (-not [string]::IsNullOrEmpty($AppId)) {
    $testAppId = $AppId
    Write-Info "  Using provided AppId: $testAppId"
} else {
    $testAppId = ""
    Write-Info "  No apps found - sandbox deploy tests will be skipped"
}

# ---------------------------------------------------------------------
Write-Test "4. Deployment Tests"
# ---------------------------------------------------------------------

if (-not [string]::IsNullOrEmpty($testAppId)) {
    # Get deployment status (might not exist yet)
    Test-Endpoint -Name "Get Deployment Status (pre-deploy)" -Method "GET" -Endpoint "/api/osa/deployment/$testAppId/status" -ExpectedStatus @(200, 404) -SkipIfNoAuth
    
    # Deploy app
    $deployResult = Test-Endpoint -Name "Deploy App" -Method "POST" -Endpoint "/api/osa/deployment/$testAppId/deploy" -ExpectedStatus @(200, 201, 202, 409) -SkipIfNoAuth
    
    if ($deployResult) {
        Write-Info "  Deployment initiated, status: $($deployResult.status)"
        
        # Wait a moment for deployment to start
        Start-Sleep -Seconds 2
        
        # Check deployment status
        $statusResult = Test-Endpoint -Name "Get Deployment Status" -Method "GET" -Endpoint "/api/osa/deployment/$testAppId/status" -SkipIfNoAuth
        
        if ($statusResult) {
            Write-Info "  Current status: $($statusResult.status)"
            if ($statusResult.url) {
                Write-Info "  URL: $($statusResult.url)"
            }
        }
        
        # Stop deployment
        Test-Endpoint -Name "Stop Deployment" -Method "POST" -Endpoint "/api/osa/deployment/$testAppId/stop" -ExpectedStatus @(200, 404) -SkipIfNoAuth
    }
} else {
    Write-Host "[SKIP] No app ID available for deployment tests" -ForegroundColor DarkYellow
    $script:skipped += 4
}

# ---------------------------------------------------------------------
Write-Test "5. Error Cases"
# ---------------------------------------------------------------------

# Test with invalid app ID
Test-Endpoint -Name "Get Invalid Deployment" -Method "GET" -Endpoint "/api/osa/deployment/invalid-uuid-here/status" -ExpectedStatus @(400, 404) -SkipIfNoAuth

# Test deploy non-existent app
Test-Endpoint -Name "Deploy Non-existent App" -Method "POST" -Endpoint "/api/osa/deployment/00000000-0000-0000-0000-000000000000/deploy" -ExpectedStatus @(404, 500) -SkipIfNoAuth

# ---------------------------------------------------------------------
Write-Test "6. App CRUD Operations"
# ---------------------------------------------------------------------

# Test PATCH endpoint (not PUT)
if (-not [string]::IsNullOrEmpty($testAppId)) {
    $patchBody = '{"app_name": "Test App Updated"}'
    Test-Endpoint -Name "Update App (PATCH)" -Method "PATCH" -Endpoint "/api/v1/osa/apps/$testAppId" -Body $patchBody -ExpectedStatus @(200, 404) -SkipIfNoAuth
} else {
    Write-Host "[SKIP] No app ID for PATCH test" -ForegroundColor DarkYellow
    $script:skipped++
}

# =====================================================================
# SUMMARY
# =====================================================================

Write-Host "`n" + "="*60 -ForegroundColor Magenta
Write-Host "  TEST RESULTS SUMMARY" -ForegroundColor Magenta
Write-Host "="*60 -ForegroundColor Magenta
Write-Host ""
Write-Success "Passed:  $script:passed"
Write-Fail "Failed:  $script:failed"
Write-Host "[SKIP] Skipped: $script:skipped" -ForegroundColor DarkYellow
Write-Host ""

$total = $script:passed + $script:failed
if ($total -gt 0) {
    $percentage = [math]::Round(($script:passed / $total) * 100, 1)
    Write-Host "Pass Rate: $percentage%" -ForegroundColor $(if ($percentage -ge 80) { "Green" } elseif ($percentage -ge 50) { "Yellow" } else { "Red" })
}

Write-Host "`n" + "-"*60 -ForegroundColor Gray
Write-Host "Usage Examples:" -ForegroundColor Gray
Write-Host "  # Run without auth (limited tests):" -ForegroundColor Gray
Write-Host "  .\test_sandbox_endpoints.ps1" -ForegroundColor White
Write-Host "" 
Write-Host "  # Run with session cookie:" -ForegroundColor Gray
Write-Host '  .\test_sandbox_endpoints.ps1 -SessionCookie "abc123..."' -ForegroundColor White
Write-Host ""
Write-Host "  # Run with specific app ID:" -ForegroundColor Gray
Write-Host '  .\test_sandbox_endpoints.ps1 -SessionCookie "abc..." -AppId "uuid-here"' -ForegroundColor White
Write-Host "-"*60 -ForegroundColor Gray
