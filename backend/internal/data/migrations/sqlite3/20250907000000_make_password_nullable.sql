-- +goose Up
-- SQLite doesn't support ALTER COLUMN directly, so we need to recreate the table
-- Create a temporary table with the new schema
CREATE TABLE users_temp (
    id           uuid                not null
        primary key,
    created_at   datetime            not null,
    updated_at   datetime            not null,
    name         text                not null,
    email        text                not null,
    password     text,
    is_superuser bool default false  not null,
    superuser    bool default false  not null,
    role         text default 'user' not null,
    activated_on datetime,
    group_users  uuid                not null
        constraint users_groups_users
            references groups
            on delete cascade
);

-- Copy data from the original table
INSERT INTO users_temp SELECT id, created_at, updated_at, name, email, password, is_superuser, superuser, role, activated_on, group_users FROM users;

-- Drop the original table
DROP TABLE users;

-- Rename the temporary table
ALTER TABLE users_temp RENAME TO users;

-- Recreate the unique index
CREATE UNIQUE INDEX users_email_key on users (email);