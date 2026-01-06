# Google APIs Research for BusinessOS

**Research Date:** January 5, 2026
**Scope:** Google Calendar API v3, Gmail API v1, Google Drive API v3

---

## Executive Summary

This research documents all data available from Google APIs relevant to a business operating system. Focus areas:

1. **Google Calendar API** - Event scheduling, attendee management, meeting links
2. **Gmail API** - Email communication, search, attachments
3. **Google Drive API** - File storage, permissions, revision history

All APIs use OAuth 2.0 authentication and share common patterns for rate limiting, pagination, and error handling.

---

## Google Calendar API v3

### Key Capabilities

- **Event Sync**: Fetch all events from calendar(s) with full metadata
- **Event Management**: Create, update, delete events (bidirectional)
- **Attendee Tracking**: See who's invited, RSVP status, optional/required
- **Meeting Links**: Extract Google Meet, Zoom, Teams conference URLs
- **Recurring Events**: Handle repeating events with expansion
- **Reminders**: Email, popup, SMS notifications
- **Multi-calendar**: Sync from multiple calendars (currently only primary)

### Data Fields Available

| Field | Type | Example | Use in BusinessOS |
|-------|------|---------|-------------------|
| Event ID | String | `abc123def456` | Primary key, idempotency |
| Summary | String | "Team standup" | Event title, node name |
| Description | String | "Discuss Q4 goals" | Event details |
| Location | String | "Conference room A" | Physical or virtual location |
| Start/End Times | RFC3339 | `2023-12-15T10:00:00-05:00` | Calendar display, scheduling |
| All-Day | Boolean | `true` | UI rendering, filters |
| Status | Enum | confirmed/tentative/cancelled | Event state tracking |
| Attendees | Array | Email, name, RSVP status | Team coordination |
| Organizer | Object | Email, name | Meeting owner identification |
| Recurrence | RRULE | `FREQ=DAILY;COUNT=5` | Recurring patterns |
| Meeting Link | String | `https://meet.google.com/...` | Direct launch link |
| Conference Type | String | Google Meet/Zoom/Teams | Which platform |
| Color ID | String | 1-11 | Visual categorization |
| Reminders | Array | Method + minutes | Notification setup |
| Attachments | Array | File metadata | Related documents |
| Event Type | Enum | working_location/focus_time/out_of_office | Availability signals |
| Created/Updated | RFC3339 | Timestamp | Sync tracking |

### Required Scopes

```
https://www.googleapis.com/auth/calendar
https://www.googleapis.com/auth/calendar.readonly
```

### Optional Scopes

```
https://www.googleapis.com/auth/calendar.events
https://www.googleapis.com/auth/calendar.events.readonly
https://www.googleapis.com/auth/calendar.settings.readonly
```

### Rate Limits & Quotas

- **Daily Quota**: 1,000,000 queries per day
- **Per-Second**: No strict limit, fair use applies
- **Pagination**: maxResults up to 2,500 per request
- **Typical Batch**: 250 events per request is optimal

### Sync Strategy

**Initial Sync (5-10 minutes)**
- Fetch all accessible calendars
- Query events from 6 months ago to +6 months future
- Use `singleEvents=true` to expand recurring events
- Paginate through results if > 250 events

**Incremental Sync (Daily)**
- Use `updatedMin` parameter to get changed events
- Store `updated` timestamp for next sync
- Handle recurring event instances properly
- Optional: Setup Watch API for push-based sync

### Current Implementation Status

**File**: `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/integrations/google/calendar.go`

**Implemented Features**:
- ✅ Sync events within date range
- ✅ Create local events and push to Google
- ✅ Delete events (bidirectional)
- ✅ Get events from database
- ✅ Fetch native Google event types (for MCP)
- ✅ Create/update with full Google API support
- ✅ Conference data support (Google Meet links)

**Missing Features**:
- ❌ Event update via UI
- ❌ Multi-calendar sync (only primary)
- ❌ Full attendee management UI
- ❌ Recurring event expansion details
- ❌ Watch API (push-based sync)
- ❌ Availability/free-busy checking

### BusinessOS Integration Points

| Module | Usage |
|--------|-------|
| **Calendar** | Primary - display all events |
| **Daily Log** | Auto-populate with today's events |
| **Tasks** | Convert all-day events to tasks if needed |
| **Nodes** | Each event becomes a node |
| **Estimated Nodes** | 20-100 per user per initial sync |

---

## Gmail API v1

### Key Capabilities

- **Email Sync**: Fetch incoming/outgoing messages with full content
- **Full-Text Search**: Powerful query syntax for filtering
- **Thread Management**: Track conversation threads
- **Label Operations**: Organize with Gmail labels
- **Attachment Handling**: Extract file metadata and content
- **Send Email**: Send via Gmail (not SMTP)
- **State Tracking**: Mark as read, starred, important, archived

### Data Fields Available

| Field | Type | Example | Use in BusinessOS |
|-------|------|---------|-------------------|
| Message ID | String | `abc123def456` | Primary key, immutable |
| Thread ID | String | `xyz789` | Conversation grouping |
| Subject | String | "Project update" | Email title, search |
| From | Email + Name | `sender@example.com` | Sender identification |
| To/Cc | Array | Multiple emails | Recipients tracking |
| Date | RFC3339 | Timestamp | Timeline, sorting |
| Body Text | String | Plain text version | Primary content |
| Body HTML | String | HTML version | Rich formatting |
| Snippet | String | Preview (100 chars) | Search result display |
| Labels | Array | INBOX, SENT, DRAFT, etc. | Organization, filtering |
| Is Read | Boolean | `true` | Inbox state |
| Is Starred | Boolean | `true` | User tagging |
| Is Important | Boolean | `true` | Auto-detected importance |
| Is Draft | Boolean | `false` | Editing state |
| Is Sent | Boolean | `true` | Outbound tracking |
| Is Archived | Boolean | `false` | Inbox management |
| Is Trash | Boolean | `false` | Deletion state |
| Size | Integer | 45000 | Storage tracking |
| Attachments | Array | Filename, MIME, size, ID | File tracking |
| Reply-To | String | Email | Response routing |

### Built-in Gmail Labels (System)

```
INBOX, SENT, DRAFT, STARRED, IMPORTANT, UNREAD, TRASH, SPAM
CATEGORY_PERSONAL, CATEGORY_SOCIAL, CATEGORY_PROMOTIONS, CATEGORY_UPDATES, CATEGORY_FORUMS
+ Custom user-created labels
```

### Powerful Search Syntax

```
in:inbox                  // Filter by folder
in:sent
in:draft
is:unread                 // Filter by state
is:starred
is:important
from:example@example.com  // Filter by sender
to:example@example.com    // Filter by recipient
subject:important         // Search subject line
has:attachment            // Has files
before:2023-01-01         // Date range
after:2023-01-01
larger:1000000            // File size
filename:pdf
```

### Required Scopes

```
https://www.googleapis.com/auth/gmail.readonly
https://www.googleapis.com/auth/gmail.modify
```

### Optional Scopes

```
https://www.googleapis.com/auth/gmail.send
https://www.googleapis.com/auth/gmail.compose
https://www.googleapis.com/auth/gmail.insert
https://www.googleapis.com/auth/gmail.labels
https://www.googleapis.com/auth/gmail.settings.basic
https://www.googleapis.com/auth/gmail.settings.sharing
```

### Rate Limits & Quotas

- **Daily Quota**: 1,000,000 requests per day
- **Per-Second**: 250 requests per second per user
- **Per Request**: 1 quota unit per call
- **Batch Operations**: Up to 1,000 messages per batch
- **Max Results**: 500 per request for messages/threads

### Sync Strategy

**Initial Sync (15-30 minutes for 2000+ emails)**
- Query with `q="in:inbox OR in:sent"` (exclude drafts/spam)
- Fetch in batches of 100-500 using pagination
- For each message, get full details (format=full)
- Download attachments only when needed

**Incremental Sync (Daily)**
- Track using `internalDate` field
- Query newer emails: `q="after:2024-12-31"`
- Update only changed emails
- Challenge: Need full fetch for content updates

**Attachment Handling**
- Store metadata only in DB
- Download on-demand using attachment ID
- Check `sizeEstimate` before downloading large files
- Store in cloud storage (GCS) or reference external

### Current Implementation Status

**File**: `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/integrations/google/gmail.go`

**Implemented Features**:
- ✅ Sync emails with full content
- ✅ Parse email headers (subject, from, to, cc, reply-to, date)
- ✅ Extract text and HTML body
- ✅ Parse and store attachments
- ✅ Map Gmail labels to status flags
- ✅ Get emails with folder filtering
- ✅ Mark as read
- ✅ Archive emails
- ✅ Move to trash
- ✅ Send emails
- ✅ Handle date parsing in multiple formats

**Missing Features**:
- ❌ Watch API (push notifications)
- ❌ Starred/important email management UI
- ❌ Label creation/organization
- ❌ Search UI integration
- ❌ Draft management
- ❌ Thread operations
- ❌ Batch operations

### BusinessOS Integration Points

| Module | Usage |
|--------|-------|
| **Communication** | Primary - display inbox, send |
| **Daily Log** | Auto-populate with key emails (starred, important) |
| **Chat** | Store conversations as memories |
| **Tasks** | Extract action items from emails |
| **Nodes** | Important emails become nodes |
| **Estimated Nodes** | 50-200 per user per initial sync |

---

## Google Drive API v3

### Key Capabilities

- **File Listing**: Browse Drive hierarchy
- **Metadata Retrieval**: Access file properties without downloading
- **Search**: Powerful query for finding files
- **Permissions**: Understand file sharing
- **Revisions**: Track file version history
- **Comments**: Access Docs/Sheets/Slides comments
- **Web Links**: Get shareable links and open-in-Drive URLs

### Data Fields Available

| Field | Type | Example | Use in BusinessOS |
|-------|------|---------|-------------------|
| File ID | String | `abc123def456` | Primary key, unique |
| Name | String | "Q4 Planning.docx" | Display name, search |
| MIME Type | String | `application/vnd.google-apps.document` | File type |
| Description | String | "Planning for Q4" | User notes |
| Size | String | "45000" | Storage tracking |
| Created | RFC3339 | Timestamp | Timeline |
| Modified | RFC3339 | Timestamp | Change tracking |
| Modified By | Object | User email/name | Last editor |
| Owners | Array | User objects | Access control |
| Parents | Array | Folder IDs | Hierarchy |
| Is Starred | Boolean | `true` | User tagging |
| Is Trashed | Boolean | `false` | Deletion state |
| Is Shared | Boolean | `true` | Sharing status |
| Owned By Me | Boolean | `true` | Ownership |
| Capabilities | Object | canEdit, canDelete, canShare | User permissions |
| Permissions | Array | Users, groups, domains | Sharing details |
| Web View Link | String | URL to Drive UI | Open in Drive |
| Web Content Link | String | Download URL | Direct access |
| Thumbnail | String/URL | Image data | Preview |
| Folder Color | String | Hex color | Visual organization |
| Custom Properties | Object | User-defined | Metadata |
| Checksum | String | MD5 hash | Integrity verification |

### Common MIME Types

```
application/vnd.google-apps.document      // Google Docs
application/vnd.google-apps.spreadsheet   // Google Sheets
application/vnd.google-apps.presentation  // Google Slides
application/vnd.google-apps.folder        // Folders
application/pdf                           // PDFs
image/jpeg, image/png                     // Images
text/plain, text/csv                      // Text files
```

### Powerful Search Syntax

```
name contains 'project'
mimeType='application/vnd.google-apps.document'
trashed=false
starred=true
'folder_id' in parents
createdTime>'2023-01-01T00:00:00'
modifiedTime>'2023-12-01T00:00:00'
ownedByMe=true
```

### Required Scopes

```
https://www.googleapis.com/auth/drive.readonly
https://www.googleapis.com/auth/drive
```

### Optional Scopes

```
https://www.googleapis.com/auth/drive.file
https://www.googleapis.com/auth/drive.appdata
https://www.googleapis.com/auth/drive.photos.readonly
https://www.googleapis.com/auth/drive.metadata
```

### Rate Limits & Quotas

- **Daily Quota**: 1,000,000 queries per day
- **Quota Units**: 1 per read, 10 per write operation
- **Storage**: 15 GB shared across all Google services
- **Batch Requests**: Not supported in v3 (use batch endpoint)

### Sync Strategy

**Initial Sync (20-40 minutes for large Drive)**
- Start from root with tree traversal
- Use 'parents' field to build folder hierarchy
- Paginate through files (max 1000 per request)
- Request only needed fields via 'fields' parameter
- Don't download content initially

**Incremental Sync (Daily)**
- Use `modifiedTime > lastSync` for delta sync
- Optional: Setup Watch API for push notifications
- Challenge: No built-in change tracking
- Approach: Periodic full scan with timestamp comparison

**File Indexing**
- Index metadata only initially
- Download content on-demand for search
- For Docs/Sheets/Slides: Parse text content
- For PDFs: Use OCR service if needed
- Cache file metadata, lazy-load content

### Current Implementation Status

**Status**: Not yet implemented in codebase (planning phase)

### BusinessOS Integration Points

| Module | Usage |
|--------|-------|
| **Storage** | Primary - file discovery and search |
| **Contexts** | Files become knowledge nodes |
| **Projects** | Organize by Drive folders |
| **Artifacts** | Attach to tasks and projects |
| **Estimated Nodes** | 50-300 per user per initial sync |

---

## Shared Integration Patterns

### OAuth 2.0 Flow (All APIs)

1. **Redirect User** to Google login with scopes
2. **Exchange Code** for access token + refresh token
3. **Store Encrypted**: tokens in `google_oauth_tokens` table
4. **Auto-Refresh**: Google SDK handles automatically
5. **Use 'me'**: Authenticated user in API calls

### Pagination Pattern

All APIs use cursor-based pagination:

```
1. Initial request: GET /endpoint?maxResults=100
2. Response includes: nextPageToken (if more results)
3. Next request: GET /endpoint?pageToken=xyz&maxResults=100
4. Repeat until no nextPageToken
```

### Timestamp Format

All timestamps use RFC3339 (ISO 8601):
```
2023-12-15T10:30:00-05:00
             ↑ Timezone included
```

### Error Handling Strategy

| Error | Action |
|-------|--------|
| 401 Unauthorized | Invalid/expired token - force re-auth |
| 403 Forbidden | Insufficient scopes - request new scopes |
| 404 Not Found | Resource deleted - skip or mark for review |
| 429 Too Many Requests | Rate limit hit - exponential backoff |
| 500+ Server Error | Transient - retry with exponential backoff |

### Rate Limiting Strategy

- **Track quota** across all API calls
- **Implement exponential backoff** for retries
- **Batch operations** where available
- **Defer non-critical syncs** if quota running low
- **Cache metadata** to reduce API calls

---

## Cross-API BusinessOS Integration

### Data Flow Architecture

```
Google APIs
    ↓
OAuth Token Management (encrypted storage)
    ↓
Integration Services (calendar.go, gmail.go, drive.go)
    ↓
PostgreSQL Database
    ├── calendar_events
    ├── emails
    ├── drive_files (future)
    └── integration_sync_log
    ↓
BusinessOS Modules
    ├── Calendar (events display)
    ├── Communication (emails)
    ├── Daily Log (populate with activities)
    ├── Tasks (convert events/emails to tasks)
    ├── Contexts (documents become knowledge)
    └── Nodes (everything becomes nodes)
```

### Sync Frequency Recommendations

| Source | Frequency | Latency | Priority |
|--------|-----------|---------|----------|
| Calendar | Daily or every 6 hours | < 1 hour acceptable | High |
| Gmail | Daily | < 4 hours acceptable | High |
| Google Drive | Daily or every 12 hours | < 4 hours acceptable | Medium |

### Node Generation Strategy

1. **Calendar Events** → 1 node per event, link to attendees
2. **Email Threads** → 1 node per thread, extract action items → tasks
3. **Drive Files** → 1 node per file, indexed for full-text search
4. **Relationships** → Connect across sources (email mentions file, etc.)

### Storage Optimization

| Source | Strategy |
|--------|----------|
| **Calendar** | Store full events, deduplicate by googleId |
| **Email** | Metadata only initially, lazy-load attachments |
| **Drive** | Index metadata, lazy-load file content |

### Privacy & Security

- **Scope Minimization**: Only request scopes user needs
- **Encryption**: Store tokens and sensitive data encrypted
- **Audit Logs**: Track all API calls for compliance
- **Data Retention**: Let users control what's synced and how long
- **Token Refresh**: Automatic, transparent to user

---

## Implementation Roadmap

### Phase 1: Calendar Foundation (Current)
- ✅ Event sync with full metadata
- ✅ Bidirectional event management
- ✅ Meeting link extraction
- ⏳ Multi-calendar support
- ⏳ Recurring event UI

### Phase 2: Email Integration (Partial)
- ✅ Email sync with content
- ✅ Folder operations (archive, trash)
- ✅ Send email
- ⏳ Search UI integration
- ⏳ Conversation threading
- ⏳ Label management

### Phase 3: Drive Integration (Planned)
- ⏳ File discovery and indexing
- ⏳ Permission tracking
- ⏳ Full-text search
- ⏳ Comment extraction

### Phase 4: Advanced Features
- ⏳ Watch API (push notifications)
- ⏳ Batch operations
- ⏳ Conflict resolution
- ⏳ Offline sync

---

## Key Metrics for BusinessOS

### Data Volume Estimates

| Source | Per-User | Total (100 users) |
|--------|----------|-------------------|
| Calendar Events (1 year) | 100-500 | 10K-50K |
| Emails (1 year) | 2,000-5,000 | 200K-500K |
| Drive Files | 100-1,000 | 10K-100K |
| **Estimated Nodes** | **500-2,500** | **50K-250K** |

### API Call Estimates

**Initial Sync**
- Calendar: 10-20 API calls
- Email: 100-200 API calls (batching)
- Drive: 50-100 API calls (pagination)
- **Total**: 200-400 API calls

**Daily Incremental Sync**
- Calendar: 5 API calls
- Email: 20-50 API calls
- Drive: 10-20 API calls
- **Total**: 35-75 API calls/day

### Storage Estimates

| Data Type | Size per User | 100 Users |
|-----------|---------------|-----------|
| Calendar (metadata) | 100 KB | 10 MB |
| Email (no attachments) | 200 MB | 20 GB |
| Drive (metadata) | 50 MB | 5 GB |
| Attachments (cached) | Variable | 10-50 GB |
| **Total** | **~250 MB** | **~25 GB** |

---

## Next Steps

1. **Complete Gmail Implementation** - Search UI, thread operations
2. **Implement Drive API** - File discovery, permissions tracking
3. **Setup Watch API** - Push-based sync for real-time updates
4. **Optimize Sync** - Batch operations, caching, offline support
5. **User Controls** - Allow per-source sync frequency, selective sync
6. **Conflict Resolution** - Handle offline changes, user preferences

---

## References

- Google Calendar API: https://developers.google.com/calendar/api
- Gmail API: https://developers.google.com/gmail/api
- Google Drive API: https://developers.google.com/drive/api
- OAuth 2.0: https://developers.google.com/identity/protocols/oauth2
- Rate Limiting: https://developers.google.com/calendar/api/guides/quota
