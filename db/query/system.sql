-- name: GetSystemByID :one
SELECT * FROM "system"
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateSystem :one
INSERT INTO "system" (
    user_creator_id,
    "name",
    description,
    is_active,
    priority,
    public_token,
    private_token
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: AddChannelForSystem :exec
INSERT INTO channel_system (
    system_id,
    channel_id
) VALUES ($1, $2);

-- name: GetSystemByPublicToken :one
SELECT * FROM "system"
WHERE public_token = $1 AND deleted_at IS NULL
LIMIT 1;
