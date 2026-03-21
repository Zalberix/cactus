-- name: CreateNewFile :one
INSERT INTO "file" (message_id, workflow_run_step_id, "name", bucket, object_key, content_type, size_bytes, hash, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetNewFileByID :one
SELECT * FROM "file"
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetNewFileByObjectKey :one
SELECT * FROM "file"
WHERE object_key = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListNewFilesByMessageID :many
SELECT * FROM "file"
WHERE message_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: ListNewFilesByRunStepID :many
SELECT * FROM "file"
WHERE workflow_run_step_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: SoftDeleteNewFile :exec
UPDATE "file"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
