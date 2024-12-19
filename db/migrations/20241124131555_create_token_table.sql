-- migrate:up
CREATE TABLE "token" (
    id_system INT NOT NULL,
    id_kind_worker INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    public_token VARCHAR(255) UNIQUE NOT NULL,
    secret_token VARCHAR(255) NOT NULL,
    PRIMARY KEY (id_system, id_kind_worker),
    FOREIGN KEY (id_system, id_kind_worker) 
        REFERENCES kind_worker_system (id_system, id_kind_worker) 
        ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS "token" CASCADE;
