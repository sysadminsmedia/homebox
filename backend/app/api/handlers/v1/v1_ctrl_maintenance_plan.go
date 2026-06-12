package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleMaintenancePlanGetAll godoc
//
//	@Summary	Query Maintenance Plans
//	@Tags		Item Maintenance
//	@Produce	json
//	@Param		id	path		string	true	"Item ID"
//	@Success	200	{array}		repo.MaintenancePlan
//	@Router		/v1/entities/{id}/maintenance/plans [GET]
//	@Security	Bearer
func (ctrl *V1Controller) HandleMaintenancePlanGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request, itemID uuid.UUID, _ struct{}) ([]repo.MaintenancePlan, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.ListPlansByItemID(auth, auth.GID, itemID)
	}

	return adapters.QueryID("id", fn, http.StatusOK)
}

// HandleMaintenancePlanCreate godoc
//
//	@Summary	Create Maintenance Plan
//	@Tags		Item Maintenance
//	@Produce	json
//	@Param		id		path		string					true	"Item ID"
//	@Param		payload	body		repo.MaintenancePlanCreate	true	"Plan Data"
//	@Success	201		{object}	repo.MaintenancePlan
//	@Router		/v1/entities/{id}/maintenance/plans [POST]
//	@Security	Bearer
func (ctrl *V1Controller) HandleMaintenancePlanCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, itemID uuid.UUID, body repo.MaintenancePlanCreate) (repo.MaintenancePlan, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.CreatePlan(auth, itemID, body)
	}

	return adapters.ActionID("id", fn, http.StatusCreated)
}

// HandleMaintenancePlanUpdate godoc
//
//	@Summary	Update Maintenance Plan
//	@Tags		Maintenance
//	@Produce	json
//	@Param		id		path		string					true	"Plan ID"
//	@Param		payload	body		repo.MaintenancePlanUpdate	true	"Plan Data"
//	@Success	200		{object}	repo.MaintenancePlan
//	@Router		/v1/maintenance/plans/{id} [PUT]
//	@Security	Bearer
func (ctrl *V1Controller) HandleMaintenancePlanUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, planID uuid.UUID, body repo.MaintenancePlanUpdate) (repo.MaintenancePlan, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.UpdatePlan(auth, planID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleMaintenancePlanDelete godoc
//
//	@Summary	Delete Maintenance Plan
//	@Tags		Maintenance
//	@Produce	json
//	@Param		id	path	string	true	"Plan ID"
//	@Success	204
//	@Router		/v1/maintenance/plans/{id} [DELETE]
//	@Security	Bearer
func (ctrl *V1Controller) HandleMaintenancePlanDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, planID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		return nil, ctrl.repo.MaintEntry.DeletePlan(auth, planID)
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}
