# Team Permissions UI

> **Priority:** P1 - High Value
> **Owner:** Javaris
> **Linear Issue:** CUS-49
> **Backend Status:** Complete (full RBAC)
> **Frontend Status:** Not Started
> **Estimated Effort:** 2-3 days

---

## Overview

The backend has full role-based access control (RBAC). We need UI to:
1. View workspace roles and permissions
2. Change member roles
3. See who has what access

---

## Backend API Endpoints (Ready to Use)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/workspaces/:id/roles` | List available roles |
| GET | `/api/workspaces/:id/members` | List members with roles |
| PUT | `/api/workspaces/:id/members/:userId` | Update member role |
| GET | `/api/workspaces/:id/role-context` | Get current user's permissions |

### Roles Hierarchy
```typescript
type WorkspaceRole = 'owner' | 'admin' | 'manager' | 'member';

// Permissions by role:
const rolePermissions = {
  owner: ['*'],  // Full control
  admin: ['manage_members', 'manage_settings', 'manage_integrations', 'view_audit'],
  manager: ['invite_members', 'manage_projects', 'view_members'],
  member: ['view_workspace', 'use_features']
};
```

### Get Roles Response
```typescript
GET /api/workspaces/:id/roles
{
  "roles": [
    { "id": "owner", "name": "Owner", "description": "Full workspace control", "member_count": 1 },
    { "id": "admin", "name": "Admin", "description": "Manage members and settings", "member_count": 2 },
    { "id": "manager", "name": "Manager", "description": "Invite members, manage projects", "member_count": 3 },
    { "id": "member", "name": "Member", "description": "Use workspace features", "member_count": 10 }
  ]
}
```

### Update Member Role
```typescript
PUT /api/workspaces/:id/members/:userId
{
  "role": "admin"
}
```

---

## Implementation Tasks

### 1. Create Permissions Page
**File:** `frontend/src/routes/(app)/settings/permissions/+page.svelte`

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { workspaceStore } from '$lib/stores/workspace';
  import { Shield, Users, ChevronRight } from 'lucide-svelte';
  import MemberRoleModal from '$lib/components/team/MemberRoleModal.svelte';

  let roles = [];
  let members = [];
  let loading = true;
  let selectedMember = null;

  onMount(async () => {
    await Promise.all([loadRoles(), loadMembers()]);
    loading = false;
  });

  async function loadRoles() {
    const res = await fetch(`/api/workspaces/${$workspaceStore.id}/roles`);
    if (res.ok) {
      const data = await res.json();
      roles = data.roles;
    }
  }

  async function loadMembers() {
    const res = await fetch(`/api/workspaces/${$workspaceStore.id}/members`);
    if (res.ok) {
      members = await res.json();
    }
  }

  function getRoleBadgeColor(role: string) {
    switch (role) {
      case 'owner': return 'bg-purple-100 text-purple-800';
      case 'admin': return 'bg-red-100 text-red-800';
      case 'manager': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  }
</script>

<div class="max-w-4xl mx-auto p-6">
  <div class="flex items-center gap-3 mb-6">
    <Shield class="w-6 h-6 text-gray-700" />
    <h1 class="text-2xl font-bold">Permissions</h1>
  </div>

  <!-- Roles Overview -->
  <div class="bg-white rounded-xl border mb-6">
    <div class="px-4 py-3 border-b">
      <h2 class="font-semibold">Workspace Roles</h2>
    </div>
    <div class="divide-y">
      {#each roles as role}
        <div class="px-4 py-3 flex items-center justify-between">
          <div>
            <span class="font-medium">{role.name}</span>
            <p class="text-sm text-gray-500">{role.description}</p>
          </div>
          <span class="text-sm text-gray-500">{role.member_count} members</span>
        </div>
      {/each}
    </div>
  </div>

  <!-- Members List -->
  <div class="bg-white rounded-xl border">
    <div class="px-4 py-3 border-b flex items-center justify-between">
      <h2 class="font-semibold">Members</h2>
      <span class="text-sm text-gray-500">{members.length} total</span>
    </div>
    <div class="divide-y">
      {#each members as member}
        <button
          class="w-full px-4 py-3 flex items-center justify-between hover:bg-gray-50 text-left"
          on:click={() => selectedMember = member}
        >
          <div class="flex items-center gap-3">
            {#if member.avatar_url}
              <img src={member.avatar_url} alt="" class="w-10 h-10 rounded-full" />
            {:else}
              <div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center">
                <Users class="w-5 h-5 text-gray-500" />
              </div>
            {/if}
            <div>
              <p class="font-medium">{member.name || member.email}</p>
              <p class="text-sm text-gray-500">{member.email}</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span class="px-2 py-1 rounded text-xs font-medium {getRoleBadgeColor(member.role)}">
              {member.role}
            </span>
            <ChevronRight class="w-4 h-4 text-gray-400" />
          </div>
        </button>
      {/each}
    </div>
  </div>
</div>

{#if selectedMember}
  <MemberRoleModal
    bind:member={selectedMember}
    {roles}
    on:updated={loadMembers}
    on:close={() => selectedMember = null}
  />
{/if}
```

### 2. Create Role Change Modal
**File:** `frontend/src/lib/components/team/MemberRoleModal.svelte`

```svelte
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { workspaceStore } from '$lib/stores/workspace';
  import { X, Shield, AlertTriangle } from 'lucide-svelte';

  export let member: any;
  export let roles: any[];

  const dispatch = createEventDispatcher();

  let selectedRole = member?.role;
  let loading = false;
  let error = '';

  $: canChangeRole = member?.role !== 'owner';
  $: isDowngrade = getRoleLevel(selectedRole) > getRoleLevel(member?.role);

  function getRoleLevel(role: string): number {
    const levels = { owner: 0, admin: 1, manager: 2, member: 3 };
    return levels[role] ?? 3;
  }

  async function updateRole() {
    if (selectedRole === member.role) {
      dispatch('close');
      return;
    }

    loading = true;
    error = '';

    try {
      const res = await fetch(
        `/api/workspaces/${$workspaceStore.id}/members/${member.id}`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ role: selectedRole })
        }
      );

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Failed to update role');
      }

      dispatch('updated');
      dispatch('close');
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center">
  <div class="absolute inset-0 bg-black/50" on:click={() => dispatch('close')}></div>

  <div class="relative bg-white rounded-xl shadow-2xl w-full max-w-md">
    <div class="flex items-center justify-between px-6 py-4 border-b">
      <h2 class="text-lg font-semibold">Change Role</h2>
      <button on:click={() => dispatch('close')} class="p-1 hover:bg-gray-100 rounded">
        <X class="w-5 h-5" />
      </button>
    </div>

    <div class="p-6">
      <!-- Member Info -->
      <div class="flex items-center gap-3 mb-6">
        {#if member.avatar_url}
          <img src={member.avatar_url} alt="" class="w-12 h-12 rounded-full" />
        {:else}
          <div class="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center">
            <Shield class="w-6 h-6 text-gray-500" />
          </div>
        {/if}
        <div>
          <p class="font-medium">{member.name || member.email}</p>
          <p class="text-sm text-gray-500">Current role: {member.role}</p>
        </div>
      </div>

      {#if !canChangeRole}
        <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
          <p class="text-yellow-800 text-sm">
            The workspace owner role cannot be changed. Transfer ownership from workspace settings.
          </p>
        </div>
      {:else}
        <!-- Role Selection -->
        <div class="space-y-2 mb-6">
          <label class="block text-sm font-medium text-gray-700 mb-2">Select New Role</label>
          {#each roles.filter(r => r.id !== 'owner') as role}
            <label
              class="flex items-center gap-3 p-3 border rounded-lg cursor-pointer hover:bg-gray-50
                {selectedRole === role.id ? 'border-blue-500 bg-blue-50' : ''}"
            >
              <input
                type="radio"
                bind:group={selectedRole}
                value={role.id}
                class="text-blue-600"
              />
              <div>
                <p class="font-medium">{role.name}</p>
                <p class="text-sm text-gray-500">{role.description}</p>
              </div>
            </label>
          {/each}
        </div>

        {#if isDowngrade}
          <div class="flex items-start gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-lg mb-4">
            <AlertTriangle class="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
            <p class="text-sm text-yellow-800">
              This will reduce {member.name || 'this user'}'s permissions in the workspace.
            </p>
          </div>
        {/if}

        {#if error}
          <p class="text-red-500 text-sm mb-4">{error}</p>
        {/if}
      {/if}
    </div>

    <div class="flex justify-end gap-3 px-6 py-4 border-t bg-gray-50">
      <button
        on:click={() => dispatch('close')}
        class="px-4 py-2 text-gray-700 hover:bg-gray-200 rounded-lg"
      >
        Cancel
      </button>
      {#if canChangeRole}
        <button
          on:click={updateRole}
          disabled={loading || selectedRole === member.role}
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? 'Saving...' : 'Update Role'}
        </button>
      {/if}
    </div>
  </div>
</div>
```

### 3. Add to Settings Navigation
**File:** Update settings page navigation

```svelte
<a href="/settings/permissions" class="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-100">
  <Shield class="w-4 h-4" />
  Permissions
</a>
```

---

## Checklist

- [ ] Create `/settings/permissions` route
- [ ] Load and display roles overview
- [ ] Display members with their roles
- [ ] Create role change modal
- [ ] Add role level badges (color-coded)
- [ ] Handle owner role restrictions
- [ ] Add to settings navigation
- [ ] Show success/error feedback
- [ ] Test role changes

---

## UX Requirements

- Clear visual hierarchy for roles (owner > admin > manager > member)
- Warn before downgrading someone's role
- Cannot change owner role (must transfer ownership separately)
- Only admins+ can change roles

---

## Future Enhancements

- Custom roles with granular permissions
- Project-level permissions
- Permission audit log
- Bulk role changes
