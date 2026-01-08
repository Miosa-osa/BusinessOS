# Changelog - December 28, 2025

## Overview

This update includes significant improvements to the artifact system, AI model routing, and settings persistence.

---

## Features Added

### 1. Artifact Version History

**Files Modified:**
- `frontend/src/routes/(app)/chat/+page.svelte`
- `frontend/src/lib/api/artifacts/artifacts.ts`
- `desktop/backend-go/internal/handlers/artifacts.go`

**Description:**
Users can now view and restore previous versions of artifacts. Each time an artifact is updated, the previous content is automatically saved as a version.

**New UI Elements:**
- History button (clock icon) in artifact detail panel
- Version history dropdown showing all previous versions
- Restore button for each version with confirmation
- Timestamp display for each version

**API Endpoints:**
- `GET /artifacts/:id/versions` - Get all versions of an artifact
- `POST /artifacts/:id/restore` - Restore a specific version

---

### 2. Sync Artifacts to Knowledge Base

**Files Modified:**
- `frontend/src/routes/(app)/chat/+page.svelte`
- `frontend/src/lib/api/artifacts/artifacts.ts`
- `desktop/backend-go/internal/handlers/artifacts.go`
- `desktop/backend-go/internal/database/queries/contexts.sql`
- `desktop/backend-go/internal/database/sqlc/contexts.sql.go`

**Description:**
Artifacts can now be synced directly to Knowledge Base contexts. This allows users to save important generated content (documents, code, etc.) to their knowledge base for future reference.

**New UI Elements:**
- "Sync to KB" button in artifact detail panel
- Context selection dropdown
- Loading and success states

**How it works:**
1. Click "Sync to KB" button on an artifact
2. Select a context from the dropdown
3. Artifact content is appended to the context with a header

---

## Bug Fixes

### 3. Token Counter Now Reflects Saved Settings

**Files Modified:**
- `frontend/src/routes/(app)/chat/+page.svelte` (lines 897-900)

**Problem:**
The token counter in chat showed hardcoded context limits (e.g., "88 / 41K") instead of the user's saved context window setting from AI Settings (e.g., 131K).

**Solution:**
- Added loading of `contextWindow` from user settings in `loadUserSettings()`
- `currentContextLimit` now checks `aiContextWindow` first before falling back to model defaults

**Code Change:**
```typescript
// Load context window setting
if (typeof settings.model_settings.contextWindow === 'number' && settings.model_settings.contextWindow > 0) {
    aiContextWindow = settings.model_settings.contextWindow;
}
```

---

### 4. Model Settings Persistence Fix

**Files Modified:**
- `frontend/src/routes/(app)/settings/ai/+page.svelte` (lines 483-527)

**Problem:**
AI Settings page wasn't loading saved model settings on mount, so users would see default values instead of their saved preferences.

**Solution:**
- `loadConfig()` now fetches saved `model_settings` from `/settings` API
- Properly populates temperature, maxTokens, contextWindow, topP, streamResponses, showUsageInChat

---

### 5. Groq Model Selection and API Routing Fix

**Files Modified:**
- `desktop/backend-go/internal/services/llm.go` (lines 82-154)

**Problem:**
When users selected Groq models (like `moonshotai/kimi-k2-instruct`), the backend would still try to route to Ollama, causing errors like:
```
failed to send request: Post https://api.ollama.com/v1/chat/completion... host
```

**Solution:**
Added intelligent provider inference based on model name:

```go
// InferProviderFromModel determines the appropriate provider based on model name
func InferProviderFromModel(model string) string {
    // Anthropic/Claude models
    if strings.HasPrefix(lowerModel, "claude") {
        return "anthropic"
    }

    // Groq models
    groqModels := []string{
        "llama-3.3-70b", "llama-3.1-70b", "llama-3.1-8b",
        "llama3-70b", "llama3-8b", "mixtral-8x7b", "gemma2-9b-it",
    }

    // OpenRouter-style models (provider/model format) -> groq
    if strings.Contains(model, "/") {
        return "groq"
    }

    // Models with -cloud suffix -> ollama_cloud
    if strings.HasSuffix(lowerModel, "-cloud") {
        return "ollama_cloud"
    }

    return "" // Use global config
}
```

**Provider Detection Logic:**
| Model Pattern | Detected Provider |
|--------------|-------------------|
| `claude*` | anthropic |
| `llama-3.x-*`, `mixtral-*`, `gemma2-*` | groq |
| `provider/model` (e.g., `moonshotai/kimi-k2-instruct`) | groq |
| `*-cloud` | ollama_cloud |
| Other | Falls back to global config |

---

## Technical Details

### State Variables Added (chat/+page.svelte)

```typescript
// Artifact version history
let showVersionHistory = $state(false);
let artifactVersions = $state<ArtifactVersion[]>([]);
let loadingVersions = $state(false);
let restoringVersion = $state(false);

// Sync to KB state
let showSyncToKBDropdown = $state(false);
let availableContextsForSync = $state<ContextListItem[]>([]);
let loadingContextsForSync = $state(false);
let syncingToKB = $state(false);

// AI context window from settings
let aiContextWindow = $state(0);
```

### New Functions Added (chat/+page.svelte)

- `loadVersionHistory(artifactId)` - Fetches version history from API
- `restoreVersion(version)` - Restores artifact to a previous version
- `formatVersionDate(dateStr)` - Formats version timestamps
- `loadContextsForSync()` - Loads available KB contexts
- `syncArtifactToContext(contextId)` - Syncs artifact to selected context

### API Functions Added (artifacts.ts)

```typescript
export async function getArtifactVersions(id: string) {
  return request<ArtifactVersion[]>(`/artifacts/${id}/versions`);
}

export async function restoreArtifactVersion(id: string, version: number) {
  return request<Artifact>(`/artifacts/${id}/restore`, { method: 'POST', body: { version } });
}

// Updated linkArtifact to support sync_to_kb
export async function linkArtifact(id: string, data: { project_id?: string; context_id?: string; sync_to_kb?: boolean }) {
  return request<Artifact>(`/artifacts/${id}/link`, { method: 'PATCH', body: data });
}
```

---

## Testing Notes

1. **Artifact Versioning**: Create an artifact, edit it multiple times, then click History to see all versions
2. **Sync to KB**: Create a context first, then sync an artifact to it
3. **Token Counter**: Set context window in AI Settings, verify chat shows correct limit
4. **Groq Routing**: Select a Groq model (or any `provider/model` format), send a message - should work without Ollama errors

---

## Files Changed Summary

| File | Changes |
|------|---------|
| `frontend/src/routes/(app)/chat/+page.svelte` | Version history UI, KB sync UI, settings loading |
| `frontend/src/routes/(app)/settings/ai/+page.svelte` | Load saved model_settings on mount |
| `frontend/src/lib/api/artifacts/artifacts.ts` | Version history + restore API functions |
| `desktop/backend-go/internal/services/llm.go` | Provider inference from model name |
| `desktop/backend-go/internal/handlers/artifacts.go` | KB sync in LinkArtifact handler |
| `desktop/backend-go/internal/database/queries/contexts.sql` | SyncArtifactToContext query |
| `desktop/backend-go/internal/database/sqlc/contexts.sql.go` | SyncArtifactToContext function |
