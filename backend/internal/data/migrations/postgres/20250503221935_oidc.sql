-- +goose Up
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

CREATE TABLE IF NOT EXISTS "oauths"
(
    "id"         uuid     NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "provider"   text     NOT NULL,
    "sub"        text     NOT NULL,
    "user_oauth" uuid     NULL,
    PRIMARY KEY (id),
    CONSTRAINT oauths_users_oauth FOREIGN KEY (user_oauth) REFERENCES users (id) ON DELETE CASCADE
);
