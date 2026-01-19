# BusinessOS Foundational Taxonomy v2

## Learning from Notion's Simplicity

[Notion's data model](https://www.notion.com/blog/data-model-behind-notion) is elegant: **Everything is a block**. Pages are blocks. Database rows are blocks. Text is a block. This unified model makes it intuitive.

Their hierarchy:
```
Workspace → Pages → Blocks
            ↓
         Databases (collections of pages)
```

Simple [terminology](https://www.makeuseof.com/beginners-guide-notion-terminology/):
- **Workspace** - Your account/space
- **Page** - Any document (can nest infinitely)
- **Block** - Any content unit (text, image, embed, even a page)
- **Database** - Structured collection of pages

---

## BusinessOS Foundational Model

### The Core Principle

**Everything the user creates is a PAGE or an ITEM.**

```
BUSINESSOS WORKSPACE
│
├── PAGES (Knowledge, Documents)
│   └── Contains: Blocks
│
├── CONVERSATIONS (Chat)
│   └── Contains: Messages → Generate: Creations
│
├── PROJECTS (Work Containers)
│   └── Contains: Tasks, Notes
│
├── AREAS (Business Structure - was "Nodes")
│   └── Contains: Child Areas, Links to Pages/Projects
│
└── RECORDS (CRM, Team)
    └── Clients, Contacts, Team Members
```

---

## The Clean Taxonomy

### Tier 1: Workspace Level

| Concept | Name | What It Is |
|---------|------|------------|
| The Platform | **BusinessOS** | The entire system |
| User's Space | **Workspace** | Everything belonging to one user |
| Main Sections | **Modules** | Dashboard, Chat, Projects, etc. |

---

### Tier 2: Module-Level Entities

#### KNOWLEDGE MODULE (was "Contexts")

| Old Name | **New Name** | What User Does |
|----------|--------------|----------------|
| Context | **Page** | User creates a Page in Knowledge |
| Context Type | **Page Type** | profile, reference, template, note |
| Block | **Block** | Content inside a Page |
| Parent Context | **Parent Page** | Nesting for hierarchy |

**User Language:**
- "I created a **Page** in my Knowledge Base"
- "This Page has **Blocks** of text and code"
- "I made it a **Template** so I can reuse it"

**Hierarchy:**
```
Knowledge Base (Module)
└── Page
    ├── Blocks (text, code, images, embeds)
    ├── Properties (custom fields)
    └── Child Pages (nested)
```

---

#### CHAT MODULE

| Old Name | **New Name** | What User Does |
|----------|--------------|----------------|
| Conversation | **Conversation** | User starts a Conversation |
| Message | **Message** | Each back-and-forth |
| Artifact | **Creation** | AI-generated output (code, doc, etc.) |

**User Language:**
- "I started a **Conversation** with the AI"
- "The AI generated a **Creation** (code snippet)"
- "I saved the Creation to my **Knowledge Base**"

**Hierarchy:**
```
Chat (Module)
└── Conversation
    ├── Messages (user + assistant)
    └── Creations (generated outputs)
        └── Can be saved as → Page (in Knowledge)
```

**Why "Creation" not "Artifact":**
- "Artifact" sounds archaeological/technical
- "Creation" is what it is - something the AI created
- Alternatives considered: Output, Generation, Result
- "Creation" is most natural: "The AI made a creation"

---

#### PROJECTS MODULE

| Old Name | **New Name** | What User Does |
|----------|--------------|----------------|
| Project | **Project** | User creates a Project |
| Task | **Task** | Work item inside Project |
| Project Note | **Note** | Comment on Project |
| Subtask | **Subtask** | Nested Task |

**User Language:**
- "I created a **Project** for the website redesign"
- "I added **Tasks** to the Project"
- "This Task has **Subtasks**"

**Hierarchy:**
```
Projects (Module)
└── Project
    ├── Tasks
    │   └── Subtasks
    ├── Notes
    └── Linked Conversations
```

---

#### AREAS MODULE (was "Nodes")

| Old Name | **New Name** | What User Does |
|----------|--------------|----------------|
| Node | **Area** | User creates an Area of focus |
| Node Type: BUSINESS | **Business Area** | Business division |
| Node Type: PROJECT | **Initiative** | Strategic initiative |
| Node Type: LEARNING | **Learning Track** | Learning focus |
| Node Type: OPERATIONAL | **Operation** | Operational focus |
| Node Health | **Status** | How the Area is doing |

**User Language:**
- "I created an **Area** for my consulting business"
- "This Area has a child **Area** for marketing"
- "The Area status is **healthy**"

**Why "Area" not "Node":**
- "Node" is a programmer term (graph theory)
- "Area" is natural: "What area are you focusing on?"
- Alternatives considered: Zone, Domain, Space, Unit
- "Area" wins because it's simple and intuitive

**Hierarchy:**
```
Areas (Module) - "My Operating System"
└── Area
    ├── Child Areas (nested)
    ├── Linked Pages (knowledge)
    ├── Linked Projects
    └── Linked Conversations
```

---

#### TASKS MODULE

| Old Name | **New Name** | What User Does |
|----------|--------------|----------------|
| Task | **Task** | User creates a Task |
| Focus Item | **Focus** | Today's priority (different!) |

**Important Distinction:**
- **Task** = Persistent work item with status, dates, project
- **Focus** = Daily priority (ephemeral, just for today)

**User Language:**
- "I have 47 **Tasks** across my projects"
- "My **Focus** for today is these 3 things"

---

#### CLIENTS MODULE (CRM)

| Old Name | **New Name** | What User Does |
|----------|--------------|----------------|
| Client | **Client** | Company or person |
| Contact | **Contact** | Person at a Client |
| Deal | **Deal** | Sales opportunity |
| Interaction | **Activity** | Logged communication |

**Hierarchy:**
```
Clients (Module)
└── Client
    ├── Contacts (people)
    ├── Deals (opportunities)
    └── Activities (calls, emails, meetings)
```

---

#### TEAM MODULE

| Old Name | **New Name** | What User Does |
|----------|--------------|----------------|
| Team Member | **Member** | Person on the team |

---

#### DAILY LOG MODULE

| Old Name | **New Name** | What User Does |
|----------|--------------|----------------|
| Daily Log | **Entry** | Daily journal entry |

**User Language:**
- "I wrote today's **Entry** in my Daily Log"

---

### Tier 3: Content Level (Inside Things)

| Container | Contains | Name |
|-----------|----------|------|
| Page | Content units | **Block** |
| Conversation | Exchanges | **Message** |
| Message | AI outputs | **Creation** |
| Project | Work items | **Task** |
| Task | Child items | **Subtask** |
| Client | People | **Contact** |
| Area | Sub-areas | **Child Area** |

---

## Complete Naming Map

### Code → User-Facing Name

| Database Table | API Route | TypeScript Type | **User Sees** |
|---------------|-----------|-----------------|---------------|
| `contexts` | `/api/knowledge` | `Page` | **Page** |
| `conversations` | `/api/conversations` | `Conversation` | **Conversation** |
| `messages` | (nested) | `Message` | **Message** |
| `artifacts` | `/api/creations` | `Creation` | **Creation** |
| `nodes` | `/api/areas` | `Area` | **Area** |
| `projects` | `/api/projects` | `Project` | **Project** |
| `tasks` | `/api/tasks` | `Task` | **Task** |
| `clients` | `/api/clients` | `Client` | **Client** |
| `client_contacts` | (nested) | `Contact` | **Contact** |
| `client_deals` | (nested) | `Deal` | **Deal** |
| `team_members` | `/api/team` | `Member` | **Member** |
| `daily_logs` | `/api/daily` | `Entry` | **Entry** |
| `focus_items` | `/api/focus` | `Focus` | **Focus** |

---

## Page Types (Knowledge Base)

| Current Code | **New Name** | Icon | Use Case |
|--------------|--------------|------|----------|
| `PERSON` | **Profile** | user | Person/company info |
| `BUSINESS` | **Reference** | book | Reference docs |
| `PROJECT` | **Template** | copy | Reusable templates |
| `CUSTOM` | **Note** | file-text | General notes |
| `DOCUMENT` | **Document** | file | Formal documents |

---

## Area Types (Operating System)

| Current Code | **New Name** | Icon | Use Case |
|--------------|--------------|------|----------|
| `BUSINESS` | **Business** | building | Business divisions |
| `PROJECT` | **Initiative** | rocket | Strategic initiatives |
| `LEARNING` | **Learning** | graduation-cap | Learning tracks |
| `OPERATIONAL` | **Operations** | settings | Operational areas |

---

## Status/Priority Naming

### Universal Priority
```
critical → high → medium → low
```

### Task Status
```
todo → in_progress → done → cancelled
```

### Project Status
```
active → paused → completed → archived
```

### Area Status (Health)
```
healthy → needs_attention → critical → not_started
```

### Client Status
```
lead → prospect → active → inactive → churned
```

---

## The Full Hierarchy Visual

```
BUSINESSOS WORKSPACE
│
├─ 📊 DASHBOARD
│  └─ Focus (today's priorities)
│
├─ 💬 CHAT
│  └─ Conversation
│     ├─ Message (user)
│     ├─ Message (assistant)
│     │  └─ Creation (generated code/doc)
│     └─ Linked Page (from Knowledge)
│
├─ 📚 KNOWLEDGE BASE
│  └─ Page
│     ├─ Block (text)
│     ├─ Block (code)
│     ├─ Block (image)
│     ├─ Properties (custom fields)
│     └─ Child Page (nested)
│
├─ 🎯 AREAS (Operating System)
│  └─ Area
│     ├─ Child Area
│     ├─ → Links to Pages
│     ├─ → Links to Projects
│     └─ → Links to Conversations
│
├─ 📁 PROJECTS
│  └─ Project
│     ├─ Task
│     │  └─ Subtask
│     ├─ Note
│     └─ → Linked Conversations
│
├─ ✅ TASKS
│  └─ Task
│     ├─ Subtask
│     ├─ Assignee (→ Member)
│     └─ Project (→ Project)
│
├─ 👥 TEAM
│  └─ Member
│     └─ Activity
│
├─ 🏢 CLIENTS
│  └─ Client
│     ├─ Contact
│     ├─ Deal
│     └─ Activity
│
├─ 📝 DAILY LOG
│  └─ Entry
│
├─ 📧 COMMUNICATION
│  ├─ Email (synced)
│  ├─ Event (calendar)
│  └─ Channel Message (Slack)
│
└─ 🔌 INTEGRATIONS
   └─ Connection (Google, Slack, etc.)
```

---

## Migration Checklist

### Phase 1: User-Facing (No Code Changes)
- [ ] Update all UI labels to new names
- [ ] Update help text and tooltips
- [ ] Update any documentation

### Phase 2: API Routes (Breaking Changes)
- [ ] `/api/contexts` → `/api/knowledge`
- [ ] `/api/nodes` → `/api/areas`
- [ ] `/api/artifacts` → `/api/creations`
- [ ] Add redirects for old routes

### Phase 3: Types & Code
- [ ] `Context` type → `Page`
- [ ] `Node` type → `Area`
- [ ] `Artifact` type → `Creation`
- [ ] Update all internal references

### Phase 4: Database (Final)
- [ ] `contexts` table → `pages`
- [ ] `nodes` table → `areas`
- [ ] `artifacts` table → `creations`

---

## Quick Reference Card

| You Say | It Means | In Code (current) |
|---------|----------|-------------------|
| "Page" | Knowledge Base document | `Context` |
| "Conversation" | Chat thread | `Conversation` |
| "Creation" | AI-generated output | `Artifact` |
| "Area" | OS focus area | `Node` |
| "Project" | Work container | `Project` |
| "Task" | Work item | `Task` |
| "Focus" | Today's priority | `FocusItem` |
| "Client" | Customer/prospect | `Client` |
| "Contact" | Person at client | `ClientContact` |
| "Member" | Team member | `TeamMember` |
| "Entry" | Daily log entry | `DailyLog` |
| "Block" | Content unit | `Block` |

---

## Why These Names?

| Old | New | Reasoning |
|-----|-----|-----------|
| Context | **Page** | Notion uses "Page", everyone understands it |
| Artifact | **Creation** | Natural language: "The AI made a creation" |
| Node | **Area** | Natural: "What area are you working on?" |
| Focus Item | **Focus** | Simpler, obvious |
| Client Contact | **Contact** | Just "Contact" when in Client context |
| Team Member | **Member** | Just "Member" when in Team context |

---

*This is the foundational taxonomy for BusinessOS. All code, documentation, and UI should follow these naming conventions.*

Sources:
- [Notion's Data Model](https://www.notion.com/blog/data-model-behind-notion)
- [Notion Terminology Guide](https://www.makeuseof.com/beginners-guide-notion-terminology/)
- [Notion Glossary](https://noteforms.com/notion-glossary)
