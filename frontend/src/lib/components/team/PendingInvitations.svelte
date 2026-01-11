<script lang="ts">
  import { Clock, X, Mail, Copy, Check } from 'lucide-svelte';
  import { currentWorkspace } from '$lib/stores/workspaces';
  import { getWorkspaceInvites, revokeWorkspaceInvite, type WorkspaceInvite } from '$lib/api/workspaces';
  import { onMount } from 'svelte';

  let invitations = $state<WorkspaceInvite[]>([]);
  let loading = $state(true);
  let copiedId = $state<string | null>(null);

  onMount(() => {
    loadInvitations();
  });

  async function loadInvitations() {
    const workspaceId = $currentWorkspace?.id;
    if (!workspaceId) {
      // Load mock data in dev mode
      if (import.meta.env.DEV) {
        invitations = getMockInvitations();
      }
      loading = false;
      return;
    }

    try {
      loading = true;
      invitations = await getWorkspaceInvites(workspaceId);
    } catch (err) {
      console.error('Failed to load invitations:', err);
      // Fall back to mock in dev
      if (import.meta.env.DEV) {
        invitations = getMockInvitations();
      }
    } finally {
      loading = false;
    }
  }

  function getMockInvitations(): WorkspaceInvite[] {
    return [
      {
        id: 'inv-1',
        workspace_id: 'mock-workspace-001',
        email: 'john.doe@example.com',
        role: 'member',
        status: 'pending',
        token: 'abc123',
        invited_by: 'mock-user-001',
        expires_at: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString(),
        created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
      },
      {
        id: 'inv-2',
        workspace_id: 'mock-workspace-001',
        email: 'jane.smith@company.com',
        role: 'admin',
        status: 'pending',
        token: 'def456',
        invited_by: 'mock-user-001',
        expires_at: new Date(Date.now() + 6 * 24 * 60 * 60 * 1000).toISOString(),
        created_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
      },
    ];
  }

  async function revokeInvite(id: string) {
    const workspaceId = $currentWorkspace?.id;
    
    // In dev/mock mode, just remove locally
    if (!workspaceId || workspaceId.startsWith('mock-')) {
      invitations = invitations.filter(i => i.id !== id);
      return;
    }

    try {
      await revokeWorkspaceInvite(workspaceId, id);
      invitations = invitations.filter(i => i.id !== id);
    } catch (err) {
      console.error('Failed to revoke invitation:', err);
    }
  }

  async function copyInviteLink(invite: WorkspaceInvite) {
    const link = `${window.location.origin}/invite/${invite.token}`;
    await navigator.clipboard.writeText(link);
    copiedId = invite.id;
    setTimeout(() => copiedId = null, 2000);
  }

  function formatExpiry(date: string) {
    const d = new Date(date);
    const now = new Date();
    const days = Math.ceil((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    if (days <= 0) return 'Expired';
    if (days === 1) return 'Expires tomorrow';
    return `Expires in ${days} days`;
  }

  function getRoleBadgeColor(role: string) {
    switch (role) {
      case 'admin': return 'bg-purple-100 text-purple-700';
      case 'manager': return 'bg-amber-100 text-amber-700';
      default: return 'bg-gray-100 text-gray-700';
    }
  }

  // Only show pending invitations
  const pendingInvitations = $derived(invitations.filter(i => i.status === 'pending'));
</script>

{#if loading}
  <div class="mx-6 mt-4 p-4 bg-gray-50 border border-gray-200 rounded-lg animate-pulse">
    <div class="h-5 w-40 bg-gray-200 rounded mb-3"></div>
    <div class="h-12 bg-gray-200 rounded"></div>
  </div>
{:else if pendingInvitations.length > 0}
  <div class="mx-6 mt-4 p-4 bg-amber-50 border border-amber-200 rounded-lg">
    <h3 class="font-medium text-amber-800 mb-3 flex items-center gap-2">
      <Clock class="w-4 h-4" />
      Pending Invitations ({pendingInvitations.length})
    </h3>

    <div class="space-y-2">
      {#each pendingInvitations as invite (invite.id)}
        <div class="flex items-center justify-between bg-white p-3 rounded-lg border border-amber-100 shadow-sm">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-full bg-amber-100 flex items-center justify-center">
              <Mail class="w-4 h-4 text-amber-600" />
            </div>
            <div>
              <p class="text-sm font-semibold text-gray-900 tracking-tight">{invite.email}</p>
              <div class="flex items-center gap-2 mt-0.5">
                <span class="text-xs px-2 py-0.5 rounded-full {getRoleBadgeColor(invite.role)}">
                  {invite.role}
                </span>
                <span class="text-xs text-gray-500">
                  {formatExpiry(invite.expires_at)}
                </span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-1">
            <button
              onclick={() => copyInviteLink(invite)}
              class="p-2 hover:bg-gray-100 rounded-lg text-gray-500 hover:text-gray-700 transition-colors"
              title="Copy invite link"
            >
              {#if copiedId === invite.id}
                <Check class="w-4 h-4 text-green-600" />
              {:else}
                <Copy class="w-4 h-4" />
              {/if}
            </button>
            <button
              onclick={() => revokeInvite(invite.id)}
              class="p-2 hover:bg-red-100 rounded-lg text-gray-500 hover:text-red-600 transition-colors"
              title="Revoke invitation"
            >
              <X class="w-4 h-4" />
            </button>
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}
