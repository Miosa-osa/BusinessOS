# BusinessOS Final Taxonomy

## Part 1: Current State (What Exists)

### The Knowledge Graph

BusinessOS is a **knowledge graph** - everything connects to everything:

```
                              NODES
                           (focus areas)
                          /     |      \
                         /      |       \
                   PROJECTS  CONTEXTS  CONVERSATIONS
                      |          |           |
                    TASKS      BLOCKS     MESSAGES
                                              |
                                          ARTIFACTS
```

**Junction Tables (Many-to-Many):**
- `node_projects` - Nodes ↔ Projects
- `node_contexts` - Nodes ↔ Contexts
- `node_conversations` - Nodes ↔ Conversations
- `project_conversations` - Projects ↔ Conversations
- `task_assignees` - Tasks ↔ Team Members
- `project_members` - Projects ↔ Team Members

---

## Part 2: Current Naming Problems

### Problem #1: "Context" is Confusing

| Where | Name Used | What User Thinks |
|-------|-----------|------------------|
| Database table | `contexts` | "What's a context?" |
| API route | `/api/contexts` | "I want knowledge/docs" |
| TypeScript type | `Context` | "Is this about context window?" |
| UI route | `/knowledge` | "Oh it's the knowledge base" |
| UI sidebar | "Knowledge" | "Why does code say Context?" |

**The Problem:** Backend says "Context", frontend says "Knowledge". Users are confused.

**Also:** `Context` has a `type` field with value `CONTEXT` - so you have "Context type Context" which is ridiculous.

---

### Problem #2: "Artifact" is Technical Jargon

| Current | What It Means | What Users Say |
|---------|---------------|----------------|
| `Artifact` | AI-generated code/doc | "The thing the AI made" |
| `artifacts` table | Generated outputs | "My creations" |
| Artifact types: CODE, DOCUMENT, MARKDOWN | Different outputs | "Code", "Doc" |

**The Problem:** "Artifact" sounds like archaeology or CI/CD pipelines. Users don't naturally say "I made an artifact."

---

### Problem #3: Node Types Overlap with Modules

| Node Type | Module That Exists | Confusion |
|-----------|-------------------|-----------|
| `PROJECT` node | Projects module | "Is a Project Node different from a Project?" |
| `PERSON` node (proposed) | Team/Clients module | "Is a Person Node different from a Team Member?" |

**The Problem:** If you have a PROJECT type Node AND a Projects module, what's the difference?

**Current Reality:**
- Node type `project` = A focus area categorized as project-related
- Projects module = Actual project management with tasks

These are DIFFERENT but named the SAME.

---

### Problem #4: Context Types are a Mess

Current `contexttype` enum:
```sql
'PERSON', 'BUSINESS', 'PROJECT', 'CUSTOM', 'document', 'DOCUMENT'
```

**Problems:**
1. Mixed case: `PERSON` vs `document`
2. Duplicate: `document` AND `DOCUMENT` (why?)
3. `PROJECT` context type vs Projects module vs `project` Node type
4. `BUSINESS` context vs `business` Node type

---

### Problem #5: Inconsistent Enum Cases

| Entity | Status Enum | Case |
|--------|-------------|------|
| Task | `todo`, `in_progress`, `done` | lowercase |
| Project | `ACTIVE`, `PAUSED`, `COMPLETED` | UPPERCASE |
| Node | `healthy`, `needs_attention` | lowercase |
| Message Role | `USER`, `user`, `ASSISTANT` | BOTH |

**The Problem:** No standard. Some uppercase, some lowercase, some both.

---

### Problem #6: "Page" vs "Context" vs "Document"

In Knowledge Base, what do you call an item?
- Code says: `Context`
- UI says: "Knowledge"
- Context type includes: `document`, `DOCUMENT`
- Artifacts include: `DOCUMENT` type

**Three different things called "Document":**
1. Context with type `document` (knowledge base item)
2. Artifact with type `DOCUMENT` (AI-generated doc)
3. Uploaded files (actual documents)

---

### Problem #7: Block Confusion

`Block` appears in TWO places:
1. **Context blocks** - Content inside a knowledge page (text, code, images)
2. **Message blocks** - Structured output in chat messages

**Same name, different things.**

---

### Problem #8: Focus vs Focus Item vs This Week Focus

| Term | Where | What |
|------|-------|------|
| `focus_items` table | Dashboard | Today's quick priorities |
| `this_week_focus` | Node field (JSONB) | Weekly priorities for a node |
| "Active Node" | Node with `is_active=true` | Current focus area |

**Three different "focus" concepts.**

---

## Part 3: Proposed New Taxonomy

### Guiding Principles

1. **One name per concept** - No synonyms
2. **User language** - What would a normal person call it?
3. **No overlaps** - Node types don't duplicate module names
4. **Consistent case** - UPPERCASE in DB, lowercase in API
5. **Clear hierarchy** - Parent contains children

---

### The New Structure

```
BUSINESSOS
│
├── WORKSPACE (user's entire space)
│
├── OPERATING SYSTEM (Node hierarchy)
│   └── NODES (focus areas with 4 types)
│       ├── Type: VENTURE (was: business)
│       ├── Type: INITIATIVE (was: project)
│       ├── Type: LEARNING
│       └── Type: OPERATIONS (was: operational)
│
├── KNOWLEDGE (was: Contexts)
│   └── PAGES (was: Context)
│       ├── Type: PROFILE (was: person/business)
│       ├── Type: REFERENCE (was: custom)
│       ├── Type: TEMPLATE
│       └── Type: DOCUMENT
│
├── CHAT
│   └── CONVERSATIONS
│       └── MESSAGES
│           └── CREATIONS (was: Artifacts)
│
├── PROJECTS (unchanged)
│   └── PROJECT
│       └── TASKS
│           └── SUBTASKS
│
├── CLIENTS (unchanged)
│   └── CLIENT
│       ├── CONTACTS
│       ├── DEALS
│       └── ACTIVITIES
│
├── TEAM (unchanged)
│   └── MEMBERS
│
├── DAILY LOG
│   └── ENTRIES
│
└── DASHBOARD
    └── FOCUS (was: Focus Items)
```

---

### Rename Map: Old → New

#### Entities

| Old Name | New Name | Reason |
|----------|----------|--------|
| `Context` | **Page** | Universal term (like Notion), everyone understands |
| `Artifact` | **Creation** | Natural: "The AI made a creation" |
| `Focus Item` | **Focus** | Simpler |
| `Team Member` | **Member** | Shorter when in Team context |
| `Client Contact` | **Contact** | Shorter when in Client context |
| `Client Interaction` | **Activity** | Clearer |
| `Daily Log` | **Entry** | Each day's entry |

#### Node Types (Avoid Module Name Overlap)

| Old Type | New Type | Reason |
|----------|----------|--------|
| `business` | **venture** | Distinct from generic "business" |
| `project` | **initiative** | Distinct from Projects module |
| `learning` | **learning** | Keep (no conflict) |
| `operational` | **operations** | Cleaner |

#### Page Types (was Context Types)

| Old Type | New Type | Reason |
|----------|----------|--------|
| `PERSON` | **profile** | Profiles of people/companies |
| `BUSINESS` | **profile** | Merge with person (both are profiles) |
| `PROJECT` | **reference** | Reference docs (not a project) |
| `CUSTOM` | **reference** | General reference |
| `document` | **document** | Keep |
| `DOCUMENT` | (delete) | Remove duplicate |

#### Creation Types (was Artifact Types)

| Old Type | New Type | Reason |
|----------|----------|--------|
| `CODE` | **code** | Lowercase standard |
| `DOCUMENT` | **document** | Lowercase standard |
| `MARKDOWN` | **markdown** | Lowercase standard |
| `REACT` | **react** | Lowercase standard |
| `HTML` | **html** | Lowercase standard |
| `SVG` | **svg** | Lowercase standard |

---

### Database Renames

| Old Table | New Table | Migration |
|-----------|-----------|-----------|
| `contexts` | `pages` | `ALTER TABLE contexts RENAME TO pages` |
| `artifacts` | `creations` | `ALTER TABLE artifacts RENAME TO creations` |
| `focus_items` | `focus` | `ALTER TABLE focus_items RENAME TO focus` |
| `node_contexts` | `node_pages` | Rename junction table |

| Old Enum | New Values |
|----------|------------|
| `nodetype` | `venture`, `initiative`, `learning`, `operations` |
| `contexttype` | `profile`, `reference`, `template`, `document` |
| `artifacttype` | `code`, `document`, `markdown`, `react`, `html`, `svg` |

---

### API Route Renames

| Old Route | New Route |
|-----------|-----------|
| `/api/contexts` | `/api/pages` |
| `/api/contexts/:id` | `/api/pages/:id` |
| `/api/artifacts` | `/api/creations` |
| `/api/artifacts/:id` | `/api/creations/:id` |

**Keep backwards compatibility:**
```go
// Redirect old routes
r.GET("/api/contexts/*path", redirectToPages)
r.GET("/api/artifacts/*path", redirectToCreations)
```

---

### TypeScript Type Renames

```typescript
// Old
interface Context { ... }
interface Artifact { ... }
type ContextType = 'PERSON' | 'BUSINESS' | 'PROJECT' | 'CUSTOM' | 'document';
type ArtifactType = 'CODE' | 'DOCUMENT' | 'MARKDOWN' | 'REACT' | 'HTML' | 'SVG';

// New
interface Page { ... }
interface Creation { ... }
type PageType = 'profile' | 'reference' | 'template' | 'document';
type CreationType = 'code' | 'document' | 'markdown' | 'react' | 'html' | 'svg';
```

---

### UI Label Changes

| Location | Old Label | New Label |
|----------|-----------|-----------|
| Sidebar | "Knowledge" | "Knowledge" (keep) |
| Page title | (varies) | "Pages" |
| Create button | "New Context" | "New Page" |
| Type selector | "Person", "Business" | "Profile" |
| Chat output | "Artifact" | "Creation" |
| Save artifact | "Save Artifact" | "Save Creation" |

---

## Part 4: The Complete Hierarchy

### Visual Map

```
BUSINESSOS WORKSPACE
│
├─ 🎯 NODES (Operating System Structure)
│  │
│  ├─ VENTURE Node: "Lunivate"
│  │  ├─ INITIATIVE Node: "BusinessOS Launch"
│  │  │  └─ Links to: Project, Pages, Conversations
│  │  └─ OPERATIONS Node: "Client Delivery"
│  │
│  └─ LEARNING Node: "AI Research"
│     └─ Links to: Pages, Conversations
│
├─ 📚 KNOWLEDGE (Pages)
│  │
│  ├─ Page (profile): "Roberto Profile"
│  │  └─ Blocks: text, image, properties
│  ├─ Page (reference): "API Documentation"
│  │  └─ Blocks: text, code
│  ├─ Page (template): "Meeting Notes Template"
│  │  └─ Blocks: structured template
│  └─ Page (document): "Q1 Strategy"
│     └─ Blocks: text, embeds
│
├─ 💬 CHAT (Conversations)
│  │
│  └─ Conversation: "Building the dashboard"
│     ├─ Message (user): "Help me build..."
│     ├─ Message (assistant): "Here's the code..."
│     │  └─ Creation (code): React component
│     └─ Linked to: Node, Project
│
├─ 📁 PROJECTS
│  │
│  └─ Project: "Dashboard Redesign"
│     ├─ Task: "Build header component"
│     │  └─ Subtask: "Add navigation"
│     ├─ Task: "Implement sidebar"
│     └─ Linked to: Node, Conversations
│
├─ 👥 TEAM
│  │
│  └─ Member: "Roberto"
│     └─ Activity: "Completed task X"
│
├─ 🏢 CLIENTS
│  │
│  └─ Client: "Acme Corp"
│     ├─ Contact: "John Smith"
│     ├─ Deal: "Enterprise License"
│     └─ Activity: "Call on Jan 5"
│
├─ 📝 DAILY LOG
│  │
│  └─ Entry: "2024-01-06"
│     └─ Content, extracted actions
│
└─ 📊 DASHBOARD
   │
   └─ Focus: "Finish taxonomy doc"
      └─ Today's priorities
```

---

## Part 5: Relationship Clarifications

### Node vs Project

| Aspect | Node | Project |
|--------|------|---------|
| **What** | Focus area in your OS | Work container with tasks |
| **Purpose** | Organize your attention | Track deliverables |
| **Contains** | Links to projects, pages, conversations | Tasks, notes, members |
| **Hierarchy** | Parent-child nodes | Flat (no sub-projects) |
| **Lifecycle** | Ongoing (archived when done) | Has start/end dates |

**A Node LINKS TO Projects. They're different.**

Example:
- Node "Q1 Growth" (type: initiative)
  - Links to Project "Website Redesign"
  - Links to Project "Marketing Campaign"
  - Links to Pages about strategy
  - Links to Conversations about planning

---

### Page vs Creation

| Aspect | Page | Creation |
|--------|------|----------|
| **What** | Document you write/maintain | Output AI generates |
| **Lives in** | Knowledge module | Chat (can save to Knowledge) |
| **Created by** | User | AI assistant |
| **Purpose** | Reference, templates, profiles | Code, documents from chat |
| **Versioned** | No (edit in place) | Yes (artifact_versions) |

**A Creation can be SAVED AS a Page.**

---

### Focus vs Active Node

| Aspect | Focus | Active Node |
|--------|-------|-------------|
| **What** | Today's priority items | Current working context |
| **Scope** | Daily tasks | May be weeks/months |
| **Count** | Multiple per day | ONE at a time |
| **Table** | `focus` (was `focus_items`) | `nodes.is_active` |

---

## Part 6: Enum Standardization

### All Enums: UPPERCASE in DB, lowercase in API

**Node Types:**
```sql
-- Database
CREATE TYPE nodetype AS ENUM ('VENTURE', 'INITIATIVE', 'LEARNING', 'OPERATIONS');

-- API Response
{ "type": "venture" }
```

**Page Types:**
```sql
-- Database
CREATE TYPE pagetype AS ENUM ('PROFILE', 'REFERENCE', 'TEMPLATE', 'DOCUMENT');

-- API Response
{ "type": "profile" }
```

**Creation Types:**
```sql
-- Database
CREATE TYPE creationtype AS ENUM ('CODE', 'DOCUMENT', 'MARKDOWN', 'REACT', 'HTML', 'SVG');

-- API Response
{ "type": "code" }
```

**Status Enums (standardize all):**
```sql
-- Task Status
CREATE TYPE taskstatus AS ENUM ('TODO', 'IN_PROGRESS', 'DONE', 'CANCELLED');

-- Project Status
CREATE TYPE projectstatus AS ENUM ('ACTIVE', 'PAUSED', 'COMPLETED', 'ARCHIVED');

-- Node Health
CREATE TYPE nodehealth AS ENUM ('HEALTHY', 'NEEDS_ATTENTION', 'CRITICAL', 'NOT_STARTED');

-- Client Status
CREATE TYPE clientstatus AS ENUM ('LEAD', 'PROSPECT', 'ACTIVE', 'INACTIVE', 'CHURNED');
```

---

## Part 7: Migration Plan

### Phase 1: Backend (Non-Breaking)

1. Create new tables alongside old:
   ```sql
   CREATE TABLE pages AS SELECT * FROM contexts;
   CREATE TABLE creations AS SELECT * FROM artifacts;
   ```

2. Create views for backwards compatibility:
   ```sql
   CREATE VIEW contexts AS SELECT * FROM pages;
   CREATE VIEW artifacts AS SELECT * FROM creations;
   ```

3. Update handlers to use new names internally

### Phase 2: API (With Redirects)

1. Add new routes:
   ```
   /api/pages (new)
   /api/creations (new)
   ```

2. Add redirects from old routes:
   ```
   /api/contexts → /api/pages
   /api/artifacts → /api/creations
   ```

3. Deprecation warnings in response headers

### Phase 3: Frontend

1. Update TypeScript types
2. Update API client calls
3. Update UI labels
4. Update component names

### Phase 4: Cleanup

1. Remove old routes
2. Drop old tables
3. Update all documentation

---

## Part 8: Final Terminology Reference

### What to Call Things

| You're Talking About | Call It | NOT |
|---------------------|---------|-----|
| The OS structure | **Node** | Area, Unit |
| A node for business | **Venture** | Business node |
| A node for projects | **Initiative** | Project node |
| Knowledge base item | **Page** | Context, Document |
| AI-generated output | **Creation** | Artifact |
| Work container | **Project** | (unchanged) |
| Work item | **Task** | (unchanged) |
| Today's priority | **Focus** | Focus Item |
| Daily journal | **Entry** | Daily Log, Log |
| Person at client | **Contact** | Client Contact |
| Sales opportunity | **Deal** | (unchanged) |
| Logged interaction | **Activity** | Interaction |

### Quick Reference Card

```
NODES (4 types):
  venture, initiative, learning, operations

PAGES (4 types):
  profile, reference, template, document

CREATIONS (6 types):
  code, document, markdown, react, html, svg

STATUSES (all uppercase in DB, lowercase in API):
  Task: todo, in_progress, done, cancelled
  Project: active, paused, completed, archived
  Node Health: healthy, needs_attention, critical, not_started
  Client: lead, prospect, active, inactive, churned
```

---

## Summary

### What Changes

| Category | Changes |
|----------|---------|
| **Renamed** | Context→Page, Artifact→Creation, FocusItem→Focus |
| **Node Types** | business→venture, project→initiative, operational→operations |
| **Page Types** | person+business→profile, custom→reference, remove duplicate |
| **Enum Case** | All UPPERCASE in DB, lowercase in API |

### What Stays

| Category | Keeps |
|----------|-------|
| **Module Names** | Projects, Tasks, Clients, Team, Chat, Dashboard |
| **Core Entities** | Node, Project, Task, Client, Member, Conversation, Message |
| **Structure** | Node hierarchy, junction tables, relationships |

---

*This is the definitive taxonomy for BusinessOS. Implement this and the naming confusion ends.*
