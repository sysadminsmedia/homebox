-- create "group_invitation_tokens" table
CREATE TABLE `group_invitation_tokens` (`id` uuid NOT NULL, `created_at` datetime NOT NULL, `updated_at` datetime NOT NULL, `token` blob NOT NULL, `expires_at` datetime NOT NULL, `uses` integer NOT NULL DEFAULT 0, `group_invitation_tokens` uuid NULL, PRIMARY KEY (`id`), CONSTRAINT `group_invitation_tokens_groups_invitation_tokens` FOREIGN KEY (`group_invitation_tokens`) REFERENCES `groups` (`id`) ON DELETE CASCADE);
-- create index "group_invitation_tokens_token_key" to table: "group_invitation_tokens"
CREATE UNIQUE INDEX `group_invitation_tokens_token_key` ON `group_invitation_tokens` (`token`);
