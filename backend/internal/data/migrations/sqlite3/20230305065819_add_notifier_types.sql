-- create "notifiers" table
CREATE TABLE `notifiers` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `name` text NOT NULL, `url` text NOT NULL, `is_active` bool NOT NULL DEFAULT true, `user_id` uuid NOT NULL, PRIMARY KEY (`id`), CONSTRAINT `notifiers_users_notifiers` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE);
-- create index "notifier_user_id" to table: "notifiers"
CREATE INDEX `notifier_user_id` ON `notifiers` (`user_id`);
-- create index "notifier_user_id_is_active" to table: "notifiers"
CREATE INDEX `notifier_user_id_is_active` ON `notifiers` (`user_id`, `is_active`);
