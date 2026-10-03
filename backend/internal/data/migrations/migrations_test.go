package migrations_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/export"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
	_ "github.com/sysadminsmedia/homebox/backend/internal/data/migrations/sqlite3"
	_ "github.com/sysadminsmedia/homebox/backend/pkgs/cgofreesqlite"
)

func TestSqliteMigrationsApplyAndMatchEnt(t *testing.T) {
	ctx := context.Background()
	c, err := ent.Open("sqlite3", "file:mig?mode=memory&cache=shared&_fk=1&_time_format=sqlite")
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	fs, err := migrations.Migrations("sqlite3")
	require.NoError(t, err)
	goose.SetBaseFS(fs)
	require.NoError(t, goose.SetDialect("sqlite3"))
	require.NoError(t, goose.Up(c.Sql(), "sqlite3"))

	g, err := c.Group.Create().SetName("g").Save(ctx)
	require.NoError(t, err)
	d, err := c.BackupDestination.Create().SetGroupID(g.ID).SetName("d").Save(ctx)
	require.NoError(t, err)
	require.Equal(t, "primary", string(d.Type))
	require.Equal(t, 7, d.KeepDaily)

	did := uuid.New()
	e, err := c.Export.Create().SetGroupID(g.ID).SetOrigin(export.OriginScheduled).SetDestinationID(did).Save(ctx)
	require.NoError(t, err)
	require.Equal(t, did, *e.DestinationID)
	_, err = c.Export.Create().SetGroupID(g.ID).Save(ctx)
	require.NoError(t, err)

	// Round-trip every nullable/updated column on the destination.
	_, err = c.BackupDestination.UpdateOneID(d.ID).SetHealthStatus("unreachable").SetHealthFailures(2).
		SetLastFingerprint("x").SetAlertedUnreachable(true).Save(ctx)
	require.NoError(t, err)
	got, err := c.BackupDestination.Get(ctx, d.ID)
	require.NoError(t, err)
	require.True(t, got.AlertedUnreachable)

	// cascade
	require.NoError(t, c.Group.DeleteOneID(g.ID).Exec(ctx))
	n, err := c.BackupDestination.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, n)
}
