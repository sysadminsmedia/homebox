-- +goose Up
-- Label templates for customizable label printing
create table if not exists label_templates
(
    id                       uuid                     not null
        primary key,
    created_at               datetime                 not null,
    updated_at               datetime                 not null,
    name                     text                     not null,
    description              text,
    width                    real     default 62.0    not null,
    height                   real     default 29.0    not null,
    preset                   text,
    is_shared                bool     default false   not null,
    canvas_data              json,
    output_format            text     default 'png'   not null,
    dpi                      integer  default 300     not null,
    owner_id                 uuid                     not null
        constraint label_templates_users_label_templates
            references users
            on delete cascade,
    group_label_templates    uuid                     not null
        constraint label_templates_groups_label_templates
            references groups
            on delete cascade
);

create index idx_label_templates_name on label_templates(name);
create index idx_label_templates_is_shared on label_templates(is_shared);
create index idx_label_templates_preset on label_templates(preset);
create index idx_label_templates_owner on label_templates(owner_id);
create index idx_label_templates_group on label_templates(group_label_templates);

-- +goose Down
drop index if exists idx_label_templates_group;
drop index if exists idx_label_templates_owner;
drop index if exists idx_label_templates_preset;
drop index if exists idx_label_templates_is_shared;
drop index if exists idx_label_templates_name;
drop table if exists label_templates;
