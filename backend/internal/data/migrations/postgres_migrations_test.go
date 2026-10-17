package migrations_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/backupdestination"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/export"
	"github.com/sysadminsmedia/homebox/backend/internal/data/migrations"
	_ "github.com/sysadminsmedia/homebox/backend/internal/data/migrations/postgres"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

func TestPostgresMigrationsAndBackupQueries(t *testing.T) {
	url := os.Getenv("HBOX_TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("set HBOX_TEST_POSTGRES_URL to run against a disposable postgres database")
	}
	ctx := context.Background()
	db, err := sql.Open("pgx", url)
	require.NoError(t, err)
	c := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	fs, err := migrations.Migrations("postgres")
	require.NoError(t, err)
	goose.SetBaseFS(fs)
	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(db, "postgres"))

	g, err := c.Group.Create().SetName("g").Save(ctx)
	require.NoError(t, err)
	d, err := c.BackupDestination.Create().SetGroupID(g.ID).SetName("d").Save(ctx)
	require.NoError(t, err)
	require.Equal(t, 7, d.KeepDaily)
	did := uuid.New()
	e, err := c.Export.Create().SetGroupID(g.ID).SetOrigin(export.OriginScheduled).SetDestinationID(did).Save(ctx)
	require.NoError(t, err)
	require.Equal(t, did, *e.DestinationID)
	_, err = c.BackupDestination.UpdateOneID(d.ID).SetHealthStatus("unreachable").SetHealthFailures(2).SetAlertedUnreachable(true).SetLastFingerprint("x").Save(ctx)
	require.NoError(t, err)

	// SFTP and WebDAV destinations: widened type constraint and new columns.
	_, err = c.BackupDestination.UpdateOneID(d.ID).
		SetType("sftp").SetUsername("u").SetSecret("sealed").SetHostKey("SHA256:x").Save(ctx)
	require.NoError(t, err)
	_, err = c.BackupDestination.UpdateOneID(d.ID).SetType("webdav").Save(ctx)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `update backup_destinations set type = 'ftp' where id = $1`, d.ID)
	require.Error(t, err, "the check constraint still rejects unknown types")
	for _, typ := range []string{"gdrive", "onedrive", "dropbox"} {
		_, err = c.BackupDestination.UpdateOneID(d.ID).SetType(backupdestination.Type(typ)).Save(ctx)
		require.NoError(t, err, typ)
	}
	_, err = c.BackupDestination.UpdateOneID(d.ID).SetType("smb").SetFrequency("cron").SetCronExpr("0 3 * * *").Save(ctx)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `update backup_destinations set frequency = 'yearly' where id = $1`, d.ID)
	require.Error(t, err, "the frequency constraint still rejects unknown values")
	_, err = c.BackupDestination.UpdateOneID(d.ID).SetType("primary").SetFrequency("daily").Save(ctx)
	require.NoError(t, err)

	// The raw-SQL change fingerprint must work on postgres placeholders too.
	bus := eventbus.New()
	st := config.Storage{PrefixPath: "/", ConnString: "file://" + os.TempDir()}
	repos := repo.New(c, bus, st, "mem://{{ .Topic }}", config.Thumbnail{}, nil)
	svc := services.New(repos,
		services.WithExportPlumbing(bus, c, st, "mem://{{ .Topic }}", "postgres"),
		services.WithBackupConfig(config.BackupConf{Enabled: true}),
	)
	fp1, err := svc.Backups.Fingerprint(ctx, g.ID)
	require.NoError(t, err)
	fp2, err := svc.Backups.Fingerprint(ctx, g.ID)
	require.NoError(t, err)
	require.Equal(t, fp1, fp2)
	_, err = repos.Tags.Create(ctx, g.ID, repo.TagCreate{Name: "t"})
	require.NoError(t, err)
	fp3, err := svc.Backups.Fingerprint(ctx, g.ID)
	require.NoError(t, err)
	require.NotEqual(t, fp1, fp3)
	res, err := svc.Backups.TestDestination(ctx, g.ID, d.ID)
	require.NoError(t, err)
	require.True(t, res.OK, res.Message)

	require.NoError(t, c.Group.DeleteOneID(g.ID).Exec(ctx))
}
