package migcheck_test

import (
	"database/sql"
	"encoding/json"
	"slices"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
)

// TestSqlitePermissionsBackfill upgrades a database holding pre-permission
// owner and member rows and asserts the backfill preserves what each could
// already do: owners get full access, while members and outstanding
// invitations get no collection-administration permissions.
func TestSqlitePermissionsBackfill(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:"+t.TempDir()+"/backfill.db?_fk=1")
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

	// Last migration before the permission system.
	if err := goose.UpTo(db, "sqlite3", 20260821000000); err != nil {
		t.Fatalf("goose up-to: %v", err)
	}

	const (
		groupID  = "00000000-0000-0000-0000-000000000001"
		ownerID  = "00000000-0000-0000-0000-000000000002"
		memberID = "00000000-0000-0000-0000-000000000003"
		inviteID = "00000000-0000-0000-0000-000000000004"
	)
	for _, q := range []string{
		`INSERT INTO groups (id, created_at, updated_at, name) VALUES ('` + groupID + `', datetime('now'), datetime('now'), 'g')`,
		`INSERT INTO users (id, created_at, updated_at, name, email, default_group_id) VALUES ('` + ownerID + `', datetime('now'), datetime('now'), 'o', 'o@example.com', '` + groupID + `')`,
		`INSERT INTO users (id, created_at, updated_at, name, email, default_group_id) VALUES ('` + memberID + `', datetime('now'), datetime('now'), 'm', 'm@example.com', '` + groupID + `')`,
		`INSERT INTO user_groups (user_id, group_id, role) VALUES ('` + ownerID + `', '` + groupID + `', 'owner')`,
		`INSERT INTO user_groups (user_id, group_id, role) VALUES ('` + memberID + `', '` + groupID + `', 'user')`,
		`INSERT INTO group_invitation_tokens (id, created_at, updated_at, token, expires_at, uses, group_invitation_tokens) VALUES ('` + inviteID + `', datetime('now'), datetime('now'), x'01', datetime('now', '+1 day'), 1, '` + groupID + `')`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("seed %q: %v", q, err)
		}
	}

	if err := goose.Up(db, "sqlite3"); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	perms := func(q string, args ...any) []string {
		t.Helper()
		var raw string
		if err := db.QueryRow(q, args...).Scan(&raw); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		var out []string
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			t.Fatalf("decode %q: %v", raw, err)
		}
		return out
	}

	owner := perms(`SELECT permissions FROM user_groups WHERE user_id = ?`, ownerID)
	if !slices.Equal(owner, []string{"*"}) {
		t.Errorf("owner permissions = %v, want [*]", owner)
	}

	adminOnly := []string{"*", "settings:manage", "members:manage", "permissions:manage", "collection:delete", "tools:wipe"}
	member := perms(`SELECT permissions FROM user_groups WHERE user_id = ?`, memberID)
	invite := perms(`SELECT permissions FROM group_invitation_tokens WHERE id = ?`, inviteID)
	for name, got := range map[string][]string{"member": member, "invitation": invite} {
		if len(got) == 0 {
			t.Errorf("%s permissions are empty; members must keep inventory access", name)
		}
		if !slices.Contains(got, "entity:*") {
			t.Errorf("%s permissions %v lack entity:*", name, got)
		}
		for _, p := range adminOnly {
			if slices.Contains(got, p) {
				t.Errorf("%s was backfilled with admin-only permission %q", name, p)
			}
		}
	}
}
