# BusinessOS Integration Infrastructure

## Executive Summary

BusinessOS has a **robust, multi-layered integration system** designed to connect with 35+ external services across multiple categories. This document details the current state, architecture patterns, and complete infrastructure.

**Current State:**
- **3 Production Integrations:** Google Calendar, Slack, Notion
- **18 MCP Tools:** AI-accessible integration tools
- **4 Stubbed Integrations:** HubSpot, GoHighLevel, Linear, Asana

**Planned Scope:**
- **35+ OAuth Integrations**
- **13+ File Import Parsers**
- **68+ MCP Tools**
- **100+ API Endpoints**
- **20+ Database Tables**

---

## Complete Integration Status Matrix

### Production Ready (3)

| Integration | Backend | Database | MCP Tools | Sync | Webhooks | Frontend UI |
|-------------|---------|----------|-----------|------|----------|-------------|
| **Google Calendar** | FULL | `google_oauth_tokens` | 5 tools | Manual | NO | NO |
| **Slack** | FULL | `slack_oauth_tokens` | 6 tools | Manual | NO | NO |
| **Notion** | FULL | `notion_oauth_tokens` | 7 tools | Manual | NO | NO |

### Stubbed / In Progress (4)

| Integration | Backend | Database | MCP Tools | Priority |
|-------------|---------|----------|-----------|----------|
| **HubSpot** | STUB | NONE | NONE | HIGH |
| **GoHighLevel** | STUB | NONE | NONE | MEDIUM |
| **Linear** | STUB | NONE | NONE | MEDIUM |
| **Asana** | STUB | NONE | NONE | HIGH |

### Planned OAuth Integrations (28+)

| Category | Providers | Count |
|----------|-----------|-------|
| **Google Ecosystem** | Gmail, Drive, Contacts, Meet, Workspace | 5 |
| **Microsoft Ecosystem** | Outlook, Teams, OneDrive, To Do, Azure AD | 5 |
| **Task Management** | ClickUp, Asana, Linear, Monday, Jira, Trello, Todoist | 7 |
| **CRM** | HubSpot, Salesforce, Pipedrive, GoHighLevel, Freshsales | 5 |
| **Video/Meetings** | Zoom, Loom, Fireflies, Fathom, TLDV, Calendly | 6 |
| **Storage** | Dropbox, Box | 2 |
| **Code/Dev** | GitHub, GitLab | 2 |
| **Communication** | Discord | 1 |
| **Support** | Zendesk, Intercom | 2 |
| **Finance** | Stripe, QuickBooks, Xero | 3 |
| **HR** | BambooHR | 1 |
| **Productivity** | Airtable, Figma, Miro, Confluence | 4 |

### File Import Providers (No OAuth)

| Provider | Format | Parser Status | Priority |
|----------|--------|---------------|----------|
| **ChatGPT** | JSON export | NOT STARTED | HIGH |
| **Claude** | JSON export | NOT STARTED | HIGH |
| **Perplexity** | JSON export | NOT STARTED | HIGH |
| **Gemini** | JSON export | NOT STARTED | MEDIUM |
| **Granola** | Meeting JSON | NOT STARTED | MEDIUM |
| **Obsidian** | Vault folder | NOT STARTED | LOW |
| **Roam Research** | JSON export | NOT STARTED | LOW |
| **Evernote** | ENEX format | NOT STARTED | LOW |
| **Notion Export** | ZIP/Markdown | NOT STARTED | LOW |
| **Apple Notes** | iCloud export | NOT STARTED | LOW |
| **CSV** | Generic CSV | NOT STARTED | MEDIUM |
| **Markdown** | Folder structure | NOT STARTED | LOW |

---

## Implemented Integrations Detail

### 1. Google Calendar

**Files:**
```
backend-go/internal/services/google_calendar.go     (283 LOC)
backend-go/internal/handlers/google_oauth.go       (200 LOC)
backend-go/internal/services/mcp_calendar.go       (200 LOC)
backend-go/internal/database/queries/google_oauth.sql
```

**Database Table:** `google_oauth_tokens`
```sql
CREATE TABLE google_oauth_tokens (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL REFERENCES users(id),
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type VARCHAR(50),
    expiry TIMESTAMPTZ,
    scopes TEXT[],
    google_email VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**API Endpoints:**
```
GET    /api/integrations/google/auth       # Initiate OAuth flow
GET    /api/integrations/google/callback   # OAuth callback handler
GET    /api/integrations/google/status     # Check connection status
DELETE /api/integrations/google            # Disconnect integration
```

**MCP Tools:**
| Tool | Description |
|------|-------------|
| `calendar_list_events` | List events within date range |
| `calendar_create_event` | Create event with attendees, recurrence, Meet links |
| `calendar_update_event` | Update existing event properties |
| `calendar_delete_event` | Delete calendar event |
| `calendar_sync_events` | Sync events to local database |

**Features:**
- Auto token refresh on expiry
- Event sync to local database
- Attendee management with email invites
- Recurrence rule support (RRULE)
- Google Meet link auto-generation
- Multiple calendar support

---

### 2. Slack

**Files:**
```
backend-go/internal/services/slack.go              (355 LOC)
backend-go/internal/handlers/slack_oauth.go        (138 LOC)
backend-go/internal/services/mcp_slack.go          (150 LOC)
backend-go/internal/database/queries/slack_oauth.sql
backend-go/docs/slack-integration.md
```

**Database Table:** `slack_oauth_tokens`
```sql
CREATE TABLE slack_oauth_tokens (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL REFERENCES users(id),
    workspace_id VARCHAR(50),
    workspace_name VARCHAR(255),
    bot_token TEXT NOT NULL,
    user_token TEXT,
    bot_user_id VARCHAR(50),
    authed_user_id VARCHAR(50),
    bot_scopes TEXT[],
    user_scopes TEXT[],
    incoming_webhook_url TEXT,
    incoming_webhook_channel VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**API Endpoints:**
```
GET    /api/integrations/slack/auth          # Initiate OAuth
GET    /api/integrations/slack/callback      # OAuth callback
GET    /api/integrations/slack/status        # Check connection
DELETE /api/integrations/slack               # Disconnect
GET    /api/integrations/slack/channels      # List channels
GET    /api/integrations/slack/notifications # Get messages
```

**MCP Tools:**
| Tool | Description |
|------|-------------|
| `slack_list_channels` | List public & private channels |
| `slack_send_message` | Send messages with thread support |
| `slack_get_channel_history` | Retrieve channel messages |
| `slack_search_messages` | Search across workspace |
| `slack_list_users` | List workspace members |
| `slack_get_user_info` | Get user profile details |

**Features:**
- Dual-token system (bot + user tokens)
- Incoming webhook URL storage
- Thread reply support
- Message search across workspace
- User directory access
- Channel membership management

---

### 3. Notion

**Files:**
```
backend-go/internal/services/notion.go             (582 LOC)
backend-go/internal/handlers/notion_oauth.go       (251 LOC)
backend-go/internal/services/mcp_notion.go         (200 LOC)
backend-go/internal/database/queries/notion_oauth.sql
```

**Database Table:** `notion_oauth_tokens`
```sql
CREATE TABLE notion_oauth_tokens (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL REFERENCES users(id),
    workspace_id VARCHAR(100),
    workspace_name VARCHAR(255),
    workspace_icon TEXT,
    access_token TEXT NOT NULL,
    bot_id VARCHAR(100),
    owner_type VARCHAR(50),
    owner_user_id VARCHAR(100),
    owner_user_name VARCHAR(255),
    owner_user_email VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**API Endpoints:**
```
GET    /api/integrations/notion/auth        # Initiate OAuth
GET    /api/integrations/notion/callback    # OAuth callback
GET    /api/integrations/notion/status      # Check connection
DELETE /api/integrations/notion             # Disconnect
GET    /api/integrations/notion/databases   # List databases
GET    /api/integrations/notion/pages       # List pages
GET    /api/integrations/notion/search      # Search workspace
POST   /api/integrations/notion/sync        # Sync database
```

**MCP Tools:**
| Tool | Description |
|------|-------------|
| `notion_list_databases` | List all accessible databases |
| `notion_get_database` | Get database schema and properties |
| `notion_query_database` | Query pages with filters/sorts |
| `notion_get_page` | Get page content and properties |
| `notion_create_page` | Create new page in database |
| `notion_update_page` | Update page properties |
| `notion_search` | Search workspace content |

**Features:**
- Pagination support with cursors
- Database schema introspection
- Page filtering and sorting
- Property manipulation (all types)
- Rich text and block support
- Search with filters

---

## Architecture Patterns

### OAuth 2.0 Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              OAuth 2.0 Flow                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  User clicks "Connect [Service]"                                            │
│       │                                                                     │
│       ▼                                                                     │
│  GET /api/integrations/{service}/auth                                       │
│       │ → Generate state token (CSRF protection)                            │
│       │ → Set state cookie                                                  │
│       │ → Return OAuth authorization URL                                    │
│       ▼                                                                     │
│  Frontend redirects to provider OAuth page                                  │
│       │                                                                     │
│       ▼                                                                     │
│  User approves permissions                                                  │
│       │                                                                     │
│       ▼                                                                     │
│  Provider redirects to /api/integrations/{service}/callback                 │
│       │ → Verify state token matches cookie                                 │
│       │ → Exchange authorization code for tokens                            │
│       │ → Fetch account info from provider                                  │
│       │ → Save tokens to database (encrypted)                               │
│       │ → Create audit log entry                                            │
│       ▼                                                                     │
│  Redirect to /settings?{service}_connected=true                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Provider Interface Pattern

```go
// internal/integrations/types.go

type Provider interface {
    // Identity
    Name() string                           // "google", "slack", "notion"
    DisplayName() string                    // "Google Calendar", "Slack"
    Category() string                       // "calendar", "communication"
    Icon() string                           // URL or icon identifier

    // OAuth
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

    // Connection Management
    GetConnectionStatus(ctx context.Context, userID string) (*ConnectionStatus, error)
    Disconnect(ctx context.Context, userID string) error

    // Token Management
    SaveToken(ctx context.Context, userID string, token *TokenResponse) error
    GetToken(ctx context.Context, userID string) (*Token, error)

    // Sync (optional)
    SupportsSync() bool
    Sync(ctx context.Context, userID string, options SyncOptions) (*SyncResult, error)
}

type TokenResponse struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Scopes       []string
    AccountID    string
    AccountName  string
    AccountEmail string
    Metadata     map[string]interface{}
}

type ConnectionStatus struct {
    Connected    bool
    ConnectedAt  *time.Time
    AccountID    string
    AccountName  string
    AccountEmail string
    Scopes       []string
    LastSyncAt   *time.Time
    SyncStatus   string // "idle", "syncing", "error"
    Error        string
}

type SyncOptions struct {
    Type      string     // "full", "incremental"
    Since     *time.Time
    Resources []string   // ["tasks", "projects"]
}

type SyncResult struct {
    Success       bool
    ItemsCreated  int
    ItemsUpdated  int
    ItemsDeleted  int
    Errors        []string
    NextSyncToken string
    Duration      time.Duration
}
```

### Registry Pattern

```go
// internal/integrations/registry.go

package integrations

import "sync"

var (
    registry = make(map[string]Provider)
    mu       sync.RWMutex
)

// Register adds a provider to the registry
func Register(p Provider) {
    mu.Lock()
    defer mu.Unlock()
    registry[p.Name()] = p
}

// Get retrieves a provider by name
func Get(name string) (Provider, bool) {
    mu.RLock()
    defer mu.RUnlock()
    p, ok := registry[name]
    return p, ok
}

// List returns all registered providers
func List() []Provider {
    mu.RLock()
    defer mu.RUnlock()
    result := make([]Provider, 0, len(registry))
    for _, p := range registry {
        result = append(result, p)
    }
    return result
}

// ListByCategory returns providers in a category
func ListByCategory(category string) []Provider {
    mu.RLock()
    defer mu.RUnlock()
    result := make([]Provider, 0)
    for _, p := range registry {
        if p.Category() == category {
            result = append(result, p)
        }
    }
    return result
}
```

### MCP Tool Pattern

```go
// internal/services/mcp.go

type MCPTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"input_schema"`
    Execute     func(ctx context.Context, userID string, params map[string]interface{}) (interface{}, error)
}

// Each integration provides its tools
func GetCalendarTools() []MCPTool { ... }
func GetSlackTools() []MCPTool { ... }
func GetNotionTools() []MCPTool { ... }

// Aggregation for chat handler
func GetAllTools(userID string, connections []string) []MCPTool {
    tools := make([]MCPTool, 0)

    for _, conn := range connections {
        switch conn {
        case "google":
            tools = append(tools, GetCalendarTools()...)
        case "slack":
            tools = append(tools, GetSlackTools()...)
        case "notion":
            tools = append(tools, GetNotionTools()...)
        }
    }

    return tools
}
```

---

## Module Integration Requirements

### Tasks Module ↔ External Task Managers

**Current Internal Structure:**
```typescript
interface Task {
    id: string;
    title: string;
    description: string;
    status: 'todo' | 'in_progress' | 'review' | 'done';
    priority: 'low' | 'medium' | 'high' | 'urgent';
    project_id?: string;
    assignee_id?: string;
    due_date?: string;
    tags: string[];
    subtasks: Subtask[];
    comments: Comment[];
    created_at: string;
    updated_at: string;
}
```

**External Integration Needs:**

| Provider | Mapping Complexity | Bi-directional | Webhooks Available |
|----------|-------------------|----------------|-------------------|
| ClickUp | HIGH (custom fields, multiple lists) | YES | YES |
| Asana | HIGH (portfolios, custom fields) | YES | YES |
| Linear | MEDIUM (cycles, projects) | YES | YES |
| Todoist | LOW (simple structure) | YES | YES |
| Monday | HIGH (columns, subitems) | YES | YES |
| Jira | HIGH (sprints, epics, workflows) | YES | YES |
| Trello | LOW (cards, lists) | YES | YES |

**Required Sync Infrastructure:**
- `synced_tasks` mapping table
- Status mapping configuration per provider
- Priority mapping configuration per provider
- Conflict resolution strategy (last-write-wins, manual, etc.)
- Webhook handlers for real-time updates
- Fallback scheduled sync (hourly/daily)

---

### Clients Module ↔ CRM Systems

**Current Internal Structure:**
```typescript
interface Client {
    id: string;
    name: string;
    email?: string;
    phone?: string;
    status: 'lead' | 'prospect' | 'active' | 'inactive';
    type: 'individual' | 'company';
    created_at: string;
    updated_at: string;
}
```

**External Integration Needs:**

| Provider | Contact Model | Company Model | Deals | Activities |
|----------|--------------|---------------|-------|------------|
| HubSpot | Contacts | Companies | Deals | Timeline |
| Salesforce | Contacts | Accounts | Opportunities | Activities |
| Pipedrive | Persons | Organizations | Deals | Activities |
| GoHighLevel | Contacts | - | Opportunities | Conversations |
| Freshsales | Contacts | Accounts | Deals | Activities |

**Required Sync Infrastructure:**
- `synced_contacts` mapping table
- `synced_companies` mapping table (optional)
- `synced_deals` mapping table
- Field mapping configuration
- Contact-to-client matching logic
- Deal stage mapping

---

### Calendar Module ↔ Calendar/Meeting Services

**Current Internal Structure:**
```typescript
interface CalendarEvent {
    id: string;
    title: string;
    description?: string;
    start_time: string;
    end_time: string;
    all_day: boolean;
    location?: string;
    meeting_type?: 'call' | 'video' | 'in_person';
    meeting_link?: string;
    attendees: Attendee[];
    html_link?: string;
    meeting_notes?: string;
    meeting_summary?: string;
    action_items?: ActionItem[];
}
```

**External Integration Needs:**

| Provider | Calendar Sync | Meeting Creation | Recordings | Transcripts |
|----------|--------------|------------------|------------|-------------|
| Google Calendar | DONE | DONE | N/A | N/A |
| Outlook | PLANNED | PLANNED | N/A | N/A |
| Zoom | PLANNED | PLANNED | YES | YES |
| Teams | PLANNED | PLANNED | YES | YES |
| Calendly | PLANNED | Import only | N/A | N/A |
| Google Meet | Via Calendar | Via Calendar | YES | NO |

**Required Sync Infrastructure:**
- `synced_events` mapping table
- `synced_meetings` mapping table (for recordings/transcripts)
- Calendar conflict detection
- Attendee availability checking
- Video call auto-detection and linking

---

### Team Module ↔ Directory Services

**Current Internal Structure:**
```typescript
interface TeamMember {
    id: string;
    name: string;
    email: string;
    role: string;
    avatar_url?: string;
    status: 'active' | 'away' | 'offline';
    active_projects: number;
    open_tasks: number;
    capacity: number;
    manager_id?: string;
    joined_at: string;
    skills: string[];
}
```

**External Integration Needs:**

| Provider | User Directory | Status Sync | Org Chart |
|----------|---------------|-------------|-----------|
| Slack | Users API | Presence API | N/A |
| Google Workspace | Directory API | N/A | Org units |
| Microsoft 365 | Graph API | Presence API | Org chart |
| BambooHR | Employees API | N/A | Reports to |
| GitHub | Members API | N/A | Teams |

**Required Sync Infrastructure:**
- `synced_team_members` mapping table
- Avatar sync and caching
- Status/presence real-time updates
- Org chart hierarchy mapping

---

### Knowledge/Contexts Module ↔ Knowledge Bases

**Current Internal Structure:**
```typescript
interface Context {
    id: string;
    title: string;
    content?: string;
    type: 'folder' | 'document' | 'link';
    parent_id?: string;
    children: Context[];
    metadata: Record<string, any>;
}
```

**External Integration Needs:**

| Provider | Sync Type | Content Type | Indexing |
|----------|-----------|--------------|----------|
| Notion | Bi-directional | Pages, Databases | Full-text |
| Google Drive | Import | Docs, Sheets, PDFs | Full-text |
| Obsidian | Import/Export | Markdown files | Full-text |
| Confluence | Import | Wiki pages | Full-text |
| Dropbox | Import | Files | Metadata only |

**Required Sync Infrastructure:**
- `synced_files` mapping table
- Content extraction pipeline
- Vector embedding generation
- Full-text search indexing

---

## Infrastructure Components

### Existing Infrastructure

| Component | Status | Location |
|-----------|--------|----------|
| OAuth handlers | 3 implemented | `internal/handlers/*_oauth.go` |
| Service layers | 3 implemented | `internal/services/*.go` |
| MCP tools | 18 implemented | `internal/services/mcp_*.go` |
| Token storage | 3 tables | `internal/database/queries/*.sql` |
| API types | Comprehensive | `frontend/src/lib/api/integrations/types.ts` |

### Missing Infrastructure

| Component | Priority | Description |
|-----------|----------|-------------|
| Integration registry | CRITICAL | Provider registration and lookup |
| Generic OAuth service | CRITICAL | Reusable OAuth flow handling |
| Settings UI | CRITICAL | User-facing integration management |
| Sync job scheduler | HIGH | Background sync orchestration |
| Webhook receiver | HIGH | Incoming webhook handling |
| Webhook dispatcher | MEDIUM | Outgoing webhook delivery |
| Audit logging | MEDIUM | Integration activity tracking |
| Field mapping config | MEDIUM | Customizable field mappings |
| Conflict resolution | MEDIUM | Sync conflict handling |
| Error handling UI | MEDIUM | User-facing error display |

---

## Complete Backend File Structure

```
backend-go/
├── internal/
│   ├── integrations/                        # NEW: Integration Framework
│   │   ├── types.go                         # Provider interface, types
│   │   ├── registry.go                      # Provider registry
│   │   ├── oauth.go                         # Generic OAuth helpers
│   │   ├── errors.go                        # Integration errors
│   │   ├── config.go                        # Environment config
│   │   │
│   │   ├── google/                          # Google Ecosystem
│   │   │   ├── provider.go                  # Google provider [NEW]
│   │   │   ├── oauth.go                     # Shared Google OAuth [NEW]
│   │   │   ├── calendar.go                  # Calendar service [EXISTS - refactor]
│   │   │   ├── gmail.go                     # Gmail service [NEW]
│   │   │   ├── drive.go                     # Drive service [NEW]
│   │   │   ├── contacts.go                  # Contacts service [NEW]
│   │   │   └── meet.go                      # Meet service [NEW]
│   │   │
│   │   ├── microsoft/                       # Microsoft Ecosystem [ALL NEW]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go                     # Graph API OAuth
│   │   │   ├── outlook.go                   # Outlook Calendar
│   │   │   ├── teams.go                     # Teams
│   │   │   ├── onedrive.go                  # OneDrive
│   │   │   └── todo.go                      # Microsoft To Do
│   │   │
│   │   ├── slack/                           # Slack [EXISTS - refactor]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go
│   │   │   ├── channels.go
│   │   │   ├── messages.go
│   │   │   ├── users.go
│   │   │   └── webhooks.go                  # [NEW]
│   │   │
│   │   ├── notion/                          # Notion [EXISTS - refactor]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go
│   │   │   ├── databases.go
│   │   │   ├── pages.go
│   │   │   └── search.go
│   │   │
│   │   ├── clickup/                         # ClickUp [ALL NEW]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go
│   │   │   ├── workspaces.go
│   │   │   ├── tasks.go
│   │   │   ├── sync.go
│   │   │   └── mapping.go
│   │   │
│   │   ├── asana/                           # Asana [ALL NEW]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go
│   │   │   ├── projects.go
│   │   │   ├── tasks.go
│   │   │   ├── sync.go
│   │   │   └── mapping.go
│   │   │
│   │   ├── linear/                          # Linear [STUB EXISTS - complete]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go
│   │   │   ├── issues.go
│   │   │   ├── projects.go
│   │   │   └── sync.go
│   │   │
│   │   ├── hubspot/                         # HubSpot [STUB EXISTS - complete]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go
│   │   │   ├── contacts.go
│   │   │   ├── companies.go
│   │   │   ├── deals.go
│   │   │   └── sync.go
│   │   │
│   │   ├── zoom/                            # Zoom [ALL NEW]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go
│   │   │   ├── meetings.go
│   │   │   ├── recordings.go
│   │   │   └── transcripts.go
│   │   │
│   │   ├── github/                          # GitHub [ALL NEW]
│   │   │   ├── provider.go
│   │   │   ├── oauth.go
│   │   │   ├── repos.go
│   │   │   ├── issues.go
│   │   │   └── prs.go
│   │   │
│   │   └── ... (20+ more providers)
│   │
│   ├── imports/                             # File Import System [ALL NEW]
│   │   ├── types.go
│   │   ├── service.go                       # Import orchestration
│   │   ├── storage.go                       # File storage (GCS/local)
│   │   ├── processor.go                     # Processing pipeline
│   │   │
│   │   ├── parsers/
│   │   │   ├── parser.go                    # Parser interface
│   │   │   ├── chatgpt.go                   # ChatGPT JSON
│   │   │   ├── claude.go                    # Claude JSON
│   │   │   ├── perplexity.go                # Perplexity JSON
│   │   │   ├── gemini.go                    # Gemini JSON
│   │   │   ├── granola.go                   # Granola meetings
│   │   │   ├── obsidian.go                  # Obsidian vault
│   │   │   ├── roam.go                      # Roam Research
│   │   │   ├── evernote.go                  # Evernote ENEX
│   │   │   ├── csv.go                       # Generic CSV
│   │   │   └── markdown.go                  # Markdown folder
│   │   │
│   │   └── enrichment/                      # AI Enrichment [ALL NEW]
│   │       ├── service.go
│   │       ├── summarizer.go
│   │       ├── topics.go
│   │       ├── entities.go
│   │       ├── knowledge.go
│   │       └── embeddings.go
│   │
│   ├── workers/                             # Background Workers [ALL NEW]
│   │   ├── types.go
│   │   ├── scheduler.go                     # Job scheduler
│   │   ├── sync_runner.go                   # Sync job runner
│   │   ├── import_runner.go                 # Import job runner
│   │   └── enrichment_runner.go             # Enrichment runner
│   │
│   ├── webhooks/                            # Webhook Infrastructure [ALL NEW]
│   │   ├── types.go
│   │   ├── receiver.go                      # Incoming webhooks
│   │   ├── dispatcher.go                    # Outgoing webhooks
│   │   ├── verifier.go                      # Signature verification
│   │   ├── retry.go                         # Retry logic
│   │   └── handlers/
│   │       ├── slack.go
│   │       ├── github.go
│   │       ├── stripe.go
│   │       └── ... (more)
│   │
│   ├── handlers/                            # HTTP Handlers
│   │   ├── integrations.go                  # Unified endpoints [NEW]
│   │   ├── imports.go                       # Import endpoints [NEW]
│   │   ├── webhooks.go                      # Webhook endpoints [NEW]
│   │   ├── google_oauth.go                  # [EXISTS]
│   │   ├── slack_oauth.go                   # [EXISTS]
│   │   └── notion_oauth.go                  # [EXISTS]
│   │
│   ├── services/                            # MCP Tools
│   │   ├── mcp.go                           # Aggregation [EXISTS]
│   │   ├── mcp_calendar.go                  # [EXISTS]
│   │   ├── mcp_slack.go                     # [EXISTS]
│   │   ├── mcp_notion.go                    # [EXISTS]
│   │   ├── mcp_gmail.go                     # [NEW]
│   │   ├── mcp_drive.go                     # [NEW]
│   │   ├── mcp_clickup.go                   # [NEW]
│   │   ├── mcp_asana.go                     # [NEW]
│   │   ├── mcp_hubspot.go                   # [NEW]
│   │   ├── mcp_zoom.go                      # [NEW]
│   │   └── mcp_search.go                    # [NEW]
│   │
│   └── database/
│       ├── migrations/
│       │   ├── 025_integration_connections.sql   # [NEW]
│       │   ├── 026_sync_jobs.sql                 # [NEW]
│       │   ├── 027_integration_audit.sql         # [NEW]
│       │   ├── 028_webhooks.sql                  # [NEW]
│       │   ├── 029_file_imports.sql              # [NEW]
│       │   ├── 030_imported_conversations.sql    # [NEW]
│       │   ├── 031_imported_messages.sql         # [NEW]
│       │   ├── 032_imported_knowledge.sql        # [NEW]
│       │   ├── 033_synced_tasks.sql              # [NEW]
│       │   ├── 034_synced_contacts.sql           # [NEW]
│       │   ├── 035_synced_events.sql             # [NEW]
│       │   ├── 036_synced_files.sql              # [NEW]
│       │   └── 037_synced_meetings.sql           # [NEW]
│       │
│       └── queries/
│           ├── google_oauth.sql                  # [EXISTS]
│           ├── slack_oauth.sql                   # [EXISTS]
│           ├── notion_oauth.sql                  # [EXISTS]
│           ├── integration_connections.sql       # [NEW]
│           ├── sync_jobs.sql                     # [NEW]
│           ├── webhooks.sql                      # [NEW]
│           └── file_imports.sql                  # [NEW]
```

---

## Complete Frontend File Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── api/
│   │   │   ├── integrations/
│   │   │   │   ├── index.ts              # Exports [EXISTS]
│   │   │   │   ├── types.ts              # Types [EXISTS - update]
│   │   │   │   ├── client.ts             # API client [NEW]
│   │   │   │   ├── google.ts             # Google API [NEW]
│   │   │   │   ├── slack.ts              # Slack API [NEW]
│   │   │   │   ├── notion.ts             # Notion API [NEW]
│   │   │   │   ├── clickup.ts            # ClickUp API [NEW]
│   │   │   │   ├── asana.ts              # Asana API [NEW]
│   │   │   │   ├── hubspot.ts            # HubSpot API [NEW]
│   │   │   │   └── zoom.ts               # Zoom API [NEW]
│   │   │   │
│   │   │   └── imports/                  # [ALL NEW]
│   │   │       ├── index.ts
│   │   │       ├── types.ts
│   │   │       └── client.ts
│   │   │
│   │   ├── stores/
│   │   │   ├── integrationsStore.ts      # [NEW]
│   │   │   ├── importsStore.ts           # [NEW]
│   │   │   └── syncStore.ts              # [NEW]
│   │   │
│   │   └── components/
│   │       ├── settings/
│   │       │   └── integrations/         # [ALL NEW]
│   │       │       ├── IntegrationsPage.svelte
│   │       │       ├── IntegrationGrid.svelte
│   │       │       ├── IntegrationCategory.svelte
│   │       │       │
│   │       │       ├── cards/
│   │       │       │   ├── IntegrationCard.svelte
│   │       │       │   ├── GoogleCard.svelte
│   │       │       │   ├── SlackCard.svelte
│   │       │       │   ├── NotionCard.svelte
│   │       │       │   └── GenericCard.svelte
│   │       │       │
│   │       │       ├── modals/
│   │       │       │   ├── ConnectModal.svelte
│   │       │       │   ├── DisconnectModal.svelte
│   │       │       │   ├── SyncSettingsModal.svelte
│   │       │       │   └── LogsModal.svelte
│   │       │       │
│   │       │       └── status/
│   │       │           ├── ConnectionStatus.svelte
│   │       │           ├── SyncStatus.svelte
│   │       │           └── SyncHistory.svelte
│   │       │
│   │       └── imports/                  # [ALL NEW]
│   │           ├── ImportsPage.svelte
│   │           ├── FileDropzone.svelte
│   │           ├── ProviderSelector.svelte
│   │           ├── ImportProgress.svelte
│   │           ├── ImportHistory.svelte
│   │           ├── ConversationList.svelte
│   │           ├── ConversationView.svelte
│   │           └── KnowledgePanel.svelte
│   │
│   └── routes/
│       └── (app)/
│           └── settings/
│               ├── integrations/         # [ALL NEW]
│               │   ├── +page.svelte
│               │   ├── +page.server.ts
│               │   └── [provider]/
│               │       ├── +page.svelte
│               │       └── callback/
│               │           └── +page.server.ts
│               │
│               └── imports/              # [ALL NEW]
│                   ├── +page.svelte
│                   ├── +page.server.ts
│                   └── [id]/
│                       └── +page.svelte
```

---

## Environment Variables

```bash
# ═══════════════════════════════════════════════════════════════════
# PRODUCTION READY
# ═══════════════════════════════════════════════════════════════════

# Google (Calendar, Gmail, Drive, Meet)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URI=http://localhost:8001/api/integrations/google/callback

# Slack
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
SLACK_REDIRECT_URI=http://localhost:8001/api/integrations/slack/callback

# Notion
NOTION_CLIENT_ID=
NOTION_CLIENT_SECRET=
NOTION_REDIRECT_URI=http://localhost:8001/api/integrations/notion/callback

# ═══════════════════════════════════════════════════════════════════
# HIGH PRIORITY (Phase 4-6)
# ═══════════════════════════════════════════════════════════════════

# ClickUp
CLICKUP_CLIENT_ID=
CLICKUP_CLIENT_SECRET=
CLICKUP_REDIRECT_URI=

# Asana
ASANA_CLIENT_ID=
ASANA_CLIENT_SECRET=
ASANA_REDIRECT_URI=

# Linear
LINEAR_CLIENT_ID=
LINEAR_CLIENT_SECRET=
LINEAR_REDIRECT_URI=

# HubSpot
HUBSPOT_CLIENT_ID=
HUBSPOT_CLIENT_SECRET=
HUBSPOT_REDIRECT_URI=

# Zoom
ZOOM_CLIENT_ID=
ZOOM_CLIENT_SECRET=
ZOOM_REDIRECT_URI=

# ═══════════════════════════════════════════════════════════════════
# AI/ENRICHMENT SERVICES
# ═══════════════════════════════════════════════════════════════════

# OpenAI (for embeddings)
OPENAI_API_KEY=

# ═══════════════════════════════════════════════════════════════════
# FUTURE INTEGRATIONS
# ═══════════════════════════════════════════════════════════════════

# Microsoft (Outlook, Teams, OneDrive)
MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
MICROSOFT_REDIRECT_URI=

# Atlassian (Jira, Trello, Confluence)
ATLASSIAN_CLIENT_ID=
ATLASSIAN_CLIENT_SECRET=
ATLASSIAN_REDIRECT_URI=

# Monday.com
MONDAY_CLIENT_ID=
MONDAY_CLIENT_SECRET=
MONDAY_REDIRECT_URI=

# Salesforce
SALESFORCE_CLIENT_ID=
SALESFORCE_CLIENT_SECRET=
SALESFORCE_REDIRECT_URI=

# Pipedrive
PIPEDRIVE_CLIENT_ID=
PIPEDRIVE_CLIENT_SECRET=
PIPEDRIVE_REDIRECT_URI=

# GoHighLevel
GHL_CLIENT_ID=
GHL_CLIENT_SECRET=
GHL_REDIRECT_URI=

# GitHub
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITHUB_REDIRECT_URI=

# GitLab
GITLAB_CLIENT_ID=
GITLAB_CLIENT_SECRET=
GITLAB_REDIRECT_URI=

# Discord
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=
DISCORD_REDIRECT_URI=

# Dropbox
DROPBOX_CLIENT_ID=
DROPBOX_CLIENT_SECRET=
DROPBOX_REDIRECT_URI=

# Stripe
STRIPE_CLIENT_ID=
STRIPE_CLIENT_SECRET=
STRIPE_REDIRECT_URI=

# Calendly
CALENDLY_CLIENT_ID=
CALENDLY_CLIENT_SECRET=
CALENDLY_REDIRECT_URI=

# Airtable
AIRTABLE_CLIENT_ID=
AIRTABLE_CLIENT_SECRET=
AIRTABLE_REDIRECT_URI=

# Figma
FIGMA_CLIENT_ID=
FIGMA_CLIENT_SECRET=
FIGMA_REDIRECT_URI=

# API Key Based (No OAuth)
FIREFLIES_API_KEY=
FATHOM_API_KEY=
```

---

## Key Strengths

1. **Consistent OAuth Pattern** - All 3 existing integrations follow identical pattern, easy to replicate
2. **MCP Tools for AI** - Integrations immediately accessible to AI agents via chat
3. **Token Management** - Secure storage with auto-refresh capability
4. **Comprehensive Types** - Frontend already has types for 30+ integrations
5. **Modular Architecture** - Each provider is self-contained

## Priority Actions

1. **Build Integration Settings UI** - Users currently cannot see or manage integrations
2. **Create Provider Registry** - Foundation for all future integrations
3. **Add File Import System** - Enable ChatGPT/Claude data import
4. **Add ClickUp/Asana** - Most requested task integrations
5. **Add HubSpot** - CRM integration critical for client management
6. **Implement Webhooks** - Real-time sync requires event-driven architecture

---

## Related Documentation

- [INTEGRATION_IMPLEMENTATION_PLAN.md](./INTEGRATION_IMPLEMENTATION_PLAN.md) - Complete implementation roadmap
- [INTEGRATIONS_MASTER_LIST.md](./INTEGRATIONS_MASTER_LIST.md) - Provider requirements matrix
- [FILE_IMPORT_ARCHITECTURE.md](./FILE_IMPORT_ARCHITECTURE.md) - File import system design
