-- +goose Up
ALTER TABLE entity_templates
    ADD COLUMN IF NOT EXISTS default_low_stock_threshold double precision NULL;