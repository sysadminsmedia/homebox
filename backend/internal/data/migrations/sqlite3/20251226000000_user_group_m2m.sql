-- +goose Up
-- Create user_groups junction table for M:M relationship
CREATE TABLE IF NOT EXISTS user_groups (
    user_id UUID NOT NULL,
    group_id UUID NOT NULL,
    PRIMARY KEY (user_id, group_id),
    CONSTRAINT user_groups_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT user_groups_group_id FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

-- Migrate existing user->group relationships to the junction table
INSERT INTO user_groups (user_id, group_id)
SELECT id, group_users FROM users WHERE group_users IS NOT NULL;

-- Add default_group_id column to users table
ALTER TABLE users ADD COLUMN default_group_id UUID;

-- Set default_group_id to the user's current group
UPDATE users SET default_group_id = group_users WHERE group_users IS NOT NULL;

-- Add foreign key constraint for default_group_id
CREATE TABLE users_new (
    id UUID NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT,
    is_superuser BOOLEAN NOT NULL DEFAULT false,
    superuser BOOLEAN NOT NULL DEFAULT false,
    role TEXT NOT NULL DEFAULT 'user',
    activated_on DATETIME,
    oidc_issuer TEXT,
    oidc_subject TEXT,
    default_group_id UUID,
    PRIMARY KEY (id),
    CONSTRAINT users_groups_users_default FOREIGN KEY (default_group_id) REFERENCES groups(id) ON DELETE SET NULL,
    UNIQUE (oidc_issuer, oidc_subject)
);

-- Copy data from old table to new table
INSERT INTO users_new (
    id, created_at, updated_at, name, email, password, is_superuser, superuser, role,
    activated_on, oidc_issuer, oidc_subject, default_group_id
)
SELECT
    id, created_at, updated_at, name, email, password, is_superuser, superuser, role,
    activated_on, oidc_issuer, oidc_subject, default_group_id
FROM users;

-- Drop old indexes
DROP INDEX IF EXISTS users_email_key;
DROP INDEX IF EXISTS users_oidc_issuer_subject_key;

-- Drop old table
DROP TABLE users;

-- Rename new table to users
ALTER TABLE users_new RENAME TO users;

-- Recreate indexes
CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS users_oidc_issuer_subject_key ON users(oidc_issuer, oidc_subject);

-- +goose Down
-- Recreate the old schema
CREATE TABLE users_old (
    id UUID NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT,
    is_superuser BOOLEAN NOT NULL DEFAULT false,
    superuser BOOLEAN NOT NULL DEFAULT false,
    role TEXT NOT NULL DEFAULT 'user',
    activated_on DATETIME,
    oidc_issuer TEXT,
    oidc_subject TEXT,
    group_users UUID NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT users_groups_users FOREIGN KEY (group_users) REFERENCES groups(id) ON DELETE CASCADE,
    UNIQUE (oidc_issuer, oidc_subject)
);

-- Copy data back, using the first group from user_groups
INSERT INTO users_old (
    id, created_at, updated_at, name, email, password, is_superuser, superuser, role,
    activated_on, oidc_issuer, oidc_subject, group_users
)
SELECT
    u.id, u.created_at, u.updated_at, u.name, u.email, u.password, u.is_superuser, u.superuser, u.role,
    u.activated_on, u.oidc_issuer, u.oidc_subject, COALESCE(u.default_group_id, (SELECT group_id FROM user_groups WHERE user_id = u.id LIMIT 1))
FROM users u;

DROP INDEX IF EXISTS users_email_key;
DROP INDEX IF EXISTS users_oidc_issuer_subject_key;
DROP TABLE users;
ALTER TABLE users_old RENAME TO users;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS users_oidc_issuer_subject_key ON users(oidc_issuer, oidc_subject);

DROP TABLE IF EXISTS user_groups;

