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

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/export"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

// createDestination and updateDestination are the only places these tests touch
// the create/update input type, which later stages widen.
func createDestination(gid uuid.UUID, s repo.BackupSettings) (repo.BackupDestinationOut, error) {
	return tSvc.Backups.CreateDestination(context.Background(), gid, repo.BackupInput{BackupSettings: s})
}

func updateDestination(gid, id uuid.UUID, s repo.BackupSettings) (repo.BackupDestinationOut, error) {
	return tSvc.Backups.UpdateDestination(context.Background(), gid, id, repo.BackupInput{BackupSettings: s})
}

func newLocalDestination(t *testing.T, gid uuid.UUID, mut func(*repo.BackupSettings)) repo.BackupDestinationOut {
	t.Helper()
	s := repo.BackupSettings{
		Name: "review", Type: destTypeLocal, ConnString: "review-" + fk.Str(6), Prefix: "homebox-backups",
		Enabled: true, ScheduleEnabled: true, Frequency: freqDaily, IntervalHours: 1, AtHour: 3, DayOfMonth: 1,
		SkipIfUnchanged: true, KeepDaily: 7, KeepWeekly: 4, KeepMonthly: 6,
		HealthIntervalMinutes: 15, AlertsEnabled: true, AlertFailureThreshold: 2, AlertStaleHours: 48,
	}
	if mut != nil {
		mut(&s)
	}
	d, err := createDestination(gid, s)
	require.NoError(t, err)
	return d
}

func TestAlertFlagsAreIndependent(t *testing.T) {
	ctx := context.Background()
	grp, err := tRepos.Groups.GroupCreate(ctx, "flags-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)
	d := newLocalDestination(t, grp.ID, nil)
	r := tRepos.BackupDestinations
	get := func() repo.BackupDestinationOut {
		got, err := r.Get(ctx, grp.ID, d.ID)
		require.NoError(t, err)
		return got
	}

	require.NoError(t, r.SetAlertedUnreachable(ctx, d.ID, true))
	require.NoError(t, r.SetAlertedFailure(ctx, d.ID, true))
	require.NoError(t, r.SetAlertedStale(ctx, d.ID, true))
	got := get()
	assert.True(t, got.AlertedUnreachable && got.AlertedFailure && got.AlertedStale)

	// Clearing one flag leaves the others exactly as they were, which a writer
	// holding a stale copy of the row could not guarantee when every update wrote
	// all three.
	require.NoError(t, r.SetAlertedFailure(ctx, d.ID, false))
	got = get()
	assert.True(t, got.AlertedUnreachable)
	assert.False(t, got.AlertedFailure)
	assert.True(t, got.AlertedStale)

	require.NoError(t, r.SetAlertedUnreachable(ctx, d.ID, false))
	got = get()
	assert.False(t, got.AlertedUnreachable)
	assert.True(t, got.AlertedStale)

	// A successful run clears failure and stale together, and only those.
	require.NoError(t, r.SetAlertedUnreachable(ctx, d.ID, true))
	require.NoError(t, r.MarkRunSucceeded(ctx, d.ID, time.Now()))
	got = get()
	assert.True(t, got.AlertedUnreachable, "a backup succeeding does not prove the destination was reachable")
	assert.False(t, got.AlertedFailure)
	assert.False(t, got.AlertedStale)
}

func TestStaleThreshold(t *testing.T) {
	at := func(freq string, interval, hours int) repo.BackupDestinationOut {
		return repo.BackupDestinationOut{BackupSettings: repo.BackupSettings{Frequency: freq, IntervalHours: interval, AlertStaleHours: hours}}
	}
	assert.Equal(t, 48*time.Hour, staleThreshold(at("daily", 1, 48)))
	assert.Equal(t, 14*24*time.Hour, staleThreshold(at(freqWeekly, 1, 48)), "a weekly schedule is given two weeks")
	assert.Equal(t, 62*24*time.Hour, staleThreshold(at("monthly", 1, 48)))
	assert.Equal(t, 12*time.Hour, staleThreshold(at("hourly", 6, 2)), "every 6 hours is given 12")
	assert.Equal(t, 1000*time.Hour, staleThreshold(at("daily", 1, 1000)), "a longer configured value wins")
}

func TestStaleThresholdFollowsCronSchedules(t *testing.T) {
	cron := func(expr string, hours int) repo.BackupDestinationOut {
		return repo.BackupDestinationOut{BackupSettings: repo.BackupSettings{Frequency: "cron", CronExpr: expr, AlertStaleHours: hours}}
	}
	// A daylight-saving change makes one daily gap 25 hours, so allow a little slack.
	between := func(got, lo, hi time.Duration, msg string) {
		assert.GreaterOrEqual(t, got, lo, msg)
		assert.LessOrEqual(t, got, hi, msg)
	}
	between(staleThreshold(cron("0 3 * * *", 48)), 48*time.Hour, 50*time.Hour, "daily cron")
	assert.Equal(t, 48*time.Hour, staleThreshold(cron("*/10 * * * *", 48)), "frequent cron keeps the configured hours")
	between(staleThreshold(cron("0 3 * * 0", 48)), 14*24*time.Hour, 14*24*time.Hour+4*time.Hour, "weekly cron")
	assert.GreaterOrEqual(t, staleThreshold(cron("0 6 * * 1-5", 48)), 2*3*24*time.Hour,
		"weekdays only: the weekend gap is three days and must not look like a problem")
	assert.GreaterOrEqual(t, staleThreshold(cron("0 2 1 * *", 48)), 2*28*24*time.Hour, "monthly cron")
	assert.Equal(t, 48*time.Hour, staleThreshold(cron("garbage", 48)), "an unparseable stored expression falls back to daily")
}

func TestCheckStaleAlerts(t *testing.T) {
	ctx := context.Background()
	grp, err := tRepos.Groups.GroupCreate(ctx, "stale-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)
	now := time.Now()
	ago := func(d time.Duration) *time.Time { v := now.Add(-d); return &v }
	const day = 24 * time.Hour

	run := func(mut func(*repo.BackupSettings), success, skipped *time.Time) bool {
		d := newLocalDestination(t, grp.ID, mut)
		if success != nil {
			require.NoError(t, tRepos.BackupDestinations.MarkRunSucceeded(ctx, d.ID, *success))
		}
		if skipped != nil {
			require.NoError(t, tRepos.BackupDestinations.MarkSkipped(ctx, d.ID, *skipped))
		}
		d, err := tRepos.BackupDestinations.Get(ctx, grp.ID, d.ID)
		require.NoError(t, err)
		tSvc.Backups.checkStale(ctx, d, now)
		d, err = tRepos.BackupDestinations.Get(ctx, grp.ID, d.ID)
		require.NoError(t, err)
		return d.AlertedStale
	}
	weekly := func(s *repo.BackupSettings) { s.Frequency = freqWeekly }

	assert.True(t, run(nil, ago(3*day), nil), "a daily destination with no backup for 3 days alerts")
	assert.False(t, run(nil, ago(1*day), nil), "a recent backup does not")
	assert.False(t, run(nil, ago(3*day), ago(2*time.Hour)), "a recent skip (nothing changed) counts as fresh")
	assert.True(t, run(nil, ago(3*day), ago(3*day)), "an old skip does not hide a stale destination")

	assert.False(t, run(weekly, ago(5*day), nil), "a weekly schedule is not stale between runs, even with the default 48 hours")
	assert.False(t, run(weekly, ago(13*day), nil), "one missed week is not yet a problem")
	assert.True(t, run(weekly, ago(15*day), nil), "two missed weeks is")

	assert.False(t, run(func(s *repo.BackupSettings) { s.AlertsEnabled = false }, ago(30*day), nil), "alerts off means no alert")
	assert.False(t, run(func(s *repo.BackupSettings) { s.AlertStaleHours = 0 }, ago(30*day), nil), "0 turns the check off")
}

func TestUpdateCannotRelocateADestinationHoldingBackups(t *testing.T) {
	ctx := context.Background()
	grp, err := tRepos.Groups.GroupCreate(ctx, "reloc-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)
	d := newLocalDestination(t, grp.ID, nil)

	edit := func(mut func(*repo.BackupSettings)) error {
		s := d.BackupSettings
		mut(&s)
		_, err := updateDestination(grp.ID, d.ID, s)
		return err
	}

	// With no backups yet, the location can still change.
	require.NoError(t, edit(func(s *repo.BackupSettings) { s.Prefix = "before-any-backup" }))
	d, err = tSvc.Backups.GetDestination(ctx, grp.ID, d.ID)
	require.NoError(t, err)

	exp, err := tRepos.Exports.CreateForDestination(ctx, grp.ID, d.ID, "scheduled")
	require.NoError(t, err)
	tSvc.Exports.RunExport(ctx, exp.ID, grp.ID)
	exp, err = tRepos.Exports.Get(ctx, grp.ID, exp.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", exp.Status, exp.Error)

	// With a stored backup, changing where it lives is refused: the file would
	// no longer be found for download, pruning or deletion.
	require.ErrorIs(t, edit(func(s *repo.BackupSettings) { s.Prefix = "somewhere-else" }), ErrBackupInvalid)
	require.ErrorIs(t, edit(func(s *repo.BackupSettings) { s.Type = destTypePrimary; s.ConnString = "" }), ErrBackupInvalid)
	require.ErrorIs(t, edit(func(s *repo.BackupSettings) { s.ConnString = "other-dir" }), ErrBackupInvalid)

	// Everything that does not move the files stays editable.
	require.NoError(t, edit(func(s *repo.BackupSettings) {
		s.Name = "renamed"
		s.KeepDaily = 3
		s.Frequency = freqWeekly
		s.AlertStaleHours = 100
	}))

	// The file is still where the row says, and still downloadable.
	store, key, err := tSvc.Backups.ArtifactLocation(ctx, grp.ID, exp)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(os.TempDir(), "homebox-backup-root", d.ConnString, filepath.FromSlash(exp.ArtifactPath)))
	require.NoError(t, err)
	_ = store.Close()
	assert.Contains(t, key, exp.ArtifactPath)

	// Once the history is gone, relocating is allowed again.
	require.NoError(t, tSvc.Backups.DeleteVersion(ctx, grp.ID, exp))
	require.NoError(t, edit(func(s *repo.BackupSettings) { s.Prefix = "now-free-to-move" }))
}

func TestDestinationBackupsAreOwnerOnly(t *testing.T) {
	fx := newOwnershipFixture(t)
	ctx := context.Background()

	d := newLocalDestination(t, fx.group.ID, nil)
	managed, err := tRepos.Exports.CreateForDestination(ctx, fx.group.ID, d.ID, "scheduled")
	require.NoError(t, err)
	plain, err := tRepos.Exports.Create(ctx, fx.group.ID)
	require.NoError(t, err)

	all, err := tRepos.Exports.ListByGroup(ctx, fx.group.ID)
	require.NoError(t, err)
	require.Len(t, all, 2)

	// The owner sees everything.
	rows, err := tSvc.Backups.VisibleExports(fx.ownerCtx, all)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
	require.NoError(t, tSvc.Backups.AuthorizeExportAccess(fx.ownerCtx, managed))
	require.NoError(t, tSvc.Backups.AuthorizeExportAccess(fx.ownerCtx, plain))

	// A member keeps ordinary exports but never sees, downloads or deletes the
	// backups the owner-only destination routes manage.
	rows, err = tSvc.Backups.VisibleExports(fx.memberCtx, all)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, plain.ID, rows[0].ID)
	assertForbidden(t, tSvc.Backups.AuthorizeExportAccess(fx.memberCtx, managed))
	require.NoError(t, tSvc.Backups.AuthorizeExportAccess(fx.memberCtx, plain))
}

func TestInterruptedRunsDoNotBlockTheDestination(t *testing.T) {
	ctx := context.Background()
	grp, err := tRepos.Groups.GroupCreate(ctx, "stuck-"+fk.Str(4), uuid.Nil)
	require.NoError(t, err)
	d := newLocalDestination(t, grp.ID, nil)

	// A run that was "running" when the server died, and never touched again.
	stuck, err := tRepos.Exports.CreateForDestination(ctx, grp.ID, d.ID, "scheduled")
	require.NoError(t, err)
	require.NoError(t, tRepos.Exports.SetRunning(ctx, grp.ID, stuck.ID))
	old := time.Now().Add(-2 * activeRunTimeout)
	_, err = tClient.Export.UpdateOneID(stuck.ID).SetUpdatedAt(old).Save(ctx)
	require.NoError(t, err)

	// A fresh run, by contrast, still blocks a second one.
	fresh := newLocalDestination(t, grp.ID, nil)
	_, err = tRepos.Exports.CreateForDestination(ctx, grp.ID, fresh.ID, "scheduled")
	require.NoError(t, err)
	_, err = tSvc.Backups.RunNow(ctx, grp.ID, fresh.ID)
	require.ErrorIs(t, err, ErrBackupInvalid, "a live run still blocks another")

	// The interrupted one is failed and a new backup can start.
	out, err := tSvc.Backups.RunNow(ctx, grp.ID, d.ID)
	require.NoError(t, err)
	assert.NotEqual(t, stuck.ID, out.ID)
	got, err := tRepos.Exports.Get(ctx, grp.ID, stuck.ID)
	require.NoError(t, err)
	assert.Equal(t, string(export.StatusFailed), got.Status)
	assert.Contains(t, got.Error, "interrupted")
}
