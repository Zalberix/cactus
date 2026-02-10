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
