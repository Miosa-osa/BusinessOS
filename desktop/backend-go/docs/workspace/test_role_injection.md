# Role Context Injection Test Plan

## What Was Implemented

### Changes Made

1. **chat_v2.go** (lines 410-425):
   - Role context injection in main chat handler
   - Uses `GetRoleContextPrompt()` from service instead of manual prompt building
   - Logs role name, hierarchy level, and permission count

2. **chat_v2.go** (lines 1414-1427):
   - Role context injection in slash command handler
   - Same pattern as main handler
   - Consistent logging

3. **base_agent_v2.go** (lines 32, 159-161, 483-487):
   - `roleContextPrompt` field added to BaseAgentV2
   - `SetRoleContextPrompt()` method implemented
   - Role prompt prepended to system prompt in `buildSystemPromptWithThinking()`

4. **agent_v2.go** (line 46):
   - `SetRoleContextPrompt()` method added to AgentV2 interface

5. **role_context.go**:
   - `GetRoleContextPrompt()` method already exists (lines 149-179)
   - Formats role information with permissions and restrictions

## How It Works

### Flow

1. **Request received** with `workspace_id` parameter
2. **Handler checks** if `roleContextService` is available
3. **Service queries** database for user's role in workspace:
   - workspace_members table (role name, hierarchy level)
   - role_permissions table (permissions by resource)
   - user_profiles table (title, department)
   - user_facts table (expertise areas)
   - project_members table (project-specific roles)
4. **Service builds** role context prompt with:
   - User role and hierarchy level
   - What user CAN do (permissions list)
   - What user CANNOT do (restrictions based on role)
   - Expertise context
5. **Handler calls** `agent.SetRoleContextPrompt(rolePrompt)`
6. **Agent prepends** role prompt to system prompt in `buildSystemPromptWithThinking()`
7. **LLM receives** complete context and respects role boundaries

### Prompt Structure

```
## User Role Context

**User:** <user_id>
**Role:** <display_name> (<role_name>)
**Hierarchy Level:** <level>
**Title:** <title>
**Department:** <department>

### What This User Can Do:
- **projects**: create, read, update
- **tasks**: create, read, update, delete
- **members**: read

### What This User CANNOT Do:
- Delete workspace (only owner can)
- Manage workspace billing
- Modify role permissions

### Important:
- Only suggest actions within this user's permissions
- Do not offer to perform restricted actions
- If user asks for something outside their role, explain they need appropriate permissions
- Tailor responses to their expertise level and department

<original system prompt>
```

## Testing

### Test Case 1: Viewer Role
**Setup:**
- User with "viewer" role (hierarchy level 5)
- No edit permissions

**Request:**
```json
{
  "message": "Help me delete all tasks",
  "workspace_id": "<workspace_uuid>"
}
```

**Expected Behavior:**
- Agent should explain user has view-only access
- Should NOT suggest delete operations
- Should offer read-only alternatives

### Test Case 2: Manager Role
**Setup:**
- User with "manager" role (hierarchy level 3)
- Can edit projects, invite members

**Request:**
```json
{
  "message": "Create a new project plan",
  "workspace_id": "<workspace_uuid>"
}
```

**Expected Behavior:**
- Agent should create project
- Should offer to invite team members
- Should respect manager capabilities

### Test Case 3: No Workspace Context
**Setup:**
- Request without workspace_id

**Request:**
```json
{
  "message": "Help me with my project"
}
```

**Expected Behavior:**
- No role context injected
- Agent behaves with default permissions
- No role-based restrictions applied

## Verification Steps

1. **Check Logs:**
   ```
   [ChatV2] Injected role context: manager (level 3, 5 permissions)
   ```

2. **Test API:**
   ```bash
   curl -X POST http://localhost:8080/api/chat/v2 \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{
       "message": "Create a new project",
       "workspace_id": "<workspace_uuid>"
     }'
   ```

3. **Verify Agent Response:**
   - Check if agent respects role boundaries
   - Verify agent doesn't suggest forbidden actions
   - Confirm permission-appropriate suggestions

## Database Setup Required

For testing, ensure these tables exist:
- `workspace_members` (user-workspace role mapping)
- `workspace_roles` (role definitions)
- `role_permissions` (role-resource permissions)
- `user_profiles` (user metadata)
- `project_members` (project-level roles)

## Success Criteria

- [x] Code compiles without errors
- [x] Role context is retrieved from service
- [x] GetRoleContextPrompt() generates formatted prompt
- [x] Agent receives role context in system prompt
- [x] Role prompt is prepended before other customizations
- [ ] Agent respects role boundaries in responses (requires live testing)
- [ ] Logging shows role injection details

## Notes

- Role context only applies when `workspace_id` is provided
- Falls back gracefully if service or data unavailable
- Order of prompt composition: role → focus → output style → personalization → thinking
- Consistent implementation in both regular and slash command handlers
