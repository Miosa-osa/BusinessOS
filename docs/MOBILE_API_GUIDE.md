# Mobile API Integration Guide

**For:** Mobile/UI Team  
**Last Updated:** January 8, 2026  
**Base URL:** `/api/mobile/v1`

---

## Overview

The Mobile API provides optimized endpoints for mobile clients with:
- **Lean payloads** - Minimal data, field selection
- **Cursor pagination** - Efficient scrolling
- **Delta sync** - Only fetch changes
- **Unix timestamps** - Smaller than ISO strings

---

## Authentication

All endpoints require session authentication via cookie:

```typescript
// All requests must include credentials
fetch('/api/mobile/v1/tasks', {
  credentials: 'include'  // Sends session cookie
});
```

---

## Quick Start

### 1. Get Current User

```typescript
const response = await fetch('/api/mobile/v1/me', {
  credentials: 'include'
});
const { user, workspace, preferences } = await response.json();
```

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
    "role": "member"
  },
  "preferences": {
    "notifications_enabled": true,
    "quiet_hours_start": "22:00",
    "quiet_hours_end": "08:00"
  }
}
```

---

## Tasks API

### List Tasks

```typescript
// Basic list
GET /api/mobile/v1/tasks

// With filters
GET /api/mobile/v1/tasks?status=pending&due=week&limit=20

// With field selection (smaller payload)
GET /api/mobile/v1/tasks?fields=id,title,status,due_date
```

**Query Parameters:**

| Param | Type | Default | Options |
|-------|------|---------|---------|
| `limit` | int | 20 | 1-50 |
| `cursor` | string | - | Pagination cursor |
| `status` | string | all | `pending`, `in_progress`, `completed` |
| `due` | string | - | `today`, `week`, `overdue` |
| `sort` | string | `due_date` | `due_date`, `priority`, `updated_at` |
| `fields` | string | all | Comma-separated field names |

**Response:**
```json
{
  "tasks": [
    {
      "id": "task_uuid",
      "title": "Review proposal",
      "status": "pending",
      "priority": "high",
      "due_date": "2026-01-10",
      "assignee": "John Smith",
      "project": "Q1 Planning",
      "updated_at": 1704844800
    }
  ],
  "cursor": "eyJpZCI6Inh4eCJ9",
  "has_more": true,
  "total": 45
}
```

### Get Task Details

```typescript
GET /api/mobile/v1/tasks/:id
```

**Response:**
```json
{
  "id": "task_uuid",
  "title": "Review proposal",
  "description": "Full description with markdown...",
  "status": "pending",
  "priority": "high",
  "due_date": "2026-01-10T17:00:00Z",
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
    {"id": "c1", "text": "Read document", "completed": true},
    {"id": "c2", "text": "Write feedback", "completed": false}
  ],
  "comments_count": 5,
  "attachments_count": 2,
  "created_at": "2026-01-01T10:00:00Z",
  "updated_at": "2026-01-08T15:30:00Z"
}
```

### Quick Create Task

```typescript
POST /api/mobile/v1/tasks/quick
Content-Type: application/json

{
  "title": "Call John about proposal",
  "due_date": "2026-01-10",
  "priority": "high"
}
```

### Update Task Status

```typescript
PUT /api/mobile/v1/tasks/:id/status
Content-Type: application/json

{
  "status": "in_progress"
}
```

**Valid statuses:** `todo`, `in_progress`, `done`, `cancelled`

### Toggle Task Completion

```typescript
PUT /api/mobile/v1/tasks/:id/toggle
// No body required - toggles between done/todo
```

---

## Notifications API

### List Notifications

```typescript
GET /api/mobile/v1/notifications?limit=20&unread_only=true
```

**Response:**
```json
{
  "notifications": [
    {
      "id": "notif_uuid",
      "type": "task.assigned",
      "title": "You were assigned: Review proposal",
      "body": "Assigned by Sarah",
      "entity_type": "task",
      "entity_id": "task_uuid",
      "priority": "normal",
      "is_read": false,
      "created_at": 1704844800
    }
  ],
  "cursor": "xxx",
  "has_more": false,
  "unread_count": 12
}
```

### Get Unread Count (Badge)

```typescript
GET /api/mobile/v1/notifications/count
```

**Response:**
```json
{
  "unread_count": 12
}
```

### Mark as Read

```typescript
// Mark specific notifications
POST /api/mobile/v1/notifications/mark-read
Content-Type: application/json

{
  "notification_ids": ["id1", "id2"]
}

// Mark all as read
POST /api/mobile/v1/notifications/mark-read
Content-Type: application/json

{
  "all": true
}
```

---

## Daily Log API

### Get Today's Log

```typescript
GET /api/mobile/v1/dailylog/today
```

**Response:**
```json
{
  "date": "2026-01-08",
  "entries": [
    {
      "id": "entry_uuid",
      "content": "Had productive meeting with client",
      "type": "note",
      "created_at": 1704844800
    },
    {
      "id": "entry_uuid_2",
      "content": "Completed proposal review",
      "type": "accomplishment",
      "linked_task_id": "task_uuid",
      "created_at": 1704848400
    }
  ],
  "summary": null,
  "mood": null
}
```

### Get History

```typescript
GET /api/mobile/v1/dailylog/history?limit=7
```

**Response:**
```json
{
  "logs": [
    {
      "date": "2026-01-07",
      "entry_count": 5,
      "summary": "Focused on client work...",
      "mood": "productive"
    }
  ],
  "has_more": true
}
```

---

## Chat/AI API

### List Threads

```typescript
GET /api/mobile/v1/chat/threads
```

**Response:**
```json
{
  "threads": [
    {
      "id": "thread_uuid",
      "title": "Task planning",
      "last_message": "You have 5 tasks due...",
      "updated_at": 1704848400
    }
  ]
}
```

### Get Chat History

```typescript
GET /api/mobile/v1/chat/history/:threadId?limit=20
```

### Send Message

```typescript
POST /api/mobile/v1/chat/message
Content-Type: application/json

{
  "thread_id": "thread_uuid",
  "message": "What tasks do I have due this week?"
}
```

**Response:**
```json
{
  "id": "msg_uuid",
  "thread_id": "thread_uuid",
  "role": "assistant",
  "content": "You have 5 tasks due this week...",
  "created_at": 1704848400
}
```

---

## Sync API

The sync endpoint returns all changes since a given timestamp.

```typescript
GET /api/mobile/v1/sync?since=2026-01-08T10:00:00Z&types=tasks,notifications
```

**Response:**
```json
{
  "server_time": "2026-01-08T16:30:00Z",
  "changes": {
    "tasks": [
      {
        "id": "task_uuid",
        "action": "updated",
        "data": { /* task object */ }
      },
      {
        "id": "task_uuid_2",
        "action": "created",
        "data": { /* task object */ }
      }
    ],
    "notifications": [
      {
        "id": "notif_uuid",
        "action": "created",
        "data": { /* notification object */ }
      }
    ]
  },
  "deleted": {
    "tasks": ["deleted_task_uuid"],
    "notifications": []
  },
  "has_more": false,
  "sync_token": "eyJ..."
}
```

### Sync Flow

```
1. Initial sync:     GET /sync?since=1970-01-01T00:00:00Z
2. Store server_time as last_sync
3. On push notification: GET /sync?since=<last_sync>
4. Merge changes into local cache
5. Update last_sync = server_time
```

---

## Push Notifications

### Register Device

```typescript
POST /api/mobile/v1/push/register
Content-Type: application/json

{
  "token": "fcm_or_apns_token",
  "platform": "ios",  // "ios", "android", "web"
  "device_id": "unique_device_uuid",
  "device_name": "iPhone 15 Pro"
}
```

### Unregister Device

```typescript
DELETE /api/mobile/v1/push/unregister
Content-Type: application/json

{
  "device_id": "unique_device_uuid"
}
```

### Push Payload Format

When the server sends a push:
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

**On receiving push with `action: "sync"`:**
1. Call `/api/mobile/v1/sync?since=<last_sync>`
2. Merge changes into local cache
3. Update badge count

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
      "expected": "ISO 8601 date"
    }
  }
}
```

### Error Codes

| HTTP | Code | Action |
|------|------|--------|
| 400 | `VALIDATION_ERROR` | Fix input |
| 401 | `UNAUTHORIZED` | Re-authenticate |
| 401 | `TOKEN_EXPIRED` | Refresh token |
| 403 | `FORBIDDEN` | No permission |
| 404 | `NOT_FOUND` | Resource missing |
| 429 | `RATE_LIMITED` | Back off |
| 500 | `INTERNAL_ERROR` | Retry with backoff |

### Retry Strategy

```typescript
async function fetchWithRetry(url: string, options: RequestInit, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    const response = await fetch(url, options);
    
    if (response.ok) return response;
    
    if (response.status === 401) {
      // Token expired - redirect to login
      throw new Error('AUTH_REQUIRED');
    }
    
    if (response.status === 429) {
      // Rate limited - get retry delay
      const retryAfter = response.headers.get('Retry-After') || '5';
      await sleep(parseInt(retryAfter) * 1000);
      continue;
    }
    
    if (response.status >= 500 && attempt < maxRetries - 1) {
      // Server error - exponential backoff
      await sleep(Math.pow(2, attempt) * 1000);
      continue;
    }
    
    throw new Error(`HTTP ${response.status}`);
  }
}
```

---

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| Read (GET) | 100 | 1 minute |
| Write (POST/PUT) | 30 | 1 minute |
| Sync | 10 | 1 minute |
| Chat | 10 | 1 minute |

**Headers returned:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704848460
```

---

## Pagination

All list endpoints use cursor-based pagination:

```typescript
// First page
const page1 = await fetch('/api/mobile/v1/tasks?limit=20');
const { tasks, cursor, has_more } = await page1.json();

// Next page (if has_more is true)
if (has_more) {
  const page2 = await fetch(`/api/mobile/v1/tasks?limit=20&cursor=${cursor}`);
}
```

---

## Endpoint Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/me` | Current user + workspace |
| GET | `/tasks` | List tasks |
| GET | `/tasks/:id` | Task details |
| POST | `/tasks/quick` | Quick create task |
| PUT | `/tasks/:id/status` | Update status |
| PUT | `/tasks/:id/toggle` | Toggle done/todo |
| GET | `/notifications` | List notifications |
| GET | `/notifications/count` | Unread count |
| POST | `/notifications/mark-read` | Mark as read |
| GET | `/dailylog/today` | Today's log |
| GET | `/dailylog/history` | Log history |
| GET | `/chat/threads` | Chat threads |
| GET | `/chat/history/:id` | Chat messages |
| POST | `/chat/message` | Send message |
| GET | `/sync` | Delta sync |
| POST | `/push/register` | Register device |
| DELETE | `/push/unregister` | Unregister device |

---

## TypeScript Types

```typescript
// Task (list view - lean)
interface MobileTask {
  id: string;
  title: string;
  status: 'todo' | 'in_progress' | 'done' | 'cancelled';
  priority: 'low' | 'normal' | 'high' | 'urgent';
  due_date?: string;
  assignee?: string;
  project?: string;
  updated_at: number;  // Unix timestamp
}

// Task (detail view - full)
interface MobileTaskDetail extends MobileTask {
  description?: string;
  assignee?: { id: string; name: string; avatar_url?: string };
  project?: { id: string; name: string };
  tags: string[];
  checklist: { id: string; text: string; completed: boolean }[];
  comments_count: number;
  attachments_count: number;
  created_at: string;
}

// Notification
interface MobileNotification {
  id: string;
  type: string;
  title: string;
  body?: string;
  entity_type?: string;
  entity_id?: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  is_read: boolean;
  created_at: number;
}

// Sync response
interface SyncResponse {
  server_time: string;
  changes: {
    tasks?: SyncChange<MobileTask>[];
    notifications?: SyncChange<MobileNotification>[];
  };
  deleted: {
    tasks: string[];
    notifications: string[];
  };
  has_more: boolean;
  sync_token: string;
}

interface SyncChange<T> {
  id: string;
  action: 'created' | 'updated';
  data: T;
}
```

---

*Questions? Contact the backend team.*
