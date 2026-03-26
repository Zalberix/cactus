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

-- name: GetMessageStatusByID :one
-- Status API: returns message + latest workflow_run status (per D-20, D-21)
SELECT m.id, m.status AS message_status, m.created_at,
       wr.id AS workflow_run_id, wr.status AS workflow_status,
       wr.started_at AS run_started_at, wr.completed_at AS run_completed_at,
       wr.error_message AS run_error_message
FROM "message" m
LEFT JOIN "workflow_run" wr ON wr.message_id = m.id
WHERE m.id = $1 AND m.deleted_at IS NULL
ORDER BY wr.id DESC
LIMIT 1;
