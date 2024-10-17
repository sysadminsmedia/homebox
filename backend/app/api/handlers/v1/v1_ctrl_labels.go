package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleLabelsGetAll godoc
//
//	@Summary	Get All Labels
//	@Tags		Labels
//	@Produce	json
//	@Success	200	{object}	[]repo.LabelOut
//	@Router		/v1/labels [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelsGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request) ([]repo.LabelSummary, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Labels.GetAll(auth, auth.GID)
	}

	return adapters.Command(fn, http.StatusOK)
}

// HandleLabelsCreate godoc
//
//	@Summary	Create Label
//	@Tags		Labels
//	@Produce	json
//	@Param		payload	body		repo.LabelCreate	true	"Label Data"
//	@Success	200		{object}	repo.LabelSummary
//	@Router		/v1/labels [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelsCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, data repo.LabelCreate) (repo.LabelOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Labels.Create(auth, auth.GID, data)
	}

	return adapters.Action(fn, http.StatusCreated)
}

// HandleLabelDelete godocs
//
//	@Summary	Delete Label
//	@Tags		Labels
//	@Produce	json
//	@Param		id	path	string	true	"Label ID"
//	@Success	204
//	@Router		/v1/labels/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.Labels.DeleteByGroup(auth, auth.GID, ID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}

// HandleLabelGet godocs
//
//	@Summary	Get Label
//	@Tags		Labels
//	@Produce	json
//	@Param		id	path		string	true	"Label ID"
//	@Success	200	{object}	repo.LabelOut
//	@Router		/v1/labels/{id} [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID) (repo.LabelOut, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.Labels.GetOneByGroup(auth, auth.GID, ID)
	}

	return adapters.CommandID("id", fn, http.StatusOK)
}

// HandleLabelUpdate godocs
//
//	@Summary	Update Label
//	@Tags		Labels
//	@Produce	json
//	@Param		id	path		string	true	"Label ID"
//	@Success	200	{object}	repo.LabelOut
//	@Router		/v1/labels/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleLabelUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, data repo.LabelUpdate) (repo.LabelOut, error) {
		auth := services.NewContext(r.Context())
		data.ID = ID
		return ctrl.repo.Labels.UpdateByGroup(auth, auth.GID, data)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}
