# Mobile API Specification

> **Version:** 1.0.0
> **Last Updated:** January 7, 2026
> **Status:** ✅ Implemented
> **Assigned:** Javaris (Primary)

---

## Quick Reference

| Endpoint | Method | Status | Description |
|----------|--------|--------|-------------|
| `/api/mobile/v1/me` | GET | ✅ | Current user profile |
| `/api/mobile/v1/tasks` | GET | ✅ | List tasks (paginated) |
| `/api/mobile/v1/tasks/:id` | GET | ✅ | Task details |
| `/api/mobile/v1/tasks/quick` | POST | ✅ | Quick create task |
| `/api/mobile/v1/tasks/:id/status` | PUT | ✅ | Update task status |
| `/api/mobile/v1/tasks/:id/toggle` | PUT | ✅ | Toggle done/todo |
| `/api/mobile/v1/notifications` | GET | ✅ | List notifications |
| `/api/mobile/v1/notifications/count` | GET | ✅ | Unread count |
| `/api/mobile/v1/notifications/mark-read` | POST | ✅ | Mark as read |
| `/api/mobile/v1/dailylog/today` | GET | ✅ | Today's log |
| `/api/mobile/v1/dailylog/history` | GET | ✅ | Log history |
| `/api/mobile/v1/sync` | GET | ✅ | Delta sync |
| `/api/mobile/v1/chat/threads` | GET | ✅ | Chat threads |
| `/api/mobile/v1/chat/history/:id` | GET | ✅ | Chat messages |
| `/api/mobile/v1/chat/message` | POST | ✅ | Send message |
| `/api/mobile/v1/push/register` | POST | ✅ | Register device |
| `/api/mobile/v1/push/unregister` | DELETE | ✅ | Unregister device |

---

## Table of Contents

- [Overview](#overview)
- [Design Principles](#design-principles)
- [Architecture](#architecture)
- [Authentication](#authentication)
- [Core Endpoints](#core-endpoints)
- [Tasks API](#tasks-api)
- [Notifications API](#notifications-api)
- [Daily Log API](#daily-log-api)
- [Chat/AI API](#chatai-api)
- [Quick Capture API](#quick-capture-api)
- [Sync Protocol](#sync-protocol)
- [Push Notifications](#push-notifications)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Caching Strategy](#caching-strategy)
- [Implementation Checklist](#implementation-checklist)

---

## Overview

The Mobile API provides optimized endpoints for mobile clients (PWA, native apps). All endpoints are workspace-scoped and designed for:

- **Minimal bandwidth** - Lean payloads, field selection
- **Fast responses** - Aggressive caching, ETags
- **Offline support** - Delta sync protocol
- **Battery efficiency** - Push-triggered sync instead of polling

### Base URL

```
/api/mobile/v1
```

### Entities Supported

| Entity | Priority | Status |
|--------|----------|--------|
| Tasks | HIGH | ✅ Implemented |
| Notifications | HIGH | ✅ Implemented |
| Daily Log | HIGH | ✅ Implemented |
| Chat/AI | HIGH | ✅ Implemented |
| Sync | HIGH | ✅ Implemented |
| Push | HIGH | ✅ Implemented |
| Projects | MEDIUM | Future |
| Clients | LOW | Future |

---

## Design Principles

### 1. Lean Payloads

```go
// ❌ Web API response (heavy)
type TaskResponse struct {
    ID          uuid.UUID          `json:"id"`
    Title       string             `json:"title"`
    Description string             `json:"description"`      // Heavy
    Status      string             `json:"status"`
    Priority    string             `json:"priority"`
    DueDate     *time.Time         `json:"due_date"`
    AssigneeID  string             `json:"assignee_id"`
    Assignee    *UserResponse      `json:"assignee"`         // Nested
    ProjectID   *uuid.UUID         `json:"project_id"`
    Project     *ProjectResponse   `json:"project"`          // Nested
    Comments    []CommentResponse  `json:"comments"`         // Array
    Attachments []AttachmentResp   `json:"attachments"`      // Array
    CreatedAt   time.Time          `json:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at"`
}

// ✅ Mobile API response (lean)
type MobileTaskResponse struct {
    ID        uuid.UUID  `json:"id"`
    Title     string     `json:"title"`
    Status    string     `json:"status"`
    Priority  string     `json:"priority"`
    DueDate   *string    `json:"due_date,omitempty"`   // ISO string
    Assignee  string     `json:"assignee,omitempty"`   // Just name
    Project   string     `json:"project,omitempty"`    // Just name
    UpdatedAt int64      `json:"updated_at"`           // Unix timestamp
}
```

### 2. Field Selection

```bash
# Request only needed fields
GET /api/mobile/v1/tasks?fields=id,title,status,due_date

# Response includes only requested fields
{
    "tasks": [
        {"id": "uuid", "title": "...", "status": "...", "due_date": "..."}
    ]
}
```

### 3. Pagination

All list endpoints are paginated:

```bash
GET /api/mobile/v1/tasks?limit=20&cursor=eyJpZCI6Inh4eCJ9
```

### 4. ETag Caching

```bash
# First request
GET /api/mobile/v1/tasks
Response: 200 OK
ETag: "abc123"

# Subsequent request
GET /api/mobile/v1/tasks
If-None-Match: "abc123"
Response: 304 Not Modified (no body)
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        MOBILE API ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         MOBILE CLIENT                                │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │    │
│  │  │   PWA       │  │  iOS App    │  │ Android App │                 │    │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘                 │    │
│  └─────────┼────────────────┼────────────────┼─────────────────────────┘    │
│            │                │                │                               │
│            └────────────────┼────────────────┘                               │
│                             │                                                │
│                             ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      API GATEWAY                                     │    │
│  │                                                                      │    │
│  │  • Rate limiting (100 req/min)                                      │    │
│  │  • JWT validation                                                    │    │
│  │  • Workspace context injection                                       │    │
│  │  • Request/Response compression (gzip)                              │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                             │                                                │
│                             ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                   MOBILE HANDLERS                                    │    │
│  │                                                                      │    │
│  │  /api/mobile/v1/                                                    │    │
│  │  ├── me                    → UserHandler                            │    │
│  │  ├── sync                  → SyncHandler                            │    │
│  │  ├── tasks/*               → MobileTaskHandler                      │    │
│  │  ├── notifications/*       → MobileNotificationHandler              │    │
│  │  ├── dailylog/*            → MobileDailyLogHandler                  │    │
│  │  ├── chat/*                → MobileChatHandler                      │    │
│  │  ├── capture/*             → CaptureHandler                         │    │
│  │  └── push/*                → PushHandler                            │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                             │                                                │
│                             ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                    EXISTING SERVICES                                 │    │
│  │                                                                      │    │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐       │    │
│  │  │ TaskSvc    │ │ NotifSvc   │ │ DailyLogSvc│ │ ChatSvc    │       │    │
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘       │    │
│  │                                                                      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Authentication

All mobile endpoints require authentication via **session cookie** (Better Auth).

### Auth Methods

#### 1. Email/Password Sign-In

```http
POST /api/auth/sign-in/email
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "your_password"
}
```

**Response:**
```json
{
    "user": {
        "id": "user_123",
        "email": "user@example.com",
        "name": "John Smith"
    },
    "session": {
        "id": "session_uuid",
        "expiresAt": "2026-02-04T12:00:00Z"
    }
}
```

**Response Headers:**
```http
Set-Cookie: better_auth.session_token=<session_token>; Path=/; HttpOnly; SameSite=Lax
```

#### 2. Google OAuth

```http
GET /api/auth/google
```

Redirects to Google OAuth flow. After success, session cookie is set automatically.

---

### Required Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Cookie` | ✅ Yes | `better_auth.session_token=<token>` |
| `X-Device-ID` | ✅ Yes | Unique device identifier (UUID recommended) |
| `Accept-Encoding` | Optional | `gzip` for compressed responses |

### Example Request

```http
GET /api/mobile/v1/tasks
Cookie: better_auth.session_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-Device-ID: 550e8400-e29b-41d4-a716-446655440000
```

---

### Session Management

#### Get Current Session

```http
GET /api/auth/session
Cookie: better_auth.session_token=<token>
```

**Response:**
```json
{
    "user": {
        "id": "user_123",
        "email": "user@example.com",
        "name": "John Smith"
    },
    "session": {
        "id": "session_uuid",
        "expiresAt": "2026-02-04T12:00:00Z"
    }
}
```

#### Sign Out

```http
POST /api/auth/sign-out
Cookie: better_auth.session_token=<token>
```

#### Sign Out All Devices

```http
POST /api/auth/logout-all
Cookie: better_auth.session_token=<token>
```

---

### Mobile App Authentication Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      MOBILE AUTH FLOW                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. SIGN IN                                                                  │
│     ─────────                                                                │
│     POST /api/auth/sign-in/email                                            │
│     Body: { "email": "...", "password": "..." }                             │
│                                                                              │
│  2. STORE SESSION TOKEN                                                      │
│     ─────────────────────                                                    │
│     Extract from response: Set-Cookie: better_auth.session_token=xxx        │
│     Store securely in device keychain/secure storage                        │
│                                                                              │
│  3. MAKE AUTHENTICATED REQUESTS                                              │
│     ──────────────────────────                                               │
│     Include headers on all /api/mobile/v1/* requests:                       │
│       Cookie: better_auth.session_token=<stored_token>                      │
│       X-Device-ID: <device_uuid>                                            │
│                                                                              │
│  4. HANDLE 401 UNAUTHORIZED                                                  │
│     ───────────────────────                                                  │
│     If 401 received → session expired → redirect to sign-in                 │
│                                                                              │
│  5. SIGN OUT                                                                 │
│     ────────                                                                 │
│     POST /api/auth/sign-out                                                 │
│     Clear stored token from device                                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

### Testing with cURL

```bash
# 1. Sign in and capture session cookie
curl -c cookies.txt -X POST http://localhost:8001/api/auth/sign-in/email \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 2. Use session cookie for mobile API
curl -b cookies.txt http://localhost:8001/api/mobile/v1/tasks \
  -H "X-Device-ID: test-device-123"
```

### Testing with PowerShell

```powershell
# 1. Sign in and save session
$loginBody = @{ email = "test@example.com"; password = "password123" } | ConvertTo-Json
$response = Invoke-WebRequest -Uri "http://localhost:8001/api/auth/sign-in/email" `
  -Method POST -ContentType "application/json" -Body $loginBody -SessionVariable session

# 2. Use session for mobile API requests
Invoke-WebRequest -Uri "http://localhost:8001/api/mobile/v1/tasks" `
  -Headers @{"X-Device-ID"="test-device-123"} -WebSession $session
```

---

## Core Endpoints

### GET /api/mobile/v1/me

Returns current user profile with workspace context.

**Response:**
```json
{
    "user": {
        "id": "user_123",
        "email": "john@example.com",
        "name": "John Smith",
        "avatar_url": "https://...",
        "timezone": "America/New_York"
    },
    "workspace": {
        "id": "ws_uuid",
        "name": "Acme Corp",
        "role": "member",
        "permissions": {
            "tasks": {"create": true, "update": true},
            "projects": {"create": false}
        }
    },
    "preferences": {
        "notifications_enabled": true,
        "quiet_hours_start": "22:00",
        "quiet_hours_end": "08:00"
    }
}
```

**Size:** ~500 bytes

---

## Tasks API

### GET /api/mobile/v1/tasks

List tasks for current user (workspace-scoped).

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `limit` | int | 20 | Max items (1-50) |
| `cursor` | string | - | Pagination cursor |
| `status` | string | - | Filter: `pending`, `in_progress`, `completed` |
| `due` | string | - | Filter: `today`, `week`, `overdue` |
| `fields` | string | all | Comma-separated fields |
| `sort` | string | `due_date` | Sort: `due_date`, `priority`, `updated_at` |

**Request:**
```http
GET /api/mobile/v1/tasks?limit=20&status=pending&due=week&fields=id,title,status,due_date,priority
```

**Response:**
```json
{
    "tasks": [
        {
            "id": "task_uuid_1",
            "title": "Review proposal",
            "status": "pending",
            "priority": "high",
            "due_date": "2026-01-05",
            "assignee": "John Smith",
            "project": "Q1 Planning",
            "updated_at": 1704384000
        },
        {
            "id": "task_uuid_2",
            "title": "Send invoice",
            "status": "pending",
            "priority": "normal",
            "due_date": "2026-01-06",
            "assignee": null,
            "project": null,
            "updated_at": 1704380400
        }
    ],
    "cursor": "eyJpZCI6InRhc2tfdXVpZF8yIn0=",
    "has_more": true,
    "total": 45
}
```

**Size:** ~50-100 bytes per task

---

### GET /api/mobile/v1/tasks/:id

Get full task details (on-demand).

**Response:**
```json
{
    "id": "task_uuid_1",
    "title": "Review proposal",
    "description": "Full description here with markdown...",
    "status": "pending",
    "priority": "high",
    "due_date": "2026-01-05T17:00:00Z",
    "assignee": {
        "id": "user_123",
        "name": "John Smith",
        "avatar_url": "https://..."
    },
    "project": {
        "id": "proj_uuid",
        "name": "Q1 Planning"
    },
    "tags": ["urgent", "client"],
    "checklist": [
        {"id": "check_1", "text": "Read document", "completed": true},
        {"id": "check_2", "text": "Write feedback", "completed": false}
    ],
    "comments_count": 5,
    "attachments_count": 2,
    "created_at": "2026-01-01T10:00:00Z",
    "updated_at": "2026-01-04T15:30:00Z"
}
```

**Size:** ~2KB

---

### POST /api/mobile/v1/tasks/quick

Quick task creation (minimal input).

**Request:**
```json
{
    "title": "Call John about proposal",
    "due_date": "2026-01-05",
    "priority": "high"
}
```

**Response:**
```json
{
    "id": "task_uuid_new",
    "title": "Call John about proposal",
    "status": "pending",
    "priority": "high",
    "due_date": "2026-01-05",
    "created_at": 1704384000
}
```

---

### PUT /api/mobile/v1/tasks/:id/status

Update task status.

**Request:**
```json
{
    "status": "in_progress"
}
```

**Valid statuses:** `todo`, `in_progress`, `done`, `cancelled`

**Response:**
```json
{
    "id": "task_uuid",
    "status": "in_progress",
    "updated_at": 1704384100
}
```

---

### PUT /api/mobile/v1/tasks/:id/toggle

Toggle task completion (done ↔ todo).

**Request:** (empty body)

**Response:**
```json
{
    "id": "task_uuid",
    "status": "done",
    "completed_at": "2026-01-04T16:00:00Z",
    "updated_at": 1704384100
}
```

---

## Notifications API

### GET /api/mobile/v1/notifications

Get unread notifications.

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `limit` | int | 20 | Max items |
| `cursor` | string | - | Pagination cursor |
| `unread_only` | bool | true | Only unread |

**Response:**
```json
{
    "notifications": [
        {
            "id": "notif_uuid_1",
            "type": "task.assigned",
            "title": "You were assigned: Review proposal",
            "body": "Assigned by Sarah",
            "entity_type": "task",
            "entity_id": "task_uuid",
            "priority": "normal",
            "is_read": false,
            "created_at": 1704384000
        }
    ],
    "cursor": "xxx",
    "has_more": false,
    "unread_count": 12
}
```

---

### GET /api/mobile/v1/notifications/count

Get unread count only (for badge).

**Response:**
```json
{
    "unread_count": 12
}
```

**Size:** ~20 bytes

---

### POST /api/mobile/v1/notifications/mark-read

Batch mark as read.

**Request:**
```json
{
    "notification_ids": ["notif_uuid_1", "notif_uuid_2"],
    "all": false
}
```

Or mark all:
```json
{
    "all": true
}
```

**Response:**
```json
{
    "marked_count": 2,
    "unread_count": 10
}
```

---

## Daily Log API

### GET /api/mobile/v1/dailylog/today

Get today's daily log entries.

**Response:**
```json
{
    "date": "2026-01-04",
    "entries": [
        {
            "id": "entry_uuid_1",
            "content": "Had productive meeting with client",
            "type": "note",
            "created_at": 1704380400
        },
        {
            "id": "entry_uuid_2",
            "content": "Completed proposal review",
            "type": "accomplishment",
            "linked_task_id": "task_uuid",
            "created_at": 1704384000
        }
    ],
    "summary": null,
    "mood": null
}
```

---

### POST /api/mobile/v1/dailylog/entry

Add entry to today's log.

**Request:**
```json
{
    "content": "Finished client presentation prep",
    "type": "accomplishment",
    "linked_task_id": "task_uuid"
}
```

**Response:**
```json
{
    "id": "entry_uuid_new",
    "content": "Finished client presentation prep",
    "type": "accomplishment",
    "linked_task_id": "task_uuid",
    "created_at": 1704384100
}
```

---

### GET /api/mobile/v1/dailylog/history

Get past daily logs (paginated).

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `limit` | int | 7 | Number of days |
| `before` | date | today | Start from date |

**Response:**
```json
{
    "logs": [
        {
            "date": "2026-01-03",
            "entry_count": 5,
            "summary": "Focused on client work and proposal...",
            "mood": "productive"
        },
        {
            "date": "2026-01-02",
            "entry_count": 3,
            "summary": "Planning session for Q1...",
            "mood": "neutral"
        }
    ],
    "has_more": true
}
```

---

## Chat/AI API

### POST /api/mobile/v1/chat/message

Send message to AI (non-streaming for mobile).

**Request:**
```json
{
    "thread_id": "thread_uuid",
    "message": "What tasks do I have due this week?",
    "context": {
        "current_view": "tasks",
        "selected_project_id": null
    }
}
```

**Response:**
```json
{
    "id": "msg_uuid",
    "thread_id": "thread_uuid",
    "role": "assistant",
    "content": "You have 5 tasks due this week:\n\n1. Review proposal (Tomorrow)\n2. Send invoice (Jan 6)\n...",
    "tool_calls": [
        {
            "tool": "list_tasks",
            "result_summary": "Found 5 tasks due this week"
        }
    ],
    "created_at": 1704384100
}
```

**Note:** Mobile uses non-streaming responses. For streaming, use WebSocket (future).

---

### GET /api/mobile/v1/chat/history

Get chat history for thread.

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `thread_id` | uuid | required | Thread ID |
| `limit` | int | 20 | Max messages |
| `before` | string | - | Cursor for pagination |

**Response:**
```json
{
    "thread_id": "thread_uuid",
    "messages": [
        {
            "id": "msg_uuid_1",
            "role": "user",
            "content": "What's on my schedule today?",
            "created_at": 1704380400
        },
        {
            "id": "msg_uuid_2",
            "role": "assistant",
            "content": "You have 3 meetings...",
            "created_at": 1704380410
        }
    ],
    "has_more": true
}
```

---

### GET /api/mobile/v1/chat/threads

List chat threads.

**Response:**
```json
{
    "threads": [
        {
            "id": "thread_uuid_1",
            "title": "Task planning",
            "last_message": "You have 5 tasks due...",
            "updated_at": 1704384100
        },
        {
            "id": "thread_uuid_2",
            "title": "Project questions",
            "last_message": "The Q1 Planning project...",
            "updated_at": 1704380400
        }
    ]
}
```

---

## Quick Capture API

### POST /api/mobile/v1/capture

Smart capture with LLM classification.

**Request:**
```json
{
    "text": "Call John tomorrow about the proposal",
    "source": "quick_capture"
}
```

**Response:**
```json
{
    "classification": {
        "type": "task",
        "confidence": 0.95,
        "reasoning": "Contains action verb 'call' and time indicator 'tomorrow'"
    },
    "created": {
        "type": "task",
        "id": "task_uuid_new",
        "title": "Call John about the proposal",
        "due_date": "2026-01-05",
        "status": "pending"
    },
    "extracted": {
        "action": "call",
        "person": "John",
        "subject": "the proposal",
        "due_date": "2026-01-05"
    }
}
```

### Classification Types

| Type | Trigger Patterns | Destination |
|------|------------------|-------------|
| `task` | Action verbs, time indicators | Create task |
| `daily_log` | Reflection, "today I", "learned" | Add to daily log |
| `inbox` | Questions, notes, low confidence | Create inbox item |

---

### POST /api/mobile/v1/capture/voice

Upload voice note for transcription + smart routing.

**Request:**
```http
POST /api/mobile/v1/capture/voice
Content-Type: multipart/form-data

audio: <binary audio file>
format: "webm" | "m4a" | "wav"
duration_seconds: 15
```

**Response:**
```json
{
    "transcription": {
        "text": "Remind me to call John tomorrow about the proposal",
        "confidence": 0.92,
        "duration_seconds": 5.2
    },
    "classification": {
        "type": "task",
        "confidence": 0.95
    },
    "created": {
        "type": "task",
        "id": "task_uuid_new",
        "title": "Call John about the proposal",
        "due_date": "2026-01-05"
    }
}
```

---

## Sync Protocol

### GET /api/mobile/v1/sync

Delta sync - get all changes since last sync.

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `since` | timestamp | Yes | ISO timestamp of last sync |
| `types` | string | No | Comma-separated: `tasks,notifications,dailylog` |

**Request:**
```http
GET /api/mobile/v1/sync?since=2026-01-04T10:00:00Z&types=tasks,notifications
X-Workspace-ID: ws_uuid
X-Device-ID: device_uuid
```

**Response:**
```json
{
    "server_time": "2026-01-04T16:30:00Z",
    "changes": {
        "tasks": [
            {
                "id": "task_uuid_1",
                "action": "updated",
                "data": {
                    "id": "task_uuid_1",
                    "title": "Review proposal",
                    "status": "completed",
                    "updated_at": 1704384000
                }
            },
            {
                "id": "task_uuid_2",
                "action": "created",
                "data": {
                    "id": "task_uuid_2",
                    "title": "New task",
                    "status": "pending",
                    "updated_at": 1704384100
                }
            }
        ],
        "notifications": [
            {
                "id": "notif_uuid_1",
                "action": "created",
                "data": {
                    "id": "notif_uuid_1",
                    "type": "task.completed",
                    "title": "Task completed: Review proposal"
                }
            }
        ]
    },
    "deleted": {
        "tasks": ["task_uuid_old"],
        "notifications": []
    },
    "has_more": false,
    "sync_token": "eyJsYXN0X3N5bmMiOiIyMDI2LTAxLTA0VDE2OjMwOjAwWiJ9"
}
```

### Sync Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SYNC PROTOCOL                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  INITIAL SYNC (App Install / First Login)                                   │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  1. Client calls: GET /api/mobile/v1/sync?since=1970-01-01T00:00:00Z       │
│  2. Server returns all data (paginated if large)                            │
│  3. Client stores server_time as last_sync                                  │
│                                                                              │
│  INCREMENTAL SYNC (Subsequent)                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  1. Push notification arrives: "New task assigned"                          │
│  2. Client wakes up, calls: GET /api/mobile/v1/sync?since=<last_sync>      │
│  3. Server returns only changes since last_sync                             │
│  4. Client merges changes into local cache                                  │
│  5. Client updates last_sync to server_time                                 │
│                                                                              │
│  CONFLICT RESOLUTION                                                         │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  • Server wins (last-write-wins)                                            │
│  • Client should refetch if local version is stale                          │
│  • Use updated_at timestamps for comparison                                 │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Push Notifications

### POST /api/mobile/v1/push/register

Register device for push notifications.

**Request:**
```json
{
    "token": "fcm_or_apns_token",
    "platform": "web" | "ios" | "android",
    "device_id": "unique_device_uuid",
    "device_name": "iPhone 15 Pro"
}
```

**Response:**
```json
{
    "registered": true,
    "device_id": "unique_device_uuid"
}
```

---

### DELETE /api/mobile/v1/push/unregister

Unregister device (logout).

**Request:**
```json
{
    "device_id": "unique_device_uuid"
}
```

**Response:**
```json
{
    "unregistered": true
}
```

---

### Push Payload Format

When server sends push:

```json
{
    "notification": {
        "title": "Task Assigned",
        "body": "You were assigned: Review proposal"
    },
    "data": {
        "type": "task.assigned",
        "entity_type": "task",
        "entity_id": "task_uuid",
        "action": "sync",
        "badge_count": 12
    }
}
```

### Push → Sync Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      PUSH-TRIGGERED SYNC                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  SERVER                              CLIENT                                  │
│  ──────                              ──────                                  │
│                                                                              │
│  1. Event: Task assigned to user                                            │
│     │                                                                        │
│     ▼                                                                        │
│  2. Send push notification ─────────────────▶ 3. Receive push              │
│     {                                              │                         │
│       "data": {                                    │                         │
│         "action": "sync",                          ▼                         │
│         "badge_count": 12                     4. Check action == "sync"     │
│       }                                            │                         │
│     }                                              ▼                         │
│                                              5. Call GET /sync?since=...    │
│                                              ◀────────────────────────────  │
│  6. Return changes ─────────────────────────▶ 7. Merge into local cache    │
│                                                    │                         │
│                                                    ▼                         │
│                                              8. Update UI + badge           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Error Handling

### Error Response Format

```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Invalid due_date format",
        "details": {
            "field": "due_date",
            "expected": "ISO 8601 date string"
        }
    }
}
```

### Error Codes

| HTTP | Code | Description |
|------|------|-------------|
| 400 | `VALIDATION_ERROR` | Invalid input |
| 400 | `INVALID_CURSOR` | Bad pagination cursor |
| 401 | `UNAUTHORIZED` | Missing/invalid token |
| 401 | `TOKEN_EXPIRED` | JWT expired |
| 403 | `FORBIDDEN` | No permission |
| 403 | `WORKSPACE_ACCESS_DENIED` | Not a member |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Version conflict |
| 422 | `CAPTURE_FAILED` | Smart capture failed |
| 429 | `RATE_LIMITED` | Too many requests |
| 500 | `INTERNAL_ERROR` | Server error |
| 503 | `SERVICE_UNAVAILABLE` | Maintenance |

### Retry Strategy

| Error | Retry? | Strategy |
|-------|--------|----------|
| 401 TOKEN_EXPIRED | Yes | Refresh token, retry once |
| 429 RATE_LIMITED | Yes | Exponential backoff |
| 500 INTERNAL_ERROR | Yes | 3 retries with backoff |
| 503 SERVICE_UNAVAILABLE | Yes | Wait for Retry-After header |
| 400, 403, 404 | No | Show error to user |

---

## Rate Limiting

### Limits

| Endpoint Category | Limit | Window |
|-------------------|-------|--------|
| Read (GET) | 100 | 1 minute |
| Write (POST/PUT) | 30 | 1 minute |
| Sync | 10 | 1 minute |
| Capture | 20 | 1 minute |
| Chat | 10 | 1 minute |

### Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704384060
```

### 429 Response

```json
{
    "error": {
        "code": "RATE_LIMITED",
        "message": "Too many requests",
        "retry_after": 45
    }
}
```

---

## Caching Strategy

### ETag Support

All GET endpoints support ETag:

```http
# Response includes ETag
GET /api/mobile/v1/tasks
ETag: "abc123def456"

# Client sends If-None-Match
GET /api/mobile/v1/tasks
If-None-Match: "abc123def456"

# If unchanged, returns 304
HTTP/1.1 304 Not Modified
```

### Cache-Control Headers

| Endpoint | Cache-Control |
|----------|---------------|
| `/me` | `private, max-age=300` (5 min) |
| `/tasks` | `private, max-age=60` (1 min) |
| `/notifications` | `private, no-cache` |
| `/sync` | `private, no-store` |

### Offline Cache Recommendations

| Entity | Cache Duration | Stale-While-Revalidate |
|--------|---------------|------------------------|
| Tasks | 1 hour | Yes |
| Projects | 1 hour | Yes |
| Notifications | 5 minutes | No |
| Daily Log | 24 hours | Yes |

---

## Implementation Checklist

### Phase 1: Core ✅ Complete

- [x] Create `mobile_handlers.go` file
- [x] `GET /api/mobile/v1/me`
- [x] `GET /api/mobile/v1/tasks`
- [x] `GET /api/mobile/v1/tasks/:id`
- [x] `POST /api/mobile/v1/tasks/quick`
- [x] `PUT /api/mobile/v1/tasks/:id/toggle`
- [x] `PUT /api/mobile/v1/tasks/:id/status`
- [x] Route registration under `/api/mobile/v1`

### Phase 2: Notifications + Daily Log ✅ Complete

- [x] `GET /api/mobile/v1/notifications`
- [x] `GET /api/mobile/v1/notifications/count`
- [x] `POST /api/mobile/v1/notifications/mark-read`
- [x] `GET /api/mobile/v1/dailylog/today`
- [x] `GET /api/mobile/v1/dailylog/history`

### Phase 3: Sync Protocol ✅ Complete

- [x] `GET /api/mobile/v1/sync`
- [x] Track `updated_at` on all synced entities
- [x] Delta sync by Unix timestamp

### Phase 4: Chat ✅ Complete

- [x] `GET /api/mobile/v1/chat/threads`
- [x] `GET /api/mobile/v1/chat/history/:id`
- [x] `POST /api/mobile/v1/chat/message`

### Phase 5: Push Notifications ✅ Complete

- [x] Push registration table (`push_devices`)
- [x] `POST /api/mobile/v1/push/register`
- [x] `DELETE /api/mobile/v1/push/unregister`

### Phase 6: Polish (Ongoing)

- [ ] ETag support
- [x] Rate limiting middleware
- [ ] Compression (gzip)
- [x] Error standardization
- [ ] API documentation (OpenAPI)

---

## Files

| File | Description |
|------|-------------|
| `internal/handlers/mobile_handlers.go` | All mobile endpoint handlers |
| `internal/handlers/mobile_types.go` | Request/response type definitions |
| `internal/handlers/mobile_utils.go` | Helper functions (cursor encoding, field selection) |
| `internal/database/queries/mobile.sql` | Optimized SQL queries for mobile |

---

## Database Additions

### New Tables Required

```sql
-- Push device registrations
CREATE TABLE push_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id),
    
    device_id VARCHAR(255) NOT NULL UNIQUE,
    device_name VARCHAR(100),
    platform VARCHAR(20) NOT NULL,  -- 'web', 'ios', 'android'
    push_token TEXT NOT NULL,
    
    is_active BOOLEAN DEFAULT TRUE,
    last_used_at TIMESTAMPTZ DEFAULT NOW(),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_push_devices_user ON push_devices(user_id);
CREATE INDEX idx_push_devices_token ON push_devices(push_token);

-- Sync tracking (for deleted items)
CREATE TABLE sync_deletions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    entity_type VARCHAR(50) NOT NULL,  -- 'task', 'notification', etc.
    entity_id UUID NOT NULL,
    deleted_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sync_deletions_workspace ON sync_deletions(workspace_id, deleted_at);

-- Inbox items (for low-confidence captures)
CREATE TABLE inbox_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    workspace_id UUID REFERENCES workspaces(id),
    
    content TEXT NOT NULL,
    source VARCHAR(50) NOT NULL,  -- 'capture', 'voice', 'email'
    suggested_type VARCHAR(50),   -- 'task', 'daily_log', null
    confidence DECIMAL(3,2),
    
    is_processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMPTZ,
    processed_to_type VARCHAR(50),
    processed_to_id UUID,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_inbox_items_user ON inbox_items(user_id, is_processed);
```

---

## Environment Variables

```env
# Mobile API
MOBILE_API_RATE_LIMIT_READ=100
MOBILE_API_RATE_LIMIT_WRITE=30
MOBILE_SYNC_MAX_ITEMS=500
MOBILE_CAPTURE_LLM_MODEL=gpt-4o-mini

# Push Notifications
FCM_PROJECT_ID=your-firebase-project
FCM_CREDENTIALS_JSON=/path/to/firebase-credentials.json
APNS_KEY_ID=xxx
APNS_TEAM_ID=xxx
APNS_KEY_PATH=/path/to/apns-key.p8

# Voice Transcription
WHISPER_API_KEY=xxx
MAX_VOICE_DURATION_SECONDS=60
```

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | Jan 4, 2026 | Initial specification |

---

*Document maintained by the BusinessOS Development Team*
