-- +goose Up
CREATE INDEX IF NOT EXISTS idx_locations_name ON locations(name);
CREATE INDEX IF NOT EXISTS idx_labels_name ON labels(name);

-- +goose Down
DROP INDEX IF EXISTS idx_locations_name;
DROP INDEX IF EXISTS idx_labels_name;
