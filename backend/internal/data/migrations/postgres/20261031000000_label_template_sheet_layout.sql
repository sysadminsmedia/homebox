-- +goose Up
-- Editable sheet layout (null follows the preset) and printer calibration offsets
ALTER TABLE "label_templates" ADD COLUMN "sheet_layout" jsonb NULL;
ALTER TABLE "label_templates" ADD COLUMN "offset_x" double precision NOT NULL DEFAULT 0;
ALTER TABLE "label_templates" ADD COLUMN "offset_y" double precision NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE "label_templates" DROP COLUMN IF EXISTS "offset_y";
ALTER TABLE "label_templates" DROP COLUMN IF EXISTS "offset_x";
ALTER TABLE "label_templates" DROP COLUMN IF EXISTS "sheet_layout";
