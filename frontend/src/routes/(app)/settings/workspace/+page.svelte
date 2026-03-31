<script lang="ts">
  import { onMount } from 'svelte';
  import { currentWorkspace, currentUserRole, currentWorkspaceRoles, currentWorkspaceMembers, currentUserRoleContext } from '$lib/stores/workspaces';
  import {
    getWorkspace,
    getWorkspaceMembers,
    getWorkspaceRoles,
    getWorkspaceInvites,
    getUserRoleContext,
    type Workspace,
    type WorkspaceMember,
    type WorkspaceRole,
    type WorkspaceInvite,
    type UserRoleContext,
  } from '$lib/api/workspaces';
  import {
    Building2,
    Users,
    Shield,
    Mail,
    Loader2,
    AlertCircle,
    Brain
  } from 'lucide-svelte';
  import WorkspaceGeneralSettings from '$lib/components/workspace/WorkspaceGeneralSettings.svelte';
  import WorkspaceMembersList from '$lib/components/workspace/WorkspaceMembersList.svelte';
  import WorkspaceRolesList from '$lib/components/workspace/WorkspaceRolesList.svelte';
  import WorkspaceInvitesList from '$lib/components/workspace/WorkspaceInvitesList.svelte';
  import WorkspaceMemoryPanel from '$lib/components/workspace/WorkspaceMemoryPanel.svelte';

  type TabType = 'general' | 'members' | 'roles' | 'invites' | 'memories';

  let activeTab = $state<TabType>('general');
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let isMockData = $state(false);

  // Workspace data
  let workspace = $state<Workspace | null>(null);
  let members = $state<WorkspaceMember[]>([]);
  let roles = $state<WorkspaceRole[]>([]);
  let invites = $state<WorkspaceInvite[]>([]);
  let roleContext = $state<UserRoleContext | null>(null);

  // Permissions
  let canManageSettings = $state(false);
  let canManageMembers = $state(false);
  let canManageRoles = $state(false);
  let canInviteMembers = $state(false);

  onMount(async () => {
    await loadWorkspaceData();
  });

  async function loadWorkspaceData() {
    if (!$currentWorkspace?.id) {
      error = 'No workspace selected';
      isLoading = false;
      return;
    }

    // Check if using mock data (mock workspace IDs start with 'mock-')
    if ($currentWorkspace.id.startsWith('mock-')) {
      if (import.meta.env.DEV) console.log('[Workspace Settings] Using mock workspace data');
      isMockData = true;
      
      // Use data from stores instead of API
      workspace = $currentWorkspace;
      roles = $currentWorkspaceRoles || [];
      members = $currentWorkspaceMembers || [];
      invites = [];
      roleContext = $currentUserRoleContext;
      
      if (roleContext) {
        checkPermissions(roleContext);
      } else {
        // Default permissions for mock data
        canManageSettings = true;
        canManageMembers = true;
        canManageRoles = true;
        canInviteMembers = true;
      }
      
      isLoading = false;
      return;
    }

    try {
      isLoading = true;
      error = null;

      // Load all data in parallel
      const [workspaceData, membersData, rolesData, invitesData, roleContextData] =
        await Promise.all([
          getWorkspace($currentWorkspace.id),
          getWorkspaceMembers($currentWorkspace.id),
          getWorkspaceRoles($currentWorkspace.id),
          getWorkspaceInvites($currentWorkspace.id).catch(() => []),
          getUserRoleContext($currentWorkspace.id),
        ]);

      workspace = workspaceData;
      members = membersData;
      roles = rolesData;
      invites = invitesData;
      roleContext = roleContextData;

      // Check permissions
      checkPermissions(roleContextData);
    } catch (err) {
      console.error('Failed to load workspace data:', err);
      error = err instanceof Error ? err.message : 'Failed to load workspace data';
    } finally {
      isLoading = false;
    }
  }

  function checkPermissions(context: UserRoleContext) {
    const perms = context.permissions;

    // Check workspace management permissions
    canManageSettings =
      perms.workspace?.manage === true ||
      context.role_name === 'owner' ||
      context.role_name === 'admin';

    // Check member management permissions
    canManageMembers =
      perms.members?.manage === true ||
      context.role_name === 'owner' ||
      context.role_name === 'admin';

    // Check role management permissions
    canManageRoles =
      perms.roles?.manage === true ||
      context.role_name === 'owner' ||
      context.role_name === 'admin';

    // Check invite permissions
    canInviteMembers =
      perms.members?.invite === true ||
      context.role_name === 'owner' ||
      context.role_name === 'admin' ||
      context.role_name === 'manager';
  }

  function handleWorkspaceUpdated(updated: Workspace) {
    workspace = updated;
    if ($currentWorkspace) {
      $currentWorkspace = { ...$currentWorkspace, ...updated };
    }
  }

  function handleMembersUpdated() {
    loadWorkspaceData();
  }

  function handleRolesUpdated() {
    loadWorkspaceData();
  }

  function handleInvitesUpdated() {
    loadWorkspaceData();
  }

  let tabs = $derived([
    {
      id: 'general' as TabType,
      label: 'General',
      icon: Building2,
      show: true,
    },
    {
      id: 'members' as TabType,
      label: 'Members',
      icon: Users,
      show: true,
      badge: members.length,
    },
    {
      id: 'roles' as TabType,
      label: 'Roles',
      icon: Shield,
      show: true,
      badge: roles.length,
    },
    {
      id: 'invites' as TabType,
      label: 'Invitations',
      icon: Mail,
      show: canManageMembers || canInviteMembers,
      badge: invites.filter((i) => i.status === 'pending').length,
    },
    {
      id: 'memories' as TabType,
      label: 'Memories',
      icon: Brain,
      show: true,
    },
  ]);
</script>

<div class="workspace-settings-page">
  <div class="settings-header">
    <div class="header-content">
      <div class="header-title">
        <Building2 class="w-6 h-6" />
        <h1>Workspace Settings</h1>
      </div>
      {#if workspace}
        <div class="workspace-info">
          <span class="workspace-name">{workspace.name}</span>
          <span class="workspace-plan">{workspace.plan_type}</span>
        </div>
      {/if}
    </div>
  </div>

  {#if isLoading}
    <div class="loading-state">
      <Loader2 class="w-8 h-8 animate-spin" />
      <p>Loading workspace settings...</p>
    </div>
  {:else if error}
    <div class="error-state">
      <AlertCircle class="w-8 h-8" />
      <p>{error}</p>
    </div>
  {:else if workspace}
    {#if isMockData}
      <div class="mock-data-banner">
        <AlertCircle class="w-4 h-4" />
        <span>Viewing demo workspace data. Create a real workspace to enable full functionality.</span>
      </div>
    {/if}
    <div class="settings-container">
      <!-- Tabs -->
      <div class="settings-tabs">
        {#each tabs as tab}
          {#if tab.show}
            <button
              class="tab-button"
              class:active={activeTab === tab.id}
              onclick={() => (activeTab = tab.id)}
            >
              <tab.icon class="w-4 h-4" />
              <span>{tab.label}</span>
              {#if tab.badge !== undefined && tab.badge > 0}
                <span class="tab-badge">{tab.badge}</span>
              {/if}
            </button>
          {/if}
        {/each}
      </div>

      <!-- Tab Content -->
      <div class="settings-content">
        {#if activeTab === 'general'}
          <WorkspaceGeneralSettings
            {workspace}
            canManage={canManageSettings}
            isOwner={roleContext?.role_name === 'owner'}
            onupdated={handleWorkspaceUpdated}
          />
        {:else if activeTab === 'members'}
          <WorkspaceMembersList
            workspaceId={workspace.id}
            {members}
            {roles}
            currentUserRole={roleContext?.role_name ?? 'member'}
            currentUserId={roleContext?.user_id ?? ''}
            canManage={canManageMembers}
            canInvite={canInviteMembers}
            onupdated={handleMembersUpdated}
          />
        {:else if activeTab === 'roles'}
          <WorkspaceRolesList
            workspaceId={workspace.id}
            {roles}
            canManage={canManageRoles}
            onupdated={handleRolesUpdated}
          />
        {:else if activeTab === 'invites'}
          <WorkspaceInvitesList
            workspaceId={workspace.id}
            {invites}
            {roles}
            canManage={canManageMembers}
            canInvite={canInviteMembers}
            onupdated={handleInvitesUpdated}
          />
        {:else if activeTab === 'memories'}
          <WorkspaceMemoryPanel />
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .workspace-settings-page {
    min-height: 100vh;
    background: var(--dbg2);
  }

  .settings-header {
    background: var(--dbg);
    border-bottom: 1px solid var(--dbd);
    padding: 1.5rem 2rem;
  }

  .header-content {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .header-title {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .header-title h1 {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--dt);
    margin: 0;
  }

  .workspace-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .workspace-name {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--dt2);
  }

  .workspace-plan {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    background: var(--bos-status-info-bg);
    color: var(--bos-status-info);
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    border-radius: 9999px;
  }

  .loading-state,
  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    gap: 1rem;
    color: var(--dt3);
  }

  .error-state {
    color: var(--bos-status-error);
  }

  .mock-data-banner {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background: var(--bos-status-warning-bg);
    border: 1px solid var(--bos-status-warning);
    border-radius: 0.5rem;
    color: var(--bos-status-warning);
    font-size: 0.875rem;
    margin: 1rem 2rem;
    max-width: 1200px;
    margin-left: auto;
    margin-right: auto;
  }

  .settings-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
  }

  .settings-tabs {
    display: flex;
    gap: 2px;
    border-bottom: 1px solid var(--dbd);
    margin-bottom: 1.5rem;
    overflow-x: auto;
  }
  .settings-tabs::-webkit-scrollbar { display: none; }

  .tab-button {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    transition: color 0.15s;
    color: var(--dt3);
    font-size: 13px;
    font-weight: 500;
    white-space: nowrap;
    margin-bottom: -1px;
  }

  .tab-button:hover {
    color: var(--dt2);
  }

  .tab-button.active {
    color: var(--dt);
    border-bottom-color: var(--dt);
    font-weight: 600;
  }

  .tab-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    background: var(--dbg3);
    color: var(--dt2);
    font-size: 11px;
    font-weight: 600;
    border-radius: 9999px;
  }

  .tab-button.active .tab-badge {
    background: var(--dt);
    color: var(--dbg);
  }

  .settings-content {
    background: var(--dbg);
    border-radius: 0.5rem;
    border: 1px solid var(--dbd);
    min-height: 400px;
  }
</style>
