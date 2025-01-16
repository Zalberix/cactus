-- name: GetSystemById :one 
SELECT * 
FROM system s
WHERE s.id = $1;

-- name: CreateSystem :one
INSERT INTO "system" (
    create_user, 
    id_priority, 
    "name", 
    description, 
    is_active
) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: AddKindWorkerForSystem :exec
INSERT INTO kind_worker_system (
    id_system, 
    id_kind_worker
) VALUES($1, $2)
RETURNING *;
