### BusinessOS Future Features - Comprehensive Development Roadmap

**Created:** December 31, 2025
**Assigned To:** Nick (Primary), Team
**Priority:** HIGH
**Status:** Planning

---

## Executive Summary

This document outlines the next major development initiatives for BusinessOS. These features will transform the platform from a single-user system to a full multi-tenant, team-collaborative platform with advanced integrations, mobile support, and intelligent automation.

### Roberto's Vision (Direct Quotes)

> "The important ones are definitely going to be the Integration stuff as well and the mobile thing will be important."

> "The Dashboards of course I want them to be custom where you have to use the OS agent or something to tell the users. Because everyone has different data in some cases, sometimes you want to view it in a different way."

> "The Team collaboration, the Workspace creation, and the Team role creation and all that stuff is super very important as well."

> "We need to make sure users can sign up and each individual user has their own memory to their agents, and then there's going to be Workspace memory as well, like based on the nodes."

> "If they have a role, it's been specific to their profile, and the agent will know based off their role and everything what they're able to do based on the project. That way, people don't do things outside their role."

---

## Feature Overview

| # | Feature Area | Priority | Complexity | Target |
|---|--------------|----------|------------|--------|
| 1 | Team/Collaboration & Workspaces | CRITICAL | High | Q1 |
| 2 | Mobile API | HIGH | Medium | Q1 |
| 3 | MCP Tool Integration | HIGH | Medium | Q1 |
| 4 | Analytics/Custom Dashboards | HIGH | High | Q2 |
| 5 | Notifications System | MEDIUM | Medium | Q2 |
| 6 | Voice/Audio Improvements | MEDIUM | Medium | Q2 |
| 7 | RAG/Embeddings Enhancement | MEDIUM | High | Q2 |
| 8 | Calendar/Scheduling | MEDIUM | Medium | Q2 |
| 9 | Webhooks & Integrations | MEDIUM | Medium | Q3 |
| 10 | Background Jobs System | MEDIUM | Medium | Q3 |

---

## FEATURE 1: Team/Collaboration & Workspaces (CRITICAL)

### 1.1 Core Requirements

Roberto's exact requirements:

> "The Team collaboration, the Workspace creation, and the Team role creation and all that stuff is super very important as well. We need to make sure that they're able to do that."

> "We need to make sure users can sign up and each individual user has their own memory to their agents, and then there's going to be Workspace memory as well, like based on the nodes."

> "It'll be individual personalized memory from when they use the system. The system's not personal, so it'll just be personal to what they're working on, to their tasks in relation."

> "If they have a role, it's been specific to their profile and stuff. The agent will know based off their role and everything what they're able to do based on the project. That way, people don't do things outside their role."
Pedro / Nick discuss best routes 
### 1.2 Key Concepts

#### Memory Hierarchy
```
WORKSPACE MEMORY (Shared across team)
 └── Based on Nodes
 └── Shared context, decisions, patterns
 └── Accessible to all workspace members

USER MEMORY (Individual)
 └── Personal to what they're working on
 └── Related to their assigned tasks
 └── Tied to their role and profile
 └── Private agent interactions
```
Pedro
#### Role-Based Agent Behavior
- Agents MUST know the user's role
- Agents MUST know what actions the user can perform
- Agents MUST restrict suggestions/actions to user's permissions
- Agents use role context to personalize responses
Pedro / nick
### 1.3 Database Schema

```sql
-- ===== WORKSPACE & TEAM SYSTEM =====

-- Workspaces (multi-tenant containers)
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Workspace Identity
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,        -- URL-friendly identifier
    description TEXT,
    logo_url VARCHAR(500),

    -- Billing & Limits
    plan_type VARCHAR(50) DEFAULT 'free',     -- 'free', 'starter', 'professional', 'enterprise'
    max_members INTEGER DEFAULT 5,
    max_projects INTEGER DEFAULT 10,
    max_storage_gb INTEGER DEFAULT 5,

    -- Settings
    settings JSONB DEFAULT '{}',

    -- Owner
    owner_id VARCHAR(255) NOT NULL,           -- User who created the workspace

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_workspaces_slug ON workspaces(slug);

-- Workspace Members
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Role Assignment
    role_id UUID REFERENCES workspace_roles(id),
    role_name VARCHAR(100),                   -- Denormalized for quick access

    -- Status
    status VARCHAR(50) DEFAULT 'active',      -- 'active', 'invited', 'suspended'
    invited_by VARCHAR(255),
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ,

    -- Permissions Override (if different from role)
    custom_permissions JSONB,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);

-- Workspace Roles (customizable per workspace)
CREATE TABLE workspace_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Role Identity
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(20),                        -- For UI display
    icon VARCHAR(50),

    -- Permissions Matrix
    permissions JSONB NOT NULL DEFAULT '{}',
    /*
    Example permissions structure:
    {
        "projects": {
            "create": true,
            "read": true,
            "update": true,
            "delete": false,
            "manage_members": false
        },
        "tasks": {
            "create": true,
            "read": true,
            "update": "own",           -- Can only update own tasks
            "delete": "own",
            "assign": false
        },
        "clients": {
            "create": false,
            "read": true,
            "update": false,
            "delete": false
        },
        "artifacts": {
            "create": true,
            "read": true,
            "update": "own",
            "delete": "own"
        },
        "settings": {
            "view": false,
            "manage": false
        },
        "team": {
            "view": true,
            "invite": false,
            "manage_roles": false
        },
        "agent": {
            "use_all_agents": true,
            "create_custom_agents": false,
            "access_workspace_memory": true,
            "modify_workspace_memory": false
        }
    }
    */

    -- Role Type
    is_system BOOLEAN DEFAULT FALSE,          -- System roles can't be deleted
    is_default BOOLEAN DEFAULT FALSE,         -- Assigned to new members by default

    -- Hierarchy
    hierarchy_level INTEGER DEFAULT 0,        -- Higher = more permissions

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(workspace_id, name)
);

CREATE INDEX idx_workspace_roles_workspace ON workspace_roles(workspace_id);

-- Default System Roles (seeded per workspace)
-- These are created when a workspace is created:
-- 1. owner (full access, cannot be deleted)
-- 2. admin (full access except billing/ownership transfer)
-- 3. manager (can manage projects, tasks, team members)
-- 4. member (standard access - create/edit own work)
-- 5. viewer (read-only access)
-- 6. guest (limited read access to specific projects)

-- Workspace Memory (shared context across team)
CREATE TABLE workspace_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Memory Identity
    title VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    content TEXT NOT NULL,

    -- Memory Type
    memory_type VARCHAR(50) NOT NULL,         -- 'decision', 'process', 'knowledge', 'pattern', 'policy'
    category VARCHAR(100),

    -- Scope
    scope_type VARCHAR(50) DEFAULT 'workspace', -- 'workspace', 'project', 'node'
    scope_id UUID,                            -- If scoped to project/node

    -- Access Control
    visibility VARCHAR(50) DEFAULT 'team',    -- 'team', 'managers', 'admins'
    created_by VARCHAR(255) NOT NULL,

    -- Relevance
    importance_score DECIMAL(3,2) DEFAULT 0.5,
    access_count INTEGER DEFAULT 0,

    -- Embeddings
    embedding vector(1536),

    -- Metadata
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',

    is_active BOOLEAN DEFAULT TRUE,
    is_pinned BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_workspace_memories_workspace ON workspace_memories(workspace_id);
CREATE INDEX idx_workspace_memories_scope ON workspace_memories(scope_type, scope_id);
CREATE INDEX idx_workspace_memories_embedding ON workspace_memories USING ivfflat (embedding vector_cosine_ops);

-- User Profiles within Workspace
CREATE TABLE user_workspace_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,

    -- Profile Info (can differ per workspace)
    display_name VARCHAR(255),
    title VARCHAR(100),                       -- Job title in this workspace
    department VARCHAR(100),
    avatar_url VARCHAR(500),

    -- Contact
    work_email VARCHAR(255),
    phone VARCHAR(50),

    -- Work Preferences
    timezone VARCHAR(50),
    working_hours JSONB,                      -- {"start": "09:00", "end": "17:00", "days": [1,2,3,4,5]}
    notification_preferences JSONB,

    -- Agent Personalization (per workspace)
    preferred_output_style VARCHAR(50),
    communication_preferences JSONB,
    expertise_areas TEXT[],

    -- User's personal memory in this workspace
    -- (References to memories table filtered by workspace_id)

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(workspace_id, user_id)
);

CREATE INDEX idx_user_workspace_profiles_workspace ON user_workspace_profiles(workspace_id);
CREATE INDEX idx_user_workspace_profiles_user ON user_workspace_profiles(user_id);

-- Project Team Assignments (who can access which projects)
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Project-specific role (can override workspace role)
    project_role VARCHAR(50),                 -- 'lead', 'contributor', 'reviewer', 'viewer'

    -- Assignment details
    assigned_by VARCHAR(255),
    assigned_at TIMESTAMPTZ DEFAULT NOW(),

    -- Notification preferences for this project
    notification_level VARCHAR(50) DEFAULT 'all', -- 'all', 'mentions', 'none'

    UNIQUE(project_id, user_id)
);

CREATE INDEX idx_project_members_project ON project_members(project_id);
CREATE INDEX idx_project_members_user ON project_members(user_id);
```
Pedro;
### 1.4 Role-Based Agent Context

The agent system MUST inject role information into context:

```go
// services/role_context.go

type UserRoleContext struct {
    UserID              string
    WorkspaceID         uuid.UUID

    // Role Info
    RoleName            string
    RoleDisplayName     string
    HierarchyLevel      int

    // Permissions
    Permissions         map[string]map[string]interface{}

    // Project-specific
    ProjectRoles        map[uuid.UUID]string  // projectID -> project role

    // Profile
    Title               string
    Department          string
    ExpertiseAreas      []string
}

// GetRoleContextPrompt generates context for the agent about user's role
func (r *UserRoleContext) GetRoleContextPrompt() string {
    return fmt.Sprintf(`
## User Role Context

**User:** %s
**Role:** %s (%s)
**Title:** %s
**Department:** %s

### What This User Can Do:
%s

### What This User CANNOT Do:
%s

### Important:
- Only suggest actions within this user's permissions
- Do not offer to perform restricted actions
- If user asks for something outside their role, explain they need appropriate permissions
- Tailor responses to their expertise level and department
`, r.UserID, r.RoleDisplayName, r.RoleName, r.Title, r.Department,
   r.formatCanDo(), r.formatCannotDo())
}
```
pedro / nick
### 1.5 API Endpoints

```
# Workspaces
POST   /api/workspaces                        # Create workspace
GET    /api/workspaces                        # List user's workspaces
GET    /api/workspaces/:id                    # Get workspace details
PUT    /api/workspaces/:id                    # Update workspace
DELETE /api/workspaces/:id                    # Delete workspace

# Workspace Members
GET    /api/workspaces/:id/members            # List members
POST   /api/workspaces/:id/members/invite     # Invite member
PUT    /api/workspaces/:id/members/:userId    # Update member role
DELETE /api/workspaces/:id/members/:userId    # Remove member

# Workspace Roles
GET    /api/workspaces/:id/roles              # List roles
POST   /api/workspaces/:id/roles              # Create custom role
PUT    /api/workspaces/:id/roles/:roleId      # Update role
DELETE /api/workspaces/:id/roles/:roleId      # Delete role

# Workspace Memory
GET    /api/workspaces/:id/memories           # List workspace memories
POST   /api/workspaces/:id/memories           # Create workspace memory
POST   /api/workspaces/:id/memories/search    # Search workspace memories

# User Profiles
GET    /api/workspaces/:id/profile            # Get my profile in workspace
PUT    /api/workspaces/:id/profile            # Update my profile

# Project Members
GET    /api/projects/:id/members              # List project members
POST   /api/projects/:id/members              # Add member to project
PUT    /api/projects/:id/members/:userId      # Update project role
DELETE /api/projects/:id/members/:userId      # Remove from project
```
javaris 
### 1.6 Testing Checklist

- [ ] Workspace creation with default roles
- [ ] Member invitation flow (email)
- [ ] Role assignment and permissions work
- [ ] Permission checks block unauthorized actions
- [ ] Agent correctly limits suggestions based on role
- [ ] Workspace memory is shared but respects visibility
- [ ] User memory is personal within workspace
- [ ] Project-level role overrides work
- [ ] Role hierarchy is respected

---
nick / pedro

## FEATURE 2: Mobile API (HIGH PRIORITY)
javaris 
### 2.1 Core Requirements

> "The mobile thing will be important."

Mobile API must be optimized for:
- Lower bandwidth usage
- Faster response times
- Offline-first data patterns
- Push notification integration
- Battery-efficient polling
Javaris 
### 2.2 API Design Principles

```go
// Mobile endpoints return lean payloads
// Use ?fields= for field selection
// Use ?include= for nested resources
// Paginate everything
// Return ETags for caching

// Example: GET /api/mobile/v1/tasks?fields=id,title,status,due_date&limit=20

type MobileTaskResponse struct {
    ID       uuid.UUID `json:"id"`
    Title    string    `json:"title"`
    Status   string    `json:"status"`
    DueDate  *string   `json:"due_date,omitempty"`
    // No heavy fields like description, comments, attachments
}

// Separate endpoints for heavy data
// GET /api/mobile/v1/tasks/:id/details   (full task)
// GET /api/mobile/v1/tasks/:id/comments  (paginated)
```
javaris
### 2.3 Mobile-Specific Endpoints

```
# Mobile API v1 (Lightweight)
GET    /api/mobile/v1/sync                    # Get changes since timestamp
GET    /api/mobile/v1/me                      # Current user (minimal)
GET    /api/mobile/v1/projects                # Projects list (minimal)
GET    /api/mobile/v1/tasks                   # Tasks list (minimal)
GET    /api/mobile/v1/tasks/:id/details       # Full task (on demand)
POST   /api/mobile/v1/tasks/quick             # Quick task creation
GET    /api/mobile/v1/notifications           # Unread notifications
POST   /api/mobile/v1/notifications/read      # Mark as read (batch)

# Voice & Quick Capture
POST   /api/mobile/v1/voice/upload            # Upload voice note
POST   /api/mobile/v1/capture/text            # Quick text capture
POST   /api/mobile/v1/capture/photo           # Quick photo capture

# Chat (Optimized)
POST   /api/mobile/v1/chat/message            # Send message (returns summary)
GET    /api/mobile/v1/chat/history            # Paginated, compressed
```
Javaris 
### 2.4 Sync Protocol

```go
// Mobile sync uses delta updates
type SyncRequest struct {
    LastSyncAt    time.Time `json:"last_sync_at"`
    DeviceID      string    `json:"device_id"`
    ResourceTypes []string  `json:"resource_types"` // ["tasks", "projects", "notifications"]
}

type SyncResponse struct {
    ServerTime time.Time              `json:"server_time"`
    Changes    map[string][]Change    `json:"changes"`
    Deleted    map[string][]uuid.UUID `json:"deleted"`
    HasMore    bool                   `json:"has_more"`
}

type Change struct {
    ID        uuid.UUID `json:"id"`
    UpdatedAt time.Time `json:"updated_at"`
    Data      any       `json:"data"`
}
```

---
nick
## FEATURE 3: MCP Tool Integration (HIGH PRIORITY)
nick
### 3.1 Core Requirements

Better tool server management for the 20+ MCP servers.
nick
### 3.2 Tool Registry System

```sql
-- MCP Tool Servers
CREATE TABLE mcp_tool_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Server Identity
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    category VARCHAR(50),                     -- 'productivity', 'development', 'communication', 'data'

    -- Connection
    server_type VARCHAR(50) NOT NULL,         -- 'stdio', 'http', 'websocket'
    connection_config JSONB NOT NULL,
    /*
    {
        "command": "npx",
        "args": ["-y", "@anthropic/mcp-server-memory"],
        "env": {"API_KEY": "..."}
    }
    */

    -- Status
    status VARCHAR(50) DEFAULT 'disconnected', -- 'connected', 'disconnected', 'error'
    last_health_check TIMESTAMPTZ,
    error_message TEXT,

    -- Usage
    is_enabled BOOLEAN DEFAULT TRUE,
    auto_start BOOLEAN DEFAULT TRUE,

    -- Access Control
    allowed_roles TEXT[] DEFAULT '{}',        -- Empty = all roles

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Tool Definitions (from MCP servers)
CREATE TABLE mcp_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id UUID NOT NULL REFERENCES mcp_tool_servers(id) ON DELETE CASCADE,

    -- Tool Identity
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Schema
    input_schema JSONB NOT NULL,

    -- Usage Tracking
    call_count INTEGER DEFAULT 0,
    last_called_at TIMESTAMPTZ,
    avg_response_time_ms INTEGER,

    -- Access Control
    is_enabled BOOLEAN DEFAULT TRUE,
    requires_confirmation BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMPTZ DEFAULT NOW()
);
```
Nick
### 3.3 API Endpoints

```
# MCP Server Management
GET    /api/mcp/servers                       # List configured servers
POST   /api/mcp/servers                       # Add new server
GET    /api/mcp/servers/:id                   # Get server details
PUT    /api/mcp/servers/:id                   # Update server config
DELETE /api/mcp/servers/:id                   # Remove server
POST   /api/mcp/servers/:id/connect           # Connect to server
POST   /api/mcp/servers/:id/disconnect        # Disconnect
GET    /api/mcp/servers/:id/health            # Health check

# Tools
GET    /api/mcp/tools                         # List all available tools
POST   /api/mcp/tools/:id/call                # Call a tool directly
GET    /api/mcp/tools/:id/schema              # Get tool input schema
```

---

## FEATURE 4: Analytics/Custom Dashboards (HIGH PRIORITY)

### 4.1 Core Requirements

Roberto's exact requirement:

> "The Dashboards of course I want them to be custom where you have to use the OS agent or something to tell the users. Because everyone has different data in some cases, sometimes you want to view it in a different way."
Javaris / nick
### 4.2 Key Concept: Agent-Configured Dashboards

Instead of static dashboards:
1. User tells OS agent what they want to see
2. Agent configures dashboard widgets
3. Dashboards are saved and can be modified via conversation
4. Different users see different dashboards based on their needs/role
javaris / nick
### 4.3 Database Schema

```sql
-- User Dashboards (agent-configured)
CREATE TABLE user_dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Dashboard Identity
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,

    -- Layout Configuration (set by agent)
    layout JSONB NOT NULL DEFAULT '[]',
    /*
    [
        {
            "widget_id": "uuid",
            "widget_type": "task_summary",
            "position": {"x": 0, "y": 0, "w": 4, "h": 2},
            "config": {...}
        }
    ]
    */

    -- Context (how was this created?)
    created_via VARCHAR(50) DEFAULT 'agent',  -- 'agent', 'manual', 'template'
    creation_context TEXT,                    -- The conversation that created it

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Dashboard Widgets
CREATE TABLE dashboard_widgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Widget Identity
    widget_type VARCHAR(100) NOT NULL,        -- 'task_summary', 'project_progress', 'chart', etc.
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Configuration Schema
    config_schema JSONB NOT NULL,             -- JSON Schema for widget config

    -- Default Config
    default_config JSONB DEFAULT '{}',

    -- Rendering
    component_name VARCHAR(100) NOT NULL,     -- Frontend component to render

    -- Access
    min_role_level INTEGER DEFAULT 0,

    is_system BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Available widget types:
-- task_summary, task_list, task_calendar
-- project_progress, project_timeline, project_burndown
-- client_pipeline, client_activity
-- team_capacity, team_workload
-- metric_card, metric_chart, metric_table
-- recent_activity, upcoming_deadlines
-- custom_query (for power users)
```
pedro / nick / javaris
### 4.4 Agent Dashboard Commands

```go
// The agent can configure dashboards through tool calls

type DashboardConfigTool struct {
    Name        string
    Description string
}

func NewDashboardConfigTool() *DashboardConfigTool {
    return &DashboardConfigTool{
        Name: "configure_dashboard",
        Description: `Configure a user's dashboard based on their needs.

        Actions:
        - add_widget: Add a new widget to dashboard
        - remove_widget: Remove a widget
        - move_widget: Change widget position
        - configure_widget: Update widget settings
        - create_dashboard: Create a new dashboard
        - set_default: Set as default dashboard

        Available widget types:
        - task_summary: Overview of tasks by status
        - project_progress: Project completion percentages
        - upcoming_deadlines: Tasks due soon
        - client_pipeline: Sales/client pipeline
        - team_workload: Team member capacity
        - metric_chart: Custom metric visualization
        - activity_feed: Recent activity stream
        `,
    }
}
```

### 4.5 Example Conversation

```
User: "I want to see my tasks organized by project, and I need to track deadlines for this week"

Agent: "I'll configure your dashboard with two widgets:
1. A task summary grouped by project
2. An upcoming deadlines widget showing this week's due dates

[Calls configure_dashboard tool]

Done! Your dashboard now shows:
- Tasks by Project (left side, 6 columns)
- This Week's Deadlines (right side, 6 columns)

Would you like me to add anything else, like a project progress chart?"
```

---
javaris 
## FEATURE 5: Notifications System

### 5.1 Database Schema

```sql
-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Notification Content
    type VARCHAR(100) NOT NULL,               -- 'task_assigned', 'mention', 'deadline', 'comment', etc.
    title VARCHAR(255) NOT NULL,
    body TEXT,

    -- Reference
    entity_type VARCHAR(50),                  -- 'task', 'project', 'comment', 'artifact'
    entity_id UUID,

    -- Sender
    sender_id VARCHAR(255),
    sender_name VARCHAR(255),

    -- Status
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,

    -- Delivery
    channels_sent TEXT[] DEFAULT '{}',        -- 'in_app', 'email', 'push'

    -- Priority
    priority VARCHAR(20) DEFAULT 'normal',    -- 'low', 'normal', 'high', 'urgent'

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_entity ON notifications(entity_type, entity_id);

-- Notification Preferences
CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Channel Preferences
    email_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    in_app_enabled BOOLEAN DEFAULT TRUE,

    -- Type-specific settings
    type_settings JSONB DEFAULT '{}',
    /*
    {
        "task_assigned": {"email": true, "push": true, "in_app": true},
        "mention": {"email": true, "push": true, "in_app": true},
        "deadline_reminder": {"email": true, "push": false, "in_app": true}
    }
    */

    -- Quiet Hours
    quiet_hours_enabled BOOLEAN DEFAULT FALSE,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    quiet_hours_timezone VARCHAR(50),

    UNIQUE(user_id, workspace_id)
);
```
javaris / nick / pedro
### 5.2 Real-Time Delivery

```go
// SSE endpoint for real-time notifications
// GET /api/notifications/stream

type NotificationEvent struct {
    Type         string       `json:"type"`         // 'notification', 'read_receipt'
    Notification *Notification `json:"notification,omitempty"`
    ReadIDs      []uuid.UUID  `json:"read_ids,omitempty"`
}
```

---

## FEATURE 6: Voice/Audio Improvements
nick / pedro 
### 6.1 Improvements Needed

- Better transcription accuracy
- Speaker diarization (who said what)
- Real-time transcription
- Voice commands for quick actions
- Audio summarization
nick / pedro
### 6.2 API Endpoints

```
POST   /api/voice/transcribe                  # Transcribe audio file
POST   /api/voice/transcribe/stream           # Real-time transcription (WebSocket)
POST   /api/voice/command                     # Process voice command
GET    /api/voice/notes                       # List voice notes
POST   /api/voice/notes/:id/summarize         # Generate summary
```

---

## FEATURE 7: RAG/Embeddings Enhancement
pedro
### 7.1 Improvements Needed

- Hybrid search (semantic + keyword)
- Better chunking strategies
- Re-ranking for relevance
- Multi-modal embeddings (images, diagrams)
- Embedding cache optimization
pedro
### 7.2 Search Service

```go
type EnhancedSearchService interface {
    // Hybrid search combining semantic and keyword
    HybridSearch(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)

    // Re-rank results for better relevance
    ReRank(ctx context.Context, query string, results []SearchResult) ([]SearchResult, error)

    // Multi-modal search
    SearchWithImage(ctx context.Context, image []byte, textQuery string) ([]SearchResult, error)
}

type SearchOptions struct {
    SemanticWeight  float64   // 0.0 to 1.0
    KeywordWeight   float64   // 0.0 to 1.0
    ReRankEnabled   bool
    MaxResults      int
    Filters         SearchFilters
}
```

---

## FEATURE 8: Calendar/Scheduling
nick
### 8.1 Google Calendar Integration Improvements

- Two-way sync
- Meeting scheduling via agent
- Availability checking
- Calendar-based task suggestions
nick
### 8.2 API Endpoints

```
GET    /api/calendar/events                   # List events
POST   /api/calendar/events                   # Create event
GET    /api/calendar/availability             # Check availability
POST   /api/calendar/schedule                 # Schedule meeting (agent)
POST   /api/calendar/sync                     # Force sync with Google
```

---

## FEATURE 9: Webhooks & External Integrations
nick / javaris
### 9.1 Database Schema

```sql
-- Webhook Endpoints
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- Webhook Config
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(255),                      -- For signature verification

    -- Events to trigger
    events TEXT[] NOT NULL,                   -- ['task.created', 'project.updated', ...]

    -- Filters
    filters JSONB DEFAULT '{}',               -- {"project_id": "uuid", ...}

    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    last_triggered_at TIMESTAMPTZ,
    failure_count INTEGER DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Webhook Delivery Log
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,

    -- Delivery Details
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,

    -- Response
    status_code INTEGER,
    response_body TEXT,
    response_time_ms INTEGER,

    -- Status
    status VARCHAR(50) NOT NULL,              -- 'success', 'failed', 'pending', 'retrying'
    attempt_count INTEGER DEFAULT 1,
    next_retry_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## FEATURE 10: Background Jobs System
nick / pedro
### 10.1 Requirements

- Reliable task queue for async operations
- Retry logic with exponential backoff
- Job scheduling (cron-like)
- Job monitoring and management
nick / pedro
### 10.2 Database Schema

```sql
-- Background Jobs
CREATE TABLE background_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Job Definition
    job_type VARCHAR(100) NOT NULL,           -- 'email_send', 'report_generate', 'sync_calendar'
    payload JSONB NOT NULL,

    -- Scheduling
    scheduled_at TIMESTAMPTZ DEFAULT NOW(),
    priority INTEGER DEFAULT 0,               -- Higher = sooner

    -- Execution
    status VARCHAR(50) DEFAULT 'pending',     -- 'pending', 'running', 'completed', 'failed', 'cancelled'
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Worker
    worker_id VARCHAR(100),
    locked_until TIMESTAMPTZ,

    -- Retries
    attempt_count INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    last_error TEXT,

    -- Result
    result JSONB,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_background_jobs_status ON background_jobs(status, scheduled_at);
CREATE INDEX idx_background_jobs_type ON background_jobs(job_type);

-- Scheduled Jobs (recurring)
CREATE TABLE scheduled_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Job Definition
    job_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,

    -- Schedule (cron expression)
    cron_expression VARCHAR(100) NOT NULL,    -- "0 9 * * 1-5" (9am weekdays)
    timezone VARCHAR(50) DEFAULT 'UTC',

    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## Implementation Priority Summary

### Q1 (Immediate)
1. **Team/Collaboration** - Foundation for multi-user
2. **Mobile API** - User acquisition
3. **MCP Tool Integration** - Core functionality

### Q2
4. **Custom Dashboards** - User experience
5. **Notifications** - Engagement
6. **Voice Improvements** - Differentiation
7. **RAG Enhancement** - Intelligence

### Q3
8. **Calendar** - Productivity
9. **Webhooks** - Integrations
10. **Background Jobs** - Infrastructure

---

## Environment Variables

```env
# Team Features
MAX_WORKSPACE_MEMBERS=50
DEFAULT_WORKSPACE_PLAN=free

# Mobile
MOBILE_API_RATE_LIMIT=100
MOBILE_SYNC_BATCH_SIZE=50

# Notifications
PUSH_NOTIFICATION_SERVICE=firebase
FCM_SERVER_KEY=xxx

# Background Jobs
JOB_WORKER_COUNT=4
JOB_MAX_RETRIES=3
JOB_RETRY_DELAY_MS=5000
```

---

## Document Version

**Version:** 1.0.0
**Created:** December 31, 2025
**Author:** Roberto (Product Vision), Claude (Technical Specification)
**Primary Assignee:** Nick

---

**This document captures the complete roadmap for BusinessOS future features. The Team/Collaboration system is the foundation that many other features depend on. Implementation should follow the priority order while maintaining all stated objectives.**