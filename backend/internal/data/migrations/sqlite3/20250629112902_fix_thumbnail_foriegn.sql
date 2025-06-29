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
    mime_type            text default 'application/octet-stream' not null,
    item_attachments     uuid
        constraint attachments_items_attachments
            references items
            on delete cascade,
    attachment_thumbnail uuid
        constraint attachments_attachments_thumbnail
            references attachments
            on delete set null
);

insert into attachments_dg_tmp(id, created_at, updated_at, type, "primary", path, title, mime_type, item_attachments,
                               attachment_thumbnail)
select id,
       created_at,
       updated_at,
       type,
       "primary",
       path,
       title,
       mime_type,
       item_attachments,
       attachment_thumbnail
from attachments;

drop table attachments;

alter table attachments_dg_tmp
    rename to attachments;

