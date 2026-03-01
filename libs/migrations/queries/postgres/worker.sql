-- name: CreateWorker :one
INSERT INTO worker (channel_id, config_id, is_active)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetWorkerByID :one
SELECT * FROM worker
WHERE id = $1
LIMIT 1;

-- name: GetChannelSlugByWorkerID :one
SELECT c.slug
FROM worker w
JOIN channel c ON c.id = w.channel_id
WHERE w.id = $1;
