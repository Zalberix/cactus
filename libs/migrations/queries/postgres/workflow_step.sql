-- name: CreateWorkflowStep :one
INSERT INTO "workflow_step" (workflow_version_id, step_type, work_type_id, worker_settings_revision_id, control_kind, control_settings, input_mapping)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetWorkflowStepByID :one
SELECT * FROM "workflow_step"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowStepsByVersionID :many
SELECT * FROM "workflow_step"
WHERE workflow_version_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: UpdateWorkflowStep :one
UPDATE "workflow_step"
SET step_type = $2,
    work_type_id = $3,
    worker_settings_revision_id = $4,
    control_kind = $5,
    control_settings = $6,
    input_mapping = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflowStep :exec
UPDATE "workflow_step"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteWorkflowStepsByVersionID :exec
UPDATE "workflow_step"
SET deleted_at = CURRENT_TIMESTAMP
WHERE workflow_version_id = $1 AND deleted_at IS NULL;
