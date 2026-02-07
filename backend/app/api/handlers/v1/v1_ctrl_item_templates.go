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

// HandleItemTemplatesGetAll godoc
//
//	@Summary	Get All Item Templates
//	@Tags		Item Templates
//	@Produce	json
//	@Success	200	{object}	[]repo.ItemTemplateSummary
//	@Router		/v1/templates [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemTemplatesGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.ItemTemplateSummary, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.ItemTemplates.GetAll(r.Context(), auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleItemTemplatesGet godoc
//
//	@Summary	Get Item Template
//	@Tags		Item Templates
//	@Produce	json
//	@Param		id	path		string	true	"Template ID"
//	@Success	200	{object}	repo.ItemTemplateOut
//	@Router		/v1/templates/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemTemplatesGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.ItemTemplateOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.ItemTemplates.GetOne(r.Context(), auth.GID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleItemTemplatesCreate godoc
//
//	@Summary	Create Item Template
//	@Tags		Item Templates
//	@Produce	json
//	@Param		payload	body		repo.ItemTemplateCreate	true	"Template Data"
//	@Success	201		{object}	repo.ItemTemplateOut
//	@Router		/v1/templates [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemTemplatesCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.ItemTemplateCreate) (repo.ItemTemplateOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.ItemTemplates.Create(r.Context(), auth.GID, body)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleItemTemplatesUpdate godoc
//
//	@Summary	Update Item Template
//	@Tags		Item Templates
//	@Produce	json
//	@Param		id		path		string					true	"Template ID"
//	@Param		payload	body		repo.ItemTemplateUpdate	true	"Template Data"
//	@Success	200		{object}	repo.ItemTemplateOut
//	@Router		/v1/templates/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemTemplatesUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.ItemTemplateUpdate) (repo.ItemTemplateOut, error) {
		auth := services.NewContext(r.Context())
		body.ID = ID
		return ctrl.repo.ItemTemplates.Update(r.Context(), auth.GID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleItemTemplatesDelete godoc
//
//	@Summary	Delete Item Template
//	@Tags		Item Templates
//	@Produce	json
//	@Param		id	path	string	true	"Template ID"
//	@Success	204
//	@Router		/v1/templates/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemTemplatesDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.ItemTemplates.Delete(r.Context(), auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

type ItemTemplateCreateItemRequest struct {
	Name        string      `json:"name"        validate:"required,min=1,max=255"`
	Description string      `json:"description" validate:"max=1000"`
	LocationID  uuid.UUID   `json:"locationId"  validate:"required"`
	TagIDs      []uuid.UUID `json:"tagIds"`
	Quantity    *int        `json:"quantity"`
}

// HandleItemTemplatesCreateItem godoc
//
//	@Summary	Create Item from Template
//	@Tags		Item Templates
//	@Produce	json
//	@Param		id		path		string							true	"Template ID"
//	@Param		payload	body		ItemTemplateCreateItemRequest	true	"Item Data"
//	@Success	201		{object}	repo.ItemOut
//	@Router		/v1/templates/{id}/create-item [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleItemTemplatesCreateItem() errchain.HandlerFunc {
	fn := func(r *http.Request, templateID uuid.UUID, body ItemTemplateCreateItemRequest) (repo.ItemOut, error) {
		auth := services.NewContext(r.Context())

		template, err := ctrl.repo.ItemTemplates.GetOne(r.Context(), auth.GID, templateID)
		if err != nil {
			return repo.ItemOut{}, err
		}

		quantity := template.DefaultQuantity
		if body.Quantity != nil {
			quantity = *body.Quantity
		}

		// Build custom fields from template
		fields := lo.Map(template.Fields, func(f repo.TemplateField, _ int) repo.ItemField {
			return repo.ItemField{
				Type:      f.Type,
				Name:      f.Name,
				TextValue: f.TextValue,
			}
		})

		// Create item with all template data in a single transaction
		return ctrl.repo.Items.CreateFromTemplate(r.Context(), auth.GID, repo.ItemCreateFromTemplate{
			Name:             body.Name,
			Description:      body.Description,
			Quantity:         quantity,
			LocationID:       body.LocationID,
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
