-- +goose Up
-- Add media_type column to label_templates table for Brother printer support
ALTER TABLE label_templates ADD COLUMN media_type VARCHAR(50);

-- +goose Down
ALTER TABLE label_templates DROP COLUMN IF EXISTS media_type;
