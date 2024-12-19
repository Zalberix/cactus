-- name: GetSystemIdByToken :one
SELECT s.id
FROM system s
INNER JOIN "token" t ON t.id_system = s.id
WHERE public_token = $1
LIMIT 1;

