-- name: CreateWorkflowRunStep :one
INSERT INTO "workflow_run_step" (workflow_run_id, workflow_step_id, worker_id, temporal_step_id, status, input_data)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetWorkflowRunStepByID :one
SELECT * FROM "workflow_run_step"
WHERE id = $1;

-- name: ListWorkflowRunStepsByRunID :many
SELECT * FROM "workflow_run_step"
WHERE workflow_run_id = $1
ORDER BY id;

-- name: UpdateWorkflowRunStepStatus :one
UPDATE "workflow_run_step"
SET status = $2, outcome = $3, output_data = $4, completed_at = $5, error_message = $6, started_at = COALESCE(started_at, $7)
WHERE id = $1
RETURNING *;

-- name: UpdateWorkflowRunStepStarted :one
UPDATE "workflow_run_step"
SET status = 'running', worker_id = $2, input_data = $3, started_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListWorkflowRunStepDetailsByRunID :many
SELECT
    wrs.id,
    wrs.workflow_run_id,
    wrs.workflow_step_id,
    wrs.worker_id,
    wrs.temporal_step_id,
    wrs.status,
    wrs.outcome,
    wrs.input_data,
    wrs.output_data,
    wrs.started_at,
    wrs.completed_at,
    wrs.error_message
FROM "workflow_run_step" wrs
WHERE wrs.workflow_run_id = $1
ORDER BY wrs.workflow_step_id, wrs.id;

-- name: GetWorkflowRunStepByRunAndStepID :one
SELECT *
FROM "workflow_run_step"
WHERE workflow_run_id = $1 AND workflow_step_id = $2
ORDER BY id DESC
LIMIT 1;

-- name: CountRunningStepsByRevisionInWindow :one
SELECT COUNT(*) AS cnt
FROM "workflow_run_step" wrs
JOIN "workflow_step" ws ON ws.id = wrs.workflow_step_id
WHERE ws.worker_settings_revision_id = $1
  AND wrs.started_at >= $2
  AND wrs.status IN ('running', 'completed');

-- name: ListWorkflowRunStepStatusesByRunID :many
-- Status API: step statuses WITHOUT input/output data (per D-21 — may be large)
SELECT wrs.id, wrs.workflow_step_id, wrs.status, wrs.outcome,
       wrs.started_at, wrs.completed_at, wrs.error_message,
       ws.step_type
FROM "workflow_run_step" wrs
JOIN "workflow_step" ws ON ws.id = wrs.workflow_step_id
WHERE wrs.workflow_run_id = $1
ORDER BY wrs.id;
