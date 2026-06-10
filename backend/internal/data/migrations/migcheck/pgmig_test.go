package migcheck_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
	_ "github.com/sysadminsmedia/homebox/backend/internal/data/migrations/postgres"
)

func TestPostgresMigrations(t *testing.T) {
	dsn := os.Getenv("HBX_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("HBX_TEST_PG_DSN not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	fs, err := migrations.Migrations("postgres")
	if err != nil {
		t.Fatal(err)
	}
	goose.SetBaseFS(fs)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "postgres"); err != nil {
		t.Fatalf("goose up: %v", err)
	}
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
	if err := goose.Down(db, "postgres"); err != nil {
		t.Fatalf("goose down: %v", err)
	}
	if err := goose.Up(db, "postgres"); err != nil {
		t.Fatalf("goose re-up: %v", err)
	}
}
