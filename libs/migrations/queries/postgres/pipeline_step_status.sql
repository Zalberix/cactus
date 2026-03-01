-- name: CreatePipelineStepStatus :one
INSERT INTO pipeline_step_status ("name", slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetPipelineStepStatusBySlug :one
SELECT * FROM pipeline_step_status
WHERE slug = $1
LIMIT 1;
