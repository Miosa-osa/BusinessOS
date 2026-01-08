# Backend Architecture

**Business OS Backend — Go + Gin + SQLC**

> Last Updated: December 18, 2025

---

## Overview

The Business OS backend is a high-performance Go API server built with Gin Gonic, using SQLC for type-safe database queries. It provides 300+ REST API endpoints for the frontend, integrates multiple AI providers, and supports real-time streaming responses.

The backend is located at `/desktop/backend-go/` and can run:
- **Standalone**: For web deployment (Cloud Run)
- **Embedded**: Inside the Electron desktop app

---

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.21+ | Core language |
| **Gin Gonic** | v1.9+ | HTTP framework |
| **pgx/v5** | Latest | PostgreSQL driver |
| **SQLC** | v2 | Type-safe SQL code generation |
| **Viper** | Latest | Configuration management |
| **Better Auth** | - | Authentication (cookie-based, frontend-driven) |

---

## Directory Structure

```
desktop/backend-go/
├── cmd/
│   └── server/
│       └── main.go                 # Server entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration (Viper)
│   ├── database/
│   │   ├── postgres.go             # Connection pool
│   │   ├── schema.sql              # Database schema (622 lines)
│   │   ├── queries/                # SQLC query definitions (16 files)
│   │   │   ├── conversations.sql
│   │   │   ├── contexts.sql
│   │   │   ├── projects.sql
│   │   │   ├── clients.sql
│   │   │   ├── team_members.sql
│   │   │   ├── nodes.sql
│   │   │   ├── tasks.sql
│   │   │   ├── artifacts.sql
│   │   │   ├── calendar.sql
│   │   │   ├── daily_logs.sql
│   │   │   ├── user_settings.sql
│   │   │   ├── usage.sql
│   │   │   ├── voice_notes.sql
│   │   │   ├── user_commands.sql
│   │   │   └── google_oauth.sql
│   │   └── sqlc/                   # Generated Go code
│   │       ├── db.go               # Database interface
│   │       ├── models.go           # Generated models
│   │       └── *.sql.go            # Query implementations
│   ├── handlers/                   # HTTP handlers (22 files)
│   │   ├── handlers.go             # Route registration
│   │   ├── response.go             # Response helpers
│   │   ├── chat.go                 # Chat & AI streaming
│   │   ├── contexts.go             # Documents
│   │   ├── projects.go             # Projects
│   │   ├── clients.go              # CRM
│   │   ├── deals.go                # Deal pipeline
│   │   ├── team.go                 # Team management
│   │   ├── nodes.go                # Business nodes
│   │   ├── dashboard.go            # Dashboard & tasks
│   │   ├── daily_logs.go           # Daily log
│   │   ├── artifacts.go            # AI artifacts
│   │   ├── settings.go             # User settings
│   │   ├── profile.go              # User profile
│   │   ├── calendar.go             # Calendar events
│   │   ├── google_oauth.go         # Google OAuth
│   │   ├── voice_notes.go          # Voice transcription
│   │   ├── transcription.go        # Whisper transcription
│   │   ├── usage.go                # Usage analytics
│   │   ├── ai_config.go            # AI configuration
│   │   ├── commands.go             # Slash commands
│   │   └── mcp.go                  # MCP tools
│   ├── middleware/
│   │   ├── auth.go                 # Better Auth validation
│   │   └── cors.go                 # CORS configuration
│   ├── services/                   # Business logic & integrations
│   │   ├── llm.go                  # LLM interface
│   │   ├── ollama.go               # Local Ollama
│   │   ├── ollama_cloud.go         # Ollama Cloud
│   │   ├── anthropic.go            # Claude/Anthropic
│   │   ├── groq.go                 # Groq
│   │   ├── whisper.go              # Whisper transcription
│   │   ├── google_calendar.go      # Google Calendar
│   │   ├── mcp.go                  # MCP client
│   │   └── usage.go                # Usage tracking
│   ├── agents/
│   │   └── agents.go               # Multi-agent system
│   ├── prompts/
│   │   └── prompts.go              # AI system prompts
│   └── tools/
│       └── artifacts.go            # Artifact tools
├── sqlc.yaml                       # SQLC configuration
├── go.mod
├── go.sum
└── .env                            # Environment config
```

---

## Configuration

### Environment Variables

```env
# Server
SERVER_PORT=8000

# Database
DATABASE_URL=postgres://user:password@localhost:5432/business_os

# AI Provider (ollama_local, ollama_cloud, anthropic, groq)
AI_PROVIDER=ollama_local

# Ollama Local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=llama3.2:3b

# Ollama Cloud
OLLAMA_CLOUD_API_KEY=your-key
OLLAMA_CLOUD_MODEL=llama3.2

# Anthropic
ANTHROPIC_API_KEY=your-key
ANTHROPIC_MODEL=claude-sonnet-4-20250514

# Groq
GROQ_API_KEY=your-key
GROQ_MODEL=llama-3.3-70b-versatile

# Google OAuth (Calendar)
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8000/api/integrations/google/callback

# CORS
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# Better Auth
SECRET_KEY=your-secret-key
```

### Config Struct

```go
type Config struct {
    DatabaseURL    string `mapstructure:"DATABASE_URL"`
    ServerPort     string `mapstructure:"SERVER_PORT"`

    // AI Providers
    AIProvider        string `mapstructure:"AI_PROVIDER"`
    OllamaLocalURL    string `mapstructure:"OLLAMA_LOCAL_URL"`
    OllamaCloudAPIKey string `mapstructure:"OLLAMA_CLOUD_API_KEY"`
    AnthropicAPIKey   string `mapstructure:"ANTHROPIC_API_KEY"`
    GroqAPIKey        string `mapstructure:"GROQ_API_KEY"`
    DefaultModel      string `mapstructure:"DEFAULT_MODEL"`

    // Google OAuth
    GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
    GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
    GoogleRedirectURI  string `mapstructure:"GOOGLE_REDIRECT_URI"`

    // CORS
    AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`
}
```

---

## Database Architecture

### SQLC Code Generation

SQLC generates type-safe Go code from SQL queries.

**sqlc.yaml:**
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/database/queries"
    schema: "internal/database/schema.sql"
    gen:
      go:
        package: "sqlc"
        out: "internal/database/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        json_tags_case_style: "snake"
        emit_empty_slices: true
        emit_pointers_for_null_types: true
```

**Regenerate after schema changes:**
```bash
sqlc generate
```

### Database Schema

27 tables organized into domains:

#### Core Tables

| Table | Purpose |
|-------|---------|
| `contexts` | Documents with blocks, properties, sharing |
| `conversations` | Chat conversations |
| `messages` | Chat messages with metadata |
| `projects` | Project management |
| `tasks` | Task management |
| `artifacts` | AI-generated content |
| `nodes` | Business structure hierarchy |

#### CRM Tables

| Table | Purpose |
|-------|---------|
| `clients` | Client profiles |
| `client_contacts` | Client contacts |
| `client_interactions` | Interaction history |
| `client_deals` | Sales pipeline |

#### Team & User Tables

| Table | Purpose |
|-------|---------|
| `team_members` | Team directory |
| `team_member_activities` | Activity tracking |
| `user_settings` | User preferences |
| `user_commands` | Custom slash commands |

#### Calendar & Logs

| Table | Purpose |
|-------|---------|
| `calendar_events` | Calendar with Google sync |
| `google_oauth_tokens` | OAuth tokens |
| `daily_logs` | Daily journal |
| `voice_notes` | Voice transcriptions |

#### Analytics Tables

| Table | Purpose |
|-------|---------|
| `ai_usage_logs` | AI request tracking |
| `mcp_usage_logs` | MCP tool tracking |
| `usage_daily_summary` | Aggregated stats |
| `system_event_logs` | Activity logs |

### Custom Types (Enums)

```sql
CREATE TYPE messagerole AS ENUM ('user', 'assistant', 'system');
CREATE TYPE artifacttype AS ENUM ('CODE', 'DOCUMENT', 'MARKDOWN', ...);
CREATE TYPE contexttype AS ENUM ('PERSON', 'BUSINESS', 'PROJECT', 'CUSTOM', 'document');
CREATE TYPE projectstatus AS ENUM ('ACTIVE', 'PAUSED', 'COMPLETED', 'ARCHIVED');
CREATE TYPE projectpriority AS ENUM ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW');
CREATE TYPE nodetype AS ENUM ('BUSINESS', 'PROJECT', 'LEARNING', 'OPERATIONAL');
CREATE TYPE nodehealth AS ENUM ('HEALTHY', 'NEEDS_ATTENTION', 'CRITICAL', 'NOT_STARTED');
CREATE TYPE taskstatus AS ENUM ('todo', 'in_progress', 'done', 'cancelled');
CREATE TYPE clientstatus AS ENUM ('lead', 'prospect', 'active', 'inactive', 'churned');
CREATE TYPE dealstage AS ENUM ('qualification', 'proposal', 'negotiation', 'closed_won', 'closed_lost');
CREATE TYPE meetingtype AS ENUM ('team', 'sales', 'onboarding', 'kickoff', ...);
```

---

## API Reference

### Route Registration

All routes are registered in `handlers/handlers.go`:

```go
func (h *Handlers) RegisterRoutes(api *gin.RouterGroup) {
    auth := middleware.AuthMiddleware(h.pool)

    // Chat routes
    chat := api.Group("/chat")
    chat.Use(auth)
    {
        chat.GET("/conversations", h.ListConversations)
        chat.POST("/message", h.SendMessage)
        // ...
    }

    // 300+ more routes...
}
```

### API Domains

#### Chat & AI (`/api/chat`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/conversations` | `ListConversations` | List all conversations |
| POST | `/conversations` | `CreateConversation` | Create new conversation |
| GET | `/conversations/:id` | `GetConversation` | Get conversation with messages |
| PUT | `/conversations/:id` | `UpdateConversation` | Update conversation |
| DELETE | `/conversations/:id` | `DeleteConversation` | Delete conversation |
| POST | `/message` | `SendMessage` | Send message (streaming) |
| GET | `/search` | `SearchConversations` | Full-text search |
| POST | `/ai/document` | `DocumentAI` | Generate document |
| POST | `/ai/analyze` | `AnalyzeContent` | Analyze content |
| POST | `/ai/extract-tasks` | `ExtractTasks` | Extract tasks from text |
| POST | `/ai/plan` | `CreatePlan` | Generate plan |

#### Contexts/Documents (`/api/contexts`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/` | `ListContexts` | List documents with filters |
| POST | `/` | `CreateContext` | Create document |
| GET | `/:id` | `GetContext` | Get document |
| PUT | `/:id` | `UpdateContext` | Update document |
| PATCH | `/:id/blocks` | `UpdateContextBlocks` | Update blocks only |
| POST | `/:id/share` | `ShareContext` | Enable public sharing |
| DELETE | `/:id/share` | `UnshareContext` | Disable sharing |
| GET | `/public/:shareId` | `GetPublicContext` | Get shared document |
| POST | `/:id/duplicate` | `DuplicateContext` | Duplicate document |
| PATCH | `/:id/archive` | `ArchiveContext` | Archive |
| PATCH | `/:id/unarchive` | `UnarchiveContext` | Unarchive |
| DELETE | `/:id` | `DeleteContext` | Delete document |
| POST | `/aggregate` | `AggregateContext` | Aggregate multiple contexts |

#### Projects (`/api/projects`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/` | `ListProjects` | List with status filter |
| POST | `/` | `CreateProject` | Create project |
| GET | `/:id` | `GetProject` | Get project with notes |
| PUT | `/:id` | `UpdateProject` | Update project |
| DELETE | `/:id` | `DeleteProject` | Delete project |
| POST | `/:id/notes` | `AddProjectNote` | Add note |

#### Clients CRM (`/api/clients`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/` | `ListClients` | List with filters |
| POST | `/` | `CreateClient` | Create client |
| GET | `/:id` | `GetClient` | Get with contacts/deals |
| PUT | `/:id` | `UpdateClient` | Update client |
| PATCH | `/:id/status` | `UpdateClientStatus` | Update status |
| DELETE | `/:id` | `DeleteClient` | Delete client |
| GET | `/:id/contacts` | `ListClientContacts` | List contacts |
| POST | `/:id/contacts` | `CreateClientContact` | Add contact |
| PUT | `/:id/contacts/:contactId` | `UpdateClientContact` | Update contact |
| DELETE | `/:id/contacts/:contactId` | `DeleteClientContact` | Delete contact |
| GET | `/:id/interactions` | `ListClientInteractions` | List interactions |
| POST | `/:id/interactions` | `CreateClientInteraction` | Log interaction |
| GET | `/:id/deals` | `ListClientDeals` | List deals |
| POST | `/:id/deals` | `CreateClientDeal` | Create deal |
| PUT | `/:id/deals/:dealId` | `UpdateClientDeal` | Update deal |

#### Deals Pipeline (`/api/deals`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/` | `ListDeals` | All deals with stage filter |
| PATCH | `/:id/stage` | `UpdateDealStage` | Move deal stage |

#### Dashboard (`/api/dashboard`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/summary` | `GetDashboardSummary` | Full dashboard data |
| GET | `/focus` | `ListFocusItems` | Today's focus items |
| POST | `/focus` | `CreateFocusItem` | Add focus item |
| PUT | `/focus/:id` | `UpdateFocusItem` | Update focus item |
| DELETE | `/focus/:id` | `DeleteFocusItem` | Delete focus item |
| GET | `/tasks` | `ListTasks` | Tasks with filters |
| POST | `/tasks` | `CreateTask` | Create task |
| PUT | `/tasks/:id` | `UpdateTask` | Update task |
| POST | `/tasks/:id/toggle` | `ToggleTask` | Toggle completion |
| DELETE | `/tasks/:id` | `DeleteTask` | Delete task |

#### Team (`/api/team`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/` | `ListTeamMembers` | List with status filter |
| POST | `/` | `CreateTeamMember` | Add member |
| GET | `/:id` | `GetTeamMember` | Get member details |
| PUT | `/:id` | `UpdateTeamMember` | Update member |
| PATCH | `/:id/status` | `UpdateTeamMemberStatus` | Update status |
| PATCH | `/:id/capacity` | `UpdateTeamMemberCapacity` | Update capacity |
| POST | `/:id/activity` | `AddTeamMemberActivity` | Log activity |
| DELETE | `/:id` | `DeleteTeamMember` | Remove member |

#### Nodes (`/api/nodes`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/` | `ListNodes` | List all nodes |
| GET | `/tree` | `GetNodeTree` | Hierarchical tree |
| GET | `/active` | `GetActiveNode` | Currently active node |
| POST | `/` | `CreateNode` | Create node |
| GET | `/:id` | `GetNode` | Get node details |
| PATCH | `/:id` | `UpdateNode` | Update node |
| POST | `/:id/activate` | `ActivateNode` | Set as active |
| POST | `/:id/deactivate` | `DeactivateNode` | Deactivate |
| DELETE | `/:id` | `DeleteNode` | Delete node |
| GET | `/:id/children` | `GetNodeChildren` | Get child nodes |
| POST | `/:id/reorder` | `ReorderNodes` | Reorder children |

#### Calendar (`/api/calendar`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/events` | `ListEvents` | Events with date range |
| GET | `/events/:id` | `GetEvent` | Get single event |
| POST | `/events` | `CreateEvent` | Create event |
| PUT | `/events/:id` | `UpdateEvent` | Update event |
| DELETE | `/events/:id` | `DeleteEvent` | Delete event |
| POST | `/sync` | `SyncCalendar` | Sync with Google |
| GET | `/today` | `GetTodayEvents` | Today's events |
| GET | `/upcoming` | `GetUpcomingEvents` | Upcoming events |

#### AI Configuration (`/api/ai`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/providers` | `GetLLMProviders` | List AI providers |
| GET | `/models` | `GetAllModels` | All available models |
| GET | `/models/local` | `GetLocalModels` | Local Ollama models |
| POST | `/models/pull` | `PullModel` | Pull Ollama model |
| GET | `/system` | `GetSystemInfo` | System info (RAM, GPU) |
| POST | `/api-key` | `SaveAPIKey` | Save provider API key |
| PUT | `/provider` | `UpdateAIProvider` | Change active provider |
| GET | `/agents` | `GetAgentPrompts` | List agents |
| GET | `/agents/:id` | `GetAgentPrompt` | Get agent prompt |
| GET | `/commands` | `ListCommands` | List slash commands |
| POST | `/commands` | `CreateUserCommand` | Create custom command |
| PUT | `/commands/:id` | `UpdateUserCommand` | Update command |
| DELETE | `/commands/:id` | `DeleteUserCommand` | Delete command |

#### Usage Analytics (`/api/usage`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/summary` | `GetUsageSummary` | Usage totals |
| GET | `/providers` | `GetUsageByProvider` | By provider |
| GET | `/models` | `GetUsageByModel` | By model |
| GET | `/agents` | `GetUsageByAgent` | By agent |
| GET | `/trend` | `GetUsageTrend` | Daily trend |
| GET | `/recent` | `GetRecentAIUsage` | Recent requests |
| GET | `/mcp` | `GetMCPUsage` | MCP tool usage |

#### Voice Notes (`/api/voice-notes`)

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/` | `ListVoiceNotes` | List recordings |
| POST | `/` | `UploadVoiceNote` | Upload & transcribe |
| GET | `/stats` | `GetVoiceNoteStats` | Transcription stats |
| GET | `/:id` | `GetVoiceNote` | Get audio file |
| DELETE | `/:id` | `DeleteVoiceNote` | Delete recording |
| POST | `/:id/retranscribe` | `RetranscribeVoiceNote` | Re-transcribe |

#### Other Routes

| Route Group | Endpoints |
|-------------|-----------|
| `/api/daily/logs` | Daily log CRUD (6 routes) |
| `/api/settings` | User settings (3 routes) |
| `/api/artifacts` | Artifacts CRUD (6 routes) |
| `/api/profile` | Profile & photos (5 routes) |
| `/api/mcp` | MCP tools (3 routes) |
| `/api/integrations/google` | Google OAuth (4 routes) |
| `/api/transcribe` | Direct transcription (2 routes) |

---

## AI System

### LLM Service Interface

```go
type LLMService interface {
    StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string) (<-chan string, <-chan error)
    StreamChatWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) *StreamResult
    ChatComplete(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, error)
    ChatCompleteWithUsage(ctx context.Context, messages []ChatMessage, systemPrompt string) (string, *TokenUsage, error)
    HealthCheck(ctx context.Context) bool
    GetModel() string
    GetProvider() string
    SetOptions(opts LLMOptions)
    GetOptions() LLMOptions
}
```

### Provider Factory

```go
func NewLLMService(cfg *config.Config, model string) LLMService {
    switch cfg.GetActiveProvider() {
    case "ollama_cloud":
        return NewOllamaCloudService(cfg, model)
    case "anthropic":
        return NewAnthropicService(cfg, model)
    case "groq":
        return NewGroqService(cfg, model)
    default:
        return NewOllamaService(cfg, model)
    }
}
```

### Supported Providers

| Provider | Service File | Models |
|----------|--------------|--------|
| **Ollama Local** | `ollama.go` | Any local model |
| **Ollama Cloud** | `ollama_cloud.go` | Qwen3 235b/480b |
| **Anthropic** | `anthropic.go` | Claude Sonnet 4, Claude Opus 4 |
| **Groq** | `groq.go` | Llama 3.3 70B, Mixtral |

### Token Usage Tracking

```go
type TokenUsage struct {
    InputTokens  int    `json:"input_tokens"`
    OutputTokens int    `json:"output_tokens"`
    TotalTokens  int    `json:"total_tokens"`
    Model        string `json:"model"`
    Provider     string `json:"provider"`
}
```

---

## Multi-Agent System

### Agent Types

```go
const (
    AgentTypeOrchestrator AgentType = "orchestrator"
    AgentTypeDocument     AgentType = "document"
    AgentTypeAnalysis     AgentType = "analysis"
    AgentTypePlanning     AgentType = "planning"
)
```

### Agent Interface

```go
type Agent interface {
    Name() string
    Description() string
    SystemPrompt() string
    Run(ctx context.Context, messages []ChatMessage) (<-chan string, <-chan error)
    SetOptions(opts LLMOptions)
}
```

### Agent Implementations

| Agent | Purpose | Prompt Source |
|-------|---------|---------------|
| `OrchestratorAgent` | Coordinates, delegates to sub-agents | `prompts.GetPrompt("orchestrator")` |
| `DocumentAgent` | Creates business documents | `prompts.GetPrompt("document")` |
| `AnalysisAgent` | Data analysis, insights | `prompts.GetPrompt("analyst")` |
| `PlanningAgent` | Planning, prioritization | `prompts.GetPrompt("planner")` |

### Focus Mode → Agent Mapping

```go
func GetAgentForFocusMode(focusMode string) AgentType {
    switch focusMode {
    case "research", "analyze":
        return AgentTypeAnalysis
    case "write":
        return AgentTypeDocument
    case "build":
        return AgentTypePlanning
    case "general":
        return AgentTypeOrchestrator
    default:
        return AgentTypeOrchestrator
    }
}
```

### Orchestrator Delegation

The orchestrator can delegate to sub-agents via `[DELEGATE:AgentName]` tags:

```go
func (o *OrchestratorAgent) Run(ctx context.Context, messages []ChatMessage) (<-chan string, <-chan error) {
    // 1. Get orchestrator's decision
    decision, _ := llm.ChatComplete(ctx, messages, o.systemPrompt)

    // 2. Check for delegation
    agentName := parseDelegation(decision)  // Parses [DELEGATE:DocumentAgent]

    if agentName != "" {
        // Delegate to sub-agent
        subAgent := o.getSubAgent(agentName)
        return subAgent.Run(ctx, messages)
    }

    // No delegation, stream orchestrator response
    return o.BaseAgent.Run(ctx, messages)
}
```

---

## Authentication

### Better Auth Middleware

Authentication is handled by Better Auth (frontend). The backend validates session cookies:

```go
// middleware/auth.go
func AuthMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Read session cookie set by Better Auth
        sessionToken, err := c.Cookie("better-auth.session_token")
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }

        // Validate session in database
        session, err := validateSession(pool, sessionToken)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid session"})
            return
        }

        // Set user ID in context
        c.Set("user_id", session.UserID)
        c.Next()
    }
}
```

### Getting User ID in Handlers

```go
func (h *Handlers) ListProjects(c *gin.Context) {
    userID := c.GetString("user_id")
    // ... use userID for queries
}
```

---

## Streaming Responses

### SSE Pattern

```go
func (h *Handlers) SendMessage(c *gin.Context) {
    // Set SSE headers
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // Create agent
    agent := agents.GetAgent(agentType, h.pool, h.cfg, userID, &convID, model)

    // Stream response
    chunks, errs := agent.Run(c.Request.Context(), messages)

    for {
        select {
        case chunk, ok := <-chunks:
            if !ok {
                return
            }
            c.Writer.WriteString(chunk)
            c.Writer.Flush()
        case err := <-errs:
            if err != nil {
                log.Printf("Stream error: %v", err)
            }
            return
        case <-c.Request.Context().Done():
            return
        }
    }
}
```

---

## SQLC Query Patterns

### Query Definition

```sql
-- queries/projects.sql

-- name: GetProject :one
SELECT * FROM projects
WHERE id = $1 AND user_id = $2;

-- name: ListProjects :many
SELECT * FROM projects
WHERE user_id = $1
  AND ($2::text IS NULL OR status = $2::projectstatus)
ORDER BY updated_at DESC;

-- name: CreateProject :one
INSERT INTO projects (user_id, name, description, status, priority)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
```

### Generated Go Code

```go
// sqlc/projects.sql.go (auto-generated)

func (q *Queries) GetProject(ctx context.Context, arg GetProjectParams) (Project, error)
func (q *Queries) ListProjects(ctx context.Context, arg ListProjectsParams) ([]Project, error)
func (q *Queries) CreateProject(ctx context.Context, arg CreateProjectParams) (Project, error)
```

### Handler Usage

```go
func (h *Handlers) GetProject(c *gin.Context) {
    userID := c.GetString("user_id")
    projectID := uuid.MustParse(c.Param("id"))

    queries := sqlc.New(h.pool)

    project, err := queries.GetProject(c.Request.Context(), sqlc.GetProjectParams{
        ID:     projectID,
        UserID: userID,
    })

    if err != nil {
        c.JSON(404, gin.H{"error": "Project not found"})
        return
    }

    c.JSON(200, project)
}
```

---

## Error Handling

### Response Helpers

```go
// handlers/response.go

func respondError(c *gin.Context, code int, message string) {
    c.JSON(code, gin.H{"error": message})
}

func respondSuccess(c *gin.Context, data interface{}) {
    c.JSON(200, data)
}

func respondCreated(c *gin.Context, data interface{}) {
    c.JSON(201, data)
}
```

### Standard Error Responses

| Code | Meaning | Example |
|------|---------|---------|
| 400 | Bad Request | Invalid JSON, missing field |
| 401 | Unauthorized | No/invalid session |
| 403 | Forbidden | Not owner of resource |
| 404 | Not Found | Resource doesn't exist |
| 500 | Server Error | Database error |

---

## Running the Server

### Development

```bash
cd desktop/backend-go

# Install dependencies
go mod download

# Generate SQLC code (after schema changes)
sqlc generate

# Run server
go run cmd/server/main.go
```

### Building

```bash
cd desktop/backend-go
go build -o server cmd/server/main.go
./server
```

### With Environment

```bash
cd desktop/backend-go

# Create .env file
cp .env.production.example .env
# Edit .env with your configuration

# Run
go run cmd/server/main.go
```

### For Desktop App

```bash
# Build backend binary for embedding
cd desktop/backend-go
go build -o server cmd/server/main.go

# The desktop app will spawn this binary as a subprocess
```

---

## Database Migrations

Schema changes are managed manually:

1. Edit `desktop/backend-go/internal/database/schema.sql`
2. Apply to database: `psql business_os < desktop/backend-go/internal/database/schema.sql`
3. Regenerate SQLC: `cd desktop/backend-go && sqlc generate`
4. Update Go handlers if needed

---

## File Uploads

### Profile Photos

```go
func (h *Handlers) UploadProfilePhoto(c *gin.Context) {
    file, _ := c.FormFile("file")

    // Save to ./uploads/profiles/
    filename := uuid.New().String() + filepath.Ext(file.Filename)
    path := filepath.Join("uploads", "profiles", filename)
    c.SaveUploadedFile(file, path)

    c.JSON(200, gin.H{"url": "/uploads/profiles/" + filename})
}
```

### Voice Notes

```go
func (h *VoiceNotesHandler) UploadVoiceNote(c *gin.Context) {
    file, _ := c.FormFile("audio")

    // Save audio file
    filename := uuid.New().String() + ".webm"
    path := filepath.Join("uploads", "voice", filename)
    c.SaveUploadedFile(file, path)

    // Transcribe with Whisper
    transcript := h.whisper.Transcribe(path)

    // Save to database
    // ...
}
```

---

## Key Files Reference

All paths relative to `desktop/backend-go/internal/`:

| File | Lines | Purpose |
|------|-------|---------|
| `handlers/handlers.go` | 313 | Route registration |
| `handlers/chat.go` | 500+ | Chat & AI streaming |
| `handlers/contexts.go` | 600+ | Document management |
| `handlers/clients.go` | 500+ | CRM handlers |
| `database/schema.sql` | 622 | Database schema |
| `services/llm.go` | 95 | LLM interface |
| `agents/agents.go` | 384 | Multi-agent system |
| `config/config.go` | 185 | Configuration |

---

## Deployment

### Web Deployment (Cloud Run)

See [DEPLOYMENT.md](./DEPLOYMENT.md) for production deployment instructions.

### Desktop App Embedding

The backend is embedded in the Electron desktop app:

1. Build the Go binary: `go build -o server cmd/server/main.go`
2. The Electron main process spawns the server as a subprocess
3. Frontend connects to `http://localhost:8000/api`

---

*This documentation reflects the backend architecture as of December 2025.*
