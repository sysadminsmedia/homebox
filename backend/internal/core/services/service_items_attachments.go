package services

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

func (svc *EntityService) AttachmentPath(ctx context.Context, gid uuid.UUID, attachmentID uuid.UUID) (*ent.Attachment, error) {
	attachment, err := svc.repo.Attachments.Get(ctx, gid, attachmentID)
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

func (svc *EntityService) AttachmentUpdate(ctx Context, gid uuid.UUID, entityID uuid.UUID, data *repo.ItemAttachmentUpdate) (repo.EntityOut, error) {
	// Update Attachment
	attachment, err := svc.repo.Attachments.Update(ctx, gid, data.ID, data)
	if err != nil {
		return repo.EntityOut{}, err
	}

	// Update Document
	attDoc := attachment
	_, err = svc.repo.Attachments.Rename(ctx, gid, attDoc.ID, data.Title)
	if err != nil {
		return repo.EntityOut{}, err
	}

	return svc.repo.Entities.GetOneByGroup(ctx, ctx.GID, entityID)
}

// AttachmentAdd adds an attachment to an entity by creating an entry in the Documents table and linking it to the Attachment
// Table and Entities table. The file provided via the reader is stored on the file system based on the provided
// relative path during construction of the service.
func (svc *EntityService) AttachmentAdd(ctx Context, entityID uuid.UUID, filename string, attachmentType attachment.Type, primary bool, file io.Reader) (repo.EntityOut, error) {
	// Get the Entity
	_, err := svc.repo.Entities.GetOneByGroup(ctx, ctx.GID, entityID)
	if err != nil {
		return repo.EntityOut{}, err
	}

	// Create the attachment
	_, err = svc.repo.Attachments.Create(ctx, entityID, repo.ItemCreateAttachment{Title: filename, Content: file}, attachmentType, primary)
	if err != nil {
		log.Err(err).Msg("failed to create attachment")
		return repo.EntityOut{}, err
	}

	return svc.repo.Entities.GetOneByGroup(ctx, ctx.GID, entityID)
}

func (svc *EntityService) AttachmentDelete(ctx context.Context, gid uuid.UUID, attachmentID uuid.UUID) error {
	// Delete the attachment
	err := svc.repo.Attachments.Delete(ctx, gid, attachmentID)
	if err != nil {
		return err
	}

	return err
}
