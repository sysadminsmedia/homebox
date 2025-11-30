-- +goose Up
create table if not exists item_templates
(
    id                           uuid                  not null
        primary key,
    created_at                   datetime              not null,
    updated_at                   datetime              not null,
    name                         text                  not null,
    description                  text,
    notes                        text,
    default_quantity             integer default 1     not null,
    default_insured              bool    default false not null,
    default_manufacturer         text,
    default_lifetime_warranty    bool    default false not null,
    default_warranty_details     text,
    include_warranty_fields      bool    default false not null,
    include_purchase_fields      bool    default false not null,
    include_sold_fields          bool    default false not null,
    group_item_templates         uuid                  not null
        constraint item_templates_groups_item_templates
            references groups
            on delete cascade
);

create table if not exists template_fields
(
    id                   uuid                  not null
        primary key,
    created_at           datetime              not null,
    updated_at           datetime              not null,
    name                 text                  not null,
    description          text,
    type                 text                  not null,
    text_value           text,
    item_template_fields uuid
        constraint template_fields_item_templates_fields
            references item_templates
            on delete cascade
);
