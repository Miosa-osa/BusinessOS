# Sandbox Integration with OSA App Generation - Handoff Document

## Overview

This document describes the newly implemented `SandboxIntegrationService` that automatically deploys OSA-generated apps to secure Docker sandboxes.

## Files Created

### 1. Core Service
**Location:** `internal/services/sandbox_integration.go`
- **Lines:** 450+
- **Purpose:** Integration layer between OSA app generation and sandbox deployment

### 2. Test Suite
**Location:** `internal/services/sandbox_integration_test.go`
- **Lines:** 450+
- **Coverage:** 12 test suites, 40+ test cases
- **Tests:** Docker image selection, command inference, config parsing, validation

### 3. Documentation
**Location:** `internal/services/SANDBOX_INTEGRATION_SUMMARY.md`
- **Purpose:** Comprehensive service documentation
- **Sections:** Architecture, usage, database integration, error handling

### 4. SQL Query Update
**Location:** `internal/database/queries/osa.sql`
- **Change:** Updated `UpdateOSAGeneratedAppStatus` to accept `deployment_url` parameter
- **Note:** Requires running `sqlc generate` to regenerate Go code

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    OSA App Generation Workflow                  │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ OnAppGenerationComplete()
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│              SandboxIntegrationService                          │
│                                                                 │
│  • Validate app status (generated/built)                       │
│  • Determine Docker image from app type                        │
│  • Build deployment configuration                              │
│  • Update status: deploying                                    │
│                                                                 │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ Deploy()
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│              SandboxDeploymentService                           │
│                                                                 │
│  • Allocate port                                               │
│  • Create Docker container                                     │
│  • Start container                                             │
│  • Monitor health                                              │
│                                                                 │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            │ Success/Failure
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│              Database Update                                    │
│                                                                 │
│  • Status: deployed (with URL)                                 │
│  • Status: failed (with error message)                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Key Features

### 1. Automatic Docker Image Selection
Maps app types to appropriate Docker images:
- **JavaScript/TypeScript:** `node:20-alpine`
- **Python:** `python:3.11-slim`
- **Go:** `golang:1.22-alpine`
- **Static:** `nginx:alpine`

### 2. Smart Start Command Inference
Determines appropriate start commands when not provided:
- React → `npm start`
- Next.js → `npm run dev`
- FastAPI → `uvicorn main:app --host 0.0.0.0 --port 8000`
- Go → `go run main.go`

### 3. Graceful Error Handling
- Updates app status to `failed` on deployment errors
- Stores error messages in database
- Logs all operations with structured logging
- Cleans up resources on failure

### 4. Resource Management
- Default memory limit: 512MB
- Default CPU quota: 50%
- Automatic port allocation
- Workspace path validation

## Integration Points

### Where to Call the Service

You need to call `OnAppGenerationComplete()` when an OSA app generation finishes. Here are three integration approaches:

#### Option 1: OSA Webhook Handler (Recommended)
```go
func HandleOSAWebhook(w http.ResponseWriter, r *http.Request) {
    var event OSAWebhookEvent
    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    if event.Type == "app.generation.complete" {
        appConfig, err := ParseAppConfig(event.Metadata)
        if err != nil {
            logger.Error("failed to parse app config", "error", err)
            return
        }

        err = integrationService.OnAppGenerationComplete(
            r.Context(),
            event.AppID,
            event.WorkspacePath,
            appConfig,
        )
        if err != nil {
            logger.Error("deployment failed", "error", err)
            // Handle error - may want to retry
        }
    }

    w.WriteHeader(http.StatusOK)
}
```

#### Option 2: Build Event Listener
```go
func WatchOSABuildEvents(ctx context.Context, integrationService *SandboxIntegrationService) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Query recent build events with status "complete"
            events, err := queries.GetRecentBuildEvents(ctx, sqlc.GetRecentBuildEventsParams{
                EventType: "build_complete",
                Limit:     10,
            })
            if err != nil {
                logger.Error("failed to fetch build events", "error", err)
                continue
            }

            for _, event := range events {
                // Get app details
                app, err := queries.GetOSAGeneratedApp(ctx, event.AppID)
                if err != nil {
                    continue
                }

                // Parse metadata
                appConfig, err := ParseAppConfig(app.Metadata)
                if err != nil {
                    continue
                }

                // Determine workspace path from event data
                workspacePath := extractWorkspacePath(event.EventData)

                // Trigger deployment
                err = integrationService.OnAppGenerationComplete(
                    ctx,
                    uuid.UUID(app.ID.Bytes),
                    workspacePath,
                    appConfig,
                )
                if err != nil {
                    logger.Error("deployment failed",
                        "app_id", app.ID,
                        "error", err)
                }
            }
        }
    }
}
```

#### Option 3: OSA Workflow Status Check
```go
func CheckWorkflowCompletion(ctx context.Context, workflowID string) error {
    // Get workflow status from OSA
    workflow, err := osaClient.GetWorkflow(ctx, workflowID)
    if err != nil {
        return err
    }

    if workflow.Status == "completed" && workflow.AppID != "" {
        // Get associated app
        app, err := queries.GetOSAGeneratedAppByWorkflowID(ctx, workflowID)
        if err != nil {
            return err
        }

        // Parse metadata
        appConfig, err := ParseAppConfig(app.Metadata)
        if err != nil {
            return err
        }

        // Deploy
        return integrationService.OnAppGenerationComplete(
            ctx,
            uuid.UUID(app.ID.Bytes),
            workflow.WorkspacePath,
            appConfig,
        )
    }

    return nil
}
```

## Service Initialization

Add to your server startup code:

```go
// In cmd/server/main.go or similar

// Create sandbox deployment service (already exists)
sandboxService, err := services.NewSandboxDeploymentService(
    pool,
    dockerClient,
    cfg,
    logger,
)
if err != nil {
    logger.Error("failed to create sandbox service", "error", err)
    os.Exit(1)
}

// Create sandbox integration service (NEW)
sandboxIntegrationService := services.NewSandboxIntegrationService(
    sandboxService,
    pool,
    cfg,
    logger,
)

// Configure workspace base path (adjust to your setup)
sandboxIntegrationService.SetWorkspaceBasePath("/path/to/osa_workspaces")

// Optional: Enable/disable auto-deploy
sandboxIntegrationService.SetAutoDeployEnabled(cfg.OSAEnabled)

// Make available to handlers
appDependencies := &AppDependencies{
    SandboxService:            sandboxService,
    SandboxIntegrationService: sandboxIntegrationService, // NEW
    // ... other dependencies
}
```

## Database Changes Required

### SQL Query Update
The `UpdateOSAGeneratedAppStatus` query has been updated to accept a `deployment_url` parameter.

**Action Required:**
```bash
cd desktop/backend-go
sqlc generate
```

This will regenerate `internal/database/sqlc/osa.sql.go` with the updated function signature:

```go
type UpdateOSAGeneratedAppStatusParams struct {
    ID            pgtype.UUID `json:"id"`
    Status        *string     `json:"status"`
    ErrorMessage  *string     `json:"error_message"`
    DeploymentUrl *string     `json:"deployment_url"` // NEW
}
```

### Schema Verification
Verify these columns exist in `osa_generated_apps`:
- `status` (VARCHAR)
- `deployment_url` (TEXT)
- `container_id` (VARCHAR) - from sandbox migration
- `sandbox_status` (VARCHAR) - from sandbox migration
- `sandbox_port` (INTEGER) - from sandbox migration
- `metadata` (JSONB) - stores AppConfig

All should exist from migrations 091 (sandbox) and earlier OSA migrations.

## App Metadata Format

Apps should store configuration in the `metadata` JSONB column:

```json
{
  "app_type": "react",
  "framework": "vite",
  "port": 3000,
  "start_command": ["npm", "run", "dev"],
  "build_command": ["npm", "run", "build"],
  "environment": {
    "NODE_ENV": "development",
    "API_URL": "http://localhost:8000"
  },
  "working_dir": "/workspace",
  "dependencies": ["react", "react-dom", "vite"],
  "requires_build": true
}
```

If `metadata` is empty or invalid, the service falls back to sensible defaults based on app type.

## REST API Handlers

You'll likely want to add these endpoints:

### 1. Deploy App to Sandbox
```go
// POST /api/v1/apps/:id/deploy
func HandleDeployAppToSandbox(w http.ResponseWriter, r *http.Request) {
    appIDStr := chi.URLParam(r, "id")
    appID, err := uuid.Parse(appIDStr)
    if err != nil {
        http.Error(w, "Invalid app ID", http.StatusBadRequest)
        return
    }

    // Get app details
    app, err := queries.GetOSAGeneratedApp(r.Context(), pgtype.UUID{Bytes: appID, Valid: true})
    if err != nil {
        http.Error(w, "App not found", http.StatusNotFound)
        return
    }

    // Parse config
    appConfig, err := services.ParseAppConfig(app.Metadata)
    if err != nil {
        http.Error(w, "Invalid app config", http.StatusBadRequest)
        return
    }

    // Determine workspace path (you need to implement this based on your setup)
    workspacePath := getWorkspacePathForApp(appID)

    // Deploy
    err = integrationService.OnAppGenerationComplete(
        r.Context(),
        appID,
        workspacePath,
        appConfig,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "deploying",
    })
}
```

### 2. Stop App Sandbox
```go
// POST /api/v1/apps/:id/sandbox/stop
func HandleStopAppSandbox(w http.ResponseWriter, r *http.Request) {
    appID, _ := uuid.Parse(chi.URLParam(r, "id"))

    err := integrationService.StopAppSandbox(r.Context(), appID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}
```

### 3. Restart App Sandbox
```go
// POST /api/v1/apps/:id/sandbox/restart
func HandleRestartAppSandbox(w http.ResponseWriter, r *http.Request) {
    appID, _ := uuid.Parse(chi.URLParam(r, "id"))

    info, err := integrationService.RestartAppSandbox(r.Context(), appID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(info)
}
```

### 4. Get App Sandbox Info
```go
// GET /api/v1/apps/:id/sandbox
func HandleGetAppSandboxInfo(w http.ResponseWriter, r *http.Request) {
    appID, _ := uuid.Parse(chi.URLParam(r, "id"))

    info, err := integrationService.GetAppSandboxInfo(r.Context(), appID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(info)
}
```

### 5. List Workspace Sandboxes
```go
// GET /api/v1/workspaces/:id/sandboxes
func HandleListWorkspaceSandboxes(w http.ResponseWriter, r *http.Request) {
    workspaceID, _ := uuid.Parse(chi.URLParam(r, "id"))

    sandboxes, err := integrationService.ListWorkspaceAppSandboxes(r.Context(), workspaceID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(sandboxes)
}
```

## Configuration

Add to `.env` or config:

```bash
# Sandbox Integration
OSA_ENABLED=true
OSA_WORKSPACE_BASE_PATH=/path/to/osa_workspaces
SANDBOX_AUTO_DEPLOY=true
```

Update `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields

    // OSA Integration
    OSAEnabled            bool   `mapstructure:"OSA_ENABLED"`
    OSAWorkspaceBasePath  string `mapstructure:"OSA_WORKSPACE_BASE_PATH"`
    SandboxAutoDeploy     bool   `mapstructure:"SANDBOX_AUTO_DEPLOY"`
}
```

## Testing

### Run Unit Tests
```bash
cd desktop/backend-go
go test ./internal/services/sandbox_integration_test.go -v
```

Expected output: All 12 test suites pass (40+ individual tests).

### Manual Testing Flow

1. **Generate an app via OSA** (or insert test data):
```sql
INSERT INTO osa_generated_apps (
    workspace_id, name, display_name, status, metadata
) VALUES (
    '<workspace-uuid>',
    'test-app',
    'Test React App',
    'generated',
    '{"app_type": "react", "port": 3000}'::jsonb
);
```

2. **Trigger deployment**:
```go
appID := uuid.MustParse("...")
workspacePath := "/tmp/test-workspace"
appConfig := &services.AppConfig{
    AppType: "react",
    Port:    3000,
}

err := integrationService.OnAppGenerationComplete(ctx, appID, workspacePath, appConfig)
```

3. **Verify in database**:
```sql
SELECT id, name, status, deployment_url, container_id, sandbox_status
FROM osa_generated_apps
WHERE id = '<app-uuid>';
```

Expected:
- `status` = 'deployed'
- `deployment_url` = 'http://localhost:XXXX'
- `container_id` populated
- `sandbox_status` = 'running'

4. **Check Docker**:
```bash
docker ps | grep test-app
```

Should show running container with port mapping.

5. **Access sandbox**:
```bash
curl http://localhost:XXXX
```

Should return app response (or error if app needs build step).

## Known Limitations & Future Work

### Current Limitations
1. **No Build Support:** Apps must be pre-built or run in development mode
2. **Single Container:** Multi-service apps (DB, cache) not supported
3. **Static Paths:** Workspace paths must be absolute
4. **No Auto-Scaling:** Fixed resource limits
5. **No SSL:** Only HTTP support

### Planned Enhancements
1. **Build Pipeline:** Detect and run build commands before deployment
2. **Dependency Installation:** Auto-run `npm install`, `pip install`, etc.
3. **Health Checks:** Custom health check endpoints per app type
4. **Environment Templates:** Pre-configured environments for popular frameworks
5. **Multi-Container:** Docker Compose support for complex apps
6. **Port Auto-Detection:** Scan app code to determine port
7. **Hot Reload:** File watcher for development mode
8. **Metrics:** Prometheus metrics for deployments

## Troubleshooting

### Error: "workspace path is empty"
**Cause:** `workspacePath` parameter is empty string
**Fix:** Ensure you pass the full path to generated app files

### Error: "app status not ready for deployment"
**Cause:** App status is not `generated` or `built`
**Fix:** Check app status in database, ensure OSA completed generation

### Error: "could not determine docker image"
**Cause:** Unknown app type (falls back to node:20-alpine)
**Fix:** Verify `app_type` in metadata is valid or add to image map

### Error: "failed to allocate port"
**Cause:** All ports in range exhausted
**Fix:** Release unused ports or increase port range in config

### Error: "container creation failed"
**Cause:** Docker daemon issue, image pull failed, or invalid config
**Fix:** Check Docker daemon is running, image exists, workspace path is valid

### Status stuck on "deploying"
**Cause:** Deployment crashed before status update
**Fix:** Check logs, manually update status, restart deployment

## Security Considerations

1. **Path Validation:** All workspace paths are made absolute to prevent traversal
2. **Resource Limits:** Memory and CPU quotas enforced per container
3. **User Quotas:** Enforced by SandboxDeploymentService
4. **Container Isolation:** Docker security via seccomp profiles
5. **Port Allocation:** Secure port allocation prevents conflicts

## Support & Documentation

- **Service Docs:** `internal/services/SANDBOX_INTEGRATION_SUMMARY.md`
- **Sandbox Docs:** `internal/services/SANDBOX_DEPLOYMENT_SERVICE.md`
- **Container Docs:** `internal/container/README.md`
- **Database Schema:** `internal/database/schema.sql`

## Checklist for Integration

- [ ] Run `sqlc generate` to regenerate database code
- [ ] Initialize `SandboxIntegrationService` in server startup
- [ ] Configure workspace base path
- [ ] Choose integration approach (webhook, polling, event listener)
- [ ] Implement chosen integration
- [ ] Add REST API handlers (optional but recommended)
- [ ] Update frontend to show sandbox status
- [ ] Test with sample app generation
- [ ] Monitor logs for errors
- [ ] Set up metrics/monitoring (future)

## Questions?

For questions or issues:
1. Check logs: `slog` structured logging with component "sandbox_integration"
2. Review error messages in `osa_generated_apps.error_message`
3. Check Docker container logs: `docker logs <container-id>`
4. Verify database state: Check app status, deployment_url, container_id

---

**Created:** 2025-02-03
**Author:** Claude Code
**Version:** 1.0
**Status:** Ready for Integration
**Task:** TASK-BE-008
