-- +goose Up
-- +goose no transaction
PRAGMA foreign_keys=OFF;

ALTER TABLE attachments ADD COLUMN entity_template_default_image uuid
    REFERENCES entity_templates(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS attachments_entity_template_default_image_key
    ON attachments(entity_template_default_image)
    WHERE entity_template_default_image IS NOT NULL;

PRAGMA foreign_keys=ON;
