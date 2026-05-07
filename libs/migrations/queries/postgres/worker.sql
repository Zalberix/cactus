-- name: CreateNewWorker :one
INSERT INTO "worker" (work_type_id, worker_settings_schema_id, "name", metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetNewWorkerByID :one
SELECT * FROM "worker"
WHERE id = $1;

-- name: ListNewWorkersByWorkTypeID :many
SELECT * FROM "worker"
WHERE work_type_id = $1
ORDER BY id;

-- name: UpdateNewWorkerHeartbeat :exec
UPDATE "worker"
SET last_heartbeat_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateNewWorkerSchema :one
UPDATE "worker"
SET worker_settings_schema_id = $2
WHERE id = $1
RETURNING *;

-- name: GetWorkerByWorkTypeAndName :one
SELECT * FROM "worker"
WHERE work_type_id = $1 AND "name" = $2
LIMIT 1;

-- name: DeleteWorker :exec
DELETE FROM "worker"
WHERE id = $1;

-- name: ListWorkflowUsagesByWorkerID :many
SELECT DISTINCT
    w.id AS workflow_id,
    w.name AS workflow_name,
    w.system_id AS system_id,
    wv.id AS workflow_version_id,
    wv.version_number AS workflow_version_number
FROM "worker" worker_row
JOIN "worker_settings_revision" wsr
    ON wsr.worker_settings_schema_id = worker_row.worker_settings_schema_id
JOIN "workflow_step" ws
    ON ws.worker_settings_revision_id = wsr.id
    AND ws.deleted_at IS NULL
JOIN "workflow_version" wv
    ON wv.id = ws.workflow_version_id
    AND wv.deleted_at IS NULL
JOIN "workflow" w
    ON w.id = wv.workflow_id
    AND w.deleted_at IS NULL
WHERE worker_row.id = $1
ORDER BY w.id, wv.version_number;
