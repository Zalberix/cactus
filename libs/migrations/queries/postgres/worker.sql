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
