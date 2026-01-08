# Comments & Mentions System - Architecture Flow

This document explains how the comments and mentions system was built and how it integrates with the existing notification infrastructure.

## Overview

The comments system provides:
- **Polymorphic comments** - Can attach to any entity (tasks, projects, notes, etc.)
- **Threaded replies** - Support for nested comment threads
- **@mentions** - Parse and notify users mentioned in comments
- **Reactions** - Emoji reactions on comments
- **Real-time notifications** - SSE, Push, and Email notifications

---

## Database Schema

### Tables Created

```
supabase/migrations/20260108040000_comments_mentions.sql
```

#### 1. `comments` table
```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,      -- Author
    entity_type VARCHAR(50) NOT NULL,   -- 'task', 'project', 'note', etc.
    entity_id UUID NOT NULL,            -- ID of the entity
    content TEXT NOT NULL,              -- Comment body (supports markdown)
    parent_id UUID,                     -- For replies (NULL = top-level)
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMPTZ,
    is_deleted BOOLEAN DEFAULT FALSE,   -- Soft delete
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

#### 2. `entity_mentions` table
```sql
CREATE TABLE entity_mentions (
    id UUID PRIMARY KEY,
    source_type VARCHAR(50) NOT NULL,   -- 'comment', 'task_description', etc.
    source_id UUID NOT NULL,            -- ID of the source (comment ID)
    mentioned_user_id VARCHAR(255),     -- Who was mentioned
    mention_text VARCHAR(255),          -- The actual @mention text
    position_in_text INTEGER,           -- Position in content
    entity_type VARCHAR(50),            -- Context entity type
    entity_id UUID,                     -- Context entity ID
    mentioned_by VARCHAR(255),          -- Who made the mention
    notified BOOLEAN DEFAULT FALSE,     -- Track notification status
    notified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ
);
```

#### 3. `comment_reactions` table
```sql
CREATE TABLE comment_reactions (
    id UUID PRIMARY KEY,
    comment_id UUID NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    emoji VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ,
    UNIQUE(comment_id, user_id, emoji)  -- One reaction per emoji per user
);
```

---

## Data Flow

### Creating a Comment

```
┌──────────────┐     ┌─────────────────┐     ┌──────────────────┐
│   Frontend   │────▶│  Comment API    │────▶│  CommentService  │
│  POST /api/  │     │  Handlers       │     │                  │
│  comments    │     └─────────────────┘     └────────┬─────────┘
└──────────────┘                                      │
                                                      ▼
                              ┌────────────────────────────────────┐
                              │         CreateComment()            │
                              │                                    │
                              │  1. Insert comment into DB         │
                              │  2. Parse @mentions from content   │
                              │  3. Store mentions in DB           │
                              │  4. Trigger notifications (async)  │
                              │  5. Return comment with author     │
                              └────────────────────────────────────┘
                                                      │
                          ┌───────────────────────────┼───────────────────────────┐
                          ▼                           ▼                           ▼
                ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
                │ OnTaskComment() │       │   OnMention()   │       │OnCommentReply() │
                │                 │       │                 │       │                 │
                │ Notify entity   │       │ Notify each     │       │ Notify parent   │
                │ owner           │       │ @mentioned user │       │ comment author  │
                └────────┬────────┘       └────────┬────────┘       └────────┬────────┘
                         │                         │                         │
                         └─────────────────────────┼─────────────────────────┘
                                                   ▼
                                    ┌──────────────────────────┐
                                    │   NotificationService    │
                                    │                          │
                                    │  - SSE (real-time)       │
                                    │  - Web Push (background) │
                                    │  - Email (async)         │
                                    │  - Store in DB           │
                                    └──────────────────────────┘
```

### Mention Parsing

The `ParseMentions()` function supports two formats:

1. **Markdown-style**: `@[Display Name](user_id)`
   - Used by rich text editors
   - Contains user ID for direct lookup

2. **Simple**: `@username`
   - Fallback for plain text
   - Requires username resolution

```go
// Example content:
"Hey @[John Doe](usr_123) can you review this? cc @jane"

// Parsed mentions:
[
  { UserID: "usr_123", Username: "John Doe", Position: 4 },
  { UserID: "", Username: "jane", Position: 48 }  // Needs resolution
]
```

---

## API Endpoints

### Comments CRUD

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/comments` | Create a comment |
| `GET` | `/api/comments?entity_type=task&entity_id=uuid` | List comments for entity |
| `GET` | `/api/comments/:id` | Get single comment |
| `PUT` | `/api/comments/:id` | Update comment content |
| `DELETE` | `/api/comments/:id` | Soft-delete comment |

### Reactions

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/comments/:id/reactions` | Add reaction |
| `DELETE` | `/api/comments/:id/reactions/:emoji` | Remove reaction |

### Entity-Specific Shortcuts

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/tasks/:id/comments` | Get task comments |
| `POST` | `/api/tasks/:id/comments` | Add task comment |
| `GET` | `/api/projects/:id/comments` | Get project comments |
| `POST` | `/api/projects/:id/comments` | Add project comment |

---

## Notification Types

### 1. Task Comment (`task_comment`)
Triggered when someone comments on a task you own/are assigned to.

```json
{
  "type": "task_comment",
  "title": "New comment on \"Fix login bug\"",
  "body": "John: Can we prioritize this?",
  "data": {
    "task_id": "uuid",
    "comment_id": "uuid"
  }
}
```

### 2. Mention (`mention`)
Triggered when you're @mentioned in a comment.

```json
{
  "type": "mention",
  "title": "John mentioned you",
  "body": "...can @you review this PR?",
  "data": {
    "source_type": "comment",
    "source_id": "uuid",
    "entity_type": "task",
    "entity_id": "uuid"
  }
}
```

### 3. Comment Reply (`comment_reply`)
Triggered when someone replies to your comment.

```json
{
  "type": "comment_reply",
  "title": "John replied to your comment",
  "body": "Good point, I'll update it.",
  "data": {
    "parent_comment_id": "uuid",
    "reply_id": "uuid"
  }
}
```

---

## File Structure

```
backend-go/
├── cmd/server/main.go                    # CommentService initialization
├── internal/
│   ├── database/
│   │   ├── queries/comments.sql          # SQLC query definitions
│   │   ├── schema.sql                    # Table definitions for SQLC
│   │   └── sqlc/comments.sql.go          # Generated query code
│   ├── handlers/
│   │   ├── handlers.go                   # Route registration
│   │   └── comment_handlers.go           # HTTP handlers
│   └── services/
│       ├── comment_service.go            # Business logic
│       └── notification_triggers.go      # OnTaskComment, OnMention, etc.
└── supabase/migrations/
    └── 20260108040000_comments_mentions.sql  # Database migration
```

---

## Integration Points

### With Existing Notification System

The comment service uses the existing `NotificationService` which provides:

- **SSE Broadcaster** - Real-time updates to connected clients
- **Web Push** - Background notifications via VAPID
- **Email** - Async email delivery via Resend
- **Database Storage** - Persistent notification history

### With Frontend

The frontend can:
1. Subscribe to SSE for real-time comment updates
2. Use the comments API to fetch/create comments
3. Display @mention autocomplete using user search
4. Show notification badges for new mentions

---

## Usage Example

```typescript
// Frontend: Create a comment with mention
const response = await fetch('/api/tasks/123/comments', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    content: 'Hey @[Alice](usr_456), can you review this task?'
  })
});

// Result: 
// - Comment created
// - Mention stored in entity_mentions
// - Alice receives notification via SSE + Push + Email
```

---

## Future Enhancements

1. **Mention Autocomplete API** - Endpoint to search users for @mention suggestions
2. **Unread Comments** - Track which comments a user has seen
3. **Comment Subscriptions** - Subscribe to entity comments without being mentioned
4. **Rich Text** - Support for images, code blocks, etc.
5. **Comment Search** - Full-text search across comments
