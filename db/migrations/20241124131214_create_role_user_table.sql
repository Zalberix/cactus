-- migrate:up
CREATE TABLE "role_user" (
    id_role INT NOT NULL,
    id_user INT NOT NULL,
    CONSTRAINT role_user_pkey PRIMARY KEY (id_role, id_user),
    CONSTRAINT role_user_id_user_fkey FOREIGN KEY (id_user) REFERENCES "user"(id) ON DELETE CASCADE,
    CONSTRAINT role_user_id_role_fkey FOREIGN KEY (id_role) REFERENCES "role"(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS "role_user" CASCADE;
