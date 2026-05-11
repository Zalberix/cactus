-- name: CreateWorkflowVersionInput :one
INSERT INTO "workflow_version_input" (workflow_version_id, name, type, required, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetWorkflowVersionInputByID :one
SELECT * FROM "workflow_version_input"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowVersionInputs :many
SELECT * FROM "workflow_version_input"
WHERE workflow_version_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: UpdateWorkflowVersionInput :one
UPDATE "workflow_version_input"
SET name = $2, type = $3, required = $4, description = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflowVersionInput :exec
UPDATE "workflow_version_input"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
