-- +goose Up
-- Give location its own column again (#1688).
--
-- 0.26.0 merged items and locations into one table with a single self-FK, which
-- made location a derived property: the nearest ancestor that is a location.
-- That can't express "child of item X, stored in Y" — setting a parent drags the
-- child's location along with it.
--
-- entity_location_entities is a sparse override, only set when the parent is a
-- non-location entity; otherwise it stays NULL and the ancestor walk still wins.
-- See repo.resolveLocationOverride.
--
-- No backfill on purpose: existing rows keep NULL and resolve exactly as before,
-- so upgrading changes nothing on screen.
alter table entities
    add column entity_location_entities uuid
        constraint entities_entities_location_entities
            references entities
            on delete set null;

create index if not exists idx_entities_location
    on entities (entity_location_entities);
