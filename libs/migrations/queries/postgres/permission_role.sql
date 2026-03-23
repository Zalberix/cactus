-- name: AddPermissionToRole :exec
INSERT INTO "permission_role" (permission_id, role_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemovePermissionFromRole :exec
DELETE FROM "permission_role"
WHERE permission_id = $1 AND role_id = $2;

-- name: ListPermissionsByRoleID :many
SELECT p.* FROM "permission" p
JOIN "permission_role" pr ON pr.permission_id = p.id
WHERE pr.role_id = $1
ORDER BY p.slug;

-- name: ListPermissionsByUserID :many
SELECT DISTINCT p.* FROM "permission" p
JOIN "permission_role" pr ON pr.permission_id = p.id
JOIN "role_user" ru ON ru.role_id = pr.role_id
WHERE ru.user_id = $1
ORDER BY p.slug;
