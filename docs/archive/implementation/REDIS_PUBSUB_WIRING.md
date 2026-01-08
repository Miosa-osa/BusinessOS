# Redis Pub/Sub Wiring for Terminal Horizontal Scaling

## Overview
Terminal pub/sub is now wired up for horizontal scaling with graceful degradation when Redis is unavailable.

## Architecture Flow

```bash
main.go
  ├─> Redis connection (optional)
  ├─> Create TerminalPubSub (if Redis available)
  ├─> Pass to handlers.NewHandlers()
  │
handlers.NewHandlers()
  ├─> Store terminalPubSub field
  ├─> Pass to NewTerminalHandler()
  │
NewTerminalHandler()
  ├─> Create Manager
  ├─> Call manager.SetPubSub() if available
  └─> Log scaling mode (single/multi-instance)
```
## Files Modified

### 1. `/cmd/server/main.go`
- Added `terminal` package import
- Created `terminalPubSub` variable after Redis connection
- Pass `terminalPubSub` to `handlers.NewHandlers()`

```go
// Create terminal pub/sub for horizontal scaling
terminalPubSub = terminal.NewTerminalPubSub(redisClient.Client(), instanceID)
log.Printf("Terminal pub/sub enabled (instance=%s)", instanceID)
```

### 2. `/internal/handlers/handlers.go`
- Added `terminalPubSub` field to `Handlers` struct
- Added `terminal` import
- Updated `NewHandlers()` to accept `terminalPubSub` parameter
- Pass pub/sub to `NewTerminalHandler()`

```go
type Handlers struct {
    pool            *pgxpool.Pool
    cfg             *config.Config
    containerMgr    *container.ContainerManager
    sessionCache    *middleware.SessionCache
    terminalPubSub  *terminal.TerminalPubSub  // Redis pub/sub for terminal scaling
}
```

### 3. `/internal/handlers/terminal.go`
- Added `pubsub` field to `TerminalHandler` struct
- Updated `NewTerminalHandler()` to accept optional `pubsub` parameter
- Call `manager.SetPubSub()` if pub/sub is available
- Added pub/sub cleanup in `Shutdown()` method

```go
func NewTerminalHandler(containerMgr *container.ContainerManager, pubsub *terminal.TerminalPubSub) *TerminalHandler {
    manager := terminal.NewManager(containerMgr)

    // Wire up pub/sub if available
    if pubsub != nil {
        manager.SetPubSub(pubsub)
        log.Printf("[Terminal] Pub/sub enabled for horizontal scaling (instance=%s)", pubsub.InstanceID())
    } else {
        log.Printf("[Terminal] Pub/sub disabled - single instance mode")
    }
    // ...
}
```

## Graceful Degradation

The system works in both single-instance and multi-instance modes:

### Single Instance (No Redis)
- `terminalPubSub` is `nil`
- Terminal manager operates without pub/sub
- Sessions are local to the server instance
- Logs: "Pub/sub disabled - single instance mode"

### Multi Instance (Redis Available)
- `terminalPubSub` is initialized with Redis client and instance ID
- Terminal manager receives pub/sub instance
- Sessions broadcast events across instances
- Logs: "Pub/sub enabled for horizontal scaling (instance=abc123)"

## Startup Logs

### With Redis:
```bash
Server instance ID: abc12345
Redis connected successfully
Session cache enabled (TTL=15m)
Terminal pub/sub enabled (instance=abc12345)
[Terminal] Pub/sub enabled for horizontal scaling (instance=abc12345)
```
### Without Redis:
```bash
Server instance ID: abc12345
Warning: Redis unavailable: connection refused
Sessions will use direct DB auth (not optimal for horizontal scaling)
[Terminal] Pub/sub disabled - single instance mode
```
## Pub/Sub Channels

When enabled, the following Redis channels are used:

1. **terminal:output** - Terminal output broadcasting (for session sharing/observation)
2. **terminal:resize** - Terminal resize events across instances
3. **terminal:sessions** - Session lifecycle events (created, closed)

## Testing

### Verify Redis Connection
```bash
curl http://localhost:3000/health/detailed
```

Check response for:
```json
{
  "components": {
    "redis": {
      "status": "connected",
      "latency_ms": 2,
      "pool_stats": {...}
    }
  }
}
```

### Verify Instance ID
```bash
curl http://localhost:3000/
```

Response includes:
```json
{
  "instance_id": "abc12345"
}
```

## Shutdown Sequence

When server shuts down:
1. Container monitor stops (if available)
2. Terminal handler closes pub/sub connections
3. Terminal manager shuts down sessions
4. Container manager cleans up
5. Redis connection closes
6. Database connection closes

## Related Files

- `/internal/terminal/pubsub.go` - Pub/sub implementation
- `/internal/terminal/manager.go` - Manager with `SetPubSub()` method
- `/internal/redis/client.go` - Redis connection management
- `/docker-compose.yml` - Redis service definition
