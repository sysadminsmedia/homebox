-- +goose Up
-- Rename purchase_time/sold_time → purchase_date/sold_date to make the
-- intended semantics explicit: at the application layer these are calendar
-- dates with no time-of-day or timezone. The misleading "_time" name caused
-- JSON round-trips to drift the day across timezones (issue #437).
--
-- This is a cosmetic rename only. ALTER TABLE ... RENAME COLUMN preserves
-- the underlying datetime type — purchase_date and sold_date remain
-- datetime after this migration runs. Date-only semantics are enforced by
-- the Go layer (internal/data/types.Date), which truncates to YYYY-MM-DD
-- on JSON marshal and via DateFromTime on every write through the repo.
-- Any value written outside that path may still carry a non-midnight time.

ALTER TABLE entities RENAME COLUMN purchase_time TO purchase_date;
ALTER TABLE entities RENAME COLUMN sold_time TO sold_date;

-- +goose Down
ALTER TABLE entities RENAME COLUMN purchase_date TO purchase_time;
ALTER TABLE entities RENAME COLUMN sold_date TO sold_time;
