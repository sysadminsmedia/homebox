-- +goose Up
ALTER TABLE `items` ADD COLUMN `sync_child_items_locations` bool NOT NULL DEFAULT (false);
