-- name: CreateSystemToken :one
INSERT INTO "system_token" (system_id, public_token, private_token, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSystemTokenByID :one
SELECT * FROM "system_token"
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetSystemTokenByPublicToken :one
SELECT * FROM "system_token"
WHERE public_token = $1 AND is_active = TRUE AND deleted_at IS NULL
LIMIT 1;

-- name: ListSystemTokensBySystemID :many
SELECT * FROM "system_token"
WHERE system_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: DeactivateSystemToken :exec
UPDATE "system_token"
SET is_active = FALSE, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: SoftDeleteSystemToken :exec
UPDATE "system_token"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
