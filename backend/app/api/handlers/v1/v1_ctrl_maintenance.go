package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleMaintenanceGetAll godoc
//
//	@Summary  Query All Maintenance
//	@Tags     Maintenance
//	@Produce  json
//	@Param    filters query    repo.MaintenanceFilters     false "which maintenance to retrieve"
//	@Success  200       {array} repo.MaintenanceEntryWithDetails[]
//	@Router   /v1/maintenance [GET]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceGetAll() errchain.HandlerFunc {
	fn := func(r *http.Request, filters repo.MaintenanceFilters) ([]repo.MaintenanceEntryWithDetails, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.GetAllMaintenance(auth, auth.GID, filters)
	}

	return adapters.Query(fn, http.StatusOK)
}

// HandleMaintenanceEntryUpdate godoc
//
//	@Summary  Update Maintenance Entry
//	@Tags     Maintenance
//	@Produce  json
//	@Param    id  path     string true "Maintenance ID"
//	@Param    payload body     repo.MaintenanceEntryUpdate true "Entry Data"
//	@Success  200     {object} repo.MaintenanceEntry
//	@Router   /v1/maintenance/{id} [PUT]
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
//	@Tags     Maintenance
//	@Produce  json
//	@Param    id  path     string true "Maintenance ID"
//	@Success  204
//	@Router   /v1/maintenance/{id} [DELETE]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceEntryDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, entryID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.MaintEntry.Delete(auth, entryID)
		return nil, err
	}

	return adapters.CommandID("id", fn, http.StatusNoContent)
}
