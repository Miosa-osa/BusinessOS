# Conflict Detector Implementation

## Overview

The Conflict Detector implements a **3-tier resolution strategy** for bidirectional sync conflicts between BusinessOS and OSA-5.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              CONFLICT DETECTION & RESOLUTION                │
└─────────────────────────────────────────────────────────────┘

1. DETECT CONFLICT
   ├── Compare timestamps (>5 sec diff = no conflict)
   ├── Compare field values (identical = no conflict)
   └── Identify conflicting fields

2. RESOLVE CONFLICT (3-Tier Strategy)
   │
   ├── TIER 1: Timestamp-based (>5 second gap)
   │   └── Auto-resolve: Use most recent version
   │
   ├── TIER 2: Field-level merge (non-critical fields)
   │   └── Auto-merge: Combine independent changes
   │
   └── TIER 3: Manual review (critical fields)
       └── Queue for user: Show conflict UI

3. STORE RESULT
   └── Save to sync_conflicts table with resolution strategy
```

## 3-Tier Resolution Strategy

### Tier 1: Timestamp-based Resolution

**When:** Time difference > 5 seconds

**Action:** Automatically use the most recent version

**Example:**
```
Local:  updated_at = 2026-01-09T10:00:00Z, name = "Old Name"
Remote: updated_at = 2026-01-09T10:00:10Z, name = "New Name"

Resolution: Use Remote (10 seconds newer)
Strategy: "timestamp_based"
```

### Tier 2: Field-level Merge

**When:** Only non-critical fields conflict AND time difference ≤ 5 seconds

**Action:** Automatically merge changes

**Non-critical fields:**
- `layout` (JSONB) - UI layout configuration
- `active_modules` (UUID[]) - Active module list
- `settings` (JSONB) - Workspace settings

**Critical fields:**
- `name` (string) - Workspace name
- `mode` (string) - Display mode (2d/3d)

**Example:**
```
Local:  layout = {"x": 10}, settings = {"theme": "dark"}
Remote: layout = {"y": 20}, settings = {"lang": "en"}

Resolution: Merge both
  layout = {"x": 10, "y": 20}
  settings = {"theme": "dark", "lang": "en"}

Strategy: "field_level_merge"
```

### Tier 3: Manual Review

**When:** Critical fields conflict AND time difference ≤ 5 seconds

**Action:** Queue for user review

**Example:**
```
Local:  name = "My Workspace", mode = "2d"
Remote: name = "Team Workspace", mode = "3d"

Resolution: NULL (requires user decision)
Strategy: "manual_review"
Status: Queued in sync_conflicts table
```

## Conflict Detection Fields

For `workspace` entities, conflicts are detected in:

| Field | Type | Critical | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ Yes | Workspace display name |
| `mode` | string | ✅ Yes | Display mode (2d/3d/hybrid) |
| `layout` | JSONB | ❌ No | UI layout positions |
| `active_modules` | UUID[] | ❌ No | Active module list |
| `settings` | JSONB | ❌ No | User preferences |

## Database Schema

```sql
CREATE TABLE sync_conflicts (
    id UUID PRIMARY KEY,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,

    -- Conflict data
    local_data JSONB NOT NULL,
    remote_data JSONB NOT NULL,
    local_updated_at TIMESTAMPTZ NOT NULL,
    remote_updated_at TIMESTAMPTZ NOT NULL,
    conflict_fields TEXT[] NOT NULL,

    -- Resolution
    resolution_strategy VARCHAR(50),
    resolved_data JSONB,
    resolved_by UUID REFERENCES users(id),  -- NULL = automatic
    resolved_at TIMESTAMPTZ,
    reasoning TEXT,

    -- Metadata
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_resolution_strategy CHECK (
        resolution_strategy IN (
            'timestamp_based',
            'field_level_merge',
            'manual_review'
        )
    )
);
```

## API Usage

### 1. Detect Conflict

```go
detector := sync.NewConflictDetector(pool, logger)

local := &sync.Workspace{
    ID:        workspaceID,
    Name:      "Local Name",
    Mode:      "2d",
    UpdatedAt: time.Now(),
}

remote := &sync.Workspace{
    ID:        workspaceID,
    Name:      "Remote Name",
    Mode:      "3d",
    UpdatedAt: time.Now().Add(2 * time.Second),
}

conflict, err := detector.DetectWorkspaceConflict(ctx, local, remote)
if err != nil {
    return err
}

if conflict == nil {
    // No conflict - safe to proceed
    return nil
}

// Conflict detected
log.Printf("Conflict detected: %v", conflict.ConflictFields)
```

### 2. Resolve Conflict

```go
resolution, err := detector.ResolveConflict(ctx, conflict)
if err != nil {
    return err
}

switch resolution.Strategy {
case sync.ResolutionTimestampBased:
    // Auto-resolved by timestamp
    log.Printf("Auto-resolved: %s", resolution.Reasoning)
    // Apply resolved_data to database

case sync.ResolutionFieldLevelMerge:
    // Auto-merged fields
    log.Printf("Auto-merged: %s", resolution.Reasoning)
    // Apply resolved_data to database

case sync.ResolutionManualReview:
    // Requires user intervention
    log.Printf("Manual review required: %s", resolution.Reasoning)
    // Store conflict and notify user
    err = detector.StoreConflict(ctx, conflict)
}
```

### 3. Query Unresolved Conflicts

```go
// Get all unresolved conflicts
conflicts, err := detector.GetUnresolvedConflicts(ctx, 10, 0)

// Get conflicts for specific workspace
conflicts, err := detector.GetUnresolvedConflictsByEntity(
    ctx,
    "workspace",
    workspaceID,
)
```

## Integration with Sync Service

```go
// In workspace sync flow
func (s *SyncService) SyncWorkspace(ctx context.Context, workspaceID uuid.UUID) error {
    // 1. Fetch local workspace
    local, err := s.db.GetOSAWorkspace(ctx, workspaceID)
    if err != nil {
        return err
    }

    // 2. Fetch remote workspace
    remote, err := s.osaClient.GetWorkspace(ctx, workspaceID)
    if err != nil {
        return err
    }

    // 3. Detect conflicts
    conflict, err := s.conflictDetector.DetectWorkspaceConflict(ctx, local, remote)
    if err != nil {
        return err
    }

    // 4. Handle conflict
    if conflict != nil {
        resolution, err := s.conflictDetector.ResolveConflict(ctx, conflict)
        if err != nil {
            return err
        }

        if resolution.Strategy == sync.ResolutionManualReview {
            // Store for user review
            return s.conflictDetector.StoreConflict(ctx, conflict)
        }

        // Apply automatic resolution
        var resolved sync.Workspace
        json.Unmarshal(resolution.ResolvedData, &resolved)
        return s.applyWorkspace(ctx, &resolved)
    }

    // 5. No conflict - proceed normally
    return s.applyWorkspace(ctx, remote)
}
```

## Testing

Run the test suite:

```bash
go test ./internal/sync -v -run TestConflict
```

### Test Coverage

- ✅ No conflict: Remote clearly newer (>5 sec)
- ✅ No conflict: Local clearly newer (>5 sec)
- ✅ No conflict: Identical data
- ✅ Conflict: Name changed (concurrent)
- ✅ Conflict: Multiple fields changed
- ✅ Resolution: Timestamp-based (remote wins)
- ✅ Resolution: Timestamp-based (local wins)
- ✅ Resolution: Field-level merge (non-critical)
- ✅ Resolution: Manual review (critical fields)
- ✅ Helper functions: JSON merge, UUID slice merge

## Monitoring

### Metrics

Track these metrics in Prometheus:

```go
// Conflict detection rate
conflict_detected_total{entity_type="workspace"}

// Resolution strategy usage
conflict_resolution_total{strategy="timestamp_based"}
conflict_resolution_total{strategy="field_level_merge"}
conflict_resolution_total{strategy="manual_review"}

// Unresolved conflicts (gauge)
conflict_unresolved_count{entity_type="workspace"}
```

### Logging

```go
// Conflict detected
logger.Warn("conflict detected",
    "workspace_id", workspaceID,
    "fields", conflictFields,
    "time_diff", timeDiff)

// Auto-resolved
logger.Info("conflict resolved",
    "workspace_id", workspaceID,
    "strategy", strategy,
    "reasoning", reasoning)

// Manual review queued
logger.Warn("conflict requires manual review",
    "workspace_id", workspaceID,
    "fields", conflictFields)
```

## Future Enhancements

1. **Conflict Resolution UI**
   - Show side-by-side diff
   - Allow user to choose local/remote/custom
   - Bulk resolve multiple conflicts

2. **Advanced Merge Strategies**
   - Operational Transform for layout
   - CRDT for active_modules list
   - Schema-aware JSON merging

3. **Conflict Prevention**
   - Optimistic locking with version numbers
   - Lock workspaces during sync
   - Real-time conflict detection via WebSockets

4. **Analytics**
   - Conflict frequency by field
   - Resolution success rate
   - Average time to manual resolution

## References

- [OSA Phase 3 Sync Design](../../../../docs/architecture/OSA_PHASE3_SYNC_DESIGN.md)
- [Sync Specification Q5: Conflict Resolution](../../../../docs/architecture/SYNC_SPECIFICATION_ANSWERS.md#q5-conflict-resolution-strategy)
- [Bidirectional Sync Best Practices](../../../../docs/architecture/BIDIRECTIONAL_SYNC_BEST_PRACTICES.md)
