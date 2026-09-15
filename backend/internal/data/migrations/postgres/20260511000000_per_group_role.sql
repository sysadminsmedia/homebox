-- +goose Up
-- Move ownership from a global users.role enum to a per-membership role on
-- user_groups. Backfill: a user is owner of their default_group_id (the group
-- they registered with — fixed at registration and not user-mutable) and is a
-- regular member of every other group they belong to. The prior global
-- users.role column is removed; it was load-bearing for cross-group wipe and
-- caused authorization bugs because every self-registered user got 'owner'
-- regardless of the group being acted on.

ALTER TABLE "user_groups"
    ADD COLUMN "role" character varying NOT NULL DEFAULT 'user';

UPDATE "user_groups"
SET "role" = 'owner'
FROM "users"
WHERE "user_groups"."user_id" = "users"."id"
  AND "user_groups"."group_id" = "users"."default_group_id";

ALTER TABLE "users" DROP COLUMN "role";

-- +goose Down
ALTER TABLE "users"
    ADD COLUMN "role" character varying NOT NULL DEFAULT 'user';

UPDATE "users"
SET "role" = 'owner'
FROM "user_groups"
WHERE "user_groups"."user_id" = "users"."id"
  AND "user_groups"."group_id" = "users"."default_group_id"
  AND "user_groups"."role" = 'owner';

ALTER TABLE "user_groups" DROP COLUMN "role";
