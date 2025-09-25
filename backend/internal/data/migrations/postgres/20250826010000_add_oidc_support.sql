-- +goose Up
-- Add OIDC support fields to users table
ALTER TABLE users ADD COLUMN auth_provider TEXT DEFAULT 'local';
ALTER TABLE users ADD COLUMN external_id TEXT;

-- Make password optional for OIDC users
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- Create index on external_id for faster lookups
CREATE INDEX idx_users_external_id ON users(external_id);
CREATE INDEX idx_users_auth_provider ON users(auth_provider);

-- +goose Down
-- Revert indexes
DROP INDEX IF EXISTS idx_users_external_id;
DROP INDEX IF EXISTS idx_users_auth_provider;

-- Revert columns
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
ALTER TABLE users DROP COLUMN IF EXISTS external_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;