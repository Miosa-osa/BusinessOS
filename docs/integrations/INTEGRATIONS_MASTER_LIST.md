# BusinessOS Integrations Master List

## Executive Summary

This is the **DEFINITIVE** integration inventory for BusinessOS. Every provider, every MCP tool, every file import parser is catalogued here.

**Total Scope:**
- **35+ OAuth Providers** across 10 categories
- **68+ MCP Tools** for AI agent access
- **13+ File Import Parsers** for manual data import
- **100+ API Endpoints** for backend operations
- **50+ Frontend Components** for settings UI

---

## Integration Categories

| Category | Provider Count | Description |
|----------|----------------|-------------|
| **Calendar** | 4 | Google Calendar, Outlook, Calendly, Apple Calendar |
| **Communication** | 4 | Slack, Microsoft Teams, Discord, Email |
| **Tasks/Projects** | 7 | ClickUp, Asana, Linear, Monday, Jira, Trello, Todoist |
| **CRM** | 5 | HubSpot, Salesforce, Pipedrive, GoHighLevel, Freshsales |
| **Storage/Docs** | 6 | Google Drive, OneDrive, Dropbox, Box, Notion, Confluence |
| **Video/Meeting** | 6 | Zoom, Google Meet, Teams, Loom, Fireflies, Fathom, tl;dv |
| **Code** | 2 | GitHub, GitLab |
| **Finance** | 3 | Stripe, QuickBooks, Xero |
| **HR** | 1 | BambooHR |
| **Design** | 2 | Figma, Miro |

---

## Complete Provider Status Matrix

### Production Ready (3)

| Provider | Category | OAuth | API | MCP Tools | Sync | UI |
|----------|----------|-------|-----|-----------|------|-----|
| **Google Calendar** | Calendar | DONE | DONE | 5 | DONE | - |
| **Slack** | Communication | DONE | DONE | 6 | PARTIAL | - |
| **Notion** | Docs | DONE | DONE | 7 | PARTIAL | - |

---

### Phase 1: Foundation Infrastructure (CRITICAL)

Core infrastructure required before adding new integrations:

| Component | Status | Priority |
|-----------|--------|----------|
| Provider Registry Pattern | NEED | CRITICAL |
| OAuth State Management | NEED | CRITICAL |
| Token Encryption | NEED | CRITICAL |
| Integration Settings UI | NEED | CRITICAL |
| Connection Status API | NEED | CRITICAL |

---

### Phase 2: File Import System (HIGH)

| Provider | File Format | Parser | Enrichment | Priority |
|----------|-------------|--------|------------|----------|
| **ChatGPT** | JSON export | NEED | NEED | HIGH |
| **Claude** | JSON export | NEED | NEED | HIGH |
| **Perplexity** | JSON export | NEED | NEED | HIGH |
| **Gemini** | JSON export | NEED | NEED | MEDIUM |
| **Granola** | Meeting JSON | NEED | NEED | MEDIUM |
| **Obsidian** | Markdown vault | NEED | NEED | LOW |
| **Roam Research** | JSON export | NEED | NEED | LOW |
| **Evernote** | ENEX XML | NEED | NEED | LOW |
| **Notion Export** | ZIP/Markdown | NEED | NEED | LOW |
| **Apple Notes** | SQLite/HTML | NEED | NEED | LOW |
| **CSV** | Generic CSV | NEED | NEED | MEDIUM |
| **Markdown** | .md files | NEED | NEED | MEDIUM |

---

### Phase 3: Google Ecosystem (HIGH)

Extends existing Google OAuth with additional scopes.

| Service | OAuth | API | MCP Tools | Sync | Priority |
|---------|-------|-----|-----------|------|----------|
| **Gmail** | Google OAuth | NEED | 5 | NEED | HIGH |
| **Google Drive** | Google OAuth | NEED | 5 | NEED | HIGH |
| **Google Meet** | Google OAuth | NEED | 3 | NEED | MEDIUM |
| **Google Contacts** | Google OAuth | NEED | 0 | NEED | LOW |

---

### Phase 4: Task/Project Integrations (HIGH)

| Provider | OAuth | API | MCP Tools | Sync | Priority |
|----------|-------|-----|-----------|------|----------|
| **ClickUp** | ClickUp OAuth | NEED | 6 | NEED | HIGH |
| **Asana** | Asana OAuth | NEED | 5 | NEED | HIGH |
| **Linear** | Linear OAuth | STUB | 5 | NEED | MEDIUM |
| **Monday.com** | Monday OAuth | NEED | 4 | NEED | MEDIUM |
| **Jira** | Atlassian OAuth | NEED | 5 | NEED | LOW |
| **Trello** | Atlassian OAuth | NEED | 4 | NEED | LOW |
| **Todoist** | Todoist OAuth | NEED | 4 | NEED | LOW |

---

### Phase 5: CRM Integrations (MEDIUM)

| Provider | OAuth | API | MCP Tools | Sync | Priority |
|----------|-------|-----|-----------|------|----------|
| **HubSpot** | HubSpot OAuth | STUB | 6 | NEED | HIGH |
| **Salesforce** | Salesforce OAuth | NEED | 4 | NEED | LOW |
| **Pipedrive** | Pipedrive OAuth | NEED | 4 | NEED | LOW |
| **GoHighLevel** | GHL OAuth | STUB | 4 | NEED | MEDIUM |
| **Freshsales** | Freshsales OAuth | NEED | 4 | NEED | LOW |

---

### Phase 6: Video/Meeting Integrations (MEDIUM)

| Provider | OAuth | API | MCP Tools | Sync | Priority |
|----------|-------|-----|-----------|------|----------|
| **Zoom** | Zoom OAuth | NEED | 4 | NEED | HIGH |
| **Loom** | Loom OAuth | NEED | 3 | NEED | LOW |
| **Fireflies.ai** | API Key | NEED | 3 | NEED | MEDIUM |
| **Fathom** | API Key | NEED | 3 | NEED | MEDIUM |
| **tl;dv** | API Key | NEED | 3 | NEED | LOW |
| **Calendly** | Calendly OAuth | NEED | 4 | NEED | MEDIUM |

---

### Phase 7: Microsoft Ecosystem (MEDIUM)

| Service | OAuth | API | MCP Tools | Sync | Priority |
|---------|-------|-----|-----------|------|----------|
| **Outlook Calendar** | Microsoft OAuth | NEED | 5 | NEED | MEDIUM |
| **Microsoft Teams** | Microsoft OAuth | NEED | 4 | NEED | MEDIUM |
| **OneDrive** | Microsoft OAuth | NEED | 4 | NEED | LOW |
| **Microsoft To Do** | Microsoft OAuth | NEED | 4 | NEED | LOW |

---

### Phase 8: Storage/Collaboration (LOW)

| Provider | OAuth | API | MCP Tools | Sync | Priority |
|----------|-------|-----|-----------|------|----------|
| **Dropbox** | Dropbox OAuth | NEED | 4 | NEED | MEDIUM |
| **Box** | Box OAuth | NEED | 4 | NEED | LOW |
| **Confluence** | Atlassian OAuth | NEED | 4 | NEED | LOW |
| **Airtable** | Airtable OAuth | NEED | 4 | NEED | LOW |

---

### Phase 9: Code/DevOps (LOW)

| Provider | OAuth | API | MCP Tools | Sync | Priority |
|----------|-------|-----|-----------|------|----------|
| **GitHub** | GitHub OAuth | NEED | 5 | NEED | LOW |
| **GitLab** | GitLab OAuth | NEED | 5 | NEED | LOW |

---

### Phase 10: Finance/HR/Design (LOW)

| Provider | OAuth | API | MCP Tools | Sync | Priority |
|----------|-------|-----|-----------|------|----------|
| **Stripe** | Stripe Connect | NEED | 4 | NEED | LOW |
| **QuickBooks** | Intuit OAuth | NEED | 4 | NEED | LOW |
| **Xero** | Xero OAuth | NEED | 4 | NEED | LOW |
| **BambooHR** | API Key | NEED | 3 | NEED | LOW |
| **Figma** | Figma OAuth | NEED | 3 | NEED | LOW |
| **Miro** | Miro OAuth | NEED | 3 | NEED | LOW |

---

## Complete MCP Tools Inventory

### Currently Implemented (18 tools)

#### Google Calendar (5 tools)
| Tool | Description | Status |
|------|-------------|--------|
| `calendar_list_events` | List events in date range | DONE |
| `calendar_create_event` | Create event with attendees | DONE |
| `calendar_update_event` | Update existing event | DONE |
| `calendar_delete_event` | Delete event | DONE |
| `calendar_sync_events` | Sync events to database | DONE |

#### Slack (6 tools)
| Tool | Description | Status |
|------|-------------|--------|
| `slack_list_channels` | List public/private channels | DONE |
| `slack_send_message` | Send message with thread support | DONE |
| `slack_get_channel_history` | Get channel messages | DONE |
| `slack_search_messages` | Search across workspace | DONE |
| `slack_list_users` | List workspace members | DONE |
| `slack_get_user_info` | Get user details | DONE |

#### Notion (7 tools)
| Tool | Description | Status |
|------|-------------|--------|
| `notion_list_databases` | List accessible databases | DONE |
| `notion_get_database` | Get database schema | DONE |
| `notion_query_database` | Query with filters/sorts | DONE |
| `notion_get_page` | Get page content | DONE |
| `notion_create_page` | Create new page | DONE |
| `notion_update_page` | Update page properties | DONE |
| `notion_search` | Search workspace | DONE |

---

### Planned MCP Tools (50+ tools)

#### Gmail (5 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `gmail_search` | Search emails | HIGH |
| `gmail_get_thread` | Get email thread | HIGH |
| `gmail_list_labels` | List labels | HIGH |
| `gmail_get_message` | Get single message | HIGH |
| `gmail_list_recent` | List recent emails | HIGH |

#### Google Drive (5 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `drive_search` | Search files | HIGH |
| `drive_get_file` | Get file metadata | HIGH |
| `drive_get_content` | Get file content | HIGH |
| `drive_list_folder` | List folder contents | HIGH |
| `drive_get_permissions` | Get sharing permissions | MEDIUM |

#### Google Meet (3 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `meet_create_meeting` | Create meeting link | MEDIUM |
| `meet_get_meeting` | Get meeting details | MEDIUM |
| `meet_list_recordings` | List recordings | MEDIUM |

#### Microsoft Outlook (5 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `outlook_list_events` | List calendar events | MEDIUM |
| `outlook_create_event` | Create event | MEDIUM |
| `outlook_list_emails` | List emails | MEDIUM |
| `outlook_get_email` | Get email | MEDIUM |
| `outlook_search` | Search emails | MEDIUM |

#### Microsoft Teams (4 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `teams_send_message` | Send message | MEDIUM |
| `teams_list_channels` | List channels | MEDIUM |
| `teams_get_presence` | Get user presence | MEDIUM |
| `teams_list_members` | List team members | MEDIUM |

#### OneDrive (4 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `onedrive_search` | Search files | LOW |
| `onedrive_get_file` | Get file | LOW |
| `onedrive_list_folder` | List folder | LOW |
| `onedrive_get_content` | Get content | LOW |

#### ClickUp (6 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `clickup_list_tasks` | List tasks | HIGH |
| `clickup_get_task` | Get task details | HIGH |
| `clickup_create_task` | Create task | HIGH |
| `clickup_update_task` | Update task | HIGH |
| `clickup_list_spaces` | List spaces | HIGH |
| `clickup_list_lists` | List lists | HIGH |

#### Asana (5 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `asana_list_tasks` | List tasks | HIGH |
| `asana_get_task` | Get task | HIGH |
| `asana_create_task` | Create task | HIGH |
| `asana_update_task` | Update task | HIGH |
| `asana_list_projects` | List projects | HIGH |

#### Linear (5 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `linear_list_issues` | List issues | MEDIUM |
| `linear_get_issue` | Get issue | MEDIUM |
| `linear_create_issue` | Create issue | MEDIUM |
| `linear_update_issue` | Update issue | MEDIUM |
| `linear_list_projects` | List projects | MEDIUM |

#### HubSpot (6 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `hubspot_get_contact` | Get contact | HIGH |
| `hubspot_list_contacts` | List contacts | HIGH |
| `hubspot_create_contact` | Create contact | HIGH |
| `hubspot_list_deals` | List deals | HIGH |
| `hubspot_create_deal` | Create deal | MEDIUM |
| `hubspot_search` | Search CRM | HIGH |

#### Zoom (4 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `zoom_create_meeting` | Schedule meeting | HIGH |
| `zoom_list_meetings` | List meetings | HIGH |
| `zoom_get_recording` | Get recording | HIGH |
| `zoom_get_transcript` | Get transcript | HIGH |

#### GitHub (5 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `github_list_repos` | List repositories | LOW |
| `github_list_issues` | List issues | LOW |
| `github_create_issue` | Create issue | LOW |
| `github_list_prs` | List pull requests | LOW |
| `github_get_pr` | Get PR details | LOW |

#### Jira (5 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `jira_list_issues` | List issues | LOW |
| `jira_get_issue` | Get issue | LOW |
| `jira_create_issue` | Create issue | LOW |
| `jira_update_issue` | Update issue | LOW |
| `jira_list_projects` | List projects | LOW |

#### Search Tools (4 tools)
| Tool | Description | Priority |
|------|-------------|----------|
| `search_imports` | Search imported conversations | MEDIUM |
| `search_knowledge` | Search extracted knowledge | MEDIUM |
| `search_files` | Search synced files | MEDIUM |
| `search_all` | Unified search across all data | MEDIUM |

---

## API Endpoints Reference

### Core Integration Endpoints

```
GET    /api/integrations                              # List all available integrations
GET    /api/integrations/status                       # Get status of all connected
GET    /api/integrations/categories                   # List integration categories

# Per-Provider Pattern (applies to ALL providers)
GET    /api/integrations/:provider                    # Get provider details
GET    /api/integrations/:provider/status             # Get connection status
GET    /api/integrations/:provider/auth               # Get OAuth URL
GET    /api/integrations/:provider/callback           # OAuth callback
DELETE /api/integrations/:provider                    # Disconnect integration
PUT    /api/integrations/:provider/settings           # Update settings
GET    /api/integrations/:provider/settings           # Get settings

# Sync Endpoints (per provider)
POST   /api/integrations/:provider/sync               # Trigger manual sync
GET    /api/integrations/:provider/sync               # Get current sync status
GET    /api/integrations/:provider/sync/history       # Get sync history
DELETE /api/integrations/:provider/sync/:job_id       # Cancel sync job
PUT    /api/integrations/:provider/sync/schedule      # Update sync schedule

# Audit & Logs
GET    /api/integrations/:provider/logs               # Get audit logs
GET    /api/integrations/:provider/errors             # Get recent errors
```

### Provider-Specific Endpoints

#### Google Calendar
```
GET    /api/integrations/google/calendars             # List calendars
GET    /api/integrations/google/calendars/:id/events  # List events
POST   /api/integrations/google/calendars/:id/events  # Create event
PUT    /api/integrations/google/calendars/:id/events/:eventId
DELETE /api/integrations/google/calendars/:id/events/:eventId
```

#### Gmail
```
GET    /api/integrations/gmail/threads                # List threads
GET    /api/integrations/gmail/threads/:id            # Get thread
GET    /api/integrations/gmail/messages               # Search messages
GET    /api/integrations/gmail/labels                 # List labels
```

#### Google Drive
```
GET    /api/integrations/drive/files                  # List/search files
GET    /api/integrations/drive/files/:id              # Get file details
GET    /api/integrations/drive/files/:id/content      # Get file content
GET    /api/integrations/drive/folders                # List folders
GET    /api/integrations/drive/folders/:id            # List folder contents
```

#### Slack
```
GET    /api/integrations/slack/channels               # List channels
GET    /api/integrations/slack/channels/:id/messages  # Get channel messages
POST   /api/integrations/slack/channels/:id/messages  # Send message
GET    /api/integrations/slack/users                  # List users
GET    /api/integrations/slack/search                 # Search messages
```

#### Notion
```
GET    /api/integrations/notion/databases             # List databases
GET    /api/integrations/notion/databases/:id         # Get database
POST   /api/integrations/notion/databases/:id/query   # Query database
GET    /api/integrations/notion/pages/:id             # Get page
POST   /api/integrations/notion/pages                 # Create page
PATCH  /api/integrations/notion/pages/:id             # Update page
GET    /api/integrations/notion/search                # Search
```

#### ClickUp
```
GET    /api/integrations/clickup/workspaces           # List workspaces
GET    /api/integrations/clickup/spaces               # List spaces
GET    /api/integrations/clickup/folders              # List folders
GET    /api/integrations/clickup/lists                # List lists
GET    /api/integrations/clickup/tasks                # List tasks
POST   /api/integrations/clickup/tasks                # Create task
PUT    /api/integrations/clickup/tasks/:id            # Update task
DELETE /api/integrations/clickup/tasks/:id            # Delete task
```

#### Asana
```
GET    /api/integrations/asana/workspaces             # List workspaces
GET    /api/integrations/asana/projects               # List projects
GET    /api/integrations/asana/tasks                  # List tasks
POST   /api/integrations/asana/tasks                  # Create task
PUT    /api/integrations/asana/tasks/:id              # Update task
```

#### Linear
```
GET    /api/integrations/linear/teams                 # List teams
GET    /api/integrations/linear/projects              # List projects
GET    /api/integrations/linear/issues                # List issues
POST   /api/integrations/linear/issues                # Create issue
PUT    /api/integrations/linear/issues/:id            # Update issue
```

#### HubSpot
```
GET    /api/integrations/hubspot/contacts             # List contacts
GET    /api/integrations/hubspot/contacts/:id         # Get contact
POST   /api/integrations/hubspot/contacts             # Create contact
PUT    /api/integrations/hubspot/contacts/:id         # Update contact
GET    /api/integrations/hubspot/companies            # List companies
GET    /api/integrations/hubspot/deals                # List deals
POST   /api/integrations/hubspot/deals                # Create deal
```

#### Zoom
```
GET    /api/integrations/zoom/meetings                # List meetings
POST   /api/integrations/zoom/meetings                # Create meeting
GET    /api/integrations/zoom/meetings/:id            # Get meeting
GET    /api/integrations/zoom/recordings              # List recordings
GET    /api/integrations/zoom/recordings/:id          # Get recording/transcript
```

#### Microsoft Teams
```
GET    /api/integrations/teams/teams                  # List teams
GET    /api/integrations/teams/channels               # List channels
POST   /api/integrations/teams/messages               # Send message
GET    /api/integrations/teams/presence               # Get presence
```

#### GitHub
```
GET    /api/integrations/github/repos                 # List repos
GET    /api/integrations/github/repos/:owner/:repo/issues
POST   /api/integrations/github/repos/:owner/:repo/issues
GET    /api/integrations/github/repos/:owner/:repo/pulls
```

### File Import Endpoints

```
POST   /api/imports/upload                            # Upload file for import
GET    /api/imports                                   # List all imports
GET    /api/imports/:id                               # Get import details
DELETE /api/imports/:id                               # Delete import
POST   /api/imports/:id/retry                         # Retry failed import
POST   /api/imports/:id/cancel                        # Cancel in-progress import

# Import Progress
GET    /api/imports/:id/progress                      # Get real-time progress (SSE)
GET    /api/imports/:id/stats                         # Get import statistics

# Imported Data Access
GET    /api/imports/conversations                     # List all imported conversations
GET    /api/imports/conversations/:id                 # Get conversation with messages
GET    /api/imports/conversations/:id/messages        # Get messages only
GET    /api/imports/conversations/:id/knowledge       # Get extracted knowledge
POST   /api/imports/conversations/:id/link            # Link to project/client

# Knowledge Access
GET    /api/imports/knowledge                         # List extracted knowledge
GET    /api/imports/knowledge/:id                     # Get knowledge item
PUT    /api/imports/knowledge/:id                     # Update/verify knowledge
POST   /api/imports/knowledge/:id/create-memory       # Create memory from knowledge

# Search
GET    /api/imports/search                            # Semantic search across imports
GET    /api/imports/search/similar                    # Find similar content
```

### Webhook Endpoints

```
# Incoming Webhooks (from providers)
POST   /api/webhooks/slack                            # Slack events
POST   /api/webhooks/github                           # GitHub events
POST   /api/webhooks/stripe                           # Stripe events
POST   /api/webhooks/hubspot                          # HubSpot events
POST   /api/webhooks/clickup                          # ClickUp events
POST   /api/webhooks/linear                           # Linear events
POST   /api/webhooks/zoom                             # Zoom events
POST   /api/webhooks/calendly                         # Calendly events

# Webhook Management
GET    /api/webhooks/subscriptions                    # List webhook subscriptions
POST   /api/webhooks/subscriptions                    # Create subscription
DELETE /api/webhooks/subscriptions/:id                # Delete subscription
GET    /api/webhooks/deliveries                       # List webhook deliveries
GET    /api/webhooks/deliveries/:id                   # Get delivery details
POST   /api/webhooks/deliveries/:id/retry             # Retry delivery
```

---

## API vs MCP Decision Guide

### When to Use Direct API (NOT MCP)

- Data sync/import (runs automatically)
- Background workers
- User authentication flows
- Bulk operations
- Bi-directional sync

### When to Use MCP Tools (AI-accessible)

- User wants AI to "check my calendar"
- User wants AI to "send a Slack message"
- User wants AI to "find files in Drive"
- User wants AI to "create a task in ClickUp"
- Real-time interactions through conversation

### When to Use BOTH

- Google Calendar (sync events + AI can check/create)
- Slack (sync channels + AI can message)
- ClickUp/Asana (sync tasks + AI can manage tasks)
- HubSpot (sync contacts + AI can lookup/create)

---

## Environment Variables

### Currently Configured (Production Ready)

```bash
# Google (Calendar, Gmail, Drive, Meet)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URI=

# Slack
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
SLACK_REDIRECT_URI=

# Notion
NOTION_CLIENT_ID=
NOTION_CLIENT_SECRET=
NOTION_REDIRECT_URI=
```

### Phase 4-6: Task & CRM Integrations

```bash
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
```

### Phase 7: Video/Meeting Integrations

```bash
# Zoom
ZOOM_CLIENT_ID=
ZOOM_CLIENT_SECRET=
ZOOM_REDIRECT_URI=

# Fireflies (API Key based)
FIREFLIES_API_KEY=

# Fathom (API Key based)
FATHOM_API_KEY=
```

### Phase 8: AI/Enrichment Services

```bash
# OpenAI (for embeddings)
OPENAI_API_KEY=
```

### Future Integrations

```bash
# Microsoft (Outlook, Teams, OneDrive)
MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
MICROSOFT_REDIRECT_URI=

# Jira/Trello/Confluence (Atlassian)
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
```

---

## Implementation Priority Summary

| Phase | Focus | Priority | Effort |
|-------|-------|----------|--------|
| **Phase 1** | Foundation Infrastructure | CRITICAL | 1-2 weeks |
| **Phase 2** | File Import System | HIGH | 1-2 weeks |
| **Phase 3** | Google Ecosystem | HIGH | 1 week |
| **Phase 4** | Task Integrations | HIGH | 1-2 weeks |
| **Phase 5** | CRM Integrations | MEDIUM | 1 week |
| **Phase 6** | Video/Meeting | MEDIUM | 1 week |
| **Phase 7** | Microsoft Ecosystem | MEDIUM | 1 week |
| **Phase 8** | Enrichment & Search | MEDIUM | 1 week |
| **Phase 9** | Webhooks | LOW | 1 week |
| **Phase 10** | Remaining Integrations | ONGOING | As needed |

---

## Summary Statistics

| Category | Count | Implemented | Planned |
|----------|-------|-------------|---------|
| OAuth Integrations | 35+ | 3 | 32+ |
| File Import Parsers | 13+ | 0 | 13+ |
| MCP Tools | 68+ | 18 | 50+ |
| API Endpoints | 100+ | ~30 | 70+ |
| Sync Workers | 30+ | 1 | 29+ |
| Frontend Components | 50+ | 0 | 50+ |

---

## Related Documents

- `INTEGRATION_IMPLEMENTATION_PLAN.md` - Complete file tree structure and implementation details
- `INTEGRATION_INFRASTRUCTURE.md` - Architecture patterns and code examples
- `FILE_IMPORT_ARCHITECTURE.md` - Detailed file import system design

---

This is the **DEFINITIVE** master list. Use alongside implementation plan for full architecture details.
