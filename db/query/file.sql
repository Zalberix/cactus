-- name: CreateFile :one
INSERT INTO file (
    message_id,
    title,
    "name",
    ext,
    url
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFilesByMessageID :many
SELECT * FROM file
WHERE message_id = $1;
