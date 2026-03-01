-- name: CreateChannel :one
INSERT INTO channel (slug, "name")
VALUES ( $1, $2)
RETURNING *;

-- name: GetChannelBySlug :one
SELECT * FROM channel
WHERE slug = $1
LIMIT 1;

-- name: GetChannels :many
SELECT * FROM channel;
