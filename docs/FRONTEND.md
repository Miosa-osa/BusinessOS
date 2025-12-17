# Frontend Architecture

**Business OS Frontend — SvelteKit 2.0 + Svelte 5**

---

## Overview

The Business OS frontend is a modern, reactive web application built with SvelteKit 2.0 and Svelte 5 Runes. It features a unique dual-mode interface: a traditional web app layout and a macOS-inspired desktop environment with window management.

---

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| **SvelteKit** | 2.0 | Meta-framework for routing, SSR, and build |
| **Svelte** | 5.x | UI framework with Runes (`$state`, `$derived`, `$effect`) |
| **TypeScript** | 5.x | Type safety throughout the codebase |
| **TailwindCSS** | 4.x | Utility-first CSS framework |
| **bits-ui** | Latest | Headless UI components (Radix-like) |
| **Motion** | Latest | Animation library |
| **Better Auth** | Latest | Authentication (cookie-based) |

---

## Directory Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── api/
│   │   │   └── client.ts          # API client (100+ methods, 1600+ lines)
│   │   ├── components/            # 75+ Svelte components
│   │   │   ├── ai-elements/       # AI chat UI components
│   │   │   ├── auth/              # Auth forms
│   │   │   ├── calendar/          # Calendar widgets
│   │   │   ├── chat/              # Chat & Focus Modes
│   │   │   ├── clients/           # CRM views
│   │   │   ├── dashboard/         # Dashboard widgets
│   │   │   ├── desktop/           # Desktop mode (Window, Dock, etc.)
│   │   │   ├── editor/            # Block-based document editor
│   │   │   ├── onboarding/        # Onboarding flows
│   │   │   ├── tasks/             # Task management views
│   │   │   ├── team/              # Team management
│   │   │   └── ui/                # Shared UI primitives
│   │   ├── stores/                # State management (10 stores)
│   │   │   ├── windowStore.ts     # Desktop window management
│   │   │   ├── desktopStore.ts    # Desktop customization
│   │   │   ├── chat.ts            # Chat state
│   │   │   ├── auth.ts            # Auth state
│   │   │   ├── projects.ts        # Projects state
│   │   │   ├── clients.ts         # CRM state
│   │   │   ├── contexts.ts        # Documents state
│   │   │   ├── team.ts            # Team state
│   │   │   ├── editor.ts          # Editor state
│   │   │   └── themeStore.ts      # Theme preferences
│   │   └── utils/                 # Utility functions
│   ├── routes/
│   │   ├── (app)/                 # Protected app routes
│   │   │   ├── +layout.svelte     # App shell with sidebar
│   │   │   ├── dashboard/         # Main dashboard
│   │   │   ├── chat/              # AI Chat & Focus Modes
│   │   │   ├── projects/          # Project management
│   │   │   ├── tasks/             # Task boards
│   │   │   ├── team/              # Team directory
│   │   │   ├── clients/           # CRM
│   │   │   ├── contexts/          # Documents
│   │   │   ├── nodes/             # Business structure
│   │   │   ├── calendar/          # Calendar
│   │   │   ├── daily/             # Daily log
│   │   │   ├── settings/          # User settings
│   │   │   ├── profile/           # User profile
│   │   │   ├── usage/             # Usage analytics
│   │   │   ├── voice-notes/       # Voice recordings
│   │   │   └── help/              # Help & docs
│   │   ├── (embed)/               # Embeddable routes
│   │   ├── docs/                  # Public documentation
│   │   ├── window/                # Desktop mode entry
│   │   ├── popup-chat/            # Embeddable chat widget
│   │   ├── login/                 # Login page
│   │   ├── register/              # Registration
│   │   ├── forgot-password/       # Password reset
│   │   ├── reset-password/        # Password reset form
│   │   └── onboarding/            # User onboarding
│   ├── app.css                    # Global styles + Tailwind
│   ├── app.html                   # HTML template
│   └── hooks.server.ts            # Server hooks (auth)
├── static/                        # Static assets
├── package.json
├── svelte.config.js
├── tailwind.config.js
└── vite.config.ts
```

---

## State Management

### Store Architecture

Business OS uses Svelte stores with a modular pattern. Each domain has its own store with async actions.

### Core Stores

#### 1. Window Store (`windowStore.ts`)

Manages the desktop environment's window state.

```typescript
interface WindowState {
  id: string;
  module: string;           // 'chat', 'dashboard', etc.
  title: string;
  x: number; y: number;     // Position
  width: number; height: number;
  minimized: boolean;
  maximized: boolean;
  snapped?: SnapZone;       // Split-screen zones
  data?: Record<string, unknown>;  // Custom window data
}

interface WindowStore {
  windows: WindowState[];
  focusedWindowId: string | null;
  windowOrder: string[];    // Z-index order
  dockPinnedItems: string[];
  desktopIcons: DesktopIcon[];
  selectedIconIds: string[];
  folders: DesktopFolder[];
}
```

**Key Methods:**
- `openWindow(module, options?)` — Open or focus a window
- `closeWindow(windowId)` — Close a window
- `minimizeWindow(windowId)` — Minimize to dock
- `toggleMaximize(windowId)` — Maximize/restore
- `snapWindow(windowId, zone, width, height)` — Snap to screen zone
- `focusWindow(windowId)` — Bring to front
- `createFolder(name, x, y, color)` — Create desktop folder
- `exportConfig()` / `importConfig(config)` — Save/load layout

**Snap Zones:**
- `left`, `right` — Half screen
- `top-left`, `top-right`, `bottom-left`, `bottom-right` — Quarter screen

#### 2. Desktop Store (`desktopStore.ts`)

Manages desktop appearance and customization.

```typescript
interface DesktopSettings {
  backgroundId: string;           // Preset background ID
  customBackgroundUrl: string | null;
  backgroundFit: BackgroundFit;   // 'cover' | 'contain' | 'fill' | 'center'
  showNoise: boolean;             // Film grain overlay
  iconStyle: IconStyle;           // 15 icon styles
  iconSize: number;               // 32-128px
  showIconLabels: boolean;
  gridSnap: boolean;
  companyName: string;            // Loading screen text
}
```

**50+ Background Presets:**
- Solid colors (18): Classic Gray, Warm Beige, Dark Mode, etc.
- Gradients (16): Sunrise, Ocean, Aurora, Cosmic, etc.
- Patterns (17): Dots, Grid, Blueprint, Carbon Fiber, etc.

**15 Icon Styles:**
- `default`, `minimal`, `rounded`, `square`, `macos`
- `macos-classic`, `outlined`, `retro`, `win95`, `glassmorphism`
- `neon`, `flat`, `gradient`, `paper`, `pixel`

#### 3. Chat Store (`chat.ts`)

Manages conversations and message streaming.

```typescript
interface ChatState {
  conversations: Conversation[];
  currentConversation: Conversation | null;
  messages: Message[];
  loading: boolean;
  streaming: boolean;
  streamingContent: string;
}
```

**Key Methods:**
- `loadConversations()` — Fetch all conversations
- `loadConversation(id)` — Load specific conversation
- `sendMessage(content, contextId?, model?)` — Send with streaming
- `newConversation()` — Start fresh
- `deleteConversation(id)` — Remove conversation
- `search(query)` — Search messages

#### 4. Other Stores

| Store | Purpose |
|-------|---------|
| `auth.ts` | User session, login state |
| `projects.ts` | Project CRUD, filtering |
| `clients.ts` | CRM state |
| `contexts.ts` | Document management |
| `team.ts` | Team members |
| `editor.ts` | Block editor state |
| `themeStore.ts` | Light/dark mode |

---

## API Client

The API client (`src/lib/api/client.ts`) is a typed HTTP client with 100+ methods.

### Configuration

```typescript
// Auto-switches between dev and prod
const API_BASE = import.meta.env.DEV
  ? 'http://localhost:8000/api'
  : '/api';
```

### Request Pattern

```typescript
class ApiClient {
  private async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      method,
      headers,
      credentials: 'include',  // Send auth cookies
      body: body ? JSON.stringify(body) : undefined
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ detail: 'Request failed' }));
      throw new Error(error.detail || 'Request failed');
    }

    return response.json();
  }
}
```

### API Domains

| Domain | Methods | Examples |
|--------|---------|----------|
| **Chat** | 10+ | `getConversations`, `sendMessage`, `searchConversations` |
| **Projects** | 6 | `getProjects`, `createProject`, `addProjectNote` |
| **Contexts** | 12 | `getContexts`, `updateContextBlocks`, `enableContextSharing` |
| **Tasks** | 6 | `getTasks`, `createTask`, `toggleTask` |
| **Team** | 7 | `getTeamMembers`, `updateTeamMemberCapacity` |
| **Clients** | 15+ | `getClients`, `createContact`, `updateDealStage` |
| **Nodes** | 10 | `getNodeTree`, `activateNode`, `reorderNode` |
| **Calendar** | 8 | `getCalendarEvents`, `syncCalendar`, `getTodayEvents` |
| **AI Config** | 10 | `getAIProviders`, `getAllModels`, `pullModel` |
| **Usage** | 7 | `getUsageSummary`, `getUsageByProvider`, `getMCPUsage` |
| **Voice** | 5 | `uploadVoiceNote`, `retranscribeVoiceNote` |

### Type Definitions

The client exports 100+ TypeScript interfaces for API responses:

```typescript
// Examples
export interface Conversation { ... }
export interface Project { ... }
export interface Context { ... }
export interface CalendarEvent { ... }
export interface DealResponse { ... }
```

---

## Component Architecture

### Component Categories

#### AI Elements (`components/ai-elements/`)

| Component | Purpose |
|-----------|---------|
| `Message.svelte` | Chat message bubble |
| `PromptInput.svelte` | Text input with send |
| `Loader.svelte` | Typing indicator |
| `CodeBlock.svelte` | Syntax-highlighted code |
| `Artifact.svelte` | Generated document card |
| `MessageActions.svelte` | Copy, regenerate actions |
| `Conversation.svelte` | Full conversation view |
| `Suggestion.svelte` | AI suggestion chip |

#### Chat Components (`components/chat/`)

| Component | Purpose |
|-----------|---------|
| `ChatInput.svelte` | Main chat input with voice |
| `EmptyState.svelte` | Welcome screen |
| `ConversationListPanel.svelte` | Sidebar with filters |
| `ConversationListItem.svelte` | Conversation row |
| `FocusModeSelector.svelte` | Focus mode cards |
| `FocusCard.svelte` | Individual focus card |
| `ProgressPanel.svelte` | Right panel progress |
| `ContextPanel.svelte` | Context/document panel |
| `UserMessage.svelte` | User message display |
| `AssistantMessage.svelte` | AI message display |
| `TypingIndicator.svelte` | Loading dots |

#### Desktop Components (`components/desktop/`)

| Component | Purpose |
|-----------|---------|
| `Window.svelte` | Draggable, resizable window |
| `Dock.svelte` | Bottom app dock |
| `MenuBar.svelte` | Top menu bar |
| `DesktopIcon.svelte` | Desktop icon (15 styles) |
| `SpotlightSearch.svelte` | ⌘+Space search |
| `FileBrowser.svelte` | Finder-like file browser |
| `FolderWindow.svelte` | Folder contents view |
| `DesktopSettingsContent.svelte` | Desktop preferences |
| `Terminal.svelte` | Terminal emulator |

#### Editor Components (`components/editor/`)

| Component | Purpose |
|-----------|---------|
| `DocumentEditor.svelte` | Main block editor |
| `Block.svelte` | Single content block |
| `BlockMenu.svelte` | Slash command menu |
| `DocumentProperties.svelte` | Notion-like properties |

#### Dashboard Components (`components/dashboard/`)

| Component | Purpose |
|-----------|---------|
| `DashboardHeader.svelte` | Header with actions |
| `TodaysFocusWidget.svelte` | Focus items widget |
| `MyTasksWidget.svelte` | Tasks widget |
| `ActiveProjectsWidget.svelte` | Projects overview |
| `RecentActivityWidget.svelte` | Activity feed |
| `QuickActionsWidget.svelte` | Quick action buttons |

---

## Desktop Mode

### Architecture

Desktop mode transforms the app into a macOS-like windowing environment.

```
┌─────────────────────────────────────────────────────────────────┐
│ MenuBar.svelte                                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌─────────────────────────────────────────┐                   │
│   │ Window.svelte (chat)                    │                   │
│   │                                         │                   │
│   │   [Route content rendered inside]       │                   │
│   │                                         │                   │
│   └─────────────────────────────────────────┘                   │
│                                                                 │
│   DesktopIcon.svelte  DesktopIcon.svelte                        │
│   [Dashboard]         [Projects]                                │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ Dock.svelte                                                     │
└─────────────────────────────────────────────────────────────────┘
```

### Window.svelte Features

```svelte
<script lang="ts">
  // Window state from windowStore
  let { window: windowState } = $props();

  // Dragging
  let isDragging = $state(false);
  let dragOffset = $state({ x: 0, y: 0 });

  // Resizing from any edge/corner
  type ResizeEdge = 'n' | 's' | 'e' | 'w' | 'ne' | 'nw' | 'se' | 'sw';
  let resizeEdge = $state<ResizeEdge | null>(null);

  // Snap preview zones
  let snapPreview = $state<SnapZone>(null);
</script>
```

**Window Controls:**
- Traffic light buttons (close, minimize, maximize)
- Double-click title bar to maximize
- Drag to edges for snap zones
- Resize from any edge or corner

### Dock.svelte Features

```svelte
// Dock state
let dockItems = $derived($windowStore.dockPinnedItems);
let openWindows = $derived($windowStore.windows);

// Collapsed voice chat bubble
let isCollapsed = $state(true);
let isRecording = $state(false);
```

**Features:**
- Pinned app icons with indicators
- Running app dots
- Minimized window restoration
- Voice recording bubble (when collapsed)
- Tooltip previews

### SpotlightSearch.svelte

Triggered by `⌘+Space`:

```svelte
// Search across all modules
let results = $derived(
  allItems.filter(item =>
    item.name.toLowerCase().includes(query.toLowerCase())
  )
);
```

**Searches:**
- Apps/Modules
- Projects
- Contexts/Documents
- Clients
- Tasks

---

## AI Focus Modes

### Focus Mode Architecture

Focus modes provide specialized AI interactions.

```typescript
interface FocusMode {
  id: string;
  name: string;
  icon: string;
  description: string;
  agent: string;        // Backend agent to use
  options: FocusOption[];
}

interface FocusOption {
  id: string;
  label: string;
  values: string[];
  default: string;
}
```

### Available Modes

| Mode | Agent | Options |
|------|-------|---------|
| **Research** | analysis_agent | Scope (Web/Docs/All), Depth, Output |
| **Analyze** | analysis_agent | Approach, Depth, Output |
| **Write** | document_agent | Format, Mode |
| **Build** | planning_agent | Create, Detail |
| **Do More** | orchestrator | Mode |

### FocusModeSelector.svelte

```svelte
<script lang="ts">
  let selectedMode = $state<FocusMode | null>(null);
  let selectedOptions = $state<Record<string, string>>({});

  function handleModeSelect(mode: FocusMode) {
    selectedMode = mode;
    // Initialize default options
    selectedOptions = Object.fromEntries(
      mode.options.map(opt => [opt.id, opt.default])
    );
  }
</script>
```

### Conversation Filtering

The ConversationListPanel supports filtering by mode:

```svelte
let modeFilter: 'all' | 'focus' | 'chat' = $state('all');
let projectFilter: string = $state('all');

const filteredConversations = $derived(() => {
  let filtered = conversations;

  if (modeFilter === 'focus') {
    filtered = filtered.filter(c => c.conversationType === 'focus');
  }

  if (projectFilter !== 'all') {
    filtered = filtered.filter(c => c.projectId === projectFilter);
  }

  return filtered;
});
```

---

## Block-Based Editor

### Block Types

The document editor supports Notion-like blocks:

| Block Type | Purpose |
|------------|---------|
| `paragraph` | Regular text |
| `heading-1/2/3` | H1, H2, H3 headers |
| `bulleted-list` | Bullet list |
| `numbered-list` | Numbered list |
| `todo` | Checkbox item |
| `quote` | Block quote |
| `code` | Code block |
| `divider` | Horizontal rule |
| `image` | Image embed |
| `callout` | Colored callout box |
| `table` | Table block |
| `embed` | URL embed |
| `artifact` | AI artifact embed |

### Block Structure

```typescript
interface Block {
  id: string;
  type: string;
  content: string | null;
  properties?: {
    language?: string;    // For code blocks
    checked?: boolean;    // For todos
    color?: string;       // For callouts
    url?: string;         // For images/embeds
    // ...
  };
  children?: Block[];     // Nested blocks
}
```

### Slash Commands

Type `/` to trigger the block menu:

```svelte
// BlockMenu.svelte
const commands = [
  { type: 'paragraph', label: 'Text', icon: '📝' },
  { type: 'heading-1', label: 'Heading 1', icon: 'H1' },
  { type: 'bulleted-list', label: 'Bullet List', icon: '•' },
  { type: 'todo', label: 'To-do', icon: '☐' },
  { type: 'code', label: 'Code', icon: '</>' },
  // ...
];
```

### Document Properties

Notion-like custom properties:

```typescript
interface PropertySchema {
  name: string;
  type: 'text' | 'select' | 'multi_select' | 'date' |
        'person' | 'relation' | 'number' | 'checkbox' | 'url' | 'email';
  options?: string[];           // For select types
  relation_type?: 'context' | 'project' | 'client';
}
```

---

## Authentication

### Better Auth Integration

```typescript
// hooks.server.ts
import { auth } from '$lib/auth';

export const handle: Handle = async ({ event, resolve }) => {
  const session = await auth.api.getSession({
    headers: event.request.headers
  });

  event.locals.session = session?.session ?? null;
  event.locals.user = session?.user ?? null;

  return resolve(event);
};
```

### Protected Routes

```svelte
<!-- (app)/+layout.svelte -->
<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';

  $effect(() => {
    if (!$page.data.session) {
      goto('/login');
    }
  });
</script>
```

---

## Routing Structure

### Route Groups

| Group | Purpose | Auth Required |
|-------|---------|---------------|
| `(app)/` | Main application | Yes |
| `(embed)/` | Embeddable widgets | Configurable |
| Root routes | Public pages | No |

### Dynamic Routes

```
/projects/[id]     → Project detail
/clients/[id]      → Client detail
/contexts/[id]     → Document editor
/nodes/[id]        → Node detail
/docs/[slug]       → Documentation page
```

### Special Routes

| Route | Purpose |
|-------|---------|
| `/window` | Desktop mode entry point |
| `/popup-chat` | Embeddable chat widget |
| `/docs` | Public documentation |

---

## Styling

### TailwindCSS 4.x

```css
/* app.css */
@import 'tailwindcss';

/* Custom CSS variables */
:root {
  --color-background: #ffffff;
  --color-foreground: #1a1a1a;
  --color-primary: #3b82f6;
  --color-border: #e5e7eb;
  /* ... */
}

.dark {
  --color-background: #1a1a1a;
  --color-foreground: #f5f5f7;
  /* ... */
}
```

### Component Patterns

```svelte
<!-- Utility classes with conditional logic -->
<button
  class="px-4 py-2 rounded-lg transition-colors
         {active ? 'bg-primary text-white' : 'bg-gray-100 hover:bg-gray-200'}"
>
  {label}
</button>
```

---

## Performance Optimizations

### Lazy Loading

```svelte
<!-- Heavy components loaded on demand -->
{#await import('./HeavyComponent.svelte') then { default: HeavyComponent }}
  <HeavyComponent />
{/await}
```

### Derived State

```typescript
// Computed values that update automatically
const filteredItems = $derived(() =>
  items.filter(item => item.status === selectedStatus)
);
```

### Debounced Updates

```typescript
// Auto-save with debouncing
import { debounce } from '$lib/utils/debounce';

const saveBlocks = debounce(async (blocks: Block[]) => {
  await api.updateContextBlocks(contextId, { blocks });
}, 1000);
```

---

## Key Patterns

### 1. Store + API Pattern

```typescript
// Store method calls API, updates state
async loadItems() {
  update(s => ({ ...s, loading: true }));
  try {
    const items = await api.getItems();
    update(s => ({ ...s, items, loading: false }));
  } catch (error) {
    update(s => ({ ...s, loading: false, error: error.message }));
  }
}
```

### 2. Optimistic Updates

```typescript
// Update UI immediately, sync with server
update(s => ({ ...s, items: [...s.items, newItem] }));
try {
  await api.createItem(newItem);
} catch {
  // Rollback on failure
  update(s => ({ ...s, items: s.items.filter(i => i.id !== newItem.id) }));
}
```

### 3. Streaming Responses

```typescript
// Read SSE stream for AI responses
const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;

  const chunk = decoder.decode(value, { stream: true });
  streamingContent += chunk;
}
```

### 4. Props with Svelte 5 Runes

```svelte
<script lang="ts">
  interface Props {
    title: string;
    onSave?: (data: Data) => void;
    children?: Snippet;
  }

  let { title, onSave, children }: Props = $props();
</script>
```

---

## Development

### Running Locally

```bash
cd frontend
npm install
npm run dev
```

### Building

```bash
npm run build
npm run preview  # Preview production build
```

### Type Checking

```bash
npm run check
```

---

## File Reference

| File | Lines | Purpose |
|------|-------|---------|
| `api/client.ts` | 1687 | API client + types |
| `stores/windowStore.ts` | 958 | Desktop window management |
| `stores/desktopStore.ts` | 789 | Desktop customization |
| `routes/(app)/chat/+page.svelte` | 4500+ | Main chat page |
| `components/desktop/Dock.svelte` | 800+ | Dock component |
| `components/desktop/Window.svelte` | 600+ | Window component |
| `components/editor/DocumentEditor.svelte` | 1000+ | Block editor |

---

*This documentation reflects the frontend architecture as of December 2025.*
