# Role Context Prompt Enhancement

## Problem
Backend logs showed role context WAS being injected successfully:
```
[ChatV2] Injected role context: owner (level 1, 6 permissions)
```

But the agent didn't mention the user's role in responses because:
1. The system prompt was 24KB long
2. Role context was buried in the middle of the prompt (after personalization, before memory context)
3. The formatting wasn't prominent enough to catch the model's attention

## Solution

### 1. Enhanced Prompt Formatting (`role_context.go`)

The role context prompt now includes:
- **Visual separators** using box-drawing characters (═══) and emojis (🔐, 🎯)
- **CRITICAL markers** to highlight importance
- **Explicit mandatory behavior** instructions
- **Clear examples** of how to acknowledge the role

Example output:
```
═══════════════════════════════════════════════════════════════════════════════
🔐 CRITICAL: USER ROLE & PERMISSIONS CONTEXT
═══════════════════════════════════════════════════════════════════════════════

You are assisting a user with the following role and permissions. This information is
CRITICAL and MUST be acknowledged in your responses when relevant.

**User:** user-123
**Workspace Role:** Owner (owner)
**Authority Level:** 1 (lower = higher authority)
**Title:** CEO
**Department:** Executive

╔═══════════════════════════════════════════════════════════════════════════════╗
║ PERMISSIONS GRANTED TO THIS USER                                             ║
╚═══════════════════════════════════════════════════════════════════════════════╝
- **projects**: create, read, update, delete
- **members**: invite, remove, manage
- **settings**: modify
- **billing**: manage
- **roles**: create, update, delete

╔═══════════════════════════════════════════════════════════════════════════════╗
║ ACTIONS RESTRICTED FROM THIS USER                                            ║
╚═══════════════════════════════════════════════════════════════════════════════╝
- None (full workspace access)

═══════════════════════════════════════════════════════════════════════════════
🎯 MANDATORY BEHAVIOR:
═══════════════════════════════════════════════════════════════════════════════
1. When the user asks "what can I do?" or similar questions, IMMEDIATELY reference
   their role (Owner) and explain their specific permissions listed above.

2. ALWAYS acknowledge their role when providing workspace-related guidance.
   Example: "As the Owner of this workspace, you have..."

3. ONLY suggest actions that are within their permission set listed above.

4. If they request something outside their permissions, politely explain:
   "I see you'd like to [action], but this requires [permission/role].
    Your current role (Owner) doesn't include this permission."

5. Tailor technical depth and business context to their title (CEO) and
   department (Executive).

═══════════════════════════════════════════════════════════════════════════════
```

### 2. Prompt Ordering (`base_agent_v2.go`)

Changed the prompt construction order to place role context at the VERY BEGINNING:

**Before:**
1. Base system prompt
2. Personalization
3. **Role context** ← buried here
4. Memory context
5. Focus mode
6. Output style
7. Thinking instructions

**After:**
1. **Role context** ← NOW AT THE TOP
2. Focus mode
3. Output style
4. Memory context
5. Base system prompt
6. Personalization (enhances but doesn't override role context)
7. Thinking instructions

### 3. Code Changes

#### `role_context.go` - Lines 148-208
- Enhanced `GetRoleContextPrompt()` with:
  - Box-drawing characters for visual prominence
  - Emojis (🔐, 🎯) to catch attention
  - "CRITICAL" and "MANDATORY" keywords
  - Explicit behavioral instructions
  - Concrete examples

#### `base_agent_v2.go` - Lines 468-543
- Refactored `buildSystemPromptWithThinking()` to:
  - Start with role context (if present)
  - Add other contexts in priority order
  - Keep personalization after base prompt
  - Add detailed logging with ✓ marker for role context placement

## Testing

### Compilation
```bash
cd desktop/backend-go
go build ./cmd/server
# ✓ Builds successfully
```

### Expected Behavior

**User asks:** "What can I do in this workspace?"

**Before this fix:**
```
Agent: "You can create projects, manage tasks, invite team members..."
[Generic response, no role acknowledgment]
```

**After this fix:**
```
Agent: "As the Owner of this workspace, you have full permissions including:
- Create, read, update, and delete projects
- Invite and remove team members
- Manage workspace settings and billing
- Create and modify roles

Your Owner role (authority level 1) gives you complete control over all workspace resources."
[Explicitly acknowledges role and lists specific permissions]
```

## Files Modified

1. `desktop/backend-go/internal/services/role_context.go`
   - Enhanced `GetRoleContextPrompt()` function (lines 148-208)

2. `desktop/backend-go/internal/agents/base_agent_v2.go`
   - Refactored `buildSystemPromptWithThinking()` function (lines 468-543)

3. `desktop/backend-go/internal/services/role_context_test.go`
   - Updated test assertions to verify new prompt format (lines 203-244)

## Why This Works

1. **Primacy Effect**: Information at the beginning of a prompt has more influence on the model's behavior
2. **Visual Prominence**: Box-drawing characters and emojis create visual anchors that LLMs recognize
3. **Explicit Instructions**: "MANDATORY BEHAVIOR" section tells the model exactly what to do
4. **Examples**: Concrete examples like "As the Owner of this workspace, you have..." give the model a template to follow
5. **Keyword Markers**: "CRITICAL" and "MUST" signal importance to the attention mechanism

## Logs to Watch

When a workspace_id is provided, you'll now see:
```
[ChatV2] Injected role context: owner (level 1, 6 permissions)
[Agent] ✓ ROLE CONTEXT placed at START of prompt (1247 chars)
```

The ✓ marker confirms role context is at the beginning of the system prompt.
