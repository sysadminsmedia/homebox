-- +goose Up
ALTER TABLE entity_templates
    ADD COLUMN default_low_stock_threshold REAL;