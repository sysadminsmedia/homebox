-- +goose Up
-- Add static, user-scoped API keys. Each key authenticates as the owning
-- user and grants the same access. Tokens are stored hashed (sha256) and
-- never persisted in plaintext.
CREATE TABLE IF NOT EXISTS "api_keys" (
    "id"           uuid NOT NULL,
    "created_at"   timestamptz NOT NULL,
    "updated_at"   timestamptz NOT NULL,
    "user_id"      uuid NOT NULL,
    "name"         character varying NOT NULL,
    "token"        bytea NOT NULL,
    "expires_at"   timestamptz NULL,
    "last_used_at" timestamptz NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "api_keys_users_api_keys" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "api_keys_token_key" ON "api_keys" ("token");
CREATE INDEX IF NOT EXISTS "apikey_token" ON "api_keys" ("token");
CREATE INDEX IF NOT EXISTS "apikey_user_id" ON "api_keys" ("user_id");

-- +goose Down
DROP TABLE IF EXISTS "api_keys";
