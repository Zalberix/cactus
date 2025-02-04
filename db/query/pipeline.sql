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

-- name: GetAllPipelineByMessageUUID :many
SELECT p.* 
FROM pipeline p 
JOIN message m ON m.id = p.id_message
WHERE m."uuid" = $1;

-- name: UpdatePipelineStatusAndWorkerByID :one
UPDATE pipeline 
SET status = $2, id_worker = $3
WHERE id = $1
RETURNING *;

-- name: GetIdPipelineByUUIDMessageAndStep :one
SELECT p.id FROM pipeline p
JOIN message m ON m.id = p.id_message
WHERE m."uuid" = $1 AND step = $2;