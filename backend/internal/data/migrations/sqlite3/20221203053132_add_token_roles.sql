-- create "auth_roles" table
CREATE TABLE `auth_roles` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `role` text NOT NULL DEFAULT 'user', `auth_tokens_roles` uuid NULL, CONSTRAINT `auth_roles_auth_tokens_roles` FOREIGN KEY (`auth_tokens_roles`) REFERENCES `auth_tokens` (`id`) ON DELETE SET NULL);
-- create index "auth_roles_auth_tokens_roles_key" to table: "auth_roles"
CREATE UNIQUE INDEX `auth_roles_auth_tokens_roles_key` ON `auth_roles` (`auth_tokens_roles`);
