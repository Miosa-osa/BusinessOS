# App Store Integration System

## Overview

The **App Store Integration System** allows BusinessOS users to discover, install, and manage external web applications within their 3D Desktop environment. Users can add popular business tools (Notion, Slack, Linear, etc.) as integrated apps that open in iframe windows with full desktop integration.

### Key Features

- Browse 100+ pre-configured popular business apps
- Add custom web apps with auto-fetched logos
- Persistent app storage per workspace
- Desktop icon integration with custom logos
- Dock integration for quick access
- Auto-open on startup support
- Workspace-specific app management

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        USER INTERACTION                          │
│  User opens App Store → Browses/Searches → Installs app         │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    FRONTEND (SvelteKit)                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ App Store Page                                            │  │
│  │ routes/(app)/app-store/+page.svelte                      │  │
│  │ - Browse popular apps (100+ pre-configured)              │  │
│  │ - Search and filter by category                          │  │
│  │ - Add custom apps via URL                                │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │                                          │
│  ┌────────────────────▼─────────────────────────────────────┐  │
│  │ AppRegistryModal Component                                │  │
│  │ lib/components/desktop/AppRegistryModal.svelte           │  │
│  │ - Displays app catalog (featured, categories)            │  │
│  │ - Install/uninstall UI                                   │  │
│  │ - Custom app form (name, URL, category)                  │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │                                          │
│  ┌────────────────────▼─────────────────────────────────────┐  │
│  │ userAppsStore (State Management)                         │  │
│  │ lib/stores/userAppsStore.ts                              │  │
│  │ - fetch(workspaceId) - Load apps                         │  │
│  │ - create(params) - Install app                           │  │
│  │ - update(id, params) - Update settings                   │  │
│  │ - delete(id) - Uninstall app                             │  │
│  │ - recordOpened(id) - Track usage                         │  │
│  │ - fetchStartupApps() - Get auto-launch apps              │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │                                          │
│                       │ Integrates with:                         │
│  ┌────────────────────▼─────────────────────────────────────┐  │
│  │ windowStore (Desktop Integration)                        │  │
│  │ lib/stores/windowStore.ts                                │  │
│  │ - registerUserApp(app) - Add desktop icon               │  │
│  │ - unregisterUserApp(id) - Remove icon                    │  │
│  │ - Desktop icons use app logos (x: -3 column)             │  │
│  │ - Module ID: "user-app-{app.id}"                         │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────────┘
                         │ HTTP Requests
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      BACKEND (Go)                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ UserAppsHandler                                           │  │
│  │ internal/handlers/user_apps.go                           │  │
│  │                                                           │  │
│  │ Endpoints:                                                │  │
│  │ - GET  /api/user-apps              → ListUserApps       │  │
│  │ - GET  /api/user-apps/:id          → GetUserApp         │  │
│  │ - POST /api/user-apps              → CreateUserApp      │  │
│  │ - PUT  /api/user-apps/:id          → UpdateUserApp      │  │
│  │ - DELETE /api/user-apps/:id        → DeleteUserApp      │  │
│  │ - PUT  /api/user-apps/:id/position → UpdateAppPosition  │  │
│  │ - POST /api/user-apps/:id/open     → RecordAppOpened    │  │
│  │ - GET  /api/user-apps/startup      → GetStartupApps     │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │                                          │
│  ┌────────────────────▼─────────────────────────────────────┐  │
│  │ FaviconFetcher Utility                                    │  │
│  │ internal/utils/favicon.go                                │  │
│  │ - Fetches app logos via Google Favicon API               │  │
│  │ - Auto-populates logo_url when creating apps             │  │
│  └────────────────────┬─────────────────────────────────────┘  │
│                       │                                          │
│  ┌────────────────────▼─────────────────────────────────────┐  │
│  │ SQLC Queries                                              │  │
│  │ internal/database/queries/user_external_apps.sql         │  │
│  │ - ListUserExternalApps (active only)                     │  │
│  │ - ListAllUserExternalApps (including inactive)           │  │
│  │ - GetUserExternalApp                                     │  │
│  │ - CreateUserExternalApp                                  │  │
│  │ - UpdateUserExternalApp                                  │  │
│  │ - DeleteUserExternalApp                                  │  │
│  │ - RecordAppOpened (updates last_opened_at)               │  │
│  │ - UpdateAppPosition (3D desktop position)                │  │
│  │ - ToggleAppActive                                        │  │
│  │ - GetStartupApps                                         │  │
│  └────────────────────┬─────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                     DATABASE (PostgreSQL)                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ user_external_apps table                                  │  │
│  │ Migration: 047_user_external_apps.sql                    │  │
│  │                                                           │  │
│  │ Columns:                                                  │  │
│  │ - id (UUID, primary key)                                 │  │
│  │ - user_id (FK → user.id)                                 │  │
│  │ - workspace_id (FK → workspaces.id)                      │  │
│  │ - name (VARCHAR, required)                               │  │
│  │ - url (TEXT, required)                                   │  │
│  │ - icon (VARCHAR, Lucide icon name)                       │  │
│  │ - color (VARCHAR, hex color)                             │  │
│  │ - logo_url (TEXT, app logo/favicon URL)                  │  │
│  │ - category (VARCHAR, e.g., "productivity")               │  │
│  │ - description (TEXT, optional)                           │  │
│  │ - position_x, position_y, position_z (INTEGER)           │  │
│  │ - iframe_config (JSONB, sandbox settings)                │  │
│  │ - is_active (BOOLEAN, soft delete)                       │  │
│  │ - open_on_startup (BOOLEAN)                              │  │
│  │ - app_type (VARCHAR, 'web' or 'native')                  │  │
│  │ - created_at, updated_at, last_opened_at                 │  │
│  │                                                           │  │
│  │ Indexes:                                                  │  │
│  │ - idx_user_external_apps_user                            │  │
│  │ - idx_user_external_apps_workspace                       │  │
│  │ - idx_user_external_apps_active (partial index)          │  │
│  │                                                           │  │
│  │ Constraints:                                              │  │
│  │ - user_external_apps_name_workspace_unique               │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Database Schema

### Table: `user_external_apps`

Created by migration `047_user_external_apps.sql`.

```sql
CREATE TABLE user_external_apps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Ownership
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

    -- App Identity
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,

    -- Visual Metadata
    icon VARCHAR(100) NOT NULL DEFAULT 'app-window',  -- Lucide icon name
    color VARCHAR(7) NOT NULL DEFAULT '#6366F1',      -- Hex color
    logo_url TEXT,                                    -- Actual app logo (auto-fetched)

    -- Categorization
    category VARCHAR(100) DEFAULT 'productivity',
    description TEXT,

    -- Desktop Positioning
    position_x INTEGER DEFAULT 0,
    position_y INTEGER DEFAULT 0,
    position_z INTEGER DEFAULT 0,

    -- Iframe Configuration
    iframe_config JSONB DEFAULT '{"sandbox": [...], "allowFullscreen": true}'::jsonb,

    -- State Management
    is_active BOOLEAN DEFAULT true,
    open_on_startup BOOLEAN DEFAULT false,

    -- App Type
    app_type VARCHAR(50) DEFAULT 'web' NOT NULL,  -- 'web' or 'native'

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_opened_at TIMESTAMPTZ
);
```

#### Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique app identifier |
| `user_id` | VARCHAR(255) | Owner of the app (FK to `user` table) |
| `workspace_id` | UUID | Workspace this app belongs to (FK to `workspaces` table) |
| `name` | VARCHAR(255) | Display name (e.g., "Notion", "Slack") |
| `url` | TEXT | Web app URL (e.g., "https://notion.so") |
| `icon` | VARCHAR(100) | **Deprecated** - Lucide icon fallback (use `logo_url` instead) |
| `color` | VARCHAR(7) | Hex color for branding (e.g., "#000000" for Notion) |
| `logo_url` | TEXT | **Auto-fetched** actual app logo/favicon URL |
| `category` | VARCHAR(100) | App category ("productivity", "communication", "design", etc.) |
| `description` | TEXT | Optional app description |
| `position_x/y/z` | INTEGER | 3D desktop position (persisted across sessions) |
| `iframe_config` | JSONB | Iframe sandbox settings and permissions |
| `is_active` | BOOLEAN | Soft delete flag (can be disabled without deletion) |
| `open_on_startup` | BOOLEAN | Auto-launch when desktop loads |
| `app_type` | VARCHAR(50) | "web" (iframe) or "native" (future support) |
| `last_opened_at` | TIMESTAMPTZ | Tracks usage for analytics |

#### Indexes

- `idx_user_external_apps_user` - User lookup
- `idx_user_external_apps_workspace` - Workspace lookup
- `idx_user_external_apps_active` - Partial index for active apps only

#### Constraints

- `user_external_apps_name_workspace_unique` - One app name per workspace (prevents duplicates)

---

## Backend Implementation

### Handlers (API Endpoints)

File: `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/handlers/user_apps.go`

#### Endpoints

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/api/user-apps?workspace_id={id}` | `ListUserApps` | Fetch all active apps for a workspace |
| GET | `/api/user-apps?workspace_id={id}&include_inactive=true` | `ListUserApps` | Fetch all apps (including inactive) |
| GET | `/api/user-apps/:id?workspace_id={id}` | `GetUserApp` | Get specific app by ID |
| POST | `/api/user-apps` | `CreateUserApp` | Install new app |
| PUT | `/api/user-apps/:id?workspace_id={id}` | `UpdateUserApp` | Update app settings |
| DELETE | `/api/user-apps/:id?workspace_id={id}` | `DeleteUserApp` | Uninstall app |
| PUT | `/api/user-apps/:id/position` | `UpdateAppPosition` | Update 3D desktop position |
| POST | `/api/user-apps/:id/open` | `RecordAppOpened` | Track app usage |
| GET | `/api/user-apps/startup?workspace_id={id}` | `GetStartupApps` | Get apps configured for auto-launch |

#### Request/Response Types

**CreateUserAppRequest:**
```go
type CreateUserAppRequest struct {
    WorkspaceID   string                 `json:"workspace_id" binding:"required"`
    Name          string                 `json:"name" binding:"required"`
    URL           string                 `json:"url" binding:"required"`
    Icon          string                 `json:"icon"`           // Optional (deprecated)
    Color         string                 `json:"color"`          // Optional, defaults to "#6366F1"
    LogoURL       string                 `json:"logo_url"`       // Optional, auto-fetched if empty
    Category      string                 `json:"category"`       // Optional, defaults to "productivity"
    Description   string                 `json:"description"`    // Optional
    IframeConfig  map[string]interface{} `json:"iframe_config"`  // Optional
    OpenOnStartup bool                   `json:"open_on_startup"` // Default: false
    AppType       string                 `json:"app_type"`       // "web" or "native"
}
```

**UpdateUserAppRequest:**
```go
type UpdateUserAppRequest struct {
    Name          *string                `json:"name"`
    URL           *string                `json:"url"`
    Icon          *string                `json:"icon"`
    Color         *string                `json:"color"`
    LogoURL       *string                `json:"logo_url"`
    Category      *string                `json:"category"`
    Description   *string                `json:"description"`
    PositionX     *int32                 `json:"position_x"`
    PositionY     *int32                 `json:"position_y"`
    PositionZ     *int32                 `json:"position_z"`
    IframeConfig  map[string]interface{} `json:"iframe_config"`
    IsActive      *bool                  `json:"is_active"`
    OpenOnStartup *bool                  `json:"open_on_startup"`
}
```

#### Logo Auto-Fetching

When creating an app without a `logo_url`, the backend automatically fetches the app's favicon:

```go
// internal/handlers/user_apps.go:234-246
logoURL := req.LogoURL
if logoURL == "" && req.AppType != "native" {
    // For web apps, fetch favicon from URL using Google's Favicon API
    fetchedLogoURL, err := h.faviconFetcher.FetchFaviconURL(req.URL)
    if err != nil {
        logging.Warn("[UserApps] Failed to fetch favicon for %s: %v", req.URL, err)
        // Continue without logo - not critical
    } else {
        logoURL = fetchedLogoURL
    }
}
```

This uses the `FaviconFetcher` utility (`internal/utils/favicon.go`) which calls Google's Favicon API to retrieve high-quality app logos.

---

## Frontend Implementation

### App Store Page

File: `/Users/rhl/Desktop/BusinessOS2/frontend/src/routes/(app)/app-store/+page.svelte`

Simple wrapper that renders the `AppRegistryModal` component as a full page:

```svelte
<script lang="ts">
    import { currentWorkspaceId } from '$lib/stores/workspaces';
    import AppRegistryModal from '$lib/components/desktop/AppRegistryModal.svelte';

    let workspaceId = $derived($currentWorkspaceId || '');
</script>

<div class="app-store-page">
    {#if workspaceId}
        <AppRegistryModal {workspaceId} isPage={true} />
    {:else}
        <div class="loading-state">Loading workspace...</div>
    {/if}
</div>
```

### AppRegistryModal Component

File: `/Users/rhl/Desktop/BusinessOS2/frontend/src/lib/components/desktop/AppRegistryModal.svelte`

The main component that powers the app store UI.

#### Features:

1. **Browse Tab** - 100+ pre-configured popular business apps
   - Featured apps (HubSpot, Salesforce, Slack, etc.)
   - Categorized apps (AI Tools, Business & CRM, Communication, etc.)
   - Search and filter functionality
   - Install with one click

2. **My Apps Tab** - User's installed apps
   - List of all installed apps
   - Toggle auto-start
   - Uninstall apps
   - Open app windows

3. **Custom Tab** - Add any web app
   - Name, URL, category inputs
   - Auto-fetch app logo from URL
   - Color picker for branding

#### Categories:

```typescript
const categories = [
    { id: 'all', name: 'All Apps', icon: Grid3X3, color: '#10B981' },
    { id: 'ai', name: 'AI Tools', icon: Brain, color: '#8B5CF6' },
    { id: 'business', name: 'Business & CRM', icon: Briefcase, color: '#F97316' },
    { id: 'productivity', name: 'Productivity', icon: TrendingUp, color: '#10B981' },
    { id: 'communication', name: 'Communication', icon: MessageSquare, color: '#3B82F6' },
    { id: 'project-management', name: 'Project Management', icon: Briefcase, color: '#F59E0B' },
    { id: 'design', name: 'Design', icon: Palette, color: '#EC4899' },
    { id: 'storage', name: 'Storage', icon: FolderOpen, color: '#14B8A6' },
    { id: 'media', name: 'Media', icon: Music, color: '#EF4444' },
    { id: 'social', name: 'Social', icon: Users, color: '#8B5CF6' }
];
```

#### Pre-configured Apps (Examples):

**Business & CRM:**
- HubSpot, Salesforce, Pipedrive, Zoho CRM, Freshsales, Close, Copper, Streak
- QuickBooks, Xero, FreshBooks, Wave, Stripe

**Communication:**
- Slack, Microsoft Teams, Discord, Zoom, Google Meet

**Project Management:**
- Asana, Trello, Monday.com, ClickUp, Jira, Linear, Basecamp

**AI Tools:**
- ChatGPT, Claude, Perplexity, Gemini, Jasper

**Design:**
- Figma, Canva, Adobe XD, Sketch

(See component for full list of 100+ apps)

### State Management: userAppsStore

File: `/Users/rhl/Desktop/BusinessOS2/frontend/src/lib/stores/userAppsStore.ts`

Svelte store that manages app state and API communication.

#### Interface:

```typescript
export interface UserApp {
    id: string;
    user_id: string;
    workspace_id: string;
    name: string;
    url: string;
    icon: string;              // Lucide icon name (deprecated)
    color: string;             // Hex color
    logo_url?: string | null;  // Actual app logo
    category?: string | null;
    description?: string | null;
    position_x?: number | null;
    position_y?: number | null;
    position_z?: number | null;
    iframe_config: Record<string, any>;
    is_active?: boolean | null;
    open_on_startup?: boolean | null;
    app_type: string;  // 'web' or 'native'
    created_at: string;
    updated_at: string;
    last_opened_at?: string | null;
}
```

#### Methods:

```typescript
// Fetch all apps for workspace
await userAppsStore.fetch(workspaceId: string, includeInactive?: boolean)

// Get specific app
const app = await userAppsStore.get(appId: string, workspaceId: string)

// Install new app
const newApp = await userAppsStore.create({
    workspace_id: string,
    name: string,
    url: string,
    icon?: string,
    color?: string,
    logo_url?: string,  // Auto-fetched if not provided
    category?: string,
    description?: string,
    iframe_config?: Record<string, any>,
    open_on_startup?: boolean,
    app_type?: string
})

// Update app
await userAppsStore.update(appId: string, workspaceId: string, {
    name?: string,
    url?: string,
    logo_url?: string,
    category?: string,
    // ... other fields
})

// Delete app
await userAppsStore.delete(appId: string, workspaceId: string)

// Update 3D position
await userAppsStore.updatePosition(appId: string, {
    position_x: number,
    position_y: number,
    position_z: number
})

// Track usage
await userAppsStore.recordOpened(appId: string)

// Get startup apps
const startupApps = await userAppsStore.fetchStartupApps(workspaceId: string)
```

#### Mock Data for Development

When the backend API is unavailable (development mode), the store automatically falls back to mock data:

```typescript
// frontend/src/lib/stores/userAppsStore.ts:438-526
function getMockUserApps(workspaceId: string): UserApp[] {
    return [
        {
            id: 'mock-app-1',
            name: 'Notion',
            url: 'https://notion.so',
            icon: 'FileText',
            color: '#000000',
            logo_url: 'https://www.notion.so/images/favicon.ico',
            category: 'productivity',
            // ...
        },
        // ... more mock apps
    ];
}
```

---

## Desktop Integration

### Window Store Integration

File: `/Users/rhl/Desktop/BusinessOS2/frontend/src/lib/stores/windowStore.ts`

User apps are automatically registered with the `windowStore` for desktop integration.

#### Registration Flow

**When app is installed:**
```typescript
// userAppsStore.ts:206-214
windowStore.registerUserApp({
    id: newApp.id,
    name: newApp.name,
    url: newApp.url,
    icon: newApp.icon,
    color: newApp.color,
    logo_url: newApp.logo_url
});
```

**windowStore.registerUserApp implementation:**
```typescript
// windowStore.ts:1169-1277
registerUserApp: (app: {
    id: string;
    name: string;
    url: string;
    icon: string;
    color: string;
    logo_url?: string | null
}) => {
    const moduleId = `user-app-${app.id}`;

    // Add to moduleDefaults for window sizing
    moduleDefaults[moduleId] = {
        title: app.name,
        width: 1000,
        height: 700,
        minWidth: 600,
        minHeight: 400,
    };

    // Create desktop icon (x: -3, dedicated column for user apps)
    const newIcon: DesktopIcon = {
        id: `icon-${moduleId}`,
        module: moduleId,
        label: app.name,
        x: -3,  // Third column from right
        y: nextY,
        type: 'app',
        customIcon: app.logo_url ? {
            type: 'image',
            imageUrl: app.logo_url,
            backgroundColor: app.color || '#6366F1',
        } : {
            type: 'lucide',
            lucideName: app.icon || 'AppWindow',
            foregroundColor: app.color || '#6366F1',
            backgroundColor: '#F3E8FF',
        }
    };

    // Add to desktop icons array
    update(state => ({
        ...state,
        desktopIcons: [...state.desktopIcons, newIcon]
    }));

    saveSettings(newState);
}
```

#### Desktop Icon Positioning

User apps are placed in a dedicated column on the desktop:
- **x: -3** - Third column from the right
- **y: 0, 1, 2, ...** - Incrementing row position

This separates user apps from core BusinessOS modules (x: -1, -2) and makes them easy to identify.

#### Logo Display Priority

1. **logo_url** (highest priority) - Actual app logo fetched from URL
2. **Lucide icon** (fallback) - Generic icon with custom color

#### Unregistration

**When app is deleted:**
```typescript
// userAppsStore.ts:329
windowStore.unregisterUserApp(appId);
```

This removes:
- Desktop icon
- Module defaults
- Any open windows for the app

---

## Usage Examples

### Installing an App (Frontend)

```typescript
import { userAppsStore } from '$lib/stores/userAppsStore';

// Install Notion
await userAppsStore.create({
    workspace_id: 'workspace-uuid',
    name: 'Notion',
    url: 'https://notion.so',
    category: 'productivity',
    description: 'Notes and documentation',
    // logo_url auto-fetched from URL
    // icon and color have defaults
});

// Result: App appears on desktop with logo
```

### Custom App Installation

```typescript
// User adds a custom app via form
await userAppsStore.create({
    workspace_id: 'workspace-uuid',
    name: 'My Custom Tool',
    url: 'https://mytool.com/dashboard',
    category: 'business',
    description: 'Internal CRM',
    color: '#FF5733',
    open_on_startup: true  // Auto-open on desktop load
});
```

### Fetching User Apps

```typescript
// Fetch all active apps
await userAppsStore.fetch('workspace-uuid');

// Access via store subscription
$userAppsStore.apps.forEach(app => {
    console.log(app.name, app.url, app.logo_url);
});

// Fetch including inactive
await userAppsStore.fetch('workspace-uuid', true);
```

### Opening an App

User apps open as iframe windows when their desktop icon is clicked. The module ID format is:

```
user-app-{app.id}
```

Example:
```typescript
// Desktop icon click handler
windowStore.openWindow(`user-app-${app.id}`, {
    title: app.name,
    data: { url: app.url }
});
```

---

## Database Queries (SQLC)

File: `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/database/queries/user_external_apps.sql`

### Key Queries

#### ListUserExternalApps
```sql
-- Get all active external apps for a workspace
SELECT * FROM user_external_apps
WHERE workspace_id = $1 AND is_active = true
ORDER BY created_at DESC;
```

#### CreateUserExternalApp
```sql
INSERT INTO user_external_apps (
    user_id, workspace_id, name, url, icon, color, logo_url,
    category, description, iframe_config, open_on_startup, app_type
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;
```

#### UpdateUserExternalApp
```sql
UPDATE user_external_apps
SET
    name = COALESCE(sqlc.narg('name'), name),
    url = COALESCE(sqlc.narg('url'), url),
    logo_url = COALESCE(sqlc.narg('logo_url'), logo_url),
    category = COALESCE(sqlc.narg('category'), category),
    -- ... other fields
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;
```

#### GetStartupApps
```sql
-- Get all apps configured to open on startup
SELECT * FROM user_external_apps
WHERE workspace_id = $1
  AND is_active = true
  AND open_on_startup = true
ORDER BY created_at ASC;
```

#### RecordAppOpened
```sql
-- Track usage
UPDATE user_external_apps
SET last_opened_at = NOW()
WHERE id = $1;
```

---

## Features in Detail

### 1. Logo Auto-Fetching

When a user installs an app without providing a logo, the system automatically fetches the app's favicon:

**Backend (Go):**
```go
// internal/handlers/user_apps.go:234-246
fetchedLogoURL, err := h.faviconFetcher.FetchFaviconURL(req.URL)
if err != nil {
    logging.Warn("[UserApps] Failed to fetch favicon for %s: %v", req.URL, err)
} else {
    logoURL = fetchedLogoURL
}
```

**FaviconFetcher** uses Google's Favicon API:
```
https://www.google.com/s2/favicons?domain={domain}&sz=128
```

This ensures high-quality app logos without manual uploads.

### 2. Workspace-Specific Apps

Each app is scoped to a workspace:
- Users in different workspaces see different app collections
- Deleting a workspace cascades to delete its apps (foreign key constraint)

```sql
workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE
```

### 3. Soft Delete (is_active)

Apps can be disabled without deletion:

```typescript
// Disable app
await userAppsStore.update(appId, workspaceId, {
    is_active: false
});

// Fetch including inactive
await userAppsStore.fetch(workspaceId, true);
```

### 4. Auto-Launch on Startup

Apps can be configured to open automatically when the desktop loads:

```typescript
await userAppsStore.create({
    // ... other fields
    open_on_startup: true
});

// Later, on desktop load:
const startupApps = await userAppsStore.fetchStartupApps(workspaceId);
startupApps.forEach(app => {
    windowStore.openWindow(`user-app-${app.id}`);
});
```

### 5. Position Persistence

The 3D desktop position is persisted across sessions:

```typescript
// When window is moved
await userAppsStore.updatePosition(appId, {
    position_x: 100,
    position_y: 200,
    position_z: 1
});

// On next load, window opens at saved position
```

### 6. Usage Tracking

The system tracks when apps are opened:

```typescript
// When app window is opened
await userAppsStore.recordOpened(appId);

// Updates `last_opened_at` timestamp
// Can be used for analytics, "recently used" lists, etc.
```

---

## Integration Points

### With Desktop Environment

1. **Desktop Icons**
   - User apps appear as icons in x: -3 column
   - Use actual app logos (not generic Lucide icons)
   - Persist position and customization

2. **Window System**
   - Apps open in iframe windows
   - Module ID format: `user-app-{app.id}`
   - Support minimize, maximize, resize

3. **Dock**
   - User apps can be pinned to dock
   - Show active state indicator

### With Iframe System

Apps render in iframe windows with sandboxing:

```json
{
  "sandbox": [
    "allow-same-origin",
    "allow-scripts",
    "allow-popups",
    "allow-forms"
  ],
  "allowFullscreen": true
}
```

This provides security while allowing apps to function.

---

## Security Considerations

### 1. Iframe Sandboxing

All user apps run in sandboxed iframes:
- `allow-same-origin` - Required for many apps to function
- `allow-scripts` - JavaScript execution
- `allow-popups` - For OAuth flows, external links
- `allow-forms` - Form submission

**Note:** Some apps (like Notion) may have X-Frame-Options that prevent iframe embedding.

### 2. Workspace Isolation

- Apps are scoped to workspaces via foreign key
- Users can't access apps from other workspaces
- Cascade delete on workspace removal

### 3. User Ownership

- Apps are tied to `user_id`
- Only the creating user (and workspace members) can manage apps

### 4. Input Validation

Backend validates:
- URL format (must be valid HTTP/HTTPS)
- Color format (hex codes)
- Category values (from predefined list)

---

## Future Enhancements

### 1. Native App Support

The system has groundwork for native app capture:
- `app_type` field supports "web" or "native"
- Future: Detect and integrate macOS/Windows apps

### 2. App Permissions

Expand iframe_config to support:
- Clipboard access
- Camera/microphone permissions
- Geolocation

### 3. App Store Analytics

Track:
- Most installed apps
- Most frequently opened apps
- Usage time per app

### 4. Team App Libraries

Workspace-level app catalogs:
- Admins pre-install apps for team
- Auto-provision apps for new workspace members

### 5. OAuth Integration

For apps requiring authentication:
- Store encrypted OAuth tokens
- Auto-login to apps
- Token refresh handling

---

## Troubleshooting

### Apps Not Loading

**Issue:** User apps don't appear on desktop after installation.

**Solution:**
1. Check `userAppsStore.fetch()` was called with correct `workspaceId`
2. Verify `windowStore.registerUserApp()` was called after creation
3. Check browser console for errors
4. Ensure app is `is_active: true`

### Logo Not Auto-Fetching

**Issue:** Apps install without logos.

**Solution:**
1. Check backend logs for FaviconFetcher errors
2. Verify URL is accessible (some apps block favicon API)
3. Manually provide `logo_url` in create request
4. Fallback to Lucide icon + color

### Iframe Not Rendering

**Issue:** App URL loads but shows blank iframe.

**Solution:**
1. Check browser console for X-Frame-Options errors
2. Some apps (Notion, Google Docs) block iframe embedding
3. Try opening in new tab instead (external link)
4. Update `iframe_config` sandbox settings

### Duplicate App Names

**Issue:** Cannot install app - name already exists.

**Solution:**
- Database constraint: one app name per workspace
- Use different name or delete existing app first
- Names are case-sensitive

---

## File Reference

### Backend Files

```
desktop/backend-go/
├── internal/
│   ├── handlers/
│   │   └── user_apps.go              # API endpoints
│   ├── database/
│   │   ├── migrations/
│   │   │   └── 047_user_external_apps.sql  # Schema migration
│   │   ├── queries/
│   │   │   └── user_external_apps.sql      # SQLC queries
│   │   └── sqlc/                      # Generated code (DO NOT EDIT)
│   └── utils/
│       └── favicon.go                 # Logo fetching utility
```

### Frontend Files

```
frontend/
├── src/
│   ├── routes/(app)/
│   │   └── app-store/
│   │       └── +page.svelte           # App Store page
│   ├── lib/
│   │   ├── stores/
│   │   │   ├── userAppsStore.ts       # User apps state management
│   │   │   └── windowStore.ts         # Desktop integration
│   │   └── components/desktop/
│   │       └── AppRegistryModal.svelte # App catalog UI
```

---

## API Reference Summary

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/user-apps?workspace_id={id}` | List active apps |
| GET | `/api/user-apps/:id?workspace_id={id}` | Get app details |
| POST | `/api/user-apps` | Install app |
| PUT | `/api/user-apps/:id?workspace_id={id}` | Update app |
| DELETE | `/api/user-apps/:id?workspace_id={id}` | Delete app |
| PUT | `/api/user-apps/:id/position` | Update position |
| POST | `/api/user-apps/:id/open` | Record usage |
| GET | `/api/user-apps/startup?workspace_id={id}` | Get startup apps |

### Frontend Store API

```typescript
// Install app
await userAppsStore.create(params: CreateUserAppParams)

// Fetch apps
await userAppsStore.fetch(workspaceId: string, includeInactive?: boolean)

// Update app
await userAppsStore.update(appId: string, workspaceId: string, params: UpdateUserAppParams)

// Delete app
await userAppsStore.delete(appId: string, workspaceId: string)

// Track usage
await userAppsStore.recordOpened(appId: string)

// Get startup apps
await userAppsStore.fetchStartupApps(workspaceId: string)
```

---

## Conclusion

The **App Store Integration System** provides a comprehensive solution for users to extend BusinessOS with external web applications. With automatic logo fetching, desktop integration, and workspace-specific management, users can create a personalized business environment with their favorite tools.

**Key Strengths:**
- Seamless desktop integration (icons, windows, dock)
- Auto-fetched app logos for professional appearance
- 100+ pre-configured popular business apps
- Workspace-scoped app management
- Usage tracking and analytics-ready
- Future-proof architecture (native app support planned)

**Documentation Version:** 1.0.0
**Last Updated:** January 2025
**Authors:** BusinessOS Team
