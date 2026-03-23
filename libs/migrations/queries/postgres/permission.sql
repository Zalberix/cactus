-- name: GetPermissionBySlug :one
SELECT * FROM "permission" WHERE slug = $1;

-- name: ListPermissions :many
SELECT * FROM "permission" ORDER BY slug;

-- name: CreatePermission :one
INSERT INTO "permission" (slug, is_private)
VALUES ($1, $2)
RETURNING *;
