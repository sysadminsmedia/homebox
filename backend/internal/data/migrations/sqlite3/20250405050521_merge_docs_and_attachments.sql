-- Add new columns to attachments and merge documents into attachments
ALTER TABLE attachments ADD COLUMN path TEXT DEFAULT '' NOT NULL;
ALTER TABLE attachments ADD COLUMN title TEXT DEFAULT '' NOT NULL;

UPDATE attachments
SET title = (SELECT title FROM documents WHERE documents.id = attachments.document_attachments),
    path = (SELECT path FROM documents WHERE documents.id = attachments.document_attachments)
WHERE EXISTS (SELECT 1 FROM documents WHERE documents.id = attachments.document_attachments);

-- Create temporary table for attachments so we can remove the document_attachments column
create table attachments_tmp
(
    id                   uuid                      not null
        primary key,
    created_at           datetime                  not null,
    updated_at           datetime                  not null,
    type                 text default 'attachment' not null,
    "primary"            bool default false        not null,
    path                 text                      not null,
    title                text                      not null,
    item_attachments uuid                      not null
        constraint attachments_items_attachments
            references items
            on delete cascade
);

-- Copy data from attachments to the temporary table
INSERT INTO attachments_tmp (id, created_at, updated_at, type, "primary", path, title, item_attachments)
SELECT id, created_at, updated_at, type, "primary", path, title, item_attachments FROM attachments;

-- Drop the old attachments table
DROP TABLE attachments;

-- Drop the documents table
DROP TABLE documents;

-- Rename the temporary table to attachments
ALTER TABLE attachments_tmp RENAME TO attachments;