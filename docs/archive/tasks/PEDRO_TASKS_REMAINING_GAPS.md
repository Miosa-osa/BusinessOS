# Pedro Tasks V2 - Remaining Gaps & Follow-up Tasks

**Date:** January 3, 2026
**Status:** Post-merge audit
**Branch:** main-dev (merged from pedro-dev)

---

## Summary

After merging Pedro's `pedro-dev` branch and running all migrations, the following gaps have been identified for follow-up work.

**Overall Assessment:** ~90% complete, remaining items are polish and integration work.

---

## Verified Complete Items

These items were flagged during audit but verified complete after migration review:

| Item | Status | Evidence |
|------|--------|----------|
| `context_retrieval_log` table | COMPLETE | Found in migration 020 |
| `conversation_summaries` table | COMPLETE | Found in migration 020 |
| Pre-seeded output styles (8) | COMPLETE | Verified in DB: conversational, professional, technical, executive, detailed, creative, tutorial, qa |
| Memory system | COMPLETE | 15 endpoints registered, frontend components exist |
| Context management | COMPLETE | 8 endpoints registered, TreeSearchPanel exists |
| Block system | COMPLETE | block_mapper.go (954 lines) |
| Document processing | COMPLETE | document_processor.go with PDF/DOCX extraction |
| Learning system | COMPLETE | 4 tables created, service implemented |
| App profiler | COMPLETE | Backend service + API endpoints |

---

## Remaining Gaps (Follow-up Tasks)

### Priority 1: Integration Tasks

#### 1.1 Context Tree Visualization Component
**Status:** Partial
**Issue:** TreeSearchPanel.svelte exists but is search-focused. Missing hierarchical tree view for browsing.
**Location:** `frontend/src/lib/components/contexts/`
**Task:** Create `ContextTreeView.svelte` component that shows full hierarchical tree with:
- Node > Project > Context Profile > Items structure
- Expand/collapse functionality
- Visual icons by item type
- Token count display

#### 1.2 Automatic Learning Triggers
**Status:** Partial
**Issue:** Learning service is manual. Auto-triggers not wired into conversation flow.
**Location:** `desktop/backend-go/internal/handlers/chat_v2.go`
**Task:** Wire up learning triggers:
- Call `LearningService.LearnFromConversation()` after each conversation turn
- Trigger pattern detection periodically
- Auto-extract memories from conversations

#### 1.3 Orchestrator V2 System Prompt
**Status:** Needs Verification
**Issue:** Enhanced orchestrator prompt from spec may not be implemented
**Location:** `desktop/backend-go/internal/prompts/agents/orchestrator.go`
**Task:** Verify or implement:
- Context-aware prompt with user profile injection
- Memory and fact injection
- Output style application
- Tool descriptions for tree_search, load_context

### Priority 2: Frontend Polish

#### 2.1 Application Profiles UI
**Status:** API Only
**Issue:** Backend complete but no Settings UI for app profiles
**Location:** `frontend/src/routes/(app)/settings/`
**Task:** Create settings page section for:
- Listing application profiles
- Creating new profile from directory/repo
- Viewing components, modules, endpoints
- Syncing profile

#### 2.2 Memories Integration in Contexts Page
**Status:** Partial
**Issue:** MemoryPanel exists for chat but not integrated into Contexts page
**Task:** Add memories tab to Contexts page showing:
- All memories
- Filter by project/node
- Memory statistics

### Priority 3: Schema Verification

#### 3.1 Voice Notes Context Linking
**Status:** Needs Verification
**Issue:** voice_notes table may need project_id, node_id columns
**Location:** `desktop/backend-go/internal/database/migrations/`
**Task:** Verify these columns exist or create migration:
```sql
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id);
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS node_id UUID REFERENCES nodes(id);
ALTER TABLE voice_notes ADD COLUMN IF NOT EXISTS is_context_source BOOLEAN DEFAULT TRUE;
```

---

## Files Cleaned Up (Not in Repo)

Removed during merge cleanup:
- `desktop/backend-go/server.exe` (57MB binary)
- `desktop/backend-go/server_err.txt` (error logs)
- `*.ps1` files (Windows-specific scripts)
- `cookies.txt`
- `check_*.txt` files
- `frontend/check_output.txt`
- `presets.json`

---

## Testing Checklist

Before considering Pedro Tasks V2 fully complete:

- [ ] Memory CRUD works through frontend
- [ ] Memory search returns relevant results
- [ ] User facts panel works in Settings
- [ ] Output style selector works
- [ ] Document upload and processing works
- [ ] Context tree search works
- [ ] Learning feedback is recorded
- [ ] Conversation summaries are generated

---

## Next Steps

1. **Immediate:** Test frontend with backend running
2. **Short-term:** Implement Priority 1 integration tasks
3. **Medium-term:** Add Priority 2 frontend polish
4. **Ongoing:** Verify Priority 3 schema items

---

*This document should be updated as gaps are addressed.*
