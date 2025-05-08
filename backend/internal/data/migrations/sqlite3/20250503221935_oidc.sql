-- +goose Up
-- +goose no transaction
-- Disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- Create "new_users" table
CREATE TABLE `new_users`
(
    `id`           uuid     NOT NULL,
    `created_at`   datetime NOT NULL,
    `updated_at`   datetime NOT NULL,
    `name`         text     NOT NULL,
    `email`        text     NOT NULL,
    `password`     text     NULL,
    `is_superuser` bool     NOT NULL DEFAULT (false),
    `superuser`    bool     NOT NULL DEFAULT (false),
    `role`         text     NOT NULL DEFAULT ('user'),
    `activated_on` datetime NULL,
    `group_users`  uuid     NOT NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `users_groups_users` FOREIGN KEY (`group_users`) REFERENCES `groups` (`id`) ON DELETE CASCADE
);
-- Copy rows from old table "users" to new temporary table "new_users"
INSERT INTO `new_users` (`id`, `created_at`, `updated_at`, `name`, `email`, `password`, `is_superuser`, `superuser`,
                         `role`, `activated_on`, `group_users`)
SELECT `id`,
       `created_at`,
       `updated_at`,
       `name`,
       `email`,
       `password`,
       `is_superuser`,
       `superuser`,
       `role`,
       `activated_on`,
       `group_users`
FROM `users`;
-- Drop "users" table after copying rows
DROP TABLE `users`;
-- Rename temporary table "new_users" to "users"
ALTER TABLE `new_users`
    RENAME TO `users`;
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX `users_email_key` ON `users` (`email`);
-- Create "oauths" table
CREATE TABLE `oauths`
(
    `id`         uuid     NOT NULL,
    `created_at` datetime NOT NULL,
    `updated_at` datetime NOT NULL,
    `provider`   text     NOT NULL,
    `sub`        text     NOT NULL,
    `user_oauth` uuid     NULL,
    PRIMARY KEY (`id`),
    CONSTRAINT `oauths_users_oauth` FOREIGN KEY (`user_oauth`) REFERENCES `users` (`id`) ON DELETE CASCADE
);
-- Create index "oauth_provider_sub" to table: "oauths"
CREATE INDEX `oauth_provider_sub` ON `oauths` (`provider`, `sub`);
-- Enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
