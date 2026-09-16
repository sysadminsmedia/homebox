-- +goose Up
ALTER TABLE entities
    ADD COLUMN low_stock_threshold REAL;