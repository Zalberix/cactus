-- migrate:up
CREATE TABLE "priority" (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    "weight" INT NOT NULL CHECK ("weight" >= 0),
    "slug" VARCHAR(255) NOT NULL,
    CONSTRAINT priority_weight_key UNIQUE ("weight"),
    CONSTRAINT priority_slug_key UNIQUE ("slug"),
    CONSTRAINT priority_name_key UNIQUE ("name")
);

-- migrate:down
DROP TABLE IF EXISTS "priority" CASCADE;
