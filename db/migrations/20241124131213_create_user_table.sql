-- migrate:up
CREATE TABLE "user" (
    id SERIAL PRIMARY KEY,
    fio VARCHAR(255) NOT NULL,
    login VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    reset_password_after_login BOOLEAN DEFAULT FALSE,
    create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_email_key UNIQUE (email),
    CONSTRAINT user_login_key UNIQUE (login)
);

-- migrate:down
DROP TABLE IF EXISTS "user" CASCADE;
