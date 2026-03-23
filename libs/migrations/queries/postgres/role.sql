-- name: CreateRole :one
INSERT INTO "role" (organization_id, "name", description, is_system)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM "role" WHERE id = $1;

-- name: ListRolesByOrgID :many
SELECT * FROM "role"
WHERE organization_id = $1
ORDER BY "name";

-- name: UpdateRole :one
UPDATE "role"
SET "name" = $2, description = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM "role" WHERE id = $1;
