-- +goose Up
-- Permission system: per-membership permission lists, tenant-scoped
-- permission groups (with user membership), and row-level access grants on
-- entities. Backfill grants every existing membership the full-access
-- wildcard "*" so upgrade behavior is unchanged and permissions added to
-- the catalog later automatically reach them; admins restrict afterwards.

-- 1. Direct permissions on tenant memberships.
ALTER TABLE user_groups ADD COLUMN permissions JSON NOT NULL DEFAULT '[]';

UPDATE user_groups SET permissions =
 '["*"]';

-- 2. Permissions applied by invitations on acceptance. Existing (and
--    unspecified future) invitations keep today's behavior: full access.
ALTER TABLE group_invitation_tokens ADD COLUMN permissions JSON NOT NULL DEFAULT
 '["*"]';

-- 3. Permission groups (tenant-scoped permission bundles).
CREATE TABLE permission_groups (
    id UUID NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    permissions JSON NOT NULL DEFAULT '[]',
    group_id UUID NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT permission_groups_groups_permission_groups FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX permissiongroup_name_group_id ON permission_groups(name, group_id);

-- 4. Permission group membership (M:M users <-> permission_groups).
CREATE TABLE permission_group_users (
    permission_group_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (permission_group_id, user_id),
    CONSTRAINT permission_group_users_permission_group_id FOREIGN KEY (permission_group_id) REFERENCES permission_groups(id) ON DELETE CASCADE,
    CONSTRAINT permission_group_users_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 5. Row-level access grants on entities. Exactly one of user_id /
--    permission_group_id is set. update/delete/attachments imply read
--    (normalized by an ent hook; can_read is authoritative for read checks).
CREATE TABLE access_grants (
    id UUID NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    can_read BOOLEAN NOT NULL DEFAULT false,
    can_update BOOLEAN NOT NULL DEFAULT false,
    can_delete BOOLEAN NOT NULL DEFAULT false,
    can_attachments BOOLEAN NOT NULL DEFAULT false,
    user_id UUID,
    permission_group_id UUID,
    entity_id UUID NOT NULL,
    group_id UUID NOT NULL,
    PRIMARY KEY (id),
    CHECK ((user_id IS NULL) <> (permission_group_id IS NULL)),
    CONSTRAINT access_grants_users_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT access_grants_permission_groups_permission_group FOREIGN KEY (permission_group_id) REFERENCES permission_groups(id) ON DELETE CASCADE,
    CONSTRAINT access_grants_entities_access_grants FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE,
    CONSTRAINT access_grants_groups_access_grants FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX accessgrant_entity_id_user_id ON access_grants(entity_id, user_id);
CREATE UNIQUE INDEX accessgrant_entity_id_permission_group_id ON access_grants(entity_id, permission_group_id);
CREATE INDEX accessgrant_user_id ON access_grants(user_id);
CREATE INDEX accessgrant_permission_group_id ON access_grants(permission_group_id);

-- +goose Down
DROP TABLE IF EXISTS access_grants;
DROP TABLE IF EXISTS permission_group_users;
DROP TABLE IF EXISTS permission_groups;
ALTER TABLE group_invitation_tokens DROP COLUMN permissions;
ALTER TABLE user_groups DROP COLUMN permissions;
