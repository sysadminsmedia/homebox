-- +goose Up
-- Modify "items" table
ALTER TABLE "items" ADD COLUMN "sync_child_items_locations" boolean NOT NULL DEFAULT false;
