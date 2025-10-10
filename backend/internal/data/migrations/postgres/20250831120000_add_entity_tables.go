package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits
func init() {
	goose.AddMigrationNoTxContext(Up20250831120000, Down20250831120000)
}

func Up20250831120000(ctx context.Context, tx *sql.DB) error {
	// Create entity_types table
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE "entity_types" (
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
		    CONSTRAINT "entity_types_groups_entity_types" FOREIGN KEY ("group_entity_types") REFERENCES "groups" ("id") ON DELETE CASCADE
		);
		CREATE INDEX "entitytype_name" ON "entity_types" ("name");
		CREATE INDEX "entitytype_is_location" ON "entity_types" ("is_location");
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_types table: %w", err)
	}

	// Create entities table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE "entities" (
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
		    "sync_child_entities_locations" boolean NOT NULL DEFAULT false,
		    "serial_number" character varying NULL,
		    "model_number" character varying NULL,
		    "manufacturer" character varying NULL,
		    "lifetime_warranty" boolean NOT NULL DEFAULT false, 
			"warranty_expires" timestamptz NULL,
			"warranty_details" character varying NULL,
			"purchase_time" timestamptz NULL,
			"purchase_from" character varying NULL,
			"purchase_price" double precision NOT NULL DEFAULT 0,
			"sold_time" timestamptz NULL, "sold_to" character varying NULL,
			"sold_price" double precision NOT NULL DEFAULT 0,
			"sold_notes" character varying NULL,
			"entity_parent" uuid NULL,
			"entity_location" uuid NULL,
			"entity_type_entities" uuid NULL,
			"group_entities" uuid NOT NULL,
			PRIMARY KEY ("id"),
			CONSTRAINT "entities_entities_parent" FOREIGN KEY ("entity_parent") REFERENCES "entities" ("id") ON DELETE SET NULL,
			CONSTRAINT "entities_entities_location" FOREIGN KEY ("entity_location") REFERENCES "entities" ("id") ON DELETE SET NULL,
			CONSTRAINT "entities_entity_types_entities" FOREIGN KEY ("entity_type_entities") REFERENCES "entity_types" ("id") ON DELETE SET NULL,
			CONSTRAINT "entities_groups_entities" FOREIGN KEY ("group_entities") REFERENCES "groups" ("id") ON DELETE CASCADE);
		CREATE INDEX "entity_name" ON "entities" ("name");
		CREATE INDEX "entity_manufacturer" ON "entities" ("manufacturer");
		CREATE INDEX "entity_model_number" ON "entities" ("model_number");
		CREATE INDEX "entity_serial_number" ON "entities" ("serial_number");
		CREATE INDEX "entity_archived" ON "entities" ("archived");
		CREATE INDEX "entity_asset_id" ON "entities" ("asset_id");
	`)
	if err != nil {
		return fmt.Errorf("failed to create entities table: %w", err)
	}

	// Fetch all groups to create default entity types for each group
	groups, err := tx.QueryContext(ctx, `SELECT id FROM "groups"`)
	if err != nil {
		return fmt.Errorf("failed to query groups: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}(groups)

	// Process each group and create default entity types
	for groups.Next() {
		var groupID uuid.UUID
		if err := groups.Scan(&groupID); err != nil {
			return fmt.Errorf("failed to scan group ID: %w", err)
		}

		// Create default 'Item' entity type for this group
		itemTypeID := uuid.New()
		now := time.Now().UTC()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entity_types" ("id", "created_at", "updated_at", "name", "description", "is_location", "group_entity_types")
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, itemTypeID, now, now, "Item", "Default item type", false, groupID)
		if err != nil {
			return fmt.Errorf("failed to create Item entity type for group %s: %w", groupID, err)
		}

		// Create default 'Location' entity type for this group
		locTypeID := uuid.New()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entity_types" ("id", "created_at", "updated_at", "name", "description", "is_location", "group_entity_types")
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, locTypeID, now, now, "Location", "Default location type", true, groupID)
		if err != nil {
			return fmt.Errorf("failed to create Location entity type for group %s: %w", groupID, err)
		}

		// Migrate existing locations to entities
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entities" (
				"id", "created_at", "updated_at", "name", "description",
				"group_entities", "entity_parent", "entity_type_entities"
			)
			SELECT
				l."id", l."created_at", l."updated_at", l."name", l."description",
				l."group_locations", l."location_children", $1
			FROM "locations" l
			WHERE l."group_locations" = $2
		`, locTypeID, groupID)
		if err != nil {
			return fmt.Errorf("failed to migrate locations to entities for group %s: %w", groupID, err)
		}

		// Migrate existing items to entities
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entities" (
			    "id", "created_at", "updated_at", "name", "description",
			    "import_ref", "notes", "quantity", "insured", "archived",
				"asset_id", "serial_number", "model_number", "manufacturer", 
			    "lifetime_warranty", "warranty_expires", "warranty_details", "purchase_time",
				"purchase_from", "purchase_price", "sold_time", "sold_to",
				"sold_price", "sold_notes", "group_entities", "entity_parent", "sync_child_entities_locations",
				"entity_type_entities")
			SELECT
			    i."id", i."created_at", i."updated_at", i."name", i."description",
			    i."import_ref", i."notes", i."quantity", i."insured", i."archived",
			    i."asset_id", i."serial_number", i."model_number", i."manufacturer",
			    i."lifetime_warranty", i."warranty_expires", i."warranty_details", i."purchase_time",
			    i."purchase_from", i."purchase_price", i."sold_time", i."sold_to",
			    i."sold_price", i."sold_notes", i."group_items", i."item_children", i."sync_child_items_locations",
			    $1
			FROM "items" i WHERE i."group_items" = $2
		`, itemTypeID, groupID)
		if err != nil {
			return fmt.Errorf("failed to migrate items to entities for group %s: %w", groupID, err)
		}
	}

	_, err = tx.ExecContext(ctx, `
		ALTER TABLE "maintenance_entries" DROP CONSTRAINT "maintenance_entries_items_maintenance_entries",
		    ADD COLUMN "entity_id" uuid NOT NULL;
		UPDATE "maintenance_entries" SET "entity_id" = "item_id";
		ALTER TABLE "maintenance_entries" ADD CONSTRAINT "maintenance_entries_entities_maintenance_entries" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON DELETE CASCADE;
		ALTER TABLE "maintenance_entries" DROP COLUMN "item_id";
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate maintence_entries to use entities: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		ALTER TABLE "attachments" DROP CONSTRAINT "attachments_no_self_reference",
		    DROP CONSTRAINT "attachments_items_attachments",
		    ALTER COLUMN "mime_type" SET NOT NULL,
		    ADD COLUMN "entity_attachments" uuid NULL,
		    ADD CONSTRAINT "attachments_entities_attachments" FOREIGN KEY ("entity_attachments") REFERENCES "entities" ("id") ON DELETE CASCADE;	
		UPDATE "attachments" SET "entity_attachments" = "item_attachments";
		ALTER TABLE "attachments" DROP COLUMN "item_attachments";
		CREATE UNIQUE INDEX "attachments_attachment_thumbnail_key" ON "attachments" ("attachment_thumbnail");
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate attachments to use entities: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE TABLE "entity_fields" (
		    "id" uuid NOT NULL,
		    "created_at" timestamptz NOT NULL,
		    "updated_at" timestamptz NOT NULL,
		    "name" character varying NOT NULL,
		    "description" character varying NULL,
		    "type" character varying NOT NULL,
		    "text_value" character varying NULL,
		    "number_value" bigint NULL,
		    "boolean_value" boolean NOT NULL DEFAULT false,
		    "time_value" timestamptz NOT NULL,
		    "entity_fields" uuid NULL,
		    PRIMARY KEY ("id"),
		    CONSTRAINT "entity_fields_entities_fields" FOREIGN KEY ("entity_fields") REFERENCES "entities" ("id") ON DELETE CASCADE);

            INSERT INTO "entity_fields" ("id", "created_at", "updated_at", "name", "description", "type", "text_value", "number_value", "boolean_value", "time_value", "entity_fields")
            SELECT "id", "created_at", "updated_at", "name", "description", "type", "text_value", "number_value", "boolean_value", "time_value", item_fields."item_fields"
			FROM "item_fields";
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_fields table and migrate data: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE TABLE "label_entities"(
		    "label_id" uuid NOT NULL,
		    "entity_id" uuid NOT NULL,
		    PRIMARY KEY ("label_id", "entity_id"),
		    CONSTRAINT "label_entities_label_id" FOREIGN KEY ("label_id") REFERENCES "labels" ("id") ON DELETE CASCADE,
		    CONSTRAINT "label_entities_entity_id" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON DELETE CASCADE);

		INSERT INTO "label_entities" ("label_id", "entity_id")
		SELECT "label_id", "item_id" FROM "label_items";

	`)
	if err != nil {
		return fmt.Errorf("failed to create label_entities table and migrate data: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		DROP TABLE IF EXISTS "label_items";
		DROP TABLE IF EXISTS "item_fields";
		DROP TABLE IF EXISTS "items";
		DROP TABLE IF EXISTS "locations";
	`)
	if err != nil {
		return fmt.Errorf("failed to drop old tables: %w", err)
	}

	return nil
}

func Down20250831120000(ctx context.Context, tx *sql.DB) error {
	return nil
}
