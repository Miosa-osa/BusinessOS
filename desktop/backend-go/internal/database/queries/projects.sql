-- name: ListProjects :many
SELECT * FROM projects
WHERE user_id = $1
  AND (sqlc.narg(status)::projectstatus IS NULL OR status = sqlc.narg(status))
ORDER BY updated_at DESC;

-- name: GetProject :one
SELECT * FROM projects
WHERE id = $1 AND user_id = $2;

-- name: CreateProject :one
INSERT INTO projects (id, user_id, name, description, status, priority, client_name, project_type, project_metadata, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET name = $2, description = $3, status = $4, priority = $5, client_name = $6, project_type = $7, project_metadata = $8, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1 AND user_id = $2;

-- name: GetProjectNotes :many
SELECT * FROM project_notes
WHERE project_id = $1
ORDER BY created_at DESC;

-- name: AddProjectNote :one
INSERT INTO project_notes (project_id, content)
VALUES ($1, $2)
RETURNING *;
