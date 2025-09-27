-- +goose Up
-- Add OIDC support fields to users table
ALTER TABLE users ADD COLUMN auth_provider TEXT DEFAULT 'local';
ALTER TABLE users ADD COLUMN external_id TEXT;

-- Make password optional for OIDC users by removing NOT NULL constraint
-- Note: SQLite doesn't support dropping constraints directly, so we'll handle this in the application layer

-- Create index on external_id for faster lookups
CREATE INDEX idx_users_external_id ON users(external_id);
CREATE INDEX idx_users_auth_provider ON users(auth_provider);

-- +goose Down
-- SQLite down migration (no column drops for compatibility)
-- Drop indexes only (columns retained)
DROP INDEX IF EXISTS idx_users_external_id;
DROP INDEX IF EXISTS idx_users_auth_provider;