-- +goose Up
alter table public.attachments
    alter column item_attachments drop not null;

alter table public.attachments
    add attachment_original uuid;

alter table public.attachments
    add constraint attachments_original_thumbnail
        foreign key (attachment_original) references public.attachments (id);

alter table public.attachments
    add constraint attachments_no_self_reference
        check (id != attachment_original);