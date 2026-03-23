-- name: CreateOrganization :one
INSERT INTO "organization" ("name", code)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM "organization"
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetOrganizationByCode :one
SELECT * FROM "organization"
WHERE code = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListOrganizations :many
SELECT * FROM "organization"
WHERE deleted_at IS NULL
ORDER BY id;

-- name: UpdateOrganization :one
UPDATE "organization"
SET "name" = $2, code = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteOrganization :exec
UPDATE "organization"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
