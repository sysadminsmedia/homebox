-- +goose Up
ALTER TABLE entities
    ADD COLUMN IF NOT EXISTS low_stock_threshold double precision NULL;

ALTER TABLE entity_templates
    ADD COLUMN IF NOT EXISTS default_low_stock_threshold double precision NULL;