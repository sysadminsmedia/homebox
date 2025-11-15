-- +goose Up
ALTER TABLE attachments ADD COLUMN mime_type TEXT DEFAULT 'application/octet-stream';

