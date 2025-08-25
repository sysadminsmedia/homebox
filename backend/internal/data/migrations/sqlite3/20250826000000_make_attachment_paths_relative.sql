-- +goose Up
-- Make attachment paths relative by removing the prefix path
-- This migration converts absolute paths to relative paths by finding the UUID/documents pattern

-- Update paths that contain "/documents/" by extracting the part starting from the UUID
-- The approach: find the "/documents/" substring, go back 37 characters (UUID + slash), 
-- and extract from there to get "uuid/documents/hash"
UPDATE attachments 
SET path = SUBSTR(path, INSTR(path, '/documents/') - 36)
WHERE path LIKE '%/documents/%' 
  AND INSTR(path, '/documents/') > 36;

-- For paths that already look like relative paths (start with UUID), leave them unchanged
-- This handles cases where the migration might be run multiple times

-- +goose Down
-- Note: This down migration cannot be safely implemented because we don't know 
-- what the original prefix paths were. This is a one-way migration.