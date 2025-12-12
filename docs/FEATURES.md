# Business OS - Features Documentation

> Last Updated: December 12, 2025

This document provides detailed information about all features in Business OS.

---

## Table of Contents

1. [Document Editor](#document-editor)
2. [Desktop Mode](#desktop-mode)
3. [Clients System](#clients-system)
4. [AI Chat](#ai-chat)
5. [Context System](#context-system)
6. [Projects & Tasks](#projects--tasks)
7. [Artifacts](#artifacts)

---

## Document Editor

Business OS includes a Notion-like block-based document editor for creating and organizing business documents.

### Block Types

| Block Type | Slash Command | Description |
|------------|---------------|-------------|
| Paragraph | `/paragraph` | Standard text block |
| Heading 1 | `/heading1`, `/h1` | Large section heading |
| Heading 2 | `/heading2`, `/h2` | Medium section heading |
| Heading 3 | `/heading3`, `/h3` | Small section heading |
| Bullet List | `/bulletlist`, `/bullet` | Unordered list |
| Numbered List | `/numberedlist`, `/numbered` | Ordered list |
| Todo | `/todo`, `/checkbox` | Checkbox item |
| Quote | `/quote`, `/blockquote` | Block quote |
| Code | `/code` | Code block with syntax highlighting |
| Divider | `/divider`, `/hr` | Horizontal divider |
| Image | `/image` | Image embed |
| Callout | `/callout` | Highlighted callout box |
| Table | `/table` | Data table |
| Embed | `/embed` | External content embed |
| Artifact | `/artifact` | Link to AI-generated artifact |

### Slash Commands

Type `/` anywhere in a document to open the command menu:
- Filter commands by typing after the slash
- Navigate with arrow keys
- Press Enter to insert
- Press Escape to cancel

### Document Properties

Add Notion-like properties to any document:

| Property Type | Description |
|---------------|-------------|
| Text | Single-line text input |
| Select | Single selection from options |
| Multi-select | Multiple selections with colored tags |
| Date | Date picker |
| Number | Numeric value |
| Checkbox | Boolean toggle |
| URL | Clickable link |
| Email | Email address |
| Relation | Link to other documents/contexts |

### Property Features

- **Add Property**: Click "+ Add property" to create new properties
- **Property Menu**: Click property name to access delete option
- **Custom Options**: Add options on-the-fly for select/multi-select
- **Color Coding**: Select options automatically get color-coded tags

### Panel Modes

Documents can be viewed in three modes:

1. **Side Panel**: Opens on the right side of the screen
   - Resizable width (400-900px)
   - Doesn't obstruct main content

2. **Center Modal**: Floating centered panel
   - 90% viewport height maximum
   - Click backdrop to close

3. **Full Screen**: Takes over entire viewport
   - Best for focused editing
   - Paper-like background

### Document Features

- **Auto-save**: Changes save automatically after 1.5 seconds of inactivity
- **Word Count**: Live word and block count in status bar
- **Custom Icons**: Choose from 40+ icons for each document
- **Cover Images**: Add cover images to documents
- **Parent Linking**: Nest documents under profiles

---

## Desktop Mode

A multi-window interface for power users who want to work with multiple pages simultaneously.

### Accessing Desktop Mode

Navigate to `/window` or click the Desktop Mode button in settings.

### Window Features

| Feature | Description |
|---------|-------------|
| **Draggable** | Drag windows by their title bar |
| **Resizable** | Resize from edges and corners |
| **Minimize** | Collapse to taskbar |
| **Maximize** | Expand to full viewport |
| **Close** | Remove window |
| **Z-Index** | Click to bring to front |

### Taskbar

- Shows all open windows
- Click to focus/restore minimized windows
- Visual indicator for active window

### Embed Mode

Pages can be opened in embed mode by adding `?embed=true` to any URL:
- Hides main navigation
- Optimized for window display
- Links within maintain embed mode

### Window Types

Open any Business OS page as a window:
- Dashboard
- Chat
- Projects
- Tasks
- Contexts/Documents
- Team
- Settings

---

## Clients System

Manage client relationships and link them across your business.

### Client Profile

| Field | Description |
|-------|-------------|
| Name | Client/company name |
| Email | Primary contact email |
| Phone | Contact phone number |
| Website | Company website |
| Address | Physical address |
| Status | active, inactive, prospect, churned |
| Notes | Free-form notes |

### Client Features

- **Project Linking**: Associate projects with clients
- **Context Profiles**: Link people/business contexts to clients
- **Filtering**: Filter by status (active, all, prospects)
- **Search**: Find clients by name

### Database Schema

```sql
CREATE TABLE clients (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    email VARCHAR,
    phone VARCHAR,
    website VARCHAR,
    address TEXT,
    status VARCHAR DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

---

## AI Chat

The AI chat system provides intelligent assistance with business context awareness.

### Chat Features

| Feature | Description |
|---------|-------------|
| **Streaming** | Real-time response streaming |
| **Project Context** | Select project to focus responses |
| **Context Profiles** | Include business knowledge |
| **Artifacts Panel** | Manage generated content |
| **Conversation History** | Persistent chat history |

### Personalized Greeting

The chat empty state shows a personalized, time-aware greeting:

| Time Range | Greeting |
|------------|----------|
| 12am - 5am | "Up late, {name}?" |
| 5am - 12pm | "Good morning, {name}" |
| 12pm - 5pm | "Good afternoon, {name}" |
| 5pm - 9pm | "Good evening, {name}" |
| 9pm - 12am | "Working late, {name}?" |

### Typewriter Effect

Below the greeting, suggestions rotate with a typewriter animation:
- "streamline your workflow"
- "automate repetitive tasks"
- "create a business proposal"
- "analyze your metrics"
- "draft a client email"
- "plan your week ahead"
- "optimize your processes"

### Context Aggregation

The `/api/contexts/aggregate` endpoint combines content from multiple sources:

```typescript
interface AggregateContextRequest {
  context_ids?: string[];
  project_ids?: string[];
  node_ids?: string[];
  include_children?: boolean;    // Include child documents
  include_artifacts?: boolean;   // Include linked artifacts
  include_tasks?: boolean;       // Include project tasks
  max_depth?: number;            // Max nesting depth (default: 2)
}
```

Returns formatted context string ready for AI consumption.

### Task Generation

Convert AI-generated artifacts into actionable tasks:
1. Click "Generate Tasks" on an artifact
2. AI extracts actionable items
3. Review and edit generated tasks
4. Select project and assign team members
5. Create tasks in one click

---

## Context System

Contexts are the knowledge base that AI uses to understand your business.

### Context Types

| Type | Icon | Description |
|------|------|-------------|
| Person | 👤 | Individual contacts |
| Business | 🏢 | Companies/organizations |
| Project | 📁 | Work initiatives |
| Document | 📄 | Standalone documents |
| Custom | ✨ | Other types |

### Profile Features

- **Custom Icons**: Choose emoji icons for profiles
- **Context Content**: Text information for AI
- **System Prompt**: Custom AI instructions
- **Child Documents**: Attach documents to profiles
- **Client Linking**: Associate with clients

### Document Organization

Documents can be:
- **Standalone**: Not attached to any profile
- **Nested**: Attached to a parent profile
- **Assigned**: Moved between profiles via UI

---

## Projects & Tasks

### Project Features

| Feature | Description |
|---------|-------------|
| Status | active, paused, completed, archived |
| Priority | critical, high, medium, low |
| Deadlines | Start and end dates |
| Client Link | Associate with client |
| Team | Assign team members |
| Notes | Project documentation |

### Task Features

| Feature | Description |
|---------|-------------|
| Status | todo, in_progress, done |
| Priority | low, medium, high |
| Due Date | Task deadline |
| Project Link | Associate with project |
| Assignee | Team member assignment |
| Description | Task details |

### Task-Project Linking

Tasks can be linked to projects:
- View tasks in project context
- Filter tasks by project
- Project completion tracking

---

## Artifacts

AI-generated content with versioning and management.

### Artifact Types

| Type | Description |
|------|-------------|
| Proposal | Business proposals |
| SOP | Standard operating procedures |
| Framework | Business frameworks |
| Report | Analysis reports |
| Plan | Strategic plans |
| Code | Code snippets |
| Markdown | General documents |

### Artifact Features

- **Version History**: Track changes over time
- **Editing**: Modify generated content
- **Saving**: Save to profiles as documents
- **Task Generation**: Convert to actionable tasks
- **Linking**: Associate with projects/contexts

### Artifact Panel

The chat interface includes an artifacts panel:
- Toggle visibility
- Filter by type
- Search artifacts
- Quick preview
- Full editing mode

---

## API Endpoints

### New Endpoints Added

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/contexts/aggregate` | POST | Aggregate context from multiple sources |
| `/api/contexts/{id}/blocks` | PUT | Update document blocks (optimized) |
| `/api/contexts/{id}/share` | POST | Generate share link |
| `/api/artifacts/{id}/link` | PATCH | Link artifact to project/context |
| `/api/artifacts?unassigned_only=true` | GET | Get unassigned artifacts |
| `/api/clients` | GET/POST | Client management |
| `/api/clients/{id}` | GET/PUT/DELETE | Single client operations |

---

## Frontend Components

### New Components

| Component | Path | Description |
|-----------|------|-------------|
| `DocumentProperties.svelte` | `/lib/components/editor/` | Notion-like property editor |
| `Block.svelte` | `/lib/components/editor/` | Document block renderer |
| `BlockMenu.svelte` | `/lib/components/editor/` | Slash command menu |
| `Tooltip.svelte` | `/lib/components/ui/` | Custom tooltip component |

### Stores

| Store | Path | Description |
|-------|------|-------------|
| `editor.ts` | `/lib/stores/` | Document editor state |
| `clients.ts` | `/lib/stores/` | Client management |
| `contexts.ts` | `/lib/stores/` | Context/document state |
| `windowStore.ts` | `/lib/stores/` | Desktop window management |
| `desktopStore.ts` | `/lib/stores/` | Desktop mode state |

---

## Database Migrations

Recent migrations added:

1. **`add_context_document_fields.py`**
   - `blocks` (JSONB) - Document block content
   - `cover_image` (VARCHAR) - Cover image URL
   - `icon` (VARCHAR) - Document icon
   - `parent_id` (UUID) - Parent context reference
   - `is_template` (BOOLEAN) - Template flag
   - `is_archived` (BOOLEAN) - Archive status
   - `last_edited_at` (TIMESTAMP) - Last edit time
   - `word_count` (INTEGER) - Word count cache
   - `is_public` (BOOLEAN) - Public sharing flag
   - `share_id` (VARCHAR) - Unique share identifier
   - `property_schema` (JSONB) - Property definitions
   - `properties` (JSONB) - Property values
   - `client_id` (UUID) - Client association

2. **`add_clients_tables.py`**
   - Full clients table with all fields
   - Foreign key relationships

---

## Configuration

### New Environment Variables

```env
# No new required variables
# All new features use existing configuration
```

### Feature Flags

Currently, all features are enabled by default. Future versions may include:
- `ENABLE_DESKTOP_MODE`
- `ENABLE_DOCUMENT_SHARING`
- `ENABLE_CLIENT_MANAGEMENT`

---

## Best Practices

### Document Organization

1. Create profiles for people/businesses first
2. Attach related documents to profiles
3. Use properties for structured data
4. Use relations to link related documents

### AI Context

1. Keep context profiles updated with relevant info
2. Use project context for focused conversations
3. Aggregate context for comprehensive AI responses

### Desktop Mode

1. Use for multi-tasking workflows
2. Arrange windows for your workflow
3. Minimize inactive windows to taskbar

---

## Troubleshooting

### Common Issues

**Dropdowns not appearing**
- Check for `overflow: hidden` on parent containers
- Ensure z-index is high enough (20+)

**Document not saving**
- Check browser console for API errors
- Verify backend is running
- Check database connection

**AI context not loading**
- Verify context IDs are valid UUIDs
- Check context aggregation endpoint

---

## Future Roadmap

Planned features:
- [ ] Real-time collaboration
- [ ] Document templates marketplace
- [ ] Advanced search across all content
- [ ] Mobile responsive desktop mode
- [ ] Keyboard shortcuts documentation
- [ ] Import/export functionality
- [ ] Webhook integrations
