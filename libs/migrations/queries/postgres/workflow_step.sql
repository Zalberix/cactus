-- name: CreateWorkflowStep :one
INSERT INTO "workflow_step" (workflow_version_id, step_type, work_type_id, worker_settings_revision_id, control_kind, control_settings, input_mapping, canvas_position)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
    canvas_position = $8,
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

-- name: UpdateWorkflowStepPosition :exec
UPDATE "workflow_step"
SET canvas_position = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListEnrichedStepsByVersionID :many
SELECT
    ws.*,
    wt.name AS work_type_name,
    wt.code AS work_type_code,
    wt.meta AS work_type_meta,
    wss.settings_schema,
    wss.input_schema,
    wss.output_schema
FROM "workflow_step" ws
LEFT JOIN "work_type" wt ON wt.id = ws.work_type_id
LEFT JOIN "worker_settings_revision" wsr ON wsr.id = ws.worker_settings_revision_id
LEFT JOIN "worker_settings_schema" wss ON wss.id = wsr.worker_settings_schema_id
WHERE ws.workflow_version_id = $1 AND ws.deleted_at IS NULL
ORDER BY ws.id;
