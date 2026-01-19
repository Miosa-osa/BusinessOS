# Audit Log Viewer

> **Priority:** P1 - High Value
> **Owner:** Roberto
> **Linear Issue:** CUS-50
> **Backend Status:** Complete (full audit logging)
> **Frontend Status:** Not Started
> **Estimated Effort:** 2-3 days

---

## Overview

The backend logs all workspace activity. We need a UI to:
1. View activity timeline
2. Filter by user, action type, date
3. See who did what and when
4. Drill down into specific events

---

## Backend API Endpoints (Ready to Use)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/workspaces/:id/audit-logs` | List audit logs |
| GET | `/api/workspaces/:id/audit-logs/:logId` | Get specific log |
| GET | `/api/workspaces/:id/audit-logs/user/:userId` | Get user activity |
| GET | `/api/workspaces/:id/audit-logs/resource/:type/:id` | Get resource history |
| GET | `/api/workspaces/:id/audit-logs/stats/actions` | Action statistics |
| GET | `/api/workspaces/:id/audit-logs/stats/active-users` | Most active users |

### List Audit Logs
```typescript
GET /api/workspaces/:id/audit-logs?limit=50&offset=0&action=create&user_id=xxx

Response:
{
  "logs": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "user_name": "John Doe",
      "user_avatar": "https://...",
      "action": "create",
      "resource_type": "project",
      "resource_id": "uuid",
      "resource_name": "Q1 Marketing",
      "details": { "fields_changed": ["name", "status"] },
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2026-01-09T12:00:00Z"
    }
  ],
  "total": 1250,
  "has_more": true
}
```

### Action Types
```typescript
type AuditAction =
  | 'create' | 'update' | 'delete'
  | 'login' | 'logout'
  | 'invite' | 'join' | 'leave'
  | 'role_change'
  | 'export' | 'import';
```

### Resource Types
```typescript
type ResourceType =
  | 'project' | 'task' | 'client'
  | 'table' | 'row'
  | 'member' | 'invitation'
  | 'integration' | 'settings';
```

---

## Implementation Tasks

### 1. Create Audit Logs Page
**File:** `frontend/src/routes/(app)/settings/audit-logs/+page.svelte`

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { workspaceStore } from '$lib/stores/workspace';
  import {
    History, Filter, ChevronDown, User, Calendar,
    Plus, Pencil, Trash2, LogIn, LogOut, UserPlus, Download, Upload
  } from 'lucide-svelte';

  let logs = [];
  let loading = true;
  let total = 0;
  let offset = 0;
  const limit = 50;

  // Filters
  let actionFilter = '';
  let userFilter = '';
  let dateFrom = '';
  let dateTo = '';
  let showFilters = false;

  const actionIcons = {
    create: Plus,
    update: Pencil,
    delete: Trash2,
    login: LogIn,
    logout: LogOut,
    invite: UserPlus,
    export: Download,
    import: Upload
  };

  const actionColors = {
    create: 'text-green-600 bg-green-100',
    update: 'text-blue-600 bg-blue-100',
    delete: 'text-red-600 bg-red-100',
    login: 'text-purple-600 bg-purple-100',
    logout: 'text-gray-600 bg-gray-100',
    invite: 'text-indigo-600 bg-indigo-100',
    export: 'text-orange-600 bg-orange-100',
    import: 'text-teal-600 bg-teal-100'
  };

  onMount(loadLogs);

  async function loadLogs() {
    loading = true;
    const params = new URLSearchParams({
      limit: String(limit),
      offset: String(offset)
    });

    if (actionFilter) params.set('action', actionFilter);
    if (userFilter) params.set('user_id', userFilter);
    if (dateFrom) params.set('from', dateFrom);
    if (dateTo) params.set('to', dateTo);

    const res = await fetch(`/api/workspaces/${$workspaceStore.id}/audit-logs?${params}`);
    if (res.ok) {
      const data = await res.json();
      logs = data.logs;
      total = data.total;
    }
    loading = false;
  }

  function formatTime(dateStr: string) {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();

    if (diff < 60000) return 'Just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`;

    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit'
    });
  }

  function getActionDescription(log: any) {
    const action = log.action;
    const type = log.resource_type;
    const name = log.resource_name || 'item';

    switch (action) {
      case 'create': return `Created ${type} "${name}"`;
      case 'update': return `Updated ${type} "${name}"`;
      case 'delete': return `Deleted ${type} "${name}"`;
      case 'login': return 'Logged in';
      case 'logout': return 'Logged out';
      case 'invite': return `Invited ${name}`;
      case 'join': return 'Joined workspace';
      case 'leave': return 'Left workspace';
      case 'role_change': return `Changed role to ${log.details?.new_role}`;
      case 'export': return `Exported ${type}`;
      case 'import': return `Imported ${type}`;
      default: return `${action} ${type}`;
    }
  }

  function nextPage() {
    offset += limit;
    loadLogs();
  }

  function prevPage() {
    offset = Math.max(0, offset - limit);
    loadLogs();
  }

  function applyFilters() {
    offset = 0;
    loadLogs();
  }

  function clearFilters() {
    actionFilter = '';
    userFilter = '';
    dateFrom = '';
    dateTo = '';
    offset = 0;
    loadLogs();
  }
</script>

<div class="max-w-5xl mx-auto p-6">
  <!-- Header -->
  <div class="flex items-center justify-between mb-6">
    <div class="flex items-center gap-3">
      <History class="w-6 h-6 text-gray-700" />
      <h1 class="text-2xl font-bold">Audit Log</h1>
    </div>
    <button
      on:click={() => showFilters = !showFilters}
      class="flex items-center gap-2 px-3 py-2 border rounded-lg hover:bg-gray-50"
    >
      <Filter class="w-4 h-4" />
      Filters
      <ChevronDown class="w-4 h-4 {showFilters ? 'rotate-180' : ''}" />
    </button>
  </div>

  <!-- Filters Panel -->
  {#if showFilters}
    <div class="bg-white border rounded-xl p-4 mb-6">
      <div class="grid grid-cols-4 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Action</label>
          <select
            bind:value={actionFilter}
            class="w-full px-3 py-2 border rounded-lg"
          >
            <option value="">All actions</option>
            <option value="create">Create</option>
            <option value="update">Update</option>
            <option value="delete">Delete</option>
            <option value="login">Login</option>
            <option value="invite">Invite</option>
            <option value="export">Export</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">From Date</label>
          <input
            type="date"
            bind:value={dateFrom}
            class="w-full px-3 py-2 border rounded-lg"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">To Date</label>
          <input
            type="date"
            bind:value={dateTo}
            class="w-full px-3 py-2 border rounded-lg"
          />
        </div>
        <div class="flex items-end gap-2">
          <button
            on:click={applyFilters}
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Apply
          </button>
          <button
            on:click={clearFilters}
            class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-lg"
          >
            Clear
          </button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Activity Timeline -->
  <div class="bg-white border rounded-xl">
    {#if loading}
      <div class="p-12 text-center">
        <div class="w-8 h-8 border-2 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
        <p class="text-gray-500">Loading activity...</p>
      </div>
    {:else if logs.length === 0}
      <div class="p-12 text-center">
        <History class="w-12 h-12 text-gray-300 mx-auto mb-3" />
        <p class="text-gray-500">No activity found</p>
      </div>
    {:else}
      <div class="divide-y">
        {#each logs as log}
          <div class="flex items-start gap-4 p-4 hover:bg-gray-50">
            <!-- User Avatar -->
            <div class="flex-shrink-0">
              {#if log.user_avatar}
                <img src={log.user_avatar} alt="" class="w-10 h-10 rounded-full" />
              {:else}
                <div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center">
                  <User class="w-5 h-5 text-gray-500" />
                </div>
              {/if}
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <span class="font-medium">{log.user_name || 'Unknown'}</span>
                <span class="text-gray-500">·</span>
                <span class="text-sm text-gray-500">{formatTime(log.created_at)}</span>
              </div>
              <p class="text-gray-700">{getActionDescription(log)}</p>
              {#if log.details?.fields_changed}
                <p class="text-sm text-gray-500 mt-1">
                  Changed: {log.details.fields_changed.join(', ')}
                </p>
              {/if}
            </div>

            <!-- Action Badge -->
            <div class="flex-shrink-0">
              <span class="inline-flex items-center gap-1 px-2 py-1 rounded text-xs font-medium {actionColors[log.action] || 'bg-gray-100 text-gray-600'}">
                <svelte:component this={actionIcons[log.action] || History} class="w-3 h-3" />
                {log.action}
              </span>
            </div>
          </div>
        {/each}
      </div>

      <!-- Pagination -->
      <div class="flex items-center justify-between px-4 py-3 border-t bg-gray-50">
        <span class="text-sm text-gray-500">
          Showing {offset + 1}-{Math.min(offset + limit, total)} of {total}
        </span>
        <div class="flex gap-2">
          <button
            on:click={prevPage}
            disabled={offset === 0}
            class="px-3 py-1 border rounded hover:bg-white disabled:opacity-50"
          >
            Previous
          </button>
          <button
            on:click={nextPage}
            disabled={offset + limit >= total}
            class="px-3 py-1 border rounded hover:bg-white disabled:opacity-50"
          >
            Next
          </button>
        </div>
      </div>
    {/if}
  </div>
</div>
```

### 2. Add to Settings Navigation
```svelte
<a href="/settings/audit-logs" class="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-100">
  <History class="w-4 h-4" />
  Audit Log
</a>
```

### 3. Activity Stats Component (Optional)
**File:** `frontend/src/lib/components/audit/ActivityStats.svelte`

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { workspaceStore } from '$lib/stores/workspace';

  let actionStats = [];
  let activeUsers = [];

  onMount(async () => {
    const [actionsRes, usersRes] = await Promise.all([
      fetch(`/api/workspaces/${$workspaceStore.id}/audit-logs/stats/actions`),
      fetch(`/api/workspaces/${$workspaceStore.id}/audit-logs/stats/active-users`)
    ]);

    if (actionsRes.ok) actionStats = await actionsRes.json();
    if (usersRes.ok) activeUsers = await usersRes.json();
  });
</script>

<div class="grid grid-cols-2 gap-4 mb-6">
  <!-- Actions This Week -->
  <div class="bg-white border rounded-xl p-4">
    <h3 class="font-medium mb-3">Actions This Week</h3>
    <div class="space-y-2">
      {#each actionStats.slice(0, 5) as stat}
        <div class="flex items-center justify-between">
          <span class="text-gray-600 capitalize">{stat.action}</span>
          <span class="font-medium">{stat.count}</span>
        </div>
      {/each}
    </div>
  </div>

  <!-- Most Active Users -->
  <div class="bg-white border rounded-xl p-4">
    <h3 class="font-medium mb-3">Most Active</h3>
    <div class="space-y-2">
      {#each activeUsers.slice(0, 5) as user}
        <div class="flex items-center justify-between">
          <span class="text-gray-600">{user.name}</span>
          <span class="font-medium">{user.action_count} actions</span>
        </div>
      {/each}
    </div>
  </div>
</div>
```

---

## Checklist

- [ ] Create `/settings/audit-logs` route
- [ ] Load and display activity timeline
- [ ] Implement filters (action, date range)
- [ ] Add pagination
- [ ] Color-code action types
- [ ] Show user avatars
- [ ] Format timestamps (relative time)
- [ ] Add to settings navigation
- [ ] Optional: Add activity stats component

---

## UX Requirements

- Clear visual distinction between action types
- Infinite scroll or pagination for large lists
- Quick filters for common actions
- Relative timestamps for recent activity

---

## Access Control

- Only workspace admins can view audit logs
- Members cannot see audit log page
- Show permission error if unauthorized

---

## Future Enhancements

- Export audit logs to CSV
- Real-time activity feed (SSE)
- Advanced search (by resource, text)
- Activity graphs and charts
