-- name: CreateOSAApp :one
INSERT INTO osa_apps (
    workspace_id,
    name,
    description,
    template_type,
    status,
    generation_context,
    deployment_config,
    app_metadata,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetOSAApp :one
SELECT * FROM osa_apps
WHERE id = $1;

-- name: GetOSAAppByName :one
SELECT * FROM osa_apps
WHERE workspace_id = $1 AND name = $2;

-- name: ListOSAAppsByWorkspace :many
SELECT * FROM osa_apps
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: ListOSAAppsByUser :many
SELECT a.* FROM osa_apps a
JOIN osa_workspaces w ON a.workspace_id = w.id
WHERE w.user_id = $1
ORDER BY a.created_at DESC;

-- name: ListOSAAppsByStatus :many
SELECT * FROM osa_apps
WHERE status = $1
ORDER BY created_at DESC;

-- name: UpdateOSAAppStatus :one
UPDATE osa_apps
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOSAAppDeploymentConfig :one
UPDATE osa_apps
SET deployment_config = deployment_config || $2::jsonb,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateOSAAppMetadata :one
UPDATE osa_apps
SET app_metadata = app_metadata || $2::jsonb,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteOSAApp :exec
DELETE FROM osa_apps
WHERE id = $1;

-- =============================================================================
-- OSA_DEPLOYMENTS QUERIES
-- =============================================================================

-- name: CreateOSADeployment :one
INSERT INTO osa_deployments (
    app_id,
    version,
    commit_sha,
    deployment_config,
    metadata,
    deployed_by
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetOSADeployment :one
SELECT * FROM osa_deployments
WHERE id = $1;

-- name: GetOSADeploymentByVersion :one
SELECT * FROM osa_deployments
WHERE app_id = $1 AND version = $2;

-- name: ListOSADeploymentsByApp :many
SELECT * FROM osa_deployments
WHERE app_id = $1
ORDER BY deployed_at DESC;

-- name: GetLatestOSADeployment :one
SELECT * FROM osa_deployments
WHERE app_id = $1
ORDER BY deployed_at DESC
LIMIT 1;

-- name: DeleteOSADeployment :exec
DELETE FROM osa_deployments
WHERE id = $1;
