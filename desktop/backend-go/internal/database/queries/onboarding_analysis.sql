-- name: CreateUserAnalysis :one
INSERT INTO onboarding_user_analysis (
    user_id,
    workspace_id,
    insights,
    interests,
    tools_used,
    profile_summary,
    email_metadata,
    total_emails_analyzed,
    sender_domains,
    detected_patterns,
    analysis_model,
    ai_provider,
    analysis_tokens_used,
    analysis_duration_ms,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
RETURNING *;

-- name: GetUserAnalysisByID :one
SELECT * FROM onboarding_user_analysis
WHERE id = $1;

-- name: GetUserAnalysisByUserAndWorkspace :one
SELECT * FROM onboarding_user_analysis
WHERE user_id = $1 AND workspace_id = $2;

-- name: UpdateUserAnalysisStatus :one
UPDATE onboarding_user_analysis
SET
    status = $2,
    error_message = $3,
    completed_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE completed_at END
WHERE id = $1
RETURNING *;

-- name: UpdateUserAnalysisResults :one
UPDATE onboarding_user_analysis
SET
    insights = $2,
    interests = $3,
    tools_used = $4,
    profile_summary = $5,
    email_metadata = $6,
    total_emails_analyzed = $7,
    sender_domains = $8,
    detected_patterns = $9,
    analysis_tokens_used = $10,
    analysis_duration_ms = $11,
    status = 'completed',
    completed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListUserAnalysesByUser :many
SELECT * FROM onboarding_user_analysis
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteUserAnalysis :exec
DELETE FROM onboarding_user_analysis
WHERE id = $1;

-- ========================================
-- STARTER APPS QUERIES
-- ========================================

-- name: CreateStarterApp :one
INSERT INTO onboarding_starter_apps (
    user_id,
    workspace_id,
    analysis_id,
    title,
    description,
    icon_emoji,
    category,
    reasoning,
    customization_prompt,
    based_on_interests,
    based_on_tools,
    base_module,
    module_customizations,
    generation_model,
    ai_provider,
    display_order,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- name: GetStarterAppByID :one
SELECT * FROM onboarding_starter_apps
WHERE id = $1;

-- name: ListStarterAppsByAnalysis :many
SELECT * FROM onboarding_starter_apps
WHERE analysis_id = $1
ORDER BY display_order ASC;

-- name: ListStarterAppsByWorkspace :many
SELECT * FROM onboarding_starter_apps
WHERE workspace_id = $1
ORDER BY display_order ASC;

-- name: UpdateStarterAppStatus :one
UPDATE onboarding_starter_apps
SET
    status = $2,
    osa_workflow_id = $3,
    error_message = $4,
    completed_at = CASE WHEN $2 = 'ready' THEN NOW() ELSE completed_at END
WHERE id = $1
RETURNING *;

-- name: UpdateStarterAppGeneration :one
UPDATE onboarding_starter_apps
SET
    status = $2,
    icon_url = $3,
    generation_tokens_used = $4,
    generation_duration_ms = $5,
    completed_at = CASE WHEN $2 = 'ready' THEN NOW() ELSE completed_at END
WHERE id = $1
RETURNING *;

-- name: GetStarterAppsByUserAndWorkspace :many
SELECT * FROM onboarding_starter_apps
WHERE user_id = $1 AND workspace_id = $2
ORDER BY display_order ASC;

-- name: DeleteStarterApp :exec
DELETE FROM onboarding_starter_apps
WHERE id = $1;

-- name: DeleteStarterAppsByAnalysis :exec
DELETE FROM onboarding_starter_apps
WHERE analysis_id = $1;

-- name: CountStarterAppsByStatus :one
SELECT COUNT(*) FROM onboarding_starter_apps
WHERE workspace_id = $1 AND status = $2;

-- ========================================
-- EMAIL METADATA QUERIES
-- ========================================

-- name: CreateEmailMetadata :one
INSERT INTO onboarding_email_metadata (
    user_id,
    analysis_id,
    email_id,
    external_id,
    sender_domain,
    sender_email,
    subject_keywords,
    body_keywords,
    detected_tools,
    detected_topics,
    category,
    sentiment,
    importance_score,
    email_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: ListEmailMetadataByAnalysis :many
SELECT * FROM onboarding_email_metadata
WHERE analysis_id = $1
ORDER BY email_date DESC;

-- name: GetEmailMetadataStats :one
SELECT
    COUNT(*) as total_emails,
    COUNT(DISTINCT sender_domain) as unique_domains,
    COUNT(DISTINCT category) as unique_categories,
    AVG(importance_score) as avg_importance
FROM onboarding_email_metadata
WHERE analysis_id = $1;

-- name: GetTopSenderDomains :many
SELECT
    sender_domain,
    COUNT(*) as email_count
FROM onboarding_email_metadata
WHERE analysis_id = $1 AND sender_domain IS NOT NULL
GROUP BY sender_domain
ORDER BY email_count DESC
LIMIT $2;

-- name: GetDetectedToolsFrequency :many
SELECT
    jsonb_array_elements_text(detected_tools) as tool,
    COUNT(*) as frequency
FROM onboarding_email_metadata
WHERE analysis_id = $1
GROUP BY tool
ORDER BY frequency DESC
LIMIT $2;

-- name: DeleteEmailMetadataByAnalysis :exec
DELETE FROM onboarding_email_metadata
WHERE analysis_id = $1;

-- ========================================
-- COMBINED QUERIES FOR ONBOARDING FLOW
-- ========================================

-- name: GetCompleteOnboardingProfile :one
SELECT
    a.*,
    (
        SELECT json_agg(s.* ORDER BY s.display_order)
        FROM onboarding_starter_apps s
        WHERE s.analysis_id = a.id
    ) as starter_apps,
    (
        SELECT COUNT(*)
        FROM onboarding_email_metadata e
        WHERE e.analysis_id = a.id
    ) as total_metadata_entries
FROM onboarding_user_analysis a
WHERE a.user_id = $1 AND a.workspace_id = $2;

-- name: CheckOnboardingProgress :one
SELECT
    a.status as analysis_status,
    COUNT(s.id) as total_apps,
    COUNT(CASE WHEN s.status = 'ready' THEN 1 END) as ready_apps,
    COUNT(CASE WHEN s.status = 'failed' THEN 1 END) as failed_apps,
    COUNT(CASE WHEN s.status = 'generating' THEN 1 END) as generating_apps
FROM onboarding_user_analysis a
LEFT JOIN onboarding_starter_apps s ON s.analysis_id = a.id
WHERE a.user_id = $1 AND a.workspace_id = $2
GROUP BY a.status;
