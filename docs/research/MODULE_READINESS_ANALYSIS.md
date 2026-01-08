# Module Readiness & Universal Foundation Analysis

**Date:** January 2025
**Purpose:** Analyze current codebase to identify the best module to start with and define universal patterns all modules need

---

# EXECUTIVE SUMMARY

## Key Findings

1. **Most Ready Module:** `Projects` - has the strongest foundation with statuses, templates, team assignment, notes, and client linking
2. **Second Most Ready:** `Tasks` - has dependencies, assignees, subtasks
3. **Third Most Ready:** `Clients` - has contacts, interactions, deals

## Universal Patterns Needed

Before integrating open-source patterns, we need these **universal foundations** that EVERY module will share:

| Pattern | Current Status | Priority |
|---------|---------------|----------|
| Activity Log | Partial (system_event_logs) | P0 |
| Comments | Per-module (not universal) | P0 |
| Attachments | Per-module (not universal) | P0 |
| Tags/Labels | Per-module (inconsistent) | P1 |
| Custom Fields | JSONB (exists, not standardized) | P1 |
| Universal Links | Partial (nodes, but complex) | P1 |
| Notifications | Not implemented | P2 |
| Webhooks | Exists but basic | P2 |

---

# 1. CURRENT MODULE ANALYSIS

## 1.1 Projects Module ⭐ MOST READY

### What Exists

```
DATABASE TABLES:
├── projects                    ✅ Core table with all fields
├── project_notes              ✅ Project-specific notes
├── project_conversations      ✅ Links to AI conversations
├── project_members            ✅ Team assignment with roles
├── project_statuses           ✅ Custom status definitions
├── project_tags               ✅ Custom labels
├── project_tag_assignments    ✅ Many-to-many tagging
├── project_documents          ✅ Document relationships
└── project_templates          ✅ Reusable templates

HANDLERS:
├── projects.go                ✅ Full CRUD
├── project_members.go         ✅ Team management
└── Related handlers           ✅ Integrated

FRONTEND:
├── /projects route            ✅ Main list view
├── /projects/[id] route       ✅ Detail view
├── ProjectsAPI module         ✅ 5+ functions
└── Components                 ✅ Multiple views
```

### What's Missing for Open-Source Parity

| Feature | Plane Has | We Need |
|---------|-----------|---------|
| Issues/Tasks | Full issue system | Link tasks better |
| Cycles/Sprints | Time-boxed iterations | New table + UI |
| Roadmap View | Visual timeline | New view |
| GitHub Integration | PR/commit linking | Integration |
| Activity Stream | All changes logged | Universal pattern |
| Comments | Issue comments | Universal pattern |

### Readiness Score: 8/10

---

## 1.2 Tasks Module ⭐ VERY READY

### What Exists

```
DATABASE TABLES:
├── tasks                      ✅ Core with parent_task_id (subtasks)
├── task_assignees            ✅ Multi-assignee support
├── task_dependencies         ✅ Predecessor/successor
└── (subtasks via parent_id)  ✅ Hierarchical

HANDLERS:
├── (via dashboard, projects) ✅ Partial direct handler

FRONTEND:
├── /tasks route              ✅ Multiple views (list, board, calendar)
├── TasksAPI module           ✅ Core functions
└── 14 components             ✅ Rich UI
```

### What's Missing

| Feature | Todoist/Linear Has | We Need |
|---------|-------------------|---------|
| Natural Language | "Buy milk tomorrow" | NLP parser |
| Quick Add | Keyboard shortcut | UI improvement |
| Recurring Tasks | Config exists | Execution engine |
| Activity Log | All changes | Universal pattern |
| Comments | Task comments | Universal pattern |

### Readiness Score: 7/10

---

## 1.3 Clients Module ⭐ READY

### What Exists

```
DATABASE TABLES:
├── clients                    ✅ Core client table
├── client_contacts           ✅ Multiple contacts per client
├── client_interactions       ✅ Calls, emails, meetings, notes
└── client_deals              ✅ Deal/opportunity tracking

HANDLERS:
├── clients.go                ✅ Full CRUD

FRONTEND:
├── /clients route            ✅ List with view switcher
├── /clients/[id] route       ✅ Detail view
├── ClientsAPI module         ✅ 15+ functions
└── 8 components              ✅ Card, Kanban, Table views
```

### What's Missing

| Feature | HubSpot/Twenty Has | We Need |
|---------|-------------------|---------|
| Pipeline View | Visual deal stages | Enhanced deals UI |
| Email Tracking | Open/click tracking | Integration |
| Sequences | Automated outreach | Automation engine |
| Enrichment | Company data | External API |
| Activity Stream | All changes | Universal pattern |

### Readiness Score: 7/10

---

## 1.4 Contexts/Knowledge Module

### What Exists

```
DATABASE TABLES:
├── contexts                   ✅ Documents/profiles/knowledge
├── nodes                      ✅ Knowledge graph nodes
├── node_metrics              ✅ Performance tracking
├── node_projects             ✅ Node-project links
├── node_contexts             ✅ Node-context links
└── node_conversations        ✅ Node-conversation links

HANDLERS:
├── contexts.go               ✅ Full CRUD
├── context_tree.go           ✅ Hierarchical navigation
├── context_injection.go      ✅ AI context injection
├── document_handler.go       ✅ Document processing
├── nodes.go                  ✅ Node management

FRONTEND:
├── /knowledge-v2 route       ✅ Knowledge base v2
├── KnowledgeAPI module       ✅ 14+ functions
└── 10+ components            ✅ Graph, document views
```

### What's Missing

| Feature | Notion/AppFlowy Has | We Need |
|---------|---------------------|---------|
| Block Editor | Rich block types | Block system |
| Database Views | Table, Board, Calendar | View system |
| Real-time Collab | Yjs integration | Enhancement |
| Page Hierarchy | Nested pages | Exists, needs UI |

### Readiness Score: 6/10 (Complex, needs block editor)

---

## 1.5 Calendar Module

### What Exists

```
DATABASE TABLES:
├── (uses integration cache)   ⚠️ Google Calendar sync
└── booking_pages (implied)    ⚠️ May need creation

HANDLERS:
├── calendar.go               ✅ Event management

FRONTEND:
├── /communication/calendar   ✅ Calendar view
├── CalendarAPI module        ✅ 7 functions
└── 6 components              ✅ Widget, event cards
```

### What's Missing

| Feature | Cal.com Has | We Need |
|---------|-------------|---------|
| Native Events | Local storage | Native table |
| Booking Pages | Public scheduling | New tables |
| Team Scheduling | Round-robin | New logic |
| Buffer Times | Auto-add | New config |

### Readiness Score: 5/10 (Heavily dependent on Google integration)

---

# 2. UNIVERSAL PATTERNS NEEDED

## 2.1 Activity Log / Audit Trail (P0)

**Purpose:** Track every change across all modules for history, undo, and compliance.

### Proposed Schema

```sql
-- Universal activity log
CREATE TABLE activity_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    -- What changed
    entity_type VARCHAR(50) NOT NULL,  -- 'project', 'task', 'client', etc.
    entity_id UUID NOT NULL,

    -- The change
    action VARCHAR(50) NOT NULL,  -- 'created', 'updated', 'deleted', 'commented', 'assigned'
    field_name VARCHAR(100),      -- Which field changed (for updates)
    old_value JSONB,              -- Previous value
    new_value JSONB,              -- New value

    -- Context
    metadata JSONB,               -- Additional context
    ip_address INET,
    user_agent TEXT,

    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- Indexes
    INDEX idx_activity_entity (entity_type, entity_id),
    INDEX idx_activity_user (user_id),
    INDEX idx_activity_created (created_at DESC)
);

-- Example usage
-- When project status changes:
INSERT INTO activity_log (user_id, entity_type, entity_id, action, field_name, old_value, new_value)
VALUES ($1, 'project', $2, 'updated', 'status', '"planning"', '"in_progress"');
```

### Go Service Interface

```go
type ActivityService interface {
    LogCreate(ctx context.Context, entityType string, entityID uuid.UUID, metadata map[string]any) error
    LogUpdate(ctx context.Context, entityType string, entityID uuid.UUID, changes []FieldChange) error
    LogDelete(ctx context.Context, entityType string, entityID uuid.UUID) error
    GetEntityHistory(ctx context.Context, entityType string, entityID uuid.UUID, limit int) ([]ActivityLog, error)
    GetUserActivity(ctx context.Context, userID uuid.UUID, since time.Time) ([]ActivityLog, error)
}

type FieldChange struct {
    Field    string
    OldValue any
    NewValue any
}
```

---

## 2.2 Universal Comments (P0)

**Purpose:** Allow commenting on ANY entity (projects, tasks, clients, deals, etc.)

### Proposed Schema

```sql
-- Universal comments table
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    -- What we're commenting on
    entity_type VARCHAR(50) NOT NULL,  -- 'project', 'task', 'client', etc.
    entity_id UUID NOT NULL,

    -- The comment
    content TEXT NOT NULL,
    content_html TEXT,                  -- Rendered HTML

    -- Threading
    parent_comment_id UUID REFERENCES comments(id),

    -- Mentions
    mentions UUID[],                    -- Array of mentioned user IDs

    -- Attachments
    attachments JSONB DEFAULT '[]',     -- [{name, url, type, size}]

    -- Status
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Indexes
    INDEX idx_comments_entity (entity_type, entity_id),
    INDEX idx_comments_parent (parent_comment_id),
    INDEX idx_comments_user (user_id)
);
```

### Go Handler Pattern

```go
// Generic comment handler that works for any entity
func (h *CommentHandler) CreateComment(c *gin.Context) {
    entityType := c.Param("entityType")  // project, task, client, etc.
    entityID := c.Param("entityId")

    // Verify entity exists and user has access
    if !h.canCommentOn(c, entityType, entityID) {
        c.JSON(403, gin.H{"error": "access denied"})
        return
    }

    // Create comment
    comment := CreateCommentRequest{}
    c.BindJSON(&comment)

    // Extract mentions, render HTML, etc.
    // Save and return
}

// Routes
router.POST("/api/:entityType/:entityId/comments", h.CreateComment)
router.GET("/api/:entityType/:entityId/comments", h.ListComments)
router.PATCH("/api/comments/:id", h.UpdateComment)
router.DELETE("/api/comments/:id", h.DeleteComment)
```

---

## 2.3 Universal Attachments (P0)

**Purpose:** Attach files to ANY entity

### Proposed Schema

```sql
-- Universal attachments table
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    -- What we're attaching to
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,

    -- File info
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,

    -- Storage
    storage_provider VARCHAR(50) DEFAULT 'local',  -- 'local', 's3', 'gcs'
    storage_path TEXT NOT NULL,
    storage_url TEXT,

    -- Preview
    thumbnail_url TEXT,
    preview_url TEXT,

    -- Metadata
    metadata JSONB DEFAULT '{}',  -- dimensions, duration, etc.

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- Indexes
    INDEX idx_attachments_entity (entity_type, entity_id)
);
```

---

## 2.4 Universal Tags/Labels (P1)

**Purpose:** Consistent tagging across all modules

### Proposed Schema

```sql
-- Universal tags (workspace-level)
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    name VARCHAR(100) NOT NULL,
    color VARCHAR(7),  -- Hex color
    description TEXT,

    -- Which modules can use this tag
    applicable_types TEXT[] DEFAULT '{}',  -- ['project', 'task', 'client'] or empty for all

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE (user_id, name)
);

-- Tag assignments (polymorphic)
CREATE TABLE tag_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE (tag_id, entity_type, entity_id),
    INDEX idx_tag_assignments_entity (entity_type, entity_id)
);
```

---

## 2.5 Universal Custom Fields (P1)

**Purpose:** Standardized approach to custom fields across modules

### Proposed Schema

```sql
-- Custom field definitions
CREATE TABLE custom_field_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    -- Which entity type this applies to
    entity_type VARCHAR(50) NOT NULL,  -- 'project', 'task', 'client', etc.

    -- Field definition
    name VARCHAR(100) NOT NULL,
    field_key VARCHAR(100) NOT NULL,  -- machine-readable key
    field_type VARCHAR(50) NOT NULL,  -- 'text', 'number', 'date', 'select', 'multiselect', 'user', 'url'

    -- Configuration
    options JSONB,  -- For select/multiselect: [{value, label, color}]
    default_value JSONB,
    is_required BOOLEAN DEFAULT FALSE,

    -- Display
    position INT DEFAULT 0,
    is_visible BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE (user_id, entity_type, field_key)
);

-- Then in each entity table, use metadata JSONB column for values
-- Example: project.metadata = {"custom_budget": 50000, "custom_region": "west"}
```

---

## 2.6 Universal Links/Relationships (P1)

**Purpose:** Connect any entity to any other entity (beyond what nodes provide)

### Proposed Schema

```sql
-- Universal entity links
CREATE TABLE entity_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    -- Source entity
    source_type VARCHAR(50) NOT NULL,
    source_id UUID NOT NULL,

    -- Target entity
    target_type VARCHAR(50) NOT NULL,
    target_id UUID NOT NULL,

    -- Relationship type
    link_type VARCHAR(50) DEFAULT 'related',  -- 'related', 'blocks', 'blocked_by', 'duplicate', 'parent', 'child'

    -- Metadata
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- Prevent duplicate links
    UNIQUE (source_type, source_id, target_type, target_id, link_type),

    -- Indexes for both directions
    INDEX idx_links_source (source_type, source_id),
    INDEX idx_links_target (target_type, target_id)
);
```

---

## 2.7 Notifications (P2)

### Proposed Schema

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    -- What triggered this
    entity_type VARCHAR(50),
    entity_id UUID,
    action VARCHAR(50) NOT NULL,  -- 'mentioned', 'assigned', 'commented', 'due_soon'

    -- The notification
    title VARCHAR(255) NOT NULL,
    body TEXT,
    url TEXT,  -- Where to navigate

    -- Actor
    actor_id UUID REFERENCES users(id),

    -- Status
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),

    INDEX idx_notifications_user (user_id, is_read, created_at DESC)
);
```

---

# 3. MIGRATION PLAN

## Phase 1: Universal Foundation (Week 1-2)

```sql
-- 1. Activity Log
CREATE TABLE activity_log (...);

-- 2. Universal Comments
CREATE TABLE comments (...);

-- 3. Universal Attachments
CREATE TABLE attachments (...);

-- Functions to auto-log activity
CREATE OR REPLACE FUNCTION log_entity_change() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO activity_log (user_id, entity_type, entity_id, action, old_value, new_value)
    VALUES (
        current_setting('app.user_id')::uuid,
        TG_TABLE_NAME,
        NEW.id,
        TG_OP,
        CASE WHEN TG_OP = 'UPDATE' THEN row_to_json(OLD) ELSE NULL END,
        row_to_json(NEW)
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

## Phase 2: Apply to Projects Module (Week 2-3)

1. Add activity logging trigger to projects table
2. Create comments endpoints for projects
3. Create attachments endpoints for projects
4. Add unified activity stream to project detail page
5. Test and validate patterns work

## Phase 3: Roll Out to All Modules (Week 3-4)

1. Apply same patterns to: Tasks, Clients, Contexts
2. Create shared frontend components for activity, comments, attachments
3. Ensure consistent API structure

## Phase 4: Open-Source Integration (Week 4+)

Now that universal patterns are in place:
1. Study Plane's issue system for Projects enhancement
2. Study Twenty's CRM patterns for Clients enhancement
3. Study Cal.com for Calendar native support
4. Add new features on top of universal foundation

---

# 4. RECOMMENDATION

## Start With: Projects Module

**Why Projects First:**

1. **Most Complete Foundation** - Already has statuses, templates, team assignment
2. **Clear Enhancement Path** - Cycles, roadmap, better issue system from Plane
3. **High User Value** - Project management is core to BusinessOS
4. **Good Test Case** - Complex enough to validate universal patterns
5. **Connected to Other Modules** - Links to tasks, clients, team, contexts

## Action Plan

```
WEEK 1: Universal Foundation
├── Create activity_log table + service
├── Create comments table + handler
├── Create attachments table + handler
└── Create shared Go interfaces

WEEK 2: Projects Enhancement
├── Apply activity logging to projects
├── Add comments to projects
├── Add attachments to projects
├── Build unified activity stream component
└── Study Plane for cycles/roadmap patterns

WEEK 3: Expand to Tasks + Clients
├── Apply patterns to tasks
├── Apply patterns to clients
├── Create shared frontend components
└── Test cross-module linking

WEEK 4: Advanced Features
├── Add cycles/sprints (from Plane)
├── Add roadmap view (from Plane)
├── Enhance deals pipeline (from Twenty)
└── Add calendar native support (from Cal.com)
```

---

# 5. FRONTEND SHARED COMPONENTS NEEDED

## Universal Components to Create

```
src/lib/components/universal/
├── ActivityStream.svelte       # Shows activity log for any entity
├── Comments.svelte             # Comment thread for any entity
├── CommentInput.svelte         # Comment input with mentions
├── Attachments.svelte          # File attachments for any entity
├── AttachmentUpload.svelte     # Drag-drop file upload
├── TagPicker.svelte            # Universal tag picker
├── EntityLink.svelte           # Link to any entity
├── EntityLinkPicker.svelte     # Search and link entities
└── index.ts                    # Exports
```

## API Patterns

```typescript
// Universal API pattern
interface UniversalAPI {
  // Activity
  getActivity(entityType: string, entityId: string): Promise<Activity[]>;

  // Comments
  getComments(entityType: string, entityId: string): Promise<Comment[]>;
  createComment(entityType: string, entityId: string, content: string): Promise<Comment>;
  updateComment(commentId: string, content: string): Promise<Comment>;
  deleteComment(commentId: string): Promise<void>;

  // Attachments
  getAttachments(entityType: string, entityId: string): Promise<Attachment[]>;
  uploadAttachment(entityType: string, entityId: string, file: File): Promise<Attachment>;
  deleteAttachment(attachmentId: string): Promise<void>;

  // Tags
  getTags(): Promise<Tag[]>;
  assignTag(entityType: string, entityId: string, tagId: string): Promise<void>;
  removeTag(entityType: string, entityId: string, tagId: string): Promise<void>;

  // Links
  getLinks(entityType: string, entityId: string): Promise<EntityLink[]>;
  createLink(source: EntityRef, target: EntityRef, linkType: string): Promise<EntityLink>;
  deleteLink(linkId: string): Promise<void>;
}
```

---

# CONCLUSION

## The Order of Operations

1. **Build Universal Patterns First** - Activity, Comments, Attachments, Tags
2. **Apply to Projects Module** - Validate patterns work
3. **Roll Out to All Modules** - Tasks, Clients, Contexts, etc.
4. **Then Add Open-Source Features** - Cycles, Roadmaps, Pipelines

This approach ensures:
- Consistent patterns across all modules
- Easier future module additions
- Clean foundation for open-source feature adoption
- Better data integrity and audit trail
- Unified user experience

**The Projects module is ready. The universal foundation is the key blocker.**
