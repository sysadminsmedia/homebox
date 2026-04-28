-- +goose Up
-- Create "exports" table
CREATE TABLE IF NOT EXISTS "exports" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "status" character varying NOT NULL DEFAULT 'pending',
    "progress" bigint NOT NULL DEFAULT 0,
    "artifact_path" character varying NULL,
    "size_bytes" bigint NOT NULL DEFAULT 0,
    "error" character varying NULL,
    "group_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "exports_groups_exports" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "export_group_id" to table: "exports"
CREATE INDEX IF NOT EXISTS "export_group_id" ON "exports" ("group_id");
-- Create index "export_group_id_status" to table: "exports"
CREATE INDEX IF NOT EXISTS "export_group_id_status" ON "exports" ("group_id", "status");
