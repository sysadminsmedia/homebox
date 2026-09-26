-- +goose Up
ALTER TABLE entities
    ADD COLUMN low_stock_threshold REAL
    CHECK (low_stock_threshold IS NULL OR low_stock_threshold >= 0);
