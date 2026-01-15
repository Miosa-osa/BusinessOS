# NATS Client for Phase 3 OSA Integration

## Overview

This NATS client provides real-time event streaming for bidirectional sync between BusinessOS and OSA. It implements the event-driven architecture defined in Phase 3 of the OSA integration.

## Features

- **Real-time Event Streaming**: Publish/Subscribe to NATS for instant sync notifications
- **Connection Management**: Automatic reconnection with exponential backoff
- **Event Types**: Support for workspace, project, and task sync events
- **Conflict Notifications**: Dedicated channel for conflict detection events
- **Graceful Degradation**: Can operate in disabled mode for development/testing
- **Health Monitoring**: Built-in health checks and connection statistics
- **Error Recovery**: Automatic retry logic and panic recovery

## Architecture

### NATS Subjects

The client uses a hierarchical subject namespace:

```
osa.sync.workspace.*     - All workspace events
  ├── osa.sync.workspace.created
  ├── osa.sync.workspace.updated
  └── osa.sync.workspace.deleted

osa.sync.project.*       - All project events
  ├── osa.sync.project.created
  ├── osa.sync.project.updated
  └── osa.sync.project.deleted

osa.sync.task.*          - All task events
  ├── osa.sync.task.created
  ├── osa.sync.task.updated
  └── osa.sync.task.deleted

osa.sync.conflict        - Conflict notifications
```

### Connection Settings

Default connection parameters (as per Phase 3 spec):

- **URL**: `nats://localhost:4222`
- **Reconnect Attempts**: 10
- **Reconnect Delay**: 2 seconds
- **Timeout**: 30 seconds
- **Ping Interval**: 20 seconds
- **Max Pings Outstanding**: 3

## Installation

The NATS library is already included in `go.mod`:

```go
github.com/nats-io/nats.go v1.48.0
```

No additional installation required.

## Configuration

### Environment Variables

Configure NATS via environment variables:

```bash
# NATS Connection
NATS_URL=nats://localhost:4222
NATS_ENABLED=true

# Connection Tuning
NATS_RECONNECT_WAIT=2s
NATS_MAX_RECONNECTS=10
NATS_TIMEOUT=30s
NATS_PING_INTERVAL=20s
NATS_MAX_PINGS_OUT=3
```

### Programmatic Configuration

```go
config := &sync.NATSConfig{
    URL:           "nats://localhost:4222",
    Enabled:       true,
    ReconnectWait: 2 * time.Second,
    MaxReconnects: 10,
    Timeout:       30 * time.Second,
    PingInterval:  20 * time.Second,
    MaxPingsOut:   3,
}

client, err := sync.NewNATSClient(config, logger)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Default Configuration

Use default settings:

```go
config := sync.DefaultNATSConfig()
client, err := sync.NewNATSClient(config, logger)
```

## Usage Examples

### 1. Basic Setup

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
config := sync.DefaultNATSConfig()

client, err := sync.NewNATSClient(config, logger)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Check health
if client.HealthCheck() {
    log.Println("NATS is healthy")
}
```

### 2. Publishing Events

#### Workspace Events

```go
// Workspace created
err := client.PublishWorkspaceCreated(
    "ws-12345",      // Workspace ID
    "user-67890",    // User ID
    map[string]interface{}{
        "name": "My Workspace",
        "mode": "2d",
        "template_type": "business_os",
    },
)

// Workspace updated
err := client.PublishWorkspaceUpdated(
    "ws-12345",
    "user-67890",
    updatedData,
    2,  // Sync version
)
```

#### Project Events

```go
err := client.PublishProjectCreated(
    "proj-123",
    "user-456",
    map[string]interface{}{
        "title": "New Project",
        "status": "active",
    },
)
```

#### Task Events

```go
err := client.PublishTaskCreated(
    "task-789",
    "user-456",
    map[string]interface{}{
        "title": "Complete feature",
        "priority": "high",
    },
)
```

#### Conflict Notifications

```go
conflict := &sync.ConflictEvent{
    EntityType:      "workspace",
    EntityID:        "ws-12345",
    ConflictFields:  []string{"name", "layout"},
    LocalData:       []byte(`{"name":"Local Name"}`),
    RemoteData:      []byte(`{"name":"Remote Name"}`),
    LocalUpdatedAt:  time.Now().Add(-1 * time.Minute),
    RemoteUpdatedAt: time.Now(),
    DetectedAt:      time.Now(),
}

err := client.PublishConflict(conflict)
```

### 3. Subscribing to Events

#### Workspace Events

```go
err := client.SubscribeWorkspaceEvents(func(event *sync.SyncEvent) error {
    log.Printf("Received: %s on %s\n", event.EventType, event.EntityID)

    switch event.EventType {
    case "workspace.created":
        // Handle creation
    case "workspace.updated":
        // Handle update
    case "workspace.deleted":
        // Handle deletion
    }

    return nil
})
```

#### Project Events

```go
err := client.SubscribeProjectEvents(func(event *sync.SyncEvent) error {
    log.Printf("Project event: %s\n", event.EventType)
    // Process project event
    return nil
})
```

#### Conflict Events

```go
err := client.SubscribeConflictEvents(func(conflict *sync.ConflictEvent) error {
    log.Printf("Conflict detected: %s (%v fields)\n",
        conflict.EntityID, conflict.ConflictFields)

    // Handle conflict - queue for review or auto-resolve
    return nil
})
```

#### Custom Subscriptions

```go
err := client.Subscribe("custom.subject.*", func(msg *nats.Msg) {
    log.Printf("Received on %s: %s\n", msg.Subject, string(msg.Data))
})
```

### 4. Monitoring and Health

```go
// Check connection status
if client.IsConnected() {
    log.Println("Connected to NATS")
}

// Health check
if !client.HealthCheck() {
    log.Println("NATS health check failed")
}

// Get statistics
stats := client.Stats()
log.Printf("Stats: %+v\n", stats)
// Output:
// Enabled: true
// Connected: true
// URL: nats://localhost:4222
// InMsgs: 123
// OutMsgs: 456
// Reconnects: 2
// Subscriptions: 3
```

### 5. Disabled Mode (Development/Testing)

```go
config := &sync.NATSConfig{
    URL:     "nats://localhost:4222",
    Enabled: false,  // Disable NATS
}

client, err := sync.NewNATSClient(config, logger)
// All operations will no-op gracefully
```

## Integration with Sync Service

### Publishing Local Changes

When a local entity is created/updated, publish to NATS:

```go
func (s *SyncService) WorkspaceUpdated(
    ctx context.Context,
    workspaceID string,
    data interface{},
    version int64,
) error {
    // Update database
    if err := s.db.UpdateWorkspace(ctx, workspaceID, data); err != nil {
        return err
    }

    // Publish to NATS for OSA to pick up
    return s.natsClient.PublishWorkspaceUpdated(
        workspaceID,
        getUserID(ctx),
        data,
        version,
    )
}
```

### Subscribing to Remote Changes

Subscribe to events from OSA:

```go
func (s *SyncService) Start(ctx context.Context) error {
    // Subscribe to workspace events from OSA
    return s.natsClient.SubscribeWorkspaceEvents(func(event *sync.SyncEvent) error {
        // Filter out our own events (echo prevention)
        if event.Source != "osa" {
            return nil
        }

        // Process remote update
        return s.handleRemoteWorkspaceUpdate(ctx, event)
    })
}

func (s *SyncService) handleRemoteWorkspaceUpdate(
    ctx context.Context,
    event *sync.SyncEvent,
) error {
    // 1. Fetch local version
    local, err := s.db.GetWorkspace(ctx, event.EntityID)
    if err != nil {
        return err
    }

    // 2. Detect conflicts
    conflict := s.conflictDetector.Detect(local, event.Data)
    if conflict != nil {
        // Publish conflict notification
        return s.natsClient.PublishConflict(conflict)
    }

    // 3. Apply update
    return s.db.UpdateWorkspace(ctx, event.EntityID, event.Data)
}
```

## Error Handling

### Automatic Reconnection

The client automatically handles disconnections:

```go
// NATS client will automatically reconnect
// Just keep publishing/subscribing normally
err := client.PublishWorkspaceCreated("ws-1", "user-1", data)
if err != nil {
    // Log error - NATS will reconnect and retry
    log.Printf("Publish failed: %v (will auto-reconnect)\n", err)
}
```

### Panic Recovery

Message handlers are wrapped with panic recovery:

```go
err := client.SubscribeWorkspaceEvents(func(event *sync.SyncEvent) error {
    // If this panics, it won't crash the application
    // Panic will be logged and recovered
    processEvent(event)
    return nil
})
```

### Retry Logic

Return errors from handlers to trigger NATS retry:

```go
err := client.SubscribeWorkspaceEvents(func(event *sync.SyncEvent) error {
    if err := processEvent(event); err != nil {
        // Returning error will cause NATS to retry
        return err
    }
    return nil
})
```

## Testing

### Unit Tests

Run unit tests (no NATS server required):

```bash
go test -v -run "TestNewNATSClient_Disabled|TestNATSClient_PublishSync_Disabled" ./internal/sync
```

### Integration Tests

Run integration tests (requires NATS server):

```bash
# Start NATS server
docker run -p 4222:4222 nats:latest

# Run tests
NATS_URL=nats://localhost:4222 go test -v ./internal/sync
```

### Test Coverage

```bash
go test -coverprofile=coverage.out ./internal/sync
go tool cover -html=coverage.out
```

## Performance

### Benchmarks

Expected performance metrics:

- **Publish Rate**: ~500K msgs/sec (local NATS)
- **Subscribe Rate**: ~400K msgs/sec
- **Latency**: <1ms (local), <10ms (network)
- **Memory**: ~50MB per 10K subscriptions

### Optimization Tips

1. **Batch Publishing**: Group related events
2. **Message Size**: Keep payloads under 1MB
3. **Subject Design**: Use specific subjects over wildcards
4. **Connection Pooling**: Share client across goroutines

## Troubleshooting

### Common Issues

#### 1. Connection Refused

```
Error: failed to connect to NATS: connection refused
```

**Solution**: Ensure NATS server is running:

```bash
docker run -p 4222:4222 nats:latest
```

#### 2. Slow Message Processing

```
Warning: message handler taking >1s
```

**Solution**: Offload heavy processing to background workers:

```go
err := client.SubscribeWorkspaceEvents(func(event *sync.SyncEvent) error {
    // Queue for async processing
    go processEventAsync(event)
    return nil  // Ack immediately
})
```

#### 3. Memory Leak

**Solution**: Always close client and unsubscribe:

```go
defer client.Close()

// Unsubscribe when done
defer client.Unsubscribe("osa.sync.workspace.*")
```

## Production Deployment

### Docker Compose

```yaml
services:
  nats:
    image: nats:2.10-alpine
    ports:
      - "4222:4222"
      - "8222:8222"  # Monitoring
    volumes:
      - nats-data:/data
    command:
      - "--jetstream"
      - "--store_dir=/data"

  businessos:
    image: businessos:latest
    environment:
      - NATS_URL=nats://nats:4222
      - NATS_ENABLED=true
    depends_on:
      - nats
```

### Kubernetes

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nats
spec:
  selector:
    app: nats
  ports:
    - port: 4222
      name: client
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: businessos
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: NATS_URL
          value: "nats://nats:4222"
```

### Monitoring

Use NATS monitoring endpoint:

```bash
curl http://localhost:8222/varz
```

Integrate with Prometheus:

```yaml
- job_name: 'nats'
  static_configs:
    - targets: ['nats:8222']
```

## Security

### TLS Configuration

```go
config := &sync.NATSConfig{
    URL: "nats://secure-server:4222",
    // Add TLS options here (future enhancement)
}
```

### Authentication

```go
// Token auth
config.URL = "nats://token@secure-server:4222"

// Username/Password
config.URL = "nats://user:pass@secure-server:4222"
```

## References

- [Phase 3 Sync Design](../../docs/architecture/OSA_PHASE3_SYNC_DESIGN.md)
- [Sync Specification Answers](../../docs/architecture/SYNC_SPECIFICATION_ANSWERS.md) (Q1-Q3)
- [NATS Documentation](https://docs.nats.io/)
- [NATS Go Client](https://github.com/nats-io/nats.go)

## Support

For issues or questions:

1. Check existing sync infrastructure: `/internal/sync/messaging/nats_client.go`
2. Review Phase 3 documentation
3. Examine test examples: `nats_client_test.go`
4. Check example usage: `nats_client_example.go`

## License

Part of BusinessOS - OSA Integration Phase 3
