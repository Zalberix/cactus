-- migrate:up
CREATE TABLE "system" (
    id SERIAL PRIMARY KEY,
    create_user INT,
    id_priority INT NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    CONSTRAINT system_create_user_fkey FOREIGN KEY (create_user) REFERENCES "user"(id) ON DELETE SET NULL,
    CONSTRAINT system_id_priority_fkey FOREIGN KEY (id_priority) REFERENCES priority(id) ON DELETE SET NULL
);

-- migrate:down
DROP TABLE IF EXISTS system CASCADE;
