-- +goose Up
-- Add label_children column to labels table for hierarchical label organization
ALTER TABLE labels ADD COLUMN label_children TEXT REFERENCES labels(id) ON DELETE CASCADE;
