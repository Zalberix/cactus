-- name: CreateRevisionRateLimit :one
INSERT INTO "revision_rate_limit" (worker_settings_revision_id, "limit", window_seconds)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListRevisionRateLimitsByRevisionID :many
SELECT * FROM "revision_rate_limit"
WHERE worker_settings_revision_id = $1
ORDER BY id;

-- name: UpdateRevisionRateLimit :one
UPDATE "revision_rate_limit"
SET "limit" = $2, window_seconds = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRevisionRateLimit :exec
DELETE FROM "revision_rate_limit"
WHERE id = $1;
