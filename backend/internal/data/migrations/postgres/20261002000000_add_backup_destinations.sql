-- +goose Up
-- Scheduled backups: per-collection destinations with their own schedule,
-- retention, health status and alert state, plus origin/destination tracking
-- on the existing exports table so scheduled artifacts are pruned by
-- retention instead of the fixed 7-day sweep.
CREATE TABLE IF NOT EXISTS "backup_destinations" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "name" character varying NOT NULL,
    "description" character varying NULL,
    "type" character varying NOT NULL DEFAULT 'primary'
        CHECK ("type" IN ('primary', 'local', 's3', 'gcs', 'azblob')),
    "conn_string" character varying NULL,
    "prefix" character varying NOT NULL DEFAULT 'homebox-backups',
    "enabled" boolean NOT NULL DEFAULT true,
    "schedule_enabled" boolean NOT NULL DEFAULT false,
    "frequency" character varying NOT NULL DEFAULT 'daily'
        CHECK ("frequency" IN ('hourly', 'daily', 'weekly', 'monthly')),
    "interval_hours" bigint NOT NULL DEFAULT 1,
    "at_hour" bigint NOT NULL DEFAULT 3,
    "at_minute" bigint NOT NULL DEFAULT 0,
    "weekday" bigint NOT NULL DEFAULT 0,
    "day_of_month" bigint NOT NULL DEFAULT 1,
    "skip_if_unchanged" boolean NOT NULL DEFAULT true,
    "keep_daily" bigint NOT NULL DEFAULT 7,
    "keep_weekly" bigint NOT NULL DEFAULT 4,
    "keep_monthly" bigint NOT NULL DEFAULT 6,
    "next_run_at" timestamptz NULL,
    "last_run_at" timestamptz NULL,
    "last_success_at" timestamptz NULL,
    "last_skipped_at" timestamptz NULL,
    "last_fingerprint" character varying NULL,
    "last_error" character varying NULL,
    "health_status" character varying NOT NULL DEFAULT 'unknown'
        CHECK ("health_status" IN ('unknown', 'healthy', 'unreachable')),
    "health_checked_at" timestamptz NULL,
    "health_error" character varying NULL,
    "health_failures" bigint NOT NULL DEFAULT 0,
    "health_interval_minutes" bigint NOT NULL DEFAULT 15,
    "alerts_enabled" boolean NOT NULL DEFAULT true,
    "alert_failure_threshold" bigint NOT NULL DEFAULT 2,
    "alert_stale_hours" bigint NOT NULL DEFAULT 48,
    "alerted_unreachable" boolean NOT NULL DEFAULT false,
    "alerted_failure" boolean NOT NULL DEFAULT false,
    "alerted_stale" boolean NOT NULL DEFAULT false,
    "group_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "backup_destinations_groups_backup_destinations" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "backupdestination_group_id" ON "backup_destinations" ("group_id");

ALTER TABLE "exports"
    ADD COLUMN IF NOT EXISTS "origin" character varying NOT NULL DEFAULT 'manual'
        CHECK ("origin" IN ('manual', 'scheduled')),
    ADD COLUMN IF NOT EXISTS "destination_id" uuid NULL;
