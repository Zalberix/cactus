-- +goose Up

ALTER TABLE "workflow_version"
    ADD COLUMN IF NOT EXISTS "name" VARCHAR(255);

CREATE TABLE IF NOT EXISTS "workflow_version_input" (
    id SERIAL PRIMARY KEY,
    workflow_version_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('string', 'number', 'integer', 'boolean', 'object', 'array')),
    required BOOLEAN DEFAULT TRUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT workflow_version_input_version_id_fkey FOREIGN KEY (workflow_version_id) REFERENCES "workflow_version"(id) ON DELETE CASCADE,
    CONSTRAINT workflow_version_input_name_uq UNIQUE (workflow_version_id, name)
);

-- +goose Down

DROP TABLE IF EXISTS "workflow_version_input" CASCADE;

ALTER TABLE "workflow_version"
    DROP COLUMN IF EXISTS "name";
