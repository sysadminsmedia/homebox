package v1

import (
	"errors"
	"net/http"
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
)

type (
	ItemAttachmentToken struct {
		Token string `json:"token"`
	}
)

// HandleItemAttachmentCreate godocs
//
//	@Summary  Create Item Attachment
//	@Tags     Items Attachments
//	@Accept   multipart/form-data
//	@Produce  json
//	@Param    id   path     string true "Item ID"
//	@Param    file formData file   true "File attachment"
//	@Param    type formData string true "Type of file"
//	@Param    primary formData bool false "Is this the primary attachment"
//	@Param    name formData string true "name of the file including extension"
//	@Success  200  {object} repo.ItemOut
//	@Failure  422  {object} validate.ErrorResponse
//	@Router   /v1/items/{id}/attachments [POST]
//	@Security Bearer
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

		attachmentType := r.FormValue("type")
		if attachmentType == "" {
			// Attempt to auto-detect the type of the file
			ext := filepath.Ext(attachmentName)

			switch strings.ToLower(ext) {
			case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff":
				attachmentType = attachment.TypePhoto.String()
			default:
				attachmentType = attachment.TypeAttachment.String()
			}
		}

		primary, err := strconv.ParseBool(r.FormValue("primary"))
		autoPrimary := false
		if err != nil {
			log.Debug().Msg("failed to parse primary from form")
			primary = false
			autoPrimary = true
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
			autoPrimary,
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
func (ctrl *V1Controller) HandleItemAttachmentDelete() errchain.HandlerFunc {
	return ctrl.handleItemAttachmentsHandler
}

// HandleItemAttachmentUpdate godocs
//
//	@Summary	Update Item Attachment
//	@Tags		Items Attachments
//	@Param		id				path		string						true	"Item ID"
//	@Param		attachment_id	path		string						true	"Attachment ID"
//	@Param		payload			body		repo.ItemAttachmentUpdate	true	"Attachment Update"
//	@Success	200				{object}	repo.ItemOut
//	@Router		/v1/items/{id}/attachments/{attachment_id} [PUT]
//	@Security	Bearer
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
		doc, err := ctrl.svc.Items.AttachmentPath(r.Context(), attachmentID)
		if err != nil {
			log.Err(err).Msg("failed to get attachment path")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}
		// w.Header().Set("Content-Disposition", "attachment; filename="+doc.Title)
		http.ServeFile(w, r, doc.Path)
		return nil

	// Delete Attachment Handler
	case http.MethodDelete:
		err = ctrl.svc.Items.AttachmentDelete(r.Context(), ctx.GID, ID, attachmentID)
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
		val, err := ctrl.svc.Items.AttachmentUpdate(ctx, ID, &attachment)
		if err != nil {
			log.Err(err).Msg("failed to delete attachment")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusOK, val)
	}

	return nil
}
