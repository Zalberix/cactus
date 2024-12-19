-- migrate:up
CREATE TABLE pipline_status (
    id SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL
);

-- migrate:down
DROP TABLE IF EXISTS pipline_status CASCADE;
