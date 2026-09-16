-- +goose Up
ALTER TABLE entities
    ADD COLUMN low_stock_threshold REAL;

ALTER TABLE entities
    ADD CONSTRAINT entities_low_stock_threshold_non_negative
    CHECK (low_stock_threshold IS NULL OR low_stock_threshold >= 0);