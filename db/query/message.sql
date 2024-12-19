-- name: CreateMessage :one
INSERT INTO message (
    id_worker,
    id_type_worker,
    id_system,
    "uuid",
    value,
    id_priority,
    send_later
) 
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetStatusMessageByUUID :one
WITH min_step AS (
    SELECT MIN(step) AS min_step
    FROM pipline
    WHERE time_end IS NULL
)
SELECT ps."name" as status FROM pipline_status ps 
JOIN pipline p ON p.id_pipline_status = ps.id
JOIN message m on m.id = p.id_message
JOIN min_step ms ON p.step = ms.min_step
WHERE m."uuid" = $1;

-- name: GetMessagesBy :many
SELECT * FROM message m 
WHERE m.id_type_worker = $1 AND m.id_system = $2;