-- +goose Up
-- +goose no transaction
-- SQLite uses dynamic typing; existing INTEGER-affinity quantity columns can store fractional values.
-- No table rewrite is required for quantity/default_quantity decimal support.

