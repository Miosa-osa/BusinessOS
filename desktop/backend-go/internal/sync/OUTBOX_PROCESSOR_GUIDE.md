# OutboxProcessor Implementation Guide

Complete guide for the `OutboxProcessor` implementation for Phase 3 OSA integration.

## Overview

The `OutboxProcessor` ensures that all sync events are reliably delivered to OSA-5, even in the face of network failures, OSA downtime, or system crashes. It provides:

- **At-least-once delivery** with idempotency support
- **Automatic retry** with exponential backoff (Q7 specification)
- **Dead Letter Queue (DLQ)** for failed events requiring manual intervention
- **Concurrent processing** with configurable worker pool
- **Monitoring and observability** through stats API

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    TRANSACTIONAL OUTBOX                     │
└─────────────────────────────────────────────────────────────┘

1. APPLICATION CREATES ENTITY
   └─▶ BEGIN TRANSACTION
       ├─▶ INSERT INTO users (...)
       ├─▶ INSERT INTO sync_outbox (aggregate_type='user', payload=...)
       └─▶ COMMIT TRANSACTION

2. OUTBOX PROCESSOR (Background Worker)
   └─▶ Poll sync_outbox every 5 seconds
       ├─▶ SELECT * FROM sync_outbox WHERE status='pending' FOR UPDATE SKIP LOCKED
       └─▶ Process events in parallel (4 workers)

3. EVENT PROCESSING
   ├─▶ Mark event as 'processing'
   ├─▶ Call OSA API (POST /api/users/sync)
   │   ├─▶ Success → Mark as 'completed'
   │   └─▶ Failure → Retry with exponential backoff
   └─▶ After 5 retries → Move to Dead Letter Queue (DLQ)

4. RETRY SCHEDULE (Q7 from SYNC_SPECIFICATION_ANSWERS.md)
   ├─▶ Retry 0: Immediate
   ├─▶ Retry 1: 1 second delay
   ├─▶ Retry 2: 2 seconds delay
   ├─▶ Retry 3: 4 seconds delay
   └─▶ Retry 4: 8 seconds delay
```

## Usage

### 1. Initialize and Start

```go
import (
    "github.com/rhl/businessos-backend/internal/sync"
    "github.com/rhl/businessos-backend/internal/integrations/osa"
)

// Create processor with 4 workers, polling every 5 seconds
processor := sync.NewOutboxProcessor(
    dbPool,        // *pgxpool.Pool
    osaClient,     // *osa.Client
    4,             // Number of workers
    5*time.Second, // Polling interval
)

// Start processing
if err := processor.Start(ctx); err != nil {
    log.Fatal(err)
}

// Graceful shutdown
defer processor.Stop()
```

### 2. Create Events in Transactions

```go
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
    return s.db.BeginFunc(ctx, func(tx pgx.Tx) error {
        queries := sqlc.New(tx)

        // 1. Create user
        user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
            Email:    req.Email,
            FullName: req.FullName,
        })
        if err != nil {
            return err
        }

        // 2. Create outbox event for sync
        payload := sync.UserSyncPayload{
            UserID:   user.ID,
            Email:    user.Email,
            FullName: user.FullName,
        }
        payloadJSON, _ := json.Marshal(payload)

        vectorClock := map[string]int{"businessos": 1}
        vectorClockJSON, _ := json.Marshal(vectorClock)

        _, err = queries.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
            AggregateType: "user",
            AggregateID:   user.ID,
            EventType:     "user_created",
            Payload:       payloadJSON,
            VectorClock:   vectorClockJSON,
            MaxAttempts:   5,
        })

        return err
    })
}
```

### 3. Monitor Processing

```go
// Get statistics
stats, err := processor.GetStats(ctx)
if err != nil {
    log.Error("failed to get stats", "error", err)
    return
}

log.Info("outbox stats",
    "pending", stats.PendingCount,
    "processing", stats.ProcessingCount,
    "completed", stats.CompletedCount,
    "failed", stats.FailedCount,
    "dlq_ready", stats.DLQReadyCount,
)
```

## Retry Logic

| Attempt | Delay | Cumulative Time |
|---------|-------|-----------------|
| 0 | Immediate | 0s |
| 1 | 1 second | 1s |
| 2 | 2 seconds | 3s |
| 3 | 4 seconds | 7s |
| 4 | 8 seconds | 15s |
| 5 (DLQ) | - | Event moved to DLQ |

## References

- [OSA Phase 3 Sync Design](../../../../docs/architecture/OSA_PHASE3_SYNC_DESIGN.md)
- [Sync Specification Q7](../../../../docs/architecture/SYNC_SPECIFICATION_ANSWERS.md#q7-retry-policy-parameters)
- [Migration 043](../../database/migrations/043_sync_outbox.sql)
- [SQLC Queries](../../database/queries/sync_outbox.sql)
