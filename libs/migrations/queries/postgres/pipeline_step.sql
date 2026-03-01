-- name: CreatePipelineStep :one
INSERT INTO pipeline_step (
    pipeline_id,
    worker_id,
    channel_id,
    step,
    time_start,
    time_end,
    pipeline_step_status_id
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPipelineStepsByPipelineID :many
SELECT * FROM pipeline_step
WHERE pipeline_id = $1 AND deleted_at IS NULL
ORDER BY step;

-- name: UpdatePipelineStepStatusAndWorker :one
UPDATE pipeline_step
SET pipeline_step_status_id = $2, worker_id = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetPipelineStepByPipelineIDAndStep :one
SELECT * FROM pipeline_step
WHERE pipeline_id = $1 AND step = $2 AND deleted_at IS NULL;
