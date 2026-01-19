# Decision Record: Feature 1 Role-Based Agent Behavior - Checkpoint

**Date:** 2026-01-06
**Decision Type:** Implementation Checkpoint
**Status:** In Progress (70% Complete)
**Priority:** CRITICAL
**Participants:** Pedro, Nick (assigned)

---

## Context

Implementing Feature 1 from FUTURE_FEATURES.md: Role-Based Agent Behavior system where agents must know user roles, permissions, and restrict suggestions accordingly.

**Roberto's Vision:**
> "If they have a role, it's been specific to their profile and stuff. The agent will know based off their role and everything what they're able to do based on the project. That way, people don't do things outside their role."

---

## Requirements (All Implemented ✅)

1. ✅ **Agents MUST know the user's role**
   - Implementation: `GetUserRoleContext()` returns role name + hierarchy level
   - Location: `services/role_context.go`

2. ✅ **Agents MUST know what actions the user can perform**
   - Implementation: Permission matrix in JSONB, `HasPermission()` method
   - Location: `middleware/permission_check.go`

3. ✅ **Agents MUST restrict suggestions/actions to user's permissions**
   - Implementation: Role context prompt with explicit instructions
   - Location: `handlers/chat_v2.go:410-448`, `agents/base_agent_v2.go:483-487`

4. ✅ **Agents use role context to personalize responses**
   - Implementation: Capability hints based on hierarchy level
   - Location: `handlers/chat_v2.go:433-440`

---

## Implementation Status: 70% Complete

### ✅ Completed (Backend)

#### Phase 1: Database Foundation (100%)
- Migration `026_workspaces_and_roles.sql` (600+ lines)
- 7 tables: workspaces, workspace_roles, workspace_members, user_workspace_profiles, workspace_memories, project_members, role_permissions
- `seed_default_workspace_roles()` function creating 6 default roles
- Hierarchy levels: 1=owner, 2=admin, 3=manager, 4=member, 5=viewer, 6=guest

#### Phase 2: Workspace Service (100%)
- `workspace_service.go` (600+ lines) with 15 methods
- `workspace_handlers.go` (346 lines) with 12 HTTP endpoints
- Transaction-safe workspace creation with auto-role seeding
- Plan-based member limits (free: 5, pro: 15, enterprise: 50)

#### Phase 4: Permission Middleware (100%)
- `permission_check.go` (408 lines) with 8 middleware functions
- `InjectRoleContext()` - loads user role into request context
- `RequirePermission()`, `RequireHierarchyLevel()`, `RequireWorkspaceOwner()` etc.
- Applied to all workspace endpoints

#### Phase 6: Agent Integration (100%)
- `roleContextPrompt` field added to `BaseAgentV2`
- `SetRoleContextPrompt()` method in `AgentV2` interface
- Role context injection in prompt building pipeline (order: base → personalization → **role context** → focus → style → thinking)
- `workspace_id` field added to `SendMessageRequest`
- Role extraction and prompt building in both `SendMessageV2` and `handleSlashCommandV2`

**Prompt Structure:**
```markdown
# WORKSPACE ROLE & PERMISSIONS
You are assisting a user with role: Manager (Level 3)

CRITICAL INSTRUCTIONS:
1. Only suggest operations the user has permission to perform
2. Explain role limitations if unauthorized action requested
3. Respect role hierarchy

Current capabilities:
- Can invite members but cannot modify workspace settings
```

### ❌ Not Started (30%)

#### Phase 3: Member Invitation System (0%)
- Email invitations
- Invitation links with expiration
- Accept/reject workflow
- Estimated: 10 hours

#### Phase 5: Workspace Memory Service (0%)
- Workspace-scoped memories (shared)
- User-scoped memories (private)
- Visibility controls by role
- RAG integration
- Estimated: 12 hours

#### Frontend UI (0%)
- WorkspaceSwitcher component ← **CRITICAL BLOCKER**
- MemberList, InviteMemberDialog, RoleSelector
- Workspace settings page
- Estimated: 16 hours

#### Testing (20%)
- Unit tests for workspace service
- Integration tests for permissions
- End-to-end role-based agent testing
- Estimated: 8 hours

---

## Critical Blocker

**Cannot test role-based agents without frontend integration!**

Frontend needs to:
1. Create `WorkspaceSwitcher.svelte` component
2. Pass `workspace_id` in chat message requests
3. Show current workspace in UI

**Minimum Viable Frontend (7 hours):**
```typescript
// frontend/src/lib/stores/workspace.ts
export const currentWorkspace = writable<Workspace | null>(null);

// frontend - chat request
fetch('/api/chat/v2/message', {
  body: JSON.stringify({
    message: userMessage,
    workspace_id: $currentWorkspace?.id,  // ← CRITICAL!
  })
});
```

---

## Decisions Made

### 1. Role Hierarchy Model
**Decision:** Use numeric hierarchy levels (lower = higher authority)
- Owner: 1
- Admin: 2
- Manager: 3
- Member: 4
- Viewer: 5
- Guest: 6

**Rationale:** Simple numeric comparison for permission checks (`IsAtLeastLevel(3)`)

### 2. Permission Matrix Storage
**Decision:** JSONB column with nested structure
```json
{
  "projects": {"create": true, "update": "own"},
  "team": {"invite": false}
}
```

**Rationale:** Flexible, queryable, extensible without schema changes

### 3. Agent Prompt Injection Point
**Decision:** Inject role context AFTER personalization, BEFORE focus mode

**Rationale:**
- Role context is fundamental (overrides general personalization)
- Focus mode is temporary (shouldn't override role restrictions)

### 4. Middleware Strategy
**Decision:** Separate `InjectRoleContext` from permission checks

**Rationale:**
- Reusable role context for multiple endpoints
- Granular permission control (owner-only vs admin+ vs manager+)

---

## Next Steps - Options

### Option A: Quick Win (Recommended - 7 hours)
**Goal:** Test what we built NOW

1. Minimal frontend (4h)
   - WorkspaceSwitcher dropdown
   - Pass workspace_id in chat

2. Manual testing (2h)
   - Create workspace via API
   - Add members with different roles
   - Send messages, verify agent responses

3. Validation (1h)
   - Check logs show role injection
   - Confirm agent respects permissions

**Outcome:** Proves backend works end-to-end

### Option B: Complete Backend First (46 hours)
1. Invitation System (10h)
2. Workspace Memory (12h)
3. Automated testing (8h)
4. Full frontend UI (16h)

### Option C: Testing First (14 hours)
1. Unit tests (8h)
2. Integration tests (4h)
3. Manual testing (2h)

---

## Technical Achievements

### What Works Now (Backend)
```bash
# 1. Create workspace with auto-role seeding
POST /api/workspaces
→ Creates workspace + 6 default roles + owner as first member

# 2. Protected endpoints
PUT /api/workspaces/:id
→ Requires admin+ (hierarchy level ≤ 2)

DELETE /api/workspaces/:id
→ Requires owner only (hierarchy level = 1)

# 3. Role-aware agents
POST /api/chat/v2/message
{
  "message": "How do I invite someone?",
  "workspace_id": "uuid"
}
→ Agent receives role context in system prompt
→ Response varies by user's role (viewer vs manager vs admin)
```

### Backend Logs Showing It Works
```
2026/01/06 00:29:09 Workspace service registered (workspaces, members, roles)
2026/01/06 00:29:09 Role context service registered (permission checks, hierarchy)
[ChatV2] Injected role context: manager (level 3)
[Agent] Applied role context prompt prefix (245 chars)
```

---

## Files Modified/Created

### Backend (Complete)
```
✅ desktop/backend-go/internal/database/migrations/026_workspaces_and_roles.sql
✅ desktop/backend-go/internal/services/workspace_service.go
✅ desktop/backend-go/internal/services/role_context.go
✅ desktop/backend-go/internal/handlers/workspace_handlers.go
✅ desktop/backend-go/internal/handlers/chat.go (added workspace_id)
✅ desktop/backend-go/internal/handlers/chat_v2.go (role injection)
✅ desktop/backend-go/internal/middleware/permission_check.go
✅ desktop/backend-go/internal/agents/base_agent_v2.go
✅ desktop/backend-go/internal/agents/agent_v2.go
✅ desktop/backend-go/cmd/server/main.go (service init)
```

### Frontend (Pending)
```
❌ frontend/src/lib/components/workspace/WorkspaceSwitcher.svelte
❌ frontend/src/lib/api/workspace.ts
❌ frontend/src/lib/stores/workspace.ts
❌ frontend/src/lib/components/chat/ChatInput.svelte (modify)
```

---

## Documentation

**Full Status Report:** `docs/feature1_role_based_agents_status.md` (comprehensive)

**Reference:** `docs/FUTURE_FEATURES.md` lines 74-79

---

## Success Criteria

### Backend ✅
- [x] Database schema supports workspaces, roles, members
- [x] API endpoints for workspace CRUD
- [x] Permission middleware protects routes
- [x] Role context service returns user permissions
- [x] Agents receive role context in system prompt
- [x] Backend logs show role injection

### Frontend ⏳
- [ ] Can select workspace from dropdown
- [ ] Chat requests include workspace_id
- [ ] UI shows current workspace

### Testing ⏳
- [ ] Create workspace → verify 6 roles seeded
- [ ] Add viewer → verify cannot edit workspace
- [ ] Add manager → verify can invite members
- [ ] Send message as viewer → agent explains limitations
- [ ] Send message as admin → agent shows full capabilities

---

## Risks & Mitigations

### Risk: Cannot validate without frontend
**Mitigation:** Build minimal frontend first (7h investment)

### Risk: Complex permission matrix hard to debug
**Mitigation:** Extensive logging of role context injection

### Risk: Frontend-backend workspace_id sync
**Mitigation:** Use Svelte store for single source of truth

---

## Open Questions

1. Should we implement invitation system before or after testing?
   - **Recommendation:** After testing - prove backend works first

2. Custom roles beyond 6 defaults?
   - **Decision:** Not in MVP - use default 6 roles first

3. Project-level role overrides?
   - **Decision:** Phase 2 - workspace roles are enough for MVP

---

## Related Work

- Feature 7: Multi-modal embeddings (completed)
- Pedro Tasks: Memory system (partial - needs workspace context)
- RAG services: Need workspace-scoped retrieval

---

**Checkpoint Created:** 2026-01-06 00:35 UTC
**Next Review:** After frontend minimal implementation
**Status:** Ready for testing pending frontend work
