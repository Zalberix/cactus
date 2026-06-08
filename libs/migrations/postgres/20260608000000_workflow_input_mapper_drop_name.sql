-- +goose Up

DROP INDEX IF EXISTS workflow_input_mapper_workflow_name_uq;

ALTER TABLE "workflow_input_mapper"
    DROP COLUMN IF EXISTS name;

-- +goose Down

ALTER TABLE "workflow_input_mapper"
    ADD COLUMN IF NOT EXISTS name VARCHAR(255);

UPDATE "workflow_input_mapper"
SET name = 'mapper-' || id::text
WHERE name IS NULL OR btrim(name) = '';

ALTER TABLE "workflow_input_mapper"
    ALTER COLUMN name SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS workflow_input_mapper_workflow_name_uq
    ON "workflow_input_mapper" (workflow_id, name)
    WHERE deleted_at IS NULL;
