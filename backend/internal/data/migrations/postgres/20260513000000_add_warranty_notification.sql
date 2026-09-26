-- +goose Up
ALTER TABLE entities ADD COLUMN notify_warranty_expiration boolean NOT NULL DEFAULT false;
