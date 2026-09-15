-- +goose Up
-- +goose no transaction
-- Move ownership from a global users.role enum to a per-membership role on
-- user_groups. Backfill: a user is owner of their default_group_id (the group
-- they registered with — fixed at registration and not user-mutable) and is a
-- regular member of every other group they belong to. The prior global
-- users.role column is removed; it was load-bearing for cross-group wipe and
-- caused authorization bugs because every self-registered user got 'owner'
-- regardless of the group being acted on.
PRAGMA foreign_keys=OFF;

-- 1. Add role to user_groups.
ALTER TABLE user_groups ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

-- 2. Backfill role=owner where the membership is the user's default group.
UPDATE user_groups
SET role = 'owner'
WHERE EXISTS (
    SELECT 1 FROM users
    WHERE users.id = user_groups.user_id
      AND users.default_group_id = user_groups.group_id
);

-- 3. Drop users.role via table rebuild (SQLite-safe).
CREATE TABLE users_new (
    id UUID NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT,
    is_superuser BOOLEAN NOT NULL DEFAULT false,
    superuser BOOLEAN NOT NULL DEFAULT false,
    activated_on DATETIME,
    oidc_issuer TEXT,
    oidc_subject TEXT,
    default_group_id UUID,
    settings JSON DEFAULT '{}',
    PRIMARY KEY (id),
    CONSTRAINT users_groups_users_default FOREIGN KEY (default_group_id) REFERENCES groups(id) ON DELETE SET NULL
);

INSERT INTO users_new (
    id, created_at, updated_at, name, email, password, is_superuser, superuser,
    activated_on, oidc_issuer, oidc_subject, default_group_id, settings
)
SELECT
    id, created_at, updated_at, name, email, password, is_superuser, superuser,
    activated_on, oidc_issuer, oidc_subject, default_group_id, settings
FROM users;

DROP INDEX IF EXISTS users_email_key;
DROP INDEX IF EXISTS users_oidc_issuer_subject_key;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS users_oidc_issuer_subject_key ON users(oidc_issuer, oidc_subject);

PRAGMA foreign_keys=ON;

-- +goose Down
-- +goose no transaction
PRAGMA foreign_keys=OFF;

CREATE TABLE users_new (
    id UUID NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT,
    is_superuser BOOLEAN NOT NULL DEFAULT false,
    superuser BOOLEAN NOT NULL DEFAULT false,
    role TEXT NOT NULL DEFAULT 'user',
    activated_on DATETIME,
    oidc_issuer TEXT,
    oidc_subject TEXT,
    default_group_id UUID,
    settings JSON DEFAULT '{}',
    PRIMARY KEY (id),
    CONSTRAINT users_groups_users_default FOREIGN KEY (default_group_id) REFERENCES groups(id) ON DELETE SET NULL
);

INSERT INTO users_new (
    id, created_at, updated_at, name, email, password, is_superuser, superuser, role,
    activated_on, oidc_issuer, oidc_subject, default_group_id, settings
)
SELECT
    u.id, u.created_at, u.updated_at, u.name, u.email, u.password, u.is_superuser, u.superuser,
    CASE
        WHEN EXISTS (
            SELECT 1 FROM user_groups ug
            WHERE ug.user_id = u.id
              AND ug.group_id = u.default_group_id
              AND ug.role = 'owner'
        ) THEN 'owner'
        ELSE 'user'
    END,
    u.activated_on, u.oidc_issuer, u.oidc_subject, u.default_group_id, u.settings
FROM users u;

DROP INDEX IF EXISTS users_email_key;
DROP INDEX IF EXISTS users_oidc_issuer_subject_key;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_key ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS users_oidc_issuer_subject_key ON users(oidc_issuer, oidc_subject);

ALTER TABLE user_groups DROP COLUMN role;

PRAGMA foreign_keys=ON;
