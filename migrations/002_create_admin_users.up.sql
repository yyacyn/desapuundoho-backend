-- 002_create_admin_users.up.sql
-- Admin users table for JWT authentication

CREATE TABLE IF NOT EXISTS admin_users (
    id            SERIAL        PRIMARY KEY,
    username      VARCHAR(100)  NOT NULL UNIQUE,
    password_hash TEXT          NOT NULL,
    role          VARCHAR(50)   NOT NULL DEFAULT 'admin',
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_users_username ON admin_users(username);
