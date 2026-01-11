import { writable, derived, get } from 'svelte/store';
import type {
  Workspace,
  WorkspaceRole,
  WorkspaceMember,
  UserWorkspaceProfile,
  UserRoleContext,
} from '$lib/api/workspaces';
import {
  getWorkspaces,
  getWorkspace,
  getWorkspaceRoles,
  getWorkspaceMembers,
  getWorkspaceProfile,
  getUserRoleContext,
} from '$lib/api/workspaces';

// ============================================================================
// STATE
// ============================================================================

/**
 * All workspaces the user has access to
 */
export const workspaces = writable<Workspace[]>([]);

/**
 * Currently selected workspace
 * This is the main state that drives workspace-aware features
 */
export const currentWorkspace = writable<Workspace | null>(null);

/**
 * Roles in the current workspace
 */
export const currentWorkspaceRoles = writable<WorkspaceRole[]>([]);

/**
 * Members in the current workspace
 */
export const currentWorkspaceMembers = writable<WorkspaceMember[]>([]);

/**
 * Current user's profile in the current workspace
 */
export const currentWorkspaceProfile = writable<UserWorkspaceProfile | null>(null);

/**
 * Current user's role context (permissions, etc)
 */
export const currentUserRoleContext = writable<UserRoleContext | null>(null);

/**
 * Loading states
 */
export const workspaceLoading = writable({
  workspaces: false,
  switching: false,
  roles: false,
  members: false,
  profile: false,
});

/**
 * Error state
 */
export const workspaceError = writable<string | null>(null);

// ============================================================================
// DERIVED STORES
// ============================================================================

/**
 * Get the current workspace ID (convenience)
 */
export const currentWorkspaceId = derived(
  currentWorkspace,
  ($currentWorkspace) => $currentWorkspace?.id ?? null
);

/**
 * Check if user has a specific permission in current workspace
 */
export const hasPermission = derived(
  currentUserRoleContext,
  ($context) => (resource: string, permission: string): boolean => {
    if (!$context) return false;
    return !!$context.permissions?.[resource]?.[permission];
  }
);

/**
 * Check if user is at least a certain hierarchy level
 */
export const isAtLeastLevel = derived(
  currentUserRoleContext,
  ($context) => (level: number): boolean => {
    if (!$context) return false;
    return $context.hierarchy_level <= level; // Lower number = higher authority
  }
);

/**
 * Get current user's role name
 */
export const currentUserRole = derived(
  currentUserRoleContext,
  ($context) => $context?.role_name ?? null
);

// ============================================================================
// ACTIONS
// ============================================================================

/**
 * Initialize workspace state - load all workspaces
 */
export async function initializeWorkspaces(): Promise<void> {
  workspaceLoading.update((s) => ({ ...s, workspaces: true }));
  workspaceError.set(null);

  try {
    const allWorkspaces = await getWorkspaces() ?? [];
    console.log(`[Workspaces] Loaded ${allWorkspaces.length} workspaces:`, allWorkspaces);
    workspaces.set(allWorkspaces);

    // If no workspace is selected and we have workspaces, select the first one
    const current = get(currentWorkspace);
    if (!current && allWorkspaces.length > 0) {
      console.log(`[Workspaces] Auto-selecting first workspace: ${allWorkspaces[0].name} (${allWorkspaces[0].id})`);
      await switchWorkspace(allWorkspaces[0].id);
    } else if (current) {
      console.log(`[Workspaces] Current workspace already set: ${current.name} (${current.id})`);
    } else if (allWorkspaces.length === 0) {
      // No workspaces - in dev mode, load mock data
      if (import.meta.env.DEV) {
        console.log('[Workspaces] No workspaces found in dev mode, loading mock data');
        loadMockWorkspaceData();
      } else {
        console.debug('[Workspaces] No workspaces available');
      }
    }
  } catch (error) {
    console.error('[Workspaces] Failed to load workspaces:', error);
    workspaceError.set(error instanceof Error ? error.message : 'Failed to load workspaces');
    
    // In dev mode, fall back to mock data on API failure
    if (import.meta.env.DEV) {
      console.log('[Workspaces] API failed in dev mode, loading mock data');
      loadMockWorkspaceData();
      workspaceError.set(null);
    }
  } finally {
    workspaceLoading.update((s) => ({ ...s, workspaces: false }));
  }
}

/**
 * Switch to a different workspace
 * This loads all workspace-specific data
 */
export async function switchWorkspace(workspaceId: string): Promise<void> {
  workspaceLoading.update((s) => ({ ...s, switching: true }));
  workspaceError.set(null);

  try {
    // Load workspace details
    const workspace = await getWorkspace(workspaceId);
    currentWorkspace.set(workspace);

    // Load workspace-specific data in parallel
    const [roles, members, profile, roleContext] = await Promise.all([
      getWorkspaceRoles(workspaceId),
      getWorkspaceMembers(workspaceId),
      getWorkspaceProfile(workspaceId).catch(() => null), // Profile might not exist yet
      getUserRoleContext(workspaceId),
    ]);

    currentWorkspaceRoles.set(roles);
    currentWorkspaceMembers.set(members);
    currentWorkspaceProfile.set(profile);
    currentUserRoleContext.set(roleContext);

    // Save to localStorage for persistence
    localStorage.setItem('businessos_current_workspace_id', workspaceId);

    console.log(`[Workspaces] Switched to workspace: ${workspace.name} (${workspace.slug})`);
    console.log(`[Workspaces] User role: ${roleContext.role_display_name} (Level ${roleContext.hierarchy_level})`);
  } catch (error) {
    console.error('[Workspaces] Failed to switch workspace:', error);
    workspaceError.set(error instanceof Error ? error.message : 'Failed to switch workspace');
    throw error;
  } finally {
    workspaceLoading.update((s) => ({ ...s, switching: false }));
  }
}

/**
 * Refresh current workspace data
 */
export async function refreshCurrentWorkspace(): Promise<void> {
  const current = get(currentWorkspace);
  if (!current) return;
  await switchWorkspace(current.id);
}

/**
 * Load saved workspace from localStorage
 */
export async function loadSavedWorkspace(): Promise<void> {
  const savedId = localStorage.getItem('businessos_current_workspace_id');
  if (savedId) {
    try {
      await switchWorkspace(savedId);
    } catch (error) {
      console.warn('[Workspaces] Failed to load saved workspace, loading all workspaces');
      await initializeWorkspaces();
    }
  } else {
    await initializeWorkspaces();
  }
}

/**
 * Clear workspace state (for logout, etc)
 */
export function clearWorkspaceState(): void {
  workspaces.set([]);
  currentWorkspace.set(null);
  currentWorkspaceRoles.set([]);
  currentWorkspaceMembers.set([]);
  currentWorkspaceProfile.set(null);
  currentUserRoleContext.set(null);
  workspaceError.set(null);
  localStorage.removeItem('businessos_current_workspace_id');
}

/**
 * Load mock workspace data for development/testing
 */
export function loadMockWorkspaceData(): void {
  console.log('[Workspaces] Loading mock data for development');
  
  const mockWorkspace: Workspace = {
    id: 'mock-workspace-001',
    name: 'Test Workspace',
    slug: 'test-workspace',
    description: 'A mock workspace for UI development',
    logo_url: null,
    plan_type: 'professional',
    max_members: 50,
    max_projects: 100,
    max_storage_gb: 50,
    owner_id: 'mock-user-001',
    settings: {},
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
  
  const mockRoles: WorkspaceRole[] = [
    { id: 'role-1', workspace_id: mockWorkspace.id, name: 'owner', display_name: 'Owner', description: 'Full access', hierarchy_level: 0, permissions: {}, is_system: true, is_default: false, color: '#10b981', icon: 'crown', created_at: '', updated_at: '' },
    { id: 'role-2', workspace_id: mockWorkspace.id, name: 'admin', display_name: 'Admin', description: 'Administrative access', hierarchy_level: 10, permissions: {}, is_system: true, is_default: false, color: '#6366f1', icon: 'shield', created_at: '', updated_at: '' },
    { id: 'role-3', workspace_id: mockWorkspace.id, name: 'manager', display_name: 'Manager', description: 'Team management', hierarchy_level: 20, permissions: {}, is_system: true, is_default: false, color: '#f59e0b', icon: 'users', created_at: '', updated_at: '' },
    { id: 'role-4', workspace_id: mockWorkspace.id, name: 'member', display_name: 'Member', description: 'Standard access', hierarchy_level: 30, permissions: {}, is_system: true, is_default: true, color: '#64748b', icon: 'user', created_at: '', updated_at: '' },
    { id: 'role-5', workspace_id: mockWorkspace.id, name: 'guest', display_name: 'Guest', description: 'Limited access', hierarchy_level: 40, permissions: {}, is_system: true, is_default: false, color: '#94a3b8', icon: 'eye', created_at: '', updated_at: '' },
  ];
  
  const mockRoleContext: UserRoleContext = {
    user_id: 'mock-user-001',
    workspace_id: mockWorkspace.id,
    role_name: 'owner',
    role_display_name: 'Owner',
    hierarchy_level: 0,
    permissions: {
      workspace: { manage: true, delete: true },
      members: { view: true, invite: true, manage: true, remove: true },
      roles: { view: true, manage: true },
      projects: { create: true, manage: true, delete: true },
    },
    title: 'Admin',
    department: 'Engineering',
    expertise_areas: [],
  };
  
  workspaces.set([mockWorkspace]);
  currentWorkspace.set(mockWorkspace);
  currentWorkspaceRoles.set(mockRoles);
  currentWorkspaceMembers.set([]);
  currentUserRoleContext.set(mockRoleContext);
  
  console.log('[Workspaces] Mock data loaded - workspace:', mockWorkspace.name);
}

