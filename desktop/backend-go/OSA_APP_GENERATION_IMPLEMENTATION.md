# OSA App Generation Endpoint Implementation

## Summary

Implemented HTTP endpoint for OSA app generation with Server-Sent Events (SSE) streaming to provide real-time progress updates to the frontend.

## Files Created/Modified

### Created Files

1. **Service Layer**: `internal/services/osa_app_service.go`
   - `OSAAppService` - Business logic for app generation
   - `GenerateApp()` - Returns SSE event channel for streaming
   - `GetAppStatus()` - Get current app status
   - `ListUserApps()` - List all user's apps
   - Automatic workspace creation if not provided
   - Polling OSA API for generation progress
   - Database integration for app tracking

2. **Tests**: `internal/services/osa_app_service_test.go`
   - Unit tests for validation logic
   - Test for non-blocking event channel
   - Structure verification tests

### Modified Files

1. **Handler**: `internal/handlers/osa_api.go`
   - Updated `HandleGenerateApp()` to use SSE streaming
   - Changed from sync JSON response to async SSE streaming
   - Added proper imports (encoding/json, fmt, services)

2. **Handlers Struct**: `internal/handlers/handlers.go`
   - Added `osaAppService *services.OSAAppService` field
   - Updated `SetOSAFileServices()` signature to include appService parameter
   - Routes updated:
     - `POST /api/osa/apps/generate` (was /api/osa/generate)
     - `GET /api/osa/apps/:app_id/status` (was /api/osa/status/:app_id)

3. **Server Initialization**: `cmd/server/main.go`
   - Added `osaAppService` variable declaration
   - Created `osaAppService` after queries initialization
   - Passed `osaAppService` to `SetOSAFileServices()`

## Architecture

### Request Flow

```
Frontend → POST /api/osa/apps/generate
   ↓
Handler (osa_api.go)
   ↓
OSAAppService.GenerateApp() → returns event channel
   ↓
Background goroutine:
   1. Create workspace (if needed)
   2. Create app record (status='generating')
   3. Generate prompt from template
   4. Call OSA API
   5. Poll for completion
   6. Update app status
   ↓
Stream events via SSE → Frontend
```

### SSE Event Types

```go
// Progress events
{
  "type": "progress",
  "content": "Initializing app generation...",
  "data": {
    "percent": 10,
    "phase": "initializing"
  }
}

// Completion event
{
  "type": "done",
  "content": "App generation completed successfully",
  "data": {
    "app_id": "uuid",
    "workspace_id": "uuid",
    "status": "active",
    "osa_app_id": "osa-generated-id",
    "deployment_url": "https://..."
  }
}

// Error event
{
  "type": "error",
  "content": "Error message"
}
```

## Database Integration

### Tables Used

1. **osa_workspaces** - User OSA workspaces
   - Auto-created if not provided
   - Default workspace name: "Default Workspace"

2. **osa_apps** - App registry (from migration 078)
   - Fields: id, workspace_id, name, description, template_type, status
   - JSONB: generation_context, deployment_config, app_metadata
   - Status values: 'generating', 'active', 'failed', 'archived'

### Queries Used

From `internal/database/queries/osa_apps.sql`:
- `CreateOSAApp` - Create new app record
- `GetOSAApp` - Get app by ID
- `ListOSAAppsByUser` - List all user's apps
- `UpdateOSAAppStatus` - Update app status

From `internal/database/queries/osa.sql`:
- `CreateOSAWorkspace` - Create workspace
- `GetOSAWorkspace` - Get workspace by ID

## Integration Points

### OSA Client

Uses existing `osa.Client` from `internal/integrations/osa/`:
- `GenerateApp(ctx, req)` - Trigger generation
- `GetAppStatus(ctx, appID, userID)` - Poll status

### Build Event Bus

Uses existing `services.BuildEventBus` for SSE infrastructure:
- Event publishing/subscribing
- Multiple concurrent streams
- Heartbeat support

### Prompt Template System

**TODO**: Integration pending with prompt template service
- Currently uses description directly as prompt
- Should generate from templates in `osa_prompt_templates` table
- See: `/Users/oso/BusinessOS/desktop/backend-go/docs/OSA_PROMPT_DESIGN.md`

## Next Steps

### 1. Regenerate SQLC Code

```bash
cd /Users/oso/BusinessOS/desktop/backend-go
sqlc generate
```

This will generate Go code for the queries in `osa_apps.sql`:
- `CreateOSAApp` function
- `GetOSAApp` function
- `UpdateOSAAppStatus` function
- `ListOSAAppsByUser` function

### 2. Verify Compilation

```bash
cd /Users/oso/BusinessOS/desktop/backend-go
go build ./cmd/server
```

Expected output: Clean build with no errors

### 3. Run Tests

```bash
# Unit tests
go test ./internal/services/osa_app_service_test.go -v

# All tests
go test ./...
```

### 4. Test Endpoint

```bash
# Start server
go run ./cmd/server

# Test health
curl http://localhost:8080/api/osa/health

# Test generation (requires auth)
curl -X POST http://localhost:8080/api/osa/apps/generate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test App",
    "description": "A test application",
    "type": "full-stack"
  }'
```

### 5. Frontend Integration

Frontend should consume SSE stream:

```javascript
const eventSource = new EventSource('/api/osa/apps/generate', {
  headers: {
    'Authorization': `Bearer ${token}`
  },
  method: 'POST',
  body: JSON.stringify({
    name: 'My App',
    description: 'App description'
  })
});

eventSource.addEventListener('progress', (e) => {
  const data = JSON.parse(e.data);
  console.log(`Progress: ${data.data.percent}% - ${data.content}`);
});

eventSource.addEventListener('done', (e) => {
  const data = JSON.parse(e.data);
  console.log('Complete!', data.data);
  eventSource.close();
});

eventSource.addEventListener('error', (e) => {
  const data = JSON.parse(e.data);
  console.error('Error:', data.content);
  eventSource.close();
});
```

## Configuration

### Environment Variables

Required for OSA integration:
- `OSA_ENABLED=true` - Enable OSA integration
- `OSA_BASE_URL=<url>` - OSA API base URL
- `OSA_SHARED_SECRET=<secret>` - Shared secret for auth

## Error Handling

### Service Layer

1. **Validation Errors** - Return immediately with error
   - Missing name or description
   - Invalid workspace ID

2. **Database Errors** - Log and return error
   - Workspace creation failed
   - App record creation failed

3. **OSA API Errors** - Update app status to 'failed', send error event
   - API call failed
   - Generation timeout (5 minutes)
   - Status polling failed

4. **Non-blocking Events** - Drop events if channel full
   - Prevents goroutine blocking
   - Logs warning on drop

### Handler Layer

1. **Auth Errors** - Return 401 Unauthorized
2. **Validation Errors** - Return 400 Bad Request
3. **Service Errors** - Return 500 Internal Server Error
4. **Streaming Errors** - Close SSE connection

## Code Quality

### Standards Followed

✅ **slog for logging** - No `fmt.Printf` usage
✅ **Context propagation** - All I/O functions accept context
✅ **Error wrapping** - Errors wrapped with context using `fmt.Errorf`
✅ **Handler → Service → Repository** - Layered architecture
✅ **No panic** - All errors handled gracefully
✅ **Non-blocking channels** - Event sending won't block

### Testing

- Unit tests for validation logic
- Mock OSA client for testing
- Non-blocking event channel verification
- Structure tests for service initialization

## Monitoring

### Logs

Service emits structured logs:
- `service="osa_app"` - Service identifier
- User ID, app ID, workspace ID
- Generation phases and progress
- Errors with full context

Example log entries:
```
INFO starting app generation user_id=... name="Test App"
INFO created app record app_id=...
INFO OSA API responded osa_app_id=... status=processing
INFO generation completed app_id=...
ERROR generation polling failed error="timeout" app_id=...
```

### Metrics

Future enhancements:
- Track generation duration
- Count successful/failed generations
- Monitor OSA API response times
- Track workspace creation rate

## Future Enhancements

1. **Prompt Template Integration**
   - Load templates from `osa_prompt_templates` table
   - User/workspace-level customization
   - Variable substitution

2. **Webhook Integration**
   - Receive OSA callbacks instead of polling
   - Reduce API calls and latency
   - Real-time status updates

3. **Retry Logic**
   - Retry failed generations
   - Exponential backoff
   - Max retry limit

4. **Cost Tracking**
   - Track OSA API usage
   - Monitor generation costs
   - User quotas

5. **Advanced Features**
   - App templates library
   - Clone existing apps
   - Version management
   - Rollback capability

## References

- OSA Integration Guide: `/Users/oso/BusinessOS/desktop/backend-go/docs/OSA_INTEGRATION_INDEX.md`
- OSA SSE Streaming: `/Users/oso/BusinessOS/desktop/backend-go/docs/OSA_SSE_STREAMING.md`
- Prompt Design: `/Users/oso/BusinessOS/desktop/backend-go/docs/OSA_PROMPT_DESIGN.md`
- Database Schema: `internal/database/migrations/078_osa_app_registry.sql`
