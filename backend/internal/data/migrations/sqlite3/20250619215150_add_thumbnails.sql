-- +goose Up
create table attachments_dg_tmp
(
    id                   uuid                      not null
        primary key,
    created_at           datetime                  not null,
    updated_at           datetime                  not null,
    type                 text default 'attachment' not null,
    "primary"            bool default false        not null,
    path                 text                      not null,
    title                text                      not null,
    item_attachments     uuid
        constraint attachments_items_attachments
            references items
            on delete cascade,
    attachment_thumbnail uuid
        constraint attachments_original_thumbnail
            references attachments
            on delete cascade
);

insert into attachments_dg_tmp(id, created_at, updated_at, type, "primary", path, title, item_attachments)
select id,
       created_at,
       updated_at,
       type,
       "primary",
       path,
       title,
       item_attachments
from attachments;

drop table attachments;

alter table attachments_dg_tmp
    rename to attachments;