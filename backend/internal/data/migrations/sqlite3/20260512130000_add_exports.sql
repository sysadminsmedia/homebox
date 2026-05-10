-- +goose Up
create table if not exists exports
(
    id            uuid                       not null
        primary key,
    created_at    datetime                   not null,
    updated_at    datetime                   not null,
    kind          text     default 'export'  not null
        check (kind in ('export', 'import')),
    status        text     default 'pending' not null
        check (status in ('pending', 'running', 'completed', 'failed')),
    progress      integer  default 0         not null,
    artifact_path text,
    size_bytes    integer  default 0         not null,
    error         text
        check (error is null or length(error) <= 1000),
    group_id      uuid                       not null
        constraint exports_groups_exports
            references groups
            on delete cascade
);

create index if not exists export_group_id
    on exports (group_id);

create index if not exists export_group_id_status
    on exports (group_id, status);
