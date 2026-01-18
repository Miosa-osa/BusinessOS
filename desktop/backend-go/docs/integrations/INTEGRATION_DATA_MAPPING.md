# Integration Data Mapping & Foundation Status

**Document Version:** 1.0
**Created:** 2026-01-05
**Purpose:** Complete reference of all data available from each integration and current implementation status

---

## Executive Summary

This document provides:
1. Complete data mapping for all integration APIs
2. Current implementation status per provider
3. Foundation gaps requiring attention
4. BusinessOS module mapping for imported data

---

## Foundation Status Overview

### Core Infrastructure

| Component | Status | Location | Notes |
|-----------|--------|----------|-------|
| Provider Interface | ✅ Complete | `integrations/types.go` | Full Provider interface defined |
| Provider Registry | ✅ Complete | `integrations/registry.go` | Thread-safe registration |
| OAuth Utilities | ✅ Complete | `integrations/oauth.go` | State generation, code exchange, refresh |
| Encryption Service | ✅ Complete | `security/encryption.go` | AES-256-GCM, global singleton |
| Credential Vault | ✅ Complete | `services/credential_vault.go` | Store/retrieve/refresh OAuth & API keys |
| Unified Handler | ⚠️ Partial | `integrations/handler.go` | Routes defined, needs user context wiring |
| Error Handling | ✅ Complete | `integrations/errors.go` | Standard error types |

### Database Migrations

| Migration | Status | Tables Created |
|-----------|--------|----------------|
| 025_integrations_module | ✅ Applied | integration_providers, user_integrations, module_integration_settings |
| 026_oauth_tokens_tables | ✅ Applied | google_oauth_tokens, slack_oauth_tokens, notion_oauth_tokens |
| 027_credential_vault | ✅ Applied | credential_vault, integration_webhooks, data_sync_mappings |
| 028_data_imports | ✅ Applied | import_jobs, imported_records, import_mapping_templates, imported_conversations |
| 029_calendar_unique_constraint | ✅ Applied | calendar_events unique constraint |
| 030_emails_and_channels | ✅ Applied | emails, channels, channel_messages, integration_sync_log |
| 031_notion_tables | ✅ Applied | notion_pages, notion_databases |

---

## Provider Implementation Status

| Provider | OAuth | Credentials | Data Sync | MCP Tools | Priority |
|----------|-------|-------------|-----------|-----------|----------|
| Google Calendar | ✅ | ✅ | ✅ | ✅ | P0 |
| Gmail | ✅ | ✅ | ⚠️ Partial | ✅ | P0 |
| Slack | ✅ | ✅ | ⚠️ Partial | ✅ | P0 |
| Notion | ✅ | ✅ | ⚠️ Partial | ✅ | P0 |
| HubSpot | ❌ | ❌ | ❌ | ❌ | P1 |
| Linear | ❌ | ❌ | ❌ | ❌ | P1 |
| Microsoft 365 | ❌ | ❌ | ❌ | ❌ | P2 |
| GitHub | ❌ | ❌ | ❌ | ❌ | P2 |
| Stripe | ❌ | ❌ | ❌ | ❌ | P3 |
| Zoom | ❌ | ❌ | ❌ | ❌ | P3 |

---

## Detailed Data Mapping by Provider

### 1. Google Calendar

**OAuth Scopes:**
```
https://www.googleapis.com/auth/calendar.readonly
https://www.googleapis.com/auth/calendar.events
https://www.googleapis.com/auth/calendar
```

**Available Data:**

| Entity | Fields | BusinessOS Target |
|--------|--------|-------------------|
| Event | id, summary, description, location, start, end, attendees, creator, organizer, status, htmlLink, hangoutLink, conferenceData, recurrence, reminders, colorId, visibility, guestsCanModify, guestsCanInviteOthers, guestsCanSeeOtherGuests | calendar_events |
| Calendar | id, summary, description, location, timeZone, colorId, backgroundColor, foregroundColor, accessRole, primary | user_calendars |
| Attendee | email, displayName, responseStatus, optional, resource, comment | calendar_attendees |
| Reminder | method, minutes | calendar_event_reminders |

**Current Implementation:**
- ✅ Event listing (FetchEvents)
- ✅ Event creation (CreateGoogleEvent)
- ✅ Event update (UpdateGoogleEvent)
- ✅ Sync to calendar_events table
- ⚠️ Need: Recurring event expansion
- ⚠️ Need: Webhook for real-time updates

---

### 2. Gmail

**OAuth Scopes:**
```
https://www.googleapis.com/auth/gmail.readonly
https://www.googleapis.com/auth/gmail.send
https://www.googleapis.com/auth/gmail.modify
https://www.googleapis.com/auth/gmail.labels
```

**Available Data:**

| Entity | Fields | BusinessOS Target |
|--------|--------|-------------------|
| Message | id, threadId, labelIds, snippet, historyId, internalDate, payload (headers, parts, body), sizeEstimate, raw | emails |
| Thread | id, historyId, messages[] | email_threads |
| Label | id, name, messageListVisibility, labelListVisibility, type, color | email_labels |
| Draft | id, message | email_drafts |
| Header | From, To, Cc, Bcc, Subject, Date, Message-ID, In-Reply-To, References | (parsed into emails) |
| Attachment | attachmentId, size, data, filename, mimeType | email_attachments |

**Current Implementation:**
- ✅ Gmail provider registered
- ✅ OAuth flow working
- ⚠️ Need: Full message sync
- ⚠️ Need: Thread reconstruction
- ⚠️ Need: Send/Reply functionality
- ⚠️ Need: Label management
- ⚠️ Need: Search functionality

---

### 3. Slack

**OAuth Scopes (Bot):**
```
channels:read
channels:history
chat:write
users:read
files:read
reactions:read
```

**OAuth Scopes (User):**
```
search:read
```

**Available Data:**

| Entity | Fields | BusinessOS Target |
|--------|--------|-------------------|
| Channel | id, name, is_channel, is_private, is_archived, is_member, num_members, topic, purpose, created, creator | channels |
| Message | type, ts, user, text, thread_ts, reply_count, reactions[], attachments[], files[], blocks[] | channel_messages |
| User | id, team_id, name, real_name, profile (email, image, title, phone), is_admin, is_owner, is_bot | slack_users |
| File | id, name, title, mimetype, filetype, size, url_private, thumb_* | slack_files |
| Reaction | name, count, users[] | message_reactions |
| Thread | thread_ts, reply_count, reply_users[], latest_reply | (part of messages) |

**Current Implementation:**
- ✅ Slack provider registered
- ✅ OAuth flow with bot + user tokens
- ✅ Channel listing
- ✅ Message history
- ✅ Send message (with threads)
- ✅ User listing
- ✅ Search messages (user scope)
- ⚠️ Need: Webhook for real-time messages
- ⚠️ Need: File upload/download
- ⚠️ Need: Reaction management

---

### 4. Notion

**OAuth Scopes:**
```
read_content
update_content
insert_content
```

**Available Data:**

| Entity | Fields | BusinessOS Target |
|--------|--------|-------------------|
| Page | id, created_time, last_edited_time, created_by, last_edited_by, cover, icon, parent, archived, url, properties{} | notion_pages |
| Database | id, created_time, last_edited_time, title[], description[], icon, cover, properties{}, parent, url, is_inline | notion_databases |
| Block | id, type, created_time, last_edited_time, has_children, archived, [type-specific content] | notion_blocks |
| User | id, type, name, avatar_url, person{email}, bot{} | notion_users |
| Property Types | title, rich_text, number, select, multi_select, date, people, files, checkbox, url, email, phone_number, formula, relation, rollup, created_time, created_by, last_edited_time, last_edited_by, status | notion_properties |

**Current Implementation:**
- ✅ Notion provider registered
- ✅ OAuth flow working
- ✅ Database listing
- ✅ Database query
- ✅ Page retrieval
- ✅ Page creation
- ✅ Search
- ⚠️ Need: Block content extraction
- ⚠️ Need: Full page sync to contexts
- ⚠️ Need: Database to BusinessOS mapping

---

### 5. HubSpot (NOT IMPLEMENTED)

**OAuth Scopes:**
```
crm.objects.contacts.read
crm.objects.contacts.write
crm.objects.companies.read
crm.objects.companies.write
crm.objects.deals.read
crm.objects.deals.write
crm.objects.owners.read
sales-email-read
```

**Available Data:**

| Entity | Fields | BusinessOS Target |
|--------|--------|-------------------|
| Contact | id, properties (firstname, lastname, email, phone, company, jobtitle, lifecyclestage, hs_lead_status), associations | clients |
| Company | id, properties (name, domain, industry, phone, city, state, country, numberofemployees, annualrevenue) | companies |
| Deal | id, properties (dealname, amount, dealstage, pipeline, closedate, hs_deal_stage_probability), associations | deals |
| Activity | id, type (EMAIL, CALL, MEETING, TASK, NOTE), properties, associations | activities |
| Owner | id, email, firstName, lastName, userId | team_members |
| Pipeline | id, label, displayOrder, stages[] | pipelines |
| Stage | id, label, displayOrder, metadata | pipeline_stages |

**Required Implementation:**
- [ ] HubSpot provider (`integrations/hubspot/`)
- [ ] OAuth flow
- [ ] Contact sync to clients
- [ ] Company sync
- [ ] Deal pipeline sync
- [ ] Activity timeline
- [ ] Import from HubSpot

---

### 6. Linear (NOT IMPLEMENTED)

**Auth Type:** OAuth 2.0 + GraphQL

**Available Data:**

| Entity | Fields | BusinessOS Target |
|--------|--------|-------------------|
| Issue | id, identifier, title, description, priority, state, assignee, creator, labels[], project, cycle, estimate, dueDate, createdAt, updatedAt, completedAt | tasks |
| Project | id, name, description, state, lead, members[], startDate, targetDate, progress, issues[] | projects |
| Cycle | id, name, number, startsAt, endsAt, completedAt, progress, issues[] | sprints |
| Team | id, name, key, description, members[], states[], labels[] | teams |
| User | id, name, displayName, email, avatarUrl, active | team_members |
| Label | id, name, color, description | labels |
| State | id, name, color, type (triage, backlog, unstarted, started, completed, canceled) | task_statuses |
| Comment | id, body, user, createdAt, updatedAt | task_comments |
| Attachment | id, title, url, source | task_attachments |

**Required Implementation:**
- [ ] Linear provider (`integrations/linear/`)
- [ ] GraphQL client setup
- [ ] Issue sync to tasks (bidirectional)
- [ ] Project sync
- [ ] Cycle tracking
- [ ] Webhook for real-time updates

---

### 7. Microsoft 365 (NOT IMPLEMENTED)

**Auth Type:** OAuth 2.0 via Azure AD

**OAuth Scopes:**
```
User.Read
Mail.Read
Mail.ReadWrite
Mail.Send
Calendars.Read
Calendars.ReadWrite
Files.Read
Files.ReadWrite
Team.ReadBasic.All
Channel.ReadBasic.All
Chat.Read
```

**Available Data:**

| Entity | Fields | BusinessOS Target |
|--------|--------|-------------------|
| Message (Mail) | id, subject, bodyPreview, body, from, toRecipients, ccRecipients, bccRecipients, receivedDateTime, hasAttachments, isRead, flag, importance | emails |
| Event (Calendar) | id, subject, body, start, end, location, attendees[], organizer, isAllDay, recurrence, showAs, importance | calendar_events |
| DriveItem | id, name, size, webUrl, createdDateTime, lastModifiedDateTime, file, folder, parentReference | documents |
| Channel | id, displayName, description, membershipType | channels |
| ChatMessage | id, body, from, createdDateTime, attachments[] | channel_messages |
| User | id, displayName, mail, jobTitle, department | team_members |

**Required Implementation:**
- [ ] Microsoft provider (`integrations/microsoft/`)
- [ ] Azure AD OAuth setup
- [ ] Outlook mail sync
- [ ] Outlook calendar sync
- [ ] OneDrive file access
- [ ] Teams integration

---

## Foundation Gaps & Required Work

### Critical Gaps (Must Fix)

#### 1. Unified Handler User Context
**File:** `integrations/handler.go`
**Issue:** Handler has TODO comments for getting userID from auth context
**Fix Required:**
```go
// Current (broken):
// userID := getUserID(c)

// Needed:
func getUserID(c *gin.Context) string {
    userID, exists := c.Get("user_id")
    if !exists {
        return ""
    }
    return userID.(string)
}
```

#### 2. Import Engine Service
**Status:** Database schema exists, service not implemented
**Files Needed:**
- `services/import_engine.go` - Core import logic
- `services/importers/` - Provider-specific importers

#### 3. Sync Scheduler
**Status:** Not implemented
**Required:** Background job to process scheduled syncs
**Files Needed:**
- `services/sync_scheduler.go`
- Integration with credential vault for auto-refresh

#### 4. Webhook Handlers
**Status:** Table exists, no handlers
**Required:** Receive and process webhooks from:
- Google (Calendar push notifications)
- Slack (Events API)
- Notion (coming soon)
- HubSpot
- Linear

### Provider Gaps

#### Google
- [ ] Gmail full sync implementation
- [ ] Gmail send/reply
- [ ] Calendar webhook registration
- [ ] Drive file listing

#### Slack
- [ ] Events API webhook handler
- [ ] File upload/download
- [ ] Channel sync to database

#### Notion
- [ ] Block content extraction
- [ ] Full page sync
- [ ] Webhook when available

#### New Providers Needed
- [ ] HubSpot (P1)
- [ ] Linear (P1)
- [ ] Microsoft 365 (P2)
- [ ] GitHub (P2)

---

## Data Flow Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DATA FLOW                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. OAuth Connection                                                        │
│     User → Frontend → /api/integrations/oauth/:provider/start               │
│                    → Provider Auth URL                                      │
│                    → Callback with code                                     │
│                    → Exchange code → Store in credential_vault              │
│                                                                             │
│  2. Initial Data Import                                                     │
│     Backend → Get credential from vault                                     │
│            → Call provider API                                              │
│            → Transform data (field mappings)                                │
│            → Store in BusinessOS tables                                     │
│            → Track in imported_records                                      │
│                                                                             │
│  3. Ongoing Sync                                                            │
│     Scheduler → Check user_integrations.sync_frequency                      │
│              → Get credential (auto-refresh if needed)                      │
│              → Fetch changes since last_synced_at                           │
│              → Apply changes to BusinessOS                                  │
│              → Log in integration_sync_log                                  │
│                                                                             │
│  4. Real-time Updates (Webhooks)                                            │
│     Provider → POST /api/webhooks/:provider                                 │
│             → Verify signature                                              │
│             → Process event                                                 │
│             → Update BusinessOS                                             │
│                                                                             │
│  5. Sorx Skill Execution                                                    │
│     Skill → GetCredential(userID, providerID)                               │
│          → Execute API call                                                 │
│          → Return result                                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Next Steps Priority

### Phase 2A: Complete Existing Providers (This Week)
1. Wire up handler.go with user context
2. Implement Gmail sync service
3. Implement Slack channel sync
4. Add webhook handlers for Google Calendar

### Phase 2B: Import Engine (Next Week)
1. Create import_engine.go service
2. Build ChatGPT/Claude conversation importer
3. Build HubSpot contact importer
4. Create field mapping UI

### Phase 3: New Providers (Following Weeks)
1. HubSpot provider (OAuth + full CRM sync)
2. Linear provider (GraphQL + bidirectional sync)
3. Microsoft 365 provider

---

## References

- [Google Calendar API](https://developers.google.com/workspace/calendar/api/v3/reference)
- [Gmail API](https://developers.google.com/gmail/api)
- [Slack API Methods](https://api.slack.com/methods)
- [Slack API Types](https://api.slack.com/types)
- [Notion API](https://developers.notion.com/reference/page-property-values)
- [HubSpot CRM API](https://developers.hubspot.com/docs/api-reference/crm-contacts-v3/guide)
- [Linear API](https://linear.app/developers/graphql)
- [Microsoft Graph API](https://learn.microsoft.com/en-us/graph/api/resources/calendar-overview)
