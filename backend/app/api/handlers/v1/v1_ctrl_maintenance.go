package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleMaintenancesGetAll godoc
//
//	@Summary  Query All Maintenances
//	@Tags     Maintenances
//	@Produce  json
//	@Param    filters query    repo.MaintenancesFilters     false "which maintenances to retrieve"
//	@Success  200       {array} repo.MaintenanceEntryWithDetails[]
//	@Router   /v1/maintenances [GET]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenancesGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request, filters repo.MaintenancesFilters) ([]repo.MaintenanceEntryWithDetails, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.GetAllMaintenances(auth, auth.GID, filters)
	}

	return adapters.Query(fn, http.StatusOK)
}

// HandleMaintenanceEntryUpdate godoc
//
//	@Summary  Update Maintenance Entry
//	@Tags     Maintenances
//	@Produce  json
//	@Param    payload body     repo.MaintenanceEntryUpdate true "Entry Data"
//	@Success  200     {object} repo.MaintenanceEntry
//	@Router   /v1/maintenances/{id} [PUT]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceEntryUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, entryID uuid.UUID, body repo.MaintenanceEntryUpdate) (repo.MaintenanceEntry, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.Update(auth, entryID, body)
	}

	return adapters.ActionID("id", fn, http.StatusOK)
}

// HandleMaintenanceEntryDelete godoc
//
//	@Summary  Delete Maintenance Entry
//	@Tags     Maintenances
//	@Produce  json
//	@Success  204
//	@Router   /v1/maintenances/{id} [DELETE]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceEntryDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, entryID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.MaintEntry.Delete(auth, entryID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}
