-- +goose Up
-- Create "groups" table
CREATE TABLE IF NOT EXISTS "groups" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "currency" character varying NOT NULL DEFAULT 'usd', PRIMARY KEY ("id"));
-- Create "documents" table
CREATE TABLE IF NOT EXISTS "documents" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "title" character varying NOT NULL, "path" character varying NOT NULL, "group_documents" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "documents_groups_documents" FOREIGN KEY ("group_documents") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "locations" table
CREATE TABLE IF NOT EXISTS "locations" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "group_locations" uuid NOT NULL, "location_children" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "locations_groups_locations" FOREIGN KEY ("group_locations") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "locations_locations_children" FOREIGN KEY ("location_children") REFERENCES "locations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create "items" table
CREATE TABLE IF NOT EXISTS "items" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "import_ref" character varying NULL, "notes" character varying NULL, "quantity" bigint NOT NULL DEFAULT 1, "insured" boolean NOT NULL DEFAULT false, "archived" boolean NOT NULL DEFAULT false, "asset_id" bigint NOT NULL DEFAULT 0, "serial_number" character varying NULL, "model_number" character varying NULL, "manufacturer" character varying NULL, "lifetime_warranty" boolean NOT NULL DEFAULT false, "warranty_expires" timestamptz NULL, "warranty_details" character varying NULL, "purchase_time" timestamptz NULL, "purchase_from" character varying NULL, "purchase_price" double precision NOT NULL DEFAULT 0, "sold_time" timestamptz NULL, "sold_to" character varying NULL, "sold_price" double precision NOT NULL DEFAULT 0, "sold_notes" character varying NULL, "group_items" uuid NOT NULL, "item_children" uuid NULL, "location_items" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "items_groups_items" FOREIGN KEY ("group_items") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "items_items_children" FOREIGN KEY ("item_children") REFERENCES "items" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "items_locations_items" FOREIGN KEY ("location_items") REFERENCES "locations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "item_archived" to table: "items"
CREATE INDEX IF NOT EXISTS "item_archived" ON "items" ("archived");
-- Create index "item_asset_id" to table: "items"
CREATE INDEX IF NOT EXISTS "item_asset_id" ON "items" ("asset_id");
-- Create index "item_manufacturer" to table: "items"
CREATE INDEX IF NOT EXISTS "item_manufacturer" ON "items" ("manufacturer");
-- Create index "item_model_number" to table: "items"
CREATE INDEX IF NOT EXISTS "item_model_number" ON "items" ("model_number");
-- Create index "item_name" to table: "items"
CREATE INDEX IF NOT EXISTS "item_name" ON "items" ("name");
-- Create index "item_serial_number" to table: "items"
CREATE INDEX IF NOT EXISTS "item_serial_number" ON "items" ("serial_number");
-- Create "attachments" table
CREATE TABLE IF NOT EXISTS "attachments" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "type" character varying NOT NULL DEFAULT 'attachment', "primary" boolean NOT NULL DEFAULT false, "document_attachments" uuid NOT NULL, "item_attachments" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "attachments_documents_attachments" FOREIGN KEY ("document_attachments") REFERENCES "documents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "attachments_items_attachments" FOREIGN KEY ("item_attachments") REFERENCES "items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "users" table
CREATE TABLE IF NOT EXISTS "users" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "email" character varying NOT NULL, "password" character varying NOT NULL, "is_superuser" boolean NOT NULL DEFAULT false, "superuser" boolean NOT NULL DEFAULT false, "role" character varying NOT NULL DEFAULT 'user', "activated_on" timestamptz NULL, "group_users" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "users_groups_users" FOREIGN KEY ("group_users") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX IF NOT EXISTS "users_email_key" ON "users" ("email");
-- Create "auth_tokens" table
CREATE TABLE IF NOT EXISTS "auth_tokens" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "token" bytea NOT NULL, "expires_at" timestamptz NOT NULL, "user_auth_tokens" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "auth_tokens_users_auth_tokens" FOREIGN KEY ("user_auth_tokens") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "auth_tokens_token_key" to table: "auth_tokens"
CREATE UNIQUE INDEX IF NOT EXISTS "auth_tokens_token_key" ON "auth_tokens" ("token");
-- Create "auth_roles" table
CREATE TABLE IF NOT EXISTS "auth_roles" ("id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY, "role" character varying NOT NULL DEFAULT 'user', "auth_tokens_roles" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "auth_roles_auth_tokens_roles" FOREIGN KEY ("auth_tokens_roles") REFERENCES "auth_tokens" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "auth_roles_auth_tokens_roles_key" to table: "auth_roles"
CREATE UNIQUE INDEX IF NOT EXISTS "auth_roles_auth_tokens_roles_key" ON "auth_roles" ("auth_tokens_roles");
-- Create "group_invitation_tokens" table
CREATE TABLE IF NOT EXISTS "group_invitation_tokens" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "token" bytea NOT NULL, "expires_at" timestamptz NOT NULL, "uses" bigint NOT NULL DEFAULT 0, "group_invitation_tokens" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "group_invitation_tokens_groups_invitation_tokens" FOREIGN KEY ("group_invitation_tokens") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "group_invitation_tokens_token_key" to table: "group_invitation_tokens"
CREATE UNIQUE INDEX IF NOT EXISTS "group_invitation_tokens_token_key" ON "group_invitation_tokens" ("token");
-- Create "item_fields" table
CREATE TABLE IF NOT EXISTS "item_fields" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "type" character varying NOT NULL, "text_value" character varying NULL, "number_value" bigint NULL, "boolean_value" boolean NOT NULL DEFAULT false, "time_value" timestamptz NOT NULL, "item_fields" uuid NULL, PRIMARY KEY ("id"), CONSTRAINT "item_fields_items_fields" FOREIGN KEY ("item_fields") REFERENCES "items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "labels" table
CREATE TABLE IF NOT EXISTS "labels" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "color" character varying NULL, "group_labels" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "labels_groups_labels" FOREIGN KEY ("group_labels") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "label_items" table
CREATE TABLE IF NOT EXISTS "label_items" ("label_id" uuid NOT NULL, "item_id" uuid NOT NULL, PRIMARY KEY ("label_id", "item_id"), CONSTRAINT "label_items_item_id" FOREIGN KEY ("item_id") REFERENCES "items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "label_items_label_id" FOREIGN KEY ("label_id") REFERENCES "labels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "maintenance_entries" table
CREATE TABLE IF NOT EXISTS "maintenance_entries" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "date" timestamptz NULL, "scheduled_date" timestamptz NULL, "name" character varying NOT NULL, "description" character varying NULL, "cost" double precision NOT NULL DEFAULT 0, "item_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "maintenance_entries_items_maintenance_entries" FOREIGN KEY ("item_id") REFERENCES "items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "notifiers" table
CREATE TABLE IF NOT EXISTS "notifiers" ("id" uuid NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, "name" character varying NOT NULL, "url" character varying NOT NULL, "is_active" boolean NOT NULL DEFAULT true, "group_id" uuid NOT NULL, "user_id" uuid NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "notifiers_groups_notifiers" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "notifiers_users_notifiers" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "notifier_group_id" to table: "notifiers"
CREATE INDEX IF NOT EXISTS "notifier_group_id" ON "notifiers" ("group_id");
-- Create index "notifier_group_id_is_active" to table: "notifiers"
CREATE INDEX IF NOT EXISTS "notifier_group_id_is_active" ON "notifiers" ("group_id", "is_active");
-- Create index "notifier_user_id" to table: "notifiers"
CREATE INDEX IF NOT EXISTS "notifier_user_id" ON "notifiers" ("user_id");
-- Create index "notifier_user_id_is_active" to table: "notifiers"
CREATE INDEX IF NOT EXISTS "notifier_user_id_is_active" ON "notifiers" ("user_id", "is_active");
