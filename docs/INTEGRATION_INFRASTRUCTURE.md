# BusinessOS Integration Infrastructure

## Executive Summary

BusinessOS has a **robust, multi-layered integration system** with:
- **3 fully implemented integrations** (Google Calendar, Slack, Notion)
- **7 additional integrations defined** in the API layer (stubbed endpoints and type definitions)
- **Consistent OAuth 2.0 pattern** across all services
- **MCP tool servers** for AI-accessible integration tools
- **20+ MCP tools** for AI to interact with external services

---

## Current Integration Status

| Integration | Backend | Database | MCP Tools | Frontend API | Frontend UI | Status |
|---|---|---|---|---|---|---|
| **Google Calendar** | FULL | FULL | 5 tools | FULL | NO | PRODUCTION READY |
| **Slack** | FULL | FULL | 6 tools | FULL | NO | PRODUCTION READY |
| **Notion** | FULL | FULL | 7 tools | FULL | NO | PRODUCTION READY |
| **HubSpot** | STUB | NONE | NONE | Types only | NO | PLANNED |
| **GoHighLevel** | STUB | NONE | NONE | Types only | NO | PLANNED |
| **Linear** | STUB | NONE | NONE | Types only | NO | PLANNED |
| **Asana** | STUB | NONE | NONE | Types only | NO | PLANNED |
| **ClickUp** | NONE | NONE | NONE | NONE | NO | NOT STARTED |
| **Todoist** | NONE | NONE | NONE | NONE | NO | NOT STARTED |
| **Microsoft 365** | NONE | NONE | NONE | NONE | NO | NOT STARTED |

---

## Implemented Integrations (Production Ready)

### 1. Google Calendar

**Files:**
- `backend-go/internal/services/google_calendar.go` (283 LOC)
- `backend-go/internal/handlers/google_oauth.go` (200 LOC)
- `backend-go/internal/services/mcp_calendar.go` (200 LOC)
- `backend-go/internal/database/queries/google_oauth.sql`

**Database Table:** `google_oauth_tokens`
```sql
id, user_id (UNIQUE), access_token, refresh_token,
token_type, expiry, scopes, google_email
```

**API Endpoints:**
```
GET  /api/integrations/google/auth      - Initiate OAuth
GET  /api/integrations/google/callback  - OAuth callback
GET  /api/integrations/google/status    - Check connection
DELETE /api/integrations/google         - Disconnect
```

**MCP Tools:**
- `calendar_list_events` - List events within date range
- `calendar_create_event` - Create events with attendees, recurrence, Meet
- `calendar_update_event` - Update existing events
- `calendar_delete_event` - Delete events
- `calendar_sync_events` - Sync to local database

**Features:**
- Auto token refresh
- Event sync to database
- Attendee management
- Recurrence rule support
- Google Meet link generation

---

### 2. Slack

**Files:**
- `backend-go/internal/services/slack.go` (355 LOC)
- `backend-go/internal/handlers/slack_oauth.go` (138 LOC)
- `backend-go/internal/services/mcp_slack.go` (150 LOC)
- `backend-go/internal/database/queries/slack_oauth.sql`
- `backend-go/docs/slack-integration.md`

**Database Table:** `slack_oauth_tokens`
```sql
id, user_id (UNIQUE), workspace_id, workspace_name,
bot_token, user_token, bot_user_id, authed_user_id,
bot_scopes (TEXT[]), user_scopes (TEXT[]),
incoming_webhook_url, incoming_webhook_channel
```

**API Endpoints:**
```
GET  /api/integrations/slack/auth         - Initiate OAuth
GET  /api/integrations/slack/callback     - OAuth callback
GET  /api/integrations/slack/status       - Check status
DELETE /api/integrations/slack            - Disconnect
GET  /api/integrations/slack/channels     - List channels
GET  /api/integrations/slack/notifications - Get messages
```

**MCP Tools:**
- `slack_list_channels` - List channels (public & private)
- `slack_send_message` - Send messages with thread support
- `slack_get_channel_history` - Retrieve channel messages
- `slack_search_messages` - Search across workspace
- `slack_list_users` - List workspace members
- `slack_get_user_info` - Get user details

**Features:**
- Dual-token system (bot + user)
- Incoming webhook URL storage
- Thread support
- Message search
- User management

---

### 3. Notion

**Files:**
- `backend-go/internal/services/notion.go` (582 LOC)
- `backend-go/internal/handlers/notion_oauth.go` (251 LOC)
- `backend-go/internal/services/mcp_notion.go` (200 LOC)
- `backend-go/internal/database/queries/notion_oauth.sql`

**Database Table:** `notion_oauth_tokens`
```sql
id, user_id (UNIQUE), workspace_id, workspace_name,
workspace_icon, access_token, bot_id, owner_type,
owner_user_id, owner_user_name, owner_user_email
```

**API Endpoints:**
```
GET  /api/integrations/notion/auth       - Initiate OAuth
GET  /api/integrations/notion/callback   - OAuth callback
GET  /api/integrations/notion/status     - Check connection
DELETE /api/integrations/notion          - Disconnect
GET  /api/integrations/notion/databases  - List databases
GET  /api/integrations/notion/pages      - List pages
GET  /api/integrations/notion/search     - Search Notion
POST /api/integrations/notion/sync       - Sync database
```

**MCP Tools:**
- `notion_list_databases` - List all accessible databases
- `notion_get_database` - Get database schema
- `notion_query_database` - Query pages with filters/sorts
- `notion_get_page` - Get specific page
- `notion_create_page` - Create new page in database
- `notion_update_page` - Update page properties
- `notion_search` - Search workspace

**Features:**
- Pagination support (cursors)
- Database schema introspection
- Page filtering and sorting
- Property manipulation
- Search with filters

---

## Module Integration Requirements

### Tasks Module

**Current State:**
- Internal task management only
- No external sync

**Data Structure:**
```typescript
Task {
  id, title, description, status, priority,
  projectId, assignee, dueDate, tags,
  subtasks[], comments[], activity[]
}
```

**Integration Needs:**

| Integration | Priority | Sync Type | Features |
|---|---|---|---|
| ClickUp | HIGH | Bi-directional | Status mapping, custom fields, dependencies |
| Asana | HIGH | Bi-directional | Portfolios, custom fields, team sync |
| Todoist | MEDIUM | Bi-directional | Natural language dates, recurring tasks |
| Linear | MEDIUM | Bi-directional | Cycles, project linking, GitHub issues |
| Microsoft To Do | LOW | Bi-directional | Outlook integration |

**Implementation Requirements:**
- Task mapping schema (internal <-> external)
- Status/priority mapping configuration
- Conflict resolution strategy
- Webhook handlers for real-time sync
- Scheduled full sync fallback

---

### Projects Module

**Current State:**
- Basic project CRUD
- Client association
- No external project management sync

**Data Structure:**
```typescript
Project {
  id, name, description, icon, status,
  project_type, priority, client_name
}
```

**Integration Needs:**

| Integration | Priority | Sync Type | Features |
|---|---|---|---|
| Monday.com | HIGH | Bi-directional | Boards, items, timelines |
| Jira | HIGH | Bi-directional | Epics, sprints, issues |
| Linear | HIGH | Bi-directional | Projects, cycles, roadmaps |
| Basecamp | MEDIUM | Bi-directional | Todos, messages, files |
| Notion | MEDIUM | Bi-directional | Database linking |

**Missing Features:**
- Project templates
- Budget/resource tracking
- Timeline/milestone management
- Dependency visualization
- Health/risk tracking

---

### Team Module

**Current State:**
- Basic team member management
- Capacity tracking
- Org chart visualization

**Data Structure:**
```typescript
TeamMember {
  id, name, email, role, avatar_url, status,
  active_projects, open_tasks, capacity,
  manager_id, joined_at, skills[], activities[]
}
```

**Integration Needs:**

| Integration | Priority | Sync Type | Features |
|---|---|---|---|
| Slack | HIGH | Bi-directional | User directory, status sync |
| Google Workspace | HIGH | Bi-directional | Directory, calendar, org chart |
| Microsoft 365 | HIGH | Bi-directional | Azure AD, Teams, Outlook |
| BambooHR | MEDIUM | Import | Employee data, time-off |
| GitHub/GitLab | LOW | Import | Developer profiles, contributions |

**Missing Features:**
- HR system integration
- Availability calendar
- Skill matrix
- Performance metrics
- Workload redistribution

---

### Clients Module

**Current State:**
- Basic client CRUD
- Status pipeline
- No CRM integration

**Data Structure:**
```typescript
Client {
  id, name, email, phone, status, type
}
```

**Integration Needs:**

| Integration | Priority | Sync Type | Features |
|---|---|---|---|
| HubSpot | HIGH | Bi-directional | Contacts, deals, activities |
| Salesforce | HIGH | Bi-directional | Accounts, opportunities, contracts |
| Pipedrive | MEDIUM | Bi-directional | Deals, activities, custom fields |
| Freshsales | MEDIUM | Bi-directional | Lead scoring, email tracking |
| Zendesk | LOW | Import | Support tickets, SLAs |

**Missing Features:**
- Contact person hierarchy
- Engagement history
- Contract management
- Deal/revenue tracking
- Communication history

---

### Calendar Module

**Current State:**
- Google Calendar connected
- Event CRUD
- Meeting notes with AI

**Data Structure:**
```typescript
CalendarEvent {
  id, title, description, start_time, end_time,
  all_day, location, meeting_type, meeting_link,
  attendees[], html_link, meeting_notes,
  meeting_summary, action_items[]
}
```

**Integration Needs:**

| Integration | Priority | Sync Type | Features |
|---|---|---|---|
| Google Calendar | DONE | Bi-directional | Full sync |
| Microsoft Outlook | HIGH | Bi-directional | O365 calendars, Teams |
| Zoom | HIGH | Bi-directional | Meeting creation, recordings |
| Calendly | MEDIUM | Import | Availability, bookings |
| Apple Calendar | LOW | Import | iCal feed |

**Missing Features:**
- Outlook integration
- Recurring event management
- Attendee availability checking
- Video call auto-detection
- Calendar analytics

---

### Nodes/Contexts Module

**Current State:**
- Hierarchical node system
- Multiple view modes (tree, graph, 3D)
- No external knowledge linking

**Integration Needs:**

| Integration | Priority | Sync Type | Features |
|---|---|---|---|
| Notion | HIGH | Bi-directional | Database pages per node |
| Obsidian | MEDIUM | Export/Import | Vault sync, graph |
| Google Drive | MEDIUM | Linking | Folder hierarchy |
| Confluence | LOW | Import | Wiki pages |
| GitHub | LOW | Linking | Repo structure |

---

## Backend Architecture

### OAuth 2.0 Pattern (All Integrations)

```
User clicks "Connect [Service]"
    │
    ▼
GET /api/integrations/{service}/auth
    │ Returns auth_url + sets CSRF cookie
    ▼
Redirect to OAuth provider
    │
    ▼
User approves
    │
    ▼
Provider redirects to /api/integrations/{service}/callback
    │
    ▼
Backend exchanges code for tokens
    │
    ▼
SaveToken() to database
    │
    ▼
Redirect to /settings?{service}_connected=true
```

### Service Layer Pattern

```go
type {Service}Service struct {
    pool         *pgxpool.Pool
    clientID     string
    clientSecret string
    redirectURI  string
    httpClient   *http.Client
}

// Standard methods:
GetAuthURL(state string) string
ExchangeCode(ctx, code string) (*Response, error)
SaveToken(ctx, userID, response)
GetToken(ctx, userID) (*Token, error)
DeleteToken(ctx, userID)
GetConnectionStatus(ctx, userID) (*Status, error)
```

### MCP Tool Pattern

```go
// Each integration provides tools to AI
GetNotionTools() []MCPTool
GetSlackTools() []MCPTool
GetCalendarTools() []MCPTool

// Tools aggregated in chat handler
tools := append([]MCPTool{}, GetNotionTools()...)
tools = append(tools, GetSlackTools()...)
```

---

## Infrastructure Gaps

### Missing Backend Infrastructure

| Component | Status | Priority | Effort |
|---|---|---|---|
| Incoming webhooks | NOT STARTED | HIGH | Medium |
| Outgoing webhooks | NOT STARTED | MEDIUM | Medium |
| Webhook signature verification | NOT STARTED | HIGH | Low |
| Webhook retry/delivery logs | NOT STARTED | MEDIUM | Medium |
| API key management | NOT STARTED | LOW | Low |
| Integration audit logging | NOT STARTED | MEDIUM | Low |
| Scheduled sync jobs | PARTIAL | MEDIUM | Medium |

### Missing Frontend Infrastructure

| Component | Status | Priority | Effort |
|---|---|---|---|
| Integration settings UI | NOT STARTED | HIGH | Medium |
| Connection status dashboard | NOT STARTED | HIGH | Low |
| Sync controls | NOT STARTED | MEDIUM | Low |
| Error handling/retry UI | NOT STARTED | MEDIUM | Medium |
| Integration logs viewer | NOT STARTED | LOW | Medium |

---

## Environment Variables

```bash
# Google Calendar
GOOGLE_CLIENT_ID=xxx
GOOGLE_CLIENT_SECRET=xxx
GOOGLE_REDIRECT_URI=http://localhost:8001/api/integrations/google/callback

# Slack
SLACK_CLIENT_ID=xxx
SLACK_CLIENT_SECRET=xxx
SLACK_REDIRECT_URI=http://localhost:8001/api/integrations/slack/callback

# Notion
NOTION_CLIENT_ID=xxx
NOTION_CLIENT_SECRET=xxx
NOTION_REDIRECT_URI=http://localhost:8001/api/integrations/notion/callback

# Future integrations will follow same pattern:
# {SERVICE}_CLIENT_ID
# {SERVICE}_CLIENT_SECRET
# {SERVICE}_REDIRECT_URI
```

---

## Implementation Roadmap

### Phase 1: Foundation (1-2 weeks)
1. Create Integration Settings UI component
2. Show existing integrations (Google, Slack, Notion) in UI
3. Add connection/disconnection buttons
4. Display sync status and last sync time

### Phase 2: Task Integrations (2-3 weeks)
1. **ClickUp Integration**
   - OAuth handler
   - Task sync service
   - Field mapping configuration
   - MCP tools
2. **Asana Integration**
   - Same pattern as ClickUp

### Phase 3: CRM Integrations (2-3 weeks)
1. **HubSpot Integration**
   - Contact & company sync
   - Deal pipeline mapping
   - Activity logging
2. **Salesforce Integration** (if needed)

### Phase 4: Calendar & Team (2 weeks)
1. **Microsoft Outlook Integration**
   - Calendar sync
   - Teams meeting support
2. **Enhanced team directory sync**
   - Google Workspace
   - Microsoft 365

### Phase 5: Webhook Infrastructure (1-2 weeks)
1. Incoming webhook handlers
2. Event routing system
3. Retry mechanism
4. Delivery logging

### Phase 6: Advanced Features (ongoing)
1. Cross-module data linking
2. Automation workflows
3. Custom integration framework
4. API marketplace

---

## File Structure

```
backend-go/
├── internal/
│   ├── handlers/
│   │   ├── google_oauth.go      # Google Calendar OAuth
│   │   ├── slack_oauth.go       # Slack OAuth
│   │   ├── notion_oauth.go      # Notion OAuth
│   │   ├── hubspot_oauth.go     # (TODO)
│   │   └── clickup_oauth.go     # (TODO)
│   ├── services/
│   │   ├── google_calendar.go   # Google Calendar API
│   │   ├── slack.go             # Slack API
│   │   ├── notion.go            # Notion API
│   │   ├── mcp_calendar.go      # Calendar MCP tools
│   │   ├── mcp_slack.go         # Slack MCP tools
│   │   ├── mcp_notion.go        # Notion MCP tools
│   │   └── mcp.go               # MCP aggregation
│   └── database/
│       └── queries/
│           ├── google_oauth.sql
│           ├── slack_oauth.sql
│           └── notion_oauth.sql

frontend/
├── src/lib/
│   ├── api/
│   │   └── integrations/
│   │       ├── integrations.ts  # API client (346 LOC)
│   │       ├── types.ts         # Type definitions
│   │       └── index.ts
│   └── components/
│       └── settings/
│           └── IntegrationPanel.svelte  # (TODO)
```

---

## Key Takeaways

### Strengths
1. **Consistent OAuth pattern** - Easy to extend
2. **MCP tools for AI** - Integrations accessible to AI agents
3. **Token management** - Secure storage with auto-refresh
4. **Comprehensive type definitions** - Frontend ready for 10+ integrations

### Priority Actions
1. **Build Integration Settings UI** - Users can't see/manage integrations
2. **Add ClickUp/Asana** - Most requested task integrations
3. **Add HubSpot** - CRM integration critical for client management
4. **Implement webhooks** - Real-time sync requires event-driven architecture
5. **Add Outlook** - Microsoft ecosystem support

### Estimated Total Effort
- **Phase 1-2:** 3-5 weeks (Foundation + Tasks)
- **Phase 3-4:** 4-5 weeks (CRM + Calendar)
- **Phase 5-6:** 3-4 weeks (Webhooks + Advanced)
- **Total:** 10-14 weeks for comprehensive integration suite
