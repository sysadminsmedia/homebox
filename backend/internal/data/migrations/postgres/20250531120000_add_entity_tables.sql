-- Create entity_types table
CREATE TABLE IF NOT EXISTS "entity_types" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "name" character varying NOT NULL,
    "description" character varying NULL,
    "icon" character varying NULL,
    "color" character varying NULL,
    "is_location" boolean NOT NULL DEFAULT false,
    "group_entity_types" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "entity_types_groups_entity_types" FOREIGN KEY ("group_entity_types") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- Create entities table
CREATE TABLE IF NOT EXISTS "entities" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "name" character varying NOT NULL,
    "description" character varying NULL,
    "import_ref" character varying NULL,
    "notes" character varying NULL,
    "quantity" bigint NOT NULL DEFAULT 1,
    "insured" boolean NOT NULL DEFAULT false,
    "archived" boolean NOT NULL DEFAULT false,
    "asset_id" bigint NOT NULL DEFAULT 0,
    "serial_number" character varying NULL,
    "model_number" character varying NULL,
    "manufacturer" character varying NULL,
    "lifetime_warranty" boolean NOT NULL DEFAULT false,
    "warranty_expires" timestamptz NULL,
    "warranty_details" character varying NULL,
    "purchase_time" timestamptz NULL,
    "purchase_from" character varying NULL,
    "purchase_price" double precision NOT NULL DEFAULT 0,
    "sold_time" timestamptz NULL,
    "sold_to" character varying NULL,
    "sold_price" double precision NOT NULL DEFAULT 0,
    "sold_notes" character varying NULL,
    "group_entities" uuid NOT NULL,
    "entity_children" uuid NULL,
    "entity_parent" uuid NULL,
    "entity_type" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "entities_groups_entities" FOREIGN KEY ("group_entities") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "entities_entities_children" FOREIGN KEY ("entity_children") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
    CONSTRAINT "entities_entities_parent" FOREIGN KEY ("entity_parent") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
    CONSTRAINT "entities_entity_types" FOREIGN KEY ("entity_type") REFERENCES "entity_types" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- Create entity_fields table
CREATE TABLE IF NOT EXISTS "entity_fields" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "name" character varying NOT NULL,
    "description" character varying NULL,
    "type" character varying NOT NULL,
    "text_value" character varying NULL,
    "number_value" bigint NULL,
    "boolean_value" boolean NOT NULL DEFAULT false,
    "time_value" timestamptz NULL,
    "entity_fields" uuid NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "entity_fields_entities_fields" FOREIGN KEY ("entity_fields") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- Create default entity types
INSERT INTO "entity_types" ("id", "created_at", "updated_at", "name", "description", "is_location", "group_entity_types")
SELECT
    gen_random_uuid(),
    NOW(),
    NOW(),
    'Item',
    'Default item type',
    false,
    id
FROM "groups";

INSERT INTO "entity_types" ("id", "created_at", "updated_at", "name", "description", "is_location", "group_entity_types")
SELECT
    gen_random_uuid(),
    NOW(),
    NOW(),
    'Location',
    'Default location type',
    true,
    id
FROM "groups";

-- Migrate locations to entities
INSERT INTO "entities" (
    "id", "created_at", "updated_at", "name", "description",
    "group_entities", "entity_children", "entity_type"
)
SELECT
    l."id", l."created_at", l."updated_at", l."name", l."description",
    l."group_locations", l."location_children",
    (SELECT et."id" FROM "entity_types" et WHERE et."name" = 'Location' AND et."group_entity_types" = l."group_locations" LIMIT 1)
FROM "locations" l;

-- Migrate items to entities
INSERT INTO "entities" (
    "id", "created_at", "updated_at", "name", "description",
    "import_ref", "notes", "quantity", "insured", "archived",
    "asset_id", "serial_number", "model_number", "manufacturer",
    "lifetime_warranty", "warranty_expires", "warranty_details",
    "purchase_time", "purchase_from", "purchase_price",
    "sold_time", "sold_to", "sold_price", "sold_notes",
    "group_entities", "entity_children", "entity_parent", "entity_type"
)
SELECT
    i."id", i."created_at", i."updated_at", i."name", i."description",
    i."import_ref", i."notes", i."quantity", i."insured", i."archived",
    i."asset_id", i."serial_number", i."model_number", i."manufacturer",
    i."lifetime_warranty", i."warranty_expires", i."warranty_details",
    i."purchase_time", i."purchase_from", i."purchase_price",
    i."sold_time", i."sold_to", i."sold_price", i."sold_notes",
    i."group_items", i."item_children",
    i."location_items",
    (SELECT et."id" FROM "entity_types" et WHERE et."name" = 'Item' AND et."group_entity_types" = i."group_items" LIMIT 1)
FROM "items" i;

-- Migrate item_fields to entity_fields
INSERT INTO "entity_fields" (
    "id", "created_at", "updated_at", "name", "description",
    "type", "text_value", "number_value", "boolean_value", "time_value", "entity_fields"
)
SELECT
    "id", "created_at", "updated_at", "name", "description",
    "type", "text_value", "number_value", "boolean_value", "time_value", "item_fields"
FROM "item_fields";

-- Update maintenance_entries to reference entities instead of items
ALTER TABLE "maintenance_entries"
ADD COLUMN "entity_id" uuid NULL;

UPDATE "maintenance_entries"
SET "entity_id" = "item_id";

ALTER TABLE "maintenance_entries"
DROP CONSTRAINT "maintenance_entries_items_maintenance_entries";

ALTER TABLE "maintenance_entries"
ADD CONSTRAINT "maintenance_entries_entities_maintenance_entries"
FOREIGN KEY ("entity_id") REFERENCES "entities" ("id")
ON UPDATE NO ACTION ON DELETE CASCADE;

ALTER TABLE "maintenance_entries"
DROP COLUMN "item_id";

-- Update attachments to reference entities instead of items
ALTER TABLE "attachments"
ADD COLUMN "entity_attachments" uuid NULL;

UPDATE "attachments"
SET "entity_attachments" = "item_attachments";

ALTER TABLE "attachments"
DROP CONSTRAINT "attachments_items_attachments";

ALTER TABLE "attachments"
ADD CONSTRAINT "attachments_entities_attachments"
FOREIGN KEY ("entity_attachments") REFERENCES "entities" ("id")
ON UPDATE NO ACTION ON DELETE CASCADE;

ALTER TABLE "attachments"
DROP COLUMN "item_attachments";

-- Update labels to reference entities
CREATE TABLE IF NOT EXISTS "label_entities" (
    "label_id" uuid NOT NULL,
    "entity_id" uuid NOT NULL,
    PRIMARY KEY ("label_id", "entity_id"),
    CONSTRAINT "label_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "label_entities_label_id" FOREIGN KEY ("label_id") REFERENCES "labels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

INSERT INTO "label_entities" ("label_id", "entity_id")
SELECT "label_id", "item_id" FROM "label_items";

-- Drop old tables (do this last)
DROP TABLE IF EXISTS "label_items";
DROP TABLE IF EXISTS "item_fields";
DROP TABLE IF EXISTS "items";
DROP TABLE IF EXISTS "locations";
