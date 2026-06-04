-- name: CreateWorkflowInputMapper :one
INSERT INTO "workflow_input_mapper" (workflow_id, name, mapper_type, rules, is_active, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetWorkflowInputMapperByID :one
SELECT * FROM "workflow_input_mapper"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowInputMappersByWorkflowID :many
SELECT * FROM "workflow_input_mapper"
WHERE workflow_id = $1 AND deleted_at IS NULL
ORDER BY is_active DESC, name, id;

-- name: UpdateWorkflowInputMapper :one
UPDATE "workflow_input_mapper"
SET name = $2,
    mapper_type = $3,
    rules = $4,
    is_active = $5,
    updated_by_user_id = $6,
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