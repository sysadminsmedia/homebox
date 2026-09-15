-- +goose Up
ALTER TABLE tags ADD COLUMN icon varchar(255) NULL;
ALTER TABLE tags ADD COLUMN tag_children uuid REFERENCES tags(id) ON DELETE SET NULL;
