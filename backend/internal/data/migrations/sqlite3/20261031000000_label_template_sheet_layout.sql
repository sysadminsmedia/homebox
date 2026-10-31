-- +goose Up
-- Editable sheet layout (null follows the preset) and printer calibration offsets
ALTER TABLE label_templates ADD COLUMN sheet_layout json;
ALTER TABLE label_templates ADD COLUMN offset_x real DEFAULT 0 NOT NULL;
ALTER TABLE label_templates ADD COLUMN offset_y real DEFAULT 0 NOT NULL;

-- +goose Down
ALTER TABLE label_templates DROP COLUMN offset_y;
ALTER TABLE label_templates DROP COLUMN offset_x;
ALTER TABLE label_templates DROP COLUMN sheet_layout;
