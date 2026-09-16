-- +goose Up
ALTER TABLE entity_templates
    ADD COLUMN default_low_stock_threshold REAL;

ALTER TABLE entity_templates
    ADD CONSTRAINT entity_templates_default_low_stock_threshold_non_negative
    CHECK (default_low_stock_threshold IS NULL OR default_low_stock_threshold >= 0);
