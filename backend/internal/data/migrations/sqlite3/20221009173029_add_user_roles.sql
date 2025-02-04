-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_users" table
CREATE TABLE `new_users` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `email` text NOT NULL, `password` text NOT NULL, `is_superuser` bool NOT NULL DEFAULT false, `role` text NOT NULL DEFAULT 'user', `superuser` bool NOT NULL DEFAULT false, `activated_on` datetime NULL, `group_users` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `users_groups_users` FOREIGN KEY (`group_users`) REFERENCES `groups` (`id`) ON DELETE CASCADE);
-- copy rows from old table "users" to new temporary table "new_users"
INSERT INTO `new_users` (`id`, `created_at`, `updated_at`, `name`, `email`, `password`, `is_superuser`, `group_users`) SELECT `id`, `created_at`, `updated_at`, `name`, `email`, `password`, `is_superuser`, `group_users` FROM `users`;
-- drop "users" table after copying rows
DROP TABLE `users`;
-- rename temporary table "new_users" to "users"
ALTER TABLE `new_users` RENAME TO `users`;
-- create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX `users_email_key` ON `users` (`email`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
