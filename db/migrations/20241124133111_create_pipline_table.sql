-- migrate:up
CREATE TABLE pipline (
    id SERIAL PRIMARY KEY,
    id_pipline_status INT NOT NULL,
    id_message INT NOT NULL,
    step INT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    time_start TIMESTAMP NULL,
    time_end TIMESTAMP NULL,
    CONSTRAINT pipline_unique UNIQUE (step, id_message),
    CONSTRAINT pipline_id_message_fkey FOREIGN KEY (id_message) REFERENCES message(id) ON DELETE CASCADE,
    CONSTRAINT pipline_id_pipline_status_fkey FOREIGN KEY (id_pipline_status) REFERENCES pipline_status(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS pipline CASCADE;
