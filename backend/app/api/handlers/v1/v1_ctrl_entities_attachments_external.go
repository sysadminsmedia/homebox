package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"go.opentelemetry.io/otel/attribute"
)

type externalAttachmentRequest struct {
	SourceType     string `json:"source_type"`
	ExternalID     string `json:"external_id"`
	Title          string `json:"title"`
	AttachmentType string `json:"attachment_type"`
}

func parseExternalHTTPURL(raw string) (*url.URL, bool) {
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil {
		return nil, false
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, schemeHTTPS) {
		return nil, false
	}
	if u.Host == "" || u.User != nil {
		return nil, false
	}
	return u, true
}

func redactExternalURLForTrace(raw string) string {
	u, ok := parseExternalHTTPURL(raw)
	if !ok {
		return ""
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// HandleEntityAttachmentExternalCreate godoc
//
//	@Summary	Create External Link Attachment
//	@Description	Links an entity to a document or URL in an external system without copying
//				the file into Homebox. The source is identified by source_type.
//	@Tags		Entities Attachments
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string					true	"Entity ID"
//	@Param		payload	body		externalAttachmentRequest	true	"External document reference"
//	@Success	201		{object}	repo.EntityOut
//	@Failure	400		{object}	validate.ErrorResponse
//	@Failure	422		{object}	validate.ErrorResponse
//	@Router		/v1/entities/{id}/attachments/external [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityAttachmentExternalCreate() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		_, span := startEntityCtrlSpan(r.Context(), "controller.V1.HandleEntityAttachmentExternalCreate")
		defer span.End()

		var body externalAttachmentRequest
		if err := server.Decode(r, &body); err != nil {
			recordCtrlSpanError(span, err)
			log.Err(err).Msg("failed to decode external attachment request")
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		body.SourceType = strings.TrimSpace(body.SourceType)
		body.ExternalID = strings.TrimSpace(body.ExternalID)

		if body.SourceType == "" {
			return server.JSON(w, http.StatusUnprocessableEntity,
				validate.NewFieldErrors().Append("source_type", "source_type is required"))
		}
		if body.ExternalID == "" {
			return server.JSON(w, http.StatusUnprocessableEntity,
				validate.NewFieldErrors().Append("external_id", "external_id is required"))
		}
		if _, ok := repo.MimeTypeForSourceType(body.SourceType); !ok {
			return server.JSON(w, http.StatusUnprocessableEntity,
				validate.NewFieldErrors().Append("source_type", fmt.Sprintf("unknown source_type %q", body.SourceType)))
		}
		if body.SourceType == "link" {
			if _, ok := parseExternalHTTPURL(body.ExternalID); !ok {
				return server.JSON(w, http.StatusUnprocessableEntity,
					validate.NewFieldErrors().Append("external_id", "external_id must be a valid http/https URL"))
			}
		}

		title := strings.TrimSpace(body.Title)
		if title == "" {
			title = body.ExternalID
		}

		id, err := ctrl.routeID(r)
		if err != nil {
			recordCtrlSpanError(span, err)
			return err
		}

		span.SetAttributes(
			attribute.String("entity.id", id.String()),
			attribute.String("integration.source_type", body.SourceType),
			attribute.String("integration.external_id", redactExternalURLForTrace(body.ExternalID)),
		)

		ctx := services.NewContext(r.Context())
		span.SetAttributes(attribute.String("group.id", ctx.GID.String()))

		attType := attachment.Type(strings.TrimSpace(body.AttachmentType))
		switch attType {
		case attachment.TypePhoto, attachment.TypeManual, attachment.TypeWarranty,
			attachment.TypeAttachment, attachment.TypeReceipt:
			// valid
		default:
			attType = attachment.TypeAttachment
		}

		item, err := ctrl.svc.Entities.AttachmentAddExternalLink(ctx, id, body.SourceType, body.ExternalID, title, attType)
		if err != nil {
			recordCtrlSpanError(span, err)
			log.Err(err).Msg("failed to add external link attachment")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusCreated, item)
	}
}
