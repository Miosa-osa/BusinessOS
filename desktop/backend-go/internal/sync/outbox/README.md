# Transactional Outbox Pattern Implementation

## Overview

This package implements the **Transactional Outbox Pattern** for reliable event publishing in a distributed system. It ensures exactly-once semantics by writing events to a database table within the same transaction as business logic, then asynchronously publishing them to NATS JetStream.

## Architecture

```
┌─────────────────┐
│ Business Logic  │
│                 │
│  BEGIN TX       │
│   ├─ Update DB  │
│   └─ Write      │◄─── Atomic operation (both succeed or both fail)
│      Outbox     │
│  COMMIT TX      │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│ Outbox Table    │
│  (Pending)      │
└─────────────────┘
         │
         ▼
┌─────────────────┐     ┌──────────────┐
│ Processor       │────▶│ NATS         │
│ (Background)    │     │ JetStream    │
│                 │     └──────────────┘
│ - Poll DB       │
│ - Publish NATS  │     ┌──────────────┐
│ - Retry Logic   │────▶│ Dead Letter  │
│ - DLQ Handling  │     │ Queue (DLQ)  │
└─────────────────┘     └──────────────┘
```

## Components

### 1. Writer (`writer.go`)

Responsible for writing events to the outbox table **within a transaction**.

**Key Methods:**
- `Write(ctx, tx, WriteRequest)` - Write event within existing transaction (recommended)
- `WriteWithPool(ctx, WriteRequest)` - Write event with its own transaction (convenience)

**Example Usage:**

```go
// Within a transaction
tx, _ := pool.Begin(ctx)
defer tx.Rollback(ctx)

// Update business data
_, err := tx.Exec(ctx, "UPDATE users SET name = $1 WHERE id = $2", newName, userID)
if err != nil {
    return err
}

// Write sync event in same transaction
writer := outbox.NewWriter(pool)
event, err := writer.Write(ctx, tx, outbox.WriteRequest{
    AggregateType: outbox.AggregateTypeUser,
    AggregateID:   userID,
    EventType:     outbox.EventTypeUpdated,
    Payload: map[string]interface{}{
        "id":   userID.String(),
        "name": newName,
    },
})
if err != nil {
    return err
}

// Commit both changes atomically
return tx.Commit(ctx)
```

### 2. Processor (`processor.go`)

Background worker that polls the outbox table and publishes events to NATS JetStream.

**Key Features:**
- ✅ Batch processing (100 messages per poll)
- ✅ Polling interval (1 second by default)
- ✅ Exponential backoff retry with jitter
- ✅ Max 5 retry attempts
- ✅ `FOR UPDATE SKIP LOCKED` for concurrent safety
- ✅ Automatic DLQ handling
- ✅ Prometheus metrics integration

**Retry Schedule:**

| Attempt | Delay (Center) | Delay (With ±25% Jitter) |
|---------|----------------|--------------------------|
| 0       | Immediate      | 0s                       |
| 1       | 1s             | 750ms - 1.25s            |
| 2       | 2s             | 1.5s - 2.5s              |
| 3       | 4s             | 3s - 5s                  |
| 4       | 8s             | 6s - 10s                 |
| 5       | 16s            | 12s - 20s                |
| 6+      | → DLQ          | Max attempts exceeded    |

**Formula:**
```
delay = min(initialBackoff * multiplier^(attempt-1), maxBackoff) ± (jitter * delay)
```

**Example Usage:**

```go
// Create processor
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
natsClient := messaging.NewNATSClient(natsURL)
processor := outbox.NewProcessor(pool, natsClient, logger)

// Start processing in background
go func() {
    if err := processor.Start(ctx); err != nil {
        logger.Error("processor failed", "error", err)
    }
}()

// Graceful shutdown
processor.Stop() // Waits for current batch to complete
```

### 3. Types (`types.go`)

Defines core data structures:

**Event:**
- `ID` - Unique event identifier
- `AggregateType` - Entity type (user, workspace, app, etc.)
- `AggregateID` - Entity identifier
- `EventType` - Change type (created, updated, deleted, restored)
- `Payload` - Full entity data as JSON
- `VectorClock` - Logical timestamp for conflict detection
- `Status` - Processing status (pending, processing, completed, failed)
- `Attempts` - Number of processing attempts
- `LastError` - Last error message
- `ScheduledFor` - Retry schedule timestamp

**Enums:**
- `SyncStatus`: `pending`, `processing`, `completed`, `failed`
- `AggregateType`: `user`, `workspace`, `app`, `project`, `task`
- `EventType`: `created`, `updated`, `deleted`, `restored`

## Database Schema

### Outbox Table (`sync_outbox`)

```sql
CREATE TABLE sync_outbox (
    id UUID PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    vector_clock JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    scheduled_for TIMESTAMPTZ
);

-- Critical index for processor queries
CREATE INDEX idx_sync_outbox_pending ON sync_outbox (created_at)
WHERE status = 'pending' AND (scheduled_for IS NULL OR scheduled_for <= NOW());
```

### Dead Letter Queue (`sync_dlq`)

```sql
CREATE TABLE sync_dlq (
    id UUID PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    vector_clock JSONB NOT NULL,
    attempts INT NOT NULL,
    last_error TEXT NOT NULL,
    failure_reason VARCHAR(255),
    original_created_at TIMESTAMPTZ NOT NULL,
    moved_to_dlq_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT
);
```

## Configuration

### ProcessorConfig

```go
type ProcessorConfig struct {
    BatchSize         int           // Messages per poll (default: 100)
    PollInterval      time.Duration // Poll frequency (default: 1s)
    MaxAttempts       int           // Max retries (default: 5)
    InitialBackoff    time.Duration // Initial delay (default: 1s)
    MaxBackoff        time.Duration // Max delay (default: 5min)
    BackoffMultiplier float64       // Backoff multiplier (default: 2.0)
    JitterFactor      float64       // Jitter ±% (default: 0.25)
}
```

### Custom Configuration Example

```go
config := outbox.ProcessorConfig{
    BatchSize:         50,
    PollInterval:      2 * time.Second,
    MaxAttempts:       3,
    InitialBackoff:    500 * time.Millisecond,
    MaxBackoff:        1 * time.Minute,
    BackoffMultiplier: 2.0,
    JitterFactor:      0.1, // ±10% jitter
}

processor := outbox.NewProcessorWithConfig(pool, nats, logger, config)
```

## NATS Integration

### NATSPublisher Interface

```go
type NATSPublisher interface {
    Publish(ctx context.Context, subject string, data []byte) error
}
```

### Subject Format

Messages are published to subjects following this pattern:
```
businessos.<aggregate_type>.<event_type>
```

**Examples:**
- `businessos.user.created`
- `businessos.workspace.updated`
- `businessos.app.deleted`
- `businessos.task.restored`

### Message Format

Published messages are JSON-encoded `Event` objects:

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "aggregate_type": "user",
  "aggregate_id": "789e0123-e89b-12d3-a456-426614174000",
  "event_type": "created",
  "payload": {
    "id": "789e0123-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "full_name": "John Doe"
  },
  "vector_clock": {
    "businessos": 1
  },
  "status": "completed",
  "attempts": 0,
  "created_at": "2026-01-09T12:00:00Z",
  "processed_at": "2026-01-09T12:00:01Z"
}
```

## Metrics

The processor integrates with the `internal/sync/metrics` package for observability:

### Available Metrics

- `OutboxPendingEvents` - Current pending events count
- `OutboxProcessingEvents` - Currently processing events count
- `OutboxCompletedEvents` - Total completed events
- `OutboxFailedEvents` - Total failed events
- `ProcessingDuration` - Event processing latency (avg, p95, p99)
- `ErrorsTotal` - Errors by type
- `LastProcessedAt` - Last successful processing timestamp
- `LastErrorAt` - Last error timestamp

### Accessing Metrics

```go
metrics := metrics.GetMetrics()
snapshot := metrics.GetSnapshot()

fmt.Printf("Pending: %d\n", snapshot.OutboxPendingEvents)
fmt.Printf("Completed: %d\n", snapshot.OutboxCompletedEvents)
fmt.Printf("Failed: %d\n", snapshot.OutboxFailedEvents)
fmt.Printf("Avg Duration: %s\n", snapshot.AvgProcessingDuration)
fmt.Printf("P95 Duration: %s\n", snapshot.P95ProcessingDuration)
```

## Error Handling

### Transient Errors (Retry)

- Network timeouts
- NATS connection failures
- Temporary database issues
- Rate limiting (429)

### Permanent Errors (DLQ)

- Invalid JSON payload
- Invalid vector clock
- Exceeded max attempts
- Non-retryable business errors

### Dead Letter Queue (DLQ)

Events that exceed max retry attempts are automatically moved to the DLQ with:
- Original event data preserved
- Full error history
- Manual resolution support
- Reprocessing capability

**Inspecting DLQ:**

```sql
-- View failed events
SELECT id, aggregate_type, event_type, attempts, last_error, moved_to_dlq_at
FROM sync_dlq
WHERE resolved = FALSE
ORDER BY moved_to_dlq_at DESC;

-- Mark as resolved
UPDATE sync_dlq
SET resolved = TRUE, resolved_at = NOW(), resolution_notes = 'Manually reprocessed'
WHERE id = '...';
```

## Production Considerations

### Scaling

1. **Multiple Processor Instances**
   - Safe to run multiple processors concurrently
   - `FOR UPDATE SKIP LOCKED` prevents duplicate processing
   - Increases throughput linearly

2. **Batch Size Tuning**
   - Larger batches = higher throughput, more memory
   - Smaller batches = lower latency, less memory
   - Recommended: 50-200 messages per batch

3. **Polling Interval**
   - Faster polling = lower latency, more DB load
   - Slower polling = higher latency, less DB load
   - Recommended: 1-5 seconds

### Monitoring

**Critical Alerts:**

1. **High Pending Count** (>1000)
   - Indicates processor falling behind
   - Action: Scale up processors or increase batch size

2. **High Failure Rate** (>5%)
   - Indicates NATS or network issues
   - Action: Investigate error logs and NATS health

3. **DLQ Growth** (>10 per hour)
   - Indicates systemic issues
   - Action: Investigate common failure patterns

4. **Processing Latency** (P95 >5s)
   - Indicates performance degradation
   - Action: Optimize queries or scale database

### Maintenance

**Cleanup Completed Events:**

```sql
-- Run daily via cron job
DELETE FROM sync_outbox
WHERE status = 'completed'
AND processed_at < NOW() - INTERVAL '7 days';
```

Or use the provided function:
```sql
SELECT cleanup_old_sync_outbox();
```

**Reprocess DLQ Events:**

```sql
-- Investigate root cause, fix issue, then:
INSERT INTO sync_outbox (
    id, aggregate_type, aggregate_id, event_type,
    payload, vector_clock, status, attempts
)
SELECT
    gen_random_uuid(), -- New ID to avoid conflicts
    aggregate_type, aggregate_id, event_type,
    payload, vector_clock, 'pending', 0
FROM sync_dlq
WHERE id = '...' AND resolved = FALSE;

-- Mark as resolved
UPDATE sync_dlq SET resolved = TRUE, resolution_notes = '...' WHERE id = '...';
```

## Testing

### Unit Tests

```bash
go test ./internal/sync/outbox -v
```

### Integration Tests

```bash
# Requires running PostgreSQL and NATS
go test ./internal/sync/outbox -tags=integration -v
```

### Test Coverage

```bash
go test ./internal/sync/outbox -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## References

- [Transactional Outbox Pattern](https://microservices.io/patterns/data/transactional-outbox.html)
- [NATS JetStream Documentation](https://docs.nats.io/nats-concepts/jetstream)
- [Exponential Backoff Algorithm](https://en.wikipedia.org/wiki/Exponential_backoff)
- [PostgreSQL FOR UPDATE SKIP LOCKED](https://www.postgresql.org/docs/current/sql-select.html#SQL-FOR-UPDATE-SHARE)

## License

Internal BusinessOS implementation - Not for external distribution.
