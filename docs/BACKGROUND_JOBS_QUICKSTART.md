# Background Jobs - Quick Start Guide 🚀

**Status:** ✅ Production Ready | **Build:** ✅ Verified | **Date:** 2026-01-08

---

## ⚡ 60-Second Quickstart

### 1. Start the Server
```bash
cd desktop/backend-go
./server.exe
```

**Expected output:**
```
INFO Initializing background jobs system...
INFO Worker started worker_id=worker-1
INFO Worker started worker_id=worker-2
INFO Worker started worker_id=worker-3
INFO Job scheduler started
INFO Background jobs routes registered
Server starting on port 8080
```

### 2. Create Your First Job
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "email_send",
    "payload": {
      "to": "test@example.com",
      "subject": "My First Background Job!",
      "body": "It works!"
    },
    "priority": 1
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "job_type": "email_send",
  "status": "pending",
  "priority": 1,
  "attempt_count": 0,
  "max_attempts": 3,
  "created_at": "2026-01-08T10:00:00Z"
}
```

### 3. Watch It Process (in server logs)
```
INFO Job acquired job_id=550e8400... job_type=email_send worker_id=worker-1
INFO Sending email to=test@example.com subject="My First Background Job!"
INFO Job completed successfully job_id=550e8400... duration=1.023s
```

**🎉 Congratulations! You just ran your first background job!**

---

## 📚 Common Use Cases

### Use Case 1: Send Email After User Action

**In your code:**
```go
// After user signup, checkout, etc.
jobsService.EnqueueJob(ctx, "email_send", map[string]interface{}{
    "to": user.Email,
    "subject": "Welcome!",
    "body": "Thanks for signing up!",
}, 1, 3, nil)
```

### Use Case 2: Generate Daily Report at 9 AM

```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "daily_report",
    "payload": {"report_type": "sales"},
    "cron_expression": "0 9 * * *",
    "timezone": "America/Sao_Paulo",
    "name": "Daily Sales Report"
  }'
```

### Use Case 3: Sync Calendar Every 15 Minutes

```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "sync_calendar",
    "payload": {"user_id": "user123"},
    "cron_expression": "*/15 * * * *",
    "name": "Calendar Sync"
  }'
```

---

## 🔧 Built-in Job Types

| Job Type | Description | Example Payload |
|----------|-------------|-----------------|
| `email_send` | Send email | `{"to": "user@example.com", "subject": "Hi", "body": "Hello!"}` |
| `report_generate` | Generate report | `{"report_type": "sales", "start_date": "2026-01-01"}` |
| `sync_calendar` | Sync calendar | `{"user_id": "123", "calendar_id": "primary"}` |

---

## 🎨 Create Your Own Job Handler

### Step 1: Create Handler Function

```go
// In any .go file in your project
func SendNotificationHandler(ctx context.Context, payload map[string]interface{}) (interface{}, error) {
    userID, _ := payload["user_id"].(string)
    message, _ := payload["message"].(string)

    slog.InfoContext(ctx, "Sending notification", "user_id", userID)

    // Your logic here
    // err := notificationService.Send(userID, message)
    // if err != nil { return nil, err }

    return map[string]interface{}{
        "sent_at": time.Now(),
        "status": "delivered",
    }, nil
}
```

### Step 2: Register in main.go (around line 700)

```go
worker.RegisterHandler("send_notification", SendNotificationHandler)
```

### Step 3: Use It

```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "send_notification",
    "payload": {
      "user_id": "user123",
      "message": "Your order shipped!"
    }
  }'
```

---

## 📊 Monitor Jobs

### List All Jobs
```bash
curl http://localhost:8080/api/background-jobs
```

### List Pending Jobs
```bash
curl "http://localhost:8080/api/background-jobs?status=pending"
```

### List Failed Jobs
```bash
curl "http://localhost:8080/api/background-jobs?status=failed"
```

### Get Specific Job
```bash
curl http://localhost:8080/api/background-jobs/{job_id}
```

---

## 🔄 Retry & Cancel

### Retry Failed Job
```bash
curl -X POST http://localhost:8080/api/background-jobs/{job_id}/retry
```

### Cancel Job
```bash
curl -X POST http://localhost:8080/api/background-jobs/{job_id}/cancel
```

---

## 📅 Cron Expression Cheat Sheet

| Expression | Meaning |
|------------|---------|
| `* * * * *` | Every minute |
| `*/5 * * * *` | Every 5 minutes |
| `0 * * * *` | Every hour (at minute 0) |
| `0 9 * * *` | Every day at 9:00 AM |
| `0 9 * * 1-5` | Weekdays at 9:00 AM |
| `0 0 * * 0` | Every Sunday at midnight |
| `0 0 1 * *` | 1st of every month at midnight |
| `*/15 9-17 * * 1-5` | Every 15 min, 9 AM-5 PM, weekdays |

**Format:** `minute hour day month day_of_week`

---

## 🧪 Test Everything

Run the automated test script:

```bash
cd desktop/backend-go/scripts/tests
./test_background_jobs_api.sh
```

This tests all 12 endpoints automatically!

---

## 🚨 Troubleshooting

### Problem: Jobs stuck in "pending"

**Check:**
1. Are workers running? Look for `Worker started` in logs
2. Is job `scheduled_at` in the future? Check job details
3. Did job exceed `max_attempts`? Check `attempt_count`

**Solution:**
```bash
# Check job details
curl http://localhost:8080/api/background-jobs/{job_id}

# Retry manually
curl -X POST http://localhost:8080/api/background-jobs/{job_id}/retry
```

### Problem: Jobs stuck in "running"

**Cause:** Worker crashed or job is taking too long

**Solution:**
Jobs auto-release after 5 minutes. Or restart server (graceful shutdown releases stuck jobs).

### Problem: Jobs failing repeatedly

**Check error:**
```bash
curl http://localhost:8080/api/background-jobs/{job_id}
```

Look at `last_error` field for details.

---

## 📖 Documentation

- **Quick Start:** `BACKGROUND_JOBS_QUICKSTART.md` (this file)
- **User Guide:** `BACKGROUND_JOBS_README.md` (complete documentation)
- **Integration Guide:** `BACKGROUND_JOBS_INTEGRATION_GUIDE.md` (how to integrate)
- **Verification:** `BACKGROUND_JOBS_VERIFICATION.md` (implementation proof)

---

## ✅ What You Have

✅ **Reliable task queue** - Won't lose jobs, atomic locking
✅ **Auto-retry** - Failed jobs retry automatically (1min → 5min → 15min)
✅ **Cron scheduling** - Run jobs on schedule (daily, hourly, etc.)
✅ **12 REST endpoints** - Full API for job management
✅ **3 workers** - Process jobs concurrently
✅ **Graceful shutdown** - Jobs complete before server stops
✅ **Production-ready** - Built with PostgreSQL, indexed, tested

---

## 🎯 Next Steps

1. **Start using it!** Create jobs via API or code
2. **Add your handlers** - Implement custom job types
3. **Schedule recurring jobs** - Set up daily/hourly tasks
4. **Monitor** - Use REST API to track job status

**Questions?** Check the full documentation in `BACKGROUND_JOBS_README.md`

---

**Happy background job processing!** 🚀
