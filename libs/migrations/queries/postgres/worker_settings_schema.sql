-- name: CreateWorkerSettingsSchema :one
INSERT INTO "worker_settings_schema" (work_type_id, "version", settings_schema, input_schema, output_schema)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetWorkerSettingsSchemaByID :one
SELECT * FROM "worker_settings_schema"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkerSettingsSchemasByWorkTypeID :many
SELECT * FROM "worker_settings_schema"
WHERE work_type_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: SoftDeleteWorkerSettingsSchema :exec
UPDATE "worker_settings_schema"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetWorkerSettingsSchemaByVersion :one
SELECT * FROM "worker_settings_schema"
WHERE work_type_id = $1 AND "version" = $2 AND deleted_at IS NULL
LIMIT 1;
