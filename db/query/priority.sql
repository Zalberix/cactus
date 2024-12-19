-- name: GetPriorityBySlug :one
SELECT * FROM priority p
WHERE p.slug = $1
LIMIT 1;

-- name: GetPriorityBySystemId :one
SELECT * FROM priority p 
JOIN system s ON s.id_priority = p.id
WHERE s.id = $1;