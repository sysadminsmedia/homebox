package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits
func init() {
	goose.AddMigrationContext(Up20241226183416, Down20241226183416)
}

func Up20241226183416(ctx context.Context, tx *sql.Tx) error {
	// Check if the 'sync_child_items_locations' column exists in the 'items' table
	columnName := "sync_child_items_locations"
	query := `
		SELECT name 
		FROM pragma_table_info('items') 
		WHERE name = 'sync_child_items_locations';
	`
	err := tx.QueryRowContext(ctx, query).Scan(&columnName)
	if err != nil {
		// Column does not exist, proceed with migration
		_, err = tx.ExecContext(ctx, `
			PRAGMA foreign_keys = off;

			ALTER TABLE items
				ADD COLUMN sync_child_items_locations BOOLEAN NOT NULL DEFAULT FALSE;

			CREATE INDEX IF NOT EXISTS item_name           ON items(name);
			CREATE INDEX IF NOT EXISTS item_manufacturer   ON items(manufacturer);
			CREATE INDEX IF NOT EXISTS item_model_number   ON items(model_number);
			CREATE INDEX IF NOT EXISTS item_serial_number  ON items(serial_number);
			CREATE INDEX IF NOT EXISTS item_archived       ON items(archived);
			CREATE INDEX IF NOT EXISTS item_asset_id       ON items(asset_id);

			PRAGMA foreign_keys = on;
		`)
		if err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}
	return nil
}

func Down20241226183416(ctx context.Context, tx *sql.Tx) error {
	// This migration is a no-op for SQLite.
	return nil
}
