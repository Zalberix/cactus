-- name: CreateKindWorker :one 
INSERT INTO kind_worker (
    "name", 
    slug, 
    config_schema, 
    config
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetKindWokerById :one 
SELECT * 
FROM kind_worker kw
WHERE kw.id = $1
LIMIT 1;