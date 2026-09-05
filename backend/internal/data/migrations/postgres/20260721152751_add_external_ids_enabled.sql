-- +goose Up
ALTER TABLE groups ADD COLUMN external_ids_enabled bool;

-- +goose Down
-- (cannot reverse)