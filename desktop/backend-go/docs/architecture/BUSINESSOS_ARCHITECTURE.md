# BusinessOS Complete Architecture Reference

> **Last Updated:** January 7, 2026
> **Version:** 2.0.0
> **Status:** Living Document - Extreme Detail Edition

---

## Table of Contents

1. [Platform Overview](#1-platform-overview)
2. [The Two-Layer Architecture](#2-the-two-layer-architecture)
3. [Module Deep Dive](#3-module-deep-dive)
4. [Complete Data Model](#4-complete-data-model)
5. [Data Flow & Relationships](#5-data-flow--relationships)
6. [Frontend Route Structure](#6-frontend-route-structure)
7. [Backend API Structure](#7-backend-api-structure)
8. [Integration Architecture](#8-integration-architecture)
9. [What's Missing / Gaps](#9-whats-missing--gaps)
10. [Implementation Roadmap](#10-implementation-roadmap)

---

## 1. Platform Overview

### 1.1 What is BusinessOS?

BusinessOS is a **unified business operating system** that combines:
- **Knowledge Management** (like Notion)
- **Project Management** (like Linear/Asana)
- **CRM** (like Salesforce)
- **Communication Hub** (Email/Calendar/Slack)
- **AI Assistant** (like ChatGPT with context)

All unified through a **Node-based Operating System** where everything connects.

### 1.2 Core Philosophy

```
"Everything is connected. Your client call connects to the project,
which connects to the tasks, which connect to the knowledge base,
which connects to the AI conversations."
```

### 1.3 Tech Stack

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND                                 │
├─────────────────────────────────────────────────────────────────┤
│  Svelte/SvelteKit + TypeScript + Tailwind CSS                   │
│  UI: Custom + Bits UI components                                │
│  State: Svelte stores                                           │
│  Auth: Better Auth                                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ REST API + SSE
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          BACKEND                                 │
├─────────────────────────────────────────────────────────────────┤
│  Go + Chi Router                                                │
│  Database: PostgreSQL (sqlc for queries)                        │
│  Cache: Redis                                                   │
│  AI: Claude API (Anthropic)                                     │
│  MCP: 20+ tool servers                                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       INTEGRATIONS                               │
├─────────────────────────────────────────────────────────────────┤
│  Google (Calendar, Gmail, Drive, Tasks, Contacts)               │
│  Microsoft (Outlook, Calendar, Teams, OneDrive)                 │
│  Notion, Slack, Linear, HubSpot, Airtable, ClickUp, Fathom     │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. The Two-Layer Architecture

### 2.1 The Core Insight

BusinessOS has **two distinct layers**:

```
┌─────────────────────────────────────────────────────────────────┐
│                 LAYER 1: OPERATING SYSTEM                        │
│                        (The Graph)                               │
│                                                                  │
│     NODES form the STRUCTURE of your business/life              │
│     They represent WHAT EXISTS and HOW IT CONNECTS              │
│                                                                  │
│     Examples:                                                    │
│     - "Lunivate LLC" (Entity Node)                              │
│     - "Q1 Growth" (Initiative Node)                             │
│     - "Frontend Development" (Operations Node)                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ linked to
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  LAYER 2: MODULES                                │
│                     (The Tools)                                  │
│                                                                  │
│     MODULES are TOOLS for doing work within the OS              │
│     They handle EXECUTION, COMMUNICATION, KNOWLEDGE             │
│                                                                  │
│     Examples:                                                    │
│     - Knowledge (Pages/Documents)                               │
│     - Projects (Tasks/Boards)                                   │
│     - Clients (CRM)                                             │
│     - Communication (Email/Calendar)                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Node Types (Operating System Layer)

| Type | Code | Description | Contains |
|------|------|-------------|----------|
| **Business/Venture** | `BUSINESS` | Company or organization | Departments, Products |
| **Project/Initiative** | `PROJECT` | Bounded initiative | Milestones, linked Projects |
| **Learning** | `LEARNING` | Knowledge acquisition | Resources, Notes |
| **Operational** | `OPERATIONAL` | Ongoing process | SOPs, Workflows |

### 2.3 Node Anatomy

Every Node has these universal sections:

```
NODE: [Name]
│
├── 📌 IDENTITY
│   ├── Name, Type, Purpose
│   ├── Owner, Health Status
│   └── Unique ID
│
├── 🔗 RELATIONSHIPS
│   ├── Parent Node (hierarchical)
│   ├── Child Nodes (hierarchical)
│   ├── Linked Projects (via node_projects)
│   ├── Linked Pages/Contexts (via node_contexts)
│   └── Linked Conversations (via node_conversations)
│
├── 📊 STATE
│   ├── Health: healthy | needs_attention | critical | not_started
│   ├── Current Status (text)
│   └── Is Active (boolean)
│
├── 🎯 FOCUS
│   ├── This Week Focus (JSONB array)
│   ├── Decision Queue (JSONB array)
│   └── Delegation Ready (JSONB array)
│
└── 📈 TRACKING
    └── Node Metrics (separate table)
```

---

## 3. Module Deep Dive

### 3.1 Module Overview Map

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            BUSINESSOS MODULES                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│  │  Dashboard   │   │    Chat      │   │  Knowledge   │   │    Nodes     │ │
│  │  (overview)  │   │   (AI)       │   │  (pages)     │   │   (OS)       │ │
│  └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘ │
│                                                                              │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│  │   Projects   │   │    Tasks     │   │   Clients    │   │    Team      │ │
│  │  (tracking)  │   │  (work)      │   │   (CRM)      │   │  (people)    │ │
│  └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘ │
│                                                                              │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│  │   Daily Log  │   │ Voice Notes  │   │Communication │   │ Integrations │ │
│  │  (journal)   │   │ (transcribe) │   │ (email/cal)  │   │  (3rd party) │ │
│  └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘ │
│                                                                              │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│  │   Settings   │   │   Profile    │   │   Terminal   │   │    Usage     │ │
│  │  (config)    │   │  (user)      │   │  (dev)       │   │  (analytics) │ │
│  └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘ │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Dashboard Module

**Route:** `/dashboard`
**Purpose:** Central overview and focus management

**Primary Data:**
- `focus_items` - Today's priorities
- `tasks` (filtered) - Urgent/due tasks
- `nodes` (active) - Currently focused nodes
- `calendar_events` (today) - Today's schedule

**Data Sources:**
```
Dashboard
├── focus_items (user's daily focus)
├── tasks WHERE status='in_progress' OR due_date=TODAY
├── nodes WHERE is_active=true
├── calendar_events WHERE date=TODAY
├── ai_usage_logs (recent activity)
└── conversations (recent chats)
```

**Key Features:**
- Focus items management
- Today's agenda
- Quick actions
- Activity feed
- Node health overview

---

### 3.3 Chat Module

**Route:** `/chat`
**Purpose:** AI conversations with context awareness

**Primary Tables:**
```sql
conversations
├── id, user_id, title
├── context_id → links to contexts (optional)
└── created_at, updated_at

messages
├── id, conversation_id (FK)
├── role: USER | ASSISTANT | SYSTEM
├── content: TEXT
└── message_metadata: JSONB

artifacts (AI creations)
├── id, user_id, conversation_id, message_id
├── title, type, language, content
├── version (integer)
└── linked to: project_id, context_id
```

**Artifact Types:**
- `CODE` - Generated code
- `DOCUMENT` - Generated docs
- `MARKDOWN` - Markdown content
- `REACT` - React components
- `HTML` - HTML markup
- `SVG` - Vector graphics

**Relationships:**
- Conversation → Context (optional, 1:1)
- Conversation → Node (via node_conversations, M:N)
- Conversation → Project (via project_conversations, M:N)
- Message → Artifacts (1:N)

---

### 3.4 Knowledge Module (Pages)

**Route:** `/knowledge-v2`
**Purpose:** Document and knowledge management (like Notion)

**Primary Table: `contexts`**
```sql
contexts (should be renamed to 'pages')
├── id, user_id
├── name (title)
├── type: PERSON | BUSINESS | PROJECT | CUSTOM | document | DOCUMENT
├── content (TEXT - legacy)
├── blocks (JSONB array - Notion-style)
├── structured_data (JSONB)
├── cover_image, icon
├── parent_id → hierarchical structure
├── is_template, is_archived
├── is_public, share_id (sharing)
├── property_schema, properties (custom fields)
├── client_id → links to clients
└── word_count, last_edited_at
```

**Block Structure (Notion-inspired):**
```typescript
interface Block {
  id: string;
  type: 'paragraph' | 'heading_1' | 'heading_2' | 'heading_3' |
        'bulleted_list_item' | 'numbered_list_item' | 'to_do' |
        'code' | 'quote' | 'divider' | 'callout' | 'toggle' |
        'image' | 'embed' | 'table';
  content: RichText[];
  properties: Record<string, any>;
  children: Block[];
  created_at: string;
  updated_at: string;
}

interface RichText {
  type: 'text' | 'mention' | 'equation';
  plain_text: string;
  text?: { content: string; link?: string };
  annotations?: {
    bold?: boolean;
    italic?: boolean;
    strikethrough?: boolean;
    underline?: boolean;
    code?: boolean;
    color?: string;
  };
}
```

**Relationships:**
- Context → Parent Context (hierarchical)
- Context → Node (via node_contexts, M:N)
- Context → Client (1:1 optional)
- Context → Projects (via project_documents, M:N)

---

### 3.5 Nodes Module (Operating System)

**Route:** `/nodes`, `/nodes/[id]`
**Purpose:** The operating system structure

**Primary Table: `nodes`**
```sql
nodes
├── id, user_id
├── parent_id → hierarchical structure
├── context_id → links to a "profile" context
├── name, type, purpose
├── health: HEALTHY | NEEDS_ATTENTION | CRITICAL | NOT_STARTED
├── current_status (TEXT)
├── this_week_focus (JSONB array)
├── decision_queue (JSONB array)
├── delegation_ready (JSONB array)
├── is_active, is_archived
└── sort_order

node_metrics
├── id, node_id
├── metric_name, metric_value
└── recorded_at
```

**Junction Tables (Many-to-Many):**
```sql
node_projects     -- Nodes ↔ Projects
node_contexts     -- Nodes ↔ Contexts/Pages
node_conversations -- Nodes ↔ Conversations
```

---

### 3.6 Projects Module

**Route:** `/projects`, `/projects/[id]`
**Purpose:** Project and work management

**Primary Table: `projects`**
```sql
projects
├── id, user_id, owner_id
├── name, description
├── status: ACTIVE | PAUSED | COMPLETED | ARCHIVED
├── priority: CRITICAL | HIGH | MEDIUM | LOW
├── client_name, client_id → links to clients
├── project_type, project_metadata (JSONB)
├── start_date, due_date, completed_at
└── visibility: private | team | public

project_statuses (custom statuses per project)
├── id, project_id
├── name, color, position
├── is_done_state, is_default

project_notes
├── id, project_id
└── content, created_at

project_conversations (M:N junction)
project_members (team assignment)
project_tags + project_tag_assignments
project_documents (links to contexts)
project_templates
```

---

### 3.7 Tasks Module

**Route:** `/tasks`
**Purpose:** Task execution and tracking

**Primary Table: `tasks`**
```sql
tasks
├── id, user_id
├── title, description
├── status: todo | in_progress | done | cancelled
├── priority: critical | high | medium | low
├── due_date, start_date, completed_at
├── project_id → belongs to project
├── assignee_id → single assignee (team_member)
├── parent_task_id → subtasks
├── custom_status_id → project_statuses
└── position (for ordering)

task_assignees (M:N for multiple assignees)
├── task_id, team_member_id
├── role, assigned_at, assigned_by

task_dependencies
├── predecessor_id, successor_id
├── dependency_type: finish_to_start | start_to_start | ...
└── lag_days
```

---

### 3.8 Clients Module (CRM)

**Route:** `/clients`, `/clients/[id]`
**Purpose:** Customer relationship management

**Primary Tables:**
```sql
clients
├── id, user_id
├── name, type: company | individual
├── email, phone, website
├── industry, company_size
├── address, city, state, zip_code, country
├── status: lead | prospect | active | inactive | churned
├── source, assigned_to
├── lifetime_value
├── tags (JSONB array), custom_fields (JSONB)
├── notes
└── last_contacted_at

client_contacts
├── id, client_id
├── name, email, phone, role
├── is_primary, notes

client_interactions
├── id, client_id, contact_id
├── type: call | email | meeting | note
├── subject, description, outcome
└── occurred_at

client_deals
├── id, client_id
├── name, value, probability
├── stage: qualification | proposal | negotiation | closed_won | closed_lost
├── expected_close_date, notes
└── closed_at
```

---

### 3.9 Team Module

**Route:** `/team`
**Purpose:** Team member management

**Primary Tables:**
```sql
team_members
├── id, user_id
├── name, email, role
├── avatar_url
├── status: AVAILABLE | BUSY | OVERLOADED | OOO
├── capacity (percentage)
├── manager_id → hierarchical
├── skills (JSONB array)
├── hourly_rate
├── share_calendar, calendar_user_id
└── joined_at

team_member_activities
├── id, member_id
├── activity_type, description
└── created_at
```

---

### 3.10 Communication Module

**Route:** `/communication`, `/communication/calendar`, `/communication/email`, `/communication/channels`
**Purpose:** Unified communication hub

**Sub-modules:**
1. **Calendar** - Google/Microsoft calendar sync
2. **Email** - Gmail/Outlook integration
3. **Channels** - Slack integration

**Calendar Events Table:**
```sql
calendar_events
├── id, user_id
├── google_event_id, calendar_id
├── title, description
├── start_time, end_time, all_day
├── location, attendees (JSONB)
├── status, visibility, html_link
├── source: google | microsoft
│
├── meeting_type: team | sales | client | internal | ...
├── context_id → linked page
├── project_id → linked project
├── client_id → linked client
│
├── recording_url, meeting_link
├── external_links (JSONB)
├── meeting_notes
├── action_items (JSONB)
└── synced_at
```

---

### 3.11 Daily Log Module

**Route:** `/daily`
**Purpose:** Daily journaling and reflection

**Primary Table:**
```sql
daily_logs
├── id, user_id
├── date (UNIQUE per user)
├── content (TEXT)
├── transcription_source
├── extracted_actions (JSONB)
├── extracted_patterns (JSONB)
└── energy_level (1-10)
```

---

### 3.12 Voice Notes Module

**Route:** `/voice-notes`
**Purpose:** Audio recording and transcription

**Primary Table:**
```sql
voice_notes
├── id, user_id
├── transcript
├── duration_seconds, word_count, words_per_minute
├── language
├── audio_file_path
├── context_id → linked page
├── project_id → linked project
└── conversation_id → linked chat
```

---

### 3.13 Integrations Module

**Route:** `/integrations`, `/integrations/[id]`
**Purpose:** Third-party service connections

**Tables:**
```sql
integration_providers (system catalog)
├── id (e.g., 'google', 'slack', 'notion')
├── name, description, category
├── icon_url, oauth_config (JSONB)
├── modules (TEXT[]) -- which modules it supports
├── skills (TEXT[]) -- which AI skills it enables
└── status

user_integrations
├── id, user_id, provider_id
├── status: connected | disconnected | error
├── access_token_encrypted, refresh_token_encrypted
├── token_expires_at, scopes
├── external_account_id/name
├── external_workspace_id/name
├── metadata, settings (JSONB)
└── connected_at, last_used_at

credential_vault (unified encrypted storage)
├── id, user_id, provider_id
├── credential_type: oauth | api_key
├── encrypted_data, encryption_version
├── expires_at, scopes
└── metadata
```

---

### 3.14 Usage/Analytics Module

**Route:** `/usage`
**Purpose:** Track AI and system usage

**Tables:**
```sql
ai_usage_logs (per request)
├── user_id, conversation_id
├── provider, model
├── input_tokens, output_tokens, total_tokens, thinking_tokens
├── agent_name, delegated_to, parent_request_id
├── request_type, duration_ms
├── context_ids, node_id, project_id
└── estimated_cost

mcp_usage_logs (tool usage)
├── user_id, tool_name, server_name
├── input_params, output_result
├── success, error_message, duration_ms
└── conversation_id, ai_request_id

usage_daily_summary (aggregated)
├── user_id, date
├── ai_requests, ai_*_tokens, ai_estimated_cost
├── provider_breakdown, model_breakdown, agent_breakdown
├── mcp_requests, mcp_tool_breakdown
├── conversations_created, messages_sent, artifacts_created
└── contexts_accessed, nodes_accessed, projects_accessed
```

---

### 3.15 Settings Module

**Route:** `/settings`, `/settings/ai`
**Purpose:** User configuration

**Tables:**
```sql
user_settings
├── user_id (UNIQUE)
├── default_model, theme
├── email_notifications, daily_summary
├── sidebar_collapsed, share_analytics
├── thinking_enabled, thinking_show_in_ui, thinking_save_traces
├── thinking_default_template_id, thinking_max_tokens
└── custom_settings (JSONB)

user_model_preferences
├── user_id
├── tier_2_model, tier_3_model, tier_4_model (JSONB)
├── tier_2/3/4_fallbacks (JSONB arrays)
├── skill_overrides (JSONB)
├── allow_model_upgrade_on_failure
├── max_latency_ms, prefer_local

user_commands (custom slash commands)
├── user_id, name, display_name
├── description, icon
├── system_prompt, context_sources
└── is_active

custom_agents
├── user_id, name, display_name
├── description, avatar
├── system_prompt, model_preference
├── temperature, max_tokens
├── capabilities, tools_enabled, context_sources
├── thinking_enabled, streaming_enabled
├── category, is_public, is_active
└── times_used, last_used_at
```

---

## 4. Complete Data Model

### 4.1 Entity Relationship Diagram (ASCII)

```
                                    USER
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
                    ▼                 ▼                 ▼
              user_settings    user_integrations    credential_vault
                    │                 │
                    │                 ▼
                    │         integration_providers
                    │
    ┌───────────────┼───────────────┬───────────────┬───────────────┐
    │               │               │               │               │
    ▼               ▼               ▼               ▼               ▼
  NODES         CONTEXTS       CONVERSATIONS    PROJECTS         CLIENTS
    │               │               │               │               │
    │               │               ▼               │               ├── contacts
    │               │           messages            │               ├── deals
    │               │               │               │               └── interactions
    │               │               ▼               │
    │               │           artifacts           │
    │               │                               │
    ├── metrics     ├── blocks                     ├── notes
    │               │   (JSONB)                    ├── statuses
    │               │                               ├── members
    │               │                               ├── tags
    │               │                               └── documents
    │               │                                   │
    │               │                                   │
    ├───────────────┼───────────────┬───────────────────┤
    │               │               │                   │
    │    node_contexts    project_conversations        │
    │               │                                   │
    │    node_conversations                            │
    │               │                                   │
    │    node_projects ─────────────────────────────────┤
    │                                                   │
    │                                                   │
    │                         TASKS ◄───────────────────┘
    │                           │
    │                           ├── task_assignees ──► team_members
    │                           ├── task_dependencies
    │                           └── parent_task_id (subtasks)
    │
    └── calendar_events (context_id, project_id, client_id)
```

### 4.2 Primary Key Patterns

| Entity | PK Type | Format |
|--------|---------|--------|
| All main tables | UUID | `gen_random_uuid()` |
| Providers | VARCHAR(50) | `google`, `slack`, etc. |
| User ID | VARCHAR(255) | From Better Auth |

### 4.3 Junction Tables Summary

| Junction Table | Connects | Type |
|----------------|----------|------|
| `node_projects` | nodes ↔ projects | M:N |
| `node_contexts` | nodes ↔ contexts | M:N |
| `node_conversations` | nodes ↔ conversations | M:N |
| `project_conversations` | projects ↔ conversations | M:N |
| `project_documents` | projects ↔ contexts | M:N |
| `project_members` | projects ↔ users/team_members | M:N |
| `project_tag_assignments` | projects ↔ project_tags | M:N |
| `task_assignees` | tasks ↔ team_members | M:N |
| `task_dependencies` | tasks ↔ tasks | M:N |

---

## 5. Data Flow & Relationships

### 5.1 The Central Hub Pattern

```
                              ┌─────────────┐
                              │    NODE     │
                              │  (the hub)  │
                              └─────────────┘
                                    │
            ┌───────────────────────┼───────────────────────┐
            │                       │                       │
            ▼                       ▼                       ▼
    ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
    │  PROJECTS   │         │  CONTEXTS   │         │CONVERSATIONS│
    │             │         │   (Pages)   │         │             │
    └─────────────┘         └─────────────┘         └─────────────┘
            │                       │                       │
            │                       │                       │
            ▼                       ▼                       ▼
    ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
    │   TASKS     │         │   BLOCKS    │         │  MESSAGES   │
    │             │         │             │         │             │
    └─────────────┘         └─────────────┘         └─────────────┘
            │                                               │
            ▼                                               ▼
    ┌─────────────┐                                 ┌─────────────┐
    │TEAM_MEMBERS │                                 │  ARTIFACTS  │
    │             │                                 │ (Creations) │
    └─────────────┘                                 └─────────────┘
```

### 5.2 Where Data Flows Most

**HIGH TRAFFIC DATA FLOWS:**

1. **Dashboard → Multiple Sources**
   ```
   Dashboard
   ├── GET /api/focus-items (today's focus)
   ├── GET /api/tasks?status=in_progress&due_before=tomorrow
   ├── GET /api/nodes?is_active=true
   ├── GET /api/calendar-events?date=today
   └── GET /api/usage/summary?date=today
   ```

2. **Chat → Context Loading**
   ```
   Chat Conversation
   ├── Loads conversation.context_id → full context
   ├── Loads node_conversations → linked nodes
   ├── Loads project_conversations → linked projects
   ├── All become AI context
   └── Artifacts saved back
   ```

3. **Project → Task Management**
   ```
   Project View
   ├── GET /api/projects/:id
   ├── GET /api/tasks?project_id=:id
   ├── GET /api/project-statuses?project_id=:id
   ├── GET /api/project-members?project_id=:id
   └── Updates flow: task status, assignees, custom status
   ```

4. **Node → Everything**
   ```
   Node Detail
   ├── GET /api/nodes/:id
   ├── GET /api/nodes/:id/projects (junction)
   ├── GET /api/nodes/:id/contexts (junction)
   ├── GET /api/nodes/:id/conversations (junction)
   ├── GET /api/node-metrics/:id
   └── PATCH updates to health, focus, decisions
   ```

### 5.3 Cross-Module Linking Patterns

**Pattern 1: Client → Everything**
```
Client "Acme Corp"
├── contexts WHERE client_id = client.id (profile pages)
├── projects WHERE client_id = client.id
├── calendar_events WHERE client_id = client.id
├── client_contacts (people at client)
├── client_deals (opportunities)
└── client_interactions (activity log)
```

**Pattern 2: Project → Everything**
```
Project "Website Redesign"
├── tasks WHERE project_id = project.id
├── project_notes
├── project_conversations (linked chats)
├── project_documents (linked pages)
├── project_members (team)
├── project_statuses (custom workflow)
├── node_projects (linked nodes)
├── calendar_events WHERE project_id = project.id
└── artifacts WHERE project_id = artifact.id
```

**Pattern 3: Calendar Event → Context**
```
Meeting "Client Kickoff"
├── context_id → meeting notes page
├── project_id → associated project
├── client_id → which client
├── meeting_notes (inline)
├── action_items (JSONB)
└── recording_url
```

---

## 6. Frontend Route Structure

### 6.1 Current Routes

```
/                           → Redirect to /dashboard
│
├── /(app)/                 → Authenticated app shell
│   ├── dashboard/          → Main overview
│   ├── chat/               → AI conversations
│   ├── knowledge-v2/       → Pages/Documents
│   │   └── [id]/           → Single page editor
│   ├── nodes/              → Operating system
│   │   └── [id]/           → Single node detail
│   ├── projects/           → Project management
│   │   └── [id]/           → Single project
│   ├── tasks/              → Task list
│   ├── clients/            → CRM
│   │   └── [id]/           → Single client
│   ├── team/               → Team members
│   ├── daily/              → Daily log/journal
│   ├── voice-notes/        → Audio transcription
│   ├── communication/      → Comms hub
│   │   ├── calendar/       → Calendar view
│   │   ├── email/          → Email inbox
│   │   └── channels/       → Slack channels
│   ├── integrations/       → Third-party connections
│   │   └── [id]/           → Integration detail
│   ├── settings/           → User settings
│   │   └── ai/             → AI preferences
│   ├── profile/            → User profile
│   ├── usage/              → Usage analytics
│   ├── terminal/           → Developer terminal
│   └── help/               → Help/docs
│
├── /(auth)/                → Authentication
│   ├── login/
│   ├── register/
│   └── forgot-password/
│
└── /api/                   → API routes (proxied to backend)
```

### 6.2 Missing Routes (Gaps)

```
MISSING:
├── /calendar               → Was deleted, now /communication/calendar
├── /knowledge              → Now /knowledge-v2
├── /nodes/active           → 404, should be handled
├── /reports                → Analytics/reporting module
├── /automations            → Workflow automation
├── /notifications          → Notification center
└── /search                 → Global search results
```

---

## 7. Backend API Structure

### 7.1 API Endpoint Map

```
/api
│
├── /auth                   → Better Auth endpoints
│
├── /conversations          → Chat
│   ├── GET /               → List conversations
│   ├── POST /              → Create conversation
│   ├── GET /:id            → Get conversation
│   ├── PATCH /:id          → Update conversation
│   ├── DELETE /:id         → Delete conversation
│   └── /messages           → Messages within
│
├── /contexts               → Knowledge/Pages
│   ├── GET /               → List contexts
│   ├── POST /              → Create context
│   ├── GET /:id            → Get context
│   ├── PATCH /:id          → Update context
│   └── DELETE /:id         → Delete context
│
├── /nodes                  → Operating System
│   ├── GET /               → List nodes
│   ├── POST /              → Create node
│   ├── GET /:id            → Get node
│   ├── PATCH /:id          → Update node
│   ├── DELETE /:id         → Delete node
│   ├── /:id/projects       → Linked projects
│   ├── /:id/contexts       → Linked contexts
│   └── /:id/conversations  → Linked conversations
│
├── /projects               → Project Management
│   ├── GET /               → List projects
│   ├── POST /              → Create project
│   ├── GET /:id            → Get project
│   ├── PATCH /:id          → Update project
│   ├── DELETE /:id         → Delete project
│   ├── /:id/tasks          → Tasks in project
│   ├── /:id/members        → Team assignment
│   └── /:id/statuses       → Custom statuses
│
├── /tasks                  → Task Management
│   ├── GET /               → List tasks
│   ├── POST /              → Create task
│   ├── GET /:id            → Get task
│   ├── PATCH /:id          → Update task
│   ├── DELETE /:id         → Delete task
│   └── /:id/subtasks       → Subtasks
│
├── /clients                → CRM
│   ├── GET /               → List clients
│   ├── POST /              → Create client
│   ├── GET /:id            → Get client
│   ├── PATCH /:id          → Update client
│   ├── DELETE /:id         → Delete client
│   ├── /:id/contacts       → Client contacts
│   ├── /:id/deals          → Client deals
│   └── /:id/interactions   → Client activities
│
├── /team                   → Team Management
│   ├── GET /members        → List members
│   ├── POST /members       → Create member
│   ├── GET /members/:id    → Get member
│   ├── PATCH /members/:id  → Update member
│   └── DELETE /members/:id → Delete member
│
├── /focus-items            → Dashboard Focus
│   ├── GET /               → List focus items
│   ├── POST /              → Create focus item
│   ├── PATCH /:id          → Update (toggle complete)
│   └── DELETE /:id         → Delete
│
├── /daily-logs             → Daily Journal
│   ├── GET /               → List logs
│   ├── GET /:date          → Get log by date
│   ├── POST /              → Create/update log
│   └── DELETE /:date       → Delete log
│
├── /voice-notes            → Voice Transcription
│   ├── GET /               → List notes
│   ├── POST /              → Create (with audio)
│   └── DELETE /:id         → Delete
│
├── /calendar               → Calendar Sync
│   ├── GET /events         → List events
│   ├── GET /events/:id     → Get event
│   ├── POST /events        → Create event
│   ├── PATCH /events/:id   → Update event
│   └── POST /sync          → Force sync
│
├── /integrations           → Third-party
│   ├── GET /providers      → List providers
│   ├── GET /               → User integrations
│   ├── POST /connect/:id   → Start OAuth
│   ├── DELETE /:id         → Disconnect
│   └── POST /:id/sync      → Force sync
│
├── /artifacts              → AI Creations
│   ├── GET /               → List artifacts
│   ├── GET /:id            → Get artifact
│   ├── PATCH /:id          → Update artifact
│   └── DELETE /:id         → Delete
│
├── /usage                  → Analytics
│   ├── GET /summary        → Daily summary
│   ├── GET /ai             → AI usage details
│   └── GET /mcp            → MCP tool usage
│
├── /settings               → User Settings
│   ├── GET /               → Get settings
│   ├── PATCH /             → Update settings
│   └── /agents             → Custom agents
│
└── /ai                     → AI Endpoints
    ├── POST /chat          → Chat completion
    ├── POST /chat/stream   → Streaming chat (SSE)
    ├── POST /extract       → Entity extraction
    └── POST /analyze       → Analysis
```

---

## 8. Integration Architecture

### 8.1 Supported Integrations

| Provider | Category | Modules | Status |
|----------|----------|---------|--------|
| **Google** | Productivity | Calendar, Gmail, Drive, Tasks, Contacts | Building |
| **Microsoft** | Productivity | Outlook, Calendar, Teams, OneDrive | Building |
| **Slack** | Communication | Channels, Messages | Building |
| **Notion** | Productivity | Pages, Databases | Building |
| **Linear** | Development | Issues, Projects, Cycles | Building |
| **HubSpot** | CRM | Contacts, Deals, Companies | Building |
| **Airtable** | Database | Bases, Tables, Records | Building |
| **ClickUp** | Projects | Spaces, Tasks, Docs | Building |
| **Fathom** | Meetings | Recordings, Transcripts | Building |

### 8.2 Integration Data Flow

```
┌─────────────┐     OAuth     ┌─────────────┐
│   User      │──────────────▶│  Provider   │
│  (Browser)  │◀──────────────│  (Google)   │
└─────────────┘   Callback    └─────────────┘
       │                             │
       │                             │
       ▼                             │
┌─────────────┐                      │
│  Frontend   │                      │
│  /integr... │                      │
└─────────────┘                      │
       │                             │
       │ API                         │
       ▼                             │
┌─────────────┐     API Calls  ┌─────────────┐
│   Backend   │──────────────▶│  Provider   │
│  Go Server  │◀──────────────│    API      │
└─────────────┘   Data         └─────────────┘
       │
       │ Store
       ▼
┌─────────────┐
│  PostgreSQL │
│ (encrypted) │
└─────────────┘
```

### 8.3 Integration Tables Mapping

| Provider | Data | Maps To |
|----------|------|---------|
| Google Calendar | Events | `calendar_events` |
| Google Calendar | Attendees | `calendar_events.attendees` (JSONB) |
| Gmail | Messages | `emails` (needs table) |
| Gmail | Threads | `email_threads` (needs table) |
| Google Contacts | Contacts | `client_contacts` or new table |
| Google Drive | Files | `drive_files` (needs table) |
| Google Tasks | Tasks | `tasks` with `source='google'` |
| Slack | Channels | `slack_channels` (needs table) |
| Slack | Messages | `slack_messages` (needs table) |
| Notion | Pages | `notion_pages` (needs table) |
| Notion | Databases | `notion_databases` (needs table) |
| Linear | Issues | `linear_issues` (needs table) |
| Linear | Projects | Link to `projects` |
| HubSpot | Contacts | `clients` + `client_contacts` |
| HubSpot | Deals | `client_deals` |
| HubSpot | Companies | `clients` |

---

## 9. What's Missing / Gaps

### 9.1 Critical Missing Features

| Feature | Impact | Effort |
|---------|--------|--------|
| **Share Modal** | Can't share pages | Medium |
| **Export (MD/PDF)** | Can't export content | Medium |
| **HTML Paste Parsing** | Poor copy/paste UX | High |
| **Database Views** | No Kanban/Calendar views in pages | High |
| **Global Search** | No search across modules | High |
| **Notifications** | No notification system | Medium |
| **Comments** | No commenting on pages | Medium |
| **Version History** | No page history | Medium |

### 9.2 Missing Database Tables

```sql
-- Email integration
CREATE TABLE emails (...);
CREATE TABLE email_threads (...);

-- Slack integration
CREATE TABLE slack_channels (...);
CREATE TABLE slack_messages (...);

-- Notion integration
CREATE TABLE notion_pages (...);
CREATE TABLE notion_databases (...);
CREATE TABLE notion_blocks (...);

-- Linear integration
CREATE TABLE linear_issues (...);
CREATE TABLE linear_projects (...);

-- Notifications
CREATE TABLE notifications (...);

-- Comments
CREATE TABLE comments (...);

-- Page versions
CREATE TABLE context_versions (...);
```

### 9.3 Missing Frontend Features

**Knowledge Module:**
- [ ] Share modal with permissions
- [ ] Export to Markdown/PDF
- [ ] HTML paste → blocks conversion
- [ ] Database views (Kanban, Calendar, Gallery, List)
- [ ] Block-level comments
- [ ] Page history/versions
- [ ] Template gallery
- [ ] AI writing assistance in blocks

**Projects Module:**
- [ ] Gantt chart view
- [ ] Timeline view
- [ ] Resource allocation view
- [ ] Project templates gallery

**Dashboard:**
- [ ] Customizable widgets
- [ ] Activity feed with all modules
- [ ] Quick actions panel
- [ ] Mobile responsive

**Global:**
- [ ] Command palette (⌘K)
- [ ] Global search
- [ ] Notification center
- [ ] Keyboard shortcuts

### 9.4 Backend API Gaps

```
MISSING ENDPOINTS:

/api/search
├── GET /global          → Search across all modules
├── GET /pages           → Search pages
└── GET /tasks           → Search tasks

/api/notifications
├── GET /                → List notifications
├── PATCH /:id/read      → Mark read
└── DELETE /:id          → Dismiss

/api/comments
├── GET /:entityType/:id → Get comments
├── POST /               → Add comment
├── PATCH /:id           → Edit comment
└── DELETE /:id          → Delete comment

/api/contexts/:id
├── GET /versions        → Version history
├── POST /restore/:versionId → Restore version
└── POST /export         → Export to format

/api/contexts/:id/share
├── POST /               → Create share link
├── GET /                → Get share settings
├── PATCH /              → Update permissions
└── DELETE /             → Remove share
```

---

## 10. Implementation Roadmap

### Phase 1: Core Stability (Current)
- [x] All modules functional
- [x] Database schema complete
- [x] Basic integrations structure
- [ ] Fix 404 routes
- [ ] Knowledge module polish

### Phase 2: Knowledge Module Complete
- [ ] Share modal
- [ ] Export functionality
- [ ] HTML paste parsing
- [ ] Block-level improvements
- [ ] Template system

### Phase 3: Database Views
- [ ] Kanban view
- [ ] Calendar view
- [ ] Gallery view
- [ ] List view with sorting/filtering

### Phase 4: Global Features
- [ ] Global search
- [ ] Notifications
- [ ] Comments
- [ ] Command palette

### Phase 5: Integration Completion
- [ ] Google suite complete
- [ ] Microsoft suite complete
- [ ] Slack complete
- [ ] Notion complete
- [ ] Linear complete

### Phase 6: Advanced Features
- [ ] Automation/workflows
- [ ] AI skills enhancement
- [ ] Mobile app
- [ ] API for external apps

---

## Appendix A: Naming Conventions

### Entity Names

| Code Name | User-Facing Name |
|-----------|------------------|
| `contexts` | Pages |
| `artifacts` | Creations |
| `focus_items` | Focus |
| `team_members` | Members |
| `client_contacts` | Contacts |
| `client_interactions` | Activities |
| `daily_logs` | Entries |

### Node Types

| Database Value | Display Name |
|----------------|--------------|
| `BUSINESS` | Business/Venture |
| `PROJECT` | Initiative |
| `LEARNING` | Learning |
| `OPERATIONAL` | Operations |

### Status Enums

| Entity | Values |
|--------|--------|
| Task | `todo`, `in_progress`, `done`, `cancelled` |
| Project | `ACTIVE`, `PAUSED`, `COMPLETED`, `ARCHIVED` |
| Node Health | `HEALTHY`, `NEEDS_ATTENTION`, `CRITICAL`, `NOT_STARTED` |
| Client | `lead`, `prospect`, `active`, `inactive`, `churned` |

---

## Appendix B: Quick Reference

### Common Queries

**Get user's active context:**
```sql
SELECT * FROM nodes WHERE user_id = ? AND is_active = true;
```

**Get project with tasks:**
```sql
SELECT p.*,
       json_agg(t.*) as tasks
FROM projects p
LEFT JOIN tasks t ON t.project_id = p.id
WHERE p.id = ?
GROUP BY p.id;
```

**Get page with all relationships:**
```sql
SELECT c.*,
       nc.node_id,
       pd.project_id
FROM contexts c
LEFT JOIN node_contexts nc ON nc.context_id = c.id
LEFT JOIN project_documents pd ON pd.document_id = c.id
WHERE c.id = ?;
```

### Key Relationships

```
Node ──M:N──┬── Projects
            ├── Contexts (Pages)
            └── Conversations

Project ──1:N──┬── Tasks
               ├── Notes
               ├── Statuses
               └── Members

Client ──1:N──┬── Contacts
              ├── Deals
              └── Interactions

Conversation ──1:N── Messages ──1:N── Artifacts
```

---

*This is the complete architecture reference for BusinessOS. Use this as the source of truth for understanding the system.*
