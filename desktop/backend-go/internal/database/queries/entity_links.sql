-- ============================================================================
-- ENTITY LINKS QUERIES
-- ============================================================================

-- name: ListEntityOutgoingLinks :many
SELECT * FROM entity_links
WHERE source_type = $1 AND source_id = $2
ORDER BY created_at DESC;

-- name: ListEntityIncomingLinks :many
SELECT * FROM entity_links
WHERE target_type = $1 AND target_id = $2
ORDER BY created_at DESC;

-- name: ListEntityAllLinks :many
SELECT
    el.*,
    CASE
        WHEN el.source_type = $1 AND el.source_id = $2 THEN 'outgoing'
        ELSE 'incoming'
    END as direction
FROM entity_links el
WHERE (el.source_type = $1 AND el.source_id = $2)
   OR (el.target_type = $1 AND el.target_id = $2 AND el.is_bidirectional = FALSE)
ORDER BY el.created_at DESC;

-- name: GetEntityLink :one
SELECT * FROM entity_links
WHERE id = $1;

-- name: CreateEntityLink :one
INSERT INTO entity_links (
    user_id,
    source_type, source_id, source_name,
    target_type, target_id, target_name,
    link_type, custom_link_type, is_bidirectional,
    description, metadata, created_by
) VALUES (
    $1,
    $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10,
    $11, $12, $13
)
RETURNING *;

-- name: UpdateEntityLink :one
UPDATE entity_links
SET description = $2,
    metadata = $3,
    source_name = $4,
    target_name = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteEntityLink :exec
DELETE FROM entity_links
WHERE id = $1;

-- name: DeleteBidirectionalLink :exec
DELETE FROM entity_links
WHERE (source_type = $1 AND source_id = $2 AND target_type = $3 AND target_id = $4)
   OR (source_type = $3 AND source_id = $4 AND target_type = $1 AND target_id = $2);

-- name: DeleteAllEntityLinks :exec
DELETE FROM entity_links
WHERE (source_type = $1 AND source_id = $2)
   OR (target_type = $1 AND target_id = $2);

-- name: ListLinksByType :many
SELECT * FROM entity_links
WHERE user_id = $1 AND link_type = $2
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_val)::int;

-- name: FindLink :one
SELECT * FROM entity_links
WHERE source_type = $1 AND source_id = $2
  AND target_type = $3 AND target_id = $4
  AND link_type = $5
LIMIT 1;

-- name: HasBlockers :one
SELECT EXISTS (
    SELECT 1 FROM entity_links
    WHERE target_type = $1 AND target_id = $2 AND link_type = 'blocks'
) as has_blockers;

-- name: ListBlockers :many
SELECT * FROM entity_links
WHERE target_type = $1 AND target_id = $2 AND link_type = 'blocks';

-- name: ListBlocking :many
SELECT * FROM entity_links
WHERE source_type = $1 AND source_id = $2 AND link_type = 'blocks';

-- name: ListDependencies :many
SELECT * FROM entity_links
WHERE source_type = $1 AND source_id = $2 AND link_type = 'depends_on';

-- name: ListDependents :many
SELECT * FROM entity_links
WHERE target_type = $1 AND target_id = $2 AND link_type = 'depends_on';

-- name: CountEntityLinks :one
SELECT COUNT(*) as count FROM entity_links
WHERE (source_type = $1 AND source_id = $2)
   OR (target_type = $1 AND target_id = $2);
