# NATS JetStream Client for BusinessOS Sync

This package provides a production-ready NATS JetStream client for bidirectional synchronization between BusinessOS and OSA-5.

## Features

- **JetStream Persistent Storage**: File-based storage with configurable retention
- **Automatic Stream Creation**: Creates `BUSINESSOS_SYNC` and `OSA_EVENTS` streams
- **Idempotent Publishing**: Uses message IDs to prevent duplicate processing
- **Durable Consumers**: Pull-based subscriptions with explicit acknowledgment
- **Automatic Retry**: Failed messages retry up to 5 times with 30s timeout
- **Connection Management**: Automatic reconnection with exponential backoff
- **Monitoring**: Built-in statistics and health checks

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      NATS JetStream                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Stream: BUSINESSOS_SYNC                                    │
│  ├── businessos.users.created.{uuid}                        │
│  ├── businessos.users.updated.{uuid}                        │
│  ├── businessos.workspaces.created.{uuid}                   │
│  ├── businessos.workspaces.updated.{uuid}                   │
│  └── businessos.apps.*                                      │
│                                                             │
│  Stream: OSA_EVENTS                                         │
│  ├── osa.builds.started.{build_id}                          │
│  ├── osa.builds.progress.{build_id}                         │
│  ├── osa.builds.completed.{build_id}                        │
│  └── osa.apps.generated.{uuid}                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Installation

```bash
go get github.com/nats-io/nats.go@latest
```

## Quick Start

### 1. Connect to NATS

```go
import (
    "log/slog"
    "github.com/rhl/businessos-backend/internal/sync/messaging"
)

logger := slog.Default()
client, err := messaging.NewNATSClient("nats://localhost:4222", logger)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### 2. Create Streams

```go
ctx := context.Background()
if err := client.CreateStreams(ctx); err != nil {
    log.Fatal(err)
}
```

### 3. Publish Events

```go
userID := uuid.New()
event := UserCreatedEvent{
    UserID:    userID,
    Email:     "user@example.com",
    FullName:  "John Doe",
    CreatedAt: time.Now(),
}

subject := messaging.SubjectUserCreated(userID)
if err := client.Publish(ctx, subject, event); err != nil {
    log.Fatal(err)
}
```

### 4. Subscribe to Events

```go
handler := func(msg *messaging.Message) error {
    log.Printf("Received: %s", msg.Subject)

    var event UserCreatedEvent
    if err := json.Unmarshal(msg.Data, &event); err != nil {
        return err
    }

    // Process event
    log.Printf("User created: %s", event.Email)

    return nil // Ack message
}

if err := client.Subscribe(ctx, messaging.SubjectAllUsers, handler); err != nil {
    log.Fatal(err)
}
```

## Subject Patterns

### BusinessOS Events (BusinessOS → OSA)

| Pattern | Description | Example |
|---------|-------------|---------|
| `businessos.users.created.{uuid}` | User created | `businessos.users.created.550e8400-e29b-41d4-a716-446655440000` |
| `businessos.users.updated.{uuid}` | User updated | `businessos.users.updated.550e8400-e29b-41d4-a716-446655440000` |
| `businessos.workspaces.created.{uuid}` | Workspace created | `businessos.workspaces.created.660e9500-e29b-41d4-a716-446655440001` |
| `businessos.workspaces.updated.{uuid}` | Workspace updated | `businessos.workspaces.updated.660e9500-e29b-41d4-a716-446655440001` |

### OSA Events (OSA → BusinessOS)

| Pattern | Description | Example |
|---------|-------------|---------|
| `osa.builds.started.{build_id}` | Build started | `osa.builds.started.build-123` |
| `osa.builds.progress.{build_id}` | Build progress | `osa.builds.progress.build-123` |
| `osa.builds.completed.{build_id}` | Build completed | `osa.builds.completed.build-123` |
| `osa.apps.generated.{uuid}` | App generated | `osa.apps.generated.770e9600-e29b-41d4-a716-446655440002` |

### Wildcard Subscriptions

| Pattern | Description |
|---------|-------------|
| `businessos.users.*` | All user events |
| `businessos.workspaces.*` | All workspace events |
| `osa.builds.*` | All build events |
| `osa.apps.*` | All app events |
| `businessos.>` | All BusinessOS events |
| `osa.>` | All OSA events |

## Configuration

### Stream Configuration

```go
StreamConfig{
    Name:      "BUSINESSOS_SYNC",
    Storage:   jetstream.FileStorage,     // Persistent
    Retention: jetstream.LimitsPolicy,    // Time/size limits
    MaxAge:    7 * 24 * time.Hour,       // 7 days
    MaxBytes:  1024 * 1024 * 1024,       // 1GB
    Discard:   jetstream.DiscardOld,     // Remove oldest when full
    Replicas:  1,                         // Single node
}
```

### Consumer Configuration

```go
ConsumerConfig{
    Durable:       "consumer_name",           // Survives restarts
    AckPolicy:     jetstream.AckExplicitPolicy, // Must explicitly ack
    AckWait:       30 * time.Second,          // Wait 30s for ack
    MaxDeliver:    5,                         // Retry up to 5 times
    MaxAckPending: 1000,                      // Max unacked messages
}
```

### Connection Options

```go
nats.Options{
    Name:              "BusinessOS-Sync-Client",
    ReconnectWait:     2 * time.Second,
    MaxReconnects:     -1,  // Infinite
    PingInterval:      20 * time.Second,
    MaxPingsOutstanding: 3,
}
```

## Message Flow

### Publishing (BusinessOS → NATS)

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  BusinessOS  │────▶│     NATS     │────▶│     OSA      │
│   Publisher  │     │  JetStream   │     │   Consumer   │
└──────────────┘     └──────────────┘     └──────────────┘
     │                     │                     │
     │ 1. Publish          │                     │
     │────────────────────▶│                     │
     │                     │ 2. Store (disk)     │
     │                     │─────────▶           │
     │ 3. Ack              │                     │
     │◀────────────────────│                     │
     │                     │ 4. Deliver          │
     │                     │────────────────────▶│
     │                     │ 5. Ack              │
     │                     │◀────────────────────│
```

### Consuming (NATS → BusinessOS)

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│     OSA      │────▶│     NATS     │────▶│  BusinessOS  │
│   Publisher  │     │  JetStream   │     │   Consumer   │
└──────────────┘     └──────────────┘     └──────────────┘
                           │                     │
                           │ 1. Pull request     │
                           │◀────────────────────│
                           │ 2. Deliver          │
                           │────────────────────▶│
                           │                     │ 3. Process
                           │                     │
                           │ 4. Ack/Nak          │
                           │◀────────────────────│
```

## Error Handling

### Automatic Retry

Messages that fail processing are automatically retried:

```go
handler := func(msg *messaging.Message) error {
    // If this returns an error, NATS will redeliver
    if err := processMessage(msg); err != nil {
        return err  // Nak (negative ack)
    }
    return nil  // Ack (acknowledge)
}
```

**Retry Behavior:**
- Failed messages are redelivered after `AckWait` (30s)
- Maximum `MaxDeliver` (5) attempts
- After 5 failures, message goes to dead letter queue (if configured)

### Connection Recovery

The client automatically reconnects on connection loss:

```go
// DisconnectErrHandler is called on disconnect
nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
    log.Warn("NATS disconnected", "error", err)
})

// ReconnectHandler is called on successful reconnect
nats.ReconnectHandler(func(nc *nats.Conn) {
    log.Info("NATS reconnected")
})
```

## Monitoring

### Health Check

```go
if !client.IsConnected() {
    log.Error("NATS client disconnected!")
}
```

### Statistics

```go
stats := client.Stats()
log.Info("NATS stats",
    "connected", stats.Connected,
    "in_msgs", stats.InMsgs,
    "out_msgs", stats.OutMsgs,
    "reconnects", stats.Reconnects,
    "active_subs", stats.ActiveSubscriptions)
```

### Prometheus Metrics (recommended)

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    messagesPublished = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "nats_messages_published_total",
            Help: "Total messages published",
        },
        []string{"subject"},
    )

    messagesReceived = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "nats_messages_received_total",
            Help: "Total messages received",
        },
        []string{"subject"},
    )
)
```

## Testing

### Unit Tests

Run unit tests (no NATS server required):

```bash
go test -v ./internal/sync/messaging/...
```

### Integration Tests

Run integration tests (requires NATS server):

```bash
# Start NATS with JetStream
docker run -d --name nats -p 4222:4222 nats:latest -js

# Run tests
export NATS_URL=nats://localhost:4222
go test -v ./internal/sync/messaging/...
```

### Benchmarks

```bash
go test -bench=. ./internal/sync/messaging/
```

## Production Deployment

### Docker Compose

```yaml
version: '3.8'
services:
  nats:
    image: nats:latest
    command: ["-js", "-sd", "/data"]
    ports:
      - "4222:4222"
      - "8222:8222"  # HTTP monitoring
    volumes:
      - nats-data:/data
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8222/healthz"]
      interval: 10s
      timeout: 5s
      retries: 3

  businessos:
    image: businessos:latest
    environment:
      - NATS_URL=nats://nats:4222
    depends_on:
      nats:
        condition: service_healthy

volumes:
  nats-data:
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nats
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: nats
        image: nats:latest
        args: ["-js", "-sd", "/data"]
        ports:
        - containerPort: 4222
        - containerPort: 8222
        volumeMounts:
        - name: data
          mountPath: /data
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8222
          initialDelaySeconds: 10
          periodSeconds: 10
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

### Environment Variables

```bash
# NATS connection
NATS_URL=nats://localhost:4222
NATS_CLUSTER_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222

# Optional: Authentication
NATS_USERNAME=businessos
NATS_PASSWORD=secret
NATS_TOKEN=your-auth-token

# Optional: TLS
NATS_TLS_CERT=/path/to/cert.pem
NATS_TLS_KEY=/path/to/key.pem
NATS_TLS_CA=/path/to/ca.pem
```

## Performance

### Benchmarks (local testing)

| Operation | Throughput | Latency (p99) |
|-----------|-----------|---------------|
| Publish | ~50,000 msg/s | <2ms |
| Subscribe | ~45,000 msg/s | <5ms |
| Round-trip | ~20,000 msg/s | <10ms |

### Tuning

```go
// Increase batch size for consumers
ConsumerConfig{
    MaxAckPending: 10000,  // Default: 1000
}

// Increase buffer sizes
nats.ReconnectBufSize(8 * 1024 * 1024)  // 8MB
```

## Troubleshooting

### Consumer Not Receiving Messages

1. Check stream exists: `nats stream list`
2. Check consumer exists: `nats consumer list BUSINESSOS_SYNC`
3. Check filter subject matches published subject
4. Check consumer is not at max unacked messages

### Messages Not Being Acknowledged

1. Ensure handler returns `nil` for successful processing
2. Check `AckWait` timeout is sufficient
3. Verify no panics in handler (use recover)

### High Memory Usage

1. Reduce `MaxAckPending` in consumer config
2. Implement backpressure in handler
3. Process messages in batches

### Connection Keeps Dropping

1. Check network stability
2. Increase `PingInterval` and `MaxPingsOutstanding`
3. Verify firewall allows NATS port (4222)

## Additional Resources

- [NATS Documentation](https://docs.nats.io/)
- [JetStream Guide](https://docs.nats.io/nats-concepts/jetstream)
- [Go Client Documentation](https://pkg.go.dev/github.com/nats-io/nats.go)
- [BusinessOS Sync Architecture](../../../docs/architecture/OSA_PHASE3_SYNC_DESIGN.md)

## License

Copyright © 2026 BusinessOS
