-- name: CreateWorkflowRunStepAttempt :one
INSERT INTO "workflow_run_step_attempt" (workflow_run_step_id, worker_id, attempt_number, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWorkflowRunStepAttemptByID :one
SELECT * FROM "workflow_run_step_attempt"
WHERE id = $1;

-- name: ListAttemptsByRunStepID :many
SELECT * FROM "workflow_run_step_attempt"
WHERE workflow_run_step_id = $1
ORDER BY attempt_number;

-- name: UpdateWorkflowRunStepAttemptStatus :one
UPDATE "workflow_run_step_attempt"
SET status = $2, error_code = $3, error_message = $4, output_data = $5, completed_at = $6
WHERE id = $1
RETURNING *;
