-- +goose Up
-- Label templates for customizable label printing
CREATE TABLE IF NOT EXISTS "label_templates" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "name" character varying NOT NULL,
    "description" character varying NULL,
    "width" double precision NOT NULL DEFAULT 62.0,
    "height" double precision NOT NULL DEFAULT 29.0,
    "preset" character varying NULL,
    "is_shared" boolean NOT NULL DEFAULT false,
    "canvas_data" jsonb NULL,
    "output_format" character varying NOT NULL DEFAULT 'png',
    "dpi" bigint NOT NULL DEFAULT 300,
    "owner_id" uuid NOT NULL,
    "group_label_templates" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "label_templates_users_label_templates" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "label_templates_groups_label_templates" FOREIGN KEY ("group_label_templates") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX "idx_label_templates_name" ON "label_templates" ("name");
CREATE INDEX "idx_label_templates_is_shared" ON "label_templates" ("is_shared");
CREATE INDEX "idx_label_templates_preset" ON "label_templates" ("preset");
CREATE INDEX "idx_label_templates_owner" ON "label_templates" ("owner_id");
CREATE INDEX "idx_label_templates_group" ON "label_templates" ("group_label_templates");

-- +goose Down
DROP INDEX IF EXISTS "idx_label_templates_group";
DROP INDEX IF EXISTS "idx_label_templates_owner";
DROP INDEX IF EXISTS "idx_label_templates_preset";
DROP INDEX IF EXISTS "idx_label_templates_is_shared";
DROP INDEX IF EXISTS "idx_label_templates_name";
DROP TABLE IF EXISTS "label_templates";
