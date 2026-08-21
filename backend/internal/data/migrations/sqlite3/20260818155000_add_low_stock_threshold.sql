-- +goose Up
ALTER TABLE entities
    ADD COLUMN low_stock_threshold REAL;

ALTER TABLE entity_templates
    ADD COLUMN default_low_stock_threshold REAL;