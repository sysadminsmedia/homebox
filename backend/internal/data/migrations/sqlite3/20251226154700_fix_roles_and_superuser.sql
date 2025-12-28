-- +goose Up
-- Force the role and superuser flags, previous nullable password migration (prior to v0.22.2)
-- caused them to flip-flop during migration for some users.
UPDATE users SET role = 'owner', is_superuser = 0, superuser = 0;
