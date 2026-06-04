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
    wv.locked_at,
    wv.published_at,
    wv.archived_at,
    wv.created_at,
    wv.updated_at,
    wv.deleted_at,
    COUNT(DISTINCT wr.id)::bigint AS run_count,
    COUNT(DISTINCT c.workflow_input_schema_id)::bigint AS compatible_schema_count
FROM "workflow_version" wv
LEFT JOIN "workflow_run" wr ON wr.workflow_version_id = wv.id
LEFT JOIN "workflow_version_input_schema_compatibility" c
    ON c.workflow_version_id = wv.id
    AND c.is_active = TRUE
    AND c.deleted_at IS NULL
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
WHERE workflow_id = $1 AND is_active = TRUE AND deleted_at IS NULL AND archived_at IS NULL
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
SET is_active = $2,
    locked_at = CASE WHEN $2 THEN COALESCE(locked_at, CURRENT_TIMESTAMP) ELSE locked_at END,
    published_at = CASE WHEN $2 THEN COALESCE(published_at, CURRENT_TIMESTAMP) ELSE published_at END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ArchiveWorkflowVersion :one
UPDATE "workflow_version"
SET archived_at = COALESCE(archived_at, CURRENT_TIMESTAMP),
    is_active = FALSE,
    updated_by_user_id = $2,
    updated_at = CURRENT_TIMESTAMP
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

-- name: IsWorkflowVersionReferenced :one
SELECT EXISTS (
    SELECT 1 FROM "workflow_run" wr WHERE wr.workflow_version_id = $1
    UNION ALL
    SELECT 1 FROM "workflow_experiment_variant" ev WHERE ev.workflow_version_id = $1 AND ev.deleted_at IS NULL
    UNION ALL
    SELECT 1 FROM "workflow_version_input_schema_compatibility" c WHERE c.workflow_version_id = $1 AND c.deleted_at IS NULL
) AS is_referenced;

-- name: SoftDeleteWorkflowVersion :exec
UPDATE "workflow_version"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;