-- name: CreateWorkerNATSSession :one
INSERT INTO "worker_nats_session" (
    worker_id, bootstrap_token_id, nats_account_public_key,
    nats_user_public_key, nats_user_jwt, permissions
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListActiveWorkerNATSSessionsByBootstrapToken :many
SELECT *
FROM "worker_nats_session"
WHERE bootstrap_token_id = $1
  AND revoked_at IS NULL
ORDER BY created_at DESC, id DESC;

-- name: RevokeWorkerNATSSession :one
UPDATE "worker_nats_session"
SET revoked_at = CURRENT_TIMESTAMP,
    revoked_by_user_id = $2
WHERE id = $1
  AND revoked_at IS NULL
RETURNING *;

-- name: RevokeWorkerNATSSessionsByBootstrapToken :many
UPDATE "worker_nats_session"
SET revoked_at = CURRENT_TIMESTAMP,
    revoked_by_user_id = $2
WHERE bootstrap_token_id = $1
  AND revoked_at IS NULL
RETURNING *;
