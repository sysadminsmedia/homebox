package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
)

const (
	freqDaily  = "daily"
	freqWeekly = "weekly"
)

func at(y int, m time.Month, d, h, mi int) time.Time {
	return time.Date(y, m, d, h, mi, 0, 0, time.Local)
}

func TestNextRun(t *testing.T) {
	// 2026-10-02 is a Friday (weekday 5).
	now := at(2026, 10, 2, 10, 30)

	cases := []struct {
		name string
		c    repo.BackupSettings
		want time.Time
	}{
		{"hourly later this hour", repo.BackupSettings{Frequency: "hourly", IntervalHours: 1, AtMinute: 45}, at(2026, 10, 2, 10, 45)},
		{"hourly next hour", repo.BackupSettings{Frequency: "hourly", IntervalHours: 1, AtMinute: 15}, at(2026, 10, 2, 11, 15)},
		{"every 6 hours", repo.BackupSettings{Frequency: "hourly", IntervalHours: 6, AtMinute: 15}, at(2026, 10, 2, 16, 15)},
		{"daily later today", repo.BackupSettings{Frequency: freqDaily, AtHour: 23, AtMinute: 0}, at(2026, 10, 2, 23, 0)},
		{"daily tomorrow", repo.BackupSettings{Frequency: freqDaily, AtHour: 3, AtMinute: 0}, at(2026, 10, 3, 3, 0)},
		{"daily exactly now moves on", repo.BackupSettings{Frequency: freqDaily, AtHour: 10, AtMinute: 30}, at(2026, 10, 3, 10, 30)},
		{"weekly same weekday later", repo.BackupSettings{Frequency: freqWeekly, Weekday: 5, AtHour: 20}, at(2026, 10, 2, 20, 0)},
		{"weekly same weekday passed", repo.BackupSettings{Frequency: freqWeekly, Weekday: 5, AtHour: 3}, at(2026, 10, 9, 3, 0)},
		{"weekly sunday", repo.BackupSettings{Frequency: freqWeekly, Weekday: 0, AtHour: 4}, at(2026, 10, 4, 4, 0)},
		{"monthly later this month", repo.BackupSettings{Frequency: "monthly", DayOfMonth: 15, AtHour: 3}, at(2026, 10, 15, 3, 0)},
		{"monthly next month", repo.BackupSettings{Frequency: "monthly", DayOfMonth: 1, AtHour: 3}, at(2026, 11, 1, 3, 0)},
		{"monthly rolls the year", repo.BackupSettings{Frequency: "monthly", DayOfMonth: 1, AtHour: 3}, at(2027, 1, 1, 3, 0)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			from := now
			if tc.name == "monthly rolls the year" {
				from = at(2026, 12, 20, 9, 0)
			}
			got := NextRun(tc.c, from)
			assert.True(t, got.Equal(tc.want), "got %s want %s", got, tc.want)
			assert.True(t, got.After(from))
		})
	}
}

func row(daysAgo int, status string) repo.ExportOut {
	return repo.ExportOut{
		ID:        uuid.New(),
		Status:    status,
		CreatedAt: time.Now().Add(-time.Duration(daysAgo)*24*time.Hour - time.Minute),
	}
}

func TestSelectPrunable(t *testing.T) {
	t.Run("all zero keeps everything", func(t *testing.T) {
		rows := []repo.ExportOut{row(0, "completed"), row(1, "completed"), row(400, "completed")}
		assert.Empty(t, SelectPrunable(rows, 0, 0, 0))
	})

	t.Run("keeps the newest per day for the last N days", func(t *testing.T) {
		var rows []repo.ExportOut
		for i := 0; i < 10; i++ {
			rows = append(rows, row(i, "completed"))
		}
		pruned := SelectPrunable(rows, 3, 0, 0)
		assert.Len(t, pruned, 7)
		for _, p := range pruned {
			for _, k := range rows[:3] {
				assert.NotEqual(t, k.ID, p.ID)
			}
		}
	})

	t.Run("two backups on the same day keep only the newest", func(t *testing.T) {
		newer := row(0, "completed")
		older := newer
		older.ID = uuid.New()
		older.CreatedAt = newer.CreatedAt.Add(-time.Second)
		pruned := SelectPrunable([]repo.ExportOut{older, newer}, 7, 0, 0)
		require.Len(t, pruned, 1)
		assert.Equal(t, older.ID, pruned[0].ID)
	})

	t.Run("newest is always kept and failed rows are ignored", func(t *testing.T) {
		rows := []repo.ExportOut{row(0, "failed"), row(30, "completed"), row(31, "completed")}
		pruned := SelectPrunable(rows, 1, 0, 0)
		require.Len(t, pruned, 1)
		assert.Equal(t, rows[2].ID, pruned[0].ID)
	})

	t.Run("weekly and monthly policies widen retention", func(t *testing.T) {
		var rows []repo.ExportOut
		for i := 0; i < 120; i += 3 {
			rows = append(rows, row(i, "completed"))
		}
		daily := SelectPrunable(rows, 3, 0, 0)
		wide := SelectPrunable(rows, 3, 4, 3)
		assert.Less(t, len(wide), len(daily))
	})
}

func TestNormalizeSettings(t *testing.T) {
	svc := &BackupService{cfg: config.BackupConf{Enabled: true, LocalRoot: t.TempDir(), AllowCustomEndpoints: true}}
	base := repo.BackupSettings{Name: "x", Type: "s3", ConnString: "s3://bucket?region=us-east-1", Frequency: freqDaily}

	ok := func(mut func(*repo.BackupSettings)) repo.BackupSettings {
		s := base
		mut(&s)
		out, err := svc.NormalizeSettings(s)
		require.NoError(t, err)
		return out
	}
	bad := func(svc *BackupService, mut func(*repo.BackupSettings)) {
		s := base
		mut(&s)
		_, err := svc.NormalizeSettings(s)
		require.ErrorIs(t, err, ErrBackupInvalid)
	}

	t.Run("accepts a plain bucket and defaults the prefix", func(t *testing.T) {
		out := ok(func(s *repo.BackupSettings) {})
		assert.Equal(t, "homebox-backups", out.Prefix)
	})
	t.Run("primary ignores the connection string", func(t *testing.T) {
		out := ok(func(s *repo.BackupSettings) { s.Type = destTypePrimary; s.ConnString = "junk" })
		assert.Empty(t, out.ConnString)
	})
	t.Run("rejects wrong scheme, userinfo and missing bucket", func(t *testing.T) {
		bad(svc, func(s *repo.BackupSettings) { s.ConnString = "gcs://bucket" })
		bad(svc, func(s *repo.BackupSettings) { s.ConnString = "s3://key:secret@bucket" })
		bad(svc, func(s *repo.BackupSettings) { s.ConnString = "s3://" })
		bad(svc, func(s *repo.BackupSettings) { s.ConnString = "" })
	})
	t.Run("rejects traversal in the prefix", func(t *testing.T) {
		bad(svc, func(s *repo.BackupSettings) { s.Prefix = "../etc" })
		bad(svc, func(s *repo.BackupSettings) { s.Prefix = `a\b` })
	})
	t.Run("local destinations stay inside the root", func(t *testing.T) {
		out := ok(func(s *repo.BackupSettings) { s.Type = destTypeLocal; s.ConnString = "nas/daily" })
		assert.Equal(t, "nas/daily", out.ConnString)
		for _, p := range []string{"", "/etc", "../x", "a/../../x", ".", "c:evil"} {
			bad(svc, func(s *repo.BackupSettings) { s.Type = destTypeLocal; s.ConnString = p })
		}
	})
	t.Run("local destinations need a configured root", func(t *testing.T) {
		noRoot := &BackupService{cfg: config.BackupConf{Enabled: true}}
		bad(noRoot, func(s *repo.BackupSettings) { s.Type = destTypeLocal; s.ConnString = "x" })
	})
	t.Run("custom endpoints can be disabled", func(t *testing.T) {
		locked := &BackupService{cfg: config.BackupConf{Enabled: true}}
		bad(locked, func(s *repo.BackupSettings) { s.ConnString = "s3://b?endpoint=http://10.0.0.5:9000" })
		_, err := locked.NormalizeSettings(base)
		require.NoError(t, err)
	})
}

// TestScheduledBackupLifecycle drives the real pipeline against the
// in-memory database: a destination-bound export writes into the local root,
// an unchanged collection is skipped on the next due tick, a change triggers a
// new run, and retention removes older versions.
func TestScheduledBackupLifecycle(t *testing.T) {
	ctx := context.Background()

	grp, err := tRepos.Groups.GroupCreate(ctx, "backup-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	settings := repo.BackupSettings{
		Name: "nas", Type: destTypeLocal, ConnString: "lifecycle-" + fk.Str(4), Prefix: "homebox-backups",
		Enabled: true, ScheduleEnabled: true, Frequency: freqDaily, IntervalHours: 1, AtHour: 3, DayOfMonth: 1,
		SkipIfUnchanged: true, KeepDaily: 1, KeepWeekly: 0, KeepMonthly: 0,
		HealthIntervalMinutes: 15, AlertsEnabled: false, AlertFailureThreshold: 2, AlertStaleHours: 0,
	}
	dest, err := tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: settings})
	require.NoError(t, err)
	require.NotNil(t, dest.NextRunAt, "enabled schedule gets a next run")

	runOnce := func() repo.ExportOut {
		exp, err := tRepos.Exports.CreateForDestination(ctx, grp.ID, dest.ID, "scheduled")
		require.NoError(t, err)
		tSvc.Exports.RunExport(ctx, exp.ID, grp.ID)
		got, err := tRepos.Exports.Get(ctx, grp.ID, exp.ID)
		require.NoError(t, err)
		return got
	}

	first := runOnce()
	require.Equal(t, "completed", first.Status, first.Error)
	require.NotNil(t, first.DestinationID)
	require.Equal(t, "scheduled", first.Origin)
	assert.Contains(t, first.ArtifactPath, "homebox-backups/"+grp.ID.String()+"/backups/")

	root := os.TempDir() + "/homebox-backup-root"
	_, err = os.Stat(filepath.Join(root, settings.ConnString, filepath.FromSlash(first.ArtifactPath)))
	require.NoError(t, err, "artifact is written under the local root")

	dest, err = tSvc.Backups.GetDestination(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	require.NotNil(t, dest.LastSuccessAt)
	require.NotEmpty(t, dest.LastFingerprint)

	// Due, but nothing changed: skipped, no new export row.
	require.NoError(t, tRepos.BackupDestinations.SetNextRun(ctx, dest.ID, time.Now().Add(-time.Minute), nil))
	tSvc.Backups.SchedulerTick(ctx)
	versions, err := tSvc.Backups.ListVersions(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	dest, _ = tSvc.Backups.GetDestination(ctx, grp.ID, dest.ID)
	require.NotNil(t, dest.LastSkippedAt)
	require.NotNil(t, dest.NextRunAt)
	assert.True(t, dest.NextRunAt.After(time.Now()))

	// Change data: the next due tick enqueues a pending scheduled export.
	_, err = tRepos.Tags.Create(ctx, grp.ID, repo.TagCreate{Name: "changed-" + fk.Str(4)})
	require.NoError(t, err)
	require.NoError(t, tRepos.BackupDestinations.SetNextRun(ctx, dest.ID, time.Now().Add(-time.Minute), nil))
	tSvc.Backups.SchedulerTick(ctx)
	versions, err = tSvc.Backups.ListVersions(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	require.Len(t, versions, 2)
	assert.Equal(t, "pending", versions[0].Status)

	// The scheduler never stacks a run on one that is still active.
	require.NoError(t, tRepos.BackupDestinations.SetNextRun(ctx, dest.ID, time.Now().Add(-time.Minute), nil))
	tSvc.Backups.SchedulerTick(ctx)
	versions, _ = tSvc.Backups.ListVersions(ctx, grp.ID, dest.ID)
	assert.Len(t, versions, 2)

	// Run the pending one; with keep_daily=1 both land on the same day, so
	// retention leaves only the newest completed version.
	tSvc.Exports.RunExport(ctx, versions[0].ID, grp.ID)
	versions, _ = tSvc.Backups.ListVersions(ctx, grp.ID, dest.ID)
	require.Len(t, versions, 1)
	assert.Equal(t, "completed", versions[0].Status)
	_, err = os.Stat(filepath.Join(root, settings.ConnString, filepath.FromSlash(first.ArtifactPath)))
	assert.True(t, os.IsNotExist(err), "pruned artifact is removed from the destination")

	// The 7-day sweep must not touch destination-bound backups.
	old, err := tRepos.Exports.ListOlderThan(ctx, time.Now().Add(time.Hour))
	require.NoError(t, err)
	for _, o := range old {
		assert.Nil(t, o.DestinationID)
	}
}

func TestBackupHealthAndAlertDedup(t *testing.T) {
	ctx := context.Background()
	grp, err := tRepos.Groups.GroupCreate(ctx, "health-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	// A local destination whose directory is a file cannot be created, so the
	// probe fails deterministically.
	sub := "blocked-" + fk.Str(4)
	root := os.TempDir() + "/homebox-backup-root"
	require.NoError(t, os.MkdirAll(root, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, sub), []byte("x"), 0o600))

	dest, err := tSvc.Backups.CreateDestination(ctx, grp.ID, repo.BackupInput{BackupSettings: repo.BackupSettings{
		Name: "blocked", Type: destTypeLocal, ConnString: sub, Prefix: "p", Enabled: true, Frequency: freqDaily,
		IntervalHours: 1, DayOfMonth: 1, HealthIntervalMinutes: 15, AlertsEnabled: true, AlertFailureThreshold: 2,
	}})
	require.NoError(t, err)

	res, err := tSvc.Backups.TestDestination(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	assert.False(t, res.OK)
	assert.NotContains(t, res.Message, os.TempDir(), "host paths must not reach the user")
	assert.Contains(t, res.Message, "<backup root>")
	dest, _ = tSvc.Backups.GetDestination(ctx, grp.ID, dest.ID)
	assert.Equal(t, "unreachable", dest.HealthStatus)
	assert.Equal(t, 1, dest.HealthFailures)
	assert.False(t, dest.AlertedUnreachable, "below the failure threshold")

	_, _ = tSvc.Backups.TestDestination(ctx, grp.ID, dest.ID)
	dest, _ = tSvc.Backups.GetDestination(ctx, grp.ID, dest.ID)
	assert.Equal(t, 2, dest.HealthFailures)
	assert.True(t, dest.AlertedUnreachable, "alerts once at the threshold")

	// Fix the destination: next probe recovers and clears the flag.
	require.NoError(t, os.Remove(filepath.Join(root, sub)))
	res, err = tSvc.Backups.TestDestination(ctx, grp.ID, dest.ID)
	require.NoError(t, err)
	assert.True(t, res.OK, res.Message)
	dest, _ = tSvc.Backups.GetDestination(ctx, grp.ID, dest.ID)
	assert.Equal(t, "healthy", dest.HealthStatus)
	assert.Equal(t, 0, dest.HealthFailures)
	assert.False(t, dest.AlertedUnreachable)
}

func TestBackupDestinationGroupIsolation(t *testing.T) {
	ctx := context.Background()
	a, err := tRepos.Groups.GroupCreate(ctx, "iso-a-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)
	b, err := tRepos.Groups.GroupCreate(ctx, "iso-b-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)

	dest, err := tSvc.Backups.CreateDestination(ctx, a.ID, repo.BackupInput{BackupSettings: repo.BackupSettings{
		Name: "primary", Type: destTypePrimary, Enabled: true, Frequency: freqDaily, IntervalHours: 1, DayOfMonth: 1,
		HealthIntervalMinutes: 15, AlertFailureThreshold: 2,
	}})
	require.NoError(t, err)

	_, err = tSvc.Backups.GetDestination(ctx, b.ID, dest.ID)
	require.Error(t, err)
	_, err = tSvc.Backups.RunNow(ctx, b.ID, dest.ID)
	require.Error(t, err)
	_, err = tSvc.Backups.UpdateDestination(ctx, b.ID, dest.ID, repo.BackupInput{BackupSettings: repo.BackupSettings{Name: "x", Type: destTypePrimary, Frequency: freqDaily}})
	require.Error(t, err)

	list, err := tSvc.Backups.ListDestinations(ctx, b.ID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestCronSchedules(t *testing.T) {
	now := at(2026, 10, 2, 10, 30) // a Friday

	t.Run("NextRun follows the expression", func(t *testing.T) {
		cases := []struct {
			expr string
			want time.Time
		}{
			{"0 3 * * *", at(2026, 10, 3, 3, 0)},
			{"*/15 * * * *", at(2026, 10, 2, 10, 45)},
			{"30 10 * * *", at(2026, 10, 3, 10, 30)}, // not now: strictly after
			{"0 4 * * 0", at(2026, 10, 4, 4, 0)},     // Sunday
			{"0 2 1 * *", at(2026, 11, 1, 2, 0)},
			{"@daily", at(2026, 10, 3, 0, 0)},
			{"@hourly", at(2026, 10, 2, 11, 0)},
			{"0 6 * * 1-5", at(2026, 10, 5, 6, 0)}, // weekdays: next is Monday
		}
		for _, tc := range cases {
			got := NextRun(repo.BackupSettings{Frequency: "cron", CronExpr: tc.expr}, now)
			assert.True(t, got.Equal(tc.want), "%s: got %s want %s", tc.expr, got, tc.want)
		}
	})

	t.Run("a time zone prefix is honoured", func(t *testing.T) {
		loc, err := time.LoadLocation("Asia/Tokyo")
		require.NoError(t, err)
		got := NextRun(repo.BackupSettings{Frequency: "cron", CronExpr: "CRON_TZ=Asia/Tokyo 0 3 * * *"}, now)
		h, m, _ := got.In(loc).Clock()
		assert.Equal(t, 3, h)
		assert.Equal(t, 0, m)
	})

	t.Run("validation", func(t *testing.T) {
		svc := &BackupService{cfg: config.BackupConf{Enabled: true}}
		base := repo.BackupSettings{Name: "x", Type: destTypePrimary, Frequency: "cron"}
		ok := func(expr string) {
			in := base
			in.CronExpr = expr
			out, err := svc.NormalizeSettings(in)
			require.NoError(t, err, expr)
			assert.Equal(t, expr, out.CronExpr)
		}
		bad := func(expr string) {
			in := base
			in.CronExpr = expr
			_, err := svc.NormalizeSettings(in)
			require.ErrorIs(t, err, ErrBackupInvalid, expr)
		}
		ok("0 3 * * *")
		ok("*/5 * * * *")
		ok("@weekly")
		ok("CRON_TZ=Europe/Paris 30 2 * * *")

		bad("")
		bad("not cron")
		bad("0 3 * *")     // too few fields
		bad("* * * * *")   // every minute
		bad("*/2 * * * *") // every 2 minutes
		bad("0,1 * * * *") // a 1-minute gap hidden in a pair
		bad("@every 1s")   // sub-minute descriptor
		bad("@every 2m")   // below the minimum gap
		bad("0 0 31 2 *")  // never matches
		bad("CRON_TZ=Nowhere/X 0 3 * * *")

		// The expression is dropped for other frequencies.
		in := base
		in.Frequency, in.CronExpr = freqDaily, "garbage"
		out, err := svc.NormalizeSettings(in)
		require.NoError(t, err)
		assert.Empty(t, out.CronExpr)
	})
}
