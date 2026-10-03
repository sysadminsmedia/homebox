package migrations_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/backupdestination"
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

// TestSqliteBackupDestinationRebuildKeepsRows applies the migrations up to the
// first backup_destinations schema, stores a row, then applies the SFTP/WebDAV
// migration, which rebuilds the table to widen the type CHECK constraint.
func TestSqliteBackupDestinationRebuildKeepsRows(t *testing.T) {
	ctx := context.Background()
	c, err := ent.Open("sqlite3", "file:migrebuild?mode=memory&cache=shared&_fk=1&_time_format=sqlite")
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	fs, err := migrations.Migrations("sqlite3")
	require.NoError(t, err)
	goose.SetBaseFS(fs)
	require.NoError(t, goose.SetDialect("sqlite3"))
	require.NoError(t, goose.UpTo(c.Sql(), "sqlite3", 20261002000000))

	gid, did := uuid.New(), uuid.New()
	_, err = c.Sql().ExecContext(ctx, `insert into groups (id, created_at, updated_at, name, currency) values (?, datetime('now'), datetime('now'), 'g', 'USD')`, gid)
	require.NoError(t, err)
	_, err = c.Sql().ExecContext(ctx, `insert into backup_destinations (id, created_at, updated_at, name, type, conn_string, keep_daily, group_id)
		values (?, datetime('now'), datetime('now'), 'old', 's3', 's3://bucket', 11, ?)`, did, gid)
	require.NoError(t, err)

	// The old CHECK constraint refuses the new type.
	_, err = c.Sql().ExecContext(ctx, `update backup_destinations set type = 'sftp' where id = ?`, did)
	require.Error(t, err)

	require.NoError(t, goose.Up(c.Sql(), "sqlite3"))

	d, err := c.BackupDestination.Get(ctx, did)
	require.NoError(t, err)
	require.Equal(t, "old", d.Name)
	require.Equal(t, "s3://bucket", d.ConnString)
	require.Equal(t, 11, d.KeepDaily)
	require.Equal(t, gid, d.GroupID)

	// The widened constraint and the new columns work.
	_, err = c.BackupDestination.UpdateOneID(did).
		SetType("sftp").SetUsername("u").SetSecret("sealed").SetHostKey("SHA256:x").Save(ctx)
	require.NoError(t, err)
	_, err = c.BackupDestination.UpdateOneID(did).SetType("webdav").Save(ctx)
	require.NoError(t, err)

	// The group cascade survived the rebuild.
	require.NoError(t, c.Group.DeleteOneID(gid).Exec(ctx))
	n, err := c.BackupDestination.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, n)
}

// TestSqliteCloudDriveRebuildKeepsCredentials checks that the second rebuild of
// backup_destinations keeps the sftp/webdav columns added by the first, and
// that the cloud-drive types are accepted afterwards.
func TestSqliteCloudDriveRebuildKeepsCredentials(t *testing.T) {
	ctx := context.Background()
	c, err := ent.Open("sqlite3", "file:migdrives?mode=memory&cache=shared&_fk=1&_time_format=sqlite")
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	fs, err := migrations.Migrations("sqlite3")
	require.NoError(t, err)
	goose.SetBaseFS(fs)
	require.NoError(t, goose.SetDialect("sqlite3"))
	require.NoError(t, goose.UpTo(c.Sql(), "sqlite3", 20261003000000))

	gid, did := uuid.New(), uuid.New()
	_, err = c.Sql().ExecContext(ctx, `insert into groups (id, created_at, updated_at, name, currency) values (?, datetime('now'), datetime('now'), 'g', 'USD')`, gid)
	require.NoError(t, err)
	_, err = c.Sql().ExecContext(ctx, `insert into backup_destinations
		(id, created_at, updated_at, name, type, conn_string, username, secret, host_key, keep_weekly, group_id)
		values (?, datetime('now'), datetime('now'), 'nas', 'sftp', 'sftp://nas:22/x', 'bob', 'sealed-blob', 'SHA256:abc', 9, ?)`, did, gid)
	require.NoError(t, err)

	// The previous constraint refuses the new types.
	_, err = c.Sql().ExecContext(ctx, `update backup_destinations set type = 'gdrive' where id = ?`, did)
	require.Error(t, err)

	require.NoError(t, goose.Up(c.Sql(), "sqlite3"))

	d, err := c.BackupDestination.Get(ctx, did)
	require.NoError(t, err)
	require.Equal(t, "sftp", string(d.Type))
	require.Equal(t, "bob", d.Username)
	require.Equal(t, "sealed-blob", d.Secret)
	require.Equal(t, "SHA256:abc", d.HostKey)
	require.Equal(t, 9, d.KeepWeekly)

	for _, typ := range []string{"gdrive", "onedrive", "dropbox"} {
		_, err = c.BackupDestination.UpdateOneID(did).SetType(backupdestination.Type(typ)).Save(ctx)
		require.NoError(t, err, typ)
	}
	_, err = c.Sql().ExecContext(ctx, `update backup_destinations set type = 'ftp' where id = ?`, did)
	require.Error(t, err, "unknown types are still rejected")

	require.NoError(t, c.Group.DeleteOneID(gid).Exec(ctx))
	n, err := c.BackupDestination.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, n)
}

// TestSqliteSMBCronRebuildKeepsRows checks the rebuild that adds the smb type,
// the cron frequency and the cron_expr column keeps every earlier column.
func TestSqliteSMBCronRebuildKeepsRows(t *testing.T) {
	ctx := context.Background()
	c, err := ent.Open("sqlite3", "file:migsmbcron?mode=memory&cache=shared&_fk=1&_time_format=sqlite")
	require.NoError(t, err)
	defer func() { _ = c.Close() }()

	fs, err := migrations.Migrations("sqlite3")
	require.NoError(t, err)
	goose.SetBaseFS(fs)
	require.NoError(t, goose.SetDialect("sqlite3"))
	require.NoError(t, goose.UpTo(c.Sql(), "sqlite3", 20261004000000))

	gid, did := uuid.New(), uuid.New()
	_, err = c.Sql().ExecContext(ctx, `insert into groups (id, created_at, updated_at, name, currency) values (?, datetime('now'), datetime('now'), 'g', 'USD')`, gid)
	require.NoError(t, err)
	_, err = c.Sql().ExecContext(ctx, `insert into backup_destinations
		(id, created_at, updated_at, name, type, conn_string, username, secret, host_key, frequency, keep_monthly, group_id)
		values (?, datetime('now'), datetime('now'), 'drive', 'gdrive', '', 'me@example.com', 'sealed', '', 'weekly', 5, ?)`, did, gid)
	require.NoError(t, err)

	// The previous constraints refuse the new type and frequency.
	_, err = c.Sql().ExecContext(ctx, `update backup_destinations set type = 'smb' where id = ?`, did)
	require.Error(t, err)
	_, err = c.Sql().ExecContext(ctx, `update backup_destinations set frequency = 'cron' where id = ?`, did)
	require.Error(t, err)

	require.NoError(t, goose.Up(c.Sql(), "sqlite3"))

	d, err := c.BackupDestination.Get(ctx, did)
	require.NoError(t, err)
	require.Equal(t, "gdrive", string(d.Type))
	require.Equal(t, "me@example.com", d.Username)
	require.Equal(t, "sealed", d.Secret)
	require.Equal(t, "weekly", string(d.Frequency))
	require.Equal(t, 5, d.KeepMonthly)
	require.Empty(t, d.CronExpr)

	_, err = c.BackupDestination.UpdateOneID(did).SetType("smb").SetFrequency("cron").SetCronExpr("0 3 * * *").Save(ctx)
	require.NoError(t, err)
	_, err = c.Sql().ExecContext(ctx, `update backup_destinations set frequency = 'yearly' where id = ?`, did)
	require.Error(t, err, "unknown frequencies are still rejected")

	require.NoError(t, c.Group.DeleteOneID(gid).Exec(ctx))
	n, err := c.BackupDestination.Query().Count(ctx)
	require.NoError(t, err)
	require.Zero(t, n)
}
