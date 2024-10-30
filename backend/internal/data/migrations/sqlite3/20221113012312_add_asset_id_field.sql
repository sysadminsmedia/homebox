-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_items" table
CREATE TABLE `new_items` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `description` text NULL, `import_ref` text NULL, `notes` text NULL, `quantity` integer NOT NULL DEFAULT 1, `insured` bool NOT NULL DEFAULT false, `archived` bool NOT NULL DEFAULT false, `asset_id` integer NOT NULL DEFAULT 0, `serial_number` text NULL, `model_number` text NULL, `manufacturer` text NULL, `lifetime_warranty` bool NOT NULL DEFAULT false, `warranty_expires` datetime NULL, `warranty_details` text NULL, `purchase_time` datetime NULL, `purchase_from` text NULL, `purchase_price` real NOT NULL DEFAULT 0, `sold_time` datetime NULL, `sold_to` text NULL, `sold_price` real NOT NULL DEFAULT 0, `sold_notes` text NULL, `group_items` uuid NOT NULL, `item_children` uuid NULL, `location_items` uuid NULL, PRIMARY KEY (`id`), CONSTRAINT `items_groups_items` FOREIGN KEY (`group_items`) REFERENCES `groups` (`id`) ON DELETE CASCADE, CONSTRAINT `items_items_children` FOREIGN KEY (`item_children`) REFERENCES `items` (`id`) ON DELETE SET NULL, CONSTRAINT `items_locations_items` FOREIGN KEY (`location_items`) REFERENCES `locations` (`id`) ON DELETE CASCADE);
-- copy rows from old table "items" to new temporary table "new_items"
INSERT INTO `new_items` (`id`, `created_at`, `updated_at`, `name`, `description`, `import_ref`, `notes`, `quantity`, `insured`, `archived`, `serial_number`, `model_number`, `manufacturer`, `lifetime_warranty`, `warranty_expires`, `warranty_details`, `purchase_time`, `purchase_from`, `purchase_price`, `sold_time`, `sold_to`, `sold_price`, `sold_notes`, `group_items`, `item_children`, `location_items`) SELECT `id`, `created_at`, `updated_at`, `name`, `description`, `import_ref`, `notes`, `quantity`, `insured`, `archived`, `serial_number`, `model_number`, `manufacturer`, `lifetime_warranty`, `warranty_expires`, `warranty_details`, `purchase_time`, `purchase_from`, `purchase_price`, `sold_time`, `sold_to`, `sold_price`, `sold_notes`, `group_items`, `item_children`, `location_items` FROM `items`;
-- drop "items" table after copying rows
DROP TABLE `items`;
-- rename temporary table "new_items" to "items"
ALTER TABLE `new_items` RENAME TO `items`;
-- create index "item_name" to table: "items"
CREATE INDEX `item_name` ON `items` (`name`);
-- create index "item_manufacturer" to table: "items"
CREATE INDEX `item_manufacturer` ON `items` (`manufacturer`);
-- create index "item_model_number" to table: "items"
CREATE INDEX `item_model_number` ON `items` (`model_number`);
-- create index "item_serial_number" to table: "items"
CREATE INDEX `item_serial_number` ON `items` (`serial_number`);
-- create index "item_archived" to table: "items"
CREATE INDEX `item_archived` ON `items` (`archived`);
-- create index "item_asset_id" to table: "items"
CREATE INDEX `item_asset_id` ON `items` (`asset_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
