-- +goose Up
-- +goose no transaction
PRAGMA foreign_keys=OFF;

ALTER TABLE entity_types ADD COLUMN entity_type_default_template uuid
    REFERENCES entity_templates(id) ON DELETE SET NULL;

PRAGMA foreign_keys=ON;
