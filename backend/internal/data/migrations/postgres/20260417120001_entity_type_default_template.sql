-- +goose Up
ALTER TABLE "entity_types"
    ADD COLUMN "entity_type_default_template" uuid NULL
        CONSTRAINT "entity_types_entity_templates_default_template"
            REFERENCES "entity_templates" ("id") ON DELETE SET NULL;