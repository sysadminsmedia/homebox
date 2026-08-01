-- +goose Up
-- +goose no transaction
PRAGMA foreign_keys = OFF;

ALTER TABLE entities ADD COLUMN external_id VARCHAR(255);

PRAGMA foreign_keys = ON;

-- +goose Down
-- (cannot reverse)
