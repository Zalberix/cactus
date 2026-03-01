-- name: CreateConfig :one
INSERT INTO config (
    "name",
    config_schema,
    config
) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetConfigByID :one
SELECT * FROM config
WHERE id = $1
LIMIT 1;
