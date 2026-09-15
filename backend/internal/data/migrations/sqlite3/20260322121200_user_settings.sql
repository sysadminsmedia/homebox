-- +goose Up
ALTER TABLE users ADD COLUMN settings JSON DEFAULT '{}';