package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"

	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
)

// ErrBackupInvalid wraps every user-correctable problem with a destination's
// settings so handlers can answer 422 instead of 500.
var ErrBackupInvalid = errors.New("invalid backup destination")

// ErrBackupDisabled is returned when scheduled backups are switched off by
// configuration.
var ErrBackupDisabled = errors.New("scheduled backups are disabled")

const (
	destTypePrimary = "primary"
	destTypeLocal   = "local"

	originManual    = "manual"
	originScheduled = "scheduled"

	healthHealthy     = "healthy"
	healthUnreachable = "unreachable"

	// failedRowRetention bounds how long failed backup rows (which hold no
	// artifact) stay in the history.
	failedRowRetention = 14 * 24 * time.Hour
	checkTimeout       = 30 * time.Second

	maxDestinationsPerGroup = 20
)

// BackupService owns scheduled backups: destination management, the
// scheduler, retention pruning, health checks and alerting. The actual zip
// building and upload is still ExportService's job; scheduled runs are just
// export rows bound to a destination.
type BackupService struct {
	repos          *repo.AllRepos
	db             *ent.Client
	exports        *ExportService
	cfg            config.BackupConf
	notifierConfig *config.NotifierConf
	dialect        string
	bus            *eventbus.EventBus
}

// Enabled reports whether scheduled backups are switched on.
func (s *BackupService) Enabled() bool { return s.cfg.Enabled }

// BackupOptions describes what the server allows, for the UI.
type BackupOptions struct {
	Enabled              bool
	LocalEnabled         bool
	AllowCustomEndpoints bool
}

// Options returns the server-side backup switches.
func (s *BackupService) Options() BackupOptions {
	return BackupOptions{
		Enabled:              s.cfg.Enabled,
		LocalEnabled:         s.cfg.LocalRoot != "",
		AllowCustomEndpoints: s.cfg.AllowCustomEndpoints,
	}
}

// TestResult is the outcome of a connection test or health probe.
type TestResult struct {
	OK        bool   `json:"ok"`
	LatencyMs int64  `json:"latencyMs"`
	Message   string `json:"message"`
}

func invalid(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrBackupInvalid, fmt.Sprintf(format, a...))
}

// NormalizeSettings validates in and fills defaults. It returns the cleaned
// settings; the input is not modified.
func (s *BackupService) NormalizeSettings(in repo.BackupSettings) (repo.BackupSettings, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, invalid("name is required")
	}

	prefix, err := cleanPrefix(in.Prefix)
	if err != nil {
		return in, err
	}
	in.Prefix = prefix

	in.ConnString = strings.TrimSpace(in.ConnString)
	switch in.Type {
	case destTypePrimary:
		in.ConnString = ""
	case destTypeLocal:
		if s.cfg.LocalRoot == "" {
			return in, invalid("local destinations are disabled; set HBOX_BACKUP_LOCAL_ROOT")
		}
		if _, err := s.localDir(in.ConnString); err != nil {
			return in, err
		}
	case "s3", "gcs", "azblob":
		if err := s.validateCloudURL(in.Type, in.ConnString); err != nil {
			return in, err
		}
	default:
		return in, invalid("unknown destination type %q", in.Type)
	}
	return in, nil
}

// cleanPrefix normalizes the key prefix: no leading/trailing slashes, no
// traversal segments, no backslashes.
func cleanPrefix(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "homebox-backups", nil
	}
	if strings.ContainsAny(p, "\\\x00") {
		return "", invalid("prefix contains invalid characters")
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return "", invalid("prefix must not contain '..'")
		}
	}
	p = strings.Trim(path.Clean("/"+p), "/")
	if p == "" {
		return "homebox-backups", nil
	}
	return p, nil
}

// localDir resolves a local destination's sub-directory under the configured
// root, refusing anything that would escape it.
func (s *BackupService) localDir(sub string) (string, error) {
	sub = strings.TrimSpace(sub)
	if sub == "" {
		return "", invalid("a directory name is required for local destinations")
	}
	if filepath.IsAbs(sub) || strings.ContainsAny(sub, "\x00:") {
		return "", invalid("local directory must be a relative path")
	}
	root, err := filepath.Abs(s.cfg.LocalRoot)
	if err != nil {
		return "", invalid("invalid local backup root")
	}
	dir := filepath.Join(root, filepath.FromSlash(sub))
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", invalid("local directory must stay inside the backup root")
	}
	return dir, nil
}

func (s *BackupService) validateCloudURL(typ, raw string) error {
	if raw == "" {
		return invalid("a connection string is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return invalid("connection string is not a valid URL")
	}
	if u.Scheme != typ {
		return invalid("connection string must start with %s://", typ)
	}
	if u.User != nil {
		return invalid("credentials must not be embedded in the connection string; use the provider's environment variables")
	}
	if u.Host == "" {
		return invalid("connection string must name a bucket or container")
	}
	if !s.cfg.AllowCustomEndpoints {
		q := u.Query()
		for _, k := range []string{"endpoint", "domain", "protocol", "hostname_immutable", "disable_https", "use_path_style"} {
			if q.Has(k) {
				return invalid("custom endpoints are disabled on this server (%s)", k)
			}
		}
	}
	return nil
}

// redact replaces the server's absolute local backup root in msg, so error
// text shown to group owners never reveals the host's directory layout.
func (s *BackupService) redact(msg string) string {
	if s.cfg.LocalRoot == "" {
		return msg
	}
	root := s.cfg.LocalRoot
	if abs, err := filepath.Abs(root); err == nil {
		msg = strings.ReplaceAll(msg, abs, "<backup root>")
	}
	return strings.ReplaceAll(msg, root, "<backup root>")
}

// openBucket opens the destination's bucket and returns it with a function
// mapping artifact paths to bucket keys.
func (s *BackupService) openBucket(ctx context.Context, d repo.BackupDestinationOut) (*blob.Bucket, func(string) string, error) {
	switch d.Type {
	case destTypePrimary:
		b, err := blob.OpenBucket(ctx, s.repos.Attachments.GetConnString())
		if err != nil {
			return nil, nil, err
		}
		return b, s.repos.Attachments.GetFullPath, nil
	case destTypeLocal:
		dir, err := s.localDir(d.ConnString)
		if err != nil {
			return nil, nil, err
		}
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, nil, fmt.Errorf("create backup directory: %s", s.redact(err.Error()))
		}
		slash := filepath.ToSlash(dir)
		if !strings.HasPrefix(slash, "/") {
			slash = "/" + slash
		}
		u := url.URL{Scheme: "file", Path: slash, RawQuery: "no_tmp_dir=true"}
		b, err := blob.OpenBucket(ctx, u.String())
		if err != nil {
			return nil, nil, errors.New(s.redact(err.Error()))
		}
		return b, func(p string) string { return p }, nil
	default:
		if err := s.validateCloudURL(d.Type, d.ConnString); err != nil {
			return nil, nil, err
		}
		b, err := blob.OpenBucket(ctx, d.ConnString)
		if err != nil {
			return nil, nil, err
		}
		return b, func(p string) string { return p }, nil
	}
}

// artifactPath builds the artifact path recorded on the export row. Primary
// storage keeps the historical "{gid}/exports/{id}.zip" layout so existing
// download and sweep logic applies unchanged; other destinations get a
// browsable "{prefix}/{gid}/backups/{timestamp}-{id}.zip".
func artifactPath(d repo.BackupDestinationOut, gid, exportID uuid.UUID, at time.Time) string {
	if d.Type == destTypePrimary {
		return fmt.Sprintf("%s/exports/%s.zip", gid, exportID)
	}
	return fmt.Sprintf("%s/%s/backups/%s-%s.zip", d.Prefix, gid, at.UTC().Format("20060102-150405"), exportID.String()[:8])
}

// artifactPrefix is the only key prefix a destination's artifacts may live
// under; handlers use it as a defence-in-depth check against tampered rows.
func artifactPrefix(d repo.BackupDestinationOut, gid uuid.UUID) string {
	if d.Type == destTypePrimary {
		return gid.String() + "/exports/"
	}
	return fmt.Sprintf("%s/%s/backups/", d.Prefix, gid)
}

// ArtifactLocation resolves where an export's artifact lives. The caller owns
// the returned bucket and must Close it.
func (s *BackupService) ArtifactLocation(ctx context.Context, gid uuid.UUID, exp repo.ExportOut) (*blob.Bucket, string, error) {
	if exp.ArtifactPath == "" {
		return nil, "", errors.New("export has no artifact")
	}
	clean := path.Clean(exp.ArtifactPath)
	var dest repo.BackupDestinationOut
	if exp.DestinationID != nil {
		d, err := s.repos.BackupDestinations.Get(ctx, gid, *exp.DestinationID)
		if err != nil {
			return nil, "", err
		}
		dest = d
	} else {
		dest = repo.BackupDestinationOut{BackupSettings: repo.BackupSettings{Type: destTypePrimary}}
	}
	if !strings.HasPrefix(clean, artifactPrefix(dest, gid)) {
		return nil, "", errors.New("artifact outside the expected prefix")
	}
	b, key, err := s.openBucket(ctx, dest)
	if err != nil {
		return nil, "", err
	}
	return b, key(clean), nil
}

// Check probes a destination by writing and deleting a small object.
func (s *BackupService) Check(ctx context.Context, d repo.BackupDestinationOut, gid uuid.UUID) TestResult {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	start := time.Now()
	fail := func(err error) TestResult {
		return TestResult{OK: false, LatencyMs: time.Since(start).Milliseconds(), Message: s.redact(err.Error())}
	}

	b, key, err := s.openBucket(ctx, d)
	if err != nil {
		return fail(err)
	}
	defer func() { _ = b.Close() }()

	probe := fmt.Sprintf(".homebox-healthcheck-%s", uuid.NewString())
	if d.Type == destTypePrimary {
		probe = fmt.Sprintf("%s/exports/%s", gid, probe)
	} else if d.Prefix != "" {
		probe = d.Prefix + "/" + probe
	}
	k := key(probe)
	if err := b.WriteAll(ctx, k, make([]byte, 1024), &blob.WriterOptions{ContentType: "application/octet-stream"}); err != nil {
		return fail(fmt.Errorf("write test object: %w", err))
	}
	if err := b.Delete(ctx, k); err != nil && gcerrors.Code(err) != gcerrors.NotFound {
		return fail(fmt.Errorf("delete test object: %w", err))
	}
	return TestResult{OK: true, LatencyMs: time.Since(start).Milliseconds(), Message: "Wrote and deleted a 1 KB test file"}
}

// ---------------------------------------------------------------------------
// Destination CRUD

func (s *BackupService) ListDestinations(ctx context.Context, gid uuid.UUID) ([]repo.BackupDestinationOut, error) {
	return s.repos.BackupDestinations.ListByGroup(ctx, gid)
}

func (s *BackupService) GetDestination(ctx context.Context, gid, id uuid.UUID) (repo.BackupDestinationOut, error) {
	return s.repos.BackupDestinations.Get(ctx, gid, id)
}

func (s *BackupService) CreateDestination(ctx context.Context, gid uuid.UUID, in repo.BackupSettings) (repo.BackupDestinationOut, error) {
	if !s.cfg.Enabled {
		return repo.BackupDestinationOut{}, ErrBackupDisabled
	}
	existing, err := s.repos.BackupDestinations.ListByGroup(ctx, gid)
	if err != nil {
		return repo.BackupDestinationOut{}, err
	}
	if len(existing) >= maxDestinationsPerGroup {
		return repo.BackupDestinationOut{}, invalid("a collection can have at most %d backup destinations", maxDestinationsPerGroup)
	}
	in, err = s.NormalizeSettings(in)
	if err != nil {
		return repo.BackupDestinationOut{}, err
	}
	return s.repos.BackupDestinations.Create(ctx, gid, in, s.initialNextRun(in))
}

func (s *BackupService) UpdateDestination(ctx context.Context, gid, id uuid.UUID, in repo.BackupSettings) (repo.BackupDestinationOut, error) {
	if !s.cfg.Enabled {
		return repo.BackupDestinationOut{}, ErrBackupDisabled
	}
	cur, err := s.repos.BackupDestinations.Get(ctx, gid, id)
	if err != nil {
		return repo.BackupDestinationOut{}, err
	}
	in, err = s.NormalizeSettings(in)
	if err != nil {
		return repo.BackupDestinationOut{}, err
	}
	moved := cur.Type != in.Type || cur.ConnString != in.ConnString || cur.Prefix != in.Prefix
	out, err := s.repos.BackupDestinations.Update(ctx, gid, id, in, s.initialNextRun(in), moved)
	if err == nil {
		s.publishMutation(gid)
	}
	return out, err
}

func (s *BackupService) initialNextRun(in repo.BackupSettings) *time.Time {
	if !in.Enabled || !in.ScheduleEnabled {
		return nil
	}
	t := NextRun(in, time.Now())
	return &t
}

// DeleteDestination removes the destination and its history rows. Backup
// files already written to the destination are left untouched.
func (s *BackupService) DeleteDestination(ctx context.Context, gid, id uuid.UUID) error {
	if _, err := s.repos.BackupDestinations.Get(ctx, gid, id); err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return err
	}
	if _, err := s.repos.Exports.DeleteByDestination(ctx, gid, id); err != nil {
		return err
	}
	if _, err := s.repos.BackupDestinations.Delete(ctx, gid, id); err != nil {
		return err
	}
	s.publishMutation(gid)
	return nil
}

// TestSettings checks unsaved settings, so the UI can verify a destination
// before it is stored.
func (s *BackupService) TestSettings(ctx context.Context, gid uuid.UUID, in repo.BackupSettings) (TestResult, error) {
	if !s.cfg.Enabled {
		return TestResult{}, ErrBackupDisabled
	}
	in, err := s.NormalizeSettings(in)
	if err != nil {
		return TestResult{}, err
	}
	return s.Check(ctx, repo.BackupDestinationOut{BackupSettings: in}, gid), nil
}

// TestDestination checks a saved destination and records the result as its
// health status.
func (s *BackupService) TestDestination(ctx context.Context, gid, id uuid.UUID) (TestResult, error) {
	d, err := s.repos.BackupDestinations.Get(ctx, gid, id)
	if err != nil {
		return TestResult{}, err
	}
	res := s.Check(ctx, d, gid)
	s.recordHealth(ctx, d, res)
	return res, nil
}

// RunNow starts an on-demand backup to a destination.
func (s *BackupService) RunNow(ctx context.Context, gid, id uuid.UUID) (repo.ExportOut, error) {
	if !s.cfg.Enabled {
		return repo.ExportOut{}, ErrBackupDisabled
	}
	d, err := s.repos.BackupDestinations.Get(ctx, gid, id)
	if err != nil {
		return repo.ExportOut{}, err
	}
	if !d.Enabled {
		return repo.ExportOut{}, invalid("destination is disabled")
	}
	active, err := s.repos.Exports.HasActiveForDestination(ctx, gid, id)
	if err != nil {
		return repo.ExportOut{}, err
	}
	if active {
		return repo.ExportOut{}, invalid("a backup to this destination is already running")
	}
	return s.exports.EnqueueForDestination(ctx, gid, id, originManual)
}

func (s *BackupService) ListVersions(ctx context.Context, gid, id uuid.UUID) ([]repo.ExportOut, error) {
	if _, err := s.repos.BackupDestinations.Get(ctx, gid, id); err != nil {
		return nil, err
	}
	return s.repos.Exports.ListByDestination(ctx, gid, id)
}

func (s *BackupService) publishMutation(gid uuid.UUID) {
	if s.bus != nil {
		s.bus.Publish(eventbus.EventExportMutation, eventbus.GroupMutationEvent{GID: gid})
	}
}

// ---------------------------------------------------------------------------
// Schedule

// NextRun returns the first scheduled time strictly after `after`, in the
// server's local time zone.
func NextRun(c repo.BackupSettings, after time.Time) time.Time {
	loc := time.Local
	after = after.In(loc)
	y, m, day := after.Date()

	switch c.Frequency {
	case "hourly":
		t := time.Date(y, m, day, after.Hour(), c.AtMinute, 0, 0, loc)
		if !t.After(after) {
			t = t.Add(time.Hour)
		}
		if c.IntervalHours > 1 {
			t = t.Add(time.Duration(c.IntervalHours-1) * time.Hour)
		}
		return t
	case "weekly":
		t := time.Date(y, m, day, c.AtHour, c.AtMinute, 0, 0, loc)
		for i := 0; i < 8; i++ {
			if int(t.Weekday()) == c.Weekday && t.After(after) {
				return t
			}
			t = t.AddDate(0, 0, 1)
		}
		return t
	case "monthly":
		t := time.Date(y, m, c.DayOfMonth, c.AtHour, c.AtMinute, 0, 0, loc)
		if !t.After(after) {
			t = time.Date(y, m+1, c.DayOfMonth, c.AtHour, c.AtMinute, 0, 0, loc)
		}
		return t
	default: // daily
		t := time.Date(y, m, day, c.AtHour, c.AtMinute, 0, 0, loc)
		if !t.After(after) {
			t = t.AddDate(0, 0, 1)
		}
		return t
	}
}

// SchedulerTick starts every backup that has come due and raises "no recent
// successful backup" alerts. It is safe to call often; each destination's
// next_run_at is advanced before its backup is enqueued so a slow run is
// never started twice.
func (s *BackupService) SchedulerTick(ctx context.Context) {
	dests, err := s.repos.BackupDestinations.ListScheduled(ctx)
	if err != nil {
		log.Err(err).Msg("backup scheduler: list destinations")
		return
	}
	now := time.Now()
	for _, d := range dests {
		s.checkStale(ctx, d, now)

		if d.NextRunAt == nil {
			if err := s.repos.BackupDestinations.SetNextRun(ctx, d.ID, NextRun(d.BackupSettings, now), nil); err != nil {
				log.Err(err).Stringer("destination_id", d.ID).Msg("backup scheduler: set next run")
			}
			continue
		}
		if d.NextRunAt.After(now) {
			continue
		}

		next := NextRun(d.BackupSettings, now)
		active, err := s.repos.Exports.HasActiveForDestination(ctx, d.GroupID, d.ID)
		if err != nil {
			log.Err(err).Stringer("destination_id", d.ID).Msg("backup scheduler: check active")
			continue
		}
		if active {
			_ = s.repos.BackupDestinations.SetNextRun(ctx, d.ID, next, nil)
			continue
		}

		if d.SkipIfUnchanged && d.LastFingerprint != "" && d.LastSuccessAt != nil {
			fp, err := s.Fingerprint(ctx, d.GroupID)
			if err == nil && fp == d.LastFingerprint {
				_ = s.repos.BackupDestinations.MarkSkipped(ctx, d.ID, now)
				_ = s.repos.BackupDestinations.SetNextRun(ctx, d.ID, next, nil)
				s.publishMutation(d.GroupID)
				continue
			}
		}

		if err := s.repos.BackupDestinations.SetNextRun(ctx, d.ID, next, &now); err != nil {
			log.Err(err).Stringer("destination_id", d.ID).Msg("backup scheduler: advance next run")
			continue
		}
		if _, err := s.exports.EnqueueForDestination(ctx, d.GroupID, d.ID, originScheduled); err != nil {
			log.Err(err).Stringer("destination_id", d.ID).Msg("backup scheduler: enqueue")
			s.recordFailure(ctx, d, "failed to start backup: "+err.Error())
		}
	}
}

func (s *BackupService) checkStale(ctx context.Context, d repo.BackupDestinationOut, now time.Time) {
	if !d.AlertsEnabled || d.AlertStaleHours <= 0 || d.AlertedStale {
		return
	}
	ref := d.CreatedAt
	if d.LastSuccessAt != nil {
		ref = *d.LastSuccessAt
	}
	if now.Sub(ref) < time.Duration(d.AlertStaleHours)*time.Hour {
		return
	}
	s.alert(ctx, d, fmt.Sprintf("no successful backup in the last %d hours.", d.AlertStaleHours))
	_ = s.repos.BackupDestinations.SetAlertFlags(ctx, d.ID, d.AlertedUnreachable, d.AlertedFailure, true)
}

// ---------------------------------------------------------------------------
// Change detection

// fingerprintSpecs lists what a backup captures. Every table with an
// updated_at column contributes its row count and newest timestamp, so edits,
// additions and deletions all change the fingerprint. tag_entities is a bare
// join table, so only its count is used.
type fingerprintSpec struct {
	table      string
	scope      string
	hasUpdated bool
}

func fingerprintSpecs() []fingerprintSpec {
	specs := make([]fingerprintSpec, 0, len(exportTables)+1)
	specs = append(specs, fingerprintSpec{table: "groups", scope: "id = ?", hasUpdated: true})
	for _, t := range exportTables {
		specs = append(specs, fingerprintSpec{table: t.name, scope: t.scope, hasUpdated: t.name != "tag_entities"})
	}
	return specs
}

// Fingerprint returns a hash that changes whenever the group's backed-up data
// or settings change.
func (s *BackupService) Fingerprint(ctx context.Context, gid uuid.UUID) (string, error) {
	h := sha256.New()
	db := s.db.Sql()
	for _, spec := range fingerprintSpecs() {
		q := "SELECT COUNT(*)"
		if spec.hasUpdated {
			q += ", MAX(updated_at)"
		}
		q += " FROM " + spec.table
		var args []any
		if spec.scope != "" {
			q += " WHERE " + rebindPlaceholders(spec.scope, s.dialect)
			for i := 0; i < strings.Count(spec.scope, "?"); i++ {
				args = append(args, gid.String())
			}
		}
		var (
			count int64
			maxTS any
		)
		var err error
		if spec.hasUpdated {
			err = db.QueryRowContext(ctx, q, args...).Scan(&count, &maxTS)
		} else {
			err = db.QueryRowContext(ctx, q, args...).Scan(&count)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("fingerprint %s: %w", spec.table, err)
		}
		_, _ = fmt.Fprintf(h, "%s|%d|%v\n", spec.table, count, normalizeScan(maxTS))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ---------------------------------------------------------------------------
// Run bookkeeping (called by ExportService)

// beforeRun captures the data fingerprint a destination backup starts from.
func (s *BackupService) beforeRun(ctx context.Context, gid uuid.UUID) string {
	fp, err := s.Fingerprint(ctx, gid)
	if err != nil {
		log.Warn().Err(err).Msg("backup: fingerprint failed; skip-if-unchanged will not apply to this run")
		return ""
	}
	return fp
}

// afterRun records the outcome of a destination-bound export, prunes old
// versions on success and raises or clears alerts.
func (s *BackupService) afterRun(ctx context.Context, gid uuid.UUID, exp repo.ExportOut, fingerprint string, runErr error) {
	if exp.DestinationID == nil {
		return
	}
	d, err := s.repos.BackupDestinations.Get(ctx, gid, *exp.DestinationID)
	if err != nil {
		return // destination deleted mid-run
	}
	if runErr != nil {
		s.recordFailure(ctx, d, runErr.Error())
		return
	}

	now := time.Now()
	if err := s.repos.BackupDestinations.MarkRunSucceeded(ctx, d.ID, now); err != nil {
		log.Err(err).Msg("backup: record success")
	}
	if fingerprint != "" {
		_ = s.repos.BackupDestinations.SetFingerprint(ctx, d.ID, fingerprint)
	}
	if d.AlertedFailure && d.AlertsEnabled {
		s.alert(ctx, d, "backups are working again.")
	}
	_ = s.repos.BackupDestinations.SetAlertFlags(ctx, d.ID, d.AlertedUnreachable, false, false)
	s.Prune(ctx, d)
	s.publishMutation(gid)
}

func (s *BackupService) recordFailure(ctx context.Context, d repo.BackupDestinationOut, msg string) {
	msg = s.redact(msg)
	if err := s.repos.BackupDestinations.MarkRunFailed(ctx, d.ID, time.Now(), msg); err != nil {
		log.Err(err).Msg("backup: record failure")
	}
	if d.AlertsEnabled && !d.AlertedFailure {
		s.alert(ctx, d, "backup failed: "+msg)
		_ = s.repos.BackupDestinations.SetAlertFlags(ctx, d.ID, d.AlertedUnreachable, true, d.AlertedStale)
	}
	s.publishMutation(d.GroupID)
}

// ---------------------------------------------------------------------------
// Retention

// SelectPrunable returns the completed backups a destination's retention
// policy no longer keeps. The newest backup is always kept. For each of the
// daily, weekly and monthly policies the newest backup in each of the most
// recent N calendar buckets is kept; a policy of 0 is unused. When all three
// are 0 nothing is pruned.
func SelectPrunable(rows []repo.ExportOut, daily, weekly, monthly int) []repo.ExportOut {
	if daily <= 0 && weekly <= 0 && monthly <= 0 {
		return nil
	}
	completed := make([]repo.ExportOut, 0, len(rows))
	for _, r := range rows {
		if r.Status == "completed" {
			completed = append(completed, r)
		}
	}
	sort.Slice(completed, func(i, j int) bool { return completed[i].CreatedAt.After(completed[j].CreatedAt) })

	keep := make(map[uuid.UUID]bool, len(completed))
	if len(completed) > 0 {
		keep[completed[0].ID] = true
	}
	mark := func(limit int, bucket func(time.Time) string) {
		if limit <= 0 {
			return
		}
		seen := map[string]bool{}
		for _, r := range completed {
			b := bucket(r.CreatedAt.In(time.Local))
			if seen[b] {
				continue
			}
			if len(seen) >= limit {
				break
			}
			seen[b] = true
			keep[r.ID] = true
		}
	}
	mark(daily, func(t time.Time) string { return t.Format("2006-01-02") })
	mark(weekly, func(t time.Time) string {
		y, w := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", y, w)
	})
	mark(monthly, func(t time.Time) string { return t.Format("2006-01") })

	var out []repo.ExportOut
	for _, r := range completed {
		if !keep[r.ID] {
			out = append(out, r)
		}
	}
	return out
}

// Prune applies the destination's retention policy, deleting each pruned
// artifact before its row, and drops long-stale failed rows.
func (s *BackupService) Prune(ctx context.Context, d repo.BackupDestinationOut) {
	rows, err := s.repos.Exports.ListByDestination(ctx, d.GroupID, d.ID)
	if err != nil {
		log.Err(err).Msg("backup prune: list")
		return
	}

	victims := SelectPrunable(rows, d.KeepDaily, d.KeepWeekly, d.KeepMonthly)
	cutoff := time.Now().Add(-failedRowRetention)
	for _, r := range rows {
		if r.Status == "failed" && r.CreatedAt.Before(cutoff) {
			victims = append(victims, r)
		}
	}
	for _, v := range victims {
		if err := s.DeleteVersion(ctx, d.GroupID, v); err != nil {
			log.Warn().Err(err).Stringer("export_id", v.ID).Msg("backup prune: delete failed; leaving for next run")
		}
	}
}

// DeleteVersion removes a backup's artifact and then its row. A missing
// artifact counts as already deleted.
func (s *BackupService) DeleteVersion(ctx context.Context, gid uuid.UUID, exp repo.ExportOut) error {
	if exp.ArtifactPath != "" {
		b, key, err := s.ArtifactLocation(ctx, gid, exp)
		if err != nil {
			return err
		}
		err = b.Delete(ctx, key)
		_ = b.Close()
		if err != nil && gcerrors.Code(err) != gcerrors.NotFound {
			return err
		}
	}
	_, err := s.repos.Exports.Delete(ctx, gid, exp.ID)
	return err
}

// ---------------------------------------------------------------------------
// Health

// HealthTick probes every enabled destination whose check interval elapsed.
func (s *BackupService) HealthTick(ctx context.Context) {
	dests, err := s.repos.BackupDestinations.ListEnabled(ctx)
	if err != nil {
		log.Err(err).Msg("backup health: list destinations")
		return
	}
	now := time.Now()
	for _, d := range dests {
		if d.HealthCheckedAt != nil && now.Sub(*d.HealthCheckedAt) < time.Duration(d.HealthIntervalMinutes)*time.Minute {
			continue
		}
		s.recordHealth(ctx, d, s.Check(ctx, d, d.GroupID))
	}
}

// recordHealth stores a probe result and raises or clears the unreachable
// alert. An outage alerts once, after AlertFailureThreshold consecutive
// failed probes, and a recovery notice follows when the destination returns.
func (s *BackupService) recordHealth(ctx context.Context, d repo.BackupDestinationOut, res TestResult) {
	now := time.Now()
	if res.OK {
		if err := s.repos.BackupDestinations.SetHealth(ctx, d.ID, healthHealthy, now, "", 0); err != nil {
			log.Err(err).Msg("backup health: record")
			return
		}
		if d.AlertedUnreachable {
			if d.AlertsEnabled {
				s.alert(ctx, d, "destination is reachable again.")
			}
			_ = s.repos.BackupDestinations.SetAlertFlags(ctx, d.ID, false, d.AlertedFailure, d.AlertedStale)
		}
		s.publishMutation(d.GroupID)
		return
	}

	failures := d.HealthFailures + 1
	if err := s.repos.BackupDestinations.SetHealth(ctx, d.ID, healthUnreachable, now, res.Message, failures); err != nil {
		log.Err(err).Msg("backup health: record")
		return
	}
	if d.AlertsEnabled && !d.AlertedUnreachable && failures >= d.AlertFailureThreshold {
		s.alert(ctx, d, fmt.Sprintf("destination is unreachable after %d failed checks: %s", failures, res.Message))
		_ = s.repos.BackupDestinations.SetAlertFlags(ctx, d.ID, true, d.AlertedFailure, d.AlertedStale)
	}
	s.publishMutation(d.GroupID)
}

// ---------------------------------------------------------------------------
// Alerts

// alert sends msg to the group's active notifiers. The in-app banner is
// driven by the persisted destination state, not by this call.
func (s *BackupService) alert(ctx context.Context, d repo.BackupDestinationOut, msg string) {
	log.Warn().Stringer("destination_id", d.ID).Str("destination", d.Name).Msg("backup alert: " + msg)

	notifiers, err := s.repos.Notifiers.GetActiveByGroup(ctx, d.GroupID)
	if err != nil {
		log.Err(err).Msg("backup alert: list notifiers")
		return
	}
	text := fmt.Sprintf("Homebox backup alert for %q: %s", d.Name, msg)
	for i := range notifiers {
		if err := validate.SendNotifierMessage(notifiers[i].URL, text, s.notifierConfig); err != nil {
			log.Err(err).Str("notifier", notifiers[i].Name).Msg("backup alert: send failed")
		}
	}
}
