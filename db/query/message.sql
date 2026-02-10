-- name: CreateMessage :one
INSERT INTO message (
    system_id,
    manifest_id,
    "uuid",
    priority,
    value,
    send_at
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMessageByUUID :one
SELECT * FROM message
WHERE "uuid" = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetMessagesBySystemID :many
SELECT * FROM message
WHERE system_id = $1 AND deleted_at IS NULL;
