package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleTagsGetAll godoc
//
//	@Summary	Get All Tags
//	@Tags		Tags
//	@Produce	json
//	@Success	200	{object}	[]repo.TagOut
//	@Router		/v1/tags [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleTagsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.TagSummary, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Tags.GetAll(auth, auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleTagsCreate godoc
//
//	@Summary	Create Tag
//	@Tags		Tags
//	@Produce	json
//	@Param		payload	body		repo.TagCreate	true	"Tag Data"
//	@Success	200		{object}	repo.TagSummary
//	@Router		/v1/tags [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleTagsCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, data repo.TagCreate) (repo.TagOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Tags.Create(auth, auth.GID, data)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleTagDelete godocs
//
//	@Summary	Delete Tag
//	@Tags		Tags
//	@Produce	json
//	@Param		id	path	string	true	"Tag ID"
//	@Success	204
//	@Router		/v1/tags/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleTagDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.Tags.DeleteByGroup(auth, auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandleTagGet godocs
//
//	@Summary	Get Tag
//	@Tags		Tags
//	@Produce	json
//	@Param		id	path		string	true	"Tag ID"
//	@Success	200	{object}	repo.TagOut
//	@Router		/v1/tags/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleTagGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.TagOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Tags.GetOneByGroup(auth, auth.GID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleTagUpdate godocs
//
//	@Summary	Update Tag
//	@Tags		Tags
//	@Produce	json
//	@Param		id	path		string	true	"Tag ID"
//	@Success	200	{object}	repo.TagOut
//	@Router		/v1/tags/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleTagUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, data repo.TagUpdate) (repo.TagOut, error) {
		auth := services.NewContext(r.Context())
		data.ID = ID
		return ctrl.repo.Tags.UpdateByGroup(auth, auth.GID, data)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}
