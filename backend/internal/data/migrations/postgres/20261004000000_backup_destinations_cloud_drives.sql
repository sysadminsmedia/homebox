-- +goose Up
-- Cloud-drive backup destinations (Google Drive, OneDrive, Dropbox): three new
-- types. Their OAuth refresh token is stored sealed in the existing secret
-- column and the connected account's name in username.
ALTER TABLE "backup_destinations" DROP CONSTRAINT IF EXISTS "backup_destinations_type_check";
ALTER TABLE "backup_destinations"
    ADD CONSTRAINT "backup_destinations_type_check"
        CHECK ("type" IN ('primary', 'local', 's3', 'gcs', 'azblob', 'sftp', 'webdav', 'gdrive', 'onedrive', 'dropbox'));
