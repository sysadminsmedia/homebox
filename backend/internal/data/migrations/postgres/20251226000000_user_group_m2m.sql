-- +goose Up
-- Create user_groups junction table for M:M relationship
CREATE TABLE IF NOT EXISTS "user_groups" (
    "user_id" uuid NOT NULL,
    "group_id" uuid NOT NULL,
    PRIMARY KEY ("user_id", "group_id"),
    CONSTRAINT "user_groups_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT "user_groups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- Migrate existing user->group relationships to the junction table
INSERT INTO "user_groups" ("user_id", "group_id")
SELECT "id", "group_users" FROM "users" WHERE "group_users" IS NOT NULL;

-- Add default_group_id column to users table
ALTER TABLE "users" ADD COLUMN "default_group_id" uuid;

-- Set default_group_id to the user's current group
UPDATE "users" SET "default_group_id" = "group_users" WHERE "group_users" IS NOT NULL;

-- Drop the old group_users foreign key constraint and column
ALTER TABLE "users" DROP CONSTRAINT "users_groups_users";
ALTER TABLE "users" DROP COLUMN "group_users";

-- Add foreign key constraint for default_group_id
ALTER TABLE "users" ADD CONSTRAINT "users_groups_users_default" FOREIGN KEY ("default_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- Recreate group_users column with foreign key
ALTER TABLE "users" ADD COLUMN "group_users" uuid;

-- Restore the group_users values from user_groups (using the default_group_id or first entry)
UPDATE "users"
SET "group_users" = COALESCE("default_group_id", (
    SELECT "group_id" FROM "user_groups" WHERE "user_id" = "users"."id" LIMIT 1
));

-- Drop the default_group_id foreign key and column
ALTER TABLE "users" DROP CONSTRAINT "users_groups_users_default";
ALTER TABLE "users" DROP COLUMN "default_group_id";

-- Add back the original foreign key constraint
ALTER TABLE "users" ADD CONSTRAINT "users_groups_users" FOREIGN KEY ("group_users") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- Drop the junction table
DROP TABLE IF EXISTS "user_groups";

