-- name: GrantWorkflowToken :exec
INSERT INTO "workflow_token" (system_token_id, workflow_id)
VALUES ($1, $2)
ON CONFLICT (system_token_id, workflow_id) DO NOTHING;

-- name: RevokeWorkflowToken :exec
DELETE FROM "workflow_token"
WHERE system_token_id = $1 AND workflow_id = $2;

-- name: ListWorkflowTokensBySystemTokenID :many
SELECT * FROM "workflow_token"
WHERE system_token_id = $1
ORDER BY workflow_id;

-- name: ListWorkflowTokensByWorkflowID :many
SELECT * FROM "workflow_token"
WHERE workflow_id = $1
ORDER BY system_token_id;

-- name: CheckWorkflowAccess :one
SELECT COUNT(*) > 0 AS has_access
FROM "workflow_token" wt
JOIN "system_token" st ON st.id = wt.system_token_id
WHERE wt.workflow_id = $1
  AND st.public_token = $2
  AND st.is_active = TRUE
  AND st.deleted_at IS NULL;
