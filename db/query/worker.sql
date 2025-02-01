-- name: CreateWorker :one
INSERT INTO worker ("uuid", is_active, id_type_worker, id_kind_worker) 
VALUES($1, $2, $3, $4)
RETURNING *;

-- name: GetWorkerByUUID :one
SELECT * 
FROM worker w
WHERE w.uuid = $1
LIMIT 1;
