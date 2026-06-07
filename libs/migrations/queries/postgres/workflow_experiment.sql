-- name: CreateWorkflowExperiment :one
INSERT INTO "workflow_experiment" (workflow_id, name, description, experiment_type, status, started_at, ended_at, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetWorkflowExperimentByID :one
SELECT * FROM "workflow_experiment"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowExperimentsByWorkflowID :many
SELECT * FROM "workflow_experiment"
WHERE workflow_id = $1 AND deleted_at IS NULL
ORDER BY status = 'active' DESC, created_at DESC, id DESC;

-- name: UpdateWorkflowExperiment :one
UPDATE "workflow_experiment"
SET name = $2,
    description = $3,
    experiment_type = $4,
    started_at = $5,
    ended_at = $6,
    updated_by_user_id = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWorkflowExperimentStatus :one
UPDATE "workflow_experiment"
SET status = $2,
    started_at = COALESCE($3, started_at),
    ended_at = COALESCE($4, ended_at),
    updated_by_user_id = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflowExperiment :exec
UPDATE "workflow_experiment"
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateWorkflowExperimentScope :one
INSERT INTO "workflow_experiment_scope" (workflow_experiment_id, workflow_input_schema_id, traffic_conditions, conditions_hash, traffic_percent, fallback_policy, fallback_workflow_version_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetWorkflowExperimentScopeByID :one
SELECT * FROM "workflow_experiment_scope"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowExperimentScopesByExperimentID :many
SELECT * FROM "workflow_experiment_scope"
WHERE workflow_experiment_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: UpdateWorkflowExperimentScope :one
UPDATE "workflow_experiment_scope"
SET workflow_input_schema_id = $2,
    traffic_conditions = $3,
    conditions_hash = $4,
    traffic_percent = $5,
    fallback_policy = $6,
    fallback_workflow_version_id = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflowExperimentScope :exec
UPDATE "workflow_experiment_scope"
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListActiveExperimentScopesForRouting :many
SELECT
    e.id AS workflow_experiment_id,
    e.workflow_id,
    e.experiment_type,
    s.id AS workflow_experiment_scope_id,
    s.workflow_input_schema_id,
    s.traffic_conditions,
    s.conditions_hash,
    s.traffic_percent,
    s.fallback_policy,
    s.fallback_workflow_version_id
FROM "workflow_experiment" e
JOIN "workflow_experiment_scope" s ON s.workflow_experiment_id = e.id AND s.deleted_at IS NULL
WHERE e.workflow_id = $1
  AND e.status = 'active'
  AND e.deleted_at IS NULL
  AND (e.started_at IS NULL OR e.started_at <= CURRENT_TIMESTAMP)
  AND (e.ended_at IS NULL OR e.ended_at > CURRENT_TIMESTAMP)
  AND s.workflow_input_schema_id = $2
  AND (sqlc.narg(experiment_id)::int IS NULL OR e.id = sqlc.narg(experiment_id)::int)
ORDER BY e.id DESC, s.id DESC;

-- name: CreateWorkflowExperimentVariant :one
INSERT INTO "workflow_experiment_variant" (workflow_experiment_scope_id, workflow_version_id, traffic_weight, is_control_group, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetWorkflowExperimentVariantByID :one
SELECT * FROM "workflow_experiment_variant"
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkflowExperimentVariantsByScopeID :many
SELECT * FROM "workflow_experiment_variant"
WHERE workflow_experiment_scope_id = $1 AND deleted_at IS NULL
ORDER BY is_control_group DESC, workflow_version_id, id;

-- name: ListActiveWorkflowExperimentVariantsByScopeID :many
SELECT * FROM "workflow_experiment_variant"
WHERE workflow_experiment_scope_id = $1
  AND is_active = TRUE
  AND deleted_at IS NULL
ORDER BY is_control_group DESC, workflow_version_id, id;

-- name: UpdateWorkflowExperimentVariant :one
UPDATE "workflow_experiment_variant"
SET workflow_version_id = $2,
    traffic_weight = $3,
    is_control_group = $4,
    is_active = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteWorkflowExperimentVariant :exec
UPDATE "workflow_experiment_variant"
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
