-- +goose Up
-- Make attachment paths relative by removing the prefix path
-- This migration converts absolute paths to relative paths by finding the UUID/documents pattern

-- Update Unix-style paths that contain "/documents/" by extracting the part starting from the UUID
-- The approach: find the "/documents/" substring, go back 37 characters (UUID + slash), 
-- and extract from there to get "uuid/documents/hash"
UPDATE attachments 
SET path = SUBSTRING(path FROM POSITION('/documents/' IN path) - 36)
WHERE path LIKE '%/documents/%' 
  AND POSITION('/documents/' IN path) > 36;

-- Update Windows-style paths that contain "\documents\" by extracting the part starting from the UUID
-- Convert backslashes to forward slashes in the process for consistency
UPDATE attachments 
SET path = REPLACE(SUBSTRING(path FROM POSITION(E'\\documents\\' IN path) - 36), E'\\', '/')
WHERE path LIKE E'%\\documents\\%'
  AND POSITION(E'\\documents\\' IN path) > 36;

-- For paths that already look like relative paths (start with UUID), leave them unchanged
-- This handles cases where the migration might be run multiple times