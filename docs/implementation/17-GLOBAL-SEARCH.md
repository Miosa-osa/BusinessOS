# Global Search (Spotlight)

> **Priority:** P0 - Critical
> **Owner:** Roberto
> **Linear Issue:** CUS-53
> **Backend Status:** Complete (multimodal search API)
> **Frontend Status:** Not Started
> **Estimated Effort:** 2-3 days

---

## Overview

Build a Spotlight-style global search modal (Cmd+K / Ctrl+K) that searches across all content types:
- Tables & rows
- Projects
- Clients
- Tasks
- Conversations
- Documents
- Team members

---

## Backend API Endpoints (Ready to Use)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/search/hybrid` | Hybrid semantic + keyword search |
| POST | `/api/search/multimodal` | Search text + images |
| GET | `/api/tables/search` | Search across tables |
| GET | `/api/contexts/search` | Search knowledge base |

### Hybrid Search Request
```typescript
POST /api/search/hybrid
{
  "query": "Q1 marketing budget",
  "types": ["project", "task", "client", "table", "context"],
  "limit": 20
}
```

### Response
```typescript
{
  "results": [
    {
      "id": "uuid",
      "type": "project",
      "title": "Q1 Marketing Campaign",
      "description": "Budget planning for...",
      "url": "/projects/uuid",
      "score": 0.95,
      "highlights": ["Q1", "marketing", "budget"]
    },
    {
      "id": "uuid",
      "type": "task",
      "title": "Review marketing budget",
      "description": "...",
      "url": "/tasks/uuid",
      "score": 0.87
    }
  ],
  "total": 15,
  "query_time_ms": 45
}
```

---

## Implementation Tasks

### 1. Create Search Modal Component
**File:** `frontend/src/lib/components/search/SpotlightSearch.svelte`

```svelte
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';
  import {
    Search, FileText, Users, FolderKanban, CheckSquare,
    Table2, MessageSquare, X
  } from 'lucide-svelte';

  export let open = false;

  let query = '';
  let results: any[] = [];
  let loading = false;
  let selectedIndex = 0;
  let inputRef: HTMLInputElement;

  const typeIcons = {
    project: FolderKanban,
    task: CheckSquare,
    client: Users,
    table: Table2,
    context: FileText,
    conversation: MessageSquare
  };

  const typeLabels = {
    project: 'Project',
    task: 'Task',
    client: 'Client',
    table: 'Table',
    context: 'Document',
    conversation: 'Chat'
  };

  // Keyboard shortcut to open
  function handleKeydown(e: KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      open = !open;
    }

    if (!open) return;

    if (e.key === 'Escape') {
      open = false;
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, results.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, 0);
    } else if (e.key === 'Enter' && results[selectedIndex]) {
      e.preventDefault();
      navigateToResult(results[selectedIndex]);
    }
  }

  onMount(() => {
    if (browser) {
      window.addEventListener('keydown', handleKeydown);
    }
  });

  onDestroy(() => {
    if (browser) {
      window.removeEventListener('keydown', handleKeydown);
    }
  });

  $: if (open && inputRef) {
    setTimeout(() => inputRef?.focus(), 50);
  }

  $: if (!open) {
    query = '';
    results = [];
    selectedIndex = 0;
  }

  // Debounced search
  let searchTimeout: ReturnType<typeof setTimeout>;
  $: {
    if (query.length >= 2) {
      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(() => search(query), 200);
    } else {
      results = [];
    }
  }

  async function search(q: string) {
    loading = true;
    try {
      const res = await fetch('/api/search/hybrid', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query: q,
          types: ['project', 'task', 'client', 'table', 'context', 'conversation'],
          limit: 10
        })
      });

      if (res.ok) {
        const data = await res.json();
        results = data.results || [];
        selectedIndex = 0;
      }
    } catch (e) {
      console.error('Search failed:', e);
    } finally {
      loading = false;
    }
  }

  function navigateToResult(result: any) {
    open = false;
    goto(result.url);
  }
</script>

{#if open}
<div class="fixed inset-0 z-50">
  <!-- Backdrop -->
  <div
    class="absolute inset-0 bg-black/50 backdrop-blur-sm"
    on:click={() => open = false}
  ></div>

  <!-- Modal -->
  <div class="absolute top-[20%] left-1/2 -translate-x-1/2 w-full max-w-2xl">
    <div class="bg-white rounded-xl shadow-2xl overflow-hidden">
      <!-- Search Input -->
      <div class="flex items-center gap-3 px-4 py-3 border-b">
        <Search class="w-5 h-5 text-gray-400" />
        <input
          bind:this={inputRef}
          bind:value={query}
          type="text"
          placeholder="Search everything... (projects, tasks, clients, docs)"
          class="flex-1 text-lg outline-none placeholder:text-gray-400"
        />
        {#if loading}
          <div class="w-5 h-5 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
        {:else}
          <kbd class="px-2 py-1 bg-gray-100 text-gray-500 text-xs rounded">ESC</kbd>
        {/if}
      </div>

      <!-- Results -->
      <div class="max-h-[400px] overflow-y-auto">
        {#if results.length > 0}
          <div class="py-2">
            {#each results as result, i}
              <button
                class="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-gray-50 {i === selectedIndex ? 'bg-blue-50' : ''}"
                on:click={() => navigateToResult(result)}
                on:mouseenter={() => selectedIndex = i}
              >
                <div class="w-8 h-8 rounded-lg bg-gray-100 flex items-center justify-center">
                  <svelte:component this={typeIcons[result.type] || FileText} class="w-4 h-4 text-gray-600" />
                </div>
                <div class="flex-1 min-w-0">
                  <p class="font-medium text-gray-900 truncate">{result.title}</p>
                  {#if result.description}
                    <p class="text-sm text-gray-500 truncate">{result.description}</p>
                  {/if}
                </div>
                <span class="text-xs text-gray-400 bg-gray-100 px-2 py-1 rounded">
                  {typeLabels[result.type] || result.type}
                </span>
              </button>
            {/each}
          </div>
        {:else if query.length >= 2 && !loading}
          <div class="py-12 text-center text-gray-500">
            <Search class="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p>No results found for "{query}"</p>
          </div>
        {:else if query.length < 2}
          <div class="py-12 text-center text-gray-500">
            <p>Type at least 2 characters to search</p>
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="px-4 py-2 border-t bg-gray-50 flex items-center gap-4 text-xs text-gray-500">
        <span class="flex items-center gap-1">
          <kbd class="px-1.5 py-0.5 bg-white border rounded">↑</kbd>
          <kbd class="px-1.5 py-0.5 bg-white border rounded">↓</kbd>
          Navigate
        </span>
        <span class="flex items-center gap-1">
          <kbd class="px-1.5 py-0.5 bg-white border rounded">Enter</kbd>
          Open
        </span>
        <span class="flex items-center gap-1">
          <kbd class="px-1.5 py-0.5 bg-white border rounded">Esc</kbd>
          Close
        </span>
      </div>
    </div>
  </div>
</div>
{/if}
```

### 2. Add to Root Layout
**File:** `frontend/src/routes/(app)/+layout.svelte`

```svelte
<script>
  import SpotlightSearch from '$lib/components/search/SpotlightSearch.svelte';
  let searchOpen = false;
</script>

<!-- Add this near the end of the layout -->
<SpotlightSearch bind:open={searchOpen} />
```

### 3. Create Search Store (Optional)
**File:** `frontend/src/lib/stores/search.ts`

```typescript
import { writable } from 'svelte/store';

export const searchOpen = writable(false);

export function openSearch() {
  searchOpen.set(true);
}

export function closeSearch() {
  searchOpen.set(false);
}

export function toggleSearch() {
  searchOpen.update(v => !v);
}
```

### 4. Add Search Button to Header
**File:** Update header/navbar component

```svelte
<button
  on:click={() => searchOpen.set(true)}
  class="flex items-center gap-2 px-3 py-1.5 bg-gray-100 rounded-lg text-gray-600 hover:bg-gray-200"
>
  <Search class="w-4 h-4" />
  <span class="text-sm">Search</span>
  <kbd class="ml-2 px-1.5 py-0.5 bg-white border rounded text-xs">⌘K</kbd>
</button>
```

---

## API Client (Optional)
**File:** `frontend/src/lib/api/search/search.ts`

```typescript
interface SearchResult {
  id: string;
  type: 'project' | 'task' | 'client' | 'table' | 'context' | 'conversation';
  title: string;
  description?: string;
  url: string;
  score: number;
  highlights?: string[];
}

interface SearchResponse {
  results: SearchResult[];
  total: number;
  query_time_ms: number;
}

export async function hybridSearch(
  query: string,
  types?: string[],
  limit = 10
): Promise<SearchResponse> {
  const response = await fetch('/api/search/hybrid', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, types, limit })
  });

  if (!response.ok) {
    throw new Error('Search failed');
  }

  return response.json();
}
```

---

## Checklist

- [ ] Create `SpotlightSearch.svelte` component
- [ ] Add Cmd+K keyboard shortcut listener
- [ ] Connect to hybrid search API
- [ ] Add result type icons and labels
- [ ] Implement keyboard navigation (up/down/enter)
- [ ] Add search button to header/navbar
- [ ] Add to root layout
- [ ] Test with different content types
- [ ] Add loading states
- [ ] Handle empty/error states

---

## UX Requirements

- **Speed:** Results should appear within 200ms of typing
- **Debounce:** 200ms debounce on input
- **Keyboard-first:** Full keyboard navigation
- **Visual feedback:** Loading spinner, selected state
- **Escape routes:** Click outside or ESC to close

---

## Future Enhancements

- Recent searches history
- Search filters (type, date range)
- Quick actions (create task, new project)
- Voice search integration
- Search analytics
