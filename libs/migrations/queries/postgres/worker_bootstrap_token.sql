-- name: CreateWorkerBootstrapToken :one
INSERT INTO "worker_bootstrap_token" (
    organization_id, work_type_id, name, description, token_hash,
    max_active_workers, expires_at, created_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetActiveWorkerBootstrapTokenByHash :one
SELECT *
FROM "worker_bootstrap_token"
WHERE token_hash = $1
  AND status = 'active'
  AND deleted_at IS NULL
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
LIMIT 1;

-- name: GetActiveWorkerBootstrapTokenByHashForUpdate :one
SELECT *
FROM "worker_bootstrap_token"
WHERE token_hash = $1
  AND status = 'active'
  AND deleted_at IS NULL
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
FOR UPDATE;

-- name: ListWorkerBootstrapTokensByOrganization :many
SELECT
    t.*,
    COALESCE(active_workers.active_worker_count, 0)::int AS active_worker_count
FROM "worker_bootstrap_token" t
LEFT JOIN LATERAL (
    SELECT COUNT(DISTINCT s.worker_id) AS active_worker_count
    FROM "worker_nats_session" s
    JOIN "worker" w ON w.id = s.worker_id
    WHERE s.bootstrap_token_id = t.id
      AND s.revoked_at IS NULL
      AND w.last_heartbeat_at > CURRENT_TIMESTAMP - INTERVAL '90 seconds'
) active_workers ON TRUE
WHERE t.organization_id = $1
  AND t.deleted_at IS NULL
ORDER BY t.created_at DESC, t.id DESC;

-- name: CountActiveWorkersByBootstrapTokenExcludingWorker :one
SELECT COUNT(DISTINCT s.worker_id)::int
FROM "worker_nats_session" s
JOIN "worker" w ON w.id = s.worker_id
WHERE s.bootstrap_token_id = sqlc.arg(bootstrap_token_id)
  AND s.revoked_at IS NULL
  AND w.last_heartbeat_at > CURRENT_TIMESTAMP - (sqlc.arg(heartbeat_timeout_seconds)::int * INTERVAL '1 second')
  AND (sqlc.arg(exclude_worker_id)::int = 0 OR s.worker_id <> sqlc.arg(exclude_worker_id)::int);

-- name: TouchWorkerBootstrapTokenUse :one
UPDATE "worker_bootstrap_token"
SET total_registration_count = total_registration_count + 1,
    last_used_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND status = 'active'
  AND deleted_at IS NULL
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
RETURNING *;

-- name: RevokeWorkerBootstrapToken :one
UPDATE "worker_bootstrap_token"
SET status = 'revoked',
    revoked_at = CURRENT_TIMESTAMP,
    revoked_by_user_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
