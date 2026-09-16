-- +goose Up
ALTER TABLE entities ADD COLUMN external_id VARCHAR(255);

-- +goose Down
-- (cannot reverse)