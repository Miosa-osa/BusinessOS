# Integration Phases Roadmap

This document outlines the phased approach for implementing integrations in BusinessOS.

---

## Current Status

| Phase | Status | Providers |
|-------|--------|-----------|
| Phase 0 | Complete | Core infrastructure, Provider interface |
| Phase 1 | Complete | Google Calendar, Slack |
| Phase 2 | In Progress | Gmail, Communication Hub |
| Phase 3 | Planned | Microsoft 365 |
| Phase 4 | Planned | Notion, Linear |
| Phase 5 | Planned | Fathom, Analytics |
| Phase 6 | Future | GitHub, Discord, Zoom |

---

## Phase 0: Core Infrastructure (Complete)

### Delivered
- Provider interface (`integrations.Provider`)
- Provider registry with auto-registration
- OAuth utilities (state generation, code exchange)
- Credential vault with encryption
- Unified handler with Gin routes
- Database schema for integrations

### Files Created
```
/internal/integrations/
├── types.go           # Provider interface, types
├── registry.go        # Provider registry
├── oauth.go           # OAuth utilities
├── handler.go         # Unified HTTP handler
└── providers/
    ├── register.go    # Provider initialization
    ├── google/        # Google provider
    └── slack/         # Slack provider
```

---

## Phase 1: Core Productivity (Complete)

### Google Calendar
- OAuth flow with calendar scopes
- Event sync to calendar_events table
- Create/update/delete events
- Rich attendee data extraction
- Meeting link extraction

### Slack
- OAuth flow with bot + user scopes
- Channel listing and history
- Message sending (threads supported)
- User listing
- Message search (user scope)

---

## Phase 2: Communication Hub (In Progress)

### Gmail Integration
**Status**: Backend in progress

#### Backend Tasks
- [x] Gmail scopes added to Google provider
- [ ] Gmail service implementation
  - [ ] `SyncEmails()` - Full inbox sync
  - [ ] `SyncIncrementalEmails()` - New emails only
  - [ ] `GetEmail()` - Single email retrieval
  - [ ] `SendEmail()` - Compose and send
  - [ ] `ReplyEmail()` - Reply to thread
  - [ ] `MarkAsRead()` / `MarkAsUnread()`
  - [ ] `Archive()` / `Trash()`
  - [ ] `Search()` - Gmail search API
- [ ] Database migration for emails table
- [ ] Email sync job (background worker)

#### Frontend Tasks
- [x] Communication Hub layout with tabs
- [x] Email tab UI with folders, list, preview
- [x] Compose modal with reply/forward
- [ ] Connect to real Gmail API
- [ ] Thread view
- [ ] Attachment handling
- [ ] Search functionality

### Channels Integration
**Status**: Frontend placeholder ready

#### Backend Tasks
- [ ] Channels sync to database
- [ ] Messages sync with pagination
- [ ] Real-time message updates (webhooks)

#### Frontend Tasks
- [x] Channels tab placeholder
- [ ] Channel list from database
- [ ] Message view
- [ ] Send message functionality

---

## Phase 3: Microsoft 365

### Outlook Calendar
**Scopes Required**:
```
Calendars.Read
Calendars.ReadWrite
```

**Features**:
- Event sync to calendar_events table
- Create/update/delete events
- Shared calendar support
- Meeting invites

### Outlook Mail
**Scopes Required**:
```
Mail.Read
Mail.ReadWrite
Mail.Send
```

**Features**:
- Email sync to emails table
- Send/reply/forward
- Folder management
- Search

### OneDrive
**Scopes Required**:
```
Files.Read
Files.ReadWrite
```

**Features**:
- File listing
- Upload/download
- Sharing

### Microsoft Teams
**Scopes Required**:
```
Team.ReadBasic.All
Channel.ReadBasic.All
Chat.Read
```

**Features**:
- Team/channel listing
- Message sync
- Meeting integration

### Implementation Steps
1. Register Azure AD application
2. Configure Graph API permissions
3. Create Microsoft provider
4. Implement OAuth flow
5. Add service methods
6. Create frontend UI

---

## Phase 4: Productivity Tools

### Notion
**OAuth Type**: OAuth 2.0

**Scopes Required**:
```
read_content
update_content
insert_content
```

**Features**:
- Page/database listing
- Page content sync
- Create/update pages
- Database queries
- Link with BusinessOS contexts

### Linear
**OAuth Type**: OAuth 2.0

**Features**:
- Issue sync
- Project listing
- Create/update issues
- Link with BusinessOS tasks
- Status updates

### Implementation Priority
1. Notion (knowledge base sync)
2. Linear (task management sync)

---

## Phase 5: Analytics & Intelligence

### Fathom Analytics
**Auth Type**: API Key

**Features**:
- Site statistics
- Real-time visitors
- Goal tracking
- Dashboard widgets

### Google Analytics
**Auth Type**: OAuth 2.0

**Features**:
- Traffic data
- User behavior
- Conversion tracking
- Custom reports

### Implementation
- API key management in credential vault
- Dashboard widgets for analytics
- Daily/weekly reports

---

## Phase 6: Communication & Development

### Discord
**OAuth Type**: OAuth 2.0

**Features**:
- Server/channel listing
- Message sync
- Bot integration
- Webhook support

### GitHub
**OAuth Type**: OAuth 2.0

**Features**:
- Repository listing
- Issue/PR sync
- Webhook integration
- Code search
- Link with BusinessOS projects

### Zoom
**OAuth Type**: OAuth 2.0

**Features**:
- Meeting scheduling
- Recording access
- Participant management
- Calendar integration

### Loom
**OAuth Type**: OAuth 2.0

**Features**:
- Video listing
- Embed support
- Sharing

---

## Database Schema Extensions

### Phase 2: Communication Hub

```sql
-- Emails table (Gmail, Outlook)
CREATE TABLE emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    thread_id VARCHAR(255),
    subject TEXT,
    snippet TEXT,
    from_email VARCHAR(255),
    from_name VARCHAR(255),
    to_emails JSONB DEFAULT '[]',
    body_text TEXT,
    body_html TEXT,
    attachments JSONB DEFAULT '[]',
    is_read BOOLEAN DEFAULT FALSE,
    is_starred BOOLEAN DEFAULT FALSE,
    labels JSONB DEFAULT '[]',
    date TIMESTAMP WITH TIME ZONE,
    synced_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, provider, external_id)
);

-- Channels table (Slack, Discord, Teams)
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_private BOOLEAN DEFAULT FALSE,
    member_count INT DEFAULT 0,
    last_message_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, provider, external_id)
);

-- Channel messages
CREATE TABLE channel_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID REFERENCES channels(id),
    external_id VARCHAR(255) NOT NULL,
    sender_name VARCHAR(255),
    content TEXT,
    attachments JSONB DEFAULT '[]',
    sent_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(channel_id, external_id)
);
```

### Phase 4: Productivity

```sql
-- External documents (Notion, Google Docs)
CREATE TABLE external_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    title TEXT,
    content TEXT,
    url TEXT,
    parent_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    synced_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, provider, external_id)
);

-- External issues (Linear, GitHub)
CREATE TABLE external_issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    title TEXT,
    description TEXT,
    status VARCHAR(100),
    priority VARCHAR(50),
    assignee VARCHAR(255),
    labels JSONB DEFAULT '[]',
    url TEXT,
    synced_at TIMESTAMP WITH TIME ZONE,
    businessos_task_id UUID REFERENCES tasks(id),
    UNIQUE(user_id, provider, external_id)
);
```

---

## API Endpoints

### Current (Phase 0-1)

```
GET  /api/integrations/providers              # List all providers
GET  /api/integrations/providers/:name        # Get provider details
GET  /api/integrations/oauth/:provider/start  # Start OAuth flow
GET  /api/integrations/oauth/:provider/callback # OAuth callback
GET  /api/integrations                        # List user integrations
GET  /api/integrations/:provider              # Get integration status
DELETE /api/integrations/:provider            # Disconnect
POST /api/integrations/:provider/sync         # Trigger sync
GET  /api/integrations/:provider/status       # Sync status
```

### Phase 2 Additions

```
# Gmail
GET  /api/communication/emails                # List emails
GET  /api/communication/emails/:id            # Get email
POST /api/communication/emails                # Send email
POST /api/communication/emails/:id/reply      # Reply to email
PUT  /api/communication/emails/:id/read       # Mark as read
DELETE /api/communication/emails/:id          # Archive/delete

# Channels
GET  /api/communication/channels              # List channels
GET  /api/communication/channels/:id          # Get channel
GET  /api/communication/channels/:id/messages # Get messages
POST /api/communication/channels/:id/messages # Send message
```

---

## Frontend Routes

### Current
```
/communication/calendar    # Calendar tab
/communication/email       # Email tab
/communication/channels    # Channels tab
/integrations              # Integration settings
/integrations/:id          # Provider settings
```

### Phase 2 Additions
```
/communication/email/:threadId     # Email thread view
/communication/email/compose       # Compose new email
/communication/channels/:channelId # Channel messages
```

---

## Skills Integration

Each integration provides skills for the AI agent:

### Google Calendar Skills
- `google_calendar.get_events` - Get calendar events
- `google_calendar.create_event` - Create new event
- `google_calendar.sync_daily_log` - Sync to daily log

### Gmail Skills
- `gmail.get_emails` - Get recent emails
- `gmail.send_email` - Send email
- `gmail.search` - Search emails

### Slack Skills
- `slack.send_message` - Send message
- `slack.message_to_task` - Convert message to task
- `slack.create_channel` - Create channel

### Notion Skills (Future)
- `notion.search` - Search pages
- `notion.create_page` - Create page
- `notion.update_page` - Update page

---

## Success Metrics

### Phase 1 (Complete)
- Users can connect Google Calendar
- Events sync within 5 seconds
- Calendar displays real events

### Phase 2 (In Progress)
- Gmail inbox displays emails
- Users can send/reply to emails
- Slack channels show real messages

### Phase 3
- Microsoft 365 OAuth works
- Outlook calendar syncs
- Outlook mail works

### Phase 4
- Notion pages sync
- Linear issues sync
- Bidirectional sync works

### Phase 5
- Analytics dashboard widgets
- Fathom data displays
- Daily reports work

---

## Timeline Estimates

| Phase | Scope | Team |
|-------|-------|------|
| Phase 2 | Gmail + Channels completion | Backend + Frontend |
| Phase 3 | Microsoft 365 | Backend focus |
| Phase 4 | Notion + Linear | Full stack |
| Phase 5 | Analytics | Backend + Dashboard |
| Phase 6 | Discord + GitHub + Zoom | Full stack |

---

## Risk Mitigation

### API Rate Limits
- Implement exponential backoff
- Cache frequently accessed data
- Use webhooks where available

### Token Expiration
- Automatic token refresh
- Graceful degradation on failure
- User notification for re-auth

### Data Consistency
- Sync conflict resolution
- Idempotent sync operations
- Audit logging

### Security
- Encrypt all tokens at rest
- Minimal scope requests
- Regular security audits
