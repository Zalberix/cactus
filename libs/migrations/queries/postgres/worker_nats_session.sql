-- name: CreateWorkerNATSSession :one
INSERT INTO "worker_nats_session" (
    worker_id, bootstrap_token_id, nats_account_public_key,
    nats_user_public_key, nats_user_jwt, nats_user_seed,
    nats_user_credentials, permissions
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetActiveWorkerNATSSessionByWorkerAndBootstrapTokenForUpdate :one
SELECT *
FROM "worker_nats_session"
WHERE worker_id = $1
  AND bootstrap_token_id = $2
  AND revoked_at IS NULL
ORDER BY created_at DESC, id DESC
LIMIT 1
FOR UPDATE;

-- name: ReplaceLatestWorkerNATSSessionByWorkerAndBootstrapToken :one
UPDATE "worker_nats_session" AS s
SET nats_account_public_key = sqlc.arg(nats_account_public_key),
    nats_user_public_key = sqlc.arg(nats_user_public_key),
    nats_user_jwt = sqlc.arg(nats_user_jwt),
    nats_user_seed = sqlc.arg(nats_user_seed),
    nats_user_credentials = sqlc.arg(nats_user_credentials),
    permissions = sqlc.arg(permissions),
    revoked_at = NULL,
    revoked_by_user_id = NULL,
    created_at = CURRENT_TIMESTAMP
WHERE s.id = (
    SELECT latest.id
    FROM "worker_nats_session" AS latest
    WHERE latest.worker_id = sqlc.arg(worker_id)
      AND latest.bootstrap_token_id = sqlc.arg(bootstrap_token_id)
    ORDER BY latest.created_at DESC, latest.id DESC
    LIMIT 1
    FOR UPDATE
)
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
