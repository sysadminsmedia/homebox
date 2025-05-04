-- +goose Up

-- Step 1: Modify "attachments" table to add new columns
ALTER TABLE "attachments" ADD COLUMN "title" character varying NOT NULL DEFAULT '', ADD COLUMN "path" character varying NOT NULL DEFAULT '';

-- Update existing rows in "attachments" with data from "documents"
UPDATE "attachments"
SET "title" = d."title",
    "path" = d."path"
FROM "documents" d
WHERE "attachments"."document_attachments" = d."id";

-- Step 3: Drop foreign key constraints referencing "documents"
ALTER TABLE "attachments" DROP CONSTRAINT IF EXISTS "attachments_documents_attachments";

-- Step 4: Drop the "document_attachments" column
ALTER TABLE "attachments" DROP COLUMN IF EXISTS "document_attachments";

-- Step 5: Drop the "documents" table
DROP TABLE IF EXISTS "documents";