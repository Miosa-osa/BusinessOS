# P2: Sync Engine (Offline Support)

> **Priority:** P2 - Nice to Have
> **Backend Status:** Complete (13 endpoints)
> **Frontend Status:** Not Started
> **Estimated Effort:** 2-3 sprints

---

## Overview

The sync engine enables offline-first functionality with background synchronization. Users can work offline and changes sync when back online.

---

## Backend API Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/sync/status` | Get sync status |
| GET | `/api/sync/full` | Full sync all data |
| GET | `/api/sync/:table` | Get table sync changes |

### Per-Table Sync
| Endpoint | Table |
|----------|-------|
| `/api/sync/contexts/sync` | Contexts |
| `/api/sync/conversations/sync` | Conversations |
| `/api/sync/projects/sync` | Projects |
| `/api/sync/tasks/sync` | Tasks |
| `/api/sync/nodes/sync` | Nodes |
| `/api/sync/clients/sync` | Clients |
| `/api/sync/calendar_events/sync` | Calendar |
| `/api/sync/daily_logs/sync` | Daily Logs |
| `/api/sync/team_members/sync` | Team Members |
| `/api/sync/artifacts/sync` | Artifacts |
| `/api/sync/focus_items/sync` | Focus Items |
| `/api/sync/user_settings/sync` | Settings |

---

## Frontend Implementation Tasks

### Phase 1: Local Storage Layer
- [ ] IndexedDB setup (Dexie.js)
- [ ] Table schemas matching backend
- [ ] CRUD operations on local DB

### Phase 2: Sync Logic
- [ ] Change tracking (dirty flags)
- [ ] Conflict detection
- [ ] Conflict resolution UI
- [ ] Background sync worker

### Phase 3: Offline UI
- [ ] Offline indicator
- [ ] Pending changes badge
- [ ] Sync status in settings
- [ ] Manual sync button

---

## Linear Issues to Create

1. **[SYNC-001]** Setup IndexedDB with Dexie
2. **[SYNC-002]** Implement change tracking
3. **[SYNC-003]** Build sync worker
4. **[SYNC-004]** Add conflict resolution
5. **[SYNC-005]** Create offline UI indicators

---

## Notes

- This is a significant architectural change
- Consider starting with read-only offline (cache)
- Full offline-first requires careful conflict handling
- Service worker for background sync
