-- name: CreateWorkflow :one
INSERT INTO "workflow" (system_id, "name", priority, input_schema, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetWorkflowByID :one
SELECT * FROM "workflow"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowsBySystemID :many
SELECT * FROM "workflow"
WHERE system_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: UpdateWorkflow :one
UPDATE "workflow"
SET "name" = $2, priority = $3, description = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWorkflowInputSchema :one
UPDATE "workflow"
SET input_schema = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflow :exec
UPDATE "workflow"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
