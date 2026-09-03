package v1

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/hay-kot/httpkit/server"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/sys/validate"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleEntityTemplatesGetAll godoc
//
//	@Summary	Get All Entity Templates
//	@Tags		Entity Templates
//	@Produce	json
//	@Success	200	{array}	repo.EntityTemplateSummary
//	@Router		/v1/templates [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTemplatesGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.EntityTemplateSummary, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTemplates.GetAll(r.Context(), auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleEntityTemplatesGet godoc
//
//	@Summary	Get Entity Template
//	@Tags		Entity Templates
//	@Produce	json
//	@Param		id	path		string	true	"Template ID"
//	@Success	200	{object}	repo.EntityTemplateOut
//	@Router		/v1/templates/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTemplatesGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.EntityTemplateOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTemplates.GetOne(r.Context(), auth.GID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleEntityTemplatesCreate godoc
//
//	@Summary	Create Entity Template
//	@Tags		Entity Templates
//	@Produce	json
//	@Param		payload	body		repo.EntityTemplateCreate	true	"Template Data"
//	@Success	201		{object}	repo.EntityTemplateOut
//	@Router		/v1/templates [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTemplatesCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.EntityTemplateCreate) (repo.EntityTemplateOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTemplates.Create(r.Context(), auth.GID, body)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleEntityTemplatesUpdate godoc
//
//	@Summary	Update Entity Template
//	@Tags		Entity Templates
//	@Produce	json
//	@Param		id		path		string						true	"Template ID"
//	@Param		payload	body		repo.EntityTemplateUpdate	true	"Template Data"
//	@Success	200		{object}	repo.EntityTemplateOut
//	@Router		/v1/templates/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTemplatesUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.EntityTemplateUpdate) (repo.EntityTemplateOut, error) {
		auth := services.NewContext(r.Context())
		body.ID = ID
		return ctrl.repo.EntityTemplates.Update(r.Context(), auth.GID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleEntityTemplatesDelete godoc
//
//	@Summary	Delete Entity Template
//	@Tags		Entity Templates
//	@Produce	json
//	@Param		id	path	string	true	"Template ID"
//	@Success	204
//	@Router		/v1/templates/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTemplatesDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.EntityTemplates.Delete(r.Context(), auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

type EntityTemplateCreateItemRequest struct {
	Name        string    `json:"name"        validate:"required,min=1,max=255"`
	Description string    `json:"description" validate:"max=1000"`
	ParentID    uuid.UUID `json:"parentId"    validate:"required"`
	// EntityTypeID is the entity type selected by the user. When set it takes
	// precedence; when empty the repository falls back to the group's default.
	EntityTypeID uuid.UUID   `json:"entityTypeId"`
	TagIDs       []uuid.UUID `json:"tagIds"`
	Quantity     *float64    `json:"quantity"`
}

// HandleEntityTemplateImageUpload godoc
//
//	@Summary		Set Entity Template Default Image
//	@Description	Stores an image on the template. Entities created from the template
//				afterwards get it as their primary photo.
//	@Tags			Entity Templates
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id		path		string	true	"Template ID"
//	@Param			file	formData	file	true	"Image file"
//	@Param			name	formData	string	true	"name of the file including extension"
//	@Success		200		{object}	repo.EntityTemplateOut
//	@Failure		422		{object}	validate.ErrorResponse
//	@Router			/v1/templates/{id}/image [POST]
//	@Security		Bearer
func (ctrl *V1Controller) HandleEntityTemplateImageUpload() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		err := r.ParseMultipartForm(ctrl.maxUploadSize << 20)
		if err != nil {
			log.Err(err).Msg("failed to parse multipart form")
			return multipartFormError(err)
		}

		errs := validate.NewFieldErrors()

		file, _, err := r.FormFile("file")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrMissingFile):
				errs = errs.Append("file", "file is required")
			default:
				log.Err(err).Msg("failed to get file from form")
				return validate.NewRequestError(err, http.StatusInternalServerError)
			}
		}

		name := r.FormValue("name")
		if name == "" {
			errs = errs.Append("name", "name is required")
		}

		if !errs.Nil() {
			return server.JSON(w, http.StatusUnprocessableEntity, errs)
		}

		name = sanitizeAttachmentName(name)
		if !isImageExtension(name) {
			errs = errs.Append("file", "file must be an image")
			return server.JSON(w, http.StatusUnprocessableEntity, errs)
		}

		id, err := ctrl.routeID(r)
		if err != nil {
			return err
		}

		auth := services.NewContext(r.Context())
		out, err := ctrl.repo.EntityTemplates.SetDefaultImage(auth, auth.GID, id, repo.ItemCreateAttachment{
			Title:   name,
			Content: file,
		})
		if err != nil {
			log.Err(err).Msg("failed to set template image")
			return validate.NewRequestError(err, http.StatusInternalServerError)
		}

		return server.JSON(w, http.StatusOK, out)
	}
}

// HandleEntityTemplateImageDelete godoc
//
//	@Summary		Delete Entity Template Default Image
//	@Description	Removes the template's default image. Entities already created from
//				the template keep the copy they were given.
//	@Tags			Entity Templates
//	@Produce		json
//	@Param			id	path		string	true	"Template ID"
//	@Success		200	{object}	repo.EntityTemplateOut
//	@Router			/v1/templates/{id}/image [DELETE]
//	@Security		Bearer
func (ctrl *V1Controller) HandleEntityTemplateImageDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, id uuid.UUID) (repo.EntityTemplateOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTemplates.ClearDefaultImage(auth, auth.GID, id)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleEntityTemplateImageGet godoc
//
//	@Summary	Get Entity Template Default Image
//	@Tags		Entity Templates
//	@Produce	application/octet-stream
//	@Param		id	path	string	true	"Template ID"
//	@Success	200	{file}	file
//	@Router		/v1/templates/{id}/image [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTemplateImageGet() errchain.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := ctrl.routeID(r)
		if err != nil {
			return err
		}

		auth := services.NewContext(r.Context())
		doc, err := ctrl.repo.EntityTemplates.GetDefaultImage(auth, auth.GID, id)
		if err != nil {
			return validate.NewRequestError(err, http.StatusNotFound)
		}

		return ctrl.serveAttachmentContent(auth, w, r, doc)
	}
}

// HandleEntityTemplatesCreateItem godoc
//
//	@Summary	Create Entity from Template
//	@Tags		Entity Templates
//	@Produce	json
//	@Param		id		path		string							true	"Template ID"
//	@Param		payload	body		EntityTemplateCreateItemRequest	true	"Entity Data"
//	@Success	201		{object}	repo.EntityOut
//	@Router		/v1/templates/{id}/create-item [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTemplatesCreateItem() errchain.HandlerFunc {
	fn := func(r *http.Request, templateID uuid.UUID, body EntityTemplateCreateItemRequest) (repo.EntityOut, error) {
		auth := services.NewContext(r.Context())

		template, err := ctrl.repo.EntityTemplates.GetOne(r.Context(), auth.GID, templateID)
		if err != nil {
			return repo.EntityOut{}, err
		}

		quantity := template.DefaultQuantity
		if body.Quantity != nil {
			quantity = *body.Quantity
		}

		// Build custom fields from template
		fields := lo.Map(template.Fields, func(f repo.TemplateField, _ int) repo.EntityFieldData {
			return repo.EntityFieldData{
				Type:         f.Type,
				Name:         f.Name,
				TextValue:    f.TextValue,
				NumberValue:  f.NumberValue,
				BooleanValue: f.BooleanValue,
			}
		})

		// Create entity with all template data in a single transaction
		return ctrl.repo.Entities.CreateFromTemplate(r.Context(), auth.GID, repo.EntityCreateFromTemplate{
			TemplateID:       templateID,
			Name:             body.Name,
			Description:      body.Description,
			Quantity:         quantity,
			ParentID:         body.ParentID,
			EntityTypeID:     body.EntityTypeID,
			TagIDs:           body.TagIDs,
			Insured:          template.DefaultInsured,
			Manufacturer:     template.DefaultManufacturer,
			ModelNumber:      template.DefaultModelNumber,
			LifetimeWarranty: template.DefaultLifetimeWarranty,
			WarrantyDetails:  template.DefaultWarrantyDetails,
			Fields:           fields,
		})
	}

	return adapters.ActionID("id", fn, http.StatusCreated)
}
