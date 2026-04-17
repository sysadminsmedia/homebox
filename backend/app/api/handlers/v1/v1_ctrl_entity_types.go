package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleEntityTypeGetAll godoc
//
//	@Summary	Get All Entity Types
//	@Tags		Entity Types
//	@Produce	json
//	@Success	200	{array}	repo.EntityTypeSummary
//	@Router		/v1/entity-types [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTypeGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.EntityTypeSummary, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTypes.GetAll(r.Context(), auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleEntityTypeCreate godoc
//
//	@Summary	Create Entity Type
//	@Tags		Entity Types
//	@Produce	json
//	@Param		payload	body		repo.EntityTypeCreate	true	"Entity Type Data"
//	@Success	201		{object}	repo.EntityTypeSummary
//	@Router		/v1/entity-types [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTypeCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.EntityTypeCreate) (repo.EntityTypeSummary, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityTypes.Create(r.Context(), auth.GID, body)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleEntityTypeUpdate godoc
//
//	@Summary	Update Entity Type
//	@Tags		Entity Types
//	@Produce	json
//	@Param		id		path		string					true	"Entity Type ID"
//	@Param		payload	body		repo.EntityTypeUpdate	true	"Entity Type Data"
//	@Success	200		{object}	repo.EntityTypeSummary
//	@Router		/v1/entity-types/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTypeUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, body repo.EntityTypeUpdate) (repo.EntityTypeSummary, error) {
		auth := services.NewContext(r.Context())
		body.ID = ID
		return ctrl.repo.EntityTypes.Update(r.Context(), auth.GID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleEntityTypeDelete godoc
//
//	@Summary	Delete Entity Type
//	@Tags		Entity Types
//	@Produce	json
//	@Param		id	path	string	true	"Entity Type ID"
//	@Success	204
//	@Router		/v1/entity-types/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTypeDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.EntityTypes.Delete(r.Context(), auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}
