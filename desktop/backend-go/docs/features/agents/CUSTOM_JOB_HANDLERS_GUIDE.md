# Custom Job Handlers - Complete Guide 🎨

**Created:** 2026-01-08
**File:** `desktop/backend-go/internal/handlers/custom_job_handlers.go`

---

## 📋 Overview

This guide explains the **7 production-ready custom job handlers** created for BusinessOS and how to use them.

---

## 🎯 Available Custom Handlers

| Handler | Purpose | Execution Time | Use Case |
|---------|---------|----------------|----------|
| **user_onboarding** | Complete user onboarding flow | ~2s | After signup |
| **workspace_export** | Export workspace data | ~3-4s | User requests export |
| **analytics_aggregation** | Aggregate analytics data | 0.5-5s | Daily/hourly metrics |
| **notification_batch** | Send bulk notifications | Variable | Announcements |
| **data_cleanup** | Remove old data | ~3s | Retention policies |
| **integration_sync** | Sync with external services | ~3s | Calendar, CRM sync |
| **backup** | Create data backups | ~7-8s | Scheduled backups |

---

## 🔧 How to Enable Custom Handlers

### Step 1: Import the Package

In `desktop/backend-go/cmd/server/main.go`, add the import:

```go
import (
    // ... existing imports ...
    "github.com/rhl/businessos-backend/internal/handlers"
)
```

### Step 2: Register Handlers

Find the worker initialization section (around line 692) and replace:

**FROM:**
```go
// Register job handlers
worker.RegisterHandler("email_send", services.ExampleEmailSendHandler)
worker.RegisterHandler("report_generate", services.ExampleReportGenerateHandler)
worker.RegisterHandler("sync_calendar", services.ExampleSyncCalendarHandler)
```

**TO:**
```go
// Register job handlers
worker.RegisterHandler("email_send", services.ExampleEmailSendHandler)
worker.RegisterHandler("report_generate", services.ExampleReportGenerateHandler)
worker.RegisterHandler("sync_calendar", services.ExampleSyncCalendarHandler)

// Register all custom handlers
handlers.RegisterAllCustomHandlers(worker.RegisterHandler)
```

### Step 3: Rebuild and Restart

```bash
cd desktop/backend-go
go build ./cmd/server
./server.exe
```

**Expected log:**
```
INFO All custom job handlers registered count=7
```

---

## 📚 Handler Details & Usage

### 1. User Onboarding Handler

**Purpose:** Automate complete user onboarding after signup

**Steps:**
1. Send welcome email
2. Create default workspace
3. Setup default projects
4. Send tutorial email

**Use in code:**
```go
// After user signs up
jobsService.EnqueueJob(ctx, "user_onboarding", map[string]interface{}{
    "user_id": newUser.ID,
    "email":   newUser.Email,
    "name":    newUser.Name,
}, 1, 3, nil)
```

**Use via API:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "user_onboarding",
    "payload": {
      "user_id": "user_123",
      "email": "john@example.com",
      "name": "John Doe"
    },
    "priority": 2
  }'
```

**Response example:**
```json
{
  "user_id": "user_123",
  "email": "john@example.com",
  "name": "John Doe",
  "steps": ["send_welcome_email", "create_default_workspace", "setup_default_projects", "send_tutorial_email"],
  "results": {
    "send_welcome_email": "completed",
    "create_default_workspace": "completed",
    "setup_default_projects": "completed",
    "send_tutorial_email": "completed"
  },
  "completed_at": "2026-01-08T13:00:00Z",
  "duration_ms": 2000,
  "status": "success"
}
```

---

### 2. Workspace Export Handler

**Purpose:** Export workspace data to PDF, CSV, or JSON

**Use in code:**
```go
jobsService.EnqueueJob(ctx, "workspace_export", map[string]interface{}{
    "workspace_id": workspaceID,
    "format":       "pdf", // or "csv", "json"
    "user_id":      userID,
}, 1, 3, nil)
```

**Use via API:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "workspace_export",
    "payload": {
      "workspace_id": "ws_456",
      "format": "pdf",
      "user_id": "user_123"
    },
    "priority": 1
  }'
```

**Response example:**
```json
{
  "export_id": "550e8400-e29b-41d4-a716-446655440000",
  "workspace_id": "ws_456",
  "format": "pdf",
  "file_name": "workspace_ws_456_export_20260108.pdf",
  "file_url": "https://storage.businessos.com/exports/ws_456/workspace_ws_456_export_20260108.pdf",
  "file_size_mb": 12.5,
  "records": 1250,
  "generated_at": "2026-01-08T13:05:00Z",
  "duration_ms": 3500,
  "status": "completed"
}
```

---

### 3. Analytics Aggregation Handler

**Purpose:** Aggregate analytics data for dashboards

**Schedule as cron job:**
```bash
# Hourly aggregation
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "analytics_aggregation",
    "payload": {
      "type": "hourly",
      "date": "2026-01-08"
    },
    "cron_expression": "0 * * * *",
    "name": "Hourly Analytics Aggregation"
  }'

# Daily aggregation at 2 AM
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "analytics_aggregation",
    "payload": {
      "type": "daily",
      "date": "2026-01-08"
    },
    "cron_expression": "0 2 * * *",
    "name": "Daily Analytics Aggregation"
  }'
```

**Response example:**
```json
{
  "aggregation_type": "daily",
  "date": "2026-01-08",
  "metrics": {
    "total_users": 1250,
    "active_users": 830,
    "new_signups": 42,
    "total_conversations": 3420,
    "total_messages": 15680,
    "avg_session_time": "18m 32s",
    "total_workspaces": 315
  },
  "aggregated_at": "2026-01-08T02:00:00Z",
  "duration_ms": 2000,
  "status": "completed"
}
```

---

### 4. Notification Batch Handler

**Purpose:** Send notifications to multiple users at once

**Use in code:**
```go
userIDs := []string{"user1", "user2", "user3", ...} // Can be thousands
jobsService.EnqueueJob(ctx, "notification_batch", map[string]interface{}{
    "type":     "email", // or "push", "sms"
    "message":  "Important system update scheduled for tonight",
    "user_ids": userIDs,
}, 2, 3, nil)
```

**Use via API:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "notification_batch",
    "payload": {
      "type": "push",
      "message": "New feature released!",
      "user_ids": ["user1", "user2", "user3", "user4", "user5"]
    },
    "priority": 2
  }'
```

**Response example:**
```json
{
  "notification_type": "push",
  "message": "New feature released!",
  "total_users": 5,
  "sent": 5,
  "failed": 0,
  "success_rate": 100,
  "sent_at": "2026-01-08T13:10:00Z",
  "duration_ms": 400,
  "status": "completed"
}
```

---

### 5. Data Cleanup Handler

**Purpose:** Remove old data based on retention policies

**Schedule as daily cleanup:**
```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "data_cleanup",
    "payload": {
      "data_type": "logs",
      "older_than_days": 30
    },
    "cron_expression": "0 3 * * *",
    "name": "Daily Log Cleanup",
    "description": "Remove logs older than 30 days"
  }'
```

**Use via API:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "data_cleanup",
    "payload": {
      "data_type": "sessions",
      "older_than_days": 7
    },
    "priority": 0
  }'
```

**Response example:**
```json
{
  "data_type": "logs",
  "older_than_days": 30,
  "deleted_count": 15420,
  "freed_space_mb": 342.5,
  "cleaned_at": "2026-01-08T03:00:00Z",
  "duration_ms": 3000,
  "status": "completed"
}
```

---

### 6. Integration Sync Handler

**Purpose:** Sync data with external services (Google Calendar, HubSpot, Slack)

**Schedule recurring sync:**
```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "integration_sync",
    "payload": {
      "integration": "google_calendar",
      "user_id": "user_123",
      "direction": "bidirectional"
    },
    "cron_expression": "*/15 * * * *",
    "name": "Google Calendar Sync"
  }'
```

**Use via API:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "integration_sync",
    "payload": {
      "integration": "hubspot",
      "user_id": "user_123",
      "direction": "pull"
    },
    "priority": 1
  }'
```

**Response example:**
```json
{
  "integration": "google_calendar",
  "user_id": "user_123",
  "direction": "bidirectional",
  "items_synced": 23,
  "last_sync_at": "2026-01-08T13:15:00Z",
  "next_sync_at": "2026-01-08T13:30:00Z",
  "duration_ms": 3000,
  "status": "completed"
}
```

---

### 7. Backup Handler

**Purpose:** Create automated backups of critical data

**Schedule daily backups:**
```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "backup",
    "payload": {
      "backup_type": "incremental",
      "target": "workspace_ws_456"
    },
    "cron_expression": "0 4 * * *",
    "name": "Daily Workspace Backup"
  }'
```

**Use via API:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "backup",
    "payload": {
      "backup_type": "full",
      "target": "system"
    },
    "priority": 3
  }'
```

**Response example:**
```json
{
  "backup_id": "550e8400-e29b-41d4-a716-446655440000",
  "backup_type": "full",
  "target": "system",
  "size_mb": 2450.5,
  "location": "s3://backups/system/550e8400-e29b-41d4-a716-446655440000.tar.gz.enc",
  "created_at": "2026-01-08T04:00:00Z",
  "expires_at": "2026-02-07T04:00:00Z",
  "duration_ms": 7800,
  "status": "completed",
  "verified": true
}
```

---

## 🔄 Common Use Case Workflows

### Workflow 1: New User Journey

```go
// 1. User signs up
newUser := createUser(email, password, name)

// 2. Enqueue onboarding job
jobsService.EnqueueJob(ctx, "user_onboarding", map[string]interface{}{
    "user_id": newUser.ID,
    "email":   newUser.Email,
    "name":    newUser.Name,
}, 2, 3, nil)

// User receives welcome email, gets workspace, sees tutorials
// All happens asynchronously in the background!
```

### Workflow 2: Scheduled System Maintenance

```bash
# 1. Daily analytics at 2 AM
POST /api/scheduled-jobs {"job_type": "analytics_aggregation", "cron": "0 2 * * *"}

# 2. Daily cleanup at 3 AM
POST /api/scheduled-jobs {"job_type": "data_cleanup", "cron": "0 3 * * *"}

# 3. Daily backup at 4 AM
POST /api/scheduled-jobs {"job_type": "backup", "cron": "0 4 * * *"}

# System maintains itself automatically!
```

### Workflow 3: Bulk Operations

```go
// Send announcement to all active users
activeUserIDs := getAllActiveUserIDs()

jobsService.EnqueueJob(ctx, "notification_batch", map[string]interface{}{
    "type":     "push",
    "message":  "System maintenance tonight at 10 PM",
    "user_ids": activeUserIDs, // Can be thousands!
}, 2, 3, nil)

// Job processes in batches of 50, handles failures automatically
```

---

## 🎨 Creating Your Own Custom Handler

### Template

```go
func YourCustomHandler(ctx context.Context, payload map[string]interface{}) (interface{}, error) {
    // 1. Extract parameters
    param1, _ := payload["param1"].(string)
    param2, _ := payload["param2"].(float64)

    // 2. Log start
    slog.InfoContext(ctx, "Starting your custom job", "param1", param1)

    // 3. Do your work
    startTime := time.Now()

    // ... your business logic here ...

    duration := time.Since(startTime)

    // 4. Log completion
    slog.InfoContext(ctx, "Custom job completed", "duration", duration)

    // 5. Return result
    return map[string]interface{}{
        "status":      "completed",
        "param1":      param1,
        "processed_at": time.Now(),
        "duration_ms": duration.Milliseconds(),
    }, nil
}
```

### Best Practices

1. **Always use slog for logging** - Not fmt.Printf
2. **Return structured results** - Use map[string]interface{}
3. **Handle errors properly** - Return errors, don't panic
4. **Log progress** - For long-running jobs
5. **Make it idempotent** - Job might run multiple times
6. **Include timing** - Track duration for monitoring
7. **Return useful data** - IDs, counts, URLs, etc.

---

## 📊 Monitoring Custom Handlers

### Check job status:
```bash
curl http://localhost:8080/api/background-jobs/{job_id}
```

### List jobs by type:
```bash
curl "http://localhost:8080/api/background-jobs?job_type=user_onboarding"
```

### Check server logs:
```
INFO Starting user onboarding user_id=user_123 email=john@example.com
INFO Onboarding step step=send_welcome_email progress=1/4
INFO Onboarding step step=create_default_workspace progress=2/4
INFO Onboarding step step=setup_default_projects progress=3/4
INFO Onboarding step step=send_tutorial_email progress=4/4
INFO User onboarding completed user_id=user_123
```

---

## ✅ Summary

You now have **7 production-ready job handlers** that cover common BusinessOS use cases:

✅ **user_onboarding** - Automate user setup
✅ **workspace_export** - Export workspace data
✅ **analytics_aggregation** - Aggregate metrics
✅ **notification_batch** - Bulk notifications
✅ **data_cleanup** - Retention policies
✅ **integration_sync** - External service sync
✅ **backup** - Automated backups

**Total:** ~400 lines of production-ready code!

**Next Steps:**
1. Register handlers in main.go
2. Rebuild the server
3. Start using them via API or code
4. Create your own custom handlers using the template

---

**File Location:** `desktop/backend-go/internal/handlers/custom_job_handlers.go`
**Documentation:** This file (CUSTOM_JOB_HANDLERS_GUIDE.md)
