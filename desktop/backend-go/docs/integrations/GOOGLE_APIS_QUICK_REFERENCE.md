# Google APIs Quick Reference

**Quick lookup for Google API data available to BusinessOS**

---

## API Endpoints at a Glance

### Google Calendar API

| What | Endpoint | Method | Key Fields Returned |
|------|----------|--------|-------------------|
| List events | `/calendars/{calendarId}/events` | GET | id, summary, start, end, attendees, conferenceData |
| Create event | `/calendars/{calendarId}/events` | POST | Same as above |
| Get event | `/calendars/{calendarId}/events/{eventId}` | GET | Full event object |
| Delete event | `/calendars/{calendarId}/events/{eventId}` | DELETE | N/A |

**Max Results**: 2,500 per request | **Rate**: 1M/day quota

---

### Gmail API

| What | Endpoint | Method | Key Fields Returned |
|------|----------|--------|-------------------|
| List messages | `/users/me/messages` | GET | id, threadId (basic) |
| Get message | `/users/me/messages/{id}` | GET | Full payload with headers, body, attachments |
| Send message | `/users/me/messages/send` | POST | id, threadId |
| Mark read | `/users/me/messages/{id}/modify` | POST | N/A (modifies labels) |
| Search emails | `/users/me/messages` + `q=query` | GET | Filtered message list |

**Max Results**: 500 per request | **Rate**: 250 req/sec, 1M/day quota

**Search Examples**: `in:inbox`, `from:user@example.com`, `has:attachment`, `before:2023-01-01`

---

### Google Drive API

| What | Endpoint | Method | Key Fields Returned |
|------|----------|--------|-------------------|
| List files | `/files` | GET | id, name, mimeType, size, createdTime, modifiedTime, owners, permissions |
| Get file | `/files/{fileId}` | GET | Full metadata |
| Download content | `/files/{fileId}?alt=media` | GET | Raw file content |
| List permissions | `/files/{fileId}/permissions` | GET | Permission objects (users/groups/domains) |
| List revisions | `/files/{fileId}/revisions` | GET | Version history with timestamps |
| List comments | `/files/{fileId}/comments` | GET | Comments + replies (Docs/Sheets/Slides) |

**Max Results**: 1,000 per request | **Rate**: 1M/day quota

**Search Examples**: `name contains 'project'`, `mimeType='application/vnd.google-apps.document'`, `trashed=false`

---

## Core Data Available

### From Calendar

```
Events (primary calendar):
├── Title/Summary
├── Description
├── Date & Time (with timezone)
├── Location
├── Status (confirmed/tentative/cancelled)
├── Attendees
│   ├── Email
│   ├── Name
│   ├── RSVP Status (accepted/declined/tentative/needsAction)
│   └── Optional (boolean)
├── Organizer (email, name)
├── Recurrence (RRULE)
├── Meeting Links (Google Meet, Zoom, Teams)
│   ├── URL
│   ├── Meeting code
│   └── Passcode
├── Reminders (email/popup/sms, time in minutes)
├── Attachments (files, links)
├── Color ID (visual categorization)
├── Event Type (working location, focus time, out of office)
├── Transparency (busy/free)
└── Created/Updated timestamps
```

### From Gmail

```
Messages/Emails:
├── Subject
├── From (name + email)
├── To (array of emails)
├── Cc (array)
├── Bcc (array)
├── Date
├── Body (text + HTML)
├── Snippet (preview, 100 chars)
├── Attachments
│   ├── Filename
│   ├── MIME type
│   ├── Size
│   └── Content (base64)
├── Labels (INBOX, SENT, DRAFT, etc.)
├── Thread ID (for conversations)
├── Flags
│   ├── Is Read
│   ├── Is Starred
│   ├── Is Important
│   ├── Is Draft
│   ├── Is Sent
│   ├── Is Archived
│   └── Is Trash
├── Size (bytes)
└── Reply-To header
```

### From Drive

```
Files & Folders:
├── Name
├── ID
├── MIME Type (determines file type)
├── Description
├── Size (bytes)
├── Created timestamp
├── Modified timestamp
├── Modified by (user email/name)
├── Owners (array of users)
├── Permissions
│   ├── Type (user/group/domain/anyone)
│   ├── Role (owner/editor/commenter/viewer)
│   ├── Email (if user/group)
│   └── Expiration time
├── Parent folder (for hierarchy)
├── Flags
│   ├── Is Starred
│   ├── Is Trashed
│   ├── Is Shared
│   └── Owned by Me
├── Capabilities (canEdit, canDelete, canShare, etc.)
├── Revision history (version numbers, timestamps, editors)
├── Comments (if Docs/Sheets/Slides)
│   ├── Author
│   ├── Content
│   ├── Timestamp
│   └── Replies
├── Web links (open in Drive, download)
├── Thumbnail (if available)
└── Custom properties (user-defined metadata)
```

---

## Authentication

All APIs use **OAuth 2.0**:

```
Client ID: (from Google Cloud Console)
Client Secret: (from Google Cloud Console)
Redirect URI: https://yourapp.com/auth/callback

Scopes needed:
- Calendar: https://www.googleapis.com/auth/calendar
- Gmail: https://www.googleapis.com/auth/gmail.modify
- Drive: https://www.googleapis.com/auth/drive.readonly
```

**Token Storage**: Encrypted in PostgreSQL `google_oauth_tokens` table

---

## Sync Frequency Recommendations

| Source | Frequency | Typical Latency |
|--------|-----------|-----------------|
| Calendar | Every 6 hours | < 1 hour |
| Gmail | Daily | < 4 hours |
| Drive | Daily | < 4 hours |

**Initial Sync**: 30-60 minutes total for all three APIs combined

---

## Rate Limits

| API | Daily Quota | Per-Second | Per-Request |
|-----|------------|-----------|------------|
| Calendar | 1M | Fair use | N/A |
| Gmail | 1M | 250 req/sec | 1 unit |
| Drive | 1M | Fair use | 1 unit (read), 10 units (write) |

**Strategy**: Implement exponential backoff for retries, batch operations where available

---

## Current Implementation Status

| API | Status | Key Features | Missing |
|-----|--------|-------------|---------|
| **Calendar** | Partial | Event CRUD, attendees, meeting links | Multi-calendar, Watch API |
| **Gmail** | Partial | Sync, read/archive/send, attachments | Search UI, threading, labels |
| **Drive** | Not started | Planning phase | All features |

---

## Common Errors & Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| 401 Unauthorized | Invalid/expired token | Re-authenticate user |
| 403 Forbidden | Missing scopes | Request new scopes |
| 404 Not Found | Resource deleted | Skip or mark for review |
| 429 Rate Limited | Too many requests | Exponential backoff + wait |
| 500+ Server Error | Google API issue | Retry after delay |

---

## Useful Query Examples

### Calendar
```
timeMin=2024-01-01T00:00:00Z
timeMax=2024-12-31T23:59:59Z
singleEvents=true          // Expand recurring events
orderBy=startTime
maxResults=250
```

### Gmail
```
q=in:inbox                 // Filter folder
q=is:unread                // Unread only
q=from:boss@example.com    // From specific person
q=has:attachment           // Has files
q=before:2024-01-01        // Date range
q=larger:1000000           // File size
```

### Drive
```
q=name contains 'project'
q=mimeType='application/vnd.google-apps.document'
q=trashed=false
q='folder_id' in parents
q=createdTime>'2024-01-01T00:00:00'
spaces=drive                // Only main drive
pageSize=1000
```

---

## Files in This Research

1. **GOOGLE_APIS_RESEARCH.json** - Complete structured data (913 lines)
2. **GOOGLE_APIS_SUMMARY.md** - Detailed markdown guide (590 lines)
3. **GOOGLE_APIS_QUICK_REFERENCE.md** - This file (quick lookups)

---

## Next Implementation Phase

**Priority Order**:
1. Complete Gmail implementation (search, threading)
2. Add Google Drive API support
3. Implement Watch API for push notifications
4. Add user controls for sync frequency
5. Setup batch operations for performance

---

**Last Updated**: January 5, 2026
