# BusinessOS Platform Taxonomy

## Executive Summary: The Naming Problem

The platform has **significant terminology inconsistencies** caused by organic growth and AI-generated naming. This document establishes the **foundational taxonomy** that must be followed.

---

## Platform Hierarchy

```
PLATFORM: BusinessOS
    │
    ├── MODULES (Top-level navigation items)
    │   ├── Dashboard
    │   ├── Chat
    │   ├── Tasks
    │   ├── Communication Hub
    │   ├── Projects
    │   ├── Team
    │   ├── Clients
    │   ├── Knowledge Base    ← PROBLEM: Called "Contexts" in code
    │   ├── Nodes
    │   ├── Daily Log
    │   ├── Usage
    │   ├── Integrations
    │   └── Settings
    │
    └── Within each MODULE are ENTITIES (the things inside)
```

---

## CRITICAL NAMING ISSUES FOUND

### Issue #1: Knowledge Base Identity Crisis

| Layer | Current Name | Should Be |
|-------|--------------|-----------|
| **UI Route** | `/knowledge` | `/knowledge` ✓ |
| **UI Label** | "Knowledge" | "Knowledge Base" |
| **API Route** | `/api/contexts` | `/api/knowledge` |
| **API Folder** | `contexts/` | `knowledge/` |
| **Database Table** | `contexts` | `knowledge_entries` |
| **Go Handler** | `contexts.go` | `knowledge.go` |
| **TypeScript Type** | `Context` | `KnowledgeEntry` |

**What's INSIDE the Knowledge Base?**

Current code calls them "Contexts" with types:
- `person` - Person profiles
- `business` - Business info
- `project` - Project docs (confusing with Projects module!)
- `custom` - Custom entries
- `document` - Documents

**RECOMMENDED: Rename to "Knowledge Entries" with types:**
- `profile` - Person/company profiles
- `reference` - Reference documents
- `template` - Reusable templates
- `note` - Notes and custom content

---

### Issue #2: Artifacts vs Documents vs Contexts

**Three different things that sound similar:**

| Entity | What It Is | Where It Lives | When Created |
|--------|-----------|----------------|--------------|
| **Artifact** | Generated code/content from AI chat | `artifacts` table | During conversations |
| **Context** (should be Knowledge Entry) | Reference documents in Knowledge Base | `contexts` table | User creates manually |
| **Document** | Uploaded files for RAG/search | `/uploads/documents` | User uploads |

**The Confusion:**
- "Context" type can be `document`
- "Artifact" type can be `DOCUMENT`
- Uploaded files are also called "Documents"

**RECOMMENDED NAMING:**
| Current | Rename To | Purpose |
|---------|-----------|---------|
| Artifact | **Artifact** (keep) | AI-generated outputs |
| Context | **Knowledge Entry** | Knowledge base items |
| Document (upload) | **File** | Uploaded files |

---

### Issue #3: Nodes Confusion

**What is a "Node"?**

Current: Nodes are hierarchical organizational units representing the "Operating System" structure.

**Node Types (current):**
- `BUSINESS` - Business units/departments
- `PROJECT` - Project initiatives (confusing with Projects module!)
- `LEARNING` - Learning tracks
- `OPERATIONAL` - Operations

**The Problem:**
- "Node type: PROJECT" vs "Projects module" - what's the difference?
- Term "Node" is too generic
- Users don't understand what Nodes are for

**RECOMMENDED: Rename to "Operating System Units" or "OS Units"**

| Current | Rename To | Purpose |
|---------|-----------|---------|
| Node | **OS Unit** | A unit in the operating system structure |
| Node type: BUSINESS | **Division** | Business division |
| Node type: PROJECT | **Initiative** | Strategic initiative |
| Node type: LEARNING | **Learning Track** | Learning path |
| Node type: OPERATIONAL | **Operation** | Operational unit |

---

### Issue #4: Status/Priority Case Inconsistency

**Task Status (lowercase):**
```
'todo', 'in_progress', 'done', 'cancelled'
```

**Project Status (UPPERCASE):**
```
'ACTIVE', 'PAUSED', 'COMPLETED', 'ARCHIVED'
```

**Message Role (MIXED):**
```
'USER', 'user', 'ASSISTANT', 'assistant'
```

**RECOMMENDED: All UPPERCASE in database, lowercase in API responses**

---

### Issue #5: Three Parallel Task Systems

| System | Tables | Purpose |
|--------|--------|---------|
| Native Tasks | `tasks` | BusinessOS tasks |
| Google Tasks | `google_task_lists`, `google_tasks` | Synced from Google |
| Microsoft To Do | `microsoft_todo_lists`, `microsoft_todo_tasks` | Synced from Microsoft |

**The Problem:** No clear strategy for syncing/merging these

**RECOMMENDED:** Native `tasks` is canonical, others sync INTO it

---

### Issue #6: Duplicate Communication Tables

| Type | Generic Tables | Provider-Specific |
|------|---------------|-------------------|
| Channels | `channels`, `channel_messages` | `slack_channels`, `slack_messages` |
| Calendar | `calendar_events` | `microsoft_calendar_events` |
| Email | `emails` | `gmail_emails` |

**RECOMMENDED:** Provider-specific tables sync INTO generic tables

---

## FOUNDATIONAL TAXONOMY

### Module Definitions

| Module | Purpose | Contains (Entities) |
|--------|---------|---------------------|
| **Dashboard** | Overview & quick actions | Focus Items, Widgets |
| **Chat** | AI conversations | Conversations, Messages |
| **Tasks** | Task management | Tasks, Subtasks |
| **Communication** | Email, calendar, messaging | Emails, Events, Channel Messages |
| **Projects** | Project management | Projects, Project Notes, Project Members |
| **Team** | Team management | Team Members, Activities |
| **Clients** | CRM | Clients, Contacts, Deals, Interactions |
| **Knowledge Base** | Reference materials | Knowledge Entries (NOT "Contexts") |
| **Nodes** → **OS Units** | Operating system structure | OS Units, Unit Metrics |
| **Daily Log** | Activity journal | Daily Logs, Log Entries |
| **Usage** | Analytics | Usage Metrics |
| **Integrations** | Third-party connections | Integration Configs, Credentials |
| **Settings** | Configuration | User Settings, AI Settings |

---

### Entity Naming Convention

**Format:** `{Module}{Entity}` in code, `{entity}` in API

| Module | Entity | Database Table | API Route | TypeScript Type |
|--------|--------|----------------|-----------|-----------------|
| Chat | Conversation | `conversations` | `/api/conversations` | `Conversation` |
| Chat | Message | `messages` | (nested) | `Message` |
| Tasks | Task | `tasks` | `/api/tasks` | `Task` |
| Projects | Project | `projects` | `/api/projects` | `Project` |
| Projects | Note | `project_notes` | (nested) | `ProjectNote` |
| Team | Member | `team_members` | `/api/team` | `TeamMember` |
| Clients | Client | `clients` | `/api/clients` | `Client` |
| Clients | Contact | `client_contacts` | (nested) | `ClientContact` |
| Clients | Deal | `client_deals` | (nested) | `Deal` |
| **Knowledge** | **Entry** | `knowledge_entries`* | `/api/knowledge` | `KnowledgeEntry` |
| **OS Units** | **Unit** | `os_units`* | `/api/os-units` | `OSUnit` |
| Artifacts | Artifact | `artifacts` | `/api/artifacts` | `Artifact` |

*Requires migration to rename

---

### Enum Standardization

**All enums should be:**
- UPPERCASE in database
- lowercase in API JSON responses
- PascalCase in TypeScript

**Task Status:**
```
DB: 'TODO', 'IN_PROGRESS', 'DONE', 'CANCELLED'
API: 'todo', 'in_progress', 'done', 'cancelled'
TS: TaskStatus.Todo, TaskStatus.InProgress, etc.
```

**Project Status:**
```
DB: 'ACTIVE', 'PAUSED', 'COMPLETED', 'ARCHIVED'
API: 'active', 'paused', 'completed', 'archived'
TS: ProjectStatus.Active, etc.
```

**Priority (Universal):**
```
DB: 'CRITICAL', 'HIGH', 'MEDIUM', 'LOW'
API: 'critical', 'high', 'medium', 'low'
TS: Priority.Critical, etc.
```

---

### Knowledge Entry Types (was Context Types)

| Current | Rename To | Icon | Purpose |
|---------|-----------|------|---------|
| `person` | `profile` | user | Person/company profiles |
| `business` | `reference` | book | Reference documents |
| `project` | `template` | copy | Reusable templates |
| `custom` | `note` | file-text | Notes and custom content |
| `document` | `document` | file | Formal documents |

---

### OS Unit Types (was Node Types)

| Current | Rename To | Icon | Purpose |
|---------|-----------|------|---------|
| `BUSINESS` | `division` | building | Business divisions |
| `PROJECT` | `initiative` | rocket | Strategic initiatives |
| `LEARNING` | `track` | graduation-cap | Learning tracks |
| `OPERATIONAL` | `operation` | cog | Operations |

---

## MIGRATION PRIORITY

### Phase 1: Critical (API Breaking)
1. Rename `/api/contexts` → `/api/knowledge` (with redirect)
2. Rename `Context` type → `KnowledgeEntry`
3. Standardize all enum cases

### Phase 2: Important (DB Migration)
1. Rename `contexts` table → `knowledge_entries`
2. Rename `nodes` table → `os_units`
3. Consolidate task systems

### Phase 3: Cleanup
1. Rename all internal references
2. Update all documentation
3. Update frontend labels

---

## RELATIONSHIP MAP

```
User
├── Conversations ──────────► Messages
│   └── linked to ─────────► Knowledge Entry (optional)
│
├── Projects ──────────────► Tasks
│   ├── has ───────────────► Project Notes
│   ├── has ───────────────► Project Members
│   └── linked from ───────► OS Units
│
├── Knowledge Base
│   └── Knowledge Entries ─► Child Entries (hierarchical)
│       └── can link to ───► OS Units, Projects
│
├── OS Units (hierarchical)
│   ├── links to ──────────► Projects
│   ├── links to ──────────► Knowledge Entries
│   └── links to ──────────► Conversations
│
├── Artifacts
│   ├── created from ──────► Conversation
│   └── can save to ───────► Knowledge Entry
│
├── Clients ───────────────► Contacts, Deals, Interactions
│
└── Team ──────────────────► Team Members
```

---

## FILE REFERENCE

| What | Current Location | Notes |
|------|------------------|-------|
| DB Schema | `backend-go/internal/database/schema.sql` | Source of truth for tables |
| API Routes | `backend-go/internal/handlers/handlers.go` | All route definitions |
| Context Handler | `backend-go/internal/handlers/contexts.go` | Should be `knowledge.go` |
| Node Handler | `backend-go/internal/handlers/nodes.go` | Should be `os_units.go` |
| Frontend Routes | `frontend/src/routes/(app)/` | URL structure |
| API Types | `frontend/src/lib/api/*/types.ts` | TypeScript definitions |
| Navigation | `frontend/src/routes/(app)/+layout.svelte` | Sidebar labels |

---

## SUMMARY: What to Call Things

| You Want To Say | Call It | NOT |
|-----------------|---------|-----|
| The docs/knowledge stuff | **Knowledge Entry** | Context |
| The OS structure things | **OS Unit** | Node |
| AI-generated code/docs | **Artifact** | Document |
| Uploaded files | **File** | Document |
| Main sections of the app | **Module** | Page, Section |
| Things inside a module | **Entity** | Item, Object |
| A business unit in OS | **Division** | Node, Business Node |
| A project in OS structure | **Initiative** | Node, Project Node |

---

*Document generated from codebase analysis. Last updated: 2025-01-06*
