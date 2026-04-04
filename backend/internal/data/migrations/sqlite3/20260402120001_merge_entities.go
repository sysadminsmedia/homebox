package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits
func init() {
	goose.AddMigrationNoTxContext(Up20260402120001, Down20260402120001)
}

func Up20260402120001(ctx context.Context, db *sql.DB) error {
	// Disable FK constraints for table recreation
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys=OFF;`); err != nil {
		return fmt.Errorf("disable FK: %w", err)
	}

	// 1. Create entity_types table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS entity_types (
			id               uuid     not null primary key,
			created_at       datetime not null,
			updated_at       datetime not null,
			name             text     not null,
			description      text,
			is_location      bool     default false not null,
			icon             text,
			group_entity_types uuid   not null
				constraint entity_types_groups_entity_types
					references groups
					on delete cascade
		);
	`); err != nil {
		return fmt.Errorf("step 1: create entity_types: %w", err)
	}

	// 2. Seed default entity types per group
	if _, err := db.ExecContext(ctx, `
		INSERT INTO entity_types (id, created_at, updated_at, name, description, is_location, group_entity_types)
		SELECT lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6))),
			   datetime('now'), datetime('now'), 'Location', '', 1, g.id
		FROM groups g;
	`); err != nil {
		return fmt.Errorf("step 2a: seed Location type: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO entity_types (id, created_at, updated_at, name, description, is_location, group_entity_types)
		SELECT lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab', abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6))),
			   datetime('now'), datetime('now'), 'Item', '', 0, g.id
		FROM groups g;
	`); err != nil {
		return fmt.Errorf("step 2b: seed Item type: %w", err)
	}

	// 3. Save location_items mapping before dropping
	if _, err := db.ExecContext(ctx, `
		CREATE TEMP TABLE _location_mapping AS
		SELECT id, location_items FROM items WHERE location_items IS NOT NULL AND item_children IS NULL;
	`); err != nil {
		return fmt.Errorf("step 3: save location mapping: %w", err)
	}

	// 4. Create new entities table
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE entities (
			id                          uuid                  not null primary key,
			created_at                  datetime              not null,
			updated_at                  datetime              not null,
			name                        text                  not null,
			description                 text,
			import_ref                  text,
			notes                       text,
			quantity                    real    default 1     not null,
			insured                     bool    default false not null,
			archived                    bool    default false not null,
			asset_id                    integer default 0     not null,
			sync_child_entity_locations bool    default false not null,
			serial_number               text,
			model_number                text,
			manufacturer                text,
			lifetime_warranty           bool    default false not null,
			warranty_expires            datetime,
			warranty_details            text,
			purchase_time               datetime,
			purchase_from               text,
			purchase_price              real    default 0     not null,
			sold_time                   datetime,
			sold_to                     text,
			sold_price                  real    default 0     not null,
			sold_notes                  text,
			group_entities              uuid                  not null
				constraint entities_groups_entities
					references groups
					on delete cascade,
			entity_type_entities        uuid                  not null
				constraint entities_entity_types_entities
					references entity_types
					on delete restrict,
			entity_children             uuid
				constraint entities_entities_children
					references entities
					on delete set null
		);
	`); err != nil {
		return fmt.Errorf("step 4: create entities table: %w", err)
	}

	// 5. INSERT items into entities
	if _, err := db.ExecContext(ctx, `
		INSERT INTO entities (
			id, created_at, updated_at, name, description,
			import_ref, notes, quantity, insured, archived, asset_id,
			sync_child_entity_locations,
			serial_number, model_number, manufacturer,
			lifetime_warranty, warranty_expires, warranty_details,
			purchase_time, purchase_from, purchase_price,
			sold_time, sold_to, sold_price, sold_notes,
			group_entities, entity_type_entities, entity_children
		)
		SELECT
			i.id, i.created_at, i.updated_at, i.name, i.description,
			i.import_ref, i.notes, i.quantity, i.insured, i.archived, i.asset_id,
			i.sync_child_items_locations,
			i.serial_number, i.model_number, i.manufacturer,
			i.lifetime_warranty, i.warranty_expires, i.warranty_details,
			i.purchase_time, i.purchase_from, i.purchase_price,
			i.sold_time, i.sold_to, i.sold_price, i.sold_notes,
			i.group_items, et.id, i.item_children
		FROM items i
		JOIN entity_types et ON et.group_entity_types = i.group_items AND et.name = 'Item';
	`); err != nil {
		return fmt.Errorf("step 5: insert items as entities: %w", err)
	}

	// 6. INSERT locations into entities
	if _, err := db.ExecContext(ctx, `
		INSERT INTO entities (
			id, created_at, updated_at, name, description,
			quantity, insured, archived, asset_id, purchase_price, sold_price,
			sync_child_entity_locations, lifetime_warranty,
			group_entities, entity_type_entities, entity_children
		)
		SELECT
			l.id, l.created_at, l.updated_at, l.name, l.description,
			1, 0, 0, 0, 0, 0, 0, 0,
			l.group_locations, et.id, l.location_children
		FROM locations l
		JOIN entity_types et ON et.group_entity_types = l.group_locations AND et.name = 'Location';
	`); err != nil {
		return fmt.Errorf("step 6: insert locations as entities: %w", err)
	}

	// 7. Reparent items under their former location
	if _, err := db.ExecContext(ctx, `
		UPDATE entities SET entity_children = (
			SELECT location_items FROM _location_mapping WHERE _location_mapping.id = entities.id
		)
		WHERE EXISTS (SELECT 1 FROM _location_mapping WHERE _location_mapping.id = entities.id);
	`); err != nil {
		return fmt.Errorf("step 7: reparent items: %w", err)
	}

	// 8. Clean up temp table and drop old tables
	if _, err := db.ExecContext(ctx, `DROP TABLE _location_mapping;`); err != nil {
		return fmt.Errorf("step 8a: drop temp table: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE items;`); err != nil {
		return fmt.Errorf("step 8b: drop items: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE locations;`); err != nil {
		return fmt.Errorf("step 8c: drop locations: %w", err)
	}

	// 9. Recreate entity_fields (was item_fields)
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE entity_fields (
			id            uuid               not null primary key,
			created_at    datetime           not null,
			updated_at    datetime           not null,
			name          text               not null,
			description   text,
			type          text               not null,
			text_value    text,
			number_value  integer,
			boolean_value bool default false not null,
			time_value    datetime           not null,
			entity_fields uuid
				constraint entity_fields_entities_fields
					references entities
					on delete cascade
		);
	`); err != nil {
		return fmt.Errorf("step 9a: create entity_fields: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO entity_fields (id, created_at, updated_at, name, description, type, text_value, number_value, boolean_value, time_value, entity_fields)
		SELECT id, created_at, updated_at, name, description, type, text_value, number_value, boolean_value, time_value, item_fields FROM item_fields;
	`); err != nil {
		return fmt.Errorf("step 9b: copy item_fields: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE item_fields;`); err != nil {
		return fmt.Errorf("step 9c: drop item_fields: %w", err)
	}

	// 10. Recreate tag_entities (was tag_items)
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE tag_entities (
			tag_id    uuid not null
				constraint tag_entities_tag_id
					references tags
					on delete cascade,
			entity_id uuid not null
				constraint tag_entities_entity_id
					references entities
					on delete cascade,
			primary key (tag_id, entity_id)
		);
	`); err != nil {
		return fmt.Errorf("step 10a: create tag_entities: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO tag_entities (tag_id, entity_id)
		SELECT tag_id, item_id FROM tag_items;
	`); err != nil {
		return fmt.Errorf("step 10b: copy tag_items: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE tag_items;`); err != nil {
		return fmt.Errorf("step 10c: drop tag_items: %w", err)
	}

	// 11. Recreate entity_templates (was item_templates)
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE entity_templates (
			id                           uuid                  not null primary key,
			created_at                   datetime              not null,
			updated_at                   datetime              not null,
			name                         text                  not null,
			description                  text,
			notes                        text,
			default_quantity             real    default 1     not null,
			default_insured              bool    default false not null,
			default_name                 text,
			default_description          text,
			default_manufacturer         text,
			default_model_number         text,
			default_lifetime_warranty    bool    default false not null,
			default_warranty_details     text,
			include_warranty_fields      bool    default false not null,
			include_purchase_fields      bool    default false not null,
			include_sold_fields          bool    default false not null,
			default_tag_ids              json,
			entity_template_location     uuid
				constraint entity_templates_entities_location
					references entities
					on delete set null,
			group_entity_templates       uuid                  not null
				constraint entity_templates_groups_entity_templates
					references groups
					on delete cascade
		);
	`); err != nil {
		return fmt.Errorf("step 11a: create entity_templates: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO entity_templates (
			id, created_at, updated_at, name, description, notes,
			default_quantity, default_insured, default_name, default_description,
			default_manufacturer, default_model_number,
			default_lifetime_warranty, default_warranty_details,
			include_warranty_fields, include_purchase_fields, include_sold_fields,
			default_tag_ids, entity_template_location, group_entity_templates
		)
		SELECT
			id, created_at, updated_at, name, description, notes,
			default_quantity, default_insured, default_name, default_description,
			default_manufacturer, default_model_number,
			default_lifetime_warranty, default_warranty_details,
			include_warranty_fields, include_purchase_fields, include_sold_fields,
			default_tag_ids, item_template_location, group_item_templates
		FROM item_templates;
	`); err != nil {
		return fmt.Errorf("step 11b: copy item_templates: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE item_templates;`); err != nil {
		return fmt.Errorf("step 11c: drop item_templates: %w", err)
	}

	// 12. Recreate template_fields with FK to entity_templates
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE template_fields_new (
			id                     uuid               not null primary key,
			created_at             datetime           not null,
			updated_at             datetime           not null,
			name                   text               not null,
			description            text,
			type                   text               not null,
			text_value             text,
			number_value           integer,
			boolean_value          bool default false,
			time_value             datetime,
			entity_template_fields uuid
				constraint template_fields_entity_templates_fields
					references entity_templates
					on delete cascade
		);
	`); err != nil {
		return fmt.Errorf("step 12a: create template_fields_new: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO template_fields_new (id, created_at, updated_at, name, description, type, text_value, number_value, boolean_value, time_value, entity_template_fields)
		SELECT id, created_at, updated_at, name, description, type, text_value, number_value, boolean_value, time_value, item_template_fields FROM template_fields;
	`); err != nil {
		return fmt.Errorf("step 12b: copy template_fields: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE template_fields;`); err != nil {
		return fmt.Errorf("step 12c: drop template_fields: %w", err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE template_fields_new RENAME TO template_fields;`); err != nil {
		return fmt.Errorf("step 12d: rename template_fields: %w", err)
	}

	// 13. Recreate maintenance_entries with entity_id FK
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE maintenance_entries_new (
			id             uuid           not null primary key,
			created_at     datetime       not null,
			updated_at     datetime       not null,
			date           datetime,
			scheduled_date datetime,
			name           text           not null,
			description    text,
			cost           real default 0 not null,
			entity_id      uuid           not null
				constraint maintenance_entries_entities_maintenance_entries
					references entities
					on delete cascade
		);
	`); err != nil {
		return fmt.Errorf("step 13a: create maintenance_entries_new: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO maintenance_entries_new (id, created_at, updated_at, date, scheduled_date, name, description, cost, entity_id)
		SELECT id, created_at, updated_at, date, scheduled_date, name, description, cost, item_id FROM maintenance_entries;
	`); err != nil {
		return fmt.Errorf("step 13b: copy maintenance_entries: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE maintenance_entries;`); err != nil {
		return fmt.Errorf("step 13c: drop maintenance_entries: %w", err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE maintenance_entries_new RENAME TO maintenance_entries;`); err != nil {
		return fmt.Errorf("step 13d: rename maintenance_entries: %w", err)
	}

	// 14. Recreate attachments with entity_attachments FK
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE attachments_new (
			id                   uuid                                    not null primary key,
			created_at           datetime                                not null,
			updated_at           datetime                                not null,
			type                 text    default 'attachment'            not null,
			"primary"            bool    default false                   not null,
			path                 text                                    not null,
			title                text                                    not null,
			mime_type            text    default 'application/octet-stream' not null,
			entity_attachments   uuid
				constraint attachments_entities_attachments
					references entities
					on delete cascade,
			attachment_thumbnail uuid
				constraint attachments_attachments_thumbnail
					references attachments_new
					on delete set null
		);
	`); err != nil {
		return fmt.Errorf("step 14a: create attachments_new: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO attachments_new (id, created_at, updated_at, type, "primary", path, title, mime_type, entity_attachments, attachment_thumbnail)
		SELECT id, created_at, updated_at, type, "primary", path, title, mime_type, item_attachments, attachment_thumbnail FROM attachments;
	`); err != nil {
		return fmt.Errorf("step 14b: copy attachments: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE attachments;`); err != nil {
		return fmt.Errorf("step 14c: drop attachments: %w", err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE attachments_new RENAME TO attachments;`); err != nil {
		return fmt.Errorf("step 14d: rename attachments: %w", err)
	}

	// 15. Recreate indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS entity_name ON entities(name);`,
		`CREATE INDEX IF NOT EXISTS entity_manufacturer ON entities(manufacturer);`,
		`CREATE INDEX IF NOT EXISTS entity_model_number ON entities(model_number);`,
		`CREATE INDEX IF NOT EXISTS entity_serial_number ON entities(serial_number);`,
		`CREATE INDEX IF NOT EXISTS entity_archived ON entities(archived);`,
		`CREATE INDEX IF NOT EXISTS entity_asset_id ON entities(asset_id);`,
		`CREATE INDEX IF NOT EXISTS idx_attachments_entity_id ON attachments(entity_attachments);`,
		`CREATE INDEX IF NOT EXISTS idx_attachments_path ON attachments(path);`,
		`CREATE INDEX IF NOT EXISTS idx_attachments_type ON attachments(type);`,
		`CREATE INDEX IF NOT EXISTS idx_attachments_thumbnail ON attachments(attachment_thumbnail);`,
	}
	for i, idx := range indexes {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("step 15: create index %d: %w", i, err)
		}
	}

	// Re-enable FK constraints
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys=ON;`); err != nil {
		return fmt.Errorf("re-enable FK: %w", err)
	}

	return nil
}

func Down20260402120001(ctx context.Context, db *sql.DB) error {
	return nil
}
