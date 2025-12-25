-- +goose Up
-- Add media_type column to label_templates table for Brother printer support
ALTER TABLE label_templates ADD COLUMN media_type TEXT;

-- +goose Down
-- SQLite doesn't support DROP COLUMN in older versions (< 3.35.0).
-- This down migration is intentionally a no-op.
-- Note: Rolling back this migration will leave the media_type column in place,
-- which may cause schema drift if comparing against a fresh database.
-- The column is nullable and unused after rollback, so this is safe.
