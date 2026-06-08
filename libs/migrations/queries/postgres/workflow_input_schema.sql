-- name: CreateWorkflowInputSchema :one
INSERT INTO "workflow_input_schema" (workflow_id, code, version_number, schema_json, status, is_default, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetWorkflowInputSchemaByID :one
SELECT * FROM "workflow_input_schema"
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetWorkflowInputSchemaByCode :one
SELECT * FROM "workflow_input_schema"
WHERE workflow_id = $1 AND code = $2 AND deleted_at IS NULL;

-- name: GetWorkflowInputSchemaByVersionNumber :one
SELECT * FROM "workflow_input_schema"
WHERE workflow_id = $1
  AND version_number = $2
  AND deleted_at IS NULL;

-- name: GetDefaultWorkflowInputSchema :one
SELECT * FROM "workflow_input_schema"
WHERE workflow_id = $1 AND is_default = TRUE AND deleted_at IS NULL;

-- name: ListWorkflowInputSchemasByWorkflowID :many
SELECT * FROM "workflow_input_schema"
WHERE workflow_id = $1 AND deleted_at IS NULL
ORDER BY is_default DESC, version_number DESC, id DESC;

-- name: SearchWorkflowInputSchemas :many
SELECT * FROM "workflow_input_schema"
WHERE workflow_id = $1
  AND deleted_at IS NULL
  AND status != 'archived'
  AND ($2::int <= 0 OR id != $2)
  AND (
    $3::text = ''
    OR code ILIKE '%' || $3 || '%'
    OR ('v' || version_number::text) ILIKE '%' || $3 || '%'
  )
ORDER BY is_default DESC, version_number DESC, id DESC
LIMIT 20;

-- name: GetNextWorkflowInputSchemaVersionNumber :one
SELECT COALESCE(MAX(version_number), 0)::int + 1 FROM "workflow_input_schema"
WHERE workflow_id = $1 AND deleted_at IS NULL;

-- name: UpdateWorkflowInputSchemaRecord :one
UPDATE "workflow_input_schema"
SET code = $2,
    version_number = $3,
    schema_json = $4,
    updated_by_user_id = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWorkflowInputSchemaStatus :one
UPDATE "workflow_input_schema"
SET status = $2,
    updated_by_user_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: ClearDefaultWorkflowInputSchema :exec
UPDATE "workflow_input_schema"
SET is_default = FALSE,
    updated_by_user_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE workflow_id = $1 AND is_default = TRUE AND deleted_at IS NULL;

-- name: SetDefaultWorkflowInputSchema :one
UPDATE "workflow_input_schema"
SET is_default = TRUE,
    updated_by_user_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status = 'active' AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflowInputSchema :exec
UPDATE "workflow_input_schema"
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: ArchiveWorkflowInputSchema :one
UPDATE "workflow_input_schema"
SET status = 'archived',
    is_default = FALSE,
    updated_by_user_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: HasInputSchemaUsage :one
SELECT EXISTS (
    SELECT 1 FROM "message" m
    WHERE m.workflow_input_schema_id = $1 AND m.deleted_at IS NULL
    UNION ALL
    SELECT 1 FROM "workflow_experiment_scope" s
    WHERE s.workflow_input_schema_id = $1 AND s.deleted_at IS NULL
) AS has_usage;

-- name: GetInputSchemaUsageSummary :one
WITH target_native_versions AS (
    SELECT c.workflow_version_id
    FROM "workflow_version_input_schema_compatibility" c
    WHERE c.workflow_input_schema_id = $1
      AND c.compatibility_type = 'native'
      AND c.deleted_at IS NULL
), mapper_source_usage AS (
    SELECT 1
    FROM "workflow_version_input_schema_compatibility" c
    WHERE c.workflow_input_schema_id = $1
      AND c.workflow_input_mapper_id IS NOT NULL
      AND c.compatibility_type != 'native'
      AND c.deleted_at IS NULL
    LIMIT 1
), mapper_target_usage AS (
    SELECT 1
    FROM "workflow_version_input_schema_compatibility" c
    JOIN target_native_versions tv ON tv.workflow_version_id = c.workflow_version_id
    WHERE c.workflow_input_mapper_id IS NOT NULL
      AND c.compatibility_type != 'native'
      AND c.deleted_at IS NULL
    LIMIT 1
)
SELECT
    EXISTS (
        SELECT 1 FROM "message" m
        WHERE m.workflow_input_schema_id = $1
          AND m.deleted_at IS NULL
    ) AS used_by_message,
    EXISTS (SELECT 1 FROM mapper_source_usage)
        OR EXISTS (SELECT 1 FROM mapper_target_usage) AS used_by_mapper,
    EXISTS (
        SELECT 1 FROM "workflow_experiment_scope" s
        JOIN "workflow_experiment" e ON e.id = s.workflow_experiment_id
        WHERE s.workflow_input_schema_id = $1
          AND s.deleted_at IS NULL
          AND e.deleted_at IS NULL
    ) AS used_by_experiment;

-- name: CreateWorkflowConfigurationAuditLog :one
INSERT INTO "workflow_configuration_audit_log" (entity_type, entity_id, action, actor_user_id, before_value, after_value)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
