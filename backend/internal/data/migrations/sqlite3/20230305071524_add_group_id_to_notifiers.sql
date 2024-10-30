-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_notifiers" table
CREATE TABLE `new_notifiers` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `url` text NOT NULL, `is_active` bool NOT NULL DEFAULT true, `group_id` uuid NOT NULL, `user_id` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `notifiers_groups_notifiers` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE, CONSTRAINT `notifiers_users_notifiers` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE);
-- copy rows from old table "notifiers" to new temporary table "new_notifiers"
INSERT INTO `new_notifiers` (`id`, `created_at`, `updated_at`, `name`, `url`, `is_active`, `user_id`) SELECT `id`, `created_at`, `updated_at`, `name`, `url`, `is_active`, `user_id` FROM `notifiers`;
-- drop "notifiers" table after copying rows
DROP TABLE `notifiers`;
-- rename temporary table "new_notifiers" to "notifiers"
ALTER TABLE `new_notifiers` RENAME TO `notifiers`;
-- create index "notifier_user_id" to table: "notifiers"
CREATE INDEX `notifier_user_id` ON `notifiers` (`user_id`);
-- create index "notifier_user_id_is_active" to table: "notifiers"
CREATE INDEX `notifier_user_id_is_active` ON `notifiers` (`user_id`, `is_active`);
-- create index "notifier_group_id" to table: "notifiers"
CREATE INDEX `notifier_group_id` ON `notifiers` (`group_id`);
-- create index "notifier_group_id_is_active" to table: "notifiers"
CREATE INDEX `notifier_group_id_is_active` ON `notifiers` (`group_id`, `is_active`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
