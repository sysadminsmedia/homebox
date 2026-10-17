-- +goose Up
-- Cloud-drive backup destinations (Google Drive, OneDrive, Dropbox): three new
-- types. Their OAuth refresh token is stored sealed in the existing secret
-- column and the connected account's name in username.
--
-- SQLite cannot alter a CHECK constraint, so the table is rebuilt. Nothing
-- references backup_destinations, so the swap is safe.
create table backup_destinations_new
(
    id                      uuid                         not null
        primary key,
    created_at              datetime                     not null,
    updated_at              datetime                     not null,
    name                    text                         not null,
    description             text,
    type                    text     default 'primary'   not null
        check (type in ('primary', 'local', 's3', 'gcs', 'azblob', 'sftp', 'webdav', 'gdrive', 'onedrive', 'dropbox')),
    conn_string             text,
    username                text,
    secret                  text,
    host_key                text,
    prefix                  text     default 'homebox-backups' not null,
    enabled                 bool     default true        not null,
    schedule_enabled        bool     default false       not null,
    frequency               text     default 'daily'     not null
        check (frequency in ('hourly', 'daily', 'weekly', 'monthly')),
    interval_hours          integer  default 1           not null,
    at_hour                 integer  default 3           not null,
    at_minute               integer  default 0           not null,
    weekday                 integer  default 0           not null,
    day_of_month            integer  default 1           not null,
    skip_if_unchanged       bool     default true        not null,
    keep_daily              integer  default 7           not null,
    keep_weekly             integer  default 4           not null,
    keep_monthly            integer  default 6           not null,
    next_run_at             datetime,
    last_run_at             datetime,
    last_success_at         datetime,
    last_skipped_at         datetime,
    last_fingerprint        text,
    last_error              text,
    health_status           text     default 'unknown'   not null
        check (health_status in ('unknown', 'healthy', 'unreachable')),
    health_checked_at       datetime,
    health_error            text,
    health_failures         integer  default 0           not null,
    health_interval_minutes integer  default 15          not null,
    alerts_enabled          bool     default true        not null,
    alert_failure_threshold integer  default 2           not null,
    alert_stale_hours       integer  default 48          not null,
    alerted_unreachable     bool     default false       not null,
    alerted_failure         bool     default false       not null,
    alerted_stale           bool     default false       not null,
    group_id                uuid                         not null
        constraint backup_destinations_groups_backup_destinations
            references groups
            on delete cascade
);

insert into backup_destinations_new
    (id, created_at, updated_at, name, description, type, conn_string, username, secret, host_key, prefix, enabled, schedule_enabled,
     frequency, interval_hours, at_hour, at_minute, weekday, day_of_month, skip_if_unchanged,
     keep_daily, keep_weekly, keep_monthly, next_run_at, last_run_at, last_success_at, last_skipped_at,
     last_fingerprint, last_error, health_status, health_checked_at, health_error, health_failures,
     health_interval_minutes, alerts_enabled, alert_failure_threshold, alert_stale_hours,
     alerted_unreachable, alerted_failure, alerted_stale, group_id)
select id, created_at, updated_at, name, description, type, conn_string, username, secret, host_key, prefix, enabled, schedule_enabled,
       frequency, interval_hours, at_hour, at_minute, weekday, day_of_month, skip_if_unchanged,
       keep_daily, keep_weekly, keep_monthly, next_run_at, last_run_at, last_success_at, last_skipped_at,
       last_fingerprint, last_error, health_status, health_checked_at, health_error, health_failures,
       health_interval_minutes, alerts_enabled, alert_failure_threshold, alert_stale_hours,
       alerted_unreachable, alerted_failure, alerted_stale, group_id
from backup_destinations;

drop table backup_destinations;

alter table backup_destinations_new rename to backup_destinations;

create index if not exists backupdestination_group_id
    on backup_destinations (group_id);
