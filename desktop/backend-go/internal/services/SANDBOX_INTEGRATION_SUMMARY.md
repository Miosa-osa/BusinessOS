# Sandbox Integration Service

## Overview

The `SandboxIntegrationService` bridges OSA app generation with sandbox container deployment. When OSA finishes generating an app, this service automatically deploys it to a secure Docker sandbox.

## Architecture

```
OSA App Generation → OnAppGenerationComplete() → SandboxDeploymentService
                              ↓
                    Database Updates (app status)
```

## Key Components

### 1. SandboxIntegrationService

**Location:** `internal/services/sandbox_integration.go`

**Responsibilities:**
- Listen for app generation completion events
- Determine appropriate Docker image based on app type
- Build deployment configuration
- Deploy sandbox container
- Handle deployment failures gracefully
- Update app status in database

### 2. Main Integration Point

```go
func (s *SandboxIntegrationService) OnAppGenerationComplete(
    ctx context.Context,
    appID uuid.UUID,
    workspacePath string,
    appConfig *AppConfig,
) error
```

**Flow:**
1. Validate inputs (app ID, workspace path)
2. Fetch app details from database
3. Check app status is deployable (`generated` or `built`)
4. Update status to `deploying`
5. Determine Docker image from app type
6. Build deployment request
7. Deploy via `SandboxDeploymentService`
8. Update status to `deployed` with URL
9. Handle errors → update status to `failed`/`error`

### 3. Docker Image Mapping

Automatically selects Docker images based on app type:

| App Type | Docker Image |
|----------|-------------|
| React, Next.js, Vue, Svelte, Angular, Node | `node:20-alpine` |
| Python, Flask, Django, FastAPI | `python:3.11-slim` |
| Go, Golang | `golang:1.22-alpine` |
| HTML, Static | `nginx:alpine` |
| Unknown | `node:20-alpine` (default) |

### 4. Start Command Inference

Infers start commands when not explicitly provided:

| App Type | Start Command |
|----------|---------------|
| React | `npm start` |
| Next.js | `npm run dev` |
| Express | `node server.js` |
| Python | `python app.py` |
| FastAPI | `uvicorn main:app --host 0.0.0.0 --port 8000` |
| Go | `go run main.go` |

### 5. Configuration Handling

**AppConfig Structure:**
```go
type AppConfig struct {
    AppType         string            // e.g., "react", "python"
    Framework       string            // e.g., "nextjs", "flask"
    Port            int               // Container port
    StartCommand    []string          // Command to start app
    BuildCommand    []string          // Build command (if needed)
    Environment     map[string]string // Environment variables
    WorkingDir      string            // Working directory
    Dependencies    []string          // Dependencies list
    RequiresBuild   bool              // Build step required
}
```

Parsed from `osa_generated_apps.metadata` JSONB field.

### 6. App Status Flow

```
generated → deploying → deployed
    ↓           ↓
  built       failed/error
```

**Deployable Statuses:** `generated`, `built`

### 7. Helper Methods

**StopAppSandbox:** Stop running sandbox
- Calls `SandboxDeploymentService.Stop()`
- Updates status to `built`

**RestartAppSandbox:** Restart sandbox
- Calls `SandboxDeploymentService.Restart()`
- Updates status to `deployed` with new URL

**RemoveAppSandbox:** Remove sandbox completely
- Calls `SandboxDeploymentService.Remove()`
- Updates status to `generated`

**GetAppSandboxInfo:** Get sandbox details
- Returns `SandboxInfo` from deployment service

**ListWorkspaceAppSandboxes:** List all sandboxes for workspace apps
- Retrieves user ID from workspace
- Lists all user sandboxes

### 8. Configuration

**Auto-Deploy Toggle:**
```go
service.SetAutoDeployEnabled(true/false)
```

**Workspace Base Path:**
```go
service.SetWorkspaceBasePath("/path/to/osa_workspaces")
```

Default: `/tmp/osa_workspaces`

## Database Integration

### Tables Used

1. **osa_generated_apps:**
   - `id`, `workspace_id`, `name`, `status`, `metadata`
   - `deployment_url`, `container_id`, `sandbox_status`

2. **osa_workspaces:**
   - `id`, `user_id` (for linking apps to users)

3. **sandbox_events:**
   - Logs deployment lifecycle events

### Status Updates

```sql
UPDATE osa_generated_apps
SET status = 'deploying',
    updated_at = NOW()
WHERE id = $1;
```

```sql
UPDATE osa_generated_apps
SET status = 'deployed',
    deployment_url = $2,
    deployed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;
```

## Error Handling

### Error Types

- `ErrAppNotFound`: App doesn't exist in database
- `ErrWorkspacePathEmpty`: No workspace path provided
- `ErrInvalidAppStatus`: App status not deployable
- `ErrDeploymentDisabled`: Auto-deploy is disabled
- `ErrNoDockerImage`: Cannot determine Docker image

### Failure Modes

1. **Invalid Input:** Returns error immediately
2. **App Not Found:** Returns `ErrAppNotFound`
3. **Wrong Status:** Returns `ErrInvalidAppStatus`
4. **Deployment Fails:** Updates status to `failed`, logs error
5. **Database Error:** Logs warning, continues if possible

## Testing

**Test File:** `internal/services/sandbox_integration_test.go`

**Test Coverage:**
- Docker image determination (10 cases)
- Start command inference (7 cases)
- App config parsing (3 cases)
- Deployable status checks (7 cases)
- Port defaults handling
- Environment variable handling
- Input validation
- Configuration setters
- JSON serialization roundtrip

## Usage Example

```go
// Create service
integrationService := NewSandboxIntegrationService(
    sandboxService,
    dbPool,
    config,
    logger,
)

// Handle app generation completion
appConfig := &AppConfig{
    AppType:      "react",
    Port:         3000,
    StartCommand: []string{"npm", "start"},
    Environment:  map[string]string{"NODE_ENV": "development"},
}

err := integrationService.OnAppGenerationComplete(
    ctx,
    appID,
    "/path/to/workspace",
    appConfig,
)
if err != nil {
    log.Error("deployment failed", "error", err)
}

// Stop sandbox
err = integrationService.StopAppSandbox(ctx, appID)

// Restart sandbox
info, err := integrationService.RestartAppSandbox(ctx, appID)

// Remove sandbox
err = integrationService.RemoveAppSandbox(ctx, appID)
```

## Integration Points

### Where to Call OnAppGenerationComplete

**Option 1: OSA Webhook Handler**
```go
func HandleOSAWebhook(ctx context.Context, event OSAEvent) error {
    if event.Type == "app_generation_complete" {
        return integrationService.OnAppGenerationComplete(
            ctx,
            event.AppID,
            event.WorkspacePath,
            event.AppConfig,
        )
    }
}
```

**Option 2: Build Event Listener**
```go
func WatchBuildEvents(ctx context.Context) {
    for event := range buildEventsChan {
        if event.Phase == "complete" && event.Status == "success" {
            integrationService.OnAppGenerationComplete(...)
        }
    }
}
```

**Option 3: Polling (Fallback)**
```go
func PollAppGenerationStatus(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        apps := getRecentlyGeneratedApps(ctx)
        for _, app := range apps {
            integrationService.OnAppGenerationComplete(...)
        }
    }
}
```

## Security Considerations

1. **Path Validation:** Workspace paths are made absolute to prevent traversal
2. **User Quotas:** Enforced by `SandboxDeploymentService`
3. **Resource Limits:** Memory (512MB) and CPU (50%) limits applied
4. **Port Allocation:** Handled securely by `SandboxPortAllocator`
5. **Container Isolation:** Docker security via AppContainerManager

## Performance

- **Async Deployment:** Can be backgrounded if needed
- **Database Queries:** Minimal (2-3 per deployment)
- **Docker Operations:** ~5-10 seconds per deployment
- **Concurrent Deploys:** Supported (handled by SandboxDeploymentService)

## Future Enhancements

1. **Build Support:** Detect and run build commands before deployment
2. **Health Checks:** Custom health check endpoints per app type
3. **Auto-Scaling:** Resource adjustment based on app metrics
4. **Blue-Green Deploy:** Zero-downtime updates
5. **Environment Templates:** Pre-configured environments per framework
6. **Dependency Installation:** Auto-install packages (npm install, pip install)
7. **Port Auto-Detection:** Scan app code to determine port
8. **Multi-Container:** Support apps with multiple services (DB, cache)

## Monitoring

**Metrics to Track:**
- Deployment success rate
- Average deployment time
- Active sandboxes per user
- Failed deployments by app type
- Resource utilization per sandbox

**Logging:**
- All operations logged with `slog`
- Component: "sandbox_integration"
- Key fields: app_id, workspace_path, app_type, status

## Dependencies

- `SandboxDeploymentService`: Core deployment logic
- `pgxpool`: Database connection
- `sqlc.Queries`: Type-safe database queries
- `config.Config`: Application configuration
- `slog.Logger`: Structured logging

## Status

✅ **Implemented:** Core integration service
✅ **Tested:** Unit tests for key functions
⏳ **Pending:** Integration with OSA event system
⏳ **Pending:** REST API handlers
⏳ **Pending:** WebSocket status updates

---

**Created:** 2025-02-03
**Version:** 1.0
**Status:** Ready for integration
