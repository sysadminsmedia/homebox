-- +goose Up
-- Printers for direct label printing
create table if not exists printers
(
    id                   uuid                       not null
        primary key,
    created_at           datetime                   not null,
    updated_at           datetime                   not null,
    name                 text                       not null,
    description          text,
    printer_type         text     default 'ipp'     not null,
    address              text                       not null,
    is_default           bool     default false     not null,
    label_width_mm       real,
    label_height_mm      real,
    dpi                  integer  default 300       not null,
    media_type           text,
    status               text     default 'unknown' not null,
    last_status_check    datetime,
    group_printers       uuid                       not null
        constraint printers_groups_printers
            references groups
            on delete cascade
);

create index idx_printers_name on printers(name);
create index idx_printers_is_default on printers(is_default);
create index idx_printers_printer_type on printers(printer_type);
create index idx_printers_group on printers(group_printers);

-- +goose Down
drop index if exists idx_printers_group;
drop index if exists idx_printers_printer_type;
drop index if exists idx_printers_is_default;
drop index if exists idx_printers_name;
drop table if exists printers;
