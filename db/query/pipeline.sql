-- name: CreatePipeline :one
INSERT INTO pipeline (
    message_id,
    parent_pipeline_id
) VALUES ($1, $2)
RETURNING *;

-- name: GetPipelinesByMessageID :many
SELECT * FROM pipeline
WHERE message_id = $1 AND deleted_at IS NULL;

-- name: GetPipelineByID :one
SELECT * FROM pipeline
WHERE id = $1 AND deleted_at IS NULL;
