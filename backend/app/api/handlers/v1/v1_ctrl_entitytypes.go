package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleEntityTypesGetAll godoc
//
//	@Summary	Query All Entity Types
//	@Tags		EntityTypes
//	@Produce	json
//	@Success	200		{array}	repo.EntityType[]
//	@Router		/v1/entitytype [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTypesGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.EntityType, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityType.GetEntityTypesByGroupID(auth, auth.GID)
	}
	return adapters.Command(fn, http.StatusOK)
}

// HandleEntityTypeGetOne godoc
//
//	@Summary	Get One Entity Type
//	@Tags		EntityTypes
//	@Produce	json
//	@Param		id	path	string	true	"Entity Type ID"
//	@Success	200	{object}	repo.EntityType
//	@Router		/v1/entitytype/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTypeGetOne() errchain.HandlerFunc {
	fn := func(r *http.Request, entityTypeID uuid.UUID) (repo.EntityType, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityType.GetOneByGroup(auth, auth.GID, entityTypeID)
	}
	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleEntityTypeCreate godoc
//
//	@Summary	Create Entity Type
//	@Tags		EntityTypes
//	@Accept		json
//	@Produce	json
//	@Param		payload	body		repo.EntityTypeCreate	true	"Entity Type Data"
//	@Success	201		{object}	repo.EntityType
//	@Router		/v1/entitytype [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTypeCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, body repo.EntityTypeCreate) (repo.EntityType, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityType.CreateEntityType(auth, auth.GID, body)
	}
	return adapters.Action(fn, http.StatusCreated)
}

// HandleEntityTypeUpdate godoc
//
//	@Summary	Update Entity Type
//	@Tags		EntityTypes
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string						true	"Entity Type ID"
//	@Param		payload	body		repo.EntityTypeUpdate		true	"Entity Type Data"
//	@Success	200		{object}	repo.EntityType
//	@Router		/v1/entitytype/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleEntityTypeUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, entityTypeID uuid.UUID, body repo.EntityTypeUpdate) (repo.EntityType, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.EntityType.UpdateEntityType(auth, auth.GID, entityTypeID, body)
	}
	return adapters.ActionID("id", fn, http.StatusOK)
}
