# Integration Testing Guide

## Overview

This document provides comprehensive testing instructions for the BusinessOS integration system.

---

## ✅ Completed Tests

### Webhook Signature Verification Tests

**File**: `internal/webhooks/signature_test.go` (465 lines)

**Test Coverage**:
- ✅ Slack signature verification (HMAC-SHA256 with timestamp)
- ✅ Linear signature verification (HMAC-SHA256)
- ✅ HubSpot v3 signature verification (HMAC-SHA256 with method+URI+body+timestamp)
- ✅ Notion signature verification (HMAC-SHA256)
- ✅ Airtable signature verification (HMAC-SHA256 with timestamp)
- ✅ Fathom signature verification (HMAC-SHA256)
- ✅ Google Channel Token verification
- ✅ Microsoft Client State verification
- ✅ Generic HMAC-SHA256 helper

**Run Tests**:
```bash
go test -v ./internal/webhooks
```

**Test Results**: ✅ **ALL PASS** (9 test suites, 25+ individual tests)

---

## 🧪 CUS-118: Live Sync Integration Tests

### Status

**Completed**:
- ✅ Webhook signature verification (9 providers)
- ✅ Replay attack protection (timestamp validation)
- ✅ Development mode support (no secret configured)

**TODO** (Future Work):
- ⏳ Webhook processing tests (requires test database)
- ⏳ Sync service tests (requires test database)
- ⏳ Conflict resolution tests (requires vector clock setup)

### Manual Webhook Testing

Use the following script to test webhook handlers manually:

```bash
#!/bin/bash
# Test Linear webhook
curl -X POST http://localhost:8001/api/webhooks/linear \
  -H "Content-Type: application/json" \
  -H "Linear-Signature: <calculated-hmac>" \
  -d '{
    "action": "create",
    "type": "Issue",
    "organizationId": "test-org",
    "data": {
      "id": "issue-123",
      "identifier": "LIN-123",
      "title": "Test Issue",
      "state": {"name": "Todo", "type": "todo"},
      "team": {"id": "team-1", "name": "Engineering"}
    }
  }'

# Test Slack webhook
curl -X POST http://localhost:8001/api/webhooks/slack/events \
  -H "Content-Type: application/json" \
  -H "X-Slack-Request-Timestamp: $(date +%s)" \
  -H "X-Slack-Signature: <calculated-hmac>" \
  -d '{
    "type": "event_callback",
    "team_id": "T123",
    "event": {
      "type": "message",
      "channel": "C123",
      "user": "U123",
      "text": "Test message",
      "ts": "1234567890.123456"
    }
  }'

# Test Google Calendar webhook
curl -X POST http://localhost:8001/api/webhooks/google/calendar \
  -H "X-Goog-Channel-ID: channel-123" \
  -H "X-Goog-Resource-ID: resource-123" \
  -H "X-Goog-Resource-State: exists" \
  -H "X-Goog-Channel-Token: user-uuid"
```

---

## 🔄 CUS-110: End-to-End Integration Testing

### OAuth Flow Testing

Test each provider's OAuth flow manually:

#### 1. Google

```bash
# Start OAuth flow
curl http://localhost:8001/api/integrations/google/connect

# Follow the authorization URL
# After callback, verify token stored:
psql -d businessos -c "SELECT provider, created_at FROM oauth_tokens WHERE provider='google';"
```

#### 2. Slack

```bash
# Start OAuth flow
curl http://localhost:8001/api/integrations/slack/connect

# After callback, verify:
psql -d businessos -c "SELECT provider, team_id FROM oauth_tokens WHERE provider='slack';"
```

#### 3. Linear

```bash
curl http://localhost:8001/api/integrations/linear/connect
```

#### 4. HubSpot

```bash
curl http://localhost:8001/api/integrations/hubspot/connect
```

### API Endpoints Testing

```bash
# List all providers
curl http://localhost:8001/api/integrations/providers

# List user integrations (requires auth token)
curl -H "Authorization: Bearer TOKEN" http://localhost:8001/api/integrations/

# Check specific integration
curl -H "Authorization: Bearer TOKEN" http://localhost:8001/api/integrations/google

# Trigger sync
curl -X POST -H "Authorization: Bearer TOKEN" \
  http://localhost:8001/api/integrations/google/sync

# Disconnect integration
curl -X DELETE -H "Authorization: Bearer TOKEN" \
  http://localhost:8001/api/integrations/google
```

### Database Verification

After running tests, verify data in database:

```sql
-- Check OAuth tokens
SELECT provider, user_id, expires_at FROM oauth_tokens;

-- Check synced calendar events
SELECT provider, external_id, title, start_time FROM synced_calendar_events;

-- Check synced tasks
SELECT provider, identifier, title, status FROM synced_tasks;

-- Check synced messages
SELECT provider, channel_id, sender_name, content FROM synced_messages;

-- Check synced contacts
SELECT provider, email, full_name, company FROM synced_contacts;

-- Check synced meetings
SELECT provider, title, duration_seconds, participant_count FROM synced_meetings;

-- Check webhook subscriptions
SELECT provider, resource_type, status, event_count FROM webhook_subscriptions;

-- Check sync tokens
SELECT provider, resource_type, last_sync_at FROM sync_tokens;
```

---

## 🖥️ CUS-111: Frontend Integration UI Verification

### Manual Testing Checklist

#### 1. Navigate to Integrations Page

```
http://localhost:5173/integrations
```

**Verify**:
- [ ] Page loads without errors
- [ ] 4 tabs visible: Connected, Available, AI Models, Decisions
- [ ] Provider browser shows all providers
- [ ] Category filtering works (Communication, CRM, Calendar, etc.)

#### 2. Test OAuth Flow

**Steps**:
1. Click "Connect" on Slack integration
2. Complete OAuth on Slack's site
3. Verify redirect back to app works
4. Check "Connected" tab shows Slack

**Expected**:
- OAuth popup/redirect works
- Callback URL processes correctly
- Integration appears in Connected tab
- Connection status indicator shows "Connected"

#### 3. Integration Settings

**Steps**:
1. Click on connected integration
2. Navigate to settings page (`/integrations/[id]`)

**Verify**:
- [ ] Sync statistics displayed
- [ ] Last sync time shown
- [ ] Event count displayed
- [ ] Disconnect button works
- [ ] Sync trigger button works

#### 4. Test Multiple Providers

Repeat OAuth flow for:
- [ ] Google (Calendar + Gmail)
- [ ] Slack
- [ ] Linear
- [ ] Notion
- [ ] HubSpot

#### 5. Error Handling

**Test**:
- Cancel OAuth flow midway (should return to app gracefully)
- Disconnect and reconnect (should work smoothly)
- Try to connect with invalid credentials (should show error)

---

## 🔍 Polling Jobs Testing

### Verify Scheduled Jobs Created

```sql
-- Check scheduled jobs
SELECT job_type, cron_expression, is_active, next_run_at
FROM scheduled_jobs
ORDER BY next_run_at;
```

**Expected Jobs**:
- `sync_google_calendar` - Every 5 minutes
- `sync_slack_messages` - Every 2 minutes
- `sync_linear_issues` - Every 1 minute
- `sync_hubspot_contacts` - Every 5 minutes
- `sync_notion_pages` - Every 2 minutes
- `sync_clickup_tasks` - Every 2 minutes
- `sync_airtable_records` - Every 1 minute
- `sync_fathom_meetings` - Every 15 minutes
- `sync_microsoft_calendar` - Every 5 minutes

### Monitor Job Execution

```sql
-- Check background jobs
SELECT job_type, status, created_at, started_at, completed_at
FROM background_jobs
WHERE job_type LIKE 'sync_%'
ORDER BY created_at DESC
LIMIT 10;
```

### Test Webhook Bypass

1. Trigger a webhook (e.g., Linear issue update)
2. Wait for polling job to run
3. Verify job skipped due to recent webhook:

```bash
# Check logs
tail -f logs/app.log | grep "Skipping.*sync (recent webhook)"
```

---

## 📊 Test Coverage Summary

### Unit Tests
- ✅ **Webhook Signatures**: 25+ tests, 100% coverage
- ⏳ **Webhook Handlers**: TODO (requires test DB)
- ⏳ **Sync Service**: TODO (requires test DB)
- ⏳ **Conflict Resolution**: TODO

### Integration Tests
- ⏳ **OAuth Flows**: Manual testing required
- ⏳ **API Endpoints**: Manual testing required
- ⏳ **Database Sync**: Manual verification required

### E2E Tests
- ⏳ **Frontend UI**: Manual testing required
- ⏳ **Full Sync Flow**: OAuth → Webhook → DB → UI

---

## 🚀 Quick Test Script

```bash
#!/bin/bash
# Quick integration test script

echo "1. Testing webhook signature verification..."
go test -v ./internal/webhooks -run TestVerify

echo "2. Building project..."
go build -o bin/server ./cmd/server

echo "3. Starting server..."
./bin/server &
SERVER_PID=$!

echo "4. Waiting for server to start..."
sleep 3

echo "5. Testing API endpoints..."
curl -s http://localhost:8001/api/integrations/providers | jq .

echo "6. Cleanup..."
kill $SERVER_PID

echo "Done! ✅"
```

---

## 📝 Test Results Log

### Test Run: 2026-01-19

**Webhook Signature Tests**:
```
✅ TestVerifySlackSignature - PASS
✅ TestVerifyLinearSignature - PASS
✅ TestVerifyHubSpotSignature - PASS
✅ TestVerifyNotionSignature - PASS
✅ TestVerifyAirtableSignature - PASS
✅ TestVerifyFathomSignature - PASS
✅ TestVerifyGoogleChannelToken - PASS
✅ TestVerifyMicrosoftClientState - PASS
✅ TestVerifyHMACSHA256 - PASS
✅ TestComputeHMACSHA256 - PASS
✅ TestSignatureVerifier_SetGetSecret - PASS
```

**Total**: 11 test suites, 25+ individual tests
**Status**: ✅ ALL PASS
**Coverage**: Signature verification - 100%

---

## 🔧 Troubleshooting

### Common Issues

**Issue**: OAuth callback 404
**Fix**: Ensure callback URLs are registered in provider dashboard

**Issue**: Webhook signature verification fails
**Fix**: Check `WEBHOOK_SECRET_<PROVIDER>` in .env file

**Issue**: Polling jobs not running
**Fix**: Verify scheduler started in main.go

**Issue**: Sync data not appearing
**Fix**: Check `oauth_tokens` table for valid, non-expired tokens

---

## 📚 References

- **OAuth Implementation**: `internal/integrations/*/provider.go`
- **Webhook Handlers**: `internal/webhooks/handler.go`
- **Sync Service**: `internal/services/sync_service.go`
- **Polling Jobs**: `internal/jobs/sync_jobs.go`
- **Database Schema**: `internal/database/migrations/053_sync_tables.sql`
