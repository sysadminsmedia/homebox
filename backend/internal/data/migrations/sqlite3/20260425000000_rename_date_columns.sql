-- +goose Up
-- Rename date-only columns to drop the misleading "_time" suffix.
-- These columns hold calendar dates (no time-of-day, no timezone) — the
-- "_time" name conflated them with full timestamps and caused every JSON
-- round-trip to risk shifting the day across timezones (issue #437).

ALTER TABLE entities RENAME COLUMN purchase_time TO purchase_date;
ALTER TABLE entities RENAME COLUMN sold_time TO sold_date;

-- +goose Down
ALTER TABLE entities RENAME COLUMN purchase_date TO purchase_time;
ALTER TABLE entities RENAME COLUMN sold_date TO sold_time;
