-- name: CreatePipelineStep :one
INSERT INTO pipeline (
    id_message, 
    status,
    step, 
    "name", 
    time_start, 
    time_end
) VALUES ($1, $2, $3, $4, $5, $6) 
RETURNING *;