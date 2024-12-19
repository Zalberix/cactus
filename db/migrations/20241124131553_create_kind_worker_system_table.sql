-- migrate:up
CREATE TABLE kind_worker_system (
    id_system INT NOT NULL,
    id_kind_worker INT NOT NULL,
    PRIMARY KEY (id_system, id_kind_worker),
    CONSTRAINT kind_worker_system_id_system_fkey FOREIGN KEY (id_system) REFERENCES system (id) ON DELETE CASCADE,
    CONSTRAINT kind_worker_system_id_kind_worker_fkey FOREIGN KEY (id_kind_worker) REFERENCES kind_worker (id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS "kind_worker_system" CASCADE;
