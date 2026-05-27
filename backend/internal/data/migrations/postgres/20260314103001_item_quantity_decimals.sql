-- +goose Up
-- +goose no transaction
ALTER TABLE items ALTER COLUMN quantity TYPE double precision USING quantity::double precision;
ALTER TABLE items ALTER COLUMN quantity SET DEFAULT 1;

ALTER TABLE item_templates ALTER COLUMN default_quantity TYPE double precision USING default_quantity::double precision;
ALTER TABLE item_templates ALTER COLUMN default_quantity SET DEFAULT 1;

