# BusinessOS Live Sync Architecture

> **Status:** Design Complete, Implementation Pending
> **Priority:** P0 - Core Feature
> **Last Updated:** January 2026

---

## Overview

Live Sync enables **real-time bidirectional data synchronization** between BusinessOS and external services. When data changes in any connected service, it automatically syncs to BusinessOS and vice versa.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        LIVE SYNC DATA FLOW                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   External Services              BusinessOS                  Users      │
│   ┌─────────────┐               ┌──────────────┐          ┌─────────┐  │
│   │ Google Cal  │──webhook──────│              │          │         │  │
│   │ Slack       │──events───────│   Webhook    │          │         │  │
│   │ Notion      │──────────────▶│   Handler    │          │         │  │
│   │ Linear      │               │              │          │         │  │
│   │ HubSpot     │               └──────┬───────┘          │         │  │
│   │ ClickUp     │                      │                  │         │  │
│   │ Airtable    │                      ▼                  │         │  │
│   │ Fathom      │               ┌──────────────┐          │   UI    │  │
│   │ Microsoft   │               │  Event Bus   │──SSE────▶│         │  │
│   └─────────────┘               │  (Pub/Sub)   │          │         │  │
│          ▲                      └──────┬───────┘          │         │  │
│          │                             │                  │         │  │
│          │                             ▼                  │         │  │
│          │                      ┌──────────────┐          │         │  │
│          │                      │   Outbox     │          │         │  │
│          │                      │  Processor   │          │         │  │
│          │                      └──────┬───────┘          │         │  │
│          │                             │                  │         │  │
│          │                             ▼                  │         │  │
│          │                      ┌──────────────┐          │         │  │
│          └──────────────────────│   Database   │◀─────────│         │  │
│                                 │  (Postgres)  │          │         │  │
│                                 └──────────────┘          └─────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Current Infrastructure (70% Complete)

| Component | Status | Location |
|-----------|--------|----------|
| Webhook Handlers | ✅ Scaffolded | `internal/webhooks/handler.go` |
| Event Bus (Pub/Sub) | ✅ Complete | `internal/services/build_event_bus.go` |
| SSE Streaming | ✅ Complete | `internal/handlers/osa_streaming.go` |
| Transactional Outbox | ✅ Complete | `internal/sync/outbox/` |
| Vector Clocks | ✅ Complete | `internal/sync/vectorclock/` |
| NATS Integration | ❌ Missing | Needs implementation |
| Sync Business Logic | ❌ Missing | Needs implementation |
| Conflict Resolution | ❌ Missing | Needs implementation |
| Polling Jobs | ❌ Missing | Needs implementation |

---

## Webhook Capabilities by Provider

| Provider | Webhooks | Events | Signature | Notes |
|----------|----------|--------|-----------|-------|
| **Google Calendar** | ✅ Push | Events CRUD | Token | No payload - needs follow-up call |
| **Slack** | ✅ Events API | Messages, Channels, Users | HMAC-SHA256 | Full payload |
| **Notion** | ✅ Native | Pages, Databases, Comments | HMAC-SHA256 | Some events aggregated |
| **Linear** | ✅ Full | Issues, Comments, Projects | HMAC-SHA256 | Full payload with `previousData` |
| **HubSpot** | ✅ CRM | Contacts, Companies, Deals | HMAC-SHA256 v3 | No engagement webhooks |
| **ClickUp** | ✅ Full | Tasks, Lists, Spaces | Shared Secret | 30+ event types |
| **Airtable** | ✅ API | Records, Schema | HMAC-SHA256 | Per-table subscriptions |
| **Fathom** | ✅ API | Recordings, Transcripts | HMAC-SHA256 | Post-meeting only |
| **Microsoft 365** | ✅ Graph | Mail, Calendar, Files | clientState | Subscriptions expire |

---

## Implementation Plan

### Phase 1: Webhook Handler Completion (Priority: P0)

Complete the existing webhook handlers to actually sync data.

#### 1.1 Google Calendar Webhook Handler

```go
// internal/webhooks/google_calendar.go
func (h *Handler) HandleGoogleCalendarWebhook(c *gin.Context) {
    // 1. Verify channel token
    channelID := c.GetHeader("X-Goog-Channel-ID")
    resourceState := c.GetHeader("X-Goog-Resource-State")

    // 2. Fetch changed events (webhook has no payload)
    events, err := h.googleService.FetchChangedEvents(ctx, userID, syncToken)

    // 3. Sync each event to database
    for _, event := range events {
        err := h.syncService.SyncCalendarEvent(ctx, userID, event)

        // 4. Write to outbox for downstream processing
        h.outboxWriter.Write(ctx, tx, outbox.WriteRequest{
            AggregateType: "calendar_event",
            AggregateID:   event.ID,
            EventType:     "synced",
            Payload:       event,
        })

        // 5. Publish to event bus for real-time UI update
        h.eventBus.Publish(EventBusEvent{
            Type: "calendar.event.synced",
            Data: event,
        })
    }
}
```

#### 1.2 Slack Events Handler

```go
// internal/webhooks/slack_events.go
func (h *Handler) HandleSlackEvent(c *gin.Context) {
    // 1. Verify HMAC-SHA256 signature
    if !h.verifySlackSignature(c) {
        c.JSON(401, gin.H{"error": "Invalid signature"})
        return
    }

    // 2. Handle URL verification challenge
    if payload.Type == "url_verification" {
        c.String(200, payload.Challenge)
        return
    }

    // 3. Process event by type
    switch payload.Event.Type {
    case "message":
        h.syncSlackMessage(ctx, payload.Event)
    case "channel_created":
        h.syncSlackChannel(ctx, payload.Event)
    case "member_joined_channel":
        h.syncSlackMember(ctx, payload.Event)
    }

    // 4. Write to outbox + publish to event bus
    // ...
}
```

#### 1.3 Linear Webhook Handler

```go
// internal/webhooks/linear.go
func (h *Handler) HandleLinearWebhook(c *gin.Context) {
    // 1. Verify HMAC-SHA256 signature
    signature := c.GetHeader("Linear-Signature")
    if !h.verifyLinearSignature(c.Request.Body, signature) {
        c.JSON(401, gin.H{"error": "Invalid signature"})
        return
    }

    // 2. Validate timestamp (replay protection)
    timestamp := payload.WebhookTimestamp
    if time.Now().Unix()-timestamp > 60 {
        c.JSON(400, gin.H{"error": "Timestamp too old"})
        return
    }

    // 3. Process by action type
    switch payload.Action {
    case "create":
        h.createLocalIssue(ctx, payload.Data)
    case "update":
        h.updateLocalIssue(ctx, payload.Data, payload.PreviousData)
    case "remove":
        h.deleteLocalIssue(ctx, payload.Data.ID)
    }
}
```

---

### Phase 2: Database Schema for Synced Data

#### 2.1 Create Sync Tables

```sql
-- migrations/XXX_sync_tables.sql

-- Calendar events from all providers
CREATE TABLE synced_calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(50) NOT NULL, -- 'google', 'microsoft'
    external_id VARCHAR(255) NOT NULL,
    title TEXT,
    description TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    location TEXT,
    attendees JSONB,
    meeting_link TEXT,
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider, external_id)
);

-- Messages from Slack, Teams, etc.
CREATE TABLE synced_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(50) NOT NULL, -- 'slack', 'microsoft_teams'
    channel_id VARCHAR(255),
    channel_name VARCHAR(255),
    external_id VARCHAR(255) NOT NULL,
    sender_id VARCHAR(255),
    sender_name VARCHAR(255),
    content TEXT,
    thread_id VARCHAR(255),
    raw_data JSONB,
    sent_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider, external_id)
);

-- Tasks from Linear, ClickUp, Asana, etc.
CREATE TABLE synced_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(50) NOT NULL, -- 'linear', 'clickup', 'hubspot'
    external_id VARCHAR(255) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status VARCHAR(100),
    priority VARCHAR(50),
    assignee_id VARCHAR(255),
    assignee_name VARCHAR(255),
    project_id VARCHAR(255),
    project_name VARCHAR(255),
    due_date TIMESTAMPTZ,
    labels JSONB,
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider, external_id)
);

-- Contacts/Leads from HubSpot, etc.
CREATE TABLE synced_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(50) NOT NULL, -- 'hubspot', 'salesforce'
    external_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    phone VARCHAR(100),
    company VARCHAR(255),
    job_title VARCHAR(255),
    lifecycle_stage VARCHAR(100),
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider, external_id)
);

-- Files from Drive, OneDrive, Dropbox
CREATE TABLE synced_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(50) NOT NULL, -- 'google_drive', 'onedrive'
    external_id VARCHAR(255) NOT NULL,
    name VARCHAR(500),
    mime_type VARCHAR(255),
    size_bytes BIGINT,
    parent_folder_id VARCHAR(255),
    web_url TEXT,
    download_url TEXT,
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider, external_id)
);

-- Meeting recordings from Fathom, etc.
CREATE TABLE synced_meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(50) NOT NULL, -- 'fathom', 'fireflies'
    external_id VARCHAR(255) NOT NULL,
    title VARCHAR(500),
    start_time TIMESTAMPTZ,
    duration_seconds INT,
    participants JSONB,
    transcript TEXT,
    summary TEXT,
    action_items JSONB,
    recording_url TEXT,
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider, external_id)
);

-- Notion pages/databases
CREATE TABLE synced_notion_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    external_id VARCHAR(255) NOT NULL,
    title TEXT,
    parent_type VARCHAR(50), -- 'database', 'page', 'workspace'
    parent_id VARCHAR(255),
    properties JSONB,
    content_preview TEXT,
    url TEXT,
    raw_data JSONB,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, external_id)
);

-- Indexes for common queries
CREATE INDEX idx_synced_calendar_user_time ON synced_calendar_events(user_id, start_time);
CREATE INDEX idx_synced_tasks_user_status ON synced_tasks(user_id, status);
CREATE INDEX idx_synced_contacts_user_email ON synced_contacts(user_id, email);
CREATE INDEX idx_synced_messages_user_channel ON synced_messages(user_id, channel_id);
CREATE INDEX idx_synced_files_user_folder ON synced_files(user_id, parent_folder_id);
```

---

### Phase 3: Webhook Subscription Management

#### 3.1 Subscription Registry

```sql
-- Track active webhook subscriptions
CREATE TABLE webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(50) NOT NULL,
    resource_type VARCHAR(100), -- 'calendar', 'messages', 'issues'
    external_subscription_id VARCHAR(255),
    webhook_url TEXT,
    webhook_secret TEXT, -- Encrypted
    events JSONB, -- Array of subscribed event types
    expires_at TIMESTAMPTZ, -- For providers with expiring subscriptions
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'expired', 'failed'
    last_event_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, provider, resource_type)
);
```

#### 3.2 Subscription Service

```go
// internal/services/webhook_subscription_service.go
type WebhookSubscriptionService struct {
    pool            *pgxpool.Pool
    providers       map[string]WebhookProvider
    encryptionKey   []byte
}

// Create subscriptions when user connects integration
func (s *WebhookSubscriptionService) CreateSubscriptions(ctx context.Context, userID uuid.UUID, provider string) error {
    p, ok := s.providers[provider]
    if !ok {
        return fmt.Errorf("unknown provider: %s", provider)
    }

    // Get user's OAuth token
    token, err := s.getToken(ctx, userID, provider)
    if err != nil {
        return err
    }

    // Create webhook subscription via provider API
    sub, err := p.CreateWebhook(ctx, token, WebhookConfig{
        URL:    fmt.Sprintf("https://api.businessos.com/webhooks/%s", provider),
        Events: p.SupportedEvents(),
    })
    if err != nil {
        return err
    }

    // Store subscription with encrypted secret
    return s.storeSubscription(ctx, userID, provider, sub)
}

// Renew expiring subscriptions (for Microsoft, Google)
func (s *WebhookSubscriptionService) RenewExpiringSubscriptions(ctx context.Context) error {
    subs, err := s.getExpiringSubscriptions(ctx, 24*time.Hour)
    for _, sub := range subs {
        if err := s.renewSubscription(ctx, sub); err != nil {
            slog.Error("failed to renew subscription", "id", sub.ID, "error", err)
        }
    }
    return nil
}
```

---

### Phase 4: Polling Fallback Jobs

For providers without webhooks or as backup:

```go
// internal/jobs/sync_jobs.go

// Run every 5 minutes
func (j *SyncJobs) SyncGoogleCalendar(ctx context.Context) {
    users, err := j.getUsersWithIntegration(ctx, "google")
    for _, user := range users {
        events, syncToken, err := j.googleService.FetchEvents(ctx, user.ID, user.SyncToken)
        for _, event := range events {
            j.syncService.UpsertCalendarEvent(ctx, user.ID, "google", event)
        }
        j.updateSyncToken(ctx, user.ID, "google", syncToken)
    }
}

// Run every 2 minutes
func (j *SyncJobs) SyncSlackMessages(ctx context.Context) {
    // Similar pattern for Slack
}

// Run every 15 minutes
func (j *SyncJobs) SyncFathomMeetings(ctx context.Context) {
    // Similar pattern for Fathom
}

// Schedule jobs
func (j *SyncJobs) Start(ctx context.Context) {
    scheduler := gocron.NewScheduler(time.UTC)

    scheduler.Every(5).Minutes().Do(j.SyncGoogleCalendar, ctx)
    scheduler.Every(2).Minutes().Do(j.SyncSlackMessages, ctx)
    scheduler.Every(15).Minutes().Do(j.SyncFathomMeetings, ctx)
    scheduler.Every(24).Hours().Do(j.RenewWebhookSubscriptions, ctx)

    scheduler.StartBlocking()
}
```

---

### Phase 5: Bidirectional Sync (BusinessOS → External)

When user creates/updates in BusinessOS, sync to external service:

```go
// internal/services/outbound_sync_service.go

func (s *OutboundSyncService) CreateTask(ctx context.Context, userID uuid.UUID, task Task) error {
    // 1. Create in BusinessOS database
    localTask, err := s.taskRepo.Create(ctx, task)

    // 2. If user has Linear connected and wants sync
    if s.shouldSyncToLinear(ctx, userID) {
        linearIssue, err := s.linearService.CreateIssue(ctx, userID, LinearIssueInput{
            Title:       task.Title,
            Description: task.Description,
            TeamID:      task.ProjectID, // Mapped
        })

        // 3. Store mapping
        s.storeSyncMapping(ctx, localTask.ID, "linear", linearIssue.ID)
    }

    // 3. If user has ClickUp connected
    if s.shouldSyncToClickUp(ctx, userID) {
        // Similar pattern
    }

    return nil
}
```

---

### Phase 6: Conflict Resolution

```go
// internal/sync/conflicts/resolver.go

type ConflictResolver struct {
    strategy ConflictStrategy // "last_write_wins", "first_write_wins", "manual"
}

func (r *ConflictResolver) Resolve(ctx context.Context, local, remote SyncedEntity) (*SyncedEntity, error) {
    // Compare vector clocks
    comparison := local.VectorClock.Compare(remote.VectorClock)

    switch comparison {
    case vectorclock.Before:
        // Remote is newer, accept remote
        return &remote, nil
    case vectorclock.After:
        // Local is newer, keep local (may need to push to remote)
        return &local, nil
    case vectorclock.Concurrent:
        // True conflict - both modified independently
        return r.resolveConflict(ctx, local, remote)
    default:
        return &remote, nil
    }
}

func (r *ConflictResolver) resolveConflict(ctx context.Context, local, remote SyncedEntity) (*SyncedEntity, error) {
    switch r.strategy {
    case "last_write_wins":
        if local.UpdatedAt.After(remote.UpdatedAt) {
            return &local, nil
        }
        return &remote, nil
    case "manual":
        // Store conflict for user review
        r.storeConflict(ctx, local, remote)
        return nil, ErrConflictRequiresManualResolution
    default:
        return &remote, nil
    }
}
```

---

## Testing Strategy

### 1. Unit Tests for Webhook Handlers

```go
// internal/webhooks/handler_test.go
func TestSlackWebhookSignatureVerification(t *testing.T) {
    // Test valid signature passes
    // Test invalid signature fails
    // Test expired timestamp fails
}

func TestLinearWebhookProcessing(t *testing.T) {
    // Test issue created event
    // Test issue updated event
    // Test issue deleted event
}
```

### 2. Integration Tests with Mock Providers

```go
// internal/integrations/integration_test.go
func TestGoogleCalendarSync(t *testing.T) {
    // 1. Mock Google Calendar API
    // 2. Trigger webhook
    // 3. Verify event synced to database
    // 4. Verify SSE event published
}
```

### 3. E2E Tests with Real Providers (Manual)

```bash
# Test webhook endpoint directly
curl -X POST http://localhost:8001/api/webhooks/linear \
  -H "Content-Type: application/json" \
  -H "Linear-Signature: $(compute_hmac)" \
  -d '{"action":"create","type":"Issue","data":{...}}'

# Verify in database
psql -c "SELECT * FROM synced_tasks WHERE provider='linear' ORDER BY synced_at DESC LIMIT 1;"
```

---

## Monitoring & Observability

### Metrics to Track

```go
// internal/sync/metrics/sync_metrics.go
var (
    WebhooksReceived = promauto.NewCounterVec(
        prometheus.CounterOpts{Name: "webhooks_received_total"},
        []string{"provider", "event_type"},
    )

    WebhookProcessingDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{Name: "webhook_processing_duration_seconds"},
        []string{"provider"},
    )

    SyncErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{Name: "sync_errors_total"},
        []string{"provider", "error_type"},
    )

    ConflictsDetected = promauto.NewCounter(
        prometheus.CounterOpts{Name: "sync_conflicts_total"},
    )
)
```

### Logging

```go
slog.Info("webhook received",
    "provider", "linear",
    "event_type", "issue.created",
    "external_id", payload.Data.ID,
)

slog.Info("entity synced",
    "provider", "linear",
    "type", "task",
    "local_id", localTask.ID,
    "external_id", linearIssue.ID,
)
```

---

## Implementation Timeline

| Phase | Tasks | Est. Time |
|-------|-------|-----------|
| **Phase 1** | Webhook handler completion | 3-4 days |
| **Phase 2** | Database schema + migrations | 1 day |
| **Phase 3** | Webhook subscription management | 2 days |
| **Phase 4** | Polling fallback jobs | 2 days |
| **Phase 5** | Bidirectional sync | 3-4 days |
| **Phase 6** | Conflict resolution | 2 days |
| **Testing** | Unit + integration tests | 2-3 days |
| **Total** | | ~15-18 days |

---

## Key Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/webhooks/handler.go` | Modify | Complete sync logic |
| `internal/webhooks/signature.go` | Create | Signature verification |
| `internal/services/sync_service.go` | Create | Core sync business logic |
| `internal/services/webhook_subscription_service.go` | Create | Manage subscriptions |
| `internal/jobs/sync_jobs.go` | Create | Polling fallback |
| `internal/sync/conflicts/resolver.go` | Create | Conflict resolution |
| `migrations/XXX_sync_tables.sql` | Create | Sync data tables |
| `internal/webhooks/*_test.go` | Create | Tests |

---

*Architecture designed: January 2026*
