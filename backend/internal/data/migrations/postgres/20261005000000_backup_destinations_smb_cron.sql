-- +goose Up
-- Native SMB destinations (new type) and custom cron schedules (new frequency
-- and cron_expr column).
ALTER TABLE "backup_destinations" DROP CONSTRAINT IF EXISTS "backup_destinations_type_check";
ALTER TABLE "backup_destinations" DROP CONSTRAINT IF EXISTS "backup_destinations_frequency_check";
ALTER TABLE "backup_destinations"
    ADD CONSTRAINT "backup_destinations_type_check"
        CHECK ("type" IN ('primary', 'local', 's3', 'gcs', 'azblob', 'sftp', 'webdav', 'gdrive', 'onedrive', 'dropbox', 'smb')),
    ADD CONSTRAINT "backup_destinations_frequency_check"
        CHECK ("frequency" IN ('hourly', 'daily', 'weekly', 'monthly', 'cron')),
    ADD COLUMN IF NOT EXISTS "cron_expr" character varying NULL;
