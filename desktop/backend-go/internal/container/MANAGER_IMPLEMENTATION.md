# Container Manager Implementation

## Overview

The container manager at `/Users/ososerious/BusinessOS-1/desktop/backend-go/internal/container/manager.go` provides a robust interface for managing Docker containers through the Docker SDK for Go (v27.5.0+incompatible).

## Implementation Details

### Core Structure

```go
type ContainerManager struct {
    cli          *client.Client     // Docker SDK client
    ctx          context.Context    // Base context for operations
    defaultImage string             // Default container image
    mu           sync.RWMutex       // Protects containers map
    containers   map[string]*ContainerInfo
}

type ContainerInfo struct {
    ID           string
    UserID       string
    Image        string
    Status       string
    CreatedAt    time.Time
    LastActivity time.Time
}
```

### Key Features

#### 1. Initialization (`NewContainerManager`)
- Creates Docker client using `client.NewClientWithOpts` with:
  - `client.FromEnv` - reads DOCKER_HOST, DOCKER_TLS_VERIFY, etc.
  - `client.WithAPIVersionNegotiation()` - auto-negotiates API version
- Verifies Docker daemon availability with 5-second timeout
- Sets default image to `ubuntu:22.04` if not specified
- Returns error if Docker daemon is not available

#### 2. Connection Management
- **IsDockerAvailable()**: Quick health check with 3-second timeout
- **Close()**: Cleanly closes Docker client connection
- **Shutdown()**: Gracefully stops all tracked containers and closes client

#### 3. Container Tracking
- **ListContainers(userID)**: Lists containers, optionally filtered by user
- **GetContainerInfo(containerID)**: Retrieves detailed container information
- **UpdateActivity(containerID)**: Updates last activity timestamp

### Integration with Existing Code

The manager integrates with other files in the package:

- **container.go**: Provides `CreateContainer`, `StartContainer`, `StopContainer`, `RemoveContainer`
- **exec.go**: Provides `CreateExec`, `AttachExec`, `ResizeExec` for interactive sessions
- **volume.go**: Provides `CreateVolume`, `RemoveVolume`, `VolumeExists` for persistent storage

### Error Handling

All operations include:
- Context timeouts (ranging from 3s to 30s depending on operation)
- Proper error wrapping with `fmt.Errorf` and `%w`
- Detailed logging using `log.Printf` with `[Container]` prefix
- Graceful degradation (e.g., Shutdown continues even if StopContainer fails)

### Thread Safety

- Uses `sync.RWMutex` for concurrent access to containers map
- Read operations use `RLock()` / `RUnlock()`
- Write operations use `Lock()` / `Unlock()`
- Safe for concurrent use across multiple goroutines

## Usage Example

```go
package main

import (
    "context"
    "log"

    "github.com/rhl/businessos-backend/internal/container"
)

func main() {
    ctx := context.Background()

    // Create container manager
    manager, err := container.NewContainerManager(ctx, "ubuntu:22.04")
    if err != nil {
        log.Fatalf("Failed to create manager: %v", err)
    }
    defer manager.Shutdown()

    // Check Docker availability
    if !manager.IsDockerAvailable() {
        log.Fatal("Docker is not available")
    }

    // Create and start a container
    containerID, err := manager.CreateContainer("user123", "ubuntu:22.04")
    if err != nil {
        log.Fatalf("Failed to create container: %v", err)
    }

    if err := manager.StartContainer(containerID); err != nil {
        log.Fatalf("Failed to start container: %v", err)
    }

    // Get container info
    info, err := manager.GetContainerInfo(containerID)
    if err != nil {
        log.Fatalf("Failed to get info: %v", err)
    }
    log.Printf("Container state: %s", info.State.Status)

    // Clean up
    manager.StopContainer(containerID, 10)
    manager.RemoveContainer(containerID, false)
}
```

## Testing

The implementation includes comprehensive tests in `manager_test.go`:

- **TestNewContainerManager**: Tests initialization with default image
- **TestNewContainerManagerWithCustomImage**: Tests custom image specification
- **TestContainerManagerClose**: Tests proper cleanup
- **TestListContainers**: Tests container tracking and filtering
- **TestUpdateActivity**: Tests activity timestamp updates

Run tests:
```bash
cd /Users/ososerious/BusinessOS-1/desktop/backend-go
go test ./internal/container/ -v
```

## Dependencies

- **Docker SDK**: `github.com/docker/docker@v27.5.0+incompatible`
- **Docker API Types**: `github.com/docker/docker/api/types`
- **Docker Client**: `github.com/docker/docker/client`

## Configuration

The manager uses environment variables from Docker:
- `DOCKER_HOST`: Docker daemon socket/TCP address
- `DOCKER_API_VERSION`: API version (auto-negotiated if not set)
- `DOCKER_CERT_PATH`: Path to TLS certificates
- `DOCKER_TLS_VERIFY`: Enable TLS verification

## Logging

All operations log to standard output with the `[Container]` prefix:
```text
2025/12/22 09:26:39 [Container] Creating new container manager with image: ubuntu:22.04
2025/12/22 09:26:39 [Container] Docker daemon connection verified
2025/12/22 09:26:39 [Container] Container manager initialized successfully
```
## Best Practices

1. **Always defer Shutdown()** to ensure proper cleanup
2. **Check IsDockerAvailable()** before performing operations
3. **Use context timeouts** for all Docker operations
4. **Handle errors appropriately** - Docker operations can fail for many reasons
5. **Monitor LastActivity** for idle container cleanup
6. **Use proper container lifecycle**: Create → Start → Use → Stop → Remove

## Security Considerations

The manager itself does not enforce security policies. Security features are implemented in:
- **container.go**: Capability dropping, resource limits, network isolation
- **volume.go**: User-specific volume isolation with labels
- **exec.go**: Working directory restrictions, environment sanitization

## Performance Characteristics

- **Initialization**: ~10-50ms (depends on Docker daemon responsiveness)
- **IsDockerAvailable**: 3s timeout
- **GetContainerInfo**: 10s timeout
- **Container operations**: Delegated to other files (container.go)
- **Memory overhead**: Minimal (only tracks metadata in map)

## Future Enhancements

Potential improvements:
1. Container event monitoring for automatic status updates
2. Metrics collection (CPU, memory, network)
3. Container image pre-pulling for faster startup
4. Connection pooling for high-concurrency scenarios
5. Automatic cleanup of idle containers
6. Health check integration
