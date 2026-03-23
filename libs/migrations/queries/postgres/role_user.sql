-- name: AssignRoleToUser :exec
INSERT INTO "role_user" (role_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE FROM "role_user"
WHERE role_id = $1 AND user_id = $2;

-- name: ListRolesByUserID :many
SELECT r.* FROM "role" r
JOIN "role_user" ru ON ru.role_id = r.id
WHERE ru.user_id = $1
ORDER BY r."name";

-- name: ListUsersByRoleID :many
SELECT u.* FROM "user" u
JOIN "role_user" ru ON ru.user_id = u.id
WHERE ru.role_id = $1 AND u.deleted_at IS NULL
ORDER BY u.last_name, u.first_name;
