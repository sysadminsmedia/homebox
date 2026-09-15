-- +goose Up
-- Single-use password reset tokens for the forgot-password flow. Tokens are
-- stored as a sha256 hash; the raw value is sent only over email/CLI exactly
-- once. used_at is set when consumed so a replay finds neither a fresh row
-- nor an unmarked one.
CREATE TABLE IF NOT EXISTS "password_reset_tokens" (
    "id"         uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "user_id"    uuid NOT NULL,
    "token"      bytea NOT NULL,
    "expires_at" timestamptz NOT NULL,
    "used_at"    timestamptz NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "password_reset_tokens_users_password_reset_tokens" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "password_reset_tokens_token_key" ON "password_reset_tokens" ("token");
CREATE INDEX IF NOT EXISTS "passwordresettokens_user_id" ON "password_reset_tokens" ("user_id");

-- +goose Down
DROP TABLE IF EXISTS "password_reset_tokens";
