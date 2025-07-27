package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"time"
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

	// Process each group and create default entity types
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
	}

	return nil
}

func Down20250831120023(ctx context.Context, tx *sql.Tx) error {
	// Drop tables in reverse order to avoid foreign key constraints
	_, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS "entity_fields";`)
	if err != nil {
		return fmt.Errorf("failed to drop entity_fields table: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS "entities";`)
	if err != nil {
		return fmt.Errorf("failed to drop entities table: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS "entity_types";`)
	if err != nil {
		return fmt.Errorf("failed to drop entity_types table: %w", err)
	}

	return nil
}
