-- name: CreateWorkTypeToken :one
INSERT INTO "work_type_token" (work_type_id, token_hash, is_active)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetWorkTypeTokenByWorkTypeID :one
SELECT * FROM "work_type_token"
WHERE work_type_id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetActiveWorkTypeTokenByHash :one
SELECT * FROM "work_type_token"
WHERE token_hash = $1 AND is_active = TRUE AND deleted_at IS NULL
LIMIT 1;

-- name: DeactivateWorkTypeToken :exec
UPDATE "work_type_token"
SET is_active = FALSE
WHERE id = $1;

-- name: SoftDeleteWorkTypeToken :exec
UPDATE "work_type_token"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
