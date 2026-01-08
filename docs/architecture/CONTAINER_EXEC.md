# Container Exec Attachment for PTY Terminal

This package provides Docker container exec attachment methods for connecting WebSocket terminals to isolated Docker containers.

## Overview

The exec attachment system allows WebSocket clients to connect to interactive shell sessions inside Docker containers with full PTY support including:

- Interactive shell (bash/zsh) execution
- Real-time bidirectional I/O streaming
- Terminal resize support
- Process lifecycle management
- Security isolation per container

## API Methods

### 1. CreateExec(containerID string, cmd []string, tty bool) (string, error)

Creates a new exec instance in a container.

**Parameters:**
- `containerID`: Target container ID
- `cmd`: Command to execute (e.g., `["/bin/bash"]`)
- `tty`: Enable TTY for interactive sessions

**Returns:**
- `execID`: Unique identifier for the exec session
- `error`: Error if creation fails

**Features:**
- Sets WorkingDir to `/workspace`
- Configures terminal environment (TERM, COLORTERM, LANG)
- Attaches stdin, stdout, stderr
- 10-second timeout

**Example:**
```go
execID, err := manager.CreateExec(containerID, []string{"/bin/bash"}, true)
```

### 2. AttachExec(execID string) (types.HijackedResponse, error)

Attaches to an existing exec instance for I/O streaming.

**Parameters:**
- `execID`: Exec instance ID from CreateExec

**Returns:**
- `HijackedResponse`: Connection with Reader/Writer for bidirectional I/O
- `error`: Error if attachment fails

**Features:**
- Returns hijacked connection for raw I/O
- `Response.Reader`: Read container output
- `Response.Conn`: Write user input
- 5-second timeout

**Example:**
```go
hijacked, err := manager.AttachExec(execID)
defer hijacked.Close()

// Stream data
go io.Copy(hijacked.Conn, wsReader)
go io.Copy(wsWriter, hijacked.Reader)
```

### 3. ResizeExec(execID string, height, width uint) error

Resizes the TTY of an exec instance.

**Parameters:**
- `execID`: Exec instance ID
- `height`: Terminal height in rows
- `width`: Terminal width in columns

**Returns:**
- `error`: Error if resize fails

**Features:**
- Updates PTY dimensions dynamically
- Should be called when browser window resizes
- 3-second timeout

**Example:**
```go
// Handle resize event from WebSocket
err := manager.ResizeExec(execID, 24, 80)
```

### 4. GetExecInfo(execID string) (container.ExecInspect, error)

Retrieves information about an exec instance.

**Parameters:**
- `execID`: Exec instance ID

**Returns:**
- `ExecInspect`: Exec status including Running, ExitCode, etc.
- `error`: Error if inspection fails

**Features:**
- Check if exec is still running
- Get exit code when completed
- 3-second timeout

**Example:**
```go
info, err := manager.GetExecInfo(execID)
if !info.Running {
    log.Printf("Exec exited with code: %d", info.ExitCode)
}
```

### Convenience Methods

#### StartExec(containerID string, cmd []string, tty bool) (string, types.HijackedResponse, error)

One-step method that creates, attaches, and starts an exec instance.

**Example:**
```go
execID, hijacked, err := manager.StartExec(containerID, []string{"/bin/bash"}, true)
defer hijacked.Close()
// Ready to stream I/O
```

#### WaitExec(execID string) (int, error)

Waits for exec to complete and returns exit code (max 5 minutes).

#### StopExec(execID string) error

Checks exec status (note: Docker doesn't provide direct exec stop).

## WebSocket Integration Pattern

### Recommended Flow:

1. **WebSocket Connection**
   ```go
   conn, _ := upgrader.Upgrade(w, r, nil)
   defer conn.Close()
   ```

2. **Get or Create Container**
   ```go
   containerID, _ := manager.CreateContainer(userID, "ubuntu:22.04")
   manager.StartContainer(containerID)
   ```

3. **Create Exec Session**
   ```go
   execID, _ := manager.CreateExec(containerID, []string{"/bin/bash"}, true)
   ```

4. **Attach for I/O**
   ```go
   hijacked, _ := manager.AttachExec(execID)
   defer hijacked.Close()
   ```

5. **Bidirectional Streaming**
   ```go
   // WebSocket → Container
   go func() {
       for {
           _, msg, err := conn.ReadMessage()
           if err != nil { return }
           hijacked.Conn.Write(msg)
       }
   }()
   
   // Container → WebSocket
   go func() {
       buf := make([]byte, 4096)
       for {
           n, err := hijacked.Reader.Read(buf)
           if err != nil { return }
           conn.WriteMessage(websocket.TextMessage, buf[:n])
       }
   }()
   ```

6. **Handle Resize**
   ```go
   // On resize message from client
   manager.ResizeExec(execID, rows, cols)
   ```

7. **Cleanup on Disconnect**
   ```go
   hijacked.Close() // Stops the exec session
   ```

## Security Features

Containers created with `CreateContainer` include:

- **Capability Dropping**: All capabilities dropped, only essential ones added (CHOWN, DAC_OVERRIDE, SETGID, SETUID)
- **Resource Limits**:
  - Memory: 512MB default
  - CPU: 50% of one core
  - PIDs: Max 100 processes
- **Network Isolation**: NetworkMode set to "none"
- **Volume Isolation**: User-specific workspace volumes
- **Working Directory**: Isolated to `/workspace`

Exec sessions inherit the container's security context.

## Error Handling

All methods include:
- Input validation (empty checks)
- Context timeouts
- Structured error wrapping
- Comprehensive logging

## Timeouts

- `execCreateTimeout`: 10 seconds
- `execAttachTimeout`: 5 seconds
- `execResizeTimeout`: 3 seconds
- `execInspectTimeout`: 3 seconds

## Example Implementation

See `exec_example.go` for detailed usage examples.

## Integration with Terminal Package

The terminal WebSocket handler (`internal/terminal/websocket.go`) should:

1. Parse connection parameters (cols, rows, containerID)
2. Create or reuse container via ContainerManager
3. Use `CreateExec` to set up shell session
4. Use `AttachExec` to get I/O connection
5. Stream data bidirectionally
6. Handle resize events with `ResizeExec`
7. Clean up on disconnect

## Testing

```bash
# Build and verify
cd backend-go
go build ./internal/container/...

# Run tests (if available)
go test ./internal/container/...
```

## Dependencies

- `github.com/docker/docker/client` - Docker SDK
- `github.com/docker/docker/api/types` - Docker API types
- `github.com/docker/docker/api/types/container` - Container types

## Notes

- Docker daemon must be running and accessible
- Container must be in "running" state before creating exec
- Each exec session is independent (multiple execs per container supported)
- Exec sessions terminate when connection closes or process exits
- Working directory defaults to `/workspace` (user's isolated volume)
