package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/schema/mixins"
)

// BackupDestination is a place scheduled (and on-demand) backups are written
// to, together with that destination's own schedule, retention policy, health
// status and alert state. Type "primary" targets the instance's main storage
// (HBOX_STORAGE_CONN_STRING); the other types use gocloud.dev/blob drivers.
//
// The gocloud drivers hold no secrets: they read credentials from the
// environment (AWS_*, GOOGLE_APPLICATION_CREDENTIALS, AZURE_*), and the service
// rejects connection strings that embed userinfo. The sftp and webdav types
// need a login, which is stored encrypted in secret.
type BackupDestination struct {
	ent.Schema
}

func (BackupDestination) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.DetailsMixin{},
		GroupMixin{
			ref:   "backup_destinations",
			field: "group_id",
		},
	}
}

func (BackupDestination) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("type").
			Values("primary", "local", "s3", "gcs", "azblob", "sftp", "webdav", "gdrive", "onedrive", "dropbox", "smb").
			Default("primary"),
		// conn_string is a sub-directory of the configured local backup root
		// for type=local, and a gocloud URL (s3://bucket?region=...) for the
		// cloud types. Unused for type=primary.
		field.String("conn_string").
			MaxLen(2048).
			Optional(),
		// username, secret and host_key serve the sftp and webdav types. For
		// the cloud-drive types (gdrive, onedrive, dropbox) username holds the
		// connected account's name and secret the OAuth refresh token.
		// secret is an AES-GCM sealed JSON blob holding the credentials; it is
		// never returned by the API.
		field.String("username").
			MaxLen(255).
			Optional(),
		field.Text("secret").
			Optional().
			Sensitive(),
		// host_key is the SSH host key fingerprint ("SHA256:...") an sftp
		// destination must present.
		field.String("host_key").
			MaxLen(255).
			Optional(),
		field.String("prefix").
			MaxLen(255).
			Default("homebox-backups"),
		field.Bool("enabled").
			Default(true),

		field.Bool("schedule_enabled").
			Default(false),
		field.Enum("frequency").
			Values("hourly", "daily", "weekly", "monthly", "cron").
			Default("daily"),
		// cron_expr is a standard 5-field cron expression (or descriptor such
		// as @daily), used when frequency is "cron".
		field.String("cron_expr").
			MaxLen(255).
			Optional(),
		field.Int("interval_hours").
			Default(1),
		field.Int("at_hour").
			Default(3),
		field.Int("at_minute").
			Default(0),
		// weekday uses Go's time.Weekday numbering: 0 = Sunday.
		field.Int("weekday").
			Default(0),
		field.Int("day_of_month").
			Default(1),
		field.Bool("skip_if_unchanged").
			Default(true),

		field.Int("keep_daily").
			Default(7),
		field.Int("keep_weekly").
			Default(4),
		field.Int("keep_monthly").
			Default(6),

		field.Time("next_run_at").
			Optional().
			Nillable(),
		field.Time("last_run_at").
			Optional().
			Nillable(),
		field.Time("last_success_at").
			Optional().
			Nillable(),
		field.Time("last_skipped_at").
			Optional().
			Nillable(),
		// last_fingerprint is the data fingerprint captured when the most
		// recent successful backup started; an unchanged fingerprint means a
		// scheduled run can be skipped.
		field.String("last_fingerprint").
			MaxLen(128).
			Optional(),
		field.String("last_error").
			MaxLen(1000).
			Optional(),

		field.Enum("health_status").
			Values("unknown", "healthy", "unreachable").
			Default("unknown"),
		field.Time("health_checked_at").
			Optional().
			Nillable(),
		field.String("health_error").
			MaxLen(1000).
			Optional(),
		field.Int("health_failures").
			Default(0),
		field.Int("health_interval_minutes").
			Default(15),

		field.Bool("alerts_enabled").
			Default(true),
		field.Int("alert_failure_threshold").
			Default(2),
		field.Int("alert_stale_hours").
			Default(48),
		// Alert de-duplication: each flag is set when its alert is sent and
		// cleared when the condition recovers, so an outage notifies once.
		field.Bool("alerted_unreachable").
			Default(false),
		field.Bool("alerted_failure").
			Default(false),
		field.Bool("alerted_stale").
			Default(false),
	}
}

func (BackupDestination) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_id"),
	}
}
