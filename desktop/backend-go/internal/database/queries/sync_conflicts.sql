-- name: CreateSyncConflict :one
INSERT INTO sync_conflicts (
    entity_type,
    entity_id,
    local_data,
    remote_data,
    local_updated_at,
    remote_updated_at,
    conflict_fields,
    detected_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW()
) RETURNING *;

-- name: GetSyncConflict :one
SELECT * FROM sync_conflicts
WHERE id = $1;

-- name: GetUnresolvedConflicts :many
SELECT * FROM sync_conflicts
WHERE resolved_at IS NULL
ORDER BY detected_at DESC
LIMIT $1 OFFSET $2;

-- name: GetUnresolvedConflictsByEntity :many
SELECT * FROM sync_conflicts
WHERE entity_type = $1
  AND entity_id = $2
  AND resolved_at IS NULL
ORDER BY detected_at DESC;

-- name: GetConflictsByEntity :many
SELECT * FROM sync_conflicts
WHERE entity_type = $1
  AND entity_id = $2
ORDER BY detected_at DESC
LIMIT $1 OFFSET $2;

-- name: ResolveSyncConflictByID :one
UPDATE sync_conflicts
SET
    resolution_strategy = $2,
    resolved_data = $3,
    resolved_by = $4,
    resolved_at = NOW(),
    reasoning = $5
WHERE id = $1
RETURNING *;

-- name: DeleteOldResolvedConflicts :exec
DELETE FROM sync_conflicts
WHERE resolved_at IS NOT NULL
  AND resolved_at < NOW() - INTERVAL '30 days';

-- name: CountUnresolvedConflicts :one
SELECT COUNT(*) FROM sync_conflicts
WHERE resolved_at IS NULL;

-- name: CountUnresolvedConflictsByEntity :one
SELECT COUNT(*) FROM sync_conflicts
WHERE entity_type = $1
  AND entity_id = $2
  AND resolved_at IS NULL;
