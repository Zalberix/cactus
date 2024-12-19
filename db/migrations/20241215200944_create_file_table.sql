-- migrate:up
CREATE TABLE file (
    id_file SERIAL PRIMARY KEY,
    id_message INT NOT NULL,
    title VARCHAR(255),
    name VARCHAR(255),
    ext VARCHAR(50),
    link TEXT,
    FOREIGN KEY (id_message) REFERENCES message (id) ON DELETE SET NULL
);

-- migrate:down
DROP TABLE IF EXISTS "file" CASCADE;
