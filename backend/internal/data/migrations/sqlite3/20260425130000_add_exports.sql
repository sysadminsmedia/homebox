-- +goose Up
create table if not exists exports
(
    id            uuid                       not null
        primary key,
    created_at    datetime                   not null,
    updated_at    datetime                   not null,
    status        text     default 'pending' not null,
    progress      integer  default 0         not null,
    artifact_path text,
    size_bytes    integer  default 0         not null,
    error         text,
    group_id      uuid                       not null
        constraint exports_groups_exports
            references groups
            on delete cascade
);

create index if not exists export_group_id
    on exports (group_id);

create index if not exists export_group_id_status
    on exports (group_id, status);
