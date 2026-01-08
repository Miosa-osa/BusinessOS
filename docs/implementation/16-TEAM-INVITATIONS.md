# Team Invitations UI

> **Priority:** P0 - Critical
> **Owner:** Javaris
> **Linear Issue:** CUS-29
> **Backend Status:** Complete
> **Frontend Status:** Not Started
> **Estimated Effort:** 1-2 days

---

## Overview

The backend has full workspace invitation support. We need the frontend UI to:
1. Send invitations via email
2. Show pending invitations
3. Revoke pending invitations
4. Accept invitations (public link)

---

## Backend API Endpoints (Ready to Use)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/workspaces/:id/invites` | Create invitation |
| GET | `/api/workspaces/:id/invites` | List pending invites |
| DELETE | `/api/workspaces/:id/invites/:inviteId` | Revoke invite |
| POST | `/api/workspaces/invites/accept` | Accept invite (public) |

### Create Invitation Request
```typescript
POST /api/workspaces/:id/invites
{
  "email": "newuser@example.com",
  "role": "member" | "manager" | "admin"
}
```

### Response
```typescript
{
  "id": "uuid",
  "email": "newuser@example.com",
  "role": "member",
  "invite_token": "abc123...",
  "invite_url": "https://app.businessos.com/invite/abc123",
  "expires_at": "2026-01-16T00:00:00Z",
  "created_at": "2026-01-09T00:00:00Z"
}
```

---

## Implementation Tasks

### 1. Create Invite Modal Component
**File:** `frontend/src/lib/components/team/InviteMemberModal.svelte`

```svelte
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { workspaceStore } from '$lib/stores/workspace';

  export let open = false;

  let email = '';
  let role = 'member';
  let loading = false;
  let error = '';

  const dispatch = createEventDispatcher();

  async function sendInvite() {
    loading = true;
    error = '';

    try {
      const response = await fetch(`/api/workspaces/${$workspaceStore.id}/invites`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, role })
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Failed to send invitation');
      }

      dispatch('invited');
      open = false;
      email = '';
      role = 'member';
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
</script>

{#if open}
<div class="fixed inset-0 z-50 flex items-center justify-center">
  <div class="absolute inset-0 bg-black/50" on:click={() => open = false}></div>

  <div class="relative bg-white rounded-xl shadow-2xl w-full max-w-md p-6">
    <h2 class="text-xl font-semibold mb-4">Invite Team Member</h2>

    <form on:submit|preventDefault={sendInvite}>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Email Address
          </label>
          <input
            type="email"
            bind:value={email}
            placeholder="colleague@company.com"
            required
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Role
          </label>
          <select
            bind:value={role}
            class="w-full px-3 py-2 border border-gray-300 rounded-lg"
          >
            <option value="member">Member</option>
            <option value="manager">Manager</option>
            <option value="admin">Admin</option>
          </select>
        </div>

        {#if error}
          <p class="text-red-500 text-sm">{error}</p>
        {/if}
      </div>

      <div class="flex justify-end gap-3 mt-6">
        <button
          type="button"
          on:click={() => open = false}
          class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={loading}
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? 'Sending...' : 'Send Invitation'}
        </button>
      </div>
    </form>
  </div>
</div>
{/if}
```

### 2. Add Invite Button to Team Page
**File:** `frontend/src/routes/(app)/team/+page.svelte`

Add to the page header:
```svelte
<script>
  import InviteMemberModal from '$lib/components/team/InviteMemberModal.svelte';
  let showInviteModal = false;
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Team</h1>
  <button
    on:click={() => showInviteModal = true}
    class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
  >
    <UserPlus class="w-4 h-4" />
    Invite Member
  </button>
</div>

<InviteMemberModal bind:open={showInviteModal} on:invited={loadTeam} />
```

### 3. Create Pending Invitations List
**File:** `frontend/src/lib/components/team/PendingInvitations.svelte`

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { workspaceStore } from '$lib/stores/workspace';
  import { Clock, X, Mail } from 'lucide-svelte';

  let invitations = [];
  let loading = true;

  onMount(loadInvitations);

  async function loadInvitations() {
    loading = true;
    const res = await fetch(`/api/workspaces/${$workspaceStore.id}/invites`);
    if (res.ok) {
      invitations = await res.json();
    }
    loading = false;
  }

  async function revokeInvite(id: string) {
    await fetch(`/api/workspaces/${$workspaceStore.id}/invites/${id}`, {
      method: 'DELETE'
    });
    invitations = invitations.filter(i => i.id !== id);
  }

  function formatExpiry(date: string) {
    const d = new Date(date);
    const now = new Date();
    const days = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    return `Expires in ${days} days`;
  }
</script>

{#if invitations.length > 0}
<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
  <h3 class="font-medium text-yellow-800 mb-3 flex items-center gap-2">
    <Clock class="w-4 h-4" />
    Pending Invitations ({invitations.length})
  </h3>

  <div class="space-y-2">
    {#each invitations as invite}
      <div class="flex items-center justify-between bg-white p-3 rounded-lg border">
        <div class="flex items-center gap-3">
          <Mail class="w-4 h-4 text-gray-400" />
          <div>
            <p class="font-medium">{invite.email}</p>
            <p class="text-sm text-gray-500">
              {invite.role} - {formatExpiry(invite.expires_at)}
            </p>
          </div>
        </div>
        <button
          on:click={() => revokeInvite(invite.id)}
          class="p-1 hover:bg-red-100 rounded text-red-500"
          title="Revoke invitation"
        >
          <X class="w-4 h-4" />
        </button>
      </div>
    {/each}
  </div>
</div>
{/if}
```

### 4. Create Accept Invitation Page
**File:** `frontend/src/routes/invite/[token]/+page.svelte`

```svelte
<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  let status: 'loading' | 'success' | 'error' = 'loading';
  let error = '';
  let workspace = null;

  onMount(async () => {
    const token = $page.params.token;

    try {
      const res = await fetch('/api/workspaces/invites/accept', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token })
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || 'Invalid or expired invitation');
      }

      workspace = await res.json();
      status = 'success';

      // Redirect to workspace after 2 seconds
      setTimeout(() => goto('/'), 2000);
    } catch (e) {
      status = 'error';
      error = e.message;
    }
  });
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50">
  <div class="bg-white p-8 rounded-xl shadow-lg max-w-md w-full text-center">
    {#if status === 'loading'}
      <div class="animate-spin w-8 h-8 border-2 border-blue-600 border-t-transparent rounded-full mx-auto mb-4"></div>
      <p class="text-gray-600">Accepting invitation...</p>
    {:else if status === 'success'}
      <div class="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-6 h-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
      </div>
      <h1 class="text-xl font-bold mb-2">Welcome to {workspace?.name}!</h1>
      <p class="text-gray-600">Redirecting you now...</p>
    {:else}
      <div class="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <svg class="w-6 h-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </div>
      <h1 class="text-xl font-bold mb-2">Invitation Failed</h1>
      <p class="text-gray-600 mb-4">{error}</p>
      <a href="/login" class="text-blue-600 hover:underline">Go to login</a>
    {/if}
  </div>
</div>
```

---

## Checklist

- [ ] Create `InviteMemberModal.svelte`
- [ ] Add invite button to team page header
- [ ] Create `PendingInvitations.svelte`
- [ ] Add pending invitations to team page
- [ ] Create `/invite/[token]` route for accepting
- [ ] Add success toast after sending invite
- [ ] Test the full flow

---

## Testing

1. Open team page
2. Click "Invite Member"
3. Enter email and select role
4. Verify invitation appears in pending list
5. Check email was sent (or copy invite link)
6. Open invite link in incognito
7. Verify user is added to workspace

---

## Notes

- Invitations expire after 7 days (backend default)
- Only managers and admins can send invitations
- Need to handle "already a member" case gracefully
