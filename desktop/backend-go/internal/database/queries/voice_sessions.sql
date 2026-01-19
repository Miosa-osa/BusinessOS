-- name: CreateVoiceSession :one
INSERT INTO voice_sessions (
    session_id,
    user_id,
    workspace_id,
    agent_role,
    state
) VALUES (
    $1, $2, $3, $4, 'active'
) RETURNING *;

-- name: GetVoiceSessionByID :one
SELECT * FROM voice_sessions
WHERE id = $1;

-- name: GetVoiceSessionBySessionID :one
SELECT * FROM voice_sessions
WHERE session_id = $1;

-- name: GetVoiceSessionBySessionIDAndUser :one
SELECT * FROM voice_sessions
WHERE session_id = $1 AND user_id = $2;

-- name: ListActiveVoiceSessionsForUser :many
SELECT * FROM voice_sessions
WHERE user_id = $1 AND state = 'active'
ORDER BY created_at DESC
LIMIT $2;

-- name: ListVoiceSessionsForWorkspace :many
SELECT * FROM voice_sessions
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateVoiceSessionState :exec
UPDATE voice_sessions
SET state = $2,
    ended_at = CASE
        WHEN $2 = 'ended' THEN NOW()
        ELSE ended_at
    END,
    last_activity_at = NOW()
WHERE session_id = $1;

-- name: UpdateVoiceSessionActivity :exec
UPDATE voice_sessions
SET last_activity_at = NOW(),
    total_messages = total_messages + 1
WHERE session_id = $1;

-- name: UpdateVoiceSessionDuration :exec
UPDATE voice_sessions
SET total_duration_seconds = $2,
    last_activity_at = NOW()
WHERE session_id = $1;

-- name: DeleteVoiceSession :exec
DELETE FROM voice_sessions
WHERE session_id = $1;

-- name: DeleteVoiceSessionByID :exec
DELETE FROM voice_sessions
WHERE id = $1;

-- Voice Session Events Queries
-- name: CreateVoiceSessionEvent :one
INSERT INTO voice_session_events (
    session_id,
    event_type,
    event_data,
    duration_ms,
    error_message
) VALUES (
    (SELECT vs.id FROM voice_sessions vs WHERE vs.session_id = $1),
    $2, $3, $4, $5
) RETURNING *;

-- name: ListVoiceSessionEvents :many
SELECT vse.*
FROM voice_session_events vse
JOIN voice_sessions vs ON vse.session_id = vs.id
WHERE vs.session_id = sqlc.arg(target_session_id)
ORDER BY vse.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetVoiceSessionEventStats :one
SELECT
    COUNT(*) as total_events,
    COUNT(*) FILTER (WHERE event_type = 'error') as error_count,
    AVG(duration_ms) FILTER (WHERE duration_ms IS NOT NULL) as avg_duration_ms,
    MAX(created_at) as last_event_at
FROM voice_session_events vse
JOIN voice_sessions vs ON vse.session_id = vs.id
WHERE vs.session_id = $1;

-- User Facts Queries (for context retrieval)
-- name: CreateUserFact :one
INSERT INTO user_facts (
    user_id,
    fact_type,
    fact_key,
    fact_value,
    confidence_score
) VALUES (
    $1, $2, $3, $4, $5
) ON CONFLICT (user_id, fact_key)
DO UPDATE SET
    fact_value = EXCLUDED.fact_value,
    confidence_score = EXCLUDED.confidence_score,
    last_confirmed_at = NOW(),
    updated_at = NOW()
RETURNING *;

-- name: GetUserFactsByUser :many
SELECT * FROM user_facts
WHERE user_id = $1 AND is_active = TRUE
ORDER BY confidence_score DESC, created_at DESC;

-- name: GetUserFactsByType :many
SELECT * FROM user_facts
WHERE user_id = $1
    AND fact_type = $2
    AND is_active = TRUE
ORDER BY confidence_score DESC;

-- name: GetUserFactByKey :one
SELECT * FROM user_facts
WHERE user_id = $1
    AND fact_key = $2
    AND is_active = TRUE;

-- name: DeleteUserFact :exec
DELETE FROM user_facts
WHERE id = $1;

-- name: DeleteUserFactByKey :exec
DELETE FROM user_facts
WHERE user_id = $1 AND fact_key = $2;
