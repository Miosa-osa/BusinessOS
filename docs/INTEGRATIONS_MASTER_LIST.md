# BusinessOS Integrations Master List

## Overview

This document maps ALL planned integrations, distinguishing between:
- **Direct API** - Backend sync, OAuth, data import/export
- **MCP Tools** - AI agent can use (read/write via conversation)
- **File Import** - Manual file upload, no OAuth needed

---

## Integration Status Matrix

### PRODUCTION READY (3)

| Provider | OAuth | Direct API | MCP Tools | Sync Worker | UI |
|----------|-------|------------|-----------|-------------|-----|
| **Google Calendar** | DONE | DONE | 5 tools | DONE | NO |
| **Slack** | DONE | DONE | 6 tools | PARTIAL | NO |
| **Notion** | DONE | DONE | 7 tools | PARTIAL | NO |

---

### HIGH PRIORITY - Knowledge Module

| Provider | OAuth | Direct API | MCP Tools | Sync Worker | Priority |
|----------|-------|------------|-----------|-------------|----------|
| **Gmail** | Google OAuth | NEED | NEED | NEED | HIGH |
| **Google Drive** | Google OAuth | NEED | NEED | NEED | HIGH |
| **HubSpot** | HubSpot OAuth | STUB | NEED | NEED | HIGH |

**Gmail Requirements:**
- OAuth: Use existing Google OAuth (add `gmail.readonly` scope)
- API: Fetch emails, extract key info
- MCP: `gmail_search`, `gmail_get_thread`, `gmail_list_labels`
- Sync: Background worker to index emails

**Google Drive Requirements:**
- OAuth: Use existing Google OAuth (add `drive.readonly` scope)
- API: List files, read contents, index documents
- MCP: `drive_search`, `drive_get_file`, `drive_list_folder`
- Sync: Background worker to index documents

**HubSpot Requirements:**
- OAuth: HubSpot OAuth 2.0 (NEW)
- API: Contacts, companies, deals, activities
- MCP: `hubspot_get_contact`, `hubspot_list_deals`, `hubspot_create_task`
- Sync: Bi-directional sync with Clients module

---

### MEDIUM PRIORITY - Task/Project Integrations

| Provider | OAuth | Direct API | MCP Tools | Sync Worker | Priority |
|----------|-------|------------|-----------|-------------|----------|
| **ClickUp** | ClickUp OAuth | NEED | NEED | NEED | HIGH |
| **Asana** | Asana OAuth | STUB | NEED | NEED | HIGH |
| **Linear** | Linear OAuth | STUB | NEED | NEED | MEDIUM |
| **Monday.com** | Monday OAuth | NEED | OPTIONAL | NEED | MEDIUM |
| **Jira** | Atlassian OAuth | NEED | OPTIONAL | NEED | LOW |
| **Trello** | Atlassian OAuth | NEED | OPTIONAL | NEED | LOW |

**ClickUp Requirements:**
- OAuth: ClickUp OAuth 2.0
- API: Tasks, lists, spaces, folders
- MCP: `clickup_list_tasks`, `clickup_create_task`, `clickup_update_task`
- Sync: Bi-directional with Tasks module

**Asana Requirements:**
- OAuth: Asana OAuth 2.0
- API: Tasks, projects, portfolios
- MCP: `asana_list_tasks`, `asana_create_task`, `asana_get_project`
- Sync: Bi-directional with Tasks/Projects modules

---

### MEDIUM PRIORITY - Communication

| Provider | OAuth | Direct API | MCP Tools | Sync Worker | Priority |
|----------|-------|------------|-----------|-------------|----------|
| **Microsoft Teams** | Microsoft OAuth | NEED | NEED | NEED | MEDIUM |
| **Discord** | Discord OAuth | NEED | OPTIONAL | OPTIONAL | LOW |
| **Dropbox** | Dropbox OAuth | NEED | OPTIONAL | NEED | MEDIUM |

**Microsoft Teams Requirements:**
- OAuth: Microsoft Graph API OAuth 2.0
- API: Messages, channels, calendar, presence
- MCP: `teams_send_message`, `teams_list_channels`, `teams_get_presence`
- Sync: Team member status, calendar sync

---

### MEDIUM PRIORITY - Video/Meeting

| Provider | OAuth | Direct API | MCP Tools | Sync Worker | Priority |
|----------|-------|------------|-----------|-------------|----------|
| **Zoom** | Zoom OAuth | NEED | NEED | NEED | HIGH |
| **Loom** | Loom OAuth | NEED | OPTIONAL | OPTIONAL | LOW |
| **Fireflies** | Fireflies API | NEED | OPTIONAL | NEED | MEDIUM |
| **Fathom** | Fathom API | NEED | OPTIONAL | NEED | MEDIUM |
| **TL;DV** | TL;DV API | NEED | OPTIONAL | NEED | LOW |

**Zoom Requirements:**
- OAuth: Zoom OAuth 2.0
- API: Schedule meetings, get recordings, transcripts
- MCP: `zoom_create_meeting`, `zoom_get_recording`, `zoom_list_meetings`
- Sync: Auto-create calendar events, pull transcripts

---

### LOW PRIORITY - CRM

| Provider | OAuth | Direct API | MCP Tools | Sync Worker | Priority |
|----------|-------|------------|-----------|-------------|----------|
| **GoHighLevel** | GHL OAuth | STUB | NEED | NEED | MEDIUM |
| **Salesforce** | Salesforce OAuth | NEED | OPTIONAL | NEED | LOW |
| **Pipedrive** | Pipedrive OAuth | NEED | OPTIONAL | NEED | LOW |

---

### FILE IMPORT ONLY (No OAuth)

| Provider | File Format | Parser | UI | Priority |
|----------|-------------|--------|-----|----------|
| **ChatGPT** | JSON export | NEED | NEED | HIGH |
| **Claude** | JSON export | NEED | NEED | HIGH |
| **Perplexity** | JSON export | NEED | NEED | MEDIUM |
| **Gemini** | JSON export | NEED | NEED | LOW |
| **Granola** | Meeting JSON | NEED | NEED | MEDIUM |

**File Import Requirements:**
- UI: File upload component in Knowledge module
- Backend: Parse JSON, extract conversations
- Storage: Import as memories/contexts
- No MCP tools needed (manual import only)

---

### NOTE-TAKING APPS (Future)

| Provider | OAuth | Direct API | MCP Tools | Priority |
|----------|-------|------------|-----------|----------|
| **Obsidian** | None (local) | NEED | OPTIONAL | LOW |
| **Roam Research** | Roam API | NEED | OPTIONAL | LOW |
| **Evernote** | Evernote OAuth | NEED | OPTIONAL | LOW |
| **Apple Notes** | None (iCloud) | N/A | N/A | NOT PLANNED |

---

## API vs MCP Decision Guide

### When to use Direct API (NOT MCP):
- Data sync/import (runs automatically)
- Background workers
- User authentication flows
- Bulk operations
- Bi-directional sync

### When to use MCP Tools (AI-accessible):
- User wants AI to "check my calendar"
- User wants AI to "send a Slack message"
- User wants AI to "find files in Drive"
- User wants AI to "create a task in ClickUp"
- Real-time interactions through conversation

### When to use BOTH:
- Google Calendar (sync events + AI can check/create)
- Slack (sync channels + AI can message)
- ClickUp/Asana (sync tasks + AI can manage tasks)
- HubSpot (sync contacts + AI can lookup/create)

---

## Backend API Endpoints Required

### Generic Integration Pattern
```
GET  /api/integrations/status                   # All statuses
GET  /api/integrations/{provider}/status        # Single status
POST /api/integrations/{provider}/auth          # Start OAuth
GET  /api/integrations/{provider}/callback      # OAuth callback
DELETE /api/integrations/{provider}             # Disconnect
POST /api/integrations/{provider}/sync          # Manual sync trigger
GET  /api/integrations/{provider}/logs          # Sync history
```

### Per-Provider Specific Endpoints

**Gmail:**
```
GET  /api/integrations/gmail/messages           # Search/list emails
GET  /api/integrations/gmail/threads/:id        # Get thread
GET  /api/integrations/gmail/labels             # List labels
```

**Google Drive:**
```
GET  /api/integrations/drive/files              # List files
GET  /api/integrations/drive/files/:id          # Get file content
GET  /api/integrations/drive/folders            # List folders
```

**ClickUp:**
```
GET  /api/integrations/clickup/tasks            # List tasks
POST /api/integrations/clickup/tasks            # Create task
PUT  /api/integrations/clickup/tasks/:id        # Update task
GET  /api/integrations/clickup/spaces           # List spaces
```

**HubSpot:**
```
GET  /api/integrations/hubspot/contacts         # List contacts
POST /api/integrations/hubspot/contacts         # Create contact
GET  /api/integrations/hubspot/deals            # List deals
GET  /api/integrations/hubspot/companies        # List companies
```

**Zoom:**
```
POST /api/integrations/zoom/meetings            # Schedule meeting
GET  /api/integrations/zoom/meetings            # List meetings
GET  /api/integrations/zoom/recordings/:id      # Get recording
```

---

## MCP Tools Inventory

### Currently Implemented (18 tools)

**Google Calendar (5):**
- `calendar_list_events`
- `calendar_create_event`
- `calendar_update_event`
- `calendar_delete_event`
- `calendar_sync_events`

**Slack (6):**
- `slack_list_channels`
- `slack_send_message`
- `slack_get_channel_history`
- `slack_search_messages`
- `slack_list_users`
- `slack_get_user_info`

**Notion (7):**
- `notion_list_databases`
- `notion_get_database`
- `notion_query_database`
- `notion_get_page`
- `notion_create_page`
- `notion_update_page`
- `notion_search`

### Planned MCP Tools

**Gmail (3):**
- `gmail_search` - Search emails
- `gmail_get_thread` - Get email thread
- `gmail_list_labels` - List email labels

**Google Drive (3):**
- `drive_search` - Search files
- `drive_get_file` - Get file content
- `drive_list_folder` - List folder contents

**ClickUp (4):**
- `clickup_list_tasks` - List tasks
- `clickup_create_task` - Create task
- `clickup_update_task` - Update task
- `clickup_get_space` - Get space info

**Asana (4):**
- `asana_list_tasks` - List tasks
- `asana_create_task` - Create task
- `asana_get_project` - Get project
- `asana_search` - Search workspace

**HubSpot (4):**
- `hubspot_get_contact` - Get contact
- `hubspot_list_deals` - List deals
- `hubspot_create_task` - Create task
- `hubspot_search` - Search CRM

**Zoom (3):**
- `zoom_create_meeting` - Schedule meeting
- `zoom_list_meetings` - List meetings
- `zoom_get_recording` - Get recording

**Microsoft Teams (3):**
- `teams_send_message` - Send message
- `teams_list_channels` - List channels
- `teams_get_presence` - Get user presence

---

## Implementation Priority

### Phase 1 (Immediate - Backend Focus)
1. **Gmail OAuth + API** - Extend Google OAuth
2. **Google Drive OAuth + API** - Extend Google OAuth
3. **ClickUp OAuth + API + MCP** - New integration
4. **File Import Parsers** - ChatGPT, Claude

### Phase 2 (Short-term)
5. **Asana OAuth + API + MCP** - New integration
6. **HubSpot OAuth + API + MCP** - New integration
7. **Zoom OAuth + API + MCP** - New integration

### Phase 3 (Medium-term)
8. **Linear OAuth + API** - New integration
9. **Microsoft Teams OAuth + API** - New integration
10. **Sync Workers** - Background jobs for all

### Phase 4 (Future)
11. Remaining CRM integrations
12. Note-taking app imports
13. Meeting transcription services

---

## Environment Variables Needed

```bash
# Google (shared for Calendar, Gmail, Drive)
GOOGLE_CLIENT_ID=xxx
GOOGLE_CLIENT_SECRET=xxx
GOOGLE_REDIRECT_URI=xxx

# Slack (DONE)
SLACK_CLIENT_ID=xxx
SLACK_CLIENT_SECRET=xxx
SLACK_REDIRECT_URI=xxx

# Notion (DONE)
NOTION_CLIENT_ID=xxx
NOTION_CLIENT_SECRET=xxx
NOTION_REDIRECT_URI=xxx

# ClickUp (NEW)
CLICKUP_CLIENT_ID=xxx
CLICKUP_CLIENT_SECRET=xxx
CLICKUP_REDIRECT_URI=xxx

# Asana (NEW)
ASANA_CLIENT_ID=xxx
ASANA_CLIENT_SECRET=xxx
ASANA_REDIRECT_URI=xxx

# HubSpot (NEW)
HUBSPOT_CLIENT_ID=xxx
HUBSPOT_CLIENT_SECRET=xxx
HUBSPOT_REDIRECT_URI=xxx

# Zoom (NEW)
ZOOM_CLIENT_ID=xxx
ZOOM_CLIENT_SECRET=xxx
ZOOM_REDIRECT_URI=xxx

# Microsoft/Teams (NEW)
MICROSOFT_CLIENT_ID=xxx
MICROSOFT_CLIENT_SECRET=xxx
MICROSOFT_REDIRECT_URI=xxx

# Linear (NEW)
LINEAR_CLIENT_ID=xxx
LINEAR_CLIENT_SECRET=xxx
LINEAR_REDIRECT_URI=xxx

# GoHighLevel (NEW)
GHL_CLIENT_ID=xxx
GHL_CLIENT_SECRET=xxx
GHL_REDIRECT_URI=xxx
```

---

## Summary

| Category | Count | Implemented | Planned |
|----------|-------|-------------|---------|
| OAuth Integrations | 15+ | 3 | 12+ |
| File Imports | 5 | 0 | 5 |
| MCP Tools | 50+ | 18 | 32+ |
| Sync Workers | 10+ | 1 | 9+ |
| Frontend UI | 15+ | 0 | 15+ |

**Total Effort Estimate:** 12-16 weeks for comprehensive integration suite
