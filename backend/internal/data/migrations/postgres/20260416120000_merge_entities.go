package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits
func init() {
	goose.AddMigrationContext(Up20260402120000, Down20260402120000)
}

func Up20260402120000(ctx context.Context, tx *sql.Tx) error {
	// 1. Create entity_types table
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS "entity_types" (
			"id" uuid NOT NULL PRIMARY KEY,
			"created_at" timestamptz NOT NULL DEFAULT now(),
			"updated_at" timestamptz NOT NULL DEFAULT now(),
			"name" character varying(255) NOT NULL,
			"description" character varying(1000) NULL,
			"is_location" boolean NOT NULL DEFAULT false,
			"icon" character varying(255) NULL,
			"group_entity_types" uuid NOT NULL,
			CONSTRAINT "entity_types_groups_entity_types"
				FOREIGN KEY ("group_entity_types") REFERENCES "groups" ("id")
				ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("step 1: create entity_types table: %w", err)
	}

	// 2. Seed default entity types per group
	_, err = tx.ExecContext(ctx, `
		INSERT INTO entity_types (id, created_at, updated_at, name, description, is_location, group_entity_types)
		SELECT gen_random_uuid(), now(), now(), 'Location', '', true, g.id FROM groups g;
	`)
	if err != nil {
		return fmt.Errorf("step 2a: seed Location entity type: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO entity_types (id, created_at, updated_at, name, description, is_location, group_entity_types)
		SELECT gen_random_uuid(), now(), now(), 'Item', '', false, g.id FROM groups g;
	`)
	if err != nil {
		return fmt.Errorf("step 2b: seed Item entity type: %w", err)
	}

	// 3. Rename items -> entities
	_, err = tx.ExecContext(ctx, `ALTER TABLE "items" RENAME TO "entities";`)
	if err != nil {
		return fmt.Errorf("step 3: rename items to entities: %w", err)
	}

	// 4. Rename columns on entities table
	_, err = tx.ExecContext(ctx, `ALTER TABLE "entities" RENAME COLUMN "group_items" TO "group_entities";`)
	if err != nil {
		return fmt.Errorf("step 4a: rename group_items: %w", err)
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE "entities" RENAME COLUMN "item_children" TO "entity_children";`)
	if err != nil {
		return fmt.Errorf("step 4b: rename item_children: %w", err)
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE "entities" RENAME COLUMN "sync_child_items_locations" TO "sync_child_entity_locations";`)
	if err != nil {
		return fmt.Errorf("step 4c: rename sync_child_items_locations: %w", err)
	}

	// 5. Add entity_type_entities column (nullable initially)
	_, err = tx.ExecContext(ctx, `ALTER TABLE "entities" ADD COLUMN "entity_type_entities" uuid NULL;`)
	if err != nil {
		return fmt.Errorf("step 5: add entity_type_entities column: %w", err)
	}

	// 6. Set entity_type on all existing entities to "Item" type
	_, err = tx.ExecContext(ctx, `
		UPDATE entities SET entity_type_entities = et.id
		FROM entity_types et
		WHERE et.group_entity_types = entities.group_entities AND et.name = 'Item';
	`)
	if err != nil {
		return fmt.Errorf("step 6: set entity_type on existing entities: %w", err)
	}

	// 7. Check for UUID collision between locations and entities
	var collisionCount int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM locations l INNER JOIN entities e ON l.id = e.id`).Scan(&collisionCount)
	if err != nil {
		return fmt.Errorf("step 7: check UUID collisions: %w", err)
	}
	if collisionCount > 0 {
		return fmt.Errorf("UUID collision detected between locations and items: %d collisions", collisionCount)
	}

	// 8. INSERT locations as entities
	_, err = tx.ExecContext(ctx, `
		INSERT INTO entities (
			id, created_at, updated_at, name, description,
			quantity, insured, archived, asset_id, purchase_price, sold_price,
			sync_child_entity_locations, lifetime_warranty,
			group_entities, entity_type_entities, entity_children
		)
		SELECT
			l.id, l.created_at, l.updated_at, l.name, l.description,
			1, false, false, 0, 0, 0, false, false,
			l.group_locations, et.id, l.location_children
		FROM locations l
		JOIN entity_types et ON et.group_entity_types = l.group_locations AND et.name = 'Location';
	`)
	if err != nil {
		return fmt.Errorf("step 8: insert locations as entities: %w", err)
	}

	// 9. Reparent top-level items under their location
	_, err = tx.ExecContext(ctx, `
		UPDATE entities SET entity_children = location_items
		WHERE location_items IS NOT NULL AND entity_children IS NULL;
	`)
	if err != nil {
		return fmt.Errorf("step 9: reparent items under locations: %w", err)
	}

	// 10. Drop location_items column
	_, err = tx.ExecContext(ctx, `ALTER TABLE "entities" DROP COLUMN "location_items";`)
	if err != nil {
		return fmt.Errorf("step 10: drop location_items column: %w", err)
	}

	// 11. Make entity_type_entities NOT NULL + FK
	_, err = tx.ExecContext(ctx, `ALTER TABLE "entities" ALTER COLUMN "entity_type_entities" SET NOT NULL;`)
	if err != nil {
		return fmt.Errorf("step 11a: set entity_type_entities NOT NULL: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		ALTER TABLE "entities" ADD CONSTRAINT "entities_entity_types_entities"
		FOREIGN KEY ("entity_type_entities") REFERENCES "entity_types" ("id") ON DELETE RESTRICT;
	`)
	if err != nil {
		return fmt.Errorf("step 11b: add entity_type FK: %w", err)
	}

	// 12. Rename item_fields -> entity_fields
	_, err = tx.ExecContext(ctx, `ALTER TABLE "item_fields" RENAME TO "entity_fields";`)
	if err != nil {
		return fmt.Errorf("step 12a: rename item_fields to entity_fields: %w", err)
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE "entity_fields" RENAME COLUMN "item_fields" TO "entity_fields";`)
	if err != nil {
		return fmt.Errorf("step 12b: rename item_fields column: %w", err)
	}

	// 13. Rename tag_items -> tag_entities
	_, err = tx.ExecContext(ctx, `ALTER TABLE "tag_items" RENAME TO "tag_entities";`)
	if err != nil {
		return fmt.Errorf("step 13a: rename tag_items to tag_entities: %w", err)
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE "tag_entities" RENAME COLUMN "item_id" TO "entity_id";`)
	if err != nil {
		return fmt.Errorf("step 13b: rename item_id column: %w", err)
	}

	// 14. Rename item_templates -> entity_templates
	_, err = tx.ExecContext(ctx, `ALTER TABLE "item_templates" RENAME TO "entity_templates";`)
	if err != nil {
		return fmt.Errorf("step 14a: rename item_templates: %w", err)
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE "entity_templates" RENAME COLUMN "group_item_templates" TO "group_entity_templates";`)
	if err != nil {
		return fmt.Errorf("step 14b: rename group_item_templates: %w", err)
	}

	_, err = tx.ExecContext(ctx, `ALTER TABLE "entity_templates" RENAME COLUMN "item_template_location" TO "entity_template_location";`)
	if err != nil {
		return fmt.Errorf("step 14c: rename item_template_location: %w", err)
	}

	// 15. Rename FK in maintenance_entries
	_, err = tx.ExecContext(ctx, `ALTER TABLE "maintenance_entries" RENAME COLUMN "item_id" TO "entity_id";`)
	if err != nil {
		return fmt.Errorf("step 15: rename maintenance_entries item_id: %w", err)
	}

	// 16. Rename FK in attachments
	_, err = tx.ExecContext(ctx, `ALTER TABLE "attachments" RENAME COLUMN "item_attachments" TO "entity_attachments";`)
	if err != nil {
		return fmt.Errorf("step 16: rename attachments item_attachments: %w", err)
	}

	// 17. Rename FK in template_fields
	_, err = tx.ExecContext(ctx, `ALTER TABLE "template_fields" RENAME COLUMN "item_template_fields" TO "entity_template_fields";`)
	if err != nil {
		return fmt.Errorf("step 17: rename template_fields item_template_fields: %w", err)
	}

	// 18. Update entity_templates location FK to reference entities instead of locations
	_, err = tx.ExecContext(ctx, `
		ALTER TABLE "entity_templates" DROP CONSTRAINT IF EXISTS "item_templates_locations_location";
	`)
	if err != nil {
		return fmt.Errorf("step 18a: drop old location FK: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		ALTER TABLE "entity_templates" ADD CONSTRAINT "entity_templates_entities_location"
		FOREIGN KEY ("entity_template_location") REFERENCES "entities" ("id") ON DELETE SET NULL;
	`)
	if err != nil {
		return fmt.Errorf("step 18b: add new location FK: %w", err)
	}

	// 19. Drop locations table
	_, err = tx.ExecContext(ctx, `DROP TABLE IF EXISTS "locations";`)
	if err != nil {
		return fmt.Errorf("step 19: drop locations table: %w", err)
	}

	return nil
}

func Down20260402120000(ctx context.Context, tx *sql.Tx) error {
	return nil
}
