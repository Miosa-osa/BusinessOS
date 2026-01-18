# Workspace Memory UI - Developer Guide

## Quick Start

### Basic Usage

```svelte
<script>
  import { WorkspaceMemoryPanel } from '$lib/components/workspace';
</script>

<WorkspaceMemoryPanel />
```

That's it! The component handles everything automatically:
- Loads memories from current workspace
- Displays all accessible memories (workspace + private + shared)
- Provides filtering, searching, pinning, deleting
- Handles sharing with team members

---

## Component API

### WorkspaceMemoryPanel

Main component for displaying and managing workspace memories.

**Props:**
```typescript
interface Props {
  onMemoryClick?: (memory: WorkspaceMemoryListItem) => void;
}
```

**Example with callback:**
```svelte
<script>
  import { WorkspaceMemoryPanel } from '$lib/components/workspace';

  function handleClick(memory) {
    console.log('Memory clicked:', memory.title);
    // Navigate, open modal, etc.
  }
</script>

<WorkspaceMemoryPanel onMemoryClick={handleClick} />
```

---

### MemoryVisibilitySelector

Standalone visibility filter component.

**Props:**
```typescript
interface Props {
  selected: MemoryVisibility | 'all';
  onChange: (visibility: MemoryVisibility | 'all') => void;
  label?: string;
}
```

**Example:**
```svelte
<script>
  import { MemoryVisibilitySelector } from '$lib/components/workspace';

  let visibility = $state('all');
</script>

<MemoryVisibilitySelector
  selected={visibility}
  onChange={(v) => visibility = v}
/>
```

---

### MemorySharingModal

Modal for changing visibility and sharing with users.

**Props:**
```typescript
interface Props {
  memory: WorkspaceMemoryListItem | null;
  onClose: () => void;
  onComplete: () => void;
}
```

**Example:**
```svelte
<script>
  import { MemorySharingModal } from '$lib/components/workspace';

  let sharingMemory = $state(null);

  function share(memory) {
    sharingMemory = memory;
  }

  function handleComplete() {
    sharingMemory = null;
    // Refresh list
  }
</script>

{#if sharingMemory}
  <MemorySharingModal
    memory={sharingMemory}
    onClose={() => sharingMemory = null}
    onComplete={handleComplete}
  />
{/if}
```

---

## API Client Functions

### Import

```typescript
import {
  createWorkspaceMemory,
  listWorkspaceMemories,
  listPrivateMemories,
  listAccessibleMemories,
  shareMemory,
  unshareMemory,
  deleteWorkspaceMemory,
  pinWorkspaceMemory
} from '$lib/api/workspaces/memory';
```

### Create Memory

```typescript
const memory = await createWorkspaceMemory(workspaceId, {
  title: 'Important Decision',
  content: 'We decided to...',
  memory_type: 'decision',
  visibility: 'workspace', // or 'private' or 'shared'
  workspace_id: workspaceId,
  tags: ['architecture', 'frontend']
});
```

### List Memories

```typescript
// List workspace-wide memories
const workspaceMemories = await listWorkspaceMemories(workspaceId);

// List private memories
const privateMemories = await listPrivateMemories(workspaceId);

// List all accessible memories (recommended)
const allMemories = await listAccessibleMemories(workspaceId, {
  visibility: 'workspace', // optional filter
  memory_type: 'decision',
  is_pinned: true,
  limit: 50
});
```

### Share Memory

```typescript
// First, set visibility to 'shared'
await updateWorkspaceMemory(workspaceId, memoryId, {
  visibility: 'shared'
});

// Then share with specific users
await shareMemory(workspaceId, memoryId, {
  user_ids: ['user-id-1', 'user-id-2']
});
```

### Unshare Memory

```typescript
await unshareMemory(workspaceId, memoryId, {
  user_ids: ['user-id-1']
});
```

### Pin Memory

```typescript
await pinWorkspaceMemory(workspaceId, memoryId, true); // pin
await pinWorkspaceMemory(workspaceId, memoryId, false); // unpin
```

### Delete Memory

```typescript
await deleteWorkspaceMemory(workspaceId, memoryId);
```

---

## Filtering

All list functions support comprehensive filtering:

```typescript
const memories = await listAccessibleMemories(workspaceId, {
  // Filter by memory type
  memory_type: 'decision', // fact, preference, decision, event, learning, context, relationship

  // Filter by visibility
  visibility: 'workspace', // workspace, private, shared

  // Filter by status
  is_pinned: true,
  is_active: true,

  // Filter by importance
  min_importance: 0.7, // 0.0 to 1.0

  // Filter by tags
  tags: ['architecture', 'frontend'],

  // Pagination
  limit: 50,
  offset: 0
});
```

---

## Visibility Levels

### Workspace
- Visible to all workspace members
- Default for team knowledge
- Green badge/icon
- Cannot be shared with specific users

### Private
- Only visible to creator
- For personal notes
- Red badge/icon
- Can be changed to shared later

### Shared
- Visible to creator + specific users
- Selective sharing
- Blue badge/icon
- Shows share count

---

## Integration with Workspace Store

The components automatically integrate with the workspace store:

```typescript
import { currentWorkspaceId, currentWorkspaceMembers } from '$lib/stores/workspaces';

// Components use these stores automatically
// No manual workspace management needed
```

**Reactive Updates:**
- When workspace changes, memories automatically reload
- When members change, sharing modal updates
- All updates are reactive via Svelte stores

---

## Styling & Theming

### CSS Custom Properties

The components use CSS custom properties for theming:

```css
--color-bg           /* Background color */
--color-bg-secondary /* Secondary background */
--color-bg-tertiary  /* Tertiary background */
--color-text         /* Text color */
--color-text-muted   /* Muted text */
--color-border       /* Border color */
```

### Dark Mode

All components support dark mode automatically:

```css
:global(.dark) .component {
  /* Dark mode styles */
}
```

### Color Scheme

```css
Workspace: #22c55e (green)
Private:   #ef4444 (red)
Shared:    #3b82f6 (blue)
Primary:   #3b82f6 (blue)
```

---

## Common Patterns

### Pattern 1: Display Memories in Page

```svelte
<script>
  import { WorkspaceMemoryPanel } from '$lib/components/workspace';
</script>

<div class="container">
  <h1>Workspace Memories</h1>
  <WorkspaceMemoryPanel />
</div>
```

### Pattern 2: Custom Memory Handler

```svelte
<script>
  import { WorkspaceMemoryPanel } from '$lib/components/workspace';
  import { goto } from '$app/navigation';

  function handleMemoryClick(memory) {
    goto(`/memories/${memory.id}`);
  }
</script>

<WorkspaceMemoryPanel onMemoryClick={handleMemoryClick} />
```

### Pattern 3: Create + List Memories

```svelte
<script>
  import { WorkspaceMemoryPanel } from '$lib/components/workspace';
  import { createWorkspaceMemory } from '$lib/api/workspaces/memory';
  import { currentWorkspaceId } from '$lib/stores/workspaces';

  async function createMemory() {
    await createWorkspaceMemory($currentWorkspaceId, {
      title: 'New Memory',
      content: 'Content here',
      memory_type: 'fact',
      visibility: 'workspace',
      workspace_id: $currentWorkspaceId
    });

    // Panel will auto-refresh
  }
</script>

<button on:click={createMemory}>Create Memory</button>
<WorkspaceMemoryPanel />
```

### Pattern 4: Filter Memories Externally

```svelte
<script>
  import { listAccessibleMemories } from '$lib/api/workspaces/memory';
  import { currentWorkspaceId } from '$lib/stores/workspaces';

  let memories = $state([]);
  let filter = $state('all');

  async function loadMemories() {
    memories = await listAccessibleMemories($currentWorkspaceId, {
      visibility: filter === 'all' ? undefined : filter
    });
  }

  $effect(() => {
    loadMemories();
  });
</script>

<select bind:value={filter}>
  <option value="all">All</option>
  <option value="workspace">Workspace</option>
  <option value="private">Private</option>
  <option value="shared">Shared</option>
</select>

{#each memories as memory}
  <div>{memory.title}</div>
{/each}
```

---

## Error Handling

All API functions can throw errors. Always use try-catch:

```typescript
try {
  const memories = await listAccessibleMemories(workspaceId);
} catch (error) {
  console.error('Failed to load memories:', error);
  // Show error toast, fallback UI, etc.
}
```

The components handle errors internally but you should handle them when using the API directly.

---

## TypeScript Types

### Core Types

```typescript
import type {
  WorkspaceMemory,
  WorkspaceMemoryListItem,
  MemoryVisibility,
  CreateWorkspaceMemoryData,
  UpdateWorkspaceMemoryData,
  WorkspaceMemoryFilters,
  ShareMemoryData,
  UnshareMemoryData
} from '$lib/api/workspaces/memory';
```

### Example Usage

```typescript
import type { WorkspaceMemoryListItem } from '$lib/api/workspaces/memory';

let memories: WorkspaceMemoryListItem[] = [];

function handleClick(memory: WorkspaceMemoryListItem) {
  console.log(memory.title);
}
```

---

## Performance Tips

### 1. Use Proper Limits

```typescript
// Good - limits results
const memories = await listAccessibleMemories(workspaceId, { limit: 50 });

// Bad - loads everything
const memories = await listAccessibleMemories(workspaceId);
```

### 2. Filter Server-Side

```typescript
// Good - filters on server
const memories = await listAccessibleMemories(workspaceId, {
  visibility: 'workspace',
  memory_type: 'decision'
});

// Bad - filters client-side
const all = await listAccessibleMemories(workspaceId);
const filtered = all.filter(m => m.visibility === 'workspace');
```

### 3. Debounce Search

```typescript
import { debounce } from '$lib/utils';

const searchMemories = debounce(async (query) => {
  // Search logic
}, 300);
```

---

## Accessibility

All components are accessible:
- ✓ Keyboard navigation
- ✓ ARIA labels
- ✓ Focus management
- ✓ Screen reader support

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Tab | Navigate through memories |
| Enter | Open memory detail |
| Space | Pin/unpin memory |
| Delete | Delete memory (when focused) |

---

## Testing

### Component Testing

```typescript
import { render, screen } from '@testing-library/svelte';
import WorkspaceMemoryPanel from '$lib/components/workspace/WorkspaceMemoryPanel.svelte';

test('displays memories', async () => {
  render(WorkspaceMemoryPanel);

  expect(screen.getByText('Workspace Memories')).toBeInTheDocument();
});
```

### API Testing

```typescript
import { listAccessibleMemories } from '$lib/api/workspaces/memory';

test('fetches memories', async () => {
  const memories = await listAccessibleMemories('workspace-id');

  expect(Array.isArray(memories)).toBe(true);
});
```

---

## Troubleshooting

### Memories not loading?
1. Check workspace ID is set: `console.log($currentWorkspaceId)`
2. Verify user is authenticated
3. Check network tab for API errors
4. Verify backend is running

### Sharing not working?
1. Ensure visibility is set to 'shared'
2. Check user has permission to share
3. Verify user IDs are correct
4. Check backend logs

### Styles not applying?
1. Ensure CSS custom properties are defined
2. Check dark mode class is present
3. Verify no CSS conflicts

---

## Best Practices

### Do's ✓
- Use `listAccessibleMemories()` for most cases
- Set proper limits on API calls
- Handle errors gracefully
- Use TypeScript types
- Test components

### Don'ts ✗
- Don't load unlimited memories
- Don't bypass visibility checks
- Don't modify store state directly
- Don't forget error handling
- Don't ignore accessibility

---

## Migration Guide

### From General Memory API

**Before:**
```typescript
import { getMemories } from '$lib/api/memory';
const memories = await getMemories({ project_id: projectId });
```

**After:**
```typescript
import { listAccessibleMemories } from '$lib/api/workspaces/memory';
const memories = await listAccessibleMemories(workspaceId);
```

### From Old Components

**Before:**
```svelte
<MemoryPanel projectId={projectId} />
```

**After:**
```svelte
<WorkspaceMemoryPanel />
```

---

## Additional Resources

- **Backend API Docs:** `/docs/api_rag_endpoints.md`
- **Database Schema:** `/desktop/backend-go/internal/database/migrations/030_memory_hierarchy_v2.sql`
- **Backend Handlers:** `/desktop/backend-go/internal/handlers/workspace_memory_handlers.go`
- **Backend Service:** `/desktop/backend-go/internal/services/memory_hierarchy_service.go`

---

## Support

For issues or questions:
1. Check this guide
2. Check implementation complete doc
3. Review backend API docs
4. Check network tab for errors
5. Ask team for help

---

**Last Updated:** 2026-01-06
**Version:** 1.0.0
**Status:** Production Ready
