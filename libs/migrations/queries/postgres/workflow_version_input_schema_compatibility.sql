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
SELECT c.*
FROM "workflow_version_input_schema_compatibility" c
JOIN "workflow_version_input_schema_compatibility" target_native
    ON target_native.workflow_version_id = c.workflow_version_id
    AND target_native.compatibility_type = 'native'
    AND target_native.is_active = TRUE
    AND target_native.deleted_at IS NULL
JOIN "workflow_input_schema" target_native_schema
    ON target_native_schema.id = target_native.workflow_input_schema_id
    AND target_native_schema.deleted_at IS NULL
    AND target_native_schema.status != 'archived'
WHERE c.workflow_version_id = $1
  AND c.workflow_input_schema_id = $2
  AND c.is_active = TRUE
  AND c.deleted_at IS NULL;

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

-- name: ListInputSchemaCompatibilityRows :many
SELECT
    c.id,
    c.workflow_version_id,
    COALESCE(wv.name, 'Version ' || wv.version_number::text) AS workflow_version_name,
    wv.version_number AS workflow_version_number,
    c.workflow_input_schema_id,
    c.compatibility_type,
    c.workflow_input_mapper_id,
    c.is_active,
    c.is_default_route
FROM "workflow_version_input_schema_compatibility" c
JOIN "workflow_version" wv ON wv.id = c.workflow_version_id
WHERE c.workflow_input_schema_id = $1
  AND c.deleted_at IS NULL
  AND wv.deleted_at IS NULL
  AND wv.archived_at IS NULL
ORDER BY c.is_default_route DESC, wv.version_number DESC, c.id DESC;

-- name: ListWorkflowRoutingVersionRows :many
WITH native_schema AS (
    SELECT
        c.workflow_version_id,
        wis.id AS native_schema_id,
        wis.code AS native_schema_code,
        wis.version_number AS native_schema_version_number,
        wis.status AS native_schema_status
    FROM "workflow_version_input_schema_compatibility" c
    JOIN "workflow_input_schema" wis ON wis.id = c.workflow_input_schema_id
    WHERE c.compatibility_type = 'native'
      AND c.is_active = TRUE
      AND c.deleted_at IS NULL
      AND wis.deleted_at IS NULL
      AND wis.status != 'archived'
)
SELECT
    wv.id AS workflow_version_id,
    wv.workflow_id,
    COALESCE(wv.name, 'Version ' || wv.version_number::text) AS workflow_version_name,
    wv.version_number AS workflow_version_number,
    wv.is_valid,
    wv.is_active,
    ns.native_schema_id,
    ns.native_schema_code,
    ns.native_schema_version_number,
    COALESCE(
        jsonb_agg(
            DISTINCT jsonb_build_object(
                'compatibility_id', c.id,
                'schema_id', wis.id,
                'schema_code', wis.code,
                'schema_version_number', wis.version_number,
                'compatibility_type', c.compatibility_type,
                'mapper_id', c.workflow_input_mapper_id,
                'support_mode', CASE
                    WHEN c.workflow_input_mapper_id IS NULL THEN 'native'
                    ELSE 'mapper'
                END,
                'is_default_route', c.is_default_route
            )
        ) FILTER (WHERE c.id IS NOT NULL AND wis.id IS NOT NULL),
        '[]'::jsonb
    ) AS supported_schemas
FROM "workflow_version" wv
JOIN native_schema ns ON ns.workflow_version_id = wv.id
LEFT JOIN "workflow_version_input_schema_compatibility" c
    ON c.workflow_version_id = wv.id
    AND c.is_active = TRUE
    AND c.deleted_at IS NULL
LEFT JOIN "workflow_input_schema" wis
    ON wis.id = c.workflow_input_schema_id
    AND wis.deleted_at IS NULL
    AND wis.status != 'archived'
WHERE wv.workflow_id = $1
  AND wv.deleted_at IS NULL
  AND wv.archived_at IS NULL
GROUP BY wv.id, ns.native_schema_id, ns.native_schema_code, ns.native_schema_version_number
ORDER BY wv.is_active DESC, wv.version_number DESC, wv.id DESC;

-- name: GetWorkflowVersionByNativeInputSchemaID :one
SELECT wv.*
FROM "workflow_version_input_schema_compatibility" c
JOIN "workflow_version" wv ON wv.id = c.workflow_version_id
WHERE c.workflow_input_schema_id = $1
  AND c.compatibility_type = 'native'
  AND c.is_active = TRUE
  AND c.deleted_at IS NULL
  AND wv.deleted_at IS NULL
  AND wv.archived_at IS NULL
LIMIT 1;

-- name: DeactivateCompatibilitiesByInputSchemaID :exec
UPDATE "workflow_version_input_schema_compatibility"
SET is_active = FALSE,
    is_default_route = FALSE,
    updated_by_user_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE workflow_input_schema_id = $1
  AND deleted_at IS NULL;

-- name: DeactivateMappedCompatibilitiesByNativeTargetSchemaID :exec
WITH target_versions AS (
    SELECT native.workflow_version_id
    FROM "workflow_version_input_schema_compatibility" native
    WHERE native.workflow_input_schema_id = $1
      AND native.compatibility_type = 'native'
      AND native.deleted_at IS NULL
)
UPDATE "workflow_version_input_schema_compatibility" c
SET is_active = FALSE,
    is_default_route = FALSE,
    updated_by_user_id = $2,
    updated_at = CURRENT_TIMESTAMP
FROM target_versions tv
WHERE c.workflow_version_id = tv.workflow_version_id
  AND c.compatibility_type != 'native'
  AND c.deleted_at IS NULL;

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
JOIN "workflow_version_input_schema_compatibility" target_native
    ON target_native.workflow_version_id = v.id
    AND target_native.compatibility_type = 'native'
    AND target_native.is_active = TRUE
    AND target_native.deleted_at IS NULL
JOIN "workflow_input_schema" target_native_schema
    ON target_native_schema.id = target_native.workflow_input_schema_id
    AND target_native_schema.deleted_at IS NULL
    AND target_native_schema.status != 'archived'
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
JOIN "workflow_version_input_schema_compatibility" target_native
    ON target_native.workflow_version_id = v.id
    AND target_native.compatibility_type = 'native'
    AND target_native.is_active = TRUE
    AND target_native.deleted_at IS NULL
JOIN "workflow_input_schema" target_native_schema
    ON target_native_schema.id = target_native.workflow_input_schema_id
    AND target_native_schema.deleted_at IS NULL
    AND target_native_schema.status != 'archived'
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
