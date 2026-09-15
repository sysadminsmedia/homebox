-- +goose Up
-- Add static, user-scoped API keys. Each key authenticates as the owning
-- user and grants the same access. Tokens are stored hashed (sha256) and
-- never persisted in plaintext.
create table if not exists api_keys
(
    id           uuid     not null
        primary key,
    created_at   datetime not null,
    updated_at   datetime not null,
    user_id      uuid     not null
        constraint api_keys_users_api_keys
            references users
            on delete cascade,
    name         text     not null,
    token        blob     not null,
    expires_at   datetime,
    last_used_at datetime
);

create unique index if not exists api_keys_token_key
    on api_keys (token);

create index if not exists apikey_token
    on api_keys (token);

create index if not exists apikey_user_id
    on api_keys (user_id);

-- +goose Down
drop table if exists api_keys;
