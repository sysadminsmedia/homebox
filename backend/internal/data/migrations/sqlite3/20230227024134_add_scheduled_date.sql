-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_maintenance_entries" table
CREATE TABLE `new_maintenance_entries` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `date` datetime NULL, `scheduled_date` datetime NULL, `name` text NOT NULL, `description` text NULL, `cost` real NOT NULL DEFAULT 0, `item_id` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `maintenance_entries_items_maintenance_entries` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE);
-- copy rows from old table "maintenance_entries" to new temporary table "new_maintenance_entries"
INSERT INTO `new_maintenance_entries` (`id`, `created_at`, `updated_at`, `date`, `name`, `description`, `cost`, `item_id`) SELECT `id`, `created_at`, `updated_at`, `date`, `name`, `description`, `cost`, `item_id` FROM `maintenance_entries`;
-- drop "maintenance_entries" table after copying rows
DROP TABLE `maintenance_entries`;
-- rename temporary table "new_maintenance_entries" to "maintenance_entries"
ALTER TABLE `new_maintenance_entries` RENAME TO `maintenance_entries`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
