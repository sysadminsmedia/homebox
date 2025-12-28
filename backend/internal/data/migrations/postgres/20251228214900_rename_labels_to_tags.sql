-- goose Up
ALTER TABLE labels RENAME TO tags;
ALTER TABLE label_items RENAME TO tag_items;