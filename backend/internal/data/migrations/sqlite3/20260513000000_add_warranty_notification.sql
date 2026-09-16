-- +goose Up
ALTER TABLE entities ADD COLUMN notify_warranty_expiration bool default false not null;
