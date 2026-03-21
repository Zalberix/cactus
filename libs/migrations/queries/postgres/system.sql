-- name: GetSystemByID :one
SELECT * FROM "system"
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateSystem :one
INSERT INTO "system" (
    organization_id,
    user_creator_id,
    "name",
    description,
    is_active,
    priority,
    public_token,
    private_token
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetSystemByPublicToken :one
SELECT * FROM "system"
WHERE public_token = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListSystemsByOrganizationID :many
SELECT * FROM "system"
WHERE organization_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: UpdateSystem :one
UPDATE "system"
SET "name" = $2, description = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteSystem :exec
UPDATE "system"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
