-- create "attachments" table
CREATE TABLE `attachments` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `type` text NOT NULL DEFAULT 'attachment', `document_attachments` uuid NOT NULL, `item_attachments` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `attachments_documents_attachments` FOREIGN KEY (`document_attachments`) REFERENCES `documents` (`id`) ON DELETE CASCADE, CONSTRAINT `attachments_items_attachments` FOREIGN KEY (`item_attachments`) REFERENCES `items` (`id`) ON DELETE CASCADE);
-- create "auth_tokens" table
CREATE TABLE `auth_tokens` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `token` blob NOT NULL, `expires_at` datetime NOT NULL, `user_auth_tokens` uuid NULL, PRIMARY KEY (`id`), CONSTRAINT `auth_tokens_users_auth_tokens` FOREIGN KEY (`user_auth_tokens`) REFERENCES `users` (`id`) ON DELETE CASCADE);
-- create index "auth_tokens_token_key" to table: "auth_tokens"
CREATE UNIQUE INDEX `auth_tokens_token_key` ON `auth_tokens` (`token`);
-- create index "authtokens_token" to table: "auth_tokens"
CREATE INDEX `authtokens_token` ON `auth_tokens` (`token`);
-- create "documents" table
CREATE TABLE `documents` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `title` text NOT NULL, `path` text NOT NULL, `group_documents` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `documents_groups_documents` FOREIGN KEY (`group_documents`) REFERENCES `groups` (`id`) ON DELETE CASCADE);
-- create "document_tokens" table
CREATE TABLE `document_tokens` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `token` blob NOT NULL, `uses` integer NOT NULL DEFAULT 1, `expires_at` datetime NOT NULL, `document_document_tokens` uuid NULL, PRIMARY KEY (`id`), CONSTRAINT `document_tokens_documents_document_tokens` FOREIGN KEY (`document_document_tokens`) REFERENCES `documents` (`id`) ON DELETE CASCADE);
-- create index "document_tokens_token_key" to table: "document_tokens"
CREATE UNIQUE INDEX `document_tokens_token_key` ON `document_tokens` (`token`);
-- create index "documenttoken_token" to table: "document_tokens"
CREATE INDEX `documenttoken_token` ON `document_tokens` (`token`);
-- create "groups" table
CREATE TABLE `groups` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `currency` text NOT NULL DEFAULT 'usd', PRIMARY KEY (`id`));
-- create "items" table
CREATE TABLE `items` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `import_ref` text NULL, `notes` text NULL, `quantity` integer NOT NULL DEFAULT 1, `insured` bool NOT NULL DEFAULT false, `serial_number` text NULL, `model_number` text NULL, `manufacturer` text NULL, `lifetime_warranty` bool NOT NULL DEFAULT false, `warranty_expires` datetime NULL, `warranty_details` text NULL, `purchase_time` datetime NULL, `purchase_from` text NULL, `purchase_price` real NOT NULL DEFAULT 0, `sold_time` datetime NULL, `sold_to` text NULL, `sold_price` real NOT NULL DEFAULT 0, `sold_notes` text NULL, `group_items` uuid NOT NULL, `location_items` uuid NULL, PRIMARY KEY (`id`), CONSTRAINT `items_groups_items` FOREIGN KEY (`group_items`) REFERENCES `groups` (`id`) ON DELETE CASCADE, CONSTRAINT `items_locations_items` FOREIGN KEY (`location_items`) REFERENCES `locations` (`id`) ON DELETE CASCADE);
-- create index "item_name" to table: "items"
CREATE INDEX `item_name` ON `items` (`name`);
-- create index "item_manufacturer" to table: "items"
CREATE INDEX `item_manufacturer` ON `items` (`manufacturer`);
-- create index "item_model_number" to table: "items"
CREATE INDEX `item_model_number` ON `items` (`model_number`);
-- create index "item_serial_number" to table: "items"
CREATE INDEX `item_serial_number` ON `items` (`serial_number`);
-- create "item_fields" table
CREATE TABLE `item_fields` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `type` text NOT NULL, `text_value` text NULL, `number_value` integer NULL, `boolean_value` bool NOT NULL DEFAULT false, `time_value` datetime NOT NULL, `item_fields` uuid NULL, PRIMARY KEY (`id`), CONSTRAINT `item_fields_items_fields` FOREIGN KEY (`item_fields`) REFERENCES `items` (`id`) ON DELETE CASCADE);
-- create "labels" table
CREATE TABLE `labels` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `color` text NULL, `group_labels` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `labels_groups_labels` FOREIGN KEY (`group_labels`) REFERENCES `groups` (`id`) ON DELETE CASCADE);
-- create "locations" table
CREATE TABLE `locations` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `group_locations` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `locations_groups_locations` FOREIGN KEY (`group_locations`) REFERENCES `groups` (`id`) ON DELETE CASCADE);
-- create "users" table
CREATE TABLE `users` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `email` text NOT NULL, `password` text NOT NULL, `is_superuser` bool NOT NULL DEFAULT false, `group_users` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `users_groups_users` FOREIGN KEY (`group_users`) REFERENCES `groups` (`id`) ON DELETE CASCADE);
-- create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX `users_email_key` ON `users` (`email`);
-- create "label_items" table
CREATE TABLE `label_items` (`label_id` uuid NOT NULL, `item_id` uuid NOT NULL, PRIMARY KEY (`label_id`, `item_id`), CONSTRAINT `label_items_label_id` FOREIGN KEY (`label_id`) REFERENCES `labels` (`id`) ON DELETE CASCADE, CONSTRAINT `label_items_item_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE);
