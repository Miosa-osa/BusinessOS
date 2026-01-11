# Background Jobs API - Testing Guide 🧪

**Date:** 2026-01-08
**Purpose:** Complete testing guide for the Background Jobs API

---

## 🚀 Prerequisites

### 1. Start the Server

```bash
cd desktop/backend-go
./server.exe
```

**Wait for this log:**
```
INFO Initializing background jobs system...
INFO Worker started worker_id=worker-1
INFO Worker started worker_id=worker-2
INFO Worker started worker_id=worker-3
INFO Job scheduler started
INFO Background jobs routes registered
Server starting on port 8080
```

### 2. Verify Server is Running

```bash
curl http://localhost:8080/health
```

**Expected response:**
```json
{"status": "ok"}
```

---

## 📋 API Tests

### Test 1: Create Background Job

**Request:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "email_send",
    "payload": {
      "to": "test@example.com",
      "subject": "Background Jobs Test",
      "body": "Testing the system!"
    },
    "priority": 1,
    "max_attempts": 3
  }'
```

**Expected response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "job_type": "email_send",
  "payload": {
    "to": "test@example.com",
    "subject": "Background Jobs Test",
    "body": "Testing the system!"
  },
  "scheduled_at": "2026-01-08T13:00:00Z",
  "priority": 1,
  "status": "pending",
  "started_at": null,
  "completed_at": null,
  "worker_id": null,
  "locked_until": null,
  "attempt_count": 0,
  "max_attempts": 3,
  "last_error": null,
  "result": null,
  "created_at": "2026-01-08T13:00:00Z"
}
```

**Server logs (within 5 seconds):**
```
INFO Job acquired job_id=550e8400... job_type=email_send worker_id=worker-1 attempt=1
INFO Sending email to=test@example.com subject="Background Jobs Test"
INFO Job completed successfully job_id=550e8400... duration=1.002s
```

---

### Test 2: List All Jobs

**Request:**
```bash
curl http://localhost:8080/api/background-jobs
```

**Expected response:**
```json
{
  "jobs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "job_type": "email_send",
      "status": "completed",
      "priority": 1,
      "attempt_count": 1,
      "created_at": "2026-01-08T13:00:00Z",
      "completed_at": "2026-01-08T13:00:05Z"
    }
  ],
  "total": 1
}
```

---

### Test 3: Get Specific Job

**Request:**
```bash
# Use the ID from Test 1
JOB_ID="550e8400-e29b-41d4-a716-446655440000"
curl http://localhost:8080/api/background-jobs/$JOB_ID
```

**Expected response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "job_type": "email_send",
  "status": "completed",
  "result": {
    "sent_at": "2026-01-08T13:00:05Z",
    "to": "test@example.com",
    "status": "sent"
  },
  "completed_at": "2026-01-08T13:00:05Z"
}
```

---

### Test 4: Filter Jobs by Status

**Request:**
```bash
# List only pending jobs
curl "http://localhost:8080/api/background-jobs?status=pending"

# List only completed jobs
curl "http://localhost:8080/api/background-jobs?status=completed"

# List only failed jobs
curl "http://localhost:8080/api/background-jobs?status=failed"
```

---

### Test 5: Filter Jobs by Type

**Request:**
```bash
curl "http://localhost:8080/api/background-jobs?job_type=email_send"
```

---

### Test 6: Create Scheduled Job (Cron)

**Request:**
```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "report_generate",
    "payload": {
      "report_type": "daily_sales",
      "recipients": ["admin@example.com"]
    },
    "cron_expression": "0 9 * * *",
    "timezone": "America/Sao_Paulo",
    "name": "Daily Sales Report",
    "description": "Generates and sends daily sales report at 9 AM"
  }'
```

**Expected response:**
```json
{
  "id": "660f9511-f39c-52e5-b827-557766551111",
  "job_type": "report_generate",
  "payload": {
    "report_type": "daily_sales",
    "recipients": ["admin@example.com"]
  },
  "cron_expression": "0 9 * * *",
  "timezone": "America/Sao_Paulo",
  "is_active": true,
  "last_run_at": null,
  "next_run_at": "2026-01-09T09:00:00-03:00",
  "name": "Daily Sales Report",
  "description": "Generates and sends daily sales report at 9 AM",
  "created_at": "2026-01-08T13:05:00Z",
  "updated_at": "2026-01-08T13:05:00Z"
}
```

---

### Test 7: List Scheduled Jobs

**Request:**
```bash
curl http://localhost:8080/api/scheduled-jobs
```

**Expected response:**
```json
{
  "scheduled_jobs": [
    {
      "id": "660f9511-f39c-52e5-b827-557766551111",
      "job_type": "report_generate",
      "cron_expression": "0 9 * * *",
      "is_active": true,
      "next_run_at": "2026-01-09T09:00:00-03:00",
      "name": "Daily Sales Report"
    }
  ],
  "total": 1
}
```

---

### Test 8: Get Scheduled Job Details

**Request:**
```bash
SCHEDULED_JOB_ID="660f9511-f39c-52e5-b827-557766551111"
curl http://localhost:8080/api/scheduled-jobs/$SCHEDULED_JOB_ID
```

---

### Test 9: Update Scheduled Job

**Request:**
```bash
curl -X PUT http://localhost:8080/api/scheduled-jobs/$SCHEDULED_JOB_ID \
  -H 'Content-Type: application/json' \
  -d '{
    "cron_expression": "0 10 * * *",
    "name": "Daily Sales Report (Updated)"
  }'
```

---

### Test 10: Disable Scheduled Job

**Request:**
```bash
curl -X POST http://localhost:8080/api/scheduled-jobs/$SCHEDULED_JOB_ID/disable
```

**Expected response:**
```json
{
  "message": "Scheduled job disabled",
  "scheduled_job": {
    "id": "660f9511-f39c-52e5-b827-557766551111",
    "is_active": false
  }
}
```

---

### Test 11: Enable Scheduled Job

**Request:**
```bash
curl -X POST http://localhost:8080/api/scheduled-jobs/$SCHEDULED_JOB_ID/enable
```

---

### Test 12: Delete Scheduled Job

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/scheduled-jobs/$SCHEDULED_JOB_ID
```

---

### Test 13: Retry Failed Job

**First, create a job that will fail:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "nonexistent_handler",
    "payload": {},
    "priority": 1
  }'
```

**This will fail because no handler is registered for "nonexistent_handler"**

**Then retry:**
```bash
FAILED_JOB_ID="<id from response>"
curl -X POST http://localhost:8080/api/background-jobs/$FAILED_JOB_ID/retry
```

---

### Test 14: Cancel Job

**Request:**
```bash
curl -X POST http://localhost:8080/api/background-jobs/$JOB_ID/cancel
```

**Expected response:**
```json
{
  "message": "Job cancelled",
  "job": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "cancelled"
  }
}
```

---

## 🎨 Testing Custom Handlers

### Test User Onboarding Handler

**Request:**
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

**Expected logs:**
```
INFO Starting user onboarding user_id=user_123 email=john@example.com
INFO Onboarding step step=send_welcome_email progress=1/4
INFO Onboarding step step=create_default_workspace progress=2/4
INFO Onboarding step step=setup_default_projects progress=3/4
INFO Onboarding step step=send_tutorial_email progress=4/4
INFO User onboarding completed user_id=user_123
INFO Job completed successfully job_id=... duration=2.001s
```

**Expected result:**
```json
{
  "user_id": "user_123",
  "email": "john@example.com",
  "name": "John Doe",
  "steps": [
    "send_welcome_email",
    "create_default_workspace",
    "setup_default_projects",
    "send_tutorial_email"
  ],
  "results": {
    "send_welcome_email": "completed",
    "create_default_workspace": "completed",
    "setup_default_projects": "completed",
    "send_tutorial_email": "completed"
  },
  "completed_at": "2026-01-08T13:10:05Z",
  "duration_ms": 2000,
  "status": "success"
}
```

---

### Test Workspace Export Handler

**Request:**
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

**Expected result:**
```json
{
  "export_id": "550e8400-e29b-41d4-a716-446655440000",
  "workspace_id": "ws_456",
  "format": "pdf",
  "file_name": "workspace_ws_456_export_20260108.pdf",
  "file_url": "https://storage.businessos.com/exports/ws_456/workspace_ws_456_export_20260108.pdf",
  "file_size_mb": 12.5,
  "records": 1250,
  "generated_at": "2026-01-08T13:15:00Z",
  "duration_ms": 3500,
  "status": "completed"
}
```

---

### Test Notification Batch Handler

**Request:**
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

**Expected result:**
```json
{
  "notification_type": "push",
  "message": "New feature released!",
  "total_users": 5,
  "sent": 5,
  "failed": 0,
  "success_rate": 100,
  "sent_at": "2026-01-08T13:20:00Z",
  "duration_ms": 400,
  "status": "completed"
}
```

---

## 🔄 Advanced Testing Scenarios

### Scenario 1: Test Retry Logic

**Create a job that will fail:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "nonexistent_handler",
    "payload": {},
    "max_attempts": 3
  }'
```

**Watch logs:**
```
INFO Job acquired job_id=xxx job_type=nonexistent_handler
ERROR Job execution failed job_id=xxx error="no handler registered"
INFO Job scheduled for retry job_id=xxx next_attempt_at=2026-01-08T13:26:00Z attempt=1/3

... 1 minute later ...

INFO Job acquired job_id=xxx job_type=nonexistent_handler attempt=2
ERROR Job execution failed job_id=xxx error="no handler registered"
INFO Job scheduled for retry job_id=xxx next_attempt_at=2026-01-08T13:32:00Z attempt=2/3

... 5 minutes later ...

INFO Job acquired job_id=xxx job_type=nonexistent_handler attempt=3
ERROR Job execution failed job_id=xxx error="no handler registered"
INFO Job marked as failed (max attempts reached) job_id=xxx
```

---

### Scenario 2: Test Priority Queue

**Create jobs with different priorities:**
```bash
# Low priority
curl -X POST http://localhost:8080/api/background-jobs \
  -d '{"job_type": "email_send", "payload": {"to": "low@example.com"}, "priority": 0}'

# High priority
curl -X POST http://localhost:8080/api/background-jobs \
  -d '{"job_type": "email_send", "payload": {"to": "high@example.com"}, "priority": 10}'

# Medium priority
curl -X POST http://localhost:8080/api/background-jobs \
  -d '{"job_type": "email_send", "payload": {"to": "medium@example.com"}, "priority": 5}'
```

**Execution order:**
```
1. high@example.com    (priority 10)
2. medium@example.com  (priority 5)
3. low@example.com     (priority 0)
```

---

### Scenario 3: Test Scheduled Jobs

**Create scheduled job:**
```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "analytics_aggregation",
    "payload": {"type": "hourly"},
    "cron_expression": "*/5 * * * *",
    "name": "Test Every 5 Minutes"
  }'
```

**Wait 5 minutes, check background_jobs:**
```bash
curl "http://localhost:8080/api/background-jobs?job_type=analytics_aggregation"
```

**You should see a new background_job created automatically!**

---

## 📊 Monitoring & Debugging

### Check Server Health

```bash
curl http://localhost:8080/health/detailed
```

### View All Jobs

```bash
curl http://localhost:8080/api/background-jobs | jq '.'
```

### View Pending Jobs

```bash
curl "http://localhost:8080/api/background-jobs?status=pending" | jq '.jobs[] | {id, job_type, priority, created_at}'
```

### View Failed Jobs

```bash
curl "http://localhost:8080/api/background-jobs?status=failed" | jq '.jobs[] | {id, job_type, last_error, attempt_count}'
```

### View Job with Full Details

```bash
curl http://localhost:8080/api/background-jobs/$JOB_ID | jq '.'
```

---

## 🧪 Automated Test Script

**Use the provided test script:**
```bash
cd desktop/backend-go/scripts/tests
./test_background_jobs_api.sh
```

**This script automatically tests all 12 endpoints!**

---

## ✅ Success Criteria

After running all tests, you should see:

- ✅ Jobs created successfully
- ✅ Jobs processed by workers (check logs)
- ✅ Jobs completed with status="completed"
- ✅ Scheduled jobs creating background jobs
- ✅ Retry logic working for failed jobs
- ✅ Priority queue respecting priorities
- ✅ All API endpoints responding correctly

---

## 🐛 Troubleshooting

### Problem: Jobs stuck in "pending"

**Check:**
```bash
# Are workers running?
# Look for "Worker started" in server logs

# Is scheduled_at in the future?
curl http://localhost:8080/api/background-jobs/$JOB_ID | jq '.scheduled_at'
```

### Problem: Jobs failing

**Check:**
```bash
# View error message
curl http://localhost:8080/api/background-jobs/$JOB_ID | jq '.last_error'

# Check server logs for details
```

### Problem: Scheduled jobs not running

**Check:**
```bash
# Is scheduler running?
# Look for "Job scheduler started" in logs

# Is scheduled job active?
curl http://localhost:8080/api/scheduled-jobs/$SCHEDULED_JOB_ID | jq '.is_active'

# Check next_run_at
curl http://localhost:8080/api/scheduled-jobs/$SCHEDULED_JOB_ID | jq '.next_run_at'
```

---

## 📝 Summary

You now have a complete testing guide covering:

✅ **14 core API tests** - All endpoints covered
✅ **Custom handler tests** - Testing the 7 custom handlers
✅ **Advanced scenarios** - Retry logic, priority queue, scheduling
✅ **Monitoring commands** - Health checks, filtering, debugging
✅ **Automated script** - One command to test everything
✅ **Troubleshooting guide** - Common issues and solutions

**Ready to test! Start the server and run the tests!** 🚀
