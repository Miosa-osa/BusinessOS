# Notification Seed Script for Business OS
# Use this script to seed or clear test notifications
#
# Prerequisites:
# 1. Backend must be running (go run ./cmd/server or start_backend.bat)
# 2. You must be logged in (have a valid session cookie)
# 3. Backend must be in development mode (ENVIRONMENT != production)

param(
    [Parameter(Position=0)]
    [ValidateSet("seed", "seed-full", "clear", "help")]
    [string]$Action = "help",
    
    [string]$BaseUrl = "http://localhost:8080",
    
    [string]$CookieFile = ""
)

$ErrorActionPreference = "Stop"

function Show-Help {
    Write-Host @"
╔══════════════════════════════════════════════════════════════╗
║           NOTIFICATION SEED SCRIPT                           ║
╚══════════════════════════════════════════════════════════════╝

USAGE:
    .\seed_notifications.ps1 <action> [-BaseUrl <url>] [-CookieFile <path>]

ACTIONS:
    seed        Create test notifications (uses service, respects batching)
    seed-full   Create test notifications with varied timestamps (bypasses batching)
    clear       Remove ALL notifications for your user
    help        Show this help message

OPTIONS:
    -BaseUrl     Backend API URL (default: http://localhost:8080)
    -CookieFile  Path to file containing session cookie (optional)

EXAMPLES:
    # Seed notifications (default backend URL)
    .\seed_notifications.ps1 seed

    # Seed with timestamps for testing date grouping
    .\seed_notifications.ps1 seed-full

    # Clear all notifications
    .\seed_notifications.ps1 clear

    # Use custom backend URL
    .\seed_notifications.ps1 seed -BaseUrl http://localhost:18080

NOTES:
    - Backend must be running in DEVELOPMENT mode
    - You must be authenticated (have a valid session)
    - If using browser, copy your session cookie to a file

TO GET YOUR SESSION COOKIE:
    1. Open your app in browser
    2. Open DevTools (F12) > Application > Cookies
    3. Copy the 'better-auth.session_token' cookie value
    4. Save to a file or use browser's cookie

"@
}

function Invoke-SeedApi {
    param(
        [string]$Endpoint,
        [string]$Method = "POST"
    )
    
    $url = "$BaseUrl/api/dev/notifications/$Endpoint"
    
    Write-Host "🔗 Calling: $Method $url" -ForegroundColor Cyan
    
    $headers = @{
        "Content-Type" = "application/json"
    }
    
    # If cookie file provided, read it
    if ($CookieFile -and (Test-Path $CookieFile)) {
        $cookie = Get-Content $CookieFile -Raw
        $headers["Cookie"] = "better-auth.session_token=$cookie"
    }
    
    try {
        $response = Invoke-RestMethod -Uri $url -Method $Method -Headers $headers -UseBasicParsing -SessionVariable session
        return $response
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $errorBody = $_.ErrorDetails.Message
        
        if ($statusCode -eq 401) {
            Write-Host "❌ Authentication failed. Please log in first." -ForegroundColor Red
            Write-Host "   Make sure you have a valid session cookie." -ForegroundColor Yellow
        }
        elseif ($statusCode -eq 403) {
            Write-Host "❌ Forbidden. Seed endpoints only work in development mode." -ForegroundColor Red
            Write-Host "   Check that ENVIRONMENT != 'production' on your backend." -ForegroundColor Yellow
        }
        else {
            Write-Host "❌ Request failed with status $statusCode" -ForegroundColor Red
            Write-Host "   $errorBody" -ForegroundColor Yellow
        }
        exit 1
    }
}

# Main execution
switch ($Action) {
    "help" {
        Show-Help
    }
    "seed" {
        Write-Host "`n🌱 Seeding notifications..." -ForegroundColor Green
        $result = Invoke-SeedApi -Endpoint "seed" -Method "POST"
        Write-Host "✅ Success! Created $($result.count) notifications" -ForegroundColor Green
        Write-Host "   $($result.message)" -ForegroundColor Gray
    }
    "seed-full" {
        Write-Host "`n🌱 Seeding notifications with timestamps..." -ForegroundColor Green
        $result = Invoke-SeedApi -Endpoint "seed-full" -Method "POST"
        Write-Host "✅ Success! Created $($result.count) notifications" -ForegroundColor Green
        Write-Host "   $($result.message)" -ForegroundColor Gray
    }
    "clear" {
        Write-Host "`n🗑️  Clearing all notifications..." -ForegroundColor Yellow
        $result = Invoke-SeedApi -Endpoint "seed" -Method "DELETE"
        Write-Host "✅ Success! Deleted $($result.deleted) notifications" -ForegroundColor Green
        Write-Host "   $($result.message)" -ForegroundColor Gray
    }
}

Write-Host "`n📋 Done! Refresh your notifications page to see the changes.`n" -ForegroundColor Cyan
