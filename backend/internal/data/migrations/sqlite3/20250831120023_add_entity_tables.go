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
	_, err := tx.ExecContext(ctx, `PRAGMA foreign_keys = OFF;`)
	if err != nil {
		return fmt.Errorf("failed to disable foreign keys: %w", err)
	}

	// Create entity_types table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE entities (
		    id uuid NOT NULL,
		    created_at datetime NOT NULL,
		    updated_at datetime NOT NULL,
		    name text NOT NULL,
		    description text NULL,
		    import_ref text NULL,
		    notes text NULL,
		    quantity integer NOT NULL DEFAULT (1),
		    insured bool NOT NULL DEFAULT (false),
		    archived bool NOT NULL DEFAULT (false),
		    asset_id integer NOT NULL DEFAULT (0),
		    sync_child_entities_locations bool NOT NULL DEFAULT (false),
		    serial_number text NULL,
		    model_number text NULL,
		    manufacturer text NULL,
		    lifetime_warranty bool NOT NULL DEFAULT (false),
		    warranty_expires datetime NULL,
		    warranty_details text NULL,
		    purchase_time datetime NULL,
		    purchase_from text NULL,
		    purchase_price real NOT NULL DEFAULT (0),
		    sold_time datetime NULL,
		    sold_to text NULL,
		    sold_price real NOT NULL DEFAULT (0),
		    sold_notes text NULL,
		    entity_parent uuid NULL,
		    entity_location uuid NULL,
		    entity_type_entities uuid NULL,
		    group_entities uuid NOT NULL, 
		    PRIMARY KEY (id), 
		    CONSTRAINT entities_entities_parent FOREIGN KEY (entity_parent) REFERENCES entities (id) ON DELETE SET NULL,
		    CONSTRAINT entities_entities_location FOREIGN KEY (entity_location) REFERENCES entities (id) ON DELETE SET NULL, 
		    CONSTRAINT entities_entity_types_entities FOREIGN KEY (entity_type_entities) REFERENCES entity_types (id) ON DELETE SET NULL,
		    CONSTRAINT entities_groups_entities FOREIGN KEY (group_entities) REFERENCES groups (id) ON DELETE CASCADE);
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_types table: %w", err)
	}

	// Create entities table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE entity_types (
		    id uuid NOT NULL,
		    created_at datetime NOT NULL,
		    updated_at datetime NOT NULL,
		    name text NOT NULL,
		    description text NULL,
		    icon text NULL,
		    color text NULL,
		    is_location bool NOT NULL DEFAULT (false),
		    group_entity_types uuid NOT NULL, 
		    PRIMARY KEY (id), 
		    CONSTRAINT entity_types_groups_entity_types FOREIGN KEY (group_entity_types) REFERENCES groups (id) ON DELETE CASCADE);
			CREATE INDEX entitytype_name ON entity_types (name);
			CREATE INDEX entitytype_is_location ON entity_types (is_location);
	`)
	if err != nil {
		return fmt.Errorf("failed to create entities table: %w", err)
	}

	// Create entity_fields table
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE entity_fields (
		    id uuid NOT NULL,
		    created_at datetime NOT NULL,
		    updated_at datetime NOT NULL,
		    name text NOT NULL,
		    description text NULL,
		    type text NOT NULL,
		    text_value text NULL,
		    number_value integer NULL,
		    boolean_value bool NOT NULL DEFAULT (false),
		    time_value datetime NOT NULL,
		    entity_fields uuid NULL, 
		    PRIMARY KEY (id), 
		    CONSTRAINT entity_fields_entities_fields FOREIGN KEY (entity_fields) REFERENCES entities (id) ON DELETE CASCADE);
        INSERT INTO entity_fields (id, created_at, updated_at, name, description, type, text_value, number_value, boolean_value, time_value, entity_fields)
        	SELECT id, created_at, updated_at, name, description, type, text_value, number_value, boolean_value, time_value, item_fields FROM item_fields;
		DROP TABLE item_fields;
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_fields table: %w", err)
	}

	// Fetch all groups to create default entity types for each group
	rows, err := tx.QueryContext(ctx, `SELECT id FROM "groups"`)
	if err != nil {
		return fmt.Errorf("failed to query groups: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

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
				"group_entities", "entity_parent", "entity_type_entities"
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
				"group_entities", "entity_type_entities"
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

	// Convert other tables that reference items or locations to reference entities instead
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE attachments_tmp (
		    id uuid NOT NULL,
		    created_at datetime NOT NULL,
		    updated_at datetime NOT NULL,
		    type text NOT NULL DEFAULT ('attachment'),
		    "primary" bool NOT NULL DEFAULT (false),
		    title text NOT NULL DEFAULT (''),
		    path text NOT NULL DEFAULT (''),
		    mime_type text NOT NULL DEFAULT ('application/octet-stream'),
		    attachment_thumbnail uuid NULL,
		    entity_attachments uuid NULL, 
		    PRIMARY KEY (id),
		    CONSTRAINT attachments_attachments_thumbnail FOREIGN KEY (attachment_thumbnail) REFERENCES attachments (id) ON DELETE SET NULL,
		    CONSTRAINT attachments_entities_attachments FOREIGN KEY (entity_attachments) REFERENCES entities (id) ON DELETE CASCADE);
		
		INSERT INTO "attachments_tmp" (
			"id", "created_at", "updated_at", "type", "primary", "title", "path",
			"mime_type", "attachment_thumbnail", "entity_attachments") 
			SELECT id, created_at, updated_at, type, "primary", title, path, mime_type, attachment_thumbnail, item_attachments FROM attachments;

		DROP TABLE attachments;

		ALTER TABLE attachments_tmp RENAME TO attachments;
		
		CREATE UNIQUE INDEX attachments_attachment_thumbnail_key ON attachments (attachment_thumbnail);
		CREATE INDEX idx_attachments_entity_id ON attachments(entity_attachments);
		CREATE INDEX idx_attachments_path ON attachments(path);
		CREATE INDEX idx_attachments_thumbnail ON attachments(attachment_thumbnail);
    `)
	if err != nil {
		return fmt.Errorf("failed to migrate attachments to reference entities: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE TABLE maintenance_entries_tmp (
		    id uuid NOT NULL,
		    created_at datetime NOT NULL,
		    updated_at datetime NOT NULL,
		    date datetime NULL,
		    scheduled_date datetime NULL,
		    name text NOT NULL,
		    description text NULL,
		    cost real NOT NULL DEFAULT (0),
		    entity_id uuid NOT NULL,
		    PRIMARY KEY (id),
		    CONSTRAINT maintenance_entries_entities_maintenance_entries FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE);

		INSERT INTO maintenance_entries_tmp (
		    "id", "created_at", "updated_at", "date", "scheduled_date", "name",
			"description", "cost", "entity_id")
			SELECT id, created_at, updated_at, date, scheduled_date, name, description, cost, item_id FROM maintenance_entries;

		DROP TABLE maintenance_entries;

		ALTER TABLE maintenance_entries_tmp RENAME TO maintenance_entries;
	`)
	if err != nil {
		return fmt.Errorf("failed to migrate maintenance entries to reference entities: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		CREATE TABLE label_entities (
		    label_id uuid NOT NULL,
		    entity_id uuid NOT NULL, PRIMARY KEY (label_id, entity_id), 
		    CONSTRAINT label_entities_label_id FOREIGN KEY (label_id) REFERENCES labels (id) ON DELETE CASCADE, 
		    CONSTRAINT label_entities_entity_id FOREIGN KEY (entity_id) REFERENCES entities (id) ON DELETE CASCADE);

		INSERT INTO label_entities (label_id, entity_id)
			SELECT label_id, item_id FROM label_items;

		DROP TABLE label_items;
    `)
	if err != nil {
		return fmt.Errorf("failed to migrate labels to reference entities: %w", err)
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

	_, err = tx.ExecContext(ctx, `PRAGMA foreign_keys = ON;`)
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return nil
}

func Down20250831120023(ctx context.Context, tx *sql.Tx) error {
	return nil
}
