-- +goose Up
-- Printers for direct label printing
CREATE TABLE IF NOT EXISTS "printers" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "name" character varying NOT NULL,
    "description" character varying NULL,
    "printer_type" character varying NOT NULL DEFAULT 'ipp',
    "address" character varying NOT NULL,
    "is_default" boolean NOT NULL DEFAULT false,
    "label_width_mm" double precision NULL,
    "label_height_mm" double precision NULL,
    "dpi" bigint NOT NULL DEFAULT 300,
    "media_type" character varying NULL,
    "status" character varying NOT NULL DEFAULT 'unknown',
    "last_status_check" timestamptz NULL,
    "group_printers" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "printers_groups_printers" FOREIGN KEY ("group_printers") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX "idx_printers_name" ON "printers" ("name");
CREATE INDEX "idx_printers_is_default" ON "printers" ("is_default");
CREATE INDEX "idx_printers_printer_type" ON "printers" ("printer_type");
CREATE INDEX "idx_printers_group" ON "printers" ("group_printers");

-- +goose Down
DROP INDEX IF EXISTS "idx_printers_group";
DROP INDEX IF EXISTS "idx_printers_printer_type";
DROP INDEX IF EXISTS "idx_printers_is_default";
DROP INDEX IF EXISTS "idx_printers_name";
DROP TABLE IF EXISTS "printers";
