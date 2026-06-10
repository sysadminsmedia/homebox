// Package migcheck_test verifies the hand-written goose migrations apply
// cleanly: full Up from an empty database, then Down and re-Up of the most
// recent migration.
package migcheck_test

import (
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"

	// Register Go-based migrations (e.g. sync_children).
	_ "github.com/sysadminsmedia/homebox/backend/internal/data/migrations/sqlite3"
	_ "github.com/sysadminsmedia/homebox/backend/pkgs/cgofreesqlite"
)

func TestSqliteMigrations(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:"+t.TempDir()+"/test.db?_fk=1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	fs, err := migrations.Migrations("sqlite3")
	if err != nil {
		t.Fatal(err)
	}
	goose.SetBaseFS(fs)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "sqlite3"); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	// Sanity: permission-system tables and columns exist.
	for _, q := range []string{
		"SELECT permissions FROM user_groups LIMIT 1",
		"SELECT permissions FROM group_invitation_tokens LIMIT 1",
		"SELECT id, name, permissions, group_id FROM permission_groups LIMIT 1",
		"SELECT permission_group_id, user_id FROM permission_group_users LIMIT 1",
		"SELECT id, can_read, can_update, can_delete, can_attachments, user_id, permission_group_id, entity_id, group_id FROM access_grants LIMIT 1",
	} {
		if _, err := db.Exec(q); err != nil {
			t.Errorf("%s: %v", q, err)
		}
	}

	// The most recent migration must also roll back and re-apply.
	if err := goose.Down(db, "sqlite3"); err != nil {
		t.Fatalf("goose down: %v", err)
	}
	if err := goose.Up(db, "sqlite3"); err != nil {
		t.Fatalf("goose re-up: %v", err)
	}
}
