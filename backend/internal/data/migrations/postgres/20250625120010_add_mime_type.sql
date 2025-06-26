-- +goose Up
ALTER TABLE public.attachments ADD COLUMN mime_type VARCHAR DEFAULT 'application/octet-stream';

