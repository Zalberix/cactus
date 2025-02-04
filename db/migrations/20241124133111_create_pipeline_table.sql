-- migrate:up
CREATE TABLE pipeline (
    id SERIAL PRIMARY KEY,
    id_message INT NOT NULL,
    "status" VARCHAR(255) NOT NULL,
    step INT NOT NULL,
    id_worker INT,
    "name" VARCHAR(255) NOT NULL,
    time_start TIMESTAMP NULL,
    time_end TIMESTAMP NULL,
    CONSTRAINT pipline_unique UNIQUE (step, id_message),
    CONSTRAINT pipline_id_message_fkey FOREIGN KEY (id_message) REFERENCES message(id) ON DELETE CASCADE,
    CONSTRAINT pipline_id_worker_fkey FOREIGN KEY (id_worker) REFERENCES worker(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS pipeline CASCADE;
