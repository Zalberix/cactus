-- migrate:up
CREATE TABLE permission_role (
    id_permission INT NOT NULL,
    id_role INT NOT NULL,
    CONSTRAINT permission_role_pkey PRIMARY KEY (id_permission, id_role),
    CONSTRAINT permission_role_id_permission_fkey FOREIGN KEY (id_permission) REFERENCES "permission"(id) ON DELETE CASCADE,
    CONSTRAINT permission_role_id_role_fkey FOREIGN KEY (id_role) REFERENCES "role"(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS permission_role CASCADE;
