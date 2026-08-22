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
ALTER TABLE "entities"
    ADD COLUMN IF NOT EXISTS "entity_location_entities" uuid NULL;

-- Separate and guarded so a database that already has the column (re-run, or a
-- partially migrated dump) doesn't fail here. The StatementBegin/End fence is
-- required — goose splits on ";" and would cut the block at the first one.
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_schema = current_schema()
          AND table_name = 'entities'
          AND constraint_name = 'entities_entities_location_entities'
    ) THEN
        ALTER TABLE "entities"
            ADD CONSTRAINT "entities_entities_location_entities"
            FOREIGN KEY ("entity_location_entities") REFERENCES "entities" ("id")
            ON UPDATE NO ACTION ON DELETE SET NULL;
    END IF;
END $$;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS "idx_entities_location" ON "entities" ("entity_location_entities");
