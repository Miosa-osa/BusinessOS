# P0: Workspaces - Team Collaboration

> **Priority:** P0 - Critical for Beta
> **Backend Status:** Complete (24 endpoints)
> **Frontend Status:** Not Started
> **Estimated Effort:** 2-3 sprints

---

## Overview

Workspaces enable multi-user collaboration in BusinessOS. Currently the entire platform is single-user. This is the **most critical missing feature** for enterprise customers paying $15K.

---

## Backend API Endpoints (Ready to Use)

### Workspace CRUD
| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/workspaces` | Create workspace |
| GET | `/api/workspaces` | List user's workspaces |
| GET | `/api/workspaces/:id` | Get workspace details |
| PUT | `/api/workspaces/:id` | Update workspace (admin+) |
| DELETE | `/api/workspaces/:id` | Delete workspace (owner only) |

### Members & Roles
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/workspaces/:id/members` | List workspace members |
| POST | `/api/workspaces/:id/members/invite` | Invite members (manager+) |
| PUT | `/api/workspaces/:id/members/:userId` | Update member role (admin+) |
| DELETE | `/api/workspaces/:id/members/:userId` | Remove member (admin+) |
| GET | `/api/workspaces/:id/roles` | List available roles |

### User Profile in Workspace
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/workspaces/:id/profile` | Get user's workspace profile |
| PUT | `/api/workspaces/:id/profile` | Update workspace profile |
| GET | `/api/workspaces/:id/role-context` | Get user's role & permissions |

### Email Invitations
| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/workspaces/:id/invites` | Create email invite (manager+) |
| GET | `/api/workspaces/:id/invites` | List pending invites (admin+) |
| DELETE | `/api/workspaces/:id/invites/:inviteId` | Revoke invite (admin+) |
| POST | `/api/workspaces/invites/accept` | Accept public invite |

### Audit Logging
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/workspaces/:id/audit-logs` | List audit logs (admin+) |
| GET | `/api/workspaces/:id/audit-logs/:logId` | Get audit log details |
| GET | `/api/workspaces/:id/audit-logs/user/:userId` | Get user activity |
| GET | `/api/workspaces/:id/audit-logs/resource/:type/:id` | Get resource history |
| GET | `/api/workspaces/:id/audit-logs/stats/actions` | Action statistics |
| GET | `/api/workspaces/:id/audit-logs/stats/active-users` | Most active users |

---

## Role Hierarchy

```
owner       → Full control, can delete workspace, transfer ownership
admin       → Manage members, settings, view audit logs
manager     → Invite members, manage projects/contexts
member      → View and edit assigned content
viewer      → Read-only access
```

---

## Frontend Implementation Tasks

### Phase 1: Core Workspace UI

#### 1.1 Workspace Selector (Global)
**File:** `src/lib/components/workspace/WorkspaceSelector.svelte`

- [ ] Dropdown in top-left of sidebar showing current workspace
- [ ] List all user's workspaces with icons/colors
- [ ] "Create Workspace" button
- [ ] Visual indicator for workspace role (owner/admin/member badge)

#### 1.2 Create Workspace Modal
**File:** `src/lib/components/workspace/CreateWorkspaceModal.svelte`

- [ ] Name input (required)
- [ ] Description textarea
- [ ] Icon/color picker
- [ ] Create button → calls `POST /api/workspaces`

#### 1.3 Workspace Settings Page
**File:** `src/routes/(app)/settings/workspace/+page.svelte`

- [ ] General settings (name, description, icon)
- [ ] Danger zone: Delete workspace (owner only)
- [ ] Transfer ownership (owner only)

### Phase 2: Member Management

#### 2.1 Members List
**File:** `src/routes/(app)/settings/workspace/members/+page.svelte`

- [ ] Table with columns: Avatar, Name, Email, Role, Joined, Actions
- [ ] Role dropdown for each member (admin+ can change)
- [ ] Remove member button with confirmation
- [ ] Invite member button

#### 2.2 Invite Modal
**File:** `src/lib/components/workspace/InviteMemberModal.svelte`

- [ ] Email input (single or comma-separated)
- [ ] Role selector (default: member)
- [ ] Optional message
- [ ] Send invite button → calls `POST /api/workspaces/:id/invites`

#### 2.3 Pending Invites List
**File:** `src/lib/components/workspace/PendingInvites.svelte`

- [ ] List pending invitations
- [ ] Resend button
- [ ] Revoke button

#### 2.4 Accept Invite Page
**File:** `src/routes/invite/[token]/+page.svelte`

- [ ] Public page for accepting invites
- [ ] Show workspace name, inviter name
- [ ] Accept/Decline buttons
- [ ] Redirect to workspace on accept

### Phase 3: Audit Logs (Admin Feature)

#### 3.1 Audit Log Page
**File:** `src/routes/(app)/settings/workspace/audit/+page.svelte`

- [ ] Searchable activity feed
- [ ] Filters: User, Action type, Date range, Resource
- [ ] Export to CSV

#### 3.2 Activity Stats Dashboard
- [ ] Most active users chart
- [ ] Actions by type pie chart
- [ ] Activity timeline

### Phase 4: API Client

#### 4.1 Workspaces API Module
**File:** `src/lib/api/workspaces/workspaces.ts` (extend existing)

```typescript
// Add these functions:
export async function createWorkspace(data: CreateWorkspaceInput): Promise<Workspace>
export async function updateWorkspace(id: string, data: UpdateWorkspaceInput): Promise<Workspace>
export async function deleteWorkspace(id: string): Promise<void>
export async function getWorkspaceMembers(id: string): Promise<WorkspaceMember[]>
export async function inviteMember(workspaceId: string, email: string, role: string): Promise<void>
export async function updateMemberRole(workspaceId: string, userId: string, role: string): Promise<void>
export async function removeMember(workspaceId: string, userId: string): Promise<void>
export async function getAuditLogs(workspaceId: string, filters?: AuditLogFilters): Promise<AuditLog[]>
// ... etc
```

#### 4.2 Workspace Store
**File:** `src/lib/stores/workspaces.ts` (extend existing)

```typescript
// Add:
currentWorkspaceId: string | null
members: WorkspaceMember[]
pendingInvites: Invite[]
userRole: WorkspaceRole
canManageMembers: boolean // derived from role
canViewAudit: boolean // derived from role
```

---

## UI/UX Requirements

### Workspace Branding
- Each workspace can have: name, description, icon (emoji), color
- Color appears as accent throughout workspace context

### Permission-Based UI
- Hide/disable controls based on user's role
- Show clear "You don't have permission" messages

### Switching Workspaces
- Should be instant (no page reload)
- All data (projects, contexts, etc.) scoped to workspace
- URL should include workspace slug: `/w/my-workspace/projects`

---

## Database Impact

Existing tables already have `workspace_id` column:
- `projects`
- `contexts`
- `nodes`
- `clients`
- `team_members`
- `conversations`
- etc.

Backend handles scoping automatically based on auth context.

---

## Testing Requirements

- [ ] Unit tests for workspace store
- [ ] Component tests for WorkspaceSelector
- [ ] E2E: Create workspace flow
- [ ] E2E: Invite and accept flow
- [ ] E2E: Role-based permission checks

---

## Linear Issues to Create

1. **[WORKSPACE-001]** Implement WorkspaceSelector component
2. **[WORKSPACE-002]** Create workspace creation flow
3. **[WORKSPACE-003]** Build members management page
4. **[WORKSPACE-004]** Implement invite system
5. **[WORKSPACE-005]** Add audit log viewer
6. **[WORKSPACE-006]** Update routing for workspace context
7. **[WORKSPACE-007]** Add workspace settings page
8. **[WORKSPACE-008]** E2E tests for workspace flows

---

## Dependencies

- None (backend complete)

## Blockers

- None identified

---

## Notes

- This is the foundation for all multi-user features
- Must be implemented before team features make sense
- Consider workspace limits for different pricing tiers
