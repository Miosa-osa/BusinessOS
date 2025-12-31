-- name: GetCachedSearchResult :one
SELECT * FROM web_search_results
WHERE query_hash = $1 AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: GetCachedSearchResultForConversation :one
SELECT * FROM web_search_results
WHERE query_hash = $1 AND conversation_id = $2 AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: SaveSearchResult :one
INSERT INTO web_search_results (
    query_hash,
    original_query,
    optimized_query,
    user_id,
    conversation_id,
    results,
    result_count,
    provider,
    search_time_ms,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: IncrementCacheHit :exec
UPDATE web_search_results
SET hit_count = hit_count + 1, last_hit_at = NOW()
WHERE id = $1;

-- name: CleanupExpiredCache :execrows
DELETE FROM web_search_results WHERE expires_at < NOW();

-- name: GetCacheStats :one
SELECT
    COUNT(*) as total_entries,
    COUNT(*) FILTER (WHERE expires_at > NOW()) as active_entries,
    COALESCE(SUM(hit_count), 0) as total_hits,
    COALESCE(AVG(search_time_ms), 0) as avg_search_time_ms
FROM web_search_results;

-- name: DeleteCacheForConversation :exec
DELETE FROM web_search_results WHERE conversation_id = $1;

-- name: DeleteCacheForUser :exec
DELETE FROM web_search_results WHERE user_id = $1;

-- name: ListSearchHistory :many
SELECT id, original_query, optimized_query, result_count, provider, search_time_ms, created_at
FROM web_search_results
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetSearchHistoryCount :one
SELECT COUNT(*) FROM web_search_results WHERE user_id = $1;

-- name: GetSearchResultById :one
SELECT * FROM web_search_results WHERE id = $1 AND user_id = $2;
