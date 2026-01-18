# OSA-5 Technical Reference

> **Complete technical documentation for the BusinessOS + OSA-5 integration**
> Version: 1.0.0 | Last Updated: January 2026

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [API Reference](#2-api-reference)
3. [Database Schema](#3-database-schema)
4. [SSE Streaming](#4-sse-streaming)
5. [Deployment](#5-deployment)
6. [Testing](#6-testing)
7. [Troubleshooting](#7-troubleshooting)

---

## 1. Architecture Overview

### System Diagram

```
+------------------------------------------------------------------+
|                         CLIENT LAYER                              |
+------------------------------------------------------------------+
|  Browser (localhost:5173)                                         |
|  +----------------------------+  +-----------------------------+  |
|  | SvelteKit Frontend        |  | Terminal (WebSocket PTY)    |  |
|  | - OSA UI Components       |  | - osa CLI commands          |  |
|  | - SSE Event Listeners     |  | - businessos_init.sh        |  |
|  +------------+---------------+  +-------------+---------------+  |
|               |                                |                  |
+---------------|--------------------------------|------------------+
                | HTTP/SSE                       | HTTP POST
                v                                v
+------------------------------------------------------------------+
|                         API LAYER (Go)                            |
+------------------------------------------------------------------+
|  BusinessOS Backend (localhost:8001)                              |
|  +------------------------+  +--------------------------------+   |
|  | Gin Router             |  | Auth Middleware                |   |
|  | - /api/osa/*           |  | - JWT Validation               |   |
|  | - /api/internal/osa/*  |  | - HMAC Webhook Verification    |   |
|  +------------------------+  +--------------------------------+   |
|                                                                   |
|  +------------------------+  +--------------------------------+   |
|  | OSAWorkflowsHandler    |  | OSAWebhooksHandler             |   |
|  | - ListWorkflows        |  | - HandleWorkflowComplete       |   |
|  | - GetWorkflow          |  | - HandleBuildEvent             |   |
|  | - GetWorkflowFiles     |  | - RegisterWebhook              |   |
|  | - InstallModule        |  +--------------------------------+   |
|  +------------------------+                                       |
|                                                                   |
|  +------------------------+  +--------------------------------+   |
|  | OSAStreamingHandler    |  | BuildEventBus (Pub/Sub)        |   |
|  | - SSE Connections      |  | - Subscribe/Unsubscribe        |   |
|  | - Heartbeat (30s)      |  | - Publish to clients           |   |
|  +------------------------+  +--------------------------------+   |
|                                                                   |
|  +------------------------+  +--------------------------------+   |
|  | OSAFileSyncService     |  | ResilientClient                |   |
|  | - Poll OSA workspace   |  | - Circuit Breaker              |   |
|  | - Parse file bundles   |  | - Retry Logic                  |   |
|  | - Update database      |  | - Fallback Cache               |   |
|  +------------------------+  +--------------------------------+   |
+------------------------------------------------------------------+
                |                                |
                | pgx/v5                         | HTTP + Webhooks
                v                                v
+---------------------------+    +----------------------------------+
|     DATA LAYER            |    |        OSA-5 LAYER               |
+---------------------------+    +----------------------------------+
| PostgreSQL (Supabase)     |    | OSA-5 Orchestrator (port 3003)   |
| - osa_modules             |    | +------------------------------+ |
| - osa_workspaces          |    | | 21-Agent System              | |
| - osa_generated_apps      |    | | - Architect Agent            | |
| - osa_build_events        |    | | - Code Agent (Svelte+Go)     | |
| - osa_sync_status         |    | | - Test Agent                 | |
| - osa_execution_history   |    | | - Integration Agent          | |
| - osa_webhooks            |    | +------------------------------+ |
+---------------------------+    |                                  |
                                 | +------------------------------+ |
+---------------------------+    | | Webhook Dispatcher           | |
| Redis (Session Cache)     |    | | - HMAC Signing               | |
+---------------------------+    | | - Retry with Backoff         | |
                                 | +------------------------------+ |
                                 +----------------------------------+
```

### Component Relationships

| Component | Responsibility | Dependencies |
|-----------|---------------|--------------|
| `OSAWorkflowsHandler` | Workflow CRUD, file listing | pgxpool, OSAFileSyncService |
| `OSAWebhooksHandler` | Incoming webhook processing | pgxpool, BuildEventBus |
| `OSAStreamingHandler` | SSE connections | BuildEventBus |
| `BuildEventBus` | Pub/sub for real-time events | None (in-memory) |
| `OSAFileSyncService` | Background polling of OSA workspace | pgxpool, filesystem |
| `ResilientClient` | HTTP calls to OSA-5 with resilience | HTTP client |

### Data Flow: Generate Module

```
1. User: osa generate "expense tracker"
   |
   v
2. Terminal: HTTP POST /api/internal/osa/generate
   {
     "description": "expense tracker",
     "type": "fullstack"
   }
   |
   v
3. Backend: ResilientClient.Generate()
   - Circuit breaker check (CLOSED = proceed)
   - Create osa_generated_apps record (status: generating)
   - Forward to OSA-5: POST http://localhost:3003/api/generate
   |
   v
4. OSA-5: 21-Agent Workflow (30-120s)
   - Architect Agent: Design structure
   - Code Agent: Generate Svelte + Go + SQL
   - Test Agent: Create integration tests
   - Integration Agent: Validate and bundle
   |
   v
5. OSA-5: POST /api/osa/webhooks/build-event (every 5s)
   Headers: X-OSA-Signature: <HMAC-SHA256>
   Body: {"event_type": "build_progress", "workflow_id": "abc123", "data": {"progress": 45}}
   |
   v
6. Backend: OSAWebhooksHandler.HandleBuildEvent()
   - Verify HMAC signature
   - Insert into osa_build_events
   - BuildEventBus.Publish(event)
   |
   v
7. SSE: Client receives event via EventSource
   data: {"event_type":"build_progress","progress_percent":45,"phase":"testing"}
   |
   v
8. OSA-5: POST /api/osa/webhooks/workflow-complete
   Body: {"event_type": "workflow.completed", "status": "success", "workflow_id": "abc123"}
   |
   v
9. Backend: Update osa_generated_apps (status: generated)
   - BuildEventBus.Publish(completion event)
   |
   v
10. Frontend: Display "Module generated successfully"
```

---

## 2. API Reference

### Authentication

All authenticated endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

The token is obtained via the session endpoint after login. The middleware extracts `user_id` from the JWT and injects it into the Gin context.

### Endpoints Summary

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/osa/health` | No | Health check |
| GET | `/api/osa/workflows` | JWT | List user workflows |
| GET | `/api/osa/workflows/:id` | JWT | Get workflow details |
| GET | `/api/osa/workflows/:id/files` | JWT | List workflow files |
| GET | `/api/osa/workflows/:id/files/:type` | JWT | Get file by type |
| GET | `/api/osa/files/:id/content` | JWT | Get file by UUID |
| POST | `/api/osa/modules/install` | JWT | Install workflow as module |
| POST | `/api/osa/sync/trigger` | JWT | Manual sync trigger |
| GET | `/api/osa/stream/build/:app_id` | JWT | SSE build progress |
| GET | `/api/osa/stream/stats` | JWT | SSE connection stats |
| POST | `/api/osa/webhooks/workflow-complete` | HMAC | Workflow completion |
| POST | `/api/osa/webhooks/build-event` | HMAC | Build progress |
| POST | `/api/osa/webhooks/register` | JWT | Register webhook |
| GET | `/api/osa/webhooks` | JWT | List webhooks |

---

### GET /api/osa/health

Health check endpoint (no authentication required).

**Response 200:**
```json
{
  "enabled": true,
  "status": "healthy",
  "version": "1.0.0",
  "osa_url": "http://localhost:3003"
}
```

---

### GET /api/osa/workflows

List all workflows for the authenticated user.

**Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/osa/workflows
```

**Response 200:**
```json
{
  "workflows": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "expense-tracker",
      "display_name": "Expense Tracker",
      "description": "Expense tracking with receipt upload",
      "workflow_id": "abc12345",
      "status": "generated",
      "files_created": 8,
      "build_status": "success",
      "created_at": "2026-01-09T10:30:00Z",
      "generated_at": "2026-01-09T10:32:00Z",
      "deployed_at": null,
      "workspace_name": "My Workspace"
    }
  ],
  "count": 1
}
```

**Status Values:**
- `generating` - OSA-5 is generating code
- `generated` - Code generation complete
- `deploying` - Module installation in progress
- `deployed` - Installed as BusinessOS module
- `failed` - Generation failed

---

### GET /api/osa/workflows/:id

Get details for a specific workflow. Accepts UUID or 8-char workflow ID prefix.

**Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/osa/workflows/abc12345
```

**Response 200:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "expense-tracker",
  "display_name": "Expense Tracker",
  "description": "Expense tracking with receipt upload",
  "workflow_id": "abc12345",
  "status": "generated",
  "files_created": 8,
  "build_status": "success",
  "metadata": {
    "analysis": "## Analysis Report\n...",
    "architecture": "## Architecture Design\n...",
    "code": "// File: expense_handler.go\n...",
    "quality": "## Quality Report\n...",
    "deployment": "## Deployment Guide\n...",
    "monitoring": "## Monitoring Setup\n...",
    "strategy": "## Implementation Strategy\n...",
    "recommendations": "## Recommendations\n..."
  },
  "error_message": null,
  "error_stack": null,
  "created_at": "2026-01-09T10:30:00Z",
  "generated_at": "2026-01-09T10:32:00Z",
  "deployed_at": null,
  "workspace_name": "My Workspace",
  "workspace_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

---

### GET /api/osa/workflows/:id/files

List all files in a workflow with deterministic UUIDs.

**Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/osa/workflows/abc12345/files
```

**Response 200:**
```json
{
  "workflow_id": "abc12345",
  "files": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "analysis.md",
      "type": "documentation",
      "size": 2048,
      "language": "markdown",
      "created_at": "2026-01-09T10:30:00Z",
      "updated_at": "2026-01-09T10:32:00Z"
    },
    {
      "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
      "name": "internal/handlers/expense_handler.go",
      "type": "backend",
      "size": 4096,
      "language": "go",
      "created_at": "2026-01-09T10:30:00Z",
      "updated_at": "2026-01-09T10:32:00Z"
    },
    {
      "id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
      "name": "frontend/src/routes/(app)/expenses/+page.svelte",
      "type": "frontend",
      "size": 3072,
      "language": "svelte",
      "created_at": "2026-01-09T10:30:00Z",
      "updated_at": "2026-01-09T10:32:00Z"
    }
  ],
  "count": 3
}
```

**File Type Categories:**
- `documentation` - Markdown files (analysis, architecture, etc.)
- `frontend` - Svelte components, TypeScript
- `backend` - Go handlers, services
- `database` - SQL migrations, queries
- `test` - Test files
- `config` - Configuration files

---

### GET /api/osa/workflows/:id/files/:type

Get raw file content by type.

**Valid Types:** `analysis`, `architecture`, `code`, `quality`, `deployment`, `monitoring`, `strategy`, `recommendations`

**Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/osa/workflows/abc12345/files/architecture
```

**Response 200:**
```json
{
  "type": "architecture",
  "content": "## Architecture Design\n\n### Overview\n...",
  "size": 2048
}
```

---

### GET /api/osa/files/:id/content

Get file content by deterministic UUID (useful for code file bundles).

**Request:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/osa/files/b2c3d4e5-f6a7-8901-bcde-f12345678901/content
```

**Response 200:**
```json
{
  "content": "package handlers\n\nimport (\n\t\"net/http\"\n...",
  "file": {
    "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "name": "internal/handlers/expense_handler.go",
    "type": "backend",
    "size": 4096,
    "language": "go",
    "created_at": "2026-01-09T10:30:00Z",
    "updated_at": "2026-01-09T10:32:00Z"
  }
}
```

---

### POST /api/osa/modules/install

Install a workflow as a BusinessOS module.

**Request:**
```bash
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "abc12345",
    "module_name": "expense-tracker",
    "install_path": "/app/modules/expenses",
    "file_ids": ["b2c3d4e5-f6a7-8901-bcde-f12345678901"]
  }' \
  http://localhost:8001/api/osa/modules/install
```

**Response 200:**
```json
{
  "success": true,
  "module_id": "770e8400-e29b-41d4-a716-446655440002",
  "message": "Module installed successfully"
}
```

---

### POST /api/osa/webhooks/workflow-complete

Webhook endpoint for OSA-5 workflow completion (HMAC authentication).

**Request:**
```bash
# Generate HMAC signature
SIGNATURE=$(echo -n '{"event_type":"workflow.completed","workflow_id":"abc12345","status":"success","data":{}}' | openssl dgst -sha256 -hmac "$OSA_WEBHOOK_SECRET" | awk '{print $2}')

curl -X POST \
  -H "X-OSA-Signature: $SIGNATURE" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "workflow.completed",
    "workflow_id": "abc12345",
    "status": "success",
    "timestamp": "2026-01-09T10:32:00Z",
    "data": {
      "files_created": 8,
      "build_status": "success"
    }
  }' \
  http://localhost:8001/api/osa/webhooks/workflow-complete
```

**Response 200:**
```json
{
  "message": "Webhook processed successfully",
  "event_type": "workflow.completed",
  "workflow_id": "abc12345"
}
```

---

### POST /api/osa/webhooks/build-event

Webhook endpoint for build progress events.

**Request:**
```bash
curl -X POST \
  -H "X-OSA-Signature: $SIGNATURE" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "build_progress",
    "workflow_id": "abc12345",
    "timestamp": "2026-01-09T10:31:00Z",
    "data": {
      "progress": 45,
      "message": "Running tests..."
    }
  }' \
  http://localhost:8001/api/osa/webhooks/build-event
```

**Response 200:**
```json
{
  "message": "Build event processed",
  "workflow_id": "abc12345"
}
```

---

## 3. Database Schema

### Entity Relationship Diagram

```
                                    +------------------+
                                    |      user        |
                                    |  (BetterAuth)    |
                                    +--------+---------+
                                             |
                              +--------------+---------------+
                              |                              |
                              v                              v
                   +------------------+           +----------------------+
                   | osa_workspaces   |           | osa_execution_history|
                   +------------------+           +----------------------+
                   | id (UUID PK)     |           | id (UUID PK)         |
                   | user_id (FK)     |           | user_id (FK)         |
                   | name             |           | app_id (FK)          |
                   | mode (2d/3d)     |           | command              |
                   | layout (JSONB)   |           | output               |
                   | template_type    |           | exit_code            |
                   +--------+---------+           +----------------------+
                            |
                            | 1:N
                            v
                   +--------------------+
                   | osa_generated_apps |
                   +--------------------+
                   | id (UUID PK)       |
                   | workspace_id (FK)  |
                   | module_id (FK)     |<---+
                   | name               |    |
                   | osa_workflow_id    |    |
                   | status             |    |
                   | metadata (JSONB)   |    |
                   +--------+-----------+    |
                            |                |
          +-----------------+----------------+----------+
          |                 |                           |
          v                 v                           v
+------------------+  +------------------+    +------------------+
| osa_build_events |  | osa_sync_status  |    | osa_modules      |
+------------------+  +------------------+    +------------------+
| id (UUID PK)     |  | id (UUID PK)     |    | id (UUID PK)     |
| app_id (FK)      |  | entity_type      |    | name             |
| event_type       |  | entity_id (FK)   |    | display_name     |
| phase            |  | sync_status      |    | schema_def (JSON)|
| progress_percent |  | sync_direction   |    | api_def (JSON)   |
| status_message   |  | local_snapshot   |    | ui_def (JSON)    |
+------------------+  +------------------+    | created_by (FK)  |
                                              +------------------+

+------------------+
| osa_webhooks     |
+------------------+
| id (UUID PK)     |
| workspace_id (FK)|
| app_id (FK)      |
| event_type       |
| webhook_url      |
| secret_key       |
| enabled          |
+------------------+
```

### Table Definitions

#### osa_modules

Registry of installed BusinessOS modules.

```sql
CREATE TABLE osa_modules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    module_type VARCHAR(50) NOT NULL,  -- 'builtin', 'generated', 'custom'
    schema_definition JSONB,            -- Database schema
    api_definition JSONB,               -- API endpoints
    ui_definition JSONB,                -- Frontend components
    created_by VARCHAR(255) REFERENCES "user"(id),
    workspace_id UUID REFERENCES osa_workspaces(id),
    status VARCHAR(50) DEFAULT 'draft', -- 'draft', 'active', 'archived', 'failed'
    version VARCHAR(50) DEFAULT '1.0.0',
    metadata JSONB DEFAULT '{}',
    tags TEXT[],
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deployed_at TIMESTAMPTZ,
    CONSTRAINT osa_modules_name_workspace_unique UNIQUE(name, workspace_id)
);

-- Indexes
CREATE INDEX idx_osa_modules_workspace ON osa_modules(workspace_id);
CREATE INDEX idx_osa_modules_created_by ON osa_modules(created_by);
CREATE INDEX idx_osa_modules_status ON osa_modules(status);
CREATE INDEX idx_osa_modules_type ON osa_modules(module_type);
CREATE INDEX idx_osa_modules_tags ON osa_modules USING GIN(tags);
```

#### osa_workspaces

User workspaces with 2D/3D layout configuration.

```sql
CREATE TABLE osa_workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    mode VARCHAR(20) DEFAULT '2d',      -- '2d', '3d', 'hybrid'
    layout JSONB DEFAULT '{}',
    active_modules UUID[] DEFAULT '{}',
    template_type VARCHAR(50) DEFAULT 'business_os',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT osa_workspaces_user_name_unique UNIQUE(user_id, name)
);

-- Indexes
CREATE INDEX idx_osa_workspaces_user ON osa_workspaces(user_id);
CREATE INDEX idx_osa_workspaces_template ON osa_workspaces(template_type);
```

#### osa_generated_apps

Tracks full-stack applications generated by OSA-5.

```sql
CREATE TABLE osa_generated_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES osa_workspaces(id) ON DELETE CASCADE,
    module_id UUID REFERENCES osa_modules(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    osa_workflow_id VARCHAR(255),       -- 8-char OSA workflow ID
    osa_sandbox_id VARCHAR(255),        -- E2B sandbox ID
    code_repository TEXT,
    deployment_url TEXT,
    status VARCHAR(50) DEFAULT 'generated',
    files_created INTEGER DEFAULT 0,
    tests_passed BOOLEAN DEFAULT false,
    build_status VARCHAR(50),
    metadata JSONB DEFAULT '{}',        -- All file contents stored here
    error_message TEXT,
    error_stack TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    generated_at TIMESTAMPTZ,
    deployed_at TIMESTAMPTZ,
    last_build_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX idx_osa_apps_workspace ON osa_generated_apps(workspace_id);
CREATE INDEX idx_osa_apps_module ON osa_generated_apps(module_id);
CREATE INDEX idx_osa_apps_status ON osa_generated_apps(status);
CREATE INDEX idx_osa_apps_workflow ON osa_generated_apps(osa_workflow_id);
```

#### osa_build_events

Real-time build and deployment events.

```sql
CREATE TABLE osa_build_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES osa_generated_apps(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES osa_workspaces(id),
    event_type VARCHAR(50) NOT NULL,    -- 'build_started', 'build_progress', 'build_completed', etc.
    event_data JSONB DEFAULT '{}',
    build_id VARCHAR(255),
    phase VARCHAR(50),                  -- 'planning', 'generation', 'testing', 'deployment'
    progress_percent INTEGER DEFAULT 0,
    status_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_osa_build_app ON osa_build_events(app_id);
CREATE INDEX idx_osa_build_workspace ON osa_build_events(workspace_id);
CREATE INDEX idx_osa_build_type ON osa_build_events(event_type);
CREATE INDEX idx_osa_build_created ON osa_build_events(created_at DESC);
CREATE INDEX idx_osa_build_build_id ON osa_build_events(build_id);
```

#### osa_sync_status

Sync state between BusinessOS and OSA-5.

```sql
CREATE TABLE osa_sync_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,   -- 'user', 'workspace', 'app', 'module'
    entity_id UUID NOT NULL,
    osa_entity_id VARCHAR(255),
    osa_entity_type VARCHAR(50),
    sync_status VARCHAR(50) DEFAULT 'pending',
    last_sync_at TIMESTAMPTZ,
    next_sync_at TIMESTAMPTZ,
    sync_direction VARCHAR(50) DEFAULT 'bidirectional',
    error_count INTEGER DEFAULT 0,
    last_error TEXT,
    local_snapshot JSONB,
    remote_snapshot JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT osa_sync_entity_unique UNIQUE(entity_type, entity_id)
);

-- Indexes
CREATE INDEX idx_osa_sync_entity ON osa_sync_status(entity_type, entity_id);
CREATE INDEX idx_osa_sync_status ON osa_sync_status(sync_status);
CREATE INDEX idx_osa_sync_next ON osa_sync_status(next_sync_at);
```

#### osa_execution_history

Terminal execution logs.

```sql
CREATE TABLE osa_execution_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    app_id UUID REFERENCES osa_generated_apps(id),
    workspace_id UUID REFERENCES osa_workspaces(id),
    command TEXT NOT NULL,
    working_directory TEXT,
    environment_vars JSONB DEFAULT '{}',
    output TEXT,
    error_output TEXT,
    exit_code INTEGER,
    duration_ms INTEGER,
    triggered_by VARCHAR(50),           -- 'user', 'agent', 'workflow', 'cron'
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_osa_exec_user ON osa_execution_history(user_id);
CREATE INDEX idx_osa_exec_app ON osa_execution_history(app_id);
CREATE INDEX idx_osa_exec_workspace ON osa_execution_history(workspace_id);
CREATE INDEX idx_osa_exec_created ON osa_execution_history(created_at DESC);
```

#### osa_webhooks

Webhook configurations for callbacks.

```sql
CREATE TABLE osa_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES osa_workspaces(id) ON DELETE CASCADE,
    app_id UUID REFERENCES osa_generated_apps(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    webhook_url TEXT NOT NULL,
    secret_key VARCHAR(255),
    enabled BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMPTZ,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_osa_webhooks_workspace ON osa_webhooks(workspace_id);
CREATE INDEX idx_osa_webhooks_app ON osa_webhooks(app_id);
CREATE INDEX idx_osa_webhooks_event ON osa_webhooks(event_type);
CREATE INDEX idx_osa_webhooks_enabled ON osa_webhooks(enabled);
```

### Running Migrations

```bash
cd desktop/backend-go
go run ./cmd/migrate
```

Migration file: `internal/database/migrations/042_osa_integration.sql`

---

## 4. SSE Streaming

### Overview

Server-Sent Events (SSE) provide real-time build progress updates to connected clients. The system uses an in-memory pub/sub event bus.

### Architecture

```
+------------------+     +-----------------+     +------------------+
| Webhook Handler  |---->|  BuildEventBus  |---->| SSE Connections  |
| (Publish)        |     |  (Pub/Sub)      |     | (Subscribe)      |
+------------------+     +-----------------+     +------------------+
                               |
                               v
                         +-------------+
                         | Subscribers |
                         | Map[ID]*Sub |
                         +-------------+
```

### Event Types

| Event Type | Description | Progress |
|------------|-------------|----------|
| `connected` | Client connected | - |
| `build_started` | Build initiated | 0% |
| `build_progress` | Build in progress | 1-99% |
| `build_completed` | Build successful | 100% |
| `build_failed` | Build failed | - |
| `test_started` | Tests running | - |
| `test_completed` | Tests passed | - |
| `deploy_started` | Deployment initiated | - |
| `deploy_completed` | Deployment successful | - |

### Event Format

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "app_id": "660e8400-e29b-41d4-a716-446655440001",
  "workspace_id": "770e8400-e29b-41d4-a716-446655440002",
  "event_type": "build_progress",
  "phase": "testing",
  "progress_percent": 75,
  "status_message": "Running integration tests...",
  "data": {
    "tests_passed": 12,
    "tests_total": 15
  },
  "timestamp": "2026-01-09T10:31:30Z"
}
```

### Client Implementation (Svelte)

```typescript
// frontend/src/lib/api/osa/streaming.ts
export function subscribeToBuildProgress(
  appId: string,
  onEvent: (event: BuildEvent) => void,
  onError?: (error: Event) => void
): () => void {
  const eventSource = new EventSource(
    `${PUBLIC_API_URL}/api/osa/stream/build/${appId}`,
    { withCredentials: true }
  );

  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onEvent(data);
    } catch (e) {
      console.error('Failed to parse SSE event:', e);
    }
  };

  eventSource.onerror = (error) => {
    if (onError) onError(error);
    // Automatic reconnection is handled by EventSource
  };

  // Return cleanup function
  return () => eventSource.close();
}
```

```svelte
<!-- frontend/src/lib/components/osa/BuildProgress.svelte -->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { subscribeToBuildProgress } from '$lib/api/osa/streaming';

  export let appId: string;

  let progress = 0;
  let phase = 'initializing';
  let message = '';
  let unsubscribe: (() => void) | null = null;

  onMount(() => {
    unsubscribe = subscribeToBuildProgress(appId, (event) => {
      progress = event.progress_percent;
      phase = event.phase || phase;
      message = event.status_message || '';
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });
</script>

<div class="build-progress">
  <div class="progress-bar" style="width: {progress}%"></div>
  <span class="phase">{phase}</span>
  <span class="message">{message}</span>
</div>
```

### Server Implementation (Go)

```go
// internal/handlers/osa_streaming.go
func (h *OSAStreamingHandler) StreamBuildProgress(c *gin.Context) {
    userID := c.MustGet("userID").(uuid.UUID)
    appID, _ := uuid.Parse(c.Param("app_id"))

    // Set SSE headers
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")
    c.Header("X-Accel-Buffering", "no")

    ctx, cancel := context.WithCancel(c.Request.Context())
    defer cancel()

    subscriber := h.eventBus.Subscribe(ctx, userID, appID)
    defer h.eventBus.Unsubscribe(subscriber.ID)

    flusher := c.Writer.(http.Flusher)

    // Send initial connection event
    c.Writer.WriteString("data: {\"type\":\"connected\"}\n\n")
    flusher.Flush()

    heartbeatTicker := time.NewTicker(30 * time.Second)
    defer heartbeatTicker.Stop()

    for {
        select {
        case event := <-subscriber.Events:
            c.Writer.WriteString(services.FormatSSE(event))
            flusher.Flush()

        case <-heartbeatTicker.C:
            c.Writer.WriteString(": heartbeat\n\n")
            flusher.Flush()

        case <-ctx.Done():
            return
        }
    }
}
```

### Heartbeat

The server sends a heartbeat comment every 30 seconds to keep the connection alive:

```
: heartbeat

```

---

## 5. Deployment

### Docker Compose (Recommended)

**File:** `docker-compose.complete.yml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: businessos-postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: businessos
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d businessos"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: businessos-redis
    ports:
      - "6379:6379"
    environment:
      - REDIS_PASSWORD=${REDIS_PASSWORD:-changeme}
    command: redis-server --requirepass "$${REDIS_PASSWORD}"
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "$REDIS_PASSWORD", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: ./desktop/backend-go
      dockerfile: Dockerfile
    container_name: businessos-backend
    ports:
      - "8001:8001"
    environment:
      - DATABASE_URL=postgresql://postgres:password@postgres:5432/businessos
      - REDIS_URL=redis://redis:6379/0
      - OSA_ENABLED=true
      - OSA_BASE_URL=http://osa-5:3003
      - OSA_WEBHOOK_SECRET=${OSA_WEBHOOK_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      osa-5:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8001/health"]
      interval: 10s
      timeout: 5s
      retries: 5

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: businessos-frontend
    ports:
      - "5173:5173"
    environment:
      - PUBLIC_API_URL=http://localhost:8001
    depends_on:
      - backend

  osa-5:
    build:
      context: ${OSA_PATH:-../OSA-5}
      dockerfile: Dockerfile
    container_name: businessos-osa-5
    ports:
      - "3003:3003"
    environment:
      - PORT=3003
      - BUSINESSOS_URL=http://backend:8001
      - WEBHOOK_SECRET=${OSA_WEBHOOK_SECRET}
    volumes:
      - generated_code:/app/generated
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3003/health"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:
  generated_code:

networks:
  default:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `SECRET_KEY` | Yes | - | JWT signing key (32+ chars) |
| `TOKEN_ENCRYPTION_KEY` | Yes | - | AES encryption key |
| `REDIS_URL` | No | - | Redis connection (optional) |
| `REDIS_PASSWORD` | No | - | Redis password |
| `OSA_ENABLED` | Yes | `false` | Enable OSA integration |
| `OSA_BASE_URL` | Yes | `http://localhost:3003` | OSA-5 server URL |
| `OSA_TIMEOUT` | No | `30s` | HTTP timeout |
| `OSA_MAX_RETRIES` | No | `3` | Max retry attempts |
| `OSA_WEBHOOK_SECRET` | Yes | - | HMAC secret (64+ hex chars) |
| `OSA_SYNC_ENABLED` | No | `true` | Enable background sync |
| `OSA_SYNC_INTERVAL` | No | `60s` | Sync poll interval |
| `OSA_WORKSPACE_PATH` | No | - | Path to OSA generated files |

### Generate Secrets

```bash
# JWT secret key
openssl rand -base64 32

# Token encryption key
openssl rand -base64 32

# Webhook HMAC secret
openssl rand -hex 32
```

### Health Checks

```bash
# Backend health
curl http://localhost:8001/health
# Expected: {"status":"healthy"}

# OSA health
curl http://localhost:8001/api/osa/health
# Expected: {"enabled":true,"status":"healthy","version":"1.0.0"}

# OSA-5 health
curl http://localhost:3003/health
# Expected: {"status":"healthy"}

# PostgreSQL
pg_isready -U postgres -h localhost -p 5432
# Expected: accepting connections

# Redis
redis-cli -a $REDIS_PASSWORD ping
# Expected: PONG
```

### One-Command Startup

```bash
./start-all.sh         # Native startup
./start-all.sh docker  # Docker Compose startup
./start-all.sh status  # Check all services
```

---

## 6. Testing

### Test Structure

```
desktop/backend-go/internal/handlers/
├── osa_workflows_test.go       # Workflow CRUD tests
├── osa_webhooks_test.go        # Webhook handler tests
├── osa_streaming_test.go       # SSE streaming tests
├── osa_integration_test.go     # End-to-end integration tests
├── osa_deployment_test.go      # Deployment handler tests
└── security_audit_test.go      # Security tests
```

### Running Tests

```bash
# Run all tests
cd desktop/backend-go
go test ./...

# Run OSA tests with verbose output
go test -v ./internal/handlers/osa_*.go

# Run specific test
go test -v -run TestOSAIntegration_EventBusToSSE ./internal/handlers/

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./internal/handlers/
```

### Integration Test Example

```go
func TestOSAIntegration_EventBusToSSE(t *testing.T) {
    logger := slog.Default()
    eventBus := services.NewBuildEventBus(logger)
    streamingHandler := NewOSAStreamingHandler(eventBus, logger)

    gin.SetMode(gin.TestMode)
    router := gin.New()

    userID := uuid.New()
    appID := uuid.New()

    router.Use(func(c *gin.Context) {
        c.Set("userID", userID)
        c.Next()
    })
    router.GET("/stream/build/:app_id", streamingHandler.StreamBuildProgress)

    ts := httptest.NewServer(router)
    defer ts.Close()

    // Connect SSE client
    resp, _ := http.Get(ts.URL + "/stream/build/" + appID.String())
    defer resp.Body.Close()

    assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

    // Publish event
    testEvent := services.BuildEvent{
        ID:              uuid.New(),
        AppID:           appID,
        EventType:       "build_progress",
        ProgressPercent: 50,
    }
    eventBus.Publish(testEvent)

    // Verify client receives event (implementation reads from resp.Body)
}
```

### Coverage Requirements

| Package | Minimum Coverage |
|---------|-----------------|
| `handlers/osa_*` | 70% |
| `services/build_event_bus` | 80% |
| `services/osa_file_sync` | 70% |

---

## 7. Troubleshooting

### Common Errors

#### `osa: command not found`

**Cause:** Terminal did not source init script.

**Solution:**
```bash
# Restart terminal session, or manually source:
source ~/BusinessOS-1/desktop/backend-go/internal/terminal/businessos_init.sh

# Verify:
type osa
```

---

#### `Connection refused` from OSA-5

**Cause:** OSA-5 not running on port 3003.

**Solution:**
```bash
# Check if running
lsof -i :3003

# Start OSA-5
cd ~/Desktop/OSA-5
npm start

# Verify
curl http://localhost:3003/health
```

---

#### `{"enabled":false}` from health check

**Cause:** `OSA_ENABLED` not set.

**Solution:**
```bash
# Edit .env
echo "OSA_ENABLED=true" >> desktop/backend-go/.env

# Restart backend
pkill -f "go run"
cd desktop/backend-go && go run ./cmd/server
```

---

#### `Invalid webhook signature`

**Cause:** `OSA_WEBHOOK_SECRET` mismatch between BusinessOS and OSA-5.

**Solution:**
```bash
# Generate new secret
SECRET=$(openssl rand -hex 32)

# Update BusinessOS
echo "OSA_WEBHOOK_SECRET=$SECRET" >> desktop/backend-go/.env

# Update OSA-5
echo "WEBHOOK_SECRET=$SECRET" >> ~/Desktop/OSA-5/.env

# Restart both services
```

---

#### Database connection fails

**Cause:** PostgreSQL not running or wrong credentials.

**Solution:**
```bash
# Check PostgreSQL status
brew services list | grep postgresql

# Start if stopped
brew services start postgresql@16

# Test connection
psql "$DATABASE_URL" -c "SELECT 1"

# Check database exists
psql postgres -c "SELECT datname FROM pg_database WHERE datname='businessos'"

# Create if missing
createdb businessos
```

---

#### SSE connection drops

**Cause:** Nginx buffering or timeout.

**Solution:**
```nginx
# nginx.conf
location /api/osa/stream/ {
    proxy_pass http://backend:8001;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_buffering off;
    proxy_cache off;
    proxy_read_timeout 3600s;
    add_header X-Accel-Buffering no;
}
```

---

#### Workflow not appearing after generation

**Cause:** File sync not running or wrong workspace path.

**Solution:**
```bash
# Check OSA workspace path
echo $OSA_WORKSPACE_PATH

# Verify files exist
ls -la $OSA_WORKSPACE_PATH/

# Check sync service logs
grep "OSAFileSync" logs/server.log

# Manual sync trigger
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/osa/sync/trigger
```

---

### Debugging Commands

```bash
# Backend logs
tail -f desktop/backend-go/logs/server.log

# Database state
psql "$DATABASE_URL" -c "SELECT id, name, status, files_created FROM osa_generated_apps ORDER BY created_at DESC LIMIT 5"

# Build events
psql "$DATABASE_URL" -c "SELECT event_type, phase, progress_percent, created_at FROM osa_build_events ORDER BY created_at DESC LIMIT 10"

# Active SSE connections
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8001/api/osa/stream/stats

# Test webhook signature
BODY='{"event_type":"test"}'
SIG=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$OSA_WEBHOOK_SECRET" | awk '{print $2}')
curl -X POST -H "X-OSA-Signature: $SIG" -H "Content-Type: application/json" -d "$BODY" \
  http://localhost:8001/api/osa/webhooks/build-event
```

---

## Appendix

### File Locations

| Component | Location |
|-----------|----------|
| Backend handlers | `desktop/backend-go/internal/handlers/osa_*.go` |
| Event bus | `desktop/backend-go/internal/services/build_event_bus.go` |
| File sync | `desktop/backend-go/internal/services/osa_file_sync.go` |
| Database migrations | `desktop/backend-go/internal/database/migrations/042_osa_integration.sql` |
| Frontend API | `frontend/src/lib/api/osa/` |
| Frontend components | `frontend/src/lib/components/osa/` |
| Terminal init | `desktop/backend-go/internal/terminal/businessos_init.sh` |
| Docker Compose | `docker-compose.complete.yml` |

### Performance Benchmarks

| Operation | p50 | p95 | p99 |
|-----------|-----|-----|-----|
| List workflows | 50ms | 150ms | 300ms |
| Get workflow | 30ms | 100ms | 200ms |
| List files | 40ms | 120ms | 250ms |
| Get file content | 20ms | 80ms | 150ms |
| Module install | 200ms | 500ms | 1000ms |
| Webhook processing | <50ms | <100ms | <200ms |
| SSE event latency | <10ms | <50ms | <100ms |

### Related Documentation

- Setup Guide: `/Users/ososerious/BusinessOS-1/GETTING_STARTED_OSA.md`
- Integration Guide: `/Users/ososerious/BusinessOS-1/OSA_INTEGRATION_GUIDE.md`
- Complete Setup: `/Users/ososerious/BusinessOS-1/COMPLETE_SETUP_GUIDE.md`
- Backend CLAUDE.md: `/Users/ososerious/BusinessOS-1/desktop/backend-go/CLAUDE.md`

---

**Document Version:** 1.0.0
**Last Updated:** January 2026
**Maintainers:** BusinessOS Engineering Team
