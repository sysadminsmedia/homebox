package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/backupdestination"
)

// BackupDestinationRepository persists backup destinations. Methods that take
// a gid refuse to act on rows owned by a different group; the scheduler
// methods (ListScheduled, ListEnabled) deliberately sweep every tenant.
type BackupDestinationRepository struct {
	db *ent.Client
}

// BackupSettings is the user-editable part of a destination.
type BackupSettings struct {
	Name        string `json:"name"        validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
	// Type is one of primary, local, s3, gcs, azblob.
	Type       string `json:"type"       validate:"required,oneof=primary local s3 gcs azblob"`
	ConnString string `json:"connString" validate:"max=2048"`
	Prefix     string `json:"prefix"     validate:"max=255"`
	Enabled    bool   `json:"enabled"`

	ScheduleEnabled bool `json:"scheduleEnabled"`
	// Frequency is one of hourly, daily, weekly, monthly.
	Frequency       string `json:"frequency"       validate:"required,oneof=hourly daily weekly monthly"`
	IntervalHours   int    `json:"intervalHours"   validate:"min=1,max=168"`
	AtHour          int    `json:"atHour"          validate:"min=0,max=23"`
	AtMinute        int    `json:"atMinute"        validate:"min=0,max=59"`
	Weekday         int    `json:"weekday"         validate:"min=0,max=6"`
	DayOfMonth      int    `json:"dayOfMonth"      validate:"min=1,max=28"`
	SkipIfUnchanged bool   `json:"skipIfUnchanged"`

	KeepDaily   int `json:"keepDaily"   validate:"min=0,max=3650"`
	KeepWeekly  int `json:"keepWeekly"  validate:"min=0,max=520"`
	KeepMonthly int `json:"keepMonthly" validate:"min=0,max=120"`

	HealthIntervalMinutes int  `json:"healthIntervalMinutes" validate:"min=1,max=1440"`
	AlertsEnabled         bool `json:"alertsEnabled"`
	AlertFailureThreshold int  `json:"alertFailureThreshold" validate:"min=1,max=100"`
	// AlertStaleHours alerts when no backup succeeded for this long. 0 disables.
	AlertStaleHours int `json:"alertStaleHours" validate:"min=0,max=8760"`
}

type BackupDestinationOut struct {
	ID        uuid.UUID `json:"id"`
	GroupID   uuid.UUID `json:"groupId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	BackupSettings

	NextRunAt       *time.Time `json:"nextRunAt,omitempty"       extensions:"x-nullable"`
	LastRunAt       *time.Time `json:"lastRunAt,omitempty"       extensions:"x-nullable"`
	LastSuccessAt   *time.Time `json:"lastSuccessAt,omitempty"   extensions:"x-nullable"`
	LastSkippedAt   *time.Time `json:"lastSkippedAt,omitempty"   extensions:"x-nullable"`
	LastError       string     `json:"lastError,omitempty"`
	HealthStatus    string     `json:"healthStatus"`
	HealthCheckedAt *time.Time `json:"healthCheckedAt,omitempty" extensions:"x-nullable"`
	HealthError     string     `json:"healthError,omitempty"`
	HealthFailures  int        `json:"healthFailures"`

	// LastFingerprint and the alerted* flags are internal bookkeeping.
	LastFingerprint    string `json:"-"`
	AlertedUnreachable bool   `json:"-"`
	AlertedFailure     bool   `json:"-"`
	AlertedStale       bool   `json:"-"`
}

func mapBackupDestination(d *ent.BackupDestination) BackupDestinationOut {
	return BackupDestinationOut{
		ID:        d.ID,
		GroupID:   d.GroupID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		BackupSettings: BackupSettings{
			Name:                  d.Name,
			Description:           d.Description,
			Type:                  string(d.Type),
			ConnString:            d.ConnString,
			Prefix:                d.Prefix,
			Enabled:               d.Enabled,
			ScheduleEnabled:       d.ScheduleEnabled,
			Frequency:             string(d.Frequency),
			IntervalHours:         d.IntervalHours,
			AtHour:                d.AtHour,
			AtMinute:              d.AtMinute,
			Weekday:               d.Weekday,
			DayOfMonth:            d.DayOfMonth,
			SkipIfUnchanged:       d.SkipIfUnchanged,
			KeepDaily:             d.KeepDaily,
			KeepWeekly:            d.KeepWeekly,
			KeepMonthly:           d.KeepMonthly,
			HealthIntervalMinutes: d.HealthIntervalMinutes,
			AlertsEnabled:         d.AlertsEnabled,
			AlertFailureThreshold: d.AlertFailureThreshold,
			AlertStaleHours:       d.AlertStaleHours,
		},
		NextRunAt:          d.NextRunAt,
		LastRunAt:          d.LastRunAt,
		LastSuccessAt:      d.LastSuccessAt,
		LastSkippedAt:      d.LastSkippedAt,
		LastError:          d.LastError,
		HealthStatus:       string(d.HealthStatus),
		HealthCheckedAt:    d.HealthCheckedAt,
		HealthError:        d.HealthError,
		HealthFailures:     d.HealthFailures,
		LastFingerprint:    d.LastFingerprint,
		AlertedUnreachable: d.AlertedUnreachable,
		AlertedFailure:     d.AlertedFailure,
		AlertedStale:       d.AlertedStale,
	}
}

func mapBackupDestinations(rows []*ent.BackupDestination) []BackupDestinationOut {
	out := make([]BackupDestinationOut, len(rows))
	for i, d := range rows {
		out[i] = mapBackupDestination(d)
	}
	return out
}

func (r *BackupDestinationRepository) ListByGroup(ctx context.Context, gid uuid.UUID) ([]BackupDestinationOut, error) {
	rows, err := r.db.BackupDestination.Query().
		Where(backupdestination.GroupID(gid)).
		Order(ent.Asc(backupdestination.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return mapBackupDestinations(rows), nil
}

// ListEnabled returns every enabled destination across all groups, for the
// health-check sweep.
func (r *BackupDestinationRepository) ListEnabled(ctx context.Context) ([]BackupDestinationOut, error) {
	rows, err := r.db.BackupDestination.Query().
		Where(backupdestination.Enabled(true)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return mapBackupDestinations(rows), nil
}

// ListScheduled returns every enabled destination with scheduling turned on,
// across all groups.
func (r *BackupDestinationRepository) ListScheduled(ctx context.Context) ([]BackupDestinationOut, error) {
	rows, err := r.db.BackupDestination.Query().
		Where(backupdestination.Enabled(true), backupdestination.ScheduleEnabled(true)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return mapBackupDestinations(rows), nil
}

func (r *BackupDestinationRepository) Get(ctx context.Context, gid, id uuid.UUID) (BackupDestinationOut, error) {
	d, err := r.db.BackupDestination.Query().
		Where(backupdestination.ID(id), backupdestination.GroupID(gid)).
		Only(ctx)
	if err != nil {
		return BackupDestinationOut{}, err
	}
	return mapBackupDestination(d), nil
}

func (r *BackupDestinationRepository) Create(ctx context.Context, gid uuid.UUID, in BackupSettings, nextRun *time.Time) (BackupDestinationOut, error) {
	c := r.db.BackupDestination.Create().
		SetGroupID(gid).
		SetName(in.Name).
		SetDescription(in.Description).
		SetType(backupdestination.Type(in.Type)).
		SetConnString(in.ConnString).
		SetPrefix(in.Prefix).
		SetEnabled(in.Enabled).
		SetScheduleEnabled(in.ScheduleEnabled).
		SetFrequency(backupdestination.Frequency(in.Frequency)).
		SetIntervalHours(in.IntervalHours).
		SetAtHour(in.AtHour).
		SetAtMinute(in.AtMinute).
		SetWeekday(in.Weekday).
		SetDayOfMonth(in.DayOfMonth).
		SetSkipIfUnchanged(in.SkipIfUnchanged).
		SetKeepDaily(in.KeepDaily).
		SetKeepWeekly(in.KeepWeekly).
		SetKeepMonthly(in.KeepMonthly).
		SetHealthIntervalMinutes(in.HealthIntervalMinutes).
		SetAlertsEnabled(in.AlertsEnabled).
		SetAlertFailureThreshold(in.AlertFailureThreshold).
		SetAlertStaleHours(in.AlertStaleHours).
		SetNillableNextRunAt(nextRun)
	d, err := c.Save(ctx)
	if err != nil {
		return BackupDestinationOut{}, err
	}
	return mapBackupDestination(d), nil
}

// Update replaces the editable settings. A changed destination invalidates
// the stored health status and fingerprint, since they described the old one.
func (r *BackupDestinationRepository) Update(ctx context.Context, gid, id uuid.UUID, in BackupSettings, nextRun *time.Time, resetState bool) (BackupDestinationOut, error) {
	u := r.db.BackupDestination.UpdateOneID(id).
		Where(backupdestination.GroupID(gid)).
		SetName(in.Name).
		SetDescription(in.Description).
		SetType(backupdestination.Type(in.Type)).
		SetConnString(in.ConnString).
		SetPrefix(in.Prefix).
		SetEnabled(in.Enabled).
		SetScheduleEnabled(in.ScheduleEnabled).
		SetFrequency(backupdestination.Frequency(in.Frequency)).
		SetIntervalHours(in.IntervalHours).
		SetAtHour(in.AtHour).
		SetAtMinute(in.AtMinute).
		SetWeekday(in.Weekday).
		SetDayOfMonth(in.DayOfMonth).
		SetSkipIfUnchanged(in.SkipIfUnchanged).
		SetKeepDaily(in.KeepDaily).
		SetKeepWeekly(in.KeepWeekly).
		SetKeepMonthly(in.KeepMonthly).
		SetHealthIntervalMinutes(in.HealthIntervalMinutes).
		SetAlertsEnabled(in.AlertsEnabled).
		SetAlertFailureThreshold(in.AlertFailureThreshold).
		SetAlertStaleHours(in.AlertStaleHours).
		SetNillableNextRunAt(nextRun)
	if nextRun == nil {
		u = u.ClearNextRunAt()
	}
	if resetState {
		u = u.SetHealthStatus(backupdestination.HealthStatusUnknown).
			ClearHealthCheckedAt().
			SetHealthError("").
			SetHealthFailures(0).
			SetLastFingerprint("").
			SetAlertedUnreachable(false)
	}
	d, err := u.Save(ctx)
	if err != nil {
		return BackupDestinationOut{}, err
	}
	return mapBackupDestination(d), nil
}

func (r *BackupDestinationRepository) Delete(ctx context.Context, gid, id uuid.UUID) (int, error) {
	return r.db.BackupDestination.Delete().
		Where(backupdestination.ID(id), backupdestination.GroupID(gid)).
		Exec(ctx)
}

// SetNextRun records the next scheduled run and, when started is non-nil, the
// time the current run was kicked off.
func (r *BackupDestinationRepository) SetNextRun(ctx context.Context, id uuid.UUID, next time.Time, started *time.Time) error {
	u := r.db.BackupDestination.UpdateOneID(id).SetNextRunAt(next)
	if started != nil {
		u = u.SetLastRunAt(*started)
	}
	return u.Exec(ctx)
}

func (r *BackupDestinationRepository) MarkSkipped(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.BackupDestination.UpdateOneID(id).SetLastSkippedAt(at).Exec(ctx)
}

func (r *BackupDestinationRepository) SetFingerprint(ctx context.Context, id uuid.UUID, fp string) error {
	return r.db.BackupDestination.UpdateOneID(id).SetLastFingerprint(fp).Exec(ctx)
}

// MarkRunSucceeded records a successful backup and clears the failure state.
func (r *BackupDestinationRepository) MarkRunSucceeded(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.BackupDestination.UpdateOneID(id).
		SetLastRunAt(at).
		SetLastSuccessAt(at).
		SetLastError("").
		SetAlertedFailure(false).
		SetAlertedStale(false).
		Exec(ctx)
}

func (r *BackupDestinationRepository) MarkRunFailed(ctx context.Context, id uuid.UUID, at time.Time, msg string) error {
	if len(msg) > 1000 {
		msg = msg[:1000]
	}
	return r.db.BackupDestination.UpdateOneID(id).
		SetLastRunAt(at).
		SetLastError(msg).
		Exec(ctx)
}

func (r *BackupDestinationRepository) SetHealth(ctx context.Context, id uuid.UUID, status string, checkedAt time.Time, errMsg string, failures int) error {
	if len(errMsg) > 1000 {
		errMsg = errMsg[:1000]
	}
	return r.db.BackupDestination.UpdateOneID(id).
		SetHealthStatus(backupdestination.HealthStatus(status)).
		SetHealthCheckedAt(checkedAt).
		SetHealthError(errMsg).
		SetHealthFailures(failures).
		Exec(ctx)
}

// SetAlertFlags persists the alert de-duplication flags.
func (r *BackupDestinationRepository) SetAlertFlags(ctx context.Context, id uuid.UUID, unreachable, failure, stale bool) error {
	return r.db.BackupDestination.UpdateOneID(id).
		SetAlertedUnreachable(unreachable).
		SetAlertedFailure(failure).
		SetAlertedStale(stale).
		Exec(ctx)
}
