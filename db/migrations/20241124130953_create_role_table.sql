-- migrate:up
CREATE TABLE "role" (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    CONSTRAINT role_slug_key UNIQUE (slug)
);

-- migrate:down
DROP TABLE IF EXISTS "role" CASCADE;
