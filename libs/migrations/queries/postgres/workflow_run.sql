-- name: CreateWorkflowRun :one
INSERT INTO "workflow_run" (workflow_version_id, message_id, temporal_workflow_id, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWorkflowRunByID :one
SELECT * FROM "workflow_run"
WHERE id = $1;

-- name: GetWorkflowRunByTemporalID :one
SELECT * FROM "workflow_run"
WHERE temporal_workflow_id = $1
LIMIT 1;

-- name: ListWorkflowRunsByMessageID :many
SELECT * FROM "workflow_run"
WHERE message_id = $1
ORDER BY id;

-- name: UpdateWorkflowRunStatus :one
UPDATE "workflow_run"
SET status = $2, completed_at = $3, error_message = $4
WHERE id = $1
RETURNING *;

-- name: UpdateWorkflowRunStarted :one
UPDATE "workflow_run"
SET status = 'running', started_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
