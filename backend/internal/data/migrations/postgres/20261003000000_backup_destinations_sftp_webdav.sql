-- +goose Up
-- SFTP and WebDAV backup destinations: two new types, a login, an encrypted
-- secret (AES-GCM sealed by the service) and the SSH host key fingerprint.
ALTER TABLE "backup_destinations" DROP CONSTRAINT IF EXISTS "backup_destinations_type_check";
ALTER TABLE "backup_destinations"
    ADD CONSTRAINT "backup_destinations_type_check"
        CHECK ("type" IN ('primary', 'local', 's3', 'gcs', 'azblob', 'sftp', 'webdav')),
    ADD COLUMN IF NOT EXISTS "username" character varying NULL,
    ADD COLUMN IF NOT EXISTS "secret" text NULL,
    ADD COLUMN IF NOT EXISTS "host_key" character varying NULL;
