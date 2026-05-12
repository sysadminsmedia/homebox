-- +goose Up
-- Single-use password reset tokens for the forgot-password flow. Tokens are
-- stored as a sha256 hash; the raw value is sent only over email/CLI exactly
-- once. used_at is set when consumed so a replay finds neither a fresh row
-- nor an unmarked one.
create table if not exists password_reset_tokens
(
    id         uuid     not null
        primary key,
    created_at datetime not null,
    updated_at datetime not null,
    user_id    uuid     not null
        constraint password_reset_tokens_users_password_reset_tokens
            references users
            on delete cascade,
    token      blob     not null,
    expires_at datetime not null,
    used_at    datetime
);

create unique index if not exists password_reset_tokens_token_key
    on password_reset_tokens (token);

create index if not exists passwordresettokens_token
    on password_reset_tokens (token);

create index if not exists passwordresettokens_user_id
    on password_reset_tokens (user_id);

-- +goose Down
drop table if exists password_reset_tokens;
