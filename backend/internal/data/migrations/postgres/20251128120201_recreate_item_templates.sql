-- +goose Up
-- Create "item_templates" table
CREATE TABLE IF NOT EXISTS "item_templates" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "name" character varying NOT NULL,
    "description" character varying NULL,
    "notes" character varying NULL,
    "default_quantity" bigint NOT NULL DEFAULT 1,
    "default_insured" boolean NOT NULL DEFAULT false,
    "default_name" character varying NULL,
    "default_description" character varying NULL,
    "default_manufacturer" character varying NULL,
    "default_model_number" character varying NULL,
    "default_lifetime_warranty" boolean NOT NULL DEFAULT false,
    "default_warranty_details" character varying NULL,
    "include_warranty_fields" boolean NOT NULL DEFAULT false,
    "include_purchase_fields" boolean NOT NULL DEFAULT false,
    "include_sold_fields" boolean NOT NULL DEFAULT false,
    "default_label_ids" jsonb NULL,
    "item_template_location" uuid NULL,
    "group_item_templates" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "item_templates_groups_item_templates" FOREIGN KEY ("group_item_templates") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "item_templates_locations_location" FOREIGN KEY ("item_template_location") REFERENCES "locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "template_fields" table
CREATE TABLE IF NOT EXISTS "template_fields" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "type" character varying NOT NULL, "text_value" character varying NULL, "item_template_fields" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "template_fields_item_templates_fields" FOREIGN KEY ("item_template_fields") REFERENCES "item_templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
