package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func redactExternalIdentifierForTrace(sourceType, externalID string) string {
	if sourceType != "link" {
		return externalID
	}

	u, err := url.Parse(strings.TrimSpace(externalID))
	if err != nil {
		return ""
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""

	pathHash := sha256.Sum256([]byte(u.EscapedPath()))
	return fmt.Sprintf("%s://%s/path:%s", u.Scheme, u.Host, hex.EncodeToString(pathHash[:8]))
}

func (svc *EntityService) AttachmentPath(ctx context.Context, gid uuid.UUID, attachmentID uuid.UUID) (*ent.Attachment, error) {
	ctx, span := entityServiceTracer().Start(ctx, "service.EntityService.AttachmentPath",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("attachment.id", attachmentID.String()),
		))
	defer span.End()

	attachment, err := svc.repo.Attachments.Get(ctx, gid, attachmentID)
	if err != nil {
		recordServiceSpanError(span, err)
		return nil, err
	}

	return attachment, nil
}

func (svc *EntityService) AttachmentUpdate(ctx Context, gid uuid.UUID, entityID uuid.UUID, data *repo.ItemAttachmentUpdate) (repo.EntityOut, error) {
	spanCtx, span := entityServiceTracer().Start(ctx.Context, "service.EntityService.AttachmentUpdate",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("entity.id", entityID.String()),
			attribute.String("attachment.id", data.ID.String()),
			attribute.String("attachment.type", data.Type),
			attribute.String("attachment.title", data.Title),
			attribute.Bool("attachment.primary", data.Primary),
		))
	defer span.End()
	ctx.Context = spanCtx

	updateCtx, updateSpan := entityServiceTracer().Start(spanCtx, "service.EntityService.AttachmentUpdate.update")
	attachment, err := svc.repo.Attachments.Update(updateCtx, gid, data.ID, data)
	if err != nil {
		recordServiceSpanError(updateSpan, err)
		updateSpan.End()
		recordServiceSpanError(span, err)
		return repo.EntityOut{}, err
	}
	updateSpan.End()

	renameCtx, renameSpan := entityServiceTracer().Start(spanCtx, "service.EntityService.AttachmentUpdate.rename")
	_, err = svc.repo.Attachments.Rename(renameCtx, gid, attachment.ID, data.Title)
	if err != nil {
		recordServiceSpanError(renameSpan, err)
		renameSpan.End()
		recordServiceSpanError(span, err)
		return repo.EntityOut{}, err
	}
	renameSpan.End()

	out, err := svc.repo.Entities.GetOneByGroup(ctx, ctx.GID, entityID)
	if err != nil {
		recordServiceSpanError(span, err)
	}
	return out, err
}

// AttachmentAdd adds an attachment to an entity by creating an entry in the Documents table and linking it to the Attachment
// Table and Entities table. The file provided via the reader is stored on the file system based on the provided
// relative path during construction of the service.
func (svc *EntityService) AttachmentAdd(ctx Context, entityID uuid.UUID, filename string, attachmentType attachment.Type, primary bool, file io.Reader) (repo.EntityOut, error) {
	spanCtx, span := entityServiceTracer().Start(ctx.Context, "service.EntityService.AttachmentAdd",
		trace.WithAttributes(
			attribute.String("group.id", ctx.GID.String()),
			attribute.String("entity.id", entityID.String()),
			attribute.String("attachment.filename", filename),
			attribute.String("attachment.type", attachmentType.String()),
			attribute.Bool("attachment.primary", primary),
		))
	defer span.End()
	ctx.Context = spanCtx

	verifyCtx, verifySpan := entityServiceTracer().Start(spanCtx, "service.EntityService.AttachmentAdd.verifyEntity")
	_, err := svc.repo.Entities.GetOneByGroup(verifyCtx, ctx.GID, entityID)
	if err != nil {
		recordServiceSpanError(verifySpan, err)
		verifySpan.End()
		recordServiceSpanError(span, err)
		return repo.EntityOut{}, err
	}
	verifySpan.End()

	createCtx, createSpan := entityServiceTracer().Start(spanCtx, "service.EntityService.AttachmentAdd.create")
	_, err = svc.repo.Attachments.Create(createCtx, entityID, repo.ItemCreateAttachment{Title: filename, Content: file}, attachmentType, primary)
	if err != nil {
		recordServiceSpanError(createSpan, err)
		createSpan.End()
		recordServiceSpanError(span, err)
		log.Err(err).Msg("failed to create attachment")
		return repo.EntityOut{}, err
	}
	createSpan.End()

	out, err := svc.repo.Entities.GetOneByGroup(ctx, ctx.GID, entityID)
	if err != nil {
		recordServiceSpanError(span, err)
	}
	return out, err
}

func (svc *EntityService) AttachmentAddExternalLink(ctx Context, entityID uuid.UUID, sourceType, externalID, title string, attType attachment.Type) (repo.EntityOut, error) {
	spanCtx, span := entityServiceTracer().Start(ctx.Context, "service.EntityService.AttachmentAddExternalLink",
		trace.WithAttributes(
			attribute.String("group.id", ctx.GID.String()),
			attribute.String("entity.id", entityID.String()),
			attribute.String("integration.source_type", sourceType),
			attribute.String("integration.external_id", redactExternalIdentifierForTrace(sourceType, externalID)),
		))
	defer span.End()
	ctx.Context = spanCtx

	mimeType, ok := repo.MimeTypeForSourceType(sourceType)
	if !ok {
		err := fmt.Errorf("unknown source_type %q", sourceType)
		recordServiceSpanError(span, err)
		return repo.EntityOut{}, err
	}

	verifyCtx, verifySpan := entityServiceTracer().Start(spanCtx, "service.EntityService.AttachmentAddExternalLink.verifyEntity")
	_, err := svc.repo.Entities.GetOneByGroup(verifyCtx, ctx.GID, entityID)
	if err != nil {
		recordServiceSpanError(verifySpan, err)
		verifySpan.End()
		recordServiceSpanError(span, err)
		return repo.EntityOut{}, err
	}
	verifySpan.End()

	createCtx, createSpan := entityServiceTracer().Start(spanCtx, "service.EntityService.AttachmentAddExternalLink.create")
	_, err = svc.repo.Attachments.CreateExternalLink(createCtx, entityID, externalID, title, mimeType, attType)
	if err != nil {
		recordServiceSpanError(createSpan, err)
		createSpan.End()
		recordServiceSpanError(span, err)
		log.Err(err).Msg("failed to create external link attachment")
		return repo.EntityOut{}, err
	}
	createSpan.End()

	out, err := svc.repo.Entities.GetOneByGroup(ctx, ctx.GID, entityID)
	if err != nil {
		recordServiceSpanError(span, err)
	}
	return out, err
}

func (svc *EntityService) AttachmentDelete(ctx context.Context, gid uuid.UUID, attachmentID uuid.UUID) error {
	ctx, span := entityServiceTracer().Start(ctx, "service.EntityService.AttachmentDelete",
		trace.WithAttributes(
			attribute.String("group.id", gid.String()),
			attribute.String("attachment.id", attachmentID.String()),
		))
	defer span.End()

	err := svc.repo.Attachments.Delete(ctx, gid, attachmentID)
	if err != nil {
		recordServiceSpanError(span, err)
		return err
	}

	return nil
}
