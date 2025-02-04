-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_auth_roles" table
CREATE TABLE `new_auth_roles` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `role` text NOT NULL DEFAULT 'user', `auth_tokens_roles` uuid NULL, CONSTRAINT `auth_roles_auth_tokens_roles` FOREIGN KEY (`auth_tokens_roles`) REFERENCES `auth_tokens` (`id`) ON DELETE CASCADE);
-- copy rows from old table "auth_roles" to new temporary table "new_auth_roles"
INSERT INTO `new_auth_roles` (`id`, `role`, `auth_tokens_roles`) SELECT `id`, `role`, `auth_tokens_roles` FROM `auth_roles`;
-- drop "auth_roles" table after copying rows
DROP TABLE `auth_roles`;
-- rename temporary table "new_auth_roles" to "auth_roles"
ALTER TABLE `new_auth_roles` RENAME TO `auth_roles`;
-- create index "auth_roles_auth_tokens_roles_key" to table: "auth_roles"
CREATE UNIQUE INDEX `auth_roles_auth_tokens_roles_key` ON `auth_roles` (`auth_tokens_roles`);
-- delete where tokens is null
DELETE FROM `auth_roles` WHERE `auth_tokens_roles` IS NULL;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;
