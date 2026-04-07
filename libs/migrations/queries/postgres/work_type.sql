-- name: CreateWorkType :one
INSERT INTO "work_type" ("name", code, description, meta)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWorkTypeByID :one
SELECT * FROM "work_type"
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetWorkTypeByCode :one
SELECT * FROM "work_type"
WHERE code = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListWorkTypes :many
SELECT * FROM "work_type"
WHERE deleted_at IS NULL
ORDER BY id;

-- name: SoftDeleteWorkType :exec
UPDATE "work_type"
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
