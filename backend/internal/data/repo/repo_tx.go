package repo

import (
	"context"

	"entgo.io/ent/dialect"
	"github.com/google/uuid"

	"github.com/sysadminsmedia/homebox/backend/internal/data/authz"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
)

// lockTenantAdmins serializes the admin-invariant transactions for one tenant
// by taking a row lock on the tenant's group row. Without it two concurrent
// transactions that each demote or remove a *different* administrator can both
// observe the other admin still present and commit, leaving the tenant with
// members but zero permissions:manage holders — a write-skew on the last-admin
// invariant that READ COMMITTED (PostgreSQL's default) does not prevent.
//
// The SELECT ... FOR UPDATE clause is only emitted on PostgreSQL. SQLite
// (modernc, registered as "sqlite3") does not support row locking and already
// serializes writers, so the lock is skipped there. Call this inside the
// transaction — passing tx.Client() — before counting admin holders.
func lockTenantAdmins(ctx context.Context, client *ent.Client, dlct string, gid uuid.UUID) error {
	if dlct != dialect.Postgres {
		return nil
	}
	// The query runs under a system context: it is an internal invariant lock,
	// not a viewer-scoped read. A missing row (tenant deleted concurrently)
	// surfaces as NotFound and aborts the transaction, which is correct.
	_, err := client.Group.Query().
		Where(group.ID(gid)).
		ForUpdate().
		Only(authz.NewSystemContext(ctx))
	return err
}
