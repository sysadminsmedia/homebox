-- +goose Up
-- +goose no transaction
PRAGMA foreign_keys = OFF;

ALTER TABLE groups ADD COLUMN external_ids_enabled boolean;

PRAGMA foreign_keys = ON;

-- +goose Down
-- (cannot reverse)