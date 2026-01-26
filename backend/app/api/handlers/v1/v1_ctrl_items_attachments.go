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

type (
	ItemAttachmentToken struct {
		Token string `json:"token"`
	}
)

func sanitizeAttachmentName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return name
}

// HandleItemAttachmentCreate godocs
//
//	@Summary	Create Item Attachment
//	@Tags		Items Attachments
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id		path		string	true	"Item ID"
//	@Param		file	formData	file	true	"File attachment"
//	@Param		type	formData	string	false	"Type of file"
//	@Param		primary	formData	bool	false	"Is this the primary attachment"
//	@Param		name	formData	string	true	"name of the file including extension"
//	@Success	200		{object}	repo.ItemOut
//	@Failure	422		{object}	validate.ErrorResponse
//	@Router		/v1/items/{id}/attachments [POST]
//	@Security	Bearer
//	@Deprecated
func (ctrl *V1Controller) HandleItemAttachmentCreate() errchain.HandlerFunc {
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

		item, err := ctrl.svc.Items.AttachmentAdd(
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

// HandleItemAttachmentGet godocs
//
//	@Summary	Get Item Attachment
//	@Tags		Items Attachments
//	@Produce	application/octet-stream
//	@Param		id				path		string	true	"Item ID"
//	@Param		attachment_id	path		string	true	"Attachment ID"
//	@Success	200				{object}	ItemAttachmentToken
//	@Router		/v1/items/{id}/attachments/{attachment_id} [GET]
//	@Security	Bearer
//	@Deprecated
func (ctrl *V1Controller) HandleItemAttachmentGet() errchain.HandlerFunc {
	return ctrl.handleItemAttachmentsHandler
}

// HandleItemAttachmentDelete godocs
//
//	@Summary	Delete Item Attachment
//	@Tags		Items Attachments
//	@Param		id				path	string	true	"Item ID"
//	@Param		attachment_id	path	string	true	"Attachment ID"
//	@Success	204
//	@Router		/v1/items/{id}/attachments/{attachment_id} [DELETE]
//	@Security	Bearer
//	@Deprecated
func (ctrl *V1Controller) HandleItemAttachmentDelete() errchain.HandlerFunc {
	return ctrl.handleItemAttachmentsHandler
}

// HandleItemAttachmentUpdate godocs
//
//	@Summary	Update Item Attachment
//	@Tags		Items Attachments
//	@Param		id				path		string						true	"Item ID"
//	@Param		attachment_id	path		string						true	"Attachment ID"
//	@Param		payload			body		repo.EntityAttachmentUpdate	true	"Attachment Update"
//	@Success	200				{object}	repo.ItemOut
//	@Router		/v1/items/{id}/attachments/{attachment_id} [PUT]
//	@Security	Bearer
//	@Deprecated
func (ctrl *V1Controller) HandleItemAttachmentUpdate() errchain.HandlerFunc {
	return ctrl.handleItemAttachmentsHandler
}

func (ctrl *V1Controller) handleItemAttachmentsHandler(w http.ResponseWriter, r *http.Request) error {
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
		doc, err := ctrl.svc.Items.AttachmentPath(r.Context(), ctx.GID, attachmentID)
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
		disposition := "inline; filename*=UTF-8''" + url.QueryEscape(doc.Title)
		w.Header().Set("Content-Disposition", disposition)
		http.ServeContent(w, r, doc.Title, doc.CreatedAt, file)
		return nil

	// Delete Attachment Handler
	case http.MethodDelete:
		err = ctrl.svc.Items.AttachmentDelete(r.Context(), ctx.GID, attachmentID)
		if err != nil {
			log.Err(err).Msg("failed to delete attachment")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusNoContent, nil)

	// Update Attachment Handler
	case http.MethodPut:
		var attachment repo.EntityAttachmentUpdate
		err = server.Decode(r, &attachment)
		if err != nil {
			log.Err(err).Msg("failed to decode attachment")
			return validate.NewRequestError(err, http.StatusBadRequest)
		}

		attachment.ID = attachmentID
		val, err := ctrl.svc.Items.AttachmentUpdate(ctx, ctx.GID, ID, &attachment)
		if err != nil {
			log.Err(err).Msg("failed to update attachment")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusOK, val)
	}

	return nil
}
