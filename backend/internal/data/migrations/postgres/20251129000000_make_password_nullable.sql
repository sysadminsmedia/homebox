-- +goose Up
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
