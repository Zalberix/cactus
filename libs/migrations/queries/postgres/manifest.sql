-- name: CreateManifest :one
INSERT INTO manifest (value)
VALUES ($1)
RETURNING *;

-- name: GetManifestByID :one
SELECT * FROM manifest
WHERE id = $1
LIMIT 1;
