package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/samber/lo"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleEntityTemplatesGetAll godoc
//
//	@Summary	Get All Entity Templates
//	@Tags		Entity Templates
//	@Produce	json
//	@Success	200	{object}	[]repo.EntityTemplateSummary
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
	Name        string      `json:"name"        validate:"required,min=1,max=255"`
	Description string      `json:"description" validate:"max=1000"`
	ParentID    uuid.UUID   `json:"parentId"    validate:"required"`
	TagIDs      []uuid.UUID `json:"tagIds"`
	Quantity    *float64    `json:"quantity"`
}

// HandleEntityTemplatesCreateItem godoc
//
//	@Summary	Create Entity from Template
//	@Tags		Entity Templates
//	@Produce	json
//	@Param		id		path		string								true	"Template ID"
//	@Param		payload	body		EntityTemplateCreateItemRequest		true	"Entity Data"
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
				Type:      f.Type,
				Name:      f.Name,
				TextValue: f.TextValue,
			}
		})

		// Create entity with all template data in a single transaction
		return ctrl.repo.Entities.CreateFromTemplate(r.Context(), auth.GID, repo.EntityCreateFromTemplate{
			Name:             body.Name,
			Description:      body.Description,
			Quantity:         quantity,
			ParentID:         body.ParentID,
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
