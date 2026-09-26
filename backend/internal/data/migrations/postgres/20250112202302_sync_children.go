package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits
func init() {
	goose.AddMigrationContext(Up20250112202302, Down20250112202302)
}

func Up20250112202302(ctx context.Context, tx *sql.Tx) error {
	// The table_schema filter is load-bearing. information_schema.columns spans
	// every schema in the database, so an unqualified table_name = 'items'
	// lookup also matches an unrelated items table living in another schema.
	// When that happened the guard concluded the column already existed, skipped
	// the ALTER, and left the real table without it — which only surfaced later
	// as a hard failure renaming the missing column in 20260416120000.
	query := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'items'
		  AND column_name = 'sync_child_items_locations';
	`
	err := tx.QueryRowContext(ctx, query).Scan(new("sync_child_items_locations"))
	if err != nil {
		// Column does not exist, proceed with migration
		_, err = tx.ExecContext(ctx, `
			ALTER TABLE "items" ADD COLUMN "sync_child_items_locations" boolean NOT NULL DEFAULT false;
		`)
		if err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}
	return nil
}

func Down20250112202302(ctx context.Context, tx *sql.Tx) error {
	// This migration is a no-op for Postgres.
	return nil
}
