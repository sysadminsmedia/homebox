-- +goose Up
-- Add parent_id column to labels table for hierarchical label organization
ALTER TABLE labels ADD COLUMN label_children UUID REFERENCES labels(id) ON DELETE CASCADE;
