-- +goose Up
-- +goose no transaction
ALTER TABLE template_fields ADD COLUMN number_value INTEGER;
ALTER TABLE template_fields ADD COLUMN boolean_value BOOLEAN DEFAULT FALSE;
ALTER TABLE template_fields ADD COLUMN time_value timestamp;