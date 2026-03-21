-- name: CreateWorkerSettingsRevision :one
INSERT INTO "worker_settings_revision" (worker_settings_schema_id, created_by_user_id, settings_data)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetWorkerSettingsRevisionByID :one
SELECT * FROM "worker_settings_revision"
WHERE id = $1;

-- name: ListWorkerSettingsRevisionsBySchemaID :many
SELECT * FROM "worker_settings_revision"
WHERE worker_settings_schema_id = $1
ORDER BY created_at DESC;
