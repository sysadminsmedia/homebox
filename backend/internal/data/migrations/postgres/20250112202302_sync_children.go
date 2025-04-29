package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pressly/goose/v3"
)

//nolint:gochecknoinits
func init() {
	goose.AddMigrationContext(Up20235011220230202, Down20235011220230202)
}

func Up20235011220230202(ctx context.Context, tx *sql.Tx) error {
	columnName := "sync_child_items_locations"
	query := `
		SELECT column_name 
		FROM information_schema.columns
		WHERE table_name = 'items' AND column_name = 'sync_child_items_locations';
	`
	err := tx.QueryRowContext(ctx, query).Scan(&columnName)
	if err != nil {
		// Column does not exist, proceed with migration
		_, err = tx.ExecContext(ctx, `
			ALTER TABLE "items" ADD COLUMN"sync_child_items_locations" boolean NOT NULL DEFAULT false;
		`)
		if err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}
	return nil
}

func Down20235011220230202(ctx context.Context, tx *sql.Tx) error {
	// This migration is a no-op for Postgres.
	return nil
}
