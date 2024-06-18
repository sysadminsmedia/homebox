package v1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hay-kot/httpkit/errchain"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/internal/web/adapters"
)

// HandleMaintenanceLogGet godoc
//
//	@Summary  Get Maintenance Log
//	@Tags     Maintenance
//	@Produce  json
//	@Success  200       {object} repo.MaintenanceLog
//	@Router   /v1/items/{id}/maintenance [GET]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceLogGet() errchain.HandlerFunc {
	fn := func(r *http.Request, ID uuid.UUID, q repo.MaintenanceLogQuery) (repo.MaintenanceLog, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.GetLog(auth, auth.GID, ID, q)
	}

	return adapters.QueryID("id", fn, http.StatusOK)
}

// HandleMaintenanceEntryCreate godoc
//
//	@Summary  Create Maintenance Entry
//	@Tags     Maintenance
//	@Produce  json
//	@Param    payload body     repo.MaintenanceEntryCreate true "Entry Data"
//	@Success  201     {object} repo.MaintenanceEntry
//	@Router   /v1/items/{id}/maintenance [POST]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceEntryCreate() errchain.HandlerFunc {
	fn := func(r *http.Request, itemID uuid.UUID, body repo.MaintenanceEntryCreate) (repo.MaintenanceEntry, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.Create(auth, itemID, body)
	}

	return adapters.ActionID("id", fn, http.StatusCreated)
}

// HandleMaintenanceEntryDelete godoc
//
//	@Summary  Delete Maintenance Entry
//	@Tags     Maintenance
//	@Produce  json
//	@Success  204
//	@Router   /v1/items/{id}/maintenance/{entry_id} [DELETE]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceEntryDelete() errchain.HandlerFunc {
	fn := func(r *http.Request, entryID uuid.UUID) (any, error) {
		auth := services.NewContext(r.Context())
		err := ctrl.repo.MaintEntry.Delete(auth, entryID)
		return nil, err
	}

	return adapters.CommandID("entry_id", fn, http.StatusNoContent)
}

// HandleMaintenanceEntryUpdate godoc
//
//	@Summary  Update Maintenance Entry
//	@Tags     Maintenance
//	@Produce  json
//	@Param    payload body     repo.MaintenanceEntryUpdate true "Entry Data"
//	@Success  200     {object} repo.MaintenanceEntry
//	@Router   /v1/items/{id}/maintenance/{entry_id} [PUT]
//	@Security Bearer
func (ctrl *V1Controller) HandleMaintenanceEntryUpdate() errchain.HandlerFunc {
	fn := func(r *http.Request, entryID uuid.UUID, body repo.MaintenanceEntryUpdate) (repo.MaintenanceEntry, error) {
		auth := services.NewContext(r.Context())
		return ctrl.repo.MaintEntry.Update(auth, entryID, body)
	}

	return adapters.ActionID("entry_id", fn, http.StatusOK)
}
