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

-- name: CloneWorkerSettingsRevision :one
INSERT INTO "worker_settings_revision" (worker_settings_schema_id, created_by_user_id, settings_data)
SELECT worker_settings_schema_id, $2, settings_data
FROM "worker_settings_revision"
WHERE id = $1
RETURNING *;

-- name: UpdateWorkerSettingsRevisionSettings :one
UPDATE "worker_settings_revision"
SET settings_data = $2
WHERE id = $1
RETURNING *;
