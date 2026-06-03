-- name: CreateWorkflowVersion :one
INSERT INTO "workflow_version" (workflow_id, created_by_user_id, version_number, name, is_valid, is_active, traffic_weight, is_control_group)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetWorkflowVersionByID :one
SELECT * FROM "workflow_version"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowVersionsByWorkflowID :many
SELECT * FROM "workflow_version"
WHERE workflow_id = $1 AND deleted_at IS NULL
ORDER BY version_number;

-- name: ListAllWorkflowVersionsByWorkflowID :many
SELECT * FROM "workflow_version"
WHERE workflow_id = $1
ORDER BY version_number;

-- name: ListWorkflowVersionSummariesByWorkflowID :many
SELECT
    wv.id,
    wv.workflow_id,
    COALESCE(wv.name, 'Version ' || wv.version_number::text) AS name,
    wv.version_number,
    wv.is_valid,
    wv.is_active,
    wv.traffic_weight,
    wv.is_control_group,
    wv.created_at,
    wv.updated_at,
    wv.deleted_at,
    COUNT(wr.id)::bigint AS run_count
FROM "workflow_version" wv
LEFT JOIN "workflow_run" wr ON wr.workflow_version_id = wv.id
WHERE wv.workflow_id = $1
GROUP BY wv.id
ORDER BY wv.is_active DESC, wv.version_number DESC;

-- name: ListWorkflowTrafficCandidatesByWorkflowID :many
SELECT
    wv.id,
    wv.workflow_id,
    wv.version_number,
    wv.is_active,
    wv.traffic_weight,
    wv.deleted_at,
    COUNT(msg.id)::bigint AS run_count
FROM "workflow_version" wv
LEFT JOIN "workflow_run" wr ON wr.workflow_version_id = wv.id
LEFT JOIN "message" msg
    ON msg.id = wr.message_id
    AND msg.created_at >= wv.traffic_updated_at
WHERE wv.workflow_id = $1
GROUP BY wv.id
ORDER BY wv.is_active DESC, wv.version_number DESC;

-- name: UpdateWorkflowVersionName :one
UPDATE "workflow_version"
SET name = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ListActiveWorkflowVersions :many
SELECT * FROM "workflow_version"
WHERE workflow_id = $1 AND is_active = TRUE AND deleted_at IS NULL
ORDER BY version_number;

-- name: UpdateWorkflowVersionValid :one
UPDATE "workflow_version"
SET is_valid = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: InvalidateWorkflowVersionsByWorkflowID :exec
UPDATE "workflow_version"
SET is_valid = FALSE, updated_at = CURRENT_TIMESTAMP
WHERE workflow_id = $1 AND deleted_at IS NULL;

-- name: UpdateWorkflowVersionActive :one
UPDATE "workflow_version"
SET is_active = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWorkflowVersionTrafficWeight :one
UPDATE "workflow_version"
SET traffic_weight = $2, traffic_updated_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWorkflowVersionTrafficWeightIncludingDeleted :one
UPDATE "workflow_version"
SET traffic_weight = $2, traffic_updated_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetMaxVersionNumberByWorkflowID :one
SELECT COALESCE(MAX(version_number), 0)::int FROM "workflow_version"
WHERE workflow_id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteWorkflowVersion :exec
UPDATE "workflow_version"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
