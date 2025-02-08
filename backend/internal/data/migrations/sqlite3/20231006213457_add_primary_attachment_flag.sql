-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_attachments" table
CREATE TABLE `new_attachments` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `type` text NOT NULL DEFAULT 'attachment', `primary` bool NOT NULL DEFAULT false, `document_attachments` uuid NOT NULL, `item_attachments` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `attachments_documents_attachments` FOREIGN KEY (`document_attachments`) REFERENCES `documents` (`id`) ON DELETE CASCADE, CONSTRAINT `attachments_items_attachments` FOREIGN KEY (`item_attachments`) REFERENCES `items` (`id`) ON DELETE CASCADE);
-- Copy rows from old table "attachments" to new temporary table "new_attachments"
INSERT INTO `new_attachments` (`id`, `created_at`, `updated_at`, `type`, `document_attachments`, `item_attachments`) SELECT `id`, `created_at`, `updated_at`, `type`, `document_attachments`, `item_attachments` FROM `attachments`;
-- Drop "attachments" table after copying rows
DROP TABLE `attachments`;
-- Rename temporary table "new_attachments" to "attachments"
ALTER TABLE `new_attachments` RENAME TO `attachments`;
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
