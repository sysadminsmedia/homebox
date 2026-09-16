-- +goose Up
ALTER TABLE "groups"
    ADD COLUMN "found_contact_enabled" boolean NOT NULL DEFAULT false;
ALTER TABLE "groups"
    ADD COLUMN "found_contact_message" character varying NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE "groups" DROP COLUMN "found_contact_message";
ALTER TABLE "groups" DROP COLUMN "found_contact_enabled";
