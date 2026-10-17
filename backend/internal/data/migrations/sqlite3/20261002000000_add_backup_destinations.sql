-- +goose Up
-- Scheduled backups: per-collection destinations with their own schedule,
-- retention, health status and alert state, plus origin/destination tracking
-- on the existing exports table so scheduled artifacts are pruned by
-- retention instead of the fixed 7-day sweep.
create table if not exists backup_destinations
(
    id                      uuid                         not null
        primary key,
    created_at              datetime                     not null,
    updated_at              datetime                     not null,
    name                    text                         not null,
    description             text,
    type                    text     default 'primary'   not null
        check (type in ('primary', 'local', 's3', 'gcs', 'azblob')),
    conn_string             text,
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

create index if not exists backupdestination_group_id
    on backup_destinations (group_id);

alter table exports
    add column origin text default 'manual' not null
        check (origin in ('manual', 'scheduled'));

alter table exports
    add column destination_id uuid;
