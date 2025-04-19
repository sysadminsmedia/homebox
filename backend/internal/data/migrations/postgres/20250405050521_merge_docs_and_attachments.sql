-- Add new columns to attachments and merge documents into attachments
ALTER TABLE attachments ADD COLUMN path TEXT DEFAULT '' NOT NULL;
ALTER TABLE attachments ADD COLUMN title TEXT DEFAULT '' NOT NULL;

UPDATE attachments
SET title = (SELECT title FROM documents WHERE documents.id = attachments.document_attachments),
    path = (SELECT path FROM documents WHERE documents.id = attachments.document_attachments)
WHERE EXISTS (SELECT 1 FROM documents WHERE documents.id = attachments.document_attachments);

ALTER TABLE attachments DROP COLUMN document_attachments;