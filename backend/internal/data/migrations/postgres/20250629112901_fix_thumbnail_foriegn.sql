-- +goose Up
alter table public.attachments
    drop constraint attachments_attachments_thumbnail;

alter table public.attachments
    add constraint attachments_attachments_thumbnail
        foreign key (attachment_thumbnail) references public.attachments
    on delete set null;