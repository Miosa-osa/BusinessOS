-- User External Apps Queries
-- Full CRUD operations for managing user-added web applications

-- name: ListUserExternalApps :many
-- Get all active external apps for a workspace
SELECT
    id,
    user_id,
    workspace_id,
    name,
    url,
    icon,
    color,
    logo_url,
    category,
    description,
    position_x,
    position_y,
    position_z,
    iframe_config,
    is_active,
    open_on_startup,
    app_type,
    created_at,
    updated_at,
    last_opened_at
FROM user_external_apps
WHERE workspace_id = $1
  AND is_active = true
ORDER BY created_at DESC;

-- name: ListAllUserExternalApps :many
-- Get all external apps for a workspace (including inactive)
SELECT
    id,
    user_id,
    workspace_id,
    name,
    url,
    icon,
    color,
    logo_url,
    category,
    description,
    position_x,
    position_y,
    position_z,
    iframe_config,
    is_active,
    open_on_startup,
    app_type,
    created_at,
    updated_at,
    last_opened_at
FROM user_external_apps
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: GetUserExternalApp :one
-- Get a specific external app by ID
SELECT
    id,
    user_id,
    workspace_id,
    name,
    url,
    icon,
    color,
    logo_url,
    category,
    description,
    position_x,
    position_y,
    position_z,
    iframe_config,
    is_active,
    open_on_startup,
    app_type,
    created_at,
    updated_at,
    last_opened_at
FROM user_external_apps
WHERE id = $1 AND workspace_id = $2;

-- name: CreateUserExternalApp :one
-- Create a new external app
INSERT INTO user_external_apps (
    user_id,
    workspace_id,
    name,
    url,
    icon,
    color,
    logo_url,
    category,
    description,
    iframe_config,
    open_on_startup,
    app_type
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING
    id,
    user_id,
    workspace_id,
    name,
    url,
    icon,
    color,
    logo_url,
    category,
    description,
    position_x,
    position_y,
    position_z,
    iframe_config,
    is_active,
    open_on_startup,
    app_type,
    created_at,
    updated_at,
    last_opened_at;

-- name: UpdateUserExternalApp :one
-- Update an existing external app
UPDATE user_external_apps
SET
    name = COALESCE(sqlc.narg('name'), name),
    url = COALESCE(sqlc.narg('url'), url),
    icon = COALESCE(sqlc.narg('icon'), icon),
    color = COALESCE(sqlc.narg('color'), color),
    logo_url = COALESCE(sqlc.narg('logo_url'), logo_url),
    category = COALESCE(sqlc.narg('category'), category),
    description = COALESCE(sqlc.narg('description'), description),
    position_x = COALESCE(sqlc.narg('position_x'), position_x),
    position_y = COALESCE(sqlc.narg('position_y'), position_y),
    position_z = COALESCE(sqlc.narg('position_z'), position_z),
    iframe_config = COALESCE(sqlc.narg('iframe_config'), iframe_config),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    open_on_startup = COALESCE(sqlc.narg('open_on_startup'), open_on_startup),
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING
    id,
    user_id,
    workspace_id,
    name,
    url,
    icon,
    color,
    logo_url,
    category,
    description,
    position_x,
    position_y,
    position_z,
    iframe_config,
    is_active,
    open_on_startup,
    app_type,
    created_at,
    updated_at,
    last_opened_at;

-- name: DeleteUserExternalApp :exec
-- Delete an external app
DELETE FROM user_external_apps
WHERE id = $1 AND workspace_id = $2;

-- name: RecordAppOpened :exec
-- Update last_opened_at timestamp when app is accessed
UPDATE user_external_apps
SET last_opened_at = NOW()
WHERE id = $1;

-- name: UpdateAppPosition :exec
-- Update app position on 3D desktop (called when user moves window)
UPDATE user_external_apps
SET
    position_x = $2,
    position_y = $3,
    position_z = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: ToggleAppActive :exec
-- Enable/disable an app without deleting
UPDATE user_external_apps
SET
    is_active = $2,
    updated_at = NOW()
WHERE id = $1 AND workspace_id = $3;

-- name: GetStartupApps :many
-- Get all apps configured to open on startup
SELECT
    id,
    user_id,
    workspace_id,
    name,
    url,
    icon,
    color,
    logo_url,
    category,
    description,
    position_x,
    position_y,
    position_z,
    iframe_config,
    is_active,
    open_on_startup,
    app_type,
    created_at,
    updated_at,
    last_opened_at
FROM user_external_apps
WHERE workspace_id = $1
  AND is_active = true
  AND open_on_startup = true
ORDER BY created_at ASC;
