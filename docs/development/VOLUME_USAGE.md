# Docker Volume Management

This package provides comprehensive Docker volume management for user workspaces in BusinessOS.

## Features

- **Volume Creation**: Create persistent Docker volumes for user workspaces
- **Volume Removal**: Safely remove volumes with usage checks
- **Volume Inspection**: Check existence and get detailed information
- **Volume Listing**: List volumes by user or all workspace volumes
- **Volume Pruning**: Clean up unused volumes
- **Statistics**: Get usage statistics across all volumes

## Usage Examples

### Basic Setup

```go
import (
    "context"
    "log"
    "github.com/rhl/businessos-backend/internal/container"
)

// Create a container manager
ctx := context.Background()
manager, err := container.NewContainerManager(ctx, "ubuntu:22.04")
if err != nil {
    log.Fatalf("Failed to create container manager: %v", err)
}
defer manager.Close()
```

### Creating a Volume

```go
userID := "user-123"

// Create a volume for the user
volumeName, err := manager.CreateVolume(userID)
if err != nil {
    log.Fatalf("Failed to create volume: %v", err)
}

log.Printf("Created volume: %s", volumeName)
// Output: Created volume: workspace_user-123
```

### Checking if a Volume Exists

```go
exists, err := manager.VolumeExists(userID)
if err != nil {
    log.Fatalf("Failed to check volume existence: %v", err)
}

if exists {
    log.Printf("Volume exists for user %s", userID)
}
```

### Getting Volume Information

```go
volumeName := "workspace_user-123"

volumeInfo, err := manager.GetVolumeInfo(volumeName)
if err != nil {
    log.Fatalf("Failed to get volume info: %v", err)
}

log.Printf("Volume: %s", volumeInfo.Name)
log.Printf("Driver: %s", volumeInfo.Driver)
log.Printf("Mountpoint: %s", volumeInfo.Mountpoint)
log.Printf("Labels: %v", volumeInfo.Labels)

// Check if volume is in use
if volumeInfo.UsageData != nil {
    log.Printf("RefCount: %d", volumeInfo.UsageData.RefCount)
    log.Printf("Size: %d bytes", volumeInfo.UsageData.Size)
}
```

### Listing User Volumes

```go
// List all volumes for a specific user
volumes, err := manager.ListUserVolumes(userID)
if err != nil {
    log.Fatalf("Failed to list user volumes: %v", err)
}

for _, vol := range volumes {
    log.Printf("Volume: %s (Driver: %s)", vol.Name, vol.Driver)
}
```

### Listing All Workspace Volumes

```go
// List all workspace volumes in the system
allVolumes, err := manager.ListAllWorkspaceVolumes()
if err != nil {
    log.Fatalf("Failed to list workspace volumes: %v", err)
}

log.Printf("Total workspace volumes: %d", len(allVolumes))
```

### Removing a Volume

```go
// Remove a user's volume (only if not in use)
err := manager.RemoveVolume(userID)
if err != nil {
    log.Printf("Failed to remove volume: %v", err)
    // Error will indicate if volume is in use
}
```

### Getting Volume Statistics

```go
stats, err := manager.GetVolumeStats()
if err != nil {
    log.Fatalf("Failed to get volume stats: %v", err)
}

log.Printf("Total volumes: %d", stats.TotalVolumes)
log.Printf("Volumes in use: %d", stats.VolumesInUse)
log.Printf("Unused volumes: %d", stats.UnusedVolumes)
```

### Pruning Unused Volumes

```go
// Remove all unused workspace volumes
removedVolumes, err := manager.PruneUnusedVolumes()
if err != nil {
    log.Fatalf("Failed to prune volumes: %v", err)
}

log.Printf("Removed %d unused volumes", len(removedVolumes))
for _, volumeName := range removedVolumes {
    log.Printf("  - %s", volumeName)
}
```

## Volume Naming Convention

Volumes are automatically named using the pattern: `workspace_{userID}`

Example:
- User ID: `user-123` → Volume: `workspace_user-123`
- User ID: `550e8400-e29b-41d4-a716-446655440000` → Volume: `workspace_550e8400-e29b-41d4-a716-446655440000`

## Volume Labels

All workspace volumes are created with the following labels:

- `app=businessos` - Identifies volumes belonging to BusinessOS
- `type=workspace` - Identifies the volume type
- `user_id={userID}` - Links the volume to a specific user

These labels enable:
- Easy filtering and querying
- Automated cleanup and maintenance
- User-specific volume management

## Integration with Containers

To mount a user's workspace volume in a container:

```go
import (
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/mount"
)

// Create the volume first
volumeName, err := manager.CreateVolume(userID)
if err != nil {
    log.Fatalf("Failed to create volume: %v", err)
}

// Configure container with the volume mounted
hostConfig := &container.HostConfig{
    Mounts: []mount.Mount{
        {
            Type:   mount.TypeVolume,
            Source: volumeName,
            Target: "/workspace",
        },
    },
}

// Create container with the volume
containerID, err := manager.CreateContainer(userID, "ubuntu:22.04", nil, hostConfig)
if err != nil {
    log.Fatalf("Failed to create container: %v", err)
}
```

## Error Handling

All volume methods return detailed errors with context:

```go
_, err := manager.CreateVolume("")
// Error: "userID cannot be empty"

err = manager.RemoveVolume("user-123")
// Error: "volume workspace_user-123 is currently in use by 2 container(s)"

_, err = manager.GetVolumeInfo("nonexistent")
// Error: "volume nonexistent not found"
```

## Timeouts

The following timeouts are configured for volume operations:

- Volume creation: 30 seconds
- Volume removal: 30 seconds
- Volume listing: 15 seconds
- Volume inspection: 10 seconds

## Safety Features

### Volume Removal Protection

`RemoveVolume()` includes safety checks:

1. Verifies the volume exists before attempting removal
2. Checks if the volume is currently in use by any containers
3. Returns an error if the volume is in use (prevents data loss)
4. Uses non-force removal for additional safety

### Concurrent Access

All volume operations use the Docker client with proper timeout contexts, ensuring:
- Operations don't hang indefinitely
- Multiple operations can be performed safely
- The Docker daemon connection is properly managed

## Best Practices

1. **Always check errors**: Volume operations can fail for various reasons
2. **Create volumes before containers**: Ensure volumes exist before mounting
3. **Clean up unused volumes**: Use `PruneUnusedVolumes()` periodically
4. **Monitor volume usage**: Use `GetVolumeStats()` for capacity planning
5. **Handle volume removal failures**: Containers must be stopped before volume removal

## Testing

Run the test suite:

```bash
# Run all tests
go test -v ./internal/container/...

# Run only unit tests (no Docker required)
go test -v ./internal/container/... -short

# Run integration tests (requires Docker)
go test -v ./internal/container/...
```

## Logging

All volume operations are logged with the `[Container]` prefix:

```bash
[Container] Volume created: workspace_user-123
[Container] Listed 3 volumes for user user-123
[Container] Volume removed: workspace_user-456
[Container] Pruned 5 unused volumes
```
