-- migrate:up
CREATE TABLE "permission" (
    id SERIAL PRIMARY KEY,
    description TEXT,
    slug VARCHAR(255) NOT NULL,
    CONSTRAINT permission_slug_key UNIQUE (slug)
);

-- migrate:down
DROP TABLE IF EXISTS "permission" CASCADE;
