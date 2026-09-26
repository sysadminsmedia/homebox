-- +goose Up
ALTER TABLE entity_templates
    ADD COLUMN IF NOT EXISTS default_low_stock_threshold double precision NULL;

ALTER TABLE entity_templates
    ADD CONSTRAINT entity_templates_default_low_stock_threshold_non_negative
    CHECK (default_low_stock_threshold IS NULL OR default_low_stock_threshold >= 0) NOT VALID;
