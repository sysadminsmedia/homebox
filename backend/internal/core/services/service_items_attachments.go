package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"io"
)

func (svc *ItemService) AttachmentPath(ctx context.Context, attachmentID uuid.UUID) (*ent.Document, error) {
	attachments, err := svc.repo.Attachments.Get(ctx, attachmentID)
	if err != nil {
		return nil, err
	}

	return attachments.Edges.Document, nil
}

func (svc *ItemService) AttachmentUpdate(ctx Context, itemID uuid.UUID, data *repo.ItemAttachmentUpdate) (repo.ItemOut, error) {
	// Update Attachment
	attachments, err := svc.repo.Attachments.Update(ctx, data.ID, data)
	if err != nil {
		return repo.ItemOut{}, err
	}

	// Update Document
	attDoc := attachments.Edges.Document
	_, err = svc.repo.Docs.Rename(ctx, attDoc.ID, data.Title)
	if err != nil {
		return repo.ItemOut{}, err
	}

	return svc.repo.Items.GetOneByGroup(ctx, ctx.GID, itemID)
}

// AttachmentAdd adds an attachment to an item by creating an entry in the Documents table and linking it to the Attachment
// Table and Items table. The file provided via the reader is stored on the file system based on the provided
// relative path during construction of the service.
func (svc *ItemService) AttachmentAdd(ctx Context, itemID uuid.UUID, filename string, attachmentType attachment.Type, file io.Reader) (repo.ItemOut, error) {
	// Get the Item
	_, err := svc.repo.Items.GetOneByGroup(ctx, ctx.GID, itemID)
	if err != nil {
		return repo.ItemOut{}, err
	}

	// Create the document
	doc, err := svc.repo.Docs.Create(ctx, ctx.GID, repo.DocumentCreate{Title: filename, Content: file})
	if err != nil {
		log.Err(err).Msg("failed to create document")
		return repo.ItemOut{}, err
	}

	// Create the attachment
	_, err = svc.repo.Attachments.Create(ctx, itemID, doc.ID, attachmentType)
	if err != nil {
		log.Err(err).Msg("failed to create attachment")
		return repo.ItemOut{}, err
	}

	return svc.repo.Items.GetOneByGroup(ctx, ctx.GID, itemID)
}

func (svc *ItemService) AttachmentDelete(ctx context.Context, gid, itemID, attachmentID uuid.UUID) error {
	// Get the Item
	_, err := svc.repo.Items.GetOneByGroup(ctx, gid, itemID)
	if err != nil {
		return err
	}

	attachments, err := svc.repo.Attachments.Get(ctx, attachmentID)
	if err != nil {
		return err
	}

	documentID := attachment.Edges.Document.GetID()

	// Delete the attachment
	err = svc.repo.Attachments.Delete(ctx, attachmentID)
	if err != nil {
		return err
	}

	// Remove File
	err = os.Remove(attachments.Edges.Document.Path)
	// Delete the document, this function also removes the file
	err = svc.repo.Docs.Delete(ctx, documentID)
	if err != nil {
		return err
	}

	return err
}
