-- +goose Up
alter table public.attachments
    alter column item_attachments drop not null;

alter table public.attachments
    add attachment_thumbnail uuid;

alter table public.attachments
    add constraint attachments_attachments_thumbnail
        foreign key (attachment_thumbnail) references public.attachments (id);

alter table public.attachments
    add constraint attachments_no_self_reference
        check (id != attachment_thumbnail);