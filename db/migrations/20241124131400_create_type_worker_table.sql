-- migrate:up
CREATE TABLE type_worker (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    CONSTRAINT type_worker_slug_key UNIQUE (slug)
    -- CONSTRAINT type_worker_name_key UNIQUE ("name")
);

-- migrate:down
DROP TABLE IF EXISTS type_worker CASCADE;
