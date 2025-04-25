-- +goose Up
create table if not exists groups
(
    id         uuid               not null
        primary key,
    created_at datetime           not null,
    updated_at datetime           not null,
    name       text               not null,
    currency   text default 'usd' not null
);

create table if not exists documents
(
    id              uuid     not null
        primary key,
    created_at      datetime not null,
    updated_at      datetime not null,
    title           text     not null,
    path            text     not null,
    group_documents uuid     not null
        constraint documents_groups_documents
            references groups
            on delete cascade
);

create table if not exists group_invitation_tokens
(
    id                      uuid              not null
        primary key,
    created_at              datetime          not null,
    updated_at              datetime          not null,
    token                   blob              not null,
    expires_at              datetime          not null,
    uses                    integer default 0 not null,
    group_invitation_tokens uuid
        constraint group_invitation_tokens_groups_invitation_tokens
            references groups
            on delete cascade
);

create unique index if not exists group_invitation_tokens_token_key
    on group_invitation_tokens (token);

create table if not exists labels
(
    id           uuid     not null
        primary key,
    created_at   datetime not null,
    updated_at   datetime not null,
    name         text     not null,
    description  text,
    color        text,
    group_labels uuid     not null
        constraint labels_groups_labels
            references groups
            on delete cascade
);

create table if not exists locations
(
    id                uuid     not null
        primary key,
    created_at        datetime not null,
    updated_at        datetime not null,
    name              text     not null,
    description       text,
    group_locations   uuid     not null
        constraint locations_groups_locations
            references groups
            on delete cascade,
    location_children uuid
        constraint locations_locations_children
            references locations
            on delete set null
);

create table if not exists items
(
    id                uuid                  not null
        primary key,
    created_at        datetime              not null,
    updated_at        datetime              not null,
    name              text                  not null,
    description       text,
    import_ref        text,
    notes             text,
    quantity          integer default 1     not null,
    insured           bool    default false not null,
    archived          bool    default false not null,
    asset_id          integer default 0     not null,
    serial_number     text,
    model_number      text,
    manufacturer      text,
    lifetime_warranty bool    default false not null,
    warranty_expires  datetime,
    warranty_details  text,
    purchase_time     datetime,
    purchase_from     text,
    purchase_price    real    default 0     not null,
    sold_time         datetime,
    sold_to           text,
    sold_price        real    default 0     not null,
    sold_notes        text,
    group_items       uuid                  not null
        constraint items_groups_items
            references groups
            on delete cascade,
    item_children     uuid
        constraint items_items_children
            references items
            on delete set null,
    location_items    uuid
        constraint items_locations_items
            references locations
            on delete cascade
);

create table if not exists attachments
(
    id                   uuid                      not null
        primary key,
    created_at           datetime                  not null,
    updated_at           datetime                  not null,
    type                 text default 'attachment' not null,
    "primary"            bool default false        not null,
    document_attachments uuid                      not null
        constraint attachments_documents_attachments
            references documents
            on delete cascade,
    item_attachments     uuid                      not null
        constraint attachments_items_attachments
            references items
            on delete cascade
);

create table if not exists item_fields
(
    id            uuid               not null
        primary key,
    created_at    datetime           not null,
    updated_at    datetime           not null,
    name          text               not null,
    description   text,
    type          text               not null,
    text_value    text,
    number_value  integer,
    boolean_value bool default false not null,
    time_value    datetime           not null,
    item_fields   uuid
        constraint item_fields_items_fields
            references items
            on delete cascade
);

create index if not exists item_archived
    on items (archived);

create index if not exists item_asset_id
    on items (asset_id);

create index if not exists item_manufacturer
    on items (manufacturer);

create index if not exists item_model_number
    on items (model_number);

create index if not exists item_name
    on items (name);

create index if not exists item_serial_number
    on items (serial_number);

create table if not exists label_items
(
    label_id uuid not null
        constraint label_items_label_id
            references labels
            on delete cascade,
    item_id  uuid not null
        constraint label_items_item_id
            references items
            on delete cascade,
    primary key (label_id, item_id)
);

create table if not exists maintenance_entries
(
    id             uuid           not null
        primary key,
    created_at     datetime       not null,
    updated_at     datetime       not null,
    date           datetime,
    scheduled_date datetime,
    name           text           not null,
    description    text,
    cost           real default 0 not null,
    item_id        uuid           not null
        constraint maintenance_entries_items_maintenance_entries
            references items
            on delete cascade
);

create table if not exists users
(
    id           uuid                not null
        primary key,
    created_at   datetime            not null,
    updated_at   datetime            not null,
    name         text                not null,
    email        text                not null,
    password     text                not null,
    is_superuser bool default false  not null,
    superuser    bool default false  not null,
    role         text default 'user' not null,
    activated_on datetime,
    group_users  uuid                not null
        constraint users_groups_users
            references groups
            on delete cascade
);

create table if not exists auth_tokens
(
    id               uuid     not null
        primary key,
    created_at       datetime not null,
    updated_at       datetime not null,
    token            blob     not null,
    expires_at       datetime not null,
    user_auth_tokens uuid
        constraint auth_tokens_users_auth_tokens
            references users
            on delete cascade
);

create table if not exists auth_roles
(
    id                integer             not null
        primary key autoincrement,
    role              text default 'user' not null,
    auth_tokens_roles uuid
        constraint auth_roles_auth_tokens_roles
            references auth_tokens
            on delete cascade
);

create unique index if not exists auth_roles_auth_tokens_roles_key
    on auth_roles (auth_tokens_roles);

create unique index if not exists auth_tokens_token_key
    on auth_tokens (token);

create index if not exists authtokens_token
    on auth_tokens (token);

create table if not exists notifiers
(
    id         uuid              not null
        primary key,
    created_at datetime          not null,
    updated_at datetime          not null,
    name       text              not null,
    url        text              not null,
    is_active  bool default true not null,
    group_id   uuid              not null
        constraint notifiers_groups_notifiers
            references groups
            on delete cascade,
    user_id    uuid              not null
        constraint notifiers_users_notifiers
            references users
            on delete cascade
);

create index if not exists notifier_group_id
    on notifiers (group_id);

create index if not exists notifier_group_id_is_active
    on notifiers (group_id, is_active);

create index if not exists notifier_user_id
    on notifiers (user_id);

create index if not exists notifier_user_id_is_active
    on notifiers (user_id, is_active);

create unique index if not exists users_email_key
    on users (email);

