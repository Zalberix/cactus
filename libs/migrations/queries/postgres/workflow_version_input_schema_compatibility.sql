-- name: CreateWorkflowVersionInputSchemaCompatibility :one
INSERT INTO "workflow_version_input_schema_compatibility" (
    workflow_version_id,
    workflow_input_schema_id,
    compatibility_type,
    workflow_input_mapper_id,
    default_values,
    is_active,
    is_default_route,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetWorkflowVersionInputSchemaCompatibilityByID :one
SELECT * FROM "workflow_version_input_schema_compatibility"
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetActiveWorkflowVersionInputSchemaCompatibilityByPair :one
SELECT * FROM "workflow_version_input_schema_compatibility"
WHERE workflow_version_id = $1
  AND workflow_input_schema_id = $2
  AND is_active = TRUE
  AND deleted_at IS NULL;

-- name: GetNativeInputSchemaForWorkflowVersion :one
SELECT wis.*
FROM "workflow_version_input_schema_compatibility" c
JOIN "workflow_input_schema" wis ON wis.id = c.workflow_input_schema_id
WHERE c.workflow_version_id = $1
  AND c.compatibility_type = 'native'
  AND c.is_active = TRUE
  AND c.deleted_at IS NULL
  AND wis.deleted_at IS NULL
ORDER BY wis.version_number DESC, wis.id DESC
LIMIT 1;

-- name: ListCompatibilitiesByWorkflowID :many
SELECT c.*
FROM "workflow_version_input_schema_compatibility" c
JOIN "workflow_version" v ON v.id = c.workflow_version_id
WHERE v.workflow_id = $1
  AND c.deleted_at IS NULL
  AND v.deleted_at IS NULL
ORDER BY c.workflow_input_schema_id, v.version_number, c.id;

-- name: ListCompatibilitiesByVersionID :many
SELECT * FROM "workflow_version_input_schema_compatibility"
WHERE workflow_version_id = $1 AND deleted_at IS NULL
ORDER BY is_default_route DESC, workflow_input_schema_id, id;

-- name: ListCompatibilitiesByInputSchemaID :many
SELECT * FROM "workflow_version_input_schema_compatibility"
WHERE workflow_input_schema_id = $1 AND deleted_at IS NULL
ORDER BY is_default_route DESC, workflow_version_id, id;

-- name: ListActiveRoutingCompatibilitiesByInputSchemaID :many
SELECT
    c.id,
    c.workflow_version_id,
    c.workflow_input_schema_id,
    c.compatibility_type,
    c.workflow_input_mapper_id,
    c.default_values,
    c.is_default_route,
    v.workflow_id,
    v.version_number,
    v.name AS workflow_version_name,
    v.is_valid,
    v.is_active,
    v.archived_at
FROM "workflow_version_input_schema_compatibility" c
JOIN "workflow_version" v ON v.id = c.workflow_version_id
WHERE c.workflow_input_schema_id = $1
  AND c.is_active = TRUE
  AND c.deleted_at IS NULL
  AND v.deleted_at IS NULL
  AND v.archived_at IS NULL
  AND v.is_active = TRUE
  AND v.is_valid = TRUE
ORDER BY c.is_default_route DESC, v.version_number DESC, v.id DESC;

-- name: GetDefaultRouteForInputSchema :one
SELECT c.*
FROM "workflow_version_input_schema_compatibility" c
JOIN "workflow_version" v ON v.id = c.workflow_version_id
WHERE c.workflow_input_schema_id = $1
  AND c.is_default_route = TRUE
  AND c.is_active = TRUE
  AND c.deleted_at IS NULL
  AND v.deleted_at IS NULL
  AND v.archived_at IS NULL
  AND v.is_active = TRUE
  AND v.is_valid = TRUE;

-- name: UpdateWorkflowVersionInputSchemaCompatibility :one
UPDATE "workflow_version_input_schema_compatibility"
SET compatibility_type = $2,
    workflow_input_mapper_id = $3,
    default_values = $4,
    is_active = $5,
    is_default_route = $6,
    updated_by_user_id = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateCompatibilityDefaultRoute :one
UPDATE "workflow_version_input_schema_compatibility"
SET is_default_route = $2,
    updated_by_user_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeactivateWorkflowVersionInputSchemaCompatibility :one
UPDATE "workflow_version_input_schema_compatibility"
SET is_active = FALSE,
    is_default_route = FALSE,
    updated_by_user_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflowVersionInputSchemaCompatibility :exec
UPDATE "workflow_version_input_schema_compatibility"
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: HasCompatibilityHistoricalRun :one
SELECT EXISTS (
    SELECT 1
    FROM "workflow_version_input_schema_compatibility" c
    JOIN "workflow_run" wr ON wr.workflow_version_id = c.workflow_version_id
    JOIN "message" m ON m.id = wr.message_id AND m.workflow_input_schema_id = c.workflow_input_schema_id
    WHERE c.id = $1
) AS has_run;
