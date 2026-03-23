-- name: CreateWorkflowVersion :one
INSERT INTO "workflow_version" (workflow_id, created_by_user_id, version_number, is_valid, is_active, traffic_weight, is_control_group)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetWorkflowVersionByID :one
SELECT * FROM "workflow_version"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowVersionsByWorkflowID :many
SELECT * FROM "workflow_version"
WHERE workflow_id = $1 AND deleted_at IS NULL
ORDER BY version_number;

-- name: ListActiveWorkflowVersions :many
SELECT * FROM "workflow_version"
WHERE workflow_id = $1 AND is_active = TRUE AND deleted_at IS NULL
ORDER BY version_number;

-- name: UpdateWorkflowVersionValid :one
UPDATE "workflow_version"
SET is_valid = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWorkflowVersionActive :one
UPDATE "workflow_version"
SET is_active = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWorkflowVersionTrafficWeight :one
UPDATE "workflow_version"
SET traffic_weight = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: GetMaxVersionNumberByWorkflowID :one
SELECT COALESCE(MAX(version_number), 0)::int FROM "workflow_version"
WHERE workflow_id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteWorkflowVersion :exec
UPDATE "workflow_version"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
