-- +goose Up
ALTER TABLE users ADD COLUMN settings JSONB DEFAULT '{}';