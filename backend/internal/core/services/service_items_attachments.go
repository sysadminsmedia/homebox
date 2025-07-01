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

func (svc *ItemService) AttachmentPath(ctx context.Context, gid uuid.UUID, attachmentID uuid.UUID) (*ent.Attachment, error) {
	attachment, err := svc.repo.Attachments.Get(ctx, gid, attachmentID)
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

func (svc *ItemService) AttachmentUpdate(ctx Context, gid uuid.UUID, itemID uuid.UUID, data *repo.ItemAttachmentUpdate) (repo.ItemOut, error) {
	// Update Attachment
	attachment, err := svc.repo.Attachments.Update(ctx, gid, data.ID, data)
	if err != nil {
		return repo.ItemOut{}, err
	}

	// Update Document
	attDoc := attachment
	_, err = svc.repo.Attachments.Rename(ctx, gid, attDoc.ID, data.Title)
	if err != nil {
		return repo.ItemOut{}, err
	}

	return svc.repo.Items.GetOneByGroup(ctx, ctx.GID, itemID)
}

// AttachmentAdd adds an attachment to an item by creating an entry in the Documents table and linking it to the Attachment
// Table and Items table. The file provided via the reader is stored on the file system based on the provided
// relative path during construction of the service.
func (svc *ItemService) AttachmentAdd(ctx Context, itemID uuid.UUID, filename string, attachmentType attachment.Type, primary bool, file io.Reader) (repo.ItemOut, error) {
	// Get the Item
	_, err := svc.repo.Items.GetOneByGroup(ctx, ctx.GID, itemID)
	if err != nil {
		return repo.ItemOut{}, err
	}

	// Create the attachment
	_, err = svc.repo.Attachments.Create(ctx, itemID, repo.ItemCreateAttachment{Title: filename, Content: file}, attachmentType, primary)
	if err != nil {
		log.Err(err).Msg("failed to create attachment")
	}

	return svc.repo.Items.GetOneByGroup(ctx, ctx.GID, itemID)
}

func (svc *ItemService) AttachmentDelete(ctx context.Context, gid uuid.UUID, id uuid.UUID, attachmentID uuid.UUID) error {
	// Delete the attachment
	err := svc.repo.Attachments.Delete(ctx, gid, id, attachmentID)
	if err != nil {
		return err
	}

	return err
}
