-- migrate:up
CREATE TABLE kind_worker (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    config_schema JSONB,
    config JSONB,
    CONSTRAINT kind_worker_slug_key UNIQUE (slug)
);

-- migrate:down
DROP TABLE IF EXISTS kind_worker CASCADE;
