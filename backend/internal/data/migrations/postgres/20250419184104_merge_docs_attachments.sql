-- Modify "attachments" table
ALTER TABLE "attachments" ADD COLUMN "title" character varying NOT NULL DEFAULT '', ADD COLUMN "path" character varying NOT NULL DEFAULT '';

-- Step 2: Migrate data from "documents" to "attachments"
UPDATE "attachments"
SET "title" = d."title",
    "path" = d."path"
FROM "documents" d
WHERE "attachments"."document_attachments" = d."id";

-- Step 3: Drop foreign key constraints referencing "documents"
ALTER TABLE "attachments" DROP CONSTRAINT "attachments_documents_attachments";
ALTER TABLE "attachments" DROP COLUMN "document_attachments";

-- Step 4: Drop the "documents" table
DROP TABLE "documents";