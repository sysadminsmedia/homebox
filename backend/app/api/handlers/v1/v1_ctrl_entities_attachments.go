package v1

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/attachment"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/memblob"
	_ "gocloud.dev/blob/s3blob"
)

func sanitizeAttachmentName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return name
}

// HandleEntityAttachmentCreate godoc
//
//	@Summary	Create Entity Attachment
//	@Tags		Entities Attachments
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id		path		string	true	"Entity ID"
//	@Param		file	formData	file	true	"File attachment"
//	@Param		type	formData	string	false	"Type of file"
//	@Param		primary	formData	bool	false	"Is this the primary attachment"
//	@Param		name	formData	string	true	"name of the file including extension"
//	@Success	201		{object}	repo.EntityOut
//	@Failure	422		{object}	validate.ErrorResponse
//	@Router		/v1/entities/{id}/attachments [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityAttachmentCreate() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseMultipartForm(ctrl.maxUploadSize << 20)
		if err != nil {
			log.Err(err).Msg("failed to parse multipart form")
			return validate.NewRequestError(errors.New("failed to parse multipart form"), http.StatusBadRequest)
		}

		errs := validate.NewFieldErrors()

		file, _, err := r.FormFile("file")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrMissingFile):
				log.Debug().Msg("file for attachment is missing")
				errs = errs.Append("file", "file is required")
			default:
				log.Err(err).Msg("failed to get file from form")
				return validate.NewRequestError(err, http.StatusInternalServerError)
			}
		}

		attachmentName := r.FormValue("name")
		if attachmentName == "" {
			log.Debug().Msg("failed to get name from form")
			errs = errs.Append("name", "name is required")
		}

		if !errs.Nil() {
			return server.JSON(w, http.StatusUnprocessableEntity, errs)
		}

		attachmentName = sanitizeAttachmentName(attachmentName)

		attachmentType := r.FormValue("type")
		if attachmentType == "" {
			// Attempt to auto-detect the type of the file
			ext := filepath.Ext(attachmentName)

			switch strings.ToLower(ext) {
			case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff", ".avif", ".ico", ".heic", ".jxl":
				attachmentType = attachment.TypePhoto.String()
			default:
				attachmentType = attachment.TypeAttachment.String()
			}
		}

		primary, err := strconv.ParseBool(r.FormValue("primary"))
		if err != nil {
			log.Debug().Msg("failed to parse primary from form")
			primary = false
		}

		id, err := ctrl.routeID(r)
		if err != nil {
			return err
		}

		ctx := services.NewContext(r.Context())

		item, err := ctrl.svc.Entities.AttachmentAdd(
			ctx,
			id,
			attachmentName,
			attachment.Type(attachmentType),
			primary,
			file,
		)
		if err != nil {
			log.Err(err).Msg("failed to add attachment")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusCreated, item)
	}
}

// HandleEntityAttachmentGet godoc
//
//	@Summary	Get Entity Attachment
//	@Tags		Entities Attachments
//	@Produce	application/octet-stream
//	@Param		id				path		string	true	"Entity ID"
//	@Param		attachment_id	path		string	true	"Attachment ID"
//	@Success	200				{file}		file
//	@Router		/v1/entities/{id}/attachments/{attachment_id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityAttachmentGet() errchain.HandlerFunc {
	return ctrl.handleEntityAttachmentsHandler
}

// HandleEntityAttachmentDelete godoc
//
//	@Summary	Delete Entity Attachment
//	@Tags		Entities Attachments
//	@Param		id				path	string	true	"Entity ID"
//	@Param		attachment_id	path	string	true	"Attachment ID"
//	@Success	204
//	@Router		/v1/entities/{id}/attachments/{attachment_id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityAttachmentDelete() errchain.HandlerFunc {
	return ctrl.handleEntityAttachmentsHandler
}

// HandleEntityAttachmentUpdate godoc
//
//	@Summary	Update Entity Attachment
//	@Tags		Entities Attachments
//	@Param		id				path		string						true	"Entity ID"
//	@Param		attachment_id	path		string						true	"Attachment ID"
//	@Param		payload			body		repo.ItemAttachmentUpdate	true	"Attachment Update"
//	@Success	200				{object}	repo.EntityOut
//	@Router		/v1/entities/{id}/attachments/{attachment_id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityAttachmentUpdate() errchain.HandlerFunc {
	return ctrl.handleEntityAttachmentsHandler
}

func (ctrl *V1Controller) handleEntityAttachmentsHandler(w http.ResponseWriter, r *http.Request) error {
	ID, err := ctrl.routeID(r)
	if err != nil {
		return err
	}

	attachmentID, err := ctrl.routeUUID(r, "attachment_id")
	if err != nil {
		return err
	}

	ctx := services.NewContext(r.Context())
	switch r.Method {
	case http.MethodGet:
		doc, err := ctrl.svc.Entities.AttachmentPath(r.Context(), ctx.GID, attachmentID)
		if err != nil {
			log.Err(err).Msg("failed to get attachment path")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		bucket, err := blob.OpenBucket(ctx, ctrl.repo.Attachments.GetConnString())
		if err != nil {
			log.Err(err).Msg("failed to open bucket")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		file, err := bucket.NewReader(ctx, ctrl.repo.Attachments.GetFullPath(doc.Path), nil)
		if err != nil {
			log.Err(err).Msg("failed to open file")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		defer func(file *blob.Reader) {
			err := file.Close()
			if err != nil {
				log.Err(err).Msg("failed to close file")
			}
		}(file)
		defer func(bucket *blob.Bucket) {
			err := bucket.Close()
			if err != nil {
				log.Err(err).Msg("failed to close bucket")
			}
		}(bucket)

		// Set the Content-Disposition header for RFC6266 compliance
		disposition := "attachment"
		if isSafeInlineType(doc.MimeType) {
			disposition = "inline"
		}
		disposition += "; filename*=UTF-8''" + url.QueryEscape(doc.Title)
		w.Header().Set("Content-Disposition", disposition)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Download-Options", "noopen")
		// Set strict CSP for all attachments to prevent any script execution
		// Even for "safe" types, we want to ensure no inline scripts can execute
		w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; sandbox;")
		http.ServeContent(w, r, doc.Title, doc.CreatedAt, file)
		return nil

	// Delete Attachment Handler
	case http.MethodDelete:
		err = ctrl.svc.Entities.AttachmentDelete(r.Context(), ctx.GID, attachmentID)
		if err != nil {
			log.Err(err).Msg("failed to delete attachment")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusNoContent, nil)

	// Update Attachment Handler
	case http.MethodPut:
		var attachment repo.ItemAttachmentUpdate
		err = server.Decode(r, &attachment)
		if err != nil {
			log.Err(err).Msg("failed to decode attachment")
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		attachment.ID = attachmentID
		val, err := ctrl.svc.Entities.AttachmentUpdate(ctx, ctx.GID, ID, &attachment)
		if err != nil {
			log.Err(err).Msg("failed to update attachment")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusOK, val)
	}

	return nil
}

// isSafeInlineType returns true if the MIME type is safe to display inline
func isSafeInlineType(mimeType string) bool {
	safeMimeTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"image/bmp":       true,
		"image/tiff":      true,
		"image/avif":      true,
		"image/ico":       true,
		"image/x-icon":    true,
		"application/pdf": true, // PDFs are generally safe with proper CSP
	}

	// Check exact match
	if safeMimeTypes[strings.ToLower(mimeType)] {
		return true
	}

	// Block any text/html, application/javascript, image/svg+xml explicitly
	dangerousTypes := []string{
		"text/html",
		"text/xml",
		"application/xhtml",
		"application/xml",
		"application/javascript",
		"text/javascript",
		"image/svg+xml",
		"image/svg",
	}

	lowerMime := strings.ToLower(mimeType)
	for _, dangerous := range dangerousTypes {
		if strings.Contains(lowerMime, dangerous) {
			return false
		}
	}

	return false
}
