-- +goose Up

-- ============================================================
-- Организация
-- ============================================================
CREATE TABLE "organization" (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    code VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT organization_code_key UNIQUE (code)
);

-- ============================================================
-- Права и роли
-- ============================================================
CREATE TABLE "permission" (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    is_private BOOLEAN DEFAULT FALSE NOT NULL,
    CONSTRAINT permission_slug_key UNIQUE (slug)
);

CREATE TABLE "role" (
    id SERIAL PRIMARY KEY,
    organization_id INT,
    "name" VARCHAR(255) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT role_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "organization"(id) ON DELETE SET NULL
);

CREATE TABLE "permission_role" (
    permission_id INT NOT NULL,
    role_id INT NOT NULL,
    CONSTRAINT permission_role_pkey PRIMARY KEY (permission_id, role_id),
    CONSTRAINT permission_role_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES "permission"(id) ON DELETE CASCADE,
    CONSTRAINT permission_role_role_id_fkey FOREIGN KEY (role_id) REFERENCES "role"(id) ON DELETE CASCADE
);

-- ============================================================
-- Пользователи
-- ============================================================
CREATE TABLE "user" (
    id SERIAL PRIMARY KEY,
    organization_id INT,
    last_name VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    patronymic VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    reset_password_after_login BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT user_email_key UNIQUE (email),
    CONSTRAINT user_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "organization"(id) ON DELETE SET NULL
);

CREATE TABLE "role_user" (
    role_id INT NOT NULL,
    user_id INT NOT NULL,
    CONSTRAINT role_user_pkey PRIMARY KEY (role_id, user_id),
    CONSTRAINT role_user_user_id_fkey FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    CONSTRAINT role_user_role_id_fkey FOREIGN KEY (role_id) REFERENCES "role"(id) ON DELETE CASCADE
);

-- ============================================================
-- Система
-- ============================================================
CREATE TABLE "system" (
    id SERIAL PRIMARY KEY,
    organization_id INT,
    user_creator_id INT,
    "name" VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    priority INT NOT NULL DEFAULT 0,
    public_token VARCHAR(255),
    private_token VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT system_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "organization"(id) ON DELETE SET NULL,
    CONSTRAINT system_user_creator_id_fkey FOREIGN KEY (user_creator_id) REFERENCES "user"(id) ON DELETE SET NULL
);

-- ============================================================
-- Типы работ
-- ============================================================
CREATE TABLE "work_type" (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    code VARCHAR(255) NOT NULL,
    description TEXT,
    meta JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT work_type_code_key UNIQUE (code)
);

CREATE TABLE "worker_bootstrap_token" (
    id SERIAL PRIMARY KEY,
    organization_id INT NOT NULL,
    work_type_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    token_hash VARCHAR(512) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
    max_active_workers INT NOT NULL DEFAULT 1 CHECK (max_active_workers > 0),
    total_registration_count INT NOT NULL DEFAULT 0 CHECK (total_registration_count >= 0),
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_by_user_id INT,
    revoked_at TIMESTAMP,
    revoked_by_user_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT worker_bootstrap_token_org_fkey FOREIGN KEY (organization_id) REFERENCES "organization"(id) ON DELETE CASCADE,
    CONSTRAINT worker_bootstrap_token_work_type_fkey FOREIGN KEY (work_type_id) REFERENCES "work_type"(id) ON DELETE CASCADE,
    CONSTRAINT worker_bootstrap_token_created_by_fkey FOREIGN KEY (created_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT worker_bootstrap_token_revoked_by_fkey FOREIGN KEY (revoked_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT worker_bootstrap_token_hash_key UNIQUE (token_hash)
);

CREATE INDEX worker_bootstrap_token_org_idx
    ON "worker_bootstrap_token" (organization_id, work_type_id)
    WHERE deleted_at IS NULL;

CREATE INDEX worker_bootstrap_token_active_hash_idx
    ON "worker_bootstrap_token" (token_hash)
    WHERE deleted_at IS NULL AND status = 'active';

-- ============================================================
-- Схемы и ревизии настроек работников
-- ============================================================
CREATE TABLE "worker_settings_schema" (
    id SERIAL PRIMARY KEY,
    work_type_id INT NOT NULL,
    "version" VARCHAR(255) NOT NULL,
    settings_schema JSONB NOT NULL,
    input_schema JSONB NOT NULL,
    output_schema JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT worker_settings_schema_work_type_id_fkey FOREIGN KEY (work_type_id) REFERENCES "work_type"(id) ON DELETE CASCADE
);

CREATE TABLE "worker_settings_revision" (
    id SERIAL PRIMARY KEY,
    worker_settings_schema_id INT NOT NULL,
    created_by_user_id INT,
    settings_data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT worker_settings_revision_schema_id_fkey FOREIGN KEY (worker_settings_schema_id) REFERENCES "worker_settings_schema"(id) ON DELETE CASCADE,
    CONSTRAINT worker_settings_revision_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL
);

CREATE TABLE "revision_rate_limit" (
    id SERIAL PRIMARY KEY,
    worker_settings_revision_id INT NOT NULL,
    "limit" INT NOT NULL,
    window_seconds INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT revision_rate_limit_revision_id_fkey FOREIGN KEY (worker_settings_revision_id) REFERENCES "worker_settings_revision"(id) ON DELETE CASCADE
);

-- ============================================================
-- Работник
-- ============================================================
CREATE TABLE "worker" (
    id SERIAL PRIMARY KEY,
    organization_id INT NOT NULL,
    work_type_id INT NOT NULL,
    worker_settings_schema_id INT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    metadata JSONB,
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_heartbeat_at TIMESTAMP,
    CONSTRAINT worker_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES "organization"(id) ON DELETE CASCADE,
    CONSTRAINT worker_work_type_id_fkey FOREIGN KEY (work_type_id) REFERENCES "work_type"(id) ON DELETE CASCADE,
    CONSTRAINT worker_schema_id_fkey FOREIGN KEY (worker_settings_schema_id) REFERENCES "worker_settings_schema"(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX worker_org_work_type_name_uq
    ON "worker" (organization_id, work_type_id, "name");

CREATE TABLE "worker_nats_session" (
    id SERIAL PRIMARY KEY,
    worker_id INT NOT NULL,
    bootstrap_token_id INT NOT NULL,
    nats_account_public_key VARCHAR(128) NOT NULL,
    nats_user_public_key VARCHAR(128) NOT NULL,
    nats_user_jwt TEXT NOT NULL,
    permissions JSONB NOT NULL DEFAULT '{}',
    revoked_at TIMESTAMP,
    revoked_by_user_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT worker_nats_session_worker_fkey FOREIGN KEY (worker_id) REFERENCES "worker"(id) ON DELETE CASCADE,
    CONSTRAINT worker_nats_session_bootstrap_fkey FOREIGN KEY (bootstrap_token_id) REFERENCES "worker_bootstrap_token"(id) ON DELETE CASCADE,
    CONSTRAINT worker_nats_session_revoked_by_fkey FOREIGN KEY (revoked_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT worker_nats_session_user_key_uq UNIQUE (nats_user_public_key)
);

CREATE INDEX worker_nats_session_active_worker_idx
    ON "worker_nats_session" (worker_id)
    WHERE revoked_at IS NULL;

-- ============================================================
-- Токены системы
-- ============================================================
CREATE TABLE "system_token" (
    id SERIAL PRIMARY KEY,
    system_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    public_token VARCHAR(512) NOT NULL,
    private_token VARCHAR(512) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT system_token_system_id_fkey FOREIGN KEY (system_id) REFERENCES "system"(id) ON DELETE CASCADE,
    CONSTRAINT system_token_system_id_name_uq UNIQUE (system_id, name)
);

-- ============================================================
-- Рабочий процесс (workflow)
-- ============================================================
CREATE TABLE "workflow" (
    id SERIAL PRIMARY KEY,
    system_id INT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    priority INT NOT NULL DEFAULT 2,
    input_schema JSONB DEFAULT '{"type":"object","properties":{}}'::jsonb NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT workflow_system_id_fkey FOREIGN KEY (system_id) REFERENCES "system"(id) ON DELETE CASCADE
);

CREATE TABLE "workflow_token" (
    system_token_id INT NOT NULL,
    workflow_id INT NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT workflow_token_pkey PRIMARY KEY (system_token_id, workflow_id),
    CONSTRAINT workflow_token_system_token_id_fkey FOREIGN KEY (system_token_id) REFERENCES "system_token"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_token_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES "workflow"(id) ON DELETE CASCADE
);

CREATE TABLE "workflow_version" (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL,
    created_by_user_id INT,
    version_number INT NOT NULL,
    "name" VARCHAR(255),
    is_valid BOOLEAN DEFAULT FALSE NOT NULL,
    is_active BOOLEAN DEFAULT FALSE NOT NULL,
    traffic_weight INT NOT NULL DEFAULT 100 CHECK (traffic_weight >= 0 AND traffic_weight <= 100),
    is_control_group BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT workflow_version_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES "workflow"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_version_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES "user"(id) ON DELETE SET NULL
);

-- ============================================================
-- Шаги рабочего процесса (DAG)
-- ============================================================
CREATE TABLE "workflow_step" (
    id SERIAL PRIMARY KEY,
    workflow_version_id INT NOT NULL,
    step_type VARCHAR(20) NOT NULL CHECK (step_type IN ('task', 'control')),
    work_type_id INT,
    worker_settings_revision_id INT,
    control_kind VARCHAR(50),
    control_settings JSONB,
    input_mapping JSONB,
    canvas_position JSONB DEFAULT '{"x":0,"y":0}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT workflow_step_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES "workflow_version"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_step_work_type_id_fkey FOREIGN KEY (work_type_id) REFERENCES "work_type"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_step_revision_id_fkey FOREIGN KEY (worker_settings_revision_id) REFERENCES "worker_settings_revision"(id) ON DELETE SET NULL,
    CONSTRAINT workflow_step_task_check CHECK (
        (step_type = 'task' AND work_type_id IS NOT NULL AND worker_settings_revision_id IS NOT NULL AND control_kind IS NULL)
        OR
        (step_type = 'control' AND work_type_id IS NULL AND worker_settings_revision_id IS NULL AND control_kind IS NOT NULL)
    )
);

CREATE TABLE "workflow_step_dependency" (
    step_id INT NOT NULL,
    depends_on_step_id INT NOT NULL,
    outcome VARCHAR(255),
    output_index INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT workflow_step_dependency_pkey PRIMARY KEY (step_id, depends_on_step_id),
    CONSTRAINT workflow_step_dependency_step_id_fkey FOREIGN KEY (step_id) REFERENCES "workflow_step"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_step_dependency_depends_on_fkey FOREIGN KEY (depends_on_step_id) REFERENCES "workflow_step"(id) ON DELETE CASCADE
);

-- ============================================================
-- Сообщение
-- ============================================================
CREATE TABLE "message" (
    id SERIAL PRIMARY KEY,
    workflow_id INT NOT NULL,
    external_message_id VARCHAR(255),
    overridden_priority INT CHECK (overridden_priority IS NULL OR (overridden_priority >= 0 AND overridden_priority <= 3)),
    value JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT message_workflow_id_fkey FOREIGN KEY (workflow_id) REFERENCES "workflow"(id) ON DELETE CASCADE
);

-- ============================================================
-- Запуск и шаги рабочего процесса
-- ============================================================
CREATE TABLE "workflow_run" (
    id SERIAL PRIMARY KEY,
    workflow_version_id INT NOT NULL,
    message_id INT NOT NULL,
    temporal_workflow_id VARCHAR(512),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    CONSTRAINT workflow_run_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES "workflow_version"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_run_message_id_fkey FOREIGN KEY (message_id) REFERENCES "message"(id) ON DELETE CASCADE
);

CREATE TABLE "workflow_run_step" (
    id SERIAL PRIMARY KEY,
    workflow_run_id INT NOT NULL,
    workflow_step_id INT NOT NULL,
    worker_id INT,
    temporal_step_id VARCHAR(512),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    outcome VARCHAR(255),
    input_data JSONB,
    output_data JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    CONSTRAINT workflow_run_step_run_id_fkey FOREIGN KEY (workflow_run_id) REFERENCES "workflow_run"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_run_step_step_id_fkey FOREIGN KEY (workflow_step_id) REFERENCES "workflow_step"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_run_step_worker_id_fkey FOREIGN KEY (worker_id) REFERENCES "worker"(id) ON DELETE SET NULL
);

CREATE TABLE "workflow_run_step_attempt" (
    id SERIAL PRIMARY KEY,
    workflow_run_step_id INT NOT NULL,
    worker_id INT,
    attempt_number INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_code VARCHAR(255),
    error_message TEXT,
    output_data JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    CONSTRAINT workflow_run_step_attempt_run_step_id_fkey FOREIGN KEY (workflow_run_step_id) REFERENCES "workflow_run_step"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_run_step_attempt_worker_id_fkey FOREIGN KEY (worker_id) REFERENCES "worker"(id) ON DELETE SET NULL
);

-- ============================================================
-- Файл
-- ============================================================
CREATE TABLE "file" (
    id SERIAL PRIMARY KEY,
    message_id INT,
    workflow_run_step_id INT,
    "name" VARCHAR(255) NOT NULL,
    bucket VARCHAR(255) NOT NULL,
    object_key VARCHAR(512) NOT NULL,
    content_type VARCHAR(255) NOT NULL,
    size_bytes INT NOT NULL,
    hash VARCHAR(512) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT file_object_key_key UNIQUE (object_key),
    CONSTRAINT file_message_id_fkey FOREIGN KEY (message_id) REFERENCES "message"(id) ON DELETE SET NULL,
    CONSTRAINT file_run_step_id_fkey FOREIGN KEY (workflow_run_step_id) REFERENCES "workflow_run_step"(id) ON DELETE SET NULL,
    CONSTRAINT file_one_fk_check CHECK (
        (message_id IS NOT NULL AND workflow_run_step_id IS NULL)
        OR (message_id IS NULL AND workflow_run_step_id IS NOT NULL)
        OR (message_id IS NULL AND workflow_run_step_id IS NULL)
    )
);

-- +goose Down
DROP TABLE IF EXISTS "file" CASCADE;
DROP TABLE IF EXISTS "workflow_run_step_attempt" CASCADE;
DROP TABLE IF EXISTS "workflow_run_step" CASCADE;
DROP TABLE IF EXISTS "workflow_run" CASCADE;
DROP TABLE IF EXISTS "message" CASCADE;
DROP TABLE IF EXISTS "workflow_step_dependency" CASCADE;
DROP TABLE IF EXISTS "workflow_step" CASCADE;
DROP TABLE IF EXISTS "workflow_version" CASCADE;
DROP TABLE IF EXISTS "workflow_token" CASCADE;
DROP TABLE IF EXISTS "workflow" CASCADE;
DROP TABLE IF EXISTS "system_token" CASCADE;
DROP TABLE IF EXISTS "worker_nats_session" CASCADE;
DROP TABLE IF EXISTS "worker" CASCADE;
DROP TABLE IF EXISTS "revision_rate_limit" CASCADE;
DROP TABLE IF EXISTS "worker_settings_revision" CASCADE;
DROP TABLE IF EXISTS "worker_settings_schema" CASCADE;
DROP TABLE IF EXISTS "worker_bootstrap_token" CASCADE;
DROP TABLE IF EXISTS "work_type" CASCADE;
DROP TABLE IF EXISTS "system" CASCADE;
DROP TABLE IF EXISTS "role_user" CASCADE;
DROP TABLE IF EXISTS "user" CASCADE;
DROP TABLE IF EXISTS "permission_role" CASCADE;
DROP TABLE IF EXISTS "role" CASCADE;
DROP TABLE IF EXISTS "permission" CASCADE;
DROP TABLE IF EXISTS "organization" CASCADE;
