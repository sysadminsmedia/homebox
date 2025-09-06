package sqlite3

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
	goose.AddMigrationContext(Up20250831120023, Down20250831120023)
}

func Up20250831120023(ctx context.Context, tx *sql.Tx) error {
	// Create entity_types table
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "entity_types" (
			"id" text NOT NULL,
			"created_at" datetime NOT NULL,
			"updated_at" datetime NOT NULL,
			"name" text NOT NULL,
			"description" text NULL,
			"icon" text NULL,
			"color" text NULL,
			"is_location" integer NOT NULL DEFAULT 0,
			"group_entity_types" text NOT NULL,
			PRIMARY KEY ("id"),
			FOREIGN KEY ("group_entity_types") REFERENCES "groups" ("id") ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_types table: %w", err)
	}

	// Create entities table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "entities" (
			"id" text NOT NULL,
			"created_at" datetime NOT NULL,
			"updated_at" datetime NOT NULL,
			"name" text NOT NULL,
			"description" text NULL,
			"import_ref" text NULL,
			"notes" text NULL,
			"quantity" integer NOT NULL DEFAULT 1,
			"insured" integer NOT NULL DEFAULT 0,
			"archived" integer NOT NULL DEFAULT 0,
			"asset_id" integer NOT NULL DEFAULT 0,
			"serial_number" text NULL,
			"model_number" text NULL,
			"manufacturer" text NULL,
			"lifetime_warranty" integer NOT NULL DEFAULT 0,
			"warranty_expires" datetime NULL,
			"warranty_details" text NULL,
			"purchase_time" datetime NULL,
			"purchase_from" text NULL,
			"purchase_price" real NOT NULL DEFAULT 0,
			"sold_time" datetime NULL,
			"sold_to" text NULL,
			"sold_price" real NOT NULL DEFAULT 0,
			"sold_notes" text NULL,
			"group_entities" text NOT NULL,
			"entity_children" text NULL,
			"entity_parent" text NULL,
			"entity_type" text NOT NULL,
			PRIMARY KEY ("id"),
			FOREIGN KEY ("group_entities") REFERENCES "groups" ("id") ON DELETE CASCADE,
			FOREIGN KEY ("entity_children") REFERENCES "entities" ("id") ON DELETE SET NULL,
			FOREIGN KEY ("entity_parent") REFERENCES "entities" ("id") ON DELETE SET NULL,
			FOREIGN KEY ("entity_type") REFERENCES "entity_types" ("id") ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create entities table: %w", err)
	}

	// Create entity_fields table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "entity_fields" (
			"id" text NOT NULL,
			"created_at" datetime NOT NULL,
			"updated_at" datetime NOT NULL,
			"name" text NOT NULL,
			"description" text NULL,
			"type" text NOT NULL,
			"text_value" text NULL,
			"number_value" integer NULL,
			"boolean_value" integer NOT NULL DEFAULT 0,
			"time_value" datetime NULL,
			"entity_fields" text NULL,
			PRIMARY KEY ("id"),
			FOREIGN KEY ("entity_fields") REFERENCES "entities" ("id") ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_fields table: %w", err)
	}

	// Fetch all groups to create default entity types for each group
	rows, err := tx.QueryContext(ctx, `SELECT id FROM "groups"`)
	if err != nil {
		return fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	// Process each group and create default entity types, and perform migrations that depend on entity types information
	for rows.Next() {
		var groupID string
		if err := rows.Scan(&groupID); err != nil {
			return fmt.Errorf("failed to scan group ID: %w", err)
		}

		// Create default 'Item' entity type for this group
		itemTypeID := uuid.New().String()
		now := time.Now().UTC()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entity_types" ("id", "created_at", "updated_at", "name", "description", "is_location", "group_entity_types")
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, itemTypeID, now, now, "Item", "Default item type", 0, groupID)
		if err != nil {
			return fmt.Errorf("failed to create Item entity type for group %s: %w", groupID, err)
		}

		// Create default 'Location' entity type for this group
		locTypeID := uuid.New().String()
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entity_types" ("id", "created_at", "updated_at", "name", "description", "is_location", "group_entity_types")
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, locTypeID, now, now, "Location", "Default location type", 1, groupID)
		if err != nil {
			return fmt.Errorf("failed to create Location entity type for group %s: %w", groupID, err)
		}

		// Migrate existing locations to entities
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entities" (
				"id", "created_at", "updated_at", "name", "description",
				"group_entities", "entity_children", "entity_type"
			)
			SELECT
				l."id", l."created_at", l."updated_at", l."name", l."description",
				l."group_locations", l."location_children", ?
			FROM "locations" l
			WHERE l."group_locations" = ?
		`, locTypeID, groupID)
		if err != nil {
			return fmt.Errorf("failed to migrate locations to entities for group %s: %w", groupID, err)
		}

		// Migrate existing items to entities
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entities" (
				"id", "created_at", "updated_at", "name", "description",
				"import_ref", "notes", "quantity", "insured", "archived", "asset_id",
				"serial_number", "model_number", "manufacturer", "lifetime_warranty",
				"warranty_expires", "warranty_details", "purchase_time", "purchase_from",
				"purchase_price", "sold_time", "sold_to", "sold_price", "sold_notes",
				"group_entities", "entity_type"
			)
			SELECT
				i."id", i."created_at", i."updated_at", i."name", i."description",
				i."import_ref", i."notes", i."quantity", i."insured", i."archived", i."asset_id",
				i."serial_number", i."model_number", i."manufacturer", i."lifetime_warranty",
				i."warranty_expires", i."warranty_details", i."purchase_time", i."purchase_from",
				i."purchase_price", i."sold_time", i."sold_to", i."sold_price", i."sold_notes",
				i."group_items", ?
			FROM "items" i
			WHERE i."group_items" = ?
		`, itemTypeID, groupID)
		if err != nil {
			return fmt.Errorf("failed to migrate items to entities for group %s: %w", groupID, err)
		}

		// Migrate existing locations to entities
		_, err = tx.ExecContext(ctx, `
			INSERT INTO "entities" (
				"id", "created_at", "updated_at", "name", "description",
				"group_entities", "entity_type"
			)
			SELECT l.id, l.created_at, l.updated_at, l.name, l.description, l.group_locations, ? FROM "locations" l WHERE l."group_locations" = ?
		`, locTypeID, groupID)
		if err != nil {
			return fmt.Errorf("failed to migrate locations to entities for group %s: %w", groupID, err)
		}
	}

	// Drop old tables
	_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS "items"`)
	if err != nil {
		return fmt.Errorf("failed to drop items table: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS "locations"`)
	if err != nil {
		return fmt.Errorf("failed to drop locations table: %w", err)
	}

	return nil
}

func Down20250831120023(ctx context.Context, tx *sql.Tx) error {
	return nil
}
