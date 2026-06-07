-- name: CreateNewMessage :one
INSERT INTO "message" (
    workflow_id,
    workflow_input_schema_id,
    external_message_id,
    idempotency_key,
    overridden_priority,
    value,
    metadata,
    status,
    error_message
)
VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7, '{}'::jsonb), $8, $9)
RETURNING *;

-- name: GetNewMessageByID :one
SELECT * FROM "message"
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetMessageByIdempotencyKey :one
SELECT * FROM "message"
WHERE workflow_id = $1
  AND idempotency_key = $2
  AND deleted_at IS NULL
LIMIT 1;

-- name: ListNewMessagesByWorkflowID :many
SELECT * FROM "message"
WHERE workflow_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateNewMessageStatus :one
UPDATE "message"
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteNewMessage :exec
UPDATE "message"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetMessageStatusByID :one
-- Status API: returns message + latest workflow_run status (per D-20, D-21)
SELECT m.id, m.status AS message_status, m.created_at,
       wr.id AS workflow_run_id, wr.status AS workflow_status,
       wr.started_at AS run_started_at, wr.completed_at AS run_completed_at,
       wr.error_message AS run_error_message
FROM "message" m
LEFT JOIN "workflow_run" wr ON wr.message_id = m.id
WHERE m.id = $1 AND m.deleted_at IS NULL
ORDER BY wr.id DESC
LIMIT 1;

-- name: GetMessageDetailByID :one
SELECT
    m.id,
    m.workflow_id,
    w.name AS workflow_name,
    m.workflow_input_schema_id,
    wis.code AS workflow_input_schema_code,
    wis.version_number AS workflow_input_schema_version_number,
    m.value AS message_value,
    m.metadata AS message_metadata,
    m.status AS message_status,
    m.error_message AS message_error_message,
    m.created_at,
    m.updated_at,
    wr.id AS workflow_run_id,
    wr.workflow_version_id,
    wv.version_number AS workflow_version_number,
    wv.name AS workflow_version_name,
    wr.workflow_experiment_id,
    wr.workflow_experiment_scope_id,
    wr.workflow_experiment_variant_id,
    wr.input_schema_compatibility_id,
    wr.selection_reason,
    wr.version_input_data,
    wr.routing_decision,
    wr.status AS workflow_status,
    wr.started_at AS run_started_at,
    wr.completed_at AS run_completed_at,
    wr.error_message AS run_error_message
FROM "message" m
JOIN "workflow" w ON w.id = m.workflow_id AND w.deleted_at IS NULL
LEFT JOIN "workflow_input_schema" wis ON wis.id = m.workflow_input_schema_id AND wis.deleted_at IS NULL
LEFT JOIN "workflow_run" wr ON wr.message_id = m.id
LEFT JOIN "workflow_version" wv ON wv.id = wr.workflow_version_id AND wv.deleted_at IS NULL
WHERE m.id = $1 AND m.deleted_at IS NULL
ORDER BY wr.id DESC
LIMIT 1;

-- name: ListMessagesByOrganizationID :many
-- List messages for all workflows belonging to systems within an organization.
-- Joins: message -> workflow -> system (filtered by organization_id), plus latest workflow_run -> workflow_version.
SELECT
    m.id,
    m.workflow_id,
    w."name" AS workflow_name,
    m.workflow_input_schema_id,
    wis.code AS workflow_input_schema_code,
    wis.version_number AS workflow_input_schema_version_number,
    m.status,
    m.created_at,
    m.updated_at,
    wv.id AS workflow_version_id,
    wv.version_number AS workflow_version_number,
    wv.name AS workflow_version_name,
    wr.workflow_experiment_id,
    wr.workflow_experiment_variant_id,
    COALESCE(wr.selection_reason, 'standard') AS selection_reason
FROM "message" m
JOIN "workflow" w ON w.id = m.workflow_id AND w.deleted_at IS NULL
JOIN "system" s ON s.id = w.system_id AND s.deleted_at IS NULL
LEFT JOIN "workflow_input_schema" wis ON wis.id = m.workflow_input_schema_id AND wis.deleted_at IS NULL
LEFT JOIN LATERAL (
    SELECT
        run.workflow_version_id,
        run.workflow_experiment_id,
        run.workflow_experiment_variant_id,
        run.selection_reason
    FROM "workflow_run" run
    WHERE run.message_id = m.id
    ORDER BY run.id DESC
    LIMIT 1
) wr ON TRUE
LEFT JOIN "workflow_version" wv ON wv.id = wr.workflow_version_id AND wv.deleted_at IS NULL
WHERE s.organization_id = $1 AND m.deleted_at IS NULL
ORDER BY m.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountMessagesByOrganizationID :one
SELECT COUNT(*) AS total
FROM "message" m
JOIN "workflow" w ON w.id = m.workflow_id AND w.deleted_at IS NULL
JOIN "system" s ON s.id = w.system_id AND s.deleted_at IS NULL
WHERE s.organization_id = $1 AND m.deleted_at IS NULL;
