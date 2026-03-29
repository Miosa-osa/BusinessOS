# Sync Package - Phase 3 Bi-Directional Synchronization

This package implements the bi-directional synchronization between BusinessOS and OSA using the **transactional outbox + event streaming** pattern.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     BusinessOS Backend (Go)                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────┐    ┌──────────────┐    ┌─────────────┐        │
│  │  Business  │───▶│    Outbox    │───▶│  Processor  │───┐    │
│  │   Logic    │    │    Writer    │    │ (Background)│   │    │
│  └────────────┘    └──────────────┘    └─────────────┘   │    │
│                           │                               │    │
│                           │                               │    │
│                           ▼                               ▼    │
│                    ┌─────────────┐              ┌──────────┐   │
│                    │  PostgreSQL │              │   NATS   │   │
│                    │   Outbox    │              │JetStream │   │
│                    │    Table    │              └──────────┘   │
│                    └─────────────┘                      │       │
│                                                         │       │
└─────────────────────────────────────────────────────────┼───────┘
                                                          │
                                                          │
                                      ┌───────────────────▼────────┐
                                      │  NATS JetStream (Global)   │
                                      │  - Streams per aggregate   │
                                      │  - Message persistence     │
                                      │  - Exactly-once delivery   │
                                      └───────────────┬────────────┘
                                                      │
                                                      │
┌─────────────────────────────────────────────────────┼───────────┐
│                           OSA (Node.js)             │           │
│                                                     │           │
│                                              ┌──────▼──────┐    │
│                                              │    NATS     │    │
│                                              │  Subscriber │    │
│                                              └──────┬──────┘    │
│                                                     │           │
│                                                     ▼           │
│                                          ┌──────────────────┐   │
│                                          │   Conflict       │   │
│                                          │   Detector       │   │
│                                          └──────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
internal/sync/
├── outbox/
│   ├── types.go        # Event and conflict type definitions
│   ├── writer.go       # Outbox event writer (transactional)
│   └── processor.go    # Background processor (stub - TODO)
├── vectorclock/
│   ├── clock.go        # Vector clock implementation
│   └── clock_test.go   # Vector clock tests
├── messaging/
│   └── nats.go         # NATS JetStream client (stub - TODO)
├── conflicts/
│   └── detector.go     # Conflict detection logic (stub - TODO)
├── metrics/
│   ├── metrics.go      # Prometheus metrics
│   └── metrics_test.go # Metrics tests
└── README.md           # This file
```

## Core Components

### 1. Outbox Writer (`outbox/writer.go`)

**Purpose**: Write synchronization events to the outbox within the same database transaction as business logic.

**Key Features**:
- ✅ Transactional writes (exactly-once semantics)
- ✅ Vector clock integration for conflict detection
- ✅ Support for all aggregate types (user, workspace, app, project, task)
- ✅ Deferred processing for retry scenarios
- ✅ Status tracking (pending, processing, completed, failed)

**Usage Example**:
```go
writer := outbox.NewWriter(pool)

// Within a transaction
tx, _ := pool.Begin(ctx)
defer tx.Rollback(ctx)

// Update workspace in database
_, err := tx.Exec(ctx, "UPDATE workspaces SET name = $1 WHERE id = $2", newName, wsID)
if err != nil {
    return err
}

// Write sync event in same transaction
event, err := writer.Write(ctx, tx, outbox.WriteRequest{
    AggregateType: outbox.AggregateTypeWorkspace,
    AggregateID:   wsID,
    EventType:     outbox.EventTypeUpdated,
    Payload:       workspaceData,
})
if err != nil {
    return err
}

// Commit both changes atomically
tx.Commit(ctx)
```

### 2. Vector Clock (`vectorclock/clock.go`)

**Purpose**: Implement logical timestamps for distributed event ordering and conflict detection.

**Key Features**:
- ✅ Increment operation for local events
- ✅ Merge operation for remote events
- ✅ Compare operation for causality detection
- ✅ JSON serialization/deserialization
- ✅ Comprehensive test coverage

**Usage Example**:
```go
// Create new vector clock
vc := vectorclock.New()

// Increment for local event
vc.Increment("businessos")

// Compare with remote clock
remoteVC := vectorclock.FromMap(remoteEvent.VectorClock)
switch vc.Compare(remoteVC) {
case -1:
    // Local is before remote - apply update
case 1:
    // Local is after remote - ignore stale update
case 0:
    // Concurrent - conflict!
}

// Merge remote clock
vc.Merge(remoteVC)
```

### 3. Metrics (`metrics/metrics.go`)

**Purpose**: Track synchronization metrics for monitoring and alerting.

**Key Metrics**:
- ✅ Outbox pending events (gauge)
- ✅ Outbox processing events (gauge)
- ✅ Outbox completed events (counter)
- ✅ Outbox failed events (counter)
- ✅ Processing duration (histogram: avg, p95, p99)
- ✅ Errors total (counter by type)
- ✅ Conflicts total (counter by type)
- ✅ Last processed timestamp
- ✅ Last error timestamp

**Usage Example**:
```go
m := metrics.GetMetrics()

// Track event processing
m.IncrementProcessingEvents()
defer m.DecrementProcessingEvents()

start := time.Now()
// ... process event ...
m.RecordProcessingDuration(time.Since(start))

if err != nil {
    m.IncrementError("network_error")
    m.IncrementFailedEvents()
} else {
    m.IncrementCompletedEvents()
}

// Get snapshot for reporting
snapshot := m.GetSnapshot()
```

## Implementation Status

### ✅ Completed
- [x] Core types and enums (SyncStatus, AggregateType, EventType, ConflictType)
- [x] Event and Conflict struct definitions
- [x] OutboxWriter with transactional support
- [x] VectorClock implementation with full test coverage
- [x] Metrics system with comprehensive tracking

### 🚧 Pending Specification Decisions
- [ ] **OutboxProcessor** - Waiting for Q7 (retry policy)
- [ ] **NATS Client** - Waiting for Q1-Q3 (endpoint strategy, connection management)
- [ ] **Conflict Detector** - Waiting for Q5 (resolution strategy)

## Database Schema Requirements

The sync package requires the following database table:

```sql
CREATE TABLE sync_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    vector_clock JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    scheduled_for TIMESTAMP,

    INDEX idx_outbox_status (status, scheduled_for),
    INDEX idx_outbox_aggregate (aggregate_type, aggregate_id)
);
```

## Configuration

**Environment Variables** (to be added):
```bash
# NATS Configuration
NATS_ENDPOINT=nats://localhost:4222
NATS_CLUSTER_ID=businessos-sync
NATS_CLIENT_ID=businessos-backend-1

# Sync Configuration
SYNC_POLL_INTERVAL=5s
SYNC_BATCH_SIZE=100
SYNC_MAX_RETRIES=5
SYNC_RETRY_BACKOFF=exponential

# Node Identification
SYNC_NODE_ID=businessos  # or instance-specific ID for multi-instance
```

## Next Steps

1. **Await Specification Decisions**:
   - Q1-Q3: NATS endpoint and connection strategy
   - Q5: Conflict resolution approach
   - Q7: Retry policy

2. **Implement Remaining Components**:
   - OutboxProcessor with background polling
   - NATS JetStream client
   - Conflict detector and resolver

3. **Integration**:
   - Add migration for `sync_outbox` table
   - Integrate OutboxWriter into existing services (workspace, user, app, project, task)
   - Set up background processor job
   - Configure monitoring and alerting

4. **Testing**:
   - Integration tests with PostgreSQL
   - End-to-end sync tests with OSA
   - Conflict scenario testing
   - Performance benchmarking

## Design Principles

1. **Exactly-Once Semantics**: Events are written transactionally with business logic
2. **Conflict Detection**: Vector clocks track causality and detect concurrent updates
3. **Fault Tolerance**: Failed events are retried with exponential backoff
4. **Observability**: Comprehensive metrics for monitoring and alerting
5. **Idempotency**: Event processing is idempotent (safe to retry)
6. **Scalability**: Horizontal scaling via NATS JetStream

## References

- [Phase 3 Sync Design](../../../../docs/architecture/OSA_PHASE3_SYNC_DESIGN.md)
- [Transactional Outbox Pattern](https://microservices.io/patterns/data/transactional-outbox.html)
- [Vector Clocks](https://en.wikipedia.org/wiki/Vector_clock)
- [NATS JetStream](https://docs.nats.io/nats-concepts/jetstream)
