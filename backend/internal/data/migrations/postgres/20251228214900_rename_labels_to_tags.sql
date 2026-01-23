-- +goose Up
ALTER TABLE labels RENAME TO tags;
ALTER TABLE label_items RENAME TO tag_items;
ALTER TABLE tag_items RENAME COLUMN label_id TO tag_id;
ALTER TABLE tags RENAME COLUMN group_labels TO group_tags;
ALTER TABLE item_templates RENAME COLUMN default_label_ids TO default_tag_ids;