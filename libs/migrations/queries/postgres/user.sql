-- name: CreateUser :one
INSERT INTO "user" (
    last_name,
    first_name,
    patronymic,
    email,
    "password",
    reset_password_after_login
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM "user"
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM "user"
WHERE email = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListUsersByOrgID :many
SELECT DISTINCT u.* FROM "user" u
JOIN "role_user" ru ON ru.user_id = u.id
JOIN "role" r ON r.id = ru.role_id
WHERE r.organization_id = $1 AND u.deleted_at IS NULL
ORDER BY u.last_name, u.first_name
LIMIT $2 OFFSET $3;

-- name: CountUsersByOrgID :one
SELECT COUNT(DISTINCT u.id)::bigint FROM "user" u
JOIN "role_user" ru ON ru.user_id = u.id
JOIN "role" r ON r.id = ru.role_id
WHERE r.organization_id = $1 AND u.deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE "user"
SET last_name = $2, first_name = $3, patronymic = $4, email = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :exec
UPDATE "user" SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
