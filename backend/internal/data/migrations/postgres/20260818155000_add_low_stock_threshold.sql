-- +goose Up
ALTER TABLE entities
    ADD COLUMN IF NOT EXISTS low_stock_threshold double precision NULL;