-- +goose Up
ALTER TABLE entity_types ADD COLUMN color varchar(255) NULL;