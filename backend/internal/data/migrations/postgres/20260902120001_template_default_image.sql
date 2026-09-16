-- +goose Up
ALTER TABLE "attachments"
    ADD COLUMN "entity_template_default_image" uuid NULL
        CONSTRAINT "attachments_entity_templates_default_image"
            REFERENCES "entity_templates" ("id") ON DELETE SET NULL;

ALTER TABLE "attachments"
    ADD CONSTRAINT "attachments_entity_template_default_image_key"
        UNIQUE ("entity_template_default_image");
