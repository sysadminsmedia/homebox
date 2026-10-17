-- +goose Up
-- Permission system: per-membership permission lists, tenant-scoped
-- permission groups (with user membership), and row-level access grants on
-- entities. The backfill preserves pre-upgrade behavior: owners get the
-- full-access wildcard "*" (so permissions added to the catalog later reach
-- them), while other members get an explicit list matching what they could
-- already do. Collection administration (settings, members, permissions,
-- deleting the collection, wiping inventory) was owner-only before this
-- migration and must stay that way, so it is excluded from the member list.

-- 1. Direct permissions on tenant memberships.
ALTER TABLE "user_groups"
    ADD COLUMN "permissions" jsonb NOT NULL DEFAULT '[]'::jsonb;

UPDATE "user_groups" SET "permissions" = CASE
    WHEN "role" = 'owner' THEN '["*"]'::jsonb
    ELSE '["entity:*","tag:manage","entitytype:manage","template:manage","notifier:manage","data:export","data:import","statistics:read","tools:run","label:print","report:generate","barcode:lookup"]'::jsonb
END;

-- 2. Permissions applied by invitations on acceptance. The default is the
--    ordinary member list (authz.DefaultMember, matching the ent schema), so
--    existing invitations — which created non-owner memberships before this
--    migration — and any invite created without permissions never confer
--    collection administration.
ALTER TABLE "group_invitation_tokens"
    ADD COLUMN "permissions" jsonb NOT NULL DEFAULT
 '["entity:*","tag:manage","entitytype:manage","template:manage","notifier:manage","data:export","data:import","statistics:read","tools:run","label:print","report:generate","barcode:lookup"]'::jsonb;

-- 3. Permission groups (tenant-scoped permission bundles).
CREATE TABLE IF NOT EXISTS "permission_groups" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "name" character varying NOT NULL,
    "description" character varying NULL
        CHECK ("description" IS NULL OR char_length("description") <= 1000),
    "permissions" jsonb NOT NULL DEFAULT '[]'::jsonb,
    "group_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "permission_groups_groups_permission_groups" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- Create index "permissiongroup_name_group_id" to table: "permission_groups"
CREATE UNIQUE INDEX IF NOT EXISTS "permissiongroup_name_group_id" ON "permission_groups" ("name", "group_id");

-- 4. Permission group membership (M:M users <-> permission_groups).
CREATE TABLE IF NOT EXISTS "permission_group_users" (
    "permission_group_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    PRIMARY KEY ("permission_group_id", "user_id"),
    CONSTRAINT "permission_group_users_permission_group_id" FOREIGN KEY ("permission_group_id") REFERENCES "permission_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "permission_group_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- 5. Row-level access grants on entities. Exactly one of user_id /
--    permission_group_id is set. update/delete/attachments imply read
--    (normalized by an ent hook; can_read is authoritative for read checks).
CREATE TABLE IF NOT EXISTS "access_grants" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "can_read" boolean NOT NULL DEFAULT false,
    "can_update" boolean NOT NULL DEFAULT false,
    "can_delete" boolean NOT NULL DEFAULT false,
    "can_attachments" boolean NOT NULL DEFAULT false,
    "user_id" uuid NULL,
    "permission_group_id" uuid NULL,
    "entity_id" uuid NOT NULL,
    "group_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CHECK (("user_id" IS NULL) <> ("permission_group_id" IS NULL)),
    CONSTRAINT "access_grants_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "access_grants_permission_groups_permission_group" FOREIGN KEY ("permission_group_id") REFERENCES "permission_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "access_grants_entities_access_grants" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "access_grants_groups_access_grants" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- Create indexes to table: "access_grants"
CREATE UNIQUE INDEX IF NOT EXISTS "accessgrant_entity_id_user_id" ON "access_grants" ("entity_id", "user_id");
CREATE UNIQUE INDEX IF NOT EXISTS "accessgrant_entity_id_permission_group_id" ON "access_grants" ("entity_id", "permission_group_id");
CREATE INDEX IF NOT EXISTS "accessgrant_user_id" ON "access_grants" ("user_id");
CREATE INDEX IF NOT EXISTS "accessgrant_permission_group_id" ON "access_grants" ("permission_group_id");

-- +goose Down
DROP TABLE IF EXISTS "access_grants";
DROP TABLE IF EXISTS "permission_group_users";
DROP TABLE IF EXISTS "permission_groups";
ALTER TABLE "group_invitation_tokens" DROP COLUMN "permissions";
ALTER TABLE "user_groups" DROP COLUMN "permissions";
