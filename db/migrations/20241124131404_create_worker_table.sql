-- migrate:up
CREATE TABLE worker (
    id SERIAL PRIMARY KEY,
    "is_active" BOOLEAN DEFAULT FALSE NOT NULL,
    id_type_worker INT NOT NULL,
    id_kind_worker INT NOT NULL,
    CONSTRAINT worker_id_type_worker_fkey FOREIGN KEY (id_type_worker) REFERENCES type_worker(id) ON DELETE SET NULL,
    CONSTRAINT worker_id_kind_worker_fkey FOREIGN KEY (id_kind_worker) REFERENCES kind_worker(id) ON DELETE SET NULL
);

-- migrate:down
DROP TABLE IF EXISTS worker CASCADE;
