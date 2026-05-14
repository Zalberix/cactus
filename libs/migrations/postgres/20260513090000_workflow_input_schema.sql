-- +goose Up

ALTER TABLE "workflow"
    ADD COLUMN input_schema JSONB;

UPDATE "workflow"
SET input_schema = input_validation;

UPDATE "workflow"
SET input_schema = '{"type":"object","properties":{}}'::jsonb
WHERE input_schema IS NULL;

ALTER TABLE "workflow"
    ALTER COLUMN input_schema SET DEFAULT '{"type":"object","properties":{}}'::jsonb,
    ALTER COLUMN input_schema SET NOT NULL;

ALTER TABLE "workflow"
    DROP COLUMN input_validation;

DROP TABLE IF EXISTS "workflow_version_input" CASCADE;

-- +goose Down

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

ALTER TABLE "workflow"
    ADD COLUMN input_validation JSONB;

UPDATE "workflow"
SET input_validation = input_schema;

ALTER TABLE "workflow"
    DROP COLUMN input_schema;
