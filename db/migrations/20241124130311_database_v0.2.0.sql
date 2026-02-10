-- migrate:up
CREATE TABLE "permission" (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    CONSTRAINT permission_slug_key UNIQUE (slug)
);

CREATE TABLE "role" (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE "permission_role" (
    permission_id INT NOT NULL,
    role_id INT NOT NULL,
    CONSTRAINT permission_role_pkey PRIMARY KEY (permission_id, role_id),
    CONSTRAINT permission_role_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES "permission"(id) ON DELETE CASCADE,
    CONSTRAINT permission_role_role_id_fkey FOREIGN KEY (role_id) REFERENCES "role"(id) ON DELETE CASCADE
);

CREATE TABLE "user" (
    id SERIAL PRIMARY KEY,
    last_name VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    patronymic VARCHAR(255),
    email VARCHAR(255) NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    reset_password_after_login BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT user_email_key UNIQUE (email)
);

CREATE TABLE "role_user" (
    role_id INT NOT NULL,
    user_id INT NOT NULL,
    CONSTRAINT role_user_pkey PRIMARY KEY (role_id, user_id),
    CONSTRAINT role_user_user_id_fkey FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    CONSTRAINT role_user_role_id_fkey FOREIGN KEY (role_id) REFERENCES "role"(id) ON DELETE CASCADE
);

CREATE TABLE "channel" (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    CONSTRAINT channel_slug_key UNIQUE (slug)
);

CREATE TABLE "config" (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    config_schema JSONB NOT NULL,
    config JSONB
);

CREATE TABLE "worker" (
    id SERIAL PRIMARY KEY,
    channel_id INT NOT NULL,
    config_id INT NOT NULL,
    is_active BOOLEAN DEFAULT FALSE NOT NULL,
    CONSTRAINT worker_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES "channel"(id) ON DELETE CASCADE,
    CONSTRAINT worker_config_id_fkey FOREIGN KEY (config_id) REFERENCES "config"(id) ON DELETE CASCADE
);

CREATE TABLE "system" (
    id SERIAL PRIMARY KEY,
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
    CONSTRAINT system_user_creator_id_fkey FOREIGN KEY (user_creator_id) REFERENCES "user"(id) ON DELETE SET NULL
);

CREATE TABLE "channel_system" (
    system_id INT NOT NULL,
    channel_id INT NOT NULL,
    PRIMARY KEY (system_id, channel_id),
    CONSTRAINT channel_system_system_id_fkey FOREIGN KEY (system_id) REFERENCES "system"(id) ON DELETE CASCADE,
    CONSTRAINT channel_system_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES "channel"(id) ON DELETE CASCADE
);

CREATE TABLE "manifest" (
    id SERIAL PRIMARY KEY,
    value JSONB NOT NULL
);

CREATE TABLE "message" (
    id SERIAL PRIMARY KEY,
    system_id INT NOT NULL,
    manifest_id INT NOT NULL,
    "uuid" UUID NOT NULL,
    priority INT NOT NULL DEFAULT 0,
    value JSONB DEFAULT '{}'::jsonb NOT NULL,
    send_at TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT message_uuid_key UNIQUE ("uuid"),
    CONSTRAINT message_system_id_fkey FOREIGN KEY (system_id) REFERENCES "system"(id) ON DELETE CASCADE,
    CONSTRAINT message_manifest_id_fkey FOREIGN KEY (manifest_id) REFERENCES "manifest"(id) ON DELETE CASCADE
);

CREATE TABLE "file" (
    id SERIAL PRIMARY KEY,
    message_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    ext VARCHAR(50) NOT NULL,
    url VARCHAR(255) NOT NULL,
    CONSTRAINT file_message_id_fkey FOREIGN KEY (message_id) REFERENCES "message"(id) ON DELETE CASCADE
);

CREATE TABLE "pipeline" (
    id SERIAL PRIMARY KEY,
    message_id INT NOT NULL,
    parent_pipeline_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT pipeline_message_id_fkey FOREIGN KEY (message_id) REFERENCES "message"(id) ON DELETE CASCADE,
    CONSTRAINT pipeline_parent_pipeline_id_fkey FOREIGN KEY (parent_pipeline_id) REFERENCES "pipeline"(id) ON DELETE CASCADE
);

CREATE TABLE "pipeline_step_status" (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    CONSTRAINT pipeline_step_status_slug_key UNIQUE (slug)
);

CREATE TABLE "pipeline_step" (
    id SERIAL PRIMARY KEY,
    pipeline_id INT NOT NULL,
    worker_id INT,
    channel_id INT NOT NULL,
    step INT NOT NULL,
    time_start TIMESTAMP NULL,
    time_end TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    pipeline_step_status_id INT NOT NULL,
    CONSTRAINT pipeline_step_unique UNIQUE (step, pipeline_id),
    CONSTRAINT pipeline_step_pipeline_id_fkey FOREIGN KEY (pipeline_id) REFERENCES "pipeline"(id) ON DELETE CASCADE,
    CONSTRAINT pipeline_step_worker_id_fkey FOREIGN KEY (worker_id) REFERENCES "worker"(id) ON DELETE SET NULL,
    CONSTRAINT pipeline_step_channel_id_fkey FOREIGN KEY (channel_id) REFERENCES "channel"(id) ON DELETE CASCADE,
    CONSTRAINT pipeline_step_pipeline_step_status_id_fkey FOREIGN KEY (pipeline_step_status_id) REFERENCES "pipeline_step_status"(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS "pipeline_step" CASCADE;
DROP TABLE IF EXISTS "pipeline_step_status" CASCADE;
DROP TABLE IF EXISTS "pipeline" CASCADE;
DROP TABLE IF EXISTS "file" CASCADE;
DROP TABLE IF EXISTS "message" CASCADE;
DROP TABLE IF EXISTS "manifest" CASCADE;
DROP TABLE IF EXISTS "channel_system" CASCADE;
DROP TABLE IF EXISTS "system" CASCADE;
DROP TABLE IF EXISTS "worker" CASCADE;
DROP TABLE IF EXISTS "config" CASCADE;
DROP TABLE IF EXISTS "channel" CASCADE;
DROP TABLE IF EXISTS "role_user" CASCADE;
DROP TABLE IF EXISTS "user" CASCADE;
DROP TABLE IF EXISTS "permission_role" CASCADE;
DROP TABLE IF EXISTS "role" CASCADE;
DROP TABLE IF EXISTS "permission" CASCADE;
