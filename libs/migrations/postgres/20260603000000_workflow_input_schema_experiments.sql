-- +goose Up

ALTER TABLE "workflow_version"
    ADD COLUMN locked_at TIMESTAMP NULL,
    ADD COLUMN published_at TIMESTAMP NULL,
    ADD COLUMN archived_at TIMESTAMP NULL,
    ADD COLUMN updated_by_user_id INT NULL;

ALTER TABLE "workflow_version"
    ADD CONSTRAINT workflow_version_updated_by_user_id_fkey
    FOREIGN KEY (updated_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX workflow_version_workflow_version_number_uq
    ON "workflow_version" (workflow_id, version_number)
    WHERE deleted_at IS NULL;

CREATE TABLE "workflow_input_schema" (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL,
    code VARCHAR(255) NOT NULL,
    version_number INT NOT NULL,
    schema_json JSONB NOT NULL,
    status VARCHAR(32) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_by_user_id INT NULL,
    updated_by_user_id INT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT workflow_input_schema_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES "workflow"(id),
    CONSTRAINT workflow_input_schema_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_input_schema_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_input_schema_status_check CHECK (status IN ('draft', 'active', 'deprecated', 'archived')),
    CONSTRAINT workflow_input_schema_default_active_check CHECK (is_default = FALSE OR status = 'active'),
    CONSTRAINT workflow_input_schema_json_object_check CHECK (jsonb_typeof(schema_json) = 'object')
);

CREATE UNIQUE INDEX workflow_input_schema_workflow_code_uq
    ON "workflow_input_schema" (workflow_id, code)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX workflow_input_schema_workflow_version_uq
    ON "workflow_input_schema" (workflow_id, version_number)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX workflow_input_schema_default_uq
    ON "workflow_input_schema" (workflow_id)
    WHERE is_default = TRUE AND deleted_at IS NULL;

CREATE INDEX workflow_input_schema_workflow_status_idx
    ON "workflow_input_schema" (workflow_id, status)
    WHERE deleted_at IS NULL;

CREATE TABLE "workflow_input_mapper" (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    mapper_type VARCHAR(32) NOT NULL,
    rules JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by_user_id INT NULL,
    updated_by_user_id INT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT workflow_input_mapper_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES "workflow"(id),
    CONSTRAINT workflow_input_mapper_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_input_mapper_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_input_mapper_type_check CHECK (mapper_type IN ('internal', 'jsonata', 'jq', 'javascript')),
    CONSTRAINT workflow_input_mapper_rules_object_check CHECK (jsonb_typeof(rules) = 'object')
);

CREATE UNIQUE INDEX workflow_input_mapper_workflow_name_uq
    ON "workflow_input_mapper" (workflow_id, name)
    WHERE deleted_at IS NULL;

CREATE INDEX workflow_input_mapper_workflow_active_idx
    ON "workflow_input_mapper" (workflow_id, is_active)
    WHERE deleted_at IS NULL;

CREATE TABLE "workflow_version_input_schema_compatibility" (
    id SERIAL PRIMARY KEY,
    workflow_version_id INT NOT NULL,
    workflow_input_schema_id INT NOT NULL,
    compatibility_type VARCHAR(32) NOT NULL,
    workflow_input_mapper_id INT NULL,
    default_values JSONB NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_default_route BOOLEAN NOT NULL DEFAULT FALSE,
    created_by_user_id INT NULL,
    updated_by_user_id INT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT workflow_version_input_schema_compatibility_version_fkey FOREIGN KEY (workflow_version_id) REFERENCES "workflow_version"(id),
    CONSTRAINT workflow_version_input_schema_compatibility_schema_fkey FOREIGN KEY (workflow_input_schema_id) REFERENCES "workflow_input_schema"(id),
    CONSTRAINT workflow_version_input_schema_compatibility_mapper_fkey FOREIGN KEY (workflow_input_mapper_id) REFERENCES "workflow_input_mapper"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_version_input_schema_compatibility_created_by_fkey FOREIGN KEY (created_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_version_input_schema_compatibility_updated_by_fkey FOREIGN KEY (updated_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_version_input_schema_compatibility_type_check CHECK (compatibility_type IN ('native', 'adapter', 'partial')),
    CONSTRAINT workflow_version_input_schema_compatibility_mapper_check CHECK (
        (compatibility_type = 'native' AND workflow_input_mapper_id IS NULL)
        OR
        (compatibility_type IN ('adapter', 'partial') AND (workflow_input_mapper_id IS NOT NULL OR default_values IS NOT NULL))
    )
);

CREATE UNIQUE INDEX workflow_version_input_schema_compatibility_pair_uq
    ON "workflow_version_input_schema_compatibility" (workflow_version_id, workflow_input_schema_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX workflow_version_input_schema_compatibility_default_route_uq
    ON "workflow_version_input_schema_compatibility" (workflow_input_schema_id)
    WHERE is_default_route = TRUE AND is_active = TRUE AND deleted_at IS NULL;

CREATE INDEX workflow_version_input_schema_compatibility_schema_active_idx
    ON "workflow_version_input_schema_compatibility" (workflow_input_schema_id, is_active)
    WHERE deleted_at IS NULL;

CREATE INDEX workflow_version_input_schema_compatibility_version_active_idx
    ON "workflow_version_input_schema_compatibility" (workflow_version_id, is_active)
    WHERE deleted_at IS NULL;

CREATE TABLE "workflow_experiment" (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    experiment_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    created_by_user_id INT NULL,
    updated_by_user_id INT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT workflow_experiment_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES "workflow"(id),
    CONSTRAINT workflow_experiment_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_experiment_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_experiment_type_check CHECK (experiment_type IN ('canary', 'experiment', 'rollout')),
    CONSTRAINT workflow_experiment_status_check CHECK (status IN ('draft', 'active', 'paused', 'completed', 'cancelled')),
    CONSTRAINT workflow_experiment_window_check CHECK (ended_at IS NULL OR started_at IS NULL OR ended_at >= started_at)
);

CREATE INDEX workflow_experiment_workflow_status_idx
    ON "workflow_experiment" (workflow_id, status)
    WHERE deleted_at IS NULL;

CREATE TABLE "workflow_experiment_scope" (
    id SERIAL PRIMARY KEY,
    workflow_experiment_id INT NOT NULL,
    workflow_input_schema_id INT NOT NULL,
    traffic_conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    conditions_hash VARCHAR(64) NOT NULL,
    traffic_percent INT NOT NULL DEFAULT 100,
    fallback_policy VARCHAR(32) NOT NULL DEFAULT 'error',
    fallback_workflow_version_id INT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT workflow_experiment_scope_experiment_fkey FOREIGN KEY (workflow_experiment_id) REFERENCES "workflow_experiment"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_experiment_scope_schema_fkey FOREIGN KEY (workflow_input_schema_id) REFERENCES "workflow_input_schema"(id),
    CONSTRAINT workflow_experiment_scope_fallback_version_fkey FOREIGN KEY (fallback_workflow_version_id) REFERENCES "workflow_version"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_experiment_scope_conditions_object_check CHECK (jsonb_typeof(traffic_conditions) = 'object'),
    CONSTRAINT workflow_experiment_scope_percent_check CHECK (traffic_percent BETWEEN 1 AND 100),
    CONSTRAINT workflow_experiment_scope_fallback_policy_check CHECK (fallback_policy IN ('error', 'default_route', 'explicit_version')),
    CONSTRAINT workflow_experiment_scope_explicit_fallback_check CHECK (fallback_policy != 'explicit_version' OR fallback_workflow_version_id IS NOT NULL),
    CONSTRAINT workflow_experiment_scope_hash_check CHECK (length(conditions_hash) = 64)
);

CREATE INDEX workflow_experiment_scope_schema_hash_idx
    ON "workflow_experiment_scope" (workflow_input_schema_id, conditions_hash)
    WHERE deleted_at IS NULL;

CREATE TABLE "workflow_experiment_variant" (
    id SERIAL PRIMARY KEY,
    workflow_experiment_scope_id INT NOT NULL,
    workflow_version_id INT NOT NULL,
    traffic_weight INT NOT NULL,
    is_control_group BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    CONSTRAINT workflow_experiment_variant_scope_fkey FOREIGN KEY (workflow_experiment_scope_id) REFERENCES "workflow_experiment_scope"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_experiment_variant_version_fkey FOREIGN KEY (workflow_version_id) REFERENCES "workflow_version"(id),
    CONSTRAINT workflow_experiment_variant_weight_check CHECK (traffic_weight BETWEEN 0 AND 100)
);

CREATE UNIQUE INDEX workflow_experiment_variant_scope_version_uq
    ON "workflow_experiment_variant" (workflow_experiment_scope_id, workflow_version_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX workflow_experiment_variant_control_group_uq
    ON "workflow_experiment_variant" (workflow_experiment_scope_id)
    WHERE is_control_group = TRUE AND is_active = TRUE AND deleted_at IS NULL;

CREATE INDEX workflow_experiment_variant_scope_active_idx
    ON "workflow_experiment_variant" (workflow_experiment_scope_id, is_active)
    WHERE deleted_at IS NULL;

CREATE TABLE "workflow_configuration_audit_log" (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(64) NOT NULL,
    entity_id INT NOT NULL,
    action VARCHAR(64) NOT NULL,
    actor_user_id INT NULL,
    before_value JSONB NULL,
    after_value JSONB NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT workflow_configuration_audit_log_actor_fkey FOREIGN KEY (actor_user_id) REFERENCES "user"(id) ON DELETE SET NULL
);

ALTER TABLE "message"
    ADD COLUMN workflow_input_schema_id INT NULL,
    ADD COLUMN idempotency_key VARCHAR(255) NULL,
    ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN error_message TEXT NULL;

ALTER TABLE "message"
    ADD CONSTRAINT message_workflow_input_schema_id_fkey
    FOREIGN KEY (workflow_input_schema_id) REFERENCES "workflow_input_schema"(id);

CREATE UNIQUE INDEX message_workflow_idempotency_key_uq
    ON "message" (workflow_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX message_workflow_input_schema_idx
    ON "message" (workflow_input_schema_id)
    WHERE deleted_at IS NULL;

ALTER TABLE "workflow_run"
    ADD COLUMN workflow_experiment_id INT NULL,
    ADD COLUMN workflow_experiment_scope_id INT NULL,
    ADD COLUMN workflow_experiment_variant_id INT NULL,
    ADD COLUMN input_schema_compatibility_id INT NULL,
    ADD COLUMN selection_reason VARCHAR(32) NOT NULL DEFAULT 'standard',
    ADD COLUMN version_input_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN routing_decision JSONB NULL;

ALTER TABLE "workflow_run"
    ADD CONSTRAINT workflow_run_experiment_id_fkey FOREIGN KEY (workflow_experiment_id) REFERENCES "workflow_experiment"(id) ON DELETE SET NULL,
    ADD CONSTRAINT workflow_run_experiment_scope_id_fkey FOREIGN KEY (workflow_experiment_scope_id) REFERENCES "workflow_experiment_scope"(id) ON DELETE SET NULL,
    ADD CONSTRAINT workflow_run_experiment_variant_id_fkey FOREIGN KEY (workflow_experiment_variant_id) REFERENCES "workflow_experiment_variant"(id) ON DELETE SET NULL,
    ADD CONSTRAINT workflow_run_input_schema_compatibility_id_fkey FOREIGN KEY (input_schema_compatibility_id) REFERENCES "workflow_version_input_schema_compatibility"(id) ON DELETE SET NULL,
    ADD CONSTRAINT workflow_run_selection_reason_check CHECK (selection_reason IN ('standard', 'canary', 'experiment', 'rollout', 'fallback')),
    ADD CONSTRAINT workflow_run_experiment_reason_check CHECK (
        selection_reason NOT IN ('canary', 'experiment', 'rollout')
        OR (workflow_experiment_id IS NOT NULL AND workflow_experiment_variant_id IS NOT NULL)
    );

CREATE INDEX workflow_run_routing_idx
    ON "workflow_run" (input_schema_compatibility_id, workflow_experiment_id, workflow_experiment_variant_id);

WITH ranked_versions AS (
    SELECT
        wv.id AS workflow_version_id,
        wv.workflow_id,
        wv.version_number,
        wv.is_active,
        wv.is_valid,
        w.input_schema,
        ROW_NUMBER() OVER (
            PARTITION BY wv.workflow_id
            ORDER BY wv.is_active DESC, wv.is_valid DESC, wv.version_number DESC, wv.id DESC
        ) AS default_rank
    FROM "workflow_version" wv
    JOIN "workflow" w ON w.id = wv.workflow_id AND w.deleted_at IS NULL
    WHERE wv.deleted_at IS NULL
)
INSERT INTO "workflow_input_schema" (
    workflow_id,
    code,
    version_number,
    schema_json,
    status,
    is_default,
    created_at,
    updated_at
)
SELECT
    workflow_id,
    'v' || version_number::text,
    version_number,
    COALESCE(input_schema, '{"type":"object","properties":{}}'::jsonb),
    CASE WHEN is_active THEN 'active' ELSE 'draft' END,
    default_rank = 1 AND is_active = TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM ranked_versions;

INSERT INTO "workflow_version_input_schema_compatibility" (
    workflow_version_id,
    workflow_input_schema_id,
    compatibility_type,
    workflow_input_mapper_id,
    default_values,
    is_active,
    is_default_route,
    created_at,
    updated_at
)
SELECT
    wv.id,
    wis.id,
    'native',
    NULL,
    '{}'::jsonb,
    TRUE,
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM "workflow_version" wv
JOIN "workflow_input_schema" wis
  ON wis.workflow_id = wv.workflow_id
 AND wis.version_number = wv.version_number
 AND wis.deleted_at IS NULL
WHERE wv.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM "workflow_version_input_schema_compatibility" c
      WHERE c.workflow_version_id = wv.id
        AND c.workflow_input_schema_id = wis.id
        AND c.deleted_at IS NULL
  );

WITH message_native_schema AS (
    SELECT DISTINCT ON (wr.message_id)
        wr.message_id,
        wis.id AS workflow_input_schema_id
    FROM "workflow_run" wr
    JOIN "workflow_version" wv ON wv.id = wr.workflow_version_id
    JOIN "workflow_input_schema" wis
      ON wis.workflow_id = wv.workflow_id
     AND wis.version_number = wv.version_number
     AND wis.deleted_at IS NULL
    ORDER BY wr.message_id, wr.id DESC
)
UPDATE "message" m
SET workflow_input_schema_id = mns.workflow_input_schema_id
FROM message_native_schema mns
WHERE mns.message_id = m.id
  AND m.workflow_input_schema_id IS NULL;

WITH fallback_schema AS (
    SELECT DISTINCT ON (workflow_id)
        workflow_id,
        id
    FROM "workflow_input_schema"
    WHERE deleted_at IS NULL
    ORDER BY workflow_id, is_default DESC, version_number DESC, id DESC
)
UPDATE "message" m
SET workflow_input_schema_id = fs.id
FROM fallback_schema fs
WHERE fs.workflow_id = m.workflow_id
  AND m.workflow_input_schema_id IS NULL;

WITH rollout_workflows AS (
    SELECT DISTINCT wv.workflow_id
    FROM "workflow_version" wv
    WHERE wv.is_active = TRUE
      AND wv.deleted_at IS NULL
      AND (wv.traffic_weight > 0 OR wv.is_control_group = TRUE)
), inserted_experiments AS (
    INSERT INTO "workflow_experiment" (workflow_id, name, description, experiment_type, status, started_at, created_at, updated_at)
    SELECT rw.workflow_id, 'Migrated Rollout', 'Backfilled from legacy workflow_version traffic weights', 'rollout', 'paused', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    FROM rollout_workflows rw
    RETURNING id, workflow_id
), inserted_scopes AS (
    INSERT INTO "workflow_experiment_scope" (workflow_experiment_id, workflow_input_schema_id, traffic_conditions, conditions_hash, traffic_percent, fallback_policy, created_at, updated_at)
    SELECT ie.id, wis.id, '{}'::jsonb, '44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a', 100, 'default_route', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    FROM inserted_experiments ie
    JOIN "workflow_input_schema" wis ON wis.workflow_id = ie.workflow_id AND wis.status = 'active' AND wis.deleted_at IS NULL
    RETURNING id, workflow_experiment_id, workflow_input_schema_id
)
INSERT INTO "workflow_experiment_variant" (workflow_experiment_scope_id, workflow_version_id, traffic_weight, is_control_group, is_active, created_at, updated_at)
SELECT iscope.id, wv.id, wv.traffic_weight, wv.is_control_group, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM inserted_scopes iscope
JOIN inserted_experiments ie ON ie.id = iscope.workflow_experiment_id
JOIN "workflow_version_input_schema_compatibility" compat
  ON compat.workflow_input_schema_id = iscope.workflow_input_schema_id
 AND compat.is_active = TRUE
 AND compat.deleted_at IS NULL
JOIN "workflow_version" wv
  ON wv.id = compat.workflow_version_id
 AND wv.workflow_id = ie.workflow_id
 AND wv.is_active = TRUE
 AND wv.deleted_at IS NULL;

UPDATE "workflow_run" wr
SET version_input_data = m.value
FROM "message" m
WHERE m.id = wr.message_id
  AND wr.version_input_data = '{}'::jsonb;

UPDATE "workflow_run" wr
SET input_schema_compatibility_id = compat.id
FROM "message" m
JOIN "workflow_version_input_schema_compatibility" compat
  ON compat.workflow_input_schema_id = m.workflow_input_schema_id
 AND compat.deleted_at IS NULL
WHERE m.id = wr.message_id
  AND compat.workflow_version_id = wr.workflow_version_id
  AND wr.input_schema_compatibility_id IS NULL;

ALTER TABLE "workflow" DROP COLUMN IF EXISTS input_schema;

ALTER TABLE "workflow_version"
    DROP COLUMN IF EXISTS traffic_weight,
    DROP COLUMN IF EXISTS traffic_updated_at,
    DROP COLUMN IF EXISTS is_control_group;

-- +goose Down

ALTER TABLE "workflow"
    ADD COLUMN IF NOT EXISTS input_schema JSONB NOT NULL DEFAULT '{"type":"object","properties":{}}'::jsonb;

WITH fallback_schema AS (
    SELECT DISTINCT ON (workflow_id)
        workflow_id,
        schema_json
    FROM "workflow_input_schema"
    WHERE deleted_at IS NULL
    ORDER BY workflow_id, is_default DESC, version_number DESC, id DESC
)
UPDATE "workflow" w
SET input_schema = fs.schema_json
FROM fallback_schema fs
WHERE fs.workflow_id = w.id;

ALTER TABLE "workflow_version"
    ADD COLUMN IF NOT EXISTS traffic_weight INT NOT NULL DEFAULT 100,
    ADD COLUMN IF NOT EXISTS traffic_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS is_control_group BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE "workflow_version" DROP CONSTRAINT IF EXISTS workflow_version_traffic_weight_check;
ALTER TABLE "workflow_version"
    ADD CONSTRAINT workflow_version_traffic_weight_check CHECK (traffic_weight >= 0 AND traffic_weight <= 100);

WITH variant_traffic AS (
    SELECT DISTINCT ON (ev.workflow_version_id)
        ev.workflow_version_id,
        ev.traffic_weight,
        ev.is_control_group
    FROM "workflow_experiment_variant" ev
    JOIN "workflow_experiment_scope" es ON es.id = ev.workflow_experiment_scope_id
    JOIN "workflow_experiment" e ON e.id = es.workflow_experiment_id
    WHERE e.name = 'Migrated Rollout'
      AND e.deleted_at IS NULL
      AND es.deleted_at IS NULL
      AND ev.deleted_at IS NULL
    ORDER BY ev.workflow_version_id, ev.updated_at DESC NULLS LAST, ev.id DESC
)
UPDATE "workflow_version" wv
SET traffic_weight = vt.traffic_weight,
    traffic_updated_at = CURRENT_TIMESTAMP,
    is_control_group = vt.is_control_group
FROM variant_traffic vt
WHERE vt.workflow_version_id = wv.id;

ALTER TABLE "workflow_run" DROP CONSTRAINT IF EXISTS workflow_run_experiment_id_fkey;
ALTER TABLE "workflow_run" DROP CONSTRAINT IF EXISTS workflow_run_experiment_scope_id_fkey;
ALTER TABLE "workflow_run" DROP CONSTRAINT IF EXISTS workflow_run_experiment_variant_id_fkey;
ALTER TABLE "workflow_run" DROP CONSTRAINT IF EXISTS workflow_run_input_schema_compatibility_id_fkey;
ALTER TABLE "workflow_run" DROP CONSTRAINT IF EXISTS workflow_run_selection_reason_check;
ALTER TABLE "workflow_run" DROP CONSTRAINT IF EXISTS workflow_run_experiment_reason_check;
DROP INDEX IF EXISTS workflow_run_routing_idx;
ALTER TABLE "workflow_run"
    DROP COLUMN IF EXISTS workflow_experiment_id,
    DROP COLUMN IF EXISTS workflow_experiment_scope_id,
    DROP COLUMN IF EXISTS workflow_experiment_variant_id,
    DROP COLUMN IF EXISTS input_schema_compatibility_id,
    DROP COLUMN IF EXISTS selection_reason,
    DROP COLUMN IF EXISTS version_input_data,
    DROP COLUMN IF EXISTS routing_decision;

DROP INDEX IF EXISTS message_workflow_input_schema_idx;
DROP INDEX IF EXISTS message_workflow_idempotency_key_uq;
ALTER TABLE "message" DROP CONSTRAINT IF EXISTS message_workflow_input_schema_id_fkey;
ALTER TABLE "message"
    DROP COLUMN IF EXISTS workflow_input_schema_id,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS error_message;

DROP TABLE IF EXISTS "workflow_configuration_audit_log";
DROP TABLE IF EXISTS "workflow_experiment_variant";
DROP TABLE IF EXISTS "workflow_experiment_scope";
DROP TABLE IF EXISTS "workflow_experiment";
DROP TABLE IF EXISTS "workflow_version_input_schema_compatibility";
DROP TABLE IF EXISTS "workflow_input_mapper";
DROP TABLE IF EXISTS "workflow_input_schema";

DROP INDEX IF EXISTS workflow_version_workflow_version_number_uq;
ALTER TABLE "workflow_version" DROP CONSTRAINT IF EXISTS workflow_version_updated_by_user_id_fkey;
ALTER TABLE "workflow_version"
    DROP COLUMN IF EXISTS locked_at,
    DROP COLUMN IF EXISTS published_at,
    DROP COLUMN IF EXISTS archived_at,
    DROP COLUMN IF EXISTS updated_by_user_id;
