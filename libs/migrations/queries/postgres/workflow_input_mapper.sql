-- name: CreateWorkflowInputMapper :one
INSERT INTO "workflow_input_mapper" (workflow_id, mapper_type, rules, is_active, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetWorkflowInputMapperByID :one
SELECT * FROM "workflow_input_mapper"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowInputMappersByWorkflowID :many
SELECT * FROM "workflow_input_mapper"
WHERE workflow_id = $1 AND deleted_at IS NULL
ORDER BY is_active DESC, id;

-- name: UpdateWorkflowInputMapper :one
UPDATE "workflow_input_mapper"
SET mapper_type = $2,
    rules = $3,
    is_active = $4,
    updated_by_user_id = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWorkflowInputMapperActive :one
UPDATE "workflow_input_mapper"
SET is_active = $2,
    updated_by_user_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflowInputMapper :exec
UPDATE "workflow_input_mapper"
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
