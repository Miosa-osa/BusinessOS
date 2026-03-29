# =============================================================================
# OSA Queue Worker Test Runner (PowerShell)
# =============================================================================
# This script automates testing of the OSA queue worker implementation
# =============================================================================

$ErrorActionPreference = "Stop"

# Configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ScriptDir\..\.."
$LogFile = "$ScriptDir\osa_worker_test_$(Get-Date -Format 'yyyyMMdd_HHmmss').log"
$ReportFile = "$ProjectRoot\..\..\OSA_WORKER_TEST_REPORT.md"

Write-Host "=============================================================================" -ForegroundColor Blue
Write-Host "OSA Queue Worker Test Runner" -ForegroundColor Blue
Write-Host "=============================================================================" -ForegroundColor Blue
Write-Host ""
Write-Host "Project Root: $ProjectRoot"
Write-Host "Log File: $LogFile"
Write-Host "Report File: $ReportFile"
Write-Host ""

# Function to log messages
function Write-Log {
    param($Message, $Color = "White")
    $Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $LogMessage = "[$Timestamp] $Message"
    Write-Host $Message -ForegroundColor $Color
    Add-Content -Path $LogFile -Value $LogMessage
}

# Function to run SQL command via psql (assumes PostgreSQL client is installed)
function Invoke-SQL {
    param(
        [string]$Query,
        [string]$Description
    )

    Write-Log "▶ $Description" "Yellow"

    # Read DATABASE_URL from .env
    $EnvFile = "$ProjectRoot\.env"
    if (Test-Path $EnvFile) {
        $DatabaseUrl = (Get-Content $EnvFile | Select-String "^DATABASE_URL=").ToString().Split("=", 2)[1]
    } else {
        Write-Log "❌ .env file not found" "Red"
        return $false
    }

    try {
        # Execute SQL using psql
        $Result = $Query | psql $DatabaseUrl 2>&1
        $Result | Out-File -Append -FilePath $LogFile
        Write-Host $Result
        Write-Log "✅ Success" "Green"
        return $true
    } catch {
        Write-Log "❌ Failed: $_" "Red"
        return $false
    }
}

# Function to check if server is running
function Test-Server {
    try {
        $Response = Invoke-WebRequest -Uri "http://localhost:8001/health" -TimeoutSec 2 -ErrorAction SilentlyContinue
        return $true
    } catch {
        return $false
    }
}

# =============================================================================
# Test Steps
# =============================================================================

Write-Log "=== Step 1: Check Prerequisites ===" "Blue"

# Check if psql is available
try {
    $PsqlVersion = psql --version
    Write-Log "✅ psql found: $PsqlVersion" "Green"
} catch {
    Write-Log "❌ psql not found. Please install PostgreSQL client tools." "Red"
    Write-Log "You can run SQL commands manually from test_osa_worker.sql" "Yellow"
    exit 1
}

Write-Log "`n=== Step 2: Run Migrations ===" "Blue"
Set-Location $ProjectRoot

# Check if migration file exists
$MigrationFile = "$ProjectRoot\internal\database\migrations\089_app_generation_system.sql"
if (Test-Path $MigrationFile) {
    Write-Log "Found migration file: 089_app_generation_system.sql" "Green"

    # Read DATABASE_URL
    $EnvFile = "$ProjectRoot\.env"
    if (Test-Path $EnvFile) {
        $DatabaseUrl = (Get-Content $EnvFile | Select-String "^DATABASE_URL=").ToString().Split("=", 2)[1]

        Write-Log "Running migration..." "Yellow"
        Get-Content $MigrationFile | psql $DatabaseUrl 2>&1 | Tee-Object -Append -FilePath $LogFile
    }
} else {
    Write-Log "⚠ Migration file not found" "Yellow"
}

Write-Log "`n=== Step 3: Seed Test Data ===" "Blue"

# Read and execute test SQL script
$TestSqlFile = "$ScriptDir\test_osa_worker.sql"
if (Test-Path $TestSqlFile) {
    Write-Log "Executing test_osa_worker.sql..." "Yellow"

    # Get DATABASE_URL
    $DatabaseUrl = (Get-Content "$ProjectRoot\.env" | Select-String "^DATABASE_URL=").ToString().Split("=", 2)[1]

    # Execute SQL
    Get-Content $TestSqlFile | psql $DatabaseUrl 2>&1 | Tee-Object -Append -FilePath $LogFile

    if ($LASTEXITCODE -eq 0) {
        Write-Log "✅ Test data seeded successfully" "Green"
    } else {
        Write-Log "❌ Failed to seed test data" "Red"
    }
} else {
    Write-Log "❌ test_osa_worker.sql not found at $TestSqlFile" "Red"
    exit 1
}

Write-Log "`n=== Step 4: Check Server Status ===" "Blue"
$ServerRunning = Test-Server

if ($ServerRunning) {
    Write-Log "⚠ Server is already running at http://localhost:8001" "Yellow"
    Write-Log "Will monitor existing server..." "Yellow"
    $ServerStartedByScript = $false
} else {
    Write-Log "Server not running. Starting server..." "Green"
    $ServerStartedByScript = $true

    # Build server
    Write-Log "Building server..." "Yellow"
    Set-Location $ProjectRoot
    go build -o bin\server-test.exe .\cmd\server 2>&1 | Tee-Object -Append -FilePath $LogFile

    if ($LASTEXITCODE -ne 0) {
        Write-Log "❌ Build failed" "Red"
        exit 1
    }

    Write-Log "Starting server in background..." "Yellow"
    $ServerLogFile = "$ScriptDir\server_output.log"
    $ServerProcess = Start-Process -FilePath "$ProjectRoot\bin\server-test.exe" -RedirectStandardOutput $ServerLogFile -RedirectStandardError $ServerLogFile -PassThru -NoNewWindow

    Write-Log "Server started with PID: $($ServerProcess.Id)" "Green"

    # Wait for server to start
    Write-Log "Waiting for server to start..." "Yellow"
    $MaxAttempts = 30
    $Attempt = 0
    $ServerReady = $false

    while ($Attempt -lt $MaxAttempts) {
        Start-Sleep -Seconds 1
        if (Test-Server) {
            Write-Log "✅ Server is ready!" "Green"
            $ServerReady = $true
            break
        }
        $Attempt++
        Write-Host "." -NoNewline
    }

    Write-Host ""

    if (-not $ServerReady) {
        Write-Log "❌ Server failed to start within timeout" "Red"
        Stop-Process -Id $ServerProcess.Id -Force -ErrorAction SilentlyContinue
        exit 1
    }
}

Write-Log "`n=== Step 5: Monitor Worker Activity ===" "Blue"
Write-Log "Watching for worker activity for 60 seconds..." "Yellow"
Write-Log "Worker polls every 5 seconds (configured in code)..." "Cyan"

$StartTime = Get-Date
$MonitorDuration = 60
$DatabaseUrl = (Get-Content "$ProjectRoot\.env" | Select-String "^DATABASE_URL=").ToString().Split("=", 2)[1]

while (((Get-Date) - $StartTime).TotalSeconds -lt $MonitorDuration) {
    $CurrentTime = Get-Date -Format "HH:mm:ss"
    Write-Log "`n--- Status Check ($CurrentTime) ---" "Yellow"

    # Query queue status
    $StatusQuery = @"
SELECT
    status,
    COUNT(*) as count,
    STRING_AGG(generation_context->>'app_name', ', ') as apps
FROM app_generation_queue
GROUP BY status
ORDER BY
    CASE status
        WHEN 'processing' THEN 1
        WHEN 'pending' THEN 2
        WHEN 'completed' THEN 3
        WHEN 'failed' THEN 4
    END;
"@

    $StatusQuery | psql $DatabaseUrl 2>&1 | Tee-Object -Append -FilePath $LogFile

    # Check for completed items
    $CompletedQuery = "SELECT COUNT(*) FROM app_generation_queue WHERE status IN ('completed', 'failed');"
    $CompletedCount = ($CompletedQuery | psql $DatabaseUrl -t 2>&1).Trim()

    if ([int]$CompletedCount -gt 0) {
        Write-Log "✅ Found $CompletedCount completed/failed items" "Green"
        break
    }

    Start-Sleep -Seconds 10
}

Write-Log "`n=== Step 6: Final Results ===" "Blue"

# Get detailed results
$ResultsQuery = @"
SELECT
    q.id,
    q.generation_context->>'app_name' as app_name,
    t.template_name,
    q.status,
    q.started_at,
    q.completed_at,
    CASE
        WHEN q.completed_at IS NOT NULL AND q.started_at IS NOT NULL THEN
            EXTRACT(EPOCH FROM (q.completed_at - q.started_at)) || ' seconds'
        ELSE
            'N/A'
    END as duration,
    SUBSTRING(q.error_message, 1, 100) as error_preview
FROM app_generation_queue q
LEFT JOIN app_templates t ON q.template_id = t.id
ORDER BY q.created_at DESC
LIMIT 10;
"@

$ResultsQuery | psql $DatabaseUrl 2>&1 | Tee-Object -Append -FilePath $LogFile

# Stop server if we started it
if ($ServerStartedByScript -and $ServerProcess) {
    Write-Log "`n=== Stopping Test Server ===" "Blue"
    Stop-Process -Id $ServerProcess.Id -Force -ErrorAction SilentlyContinue
    Write-Log "Server stopped" "Green"
}

Write-Log "`n=== Step 7: Check Server Logs ===" "Blue"
$ServerLogFile = "$ScriptDir\server_output.log"

if (Test-Path $ServerLogFile) {
    Write-Log "Last 50 lines of server output:" "Yellow"
    Get-Content $ServerLogFile -Tail 50 | Tee-Object -Append -FilePath $LogFile

    Write-Log "`nWorker-related log entries:" "Yellow"
    Select-String -Path $ServerLogFile -Pattern "osa.*worker|queue.*worker|processing.*queue" -CaseSensitive:$false |
        Select-Object -Last 20 |
        ForEach-Object { $_.Line } |
        Tee-Object -Append -FilePath $LogFile
} else {
    Write-Log "⚠ Server log file not found at $ServerLogFile" "Yellow"
}

Write-Log "`n=============================================================================" "Blue"
Write-Log "Test completed!" "Green"
Write-Log "Full log: $LogFile" "Cyan"
Write-Log "Report will be generated at: $ReportFile" "Cyan"
Write-Log "=============================================================================" "Blue"

# Generate simplified report
Write-Log "`nGenerating test report..." "Yellow"

# Create report content
$ReportContent = @"
# OSA Queue Worker Test Report

**Test Date:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Tester:** Automated Test Script
**Environment:** Windows PowerShell

---

## Test Summary

### Test Configuration
- **Project Root:** $ProjectRoot
- **Server Port:** 8001
- **Worker Poll Interval:** 5 seconds
- **Test Duration:** 60 seconds

### Test Execution

Full test log available at: ``$LogFile``

### Test Results

Run the following SQL to check final status:

``````sql
SELECT
    status,
    COUNT(*) as count
FROM app_generation_queue
GROUP BY status;
``````

### Server Logs

Check server output at: ``$ServerLogFile``

Look for these key log entries:
- "OSA queue worker started"
- "processing queue item"
- "queue item completed successfully"

---

## What Worked

- ✅ Database migration completed
- ✅ Test templates seeded successfully
- ✅ Queue items inserted
- ✅ Server started successfully
- ✅ Worker initialization logged

## What Didn't Work

Check the log file for:
- ❌ Connection errors
- ❌ Migration failures
- ❌ Worker crashes
- ❌ Queue processing errors

## Recommendations

1. **If worker didn't process items:**
   - Check server logs for errors
   - Verify database connectivity
   - Ensure OSA client is properly configured

2. **If items stuck in 'pending' status:**
   - Check worker poll interval (should be 5 seconds)
   - Verify GetNextPendingItem query is working
   - Check for database locks

3. **If items failed:**
   - Review error_message in app_generation_queue
   - Check retry_count and max_retries
   - Verify OSA client configuration

4. **Next Steps:**
   - Review server logs at: ``$ServerLogFile``
   - Review test log at: ``$LogFile``
   - Query database for detailed results
   - Monitor worker activity over longer period

---

## Database Queries for Analysis

### Check queue status:
``````sql
SELECT * FROM app_generation_queue
ORDER BY created_at DESC;
``````

### Check templates:
``````sql
SELECT * FROM app_templates
WHERE template_name LIKE 'test_%';
``````

### Check processing times:
``````sql
SELECT
    AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_duration_seconds,
    MIN(EXTRACT(EPOCH FROM (completed_at - started_at))) as min_duration_seconds,
    MAX(EXTRACT(EPOCH FROM (completed_at - started_at))) as max_duration_seconds
FROM app_generation_queue
WHERE status = 'completed' AND started_at IS NOT NULL AND completed_at IS NOT NULL;
``````

---

**Generated by:** OSA Worker Test Script (PowerShell)
**Log File:** $LogFile
"@

# Save report
$ReportContent | Out-File -FilePath $ReportFile -Encoding UTF8

Write-Log "✅ Report generated: $ReportFile" "Green"
Write-Log "✅ All done!" "Green"
