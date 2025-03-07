-- name: GetTypeWorkerBySlug :one
SELECT * FROM type_worker
WHERE slug = $1
LIMIT 1;

-- name: GetTypeWorkers :many
SELECT * FROM type_worker;

-- name: CreateTypeWorker :one
INSERT INTO type_worker (slug, "name")
VALUES($1, $2)
RETURNING *;
