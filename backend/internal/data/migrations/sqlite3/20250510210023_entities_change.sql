-- +goose Up
-- +goose no transaction
create table entities
(
    id                            uuid                  not null
        primary key,
    type                          text                  not null,
    created_at                    datetime              not null,
    updated_at                    datetime              not null,
    name                          text                  not null,
    description                   text,
    import_ref                    text,
    notes                         text,
    quantity                      integer default 1     not null,
    insured                       bool    default false not null,
    archived                      bool    default false not null,
    asset_id                      integer default 0     not null,
    serial_number                 text,
    model_number                  text,
    manufacturer                  text,
    lifetime_warranty             bool    default false not null,
    warranty_expires              datetime,
    warranty_details              text,
    purchase_time                 datetime,
    purchase_from                 text,
    purchase_price                real    default 0     not null,
    sold_time                     datetime,
    sold_to                       text,
    sold_price                    real    default 0     not null,
    sold_notes                    text,
    group_entities                uuid                  not null
        constraint entities_groups_entities
            references groups
            on delete cascade,
    entity_children               uuid
        constraint entities_entities_children
            references entities
            on delete set null,
    location_entities             uuid
        constraint entities_locations_entities
            references entities
            on delete cascade,
    sync_child_entities_locations BOOLEAN default FALSE not null
);

create index entity_archived
    on entities (archived);

create index entity_asset_id
    on entities (asset_id);

create index entity_manufacturer
    on entities (manufacturer);

create index entity_model_number
    on entities (model_number);

create index entity_name
    on entities (name);

create index entity_serial_number
    on entities (serial_number);

PRAGMA FOREIGN_KEYS = OFF;

-- Migrate the item_fields table to the new entity_fields table
create table entity_fields
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
    entity_fields uuid
        constraint entity_fields_entities_fields
            references entities
            on delete cascade
);

insert into entity_fields(id, created_at, updated_at, name, description, type, text_value, number_value,
                          boolean_value, time_value, entity_fields)
select id,
       created_at,
       updated_at,
       name,
       description,
       type,
       text_value,
       number_value,
       boolean_value,
       time_value,
       item_fields
from item_fields;

drop table item_fields;

-- Update maintenance_entries to use the new entities table
create table maintenance_entries_dg_tmp
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
    entity_id      uuid           not null
        constraint maintenance_entries_entities_maintenance_entries
            references entities
            on delete cascade
);

insert into maintenance_entries_dg_tmp(id, created_at, updated_at, date, scheduled_date, name, description, cost,
                                       entity_id)
select id,
       created_at,
       updated_at,
       date,
       scheduled_date,
       name,
       description,
       cost,
       item_id
from maintenance_entries;

drop table maintenance_entries;

alter table maintenance_entries_dg_tmp
    rename to maintenance_entries;

-- Migrate the locations first
INSERT INTO entities (id, type, created_at, updated_at, name, description, group_entities, entity_children)
SELECT id,
       'location',
       created_at,
       updated_at,
       name,
       description,
       group_locations,
       location_children
FROM locations;

-- Then migrate the items
INSERT INTO entities (id, type, created_at, updated_at, name, description, import_ref, notes, quantity, insured,
                      archived, asset_id, serial_number, model_number, manufacturer, lifetime_warranty,
                      warranty_expires, warranty_details, purchase_time, purchase_from, purchase_price, sold_time,
                      sold_to, sold_price, sold_notes, group_entities, entity_children, location_entities)
SELECT id,
       'item',
       created_at,
       updated_at,
       name,
       description,
       import_ref,
       notes,
       quantity,
       insured,
       archived,
       asset_id,
       serial_number,
       model_number,
       manufacturer,
       lifetime_warranty,
       warranty_expires,
       warranty_details,
       purchase_time,
       purchase_from,
       purchase_price,
       sold_time,
       sold_to,
       sold_price,
       sold_notes,
       group_items,
       item_children,
       location_items
FROM items;

PRAGMA FOREIGN_KEYS = ON;

DROP TABLE locations;
DROP TABLE items;