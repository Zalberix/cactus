-- name: CreateNewMessage :one
INSERT INTO "message" (workflow_id, external_message_id, overridden_priority, value, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetNewMessageByID :one
SELECT * FROM "message"
WHERE id = $1 AND deleted_at IS NULL;

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
